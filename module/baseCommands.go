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
	if err := self.registerInfo(); err != nil {
		self.Logger.Errorln(err)
	}
	if err := self.registerAdd(); err != nil {
		self.Logger.Errorln(err)
	}
	if err := self.registerRem(); err != nil {
		self.Logger.Errorln(err)
	}
	if err := self.registerClear(); err != nil {
		self.Logger.Errorln(err)
	}
	if err := self.registerList(); err != nil {
		self.Logger.Errorln(err)
	}
	if err := self.registerEnable(); err != nil {
		self.Logger.Errorln(err)
	}
	if err := self.registerLogs(); err != nil {
		self.Logger.Errorln(err)
	}
	if err := self.registerLogs2(); err != nil {
		self.Logger.Errorln(err)
	}
	if err := self.registerClearLogs(); err != nil {
		self.Logger.Errorln(err)
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

		unAU, allUsr := self.GetROAllowed(User)
		unAC, allChn := self.GetROAllowed(Chan)
		unDU, denUsr := self.GetRODenyed(User)
		unDC, denChn := self.GetRODenyed(Chan)

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
			*allUsr, *denUsr, *allChn, *denChn)

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

		if groups["mode"] == "allow" {
			if err := self.Allow(nick); err != nil {
				self.Logger.Errorln("Module.registerAdd()", err.Error())
				log.Printf("Error allowing %v: %v\n", nick, err)
			} else {
				self.Logger.Infoln("Allowed", nick)
				log.Println("Allowed", nick)
			}
		} else { // "deny"
			if err := self.Deny(nick); err != nil {
				self.Logger.Errorln("Module.registerAdd()", err.Error())
				log.Printf("Error blocking %v: %v\n", nick, err)
			} else {
				self.Logger.Infoln("Blocked", nick)
				log.Println("Blocked", nick)
			}
		}
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

		if groups["mode"] == "allow" {
			if err := self.RemAllowed(nick); err != nil {
				self.Logger.Errorln("Module.registerRem()", err.Error())
				log.Printf("Error removing %v from allowed: %v\n", nick, err)
			} else {
				self.Logger.Infoln("Removed", nick)
				log.Println("Removed", nick)
			}
		} else { // "deny"
			if err := self.RemDenyed(nick); err != nil {
				self.Logger.Errorln("Module.registerRem()", err.Error())
				log.Printf("Error removing %v from denied: %v\n", nick, err)
			} else {
				self.Logger.Infoln("Removed", nick)
				log.Println("Removed", nick)
			}
		}
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

		if groups["mode"] == "allow" {
			if groups["type"] == "user" {
				self.ClearAllowed(User)
				self.Logger.Infoln("Cleared allowUser list")
				log.Println("Cleared allowUser list")
			} else { // "chan"
				self.ClearAllowed(Chan)
				self.Logger.Infoln("Cleared allowChan list")
				log.Println("Cleared allowChan list")
			}
		} else { // "deny"
			if groups["type"] == "user" {
				self.ClearDenyed(User)
				self.Logger.Infoln("Cleared denyUser list")
				log.Println("Cleared denyUser list")
			} else { // "chan"
				self.ClearDenyed(Chan)
				self.Logger.Infoln("Cleared denyChan list")
				log.Println("Cleared denyChan list")
			}
		}
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

		if groups["mode"] == "allow" {
			if groups["type"] == "user" {
				unlock, list := self.GetROAllowed(User)
				log.Println("Allowed users:", *list)

				unlock <- true
			} else { // "chan"
				unlock, list := self.GetRODenyed(Chan)
				log.Println("Allowed channels:", *list)

				unlock <- true
			}
		} else { // "deny"
			if groups["type"] == "user" {
				unlock, list := self.GetRODenyed(User)

				log.Println("Denyed users:", *list)
				unlock <- true
			} else { // "chan"
				unlock, list := self.GetRODenyed(Chan)

				log.Println("Denyed channels:", *list)
				unlock <- true
			}
		}
	})

	return err
}

// Enable or disable module
func (self *Module) registerEnable() error {
	re := regexp.MustCompile(`^(?i)(?P<cmd>en|dis)able$`)

	err := self.Console.RegisterRegexp(re, func(s string) {
		s = strings.ToLower(s)

		groups, _ := matchGroups(re, s)

		switch groups["cmd"] {
		case "en":
			self.Enable()
			self.Logger.Infoln("Enabled", self.Name())
			log.Println("Enabled", self.Name())
		case "dis":
			self.Disable()
			self.Logger.Infoln("Disabled", self.Name())
			log.Println("Disabled", self.Name())
		}
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
