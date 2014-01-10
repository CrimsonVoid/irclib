package module

import (
	"log"
	"regexp"
	"strconv"
	"strings"
)

func init() {
	// Use log to print to stdout because it is thread safe
	log.SetFlags(0)
}

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
		var color string
		if self.Enabled() {
			color = "\x1b[32m" // Green
		} else {
			color = "\x1b[31m" // Red
		}

		strOut := ""
		if self.String != nil {
			strOut = "\n\t" + self.String() + "\n"
		}

		unAU, alwUsr := self.GetROAllowed(User)
		unAC, alwChn := self.GetROAllowed(Chan)
		unDU, dnyUsr := self.GetRODenyed(User)
		unDC, dnyChn := self.GetRODenyed(Chan)

		log.Printf("%v%v\x1b[0m\n\t%v\n%v"+
			"\n\tIRC Commands\n\t\t%v"+
			"\n\tConsole Commands\n\t\t%v\n\n"+

			"\tAllowed Users: %v\n"+
			"\tBlocked Users: %v\n\n"+

			"\tAllowed Chans: %v\n"+
			"\tBlocked Chans: %v\n",

			color, self.Name(), self.Description(),
			strOut,
			strings.Join(self.StringCommands(), "\n\t\t"),
			strings.Join(self.Console.String(), "\n\t\t"),
			alwUsr, dnyUsr, alwChn, dnyChn)

		// Unlock
		unAU <- true
		unAC <- true
		unDU <- true
		unDC <- true
	})

	return err
}

// Allow or deny 'nick'
func (self *Module) registerAdd() error {
	re := regexp.MustCompile(`^(?i)(?P<mode>allow|deny)\s(?P<nick>.*)$`)

	err := self.Console.RegisterRegexp(re, func(s string) {
		s = strings.ToLower(s)

		// Can ignore error since match is already guaranteed
		groups, _ := matchGroups(re, s)
		nick := groups["nick"]
		var (
			err    error
			errMsg string
			msg    string
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
			log.Printf("Error %v %v: %v\n", errMsg, nick, err)

			return
		}

		self.Logger.Infoln(msg, nick)
		log.Println(msg, nick)
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
			log.Printf("Error removing %v from %v: %v\n", nick, errMsg, err)

			return
		}

		self.Logger.Infoln("Removed", nick)
		log.Println("Removed", nick)
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
		msg := ""

		switch groups["mode"] {
		case "allow":
			if groups["type"] == "user" {
				self.ClearAllowed(User)
				msg = "allowUser"
			} else { // "chan"
				self.ClearAllowed(Chan)
				msg = "allowChan"
			}
		default: // case "deny":
			if groups["type"] == "user" {
				self.ClearDenyed(User)
				msg = "denyUser"
			} else { // "chan"
				self.ClearDenyed(Chan)
				msg = "denyChan"
			}
		}

		self.Logger.Infof("Cleared %v list\n", msg)
		log.Printf("Cleared %v list\n", msg)
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
			msg    string
		)

		switch groups["mode"] {
		case "allow":
			if groups["type"] == "user" {
				unlock, list = self.GetROAllowed(User)
				msg = "Allowed users:"
			} else { // "chan"
				unlock, list = self.GetRODenyed(Chan)
				msg = "Allowed channels:"
			}
		default: // case "deny":
			if groups["type"] == "user" {
				unlock, list = self.GetRODenyed(User)
				msg = "Denyed users:"
			} else { // "chan"
				unlock, list = self.GetRODenyed(Chan)
				msg = "Denyed channels:"
			}
		}

		log.Println(msg, list)
		unlock <- true
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
		log.Println(status, self.Name())
	})

	return err
}

// Show last 10 logs
func (self *Module) registerLogs() error {
	err := self.Console.Register("logs", func(s string) {
		s = strings.ToLower(s)

		logs := self.Logger.TailLogs(10)

		log.Printf("%v\nShowing %v of %v logs\n",
			strings.Join(logs, "\n"),
			len(logs),
			self.Logger.LenLogs())
	})

	return err
}

// Print logs
func (self *Module) registerLogs2() error {
	re := regexp.MustCompile(`^(?i)(?P<cmd>head|tail)(\s(?P<num>-?\d+))?$`)

	err := self.Console.RegisterRegexp(re, func(s string) {
		s = strings.ToLower(s)

		groups, _ := matchGroups(re, s)

		var num int
		var err error
		if groups["num"] == "" {
			num = 10
		} else {
			num, err = strconv.Atoi(groups["num"])

			if err != nil {
				self.Logger.Errorln("Module.registerLogs():", err.Error())
				log.Println("Module.registerLogs(): ", err.Error())

				return
			}
		}

		switch groups["cmd"] {
		case "head":
			log.Println(strings.Join(self.Logger.Logs(num), "\n"))
		default: // case "tail":
			log.Println(strings.Join(self.Logger.TailLogs(num), "\n"))
		}
	})

	return err
}

// Clear logs
func (self *Module) registerClearLogs() error {
	err := self.Console.Register("clear logs", func(s string) {
		self.Logger.ClearLogs()
		log.Println("Logs cleared")
	})

	return err
}
