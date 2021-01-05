package action

// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2020 Intel Corporation

import (
	"encoding/base64"
	"encoding/json"
	"strings"

	"github.com/open-ness/EMCO/src/clm/pkg/cluster"
	jh "github.com/open-ness/EMCO/src/genericactioncontroller/pkg/jsonapihelper"
	"github.com/open-ness/EMCO/src/genericactioncontroller/pkg/module"
	"github.com/open-ness/EMCO/src/orchestrator/pkg/appcontext"
	log "github.com/open-ness/EMCO/src/orchestrator/pkg/infra/logutils"
	pkgerrors "github.com/pkg/errors"
)

// isValidClusterToApplyByLabel checks if the cluster being authenticated for(acName) falls under the given label(cLabel) and provider(cProvider)
func isValidClusterToApplyByLabel(cProvider, acName, cLabel, cMode string) (bool, error) {

	clusterNamesList, err := cluster.NewClusterClient().GetClustersWithLabel(cProvider, cLabel)
	if err != nil {
		return false, err
	}
	acName = strings.Split(acName, SEPARATOR)[1]
	for _, cn := range clusterNamesList {
		
		if cn == acName && cMode == "allow" {
			return true, nil
		}
	}
	return false, nil
}

// isValidClusterToApplyByName checks if a given cluster(gcName) under a provider(cProvider) matches with the cluster which is authenticated for(acName).
func isValidClusterToApplyByName(cProvider, acName, gcName, cMode string) (bool, error) {

	clusterNamesList, err := cluster.NewClusterClient().GetClusters(cProvider)
	if err != nil {
		return false, err
	}
	acName = strings.Split(acName, SEPARATOR)[1]
	for _, cn := range clusterNamesList {
		if cn.Metadata.Name == acName && cMode == "allow" {
			return true, nil
		}
	}
	return false, nil
}

// UpdateContextDirectly is invoked in cases when the resource is neither configMap nor secret. For eg, network policy. It takes in :
/*
p - ProjectName
ca - CompAppName
cv - CompAppVersion
dig - DepIntentGrpName
GenK8sIntName - GenK8sIntentName
ac - AppContext
cz - Customization
*/
func UpdateContextDirectly(p, ca, cv, dig, GenK8sIntName string, rs module.Resource, ac appcontext.AppContext, cz module.Customization) error {

	cSpecific := strings.ToLower(cz.Spec.ClusterSpecific)
	cScope := strings.ToLower(cz.Spec.ClusterInfo.Scope)
	cProvider := cz.Spec.ClusterInfo.ClusterProvider
	cName := cz.Spec.ClusterInfo.ClusterName
	cLabel := cz.Spec.ClusterInfo.ClusterLabel
	cMode := strings.ToLower(cz.Spec.ClusterInfo.Mode)

	appName := rs.Spec.AppName
	clusters, err := ac.GetClusterNames(appName)
	if err != nil {
		log.Error("Error GetClusterNames", log.Fields{"appName": appName, "project": p, "CompApp": ca, "CompVer": cv, "DepIntentGrp": dig, "intentName": GenK8sIntName, "resourceName": rs.Metadata.Name})
		return pkgerrors.Errorf("Internal error")
	}

	rsFilecontent, err := module.NewResourceClient().GetResourceContent(rs.Metadata.Name, p, ca, cv, dig, GenK8sIntName)
	if err != nil {
		log.Error("Error during UpdateContextDirectly while getting ResourceContent", log.Fields{
			"Error":        err.Error(),
			"ResourceName": rs.Metadata.Name,
			"Project":      p,
			"CompositeApp": ca,
			"Version":      cv,
			"DepIntentGrp": dig,
			"GenK8sIntent": GenK8sIntName,
		})
		return err
	}
	dataBytes, err := base64.StdEncoding.DecodeString(rsFilecontent.FileContent)
	if err != nil {
		log.Error("Error DecodeString", log.Fields{"ResourceName": rs.Metadata.Name})
		return err
	}

	// BEGIN : applying to clusters

	for _, c := range clusters {

		if cSpecific == "true" && cScope == "label" {
			allow, err := isValidClusterToApplyByLabel(cProvider, c, cLabel, cMode)
			if err != nil {
				log.Error("Error ApplyToClusterByLabel", log.Fields{"Provider": cProvider, "ClusterName": cName, "ClusterLabel": cLabel, "Mode": cMode})
				return pkgerrors.Errorf("Internal error")
			}
			if !allow {
				continue
			}
		}
		if cSpecific == "true" && cScope == "name" {
			allow, err := isValidClusterToApplyByName(cProvider, c, cName, cMode)
			if err != nil {
				log.Error("Error ApplyClusterByName", log.Fields{"Provider": cProvider, "GivenClusterName": cName, "AutheticatingForCluste": c, "Mode": cMode})
				return pkgerrors.Errorf("Internal error")
			}
			if !allow {
				continue
			}
		}

		ch, err := ac.GetClusterHandle(appName, c)
		if err != nil {
			log.Error("Error GetClusterHandle", log.Fields{"appName": appName, "project": p, "CompApp": ca, "CompVer": cv, "DepIntentGrp": dig, "intentName": GenK8sIntName, "resourceName": rs.Metadata.Name})
			return pkgerrors.Errorf("Internal error")
		}

		resName := rs.Spec.ResourceGVK.Name + SEPARATOR + rs.Spec.ResourceGVK.Kind
		// Add the new resource
		_, err = ac.AddResource(ch, resName, string(dataBytes))
		if err != nil {
			log.Error("Error AddResource", log.Fields{"appName": appName, "project": p, "CompApp": ca, "CompVer": cv, "DepIntentGrp": dig, "intentName": GenK8sIntName, "resourceName": rs.Metadata.Name})
			return pkgerrors.Errorf("Internal error")
		}
		// update the resource order
		resorder, err := ac.GetResourceInstruction(appName, c, "order")
		if err != nil {
			log.Error("Error getting Resource order", log.Fields{
				"error":        err,
				"app name":     appName,
				"cluster name": c,
			})
			return pkgerrors.Errorf("Internal error")
		}
		aov := make(map[string][]string)
		json.Unmarshal([]byte(resorder.(string)), &aov)
		aov["resorder"] = append(aov["resorder"], resName)
		jresord, _ := json.Marshal(aov)

		_, err = ac.AddInstruction(ch, "resource", "order", string(jresord))
		if err != nil {
			log.Error("Error updating Resource order", log.Fields{
				"error":        err,
				"app name":     appName,
				"cluster name": c,
			})
			return pkgerrors.Errorf("Internal error")
		}
	}
	return nil
}

// UpdateContextExistingResourceUsingPatchArray updates the context for an existing resource
func UpdateContextExistingResourceUsingPatchArray(p, ca, cv, dig, GenK8sIntName string, rs module.Resource, ac appcontext.AppContext, cz module.Customization) error {
	//resName := rs.Spec.ResourceGVK.Name + SEPARATOR + rs.Spec.ResourceGVK.Kind
	appName := rs.Spec.AppName

	byteValuePatchJSON, err := jh.GetPatchFromPatchJSON(cz.Spec.PatchJSON)
	if err != nil {
		log.Error(":: Error GetPatchFromPatchJSON ::", log.Fields{"Error": err})
		return pkgerrors.Errorf("Internal error")
	}
	log.Error("byteValuePatchJSON", log.Fields{"byteValuePatchJSON": string(byteValuePatchJSON)})
	//dataBytes, err := getModifiedYaml(rs)

	clusters, err := ac.GetClusterNames(appName)
	if err != nil {
		log.Error("Error GetClusterNames", log.Fields{"appName": appName, "project": p, "CompApp": ca, "CompVer": cv, "DepIntentGrp": dig, "intentName": GenK8sIntName, "resourceName": rs.Metadata.Name})
		return pkgerrors.Errorf("Internal error")
	}

	cSpecific := strings.ToLower(cz.Spec.ClusterSpecific)
	cScope := strings.ToLower(cz.Spec.ClusterInfo.Scope)
	cProvider := cz.Spec.ClusterInfo.ClusterProvider
	cName := cz.Spec.ClusterInfo.ClusterName
	cLabel := cz.Spec.ClusterInfo.ClusterLabel
	cMode := strings.ToLower(cz.Spec.ClusterInfo.Mode)

	for _, c := range clusters {

		if cSpecific == "true" && cScope == "label" {
			allow, err := isValidClusterToApplyByLabel(cProvider, c, cLabel, cMode)
			if err != nil {
				log.Error("Error ApplyToClusterByLabel", log.Fields{"Provider": cProvider, "ClusterName": cName, "ClusterLabel": cLabel, "Mode": cMode})
				return pkgerrors.Errorf("Internal error")
			}
			if !allow {
				continue
			}
		}
		if cSpecific == "true" && cScope == "name" {
			allow, err := isValidClusterToApplyByName(cProvider, c, cName, cMode)
			if err != nil {
				log.Error("Error ApplyClusterByName", log.Fields{"Provider": cProvider, "GivenClusterName": cName, "AutheticatingForCluste": c, "Mode": cMode})
				return pkgerrors.Errorf("Internal error")
			}
			if !allow {
				continue
			}
		}

		rh, err := ac.GetResourceHandle(rs.Spec.AppName, c,
			strings.Join([]string{rs.Spec.ResourceGVK.Name, rs.Spec.ResourceGVK.Kind}, SEPARATOR))
		if err != nil {
			log.Error("App Context resource handle not found", log.Fields{
				"project":                 p,
				"composite app":           ca,
				"composite app version":   cv,
				"deployment intent group": dig,
				"GenK8sIntent":            GenK8sIntName,
				"AppName":                 rs.Spec.AppName,
				"Resource name":           rs.Metadata.Name,
				"resource kind":           rs.Spec.ResourceGVK.Kind,
			})
			continue
		}

		r, err := ac.GetValue(rh)
		if err != nil {
			log.Error("Error retrieving resource from App Context", log.Fields{
				"error":           err,
				"resource handle": rh,
			})
			continue
		}
		log.Info("manifest file for the resource", log.Fields{"ResName": rs.Spec.ResourceGVK.Name, "Manifest-File": r.(string)})

		dataBytes, err := jh.GenerateModifiedYamlFileForExistingResources(byteValuePatchJSON, []byte(r.(string)), rs.Metadata.Name)
		if err != nil {
			log.Error("Error retrieving dataBytes", log.Fields{
				"error": err,
			})
			continue
		}
		log.Info("Data bytes", log.Fields{"dataBytes": string(dataBytes)})

		err = ac.UpdateResourceValue(rh, string(dataBytes))
		if err != nil {
			log.Error("Error UpdateResourceValue", log.Fields{"appName": appName, "project": p, "CompApp": ca, "CompVer": cv, "DepIntentGrp": dig, "intentName": GenK8sIntName, "resourceName": rs.Metadata.Name})
			return pkgerrors.Errorf("Internal error")
		}
		log.Info("Resource updated in AppContext", log.Fields{"appName": appName, "project": p, "CompApp": ca, "CompVer": cv, "DepIntentGrp": dig, "intentName": GenK8sIntName, "resourceName": rs.Metadata.Name})
	}
	return nil

}
