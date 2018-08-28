// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2020 Intel Corporation

package module

import (
	"context"
	"sync"
	"time"

	log "github.com/open-ness/EMCO/src/orchestrator/pkg/infra/logutils"
	"github.com/open-ness/EMCO/src/orchestrator/pkg/infra/rpc"
	readynotifypb "github.com/open-ness/EMCO/src/rsync/pkg/grpc/readynotify"
)

// const rsyncName = "rsync"

/*
RsyncInfo consists of rsyncName, hostName and portNumber.
*/
type RsyncInfo struct {
	RsyncName  string
	hostName   string
	portNumber int
}

var rsyncInfo RsyncInfo
var mutex = &sync.Mutex{}

// InitRsyncClient initializes connections to the Resource Synchronizer service
func initRsyncClient() bool {
	if (RsyncInfo{}) == rsyncInfo {
		mutex.Lock()
		defer mutex.Unlock()
		log.Error("[ReadyNotify gRPC] RsyncInfo not set - InitRsyncClient failed", log.Fields{
			"Rsyncname":  rsyncInfo.RsyncName,
			"Hostname":   rsyncInfo.hostName,
			"PortNumber": rsyncInfo.portNumber,
		})
		return false
	}
	rpc.UpdateRpcConn(rsyncInfo.RsyncName, rsyncInfo.hostName, rsyncInfo.portNumber)
	return true
}

// NewRsyncInfo shall return a newly created RsyncInfo object
func NewRsyncInfo(rName, h string, pN int) RsyncInfo {
	mutex.Lock()
	defer mutex.Unlock()
	rsyncInfo = RsyncInfo{RsyncName: rName, hostName: h, portNumber: pN}
	return rsyncInfo

}

// InvokeReadyNotify will make a gRPC call to the resource synchronizer and will
// will subscribe DCM to alerts from the rsync gRPC server ("ready-notify")
func InvokeReadyNotify(appContextId string) error {
	var err error
	_, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	conn := rpc.GetRpcConn(rsyncName)
	if conn == nil {
		initRsyncClient()
		conn = rpc.GetRpcConn(rsyncName)
	}

	client := readynotifypb.NewReadyNotifyClient(conn)
	go subscribe(client, appContextId)
	if err != nil {
		log.Error("[ReadyNotify gRPC] Error invoking", log.Fields{"err": err})
	}
	return err
}

func processAlert(stream readynotifypb.ReadyNotify_AlertClient) {
	var appContextId string
	var lcc *LogicalCloudClient
	var dcc *ClusterClient

	resp, err := stream.Recv()
	if err != nil {
		log.Error("[ReadyNotify gRPC] Failed to receive notification", log.Fields{"err": err})
		return
	}

	appContextId = resp.AppContext
	log.Info("[ReadyNotify gRPC] Received alert from rsync", log.Fields{"appContextId": appContextId, "err": err})

	// if this point is reached, it means all clusters' certificates have been issued,
	// so it's time for DCM to build all the L1 kubeconfigs and store them in CloudConfig

	// Get the actual Logical Cloud via the known AppContext ID
	lcc = NewLogicalCloudClient() // in logicalcloud.go
	project, logicalCloud, err := lcc.util.GetLogicalCloudFromContext(lcc.storeName, appContextId)
	if err != nil {
		log.Error("[ReadyNotify gRPC] Couldn't get Logical Cloud using AppContext ID", log.Fields{"err": err})
		return
	}
	log.Info("[ReadyNotify gRPC] Project and Logical Cloud obtained", log.Fields{"project": project, "logicalCloud": logicalCloud})

	// Get all clusters of the Logical Cloud
	dcc = NewClusterClient() // in cluster.go
	clusterList, err := dcc.GetAllClusters(project, logicalCloud)
	if err != nil {
		log.Error("[ReadyNotify gRPC] Failed getting all clusters of Logical Cloud", log.Fields{"logicalCloud": logicalCloud, "project": project})
		return
	}
	for _, cluster := range clusterList {
		_, err = dcc.GetClusterConfig(project, logicalCloud, cluster.MetaData.ClusterReference)
		// discard kubeconfig returned because it's not needed here
		if err != nil {
			log.Error("[ReadyNotify gRPC] Generating kubeconfig or storing CloudConfig failed", log.Fields{"logicalCloud": logicalCloud, "project": project, "cluster": cluster.MetaData.ClusterReference})
			return
		}
		log.Info("[ReadyNotify gRPC] Generated kubeconfig and created CloudConfig for cluster", log.Fields{"project": project, "logicalCloud": logicalCloud, "cluster": cluster.MetaData.ClusterReference})
		// if this point is reached, the kubeconfig is already stored in CloudConfig
	}
	log.Info("[ReadyNotify gRPC] All CloudConfigs for Logical Cloud have been created", log.Fields{"project": project, "logicalCloud": logicalCloud})
}

func subscribe(client readynotifypb.ReadyNotifyClient, appContextId string) {
	stream, err := client.Alert(context.Background(), &readynotifypb.Topic{ClientName: "dcm", AppContext: appContextId})
	if err != nil {
		log.Error("[ReadyNotify gRPC] Failed to subscribe to alerts", log.Fields{"err": err, "appContextId": appContextId})
	}

	log.Info("[ReadyNotify gRPC] Subscribing to alerts about appcontext ID", log.Fields{"appContextId": appContextId})
	go processAlert(stream)

	stream.CloseSend()
}
