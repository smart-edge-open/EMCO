// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2020 Intel Corporation

package contextupdateserver

import (
	"context"
	"errors"
	"fmt"

	"github.com/open-ness/EMCO/src/hpa-ac/internal/action"
	contextpb "github.com/open-ness/EMCO/src/orchestrator/pkg/grpc/contextupdate"
	log "github.com/open-ness/EMCO/src/orchestrator/pkg/infra/logutils"
)

type contextupdateServer struct {
	contextpb.UnimplementedContextupdateServer
}

func (cs *contextupdateServer) UpdateAppContext(ctx context.Context, req *contextpb.ContextUpdateRequest) (*contextpb.ContextUpdateResponse, error) {
	log.Info("Received Update App Context request .. start", log.Fields{"req": req})

	if (req != nil) && (len(req.AppContext) > 0) {
		err := action.UpdateAppContext(req.IntentName, req.AppContext)
		if err != nil {
			log.Error("Received Update App Context request .. internal error.", log.Fields{"req": req, "err": err})
			return &contextpb.ContextUpdateResponse{AppContextUpdated: false, AppContextUpdateMessage: err.Error()}, nil
		}
	} else {
		log.Error("Received Update App Context request .. invalid request error.", log.Fields{"req": req})
		return &contextpb.ContextUpdateResponse{AppContextUpdated: false, AppContextUpdateMessage: errors.New("invalid request error").Error()}, nil
	}

	log.Info("Received Update App Context request .. end", log.Fields{"req": req})
	return &contextpb.ContextUpdateResponse{AppContextUpdated: true, AppContextUpdateMessage: fmt.Sprintf("Successful application of intent %v to %v", req.IntentName, req.AppContext)}, nil
}

// NewContextUpdateServer exported
func NewContextupdateServer() *contextupdateServer {
	s := &contextupdateServer{}
	return s
}
