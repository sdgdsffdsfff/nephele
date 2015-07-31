package util

import (
	"regexp"
	"strconv"
)

type RegexpExt struct {
	*regexp.Regexp
}

func (r *RegexpExt) FindStringSubmatchMap(s string) (map[string]string, bool) {
	captures := make(map[string]string)
	match := r.FindStringSubmatch(s)
	if match == nil {
		return captures, false
	}
	for i, name := range r.SubexpNames() {
		if i == 0 {
			captures[":0"] = s
		}
		if name == "" {
			if match[i] != "" {
				captures[":"+strconv.Itoa(i)] = match[i]
			}
		} else {
			captures[name] = match[i]
		}
	}
	return captures, true
}
