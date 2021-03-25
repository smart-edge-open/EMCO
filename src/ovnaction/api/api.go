// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2020 Intel Corporation
package api

import (
	"fmt"
	"reflect"

	"github.com/gorilla/mux"
	moduleLib "github.com/open-ness/EMCO/src/ovnaction/pkg/module"
)

var moduleClient *moduleLib.Client

// For the given client and testClient, if the testClient is not null and
// implements the client manager interface corresponding to client, then
// return the testClient, otherwise return the client.
func setClient(client, testClient interface{}) interface{} {
	switch cl := client.(type) {
	case *moduleLib.NetControlIntentClient:
		if testClient != nil && reflect.TypeOf(testClient).Implements(reflect.TypeOf((*moduleLib.NetControlIntentManager)(nil)).Elem()) {
			c, ok := testClient.(moduleLib.NetControlIntentManager)
			if ok {
				return c
			}
		}
	case *moduleLib.WorkloadIntentClient:
		if testClient != nil && reflect.TypeOf(testClient).Implements(reflect.TypeOf((*moduleLib.WorkloadIntentManager)(nil)).Elem()) {
			c, ok := testClient.(moduleLib.WorkloadIntentManager)
			if ok {
				return c
			}
		}
	case *moduleLib.WorkloadIfIntentClient:
		if testClient != nil && reflect.TypeOf(testClient).Implements(reflect.TypeOf((*moduleLib.WorkloadIfIntentManager)(nil)).Elem()) {
			c, ok := testClient.(moduleLib.WorkloadIfIntentManager)
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

	moduleClient = moduleLib.NewClient()

	router := mux.NewRouter().PathPrefix("/v2").Subrouter()

	netcontrolintentHandler := netcontrolintentHandler{
		client: setClient(moduleClient.NetControlIntent, testClient).(moduleLib.NetControlIntentManager),
	}
	router.HandleFunc("/projects/{project}/composite-apps/{composite-app-name}/{version}/deployment-intent-groups/{deployment-intent-group-name}/network-controller-intent", netcontrolintentHandler.createHandler).Methods("POST")
	router.HandleFunc("/projects/{project}/composite-apps/{composite-app-name}/{version}/deployment-intent-groups/{deployment-intent-group-name}/network-controller-intent", netcontrolintentHandler.getHandler).Methods("GET")
	router.HandleFunc("/projects/{project}/composite-apps/{composite-app-name}/{version}/deployment-intent-groups/{deployment-intent-group-name}/network-controller-intent/{name}", netcontrolintentHandler.putHandler).Methods("PUT")
	router.HandleFunc("/projects/{project}/composite-apps/{composite-app-name}/{version}/deployment-intent-groups/{deployment-intent-group-name}/network-controller-intent/{name}", netcontrolintentHandler.getHandler).Methods("GET")
	router.HandleFunc("/projects/{project}/composite-apps/{composite-app-name}/{version}/deployment-intent-groups/{deployment-intent-group-name}/network-controller-intent/{name}", netcontrolintentHandler.deleteHandler).Methods("DELETE")

	workloadintentHandler := workloadintentHandler{
		client: setClient(moduleClient.WorkloadIntent, testClient).(moduleLib.WorkloadIntentManager),
	}
	router.HandleFunc("/projects/{project}/composite-apps/{composite-app-name}/{version}/deployment-intent-groups/{deployment-intent-group-name}/network-controller-intent/{net-control-intent}/workload-intents", workloadintentHandler.createHandler).Methods("POST")
	router.HandleFunc("/projects/{project}/composite-apps/{composite-app-name}/{version}/deployment-intent-groups/{deployment-intent-group-name}/network-controller-intent/{net-control-intent}/workload-intents", workloadintentHandler.getHandler).Methods("GET")
	router.HandleFunc("/projects/{project}/composite-apps/{composite-app-name}/{version}/deployment-intent-groups/{deployment-intent-group-name}/network-controller-intent/{net-control-intent}/workload-intents/{name}", workloadintentHandler.putHandler).Methods("PUT")
	router.HandleFunc("/projects/{project}/composite-apps/{composite-app-name}/{version}/deployment-intent-groups/{deployment-intent-group-name}/network-controller-intent/{net-control-intent}/workload-intents/{name}", workloadintentHandler.getHandler).Methods("GET")
	router.HandleFunc("/projects/{project}/composite-apps/{composite-app-name}/{version}/deployment-intent-groups/{deployment-intent-group-name}/network-controller-intent/{net-control-intent}/workload-intents/{name}", workloadintentHandler.deleteHandler).Methods("DELETE")

	workloadifintentHandler := workloadifintentHandler{
		client: setClient(moduleClient.WorkloadIfIntent, testClient).(moduleLib.WorkloadIfIntentManager),
	}
	router.HandleFunc("/projects/{project}/composite-apps/{composite-app-name}/{version}/deployment-intent-groups/{deployment-intent-group-name}/network-controller-intent/{net-control-intent}/workload-intents/{workload-intent}/interfaces", workloadifintentHandler.createHandler).Methods("POST")
	router.HandleFunc("/projects/{project}/composite-apps/{composite-app-name}/{version}/deployment-intent-groups/{deployment-intent-group-name}/network-controller-intent/{net-control-intent}/workload-intents/{workload-intent}/interfaces", workloadifintentHandler.getHandler).Methods("GET")
	router.HandleFunc("/projects/{project}/composite-apps/{composite-app-name}/{version}/deployment-intent-groups/{deployment-intent-group-name}/network-controller-intent/{net-control-intent}/workload-intents/{workload-intent}/interfaces/{name}", workloadifintentHandler.putHandler).Methods("PUT")
	router.HandleFunc("/projects/{project}/composite-apps/{composite-app-name}/{version}/deployment-intent-groups/{deployment-intent-group-name}/network-controller-intent/{net-control-intent}/workload-intents/{workload-intent}/interfaces/{name}", workloadifintentHandler.getHandler).Methods("GET")
	router.HandleFunc("/projects/{project}/composite-apps/{composite-app-name}/{version}/deployment-intent-groups/{deployment-intent-group-name}/network-controller-intent/{net-control-intent}/workload-intents/{workload-intent}/interfaces/{name}", workloadifintentHandler.deleteHandler).Methods("DELETE")

	return router
}
