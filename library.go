package irclib

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"regexp"
	"strings"
	"sync"

	"github.com/crimsonvoid/console"
	"github.com/crimsonvoid/irclib/module"
	irc "github.com/fluffle/goirc/client"
)

type ModManager struct {
	Conn   *irc.Conn // IRC Connection
	Config *BotInfo  // Bot config

	core    *module.Module   // Core "master" module
	modules []*module.Module // List of registered modules
	mut     sync.RWMutex
	running bool

	cons *console.Console // Console to get input

	Quit chan bool // Quit chan to block until a successful disconnect or force disconnect
}

// Returns a new ModManager configured with a JSON file. Errors indicate file
// reading or JSON unmarshalling
func New(fileName string) (*ModManager, error) {
	file, err := ioutil.ReadFile(fileName)
	if err != nil {
		return nil, err
	}

	servInfo := new(ServerInfo)
	err = json.Unmarshal(file, servInfo)
	if err != nil {
		return nil, err
	}

	return NewManager(servInfo)
}

// Create a new ModManager from a ServerInfo config
func NewManager(serverInfo *ServerInfo) (*ModManager, error) {
	ircCfg, err := serverInfo.configServer()
	if err != nil {
		return nil, err
	}
	con := irc.Client(ircCfg)

	// copy Chans and Accesss to allow serverInfo to be marked for GC
	chans := make([]string, len(serverInfo.Botinfo.Chans))
	copy(chans, serverInfo.Botinfo.Chans)

	access := access{
		list: make(map[string][]string, len(serverInfo.Botinfo.Access.list)),
	}

	for name, list := range serverInfo.Botinfo.Access.list {
		l := make([]string, len(list))
		copy(l, list)
		access.list[name] = l
	}

	m := &ModManager{
		core:    newCore(),
		modules: make([]*module.Module, 0, 5),
		cons:    console.New(),

		Conn: con,
		Config: &BotInfo{
			Chans:  chans,
			Access: access,
		},
		Quit: make(chan bool),
	}
	m.registerCoreCommands()
	m.registerCommands()

	return m, nil
}

// Registers a unique module; "core" is a reserved module name. If a module with
// the same name is already registered an error is returned
func (self *ModManager) Register(mod *module.Module) error {
	name := mod.Name()

	if mod.Conn != nil {
		return errors.New("Module is registered with a ModManager")
	}

	if name == self.core.Name() {
		return errors.New("Module name is already registered")
	}

	self.mut.Lock()
	defer self.mut.Unlock()

	for _, v := range self.modules {
		if v.Name() == name {
			return errors.New("Module name is already registered")
		}
	}

	mod.Conn = self.Conn
	self.modules = append(self.modules, mod)

	return nil
}

// Connect to IRC. If a registered module's PreStart() returns an error it does
// not attempt to call Start(). A map of module names to error is returned if
// there were any errors or nil if none. Errors are logged to the core module.
// If there are errors in "core" from the map then Connect() failed
func (self *ModManager) Connect() map[string]error {
	self.mut.Lock()
	defer self.mut.Unlock()

	errMap := make(map[string]error)
	errMut := new(sync.RWMutex)
	wg := new(sync.WaitGroup)

	if self.running {
		errMap["core"] = errors.New("ModManager is already running")

		return errMap
	}

	if self.Conn == nil || self.Config == nil {
		errMap["core"] = errors.New("Improperly configured ModManager")

		return errMap
	}

	self.setupHandlers()
	go self.cons.Monitor()

	// module.PreStart()
	wg.Add(len(self.modules))
	for _, mod := range self.modules {
		if mod.Conn == nil {
			mod.Conn = self.Conn
		}

		go func(mod *module.Module) {
			defer wg.Done()

			err := mod.PreStart()
			if err == nil {
				return
			}

			self.core.Logger.Errorf("%v.PreStart() error: %v\n", mod.Name(), err)

			errMut.Lock()
			errMap[mod.Name()] = err
			errMut.Unlock()
		}(mod)
	}
	wg.Wait()

	// Connect to IRC
	if err := self.Conn.Connect(); err != nil {
		self.core.Logger.Errorf("Error connecting: %v\n", err)
		errMap["core"] = err

		return errMap
	}

	// module.Start()
	wg.Add(len(self.modules))
	for _, mod := range self.modules {
		go func(mod *module.Module) {
			defer wg.Done()

			errMut.RLock()
			_, ok := errMap[mod.Name()]
			errMut.RUnlock()
			if ok {
				return
			}

			err := mod.Start()
			if err == nil {
				return
			}

			self.core.Logger.Errorf("%v.Start() error: %v\n", mod.Name(), err)

			errMut.Lock()
			errMap[mod.Name()] = err
			errMut.Unlock()
		}(mod)
	}
	wg.Wait()

	self.running = true

	log.Printf("%v%v connected to %v%v\n",
		console.FgGreen, self.Conn.Config().Me.Nick, self.Conn.Config().Server,
		console.Reset)
	self.core.Logger.Infof("%v connected to %v\n",
		self.Conn.Config().Me.Nick, self.Conn.Config().Server)

	return errMap
}

// Disconnects and cleans up. Returns a map of module names to errors if
// module.Exit() returned an error. Errors returned by module.Exit() are logged
// to core. If there are any errors when calling module.Exit() ModManager will
// not finish cleanup
func (self *ModManager) Disconnect() map[string]error {
	self.mut.Lock()
	defer self.mut.Unlock()

	errMap := make(map[string]error)
	errMut := new(sync.RWMutex)
	wg := new(sync.WaitGroup)

	if !self.running {
		errMap["core"] = errors.New("ModManager is not running")

		return errMap
	}

	wg.Add(len(self.modules))
	for _, mod := range self.modules {
		go func(mod *module.Module) {
			defer wg.Done()

			err := mod.Exit()
			if err == nil {
				return
			}

			self.core.Logger.Errorf("%v.Exit() error: %v\n", mod.Name(), err)

			errMut.Lock()
			errMap[mod.Name()] = err
			errMut.Unlock()
		}(mod)
	}
	wg.Wait()

	if len(errMap) != 0 {
		self.core.Logger.Infoln("Errors when attempting to disconnect")

		for modName, err := range errMap {
			self.core.Logger.Errorf("irclibrary.Disconnect() %v: %v\n", modName, err)
		}

		return errMap
	}

	self.core.Logger.Infoln("Disconnected without errors")

	if err := self.core.Exit(); err != nil {
		log.Println("Core breach! Unsuccessful Exit():", err)
		errMap["core"] = err

		return errMap
	}

	self.cons.Stop()
	if self.Conn.Connected() {
		self.Conn.Quit()
	}

	self.running = false

	return errMap
}

// Force Disconnect all modules, returning a map of modules to a list or errors
func (self *ModManager) ForceDisconnect() map[string][]error {
	self.mut.Lock()
	defer self.mut.RUnlock()

	errMap := make(map[string][]error)
	errMut := new(sync.Mutex)
	wg := new(sync.WaitGroup)

	wg.Add(len(self.modules))
	for _, mod := range self.modules {
		go func(mod *module.Module) {
			defer wg.Done()

			err := mod.ForceExit()
			if err == nil {
				return
			}

			self.core.Logger.Errorf("%v.ForceExit() error: %v\n", mod.Name(), err)

			errMut.Lock()
			errMap[mod.Name()] = err
			errMut.Unlock()
		}(mod)
	}
	wg.Wait()

	if len(errMap) == 0 {
		self.core.Logger.Infoln("Force disconnected without errors")
	} else {
		self.core.Logger.Infoln("Errors when attempting to force disconnect")

		for modName, errs := range errMap {
			for _, err := range errs {
				self.core.Logger.Errorf("irclibrary.ForceDisconnect() %v: %v\n", modName, err)
			}
		}
	}

	if err := self.core.ForceExit(); err != nil {
		log.Println("Core breach! unsuccessful ForceExit():", err)
		errMap["core"] = err

	}

	self.cons.Stop()

	if self.Conn.Connected() {
		self.Conn.Quit()
	}

	return errMap
}

// Force disconnect a module aggregating all errors and returning that list
func (self *ModManager) ForceDisconnectModule(modName string) []error {
	self.mut.RLock()
	defer self.mut.RUnlock()

	for _, mod := range self.modules {
		if mod.Name() == modName {
			return mod.ForceExit()
		}
	}

	return nil
}

func (self *ModManager) Running() bool {
	self.mut.RLock()
	defer self.mut.RUnlock()

	return self.running
}

// Register console commands
func (self *ModManager) registerCommands() {
	// Register console handler for commands
	re := regexp.MustCompile(`^:(?P<name>\w+)\s(?P<command>.+)$`)
	self.cons.RegisterRegexp(re, func(s string) {
		groups, err := matchGroups(re, s)
		if err != nil {
			return
		}

		groups["name"] = strings.ToLower(groups["name"])
		if groups["name"] == "core" {
			go self.core.Console.Parse(groups["command"])
			return
		}

		for _, mod := range self.modules {
			if mod.Name() == groups["name"] {
				go mod.Console.Parse(groups["command"])
				break
			}
		}
	})

	// Register quit and forece quit
	self.cons.Register(":q", func(s string) {
		self.coreDisconnect()
	})

	re2 := regexp.MustCompile(`^f(orce\s)?quit\s(?P<module>.*)?$`)
	self.cons.RegisterRegexp(re2, func(trigger string) {
		self.coreForceDisconnect(re2, trigger)
	})
}
