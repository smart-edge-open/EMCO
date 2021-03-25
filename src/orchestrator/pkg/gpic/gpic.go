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

// ClusterList consists of mandatoryClusters and OptionalClusters
type ClusterList struct {
	MandatoryClusters []ClusterGroup
	OptionalClusters  []ClusterGroup
}

//ClusterGroup consists of Clusters and GroupNumber. All the clusters under the same clusterGroup belong to the same groupNumber
type ClusterGroup struct {
	Clusters    []ClusterWithName
	GroupNumber string
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
var intentResolverHelper = func(pn, cn, cln string, clusters []ClusterWithName) ([]ClusterWithName, error) {
	if cln == "" && cn != "" {
		eachClusterWithName := ClusterWithName{pn, cn}
		clusters = append(clusters, eachClusterWithName)
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
			clusters = append(clusters, eachClusterWithPN)
			log.Printf("Added Cluster :: %s through its label: %s ", eachClusterName, cln)
		}
	}
	return clusters, nil
}

// IntentResolver shall help to resolve the given intent into 2 lists of clusters where the app need to be deployed.
func IntentResolver(intent IntentStruc) (ClusterList, error) {
	var mc []ClusterWithName
	var mClusters []ClusterGroup
	var err error
	var oClusters []ClusterGroup
	index := 0
	for _, eachAllOf := range intent.AllOfArray {
		mc, err := intentResolverHelper(eachAllOf.ProviderName, eachAllOf.ClusterName, eachAllOf.ClusterLabelName, mc)
		if err != nil {
			return ClusterList{}, pkgerrors.Wrap(err, "intentResolverHelper error")
		}
		for _, eachMC := range mc {
			index++
			var arrCname []ClusterWithName
			arrCname = append(arrCname, eachMC)
			eachMandatoryCluster := ClusterGroup{Clusters: arrCname, GroupNumber: strconv.Itoa(index)}
			mClusters = append(mClusters, eachMandatoryCluster)
		}

		if len(eachAllOf.AnyOfArray) > 0 {
			index++
			for _, eachAnyOf := range eachAllOf.AnyOfArray {
				var opc []ClusterWithName
				opc, err = intentResolverHelper(eachAnyOf.ProviderName, eachAnyOf.ClusterName, eachAnyOf.ClusterLabelName, opc)
				if err != nil {
					return ClusterList{}, pkgerrors.Wrap(err, "intentResolverHelper error")
				}
				eachOptionalCluster := ClusterGroup{Clusters: opc, GroupNumber: strconv.Itoa(index)}
				oClusters = append(oClusters, eachOptionalCluster)
			}
		}
	}
	if len(intent.AnyOfArray) > 0 {
		index++
		for _, eachAnyOf := range intent.AnyOfArray {
			var opc []ClusterWithName
			opc, err = intentResolverHelper(eachAnyOf.ProviderName, eachAnyOf.ClusterName, eachAnyOf.ClusterLabelName, opc)
			if err != nil {
				return ClusterList{}, pkgerrors.Wrap(err, "intentResolverHelper error")
			}
			eachOptionalCluster := ClusterGroup{Clusters: opc, GroupNumber: strconv.Itoa(index)}
			oClusters = append(oClusters, eachOptionalCluster)
		}
	}
	clusterList := ClusterList{MandatoryClusters: mClusters, OptionalClusters: oClusters}
	return clusterList, nil
}
