package irclib

import (
	"github.com/crimsonvoid/irclib/module"
	irc "github.com/fluffle/goirc/client"
)

func (self *ModManager) setupHandlers() {
	config := self.Conn.Config()

	// Identify to NickServ and join channels
	self.Conn.HandleFunc(irc.CONNECTED, func(con *irc.Conn, line *irc.Line) {
		con.Privmsg("NickServ", "IDENTIFY "+config.Pass)

		self.mut.RLock()
		defer self.mut.RLock()

		for _, ch := range self.Config.Chans {
			con.Join(ch)
		}

	})

	// Iterate over EventList and register functions
	events := module.RegisteredEvents()
	for i := range events {
		event := string(events[i])

		self.Conn.HandleFunc(event, func(con *irc.Conn, line *irc.Line) {
			go self.run(event, line)
		})
	}
}

func (self *ModManager) run(event string, line *irc.Line) {
	self.mut.RLock()
	defer self.mut.RUnlock()

	for _, mod := range self.modules {
		// Module should check if enabled, not handlers
		go mod.Handle(line.Text(), module.Event(event), line)
	}
}
