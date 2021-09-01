// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2020 Intel Corporation

package resources

// ClusterResourceObj ... Struct to container cluster resource objects
type ClusterResourceObj struct {
	ClusterName      string
	AllocatableRs    *GenericResource
	NonAllocatableRs *NFDResource
}

// ClusterResourceObjMap ...
type ClusterResourceObjMap map[string]ClusterResourceObj

// ClusterResourceInfo ...
type ClusterResourceInfo struct {
	// Cluster name
	ClusterName string `json:"cluster-name"`
	// Cluster Available Resource Count
	ClusterAvailResCount int64 `json:"cluster-avail-res-count"`
	// Cluster Available CPU Count
	ClusterAvailResCountOrig int64 `json:"cluster-avail-res-count-orig"`
	// Cluster Node Max Available CPU Count
	NodeMaxAvailResCount int64 `json:"node-max-avail-res-count"`
	// Cluster Node Max Available CPU Count
	NodeMaxAvailResCountOrig int64 `json:"node-max-avail-res-count-orig"`
	// Qualified nodes satisfying resource rules
	QualifiedNodes []string `json:"qualified-nodes"`
}

// ClusterResourceInfoMap ...
type ClusterResourceInfoMap map[string]ClusterResourceInfo
