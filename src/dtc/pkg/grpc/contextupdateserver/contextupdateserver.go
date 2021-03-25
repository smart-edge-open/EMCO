// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2020 Intel Corporation

package contextupdateserver

import (
	"context"
	"fmt"
	"sort"

	contextpb "github.com/open-ness/EMCO/src/orchestrator/pkg/grpc/contextupdate"
	log "github.com/open-ness/EMCO/src/orchestrator/pkg/infra/logutils"

	"github.com/open-ness/EMCO/src/orchestrator/pkg/module/controller"
	client "github.com/open-ness/EMCO/src/dtc/pkg/grpc/contextupdateclient"

)

type contextupdateServer struct {
	contextpb.UnimplementedContextupdateServer
}

func (cs *contextupdateServer) UpdateAppContext(ctx context.Context, req *contextpb.ContextUpdateRequest) (*contextpb.ContextUpdateResponse, error) {
	log.Info("Received Update App Context request", log.Fields{
		"AppContextId": req.AppContext,
		"IntentName":   req.IntentName,
	})

	cc := controller.NewControllerClient("dtccontroller", "dtccontrollermetadata")
	clist, err := cc.GetControllers()
	if err != nil {
                log.Error("Error getting controllers", log.Fields{
                        "error": err,
                })
		return &contextpb.ContextUpdateResponse{AppContextUpdated: false, AppContextUpdateMessage: fmt.Sprintf("Error getting controllers for intent %v and Id: %v", req.IntentName, req.AppContext)}, err
	}

	// Sort the list based on priority
	sort.SliceStable(clist, func(i, j int) bool {
		return clist[i].Spec.Priority > clist[j].Spec.Priority

	})

	for _, c := range clist {
		err := client.InvokeContextUpdate(c.Metadata.Name, req.IntentName, req.AppContext)
		if err != nil {
			log.Error("invoke context update failed for sub controller", log.Fields{
				"error": err,
				"Name":  c.Metadata.Name,
			})
		}
	}

	return &contextpb.ContextUpdateResponse{AppContextUpdated: true, AppContextUpdateMessage: fmt.Sprintf("Successful application of intent %v to %v", req.IntentName, req.AppContext)}, nil
}

// NewContextUpdateServer exported
func NewContextupdateServer() *contextupdateServer {
	s := &contextupdateServer{}
	return s
}
