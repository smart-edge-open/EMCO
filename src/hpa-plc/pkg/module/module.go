// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2020 Intel Corporation

package module

import (
	"encoding/json"

	hpaModel "github.com/open-ness/EMCO/src/hpa-plc/pkg/model"
)

// HpaPlacementManager .. Manager is an interface exposing the HpaPlacementIntent functionality
type HpaPlacementManager interface {
	// intents
	AddIntent(a hpaModel.DeploymentHpaIntent, p string, ca string, v string, di string, exists bool) (hpaModel.DeploymentHpaIntent, error)
	GetIntent(i string, p string, ca string, v string, di string) (hpaModel.DeploymentHpaIntent, bool, error)
	GetAllIntents(p, ca, v, di string) ([]hpaModel.DeploymentHpaIntent, error)
	GetAllIntentsByApp(app, p, ca, v, di string) ([]hpaModel.DeploymentHpaIntent, error)
	GetIntentByName(i, p, ca, v, di string) (hpaModel.DeploymentHpaIntent, error)
	DeleteIntent(i string, p string, ca string, v string, di string) error

	// consumers
	AddConsumer(a hpaModel.HpaResourceConsumer, p string, ca string, v string, di string, i string, exists bool) (hpaModel.HpaResourceConsumer, error)
	GetConsumer(cn string, p string, ca string, v string, di string, i string) (hpaModel.HpaResourceConsumer, bool, error)
	GetAllConsumers(p, ca, v, di, i string) ([]hpaModel.HpaResourceConsumer, error)
	GetConsumerByName(cn, p, ca, v, di, i string) (hpaModel.HpaResourceConsumer, error)
	DeleteConsumer(cn, p string, ca string, v string, di string, i string) error

	// resources
	AddResource(a hpaModel.HpaResourceRequirement, p string, ca string, v string, di string, i string, cn string, exists bool) (hpaModel.HpaResourceRequirement, error)
	GetResource(rn string, p string, ca string, v string, di string, i string, cn string) (hpaModel.HpaResourceRequirement, bool, error)
	GetAllResources(p, ca, v, di, i, cn string) ([]hpaModel.HpaResourceRequirement, error)
	GetResourceByName(rn, p, ca, v, di, i, cn string) (hpaModel.HpaResourceRequirement, error)
	DeleteResource(rn string, p string, ca string, v string, di string, i string, cn string) error
}

// HpaPlacementClient implements the HpaPlacementManager interface
type HpaPlacementClient struct {
	db hpaModel.ClientDBInfo
}

// NewHpaPlacementClient returns an instance of the HpaPlacementClient
func NewHpaPlacementClient() *HpaPlacementClient {
	return &HpaPlacementClient{
		db: hpaModel.ClientDBInfo{
			StoreName:   "HpaPlacementController",
			TagMetaData: "HpaPlacementControllerMetadata",
			TagContent:  "HpaPlacementControllerContent",
			TagState:    "HpaPlacementControllerStateInfo",
		},
	}
}

// HpaIntentKey ... consists of intent name, Project name, CompositeApp name,
// CompositeApp version, deployment intent group
type HpaIntentKey struct {
	IntentName            string `json:"intentname"`
	Project               string `json:"project"`
	CompositeApp          string `json:"compositeapp"`
	Version               string `json:"compositeappversion"`
	DeploymentIntentGroup string `json:"deploymentintentgroup"`
}

// We will use json marshalling to convert to string to
// preserve the underlying structure.
func (ik HpaIntentKey) String() string {
	out, err := json.Marshal(ik)
	if err != nil {
		return ""
	}

	return string(out)
}

// HpaConsumerKey ... consists of Name if the Consumer name, Project name, CompositeApp name,
// CompositeApp version, Deployment intent group, Intent name
type HpaConsumerKey struct {
	ConsumerName          string `json:"consumername"`
	IntentName            string `json:"intentname"`
	Project               string `json:"project"`
	CompositeApp          string `json:"compositeapp"`
	Version               string `json:"compositeappversion"`
	DeploymentIntentGroup string `json:"deploymentintentgroup"`
}

// We will use json marshalling to convert to string to
// preserve the underlying structure.
func (ck HpaConsumerKey) String() string {
	out, err := json.Marshal(ck)
	if err != nil {
		return ""
	}

	return string(out)
}

// HpaResourceKey ... consists of Name of the Resource name, Project name, CompositeApp name,
// CompositeApp version, Deployment intent group, Intent name, Consumer name
type HpaResourceKey struct {
	ResourceName          string `json:"resourcename"`
	ConsumerName          string `json:"consumername"`
	IntentName            string `json:"intentname"`
	Project               string `json:"project"`
	CompositeApp          string `json:"compositeapp"`
	Version               string `json:"compositeappversion"`
	DeploymentIntentGroup string `json:"deploymentintentgroup"`
}

// We will use json marshalling to convert to string to
// preserve the underlying structure.
func (rk HpaResourceKey) String() string {
	out, err := json.Marshal(rk)
	if err != nil {
		return ""
	}

	return string(out)
}
