// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2020 Intel Corporation

package module

import (
	"encoding/json"
	"strings"

	"github.com/open-ness/EMCO/src/orchestrator/pkg/infra/db"

	pkgerrors "github.com/pkg/errors"
)

// Project contains the metaData for Projects
type Project struct {
	MetaData ProjectMetaData `json:"metadata"`
}

// ProjectMetaData contains the parameters for creating a project
type ProjectMetaData struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	UserData1   string `json:"userData1"`
	UserData2   string `json:"userData2"`
}

// ProjectKey is the key structure that is used in the database
type ProjectKey struct {
	ProjectName string `json:"project"`
}

// We will use json marshalling to convert to string to
// preserve the underlying structure.
func (pk ProjectKey) String() string {
	out, err := json.Marshal(pk)
	if err != nil {
		return ""
	}

	return string(out)
}

// ProjectManager is an interface exposes the Project functionality
type ProjectManager interface {
	CreateProject(pr Project, exists bool) (Project, error)
	GetProject(name string) (Project, error)
	DeleteProject(name string) error
	GetAllProjects() ([]Project, error)
}

// ProjectClient implements the ProjectManager
// It will also be used to maintain some localized state
type ProjectClient struct {
	storeName           string
	tagMeta, tagContent string
}

// NewProjectClient returns an instance of the ProjectClient
// which implements the ProjectManager
func NewProjectClient() *ProjectClient {
	return &ProjectClient{
		storeName: "orchestrator",
		tagMeta:   "projectmetadata",
	}
}

// CreateProject a new collection based on the project
func (v *ProjectClient) CreateProject(p Project, exists bool) (Project, error) {

	//Construct the composite key to select the entry
	key := ProjectKey{
		ProjectName: p.MetaData.Name,
	}

	//Check if this Project already exists
	_, err := v.GetProject(p.MetaData.Name)
	if err == nil && !exists {
		return Project{}, pkgerrors.New("Project already exists")
	}

	err = db.DBconn.Insert(v.storeName, key, nil, v.tagMeta, p)
	if err != nil {
		return Project{}, pkgerrors.Wrap(err, "Creating DB Entry")
	}

	return p, nil
}

// GetProject returns the Project for corresponding name
func (v *ProjectClient) GetProject(name string) (Project, error) {

	//Construct the composite key to select the entry
	key := ProjectKey{
		ProjectName: name,
	}
	value, err := db.DBconn.Find(v.storeName, key, v.tagMeta)
	if err != nil {
		return Project{}, pkgerrors.Wrap(err, "db Find error")
	} else if len(value) == 0 {
		return Project{}, pkgerrors.New("Project not found")
	}

	//value is a byte array
	if value != nil {
		proj := Project{}
		err = db.DBconn.Unmarshal(value[0], &proj)
		if err != nil {
			return Project{}, pkgerrors.Wrap(err, "Unmarshaling Value")
		}
		return proj, nil
	}

	return Project{}, pkgerrors.New("Error getting Project")
}

// GetAllProjects returns all the projects
func (v *ProjectClient) GetAllProjects() ([]Project, error) {
	key := ProjectKey{
		ProjectName: "",
	}

	var res []Project
	values, err := db.DBconn.Find(v.storeName, key, v.tagMeta)
	if err != nil {
		return []Project{}, pkgerrors.Wrap(err, "db Find error")
	}

	for _, value := range values {
		p := Project{}
		err = db.DBconn.Unmarshal(value, &p)
		if err != nil {
			return []Project{}, pkgerrors.Wrap(err, "Unmarshaling Project")
		}
		res = append(res, p)
	}
	return res, nil
}

// DeleteProject the  Project from database
func (v *ProjectClient) DeleteProject(name string) error {

	//Construct the composite key to select the entry
	key := ProjectKey{
		ProjectName: name,
	}
	err := db.DBconn.Remove(v.storeName, key)
	if err != nil {
		if strings.Contains(err.Error(), "Error finding:") {
			return pkgerrors.Wrap(err, "db Remove error - not found")
		} else if strings.Contains(err.Error(), "Can't delete parent without deleting child") {
			return pkgerrors.Wrap(err, "db Remove error - conflict")
		} else {
			return pkgerrors.Wrap(err, "db Remove error - general")
		}
	}

	//TODO: Delete the collection when the project is deleted
	return nil
}
