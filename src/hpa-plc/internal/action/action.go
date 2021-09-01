// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2020 Intel Corporation

package action

import (
	"context"
	"sort"
	"strings"

	pkgerrors "github.com/pkg/errors"

	"github.com/open-ness/EMCO/src/orchestrator/pkg/appcontext"
	log "github.com/open-ness/EMCO/src/orchestrator/pkg/infra/logutils"
	orchUtils "github.com/open-ness/EMCO/src/orchestrator/pkg/infra/utils"
	orchModuleLib "github.com/open-ness/EMCO/src/orchestrator/pkg/module"

	clmcontrollerpb "github.com/open-ness/EMCO/src/clm/pkg/grpc/controller-eventchannel"
	hpaModel "github.com/open-ness/EMCO/src/hpa-plc/pkg/model"
	hpaModuleLib "github.com/open-ness/EMCO/src/hpa-plc/pkg/module"
	intentRs "github.com/open-ness/EMCO/src/hpa-plc/pkg/resources"
	hpaUtils "github.com/open-ness/EMCO/src/hpa-plc/pkg/utils"
)

// FilterClusters .. Filter clusters based on hpa-intents attached to the AppContext ID
func FilterClusters(appContextID string) error {
	var ac appcontext.AppContext
	log.Warn("FilterClusters .. start", log.Fields{"appContextID": appContextID})
	_, err := ac.LoadAppContext(appContextID)
	if err != nil {
		log.Error("FilterClusters .. Error getting AppContext", log.Fields{"appContextID": appContextID})
		return pkgerrors.Wrapf(err, "FilterClusters .. Error getting AppContext with Id: %v", appContextID)
	}

	caMeta, err := ac.GetCompositeAppMeta()
	if err != nil {
		log.Error("FilterClusters .. Error getting metadata for AppContext", log.Fields{"appContextID": appContextID})
		return pkgerrors.Wrapf(err, "FilterClusters .. Error getting metadata for AppContext with Id: %v", appContextID)
	}

	project := caMeta.Project
	compositeApp := caMeta.CompositeApp
	compositeAppVersion := caMeta.Version
	deploymentIntentGroup := caMeta.DeploymentIntentGroup

	log.Info("FilterClusters .. AppContext details", log.Fields{"project": project, "compositeApp": compositeApp, "deploymentGroup": deploymentIntentGroup})

	// Get all apps in this composite app
	apps, err := orchModuleLib.NewAppClient().GetApps(project, compositeApp, compositeAppVersion)
	if err != nil {
		log.Error("FilterClusters .. Not finding the compositeApp attached apps", log.Fields{"appContextID": appContextID, "compositeApp": compositeApp})
		return pkgerrors.Wrapf(err, "FilterClusters .. Not finding the compositeApp[%s] attached apps", compositeApp)
	}
	allAppNames := make([]string, 0)
	for _, a := range apps {
		allAppNames = append(allAppNames, a.Metadata.Name)
	}
	log.Info("FilterClusters .. Applications attached to compositeApp",
		log.Fields{"appContextID": appContextID, "project": project, "compositeApp": compositeApp, "deploymentGroup": deploymentIntentGroup, "app-names": allAppNames})

	// Dump group-clusters map
	for index, eachApp := range allAppNames {
		grpMap, _ := ac.GetClusterGroupMap(eachApp)
		log.Warn("FilterClusters .. ClusterGroupMap dump before invoking HPA Placement filtering",
			log.Fields{"index": index, "appContextID": appContextID, "appName": eachApp, "group-map_size": len(grpMap), "groupMap": grpMap})
	}

	// Iterate through all apps of the Composite App
	for appIndex, eachApp := range allAppNames {
		// Handle all hpa Intents of the app
		hpaIntents, err := hpaModuleLib.NewHpaPlacementClient().GetAllIntentsByApp(eachApp, project, compositeApp, compositeAppVersion, deploymentIntentGroup)
		if err != nil {
			log.Error("FilterClusters .. Error getting hpa Intents", log.Fields{"project": project, "compositeApp": compositeApp, "deploymentGroup": deploymentIntentGroup})
			return pkgerrors.Wrapf(err, "FilterClusters .. Error getting hpa Intents for project[%v] compositeApp[%v] compositeVersion[%v] deploymentGroup[%v] not found", project, compositeApp, compositeAppVersion, deploymentIntentGroup)
		}

		// Continue with other apps as the current app does not have intents associated
		if len(hpaIntents) == 0 {
			log.Info("FilterClusters .. No hpa Intents", log.Fields{"project": project, "compositeApp": compositeApp, "deploymentGroup": deploymentIntentGroup, "app-nme": eachApp})
			continue
		}

		// initialize
		var clusterResourceObjMap = make(intentRs.ClusterResourceObjMap)
		var clusterResourceInfoMap = make(intentRs.ClusterResourceInfoMap)
		var kubeResToHpaResourceMap = make(map[string](hpaModel.HpaResourceRequirement))

		log.Info("FilterClusters .. Intents attached to app", log.Fields{"app-index": appIndex,
			"app-name":        eachApp,
			"len-hpa-intents": len(hpaIntents), "hpa-intents": hpaIntents})
		for index, hpaIntent := range hpaIntents {
			log.Info("FilterClusters .. hpaIntents filtering details => ", log.Fields{
				"app-index":               appIndex,
				"intent-index":            index,
				"hpa-intent":              hpaIntent,
				"project":                 project,
				"composite-app":           compositeApp,
				"composite-app-version":   compositeAppVersion,
				"deployment-intent-group": deploymentIntentGroup,
				"hpa-intent-name":         hpaIntent.MetaData.Name,
				"app-name":                hpaIntent.Spec.AppName,
			})

			grpMap, err := ac.GetClusterGroupMap(hpaIntent.Spec.AppName)
			if err != nil {
				log.Error("FilterClusters .. Error getting GroupMap for app", log.Fields{"appName": hpaIntent.Spec.AppName, "groupMap": grpMap})
				return pkgerrors.Wrapf(err, "FilterClusters .. Error getting GroupMap for app[%s], groupMap[%s]", hpaIntent.Spec.AppName, grpMap)
			}
			log.Info("FilterClusters .. ClusterGroupMap", log.Fields{"GroupMap": grpMap})
			for gn, clusters := range grpMap {
				log.Info("FilterClusters .. GetClusterGroupMap details.", log.Fields{"group_number": gn, "anyof-clusters": clusters})

				// Final HPA Qualified clusters list
				hpaQualifiedClusterToNodesMap := make(map[string]([]string))
				hpaQualifiedClusters := make([]string, 0)
				hpaQualifiedNodes := make([]string, 0)

				// Get all clusters for the current App from the AppContext
				getclusters, err := ac.GetClusterNames(hpaIntent.Spec.AppName)
				log.Info("FilterClusters .. GetClusterNames for app Info.", log.Fields{
					"clusters":                getclusters,
					"project":                 project,
					"composite app":           compositeApp,
					"composite app version":   compositeAppVersion,
					"deployment-intent-group": deploymentIntentGroup,
					"hpa-intent-name":         hpaIntent.MetaData.Name,
					"app-name":                hpaIntent.Spec.AppName,
					"err":                     err,
				})

				// Interate through all clusters and populate cluster info
				for _, cl := range clusters {
					// Populate cluster info
					var clusterResourceInfo intentRs.ClusterResourceInfo
					clusterResourceInfo.ClusterName = cl
					clusterResourceInfoMap[cl] = clusterResourceInfo
				} // for clusters

				// Initialize cluster/node resource Info
				err = initializeResourceInfo(context.TODO(), &hpaIntent, &clusterResourceInfoMap, &clusterResourceObjMap)
				if err != nil {
					log.Error("FilterClusters .. Hpa Intent initializeResourceInfo error=> ", log.Fields{
						"clusters": clusters,
						"err":      err})

					return pkgerrors.Wrapf(err, "filterResource .. Hpa Intent initializeResourceInfo error intent-name[%v] app-name[%v] err[%v]", hpaIntent.MetaData.Name, hpaIntent.Spec.AppName, err)
				}
				log.Info("FilterClusters .. Hpa Intent initializeResourceInfo ", log.Fields{
					"intent-name":           hpaIntent.MetaData.Name,
					"app-name":              hpaIntent.Spec.AppName,
					"clusters":              clusters,
					"clusterResourceObjMap": clusterResourceObjMap})

				// Handle all hpa Consumers
				hpaConsumers, err := hpaModuleLib.NewHpaPlacementClient().GetAllConsumers(project, compositeApp, compositeAppVersion, deploymentIntentGroup, hpaIntent.MetaData.Name)
				if err != nil {
					log.Error("FilterClusters .. Error GetAllConsumers.", log.Fields{
						"hpa-intent": hpaIntent,
						"err":        err})
					return pkgerrors.Wrapf(err, "FilterClusters .. Error GetAllConsumers. Intent[%v] for project[%v] comp-app[%v] comp-app-version[%v] not found", hpaIntent.MetaData.Name, project, compositeApp, compositeAppVersion)
				}

				// Continue with other apps as the current app does not have intents associated
				if len(hpaConsumers) == 0 {
					log.Info("FilterClusters .. No hpa Consumers Resources for the intent", log.Fields{"project": project, "compositeApp": compositeApp, "deploymentGroup": deploymentIntentGroup, "app-nme": eachApp, "hpa-intent": hpaIntent})
					continue
				}

				// Handle hpa consumer
				log.Info("FilterClusters .. List Of Consumers", log.Fields{"len_hpa-consumers": len(hpaConsumers), "hpa-consumers": hpaConsumers})
				for index, hpaConsumer := range hpaConsumers {
					log.Info("FilterClusters .. HpaConsumers .. start", log.Fields{"index": index, "app": hpaIntent.Spec.AppName, "hpa-intent": hpaIntent,
						"hpa-consumer": hpaConsumer})

					// Handle all hpa Resurces
					hpaResources, err := hpaModuleLib.NewHpaPlacementClient().GetAllResources(project, compositeApp, compositeAppVersion, deploymentIntentGroup, hpaIntent.MetaData.Name, hpaConsumer.MetaData.Name)
					if err != nil {
						log.Error("FilterClusters .. Error GetAllResources.", log.Fields{
							"hpa-intent":   hpaIntent,
							"hpa-consumer": hpaConsumer,
							"err":          err})

						return pkgerrors.Wrapf(err, "FilterClusters .. Error GetAllResources. Intent[%v] consumer[%v] for project[%v] comp-app[%v] comp-app-version[%v] not found", hpaIntent.MetaData.Name, hpaConsumer.MetaData.Name, project, compositeApp, compositeAppVersion)
					}

					// Continue with other apps as the current app does not have intents associated
					if len(hpaResources) == 0 {
						log.Info("FilterClusters .. No hpa Resources for the consumer", log.Fields{"project": project, "compositeApp": compositeApp, "deploymentGroup": deploymentIntentGroup, "app-nme": eachApp, "hpa-intent": hpaIntent, "hpa-consumer": hpaConsumer})
						continue
					}

					log.Warn("FilterClusters .. Placement start",
						log.Fields{"appContextID": appContextID,
							"hpa-intent":   hpaIntent,
							"hpa-consumer": hpaConsumer,
							"app-name":     hpaIntent.Spec.AppName,
							"number-hpa-qualified-cluster-to-nodes-map": len(hpaQualifiedClusterToNodesMap),
							"hpa-qualified-cluster-to-nodes-map":        hpaQualifiedClusterToNodesMap,
							"number-qualified-clusters":                 len(hpaQualifiedClusters),
							"qualified-clusters":                        hpaQualifiedClusters,
							"number-qualified-nodes":                    len(hpaQualifiedNodes),
							"qualified-nodes":                           hpaQualifiedNodes,
							"input-clusters":                            clusters})

					var isNonAllocResPresent bool = false
					var isAllocResPresent bool = false
					// Handle NonAllocatable hpa resource filtering
					for index, hpaResource := range hpaResources {
						log.Info("FilterClusters .. HpaNonAllocResources .. start", log.Fields{"index": index, "app": hpaIntent.Spec.AppName, "hpa-intent": hpaIntent,
							"hpa-consumer": hpaConsumer, "hpa-resource": hpaResource, "clusters": clusters})
						hpaResourceLocal := hpaResource
						if !(*(hpaResource.Spec.Allocatable)) {
							isNonAllocResPresent = true
							status, qualifiedClusterToNodesMap, err := filterNonAllocResource(hpaIntent.MetaData.Name, hpaIntent.Spec.AppName, &hpaResourceLocal, &clusterResourceObjMap, &clusterResourceInfoMap, clusters)
							if !status {
								log.Error("FilterClusters ..  filterNonAllocResource Failed. None of the clusters match the hpa-nonalloc-resource rules!!",
									log.Fields{"appContextID": appContextID,
										"intent-name":                        hpaIntent.MetaData.Name,
										"app-name":                           hpaIntent.Spec.AppName,
										"hpa-consumer":                       hpaConsumer,
										"hpa-resource":                       hpaResource,
										"len-qualified-cluster-to-nodes-map": len(qualifiedClusterToNodesMap), "qualified-cluster-to-nodes-map": qualifiedClusterToNodesMap,
										"input-clusters": clusters, "err": err})
								return pkgerrors.Errorf("FilterClusters .. filterNonAllocResource Failed. None of the clusters[%v] match the hpa-nonalloc-resource rules for appContextID[%v] intent-name[%v] app-name[%v] hpa-consumer[%v] hpa-resource[%v]",
									clusters, appContextID, hpaIntent.MetaData.Name, hpaIntent.Spec.AppName, hpaConsumer, hpaResource)

							} else {
								log.Info("FilterClusters .. filterNonAllocResource success",
									log.Fields{"hpa-intent-name": hpaIntent.MetaData.Name,
										"app-name":     hpaIntent.Spec.AppName,
										"hpa-consumer": hpaConsumer,
										"hpaResource":  hpaResource,
										"number-hpa-qualified-cluster-to-nodes-map": len(hpaQualifiedClusterToNodesMap),
										"hpa-qualified-cluster-to-nodes-map":        hpaQualifiedClusterToNodesMap,
										"number-qualified-cluster-to-nodes-map":     len(qualifiedClusterToNodesMap),
										"qualified-cluster-to-nodes-map":            qualifiedClusterToNodesMap,
									})
							}

							if len(hpaQualifiedClusterToNodesMap) == 0 {
								hpaQualifiedClusterToNodesMap = make(map[string]([]string))
								hpaQualifiedClusterToNodesMap = qualifiedClusterToNodesMap
							}

							// intersect NonAllocRes Node maps
							hpaQualifiedClusters = make([]string, 0)
							isCommonNonAllocClustersPresent := false
							for hpaCluster, hpaNodes := range hpaQualifiedClusterToNodesMap {
								for hl, hn := range qualifiedClusterToNodesMap {
									if hpaCluster == hl {
										hpaQualifiedNodes := hpaUtils.GetSliceIntersect(hpaNodes, hn)
										if len(hpaQualifiedNodes) > 0 {
											isCommonNonAllocClustersPresent = true
											if !hpaUtils.IsInSlice(hl, hpaQualifiedClusters) {
												hpaQualifiedClusters = append(hpaQualifiedClusters, hl)
											}
											hpaQualifiedClusterToNodesMap[hpaCluster] = hpaQualifiedNodes
											log.Info("FilterClusters .. Emco HpaNonAllocResource Filter Cluster Node candidates for resource: ",
												log.Fields{"hpa-intent-name": hpaIntent.MetaData.Name,
													"app-name":    hpaIntent.Spec.AppName,
													"hpaResource": hpaResource,
													"number-hpa-qualified-cluster-to-nodes-map": len(hpaQualifiedClusterToNodesMap),
													"hpa-qualified-cluster-to-nodes-map":        hpaQualifiedClusterToNodesMap,
													"number-qualified-cluster-to-nodes-map":     len(qualifiedClusterToNodesMap),
													"qualified-cluster-to-nodes-map":            qualifiedClusterToNodesMap,
													"number-qualified-clusters":                 len(hpaQualifiedClusters),
													"qualified-clusters":                        hpaQualifiedClusters,
													"number-qualified-nodes":                    len(hpaQualifiedNodes),
													"qualified-nodes":                           hpaQualifiedNodes})
										} else {
											// delete clusterName entry from hpaQualifiedClusterToNodesMap
											delete(hpaQualifiedClusterToNodesMap, hpaCluster)

											// delete clusterName from hpaQualifiedClusters
											for idx, val := range hpaQualifiedClusters {
												if val == hpaCluster {
													hpaQualifiedClusters = append(hpaQualifiedClusters[:idx], hpaQualifiedClusters[idx+1:]...)
													break
												}
											}
										}
									} //if hpaCluster == hl {
								} //for hl, hn
							} //for hpaCluster, hpaNodes
							if !isCommonNonAllocClustersPresent {
								log.Error("FilterClusters .. No common clusters. None of the clusters match the hpa-non-alloc-resource rules!! ",
									log.Fields{"appContextID": appContextID, "hpa-intent-name": hpaIntent.MetaData.Name, "isNonAllocResPresent": isNonAllocResPresent,
										"app-name":                        hpaIntent.Spec.AppName,
										"hpaResource":                     hpaResource,
										"isCommonNonAllocClustersPresent": isCommonNonAllocClustersPresent,
										"number-hpa-qualified-cluster-to-nodes-map": len(hpaQualifiedClusterToNodesMap),
										"hpa-qualified-cluster-to-nodes-map":        hpaQualifiedClusterToNodesMap,
										"number-qualified-cluster-to-nodes-map":     len(qualifiedClusterToNodesMap),
										"qualified-cluster-to-nodes-map":            qualifiedClusterToNodesMap,
										"number-qualified-clusters":                 len(hpaQualifiedClusters),
										"qualified-clusters":                        hpaQualifiedClusters,
										"number-qualified-nodes":                    len(hpaQualifiedNodes),
										"qualified-nodes":                           hpaQualifiedNodes})

								return pkgerrors.Errorf("FilterClusters .. No common clusters. None of the clusters match the hpa-non-alloc-resource rules for appContextID[%v] isNonAllocResPresent[%v] intent-name[%v] app-name[%v] hpa-consumer[%v] hpa-resource[%v]",
									appContextID, isNonAllocResPresent, hpaIntent.MetaData.Name, hpaIntent.Spec.AppName, hpaConsumer, hpaResource)
							}
						} // for hpa-resource .. Handle NonAllocatable hpa resource filtering
					} // for hpa-resource .. Handle Non-Allocatable hpa resource filtering

					if isNonAllocResPresent {
						if len(hpaQualifiedClusters) > 0 {
							log.Info("FilterClusters .. Emco hpa-non-alloc-resource Deployment Cluster candidates: ",
								log.Fields{"appContextID": appContextID,
									"hpa-intent": hpaIntent,
									"app-name":   hpaIntent.Spec.AppName,
									"number-hpa-qualified-cluster-to-nodes-map": len(hpaQualifiedClusterToNodesMap),
									"hpa-qualified-cluster-to-nodes-map":        hpaQualifiedClusterToNodesMap,
									"number-qualified-clusters":                 len(hpaQualifiedClusters),
									"qualified-clusters":                        hpaQualifiedClusters,
									"input-clusters":                            clusters})
						} else {
							log.Error("FilterClusters .. Failure Before checking AllocResource rules .. None of the clusters match the hpa-non-alloc-resource rules!!",
								log.Fields{"appContextID": appContextID,
									"hpa-intent": hpaIntent,
									"number-hpa-qualified-cluster-to-nodes-map": len(hpaQualifiedClusterToNodesMap),
									"hpa-qualified-cluster-to-nodes-map":        hpaQualifiedClusterToNodesMap,
									"number-qualified-clusters":                 len(hpaQualifiedClusters),
									"qualified-clusters":                        hpaQualifiedClusters})
							return pkgerrors.Errorf("FilterClusters .. None of the clusters match the hpa-non-alloc-resource rules for appContextID[%v] intent-name[%v] app-name[%v]",
								appContextID, hpaIntent.MetaData.Name, hpaIntent.Spec.AppName)
						}
					}

					// Handle Allocatable hpa resource filtering
					if hpaConsumer.Spec.Replicas <= 0 {
						log.Info("FilterClusters .. replicas are not specified in HpaConsumer spec, setting it to 1", log.Fields{"hpa-intent": hpaIntent,
							"hpa-consumer": hpaConsumer})
						hpaConsumer.Spec.Replicas = 1
					}
					var replicaCount int64
					for replicaCount = 1; replicaCount <= hpaConsumer.Spec.Replicas; replicaCount++ {
						log.Info("FilterClusters .. List Of Resources", log.Fields{"replica-count": replicaCount, "hpa-consumer": hpaConsumer, "len_hpa-resources": len(hpaResources), "hpa-resources": hpaResources})
						for index, hpaResource := range hpaResources {
							log.Info("FilterClusters .. HpaAllocResources .. start", log.Fields{"index": index, "app": hpaIntent.Spec.AppName, "hpa-intent": hpaIntent,
								"hpa-consumer": hpaConsumer, "hpa-resource": hpaResource, "clusters": clusters})
							hpaResourceLocal := hpaResource

							if *(hpaResource.Spec.Allocatable) {
								isAllocResPresent = true
								kubeResToHpaResourceMap[hpaResource.Spec.Resource.Name] = hpaResource
								status, qualifiedClusterToNodesMap, err := filterAllocResource(hpaIntent.MetaData.Name, hpaIntent.Spec.AppName, &hpaResourceLocal, &clusterResourceObjMap, &clusterResourceInfoMap, clusters, (replicaCount == hpaConsumer.Spec.Replicas))
								if !status {
									log.Error("FilterClusters .. filterAllocResource Failed .. None of the clusters match the hpa-alloc-resource rules!!",
										log.Fields{"appContextID": appContextID, "isNonAllocResPresent": isNonAllocResPresent,
											"intent-name":                        hpaIntent.MetaData.Name,
											"app-name":                           hpaIntent.Spec.AppName,
											"replica-count":                      replicaCount,
											"hpa-consumer":                       hpaConsumer,
											"hpa-resource":                       hpaResource,
											"len-qualified-cluster-to-nodes-map": len(qualifiedClusterToNodesMap), "qualified-cluster-to-nodes-map": qualifiedClusterToNodesMap,
											"input-clusters": clusters, "err": err})

									return pkgerrors.Wrapf(err, "FilterClusters .. filterAllocResource Failed .. None of the clusters match the hpa-alloc-resource rules for appContextID[%v] isNonAllocResPresent[%v] intent-name[%v] app-name[%v] replica-count[%v] hpa-consumer[%v] hpa-resource[%v] err[%v]",
										appContextID, isNonAllocResPresent, hpaIntent.MetaData.Name, hpaIntent.Spec.AppName, replicaCount, hpaConsumer, hpaResource, err)
								} else {
									log.Info("FilterClusters .. filterAllocResource success",
										log.Fields{"hpa-intent-name": hpaIntent.MetaData.Name,
											"app-name":      hpaIntent.Spec.AppName,
											"replica-count": replicaCount,
											"hpa-consumer":  hpaConsumer,
											"hpaResource":   hpaResource,
											"number-hpa-qualified-cluster-to-nodes-map": len(hpaQualifiedClusterToNodesMap),
											"hpa-qualified-cluster-to-nodes-map":        hpaQualifiedClusterToNodesMap,
											"number-qualified-cluster-to-nodes-map":     len(qualifiedClusterToNodesMap),
											"qualified-cluster-to-nodes-map":            qualifiedClusterToNodesMap,
										})
								}

								if len(hpaQualifiedClusterToNodesMap) == 0 {
									hpaQualifiedClusterToNodesMap = make(map[string]([]string))
									hpaQualifiedClusterToNodesMap = qualifiedClusterToNodesMap
								}

								// intersect AllocRes Node maps
								isCommonAllocClustersPresent := false
								for hpaCluster, hpaNodes := range hpaQualifiedClusterToNodesMap {
									for hl, hn := range qualifiedClusterToNodesMap {
										if hpaCluster == hl {
											hpaQualifiedNodes := hpaUtils.GetSliceIntersect(hpaNodes, hn)
											if len(hpaQualifiedNodes) > 0 {
												isCommonAllocClustersPresent = true
												if !hpaUtils.IsInSlice(hl, hpaQualifiedClusters) {
													hpaQualifiedClusters = append(hpaQualifiedClusters, hl)
												}

												hpaQualifiedClusterToNodesMap[hpaCluster] = hpaQualifiedNodes

												// Update node accounting for the matched cluster
												// Fetch allocatable resource obj
												rsAllocatable := clusterResourceObjMap[hpaCluster].AllocatableRs
												nodeMap := make(map[string]int64)
												for k, v := range rsAllocatable.GetNodeResMap(hpaResource.Spec.Resource.Name) {
													nodeMap[k] = v
												}
												if len(nodeMap) > 0 {
													nodeChosenForAccounting := findNodeForAccounting(nodeMap, hpaQualifiedNodes, hpaResource)
													if nodeChosenForAccounting != "" {
														rsAllocatable.UpdateNodeResourceCounts(nodeChosenForAccounting, hpaResource)
													}

													log.Info("filterAllocResource .. Updated node accounting for the matched cluster.", log.Fields{
														"nodeChosenForAccounting": nodeChosenForAccounting,
														"cluster-chosen":          hpaCluster,
														"resource-name":           hpaResource.Spec.Resource.Name,
														"nodeMapOrig":             nodeMap,
														"nodeMapUpdated":          rsAllocatable.GetNodeResMap(hpaResource.Spec.Resource.Name),
														"QualifiedNodes":          hpaQualifiedNodes,
														"hpaResource":             hpaResource})
												}

												log.Info("FilterClusters .. Emco HpaAllocResource Filter Cluster Node candidates for resource: ",
													log.Fields{"hpa-intent-name": hpaIntent.MetaData.Name,
														"app-name":      hpaIntent.Spec.AppName,
														"replica-count": replicaCount,
														"hpaResource":   hpaResource,
														"number-hpa-qualified-cluster-to-nodes-map": len(hpaQualifiedClusterToNodesMap),
														"hpa-qualified-cluster-to-nodes-map":        hpaQualifiedClusterToNodesMap,
														"number-qualified-cluster-to-nodes-map":     len(qualifiedClusterToNodesMap),
														"qualified-cluster-to-nodes-map":            qualifiedClusterToNodesMap,
														"number-qualified-clusters":                 len(hpaQualifiedClusters),
														"qualified-clusters":                        hpaQualifiedClusters,
														"number-qualified-nodes":                    len(hpaQualifiedNodes),
														"qualified-nodes":                           hpaQualifiedNodes})
											} else {
												// delete clusterName entry from hpaQualifiedClusterToNodesMap
												delete(hpaQualifiedClusterToNodesMap, hpaCluster)

												// delete clusterName from hpaQualifiedClusters
												for idx, val := range hpaQualifiedClusters {
													if val == hpaCluster {
														hpaQualifiedClusters = append(hpaQualifiedClusters[:idx], hpaQualifiedClusters[idx+1:]...)
														break
													}
												}
											}
										} //if hpaCluster == hl {
									} //for hl, hn
								} //for hpaCluster, hpaNodes

								if !isCommonAllocClustersPresent {
									log.Error("FilterClusters .. No common clusters. None of the clusters match the hpa-alloc-resource rules!! ",
										log.Fields{"appContextID": appContextID, "hpa-intent-name": hpaIntent.MetaData.Name, "isNonAllocResPresent": isNonAllocResPresent,
											"app-name":                     hpaIntent.Spec.AppName,
											"replica-count":                replicaCount,
											"hpaResource":                  hpaResource,
											"isCommonAllocClustersPresent": isCommonAllocClustersPresent,
											"number-hpa-qualified-cluster-to-nodes-map": len(hpaQualifiedClusterToNodesMap),
											"hpa-qualified-cluster-to-nodes-map":        hpaQualifiedClusterToNodesMap,
											"number-qualified-cluster-to-nodes-map":     len(qualifiedClusterToNodesMap),
											"qualified-cluster-to-nodes-map":            qualifiedClusterToNodesMap,
											"number-qualified-clusters":                 len(hpaQualifiedClusters),
											"qualified-clusters":                        hpaQualifiedClusters,
											"number-qualified-nodes":                    len(hpaQualifiedNodes),
											"qualified-nodes":                           hpaQualifiedNodes})

									return pkgerrors.Errorf("FilterClusters .. No common clusters. None of the clusters match the hpa-alloc-resource rules for appContextID[%v] isNonAllocResPresent[%v] intent-name[%v] app-name[%v] replica-count[%d] hpa-consumer[%v] hpa-resource[%v]",
										appContextID, isNonAllocResPresent, hpaIntent.MetaData.Name, hpaIntent.Spec.AppName, replicaCount, hpaConsumer, hpaResource)
								}
							}
						} // for hpa-resource .. Handle Allocatable hpa resource filtering
					} // for replicaCount

					if len(hpaQualifiedClusters) > 0 {
						log.Warn("FilterClusters .. Placement end. Emco Placement qualified Deployment Cluster candidates: ",
							log.Fields{"appContextID": appContextID,
								"isNonAllocResPresent": isNonAllocResPresent, "isAllocResPresent": isAllocResPresent,
								"hpa-intent": hpaIntent,
								"app-name":   hpaIntent.Spec.AppName,
								"number-hpa-qualified-cluster-to-nodes-map": len(hpaQualifiedClusterToNodesMap),
								"hpa-qualified-cluster-to-nodes-map":        hpaQualifiedClusterToNodesMap,
								"number-hpa-qualified-clusters":             len(hpaQualifiedClusters),
								"hpa-qualified-clusters":                    hpaQualifiedClusters,
								"input-clusters":                            clusters})
					} else {
						log.Error("FilterClusters .. Placement end. Failed Emco Placement: None of the clusters match the intent rules!!",
							log.Fields{"appContextID": appContextID,
								"isNonAllocResPresent": isNonAllocResPresent, "isAllocResPresent": isAllocResPresent,
								"hpa-intent": hpaIntent,
								"app-name":   hpaIntent.Spec.AppName,
								"number-hpa-qualified-cluster-to-nodes-map": len(hpaQualifiedClusterToNodesMap),
								"hpa-qualified-cluster-to-nodes-map":        hpaQualifiedClusterToNodesMap,
								"number-hpa-qualified-clusters":             len(hpaQualifiedClusters),
								"hpa-qualified-clusters":                    hpaQualifiedClusters,
								"input-clusters":                            clusters})
						return pkgerrors.Errorf("FilterClusters .. None of the clusters[%v] match the hpa-resource rules for appContextID[%v] isNonAllocResPresent[%v] isAllocResPresent[%v] intent-name[%v] app-name[%v]",
							clusters, appContextID, isNonAllocResPresent, isAllocResPresent,
							hpaIntent.MetaData.Name, hpaIntent.Spec.AppName)
					}
				}

				// Delete extra clusters not matching HPA rules
				clustersExtra := orchUtils.GetSliceSubtract(clusters, hpaQualifiedClusters)
				log.Info("filterResource .. Extra clusters to be deleted", log.Fields{"app-name": hpaIntent.Spec.AppName, "group-name": gn, "input-clusters": clusters, "hpa-clusters": hpaQualifiedClusters, "extra-clusters": clustersExtra})
				for i, clExtra := range clustersExtra {
					log.Info("filterResource .. Delete non-qualified cluster", log.Fields{"cluster-index": i, "cluster": clExtra, "appname": hpaIntent.Spec.AppName})

					// Delete the cluster from AppContext if not matching HPA rules
					ch, err := ac.GetClusterHandle(hpaIntent.Spec.AppName, clExtra)
					if err != nil {
						log.Error("filterResource .. Unable to get cluster handle", log.Fields{"cluster": clExtra, "appname": hpaIntent.Spec.AppName})
						return pkgerrors.Wrapf(err, "filterResource .. Unable to get cluster handle. appName[%s] cluster[%s]", hpaIntent.Spec.AppName, clExtra)
					}
					err = ac.DeleteCluster(ch)
					if err != nil {
						log.Error("filterResource .. Unable to delete cluster", log.Fields{"cluster": clExtra, "appname": hpaIntent.Spec.AppName})
						return pkgerrors.Wrapf(err, "filterResource .. Unable to delete cluster. appName[%s] cluster[%s]", hpaIntent.Spec.AppName, clExtra)
					}
				}
			} // for gn, clusters := range grpMap {
		} // for hpa-intent
	} // for index, eachApp := range allAppNames {

	// Dump group-clusters map
	for index, eachApp := range allAppNames {
		grpMap, _ := ac.GetClusterGroupMap(eachApp)
		log.Warn("FilterClusters .. ClusterGroupMap dump after invoking HPA Placement filtering.",
			log.Fields{"index": index, "appContextID": appContextID,
				"project": project, "compositeApp": compositeApp,
				"all-app-names": allAppNames, "deploymentGroup": deploymentIntentGroup,
				"appName": eachApp, "group-map_size": len(grpMap), "groupMap": grpMap})
	}
	return nil
}

//filterAllocResource ... filter w.r.t hpa resource
func filterAllocResource(intentName string, appName string, hpaResource *hpaModel.HpaResourceRequirement, clusterResourceObjMap *intentRs.ClusterResourceObjMap, clusterResourceInfoMap *intentRs.ClusterResourceInfoMap, clusters []string, rollbackNeeded bool) (bool, map[string]([]string), error) {
	log.Info("filterAllocResource .. start", log.Fields{
		"intent-name":    intentName,
		"app-name":       appName,
		"clusters":       clusters,
		"rollbackNeeded": rollbackNeeded,
		"hpa-resource":   hpaResource,
	})

	var clusterToNodesMap map[string]([]string)
	clusterToNodesMap = make(map[string]([]string))
	matched := false

	// Populate cluster/node resource Info
	err := PopulateClustersResources(context.TODO(), hpaResource, clusterResourceInfoMap, clusterResourceObjMap)
	if err != nil {
		log.Error("filterAllocResource .. Hpa Intent PopulateClustersResources error=> ", log.Fields{
			"intent-name": intentName,
			"app-name":    appName,
			"clusters":    clusters,
			"err":         err})

		return false, nil, pkgerrors.Wrapf(err, "filterAllocResource .. Hpa Intent PopulateClustersResources error intent-name[%v] app-name[%v] err[%v]", intentName, appName, err)
	}
	log.Info("filterAllocResource .. Hpa Intent PopulateClustersResources ", log.Fields{
		"intent-name":           intentName,
		"app-name":              appName,
		"clusters":              clusters,
		"clusterResourceObjMap": clusterResourceObjMap})

	qualifiedClusters := make([]string, 0)
	qualifiedNodes := make([]string, 0)
	// Interate through all clusters and filter qualified nodes
	for _, cl := range clusters {
		log.Info("filterAllocResource .. resource is allocatable => ", log.Fields{"hpa-resource-name": hpaResource.MetaData.Name})

		// if resource request is not specified but limit specified them request = limit
		if hpaResource.Spec.Resource.Requests == 0 {
			hpaResource.Spec.Resource.Requests = hpaResource.Spec.Resource.Limits
		}

		// instantiate allocatable resource
		rsAllocatable := (*clusterResourceObjMap)[cl].AllocatableRs

		// Check if clusterResourceInfoMap contains data
		if len(*clusterResourceInfoMap) > 0 {
			// Check if the cluster name exists in the map
			if _, ok := (*clusterResourceInfoMap)[cl]; ok {
				clusterResourceInfo := (*clusterResourceInfoMap)[cl]

				matched = rsAllocatable.Qualified(context.TODO(), clusterResourceInfo.ClusterName, *hpaResource)
				qualifiedNodes = rsAllocatable.GetQualifiedNodes(hpaResource.Spec.Resource.Name)

				if len(qualifiedNodes) > 0 {
					matched = true
				} else {
					if rollbackNeeded {
						// rollback node accounting for the not-matched cluster
						log.Info("filterAllocResource .. No Qualified Cluster Nodes .. Rollback node accounting for the not-matched cluster.",
							log.Fields{"resource-name": hpaResource.Spec.Resource.Name, "nodeMap": rsAllocatable.GetNodeResMap(hpaResource.Spec.Resource.Name), "cluster_info": clusterResourceInfo, "cluster": cl, "hpa-resource": hpaResource})
						rsAllocatable.RollbackAccounting(hpaResource.Spec.Resource.Name)
					}
				}
			}
		} // if len(clusterResourceInfoMap) > 0 {

		// if found a qualified cluster, add to the qualified cluster list
		if matched {
			qualifiedClusters = append(qualifiedClusters, cl)
			clusterToNodesMap[cl] = make([]string, 0)
			clusterToNodesMap[cl] = qualifiedNodes

			log.Info("HPA filterAllocResource .. Found a qualified cluster", log.Fields{
				"intent-name":        intentName,
				"app-name":           appName,
				"hpa-resource":       hpaResource,
				"qualified-cluster":  cl,
				"qualified-clusters": qualifiedClusters,
				"qualified-nodes":    qualifiedNodes,
				"clusters":           clusters})
		}
	} // for clusters

	log.Info("HPA filterAllocResource .. end", log.Fields{
		"intent-name":          intentName,
		"app-name":             appName,
		"clusters":             clusters,
		"qualified-clusters":   qualifiedClusters,
		"qualified-nodes":      qualifiedNodes,
		"cluster-to-nodes-map": clusterToNodesMap,
		"hpa-resource":         hpaResource,
	})

	if len(qualifiedClusters) == 0 {
		log.Error("filterAllocResource .. None of Cluster match hpa resource",
			log.Fields{"clusters": clusters, "hpa-resource": hpaResource})
		return false, nil, pkgerrors.Errorf("filterAllocResource .. None of Cluster match hpa resource[%s]", hpaResource.MetaData.Name)
	}
	return true, clusterToNodesMap, nil
}

//filterNonAllocResource ... filter w.r.t hpa resource
func filterNonAllocResource(intentName string, appName string, hpaResource *hpaModel.HpaResourceRequirement, clusterResourceObjMap *intentRs.ClusterResourceObjMap, clusterResourceInfoMap *intentRs.ClusterResourceInfoMap, clusters []string) (bool, map[string]([]string), error) {
	log.Info("filterNonAllocResource .. start", log.Fields{
		"intent-name":  intentName,
		"app-name":     appName,
		"clusters":     clusters,
		"hpa-resource": hpaResource,
	})

	var clusterToNodesMap map[string]([]string)
	clusterToNodesMap = make(map[string]([]string))
	matched := false

	// Populate cluster/node resource Info
	err := PopulateClustersResources(context.TODO(), hpaResource, clusterResourceInfoMap, clusterResourceObjMap)
	if err != nil {
		log.Error("filterNonAllocResource .. Hpa Intent PopulateClustersResources error=> ", log.Fields{
			"intent-name": intentName,
			"app-name":    appName,
			"clusters":    clusters,
			"err":         err})

		return false, nil, pkgerrors.Wrapf(err, "filterNonAllocResource .. Hpa Intent PopulateClustersResources error intent-name[%v] app-name[%v] err[%v]", intentName, appName, err)
	}
	log.Info("filterNonAllocResource .. Hpa Intent PopulateClustersResources ", log.Fields{
		"intent-name":           intentName,
		"app-name":              appName,
		"clusters":              clusters,
		"clusterResourceObjMap": clusterResourceObjMap})

	qualifiedClusters := make([]string, 0)
	qualifiedNodes := make([]string, 0)
	// Interate through all clusters and filter qualified nodes
	for _, cl := range clusters {

		log.Info("filterNonAllocResource .. resource is non-allocatable => ", log.Fields{"cluster-name": cl, "hpa-resource-name": hpaResource.MetaData.Name})

		// instantiate non-allocatable resource
		rsNonAllocatable := (*clusterResourceObjMap)[cl].NonAllocatableRs
		if rsNonAllocatable != nil {
			// Check if clusterResourceInfoMap contains data
			if len(*clusterResourceInfoMap) > 0 {
				// Check if the cluster name exists in the map
				if _, ok := (*clusterResourceInfoMap)[cl]; ok {
					clusterResourceInfo := (*clusterResourceInfoMap)[cl]

					matched = rsNonAllocatable.Qualified(context.TODO(), clusterResourceInfo.ClusterName, *hpaResource)
					qualifiedNodes = rsNonAllocatable.GetQualifiedNodes()
					if len(qualifiedNodes) > 0 {
						matched = true
					}
					log.Info("HpaIntent Filtering Found non-allocatable res=> ", log.Fields{"matched": matched, "cluster_name": clusterResourceInfo.ClusterName, "qualified_nodes": qualifiedNodes})
				}
			}
		} else {
			log.Info("filterNonAllocResource .. clusterResourceObjMap dump", log.Fields{"len_clusterResourceObjMap": len(*clusterResourceObjMap)})
			index := 0
			for k, cluster := range *clusterResourceObjMap {
				log.Info("filterNonAllocResource .. clusterResourceObjMap cluster dump=>", log.Fields{
					"index":                    index,
					"key":                      k,
					"ClusterName":              cluster.ClusterName,
					"AllocatableRs":            cluster.AllocatableRs,
					"NonAllocatableRs":         cluster.NonAllocatableRs,
					"ClusterResourceCount":     cluster.AllocatableRs.GetClusterResourceCount(hpaResource.Spec.Resource.Name),
					"ClusterResourceCountOrig": cluster.AllocatableRs.GetClusterResourceCountOrig(hpaResource.Spec.Resource.Name),
					"NodeResMap":               cluster.AllocatableRs.GetNodeResMap(hpaResource.Spec.Resource.Name)})

				index++
			}
			log.Error("filterNonAllocResource .. Unable to find rsNonAllocatable object", log.Fields{"cluster": cl, "clusterResourceObjMap": clusterResourceObjMap, "hpa-resource-name": hpaResource.MetaData.Name})
			return false, nil, pkgerrors.Wrapf(err, "filterNonAllocResource .. Unable to find rsNonAllocatable object[%s]", cl)
		}

		// if found a qualified cluster, add to the qualified cluster list
		if matched {
			qualifiedClusters = append(qualifiedClusters, cl)
			clusterToNodesMap[cl] = make([]string, 0)
			clusterToNodesMap[cl] = qualifiedNodes

			log.Info("HPA filterNonAllocResource .. Found a qualified cluster", log.Fields{
				"intent-name":        intentName,
				"app-name":           appName,
				"hpa-resource":       hpaResource,
				"qualified-cluster":  cl,
				"qualified-clusters": qualifiedClusters,
				"qualified-nodes":    qualifiedNodes,
				"clusters":           clusters})
		}
	} // for clusters

	log.Info("HPA filterNonAllocResource .. end", log.Fields{
		"intent-name":          intentName,
		"app-name":             appName,
		"clusters":             clusters,
		"qualified-clusters":   qualifiedClusters,
		"qualified-nodes":      qualifiedNodes,
		"cluster-to-nodes-map": clusterToNodesMap,
		"hpa-resource":         hpaResource,
	})

	if len(qualifiedClusters) == 0 {
		log.Error("filterNonAllocResource .. None of Cluster match hpa resource",
			log.Fields{"clusters": clusters, "hpa-resource": hpaResource})
		return false, nil, pkgerrors.Errorf("filterNonAllocResource .. None of Cluster match hpa resource[%s]", hpaResource.MetaData.Name)
	}
	return true, clusterToNodesMap, nil
}

// PopulateClustersResources ... Populate model with cluster resource info
func PopulateClustersResources(ctx context.Context, hpaResource *hpaModel.HpaResourceRequirement, clusters *intentRs.ClusterResourceInfoMap, clusterResourceObjMap *intentRs.ClusterResourceObjMap) error {
	log.Info("PopulateClustersResources .. start", log.Fields{"clustersCount": len(*clusters), "hpaResource": hpaResource})

	if len(*clusters) > 0 {
		// Populate cluster resources
		for _, cluster := range *clusters {
			log.Info("PopulateClustersResources .. cluster populate resources .. start", log.Fields{"ClusterName": cluster.ClusterName})

			// instantiate nfd resource
			rsNonAllocatable := (*clusterResourceObjMap)[cluster.ClusterName].NonAllocatableRs
			// instantiate generic resource
			rsAllocatable := (*clusterResourceObjMap)[cluster.ClusterName].AllocatableRs

			if !(*hpaResource.Spec.Allocatable) && (rsNonAllocatable != nil) {
				log.Info("PopulateClustersResources .. Non-Allocatable resources", log.Fields{"ClusterName": cluster.ClusterName, "is_allocatable": hpaResource.Spec.Allocatable, "hpaResource-spec": hpaResource.Spec})
				// Pull Cluster labels from db & Tokenize received cluster-detail into provider-name & cluster-name
				if strings.Contains(cluster.ClusterName, "+") {
					tokens := strings.Split(cluster.ClusterName, "+")
					nodeLabels, err := GetKubeClusterLabels(tokens[0], tokens[1])
					if err != nil {
						log.Error("PopulateClustersResources .. Unable to find the cluster labels", log.Fields{"cluster": cluster.ClusterName, "hpa-resource-name": hpaResource.MetaData.Name})
					} else {
						err = rsNonAllocatable.SetResourceInfo(ctx, cluster.ClusterName, nodeLabels)
						if err != nil {
							log.Error("PopulateClustersResources .. NonAllocatable SetResourceInfo Resource failed for a cluster.", log.Fields{"cluster": cluster, "err": err})
							//return pkgerrors.Wrapf(err, "PopulateClustersResources ..  NonAllocatable SetResourceInfo Resource failed err[%v]", err)
						} else {
							log.Info("PopulateClustersResources .. resource is non-allocatable. cluster-labels=> ", log.Fields{"hpa-resource-spec": hpaResource.Spec, "node-labels": nodeLabels})
						}
					}
				} else {
					log.Error("PopulateClustersResources .. Not a valid cluster name", log.Fields{"cluster": cluster.ClusterName, "hpa-resource-spec": hpaResource.Spec})
				}
			} else if *(hpaResource.Spec.Allocatable) && (rsAllocatable != nil) {
				log.Info("PopulateClustersResources .. Received Allocatable resource", log.Fields{"hpa-resource": hpaResource, "rs-allocatable-info": rsAllocatable})
				if !rsAllocatable.IsResourceAlreadyPopulated(hpaResource.Spec.Resource.Name) {
					log.Info("PopulateClustersResources .. Allocatable resource .. start", log.Fields{"ClusterName": cluster.ClusterName, "is_allocatable": hpaResource.Spec.Allocatable, "hpaResource-spec": hpaResource.Spec, "rs-allocatable-populated": rsAllocatable.ConfigLoaded})
					clusterTotal, nodeTotal, err := rsAllocatable.PopulateResourceInfo(ctx, cluster.ClusterName, *hpaResource)
					if err != nil {
						log.Error("PopulateClustersResources .. Allocatable PopulateResourceInfo Resource failed for a cluster.", log.Fields{"cluster": cluster, "err": err})
						//return pkgerrors.Wrapf(err, "PopulateClustersResources ..  Allocatable PopulateResourceInfo Resource failed. err[%v]", err)
					} else {
						cluster.ClusterAvailResCount = rsAllocatable.GetClusterResourceCount(hpaResource.Spec.Resource.Name)
						cluster.NodeMaxAvailResCount = rsAllocatable.GetNodeResourceAvailMaxCount(hpaResource.Spec.Resource.Name)
						cluster.ClusterAvailResCountOrig = rsAllocatable.GetClusterResourceCountOrig(hpaResource.Spec.Resource.Name)
						cluster.NodeMaxAvailResCountOrig = rsAllocatable.GetNodeResourceAvailMaxCountOrig(hpaResource.Spec.Resource.Name)
					}

					log.Info("PopulateClustersResources .. Allocatable resource .. end",
						log.Fields{"ClusterName": cluster.ClusterName, "hpa-resource-spec": hpaResource.Spec, "clusterTotal": clusterTotal, "nodeTotal": nodeTotal})
				} else {
					log.Info("PopulateClustersResources .. Allocatable resource already populated",
						log.Fields{"ClusterName": cluster.ClusterName, "hpa-resource-spec": hpaResource.Spec})
				}
			}
			log.Info("PopulateClustersResources .. cluster populate resources .. end",
				log.Fields{"cluster-info": cluster})
		} // for clusters
	} //if len(clusters) > 0 {

	log.Info("PopulateClustersResources .. clusterResourceObjMap dump", log.Fields{"len_clusterResourceObjMap": len(*clusterResourceObjMap)})
	index := 0
	for k, cluster := range *clusterResourceObjMap {
		log.Info("PopulateClustersResources .. clusterResourceObjMap cluster dump=>", log.Fields{
			"index":        index,
			"key":          k,
			"cluster-info": cluster})
		index++
	}
	log.Info("PopulateClustersResources .. end", log.Fields{"clustersCount": len(*clusters), "hpaResource": hpaResource, "clusterResourceObjMapCount": len((*clusterResourceObjMap))})
	return nil
}

// initializeResourceInfo ... initialize model
func initializeResourceInfo(ctx context.Context, hpaIntent *hpaModel.DeploymentHpaIntent, clusters *intentRs.ClusterResourceInfoMap, clusterResourceObjMap *intentRs.ClusterResourceObjMap) error {
	log.Info("initializeResourceInfo .. start",
		log.Fields{"hpa-intent": hpaIntent, "clustersCount": len(*clusters), "clusters": clusters})

	if len(*clusters) > 0 {
		// Populate cluster resources
		for _, cluster := range *clusters {
			log.Info("initializeResourceInfo .. cluster populate resources .. start", log.Fields{"ClusterName": cluster.ClusterName})

			// instantiate nfd resource
			var rsNonAllocatable intentRs.NFDResource
			if (*clusterResourceObjMap)[cluster.ClusterName].NonAllocatableRs == nil {
				rsNonAllocatable = intentRs.NFDResource{}
				log.Info("initializeResourceInfo .. non-allocatable cluster resources never poulated, create new", log.Fields{"ClusterName": cluster.ClusterName, "rsNonAllocatable": rsNonAllocatable})
			} else {
				rsNonAllocatable = *(*clusterResourceObjMap)[cluster.ClusterName].NonAllocatableRs
				log.Info("initializeResourceInfo .. non-allocatable cluster resources already poulated, use them", log.Fields{"ClusterName": cluster.ClusterName, "rsNonAllocatable": rsNonAllocatable})
			}

			// instantiate generic resource
			var rsAllocatable intentRs.GenericResource
			if (*clusterResourceObjMap)[cluster.ClusterName].AllocatableRs == nil {
				log.Info("initializeResourceInfo .. allocatable cluster resources never poulated, create new", log.Fields{"ClusterName": cluster.ClusterName, "rsAllocatable": rsAllocatable})
				rsAllocatable := intentRs.GenericResource{}
				rsAllocatable.Initialize()
			} else {
				rsAllocatable = *(*clusterResourceObjMap)[cluster.ClusterName].AllocatableRs
				log.Info("initializeResourceInfo .. allocatable cluster resources already poulated, use them", log.Fields{"ClusterName": cluster.ClusterName, "rsAllocatable": rsAllocatable})
			}

			// Fill the cluster resource map
			(*clusterResourceObjMap)[cluster.ClusterName] = intentRs.ClusterResourceObj{ClusterName: cluster.ClusterName, AllocatableRs: &rsAllocatable, NonAllocatableRs: &rsNonAllocatable}

			log.Info("initializeResourceInfo .. cluster populate resources .. end", log.Fields{"ClusterName": cluster.ClusterName})
		} // for clusters
	} //if len(clusters) > 0 {

	log.Info("initializeResourceInfo .. clusterResourceObjMap dump", log.Fields{"len_clusterResourceObjMap": len(*clusterResourceObjMap)})
	index := 0
	for k, cluster := range *clusterResourceObjMap {
		log.Info("initializeResourceInfo .. clusterResourceObjMap cluster dump=>", log.Fields{
			"index":            index,
			"key":              k,
			"ClusterName":      cluster.ClusterName,
			"AllocatableRs":    cluster.AllocatableRs,
			"NonAllocatableRs": cluster.NonAllocatableRs})
		index++
	}
	log.Info("initializeResourceInfo .. end",
		log.Fields{"hpa-intent": hpaIntent, "clustersCount": len(*clusters), "clusters": clusters})
	return nil
}

// Publish ... Publish event
func Publish(ctx context.Context, req *clmcontrollerpb.ClmControllerEventRequest) error {

	log.Info("Publish .. start", log.Fields{"req": req, "event": req.Event.String()})

	var err error = nil
	switch req.Event {
	case clmcontrollerpb.ClmControllerEventType_CLUSTER_CREATED, clmcontrollerpb.ClmControllerEventType_CLUSTER_UPDATED:
		err = SaveClusterLabelsDB(req.ProviderName, req.ClusterName)
	case clmcontrollerpb.ClmControllerEventType_CLUSTER_DELETED:
		err = DeleteKubeClusterLabelsDB(req.ProviderName, req.ClusterName)
	default:
		log.Warn("Publish .. Received Unknown event", log.Fields{"req": req, "event": req.Event.String()})
	}
	if err != nil {
		return pkgerrors.Wrapf(err, "Error while saving Cluster labels[%v]", *req)
	}

	log.Info("Publish .. end", log.Fields{"req": req, "event": req.Event.String()})

	return nil
}

func findNodeForAccounting(nodeMap map[string]int64, qualifiedNodes []string, hpaResource hpaModel.HpaResourceRequirement) string {
	log.Info("findNodeForAccounting .. start", log.Fields{"nodeMap": nodeMap, "qualifiedNodes": qualifiedNodes, "hpaResource": hpaResource})

	var nodeChosen string = ""
	// create reverse map for the request
	revMap := make(map[int64][]string)
	for k, v := range nodeMap {
		if v >= hpaResource.Spec.Resource.Requests {
			if _, found := revMap[v]; found {
				revMap[v] = append(revMap[v], k)
			} else {
				revMap[v] = append(revMap[v], k)
			}
		}
	}

	// Sort the Keys of the reverse map (descending order)
	values := make([]int64, 0, len(revMap))
	for k := range revMap {
		values = append(values, k)
	}
	if len(values) > 0 {
		sort.Slice(values, func(i, j int) bool { return values[i] > values[j] })
	}

	for _, value := range values {
		if len(revMap[value]) > 0 {
			if _, ok := revMap[value]; ok {
				for _, nodeName := range revMap[value] {
					if hpaUtils.IsInSlice(nodeName, qualifiedNodes) {
						nodeChosen = nodeName
						break
					}
				}
			}
		}
	} // for
	log.Info("findNodeForAccounting .. end", log.Fields{"nodeChosen": nodeChosen, "nodeMap": nodeMap, "revMap": revMap, "qualifiedNodes": qualifiedNodes, "hpaResource": hpaResource})
	return nodeChosen
}
