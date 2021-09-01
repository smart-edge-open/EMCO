// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2020 Intel Corporation

package resources

import (
	"context"

	hpaModel "github.com/open-ness/EMCO/src/hpa-plc/pkg/model"
)

// Resources interface of a intent
type Resources interface {
	// Initialize the object
	Initialize()
	// Populate Resource Info
	PopulateResourceInfo(ctx context.Context, clusterName string, hpaResource hpaModel.HpaResourceRequirement) (int64, map[string]int64, error)
	// Qualified checks whether required resources are met
	Qualified(ctx context.Context, clusterName string, hpaResource hpaModel.HpaResourceRequirement) bool
	// Get cluster resource count
	GetClusterResourceCount(res string) int64
	// Set cluster resource count
	SetClusterResourceCount(res string, val int64)
	// Get max resource available count of the cluster node
	// example: if two nodes of a cluster as available cpu count as 2 & 5 then this function will return 5
	GetNodeResourceAvailMaxCount(res string) int64
	// Get cluster resource count
	GetClusterResourceCountOrig(res string) int64
	// Get max resource available count of the cluster node
	// example: if two nodes of a cluster as available cpu count as 2 & 5 then this function will return 5
	GetNodeResourceAvailMaxCountOrig(res string) int64
	// Get qualified nodes list
	GetQualifiedNodes(res string) []string
	// Get node resource map
	GetNodeResMap(res string) map[string]int64
	// Set node resource map
	SetNodeResMap(res string, newMap map[string]int64)
	// Get original node resource map
	GetNodeResMapOrig(res string) map[string]int64
	// Update max resource available count of the cluster node
	UpdateNodeResourceAvailMaxCount(res string, nodeMap map[string]int64)
	// Set max resource available count of the cluster node
	SetNodeResourceAvailMaxCount(res string, val int64)
	// Update resource stats
	UpdateNodeResourceCounts(nodeName string, hpaResource hpaModel.HpaResourceRequirement)
	// Rollback accounting
	RollbackAccounting(res string) error
	// Is Resource already Populated
	IsResourceAlreadyPopulated(res string) bool
}
