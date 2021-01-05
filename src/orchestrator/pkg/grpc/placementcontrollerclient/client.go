// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2020 Intel Corporation

package placementcontrollerclient

import (
	"context"
	"time"

	plsctrlclientpb "github.com/open-ness/EMCO/src/orchestrator/pkg/grpc/placementcontroller"
	log "github.com/open-ness/EMCO/src/orchestrator/pkg/infra/logutils"
	"github.com/open-ness/EMCO/src/orchestrator/pkg/infra/rpc"
	"github.com/open-ness/EMCO/src/orchestrator/pkg/module/controller"
	pkgerrors "github.com/pkg/errors"
)

// InvokeFilterClusters ..  will make the grpc call to the specified controller
func InvokeFilterClusters(plsCtrl controller.Controller, appContextId string) error {
	controllerName := plsCtrl.Metadata.Name
	log.Info("FilterClusters .. start", log.Fields{"controllerName": controllerName, "Host": plsCtrl.Spec.Host, "Port": plsCtrl.Spec.Port, "appContextId": appContextId})
	var err error
	var rpcClient plsctrlclientpb.PlacementControllerClient
	var ctrlRes *plsctrlclientpb.ResourceResponse
	ctx, cancel := context.WithTimeout(context.Background(), 600*time.Second)
	defer cancel()

	// Fetch Grpc Connection handle
	conn := rpc.GetRpcConn(plsCtrl.Metadata.Name)
	if conn != nil {
		rpcClient = plsctrlclientpb.NewPlacementControllerClient(conn)
		ctrlReq := new(plsctrlclientpb.ResourceRequest)
		ctrlReq.AppContext = appContextId
		ctrlRes, err = rpcClient.FilterClusters(ctx, ctrlReq)
		if err == nil {
			log.Info("Response from FilterClusters GRPC call", log.Fields{"status": ctrlRes.Status, "message": ctrlRes.Message})
		}
	} else {
		log.Error("FilterClusters Failed - Could not get client connection", log.Fields{"controllerName": controllerName, "appContextId": appContextId})
		return pkgerrors.Errorf("FilterClusters Failed - Could not get client connection. controllerName[%v] appContextId[%v]", controllerName, appContextId)
	}

	if err == nil {
		if ctrlRes.Status {
			log.Info("FilterClusters Successful", log.Fields{
				"Controller": controllerName,
				"AppContext": appContextId,
				"Message":    ctrlRes.Message})
			return nil
		}
		log.Error("FilterClusters UnSuccessful - Received message", log.Fields{"message": ctrlRes.Message, "controllerName": controllerName, "appContextId": appContextId})
		return pkgerrors.Errorf("FilterClusters UnSuccessful - Received message[%v] for controllerName[%v] appContextId[%v]", ctrlRes.Message, controllerName, appContextId)
	}
	log.Error("FilterClusters Failed - Received error message", log.Fields{"controllerName": controllerName, "appContextId": appContextId})
	return err
}
