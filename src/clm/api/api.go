// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2020 Intel Corporation
package api

import (
	"fmt"
	"reflect"

	"github.com/gorilla/mux"
	"github.com/open-ness/EMCO/src/clm/pkg/cluster"
	controller "github.com/open-ness/EMCO/src/clm/pkg/controller"
	"github.com/open-ness/EMCO/src/clm/pkg/module"
)

var moduleClient *module.Client
var moduleController *module.Client

// For the given client and testClient, if the testClient is not null and
// implements the client manager interface corresponding to client, then
// return the testClient, otherwise return the client.
func setClient(client, testClient interface{}) interface{} {
	switch cl := client.(type) {
	case *cluster.ClusterClient:
		if testClient != nil && reflect.TypeOf(testClient).Implements(reflect.TypeOf((*cluster.ClusterManager)(nil)).Elem()) {
			c, ok := testClient.(cluster.ClusterManager)
			if ok {
				return c
			}
		}
	case *controller.ControllerClient:
		if testClient != nil && reflect.TypeOf(testClient).Implements(reflect.TypeOf((*controller.ControllerManager)(nil)).Elem()) {
			c, ok := testClient.(controller.ControllerManager)
			if ok {
				return c
			}
		}
	default:
		fmt.Printf("unknown type %T\n", cl)
	}
	return client
}

// NewRouter creates a router that registers the various urls that are supported
// testClient parameter allows unit testing for a given client
func NewRouter(testClient interface{}) *mux.Router {

	moduleClient = module.NewClient()
	moduleController = module.NewController()

	router := mux.NewRouter().PathPrefix("/v2").Subrouter()

	clusterHandler := clusterHandler{
		client: setClient(moduleClient.Cluster, testClient).(cluster.ClusterManager),
	}
	router.HandleFunc("/cluster-providers", clusterHandler.createClusterProviderHandler).Methods("POST")
	router.HandleFunc("/cluster-providers", clusterHandler.getClusterProviderHandler).Methods("GET")
	router.HandleFunc("/cluster-providers/{name}", clusterHandler.putClusterProviderHandler).Methods("PUT")
	router.HandleFunc("/cluster-providers/{name}", clusterHandler.getClusterProviderHandler).Methods("GET")
	router.HandleFunc("/cluster-providers/{name}", clusterHandler.deleteClusterProviderHandler).Methods("DELETE")
	router.HandleFunc("/cluster-providers/{provider-name}/clusters", clusterHandler.createClusterHandler).Methods("POST")
	router.HandleFunc("/cluster-providers/{provider-name}/clusters", clusterHandler.getClusterHandler).Methods("GET")
	router.HandleFunc("/cluster-providers/{provider-name}/clusters", clusterHandler.getClusterHandler).Queries("label", "{label}")
	router.HandleFunc("/cluster-providers/{provider-name}/clusters/{name}", clusterHandler.getClusterHandler).Methods("GET")
	router.HandleFunc("/cluster-providers/{provider-name}/clusters/{name}", clusterHandler.deleteClusterHandler).Methods("DELETE")
	router.HandleFunc("/cluster-providers/{provider-name}/clusters/{cluster-name}/labels", clusterHandler.createClusterLabelHandler).Methods("POST")
	router.HandleFunc("/cluster-providers/{provider-name}/clusters/{cluster-name}/labels", clusterHandler.getClusterLabelHandler).Methods("GET")
	router.HandleFunc("/cluster-providers/{provider-name}/clusters/{cluster-name}/labels/{label}", clusterHandler.putClusterLabelHandler).Methods("PUT")
	router.HandleFunc("/cluster-providers/{provider-name}/clusters/{cluster-name}/labels/{label}", clusterHandler.getClusterLabelHandler).Methods("GET")
	router.HandleFunc("/cluster-providers/{provider-name}/clusters/{cluster-name}/labels/{label}", clusterHandler.deleteClusterLabelHandler).Methods("DELETE")
	router.HandleFunc("/cluster-providers/{provider-name}/clusters/{cluster-name}/kv-pairs", clusterHandler.createClusterKvPairsHandler).Methods("POST")
	router.HandleFunc("/cluster-providers/{provider-name}/clusters/{cluster-name}/kv-pairs", clusterHandler.getClusterKvPairsHandler).Methods("GET")
	router.HandleFunc("/cluster-providers/{provider-name}/clusters/{cluster-name}/kv-pairs/{kvpair}", clusterHandler.putClusterKvPairsHandler).Methods("PUT")
	router.HandleFunc("/cluster-providers/{provider-name}/clusters/{cluster-name}/kv-pairs/{kvpair}", clusterHandler.getClusterKvPairsHandler).Methods("GET")
	router.HandleFunc("/cluster-providers/{provider-name}/clusters/{cluster-name}/kv-pairs/{kvpair}", clusterHandler.getClusterKvPairsHandler).Queries("key", "{key}")
	router.HandleFunc("/cluster-providers/{provider-name}/clusters/{cluster-name}/kv-pairs/{kvpair}", clusterHandler.deleteClusterKvPairsHandler).Methods("DELETE")

	controlHandler := controllerHandler{
		client: setClient(moduleController.Controller, testClient).(controller.ControllerManager),
	}
	router.HandleFunc("/clm-controllers", controlHandler.createHandler).Methods("POST")
	router.HandleFunc("/clm-controllers", controlHandler.getHandler).Methods("GET")
	router.HandleFunc("/clm-controllers/{controller-name}", controlHandler.putHandler).Methods("PUT")
	router.HandleFunc("/clm-controllers/{controller-name}", controlHandler.getHandler).Methods("GET")
	router.HandleFunc("/clm-controllers/{controller-name}", controlHandler.deleteHandler).Methods("DELETE")

	return router
}
