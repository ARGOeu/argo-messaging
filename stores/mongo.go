package stores

import (
	"errors"
	"log"
	"time"

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
	// log.Printf("Session Closed...")
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

		log.Fatalf("%s\t%s\t%s", "FATAL", "STORE", err.Error())

	}

	mong.Session = session

	log.Printf("%s\t%s\t%s: %s", "INFO", "STORE", "Connected to Mongo", mong.Server)
}

// UpdateSubPull updates next offset and sets timestamp for Ack
func (mong *MongoStore) UpdateSubPull(name string, nextOff int64, ts string) {

	db := mong.Session.DB(mong.Database)
	c := db.C("subscriptions")

	doc := bson.M{"name": name}
	change := bson.M{"$set": bson.M{"next_offset": nextOff, "pending_ack": ts}}
	err := c.Update(doc, change)
	if err != nil {
		log.Fatalf("%s\t%s\t%s", "FATAL", "STORE", err.Error())
	}

}

// UpdateSubOffsetAck updates a subscription offset after Ack
func (mong *MongoStore) UpdateSubOffsetAck(name string, offset int64, ts string) error {

	db := mong.Session.DB(mong.Database)
	c := db.C("subscriptions")

	// Get Info
	res := QSub{}
	err := c.Find(bson.M{"name": name}).One(&res)

	// check if no ack pending
	if res.NextOffset == 0 {
		return errors.New("no ack pending")
	}

	// check if ack offset is wrong - wrong ack
	if offset < res.Offset || offset > res.NextOffset {
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

	doc := bson.M{"name": name}
	change := bson.M{"$set": bson.M{"offset": offset, "next_offset": 0, "pending_ack": ""}}
	err = c.Update(doc, change)
	if err != nil {
		log.Fatalf("%s\t%s\t%s", "FATAL", "STORE", err.Error())
	}

	return nil

}

// UpdateSubOffset updates a subscription offset
func (mong *MongoStore) UpdateSubOffset(name string, offset int64) {

	db := mong.Session.DB(mong.Database)
	c := db.C("subscriptions")

	doc := bson.M{"name": name}
	change := bson.M{"$set": bson.M{"offset": offset, "next_offset": 0, "pending_ack": ""}}
	err := c.Update(doc, change)
	if err != nil {
		log.Fatalf("%s\t%s\t%s", "FATAL", "STORE", err.Error())
	}

}

// QueryTopics Query Subscription info from store
func (mong *MongoStore) QueryTopics() []QTopic {

	db := mong.Session.DB(mong.Database)
	c := db.C("topics")
	var results []QTopic
	err := c.Find(bson.M{}).All(&results)

	if err != nil {
		log.Fatalf("%s\t%s\t%s", "FATAL", "STORE", err.Error())
	}

	return results
}

//HasResourceRoles returns the roles of a user in a project
func (mong *MongoStore) HasResourceRoles(resource string, roles []string) bool {

	db := mong.Session.DB(mong.Database)
	c := db.C("roles")
	var results []QRole
	err := c.Find(bson.M{"resource": resource, "roles": bson.M{"$in": roles}}).All(&results)
	if err != nil {
		log.Fatalf("%s\t%s\t%s", "FATAL", "STORE", err.Error())
	}

	if len(results) > 0 {
		return true
	}

	return false

}

//GetUserRoles returns the roles of a user in a project
func (mong *MongoStore) GetUserRoles(project string, token string) []string {

	db := mong.Session.DB(mong.Database)
	c := db.C("users")
	var results []QUser
	err := c.Find(bson.M{"project": project, "token": token}).All(&results)
	if err != nil {
		log.Fatalf("%s\t%s\t%s", "FATAL", "STORE", err.Error())
	}

	if len(results) == 0 {
		return []string{}
	}

	if len(results) > 1 {
		log.Printf("%s\t%s\t%s: %s", "WARNING", "STORE", "Multiple users with the same token", token)

	}

	return results[0].Roles

}

// QueryOneSub queries and returns specific sub of project
func (mong *MongoStore) QueryOneSub(project string, sub string) (QSub, error) {

	db := mong.Session.DB(mong.Database)
	c := db.C("subscriptions")
	var results []QSub
	err := c.Find(bson.M{"project": project, "name": sub}).All(&results)
	if err != nil {
		log.Fatalf("%s\t%s\t%s", "FATAL", "STORE", err.Error())
	}

	if len(results) > 0 {
		return results[0], nil
	}

	return QSub{}, errors.New("empty")
}

// HasProject Returns true if project exists
func (mong *MongoStore) HasProject(project string) bool {

	db := mong.Session.DB(mong.Database)
	c := db.C("projects")
	var results []QProject
	err := c.Find(bson.M{}).All(&results)
	if err != nil {
		log.Fatalf("%s\t%s\t%s", "FATAL", "STORE", err.Error())
	}

	if len(results) > 0 {
		return true
	}
	return false
}

// InsertTopic inserts a topic to the store
func (mong *MongoStore) InsertTopic(project string, name string) error {
	topic := QTopic{project, name}
	return mong.InsertResource("topics", topic)
}

// InsertSub inserts a subscription to the store
func (mong *MongoStore) InsertSub(project string, name string, topic string, offset int64, ack int, push string) error {
	sub := QSub{project, name, topic, offset, 0, "", push, ack}
	return mong.InsertResource("subscriptions", sub)
}

// RemoveTopic removes a topic from the store
func (mong *MongoStore) RemoveTopic(project string, name string) error {
	topic := QTopic{project, name}
	return mong.RemoveResource("topics", topic)
}

// RemoveSub removes a subscription from the store
func (mong *MongoStore) RemoveSub(project string, name string) error {
	sub := bson.M{"project": project, "name": name}
	return mong.RemoveResource("subscriptions", sub)
}

// ModSubPush modifies the push configuration
func (mong *MongoStore) ModSubPush(project string, name string, push string) error {
	db := mong.Session.DB(mong.Database)
	c := db.C("subscriptions")

	err := c.Update(bson.M{"project": project, "name": name}, bson.M{"$set": bson.M{"push_endpoint": push}})
	return err
}

// InsertResource inserts a new topic object to the datastore
func (mong *MongoStore) InsertResource(col string, res interface{}) error {

	db := mong.Session.DB(mong.Database)
	c := db.C(col)

	err := c.Insert(res)
	if err != nil {
		log.Fatalf("%s\t%s\t%s", "FATAL", "STORE", err.Error())
	}

	return err
}

// RemoveResource removes a resource from the store
func (mong *MongoStore) RemoveResource(col string, res interface{}) error {

	db := mong.Session.DB(mong.Database)
	c := db.C(col)

	err := c.Remove(res)
	if err != nil {
		log.Fatalf("%s\t%s\t%s", "FATAL", "STORE", err.Error())
	}

	return err
}

// QuerySubs Query Subscription info from store
func (mong *MongoStore) QuerySubs() []QSub {

	db := mong.Session.DB(mong.Database)
	c := db.C("subscriptions")
	var results []QSub
	err := c.Find(bson.M{}).All(&results)
	if err != nil {
		log.Fatalf("%s\t%s\t%s", "FATAL", "STORE", err.Error())
	}
	return results

}

// QueryPushSubs retrieves subscriptions that have a push_endpoint defined
func (mong *MongoStore) QueryPushSubs() []QSub {

	db := mong.Session.DB(mong.Database)
	c := db.C("subscriptions")
	var results []QSub
	err := c.Find(bson.M{"push_endpoint": bson.M{"$ne": nil}}).All(&results)
	if err != nil {
		log.Fatalf("%s\t%s\t%s", "FATAL", "STORE", err.Error())
	}
	return results

}
