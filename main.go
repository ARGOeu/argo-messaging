package main

import (
	"fmt"
	"html"
	"log"
	"net/http"

	"strconv"

	"github.com/ARGOeu/argo-messaging/config"
	"github.com/gorilla/mux"
)

func main() {
	defer broker.CloseConnections()
	broker.Initialize(config.Kafka.Server)
	router := mux.NewRouter().StrictSlash(true)
	router.HandleFunc("/", ReflectRoute)
	router.HandleFunc("/v1/", ReflectRoute)
	router.HandleFunc("/v1/pub/{topic}", RawPublish)
	router.HandleFunc("/v1/sub/{topic}", RawConsume)
	log.Fatal(http.ListenAndServe(":8080", router))
}

// ReflectRoute is a temporary function which confirms that the api is alive and reflects the route used
func ReflectRoute(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "API alive!\nRoute: %q", html.EscapeString(r.URL.Path))
}

// RawPublish This is a temporary api call to test pub functionality until publish model is implemented
func RawPublish(w http.ResponseWriter, r *http.Request) {
	urlValues := r.URL.Query()
	urlVars := mux.Vars(r)
	m, p, o := broker.Publish(urlVars["topic"], urlValues.Get("message"))
	fmt.Fprintf(w, "Message Published!")
	fmt.Fprintf(w, "\nkafka endpoint: "+config.Kafka.Server)
	fmt.Fprintf(w, "\nTopic:"+m)
	fmt.Fprintf(w, "\nPartition: "+strconv.Itoa(p))
	fmt.Fprintf(w, "\nOffset: "+strconv.Itoa(o))
	fmt.Fprintf(w, "\nMessage:"+urlValues.Get("message"))
}

// RawConsume to test if we can consume
func RawConsume(w http.ResponseWriter, r *http.Request) {
	urlValues := r.URL.Query()
	urlVars := mux.Vars(r)
	offset, _ := strconv.ParseInt(urlValues.Get("offset"), 10, 64)
	fmt.Fprintf(w, "\nkafka endpoint: "+config.Kafka.Server+"\n")
	m := broker.Consume(urlVars["topic"], offset)
	for index, value := range m {
		fmt.Fprintf(w, "message "+strconv.Itoa(index)+" : "+value+"\n")
	}

}
