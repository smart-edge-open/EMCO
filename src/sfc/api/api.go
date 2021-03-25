// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2021 Intel Corporation
package api

import (
	"fmt"
	"reflect"

	"github.com/gorilla/mux"
	"github.com/open-ness/EMCO/src/sfc/pkg/module"
)

var moduleSfcIntentClient *module.SfcIntentClient
var moduleSfcClientSelectorIntentClient *module.SfcClientSelectorIntentClient
var moduleSfcProviderNetworkIntentClient *module.SfcProviderNetworkIntentClient

// Used to store backend implementations objects
// Also simplifies mocking for unit testing purposes
type sfcIntentHandler struct {
	// Interface that implements SFC intent operations
	// We will set this variable with a mock interface for testing
	client module.SfcIntentManager
}
type sfcClientSelectorIntentHandler struct {
	// Interface that implements SFC intent operations
	// We will set this variable with a mock interface for testing
	client module.SfcClientSelectorIntentManager
}
type sfcProviderNetworkIntentHandler struct {
	// Interface that implements SFC intent operations
	// We will set this variable with a mock interface for testing
	client module.SfcProviderNetworkIntentManager
}

// For the given client and testClient, if the testClient is not null and
// implements the client manager interface corresponding to client, then
// return the testClient, otherwise return the client.
func setClient(client, testClient interface{}) interface{} {
	switch cl := client.(type) {
	case *module.SfcIntentClient:
		if testClient != nil && reflect.TypeOf(testClient).Implements(reflect.TypeOf((*module.SfcIntentManager)(nil)).Elem()) {
			c, ok := testClient.(module.SfcIntentManager)
			if ok {
				return c
			}
		}
	case *module.SfcClientSelectorIntentClient:
		if testClient != nil && reflect.TypeOf(testClient).Implements(reflect.TypeOf((*module.SfcClientSelectorIntentManager)(nil)).Elem()) {
			c, ok := testClient.(module.SfcClientSelectorIntentManager)
			if ok {
				return c
			}
		}
	case *module.SfcProviderNetworkIntentClient:
		if testClient != nil && reflect.TypeOf(testClient).Implements(reflect.TypeOf((*module.SfcProviderNetworkIntentManager)(nil)).Elem()) {
			c, ok := testClient.(module.SfcProviderNetworkIntentManager)
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

	moduleClient := module.NewClient()

	router := mux.NewRouter().PathPrefix("/v2").Subrouter()

	const sfcIntentsURL = "/projects/{project-name}/composite-apps/{composite-app-name}/{composite-app-version}/deployment-intent-groups/{deployment-intent-group-name}/network-controller-intent/{net-control-intent}/network-chains"
	const sfcIntentsGetURL = sfcIntentsURL + "/{sfc-name}"
	const sfcClientSelectorIntentsURL = sfcIntentsGetURL + "/client-selectors"
	const sfcClientSelectorIntentsGetURL = sfcClientSelectorIntentsURL + "/{sfc-client-selector}"
	const sfcProviderNetworkIntentsURL = sfcIntentsGetURL + "/provider-networks"
	const sfcProviderNetworkIntentsGetURL = sfcProviderNetworkIntentsURL + "/{sfc-provider-network}"

	sfcHandler := sfcIntentHandler{
		client: setClient(moduleClient.SfcIntent, testClient).(module.SfcIntentManager),
	}
	router.HandleFunc(sfcIntentsURL, sfcHandler.createSfcHandler).Methods("POST")
	router.HandleFunc(sfcIntentsURL, sfcHandler.getSfcHandler).Methods("GET")
	router.HandleFunc(sfcIntentsGetURL, sfcHandler.putSfcHandler).Methods("PUT")
	router.HandleFunc(sfcIntentsGetURL, sfcHandler.getSfcHandler).Methods("GET")
	router.HandleFunc(sfcIntentsGetURL, sfcHandler.deleteSfcHandler).Methods("DELETE")

	sfcClientSelectorHandler := sfcClientSelectorIntentHandler{
		client: setClient(moduleClient.SfcClientSelectorIntent, testClient).(module.SfcClientSelectorIntentManager),
	}
	router.HandleFunc(sfcClientSelectorIntentsURL, sfcClientSelectorHandler.createClientSelectorHandler).Methods("POST")
	router.HandleFunc(sfcClientSelectorIntentsURL, sfcClientSelectorHandler.getClientSelectorHandler).Methods("GET")
	router.HandleFunc(sfcClientSelectorIntentsGetURL, sfcClientSelectorHandler.putClientSelectorHandler).Methods("PUT")
	router.HandleFunc(sfcClientSelectorIntentsGetURL, sfcClientSelectorHandler.getClientSelectorHandler).Methods("GET")
	router.HandleFunc(sfcClientSelectorIntentsGetURL, sfcClientSelectorHandler.deleteClientSelectorHandler).Methods("DELETE")

	sfcProviderNetworkHandler := sfcProviderNetworkIntentHandler{
		client: setClient(moduleClient.SfcProviderNetworkIntent, testClient).(module.SfcProviderNetworkIntentManager),
	}
	router.HandleFunc(sfcProviderNetworkIntentsURL, sfcProviderNetworkHandler.createProviderNetworkHandler).Methods("POST")
	router.HandleFunc(sfcProviderNetworkIntentsURL, sfcProviderNetworkHandler.getProviderNetworkHandler).Methods("GET")
	router.HandleFunc(sfcProviderNetworkIntentsGetURL, sfcProviderNetworkHandler.putProviderNetworkHandler).Methods("PUT")
	router.HandleFunc(sfcProviderNetworkIntentsGetURL, sfcProviderNetworkHandler.getProviderNetworkHandler).Methods("GET")
	router.HandleFunc(sfcProviderNetworkIntentsGetURL, sfcProviderNetworkHandler.deleteProviderNetworkHandler).Methods("DELETE")

	return router
}
