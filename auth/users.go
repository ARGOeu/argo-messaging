package auth

import "github.com/ARGOeu/argo-messaging/stores"

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
func Authenticate(project string, token string, store stores.Store) []string {
	return store.GetUserRoles(project, token)
}

// Authorize based on resource and  role information
func Authorize(resource string, roles []string, store stores.Store) bool {
	return store.HasResourceRoles(resource, roles)
}
