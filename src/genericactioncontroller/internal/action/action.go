package action

// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2020 Intel Corporation

import (
	"encoding/base64"
	"encoding/json"
	"strings"

	jh "github.com/open-ness/EMCO/src/genericactioncontroller/pkg/jsonapihelper"
	"github.com/open-ness/EMCO/src/genericactioncontroller/pkg/module"
	"github.com/open-ness/EMCO/src/orchestrator/pkg/appcontext"
	log "github.com/open-ness/EMCO/src/orchestrator/pkg/infra/logutils"
	pkgerrors "github.com/pkg/errors"
)

// SEPARATOR used while creating resourceNames to store in etcd
const SEPARATOR = "+"

// UpdateAppContext is the method which calls the backend logic of this controller.
func UpdateAppContext(intentName, appContextID string) error {
	log.Info("Begin updating app context ", log.Fields{"intent-name": intentName, "appcontext": appContextID})

	var ac appcontext.AppContext
	_, err := ac.LoadAppContext(appContextID)
	if err != nil {
		log.Error("Loading AppContext failed ", log.Fields{"intent-name": intentName, "appcontext": appContextID, "Error": err.Error()})
		return pkgerrors.Errorf("Internal error")
	}

	caMeta, err := ac.GetCompositeAppMeta()
	if err != nil {
		log.Error("Error getting metadata for AppContext ", log.Fields{"intent-name": intentName, "appcontext": appContextID, "Error": err.Error()})
		return pkgerrors.Errorf("Internal error")
	}

	p := caMeta.Project
	ca := caMeta.CompositeApp
	cv := caMeta.Version
	dig := caMeta.DeploymentIntentGroup

	// get all the resources under this intent
	resources, err := module.NewResourceClient().GetAllResources(p, ca, cv, dig, intentName)
	if err != nil {
		log.Error("Error GetAllResources", log.Fields{"project": p, "CompApp": ca, "CompVer": cv, "DepIntentGrp": dig, "IntentName": intentName})
		return pkgerrors.Errorf("Internal error")
	}

	for _, rs := range resources {

		// if resource is either configMap or secret, it must have customization files.
		// go through each customization, find customization content, get the patchfiles,
		// generate the modified contentfiles and update context for each of the valid cluster
		czs, err := module.NewCustomizationClient().GetAllCustomization(p, ca, cv, dig, intentName, rs.Metadata.Name)
		if err != nil {
			log.Error("Error GetAllCustomization", log.Fields{"project": p, "CompApp": ca, "CompVer": cv, "DepIntentGrp": dig, "intentName": intentName, "resourceName": rs.Metadata.Name})
			return pkgerrors.Errorf("Internal error")
		}

		var cSpecific, cScope, cProvider, cName, cLabel, cMode string

		for _, cz := range czs {

			// check if clusterSpecific is true, and then get everything from the clusterInfo
			if strings.ToLower(cz.Spec.ClusterSpecific) == "true" && (module.ClusterInfo{}) != cz.Spec.ClusterInfo {
				cSpecific = strings.ToLower(cz.Spec.ClusterSpecific)
				cScope = strings.ToLower(cz.Spec.ClusterInfo.Scope)
				cProvider = cz.Spec.ClusterInfo.ClusterProvider
				cName = cz.Spec.ClusterInfo.ClusterName
				cLabel = cz.Spec.ClusterInfo.ClusterLabel
				cMode = strings.ToLower(cz.Spec.ClusterInfo.Mode)
			}

			// if resource kind is neither configMap nor secret, directly update the context
			if strings.ToLower(rs.Spec.ResourceGVK.Kind) != "configmap" && strings.ToLower(rs.Spec.ResourceGVK.Kind) != "secret" && strings.ToLower(rs.Spec.NewObject) == "true" {
				return UpdateContextDirectly(p, ca, cv, dig, intentName, rs, ac, cz)
			}

			if strings.ToLower(cz.Spec.PatchType) == "json" && strings.ToLower(rs.Spec.NewObject) == "false" {
				return UpdateContextExistingResourceUsingPatchArray(p, ca, cv, dig, intentName, rs, ac, cz)
			}

			dataArr, err := module.NewCustomizationClient().GetCustomizationContent(cz.Metadata.Name, p, ca, cv, dig, intentName, rs.Metadata.Name)
			if err != nil {
				log.Error("Error GetCustomizationContent", log.Fields{
					"CustomizationName": cz.Metadata.Name,
					"project":           p, "CompApp": ca,
					"CompVer":      cv,
					"DepIntentGrp": dig,
					"intentName":   intentName,
					"resourceName": rs.Metadata.Name})
				return pkgerrors.Errorf("Internal error")
			}
			var dataBytes []byte
			var dataFiles [][]byte

			for _, content := range dataArr.FileContents {
				
				if strings.ToLower(rs.Spec.ResourceGVK.Kind) == "secret" {
					dataBytes = []byte(content)
				} else {
					dataBytes, err = base64.StdEncoding.DecodeString(content)
				}
				
				if err != nil {
					log.Error(":: Base64 encoding error ::", log.Fields{"Error": err})
					return pkgerrors.Errorf("Internal error")
				}
				dataFiles = append(dataFiles, dataBytes)
			}

			byteValuePatchJSON, err := jh.GetPatchFromFile(dataArr.FileNames)
			if err != nil {
				log.Error(":: Error GetPatchFromFile ::", log.Fields{"Error": err})
				return pkgerrors.Errorf("Internal error")
			}
			modifiedConfigMap, err := jh.GenerateModifiedConfigFile(dataFiles, byteValuePatchJSON, dataArr.FileNames, rs.Spec.ResourceGVK.Name, rs.Spec.ResourceGVK.Kind)
			if err != nil {
				log.Error(":: Error GenerateModifiedConfigFile ::", log.Fields{"Error": err})
				return pkgerrors.Errorf("Internal error")
			}

			appName := rs.Spec.AppName
			clusters, err := ac.GetClusterNames(appName)
			if err != nil {
				log.Error("Error GetClusterNames", log.Fields{"appName": appName, "project": p, "CompApp": ca, "CompVer": cv, "DepIntentGrp": dig, "intentName": intentName, "resourceName": rs.Metadata.Name})
				return pkgerrors.Errorf("Internal error")
			}

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
					log.Error("Error GetClusterHandle", log.Fields{"appName": appName, "project": p, "CompApp": ca, "CompVer": cv, "DepIntentGrp": dig, "intentName": intentName, "resourceName": rs.Metadata.Name})
					return pkgerrors.Errorf("Internal error")
				}

				resName := rs.Spec.ResourceGVK.Name + SEPARATOR + rs.Spec.ResourceGVK.Kind
				// Add the new resource
				_, err = ac.AddResource(ch, resName, string(modifiedConfigMap))
				if err != nil {
					log.Error("Error AddResource", log.Fields{"appName": appName, "project": p, "CompApp": ca, "CompVer": cv, "DepIntentGrp": dig, "intentName": intentName, "resourceName": rs.Metadata.Name})
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
		}
	}
	return nil
}
