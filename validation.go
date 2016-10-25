package main

import (
	"regexp"
	"strconv"
	"strings"
)

func validName(name string) bool {
	r, _ := regexp.Compile("^[a-zA-Z0-9_-]+$")
	return r.Match([]byte(name))
}

// validAckID checks the validity of an AckID string against a given project and subscription
func validAckID(project string, sub string, ackID string) bool {

	tokens := strings.Split(ackID, "/")

	if len(tokens) != 4 || tokens[0] != "projects" || tokens[1] != project || tokens[2] != "subscriptions" {
		return false
	}

	subTokens := strings.Split(tokens[3], ":")
	if len(subTokens) != 2 || subTokens[0] != sub {
		return false
	}
	_, err := strconv.ParseInt(subTokens[1], 10, 64)
	if err != nil {

		return false
	}

	return true
}
