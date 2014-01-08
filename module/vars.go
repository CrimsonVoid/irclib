package module

import (
	"sync"
)

type Event string

// IRC events to trigger on
const (
	REGISTER     Event = "REGISTER"
	CONNECTED          = "CONNECTED"
	DISCONNECTED       = "DISCONNECTED"
	ACTION             = "ACTION"
	AWAY               = "AWAY"
	CTCP               = "CTCP"
	CTCPREPLY          = "CTCPREPLY"
	INVITE             = "INVITE"
	JOIN               = "JOIN"
	KICK               = "KICK"
	MODE               = "MODE"
	NICK               = "NICK"
	NOTICE             = "NOTICE"
	OPER               = "OPER"
	PART               = "PART"
	PASS               = "PASS"
	PING               = "PING"
	PONG               = "PONG"
	PRIVMSG            = "PRIVMSG"
	QUIT               = "QUIT"
	TOPIC              = "TOPIC"
	USER               = "USER"
	VERSION            = "VERSION"
	VHOST              = "VHOST"
	WHO                = "WHO"
	WHOIS              = "WHOIS"
)

// Events is a slice of Events which are registered when irclibrary.Connect()
// is called. This is export primarily for the use of irclibrary and should not
// need to be modified by the user
var (
	Events    = make([]Event, 0, 5)
	eventsMut sync.RWMutex
)

var (
	logDir string // Module specific log directory
)
