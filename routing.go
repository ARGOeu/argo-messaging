package main

import (
	"net/http"

	log "github.com/sirupsen/logrus"

	"github.com/ARGOeu/argo-messaging/brokers"
	"github.com/ARGOeu/argo-messaging/config"
	oldPush "github.com/ARGOeu/argo-messaging/push"
	push "github.com/ARGOeu/argo-messaging/push/grpc/client"
	"github.com/ARGOeu/argo-messaging/stores"
	gorillaContext "github.com/gorilla/context"
	"github.com/gorilla/mux"
)

// API Object that holds routing information and the router itself
type API struct {
	Router *mux.Router // router Object reference
	Routes []APIRoute  // routes definitions list
}

// APIRoute Item describing an API route
type APIRoute struct {
	Name    string           // name of the route for description purposes
	Method  string           // GET, POST, PUT string litterals
	Path    string           // API call path with urlVars included
	Handler http.HandlerFunc // Handler Function to be used
}

// NewRouting creates a new routing object including mux.Router and routes definitions
func NewRouting(cfg *config.APICfg, brk brokers.Broker, str stores.Store, mgr *oldPush.Manager, c push.Client, routes []APIRoute) *API {
	// Create the api Object
	ar := API{}
	// Create a new router and reference him in API object
	ar.Router = mux.NewRouter().StrictSlash(false)
	// reference routes input in API object too keep info centralized
	ar.Routes = routes

	// For each route
	for _, route := range ar.Routes {

		// prepare handle wrappers
		var handler http.HandlerFunc
		handler = route.Handler

		handler = WrapLog(handler, route.Name)

		// skip authentication/authorization for the health status and profile api calls
		if route.Name != "ams:healthStatus" && "users:profile" != route.Name && route.Name != "version:list" && route.Name != "registrations:newUser" {
			handler = WrapAuthorize(handler, route.Name)
			handler = WrapAuthenticate(handler)
		}

		handler = WrapValidate(handler)
		handler = WrapConfig(handler, cfg, brk, str, mgr, c)

		ar.Router.
			PathPrefix("/v1").
			Methods(route.Method).
			Path(route.Path).
			Name(route.Name).
			Handler(gorillaContext.ClearHandler(handler))
	}

	log.Info("API", "\t", "API Router initialized! Ready to start listening...")
	// Return reference to API object
	return &ar
}

// Global list populated with default routes
var defaultRoutes = []APIRoute{

	{"ams:metrics", "GET", "/metrics", OpMetrics},
	{"ams:healthStatus", "GET", "/status", HealthCheck},
	{"ams:dailyMessageAverage", "GET", "/metrics/daily-message-average", DailyMessageAverage},
	{"users:byToken", "GET", "/users:byToken/{token}", UserListByToken},
	{"users:byUUID", "GET", "/users:byUUID/{uuid}", UserListByUUID},
	{"users:list", "GET", "/users", UserListAll},
	{"users:profile", "GET", "/users/profile", UserProfile},
	{"users:show", "GET", "/users/{user}", UserListOne},
	{"users:refreshToken", "POST", "/users/{user}:refreshToken", RefreshToken},
	{"users:create", "POST", "/users/{user}", UserCreate},
	{"users:update", "PUT", "/users/{user}", UserUpdate},
	{"users:delete", "DELETE", "/users/{user}", UserDelete},
	{"registrations:newUser", "POST", "/registrations/{user}", RegisterUser},
	{"projects:list", "GET", "/projects", ProjectListAll},
	{"projects:metrics", "GET", "/projects/{project}:metrics", ProjectMetrics},
	{"projects:addUser", "POST", "/projects/{project}/members/{user}:add", ProjectUserAdd},
	{"projects:removeUser", "POST", "/projects/{project}/members/{user}:remove", ProjectUserRemove},
	{"projects:showUser", "GET", "/projects/{project}/members/{user}", ProjectUserListOne},
	{"projects:createUser", "POST", "/projects/{project}/members/{user}", ProjectUserCreate},
	{"projects:updateUser", "PUT", "/projects/{project}/members/{user}", ProjectUserUpdate},
	{"projects:listUsers", "GET", "/projects/{project}/members", ProjectListUsers},
	{"projects:show", "GET", "/projects/{project}", ProjectListOne},
	{"projects:create", "POST", "/projects/{project}", ProjectCreate},
	{"projects:update", "PUT", "/projects/{project}", ProjectUpdate},
	{"projects:delete", "DELETE", "/projects/{project}", ProjectDelete},
	{"subscriptions:list", "GET", "/projects/{project}/subscriptions", SubListAll},
	{"subscriptions:listByTopic", "GET", "/projects/{project}/topics/{topic}/subscriptions", ListSubsByTopic},
	{"subscriptions:offsets", "GET", "/projects/{project}/subscriptions/{subscription}:offsets", SubGetOffsets},
	{"subscriptions:timeToOffset", "GET", "/projects/{project}/subscriptions/{subscription}:timeToOffset", SubTimeToOffset},
	{"subscriptions:acl", "GET", "/projects/{project}/subscriptions/{subscription}:acl", SubACL},
	{"subscriptions:metrics", "GET", "/projects/{project}/subscriptions/{subscription}:metrics", SubMetrics},
	{"subscriptions:show", "GET", "/projects/{project}/subscriptions/{subscription}", SubListOne},
	{"subscriptions:create", "PUT", "/projects/{project}/subscriptions/{subscription}", SubCreate},
	{"subscriptions:delete", "DELETE", "/projects/{project}/subscriptions/{subscription}", SubDelete},
	{"subscriptions:pull", "POST", "/projects/{project}/subscriptions/{subscription}:pull", SubPull},
	{"subscriptions:acknowledge", "POST", "/projects/{project}/subscriptions/{subscription}:acknowledge", SubAck},
	{"subscriptions:verifyPushEndpoint", "POST", "/projects/{project}/subscriptions/{subscription}:verifyPushEndpoint", SubVerifyPushEndpoint},
	{"subscriptions:modifyAckDeadline", "POST", "/projects/{project}/subscriptions/{subscription}:modifyAckDeadline", SubModAck},
	{"subscriptions:modifyPushConfig", "POST", "/projects/{project}/subscriptions/{subscription}:modifyPushConfig", SubModPush},
	{"subscriptions:modifyOffset", "POST", "/projects/{project}/subscriptions/{subscription}:modifyOffset", SubSetOffset},
	{"subscriptions:modifyAcl", "POST", "/projects/{project}/subscriptions/{subscription}:modifyAcl", SubModACL},
	{"topics:list", "GET", "/projects/{project}/topics", TopicListAll},
	{"topics:acl", "GET", "/projects/{project}/topics/{topic}:acl", TopicACL},
	{"topics:metrics", "GET", "/projects/{project}/topics/{topic}:metrics", TopicMetrics},
	{"topics:show", "GET", "/projects/{project}/topics/{topic}", TopicListOne},
	{"topics:create", "PUT", "/projects/{project}/topics/{topic}", TopicCreate},
	{"topics:delete", "DELETE", "/projects/{project}/topics/{topic}", TopicDelete},
	{"topics:publish", "POST", "/projects/{project}/topics/{topic}:publish", TopicPublish},
	{"topics:modifyAcl", "POST", "/projects/{project}/topics/{topic}:modifyAcl", TopicModACL},
	{"schemas:validateMessage", "POST", "/projects/{project}/schemas/{schema}:validate", SchemaValidateMessage},
	{"schemas:create", "POST", "/projects/{project}/schemas/{schema}", SchemaCreate},
	{"schemas:show", "GET", "/projects/{project}/schemas/{schema}", SchemaListOne},
	{"schemas:list", "GET", "/projects/{project}/schemas", SchemaListAll},
	{"schemas:update", "PUT", "/projects/{project}/schemas/{schema}", SchemaUpdate},
	{"schemas:delete", "DELETE", "/projects/{project}/schemas/{schema}", SchemaDelete},
	{"version:list", "GET", "/version", ListVersion},
}
