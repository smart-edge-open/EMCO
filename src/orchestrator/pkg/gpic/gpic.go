// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2020 Intel Corporation

package gpic

/*
 gpic stands for GenericPlacementIntent Controller.
 This file pertains to the implementation and handling of generic placement intents
*/

import (
	"github.com/open-ness/EMCO/src/clm/pkg/cluster"
	pkgerrors "github.com/pkg/errors"
	"log"
	"strconv"
)

// ClusterList consists of mandatoryClusters and clusterGroups
type ClusterList struct {
	MandatoryClusters []ClusterWithName
	ClusterGroups     []ClusterGroup
}

//ClusterGroup consists of a list of optionalClusters and a groupNumber. All the clusters under the optional clusters belong to same groupNumber
type ClusterGroup struct {
	OptionalClusters []ClusterWithName
	GroupNumber      string
}

// ClusterWithName has two fields - ProviderName and ClusterName
type ClusterWithName struct {
	ProviderName string
	ClusterName  string
}

// ClusterWithLabel has two fields - ProviderName and ClusterLabel
type ClusterWithLabel struct {
	ProviderName string
	ClusterLabel string
}

// IntentStruc consists of AllOfArray and AnyOfArray
type IntentStruc struct {
	AllOfArray []AllOf `json:"allOf,omitempty"`
	AnyOfArray []AnyOf `json:"anyOf,omitempty"`
}

// AllOf consists if ProviderName, ClusterName, ClusterLabelName and AnyOfArray. Any of them can be empty
type AllOf struct {
	ProviderName     string  `json:"provider-name,omitempty"`
	ClusterName      string  `json:"cluster-name,omitempty"`
	ClusterLabelName string  `json:"cluster-label-name,omitempty"`
	AnyOfArray       []AnyOf `json:"anyOf,omitempty"`
}

// AnyOf consists of Array of ProviderName & ClusterLabelNames
type AnyOf struct {
	ProviderName     string `json:"provider-name,omitempty"`
	ClusterName      string `json:"cluster-name,omitempty"`
	ClusterLabelName string `json:"cluster-label-name,omitempty"`
}

// intentResolverHelper helps to populate the cluster lists
func intentResolverHelper(pn, cn, cln string, clustersWithName []ClusterWithName) ([]ClusterWithName, error) {
	if cln == "" && cn != "" {
		eachClusterWithName := ClusterWithName{pn, cn}
		clustersWithName = append(clustersWithName, eachClusterWithName)
		log.Printf("Added Cluster: %s ", cn)
	}
	if cn == "" && cln != "" {
		//Finding cluster names for the clusterlabel
		clusterNamesList, err := cluster.NewClusterClient().GetClustersWithLabel(pn, cln)
		if err != nil {
			return []ClusterWithName{}, pkgerrors.Wrap(err, "Error getting clusterLabels")
		}
		// Populate the clustersWithName array with the clusternames found above
		for _, eachClusterName := range clusterNamesList {
			eachClusterWithPN := ClusterWithName{pn, eachClusterName}
			clustersWithName = append(clustersWithName, eachClusterWithPN)
			log.Printf("Added Cluster :: %s through its label: %s ", eachClusterName, cln)
		}
	}
	return clustersWithName, nil
}

// IntentResolver shall help to resolve the given intent into 2 lists of clusters where the app need to be deployed.
func IntentResolver(intent IntentStruc) (ClusterList, error) {
	var mc []ClusterWithName
	var err error
	var cg []ClusterGroup
	index := 0
	for _, eachAllOf := range intent.AllOfArray {
		mc, err = intentResolverHelper(eachAllOf.ProviderName, eachAllOf.ClusterName, eachAllOf.ClusterLabelName, mc)
		if err != nil {
			return ClusterList{}, pkgerrors.Wrap(err, "intentResolverHelper error")
		}
		if len(eachAllOf.AnyOfArray) > 0 {
			for _, eachAnyOf := range eachAllOf.AnyOfArray {
				var opc []ClusterWithName
				opc, err = intentResolverHelper(eachAnyOf.ProviderName, eachAnyOf.ClusterName, eachAnyOf.ClusterLabelName, opc)
				index++
				if err != nil {
					return ClusterList{}, pkgerrors.Wrap(err, "intentResolverHelper error")
				}
				eachClustergroup := ClusterGroup{OptionalClusters: opc, GroupNumber: strconv.Itoa(index)}
				cg = append(cg, eachClustergroup)
			}
		}
	}
	if len(intent.AnyOfArray) > 0 {
		var opc []ClusterWithName
		for _, eachAnyOf := range intent.AnyOfArray {
			opc, err = intentResolverHelper(eachAnyOf.ProviderName, eachAnyOf.ClusterName, eachAnyOf.ClusterLabelName, opc)
			index++
			if err != nil {
				return ClusterList{}, pkgerrors.Wrap(err, "intentResolverHelper error")
			}
			eachClustergroup := ClusterGroup{OptionalClusters: opc, GroupNumber: strconv.Itoa(index)}
			cg = append(cg, eachClustergroup)
		}
	}
	clusterList := ClusterList{MandatoryClusters: mc, ClusterGroups: cg}
	return clusterList, nil
}
