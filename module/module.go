// Package for creating and managing Module's

package module

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"regexp"
	"strings"
	"sync"

	irc "github.com/fluffle/goirc/client"
)

type re struct {
	trigger *regexp.Regexp
	fn      func(*irc.Line)
}

type eventTrigger struct {
	event   Event // IRC event
	trigger string
}

// Module repersents the state of the module
type Module struct {
	moduleConfig

	// IRC Connection. Conn is not assigned until it is registered
	Conn *irc.Conn

	// Connect functions to call before or after IRC connection
	// Disconnect is called after disconnected from IRC
	// Errors are logged to to the module Logger
	Preconnect, Connected, Disconnect func() error

	running bool
	file    *os.File      // File to write logs to
	bufFile *bufio.Writer // Buffered writer of Module.file

	stTriggers   map[eventTrigger][]func(*irc.Line)
	reTriggers   map[Event][]*re
	stMut, reMut sync.RWMutex

	Console *Console // Console handler; commands are triggered with ":moduleName <command>"
	Logger  *Logger
}

// Read a JSON file and return a configured Module. Errors indicate a failure to
// parse the file or an incomplete configuration
func New(configFile string) (*Module, error) {
	file, err := ioutil.ReadFile(configFile)
	if err != nil {
		return nil, err
	}

	modInfo := new(ModuleInfo)
	err = json.Unmarshal(file, modInfo)
	if err != nil {
		return nil, err
	}

	return modInfo.NewModule()
}

// Returns a configured Module from ModuleInfo. ModuleInfo.Name and
// ModuleInfo.Description can not be an empty string
func (self *ModuleInfo) NewModule() (*Module, error) {
	if self.Name == "" || self.Description == "" {
		return nil, fmt.Errorf("Improperly configured ModuleInfo")
	}

	if self.LogDir == "" {
		self.LogDir = logDir
	} else if self.LogDir[len(self.LogDir)-1:] != "/" {
		self.LogDir = self.LogDir + "/"
	}

	toLowerSlice(self.AllowUser)
	toLowerSlice(self.DenyUser)
	toLowerSlice(self.AllowChan)
	toLowerSlice(self.DenyChan)

	mod := &Module{
		moduleConfig: moduleConfig{
			m: *self,
		},

		stTriggers: make(map[eventTrigger][]func(*irc.Line)),
		reTriggers: make(map[Event][]*re),
		Console:    newConsole(),
	}

	if err := mod.createLogger(); err != nil {
		return nil, err
	}
	mod.registerBaseCommands()

	return mod, nil
}

// Creates a Logger if necessay and calls Preconnect() if applicable. This is
// exported for use by library and most likely doesn't need to be called by the
// user. Error is non-nil if a logger could not be created or Preconnect() returned
// an error; errors returned by Preconnect() are logged
func (self *Module) PreStart() error {
	self.mu.Lock()
	defer self.mu.Unlock()

	if self.file == nil || self.bufFile == nil {
		if err := self.createLogger(); err != nil {
			consLog.Println(self.Name(), "error creating log file", err)

			return err
		}
	}

	if self.Preconnect == nil {
		return nil
	}

	err := self.Preconnect()
	if err != nil {
		self.Logger.Errorln(err)
	}

	return err
}

// Creates a Logger if necessay and calls Connet() if applicable. This is
// exported for use by library and most likely does not need to be called by the
// user. Error is non-nil if a logger could not be created, module is already
// running or Connect() returned an error; errors returned by Connect() are logged
func (self *Module) Start() error {
	self.mu.Lock()
	defer self.mu.Unlock()

	if self.running {
		return fmt.Errorf("Module.Start(): %v is already running", self.m.Name)
	}

	if self.file == nil || self.bufFile == nil {
		if err := self.createLogger(); err != nil {
			return fmt.Errorf("Module.Start(): %v", err.Error())
		}
	}

	self.running = true

	if self.Connected == nil {
		return nil
	}

	if err := self.Connected(); err != nil {
		return fmt.Errorf("Module.Start(): %v", err.Error())
	}

	return nil
}

// Calls Disconnect(), cleans up, and exits. Errors returned by Disconnect() are
// logged and Exit() continues. If there is an error at any other point the error
// is returned and should be assumed that cleanup did not complete.
func (self *Module) Exit() error {
	self.mu.Lock()
	defer self.mu.Unlock()

	if !self.running {
		return fmt.Errorf("Module.Exit(): %v is not running", self.m.Name)
	}

	if self.Disconnect != nil {
		if err := self.Disconnect(); err != nil {
			self.Logger.Errorln(err)
		}
	}

	self.Logger.exit()

	if err := self.bufFile.Flush(); err != nil {
		return err
	}

	if err := self.file.Close(); err != nil {
		return err
	}

	self.running = false
	self.file, self.bufFile = nil, nil

	return nil
}

// Calls Disconnect(), cleans up, and exits. ForceExit() continues on errors,
// which are aggregated and returned in a slice. If there are no errors `nil`
// is returned
func (self *Module) ForceExit() []error {
	self.mu.Lock()
	defer self.mu.Unlock()

	errs := make([]error, 0, 5)

	if !self.running {
		errs = append(errs, fmt.Errorf("Module.ForceExit(): %v is not running", self.m.Name))

		return errs
	}

	if self.Disconnect != nil {
		if err := self.Disconnect(); err != nil {
			errs = append(errs, err)
		}
	}

	self.Logger.exit()

	if err := self.bufFile.Flush(); err != nil {
		errs = append(errs, err)
	}

	if err := self.file.Close(); err != nil {
		errs = append(errs, err)
	}

	self.running = false
	self.file, self.bufFile = nil, nil

	if len(errs) == 0 {
		return nil
	}

	return errs
}

func (self *Module) Register(eventMode Event, trigger interface{}, fn func(*irc.Line)) {
	switch trigger.(type) {
	case string:
		self.registerString(eventMode, trigger.(string), fn)
	case *regexp.Regexp:
		self.registerRegexp(eventMode, trigger.(*regexp.Regexp), fn)
	case regexp.Regexp:
		re := trigger.(regexp.Regexp)
		self.registerRegexp(eventMode, &re, fn)
	}
}

// Register a function that is called when an Event of eventMode is triggered and
// trigger equals input. trigger is lowered before registering.
func (self *Module) registerString(eventMode Event, trigger string, fn func(*irc.Line)) {
	trigger = strings.ToLower(trigger)
	eventMode = Event(strings.ToUpper(string(eventMode)))

	appendEvent(eventMode)
	evT := eventTrigger{eventMode, trigger}

	self.stMut.Lock()
	defer self.stMut.Unlock()

	fns := self.stTriggers[evT]
	fns = append(fns, fn)
	self.stTriggers[evT] = fns
}

// Register a function that is called when an Event of eventMode is triggered and
// trigger equals input.
func (self *Module) registerRegexp(eventMode Event, trigger *regexp.Regexp, fn func(*irc.Line)) {
	eventMode = Event(strings.ToUpper(string(eventMode)))

	appendEvent(eventMode)
	reM := &re{trigger, fn}

	self.reMut.Lock()
	defer self.reMut.Unlock()

	fns := self.reTriggers[eventMode]
	fns = append(fns, reM)
	self.reTriggers[eventMode] = fns
}

// Handles triggers if module is enabled and user/chan is allowed. This is mainly
// exported for use by library and should not have to be called by the user
func (self *Module) Handle(eventMode Event, trigger string, line *irc.Line) {
	// Filtered by: denyUser, allowUser, denyChan, allowChan
	if !self.Enabled() ||
		self.InDenyed(line.Nick) ||
		// Empty allowUser list => allow all
		(self.LenAllowed(UC_User) != 0 && !self.InAllowed(line.Nick)) ||
		self.InDenyed(line.Target()) ||
		// Empty denyChan list => allow all
		(self.LenAllowed(UC_Chan) != 0 && !self.InAllowed(line.Target())) {

		return
	}

	eventMode = Event(strings.ToUpper(string(eventMode)))

	go self.handleString(eventMode, trigger, line)
	go self.handleRegexp(eventMode, trigger, line)
}

func (self *Module) handleString(eventMode Event, trigger string, line *irc.Line) {
	trigger = strings.ToLower(trigger)
	evT := eventTrigger{eventMode, trigger}

	self.stMut.RLock()
	defer self.stMut.RUnlock()

	for _, fn := range self.stTriggers[evT] {
		go fn(line.Copy())
	}
}

func (self *Module) handleRegexp(eventMode Event, trigger string, line *irc.Line) {
	self.reMut.RLock()
	defer self.reMut.RUnlock()

	for _, reM := range self.reTriggers[eventMode] {
		if reM.trigger.MatchString(trigger) {
			go reM.fn(line.Copy())
		}
	}
}

// Returns a list of IRC commands registered
func (self *Module) StringCommands() []string {
	output := make([]string, 0, len(self.stTriggers)+len(self.reTriggers))

	self.stMut.RLock()
	for evTrig := range self.stTriggers {
		output = append(output, fmt.Sprintf("[%-12v] %v", evTrig.event, evTrig.trigger))
	}
	self.stMut.RUnlock()

	self.reMut.RLock()
	for event, reS := range self.reTriggers {
		for _, re := range reS {
			output = append(output, fmt.Sprintf("[%-12v] %v", event, re.trigger))
		}
	}
	self.reMut.RUnlock()

	return output
}

func (self *Module) createLogger() error {
	if self.LogDir() == "" {
		self.SetLogDir(logDir)
	} else if lDir := self.LogDir(); lDir[len(logDir)-1] != '/' {
		self.SetLogDir(lDir + "/")
	}

	if err := os.MkdirAll(self.LogDir(), 0755); err != nil || os.IsNotExist(err) {
		return err
	}

	logName := fmt.Sprintf("%v%v.log", self.LogDir(), self.Name())
	file, err := os.OpenFile(logName, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0644)
	if err != nil {
		return err
	}

	self.file = file
	self.bufFile = bufio.NewWriter(self.file)
	self.Logger = newLogger(self.bufFile, "", Lpriority|LstdFlags, Pinfo)

	return nil
}

func SetLogDir(logdir string) {
	if logdir[len(logdir)-1] == '/' {
		logDir = logdir
	} else {
		logDir = logdir + "/"
	}
}
