// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2020 Intel Corporation

package context

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/open-ness/EMCO/src/orchestrator/pkg/appcontext"
	log "github.com/open-ness/EMCO/src/orchestrator/pkg/infra/logutils"
	"github.com/open-ness/EMCO/src/orchestrator/pkg/resourcestatus"
	pkgerrors "github.com/pkg/errors"
	//	. "github.com/open-ness/EMCO/src/rsync/pkg/types"
)

type AppContextUtils struct {
	ac appcontext.AppContext
}

//GetAppContextFlag gets the stop flag
func (a *AppContextUtils) GetAppContextFlag(key string) (bool, error) {
	h, err := a.ac.GetCompositeAppHandle()
	if err != nil {
		log.Error("Error GetAppContextFlag", log.Fields{"err": err})
		// Treat an error as stop
		return true, err
	}
	sh, err := a.ac.GetLevelHandle(h, key)
	if sh != nil {
		if v, err := a.ac.GetValue(sh); err == nil {
			return v.(bool), nil
		}
	}
	return true, err
}

//UpdateAppContextFlag to update flags
func (a *AppContextUtils) UpdateAppContextFlag(key string, b bool) error {
	h, err := a.ac.GetCompositeAppHandle()
	if err != nil {
		log.Error("Error UpdateAppContextFlag", log.Fields{"err": err})
		return err
	}
	sh, err := a.ac.GetLevelHandle(h, key)
	if sh == nil {
		_, err = a.ac.AddLevelValue(h, key, b)
	} else {
		err = a.ac.UpdateValue(sh, b)
	}
	if err != nil {
		log.Error("Error UpdateAppContextFlag", log.Fields{"err": err})
	}
	return err

}

//UpdateAppContextStatus updates a field in AppContext
func (a *AppContextUtils) UpdateAppContextStatus(key string, status interface{}) error {
	//var acStatus appcontext.AppContextStatus = appcontext.AppContextStatus{}
	hc, err := a.ac.GetCompositeAppHandle()
	if err != nil {
		log.Error("Error UpdateAppContextStatus", log.Fields{"err": err})
		return err
	}
	dsh, err := a.ac.GetLevelHandle(hc, key)
	if dsh == nil {
		_, err = a.ac.AddLevelValue(hc, key, status)
	} else {
		err = a.ac.UpdateValue(dsh, status)
	}
	if err != nil {
		log.Error("Error UpdateAppContextStatus", log.Fields{"err": err})
	}
	return err

}

//GetAppContextStatus gets the status
func (a *AppContextUtils) GetAppContextStatus(key string) (appcontext.AppContextStatus, error) {
	var acStatus appcontext.AppContextStatus = appcontext.AppContextStatus{}

	hc, err := a.ac.GetCompositeAppHandle()
	if err != nil {
		log.Error("Error GetAppContextStatus", log.Fields{"err": err})
		return acStatus, err
	}
	dsh, err := a.ac.GetLevelHandle(hc, key)
	if dsh != nil {
		v, err := a.ac.GetValue(dsh)
		if err != nil {
			log.Error("Error GetAppContextStatus", log.Fields{"err": err})
			return acStatus, err
		}
		//s := fmt.Sprintf("%v", v)
		//acStatus.Status = appcontext.StatusValue(s)
		acStatus = appcontext.AppContextStatus{}
		js, err := json.Marshal(v)
		if err != nil {
			log.Error("Error GetAppContextStatus", log.Fields{"err": err})
			return acStatus, err
		}
		err = json.Unmarshal(js, &acStatus)
		if err != nil {
			log.Error("Error GetAppContextStatus", log.Fields{"err": err})
			return acStatus, err
		}
	}
	return acStatus, err
}

// SetClusterReadyStatus gets the cluster ready status
func (a *AppContextUtils) SetClusterReadyStatus(app, cluster string, status appcontext.StatusValue) {
	ch, err := a.ac.GetClusterHandle(app, cluster)
	if err != nil {
		return
	}
	rsh, _ := a.ac.GetLevelHandle(ch, "readystatus")
	// If readystatus handle was not found, then create it
	if rsh == nil {
		a.ac.AddLevelValue(ch, "readystatus", status)
	} else {
		a.ac.UpdateStatusValue(rsh, status)
	}
	return
}

// GetClusterReadyStatus sets the cluster ready status
// does not return an error, just a status of Unknown if the cluster readystatus key does
// not exist or any other error occurs.
func (a *AppContextUtils) GetClusterReadyStatus(app, cluster string) appcontext.StatusValue {
	ch, err := a.ac.GetClusterHandle(app, cluster)
	if err != nil {
		return appcontext.ClusterReadyStatusEnum.Unknown
	}
	rsh, _ := a.ac.GetLevelHandle(ch, "readystatus")
	if rsh != nil {
		status, err := a.ac.GetValue(rsh)
		if err != nil {
			return appcontext.ClusterReadyStatusEnum.Unknown
		}
		return status.(appcontext.StatusValue)
	}

	return appcontext.ClusterReadyStatusEnum.Unknown
}

// GetRes Reads resource
func (a *AppContextUtils) GetRes(name string, app string, cluster string) ([]byte, interface{}, error) {
	var byteRes []byte

	rh, err := a.ac.GetResourceHandle(app, cluster, name)
	if err != nil {
		log.Error("Error GetRes", log.Fields{"err": err})
		return nil, nil, err
	}
	sh, err := a.ac.GetLevelHandle(rh, "status")
	if err != nil {
		statusPending := resourcestatus.ResourceStatus{
			Status: resourcestatus.RsyncStatusEnum.Pending,
		}
		sh, err = a.ac.AddLevelValue(rh, "status", statusPending)
		if err != nil {
			log.Error("Error GetRes", log.Fields{"err": err})
			return nil, nil, err
		}
	}
	resval, err := a.ac.GetValue(rh)
	if err != nil {
		log.Error("Error GetRes", log.Fields{"err": err})
		return nil, sh, err
	}
	if resval != "" {
		result := strings.Split(name, "+")
		if result[0] == "" {
			log.Error("Error GetRes, Resource name is nil", log.Fields{})
			return nil, sh, pkgerrors.Errorf("Resource name is nil %s:", name)
		}
		byteRes = []byte(fmt.Sprintf("%v", resval.(interface{})))
	} else {
		log.Error("Error GetRes, Resource name is nil", log.Fields{})
		return nil, sh, pkgerrors.Errorf("Resource value is nil %s", name)
	}
	return byteRes, sh, nil
}

// GetSubResApprove Reads sub resource
func (a *AppContextUtils) GetSubResApprove(name string, app string, cluster string) ([]byte, interface{}, error) {
	var byteRes []byte
	rh, err := a.ac.GetResourceHandle(app, cluster, name)
	if err != nil {
		log.Error("Error GetSubResApprove", log.Fields{"err": err})
		return nil, nil, err
	}
	// Check if Subresource defined
	sh, err := a.ac.GetLevelHandle(rh, "subresource/approval")
	if err != nil {
		return nil, nil, err
	}
	resval, err := a.ac.GetValue(sh)
	if err != nil {
		return nil, sh, err
	}
	if resval != "" {
		byteRes = []byte(fmt.Sprintf("%v", resval.(interface{})))
	} else {
		log.Error("Error GetSubResApprove, Resource name is nil", log.Fields{})
		return nil, sh, pkgerrors.Errorf("SubResource value is nil %s", name)
	}
	return byteRes, sh, nil
}

//GetNamespace reads namespace from metadata
func (a *AppContextUtils) GetNamespace() (string, string) {

	namespace := "default"
	level := "0"
	appmeta, err := a.ac.GetCompositeAppMeta()
	if err == nil {
		namespace = appmeta.Namespace
		level = appmeta.Level
	}
	log.Info("CloudConfig for this app will be looked up using level and namespace specified", log.Fields{
		"level":     level,
		"namespace": namespace,
	})
	return namespace, level
}

// PutRes copies resource into appContext
func (a *AppContextUtils) PutRes(name string, app string, cluster string, data []byte) error {

	rh, err := a.ac.GetResourceHandle(app, cluster, name)
	if err != nil {
		log.Error("Error GetResourceHandle", log.Fields{"err": err})
		return err
	}
	handle, _ := a.ac.GetLevelHandle(rh, "definition")
	// If definition handle was not found, then create it
	if handle == nil {
		a.ac.AddLevelValue(rh, "definition", string(data))
	} else {
		a.ac.UpdateStatusValue(handle, string(data))
	}
	return nil
}

//GetAppContextFlag gets the statusappctxid
func (a *AppContextUtils) GetStatusAppContext(key string) (string, error) {
	h, err := a.ac.GetCompositeAppHandle()
	if err != nil {
		log.Error("Error GetAppContextFlag", log.Fields{"err": err})
		return "", err
	}
	sh, err := a.ac.GetLevelHandle(h, key)
	if sh != nil {
		if v, err := a.ac.GetValue(sh); err == nil {
			return v.(string), nil
		}
	}
	return "", err
}

// Add resource level for a status
// Function adds any missing levels to AppContext
func (a *AppContextUtils) AddResourceStatus(name string, app string, cluster string, status interface{}) error {
	var rh, ch, ah interface{}

	rh, err := a.ac.GetResourceHandle(app, cluster, name)
	if err != nil {
		// Assume the resource doesn't exist
		h, err := a.ac.GetCompositeAppHandle()
		if err != nil {
			log.Error("Composite App Handle not found", log.Fields{"err": err})
			return err
		}
		// Check if App exists if not add handle
		ah, err = a.ac.GetAppHandle(app)
		if err != nil {
			//Add App level
			ah, err = a.ac.AddApp(h, app)
			if err != nil {
				log.Error("Unable to add application to context for status", log.Fields{"err": err})
				return err
			}
		}
		ch, err = a.ac.GetClusterHandle(app, cluster)
		if err != nil {
			ch, err = a.ac.AddCluster(ah, cluster)
			if err != nil {
				log.Error("Unable to add cluster to context for status", log.Fields{"err": err})
				return err
			}
		}
		rh, err = a.ac.AddResource(ch, name, nil)
		if err != nil {
			log.Error("Unable to add resource to context for status", log.Fields{"err": err})
			return err
		}
	}
	sh, err := a.ac.GetLevelHandle(rh, "status")
	if err != nil {
		sh, err = a.ac.AddLevelValue(rh, "status", status)
		if err != nil {
			log.Error("Error add status to resource", log.Fields{"err": err})
			return err
		}
	} else {
		a.ac.UpdateStatusValue(sh, status)
	}
	return nil
}
