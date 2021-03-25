// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2020 Intel Corporation

package module

import (
	log "github.com/open-ness/EMCO/src/orchestrator/pkg/infra/logutils"
	pkgerrors "github.com/pkg/errors"
)

type Instantiator struct {
	project              string
	compositeApp         string
	compAppVersion       string
	deploymentIntent     string
	deploymentIntenetGrp DeploymentIntentGroup
}

// MakeAppContext shall make an app context and store the app context into etcd. This shall return contextForCompositeApp
func (i *Instantiator) MakeAppContext() (contextForCompositeApp, error) {

	rName := i.deploymentIntenetGrp.Spec.Version //rName is releaseName
	overrideValues := i.deploymentIntenetGrp.Spec.OverrideValuesObj
	cp := i.deploymentIntenetGrp.Spec.Profile

	gIntent, err := findGenericPlacementIntent(i.project, i.compositeApp, i.compAppVersion, i.deploymentIntent)
	if err != nil {
		return contextForCompositeApp{}, err
	}

	log.Info(":: The name of the GenPlacIntent ::", log.Fields{"GenPlmtIntent": gIntent})
	log.Info(":: DeploymentIntentGroup, ReleaseName, CompositeProfile ::", log.Fields{"dIGrp": i.deploymentIntenetGrp.MetaData.Name, "releaseName": rName, "cp": cp})

	allApps, err := NewAppClient().GetApps(i.project, i.compositeApp, i.compAppVersion)
	if err != nil {
		return contextForCompositeApp{}, pkgerrors.Wrap(err, "Not finding the apps")
	}

	dcmClusters, namespace, level, err := getLogicalCloudInfo(i.project, i.deploymentIntenetGrp.Spec.LogicalCloud)
	if err != nil {
		return contextForCompositeApp{}, err
	}

	cca, err := makeAppContextForCompositeApp(i.project, i.compositeApp, i.compAppVersion, rName, i.deploymentIntent, namespace, level)
	if err != nil {
		return contextForCompositeApp{}, err
	}

	err = storeAppContextIntoRunTimeDB(allApps, cca, overrideValues, dcmClusters, i.project, i.compositeApp, i.compAppVersion, rName, cp, gIntent, i.deploymentIntent, namespace)
	if err != nil {
		return contextForCompositeApp{}, pkgerrors.Wrap(err, "Error in storeAppContextIntoETCd")
	}

	return cca, nil
}
