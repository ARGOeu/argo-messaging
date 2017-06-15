package stores

import (
	"errors"
	"time"

	log "github.com/Sirupsen/logrus"

	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

// MongoStore holds configuration
type MongoStore struct {
	Server   string
	Database string
	Session  *mgo.Session
}

// NewMongoStore creates new mongo store
func NewMongoStore(server string, db string) *MongoStore {
	mong := MongoStore{}
	mong.Server = server
	mong.Database = db
	return &mong
}

// Close is used to close session
func (mong *MongoStore) Close() {
	mong.Session.Close()

}

// Clone the store with  a cloned session
func (mong *MongoStore) Clone() Store {
	nStore := NewMongoStore(mong.Server, mong.Database)
	nStore.Session = mong.Session.Clone()
	return nStore
}

// Initialize initializes the mongo store struct
func (mong *MongoStore) Initialize() {

	session, err := mgo.Dial(mong.Server)
	if err != nil {

		log.Fatal("STORE", "\t", err.Error())

	}

	mong.Session = session

	log.Info("STORE", "\t", "Connected to Mongo: ", mong.Server)
}

// QueryProjects queries the database for a specific project or a list of all projects
func (mong *MongoStore) QueryProjects(uuid string, name string) ([]QProject, error) {

	query := bson.M{}

	if name != "" {

		query = bson.M{"name": name}

	} else if uuid != "" {
		query = bson.M{"uuid": uuid}
	}

	db := mong.Session.DB(mong.Database)
	c := db.C("projects")
	var results []QProject
	err := c.Find(query).All(&results)
	if err != nil {
		log.Fatal("STORE", "\t", err.Error())
	}

	if len(results) > 0 {
		return results, nil
	}

	return results, errors.New("not found")
}

// UpdateProject updates project information
func (mong *MongoStore) UpdateProject(projectUUID string, name string, description string, modifiedOn time.Time) error {
	db := mong.Session.DB(mong.Database)
	c := db.C("projects")

	doc := bson.M{"uuid": projectUUID}
	results, err := mong.QueryProjects(projectUUID, "")
	if err != nil {
		return err
	}

	curPr := results[0]
	curPr.ModifiedOn = modifiedOn // modifiedOn should always be updated

	if name != "" {
		// Check if name is going to change and if that name already exists
		if name != curPr.Name {
			if sameRes, _ := mong.QueryProjects("", name); len(sameRes) > 0 {
				return errors.New("invalid project name change, name already exists")
			}
		}
		curPr.Name = name
	}

	if description != "" {
		curPr.Description = description
	}

	change := bson.M{"$set": curPr}

	err = c.Update(doc, change)

	return err

}

// UpdateUserToken updates user's token
func (mong *MongoStore) UpdateUserToken(uuid string, token string) error {

	db := mong.Session.DB(mong.Database)
	c := db.C("users")

	doc := bson.M{"uuid": uuid}
	change := bson.M{"$set": bson.M{"token": token}}

	err := c.Update(doc, change)

	return err

}

// UpdateUser updates user information
func (mong *MongoStore) UpdateUser(uuid string, projects []QProjectRoles, name string, email string, serviceRoles []string, modifiedOn time.Time) error {
	db := mong.Session.DB(mong.Database)
	c := db.C("users")

	doc := bson.M{"uuid": uuid}
	results, err := mong.QueryUsers("", uuid, "")
	if err != nil {
		return err
	}

	curUsr := results[0]

	if name != "" {
		// Check if name is going to change and if that name already exists
		if name != curUsr.Name {
			if sameRes, _ := mong.QueryUsers("", "", name); len(sameRes) > 0 {
				return errors.New("invalid user name change, name already exists")
			}
		}
		curUsr.Name = name
	}

	if email != "" {
		curUsr.Email = email
	}

	if projects != nil {
		curUsr.Projects = projects
	}

	if serviceRoles != nil {
		curUsr.ServiceRoles = serviceRoles
	}

	curUsr.ModifiedOn = modifiedOn

	change := bson.M{"$set": curUsr}

	err = c.Update(doc, change)

	return err

}

// UpdateSubPull updates next offset and sets timestamp for Ack
func (mong *MongoStore) UpdateSubPull(projectUUID string, name string, nextOff int64, ts string) error {

	db := mong.Session.DB(mong.Database)
	c := db.C("subscriptions")

	doc := bson.M{"project_uuid": projectUUID, "name": name}
	change := bson.M{"$set": bson.M{"next_offset": nextOff, "pending_ack": ts}}
	err := c.Update(doc, change)
	if err != nil && err != mgo.ErrNotFound {
		log.Fatal("STORE", "\t", err.Error())
	}

	return err

}

// UpdateSubOffsetAck updates a subscription offset after Ack
func (mong *MongoStore) UpdateSubOffsetAck(projectUUID string, name string, offset int64, ts string) error {

	db := mong.Session.DB(mong.Database)
	c := db.C("subscriptions")

	// Get Info
	res := QSub{}
	err := c.Find(bson.M{"project_uuid": projectUUID, "name": name}).One(&res)

	// check if no ack pending
	if res.NextOffset == 0 {
		return errors.New("no ack pending")
	}

	// check if ack offset is wrong - wrong ack
	if offset <= res.Offset || offset > res.NextOffset {
		return errors.New("wrong ack")
	}

	// check if ack has timeout
	zSec := "2006-01-02T15:04:05Z"
	timeGiven, _ := time.Parse(zSec, ts)
	timeRef, _ := time.Parse(zSec, res.PendingAck)
	durSec := timeGiven.Sub(timeRef).Seconds()

	if int(durSec) > res.Ack {
		return errors.New("ack timeout")
	}

	doc := bson.M{"project_uuid": projectUUID, "name": name}
	change := bson.M{"$set": bson.M{"offset": offset, "next_offset": 0, "pending_ack": ""}}
	err = c.Update(doc, change)
	if err != nil && err != mgo.ErrNotFound {
		log.Fatal("STORE", "\t", err.Error())
	}

	return nil

}

// UpdateSubOffset updates a subscription offset
func (mong *MongoStore) UpdateSubOffset(projectUUID string, name string, offset int64) {

	db := mong.Session.DB(mong.Database)
	c := db.C("subscriptions")

	doc := bson.M{"project_uuid": projectUUID, "name": name}
	change := bson.M{"$set": bson.M{"offset": offset, "next_offset": 0, "pending_ack": ""}}
	err := c.Update(doc, change)
	if err != nil && err != mgo.ErrNotFound {
		log.Fatal("STORE", "\t", err.Error())
	}

}

// HasUsers accepts a user array of usernames and returns the not found
func (mong *MongoStore) HasUsers(projectUUID string, users []string) (bool, []string) {
	db := mong.Session.DB(mong.Database)
	var results []QUser
	var notFound []string
	c := db.C("users")

	err := c.Find(bson.M{"projects": bson.M{"$elemMatch": bson.M{"project_uuid": projectUUID}}, "name": bson.M{"$in": users}}).All(&results)
	if err != nil {
		log.Fatal("STORE", "\t", err.Error())
	}

	// for each given username
	for _, username := range users {
		found := false
		// loop through all found users
		for _, user := range results {
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

// QueryACL queries topic or subscription for a list of authorized users
func (mong *MongoStore) QueryACL(projectUUID string, resource string, name string) (QAcl, error) {

	db := mong.Session.DB(mong.Database)
	var results []QAcl
	var c *mgo.Collection
	if resource != "topics" && resource != "subscriptions" {
		return QAcl{}, errors.New("wrong resource type")
	}

	c = db.C(resource)

	err := c.Find(bson.M{"project_uuid": projectUUID, "name": name}).All(&results)

	if err != nil {
		log.Fatal("STORE", "\t", err.Error())
	}

	if len(results) > 0 {
		return results[0], nil
	}

	return QAcl{}, errors.New("not found")
}

// QueryUsers queries user(s) information belonging to a project
func (mong *MongoStore) QueryUsers(projectUUID string, uuid string, name string) ([]QUser, error) {

	// By default return all users
	query := bson.M{}
	// If project UUID is given return users that belong to the project
	if projectUUID != "" {
		query = bson.M{"project_uuid": projectUUID}
		if uuid != "" {
			query = bson.M{"project_uuid": projectUUID, "uuid": uuid}
		} else if name != "" {
			query = bson.M{"project_uuid": projectUUID, "name": name}
		}
	} else {
		if uuid != "" {
			query = bson.M{"uuid": uuid}
		} else if name != "" {
			query = bson.M{"name": name}

		}
	}

	db := mong.Session.DB(mong.Database)
	c := db.C("users")
	var results []QUser
	err := c.Find(query).All(&results)

	if err != nil {
		log.Fatal("STORE", "\t", err.Error())
	}

	return results, err
}

// QueryTopics Query Subscription info from store
func (mong *MongoStore) QueryTopics(projectUUID string, name string) ([]QTopic, error) {

	// By default return all topics of a given project
	query := bson.M{"project_uuid": projectUUID}

	// If name is given return only the specific topic
	if name != "" {
		query = bson.M{"project_uuid": projectUUID, "name": name}
	}
	db := mong.Session.DB(mong.Database)
	c := db.C("topics")
	var results []QTopic
	err := c.Find(query).All(&results)

	if err != nil {
		log.Fatal("STORE", "\t", err.Error())
	}

	return results, err
}

// IncrementTopicMsgNum increments the number of messages published in a topic
func (mong *MongoStore) IncrementTopicMsgNum(projectUUID string, name string, num int64) error {

	db := mong.Session.DB(mong.Database)
	c := db.C("topics")

	doc := bson.M{"project_uuid": projectUUID, "name": name}
	change := bson.M{"$inc": bson.M{"msg_num": num}}

	err := c.Update(doc, change)

	return err

}

//IncrementTopicBytes increases the total number of bytes published in a topic
func (mong *MongoStore) IncrementTopicBytes(projectUUID string, name string, totalBytes int64) error {
	db := mong.Session.DB(mong.Database)
	c := db.C("topics")

	doc := bson.M{"project_uuid": projectUUID, "name": name}
	change := bson.M{"$inc": bson.M{"total_bytes": totalBytes}}

	err := c.Update(doc, change)

	return err
}

// IncrementTopicMsgNum increments the number of messages pulled in a subscription
func (mong *MongoStore) IncrementSubMsgNum(projectUUID string, name string, num int64) error {

	db := mong.Session.DB(mong.Database)
	c := db.C("subscriptions")

	doc := bson.M{"project_uuid": projectUUID, "name": name}
	change := bson.M{"$inc": bson.M{"msg_num": num}}

	err := c.Update(doc, change)

	return err

}

//IncrementSubMsgNum increases the total number of bytes consumed from a subscripion
func (mong *MongoStore) IncrementSubBytes(projectUUID string, name string, totalBytes int64) error {
	db := mong.Session.DB(mong.Database)
	c := db.C("subscriptions")

	doc := bson.M{"project_uuid": projectUUID, "name": name}
	change := bson.M{"$inc": bson.M{"total_bytes": totalBytes}}

	err := c.Update(doc, change)

	return err
}

//HasResourceRoles returns the roles of a user in a project
func (mong *MongoStore) HasResourceRoles(resource string, roles []string) bool {

	db := mong.Session.DB(mong.Database)
	c := db.C("roles")
	var results []QRole
	err := c.Find(bson.M{"resource": resource, "roles": bson.M{"$in": roles}}).All(&results)
	if err != nil {
		log.Fatal("STORE", "\t", err.Error())
	}

	if len(results) > 0 {
		return true
	}

	return false

}

//GetAllRoles returns a list of all available roles
func (mong *MongoStore) GetAllRoles() []string {

	db := mong.Session.DB(mong.Database)
	c := db.C("roles")
	var results []string
	err := c.Find(nil).Distinct("roles", &results)
	if err != nil {
		log.Fatal("STORE", "\t", err.Error())
	}
	return results
}

//GetUserRoles returns the roles of a user in a project
func (mong *MongoStore) GetUserRoles(projectUUID string, token string) ([]string, string) {

	db := mong.Session.DB(mong.Database)
	c := db.C("users")
	var results []QUser

	err := c.Find(bson.M{"token": token}).All(&results)

	if err != nil {
		log.Fatal("STORE", "\t", err.Error())
	}

	if len(results) == 0 {
		return []string{}, ""
	}

	if len(results) > 1 {
		log.Warning("STORE", "\t", "Multiple users with the same token", token)

	}

	// Search the found user for project roles
	return results[0].getProjectRoles(projectUUID), results[0].Name

}

// QueryOneSub queries and returns specific sub of project
func (mong *MongoStore) QueryOneSub(projectUUID string, name string) (QSub, error) {

	db := mong.Session.DB(mong.Database)
	c := db.C("subscriptions")
	var results []QSub
	err := c.Find(bson.M{"project_uuid": projectUUID, "name": name}).All(&results)
	if err != nil {
		log.Fatal("STORE", "\t", err.Error())
	}

	if len(results) > 0 {
		return results[0], nil
	}

	return QSub{}, errors.New("empty")
}

// HasProject Returns true if project exists
func (mong *MongoStore) HasProject(name string) bool {

	db := mong.Session.DB(mong.Database)
	c := db.C("projects")
	var results []QProject
	err := c.Find(bson.M{"name": name}).All(&results)
	if err != nil {
		log.Fatal("STORE", "\t", err.Error())
	}

	if len(results) > 0 {
		return true
	}
	return false
}

// InsertTopic inserts a topic to the store
func (mong *MongoStore) InsertTopic(projectUUID string, name string) error {
	topic := QTopic{ProjectUUID: projectUUID, Name: name, MsgNum: 0, TotalBytes: 0}
	return mong.InsertResource("topics", topic)
}

// InsertUser inserts a new user to the store
func (mong *MongoStore) InsertUser(uuid string, projects []QProjectRoles, name string, token string, email string, serviceRoles []string, createdOn time.Time, modifiedOn time.Time, createdBy string) error {
	user := QUser{UUID: uuid, Name: name, Email: email, Token: token, Projects: projects, ServiceRoles: serviceRoles, CreatedOn: createdOn, ModifiedOn: modifiedOn, CreatedBy: createdBy}
	return mong.InsertResource("users", user)
}

// InsertProject inserts a project to the store
func (mong *MongoStore) InsertProject(uuid string, name string, createdOn time.Time, modifiedOn time.Time, createdBy string, description string) error {
	project := QProject{UUID: uuid, Name: name, CreatedOn: createdOn, ModifiedOn: modifiedOn, CreatedBy: createdBy, Description: description}
	return mong.InsertResource("projects", project)
}

// InsertSub inserts a subscription to the store
func (mong *MongoStore) InsertSub(projectUUID string, name string, topic string, offset int64, ack int, push string, rPolicy string, rPeriod int) error {
	sub := QSub{projectUUID, name, topic, offset, 0, "", push, ack, rPolicy, rPeriod, 0, 0}
	return mong.InsertResource("subscriptions", sub)
}

// RemoveProjectTopics removes all topics related to a project UUID
func (mong *MongoStore) RemoveProjectTopics(projectUUID string) error {
	topicMatch := bson.M{"project_uuid": projectUUID}
	return mong.RemoveAll("topics", topicMatch)
}

// RemoveProjectSubs removes all subscriptions related to a project UUID
func (mong *MongoStore) RemoveProjectSubs(projectUUID string) error {
	subMatch := bson.M{"project_uuid": projectUUID}
	return mong.RemoveAll("subscriptions", subMatch)
}

// RemoveProject removes a project from the store
func (mong *MongoStore) RemoveProject(uuid string) error {
	project := bson.M{"uuid": uuid}
	return mong.RemoveResource("projects", project)
}

// RemoveTopic removes a topic from the store
func (mong *MongoStore) RemoveTopic(projectUUID string, name string) error {
	topic := bson.M{"project_uuid": projectUUID, "name": name}
	return mong.RemoveResource("topics", topic)
}

// RemoveUser removes a user entry from the store
func (mong *MongoStore) RemoveUser(uuid string) error {
	user := bson.M{"uuid": uuid}
	return mong.RemoveResource("users", user)
}

// RemoveSub removes a subscription from the store
func (mong *MongoStore) RemoveSub(projectUUID string, name string) error {
	sub := bson.M{"project_uuid": projectUUID, "name": name}
	return mong.RemoveResource("subscriptions", sub)
}

// ModACL modifies the push configuration
func (mong *MongoStore) ModACL(projectUUID string, resource string, name string, acl []string) error {
	db := mong.Session.DB(mong.Database)

	if resource != "topics" && resource != "subscriptions" {
		return errors.New("wrong resource type")
	}

	c := db.C(resource)

	err := c.Update(bson.M{"project_uuid": projectUUID, "name": name}, bson.M{"$set": bson.M{"acl": acl}})
	return err
}

// ModSubPush modifies the push configuration
func (mong *MongoStore) ModSubPush(projectUUID string, name string, push string, rPolicy string, rPeriod int) error {
	db := mong.Session.DB(mong.Database)
	c := db.C("subscriptions")

	err := c.Update(bson.M{"project_uuid": projectUUID, "name": name}, bson.M{"$set": bson.M{"push_endpoint": push, "retry_policy": rPolicy, "retry_period": rPeriod}})
	return err
}

// InsertResource inserts a new topic object to the datastore
func (mong *MongoStore) InsertResource(col string, res interface{}) error {

	db := mong.Session.DB(mong.Database)
	c := db.C(col)

	err := c.Insert(res)
	if err != nil {
		log.Fatal("STORE", "\t", err.Error())
	}

	return err
}

// RemoveAll removes all  occurences matched with a resource from the store
func (mong *MongoStore) RemoveAll(col string, res interface{}) error {
	db := mong.Session.DB(mong.Database)
	c := db.C(col)

	_, err := c.RemoveAll(res)

	return err
}

// RemoveResource removes a resource from the store
func (mong *MongoStore) RemoveResource(col string, res interface{}) error {

	db := mong.Session.DB(mong.Database)
	c := db.C(col)

	err := c.Remove(res) // if not found mgo returns error string "not found"

	return err
}

// QuerySubs Query Subscription info from store
func (mong *MongoStore) QuerySubs(projectUUID string, name string) ([]QSub, error) {

	query := bson.M{"project_uuid": projectUUID}
	// If name is given return only the specific topic
	if name != "" {
		query = bson.M{"project_uuid": projectUUID, "name": name}

	}

	db := mong.Session.DB(mong.Database)
	c := db.C("subscriptions")
	var results []QSub
	err := c.Find(query).All(&results)
	if err != nil {
		log.Fatal("STORE", "\t", err.Error())
	}
	return results, err

}

// QueryPushSubs retrieves subscriptions that have a push_endpoint defined
func (mong *MongoStore) QueryPushSubs() []QSub {

	db := mong.Session.DB(mong.Database)
	c := db.C("subscriptions")
	var results []QSub
	err := c.Find(bson.M{"push_endpoint": bson.M{"$ne": nil}}).All(&results)
	if err != nil {
		log.Fatal("STORE", "\t", err.Error())
	}
	return results

}
