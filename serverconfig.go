package irclib

import (
	"errors"
	"fmt"
	"time"

	irc "github.com/fluffle/goirc/client"
)

type BotInfo struct {
	Chans  []string
	Access access
}

type Network struct {
	SSL      bool
	Server   string
	Port     int
	PingFreq int
	SplitLen int
	Flood    bool
	Tracking bool
}

type Groups struct {
	Users []string
}

type ServerInfo struct {
	Nick, Ident, Name string
	Pass              string
	Channels          []string
	Version           string
	QuitMessage       string

	Network Network
	Access  map[string]Groups
}

func (serverInfo *ServerInfo) configServer() (*irc.Config, error) {
	// Check there is enough info to set up a server
	switch {
	case serverInfo.Nick == "":
		return nil, errors.New("Specify a Nick in the config file")
	case serverInfo.Network.Server == "":
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
	cfg.SSL = serverInfo.Network.SSL
	cfg.Server = fmt.Sprintf("%v:%v",
		serverInfo.Network.Server,
		serverInfo.Network.Port)

	if serverInfo.Version != "" {
		cfg.Version = serverInfo.Version
	}
	if serverInfo.QuitMessage != "" {
		cfg.QuitMessage = serverInfo.QuitMessage
	}
	if serverInfo.Network.PingFreq > 0 {
		cfg.PingFreq = time.Duration(serverInfo.Network.PingFreq) * time.Second
	}
	if serverInfo.Network.SplitLen > 0 {
		cfg.SplitLen = serverInfo.Network.SplitLen
	}
	// Default value of bool is false which is default for cfg.Flood
	cfg.Flood = serverInfo.Network.Flood

	return cfg, nil
}
