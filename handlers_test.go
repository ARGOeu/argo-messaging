package main

import (
	"log"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ARGOeu/argo-messaging/Godeps/_workspace/src/github.com/stretchr/testify/suite"
)

type HandlerTestSuite struct {
	suite.Suite
}

func (suite *HandlerTestSuite) TestTopicListOne() {

	req, err := http.NewRequest("GET", "http://localhost:8080/v1/projects/ARGO/topics/topic1", nil)
	if err != nil {
		log.Fatal(err)
	}

	expResp := `{
   "name": "/project/ARGO/topics/topic1"
}`

	w := httptest.NewRecorder()
	TopicListOne(w, req)
	suite.Equal(200, w.Code)
	suite.Equal(expResp, w.Body.String())
}

func (suite *HandlerTestSuite) TestTopicListAll() {

	req, err := http.NewRequest("GET", "http://localhost:8080/v1/projects/ARGO", nil)
	if err != nil {
		log.Fatal(err)
	}

	expResp := `{
   "topics": [
      {
         "name": "/project/ARGO/topics/topic1"
      },
      {
         "name": "/project/ARGO/topics/topic2"
      }
   ]
}`

	w := httptest.NewRecorder()
	TopicListAll(w, req)
	suite.Equal(200, w.Code)
	suite.Equal(expResp, w.Body.String())
}

func TestFactorsTestSuite(t *testing.T) {
	suite.Run(t, new(HandlerTestSuite))
}
