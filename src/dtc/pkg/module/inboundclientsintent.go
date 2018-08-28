// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2020 Intel Corporation

package module

import (
	"strings"

	"github.com/open-ness/EMCO/src/orchestrator/pkg/infra/db"
	pkgerrors "github.com/pkg/errors"
)

type InboundClientsIntent struct {
	Metadata Metadata                 `json:"metadata"`
	Spec     InboundClientsIntentSpec `json:"spec"`
}

type InboundClientsIntentSpec struct {
	AppName     string   `json:"application"`
	ServiceName string   `json:"servicename"`
	Namespaces  []string `json:"namespaces"`
	IpRange     []string `json:"cidrs"`
}

type InboundClientsIntentManager interface {
	CreateClientsInboundIntent(tci InboundClientsIntent, project, compositeapp, compositeappversion, deploymentIntentGroupName, trafficIntentGroupName, inboundIntentName string, exists bool) (InboundClientsIntent, error)
	GetClientsInboundIntents(project, compositeapp, compositeappversion, deploymentIntentGroupName, trafficintentgroupname, inboundIntentName string) ([]InboundClientsIntent, error)
	GetClientsInboundIntent(name, project, compositeapp, compositeappversion, deploymentIntentGroupName, trafficintentgroupname, inboundIntentName string) (InboundClientsIntent, error)
	DeleteClientsInboundIntent(name, project, compositeapp, compositeappversion, deploymentintentgroupname, trafficintentgroupname, inboundserverintentname string) error
}

type InboundClientsIntentDbClient struct {
	db ClientDbInfo
}

// ClientsInboundIntentKey is the key structure that is used in the database
type InboundClientsIntentKey struct {
	Project                   string `json:"project"`
	CompositeApp              string `json:"compositeapp"`
	CompositeAppVersion       string `json:"compositeappversion"`
	DeploymentIntentGroupName string `json:"deploymentintentgroupname"`
	TrafficIntentGroupName    string `json:"trafficintentgroupname"`
	InboundServerIntentName   string `json:"inboundserverintentname"`
	InboundClientsIntentName  string `json:"inboundclientsintentname"`
}

func NewClientsInboundIntentClient() *InboundClientsIntentDbClient {
	return &InboundClientsIntentDbClient{
		db: ClientDbInfo{
			storeName: "orchestrator",
			tagMeta:   "inboundclientsintentmetadata",
		},
	}
}

func (v InboundClientsIntentDbClient) CreateClientsInboundIntent(ici InboundClientsIntent, project, compositeapp, compositeappversion, deploymentintentgroupname, trafficintentgroupname, inboundserverintentname string, exists bool) (InboundClientsIntent, error) {

	//Construct key and tag to select the entry
	key := InboundClientsIntentKey{
		Project:                   project,
		CompositeApp:              compositeapp,
		CompositeAppVersion:       compositeappversion,
		DeploymentIntentGroupName: deploymentintentgroupname,
		TrafficIntentGroupName:    trafficintentgroupname,
		InboundServerIntentName:   inboundserverintentname,
		InboundClientsIntentName:  ici.Metadata.Name,
	}

	//Check if the Inbound Server Intent exists
	_, err := NewServerInboundIntentClient().GetServerInboundIntent(inboundserverintentname, project, compositeapp, compositeappversion, deploymentintentgroupname, trafficintentgroupname)
	if err != nil {
		return InboundClientsIntent{}, pkgerrors.Errorf("Inbound Server Intent %v does not exist", inboundserverintentname)
	}

	//Check if this InboundClientsIntent already exists
	_, err = v.GetClientsInboundIntent(ici.Metadata.Name, project, compositeapp, compositeappversion, deploymentintentgroupname, trafficintentgroupname, inboundserverintentname)
	if err == nil && !exists {
		return InboundClientsIntent{}, pkgerrors.New("InboundClientsIntent already exists")
	}

	err = db.DBconn.Insert(v.db.storeName, key, nil, v.db.tagMeta, ici)
	if err != nil {
		return InboundClientsIntent{}, pkgerrors.Wrap(err, "Creating DB Entry")
	}

	return ici, nil

}

// GetClientsInboundIntent returns the InboundClientsIntent
func (v *InboundClientsIntentDbClient) GetClientsInboundIntent(name, project, compositeapp, compositeappversion, deploymentintentgroupname, trafficintentgroupname, inboundserverintentname string) (InboundClientsIntent, error) {

	//Construct key and tag to select the entry
	key := InboundClientsIntentKey{
		Project:                   project,
		CompositeApp:              compositeapp,
		CompositeAppVersion:       compositeappversion,
		DeploymentIntentGroupName: deploymentintentgroupname,
		TrafficIntentGroupName:    trafficintentgroupname,
		InboundServerIntentName:   inboundserverintentname,
		InboundClientsIntentName:  name,
	}

	value, err := db.DBconn.Find(v.db.storeName, key, v.db.tagMeta)
	if err != nil {
		return InboundClientsIntent{}, pkgerrors.Wrap(err, "db Find error")
	}

	//value is a byte array
	if value != nil {
		ici := InboundClientsIntent{}
		err = db.DBconn.Unmarshal(value[0], &ici)
		if err != nil {
			return InboundClientsIntent{}, pkgerrors.Wrap(err, "db Unmarshal error")
		}
		return ici, nil
	}

	return InboundClientsIntent{}, pkgerrors.New("Error getting InboundClientsIntent")
}

// GetClientsInboundIntents returns all of the InboundClientsIntent for corresponding name
func (v *InboundClientsIntentDbClient) GetClientsInboundIntents(project, compositeapp, compositeappversion, deploymentintentgroupname, trafficintentgroupname, inboundserverintentname string) ([]InboundClientsIntent, error) {

	//Construct key and tag to select the entry
	key := InboundClientsIntentKey{
		Project:                   project,
		CompositeApp:              compositeapp,
		CompositeAppVersion:       compositeappversion,
		DeploymentIntentGroupName: deploymentintentgroupname,
		TrafficIntentGroupName:    trafficintentgroupname,
		InboundServerIntentName:   inboundserverintentname,
		InboundClientsIntentName:  "",
	}

	var resp []InboundClientsIntent
	values, err := db.DBconn.Find(v.db.storeName, key, v.db.tagMeta)
	if err != nil {
		return []InboundClientsIntent{}, pkgerrors.Wrap(err, "Get InboundClientsIntent")
	}

	for _, value := range values {
		ici := InboundClientsIntent{}
		err = db.DBconn.Unmarshal(value, &ici)
		if err != nil {
			return []InboundClientsIntent{}, pkgerrors.Wrap(err, "Unmarshalling Value")
		}
		resp = append(resp, ici)
	}

	return resp, nil

}

// Delete the  ClientsInboundIntent from database
func (v *InboundClientsIntentDbClient) DeleteClientsInboundIntent(name, project, compositeapp, compositeappversion, deploymentintentgroupname, trafficintentgroupname, inboundserverintentname string) error {

	//Construct key and tag to select the entry
	key := InboundClientsIntentKey{
		Project:                   project,
		CompositeApp:              compositeapp,
		CompositeAppVersion:       compositeappversion,
		DeploymentIntentGroupName: deploymentintentgroupname,
		TrafficIntentGroupName:    trafficintentgroupname,
		InboundServerIntentName:   inboundserverintentname,
		InboundClientsIntentName:  name,
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
