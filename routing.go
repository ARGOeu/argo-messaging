package main

import (
	"log"
	"net/http"

	"github.com/ARGOeu/argo-messaging/Godeps/_workspace/src/github.com/gorilla/mux"
	"github.com/ARGOeu/argo-messaging/brokers"
	"github.com/ARGOeu/argo-messaging/config"
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
func NewRouting(cfg *config.KafkaCfg, brk brokers.Broker, routes []APIRoute) *API {
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
		handler = WrapConfig(handler, cfg, brk)

		ar.Router.
			PathPrefix("/v1").
			Methods(route.Method).
			Path(route.Path).
			Handler(handler)
	}

	log.Printf("INFO\tAPI\tAPI Router initialized! Ready to start listening...")
	// Return reference to API object
	return &ar
}

// Global list populated with default routes
var defaultRoutes = []APIRoute{
	{"Topics List All", "GET", "/projects/{project}/topics", TopicListAll},
	{"Topics List One", "GET", "/projects/{project}/topics/{topic}", TopicListOne},
	{"Topics Publish", "POST", "/projects/{project}/topics/{topic}:publish", TopicPublish},
}
