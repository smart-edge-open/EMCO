package module

// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2020 Intel Corporation

import (
	"encoding/json"
	"strings"

	"github.com/open-ness/EMCO/src/orchestrator/pkg/infra/db"
	pkgerrors "github.com/pkg/errors"
)

// GenericK8sIntent consists of metadata
type GenericK8sIntent struct {
	Metadata Metadata `json:"metadata"`
}

// GenericK8sIntentKey consists generick8sintentName, project, compositeApp, compositeAppVersion, deploymentIntentGroupName
type GenericK8sIntentKey struct {
	GenericK8sIntent    string `json:"generick8sintent"`
	Project             string `json:"project"`
	CompositeApp        string `json:"compositeapp"`
	CompositeAppVersion string `json:"compositeappversion"`
	DigName             string `json:"deploymentintentgroup"`
}

// GenericK8sIntentManager is an interface exposing the GenericK8sIntent functionality
type GenericK8sIntentManager interface {
	CreateGenericK8sIntent(gki GenericK8sIntent, project, compositeapp, compositeappversion, dig string, exists bool) (GenericK8sIntent, error)
	GetGenericK8sIntent(gkiName, project, compositeapp, compositeappversion, dig string) (GenericK8sIntent, error)
	GetAllGenericK8sIntents(project, compositeapp, compositeappversion, dig string) ([]GenericK8sIntent, error)
	DeleteGenericK8sIntent(gkiName, project, compositeapp, compositeappversion, dig string) error
}

// GenericK8sIntentClient consists of the clientInfo
type GenericK8sIntentClient struct {
	db ClientDbInfo
}

// We will use json marshalling to convert to string to
// preserve the underlying structure.
func (gk GenericK8sIntentKey) String() string {
	out, err := json.Marshal(gk)
	if err != nil {
		return ""
	}
	return string(out)
}

// NewGenericK8sIntentClient returns an instance of the GenericK8sIntentClient
func NewGenericK8sIntentClient() *GenericK8sIntentClient {
	return &GenericK8sIntentClient{
		db: ClientDbInfo{
			storeName: "orchestrator",
			tagMeta:   "generick8sintentmetadata",
		},
	}
}

// CreateGenericK8sIntent creates a new GenericK8sIntent
func (g *GenericK8sIntentClient) CreateGenericK8sIntent(gki GenericK8sIntent, project, compositeapp, compositeappversion, dig string, exists bool) (GenericK8sIntent, error) {

	key := GenericK8sIntentKey{
		GenericK8sIntent:    gki.Metadata.Name,
		Project:             project,
		CompositeApp:        compositeapp,
		CompositeAppVersion: compositeappversion,
		DigName:             dig,
	}

	//Check if this GenericK8sIntent already exists
	_, err := g.GetGenericK8sIntent(gki.Metadata.Name, project, compositeapp, compositeappversion, dig)
	if err == nil && !exists {
		return GenericK8sIntent{}, pkgerrors.New("GenericK8sIntent already exists")
	}

	err = db.DBconn.Insert(g.db.storeName, key, nil, g.db.tagMeta, gki)
	if err != nil {
		return GenericK8sIntent{}, pkgerrors.Wrap(err, "Creating DB Entry")
	}

	return gki, nil
}

// GetGenericK8sIntent returns GenericK8sIntent with the corresponding name
func (g *GenericK8sIntentClient) GetGenericK8sIntent(gkiName, project, compositeapp, compositeappversion, dig string) (GenericK8sIntent, error) {

	//Construct key and tag to select the entry
	key := GenericK8sIntentKey{
		GenericK8sIntent:    gkiName,
		Project:             project,
		CompositeApp:        compositeapp,
		CompositeAppVersion: compositeappversion,
		DigName:             dig,
	}

	value, err := db.DBconn.Find(g.db.storeName, key, g.db.tagMeta)
	if err != nil {
		return GenericK8sIntent{}, pkgerrors.Wrap(err, "db Find error")
	}

	//value is a byte array
	if value != nil {
		gki := GenericK8sIntent{}
		err = db.DBconn.Unmarshal(value[0], &gki)
		if err != nil {
			return GenericK8sIntent{}, pkgerrors.Wrap(err, "Unmarshalling Value")
		}
		return gki, nil
	}

	return GenericK8sIntent{}, pkgerrors.New("Error getting GenericK8sIntent")
}

// GetAllGenericK8sIntents returns all of the GenericK8sIntent for corresponding name
func (g *GenericK8sIntentClient) GetAllGenericK8sIntents(project, compositeapp, compositeappversion, dig string) ([]GenericK8sIntent, error) {

	//Construct key and tag to select the entry
	key := GenericK8sIntentKey{
		GenericK8sIntent:    "",
		Project:             project,
		CompositeApp:        compositeapp,
		CompositeAppVersion: compositeappversion,
		DigName:             dig,
	}

	var resp []GenericK8sIntent
	values, err := db.DBconn.Find(g.db.storeName, key, g.db.tagMeta)
	if err != nil {
		return []GenericK8sIntent{}, pkgerrors.Wrap(err, "db Find error")
	}

	for _, value := range values {
		gki := GenericK8sIntent{}
		err = db.DBconn.Unmarshal(value, &gki)
		if err != nil {
			return []GenericK8sIntent{}, pkgerrors.Wrap(err, "Unmarshalling Value")
		}
		resp = append(resp, gki)
	}

	return resp, nil
}

// DeleteGenericK8sIntent delete the GenericK8sIntent entry from the database
func (g *GenericK8sIntentClient) DeleteGenericK8sIntent(gkiName, project, compositeapp, compositeappversion, dig string) error {

	//Construct key and tag to select the entry
	key := GenericK8sIntentKey{
		GenericK8sIntent:    gkiName,
		Project:             project,
		CompositeApp:        compositeapp,
		CompositeAppVersion: compositeappversion,
		DigName:             dig,
	}

	err := db.DBconn.Remove(g.db.storeName, key)
	if err != nil {
		if strings.Contains(err.Error(), "Error finding:") {
			return pkgerrors.Wrap(err, "db Remove error - not found")
		} else if strings.Contains(err.Error(), "Can't delete parent without deleting child") {
			return pkgerrors.Wrap(err, "db Remove error - conflict")
		} else {
			return pkgerrors.Wrap(err, "db Remove error - general")
		}
	}

	return nil
}
