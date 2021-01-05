package module

// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2020 Intel Corporation

import (
	"encoding/json"
	"strings"

	"github.com/open-ness/EMCO/src/orchestrator/pkg/infra/db"
	pkgerrors "github.com/pkg/errors"
)

// Customization consists of metadata and Spec
type Customization struct {
	Metadata Metadata      `json:"metadata"`
	Spec     CustomizeSpec `json:"spec"`
}

// CustomizeSpec consists of ClusterSpecific and ClusterInfo
type CustomizeSpec struct {
	ClusterSpecific string                   `json:"clusterspecific"`
	ClusterInfo     ClusterInfo              `json:"clusterinfo"`
	PatchType       string                   `json:"patchType,omitempty"`
	PatchJSON       []map[string]interface{} `json:"patchJson,omitempty"`
}

// ClusterInfo consists of scope, Clusterprovider, ClusterName, ClusterLabel and Mode
type ClusterInfo struct {
	Scope           string `json:"scope"`
	ClusterProvider string `json:"clusterprovider"`
	ClusterName     string `json:"clustername"`
	ClusterLabel    string `json:"clusterlabel"`
	Mode            string `json:"mode"`
}

// SpecFileContent contains the array of file contents
type SpecFileContent struct {
	FileContents []string
	FileNames    []string
}

// CustomizationKey consists of CustomizationName, project, CompApp, CompAppVersion, DeploymentIntentGroupName, GenericIntentName, ResourceName
type CustomizationKey struct {
	Customization       string `json:"customization"`
	Project             string `json:"project"`
	CompositeApp        string `json:"compositeapp"`
	CompositeAppVersion string `json:"compositeappversion"`
	DigName             string `json:"deploymentintentgroup"`
	GenericK8sIntent    string `json:"generick8sintent"`
	Resource            string `json:"resource"`
}

// CustomizationManager exposes all the functionalities of customization
type CustomizationManager interface {
	CreateCustomization(c Customization, t SpecFileContent, p, ca, cv, dig, gi, rs string, exists bool) (Customization, error)
	GetCustomization(c, p, ca, cv, dig, gi, rs string) (Customization, error)
	GetCustomizationContent(c, p, ca, cv, dig, gi, rs string) (SpecFileContent, error)
	GetAllCustomization(p, ca, cv, dig, gi, rs string) ([]Customization, error)
	DeleteCustomization(c, p, ca, cv, dig, gi, rs string) error
}

// CustomizationClientDbInfo consists of tableName and columns
type CustomizationClientDbInfo struct {
	storeName  string // name of the mongodb collection to use for client documents
	tagMeta    string // attribute key name for the json data of a client document
	tagContent string // attribute key name for the file data of a client document
	//tagFileName string // attribute key name for storing all the file names
}

// CustomizationClient consists of CustomizationClientDbInfo
type CustomizationClient struct {
	db CustomizationClientDbInfo
}

// We will use json marshalling to convert to string to
// preserve the underlying structure.
func (ck CustomizationKey) String() string {
	out, err := json.Marshal(ck)
	if err != nil {
		return ""
	}
	return string(out)
}

// NewCustomizationClient returns an instance of the CustomizationClient
func NewCustomizationClient() *CustomizationClient {
	return &CustomizationClient{
		db: CustomizationClientDbInfo{
			storeName:  "customization",
			tagMeta:    "customizationmetadata",
			tagContent: "customizationcontent",
			//tagFileName: "customizationfiles",
		},
	}
}

// CreateCustomization creates a new Customization
func (cc *CustomizationClient) CreateCustomization(c Customization, t SpecFileContent, p, ca, cv, dig, gi, rs string, exists bool) (Customization, error) {

	key := CustomizationKey{
		Customization:       c.Metadata.Name,
		Project:             p,
		CompositeApp:        ca,
		CompositeAppVersion: cv,
		DigName:             dig,
		GenericK8sIntent:    gi,
		Resource:            rs,
	}

	_, err := cc.GetCustomization(c.Metadata.Name, p, ca, cv, dig, gi, rs)
	if err == nil && !exists {
		return Customization{}, pkgerrors.New("Customization already exists")
	}

	err = db.DBconn.Insert(cc.db.storeName, key, nil, cc.db.tagMeta, c)
	if err != nil {
		return Customization{}, pkgerrors.Wrap(err, "Creating DB Entry")
	}

	err = db.DBconn.Insert(cc.db.storeName, key, nil, cc.db.tagContent, t)
	if err != nil {
		return Customization{}, pkgerrors.Wrap(err, "Creating DB Entry")
	}

	return c, nil

}

// GetCustomization returns Customization
func (cc *CustomizationClient) GetCustomization(c, p, ca, cv, dig, gi, rs string) (Customization, error) {

	key := CustomizationKey{
		Customization:       c,
		Project:             p,
		CompositeApp:        ca,
		CompositeAppVersion: cv,
		DigName:             dig,
		GenericK8sIntent:    gi,
		Resource:            rs,
	}

	value, err := db.DBconn.Find(cc.db.storeName, key, cc.db.tagMeta)
	if err != nil {
		return Customization{}, pkgerrors.Wrap(err, "db Find error")
	}

	//value is a byte array
	if value != nil {
		c := Customization{}
		err = db.DBconn.Unmarshal(value[0], &c)
		if err != nil {
			return Customization{}, pkgerrors.Wrap(err, "Unmarshalling Value")
		}
		return c, nil
	}

	return Customization{}, pkgerrors.New("Error getting Customization")

}

// GetAllCustomization returns all the customization objects
func (cc *CustomizationClient) GetAllCustomization(p, ca, cv, dig, gi, rs string) ([]Customization, error) {

	key := CustomizationKey{
		Customization:       "",
		Project:             p,
		CompositeApp:        ca,
		CompositeAppVersion: cv,
		DigName:             dig,
		GenericK8sIntent:    gi,
		Resource:            rs,
	}

	var czs []Customization
	values, err := db.DBconn.Find(cc.db.storeName, key, cc.db.tagMeta)
	if err != nil {
		return []Customization{}, pkgerrors.Wrap(err, "db Find error")
	}

	for _, value := range values {
		cz := Customization{}
		err = db.DBconn.Unmarshal(value, &cz)
		if err != nil {
			return []Customization{}, pkgerrors.Wrap(err, "Unmarshalling Value")
		}
		czs = append(czs, cz)
	}

	return czs, nil
}

// GetCustomizationContent returns the customizationContent
func (cc *CustomizationClient) GetCustomizationContent(c, p, ca, cv, dig, gi, rs string) (SpecFileContent, error) {
	key := CustomizationKey{
		Customization:       c,
		Project:             p,
		CompositeApp:        ca,
		CompositeAppVersion: cv,
		DigName:             dig,
		GenericK8sIntent:    gi,
		Resource:            rs,
	}

	value, err := db.DBconn.Find(cc.db.storeName, key, cc.db.tagContent)
	if err != nil {
		return SpecFileContent{}, pkgerrors.Wrap(err, "db Find error")
	}

	if value != nil {
		sFileContent := SpecFileContent{}

		err = db.DBconn.Unmarshal(value[0], &sFileContent)
		if err != nil {
			return SpecFileContent{}, pkgerrors.Wrap(err, "Unmarshalling Value")
		}
		return sFileContent, nil
	}

	return SpecFileContent{}, pkgerrors.New("Error getting CustomizationSpecFileContent")
}

// DeleteCustomization deletes Customization
func (cc *CustomizationClient) DeleteCustomization(c, p, ca, cv, dig, gi, rs string) error {

	key := CustomizationKey{
		Customization:       c,
		Project:             p,
		CompositeApp:        ca,
		CompositeAppVersion: cv,
		DigName:             dig,
		GenericK8sIntent:    gi,
		Resource:            rs,
	}

	err := db.DBconn.Remove(cc.db.storeName, key)
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
