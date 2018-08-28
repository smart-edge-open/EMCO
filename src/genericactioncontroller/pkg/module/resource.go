package module

// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2020 Intel Corporation

import (
	"strings"

	"github.com/open-ness/EMCO/src/orchestrator/pkg/infra/db"
	pkgerrors "github.com/pkg/errors"
)

// Resource consists of metadata and Spec
type Resource struct {
	Metadata Metadata     `json:"metadata"`
	Spec     ResourceSpec `json:"spec"`
}

// ResourceSpec consists of AppName, NewObject, ExistingResource
type ResourceSpec struct {
	AppName     string      `json:"appname"`
	NewObject   string      `json:"newobject"`
	ResourceGVK ResourceGVK `json:"resourcegvk,omitempty"`
}

// ResourceGVK consists of ApiVersion, Kind, Name
type ResourceGVK struct {
	APIVersion string `json:"apiversion"`
	Kind       string `json:"kind"`
	Name       string `json:"name"`
}

// ResourceFileContent contains the content of resourceTemplate
type ResourceFileContent struct {
	FileContent string `json:"filecontent"`
}

// ResourceKey consists of resourceName, ProjectName, CompAppName, CompAppVersion, DeploymentIntentgroupName, GenericK8sIntentName
type ResourceKey struct {
	Resource            string `json:"resource"`
	Project             string `json:"project"`
	CompositeApp        string `json:"compositeapp"`
	CompositeAppVersion string `json:"compositeappversion"`
	DigName             string `json:"deploymentintentgroup"`
	GenericK8sIntent    string `json:"generick8sintent"`
}

// ResourceManager is an interface that exposes resource related functionalities
type ResourceManager interface {
	CreateResource(b Resource, t ResourceFileContent, p, ca, cv, dig, gi string, exists bool) (Resource, error)
	GetResource(name, p, ca, cv, dig, gi string) (Resource, error)
	GetResourceContent(brName, p, ca, cv, dig, gi string) (ResourceFileContent, error)
	GetAllResources(p, ca, cv, dig, gi string) ([]Resource, error)
	DeleteResource(brName, p, ca, cv, dig, gi string) error
}

type clientDbInfo struct {
	storeName  string // name of the mongodb collection to use for client documents
	tagMeta    string // attribute key name for the json data of a client document
	tagContent string // attribute key name for the file data of a client document
}

// ResourceClient implements the resourceManager
type ResourceClient struct {
	db clientDbInfo
}

// NewResourceClient returns an instance of the resourceClient
// which implements the Manager
func NewResourceClient() *ResourceClient {
	return &ResourceClient{
		db: clientDbInfo{
			storeName:  "resource",
			tagMeta:    "resourcemetadata",
			tagContent: "resourcecontent",
		},
	}
}

// CreateResource creates a resource
func (rc *ResourceClient) CreateResource(r Resource, t ResourceFileContent, p, ca, cv, dig, gi string, exists bool) (Resource, error) {

	key := ResourceKey{
		Resource:            r.Metadata.Name,
		Project:             p,
		CompositeApp:        ca,
		CompositeAppVersion: cv,
		DigName:             dig,
		GenericK8sIntent:    gi,
	}

	_, err := rc.GetResource(r.Metadata.Name, p, ca, cv, dig, gi)
	if err == nil && !exists {
		return Resource{}, pkgerrors.New("resource already exists")
	}
	err = db.DBconn.Insert(rc.db.storeName, key, nil, rc.db.tagMeta, r)
	if err != nil {
		return Resource{}, pkgerrors.Wrap(err, "Creating DB Entry")
	}

	err = db.DBconn.Insert(rc.db.storeName, key, nil, rc.db.tagContent, t)
	if err != nil {
		return Resource{}, pkgerrors.Wrap(err, "Creating DB Entry")
	}
	return r, nil
}

// GetResource returns a resource
func (rc *ResourceClient) GetResource(brName, p, ca, cv, dig, gi string) (Resource, error) {

	key := ResourceKey{
		Resource:            brName,
		Project:             p,
		CompositeApp:        ca,
		CompositeAppVersion: cv,
		DigName:             dig,
		GenericK8sIntent:    gi,
	}

	value, err := db.DBconn.Find(rc.db.storeName, key, rc.db.tagMeta)
	if err != nil {
		return Resource{}, pkgerrors.Wrap(err, "db Find error")
	}

	//value is a byte array
	if value != nil {
		br := Resource{}
		err = db.DBconn.Unmarshal(value[0], &br)
		if err != nil {
			return Resource{}, pkgerrors.Wrap(err, "Unmarshalling Value")
		}
		return br, nil
	}

	return Resource{}, pkgerrors.New("Error getting resource")
}

// GetAllResources shall return all the resources for the intent
func (rc *ResourceClient) GetAllResources(p, ca, cv, dig, gi string) ([]Resource, error) {

	key := ResourceKey{
		Resource:            "",
		Project:             p,
		CompositeApp:        ca,
		CompositeAppVersion: cv,
		DigName:             dig,
		GenericK8sIntent:    gi,
	}

	var brs []Resource
	values, err := db.DBconn.Find(rc.db.storeName, key, rc.db.tagMeta)
	if err != nil {
		return []Resource{}, pkgerrors.Wrap(err, "db Find error")
	}

	for _, value := range values {
		br := Resource{}
		err = db.DBconn.Unmarshal(value, &br)
		if err != nil {
			return []Resource{}, pkgerrors.Wrap(err, "Unmarshalling Value")
		}
		brs = append(brs, br)
	}

	return brs, nil
}

// GetResourceContent returns the content of the resourceTemplate
func (rc *ResourceClient) GetResourceContent(rName, p, ca, cv, dig, gi string) (ResourceFileContent, error) {
	key := ResourceKey{
		Resource:            rName,
		Project:             p,
		CompositeApp:        ca,
		CompositeAppVersion: cv,
		DigName:             dig,
		GenericK8sIntent:    gi,
	}

	value, err := db.DBconn.Find(rc.db.storeName, key, rc.db.tagContent)
	if err != nil {
		return ResourceFileContent{}, pkgerrors.Wrap(err, "db Find error")
	}

	if value != nil {
		rfc := ResourceFileContent{}
		err = db.DBconn.Unmarshal(value[0], &rfc)
		if err != nil {
			return ResourceFileContent{}, pkgerrors.Wrap(err, "Unmarshalling Value")
		}
		return rfc, nil
	}

	return ResourceFileContent{}, pkgerrors.New("Error getting resourceTemplateContent")
}

// DeleteResource deletes a given resource
func (rc *ResourceClient) DeleteResource(rName, p, ca, cv, dig, gi string) error {
	key := ResourceKey{
		Resource:            rName,
		Project:             p,
		CompositeApp:        ca,
		CompositeAppVersion: cv,
		DigName:             dig,
		GenericK8sIntent:    gi,
	}

	err := db.DBconn.Remove(rc.db.storeName, key)
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
