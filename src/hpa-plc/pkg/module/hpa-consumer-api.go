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
AddConsumer ... AddConsumer adds a given consumenr to the hpa-intent-name and stores in the db.
Other input parameters for it - projectName, compositeAppName, version, DeploymentIntentgroupName, intentName
*/
func (c *HpaPlacementClient) AddConsumer(a hpaModel.HpaResourceConsumer, p string, ca string, v string, di string, i string, exists bool) (hpaModel.HpaResourceConsumer, error) {
	//Check for the Consumer already exists here.
	res, dependentErrStaus, err := c.GetConsumer(a.MetaData.Name, p, ca, v, di, i)
	if err != nil && dependentErrStaus == true {
		log.Error("AddConsumer ... Consumer dependency check failed", log.Fields{"intent-name": a.MetaData.Name, "err": err, "res-received": res})
		return hpaModel.HpaResourceConsumer{}, err
	} else if err == nil && !exists {
		log.Error("AddConsumer ... Consumer already exists", log.Fields{"consumer-name": a.MetaData.Name, "err": err, "res-received": res})
		return hpaModel.HpaResourceConsumer{}, pkgerrors.New("Consumer already exists")
	}
	dbKey := HpaConsumerKey{
		ConsumerName:          a.MetaData.Name,
		IntentName:            i,
		Project:               p,
		CompositeApp:          ca,
		Version:               v,
		DeploymentIntentGroup: di,
	}

	log.Info("AddConsumer ... Creating DB entry entry", log.Fields{"StoreName": c.db.StoreName, "key": dbKey, "project": p, "composite-app": ca, "composite-app-ver": v, "dep-group": di, "intent-name": i, "consumer-name": a.MetaData.Name})
	err = db.DBconn.Insert(c.db.StoreName, dbKey, nil, c.db.TagMetaData, a)
	if err != nil {
		log.Error("AddConsumer ...  DB Error .. Creating DB entry error", log.Fields{"consumer-name": a.MetaData.Name, "err": err})
		return hpaModel.HpaResourceConsumer{}, pkgerrors.Wrap(err, "DB Error .. Creating DB entry error")
	}
	return a, nil
}

/*
GetConsumer ... takes in an IntentName, ProjectName, CompositeAppName, Version, DeploymentIntentGroup and intentName.
It returns the Consumer.
*/
func (c *HpaPlacementClient) GetConsumer(cn string, p string, ca string, v string, di string, i string) (hpaModel.HpaResourceConsumer, bool, error) {
	// check whether dependencies are met
	err := CheckPlacementConsumerDependency(p, ca, v, di, i)
	if err != nil {
		return hpaModel.HpaResourceConsumer{}, true, err
	}

	dbKey := HpaConsumerKey{
		ConsumerName:          cn,
		IntentName:            i,
		Project:               p,
		CompositeApp:          ca,
		Version:               v,
		DeploymentIntentGroup: di,
	}

	result, err := db.DBconn.Find(c.db.StoreName, dbKey, c.db.TagMetaData)
	if err != nil {
		log.Error("GetConsumer ... DB Error .. Get Consumer error", log.Fields{"consumer-name": cn})
		return hpaModel.HpaResourceConsumer{}, false, pkgerrors.Wrap(err, "DB Error .. Get Consumer error")
	}

	if result != nil {
		a := hpaModel.HpaResourceConsumer{}
		err = db.DBconn.Unmarshal(result[0], &a)
		if err != nil {
			log.Error("GetConsumer ... Unmarshalling  HpaConsumer error", log.Fields{"consumer-name": cn})
			return hpaModel.HpaResourceConsumer{}, false, pkgerrors.Wrap(err, "Unmarshalling  HpaConsumer error")
		}
		return a, false, nil
	}
	return hpaModel.HpaResourceConsumer{}, false, pkgerrors.New("Error getting Consumer")
}

/*
GetAllConsumers ... takes in projectName, CompositeAppName, CompositeAppVersion, DeploymentGroup,
DeploymentIntentName . It returns ListOfConsumers.
*/
func (c HpaPlacementClient) GetAllConsumers(p, ca, v, di, i string) ([]hpaModel.HpaResourceConsumer, error) {
	// check whether dependencies are met
	err := CheckPlacementConsumerDependency(p, ca, v, di, i)
	if err != nil {
		return []hpaModel.HpaResourceConsumer{}, err
	}

	dbKey := HpaConsumerKey{
		ConsumerName:          "",
		IntentName:            i,
		Project:               p,
		CompositeApp:          ca,
		Version:               v,
		DeploymentIntentGroup: di,
	}

	result, err := db.DBconn.Find(c.db.StoreName, dbKey, c.db.TagMetaData)
	if err != nil {
		log.Error("GetAllConsumers ... DB Error .. Get HpaConsumers db error", log.Fields{"intent-name": i})
		return []hpaModel.HpaResourceConsumer{}, pkgerrors.Wrap(err, "DB Error .. Get HpaConsumers db error")
	}
	log.Info("GetAllConsumers ... db result", log.Fields{"StoreName": c.db.StoreName, "key": dbKey, "project": p, "composite-app": ca, "composite-app-ver": v, "dep-group": di})

	var listOfMapOfConsumers []hpaModel.HpaResourceConsumer
	for i := range result {
		a := hpaModel.HpaResourceConsumer{}
		if result[i] != nil {
			err = db.DBconn.Unmarshal(result[i], &a)
			if err != nil {
				log.Error("GetAllConsumers ... Unmarshalling Consumer error.", log.Fields{"index": i, "consumer": result[i], "err": err})
				return []hpaModel.HpaResourceConsumer{}, pkgerrors.Wrap(err, "Unmarshalling Consumer error")
			}
			listOfMapOfConsumers = append(listOfMapOfConsumers, a)
		}
	}

	return listOfMapOfConsumers, nil
}

/*
GetConsumerByName ... takes in IntentName, projectName, CompositeAppName, CompositeAppVersion,
deploymentIntentGroupName and intentName returns the list of consumers under the IntentName.
*/
func (c HpaPlacementClient) GetConsumerByName(cn, p, ca, v, di, i string) (hpaModel.HpaResourceConsumer, error) {
	// check whether dependencies are met
	err := CheckPlacementConsumerDependency(p, ca, v, di, i)
	if err != nil {
		return hpaModel.HpaResourceConsumer{}, err
	}

	dbKey := HpaConsumerKey{
		ConsumerName:          cn,
		IntentName:            i,
		Project:               p,
		CompositeApp:          ca,
		Version:               v,
		DeploymentIntentGroup: di,
	}

	result, err := db.DBconn.Find(c.db.StoreName, dbKey, c.db.TagMetaData)
	if err != nil {
		log.Error("GetConsumerByName ... DB Error .. Get HpaConsumer error", log.Fields{"consumer-name": cn})
		return hpaModel.HpaResourceConsumer{}, pkgerrors.Wrap(err, "DB Error .. Get HpaConsumer error")
	}
	var a hpaModel.HpaResourceConsumer
	err = db.DBconn.Unmarshal(result[0], &a)
	if err != nil {
		log.Error("GetConsumerByName ... Unmarshalling Consumer error", log.Fields{"consumer-name": cn})
		return hpaModel.HpaResourceConsumer{}, pkgerrors.Wrap(err, "Unmarshalling Consumer error")
	}
	return a, nil
}

// DeleteConsumer ... deletes a given intent consumer tied to project, composite app and deployment intent group, intent name
func (c HpaPlacementClient) DeleteConsumer(cn, p string, ca string, v string, di string, i string) error {
	dbKey := HpaConsumerKey{
		ConsumerName:          cn,
		IntentName:            i,
		Project:               p,
		CompositeApp:          ca,
		Version:               v,
		DeploymentIntentGroup: di,
	}

	//Check for the Consumer already exists
	_, _, err := c.GetConsumer(cn, p, ca, v, di, i)
	if err != nil {
		log.Error("DeleteConsumer ... Consumer does not exist", log.Fields{"consumer-name": cn, "err": err})
		return pkgerrors.Wrapf(err, "Consumer[%s] does not exist", cn)
	}

	log.Info("DeleteConsumer ... Delete Hpa Consumer entry", log.Fields{"StoreName": c.db.StoreName, "key": dbKey, "project": p, "composite-app": ca, "composite-app-ver": v, "dep-group": di, "intent-name": i, "consumer-name": cn})
	err = db.DBconn.Remove(c.db.StoreName, dbKey)
	if err != nil {
		log.Error("DeleteConsumer ... DB Error .. Delete Hpa Consumer entry error", log.Fields{"err": err, "StoreName": c.db.StoreName, "key": dbKey, "project": p, "composite-app": ca, "composite-app-ver": v, "dep-group": di, "intent-name": i, "consumer-name": cn})
		return pkgerrors.Wrap(err, "DB Error .. Delete Hpa Consumer entry error")
	}
	return nil
}

// CheckPlacementConsumerDependency  ... check whether consumer(example: project, composite app..) dependncies exist
func CheckPlacementConsumerDependency(p string, ca string, v string, di string, i string) error {

	//Check intent dependencies
	err := CheckPlacementIntentDependency(p, ca, v, di)
	if err != nil {
		return err
	}

	// check if the intetName exists
	_, _, err = NewHpaPlacementClient().GetIntent(i, p, ca, v, di)
	if err != nil {
		return pkgerrors.Wrapf(err, "dependency not found .. Unable to find the intent-name[%v]", i)
	}

	return nil
}
