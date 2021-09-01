// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2020 Intel Corporation

package api

import (
	"reflect"

	"github.com/gorilla/mux"

	log "github.com/open-ness/EMCO/src/orchestrator/pkg/infra/logutils"

	moduleLib "github.com/open-ness/EMCO/src/hpa-plc/pkg/module"
)

var moduleClient *moduleLib.HpaPlacementClient
var hpaIntentJSONFile string = "json-schemas/placement-hpa-intent.json"
var hpaConsumerJSONFile string = "json-schemas/placement-hpa-consumer.json"
var hpaResourceJSONFile string = "json-schemas/placement-hpa-resource.json"

// HpaPlacementIntentHandler .. Used to store backend implementations objects
// Also simplifies mocking for unit testing purposes
type HpaPlacementIntentHandler struct {
	// Interface that implements Cluster operations
	// We will set this variable with a mock interface for testing
	client moduleLib.HpaPlacementManager
}

// For the given client and testClient, if the testClient is not null and
// implements the client manager interface corresponding to client, then
// return the testClient, otherwise return the client.
func setClient(client, testClient interface{}) interface{} {
	switch cl := client.(type) {
	case *moduleLib.HpaPlacementClient:
		if testClient != nil && reflect.TypeOf(testClient).Implements(reflect.TypeOf((*moduleLib.HpaPlacementManager)(nil)).Elem()) {
			c, ok := testClient.(moduleLib.HpaPlacementManager)
			if ok {
				return c
			}
		}
	default:
		log.Error(":: setClient .. unknown type ::", log.Fields{"client-type": cl})
	}
	return client
}

// NewRouter creates a router that registers the various urls that are supported
// testClient parameter allows unit testing for a given client
func NewRouter(testClient interface{}) *mux.Router {
	moduleClient = moduleLib.NewHpaPlacementClient()

	router := mux.NewRouter().PathPrefix("/v2").Subrouter()

	hpaPlacementIntentHandler := HpaPlacementIntentHandler{
		client: setClient(moduleClient, testClient).(moduleLib.HpaPlacementManager),
	}

	const emcoHpaIntentsURL = "/projects/{project-name}/composite-apps/{composite-app-name}/{composite-app-version}/deployment-intent-groups/{deployment-intent-group-name}/hpa-intents"
	const emcoHpaIntentsGetURL = emcoHpaIntentsURL + "/{intent-name}"
	const emcoHpaConsumersURL = emcoHpaIntentsGetURL + "/hpa-resource-consumers"
	const emcoHpaConsumersGetURL = emcoHpaConsumersURL + "/{consumer-name}"
	const emcoHpaResourcesURL = emcoHpaConsumersGetURL + "/resource-requirements"
	const emcoHpaResourcesGetURL = emcoHpaResourcesURL + "/{resource-name}"

	// hpa-intent => /projects/{project-name}/composite-apps/{composite-app-name}/{composite-app-version}/deployment-intent-groups/{deployment-intent-group-name}/hpa-intents
	router.HandleFunc(emcoHpaIntentsURL, hpaPlacementIntentHandler.addHpaIntentHandler).Methods("POST")
	router.HandleFunc(emcoHpaIntentsGetURL, hpaPlacementIntentHandler.getHpaIntentHandler).Methods("GET")
	router.HandleFunc(emcoHpaIntentsURL, hpaPlacementIntentHandler.getHpaIntentHandler).Methods("GET")
	router.HandleFunc(emcoHpaIntentsURL, hpaPlacementIntentHandler.getHpaIntentByNameHandler).Queries("intent", "{intent-name}")
	router.HandleFunc(emcoHpaIntentsGetURL, hpaPlacementIntentHandler.putHpaIntentHandler).Methods("PUT")
	router.HandleFunc(emcoHpaIntentsGetURL, hpaPlacementIntentHandler.deleteHpaIntentHandler).Methods("DELETE")
	router.HandleFunc(emcoHpaIntentsURL, hpaPlacementIntentHandler.deleteAllHpaIntentsHandler).Methods("DELETE")

	// hpa-consumer => /projects/{project-name}/composite-apps/{composite-app-name}/{composite-app-version}/deployment-intent-groups/{deployment-intent-group-name}/hpa-intents/{intent-name}/hpa-resource-consumers
	router.HandleFunc(emcoHpaConsumersURL, hpaPlacementIntentHandler.addHpaConsumerHandler).Methods("POST")
	router.HandleFunc(emcoHpaConsumersGetURL, hpaPlacementIntentHandler.getHpaConsumerHandler).Methods("GET")
	router.HandleFunc(emcoHpaConsumersURL, hpaPlacementIntentHandler.getHpaConsumerHandler).Methods("GET")
	router.HandleFunc(emcoHpaConsumersURL, hpaPlacementIntentHandler.getHpaConsumerHandlerByName).Queries("consumer", "{consumer-name}")
	router.HandleFunc(emcoHpaConsumersGetURL, hpaPlacementIntentHandler.putHpaConsumerHandler).Methods("PUT")
	router.HandleFunc(emcoHpaConsumersGetURL, hpaPlacementIntentHandler.deleteHpaConsumerHandler).Methods("DELETE")
	router.HandleFunc(emcoHpaConsumersURL, hpaPlacementIntentHandler.deleteAllHpaConsumersHandler).Methods("DELETE")

	// hpa-resource => /projects/{project-name}/composite-apps/{composite-app-name}/{composite-app-version}/deployment-intent-groups/{deployment-intent-group-name}/hpa-intents/{intent-name}/hpa-resource-consumers/{consumer-name}/resource-requirements
	router.HandleFunc(emcoHpaResourcesURL, hpaPlacementIntentHandler.addHpaResourceHandler).Methods("POST")
	router.HandleFunc(emcoHpaResourcesGetURL, hpaPlacementIntentHandler.getHpaResourceHandler).Methods("GET")
	router.HandleFunc(emcoHpaResourcesURL, hpaPlacementIntentHandler.getHpaResourceHandler).Methods("GET")
	router.HandleFunc(emcoHpaResourcesURL, hpaPlacementIntentHandler.getHpaResourceHandlerByName).Queries("resource", "{resource-name}")
	router.HandleFunc(emcoHpaResourcesGetURL, hpaPlacementIntentHandler.putHpaResourceHandler).Methods("PUT")
	router.HandleFunc(emcoHpaResourcesGetURL, hpaPlacementIntentHandler.deleteHpaResourceHandler).Methods("DELETE")
	router.HandleFunc(emcoHpaResourcesURL, hpaPlacementIntentHandler.deleteAllHpaResourcesHandler).Methods("DELETE")

	return router
}
