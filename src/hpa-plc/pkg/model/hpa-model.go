// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2020 Intel Corporation

package model

import (
	mtypes "github.com/open-ness/EMCO/src/orchestrator/pkg/module/types"
)

// ClientDBInfo ... to save hpa data to db
type ClientDBInfo struct {
	StoreName   string // name of the mongodb collection to use for client documents
	TagMetaData string // attribute key name for the json data of a client document
	TagContent  string // attribute key name for the file data of a client document
	TagState    string // attribute key name for StateInfo object in the cluster
}

// DeploymentHpaIntentSpec .. DeploymentHpaIntent spec
type DeploymentHpaIntentSpec struct {
	// Target Application Name of the Hpa intent (e.g. Prometheus)
	AppName string `json:"app-name,omitempty"`
}

// DeploymentHpaIntent ..
type DeploymentHpaIntent struct {
	// Intent Metadata
	MetaData mtypes.Metadata `json:"metadata,omitempty"`

	// Intent Spec
	Spec DeploymentHpaIntentSpec `json:"spec,omitempty"`
}

// HpaResourceConsumerSpec .. HpaIntent ResourceConsumer spec
type HpaResourceConsumerSpec struct {
	// K8s version (e.g apps/v1)
	APIVersion string `json:"api-version,omitempty"`
	// Type of object (e.g. Deployment)
	Kind string `json:"kind,omitempty"`
	// Replicas of the consumer object(e.g deployment)
	Replicas int64 `json:"replicas,omitempty"`
	// From metadata/name field of the consumer object(e.g deployment)
	Name string `json:"name,omitempty"`
	// Container name of the consumer object(e.g. deployment). This field is required for allocatable resources only
	ContainerName string `json:"container-name,omitempty"`
}

// HpaResourceConsumer .. Intent mapping to K8s
type HpaResourceConsumer struct {
	// Intent Metadata
	MetaData mtypes.Metadata `json:"metadata,omitempty"`

	// Intent Spec
	Spec HpaResourceConsumerSpec `json:"spec,omitempty"`
}

// NonAllocatableResources ..
type NonAllocatableResources struct {
	// kubernetes label key
	Key string `json:"key,omitempty"`
	// kubernetes label value
	Value string `json:"value,omitempty"`
}

// AllocatableResources ..
type AllocatableResources struct {
	// The requested resource type  (e.g. nvidia.com/gpu)
	Name string `json:"name,omitempty"`
	// The requested number of resource instances. (e.g memory is expressed as bytes by default)
	Requests int64 `json:"requests,omitempty"`
	// The limit of resource instances. (e.g memory is expressed as bytes by default)
	Limits int64 `json:"limits,omitempty"`
	// resource units.(e.g MB for memory resource represents Mega Bytes)
	Units string `json:"units,omitempty"`
}

// HpaResourceRequirementDetails .. One of Allocatable/NonAllocatable
type HpaResourceRequirementDetails struct {
	AllocatableResources
	NonAllocatableResources
}

// HpaResourceRequirementSpec .. Hpa resource spec
type HpaResourceRequirementSpec struct {
	// Whether resource is allocatble
	Allocatable *bool `json:"allocatable,omitempty"`
	// Whether requested resource type is mandatory or optional
	Mandatory bool `json:"mandatory,omitempty"`
	// Whether requested resource type is mandatory or optional
	Weight int32 `json:"weight,omitempty"`
	// Resource spec
	Resource HpaResourceRequirementDetails `json:"resource,omitempty"`
}

// HpaResourceRequirement .. HpaIntent Resource Requirements
type HpaResourceRequirement struct {
	// Intent Metadata
	MetaData mtypes.Metadata `json:"metadata,omitempty"`

	// Intent Spec
	Spec HpaResourceRequirementSpec `json:"spec,omitempty"`
}
