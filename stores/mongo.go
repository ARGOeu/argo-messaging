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

// QSubs are the results of the Qsub query
type QSubs struct {
	Project string `bson:"project"`
	Name    string `bson:"name"`
	Topic   string `bson:"topic"`
	Offset  int64  `bson:"offset"`
}

// QTopics are the results of the QTopic query
type QTopics struct {
	Project string `bson:"project"`
	Name    string `bson:"name"`
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
func (mong *MongoStore) QueryTopics() []QTopics {

	session, err := mgo.Dial(mong.Server)
	if err != nil {
		panic(err)
	}
	defer session.Close()

	db := session.DB(mong.Database)
	c := db.C("topics")
	var results []QTopics
	err = c.Find(bson.M{}).All(&results)

	if err != nil {
		log.Fatal(err)
	}

	return results
}

// QuerySubs Query Subscription info from store
func (mong *MongoStore) QuerySubs() []QSubs {

	session, err := mgo.Dial(mong.Server)
	if err != nil {
		panic(err)
	}
	defer session.Close()

	db := session.DB(mong.Database)
	c := db.C("subscriptions")
	var results []QSubs
	err = c.Find(bson.M{}).All(&results)
	if err != nil {
		log.Fatal(err)
	}
	return results

}
