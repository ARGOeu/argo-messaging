package stores

import (
	"errors"
	"sort"
	"strconv"
	"time"
)

// MockStore holds configuration
type MockStore struct {
	Server             string
	Database           string
	SubList            []QSub
	TopicList          []QTopic
	DailyTopicMsgCount []QDailyTopicMsgCount
	ProjectList        []QProject
	UserList           []QUser
	RoleList           []QRole
	Session            bool
	TopicsACL          map[string]QAcl
	SubsACL            map[string]QAcl
	OpMetrics          map[string]QopMetric
}

// QueryACL Topic/Subscription ACL
func (mk *MockStore) QueryACL(projectUUID string, resource string, name string) (QAcl, error) {
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

func (mk *MockStore) InsertOpMetric(hostname string, cpu float64, mem float64) error {
	qOp := QopMetric{hostname, cpu, mem}
	mk.OpMetrics[hostname] = qOp
	return nil
}

// Close is used to close session
func (mk *MockStore) Close() {
	mk.Session = false
}

// InsertUser inserts a new user to the store
func (mk *MockStore) InsertUser(uuid string, projects []QProjectRoles, name string, token string, email string, serviceRoles []string, createdOn time.Time, modifiedOn time.Time, createdBy string) error {
	user := QUser{UUID: uuid, Name: name, Email: email, Projects: projects, Token: token, ServiceRoles: serviceRoles, CreatedOn: createdOn, ModifiedOn: modifiedOn, CreatedBy: createdBy}
	mk.UserList = append(mk.UserList, user)
	return nil
}

//GetAllRoles returns a list of all available roles
func (mk *MockStore) GetAllRoles() []string {
	return []string{"service_admin", "admin", "project_admin", "viewer", "consumer", "producer", "publisher", "push_worker"}
}

// UpdateUserToken updates user's token
func (mk *MockStore) UpdateUserToken(uuid string, token string) error {
	for i, item := range mk.UserList {
		if item.UUID == uuid {
			mk.UserList[i].Token = token
			return nil
		}
	}

	return errors.New("not found")

}

func (mk *MockStore) GetOpMetrics() []QopMetric {
	results := []QopMetric{}
	for _, v := range mk.OpMetrics {
		results = append(results, v)
	}
	return results
}

func (mk *MockStore) AppendToUserProjects(userUUID string, projectUUID string, pRoles ...string) error {

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
func (mk *MockStore) UpdateUser(uuid string, projects []QProjectRoles, name string, email string, serviceRoles []string, modifiedOn time.Time) error {

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

			mk.UserList[i].ModifiedOn = modifiedOn

			return nil
		}
	}

	return errors.New("not found")

}

// HasUsers accepts a user array of usernames and returns the not found
func (mk *MockStore) HasUsers(projectUUID string, users []string) (bool, []string) {

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
func (mk *MockStore) ModACL(projectUUID string, resource string, name string, acl []string) error {
	newAcl := QAcl{ACL: acl}
	if resource == "topics" {
		if _, exists := mk.TopicsACL[name]; exists {
			mk.TopicsACL[name] = newAcl
			return nil
		}
	} else if resource == "subscriptions" {
		if _, exists := mk.SubsACL[name]; exists {
			mk.SubsACL[name] = newAcl
			return nil
		}
	}

	return errors.New("wrong resource type")
}

func (mk *MockStore) AppendToACL(projectUUID string, resource string, name string, acl []string) error {
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

func (mk *MockStore) RemoveFromACL(projectUUID string, resource string, name string, acl []string) error {
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
func (mk *MockStore) UpdateProject(projectUUID string, name string, description string, modifiedOn time.Time) error {

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
func (mk *MockStore) QueryDailyProjectMsgCount(projectUUID string) ([]QDailyProjectMsgCount, error) {

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

//IncrementTopicMsgNum increase number of messages published in a topic
func (mk *MockStore) IncrementTopicMsgNum(projectUUID string, name string, num int64) error {

	for i, item := range mk.TopicList {
		if item.ProjectUUID == projectUUID && item.Name == name {
			mk.TopicList[i].MsgNum += num
			return nil
		}
	}

	return errors.New("not found")
}

//IncrementTopicMsgNum increase number of messages published in a topic
func (mk *MockStore) IncrementDailyTopicMsgCount(projectUUID string, topicName string, num int64, date time.Time) error {

	for i, item := range mk.DailyTopicMsgCount {
		if item.ProjectUUID == projectUUID && item.TopicName == topicName && item.Date.Equal(date) {
			mk.DailyTopicMsgCount[i].NumberOfMessages += num
			return nil
		}
	}

	mk.DailyTopicMsgCount = append(mk.DailyTopicMsgCount, QDailyTopicMsgCount{Date: date, ProjectUUID: projectUUID, TopicName: topicName, NumberOfMessages: num})
	return nil
}

//IncrementTopicBytes increases the total number of bytes published in a topic
func (mk *MockStore) IncrementTopicBytes(projectUUID string, name string, totalBytes int64) error {
	for i, item := range mk.TopicList {
		if item.ProjectUUID == projectUUID && item.Name == name {
			mk.TopicList[i].TotalBytes += totalBytes
			return nil
		}
	}

	return errors.New("not found")
}

//IncrementSubBytes increases the total number of bytes published in a subscription
func (mk *MockStore) IncrementSubBytes(projectUUID string, name string, totalBytes int64) error {
	for i, item := range mk.SubList {
		if item.ProjectUUID == projectUUID && item.Name == name {
			mk.SubList[i].TotalBytes += totalBytes
			return nil
		}
	}

	return errors.New("not found")
}

//IncrementSubMsgNum increase number of messages pulled in a subscription
func (mk *MockStore) IncrementSubMsgNum(projectUUID string, name string, num int64) error {

	for i, item := range mk.SubList {
		if item.ProjectUUID == projectUUID && item.Name == name {
			mk.SubList[i].MsgNum += num
			return nil
		}
	}

	return errors.New("not found")
}

// UpdateSubOffset updates the offset of the current subscription
func (mk *MockStore) UpdateSubOffset(projectUUID string, name string, offset int64) {

}

// ModSubPush modifies the subscription ack
func (mk *MockStore) ModAck(projectUUID string, name string, ack int) error {
	for i, item := range mk.SubList {
		if item.ProjectUUID == projectUUID && item.Name == name {
			mk.SubList[i].Ack = ack

			return nil
		}
	}

	return errors.New("not found")
}

// ModSubPush modifies the subscription push configuration
func (mk *MockStore) ModSubPush(projectUUID string, name string, push string, rPolicy string, rPeriod int) error {
	for i, item := range mk.SubList {
		if item.ProjectUUID == projectUUID && item.Name == name {
			mk.SubList[i].PushEndpoint = push
			mk.SubList[i].RetPolicy = rPolicy
			mk.SubList[i].RetPeriod = rPeriod
			return nil
		}
	}
	return errors.New("not found")
}

// ModSubPush modifies the subscription push configuration
func (mk *MockStore) ModSubPushStatus(projectUUID string, name string, status string) error {
	for i, item := range mk.SubList {
		if item.ProjectUUID == projectUUID && item.Name == name {
			mk.SubList[i].PushStatus = status
			return nil
		}
	}
	return errors.New("not found")
}

// UpdateSubOffsetAck updates the offset of the current subscription
func (mk *MockStore) UpdateSubOffsetAck(projectUUID string, name string, offset int64, ts string) error {
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
func (mk *MockStore) QueryProjects(uuid string, name string) ([]QProject, error) {

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
func (mk *MockStore) QueryUsers(projectUUID string, uuid string, name string) ([]QUser, error) {
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

func (mk *MockStore) PaginatedQueryUsers(pageToken string, pageSize int32) ([]QUser, int32, string, error) {

	var qUsers []QUser
	var totalSize int32
	var nextPageToken string
	var err error
	var pg int
	var limit int

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

	totalSize = int32(len(mk.UserList))

	if len(qUsers) > 0 && len(qUsers) == int(pageSize)+1 {
		nextPageToken = strconv.Itoa(qUsers[int(pageSize)].ID.(int))
		qUsers = qUsers[:len(qUsers)-1]
	}

	return qUsers, totalSize, nextPageToken, err

}

// UpdateSubPull updates next offset info after a pull
func (mk *MockStore) UpdateSubPull(projectUUID string, name string, offset int64, ts string) error {
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
	qtop1 := QTopic{0, "argo_uuid", "topic1", 0, 0}
	qtop2 := QTopic{1, "argo_uuid", "topic2", 0, 0}
	qtop3 := QTopic{2, "argo_uuid", "topic3", 0, 0}
	mk.TopicList = append(mk.TopicList, qtop1)
	mk.TopicList = append(mk.TopicList, qtop2)
	mk.TopicList = append(mk.TopicList, qtop3)

	// populate Subscriptions
	qsub1 := QSub{0, "argo_uuid", "sub1", "topic1", 0, 0, "", "", 10, "", 0, 0, 0, ""}
	qsub2 := QSub{1, "argo_uuid", "sub2", "topic2", 0, 0, "", "", 10, "", 0, 0, 0, ""}
	qsub3 := QSub{2, "argo_uuid", "sub3", "topic3", 0, 0, "", "", 10, "", 0, 0, 0, ""}
	qsub4 := QSub{3, "argo_uuid", "sub4", "topic4", 0, 0, "", "endpoint.foo", 10, "linear", 300, 0, 0, "push enabled"}
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

	// populate daily msg count for topics
	dc1 := QDailyTopicMsgCount{time.Date(2018, 10, 1, 0, 0, 0, 0, time.UTC), "argo_uuid", "topic1", 40}
	dc2 := QDailyTopicMsgCount{time.Date(2018, 10, 2, 0, 0, 0, 0, time.UTC), "argo_uuid", "topic1", 30}
	dc3 := QDailyTopicMsgCount{time.Date(2018, 10, 1, 0, 0, 0, 0, time.UTC), "argo_uuid", "topic2", 70}
	dc4 := QDailyTopicMsgCount{time.Date(2018, 10, 1, 0, 0, 0, 0, time.UTC), "argo_uuid", "topic3", 0}
	mk.DailyTopicMsgCount = append(mk.DailyTopicMsgCount, dc1, dc2, dc3, dc4)

	// populate Users
	qRole := []QProjectRoles{QProjectRoles{"argo_uuid", []string{"consumer", "publisher"}}}
	qUsr := QUser{0, "uuid0", qRole, "Test", "S3CR3T", "Test@test.com", []string{}, created, modified, ""}

	mk.UserList = append(mk.UserList, qUsr)

	qRoleConsumerPub := []QProjectRoles{QProjectRoles{"argo_uuid", []string{"publisher", "consumer"}}}

	mk.UserList = append(mk.UserList, QUser{1, "uuid1", qRole, "UserA", "S3CR3T1", "foo-email", []string{}, created, modified, ""})
	mk.UserList = append(mk.UserList, QUser{2, "uuid2", qRole, "UserB", "S3CR3T2", "foo-email", []string{}, created, modified, "uuid1"})
	mk.UserList = append(mk.UserList, QUser{3, "uuid3", qRoleConsumerPub, "UserX", "S3CR3T3", "foo-email", []string{}, created, modified, "uuid1"})
	mk.UserList = append(mk.UserList, QUser{4, "uuid4", qRoleConsumerPub, "UserZ", "S3CR3T4", "foo-email", []string{}, created, modified, "uuid1"})
	mk.UserList = append(mk.UserList, QUser{5, "same_uuid", qRoleConsumerPub, "UserSame1", "S3CR3T41", "foo-email", []string{}, created, modified, "uuid1"})
	mk.UserList = append(mk.UserList, QUser{6, "same_uuid", qRoleConsumerPub, "UserSame2", "S3CR3T42", "foo-email", []string{}, created, modified, "uuid1"})
	mk.UserList = append(mk.UserList, QUser{7, "uuid7", []QProjectRoles{}, "push_worker_0", "push_token", "foo-email", []string{"push_worker"}, created, modified, ""})

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

}

// QueryOneSub returns one sub exactly
func (mk *MockStore) QueryOneSub(projectUUID string, name string) (QSub, error) {
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

func (mk *MockStore) GetUserFromToken(token string) (QUser, error) {
	for _, item := range mk.UserList {

		if item.Token == token {
			return item, nil

		}
	}

	return QUser{}, errors.New("not found")

}

// GetUserRoles returns the roles of a user in a project
func (mk *MockStore) GetUserRoles(projectUUID string, token string) ([]string, string) {
	for _, item := range mk.UserList {

		if item.Token == token {
			return item.getProjectRoles(projectUUID), item.Name

		}
	}

	return []string{}, ""
}

//HasResourceRoles returns the roles of a user in a project
func (mk *MockStore) HasResourceRoles(resource string, roles []string) bool {

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
func (mk *MockStore) HasProject(name string) bool {
	for _, item := range mk.ProjectList {
		if item.Name == name {
			return true
		}
	}

	return false
}

// InsertTopic inserts a new topic object to the store
func (mk *MockStore) InsertTopic(projectUUID string, name string) error {
	topic := QTopic{ID: len(mk.TopicList), ProjectUUID: projectUUID, Name: name, MsgNum: 0, TotalBytes: 0}
	mk.TopicList = append(mk.TopicList, topic)
	return nil
}

// InsertSub inserts a new sub object to the store
func (mk *MockStore) InsertSub(projectUUID string, name string, topic string, offset int64, ack int, push string, rPolicy string, rPeriod int) error {
	sub := QSub{
		ID:           len(mk.SubList),
		ProjectUUID:  projectUUID,
		Name:         name,
		Topic:        topic,
		Offset:       offset,
		Ack:          ack,
		PushEndpoint: push,
		RetPolicy:    rPolicy,
		RetPeriod:    rPeriod,
		MsgNum:       0,
		TotalBytes:   0,
	}
	mk.SubList = append(mk.SubList, sub)
	mk.SubsACL[name] = QAcl{}
	return nil
}

// InsertProject inserts a project to the store
func (mk *MockStore) InsertProject(uuid string, name string, createdOn time.Time, modifiedOn time.Time, createdBy string, description string) error {
	project := QProject{UUID: uuid, Name: name, CreatedOn: createdOn, ModifiedOn: modifiedOn, CreatedBy: createdBy, Description: description}
	mk.ProjectList = append(mk.ProjectList, project)
	return nil
}

// RemoveProject removes an existing project
func (mk *MockStore) RemoveProject(uuid string) error {
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
func (mk *MockStore) RemoveTopic(projectUUID string, name string) error {
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
func (mk *MockStore) RemoveUser(uuid string) error {
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
func (mk *MockStore) RemoveProjectTopics(projectUUID string) error {
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
func (mk *MockStore) RemoveProjectSubs(projectUUID string) error {
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

// RemoveSub removes an existing sub from the store
func (mk *MockStore) RemoveSub(projectUUID string, name string) error {
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
func (mk *MockStore) QueryPushSubs() []QSub {
	return mk.SubList
}

// QuerySubs Query Subscription info from store
func (mk *MockStore) QuerySubs(projectUUID string, name string, pageToken string, pageSize int32) ([]QSub, int32, string, error) {

	var qSubs []QSub
	var totalSize int32
	var nextPageToken string
	var err error
	var pg int
	var limit int
	var counter int

	for _, sub := range mk.SubList {
		if sub.ProjectUUID == projectUUID {
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

		totalSize = int32(counter)

		if len(qSubs) > 0 && len(qSubs) == int(pageSize)+1 {
			nextPageToken = strconv.Itoa(qSubs[int(pageSize)].ID.(int))
			qSubs = qSubs[:len(qSubs)-1]
		}

	case false:
		for _, sub := range mk.SubList {
			if sub.ProjectUUID == projectUUID && sub.Name == name {
				qSubs = append(qSubs, sub)
				break
			}
		}

	}

	return qSubs, totalSize, nextPageToken, nil

}

func (mk *MockStore) QuerySubsByTopic(projectUUID, topic string) ([]QSub, error) {
	result := []QSub{}
	for _, item := range mk.SubList {
		if projectUUID == item.ProjectUUID && item.Topic == topic {
			result = append(result, item)
		}
	}
	return result, nil
}

func (mk *MockStore) QuerySubsByACL(projectUUID, user string) ([]QSub, error) {

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

func (mk *MockStore) QueryTopicsByACL(projectUUID, user string) ([]QTopic, error) {

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
func (mk *MockStore) QueryTopics(projectUUID string, name string, pageToken string, pageSize int32) ([]QTopic, int32, string, error) {

	var qTopics []QTopic
	var totalSize int32
	var nextPageToken string
	var err error
	var pg int
	var limit int
	var counter int

	for _, topic := range mk.TopicList {
		if topic.ProjectUUID == projectUUID {
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

					qTopics = append(qTopics, topic)
					limit--

				}

			} else {

				if topic.ProjectUUID == projectUUID {

					qTopics = append(qTopics, topic)
					limit--

				}
			}

		}

		totalSize = int32(counter)

		if len(qTopics) > 0 && len(qTopics) == int(pageSize)+1 {
			nextPageToken = strconv.Itoa(qTopics[int(pageSize)].ID.(int))
			qTopics = qTopics[:len(qTopics)-1]
		}

	case false:
		for _, topic := range mk.TopicList {
			if topic.ProjectUUID == projectUUID && topic.Name == name {
				qTopics = append(qTopics, topic)
				break
			}
		}

	}

	return qTopics, totalSize, nextPageToken, nil
}

//IncrementTopicMsgNum increase number of messages published in a topic
func (mk *MockStore) QueryDailyTopicMsgCount(projectUUID string, topicName string, date time.Time) ([]QDailyTopicMsgCount, error) {

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
