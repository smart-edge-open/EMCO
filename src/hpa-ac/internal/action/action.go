// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2020 Intel Corporation

package action

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/open-ness/EMCO/src/orchestrator/pkg/appcontext"
	log "github.com/open-ness/EMCO/src/orchestrator/pkg/infra/logutils"

	hpaActionUtils "github.com/open-ness/EMCO/src/hpa-ac/pkg/utils"
	hpaModel "github.com/open-ness/EMCO/src/hpa-plc/pkg/model"
	hpaModuleLib "github.com/open-ness/EMCO/src/hpa-plc/pkg/module"
	orchModuleLib "github.com/open-ness/EMCO/src/orchestrator/pkg/module"

	jyaml "github.com/ghodss/yaml"
	pkgerrors "github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
)

const SEPARATOR = "+"

// UpdateAppContext breaks down the spec from hpa placement controller and updates appcontext for rsync
func UpdateAppContext(intentName, appContextID string) error {
	log.Info("UpdateAppContext HPA .. start", log.Fields{"intent-name": intentName, "appcontext": appContextID})

	var ac appcontext.AppContext
	_, err := ac.LoadAppContext(appContextID)
	if err != nil {
		log.Error("UpdateAppContext HPA ..Loading AppContext failed.", log.Fields{"intent-name": intentName, "appcontext": appContextID, "Error": err})
		return pkgerrors.Errorf("UpdateAppContext HPA .. Error in loading AppContext failed. Internal error")
	}

	caMeta, err := ac.GetCompositeAppMeta()
	if err != nil {
		log.Error("UpdateAppContext HPA .. Error in getting App metadata.", log.Fields{"intent-name": intentName, "appcontext": appContextID, "Error": err})
		return pkgerrors.Errorf("UpdateAppContext HPA .. Error in getting App metadata. Internal error")
	}

	project := caMeta.Project
	compositeApp := caMeta.CompositeApp
	compositeAppVersion := caMeta.Version
	deploymentIntentGroup := caMeta.DeploymentIntentGroup

	// Get all apps in this composite app
	apps, err := orchModuleLib.NewAppClient().GetApps(project, compositeApp, compositeAppVersion)
	if err != nil {
		log.Error("UpdateAppContext HPA .. Not finding the compositeApp attached apps", log.Fields{"appContextID": appContextID, "compositeApp": compositeApp, "caMeta": caMeta, "err": err})
		return nil
	}
	allAppNames := make([]string, 0)
	for _, a := range apps {
		allAppNames = append(allAppNames, a.Metadata.Name)
	}
	log.Info("UpdateAppContext HPA .. Applications attached to compositeApp", log.Fields{"compositeApp": compositeApp, "app-names": allAppNames})

	// Iterate through all apps of the Composite App
	for appIndex, eachApp := range allAppNames {
		// Handle all hpa Intents of the app
		hpaIntents, err := hpaModuleLib.NewHpaPlacementClient().GetAllIntentsByApp(eachApp, project, compositeApp, compositeAppVersion, deploymentIntentGroup)
		if err != nil {
			log.Error("FilterClusters .. Error getting hpa Intents", log.Fields{"project": project, "compositeApp": compositeApp, "deploymentGroup": deploymentIntentGroup})
			//return pkgerrors.Wrapf(err, "FilterClusters .. Error getting hpa Intents for project[%v] compositeApp[%v] compositeVersion[%v] deploymentGroup[%v] not found", project, compositeApp, compositeAppVersion, deploymentIntentGroup)
			return nil
		}

		log.Info("UpdateAppContext HPA .. Intents attached to app", log.Fields{"app-index": appIndex,
			"app-name":        eachApp,
			"len-hpa-intents": len(hpaIntents), "hpa-intents": hpaIntents})

		// loop through all hpa Intents
		for index, hpaIntent := range hpaIntents {
			log.Info("UpdateAppContext HPA .. start.", log.Fields{
				"app-index":               appIndex,
				"app-name":                eachApp,
				"intent-index":            index,
				"project":                 project,
				"composite-app":           compositeApp,
				"composite-app-version":   compositeAppVersion,
				"deployment-intent-group": deploymentIntentGroup,
				"hpa-intent-name":         hpaIntent.MetaData.Name,
				"index":                   index,
				"hpa-intent":              hpaIntent,
			})

			// Handle all hpa Consumers
			hpaConsumers, err := hpaModuleLib.NewHpaPlacementClient().GetAllConsumers(project, compositeApp, compositeAppVersion, deploymentIntentGroup, hpaIntent.MetaData.Name)
			if err != nil {
				log.Error("UpdateAppContext HPA .. Error in GetAllConsumers.", log.Fields{
					"hpa-intent-name": hpaIntent.MetaData.Name,
					"app-name":        hpaIntent.Spec.AppName,
					"err":             err})

				//return pkgerrors.Wrapf(err, "UpdateAppContext HPA .. Error in GetAllConsumers. Intent %v for %v/%v%v not found", hpaIntent.MetaData.Name, project, compositeApp, compositeAppVersion)
				return nil
			}

			// Handle hpa consumer
			for index, hpaConsumer := range hpaConsumers {
				log.Info("UpdateAppContext HPA .. Fetching HpaConsumers .. start", log.Fields{"index": index, "app": hpaIntent.Spec.AppName, "hpa-intent": hpaIntent,
					"hpa-consumer": hpaConsumer})

				// Handle all hpa Resources
				hpaResources, err := hpaModuleLib.NewHpaPlacementClient().GetAllResources(project, compositeApp, compositeAppVersion, deploymentIntentGroup, hpaIntent.MetaData.Name, hpaConsumer.MetaData.Name)
				if err != nil {
					log.Error("UpdateAppContext HPA .. Error in GetAllResources.", log.Fields{
						"hpa-intent-name": hpaIntent.MetaData.Name,
						"app-name":        hpaIntent.Spec.AppName,
						"err":             err})

					//return pkgerrors.Wrapf(err, "UpdateAppContext HPA .. Error in GetAllResources. Intent %v for %v/%v%v not found", hpaIntent.MetaData.Name, project, compositeApp, compositeAppVersion)
					return nil
				}
				// Add resources to consumer deployment spec
				for index, hpaResource := range hpaResources {
					log.Info("UpdateAppContext HPA .. adding resource to consumer spec", log.Fields{"index": index, "app-name": hpaIntent.Spec.AppName, "hpa-intent": hpaIntent,
						"hpa-consumer": hpaConsumer, "hpa-resource": hpaResource})

					// If consumer spec name in resource key matches this consumer spec name, we can add the resource to this consumer spec
					// Assuming all consumer is deployment here. Can be of other types
					hpaclusters, err := ac.GetClusterNames(hpaIntent.Spec.AppName)
					if err != nil {
						log.Error("UpdateAppContext HPA .. Error in GetClusterNames.", log.Fields{
							"hpa-intent-name": hpaIntent.MetaData.Name,
							"app-name":        hpaIntent.Spec.AppName,
							"err":             err})

						return pkgerrors.Wrapf(err, "UpdateAppContext HPA .. Error in GetClusterNames. intent-name[%v] app-name[%v].",
							hpaIntent.MetaData.Name, hpaIntent.Spec.AppName)
					}
					for clusterindex, cluster := range hpaclusters {
						var deployRes []string
						deployRes = make([]string, 0)
						if len(hpaConsumer.Spec.Name) == 0 {
							deployResDB, err := ac.GetResourceNames(hpaIntent.Spec.AppName, cluster)
							if err != nil {
								log.Error("UpdateAppContext HPA .. Error in GetResourceNames.", log.Fields{
									"hpa-intent-name": hpaIntent.MetaData.Name,
									"app-name":        hpaIntent.Spec.AppName,
									"err":             err})
								return pkgerrors.Wrapf(err, "UpdateAppContext HPA .. Error in GetResourceNames. hpa-intent-name[%v] app-name[%v] err[%v]",
									hpaIntent.MetaData.Name, hpaIntent.Spec.AppName, err)
							}

							// Filter Deployment resources
							for _, k := range deployResDB {
								resName, kind := strings.ToUpper(k), strings.ToUpper("Deployment")
								if strings.Contains(resName, kind) {
									deployRes = append(deployRes, k)
								}
							}
							log.Info("UpdateAppContext HPA .. consumer kind name not specified", log.Fields{
								"hpa-intent-name": hpaIntent.MetaData.Name,
								"app-name":        hpaIntent.Spec.AppName,
								"deployRes":       deployRes})
						} else {
							deployRes = append(deployRes, hpaConsumer.Spec.Name+SEPARATOR+"Deployment")
						}

						log.Info("UpdateAppContext HPA ..  Fetched resources from app context.", log.Fields{
							"cluster-index":        clusterindex,
							"cluster":              cluster,
							"hpa-intent-name":      hpaIntent.MetaData.Name,
							"app-name":             hpaIntent.Spec.AppName,
							"deployment-resources": deployRes})

						for _, resName := range deployRes {
							res, err := getResource(ac, resName, cluster, hpaIntent.Spec.AppName)
							if err != nil {
								log.Error("UpdateAppContext HPA .. Error in fetching resource handle from app context.", log.Fields{
									"hpa-intent-name": hpaIntent.MetaData.Name,
									"cluster-index":   clusterindex,
									"app-name":        hpaIntent.Spec.AppName,
									"res-name":        resName,
									"cluster":         cluster,
									"err":             err})
								return pkgerrors.Wrapf(err, "UpdateAppContext HPA .. Error in fetching resource[%v] handle of cluster[%v] from app hpa-intent-name[%v] app-name[%v]",
									resName, cluster, hpaIntent.MetaData.Name, hpaIntent.Spec.AppName)
							}

							//Decode the yaml to create a runtime.Object
							unstruct := &unstructured.Unstructured{}
							//Ignore the returned obj as we expect the data in unstruct
							_, err = hpaActionUtils.DecodeYAMLData(string(res), unstruct)
							if err != nil {
								log.Error("UpdateAppContext HPA .. Error in decoding deployment obj.", log.Fields{
									"hpa-intent-name": hpaIntent.MetaData.Name,
									"app-name":        hpaIntent.Spec.AppName,
									"yaml":            string(res),
									"yaml_unstruct":   unstruct,
									"err":             err})
								return pkgerrors.Wrapf(err, "UpdateAppContext HPA .. Error in decoding deployment obj. hpa-intent-name[%v] app-name[%v] err[%v]",
									hpaIntent.MetaData.Name, hpaIntent.Spec.AppName, err)
							}

							// addResource spec to Deployment spec
							err = addResourceSpecToDepSpec(unstruct, hpaConsumer, hpaResource)
							if err != nil {
								log.Error("UpdateAppContext HPA .. Error in adding resource-spec to deployment-spec.", log.Fields{
									"hpa-intent-name": hpaIntent.MetaData.Name,
									"app-name":        hpaIntent.Spec.AppName,
									"err":             err})
								return pkgerrors.Wrapf(err, "UpdateAppContext HPA .. Error in adding resource-spec to hpa-intent-name[%v] app-name[%v]",
									hpaIntent.MetaData.Name, hpaIntent.Spec.AppName)
							}

							// Marshal object back to yaml format (via json - seems to eliminate most clutter)
							j, err := json.Marshal(unstruct)
							if err != nil {
								log.Error("UpdateAppContext HPA .. Error marshalling resource to JSON", log.Fields{
									"error": err,
								})
								return pkgerrors.Wrapf(err, "UpdateAppContext HPA .. Error marshalling resource to JSON. hpa-intent-name[%v] app-name[%v] err[%v]",
									hpaIntent.MetaData.Name, hpaIntent.Spec.AppName, err)
							}
							y, err := jyaml.JSONToYAML(j)
							if err != nil {
								log.Error("UpdateAppContext HPA .. Error marshalling resource to YAML", log.Fields{
									"error": err,
								})
								return pkgerrors.Wrapf(err, "UpdateAppContext HPA .. Error marshalling resource to YAML. hpa-intent-name[%v] app-name[%v] err[%v]",
									hpaIntent.MetaData.Name, hpaIntent.Spec.AppName, err)
							}

							// Update resource in AppContext
							log.Info("UpdateAppContext HPA .. Update spec in db", log.Fields{"val": string(y)})
							err = updateResource(ac, resName, cluster, hpaIntent.Spec.AppName, string(y))
							if err != nil {
								log.Error("UpdateAppContext HPA .. error while updating app context resource handle", log.Fields{
									"error": err,
								})
								return pkgerrors.Wrapf(err, "UpdateAppContext HPA .. error while updating app context resource handle. hpa-intent-name[%v] app-name[%v] err[%v]",
									hpaIntent.MetaData.Name, hpaIntent.Spec.AppName, err)
							}
						}
					} // for clusterindex, cluster := range hpaclusters {
				} // for index, hpaResource := range hpaResources {
			} // for index, hpaConsumer := range hpaConsumers {
		} // for index, hpaIntent := range hpaIntents {
	} // for appIndex, eachApp := range allAppNames {

	log.Info("UpdateAppContext HPA .. end", log.Fields{
		"appcontext":              appContextID,
		"intent-name":             intentName,
		"project":                 project,
		"composite-app":           compositeApp,
		"composite-app-version":   compositeAppVersion,
		"deployment-intent-group": deploymentIntentGroup,
	})
	return nil
}

func getResource(ac appcontext.AppContext, name string, cluster string, app string) ([]byte, error) {
	log.Info("getResource .. start.", log.Fields{
		"res-name": name,
		"app-name": app,
		"cluster":  cluster})
	var byteRes []byte
	rh, err := ac.GetResourceHandle(app, cluster, name)
	if err != nil {
		log.Error("getResource .. App Context resource handle not found", log.Fields{
			"resource-name": name,
			"cluster":       cluster,
			"app":           app,
		})
		return nil, err
	}
	r, err := ac.GetValue(rh)
	if err != nil {
		log.Error("getResource .. Error retrieving resource from App Context", log.Fields{
			"error":           err,
			"resource handle": rh,
		})
		return nil, err
	}

	byteRes = []byte(fmt.Sprintf("%v", r.(interface{})))
	log.Info("getResource .. end.", log.Fields{
		"res-name": name,
		"app-name": app,
		"cluster":  cluster,
		"depSpec":  string(byteRes)})
	return byteRes, nil
}

func updateResource(ac appcontext.AppContext, name string, cluster string, app string, spec string) error {
	log.Info("updateResource .. start.", log.Fields{
		"res-name": name,
		"app-name": app,
		"cluster":  cluster})

	rh, err := ac.GetResourceHandle(app, cluster, name)
	if err != nil {
		log.Error("updateResource .. App Context resource handle not found", log.Fields{
			"resource-name": name,
			"cluster":       cluster,
			"app":           app,
		})
		return err
	}

	// Update resource in AppContext
	err = ac.UpdateResourceValue(rh, spec)
	if err != nil {
		log.Error("updateResource ..  updating app context resource handle", log.Fields{
			"error":           err,
			"resource handle": rh,
		})
		return err
	}
	log.Info("updateResource .. end.", log.Fields{
		"res-name": name,
		"app-name": app,
		"cluster":  cluster})

	return nil
}

func addResourceSpecToDepSpec(unstruct *unstructured.Unstructured, hpaConsumer hpaModel.HpaResourceConsumer, hpaResource hpaModel.HpaResourceRequirement) error {
	log.Warn("addResourceSpecToDepSpec .. start.", log.Fields{
		"hpa-consumer": hpaConsumer,
		"hpa-resource": hpaResource,
	})

	hpaCSpec := hpaConsumer.Spec
	hpaRSpec := hpaResource.Spec

	metadata, ok := unstruct.Object["metadata"].(map[string]interface{})
	if !ok {
		log.Error("addResourceSpecToDepSpec .. Error converting metadata to map.", nil)
		///return pkgerrors.Errorf("addResourceSpecToDepSpec .. Error converting metadata to map. consumer-spec-name[%v] resource-spec-name[%v]", hpaCSpec.Name, hpaRSpec.Resource.Name)
		return nil
	}
	// Look for the ObjectMeta
	podMeta := &metav1.ObjectMeta{}
	err := runtime.DefaultUnstructuredConverter.FromUnstructured(metadata, podMeta)
	if err != nil {
		log.Error("addResourceSpecToDepSpec .. ObjectMeta not found.", log.Fields{"err": err.Error()})
		//return pkgerrors.Wrapf(err, "addResourceSpecToDepSpec .. ObjectMeta not found. err[%v]", err)
		return nil
	}

	spec, ok := unstruct.Object["spec"].(map[string]interface{})
	if !ok {
		log.Error("addResourceSpecToDepSpec .. Error converting spec to map.", nil)
		//return pkgerrors.Errorf("addResourceSpecToDepSpec .. Error converting spec to map. consumer-spec-name[%v] resource-spec-name[%v]", hpaCSpec.Name, hpaRSpec.Resource.Name)
		return nil
	}

	replicas, ok := spec["replicas"]
	if !ok {
		log.Error("addResourceSpecToDepSpec .. Error fetching replicas frpm spec.", log.Fields{"spec": spec})
	} else {
		log.Info("addResourceSpecToDepSpec .. replicas fetched from spec.", log.Fields{"replicas": replicas})
		if hpaCSpec.Replicas > 0 {
			spec["replicas"] = hpaCSpec.Replicas
		}
	}

	template, ok := spec["template"].(map[string]interface{})
	if !ok {
		log.Error("addResourceSpecToDepSpec .. Error converting template to map.", nil)
		//return pkgerrors.Errorf("addResourceSpecToDepSpec .. Error converting template to map. consumer-spec-name[%v] resource-spec-name[%v]", hpaCSpec.Name, hpaRSpec.Resource.Name)
		return nil
	}
	// Look for the PodTemplateSpec
	podTemplateSpec := &corev1.PodTemplateSpec{}
	err = runtime.DefaultUnstructuredConverter.FromUnstructured(template, podTemplateSpec)
	if err != nil {
		log.Error("addResourceSpecToDepSpec .. PodTemplateSpec not found.", log.Fields{"pod-spec": spec, "template": template, "pod-template-spec": podTemplateSpec, "err": err})
		return pkgerrors.Wrapf(err, "addResourceSpecToDepSpec .. PodTemplateSpec not found")
	}

	log.Info("addResourceSpecToDepSpec .. adding resource-spec.", log.Fields{
		"hpa-consumer-spec":      hpaCSpec,
		"hpa-resource-spec":      hpaRSpec,
		"pod-spec":               spec,
		"pod-meta-data":          podMeta,
		"pod-template-spec-orig": podTemplateSpec,
	})

	// input validation of Consumer spec Name
	if (len(hpaCSpec.Name) == 0) || (hpaCSpec.Name != podMeta.Name) {
		log.Error("addResourceSpecToDepSpec .. consumer-spec Name mis-match",
			log.Fields{"hpaCSpec-spec-name": hpaCSpec.Name,
				"pod-spec-name": podMeta.Name})
		return fmt.Errorf("addResourceSpecToDepSpec .. consumer-spec Name mis-match. "+
			"hpaCSpec-spec-name[%v] pod-spec-name[%v]", hpaCSpec.Name, podMeta.Name)
	}

	// processing non-allocatable resource
	if !(*(hpaRSpec.Allocatable)) {
		rkey := hpaRSpec.Resource.NonAllocatableResources.Key
		rval := hpaRSpec.Resource.NonAllocatableResources.Value
		if len(podTemplateSpec.Spec.NodeSelector) == 0 {
			podTemplateSpec.Spec.NodeSelector = make(map[string]string)
		}

		// set pod spec node selector
		podTemplateSpec.Spec.NodeSelector[rkey] = rval

		log.Info("addResourceSpecToDepSpec .. non-allocatable-resource added to spec.",
			log.Fields{"pod-spec-name": podMeta.Name, "label-key": rkey, "label-val": rval})
	} else { // processing allocatable resource
		// input validation of consumer spec container name
		containerMatched := false
		podContainerNames := make([]string, 0)
		for _, container := range podTemplateSpec.Spec.Containers {
			podContainerNames = append(podContainerNames, container.Name)
			if hpaCSpec.ContainerName == container.Name {
				containerMatched = true
			}
		}
		if !containerMatched {
			log.Error("addResourceSpecToDepSpec .. consumer-spec container-name mis-match",
				log.Fields{"consumer-spec-container-name": hpaCSpec.ContainerName,
					"pod-containers": podContainerNames})
			return fmt.Errorf("addResourceSpecToDepSpec .. consumer-spec container-name mis-match. "+
				"consumer-spec-container-name[%v] pod-containers[%v]", hpaCSpec.ContainerName, podContainerNames)
		}

		// Get the resource requirements for the spec
		for index, container := range podTemplateSpec.Spec.Containers {
			// container name matches
			if hpaCSpec.ContainerName == container.Name {
				// Hpa resource spec parametes
				rname := hpaRSpec.Resource.AllocatableResources.Name
				rreq := hpaRSpec.Resource.AllocatableResources.Requests
				rlimit := hpaRSpec.Resource.AllocatableResources.Limits

				// convert memory units from MB to Bytes
				if rname == string(corev1.ResourceMemory) {
					// convert Mega Bytes to Bytes
					rreq *= 1000 * 1000
					rlimit *= 1000 * 1000
				}

				var resourceReq corev1.ResourceList = make(corev1.ResourceList)
				var resourceLimits corev1.ResourceList = make(corev1.ResourceList)

				// Loop through container's existing resource spec
				for key, val := range container.Resources.Requests {
					resourceReq[key] = val
				}
				for key, val := range container.Resources.Limits {
					resourceLimits[key] = val
				}

				// Fill container with new resource spec
				resourceReq[corev1.ResourceName(rname)] = *(resource.NewQuantity(rreq, ""))
				if hpaRSpec.Resource.AllocatableResources.Limits > 0 {
					resourceLimits[corev1.ResourceName(rname)] = *(resource.NewQuantity(rlimit, ""))
				}

				container.Resources.Requests = resourceReq
				container.Resources.Limits = resourceLimits

				// set constainer spec
				podTemplateSpec.Spec.Containers[index] = container

				log.Info("addResourceSpecToDepSpec .. allocatable-resource added to spec.",
					log.Fields{"pod-spec-name": podMeta.Name,
						"container-name": container.Name,
						"resource-name":  rname,
						"reqs":           container.Resources.Requests,
						"limits":         container.Resources.Limits})
			}
		} // for index, container := range podTemplateSpec.Spec.Containers
	} // if !(*(hpaRSpec.Allocatable))

	updatedTemplate, err := runtime.DefaultUnstructuredConverter.ToUnstructured(podTemplateSpec)
	if err != nil {
		log.Error("addResourceSpecToDepSpec .. unable to convert podTemplateSpec.", log.Fields{"err": err})
		return pkgerrors.Wrapf(err, "addResourceSpecToDepSpec .. unable to convert podTemplateSpec")
	}

	//Set the template
	spec["template"] = updatedTemplate

	log.Warn("addResourceSpecToDepSpec .. end.", log.Fields{
		"hpa-consumer":              hpaConsumer,
		"hpa-resource":              hpaResource,
		"pod-spec":                  spec,
		"pod-meta-data":             podMeta,
		"pod-template-spec-orig":    podTemplateSpec,
		"pod-template-spec-updated": updatedTemplate,
	})

	return nil
}
