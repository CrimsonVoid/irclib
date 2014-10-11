package module

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
	"sync"
)

type st struct {
	trigger string
	fn      func(string)
}

type reCon struct {
	trigger *regexp.Regexp
	fn      func(string)
}

// Console struct to manage and parse commands
type Console struct {
	stTriggers []*st
	stMut      sync.RWMutex

	reTriggers []*reCon
	reMut      sync.RWMutex
}

func newConsole() *Console {
	return &Console{
		stTriggers: make([]*st, 0, 5),
		reTriggers: make([]*reCon, 0, 5),
	}
}

// Register a `string` or `regexp.Regexp` returning an error if `trigger` is already registered
func (self *Console) Register(trigger interface{}, fn func(string)) error {
	switch t := trigger.(type) {
	case string:
		return self.registerString(t, fn)
	case *regexp.Regexp:
		return self.registerRegexp(t, fn)
	case regexp.Regexp:
		return self.registerRegexp(&t, fn)
	default:
		return errors.New("Need a string or regexp.Regexp")
	}
}

// Register console commands; strings are lowered. Registered with "command" but triggered
// as ":moduleName command". Returns an error if 'trigger' is already registered
func (self *Console) registerString(trigger string, fn func(string)) error {
	trigger = strings.ToLower(trigger)

	self.stMut.Lock()
	defer self.stMut.Unlock()

	for _, v := range self.stTriggers {
		if v.trigger == trigger {
			return fmt.Errorf("Console.Register(): %v is already registered", trigger)
		}
	}

	self.stTriggers = append(self.stTriggers, &st{trigger, fn})

	return nil
}

// Register console commands. Registered with "command" but triggered
// as ":moduleName command". Returns an error if trigger.String() is already registered
func (self *Console) registerRegexp(trigger *regexp.Regexp, fn func(string)) error {
	self.reMut.Lock()
	defer self.reMut.Unlock()

	rStr := trigger.String()
	for _, v := range self.reTriggers {
		if v.trigger.String() == rStr {
			return fmt.Errorf("Console.RegisterRegexp(): %v is already registered", trigger)
		}
	}

	self.reTriggers = append(self.reTriggers, &reCon{trigger, fn})

	return nil
}

// Unregister a `string` or `regexp.Regexp` returning an error if `trigger` was not registered
func (self *Console) Unregister(trigger interface{}) error {
	switch t := trigger.(type) {
	case string:
		return self.unregisterString(t)
	case *regexp.Regexp:
		return self.unregisterRegexp(t)
	case regexp.Regexp:
		return self.unregisterRegexp(&t)
	default:
		return errors.New("Need a string or regexp.Regexp")
	}
}

// Unregister trigger. Return an error if trigger is not registered
func (self *Console) unregisterString(trigger string) error {
	trigger = strings.ToLower(trigger)

	self.stMut.Lock()
	defer self.stMut.Unlock()

	for i, v := range self.stTriggers {
		if v.trigger != trigger {
			continue
		}

		trigLen := len(self.stTriggers) - 1
		self.stTriggers[i] = self.stTriggers[trigLen]
		self.stTriggers[trigLen] = nil
		self.stTriggers = self.stTriggers[:trigLen]

		return nil
	}

	return fmt.Errorf("Console.Unregister(): %v is not registered", trigger)
}

// Unregister trigger. Return an error if trigger.String() is not registered
func (self *Console) unregisterRegexp(trigger *regexp.Regexp) error {
	self.reMut.Lock()
	defer self.reMut.Unlock()

	rStr := trigger.String()
	for i, v := range self.reTriggers {
		if v.trigger.String() != rStr {
			continue
		}

		trigLen := len(self.reTriggers) - 1
		self.reTriggers[i] = self.reTriggers[trigLen]
		self.reTriggers[trigLen] = nil
		self.reTriggers = self.reTriggers[:trigLen]

		return nil
	}

	return fmt.Errorf("Console.UnregisterRegexp(): %v is not registered", trigger.String())
}

// Parse cmd and run associated functions. 'cmd' is lowered before parsing string
// triggers. String and regexp functions spawn goroutines
func (self *Console) Parse(cmd string) {
	go self.parseString(cmd)
	go self.parseRegexp(cmd)
}

// Parse strings and if matched call associated function
func (self *Console) parseString(cmd string) {
	cmd = strings.ToLower(cmd)

	self.stMut.RLock()
	defer self.stMut.RUnlock()

	for _, v := range self.stTriggers {
		if v.trigger == cmd {
			v.fn(cmd)

			return
		}
	}
}

// Parse strings and if matched call associated function in it's own goroutine
func (self *Console) parseRegexp(cmd string) {
	self.reMut.RLock()
	defer self.reMut.RUnlock()

	for _, v := range self.reTriggers {
		if v.trigger.FindStringSubmatch(cmd) != nil {
			go v.fn(cmd)
		}
	}
}

// Returns a slice of strings of all string and regexp commands registered
func (self *Console) String() []string {
	naiveLen := len(self.stTriggers) + len(self.reTriggers)
	output := make([]string, 0, naiveLen)

	self.stMut.RLock()
	for _, v := range self.stTriggers {
		output = append(output, fmt.Sprintf("%v", v.trigger))
	}
	self.stMut.RUnlock()

	self.reMut.RLock()
	for _, v := range self.reTriggers {
		output = append(output, fmt.Sprintf("%v", v.trigger))
	}
	self.reMut.RUnlock()

	return output
}
