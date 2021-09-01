// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2020 Intel Corporation

package resources

import (
	"context"
	"sort"

	hpaModel "github.com/open-ness/EMCO/src/hpa-plc/pkg/model"
	hpaUtils "github.com/open-ness/EMCO/src/hpa-plc/pkg/utils"
	log "github.com/open-ness/EMCO/src/orchestrator/pkg/infra/logutils"
	connector "github.com/open-ness/EMCO/src/rsync/pkg/connector"
	corev1 "k8s.io/api/core/v1"
)

// GenericResource ... holds the definition of a Generic resource
type GenericResource struct {
	clusterAvailCount     map[string]int64
	nodeMaxAvailCount     map[string]int64
	clusterAvailCountOrig map[string]int64
	nodeMaxAvailCountOrig map[string]int64
	nodeResMap            map[string](map[string]int64)
	nodeResMapOrig        map[string](map[string]int64)
	qualifiedNodes        map[string]([]string)
	clusterName           map[string]string
	ConfigLoaded          map[string]bool
	isInitialized         bool
}

// Initialize ..
func (p *GenericResource) Initialize() {
	p.clusterAvailCount = make(map[string]int64)
	p.nodeMaxAvailCount = make(map[string]int64)
	p.clusterAvailCountOrig = make(map[string]int64)
	p.nodeMaxAvailCountOrig = make(map[string]int64)
	p.nodeResMap = make(map[string](map[string]int64))
	p.nodeResMapOrig = make(map[string](map[string]int64))
	p.qualifiedNodes = make(map[string]([]string))
	p.clusterName = make(map[string]string)
	p.ConfigLoaded = make(map[string]bool)

	p.isInitialized = true
}

// Qualified checks whether required Generic resources are met
func (p *GenericResource) Qualified(ctx context.Context, clusterName string, hpaResource hpaModel.HpaResourceRequirement) bool {
	// populate resource info if it's not already loaded
	if !(p.ConfigLoaded[hpaResource.Spec.Resource.Name]) {
		_, _, err := p.PopulateResourceInfo(ctx, clusterName, hpaResource)
		if err != nil {
			log.Error("Qualified PopulateResourceInfo Generic failed", log.Fields{"err": err})
			return false
		}
	}

	// qualified nodes are for each qualify check, clear previous qualified nodes
	p.qualifiedNodes = nil

	log.Info("Qualified Generic .. start => ", log.Fields{
		"resource-name":               hpaResource.Spec.Resource.Name,
		"hpa-resource":                hpaResource,
		"GenericClusterAvailable":     p.clusterAvailCount,
		"GenericNodeMaxAvailable":     p.nodeMaxAvailCount,
		"GenericClusterAvailableOrig": p.clusterAvailCountOrig,
		"GenericNodeMaxAvailableOrig": p.nodeMaxAvailCountOrig,
		"nodeGenericMap":              p.nodeResMap,
		"qualifiedNodes":              p.qualifiedNodes,
		"clusterName":                 p.clusterName})

	// if limits are mentioned or zero then assign requests to limits
	if hpaResource.Spec.Resource.Limits == 0 {
		log.Info("Qualified Generic .. limits are zero => ", log.Fields{"hpa-resource-spec": hpaResource.Spec})
		hpaResource.Spec.Resource.Limits = hpaResource.Spec.Resource.Requests
	}

	// convert memory units from MB to Bytes
	if hpaResource.Spec.Resource.Name == string(corev1.ResourceMemory) {
		// convert Mega Bytes to Bytes
		hpaResource.Spec.Resource.Requests *= 1000 * 1000
		hpaResource.Spec.Resource.Limits *= 1000 * 1000
	}

	matched := false
	p.qualifiedNodes = make(map[string][]string)
	p.qualifiedNodes[hpaResource.Spec.Resource.Name] = make([]string, 0)
	if (p.clusterAvailCount[hpaResource.Spec.Resource.Name] >= hpaResource.Spec.Resource.Requests) && (p.nodeMaxAvailCount[hpaResource.Spec.Resource.Name] >= hpaResource.Spec.Resource.Requests) {
		for resource, nodeMap := range p.nodeResMap {
			if resource == hpaResource.Spec.Resource.Name {
				for key, value := range nodeMap {
					if hpaResource.Spec.Resource.Requests <= value {
						// add to the qualified nodes list
						if !hpaUtils.IsInSlice(key, p.qualifiedNodes[hpaResource.Spec.Resource.Name]) {
							p.qualifiedNodes[hpaResource.Spec.Resource.Name] = append(p.qualifiedNodes[hpaResource.Spec.Resource.Name], key)
						}

						matched = true
						log.Info("Qualified cluster node matched required res .. Generic => ", log.Fields{
							"resource_name":           hpaResource.Spec.Resource.Name,
							"request":                 hpaResource.Spec.Resource.Requests,
							"limit":                   hpaResource.Spec.Resource.Limits,
							"node_name":               key,
							"GenericNodeavailable":    p.nodeResMap[hpaResource.Spec.Resource.Name],
							"GenericNodeMaxAvailable": p.nodeMaxAvailCount,
							"GenericClusterAvailable": p.clusterAvailCount})
					}
				} // for nodes
			}
		}
	} // if ((p.clusterAvailCount >= requestTotalRes)

	// if no qualified clusters found, remove the existing Qualified nodes from list & rollback the nodeMap
	if !matched {
		p.qualifiedNodes = nil
	}

	log.Info("Qualified Generic .. end => ", log.Fields{
		"resource-name":               hpaResource.Spec.Resource.Name,
		"matched":                     (len(p.qualifiedNodes) > 0),
		"hpa-resource":                hpaResource,
		"GenericClusterAvailable":     p.clusterAvailCount,
		"GenericNodeMaxAvailable":     p.nodeMaxAvailCount,
		"GenericClusterAvailableOrig": p.clusterAvailCountOrig,
		"GenericNodeMaxAvailableOrig": p.nodeMaxAvailCountOrig,
		"nodeGenericMap":              p.nodeResMap,
		"qualifiedNodes":              p.qualifiedNodes,
		"clusterName":                 p.clusterName})

	return (len(p.qualifiedNodes) > 0)
}

// PopulateResourceInfo .. Fetch Generic resource info based on KubeConfig
func (p *GenericResource) PopulateResourceInfo(ctx context.Context, clusterName string, hpaResource hpaModel.HpaResourceRequirement) (int64, map[string]int64, error) {
	log.Info("PopulateResourceInfo Generic .. start => ", log.Fields{"clusterName": clusterName, "hpaResource": hpaResource})

	// Initialize models if not not initialized already
	if !p.isInitialized {
		p.Initialize()
	}

	p.clusterAvailCount[hpaResource.Spec.Resource.Name] = 0
	p.ConfigLoaded[hpaResource.Spec.Resource.Name] = false
	p.clusterName[hpaResource.Spec.Resource.Name] = clusterName
	var err error

	// Connect to Cluster
	con := connector.Connection{}
	con.Init("cluster")
	//con := connector.Init("cluster")
	//Cleanup
	defer con.RemoveClient()

	// Get Kube Client handle
	c, err := con.GetClient(clusterName, "0", "default")
	if err != nil {
		log.Error("Error in creating kubeconfig client", log.Fields{
			"error":        err,
			"cluster-name": clusterName,
		})
		return 0, nil, err
	}

	if clusterGenericCount, _, nodeGenericMap, err := c.GetAvailableNodeResources(context.TODO(), hpaResource.Spec.Resource.Name); err == nil {
		p.clusterAvailCount[hpaResource.Spec.Resource.Name] = clusterGenericCount
		p.nodeResMap[hpaResource.Spec.Resource.Name] = make(map[string]int64)
		p.nodeResMap[hpaResource.Spec.Resource.Name] = nodeGenericMap
		// Keep the original map as this is needed to rollback accounting if a different cluster is chosen for deployment
		for k, v := range nodeGenericMap {
			(p.nodeResMapOrig[hpaResource.Spec.Resource.Name]) = make(map[string]int64)
			(p.nodeResMapOrig[hpaResource.Spec.Resource.Name])[k] = v
		}
		if clusterGenericCount > 0 {
			p.UpdateNodeResourceAvailMaxCount(hpaResource.Spec.Resource.Name, p.nodeResMap[hpaResource.Spec.Resource.Name])
			p.clusterAvailCountOrig = p.clusterAvailCount
			p.nodeMaxAvailCountOrig = p.nodeMaxAvailCount
			log.Info("PopulateResourceInfo Generic nodeMaxAvailCount => ", log.Fields{"nodeMaxAvailCount": p.nodeMaxAvailCount})
		}
	}

	p.ConfigLoaded[hpaResource.Spec.Resource.Name] = true
	log.Info("PopulateResourceInfo Generic .. end => ", log.Fields{
		"hpaResource":              hpaResource,
		"clusterAvailGenericCount": p.clusterAvailCount,
		"nodeMaxAvailGenericCount": p.nodeMaxAvailCount,
		"nodeResMap":               p.nodeResMap,
		"clusterName":              clusterName})
	return p.clusterAvailCount[hpaResource.Spec.Resource.Name], p.nodeResMap[hpaResource.Spec.Resource.Name], err
}

// GetClusterResourceCount .. Get cluster resource count
func (p *GenericResource) GetClusterResourceCount(res string) int64 {
	return p.clusterAvailCount[res]
}

// SetClusterResourceCount .. Set cluster resource count
func (p *GenericResource) SetClusterResourceCount(res string, val int64) {
	p.clusterAvailCount[res] = val
}

// GetNodeResourceAvailMaxCount .. Get max resource available count of the cluster node
// example: if two nodes of a cluster as available Generic count as 2 & 5 then this function will return 5
func (p *GenericResource) GetNodeResourceAvailMaxCount(res string) int64 {
	return p.nodeMaxAvailCount[res]
}

// GetClusterResourceCountOrig .. Get cluster resource count
func (p *GenericResource) GetClusterResourceCountOrig(res string) int64 {
	return p.clusterAvailCountOrig[res]
}

// GetNodeResourceAvailMaxCountOrig .. Get max resource available count of the cluster node
// example: if two nodes of a cluster as available Generic count as 2 & 5 then this function will return 5
func (p *GenericResource) GetNodeResourceAvailMaxCountOrig(res string) int64 {
	return p.nodeMaxAvailCountOrig[res]
}

// GetQualifiedNodes .. Get qualified nodes list
func (p *GenericResource) GetQualifiedNodes(res string) []string {
	return p.qualifiedNodes[res]
}

// UpdateNodeResourceAvailMaxCount .. Updat max resource available count of the cluster node
func (p *GenericResource) UpdateNodeResourceAvailMaxCount(res string, nodeMap map[string]int64) {
	// sort the node available resource count
	if len(nodeMap) > 0 {
		values := make([]int64, 0, len(nodeMap))
		for _, v := range nodeMap {
			values = append(values, v)
		}
		if len(values) > 0 {
			sort.Slice(values, func(i, j int) bool { return values[i] > values[j] })
			p.nodeMaxAvailCount[res] = values[0]
		}
	}
}

// SetNodeResourceAvailMaxCount ... Set max resource available count of the cluster node
func (p *GenericResource) SetNodeResourceAvailMaxCount(res string, val int64) {
	p.nodeMaxAvailCount[res] = val
}

// UpdateNodeResourceCounts ... Update resource stats
func (p *GenericResource) UpdateNodeResourceCounts(nodeName string, hpaResource hpaModel.HpaResourceRequirement) {
	log.Info("UpdateNodeResourceCounts .. start", log.Fields{"hpa-resource": hpaResource, "node-name": nodeName, "nodeMap": p.nodeResMap})
	// update the available counts
	p.clusterAvailCount[hpaResource.Spec.Resource.Name] -= hpaResource.Spec.Resource.Requests
	(p.nodeResMap[hpaResource.Spec.Resource.Name])[nodeName] -= hpaResource.Spec.Resource.Requests
	p.UpdateNodeResourceAvailMaxCount(hpaResource.Spec.Resource.Name, p.nodeResMap[hpaResource.Spec.Resource.Name])
}

// GetNodeResMap ... Get node resource map
func (p *GenericResource) GetNodeResMap(res string) map[string]int64 {
	return p.nodeResMap[res]
}

// GetNodeResMapOrig ... Get original node resource map
func (p *GenericResource) GetNodeResMapOrig(res string) map[string]int64 {
	return p.nodeResMapOrig[res]
}

// SetNodeResMap ... Set node resource map
func (p *GenericResource) SetNodeResMap(res string, newMap map[string]int64) {
	// Update nodemap
	for k, v := range newMap {
		(p.nodeResMap[res])[k] = v
	}
}

// RollbackAccounting ... Rollback resource accounting
func (p *GenericResource) RollbackAccounting(res string) error {
	log.Info("RollbackAccounting .. start", log.Fields{"resource": res, "nodeMap": p.nodeResMap, "nodeMapOrig": p.nodeResMapOrig})
	p.SetNodeResMap(res, p.GetNodeResMapOrig(res))
	p.SetClusterResourceCount(res, p.GetClusterResourceCountOrig(res))
	p.SetNodeResourceAvailMaxCount(res, p.GetNodeResourceAvailMaxCountOrig(res))
	return nil
}

// IsResourceAlreadyPopulated .. Check if resource already populated
func (p *GenericResource) IsResourceAlreadyPopulated(res string) bool {
	return p.ConfigLoaded[res]
}
