// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2020 Intel Corporation

package context

import (
	"fmt"
	"reflect"
	"strings"
	"sort"
	"sync"

	"github.com/open-ness/EMCO/src/orchestrator/pkg/appcontext"
	. "github.com/open-ness/EMCO/src/rsync/pkg/types"
	pkgerrors "github.com/pkg/errors"
	//"time"
)
// Match stores information about resources applied in clusters
type Match struct {
	// Collects all resources that are deleted
	DeleteMatchList sync.Map
	// Collects all resources that are applied
	ApplyMatchList  sync.Map
	// Collects all resources that are currently applied on the cluster
	ResourceList    sync.Map
}
// MatchList to collect resources
var MatchList Match

// MockConnector mocks connector interface
type MockConnector struct {
	sync.Mutex
	cid     string
	Clients *sync.Map
}

func (c *MockConnector) GetClientInternal(cluster string, level string, namespace string) (ClientProvider, error) {
	c.Lock()
	defer c.Unlock()
	if c.Clients == nil {
		c.Clients = new(sync.Map)
	}
	_, ok := c.Clients.Load(cluster)
	if !ok {
		m := MockClient{cluster: cluster, retryCounter: 0, deletedCounter: 0}
		m.lock = new(sync.Mutex)
		c.Clients.Store(cluster, m)
	}
	m, _ := c.Clients.Load(cluster)
	n := m.(MockClient)
	return &n, nil
}

func (c *MockConnector) RemoveClient() {

}
func (c *MockConnector) StartClusterWatcher(cluster string) error {
	return nil

}
func (c *MockConnector) GetStatusCR(label string) ([]byte, error) {
	return nil, nil
}
func (c *MockConnector) Init(id interface{}) error {
	c.cid = fmt.Sprintf("%v", id)
	MatchList.DeleteMatchList = sync.Map{}
	MatchList.ApplyMatchList = sync.Map{}
	MatchList.ResourceList = sync.Map{}
	return nil
}
// MockClient mocks client
type MockClient struct {
	lock           *sync.Mutex
	cluster        string
	retryCounter   int
	deletedCounter int
	applyCounter   int
}
// Apply Collects resources applied to cluster
func (m *MockClient) Apply(content []byte) error {
	m.lock.Lock()
	defer m.lock.Unlock()
	m.applyCounter = m.applyCounter + 1
	// Simulate delay
	//time.Sleep(1 * time.Millisecond)
	if len(content) <= 0 {
		return nil
	}
	i, ok := MatchList.ApplyMatchList.Load(m.cluster)
	var str string
	if !ok {
		str = string(content)
	} else {
		str = fmt.Sprintf("%v", i) + "," + string(content)
	}
	MatchList.ApplyMatchList.Store(m.cluster, str)

	i, ok = MatchList.ResourceList.Load(m.cluster)
	if !ok {
		str = string(content)
	} else {
		x := fmt.Sprintf("%v", i)
		if x != "" {
			str = fmt.Sprintf("%v", i) + "," + string(content)
		} else {
			str = string(content)
		}
	}
	MatchList.ResourceList.Store(m.cluster, str)

	return nil
}
// Delete Collects resources deleted from cluster
func (m *MockClient) Delete(content []byte) error {
	m.lock.Lock()
	defer m.lock.Unlock()
	m.deletedCounter = m.deletedCounter + 1
	if len(content) <= 0 {
		return nil
	}
	i, ok := MatchList.DeleteMatchList.Load(m.cluster)
	var str string
	if !ok {
		str = string(content)
	} else {
		str = fmt.Sprintf("%v", i) + "," + string(content)
	}
	MatchList.DeleteMatchList.Store(m.cluster, str)

	// Remove the resource from resourcre list
	i, ok = MatchList.ResourceList.Load(m.cluster)
	if !ok {
		fmt.Println("Deleting resource not applied on cluster", m.cluster)
		return pkgerrors.Errorf("Deleting resource not applied on cluster " + m.cluster)
	} else {
		// Delete it from the string
		a := strings.Split(fmt.Sprintf("%v", i), ",")
		for idx, v := range a {
			if v == string(content) {
				a = append(a[:idx], a[idx+1:]...)
				break
			}
		}
		if len(a) > 0 {
			str = strings.Join(a, ",")
		} else {
			str = ""
		}
		MatchList.ResourceList.Store(m.cluster, str)
	}
	return nil
}
func (m *MockClient) Approve(name string, sa []byte) error {
	return nil
}
func (m *MockClient) Get(gvkRes []byte, namespace string) ([]byte, error) {
	b := []byte("test")
	return b, nil
}
func (m *MockClient) IsReachable() error {
	if m.cluster == "provider1+cluster1" {
		if m.retryCounter < 0 {
			fmt.Println("Counter: ", m.retryCounter)
			m.retryCounter = m.retryCounter + 1
			return pkgerrors.Errorf("Unreachable: " + m.cluster)
		}
	}
	return nil
}
func (m *MockClient) TagResource(res []byte, label string) ([]byte, error) {
	return res, nil
}

func LoadMap(str string) map[string]string {
	m := make(map[string]string)
	if str == "apply" {
		MatchList.ApplyMatchList.Range(func(k, v interface{}) bool {
			m[fmt.Sprint(k)] = v.(string)
			return true
		})
	} else if str == "delete" {
		MatchList.DeleteMatchList.Range(func(k, v interface{}) bool {
			m[fmt.Sprint(k)] = v.(string)
			return true
		})
	} else if str == "resource" {
		MatchList.ResourceList.Range(func(k, v interface{}) bool {
			m[fmt.Sprint(k)] = v.(string)
			return true
		})
	}
	return m
}

func CompareMaps(m, n map[string]string) bool {
	var m1, n1 map[string][]string
	m1 = make(map[string][]string)
	n1 = make(map[string][]string)
	for k, v := range m {
		a := strings.Split(v, ",")
		sort.Strings(a)
		m1[k] = a
	}
	for k, v := range n {
		a := strings.Split(v, ",")
		sort.Strings(a)
		n1[k] = a
	}
	return reflect.DeepEqual(m1, n1)
}

func GetAppContextStatus(cid interface{}, key string) (string, error) {
	//var acStatus appcontext.AppContextStatus = appcontext.AppContextStatus{}
	ac := appcontext.AppContext{}
	_, err := ac.LoadAppContext(cid)
	if err != nil {
		return "", err
	}
	hc, err := ac.GetCompositeAppHandle()
	if err != nil {
		return "", err
	}
	dsh, err := ac.GetLevelHandle(hc, key)
	if dsh != nil {
		v, err := ac.GetValue(dsh)
		if err != nil {
			return "", err
		}
		str := fmt.Sprintf("%v", v)
		return str, nil
	}
	return "", err
}

func UpdateAppContextFlag(cid interface{}, key string, b bool) error {
	ac := appcontext.AppContext{}
	_, err := ac.LoadAppContext(cid)
	if err != nil {
		return err
	}
	h, err := ac.GetCompositeAppHandle()
	if err != nil {

		return err
	}
	sh, err := ac.GetLevelHandle(h, key)
	if sh == nil {
		_, err = ac.AddLevelValue(h, key, b)
	} else {
		err = ac.UpdateValue(sh, b)
	}
	return err
}