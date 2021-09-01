// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2020 Intel Corporation

package api

import (
	"encoding/json"
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
addHpaIntentHandler handles the URL
URL: /v2/projects/{project-name}/composite-apps/{composite-app-name}/{version}/
deployment-intent-groups/{deployment-intent-group-name}/hpa-intents
*/
// Add Hpa Intent in Deployment Group
func (h HpaPlacementIntentHandler) addHpaIntentHandler(w http.ResponseWriter, r *http.Request) {
	var hpa hpaModel.DeploymentHpaIntent
	reqDump, _ := httputil.DumpRequest(r, true)
	log.Info(":: addHpaIntentHandler .. start ::", log.Fields{"req": string(reqDump)})

	err := json.NewDecoder(r.Body).Decode(&hpa)
	switch {
	case err == io.EOF:
		log.Error(":: addHpaIntentHandler .. Empty POST body ::", log.Fields{"Error": err})
		http.Error(w, "addHpaIntentHandler .. Empty POST body", http.StatusBadRequest)
		return
	case err != nil:
		log.Error(":: addHpaIntentHandler .. Error decoding POST body ::", log.Fields{"Error": err})
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	// Validate json schema
	err, httpError := validation.ValidateJsonSchemaData(hpaIntentJSONFile, hpa)
	if err != nil {
		log.Error(":: addHpaIntentHandler .. JSON validation failed ::", log.Fields{"Error": err})
		http.Error(w, err.Error(), httpError)
		return
	}

	vars := mux.Vars(r)
	p := vars["project-name"]
	ca := vars["composite-app-name"]
	v := vars["composite-app-version"]
	d := vars["deployment-intent-group-name"]

	log.Info(":: addHpaIntentHandler .. Req ::", log.Fields{"project": p, "composite-app": ca, "composite-app-ver": v, "dep-group": d, "intent-name": hpa.MetaData.Name})
	intent, err := h.client.AddIntent(hpa, p, ca, v, d, false)
	if err != nil {
		log.Error(":: addHpaIntentHandler .. AddIntent error ::", log.Fields{"Error": err})
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
	err = json.NewEncoder(w).Encode(intent)
	if err != nil {
		log.Error(":: addHpaIntentHandler ..  Encoder error ::", log.Fields{"Error": err})
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	log.Info(":: addHpaIntentHandler .. end ::", log.Fields{"intent": intent})
}

/*
getIntentByNameHandler handles the URL
URL: /v2/projects/{project-name}/composite-apps/{composite-app-name}/{version}/
deployment-intent-groups/{deployment-intent-group-name}?intent=<intent>
*/
// Query Hpa Intent in Deployment Group
func (h HpaPlacementIntentHandler) getHpaIntentByNameHandler(w http.ResponseWriter, r *http.Request) {
	reqDump, _ := httputil.DumpRequest(r, true)
	log.Info(":: getHpaIntentByNameHandler .. start ::", log.Fields{"req": string(reqDump)})

	p, ca, v, di, _, err := parseHpaIntentReqParameters(&w, r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	iN := r.URL.Query().Get("intent")
	if iN == "" {
		log.Error(":: getHpaIntentByNameHandler .. Missing intent-name in request ::", log.Fields{"Error": http.StatusBadRequest})
		http.Error(w, "getHpaIntentByNameHandler .. Missing intent-name in request", http.StatusBadRequest)
		return
	}

	intent, err := h.client.GetIntentByName(iN, p, ca, v, di)
	if err != nil {
		log.Error(":: getHpaIntentByNameHandler .. GetIntentByName error ::", log.Fields{"intent-name": iN, "Error": err})
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(intent)
	if err != nil {
		log.Error(":: getHpaIntentByNameHandler .. Encoder error ::", log.Fields{"Error": err})
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	log.Info(":: getHpaIntentByNameHandler .. end ::", log.Fields{"intent": intent})
}

/*
getHpaIntentHandler/getHpaIntentHandlers handles the URL
URL: /v2/projects/{project-name}/composite-apps/{composite-app-name}/{version}/
deployment-intent-groups/{deployment-intent-group-name}/hpa-intents/{intent-name}
*/
// Get Hpa Intent in Deployment Group
func (h HpaPlacementIntentHandler) getHpaIntentHandler(w http.ResponseWriter, r *http.Request) {
	reqDump, _ := httputil.DumpRequest(r, true)
	log.Info(":: getHpaIntentHandler .. start ::", log.Fields{"req": string(reqDump)})
	p, ca, v, di, name, err := parseHpaIntentReqParameters(&w, r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	log.Info(":: getHpaIntentHandler .. Req ::", log.Fields{"project": p, "composite-app": ca, "composite-app-ver": v, "dep-group": di, "intent-name": name})

	var intents interface{}
	if len(name) == 0 {
		intents, err = h.client.GetAllIntents(p, ca, v, di)
		if err != nil {
			log.Error(":: getHpaIntentHandler .. GetAllInents failure ::", log.Fields{"Error": err})
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
	} else {
		intents, _, err = h.client.GetIntent(name, p, ca, v, di)
		if err != nil {
			log.Error(":: getHpaIntentHandler .. GetIntent failure ::", log.Fields{"Error": err})
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(intents)
	if err != nil {
		log.Error(":: getHpaIntentHandler .. Encoder failure ::", log.Fields{"Error": err})
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	log.Info(":: getHpaIntentHandler .. end ::", log.Fields{"intents": intents})
}

/*
putHpaPlacementIntentHandler handles the URL
URL: /v2/projects/{project-name}/composite-apps/{composite-app-name}/{version}/
deployment-intent-groups/{deployment-intent-group-name}/hpa-intents/{intent-name}
*/
// Update Hpa Intent in Deployment Group
func (h HpaPlacementIntentHandler) putHpaIntentHandler(w http.ResponseWriter, r *http.Request) {
	var hpa hpaModel.DeploymentHpaIntent
	reqDump, _ := httputil.DumpRequest(r, true)
	log.Info(":: putHpaIntentHandler .. start ::", log.Fields{"req": string(reqDump)})

	err := json.NewDecoder(r.Body).Decode(&hpa)
	switch {
	case err == io.EOF:
		log.Error(":: putHpaIntentHandler .. Empty PUT body ::", log.Fields{"Error": err})
		http.Error(w, "Empty body", http.StatusBadRequest)
		return
	case err != nil:
		log.Error(":: putHpaIntentHandler .. decoding resource PUT body ::", log.Fields{"Error": err})
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	// Validate json schema
	err, httpError := validation.ValidateJsonSchemaData(hpaIntentJSONFile, hpa)
	if err != nil {
		log.Error(":: putHpaIntentHandler .. JSON validation failed ::", log.Fields{"Error": err})
		http.Error(w, err.Error(), httpError)
		return
	}

	// Parse request parameters
	p, ca, v, di, name, err := parseHpaIntentReqParameters(&w, r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Name in URL should match name in body
	if hpa.MetaData.Name != name {
		log.Error(":: putHpaIntentHandler .. Mismatched name in PUT request ::", log.Fields{"bodyname": hpa.MetaData.Name, "name": name})
		http.Error(w, "putHpaIntentHandler .. Mismatched name in PUT request", http.StatusBadRequest)
		return
	}

	log.Info(":: putHpaIntentHandler .. Req ::", log.Fields{"project": p, "composite-app": ca, "composite-app-ver": v, "dep-group": di, "intent-name": name})
	intent, err := h.client.AddIntent(hpa, p, ca, v, di, true)
	if err != nil {
		log.Error(":: putHpaIntentHandler .. AddIntent error ::", log.Fields{"Error": err})
		if strings.Contains(err.Error(), "not found") {
			http.Error(w, err.Error(), http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(intent)
	if err != nil {
		log.Error(":: putHpaIntentHandler .. encoding failure ::", log.Fields{"Error": err})
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	log.Info(":: putHpaIntentHandler .. end ::", log.Fields{"intent": intent})
}

/*
deleteHpaIntentHandler handles the URL
URL: /v2/projects/{project-name}/composite-apps/{composite-app-name}/{version}/
deployment-intent-groups/{deployment-intent-group-name}/hpa-intents/{intent-name}
*/
// Delete Hpa Intent in Deployment Group
func (h HpaPlacementIntentHandler) deleteHpaIntentHandler(w http.ResponseWriter, r *http.Request) {
	reqDump, _ := httputil.DumpRequest(r, true)
	log.Info(":: deleteHpaIntentHandler .. start ::", log.Fields{"req": string(reqDump)})

	// Parse request parameters
	p, ca, v, di, name, err := parseHpaIntentReqParameters(&w, r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	log.Info(":: deleteHpaIntentHandler .. Req ::", log.Fields{"project": p, "composite-app": ca, "composite-app-ver": v, "dep-group": di, "intent-name": name})

	_, _, err = h.client.GetIntent(name, p, ca, v, di)
	if err != nil {
		log.Error(":: deleteHpaIntentHandler .. GetIntent failure ::", log.Fields{"Error": err})
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	err = h.client.DeleteIntent(name, p, ca, v, di)
	if err != nil {
		log.Error(":: deleteHpaIntentHandler .. DeleteIntent failure ::", log.Fields{"Error": err})
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	w.WriteHeader(http.StatusNoContent)
	log.Info(":: deleteHpaIntentHandler .. end ::", log.Fields{"req": string(reqDump)})
}

/*
deleteAllHpaIntentsHandler handles the URL
URL: /v2/projects/{project-name}/composite-apps/{composite-app-name}/{version}/
deployment-intent-groups/{deployment-intent-group-name}/hpa-intents
*/
// Delete all Hpa Intents in Deployment Group
func (h HpaPlacementIntentHandler) deleteAllHpaIntentsHandler(w http.ResponseWriter, r *http.Request) {
	reqDump, _ := httputil.DumpRequest(r, true)
	log.Info(":: deleteAllHpaIntentsHandler .. start ::", log.Fields{"req": string(reqDump)})

	// Parse request parameters
	p, ca, v, di, name, err := parseHpaIntentReqParameters(&w, r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	log.Info(":: deleteAllHpaIntentsHandler .. Req ::", log.Fields{"project": p, "composite-app": ca, "composite-app-ver": v, "dep-group": di, "intent-name": name})

	hpaintents, err := h.client.GetAllIntents(p, ca, v, di)
	if err != nil {
		log.Error(":: deleteAllHpaIntentsHandler .. GetAllIntents failure ::", log.Fields{"Error": err})
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	for _, hpaIntent := range hpaintents {
		err = h.client.DeleteIntent(hpaIntent.MetaData.Name, p, ca, v, di)
		if err != nil {
			log.Error(":: deleteAllHpaIntentsHandler .. DeleteIntent failure ::", log.Fields{"intent-name": hpaIntent.MetaData.Name, "Error": err})
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
	}

	w.WriteHeader(http.StatusNoContent)
	log.Info(":: deleteAllHpaIntentsHandler .. end ::", log.Fields{"req": string(reqDump)})
}

/* Parse Http request Parameters */
func parseHpaIntentReqParameters(w *http.ResponseWriter, r *http.Request) (string, string, string, string, string, error) {
	vars := mux.Vars(r)

	i := vars["intent-name"]

	p := vars["project-name"]
	if p == "" {
		log.Error(":: parseHpaIntentReqParameters ..  Missing projectName in request ::", log.Fields{"Error": http.StatusBadRequest})
		http.Error(*w, "parseHpaIntentReqParameters .. Missing projectName in request", http.StatusBadRequest)
		return "", "", "", "", "", pkgerrors.New("Missing project-name")
	}
	ca := vars["composite-app-name"]
	if ca == "" {
		log.Error(":: parseHpaIntentReqParameters ..  Missing compositeAppName in request ::", log.Fields{"Error": http.StatusBadRequest})
		http.Error(*w, "parseHpaIntentReqParameters .. Missing compositeAppName in request", http.StatusBadRequest)
		return "", "", "", "", "", pkgerrors.New("Missing composite-app-name")
	}

	v := vars["composite-app-version"]
	if v == "" {
		log.Error(":: parseHpaIntentReqParameters ..  version intentName in request ::", log.Fields{"Error": http.StatusBadRequest})
		http.Error(*w, "parseHpaIntentReqParameters .. Missing version of compositeApp in request", http.StatusBadRequest)
		return "", "", "", "", "", pkgerrors.New("Missing composite-app-version")
	}

	di := vars["deployment-intent-group-name"]
	if di == "" {
		log.Error(":: parseHpaIntentReqParameters ..  Missing DeploymentIntentGroup in request ::", log.Fields{"Error": http.StatusBadRequest})
		http.Error(*w, "parseHpaIntentReqParameters .. Missing name of DeploymentIntentGroup in request", http.StatusBadRequest)
		return "", "", "", "", "", pkgerrors.New("Missing deployment-intent-group-name")
	}

	return p, ca, v, di, i, nil
}
