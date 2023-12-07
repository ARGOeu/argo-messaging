package projects

import (
	"context"
	"encoding/json"
	"errors"

	"time"

	"github.com/ARGOeu/argo-messaging/stores"
)

// Project is the struct that holds ProjectUUID information
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

// ExportJSON exports ProjectUUID to json format
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

// GetFromJSON retrieves ProjectUUID info From JSON string
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
func Find(ctx context.Context, uuid string, name string, store stores.Store) (Projects, error) {
	result := Projects{}
	// if project string empty, returns all projects
	projects, err := store.QueryProjects(ctx, uuid, name)

	for _, item := range projects {
		// Get Username from user uuid
		username := ""
		if item.CreatedBy != "" {
			usr, err := store.QueryUsers(ctx, "", item.CreatedBy, "")
			if err == nil && len(usr) > 0 {
				username = usr[0].Name
			}
		}
		curProject := NewProject(item.UUID, item.Name, item.CreatedOn.UTC(), item.ModifiedOn.UTC(), username, item.Description)
		result.List = append(result.List, curProject)
	}

	return result, err
}

// GetNameByUUID queries projects by UUID and returns the project name. If not found, returns an empty string
func GetNameByUUID(ctx context.Context, uuid string, store stores.Store) string {
	result := ""

	if uuid != "" {
		projects, err := store.QueryProjects(ctx, uuid, "")
		if len(projects) > 0 && err == nil {
			result = projects[0].Name
		}
	}

	return result
}

// GetUUIDByName queries project by name and returns the corresponding UUID
func GetUUIDByName(ctx context.Context, name string, store stores.Store) string {
	result := ""

	if name != "" {
		projects, err := store.QueryProjects(ctx, "", name)
		if len(projects) > 0 && err == nil {
			result = projects[0].UUID
		}
	}

	return result
}

// ExistsWithName returns true if a project with name exists
func ExistsWithName(ctx context.Context, name string, store stores.Store) bool {
	if name == "" {
		return false
	}

	result := false

	projects, err := store.QueryProjects(ctx, "", name)
	if len(projects) > 0 && err == nil {
		result = true
	}

	return result

}

// ExistsWithUUID return true if a project with uuid exists
func ExistsWithUUID(ctx context.Context, uuid string, store stores.Store) bool {
	if uuid == "" {
		return false
	}

	result := false

	projects, err := store.QueryProjects(ctx, uuid, "")
	if len(projects) > 0 && err == nil {
		result = true
	}

	return result
}

// HasProject if store contains a project with the specific name
func HasProject(ctx context.Context, name string, store stores.Store) bool {
	projects, _ := store.QueryProjects(ctx, "", name)

	return len(projects) > 0

}

// CreateProject creates a new project
func CreateProject(ctx context.Context, uuid string, name string, createdOn time.Time, createdBy string, description string, store stores.Store) (Project, error) {
	// check if project with the same name exists
	if ExistsWithName(ctx, name, store) {
		return Project{}, errors.New("exists")
	}

	if err := store.InsertProject(ctx, uuid, name, createdOn, createdOn, createdBy, description); err != nil {
		return Project{}, errors.New("backend error")
	}

	// reflect stored object
	stored, err := Find(ctx, "", name, store)

	return stored.One(), err
}

// UpdateProject creates a new project
func UpdateProject(ctx context.Context, uuid string, name string, description string, modifiedOn time.Time, store stores.Store) (Project, error) {
	// ProjectUUID with uuid should exist to be updated

	// check if project with the same name exists
	if ExistsWithUUID(ctx, uuid, store) == false {
		return Project{}, errors.New("not found")
	}

	if err := store.UpdateProject(ctx, uuid, name, description, modifiedOn); err != nil {
		return Project{}, err
	}

	// reflect stored object
	stored, err := Find(ctx, uuid, name, store)
	return stored.One(), err
}

// RemoveProject removes project
func RemoveProject(ctx context.Context, uuid string, store stores.Store) error {
	// ProjectUUID with uuid should exist to be updated

	// check if project with the same name exists
	if ExistsWithUUID(ctx, uuid, store) == false {
		return errors.New("not found")
	}

	// Remove project it self
	if err := store.RemoveProject(ctx, uuid); err != nil {

		if err.Error() == "not found" {
			return err
		}

		return errors.New("backend error")
	}

	// Remove topics attached to this project
	if err := store.RemoveProjectTopics(ctx, uuid); err != nil {

		if err.Error() == "not found" {
			return err
		}

		return errors.New("backend error")
	}

	// Remove subscriptions attached to this project
	if err := store.RemoveProjectSubs(ctx, uuid); err != nil {

		if err.Error() == "not found" {
			return err
		}

		return errors.New("backend error")
	}

	if err := store.RemoveProjectDailyMessageCounters(ctx, uuid); err != nil {
		if err.Error() == "not found" {
			return err
		}

		return errors.New("backend error")
	}

	return nil

}
