// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2020 Intel Corporation

package module

/*
This file deals with the interaction of instantiation flow and etcd.
It contains methods for creating appContext, saving cluster and resource details to etcd.

*/
import (
	"encoding/json"
	"io/ioutil"
	"time"

	"github.com/open-ness/EMCO/src/orchestrator/pkg/appcontext"
	gpic "github.com/open-ness/EMCO/src/orchestrator/pkg/gpic"
	"github.com/open-ness/EMCO/src/orchestrator/pkg/infra/db"
	log "github.com/open-ness/EMCO/src/orchestrator/pkg/infra/logutils"
	"github.com/open-ness/EMCO/src/orchestrator/pkg/state"
	"github.com/open-ness/EMCO/src/orchestrator/utils"
	"github.com/open-ness/EMCO/src/orchestrator/utils/helm"
	pkgerrors "github.com/pkg/errors"
)

// resource consists of name of reource
type resource struct {
	name        string
	filecontent string
}

type contextForCompositeApp struct {
	context            appcontext.AppContext
	ctxval             interface{}
	compositeAppHandle interface{}
}

// TODO move into a better place or reuse existing struct
type K8sResource struct {
	Metadata MetadataList `yaml:"metadata"`
}

// TODO move into a better place or reuse existing struct
type MetadataList struct {
	Namespace string `yaml:"namespace"`
}

type appOrderInstr struct {
	Apporder []string `json:"apporder"`
}

type appDepInstr struct {
	AppDepMap map[string]string `json:"appdependency"`
}

// makeAppContext creates an appContext for a compositeApp and returns the output as contextForCompositeApp
func makeAppContextForCompositeApp(p, ca, v, rName, dig string, namespace string, level string) (contextForCompositeApp, error) {
	context := appcontext.AppContext{}
	ctxval, err := context.InitAppContext()
	if err != nil {
		return contextForCompositeApp{}, pkgerrors.Wrap(err, "Error creating AppContext CompositeApp")
	}
	compositeHandle, err := context.CreateCompositeApp()
	if err != nil {
		return contextForCompositeApp{}, pkgerrors.Wrap(err, "Error creating CompositeApp handle")
	}
	err = context.AddCompositeAppMeta(appcontext.CompositeAppMeta{Project: p, CompositeApp: ca, Version: v, Release: rName, DeploymentIntentGroup: dig, Namespace: namespace, Level: level})
	if err != nil {
		return contextForCompositeApp{}, pkgerrors.Wrap(err, "Error Adding CompositeAppMeta")
	}

	m, err := context.GetCompositeAppMeta()

	log.Info(":: The meta data stored in the runtime context :: ", log.Fields{"Project": m.Project, "CompositeApp": m.CompositeApp, "Version": m.Version, "Release": m.Release, "DeploymentIntentGroup": m.DeploymentIntentGroup})

	cca := contextForCompositeApp{context: context, ctxval: ctxval, compositeAppHandle: compositeHandle}

	return cca, nil

}

// deleteAppContext removes an appcontext
func deleteAppContext(ct appcontext.AppContext) error {
	err := ct.DeleteCompositeApp()
	if err != nil {
		log.Warn(":: Error deleting AppContext ::", log.Fields{"Error": err})
		return pkgerrors.Wrapf(err, "Error Deleteing AppContext")
	}
	return nil
}

// getResources shall take in the sorted templates and output the resources
// which consists of name(name+kind) and filecontent
func getResources(st []helm.KubernetesResourceTemplate) ([]resource, error) {
	var resources []resource
	for _, t := range st {
		yamlStruct, err := utils.ExtractYamlParameters(t.FilePath)
		yamlFile, err := ioutil.ReadFile(t.FilePath)
		if err != nil {
			return nil, pkgerrors.Wrap(err, "Failed to get the resources..")
		}
		n := yamlStruct.Metadata.Name + SEPARATOR + yamlStruct.Kind
		// This might happen when the rendered file just has some comments inside, no real k8s object.
		if n == SEPARATOR {
			log.Info(":: Ignoring, Unable to render the template ::", log.Fields{"YAML PATH": t.FilePath})
			continue
		}

		resources = append(resources, resource{name: n, filecontent: string(yamlFile)})

		log.Info(":: Added resource into resource-order ::", log.Fields{"ResourceName": n})
	}
	return resources, nil
}

func addResourcesToCluster(ct appcontext.AppContext, ch interface{}, resources []resource, namespace string) error {

	var resOrderInstr struct {
		Resorder []string `json:"resorder"`
	}

	var resDepInstr struct {
		Resdep map[string]string `json:"resdependency"`
	}
	resdep := make(map[string]string)

	for _, resource := range resources {
		log.Info(":: RESOURCE ::", log.Fields{"filecontent": resource.filecontent})

		// // parse filecontent and replace namespace
		// yamlFile := K8sResource{}
		// // unmarshalling properly so there's no doubt it's metadata/namespace:
		// _ = yaml.Unmarshal([]byte(resource.filecontent), &yamlFile)
		// log.Info(":: yaml ::", log.Fields{"yamlFile": yamlFile.Metadata})
		// // string-based replace to avoid full unmarshal/marshal overhead:
		// // TODO: document rare issues: providing yaml not in lower-case; not respecting the single-space separator
		// resource.filecontent = strings.Replace(resource.filecontent, "namespace: "+yamlFile.Metadata.Namespace, "namespace: "+namespace, 1)
		// //

		resOrderInstr.Resorder = append(resOrderInstr.Resorder, resource.name)
		resdep[resource.name] = "go"
		_, err := ct.AddResource(ch, resource.name, resource.filecontent)
		if err != nil {
			cleanuperr := ct.DeleteCompositeApp()
			if cleanuperr != nil {
				log.Info(":: Error Cleaning up AppContext after add resource failure ::", log.Fields{"Resource": resource.name, "Error": cleanuperr.Error})
			}
			return pkgerrors.Wrapf(err, "Error adding resource ::%s to AppContext", resource.name)
		}
		jresOrderInstr, _ := json.Marshal(resOrderInstr)
		resDepInstr.Resdep = resdep
		jresDepInstr, _ := json.Marshal(resDepInstr)
		_, err = ct.AddInstruction(ch, "resource", "order", string(jresOrderInstr))
		_, err = ct.AddInstruction(ch, "resource", "dependency", string(jresDepInstr))
		if err != nil {
			cleanuperr := ct.DeleteCompositeApp()
			if cleanuperr != nil {
				log.Info(":: Error Cleaning up AppContext after add instruction failure ::", log.Fields{"Resource": resource.name, "Error": cleanuperr.Error})
			}
			return pkgerrors.Wrapf(err, "Error adding instruction for resource ::%s to AppContext", resource.name)
		}
	}
	return nil
}

//addClustersToAppContext method shall add cluster details save into etcd
func addClustersToAppContextHelper(cg []gpic.ClusterGroup, ct appcontext.AppContext, appHandle interface{}, resources []resource, namespace string) error {
	for _, eachGrp := range cg {
		oc := eachGrp.Clusters
		gn := eachGrp.GroupNumber

		for _, eachCluster := range oc {
			p := eachCluster.ProviderName
			n := eachCluster.ClusterName

			clusterhandle, err := ct.AddCluster(appHandle, p+SEPARATOR+n)

			if err != nil {
				cleanuperr := ct.DeleteCompositeApp()
				if cleanuperr != nil {
					log.Info(":: Error Cleaning up AppContext after add cluster failure ::", log.Fields{"cluster-provider": p, "cluster-name": n, "GroupName": gn, "Error": cleanuperr.Error})
				}
				return pkgerrors.Wrapf(err, "Error adding Cluster(provider::%s and name::%s) to AppContext", p, n)
			}
			log.Info(":: Added cluster ::", log.Fields{"Cluster ": p + SEPARATOR + n})

			err = ct.AddClusterMetaGrp(clusterhandle, gn)
			if err != nil {
				cleanuperr := ct.DeleteCompositeApp()
				if cleanuperr != nil {
					log.Info(":: Error Cleaning up AppContext after add cluster failure ::", log.Fields{"cluster-provider": p, "cluster-name": n, "GroupName": gn, "Error": cleanuperr.Error})
				}
				return pkgerrors.Wrapf(err, "Error adding Cluster(provider::%s and name::%s) to AppContext", p, n)
			}
			log.Info(":: Added cluster ::", log.Fields{"Cluster ": p + SEPARATOR + n, "GroupNumber ": gn})

			err = addResourcesToCluster(ct, clusterhandle, resources, namespace)
			if err != nil {
				return pkgerrors.Wrapf(err, "Error adding Resources to Cluster(provider::%s, name::%s and groupName:: %s) to AppContext", p, n, gn)
			}
		}
	}
	return nil
}

func addClustersToAppContext(l gpic.ClusterList, ct appcontext.AppContext, appHandle interface{}, resources []resource, namespace string) error {
	mClusters := l.MandatoryClusters
	oClusters := l.OptionalClusters

	err := addClustersToAppContextHelper(mClusters, ct, appHandle, resources, namespace)
	if err != nil {
		return err
	}
	log.Info("::Added mandatory clusters to the AppContext", log.Fields{})

	err = addClustersToAppContextHelper(oClusters, ct, appHandle, resources, namespace)
	if err != nil {
		return err
	}
	log.Info("::Added optional clusters to the AppContext", log.Fields{})
	return nil
}

/*
verifyResources method is just to check if the resource handles are correctly saved.
*/
func verifyResources(l gpic.ClusterList, ct appcontext.AppContext, resources []resource, appName string) error {

	for _, cg := range l.OptionalClusters {
		gn := cg.GroupNumber
		oc := cg.Clusters
		for _, eachCluster := range oc {
			p := eachCluster.ProviderName
			n := eachCluster.ClusterName
			cn := p + SEPARATOR + n

			for _, res := range resources {
				rh, err := ct.GetResourceHandle(appName, cn, res.name)
				if err != nil {
					return pkgerrors.Wrapf(err, "Error getting resource handle for resource :: %s, app:: %s, cluster :: %s, groupName :: %s", appName, res.name, cn, gn)
				}
				log.Info(":: GetResourceHandle ::", log.Fields{"ResourceHandler": rh, "appName": appName, "Cluster": cn, "Resource": res.name})
			}
		}
		grpMap, err := ct.GetClusterGroupMap(appName)
		if err != nil {
			return pkgerrors.Wrapf(err, "Error getting GetGroupMap for app:: %s, groupName :: %s", appName, gn)
		}
		log.Info(":: GetGroupMapReults ::", log.Fields{"GroupMap": grpMap})
	}

	for _, mClusters := range l.MandatoryClusters {
		for _, mc := range mClusters.Clusters {
			p := mc.ProviderName
			n := mc.ClusterName
			cn := p + SEPARATOR + n
			for _, res := range resources {
				rh, err := ct.GetResourceHandle(appName, cn, res.name)
				if err != nil {
					return pkgerrors.Wrapf(err, "Error getting resoure handle for resource :: %s, app:: %s, cluster :: %s", appName, res.name, cn)
				}
				log.Info(":: GetResourceHandle ::", log.Fields{"ResourceHandler": rh, "appName": appName, "Cluster": cn, "Resource": res.name})
			}
		}
	}
	return nil
}

func storeAppContextIntoRunTimeDB(allApps []App, cxtForCApp contextForCompositeApp, overrideValues []OverrideValues, dcmClusters []Cluster, p, ca, v, rName, cp, gIntent, di, namespace string) error {

	context := cxtForCApp.context
	// for recording the app order instruction
	var appOrdInsStr appOrderInstr
	// for recording the app dependency
	var appDepStr appDepInstr
	appDepStr.AppDepMap = make(map[string]string)

	for _, eachApp := range allApps {
		appOrdInsStr.Apporder = append(appOrdInsStr.Apporder, eachApp.Metadata.Name)
		appDepStr.AppDepMap[eachApp.Metadata.Name] = "go"

		sortedTemplates, err := GetSortedTemplateForApp(eachApp.Metadata.Name, p, ca, v, rName, cp, namespace, overrideValues)

		if err != nil {
			deleteAppContext(context)
			log.Error("Unable to get the sorted templates for app", log.Fields{"AppName": eachApp.Metadata.Name})
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
		if listOfClusters.MandatoryClusters == nil && listOfClusters.OptionalClusters == nil {
			deleteAppContext(context)
			log.Error("No compatible clusters have been provided to the Deployment Intent Group", log.Fields{"listOfClusters": listOfClusters})
			return pkgerrors.New("No compatible clusters have been provided to the Deployment Intent Group")
		}

		if err := checkClusters(listOfClusters, dcmClusters); err != nil {
			return err
		}

		//BEGIN: storing into etcd
		// Add an app to the app context
		apphandle, err := context.AddApp(cxtForCApp.compositeAppHandle, eachApp.Metadata.Name)
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
	jappOrderInstr, err := json.Marshal(appOrdInsStr)
	if err != nil {
		deleteAppContext(context)
		return pkgerrors.Wrap(err, "Error marshalling app order instruction")
	}

	jappDepInstr, err := json.Marshal(appDepStr.AppDepMap)
	if err != nil {
		deleteAppContext(context)
		return pkgerrors.Wrap(err, "Error marshalling app dependency instruction")
	}
	_, err = context.AddInstruction(cxtForCApp.compositeAppHandle, "app", "order", string(jappOrderInstr))
	if err != nil {
		deleteAppContext(context)
		return pkgerrors.Wrap(err, "Error adding app dependency instruction")
	}
	_, err = context.AddInstruction(cxtForCApp.compositeAppHandle, "app", "dependency", string(jappDepInstr))
	if err != nil {
		deleteAppContext(context)
		return pkgerrors.Wrap(err, "Error adding app dependency instruction")
	}
	//END: storing into etcd

	return nil
}

func storeAppContextIntoMetaDB(ctxval interface{}, storeName string, colName string, s state.StateInfo, p, ca, v, di string) error {

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
		Revision:  1,
	}
	s.StatusContextId = ctxval.(string)
	s.Actions = append(s.Actions, a)
	err := db.DBconn.Insert(storeName, key, nil, colName, s)
	if err != nil {
		log.Warn(":: Error updating DeploymentIntentGroup state in DB ::", log.Fields{"Error": err.Error(), "DeploymentIntentGroup": di, "CompositeApp": ca, "CompositeAppVersion": v, "Project": p, "AppContext": ctxval.(string)})
		return pkgerrors.Wrap(err, "Error adding DeploymentIntentGroup state to DB")
	}
	// END:: save the context in the orchestrator db record
	return nil
}

func handleStateInfo(p, ca, v, di string) (state.StateInfo, error) {

	s, err := NewDeploymentIntentGroupClient().GetDeploymentIntentGroupState(di, p, ca, v)
	if err != nil {
		return state.StateInfo{}, pkgerrors.Errorf("Error retrieving DeploymentIntentGroup stateInfo: " + di)
	}

	stateVal, err := state.GetCurrentStateFromStateInfo(s)
	if err != nil {
		return state.StateInfo{}, pkgerrors.Errorf("Error getting current state from DeploymentIntentGroup stateInfo: " + di)
	}
	switch stateVal {
	case state.StateEnum.Approved:
		break
	case state.StateEnum.Terminated:
		break // TODO - ideally, should check that all resources have completed being terminated
	case state.StateEnum.TerminateStopped:
		break
	case state.StateEnum.Created:
		return state.StateInfo{}, pkgerrors.Errorf("DeploymentIntentGroup must be Approved before instantiating" + di)
	case state.StateEnum.Applied:
		return state.StateInfo{}, pkgerrors.Errorf("DeploymentIntentGroup is in an invalid state" + di)
	case state.StateEnum.InstantiateStopped:
		return state.StateInfo{}, pkgerrors.Errorf("DeploymentIntentGroup has already been instantiated and stopped" + di)
	case state.StateEnum.Instantiated:
		return state.StateInfo{}, pkgerrors.Errorf("DeploymentIntentGroup has already been instantiated" + di)
	default:
		return state.StateInfo{}, pkgerrors.Errorf("DeploymentIntentGroup is in an unknown state" + stateVal)
	}
	return s, nil
}
