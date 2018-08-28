package api

// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2020 Intel Corporation

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"io"
	"io/ioutil"
	"mime"
	"mime/multipart"
	"strings"

	//"mime"
	//"mime/multipart"
	"net/http"
	"net/textproto"

	//"net/textproto"
	//"strings"

	"github.com/gorilla/mux"
	moduleLib "github.com/open-ness/EMCO/src/genericactioncontroller/pkg/module"
	"github.com/open-ness/EMCO/src/orchestrator/pkg/infra/logutils"
	log "github.com/open-ness/EMCO/src/orchestrator/pkg/infra/logutils"
	"github.com/open-ness/EMCO/src/orchestrator/pkg/infra/validation"
)

var czJSONFile string = "json-schemas/customization.json"

type customizationHandler struct {
	// Interface that implements resource operations
	// We will set this variable with a mock interface for testing
	client moduleLib.CustomizationManager
}

func (ch customizationHandler) createCustomizationHandler(w http.ResponseWriter, r *http.Request) {
	var cz moduleLib.Customization
	//var cc moduleLib.SpecFileContent
	var contentArray []string
	var fileNameArray []string

	vars := mux.Vars(r)

	err := r.ParseMultipartForm(16777216)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	jsn := bytes.NewBuffer([]byte(r.FormValue("metadata")))
	err = json.NewDecoder(jsn).Decode(&cz)
	switch {
	case err == io.EOF:
		log.Error(":: Empty customization POST body ::", log.Fields{"Error": err})
		http.Error(w, "Empty body", http.StatusBadRequest)
		return
	case err != nil:
		log.Error(":: Error decoding resource POST body ::", log.Fields{"Error": err})
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	err, httpError := validation.ValidateJsonSchemaData(czJSONFile, cz)
	if err != nil {
		log.Error(":: JSON validation failed ::", log.Fields{"Error": err})
		http.Error(w, err.Error(), httpError)
		return
	}

	if strings.ToLower(cz.Spec.ClusterSpecific) == "true" && (moduleLib.ClusterInfo{}) == cz.Spec.ClusterInfo {
		log.Error(":: ClusterInfo missing when ClusterSpecific is true ::", log.Fields{})
		http.Error(w, "ClusterInfo missing", httpError)
		return
	}

	if strings.ToLower(cz.Spec.ClusterSpecific) == "true" && strings.ToLower(cz.Spec.ClusterInfo.Scope) == "label" && cz.Spec.ClusterInfo.ClusterLabel == "" {
		log.Error(":: ClusterLabel missing when ClusterSpecific is true and  ClusterScope is label::", log.Fields{})
		http.Error(w, "ClusterLabel missing", httpError)
		return
	}

	if strings.ToLower(cz.Spec.ClusterSpecific) == "true" && strings.ToLower(cz.Spec.ClusterInfo.Scope) == "name" && cz.Spec.ClusterInfo.ClusterName == "" {
		log.Error(":: ClusterName missing when ClusterSpecific is true and  ClusterScope is name::", log.Fields{})
		http.Error(w, "ClusterName missing", httpError)
		return
	}

	// BEGIN: Customization file processing
	formData := r.MultipartForm

	//get the *fileheaders
	files := formData.File["files"]

	for i := range files {
		file, err := files[i].Open()
		defer file.Close()
		if err != nil {
			logutils.Info("Unable to open file", log.Fields{"FileName": files[i].Filename})
			http.Error(w, err.Error(), http.StatusUnprocessableEntity)
			return
		}
		content, err := ioutil.ReadAll(file)
		if err != nil {
			log.Error(":: File read failed ::", log.Fields{"Error": err})
			http.Error(w, "Unable to read file", http.StatusUnprocessableEntity)
			return
		}
		contStr := base64.StdEncoding.EncodeToString(content)
		contentArray = append(contentArray, contStr)
		fileNameArray = append(fileNameArray, files[i].Filename)
		logutils.Info("Appended file", log.Fields{"FileName": files[i].Filename})

	}

	specFC := moduleLib.SpecFileContent{FileContents: contentArray, FileNames: fileNameArray}
	// END: Customization file processing

	if cz.Metadata.Name == "" {
		log.Error(":: Missing name in POST request ::", log.Fields{"Error": err})
		http.Error(w, "Missing name in POST request", http.StatusBadRequest)
		return
	}

	p := vars["project"]
	ca := vars["composite-app-name"]
	cv := vars["version"]
	dig := vars["deployment-intent-group-name"]
	gki := vars["intent-name"]
	rs := vars["resource-name"]

	ret, err := ch.client.CreateCustomization(cz, specFC, p, ca, cv, dig, gki, rs, false)
	if err != nil {
		log.Error(":: Create customization failure::", log.Fields{"Error": err})
		if strings.Contains(err.Error(), "Customization already exists") {
			http.Error(w, err.Error(), http.StatusConflict)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	err = json.NewEncoder(w).Encode(ret)
	if err != nil {
		log.Error(":: Encoding error ::", log.Fields{"Error": err})
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

}

func (ch customizationHandler) putCustomizationHandler(w http.ResponseWriter, r *http.Request) {
	var cz moduleLib.Customization
	//var cc moduleLib.SpecFileContent
	var contentArray []string

	vars := mux.Vars(r)

	err := r.ParseMultipartForm(16777216)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	jsn := bytes.NewBuffer([]byte(r.FormValue("metadata")))
	err = json.NewDecoder(jsn).Decode(&cz)
	switch {
	case err == io.EOF:
		log.Error(":: Empty customization POST body ::", log.Fields{"Error": err})
		http.Error(w, "Empty body", http.StatusBadRequest)
		return
	case err != nil:
		log.Error(":: Error decoding resource POST body ::", log.Fields{"Error": err})
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	err, httpError := validation.ValidateJsonSchemaData(czJSONFile, cz)
	if err != nil {
		log.Error(":: JSON validation failed ::", log.Fields{"Error": err})
		http.Error(w, err.Error(), httpError)
		return
	}

	// BEGIN: Customization file processing
	formData := r.MultipartForm

	//get the *fileheaders
	files := formData.File["files"]

	for i := range files {
		file, err := files[i].Open()
		defer file.Close()
		if err != nil {
			logutils.Info("Unable to open file", log.Fields{"FileName": files[i].Filename})
			http.Error(w, err.Error(), http.StatusUnprocessableEntity)
			return
		}
		content, err := ioutil.ReadAll(file)
		if err != nil {
			log.Error(":: File read failed ::", log.Fields{"Error": err})
			http.Error(w, "Unable to read file", http.StatusUnprocessableEntity)
			return
		}
		contStr := base64.StdEncoding.EncodeToString(content)
		contentArray = append(contentArray, contStr)
		logutils.Info("Appended file", log.Fields{"FileName": files[i].Filename})

	}

	specFC := moduleLib.SpecFileContent{FileContents: contentArray}
	// END: Customization file processing

	if cz.Metadata.Name == "" {
		log.Error(":: Missing name in PUT request ::", log.Fields{"Error": err})
		http.Error(w, "Missing name in PUT request", http.StatusBadRequest)
		return
	}

	p := vars["project"]
	ca := vars["composite-app-name"]
	cv := vars["version"]
	dig := vars["deployment-intent-group-name"]
	gki := vars["intent-name"]
	rs := vars["resource-name"]

	ret, err := ch.client.CreateCustomization(cz, specFC, p, ca, cv, dig, gki, rs, true)
	if err != nil {
		log.Error(":: Create customization failure while PUT::", log.Fields{"Error": err})
		if strings.Contains(err.Error(), "Customization already exists") {
			http.Error(w, err.Error(), http.StatusConflict)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	err = json.NewEncoder(w).Encode(ret)
	if err != nil {
		log.Error(":: Encoding error ::", log.Fields{"Error": err})
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

}

func (ch customizationHandler) getCustomizationHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	name := vars["name"]
	p := vars["project"]
	ca := vars["composite-app-name"]
	cv := vars["version"]
	dig := vars["deployment-intent-group-name"]
	gki := vars["intent-name"]
	rs := vars["resource-name"]

	if len(name) == 0 {

		var czList []moduleLib.Customization

		ret, err := ch.client.GetAllCustomization(p, ca, cv, dig, gki, rs)
		if err != nil {
			log.Error(":: GetAllCustomization failure::", log.Fields{"Error": err})
			if strings.Contains(err.Error(), "db Find error") {
				http.Error(w, err.Error(), http.StatusNotFound)
			} else {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			return
		}

		for _, cz := range ret {
			czList = append(czList, moduleLib.Customization{Metadata: cz.Metadata, Spec: cz.Spec})
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		err = json.NewEncoder(w).Encode(czList)
		if err != nil {
			log.Error(":: Encoding customization failure::", log.Fields{"Error": err})
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		return
	}

	accepted, _, err := mime.ParseMediaType(r.Header.Get("Accept"))
	if err != nil {
		log.Error(":: Mime parser failure::", log.Fields{"Error": err})
		http.Error(w, err.Error(), http.StatusNotAcceptable)
		return
	}

	var cz moduleLib.Customization
	var specFC moduleLib.SpecFileContent

	cz, err = ch.client.GetCustomization(name, p, ca, cv, dig, gki, rs)
	if err != nil {
		log.Error(":: GetCustomization failure::", log.Fields{"Error": err})
		if strings.Contains(err.Error(), "db Find error") {
			http.Error(w, err.Error(), http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	specFC, err = ch.client.GetCustomizationContent(name, p, ca, cv, dig, gki, rs)
	if err != nil {
		log.Error(":: GetCustomizationContent failure::", log.Fields{"Error": err})
		if strings.Contains(err.Error(), "db Find error") {
			http.Error(w, err.Error(), http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	switch accepted {
	case "multipart/form-data":
		mpw := multipart.NewWriter(w)
		w.Header().Set("Content-Type", mpw.FormDataContentType())
		w.WriteHeader(http.StatusOK)
		pw, err := mpw.CreatePart(textproto.MIMEHeader{"Content-Type": {"application/json"}, "Content-Disposition": {"form-data; name=customization"}})
		if err != nil {
			log.Error(":: multipart/form-data :: application/json failure::", log.Fields{"Error": err})
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if err := json.NewEncoder(pw).Encode(cz); err != nil {
			log.Error(":: multipart/form-data :: encoding failure", log.Fields{"Error": err})
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		pw, err = mpw.CreatePart(textproto.MIMEHeader{"Content-Type": {"application/octet-stream"}, "Content-Disposition": {"form-data; name=files; filename=customizationFile"}})
		if err != nil {
			log.Error(":: multipart/form-data :: application/octet-stream failure ::", log.Fields{"Error": err})
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		for _, content := range specFC.FileContents {
			brBytes, err := base64.StdEncoding.DecodeString(content)
			if err != nil {
				log.Error(":: multipart/form-data :: application/octet-stream decode failure ::", log.Fields{"Error": err})
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			_, err = pw.Write(brBytes)
			if err != nil {
				log.Error(":: FileWriter failure ::", log.Fields{"Error": err})
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		}

	case "application/json":
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		err = json.NewEncoder(w).Encode(cz)
		if err != nil {
			log.Error(":: application/json encoding failure::", log.Fields{"Error": err})
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

	case "application/octet-stream":
		w.Header().Set("Content-Type", "application/octet-stream")
		w.WriteHeader(http.StatusOK)

		for _, content := range specFC.FileContents {
			czBytes, err := base64.StdEncoding.DecodeString(content)
			if err != nil {
				log.Error(":: application/octet-stream failure::", log.Fields{"Error": err})
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			_, err = w.Write(czBytes)
			if err != nil {
				log.Error(":: FileWriter failure ::", log.Fields{"Error": err})
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		}

	default:
		http.Error(w, "set Accept: multipart/form-data, application/json or application/octet-stream", http.StatusMultipleChoices)
		return

	}
}

func (ch customizationHandler) deleteCustomizationHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	name := vars["name"]
	p := vars["project"]
	ca := vars["composite-app-name"]
	cv := vars["version"]
	dig := vars["deployment-intent-group-name"]
	gki := vars["intent-name"]
	rs := vars["resource-name"]

	err := ch.client.DeleteCustomization(name, p, ca, cv, dig, gki, rs)
	if err != nil {
		log.Error(":: DeleteCustomization failure ::", log.Fields{"Error": err})
		if strings.Contains(err.Error(), "not found") {
			http.Error(w, err.Error(), http.StatusNotFound)
		} else if strings.Contains(err.Error(), "conflict") {
			http.Error(w, err.Error(), http.StatusConflict)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
