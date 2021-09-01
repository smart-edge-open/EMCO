// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2020 Intel Corporation

package placementcontroller

import (
	"context"
	"errors"
	"fmt"

	"github.com/open-ness/EMCO/src/hpa-plc/internal/action"
	placementcontrollerpb "github.com/open-ness/EMCO/src/orchestrator/pkg/grpc/placementcontroller"
	log "github.com/open-ness/EMCO/src/orchestrator/pkg/infra/logutils"
)

// HpaPlacementcontrollerServer ...
type HpaPlacementcontrollerServer struct {
}

// FilterClusters ...
func (cs *HpaPlacementcontrollerServer) FilterClusters(ctx context.Context, req *placementcontrollerpb.ResourceRequest) (*placementcontrollerpb.ResourceResponse, error) {
	log.Info("Received HPA FilterClusters request .. start", log.Fields{"ctx": ctx, "req": req})

	if (req != nil) && (len(req.AppContext) > 0) {
		err := action.FilterClusters(req.AppContext)
		if err != nil {
			log.Error("Received HPA FilterClusters request .. internal error.", log.Fields{"req": req, "err": err})
			return &placementcontrollerpb.ResourceResponse{AppContext: req.AppContext, Status: false, Message: err.Error()}, nil
		}
	} else {
		log.Error("Received HPA FilterClusters request .. invalid request error.", log.Fields{"req": req})
		return &placementcontrollerpb.ResourceResponse{Status: false, Message: errors.New("invalid request error").Error()}, nil
	}

	log.Info("Received HPA FilterClusters request .. end", log.Fields{"req": req})
	return &placementcontrollerpb.ResourceResponse{AppContext: req.AppContext, Status: true, Message: fmt.Sprintf("Successful HPA Filtering of clusters for AppCtx[%v]", req.AppContext)}, nil
}

// NewHpaPlacementControllerServer ...
func NewHpaPlacementControllerServer() *HpaPlacementcontrollerServer {
	s := &HpaPlacementcontrollerServer{}
	return s
}
