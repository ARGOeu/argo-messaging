package main

import (
	"crypto/tls"
	"fmt"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"net/http"
	"strconv"

	"github.com/ARGOeu/argo-messaging/brokers"
	"github.com/ARGOeu/argo-messaging/config"
	oldPush "github.com/ARGOeu/argo-messaging/push"
	push "github.com/ARGOeu/argo-messaging/push/grpc/client"
	"github.com/ARGOeu/argo-messaging/stores"
	"github.com/ARGOeu/argo-messaging/version"
	"github.com/gorilla/handlers"
	log "github.com/sirupsen/logrus"
)

func init() {
	// don't use colors in output
	log.SetFormatter(
		&log.TextFormatter{
			FullTimestamp: true,
			DisableColors: true},
	)

	// display binary version information during start up
	version.LogInfo()

}

func main() {

	// create and load configuration object
	cfg := config.NewAPICfg("LOAD")

	// create the store
	store := stores.NewMongoStore(cfg.StoreHost, cfg.StoreDB)
	store.Initialize()

	// create and initialize broker based on configuration
	broker := brokers.NewKafkaBroker(cfg.GetBrokerInfo())
	defer broker.CloseConnections()

	mgr := &oldPush.Manager{}

	// ams push server pushClient
	pushClient := push.NewGrpcClient(cfg)
	err := pushClient.Dial()
	if err != nil {
		log.WithFields(
			log.Fields{
				"type":            "backend_log",
				"backend_service": "ams-push-server",
				"backend_hosts":   pushClient.Target(),
			},
		).Error(err.Error())
	}

	defer pushClient.Close()

	// create and initialize API routing object
	API := NewRouting(cfg, broker, store, mgr, pushClient, defaultRoutes)

	//Configure TLS support only
	config := &tls.Config{
		MinVersion:               tls.VersionTLS10,
		PreferServerCipherSuites: true,
	}

	// Initialize CORS specifics
	xReqWithConType := handlers.AllowedHeaders([]string{"X-Requested-With", "Content-Type"})
	allowVerbs := handlers.AllowedMethods([]string{"OPTIONS", "POST", "GET", "PUT", "DELETE", "HEAD"})
	// Initialize server wth proper parameters
	server := &http.Server{Addr: ":" + strconv.Itoa(cfg.Port), Handler: handlers.CORS(xReqWithConType, allowVerbs)(API.Router), TLSConfig: config}

	UQ()

	// Web service binds to server. Requests served over HTTPS.
	err = server.ListenAndServeTLS(cfg.Cert, cfg.CertKey)
	if err != nil {
		log.Fatal("API", "\t", "ListenAndServe:", err)
	}

}

func UQ() {

	query := []bson.M{

		// create a unique pair for every user and each one of its projects
		// even if a user doesn't belong to any projects, keep him in the grand total result
		{
			"$unwind": bson.M{
				"path":                       "$projects",
				"preserveNullAndEmptyArrays": true,
			},
		},

		// for each project uuid look up the additional project details
		//from the projects collection
		{
			"$lookup": bson.M{
				"from":         "projects",
				"localField":   "projects.project_uuid",
				"foreignField": "uuid",
				"as":           "project_info",
			},
		},

		// project_uuid can only map to 1 item
		// we can unwind the project_info array since it will always contain 1 item
		{
			"$unwind": bson.M{
				"path":                       "$project_info",
				"preserveNullAndEmptyArrays": true,
			},
		},

		// for each user/project load the respective topics that the user belongs to their acl
		{
			"$lookup": bson.M{
				"from": "topics",
				"let": bson.M{
					"q_user_uuid": "$uuid",
					"q_proj_uuid": "$project_info.uuid",
				},
				"pipeline": []bson.M{
					{
						"$match": bson.M{
							"$expr": bson.M{
								"$and": []bson.M{
									{
										"$in": []string{"$$q_user_uuid", "$acl"},
									},
									{
										"$eq": []string{"$$q_proj_uuid", "$project_uuid"},
									},
								},
							},
						},
					},
				},
				"as": "topic_info",
			},
		},

		// for each user/project load the respective subscriptions that the user belongs to their acl
		{
			"$lookup": bson.M{
				"from": "subscriptions",
				"let": bson.M{
					"q_user_uuid": "$uuid",
					"q_proj_uuid": "$project_info.uuid",
				},
				"pipeline": []bson.M{
					{
						"$match": bson.M{
							"$expr": bson.M{
								"$and": []bson.M{
									{
										"$in": []string{"$$q_user_uuid", "$acl"},
									},
									{
										"$eq": []string{"$$q_proj_uuid", "$project_uuid"},
									},
								},
							},
						},
					},
				},
				"as": "sub_info",
			},
		},

		// unwind the topics array in order to group and project the wanted view
		// of an array of just topic names, e.g. ["t1", "t2", "t3"]
		{
			"$unwind": bson.M{
				"path":                       "$topic_info",
				"preserveNullAndEmptyArrays": true,
			},
		},

		// unwind the subs array in order to group and project the wanted view
		// of an array of just sub names, e.g. ["s1", "s2", "s3"]
		{
			"$unwind": bson.M{
				"path":                       "$sub_info",
				"preserveNullAndEmptyArrays": true,
			},
		},

		{
			"$group": bson.M{
				"_id": bson.M{
					"_id":          "$_id",
					"project_uuid": "$project_info.uuid",
				},
				"topics": bson.M{
					"$addToSet": "$topic_info.name",
				},
				"subscriptions": bson.M{
					"$addToSet": "$sub_info.name",
				},
				"name": bson.M{
					"$first": "$name",
				},
				"uuid": bson.M{
					"$first": "uuid",
				},
				"email": bson.M{
					"$first": "email",
				},
				"token": bson.M{
					"$first": "token",
				},
				"service_roles": bson.M{
					"$first": "service_roles",
				},
				"project_info": bson.M{
					"$first": "project_info",
				},
				"projects": bson.M{
					"$first": "projects",
				},
				"created_on": bson.M{
					"$first": "created_on",
				},
				"modified_on": bson.M{
					"$first": "modified_on",
				},
			},
		},

		{
			"$project": bson.M{
				"_id":  "$_id._id",
				"uuid": 1,
				"project_info": bson.M{
					"name":          "$project_info.name",
					"roles":         "$projects.roles",
					"topics":        "$topics",
					"subscriptions": "$subscriptions",
				},
				"name":          1,
				"token":         1,
				"email":         1,
				"service_roles": 1,
				"created_on":    1,
				"modified_on":   1,
			},
		},

		// group by user id and push all projects into a single array
		{
			"$group": bson.M{
				"_id": "$_id",
				"root": bson.M{
					"$mergeObjects": "$$ROOT",
				},
				"projects": bson.M{
					"$push": "$project_info",
				},
			},
		},
		{
			"$replaceRoot": bson.M{
				"newRoot": bson.M{
					"$mergeObjects": []string{"$root", "$$ROOT"},
				},
			},
		},
		// remove the placeholder fields of root and project info
		{
			"$project": bson.M{
				"root":         0,
				"project_info": 0,
			},
		},
	}

	session, err := mgo.Dial("127.0.0.1")
	if err != nil {
		panic(err)
	}

	db := session.DB("argo_msg")
	c := db.C("users")

	res := []map[string]interface{}{}

	err = c.Pipe(query).All(&res)
	if err != nil {
		panic(err)
	}

	fmt.Printf("\n%+v\n", res)

}
