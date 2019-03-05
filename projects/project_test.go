package projects

import (
	"errors"
	"testing"
	"time"

	"github.com/ARGOeu/argo-messaging/stores"
	"github.com/stretchr/testify/suite"
)

type ProjectsTestSuite struct {
	suite.Suite
}

func (suite *ProjectsTestSuite) TestProjects() {
	store := stores.NewMockStore("mockhost", "mockbase")
	tm := time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC)

	item1 := NewProject("argo_uuid", "ARGO", tm, tm, "UserA", "simple project")
	item2 := NewProject("argo_uuid2", "ARGO2", tm, tm, "UserA", "simple project")
	ep1 := Projects{List: []Project{item1}}
	ep2 := Projects{List: []Project{item2}}
	ep3 := Projects{List: []Project{item1, item2}}
	ep4 := Projects{}

	p1, err := Find("", "ARGO", store)
	suite.Equal(ep1, p1)
	suite.Equal(nil, err)
	p2, err := Find("", "ARGO2", store)
	suite.Equal(ep2, p2)
	suite.Equal(nil, err)
	p3, err := Find("", "", store)
	suite.Equal(ep3, p3)
	suite.Equal(nil, err)
	p4, err := Find("", "FOO", store)

	suite.Equal(ep4, p4)
	suite.Equal(errors.New("not found"), err)

	// Create new project
	itemNew := NewProject("uuid_new", "BRAND_NEW", tm, tm, "UserA", "brand new project")

	reflect, err := CreateProject("uuid_new", "BRAND_NEW", tm, "uuid1", "brand new project", store)

	expNew := Projects{List: []Project{itemNew}}
	expAllNew := Projects{List: []Project{item1, item2, itemNew}}

	pNew, err := Find("", "BRAND_NEW", store)

	suite.Equal(expNew.List[0], reflect)
	suite.Equal(expNew, pNew)
	suite.Equal(nil, err)

	// Test GetNameByUUID
	suite.Equal("BRAND_NEW", GetNameByUUID("uuid_new", store))

	pAllNew, err := Find("", "", store)

	suite.Equal(expAllNew, pAllNew)
	suite.Equal(nil, err)

	// Test Export Json
	expJSON := `{
   "projects": [
      {
         "name": "ARGO",
         "created_on": "2009-11-10T23:00:00Z",
         "modified_on": "2009-11-10T23:00:00Z",
         "created_by": "UserA",
         "description": "simple project"
      },
      {
         "name": "ARGO2",
         "created_on": "2009-11-10T23:00:00Z",
         "modified_on": "2009-11-10T23:00:00Z",
         "created_by": "UserA",
         "description": "simple project"
      },
      {
         "name": "BRAND_NEW",
         "created_on": "2009-11-10T23:00:00Z",
         "modified_on": "2009-11-10T23:00:00Z",
         "created_by": "UserA",
         "description": "brand new project"
      }
   ]
}`
	outJSON, err := pAllNew.ExportJSON()
	suite.Equal(expJSON, outJSON)
	suite.Equal(nil, err)

	// Test Get from json
	prJSON := `{

    "description":"project with only description"
  }`
	expGen01 := Project{Description: "project with only description"}

	prGen01, err := GetFromJSON([]byte(prJSON))
	suite.Equal(expGen01, prGen01)
	suite.Equal(nil, err)

	prJSON2 := `{
  "created_on": "2009-11-10T23:00:00Z",
  "modified_on": "2009-11-10T23:00:00Z",
  "description":"another description"
  }`

	zuluForm := "2006-01-02T15:04:05Z"

	expGen02 := Project{CreatedOn: tm.Format(zuluForm), ModifiedOn: tm.Format(zuluForm), Description: "another description"}

	prGen02, err := GetFromJSON([]byte(prJSON2))
	suite.Equal(expGen02, prGen02)
	suite.Equal(nil, err)

	// Test erroneous json
	prJSON3 := `{
  "created_on": "2009-11-10T23:00:00Z",
  "modified_other description"
  }`
	expGen03 := Project{}

	prGen03, err := GetFromJSON([]byte(prJSON3))
	suite.Equal(expGen03, prGen03)
	suite.Equal(true, err != nil)

	// Test updates

	expUpdJSON := `{
   "projects": [
      {
         "name": "NEW_ARGO",
         "created_on": "2009-11-10T23:00:00Z",
         "modified_on": "2009-11-10T23:00:00Z",
         "created_by": "UserA",
         "description": "a new description and name for  project"
      },
      {
         "name": "ARGO2",
         "created_on": "2009-11-10T23:00:00Z",
         "modified_on": "2009-11-10T23:00:00Z",
         "created_by": "UserA",
         "description": "this project has only description changed"
      },
      {
         "name": "ONLY_NAME_CHANGED",
         "created_on": "2009-11-10T23:00:00Z",
         "modified_on": "2009-11-10T23:00:00Z",
         "created_by": "UserA",
         "description": "brand new project"
      }
   ]
}`

	UpdateProject("argo_uuid", "NEW_ARGO", "a new description and name for  project", tm, store)
	UpdateProject("argo_uuid2", "", "this project has only description changed", tm, store)
	UpdateProject("uuid_new", "ONLY_NAME_CHANGED", "", tm, store)

	pAllUpdated, _ := Find("", "", store)
	outAllUpdJSON, _ := pAllUpdated.ExportJSON()

	suite.Equal(expUpdJSON, outAllUpdJSON)

	// Test removing project
	RemoveProject("argo_uuid", store)
	pRemoved, err := Find("argo_uuid", "", store)
	suite.Equal(Projects{}, pRemoved)
	suite.Equal(errors.New("not found"), err)
	// Check to see that also projects topics and subscriptions have been removed from the store

	resTop, _, _, _ := store.QueryTopics("argo_uuid", "", "", "", 0)
	suite.Equal(0, len(resTop))
	resSub, _, _, _ := store.QuerySubs("argo_uuid", "", "", "", 0)
	suite.Equal(0, len(resSub))

}

func TestProjectsTestSuite(t *testing.T) {
	suite.Run(t, new(ProjectsTestSuite))
}
