// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2021 Intel Corporation

package module

import (
	"strings"

	"github.com/open-ness/EMCO/src/orchestrator/pkg/infra/db"
	"github.com/open-ness/EMCO/src/sfc/pkg/model"

	pkgerrors "github.com/pkg/errors"
)

// CreateSfcClientSelectorIntent - create a new SfcClientSelectorIntent
func (v *SfcClientSelectorIntentClient) CreateSfcClientSelectorIntent(intent model.SfcClientSelectorIntent, pr, ca, caver, dig, netctrlint, sfcIntent string, exists bool) (model.SfcClientSelectorIntent, error) {
	//Construct key and tag to select the entry
	key := model.SfcClientSelectorIntentKey{
		Project:                 pr,
		CompositeApp:            ca,
		CompositeAppVersion:     caver,
		DigName:                 dig,
		NetControlIntent:        netctrlint,
		SfcIntent:               sfcIntent,
		SfcClientSelectorIntent: intent.Metadata.Name,
	}

	endKey := model.SfcEndKey{
		ChainEnd: intent.Spec.ChainEnd,
	}

	//Check if the SFC Intent exists
	_, err := NewSfcIntentClient().GetSfcIntent(sfcIntent, pr, ca, caver, dig, netctrlint)
	if err != nil {
		return model.SfcClientSelectorIntent{}, pkgerrors.Errorf("Parent SFC Intent does not exist: %v", sfcIntent)
	}

	//Check if this SFC Client Selector Intent already exists
	_, err = v.GetSfcClientSelectorIntent(intent.Metadata.Name, pr, ca, caver, dig, netctrlint, sfcIntent)
	if err == nil && !exists {
		return model.SfcClientSelectorIntent{}, pkgerrors.New("SFC Client Selector Intent already exists")
	}

	err = db.DBconn.Insert(v.db.storeName, key, endKey, v.db.tagMeta, intent)
	if err != nil {
		return model.SfcClientSelectorIntent{}, pkgerrors.Wrap(err, "Creating DB Entry")
	}

	return intent, nil
}

// GetSfcClientSelectorIntent returns the SfcClientSelectorIntent for corresponding name
func (v *SfcClientSelectorIntentClient) GetSfcClientSelectorIntent(name, pr, ca, caver, dig, netctrlint, sfcIntent string) (model.SfcClientSelectorIntent, error) {
	//Construct key and tag to select the entry
	key := model.SfcClientSelectorIntentKey{
		Project:                 pr,
		CompositeApp:            ca,
		CompositeAppVersion:     caver,
		DigName:                 dig,
		NetControlIntent:        netctrlint,
		SfcIntent:               sfcIntent,
		SfcClientSelectorIntent: name,
	}

	value, err := db.DBconn.Find(v.db.storeName, key, v.db.tagMeta)
	if err != nil {
		return model.SfcClientSelectorIntent{}, pkgerrors.Wrap(err, "db Find error")
	} else if len(value) == 0 {
		return model.SfcClientSelectorIntent{}, pkgerrors.New("SFC Client Selector Intent not found")
	}

	//value is a byte array
	if value != nil {
		intent := model.SfcClientSelectorIntent{}
		err = db.DBconn.Unmarshal(value[0], &intent)
		if err != nil {
			return model.SfcClientSelectorIntent{}, pkgerrors.Wrap(err, "Unmarshalling Value")
		}
		return intent, nil
	}

	return model.SfcClientSelectorIntent{}, pkgerrors.New("Error getting SFC Client Selector Intent")
}

// GetAllSfcClientSelectorIntent returns all of the SFC Intents for for the given network control intent
func (v *SfcClientSelectorIntentClient) GetAllSfcClientSelectorIntents(pr, ca, caver, dig, netctrlint, sfcIntent string) ([]model.SfcClientSelectorIntent, error) {
	//Construct key and tag to select the entry
	key := model.SfcClientSelectorIntentKey{
		Project:                 pr,
		CompositeApp:            ca,
		CompositeAppVersion:     caver,
		DigName:                 dig,
		NetControlIntent:        netctrlint,
		SfcIntent:               sfcIntent,
		SfcClientSelectorIntent: "",
	}

	resp := make([]model.SfcClientSelectorIntent, 0)

	// Verify the SFC intent exists
	_, err := NewSfcIntentClient().GetSfcIntent(sfcIntent, pr, ca, caver, dig, netctrlint)
	if err != nil {
		return resp, err
	}

	values, err := db.DBconn.Find(v.db.storeName, key, v.db.tagMeta)
	if err != nil {
		return resp, pkgerrors.Wrap(err, "db Find error")
	}

	for _, value := range values {
		cp := model.SfcClientSelectorIntent{}
		err = db.DBconn.Unmarshal(value, &cp)
		if err != nil {
			return resp, pkgerrors.Wrap(err, "Unmarshalling Value")
		}
		resp = append(resp, cp)
	}

	return resp, nil
}

// GetSfcClientSelectorIntentsByEnd returns all of the SFC Client Selector Intents for for the given network control intent
// and specified end of the chain
func (v *SfcClientSelectorIntentClient) GetSfcClientSelectorIntentsByEnd(pr, ca, caver, dig, netctrlint, sfcIntent, chainEnd string) ([]model.SfcClientSelectorIntent, error) {
	//Construct key and tag to select the entry
	key := model.SfcClientSelectorIntentByEndKey{
		Project:             pr,
		CompositeApp:        ca,
		CompositeAppVersion: caver,
		DigName:             dig,
		NetControlIntent:    netctrlint,
		SfcIntent:           sfcIntent,
		ChainEnd:            chainEnd,
	}

	resp := make([]model.SfcClientSelectorIntent, 0)

	// Verify the SFC intent exists
	_, err := NewSfcIntentClient().GetSfcIntent(sfcIntent, pr, ca, caver, dig, netctrlint)
	if err != nil {
		return resp, err
	}

	values, err := db.DBconn.Find(v.db.storeName, key, v.db.tagMeta)
	if err != nil {
		return resp, pkgerrors.Wrap(err, "db Find error haha")
	}

	for _, value := range values {
		cp := model.SfcClientSelectorIntent{}
		err = db.DBconn.Unmarshal(value, &cp)
		if err != nil {
			return resp, pkgerrors.Wrap(err, "Unmarshalling Value")
		}
		resp = append(resp, cp)
	}

	return resp, nil
}

// DeleteSfcClientSelectorIntent deletes the SfcClientSelectorIntent from the database
func (v *SfcClientSelectorIntentClient) DeleteSfcClientSelectorIntent(name, pr, ca, caver, dig, netctrlint, sfcIntent string) error {

	//Construct key and tag to select the entry
	key := model.SfcClientSelectorIntentKey{
		Project:                 pr,
		CompositeApp:            ca,
		CompositeAppVersion:     caver,
		DigName:                 dig,
		NetControlIntent:        netctrlint,
		SfcIntent:               sfcIntent,
		SfcClientSelectorIntent: name,
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
