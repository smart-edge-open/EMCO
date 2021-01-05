// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2020 Intel Corporation

package clmcontrollereventchannelclient

import (
	"context"
	"time"

	clmcontrollerpb "github.com/open-ness/EMCO/src/clm/pkg/grpc/controller-eventchannel"
	clmModel "github.com/open-ness/EMCO/src/clm/pkg/model"
	log "github.com/open-ness/EMCO/src/orchestrator/pkg/infra/logutils"
	"github.com/open-ness/EMCO/src/orchestrator/pkg/infra/rpc"
	pkgerrors "github.com/pkg/errors"
)

// SendControllerEvent ..  will make the grpc call to the specified controller
func SendControllerEvent(providerName string, clusterName string, event clmcontrollerpb.ClmControllerEventType, clmCtrl clmModel.Controller) error {
	controllerName := clmCtrl.Metadata.Name
	log.Info("SendControllerEvent .. start", log.Fields{"provider-name": providerName, "cluster-name": clusterName, "event": event, "controller": clmCtrl})
	var err error
	var rpcClient clmcontrollerpb.ClmControllerEventChannelClient
	var ctrlRes *clmcontrollerpb.ClmControllerEventResponse
	ctx, cancel := context.WithTimeout(context.Background(), 600*time.Second)
	defer cancel()

	// Fetch Grpc Connection handle
	conn := rpc.GetRpcConn(clmCtrl.Metadata.Name)
	if conn != nil {
		rpcClient = clmcontrollerpb.NewClmControllerEventChannelClient(conn)
		ctrlReq := new(clmcontrollerpb.ClmControllerEventRequest)
		ctrlReq.ProviderName = providerName
		ctrlReq.ClusterName = clusterName
		ctrlReq.Event = event
		log.Info("SendControllerEvent .. Sending event", log.Fields{"controller": clmCtrl, "ctrlReq": ctrlReq})
		ctrlRes, err = rpcClient.Publish(ctx, ctrlReq)
		if err == nil {
			log.Info("Response from SendControllerEvent GRPC call", log.Fields{"status": ctrlRes.Status, "message": ctrlRes.Message})
		}
	} else {
		log.Error("SendControllerEvent Failed - Could not get client connection to grpc-server.", log.Fields{"controllerName": controllerName})
		return pkgerrors.Errorf("SendControllerEvent Failed - Could not get client connection to grpc-server[%v]", controllerName)
	}

	if err == nil {
		if ctrlRes.Status {
			log.Info("SendControllerEvent Successful", log.Fields{
				"provider-name": providerName,
				"cluster-name":  clusterName,
				"event":         event,
				"controller":    clmCtrl})
			return nil
		}
		log.Error("SendControllerEvent UnSuccessful - Received message", log.Fields{"message": ctrlRes.Message, "controllerName": controllerName})
		return pkgerrors.Errorf("SendControllerEvent UnSuccessful - Received message: %v", ctrlRes.Message)
	}
	log.Error("SendControllerEvent Failed - Received error message", log.Fields{"message": ctrlRes.Message, "controllerName": controllerName})
	return err
}
