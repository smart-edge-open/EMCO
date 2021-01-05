// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2020 Intel Corporation

package module

import (
	"strings"

	"github.com/open-ness/EMCO/src/orchestrator/pkg/infra/db"
	pkgerrors "github.com/pkg/errors"
)

type InboundServerIntent struct {
	Metadata Metadata               `json:"metadata"`
	Spec     InbondServerIntentSpec `json:"spec"`
}

type InbondServerIntentSpec struct {
	AppName         string `json:"appName"`
	AppLabel        string `json:"appLabel"`
	ServiceName     string `json:"serviceName"`
	ExternalName    string `json:"externalName", default:""`
	Port            int    `json:"port"`
	Protocol        string `json:"protocol"`
	ExternalSupport bool   `json:"externalSupport", default:false`
}

type InboundServerIntentManager interface {
	CreateServerInboundIntent(isi InboundServerIntent, project, compositeapp, compositeappversion, deploymentIntentGroupName, trafficintentgroupname string, exists bool) (InboundServerIntent, error)
	GetServerInboundIntent(name, project, compositeapp, compositeappversion, dig, trafficintentgroupname string) (InboundServerIntent, error)

	GetServerInboundIntents(project, compositeapp, compositeappversion, dig, intentName string) ([]InboundServerIntent, error)
	DeleteServerInboundIntent(name, project, compositeapp, compositeappversion, dig, trafficintentgroupname string) error
}

type InboundServerIntentDbClient struct {
	db ClientDbInfo
}

// ServerInboundIntentKey is the key structure that is used in the database
type InboundServerIntentKey struct {
	Project                   string `json:"project"`
	CompositeApp              string `json:"compositeapp"`
	CompositeAppVersion       string `json:"compositeappversion"`
	DeploymentIntentGroupName string `json:"deploymentintentgroupname"`
	TrafficIntentGroupName    string `json:"trafficintentgroupname"`
	ServerInboundIntentName   string `json:"inboundserverintentname"`
}

func NewServerInboundIntentClient() *InboundServerIntentDbClient {
	return &InboundServerIntentDbClient{
		db: ClientDbInfo{
			storeName: "orchestrator",
			tagMeta:   "inboundserverintentmetadata",
		},
	}
}

func (v InboundServerIntentDbClient) CreateServerInboundIntent(isi InboundServerIntent, project, compositeapp, compositeappversion, deploymentintentgroupname, trafficintentgroupname string, exists bool) (InboundServerIntent, error) {

	//Construct key and tag to select the entry
	key := InboundServerIntentKey{
		Project:                   project,
		CompositeApp:              compositeapp,
		CompositeAppVersion:       compositeappversion,
		DeploymentIntentGroupName: deploymentintentgroupname,
		TrafficIntentGroupName:    trafficintentgroupname,
		ServerInboundIntentName:   isi.Metadata.Name,
	}

	//Check if the Traffic Group Intent exists
	_, err := NewTrafficGroupIntentClient().GetTrafficGroupIntent(trafficintentgroupname, project, compositeapp, compositeappversion, deploymentintentgroupname)
	if err != nil {
		return InboundServerIntent{}, pkgerrors.Errorf("Traffic Group Intent %v does not exist", trafficintentgroupname)
	}

	//Check if this ServerInboundIntent already exists
	_, err = v.GetServerInboundIntent(isi.Metadata.Name, project, compositeapp, compositeappversion, deploymentintentgroupname, trafficintentgroupname)
	if err == nil && !exists {
		return InboundServerIntent{}, pkgerrors.New("ServerInboundIntent already exists")
	}

	err = db.DBconn.Insert(v.db.storeName, key, nil, v.db.tagMeta, isi)
	if err != nil {
		return InboundServerIntent{}, pkgerrors.Wrap(err, "Creating DB Entry")
	}

	return isi, nil
}

// GetServerInboundIntent returns the ServerInboundIntent for corresponding name
func (v *InboundServerIntentDbClient) GetServerInboundIntent(name, project, compositeapp, compositeappversion, dig, trafficintentgroupname string) (InboundServerIntent, error) {

	//Construct key and tag to select the entry
	key := InboundServerIntentKey{
		Project:                   project,
		CompositeApp:              compositeapp,
		CompositeAppVersion:       compositeappversion,
		DeploymentIntentGroupName: dig,
		TrafficIntentGroupName:    trafficintentgroupname,
		ServerInboundIntentName:   name,
	}

	value, err := db.DBconn.Find(v.db.storeName, key, v.db.tagMeta)
	if err != nil {
		return InboundServerIntent{}, pkgerrors.Wrap(err, "db Find error")
	}

	//value is a byte array
	if value != nil {
		wi := InboundServerIntent{}
		err = db.DBconn.Unmarshal(value[0], &wi)
		if err != nil {
			return InboundServerIntent{}, pkgerrors.Wrap(err, "db Unmarshal error")
		}
		return wi, nil
	}

	return InboundServerIntent{}, pkgerrors.New("Error getting ServerInboundIntent")
}

// GetServerInboundIntents returns all of the ServerInboundIntents
func (v *InboundServerIntentDbClient) GetServerInboundIntents(project, compositeapp, compositeappversion, deploymentintentgroupname, trafficintentgroupname string) ([]InboundServerIntent, error) {

	//Construct key and tag to select the entry
	key := InboundServerIntentKey{
		Project:                   project,
		CompositeApp:              compositeapp,
		CompositeAppVersion:       compositeappversion,
		DeploymentIntentGroupName: deploymentintentgroupname,
		TrafficIntentGroupName:    trafficintentgroupname,
		ServerInboundIntentName:   "",
	}

	var resp []InboundServerIntent
	values, err := db.DBconn.Find(v.db.storeName, key, v.db.tagMeta)
	if err != nil {
		return []InboundServerIntent{}, pkgerrors.Wrap(err, "Get ServerInboundIntents")
	}

	for _, value := range values {
		is := InboundServerIntent{}
		err = db.DBconn.Unmarshal(value, &is)
		if err != nil {
			return []InboundServerIntent{}, pkgerrors.Wrap(err, "Unmarshalling Value")
		}
		resp = append(resp, is)
	}

	return resp, nil
}

// Delete the  ServerInboundIntents from database
func (v *InboundServerIntentDbClient) DeleteServerInboundIntent(name, project, compositeapp, compositeappversion, dig, trafficintentgroupname string) error {

	//Construct key and tag to select the entry
	key := InboundServerIntentKey{
		Project:                   project,
		CompositeApp:              compositeapp,
		CompositeAppVersion:       compositeappversion,
		DeploymentIntentGroupName: dig,
		TrafficIntentGroupName:    trafficintentgroupname,
		ServerInboundIntentName:   name,
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
