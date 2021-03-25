package api

import (
	"encoding/json"
	"github.com/gorilla/mux"
	"io"
	"net/http"
	moduleLib "github.com/open-ness/EMCO/src/orchestrator/pkg/module"
	log "github.com/open-ness/EMCO/src/orchestrator/pkg/infra/logutils"
	"github.com/open-ness/EMCO/src/orchestrator/pkg/infra/validation"
	utils "github.com/open-ness/EMCO/src/orchestrator/utils"
)

var migrateJSONFile string = "json-schemas/migrate.json"
var rollbackJSONFile string = "json-schemas/rollback.json"

/* Used to store backend implementation objects
Also simplifies mocking for unit testing purposes
*/
type updateHandler struct {
	client moduleLib.InstantiationManager
}

func (h updateHandler) migrateHandler(w http.ResponseWriter, r *http.Request) {
	var migrate moduleLib.MigrateJson

	vars := mux.Vars(r)
	p := vars["project-name"]
	ca := vars["composite-app-name"]
	v := vars["composite-app-version"]
	di := vars["deployment-intent-group-name"]

	err := json.NewDecoder(r.Body).Decode(&migrate)
	log.Info("migrateJson:", log.Fields{"json:": migrate})
	switch {
	case err == io.EOF:
		log.Error(err.Error(), log.Fields{})
		http.Error(w, "Empty body", http.StatusBadRequest)
		w.WriteHeader(http.StatusBadRequest)
		return
	case err != nil:
		log.Error(err.Error(), log.Fields{})
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		w.WriteHeader(http.StatusUnprocessableEntity)
		return
	}

	// Verify JSON Body
	err, httpError := validation.ValidateJsonSchemaData(migrateJSONFile, migrate)
	if err != nil {
		log.Error(err.Error(), log.Fields{})
		http.Error(w, err.Error(), httpError)
		w.WriteHeader(httpError)
		return
	}

	tCav := migrate.Spec.TargetCompositeAppVersion
	tDig := migrate.Spec.TargetDigName

	log.Info("targetDeploymentName and targetCompositeAppVersion", log.Fields{"targetDeploymentName": tDig, "targetCompositeAppVersion": tCav})
	iErr := h.client.Migrate(p, ca, v, tCav, di, tDig)
	if iErr != nil {
		log.Error(iErr.Error(), log.Fields{})
		utils.HandleLogicalCloudError(iErr.Error(), &w)
		return
	}
	log.Info("migrateHandler ... end ", log.Fields{"project": p, "composite-app": ca, "composite-app-ver": v,
		"targetCompositeAppVersion": tCav, "dep-group": di, "targetDigName": tDig, "return-value": iErr})
	w.WriteHeader(http.StatusAccepted)
}

func (h updateHandler) updateHandler(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)
	p := vars["project-name"]
	ca := vars["composite-app-name"]
	v := vars["composite-app-version"]
	di := vars["deployment-intent-group-name"]

	revisionID, iErr := h.client.Update(p, ca, v, di)
	if iErr != nil {
		log.Error(iErr.Error(), log.Fields{})
		utils.HandleLogicalCloudError(iErr.Error(), &w)
		return
	}
	log.Info("updateHandler ... end ", log.Fields{"project": p, "composite-app": ca, "composite-app-ver": v,
		"dep-group": di, "return-value": iErr})
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	err := json.NewEncoder(w).Encode(revisionID)
	if err != nil {
		log.Error(err.Error(), log.Fields{})
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}


func (h updateHandler) rollbackHandler(w http.ResponseWriter, r *http.Request) {
	var rollback moduleLib.RollbackJson

	vars := mux.Vars(r)
	p := vars["project-name"]
	ca := vars["composite-app-name"]
	v := vars["composite-app-version"]
	di := vars["deployment-intent-group-name"]


	err := json.NewDecoder(r.Body).Decode(&rollback)
	log.Info("rollbackJson:", log.Fields{"json:": rollback})
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
	err, httpError := validation.ValidateJsonSchemaData(rollbackJSONFile, rollback)
	if err != nil {
		log.Error(err.Error(), log.Fields{})
		http.Error(w, err.Error(), httpError)
		return
	}

	rbRev := rollback.Spec.Revison

	iErr := h.client.Rollback(p, ca, v, di, rbRev)
	if iErr != nil {
		log.Error(iErr.Error(), log.Fields{})
		utils.HandleLogicalCloudError(iErr.Error(), &w)
		return
	}
	log.Info("rollbackHandler ... end ", log.Fields{"project": p, "composite-app": ca, "composite-app-ver": v,
		"dep-group": di, "revision": rbRev, "return-value": iErr})
	w.WriteHeader(http.StatusAccepted)

}