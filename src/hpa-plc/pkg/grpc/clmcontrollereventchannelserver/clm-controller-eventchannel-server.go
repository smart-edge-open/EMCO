// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2020 Intel Corporation

package clmcontrollereventchannel

import (
	"context"
	"errors"
	"fmt"

	clmcontrollerpb "github.com/open-ness/EMCO/src/clm/pkg/grpc/controller-eventchannel"
	"github.com/open-ness/EMCO/src/hpa-plc/internal/action"
	log "github.com/open-ness/EMCO/src/orchestrator/pkg/infra/logutils"
)

// ClmControllerEventChannelServer ...
type ClmControllerEventChannelServer struct {
}

// Publish ...
func (cs *ClmControllerEventChannelServer) Publish(ctx context.Context, req *clmcontrollerpb.ClmControllerEventRequest) (*clmcontrollerpb.ClmControllerEventResponse, error) {
	log.Info("Publish request .. start", log.Fields{"req": req})

	if (req != nil) && (len(req.ClusterName) > 0) {
		err := action.Publish(ctx, req)
		if err != nil {
			log.Error("Publish request .. internal error.", log.Fields{"req": req, "err": err})
			return &clmcontrollerpb.ClmControllerEventResponse{ProviderName: req.ProviderName, ClusterName: req.ClusterName, Status: false, Message: err.Error()}, nil
		}
	} else {
		log.Error("Publish request .. invalid request error.", log.Fields{"req": req})
		return &clmcontrollerpb.ClmControllerEventResponse{Status: false, Message: errors.New("invalid request error").Error()}, nil
	}

	log.Info("Publish request .. end", log.Fields{"req": req})
	return &clmcontrollerpb.ClmControllerEventResponse{ProviderName: req.ProviderName, ClusterName: req.ClusterName, Status: true, Message: fmt.Sprintf("Successfully Published for ProviderName[%v] ClusterName[%v]", req.ProviderName, req.ClusterName)}, nil
}

// NewControllerEventchannelServer ...
func NewControllerEventchannelServer() *ClmControllerEventChannelServer {
	s := &ClmControllerEventChannelServer{}
	return s
}
