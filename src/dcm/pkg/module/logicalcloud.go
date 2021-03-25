// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2020 Intel Corporation

package module

import (
	"encoding/json"

	"github.com/open-ness/EMCO/src/orchestrator/pkg/appcontext"
	"github.com/open-ness/EMCO/src/orchestrator/pkg/infra/db"
	log "github.com/open-ness/EMCO/src/orchestrator/pkg/infra/logutils"
	"github.com/open-ness/EMCO/src/orchestrator/pkg/module"

	pkgerrors "github.com/pkg/errors"
)

// LogicalCloud contains the parameters needed for a Logical Cloud
type LogicalCloud struct {
	MetaData      MetaDataList `json:"metadata"`
	Specification Spec         `json:"spec"`
}

// MetaData contains the parameters needed for metadata
type MetaDataList struct {
	LogicalCloudName string `json:"name"`
	Description      string `json:"description"`
	UserData1        string `json:"userData1"`
	UserData2        string `json:"userData2"`
}

// Spec contains the parameters needed for spec
type Spec struct {
	NameSpace string   `json:"namespace"`
	Level     string   `json:"level"`
	User      UserData `json:"user"`
}

// UserData contains the parameters needed for user
type UserData struct {
	UserName string `json:"user-name"`
	Type     string `json:"type"`
}

// LogicalCloudKey is the key structure that is used in the database
type LogicalCloudKey struct {
	Project          string `json:"project"`
	LogicalCloudName string `json:"logical-cloud-name"`
}

// AppContextKey is an alternative key to access logical clouds
type AppContextKey struct {
	LCContext string `json:"lccontext"`
}

// LogicalCloudManager is an interface that exposes the connection
// functionality
type LogicalCloudManager interface {
	Create(project string, c LogicalCloud) (LogicalCloud, error)
	Get(project, name string) (LogicalCloud, error)
	GetAll(project string) ([]LogicalCloud, error)
	Delete(project, name string) error
	Update(project, name string, c LogicalCloud) (LogicalCloud, error)
}

// LogicalCloudClient implements the LogicalCloudManager
// It will also be used to maintain some localized state
type LogicalCloudClient struct {
	storeName  string
	tagMeta    string
	tagContext string
}

// LogicalCloudClient returns an instance of the LogicalCloudClient
// which implements the LogicalCloudManager
func NewLogicalCloudClient() *LogicalCloudClient {
	return &LogicalCloudClient{
		storeName:  "orchestrator",
		tagMeta:    "logicalcloud",
		tagContext: "lccontext",
	}
}

// Create entry for the logical cloud resource in the database
func (v *LogicalCloudClient) Create(project string, c LogicalCloud) (LogicalCloud, error) {

	//Construct key consisting of name
	key := LogicalCloudKey{
		Project:          project,
		LogicalCloudName: c.MetaData.LogicalCloudName,
	}

	//Check if project exists
	err := CheckProject(project)
	if err != nil {
		return LogicalCloud{}, pkgerrors.Wrap(err, "Unable to find the project")
	}

	//Check if this Logical Cloud already exists
	_, err = v.Get(project, c.MetaData.LogicalCloudName)
	if err == nil {
		return LogicalCloud{}, pkgerrors.Wrap(err, "Logical Cloud already exists")
	}

	// if Logical Cloud Level is not specified, it defaults to 1:
	if c.Specification.Level == "" {
		c.Specification.Level = "1"
	}

	err = db.DBconn.Insert(v.storeName, key, nil, v.tagMeta, c)
	if err != nil {
		return LogicalCloud{}, pkgerrors.Wrap(err, "Error creating DB Entry")
	}

	return c, nil
}

// Get returns Logical Cloud corresponding to logical cloud name
func (v *LogicalCloudClient) Get(project, logicalCloudName string) (LogicalCloud, error) {

	//Construct the composite key to select the entry
	key := LogicalCloudKey{
		Project:          project,
		LogicalCloudName: logicalCloudName,
	}
	value, err := db.DBconn.Find(v.storeName, key, v.tagMeta)
	if err != nil {
		return LogicalCloud{}, pkgerrors.Wrap(err, "Error getting Logical Cloud")
	}

	//value is a byte array
	if value != nil {
		lc := LogicalCloud{}
		err = db.DBconn.Unmarshal(value[0], &lc)
		if err != nil {
			return LogicalCloud{}, pkgerrors.Wrap(err, "Error unmarshaling value")
		}
		return lc, nil
	}

	return LogicalCloud{}, pkgerrors.New("Logical Cloud does not exist")
}

// GetAll returns Logical Clouds in the project
func (v *LogicalCloudClient) GetAll(project string) ([]LogicalCloud, error) {

	//Construct the composite key to select the entry
	key := LogicalCloudKey{
		Project:          project,
		LogicalCloudName: "",
	}

	var resp []LogicalCloud
	values, err := db.DBconn.Find(v.storeName, key, v.tagMeta)
	if err != nil {
		return []LogicalCloud{}, pkgerrors.Wrap(err, "Error getting Logical Clouds")
	}

	for _, value := range values {
		lc := LogicalCloud{}
		err = db.DBconn.Unmarshal(value, &lc)
		if err != nil {
			return []LogicalCloud{}, pkgerrors.Wrap(err, "Unmarshaling values")
		}
		resp = append(resp, lc)
	}

	return resp, nil
}

// Delete the Logical Cloud entry from database
func (v *LogicalCloudClient) Delete(project, logicalCloudName string) error {

	//Construct the composite key to select the entry
	key := LogicalCloudKey{
		Project:          project,
		LogicalCloudName: logicalCloudName,
	}
	//Check if this Logical Cloud exists
	_, err := v.Get(project, logicalCloudName)
	if err != nil {
		return pkgerrors.New("Logical Cloud does not exist")
	}

	context, _, err := GetLogicalCloudContext(v.storeName, key, v.tagContext, project, logicalCloudName)
	// If there's no context for Logical Cloud, just go ahead and delete it now
	if err != nil {
		err = db.DBconn.Remove(v.storeName, key)
		if err != nil {
			return pkgerrors.Wrap(err, "Error when deleting Logical Cloud (scenario with no context)")
		}
		return nil
	}

	// Make sure rsync status for this logical cloud is Terminated,
	// otherwise we can't remove appcontext yet
	acStatus, err := GetAppContextStatus(context)
	if err != nil {
		return err
	}
	switch acStatus.Status {
	case appcontext.AppContextStatusEnum.Terminating:
		log.Error("The Logical Cloud can't be deleted yet, it is being terminated", log.Fields{"logicalcloud": logicalCloudName})
		return pkgerrors.New("The Logical Cloud can't be deleted yet, it is being terminated")
	case appcontext.AppContextStatusEnum.Instantiated:
		log.Error("The Logical Cloud is instantiated, please terminate first", log.Fields{"logicalcloud": logicalCloudName})
		return pkgerrors.New("The Logical Cloud is instantiated, please terminate first")
	case appcontext.AppContextStatusEnum.Instantiating:
		log.Error("The Logical Cloud is instantiating, please wait and then terminate", log.Fields{"logicalcloud": logicalCloudName})
		return pkgerrors.New("The Logical Cloud is instantiating, please wait and then terminate")
	case appcontext.AppContextStatusEnum.InstantiateFailed:
		log.Error("The Logical Cloud has failed instantiating, for safety please terminate and try again", log.Fields{"logicalcloud": logicalCloudName})
		return pkgerrors.New("The Logical Cloud has failed instantiating, for safety please terminate and try again")
	case appcontext.AppContextStatusEnum.TerminateFailed:
		log.Info("The Logical Cloud has failed terminating, proceeding with the delete operation", log.Fields{"logicalcloud": logicalCloudName})
		// try to delete anyway since termination failed
		fallthrough
	case appcontext.AppContextStatusEnum.Terminated:
		// remove the appcontext
		err := context.DeleteCompositeApp()
		if err != nil {
			log.Error("Error deleting AppContext CompositeApp Logical Cloud", log.Fields{"logicalcloud": logicalCloudName})
			return pkgerrors.Wrap(err, "Error deleting AppContext CompositeApp Logical Cloud")
		}

		err = db.DBconn.Remove(v.storeName, key)
		if err != nil {
			log.Error("Error when deleting Logical Cloud (scenario with Terminated status)", log.Fields{"logicalcloud": logicalCloudName})
			return pkgerrors.Wrap(err, "Error when deleting Logical Cloud (scenario with Terminated status)")
		}
		log.Info("Deleted Logical Cloud", log.Fields{"logicalcloud": logicalCloudName})
		return nil
	default:
		log.Error("The Logical Cloud isn't in an expected status so not taking any action", log.Fields{"logicalcloud": logicalCloudName, "status": acStatus.Status})
		return pkgerrors.New("The Logical Cloud isn't in an expected status so not taking any action")
	}
}

// Update an entry for the Logical Cloud in the database
func (v *LogicalCloudClient) Update(project, logicalCloudName string, c LogicalCloud) (LogicalCloud, error) {

	key := LogicalCloudKey{
		Project:          project,
		LogicalCloudName: logicalCloudName,
	}
	// Check for mismatch, logicalCloudName and payload logical cloud name
	if c.MetaData.LogicalCloudName != logicalCloudName {
		return LogicalCloud{}, pkgerrors.New("Logical Cloud name mismatch")
	}
	//Check if this Logical Cloud exists
	_, err := v.Get(project, logicalCloudName)
	if err != nil {
		return LogicalCloud{}, pkgerrors.New("Logical Cloud does not exist")
	}
	err = db.DBconn.Insert(v.storeName, key, nil, v.tagMeta, c)
	if err != nil {
		return LogicalCloud{}, pkgerrors.Wrap(err, "Updating DB Entry")
	}
	return c, nil
}

// GetLogicalCloudContext returns the AppContext for corresponding provider and name
func GetLogicalCloudContext(storeName string, key db.Key, meta string, project string, name string) (appcontext.AppContext, string, error) {

	value, err := db.DBconn.Find(storeName, key, meta)
	if err != nil {
		return appcontext.AppContext{}, "", pkgerrors.Wrap(err, "Get Logical Cloud Context")
	}

	//value is a [][]byte
	if value != nil {
		ctxVal := string(value[0])
		var lcc appcontext.AppContext
		_, err = lcc.LoadAppContext(ctxVal)
		if err != nil {
			return appcontext.AppContext{}, "", pkgerrors.Wrap(err, "Reinitializing Logical Cloud AppContext")
		}
		return lcc, ctxVal, nil
	}

	return appcontext.AppContext{}, "", pkgerrors.New("Error getting Logical Cloud AppContext")
}

// GetLogicalCloudFromContext returns the pair (project, logical cloud name) for a given AppContext
func GetLogicalCloudFromContext(storeName string, appContextID string) (string, string, error) {
	key := AppContextKey{
		LCContext: appContextID,
	}
	log.Info("GetLogicalCloudFromContext", log.Fields{"appContextID": appContextID})

	values, err := db.DBconn.Find(storeName, key, "logical-cloud-name")
	if err != nil {
		log.Error("Couldn't fetch logical cloud", log.Fields{"err": err})
		return "", "", pkgerrors.Wrap(err, "Couldn't fetch logical cloud")
	}
	logicalCloudName := string(values[0])
	log.Info("", log.Fields{"logicalCloudName": logicalCloudName})

	values, err = db.DBconn.Find(storeName, key, "project")
	if err != nil {
		log.Error("Couldn't fetch project", log.Fields{"err": err})
		return "", "", pkgerrors.Wrap(err, "Couldn't fetch project")
	}
	project := string(values[0])
	log.Info("", log.Fields{"project": project})

	return project, logicalCloudName, nil
}

// CheckProject if the project exists
func CheckProject(project string) error {
	_, err := module.NewProjectClient().GetProject(project)
	if err != nil {
		return pkgerrors.New("Unable to find the project")
	}

	return nil
}

// CheckLogicalCloud checks if logical cloud exists
func CheckLogicalCloud(lcClient LogicalCloudManager, project string, logicalCloud string) error {
	_, err := lcClient.Get(project, logicalCloud)
	if err != nil {
		return pkgerrors.New("Unable to find the logical cloud")
	}

	return nil
}

// GetAppContextStatus returns the Status for a particular AppContext
func GetAppContextStatus(ac appcontext.AppContext) (*appcontext.AppContextStatus, error) {

	h, err := ac.GetCompositeAppHandle()
	if err != nil {
		return nil, err
	}
	sh, err := ac.GetLevelHandle(h, "status")
	if err != nil {
		return nil, err
	}
	s, err := ac.GetValue(sh)
	if err != nil {
		return nil, err
	}
	acStatus := appcontext.AppContextStatus{}
	js, _ := json.Marshal(s)
	json.Unmarshal(js, &acStatus)

	return &acStatus, nil
}
