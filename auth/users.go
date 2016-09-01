package auth

import (
	"github.com/ARGOeu/argo-messaging/stores"
	"github.com/ARGOeu/argo-messaging/subscriptions"
	"github.com/ARGOeu/argo-messaging/topics"
)

// User is the struct that holds user information
type User struct {
	Name    string
	Email   string
	Project string
	Token   string
	Roles   []string
}

// Users holds a list of available users
type Users struct {
	List []Users
}

// Authenticate based on token
func Authenticate(project string, token string, store stores.Store) ([]string, string) {
	roles, user := store.GetUserRoles(project, token)
	return roles, user
}

// IsPublisher Checks if a user is publisher
func IsPublisher(roles []string) bool {
	for _, role := range roles {
		if role == "publisher" {
			return true
		}
	}

	return false
}

// IsConsumer Checks if a user is consumer
func IsConsumer(roles []string) bool {
	for _, role := range roles {
		if role == "consumer" {
			return true
		}
	}

	return false
}

// PerResource  (for topics and subscriptions)
func PerResource(project string, resType string, resName string, user string, store stores.Store) bool {
	if resType == "topic" {
		tACL, _ := topics.GetTopicACL(project, resName, store)
		for _, item := range tACL.AuthUsers {
			if item == user {
				return true
			}
		}
	} else if resType == "subscription" {
		sACL, _ := subscriptions.GetSubACL(project, resName, store)
		for _, item := range sACL.AuthUsers {
			if item == user {
				return true
			}
		}
	}

	return false
}

// Authorize based on resource and  role information
func Authorize(resource string, roles []string, store stores.Store) bool {
	// check if _admin_ is in roles
	for _, role := range roles {
		if role == "_admin_" {
			return true
		}
	}

	return store.HasResourceRoles(resource, roles)
}
