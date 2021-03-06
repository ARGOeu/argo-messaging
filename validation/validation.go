package validation

import (
	"net/url"
	"regexp"
	"strconv"
	"strings"
)

func ValidName(name string) bool {
	r, _ := regexp.Compile("^[a-zA-Z0-9_-]+$")
	return r.Match([]byte(name))
}

// ValidAckID checks the validity of an AckID string against a given project and subscription
func ValidAckID(project string, sub string, ackID string) bool {

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

// IsValidHTTPS checks if a url string is valid https url
func IsValidHTTPS(urlStr string) bool {
	u, err := url.ParseRequestURI(urlStr)
	if err != nil {
		return false
	}
	// If a valid url is in form without slashes after scheme consider it invalid.
	// If a valid url doesn't have https as a scheme consider it invalid
	if u.Host == "" || u.Scheme != "https" {
		return false
	}

	return true
}
