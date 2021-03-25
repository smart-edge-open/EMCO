// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2021 Intel Corporation
package model

type Metadata struct {
	Name        string `json:"name" yaml:"name"`
	Description string `json:"description" yaml:"-"`
	UserData1   string `json:"userData1" yaml:"-"`
	UserData2   string `json:"userData2" yaml:"-"`
}

// SfcIntentKey is the key structure that is used in the database
type SfcClientIntentKey struct {
	Project             string `json:"project"`
	CompositeApp        string `json:"compositeapp"`
	CompositeAppVersion string `json:"compositeappversion"`
	DigName             string `json:"deploymentintentgroup"`
	NetControlIntent    string `json:"netcontrolintent"`
	SfcClientIntent     string `json:"sfcclientintent"`
}

// SfcIntent defines the high level structure of a network chain document
type SfcClientIntent struct {
	Metadata Metadata            `json:"metadata" yaml:"metadata"`
	Spec     SfcClientIntentSpec `json:"spec" yaml:"spec"`
}

// SfcIntentSpec contains the specification of a network chain
type SfcClientIntentSpec struct {
	ChainEnd                   string `json:"chainEnd"`
	ChainName                  string `json:"chainName"`
	ChainCompositeApp          string `json:"chainCompositeApp"`
	ChainCompositeAppVersion   string `json:"chainCompositeAppVersion"`
	ChainDeploymentIntentGroup string `json:"chainDeploymentIntentGroup"`
	ChainNetControlIntent      string `json:"chainNetControlIntent"`
	AppName                    string `json:"appName"`
	WorkloadResource           string `json:"workloadResource"`
	ResourceType               string `json:"resourceType"`
}
