// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2020 Intel Corporation

package module

import (
	"encoding/json"
	"strings"

	"github.com/open-ness/EMCO/src/orchestrator/pkg/infra/db"

	pkgerrors "github.com/pkg/errors"
)

// CompositeApp contains metadata and spec for CompositeApps
type CompositeApp struct {
	Metadata CompositeAppMetaData `json:"metadata"`
	Spec     CompositeAppSpec     `json:"spec"`
}

//CompositeAppMetaData contains the parameters needed for CompositeApps
type CompositeAppMetaData struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	UserData1   string `json:"userData1"`
	UserData2   string `json:"userData2"`
}

//CompositeAppSpec contains the Version of the CompositeApp
type CompositeAppSpec struct {
	Version string `json:"version"`
}

// CompositeAppKey is the key structure that is used in the database
type CompositeAppKey struct {
	CompositeAppName string `json:"compositeapp"`
	Version          string `json:"compositeappversion"`
	Project          string `json:"project"`
}

// We will use json marshalling to convert to string to
// preserve the underlying structure.
func (cK CompositeAppKey) String() string {
	out, err := json.Marshal(cK)
	if err != nil {
		return ""
	}
	return string(out)
}

// CompositeAppManager is an interface exposes the CompositeApp functionality
type CompositeAppManager interface {
	CreateCompositeApp(c CompositeApp, p string, exists bool) (CompositeApp, error)
	GetCompositeApp(name string, version string, p string) (CompositeApp, error)
	GetAllCompositeApps(p string) ([]CompositeApp, error)
	DeleteCompositeApp(name string, version string, p string) error
}

// CompositeAppClient implements the CompositeAppManager
// It will also be used to maintain some localized state
type CompositeAppClient struct {
	storeName string
	tagMeta   string
}

// NewCompositeAppClient returns an instance of the CompositeAppClient
// which implements the CompositeAppManager
func NewCompositeAppClient() *CompositeAppClient {
	return &CompositeAppClient{
		storeName: "orchestrator",
		tagMeta:   "compositeappmetadata",
	}
}

// CreateCompositeApp creates a new collection based on the CompositeApp
func (v *CompositeAppClient) CreateCompositeApp(c CompositeApp, p string, exists bool) (CompositeApp, error) {

	//Construct the composite key to select the entry
	key := CompositeAppKey{
		CompositeAppName: c.Metadata.Name,
		Version:          c.Spec.Version,
		Project:          p,
	}

	//Check if this CompositeApp already exists
	_, err := v.GetCompositeApp(c.Metadata.Name, c.Spec.Version, p)
	if err == nil && !exists{
		return CompositeApp{}, pkgerrors.New("CompositeApp already exists")
	}

	//Check if Project exists
	_, err = NewProjectClient().GetProject(p)
	if err != nil {
		return CompositeApp{}, pkgerrors.New("Unable to find the project")
	}

	err = db.DBconn.Insert(v.storeName, key, nil, v.tagMeta, c)
	if err != nil {
		return CompositeApp{}, pkgerrors.Wrap(err, "Creating DB Entry")
	}

	return c, nil
}

// GetCompositeApp returns the CompositeApp for corresponding name
func (v *CompositeAppClient) GetCompositeApp(name string, version string, p string) (CompositeApp, error) {

	//Construct the composite key to select the entry
	key := CompositeAppKey{
		CompositeAppName: name,
		Version:          version,
		Project:          p,
	}
	value, err := db.DBconn.Find(v.storeName, key, v.tagMeta)
	if err != nil {
		return CompositeApp{}, pkgerrors.Wrap(err, "db Find error")
	} else if len(value) == 0 {
		return CompositeApp{}, pkgerrors.New("CompositeApp not found")
	}

	//value is a byte array
	if value != nil {
		compApp := CompositeApp{}
		err = db.DBconn.Unmarshal(value[0], &compApp)
		if err != nil {
			return CompositeApp{}, pkgerrors.Wrap(err, "Unmarshaling Value")
		}
		return compApp, nil
	}

	return CompositeApp{}, pkgerrors.New("Error getting composite application")
}

// GetAllCompositeApps returns all the compositeApp for a given project
func (v *CompositeAppClient) GetAllCompositeApps(p string) ([]CompositeApp, error) {

	_, err := NewProjectClient().GetProject(p)
	if err != nil {
		return []CompositeApp{}, pkgerrors.New("Unable to find the project")
	}

	key := CompositeAppKey{
		CompositeAppName: "",
		Version:          "",
		Project:          p,
	}

	var caList []CompositeApp
	values, err := db.DBconn.Find(v.storeName, key, v.tagMeta)
	if err != nil {
		return []CompositeApp{}, pkgerrors.Wrap(err, "db Find error")
	} else if len(values) == 0 {
		return []CompositeApp{}, pkgerrors.New("CompositeApps not found")
	}

	for _, value := range values {
		ca := CompositeApp{}
		err = db.DBconn.Unmarshal(value, &ca)
		if err != nil {
			return []CompositeApp{}, pkgerrors.Wrap(err, "Unmarshaling CompositeApp")
		}
		caList = append(caList, ca)
	}

	return caList, nil
}

// DeleteCompositeApp deletes the  CompositeApp from database
func (v *CompositeAppClient) DeleteCompositeApp(name string, version string, p string) error {

	//Construct the composite key to select the entry
	key := CompositeAppKey{
		CompositeAppName: name,
		Version:          version,
		Project:          p,
	}
	err := db.DBconn.Remove(v.storeName, key)
	if err != nil {
		if strings.Contains(err.Error(), "Error finding:") {
			return pkgerrors.Wrap(err, "db Remove error - not found")
		} else if strings.Contains(err.Error(), "Can't delete parent without deleting child") {
			return pkgerrors.Wrap(err, "db Remove error - conflict")
		} else if strings.Contains(err.Error(), "not found") {
			return pkgerrors.Wrap(err, "Composite App not found")
		} else {
			return pkgerrors.Wrap(err, "db Remove error - general")
		}
	}

	return nil
}
