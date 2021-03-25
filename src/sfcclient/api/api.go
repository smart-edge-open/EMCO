// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2021 Intel Corporation
package api

import (
	"fmt"
	"reflect"

	"github.com/gorilla/mux"
	"github.com/open-ness/EMCO/src/sfcclient/pkg/module"
)

var moduleClient *module.SfcClient

// For the given client and testClient, if the testClient is not null and
// implements the client manager interface corresponding to client, then
// return the testClient, otherwise return the client.
func setClient(client, testClient interface{}) interface{} {
	switch cl := client.(type) {
	case *module.SfcClient:
		if testClient != nil && reflect.TypeOf(testClient).Implements(reflect.TypeOf((*module.SfcManager)(nil)).Elem()) {
			c, ok := testClient.(module.SfcManager)
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

	moduleClient = module.NewSfcClient()

	router := mux.NewRouter().PathPrefix("/v2").Subrouter()

	const sfcClientIntentsURL = "/projects/{project-name}/composite-apps/{composite-app-name}/{composite-app-version}/deployment-intent-groups/{deployment-intent-group-name}/network-controller-intent/{net-control-intent}/sfc-clients"
	const sfcClientIntentsGetURL = sfcClientIntentsURL + "/{sfc-client-name}"

	sfcHandler := sfcHandler{
		client: setClient(moduleClient, testClient).(module.SfcManager),
	}
	router.HandleFunc(sfcClientIntentsURL, sfcHandler.createHandler).Methods("POST")
	router.HandleFunc(sfcClientIntentsURL, sfcHandler.getHandler).Methods("GET")
	router.HandleFunc(sfcClientIntentsGetURL, sfcHandler.putHandler).Methods("PUT")
	router.HandleFunc(sfcClientIntentsGetURL, sfcHandler.getHandler).Methods("GET")
	router.HandleFunc(sfcClientIntentsGetURL, sfcHandler.deleteHandler).Methods("DELETE")

	return router
}
