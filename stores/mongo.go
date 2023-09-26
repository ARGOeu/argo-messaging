package stores

import (
	"context"
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

func (mong *MongoStore) logErrorAndCrash(ctx context.Context, funcName string, err error) {
	log.WithFields(
		log.Fields{
			"trace_id":        ctx.Value("trace_id"),
			"type":            "backend_log",
			"function":        funcName,
			"backend_service": "mongo",
			"backend_hosts":   mong.Server,
		},
	).Fatal(err.Error())
}

// SubscriptionsCount returns the amount of subscriptions created in the given time period
func (mong *MongoStore) SubscriptionsCount(ctx context.Context, startDate, endDate time.Time, projectUUIDs []string) (map[string]int64, error) {
	return mong.getDocCountForCollection(ctx, startDate, endDate, "subscriptions", projectUUIDs)
}

// TopicsCount returns the amount of topics created in the given time period
func (mong *MongoStore) TopicsCount(ctx context.Context, startDate, endDate time.Time, projectUUIDs []string) (map[string]int64, error) {
	return mong.getDocCountForCollection(ctx, startDate, endDate, "topics", projectUUIDs)
}

// UsersCount returns the amount of users created in the given time period
func (mong *MongoStore) UsersCount(ctx context.Context, startDate, endDate time.Time, projectUUIDs []string) (map[string]int64, error) {
	var resourceCounts []QProjectResourceCount

	condQuery := []bson.M{
		{
			"created_on": bson.M{
				"$gte": startDate,
			},
		},
		{
			"created_on": bson.M{
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
			"$unwind": "$projects",
		},
		{
			"$match": bson.M{
				"$and": condQuery,
			},
		},
		{
			"$group": bson.M{
				"_id": bson.M{
					"project_uuid": "$projects.project_uuid",
				},
				"resource_count": bson.M{
					"$sum": 1,
				},
			},
		},
		{
			"$project": bson.M{
				"_id":            0,
				"project_uuid":   "$_id.project_uuid",
				"resource_count": 1,
			},
		},
	}

	c := mong.Session.DB(mong.Database).C("users")

	if err := c.Pipe(query).All(&resourceCounts); err != nil {
		mong.logErrorAndCrash(ctx, "UsersCount", err)
	}

	res := map[string]int64{}

	for _, t := range resourceCounts {
		res[t.ProjectUUID] = t.Count
	}

	return res, nil
}

// getDocCountForCollection returns the document count for a collection in a given time period
// collection should support field created_on
func (mong *MongoStore) getDocCountForCollection(ctx context.Context, startDate, endDate time.Time, col string, projectUUIDs []string) (map[string]int64, error) {

	var resourceCounts []QProjectResourceCount

	condQuery := []bson.M{
		{
			"created_on": bson.M{
				"$gte": startDate,
			},
		},
		{
			"created_on": bson.M{
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
				"resource_count": bson.M{
					"$sum": 1,
				},
			},
		},
		{
			"$project": bson.M{
				"_id":            0,
				"project_uuid":   "$_id.project_uuid",
				"resource_count": 1,
			},
		},
	}

	c := mong.Session.DB(mong.Database).C(col)

	if err := c.Pipe(query).All(&resourceCounts); err != nil {
		mong.logErrorAndCrash(ctx, "getDocCountForCollection", err)
	}

	res := map[string]int64{}

	for _, t := range resourceCounts {
		res[t.ProjectUUID] = t.Count
	}

	return res, nil
}

// QueryProjects queries the database for a specific project or a list of all projects
func (mong *MongoStore) QueryProjects(ctx context.Context, uuid string, name string) ([]QProject, error) {

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
		mong.logErrorAndCrash(ctx, "QueryProjects", err)
	}

	if len(results) > 0 {
		return results, nil
	}

	return results, errors.New("not found")
}

// UpdateProject updates project information
func (mong *MongoStore) UpdateProject(ctx context.Context, projectUUID string, name string, description string, modifiedOn time.Time) error {
	db := mong.Session.DB(mong.Database)
	c := db.C("projects")

	doc := bson.M{"uuid": projectUUID}
	results, err := mong.QueryProjects(ctx, projectUUID, "")
	if err != nil {
		return err
	}

	curPr := results[0]
	curPr.ModifiedOn = modifiedOn // modifiedOn should always be updated

	if name != "" {
		// Check if name is going to change and if that name already exists
		if name != curPr.Name {
			if sameRes, _ := mong.QueryProjects(ctx, "", name); len(sameRes) > 0 {
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

	return nil

}

// RegisterUser inserts a new user registration to the database
func (mong *MongoStore) RegisterUser(ctx context.Context, uuid, name, firstName, lastName, email, org, desc, registeredAt, atkn, status string) error {

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

	return mong.InsertResource(ctx, "user_registrations", ur)
}

// DeleteRegistration removes the respective registration from the
func (mong *MongoStore) DeleteRegistration(ctx context.Context, uuid string) error {
	return mong.RemoveResource(ctx, "user_registrations", bson.M{"uuid": uuid})
}

func (mong *MongoStore) QueryRegistrations(ctx context.Context, regUUID, status, activationToken, name, email, org string) ([]QUserRegistration, error) {

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

func (mong *MongoStore) UpdateRegistration(ctx context.Context, regUUID, status, declineComment, modifiedBy, modifiedAt string) error {

	db := mong.Session.DB(mong.Database)
	c := db.C("user_registrations")

	ur := bson.M{"uuid": regUUID}
	change := bson.M{
		"$set": bson.M{
			"status":           status,
			"decline_comment":  declineComment,
			"modified_by":      modifiedBy,
			"modified_at":      modifiedAt,
			"activation_token": "",
		},
	}
	return c.Update(ur, change)
}

// UpdateUserToken updates user's token
func (mong *MongoStore) UpdateUserToken(ctx context.Context, uuid string, token string) error {

	db := mong.Session.DB(mong.Database)
	c := db.C("users")

	doc := bson.M{"uuid": uuid}
	change := bson.M{"$set": bson.M{"token": token}}

	err := c.Update(doc, change)

	return err

}

// AppendToUserProjects appends a new unique project to the user's projects
func (mong *MongoStore) AppendToUserProjects(ctx context.Context, userUUID string, projectUUID string, pRoles ...string) error {

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
		mong.logErrorAndCrash(ctx, "AppendToUserProjects", err)
	}

	return nil
}

// UpdateUser updates user information
func (mong *MongoStore) UpdateUser(ctx context.Context, uuid, fname, lname, org, desc string, projects []QProjectRoles, name string, email string, serviceRoles []string, modifiedOn time.Time) error {
	db := mong.Session.DB(mong.Database)
	c := db.C("users")

	doc := bson.M{"uuid": uuid}
	results, err := mong.QueryUsers(ctx, "", uuid, "")
	if err != nil {
		return err
	}

	curUsr := results[0]

	if name != "" {
		// Check if name is going to change and if that name already exists
		if name != curUsr.Name {
			if sameRes, _ := mong.QueryUsers(ctx, "", "", name); len(sameRes) > 0 {
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
func (mong *MongoStore) UpdateSubPull(ctx context.Context, projectUUID string, name string, nextOff int64, ts string) error {

	db := mong.Session.DB(mong.Database)
	c := db.C("subscriptions")

	doc := bson.M{"project_uuid": projectUUID, "name": name}
	change := bson.M{"$set": bson.M{"next_offset": nextOff, "pending_ack": ts}}
	err := c.Update(doc, change)
	if err != nil && err != mgo.ErrNotFound {
		mong.logErrorAndCrash(ctx, "UpdateSubPull", err)
	}

	return err

}

// UpdateSubOffsetAck updates a subscription offset after Ack
func (mong *MongoStore) UpdateSubOffsetAck(ctx context.Context, projectUUID string, name string, offset int64, ts string) error {

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
		mong.logErrorAndCrash(ctx, "UpdateSubOffsetAck", err)
	}

	return nil

}

// UpdateSubOffset updates a subscription offset
func (mong *MongoStore) UpdateSubOffset(ctx context.Context, projectUUID string, name string, offset int64) {

	db := mong.Session.DB(mong.Database)
	c := db.C("subscriptions")

	doc := bson.M{"project_uuid": projectUUID, "name": name}
	change := bson.M{"$set": bson.M{"offset": offset, "next_offset": 0, "pending_ack": ""}}
	err := c.Update(doc, change)
	if err != nil && err != mgo.ErrNotFound {
		mong.logErrorAndCrash(ctx, "UpdateSubOffset", err)
	}

}

// HasUsers accepts a user array of usernames and returns the not found
func (mong *MongoStore) HasUsers(ctx context.Context, projectUUID string, users []string) (bool, []string) {
	db := mong.Session.DB(mong.Database)
	var results []QUser
	var notFound []string
	c := db.C("users")

	err := c.Find(bson.M{"projects": bson.M{"$elemMatch": bson.M{"project_uuid": projectUUID}}, "name": bson.M{"$in": users}}).All(&results)
	if err != nil {
		mong.logErrorAndCrash(ctx, "HasUsers", err)
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
func (mong *MongoStore) QueryACL(ctx context.Context, projectUUID string, resource string, name string) (QAcl, error) {

	db := mong.Session.DB(mong.Database)
	var results []QAcl
	var c *mgo.Collection
	if resource != "topics" && resource != "subscriptions" {
		return QAcl{}, errors.New("wrong resource type")
	}

	c = db.C(resource)

	err := c.Find(bson.M{"project_uuid": projectUUID, "name": name}).All(&results)

	if err != nil {
		mong.logErrorAndCrash(ctx, "QueryACL", err)
	}

	if len(results) > 0 {
		return results[0], nil
	}

	return QAcl{}, errors.New("not found")
}

// QueryUsers queries user(s) information belonging to a project
func (mong *MongoStore) QueryUsers(ctx context.Context, projectUUID string, uuid string, name string) ([]QUser, error) {

	// By default, return all users
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
		mong.logErrorAndCrash(ctx, "QueryUsers", err)
	}

	return results, err
}

// PaginatedQueryUsers returns a page of users
func (mong *MongoStore) PaginatedQueryUsers(ctx context.Context, pageToken string, pageSize int64, projectUUID string) ([]QUser, int64, string, error) {

	var qUsers []QUser
	var totalSize int64
	var limit int64
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
		mong.logErrorAndCrash(ctx, "PaginatedQueryUsers", err)
	}
	totalSize = int64(size)

	// now take into account if pagination is enabled and change the query accordingly
	// first check if an pageToken is provided and whether is a valid bson ID
	if pageToken != "" {
		if ok = bson.IsObjectIdHex(pageToken); !ok {
			err = fmt.Errorf("Page token %v is not a valid bson ObjectId", pageToken)
			log.WithFields(
				log.Fields{
					"type":            "backend_log",
					"trace_id":        ctx.Value("trace_id"),
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
		mong.logErrorAndCrash(ctx, "PaginatedQueryUsers-2", err)
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

// QuerySubsByTopic returns subscriptions of a specific topic
func (mong *MongoStore) QuerySubsByTopic(ctx context.Context, projectUUID, topic string) ([]QSub, error) {
	// By default, return all subs of a given project
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
		mong.logErrorAndCrash(ctx, "QuerySubsByTopic", err)
	}

	return results, err
}

// QuerySubsByACL returns subscriptions that a specific username has access to
func (mong *MongoStore) QuerySubsByACL(ctx context.Context, projectUUID, user string) ([]QSub, error) {
	// By default, return all subs of a given project
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
		mong.logErrorAndCrash(ctx, "QuerySubsByACL", err)
	}

	return results, err
}

// QueryTopicsByACL returns topics that a specific username has access to
func (mong *MongoStore) QueryTopicsByACL(ctx context.Context, projectUUID, user string) ([]QTopic, error) {
	// By default, return all topics of a given project
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
		mong.logErrorAndCrash(ctx, "QueryTopicsByACL", err)
	}

	return results, err
}

// QueryTopics Query Subscription info from store
func (mong *MongoStore) QueryTopics(ctx context.Context, projectUUID, userUUID, name, pageToken string, pageSize int64) ([]QTopic, int64, string, error) {

	var err error
	var totalSize int64
	var limit int64
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
					"trace_id":        ctx.Value("trace_id"),
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
		mong.logErrorAndCrash(ctx, "QueryTopics", err)
	}

	if name == "" {

		countQuery := bson.M{"project_uuid": projectUUID}
		if userUUID != "" {
			countQuery["acl"] = bson.M{"$in": []string{userUUID}}
		}

		if size, err = c.Find(countQuery).Count(); err != nil {
			mong.logErrorAndCrash(ctx, "QueryTopics-2", err)

		}

		totalSize = int64(size)

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
func (mong *MongoStore) UpdateTopicLatestPublish(ctx context.Context, projectUUID string, name string, date time.Time) error {

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
func (mong *MongoStore) UpdateTopicPublishRate(ctx context.Context, projectUUID string, name string, rate float64) error {

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
func (mong *MongoStore) UpdateSubLatestConsume(ctx context.Context, projectUUID string, name string, date time.Time) error {

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
func (mong *MongoStore) UpdateSubConsumeRate(ctx context.Context, projectUUID string, name string, rate float64) error {

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
func (mong *MongoStore) QueryDailyTopicMsgCount(ctx context.Context, projectUUID string, topicName string, date time.Time) ([]QDailyTopicMsgCount, error) {

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
		mong.logErrorAndCrash(ctx, "QueryDailyTopicMsgCount", err)
	}

	return qDailyTopicMsgCount, err
}

// IncrementTopicMsgNum increments the number of messages published in a topic
func (mong *MongoStore) IncrementTopicMsgNum(ctx context.Context, projectUUID string, name string, num int64) error {

	db := mong.Session.DB(mong.Database)
	c := db.C("topics")

	doc := bson.M{"project_uuid": projectUUID, "name": name}
	change := bson.M{"$inc": bson.M{"msg_num": num}}

	err := c.Update(doc, change)

	return err

}

// IncrementDailyTopicMsgCount increments the daily count of published messages to a specific topic
func (mong *MongoStore) IncrementDailyTopicMsgCount(ctx context.Context, projectUUID string, topicName string, num int64, date time.Time) error {

	db := mong.Session.DB(mong.Database)
	c := db.C("daily_topic_msg_count")

	doc := bson.M{"date": date, "project_uuid": projectUUID, "topic_name": topicName}
	change := bson.M{"$inc": bson.M{"msg_count": num}}

	_, err := c.Upsert(doc, change)

	return err

}

// IncrementTopicBytes increases the total number of bytes published in a topic
func (mong *MongoStore) IncrementTopicBytes(ctx context.Context, projectUUID string, name string, totalBytes int64) error {
	db := mong.Session.DB(mong.Database)
	c := db.C("topics")

	doc := bson.M{"project_uuid": projectUUID, "name": name}
	change := bson.M{"$inc": bson.M{"total_bytes": totalBytes}}

	err := c.Update(doc, change)

	return err
}

// IncrementSubMsgNum increments the number of messages pulled in a subscription
func (mong *MongoStore) IncrementSubMsgNum(ctx context.Context, projectUUID string, name string, num int64) error {

	db := mong.Session.DB(mong.Database)
	c := db.C("subscriptions")

	doc := bson.M{"project_uuid": projectUUID, "name": name}
	change := bson.M{"$inc": bson.M{"msg_num": num}}

	err := c.Update(doc, change)

	return err

}

// IncrementSubBytes increases the total number of bytes consumed from a subscription
func (mong *MongoStore) IncrementSubBytes(ctx context.Context, projectUUID string, name string, totalBytes int64) error {
	db := mong.Session.DB(mong.Database)
	c := db.C("subscriptions")

	doc := bson.M{"project_uuid": projectUUID, "name": name}
	change := bson.M{"$inc": bson.M{"total_bytes": totalBytes}}

	err := c.Update(doc, change)

	return err
}

// HasResourceRoles returns the roles of a user in a project
func (mong *MongoStore) HasResourceRoles(ctx context.Context, resource string, roles []string) bool {

	db := mong.Session.DB(mong.Database)
	c := db.C("roles")
	var results []QRole
	err := c.Find(bson.M{"resource": resource, "roles": bson.M{"$in": roles}}).All(&results)
	if err != nil {
		mong.logErrorAndCrash(ctx, "HasResourceRoles", err)
	}

	if len(results) > 0 {
		return true
	}

	return false

}

// GetAllRoles returns a list of all available roles
func (mong *MongoStore) GetAllRoles(ctx context.Context) []string {

	db := mong.Session.DB(mong.Database)
	c := db.C("roles")
	var results []string
	err := c.Find(nil).Distinct("roles", &results)
	if err != nil {
		mong.logErrorAndCrash(ctx, "GetAllRoles", err)
	}
	return results
}

// GetOpMetrics returns the operational metrics from datastore
func (mong *MongoStore) GetOpMetrics(ctx context.Context) []QopMetric {

	db := mong.Session.DB(mong.Database)
	c := db.C("op_metrics")
	var results []QopMetric

	err := c.Find(bson.M{}).All(&results)

	if err != nil {
		mong.logErrorAndCrash(ctx, "GetOpMetrics", err)
	}

	return results

}

// GetUserRoles returns the roles of a user in a project
func (mong *MongoStore) GetUserRoles(ctx context.Context, projectUUID string, token string) ([]string, string) {

	db := mong.Session.DB(mong.Database)
	c := db.C("users")
	var results []QUser

	err := c.Find(bson.M{"token": token}).All(&results)

	if err != nil {
		mong.logErrorAndCrash(ctx, "GetUserRoles", err)
	}

	if len(results) == 0 {
		return []string{}, ""
	}

	if len(results) > 1 {
		log.WithFields(
			log.Fields{
				"type":            "backend_log",
				"trace_id":        ctx.Value("trace_id"),
				"token":           token,
				"backend_service": "mongo",
				"backend_hosts":   mong.Server,
			},
		).Warning("Multiple users with the same token")
	}

	// Search the found user for project roles
	return results[0].getProjectRoles(projectUUID), results[0].Name

}

// GetUserFromToken returns user information from a specific token
func (mong *MongoStore) GetUserFromToken(ctx context.Context, token string) (QUser, error) {

	db := mong.Session.DB(mong.Database)
	c := db.C("users")
	var results []QUser

	err := c.Find(bson.M{"token": token}).All(&results)

	if err != nil {
		mong.logErrorAndCrash(ctx, "GetUserFromToken", err)

	}

	if len(results) == 0 {
		return QUser{}, errors.New("not found")
	}

	if len(results) > 1 {
		log.WithFields(
			log.Fields{
				"type":            "backend_log",
				"trace_id":        ctx.Value("trace_id"),
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
func (mong *MongoStore) QueryOneSub(ctx context.Context, projectUUID string, name string) (QSub, error) {

	db := mong.Session.DB(mong.Database)
	c := db.C("subscriptions")
	var results []QSub
	err := c.Find(bson.M{"project_uuid": projectUUID, "name": name}).All(&results)
	if err != nil {
		mong.logErrorAndCrash(ctx, "QueryOneSub", err)

	}

	if len(results) > 0 {
		return results[0], nil
	}

	return QSub{}, errors.New("empty")
}

// HasProject Returns true if project exists
func (mong *MongoStore) HasProject(ctx context.Context, name string) bool {

	db := mong.Session.DB(mong.Database)
	c := db.C("projects")
	var results []QProject
	err := c.Find(bson.M{"name": name}).All(&results)
	if err != nil {
		mong.logErrorAndCrash(ctx, "HasProject", err)

	}

	if len(results) > 0 {
		return true
	}
	return false
}

// InsertTopic inserts a topic to the store
func (mong *MongoStore) InsertTopic(ctx context.Context, projectUUID string, name string, schemaUUID string, createdOn time.Time) error {

	topic := QTopic{
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

	return mong.InsertResource(ctx, "topics", topic)
}

// LinkTopicSchema links the topic with a schema
func (mong *MongoStore) LinkTopicSchema(ctx context.Context, projectUUID, name, schemaUUID string) error {

	db := mong.Session.DB(mong.Database)
	c := db.C("topics")

	err := c.Update(bson.M{"project_uuid": projectUUID, "name": name}, bson.M{"$set": bson.M{"schema_uuid": schemaUUID}})
	if err != nil {
		mong.logErrorAndCrash(ctx, "LinkTopicSchema", err)

	}
	return nil
}

// InsertOpMetric inserts an operational metric
func (mong *MongoStore) InsertOpMetric(ctx context.Context, hostname string, cpu float64, mem float64) error {
	opMetric := QopMetric{Hostname: hostname, CPU: cpu, MEM: mem}
	db := mong.Session.DB(mong.Database)
	c := db.C("op_metrics")

	upsertdata := bson.M{"$set": opMetric}

	_, err := c.UpsertId(opMetric.Hostname, upsertdata)
	if err != nil {
		mong.logErrorAndCrash(ctx, "InsertOpMetric", err)

	}

	return err
}

// InsertUser inserts a new user to the store
func (mong *MongoStore) InsertUser(ctx context.Context, uuid string, projects []QProjectRoles, name string, fname string, lname string, org string, desc string, token string, email string, serviceRoles []string, createdOn time.Time, modifiedOn time.Time, createdBy string) error {
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
	return mong.InsertResource(ctx, "users", user)
}

// InsertProject inserts a project to the store
func (mong *MongoStore) InsertProject(ctx context.Context, uuid string, name string, createdOn time.Time, modifiedOn time.Time, createdBy string, description string) error {
	project := QProject{UUID: uuid, Name: name, CreatedOn: createdOn, ModifiedOn: modifiedOn, CreatedBy: createdBy, Description: description}
	return mong.InsertResource(ctx, "projects", project)
}

// InsertSub inserts a subscription to the store
func (mong *MongoStore) InsertSub(ctx context.Context, projectUUID string, name string, topic string,
	offset int64, ack int, pushCfg QPushConfig, createdOn time.Time) error {
	sub := QSub{
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
	return mong.InsertResource(ctx, "subscriptions", sub)
}

// RemoveProjectTopics removes all topics related to a project UUID
func (mong *MongoStore) RemoveProjectTopics(ctx context.Context, projectUUID string) error {
	topicMatch := bson.M{"project_uuid": projectUUID}
	return mong.RemoveAll(ctx, "topics", topicMatch)
}

// RemoveProjectSubs removes all subscriptions related to a project UUID
func (mong *MongoStore) RemoveProjectSubs(ctx context.Context, projectUUID string) error {
	subMatch := bson.M{"project_uuid": projectUUID}
	return mong.RemoveAll(ctx, "subscriptions", subMatch)
}

// RemoveProjectDailyMessageCounters removes all message counts related to a project UUID
func (mong *MongoStore) RemoveProjectDailyMessageCounters(ctx context.Context, projectUUID string) error {
	return mong.RemoveAll(ctx, "daily_topic_msg_count", bson.M{"project_uuid": projectUUID})
}

// QueryTotalMessagesPerProject returns the total amount of messages per project for the given time window
func (mong *MongoStore) QueryTotalMessagesPerProject(ctx context.Context, projectUUIDs []string, startDate time.Time, endDate time.Time) ([]QProjectMessageCount, error) {

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
		mong.logErrorAndCrash(ctx, "QueryTotalMessagesPerProject", err)
	}

	return qdp, err
}

// QueryDailyProjectMsgCount queries the total messages per day for a given project
func (mong *MongoStore) QueryDailyProjectMsgCount(ctx context.Context, projectUUID string) ([]QDailyProjectMsgCount, error) {

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

	fmt.Println(query)
	fmt.Println("2")
	if err = c.Pipe(query).All(&qdp); err != nil {
		mong.logErrorAndCrash(ctx, "QueryDailyProjectMsgCount", err)
	}

	return qdp, err

}

// RemoveProject removes a project from the store
func (mong *MongoStore) RemoveProject(ctx context.Context, uuid string) error {
	project := bson.M{"uuid": uuid}
	return mong.RemoveResource(ctx, "projects", project)
}

// RemoveTopic removes a topic from the store
func (mong *MongoStore) RemoveTopic(ctx context.Context, projectUUID string, name string) error {
	topic := bson.M{"project_uuid": projectUUID, "name": name}
	return mong.RemoveResource(ctx, "topics", topic)
}

// RemoveUser removes a user entry from the store
func (mong *MongoStore) RemoveUser(ctx context.Context, uuid string) error {
	user := bson.M{"uuid": uuid}
	return mong.RemoveResource(ctx, "users", user)
}

// RemoveSub removes a subscription from the store
func (mong *MongoStore) RemoveSub(ctx context.Context, projectUUID string, name string) error {
	sub := bson.M{"project_uuid": projectUUID, "name": name}
	return mong.RemoveResource(ctx, "subscriptions", sub)
}

// ExistsInACL checks if a user is part of a topic's or sub's acl
func (mong *MongoStore) ExistsInACL(ctx context.Context, projectUUID string, resource string, resourceName string, userUUID string) error {

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
func (mong *MongoStore) ModACL(ctx context.Context, projectUUID string, resource string, name string, acl []string) error {
	db := mong.Session.DB(mong.Database)

	if resource != "topics" && resource != "subscriptions" {
		return errors.New("wrong resource type")
	}

	c := db.C(resource)

	err := c.Update(bson.M{"project_uuid": projectUUID, "name": name}, bson.M{"$set": bson.M{"acl": acl}})
	return err
}

// AppendToACL adds additional users to an existing ACL
func (mong *MongoStore) AppendToACL(ctx context.Context, projectUUID string, resource string, name string, acl []string) error {

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

// RemoveFromACL removes users for a given ACL
func (mong *MongoStore) RemoveFromACL(ctx context.Context, projectUUID string, resource string, name string, acl []string) error {

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
func (mong *MongoStore) ModAck(ctx context.Context, projectUUID string, name string, ack int) error {
	db := mong.Session.DB(mong.Database)
	c := db.C("subscriptions")
	err := c.Update(bson.M{"project_uuid": projectUUID, "name": name}, bson.M{"$set": bson.M{"ack": ack}})
	return err
}

// ModSubPush modifies the push configuration
func (mong *MongoStore) ModSubPush(ctx context.Context, projectUUID string, name string, pushCfg QPushConfig) error {
	db := mong.Session.DB(mong.Database)
	c := db.C("subscriptions")

	err := c.Update(bson.M{
		"project_uuid": projectUUID,
		"name":         name,
	},
		bson.M{"$set": bson.M{
			"push_type":            pushCfg.Type,
			"push_endpoint":        pushCfg.PushEndpoint,
			"authorization_type":   pushCfg.AuthorizationType,
			"authorization_header": pushCfg.AuthorizationHeader,
			"max_messages":         pushCfg.MaxMessages,
			"retry_policy":         pushCfg.RetPolicy,
			"retry_period":         pushCfg.RetPeriod,
			"verification_hash":    pushCfg.VerificationHash,
			"verified":             pushCfg.Verified,
			"mattermost_url":       pushCfg.MattermostUrl,
			"mattermost_username":  pushCfg.MattermostUsername,
			"mattermost_channel":   pushCfg.MattermostChannel,
			"base_64_decode":       pushCfg.Base64Decode,
		},
		})
	return err
}

// InsertResource inserts a new topic object to the datastore
func (mong *MongoStore) InsertResource(ctx context.Context, col string, res interface{}) error {

	db := mong.Session.DB(mong.Database)
	c := db.C(col)

	err := c.Insert(res)
	if err != nil {
		log.Fatal("STORE", "\t", err.Error())
	}

	return err
}

// RemoveAll removes all occurrences matched with a resource from the store
func (mong *MongoStore) RemoveAll(ctx context.Context, col string, res interface{}) error {
	db := mong.Session.DB(mong.Database)
	c := db.C(col)

	_, err := c.RemoveAll(res)

	return err
}

// RemoveResource removes a resource from the store
func (mong *MongoStore) RemoveResource(ctx context.Context, col string, res interface{}) error {

	db := mong.Session.DB(mong.Database)
	c := db.C(col)

	err := c.Remove(res) // if not found mgo returns error string "not found"

	return err
}

// QuerySubs Query Subscription info from store
func (mong *MongoStore) QuerySubs(ctx context.Context, projectUUID, userUUID, name, pageToken string, pageSize int64) ([]QSub, int64, string, error) {

	var err error
	var totalSize int64
	var limit int64
	var nextPageToken string
	var qSubs []QSub
	var ok bool
	var size int

	// By default, return all subs of a given project
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

	// first check if an pageToken is provided and whether is a valid bson ID
	if pageToken != "" {
		if ok = bson.IsObjectIdHex(pageToken); !ok {
			err = fmt.Errorf("Page token %v is not a valid bson ObjectId", pageToken)
			log.WithFields(
				log.Fields{
					"type":            "backend_log",
					"trace_id":        ctx.Value("trace_id"),
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
		mong.logErrorAndCrash(ctx, "QuerySubs", err)
	}

	if name == "" {

		countQuery := bson.M{"project_uuid": projectUUID}
		if userUUID != "" {
			countQuery["acl"] = bson.M{"$in": []string{userUUID}}
		}

		if size, err = c.Find(countQuery).Count(); err != nil {
			mong.logErrorAndCrash(ctx, "QuerySubs-2", err)

		}

		totalSize = int64(size)

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
func (mong *MongoStore) QueryPushSubs(ctx context.Context) []QSub {

	db := mong.Session.DB(mong.Database)
	c := db.C("subscriptions")
	var results []QSub
	err := c.Find(bson.M{"push_endpoint": bson.M{"$ne": ""}}).All(&results)
	if err != nil {
		mong.logErrorAndCrash(ctx, "QueryPushSubs", err)

	}
	return results

}

func (mong *MongoStore) InsertSchema(ctx context.Context, projectUUID, schemaUUID, name, schemaType, rawSchemaString string) error {
	sub := QSchema{
		ProjectUUID: projectUUID,
		UUID:        schemaUUID,
		Name:        name,
		Type:        schemaType,
		RawSchema:   rawSchemaString,
	}
	return mong.InsertResource(ctx, "schemas", sub)
}

func (mong *MongoStore) QuerySchemas(ctx context.Context, projectUUID, schemaUUID, name string) ([]QSchema, error) {

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
		mong.logErrorAndCrash(ctx, "QuerySchemas", err)

	}

	return results, nil
}

// UpdateSchema updates the fields of a schema
func (mong *MongoStore) UpdateSchema(ctx context.Context, schemaUUID, name, schemaType, rawSchemaString string) error {

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
func (mong *MongoStore) DeleteSchema(ctx context.Context, schemaUUID string) error {

	db := mong.Session.DB(mong.Database)
	c := db.C("schemas")

	selector := bson.M{"uuid": schemaUUID}

	err := c.Remove(selector)

	if err != nil {
		mong.logErrorAndCrash(ctx, "DeleteSchema", err)

	}

	topics := db.C("topics")

	topicSelector := bson.M{"schema_uuid": schemaUUID}
	change := bson.M{
		"$set": bson.M{
			"schema_uuid": "",
		},
	}
	_, err = topics.UpdateAll(topicSelector, change)
	if err != nil {
		mong.logErrorAndCrash(ctx, "DeleteSchema-2", err)

	}

	return nil

}
