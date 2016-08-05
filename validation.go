package main

import "regexp"

func validName(name string) bool {
	r, _ := regexp.Compile("^[a-zA-Z0-9_-]+$")
	return r.Match([]byte(name))
}
