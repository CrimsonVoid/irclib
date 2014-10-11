package module

import (
	"fmt"
	"sync"
)

type userChan bool

const (
	UC_User userChan = true
	UC_Chan          = false
)

// Allow user or channels; channels should begin with '#'. Returns an error
// if 'target' is already in the slice. Does not remove user or channel from
// allow(User|Chan) slices
func (self *moduleConfig) Allow(target string) error {
	if len(target) == 0 {
		return fmt.Errorf("moduleConfig.Allow(): Length of string is 0")
	}

	if target[0] == '#' {
		self.mu.Lock()
		defer self.mu.Unlock()

		return add(&self.m.AllowChan, target)
	} else {
		self.mu.Lock()
		defer self.mu.Unlock()

		return add(&self.m.AllowUser, target)
	}
}

// Deny user or channels; channels should begin with '#'. Returns an error
// if 'target' is already in the slice. Does not remove user or channel from
// deny(User|Chan) slices
func (self *moduleConfig) Deny(target string) error {
	if len(target) == 0 {
		return fmt.Errorf("moduleConfig.Deny(): Length of string is 0")
	}

	if target[0] == '#' {
		self.mu.Lock()
		defer self.mu.Unlock()

		return add(&self.m.DenyChan, target)
	} else {
		self.mu.Lock()
		defer self.mu.Unlock()

		return add(&self.m.DenyUser, target)
	}
}

// Removes user or channel from respective allowed slices; channels should begin
// with '#'. Returns an error if 'target' is not in the slice
func (self *moduleConfig) RemAllowed(target string) error {
	if len(target) == 0 {
		return fmt.Errorf("moduleConfig.RemAllowed(): Length of string is 0")
	}

	if target[0] == '#' {
		self.mu.Lock()
		defer self.mu.Unlock()

		return remove(&self.m.AllowChan, target)
	} else {
		self.mu.Lock()
		defer self.mu.Unlock()

		return remove(&self.m.AllowUser, target)
	}
}

// Removes user or channel from respective denyed slices; channels should begin
// with '#'. Returns an error if 'target' is not in the slice
func (self *moduleConfig) RemDenyed(target string) error {
	if len(target) == 0 {
		return fmt.Errorf("moduleConfig.RemDenyed(): Length of string is 0")
	}

	if target[0] == '#' {
		self.mu.Lock()
		defer self.mu.Unlock()

		return remove(&self.m.DenyChan, target)
	} else {
		self.mu.Lock()
		defer self.mu.Unlock()

		return remove(&self.m.DenyUser, target)
	}
}

// Returns a copy (not slice header) of allow(User|Chan) slices. If copies are
// too expensive look at GetROAllowed()
func (self *moduleConfig) GetAllowed(u userChan) []string {
	switch u {
	case UC_User:
		self.mu.RLock()
		defer self.mu.RUnlock()

		return copySlice(self.m.AllowUser)
	default: // case UC_Chan:
		self.mu.RLock()
		defer self.mu.RUnlock()

		return copySlice(self.m.AllowChan)
	}
}

// Returns a copy (not slice header) of deny(User|Chan) slices. If copies are
// too expensive look at GetRODenyed()
func (self *moduleConfig) GetDenyed(u userChan) []string {
	switch u {
	case UC_User:
		self.mu.RLock()
		defer self.mu.RUnlock()

		return copySlice(self.m.DenyUser)
	default: // case UC_Chan:
		self.mu.RLock()
		defer self.mu.RUnlock()

		return copySlice(self.m.DenyChan)
	}
}

// Returns a copy of the appropriate slice header to avoid copies. This is a
// promise by the user to not modify the slice and enforced by a RLock() on the
// slice. The lock is released by closing the chan or sending a boolean value
func (self *moduleConfig) GetROAllowed(u userChan) ([]string, chan<- bool) {
	switch u {
	case UC_User:
		return self.m.AllowUser, roLock(&self.mu)
	default: // case UC_Chan:
		return self.m.AllowChan, roLock(&self.mu)
	}
}

// Returns a copy of the appropriate slice header to avoid copies. This is a
// promise by the user to not modify the slice and enforced by a RLock() on the
// slice. The lock is released by closing the chan or sending a boolean value
func (self *moduleConfig) GetRODenyed(u userChan) ([]string, chan<- bool) {
	switch u {
	case UC_User:
		return self.m.DenyUser, roLock(&self.mu)
	default: // case UC_Chan:
		return self.m.DenyChan, roLock(&self.mu)
	}
}

// Clears allow(User|Chan) slices
func (self *moduleConfig) ClearAllowed(u userChan) {
	switch u {
	case UC_User:
		self.mu.Lock()
		defer self.mu.Unlock()

		clear(&self.m.AllowUser)
	default: // case UC_Chan:
		self.mu.Lock()
		defer self.mu.Unlock()

		clear(&self.m.AllowChan)
	}
}

// Clears deny(User|Chan) slices
func (self *moduleConfig) ClearDenyed(u userChan) {
	switch u {
	case UC_User:
		self.mu.Lock()
		defer self.mu.Unlock()

		clear(&self.m.DenyUser)
	default: // case UC_Chan:
		self.mu.Lock()
		defer self.mu.Unlock()

		clear(&self.m.DenyChan)
	}
}

// Returns true if 'target' is in allow(User|Chan); chans should begin with '#'
func (self *moduleConfig) InAllowed(target string) bool {
	if len(target) == 0 {
		return false
	}

	if target[0] == '#' {
		self.mu.RLock()
		defer self.mu.RUnlock()

		return inSlice(self.m.AllowChan, target)
	} else {
		self.mu.RLock()
		defer self.mu.RUnlock()

		return inSlice(self.m.AllowUser, target)
	}
}

// Returns true if 'target' is in deny(User|Chan); chans should begin with '#'
func (self *moduleConfig) InDenyed(target string) bool {
	if len(target) == 0 {
		return false
	}

	if target[0] == '#' {
		self.mu.RLock()
		defer self.mu.RUnlock()

		return inSlice(self.m.DenyChan, target)
	} else {
		self.mu.RLock()
		defer self.mu.RUnlock()

		return inSlice(self.m.DenyUser, target)
	}
}

// Returns the length of allow(User|Chan)
func (self *moduleConfig) LenAllowed(u userChan) int {
	switch u {
	case UC_User:
		self.mu.RLock()
		defer self.mu.RUnlock()

		return len(self.m.AllowUser)
	default: // case UC_Chan:
		self.mu.Lock()
		defer self.mu.Unlock()

		return len(self.m.AllowChan)
	}
}

// Returns the length of allow(User|Chan)
func (self *moduleConfig) LenDenyed(u userChan) int {
	switch u {
	case UC_User:
		self.mu.RLock()
		defer self.mu.RUnlock()

		return len(self.m.DenyUser)
	default: // case UC_Chan:
		self.mu.Lock()
		defer self.mu.Unlock()

		return len(self.m.DenyChan)
	}
}

func roLock(mu *sync.RWMutex) chan<- bool {
	unlock := make(chan bool)
	mu.RLock()

	go func() {
		<-unlock

		mu.RUnlock()
	}()

	return unlock
}
