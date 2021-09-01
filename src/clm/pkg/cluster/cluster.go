// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2020 Intel Corporation

package cluster

import (
	"strings"
	"time"

	"github.com/open-ness/EMCO/src/orchestrator/pkg/appcontext"
	"github.com/open-ness/EMCO/src/orchestrator/pkg/infra/db"
	log "github.com/open-ness/EMCO/src/orchestrator/pkg/infra/logutils"
	mtypes "github.com/open-ness/EMCO/src/orchestrator/pkg/module/types"
	"github.com/open-ness/EMCO/src/orchestrator/pkg/state"
	rsync "github.com/open-ness/EMCO/src/rsync/pkg/db"
	pkgerrors "github.com/pkg/errors"

	clmController "github.com/open-ness/EMCO/src/clm/pkg/controller"
	clmcontrollerpb "github.com/open-ness/EMCO/src/clm/pkg/grpc/controller-eventchannel"

	clmcontrollereventchannelclient "github.com/open-ness/EMCO/src/clm/pkg/grpc/controllereventchannelclient"
)

type clientDbInfo struct {
	storeName string // name of the mongodb collection to use for client documents
	tagMeta   string // attribute key name for the json data of a client document
	tagState  string // attribute key name for StateInfo object in the cluster
}

// ClusterProvider contains the parameters needed for ClusterProviders
type ClusterProvider struct {
	Metadata mtypes.Metadata `json:"metadata"`
}

type Cluster struct {
	Metadata mtypes.Metadata `json:"metadata"`
}

type ClusterWithLabels struct {
	Metadata mtypes.Metadata `json:"metadata"`
	Labels   []ClusterLabel  `json:"labels"`
}

type ClusterContent struct {
	Kubeconfig string `json:"kubeconfig"`
}

type ClusterLabel struct {
	LabelName string `json:"label-name"`
}

type ClusterKvPairs struct {
	Metadata mtypes.Metadata `json:"metadata"`
	Spec     ClusterKvSpec   `json:"spec"`
}

type ClusterKvSpec struct {
	Kv []map[string]interface{} `json:"kv"`
}

// ClusterProviderKey is the key structure that is used in the database
type ClusterProviderKey struct {
	ClusterProviderName string `json:"provider"`
}

// ClusterKey is the key structure that is used in the database
type ClusterKey struct {
	ClusterProviderName string `json:"provider"`
	ClusterName         string `json:"cluster"`
}

// ClusterLabelKey is the key structure that is used in the database
type ClusterLabelKey struct {
	ClusterProviderName string `json:"provider"`
	ClusterName         string `json:"cluster"`
	ClusterLabelName    string `json:"label"`
}

// LabelKey is the key structure that is used in the database
type LabelKey struct {
	ClusterProviderName string `json:"provider"`
	ClusterLabelName    string `json:"label"`
}

// ClusterKvPairsKey is the key structure that is used in the database
type ClusterKvPairsKey struct {
	ClusterProviderName string `json:"provider"`
	ClusterName         string `json:"cluster"`
	ClusterKvPairsName  string `json:"kvname"`
}

const SEPARATOR = "+"
const CONTEXT_CLUSTER_APP = "network-intents"
const CONTEXT_CLUSTER_RESOURCE = "network-intents"

// ClusterManager is an interface exposes the Cluster functionality
type ClusterManager interface {
	CreateClusterProvider(pr ClusterProvider, exists bool) (ClusterProvider, error)
	GetClusterProvider(name string) (ClusterProvider, error)
	GetClusterProviders() ([]ClusterProvider, error)
	DeleteClusterProvider(name string) error
	CreateCluster(provider string, pr Cluster, qr ClusterContent) (Cluster, error)
	GetCluster(provider, name string) (Cluster, error)
	GetClusterContent(provider, name string) (ClusterContent, error)
	GetClusterState(provider, name string) (state.StateInfo, error)
	GetClusters(provider string) ([]Cluster, error)
	GetClustersWithLabel(provider, label string) ([]string, error)
	GetAllClustersAndLabels(provider string) ([]ClusterWithLabels, error)
	DeleteCluster(provider, name string) error
	CreateClusterLabel(provider, cluster string, pr ClusterLabel, exists bool) (ClusterLabel, error)
	GetClusterLabel(provider, cluster, label string) (ClusterLabel, error)
	GetClusterLabels(provider, cluster string) ([]ClusterLabel, error)
	DeleteClusterLabel(provider, cluster, label string) error
	CreateClusterKvPairs(provider, cluster string, pr ClusterKvPairs, exists bool) (ClusterKvPairs, error)
	GetClusterKvPairs(provider, cluster, kvpair string) (ClusterKvPairs, error)
	GetClusterKvPairsValue(provider, cluster, kvpair, kvkey string) (interface{}, error)
	GetAllClusterKvPairs(provider, cluster string) ([]ClusterKvPairs, error)
	DeleteClusterKvPairs(provider, cluster, kvpair string) error
}

// ClusterClient implements the Manager
// It will also be used to maintain some localized state
type ClusterClient struct {
	db clientDbInfo
}

// NewClusterClient returns an instance of the ClusterClient
// which implements the Manager
func NewClusterClient() *ClusterClient {
	return &ClusterClient{
		db: clientDbInfo{
			storeName: "cluster",
			tagMeta:   "clustermetadata",
			tagState:  "stateInfo",
		},
	}
}

// CreateClusterProvider - create a new Cluster Provider
func (v *ClusterClient) CreateClusterProvider(p ClusterProvider, exists bool) (ClusterProvider, error) {

	//Construct key and tag to select the entry
	key := ClusterProviderKey{
		ClusterProviderName: p.Metadata.Name,
	}

	//Check if this ClusterProvider already exists
	_, err := v.GetClusterProvider(p.Metadata.Name)
	if err == nil && !exists {
		return ClusterProvider{}, pkgerrors.New("ClusterProvider already exists")
	}

	err = db.DBconn.Insert(v.db.storeName, key, nil, v.db.tagMeta, p)
	if err != nil {
		return ClusterProvider{}, pkgerrors.Wrap(err, "Creating DB Entry")
	}

	return p, nil
}

// GetClusterProvider returns the ClusterProvider for corresponding name
func (v *ClusterClient) GetClusterProvider(name string) (ClusterProvider, error) {

	//Construct key and tag to select the entry
	key := ClusterProviderKey{
		ClusterProviderName: name,
	}

	value, err := db.DBconn.Find(v.db.storeName, key, v.db.tagMeta)
	if err != nil {
		return ClusterProvider{}, pkgerrors.Wrap(err, "db Find error")
	} else if len(value) == 0 {
		return ClusterProvider{}, pkgerrors.New("Cluster provider not found")
	}

	//value is a byte array
	if value != nil {
		cp := ClusterProvider{}
		err = db.DBconn.Unmarshal(value[0], &cp)
		if err != nil {
			return ClusterProvider{}, pkgerrors.Wrap(err, "db Unmarshal error")
		}
		return cp, nil
	}

	return ClusterProvider{}, pkgerrors.New("Error getting ClusterProvider")
}

// GetClusterProviderList returns all of the ClusterProvider for corresponding name
func (v *ClusterClient) GetClusterProviders() ([]ClusterProvider, error) {

	//Construct key and tag to select the entry
	key := ClusterProviderKey{
		ClusterProviderName: "",
	}

	values, err := db.DBconn.Find(v.db.storeName, key, v.db.tagMeta)
	if err != nil {
		return []ClusterProvider{}, pkgerrors.Wrap(err, "db Find error")
	}

	resp := make([]ClusterProvider, 0)
	for _, value := range values {
		cp := ClusterProvider{}
		err = db.DBconn.Unmarshal(value, &cp)
		if err != nil {
			return []ClusterProvider{}, pkgerrors.Wrap(err, "db Unmarshal error")
		}
		resp = append(resp, cp)
	}

	return resp, nil
}

// DeleteClusterProvider the  ClusterProvider from database
func (v *ClusterClient) DeleteClusterProvider(name string) error {

	//Construct key and tag to select the entry
	key := ClusterProviderKey{
		ClusterProviderName: name,
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

// CreateCluster - create a new Cluster for a cluster-provider
func (v *ClusterClient) CreateCluster(provider string, p Cluster, q ClusterContent) (Cluster, error) {

	//Construct key and tag to select the entry
	key := ClusterKey{
		ClusterProviderName: provider,
		ClusterName:         p.Metadata.Name,
	}

	//Verify ClusterProvider already exists
	_, err := v.GetClusterProvider(provider)
	if err != nil {
		return Cluster{}, pkgerrors.New("ClusterProvider does not exist")
	}

	//Check if this Cluster already exists
	_, err = v.GetCluster(provider, p.Metadata.Name)
	if err == nil {
		return Cluster{}, pkgerrors.New("Cluster already exists")
	}

	err = db.DBconn.Insert(v.db.storeName, key, nil, v.db.tagMeta, p)
	if err != nil {
		return Cluster{}, pkgerrors.Wrap(err, "Creating DB Entry")
	}

	// Add the stateInfo record
	s := state.StateInfo{}
	a := state.ActionEntry{
		State:     state.StateEnum.Created,
		ContextId: "",
		TimeStamp: time.Now(),
	}
	s.Actions = append(s.Actions, a)

	err = db.DBconn.Insert(v.db.storeName, key, nil, v.db.tagState, s)
	if err != nil {
		return Cluster{}, pkgerrors.Wrap(err, "Creating cluster StateInfo")
	}

	ccc := rsync.NewCloudConfigClient()

	_, err = ccc.CreateCloudConfig(provider, p.Metadata.Name, "0", "default", q.Kubeconfig)
	if err != nil {
		return Cluster{}, pkgerrors.Wrap(err, "Error creating cloud config")
	}

	// Loop through CLM controllers and publish CLUSTER_CREATE event
	client := clmController.NewControllerClient()
	ctrls, _ := client.GetControllers()
	for _, c := range ctrls {
		log.Info("CLM CreateController .. controller info.", log.Fields{"provider-name": provider, "cluster-name": p.Metadata.Name, "Controller": c})
		err = clmcontrollereventchannelclient.SendControllerEvent(provider, p.Metadata.Name, clmcontrollerpb.ClmControllerEventType_CLUSTER_CREATED, c)
		if err != nil {
			log.Error("CLM CreateController .. Failed publishing event to clm-controller.", log.Fields{"provider-name": provider, "cluster-name": p.Metadata.Name, "Controller": c})
			return Cluster{}, pkgerrors.Wrapf(err, "CLM failed publishing event to clm-controller[%v]", c.Metadata.Name)
		}
	}

	return p, nil
}

// GetCluster returns the Cluster for corresponding provider and name
func (v *ClusterClient) GetCluster(provider, name string) (Cluster, error) {
	//Construct key and tag to select the entry
	key := ClusterKey{
		ClusterProviderName: provider,
		ClusterName:         name,
	}

	value, err := db.DBconn.Find(v.db.storeName, key, v.db.tagMeta)
	if err != nil {
		return Cluster{}, pkgerrors.Wrap(err, "db Find error")
	} else if len(value) == 0 {
		return Cluster{}, pkgerrors.New("Cluster not found")
	}

	//value is a byte array
	if value != nil {
		cl := Cluster{}
		err = db.DBconn.Unmarshal(value[0], &cl)
		if err != nil {
			return Cluster{}, pkgerrors.Wrap(err, "Unmarshalling Value")
		}
		return cl, nil
	}

	return Cluster{}, pkgerrors.New("Error getting Cluster")
}

// GetClusterContent returns the ClusterContent for corresponding provider and name
func (v *ClusterClient) GetClusterContent(provider, name string) (ClusterContent, error) {

	// Fetch the kubeconfig from rsync according to new workflow
	ccc := rsync.NewCloudConfigClient()

	cconfig, err := ccc.GetCloudConfig(provider, name, "0", "")
	if err != nil {
		if strings.Contains(err.Error(), "Finding CloudConfig failed") {
			return ClusterContent{}, pkgerrors.Wrap(err, "GetCloudConfig error - not found")
		} else {
			return ClusterContent{}, pkgerrors.Wrap(err, "GetCloudConfig error - general")
		}
	}

	ccontent := ClusterContent{}
	ccontent.Kubeconfig = cconfig.Config

	return ccontent, nil
}

// GetClusterState returns the StateInfo structure for corresponding cluster provider and cluster
func (v *ClusterClient) GetClusterState(provider, name string) (state.StateInfo, error) {
	//Construct key and tag to select the entry
	key := ClusterKey{
		ClusterProviderName: provider,
		ClusterName:         name,
	}

	result, err := db.DBconn.Find(v.db.storeName, key, v.db.tagState)
	if err != nil {
		return state.StateInfo{}, pkgerrors.Wrap(err, "db Find error")
	} else if len(result) == 0 {
		return state.StateInfo{}, pkgerrors.New("Cluster StateInfo not found")
	}

	if result != nil {
		s := state.StateInfo{}
		err = db.DBconn.Unmarshal(result[0], &s)
		if err != nil {
			return state.StateInfo{}, pkgerrors.Wrap(err, "Unmarshalling Cluster StateInfo")
		}
		return s, nil
	}

	return state.StateInfo{}, pkgerrors.New("Error getting Cluster StateInfo")
}

// GetClusters returns all the Clusters for corresponding provider
func (v *ClusterClient) GetClusters(provider string) ([]Cluster, error) {
	//Construct key and tag to select the entry
	key := ClusterKey{
		ClusterProviderName: provider,
		ClusterName:         "",
	}

	//Verify Cluster provider exists
	_, err := v.GetClusterProvider(provider)
	if err != nil {
		return []Cluster{}, err
	}

	values, err := db.DBconn.Find(v.db.storeName, key, v.db.tagMeta)
	if err != nil {
		return []Cluster{}, pkgerrors.Wrap(err, "db Find error")
	}

	resp := make([]Cluster, 0)
	for _, value := range values {
		cp := Cluster{}
		err = db.DBconn.Unmarshal(value, &cp)
		if err != nil {
			return []Cluster{}, pkgerrors.Wrap(err, "Unmarshalling Value")
		}
		resp = append(resp, cp)
	}

	return resp, nil
}

// GetAllClustersAndLabels returns all the the clusters and their labels
func (v *ClusterClient) GetAllClustersAndLabels(provider string) ([]ClusterWithLabels, error) {

	// Get All clusters
	cl, err := v.GetClusters(provider)
	if err != nil {
		return []ClusterWithLabels{}, err
	}

	resp := make([]ClusterWithLabels, len(cl))

	// Get all cluster labels
	for k, value := range cl {
		resp[k].Metadata = value.Metadata
		resp[k].Labels, err = v.GetClusterLabels(provider, value.Metadata.Name)
		if err != nil {
			return []ClusterWithLabels{}, err
		}
	}
	return resp, nil
}

// GetClustersWithLabel returns all the Clusters with Labels for provider
// Support Query like /cluster-providers/{Provider}/clusters?label={label}
func (v *ClusterClient) GetClustersWithLabel(provider, label string) ([]string, error) {
	//Construct key and tag to select the entry
	key := LabelKey{
		ClusterProviderName: provider,
		ClusterLabelName:    label,
	}

	//Verify Cluster provider exists
	_, err := v.GetClusterProvider(provider)
	if err != nil {
		return []string{}, err
	}

	values, err := db.DBconn.Find(v.db.storeName, key, "cluster")
	if err != nil {
		return []string{}, pkgerrors.Wrap(err, "db Find error")
	}

	resp := make([]string, 0)
	for _, value := range values {
		cp := string(value)
		resp = append(resp, cp)
	}

	return resp, nil
}

// DeleteCluster the  Cluster from database
func (v *ClusterClient) DeleteCluster(provider, name string) error {
	//Construct key and tag to select the entry
	key := ClusterKey{
		ClusterProviderName: provider,
		ClusterName:         name,
	}
	s, err := v.GetClusterState(provider, name)
	if err != nil {
		// If the StateInfo cannot be found, then a proper cluster record is not present.
		// Call the DB delete to clean up any errant record without a StateInfo element that may exist.
		err = db.DBconn.Remove(v.db.storeName, key)
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

	stateVal, err := state.GetCurrentStateFromStateInfo(s)
	if err != nil {
		return pkgerrors.Errorf("Error getting current state from Cluster stateInfo: " + name)
	}

	if stateVal == state.StateEnum.Applied || stateVal == state.StateEnum.InstantiateStopped {
		return pkgerrors.Errorf("Cluster network intents must be terminated before it can be deleted " + name)
	}

	// remove the app contexts associated with this cluster
	if stateVal == state.StateEnum.Terminated || stateVal == state.StateEnum.TerminateStopped {
		// Verify that the appcontext has completed terminating
		ctxid := state.GetLastContextIdFromStateInfo(s)
		acStatus, err := state.GetAppContextStatus(ctxid)
		if err == nil &&
			!(acStatus.Status == appcontext.AppContextStatusEnum.Terminated || acStatus.Status == appcontext.AppContextStatusEnum.TerminateFailed) {
			return pkgerrors.Errorf("Network intents for cluster have not completed terminating " + name)
		}

		for _, id := range state.GetContextIdsFromStateInfo(s) {
			context, err := state.GetAppContextFromId(id)
			if err != nil {
				return pkgerrors.Wrap(err, "Error getting appcontext from Cluster StateInfo")
			}
			err = context.DeleteCompositeApp()
			if err != nil {
				return pkgerrors.Wrap(err, "Error deleting appcontext for Cluster")
			}
		}
	}

	err = db.DBconn.Remove(v.db.storeName, key)
	if err != nil {
		return pkgerrors.Wrap(err, "Delete Cluster Entry;")
	}

	ccc := rsync.NewCloudConfigClient()

	err = ccc.DeleteCloudConfig(provider, name, "0", "default")
	if err != nil {
		return pkgerrors.Wrap(err, "Error deleting cloud config")
	}

	// Loop through CLM controllers and publish CLUSTER_DELETE event
	client := clmController.NewControllerClient()
	vals, _ := client.GetControllers()
	for _, v := range vals {
		log.Info("DeleteCluster .. controller info.", log.Fields{"provider-name": provider, "cluster-name": name, "Controller": v})
		err = clmcontrollereventchannelclient.SendControllerEvent(provider, name, clmcontrollerpb.ClmControllerEventType_CLUSTER_DELETED, v)
		if err != nil {
			log.Error("DeleteCluster .. Failed publishing event to controller.", log.Fields{"provider-name": provider, "cluster-name": name, "Controller": v})
		}
	}

	return nil
}

// CreateClusterLabel - create a new Cluster Label mongo document for a cluster-provider/cluster
func (v *ClusterClient) CreateClusterLabel(provider string, cluster string, p ClusterLabel, exists bool) (ClusterLabel, error) {
	//Construct key and tag to select the entry
	key := ClusterLabelKey{
		ClusterProviderName: provider,
		ClusterName:         cluster,
		ClusterLabelName:    p.LabelName,
	}

	//Verify Cluster already exists
	_, err := v.GetCluster(provider, cluster)
	if err != nil {
		return ClusterLabel{}, pkgerrors.New("Cluster does not exist")
	}

	//Check if this ClusterLabel already exists
	_, err = v.GetClusterLabel(provider, cluster, p.LabelName)
	if err == nil && !exists {
		return ClusterLabel{}, pkgerrors.New("Cluster Label already exists")
	}

	err = db.DBconn.Insert(v.db.storeName, key, nil, v.db.tagMeta, p)
	if err != nil {
		return ClusterLabel{}, pkgerrors.Wrap(err, "Creating DB Entry")
	}

	return p, nil
}

// GetClusterLabel returns the Cluster for corresponding provider, cluster and label
func (v *ClusterClient) GetClusterLabel(provider, cluster, label string) (ClusterLabel, error) {
	//Construct key and tag to select the entry
	key := ClusterLabelKey{
		ClusterProviderName: provider,
		ClusterName:         cluster,
		ClusterLabelName:    label,
	}

	value, err := db.DBconn.Find(v.db.storeName, key, v.db.tagMeta)
	if err != nil {
		return ClusterLabel{}, pkgerrors.Wrap(err, "db Find error")
	} else if len(value) == 0 {
		return ClusterLabel{}, pkgerrors.New("Cluster label not found")
	}

	//value is a byte array
	if value != nil {
		cl := ClusterLabel{}
		err = db.DBconn.Unmarshal(value[0], &cl)
		if err != nil {
			return ClusterLabel{}, pkgerrors.Wrap(err, "Unmarshalling Value")
		}
		return cl, nil
	}

	return ClusterLabel{}, pkgerrors.New("Error getting Cluster")
}

// GetClusterLabels returns the Cluster Labels for corresponding provider and cluster
func (v *ClusterClient) GetClusterLabels(provider, cluster string) ([]ClusterLabel, error) {
	// Construct key and tag to select the entry
	key := ClusterLabelKey{
		ClusterProviderName: provider,
		ClusterName:         cluster,
		ClusterLabelName:    "",
	}

	// Verify Cluster already exists
	_, err := v.GetCluster(provider, cluster)
	if err != nil {
		return []ClusterLabel{}, err
	}

	values, err := db.DBconn.Find(v.db.storeName, key, v.db.tagMeta)
	if err != nil {
		return []ClusterLabel{}, pkgerrors.Wrap(err, "db Find error")
	}

	resp := make([]ClusterLabel, 0)
	for _, value := range values {
		cp := ClusterLabel{}
		err = db.DBconn.Unmarshal(value, &cp)
		if err != nil {
			return []ClusterLabel{}, pkgerrors.Wrap(err, "Unmarshalling Value")
		}
		resp = append(resp, cp)
	}

	return resp, nil
}

// DeleteClusterLabel ... Delete the Cluster Label from database
func (v *ClusterClient) DeleteClusterLabel(provider, cluster, label string) error {
	//Construct key and tag to select the entry
	key := ClusterLabelKey{
		ClusterProviderName: provider,
		ClusterName:         cluster,
		ClusterLabelName:    label,
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

// CreateClusterKvPairs - Create a New Cluster KV pairs document
func (v *ClusterClient) CreateClusterKvPairs(provider string, cluster string, p ClusterKvPairs, exists bool) (ClusterKvPairs, error) {
	key := ClusterKvPairsKey{
		ClusterProviderName: provider,
		ClusterName:         cluster,
		ClusterKvPairsName:  p.Metadata.Name,
	}

	//Verify Cluster already exists
	_, err := v.GetCluster(provider, cluster)
	if err != nil {
		return ClusterKvPairs{}, pkgerrors.New("Cluster does not exist")
	}

	//Check if this ClusterKvPairs already exists
	_, err = v.GetClusterKvPairs(provider, cluster, p.Metadata.Name)
	if err == nil && !exists {
		return ClusterKvPairs{}, pkgerrors.New("Cluster KV Pair already exists")
	}

	err = db.DBconn.Insert(v.db.storeName, key, nil, v.db.tagMeta, p)
	if err != nil {
		return ClusterKvPairs{}, pkgerrors.Wrap(err, "Creating DB Entry")
	}

	return p, nil
}

// GetClusterKvPairs returns the Cluster KeyValue pair for corresponding provider, cluster and KV pair name
func (v *ClusterClient) GetClusterKvPairs(provider, cluster, kvpair string) (ClusterKvPairs, error) {
	//Construct key and tag to select entry
	key := ClusterKvPairsKey{
		ClusterProviderName: provider,
		ClusterName:         cluster,
		ClusterKvPairsName:  kvpair,
	}

	value, err := db.DBconn.Find(v.db.storeName, key, v.db.tagMeta)
	if err != nil {
		return ClusterKvPairs{}, pkgerrors.Wrap(err, "db Find error")
	} else if len(value) == 0 {
		return ClusterKvPairs{}, pkgerrors.New("Cluster key value pair not found")
	}

	//value is a byte array
	if value != nil {
		ckvp := ClusterKvPairs{}
		err = db.DBconn.Unmarshal(value[0], &ckvp)
		if err != nil {
			return ClusterKvPairs{}, pkgerrors.Wrap(err, "Unmarshalling Value")
		}
		return ckvp, nil
	}

	return ClusterKvPairs{}, pkgerrors.New("Error getting Cluster KV pairs")
}

// GetClusterKvPairsValue returns the value of the key from the corresponding provider, cluster and KV pair name
func (v *ClusterClient) GetClusterKvPairsValue(provider, cluster, kvpair, kvkey string) (interface{}, error) {
	//Construct key and tag to select entry
	key := ClusterKvPairsKey{
		ClusterProviderName: provider,
		ClusterName:         cluster,
		ClusterKvPairsName:  kvpair,
	}

	value, err := db.DBconn.Find(v.db.storeName, key, v.db.tagMeta)
	if err != nil {
		return ClusterKvPairs{}, pkgerrors.Wrap(err, "db Find error")
	} else if len(value) == 0 {
		return Cluster{}, pkgerrors.New("Cluster key value pair not found")
	}

	//value is a byte array
	if value != nil {
		ckvp := ClusterKvPairs{}
		err = db.DBconn.Unmarshal(value[0], &ckvp)
		if err != nil {
			return nil, pkgerrors.Wrap(err, "Unmarshalling Value")
		}

		for _, kvmap := range ckvp.Spec.Kv {
			if val, ok := kvmap[kvkey]; ok {
				return struct {
					Value interface{} `json:"value"`
				}{Value: val}, nil
			}
		}
		return nil, pkgerrors.New("Cluster KV pair key value not found")
	}

	return nil, pkgerrors.New("Error getting Cluster KV pair")
}

// GetAllClusterKvPairs returns the Cluster Kv Pairs for corresponding provider and cluster
func (v *ClusterClient) GetAllClusterKvPairs(provider, cluster string) ([]ClusterKvPairs, error) {
	//Construct key and tag to select the entry
	key := ClusterKvPairsKey{
		ClusterProviderName: provider,
		ClusterName:         cluster,
		ClusterKvPairsName:  "",
	}

	// Verify Cluster exists
	_, err := v.GetCluster(provider, cluster)
	if err != nil {
		return []ClusterKvPairs{}, err
	}

	values, err := db.DBconn.Find(v.db.storeName, key, v.db.tagMeta)
	if err != nil {
		return []ClusterKvPairs{}, pkgerrors.Wrap(err, "db Find error")
	}

	resp := make([]ClusterKvPairs, 0)
	for _, value := range values {
		cp := ClusterKvPairs{}
		err = db.DBconn.Unmarshal(value, &cp)
		if err != nil {
			return []ClusterKvPairs{}, pkgerrors.Wrap(err, "Unmarshalling Value")
		}
		resp = append(resp, cp)
	}

	return resp, nil
}

// DeleteClusterKvPairs the  ClusterKvPairs from database
func (v *ClusterClient) DeleteClusterKvPairs(provider, cluster, kvpair string) error {
	//Construct key and tag to select entry
	key := ClusterKvPairsKey{
		ClusterProviderName: provider,
		ClusterName:         cluster,
		ClusterKvPairsName:  kvpair,
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
