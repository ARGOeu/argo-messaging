package auth

import (
	"crypto/rand"
	"crypto/sha1"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/ARGOeu/argo-messaging/projects"
	"github.com/ARGOeu/argo-messaging/stores"
	log "github.com/sirupsen/logrus"
	"time"
)

// User is the struct that holds user information
type User struct {
	UUID         string         `json:"uuid"`
	Projects     []ProjectRoles `json:"projects"`
	Name         string         `json:"name"`
	Token        string         `json:"token"`
	Email        string         `json:"email"`
	ServiceRoles []string       `json:"service_roles"`
	CreatedOn    string         `json:"created_on,omitempty"`
	ModifiedOn   string         `json:"modified_on,omitempty"`
	CreatedBy    string         `json:"created_by,omitempty"`
}

// ProjectRoles is the struct that hold project and role information of the user
type ProjectRoles struct {
	Project string   `json:"project"`
	Roles   []string `json:"roles"`
	Topics  []string `json:"topics"`
	Subs    []string `json:"subscriptions"`
}

// Users holds a list of available users
type Users struct {
	List []User `json:"users,omitempty"`
}

// PaginatedUsers holds information about a users' page and how to access the next page
type PaginatedUsers struct {
	Users         []User `json:"users"`
	NextPageToken string `json:"nextPageToken"`
	TotalSize     int32  `json:"totalSize"`
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

// ExportJSON exports Projects list to json format
func (pus *PaginatedUsers) ExportJSON() (string, error) {
	output, err := json.MarshalIndent(pus, "", "   ")
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
func NewUser(uuid string, projects []ProjectRoles, name string, token string, email string, serviceRoles []string, createdOn time.Time, modifiedOn time.Time, createdBy string) User {
	zuluForm := "2006-01-02T15:04:05Z"
	return User{UUID: uuid, Projects: projects, Name: name, Token: token, Email: email, ServiceRoles: serviceRoles, CreatedOn: createdOn.Format(zuluForm), ModifiedOn: modifiedOn.Format(zuluForm), CreatedBy: createdBy}
}

func GetPushWorker(pwToken string, store stores.Store) (User, error) {

	pw, err := GetUserByToken(pwToken, store)
	if err != nil {
		log.Errorf("Could not retrieve push worker user with token %v, %v", pwToken, err.Error())
		return User{}, errors.New("push_500")
	}

	return pw, nil
}

// GetUserByToken returns a specific user by his token
func GetUserByToken(token string, store stores.Store) (User, error) {
	result := User{}

	user, err := store.GetUserFromToken(token)

	if err != nil {
		return result, err
	}

	usernameC := ""
	if user.CreatedBy != "" {
		usr, err := store.QueryUsers("", user.CreatedBy, "")
		if err == nil && len(usr) > 0 {
			usernameC = usr[0].Name

		}
	}

	pRoles := []ProjectRoles{}
	for _, pItem := range user.Projects {
		prName := projects.GetNameByUUID(pItem.ProjectUUID, store)
		// Get User topics and subscriptions

		topicList, _ := store.QueryTopicsByACL(pItem.ProjectUUID, user.UUID)
		topicNames := []string{}
		for _, tpItem := range topicList {
			topicNames = append(topicNames, tpItem.Name)
		}

		subList, _ := store.QuerySubsByACL(pItem.ProjectUUID, user.UUID)
		subNames := []string{}
		for _, sbItem := range subList {
			subNames = append(subNames, sbItem.Name)
		}
		pRoles = append(pRoles, ProjectRoles{Project: prName, Roles: pItem.Roles, Topics: topicNames, Subs: subNames})
	}

	curUser := NewUser(user.UUID, pRoles, user.Name, user.Token, user.Email, user.ServiceRoles, user.CreatedOn.UTC(), user.ModifiedOn.UTC(), usernameC)

	result = curUser

	return result, err
}

// FindUsers returns a specific user or a list of all available users belonging to a  project in the datastore.
func FindUsers(projectUUID string, uuid string, name string, store stores.Store) (Users, error) {
	result := Users{}

	users, err := store.QueryUsers(projectUUID, uuid, name)

	for _, item := range users {

		// Get Username from user uuid

		usernameC := ""
		if item.CreatedBy != "" {
			usr, err := store.QueryUsers("", item.CreatedBy, "")
			if err == nil && len(usr) > 0 {
				usernameC = usr[0].Name

			}
		}

		pRoles := []ProjectRoles{}
		for _, pItem := range item.Projects {
			prName := projects.GetNameByUUID(pItem.ProjectUUID, store)
			// Get User topics and subscriptions

			topicList, _ := store.QueryTopicsByACL(pItem.ProjectUUID, item.UUID)
			topicNames := []string{}
			for _, tpItem := range topicList {
				topicNames = append(topicNames, tpItem.Name)
			}

			subList, _ := store.QuerySubsByACL(pItem.ProjectUUID, item.UUID)
			subNames := []string{}
			for _, sbItem := range subList {
				subNames = append(subNames, sbItem.Name)
			}
			pRoles = append(pRoles, ProjectRoles{Project: prName, Roles: pItem.Roles, Topics: topicNames, Subs: subNames})
		}

		curUser := NewUser(item.UUID, pRoles, item.Name, item.Token, item.Email, item.ServiceRoles, item.CreatedOn.UTC(), item.ModifiedOn.UTC(), usernameC)

		result.List = append(result.List, curUser)
	}

	if len(result.List) == 0 {
		err = errors.New("not found")
	}

	return result, err
}

// PaginatedFindUsers returns a page of users
func PaginatedFindUsers(pageToken string, pageSize int32, store stores.Store) (PaginatedUsers, error) {

	var totalSize int32
	var nextPageToken string
	var err error
	var users []stores.QUser
	var pageTokenBytes []byte

	// decode the base64 pageToken
	if pageTokenBytes, err = base64.StdEncoding.DecodeString(pageToken); err != nil {
		log.Errorf("Page token %v produced an error while being decoded to base64: %v", pageToken, err.Error())
		return PaginatedUsers{}, err
	}

	result := PaginatedUsers{Users: []User{}}

	if users, totalSize, nextPageToken, err = store.PaginatedQueryUsers(string(pageTokenBytes), pageSize); err != nil {
		return result, err
	}

	for _, item := range users {

		// Get Username from user uuid

		usernameC := ""
		if item.CreatedBy != "" {
			usr, err := store.QueryUsers("", item.CreatedBy, "")
			if err == nil && len(usr) > 0 {
				usernameC = usr[0].Name

			}
		}

		pRoles := []ProjectRoles{}
		for _, pItem := range item.Projects {
			prName := projects.GetNameByUUID(pItem.ProjectUUID, store)
			// Get User topics and subscriptions

			topicList, _ := store.QueryTopicsByACL(pItem.ProjectUUID, item.UUID)
			topicNames := []string{}
			for _, tpItem := range topicList {
				topicNames = append(topicNames, tpItem.Name)
			}

			subList, _ := store.QuerySubsByACL(pItem.ProjectUUID, item.UUID)
			subNames := []string{}
			for _, sbItem := range subList {
				subNames = append(subNames, sbItem.Name)
			}
			pRoles = append(pRoles, ProjectRoles{Project: prName, Roles: pItem.Roles, Topics: topicNames, Subs: subNames})
		}

		curUser := NewUser(item.UUID, pRoles, item.Name, item.Token, item.Email, item.ServiceRoles, item.CreatedOn.UTC(), item.ModifiedOn.UTC(), usernameC)

		result.Users = append(result.Users, curUser)
	}

	//encode to base64 the next page token
	result.NextPageToken = base64.StdEncoding.EncodeToString([]byte(nextPageToken))
	result.TotalSize = totalSize

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

func GetUserByUUID(uuid string, store stores.Store) (User, error) {

	var result User

	users, err := store.QueryUsers("", uuid, "")

	if err != nil {
		return User{}, err
	}

	if len(users) == 0 {
		return User{}, errors.New("not found")
	}

	if len(users) > 1 {
		return User{}, errors.New("multiple uuids")

	}

	user := users[0]

	//convert the Quser to User
	usernameC := ""
	if user.CreatedBy != "" {
		usr, err := store.QueryUsers("", user.CreatedBy, "")
		if err == nil && len(usr) > 0 {
			usernameC = usr[0].Name

		}
	}

	pRoles := []ProjectRoles{}
	for _, pItem := range user.Projects {
		prName := projects.GetNameByUUID(pItem.ProjectUUID, store)
		// Get User topics and subscriptions

		topicList, _ := store.QueryTopicsByACL(pItem.ProjectUUID, user.UUID)
		topicNames := []string{}
		for _, tpItem := range topicList {
			topicNames = append(topicNames, tpItem.Name)
		}

		subList, _ := store.QuerySubsByACL(pItem.ProjectUUID, user.UUID)
		subNames := []string{}
		for _, sbItem := range subList {
			subNames = append(subNames, sbItem.Name)
		}
		pRoles = append(pRoles, ProjectRoles{Project: prName, Roles: pItem.Roles, Topics: topicNames, Subs: subNames})
	}

	curUser := NewUser(user.UUID, pRoles, user.Name, user.Token, user.Email, user.ServiceRoles, user.CreatedOn, user.ModifiedOn, usernameC)

	result = curUser

	return result, nil

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

// UpdateUserToken updates an existing user's token
func UpdateUserToken(uuid string, token string, store stores.Store) (User, error) {
	if err := store.UpdateUserToken(uuid, token); err != nil {
		return User{}, err
	}
	// reflect stored object
	stored, err := FindUsers("", uuid, "", store)
	return stored.One(), err
}

// AppendToUserProjects appends a unique project to the user's project list
func AppendToUserProjects(userUUID string, projectUUID string, store stores.Store, pRoles ...string) error {

	pName := projects.GetNameByUUID(projectUUID, store)
	if pName == "" {
		return fmt.Errorf("invalid project %v", projectUUID)
	}

	validRoles := store.GetAllRoles()

	for _, role := range pRoles {
		if !IsRoleValid(role, validRoles) {
			return fmt.Errorf("invalid role %v", role)
		}
	}

	err := store.AppendToUserProjects(userUUID, projectUUID, pRoles...)
	if err != nil {
		return err
	}

	return nil
}

// UpdateUser updates an existing user's information
func UpdateUser(uuid string, name string, projectList []ProjectRoles, email string, serviceRoles []string, modifiedOn time.Time, store stores.Store) (User, error) {

	prList := []stores.QProjectRoles{}

	validRoles := store.GetAllRoles()

	if projectList != nil {
		for _, item := range projectList {
			prUUID := projects.GetUUIDByName(item.Project, store)
			// If project name doesn't reflect a uuid, then is non existent
			if prUUID == "" {
				return User{}, errors.New("invalid project: " + item.Project)
			}

			// Check roles

			for _, roleItem := range item.Roles {
				if IsRoleValid(roleItem, validRoles) == false {
					return User{}, errors.New("invalid role: " + roleItem)
				}
			}
			prList = append(prList, stores.QProjectRoles{ProjectUUID: prUUID, Roles: item.Roles})
		}

	} else {
		prList = nil
	}

	if serviceRoles != nil {
		for _, roleItem := range serviceRoles {
			if IsRoleValid(roleItem, validRoles) == false {
				return User{}, errors.New("invalid role: " + roleItem)
			}
		}
	}

	if err := store.UpdateUser(uuid, prList, name, email, serviceRoles, modifiedOn); err != nil {
		return User{}, err
	}

	// reflect stored object
	stored, err := FindUsers("", uuid, "", store)
	return stored.One(), err
}

// CreateUser creates a new user
func CreateUser(uuid string, name string, projectList []ProjectRoles, token string, email string, serviceRoles []string, createdOn time.Time, createdBy string, store stores.Store) (User, error) {
	// check if project with the same name exists
	if ExistsWithName(name, store) {
		return User{}, errors.New("exists")
	}

	// Prep project roles for datastore insert
	prList := []stores.QProjectRoles{}
	for _, item := range projectList {
		prUUID := projects.GetUUIDByName(item.Project, store)
		// If project name doesn't reflect a uuid, then is non existent
		if prUUID == "" {
			return User{}, errors.New("invalid project: " + item.Project)
		}

		// Check roles
		validRoles := store.GetAllRoles()
		for _, roleItem := range item.Roles {
			if IsRoleValid(roleItem, validRoles) == false {
				return User{}, errors.New("invalid role: " + roleItem)
			}
		}
		prList = append(prList, stores.QProjectRoles{ProjectUUID: prUUID, Roles: item.Roles})
	}

	if err := store.InsertUser(uuid, prList, name, token, email, serviceRoles, createdOn, createdOn, createdBy); err != nil {
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

// IsPushWorker Checks if a user is a push worker
func IsPushWorker(roles []string) bool {
	for _, role := range roles {
		if role == "push_worker" {
			return true
		}
	}

	return false
}

// IsProjectAdmin checks if the user is a project admin
func IsProjectAdmin(roles []string) bool {
	for _, role := range roles {
		if role == "project_admin" {
			return true
		}
	}

	return false
}

// IsServiceAdmin checks if the user is a service admin
func IsServiceAdmin(roles []string) bool {
	for _, role := range roles {
		if role == "service_admin" {
			return true
		}
	}

	return false
}

// RemoveUser removes an existing user
func RemoveUser(uuid string, store stores.Store) error {
	return store.RemoveUser(uuid)
}

// IsRoleValid checks if a role is a valid against a list of valid roles
func IsRoleValid(role string, validRoles []string) bool {
	for _, roleItem := range validRoles {
		if roleItem == role {
			return true
		}
	}
	return false
}

// AreValidUsers accepts a user array of usernames and checks if users exist in the store
func AreValidUsers(projectUUID string, users []string, store stores.Store) (bool, error) {
	found, notFound := store.HasUsers(projectUUID, users)
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
func PerResource(project string, resType string, resName string, userUUID string, store stores.Store) bool {

	if resType == "topics" || resType == "subscriptions" {
		err := store.ExistsInACL(project, resType, resName, userUUID)
		if err != nil {
			return false
			log.Errorln(err.Error())
		}

		return true

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
