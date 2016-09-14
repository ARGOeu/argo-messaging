package projects

import (
	"encoding/json"
	"errors"

	"github.com/ARGOeu/argo-messaging/stores"
)

import "time"

// Project is the struct that holds Project information
type Project struct {
	UUID        string    `json:"-"`
	Name        string    `json:"name,omitempty"`
	CreatedOn   time.Time `json:"created_on,omitempty"`
	ModifiedOn  time.Time `json:"modified_on,omitempty"`
	CreatedBy   string    `json:"created_by,omitempty"`
	Description string    `json:"description,omitempty"`
}

// Projects holds a list of available projects
type Projects struct {
	List []Project `json:"projects,omitempty"`
}

// ExportJSON exports Project to json format
func (p *Project) ExportJSON() (string, error) {
	output, err := json.MarshalIndent(p, "", "   ")
	return string(output[:]), err
}

// ExportJSON exports Projects list to json format
func (ps *Projects) ExportJSON() (string, error) {
	output, err := json.MarshalIndent(ps, "", "   ")
	return string(output[:]), err
}

// GetFromJSON retrieves Project info From JSON string
func GetFromJSON(input []byte) (Project, error) {
	p := Project{}
	err := json.Unmarshal([]byte(input), &p)
	return p, err
}

// NewProject accepts parameters and creates a new project
func NewProject(uuid string, name string, createdOn time.Time, modifiedOn time.Time, createdBy string, description string) Project {
	return Project{UUID: uuid, Name: name, CreatedOn: createdOn, ModifiedOn: modifiedOn, CreatedBy: createdBy, Description: description}
}

// Find returns a specific project or a list of all available projects in the datastore.
// To return all projects use an empty project string parameter
func Find(name string, store stores.Store) (Projects, error) {
	result := Projects{}
	// if project string empty, returns all projects
	projects, err := store.QueryProjects(name,"")
	for _, item := range projects {
		curProject := NewProject(item.UUID, item.Name, item.CreatedOn, item.ModifiedOn, item.CreatedBy, item.Description)
		result.List = append(result.List, curProject)
	}

	return result, err
}

// GetNameByUUID queries projects by UUID and returns the project name. If not found, returns an empty string
func GetNameByUUID(uuid string, store stores.Store) (string) {
	result := ""
	// if project string empty, returns all projects
	projects, err := store.QueryProjects("",uuid)
	if len(projects) > 0 && err == nil {
		result = projects[0].Name
	}

	return result;
}

// GetNameByUUID queries projects by UUID and returns the project name. If not found, returns an empty string
func GetUUIDByName(name string, store stores.Store) (string) {
	result := ""
	// if project string empty, returns all projects
	projects, err := store.QueryProjects(name,"")
	if len(projects) > 0 && err == nil {
		result = projects[0].UUID
	}

	return result;
}


// HasProject if store contains a project with the specific name
func HasProject(name string, store stores.Store) bool {
	projects, _ := store.QueryProjects(name,"")
	return len(projects) > 0
}

// CreateProject creates a new project
func CreateProject(uuid string, name string, createdOn time.Time, createdBy string, description string, store stores.Store) (Project, error) {
	if HasProject(name, store) {
		return Project{}, errors.New("exists")
	}

	projNew := NewProject(uuid, name, createdOn, createdOn, createdBy, description)
	err := store.InsertProject(uuid, name, createdOn, createdOn, createdBy, description)

	return projNew, err
}
