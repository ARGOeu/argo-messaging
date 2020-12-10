package stores

import (
	"errors"
	"time"

	log "github.com/sirupsen/logrus"

	"fmt"

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

	// Iterate trying to connect
	for {

		log.WithFields(
			log.Fields{
				"type":            "backend_log",
				"backend_service": "mongo",
				"backend_hosts":   mong.Server,
			},
		).Info("Trying to connect to Mongo")

		session, err := mgo.Dial(mong.Server)
		if err != nil {
			// If connection to datastore failed log error and retry
			log.WithFields(
				log.Fields{
					"type":   "backend_log",
					"server": mong.Server,
				},
			).Error(err.Error())
		} else {
			// If connection succesfull continue
			mong.Session = session
			log.WithFields(
				log.Fields{
					"type":            "backend_log",
					"backend_service": "mongo",
					"backend_hosts":   mong.Server,
				},
			).Info("Connection to Mongo established successfully")
			break // connected so continue
		}
	}
}

// SubscriptionsCount returns the amount of subscriptions created in the given time period
func (mong *MongoStore) SubscriptionsCount(startDate, endDate time.Time) (int, error) {
	return mong.getDocCountForCollection(startDate, endDate, "subscriptions")
}

// TopicsCount returns the amount of topics created in the given time period
func (mong *MongoStore) TopicsCount(startDate, endDate time.Time) (int, error) {
	return mong.getDocCountForCollection(startDate, endDate, "topics")
}

// UserCount returns the amount of users created in the given time period
func (mong *MongoStore) UsersCount(startDate, endDate time.Time) (int, error) {
	return mong.getDocCountForCollection(startDate, endDate, "users")
}

// getDocCountForCollection returns the document count for a collection in a given time period
// collection should support field created_on
func (mong *MongoStore) getDocCountForCollection(startDate, endDate time.Time, col string) (int, error) {

	query := bson.M{
		"created_on": bson.M{
			"$gte": startDate,
			"$lte": endDate,
		},
	}

	db := mong.Session.DB(mong.Database)
	c := db.C(col)

	count, err := c.Find(query).Count()
	if err != nil {
		log.WithFields(
			log.Fields{
				"type":            "backend_log",
				"backend_service": "mongo",
				"backend_hosts":   mong.Server,
			},
		).Fatal(err.Error())
	}

	return count, nil
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
		log.WithFields(
			log.Fields{
				"type":            "backend_log",
				"backend_service": "mongo",
				"backend_hosts":   mong.Server,
			},
		).Fatal(err.Error())
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

// RegisterUser inserts a new user registration to the database
func (mong *MongoStore) RegisterUser(uuid, name, firstName, lastName, email, org, desc, registeredAt, atkn, status string) error {

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

	return mong.InsertResource("user_registrations", ur)
}

func (mong *MongoStore) QueryRegistrations(regUUID, status, activationToken, name, email, org string) ([]QUserRegistration, error) {

	query := bson.M{}

	if regUUID != "" {
		query["uuid"] = regUUID
	}

	if status != "" {
		query["status"] = status
	}

	if activationToken != "" {
		query["activation_token"] = activationToken
	}

	if name != "" {
		query["name"] = name
	}

	if email != "" {
		query["email"] = email
	}

	if org != "" {
		query["organization"] = org
	}

	qur := []QUserRegistration{}

	db := mong.Session.DB(mong.Database)
	c := db.C("user_registrations")
	err := c.Find(query).All(&qur)
	if err != nil {
		return qur, err
	}

	return qur, nil
}

func (mong *MongoStore) UpdateRegistration(regUUID, status, modifiedBy, modifiedAt string) error {

	db := mong.Session.DB(mong.Database)
	c := db.C("user_registrations")

	ur := bson.M{"uuid": regUUID}
	change := bson.M{
		"$set": bson.M{
			"status":           status,
			"modified_by":      modifiedBy,
			"modified_at":      modifiedAt,
			"activation_token": "",
		},
	}
	return c.Update(ur, change)
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

// AppendToUserProjects appends a new unique project to the user's projects
func (mong *MongoStore) AppendToUserProjects(userUUID string, projectUUID string, pRoles ...string) error {

	db := mong.Session.DB(mong.Database)
	c := db.C("users")

	err := c.Update(
		bson.M{"uuid": userUUID},
		bson.M{
			"$addToSet": bson.M{
				"projects": QProjectRoles{
					ProjectUUID: projectUUID,
					Roles:       pRoles,
				},
			},
		},
	)

	if err != nil {
		log.WithFields(
			log.Fields{
				"type":            "backend_log",
				"backend_service": "mongo",
				"backend_hosts":   mong.Server,
			},
		).Fatal(err.Error())
	}

	return nil
}

// UpdateUser updates user information
func (mong *MongoStore) UpdateUser(uuid, fname, lname, org, desc string, projects []QProjectRoles, name string, email string, serviceRoles []string, modifiedOn time.Time) error {
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

	if fname != "" {
		curUsr.FirstName = fname
	}

	if lname != "" {
		curUsr.LastName = lname
	}

	if org != "" {
		curUsr.Organization = org
	}

	if desc != "" {
		curUsr.Description = desc
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
		log.WithFields(
			log.Fields{
				"type":            "backend_log",
				"backend_service": "mongo",
				"backend_hosts":   mong.Server,
			},
		).Fatal(err.Error())
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
		log.WithFields(
			log.Fields{
				"type":            "backend_log",
				"backend_service": "mongo",
				"backend_hosts":   mong.Server,
			},
		).Fatal(err.Error())
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
		log.WithFields(
			log.Fields{
				"type":            "backend_log",
				"backend_service": "mongo",
				"backend_hosts":   mong.Server,
			},
		).Fatal(err.Error())
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
		log.WithFields(
			log.Fields{
				"type":            "backend_log",
				"backend_service": "mongo",
				"backend_hosts":   mong.Server,
			},
		).Fatal(err.Error())
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
		log.WithFields(
			log.Fields{
				"type":            "backend_log",
				"backend_service": "mongo",
				"backend_hosts":   mong.Server,
			},
		).Fatal(err.Error())
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
		query = bson.M{"projects.project_uuid": projectUUID}
		if uuid != "" {
			query = bson.M{"projects.project_uuid": projectUUID, "uuid": uuid}
		} else if name != "" {
			query = bson.M{"projects.project_uuid": projectUUID, "name": name}
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
		log.WithFields(
			log.Fields{
				"type":            "backend_log",
				"backend_service": "mongo",
				"backend_hosts":   mong.Server,
			},
		).Fatal(err.Error())
	}

	return results, err
}

// PaginatedQueryUsers returns a page of users
func (mong *MongoStore) PaginatedQueryUsers(pageToken string, pageSize int32, projectUUID string) ([]QUser, int32, string, error) {

	var qUsers []QUser
	var totalSize int32
	var limit int32
	var size int
	var nextPageToken string
	var err error
	var ok bool
	var query bson.M

	// if the page size is other than zero(where zero means, no limit), try to grab one more document to check if there
	// will be a next page after the current one
	if pageSize > 0 {
		limit = pageSize + 1
	}

	// if projectUUID is empty string return all users, if projectUUID has a non empty value
	// query users that belong to that project
	if projectUUID != "" {
		query = bson.M{
			"projects": bson.M{
				"$elemMatch": bson.M{
					"project_uuid": projectUUID,
				},
			},
		}
	}

	// select db collection
	db := mong.Session.DB(mong.Database)
	c := db.C("users")

	// check the total of the users selected by the query not taking into account pagination
	if size, err = c.Find(query).Count(); err != nil {
		log.WithFields(
			log.Fields{
				"type":            "backend_log",
				"backend_service": "mongo",
				"backend_hosts":   mong.Server,
			},
		).Fatal(err.Error())
	}
	totalSize = int32(size)

	// now take into account if pagination is enabled and change the query accordingly
	// first check if an pageToken is provided and whether or not is a valid bson ID
	if pageToken != "" {
		if ok = bson.IsObjectIdHex(pageToken); !ok {
			err = fmt.Errorf("Page token %v is not a valid bson ObjectId", pageToken)
			log.WithFields(
				log.Fields{
					"type":            "backend_log",
					"backend_service": "mongo",
					"page_token":      pageToken,
				},
			).Error("Page token is not a valid bson ObjectId")
			return qUsers, totalSize, nextPageToken, err
		}

		bsonID := bson.ObjectIdHex(pageToken)
		// now that the paginated query is constructed from start take into account again
		// if projectUUID is provided to query only the users of a given project
		if projectUUID != "" {
			query = bson.M{
				"projects": bson.M{
					"$elemMatch": bson.M{
						"project_uuid": projectUUID,
					},
				},
				"_id": bson.M{
					"$lte": bsonID,
				},
			}

		} else {

			query = bson.M{
				"_id": bson.M{
					"$lte": bsonID,
				},
			}
		}

	}

	if err = c.Find(query).Sort("-_id").Limit(int(limit)).All(&qUsers); err != nil {
		log.WithFields(
			log.Fields{
				"type":            "backend_log",
				"backend_service": "mongo",
				"backend_hosts":   mong.Server,
			},
		).Fatal(err.Error())
	}

	// if the amount of users that were found was equal to the limit, its a sign that there are users to populate the next page
	// so pick the last element's pageToken to use as the starting point for the next page
	// and eliminate the extra element from the current response
	if pageSize > 0 && len(qUsers) > 0 && len(qUsers) == int(limit) {

		nextPageToken = qUsers[limit-1].ID.(bson.ObjectId).Hex()
		qUsers = qUsers[:len(qUsers)-1]
	}

	return qUsers, totalSize, nextPageToken, err
}

//QuerySubsByTopic returns subscriptions of a specific topic
func (mong *MongoStore) QuerySubsByTopic(projectUUID, topic string) ([]QSub, error) {
	// By default return all subs of a given project
	query := bson.M{"project_uuid": projectUUID}

	// If topic is given return only the specific topic
	if topic != "" {
		query = bson.M{"project_uuid": projectUUID, "topic": topic}
	}
	db := mong.Session.DB(mong.Database)
	c := db.C("subscriptions")
	var results []QSub
	err := c.Find(query).All(&results)

	if err != nil {
		log.WithFields(
			log.Fields{
				"type":            "backend_log",
				"backend_service": "mongo",
				"backend_hosts":   mong.Server,
			},
		).Fatal(err.Error())
	}

	return results, err
}

//QuerySubsByACL returns subscriptions that a specific username has access to
func (mong *MongoStore) QuerySubsByACL(projectUUID, user string) ([]QSub, error) {
	// By default return all subs of a given project
	query := bson.M{"project_uuid": projectUUID}

	// If name is given return only the specific topic
	if user != "" {
		query = bson.M{"project_uuid": projectUUID, "acl": user}
	}
	db := mong.Session.DB(mong.Database)
	c := db.C("subscriptions")
	var results []QSub
	err := c.Find(query).All(&results)

	if err != nil {
		log.WithFields(
			log.Fields{
				"type":            "backend_log",
				"backend_service": "mongo",
				"backend_hosts":   mong.Server,
			},
		).Fatal(err.Error())
	}

	return results, err
}

//QueryTopicsByACL returns topics that a specific username has access to
func (mong *MongoStore) QueryTopicsByACL(projectUUID, user string) ([]QTopic, error) {
	// By default return all topics of a given project
	query := bson.M{"project_uuid": projectUUID}

	// If name is given return only the specific topic
	if user != "" {
		query = bson.M{"project_uuid": projectUUID, "acl": user}
	}
	db := mong.Session.DB(mong.Database)
	c := db.C("topics")
	var results []QTopic
	err := c.Find(query).All(&results)

	if err != nil {
		log.WithFields(
			log.Fields{
				"type":            "backend_log",
				"backend_service": "mongo",
				"backend_hosts":   mong.Server,
			},
		).Fatal(err.Error())
	}

	return results, err
}

// QueryTopics Query Subscription info from store
func (mong *MongoStore) QueryTopics(projectUUID, userUUID, name, pageToken string, pageSize int32) ([]QTopic, int32, string, error) {

	var err error
	var totalSize int32
	var limit int32
	var nextPageToken string
	var qTopics []QTopic
	var ok bool
	var size int

	// By default return all topics of a given project
	query := bson.M{"project_uuid": projectUUID}

	// find all the topics for a specific user
	if userUUID != "" {
		query["acl"] = bson.M{"$in": []string{userUUID}}
	}

	// if the page size is other than zero(where zero means, no limit), try to grab one more document to check if there
	// will be a next page after the current one
	if pageSize > 0 {

		limit = pageSize + 1

	}

	// first check if an pageToken is provided and whether or not is a valid bson ID
	if pageToken != "" {
		if ok = bson.IsObjectIdHex(pageToken); !ok {
			err = fmt.Errorf("Page token %v is not a valid bson ObjectId", pageToken)
			log.WithFields(
				log.Fields{
					"type":            "backend_log",
					"backend_service": "mongo",
					"page_token":      pageToken,
				},
			).Error("Page token is not a valid bson ObjectId")
			return qTopics, totalSize, nextPageToken, err
		}

		bsonID := bson.ObjectIdHex(pageToken)

		query["_id"] = bson.M{"$lte": bsonID}

	} else if name != "" {

		query["name"] = name
	}

	db := mong.Session.DB(mong.Database)
	c := db.C("topics")

	if err = c.Find(query).Sort("-_id").Limit(int(limit)).All(&qTopics); err != nil {
		log.WithFields(
			log.Fields{
				"type":            "backend_log",
				"backend_service": "mongo",
				"backend_hosts":   mong.Server,
			},
		).Fatal(err.Error())
	}

	if name == "" {

		countQuery := bson.M{"project_uuid": projectUUID}
		if userUUID != "" {
			countQuery["acl"] = bson.M{"$in": []string{userUUID}}
		}

		if size, err = c.Find(countQuery).Count(); err != nil {
			log.WithFields(
				log.Fields{
					"type":            "backend_log",
					"backend_service": "mongo",
					"backend_hosts":   mong.Server,
				},
			).Fatal(err.Error())
		}

		totalSize = int32(size)

		// if the amount of topics that were found was equal to the limit, its a sign that there are topics to populate the next page
		// so pick the last element's pageToken to use as the starting point for the next page
		// and eliminate the extra element from the current response
		if len(qTopics) > 0 && len(qTopics) == int(limit) {

			nextPageToken = qTopics[limit-1].ID.(bson.ObjectId).Hex()
			qTopics = qTopics[:len(qTopics)-1]
		}
	}

	return qTopics, totalSize, nextPageToken, err

}

// UpdateTopicLatestPublish updates the topic's latest publish time
func (mong *MongoStore) UpdateTopicLatestPublish(projectUUID string, name string, date time.Time) error {

	db := mong.Session.DB(mong.Database)
	c := db.C("topics")

	doc := bson.M{
		"project_uuid": projectUUID,
		"name":         name,
	}

	change := bson.M{
		"$set": bson.M{
			"latest_publish": date,
		},
	}

	return c.Update(doc, change)
}

// UpdateTopicPublishRate updates the topic's publishing rate
func (mong *MongoStore) UpdateTopicPublishRate(projectUUID string, name string, rate float64) error {

	db := mong.Session.DB(mong.Database)
	c := db.C("topics")

	doc := bson.M{
		"project_uuid": projectUUID,
		"name":         name,
	}

	change := bson.M{
		"$set": bson.M{
			"publish_rate": rate,
		},
	}

	return c.Update(doc, change)
}

// UpdateSubLatestConsume updates the subscription's latest consume time
func (mong *MongoStore) UpdateSubLatestConsume(projectUUID string, name string, date time.Time) error {

	db := mong.Session.DB(mong.Database)
	c := db.C("subscriptions")

	doc := bson.M{
		"project_uuid": projectUUID,
		"name":         name,
	}

	change := bson.M{
		"$set": bson.M{
			"latest_consume": date,
		},
	}

	return c.Update(doc, change)
}

// UpdateSubConsumeRate updates the subscription's consume rate
func (mong *MongoStore) UpdateSubConsumeRate(projectUUID string, name string, rate float64) error {

	db := mong.Session.DB(mong.Database)
	c := db.C("subscriptions")

	doc := bson.M{
		"project_uuid": projectUUID,
		"name":         name,
	}

	change := bson.M{
		"$set": bson.M{
			"consume_rate": rate,
		},
	}

	return c.Update(doc, change)
}

// QueryDailyTopicMsgCount returns results regarding the number of messages published to a topic
func (mong *MongoStore) QueryDailyTopicMsgCount(projectUUID string, topicName string, date time.Time) ([]QDailyTopicMsgCount, error) {

	var err error
	var qDailyTopicMsgCount []QDailyTopicMsgCount
	var query bson.M

	// represents an empty time object
	var zeroValueTime time.Time

	query = bson.M{"date": date, "project_uuid": projectUUID, "topic_name": topicName}

	// if nothing's specified return the whole collection
	if projectUUID == "" && topicName == "" && date == zeroValueTime {
		query = bson.M{}
	}

	if projectUUID != "" && topicName != "" && date == zeroValueTime {
		query = bson.M{"project_uuid": projectUUID, "topic_name": topicName}
	}

	db := mong.Session.DB(mong.Database)
	c := db.C("daily_topic_msg_count")

	err = c.Find(query).Sort("-date").Limit(30).All(&qDailyTopicMsgCount)

	if err != nil {
		log.WithFields(
			log.Fields{
				"type":            "backend_log",
				"backend_service": "mongo",
				"backend_hosts":   mong.Server,
			},
		).Fatal(err.Error())
	}

	return qDailyTopicMsgCount, err
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

// IncrementDailyTopicMsgCount increments the daily count of published messages to a specific topic
func (mong *MongoStore) IncrementDailyTopicMsgCount(projectUUID string, topicName string, num int64, date time.Time) error {

	db := mong.Session.DB(mong.Database)
	c := db.C("daily_topic_msg_count")

	doc := bson.M{"date": date, "project_uuid": projectUUID, "topic_name": topicName}
	change := bson.M{"$inc": bson.M{"msg_count": num}}

	_, err := c.Upsert(doc, change)

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

// IncrementSubMsgNum increments the number of messages pulled in a subscription
func (mong *MongoStore) IncrementSubMsgNum(projectUUID string, name string, num int64) error {

	db := mong.Session.DB(mong.Database)
	c := db.C("subscriptions")

	doc := bson.M{"project_uuid": projectUUID, "name": name}
	change := bson.M{"$inc": bson.M{"msg_num": num}}

	err := c.Update(doc, change)

	return err

}

//IncrementSubBytes increases the total number of bytes consumed from a subscripion
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
		log.WithFields(
			log.Fields{
				"type":            "backend_log",
				"backend_service": "mongo",
				"backend_hosts":   mong.Server,
			},
		).Fatal(err.Error())
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
		log.WithFields(
			log.Fields{
				"type":            "backend_log",
				"backend_service": "mongo",
				"backend_hosts":   mong.Server,
			},
		).Fatal(err.Error())
	}
	return results
}

//GetOpMetrics returns the operational metrics from datastore
func (mong *MongoStore) GetOpMetrics() []QopMetric {

	db := mong.Session.DB(mong.Database)
	c := db.C("op_metrics")
	var results []QopMetric

	err := c.Find(bson.M{}).All(&results)

	if err != nil {
		log.WithFields(
			log.Fields{
				"type":            "backend_log",
				"backend_service": "mongo",
				"backend_hosts":   mong.Server,
			},
		).Fatal(err.Error())
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
		log.WithFields(
			log.Fields{
				"type":            "backend_log",
				"backend_service": "mongo",
				"backend_hosts":   mong.Server,
			},
		).Fatal(err.Error())
	}

	if len(results) == 0 {
		return []string{}, ""
	}

	if len(results) > 1 {
		log.WithFields(
			log.Fields{
				"type":            "backend_log",
				"token":           token,
				"backend_service": "mongo",
				"backend_hosts":   mong.Server,
			},
		).Warning("Multiple users with the same token")
	}

	// Search the found user for project roles
	return results[0].getProjectRoles(projectUUID), results[0].Name

}

//GetUserFromToken returns user information from a specific token
func (mong *MongoStore) GetUserFromToken(token string) (QUser, error) {

	db := mong.Session.DB(mong.Database)
	c := db.C("users")
	var results []QUser

	err := c.Find(bson.M{"token": token}).All(&results)

	if err != nil {
		log.WithFields(
			log.Fields{
				"type":            "backend_log",
				"backend_service": "mongo",
				"backend_hosts":   mong.Server,
			},
		).Fatal(err.Error())
	}

	if len(results) == 0 {
		return QUser{}, errors.New("not found")
	}

	if len(results) > 1 {
		log.WithFields(
			log.Fields{
				"type":            "backend_log",
				"token":           token,
				"backend_service": "mongo",
				"backend_hosts":   mong.Server,
			},
		).Warning("Multiple users with the same token")
	}

	// Search the found user for project roles
	return results[0], err

}

// QueryOneSub queries and returns specific sub of project
func (mong *MongoStore) QueryOneSub(projectUUID string, name string) (QSub, error) {

	db := mong.Session.DB(mong.Database)
	c := db.C("subscriptions")
	var results []QSub
	err := c.Find(bson.M{"project_uuid": projectUUID, "name": name}).All(&results)
	if err != nil {
		log.WithFields(
			log.Fields{
				"type":            "backend_log",
				"backend_service": "mongo",
				"backend_hosts":   mong.Server,
			},
		).Fatal(err.Error())
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
		log.WithFields(
			log.Fields{
				"type":            "backend_log",
				"backend_service": "mongo",
				"backend_hosts":   mong.Server,
			},
		).Fatal(err.Error())
	}

	if len(results) > 0 {
		return true
	}
	return false
}

// InsertTopic inserts a topic to the store
func (mong *MongoStore) InsertTopic(projectUUID string, name string, schemaUUID string, createdOn time.Time) error {

	topic := QTopic{
		ProjectUUID:   projectUUID,
		Name:          name,
		MsgNum:        0,
		TotalBytes:    0,
		LatestPublish: time.Time{},
		PublishRate:   0,
		SchemaUUID:    schemaUUID,
		CreatedOn:     createdOn,
	}

	return mong.InsertResource("topics", topic)
}

// InsertOpMetric inserts an operational metric
func (mong *MongoStore) InsertOpMetric(hostname string, cpu float64, mem float64) error {
	opMetric := QopMetric{Hostname: hostname, CPU: cpu, MEM: mem}
	db := mong.Session.DB(mong.Database)
	c := db.C("op_metrics")

	upsertdata := bson.M{"$set": opMetric}

	_, err := c.UpsertId(opMetric.Hostname, upsertdata)
	if err != nil {
		log.WithFields(
			log.Fields{
				"type":            "backend_log",
				"backend_service": "mongo",
				"backend_hosts":   mong.Server,
			},
		).Error(err.Error())
	}

	return err
}

// InsertUser inserts a new user to the store
func (mong *MongoStore) InsertUser(uuid string, projects []QProjectRoles, name string, fname string, lname string, org string, desc string, token string, email string, serviceRoles []string, createdOn time.Time, modifiedOn time.Time, createdBy string) error {
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
	return mong.InsertResource("users", user)
}

// InsertProject inserts a project to the store
func (mong *MongoStore) InsertProject(uuid string, name string, createdOn time.Time, modifiedOn time.Time, createdBy string, description string) error {
	project := QProject{UUID: uuid, Name: name, CreatedOn: createdOn, ModifiedOn: modifiedOn, CreatedBy: createdBy, Description: description}
	return mong.InsertResource("projects", project)
}

// InsertSub inserts a subscription to the store
func (mong *MongoStore) InsertSub(projectUUID string, name string, topic string, offset int64, maxMessages int64, authzType string, authzHeader string, ack int, push string, rPolicy string, rPeriod int, vhash string, verified bool, createdOn time.Time) error {
	sub := QSub{
		ProjectUUID:         projectUUID,
		Name:                name,
		Topic:               topic,
		Offset:              offset,
		NextOffset:          0,
		PendingAck:          "",
		Ack:                 ack,
		MaxMessages:         maxMessages,
		AuthorizationType:   authzType,
		AuthorizationHeader: authzHeader,
		PushEndpoint:        push,
		RetPolicy:           rPolicy,
		RetPeriod:           rPeriod,
		VerificationHash:    vhash,
		Verified:            verified,
		MsgNum:              0,
		TotalBytes:          0,
		CreatedOn:           createdOn,
	}
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

// QueryTotalMessagesPerProject returns the total amount of messages per project for the given time window
func (mong *MongoStore) QueryTotalMessagesPerProject(projectUUIDs []string, startDate time.Time, endDate time.Time) ([]QProjectMessageCount, error) {

	var err error
	var qdp []QProjectMessageCount

	c := mong.Session.DB(mong.Database).C("daily_topic_msg_count")

	if endDate.Before(startDate) {
		startDate, endDate = endDate, startDate
	}

	days := 1
	if !endDate.Equal(startDate) {
		days = int(endDate.Sub(startDate).Hours() / 24)
		// add an extra day to compensate for the fact that we need the starting day included as well
		// e.g. Aug 1 to Aug 31 should be calculated as 31 days and not as 30
		days += 1
	}

	condQuery := []bson.M{
		{
			"date": bson.M{
				"$gte": startDate,
			},
		},
		{
			"date": bson.M{
				"$lte": endDate,
			},
		},
	}

	if len(projectUUIDs) > 0 {
		condQuery = append(condQuery, bson.M{
			"project_uuid": bson.M{
				"$in": projectUUIDs,
			},
		},
		)
	}

	query := []bson.M{
		{
			"$match": bson.M{
				"$and": condQuery,
			},
		},
		{
			"$group": bson.M{
				"_id": bson.M{
					"project_uuid": "$project_uuid",
				},
				"msg_count": bson.M{
					"$sum": "$msg_count",
				},
			},
		},
		{
			"$project": bson.M{
				"_id":          0,
				"project_uuid": "$_id.project_uuid",
				"msg_count":    1,
				"avg_daily_msg": bson.M{
					"$divide": []interface{}{"$msg_count", days},
				},
			},
		},
	}

	if err = c.Pipe(query).All(&qdp); err != nil {
		log.WithFields(
			log.Fields{
				"type":            "backend_log",
				"backend_service": "mongo",
				"backend_hosts":   mong.Server,
			},
		).Fatal(err.Error())
	}

	return qdp, err
}

// QueryDailyProjectMsgCount queries the total messages per day for a given project
func (mong *MongoStore) QueryDailyProjectMsgCount(projectUUID string) ([]QDailyProjectMsgCount, error) {

	var err error
	var qdp []QDailyProjectMsgCount

	c := mong.Session.DB(mong.Database).C("daily_topic_msg_count")

	query := []bson.M{
		{
			"$match": bson.M{
				"project_uuid": projectUUID,
			},
		},
		{
			"$group": bson.M{
				"_id": bson.M{
					"date": "$date",
				},
				"msg_count": bson.M{
					"$sum": "$msg_count",
				},
			},
		},
		{
			"$sort": bson.M{
				"_id": -1,
			},
		},
		{
			"$limit": 30,
		},
		{
			"$project": bson.M{
				"_id":       0,
				"date":      "$_id.date",
				"msg_count": 1,
			},
		},
	}

	if err = c.Pipe(query).All(&qdp); err != nil {
		log.WithFields(
			log.Fields{
				"type":            "backend_log",
				"backend_service": "mongo",
				"backend_hosts":   mong.Server,
			},
		).Fatal(err.Error())
	}

	return qdp, err

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

// ExistsInACL checks if a user is part of a topic's or sub's acl
func (mong *MongoStore) ExistsInACL(projectUUID string, resource string, resourceName string, userUUID string) error {

	db := mong.Session.DB(mong.Database)

	if resource != "topics" && resource != "subscriptions" {
		return errors.New("wrong resource type")
	}

	c := db.C(resource)

	query := bson.M{
		"project_uuid": projectUUID,
		"name":         resourceName,
		"acl": bson.M{
			"$in": []string{userUUID},
		},
	}

	res := map[string]interface{}{}
	return c.Find(query).One(&res)
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

// AppendToACL adds additional users to an existing ACL
func (mong *MongoStore) AppendToACL(projectUUID string, resource string, name string, acl []string) error {

	db := mong.Session.DB(mong.Database)

	if resource != "topics" && resource != "subscriptions" {
		return errors.New("wrong resource type")
	}

	c := db.C(resource)

	err := c.Update(
		bson.M{
			"project_uuid": projectUUID,
			"name":         name,
		},
		bson.M{
			"$addToSet": bson.M{
				"acl": bson.M{
					"$each": acl,
				},
			}})
	return err
}

// RemoveFromACL remves users for a given ACL
func (mong *MongoStore) RemoveFromACL(projectUUID string, resource string, name string, acl []string) error {

	db := mong.Session.DB(mong.Database)

	if resource != "topics" && resource != "subscriptions" {
		return errors.New("wrong resource type")
	}

	c := db.C(resource)

	err := c.Update(
		bson.M{
			"project_uuid": projectUUID,
			"name":         name,
		},
		bson.M{
			"$pullAll": bson.M{
				"acl": acl,
			},
		})

	return err
}

// ModAck modifies the subscription's ack timeout field in mongodb
func (mong *MongoStore) ModAck(projectUUID string, name string, ack int) error {
	log.Info("Modifying Ack Deadline", ack)
	db := mong.Session.DB(mong.Database)
	c := db.C("subscriptions")
	err := c.Update(bson.M{"project_uuid": projectUUID, "name": name}, bson.M{"$set": bson.M{"ack": ack}})
	return err
}

// ModSubPush modifies the push configuration
func (mong *MongoStore) ModSubPush(projectUUID string, name string, push string, authzType string, authzValue string, maxMessages int64, rPolicy string, rPeriod int, vhash string, verified bool) error {
	db := mong.Session.DB(mong.Database)
	c := db.C("subscriptions")

	err := c.Update(bson.M{
		"project_uuid": projectUUID,
		"name":         name,
	},
		bson.M{"$set": bson.M{
			"push_endpoint":        push,
			"authorization_type":   authzType,
			"authorization_header": authzValue,
			"max_messages":         maxMessages,
			"retry_policy":         rPolicy,
			"retry_period":         rPeriod,
			"verification_hash":    vhash,
			"verified":             verified,
		},
		})
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
func (mong *MongoStore) QuerySubs(projectUUID, userUUID, name, pageToken string, pageSize int32) ([]QSub, int32, string, error) {

	var err error
	var totalSize int32
	var limit int32
	var nextPageToken string
	var qSubs []QSub
	var ok bool
	var size int

	// By default return all subs of a given project
	query := bson.M{"project_uuid": projectUUID}

	// find all the subscriptions for a specific user
	if userUUID != "" {
		query["acl"] = bson.M{"$in": []string{userUUID}}
	}

	// if the page size is other than zero(where zero means, no limit), try to grab one more document to check if there
	// will be a next page after the current one
	if pageSize > 0 {

		limit = pageSize + 1

	}

	// first check if an pageToken is provided and whether or not is a valid bson ID
	if pageToken != "" {
		if ok = bson.IsObjectIdHex(pageToken); !ok {
			err = fmt.Errorf("Page token %v is not a valid bson ObjectId", pageToken)
			log.WithFields(
				log.Fields{
					"type":            "backend_log",
					"backend_service": "mongo",
					"page_token":      pageToken,
				},
			).Error("Page token is not a valid bson ObjectId")
			return qSubs, totalSize, nextPageToken, err
		}

		bsonID := bson.ObjectIdHex(pageToken)

		query["_id"] = bson.M{"$lte": bsonID}

	} else if name != "" {

		query["name"] = name
	}

	db := mong.Session.DB(mong.Database)
	c := db.C("subscriptions")

	if err = c.Find(query).Sort("-_id").Limit(int(limit)).All(&qSubs); err != nil {
		log.WithFields(
			log.Fields{
				"type":            "backend_log",
				"backend_service": "mongo",
				"backend_hosts":   mong.Server,
			},
		).Fatal(err.Error())
	}

	if name == "" {

		countQuery := bson.M{"project_uuid": projectUUID}
		if userUUID != "" {
			countQuery["acl"] = bson.M{"$in": []string{userUUID}}
		}

		if size, err = c.Find(countQuery).Count(); err != nil {
			log.WithFields(
				log.Fields{
					"type":            "backend_log",
					"backend_service": "mongo",
					"backend_hosts":   mong.Server,
				},
			).Fatal(err.Error())
		}

		totalSize = int32(size)

		// if the amount of subscriptions that were found was equal to the limit, its a sign that there are subscriptions to populate the next page
		// so pick the last element's pageToken to use as the starting point for the next page
		// and eliminate the extra element from the current response
		if len(qSubs) > 0 && len(qSubs) == int(limit) {

			nextPageToken = qSubs[limit-1].ID.(bson.ObjectId).Hex()
			qSubs = qSubs[:len(qSubs)-1]
		}
	}

	return qSubs, totalSize, nextPageToken, err

}

// QueryPushSubs retrieves subscriptions that have a push_endpoint defined
func (mong *MongoStore) QueryPushSubs() []QSub {

	db := mong.Session.DB(mong.Database)
	c := db.C("subscriptions")
	var results []QSub
	err := c.Find(bson.M{"push_endpoint": bson.M{"$ne": ""}}).All(&results)
	if err != nil {
		log.WithFields(
			log.Fields{
				"type":            "backend_log",
				"backend_service": "mongo",
				"backend_hosts":   mong.Server,
			},
		).Fatal(err.Error())
	}
	return results

}

func (mong *MongoStore) InsertSchema(projectUUID, schemaUUID, name, schemaType, rawSchemaString string) error {
	sub := QSchema{
		ProjectUUID: projectUUID,
		UUID:        schemaUUID,
		Name:        name,
		Type:        schemaType,
		RawSchema:   rawSchemaString,
	}
	return mong.InsertResource("schemas", sub)
}

func (mong *MongoStore) QuerySchemas(projectUUID, schemaUUID, name string) ([]QSchema, error) {

	db := mong.Session.DB(mong.Database)
	c := db.C("schemas")

	var results []QSchema

	query := bson.M{"project_uuid": projectUUID}

	if name != "" {
		query["name"] = name
	}

	if schemaUUID != "" {
		query["uuid"] = schemaUUID
	}

	err := c.Find(query).All(&results)
	if err != nil {
		log.WithFields(
			log.Fields{
				"type":            "backend_log",
				"backend_service": "mongo",
				"backend_hosts":   mong.Server,
			},
		).Fatal(err.Error())
	}

	return results, nil
}

// UpdateSchema updates the fields of a schema
func (mong *MongoStore) UpdateSchema(schemaUUID, name, schemaType, rawSchemaString string) error {

	db := mong.Session.DB(mong.Database)
	c := db.C("schemas")

	selector := bson.M{"uuid": schemaUUID}

	updates := bson.M{}

	if name != "" {
		updates["name"] = name
	}

	if schemaType != "" {
		updates["type"] = schemaType
	}

	if rawSchemaString != "" {
		updates["raw_schema"] = rawSchemaString
	}

	change := bson.M{"$set": updates}

	return c.Update(selector, change)
}

// DeleteSchema removes the schema from the store
// It also clears all the respective topics from the schema_uuid of the deleted schema
func (mong *MongoStore) DeleteSchema(schemaUUID string) error {

	db := mong.Session.DB(mong.Database)
	c := db.C("schemas")

	selector := bson.M{"uuid": schemaUUID}

	err := c.Remove(selector)

	if err != nil {
		log.WithFields(
			log.Fields{
				"type":            "backend_log",
				"backend_service": "mongo",
				"backend_hosts":   mong.Server,
			},
		).Fatal(err.Error())
	}

	topics := db.C("topics")

	topicSelector := bson.M{"schema_uuid": schemaUUID}
	change := bson.M{
		"$set": bson.M{
			"schema_uuid": "",
		},
	}
	topics.UpdateAll(topicSelector, change)

	if err != nil {
		log.WithFields(
			log.Fields{
				"type":            "backend_log",
				"backend_service": "mongo",
				"backend_hosts":   mong.Server,
			},
		).Fatal(err.Error())
	}

	return nil

}
