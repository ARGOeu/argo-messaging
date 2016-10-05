package auth

import (
	"crypto/rand"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"errors"

	"github.com/ARGOeu/argo-messaging/stores"
	"github.com/ARGOeu/argo-messaging/subscriptions"
	"github.com/ARGOeu/argo-messaging/topics"
)

// User is the struct that holds user information
type User struct {
	UUID         string         `json:"-"`
	Projects     []ProjectRoles `json:"projects,omitempty"`
	Name         string         `json:"name"`
	Token        string         `json:"token,omitempty"`
	Email        string         `json:"email,omitempty"`
	ServiceRoles []string       `json:"service_roles,omitempty"`
}

// ProjectRoles is the struct that hold project and role information of the user
type ProjectRoles struct {
	ProjectUUID string   `json:"project_uuid"`
	Roles       []string `json:"roles"`
}

// Users holds a list of available users
type Users struct {
	List []User `json:"users,omitempty"`
}

// ExportJSON exports Project to json format
func (u *User) ExportJSON() (string, error) {
	output, err := json.MarshalIndent(u, "", "   ")
	return string(output[:]), err
}

// ExportJSON exports Projects list to json format
func (us *Users) ExportJSON() (string, error) {
	output, err := json.MarshalIndent(us, "", "   ")
	return string(output[:]), err
}

// Empty returns true if users list is empty
func (us *Users) Empty() bool {
	if us.List == nil {
		return true
	}
	return len(us.List) <= 0
}

// One returns the first user if a user list is not empty
func (us *Users) One() User {
	if us.Empty() == false {
		return us.List[0]
	}
	return User{}
}

// GetUserFromJSON retrieves User info From JSON string
func GetUserFromJSON(input []byte) (User, error) {
	u := User{}
	err := json.Unmarshal([]byte(input), &u)
	return u, err
}

// NewUser accepts parameters and creates a new user
func NewUser(uuid string, projects []ProjectRoles, name string, token string, email string, serviceRoles []string) User {
	return User{UUID: uuid, Projects: projects, Name: name, Token: token, Email: email, ServiceRoles: serviceRoles}
}

// FindUsers returns a specific user or a list of all available users belonging to a  project in the datastore.
func FindUsers(projectUUID string, uuid string, name string, store stores.Store) (Users, error) {
	result := Users{}

	users, err := store.QueryUsers(projectUUID, uuid, name)
	for _, item := range users {
		pRoles := []ProjectRoles{}
		for _, pItem := range item.Projects {
			pRoles = append(pRoles, ProjectRoles{ProjectUUID: pItem.ProjectUUID, Roles: pItem.Roles})
		}
		curUser := NewUser(item.UUID, pRoles, item.Name, item.Token, item.Email, item.ServiceRoles)
		result.List = append(result.List, curUser)
	}

	return result, err
}

// Authenticate based on token
func Authenticate(projectUUID string, token string, store stores.Store) ([]string, string) {
	roles, user := store.GetUserRoles(projectUUID, token)

	return roles, user
}

// ExistsWithName returns true if a user with name exists
func ExistsWithName(name string, store stores.Store) bool {
	result := false

	users, err := store.QueryUsers("", "", name)
	if len(users) > 0 && err == nil {
		result = true
	}

	return result
}

// ExistsWithUUID return true if a user with uuid exists
func ExistsWithUUID(uuid string, store stores.Store) bool {
	result := false

	users, err := store.QueryUsers("", uuid, "")
	if len(users) > 0 && err == nil {
		result = true
	}

	return result
}

// GetNameByUUID queries user by UUID and returns the user's name. If not found, returns an empty string
func GetNameByUUID(uuid string, store stores.Store) string {
	result := ""
	users, err := store.QueryUsers("", uuid, "")
	if len(users) > 0 && err == nil {
		result = users[0].Name
	}

	return result
}

// GetUUIDByName queries user by name and returns the corresponding UUID
func GetUUIDByName(name string, store stores.Store) string {
	result := ""
	users, err := store.QueryUsers("", "", name)
	if len(users) > 0 && err == nil {
		result = users[0].UUID
	}

	return result
}

// CreateUser creates a new user
func CreateUser(uuid string, name string, projects []ProjectRoles, token string, email string, serviceRoles []string, store stores.Store) (User, error) {
	// check if project with the same name exists
	if ExistsWithName(name, store) {
		return User{}, errors.New("exists")
	}

	// Prep project roles for datastore insert
	prList := []stores.QProjectRoles{}
	for _, item := range projects {
		prList = append(prList, stores.QProjectRoles{ProjectUUID: item.ProjectUUID, Roles: item.Roles})
	}

	if err := store.InsertUser(uuid, prList, name, token, email, serviceRoles); err != nil {
		return User{}, errors.New("backend error")
	}

	// reflect stored object
	stored, err := FindUsers("", "", name, store)
	return stored.One(), err
}

// GenToken generates a new token
func GenToken() (string, error) {
	tokenLen := 32
	tokenBytes := make([]byte, tokenLen)
	if _, err := rand.Read(tokenBytes); err != nil {
		return "", err
	}
	sha1Bytes := sha1.Sum(tokenBytes)
	return hex.EncodeToString(sha1Bytes[:]), nil
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

// AreValidUsers accepts a user array of usernames and checks if users exist in the store
func AreValidUsers(project string, users []string, store stores.Store) (bool, error) {
	found, notFound := store.HasUsers(project, users)
	if found {
		return true, nil
	}

	var list string

	for i, username := range notFound {
		if i == 0 {
			list = list + username
		} else {
			list = list + ", " + username
		}

	}
	return false, errors.New("User(s): " + list + " do not exist")

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
