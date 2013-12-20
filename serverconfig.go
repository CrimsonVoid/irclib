package irclibrary

import (
	"errors"
	"fmt"
	"time"

	irc "github.com/fluffle/goirc/client"
)

type ServerInfo struct {
	Nick, Ident, Name string

	Pass, Server string
	SSL          bool
	Port         int

	Version, QuitMessage string
	PingFreq, SplitLen   int
	Flood, Tracking      bool

	Botinfo BotInfo
}

type BotInfo struct {
	Chans  []string
	Access access
}

func (serverInfo *ServerInfo) configServer() (*irc.Config, error) {
	// Check there is enough info to set up a server
	switch {
	case serverInfo.Nick == "":
		return nil, errors.New("Specify a Nick in the config file")
	case serverInfo.Server == "":
		return nil, errors.New("Specify a Server in the config file")
	}

	if serverInfo.Ident == "" {
		serverInfo.Ident = serverInfo.Nick
	}
	if serverInfo.Name == "" {
		serverInfo.Name = serverInfo.Nick
	}

	cfg := irc.NewConfig(serverInfo.Nick, serverInfo.Ident, serverInfo.Name)

	cfg.Pass = serverInfo.Pass
	// if serverInfo.SSL is true, user will have to configure it themself
	cfg.SSL = serverInfo.SSL
	cfg.Server = fmt.Sprintf("%v:%v", serverInfo.Server, serverInfo.Port)

	if serverInfo.Version != "" {
		cfg.Version = serverInfo.Version
	}
	if serverInfo.QuitMessage != "" {
		cfg.QuitMessage = serverInfo.QuitMessage
	}
	if serverInfo.PingFreq > 0 {
		cfg.PingFreq = time.Duration(serverInfo.PingFreq) * time.Second
	}
	if serverInfo.SplitLen > 0 {
		cfg.SplitLen = serverInfo.SplitLen
	}
	// Default value of bool is false which is default for cfg.Flood
	cfg.Flood = serverInfo.Flood

	return cfg, nil
}
