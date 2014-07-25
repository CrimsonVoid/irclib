package irclib

import (
	"sync"
)

type access struct {
	list map[string][]string
	mut  sync.RWMutex
}

// Add `nick` to `group` if `nick` is not in `group`
func (self *access) Add(nick, group string) bool {
	self.mut.Lock()
	defer self.mut.RUnlock()

	if self.inGroup(nick, group) {
		return false
	}

	self.list[group] = append(self.list[group], nick)
	return true
}

// Removes `nick` from `group`
func (self *access) Remove(nick, group string) bool {
	self.mut.Lock()
	defer self.mut.RUnlock()

	list, ok := self.list[group]
	if !ok {
		return false
	}

	return remove(&list, nick)
}

// Returns a map[string][]string of groups from access list. If a requested group
// is not in the access list that value is not added to the returned map
func (self *access) Groups(groups ...string) map[string][]string {
	self.mut.RLock()
	defer self.mut.RUnlock()

	if len(groups) == 0 {
		groups = make([]string, 0, len(self.list))

		for grp := range self.list {
			groups = append(groups, grp)
		}
	}

	accessGroup := make(map[string][]string)
	for _, grp := range groups {
		list, ok := self.list[grp]
		if !ok {
			continue
		}

		l := make([]string, len(list))
		copy(l, list)
		accessGroup[grp] = l
	}

	return accessGroup
}

// Returns the first group `nick` occurs or an empty string if `nick` is not is
// the list of groups provided
func (self *access) InGroups(nick string, groups ...string) string {
	self.mut.RLock()
	defer self.mut.RUnlock()

	for _, grp := range groups {
		if self.inGroup(nick, grp) {
			return grp
		}
	}

	return ""
}

// Helper function for access.InGroups()
func (self *access) inGroup(nick, group string) bool {
	// Locked by callee
	for _, u := range (*self).list[group] {
		if u == nick {
			return true
		}
	}

	return false
}
