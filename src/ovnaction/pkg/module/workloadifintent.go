// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2020 Intel Corporation

package module

import (
	"strings"

	"github.com/open-ness/EMCO/src/orchestrator/pkg/infra/db"

	pkgerrors "github.com/pkg/errors"
)

// WorkloadIfIntent contains the parameters needed for dynamic networks
type WorkloadIfIntent struct {
	Metadata Metadata             `json:"metadata"`
	Spec     WorkloadIfIntentSpec `json:"spec"`
}

type WorkloadIfIntentSpec struct {
	IfName         string `json:"interface"`
	NetworkName    string `json:"name"`
	DefaultGateway string `json:"defaultGateway"`       // optional, default value is "false"
	IpAddr         string `json:"ipAddress,omitempty"`  // optional, if not provided then will be dynamically allocated
	MacAddr        string `json:"macAddress,omitempty"` // optional, if not provided then will be dynamically allocated
}

// WorkloadIfIntentKey is the key structure that is used in the database
type WorkloadIfIntentKey struct {
	Project             string `json:"provider"`
	CompositeApp        string `json:"compositeapp"`
	CompositeAppVersion string `json:"compositeappversion"`
	DigName             string `json:"deploymentintentgroup"`
	NetControlIntent    string `json:"netcontrolintent"`
	WorkloadIntent      string `json:"workloadintent"`
	WorkloadIfIntent    string `json:"workloadifintent"`
}

// Manager is an interface exposing the WorkloadIfIntent functionality
type WorkloadIfIntentManager interface {
	CreateWorkloadIfIntent(wi WorkloadIfIntent, project, compositeapp, compositeappversion, dig, netcontrolintent, workloadintent string, exists bool) (WorkloadIfIntent, error)
	GetWorkloadIfIntent(name, project, compositeapp, compositeappversion, dig, netcontrolintent, workloadintent string) (WorkloadIfIntent, error)
	GetWorkloadIfIntents(project, compositeapp, compositeappversion, dig, netcontrolintent, workloadintent string) ([]WorkloadIfIntent, error)
	DeleteWorkloadIfIntent(name, project, compositeapp, compositeappversion, dig, netcontrolintent, workloadintent string) error
}

// WorkloadIfIntentClient implements the Manager
// It will also be used to maintain some localized state
type WorkloadIfIntentClient struct {
	db ClientDbInfo
}

// NewWorkloadIfIntentClient returns an instance of the WorkloadIfIntentClient
// which implements the Manager
func NewWorkloadIfIntentClient() *WorkloadIfIntentClient {
	return &WorkloadIfIntentClient{
		db: ClientDbInfo{
			storeName: "orchestrator",
			tagMeta:   "workloadifintentmetadata",
		},
	}
}

// CreateWorkloadIfIntent - create a new WorkloadIfIntent
func (v *WorkloadIfIntentClient) CreateWorkloadIfIntent(wif WorkloadIfIntent, project, compositeapp, compositeappversion, dig, netcontrolintent, workloadintent string, exists bool) (WorkloadIfIntent, error) {

	//Construct key and tag to select the entry
	key := WorkloadIfIntentKey{
		Project:             project,
		CompositeApp:        compositeapp,
		CompositeAppVersion: compositeappversion,
		DigName:             dig,
		NetControlIntent:    netcontrolintent,
		WorkloadIntent:      workloadintent,
		WorkloadIfIntent:    wif.Metadata.Name,
	}

	//Check if the Workload Intent exists
	_, err := NewWorkloadIntentClient().GetWorkloadIntent(workloadintent, project, compositeapp, compositeappversion, dig, netcontrolintent)
	if err != nil {
		return WorkloadIfIntent{}, pkgerrors.Errorf("Workload Intent %v does not exist", workloadintent)
	}

	//Check if this WorkloadIfIntent already exists
	_, err = v.GetWorkloadIfIntent(wif.Metadata.Name, project, compositeapp, compositeappversion, dig, netcontrolintent, workloadintent)
	if err == nil && !exists {
		return WorkloadIfIntent{}, pkgerrors.New("WorkloadIfIntent already exists")
	}

	err = db.DBconn.Insert(v.db.storeName, key, nil, v.db.tagMeta, wif)
	if err != nil {
		return WorkloadIfIntent{}, pkgerrors.Wrap(err, "Creating DB Entry")
	}

	return wif, nil
}

// GetWorkloadIfIntent returns the WorkloadIfIntent for corresponding name
func (v *WorkloadIfIntentClient) GetWorkloadIfIntent(name, project, compositeapp, compositeappversion, dig, netcontrolintent, workloadintent string) (WorkloadIfIntent, error) {

	//Construct key and tag to select the entry
	key := WorkloadIfIntentKey{
		Project:             project,
		CompositeApp:        compositeapp,
		CompositeAppVersion: compositeappversion,
		DigName:             dig,
		NetControlIntent:    netcontrolintent,
		WorkloadIntent:      workloadintent,
		WorkloadIfIntent:    name,
	}

	value, err := db.DBconn.Find(v.db.storeName, key, v.db.tagMeta)
	if err != nil {
		return WorkloadIfIntent{}, pkgerrors.Wrap(err, "db Find error")
	}

	//value is a byte array
	if value != nil {
		wif := WorkloadIfIntent{}
		err = db.DBconn.Unmarshal(value[0], &wif)
		if err != nil {
			return WorkloadIfIntent{}, pkgerrors.Wrap(err, "Unmarshalling Value")
		}
		return wif, nil
	}

	return WorkloadIfIntent{}, pkgerrors.New("Error getting WorkloadIfIntent")
}

// GetWorkloadIfIntentList returns all of the WorkloadIfIntent for corresponding name
func (v *WorkloadIfIntentClient) GetWorkloadIfIntents(project, compositeapp, compositeappversion, dig, netcontrolintent, workloadintent string) ([]WorkloadIfIntent, error) {

	//Construct key and tag to select the entry
	key := WorkloadIfIntentKey{
		Project:             project,
		CompositeApp:        compositeapp,
		CompositeAppVersion: compositeappversion,
		DigName:             dig,
		NetControlIntent:    netcontrolintent,
		WorkloadIntent:      workloadintent,
		WorkloadIfIntent:    "",
	}

	var resp []WorkloadIfIntent
	values, err := db.DBconn.Find(v.db.storeName, key, v.db.tagMeta)
	if err != nil {
		return []WorkloadIfIntent{}, pkgerrors.Wrap(err, "db Find error")
	}

	for _, value := range values {
		wif := WorkloadIfIntent{}
		err = db.DBconn.Unmarshal(value, &wif)
		if err != nil {
			return []WorkloadIfIntent{}, pkgerrors.Wrap(err, "Unmarshalling Value")
		}
		resp = append(resp, wif)
	}

	return resp, nil
}

// Delete the  WorkloadIfIntent from database
func (v *WorkloadIfIntentClient) DeleteWorkloadIfIntent(name, project, compositeapp, compositeappversion, dig, netcontrolintent, workloadintent string) error {

	//Construct key and tag to select the entry
	key := WorkloadIfIntentKey{
		Project:             project,
		CompositeApp:        compositeapp,
		CompositeAppVersion: compositeappversion,
		DigName:             dig,
		NetControlIntent:    netcontrolintent,
		WorkloadIntent:      workloadintent,
		WorkloadIfIntent:    name,
	}

	err := db.DBconn.Remove(v.db.storeName, key)
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
