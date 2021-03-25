// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2020 Intel Corporation

package servicediscovery_test

import (
	"encoding/json"
	"fmt"

	"github.com/golang/mock/gomock"
	"github.com/golang/protobuf/proto"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"gopkg.in/yaml.v2"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	"github.com/open-ness/EMCO/src/clm/pkg/cluster"
	"github.com/open-ness/EMCO/src/sds/internal/servicediscovery"
	"github.com/open-ness/EMCO/src/dtc/pkg/module"
	"github.com/open-ness/EMCO/src/orchestrator/pkg/appcontext"
	"github.com/open-ness/EMCO/src/orchestrator/pkg/infra/contextdb"
	"github.com/open-ness/EMCO/src/orchestrator/pkg/infra/db"
	"github.com/open-ness/EMCO/src/orchestrator/pkg/module/controller"
	mtypes "github.com/open-ness/EMCO/src/orchestrator/pkg/module/types"
	"github.com/open-ness/EMCO/src/rsync/pkg/grpc/installapp"
	pkgerrors "github.com/pkg/errors"
)

type contextForCompositeApp struct {
	context            appcontext.AppContext
	ctxval             interface{}
	compositeAppHandle interface{}
}

// rpcMsg impements the gomock.Matcher interface
type rpcMsg struct {
	msg proto.Message
}

type _grpcReq proto.Message

type _grpcRsp []interface{}

type grpcSignature struct {
	grpcReq _grpcReq
	grpcRsp _grpcRsp
}

func (r *rpcMsg) Matches(msg interface{}) bool {
	m, ok := msg.(proto.Message)
	if !ok {
		return false
	}
	return proto.Equal(m, r.msg)
}

func (r *rpcMsg) String() string {
	return fmt.Sprintf("is %s", r.msg)
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

func CreateClusterLabel(provider, clu string) error {

	client := cluster.NewClusterClient()

	cp := cluster.ClusterProvider{
		Metadata: mtypes.Metadata{
			Name:        provider,
			Description: "add provider",
			UserData1:   "user data1",
			UserData2:   "user data2",
		},
	}

	_, _ = client.CreateClusterProvider(cp, false)
	c := cluster.Cluster{
		Metadata: mtypes.Metadata{
			Name:        clu,
			Description: "add cluster",
			UserData1:   "user data1",
			UserData2:   "user data2",
		},
	}
	cc := cluster.ClusterContent{
		Kubeconfig: "dummydata",
	}
	_, _ = client.CreateCluster(provider, c, cc)

	cl := cluster.ClusterLabel{
		LabelName: "networkpolicy-supported",
	}

	_, _ = client.CreateClusterLabel(provider, clu, cl, false)

	return nil

}

func createService(serviceName string) (string, error) {

	var appServicePorts []corev1.ServicePort

	var externalport intstr.IntOrString
	var appServicePort corev1.ServicePort

	externalport = intstr.IntOrString{IntVal: 30080}
	appServicePort = corev1.ServicePort{
		Name:       "30080",
		Protocol:   corev1.ProtocolTCP,
		Port:       30080,
		TargetPort: externalport,
	}
	appServicePorts = append(appServicePorts, appServicePort)

	service := servicediscovery.ServiceResource{
		APIVersion: "v1",
		Kind:       "Service",
		MetaData: metav1.ObjectMeta{
			Name:      serviceName,
			Namespace: "default",
		},
		Specification: servicediscovery.Specs{
			Ports:           appServicePorts,
			ClusterIP:       corev1.ClusterIPNone,
			SessionAffinity: "None",
			Types:           corev1.ServiceTypeNodePort,
		},
	}

	serviceData, err := yaml.Marshal(&service)
	if err != nil {
		return "", err
	}

	return string(serviceData), nil
}

type op func(arg0, arg1 interface{}, arg2 ...interface{}) *gomock.Call

// testMockGrpc ... Add API to grpc mock
func testMockGrpc(f op, grpcReq _grpcReq, grpcResp _grpcRsp) {
	arg2 := gomock.Any()

	if grpcReq != nil {
		arg2 = &rpcMsg{msg: grpcReq}
	}

	f(gomock.Any(), arg2).Return(grpcResp...).AnyTimes()

}

var _ = Describe("Action", func() {

	var (
		mdb *db.MockDB
		edb *contextdb.MockConDb

		TGI    module.TrafficGroupIntent
		TGIDBC *module.TrafficGroupIntentDbClient

		ISI    module.InboundServerIntent
		ISIDBC *module.InboundServerIntentDbClient

		ICI    module.InboundClientsIntent
		ICIDBC *module.InboundClientsIntentDbClient

		clusterResourceStatus = "{\"ready\":false,\"resourceCount\":0,\"serviceStatuses\":[{\"kind\":\"Service\",\"apiVersion\":\"v1\",\"metadata\":{\"name\":\"server-svc\",\"namespace\":\"default\",\"selfLink\":\"/api/v1/namespaces/default/services/http-service\",\"uid\":\"c3ea5728-c6be-469c-9e5a-32be20b31298\",\"resourceVersion\":\"38804882\",\"creationTimestamp\":\"2021-03-02T00:54:20Z\",\"labels\":{\"app.kubernetes.io/instance\":\"r1\",\"app.kubernetes.io/managed-by\":\"dccf\",\"app.kubernetes.io/name\":\"http-server\",\"app.kubernetes.io/version\":\"1.16.0\",\"emco/deployment-id\":\"5008073919127743612-http-server\",\"helm.sh/chart\":\"http-server-0.1.0\"},\"annotations\":{\"kubectl.kubernetes.io/last-applied-configuration\":\"{\\\"apiVersion\\\":\\\"v1\\\",\\\"kind\\\":\\\"Service\\\",\\\"metadata\\\":{\\\"annotations\\\":{},\\\"labels\\\":{\\\"app.kubernetes.io/instance\\\":\\\"r1\\\",\\\"app.kubernetes.io/managed-by\\\":\\\"dccf\\\",\\\"app.kubernetes.io/name\\\":\\\"http-server\\\",\\\"app.kubernetes.io/version\\\":\\\"1.16.0\\\",\\\"emco/deployment-id\\\":\\\"5008073919127743612-http-server\\\",\\\"helm.sh/chart\\\":\\\"http-server-0.1.0\\\"},\\\"name\\\":\\\"http-service\\\",\\\"namespace\\\":\\\"default\\\"},\\\"spec\\\":{\\\"ports\\\":[{\\\"name\\\":\\\"http-service-tcp\\\",\\\"nodePort\\\":30080,\\\"port\\\":30080,\\\"protocol\\\":\\\"TCP\\\",\\\"targetPort\\\":3333}],\\\"selector\\\":{\\\"app.kubernetes.io/instance\\\":\\\"r1\\\",\\\"app.kubernetes.io/managed-by\\\":\\\"dccf\\\",\\\"app.kubernetes.io/name\\\":\\\"http-server\\\"},\\\"type\\\":\\\"LoadBalancer\\\"}}\\n\"}},\"spec\":{\"ports\":[{\"name\":\"http-service-tcp\",\"protocol\":\"TCP\",\"port\":30080,\"targetPort\":3333,\"nodePort\":30080}],\"selector\":{\"app.kubernetes.io/instance\":\"r1\",\"app.kubernetes.io/managed-by\":\"dccf\",\"app.kubernetes.io/name\":\"http-server\"},\"clusterIP\":\"10.0.188.28\",\"type\":\"LoadBalancer\",\"sessionAffinity\":\"None\",\"externalTrafficPolicy\":\"Cluster\"},\"status\":{\"loadBalancer\":{\"ingress\":[{\"ip\":\"20.62.186.71\"}]}}}]}"

		expectedOut string = `apiVersion: v1
kind: Service
metadata:
  name: server-svc
  generatename: ""
  namespace: ""
  selflink: ""
  uid: ""
  resourceversion: ""
  generation: 0
  creationtimestamp: "0001-01-01T00:00:00Z"
  deletiontimestamp: null
  deletiongraceperiodseconds: null
  labels: {}
  annotations: {}
  ownerreferences: []
  finalizers: []
  clustername: ""
  managedfields: []
spec:
  clusterIP: None
  ports:
  - name: "30080"
    protocol: TCP
    appprotocol: null
    port: 30080
    targetport:
      type: 0
      intval: 30080
      strval: ""
    nodeport: 0
  sessionAffinity: None
  type: ClusterIP
`
	)

	BeforeEach(func() {
		//Add rsync as controler to mongodb
		mdb = &db.MockDB{
			Items: []map[string]map[string][]byte{
				{
					controller.ControllerKey{ControllerName: "rsync"}.String(): {
						"controllermetadata": []byte(
							"{\"metadata\":{" +
								"\"name\":\"rsync\"" +
								"}," +
								"\"spec\":{" +
								"\"host\":\"132.156.0.10\"," +
								"\"port\": 8080 }}"),
					},
					controller.ControllerKey{ControllerName: ""}.String(): {
						"controllermetadata": []byte(
							"{\"metadata\":{" +
								"\"name\":\"rsync\"" +
								"}," +
								"\"spec\":{" +
								"\"host\":\"132.156.0.10\"," +
								"\"port\": 8080 }}"),
					},
				},
			},
		}
		mdb.Err = nil
		db.DBconn = mdb
		edb = new(contextdb.MockConDb)
		edb.Err = nil
		contextdb.Db = edb

		TGIDBC = module.NewTrafficGroupIntentClient()
		TGI = module.TrafficGroupIntent{
			Metadata: module.Metadata{
				Name:        "testtgi",
				Description: "traffic group intent",
				UserData1:   "user data1",
				UserData2:   "user data2",
			},
		}

		ISIDBC = module.NewServerInboundIntentClient()
		ISI = module.InboundServerIntent{
			Metadata: module.Metadata{
				Name:        "testisi",
				Description: "inbound server intent",
				UserData1:   "user data1",
				UserData2:   "user data2",
			},
			Spec: module.InbondServerIntentSpec{
				AppName:         "server",
				AppLabel:        "app=server",
				ServiceName:     "server-svc",
				ExternalName:    "",
				Port:            4443,
				Protocol:        "TCP",
				ExternalSupport: false,
			},
		}

		ICIDBC = module.NewClientsInboundIntentClient()
		ICI = module.InboundClientsIntent{
			Metadata: module.Metadata{
				Name:        "testici",
				Description: "inbound client intent",
				UserData1:   "user data1",
				UserData2:   "user data2",
			},
			Spec: module.InboundClientsIntentSpec{
				AppName:     "client",
				AppLabel:    "app=client",
				ServiceName: "client-svc",
				Namespaces:  []string{},
				IpRange:     []string{},
			},
		}
	})

	Describe("App context", func() {
		It("cover invalid context error", func() {
			var ac appcontext.AppContext
			_, err := ac.LoadAppContext("dummycontextid")
			//Expect(err).To(BeNil())
			// TODO: Pls check why the BeNil() fails
			Expect(err).To(HaveOccurred())
			err = servicediscovery.DeployServiceEntry(ac, "dummycontextid", "server", "client", "server-svc")
			Expect(err).To(HaveOccurred())
		})
		It("cover invalid meta data error", func() {
			context := appcontext.AppContext{}
			ctxval, err := context.InitAppContext()
			Expect(err).To(BeNil())
			contextID := fmt.Sprintf("%v", ctxval)
			err = servicediscovery.DeployServiceEntry(context, contextID, "server1", "client", "server-svc")
			Expect(err).To(HaveOccurred())
		})
		It("cover error getting server inbound intents", func() {
			cfca, err := makeAppContextForCompositeApp("project1", "ca", "v2", "r1", "dig", "n1", "app")
			Expect(err).To(BeNil())
			sap, err := cfca.context.AddApp(cfca.compositeAppHandle, "server")
			Expect(err).To(BeNil())
			sapc1, err := cfca.context.AddCluster(sap, "provider1+cluster1")
			Expect(err).To(BeNil())
			_, err = cfca.context.AddResource(sapc1, "r1", "dummy test resource")
			Expect(err).To(BeNil())
			resOrder, err := json.Marshal(map[string][]string{"resorder": []string{"r1"}})
			_, err = cfca.context.AddInstruction(sapc1, "resource", "order", string(resOrder))

			cap, err := cfca.context.AddApp(cfca.compositeAppHandle, "client")
			Expect(err).To(BeNil())
			capc2, err := cfca.context.AddCluster(cap, "provider2+cluster2")
			Expect(err).To(BeNil())
			_, err = cfca.context.AddResource(capc2, "r1", "dummy test resource")
			Expect(err).To(BeNil())
			resOrder, err = json.Marshal(map[string][]string{"resorder": []string{"r1"}})
			_, err = cfca.context.AddInstruction(capc2, "resource", "order", string(resOrder))
			contextID := fmt.Sprintf("%v", cfca.ctxval)
			var ac appcontext.AppContext
			_, err = ac.LoadAppContext(contextID)
			Expect(err).To(BeNil())
			err = servicediscovery.DeployServiceEntry(ac, contextID, "server", "client", "server-svc")
			Expect(err).To(HaveOccurred())
		})
		It("cover error getting clients inbound intents", func() {
			cfca, err := makeAppContextForCompositeApp("project1", "ca", "v2", "r1", "dig", "n1", "app")
			Expect(err).To(BeNil())
			sap, err := cfca.context.AddApp(cfca.compositeAppHandle, "server")
			Expect(err).To(BeNil())
			sapc1, err := cfca.context.AddCluster(sap, "provider1+cluster1")
			Expect(err).To(BeNil())
			_, err = cfca.context.AddResource(sapc1, "r1", "dummy test resource")
			Expect(err).To(BeNil())
			resOrder, err := json.Marshal(map[string][]string{"resorder": []string{"r1"}})
			_, err = cfca.context.AddInstruction(sapc1, "resource", "order", string(resOrder))

			cap, err := cfca.context.AddApp(cfca.compositeAppHandle, "client")
			Expect(err).To(BeNil())
			capc2, err := cfca.context.AddCluster(cap, "provider2+cluster2")
			Expect(err).To(BeNil())
			_, err = cfca.context.AddResource(capc2, "r1", "dummy test resource")
			Expect(err).To(BeNil())
			resOrder, err = json.Marshal(map[string][]string{"resorder": []string{"r1"}})
			_, err = cfca.context.AddInstruction(capc2, "resource", "order", string(resOrder))
			tgi, err := (*TGIDBC).CreateTrafficGroupIntent(TGI, "project1", "ca", "v2", "dig", false)
			Expect(tgi).To(Equal(TGI))
			Expect(err).To(BeNil())
			isi, err := (*ISIDBC).CreateServerInboundIntent(ISI, "project1", "ca", "v2", "dig", "testtgi", false)
			Expect(isi).To(Equal(ISI))
			Expect(err).To(BeNil())
			contextID := fmt.Sprintf("%v", cfca.ctxval)
			var ac appcontext.AppContext
			_, err = ac.LoadAppContext(contextID)
			Expect(err).To(BeNil())
			err = servicediscovery.DeployServiceEntry(ac, contextID, "server", "client", "server-svc")
			Expect(err).To(HaveOccurred())
		})
		It("cover invalid cluster name", func() {
			cfca, err := makeAppContextForCompositeApp("project1", "ca", "v2", "r1", "dig", "n1", "app")
			Expect(err).To(BeNil())
			sap, err := cfca.context.AddApp(cfca.compositeAppHandle, "server")
			Expect(err).To(BeNil())
			sapc1, err := cfca.context.AddCluster(sap, "provider1-cluster1")
			Expect(err).To(BeNil())
			_, err = cfca.context.AddResource(sapc1, "r1", "dummy test resource")
			Expect(err).To(BeNil())
			resOrder, err := json.Marshal(map[string][]string{"resorder": []string{"r1"}})
			_, err = cfca.context.AddInstruction(sapc1, "resource", "order", string(resOrder))

			cap, err := cfca.context.AddApp(cfca.compositeAppHandle, "client")
			Expect(err).To(BeNil())
			capc2, err := cfca.context.AddCluster(cap, "provider2-cluster2")
			Expect(err).To(BeNil())
			_, err = cfca.context.AddResource(capc2, "r1", "dummy test resource")
			Expect(err).To(BeNil())
			resOrder, err = json.Marshal(map[string][]string{"resorder": []string{"r1"}})
			_, err = cfca.context.AddInstruction(capc2, "resource", "order", string(resOrder))
			tgi, err := (*TGIDBC).CreateTrafficGroupIntent(TGI, "project1", "ca", "v2", "dig", false)
			Expect(tgi).To(Equal(TGI))
			Expect(err).To(BeNil())
			isi, err := (*ISIDBC).CreateServerInboundIntent(ISI, "project1", "ca", "v2", "dig", "testtgi", false)
			Expect(isi).To(Equal(ISI))
			Expect(err).To(BeNil())
			ici, err := (*ICIDBC).CreateClientsInboundIntent(ICI, "project1", "ca", "v2", "dig", "testtgi", "testisi", false)
			Expect(ici).To(Equal(ICI))
			Expect(err).To(BeNil())
			contextID := fmt.Sprintf("%v", cfca.ctxval)
			var ac appcontext.AppContext
			_, err = ac.LoadAppContext(contextID)
			Expect(err).To(BeNil())
			err = servicediscovery.DeployServiceEntry(ac, contextID, "server", "client", "server-svc")
			Expect(err).To(HaveOccurred())
		})
		It("successful Creation of Child Appcontext", func() {
			cfca, err := makeAppContextForCompositeApp("project1", "ca", "v2", "r1", "dig", "n1", "app")
			Expect(err).To(BeNil())

			// Server
			sap, err := cfca.context.AddApp(cfca.compositeAppHandle, "server")
			Expect(err).To(BeNil())
			sapc1, err := cfca.context.AddCluster(sap, "provider1+cluster1")
			Expect(err).To(BeNil())
			serviceData, err := createService("server-svc")
			Expect(err).To(BeNil())
			_, err = cfca.context.AddResource(sapc1, "server-svc", serviceData)
			Expect(err).To(BeNil())
			resOrder, err := json.Marshal(map[string][]string{"resorder": []string{"server-svc"}})
			_, err = cfca.context.AddInstruction(sapc1, "resource", "order", string(resOrder))

			// Client
			cap, err := cfca.context.AddApp(cfca.compositeAppHandle, "client")
			Expect(err).To(BeNil())
			capc2, err := cfca.context.AddCluster(cap, "provider2+cluster2")
			Expect(err).To(BeNil())
			_, err = cfca.context.AddResource(capc2, "r1", "dummy test resource")
			Expect(err).To(BeNil())
			resOrder, err = json.Marshal(map[string][]string{"resorder": []string{"r1"}})
			_, err = cfca.context.AddInstruction(capc2, "resource", "order", string(resOrder))

			tgi, err := (*TGIDBC).CreateTrafficGroupIntent(TGI, "project1", "ca", "v2", "dig", false)
			Expect(tgi).To(Equal(TGI))
			Expect(err).To(BeNil())
			isi, err := (*ISIDBC).CreateServerInboundIntent(ISI, "project1", "ca", "v2", "dig", "testtgi", false)
			Expect(isi).To(Equal(ISI))
			Expect(err).To(BeNil())
			ici, err := (*ICIDBC).CreateClientsInboundIntent(ICI, "project1", "ca", "v2", "dig", "testtgi", "testisi", false)
			Expect(ici).To(Equal(ICI))
			Expect(err).To(BeNil())

			_ = CreateClusterLabel("provider1", "cluster1")
			_ = CreateClusterLabel("provider2", "cluster2")

			contextID := fmt.Sprintf("%v", cfca.ctxval)
			var ac appcontext.AppContext
			_, err = ac.LoadAppContext(contextID)
			Expect(err).To(BeNil())

			// Update the cluster resource status
			acrh := "/context/" + contextID + "/app/" + "server" + "/cluster/" + "provider1+cluster1" + "/status/"
			err = cfca.context.UpdateResourceValue(acrh, clusterResourceStatus)
			Expect(err).To(BeNil())

			type appContextStatus struct {
				Status string `json:"status"`
			}

			// Update the app context status to Instantiated
			acStatusHandle := "/context/" + contextID + "/status/"
			err = cfca.context.UpdateStatusValue(acStatusHandle, &appContextStatus{Status: "Instantiated"})
			Expect(err).To(BeNil())

			// Create mock definitions for rsync installapp
			var gs = &grpcSignature{}
			gs.grpcReq = nil
			gs.grpcRsp = []interface{}{&installapp.InstallAppResponse{
				AppContextInstalled: true,
			}, nil}

			testMockGrpc(mockApplication.EXPECT().InstallApp, gs.grpcReq, gs.grpcRsp)

			err = servicediscovery.DeployServiceEntry(ac, contextID, "server", "client", "server-svc")
			Expect(err).To(BeNil())
			// Get the parent composite app meta and the child context ID
			m, err := ac.GetCompositeAppMeta()
			Expect(err).To(BeNil())

			for _, childContextID := range m.ChildContextIDs {
				serviceDiscoveryHandle := "/context/" + childContextID + "/app/service-discovery/cluster/provider2+cluster2/resource/server-svc+Service/"
				var childContext appcontext.AppContext
				_, err := childContext.LoadAppContext(childContextID)
				Expect(err).To(BeNil())
				v, err := childContext.GetValue(serviceDiscoveryHandle)
				Expect(err).To(BeNil())
				Expect(v).To(Equal(expectedOut))
			}
		})

	})
})
