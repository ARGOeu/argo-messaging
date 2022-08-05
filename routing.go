package main

import (
	"net/http"

	log "github.com/sirupsen/logrus"

	"github.com/ARGOeu/argo-messaging/brokers"
	"github.com/ARGOeu/argo-messaging/config"
	"github.com/ARGOeu/argo-messaging/handlers"
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

	tokenExtractStrategy := handlers.GetRequestTokenExtractStrategy(cfg.AuthOption())

	// For each route
	for _, route := range ar.Routes {

		// prepare handle wrappers
		var handler http.HandlerFunc
		handler = route.Handler

		handler = handlers.WrapLog(handler, route.Name)

		// skip authentication/authorization for the health status and profile api calls
		if route.Name != "ams:healthStatus" && "users:profile" != route.Name && route.Name != "version:list" {
			handler = handlers.WrapAuthorize(handler, route.Name, tokenExtractStrategy)
			handler = handlers.WrapAuthenticate(handler, tokenExtractStrategy)
		}

		handler = handlers.WrapValidate(handler)
		handler = handlers.WrapConfig(handler, cfg, brk, str, mgr, c)

		ar.Router.
			PathPrefix("/v1").
			Methods(route.Method).
			Path(route.Path).
			Name(route.Name).
			Handler(gorillaContext.ClearHandler(handler))
	}

	log.WithFields(
		log.Fields{
			"type": "service_log",
		},
	).Info("API Router initialized! Ready to start listening...")

	// Return reference to API object
	return &ar
}

// Global list populated with default routes
var defaultRoutes = []APIRoute{

	{"ams:metrics", "GET", "/metrics", handlers.OpMetrics},
	{"ams:healthStatus", "GET", "/status", handlers.HealthCheck},
	{"ams:vaMetrics", "GET", "/metrics/va_metrics", handlers.VaMetrics},
	{"users:byToken", "GET", "/users:byToken/{token}", handlers.UserListByToken},
	{"users:byUUID", "GET", "/users:byUUID/{uuid}", handlers.UserListByUUID},
	{"users:list", "GET", "/users", handlers.UserListAll},
	{"users:profile", "GET", "/users/profile", handlers.UserProfile},
	{"users:show", "GET", "/users/{user}", handlers.UserListOne},
	{"users:refreshToken", "POST", "/users/{user}:refreshToken", handlers.RefreshToken},
	{"users:create", "POST", "/users/{user}", handlers.UserCreate},
	{"users:update", "PUT", "/users/{user}", handlers.UserUpdate},
	{"users:delete", "DELETE", "/users/{user}", handlers.UserDelete},
	{"registrations:newUser", "POST", "/registrations", handlers.RegisterUser},
	{"registrations:acceptNewUser", "POST", "/registrations/{uuid}:accept", handlers.AcceptRegisterUser},
	{"registrations:declineNewUser", "POST", "/registrations/{uuid}:decline", handlers.DeclineRegisterUser},
	{"registrations:delete", "DELETE", "/registrations/{uuid}", handlers.DeleteRegistration},
	{"registrations:show", "GET", "/registrations/{uuid}", handlers.ListOneRegistration},
	{"registrations:list", "GET", "/registrations", handlers.ListAllRegistrations},
	{"projects:list", "GET", "/projects", handlers.ProjectListAll},
	{"projects:metrics", "GET", "/projects/{project}:metrics", handlers.ProjectMetrics},
	{"projects:addUser", "POST", "/projects/{project}/members/{user}:add", handlers.ProjectUserAdd},
	{"projects:removeUser", "POST", "/projects/{project}/members/{user}:remove", handlers.ProjectUserRemove},
	{"projects:showUser", "GET", "/projects/{project}/members/{user}", handlers.ProjectUserListOne},
	{"projects:createUser", "POST", "/projects/{project}/members/{user}", handlers.ProjectUserCreate},
	{"projects:updateUser", "PUT", "/projects/{project}/members/{user}", handlers.ProjectUserUpdate},
	{"projects:listUsers", "GET", "/projects/{project}/members", handlers.ProjectListUsers},
	{"projects:show", "GET", "/projects/{project}", handlers.ProjectListOne},
	{"projects:create", "POST", "/projects/{project}", handlers.ProjectCreate},
	{"projects:update", "PUT", "/projects/{project}", handlers.ProjectUpdate},
	{"projects:delete", "DELETE", "/projects/{project}", handlers.ProjectDelete},
	{"subscriptions:list", "GET", "/projects/{project}/subscriptions", handlers.SubListAll},
	{"subscriptions:listByTopic", "GET", "/projects/{project}/topics/{topic}/subscriptions", handlers.ListSubsByTopic},
	{"subscriptions:offsets", "GET", "/projects/{project}/subscriptions/{subscription}:offsets", handlers.SubGetOffsets},
	{"subscriptions:timeToOffset", "GET", "/projects/{project}/subscriptions/{subscription}:timeToOffset", handlers.SubTimeToOffset},
	{"subscriptions:acl", "GET", "/projects/{project}/subscriptions/{subscription}:acl", handlers.SubACL},
	{"subscriptions:metrics", "GET", "/projects/{project}/subscriptions/{subscription}:metrics", handlers.SubMetrics},
	{"subscriptions:show", "GET", "/projects/{project}/subscriptions/{subscription}", handlers.SubListOne},
	{"subscriptions:create", "PUT", "/projects/{project}/subscriptions/{subscription}", handlers.SubCreate},
	{"subscriptions:delete", "DELETE", "/projects/{project}/subscriptions/{subscription}", handlers.SubDelete},
	{"subscriptions:pull", "POST", "/projects/{project}/subscriptions/{subscription}:pull", handlers.SubPull},
	{"subscriptions:acknowledge", "POST", "/projects/{project}/subscriptions/{subscription}:acknowledge", handlers.SubAck},
	{"subscriptions:verifyPushEndpoint", "POST", "/projects/{project}/subscriptions/{subscription}:verifyPushEndpoint", handlers.SubVerifyPushEndpoint},
	{"subscriptions:modifyAckDeadline", "POST", "/projects/{project}/subscriptions/{subscription}:modifyAckDeadline", handlers.SubModAck},
	{"subscriptions:modifyPushConfig", "POST", "/projects/{project}/subscriptions/{subscription}:modifyPushConfig", handlers.SubModPush},
	{"subscriptions:modifyOffset", "POST", "/projects/{project}/subscriptions/{subscription}:modifyOffset", handlers.SubSetOffset},
	{"subscriptions:modifyAcl", "POST", "/projects/{project}/subscriptions/{subscription}:modifyAcl", handlers.SubModACL},
	{"topics:list", "GET", "/projects/{project}/topics", handlers.TopicListAll},
	{"topics:acl", "GET", "/projects/{project}/topics/{topic}:acl", handlers.TopicACL},
	{"topics:metrics", "GET", "/projects/{project}/topics/{topic}:metrics", handlers.TopicMetrics},
	{"topics:show", "GET", "/projects/{project}/topics/{topic}", handlers.TopicListOne},
	{"topics:create", "PUT", "/projects/{project}/topics/{topic}", handlers.TopicCreate},
	{"topics:delete", "DELETE", "/projects/{project}/topics/{topic}", handlers.TopicDelete},
	{"topics:publish", "POST", "/projects/{project}/topics/{topic}:publish", handlers.TopicPublish},
	{"topics:modifyAcl", "POST", "/projects/{project}/topics/{topic}:modifyAcl", handlers.TopicModACL},
	{"topics:attachSchema", "POST", "/projects/{project}/topics/{topic}:attachSchema", handlers.TopicAttachSchema},
	{"topics:detachSchema", "POST", "/projects/{project}/topics/{topic}:detachSchema", handlers.TopicDetachSchema},
	{"schemas:validateMessage", "POST", "/projects/{project}/schemas/{schema}:validate", handlers.SchemaValidateMessage},
	{"schemas:create", "POST", "/projects/{project}/schemas/{schema}", handlers.SchemaCreate},
	{"schemas:show", "GET", "/projects/{project}/schemas/{schema}", handlers.SchemaListOne},
	{"schemas:list", "GET", "/projects/{project}/schemas", handlers.SchemaListAll},
	{"schemas:update", "PUT", "/projects/{project}/schemas/{schema}", handlers.SchemaUpdate},
	{"schemas:delete", "DELETE", "/projects/{project}/schemas/{schema}", handlers.SchemaDelete},
	{"version:list", "GET", "/version", handlers.ListVersion},
}
