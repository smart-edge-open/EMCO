// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2020 Intel Corporation

package servicediscovery

import (
	"context"
	"time"

	pkgerrors "github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"

	"github.com/open-ness/EMCO/src/dtc/internal/utils"
	"github.com/open-ness/EMCO/src/dtc/pkg/grpc/rsyncclient"
	"github.com/open-ness/EMCO/src/dtc/pkg/module"
	"github.com/open-ness/EMCO/src/orchestrator/pkg/appcontext"
	log "github.com/open-ness/EMCO/src/orchestrator/pkg/infra/logutils"
	"github.com/open-ness/EMCO/src/orchestrator/pkg/state"
	readynotifypb "github.com/open-ness/EMCO/src/rsync/pkg/grpc/readynotify"
)

const (
	parentACInstantiatedTimeout  = time.Duration(120) * time.Second
	pollingInterval              = 5
	serverAppDeploymentTimeout   = time.Duration(300) * time.Second
	loadbalancerIngressIPTimeout = time.Duration(120) * time.Second
	compositeApp                 = "service-discovery"
	transportError               = "rpc error: code = Unavailable desc = transport is closing"
)

// CreateAppContext Action applies the supplied intent against the given AppContext ID
func CreateAppContext(intentName, appContextID string) error {
	var ac appcontext.AppContext
	_, err := ac.LoadAppContext(appContextID)
	if err != nil {
		log.Error("Error loading AppContext", log.Fields{
			"error": err,
		})
		return pkgerrors.Wrapf(err, "Error getting AppContext with Id: %v", appContextID)
	}

	caMeta, err := ac.GetCompositeAppMeta()
	if err != nil {
		log.Error("Error getting metadata from AppContext", log.Fields{
			"error": err,
		})
		return pkgerrors.Wrapf(err, "Error getting metadata from AppContext with Id: %v", appContextID)
	}

	project := caMeta.Project
	compositeapp := caMeta.CompositeApp
	compositeappversion := caMeta.Version
	deployIntentGroup := caMeta.DeploymentIntentGroup

	// Get all server inbound intents
	iss, err := module.NewServerInboundIntentClient().GetServerInboundIntents(project, compositeapp, compositeappversion, deployIntentGroup, intentName)
	if err != nil {
		log.Error("Error getting server inbound intents", log.Fields{
			"error": err,
		})
		return pkgerrors.Wrapf(err, "Error getting server inbound intents %v for %v/%v%v/%v not found", intentName, project, compositeapp, deployIntentGroup, compositeappversion)
	}

	for _, is := range iss {

		ics, err := module.NewClientsInboundIntentClient().GetClientsInboundIntents(project,
			compositeapp,
			compositeappversion,
			deployIntentGroup,
			intentName,
			is.Metadata.Name)
		if err != nil {
			log.Error("Error getting clients inbound intents", log.Fields{
				"error": err,
			})
			return pkgerrors.Wrapf(err,
				"Error getting clients inbound intents %v under server inbound intent %v for %v/%v%v/%v not found",
				is.Metadata.Name, intentName, project, compositeapp, compositeappversion, deployIntentGroup)
		}

		clientSets := make(map[string]string, 1)
		server := is.Spec.AppName
		for _, ic := range ics {

			if ic.Spec.AppName != "" && is.Spec.ServiceName != "" && is.Spec.AppName != "" {

				clientSets[ic.Spec.AppName] = is.Spec.ServiceName

			} else {
				log.Info("Either client's app name or server's app name or service name is empty", log.Fields{
					"ClientAppName": ic.Spec.AppName, "ServiceName": is.Spec.ServiceName, "ServerAppName": is.Spec.AppName,
				})
			}
		}

		// Register for an appcontext alert and receive the stream handle
		stream, client, err := rsyncclient.InvokeReadyNotify(appContextID)
		if err != nil {
			log.Error("Error in callRsyncReadyNotify", log.Fields{
				"error": err, "appContextID": appContextID,
			})
			return pkgerrors.Wrap(err, "Error in callRsyncReadyNotify")
		}

		go processServiceDiscovery(appContextID, clientSets, server, stream, client)
	}

	return nil
}

// processServiceDiscovery will fetch all the service related specs from the deployed apps and
// deploy the service entry on the clusters
func processServiceDiscovery(appContextID string, clientSets map[string]string, server string, stream readynotifypb.ReadyNotify_AlertClient, cl readynotifypb.ReadyNotifyClient) error {

	var ac appcontext.AppContext
	_, err := ac.LoadAppContext(appContextID)
	if err != nil {
		log.Error("Error getting AppContext with Id", log.Fields{
			"error": err, "appContextID": appContextID,
		})
		return pkgerrors.Wrapf(err, "Error getting AppContext with Id: %v", appContextID)
	}

	// Obtain the service related specs associated with the server app
	err = processAlertForServiceDiscovery(stream, appContextID, server)
	if err != nil {
		log.Error("Unable to process the alert for service discovery", log.Fields{"err": err.Error()})
	}

	// call unsubscribe
	_, err = cl.Unsubscribe(context.Background(), &readynotifypb.Topic{ClientName: "dtc", AppContext: appContextID})
	if err != nil {
		log.Error("[ReadyNotify gRPC] Failed to unsubscribe to alerts", log.Fields{"err": err, "appContextId": appContextID})
		return err
	}

	// close the stream
	stream.CloseSend()

	// create a new child app contexts and link it to the parent's meta data and
	// deploy these new child app contexts which contains the virtual service entries

	// Get the appcontext status value
	acStatus, err := state.GetAppContextStatus(appContextID)
	if err != nil {
		log.Error("Unable to get the parent's app context status", log.Fields{"err": err.Error()})
		return pkgerrors.Wrap(err, "Unable to get the status of the app context")
	}
	if acStatus.Status == appcontext.AppContextStatusEnum.Instantiated {
		for client, service := range clientSets {
			err := DeployServiceEntry(ac, appContextID, server, client, service)
			if err != nil {
				log.Error("Unable to deploy the virtual service entry", log.Fields{"err": err.Error(), "clientApp": client})
				return pkgerrors.Wrap(err, "Unable to deploy the virtual service entry for the client: "+client)
			}

		}
	}

	return nil
}

func processAlertForServiceDiscovery(stream readynotifypb.ReadyNotify_AlertClient, appContextID string, serverAppName string) error {

	for {

		loadBalancerIPSet := false
		// Now check whether the parent app context has been "Instantiated".

		acStatus, err := state.GetAppContextStatus(appContextID)
		if err != nil {
			log.Warn("[ReadyNotify gRPC] Unable to get the status of the app context", log.Fields{"err": err, "appContextID": appContextID})
			continue
		}

		if acStatus.Status == appcontext.AppContextStatusEnum.Instantiated {
			log.Info("Parent's app context is in 'Instantiated' state. Checking for the app to be deployed successfully", log.Fields{"appContextID": appContextID})
			// Now check the status of the app deployed
			condition, err := utils.CheckDeploymentStatus(appContextID, serverAppName)
			if err != nil {
				log.Error("Unable to check the deployment status of the server app", log.Fields{"err": err.Error(), "serverApp": serverAppName})
				return pkgerrors.Wrap(err, "Unable to check the deployment status of the server app")
			}
			if condition {
				// Server App has been successfully deployed
				log.Info("Server App has been successfully deployed", log.Fields{"serverApp": serverAppName})

				// Check for the loadbalancer external IP
				var ac appcontext.AppContext
				_, err := ac.LoadAppContext(appContextID)
				if err != nil {
					log.Error("Error loading AppContext", log.Fields{
						"error": err,
					})
					return pkgerrors.Wrap(err, "Error loading AppContext")
				}

				// Get the clusters in the appcontext for this app
				clusters, err := ac.GetClusterNames(serverAppName)
				if err != nil {
					log.Error("Unable to get the cluster names",
						log.Fields{"AppName": serverAppName, "Error": err})
					return pkgerrors.Wrap(err, "Unable to get the cluster names")
				}
				for _, cluster := range clusters {
					rbValue, err := utils.GetClusterResources(appContextID, serverAppName, cluster)
					if err != nil {
						log.Error("Unable to get the cluster resources",
							log.Fields{"Cluster": cluster, "AppName": serverAppName, "Error": err})
						return pkgerrors.Wrap(err, "Unable to get the cluster resources")
					}
					for _, s := range rbValue.ServiceStatuses {

						if s.Spec.Type == corev1.ServiceTypeLoadBalancer {
							if len(s.Status.LoadBalancer.Ingress) != 0 {
								for _, ingress := range s.Status.LoadBalancer.Ingress {
									if ingress.IP != "" {
										log.Info("Server App's service has the loadbalancer ingress IP", log.Fields{"serverApp": serverAppName, "loadbalancerIP": ingress.IP})
										loadBalancerIPSet = true
										return nil
									}
								}
							} else {
								log.Info("Server App's service doesn't has the loadbalancer ingress IP assigned yet", log.Fields{"serverApp": serverAppName})
							}
						} else {
							log.Info("Server App's service type is not loadbalancer and hence exiting the ready notify stream channel", log.Fields{"serverApp": serverAppName})
							return nil
						}

					}
				}
			} else {
				log.Info("Server App has not been deployed yet", log.Fields{"serverApp": serverAppName})
			}
		} else if acStatus.Status == appcontext.AppContextStatusEnum.Instantiating {
			log.Info("Parent's app context is still in 'Instantiating' state", log.Fields{"appContextID": appContextID})
		} else { // If the parent's appContext is "Terminating/Terminated/InstantiateFailed/TerminateFailed"
			log.Error("Parent's app context is not in 'Instantiated' state", log.Fields{"err": err.Error()})
			return pkgerrors.Wrap(err, "Parent's app context is not in 'Instantiated' state")
		}

		if !loadBalancerIPSet {
			// Here the code gets blocked until the load balancer external IP is obtained
			if stream != nil {
				resp, err := stream.Recv()
				if err != nil {
					log.Error("[ReadyNotify gRPC] Failed to receive notification", log.Fields{"err": err.Error()})
					if err.Error() == transportError {
						time.Sleep(5 * time.Second)
						// If rsync crashes/restarts while the stream channel is active then poll until the load balancer IP is obtained
						continue
					}
					return pkgerrors.Wrap(err, "Failed to receive notification")
				}

				appContextID = resp.AppContext
				log.Info("[ReadyNotify gRPC] Received alert from rsync", log.Fields{"appContextId": appContextID, "err": err})
			} else {
				// if stream is nil then poll until the load balancer IP is obtained
				time.Sleep(5 * time.Second)
				continue
			}

		}

	}

}
