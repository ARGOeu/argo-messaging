package stores

import (
	"context"
	"errors"
	"fmt"
	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"time"
)

const TopicsCollection string = "topics"
const SubscriptionsCollection string = "subscriptions"
const DailyTopicMsgCountCollection string = "daily_topic_msg_count"
const UsersCollection string = "users"
const ProjectsCollection string = "projects"
const UserRegistrationsCollection string = "user_registrations"
const SchemasCollection string = "schemas"
const OpMetricsCollection string = "op_metrics"
const RolesCollection string = "roles"

type DocNotFound struct{}

func (DocNotFound) Error() string {
	return "not found"
}

type findQueryProcessor[T any] struct {
	collection *mongo.Collection
}

func (p findQueryProcessor[T]) execute(ctx context.Context, query bson.M, opts ...*options.FindOptions) ([]T, error) {
	var results []T

	cursor, err := p.collection.Find(ctx, query, opts...)
	if err != nil {
		return results, err
	}

	for cursor.Next(ctx) {
		var result T
		err := cursor.Decode(&result)
		if err != nil {
			return results, err
		}
		results = append(results, result)
	}

	if err := cursor.Err(); err != nil {
		return results, err
	}

	return results, nil

}

type MongoStoreWithOfficialDriver struct {
	Server   string
	Database string
	database *mongo.Database
	client   *mongo.Client

	topicsCollection              *mongo.Collection
	topicsDailyMsgCountCollection *mongo.Collection
	subscriptionsCollection       *mongo.Collection
	usersCollection               *mongo.Collection
	projectsCollection            *mongo.Collection
	userRegistrationsCollection   *mongo.Collection
	schemasCollection             *mongo.Collection
	rolesCollection               *mongo.Collection
	opMetricsCollection           *mongo.Collection

	topicsFindQueryProcessor            findQueryProcessor[QTopic]
	subsFindQueryProcessor              findQueryProcessor[QSub]
	usersFindQueryProcessor             findQueryProcessor[QUser]
	projectsFindQueryProcessor          findQueryProcessor[QProject]
	userRegistrationsFindQueryProcessor findQueryProcessor[QUserRegistration]
	schemasFindQueryProcessor           findQueryProcessor[QSchema]
}

func NewMongoStoreWithOfficialDriver(server, database string) *MongoStoreWithOfficialDriver {
	return &MongoStoreWithOfficialDriver{
		Server:   server,
		Database: database,
	}
}

func (store *MongoStoreWithOfficialDriver) Initialize() {

	mongoDBUri := fmt.Sprintf("mongodb://%s", store.Server)

	for {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		log.WithFields(
			log.Fields{
				"type":            "backend_log",
				"backend_service": "mongo",
				"backend_hosts":   store.Server,
			},
		).Info("Trying to connect to Mongo")
		client, err := mongo.Connect(ctx, options.Client().ApplyURI(mongoDBUri))
		if err != nil {
			log.WithFields(
				log.Fields{
					"type":            "backend_log",
					"backend_service": "mongo",
					"backend_hosts":   store.Server,
				},
			).Error(err.Error())
			continue
		}
		store.client = client
		cancel()
		break
	}

	log.WithFields(
		log.Fields{
			"type":            "backend_log",
			"backend_service": "mongo",
			"backend_hosts":   store.Server,
		},
	).Info("Connection to Mongo established successfully")

	for {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		log.WithFields(
			log.Fields{
				"type":            "backend_log",
				"backend_service": "mongo",
				"backend_hosts":   store.Server,
			},
		).Info("Trying to ping Mongo")
		err := store.client.Ping(ctx, readpref.Primary())
		if err != nil {
			log.WithFields(
				log.Fields{
					"type":            "backend_log",
					"backend_service": "mongo",
					"backend_hosts":   store.Server,
				},
			).Error(err.Error())
			continue
		}
		cancel()
		break
	}

	log.WithFields(
		log.Fields{
			"type":            "backend_log",
			"backend_service": "mongo",
			"backend_hosts":   store.Server,
		},
	).Info("Mongo Deployment is up and running")
	store.database = store.client.Database(store.Database)

	store.topicsCollection = store.database.Collection(TopicsCollection)
	store.topicsFindQueryProcessor = findQueryProcessor[QTopic]{
		collection: store.topicsCollection,
	}

	store.subscriptionsCollection = store.database.Collection(SubscriptionsCollection)
	store.subsFindQueryProcessor = findQueryProcessor[QSub]{
		collection: store.subscriptionsCollection,
	}

	store.usersCollection = store.database.Collection(UsersCollection)
	store.usersFindQueryProcessor = findQueryProcessor[QUser]{
		collection: store.usersCollection,
	}

	store.projectsCollection = store.database.Collection(ProjectsCollection)
	store.projectsFindQueryProcessor = findQueryProcessor[QProject]{
		collection: store.projectsCollection,
	}

	store.userRegistrationsCollection = store.database.Collection(UserRegistrationsCollection)
	store.userRegistrationsFindQueryProcessor = findQueryProcessor[QUserRegistration]{
		collection: store.userRegistrationsCollection,
	}

	store.schemasCollection = store.database.Collection(SchemasCollection)
	store.schemasFindQueryProcessor = findQueryProcessor[QSchema]{
		collection: store.schemasCollection,
	}

	store.topicsDailyMsgCountCollection = store.database.Collection(DailyTopicMsgCountCollection)
	store.rolesCollection = store.database.Collection(RolesCollection)
	store.opMetricsCollection = store.database.Collection(OpMetricsCollection)
}

func (store *MongoStoreWithOfficialDriver) Close() {
	if store.client != nil {
		if err := store.client.Disconnect(context.Background()); err != nil {
			log.Fatalf("Could not disconnect mongo client, %s", err.Error())
		}
	}
}

func (store *MongoStoreWithOfficialDriver) Clone() Store {
	return store
}

func (store *MongoStoreWithOfficialDriver) logErrorAndCrash(ctx context.Context, funcName string, err error) {
	log.WithFields(
		log.Fields{
			"trace_id":        ctx.Value("trace_id"),
			"type":            "backend_log",
			"function":        funcName,
			"backend_service": "mongo",
			"backend_hosts":   store.Server,
		},
	).Fatal(err.Error())
}

func (store *MongoStoreWithOfficialDriver) deleteOne(
	ctx context.Context,
	collection *mongo.Collection,
	query bson.M) error {
	dr, err := collection.DeleteOne(ctx, query)
	if err != nil {
		return err
	}
	if dr.DeletedCount == 0 {
		return DocNotFound{}
	}
	return nil
}

func (store *MongoStoreWithOfficialDriver) upsert(ctx context.Context, doc interface{},
	change interface{}, collection *mongo.Collection) error {
	updateOptions := options.Update().SetUpsert(true)
	_, err := collection.UpdateOne(ctx, doc, change, updateOptions)
	return err
}

// getDocCountForCollectionPerProject returns the document count for a collection in a given time period
// collection should support field created_on. The results are projected on a per-project counter.
func (store *MongoStoreWithOfficialDriver) getDocCountForCollectionPerProject(ctx context.Context, startDate,
	endDate time.Time, projectUUIDs []string, collection *mongo.Collection) (map[string]int64, error) {

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

	res := map[string]int64{}

	cursor, err := collection.Aggregate(ctx, query)
	if err != nil {
		return nil, err
	}

	err = cursor.All(ctx, &resourceCounts)
	if err != nil {
		return nil, err
	}

	for _, t := range resourceCounts {
		res[t.ProjectUUID] = t.Count
	}

	return res, nil
}

// ##### OP METRICS QUERIES #####

// InsertOpMetric inserts an operational metric
func (store *MongoStoreWithOfficialDriver) InsertOpMetric(ctx context.Context, hostname string, cpu float64, mem float64) error {
	doc := bson.M{"hostname": hostname, "cpu": cpu, "mem": mem}
	change := bson.M{"$set": bson.M{"hostname": hostname, "cpu": cpu, "mem": mem}}
	err := store.upsert(ctx, doc, change, store.opMetricsCollection)
	if err != nil {
		store.logErrorAndCrash(ctx, "InsertOpMetric", err)
		return err
	}
	return nil
}

// GetOpMetrics returns the operational metrics from datastore
func (store *MongoStoreWithOfficialDriver) GetOpMetrics(ctx context.Context) []QopMetric {
	var results []QopMetric
	cursor, err := store.opMetricsCollection.Find(ctx, bson.M{})
	if err != nil {
		store.logErrorAndCrash(ctx, "GetOpMetrics", err)
		return results
	}
	err = cursor.All(ctx, &results)
	if err != nil {
		store.logErrorAndCrash(ctx, "GetOpMetrics", err)
		return results
	}
	return results
}

// ##### ROLES QUERIES #####

func (store *MongoStoreWithOfficialDriver) HasResourceRoles(ctx context.Context, resource string, roles []string) bool {
	query := bson.M{
		"resource": resource,
		"roles": bson.M{
			"$in": roles,
		},
	}
	var results []QRole
	cursor, err := store.rolesCollection.Find(ctx, query)
	if err != nil {
		store.logErrorAndCrash(ctx, "HasResourceRoles", err)
		return false
	}
	err = cursor.All(ctx, &results)
	if err != nil {
		store.logErrorAndCrash(ctx, "HasResourceRoles", err)
		return false
	}
	return len(results) > 0
}

func (store *MongoStoreWithOfficialDriver) InsertResourceRoles(ctx context.Context, resource string, roles []string) error {
	role := QRole{
		Name:  resource,
		Roles: roles,
	}
	_, err := store.rolesCollection.InsertOne(ctx, role)
	if err != nil {
		store.logErrorAndCrash(ctx, "InsertResourceRoles", err)
		return err
	}
	return nil
}

// GetAllRoles returns a list of all available roles
func (store *MongoStoreWithOfficialDriver) GetAllRoles(ctx context.Context) []string {
	cursor, err := store.rolesCollection.Distinct(ctx, "roles", bson.M{})
	if err != nil {
		store.logErrorAndCrash(ctx, "GetAllRoles", err)
		return []string{}
	}
	var roles []string
	for _, i := range cursor {
		roles = append(roles, i.(string))
	}
	return roles

}

// ##### ACL QUERIES ######

// QueryACL queries topic or subscription for a list of authorized users
func (store *MongoStoreWithOfficialDriver) QueryACL(ctx context.Context, projectUUID string, resource string, name string) (QAcl, error) {

	var results []QAcl
	var c *mongo.Collection

	if resource == "topics" {
		c = store.topicsCollection
	} else if resource == "subscriptions" {
		c = store.subscriptionsCollection
	} else {
		return QAcl{}, errors.New("wrong resource type")
	}

	query := bson.M{"project_uuid": projectUUID, "name": name}

	cursor, err := c.Find(ctx, query)
	if err != nil {
		store.logErrorAndCrash(ctx, "QueryACL", err)
		return QAcl{}, err
	}
	err = cursor.All(ctx, &results)
	if err != nil {
		store.logErrorAndCrash(ctx, "QueryACL", err)
		return QAcl{}, err
	}
	if len(results) > 0 {
		return results[0], nil
	}

	return QAcl{}, DocNotFound{}
}

// ExistsInACL checks if a user is part of a topic's or sub's acl
func (store *MongoStoreWithOfficialDriver) ExistsInACL(ctx context.Context, projectUUID string, resource string, resourceName string, userUUID string) error {

	var c *mongo.Collection
	if resource == "topics" {
		c = store.topicsCollection
	} else if resource == "subscriptions" {
		c = store.subscriptionsCollection
	} else {
		return errors.New("wrong resource type")
	}

	query := bson.M{
		"project_uuid": projectUUID,
		"name":         resourceName,
		"acl": bson.M{
			"$in": []string{userUUID},
		},
	}

	err := c.FindOne(ctx, query).Err()
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return DocNotFound{}
		} else {
			store.logErrorAndCrash(ctx, "ExistsInAcl", err)
			return err
		}
	}
	return nil
}

// ModACL modifies the push configuration
func (store *MongoStoreWithOfficialDriver) ModACL(ctx context.Context, projectUUID string,
	resource string, name string, acl []string) error {
	var c *mongo.Collection
	if resource == "topics" {
		c = store.topicsCollection
	} else if resource == "subscriptions" {
		c = store.subscriptionsCollection
	} else {
		return errors.New("wrong resource type")
	}
	doc := bson.M{"project_uuid": projectUUID, "name": name}
	change := bson.M{"$set": bson.M{"acl": acl}}
	_, err := c.UpdateOne(ctx, doc, change)
	if err != nil {
		store.logErrorAndCrash(ctx, "ModACL", err)
		return err
	}

	return nil
}

// AppendToACL adds additional users to an existing ACL
func (store *MongoStoreWithOfficialDriver) AppendToACL(ctx context.Context, projectUUID string, resource string, name string, acl []string) error {

	var c *mongo.Collection
	if resource == "topics" {
		c = store.topicsCollection
	} else if resource == "subscriptions" {
		c = store.subscriptionsCollection
	} else {
		return errors.New("wrong resource type")
	}
	doc := bson.M{
		"project_uuid": projectUUID,
		"name":         name,
	}
	change := bson.M{
		"$addToSet": bson.M{
			"acl": bson.M{
				"$each": acl,
			},
		}}
	_, err := c.UpdateOne(ctx, doc, change)
	if err != nil {
		store.logErrorAndCrash(ctx, "AppendToACL", err)
		return err
	}

	return nil
}

// RemoveFromACL removes users for a given ACL
func (store *MongoStoreWithOfficialDriver) RemoveFromACL(ctx context.Context, projectUUID string, resource string, name string, acl []string) error {
	var c *mongo.Collection
	if resource == "topics" {
		c = store.topicsCollection
	} else if resource == "subscriptions" {
		c = store.subscriptionsCollection
	} else {
		return errors.New("wrong resource type")
	}

	doc := bson.M{
		"project_uuid": projectUUID,
		"name":         name,
	}
	change := bson.M{
		"$pullAll": bson.M{
			"acl": acl,
		},
	}
	_, err := c.UpdateOne(ctx, doc, change)
	if err != nil {
		store.logErrorAndCrash(ctx, "RemoveFromACL", err)
		return err
	}
	return nil
}

// ##### SCHEMA QUERIES #####

func (store *MongoStoreWithOfficialDriver) InsertSchema(ctx context.Context, projectUUID, schemaUUID, name,
	schemaType, rawSchemaString string) error {
	schema := QSchema{
		ProjectUUID: projectUUID,
		UUID:        schemaUUID,
		Name:        name,
		Type:        schemaType,
		RawSchema:   rawSchemaString,
	}
	_, err := store.schemasCollection.InsertOne(ctx, schema)
	if err != nil {
		store.logErrorAndCrash(ctx, "InsertSchema", err)
		return err
	}
	return nil
}

func (store *MongoStoreWithOfficialDriver) QuerySchemas(ctx context.Context, projectUUID, schemaUUID, name string) ([]QSchema, error) {

	query := bson.M{"project_uuid": projectUUID}

	if name != "" {
		query["name"] = name
	}

	if schemaUUID != "" {
		query["uuid"] = schemaUUID
	}

	results, err := store.schemasFindQueryProcessor.execute(ctx, query)
	if err != nil {
		store.logErrorAndCrash(ctx, "QuerySchemas", err)
		return nil, err
	}

	return results, nil

}

func (store *MongoStoreWithOfficialDriver) UpdateSchema(ctx context.Context, schemaUUID, name, schemaType, rawSchemaString string) error {

	doc := bson.M{"uuid": schemaUUID}

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
	_, err := store.schemasCollection.UpdateOne(ctx, doc, change)
	if err != nil {
		store.logErrorAndCrash(ctx, "QuerySchemas", err)
		return err
	}
	return nil
}

// DeleteSchema removes the schema from the store
// It also clears all the respective topics from the schema_uuid of the deleted schema
func (store *MongoStoreWithOfficialDriver) DeleteSchema(ctx context.Context, schemaUUID string) error {
	schemaQuery := bson.M{"uuid": schemaUUID}
	_, err := store.schemasCollection.DeleteOne(ctx, schemaQuery)
	if err != nil {
		store.logErrorAndCrash(ctx, "DeleteSchema", err)
		return err
	}
	doc := bson.M{"schema_uuid": schemaUUID}
	change := bson.M{
		"$set": bson.M{
			"schema_uuid": "",
		},
	}
	_, err = store.topicsCollection.UpdateMany(ctx, doc, change)
	if err != nil {
		store.logErrorAndCrash(ctx, "DeleteSchema-2", err)
		return err
	}

	return nil
}

// ##### PROJECT QUERIES #####

// QueryProjects queries the database for a specific project or a list of all projects
func (store *MongoStoreWithOfficialDriver) QueryProjects(ctx context.Context, uuid string, name string) ([]QProject, error) {

	query := bson.M{}

	if name != "" {

		query = bson.M{"name": name}

	} else if uuid != "" {
		query = bson.M{"uuid": uuid}
	}

	results, err := store.projectsFindQueryProcessor.execute(ctx, query)
	if err != nil {
		store.logErrorAndCrash(ctx, "QueryProjects", err)
		return nil, err
	}

	if len(results) > 0 {
		return results, nil
	}

	return results, errors.New("not found")
}

// UpdateProject updates project information
func (store *MongoStoreWithOfficialDriver) UpdateProject(ctx context.Context, projectUUID string, name string,
	description string, modifiedOn time.Time) error {

	doc := bson.M{"uuid": projectUUID}
	results, err := store.QueryProjects(ctx, projectUUID, "")
	if err != nil {
		store.logErrorAndCrash(ctx, "UpdateProject", err)
		return err
	}

	curPr := results[0]
	curPr.ModifiedOn = modifiedOn // modifiedOn should always be updated

	if name != "" {
		// Check if name is going to change and if that name already exists
		if name != curPr.Name {
			if sameRes, _ := store.QueryProjects(ctx, "", name); len(sameRes) > 0 {
				return errors.New("invalid project name change, name already exists")
			}
		}
		curPr.Name = name
	}

	if description != "" {
		curPr.Description = description
	}

	change := bson.M{"$set": curPr}
	_, err = store.projectsCollection.UpdateOne(ctx, doc, change)
	if err != nil {
		store.logErrorAndCrash(ctx, "UpdateProject", err)
		return err
	}
	return nil
}

// RemoveProject removes a project from the store
func (store *MongoStoreWithOfficialDriver) RemoveProject(ctx context.Context, uuid string) error {
	project := bson.M{"uuid": uuid}
	_, err := store.projectsCollection.DeleteOne(ctx, project)
	if err != nil {
		store.logErrorAndCrash(ctx, "RemoveProject", err)
		return err
	}
	return nil
}

// RemoveProjectTopics removes all topics related to a project UUID
func (store *MongoStoreWithOfficialDriver) RemoveProjectTopics(ctx context.Context, projectUUID string) error {
	topics := bson.M{"project_uuid": projectUUID}
	_, err := store.topicsCollection.DeleteMany(ctx, topics)
	if err != nil {
		store.logErrorAndCrash(ctx, "RemoveProjectTopics", err)
		return err
	}
	return nil
}

// RemoveProjectSubs removes all subscriptions related to a project UUID
func (store *MongoStoreWithOfficialDriver) RemoveProjectSubs(ctx context.Context, projectUUID string) error {
	subs := bson.M{"project_uuid": projectUUID}
	_, err := store.subscriptionsCollection.DeleteMany(ctx, subs)
	if err != nil {
		store.logErrorAndCrash(ctx, "RemoveProjectSubs", err)
		return err
	}
	return nil
}

// RemoveProjectDailyMessageCounters removes all message counts related to a project UUID
func (store *MongoStoreWithOfficialDriver) RemoveProjectDailyMessageCounters(ctx context.Context, projectUUID string) error {
	counts := bson.M{"project_uuid": projectUUID}
	_, err := store.topicsDailyMsgCountCollection.DeleteMany(ctx, counts)
	if err != nil {
		store.logErrorAndCrash(ctx, "RemoveProjectDailyMessageCounters", err)
		return err
	}
	return nil
}

// QueryDailyProjectMsgCount queries the total messages per day for a given project
func (store *MongoStoreWithOfficialDriver) QueryDailyProjectMsgCount(ctx context.Context, projectUUID string) ([]QDailyProjectMsgCount, error) {

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
	var res []QDailyProjectMsgCount
	cursor, err := store.topicsDailyMsgCountCollection.Aggregate(ctx, query)
	if err != nil {
		return res, err
	}

	err = cursor.All(ctx, &res)
	if err != nil {
		store.logErrorAndCrash(ctx, "QueryDailyProjectMsgCount", err)
		return res, err
	}
	return res, nil
}

func (store *MongoStoreWithOfficialDriver) QueryTotalMessagesPerProject(ctx context.Context, projectUUIDs []string, startDate time.Time, endDate time.Time) ([]QProjectMessageCount, error) {

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

	var res []QProjectMessageCount
	cursor, err := store.topicsDailyMsgCountCollection.Aggregate(ctx, query)
	if err != nil {
		return res, err
	}

	err = cursor.All(ctx, &res)
	if err != nil {
		store.logErrorAndCrash(ctx, "QueryTotalMessagesPerProject", err)
		return res, err
	}
	return res, nil
}

func (store *MongoStoreWithOfficialDriver) InsertProject(ctx context.Context, uuid string, name string,
	createdOn time.Time, modifiedOn time.Time, createdBy string, description string) error {
	project := QProject{
		UUID:        uuid,
		Name:        name,
		CreatedOn:   createdOn,
		ModifiedOn:  modifiedOn,
		CreatedBy:   createdBy,
		Description: description,
	}
	_, err := store.projectsCollection.InsertOne(ctx, project)
	if err != nil {
		store.logErrorAndCrash(ctx, "InsertProject", err)
		return err
	}
	return nil
}

func (store *MongoStoreWithOfficialDriver) HasProject(ctx context.Context, name string) bool {
	query := bson.M{"name": name}
	results, err := store.projectsFindQueryProcessor.execute(ctx, query)
	if err != nil {
		store.logErrorAndCrash(ctx, "HasProject", err)
		return false
	}

	if len(results) > 0 {
		return true
	}

	return false
}

// ##### USER REGISTRATIONS QUERIES #####

// RegisterUser inserts a new user registration to the database
func (store *MongoStoreWithOfficialDriver) RegisterUser(ctx context.Context, uuid, name, firstName, lastName, email,
	org, desc, registeredAt, atkn, status string) error {
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
	_, err := store.userRegistrationsCollection.InsertOne(ctx, ur)
	if err != nil {
		store.logErrorAndCrash(ctx, "RegisterUSer", err)
		return err
	}
	return nil
}

// DeleteRegistration removes the respective registration from the
func (store *MongoStoreWithOfficialDriver) DeleteRegistration(ctx context.Context, uuid string) error {
	query := bson.M{"uuid": uuid}
	err := store.deleteOne(ctx, store.userRegistrationsCollection, query)
	if err != nil {
		if (err == DocNotFound{}) {
			return err
		} else {
			store.logErrorAndCrash(ctx, "RemoveSub", err)
			return err
		}
	}
	return nil
}

func (store *MongoStoreWithOfficialDriver) QueryRegistrations(ctx context.Context, regUUID, status, activationToken, name, email, org string) ([]QUserRegistration, error) {
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

	res, err := store.userRegistrationsFindQueryProcessor.execute(ctx, query)
	if err != nil {
		store.logErrorAndCrash(ctx, "QueryRegistrations", err)
		return nil, err
	}

	return res, nil
}

func (store *MongoStoreWithOfficialDriver) UpdateRegistration(ctx context.Context, regUUID, status, declineComment, modifiedBy, modifiedAt string) error {

	doc := bson.M{"uuid": regUUID}
	change := bson.M{
		"$set": bson.M{
			"status":           status,
			"decline_comment":  declineComment,
			"modified_by":      modifiedBy,
			"modified_at":      modifiedAt,
			"activation_token": "",
		},
	}
	_, err := store.userRegistrationsCollection.UpdateOne(ctx, doc, change)
	if err != nil {
		store.logErrorAndCrash(ctx, "UpdateRegistrations", err)
		return err
	}

	return nil
}

// ###### USER QUERIES ######

// HasUsers accepts a user array of usernames and returns the not found
func (store *MongoStoreWithOfficialDriver) HasUsers(ctx context.Context, projectUUID string, users []string) (bool, []string) {
	var results []QUser
	var notFound []string

	query := bson.M{
		"projects": bson.M{
			"$elemMatch": bson.M{
				"project_uuid": projectUUID,
			},
		},
		"name": bson.M{"$in": users},
	}

	results, err := store.usersFindQueryProcessor.execute(ctx, query)
	if err != nil {
		store.logErrorAndCrash(ctx, "HasUsers", err)
		return false, []string{err.Error()}
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

// PaginatedQueryUsers returns a page of users
func (store *MongoStoreWithOfficialDriver) PaginatedQueryUsers(ctx context.Context, pageToken string, pageSize int64,
	projectUUID string) ([]QUser, int64, string, error) {

	var qUsers []QUser
	var totalSize int64
	var limit int64
	var nextPageToken string
	var err error
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

	// check the total of the users selected by the query not taking into account pagination
	totalSize, err = store.usersCollection.CountDocuments(ctx, query)
	if err != nil {
		store.logErrorAndCrash(ctx, "PaginatedQueryUsers", err)
	}

	// now take into account if pagination is enabled and change the query accordingly
	// first check if an pageToken is provided and whether is a valid bson ID
	if pageToken != "" {
		bsonID, err := primitive.ObjectIDFromHex(pageToken)
		if err != nil {
			err = fmt.Errorf("page token %s is not a valid bson ObjectId. %s", pageToken, err.Error())
			log.WithFields(
				log.Fields{
					"type":            "backend_log",
					"trace_id":        ctx.Value("trace_id"),
					"backend_service": "mongo",
					"page_token":      pageToken,
					"err":             err.Error(),
				},
			).Error("Page token is not a valid bson ObjectId")
			return qUsers, totalSize, nextPageToken, err
		}

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
	findOptions := options.Find().SetLimit(limit).SetSort(bson.M{"_id": -1})
	qUsers, err = store.usersFindQueryProcessor.execute(ctx, query, findOptions)
	if err != nil {
		store.logErrorAndCrash(ctx, "PaginatedQueryUsers-2", err)
	}

	// if the amount of users that were found was equal to the limit, its a sign that there are users to populate the next page
	// so pick the last element's pageToken to use as the starting point for the next page
	// and eliminate the extra element from the current response
	if pageSize > 0 && len(qUsers) > 0 && len(qUsers) == int(limit) {

		nextPageToken = qUsers[limit-1].ID.(primitive.ObjectID).Hex()
		qUsers = qUsers[:len(qUsers)-1]
	}

	return qUsers, totalSize, nextPageToken, err

}

// QueryUsers queries user(s) information belonging to a project
func (store *MongoStoreWithOfficialDriver) QueryUsers(ctx context.Context, projectUUID string,
	uuid string, name string) ([]QUser, error) {

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

	results, err := store.usersFindQueryProcessor.execute(ctx, query)
	if err != nil {
		store.logErrorAndCrash(ctx, "QueryUsers", err)
	}

	return results, err
}

// UpdateUser updates user information
func (store *MongoStoreWithOfficialDriver) UpdateUser(ctx context.Context, uuid, fname, lname, org,
	desc string, projects []QProjectRoles, name string, email string, serviceRoles []string, modifiedOn time.Time) error {

	doc := bson.M{"uuid": uuid}
	results, err := store.QueryUsers(ctx, "", uuid, "")
	if err != nil {
		return err
	}

	curUsr := results[0]

	if name != "" {
		// Check if name is going to change and if that name already exists
		if name != curUsr.Name {
			if sameRes, _ := store.QueryUsers(ctx, "", "", name); len(sameRes) > 0 {
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

	_, err = store.usersCollection.UpdateOne(ctx, doc, change)
	if err != nil {
		store.logErrorAndCrash(ctx, "UpdateUser", err)
	}
	return err

}

// AppendToUserProjects appends a new unique project to the user's projects
func (store *MongoStoreWithOfficialDriver) AppendToUserProjects(ctx context.Context, userUUID string, projectUUID string, pRoles ...string) error {
	doc := bson.M{"uuid": userUUID}
	change := bson.M{
		"$addToSet": bson.M{
			"projects": QProjectRoles{
				ProjectUUID: projectUUID,
				Roles:       pRoles,
			},
		},
	}
	_, err := store.usersCollection.UpdateOne(ctx, doc, change)
	if err != nil {
		store.logErrorAndCrash(ctx, "AppendToUserProjects", err)
	}
	return err
}

// UpdateUserToken updates user's token
func (store *MongoStoreWithOfficialDriver) UpdateUserToken(ctx context.Context, uuid string, token string) error {
	doc := bson.M{"uuid": uuid}
	change := bson.M{"$set": bson.M{"token": token}}
	_, err := store.usersCollection.UpdateOne(ctx, doc, change)
	if err != nil {
		store.logErrorAndCrash(ctx, "UpdateUserToken", err)
	}
	return err
}

func (store *MongoStoreWithOfficialDriver) RemoveUser(ctx context.Context, uuid string) error {
	user := bson.M{"uuid": uuid}
	_, err := store.usersCollection.DeleteOne(ctx, user)
	if err != nil {
		store.logErrorAndCrash(ctx, "RemoveUser", err)
	}
	return err
}

// InsertUser inserts a new user to the store
func (store *MongoStoreWithOfficialDriver) InsertUser(ctx context.Context, uuid string, projects []QProjectRoles,
	name string, firstName string, lastName string, org string, desc string, token string, email string, serviceRoles []string, createdOn time.Time, modifiedOn time.Time, createdBy string) error {
	user := QUser{
		UUID:         uuid,
		Name:         name,
		Email:        email,
		Token:        token,
		FirstName:    firstName,
		LastName:     lastName,
		Organization: org,
		Description:  desc,
		Projects:     projects,
		ServiceRoles: serviceRoles,
		CreatedOn:    createdOn,
		ModifiedOn:   modifiedOn,
		CreatedBy:    createdBy,
	}
	_, err := store.usersCollection.InsertOne(ctx, user)
	if err != nil {
		store.logErrorAndCrash(ctx, "InsertUser", err)
	}
	return err
}

// GetUserFromToken returns user information from a specific token
func (store *MongoStoreWithOfficialDriver) GetUserFromToken(ctx context.Context, token string) (QUser, error) {

	query := bson.M{"token": token}
	results, err := store.usersFindQueryProcessor.execute(ctx, query)

	if err != nil {
		store.logErrorAndCrash(ctx, "GetUserFromToken", err)
		return QUser{}, err
	}

	if len(results) == 0 {
		return QUser{}, DocNotFound{}
	}

	if len(results) > 1 {
		log.WithFields(
			log.Fields{
				"type":            "backend_log",
				"trace_id":        ctx.Value("trace_id"),
				"token":           token,
				"backend_service": "mongo",
				"backend_hosts":   store.Server,
			},
		).Warning("Multiple users with the same token")
	}

	// Search the found user for project roles
	return results[0], err
}

// UsersCount returns the amount of users created in the given time period per project
func (store *MongoStoreWithOfficialDriver) UsersCount(ctx context.Context, startDate, endDate time.Time, projectUUIDs []string) (map[string]int64, error) {

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
	res := map[string]int64{}

	cursor, err := store.usersCollection.Aggregate(ctx, query)
	if err != nil {
		return res, err
	}

	err = cursor.All(ctx, &resourceCounts)
	if err != nil {
		store.logErrorAndCrash(ctx, "UsersCount", err)
		return res, err
	}

	for _, t := range resourceCounts {
		res[t.ProjectUUID] = t.Count
	}

	return res, nil

}

func (store *MongoStoreWithOfficialDriver) GetUserRoles(ctx context.Context, projectUUID string, token string) ([]string, string) {

	query := bson.M{"token": token}
	results, err := store.usersFindQueryProcessor.execute(ctx, query)

	if err != nil {
		store.logErrorAndCrash(ctx, "GetUserRoles", err)
		return []string{}, err.Error()
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
				"backend_hosts":   store.Server,
			},
		).Warning("Multiple users with the same token")
	}

	// Search the found user for project roles
	return results[0].getProjectRoles(projectUUID), results[0].Name

}

// ##### SUBSCRIPTION QUERIES #####

// QuerySubsByTopic returns subscriptions of a specific topic
func (store *MongoStoreWithOfficialDriver) QuerySubsByTopic(ctx context.Context, projectUUID, topic string) ([]QSub, error) {

	// By default, return all subs of a given project
	query := bson.M{"project_uuid": projectUUID}

	// If topic is given return only the specific topic
	if topic != "" {
		query = bson.M{"project_uuid": projectUUID, "topic": topic}
	}

	results, err := store.subsFindQueryProcessor.execute(ctx, query)
	if err != nil {
		store.logErrorAndCrash(ctx, "QuerySubsByTopic", err)
	}

	return results, err
}

// QuerySubsByACL returns subscriptions that a specific username has access to
func (store *MongoStoreWithOfficialDriver) QuerySubsByACL(ctx context.Context, projectUUID, user string) ([]QSub, error) {
	// By default, return all subs of a given project
	query := bson.M{"project_uuid": projectUUID}

	// If name is given return only the specific topic
	if user != "" {
		query = bson.M{"project_uuid": projectUUID, "acl": user}
	}
	results, err := store.subsFindQueryProcessor.execute(ctx, query)
	if err != nil {
		store.logErrorAndCrash(ctx, "QuerySubsByACL", err)
	}
	return results, err
}

// QuerySubs Query Subscription info from store
func (store *MongoStoreWithOfficialDriver) QuerySubs(ctx context.Context, projectUUID string, userUUID string, name string, pageToken string, pageSize int64) ([]QSub, int64, string, error) {
	var err error
	var totalSize int64
	var limit int64
	var nextPageToken string
	var qSubs []QSub

	// By default, return all topics of a given project
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

	// first check if an pageToken is provided and whether is a valid bson ID
	if pageToken != "" {
		bsonID, err := primitive.ObjectIDFromHex(pageToken)
		if err != nil {
			err = fmt.Errorf("page token %s is not a valid bson ObjectId. %s", pageToken, err.Error())
			log.WithFields(
				log.Fields{
					"type":            "backend_log",
					"trace_id":        ctx.Value("trace_id"),
					"backend_service": "mongo",
					"page_token":      pageToken,
					"err":             err.Error(),
				},
			).Error("Page token is not a valid bson ObjectId")
			return qSubs, totalSize, nextPageToken, err
		}

		query["_id"] = bson.M{"$lte": bsonID}

	} else if name != "" {

		query["name"] = name
	}

	limitFindOptions := options.Find().SetLimit(limit).SetSort(bson.M{"_id": -1})
	qSubs, err = store.subsFindQueryProcessor.execute(ctx, query, limitFindOptions)

	if name == "" {

		countQuery := bson.M{"project_uuid": projectUUID}
		if userUUID != "" {
			countQuery["acl"] = bson.M{"$in": []string{userUUID}}
		}

		totalSize, err = store.subscriptionsCollection.CountDocuments(ctx, countQuery)

		if err != nil {
			store.logErrorAndCrash(ctx, "QuerySubs", err)

		}

		// if the amount of topics that were found was equal to the limit, its a sign that there are topics to populate the next page
		// so pick the last element's pageToken to use as the starting point for the next page
		// and eliminate the extra element from the current response
		if len(qSubs) > 0 && len(qSubs) == int(limit) {

			nextPageToken = qSubs[limit-1].ID.(primitive.ObjectID).Hex()
			qSubs = qSubs[:len(qSubs)-1]
		}
	}

	return qSubs, totalSize, nextPageToken, err

}

// UpdateSubLatestConsume updates the subscription's latest consume time
func (store *MongoStoreWithOfficialDriver) UpdateSubLatestConsume(ctx context.Context, projectUUID string, name string, date time.Time) error {
	doc := bson.M{
		"project_uuid": projectUUID,
		"name":         name,
	}

	change := bson.M{
		"$set": bson.M{
			"latest_consume": date,
		},
	}
	_, err := store.subscriptionsCollection.UpdateOne(ctx, doc, change)
	if err != nil {
		store.logErrorAndCrash(ctx, "UpdateSubLatestConsume", err)
	}
	return err
}

// UpdateSubConsumeRate updates the subscription's consume rate
func (store *MongoStoreWithOfficialDriver) UpdateSubConsumeRate(ctx context.Context, projectUUID string, name string, rate float64) error {
	doc := bson.M{
		"project_uuid": projectUUID,
		"name":         name,
	}

	change := bson.M{
		"$set": bson.M{
			"consume_rate": rate,
		},
	}
	_, err := store.subscriptionsCollection.UpdateOne(ctx, doc, change)
	if err != nil {
		store.logErrorAndCrash(ctx, "UpdateSubConsumeRate", err)
	}
	return err
}

// RemoveSub removes a subscription from the store
func (store *MongoStoreWithOfficialDriver) RemoveSub(ctx context.Context, projectUUID string, name string) error {
	sub := bson.M{"project_uuid": projectUUID, "name": name}
	err := store.deleteOne(ctx, store.subscriptionsCollection, sub)
	if err != nil {
		if (err == DocNotFound{}) {
			return err
		} else {
			store.logErrorAndCrash(ctx, "RemoveSub", err)
			return err
		}
	}
	return nil
}

// IncrementSubBytes increases the total number of bytes consumed from a subscription
func (store *MongoStoreWithOfficialDriver) IncrementSubBytes(ctx context.Context, projectUUID string, name string, totalBytes int64) error {
	doc := bson.M{"project_uuid": projectUUID, "name": name}
	change := bson.M{"$inc": bson.M{"total_bytes": totalBytes}}
	_, err := store.subscriptionsCollection.UpdateOne(ctx, doc, change)
	if err != nil {
		store.logErrorAndCrash(ctx, "IncrementSubBytes", err)
	}
	return err
}

// IncrementSubMsgNum increments the number of messages pulled in a subscription
func (store *MongoStoreWithOfficialDriver) IncrementSubMsgNum(ctx context.Context, projectUUID string, name string, num int64) error {
	doc := bson.M{"project_uuid": projectUUID, "name": name}
	change := bson.M{"$inc": bson.M{"msg_num": num}}
	_, err := store.subscriptionsCollection.UpdateOne(ctx, doc, change)
	if err != nil {
		store.logErrorAndCrash(ctx, "IncrementSubMsgNum", err)
	}
	return err
}

func (store *MongoStoreWithOfficialDriver) InsertSub(ctx context.Context, projectUUID string, name string, topic string,
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
	_, err := store.subscriptionsCollection.InsertOne(ctx, sub)
	if err != nil {
		store.logErrorAndCrash(ctx, "InsertSub", err)
	}
	return err
}

// QueryOneSub queries and returns specific sub of project
func (store *MongoStoreWithOfficialDriver) QueryOneSub(ctx context.Context, projectUUID string, name string) (QSub, error) {
	query := bson.M{"project_uuid": projectUUID, "name": name}
	results, err := store.subsFindQueryProcessor.execute(ctx, query)
	if err != nil {
		store.logErrorAndCrash(ctx, "QueryOneSub", err)

	}

	if len(results) > 0 {
		return results[0], nil
	}

	return QSub{}, errors.New("empty")
}

// QueryPushSubs retrieves subscriptions that have a push_endpoint defined
func (store *MongoStoreWithOfficialDriver) QueryPushSubs(ctx context.Context) []QSub {
	query := bson.M{"push_endpoint": bson.M{"$ne": ""}}
	results, err := store.subsFindQueryProcessor.execute(ctx, query)
	if err != nil {
		store.logErrorAndCrash(ctx, "QueryPushSubs", err)

	}
	return results
}

// SubscriptionsCount returns the amount of subscriptions created in the given time period per project
func (store *MongoStoreWithOfficialDriver) SubscriptionsCount(ctx context.Context, startDate, endDate time.Time,
	projectUUIDs []string) (map[string]int64, error) {

	res, err := store.getDocCountForCollectionPerProject(ctx, startDate, endDate, projectUUIDs, store.subscriptionsCollection)
	if err != nil {
		store.logErrorAndCrash(ctx, "SubscriptionsCount", err)
	}
	return res, err
}

// ModAck modifies the subscription's ack timeout field in mongodb
func (store *MongoStoreWithOfficialDriver) ModAck(ctx context.Context, projectUUID string, name string, ack int) error {
	doc := bson.M{"project_uuid": projectUUID, "name": name}
	change := bson.M{"$set": bson.M{"ack": ack}}
	_, err := store.subscriptionsCollection.UpdateOne(ctx, doc, change)
	if err != nil {
		store.logErrorAndCrash(ctx, "ModAck", err)
	}
	return err
}

// UpdateSubOffset updates a subscription offset
func (store *MongoStoreWithOfficialDriver) UpdateSubOffset(ctx context.Context, projectUUID string, name string, offset int64) {
	doc := bson.M{"project_uuid": projectUUID, "name": name}
	change := bson.M{"$set": bson.M{"offset": offset, "next_offset": 0, "pending_ack": ""}}
	_, err := store.subscriptionsCollection.UpdateOne(ctx, doc, change)
	if err != nil {
		store.logErrorAndCrash(ctx, "UpdateSubOffset", err)
	}
}

func (store *MongoStoreWithOfficialDriver) UpdateSubPull(ctx context.Context, projectUUID string, name string,
	nextOff int64, ts string) error {

	doc := bson.M{"project_uuid": projectUUID, "name": name}
	change := bson.M{"$set": bson.M{"next_offset": nextOff, "pending_ack": ts}}
	_, err := store.subscriptionsCollection.UpdateOne(ctx, doc, change)
	if err != nil {
		store.logErrorAndCrash(ctx, "UpdateSubPull", err)
		return err
	}
	return nil
}

func (store *MongoStoreWithOfficialDriver) UpdateSubOffsetAck(ctx context.Context, projectUUID string, name string,
	offset int64, ts string) error {

	// Get Info
	res := QSub{}
	query := bson.M{"project_uuid": projectUUID, "name": name}
	results, err := store.subsFindQueryProcessor.execute(ctx, query)
	if err != nil {
		store.logErrorAndCrash(ctx, "UpdateSubOffsetAck", err)
		return err
	}
	if len(results) == 0 {
		return errors.New("sub not found during UpdateSubOffsetAck")
	}
	res = results[0]

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
	change := bson.M{
		"$set": bson.M{
			"offset":      offset,
			"next_offset": 0,
			"pending_ack": "",
		},
	}
	_, err = store.subscriptionsCollection.UpdateOne(ctx, doc, change)
	if err != nil {
		store.logErrorAndCrash(ctx, "UpdateSubOffsetAck", err)
	}
	return err
}

// ModSubPush modifies the push configuration
func (store *MongoStoreWithOfficialDriver) ModSubPush(ctx context.Context, projectUUID string,
	name string, pushCfg QPushConfig) error {

	doc := bson.M{
		"project_uuid": projectUUID,
		"name":         name,
	}

	change := bson.M{"$set": bson.M{
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
	}
	ur, err := store.subscriptionsCollection.UpdateOne(ctx, doc, change)
	if err != nil {
		store.logErrorAndCrash(ctx, "ModSubPush", err)
		return err
	}
	if ur.MatchedCount == 0 {
		return DocNotFound{}
	}

	return nil
}

// ###### TOPIC QUERIES ######

func (store *MongoStoreWithOfficialDriver) LinkTopicSchema(ctx context.Context, projectUUID, name, schemaUUID string) error {
	doc := bson.M{"project_uuid": projectUUID, "name": name}
	change := bson.M{"$set": bson.M{"schema_uuid": schemaUUID}}
	_, err := store.topicsCollection.UpdateOne(ctx, doc, change)
	if err != nil {
		store.logErrorAndCrash(ctx, "LinkTopicSchema", err)
		return err

	}
	return nil
}

// QueryTopicsByACL returns topics that a specific username has access to
func (store *MongoStoreWithOfficialDriver) QueryTopicsByACL(ctx context.Context, projectUUID, user string) ([]QTopic, error) {

	// By default, return all topics of a given project
	query := bson.M{"project_uuid": projectUUID}

	// If name is given return only the specific topic
	if user != "" {
		query = bson.M{"project_uuid": projectUUID, "acl": user}
	}

	results, err := store.topicsFindQueryProcessor.execute(ctx, query)

	if err != nil {
		store.logErrorAndCrash(ctx, "QueryTopicsByACL", err)
	}

	return results, err

}

func (store *MongoStoreWithOfficialDriver) QueryTopics(ctx context.Context, projectUUID string, userUUID string, name string, pageToken string, pageSize int64) ([]QTopic, int64, string, error) {

	var err error
	var totalSize int64
	var limit int64
	var nextPageToken string
	var qTopics []QTopic

	// By default, return all topics of a given project
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

	// first check if an pageToken is provided and whether is a valid bson ID
	if pageToken != "" {
		bsonID, err := primitive.ObjectIDFromHex(pageToken)
		if err != nil {
			err = fmt.Errorf("page token %s is not a valid bson ObjectId. %s", pageToken, err.Error())
			log.WithFields(
				log.Fields{
					"type":            "backend_log",
					"trace_id":        ctx.Value("trace_id"),
					"backend_service": "mongo",
					"page_token":      pageToken,
					"err":             err.Error(),
				},
			).Error("Page token is not a valid bson ObjectId")
			return qTopics, totalSize, nextPageToken, err
		}

		query["_id"] = bson.M{"$lte": bsonID}

	} else if name != "" {

		query["name"] = name
	}

	limitFindOptions := options.Find().SetLimit(limit).SetSort(bson.M{"_id": -1})
	qTopics, err = store.topicsFindQueryProcessor.execute(ctx, query, limitFindOptions)

	if name == "" {

		countQuery := bson.M{"project_uuid": projectUUID}
		if userUUID != "" {
			countQuery["acl"] = bson.M{"$in": []string{userUUID}}
		}

		totalSize, err = store.topicsCollection.CountDocuments(ctx, countQuery)

		if err != nil {
			store.logErrorAndCrash(ctx, "QueryTopics-2", err)

		}

		// if the amount of topics that were found was equal to the limit, its a sign that there are topics to populate the next page
		// so pick the last element's pageToken to use as the starting point for the next page
		// and eliminate the extra element from the current response
		if len(qTopics) > 0 && len(qTopics) == int(limit) {

			nextPageToken = qTopics[limit-1].ID.(primitive.ObjectID).Hex()
			qTopics = qTopics[:len(qTopics)-1]
		}
	}

	return qTopics, totalSize, nextPageToken, err

}

func (store *MongoStoreWithOfficialDriver) UpdateTopicLatestPublish(ctx context.Context, projectUUID string, name string, date time.Time) error {
	doc := bson.M{
		"project_uuid": projectUUID,
		"name":         name,
	}

	change := bson.M{
		"$set": bson.M{
			"latest_publish": date,
		},
	}

	_, err := store.topicsCollection.UpdateOne(ctx, doc, change)
	if err != nil {
		store.logErrorAndCrash(ctx, "UpdateTopicLatestPublish", err)
	}
	return err
}

func (store *MongoStoreWithOfficialDriver) UpdateTopicPublishRate(ctx context.Context, projectUUID string, name string, rate float64) error {
	doc := bson.M{
		"project_uuid": projectUUID,
		"name":         name,
	}

	change := bson.M{
		"$set": bson.M{
			"publish_rate": rate,
		},
	}

	_, err := store.topicsCollection.UpdateOne(ctx, doc, change)
	if err != nil {
		store.logErrorAndCrash(ctx, "UpdateTopicPublishRate", err)
	}
	return err
}

func (store *MongoStoreWithOfficialDriver) RemoveTopic(ctx context.Context, projectUUID string, name string) error {
	query := bson.M{"project_uuid": projectUUID, "name": name}
	err := store.deleteOne(ctx, store.topicsCollection, query)
	if err != nil {
		if (err == DocNotFound{}) {
			return err
		} else {
			store.logErrorAndCrash(ctx, "RemoveTopic", err)
			return err
		}
	}
	return nil
}

func (store *MongoStoreWithOfficialDriver) InsertTopic(ctx context.Context, projectUUID string, name string, schemaUUID string, createdOn time.Time) error {
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
	_, err := store.topicsCollection.InsertOne(ctx, topic)
	if err != nil {
		store.logErrorAndCrash(ctx, "InsertTopic", err)
	}
	return err
}

// TopicsCount returns the amount of topics created in the given time period per project
func (store *MongoStoreWithOfficialDriver) TopicsCount(ctx context.Context, startDate, endDate time.Time, projectUUIDs []string) (map[string]int64, error) {
	res, err := store.getDocCountForCollectionPerProject(ctx, startDate, endDate, projectUUIDs, store.topicsCollection)
	if err != nil {
		store.logErrorAndCrash(ctx, "TopicsCount", err)
	}
	return res, err
}

// QueryDailyTopicMsgCount returns results regarding the number of messages published to a topic
func (store *MongoStoreWithOfficialDriver) QueryDailyTopicMsgCount(ctx context.Context, projectUUID string, topicName string, date time.Time) ([]QDailyTopicMsgCount, error) {

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

	limit := int64(30)
	findOptions := options.Find().SetLimit(limit).SetSort(bson.M{"date": -1})
	cursor, err := store.topicsDailyMsgCountCollection.Find(ctx, query, findOptions)
	if err != nil {
		store.logErrorAndCrash(ctx, "QueryDailyTopicMsgCount", err)
		return qDailyTopicMsgCount, err
	}

	err = cursor.All(ctx, &qDailyTopicMsgCount)
	if err != nil {
		store.logErrorAndCrash(ctx, "QueryDailyTopicMsgCount-2", err)
		return qDailyTopicMsgCount, err
	}

	return qDailyTopicMsgCount, err
}

// IncrementTopicMsgNum increments the number of messages published in a topic
func (store *MongoStoreWithOfficialDriver) IncrementTopicMsgNum(ctx context.Context, projectUUID string,
	name string, num int64) error {

	doc := bson.M{"project_uuid": projectUUID, "name": name}
	change := bson.M{"$inc": bson.M{"msg_num": num}}

	_, err := store.topicsCollection.UpdateOne(ctx, doc, change)
	if err != nil {
		store.logErrorAndCrash(ctx, "IncrementTopicMsgNum", err)
	}
	return err
}

// IncrementDailyTopicMsgCount increments the daily count of published messages to a specific topic
func (store *MongoStoreWithOfficialDriver) IncrementDailyTopicMsgCount(ctx context.Context, projectUUID string,
	topicName string, num int64, date time.Time) error {

	doc := bson.M{"date": date, "project_uuid": projectUUID, "topic_name": topicName}
	change := bson.M{"$inc": bson.M{"msg_count": num}}
	err := store.upsert(ctx, doc, change, store.topicsDailyMsgCountCollection)
	if err != nil {
		store.logErrorAndCrash(ctx, "IncrementDailyTopicMsgCount", err)
	}
	return err
}

// IncrementTopicBytes increases the total number of bytes published in a topic
func (store *MongoStoreWithOfficialDriver) IncrementTopicBytes(ctx context.Context, projectUUID string,
	name string, totalBytes int64) error {

	doc := bson.M{"project_uuid": projectUUID, "name": name}
	change := bson.M{"$inc": bson.M{"total_bytes": totalBytes}}

	_, err := store.topicsCollection.UpdateOne(ctx, doc, change)
	if err != nil {
		store.logErrorAndCrash(ctx, "IncrementTopicBytes", err)
	}
	return err
}
