package module

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
)

// Mostly private utility functions

// Adds 'target' to 'list' if 'target' is not in 'list', otherwise returns an error
func add(list *[]string, target string) error {
	target = strings.ToLower(target)

	if inSlice(*list, target) {
		return errors.New(target + " is already in slice")
	}

	*list = append(*list, target)

	return nil
}

// Removes 'target' from 'list'; returns an error if 'target' is not in 'list'
// This will not preserve order
func remove(list *[]string, target string) error {
	target = strings.ToLower(target)

	for i, v := range *list {
		if target != v {
			continue
		}

		listLen := len(*list) - 1
		(*list)[i] = (*list)[listLen]
		(*list)[listLen] = ""
		(*list) = (*list)[:listLen]

		return nil
	}

	return errors.New(target + " is not in slice")
}

// Sets 'list' to point to a new slice of strings
func clear(list *[]string) {
	*list = make([]string, 0, 5)
}

// Return a copy of list
func copySlice(list []string) []string {
	s := make([]string, len(list))
	copy(s, list)

	return s
}

// Return true if 'target' is in 'list'
func inSlice(list []string, target string) bool {
	target = strings.ToLower(target)

	for _, v := range list {
		if v == target {
			return true
		}
	}

	return false
}

func toLowerSlice(list []string) {
	for i, str := range list {
		list[i] = strings.ToLower(str)
	}
}

// Helper function to match named groups
func matchGroups(reg *regexp.Regexp, s string) (map[string]string, error) {
	groups := make(map[string]string)
	res := reg.FindStringSubmatch(s)
	if res == nil {
		return nil, fmt.Errorf("%v did not match regexp", s)
	}

	groupNames := reg.SubexpNames()
	for k, v := range groupNames {
		if v != "" {
			groups[v] = res[k]
		}
	}

	return groups, nil
}
