package module

import (
	"sync"
)

// ModuleInfo repersents configuration fields that can be loaded from JSON files
type ModuleInfo struct {
	Name        string // Unique name or registering and triggering console commands
	Description string // Description of module
	LogDir      string // Directory to keep logs, defaults to ./logs/
	Enabled     bool   // Flag to see if module is enabled

	// Filtered by: denyUser, allowUser, denyChan, allowChan
	// ToLower is called on slices when creating a Module
	AllowUser, DenyUser []string // Slice of allowed or denyed users
	AllowChan, DenyChan []string // Slice of allowed or denyed chans
}

// A copy of ModuleInfo but fields are not exported
type moduleConfig struct {
	name, description string
	logDir            string
	enabled           bool
	mu                sync.RWMutex

	// Filtered by: denyUser, allowUser, denyChan, allowChan
	allowUser, denyUser []string
	allowChan, denyChan []string
	userMut, chanMut    sync.RWMutex
}

func (self *moduleConfig) Name() string {
	self.mu.RLock()
	defer self.mu.RUnlock()

	name := self.name
	return name
}

func (self *moduleConfig) SetName(name string) {
	self.mu.Lock()
	defer self.mu.Unlock()

	self.name = name
}

func (self *moduleConfig) Description() string {
	self.mu.RLock()
	defer self.mu.RUnlock()

	desc := self.description
	return desc
}

func (self *moduleConfig) SetDescription(desc string) {
	self.mu.Lock()
	defer self.mu.Unlock()

	self.description = desc
}

func (self *moduleConfig) Enabled() bool {
	self.mu.RLock()
	defer self.mu.RUnlock()

	enabled := self.enabled
	return enabled
}

func (self *moduleConfig) SetEnabled(en bool) {
	self.mu.Lock()
	defer self.mu.Unlock()

	self.enabled = en
}

func (self *moduleConfig) Enable() {
	self.SetEnabled(true)
}

func (self *moduleConfig) Disable() {
	self.SetEnabled(false)
}

func (self *moduleConfig) LogDir() string {
	self.mu.RLock()
	defer self.mu.RUnlock()

	logDir := self.logDir
	return logDir
}

func (self *moduleConfig) SetLogDir(logDir string) {
	self.mu.Lock()
	defer self.mu.Unlock()

	self.logDir = logDir
}
