// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2020 Intel Corporation

package connector

import (
	"encoding/base64"
	"errors"
	"fmt"
	"os"
	"strings"
	"sync"
	types "github.com/open-ness/EMCO/src/rsync/pkg/types"
	log "github.com/open-ness/EMCO/src/orchestrator/pkg/infra/logutils"
	kubeclient "github.com/open-ness/EMCO/src/rsync/pkg/client"
	"github.com/open-ness/EMCO/src/rsync/pkg/db"
	pkgerrors "github.com/pkg/errors"
)

// IsTestKubeClient .. global variable used during unit-tests to check whether a fake kube client object has to be instantiated
var IsTestKubeClient bool = false

// Connection is for a cluster
type Connection struct {
	Cid     string
	Clients map[string]*kubeclient.Client
	sync.Mutex
}

const basePath string = "/tmp/rsync/"

// Init Connection for an app context
func (c *Connection) Init(id interface{}) error {
	log.Info("Init with interface", log.Fields{})
	c.Clients = make(map[string]*kubeclient.Client)
	c.Cid = fmt.Sprintf("%v", id)
	return nil
}

// GetKubeConfig uses the connectivity client to get the kubeconfig based on the name
// of the clustername.
func GetKubeConfig(clustername string, level string, namespace string) ([]byte, error) {
	if !strings.Contains(clustername, "+") {
		return nil, pkgerrors.New("Not a valid cluster name")
	}
	strs := strings.Split(clustername, "+")
	if len(strs) != 2 {
		return nil, pkgerrors.New("Not a valid cluster name")
	}

	ccc := db.NewCloudConfigClient()

	log.Info("Querying CloudConfig", log.Fields{"strs": strs, "level": level, "namespace": namespace})
	cconfig, err := ccc.GetCloudConfig(strs[0], strs[1], level, namespace)
	if err != nil {
		return nil, pkgerrors.New("Get kubeconfig failed")
	}
	log.Info("Successfully looked up CloudConfig", log.Fields{".Provider": cconfig.Provider, ".Cluster": cconfig.Cluster, ".Level": cconfig.Level, ".Namespace": cconfig.Namespace})

	dec, err := base64.StdEncoding.DecodeString(cconfig.Config)
	if err != nil {
		return nil, err
	}
	return dec, nil
}

// GetClient returns client for the cluster
func (c *Connection) GetClient(cluster string, level string, namespace string) (*kubeclient.Client, error) {
	c.Lock()
	defer c.Unlock()

	// Check if Fake kube client is required(it's true for unit tests)
	log.Info("GetClient .. start", log.Fields{"IsTestKubeClient": fmt.Sprintf("%v", IsTestKubeClient)})
	if IsTestKubeClient {
		return kubeclient.NewKubeFakeClient()
	}

	client, ok := c.Clients[cluster]
	if !ok {
		// Get file from DB
		dec, err := GetKubeConfig(cluster, level, namespace)
		if err != nil {
			return nil, err
		}
		var kubeConfigPath string = basePath + c.Cid + "/" + cluster + "/"
		if _, err := os.Stat(kubeConfigPath); os.IsNotExist(err) {
			err = os.MkdirAll(kubeConfigPath, 0700)
			if err != nil {
				return nil, err
			}
		}
		kubeConfig := kubeConfigPath + "config"
		f, err := os.Create(kubeConfig)
		if err != nil {
			return nil, err
		}
		_, err = f.Write(dec)
		if err != nil {
			return nil, err
		}
		client = kubeclient.New("", kubeConfig, namespace)
		if client != nil {
			c.Clients[cluster] = client
		} else {
			return nil, errors.New("failed to connect with the cluster")
		}
	}
	return client, nil
}

func (c *Connection) RemoveClient() {
	c.Lock()
	defer c.Unlock()
	err := os.RemoveAll(basePath + "/" + c.Cid)
	if err != nil {
		log.Error("Warning: Deleting kubepath", log.Fields{"err": err})
	}
}

func (c *Connection) GetClientInternal(cluster string, level string, namespace string) (types.ClientProvider, error) {
	return c.GetClient(cluster, level, namespace)
}
