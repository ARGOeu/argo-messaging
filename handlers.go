package main

import (
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/ARGOeu/argo-messaging/topics"
)

// TopicListOne (GET) one topic
func TopicListOne(w http.ResponseWriter, r *http.Request) {

	code := http.StatusOK
	output := []byte("")
	contentType := "application/json"
	charset := "utf-8"
	urlVars := strings.Split(r.URL.Path, "/")

	tp := topics.Topics{}
	tp.LoadFromCfg(kafkaCfg)
	res := topics.Topic{}
	res = tp.GetTopicByName(urlVars[3], urlVars[5])
	resJSON, err := res.ExportJSON()
	if err != nil {
		log.Panic("Error in response transformation to JSON", err)
	}
	output = []byte(resJSON)
	w.Header().Add("Content-Type", fmt.Sprintf("%s; charset=%s", contentType, charset))
	w.WriteHeader(code)
	w.Write(output)
}

// TopicListAll (GET) all topics
func TopicListAll(w http.ResponseWriter, r *http.Request) {

	code := http.StatusOK
	output := []byte("")
	contentType := "application/json"
	charset := "utf-8"

	urlVars := strings.Split(r.URL.Path, "/")
	tp := topics.Topics{}
	tp.LoadFromCfg(kafkaCfg)
	res := topics.Topics{}
	res = tp.GetTopicsByProject(urlVars[3])
	resJSON, err := res.ExportJSON()
	if err != nil {
		log.Panic("Error in response transformation to JSON", err)
	}
	output = []byte(resJSON)
	w.Header().Add("Content-Type", fmt.Sprintf("%s; charset=%s", contentType, charset))
	w.WriteHeader(code)
	w.Write(output)
}
