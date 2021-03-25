// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2020 Intel Corporation

package types

import (
	"k8s.io/apimachinery/pkg/runtime/schema"

	"github.com/open-ness/EMCO/src/orchestrator/pkg/appcontext"
)

const (
	CurrentStateKey         string = "rsync/state/CurrentState"
	DesiredStateKey         string = "rsync/state/DesiredState"
	PendingTerminateFlagKey string = "rsync/state/PendingTerminateFlag"
	AppContextEventQueueKey string = "rsync/AppContextEventQueue"
	StatusKey               string = "status"
	StopFlagKey             string = "stopflag"
	StatusAppContextIDKey   string = "statusappctxid"
)

// RsyncEvent is event Rsync handles
type RsyncEvent string

// Rsync Event types
const (
	InstantiateEvent     RsyncEvent = "Instantiate"
	TerminateEvent       RsyncEvent = "Terminate"
	ReadEvent            RsyncEvent = "Read"
	AddChildContextEvent RsyncEvent = "AddChildContext"
	UpdateEvent          RsyncEvent = "Update"
	// This is an internal event
	UpdateModifyEvent          RsyncEvent = "UpdateModify"
)

// RsyncOperation is operation Rsync handles
type RsyncOperation int

// Rsync Operations
const (
	OpApply RsyncOperation = iota
	OpDelete
	OpRead
)

func (d RsyncOperation) String() string {
	return [...]string{"Apply", "Delete", "Read"}[d]
}

// StateChange represents a state change rsync handles
type StateChange struct {
	// Name of the event
	Event RsyncEvent
	// List of states that can handle this event
	SState []appcontext.StatusValue
	// Dst state if the state transition was successful
	DState appcontext.StatusValue
	// Current state if the state transition was successful
	CState appcontext.StatusValue
	// Error state if the state transition was unsuccessful
	ErrState appcontext.StatusValue
}

// StateChanges represent State Machine for the AppContext
var StateChanges = map[RsyncEvent]StateChange{

	InstantiateEvent: StateChange{
		SState: []appcontext.StatusValue{
			appcontext.AppContextStatusEnum.Created,
			appcontext.AppContextStatusEnum.Instantiated,
			appcontext.AppContextStatusEnum.InstantiateFailed,
			appcontext.AppContextStatusEnum.Instantiating},
		DState:   appcontext.AppContextStatusEnum.Instantiated,
		CState:   appcontext.AppContextStatusEnum.Instantiating,
		ErrState: appcontext.AppContextStatusEnum.InstantiateFailed,
	},
	TerminateEvent: StateChange{
		SState: []appcontext.StatusValue{
			appcontext.AppContextStatusEnum.InstantiateFailed,
			appcontext.AppContextStatusEnum.Instantiating,
			appcontext.AppContextStatusEnum.Instantiated,
			appcontext.AppContextStatusEnum.TerminateFailed,
			appcontext.AppContextStatusEnum.Terminating},
		DState:   appcontext.AppContextStatusEnum.Terminated,
		CState:   appcontext.AppContextStatusEnum.Terminating,
		ErrState: appcontext.AppContextStatusEnum.TerminateFailed,
	},
	UpdateEvent: StateChange{
		SState: []appcontext.StatusValue{
			appcontext.AppContextStatusEnum.Instantiated},
		DState:   appcontext.AppContextStatusEnum.Updated,
		CState:   appcontext.AppContextStatusEnum.Updating,
		ErrState: appcontext.AppContextStatusEnum.UpdateFailed,
	},
	UpdateModifyEvent: StateChange{
		SState: []appcontext.StatusValue{
			appcontext.AppContextStatusEnum.Created,
			appcontext.AppContextStatusEnum.Updated},
		DState:   appcontext.AppContextStatusEnum.Instantiated,
		CState:   appcontext.AppContextStatusEnum.Instantiating,
		ErrState: appcontext.AppContextStatusEnum.InstantiateFailed,
	},
	ReadEvent: StateChange{
		SState: []appcontext.StatusValue{
			appcontext.AppContextStatusEnum.Created,
			appcontext.AppContextStatusEnum.Instantiated,
			appcontext.AppContextStatusEnum.InstantiateFailed,
			appcontext.AppContextStatusEnum.Instantiating},
		DState:   appcontext.AppContextStatusEnum.Instantiated,
		CState:   appcontext.AppContextStatusEnum.Instantiating,
		ErrState: appcontext.AppContextStatusEnum.InstantiateFailed,
	},
}

// Resource Dependency Structures
type Resource struct {
	App string                  `json:"app,omitempty"`
	Res string                  `json:"name,omitempty"`
	GVK schema.GroupVersionKind `json:"gvk,omitempty"`
}
// Criteria for Resource dependency
type Criteria struct {
	// Ready or deployed
	OpStatus string `json:"opstatus,omitempty"`
	// Wait time in seconds
	Wait string `json:"wait,omitempty"`
}

// Dependency Structures
type Dependency struct {
	Resource Resource `json:"resource,omitempty"`
	criteria Criteria `json:"criteria,omitempty"`
}

// ResourceDependency structure
type ResourceDependency struct {
	Resource Resource     `json:"resource,omitempty"`
	Dep      []Dependency `json:"dependency,omitempty"`
}

// CompositeApp Structures
type CompositeApp struct {
	Name         string                      `json:"name,omitempty"`
	CompMetadata appcontext.CompositeAppMeta `json:"compmetadat,omitempty"`
	AppOrder     []string                    `json:"appOrder,omitempty"`
	Apps         map[string]*App              `json:"apps,omitempty"`
}
// AppResource represents a resource
type AppResource struct {
	Name       string              `json:"name,omitempty"`
	Data       interface{}         `json:"data,omitempty"`
	Dependency map[string]*Criteria `json:"depenedency,omitempty"`
	// Needed to suport updates
	Skip bool `json:"bool,omitempty"`
}
// Cluster is a cluster within an App
type Cluster struct {
	Name      string                 `json:"name,omitempty"`
	ResOrder  []string               `json:"reorder,omitempty"`
	Resources map[string]*AppResource `json:"resources,omitempty"`
	// Needed to suport updates
	Skip bool `json:"bool,omitempty"`
}
// App is an app within a composite app
type App struct {
	Name       string              `json:"name,omitempty"`
	Clusters   map[string]*Cluster  `json:"clusters,omitempty"`
	Dependency map[string]*Criteria `json:"dependency,omitempty"`
	// Needed to suport updates
	Skip       bool                `json:"bool,omitempty"`
}
// ClientProvider is interface for client
type ClientProvider interface {
	Apply(content []byte) error
	Delete(content []byte) error
	Get(gvkRes []byte, namespace string) ([]byte, error)
	Approve(name string, sa []byte) error
	IsReachable() error
	TagResource([]byte, string) ([]byte, error)
}
// Connector is interface for connection to Cluster
type Connector interface {
	Init(id interface{}) error
	GetClientInternal(cluster string, level string, namespace string) (ClientProvider, error)
	RemoveClient()
	StartClusterWatcher(cluster string) error
	GetStatusCR(label string) ([]byte, error)
}
// AppContextQueueElement element in per AppContext Queue
type AppContextQueueElement struct {
	Event RsyncEvent `json:"event"`
	// Only valid in case of update events
	UCID string `json:"uCID,omitempty"`
	// Status - Pending, Done, Error, skip
	Status string `json:"status"`
}
// AppContextQueue per AppContext queue
type AppContextQueue struct {
	AcQueue []AppContextQueueElement
}
