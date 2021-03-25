// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2020 Intel Corporation

package module

import (
	"github.com/open-ness/EMCO/src/orchestrator/pkg/infra/db"
	pkgerrors "github.com/pkg/errors"
)

// Quota contains the parameters needed for a Quota
type Quota struct {
	MetaData QMetaDataList `json:"metadata"`
	// Specification QSpec         `json:"spec"`
	Specification map[string]string `json:"spec"`
}

// MetaData contains the parameters needed for metadata
type QMetaDataList struct {
	QuotaName   string `json:"name"`
	Description string `json:"description"`
	UserData1   string `json:"userData1"`
	UserData2   string `json:"userData2"`
}

// TODO: use QSpec fields to validate quota keys
// Spec contains the parameters needed for spec
type QSpec struct {
	LimitsCPU                   string `json:"limits.cpu"`
	LimitsMemory                string `json:"limits.memory"`
	RequestsCPU                 string `json:"requests.cpu"`
	RequestsMemory              string `json:"requests.memory"`
	RequestsStorage             string `json:"requests.storage"`
	LimitsEphemeralStorage      string `json:"limits.ephemeral.storage"`
	PersistentVolumeClaims      string `json:"persistentvolumeclaims"`
	Pods                        string `json:"pods"`
	ConfigMaps                  string `json:"configmaps"`
	ReplicationControllers      string `json:"replicationcontrollers"`
	ResourceQuotas              string `json:"resourcequotas"`
	Services                    string `json:"services"`
	ServicesLoadBalancers       string `json:"services.loadbalancers"`
	ServicesNodePorts           string `json:"services.nodeports"`
	Secrets                     string `json:"secrets"`
	CountReplicationControllers string `json:"count/replicationcontrollers"`
	CountDeploymentsApps        string `json:"count/deployments.apps"`
	CountReplicasetsApps        string `json:"count/replicasets.apps"`
	CountStatefulSets           string `json:"count/statefulsets.apps"`
	CountJobsBatch              string `json:"count/jobs.batch"`
	CountCronJobsBatch          string `json:"count/cronjobs.batch"`
	CountDeploymentsExtensions  string `json:"count/deployments.extensions"`
}

// QuotaKey is the key structure that is used in the database
type QuotaKey struct {
	Project          string `json:"project"`
	LogicalCloudName string `json:"logical-cloud-name"`
	QuotaName        string `json:"qname"`
}

// QuotaManager is an interface that exposes the connection
// functionality
type QuotaManager interface {
	CreateQuota(project, logicalCloud string, c Quota) (Quota, error)
	GetQuota(project, logicalCloud, name string) (Quota, error)
	GetAllQuotas(project, logicalCloud string) ([]Quota, error)
	DeleteQuota(project, logicalCloud, name string) error
	UpdateQuota(project, logicalCloud, name string, c Quota) (Quota, error)
}

// QuotaClient implements the QuotaManager
// It will also be used to maintain some localized state
type QuotaClient struct {
	storeName string
	tagMeta   string
}

// QuotaClient returns an instance of the QuotaClient
// which implements the QuotaManager
func NewQuotaClient() *QuotaClient {
	return &QuotaClient{
		storeName: "orchestrator",
		tagMeta:   "quota",
	}
}

// Create entry for the quota resource in the database
func (v *QuotaClient) CreateQuota(project, logicalCloud string, c Quota) (Quota, error) {

	//Construct key consisting of name
	key := QuotaKey{
		Project:          project,
		LogicalCloudName: logicalCloud,
		QuotaName:        c.MetaData.QuotaName,
	}
	lcClient := NewLogicalCloudClient()

	//Check if project exists
	err := CheckProject(project)
	if err != nil {
		return Quota{}, pkgerrors.New("Unable to find the project")
	}
	//check if logical cloud exists
	err = CheckLogicalCloud(lcClient, project, logicalCloud)
	if err != nil {
		return Quota{}, pkgerrors.New("Unable to find the logical cloud")
	}
	//Check if Logical Cloud Level 0 & then avoid creating Quotas
	lc, err := lcClient.Get(project, logicalCloud)
	if err != nil {
		return Quota{}, pkgerrors.Wrap(err, "Logical Cloud does not exist")
	}
	if lc.Specification.Level == "0" {
		return Quota{}, pkgerrors.New("Cluster Quotas not allowed for Logical Cloud Level 0")
	}
	//Check if this Quota already exists
	_, err = v.GetQuota(project, logicalCloud, c.MetaData.QuotaName)
	if err == nil {
		return Quota{}, pkgerrors.New("Quota already exists")
	}

	err = db.DBconn.Insert(v.storeName, key, nil, v.tagMeta, c)
	if err != nil {
		return Quota{}, pkgerrors.Wrap(err, "Creating DB Entry")
	}

	return c, nil
}

// Get returns Quota for corresponding quota name
func (v *QuotaClient) GetQuota(project, logicalCloud, quotaName string) (Quota, error) {

	//Construct the composite key to select the entry
	key := QuotaKey{
		Project:          project,
		LogicalCloudName: logicalCloud,
		QuotaName:        quotaName,
	}
	value, err := db.DBconn.Find(v.storeName, key, v.tagMeta)
	if err != nil {
		return Quota{}, pkgerrors.Wrap(err, "Quota")
	}

	//value is a byte array
	if value != nil {
		q := Quota{}
		err = db.DBconn.Unmarshal(value[0], &q)
		if err != nil {
			return Quota{}, pkgerrors.Wrap(err, "Unmarshaling value")
		}
		return q, nil
	}

	return Quota{}, pkgerrors.New("Cluster Quota does not exist")
}

// GetAll returns all cluster quotas in the logical cloud
func (v *QuotaClient) GetAllQuotas(project, logicalCloud string) ([]Quota, error) {
	//Construct the composite key to select the entry
	key := QuotaKey{
		Project:          project,
		LogicalCloudName: logicalCloud,
		QuotaName:        "",
	}
	var resp []Quota
	values, err := db.DBconn.Find(v.storeName, key, v.tagMeta)
	if err != nil {
		return []Quota{}, pkgerrors.Wrap(err, "Error getting all Quotas")
	}

	for _, value := range values {
		q := Quota{}
		err = db.DBconn.Unmarshal(value, &q)
		if err != nil {
			return []Quota{}, pkgerrors.Wrap(err, "Unmarshaling value")
		}
		resp = append(resp, q)
	}

	return resp, nil
}

// Delete the Quota entry from database
func (v *QuotaClient) DeleteQuota(project, logicalCloud, quotaName string) error {
	//Construct the composite key to select the entry
	key := QuotaKey{
		Project:          project,
		LogicalCloudName: logicalCloud,
		QuotaName:        quotaName,
	}
	err := db.DBconn.Remove(v.storeName, key)
	if err != nil {
		return pkgerrors.Wrap(err, "Delete Quota")
	}
	return nil
}

// Update an entry for the Quota in the database
func (v *QuotaClient) UpdateQuota(project, logicalCloud, quotaName string, c Quota) (Quota, error) {

	key := QuotaKey{
		Project:          project,
		LogicalCloudName: logicalCloud,
		QuotaName:        quotaName,
	}
	//Check quota URL name against the quota json name
	if c.MetaData.QuotaName != quotaName {
		return Quota{}, pkgerrors.New("Update Error - Quota name mismatch")
	}
	//Check if this Quota exists
	_, err := v.GetQuota(project, logicalCloud, quotaName)
	if err != nil {
		return Quota{}, pkgerrors.New("Cluster Quota does not exist")
	}
	err = db.DBconn.Insert(v.storeName, key, nil, v.tagMeta, c)
	if err != nil {
		return Quota{}, pkgerrors.Wrap(err, "Updating DB Entry")
	}
	return c, nil
}
