package module

import (
	"errors"
	"sync"
)

type UserChan bool

const (
	User UserChan = true
	Chan          = false
)

// Allow user or channels; channels should begin with '#'. Returns an error
// if 'target' is already in the slice. Does not remove user or channel from
// allow(User|Chan) slices
func (self *moduleConfig) Allow(target string) error {
	if len(target) == 0 {
		return errors.New("moduleConfig.Allow(): Length of string " + target + " is 0")
	}

	if target[0] == '#' {
		self.chanMut.Lock()
		defer self.chanMut.Unlock()

		return add(&self.allowChan, target)
	} else {
		self.userMut.Lock()
		defer self.userMut.Unlock()

		return add(&self.allowUser, target)
	}
}

// Deny user or channels; channels should begin with '#'. Returns an error
// if 'target' is already in the slice. Does not remove user or channel from
// deny(User|Chan) slices
func (self *moduleConfig) Deny(target string) error {
	if len(target) == 0 {
		return errors.New("moduleConfig.Deny(): Length of string " + target + " is 0")
	}

	if target[0] == '#' {
		self.chanMut.Lock()
		defer self.chanMut.Unlock()

		return add(&self.denyChan, target)
	} else {
		self.userMut.Lock()
		defer self.userMut.Unlock()

		return add(&self.denyUser, target)
	}
}

// Removes user or channel from respective allowed slices; channels should begin
// with '#'. Returns an error if 'target' is not in the slice
func (self *moduleConfig) RemAllowed(target string) error {
	if len(target) == 0 {
		return errors.New("moduleConfig.RemAllowed(): Length of string " + target + " is 0")
	}

	if target[0] == '#' {
		self.chanMut.Lock()
		defer self.chanMut.Unlock()

		return remove(&self.allowChan, target)
	} else {
		self.userMut.Lock()
		defer self.userMut.Unlock()

		return remove(&self.allowUser, target)
	}
}

// Removes user or channel from respective denyed slices; channels should begin
// with '#'. Returns an error if 'target' is not in the slice
func (self *moduleConfig) RemDenyed(target string) error {
	if len(target) == 0 {
		return errors.New("moduleConfig.RemDenyed(): Length of string " + target + " is 0")
	}

	if target[0] == '#' {
		self.chanMut.Lock()
		defer self.chanMut.Unlock()

		return remove(&self.denyChan, target)
	} else {
		self.userMut.Lock()
		defer self.userMut.Unlock()

		return remove(&self.denyUser, target)
	}
}

// Returns a copy (not slice header) of allow(User|Chan) slices. If copies are
// too expensive look at GetROAllowed()
func (self *moduleConfig) GetAllowed(u UserChan) []string {
	switch u {
	case User:
		self.userMut.RLock()
		defer self.userMut.RUnlock()

		return copySlice(self.allowUser)
	case Chan:
		self.chanMut.RLock()
		defer self.chanMut.RUnlock()

		return copySlice(self.allowChan)
	// This should never trigger; only here to silence compiler errors
	default:
		return nil
	}
}

// Returns a copy (not slice header) of deny(User|Chan) slices. If copies are
// too expensive look at GetRODenyed()
func (self *moduleConfig) GetDenyed(u UserChan) []string {
	switch u {
	case User:
		self.userMut.RLock()
		defer self.userMut.RUnlock()

		return copySlice(self.denyUser)
	case Chan:
		self.chanMut.RLock()
		defer self.chanMut.RUnlock()

		return copySlice(self.denyChan)
	// This should never trigger; only here to silence compiler errors
	default:
		return nil
	}
}

// Returns a pointer of the slice header to avoid copies. This is a promise by
// the user to not modify the slice and enforced by a RLock() on the slice. The
// lock is released and slice pointer set to nil when a boolean value is sent on
// the returned channel
func (self *moduleConfig) GetROAllowed(u UserChan) (chan<- bool, []string) {
	switch u {
	case User:
		return roLock(&self.userMut), self.allowUser
	case Chan:
		return roLock(&self.chanMut), self.allowChan
	default:
		return nil, nil
	}
}

// Returns a pointer of the slice header to avoid copies. This is a promise by
// the user to not modify the slice and enforced by a RLock() on the slice. The
// lock is released and slice pointer set to nil when a boolean value is sent on
// the returned channel
func (self *moduleConfig) GetRODenyed(u UserChan) (chan<- bool, []string) {
	switch u {
	case User:
		return roLock(&self.userMut), self.denyUser
	case Chan:
		return roLock(&self.chanMut), self.denyChan
	default:
		return nil, nil
	}
}

// Clears allow(User|Chan) slices
func (self *moduleConfig) ClearAllowed(u UserChan) {
	switch u {
	case User:
		self.userMut.Lock()
		defer self.userMut.Unlock()

		clear(&self.allowUser)
	case Chan:
		self.chanMut.Lock()
		defer self.chanMut.Unlock()

		clear(&self.allowChan)
	}
}

// Clears deny(User|Chan) slices
func (self *moduleConfig) ClearDenyed(u UserChan) {
	switch u {
	case User:
		self.userMut.Lock()
		defer self.userMut.Unlock()

		clear(&self.denyUser)
	case Chan:
		self.chanMut.Lock()
		defer self.chanMut.Unlock()

		clear(&self.denyChan)
	}
}

// Returns true if 'target' is in allow(User|Chan); chans should begin with '#'
func (self *moduleConfig) InAllowed(target string) bool {
	if len(target) == 0 {
		return false
	}

	if target[0] == '#' {
		self.chanMut.RLock()
		defer self.chanMut.RUnlock()

		return inSlice(self.allowChan, target)
	} else {
		self.userMut.RLock()
		defer self.userMut.RUnlock()

		return inSlice(self.allowUser, target)
	}
}

// Returns true if 'target' is in deny(User|Chan); chans should begin with '#'
func (self *moduleConfig) InDenyed(target string) bool {
	if len(target) == 0 {
		return false
	}

	if target[0] == '#' {
		self.chanMut.RLock()
		defer self.chanMut.RUnlock()

		return inSlice(self.denyChan, target)
	} else {
		self.userMut.RLock()
		defer self.userMut.RUnlock()

		return inSlice(self.denyUser, target)
	}
}

// Returns the length of allow(User|Chan)
func (self *moduleConfig) LenAllowed(u UserChan) int {
	switch u {
	case User:
		self.userMut.RLock()
		defer self.userMut.RUnlock()

		return len(self.allowUser)
	default: // case Chan:
		self.chanMut.Lock()
		defer self.chanMut.Unlock()

		return len(self.allowChan)
	}
}

// Returns the length of allow(User|Chan)
func (self *moduleConfig) LenDenyed(u UserChan) int {
	switch u {
	case User:
		self.userMut.RLock()
		defer self.userMut.RUnlock()

		return len(self.denyUser)
	default: // case Chan:
		self.chanMut.Lock()
		defer self.chanMut.Unlock()

		return len(self.denyChan)
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
