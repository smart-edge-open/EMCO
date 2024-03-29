// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2020 Intel Corporation

package module

import (
	"encoding/json"
	"fmt"

	nettypes "github.com/k8snetworkplumbingwg/network-attachment-definition-client/pkg/apis/k8s.cni.cncf.io/v1"
	netutils "github.com/k8snetworkplumbingwg/network-attachment-definition-client/pkg/utils"
	log "github.com/open-ness/EMCO/src/orchestrator/pkg/infra/logutils"
	v1 "k8s.io/api/apps/v1"
	batch "k8s.io/api/batch/v1"
	batchv1beta1 "k8s.io/api/batch/v1beta1"
	v1core "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/client-go/kubernetes/scheme"
	_ "k8s.io/kubernetes/pkg/apis/apps/install"
	_ "k8s.io/kubernetes/pkg/apis/batch/install"
	_ "k8s.io/kubernetes/pkg/apis/core/install"
	_ "k8s.io/kubernetes/pkg/apis/extensions/install"

	pkgerrors "github.com/pkg/errors"
)

type NfnAnnotation struct {
	CniType   string                 `json:"type"`
	Interface []WorkloadIfIntentSpec `json:"interface"`
}

const NfnAnnotationKey = "k8s.plugin.opnfv.org/nfn-network"

// ParsePodTemplateNetworkAnnotation parses Pod annotation in PodTemplate
func ParsePodTemplateNetworkAnnotation(pt *v1core.PodTemplateSpec) ([]*nettypes.NetworkSelectionElement, error) {
	netAnnot := pt.Annotations[nettypes.NetworkAttachmentAnnot]
	defaultNamespace := pt.Namespace

	if len(netAnnot) == 0 {
		return nil, pkgerrors.Errorf("No kubernetes network annotation found")
	}

	networks, err := netutils.ParseNetworkAnnotation(netAnnot, defaultNamespace)
	if err != nil {
		return nil, err
	}
	return networks, nil
}

// GetPodTemplateNetworkAnnotation gets Pod Nfn annotation in PodTemplate
func GetPodTemplateNfnAnnotation(pt *v1core.PodTemplateSpec) NfnAnnotation {
	var nfn NfnAnnotation

	nfnAnnot := pt.Annotations[NfnAnnotationKey]
	if len(nfnAnnot) == 0 {
		return nfn
	}

	err := json.Unmarshal([]byte(nfnAnnot), &nfn)
	if err != nil {
		log.Warn("Error unmarshalling nfn annotation", log.Fields{
			"annotation": nfnAnnot,
		})
	}
	return nfn
}

// GetPodNetworkAnnotation gets Pod Nfn annotation in PodTemplate
func GetPodNfnAnnotation(p *v1core.Pod) NfnAnnotation {
	var nfn NfnAnnotation

	nfnAnnot := p.Annotations[NfnAnnotationKey]
	if len(nfnAnnot) == 0 {
		return nfn
	}

	err := json.Unmarshal([]byte(nfnAnnot), &nfn)
	if err != nil {
		log.Warn("Error unmarshalling nfn annotation", log.Fields{
			"annotation": nfnAnnot,
		})
	}
	return nfn
}

func addNetworkAnnotation(a nettypes.NetworkSelectionElement, as []*nettypes.NetworkSelectionElement) []*nettypes.NetworkSelectionElement {
	var netElements []*nettypes.NetworkSelectionElement

	found := false
	for _, e := range as {
		if e.Name == a.Name {
			found = true
		}
		netElements = append(netElements, e)
	}
	if !found {
		netElements = append(netElements, &a)
	}

	return netElements
}

// Add the interfaces in the 'new' parameter to the nfn annotation
func newNfnIfs(nfn NfnAnnotation, new []WorkloadIfIntentSpec) NfnAnnotation {
	// Prepare a new interface list - combining the original plus new ones
	var newNfn NfnAnnotation

	if nfn.CniType != CNI_TYPE_OVN4NFV {
		if len(nfn.CniType) > 0 {
			log.Warn("Error existing nfn cnitype is invalid", log.Fields{
				"existing cnitype": nfn.CniType,
				"using cnitype":    CNI_TYPE_OVN4NFV,
			})
		}
	}
	newNfn.CniType = CNI_TYPE_OVN4NFV

	// update any interfaces already in the list with the updated interface
	for _, i := range nfn.Interface {
		for _, j := range new {
			if i.NetworkName == j.NetworkName && i.IfName == j.IfName {
				i.DefaultGateway = j.DefaultGateway
				i.IpAddr = j.IpAddr
				i.MacAddr = j.MacAddr
				break
			}
		}
		newNfn.Interface = append(newNfn.Interface, i)
	}

	// add new interfaces not present in original list
	for _, j := range new {
		found := false
		for _, i := range nfn.Interface {
			if i.NetworkName == j.NetworkName && i.IfName == j.IfName {
				found = true
				break
			}
		}
		if !found {
			newNfn.Interface = append(newNfn.Interface, j)
		}
	}
	return newNfn
}

func updatePodTemplateNetworkAnnotation(pt *v1core.PodTemplateSpec, a nettypes.NetworkSelectionElement) {
	netAnnotation, _ := ParsePodTemplateNetworkAnnotation(pt)
	elements := addNetworkAnnotation(a, netAnnotation)
	j, err := json.Marshal(elements)
	if err != nil {
		log.Error("Existing network annotation has invalid format", log.Fields{
			"error": err,
		})
		return
	}
	if pt.Annotations == nil {
		pt.Annotations = make(map[string]string)
	}
	pt.Annotations[nettypes.NetworkAttachmentAnnot] = string(j)
}

// Add a network annotation to the resource
func AddNetworkAnnotation(r interface{}, a nettypes.NetworkSelectionElement) {

	switch o := r.(type) {
	case *batch.Job:
		updatePodTemplateNetworkAnnotation(&o.Spec.Template, a)
	case *batchv1beta1.CronJob:
		updatePodTemplateNetworkAnnotation(&o.Spec.JobTemplate.Spec.Template, a)
	case *v1.DaemonSet:
		updatePodTemplateNetworkAnnotation(&o.Spec.Template, a)
	case *v1.Deployment:
		updatePodTemplateNetworkAnnotation(&o.Spec.Template, a)
	case *v1.ReplicaSet:
		updatePodTemplateNetworkAnnotation(&o.Spec.Template, a)
	case *v1.StatefulSet:
		updatePodTemplateNetworkAnnotation(&o.Spec.Template, a)
	case *v1core.Pod:
		netAnnotation, _ := netutils.ParsePodNetworkAnnotation(o)
		elements := addNetworkAnnotation(a, netAnnotation)
		j, err := json.Marshal(elements)
		if err != nil {
			log.Error("Existing network annotation has invalid format", log.Fields{
				"error": err,
			})
			break
		}
		if o.Annotations == nil {
			o.Annotations = make(map[string]string)
		}
		o.Annotations[nettypes.NetworkAttachmentAnnot] = string(j)
		return
	case *v1core.ReplicationController:
		updatePodTemplateNetworkAnnotation(o.Spec.Template, a)
	default:
		typeStr := fmt.Sprintf("%T", o)
		log.Warn("Network annotations not supported for resource type", log.Fields{
			"resource type": typeStr,
		})
	}
}

func updatePodTemplateNfnAnnotation(pt *v1core.PodTemplateSpec, new []WorkloadIfIntentSpec) {
	nfnAnnotation := GetPodTemplateNfnAnnotation(pt)
	newNfnAnnotation := newNfnIfs(nfnAnnotation, new)
	j, err := json.Marshal(newNfnAnnotation)
	if err != nil {
		log.Error("Network nfn annotation has invalid format", log.Fields{
			"error": err,
		})
		return
	}
	if pt.Annotations == nil {
		pt.Annotations = make(map[string]string)
	}
	pt.Annotations[NfnAnnotationKey] = string(j)
}

// Add an nfn annotation to the resource
func AddNfnAnnotation(r interface{}, new []WorkloadIfIntentSpec) {

	switch o := r.(type) {
	case *batch.Job:
		updatePodTemplateNfnAnnotation(&o.Spec.Template, new)
	case *batchv1beta1.CronJob:
		updatePodTemplateNfnAnnotation(&o.Spec.JobTemplate.Spec.Template, new)
	case *v1.DaemonSet:
		updatePodTemplateNfnAnnotation(&o.Spec.Template, new)
		return
	case *v1.Deployment:
		updatePodTemplateNfnAnnotation(&o.Spec.Template, new)
		return
	case *v1.ReplicaSet:
		updatePodTemplateNfnAnnotation(&o.Spec.Template, new)
	case *v1.StatefulSet:
		updatePodTemplateNfnAnnotation(&o.Spec.Template, new)
	case *v1core.Pod:
		nfnAnnotation := GetPodNfnAnnotation(o)
		newNfnAnnotation := newNfnIfs(nfnAnnotation, new)
		j, err := json.Marshal(newNfnAnnotation)
		if err != nil {
			log.Error("Network nfn annotation has invalid format", log.Fields{
				"error": err,
			})
			break
		}
		if o.Annotations == nil {
			o.Annotations = make(map[string]string)
		}
		o.Annotations[NfnAnnotationKey] = string(j)
		return
	case *v1core.ReplicationController:
		updatePodTemplateNfnAnnotation(o.Spec.Template, new)
		return
	default:
		typeStr := fmt.Sprintf("%T", o)
		log.Warn("Network nfn annotations not supported for resource type", log.Fields{
			"resource type": typeStr,
		})
	}
}
// Add Annotations for non standard K8s resources with template section
func AddTemplateAnnotation(r interface{}, a nettypes.NetworkSelectionElement, new []WorkloadIfIntentSpec) ([]byte, error) {

	//Decode the yaml to create a runtime.Object
	unstruct := &unstructured.Unstructured{}
	decode := scheme.Codecs.UniversalDeserializer().Decode
	_, _, err := decode([]byte(r.(string)), nil, unstruct)
	if err != nil {
		return nil, pkgerrors.Wrap(err, "Deserialize YAML error")
	}
	// Check if template section is available
	_, found, err := unstructured.NestedMap(unstruct.Object, "spec", "template")
	if err != nil  || !found {
		log.Error("Error NestedMap no template section", log.Fields{"unstruct": unstruct})
		return nil, err
	}
	err = updateTemplateNfnAnnotation(unstruct, new)
	if err != nil {
		return nil, pkgerrors.Wrap(err, "updateTemplateNfnAnnotation error")
	}
	err = updateTemplateNetworkAnnotation(unstruct, a)
	if err != nil {
		return nil, pkgerrors.Wrap(err, "updateTemplateNetworkAnnotation error")
	}
	// Convert to Json
	b, err := unstruct.MarshalJSON()
	if err != nil {
		return nil, err
	}
	return b, nil
}
// Add NFN annotation to resources with template section
func updateTemplateNfnAnnotation(unstruct *unstructured.Unstructured, new []WorkloadIfIntentSpec) error {
	var nfnAnnotation NfnAnnotation
	annotations, found, err := unstructured.NestedMap(unstruct.Object, "spec", "template", "metadata", "annotations")
	if err != nil   {
		log.Error("Error NestedMap for template metadata annotations", log.Fields{"unstruct": unstruct})
		return err
	}
	if !found || annotations == nil {
		annotations = make(map[string]interface{})
	}

	nfnAnnotIntf := annotations[NfnAnnotationKey]
	if nfnAnnotIntf != nil {
		nfnAnnot := fmt.Sprintf("%v", nfnAnnotIntf)
		if len(nfnAnnot) != 0 {
			err := json.Unmarshal([]byte(nfnAnnot), &nfnAnnotation)
			if err != nil {
				log.Warn("Error unmarshalling nfn annotation", log.Fields{
					"annotation": nfnAnnot,
				})
				return err
			}
		}
	}
	newNfnAnnotation := newNfnIfs(nfnAnnotation, new)
	j, err := json.Marshal(newNfnAnnotation)
	if err != nil {
		log.Error("Network nfn annotation has invalid format", log.Fields{
			"error": err,
		})
		return err
	}
	annotations[NfnAnnotationKey] = string(j)
	if err := unstructured.SetNestedMap(unstruct.Object, annotations, "spec", "template", "metadata", "annotations"); err != nil {
		log.Error("Error adding template with nfn annotations", log.Fields{"err": err})
	}
	return err
}
// Add network annotation to resources with template section
func updateTemplateNetworkAnnotation(unstruct *unstructured.Unstructured, a nettypes.NetworkSelectionElement) error {
	//netAnnotation, _ := ParsePodTemplateNetworkAnnotation(pt)
	annotations, found, err := unstructured.NestedMap(unstruct.Object, "spec", "template", "metadata", "annotations")
	if err != nil   {
		log.Error("Error NestedMap for template metadata annotations", log.Fields{"unstruct": unstruct})
		return err
	}
	if !found || annotations == nil {
		annotations = make(map[string]interface{})
	}
	defaultNamespace := unstruct.GetNamespace()
	netAnnotIntf := annotations[nettypes.NetworkAttachmentAnnot]
	var netAnnotation []*nettypes.NetworkSelectionElement
	if netAnnotIntf != nil {
		netAnnot := fmt.Sprintf("%v", netAnnotIntf)
		netAnnotation, err = netutils.ParseNetworkAnnotation(netAnnot, defaultNamespace)
		if err != nil {
			return err
		}
	} else {
		netAnnotation = nil
	}
	elements := addNetworkAnnotation(a, netAnnotation)
	j, err := json.Marshal(elements)
	if err != nil {
		log.Error("Existing network annotation has invalid format", log.Fields{
			"error": err,
		})
		return err
	}
	annotations[nettypes.NetworkAttachmentAnnot] = string(j)
	if err := unstructured.SetNestedMap(unstruct.Object, annotations, "spec", "template", "metadata", "annotations"); err != nil {
		log.Error("Error adding template with nfn annotations", log.Fields{"err": err})
	}
	return err
}
