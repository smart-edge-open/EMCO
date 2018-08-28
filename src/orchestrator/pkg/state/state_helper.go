// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2020 Intel Corporation

package state

import (
	"encoding/json"

	"github.com/open-ness/EMCO/src/orchestrator/pkg/appcontext"

	pkgerrors "github.com/pkg/errors"
)

// GetAppContextFromStateInfo loads the appcontext present in the StateInfo input
func GetAppContextFromId(ctxid string) (appcontext.AppContext, error) {
	var cc appcontext.AppContext
	_, err := cc.LoadAppContext(ctxid)
	if err != nil {
		return appcontext.AppContext{}, err
	}
	return cc, nil
}

// GetCurrentStateFromStatInfo gets the last (current) state from StateInfo
func GetCurrentStateFromStateInfo(s StateInfo) (StateValue, error) {
	alen := len(s.Actions)
	if alen == 0 {
		return StateEnum.Undefined, pkgerrors.Errorf("No state information")
	}
	return s.Actions[alen-1].State, nil
}

// GetLastContextFromStatInfo gets the last (most recent) context id from StateInfo
func GetLastContextIdFromStateInfo(s StateInfo) string {
	alen := len(s.Actions)
	if alen > 0 {
		return s.Actions[alen-1].ContextId
	} else {
		return ""
	}
}

// GetContextIdsFromStatInfo return a list of the unique AppContext Ids in the StateInfo
func GetContextIdsFromStateInfo(s StateInfo) []string {
	m := make(map[string]string)

	for _, a := range s.Actions {
		if a.ContextId != "" {
			m[a.ContextId] = ""
		}
	}

	ids := make([]string, len(m))
	i := 0
	for k := range m {
		ids[i] = k
		i++
	}

	return ids
}

func GetAppContextStatus(ctxid string) (appcontext.AppContextStatus, error) {

	ac, err := GetAppContextFromId(ctxid)
	if err != nil {
		return appcontext.AppContextStatus{}, err
	}

	h, err := ac.GetCompositeAppHandle()
	if err != nil {
		return appcontext.AppContextStatus{}, err
	}
	sh, err := ac.GetLevelHandle(h, "status")
	if err != nil {
		return appcontext.AppContextStatus{}, err
	}
	s, err := ac.GetValue(sh)
	if err != nil {
		return appcontext.AppContextStatus{}, err
	}
	acStatus := appcontext.AppContextStatus{}
	js, _ := json.Marshal(s)
	json.Unmarshal(js, &acStatus)

	return acStatus, nil

}

func UpdateAppContextStopFlag(ctxid string, sf bool) error {
	ac, err := GetAppContextFromId(ctxid)
	if err != nil {
		return err
	}
	hc, err := ac.GetCompositeAppHandle()
	if err != nil {
		return err
	}
	sh, err := ac.GetLevelHandle(hc, "stopflag")
	if sh == nil {
		_, err = ac.AddLevelValue(hc, "stopflag", sf)
	} else {
		err = ac.UpdateValue(sh, sf)
	}
	if err != nil {
		return err
	}
	return nil
}
