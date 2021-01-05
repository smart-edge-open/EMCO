// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2020 Intel Corporation

package api

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"io"
	"io/ioutil"
	"mime"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"strings"

	log "github.com/open-ness/EMCO/src/orchestrator/pkg/infra/logutils"
	"github.com/open-ness/EMCO/src/orchestrator/pkg/infra/validation"
	moduleLib "github.com/open-ness/EMCO/src/orchestrator/pkg/module"

	"github.com/gorilla/mux"
	pkgerrors "github.com/pkg/errors"
)

var appProfileJSONFile string = "json-schemas/metadata.json"

/*maxMemory - ParseMultipartForm method used in the multipart form handling, parses a request body as multipart/form-data. 
The whole request body is parsed and up to a total of maxMemory bytes of its file parts are stored in memory, with the remainder stored on disk in temporary files.
*/
var maxMemory int64 = 16777216
var oneGB int64 = 1073741824

/* Used to store backend implementation objects
Also simplifies mocking for unit testing purposes
*/
type appProfileHandler struct {
	client moduleLib.AppProfileManager
}

// createAppProfileHandler handles the create operation
func (h appProfileHandler) createAppProfileHandler(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)
	project := vars["project-name"]
	compositeApp := vars["composite-app-name"]
	compositeAppVersion := vars["composite-app-version"]
	compositeProfile := vars["composite-profile-name"]

	var ap moduleLib.AppProfile
	var ac moduleLib.AppProfileContent

	// Implemenation using multipart form
	// Review and enable/remove at a later date
	// Set Max size to 16mb here
	err := r.ParseMultipartForm(maxMemory)
	if err != nil {
		log.Error(err.Error(), log.Fields{})
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	jsn := bytes.NewBuffer([]byte(r.FormValue("metadata")))
	err = json.NewDecoder(jsn).Decode(&ap)
	switch {
	case err == io.EOF:
		log.Error(err.Error(), log.Fields{})
		http.Error(w, "Empty body", http.StatusBadRequest)
		return
	case err != nil:
		log.Error(err.Error(), log.Fields{})
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	// Verify JSON Body
	err, httpError := validation.ValidateJsonSchemaData(appProfileJSONFile, ap)
	if err != nil {
		log.Error(err.Error(), log.Fields{})
		http.Error(w, err.Error(), httpError)
		return
	}
	//Read the file section and ignore the header
	file, _, err := r.FormFile("file")
	
	if err != nil {
		log.Error(err.Error(), log.Fields{})
		http.Error(w, "Unable to process file", http.StatusUnprocessableEntity)
		return
	}

	defer file.Close()

	//Convert the file content to base64 for storage
	content, err := ioutil.ReadAll(file)
	if err != nil {
		log.Error(err.Error(), log.Fields{})
		http.Error(w, "Unable to read file", http.StatusUnprocessableEntity)
		return
	}
	// Limit file Size to 1 GB
	if len(content) > int(oneGB) {
		log.Error("File Size Exceeds 1 GB", log.Fields{})
		http.Error(w, "File Size Exceeds 1 GB", http.StatusUnprocessableEntity)
		return
	}
	err = validation.IsTarGz(bytes.NewBuffer(content))
	if err != nil {
		log.Error(err.Error(), log.Fields{})
		http.Error(w, "Error in file format", http.StatusUnprocessableEntity)
		return
	}

	ac.Profile = base64.StdEncoding.EncodeToString(content)

	// Name is required.
	if ap.Metadata.Name == "" {
		log.Error("Missing name in POST request", log.Fields{})
		http.Error(w, "Missing name in POST request", http.StatusBadRequest)
		return
	}

	ret, createErr := h.client.CreateAppProfile(project, compositeApp, compositeAppVersion, compositeProfile, ap, ac, false)
	if createErr != nil {
		log.Error(createErr.Error(), log.Fields{})
		if strings.Contains(createErr.Error(), "Unable to find the compositeProfile") {
			http.Error(w, createErr.Error(), http.StatusNotFound)
		} else if strings.Contains(createErr.Error(), "App already has an AppProfile") {
			http.Error(w, createErr.Error(), http.StatusConflict)
		} else if strings.Contains(createErr.Error(), "AppProfile already exists") {
			http.Error(w, createErr.Error(), http.StatusConflict)
		} else {
			http.Error(w, createErr.Error(), http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusCreated)
	err = json.NewEncoder(w).Encode(ret)
	if err != nil {
		log.Error(err.Error(), log.Fields{})
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// getHandler handles the GET operations on AppProfile
func (h appProfileHandler) getAppProfileHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	project := vars["project-name"]
	compositeApp := vars["composite-app-name"]
	compositeAppVersion := vars["composite-app-version"]
	compositeProfile := vars["composite-profile-name"]
	name := vars["app-profile"]
	appName := r.URL.Query().Get("app-name")

	if len(name) != 0 && len(appName) != 0 {
		log.Error("Invalid query - contains both app-profile and app-name", log.Fields{})
		http.Error(w, pkgerrors.New("Invalid query - contains both app-profile and app-name").Error(), http.StatusInternalServerError)
		return
	}

	// handle the get all app profiles case - return a list of only the json parts
	if len(name) == 0 && len(appName) == 0 {
		var retList []moduleLib.AppProfile

		ret, err := h.client.GetAppProfiles(project, compositeApp, compositeAppVersion, compositeProfile)
		if err != nil {
			log.Error(err.Error(), log.Fields{})
			if strings.Contains(err.Error(), "db Find error") {
				http.Error(w, err.Error(), http.StatusNotFound)
			} else if strings.Contains(err.Error(), "not found") {
				http.Error(w, err.Error(), http.StatusNotFound)
			} else {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			return
		}

		for _, ap := range ret {
			retList = append(retList, ap)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		err = json.NewEncoder(w).Encode(retList)
		if err != nil {
			log.Error(err.Error(), log.Fields{})
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		return
	}

	accepted, _, err := mime.ParseMediaType(r.Header.Get("Accept"))
	if err != nil {
		log.Error(err.Error(), log.Fields{})
		http.Error(w, err.Error(), http.StatusNotAcceptable)
		return
	}

	var retAppProfile moduleLib.AppProfile
	var retAppProfileContent moduleLib.AppProfileContent

	if len(appName) != 0 {
		retAppProfile, err = h.client.GetAppProfileByApp(project, compositeApp, compositeAppVersion, compositeProfile, appName)
		if err != nil {
			log.Error(err.Error(), log.Fields{})
			if strings.Contains(err.Error(), "db Find error") {
				http.Error(w, err.Error(), http.StatusNotFound)
			} else if strings.Contains(err.Error(), "not found") {
				http.Error(w, err.Error(), http.StatusNotFound)
			} else {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			return
		}

		retAppProfileContent, err = h.client.GetAppProfileContentByApp(project, compositeApp, compositeAppVersion, compositeProfile, appName)
		if err != nil {
			log.Error(err.Error(), log.Fields{})
			if strings.Contains(err.Error(), "db Find error") {
				http.Error(w, err.Error(), http.StatusNotFound)
			} else if strings.Contains(err.Error(), "not found") {
				http.Error(w, err.Error(), http.StatusNotFound)
			} else {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			return
		}
	} else {
		retAppProfile, err = h.client.GetAppProfile(project, compositeApp, compositeAppVersion, compositeProfile, name)
		if err != nil {
			log.Error(err.Error(), log.Fields{})
			if strings.Contains(err.Error(), "db Find error") {
				http.Error(w, err.Error(), http.StatusNotFound)
			} else if strings.Contains(err.Error(), "not found") {
				http.Error(w, err.Error(), http.StatusNotFound)
			} else {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			return
		}

		retAppProfileContent, err = h.client.GetAppProfileContent(project, compositeApp, compositeAppVersion, compositeProfile, name)
		if err != nil {
			log.Error(err.Error(), log.Fields{})
			if strings.Contains(err.Error(), "db Find error") {
				http.Error(w, err.Error(), http.StatusNotFound)
			} else {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			return
		}
	}

	switch accepted {
	case "multipart/form-data":
		mpw := multipart.NewWriter(w)
		w.Header().Set("Content-Type", mpw.FormDataContentType())
		w.WriteHeader(http.StatusOK)
		pw, err := mpw.CreatePart(textproto.MIMEHeader{"Content-Type": {"application/json"}, "Content-Disposition": {"form-data; name=metadata"}})
		if err != nil {
			log.Error(err.Error(), log.Fields{})
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if err := json.NewEncoder(pw).Encode(retAppProfile); err != nil {
			log.Error(err.Error(), log.Fields{})
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		pw, err = mpw.CreatePart(textproto.MIMEHeader{"Content-Type": {"application/octet-stream"}, "Content-Disposition": {"form-data; name=file"}})
		if err != nil {
			log.Error(err.Error(), log.Fields{})
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		kcBytes, err := base64.StdEncoding.DecodeString(retAppProfileContent.Profile)
		if err != nil {
			log.Error(err.Error(), log.Fields{})
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		_, err = pw.Write(kcBytes)
		if err != nil {
			log.Error(err.Error(), log.Fields{})
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	case "application/json":
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		err = json.NewEncoder(w).Encode(retAppProfile)
		if err != nil {
			log.Error(err.Error(), log.Fields{})
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	case "application/octet-stream":
		w.Header().Set("Content-Type", "application/octet-stream")
		w.WriteHeader(http.StatusOK)
		kcBytes, err := base64.StdEncoding.DecodeString(retAppProfileContent.Profile)
		if err != nil {
			log.Error(err.Error(), log.Fields{})
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		_, err = w.Write(kcBytes)
		if err != nil {
			log.Error(err.Error(), log.Fields{})
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	default:
		log.Error("HEADER missing::set Accept: multipart/form-data, application/json or application/octet-stream", log.Fields{})
		http.Error(w, "set Accept: multipart/form-data, application/json or application/octet-stream", http.StatusMultipleChoices)
		return
	}
}

// deleteHandler handles the delete operations on AppProfile
func (h appProfileHandler) deleteAppProfileHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	project := vars["project-name"]
	compositeApp := vars["composite-app-name"]
	compositeAppVersion := vars["composite-app-version"]
	compositeProfile := vars["composite-profile-name"]
	name := vars["app-profile"]

	err := h.client.DeleteAppProfile(project, compositeApp, compositeAppVersion, compositeProfile, name)
	if err != nil {
		log.Error(err.Error(), log.Fields{})
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

func (h appProfileHandler) updateAppProfileHandler(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)
	project := vars["project-name"]
	compositeApp := vars["composite-app-name"]
	compositeAppVersion := vars["composite-app-version"]
	compositeProfile := vars["composite-profile-name"]

	var ap moduleLib.AppProfile
	var ac moduleLib.AppProfileContent

	// Implemenation using multipart form
	// Review and enable/remove at a later date
	// Set Max size to 16mb here
	err := r.ParseMultipartForm(maxMemory)
	if err != nil {
		log.Error(err.Error(), log.Fields{})
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	jsn := bytes.NewBuffer([]byte(r.FormValue("metadata")))
	err = json.NewDecoder(jsn).Decode(&ap)
	switch {
	case err == io.EOF:
		log.Error(err.Error(), log.Fields{})
		http.Error(w, "Empty body", http.StatusBadRequest)
		return
	case err != nil:
		log.Error(err.Error(), log.Fields{})
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	// Verify JSON Body
	err, httpError := validation.ValidateJsonSchemaData(appProfileJSONFile, ap)
	if err != nil {
		log.Error(err.Error(), log.Fields{})
		http.Error(w, err.Error(), httpError)
		return
	}
	//Read the file section and ignore the header
	file, _, err := r.FormFile("file")
	if err != nil {
		log.Error(err.Error(), log.Fields{})
		http.Error(w, "Unable to process file", http.StatusUnprocessableEntity)
		return
	}

	defer file.Close()

	//Convert the file content to base64 for storage
	content, err := ioutil.ReadAll(file)
	if err != nil {
		log.Error(err.Error(), log.Fields{})
		http.Error(w, "Unable to read file", http.StatusUnprocessableEntity)
		return
	}
	// Limit file Size to 1 GB
	if len(content) > int(oneGB) {
		log.Error("File Size Exceeds 1 GB", log.Fields{})
		http.Error(w, "File Size Exceeds 1 GB", http.StatusUnprocessableEntity)
		return
	}
	err = validation.IsTarGz(bytes.NewBuffer(content))
	if err != nil {
		log.Error(err.Error(), log.Fields{})
		http.Error(w, "Error in file format", http.StatusUnprocessableEntity)
		return
	}

	ac.Profile = base64.StdEncoding.EncodeToString(content)

	// Name is required.
	if ap.Metadata.Name == "" {
		log.Error("Missing name in POST request", log.Fields{})
		http.Error(w, "Missing name in POST request", http.StatusBadRequest)
		return
	}

	ret, createErr := h.client.CreateAppProfile(project, compositeApp, compositeAppVersion, compositeProfile, ap, ac, true)
	if createErr != nil {
		log.Error(createErr.Error(), log.Fields{})
		if strings.Contains(createErr.Error(), "Unable to find the compositeProfile") {
			http.Error(w, createErr.Error(), http.StatusNotFound)
		} else if strings.Contains(createErr.Error(), "App already has an AppProfile") {
			http.Error(w, createErr.Error(), http.StatusConflict)
		} else {
			http.Error(w, createErr.Error(), http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusCreated)
	err = json.NewEncoder(w).Encode(ret)
	if err != nil {
		log.Error(err.Error(), log.Fields{})
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}