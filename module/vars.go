package module

import (
	"log"
	"os"
	"sync"
)

type Event string

// IRC events to trigger on
const (
	E_REGISTER     Event = "REGISTER"
	E_CONNECTED    Event = "CONNECTED"
	E_DISCONNECTED Event = "DISCONNECTED"
	E_ACTION       Event = "ACTION"
	E_AWAY         Event = "AWAY"
	E_CTCP         Event = "CTCP"
	E_CTCPREPLY    Event = "CTCPREPLY"
	E_INVITE       Event = "INVITE"
	E_JOIN         Event = "JOIN"
	E_KICK         Event = "KICK"
	E_MODE         Event = "MODE"
	E_NICK         Event = "NICK"
	E_NOTICE       Event = "NOTICE"
	E_OPER         Event = "OPER"
	E_PART         Event = "PART"
	E_PASS         Event = "PASS"
	E_PING         Event = "PING"
	E_PONG         Event = "PONG"
	E_PRIVMSG      Event = "PRIVMSG"
	E_QUIT         Event = "QUIT"
	E_TOPIC        Event = "TOPIC"
	E_USER         Event = "USER"
	E_VERSION      Event = "VERSION"
	E_VHOST        Event = "VHOST"
	E_WHO          Event = "WHO"
	E_WHOIS        Event = "WHOIS"
)

// Events is a slice of Events which are registered when irclibrary.Connect()
// is called. This is export primarily for the use of irclibrary and should not
// need to be modified by the user
var (
	Events    = make([]Event, 0, 5)
	eventsMut sync.RWMutex
)

var (
	logDir  = "./logs/" // Module specific log directory
	consLog = log.New(os.Stdout, "", 0)
)
