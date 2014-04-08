package irclib

import (
	"fmt"
	"regexp"
	"strings"
)

// Removes 'target' from 'list'; returns an error if 'target' is not in 'list'
// This will not preserve order
func remove(list *[]string, target string) bool {
	target = strings.ToLower(target)

	for i, v := range *list {
		if target != v {
			continue
		}

		listLen := len(*list) - 1
		(*list)[i] = (*list)[listLen]
		(*list)[listLen] = ""
		(*list) = (*list)[:listLen]

		return true
	}

	return false
}

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
