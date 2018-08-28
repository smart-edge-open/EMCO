// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2020 Intel Corporation
package api

import (
	"fmt"
	"reflect"

	"github.com/gorilla/mux"
	"github.com/open-ness/EMCO/src/dtc/pkg/module"
)

var moduleClient *module.Client

// For the given client and testClient, if the testClient is not null and
// implements the client manager interface corresponding to client, then
// return the testClient, otherwise return the client.
func setClient(client, testClient interface{}) interface{} {
        switch cl := client.(type) {
        case *module.TrafficGroupIntentDbClient:
                if testClient != nil && reflect.TypeOf(testClient).Implements(reflect.TypeOf((*module.TrafficGroupIntentManager)(nil)).Elem()) {
                        c, ok := testClient.(module.TrafficGroupIntentManager)
                        if ok {
                                return c
                        }
                }
        case *module.InboundServerIntentDbClient:
                if testClient != nil && reflect.TypeOf(testClient).Implements(reflect.TypeOf((*module.InboundServerIntentManager)(nil)).Elem()) {
                        c, ok := testClient.(module.InboundServerIntentManager)
                        if ok {
                                return c
                        }
                }
        case *module.InboundClientsIntentDbClient:
                if testClient != nil && reflect.TypeOf(testClient).Implements(reflect.TypeOf((*module.InboundClientsIntentManager)(nil)).Elem()) {
                        c, ok := testClient.(module.InboundClientsIntentManager)
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
func NewRouter(testClient interface{}) *mux.Router {

	moduleClient = module.NewClient()

	router := mux.NewRouter().PathPrefix("/v2").Subrouter()
	trafficgroupintentHandler := trafficgroupintentHandler{
		client: setClient(moduleClient.TrafficGroupIntent, testClient).(module.TrafficGroupIntentManager),
	}
	router.HandleFunc("/projects/{project}/composite-apps/{composite-app-name}/{version}/deployment-intent-groups/{deployment-intent-group-name}/traffic-group-intents", trafficgroupintentHandler.createHandler).Methods("POST")
	router.HandleFunc("/projects/{project}/composite-apps/{composite-app-name}/{version}/deployment-intent-groups/{deployment-intent-group-name}/traffic-group-intents", trafficgroupintentHandler.getHandler).Methods("GET")
	router.HandleFunc("/projects/{project}/composite-apps/{composite-app-name}/{version}/deployment-intent-groups/{deployment-intent-group-name}/traffic-group-intents/{name}", trafficgroupintentHandler.getHandler).Methods("GET")
	router.HandleFunc("/projects/{project}/composite-apps/{composite-app-name}/{version}/deployment-intent-groups/{deployment-intent-group-name}/traffic-group-intents/{name}", trafficgroupintentHandler.putHandler).Methods("PUT")
	router.HandleFunc("/projects/{project}/composite-apps/{composite-app-name}/{version}/deployment-intent-groups/{deployment-intent-group-name}/traffic-group-intents/{name}", trafficgroupintentHandler.deleteHandler).Methods("DELETE")

	inboundserverintentHandler := inboundserverintentHandler{
		client: setClient(moduleClient.ServerInboundIntent, testClient).(module.InboundServerIntentManager),
	}
	router.HandleFunc("/projects/{project}/composite-apps/{composite-app-name}/{version}/deployment-intent-groups/{deployment-intent-group-name}/traffic-group-intents/{traffic-group-intent-name}/inbound-intents", inboundserverintentHandler.createHandler).Methods("POST")
	router.HandleFunc("/projects/{project}/composite-apps/{composite-app-name}/{version}/deployment-intent-groups/{deployment-intent-group-name}/traffic-group-intents/{traffic-group-intent-name}/inbound-intents/{name}", inboundserverintentHandler.getHandler).Methods("GET")
	router.HandleFunc("/projects/{project}/composite-apps/{composite-app-name}/{version}/deployment-intent-groups/{deployment-intent-group-name}/traffic-group-intents/{traffic-group-intent-name}/inbound-intents", inboundserverintentHandler.getHandler).Methods("GET")
	router.HandleFunc("/projects/{project}/composite-apps/{composite-app-name}/{version}/deployment-intent-groups/{deployment-intent-group-name}/traffic-group-intents/{traffic-group-intent-name}/inbound-intents/{name}", inboundserverintentHandler.putHandler).Methods("PUT")
	router.HandleFunc("/projects/{project}/composite-apps/{composite-app-name}/{version}/deployment-intent-groups/{deployment-intent-group-name}/traffic-group-intents/{traffic-group-intent-name}/inbound-intents/{name}", inboundserverintentHandler.deleteHandler).Methods("DELETE")

	inboundclientsintentHandler := inboundclientsintentHandler{
		client: setClient(moduleClient.ClientsInboundIntent, testClient).(module.InboundClientsIntentManager),
	}
	router.HandleFunc("/projects/{project}/composite-apps/{composite-app-name}/{version}/deployment-intent-groups/{deployment-intent-group-name}/traffic-group-intents/{traffic-group-intent-name}/inbound-intents/{intent-name}/clients", inboundclientsintentHandler.createHandler).Methods("POST")
	router.HandleFunc("/projects/{project}/composite-apps/{composite-app-name}/{version}/deployment-intent-groups/{deployment-intent-group-name}/traffic-group-intents/{traffic-group-intent-name}/inbound-intents/{intent-name}/clients", inboundclientsintentHandler.getHandler).Methods("GET")
	router.HandleFunc("/projects/{project}/composite-apps/{composite-app-name}/{version}/deployment-intent-groups/{deployment-intent-group-name}/traffic-group-intents/{traffic-group-intent-name}/inbound-intents/{intent-name}/clients/{name}", inboundclientsintentHandler.getHandler).Methods("GET")
	router.HandleFunc("/projects/{project}/composite-apps/{composite-app-name}/{version}/deployment-intent-groups/{deployment-intent-group-name}/traffic-group-intents/{traffic-group-intent-name}/inbound-intents/{intent-name}/clients/{name}", inboundclientsintentHandler.putHandler).Methods("PUT")
	router.HandleFunc("/projects/{project}/composite-apps/{composite-app-name}/{version}/deployment-intent-groups/{deployment-intent-group-name}/traffic-group-intents/{traffic-group-intent-name}/inbound-intents/{intent-name}/clients/{name}", inboundclientsintentHandler.deleteHandler).Methods("DELETE")

	return router
}
