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
	E_CONNECTED          = "CONNECTED"
	E_DISCONNECTED       = "DISCONNECTED"
	E_ACTION             = "ACTION"
	E_AWAY               = "AWAY"
	E_CTCP               = "CTCP"
	E_CTCPREPLY          = "CTCPREPLY"
	E_INVITE             = "INVITE"
	E_JOIN               = "JOIN"
	E_KICK               = "KICK"
	E_MODE               = "MODE"
	E_NICK               = "NICK"
	E_NOTICE             = "NOTICE"
	E_OPER               = "OPER"
	E_PART               = "PART"
	E_PASS               = "PASS"
	E_PING               = "PING"
	E_PONG               = "PONG"
	E_PRIVMSG            = "PRIVMSG"
	E_QUIT               = "QUIT"
	E_TOPIC              = "TOPIC"
	E_USER               = "USER"
	E_VERSION            = "VERSION"
	E_VHOST              = "VHOST"
	E_WHO                = "WHO"
	E_WHOIS              = "WHOIS"
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
