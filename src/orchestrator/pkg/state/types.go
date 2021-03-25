// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2020 Intel Corporation

package state

import "time"

// StateInfo struct is used to maintain the values for state, contextid, (and other)
// information about resources which can be instantiated via rsync.
// The last Actions entry holds the current state of the container object.
type StateInfo struct {
	// Same Status AppContext between instantiation and termination for a DIG
	StatusContextId string `json:"statusctxid"`
	Actions []ActionEntry `json:"actions"`
}

// ActionEntry is used to keep track of the time an action (e.g. Created, Instantiate, Terminate) was invoked
// For actions where an AppContext is relevent, the ContextId field will be non-zero length
type ActionEntry struct {
	State     StateValue `json:"state"`
	ContextId string     `json:"instance"`
	TimeStamp time.Time  `json:"time"`
	Revision  int64      `json:"revision"`
}

type StateValue = string

type states struct {
	Undefined          StateValue
	Created            StateValue
	Approved           StateValue
	Applied            StateValue
	Instantiated       StateValue
	Terminated         StateValue
	InstantiateStopped StateValue
	TerminateStopped   StateValue
	Updated		StateValue
}

var StateEnum = &states{
	Undefined:          "Undefined",
	Created:            "Created",
	Approved:           "Approved",
	Applied:            "Applied",
	Instantiated:       "Instantiated",
	Terminated:         "Terminated",
	InstantiateStopped: "InstantiateStopped",
	TerminateStopped:   "TerminateStopped",
	Updated:            "Updated",
}
