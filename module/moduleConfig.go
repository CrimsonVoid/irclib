package module

import (
	"sync"
)

// ModuleInfo repersents configuration fields that can be loaded from TOML files
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
	m  ModuleInfo
	mu sync.RWMutex
}

// Returns the module name
func (self *moduleConfig) Name() string {
	self.mu.RLock()
	defer self.mu.RUnlock()

	return self.m.Name
}

// Returns the module description
func (self *moduleConfig) Description() string {
	self.mu.RLock()
	defer self.mu.RUnlock()

	return self.m.Description
}

// Set module description
func (self *moduleConfig) SetDescription(desc string) {
	self.mu.Lock()
	defer self.mu.Unlock()

	self.m.Description = desc
}

// Returns `true` if the module is enabled
func (self *moduleConfig) Enabled() bool {
	self.mu.RLock()
	defer self.mu.RUnlock()

	return self.m.Enabled
}

// Set the status of the module
func (self *moduleConfig) SetEnabled(en bool) {
	self.mu.Lock()
	defer self.mu.Unlock()

	self.m.Enabled = en
}

// Helper function to enable the module. Calls `moduleConfig.SetEnabled(true)`
func (self *moduleConfig) Enable() {
	self.SetEnabled(true)
}

// Helper function to disable the module. Calls `moduleConfig.SetEnabled(false)`
func (self *moduleConfig) Disable() {
	self.SetEnabled(false)
}

// Returns the directory logs are saved to. Full log path is "moduleConfig.logDir/moduleConfig.name"
func (self *moduleConfig) LogDir() string {
	self.mu.RLock()
	defer self.mu.RUnlock()

	return self.m.LogDir
}

// Sets the directory logs are saved to. This does not take effect until the module is restarted
func (self *moduleConfig) SetLogDir(logDir string) {
	self.mu.Lock()
	defer self.mu.Unlock()

	self.m.LogDir = logDir
}
