package main

import "regexp"

func validName(name string) bool {
	r, _ := regexp.Compile("^\\w+?$")
	return r.Match([]byte(name))
}
