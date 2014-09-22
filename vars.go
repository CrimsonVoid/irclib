package irclib

import (
	"log"
	"os"
)

var (
	consLog = log.New(os.Stdout, "", 0)
)
