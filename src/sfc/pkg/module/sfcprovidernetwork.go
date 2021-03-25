// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2021 Intel Corporation

package module

import (
	"strings"

	"github.com/open-ness/EMCO/src/orchestrator/pkg/infra/db"
	"github.com/open-ness/EMCO/src/sfc/pkg/model"

	pkgerrors "github.com/pkg/errors"
)

// CreateSfcProviderNetworkIntent - create a new SfcProviderNetworkIntent
func (v *SfcProviderNetworkIntentClient) CreateSfcProviderNetworkIntent(intent model.SfcProviderNetworkIntent, pr, ca, caver, dig, netctrlint, sfcIntent string, exists bool) (model.SfcProviderNetworkIntent, error) {
	//Construct key and tag to select the entry
	key := model.SfcProviderNetworkIntentKey{
		Project:                  pr,
		CompositeApp:             ca,
		CompositeAppVersion:      caver,
		DigName:                  dig,
		NetControlIntent:         netctrlint,
		SfcIntent:                sfcIntent,
		SfcProviderNetworkIntent: intent.Metadata.Name,
	}

	//Check if the SFC Intent exists
	_, err := NewSfcIntentClient().GetSfcIntent(sfcIntent, pr, ca, caver, dig, netctrlint)
	if err != nil {
		return model.SfcProviderNetworkIntent{}, pkgerrors.Errorf("Parent SFC Intent does not exist: %v", sfcIntent)
	}

	//Check if this SFC Provider Network Intent already exists
	_, err = v.GetSfcProviderNetworkIntent(intent.Metadata.Name, pr, ca, caver, dig, netctrlint, sfcIntent)
	if err == nil && !exists {
		return model.SfcProviderNetworkIntent{}, pkgerrors.New("SFC Provider Network Intent already exists")
	}

	err = db.DBconn.Insert(v.db.storeName, key, nil, v.db.tagMeta, intent)
	if err != nil {
		return model.SfcProviderNetworkIntent{}, pkgerrors.Wrap(err, "Creating DB Entry")
	}

	return intent, nil
}

// GetSfcProviderNetworkIntent returns the SfcProviderNetworkIntent for corresponding name
func (v *SfcProviderNetworkIntentClient) GetSfcProviderNetworkIntent(name, pr, ca, caver, dig, netctrlint, sfcIntent string) (model.SfcProviderNetworkIntent, error) {
	//Construct key and tag to select the entry
	key := model.SfcProviderNetworkIntentKey{
		Project:                  pr,
		CompositeApp:             ca,
		CompositeAppVersion:      caver,
		DigName:                  dig,
		NetControlIntent:         netctrlint,
		SfcIntent:                sfcIntent,
		SfcProviderNetworkIntent: name,
	}

	value, err := db.DBconn.Find(v.db.storeName, key, v.db.tagMeta)
	if err != nil {
		return model.SfcProviderNetworkIntent{}, pkgerrors.Wrap(err, "db Find error")
	} else if len(value) == 0 {
		return model.SfcProviderNetworkIntent{}, pkgerrors.New("SFC Provider Network Intent not found")
	}

	//value is a byte array
	if value != nil {
		intent := model.SfcProviderNetworkIntent{}
		err = db.DBconn.Unmarshal(value[0], &intent)
		if err != nil {
			return model.SfcProviderNetworkIntent{}, pkgerrors.Wrap(err, "Unmarshalling Value")
		}
		return intent, nil
	}

	return model.SfcProviderNetworkIntent{}, pkgerrors.New("Error getting SFC Provider Network Intent")
}

// GetAllSfcProviderNetworkIntent returns all of the SFC Intents for for the given network control intent
func (v *SfcProviderNetworkIntentClient) GetAllSfcProviderNetworkIntents(pr, ca, caver, dig, netctrlint, sfcIntent string) ([]model.SfcProviderNetworkIntent, error) {
	//Construct key and tag to select the entry
	key := model.SfcProviderNetworkIntentKey{
		Project:                  pr,
		CompositeApp:             ca,
		CompositeAppVersion:      caver,
		DigName:                  dig,
		NetControlIntent:         netctrlint,
		SfcIntent:                sfcIntent,
		SfcProviderNetworkIntent: "",
	}

	resp := make([]model.SfcProviderNetworkIntent, 0)

	// verify SFC Intent exists
	_, err := NewSfcIntentClient().GetSfcIntent(sfcIntent, pr, ca, caver, dig, netctrlint)
	if err != nil {
		return resp, err
	}

	values, err := db.DBconn.Find(v.db.storeName, key, v.db.tagMeta)
	if err != nil {
		return resp, pkgerrors.Wrap(err, "db Find error")
	}

	for _, value := range values {
		cp := model.SfcProviderNetworkIntent{}
		err = db.DBconn.Unmarshal(value, &cp)
		if err != nil {
			return resp, pkgerrors.Wrap(err, "Unmarshalling Value")
		}
		resp = append(resp, cp)
	}

	return resp, nil
}

// GetSfcProviderNetworkIntentByEnd returns all of the SFC Provider Network Intents for for the given network control intent
func (v *SfcProviderNetworkIntentClient) GetSfcProviderNetworkIntentsByEnd(pr, ca, caver, dig, netctrlint, sfcIntent, chainEnd string) ([]model.SfcProviderNetworkIntent, error) {
	//Construct key and tag to select the entry
	key := model.SfcProviderNetworkIntentByEndKey{
		Project:             pr,
		CompositeApp:        ca,
		CompositeAppVersion: caver,
		DigName:             dig,
		NetControlIntent:    netctrlint,
		SfcIntent:           sfcIntent,
		ChainEnd:            chainEnd,
	}

	resp := make([]model.SfcProviderNetworkIntent, 0)

	// verify SFC Intent exists
	_, err := NewSfcIntentClient().GetSfcIntent(sfcIntent, pr, ca, caver, dig, netctrlint)
	if err != nil {
		return resp, err
	}

	values, err := db.DBconn.Find(v.db.storeName, key, v.db.tagMeta)
	if err != nil {
		return resp, pkgerrors.Wrap(err, "db Find error")
	}

	for _, value := range values {
		cp := model.SfcProviderNetworkIntent{}
		err = db.DBconn.Unmarshal(value, &cp)
		if err != nil {
			return resp, pkgerrors.Wrap(err, "Unmarshalling Value")
		}
		resp = append(resp, cp)
	}

	return resp, nil
}

// DeleteSfcProviderNetworkIntent deletes the SfcProviderNetworkIntent from the database
func (v *SfcProviderNetworkIntentClient) DeleteSfcProviderNetworkIntent(name, pr, ca, caver, dig, netctrlint, sfcIntent string) error {

	//Construct key and tag to select the entry
	key := model.SfcProviderNetworkIntentKey{
		Project:                  pr,
		CompositeApp:             ca,
		CompositeAppVersion:      caver,
		DigName:                  dig,
		NetControlIntent:         netctrlint,
		SfcIntent:                sfcIntent,
		SfcProviderNetworkIntent: name,
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
