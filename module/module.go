// Package for creating and managing Module's

package module

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
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

	// Function used for ":moduleName info" console command
	String func() string

	running bool
	file    *os.File      // File to write logs to
	bufFile *bufio.Writer // Buffered writer of Module.file

	stTriggers   map[eventTrigger][]func(*irc.Line)
	reTriggers   map[Event][]*re
	stMut, reMut sync.RWMutex

	Console *Console // Console handler; commands are triggered with ":moduleName <command>"
	Logger  *Logger
}

func init() {
	devNil, err := os.Open(os.DevNull)
	if err != nil {
		panic(err)
	}

	flagSet := flag.NewFlagSet(os.Args[0], flag.PanicOnError)
	flagSet.SetOutput(devNil)

	flagSet.StringVar(&logDir, "logDir", "./logs", "Set log directory")
	flagSet.Parse(os.Args[1:])

	if logDir[len(logDir)-1] != '/' {
		logDir = logDir + "/"
	}

	if err := devNil.Close(); err != nil {
		panic(err)
	}
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
	} else if self.LogDir[len(logDir)-1:] != "/" {
		self.LogDir = self.LogDir + "/"
	}

	toLowerSlice(self.AllowUser)
	toLowerSlice(self.DenyUser)
	toLowerSlice(self.AllowChan)
	toLowerSlice(self.DenyChan)

	mod := &Module{
		moduleConfig: moduleConfig{
			name:        self.Name,
			description: self.Description,
			logDir:      self.LogDir,
			enabled:     self.Enabled,

			allowUser: self.AllowUser,
			denyUser:  self.DenyUser,
			allowChan: self.AllowChan,
			denyChan:  self.DenyChan,
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
	if self.file == nil || self.bufFile == nil {
		if err := self.createLogger(); err != nil {
			log.Println(self.Name(), "error creating log file", err)

			return err
		}
	}

	if self.Preconnect != nil {
		if err := self.Preconnect(); err != nil {
			self.Logger.Errorln(err)

			return err
		}
	}

	return nil
}

// Creates a Logger if necessay and calls Connet() if applicable. This is
// exported for use by library and most likely doesn't need to be called by the
// user. Error is non-nil if a logger could not be created, module is already
// running or Connect() returned an error; errors returned by Connect() are logged
func (self *Module) Start() error {
	if self.running {
		return fmt.Errorf("Module.Start(): %v is already running", self.Name())
	}

	if self.file == nil || self.bufFile == nil {
		if err := self.createLogger(); err != nil {
			return fmt.Errorf("Module.Start(): %v", err.Error())
		}
	}

	self.running = true

	if self.Connected != nil {
		if err := self.Connected(); err != nil {
			return fmt.Errorf("Module.Start(): %v", err.Error())
		}
	}

	return nil
}

// Calls Disconnect(), cleans up, and exits. Errors returned by Disconnect() are
// logged and Exit() continues. If there is an error at any point the error is
// returned and should be assumed that cleanup did not complete.
func (self *Module) Exit() error {
	if !self.running {
		return fmt.Errorf("Module.Exit(): %v is not running", self.Name())
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
// which are aggregated and returned in a slice
func (self *Module) ForceExit() []error {
	errs := make([]error, 0, 3)

	if self.Disconnect != nil {
		if err := self.Disconnect(); err != nil {
			errs = append(errs, err)
		}
	}

	self.Logger.exit()
	if self.bufFile != nil {
		if err := self.bufFile.Flush(); err != nil {
			errs = append(errs, err)
		}
	}
	if self.file != nil {
		if err := self.file.Close(); err != nil {
			errs = append(errs, err)
		}
	}

	self.file, self.bufFile = nil, nil

	if len(errs) == 0 {
		return nil
	} else {
		return errs
	}
}

// Register a function that is called when an Event of eventMode is triggered and
// trigger equals input. trigger is lowered before registering.
func (self *Module) Register(trigger string, eventMode Event, fn func(*irc.Line)) {
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
func (self *Module) RegisterRegexp(trigger *regexp.Regexp, eventMode Event, fn func(*irc.Line)) {
	eventMode = Event(strings.ToUpper(string(eventMode)))
	appendEvent(eventMode)

	reM := &re{
		trigger: trigger,
		fn:      fn,
	}

	self.reMut.Lock()
	defer self.reMut.Unlock()

	fns := self.reTriggers[eventMode]
	fns = append(fns, reM)
	self.reTriggers[eventMode] = fns
}

// Handles triggers if module is enabled and user/chan is allowed. This is mainly
// exported for use by library and should not have to be called by the user
func (self *Module) Handle(trigger string, eventMode Event, line *irc.Line) {
	// Filtered by: denyUser, allowUser, denyChan, allowChan
	if !self.Enabled() ||
		self.InDenyed(line.Nick) ||
		// Empty allowUser list => allow all
		(self.LenAllowed(User) != 0 && !self.InAllowed(line.Nick)) ||
		self.InDenyed(line.Target()) ||
		// Empty denyChan list => allow all
		(self.LenAllowed(Chan) != 0 && !self.InAllowed(line.Target())) {

		return
	}

	eventMode = Event(strings.ToUpper(string(eventMode)))

	go self.handleString(trigger, eventMode, line)
	go self.handleRegexp(trigger, eventMode, line)
}

func (self *Module) handleString(trigger string, eventMode Event, line *irc.Line) {
	trigger = strings.ToLower(trigger)
	evT := eventTrigger{eventMode, trigger}

	self.stMut.RLock()
	defer self.stMut.RUnlock()

	for _, fn := range self.stTriggers[evT] {
		go fn(line.Copy())
	}
}

func (self *Module) handleRegexp(trigger string, eventMode Event, line *irc.Line) {
	self.reMut.RLock()
	defer self.reMut.RUnlock()

	for _, reM := range self.reTriggers[eventMode] {
		if reM.trigger.FindStringSubmatch(trigger) != nil {
			go reM.fn(line.Copy())
		}
	}
}

// Returns a list of IRC commands registered
func (self *Module) StringCommands() []string {
	output := make([]string, 0, len(self.stTriggers)+len(self.reTriggers))
	self.stMut.RLock()
	self.reMut.RLock()
	defer self.stMut.RUnlock()
	defer self.reMut.RUnlock()

	for evTrig, _ := range self.stTriggers {
		output = append(output, fmt.Sprintf("[%-12v] %v", evTrig.event, evTrig.trigger))
	}

	for event, reS := range self.reTriggers {
		for _, re := range reS {
			output = append(output, fmt.Sprintf("[%-12v] %v", event, re.trigger))
		}
	}

	return output
}

func (self *Module) createLogger() error {
	if self.LogDir() == "" {
		self.SetLogDir(logDir)
	} else if lDir := self.LogDir(); lDir[len(logDir)-1] != '/' {
		self.SetLogDir(lDir + "/")
	}

	if err := os.MkdirAll(self.LogDir(), 755); err != nil || os.IsNotExist(err) {
		return err
	}

	logName := fmt.Sprintf("%v%v.log", self.LogDir(), self.Name())
	file, err := os.OpenFile(logName, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 644)
	if err != nil {
		return err
	}

	self.file = file
	self.bufFile = bufio.NewWriter(self.file)
	self.Logger = newLogger(self.bufFile, "", Lpriority|LstdFlags, Pinfo)
	go self.Logger.start()

	return nil
}
