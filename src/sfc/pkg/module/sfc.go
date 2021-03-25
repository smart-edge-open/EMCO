// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2021 Intel Corporation

package module

import (
	"strings"

	"github.com/open-ness/EMCO/src/orchestrator/pkg/infra/db"
	ovn "github.com/open-ness/EMCO/src/ovnaction/pkg/module"
	"github.com/open-ness/EMCO/src/sfc/pkg/model"

	pkgerrors "github.com/pkg/errors"
)

// CreateSfcIntent - create a new SfcIntent
func (v *SfcIntentClient) CreateSfcIntent(intent model.SfcIntent, pr, ca, caver, dig, netctrlint string, exists bool) (model.SfcIntent, error) {
	//Construct key and tag to select the entry
	key := model.SfcIntentKey{
		Project:             pr,
		CompositeApp:        ca,
		CompositeAppVersion: caver,
		DigName:             dig,
		NetControlIntent:    netctrlint,
		SfcIntent:           intent.Metadata.Name,
	}

	// Check for existence of parent NetControlIntent resource
	_, err := ovn.NewNetControlIntentClient().GetNetControlIntent(netctrlint, pr, ca, caver, dig)
	if err != nil {
		return model.SfcIntent{}, pkgerrors.Errorf("Parent NetControlIntent resource does not exist: %v", netctrlint)
	}

	//Check if this SFC Intent already exists
	_, err = v.GetSfcIntent(intent.Metadata.Name, pr, ca, caver, dig, netctrlint)
	if err == nil && !exists {
		return model.SfcIntent{}, pkgerrors.New("SFC Intent already exists")
	}

	err = db.DBconn.Insert(v.db.storeName, key, nil, v.db.tagMeta, intent)
	if err != nil {
		return model.SfcIntent{}, pkgerrors.Wrap(err, "Creating DB Entry")
	}

	return intent, nil
}

// GetSfcIntent returns the SfcIntent for corresponding name
func (v *SfcIntentClient) GetSfcIntent(name, pr, ca, caver, dig, netctrlint string) (model.SfcIntent, error) {
	//Construct key and tag to select the entry
	key := model.SfcIntentKey{
		Project:             pr,
		CompositeApp:        ca,
		CompositeAppVersion: caver,
		DigName:             dig,
		NetControlIntent:    netctrlint,
		SfcIntent:           name,
	}

	value, err := db.DBconn.Find(v.db.storeName, key, v.db.tagMeta)
	if err != nil {
		return model.SfcIntent{}, pkgerrors.Wrap(err, "db Find error")
	} else if len(value) == 0 {
		return model.SfcIntent{}, pkgerrors.New("SFC Intent not found")
	}

	//value is a byte array
	if value != nil {
		intent := model.SfcIntent{}
		err = db.DBconn.Unmarshal(value[0], &intent)
		if err != nil {
			return model.SfcIntent{}, pkgerrors.Wrap(err, "Unmarshalling Value")
		}
		return intent, nil
	}

	return model.SfcIntent{}, pkgerrors.New("Error getting SFC Intent")
}

// GetAllSfcIntent returns all of the SFC Intents for for the given network control intent
func (v *SfcIntentClient) GetAllSfcIntents(pr, ca, caver, dig, netctrlint string) ([]model.SfcIntent, error) {
	//Construct key and tag to select the entry
	key := model.SfcIntentKey{
		Project:             pr,
		CompositeApp:        ca,
		CompositeAppVersion: caver,
		DigName:             dig,
		NetControlIntent:    netctrlint,
		SfcIntent:           "",
	}

	resp := make([]model.SfcIntent, 0)

	// Verify the Net Control Intent exists
	_, err := ovn.NewNetControlIntentClient().GetNetControlIntent(netctrlint, pr, ca, caver, dig)
	if err != nil {
		return resp, err
	}

	values, err := db.DBconn.Find(v.db.storeName, key, v.db.tagMeta)
	if err != nil {
		return []model.SfcIntent{}, pkgerrors.Wrap(err, "db Find error")
	}

	for _, value := range values {
		cp := model.SfcIntent{}
		err = db.DBconn.Unmarshal(value, &cp)
		if err != nil {
			return []model.SfcIntent{}, pkgerrors.Wrap(err, "Unmarshalling Value")
		}
		resp = append(resp, cp)
	}

	return resp, nil
}

// DeleteSfcIntent deletes the SfcIntent from the database
func (v *SfcIntentClient) DeleteSfcIntent(name, pr, ca, caver, dig, netctrlint string) error {

	//Construct key and tag to select the entry
	key := model.SfcIntentKey{
		Project:             pr,
		CompositeApp:        ca,
		CompositeAppVersion: caver,
		DigName:             dig,
		NetControlIntent:    netctrlint,
		SfcIntent:           name,
	}

	err := db.DBconn.Remove(v.db.storeName, key)
	if err != nil {
		if strings.Contains(err.Error(), "Error finding:") {
			return pkgerrors.Wrap(err, "db Remove error - not found")
		} else if strings.Contains(err.Error(), "Can't delete parent without deleting child") {
			return pkgerrors.Wrap(err, "db Remove error - conflict")
		} else {
			return pkgerrors.Wrap(err, "db Remove error - general")
		}
	}

	return nil
}
