package stores

import (
	"context"
	"errors"
	"sort"
	"strconv"
	"time"
)

// MockStore holds configuration
type MockStore struct {
	Server             string
	Database           string
	UserRegistrations  []QUserRegistration
	SubList            []QSub
	TopicList          []QTopic
	DailyTopicMsgCount []QDailyTopicMsgCount
	ProjectList        []QProject
	UserList           []QUser
	RoleList           []QRole
	SchemaList         []QSchema
	Session            bool
	TopicsACL          map[string]QAcl
	SubsACL            map[string]QAcl
	OpMetrics          map[string]QopMetric
}

func (mk *MockStore) TopicsCount(ctx context.Context, startDate, endDate time.Time, projectUUIDs []string) (map[string]int64, error) {

	res := map[string]int64{}

	if len(projectUUIDs) == 0 {
		for _, tp := range mk.TopicList {
			if tp.CreatedOn.After(startDate) && tp.CreatedOn.Before(endDate) {
				res[tp.ProjectUUID] += 1
			}
		}
	} else {
		for _, tp := range mk.TopicList {
			if tp.CreatedOn.After(startDate) && tp.CreatedOn.Before(endDate) {
				for _, puuid := range projectUUIDs {
					if tp.ProjectUUID == puuid {
						res[tp.ProjectUUID] += 1
					}
				}

			}
		}
	}

	return res, nil
}

func (mk *MockStore) SubscriptionsCount(ctx context.Context, startDate, endDate time.Time, projectUUIDs []string) (map[string]int64, error) {

	res := map[string]int64{}

	if len(projectUUIDs) == 0 {
		for _, sub := range mk.SubList {
			if sub.CreatedOn.After(startDate) && sub.CreatedOn.Before(endDate) {
				res[sub.ProjectUUID] += 1
			}
		}
	} else {
		for _, sub := range mk.SubList {
			if sub.CreatedOn.After(startDate) && sub.CreatedOn.Before(endDate) {
				for _, puuid := range projectUUIDs {
					if sub.ProjectUUID == puuid {
						res[sub.ProjectUUID] += 1
					}
				}

			}
		}
	}

	return res, nil
}

func (mk *MockStore) UsersCount(ctx context.Context, startDate, endDate time.Time, projectUUIDs []string) (map[string]int64, error) {

	res := map[string]int64{}

	if len(projectUUIDs) == 0 {
		for _, user := range mk.UserList {
			if user.CreatedOn.After(startDate) && user.CreatedOn.Before(endDate) {
				for _, proj := range user.Projects {
					res[proj.ProjectUUID] += 1
				}
			}
		}
	} else {
		for _, user := range mk.UserList {
			if user.CreatedOn.After(startDate) && user.CreatedOn.Before(endDate) {
				for _, puuid := range projectUUIDs {
					for _, proj := range user.Projects {
						if puuid == proj.ProjectUUID {
							res[proj.ProjectUUID] += 1
						}
					}
				}

			}
		}
	}

	return res, nil
}

// QueryACL Topic/Subscription ACL
func (mk *MockStore) QueryACL(ctx context.Context, projectUUID string, resource string, name string) (QAcl, error) {
	if resource == "topics" {
		if _, exists := mk.TopicsACL[name]; exists {
			return mk.TopicsACL[name], nil
		}
	} else if resource == "subscriptions" {
		if _, exists := mk.SubsACL[name]; exists {
			return mk.SubsACL[name], nil
		}
	}

	return QAcl{}, errors.New("not found")
}

// NewMockStore creates new mock store
func NewMockStore(server string, database string) *MockStore {
	mk := MockStore{}
	mk.Server = server
	mk.Database = database
	mk.Session = true
	mk.Initialize()
	return &mk
}

// InsertOpMetric inserts a new operation metric
func (mk *MockStore) InsertOpMetric(ctx context.Context, hostname string, cpu float64, mem float64) error {
	qOp := QopMetric{hostname, cpu, mem}
	mk.OpMetrics[hostname] = qOp
	return nil
}

// Close is used to close session
func (mk *MockStore) Close() {
	mk.Session = false
}

// InsertUser inserts a new user to the store
func (mk *MockStore) InsertUser(ctx context.Context, uuid string, projects []QProjectRoles, name string, fname string, lname string, org string, desc string, token string, email string, serviceRoles []string, createdOn time.Time, modifiedOn time.Time, createdBy string) error {
	user := QUser{
		UUID:         uuid,
		Name:         name,
		Email:        email,
		Token:        token,
		FirstName:    fname,
		LastName:     lname,
		Organization: org,
		Description:  desc,
		Projects:     projects,
		ServiceRoles: serviceRoles,
		CreatedOn:    createdOn,
		ModifiedOn:   modifiedOn,
		CreatedBy:    createdBy,
	}
	mk.UserList = append(mk.UserList, user)
	return nil
}

func (mk *MockStore) RegisterUser(ctx context.Context, uuid, name, firstName, lastName, email, org, desc, registeredAt, atkn, status string) error {

	ur := QUserRegistration{
		UUID:            uuid,
		Name:            name,
		FirstName:       firstName,
		LastName:        lastName,
		Email:           email,
		Organization:    org,
		Description:     desc,
		RegisteredAt:    registeredAt,
		ActivationToken: atkn,
		Status:          status,
	}

	mk.UserRegistrations = append(mk.UserRegistrations, ur)
	return nil
}

func (mk *MockStore) QueryRegistrations(ctx context.Context, regUUID, status, activationToken, name, email, org string) ([]QUserRegistration, error) {

	if regUUID == "" && status == "" && activationToken == "" && name == "" && email == "" && org == "" {
		return mk.UserRegistrations, nil
	}

	reqs := []QUserRegistration{}

	if regUUID != "" {
		for idx, req := range mk.UserRegistrations {
			if req.UUID == regUUID {
				reqs = append(reqs, mk.UserRegistrations[idx])
				continue
			}
		}
	} else {
		reqs = mk.UserRegistrations
	}

	if status != "" {
		tempReqs := []QUserRegistration{}
		for idx, req := range reqs {
			if req.Status == status {
				tempReqs = append(tempReqs, reqs[idx])
			}
		}
		reqs = tempReqs
	}

	if activationToken != "" {
		tempReqs := []QUserRegistration{}
		for idx, req := range reqs {
			if req.ActivationToken == activationToken {
				tempReqs = append(tempReqs, reqs[idx])
			}
		}
		reqs = tempReqs
	}

	if name != "" {
		tempReqs := []QUserRegistration{}
		for idx, req := range reqs {
			if req.Name == name {
				tempReqs = append(tempReqs, reqs[idx])
			}
		}
		reqs = tempReqs
	}

	if email != "" {
		tempReqs := []QUserRegistration{}
		for idx, req := range reqs {
			if req.Email == email {
				tempReqs = append(tempReqs, reqs[idx])
			}
		}
		reqs = tempReqs
	}

	if org != "" {
		tempReqs := []QUserRegistration{}
		for idx, req := range reqs {
			if req.Organization == org {
				tempReqs = append(tempReqs, reqs[idx])
			}
		}
		reqs = tempReqs
	}

	return reqs, nil
}

func (mk *MockStore) UpdateRegistration(ctx context.Context, regUUID, status, declineComment, modifiedBy, modifiedAt string) error {

	for idx, ur := range mk.UserRegistrations {
		if ur.UUID == regUUID {
			mk.UserRegistrations[idx].Status = status
			mk.UserRegistrations[idx].DeclineComment = declineComment
			mk.UserRegistrations[idx].ModifiedBy = modifiedBy
			mk.UserRegistrations[idx].ModifiedAt = modifiedAt
			mk.UserRegistrations[idx].ActivationToken = ""
		}
	}

	return nil
}

// GetAllRoles returns a list of all available roles
func (mk *MockStore) GetAllRoles(ctx context.Context) []string {
	return []string{"service_admin", "admin", "project_admin", "viewer", "consumer", "producer", "publisher", "push_worker"}
}

// UpdateUserToken updates user's token
func (mk *MockStore) UpdateUserToken(ctx context.Context, uuid string, token string) error {
	for i, item := range mk.UserList {
		if item.UUID == uuid {
			mk.UserList[i].Token = token
			return nil
		}
	}

	return errors.New("not found")

}

// GetOpMetrics returns operation metrics
func (mk *MockStore) GetOpMetrics(ctx context.Context) []QopMetric {
	results := []QopMetric{}
	for _, v := range mk.OpMetrics {
		results = append(results, v)
	}
	return results
}

// AppendToUserProjects adds project and specific roles to a users role list
func (mk *MockStore) AppendToUserProjects(ctx context.Context, userUUID string, projectUUID string, pRoles ...string) error {

	for idx, user := range mk.UserList {

		if user.UUID == userUUID {
			projectFound := false

			for _, p := range user.Projects {

				if p.ProjectUUID == projectUUID {
					projectFound = true
					break
				}
			}

			if !projectFound {
				mk.UserList[idx].Projects = append(mk.UserList[idx].Projects, QProjectRoles{
					ProjectUUID: projectUUID,
					Roles:       pRoles,
				})
			}

			break
		}
	}

	return nil

}

// UpdateUser updates user information
func (mk *MockStore) UpdateUser(ctx context.Context, uuid, fname, lname, org, desc string, projects []QProjectRoles, name string, email string, serviceRoles []string, modifiedOn time.Time) error {

	for i, item := range mk.UserList {
		if item.UUID == uuid {
			if projects != nil {
				mk.UserList[i].Projects = projects
			}
			if serviceRoles != nil {
				mk.UserList[i].ServiceRoles = serviceRoles
			}
			if name != "" {
				mk.UserList[i].Name = name
			}
			if email != "" {
				mk.UserList[i].Email = email
			}

			if fname != "" {
				mk.UserList[i].FirstName = fname
			}

			if lname != "" {
				mk.UserList[i].LastName = lname
			}

			if org != "" {
				mk.UserList[i].Organization = org
			}

			if desc != "" {
				mk.UserList[i].Description = desc
			}

			mk.UserList[i].ModifiedOn = modifiedOn

			return nil
		}
	}

	return errors.New("not found")

}

// HasUsers accepts a user array of usernames and returns the not found
func (mk *MockStore) HasUsers(ctx context.Context, projectUUID string, users []string) (bool, []string) {

	var notFound []string

	// for each given username
	for _, username := range users {
		found := false
		// loop through all found users
		for _, user := range mk.UserList {
			if username == user.Name {
				found = true
			}
		}
		// if not found add it to the notFound
		if !found {
			notFound = append(notFound, username)
		}

	}

	return len(notFound) == 0, notFound
}

// ModACL changes the acl in a function
func (mk *MockStore) ModACL(ctx context.Context, projectUUID string, resource string, name string, acl []string) error {
	newACL := QAcl{ACL: acl}
	if resource == "topics" {
		if _, exists := mk.TopicsACL[name]; exists {
			mk.TopicsACL[name] = newACL
			return nil
		}
	} else if resource == "subscriptions" {
		if _, exists := mk.SubsACL[name]; exists {
			mk.SubsACL[name] = newACL
			return nil
		}
	}

	return errors.New("wrong resource type")
}

// AppendToACL adds given users to an existing ACL
func (mk *MockStore) AppendToACL(ctx context.Context, projectUUID string, resource string, name string, acl []string) error {
	if resource == "topics" {
		if qACL, exists := mk.TopicsACL[name]; exists {
			qACL.ACL = appendUniqueValues(qACL.ACL, acl...)
			mk.TopicsACL[name] = qACL
			return nil
		}
	} else if resource == "subscriptions" {
		if qACL, exists := mk.SubsACL[name]; exists {
			qACL.ACL = appendUniqueValues(qACL.ACL, acl...)
			mk.SubsACL[name] = qACL
			return nil
		}
	} else {
		return errors.New("wrong resource type")
	}

	return errors.New("no acl found")
}

func appendUniqueValues(existingValues []string, newValues ...string) []string {
	for _, value := range newValues {
		found := false
		for _, ev := range existingValues {
			if ev == value {
				found = true
				break
			}
		}
		if !found {
			existingValues = append(existingValues, value)
		}
	}
	return existingValues
}

// RemoveFromACL removes given users from an existing acl
func (mk *MockStore) RemoveFromACL(ctx context.Context, projectUUID string, resource string, name string, acl []string) error {
	if resource == "topics" {
		if qACL, exists := mk.TopicsACL[name]; exists {
			qACL.ACL = removeValues(qACL.ACL, acl...)
			mk.TopicsACL[name] = qACL
			return nil
		}
	} else if resource == "subscriptions" {
		if qACL, exists := mk.SubsACL[name]; exists {
			qACL.ACL = removeValues(qACL.ACL, acl...)
			mk.SubsACL[name] = qACL
			return nil
		}
	} else {
		return errors.New("wrong resource type")
	}

	return errors.New("no acl found")
}

func removeValues(existingValues []string, valuesToRemove ...string) []string {

	for _, value := range valuesToRemove {
		existingValues = removeSingleValue(existingValues, value)
	}

	return existingValues
}

func removeSingleValue(existingValues []string, valueToRemove string) []string {

	for idx, value := range existingValues {
		if value == valueToRemove {
			existingValues = append(existingValues[:idx], existingValues[idx+1:]...)
		}
	}

	return existingValues
}

// UpdateProject updates project information
func (mk *MockStore) UpdateProject(ctx context.Context, projectUUID string, name string, description string, modifiedOn time.Time) error {

	for i, item := range mk.ProjectList {
		if item.UUID == projectUUID {
			if description != "" {
				mk.ProjectList[i].Description = description
			}
			if name != "" {
				mk.ProjectList[i].Name = name
			}

			mk.ProjectList[i].ModifiedOn = modifiedOn
			return nil
		}
	}

	return errors.New("not found")

}

// QueryDailyProjectMsgCount retrieves the number of total messages that have been published to all project's topics daily
func (mk *MockStore) QueryDailyProjectMsgCount(ctx context.Context, projectUUID string) ([]QDailyProjectMsgCount, error) {

	var qDps []QDailyProjectMsgCount
	var ok bool
	var msgs int64

	var msgCounts = make(map[time.Time]int64)

	// group the number of messages by date
	for _, dp := range mk.DailyTopicMsgCount {

		if dp.ProjectUUID == projectUUID {

			if msgs, ok = msgCounts[dp.Date]; ok {
				msgCounts[dp.Date] = msgs + dp.NumberOfMessages
			} else {
				msgCounts[dp.Date] = dp.NumberOfMessages
			}

		}

	}

	for key, value := range msgCounts {

		qDps = append(qDps, QDailyProjectMsgCount{key, value})
	}

	// sort in descending order
	sort.Slice(qDps, func(i, j int) bool { return qDps[i].Date.After(qDps[j].Date) })

	return qDps, nil
}

// IncrementTopicMsgNum increase number of messages published in a topic
func (mk *MockStore) IncrementTopicMsgNum(ctx context.Context, projectUUID string, name string, num int64) error {

	for i, item := range mk.TopicList {
		if item.ProjectUUID == projectUUID && item.Name == name {
			mk.TopicList[i].MsgNum += num
			return nil
		}
	}

	return errors.New("not found")
}

// IncrementDailyTopicMsgCount increase number of messages published in a topic
func (mk *MockStore) IncrementDailyTopicMsgCount(ctx context.Context, projectUUID string, topicName string, num int64, date time.Time) error {

	for i, item := range mk.DailyTopicMsgCount {
		if item.ProjectUUID == projectUUID && item.TopicName == topicName && item.Date.Equal(date) {
			mk.DailyTopicMsgCount[i].NumberOfMessages += num
			return nil
		}
	}

	mk.DailyTopicMsgCount = append(mk.DailyTopicMsgCount, QDailyTopicMsgCount{Date: date, ProjectUUID: projectUUID, TopicName: topicName, NumberOfMessages: num})
	return nil
}

// IncrementTopicBytes increases the total number of bytes published in a topic
func (mk *MockStore) IncrementTopicBytes(ctx context.Context, projectUUID string, name string, totalBytes int64) error {
	for i, item := range mk.TopicList {
		if item.ProjectUUID == projectUUID && item.Name == name {
			mk.TopicList[i].TotalBytes += totalBytes
			return nil
		}
	}

	return errors.New("not found")
}

// IncrementSubBytes increases the total number of bytes published in a subscription
func (mk *MockStore) IncrementSubBytes(ctx context.Context, projectUUID string, name string, totalBytes int64) error {
	for i, item := range mk.SubList {
		if item.ProjectUUID == projectUUID && item.Name == name {
			mk.SubList[i].TotalBytes += totalBytes
			return nil
		}
	}

	return errors.New("not found")
}

// IncrementSubMsgNum increase number of messages pulled in a subscription
func (mk *MockStore) IncrementSubMsgNum(ctx context.Context, projectUUID string, name string, num int64) error {

	for i, item := range mk.SubList {
		if item.ProjectUUID == projectUUID && item.Name == name {
			mk.SubList[i].MsgNum += num
			return nil
		}
	}

	return errors.New("not found")
}

// UpdateSubOffset updates the offset of the current subscription
func (mk *MockStore) UpdateSubOffset(ctx context.Context, projectUUID string, name string, offset int64) {

}

// ModAck modifies the subscription ack
func (mk *MockStore) ModAck(ctx context.Context, projectUUID string, name string, ack int) error {
	for i, item := range mk.SubList {
		if item.ProjectUUID == projectUUID && item.Name == name {
			mk.SubList[i].Ack = ack

			return nil
		}
	}

	return errors.New("not found")
}

// ModSubPush modifies the subscription push configuration
func (mk *MockStore) ModSubPush(ctx context.Context, projectUUID string, name string, config QPushConfig) error {
	for i, item := range mk.SubList {
		if item.ProjectUUID == projectUUID && item.Name == name {
			mk.SubList[i].PushType = config.Type
			mk.SubList[i].PushEndpoint = config.PushEndpoint
			mk.SubList[i].AuthorizationType = config.AuthorizationType
			mk.SubList[i].AuthorizationHeader = config.AuthorizationHeader
			mk.SubList[i].MaxMessages = config.MaxMessages
			mk.SubList[i].RetPolicy = config.RetPolicy
			mk.SubList[i].RetPeriod = config.RetPeriod
			mk.SubList[i].VerificationHash = config.VerificationHash
			mk.SubList[i].Verified = config.Verified
			mk.SubList[i].MattermostUrl = config.MattermostUrl
			mk.SubList[i].MattermostUsername = config.MattermostUsername
			mk.SubList[i].MattermostChannel = config.MattermostChannel
			return nil
		}
	}
	return errors.New("not found")
}

// UpdateSubOffsetAck updates the offset of the current subscription
func (mk *MockStore) UpdateSubOffsetAck(ctx context.Context, projectUUID string, name string, offset int64, ts string) error {
	// find sub
	sub := QSub{}

	for _, item := range mk.SubList {
		if item.ProjectUUID == projectUUID && item.Name == name {
			sub = item
		}
	}

	// check if no ack pending
	if sub.NextOffset == 0 {
		return errors.New("no ack pending")
	}

	// check if ack offset is wrong - wrong ack
	if offset < sub.Offset || offset > sub.NextOffset {
		return errors.New("wrong ack")
	}

	// check if ack has timeout
	zSec := "2006-01-02T15:04:05Z"
	timeGiven, _ := time.Parse(zSec, ts)
	timeRef, _ := time.Parse(zSec, sub.PendingAck)
	durSec := timeGiven.Sub(timeRef).Seconds()

	if int(durSec) > sub.Ack {
		return errors.New("ack timeout")
	}

	return nil
}

// QueryProjects function queries for a specific project or for a list of all projects
func (mk *MockStore) QueryProjects(ctx context.Context, uuid string, name string) ([]QProject, error) {

	result := []QProject{}
	if name == "" && uuid == "" {
		result = mk.ProjectList
	} else if name != "" {
		for _, item := range mk.ProjectList {
			if item.Name == name {
				result = append(result, item)
				break
			}
		}
	} else if uuid != "" {
		for _, item := range mk.ProjectList {

			if item.UUID == uuid {
				result = append(result, item)
				break
			}
		}
	}

	if len(result) > 0 {
		return result, nil
	}

	return result, errors.New("not found")

}

// QueryUsers queries the datastore for user information
func (mk *MockStore) QueryUsers(ctx context.Context, projectUUID string, uuid string, name string) ([]QUser, error) {
	result := []QUser{}

	if name == "" && uuid == "" && projectUUID == "" {
		for _, item := range mk.UserList {
			result = append(result, item)
		}
	} else if name == "" && uuid == "" && projectUUID != "" {
		for _, item := range mk.UserList {
			if item.isInProject(projectUUID) {
				result = append(result, item)
			}
		}
	} else if uuid != "" {
		for _, item := range mk.UserList {
			if item.UUID == uuid {
				result = append(result, item)
			}
		}
	} else if name != "" {
		for _, item := range mk.UserList {
			if item.Name == name {
				result = append(result, item)
				break
			}
		}
	}

	if len(result) > 0 {
		return result, nil
	}

	return result, errors.New("not found")

}

// PaginatedQueryUsers provides query to the list of users using pagination parameters
func (mk *MockStore) PaginatedQueryUsers(ctx context.Context, pageToken string, pageSize int64, projectUUID string) ([]QUser, int64, string, error) {

	var qUsers []QUser
	var nextPageToken string
	var err error
	var pg int
	var limit int

	totalSize := int64(0)

	if pageSize == 0 {
		limit = len(mk.UserList)
	} else {
		limit = int(pageSize) + 1
	}

	if pageToken != "" {
		if pg, err = strconv.Atoi(pageToken); err != nil {
			return qUsers, totalSize, nextPageToken, err
		}
	}

	sort.Slice(mk.UserList, func(i, j int) bool {
		id1 := mk.UserList[i].ID.(int)
		id2 := mk.UserList[j].ID.(int)
		return id1 > id2
	})

	for _, user := range mk.UserList {

		if projectUUID != "" {
			found := false
			for _, project := range user.Projects {

				if projectUUID == project.ProjectUUID {
					found = true
				}
			}
			if !found {
				continue
			}

		}

		if limit == 0 {
			break
		}

		if pageToken != "" {

			if user.ID.(int) <= pg {

				qUsers = append(qUsers, user)
				limit--

			}

		} else {

			qUsers = append(qUsers, user)
			limit--

		}

	}

	if pageSize > 0 && len(qUsers) > 0 && len(qUsers) == int(pageSize)+1 {
		nextPageToken = strconv.Itoa(qUsers[int(pageSize)].ID.(int))
		qUsers = qUsers[:len(qUsers)-1]
	}

	if projectUUID == "" {
		totalSize = int64(len(qUsers))
	} else {
		for _, user := range mk.UserList {
			for _, project := range user.Projects {
				if projectUUID == project.ProjectUUID {
					totalSize++
				}
			}
		}
	}

	return qUsers, totalSize, nextPageToken, err

}

// UpdateSubPull updates next offset info after a pull
func (mk *MockStore) UpdateSubPull(ctx context.Context, projectUUID string, name string, offset int64, ts string) error {
	for i, item := range mk.SubList {
		if item.ProjectUUID == projectUUID && item.Name == name {
			mk.SubList[i].NextOffset = offset
			mk.SubList[i].PendingAck = ts
			return nil
		}
	}
	return errors.New("not found")

}

// Initialize is used to initialize the mock
func (mk *MockStore) Initialize() {
	mk.OpMetrics = make(map[string]QopMetric)

	// populate topics
	qtop4 := QTopic{3, "argo_uuid", "topic4", 0, 0, time.Date(0, 0, 0, 0, 0, 0, 0, time.UTC), 0, "", time.Date(2020, 11, 19, 0, 0, 0, 0, time.UTC), []string{}}
	qtop3 := QTopic{2, "argo_uuid", "topic3", 0, 0, time.Date(2019, 5, 7, 0, 0, 0, 0, time.UTC), 8.99, "schema_uuid_3", time.Date(2020, 11, 20, 0, 0, 0, 0, time.UTC), []string{}}
	qtop2 := QTopic{1, "argo_uuid", "topic2", 0, 0, time.Date(2019, 5, 8, 0, 0, 0, 0, time.UTC), 5.45, "schema_uuid_1", time.Date(2020, 11, 21, 0, 0, 0, 0, time.UTC), []string{}}
	qtop1 := QTopic{0, "argo_uuid", "topic1", 0, 0, time.Date(2019, 5, 6, 0, 0, 0, 0, time.UTC), 10, "", time.Date(2020, 11, 22, 0, 0, 0, 0, time.UTC), []string{}}
	mk.TopicList = append(mk.TopicList, qtop1)
	mk.TopicList = append(mk.TopicList, qtop2)
	mk.TopicList = append(mk.TopicList, qtop3)
	mk.TopicList = append(mk.TopicList, qtop4)

	// populate Subscriptions
	qsub1 := QSub{0, "argo_uuid", "sub1", "topic1", 0, 0, "",
		"", "", 0, "", "", 10, "",
		0, 0, 0, "", false, "", "", "", false,
		time.Date(2019, 5, 6, 0, 0, 0, 0, time.UTC),
		10, time.Date(2020, 11, 19, 0, 0, 0, 0, time.UTC), []string{}}

	qsub2 := QSub{1, "argo_uuid", "sub2", "topic2", 0, 0, "",
		"", "", 0, "", "", 10, "",
		0, 0, 0, "", false, "", "", "", false,
		time.Date(2019, 5, 7, 0, 0, 0, 0, time.UTC),
		8.99, time.Date(2020, 11, 20, 0, 0, 0, 0, time.UTC), []string{}}

	qsub3 := QSub{2, "argo_uuid", "sub3", "topic3", 0, 0, "",
		"", "", 0, "", "", 10, "",
		0, 0, 0, "", false, "", "", "", false,
		time.Date(2019, 5, 8, 0, 0, 0, 0, time.UTC),
		5.45, time.Date(2020, 11, 21, 0, 0, 0, 0, time.UTC), []string{}}

	qsub4 := QSub{3, "argo_uuid", "sub4", "topic4", 0, 0, "",
		"http_endpoint", "endpoint.foo", 1, "autogen",
		"auth-header-1", 10, "linear", 300, 0, 0,
		"push-id-1", true, "", "", "", true,
		time.Date(0, 0, 0, 0, 0, 0, 0, time.UTC),
		0, time.Date(2020, 11, 22, 0, 0, 0, 0, time.UTC), []string{}}
	mk.SubList = append(mk.SubList, qsub1)
	mk.SubList = append(mk.SubList, qsub2)
	mk.SubList = append(mk.SubList, qsub3)
	mk.SubList = append(mk.SubList, qsub4)

	// populate Projects
	created := time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC)
	modified := time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC)
	qPr := QProject{UUID: "argo_uuid", Name: "ARGO", CreatedOn: created, ModifiedOn: modified, CreatedBy: "uuid1", Description: "simple project"}
	qPr2 := QProject{UUID: "argo_uuid2", Name: "ARGO2", CreatedOn: created, ModifiedOn: modified, CreatedBy: "uuid1", Description: "simple project"}
	mk.ProjectList = append(mk.ProjectList, qPr)
	mk.ProjectList = append(mk.ProjectList, qPr2)

	// populate schemas
	//{
	// 			"type": "object",
	//			 "properties": {
	// 			  "name":        { "type": "string" },
	//  			  "email":        { "type": "string" },
	// 			  "address":    { "type": "string" },
	//  			  "telephone": { "type": "string" }
	//	 },
	// 	"required": ["name", "email"]
	//}
	//}
	// the above schema base64 encoded
	s := "eyJwcm9wZXJ0aWVzIjp7ImFkZHJlc3MiOnsidHlwZSI6InN0cmluZyJ9LCJlbWFpbCI6eyJ0eXBlIjoic3RyaW5nIn0sIm5hbWUiOnsidHlwZSI6InN0cmluZyJ9LCJ0ZWxlcGhvbmUiOnsidHlwZSI6InN0cmluZyJ9fSwicmVxdWlyZWQiOlsibmFtZSIsImVtYWlsIl0sInR5cGUiOiJvYmplY3QifQ=="
	qSchema1 := QSchema{UUID: "schema_uuid_1", ProjectUUID: "argo_uuid", Type: "json", Name: "schema-1", RawSchema: s}
	qSchema2 := QSchema{UUID: "schema_uuid_2", ProjectUUID: "argo_uuid", Type: "json", Name: "schema-2", RawSchema: s}
	// {
	//		"namespace": "user.avro",
	//		"type": "record",
	//		"name": "User",
	//		"fields": [
	//		{"name": "username", "type":"string"},
	//		{"name": "phone", "type": "int"}
	//      ]
	// }
	avros := "eyJmaWVsZHMiOlt7Im5hbWUiOiJ1c2VybmFtZSIsInR5cGUiOiJzdHJpbmcifSx7Im5hbWUiOiJwaG9uZSIsInR5cGUiOiJpbnQifV0sIm5hbWUiOiJVc2VyIiwibmFtZXNwYWNlIjoidXNlci5hdnJvIiwidHlwZSI6InJlY29yZCJ9"
	qSchema3 := QSchema{UUID: "schema_uuid_3", ProjectUUID: "argo_uuid", Type: "avro", Name: "schema-3", RawSchema: avros}
	mk.SchemaList = append(mk.SchemaList, qSchema1, qSchema2, qSchema3)

	// populate daily msg count for topics
	dc1 := QDailyTopicMsgCount{time.Date(2018, 10, 1, 0, 0, 0, 0, time.UTC), "argo_uuid", "topic1", 40}
	dc2 := QDailyTopicMsgCount{time.Date(2018, 10, 2, 0, 0, 0, 0, time.UTC), "argo_uuid", "topic1", 30}
	dc3 := QDailyTopicMsgCount{time.Date(2018, 10, 1, 0, 0, 0, 0, time.UTC), "argo_uuid", "topic2", 70}
	dc4 := QDailyTopicMsgCount{time.Date(2018, 10, 1, 0, 0, 0, 0, time.UTC), "argo_uuid", "topic3", 0}
	mk.DailyTopicMsgCount = append(mk.DailyTopicMsgCount, dc1, dc2, dc3, dc4)

	// populate Users
	qRole := []QProjectRoles{QProjectRoles{"argo_uuid", []string{"consumer", "publisher"}}}
	qRoleB := []QProjectRoles{QProjectRoles{"argo_uuid2", []string{"consumer", "publisher"}}}
	qUsr := QUser{0, "uuid0", qRole, "Test", "", "", "", "", "S3CR3T", "Test@test.com", []string{}, created, modified, ""}

	mk.UserList = append(mk.UserList, qUsr)

	qRoleConsumerPub := []QProjectRoles{QProjectRoles{"argo_uuid", []string{"publisher", "consumer"}}}

	mk.UserList = append(mk.UserList, QUser{1, "uuid1", qRole, "UserA", "FirstA", "LastA", "OrgA", "DescA", "S3CR3T1", "foo-email", []string{}, created, modified, ""})
	mk.UserList = append(mk.UserList, QUser{2, "uuid2", qRole, "UserB", "", "", "", "", "S3CR3T2", "foo-email", []string{}, created, modified, "uuid1"})
	mk.UserList = append(mk.UserList, QUser{3, "uuid3", qRoleConsumerPub, "UserX", "", "", "", "", "S3CR3T3", "foo-email", []string{}, created, modified, "uuid1"})
	mk.UserList = append(mk.UserList, QUser{4, "uuid4", qRoleConsumerPub, "UserZ", "", "", "", "", "S3CR3T4", "foo-email", []string{}, created, modified, "uuid1"})
	mk.UserList = append(mk.UserList, QUser{5, "same_uuid", qRoleConsumerPub, "UserSame1", "", "", "", "", "S3CR3T41", "foo-email", []string{}, created, modified, "uuid1"})
	mk.UserList = append(mk.UserList, QUser{6, "same_uuid", qRoleConsumerPub, "UserSame2", "", "", "", "", "S3CR3T42", "foo-email", []string{}, created, modified, "uuid1"})
	mk.UserList = append(mk.UserList, QUser{7, "uuid7", []QProjectRoles{}, "push_worker_0", "", "", "", "", "push_token", "foo-email", []string{"push_worker"}, created, modified, ""})
	mk.UserList = append(mk.UserList, QUser{8, "uuid8", qRoleB, "UserZ", "", "", "", "", "S3CR3T1", "foo-email", []string{}, created, modified, ""})

	qRole1 := QRole{"topics:list_all", []string{"admin", "reader", "publisher"}}
	qRole2 := QRole{"topics:publish", []string{"admin", "publisher"}}
	mk.RoleList = append(mk.RoleList, qRole1)
	mk.RoleList = append(mk.RoleList, qRole2)

	qTopicACL01 := QAcl{[]string{"uuid1", "uuid2"}}
	qTopicACL02 := QAcl{[]string{"uuid1", "uuid2", "uuid4"}}
	qTopicACL03 := QAcl{[]string{"uuid3"}}

	qSubACL01 := QAcl{[]string{"uuid1", "uuid2"}}
	qSubACL02 := QAcl{[]string{"uuid1", "uuid3"}}
	qSubACL03 := QAcl{[]string{"uuid4", "uuid2", "uuid1"}}
	qSubACL04 := QAcl{[]string{"uuid2", "uuid4", "uuid7"}}

	mk.TopicsACL = make(map[string]QAcl)
	mk.SubsACL = make(map[string]QAcl)

	mk.TopicsACL["topic1"] = qTopicACL01
	mk.TopicsACL["topic2"] = qTopicACL02
	mk.TopicsACL["topic3"] = qTopicACL03

	mk.SubsACL["sub1"] = qSubACL01
	mk.SubsACL["sub2"] = qSubACL02
	mk.SubsACL["sub3"] = qSubACL03
	mk.SubsACL["sub4"] = qSubACL04

	// Populate user registrations
	ur1 := QUserRegistration{
		UUID:            "ur-uuid1",
		Name:            "urname",
		FirstName:       "urfname",
		LastName:        "urlname",
		Organization:    "urorg",
		Description:     "urdesc",
		Email:           "uremail",
		ActivationToken: "uratkn-1",
		Status:          "pending",
		RegisteredAt:    "2019-05-12T22:26:58Z",
		ModifiedBy:      "uuid1",
		ModifiedAt:      "2020-05-15T22:26:58Z",
	}

	mk.UserRegistrations = append(mk.UserRegistrations, ur1)
}

func (mk *MockStore) InsertResourceRoles(ctx context.Context, resource string, roles []string) error {
	mk.RoleList = append(mk.RoleList, QRole{Name: resource, Roles: roles})
	return nil
}

func (mk *MockStore) QueryTotalMessagesPerProject(ctx context.Context, projectUUIDs []string, startDate time.Time, endDate time.Time) ([]QProjectMessageCount, error) {

	projectCount := make(map[string]int64)

	qpc := make([]QProjectMessageCount, 0)

	if endDate.Before(startDate) {
		startDate, endDate = endDate, startDate
	}

	days := int64(1)
	if !endDate.Equal(startDate) {
		days = int64(endDate.Sub(startDate).Hours() / 24)
		// add an extra day to compensate for the fact that we need the starting day included as well
		// e.g. Aug 1 to Aug 31 should be calculated as 31 days and not as 30
		days += 1
	}

	if len(projectUUIDs) == 0 {
		for _, c := range mk.DailyTopicMsgCount {
			if c.Date.After(startDate) && c.Date.Before(endDate) {
				count, ok := projectCount[c.ProjectUUID]
				if ok {
					projectCount[c.ProjectUUID] = c.NumberOfMessages + count
				} else {
					projectCount[c.ProjectUUID] = c.NumberOfMessages
				}
			}
		}
	} else {
		for _, pUUID := range projectUUIDs {
			for _, c := range mk.DailyTopicMsgCount {
				if pUUID == c.ProjectUUID && c.Date.After(startDate) && c.Date.Before(endDate) {
					count, ok := projectCount[c.ProjectUUID]
					if ok {
						projectCount[c.ProjectUUID] = c.NumberOfMessages + count
					} else {
						projectCount[c.ProjectUUID] = c.NumberOfMessages
					}
				}
			}
		}
	}

	for k, v := range projectCount {
		qpc = append(qpc, QProjectMessageCount{
			ProjectUUID:          k,
			NumberOfMessages:     v,
			AverageDailyMessages: float64(v / days),
		})
	}

	return qpc, nil
}

// QueryOneSub returns one sub exactly
func (mk *MockStore) QueryOneSub(ctx context.Context, projectUUID string, name string) (QSub, error) {
	for _, item := range mk.SubList {
		if item.Name == name && item.ProjectUUID == projectUUID {
			return item, nil
		}
	}

	return QSub{}, errors.New("empty")
}

// Clone the store
func (mk *MockStore) Clone() Store {
	return mk
}

// GetUserFromToken retrieves specific user info from a given token
func (mk *MockStore) GetUserFromToken(ctx context.Context, token string) (QUser, error) {
	for _, item := range mk.UserList {

		if item.Token == token {
			return item, nil

		}
	}

	return QUser{}, errors.New("not found")

}

// GetUserRoles returns the roles of a user in a project
func (mk *MockStore) GetUserRoles(ctx context.Context, projectUUID string, token string) ([]string, string) {
	for _, item := range mk.UserList {

		if item.Token == token {
			return item.getProjectRoles(projectUUID), item.Name

		}
	}

	return []string{}, ""
}

// HasResourceRoles returns the roles of a user in a project
func (mk *MockStore) HasResourceRoles(ctx context.Context, resource string, roles []string) bool {

	for _, item := range mk.RoleList {
		if item.Name == resource {
			for _, subitem := range item.Roles {
				for _, roleItem := range roles {
					if roleItem == subitem {
						return true
					}
				}
			}
		}

	}

	return false

}

// HasProject returns true if project exists in store
func (mk *MockStore) HasProject(ctx context.Context, name string) bool {
	for _, item := range mk.ProjectList {
		if item.Name == name {
			return true
		}
	}

	return false
}

// InsertTopic inserts a new topic object to the store
func (mk *MockStore) InsertTopic(ctx context.Context, projectUUID string, name string, schemaUUID string, createdOn time.Time) error {
	topic := QTopic{
		ID:            len(mk.TopicList),
		ProjectUUID:   projectUUID,
		Name:          name,
		MsgNum:        0,
		TotalBytes:    0,
		LatestPublish: time.Time{},
		PublishRate:   0,
		SchemaUUID:    schemaUUID,
		CreatedOn:     createdOn,
		ACL:           []string{},
	}
	mk.TopicList = append(mk.TopicList, topic)
	return nil
}

func (mk *MockStore) LinkTopicSchema(ctx context.Context, projectUUID, name, schemaUUID string) error {
	for idx, topic := range mk.TopicList {
		if topic.Name == name && topic.ProjectUUID == projectUUID {
			mk.TopicList[idx].SchemaUUID = schemaUUID
			break
		}
	}
	return nil
}

// InsertSub inserts a new sub object to the store
func (mk *MockStore) InsertSub(ctx context.Context, projectUUID string, name string, topic string,
	offset int64, ack int, pushCfg QPushConfig, createdOn time.Time) error {
	sub := QSub{
		ID:                  len(mk.SubList),
		ProjectUUID:         projectUUID,
		Name:                name,
		Topic:               topic,
		Offset:              offset,
		NextOffset:          0,
		PendingAck:          "",
		Ack:                 ack,
		PushType:            pushCfg.Type,
		MaxMessages:         pushCfg.MaxMessages,
		AuthorizationType:   pushCfg.AuthorizationType,
		AuthorizationHeader: pushCfg.AuthorizationHeader,
		PushEndpoint:        pushCfg.PushEndpoint,
		RetPolicy:           pushCfg.RetPolicy,
		RetPeriod:           pushCfg.RetPeriod,
		VerificationHash:    pushCfg.VerificationHash,
		Verified:            pushCfg.Verified,
		MattermostUrl:       pushCfg.MattermostUrl,
		MattermostChannel:   pushCfg.MattermostChannel,
		MattermostUsername:  pushCfg.MattermostUsername,
		Base64Decode:        pushCfg.Base64Decode,
		MsgNum:              0,
		TotalBytes:          0,
		CreatedOn:           createdOn,
		ACL:                 []string{},
	}

	mk.SubList = append(mk.SubList, sub)
	mk.SubsACL[name] = QAcl{}
	return nil
}

// InsertProject inserts a project to the store
func (mk *MockStore) InsertProject(ctx context.Context, uuid string, name string, createdOn time.Time, modifiedOn time.Time, createdBy string, description string) error {
	project := QProject{UUID: uuid, Name: name, CreatedOn: createdOn, ModifiedOn: modifiedOn, CreatedBy: createdBy, Description: description}
	mk.ProjectList = append(mk.ProjectList, project)
	return nil
}

// RemoveProject removes an existing project
func (mk *MockStore) RemoveProject(ctx context.Context, uuid string) error {
	for i, project := range mk.ProjectList {
		if project.UUID == uuid {
			// found item at i, remove it using index
			mk.ProjectList = append(mk.ProjectList[:i], mk.ProjectList[i+1:]...)
			return nil
		}
	}

	return errors.New("not found")
}

// RemoveTopic removes an existing topic
func (mk *MockStore) RemoveTopic(ctx context.Context, projectUUID string, name string) error {
	for i, topic := range mk.TopicList {
		if topic.Name == name && topic.ProjectUUID == projectUUID {
			// found item at i, remove it using index
			mk.TopicList = append(mk.TopicList[:i], mk.TopicList[i+1:]...)
			return nil
		}
	}

	return errors.New("not found")
}

// RemoveUser removes an existing user
func (mk *MockStore) RemoveUser(ctx context.Context, uuid string) error {
	for i, user := range mk.UserList {
		if user.UUID == uuid {
			// found item at i, remove it using index
			mk.UserList = append(mk.UserList[:i], mk.UserList[i+1:]...)
			return nil
		}
	}

	return errors.New("not found")
}

// RemoveProjectTopics removes all topics belonging to a specific project uuid
func (mk *MockStore) RemoveProjectTopics(ctx context.Context, projectUUID string) error {
	found := false
	newList := []QTopic{}
	for _, topic := range mk.TopicList {
		if topic.ProjectUUID != projectUUID {
			// found item at i, remove it using index
			newList = append(newList, topic)
		} else {
			found = true
		}
	}
	mk.TopicList = newList
	if found {
		return nil
	}
	return errors.New("not found")
}

// RemoveProjectSubs removes all existing subs belonging to a specific project uuid
func (mk *MockStore) RemoveProjectSubs(ctx context.Context, projectUUID string) error {
	found := false
	newList := []QSub{}
	for _, sub := range mk.SubList {
		if sub.ProjectUUID != projectUUID {
			// found item at i, remove it using index
			newList = append(newList, sub)
		} else {
			found = true
		}
	}
	mk.SubList = newList
	if found {
		return nil
	}
	return errors.New("not found")
}

// RemoveProjectDailyMessageCounters removes all existing message counters belonging to a specific project uuid
func (mk *MockStore) RemoveProjectDailyMessageCounters(ctx context.Context, projectUUID string) error {
	found := false
	newList := []QDailyTopicMsgCount{}
	for _, qc := range mk.DailyTopicMsgCount {
		if qc.ProjectUUID != projectUUID {
			// found item at i, remove it using index
			newList = append(newList, qc)
		} else {
			found = true
		}
	}
	mk.DailyTopicMsgCount = newList
	if found {
		return nil
	}
	return errors.New("not found")
}

// RemoveSub removes an existing sub from the store
func (mk *MockStore) RemoveSub(ctx context.Context, projectUUID string, name string) error {
	for i, sub := range mk.SubList {
		if sub.Name == name && sub.ProjectUUID == projectUUID {
			// found item at i, remove it using index
			mk.SubList = append(mk.SubList[:i], mk.SubList[i+1:]...)
			return nil
		}
	}

	return errors.New("not found")
}

// QueryPushSubs Query push Subscription info from store
func (mk *MockStore) QueryPushSubs(ctx context.Context) []QSub {
	return mk.SubList
}

// QuerySubs Query Subscription info from store
func (mk *MockStore) QuerySubs(ctx context.Context, projectUUID, userUUID, name, pageToken string, pageSize int64) ([]QSub, int64, string, error) {

	var qSubs []QSub
	var totalSize int64
	var nextPageToken string
	var err error
	var pg int
	var limit int
	var counter int

	for _, sub := range mk.SubList {
		if sub.ProjectUUID == projectUUID {

			if userUUID != "" {
				if !mk.existsInACL(ctx, "subscriptions", sub.Name, userUUID) {
					continue
				}

			}

			counter++
		}
	}

	switch name == "" {
	case true:

		if pageSize == 0 {
			limit = counter
		} else {
			limit = int(pageSize) + 1
		}

		if pageToken != "" {
			if pg, err = strconv.Atoi(pageToken); err != nil {
				return qSubs, totalSize, nextPageToken, err
			}
		}

		sort.Slice(mk.SubList, func(i, j int) bool {
			id1 := mk.SubList[i].ID.(int)
			id2 := mk.SubList[j].ID.(int)
			return id1 > id2
		})

		for _, sub := range mk.SubList {

			if limit == 0 {
				break
			}

			if pageToken != "" {

				if sub.ID.(int) <= pg && sub.ProjectUUID == projectUUID {

					if userUUID != "" {
						if !mk.existsInACL(ctx, "subscriptions", sub.Name, userUUID) {
							continue
						}
					}

					qSubs = append(qSubs, sub)
					limit--

				}

			} else {

				if sub.ProjectUUID == projectUUID {

					qSubs = append(qSubs, sub)
					limit--

				}
			}

		}

		totalSize = int64(counter)

		if len(qSubs) > 0 && len(qSubs) == int(pageSize)+1 {
			nextPageToken = strconv.Itoa(qSubs[int(pageSize)].ID.(int))
			qSubs = qSubs[:len(qSubs)-1]
		}

	case false:
		for _, sub := range mk.SubList {
			if sub.ProjectUUID == projectUUID && sub.Name == name {

				if userUUID != "" {
					if !mk.existsInACL(ctx, "subscriptions", sub.Name, userUUID) {
						continue
					}
				}
				qSubs = append(qSubs, sub)
				break
			}
		}

	}

	return qSubs, totalSize, nextPageToken, nil

}

// QuerySubsByTopic returns subscriptions attached to a given topic
func (mk *MockStore) QuerySubsByTopic(ctx context.Context, projectUUID, topic string) ([]QSub, error) {
	result := []QSub{}
	for _, item := range mk.SubList {
		if projectUUID == item.ProjectUUID && item.Topic == topic {
			result = append(result, item)
		}
	}
	return result, nil
}

// QuerySubsByACL returns subscriptions that contain a specific user in their ACL
func (mk *MockStore) QuerySubsByACL(ctx context.Context, projectUUID, user string) ([]QSub, error) {

	result := []QSub{}
	for _, item := range mk.SubList {
		if projectUUID == item.ProjectUUID {
			for _, usr := range mk.SubsACL[item.Name].ACL {
				if usr == user {
					result = append(result, item)
				}
			}
		}
	}

	return result, nil
}

// QueryTopicsByACL returns topics that contain a specific user in their ACL
func (mk *MockStore) QueryTopicsByACL(ctx context.Context, projectUUID, user string) ([]QTopic, error) {

	result := []QTopic{}
	for _, item := range mk.TopicList {
		if projectUUID == item.ProjectUUID {
			for _, usr := range mk.TopicsACL[item.Name].ACL {
				if usr == user {
					result = append(result, item)
				}
			}
		}
	}

	return result, nil
}

// QueryTopics Query Subscription info from store
func (mk *MockStore) QueryTopics(ctx context.Context, projectUUID, userUUID, name, pageToken string, pageSize int64) ([]QTopic, int64, string, error) {

	var qTopics []QTopic
	var totalSize int64
	var nextPageToken string
	var err error
	var pg int
	var limit int
	var counter int

	for _, topic := range mk.TopicList {
		if topic.ProjectUUID == projectUUID {

			if userUUID != "" {
				if !mk.existsInACL(ctx, "topics", topic.Name, userUUID) {
					continue
				}

			}
			counter++
		}
	}

	switch name == "" {
	case true:

		if pageSize == 0 {
			limit = counter
		} else {
			limit = int(pageSize) + 1
		}

		if pageToken != "" {
			if pg, err = strconv.Atoi(pageToken); err != nil {
				return qTopics, totalSize, nextPageToken, err
			}
		}

		sort.Slice(mk.TopicList, func(i, j int) bool {
			id1 := mk.TopicList[i].ID.(int)
			id2 := mk.TopicList[j].ID.(int)
			return id1 > id2
		})

		for _, topic := range mk.TopicList {

			if limit == 0 {
				break
			}

			if pageToken != "" {

				if topic.ID.(int) <= pg && topic.ProjectUUID == projectUUID {

					if userUUID != "" {
						if !mk.existsInACL(ctx, "topics", topic.Name, userUUID) {
							continue
						}
					}

					qTopics = append(qTopics, topic)
					limit--

				}

			} else {

				if topic.ProjectUUID == projectUUID {

					if userUUID != "" {
						if !mk.existsInACL(ctx, "topics", topic.Name, userUUID) {
							continue
						}
					}

					qTopics = append(qTopics, topic)
					limit--

				}
			}

		}

		totalSize = int64(counter)

		if len(qTopics) > 0 && len(qTopics) == int(pageSize)+1 {
			nextPageToken = strconv.Itoa(qTopics[int(pageSize)].ID.(int))
			qTopics = qTopics[:len(qTopics)-1]
		}

	case false:

		for _, topic := range mk.TopicList {
			if topic.ProjectUUID == projectUUID && topic.Name == name {

				if userUUID != "" {
					if !mk.existsInACL(ctx, "topics", topic.Name, userUUID) {
						continue
					}
				}

				qTopics = append(qTopics, topic)
				break
			}
		}

	}

	return qTopics, totalSize, nextPageToken, nil
}

func (mk *MockStore) existsInACL(ctx context.Context, resource, resourceName, userUUID string) bool {

	var acl QAcl

	if resource == "subscriptions" {
		acl = mk.SubsACL[resourceName]
	} else if resource == "topics" {
		acl = mk.TopicsACL[resourceName]
	}

	for _, u := range acl.ACL {
		if u == userUUID {
			return true
		}

	}

	return false

}

// Checks if a users exists in an ACL resource (topic or subscription)
func (mk *MockStore) ExistsInACL(ctx context.Context, projectUUID string, resource string, resourceName string, userUUID string) error {

	var acl QAcl

	if resource == "subscriptions" {
		acl = mk.SubsACL[resourceName]
	} else if resource == "topics" {
		acl = mk.TopicsACL[resourceName]
	}

	for _, u := range acl.ACL {
		if u == userUUID {
			return nil
		}

	}

	return errors.New("not found")
}

// UpdateTopicLatestPublish updates the topic's latest publish time
func (mk *MockStore) UpdateTopicLatestPublish(ctx context.Context, projectUUID string, name string, date time.Time) error {
	for idx, topic := range mk.TopicList {
		if topic.ProjectUUID == projectUUID && topic.Name == name {
			mk.TopicList[idx].LatestPublish = date
			return nil
		}
	}
	return errors.New("topic not found")
}

// UpdateTopicPublishRate updates the topic's publishing rate
func (mk *MockStore) UpdateTopicPublishRate(ctx context.Context, projectUUID string, name string, rate float64) error {
	for idx, topic := range mk.TopicList {
		if topic.ProjectUUID == projectUUID && topic.Name == name {
			mk.TopicList[idx].PublishRate = rate
			return nil
		}
	}
	return errors.New("topic not found")
}

// UpdateSubLatestConsume updates the subscription's latest consume time
func (mk *MockStore) UpdateSubLatestConsume(ctx context.Context, projectUUID string, name string, date time.Time) error {
	for idx, topic := range mk.SubList {
		if topic.ProjectUUID == projectUUID && topic.Name == name {
			mk.SubList[idx].LatestConsume = date
			return nil
		}
	}
	return errors.New("subscription not found")
}

// UpdateSubConsumeRate updates the subscription's consume rate
func (mk *MockStore) UpdateSubConsumeRate(ctx context.Context, projectUUID string, name string, rate float64) error {
	for idx, topic := range mk.SubList {
		if topic.ProjectUUID == projectUUID && topic.Name == name {
			mk.SubList[idx].ConsumeRate = rate
			return nil
		}
	}
	return errors.New("subscription not found")
}

// QueryDailyTopicMsgCount returns results regarding the number of messages published to a topic
func (mk *MockStore) QueryDailyTopicMsgCount(ctx context.Context, projectUUID string, topicName string, date time.Time) ([]QDailyTopicMsgCount, error) {

	var qds []QDailyTopicMsgCount
	var zeroValueTime time.Time

	if projectUUID == "" && topicName == "" && date.Equal(zeroValueTime) {

		qds = mk.DailyTopicMsgCount
	}

	if projectUUID != "" && topicName != "" && date.Equal(zeroValueTime) {
		for _, item := range mk.DailyTopicMsgCount {
			if item.ProjectUUID == projectUUID && item.TopicName == topicName {
				qds = append(qds, item)
			}
		}
	}

	if projectUUID != "" && topicName != "" && !date.Equal(zeroValueTime) {
		for _, item := range mk.DailyTopicMsgCount {
			if item.ProjectUUID == projectUUID && item.TopicName == topicName && item.Date.Equal(date) {
				qds = append(qds, item)
			}
		}
	}

	// sort in descending order
	sort.Slice(qds, func(i, j int) bool { return qds[i].Date.After(qds[j].Date) })

	return qds, nil
}

func (mk *MockStore) InsertSchema(ctx context.Context, projectUUID, schemaUUID, name, schemaType, rawSchemaString string) error {
	mk.SchemaList = append(mk.SchemaList, QSchema{
		ProjectUUID: projectUUID,
		UUID:        schemaUUID,
		Name:        name,
		Type:        schemaType,
		RawSchema:   rawSchemaString,
	})

	return nil
}

func (mk *MockStore) QuerySchemas(ctx context.Context, projectUUID, schemaUUID, name string) ([]QSchema, error) {

	qSchemas := []QSchema{}

	if schemaUUID == "" && name == "" {
		for _, qs := range mk.SchemaList {
			if qs.ProjectUUID == projectUUID {
				qSchemas = append(qSchemas, qs)
			}
		}
	}

	if schemaUUID != "" && name != "" {
		for _, qs := range mk.SchemaList {
			if qs.ProjectUUID == projectUUID && qs.UUID == schemaUUID && qs.Name == name {
				qSchemas = append(qSchemas, qs)
			}
		}
	}

	if schemaUUID != "" && name == "" {
		for _, qs := range mk.SchemaList {
			if qs.ProjectUUID == projectUUID && qs.UUID == schemaUUID {
				qSchemas = append(qSchemas, qs)
			}
		}
	}

	if schemaUUID == "" && name != "" {
		for _, qs := range mk.SchemaList {
			if qs.ProjectUUID == projectUUID && qs.Name == name {
				qSchemas = append(qSchemas, qs)
			}
		}
	}

	return qSchemas, nil
}

func (mk *MockStore) UpdateSchema(ctx context.Context, schemaUUID, name, schemaType, rawSchemaString string) error {

	for idx, s := range mk.SchemaList {
		if s.UUID == schemaUUID {

			if name != "" {
				mk.SchemaList[idx].Name = name
			}

			if schemaType != "" {
				mk.SchemaList[idx].Type = schemaType
			}

			if rawSchemaString != "" {
				mk.SchemaList[idx].RawSchema = rawSchemaString
			}

			return nil
		}
	}

	return errors.New("not found")
}
func (mk *MockStore) DeleteSchema(ctx context.Context, schemaUUID string) error {

	for idx, s := range mk.SchemaList {
		if s.UUID == schemaUUID {
			mk.SchemaList = append(mk.SchemaList[:idx], mk.SchemaList[idx+1:]...)

			for idx, t := range mk.TopicList {
				if t.SchemaUUID == schemaUUID {
					mk.TopicList[idx].SchemaUUID = ""
				}
			}

			return nil
		}
	}

	return errors.New("not found")
}

func (mk *MockStore) DeleteRegistration(ctx context.Context, regUUID string) error {

	for idx, s := range mk.UserRegistrations {
		if s.UUID == regUUID {
			mk.UserRegistrations = append(mk.UserRegistrations[:idx], mk.UserRegistrations[idx+1:]...)
			return nil
		}
	}

	return errors.New("not found")
}
