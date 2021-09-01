package action_test

import (
	"fmt"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/sirupsen/logrus"

	"github.com/open-ness/EMCO/src/hpa-ac/internal/action"
	"github.com/open-ness/EMCO/src/orchestrator/pkg/appcontext"
	"github.com/open-ness/EMCO/src/orchestrator/pkg/infra/contextdb"
	"github.com/open-ness/EMCO/src/orchestrator/pkg/infra/db"

	hpaMod "github.com/open-ness/EMCO/src/hpa-plc/pkg/module"
	orchLog "github.com/open-ness/EMCO/src/orchestrator/pkg/infra/logutils"
	orchMod "github.com/open-ness/EMCO/src/orchestrator/pkg/module"
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

	fmt.Printf("\n================== hpa-actioncontroller TestAction .. start ==================\n")

	orchLog.SetLoglevel(logrus.InfoLevel)
	RegisterFailHandler(Fail)
	RunSpecs(t, "HPA-ACTION-CONTROLLER")

	fmt.Printf("\n================== hpa-actioncontroller TestAction .. end ==================\n")
}

var _ = Describe("HPA-ACTION-CONTROLLER", func() {

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
		deploymentName1       string = "r1-http-client"
		containerName1        string = "http-client-1"
		hpaIntentName2        string = "hpa-intent-2"
		hpaConsumerName2      string = "hpa-consumer-2"
		hpaAllocResourceName1 string = "hpa-alloc-resource-1"

		numberOfReplicas0 int64 = 0
		numberOfReplicas2 int64 = 2

		deploymentSpec string = `apiVersion: networking.k8s.io/v1
kind: Deployment
metadata:
  name: r1-http-client
spec:
  replicas: 1
  template:
    metadata:
      labels:
        app: http-client
    spec:
      containers:
      - name: http-client-1
`

		badDeploymentSpec string = `apiVersion: networking.k8s.io/v1
kind: Deployment
metadata
  name: r1-http-client
spec:
  replicas: 1
  template:
    metadata:
      labels:
        app: http-client
    spec:
      containers:
      - name: http-client-1
`

		badDeploymentSpecNoMeta string = `apiVersion: networking.k8s.io/v1
kind: Deployment
spec:
  replicas: 1
  template:
    metadata:
      labels:
        app: http-client
    spec:
      containers:
      - name: http-client-1
`

		badDeploymentSpec1 string = `apiVersion: networking.k8s.io/v1
kind: Deployment
metadata:
  name: r1-http-client
`
		badDeploymentSpec2 string = `apiVersion: networking.k8s.io/v1
kind: Deployment
metadata:
  name: r1-http-client
spec:
  replicas: 1
`

		badDeploymentSpec3 string = `apiVersion: networking.k8s.io/v1
kind: Deployment
metadata:
  name: r1-http-client-2
spec:
  replicas: 1
  template:
    metadata:
      labels:
        app: http-client
    spec:
      containers:
      - name: http-client-1
`

		badDeploymentSpec4 string = `apiVersion: networking.k8s.io/v1
kind: Deployment
metadata:
  name: r1-http-client
spec:
  replicas: 1
  template:
    metadata:
      labels:
        app: http-client
    spec:
      containers:
      - name: http-client-2
`
	)

	var cfca contextForCompositeApp
	var capcl interface{}
	var cap interface{}
	var sap interface{}

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
					orchMod.DeploymentIntentGroupKey{Name: dig, Project: project, CompositeApp: compApp, Version: version}.String(): {
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
								"\"name\":\"" + deploymentName1 + "\"," +
								"\"container-name\":\"" + containerName1 + "\"}" +
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
								"\"resource\":{\"name\":\"cpu\", \"requests\":1, \"limits\":1}" +
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

		cap, err = cfca.context.AddApp(cfca.compositeAppHandle, app1)
		Expect(err).To(BeNil())
		capcl, err = cfca.context.AddCluster(cap, "provider1-cluster1")
		Expect(err).To(BeNil())

		sap, err = cfca.context.AddApp(cfca.compositeAppHandle, app2)
		Expect(err).To(BeNil())
		_, err = cfca.context.AddCluster(sap, "provider1-cluster2")
		Expect(err).To(BeNil())
	})

	Describe("Update context", func() {
		It("*** GINKGO TESTCASE: successful allocatable-resource update-context", func() {
			_, err := cfca.context.AddResource(capcl, deploymentName1+"+Deployment", deploymentSpec)
			Expect(err).To(BeNil())

			contextID := fmt.Sprintf("%v", cfca.ctxval)
			err = action.UpdateAppContext("hpa-action-controller", contextID)
			Expect(err).To(BeNil())
		})

		It("*** GINKGO TESTCASE: successful allocatable-resource update-context with replica-count > 1", func() {
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
						"\"replicas\":" + fmt.Sprintf("%d", numberOfReplicas2) + "," +
						"\"name\":\"" + deploymentName1 + "\"," +
						"\"container-name\":\"" + containerName1 + "\"}" +
						"}"),
			}
			_, err := cfca.context.AddResource(capcl, deploymentName1+"+Deployment", deploymentSpec)
			Expect(err).To(BeNil())

			contextID := fmt.Sprintf("%v", cfca.ctxval)
			err = action.UpdateAppContext("hpa-action-controller", contextID)
			Expect(err).To(BeNil())
		})

		It("*** GINKGO TESTCASE: successful allocatable-resource update-context with replica-count = 0", func() {
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
						"\"replicas\":" + fmt.Sprintf("%d", numberOfReplicas0) + "," +
						"\"name\":\"" + deploymentName1 + "\"," +
						"\"container-name\":\"" + containerName1 + "\"}" +
						"}"),
			}
			_, err := cfca.context.AddResource(capcl, deploymentName1+"+Deployment", deploymentSpec)
			Expect(err).To(BeNil())

			contextID := fmt.Sprintf("%v", cfca.ctxval)
			err = action.UpdateAppContext("hpa-action-controller", contextID)
			Expect(err).To(BeNil())
		})

		It("*** GINKGO TESTCASE: successful memory allocatable-resource update-context even if limits is zero", func() {
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
						"\"resource\":{\"name\":\"memory\", \"requests\":1000}" +
						"}" +
						"}"),
			}

			_, err := cfca.context.AddResource(capcl, deploymentName1+"+Deployment", deploymentSpec)
			Expect(err).To(BeNil())

			contextID := fmt.Sprintf("%v", cfca.ctxval)
			err = action.UpdateAppContext("hpa-action-controller", contextID)
			Expect(err).To(BeNil())
		})

		It("*** GINKGO TESTCASE: successful memory allocatable-resource update-context even if limits is zero", func() {
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
						"\"resource\":{\"name\":\"memory\", \"requests\":1000}" +
						"}" +
						"}"),
			}

			_, err := cfca.context.AddResource(capcl, deploymentName1+"+Deployment", deploymentSpec)
			Expect(err).To(BeNil())

			contextID := fmt.Sprintf("%v", cfca.ctxval)
			err = action.UpdateAppContext("hpa-action-controller", contextID)
			Expect(err).To(BeNil())
		})

		It("*** GINKGO TESTCASE: successful non-allocatable-resource update-context", func() {
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
						"\"resource\":{\"key\":\"cpu\", \"value\":\"cpu-value\"}" +
						"}" +
						"}"),
			}

			_, err := cfca.context.AddResource(capcl, deploymentName1+"+Deployment", deploymentSpec)
			Expect(err).To(BeNil())

			contextID := fmt.Sprintf("%v", cfca.ctxval)
			err = action.UpdateAppContext("hpa-action-controller", contextID)
			Expect(err).To(BeNil())
		})

		It("*** GINKGO TESTCASE: successful allocatable-resource update-context even if limits is zero", func() {
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
						"\"resource\":{\"name\":\"cpu\", \"requests\":1}" +
						"}" +
						"}"),
			}

			_, err := cfca.context.AddResource(capcl, deploymentName1+"+Deployment", deploymentSpec)
			Expect(err).To(BeNil())

			contextID := fmt.Sprintf("%v", cfca.ctxval)
			err = action.UpdateAppContext("hpa-action-controller", contextID)
			Expect(err).To(BeNil())
		})

		It("*** GINKGO TESTCASE: successful update-context due to non-presence of metadata in deployment-spec", func() {
			_, err := cfca.context.AddResource(capcl, deploymentName1+"+Deployment", badDeploymentSpecNoMeta)
			Expect(err).To(BeNil())

			contextID := fmt.Sprintf("%v", cfca.ctxval)
			err = action.UpdateAppContext("hpa-action-controller", contextID)
			Expect(err).To(BeNil())
		})

		It("*** GINKGO TESTCASE: successful update-context due to non-presence of spec in deployment-spec", func() {
			_, err := cfca.context.AddResource(capcl, deploymentName1+"+Deployment", badDeploymentSpec1)
			Expect(err).To(BeNil())

			contextID := fmt.Sprintf("%v", cfca.ctxval)
			err = action.UpdateAppContext("hpa-action-controller", contextID)
			Expect(err).To(BeNil())
		})

		It("*** GINKGO TESTCASE: successful update-context due to non-presence of spec template in deployment-spec", func() {
			_, err := cfca.context.AddResource(capcl, deploymentName1+"+Deployment", badDeploymentSpec2)
			Expect(err).To(BeNil())

			contextID := fmt.Sprintf("%v", cfca.ctxval)
			err = action.UpdateAppContext("hpa-action-controller", contextID)
			Expect(err).To(BeNil())
		})

		It("*** GINKGO TESTCASE: unsuccessful update-context due to invalid deploymentspec in etcd", func() {
			contextID := fmt.Sprintf("%v", cfca.ctxval)
			err := action.UpdateAppContext("hpa-action-controller", contextID)
			Expect(err).To(HaveOccurred())
		})

		It("*** GINKGO TESTCASE: unsuccessful update-context due to bad deployment-name allocatable-resource hpa-resource spec", func() {
			_, err := cfca.context.AddResource(capcl, deploymentName1+"+Deployment", badDeploymentSpec3)
			Expect(err).To(BeNil())

			contextID := fmt.Sprintf("%v", cfca.ctxval)
			err = action.UpdateAppContext("hpa-action-controller", contextID)
			Expect(err).To(HaveOccurred())
		})

		It("*** GINKGO TESTCASE: unsuccessful update-context due to bad deployment-name non-allocatable-resource hpa-resource spec", func() {
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
						"\"resource\":{\"key\":\"cpu\", \"value\":\"cpu-value\"}" +
						"}" +
						"}"),
			}
			_, err := cfca.context.AddResource(capcl, deploymentName1+"+Deployment", badDeploymentSpec3)
			Expect(err).To(BeNil())

			contextID := fmt.Sprintf("%v", cfca.ctxval)
			err = action.UpdateAppContext("hpa-action-controller", contextID)
			Expect(err).To(HaveOccurred())
		})

		It("*** GINKGO TESTCASE: unsuccessful update-context due to bad container-name allocatable-resource hpa-resource spec", func() {
			_, err := cfca.context.AddResource(capcl, deploymentName1+"+Deployment", badDeploymentSpec4)
			Expect(err).To(BeNil())

			contextID := fmt.Sprintf("%v", cfca.ctxval)
			err = action.UpdateAppContext("hpa-action-controller", contextID)
			Expect(err).To(HaveOccurred())
		})

		It("*** GINKGO TESTCASE: successful update-context when hpa-intent app is associated with composite-app with NO apps", func() {
			(mdb.Items[0])[orchMod.AppKey{App: "", Project: project, CompositeApp: compApp, CompositeAppVersion: version}.String()] = nil
			contextID := fmt.Sprintf("%v", cfca.ctxval)
			err := action.UpdateAppContext("hpa-action-controller", contextID)
			Expect(err).To(BeNil())
		})

		It("*** GINKGO TESTCASE: successful update-context when hpa-intent app is not one of composite-app apps", func() {
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
			err := action.UpdateAppContext("hpa-action-controller", contextID)
			Expect(err).To(BeNil())
		})

		It("*** GINKGO TESTCASE: successful update-context for non-existing hpa-intents", func() {
			(mdb.Items[0])[hpaMod.HpaIntentKey{IntentName: "",
				Project: project, CompositeApp: compApp,
				Version: version, DeploymentIntentGroup: dig}.String()] = nil
			contextID := fmt.Sprintf("%v", cfca.ctxval)
			err := action.UpdateAppContext("hpa-action-controller", contextID)
			Expect(err).To(BeNil())
		})

		It("*** GINKGO TESTCASE: successful update-context for non-existing hpa-intent", func() {
			(mdb.Items[0])[hpaMod.HpaIntentKey{IntentName: "",
				Project: project, CompositeApp: compApp,
				Version: version, DeploymentIntentGroup: dig}.String()] = map[string][]byte{
				"HpaPlacementControllerMetadata": []byte(
					"{" +
						"\"metadata\" : {" +
						"\"Name\":\"" + hpaIntentName2 + "\"," +
						"\"Description\":\"Test Intent for unit testing\"," +
						"\"UserData1\": \"userData1\"," +
						"\"UserData2\":\"userData2\"}," +
						"\"spec\" : {" +
						"\"app-name\":\"" + app1 + "\"}" +
						"}"),
			}
			contextID := fmt.Sprintf("%v", cfca.ctxval)
			err := action.UpdateAppContext("hpa-action-controller", contextID)
			Expect(err).To(BeNil())
		})

		It("*** GINKGO TESTCASE: successful update-context for non-existing hpa-consumers", func() {
			(mdb.Items[0])[hpaMod.HpaConsumerKey{ConsumerName: "", IntentName: hpaIntentName1,
				Project: project, CompositeApp: compApp,
				Version: version, DeploymentIntentGroup: dig}.String()] = nil
			contextID := fmt.Sprintf("%v", cfca.ctxval)
			err := action.UpdateAppContext("hpa-action-controller", contextID)
			Expect(err).To(BeNil())
		})

		It("*** GINKGO TESTCASE: successful update-context for non-existing hpa-consumer", func() {
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
			err := action.UpdateAppContext("hpa-action-controller", contextID)
			Expect(err).To(BeNil())
		})

		It("*** GINKGO TESTCASE: successful update-context due to non-existing hpa-resources", func() {
			(mdb.Items[0])[hpaMod.HpaResourceKey{ResourceName: "", ConsumerName: hpaConsumerName1, IntentName: hpaIntentName1,
				Project: project, CompositeApp: compApp,
				Version: version, DeploymentIntentGroup: dig}.String()] = nil
			_, err := cfca.context.AddResource(capcl, deploymentName1+"+Deployment", badDeploymentSpec3)
			Expect(err).To(BeNil())

			contextID := fmt.Sprintf("%v", cfca.ctxval)
			err = action.UpdateAppContext("hpa-action-controller", contextID)
			Expect(err).To(BeNil())
		})

		It("*** GINKGO TESTCASE: successful update-context due to non-existing hpa-resource", func() {
			(mdb.Items[0])[hpaMod.HpaResourceKey{ResourceName: "", ConsumerName: hpaConsumerName1, IntentName: hpaIntentName1,
				Project: project, CompositeApp: compApp,
				Version: version, DeploymentIntentGroup: dig}.String()] = map[string][]byte{}
			_, err := cfca.context.AddResource(capcl, deploymentName1+"+Deployment", badDeploymentSpec3)
			Expect(err).To(BeNil())

			contextID := fmt.Sprintf("%v", cfca.ctxval)
			err = action.UpdateAppContext("hpa-action-controller", contextID)
			Expect(err).To(BeNil())
		})

		It("*** GINKGO TESTCASE: failed update-context for non-existing resource deployment spec", func() {
			_, err := cfca.context.AddResource(capcl, deploymentName1+"+Deployment", badDeploymentSpec)
			Expect(err).To(BeNil())

			contextID := fmt.Sprintf("%v", cfca.ctxval)
			err = action.UpdateAppContext("hpa-action-controller", contextID)
			Expect(err).To(HaveOccurred())
		})

		It("*** GINKGO TESTCASE: failed update-context with nil contextID", func() {
			err := action.UpdateAppContext("hpa-action-controller", "")
			Expect(err).To(HaveOccurred())
		})

		It("*** GINKGO TESTCASE: successful update-context with nil contextID", func() {
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
						"\"name\":\"\"," +
						"\"container-name\":\"" + containerName1 + "\"}" +
						"}"),
			}

			(mdb.Items[0])[hpaMod.HpaConsumerKey{ConsumerName: hpaConsumerName1, IntentName: hpaIntentName1,
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
						"\"name\":\"\"," +
						"\"container-name\":\"" + containerName1 + "\"}" +
						"}"),
			}

			contextID := fmt.Sprintf("%v", cfca.ctxval)
			err := action.UpdateAppContext("hpa-action-controller", contextID)
			Expect(err).To(BeNil())
		})
	})
})
