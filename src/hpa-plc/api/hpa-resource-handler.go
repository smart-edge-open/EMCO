// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2020 Intel Corporation

package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httputil"
	"strings"

	"github.com/gorilla/mux"
	pkgerrors "github.com/pkg/errors"

	log "github.com/open-ness/EMCO/src/orchestrator/pkg/infra/logutils"
	"github.com/open-ness/EMCO/src/orchestrator/pkg/infra/validation"

	hpaModel "github.com/open-ness/EMCO/src/hpa-plc/pkg/model"
)

/*
addHpaResourceHandler handles the URL
URL: /v2/projects/{project-name}/composite-apps/{composite-app-name}/{version}/
deployment-intent-groups/{deployment-intent-group-name}/hpa-intents/{intent-name}/hpa-resource-consumers/{consumer-name}/resource-requirements
*/
// Add Hpa Intent resource
func (h HpaPlacementIntentHandler) addHpaResourceHandler(w http.ResponseWriter, r *http.Request) {
	var hpa hpaModel.HpaResourceRequirement
	reqDump, _ := httputil.DumpRequest(r, true)
	log.Info(":: addHpaResourceHandler .. start ::", log.Fields{"req": string(reqDump)})

	err := json.NewDecoder(r.Body).Decode(&hpa)
	switch {
	case err == io.EOF:
		log.Error(":: addHpaResourceHandler .. Empty addHpaResourceHandler POST body ::", log.Fields{"Error": err})
		http.Error(w, "Empty body", http.StatusBadRequest)
		return
	case err != nil:
		log.Error(":: addHpaResourceHandler .. Error decoding addHpaResourceHandler POST body ::", log.Fields{"Error": err})
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	// Verify JSON Body
	err, httpError := validation.ValidateJsonSchemaData(hpaResourceJSONFile, hpa)
	if err != nil {
		log.Error(":: addHpaResourceHandler .. JSON validation failed ::", log.Fields{"Error": err})
		http.Error(w, err.Error(), httpError)
		return
	}

	vars := mux.Vars(r)
	p := vars["project-name"]
	ca := vars["composite-app-name"]
	v := vars["composite-app-version"]
	di := vars["deployment-intent-group-name"]
	i := vars["intent-name"]
	cn := vars["consumer-name"]

	// check resource dependencies(consumer) validity
	if !validateDependents(&w, &h, &hpa, p, ca, v, di, i, cn) {
		return
	}

	log.Info(":: AddResource .. Req ::", log.Fields{"project": p, "composite-app": ca, "composite-app-ver": v, "dep-group": di, "intent-name": i, "consumer-name": cn, "resource-name": hpa.MetaData.Name})
	resource, err := h.client.AddResource(hpa, p, ca, v, di, i, cn, false)
	if err != nil {
		log.Error(":: addHpaResourceHandler .. AddResource error ::", log.Fields{"Error": err})
		if strings.Contains(err.Error(), "not found") {
			http.Error(w, err.Error(), http.StatusNotFound)
		} else if (strings.Contains(err.Error(), "conflict")) || (strings.Contains(err.Error(), "already exists")) {
			http.Error(w, err.Error(), http.StatusConflict)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	err = json.NewEncoder(w).Encode(resource)
	if err != nil {
		log.Error(":: addHpaResourceHandler ..  Encoder error ::", log.Fields{"Error": err})
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	log.Info(":: addHpaResourceHandler .. end ::", log.Fields{"resource": resource})
}

/*
getHpaResourceHandlerByName handles the URL
URL: /v2/projects/{project-name}/composite-apps/{composite-app-name}/{version}/
deployment-intent-groups/{deployment-intent-group-name}/hpa-intents/{intent-name}/hpa-resource-consumers/{consumer-name}/resource-requirements?resource=<resource>
*/
// Query Hpa Intent resource
func (h HpaPlacementIntentHandler) getHpaResourceHandlerByName(w http.ResponseWriter, r *http.Request) {
	reqDump, _ := httputil.DumpRequest(r, true)
	log.Info(":: getHpaResourceHandlerByName .. start ::", log.Fields{"req": string(reqDump)})

	p, ca, v, di, i, cn, _, err := parseHpaResourceReqParameters(&w, r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	rN := r.URL.Query().Get("resource")
	if rN == "" {
		log.Error(":: getHpaResourceHandlerByName .. Missing intent-name in request ::", log.Fields{"Error": http.StatusBadRequest})
		http.Error(w, "Missing intent-name in request", http.StatusBadRequest)
		return
	}

	resource, err := h.client.GetResourceByName(rN, p, ca, v, di, i, cn)
	if err != nil {
		log.Error(":: getHpaResourceHandlerByName .. GetIntentByName error ::", log.Fields{"Error": err})
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(resource)
	if err != nil {
		log.Error(":: getHpaResourceHandlerByName .. Encoder error ::", log.Fields{"Error": err})
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	log.Info(":: getHpaResourceHandlerByName .. end ::", log.Fields{"resource": resource})
}

/*
getHpaResourceHandler/getHpaResourceHandlers handles the URL
URL: /v2/projects/{project-name}/composite-apps/{composite-app-name}/{version}/
deployment-intent-groups/{deployment-intent-group-name}/hpa-intents/{intent-name}/hpa-resource-consumers/{consumer-name}/resource-requirements/{resource-name}
*/
// Get Hpa Intent resource
func (h HpaPlacementIntentHandler) getHpaResourceHandler(w http.ResponseWriter, r *http.Request) {
	reqDump, _ := httputil.DumpRequest(r, true)
	log.Info(":: getHpaResourceHandler .. start ::", log.Fields{"req": string(reqDump)})
	p, ca, v, di, i, cn, name, err := parseHpaResourceReqParameters(&w, r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	log.Info(":: getHpaResourceHandler .. Req ::", log.Fields{"project": p, "composite-app": ca, "composite-app-ver": v, "dep-group": di, "intent-name": name})

	var resources interface{}
	if len(name) == 0 {
		resources, err = h.client.GetAllResources(p, ca, v, di, i, cn)
		if err != nil {
			log.Error(":: getHpaResourceHandler .. GetAllResources failure ::", log.Fields{"Error": err})
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
	} else {
		resources, _, err = h.client.GetResource(name, p, ca, v, di, i, cn)
		if err != nil {
			log.Error(":: getHpaResourceHandler .. GetResource failure ::", log.Fields{"Error": err})
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(resources)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	log.Info(":: getHpaResourceHandler .. end ::", log.Fields{"resources": resources})
}

/*
putHpaResourceHandler handles the URL
URL: /v2/projects/{project-name}/composite-apps/{composite-app-name}/{version}/
deployment-intent-groups/{deployment-intent-group-name}/hpa-intents/{intent-name}/hpa-resource-consumers/{consumer-name}/resource-requirements/{resource-name}
*/
// Update Hpa Intent resource
func (h HpaPlacementIntentHandler) putHpaResourceHandler(w http.ResponseWriter, r *http.Request) {
	var hpa hpaModel.HpaResourceRequirement
	reqDump, _ := httputil.DumpRequest(r, true)
	log.Info(":: putHpaResourceHandler .. start ::", log.Fields{"req": string(reqDump)})

	err := json.NewDecoder(r.Body).Decode(&hpa)
	switch {
	case err == io.EOF:
		log.Error(":: putHpaResourceHandler .. Empty PUT body ::", log.Fields{"Error": err})
		http.Error(w, "putHpaResourceHandler .. Empty body", http.StatusBadRequest)
		return
	case err != nil:
		log.Error(":: putHpaResourceHandler .. decoding resource PUT body ::", log.Fields{"Error": err})
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	// Verify JSON Body
	err, httpError := validation.ValidateJsonSchemaData(hpaResourceJSONFile, hpa)
	if err != nil {
		log.Error(":: putHpaResourceHandler .. JSON validation failed ::", log.Fields{"Error": err})
		http.Error(w, err.Error(), httpError)
		return
	}

	// Validate Hpa resource req
	p, ca, v, di, i, cn, name, err := parseHpaResourceReqParameters(&w, r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// check resource dependencies(consumer) validity
	if !validateDependents(&w, &h, &hpa, p, ca, v, di, i, cn) {
		return
	}

	// Name in URL should match name in body
	if hpa.MetaData.Name != name {
		log.Error(":: putHpaResourceHandler .. Mismatched name in PUT request ::", log.Fields{"bodyname": hpa.MetaData.Name, "name": name})
		http.Error(w, "putHpaResourceHandler .. Mismatched name in PUT request", http.StatusBadRequest)
		return
	}

	log.Info(":: putHpaResourceHandler .. Req ::", log.Fields{"project": p, "composite-app": ca, "composite-app-ver": v, "dep-group": di, "intent-name": name})
	resource, err := h.client.AddResource(hpa, p, ca, v, di, i, cn, true)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			http.Error(w, err.Error(), http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(resource)
	if err != nil {
		log.Error(":: putHpaResourceHandler .. encoding failure ::", log.Fields{"Error": err})
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	log.Info(":: putHpaResourceHandler .. end ::", log.Fields{"req": string(reqDump)})
}

/*
deleteHpaResourceHandler handles the URL
URL: /v2/projects/{project-name}/composite-apps/{composite-app-name}/{version}/
deployment-intent-groups/{deployment-intent-group-name}/hpa-intents/{intent-name}/hpa-resource-consumers/{consumer-name}/resource-requirements/{resource-name}
*/
// Delete Hpa Intent resource
func (h HpaPlacementIntentHandler) deleteHpaResourceHandler(w http.ResponseWriter, r *http.Request) {
	reqDump, _ := httputil.DumpRequest(r, true)
	log.Info(":: deleteHpaResourceHandler .. start ::", log.Fields{"req": string(reqDump)})

	// Validate Hpa resource req
	p, ca, v, di, i, cn, name, err := parseHpaResourceReqParameters(&w, r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	log.Info(":: deleteHpaResourceHandler .. Req ::", log.Fields{"project": p, "composite-app": ca, "composite-app-ver": v, "dep-group": di, "intent-name": i, "consumer-name": cn, "resource-name": name})

	_, _, err = h.client.GetResource(name, p, ca, v, di, i, cn)
	if err != nil {
		log.Error(":: deleteHpaResourceHandler .. GetResource failure ::", log.Fields{"Error": err})
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	err = h.client.DeleteResource(name, p, ca, v, di, i, cn)
	if err != nil {
		log.Error(":: deleteHpaResourceHandler .. DeleteResource failure ::", log.Fields{"Error": err})
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
	log.Info(":: deleteHpaResourceHandler .. end ::", log.Fields{"req": string(reqDump)})
}

/*
deleteAllHpaResourcesHandler handles the URL
URL: /v2/projects/{project-name}/composite-apps/{composite-app-name}/{version}/
deployment-intent-groups/{deployment-intent-group-name}/hpa-intents/{intent-name}/hpa-resource-consumers/{consumer-name}/resource-requirements
*/
// Delete Hpa Intent resource
func (h HpaPlacementIntentHandler) deleteAllHpaResourcesHandler(w http.ResponseWriter, r *http.Request) {
	reqDump, _ := httputil.DumpRequest(r, true)
	log.Info(":: deleteAllHpaResourceHandler .. start ::", log.Fields{"req": string(reqDump)})

	// Validate Hpa resource req
	p, ca, v, di, i, cn, name, err := parseHpaResourceReqParameters(&w, r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	log.Info(":: deleteAllHpaResourcesHandler .. Req ::", log.Fields{"project": p, "composite-app": ca, "composite-app-ver": v, "dep-group": di, "intent-name": i, "consumer-name": cn, "resource-name": name})

	hpaResources, err := h.client.GetAllResources(p, ca, v, di, i, cn)
	if err != nil {
		log.Error(":: deleteAllHpaResourcesHandler .. GetAllResources failure ::", log.Fields{"Error": err})
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	for _, hpaResource := range hpaResources {
		err = h.client.DeleteResource(hpaResource.MetaData.Name, p, ca, v, di, i, cn)
		if err != nil {
			log.Error(":: deleteAllHpaResourcesHandler .. DeleteResource failure ::", log.Fields{"Error": err})
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
	w.WriteHeader(http.StatusNoContent)
	log.Info(":: deleteAllHpaResourcesHandler .. end ::", log.Fields{"req": string(reqDump)})
}

/* Parse Http request Parameters */
func parseHpaResourceReqParameters(w *http.ResponseWriter, r *http.Request) (string, string, string, string, string, string, string, error) {
	vars := mux.Vars(r)

	rn := vars["resource-name"]

	cn := vars["consumer-name"]
	if cn == "" {
		log.Error(":: parseHpaConsumerReqParameters ..  Missing consumerName in request ::", log.Fields{"Error": http.StatusBadRequest})
		http.Error(*w, "parseHpaConsumerReqParameters .. Missing name of consumerName in request", http.StatusBadRequest)
		return "", "", "", "", "", "", "", pkgerrors.New("Missing consumer-name")
	}

	p, ca, v, di, i, cn, err := parseHpaConsumerReqParameters(w, r)
	if err != nil {
		log.Error(":: parseHpaResourceReqParameters .. Failed Consumer validation ::", log.Fields{"Error": http.StatusBadRequest})
		http.Error(*w, "parseHpaResourceReqParameters .. Failed Consumer validation", http.StatusBadRequest)
		return "", "", "", "", "", "", "", err
	}

	return p, ca, v, di, i, cn, rn, nil
}

/* Valdate consumer spec */
func validateConsumerSpec(w *http.ResponseWriter, name string) {
	if name == "" {
		err := fmt.Errorf("param[%s] is empty", name)
		log.Error(":: addHpaResourceHandler .. JSON validation failed ::", log.Fields{"Error": err})
		http.Error(*w, err.Error(), http.StatusBadRequest)
		return
	}
}

func validateDependents(w *http.ResponseWriter, h *HpaPlacementIntentHandler, hpa *hpaModel.HpaResourceRequirement, p string, ca string, v string, di string, i string, cn string) bool {
	log.Info(":: validateDependents .. start", log.Fields{"hpa-resource": hpa})

	//Check for the Consumer already exists here.
	res, _, err := h.client.GetConsumer(cn, p, ca, v, di, i)
	if err != nil {
		log.Error(":: validateDependents .. Consumer not found.", log.Fields{"consumer-name": cn, "Error": err})
		http.Error(*w, err.Error(), http.StatusNotFound)
		return false
	}

	// validate spec
	deployName := res.Spec.Name
	containerName := res.Spec.ContainerName
	// if non-allocatable
	if !(*hpa.Spec.Allocatable) {
		if deployName == "" {
			validateConsumerSpec(w, deployName)
			return false
		}

		if (hpa.Spec.Resource.NonAllocatableResources == hpaModel.NonAllocatableResources{}) {
			err := fmt.Errorf("JsonSchemaValidation: Document Validation failed .. resource-spec allocatable and resource field value mismatch. hpa-resource-name[%v] allocatable[%v] hpa-resource-spec[%v]", hpa.MetaData.Name, hpa.Spec.Allocatable, hpa.Spec)
			log.Error(":: validateDependents .. JSON validation failed ::", log.Fields{"Error": err})
			http.Error(*w, err.Error(), http.StatusBadRequest)
			return false
		}
	} else {
		if deployName == "" {
			validateConsumerSpec(w, deployName)
			return false
		}

		if containerName == "" {
			validateConsumerSpec(w, containerName)
			return false
		}

		if (hpa.Spec.Resource.AllocatableResources == hpaModel.AllocatableResources{}) {
			err := fmt.Errorf("JsonSchemaValidation: Document Validation failed .. resource-spec allocatable and resource field value mismatch. hpa-resource-name[%v] allocatable[%v] hpa-resource-spec[%v]", hpa.MetaData.Name, hpa.Spec.Allocatable, hpa.Spec)
			log.Error(":: validateDependents .. JSON validation failed ::", log.Fields{"Error": err})
			http.Error(*w, err.Error(), http.StatusBadRequest)
			return false
		}

		if (hpa.Spec.Resource.AllocatableResources.Limits > 0) && (hpa.Spec.Resource.AllocatableResources.Requests > hpa.Spec.Resource.AllocatableResources.Limits) {
			err := fmt.Errorf("JsonSchemaValidation: Document Validation failed .. resource-spec requests must be less or equal to limits. hpa-resource-name[%v] hpa-resource-spec[%v]", hpa.MetaData.Name, hpa.Spec)
			log.Error(":: validateDependents .. JSON validation failed ::", log.Fields{"Error": err})
			http.Error(*w, err.Error(), http.StatusBadRequest)
			return false
		}
	}

	return true
}
