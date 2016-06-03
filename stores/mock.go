package stores

import (
	"errors"
	"time"
)

// MockStore holds configuration
type MockStore struct {
	Server      string
	Database    string
	SubList     []QSub
	TopicList   []QTopic
	ProjectList []QProject
	UserList    []QUser
	RoleList    []QRole
	Session     bool
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

// Close is used to close session
func (mk *MockStore) Close() {
	mk.Session = false
}

// UpdateSubOffset updates the offset of the current subscription
func (mk *MockStore) UpdateSubOffset(name string, offset int64) {

}

// UpdateSubOffsetAck updates the offset of the current subscription
func (mk *MockStore) UpdateSubOffsetAck(name string, offset int64, ts string) error {
	// find sub
	sub := QSub{}
	for _, item := range mk.QuerySubs() {
		if item.Name == name {
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

// UpdateSubPull updates next offset info after a pull
func (mk *MockStore) UpdateSubPull(name string, offset int64, ts string) {

}

// Initialize is used to initalize the mock
func (mk *MockStore) Initialize() {

	// populate topics
	qtop1 := QTopic{"ARGO", "topic1"}
	qtop2 := QTopic{"ARGO", "topic2"}
	qtop3 := QTopic{"ARGO", "topic3"}
	mk.TopicList = append(mk.TopicList, qtop1)
	mk.TopicList = append(mk.TopicList, qtop2)
	mk.TopicList = append(mk.TopicList, qtop3)

	// populate Subscriptions
	qsub1 := QSub{"ARGO", "sub1", "topic1", 0, 0, "", 10}
	qsub2 := QSub{"ARGO", "sub2", "topic2", 0, 0, "", 10}
	qsub3 := QSub{"ARGO", "sub3", "topic3", 0, 0, "", 10}
	mk.SubList = append(mk.SubList, qsub1)
	mk.SubList = append(mk.SubList, qsub2)
	mk.SubList = append(mk.SubList, qsub3)

	// populate Projects
	qPr := QProject{"ARGO"}
	mk.ProjectList = append(mk.ProjectList, qPr)

	// populate Users
	qUsr := QUser{"Test", "Test@test.com", "ARGO", "S3CR3T", []string{"admin", "member"}}
	mk.UserList = append(mk.UserList, qUsr)

	qRole1 := QRole{"topics:list_all", []string{"admin", "reader", "publisher"}}
	qRole2 := QRole{"topics:publish", []string{"admin", "publisher"}}
	mk.RoleList = append(mk.RoleList, qRole1)
	mk.RoleList = append(mk.RoleList, qRole2)

}

// Clone the store
func (mk *MockStore) Clone() Store {
	return mk
}

// GetUserRoles returns the roles of a user in a project
func (mk *MockStore) GetUserRoles(project string, token string) []string {
	for _, item := range mk.UserList {
		if item.Project == project && item.Token == token {
			return item.Roles
		}
	}

	return []string{}
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
func (mk *MockStore) HasProject(project string) bool {
	for _, item := range mk.ProjectList {
		if item.Name == project {
			return true
		}
	}

	return false
}

// InsertTopic inserts a new topic object to the store
func (mk *MockStore) InsertTopic(project string, name string) error {
	topic := QTopic{project, name}
	mk.TopicList = append(mk.TopicList, topic)
	return nil
}

// InsertSub inserts a new sub object to the store
func (mk *MockStore) InsertSub(project string, name string, topic string, offset int64, ack int) error {
	sub := QSub{project, name, topic, offset, 0, "", ack}
	mk.SubList = append(mk.SubList, sub)
	return nil
}

// RemoveTopic removes an existing topic
func (mk *MockStore) RemoveTopic(project string, name string) error {
	for i, topic := range mk.TopicList {
		if topic.Name == name && topic.Project == project {
			// found item at i, remove it using index
			mk.TopicList = append(mk.TopicList[:i], mk.TopicList[i+1:]...)
			return nil
		}
	}

	return errors.New("not found")
}

// RemoveSub removes an existing sub from the store
func (mk *MockStore) RemoveSub(project string, name string) error {
	for i, sub := range mk.SubList {
		if sub.Name == name && sub.Project == project {
			// found item at i, remove it using index
			mk.SubList = append(mk.SubList[:i], mk.SubList[i+1:]...)
			return nil
		}
	}

	return errors.New("not found")
}

// QuerySubs Query Subscription info from store
func (mk *MockStore) QuerySubs() []QSub {
	return mk.SubList
}

// QueryTopics Query Subscription info from store
func (mk *MockStore) QueryTopics() []QTopic {
	return mk.TopicList
}
