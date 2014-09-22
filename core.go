package irclib

import (
	"fmt"
	"regexp"

	"github.com/crimsonvoid/console/styles"
	"github.com/crimsonvoid/irclib/module"
)

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

	if err := core.PreStart(); err != nil {
		panic(err)
	}

	if err := core.Start(); err != nil {
		panic(err)
	}

	return core
}

func (self *ModManager) registerCoreCommands() {
	self.core.Conn = self.Conn

	errFns := []func() error{
		// Quit
		self.regCoreQuit,
		// Force quit, optional module name
		self.regCoreForceQuit,
		// List modules
		self.regCoreListModules,
		// Join or Part channels
		self.regCoreChanManage,
		// Print access list
		self.regCoreAccessList,
		// Add or remove nicks from access list
		self.regCoreAccessManip,
	}

	for _, fn := range errFns {
		if err := fn(); err != nil {
			panic(err)
		}
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
	err := self.core.Console.Register(re, func(trigger string) {
		self.coreForceDisconnect(trigger)
	})

	return err
}

func (self *ModManager) regCoreListModules() error {
	err := self.core.Console.Register("list", func(trigger string) {
		self.mut.RLock()
		defer self.mut.RUnlock()

		msg := ""
		for _, mod := range self.modules {
			color := styles.Green
			if !mod.Enabled() {
				color = styles.Red
			}

			msg += fmt.Sprintf("%v - %v\n",
				color.Fg("%v", mod.Name()),
				mod.Description())
		}

		consLog.Print(msg)
	})

	return err
}

func (self *ModManager) regCoreChanManage() error {
	re := regexp.MustCompile(`^(?P<cmd>join|part)\s(?P<chan>.*)$`)
	err := self.core.Console.Register(re, func(trigger string) {
		groups, _ := matchGroups(re, trigger)
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
			consLog.Printf("%v\n  %v\n", grp, nicks)
		}
	})

	return err
}

func (self *ModManager) regCoreAccessManip() error {
	re := regexp.MustCompile(`^access\s(?P<cmd>add|rem)\s(?P<group>.*)\s(?P<nick>.*)$`)
	err := self.core.Console.Register(re, func(trigger string) {
		groups, _ := matchGroups(re, trigger)
		msg := ""

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

		consLog.Printf(msg, groups["nick"], groups["group"])
		self.core.Logger.Infof(msg, groups["nick"], groups["group"])
	})

	return err
}

func (self *ModManager) coreDisconnect() {
	errors := self.Disconnect()

	if len(errors) == 0 {
		consLog.Println(styles.Green.Fg("Disconnected without errors"))

		self.Quit <- true

		return
	}

	out := styles.Red.Fg("Errors when attempting to disconnect\n")
	for modName, err := range errors {
		out += fmt.Sprintf("  %v: %v\n", modName, err)
	}
	consLog.Print(out)

	self.Quit <- false
}

func (self *ModManager) coreForceDisconnect(trigger string) {
	defer func() {
		self.Quit <- true
	}()

	re := regexp.MustCompile(`^f(orce\s)?quit\s(?P<module>.*)?$`)
	errMap := make(map[string][]error)
	groups, _ := matchGroups(re, trigger)

	if modName, ok := groups["module"]; ok {
		errs := self.ForceDisconnectModule(modName)
		if len(errs) == 0 {
			consLog.Println(styles.Green.Fg("Force disconnected module %v without errors",
				modName))

			return
		}

		errMap[modName] = errs
	} else if errMap = self.ForceDisconnect(); len(errMap) == 0 {
		consLog.Println(styles.Green.Fg("Force disconnected without errors"))

		return
	}

	out := styles.Red.Fg("Errors when attempting to force disconnect %v\n",
		groups["module"])

	for modName, errs := range errMap {
		errStr := fmt.Sprintf("  %v", modName)
		for _, err := range errs {
			errStr += fmt.Sprintf("    %v\n", err)
		}

		out += errStr
	}
	consLog.Print(out)
}
