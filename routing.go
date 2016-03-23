package main

import (
	"log"
	"net/http"

	"github.com/ARGOeu/argo-messaging/brokers"
	"github.com/ARGOeu/argo-messaging/config"
	"github.com/ARGOeu/argo-messaging/stores"
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
func NewRouting(cfg *config.APICfg, brk brokers.Broker, str stores.Store, routes []APIRoute) *API {
	// Create the api Object
	ar := API{}
	// Create a new router and reference him in API object
	ar.Router = mux.NewRouter().StrictSlash(true)
	// reference routes input in API object too keep info centralized
	ar.Routes = routes

	// For each route
	for _, route := range ar.Routes {

		// prepare handle wrappers
		var handler http.HandlerFunc
		handler = route.Handler

		handler = WrapLog(handler, route.Name)

		if cfg.Author && cfg.Authen {

			handler = WrapAuthorize(handler, route.Name)
		}

		if cfg.Authen {

			handler = WrapAuthenticate(handler)
		}
		handler = WrapConfig(handler, cfg, brk, str)

		ar.Router.
			PathPrefix("/v1").
			Methods(route.Method).
			Path(route.Path).
			Handler(handler)
	}

	if cfg.Authen {
		log.Printf("INFO\tAPI\tAPI Authentication mechanism enabled")
	}

	if cfg.Author {
		log.Printf("INFO\tAPI\tAPI Authorization mechanism enabled")
	}

	log.Printf("INFO\tAPI\tAPI Router initialized! Ready to start listening...")
	// Return reference to API object
	return &ar
}

// Global list populated with default routes
var defaultRoutes = []APIRoute{
	{"subscriptions:list", "GET", "/projects/{project}/subscriptions", SubListAll},
	{"subscriptions:show", "GET", "/projects/{project}/subscriptions/{subscription}", SubListOne},
	{"subscriptions:create", "PUT", "/projects/{project}/subscriptions/{subscription}", SubCreate},
	{"subscriptions:pull", "POST", "/projects/{project}/subscriptions/{subscription}:pull", SubPull},
	{"topics:list", "GET", "/projects/{project}/topics", TopicListAll},
	{"topics:show", "GET", "/projects/{project}/topics/{topic}", TopicListOne},
	{"topics:create", "PUT", "/projects/{project}/topics/{topic}", TopicCreate},
	{"topics:delete", "DELETE", "/projects/{project}/topics/{topic}", TopicDelete},
	{"topics:publish", "POST", "/projects/{project}/topics/{topic}:publish", TopicPublish},
}
