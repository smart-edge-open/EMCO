// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2020 Intel Corporation

package resources

import (
	"context"

	hpaModel "github.com/open-ness/EMCO/src/hpa-plc/pkg/model"
)

// NonAllocatableResources interface of a intent
type NonAllocatableResources interface {
	// Populate NonAllocatableResources Info
	PopulateResourceInfo(ctx context.Context, clusterName string) error
	// Set NonAllocatableResources Info
	SetResourceInfo(ctx context.Context, clusterName string, nodeLabels map[string](map[string]string)) error
	// Qualified checks whether required NonAllocatableResources are met
	Qualified(ctx context.Context, clusterName string, hpaResource hpaModel.HpaResourceRequirement) bool
	// Get qualified nodes list
	GetQualifiedNodes() []string
}
