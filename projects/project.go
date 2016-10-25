package projects

import (
	"encoding/json"
	"errors"

	"time"

	"github.com/ARGOeu/argo-messaging/stores"
)

// Project is the struct that holds Project information
type Project struct {
	UUID        string `json:"-"`
	Name        string `json:"name,omitempty"`
	CreatedOn   string `json:"created_on,omitempty"`
	ModifiedOn  string `json:"modified_on,omitempty"`
	CreatedBy   string `json:"created_by,omitempty"`
	Description string `json:"description,omitempty"`
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

// Empty returns true if projects list is empty
func (ps *Projects) Empty() bool {
	if ps.List == nil {
		return true
	}
	return len(ps.List) <= 0
}

// One returns the first project if a projects list is not empty
func (ps *Projects) One() Project {
	if ps.Empty() == false {
		return ps.List[0]
	}
	return Project{}
}

// GetFromJSON retrieves Project info From JSON string
func GetFromJSON(input []byte) (Project, error) {
	p := Project{}
	err := json.Unmarshal([]byte(input), &p)
	return p, err
}

// NewProject accepts parameters and creates a new project
func NewProject(uuid string, name string, createdOn time.Time, modifiedOn time.Time, createdBy string, description string) Project {
	zuluForm := "2006-01-02T15:04:05Z"
	return Project{UUID: uuid, Name: name, CreatedOn: createdOn.Format(zuluForm), ModifiedOn: modifiedOn.Format(zuluForm), CreatedBy: createdBy, Description: description}
}

// Find returns a specific project or a list of all available projects in the datastore.
// To return all projects use an empty project string parameter
func Find(uuid string, name string, store stores.Store) (Projects, error) {
	result := Projects{}
	// if project string empty, returns all projects
	projects, err := store.QueryProjects(uuid, name)

	for _, item := range projects {
		// Get Username from user uuid
		username := ""
		usr, err := store.QueryUsers("", item.CreatedBy, "")
		if err == nil {
			username = usr[0].Name
		}
		curProject := NewProject(item.UUID, item.Name, item.CreatedOn, item.ModifiedOn, username, item.Description)
		result.List = append(result.List, curProject)
	}

	return result, err
}

// GetNameByUUID queries projects by UUID and returns the project name. If not found, returns an empty string
func GetNameByUUID(uuid string, store stores.Store) string {
	result := ""
	// if project string empty, returns all projects
	projects, err := store.QueryProjects(uuid, "")

	if len(projects) > 0 && err == nil {
		result = projects[0].Name
	}

	return result
}

// GetUUIDByName queries project by name and returns the corresponding UUID
func GetUUIDByName(name string, store stores.Store) string {
	result := ""
	// if project string empty, returns all projects
	projects, err := store.QueryProjects("", name)
	if len(projects) > 0 && err == nil {
		result = projects[0].UUID
	}
	return result
}

// ExistsWithName returns true if a project with name exists
func ExistsWithName(name string, store stores.Store) bool {
	if name == "" {
		return false
	}

	result := false

	projects, err := store.QueryProjects("", name)
	if len(projects) > 0 && err == nil {
		result = true
	}

	return result

}

// ExistsWithUUID return true if a project with uuid exists
func ExistsWithUUID(uuid string, store stores.Store) bool {
	if uuid == "" {
		return false
	}

	result := false

	projects, err := store.QueryProjects(uuid, "")
	if len(projects) > 0 && err == nil {
		result = true
	}

	return result
}

// HasProject if store contains a project with the specific name
func HasProject(name string, store stores.Store) bool {
	projects, _ := store.QueryProjects("", name)

	return len(projects) > 0

}

// CreateProject creates a new project
func CreateProject(uuid string, name string, createdOn time.Time, createdBy string, description string, store stores.Store) (Project, error) {

	// check if project with the same name exists
	if ExistsWithName(name, store) {
		return Project{}, errors.New("exists")
	}

	if err := store.InsertProject(uuid, name, createdOn, createdOn, createdBy, description); err != nil {
		return Project{}, errors.New("backend error")
	}

	// reflect stored object
	stored, err := Find("", name, store)
	return stored.One(), err
}

// UpdateProject creates a new project
func UpdateProject(uuid string, name string, description string, modifiedOn time.Time, store stores.Store) (Project, error) {
	// Project with uuid should exist to be updated

	// check if project with the same name exists
	if ExistsWithUUID(uuid, store) == false {
		return Project{}, errors.New("not found")
	}

	if err := store.UpdateProject(uuid, name, description, modifiedOn); err != nil {
		return Project{}, err
	}

	// reflect stored object
	stored, err := Find(uuid, name, store)
	return stored.One(), err
}

// RemoveProject removes project
func RemoveProject(uuid string, store stores.Store) error {
	// Project with uuid should exist to be updated

	// check if project with the same name exists
	if ExistsWithUUID(uuid, store) == false {
		return errors.New("not found")
	}

	// Remove project it self
	if err := store.RemoveProject(uuid); err != nil {

		if err.Error() == "not found" {
			return err
		}

		return errors.New("backend error")
	}

	// Remove topics attached to this project
	if err := store.RemoveProjectTopics(uuid); err != nil {

		if err.Error() == "not found" {
			return err
		}

		return errors.New("backend error")
	}

	// Remove subscriptions attached to this project
	if err := store.RemoveProjectSubs(uuid); err != nil {

		if err.Error() == "not found" {
			return err
		}

		return errors.New("backend error")
	}

	return nil

}
