// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2020 Intel Corporation

package networkintents

import (
	"strings"

	clusterPkg "github.com/open-ness/EMCO/src/clm/pkg/cluster"
	ncmtypes "github.com/open-ness/EMCO/src/ncm/pkg/module/types"
	nettypes "github.com/open-ness/EMCO/src/ncm/pkg/networkintents/types"
	"github.com/open-ness/EMCO/src/orchestrator/pkg/infra/db"
	mtypes "github.com/open-ness/EMCO/src/orchestrator/pkg/module/types"
	"github.com/open-ness/EMCO/src/orchestrator/pkg/state"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	pkgerrors "github.com/pkg/errors"
)

// ProviderNet contains the parameters needed for dynamic networks
type ProviderNet struct {
	Metadata mtypes.Metadata `json:"metadata"`
	Spec     ProviderNetSpec `json:"spec"`
}

type ProviderNetSpec struct {
	CniType         string                `json:"cniType" yaml:"cniType"`
	Ipv4Subnets     []nettypes.Ipv4Subnet `json:"ipv4Subnets" yaml:"ipv4Subnets"`
	ProviderNetType string                `json:"providerNetType" yaml:"providerNetType"`
	Vlan            nettypes.Vlan         `json:"vlan" yaml:"vlan"`
}

// structure for the Network Custom Resource
type CrProviderNet struct {
	ApiVersion      string            `yaml:"apiVersion"`
	Kind            string            `yaml:"kind"`
	MetaData        metav1.ObjectMeta `yaml:"metadata"`
	ProviderNetSpec ProviderNetSpec   `yaml:"spec"`
}

const PROVIDER_NETWORK_APIVERSION = "k8s.plugin.opnfv.org/v1alpha1"
const PROVIDER_NETWORK_KIND = "ProviderNetwork"

// ProviderNetKey is the key structure that is used in the database
type ProviderNetKey struct {
	ClusterProviderName string `json:"provider"`
	ClusterName         string `json:"cluster"`
	ProviderNetName     string `json:"providernet"`
}

// Manager is an interface exposing the ProviderNet functionality
type ProviderNetManager interface {
	CreateProviderNet(pr ProviderNet, clusterProvider, cluster string, exists bool) (ProviderNet, error)
	GetProviderNet(name, clusterProvider, cluster string) (ProviderNet, error)
	GetProviderNets(clusterProvider, cluster string) ([]ProviderNet, error)
	DeleteProviderNet(name, clusterProvider, cluster string) error
}

// ProviderNetClient implements the Manager
// It will also be used to maintain some localized state
type ProviderNetClient struct {
	db ncmtypes.ClientDbInfo
}

// NewProviderNetClient returns an instance of the ProviderNetClient
// which implements the Manager
func NewProviderNetClient() *ProviderNetClient {
	return &ProviderNetClient{
		db: ncmtypes.ClientDbInfo{
			StoreName: "cluster",
			TagMeta:   "networkmetadata",
		},
	}
}

// CreateProviderNet - create a new ProviderNet
func (v *ProviderNetClient) CreateProviderNet(p ProviderNet, clusterProvider, cluster string, exists bool) (ProviderNet, error) {

	// verify cluster exists and in state to add provider networks
	s, err := clusterPkg.NewClusterClient().GetClusterState(clusterProvider, cluster)
	if err != nil {
		return ProviderNet{}, pkgerrors.New("Unable to find the cluster")
	}
	stateVal, err := state.GetCurrentStateFromStateInfo(s)
	if err != nil {
		return ProviderNet{}, pkgerrors.Errorf("Error getting current state from Cluster stateInfo: " + cluster)
	}
	switch stateVal {
	case state.StateEnum.Approved:
		return ProviderNet{}, pkgerrors.Errorf("Cluster is in an invalid state: " + cluster + " " + state.StateEnum.Approved)
	case state.StateEnum.Terminated:
		break
	case state.StateEnum.Created:
		break
	case state.StateEnum.Applied:
		return ProviderNet{}, pkgerrors.Errorf("Existing cluster provider network intents must be terminated before creating: " + cluster)
	case state.StateEnum.Instantiated:
		return ProviderNet{}, pkgerrors.Errorf("Cluster is in an invalid state: " + cluster + " " + state.StateEnum.Instantiated)
	default:
		return ProviderNet{}, pkgerrors.Errorf("Cluster is in an invalid state: " + cluster + " " + stateVal)
	}

	//Construct key and tag to select the entry
	key := ProviderNetKey{
		ClusterProviderName: clusterProvider,
		ClusterName:         cluster,
		ProviderNetName:     p.Metadata.Name,
	}

	//Check if this ProviderNet already exists
	_, err = v.GetProviderNet(p.Metadata.Name, clusterProvider, cluster)
	if err == nil && !exists {
		return ProviderNet{}, pkgerrors.New("ProviderNet already exists")
	}

	err = db.DBconn.Insert(v.db.StoreName, key, nil, v.db.TagMeta, p)
	if err != nil {
		return ProviderNet{}, pkgerrors.Wrap(err, "Creating DB Entry")
	}

	return p, nil
}

// GetProviderNet returns the ProviderNet for corresponding name
func (v *ProviderNetClient) GetProviderNet(name, clusterProvider, cluster string) (ProviderNet, error) {

	//Construct key and tag to select the entry
	key := ProviderNetKey{
		ClusterProviderName: clusterProvider,
		ClusterName:         cluster,
		ProviderNetName:     name,
	}

	value, err := db.DBconn.Find(v.db.StoreName, key, v.db.TagMeta)
	if err != nil {
		return ProviderNet{}, pkgerrors.Wrap(err, "db Find error")
	}

	//value is a byte array
	if value != nil {
		cp := ProviderNet{}
		err = db.DBconn.Unmarshal(value[0], &cp)
		if err != nil {
			return ProviderNet{}, pkgerrors.Wrap(err, "Unmarshalling Value")
		}
		return cp, nil
	}

	return ProviderNet{}, pkgerrors.New("Error getting ProviderNet")
}

// GetProviderNetList returns all of the ProviderNet for corresponding name
func (v *ProviderNetClient) GetProviderNets(clusterProvider, cluster string) ([]ProviderNet, error) {

	//Construct key and tag to select the entry
	key := ProviderNetKey{
		ClusterProviderName: clusterProvider,
		ClusterName:         cluster,
		ProviderNetName:     "",
	}

	var resp []ProviderNet
	values, err := db.DBconn.Find(v.db.StoreName, key, v.db.TagMeta)
	if err != nil {
		return []ProviderNet{}, pkgerrors.Wrap(err, "db Find error")
	}

	for _, value := range values {
		cp := ProviderNet{}
		err = db.DBconn.Unmarshal(value, &cp)
		if err != nil {
			return []ProviderNet{}, pkgerrors.Wrap(err, "Unmarshalling Value")
		}
		resp = append(resp, cp)
	}

	return resp, nil
}

// Delete the  ProviderNet from database
func (v *ProviderNetClient) DeleteProviderNet(name, clusterProvider, cluster string) error {
	// verify cluster is in a state where provider network intent can be deleted
	s, err := clusterPkg.NewClusterClient().GetClusterState(clusterProvider, cluster)
	if err != nil {
		return pkgerrors.New("Unable to find the cluster")
	}
	stateVal, err := state.GetCurrentStateFromStateInfo(s)
	if err != nil {
		return pkgerrors.Errorf("Error getting current state from Cluster stateInfo: " + cluster)
	}
	switch stateVal {
	case state.StateEnum.Approved:
		return pkgerrors.Errorf("Cluster is in an invalid state: " + cluster + " " + state.StateEnum.Approved)
	case state.StateEnum.Terminated:
		break
	case state.StateEnum.TerminateStopped:
		break
	case state.StateEnum.Created:
		break
	case state.StateEnum.Applied:
		return pkgerrors.Errorf("Cluster provider network intents must be terminated before deleting: " + cluster)
	case state.StateEnum.InstantiateStopped:
		return pkgerrors.Errorf("Cluster provider network intents must be terminated before deleting: " + cluster)
	case state.StateEnum.Instantiated:
		return pkgerrors.Errorf("Cluster is in an invalid state: " + cluster + " " + state.StateEnum.Instantiated)
	default:
		return pkgerrors.Errorf("Cluster is in an invalid state: " + cluster + " " + stateVal)
	}

	//Construct key and tag to select the entry
	key := ProviderNetKey{
		ClusterProviderName: clusterProvider,
		ClusterName:         cluster,
		ProviderNetName:     name,
	}

	err = db.DBconn.Remove(v.db.StoreName, key)
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
