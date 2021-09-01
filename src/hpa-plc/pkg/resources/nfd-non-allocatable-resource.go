// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2020 Intel Corporation

package resources

import (
	"context"

	hpaModel "github.com/open-ness/EMCO/src/hpa-plc/pkg/model"
	hpaUtils "github.com/open-ness/EMCO/src/hpa-plc/pkg/utils"
	log "github.com/open-ness/EMCO/src/orchestrator/pkg/infra/logutils"
	connector "github.com/open-ness/EMCO/src/rsync/pkg/connector"
)

// NFDResource holds the definition of a NFD resource
type NFDResource struct {
	clusterName    string
	configLoaded   bool
	qualifiedNodes []string
	nodeLabels     map[string](map[string]string)
}

// Qualified checks whether required NFD resources are met
func (p *NFDResource) Qualified(ctx context.Context, clusterName string, hpaResource hpaModel.HpaResourceRequirement) bool {
	// populate resource info if it's not already loaded
	if !p.configLoaded {
		err := p.PopulateResourceInfo(ctx, clusterName)
		if err != nil {
			log.Error("Qualified PopulateResourceInfo NFD failed", log.Fields{"err": err})
			return false
		}
	}

	log.Info("Qualified NFD .. start => ", log.Fields{
		"resource_name":  hpaResource.Spec.Resource.Name,
		"label-key":      hpaResource.Spec.Resource.Key,
		"label_value":    hpaResource.Spec.Resource.Value,
		"nodeLables":     p.nodeLabels,
		"qualifiedNodes": p.qualifiedNodes,
		"clusterName":    p.clusterName})

	matched := false

	for key, value := range p.nodeLabels {
		// Atleast one of the nodes should have the label
		if val, ok := value[hpaResource.Spec.Resource.Key]; ok {
			if val == hpaResource.Spec.Resource.Value {
				matched = true
				// add to the qualified nodes list
				if !hpaUtils.IsInSlice(key, p.qualifiedNodes) {
					p.qualifiedNodes = append(p.qualifiedNodes, key)
				}
			}
		}
	} // for key, value := range p.nodeResMap {

	// if no qualified clusters found, remove the existing Qualified nodes from list
	if !matched {
		p.qualifiedNodes = nil
	}

	log.Info("Qualified NFD .. end => ", log.Fields{
		"matched":        matched,
		"resource_name":  hpaResource.Spec.Resource.Name,
		"label-key":      hpaResource.Spec.Resource.Key,
		"label_value":    hpaResource.Spec.Resource.Value,
		"nodeLables":     p.nodeLabels,
		"qualifiedNodes": p.qualifiedNodes,
		"clusterName":    p.clusterName})

	return matched
}

// PopulateResourceInfo .. Fetch NFD resource info based on KubeConfig
func (p *NFDResource) PopulateResourceInfo(ctx context.Context, clusterName string) error {
	log.Info("PopulateResourceInfo NFD .. start => ", log.Fields{"clusterName": clusterName})

	p.configLoaded = false
	p.clusterName = clusterName
	p.qualifiedNodes = make([]string, 0)
	p.nodeLabels = make(map[string](map[string]string))

	// Connect to Cluster
	con := connector.Connection{}
	con.Init("clsuter")
	//Cleanup
	defer con.RemoveClient()

	// Get Kube Client handle
	c, err := con.GetClient(clusterName, "0", "default")
	if err != nil {
		log.Error("Error in creating kubeconfig client", log.Fields{
			"error":        err,
			"cluster-name": clusterName,
		})
		return err
	}

	if c != nil {
		if nodeLabels, err := c.GetNodeLabels(context.TODO()); err == nil {
			p.nodeLabels = nodeLabels
		}
		p.configLoaded = true
	} else {
		log.Error("Error in getting kubeconfig client .. null pointer", log.Fields{
			"connection":   c,
			"error":        err,
			"cluster-name": clusterName,
		})
	}

	log.Info("PopulateResourceInfo NFD .. end => ", log.Fields{
		"nodeLabels":  p.nodeLabels,
		"clusterName": clusterName})
	return err
}

// SetResourceInfo .. Set NFD resource
func (p *NFDResource) SetResourceInfo(ctx context.Context, clusterName string, nodeLabels map[string](map[string]string)) error {
	log.Info("SetResourceInfo NFD .. start => ", log.Fields{"clusterName": clusterName})

	p.configLoaded = true
	p.clusterName = clusterName
	p.qualifiedNodes = make([]string, 0)
	p.nodeLabels = make(map[string](map[string]string))
	p.nodeLabels = nodeLabels

	log.Info("SetResourceInfo NFD .. end => ", log.Fields{
		"nodeLabels":  p.nodeLabels,
		"clusterName": clusterName})
	return nil
}

// GetQualifiedNodes .. Get qualified nodes list
func (p *NFDResource) GetQualifiedNodes() []string {
	return p.qualifiedNodes
}
