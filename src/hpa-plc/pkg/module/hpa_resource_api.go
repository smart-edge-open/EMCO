// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2020 Intel Corporation

package module

import (
	"github.com/open-ness/EMCO/src/orchestrator/pkg/infra/db"

	hpaModel "github.com/open-ness/EMCO/src/hpa-plc/pkg/model"
	log "github.com/open-ness/EMCO/src/orchestrator/pkg/infra/logutils"
	pkgerrors "github.com/pkg/errors"
)

/*
 AddResource ... AddResource adds a given consumer to the hpa-intent-name and stores in the db.
 Other input parameters for it - projectName, compositeAppName, version, DeploymentIntentgroupName, intentName, consumerName
*/
func (c *HpaPlacementClient) AddResource(a hpaModel.HpaResourceRequirement, p string, ca string, v string, di string, i string, cn string, exists bool) (hpaModel.HpaResourceRequirement, error) {
	//Check for the Resource already exists here.
	res, dependentErrStaus, err := c.GetResource(a.MetaData.Name, p, ca, v, di, i, cn)
	if err != nil && dependentErrStaus == true {
		log.Error("AddResource ... Resource dependency check failed", log.Fields{"intent-name": a.MetaData.Name, "err": err, "res-received": res})
		return hpaModel.HpaResourceRequirement{}, err
	} else if err == nil && !exists {
		log.Error("AddResource ... Resource already exists", log.Fields{"resource-name": a.MetaData.Name, "err": err, "res-received": res})
		return hpaModel.HpaResourceRequirement{}, pkgerrors.New("Resource already exists")
	}

	dbKey := HpaResourceKey{
		ResourceName:          a.MetaData.Name,
		ConsumerName:          cn,
		IntentName:            i,
		Project:               p,
		CompositeApp:          ca,
		Version:               v,
		DeploymentIntentGroup: di,
	}

	log.Info("AddResource ... Creating DB entry", log.Fields{"StoreName": c.db.StoreName, "key": dbKey, "project": p, "composite-app": ca, "composite-app-ver": v, "dep-group": di, "intent-name": i, "consumer-name": cn, "resource-name": a.MetaData.Name})
	err = db.DBconn.Insert(c.db.StoreName, dbKey, nil, c.db.TagMetaData, a)
	if err != nil {
		log.Error("AddResource ... DB Error .. Creating DB entry error", log.Fields{"resource-name": a.MetaData.Name})
		return hpaModel.HpaResourceRequirement{}, pkgerrors.Wrap(err, "DB Error .. Creating DB entry error")
	}
	return a, nil
}

/*
 GetResource ... takes in an ConsumerName, IntentName, ProjectName, CompositeAppName, Version, DeploymentIntentGroup and intentName.
 It returns the Resource.
*/
func (c *HpaPlacementClient) GetResource(rn string, p string, ca string, v string, di string, i string, cn string) (hpaModel.HpaResourceRequirement, bool, error) {
	// check whether dependencies are met
	err := CheckPlacementResourceDependency(p, ca, v, di, i, cn)
	if err != nil {
		return hpaModel.HpaResourceRequirement{}, true, err
	}

	dbKey := HpaResourceKey{
		ResourceName:          rn,
		ConsumerName:          cn,
		IntentName:            i,
		Project:               p,
		CompositeApp:          ca,
		Version:               v,
		DeploymentIntentGroup: di,
	}

	result, err := db.DBconn.Find(c.db.StoreName, dbKey, c.db.TagMetaData)
	if err != nil {
		log.Error("GetResource ... DB Error .. Get Resource error", log.Fields{"resource-name": rn})
		return hpaModel.HpaResourceRequirement{}, false, pkgerrors.Wrap(err, "DB Error .. Get Resource error")
	}

	if result != nil {
		a := hpaModel.HpaResourceRequirement{}
		err = db.DBconn.Unmarshal(result[0], &a)
		if err != nil {
			log.Error("GetResource ... Unmarshalling HpaResource error", log.Fields{"resource-name": rn})
			return hpaModel.HpaResourceRequirement{}, false, pkgerrors.Wrap(err, "Unmarshalling HpaResource error")
		}
		return a, false, nil
	}
	return hpaModel.HpaResourceRequirement{}, false, pkgerrors.New("Error getting Resource")
}

/*
 GetAllResources ... takes in projectName, CompositeAppName, CompositeAppVersion, DeploymentGroup,
 DeploymentIntentName, ConsumerName . It returns ListOfResources.
*/
func (c HpaPlacementClient) GetAllResources(p, ca, v, di, i, cn string) ([]hpaModel.HpaResourceRequirement, error) {
	// check whether dependencies are met
	err := CheckPlacementResourceDependency(p, ca, v, di, i, cn)
	if err != nil {
		return []hpaModel.HpaResourceRequirement{}, err
	}

	dbKey := HpaResourceKey{
		ResourceName:          "",
		ConsumerName:          cn,
		IntentName:            i,
		Project:               p,
		CompositeApp:          ca,
		Version:               v,
		DeploymentIntentGroup: di,
	}

	result, err := db.DBconn.Find(c.db.StoreName, dbKey, c.db.TagMetaData)
	if err != nil {
		log.Error("GetAllResources ... DB Error .. Get HpaResources db error", log.Fields{"consumer-name": cn})
		return []hpaModel.HpaResourceRequirement{}, pkgerrors.Wrap(err, "DB Error .. Get HpaResources db error")
	}
	log.Info("GetAllResources ... db result", log.Fields{"StoreName": c.db.StoreName, "key": dbKey, "project": p, "composite-app": ca, "composite-app-ver": v, "dep-group": di, "consumer-name": cn})

	var listOfMapOfResources []hpaModel.HpaResourceRequirement
	for i := range result {
		a := hpaModel.HpaResourceRequirement{}
		if result[i] != nil {
			err = db.DBconn.Unmarshal(result[i], &a)
			if err != nil {
				log.Error("GetAllResources ... Unmarshalling Resources error", log.Fields{"consumer-name": cn})
				return []hpaModel.HpaResourceRequirement{}, pkgerrors.Wrap(err, "Unmarshalling Resources error")
			}
			listOfMapOfResources = append(listOfMapOfResources, a)
		}
	}
	return listOfMapOfResources, nil
}

/*
 GetResourceByName ... takes in IntentName, projectName, CompositeAppName, CompositeAppVersion,
 deploymentIntentGroupName, intentName and consumerName returns the list of resource under the consumerName.
*/
func (c HpaPlacementClient) GetResourceByName(rn, p, ca, v, di, i, cn string) (hpaModel.HpaResourceRequirement, error) {
	// check whether dependencies are met
	err := CheckPlacementResourceDependency(p, ca, v, di, i, cn)
	if err != nil {
		return hpaModel.HpaResourceRequirement{}, err
	}

	dbKey := HpaResourceKey{
		ResourceName:          rn,
		ConsumerName:          cn,
		IntentName:            i,
		Project:               p,
		CompositeApp:          ca,
		Version:               v,
		DeploymentIntentGroup: di,
	}

	result, err := db.DBconn.Find(c.db.StoreName, dbKey, c.db.TagMetaData)
	if err != nil {
		log.Error("GetResourceByName ... DB Error .. Get HpaResource error", log.Fields{"resource-name": rn})
		return hpaModel.HpaResourceRequirement{}, pkgerrors.Wrap(err, "DB Error .. Get HpaResource error")
	}
	var a hpaModel.HpaResourceRequirement
	err = db.DBconn.Unmarshal(result[0], &a)
	if err != nil {
		log.Error("GetResourceByName ... Unmarshalling Resource error", log.Fields{"resource-name": rn})
		return hpaModel.HpaResourceRequirement{}, pkgerrors.Wrap(err, "Unmarshalling Resource error")
	}
	return a, nil
}

// DeleteResource ... deletes a given resource tied to project, composite app and deployment intent group, intent name, consumer name
func (c HpaPlacementClient) DeleteResource(rn string, p string, ca string, v string, di string, i string, cn string) error {
	dbKey := HpaResourceKey{
		ResourceName:          rn,
		ConsumerName:          cn,
		IntentName:            i,
		Project:               p,
		CompositeApp:          ca,
		Version:               v,
		DeploymentIntentGroup: di,
	}

	//Check for the Resource already exists
	_, _, err := c.GetResource(rn, p, ca, v, di, i, cn)
	if err != nil {
		log.Error("DeleteResource ... Resource does not exist", log.Fields{"resource-name": rn, "err": err})
		return pkgerrors.Wrapf(err, "Resource[%s] does not exist", rn)
	}

	log.Info("DeleteResource ... Delete Hpa Consumer entry", log.Fields{"StoreName": c.db.StoreName, "key": dbKey, "project": p, "composite-app": ca, "composite-app-ver": v, "dep-group": di, "intent-name": i, "consumer-name": cn, "resource-name": rn})
	err = db.DBconn.Remove(c.db.StoreName, dbKey)
	if err != nil {
		log.Error("DeleteResource ... DB Error .. Delete Hpa Resource entry error", log.Fields{"err": err, "StoreName": c.db.StoreName, "key": dbKey, "project": p, "composite-app": ca, "composite-app-ver": v, "dep-group": di, "intent-name": i, "consumer-name": cn})
		return pkgerrors.Wrap(err, "DB Error .. Delete Hpa Resource entry error")
	}
	return nil
}

// CheckPlacementResourceDependency  ... check whether resource(example: project, composite app..) dependncies exist
func CheckPlacementResourceDependency(p string, ca string, v string, di string, i string, cn string) error {

	//Check the consumer intent dependencies
	err := CheckPlacementConsumerDependency(p, ca, v, di, i)
	if err != nil {
		return err
	}

	// check if the consumerName exists
	_, _, err = NewHpaPlacementClient().GetConsumer(cn, p, ca, v, di, i)
	if err != nil {
		return pkgerrors.Wrapf(err, "dependency not found .. Unable to find the consumer-name[%v]", cn)
	}

	return nil
}
