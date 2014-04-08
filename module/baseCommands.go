package module

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/crimsonvoid/console"
)

// Register commands, logs errors, and continues
func (self *Module) registerBaseCommands() {
	registerErrors := []error{
		self.registerInfo(),
		self.registerAdd(),
		self.registerRem(),
		self.registerClear(),
		self.registerList(),
		self.registerEnable(),
		self.registerLogs(),
		self.registerLogs2(),
		self.registerClearLogs(),
	}

	for _, err := range registerErrors {
		if err != nil {
			self.Logger.Errorln(err)
		}
	}
}

// Print info about module. Triggerd with 'info'
func (self *Module) registerInfo() error {
	err := self.Console.Register("info", func(s string) {
		color := console.C_FgGreen
		if !self.Enabled() {
			color = console.C_FgRed
		}

		strOut := ""
		if self.String != nil {
			strOut = fmt.Sprintf("\n\t%v\n", self.String())
		}

		unAU, alwUsr := self.GetROAllowed(UC_User)
		unAC, alwChn := self.GetROAllowed(UC_Chan)
		unDU, dnyUsr := self.GetRODenyed(UC_User)
		unDC, dnyChn := self.GetRODenyed(UC_Chan)

		consLog.Printf("%v%v%v\n\t%v\n%v"+
			"\n\tIRC Commands\n\t\t%v"+
			"\n\tConsole Commands\n\t\t%v\n\n"+

			"\tAllowed Users: %v\n"+
			"\tBlocked Users: %v\n\n"+

			"\tAllowed Chans: %v\n"+
			"\tBlocked Chans: %v\n",

			color, console.C_Reset, self.Name(), self.Description(),
			strOut,
			strings.Join(self.StringCommands(), "\n\t\t"),
			strings.Join(self.Console.String(), "\n\t\t"),
			alwUsr, dnyUsr, alwChn, dnyChn,
		)

		// Release locks
		close(unAU)
		close(unAC)
		close(unDU)
		close(unDC)
	})

	return err
}

// Allow or deny 'nick'
func (self *Module) registerAdd() error {
	re := regexp.MustCompile(`^(?i)(?P<mode>allow|deny)\s(?P<nick>.*)$`)

	err := self.Console.RegisterRegexp(re, func(s string) {
		s = strings.ToLower(s)
		// Can ignore error since match is guaranteed
		groups, _ := matchGroups(re, s)
		nick := groups["nick"]

		var (
			err    error
			errMsg string
		)

		switch groups["mode"] {
		case "allow":
			err = self.Allow(nick)
			errMsg = "allowing"
		default: // case "deny":
			err = self.Deny(nick)
			errMsg = "blocking"
		}

		if err != nil {
			self.Logger.Errorln("Module.registerAdd()", err.Error())
			consLog.Printf("Error %v %v: %v\n", errMsg, nick, err)

			return
		}

		self.Logger.Infoln("Allowed", nick)
		consLog.Println("Allowed", nick)
	})

	return err
}

// Remove 'nick' from allow or deny slices
func (self *Module) registerRem() error {
	re := regexp.MustCompile(`^(?i)rem\s(?P<mode>allow|deny)\s(?P<nick>.*)$`)

	err := self.Console.RegisterRegexp(re, func(s string) {
		s = strings.ToLower(s)

		// Can ignore error since match is already guaranteed
		groups, _ := matchGroups(re, s)
		nick := groups["nick"]

		var (
			err    error
			errMsg string
		)

		switch groups["mode"] {
		case "allow":
			err = self.RemAllowed(nick)
			errMsg = "allowed"
		default: // case "deny":
			err = self.RemDenyed(nick)
			errMsg = "denied"
		}

		if err != nil {
			self.Logger.Errorln("Module.registerRem()", err.Error())
			consLog.Printf("Error removing %v from %v: %v\n", nick, errMsg, err)

			return
		}

		self.Logger.Infoln("Removed", nick)
		consLog.Println("Removed", nick)
	})

	return err
}

// Clear (allow|deny)(user|chan)
func (self *Module) registerClear() error {
	re := regexp.MustCompile(`^(?i)clear\s(?P<mode>allow|deny)(?P<type>user|chan)$`)

	err := self.Console.RegisterRegexp(re, func(s string) {
		s = strings.ToLower(s)

		// Can ignore error since match is already guaranteed
		groups, _ := matchGroups(re, s)

		clType, msg := UC_User, "User"
		if groups["type"] == "chan" {
			clType, msg = UC_Chan, "Chan"
		}

		switch groups["mode"] {
		case "allow":
			self.ClearAllowed(clType)
			msg = "allow" + msg
		default: // case "deny":
			self.ClearDenyed(clType)
			msg = "deny" + msg
		}

		self.Logger.Infof("Cleared %v list\n", msg)
		consLog.Printf("Cleared %v list\n", msg)
	})

	return err
}

// Print (allow|deny)(user|chan) slice
func (self *Module) registerList() error {
	re := regexp.MustCompile(`^(?i)list\s(?P<mode>allow|deny)(?P<type>user|chan)$`)

	err := self.Console.RegisterRegexp(re, func(s string) {
		s = strings.ToLower(s)

		// Can ignore error since match is already guaranteed
		groups, _ := matchGroups(re, s)
		var (
			unlock chan<- bool
			list   []string
		)

		lsType, msg := UC_User, "users: "
		if groups["type"] == "chan" {
			lsType, msg = UC_Chan, "channels: "
		}

		switch groups["mode"] {
		case "allow":
			unlock, list = self.GetROAllowed(lsType)
			msg = "Allowed " + msg
		default: // case "deny"
			unlock, list = self.GetRODenyed(lsType)
			msg = "Denyed " + msg
		}

		consLog.Println(msg, list)
		close(unlock)
	})

	return err
}

// Enable or disable module
func (self *Module) registerEnable() error {
	re := regexp.MustCompile(`^(?i)(?P<cmd>en|dis)able$`)

	err := self.Console.RegisterRegexp(re, func(s string) {
		s = strings.ToLower(s)

		groups, _ := matchGroups(re, s)
		status := ""

		switch groups["cmd"] {
		case "en":
			self.Enable()
			status = "Enabled"
		default: // case "dis":
			self.Disable()
			status = "Disabled"
		}

		self.Logger.Infoln(status, self.Name())
		consLog.Println(status, self.Name())
	})

	return err
}

// Show last 10 logs
func (self *Module) registerLogs() error {
	err := self.Console.Register("logs", func(s string) {
		s = strings.ToLower(s)
		logs := self.Logger.TailLogs(10)

		consLog.Printf("%v\nShowing %v of %v logs\n",
			strings.Join(logs, "\n"), len(logs), self.Logger.LenLogs(),
		)
	})

	return err
}

// Print logs
func (self *Module) registerLogs2() error {
	re := regexp.MustCompile(`^(?i)(?P<cmd>head|tail)(\s(?P<num>-?\d+))?$`)

	err := self.Console.RegisterRegexp(re, func(s string) {
		s = strings.ToLower(s)
		groups, _ := matchGroups(re, s)

		var (
			num int
			err error
		)

		if groups["num"] == "" {
			num = 10
		} else {
			num, err = strconv.Atoi(groups["num"])

			if err != nil {
				self.Logger.Errorln("Module.registerLogs():", err.Error())
				consLog.Println("Module.registerLogs(): ", err.Error())

				return
			}
		}

		switch groups["cmd"] {
		case "head":
			consLog.Println(strings.Join(self.Logger.Logs(num), "\n"))
		default: // case "tail":
			consLog.Println(strings.Join(self.Logger.TailLogs(num), "\n"))
		}
	})

	return err
}

// Clear logs
func (self *Module) registerClearLogs() error {
	err := self.Console.Register("clear logs", func(s string) {
		self.Logger.ClearLogs()
		consLog.Println("Logs cleared")
	})

	return err
}
