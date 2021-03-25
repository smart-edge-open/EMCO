// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2020 Intel Corporation

package module

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"strconv"
	"strings"

	"github.com/open-ness/EMCO/src/orchestrator/pkg/appcontext"
	"github.com/open-ness/EMCO/src/orchestrator/pkg/appcontext/subresources"
	"github.com/open-ness/EMCO/src/orchestrator/pkg/grpc/installappclient"
	"github.com/open-ness/EMCO/src/orchestrator/pkg/infra/db"
	log "github.com/open-ness/EMCO/src/orchestrator/pkg/infra/logutils"
	"github.com/open-ness/EMCO/src/orchestrator/pkg/module/controller"
	rsync "github.com/open-ness/EMCO/src/rsync/pkg/db"
	pkgerrors "github.com/pkg/errors"
	"gopkg.in/yaml.v2"
	certificatesv1beta1 "k8s.io/api/certificates/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// rsyncName denotes the name of the rsync controller
const rsyncName = "rsync"

type Resource struct {
	ApiVersion    string         `yaml:"apiVersion"`
	Kind          string         `yaml:"kind"`
	MetaData      MetaDatas      `yaml:"metadata"`
	Specification Specs          `yaml:"spec,omitempty"`
	Rules         []RoleRules    `yaml:"rules,omitempty"`
	Subjects      []RoleSubjects `yaml:"subjects,omitempty"`
	RoleRefs      RoleRef        `yaml:"roleRef,omitempty"`
}

type MetaDatas struct {
	Name      string `yaml:"name"`
	Namespace string `yaml:"namespace,omitempty"`
}

type Specs struct {
	Request string   `yaml:"request,omitempty"`
	Usages  []string `yaml:"usages,omitempty"`
	// TODO: validate quota keys
	// //Hard           logicalcloud.QSpec    `yaml:"hard,omitempty"`
	// Hard QSpec `yaml:"hard,omitempty"`
	Hard map[string]string `yaml:"hard,omitempty"`
}

type RoleRules struct {
	ApiGroups []string `yaml:"apiGroups"`
	Resources []string `yaml:"resources"`
	Verbs     []string `yaml:"verbs"`
}

type RoleSubjects struct {
	Kind     string `yaml:"kind"`
	Name     string `yaml:"name"`
	ApiGroup string `yaml:"apiGroup"`
}

type RoleRef struct {
	Kind     string `yaml:"kind"`
	Name     string `yaml:"name"`
	ApiGroup string `yaml:"apiGroup"`
}

func cleanupCompositeApp(context appcontext.AppContext, err error, reason string, details []string) error {
	cleanuperr := context.DeleteCompositeApp()
	newerr := pkgerrors.Wrap(err, reason)
	if cleanuperr != nil {
		log.Warn("Error cleaning AppContext, ", log.Fields{
			"Related details": details,
		})
		// this would be useful: https://godoc.org/go.uber.org/multierr
		return pkgerrors.Wrap(err, "After previous error, cleaning the AppContext also failed.")
	}
	return newerr
}

func createNamespace(logicalcloud LogicalCloud) (string, string, error) {

	name := logicalcloud.Specification.NameSpace

	namespace := Resource{
		ApiVersion: "v1",
		Kind:       "Namespace",
		MetaData: MetaDatas{
			Name: name,
		},
	}

	nsData, err := yaml.Marshal(&namespace)
	if err != nil {
		return "", "", err
	}

	return string(nsData), strings.Join([]string{name, "+Namespace"}, ""), nil
}

func createRoles(logicalcloud LogicalCloud, userpermissions []UserPermission) ([]string, []string, error) {
	var name string
	var kind string
	var datas []string
	var names []string

	roleCount := len(userpermissions)
	datas = make([]string, roleCount, roleCount)
	names = make([]string, roleCount, roleCount)

	for i, up := range userpermissions {
		if up.Specification.Namespace == "" {
			name = strings.Join([]string{logicalcloud.MetaData.LogicalCloudName, "-clusterRole", strconv.Itoa(i)}, "")
			kind = "ClusterRole"
		} else {
			name = strings.Join([]string{logicalcloud.MetaData.LogicalCloudName, "-role", strconv.Itoa(i)}, "")
			kind = "Role"
		}

		role := Resource{
			ApiVersion: "rbac.authorization.k8s.io/v1beta1",
			Kind:       kind,
			MetaData: MetaDatas{
				Name: name,
				// Namespace: logicalcloud.Specification.NameSpace,
			},
			Rules: []RoleRules{RoleRules{
				ApiGroups: up.Specification.APIGroups,
				Resources: up.Specification.Resources,
				Verbs:     up.Specification.Verbs,
			},
			},
		}
		if up.Specification.Namespace != "" {
			role.MetaData.Namespace = up.Specification.Namespace
		}

		roleData, err := yaml.Marshal(&role)
		if err != nil {
			return []string{}, []string{}, err
		}

		datas[i] = string(roleData)
		names[i] = strings.Join([]string{name, "+", kind}, "")
	}

	return datas, names, nil
}

func createRoleBindings(logicalcloud LogicalCloud, userpermissions []UserPermission) ([]string, []string, error) {
	var name string
	var kind string
	var kindbinding string
	var datas []string
	var names []string

	roleCount := len(userpermissions)
	datas = make([]string, roleCount, roleCount)
	names = make([]string, roleCount, roleCount)

	for i, up := range userpermissions {
		if up.Specification.Namespace == "" {
			name = strings.Join([]string{logicalcloud.MetaData.LogicalCloudName, "-clusterRoleBinding", strconv.Itoa(i)}, "")
			kind = "ClusterRole"
			kindbinding = "ClusterRoleBinding"
		} else {
			name = strings.Join([]string{logicalcloud.MetaData.LogicalCloudName, "-roleBinding", strconv.Itoa(i)}, "")
			kind = "Role"
			kindbinding = "RoleBinding"
		}

		roleBinding := Resource{
			ApiVersion: "rbac.authorization.k8s.io/v1beta1",
			Kind:       kindbinding,
			MetaData: MetaDatas{
				Name: name,
			},
			Subjects: []RoleSubjects{RoleSubjects{
				Kind:     "User",
				Name:     logicalcloud.Specification.User.UserName,
				ApiGroup: "",
			},
			},

			RoleRefs: RoleRef{
				Kind:     kind,
				ApiGroup: "",
			},
		}
		if up.Specification.Namespace != "" {
			roleBinding.MetaData.Namespace = up.Specification.Namespace
			roleBinding.RoleRefs.Name = strings.Join([]string{logicalcloud.MetaData.LogicalCloudName, "-role", strconv.Itoa(i)}, "")
		} else {
			roleBinding.RoleRefs.Name = strings.Join([]string{logicalcloud.MetaData.LogicalCloudName, "-clusterRole", strconv.Itoa(i)}, "")
		}

		rBData, err := yaml.Marshal(&roleBinding)
		if err != nil {
			return []string{}, []string{}, err
		}
		datas[i] = string(rBData)
		names[i] = strings.Join([]string{name, "+", kindbinding}, "")
	}

	return datas, names, nil
}

func createQuota(quota []Quota, namespace string) (string, string, error) {

	lcQuota := quota[0]
	name := lcQuota.MetaData.QuotaName

	q := Resource{
		ApiVersion: "v1",
		Kind:       "ResourceQuota",
		MetaData: MetaDatas{
			Name:      name,
			Namespace: namespace,
		},
		Specification: Specs{
			Hard: lcQuota.Specification,
		},
	}

	qData, err := yaml.Marshal(&q)
	if err != nil {
		return "", "", err
	}

	return string(qData), strings.Join([]string{name, "+ResourceQuota"}, ""), nil
}

func createUserCSR(logicalcloud LogicalCloud) (string, string, string, error) {

	KEYSIZE := 4096
	userName := logicalcloud.Specification.User.UserName
	name := strings.Join([]string{logicalcloud.MetaData.LogicalCloudName, "-user-csr"}, "")

	key, err := rsa.GenerateKey(rand.Reader, KEYSIZE)
	if err != nil {
		return "", "", "", err
	}

	csrTemplate := x509.CertificateRequest{Subject: pkix.Name{CommonName: userName}}

	csrCert, err := x509.CreateCertificateRequest(rand.Reader, &csrTemplate, key)
	if err != nil {
		return "", "", "", err
	}

	//Encode csr
	csr := pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE REQUEST",
		Bytes: csrCert,
	})

	csrObj := Resource{
		ApiVersion: "certificates.k8s.io/v1beta1",
		Kind:       "CertificateSigningRequest",
		MetaData: MetaDatas{
			Name: name,
		},
		Specification: Specs{
			Request: base64.StdEncoding.EncodeToString(csr),
			Usages:  []string{"digital signature", "key encipherment"},
		},
	}

	csrData, err := yaml.Marshal(&csrObj)
	if err != nil {
		return "", "", "", err
	}

	keyData := base64.StdEncoding.EncodeToString(pem.EncodeToMemory(
		&pem.Block{
			Type:  "RSA PRIVATE KEY",
			Bytes: x509.MarshalPKCS1PrivateKey(key),
		},
	))
	if err != nil {
		return "", "", "", err
	}

	return string(csrData), string(keyData), strings.Join([]string{name, "+CertificateSigningRequest"}, ""), nil
}

func createApprovalSubresource(logicalcloud LogicalCloud) (string, error) {
	subresource := subresources.ApprovalSubresource{
		Message:        "Approved for Logical Cloud authentication",
		Reason:         "LogicalCloud",
		Type:           string(certificatesv1beta1.CertificateApproved),
		LastUpdateTime: metav1.Now().Format("2006-01-02T15:04:05Z"),
	}
	csrData, err := json.Marshal(subresource)
	return string(csrData), err
}

/*
queryDBAndSetRsyncInfo queries the MCO db to find the record the sync controller
and then sets the RsyncInfo global variable.
*/
func queryDBAndSetRsyncInfo() (installappclient.RsyncInfo, error) {
	client := controller.NewControllerClient("controller", "controllermetadata")
	vals, _ := client.GetControllers()
	for _, v := range vals {
		if v.Metadata.Name == rsyncName {
			log.Info("Initializing RPC connection to resource synchronizer", log.Fields{
				"Controller": v.Metadata.Name,
			})
			rsyncInfo := installappclient.NewRsyncInfo(v.Metadata.Name, v.Spec.Host, v.Spec.Port)
			return rsyncInfo, nil
		}
	}
	return installappclient.RsyncInfo{}, pkgerrors.Errorf("queryRsyncInfoInMCODB Failed - Could not get find rsync by name : %v", rsyncName)
}

// callRsyncInstall method shall take in the app context id and invoke the rsync service via grpc
func callRsyncInstall(contextid interface{}) error {
	rsyncInfo, err := queryDBAndSetRsyncInfo()
	log.Info("Calling rsync", log.Fields{
		"RsyncName": rsyncInfo.RsyncName,
	})
	if err != nil {
		log.Error("", log.Fields{"err": err})
		return err
	}

	appContextID := fmt.Sprintf("%v", contextid)
	err = installappclient.InvokeInstallApp(appContextID)
	if err != nil {
		log.Error("", log.Fields{"err": err})
		return err
	}
	return nil
}

// callRsyncReadyNotify method shall take in the app context id and invoke the rsync ready-notify grpc api
func callRsyncReadyNotify(contextid interface{}) error {
	rsyncInfo, err := queryDBAndSetRsyncInfo()
	log.Info("Calling rsync", log.Fields{
		"RsyncName": rsyncInfo.RsyncName,
	})
	if err != nil {
		log.Error("", log.Fields{"err": err})
		return err
	}

	appContextID := fmt.Sprintf("%v", contextid)
	return InvokeReadyNotify(appContextID) // see dcm/pkg/module/client.go
}

// callRsyncUninstall method shall take in the app context id and invoke the rsync service via grpc
func callRsyncUninstall(contextid interface{}) error {
	rsyncInfo, err := queryDBAndSetRsyncInfo()
	log.Info("Calling rsync", log.Fields{
		"RsyncName": rsyncInfo.RsyncName,
	})
	if err != nil {
		log.Error("", log.Fields{"err": err})
		return err
	}

	appContextID := fmt.Sprintf("%v", contextid)
	err = installappclient.InvokeUninstallApp(appContextID)
	if err != nil {
		log.Error("", log.Fields{"err": err})
		return err
	}
	return nil
}

// Instantiate prepares all yaml resources to be given to the clusters via rsync,
// then creates an appcontext with such resources and asks rsync to instantiate the logical cloud
func Instantiate(project string, logicalcloud LogicalCloud, clusterList []Cluster,
	quotaList []Quota, userPermissionList []UserPermission) error {

	APP := "logical-cloud"
	logicalCloudName := logicalcloud.MetaData.LogicalCloudName
	level := logicalcloud.Specification.Level

	lcclient := NewLogicalCloudClient()
	lckey := LogicalCloudKey{
		LogicalCloudName: logicalCloudName,
		Project:          project,
	}

	// Check if there was a previous context for this logical cloud
	ac, cid, err := GetLogicalCloudContext(lcclient.storeName, lckey, lcclient.tagContext, project, logicalCloudName)
	if cid != "" {
		// Make sure rsync status for this logical cloud is Terminated,
		// otherwise we can't re-instantiate logical cloud yet
		acStatus, err := GetAppContextStatus(ac)
		if err != nil {
			return err
		}
		switch acStatus.Status {
		case appcontext.AppContextStatusEnum.Terminated:
			// We now know Logical Cloud has terminated, so let's update the entry before we process the instantiate
			err = db.DBconn.RemoveTag(lcclient.storeName, lckey, lcclient.tagContext)
			if err != nil {
				log.Error("Error removing lccontext tag from Logical Cloud", log.Fields{"logicalcloud": logicalCloudName})
				return pkgerrors.Wrap(err, "Error removing lccontext tag from Logical Cloud")
			}
			// And fully delete the old AppContext
			err := ac.DeleteCompositeApp()
			if err != nil {
				log.Error("Error deleting AppContext CompositeApp Logical Cloud", log.Fields{"logicalcloud": logicalCloudName})
				return pkgerrors.Wrap(err, "Error deleting AppContext CompositeApp Logical Cloud")
			}
		case appcontext.AppContextStatusEnum.Terminating:
			log.Error("The Logical Cloud can't be re-instantiated yet, it is being terminated", log.Fields{"logicalcloud": logicalCloudName})
			return pkgerrors.New("The Logical Cloud can't be re-instantiated yet, it is being terminated")
		case appcontext.AppContextStatusEnum.Instantiated:
			log.Error("The Logical Cloud is already instantiated", log.Fields{"logicalcloud": logicalCloudName})
			return pkgerrors.New("The Logical Cloud is already instantiated")
		case appcontext.AppContextStatusEnum.Instantiating:
			log.Error("The Logical Cloud is already instantiating", log.Fields{"logicalcloud": logicalCloudName})
			return pkgerrors.New("The Logical Cloud is already instantiating")
		case appcontext.AppContextStatusEnum.InstantiateFailed:
			log.Error("The Logical Cloud has failed instantiating before, please terminate and try again", log.Fields{"logicalcloud": logicalCloudName})
			return pkgerrors.New("The Logical Cloud has failed instantiating before, please terminate and try again")
		case appcontext.AppContextStatusEnum.TerminateFailed:
			log.Error("The Logical Cloud has failed terminating, please try to terminate again or delete the Logical Cloud", log.Fields{"logicalcloud": logicalCloudName})
			return pkgerrors.New("The Logical Cloud has failed terminating, please try to terminate again or delete the Logical Cloud")
		default:
			log.Error("The Logical Cloud isn't in an expected status so not taking any action", log.Fields{"logicalcloud": logicalCloudName, "status": acStatus.Status})
			return pkgerrors.New("The Logical Cloud isn't in an expected status so not taking any action")
		}
	}

	// if this is an L0 logical cloud, only the following will be done as part of instantiate
	// ================================================================================
	if level == "0" {
		l0ns := ""
		// cycle through all clusters to obtain and validate the single level-0 namespace to use
		// the namespace of each cluster is retrieved from CloudConfig in rsync
		for _, cluster := range clusterList {

			ccc := rsync.NewCloudConfigClient()
			log.Info("Asking rsync's CloudConfig for this cluster's namespace at level-0", log.Fields{"cluster": cluster.Specification.ClusterName})
			ns, err := ccc.GetNamespace(
				cluster.Specification.ClusterProvider,
				cluster.Specification.ClusterName,
			)
			if err != nil {
				if err.Error() == "No CloudConfig was returned" {
					return pkgerrors.New("It looks like the cluster provided as reference does not exist")
				}
				return pkgerrors.Wrap(err, "Couldn't determine namespace for L0 logical cloud")
			}
			// we're checking here if any of the clusters have a differently-named namespace at level 0 and, if so,
			// we abort the instantiate operation because a single namespace name for this logical cloud cannot be inferred
			if len(l0ns) > 0 && ns != l0ns {
				log.Error("The clusters associated to this L0 logical cloud don't all share the same namespace name", log.Fields{"logicalcloud": logicalCloudName})
				return pkgerrors.New("The clusters associated to this L0 logical cloud don't all share the same namespace name")
			}
			l0ns = ns
		}
		// if l0ns is still empty, something definitely went wrong so we can't let this pass
		if len(l0ns) == 0 {
			log.Error("Something went wrong as no cluster namespaces got checked", log.Fields{"logicalcloud": logicalCloudName})
			return pkgerrors.New("Something went wrong as no cluster namespaces got checked")
		}
		// at this point we know what namespace name to give to the logical cloud
		logicalcloud.Specification.NameSpace = l0ns
		// the following is an update operation:
		err = db.DBconn.Insert(lcclient.storeName, lckey, nil, lcclient.tagMeta, logicalcloud)
		if err != nil {
			log.Error("Failed to update L0 logical cloud with a namespace name", log.Fields{"logicalcloud": logicalCloudName, "namespace": l0ns})
			return pkgerrors.Wrap(err, "Failed to update L0 logical cloud with a namespace name")
		}
		log.Info("The L0 logical cloud has been updated with a namespace name", log.Fields{"logicalcloud": logicalCloudName, "namespace": l0ns})

		// prepare empty-shell appcontext for the L0 LC in order to officially set it as Instantiated
		context := appcontext.AppContext{}
		ctxVal, err := context.InitAppContext()
		if err != nil {
			return pkgerrors.Wrap(err, "Error creating L0 LC AppContext")
		}

		handle, err := context.CreateCompositeApp()
		if err != nil {
			return pkgerrors.Wrap(err, "Error creating L0 LC AppContext CompositeApp")
		}

		appHandle, err := context.AddApp(handle, APP)
		if err != nil {
			return cleanupCompositeApp(context, err, "Error adding App to L0 LC AppContext", []string{logicalCloudName, ctxVal.(string)})
		}

		// iterate through cluster list and add all the clusters (as empty-shells)
		for _, cluster := range clusterList {
			clusterName := strings.Join([]string{cluster.Specification.ClusterProvider, "+", cluster.Specification.ClusterName}, "")
			clusterHandle, err := context.AddCluster(appHandle, clusterName)
			// pre-build array to pass to cleanupCompositeApp() [for performance]
			details := []string{logicalCloudName, clusterName, ctxVal.(string)}

			if err != nil {
				return cleanupCompositeApp(context, err, "Error adding Cluster to L0 LC AppContext", details)
			}

			// resource-level order is mandatory too for an empty-shell appcontext
			resOrder, err := json.Marshal(map[string][]string{"resorder": []string{}})
			if err != nil {
				return pkgerrors.Wrap(err, "Error creating resource order JSON")
			}
			_, err = context.AddInstruction(clusterHandle, "resource", "order", string(resOrder))
			if err != nil {
				return cleanupCompositeApp(context, err, "Error adding resource-level order to L0 LC AppContext", details)
			}
			// TODO add resource-level dependency as well
			// app-level order is mandatory too for an empty-shell appcontext
			appOrder, err := json.Marshal(map[string][]string{"apporder": []string{APP}})
			if err != nil {
				return pkgerrors.Wrap(err, "Error creating app order JSON")
			}
			_, err = context.AddInstruction(handle, "app", "order", string(appOrder))
			if err != nil {
				return cleanupCompositeApp(context, err, "Error adding app-level order to L0 LC AppContext", details)
			}
			// TODO add app-level dependency as well
			// TODO move app-level order/dependency out of loop
		}

		// save the context in the logicalcloud db record
		err = db.DBconn.Insert("orchestrator", lckey, nil, "lccontext", ctxVal)
		if err != nil {
			return cleanupCompositeApp(context, err, "Error adding L0 LC AppContext to DB", []string{logicalCloudName, ctxVal.(string)})
		}

		// call resource synchronizer to instantiate the CRs in the cluster
		err = callRsyncInstall(ctxVal)
		if err != nil {
			log.Error("Failed calling rsync install-app", log.Fields{"err": err})
			return pkgerrors.Wrap(err, "Failed calling rsync install-app")
		}

		log.Info("The L0 logical cloud is now associated with an empty-shell appcontext and is ready to be used", log.Fields{"logicalcloud": logicalCloudName, "namespace": l0ns})
		return nil
	}

	if len(userPermissionList) == 0 {
		return pkgerrors.Wrap(err, "Level-1 Logical Clouds require at least a User Permission assigned to its primary namespace")
	}
	primaryUP := false
	for _, up := range userPermissionList {
		if up.Specification.Namespace == logicalcloud.Specification.NameSpace {
			primaryUP = true
			break
		}
	}
	if !primaryUP {
		return pkgerrors.Wrap(err, "Level-1 Logical Clouds require a User Permission assigned to its primary namespace")
	}

	if len(quotaList) == 0 {
		return pkgerrors.Wrap(err, "Level-1 Logical Clouds require a Quota to be associated first")
	}

	// Get resources to be added
	namespace, namespaceName, err := createNamespace(logicalcloud)
	if err != nil {
		return pkgerrors.Wrap(err, "Error Creating Namespace YAML for logical cloud")
	}

	roles, roleNames, err := createRoles(logicalcloud, userPermissionList)
	if err != nil {
		return pkgerrors.Wrap(err, "Error Creating Roles/ClusterRoles YAMLs for logical cloud")
	}

	roleBindings, roleBindingNames, err := createRoleBindings(logicalcloud, userPermissionList)
	if err != nil {
		return pkgerrors.Wrap(err, "Error Creating RoleBindings/ClusterRoleBindings YAMLs for logical cloud")
	}

	quota, quotaName, err := createQuota(quotaList, logicalcloud.Specification.NameSpace)
	if err != nil {
		return pkgerrors.Wrap(err, "Error Creating Quota YAML for logical cloud")
	}

	csr, key, csrName, err := createUserCSR(logicalcloud)
	if err != nil {
		return pkgerrors.Wrap(err, "Error Creating User CSR and Key for logical cloud")
	}

	approval, err := createApprovalSubresource(logicalcloud)

	// From this point on, we are dealing with a new context (not "ac" from above, which is either old or never existed)
	context := appcontext.AppContext{}
	ctxVal, err := context.InitAppContext()
	if err != nil {
		return pkgerrors.Wrap(err, "Error creating AppContext")
	}

	handle, err := context.CreateCompositeApp()
	if err != nil {
		return pkgerrors.Wrap(err, "Error creating AppContext CompositeApp")
	}

	appHandle, err := context.AddApp(handle, APP)
	if err != nil {
		return cleanupCompositeApp(context, err, "Error adding App to AppContext", []string{logicalCloudName, ctxVal.(string)})
	}

	// Iterate through cluster list and add all the clusters
	for _, cluster := range clusterList {
		clusterName := strings.Join([]string{cluster.Specification.ClusterProvider, "+", cluster.Specification.ClusterName}, "")
		clusterHandle, err := context.AddCluster(appHandle, clusterName)
		// pre-build array to pass to cleanupCompositeApp() [for performance]
		details := []string{logicalCloudName, clusterName, ctxVal.(string)}

		if err != nil {
			return cleanupCompositeApp(context, err, "Error adding Cluster to AppContext", details)
		}

		// Add namespace resource to each cluster
		_, err = context.AddResource(clusterHandle, namespaceName, namespace)
		if err != nil {
			return cleanupCompositeApp(context, err, "Error adding Namespace Resource to AppContext", details)
		}

		// Add csr resource to each cluster
		csrHandle, err := context.AddResource(clusterHandle, csrName, csr)
		if err != nil {
			return cleanupCompositeApp(context, err, "Error adding CSR Resource to AppContext", details)
		}

		// Add csr approval as a subresource of csr:
		_, err = context.AddLevelValue(csrHandle, "subresource/approval", approval)
		if err != nil {
			return cleanupCompositeApp(context, err, "Error approving CSR via AppContext", details)
		}

		// Add private key to MongoDB
		err = db.DBconn.Insert("orchestrator", lckey, nil, "privatekey", key)
		if err != nil {
			return cleanupCompositeApp(context, err, "Error adding private key to DB", details)
		}

		// Add [Cluster]Role resources to each cluster
		for i, roleName := range roleNames {
			_, err = context.AddResource(clusterHandle, roleName, roles[i])
			if err != nil {
				return cleanupCompositeApp(context, err, "Error adding [Cluster]Role Resource to AppContext", details)
			}
		}

		// Add [Cluster]RoleBinding resource to each cluster
		for i, roleBindingName := range roleBindingNames {
			_, err = context.AddResource(clusterHandle, roleBindingName, roleBindings[i])
			if err != nil {
				return cleanupCompositeApp(context, err, "Error adding [Cluster]RoleBinding Resource to AppContext", details)
			}
		}

		// Add quota resource to each cluster
		_, err = context.AddResource(clusterHandle, quotaName, quota)
		if err != nil {
			return cleanupCompositeApp(context, err, "Error adding quota Resource to AppContext", details)
		}

		// Add Subresource Order and Subresource Dependency
		subresOrder, err := json.Marshal(map[string][]string{"subresorder": []string{"approval"}})
		if err != nil {
			return pkgerrors.Wrap(err, "Error creating subresource order JSON")
		}
		subresDependency, err := json.Marshal(map[string]map[string]string{"subresdependency": map[string]string{"approval": "go"}})

		// Add Resource Order
		resorderList := []string{namespaceName, quotaName, csrName}
		resorderList = append(resorderList, roleNames...)
		resorderList = append(resorderList, roleBindingNames...)
		resOrder, err := json.Marshal(map[string][]string{"resorder": resorderList})
		if err != nil {
			return pkgerrors.Wrap(err, "Error creating resource order JSON")
		}

		// Add Resource Dependency
		resdep := map[string]string{namespaceName: "go",
			quotaName: strings.Join(
				[]string{"wait on ", namespaceName}, ""),
			csrName: strings.Join(
				[]string{"wait on ", quotaName}, "")}
		// Add [Cluster]Role and [Cluster]RoleBinding resources to dependency graph
		for i, roleName := range roleNames {
			resdep[roleName] = strings.Join([]string{"wait on ", csrName}, "")
			resdep[roleBindingNames[i]] = strings.Join([]string{"wait on ", roleName}, "")
		}
		resDependency, err := json.Marshal(map[string]map[string]string{"resdependency": resdep})

		// Add App Order and App Dependency
		appOrder, err := json.Marshal(map[string][]string{"apporder": []string{APP}})
		if err != nil {
			return pkgerrors.Wrap(err, "Error creating app order JSON")
		}
		appDependency, err := json.Marshal(map[string]map[string]string{"appdependency": map[string]string{APP: "go"}})
		if err != nil {
			return pkgerrors.Wrap(err, "Error creating app dependency JSON")
		}

		// Add Resource-level Order and Dependency
		_, err = context.AddInstruction(clusterHandle, "resource", "order", string(resOrder))
		if err != nil {
			return cleanupCompositeApp(context, err, "Error adding instruction order to AppContext", details)
		}
		_, err = context.AddInstruction(clusterHandle, "resource", "dependency", string(resDependency))
		if err != nil {
			return cleanupCompositeApp(context, err, "Error adding instruction dependency to AppContext", details)
		}
		_, err = context.AddInstruction(csrHandle, "subresource", "order", string(subresOrder))
		if err != nil {
			return cleanupCompositeApp(context, err, "Error adding instruction order to AppContext", details)
		}
		_, err = context.AddInstruction(csrHandle, "subresource", "dependency", string(subresDependency))
		if err != nil {
			return cleanupCompositeApp(context, err, "Error adding instruction dependency to AppContext", details)
		}

		// Add App-level Order and Dependency
		_, err = context.AddInstruction(handle, "app", "order", string(appOrder))
		if err != nil {
			return cleanupCompositeApp(context, err, "Error adding app-level order to AppContext", details)
		}
		_, err = context.AddInstruction(handle, "app", "dependency", string(appDependency))
		if err != nil {
			return cleanupCompositeApp(context, err, "Error adding app-level dependency to AppContext", details)
		}
	}
	// save the context in the logicalcloud db record
	err = db.DBconn.Insert("orchestrator", lckey, nil, "lccontext", ctxVal)
	if err != nil {
		return cleanupCompositeApp(context, err, "Error adding AppContext to DB", []string{logicalCloudName, ctxVal.(string)})
	}

	// call resource synchronizer to instantiate the CRs in the cluster
	err = callRsyncInstall(ctxVal)
	if err != nil {
		return err
	}

	// call grpc streaming api in rsync, which launches a goroutine to wait for the response of
	// every cluster (function should know how many clusters are expected and only finish when
	// all respective certificates have been obtained and all kubeconfigs stored in CloudConfig)
	err = callRsyncReadyNotify(ctxVal)
	if err != nil {
		log.Error("Failed calling rsync ready-notify", log.Fields{"err": err})
		return pkgerrors.Wrap(err, "Failed calling rsync ready-notify")
	}

	return nil

}

// Terminate asks rsync to terminate the logical cloud
func Terminate(project string, logicalcloud LogicalCloud, clusterList []Cluster,
	quotaList []Quota) error {

	logicalCloudName := logicalcloud.MetaData.LogicalCloudName
	level := logicalcloud.Specification.Level
	namespace := logicalcloud.Specification.NameSpace

	lcclient := NewLogicalCloudClient()
	lckey := LogicalCloudKey{
		LogicalCloudName: logicalcloud.MetaData.LogicalCloudName,
		Project:          project,
	}

	ac, cid, err := GetLogicalCloudContext(lcclient.storeName, lckey, lcclient.tagContext, project, logicalCloudName)
	if err != nil {
		return pkgerrors.Wrapf(err, "Logical Cloud is not instantiated")
	}

	// Check if there was a previous context for this logical cloud
	if cid != "" {
		// Make sure rsync status for this logical cloud is Terminated,
		// otherwise we can't re-instantiate logical cloud yet
		acStatus, err := GetAppContextStatus(ac)
		if err != nil {
			return err
		}
		switch acStatus.Status {
		case appcontext.AppContextStatusEnum.Terminated:
			log.Error("The Logical Cloud has already been terminated", log.Fields{"logicalcloud": logicalCloudName})
			return pkgerrors.New("The Logical Cloud has already been terminated")
		case appcontext.AppContextStatusEnum.Terminating:
			log.Error("The Logical Cloud is already being terminated", log.Fields{"logicalcloud": logicalCloudName})
			return pkgerrors.New("The Logical Cloud is already being terminated")
		case appcontext.AppContextStatusEnum.Instantiating:
			log.Error("The Logical Cloud is still instantiating", log.Fields{"logicalcloud": logicalCloudName})
			return pkgerrors.New("The Logical Cloud is still instantiating")
		case appcontext.AppContextStatusEnum.TerminateFailed:
			// try to terminate anyway
			fallthrough
		case appcontext.AppContextStatusEnum.InstantiateFailed:
			// try to terminate anyway
			fallthrough
		case appcontext.AppContextStatusEnum.Instantiated:
			// call resource synchronizer to delete the CRs from every cluster of the logical cloud
			err = callRsyncUninstall(cid)
			if err != nil {
				return err
			}
			// destroy kubeconfigs from CloudConfig if this is an L1 logical cloud
			if level == "1" {

				ccc := rsync.NewCloudConfigClient()
				for _, cluster := range clusterList {
					log.Info("Destroying CloudConfig of logicalcloud/cluster pair via rsync", log.Fields{"cluster": cluster.Specification.ClusterName, "logicalcloud": logicalCloudName, "level": level})
					err = ccc.DeleteCloudConfig(
						cluster.Specification.ClusterProvider,
						cluster.Specification.ClusterName,
						level,
						namespace,
					)

					if err != nil {
						log.Error("Failed destroying at least one CloudConfig of L1 LC", log.Fields{"cluster": cluster, "err": err})
						// continue terminating and removing any remaining CloudConfigs
						// (this happens when terminating a Logical Cloud before all kubeconfigs had a chance to be generated)
					}
				}
			}
		default:
			log.Error("The Logical Cloud isn't in an expected status so not taking any action", log.Fields{"logicalcloud": logicalCloudName, "status": acStatus.Status})
			return pkgerrors.New("The Logical Cloud isn't in an expected status so not taking any action")
		}
	}
	return nil
}

// Stop asks rsync to stop the instantiation or termination of the logical cloud
func Stop(project string, logicalcloud LogicalCloud) error {

	logicalCloudName := logicalcloud.MetaData.LogicalCloudName

	lcclient := NewLogicalCloudClient()
	lckey := LogicalCloudKey{
		LogicalCloudName: logicalcloud.MetaData.LogicalCloudName,
		Project:          project,
	}

	ac, cid, err := GetLogicalCloudContext(lcclient.storeName, lckey, lcclient.tagContext, project, logicalCloudName)
	if err != nil {
		return pkgerrors.Wrapf(err, "Logical Cloud doesn't seem instantiated: %v", logicalCloudName)
	}

	// Check if there was a previous context for this logical cloud
	if cid != "" {
		acStatus, err := GetAppContextStatus(ac)
		if err != nil {
			return err
		}
		if acStatus.Status != appcontext.AppContextStatusEnum.Instantiating &&
			acStatus.Status != appcontext.AppContextStatusEnum.Terminating {
			return pkgerrors.Errorf("Logical Cloud is not instantiating or terminating:" + logicalCloudName)
		}

		// DCM doesn't support StateInfo today, so the Stop operation is effectively a stub
		return pkgerrors.New("Logical Clouds can't be stopped")
	}
	return pkgerrors.New("Logical Cloud is not instantiated: " + logicalCloudName)
}
