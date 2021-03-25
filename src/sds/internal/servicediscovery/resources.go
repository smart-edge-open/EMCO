// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2020 Intel Corporation

package servicediscovery

import (
	"fmt"

	pkgerrors "github.com/pkg/errors"
	"gopkg.in/yaml.v2"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	"github.com/open-ness/EMCO/src/sds/internal/utils"
	rb "github.com/open-ness/EMCO/src/monitor/pkg/apis/k8splugin/v1alpha1"
	"github.com/open-ness/EMCO/src/orchestrator/pkg/appcontext"
	log "github.com/open-ness/EMCO/src/orchestrator/pkg/infra/logutils"
	"github.com/open-ness/EMCO/src/orchestrator/pkg/state"
	connector "github.com/open-ness/EMCO/src/rsync/pkg/connector"
)

// serviceEntry provides the contents of the virtual proxy service
type serviceEntry struct {
	// Name of the service deployed
	name string
	// Namespace defined for the deployed service
	namespace string
	// IP defines the service address. NodeIP for type clusterIP/NodePort and loadbalancer ingress IP for loadbalancer type
	ip string
	// ServicePorts defines the list of service ports exposed
	servicePorts []corev1.ServicePort
}

// ServiceResource defines the service k8s object
type ServiceResource struct {
	APIVersion    string            `yaml:"apiVersion"`
	Kind          string            `yaml:"kind"`
	MetaData      metav1.ObjectMeta `yaml:"metadata"`
	Specification Specs             `yaml:"spec,omitempty"`
}

// Specs defines the service spec
type Specs struct {
	ClusterIP       string               `yaml:"clusterIP,omitempty"`
	Ports           []corev1.ServicePort `yaml:"ports,omitempty"`
	SessionAffinity string               `yaml:"sessionAffinity,omitempty"`
	Types           corev1.ServiceType   `yaml:"type,omitempty"`
}

// EndpointResource defines the endpoint k8s object
type EndpointResource struct {
	APIVersion string                  `yaml:"apiVersion"`
	Kind       string                  `yaml:"kind"`
	MetaData   metav1.ObjectMeta       `yaml:"metadata"`
	Subsets    []corev1.EndpointSubset `yaml:"subsets,omitempty"`
}

// getClusterServiceSpecs takes in a ResourceBundleStatus CR and returns the service relates specs
func getClusterServiceSpecs(ac appcontext.AppContext, appContextID string, rbData *rb.ResourceBundleStatus, serviceName string,
	serverApp string, cluster string) (serviceEntry, error) {

	virtualService := serviceEntry{}

	for _, s := range rbData.ServiceStatuses {
		if !utils.CompareResource(s.Name, serviceName) {
			continue
		}

		switch s.Spec.Type {
		// If the service Type is NodePort, then obtain the node IP where the server app is running
		case corev1.ServiceTypeNodePort:
			var connector *connector.Connection
			err := connector.Init(appContextID)
			if err != nil {
				return virtualService, pkgerrors.New("unable to Initialize connection")
			}

			kubeClient, err := connector.GetClient(cluster, "0", "default")
			if err != nil {
				log.Error("unable to connect to the cluster",
					log.Fields{"Cluster": cluster, "Error": err})
				return virtualService, pkgerrors.New("unable to connect to the cluster")
			}
			nodeIP, err := kubeClient.GetMasterNodeIP()
			if err != nil {
				log.Error("unable to get the master node IP",
					log.Fields{"Cluster": cluster, "Error": err})
				return virtualService, pkgerrors.New("unable to retrieve the master node IP")
			}
			virtualService.ip = nodeIP
			virtualService.name = serviceName
			virtualService.servicePorts = s.Spec.Ports
			break

		// If the service Type is LoadBalancer, then poll until the ingress loadbalancer IP is obtained
		case corev1.ServiceTypeLoadBalancer:

			// Get the appcontext status value
			acStatus, err := state.GetAppContextStatus(appContextID)
			if err != nil {
				log.Error("Unable to get the status of the app context",
					log.Fields{"appContextID": appContextID, "Error": err})
				return virtualService, pkgerrors.Wrap(err, "Unable to get the status of the app context")
			}
			if acStatus.Status == appcontext.AppContextStatusEnum.Instantiated {
				rbValue, err := utils.GetClusterResources(appContextID, serverApp, cluster)
				if err != nil {
					log.Error("Unable to get the cluster resources",
						log.Fields{"Cluster": cluster, "AppName": serverApp, "Error": err})
					return virtualService, pkgerrors.Wrap(err, "Unable to get the status of the app context")
				}
				for _, s := range rbValue.ServiceStatuses {
					if !utils.CompareResource(s.Name, serviceName) {
						continue
					}

					if len(s.Status.LoadBalancer.Ingress) != 0 {
						for _, ingress := range s.Status.LoadBalancer.Ingress {
							if ingress.IP != "" {
								virtualService.ip = ingress.IP
								virtualService.name = serviceName
								virtualService.servicePorts = s.Spec.Ports
								// Obtain the loadbalancer IP
								log.Info("Loadbalancer ingress IP", log.Fields{"IP": virtualService.ip, "service": serviceName})
								return virtualService, nil
							}
						}
					}
				}
			} else {
				log.Info("Parent's app context is not in 'Instantiated' state",
					log.Fields{"appContextID": appContextID})
				return virtualService, pkgerrors.Wrap(err, "Parent's app context is not in 'Instantiated' state")
			}

			if err != nil {
				log.Error("Polling for loadbalance ingress IP failed",
					log.Fields{"appContextID": appContextID, "Error": err})
				return virtualService, pkgerrors.Wrap(err, "Polling for loadbalance ingress IP failed")
			}
			break

		default:
			log.Info("service deployed is neighter Nodeport/LoadBalancer type",
				log.Fields{"appContextID": appContextID, "serviceName": serviceName})
			return virtualService, pkgerrors.New("service deployed is neighter Nodeport/LoadBalancer type")
		}
	}

	// Get the parent composite app meta
	m, err := ac.GetCompositeAppMeta()
	if err != nil {
		log.Error("Error getting CompositeAppMeta",
			log.Fields{"Cluster": cluster, "AppName": serverApp, "Error": err})
		return virtualService, pkgerrors.New("Error getting CompositeAppMeta")
	}

	// Add the namespace
	virtualService.namespace = m.Namespace

	log.Info("The virtual service entries inside are", log.Fields{"virtualService": virtualService})
	return virtualService, nil
}

// createService creates a YAML template for rsync to deploy the service object
func createService(se serviceEntry) (string, error) {

	var appServicePorts []corev1.ServicePort

	// Assign the port, target port and protocol
	for _, servicePort := range se.servicePorts {
		var externalport intstr.IntOrString
		var appServicePort corev1.ServicePort

		externalport = intstr.IntOrString{IntVal: servicePort.Port}
		appServicePort = corev1.ServicePort{
			Name:       fmt.Sprint(servicePort.Port),
			Protocol:   servicePort.Protocol,
			Port:       servicePort.Port,
			TargetPort: externalport,
		}
		appServicePorts = append(appServicePorts, appServicePort)
	}

	service := ServiceResource{
		APIVersion: "v1",
		Kind:       "Service",
		MetaData: metav1.ObjectMeta{
			Name:      se.name,
			Namespace: se.namespace,
		},
		Specification: Specs{
			Ports:           appServicePorts,
			ClusterIP:       corev1.ClusterIPNone,
			SessionAffinity: "None",
			Types:           corev1.ServiceTypeClusterIP,
		},
	}

	serviceData, err := yaml.Marshal(&service)
	if err != nil {
		return "", err
	}

	return string(serviceData), nil
}

// createEndpoint creates a YAML template for rsync to deploy the endpoint object
func createEndpoint(se serviceEntry) (string, error) {
	var endpointPorts []corev1.EndpointPort

	for _, servicePort := range se.servicePorts {
		var endpointPort corev1.EndpointPort

		endpointPort = corev1.EndpointPort{
			Name: fmt.Sprint(servicePort.Port),
			Port: servicePort.Port,
		}
		endpointPorts = append(endpointPorts, endpointPort)
	}

	endpoint := EndpointResource{
		APIVersion: "v1",
		Kind:       "Endpoints",
		MetaData: metav1.ObjectMeta{
			Name:      se.name,
			Namespace: se.namespace,
		},
		Subsets: []corev1.EndpointSubset{
			{
				Addresses: []corev1.EndpointAddress{
					{
						IP: se.ip,
					},
				},
				Ports: endpointPorts,
			},
		},
	}

	endpointData, err := yaml.Marshal(&endpoint)
	if err != nil {
		return "", err
	}

	return string(endpointData), nil
}
