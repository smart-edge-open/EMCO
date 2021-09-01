package action_test

import (
	"context"
	"fmt"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/sirupsen/logrus"

	clmcontrollerpb "github.com/open-ness/EMCO/src/clm/pkg/grpc/controller-eventchannel"
	"github.com/open-ness/EMCO/src/hpa-plc/internal/action"
	"github.com/open-ness/EMCO/src/orchestrator/pkg/appcontext"
	"github.com/open-ness/EMCO/src/orchestrator/pkg/infra/contextdb"
	"github.com/open-ness/EMCO/src/orchestrator/pkg/infra/db"

	hpaMod "github.com/open-ness/EMCO/src/hpa-plc/pkg/module"
	orchLog "github.com/open-ness/EMCO/src/orchestrator/pkg/infra/logutils"
	orchMod "github.com/open-ness/EMCO/src/orchestrator/pkg/module"
	"github.com/open-ness/EMCO/src/rsync/pkg/connector"
	pkgerrors "github.com/pkg/errors"
)

type contextForCompositeApp struct {
	context            appcontext.AppContext
	ctxval             interface{}
	compositeAppHandle interface{}
}

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
	compMetadata := appcontext.CompositeAppMeta{Project: p, CompositeApp: ca, Version: v, Release: rName, DeploymentIntentGroup: dig, Namespace: namespace, Level: level}
	err = context.AddCompositeAppMeta(compMetadata)
	if err != nil {
		return contextForCompositeApp{}, pkgerrors.Wrap(err, "Error Adding CompositeAppMeta")
	}

	cca := contextForCompositeApp{context: context, ctxval: ctxval, compositeAppHandle: compositeHandle}

	return cca, nil
}

func TestAction(t *testing.T) {

	fmt.Printf("\n================== hpa-placementcontroller TestAction .. start ==================\n")

	orchLog.SetLoglevel(logrus.InfoLevel)
	RegisterFailHandler(Fail)
	RunSpecs(t, "HPA-PLACEMENT-CONTROLLER")

	fmt.Printf("\n================== hpa-placementcontroller TestAction .. end ==================\n")
}

var _ = Describe("HPA-PLACEMENT-CONTROLLER", func() {

	var (
		mdb *db.MockDB
		edb *contextdb.MockConDb

		project    string = "p"
		compApp    string = "ca"
		version    string = "v1"
		dig        string = "dig"
		app1       string = "client"
		app2       string = "server"
		logicCloud string = "default"
		release    string = "r1"
		namespace  string = "n1"

		hpaIntentName1        string = "hpa-intent-1"
		hpaConsumerName1      string = "hpa-consumer-1"
		replicaCount          int64  = 1
		replicaCountMax       int64  = 20
		deploymentName1       string = "r1-http-client"
		containerName1        string = "http-client-1"
		containerName2        string = "http-client-2"
		hpaIntentName2        string = "hpa-intent-2"
		hpaIntentName3        string = "hpa-intent-3"
		hpaConsumerName2      string = "hpa-consumer-2"
		hpaAllocResourceName1 string = "hpa-alloc-resource-1"
	)

	var cfca contextForCompositeApp

	BeforeEach(func() {
		//logrus.SetOutput(ioutil.Discard)
		fmt.Printf("\n\n================== GINKGO ACTION TESTCASE START .. [%v] ==================\n\n", CurrentGinkgoTestDescription().TestText)

		// mongo mockdb
		mdb = &db.MockDB{
			Items: []map[string]map[string][]byte{
				{
					orchMod.ProjectKey{ProjectName: project}.String(): {
						"projectmetadata": []byte(
							"{\"project-name\":\"" + project + "\"," +
								"\"description\":\"Test project for unit testing\"}"),
					},
					orchMod.CompositeAppKey{CompositeAppName: compApp,
						Version: version, Project: project}.String(): {
						"compositeappmetadata": []byte(
							"{\"metadata\":{" +
								"\"name\":\"" + compApp + "\"," +
								"\"description\":\"description\"," +
								"\"userData1\":\"user data\"," +
								"\"userData2\":\"user data\"" +
								"}," +
								"\"spec\":{" +
								"\"version\":\"version of the composite app\"}}"),
					},
					orchMod.AppKey{App: "", Project: project, CompositeApp: compApp, CompositeAppVersion: version}.String(): {
						"appmetadata": []byte(
							"{" +
								"\"metadata\": {" +
								"\"name\":\"" + app1 + "\"," +
								"\"description\": \"Test App for unit testing\"," +
								"\"userData1\": \"userData1\"," +
								"\"userData2\": \"userData2\"}" +
								"}"),
						"appcontent": []byte(
							"{" +
								"\"FileContent\": \"sample file content\"" +
								"}"),
					},
					orchMod.DeploymentIntentGroupKey{
						Name:         dig,
						Project:      project,
						CompositeApp: compApp,
						Version:      version,
					}.String(): {
						"deploymentintentgroupmetadata": []byte(
							"{" +
								"\"metadata\":{" +
								"\"name\":\"" + dig + "\"," +
								"\"description\":\"DescriptionTestDeploymentIntentGroup\"," +
								"\"userData1\": \"userData1\"," +
								"\"userData2\": \"userData2\"}" +
								"}"),
					},
					hpaMod.HpaIntentKey{IntentName: "",
						Project: project, CompositeApp: compApp,
						Version: version, DeploymentIntentGroup: dig}.String(): {
						"HpaPlacementControllerMetadata": []byte(
							"{" +
								"\"metadata\" : {" +
								"\"Name\":\"" + hpaIntentName1 + "\"," +
								"\"Description\":\"Test Intent for unit testing\"," +
								"\"UserData1\": \"userData1\"," +
								"\"UserData2\":\"userData2\"}," +
								"\"spec\" : {" +
								"\"app-name\":\"" + app1 + "\"}" +
								"}"),
					},
					hpaMod.HpaIntentKey{IntentName: hpaIntentName1,
						Project: project, CompositeApp: compApp,
						Version: version, DeploymentIntentGroup: dig}.String(): {
						"HpaPlacementControllerMetadata": []byte(
							"{" +
								"\"metadata\" : {" +
								"\"Name\":\"" + hpaIntentName1 + "\"," +
								"\"Description\":\"Test Intent for unit testing\"," +
								"\"UserData1\": \"userData1\"," +
								"\"UserData2\":\"userData2\"}," +
								"\"spec\" : {" +
								"\"app-name\":\"" + app1 + "\"}" +
								"}"),
					},
					hpaMod.HpaConsumerKey{ConsumerName: "", IntentName: hpaIntentName1,
						Project: project, CompositeApp: compApp,
						Version: version, DeploymentIntentGroup: dig}.String(): {
						"HpaPlacementControllerMetadata": []byte(
							"{" +
								"\"metadata\" : {" +
								"\"Name\":\"" + hpaConsumerName1 + "\"," +
								"\"Description\":\"Test Consumer for unit testing\"," +
								"\"UserData1\": \"userData1\"," +
								"\"UserData2\":\"userData2\"}," +
								"\"spec\" : {" +
								"\"replicas\":" + fmt.Sprint(replicaCount) + "," +
								"\"name\":\"" + deploymentName1 + "\"," +
								"\"container-name\":\"" + containerName1 + "\"}" +
								"}"),
					},
					hpaMod.HpaConsumerKey{ConsumerName: hpaConsumerName1, IntentName: hpaIntentName1,
						Project: project, CompositeApp: compApp,
						Version: version, DeploymentIntentGroup: dig}.String(): {
						"HpaPlacementControllerMetadata": []byte(
							"{" +
								"\"metadata\" : {" +
								"\"Name\":\"" + hpaConsumerName1 + "\"," +
								"\"Description\":\"Test Consumer for unit testing\"," +
								"\"UserData1\": \"userData1\"," +
								"\"UserData2\":\"userData2\"}," +
								"\"spec\" : {" +
								"\"replicas\":" + fmt.Sprint(replicaCount) + "," +
								"\"name\":\"" + deploymentName1 + "\"," +
								"\"container-name\":\"" + containerName1 + "\"}" +
								"}"),
					},
					hpaMod.HpaResourceKey{ResourceName: "", ConsumerName: hpaConsumerName1, IntentName: hpaIntentName1,
						Project: project, CompositeApp: compApp,
						Version: version, DeploymentIntentGroup: dig}.String(): {
						"HpaPlacementControllerMetadata": []byte(
							"{" +
								"\"metadata\" : {" +
								"\"Name\":\"" + hpaAllocResourceName1 + "\"," +
								"\"Description\":\"Test Resource for unit testing\"," +
								"\"UserData1\":\"userData1\"," +
								"\"UserData2\":\"userData2\"}," +
								"\"spec\" : {" +
								"\"allocatable\":true," +
								"\"resource\":{\"name\":\"cpu\", \"requests\":15, \"limits\":15}" +
								"}" +
								"}"),
					},
				},
				{
					orchMod.ProjectKey{ProjectName: project}.String(): {
						"projectmetadata": []byte(
							"{\"project-name\":\"" + project + "\"," +
								"\"description\":\"Test project for unit testing\"}"),
					},
					orchMod.CompositeAppKey{CompositeAppName: compApp,
						Version: version, Project: project}.String(): {
						"compositeappmetadata": []byte(
							"{\"metadata\":{" +
								"\"name\":\"" + compApp + "\"," +
								"\"description\":\"description\"," +
								"\"userData1\":\"user data\"," +
								"\"userData2\":\"user data\"" +
								"}," +
								"\"spec\":{" +
								"\"version\":\"version of the composite app\"}}"),
					},
					orchMod.AppKey{App: "", Project: project, CompositeApp: compApp, CompositeAppVersion: version}.String(): {
						"appmetadata": []byte(
							"{" +
								"\"metadata\": {" +
								"\"name\":\"" + app2 + "\"," +
								"\"description\": \"Test App for unit testing\"," +
								"\"userData1\": \"userData1\"," +
								"\"userData2\": \"userData2\"}" +
								"}"),
						"appcontent": []byte(
							"{" +
								"\"FileContent\": \"sample file content\"" +
								"}"),
					},
					orchMod.DeploymentIntentGroupKey{
						Name:         dig,
						Project:      project,
						CompositeApp: compApp,
						Version:      version,
					}.String(): {
						"deploymentintentgroupmetadata": []byte(
							"{" +
								"\"metadata\":{" +
								"\"name\":\"" + dig + "\"," +
								"\"description\":\"DescriptionTestDeploymentIntentGroup\"," +
								"\"userData1\": \"userData1\"," +
								"\"userData2\": \"userData2\"}" +
								"}"),
					},
					hpaMod.HpaIntentKey{IntentName: "",
						Project: project, CompositeApp: compApp,
						Version: version, DeploymentIntentGroup: dig}.String(): {
						"HpaPlacementControllerMetadata": []byte(
							"{" +
								"\"metadata\" : {" +
								"\"Name\":\"" + hpaIntentName2 + "\"," +
								"\"Description\":\"Test Intent for unit testing\"," +
								"\"UserData1\": \"userData1\"," +
								"\"UserData2\":\"userData2\"}," +
								"\"spec\" : {" +
								"\"app-name\":\"" + app2 + "\"}" +
								"}"),
					},
					hpaMod.HpaIntentKey{IntentName: hpaIntentName2,
						Project: project, CompositeApp: compApp,
						Version: version, DeploymentIntentGroup: dig}.String(): {
						"HpaPlacementControllerMetadata": []byte(
							"{" +
								"\"metadata\" : {" +
								"\"Name\":\"" + hpaIntentName2 + "\"," +
								"\"Description\":\"Test Intent for unit testing\"," +
								"\"UserData1\": \"userData1\"," +
								"\"UserData2\":\"userData2\"}," +
								"\"spec\" : {" +
								"\"app-name\":\"" + app2 + "\"}" +
								"}"),
					},
					hpaMod.HpaConsumerKey{ConsumerName: "", IntentName: hpaIntentName2,
						Project: project, CompositeApp: compApp,
						Version: version, DeploymentIntentGroup: dig}.String(): {
						"HpaPlacementControllerMetadata": []byte(
							"{" +
								"\"metadata\" : {" +
								"\"Name\":\"" + hpaConsumerName2 + "\"," +
								"\"Description\":\"Test Consumer for unit testing\"," +
								"\"UserData1\": \"userData1\"," +
								"\"UserData2\":\"userData2\"}," +
								"\"spec\" : {" +
								"\"replicas\":" + fmt.Sprint(replicaCount) + "," +
								"\"name\":\"" + deploymentName1 + "\"," +
								"\"container-name\":\"" + containerName2 + "\"}" +
								"}"),
					},
					hpaMod.HpaConsumerKey{ConsumerName: hpaConsumerName2, IntentName: hpaIntentName2,
						Project: project, CompositeApp: compApp,
						Version: version, DeploymentIntentGroup: dig}.String(): {
						"HpaPlacementControllerMetadata": []byte(
							"{" +
								"\"metadata\" : {" +
								"\"Name\":\"" + hpaConsumerName2 + "\"," +
								"\"Description\":\"Test Consumer for unit testing\"," +
								"\"UserData1\": \"userData1\"," +
								"\"UserData2\":\"userData2\"}," +
								"\"spec\" : {" +
								"\"replicas\":" + fmt.Sprint(replicaCount) + "," +
								"\"name\":\"" + deploymentName1 + "\"," +
								"\"container-name\":\"" + containerName2 + "\"}" +
								"}"),
					},
					hpaMod.HpaResourceKey{ResourceName: "", ConsumerName: hpaConsumerName2, IntentName: hpaIntentName2,
						Project: project, CompositeApp: compApp,
						Version: version, DeploymentIntentGroup: dig}.String(): {
						"HpaPlacementControllerMetadata": []byte(
							"{" +
								"\"metadata\" : {" +
								"\"Name\":\"" + hpaAllocResourceName1 + "\"," +
								"\"Description\":\"Test Resource for unit testing\"," +
								"\"UserData1\":\"userData1\"," +
								"\"UserData2\":\"userData2\"}," +
								"\"spec\" : {" +
								"\"allocatable\":true," +
								"\"resource\":{\"name\":\"cpu\", \"requests\":15, \"limits\":15}" +
								"}" +
								"}"),
					},
				},
			},
		}
		mdb.Err = nil
		db.DBconn = mdb

		// etcd mockdb
		edb = new(contextdb.MockConDb)
		edb.Err = nil
		contextdb.Db = edb

		// Initialize etcd with default values
		var err error
		cfca, err = makeAppContextForCompositeApp(project, compApp, version, release, dig, namespace, logicCloud)
		Expect(err).To(BeNil())

		cap, err := cfca.context.AddApp(cfca.compositeAppHandle, app1)
		Expect(err).To(BeNil())

		ch, err := cfca.context.AddCluster(cap, "provider1-cluster1")
		Expect(err).To(BeNil())

		sap, err := cfca.context.AddApp(cfca.compositeAppHandle, app2)
		Expect(err).To(BeNil())

		sh, err := cfca.context.AddCluster(sap, "provider1-cluster2")
		Expect(err).To(BeNil())
		err = cfca.context.AddClusterMetaGrp(ch, "1")
		Expect(err).To(BeNil())
		err = cfca.context.AddClusterMetaGrp(sh, "1")
		Expect(err).To(BeNil())

		
		// Use Kube Fake client for unit-testing
		connector.IsTestKubeClient = true
	})

	AfterEach(func() {
		mdb = nil
		db.DBconn = mdb

		// etcd mockdb
		edb = nil
		contextdb.Db = edb
	})

	Describe("Filter Clusters", func() {

		It("*** GINKGO ACTION TESTCASE: successful allocatable-resource filter-clusters", func() {
			contextID := fmt.Sprintf("%v", cfca.ctxval)
			err := action.FilterClusters(contextID)
			Expect(err).To(BeNil())
		})

		It("*** GINKGO ACTION TESTCASE: successful non-allocatable-resource filter-cluster", func() {
			(mdb.Items[0])[hpaMod.HpaResourceKey{ResourceName: "", ConsumerName: hpaConsumerName1, IntentName: hpaIntentName1,
				Project: project, CompositeApp: compApp,
				Version: version, DeploymentIntentGroup: dig}.String()] = map[string][]byte{
				"HpaPlacementControllerMetadata": []byte(
					"{" +
						"\"metadata\" : {" +
						"\"Name\":\"" + hpaAllocResourceName1 + "\"," +
						"\"Description\":\"Test Resource for unit testing\"," +
						"\"UserData1\":\"userData1\"," +
						"\"UserData2\":\"userData2\"}," +
						"\"spec\" : {" +
						"\"allocatable\":false," +
						"\"resource\":{\"key\":\"feature.node.kubernetes.io/intel_qat\", \"value\":\"true\"}" +
						"}" +
						"}"),
			}

			contextID := fmt.Sprintf("%v", cfca.ctxval)
			err := action.FilterClusters(contextID)
			Expect(err).To(BeNil())
		})

		It("*** GINKGO ACTION TESTCASE: unsuccessful action to due to invalid app-context", func() {
			err := action.FilterClusters("1234")
			Expect(err).To(HaveOccurred())
		})

		It("*** GINKGO ACTION TESTCASE: unsuccessful allocatable-resource filter-clusters due to invalid request count", func() {
			(mdb.Items[0])[hpaMod.HpaResourceKey{ResourceName: "", ConsumerName: hpaConsumerName1, IntentName: hpaIntentName1,
				Project: project, CompositeApp: compApp,
				Version: version, DeploymentIntentGroup: dig}.String()] = map[string][]byte{
				"HpaPlacementControllerMetadata": []byte(
					"{" +
						"\"metadata\" : {" +
						"\"Name\":\"" + hpaAllocResourceName1 + "\"," +
						"\"Description\":\"Test Resource for unit testing\"," +
						"\"UserData1\":\"userData1\"," +
						"\"UserData2\":\"userData2\"}," +
						"\"spec\" : {" +
						"\"allocatable\":true," +
						"\"resource\":{\"name\":\"cpu\", \"requests\":20, \"limits\":20}" +
						"}" +
						"}"),
			}

			contextID := fmt.Sprintf("%v", cfca.ctxval)
			err := action.FilterClusters(contextID)
			Expect(err).To(HaveOccurred())
		})

		It("*** GINKGO ACTION TESTCASE: unsuccessful allocatable-resource filter-clusters due to non-exsiting k8s resource name", func() {
			(mdb.Items[0])[hpaMod.HpaResourceKey{ResourceName: "", ConsumerName: hpaConsumerName1, IntentName: hpaIntentName1,
				Project: project, CompositeApp: compApp,
				Version: version, DeploymentIntentGroup: dig}.String()] = map[string][]byte{
				"HpaPlacementControllerMetadata": []byte(
					"{" +
						"\"metadata\" : {" +
						"\"Name\":\"" + hpaAllocResourceName1 + "\"," +
						"\"Description\":\"Test Resource for unit testing\"," +
						"\"UserData1\":\"userData1\"," +
						"\"UserData2\":\"userData2\"}," +
						"\"spec\" : {" +
						"\"allocatable\":true," +
						"\"resource\":{\"name\":\"cpu1\", \"requests\":1, \"limits\":1}" +
						"}" +
						"}"),
			}

			contextID := fmt.Sprintf("%v", cfca.ctxval)
			err := action.FilterClusters(contextID)
			Expect(err).To(HaveOccurred())
		})

		It("*** GINKGO ACTION TESTCASE: unsuccessful non-allocatable-resource filter-cluster due to non-existing label", func() {
			(mdb.Items[0])[hpaMod.HpaResourceKey{ResourceName: "", ConsumerName: hpaConsumerName1, IntentName: hpaIntentName1,
				Project: project, CompositeApp: compApp,
				Version: version, DeploymentIntentGroup: dig}.String()] = map[string][]byte{
				"HpaPlacementControllerMetadata": []byte(
					"{" +
						"\"metadata\" : {" +
						"\"Name\":\"" + hpaAllocResourceName1 + "\"," +
						"\"Description\":\"Test Resource for unit testing\"," +
						"\"UserData1\":\"userData1\"," +
						"\"UserData2\":\"userData2\"}," +
						"\"spec\" : {" +
						"\"allocatable\":false," +
						"\"resource\":{\"key\":\"feature.node.kubernetes1.io/intel_qa\", \"value\":\"true\"}" +
						"}" +
						"}"),
			}

			contextID := fmt.Sprintf("%v", cfca.ctxval)
			err := action.FilterClusters(contextID)
			Expect(err).To(HaveOccurred())
		})

		It("*** GINKGO ACTION TESTCASE: unsuccessful filter-cluster when there are no apps in composite-app apps", func() {
			(mdb.Items[0])[orchMod.AppKey{App: "", Project: project, CompositeApp: compApp, CompositeAppVersion: version}.String()] = nil
			contextID := fmt.Sprintf("%v", cfca.ctxval)
			err := action.FilterClusters(contextID)
			Expect(err).To(HaveOccurred())
		})

		It("*** GINKGO ACTION TESTCASE: successful filter-cluster when hpa-intent app is not one of composite-app apps", func() {
			(mdb.Items[0])[orchMod.AppKey{App: "", Project: project, CompositeApp: compApp, CompositeAppVersion: version}.String()] = map[string][]byte{
				"appmetadata": []byte(
					"{" +
						"\"metadata\": {" +
						"\"name\":\"" + app2 + "\"," +
						"\"description\": \"Test App for unit testing\"," +
						"\"userData1\": \"userData1\"," +
						"\"userData2\": \"userData2\"}" +
						"}"),
				"appcontent": []byte(
					"{" +
						"\"FileContent\": \"sample file content\"" +
						"}"),
			}
			contextID := fmt.Sprintf("%v", cfca.ctxval)
			err := action.FilterClusters(contextID)
			Expect(err).To(BeNil())
		})

		It("failed filter-cluster with NO hpa-intents", func() {
			(mdb.Items[0])[hpaMod.HpaIntentKey{IntentName: "",
				Project: project, CompositeApp: compApp,
				Version: version, DeploymentIntentGroup: dig}.String()] = nil
			contextID := fmt.Sprintf("%v", cfca.ctxval)
			err := action.FilterClusters(contextID)
			Expect(err).To(BeNil())

		})

		It("*** GINKGO ACTION TESTCASE: unsuccessful hpa-intent filter-cluster for non-existing hpa-consumers", func() {
			(mdb.Items[0])[hpaMod.HpaIntentKey{IntentName: "",
				Project: project, CompositeApp: compApp,
				Version: version, DeploymentIntentGroup: dig}.String()] = map[string][]byte{
				"HpaPlacementControllerMetadata": []byte(
					"{" +
						"\"metadata\" : {" +
						"\"Name\":\"" + hpaIntentName3 + "\"," +
						"\"Description\":\"Test Intent for unit testing\"," +
						"\"UserData1\": \"userData1\"," +
						"\"UserData2\":\"userData2\"}," +
						"\"spec\" : {" +
						"\"app-name\":\"" + app1 + "\"}" +
						"}"),
			}
			contextID := fmt.Sprintf("%v", cfca.ctxval)
			err := action.FilterClusters(contextID)
			Expect(err).To(HaveOccurred())

		})

		It("*** GINKGO ACTION TESTCASE: unsuccessful filter-cluster for non-existing hpa-consumer", func() {
			(mdb.Items[0])[hpaMod.HpaConsumerKey{ConsumerName: "", IntentName: hpaIntentName1,
				Project: project, CompositeApp: compApp,
				Version: version, DeploymentIntentGroup: dig}.String()] = map[string][]byte{
				"HpaPlacementControllerMetadata": []byte(
					"{" +
						"\"metadata\" : {" +
						"\"Name\":\"" + hpaConsumerName2 + "\"," +
						"\"Description\":\"Test Consumer for unit testing\"," +
						"\"UserData1\": \"userData1\"," +
						"\"UserData2\":\"userData2\"}," +
						"\"spec\" : {" +
						"\"name\":\"" + deploymentName1 + "\"," +
						"\"container-name\":\"" + containerName1 + "\"}" +
						"}"),
			}
			contextID := fmt.Sprintf("%v", cfca.ctxval)
			err := action.FilterClusters(contextID)
			Expect(err).To(HaveOccurred())
		})

		It("*** GINKGO ACTION TESTCASE: successful filter-cluster with NO hpa-resources", func() {
			(mdb.Items[0])[hpaMod.HpaResourceKey{ResourceName: "", ConsumerName: hpaConsumerName1, IntentName: hpaIntentName1,
				Project: project, CompositeApp: compApp,
				Version: version, DeploymentIntentGroup: dig}.String()] = nil

			contextID := fmt.Sprintf("%v", cfca.ctxval)
			err := action.FilterClusters(contextID)
			Expect(err).To(BeNil())
		})

		It("*** GINKGO ACTION TESTCASE: successful non-allocatable resource filter-cluster with maximum replicaCount", func() {
			(mdb.Items[0])[hpaMod.HpaConsumerKey{ConsumerName: "", IntentName: hpaIntentName1,
				Project: project, CompositeApp: compApp,
				Version: version, DeploymentIntentGroup: dig}.String()] = map[string][]byte{
				"HpaPlacementControllerMetadata": []byte(
					"{" +
						"\"metadata\" : {" +
						"\"Name\":\"" + hpaConsumerName1 + "\"," +
						"\"Description\":\"Test Consumer for unit testing\"," +
						"\"UserData1\": \"userData1\"," +
						"\"UserData2\":\"userData2\"}," +
						"\"spec\" : {" +
						"\"replicas\":" + fmt.Sprint(replicaCountMax) + "," +
						"\"name\":\"" + deploymentName1 + "\"," +
						"\"container-name\":\"" + containerName1 + "\"}" +
						"}"),
			}

			(mdb.Items[0])[hpaMod.HpaResourceKey{ResourceName: "", ConsumerName: hpaConsumerName1, IntentName: hpaIntentName1,
				Project: project, CompositeApp: compApp,
				Version: version, DeploymentIntentGroup: dig}.String()] = map[string][]byte{
				"HpaPlacementControllerMetadata": []byte(
					"{" +
						"\"metadata\" : {" +
						"\"Name\":\"" + hpaAllocResourceName1 + "\"," +
						"\"Description\":\"Test Resource for unit testing\"," +
						"\"UserData1\":\"userData1\"," +
						"\"UserData2\":\"userData2\"}," +
						"\"spec\" : {" +
						"\"allocatable\":false," +
						"\"resource\":{\"key\":\"feature.node.kubernetes.io/intel_qat\", \"value\":\"true\"}" +
						"}" +
						"}"),
			}

			contextID := fmt.Sprintf("%v", cfca.ctxval)
			err := action.FilterClusters(contextID)
			Expect(err).To(BeNil())
		})

		It("*** GINKGO ACTION TESTCASE: unsuccessful filter-clusters with original kube client", func() {
			// Use Kube Fake client for unit-testing
			connector.IsTestKubeClient = false

			contextID := fmt.Sprintf("%v", cfca.ctxval)
			err := action.FilterClusters(contextID)
			Expect(err).To(HaveOccurred())
		})

		It("*** GINKGO ACTION TESTCASE: unsuccessful filter-cluster with maximum replicaCount", func() {
			(mdb.Items[0])[hpaMod.HpaConsumerKey{ConsumerName: "", IntentName: hpaIntentName1,
				Project: project, CompositeApp: compApp,
				Version: version, DeploymentIntentGroup: dig}.String()] = map[string][]byte{
				"HpaPlacementControllerMetadata": []byte(
					"{" +
						"\"metadata\" : {" +
						"\"Name\":\"" + hpaConsumerName1 + "\"," +
						"\"Description\":\"Test Consumer for unit testing\"," +
						"\"UserData1\": \"userData1\"," +
						"\"UserData2\":\"userData2\"}," +
						"\"spec\" : {" +
						"\"replicas\":" + fmt.Sprint(replicaCountMax) + "," +
						"\"name\":\"" + deploymentName1 + "\"," +
						"\"container-name\":\"" + containerName1 + "\"}" +
						"}"),
			}

			(mdb.Items[0])[hpaMod.HpaResourceKey{ResourceName: "", ConsumerName: hpaConsumerName1, IntentName: hpaIntentName1,
				Project: project, CompositeApp: compApp,
				Version: version, DeploymentIntentGroup: dig}.String()] = map[string][]byte{
				"HpaPlacementControllerMetadata": []byte(
					"{" +
						"\"metadata\" : {" +
						"\"Name\":\"" + hpaAllocResourceName1 + "\"," +
						"\"Description\":\"Test Resource for unit testing\"," +
						"\"UserData1\":\"userData1\"," +
						"\"UserData2\":\"userData2\"}," +
						"\"spec\" : {" +
						"\"allocatable\":true," +
						"\"resource\":{\"name\":\"cpu\", \"requests\":5, \"limits\":5}" +
						"}" +
						"}"),
			}
			contextID := fmt.Sprintf("%v", cfca.ctxval)
			err := action.FilterClusters(contextID)
			Expect(err).To(HaveOccurred())
		})

		It("*** GINKGO ACTION TESTCASE: successful filter-cluster with NO hpa-consumers", func() {
			(mdb.Items[0])[hpaMod.HpaConsumerKey{ConsumerName: "", IntentName: hpaIntentName1,
				Project: project, CompositeApp: compApp,
				Version: version, DeploymentIntentGroup: dig}.String()] = nil
			contextID := fmt.Sprintf("%v", cfca.ctxval)
			err := action.FilterClusters(contextID)
			Expect(err).To(BeNil())
		})

		It("*** GINKGO ACTION TESTCASE: successful publish received cluster-create event", func() {
			var req clmcontrollerpb.ClmControllerEventRequest
			req.ProviderName = "provider1"
			req.ClusterName = "cluster1"
			req.Event = clmcontrollerpb.ClmControllerEventType_CLUSTER_CREATED

			err := action.Publish(context.TODO(), &req)
			Expect(err).To(BeNil())
		})

		It("*** GINKGO ACTION TESTCASE: successful publish received cluster-modify event", func() {
			var req clmcontrollerpb.ClmControllerEventRequest
			req.ProviderName = "provider1"
			req.ClusterName = "cluster1"
			req.Event = clmcontrollerpb.ClmControllerEventType_CLUSTER_UPDATED

			err := action.Publish(context.TODO(), &req)
			Expect(err).To(BeNil())
		})

		It("*** GINKGO ACTION TESTCASE: unsuccessful publish received unknown event", func() {
			var req clmcontrollerpb.ClmControllerEventRequest
			req.ProviderName = "provider1"
			req.ClusterName = "cluster1"
			req.Event = 5

			err := action.Publish(context.TODO(), &req)
			Expect(err).To(BeNil())
		})

		It("*** GINKGO ACTION TESTCASE: successful publish received cluster-delete event", func() {
			var req clmcontrollerpb.ClmControllerEventRequest

			req.ProviderName = "provider1"
			req.ClusterName = "cluster1"
			req.Event = clmcontrollerpb.ClmControllerEventType_CLUSTER_CREATED
			err := action.Publish(context.TODO(), &req)
			Expect(err).To(BeNil())

			req.ProviderName = "provider1"
			req.ClusterName = "cluster1"
			req.Event = clmcontrollerpb.ClmControllerEventType_CLUSTER_DELETED
			err = action.Publish(context.TODO(), &req)
			Expect(err).To(BeNil())
		})

		It("*** GINKGO ACTION TESTCASE: successful get labels", func() {
			var req clmcontrollerpb.ClmControllerEventRequest
			req.ProviderName = "provider1"
			req.ClusterName = "cluster1"
			req.Event = clmcontrollerpb.ClmControllerEventType_CLUSTER_CREATED

			err := action.Publish(context.TODO(), &req)
			Expect(err).To(BeNil())

			_, err = action.GetKubeClusterLabels("provider1", "cluster1")
			Expect(err).To(BeNil())
		})

		It("*** GINKGO ACTION TESTCASE: unsuccessful publish cluster-create", func() {
			mdb.Err = pkgerrors.New("Error")
			var req clmcontrollerpb.ClmControllerEventRequest
			req.ProviderName = "provider1"
			req.ClusterName = "cluster1"
			req.Event = clmcontrollerpb.ClmControllerEventType_CLUSTER_CREATED

			err := action.Publish(context.TODO(), &req)
			Expect(err).To(HaveOccurred())
		})

		It("*** GINKGO ACTION TESTCASE: unsuccessful publish cluster-delete", func() {
			var req clmcontrollerpb.ClmControllerEventRequest

			req.ProviderName = "provider1"
			req.ClusterName = "cluster1"
			req.Event = clmcontrollerpb.ClmControllerEventType_CLUSTER_CREATED
			err := action.Publish(context.TODO(), &req)
			Expect(err).To(BeNil())

			req.ProviderName = "provider1"
			req.ClusterName = "cluster1"
			mdb.Err = pkgerrors.New("Error")
			req.Event = clmcontrollerpb.ClmControllerEventType_CLUSTER_DELETED
			err = action.Publish(context.TODO(), &req)
			Expect(err).To(HaveOccurred())
		})

		It("*** GINKGO ACTION TESTCASE: unsuccessful get labels cluster-create", func() {
			var req clmcontrollerpb.ClmControllerEventRequest
			req.ProviderName = "provider1"
			req.ClusterName = "cluster1"
			req.Event = clmcontrollerpb.ClmControllerEventType_CLUSTER_CREATED

			err := action.Publish(context.TODO(), &req)
			Expect(err).To(BeNil())

			mdb.Err = pkgerrors.New("Error")
			_, err = action.GetKubeClusterLabels("provider1", "cluster1")
			Expect(err).To(HaveOccurred())
		})

	})
})
