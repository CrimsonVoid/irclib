package irclibrary

import (
	"fmt"
	"log"
	"regexp"
	"time"

	"github.com/crimsonvoid/console"
	"github.com/crimsonvoid/irclibrary/module"
)

func init() {
	log.SetFlags(0)
}

func newCore() *module.Module {
	modInfo := module.ModuleInfo{
		Name:        "core",
		Description: "IRC Library core module",
		Enabled:     true,
	}
	core, err := modInfo.NewModule()
	if err != nil {
		panic(err)
	}

	if err := core.Start(); err != nil {
		panic(err)
	}

	return core
}

func (self *ModManager) registerCoreCommands() {
	// Quit
	if err := self.regCoreQuit(); err != nil {
		panic(err)
	}
	// Force quit, optional module name
	if err := self.regCoreForceQuit(); err != nil {
		panic(err)
	}
	// List modules
	if err := self.regCoreListModules(); err != nil {
		panic(err)
	}
	// Join or Part channels
	if err := self.regCoreChanManage(); err != nil {
		panic(err)
	}
	// Print access list
	if err := self.regCoreAccessList(); err != nil {
		panic(err)
	}
	// Add or remove nicks from access list
	if err := self.regCoreAccessManip(); err != nil {
		panic(err)
	}
}

func (self *ModManager) regCoreQuit() error {
	err := self.core.Console.Register("quit", func(trigger string) {
		self.coreDisconnect()
	})

	return err
}

func (self *ModManager) regCoreForceQuit() error {
	re := regexp.MustCompile(`^f(orce\s)?quit\s(?P<module>.*)?$`)
	err := self.core.Console.RegisterRegexp(re, func(trigger string) {
		self.coreForceDisconnect(re, trigger)
	})

	return err
}

func (self *ModManager) regCoreListModules() error {
	err := self.core.Console.Register("list", func(trigger string) {
		for _, mod := range self.modules {
			var color string
			if mod.Enabled() {
				color = console.FgGreen
			} else {
				color = console.FgRed
			}

			log.Printf("%v%v%v - %v\n", color, mod.Name(), console.Reset, mod.Description())
		}
	})

	return err
}

func (self *ModManager) regCoreChanManage() error {
	re2 := regexp.MustCompile(`^(?P<cmd>join|part)\s(?P<chan>.*)$`)
	err := self.core.Console.RegisterRegexp(re2, func(trigger string) {
		groups, _ := matchGroups(re2, trigger)
		channel := groups["chan"]
		if channel[0] != '#' {
			channel = "#" + channel
		}

		if groups["cmd"] == "join" {
			self.Conn.Join(channel)
			self.core.Logger.Infoln("Joined", channel)
		} else {
			self.Conn.Part(channel)
			self.core.Logger.Infoln("Parted", channel)
		}
	})

	return err
}

func (self *ModManager) regCoreAccessList() error {
	err := self.core.Console.Register("access list", func(trigger string) {
		for grp, nicks := range self.Config.Access.Groups() {
			log.Printf("%v\n  %v\n", grp, nicks)
		}
	})

	return err
}

func (self *ModManager) regCoreAccessManip() error {
	re := regexp.MustCompile(`^access\s(?P<cmd>add|rem)\s(?P<group>.*)\s(?P<nick>.*)$`)
	err := self.core.Console.RegisterRegexp(re, func(trigger string) {
		groups, _ := matchGroups(re, trigger)
		var msg string

		if groups["cmd"] == "add" {
			if self.Config.Access.Add(groups["nick"], groups["group"]) {
				msg = "Added %v to %v\n"
			} else {
				msg = "%v is already in %v\n"
			}
		} else {
			if self.Config.Access.Remove(groups["nick"], groups["group"]) {
				msg = "Removed %v from %v\n"
			} else {
				msg = "%v is not in %v\n"
			}
		}

		log.Printf(msg, groups["nick"], groups["group"])
		self.core.Logger.Infof(msg, groups["nick"], groups["group"])
	})

	return err
}

func (self *ModManager) coreDisconnect() {
	errors := self.Disconnect()
	if len(errors) == 0 {
		log.Printf("%vDisconnected without errors%v\n", console.FgGreen, console.Reset)
		<-time.After(time.Second * 2)
		self.Quit <- true
		return
	}

	out := fmt.Sprintf("%vErrors when attempting to disconnect%v\n", console.FgRed, console.Reset)
	for modName, err := range errors {
		out += fmt.Sprintf("  %v: %v\n", modName, err)
	}
	log.Print(out)

	self.Quit <- false
}

func (self *ModManager) coreForceDisconnect(re *regexp.Regexp, trigger string) {
	defer func() {
		<-time.After(time.Second * 2)
		self.Quit <- true
	}()

	// re := regexp.MustCompile(`^f(orce\s)?quit\s(?P<module>.*)?$`)
	errMap := make(map[string][]error)
	groups, _ := matchGroups(re, trigger)

	if modName, ok := groups["module"]; ok {
		errs := self.ForceDisconnectModule(modName)
		if len(errs) == 0 {
			log.Printf("%vForce disconnected module %v without errors%v\n",
				console.FgGreen, modName, console.Reset)

			return
		}

		errMap[modName] = errs
	} else if errMap = self.ForceDisconnect(); len(errMap) == 0 {
		log.Printf("%vForce disconnected without errors%v\n",
			console.FgGreen, console.Reset)

		return
	}

	out := fmt.Sprintf("%vErrors when attempting to force disconnect %v%v\n",
		console.FgRed, groups["module"], console.Reset)

	for modName, errs := range errMap {
		errStr := fmt.Sprintf("  %v", modName)
		for _, err := range errs {
			errStr += fmt.Sprintf("    %v\n", err)
		}

		out += errStr
	}
	log.Print(out)
}