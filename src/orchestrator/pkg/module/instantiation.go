// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2020 Intel Corporation

package module

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/open-ness/EMCO/src/orchestrator/pkg/appcontext"
	gpic "github.com/open-ness/EMCO/src/orchestrator/pkg/gpic"
	"github.com/open-ness/EMCO/src/orchestrator/pkg/infra/db"
	log "github.com/open-ness/EMCO/src/orchestrator/pkg/infra/logutils"
	"github.com/open-ness/EMCO/src/orchestrator/pkg/state"
	"github.com/open-ness/EMCO/src/orchestrator/pkg/status"
	"github.com/open-ness/EMCO/src/orchestrator/utils/helm"
	pkgerrors "github.com/pkg/errors"
)

// ManifestFileName is the name given to the manifest file in the profile package
const ManifestFileName = "manifest.yaml"

// GenericPlacementIntentName denotes the generic placement intent name
const GenericPlacementIntentName = "genericPlacementIntent"

// SEPARATOR used while creating clusternames to store in etcd
const SEPARATOR = "+"

// InstantiationClient implements the InstantiationManager
type InstantiationClient struct {
	db InstantiationClientDbInfo
}

type DeploymentStatus struct {
	Project              string `json:"project,omitempty"`
	CompositeAppName     string `json:"composite-app-name,omitempty"`
	CompositeAppVersion  string `json:"composite-app-version,omitempty"`
	CompositeProfileName string `json:"composite-profile-name,omitempty"`
	status.StatusResult  `json:",inline"`
}

/*
InstantiationKey used in storing the contextid in the momgodb
It consists of
ProjectName,
CompositeAppName,
CompositeAppVersion,
DeploymentIntentGroup
*/
type InstantiationKey struct {
	Project               string
	CompositeApp          string
	Version               string
	DeploymentIntentGroup string
}

// InstantiationManager is an interface which exposes the
// InstantiationManager functionalities
type InstantiationManager interface {
	Approve(p string, ca string, v string, di string) error
	Instantiate(p string, ca string, v string, di string) error
	Status(p, ca, v, di, qInstance, qType, qOutput string, qApps, qClusters, qResources []string) (DeploymentStatus, error)
	Terminate(p string, ca string, v string, di string) error
	Stop(p string, ca string, v string, di string) error
}

// InstantiationClientDbInfo consists of storeName and tagState
type InstantiationClientDbInfo struct {
	storeName string // name of the mongodb collection to use for Instantiationclient documents
	tagState  string // attribute key name for context object in App Context
}

// NewInstantiationClient returns an instance of InstantiationClient
func NewInstantiationClient() *InstantiationClient {
	return &InstantiationClient{
		db: InstantiationClientDbInfo{
			storeName: "orchestrator",
			tagState:  "stateInfo",
		},
	}
}

//Approve approves an instantiation
func (c InstantiationClient) Approve(p string, ca string, v string, di string) error {
	s, err := NewDeploymentIntentGroupClient().GetDeploymentIntentGroupState(di, p, ca, v)
	if err != nil {
		log.Info("DeploymentIntentGroup has no state info ", log.Fields{"DeploymentIntentGroup: ": di})
		return pkgerrors.Wrap(err, "DeploymentIntentGroup has no state info: "+di)
	}
	stateVal, err := state.GetCurrentStateFromStateInfo(s)
	if err != nil {
		log.Info("Error getting current state from DeploymentIntentGroup stateInfo", log.Fields{"DeploymentIntentGroup ": di})
		return pkgerrors.Errorf("Error getting current state from DeploymentIntentGroup stateInfo: " + di)
	}
	switch stateVal {
	case state.StateEnum.Approved:
		return nil
	case state.StateEnum.Terminated:
		break
	case state.StateEnum.Created:
		break
	case state.StateEnum.Applied:
		return pkgerrors.Errorf("DeploymentIntentGroup is in an invalid state" + stateVal)
	case state.StateEnum.Instantiated:
		return pkgerrors.Errorf("DeploymentIntentGroup has already been instantiated" + di)
	default:
		return pkgerrors.Errorf("DeploymentIntentGroup is in an unknown state" + stateVal)
	}

	key := DeploymentIntentGroupKey{
		Name:         di,
		Project:      p,
		CompositeApp: ca,
		Version:      v,
	}
	a := state.ActionEntry{
		State:     state.StateEnum.Approved,
		ContextId: "",
		TimeStamp: time.Now(),
	}
	s.Actions = append(s.Actions, a)

	err = db.DBconn.Insert(c.db.storeName, key, nil, c.db.tagState, s)
	if err != nil {
		return pkgerrors.Wrap(err, "Error updating the stateInfo of the DeploymentIntentGroup: "+di)
	}

	return nil
}

func getOverrideValuesByAppName(ov []OverrideValues, a string) map[string]string {
	for _, eachOverrideVal := range ov {
		if eachOverrideVal.AppName == a {
			return eachOverrideVal.ValuesObj
		}
	}
	return map[string]string{}
}

/*
findGenericPlacementIntent takes in projectName, CompositeAppName, CompositeAppVersion, DeploymentIntentName
and returns the name of the genericPlacementIntentName. Returns empty value if string not found.
*/
func findGenericPlacementIntent(p, ca, v, di string) (string, error) {
	var gi string
	iList, err := NewIntentClient().GetAllIntents(p, ca, v, di)
	if err != nil {
		return gi, err
	}
	for _, eachMap := range iList.ListOfIntents {
		if gi, found := eachMap[GenericPlacementIntentName]; found {
			log.Info(":: Name of the generic-placement-intent found ::", log.Fields{"GenPlmtIntent": gi})
			return gi, nil
		}
	}
	log.Info(":: generic-placement-intent not found ! ::", log.Fields{"Searched for GenPlmtIntent": GenericPlacementIntentName})
	return gi, pkgerrors.New("Generic-placement-intent not found")
}

// GetSortedTemplateForApp returns the sorted templates.
//It takes in arguments - appName, project, compositeAppName, releaseName, compositeProfileName, array of override values
func GetSortedTemplateForApp(appName, p, ca, v, rName, cp string, overrideValues []OverrideValues) ([]helm.KubernetesResourceTemplate, error) {

	log.Info(":: Processing App ::", log.Fields{"appName": appName})

	var sortedTemplates []helm.KubernetesResourceTemplate

	aC, err := NewAppClient().GetAppContent(appName, p, ca, v)
	if err != nil {
		return sortedTemplates, pkgerrors.Wrap(err, fmt.Sprint("Not finding the content of app:: ", appName))
	}
	appContent, err := base64.StdEncoding.DecodeString(aC.FileContent)
	if err != nil {
		return sortedTemplates, pkgerrors.Wrap(err, "Fail to convert to byte array")
	}

	log.Info(":: Got the app content.. ::", log.Fields{"appName": appName})

	appPC, err := NewAppProfileClient().GetAppProfileContentByApp(p, ca, v, cp, appName)
	if err != nil {
		return sortedTemplates, pkgerrors.Wrap(err, fmt.Sprintf("Not finding the appProfileContent for:: %s", appName))
	}
	appProfileContent, err := base64.StdEncoding.DecodeString(appPC.Profile)
	if err != nil {
		return sortedTemplates, pkgerrors.Wrap(err, "Fail to convert to byte array")
	}

	log.Info(":: Got the app Profile content .. ::", log.Fields{"appName": appName})

	overrideValuesOfApp := getOverrideValuesByAppName(overrideValues, appName)
	//Convert override values from map to array of strings of the following format
	//foo=bar
	overrideValuesOfAppStr := []string{}
	if overrideValuesOfApp != nil {
		for k, v := range overrideValuesOfApp {
			overrideValuesOfAppStr = append(overrideValuesOfAppStr, k+"="+v)
		}
	}

	sortedTemplates, err = helm.NewTemplateClient("", "default", rName,
		ManifestFileName).Resolve(appContent,
		appProfileContent, overrideValuesOfAppStr,
		appName)

	log.Info(":: Total no. of sorted templates ::", log.Fields{"len(sortedTemplates):": len(sortedTemplates)})

	return sortedTemplates, err
}

func calculateDirPath(fp string) string {
	sa := strings.Split(fp, "/")
	return "/" + sa[1] + "/" + sa[2] + "/"
}

func cleanTmpfiles(sortedTemplates []helm.KubernetesResourceTemplate) error {
	dp := calculateDirPath(sortedTemplates[0].FilePath)
	for _, st := range sortedTemplates {
		log.Info("Clean up ::", log.Fields{"file: ": st.FilePath})
		err := os.Remove(st.FilePath)
		if err != nil {
			log.Error("Error while deleting file", log.Fields{"file: ": st.FilePath})
			return err
		}
	}
	err := os.RemoveAll(dp)
	if err != nil {
		log.Error("Error while deleting dir", log.Fields{"Dir: ": dp})
		return err
	}
	log.Info("Clean up temp-dir::", log.Fields{"Dir: ": dp})
	return nil
}

/*
Instantiate methods takes in projectName, compositeAppName, compositeAppVersion,
DeploymentIntentName. This method is responsible for template resolution, intent
resolution, creation and saving of context for saving into etcd.
*/
func (c InstantiationClient) Instantiate(p string, ca string, v string, di string) error {

	dIGrp, err := NewDeploymentIntentGroupClient().GetDeploymentIntentGroup(di, p, ca, v)
	if err != nil {
		return pkgerrors.Wrap(err, "Not finding the deploymentIntentGroup")
	}

	s, err := NewDeploymentIntentGroupClient().GetDeploymentIntentGroupState(di, p, ca, v)
	if err != nil {
		return pkgerrors.Errorf("Error retrieving DeploymentIntentGroup stateInfo: " + di)
	}
	stateVal, err := state.GetCurrentStateFromStateInfo(s)
	if err != nil {
		return pkgerrors.Errorf("Error getting current state from DeploymentIntentGroup stateInfo: " + di)
	}
	switch stateVal {
	case state.StateEnum.Approved:
		break
	case state.StateEnum.Terminated:
		break // TODO - ideally, should check that all resources have completed being terminated
	case state.StateEnum.TerminateStopped:
		break
	case state.StateEnum.Created:
		return pkgerrors.Errorf("DeploymentIntentGroup must be Approved before instantiating" + di)
	case state.StateEnum.Applied:
		return pkgerrors.Errorf("DeploymentIntentGroup is in an invalid state" + di)
	case state.StateEnum.InstantiateStopped:
		return pkgerrors.Errorf("DeploymentIntentGroup has already been instantiated and stopped" + di)
	case state.StateEnum.Instantiated:
		return pkgerrors.Errorf("DeploymentIntentGroup has already been instantiated" + di)
	default:
		return pkgerrors.Errorf("DeploymentIntentGroup is in an unknown state" + stateVal)
	}

	rName := dIGrp.Spec.Version //rName is releaseName
	overrideValues := dIGrp.Spec.OverrideValuesObj
	cp := dIGrp.Spec.Profile
	lc := dIGrp.Spec.LogicalCloud

	gIntent, err := findGenericPlacementIntent(p, ca, v, di)
	if err != nil {
		return err
	}

	log.Info(":: The name of the GenPlacIntent ::", log.Fields{"GenPlmtIntent": gIntent})
	log.Info(":: DeploymentIntentGroup, ReleaseName, CompositeProfile ::", log.Fields{"dIGrp": dIGrp.MetaData.Name, "releaseName": rName, "cp": cp})

	allApps, err := NewAppClient().GetApps(p, ca, v)
	if err != nil {
		return pkgerrors.Wrap(err, "Not finding the apps")
	}

	// check that specified logical cloud is already instantiated
	dcmCloudClient := NewLogicalCloudClient()
	logicalCloud, err := dcmCloudClient.Get(p, lc)
	log.Info(":: logicalCloud ::", log.Fields{"logicalCloud": logicalCloud})
	if err != nil {
		log.Error("Failed to obtain Logical Cloud specified", log.Fields{"error": err.Error()})
		return pkgerrors.Wrap(err, "Failed to obtain Logical Cloud specified")
	}
	lckey := LogicalCloudKey{
		Project:          p,
		LogicalCloudName: lc,
	}
	ac, cid, err := dcmCloudClient.util.GetLogicalCloudContext(dcmCloudClient.storeName, lckey, dcmCloudClient.tagContext, p, lc)
	if err != nil {
		log.Error("Error reading Logical Cloud context", log.Fields{"error": err.Error()})
		return pkgerrors.Wrap(err, "Error reading Logical Cloud context")
	}
	if cid == "" {
		log.Error("The Logical Cloud has never been instantiated", log.Fields{"cid": cid})
		return pkgerrors.New("The Logical Cloud has never been instantiated")
	}

	// the namespace where the resources of this app are supposed to deployed to
	namespace := logicalCloud.Specification.NameSpace
	log.Info("Namespace for this logical cloud", log.Fields{"namespace": namespace})
	// level of the logical cloud (0 - admin, 1 - custom)
	level := logicalCloud.Specification.Level

	// make sure rsync status for this logical cloud is Instantiated (instantiated),
	// otherwise the cloud isn't ready to receive the application being instantiated
	acStatus, err := dcmCloudClient.util.GetAppContextStatus(ac)
	if err != nil {
		return err
	}
	switch acStatus.Status {
	case appcontext.AppContextStatusEnum.Instantiated:
		log.Info("The Logical Cloud is instantiated, proceeding with DIG instantiation.", log.Fields{"logicalcloud": lc})
	case appcontext.AppContextStatusEnum.Terminated:
		log.Error("The Logical Cloud is not currently instantiated (has been terminated).", log.Fields{"logicalcloud": lc})
		return pkgerrors.New("The Logical Cloud is not currently instantiated (has been terminated).")
	case appcontext.AppContextStatusEnum.Instantiating:
		log.Error("The Logical Cloud is still instantiating, please wait and try again.", log.Fields{"logicalcloud": lc})
		return pkgerrors.New("The Logical Cloud is still instantiating, please wait and try again.")
	case appcontext.AppContextStatusEnum.Terminating:
		log.Error("The Logical Cloud is terminating, so it can no longer receive DIGs.", log.Fields{"logicalcloud": lc})
		return pkgerrors.New("The Logical Cloud is terminating, so it can no longer receive DIGs.")
	case appcontext.AppContextStatusEnum.InstantiateFailed:
		log.Error("The Logical Cloud has failed instanting, so it can't receive DIGs.", log.Fields{"logicalcloud": lc})
		return pkgerrors.New("The Logical Cloud has failed instanting, so it can't receive DIGs.")
	case appcontext.AppContextStatusEnum.TerminateFailed:
		log.Error("The Logical Cloud has failed terminating, so for safety it can no longer receive DIGs.", log.Fields{"logicalcloud": lc})
		return pkgerrors.New("The Logical Cloud has failed terminating, so for safety it can no longer receive DIGs.")
	default:
		log.Error("The Logical Cloud isn't in an expected status so not taking any action.", log.Fields{"logicalcloud": lc, "status": acStatus.Status})
		return pkgerrors.New("The Logical Cloud isn't in an expected status so not taking any action.")
	}

	// get all clusters from specified logical cloud (LC)
	// [candidate in the future for emco-lib]
	dcmClusterClient := NewClusterClient()
	dcmClusters, _ := dcmClusterClient.GetAllClusters(p, lc)
	log.Info(":: dcmClusters ::", log.Fields{"dcmClusters": dcmClusters})

	cca, err := makeAppContextForCompositeApp(p, ca, v, rName, di, namespace, level)
	if err != nil {
		return err
	}
	context := cca.context
	ctxval := cca.ctxval
	compositeHandle := cca.compositeAppHandle

	var appOrderInstr struct {
		Apporder []string `json:"apporder"`
	}

	var appDepInstr struct {
		Appdep map[string]string `json:"appdependency"`
	}
	appdep := make(map[string]string)

	// Add composite app using appContext
	for _, eachApp := range allApps {
		appOrderInstr.Apporder = append(appOrderInstr.Apporder, eachApp.Metadata.Name)
		appdep[eachApp.Metadata.Name] = "go"

		sortedTemplates, err := GetSortedTemplateForApp(eachApp.Metadata.Name, p, ca, v, rName, cp, overrideValues)

		if err != nil {
			deleteAppContext(context)
			log.Error("Unable to get the sorted templates for app", log.Fields{})
			return pkgerrors.Wrap(err, "Unable to get the sorted templates for app")
		}

		log.Info(":: Resolved all the templates ::", log.Fields{"appName": eachApp.Metadata.Name, "SortedTemplate": sortedTemplates})

		resources, err := getResources(sortedTemplates)
		if err != nil {
			deleteAppContext(context)
			return pkgerrors.Wrapf(err, "Unable to get the resources for app :: %s", eachApp.Metadata.Name)
		}

		defer cleanTmpfiles(sortedTemplates)

		specData, err := NewAppIntentClient().GetAllIntentsByApp(eachApp.Metadata.Name, p, ca, v, gIntent, di)
		if err != nil {
			deleteAppContext(context)
			return pkgerrors.Wrap(err, "Unable to get the intents for app")
		}
		// listOfClusters shall have both mandatoryClusters and optionalClusters where the app needs to be installed.
		listOfClusters, err := gpic.IntentResolver(specData.Intent)
		if err != nil {
			deleteAppContext(context)
			return pkgerrors.Wrap(err, "Unable to get the intents resolved for app")
		}

		log.Info(":: listOfClusters ::", log.Fields{"listOfClusters": listOfClusters})
		if listOfClusters.MandatoryClusters == nil && listOfClusters.ClusterGroups == nil {
			deleteAppContext(context)
			log.Error("No compatible clusters have been provided to the Deployment Intent Group", log.Fields{"listOfClusters": listOfClusters})
			return pkgerrors.New("No compatible clusters have been provided to the Deployment Intent Group")
		}

		// make sure LC can support DIG by validating DIG clusters against LC clusters
		mandatoryClusters := make([]gpic.ClusterWithName, len(listOfClusters.MandatoryClusters))
		_ = copy(mandatoryClusters, listOfClusters.MandatoryClusters)
		for _, dcluster := range dcmClusters {
			for i, cluster := range mandatoryClusters {
				if cluster.ProviderName == dcluster.Specification.ClusterProvider {
					// remove the cluster from slice since it's part of the LC
					lastIndex := len(mandatoryClusters) - 1
					mandatoryClusters[i] = mandatoryClusters[lastIndex]
					mandatoryClusters = mandatoryClusters[:lastIndex]
					// we're done checking this DCM cluster
					break
				}
			}
		}
		if len(mandatoryClusters) > 0 {
			log.Error("The specified Logical Cloud doesn't provide the mandatory clusters", log.Fields{"mandatoryClusters": mandatoryClusters})
			return pkgerrors.New("The specified Logical Cloud doesn't provide the mandatory clusters")
		}

		//BEGIN: storing into etcd
		// Add an app to the app context
		apphandle, err := context.AddApp(compositeHandle, eachApp.Metadata.Name)
		if err != nil {
			deleteAppContext(context)
			return pkgerrors.Wrap(err, "Error adding App to AppContext")
		}
		err = addClustersToAppContext(listOfClusters, context, apphandle, resources, namespace)
		if err != nil {
			deleteAppContext(context)
			return pkgerrors.Wrap(err, "Error while adding cluster and resources to app")
		}
		err = verifyResources(listOfClusters, context, resources, eachApp.Metadata.Name)
		if err != nil {
			deleteAppContext(context)
			return pkgerrors.Wrap(err, "Error while verifying resources in app: ")
		}

	}

	jappOrderInstr, err := json.Marshal(appOrderInstr)
	if err != nil {
		deleteAppContext(context)
		return pkgerrors.Wrap(err, "Error marshalling app order instruction")
	}
	appDepInstr.Appdep = appdep
	jappDepInstr, err := json.Marshal(appDepInstr)
	if err != nil {
		deleteAppContext(context)
		return pkgerrors.Wrap(err, "Error marshalling app dependency instruction")
	}
	_, err = context.AddInstruction(compositeHandle, "app", "order", string(jappOrderInstr))
	if err != nil {
		deleteAppContext(context)
		return pkgerrors.Wrap(err, "Error adding app dependency instruction")
	}
	_, err = context.AddInstruction(compositeHandle, "app", "dependency", string(jappDepInstr))
	if err != nil {
		deleteAppContext(context)
		return pkgerrors.Wrap(err, "Error adding app dependency instruction")
	}
	//END: storing into etcd

	// BEGIN: scheduler code

	pl, mapOfControllers, err := getPrioritizedControllerList(p, ca, v, di)
	if err != nil {
		return pkgerrors.Wrap(err, "Error adding getting prioritized controller list")
	}
	log.Info("Priority Based List ", log.Fields{"PlacementControllers::": pl.pPlaCont,
		"ActionControllers::": pl.pActCont, "mapOfControllers::": mapOfControllers})

	err = callGrpcForControllerList(pl.pPlaCont, mapOfControllers, ctxval)
	if err != nil {
		deleteAppContext(context)
		return pkgerrors.Wrap(err, "Error calling gRPC for placement controller list")
	}

	err = deleteExtraClusters(allApps, context)
	if err != nil {
		deleteAppContext(context)
		return pkgerrors.Wrap(err, "Error deleting extra clusters")
	}

	err = callGrpcForControllerList(pl.pActCont, mapOfControllers, ctxval)
	log.Warn("", log.Fields{"pl.pActCont::": pl.pActCont})
	log.Warn("", log.Fields{"mapOfControllers::": mapOfControllers})
	log.Warn("", log.Fields{"ctxval::": ctxval})
	if err != nil {
		deleteAppContext(context)
		return pkgerrors.Wrap(err, "Error calling gRPC for action controller list")
	}
	// END: Scheduler code

	// BEGIN : Rsync code
	err = callRsyncInstall(ctxval)
	if err != nil {
		deleteAppContext(context)
		return pkgerrors.Wrap(err, "Error calling rsync")
	}
	// END : Rsync code

	// BEGIN:: save the context in the orchestrator db record
	key := DeploymentIntentGroupKey{
		Name:         di,
		Project:      p,
		CompositeApp: ca,
		Version:      v,
	}
	a := state.ActionEntry{
		State:     state.StateEnum.Instantiated,
		ContextId: ctxval.(string),
		TimeStamp: time.Now(),
	}
	s.Actions = append(s.Actions, a)
	err = db.DBconn.Insert(c.db.storeName, key, nil, c.db.tagState, s)
	if err != nil {
		log.Warn(":: Error updating DeploymentIntentGroup state in DB ::", log.Fields{"Error": err.Error(), "GPIntent": gIntent, "DeploymentIntentGroup": di, "CompositeApp": ca, "CompositeAppVersion": v, "Project": p, "AppContext": ctxval.(string)})
		return pkgerrors.Wrap(err, "Error adding DeploymentIntentGroup state to DB")
	}
	// END:: save the context in the orchestrator db record

	log.Info(":: Done with instantiation call to rsync... ::", log.Fields{"CompositeAppName": ca})
	return err
}

/*
Status takes in projectName, compositeAppName, compositeAppVersion,
DeploymentIntentName. This method is responsible obtaining the status of
the deployment, which is made available in the appcontext.
*/
func (c InstantiationClient) Status(p, ca, v, di, qInstance, qType, qOutput string, qApps, qClusters, qResources []string) (DeploymentStatus, error) {

	dIGrp, err := NewDeploymentIntentGroupClient().GetDeploymentIntentGroup(di, p, ca, v)
	if err != nil {
		return DeploymentStatus{}, pkgerrors.Wrap(err, "Not finding the deploymentIntentGroup")
	}

	diState, err := NewDeploymentIntentGroupClient().GetDeploymentIntentGroupState(di, p, ca, v)
	if err != nil {
		return DeploymentStatus{}, pkgerrors.Wrap(err, "deploymentIntentGroup state not found: "+di)
	}

	// Get all apps in this composite app
	apps, err := NewAppClient().GetApps(p, ca, v)
	if err != nil {
		return DeploymentStatus{}, pkgerrors.Wrap(err, "Not finding the apps")
	}
	allApps := make([]string, 0)
	for _, a := range apps {
		allApps = append(allApps, a.Metadata.Name)
	}

	statusResponse, err := status.PrepareStatusResult(diState, allApps, qInstance, qType, qOutput, qApps, qClusters, qResources)
	if err != nil {
		return DeploymentStatus{}, err
	}
	statusResponse.Name = di
	diStatus := DeploymentStatus{
		Project:              p,
		CompositeAppName:     ca,
		CompositeAppVersion:  v,
		CompositeProfileName: dIGrp.Spec.Profile,
		StatusResult:         statusResponse,
	}

	return diStatus, nil
}

/*
Terminate takes in projectName, compositeAppName, compositeAppVersion,
DeploymentIntentName and calls rsync to terminate.
*/
func (c InstantiationClient) Terminate(p string, ca string, v string, di string) error {

	s, err := NewDeploymentIntentGroupClient().GetDeploymentIntentGroupState(di, p, ca, v)
	if err != nil {
		return pkgerrors.Wrap(err, "DeploymentIntentGroup has no state info: "+di)
	}

	stateVal, err := state.GetCurrentStateFromStateInfo(s)
	if err != nil {
		return pkgerrors.Errorf("Error getting current state from DeploymentIntentGroup stateInfo: " + di)
	}

	if stateVal != state.StateEnum.Instantiated && stateVal != state.StateEnum.InstantiateStopped {
		return pkgerrors.Errorf("DeploymentIntentGroup is not instantiated :" + di)
	}

	currentCtxId := state.GetLastContextIdFromStateInfo(s)
	err = callRsyncUninstall(currentCtxId)
	if err != nil {
		return err
	}

	key := DeploymentIntentGroupKey{
		Name:         di,
		Project:      p,
		CompositeApp: ca,
		Version:      v,
	}
	a := state.ActionEntry{
		State:     state.StateEnum.Terminated,
		ContextId: currentCtxId,
		TimeStamp: time.Now(),
	}
	s.Actions = append(s.Actions, a)

	err = db.DBconn.Insert(c.db.storeName, key, nil, c.db.tagState, s)
	if err != nil {
		return pkgerrors.Wrap(err, "Error updating the stateInfo of the DeploymentIntentGroup: "+di)
	}

	return nil
}

/*
Stop takes in projectName, compositeAppName, compositeAppVersion,
DeploymentIntentName and sets the stopFlag in the associated appContext.
*/
func (c InstantiationClient) Stop(p string, ca string, v string, di string) error {

	s, err := NewDeploymentIntentGroupClient().GetDeploymentIntentGroupState(di, p, ca, v)
	if err != nil {
		return pkgerrors.Wrap(err, "DeploymentIntentGroup has no state info: "+di)
	}

	stateVal, err := state.GetCurrentStateFromStateInfo(s)
	if err != nil {
		return pkgerrors.Errorf("Error getting current state from DeploymentIntentGroup stateInfo: " + di)
	}
	stopState := state.StateEnum.Undefined
	switch stateVal {
	case state.StateEnum.Approved:
		return pkgerrors.Errorf("DeploymentIntentGroup has not been instantiated:" + di)
	case state.StateEnum.Instantiated:
		stopState = state.StateEnum.InstantiateStopped
		break
	case state.StateEnum.Terminated:
		stopState = state.StateEnum.TerminateStopped
		break
	case state.StateEnum.Applied:
		return pkgerrors.Wrap(err, "DeploymentIntentGroup is in an invalid state:"+di)
		break
	case state.StateEnum.TerminateStopped:
		return pkgerrors.Wrap(err, "DeploymentIntentGroup termination already stopped: "+di)
	case state.StateEnum.InstantiateStopped:
		return pkgerrors.Wrap(err, "DeploymentIntentGroup instantiation already stopped: "+di)
	case state.StateEnum.Created:
		return pkgerrors.Wrap(err, "DeploymentIntentGroup have not been approved: "+di)
	default:
		return pkgerrors.Wrap(err, "DeploymentIntentGroup is in an invalid state: "+di+" "+stateVal)
	}

	currentCtxId := state.GetLastContextIdFromStateInfo(s)

	acStatus, err := state.GetAppContextStatus(currentCtxId)
	if err != nil {
		return err
	}
	if acStatus.Status != appcontext.AppContextStatusEnum.Instantiating &&
		acStatus.Status != appcontext.AppContextStatusEnum.Terminating {
		return pkgerrors.Errorf("DeploymentIntentGroup is not instantiating or terminating:" + di)
	}
	err = state.UpdateAppContextStopFlag(currentCtxId, true)
	if err != nil {
		return err
	}

	key := DeploymentIntentGroupKey{
		Name:         di,
		Project:      p,
		CompositeApp: ca,
		Version:      v,
	}
	a := state.ActionEntry{
		State:     stopState,
		ContextId: currentCtxId,
		TimeStamp: time.Now(),
	}
	s.Actions = append(s.Actions, a)

	err = db.DBconn.Insert(c.db.storeName, key, nil, c.db.tagState, s)
	if err != nil {
		return pkgerrors.Wrap(err, "Error updating the stateInfo of the DeploymentIntentGroup: "+di)
	}

	return nil
}
