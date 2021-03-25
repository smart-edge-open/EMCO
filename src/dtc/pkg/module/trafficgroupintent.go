// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2020 Intel Corporation

package module

import (
	"strings"

	"github.com/open-ness/EMCO/src/orchestrator/pkg/infra/db"
	pkgerrors "github.com/pkg/errors"
)

type TrafficGroupIntent struct {
	Metadata Metadata `json:"metadata"`
}

type TrafficGroupIntentManager interface {
	CreateTrafficGroupIntent(tci TrafficGroupIntent, project, compositeapp, compositeappversion, deploymentIntentGroupName string, exists bool) (TrafficGroupIntent, error)

	GetTrafficGroupIntent(name, project, compositeapp, compositeappversion, dig string) (TrafficGroupIntent, error)
	GetTrafficGroupIntents(project, compositeapp, compositeappversion, dig string) ([]TrafficGroupIntent, error)
	DeleteTrafficGroupIntent(name, project, compositeapp, compositeappversion, dig string) error
}

type TrafficGroupIntentDbClient struct {
	db ClientDbInfo
}

// TrafficGroupIntentKey is the key structure that is used in the database
type TrafficGroupIntentKey struct {
	TrafficGroupIntentName    string `json:"trafficgroupintentname"`
	Project                   string `json:"project"`
	CompositeApp              string `json:"compositeapp"`
	CompositeAppVersion       string `json:"compositeappversion"`
	DeploymentIntentGroupName string `json:"deploymentintentgroupname"`
}

func NewTrafficGroupIntentClient() *TrafficGroupIntentDbClient {
	return &TrafficGroupIntentDbClient{
		db: ClientDbInfo{
			storeName: "orchestrator",
			tagMeta:   "trafficgroupintentmetadata",
		},
	}
}

func (v TrafficGroupIntentDbClient) CreateTrafficGroupIntent(tci TrafficGroupIntent, project, compositeapp, compositeappversion, deploymentintentgroupname string, exists bool) (TrafficGroupIntent, error) {

	//Construct key and tag to select the entry
	key := TrafficGroupIntentKey{
		TrafficGroupIntentName:    tci.Metadata.Name,
		Project:                   project,
		CompositeApp:              compositeapp,
		CompositeAppVersion:       compositeappversion,
		DeploymentIntentGroupName: deploymentintentgroupname,
	}
	//Check if this TrafficGroupIntent already exists
	_, err := v.GetTrafficGroupIntent(tci.Metadata.Name, project, compositeapp, compositeappversion, deploymentintentgroupname)
	if err == nil && !exists {
		return TrafficGroupIntent{}, pkgerrors.New("TrafficGroupIntent already exists")
	}

	err = db.DBconn.Insert(v.db.storeName, key, nil, v.db.tagMeta, tci)
	if err != nil {
		return TrafficGroupIntent{}, pkgerrors.Wrap(err, "Creating DB Entry")
	}

	return tci, nil
}

// GetTrafficGroupIntent returns the TrafficGroupIntent for corresponding name
func (v *TrafficGroupIntentDbClient) GetTrafficGroupIntent(name, project, compositeapp, compositeappversion, dig string) (TrafficGroupIntent, error) {

	//Construct key and tag to select the entry
	key := TrafficGroupIntentKey{
		TrafficGroupIntentName:    name,
		Project:                   project,
		CompositeApp:              compositeapp,
		CompositeAppVersion:       compositeappversion,
		DeploymentIntentGroupName: dig,
	}

	value, err := db.DBconn.Find(v.db.storeName, key, v.db.tagMeta)
	if err != nil {
		return TrafficGroupIntent{}, pkgerrors.Wrap(err, "db Find error")
	} else if len(value) == 0 {
		return TrafficGroupIntent{}, pkgerrors.New("Traffic group intent not found")

	}

	//value is a byte array
	if value != nil {
		tgi := TrafficGroupIntent{}
		err = db.DBconn.Unmarshal(value[0], &tgi)
		if err != nil {
			return TrafficGroupIntent{}, pkgerrors.Wrap(err, "db Unmarshal error")
		}
		return tgi, nil
	}

	return TrafficGroupIntent{}, pkgerrors.New("Error getting TrafficGroupIntent")

}

// GetTrafficGroupIntents returns all of the TrafficGroupIntents
func (v *TrafficGroupIntentDbClient) GetTrafficGroupIntents(project, compositeapp, compositeappversion, dig string) ([]TrafficGroupIntent, error) {

	//Construct key and tag to select the entry
	key := TrafficGroupIntentKey{
		TrafficGroupIntentName:    "",
		Project:                   project,
		CompositeApp:              compositeapp,
		CompositeAppVersion:       compositeappversion,
		DeploymentIntentGroupName: dig,
	}

	var resp []TrafficGroupIntent
	values, err := db.DBconn.Find(v.db.storeName, key, v.db.tagMeta)
	if err != nil {
		return []TrafficGroupIntent{}, pkgerrors.Wrap(err, "db Find error")
	}

	for _, value := range values {
		tgi := TrafficGroupIntent{}
		err = db.DBconn.Unmarshal(value, &tgi)
		if err != nil {
			return []TrafficGroupIntent{}, pkgerrors.Wrap(err, "Unmarshalling Value")
		}
		resp = append(resp, tgi)
	}

	return resp, nil
}

// Delete the  TrafficGroupIntent from database
func (v *TrafficGroupIntentDbClient) DeleteTrafficGroupIntent(name, project, compositeapp, compositeappversion, dig string) error {

	//Construct key and tag to select the entry
	key := TrafficGroupIntentKey{
		TrafficGroupIntentName:    name,
		Project:                   project,
		CompositeApp:              compositeapp,
		CompositeAppVersion:       compositeappversion,
		DeploymentIntentGroupName: dig,
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
