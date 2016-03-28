package stores

import (
	"log"

	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

// MongoStore holds configuration
type MongoStore struct {
	Server   string
	Database string
	// Session  *mgo.Session
}

// NewMongoStore creates new mongo store
func NewMongoStore(server string, db string) *MongoStore {
	mong := MongoStore{}
	mong.Initialize(server, db)
	return &mong
}

// Close is used to close session
func (mong *MongoStore) Close() {
	// mong.Session.Close()
}

// Initialize initializes the mongo store struct
func (mong *MongoStore) Initialize(server string, database string) {

	mong.Server = server
	mong.Database = database

	session, err := mgo.Dial(server)
	if err != nil && session != nil {
		panic(err)
	}

	// mong.Session = session

	log.Printf("%s\t%s\t%s:%s", "INFO", "STORE", "Connected to Mongo: ", mong.Server)
}

// UpdateSubOffset updates a subscription offset
func (mong *MongoStore) UpdateSubOffset(name string, offset int64) {

	session, err := mgo.Dial(mong.Server)
	if err != nil {
		panic(err)
	}

	defer session.Close()
	db := session.DB(mong.Database)
	c := db.C("subscriptions")

	doc := bson.M{"name": name}
	change := bson.M{"$set": bson.M{"offset": offset}}
	err = c.Update(doc, change)
	if err != nil {
		log.Fatal(err)
	}

}

// QueryTopics Query Subscription info from store
func (mong *MongoStore) QueryTopics() []QTopic {

	session, err := mgo.Dial(mong.Server)
	if err != nil {
		panic(err)
	}
	defer session.Close()

	db := session.DB(mong.Database)
	c := db.C("topics")
	var results []QTopic
	err = c.Find(bson.M{}).All(&results)

	if err != nil {
		log.Fatal(err)
	}

	return results
}

//HasResourceRoles returns the roles of a user in a project
func (mong *MongoStore) HasResourceRoles(resource string, roles []string) bool {
	session, err := mgo.Dial(mong.Server)
	if err != nil {
		panic(err)
	}
	defer session.Close()

	db := session.DB(mong.Database)
	c := db.C("roles")
	var results []QRole
	err = c.Find(bson.M{"resource": resource, "roles": bson.M{"$in": roles}}).All(&results)
	if err != nil {
		log.Fatal(err)
	}

	if len(results) > 0 {
		return true
	}

	return false

}

//GetUserRoles returns the roles of a user in a project
func (mong *MongoStore) GetUserRoles(project string, token string) []string {
	session, err := mgo.Dial(mong.Server)
	if err != nil {
		panic(err)
	}
	defer session.Close()

	db := session.DB(mong.Database)
	c := db.C("users")
	var results []QUser
	err = c.Find(bson.M{"project": project, "token": token}).All(&results)
	if err != nil {
		log.Fatal(err)
	}

	if len(results) == 0 {
		return []string{}
	}

	if len(results) > 1 {
		log.Printf("%s\t%s\t%s:%s", "WARNING", "STORE", "Multiple users with the same token: ", token)

	}

	return results[0].Roles

}

// HasProject Returns true if project exists
func (mong *MongoStore) HasProject(project string) bool {
	session, err := mgo.Dial(mong.Server)
	if err != nil {
		panic(err)
	}
	defer session.Close()

	db := session.DB(mong.Database)
	c := db.C("projects")
	var results []QProject
	err = c.Find(bson.M{}).All(&results)
	if err != nil {
		log.Fatal(err)
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
func (mong *MongoStore) InsertSub(project string, name string, topic string, offset int64) error {
	sub := QSub{project, name, topic, offset}
	return mong.InsertResource("subscriptions", sub)
}

// InsertResource inserts a new topic object to the datastore
func (mong *MongoStore) InsertResource(col string, res interface{}) error {
	session, err := mgo.Dial(mong.Server)
	if err != nil {
		panic(err)
	}
	defer session.Close()

	db := session.DB(mong.Database)
	c := db.C(col)

	err = c.Insert(res)
	if err != nil {
		log.Fatal(err)
	}

	return err
}

// QuerySubs Query Subscription info from store
func (mong *MongoStore) QuerySubs() []QSub {

	session, err := mgo.Dial(mong.Server)
	if err != nil {
		panic(err)
	}
	defer session.Close()

	db := session.DB(mong.Database)
	c := db.C("subscriptions")
	var results []QSub
	err = c.Find(bson.M{}).All(&results)
	if err != nil {
		log.Fatal(err)
	}
	return results

}
