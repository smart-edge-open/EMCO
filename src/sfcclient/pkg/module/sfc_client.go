// SPDX-License-Identifier: Apache-2.0
// C[]model.SfcClientIntent{}opyright (c) 2021 Intel Corporation

package module

import (
	"strings"

	"github.com/open-ness/EMCO/src/orchestrator/pkg/infra/db"
	ovn "github.com/open-ness/EMCO/src/ovnaction/pkg/module"
	"github.com/open-ness/EMCO/src/sfcclient/pkg/model"

	pkgerrors "github.com/pkg/errors"
)

// CreateSfcClientIntent - create a new SfcClientIntent
func (v *SfcClient) CreateSfcClientIntent(intent model.SfcClientIntent, pr, ca, caver, dig, netctrlint string, exists bool) (model.SfcClientIntent, error) {
	//Construct key and tag to select the entry
	key := model.SfcClientIntentKey{
		Project:             pr,
		CompositeApp:        ca,
		CompositeAppVersion: caver,
		DigName:             dig,
		NetControlIntent:    netctrlint,
		SfcClientIntent:     intent.Metadata.Name,
	}

	// Check for existence of parent NetControlIntent resource
	_, err := ovn.NewNetControlIntentClient().GetNetControlIntent(netctrlint, pr, ca, caver, dig)
	if err != nil {
		return model.SfcClientIntent{}, pkgerrors.Errorf("Parent NetControlIntent resource does not exist: %v", netctrlint)
	}

	//Check if this SFC Client Intent already exists
	_, err = v.GetSfcClientIntent(intent.Metadata.Name, pr, ca, caver, dig, netctrlint)
	if err == nil && !exists {
		return model.SfcClientIntent{}, pkgerrors.New("SFC Client Intent already exists")
	}

	err = db.DBconn.Insert(v.db.storeName, key, nil, v.db.tagMeta, intent)
	if err != nil {
		return model.SfcClientIntent{}, pkgerrors.Wrap(err, "Creating DB Entry")
	}

	return intent, nil
}

// GetSfcClientIntent returns the SfcClientIntent for corresponding name
func (v *SfcClient) GetSfcClientIntent(name, pr, ca, caver, dig, netctrlint string) (model.SfcClientIntent, error) {
	//Construct key and tag to select the entry
	key := model.SfcClientIntentKey{
		Project:             pr,
		CompositeApp:        ca,
		CompositeAppVersion: caver,
		DigName:             dig,
		NetControlIntent:    netctrlint,
		SfcClientIntent:     name,
	}

	value, err := db.DBconn.Find(v.db.storeName, key, v.db.tagMeta)
	if err != nil {
		return model.SfcClientIntent{}, pkgerrors.Wrap(err, "db Find error")
	} else if len(value) == 0 {
		return model.SfcClientIntent{}, pkgerrors.New("SFC Client Intent not found")
	}

	//value is a byte array
	if value != nil {
		intent := model.SfcClientIntent{}
		err = db.DBconn.Unmarshal(value[0], &intent)
		if err != nil {
			return model.SfcClientIntent{}, pkgerrors.Wrap(err, "Unmarshalling Value")
		}
		return intent, nil
	}

	return model.SfcClientIntent{}, pkgerrors.New("Error getting SFC Client Intent")
}

// GetAllSfcClientIntent returns all of the SFC Client Intents for for the given network control intent
func (v *SfcClient) GetAllSfcClientIntents(pr, ca, caver, dig, netctrlint string) ([]model.SfcClientIntent, error) {
	//Construct key and tag to select the entry
	key := model.SfcClientIntentKey{
		Project:             pr,
		CompositeApp:        ca,
		CompositeAppVersion: caver,
		DigName:             dig,
		NetControlIntent:    netctrlint,
		SfcClientIntent:     "",
	}

	resp := make([]model.SfcClientIntent, 0)

	// Verify the Net Control Intent exists
	_, err := ovn.NewNetControlIntentClient().GetNetControlIntent(netctrlint, pr, ca, caver, dig)
	if err != nil {
		return resp, err
	}

	values, err := db.DBconn.Find(v.db.storeName, key, v.db.tagMeta)
	if err != nil {
		return resp, pkgerrors.Wrap(err, "db Find error")
	}

	for _, value := range values {
		cp := model.SfcClientIntent{}
		err = db.DBconn.Unmarshal(value, &cp)
		if err != nil {
			return resp, pkgerrors.Wrap(err, "Unmarshalling Value")
		}
		resp = append(resp, cp)
	}

	return resp, nil
}

// DeleteSfcClientIntent deletes the SfcClientIntent from the database
func (v *SfcClient) DeleteSfcClientIntent(name, pr, ca, caver, dig, netctrlint string) error {

	//Construct key and tag to select the entry
	key := model.SfcClientIntentKey{
		Project:             pr,
		CompositeApp:        ca,
		CompositeAppVersion: caver,
		DigName:             dig,
		NetControlIntent:    netctrlint,
		SfcClientIntent:     name,
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
