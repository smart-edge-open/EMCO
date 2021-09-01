// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2020 Intel Corporation

package module

import (
	"github.com/open-ness/EMCO/src/orchestrator/pkg/infra/db"

	hpaModel "github.com/open-ness/EMCO/src/hpa-plc/pkg/model"
	log "github.com/open-ness/EMCO/src/orchestrator/pkg/infra/logutils"
	orchMod "github.com/open-ness/EMCO/src/orchestrator/pkg/module"
	pkgerrors "github.com/pkg/errors"
)

/*
AddIntent adds a given intent to the deployment-intent-group and stores in the db.
Other input parameters for it - projectName, compositeAppName, version, DeploymentIntentgroupName
*/
func (c *HpaPlacementClient) AddIntent(a hpaModel.DeploymentHpaIntent, p string, ca string, v string, di string, exists bool) (hpaModel.DeploymentHpaIntent, error) {
	//Check for the intent already exists here.
	res, dependentErrStaus, err := c.GetIntent(a.MetaData.Name, p, ca, v, di)
	if err != nil && dependentErrStaus == true {
		log.Error("AddIntent ... Intent dependency check failed", log.Fields{"intent-name": a.MetaData.Name, "err": err, "res-received": res})
		return hpaModel.DeploymentHpaIntent{}, err
	} else if (err == nil) && (!exists) {
		log.Error("AddIntent ... Intent already exists", log.Fields{"intent-name": a.MetaData.Name, "err": err, "res-received": res})
		return hpaModel.DeploymentHpaIntent{}, pkgerrors.New("Intent already exists")
	}

	dbKey := HpaIntentKey{
		IntentName:            a.MetaData.Name,
		Project:               p,
		CompositeApp:          ca,
		Version:               v,
		DeploymentIntentGroup: di,
	}

	log.Info("AddIntent ... Creating DB entry", log.Fields{"StoreName": c.db.StoreName, "key": dbKey, "project": p, "composite-app": ca, "composite-app-ver": v, "dep-group": di, "intent-name": a.MetaData.Name})
	err = db.DBconn.Insert(c.db.StoreName, dbKey, nil, c.db.TagMetaData, a)
	if err != nil {
		log.Error("AddIntent ... DB Error .. Creating DB entry error", log.Fields{"StoreName": c.db.StoreName, "akey": dbKey, "project": p, "composite-app": ca, "composite-app-ver": v, "dep-group": di, "intent-name": a.MetaData.Name})
		return hpaModel.DeploymentHpaIntent{}, pkgerrors.Wrap(err, "DB Error .. Creating DB entry")
	}
	return a, nil
}

/*
GetIntent takes in an IntentName, ProjectName, CompositeAppName, Version and DeploymentIntentGroup.
It returns the Intent.
*/
func (c *HpaPlacementClient) GetIntent(i string, p string, ca string, v string, di string) (hpaModel.DeploymentHpaIntent, bool, error) {
	// check whether dependencies are met
	err := CheckPlacementIntentDependency(p, ca, v, di)
	if err != nil {
		return hpaModel.DeploymentHpaIntent{}, true, err
	}

	dbKey := HpaIntentKey{
		IntentName:            i,
		Project:               p,
		CompositeApp:          ca,
		Version:               v,
		DeploymentIntentGroup: di,
	}

	result, err := db.DBconn.Find(c.db.StoreName, dbKey, c.db.TagMetaData)
	if err != nil {
		log.Error("GetIntent ... DB Error .. Get Intent error", log.Fields{"intent-name": i, "err": err})
		return hpaModel.DeploymentHpaIntent{}, false, pkgerrors.Wrap(err, "DB Error .. Get Intent error")
	}

	if result != nil {
		a := hpaModel.DeploymentHpaIntent{}
		err = db.DBconn.Unmarshal(result[0], &a)
		if err != nil {
			log.Error("GetIntent ... Unmarshalling  Intent error", log.Fields{"intent-name": i})
			return hpaModel.DeploymentHpaIntent{}, false, pkgerrors.Wrap(err, "Unmarshalling  HpaIntent error")
		}
		return a, false, nil
	}
	return hpaModel.DeploymentHpaIntent{}, false, pkgerrors.New("Error getting Intent")
}

/*
GetAllIntents takes in projectName, CompositeAppName, CompositeAppVersion,
DeploymentIntentName . It returns ListOfIntents.
*/
func (c HpaPlacementClient) GetAllIntents(p string, ca string, v string, di string) ([]hpaModel.DeploymentHpaIntent, error) {
	// check whether dependencies are met
	err := CheckPlacementIntentDependency(p, ca, v, di)
	if err != nil {
		return []hpaModel.DeploymentHpaIntent{}, err
	}

	dbKey := HpaIntentKey{
		IntentName:            "",
		Project:               p,
		CompositeApp:          ca,
		Version:               v,
		DeploymentIntentGroup: di,
	}

	result, err := db.DBconn.Find(c.db.StoreName, dbKey, c.db.TagMetaData)
	if err != nil {
		log.Error("GetAllIntents ... DB Error .. Get HpaIntents db error", log.Fields{"StoreName": c.db.StoreName, "project": p, "composite-app": ca, "composite-app-ver": v, "dep-group": di, "len_result": len(result), "err": err})
		return []hpaModel.DeploymentHpaIntent{}, pkgerrors.Wrap(err, "DB Error .. Get HpaIntents db error")
	}
	log.Info("GetAllIntents ... db result", log.Fields{"StoreName": c.db.StoreName, "key": dbKey, "project": p, "composite-app": ca, "composite-app-ver": v, "dep-group": di})

	var listOfIntents []hpaModel.DeploymentHpaIntent
	for i := range result {
		a := hpaModel.DeploymentHpaIntent{}
		if result[i] != nil {
			err = db.DBconn.Unmarshal(result[i], &a)
			if err != nil {
				log.Error("GetAllIntents ... Unmarshalling HpaIntents error", log.Fields{"deploymentgroup": di})
				return []hpaModel.DeploymentHpaIntent{}, pkgerrors.Wrap(err, "Unmarshalling Intent error")
			}
			listOfIntents = append(listOfIntents, a)
		}
	}
	return listOfIntents, nil
}

/*
GetAllIntentsByApp takes in appName, projectName, CompositeAppName, CompositeAppVersion,
DeploymentIntentName . It returns ListOfIntents.
*/
func (c HpaPlacementClient) GetAllIntentsByApp(app string, p string, ca string, v string, di string) ([]hpaModel.DeploymentHpaIntent, error) {
	// check whether dependencies are met
	err := CheckPlacementIntentDependency(p, ca, v, di)
	if err != nil {
		return []hpaModel.DeploymentHpaIntent{}, err
	}

	dbKey := HpaIntentKey{
		IntentName:            "",
		Project:               p,
		CompositeApp:          ca,
		Version:               v,
		DeploymentIntentGroup: di,
	}

	result, err := db.DBconn.Find(c.db.StoreName, dbKey, c.db.TagMetaData)
	if err != nil {
		log.Error("GetAllIntentsByApp .. DB Error", log.Fields{"StoreName": c.db.StoreName, "project": p, "composite-app": ca, "composite-app-ver": v, "dep-group": di, "len_result": len(result), "err": err})
		return []hpaModel.DeploymentHpaIntent{}, pkgerrors.Wrap(err, "GetAllIntentsByApp .. DB Error")
	}
	log.Info("GetAllIntentsByApp ... db result",
		log.Fields{"StoreName": c.db.StoreName, "key": dbKey, "project": p, "composite-app": ca, "composite-app-ver": v, "dep-group": di, "app-name": app})

	var listOfIntents []hpaModel.DeploymentHpaIntent
	for i := range result {
		a := hpaModel.DeploymentHpaIntent{}
		if result[i] != nil {
			err = db.DBconn.Unmarshal(result[i], &a)
			if err != nil {
				log.Error("GetAllIntentsByApp ... Unmarshalling HpaIntents error", log.Fields{"deploymentgroup": di})
				return []hpaModel.DeploymentHpaIntent{}, pkgerrors.Wrap(err, "Unmarshalling Intent error")
			}
			if a.Spec.AppName == app {
				listOfIntents = append(listOfIntents, a)
			}
		}
	}
	return listOfIntents, nil
}

/*
GetIntentByName takes in IntentName, projectName, CompositeAppName, CompositeAppVersion
and deploymentIntentGroupName returns the list of intents under the IntentName.
*/
func (c HpaPlacementClient) GetIntentByName(i string, p string, ca string, v string, di string) (hpaModel.DeploymentHpaIntent, error) {
	// check whether dependencies are met
	err := CheckPlacementIntentDependency(p, ca, v, di)
	if err != nil {
		return hpaModel.DeploymentHpaIntent{}, err
	}

	dbKey := HpaIntentKey{
		IntentName:            i,
		Project:               p,
		CompositeApp:          ca,
		Version:               v,
		DeploymentIntentGroup: di,
	}

	result, err := db.DBconn.Find(c.db.StoreName, dbKey, c.db.TagMetaData)
	if err != nil {
		log.Error("GetIntentByName ... DB Error .. Get HpaIntent error", log.Fields{"intent-name": i})
		return hpaModel.DeploymentHpaIntent{}, pkgerrors.Wrap(err, "DB Error .. Get HpaIntent error")
	}
	var a hpaModel.DeploymentHpaIntent
	err = db.DBconn.Unmarshal(result[0], &a)
	if err != nil {
		log.Error("GetIntentByName ...  Unmarshalling HpaIntent error", log.Fields{"intent-name": i})
		return hpaModel.DeploymentHpaIntent{}, pkgerrors.Wrap(err, "Unmarshalling Intent error")
	}
	return a, nil
}

// DeleteIntent deletes a given intent tied to project, composite app and deployment intent group
func (c HpaPlacementClient) DeleteIntent(i string, p string, ca string, v string, di string) error {
	dbKey := HpaIntentKey{
		IntentName:            i,
		Project:               p,
		CompositeApp:          ca,
		Version:               v,
		DeploymentIntentGroup: di,
	}

	//Check for the Intent already exists
	_, _, err := c.GetIntent(i, p, ca, v, di)
	if err != nil {
		log.Error("DeleteIntent ... Intent does not exist", log.Fields{"Intent-name": i, "err": err})
		return pkgerrors.Wrapf(err, "DB Error Intent[%s] does not exist", i)
	}

	log.Info("DeleteIntent ... Delete Hpa Intent entry", log.Fields{"StoreName": c.db.StoreName, "key": dbKey, "project": p, "composite-app": ca, "composite-app-ver": v, "dep-group": di, "intent-name": i})
	err = db.DBconn.Remove(c.db.StoreName, dbKey)
	if err != nil {
		log.Error("DeleteIntent ... DB Error .. Delete Hpa Intent entry error", log.Fields{"err": err, "StoreName": c.db.StoreName, "key": dbKey, "project": p, "composite-app": ca, "composite-app-ver": v, "dep-group": di, "intent-name": i})
		return pkgerrors.Wrapf(err, "DB Error .. Delete Hpa Intent[%s] DB Error", i)
	}
	return nil
}

// CheckPlacementIntentDependency  ... check whether intent(example: project, composite app..) dependncies exist
func CheckPlacementIntentDependency(p string, ca string, v string, di string) error {

	//Check if project exists
	_, err := orchMod.NewProjectClient().GetProject(p)
	if err != nil {
		return pkgerrors.Wrapf(err, "dependency not found .. Unable to find the project[%v]", p)
	}

	// check if compositeApp exists
	_, err = orchMod.NewCompositeAppClient().GetCompositeApp(ca, v, p)
	if err != nil {
		return pkgerrors.Wrapf(err, "dependency not found .. Unable to find the composite-app[%v]", ca)
	}

	// check if the deploymentIntentGrpName exists
	_, err = orchMod.NewDeploymentIntentGroupClient().GetDeploymentIntentGroup(di, p, ca, v)
	if err != nil {
		return pkgerrors.Wrapf(err, "dependency not found .. Unable to find the deployment-intent-group-name[%v]", di)
	}
	return nil
}
