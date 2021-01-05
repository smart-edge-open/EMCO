// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2020 Intel Corporation

package model

import (
	"encoding/json"

	mtypes "github.com/open-ness/EMCO/src/orchestrator/pkg/module/types"
)

// Controller contains the parameters needed for Controllers
// It implements the interface for managing the Controllers
type Controller struct {
	Metadata mtypes.Metadata `json:"metadata,omitempty"`
	Spec     ControllerSpec  `json:"spec,omitempty"`
}

type ControllerSpec struct {
	Host     string `json:"host,omitempty"`
	Port     int    `json:"port,omitempty"`
	Priority int    `json:"priority,omitempty"`
}

const MinControllerPriority = 1
const MaxControllerPriority = 1000000

// ControllerKey is the key structure that is used in the database
type ControllerKey struct {
	ControllerName string `json:"clm-controller-name"`
}

// We will use json marshalling to convert to string to
// preserve the underlying structure.
func (mk ControllerKey) String() string {
	out, err := json.Marshal(mk)
	if err != nil {
		return ""
	}

	return string(out)
}
