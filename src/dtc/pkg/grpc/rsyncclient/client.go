// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2021 Intel Corporation

package rsyncclient

import (
	"context"
	"fmt"
	"sync"
	"time"

	pkgerrors "github.com/pkg/errors"
	"google.golang.org/grpc"

	"github.com/open-ness/EMCO/src/orchestrator/pkg/grpc/installappclient"
	log "github.com/open-ness/EMCO/src/orchestrator/pkg/infra/logutils"
	"github.com/open-ness/EMCO/src/orchestrator/pkg/infra/rpc"
	"github.com/open-ness/EMCO/src/orchestrator/pkg/module/controller"
	readynotifypb "github.com/open-ness/EMCO/src/rsync/pkg/grpc/readynotify"
)

const (
	rsyncName = "rsync"
)

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

// InitRsyncClient initializes connctions to the Resource Synchronizer service
func initRsyncClient() bool {
	if (RsyncInfo{}) == rsyncInfo {
		mutex.Lock()
		defer mutex.Unlock()
		log.Error("RsyncInfo not set. InitRsyncClient failed", log.Fields{
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

// QueryDBAndSetRsyncInfo queries the MCO db to find the record the sync controller and then sets the RsyncInfo global variable
func QueryDBAndSetRsyncInfo() (installappclient.RsyncInfo, error) {
	client := controller.NewControllerClient()
	vals, _ := client.GetControllers()
	for _, v := range vals {
		if v.Metadata.Name == rsyncName {
			log.Info("Initializing RPC connection to resource synchronizer", log.Fields{
				"Controller": v.Metadata.Name,
			})
			rsyncInfoInstallAppClient := installappclient.NewRsyncInfo(v.Metadata.Name, v.Spec.Host, v.Spec.Port)
			rsyncInfo = NewRsyncInfo(v.Metadata.Name, v.Spec.Host, v.Spec.Port)
			return rsyncInfoInstallAppClient, nil
		}
	}
	return installappclient.RsyncInfo{}, pkgerrors.Errorf("queryRsyncInfoInMCODB Failed - Could not get find rsync by name : %v", rsyncName)
}

// CallRsyncInstall method shall take in the app context id and invoke the rsync service via grpc
func CallRsyncInstall(contextid interface{}) error {
	rsyncInfo, err := QueryDBAndSetRsyncInfo()
	log.Info("Calling the Rsync ", log.Fields{
		"RsyncName": rsyncInfo.RsyncName,
	})
	if err != nil {
		return err
	}

	appContextID := fmt.Sprintf("%v", contextid)
	err = installappclient.InvokeInstallApp(appContextID)
	if err != nil {
		return err
	}
	return nil
}

// InvokeReadyNotify will make a gRPC call to the resource synchronizer and will
// will subscribe DCM to alerts from the rsync gRPC server ("ready-notify")
func InvokeReadyNotify(appContextID string) (readynotifypb.ReadyNotify_AlertClient, readynotifypb.ReadyNotifyClient, error) {

	var stream readynotifypb.ReadyNotify_AlertClient
	var client readynotifypb.ReadyNotifyClient
	_, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	rsyncInfo, err := QueryDBAndSetRsyncInfo()
	log.Info("Calling rsync", log.Fields{
		"RsyncName": rsyncInfo.RsyncName,
	})
	if err != nil {
		log.Error("", log.Fields{"err": err})
		return stream, client, pkgerrors.Wrapf(err, "Unable to find the rsync info from MCO db")
	}

	conn := rpc.GetRpcConn(rsyncName)
	if conn == nil {
		initRsyncClient()
		conn = rpc.GetRpcConn(rsyncName)
		if conn == nil {
			return stream, client, pkgerrors.Wrapf(err, "Unable to connect to rsync")
		}
	}

	client = readynotifypb.NewReadyNotifyClient(conn)

	if client != nil {
		stream, err = client.Alert(context.Background(), &readynotifypb.Topic{ClientName: "dtc", AppContext: appContextID}, grpc.WaitForReady(true))
		if err != nil {
			log.Error("[ReadyNotify gRPC] Failed to subscribe to alerts", log.Fields{"err": err, "appContextId": appContextID})
			time.Sleep(5 * time.Second)
			InvokeReadyNotify(appContextID)
		}

		log.Info("[ReadyNotify gRPC] Subscribing to alerts about appcontext ID", log.Fields{"appContextId": appContextID})

	}

	return stream, client, nil
}
