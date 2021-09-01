// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2020 Intel Corporation

package api

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/sirupsen/logrus"

	"github.com/open-ness/EMCO/src/orchestrator/pkg/infra/contextdb"
	"github.com/open-ness/EMCO/src/orchestrator/pkg/infra/db"

	hpaMod "github.com/open-ness/EMCO/src/hpa-plc/pkg/module"
	orchLog "github.com/open-ness/EMCO/src/orchestrator/pkg/infra/logutils"
	orchMod "github.com/open-ness/EMCO/src/orchestrator/pkg/module"
)

func init() {
	hpaIntentJSONFile = "../json-schemas/placement-hpa-intent.json"
	orchLog.SetLoglevel(logrus.InfoLevel)
}

func TestHpaHTTPApi(t *testing.T) {

	fmt.Printf("\n================== hpa-placementcontroller TestAction .. start ==================\n")

	orchLog.SetLoglevel(logrus.InfoLevel)
	RegisterFailHandler(Fail)
	RunSpecs(t, "TestHpaApi")

	fmt.Printf("\n================== hpa-placementcontroller TestAction .. end ==================\n")
}

var _ = Describe("TestHpaApi ***", func() {

	var (
		mdb *db.MockDB
		edb *contextdb.MockConDb

		project string = "p"
		compApp string = "ca"
		version string = "v1"
		digroup string = "dig"
		app1    string = "client"

		hpaIntentName1   string = "testIntent"
		hpaConsumerName1 string = "testConsumer"
		replicaCount     int64  = 1
		deploymentName1  string = "r1-http-client"
		containerName1   string = "http-client-1"
	)

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
					orchMod.DeploymentIntentGroupKey{Name: digroup, Project: project, CompositeApp: compApp, Version: version}.String(): {
						"deploymentintentgroupmetadata": []byte(
							"{" +
								"\"metadata\":{" +
								"\"name\":\"" + digroup + "\"," +
								"\"description\":\"DescriptionTestDeploymentIntentGroup\"," +
								"\"userData1\": \"userData1\"," +
								"\"userData2\": \"userData2\"}" +
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

	})

	It("*** GINKGO API TESTCASE: successful create intents ***", func() {

		reader := bytes.NewBuffer([]byte(`{
			"metadata" : {
				"name": "testIntent",
				"description": "Test Intent used for unit testing",
				"userData1": "data1",
				"userData2": "data2"
			},
			"spec" : {
				"app-name":"app1"
			}
		}`))

		request := httptest.NewRequest("POST", "/v2/projects/p/composite-apps/ca/v1/deployment-intent-groups/dig/hpa-intents", reader)
		resp := executeRequest(request, NewRouter(nil))
		Expect(resp.StatusCode).To(Equal(http.StatusCreated))

		reader = bytes.NewBuffer([]byte(`{
			"metadata" : {
				"name": "testIntent",
				"description": "Test Intent used for unit testing",
				"userData1": "data1",
				"userData2": "data2"
			},
			"spec" : {
				"app-name":"app1"
			}
		}`))

		request = httptest.NewRequest("POST", "/v2/projects/p/composite-apps/ca/v1/deployment-intent-groups/dig/hpa-intents", reader)
		resp = executeRequest(request, NewRouter(nil))
		Expect(resp.StatusCode).To(Equal(http.StatusConflict))
	})

	It("*** GINKGO API TESTCASE: successful create consumer ***", func() {

		(mdb.Items[0])[hpaMod.HpaIntentKey{IntentName: hpaIntentName1,
			Project: project, CompositeApp: compApp,
			Version: version, DeploymentIntentGroup: digroup}.String()] = map[string][]byte{
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
		}

		reader := bytes.NewBuffer([]byte(`{
			"metadata" : {
				"name": "testConsumer",
				"description": "Test Consumer used for unit testing",
				"userData1": "data1",
				"userData2": "data2"
			},
			"spec" : {
				"replicas": 1,
				"name":          "deployment-1",
				"container-name": "container-1"
			}
		}`))

		request := httptest.NewRequest("POST", "/v2/projects/p/composite-apps/ca/v1/deployment-intent-groups/dig/hpa-intents/testIntent/hpa-resource-consumers", reader)
		resp := executeRequest(request, NewRouter(nil))
		Expect(resp.StatusCode).To(Equal(http.StatusCreated))

		reader = bytes.NewBuffer([]byte(`{
			"metadata" : {
				"name": "testConsumer",
				"description": "Test Consumer used for unit testing",
				"userData1": "data1",
				"userData2": "data2"
			},
			"spec" : {
				"replicas": 1,
				"name":          "deployment-1",
				"container-name": "container-1"
			}
		}`))

		request = httptest.NewRequest("POST", "/v2/projects/p/composite-apps/ca/v1/deployment-intent-groups/dig/hpa-intents/testIntent/hpa-resource-consumers", reader)
		resp = executeRequest(request, NewRouter(nil))
		Expect(resp.StatusCode).To(Equal(http.StatusConflict))
	})

	It("*** GINKGO API TESTCASE: successful create allocatable resource ***", func() {

		(mdb.Items[0])[hpaMod.HpaIntentKey{IntentName: hpaIntentName1,
			Project: project, CompositeApp: compApp,
			Version: version, DeploymentIntentGroup: digroup}.String()] = map[string][]byte{
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
		}

		(mdb.Items[0])[hpaMod.HpaConsumerKey{ConsumerName: hpaConsumerName1, IntentName: hpaIntentName1,
			Project: project, CompositeApp: compApp,
			Version: version, DeploymentIntentGroup: digroup}.String()] = map[string][]byte{
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
		}

		reader := bytes.NewBuffer([]byte(`{
			"metadata" : {
				"name": "testResource",
				"description": "Test Resource used for unit testing",
				"userData1": "data1",
				"userData2": "data2"
			},
			"spec" : {
				"allocatable":          true,
				"resource" : {"name":"cpu", "requests":1,"limits":1}
			}
		}`))

		request := httptest.NewRequest("POST", "/v2/projects/p/composite-apps/ca/v1/deployment-intent-groups/dig/hpa-intents/testIntent/hpa-resource-consumers/testConsumer/resource-requirements", reader)
		resp := executeRequest(request, NewRouter(nil))
		Expect(resp.StatusCode).To(Equal(http.StatusCreated))

		reader = bytes.NewBuffer([]byte(`{
			"metadata" : {
				"name": "testResource",
				"description": "Test Resource used for unit testing",
				"userData1": "data1",
				"userData2": "data2"
			},
			"spec" : {
				"allocatable":          true,
				"resource" : {"name":"cpu", "requests":1,"limits":1}
			}
		}`))

		request = httptest.NewRequest("POST", "/v2/projects/p/composite-apps/ca/v1/deployment-intent-groups/dig/hpa-intents/testIntent/hpa-resource-consumers/testConsumer/resource-requirements", reader)
		resp = executeRequest(request, NewRouter(nil))
		Expect(resp.StatusCode).To(Equal(http.StatusConflict))
	})

	It("*** GINKGO API TESTCASE: successful create non-allocatable resource ***", func() {

		(mdb.Items[0])[hpaMod.HpaIntentKey{IntentName: hpaIntentName1,
			Project: project, CompositeApp: compApp,
			Version: version, DeploymentIntentGroup: digroup}.String()] = map[string][]byte{
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
		}

		(mdb.Items[0])[hpaMod.HpaConsumerKey{ConsumerName: hpaConsumerName1, IntentName: hpaIntentName1,
			Project: project, CompositeApp: compApp,
			Version: version, DeploymentIntentGroup: digroup}.String()] = map[string][]byte{
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
		}

		reader := bytes.NewBuffer([]byte(`{
			"metadata" : {
				"name": "testResource",
				"description": "Test Resource used for unit testing",
				"userData1": "data1",
				"userData2": "data2"
			},
			"spec" : {
				"allocatable":          false,
				"resource" : {"key":"vpu", "value":"yes"}
			}
		}`))

		request := httptest.NewRequest("POST", "/v2/projects/p/composite-apps/ca/v1/deployment-intent-groups/dig/hpa-intents/testIntent/hpa-resource-consumers/testConsumer/resource-requirements", reader)
		resp := executeRequest(request, NewRouter(nil))
		Expect(resp.StatusCode).To(Equal(http.StatusCreated))

		reader = bytes.NewBuffer([]byte(`{
			"metadata" : {
				"name": "testResource",
				"description": "Test Resource used for unit testing",
				"userData1": "data1",
				"userData2": "data2"
			},
			"spec" : {
				"allocatable":          false,
				"resource" : {"key":"vpu", "value":"yes"}
			}
		}`))

		request = httptest.NewRequest("POST", "/v2/projects/p/composite-apps/ca/v1/deployment-intent-groups/dig/hpa-intents/testIntent/hpa-resource-consumers/testConsumer/resource-requirements", reader)
		resp = executeRequest(request, NewRouter(nil))
		Expect(resp.StatusCode).To(Equal(http.StatusConflict))
	})

	It("*** GINKGO API TESTCASE: successful intent crud ***", func() {

		reader := bytes.NewBuffer([]byte(`{
			"metadata" : {
				"name": "testIntent",
				"description": "Test Intent used for unit testing",
				"userData1": "data1",
				"userData2": "data2"
			},
			"spec" : {
				"app-name":"app1"
			}
		}`))

		request := httptest.NewRequest("POST", "/v2/projects/p/composite-apps/ca/v1/deployment-intent-groups/dig/hpa-intents", reader)
		resp := executeRequest(request, NewRouter(nil))
		Expect(resp.StatusCode).To(Equal(http.StatusCreated))

		request = httptest.NewRequest("GET", "/v2/projects/p/composite-apps/ca/v1/deployment-intent-groups/dig/hpa-intents/testIntent", nil)
		resp = executeRequest(request, NewRouter(nil))
		Expect(resp.StatusCode).To(Equal(http.StatusOK))

		reader = bytes.NewBuffer([]byte(`{
			"metadata" : {
				"name": "testIntent",
				"description": "Test Intent used for unit testing",
				"userData1": "data1 updated",
				"userData2": "data2 updated"
			},
			"spec" : {
				"app-name":"app2"
			}
		}`))

		request = httptest.NewRequest("PUT", "/v2/projects/p/composite-apps/ca/v1/deployment-intent-groups/dig/hpa-intents/testIntent", reader)
		resp = executeRequest(request, NewRouter(nil))
		Expect(resp.StatusCode).To(Equal(http.StatusOK))

		request = httptest.NewRequest("DELETE", "/v2/projects/p/composite-apps/ca/v1/deployment-intent-groups/dig/hpa-intents/testIntent", nil)
		resp = executeRequest(request, NewRouter(nil))
		Expect(resp.StatusCode).To(Equal(http.StatusNoContent))

	})

	It("*** GINKGO API TESTCASE: successful consumer crud ***", func() {

		(mdb.Items[0])[hpaMod.HpaIntentKey{IntentName: hpaIntentName1,
			Project: project, CompositeApp: compApp,
			Version: version, DeploymentIntentGroup: digroup}.String()] = map[string][]byte{
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
		}

		reader := bytes.NewBuffer([]byte(`{
			"metadata" : {
				"name": "testConsumer",
				"description": "Test Consumer used for unit testing",
				"userData1": "data1",
				"userData2": "data2"
			},
			"spec" : {
				"replicas": 1,
				"name":          "deployment-1",
				"container-name": "container-1"
			}
		}`))

		request := httptest.NewRequest("POST", "/v2/projects/p/composite-apps/ca/v1/deployment-intent-groups/dig/hpa-intents/testIntent/hpa-resource-consumers", reader)
		resp := executeRequest(request, NewRouter(nil))
		Expect(resp.StatusCode).To(Equal(http.StatusCreated))

		request = httptest.NewRequest("GET", "/v2/projects/p/composite-apps/ca/v1/deployment-intent-groups/dig/hpa-intents/testIntent/hpa-resource-consumers/testConsumer", nil)
		resp = executeRequest(request, NewRouter(nil))
		Expect(resp.StatusCode).To(Equal(http.StatusOK))

		reader = bytes.NewBuffer([]byte(`{
			"metadata" : {
				"name": "testConsumer",
				"description": "Test Consumer used for unit testing",
				"userData1": "data1 updated",
				"userData2": "data2 updated"
			},
			"spec" : {
				"replicas": 1,
				"name":          "deployment-2",
				"container-name": "container-3"
			}
		}`))

		request = httptest.NewRequest("PUT", "/v2/projects/p/composite-apps/ca/v1/deployment-intent-groups/dig/hpa-intents/testIntent/hpa-resource-consumers/testConsumer", reader)
		resp = executeRequest(request, NewRouter(nil))
		Expect(resp.StatusCode).To(Equal(http.StatusOK))

		request = httptest.NewRequest("DELETE", "/v2/projects/p/composite-apps/ca/v1/deployment-intent-groups/dig/hpa-intents/testIntent/hpa-resource-consumers/testConsumer", nil)
		resp = executeRequest(request, NewRouter(nil))
		Expect(resp.StatusCode).To(Equal(http.StatusNoContent))

	})

	It("*** GINKGO API TESTCASE: successful allocatable resource crud ***", func() {

		(mdb.Items[0])[hpaMod.HpaIntentKey{IntentName: hpaIntentName1,
			Project: project, CompositeApp: compApp,
			Version: version, DeploymentIntentGroup: digroup}.String()] = map[string][]byte{
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
		}

		(mdb.Items[0])[hpaMod.HpaConsumerKey{ConsumerName: hpaConsumerName1, IntentName: hpaIntentName1,
			Project: project, CompositeApp: compApp,
			Version: version, DeploymentIntentGroup: digroup}.String()] = map[string][]byte{
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
		}

		reader := bytes.NewBuffer([]byte(`{
			"metadata" : {
				"name": "testResource",
				"description": "Test Resource used for unit testing",
				"userData1": "data1",
				"userData2": "data2"
			},
			"spec" : {
				"allocatable":          true,
				"resource" : {"name":"cpu", "requests":1,"limits":1}
			}
		}`))

		request := httptest.NewRequest("POST", "/v2/projects/p/composite-apps/ca/v1/deployment-intent-groups/dig/hpa-intents/testIntent/hpa-resource-consumers/testConsumer/resource-requirements", reader)
		resp := executeRequest(request, NewRouter(nil))
		Expect(resp.StatusCode).To(Equal(http.StatusCreated))

		request = httptest.NewRequest("GET", "/v2/projects/p/composite-apps/ca/v1/deployment-intent-groups/dig/hpa-intents/testIntent/hpa-resource-consumers/testConsumer/resource-requirements/testResource", nil)
		resp = executeRequest(request, NewRouter(nil))
		Expect(resp.StatusCode).To(Equal(http.StatusOK))

		reader = bytes.NewBuffer([]byte(`{
			"metadata" : {
				"name": "testResource",
				"description": "Test Resource used for unit testing update",
				"userData1": "data1 update",
				"userData2": "data2 update"
			},
			"spec" : {
				"allocatable":          true,
				"resource" : {"name":"cpu", "requests":2,"limits":2}
			}
		}`))

		request = httptest.NewRequest("PUT", "/v2/projects/p/composite-apps/ca/v1/deployment-intent-groups/dig/hpa-intents/testIntent/hpa-resource-consumers/testConsumer/resource-requirements/testResource", reader)
		resp = executeRequest(request, NewRouter(nil))
		Expect(resp.StatusCode).To(Equal(http.StatusOK))

		request = httptest.NewRequest("DELETE", "/v2/projects/p/composite-apps/ca/v1/deployment-intent-groups/dig/hpa-intents/testIntent/hpa-resource-consumers/testConsumer/resource-requirements/testResource", nil)
		resp = executeRequest(request, NewRouter(nil))
		Expect(resp.StatusCode).To(Equal(http.StatusNoContent))

	})

	It("*** GINKGO API TESTCASE: successful non-allocatable resource crud ***", func() {

		(mdb.Items[0])[hpaMod.HpaIntentKey{IntentName: hpaIntentName1,
			Project: project, CompositeApp: compApp,
			Version: version, DeploymentIntentGroup: digroup}.String()] = map[string][]byte{
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
		}

		(mdb.Items[0])[hpaMod.HpaConsumerKey{ConsumerName: hpaConsumerName1, IntentName: hpaIntentName1,
			Project: project, CompositeApp: compApp,
			Version: version, DeploymentIntentGroup: digroup}.String()] = map[string][]byte{
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
		}

		reader := bytes.NewBuffer([]byte(`{
			"metadata" : {
				"name": "testResource",
				"description": "Test Resource used for unit testing",
				"userData1": "data1",
				"userData2": "data2"
			},
			"spec" : {
				"allocatable":          false,
				"resource" : {"key":"vpu", "value":"yes"}
			}
		}`))

		request := httptest.NewRequest("POST", "/v2/projects/p/composite-apps/ca/v1/deployment-intent-groups/dig/hpa-intents/testIntent/hpa-resource-consumers/testConsumer/resource-requirements", reader)
		resp := executeRequest(request, NewRouter(nil))
		Expect(resp.StatusCode).To(Equal(http.StatusCreated))

		request = httptest.NewRequest("GET", "/v2/projects/p/composite-apps/ca/v1/deployment-intent-groups/dig/hpa-intents/testIntent/hpa-resource-consumers/testConsumer/resource-requirements/testResource", nil)
		resp = executeRequest(request, NewRouter(nil))
		Expect(resp.StatusCode).To(Equal(http.StatusOK))

		reader = bytes.NewBuffer([]byte(`{
			"metadata" : {
				"name": "testResource",
				"description": "Test Resource used for unit testing updated",
				"userData1": "data1 updated",
				"userData2": "data2 updated"
			},
			"spec" : {
				"allocatable":          false,
				"resource" : {"key":"vpu-2", "value":"no"}
			}
		}`))

		request = httptest.NewRequest("PUT", "/v2/projects/p/composite-apps/ca/v1/deployment-intent-groups/dig/hpa-intents/testIntent/hpa-resource-consumers/testConsumer/resource-requirements/testResource", reader)
		resp = executeRequest(request, NewRouter(nil))
		Expect(resp.StatusCode).To(Equal(http.StatusOK))

		request = httptest.NewRequest("DELETE", "/v2/projects/p/composite-apps/ca/v1/deployment-intent-groups/dig/hpa-intents/testIntent/hpa-resource-consumers/testConsumer/resource-requirements/testResource", nil)
		resp = executeRequest(request, NewRouter(nil))
		Expect(resp.StatusCode).To(Equal(http.StatusNoContent))
	})

	It("*** GINKGO API TESTCASE: unsuccessful create intents .. dependency(project) failed ***", func() {
		reader := bytes.NewBuffer([]byte(`{
			"metadata" : {
				"name": "testIntent",
				"description": "Test Intent used for unit testing",
				"userData1": "data1",
				"userData2": "data2"
			},
			"spec" : {
				"app-name":"app1"
			}
		}`))

		request := httptest.NewRequest("POST", "/v2/projects/p1/composite-apps/ca/v1/deployment-intent-groups/dig/hpa-intents", reader)
		resp := executeRequest(request, NewRouter(nil))
		Expect(resp.StatusCode).To(Equal(http.StatusNotFound))
	})

	It("*** GINKGO API TESTCASE: unsuccessful create intents .. dependency(composite-app) failed ***", func() {
		reader := bytes.NewBuffer([]byte(`{
			"metadata" : {
				"name": "testIntent",
				"description": "Test Intent used for unit testing",
				"userData1": "data1",
				"userData2": "data2"
			},
			"spec" : {
				"app-name":"app1"
			}
		}`))

		request := httptest.NewRequest("POST", "/v2/projects/p/composite-apps/ca1/v1/deployment-intent-groups/dig/hpa-intents", reader)
		resp := executeRequest(request, NewRouter(nil))
		Expect(resp.StatusCode).To(Equal(http.StatusNotFound))
	})

	It("*** GINKGO API TESTCASE: unsuccessful create intents .. dependency(depgroup) failed ***", func() {
		reader := bytes.NewBuffer([]byte(`{
			"metadata" : {
				"name": "testIntent",
				"description": "Test Intent used for unit testing",
				"userData1": "data1",
				"userData2": "data2"
			},
			"spec" : {
				"app-name":"app1"
			}
		}`))

		request := httptest.NewRequest("POST", "/v2/projects/p/composite-apps/ca/v1/deployment-intent-groups/dig1/hpa-intents", reader)
		resp := executeRequest(request, NewRouter(nil))
		Expect(resp.StatusCode).To(Equal(http.StatusNotFound))
	})

	It("*** GINKGO API TESTCASE: unsuccessful create consumer .. dependency(intentname) failed ***", func() {

		(mdb.Items[0])[hpaMod.HpaIntentKey{IntentName: hpaIntentName1,
			Project: project, CompositeApp: compApp,
			Version: version, DeploymentIntentGroup: digroup}.String()] = map[string][]byte{
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
		}

		reader := bytes.NewBuffer([]byte(`{
			"metadata" : {
				"name": "testConsumer",
				"description": "Test Consumer used for unit testing",
				"userData1": "data1",
				"userData2": "data2"
			},
			"spec" : {
				"replicas": 1,
				"name":          "deployment-1",
				"container-name": "container-1"
			}
		}`))

		request := httptest.NewRequest("POST", "/v2/projects/p/composite-apps/ca/v1/deployment-intent-groups/dig/hpa-intents/testIntent1/hpa-resource-consumers", reader)
		resp := executeRequest(request, NewRouter(nil))
		Expect(resp.StatusCode).To(Equal(http.StatusNotFound))

	})

	It("*** GINKGO API TESTCASE: unsuccessful create allocatable resource .. dependency(consumername) failed ***", func() {

		(mdb.Items[0])[hpaMod.HpaIntentKey{IntentName: hpaIntentName1,
			Project: project, CompositeApp: compApp,
			Version: version, DeploymentIntentGroup: digroup}.String()] = map[string][]byte{
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
		}

		(mdb.Items[0])[hpaMod.HpaConsumerKey{ConsumerName: hpaConsumerName1, IntentName: hpaIntentName1,
			Project: project, CompositeApp: compApp,
			Version: version, DeploymentIntentGroup: digroup}.String()] = map[string][]byte{
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
		}

		reader := bytes.NewBuffer([]byte(`{
			"metadata" : {
				"name": "testResource",
				"description": "Test Resource used for unit testing",
				"userData1": "data1",
				"userData2": "data2"
			},
			"spec" : {
				"allocatable":          true,
				"resource" : {"name":"cpu", "requests":1,"limits":1}
			}
		}`))

		request := httptest.NewRequest("POST", "/v2/projects/p/composite-apps/ca/v1/deployment-intent-groups/dig/hpa-intents/testIntent/hpa-resource-consumers/testConsumer1/resource-requirements", reader)
		resp := executeRequest(request, NewRouter(nil))
		Expect(resp.StatusCode).To(Equal(http.StatusNotFound))
	})

})
