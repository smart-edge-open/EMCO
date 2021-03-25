// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2020 Intel Corporation

package contextupdateserver

import (
	"context"
	"fmt"

	"github.com/open-ness/EMCO/src/nps/internal/networkpolicy"
	contextpb "github.com/open-ness/EMCO/src/dtc/pkg/grpc/contextupdate"
	log "github.com/open-ness/EMCO/src/orchestrator/pkg/infra/logutils"
)

type contextupdateServer struct {
	contextpb.UnimplementedContextupdateServer
}

func (cs *contextupdateServer) UpdateAppContext(ctx context.Context, req *contextpb.ContextUpdateRequest) (*contextpb.ContextUpdateResponse, error) {
	log.Info("Received Update App Context request", log.Fields{
		"AppContextId": req.AppContext,
		"IntentName":   req.IntentName,
	})


	err := networkpolicy.UpdateAppContext(req.IntentName, req.AppContext)
	if err != nil {
		return &contextpb.ContextUpdateResponse{AppContextUpdated: false, AppContextUpdateMessage: err.Error()}, nil
	}

	return &contextpb.ContextUpdateResponse{AppContextUpdated: true, AppContextUpdateMessage: fmt.Sprintf("Successful application of intent %v to %v", req.IntentName, req.AppContext)}, nil
}

// NewContextUpdateServer exported
func NewContextupdateServer() *contextupdateServer {
	s := &contextupdateServer{}
	return s
}
