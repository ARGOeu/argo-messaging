package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/ARGOeu/argo-messaging/projects"
	"github.com/ARGOeu/argo-messaging/stores"
	log "github.com/sirupsen/logrus"
)

const (
	AcceptedRegistrationStatus = "accepted"
	PendingRegistrationStatus  = "pending"
	DeclinedRegistrationStatus = "declined"
)

// User is the struct that holds user information
type User struct {
	UUID         string         `json:"uuid"`
	Projects     []ProjectRoles `json:"projects,omitempty"`
	Name         string         `json:"name"`
	FirstName    string         `json:"first_name,omitempty"`
	LastName     string         `json:"last_name,omitempty"`
	Organization string         `json:"organization,omitempty"`
	Description  string         `json:"description,omitempty"`
	Token        string         `json:"token,omitempty"`
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
	TotalSize     int64  `json:"totalSize"`
}

// UserRegistration holds information about a new user registration
type UserRegistration struct {
	UUID            string `json:"uuid"`
	Name            string `json:"name"`
	FirstName       string `json:"first_name"`
	LastName        string `json:"last_name"`
	Organization    string `json:"organization"`
	Description     string `json:"description"`
	Email           string `json:"email"`
	Status          string `json:"status"`
	DeclineComment  string `json:"decline_comment,omitempty"`
	ActivationToken string `json:"activation_token,omitempty"`
	RegisteredAt    string `json:"registered_at"`
	ModifiedBy      string `json:"modified_by,omitempty"`
	ModifiedAt      string `json:"modified_at,omitempty"`
}

// UserRegistrationsList holds a list with all the user registrations in the service
type UserRegistrationsList struct {
	UserRegistrations []UserRegistration `json:"user_registrations"`
}

// ExportJSON exports User to json format
func (u *User) ExportJSON() (string, error) {
	output, err := json.MarshalIndent(u, "", "   ")
	return string(output[:]), err
}

// ExportJSON exports Users list to json format
func (us *Users) ExportJSON() (string, error) {
	output, err := json.MarshalIndent(us, "", "   ")
	return string(output[:]), err
}

// ExportJSON exports Paginated users list to json format
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

// RegisterUser registers a new user to the store
func RegisterUser(ctx context.Context, uuid, name, fname, lname, email, org, desc, registeredAt, atkn, status string, str stores.Store) (UserRegistration, error) {

	err := str.RegisterUser(ctx, uuid, name, fname, lname, email, org, desc, registeredAt, atkn, status)
	if err != nil {
		return UserRegistration{}, err
	}

	return UserRegistration{
		UUID:            uuid,
		Name:            name,
		FirstName:       fname,
		LastName:        lname,
		Email:           email,
		Organization:    org,
		Description:     desc,
		RegisteredAt:    registeredAt,
		ActivationToken: atkn,
		Status:          status,
	}, nil
}

// DeleteUserRegistration removes the respective registration from the store based on the given uuid
func DeleteUserRegistration(ctx context.Context, regUUID string, str stores.Store) error {
	return str.DeleteRegistration(ctx, regUUID)
}

func FindUserRegistration(ctx context.Context, regUUID, status string, str stores.Store) (UserRegistration, error) {

	q, err := str.QueryRegistrations(ctx, regUUID, status, "", "", "", "")
	if err != nil {
		return UserRegistration{}, err
	}

	if len(q) == 0 {
		return UserRegistration{}, errors.New("not found")
	}

	usernameC := ""
	if q[0].ModifiedBy != "" {
		usr, err := str.QueryUsers(ctx, "", q[0].ModifiedBy, "")
		if err == nil && len(usr) > 0 {
			usernameC = usr[0].Name

		}
	}

	ur := UserRegistration{
		UUID:            q[0].UUID,
		Name:            q[0].Name,
		FirstName:       q[0].FirstName,
		LastName:        q[0].LastName,
		Email:           q[0].Email,
		ActivationToken: q[0].ActivationToken,
		Status:          q[0].Status,
		DeclineComment:  q[0].DeclineComment,
		Organization:    q[0].Organization,
		Description:     q[0].Description,
		RegisteredAt:    q[0].RegisteredAt,
		ModifiedBy:      usernameC,
		ModifiedAt:      q[0].ModifiedAt,
	}

	return ur, nil
}

func FindUserRegistrations(ctx context.Context, status, activationToken, name, email, org string, str stores.Store) (UserRegistrationsList, error) {

	q, err := str.QueryRegistrations(ctx, "", status, activationToken, name, email, org)
	if err != nil {
		return UserRegistrationsList{}, err
	}

	urList := UserRegistrationsList{
		UserRegistrations: []UserRegistration{},
	}

	for _, ur := range q {

		usernameC := ""
		if ur.ModifiedBy != "" {
			usr, err := str.QueryUsers(ctx, "", ur.ModifiedBy, "")
			if err == nil && len(usr) > 0 {
				usernameC = usr[0].Name

			}
		}

		tempUR := UserRegistration{
			UUID:            ur.UUID,
			Name:            ur.Name,
			FirstName:       ur.FirstName,
			LastName:        ur.LastName,
			Email:           ur.Email,
			ActivationToken: ur.ActivationToken,
			Status:          ur.Status,
			DeclineComment:  ur.DeclineComment,
			Organization:    ur.Organization,
			Description:     ur.Description,
			RegisteredAt:    ur.RegisteredAt,
			ModifiedBy:      usernameC,
			ModifiedAt:      ur.ModifiedAt,
		}

		urList.UserRegistrations = append(urList.UserRegistrations, tempUR)
	}

	return urList, nil
}

func UpdateUserRegistration(ctx context.Context, regUUID, status, declineComment, modifiedBy string, modifiedAt time.Time, refStr stores.Store) error {
	// only accept decline comment with the decline status action
	if status != DeclinedRegistrationStatus {
		declineComment = ""
	}
	return refStr.UpdateRegistration(ctx, regUUID, status, declineComment, modifiedBy, modifiedAt.UTC().Format("2006-01-02T15:04:05Z"))
}

// NewUser accepts parameters and creates a new user
func NewUser(uuid string, projects []ProjectRoles, name string, fname string, lname string, org string, desc string, token string, email string, serviceRoles []string, createdOn time.Time, modifiedOn time.Time, createdBy string) User {
	zuluForm := "2006-01-02T15:04:05Z"
	return User{
		UUID:         uuid,
		Projects:     projects,
		Name:         name,
		FirstName:    fname,
		LastName:     lname,
		Organization: org,
		Description:  desc,
		Token:        token,
		Email:        email,
		ServiceRoles: serviceRoles,
		CreatedOn:    createdOn.Format(zuluForm),
		ModifiedOn:   modifiedOn.Format(zuluForm),
		CreatedBy:    createdBy}
}

// GetPushWorker returns a push worker user by token
func GetPushWorker(ctx context.Context, pwToken string, store stores.Store) (User, error) {

	pw, err := GetUserByToken(ctx, pwToken, store)
	if err != nil {
		log.WithFields(
			log.Fields{
				"trace_id": ctx.Value("trace_id"),
				"type":     "service_log",
				"token":    pwToken,
				"error":    err.Error(),
			},
		).Error("Could not retrieve push worker user")
		return User{}, errors.New("push_500")
	}

	return pw, nil
}

// GetUserByToken returns a specific user by his token
func GetUserByToken(ctx context.Context, token string, store stores.Store) (User, error) {
	result := User{}

	user, err := store.GetUserFromToken(ctx, token)

	if err != nil {
		return result, err
	}

	usernameC := ""
	if user.CreatedBy != "" {
		usr, err := store.QueryUsers(ctx, "", user.CreatedBy, "")
		if err == nil && len(usr) > 0 {
			usernameC = usr[0].Name

		}
	}

	pRoles := []ProjectRoles{}
	for _, pItem := range user.Projects {
		prName := projects.GetNameByUUID(ctx, pItem.ProjectUUID, store)
		// Get User topics and subscriptions

		topicList, _ := store.QueryTopicsByACL(ctx, pItem.ProjectUUID, user.UUID)
		topicNames := []string{}
		for _, tpItem := range topicList {
			topicNames = append(topicNames, tpItem.Name)
		}

		subList, _ := store.QuerySubsByACL(ctx, pItem.ProjectUUID, user.UUID)
		subNames := []string{}
		for _, sbItem := range subList {
			subNames = append(subNames, sbItem.Name)
		}
		pRoles = append(pRoles, ProjectRoles{Project: prName, Roles: pItem.Roles, Topics: topicNames, Subs: subNames})
	}

	curUser := NewUser(user.UUID, pRoles, user.Name, user.FirstName,
		user.LastName, user.Organization, user.Description, user.Token, user.Email,
		user.ServiceRoles, user.CreatedOn.UTC(), user.ModifiedOn.UTC(), usernameC)

	result = curUser

	return result, err
}

// FindUsers returns a specific user or a list of all available users belonging to a  project in the datastore.
func FindUsers(ctx context.Context, projectUUID string, uuid string, name string, priviledged bool, store stores.Store) (Users, error) {
	result := Users{}

	users, err := store.QueryUsers(ctx, projectUUID, uuid, name)

	for _, item := range users {

		// Get Username from user uuid

		// Get Username from user uuid
		serviceRoles := []string{}
		token := ""
		usernameC := ""

		// if call made by priviledged user (superuser), show service roles, token and user creator info
		if priviledged {
			if item.CreatedBy != "" {
				usr, err := store.QueryUsers(ctx, "", item.CreatedBy, "")
				if err == nil && len(usr) > 0 {
					usernameC = usr[0].Name

				}
			}
			token = item.Token
			serviceRoles = item.ServiceRoles
		}

		pRoles := []ProjectRoles{}
		for _, pItem := range item.Projects {
			// if user not priviledged (not superuser) and queried projectUUID doesn't
			// match current role item's project UUID, skip the item
			if !priviledged && pItem.ProjectUUID != projectUUID {
				continue
			}
			prName := projects.GetNameByUUID(ctx, pItem.ProjectUUID, store)
			// Get User topics and subscriptions

			topicList, _ := store.QueryTopicsByACL(ctx, pItem.ProjectUUID, item.UUID)
			topicNames := []string{}
			for _, tpItem := range topicList {
				topicNames = append(topicNames, tpItem.Name)
			}

			subList, _ := store.QuerySubsByACL(ctx, pItem.ProjectUUID, item.UUID)
			subNames := []string{}
			for _, sbItem := range subList {
				subNames = append(subNames, sbItem.Name)
			}

			// avoid null json reference with empty roles list
			_pRoles := []string{}
			if pItem.Roles != nil {
				_pRoles = pItem.Roles
			}

			pRoles = append(pRoles, ProjectRoles{Project: prName, Roles: _pRoles, Topics: topicNames, Subs: subNames})
		}

		curUser := NewUser(item.UUID, pRoles, item.Name, item.FirstName, item.LastName,
			item.Organization, item.Description, token, item.Email, serviceRoles,
			item.CreatedOn.UTC(), item.ModifiedOn.UTC(), usernameC)

		result.List = append(result.List, curUser)
	}

	if len(result.List) == 0 {
		err = errors.New("not found")
	}

	return result, err
}

// PaginatedFindUsers returns a page of users
func PaginatedFindUsers(ctx context.Context, pageToken string, pageSize int64, projectUUID string, privileged, detailedView bool, store stores.Store) (PaginatedUsers, error) {

	var totalSize int64
	var nextPageToken string
	var err error
	var users []stores.QUser
	var pageTokenBytes []byte

	// decode the base64 pageToken
	if pageTokenBytes, err = base64.StdEncoding.DecodeString(pageToken); err != nil {
		log.WithFields(
			log.Fields{
				"trace_id":   ctx.Value("trace_id"),
				"type":       "request_log",
				"page_token": pageToken,
				"error":      err.Error(),
			},
		).Error("error while decoding to base64")
		return PaginatedUsers{}, err
	}

	result := PaginatedUsers{Users: []User{}}

	if users, totalSize, nextPageToken, err = store.PaginatedQueryUsers(ctx, string(pageTokenBytes), pageSize, projectUUID); err != nil {
		return result, err
	}

	for _, item := range users {

		// Get Username from user uuid
		serviceRoles := []string{}
		token := ""
		usernameC := ""
		// if call made by priviledged user (superuser), show service roles, token and user creator info
		if privileged {
			if item.CreatedBy != "" {
				usr, err := store.QueryUsers(ctx, "", item.CreatedBy, "")
				if err == nil && len(usr) > 0 {
					usernameC = usr[0].Name

				}
			}
			token = item.Token
			serviceRoles = item.ServiceRoles
		}

		var pRoles []ProjectRoles

		if detailedView {

			for _, pItem := range item.Projects {
				// if user not priviledged (not superuser) and queried projectUUID doesn't
				// match current role item's project UUID, skip the item
				if !privileged && pItem.ProjectUUID != projectUUID {
					continue
				}
				prName := projects.GetNameByUUID(ctx, pItem.ProjectUUID, store)

				// Get User topics and subscriptions
				topicList, _ := store.QueryTopicsByACL(ctx, pItem.ProjectUUID, item.UUID)
				topicNames := []string{}
				for _, tpItem := range topicList {
					topicNames = append(topicNames, tpItem.Name)
				}

				subList, _ := store.QuerySubsByACL(ctx, pItem.ProjectUUID, item.UUID)
				subNames := []string{}
				for _, sbItem := range subList {
					subNames = append(subNames, sbItem.Name)
				}
				pRoles = append(pRoles, ProjectRoles{Project: prName, Roles: pItem.Roles, Topics: topicNames, Subs: subNames})
			}
		}

		curUser := NewUser(item.UUID, pRoles, item.Name, item.FirstName, item.LastName,
			item.Organization, item.Description, token, item.Email, serviceRoles,
			item.CreatedOn.UTC(), item.ModifiedOn.UTC(), usernameC)

		result.Users = append(result.Users, curUser)
	}

	//encode to base64 the next page token
	result.NextPageToken = base64.StdEncoding.EncodeToString([]byte(nextPageToken))
	result.TotalSize = totalSize

	return result, err
}

// Authenticate based on token
func Authenticate(ctx context.Context, projectUUID string, token string, store stores.Store) ([]string, string) {
	return store.GetUserRoles(ctx, projectUUID, token)
}

// ExistsWithName returns true if a user with name exists
func ExistsWithName(ctx context.Context, name string, store stores.Store) bool {
	result := false

	users, err := store.QueryUsers(ctx, "", "", name)
	if len(users) > 0 && err == nil {
		result = true
	}

	return result
}

// ExistsWithUUID return true if a user with uuid exists
func ExistsWithUUID(ctx context.Context, uuid string, store stores.Store) bool {
	result := false

	users, err := store.QueryUsers(ctx, "", uuid, "")
	if len(users) > 0 && err == nil {
		result = true
	}

	return result
}

// GetNameByUUID queries user by UUID and returns the user's name. If not found, returns an empty string
func GetNameByUUID(ctx context.Context, uuid string, store stores.Store) string {
	result := ""
	users, err := store.QueryUsers(ctx, "", uuid, "")
	if len(users) > 0 && err == nil {
		result = users[0].Name
	}

	return result
}

// GetUserByUUID returns user information by UUID
func GetUserByUUID(ctx context.Context, uuid string, store stores.Store) (User, error) {

	var result User

	users, err := store.QueryUsers(ctx, "", uuid, "")

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
		usr, err := store.QueryUsers(ctx, "", user.CreatedBy, "")
		if err == nil && len(usr) > 0 {
			usernameC = usr[0].Name

		}
	}

	pRoles := []ProjectRoles{}
	for _, pItem := range user.Projects {
		prName := projects.GetNameByUUID(ctx, pItem.ProjectUUID, store)
		// Get User topics and subscriptions

		topicList, _ := store.QueryTopicsByACL(ctx, pItem.ProjectUUID, user.UUID)
		topicNames := []string{}
		for _, tpItem := range topicList {
			topicNames = append(topicNames, tpItem.Name)
		}

		subList, _ := store.QuerySubsByACL(ctx, pItem.ProjectUUID, user.UUID)
		subNames := []string{}
		for _, sbItem := range subList {
			subNames = append(subNames, sbItem.Name)
		}
		pRoles = append(pRoles, ProjectRoles{Project: prName, Roles: pItem.Roles, Topics: topicNames, Subs: subNames})
	}

	curUser := NewUser(user.UUID, pRoles, user.Name, user.FirstName,
		user.LastName, user.Organization, user.Description, user.Token, user.Email,
		user.ServiceRoles, user.CreatedOn.UTC(), user.ModifiedOn.UTC(), usernameC)

	result = curUser

	return result, nil

}

// GetUUIDByName queries user by name and returns the corresponding UUID
func GetUUIDByName(ctx context.Context, name string, store stores.Store) string {
	result := ""
	users, err := store.QueryUsers(ctx, "", "", name)

	if len(users) > 0 && err == nil {
		result = users[0].UUID
	}

	return result
}

// UpdateUserToken updates an existing user's token
func UpdateUserToken(ctx context.Context, uuid string, token string, store stores.Store) (User, error) {
	if err := store.UpdateUserToken(ctx, uuid, token); err != nil {
		return User{}, err
	}
	// reflect stored object
	stored, err := FindUsers(ctx, "", uuid, "", true, store)
	return stored.One(), err
}

// AppendToUserProjects appends a unique project to the user's project list
func AppendToUserProjects(ctx context.Context, userUUID string, projectUUID string, store stores.Store, pRoles ...string) error {

	pName := projects.GetNameByUUID(ctx, projectUUID, store)
	if pName == "" {
		return fmt.Errorf("invalid project %v", projectUUID)
	}

	validRoles := store.GetAllRoles(ctx)

	for _, role := range pRoles {
		if !IsRoleValid(role, validRoles) {
			return fmt.Errorf("invalid role %v", role)
		}
	}

	err := store.AppendToUserProjects(ctx, userUUID, projectUUID, pRoles...)
	if err != nil {
		return err
	}

	return nil
}

// UpdateUser updates an existing user's information
// IF the function caller needs to have a view on the updated user object it can set the reflectObj to true
func UpdateUser(ctx context.Context, uuid, firstName, lastName, organization, description string, name string, projectList []ProjectRoles, email string, serviceRoles []string, modifiedOn time.Time, reflectObj bool, store stores.Store) (User, error) {

	prList := []stores.QProjectRoles{}

	validRoles := store.GetAllRoles(ctx)

	var duplicates []string
	// Prep project roles for datastore insert
	if projectList != nil {
		for _, item := range projectList {

			// if no name has been given for the project, skip it
			if item.Project == "" {
				continue
			}

			// check if project is encountered before by consulting duplicate list
			for _, dItem := range duplicates {
				if dItem == item.Project {
					return User{}, errors.New("duplicate reference of project " + dItem)
				}
			}

			duplicates = append(duplicates, item.Project)

			prUUID := projects.GetUUIDByName(ctx, item.Project, store)
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

	if serviceRoles != nil && len(serviceRoles) > 0 {
		for _, roleItem := range serviceRoles {
			if IsRoleValid(roleItem, validRoles) == false {
				return User{}, errors.New("invalid role: " + roleItem)
			}
		}
	}

	if err := store.UpdateUser(ctx, uuid, firstName, lastName, organization, description, prList, name, email, serviceRoles, modifiedOn); err != nil {
		return User{}, err
	}

	// reflect stored object
	if reflectObj {
		stored, err := FindUsers(ctx, "", uuid, "", true, store)
		return stored.One(), err
	}

	return User{}, nil
}

// CreateUser creates a new user
func CreateUser(ctx context.Context, uuid string, name string, fname string, lname string, org string, desc string, projectList []ProjectRoles, token string, email string, serviceRoles []string, createdOn time.Time, createdBy string, store stores.Store) (User, error) {
	// check if project with the same name exists
	if ExistsWithName(ctx, name, store) {
		return User{}, errors.New("exists")
	}

	validRoles := store.GetAllRoles(ctx)

	var duplicates []string
	// Prep project roles for datastore insert
	prList := []stores.QProjectRoles{}

	for _, item := range projectList {

		// if no name has been given for the project, skip it
		if item.Project == "" {
			continue
		}

		// check if project is encountered before by consulting duplicate list
		for _, dItem := range duplicates {
			if dItem == item.Project {
				return User{}, errors.New("duplicate reference of project " + dItem)
			}
		}

		// add project name to duplicate check list
		duplicates = append(duplicates, item.Project)

		prUUID := projects.GetUUIDByName(ctx, item.Project, store)
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

	if serviceRoles != nil && len(serviceRoles) > 0 {
		for _, roleItem := range serviceRoles {
			if IsRoleValid(roleItem, validRoles) == false {
				return User{}, errors.New("invalid role: " + roleItem)
			}
		}
	}

	if err := store.InsertUser(ctx, uuid, prList, name, fname, lname, org, desc, token, email, serviceRoles, createdOn, createdOn, createdBy); err != nil {
		return User{}, errors.New("backend error")
	}

	// reflect stored object
	stored, err := FindUsers(ctx, "", "", name, true, store)
	return stored.One(), err
}

// GenToken generates a new token
func GenToken() (string, error) {
	tokenLen := 32
	tokenBytes := make([]byte, tokenLen)
	if _, err := rand.Read(tokenBytes); err != nil {
		return "", err
	}
	sha1Bytes := sha256.Sum256(tokenBytes)
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

// IsAdminViewer checks if the user is an admon viewer
func IsAdminViewer(roles []string) bool {
	for _, role := range roles {
		if role == "admin_viewer" {
			return true
		}
	}

	return false
}

// RemoveUser removes an existing user
func RemoveUser(ctx context.Context, uuid string, store stores.Store) error {
	return store.RemoveUser(ctx, uuid)
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
func AreValidUsers(ctx context.Context, projectUUID string, users []string, store stores.Store) (bool, error) {
	found, notFound := store.HasUsers(ctx, projectUUID, users)
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
func PerResource(ctx context.Context, project string, resType string, resName string, userUUID string, store stores.Store) bool {

	if resType == "topics" || resType == "subscriptions" {
		err := store.ExistsInACL(ctx, project, resType, resName, userUUID)
		if err != nil {
			log.WithFields(
				log.Fields{
					"trace_id": ctx.Value("trace_id"),
					"type":     "system_log",
				},
			).Error(err.Error())
			return false
		}

		return true

	}
	return false
}

// Authorize based on resource and  role information
func Authorize(ctx context.Context, resource string, roles []string, store stores.Store) bool {
	// check if _admin_ is in roles
	for _, role := range roles {
		if role == "_admin_" {
			return true
		}
	}

	return store.HasResourceRoles(ctx, resource, roles)
}
