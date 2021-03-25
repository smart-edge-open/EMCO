// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2020 Intel Corporation


package module


import (
	"fmt"
	log "github.com/open-ness/EMCO/src/orchestrator/pkg/infra/logutils"
	rsyncclient "github.com/open-ness/EMCO/src/orchestrator/pkg/grpc/updateappclient"
	
)

func callRsyncUpdate(FromContextid, ToContextid interface{}) error {
	rsyncInfo, err := queryDBAndSetRsyncInfo()
	log.Info("Calling the Rsync ", log.Fields{
		"RsyncName": rsyncInfo.RsyncName,
	})
	if err != nil {
		return err
	}

	fromAppContextID := fmt.Sprintf("%v", FromContextid)
	toAppContextID := fmt.Sprintf("%v", ToContextid)
	err = rsyncclient.InvokeUpdateApp(fromAppContextID, toAppContextID)
	if err != nil {
		return err
	}
	return nil
}