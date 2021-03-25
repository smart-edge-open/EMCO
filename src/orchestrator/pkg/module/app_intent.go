// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2020 Intel Corporation

package module

/*
This file deals with the backend implementation of
Adding/Querying AppIntents for each application in the composite-app
*/

import (
	"encoding/json"
	"reflect"
	"strings"

	gpic "github.com/open-ness/EMCO/src/orchestrator/pkg/gpic"
	"github.com/open-ness/EMCO/src/orchestrator/pkg/infra/db"
	pkgerrors "github.com/pkg/errors"
)

// AppIntent has two components - metadata, spec
type AppIntent struct {
	MetaData MetaData `json:"metadata,omitempty"`
	Spec     SpecData `json:"spec,omitempty"`
}

// MetaData has - name, description, userdata1, userdata2
type MetaData struct {
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
	UserData1   string `json:"userData1,omitempty"`
	UserData2   string `json:"userData2,omitempty"`
}

// SpecData consists of appName and intent
type SpecData struct {
	AppName string           `json:"app-name,omitempty"`
	Intent  gpic.IntentStruc `json:"intent,omitempty"`
}

// AppIntentManager is an interface which exposes the
// AppIntentManager functionalities
type AppIntentManager interface {
	CreateAppIntent(a AppIntent, p string, ca string, v string, i string, digName string) (AppIntent, error)
	GetAppIntent(ai string, p string, ca string, v string, i string, digName string) (AppIntent, error)
	GetAllIntentsByApp(aN, p, ca, v, i, digName string) (SpecData, error)
	GetAllAppIntents(p, ca, v, i, digName string) ([]AppIntent, error)
	DeleteAppIntent(ai string, p string, ca string, v string, i string, digName string) error
}

//AppIntentQueryKey required for query
type AppIntentQueryKey struct {
	AppName string `json:"app-name"`
}

// AppIntentKey is used as primary key
type AppIntentKey struct {
	Name                      string `json:"appintent"`
	Project                   string `json:"project"`
	CompositeApp              string `json:"compositeapp"`
	Version                   string `json:"compositeappversion"`
	Intent                    string `json:"genericplacement"`
	DeploymentIntentGroupName string `json:"deploymentintentgroup"`
}

// AppIntentFindByAppKey required for query
type AppIntentFindByAppKey struct {
	Project                   string `json:"project"`
	CompositeApp              string `json:"compositeapp"`
	CompositeAppVersion       string `json:"compositeappversion"`
	Intent                    string `json:"genericplacement"`
	DeploymentIntentGroupName string `json:"deploymentintentgroup"`
	AppName                   string `json:"app-name"`
}

// ApplicationsAndClusterInfo type represents the list of
type ApplicationsAndClusterInfo struct {
	ArrayOfAppClusterInfo []AppClusterInfo `json:"applications"`
}

// AppClusterInfo is a type linking the app and the clusters
// on which they need to be installed.
type AppClusterInfo struct {
	Name       string       `json:"name"`
	AllOfArray []gpic.AllOf `json:"allOf,omitempty"`
	AnyOfArray []gpic.AnyOf `json:"anyOf,omitempty"`
}

// We will use json marshalling to convert to string to
// preserve the underlying structure.
func (ak AppIntentKey) String() string {
	out, err := json.Marshal(ak)
	if err != nil {
		return ""
	}
	return string(out)
}

// AppIntentClient implements the AppIntentManager interface
type AppIntentClient struct {
	storeName   string
	tagMetaData string
}

// NewAppIntentClient returns an instance of AppIntentClient
func NewAppIntentClient() *AppIntentClient {
	return &AppIntentClient{
		storeName:   "orchestrator",
		tagMetaData: "appintentmetadata",
	}
}

// CreateAppIntent creates an entry for AppIntent in the db.
// Other input parameters for it - projectName, compositeAppName, version, intentName and deploymentIntentGroupName.
func (c *AppIntentClient) CreateAppIntent(a AppIntent, p string, ca string, v string, i string, digName string) (AppIntent, error) {

	//Check for the AppIntent already exists here.
	res, err := c.GetAppIntent(a.MetaData.Name, p, ca, v, i, digName)
	if !reflect.DeepEqual(res, AppIntent{}) {
		return AppIntent{}, pkgerrors.New("AppIntent already exists")
	}

	//Check if project exists
	_, err = NewProjectClient().GetProject(p)
	if err != nil {
		return AppIntent{}, pkgerrors.New("Unable to find the project")
	}

	// check if compositeApp exists
	_, err = NewCompositeAppClient().GetCompositeApp(ca, v, p)
	if err != nil {
		return AppIntent{}, pkgerrors.New("Unable to find the composite-app")
	}

	// check if Intent exists
	_, err = NewGenericPlacementIntentClient().GetGenericPlacementIntent(i, p, ca, v, digName)
	if err != nil {
		return AppIntent{}, pkgerrors.New("Unable to find the intent")
	}

	// check if the deploymentIntentGrpName exists
	_, err = NewDeploymentIntentGroupClient().GetDeploymentIntentGroup(digName, p, ca, v)
	if err != nil {
		return AppIntent{}, pkgerrors.New("Unable to find the deployment-intent-group-name")
	}

	akey := AppIntentKey{
		Name:                      a.MetaData.Name,
		Project:                   p,
		CompositeApp:              ca,
		Version:                   v,
		Intent:                    i,
		DeploymentIntentGroupName: digName,
	}

	qkey := AppIntentQueryKey{
		AppName: a.Spec.AppName,
	}

	err = db.DBconn.Insert(c.storeName, akey, qkey, c.tagMetaData, a)
	if err != nil {
		return AppIntent{}, pkgerrors.Wrap(err, "Create DB entry error")
	}

	return a, nil
}

// GetAppIntent shall take arguments - name of the app intent, name of the project, name of the composite app, version of the composite app,intent name and deploymentIntentGroupName. It shall return the AppIntent
func (c *AppIntentClient) GetAppIntent(ai string, p string, ca string, v string, i string, digName string) (AppIntent, error) {

	k := AppIntentKey{
		Name:                      ai,
		Project:                   p,
		CompositeApp:              ca,
		Version:                   v,
		Intent:                    i,
		DeploymentIntentGroupName: digName,
	}

	result, err := db.DBconn.Find(c.storeName, k, c.tagMetaData)
	if err != nil {
		return AppIntent{}, pkgerrors.Wrap(err, "db Find error")
	}

	if result != nil {
		a := AppIntent{}
		err = db.DBconn.Unmarshal(result[0], &a)
		if err != nil {
			return AppIntent{}, pkgerrors.Wrap(err, "Unmarshalling  AppIntent")
		}
		return a, nil

	}
	return AppIntent{}, pkgerrors.New("Error getting AppIntent")
}

/*
GetAllIntentsByApp queries intent by AppName, it takes in parameters AppName, CompositeAppName, CompositeNameVersion,
GenericPlacementIntentName & DeploymentIntentGroupName. Returns SpecData which contains
all the intents for the app.
*/
func (c *AppIntentClient) GetAllIntentsByApp(aN, p, ca, v, i, digName string) (SpecData, error) {
	k := AppIntentFindByAppKey{
		Project:                   p,
		CompositeApp:              ca,
		CompositeAppVersion:       v,
		Intent:                    i,
		DeploymentIntentGroupName: digName,
		AppName:                   aN,
	}
	result, err := db.DBconn.Find(c.storeName, k, c.tagMetaData)
	if err != nil {
		return SpecData{}, pkgerrors.Wrap(err, "db Find error")
	}
	if len(result) == 0 {
		return SpecData{}, nil
	}
	
	var a AppIntent
	err = db.DBconn.Unmarshal(result[0], &a)
	if err != nil {
		return SpecData{}, pkgerrors.Wrap(err, "Unmarshalling  AppIntent")
	}
	return a.Spec, nil

}

/*
GetAllAppIntents takes in paramaters ProjectName, CompositeAppName, CompositeNameVersion
and GenericPlacementIntentName,DeploymentIntentGroupName. Returns an array of AppIntents
*/
func (c *AppIntentClient) GetAllAppIntents(p, ca, v, i, digName string) ([]AppIntent, error) {
	k := AppIntentKey{
		Name:                      "",
		Project:                   p,
		CompositeApp:              ca,
		Version:                   v,
		Intent:                    i,
		DeploymentIntentGroupName: digName,
	}
	result, err := db.DBconn.Find(c.storeName, k, c.tagMetaData)
	if err != nil {
		return []AppIntent{}, pkgerrors.Wrap(err, "db Find error")
	}

	var appIntents []AppIntent

	if len(result) != 0 {
		for i := range result {
			aI := AppIntent{}
			err = db.DBconn.Unmarshal(result[i], &aI)
			if err != nil {
				return []AppIntent{}, pkgerrors.Wrap(err, "Unmarshalling  AppIntent")
			}
			appIntents = append(appIntents, aI)
		}
	}

	return appIntents, err
}

// DeleteAppIntent delete an AppIntent
func (c *AppIntentClient) DeleteAppIntent(ai string, p string, ca string, v string, i string, digName string) error {
	k := AppIntentKey{
		Name:                      ai,
		Project:                   p,
		CompositeApp:              ca,
		Version:                   v,
		Intent:                    i,
		DeploymentIntentGroupName: digName,
	}

	err := db.DBconn.Remove(c.storeName, k)
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
