package api

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"fmt"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	"github.com/open-ness/EMCO/src/dtc/api/mocks"
	"github.com/open-ness/EMCO/src/dtc/pkg/module"
	orcmod "github.com/open-ness/EMCO/src/orchestrator/pkg/module"
	"github.com/open-ness/EMCO/src/orchestrator/pkg/infra/db"
	pkgerrors "github.com/pkg/errors"
)

func init() {
	inServerIntJSONFile = "../json-schemas/inbound-server.json"
}

var _ = Describe("Inboundserverintenthandler", func() {

	var (
		mdb *db.MockDB
	)
	type testCase struct {
		inputName    string
		inputReader  io.Reader
		inStruct     module.InboundServerIntent
		mockError    error
		mockVal      module.InboundServerIntent
		mockVals     []module.InboundServerIntent
		expectedCode int
		client	     *mocks.InboundServerIntentManager
	}

	BeforeEach(func() {

		mdb = new(db.MockDB)
		mdb.Err = nil
		db.DBconn = mdb
		c := orcmod.NewClient()
		_, err := c.Project.CreateProject(orcmod.Project{MetaData: orcmod.ProjectMetaData{Name: "test-project", Description: "test", UserData1: "userData1", UserData2: "userData2"}}, false)
		if err != nil {
			fmt.Println(err)
			return
		}
		_, err = c.CompositeApp.CreateCompositeApp(orcmod.CompositeApp{Metadata: orcmod.CompositeAppMetaData{Name: "test-compositeapp", Description: "test", UserData1: "userData1", UserData2: "userData2"}, Spec: orcmod.CompositeAppSpec{ Version: "v1"}}, "test-project", false)
		if err != nil {
			fmt.Println(err)
			return
		}

		list := []orcmod.OverrideValues{}
		_, err = c.DeploymentIntentGroup.CreateDeploymentIntentGroup(orcmod.DeploymentIntentGroup{MetaData: orcmod.DepMetaData{Name: "test-dig", Description: "test", UserData1: "userData1", UserData2: "userData2"}, Spec: orcmod.DepSpecData{ Profile: "prof1", Version: "v1", OverrideValuesObj: list, LogicalCloud: "lc1"}}, "test-project", "test-compositeapp", "v1")
		if err != nil {
			fmt.Println(err)
			return
		}
	})

	DescribeTable("Create InboundServerIntent tests",
		func(t testCase) {
			// set up client mock responses

			t.client.On("CreateServerInboundIntent", t.inStruct, "test-project", "test-compositeapp", "v1", "test-dig", "testtrafficgroupintent", false).Return(t.mockVal, t.mockError)

			// make HTTP request
			request := httptest.NewRequest("POST", "/v2/projects/test-project/composite-apps/test-compositeapp/v1/deployment-intent-groups/test-dig/traffic-group-intents/testtrafficgroupintent/inbound-intents", t.inputReader)
			resp := executeRequest(request, NewRouter(t.client))

			//Check returned code
			Expect(resp.StatusCode).To(Equal(t.expectedCode))

			//Check returned body
			got := module.InboundServerIntent{}
			json.NewDecoder(resp.Body).Decode(&got)
			Expect(got).To(Equal(t.mockVal))
		},

		Entry("successful create", testCase{
			expectedCode: http.StatusCreated,
			inputReader: bytes.NewBuffer([]byte(`{
				"metadata": {
					"name": "testinboundserverintent",
					"description": "test inbound server group intent",
					"userData1": "some user data 1",
					"userData2": "some user data 2"
				},
				"spec": {
					"appName":	   "test-app",
					"appLabel":	   "test-applabel",
					"serviceName":	   "test-servicename",
					"externalName":    "test-externalname",
					"port":		   6666,
					"protocol":	   "TCP",
					"externalSupport": false
				}
			}`)),
			inStruct: module.InboundServerIntent{
				Metadata: module.Metadata{
					Name:	     "testinboundserverintent",
					Description: "test inbound server group intent",
					UserData1:   "some user data 1",
					UserData2:   "some user data 2",
				},
				Spec: module.InbondServerIntentSpec{
					AppName:	 "test-app",
					AppLabel:	 "test-applabel",
					ServiceName:	 "test-servicename",
					ExternalName:	 "test-externalname",
					Port:		 6666,
					Protocol:	 "TCP",
					ExternalSupport: false,
				},
			},
			mockError: nil,
			mockVal: module.InboundServerIntent{
				Metadata: module.Metadata{
					Name:	     "testinboundserverintent",
					Description: "test inbound server group intent",
					UserData1:   "some user data 1",
					UserData2:   "some user data 2",
				},
				Spec: module.InbondServerIntentSpec{
					AppName:	 "test-app",
					AppLabel:	 "test-applabel",
					ServiceName:	 "test-servicename",
					ExternalName:	 "test-externalname",
					Port:		 6666,
					Protocol:	 "TCP",
					ExternalSupport: false,
				},
			},
			client: &mocks.InboundServerIntentManager{},
		}),

		Entry("fails due to empty body", testCase{
			expectedCode: http.StatusBadRequest,
			inStruct:     module.InboundServerIntent{},
			mockError:    nil,
			mockVal:      module.InboundServerIntent{},
			client:       &mocks.InboundServerIntentManager{},
		}),

		Entry("fails due to missing name", testCase{
			expectedCode: http.StatusBadRequest,
			inStruct:     module.InboundServerIntent{},
			inputReader: bytes.NewBuffer([]byte(`{
				"metadata": {
					"description": "test inbound server group intent",
					"userData1": "some user data 1",
					"userData2": "some user data 2"
				},
				"spec": {
					"appName":	   "test-app",
					"appLabel":	   "test-applabel",
					"serviceName":	   "test-servicename",
					"externalName":    "test-externalname",
					"port":		   6666,
					"protocol":	   "TCP",
					"externalSupport": false
				}
			}`)),
			mockError:    nil,
			mockVal:      module.InboundServerIntent{},
			client:       &mocks.InboundServerIntentManager{},
		}),
		Entry("fails due to missing app name", testCase{
			expectedCode: http.StatusBadRequest,
			inStruct:     module.InboundServerIntent{},
			inputReader: bytes.NewBuffer([]byte(`{
				"metadata": {
					"name": "testinboundserverintent",
					"description": "test inbound server group intent",
					"userData1": "some user data 1",
					"userData2": "some user data 2"
				},
				"spec": {
					"appLabel":	   "test-applabel",
					"serviceName":	   "test-servicename",
					"externalName":    "test-externalname",
					"port":		   6666,
					"protocol":	   "TCP",
					"externalSupport": false
				}
			}`)),
			mockError:    nil,
			mockVal:      module.InboundServerIntent{},
			client:       &mocks.InboundServerIntentManager{},
		}),
		Entry("fails due to missing port", testCase{
			expectedCode: http.StatusBadRequest,
			inStruct:     module.InboundServerIntent{},
			inputReader: bytes.NewBuffer([]byte(`{
				"metadata": {
					"name": "testinboundserverintent",
					"description": "test inbound server group intent",
					"userData1": "some user data 1",
					"userData2": "some user data 2"
				},
				"spec": {
					"appName":	   "test-app",
					"appLabel":	   "test-applabel",
					"serviceName":	   "test-servicename",
					"externalName":    "test-externalname",
					"protocol":	   "TCP",
					"externalSupport": false
				}
			}`)),
			mockError:    nil,
			mockVal:      module.InboundServerIntent{},
			client:       &mocks.InboundServerIntentManager{},
		}),
		Entry("fails due to missing protocol", testCase{
			expectedCode: http.StatusBadRequest,
			inStruct:     module.InboundServerIntent{},
			inputReader: bytes.NewBuffer([]byte(`{
				"metadata": {
					"name": "testinboundserverintent",
					"description": "test inbound server group intent",
					"userData1": "some user data 1",
					"userData2": "some user data 2"
				},
				"spec": {
					"appName":	   "test-app",
					"appLabel":	   "test-applabel",
					"serviceName":	   "test-servicename",
					"externalName":    "test-externalname",
					"port":		   6666,
					"externalSupport": false
				}
			}`)),
			mockError:    nil,
			mockVal:      module.InboundServerIntent{},
			client:       &mocks.InboundServerIntentManager{},
		}),
		Entry("fails due to empty app name", testCase{
			expectedCode: http.StatusBadRequest,
			inStruct:     module.InboundServerIntent{},
			inputReader: bytes.NewBuffer([]byte(`{
				"metadata": {
					"name": "testinboundserverintent",
					"description": "test inbound server group intent",
					"userData1": "some user data 1",
					"userData2": "some user data 2"
				},
				"spec": {
					"appName":	   "",
					"appLabel":	   "test-applabel",
					"serviceName":	   "test-servicename",
					"externalName":    "test-externalname",
					"port":		   6666,
					"protocol":	   "TCP",
					"externalSupport": false
				}
			}`)),
			mockError:    nil,
			mockVal:      module.InboundServerIntent{},
			client:       &mocks.InboundServerIntentManager{},
		}),
		Entry("fails due to other json validation error", testCase{
			// name field has an '=' character
			expectedCode: http.StatusBadRequest,
			inputReader: bytes.NewBuffer([]byte(`{
				"metadata": {
					"name": "test=testinboundserverintent",
					"description": "test inbound server group intent",
					"userData1": "some user data 1",
					"userData2": "some user data 2"
				},
				"spec": {
					"appName":	   "test-app",
					"appLabel":	   "test-applabel",
					"serviceName":	   "test-servicename",
					"externalName":    "test-externalname",
					"port":		   6666,
					"protocol":	   "TCP",
					"externalSupport": false
				}
			}`)),
			inStruct:  module.InboundServerIntent{},
			mockError: nil,
			mockVal:   module.InboundServerIntent{},
			client:    &mocks.InboundServerIntentManager{},
		}),
		Entry("fails due to entry already exists", testCase{
			expectedCode: http.StatusConflict,
			inputReader: bytes.NewBuffer([]byte(`{
				"metadata": {
					"name": "testinboundserverintent",
					"description": "test inbound server group intent",
					"userData1": "some user data 1",
					"userData2": "some user data 2"
				},
				"spec": {
					"appName":	   "test-app",
					"appLabel":	   "test-applabel",
					"serviceName":	   "test-servicename",
					"externalName":    "test-externalname",
					"port":		   6666,
					"protocol":	   "TCP",
					"externalSupport": false
				}
			}`)),
			inStruct: module.InboundServerIntent{
				Metadata: module.Metadata{
					Name:	     "testinboundserverintent",
					Description: "test inbound server group intent",
					UserData1:   "some user data 1",
					UserData2:   "some user data 2",
				},
				Spec: module.InbondServerIntentSpec{
					AppName:	 "test-app",
					AppLabel:	 "test-applabel",
					ServiceName:	 "test-servicename",
					ExternalName:	 "test-externalname",
					Port:		 6666,
					Protocol:	 "TCP",
					ExternalSupport: false,
				},
			},
			mockError: pkgerrors.New("InboundServerIntent already exists"),
			mockVal: module.InboundServerIntent{},
			client: &mocks.InboundServerIntentManager{},
		}),
	       Entry("fails due to traffic group does not exist error", testCase{
			expectedCode: http.StatusNotFound,
			inputReader: bytes.NewBuffer([]byte(`{
				"metadata": {
					"name": "testinboundserverintent",
					"description": "test inbound server group intent",
					"userData1": "some user data 1",
					"userData2": "some user data 2"
				},
				"spec": {
					"appName":	   "test-app",
					"appLabel":	   "test-applabel",
					"serviceName":	   "test-servicename",
					"externalName":    "test-externalname",
					"port":		   6666,
					"protocol":	   "TCP",
					"externalSupport": false
				}
			}`)),
			inStruct: module.InboundServerIntent{
				Metadata: module.Metadata{
					Name:	     "testinboundserverintent",
					Description: "test inbound server group intent",
					UserData1:   "some user data 1",
					UserData2:   "some user data 2",
				},
				Spec: module.InbondServerIntentSpec{
					AppName:	 "test-app",
					AppLabel:	 "test-applabel",
					ServiceName:	 "test-servicename",
					ExternalName:	 "test-externalname",
					Port:		 6666,
					Protocol:	 "TCP",
					ExternalSupport: false,
				},
			},
			mockError: pkgerrors.New("does not exist"),
			mockVal: module.InboundServerIntent{},
			client: &mocks.InboundServerIntentManager{},
		}),
		Entry("fails due to db error", testCase{
			expectedCode: http.StatusInternalServerError,
			inputReader: bytes.NewBuffer([]byte(`{
				"metadata": {
					"name": "testinboundserverintent",
					"description": "test inbound server group intent",
					"userData1": "some user data 1",
					"userData2": "some user data 2"
				},
				"spec": {
					"appName":	   "test-app",
					"appLabel":	   "test-applabel",
					"serviceName":	   "test-servicename",
					"externalName":    "test-externalname",
					"port":		   6666,
					"protocol":	   "TCP",
					"externalSupport": false
				}
			}`)),
			inStruct: module.InboundServerIntent{
				Metadata: module.Metadata{
					Name:	     "testinboundserverintent",
					Description: "test inbound server group intent",
					UserData1:   "some user data 1",
					UserData2:   "some user data 2",
				},
				Spec: module.InbondServerIntentSpec{
					AppName:	 "test-app",
					AppLabel:	 "test-applabel",
					ServiceName:	 "test-servicename",
					ExternalName:	 "test-externalname",
					Port:		 6666,
					Protocol:	 "TCP",
					ExternalSupport: false,
				},
			},
			mockError: pkgerrors.New("Creating DB Entry"),
			mockVal: module.InboundServerIntent{},
			client: &mocks.InboundServerIntentManager{},
		}),
	)
	DescribeTable("Put InboundServerIntent tests",
		func(t testCase) {
			// set up client mock responses

			t.client.On("CreateServerInboundIntent", t.inStruct, "test-project", "test-compositeapp", "v1", "test-dig", "testtrafficgroupintent", true).Return(t.mockVal, t.mockError)

			// make HTTP request
			request := httptest.NewRequest("PUT", "/v2/projects/test-project/composite-apps/test-compositeapp/v1/deployment-intent-groups/test-dig/traffic-group-intents/testtrafficgroupintent/inbound-intents/"+t.inputName, t.inputReader)
			resp := executeRequest(request, NewRouter(t.client))

			//Check returned code
			Expect(resp.StatusCode).To(Equal(t.expectedCode))

			//Check returned body
			got := module.InboundServerIntent{}
			json.NewDecoder(resp.Body).Decode(&got)
			Expect(got).To(Equal(t.mockVal))
		},

		Entry("successful put", testCase{
			inputName:    "testinboundserverintent",
			expectedCode: http.StatusCreated,
			inputReader: bytes.NewBuffer([]byte(`{
				"metadata": {
					"name": "testinboundserverintent",
					"description": "test inbound server group intent",
					"userData1": "some user data 1",
					"userData2": "some user data 2"
				},
				"spec": {
					"appName":	   "test-app",
					"appLabel":	   "test-applabel",
					"serviceName":	   "test-servicename",
					"externalName":    "test-externalname",
					"port":		   6666,
					"protocol":	   "TCP",
					"externalSupport": false
				}
			}`)),
			inStruct: module.InboundServerIntent{
				Metadata: module.Metadata{
					Name:	     "testinboundserverintent",
					Description: "test inbound server group intent",
					UserData1:   "some user data 1",
					UserData2:   "some user data 2",
				},
				Spec: module.InbondServerIntentSpec{
					AppName:	 "test-app",
					AppLabel:	 "test-applabel",
					ServiceName:	 "test-servicename",
					ExternalName:	 "test-externalname",
					Port:		 6666,
					Protocol:	 "TCP",
					ExternalSupport: false,
				},
			},
			mockError: nil,
			mockVal: module.InboundServerIntent{
				Metadata: module.Metadata{
					Name:	     "testinboundserverintent",
					Description: "test inbound server group intent",
					UserData1:   "some user data 1",
					UserData2:   "some user data 2",
				},
				Spec: module.InbondServerIntentSpec{
					AppName:	 "test-app",
					AppLabel:	 "test-applabel",
					ServiceName:	 "test-servicename",
					ExternalName:	 "test-externalname",
					Port:		 6666,
					Protocol:	 "TCP",
					ExternalSupport: false,
				},
			},
			client: &mocks.InboundServerIntentManager{},
		}),

		Entry("fails due to empty body", testCase{
			inputName:    "testinboundserverintent",
			expectedCode: http.StatusBadRequest,
			inStruct:     module.InboundServerIntent{},
			mockError:    nil,
			mockVal:      module.InboundServerIntent{},
			client:       &mocks.InboundServerIntentManager{},
		}),

		Entry("fails due to missing name", testCase{
			inputName:    "testinboundserverintent",
			expectedCode: http.StatusBadRequest,
			inStruct:     module.InboundServerIntent{},
			inputReader: bytes.NewBuffer([]byte(`{
				"metadata": {
					"description": "test inbound server group intent",
					"userData1": "some user data 1",
					"userData2": "some user data 2"
				},
				"spec": {
					"appName":	   "test-app",
					"appLabel":	   "test-applabel",
					"serviceName":	   "test-servicename",
					"externalName":    "test-externalname",
					"port":		   6666,
					"protocol":	   "TCP",
					"externalSupport": false
				}
			}`)),
			mockError:    nil,
			mockVal:      module.InboundServerIntent{},
			client:       &mocks.InboundServerIntentManager{},
		}),
		Entry("fails due to other json validation error", testCase{
			// name field has an '=' character
			inputName:    "testinboundserverintent",
			expectedCode: http.StatusBadRequest,
			inputReader: bytes.NewBuffer([]byte(`{
				"metadata": {
					"name": "test=testinboundserverintent",
					"description": "test inbound server group intent",
					"userData1": "some user data 1",
					"userData2": "some user data 2"
				},
				"spec": {
					"appName":	   "test-app",
					"appLabel":	   "test-applabel",
					"serviceName":	   "test-servicename",
					"externalName":    "test-externalname",
					"port":		   6666,
					"protocol":	   "TCP",
					"externalSupport": false
				}
			}`)),
			inStruct:  module.InboundServerIntent{},
			mockError: nil,
			mockVal:   module.InboundServerIntent{},
			client:    &mocks.InboundServerIntentManager{},
		}),
		Entry("fails due to mismatched name", testCase{
			inputName:    "testinboundserverintentXYZ",
			expectedCode: http.StatusBadRequest,
			inputReader: bytes.NewBuffer([]byte(`{
				"metadata": {
					"name": "testinboundserverintent",
					"description": "test inbound server group intent",
					"userData1": "some user data 1",
					"userData2": "some user data 2"
				},
				"spec": {
					"appName":	   "test-app",
					"appLabel":	   "test-applabel",
					"serviceName":	   "test-servicename",
					"externalName":    "test-externalname",
					"port":		   6666,
					"protocol":	   "TCP",
					"externalSupport": false
				}
			}`)),
			inStruct: module.InboundServerIntent{
				Metadata: module.Metadata{
					Name:	     "testinboundserverintent",
					Description: "test inbound server group intent",
					UserData1:   "some user data 1",
					UserData2:   "some user data 2",
				},
				Spec: module.InbondServerIntentSpec{
					AppName:	 "test-app",
					AppLabel:	 "test-applabel",
					ServiceName:	 "test-servicename",
					ExternalName:	 "test-externalname",
					Port:		 6666,
					Protocol:	 "TCP",
					ExternalSupport: false,
				},
			},
			mockError: pkgerrors.New("InboundServerIntent already exists"),
			mockVal: module.InboundServerIntent{},
			client: &mocks.InboundServerIntentManager{},
		}),
	       Entry("fails due to traffic group does not exist error", testCase{
			inputName:    "testinboundserverintent",
			expectedCode: http.StatusNotFound,
			inputReader: bytes.NewBuffer([]byte(`{
				"metadata": {
					"name": "testinboundserverintent",
					"description": "test inbound server group intent",
					"userData1": "some user data 1",
					"userData2": "some user data 2"
				},
				"spec": {
					"appName":	   "test-app",
					"appLabel":	   "test-applabel",
					"serviceName":	   "test-servicename",
					"externalName":    "test-externalname",
					"port":		   6666,
					"protocol":	   "TCP",
					"externalSupport": false
				}
			}`)),
			inStruct: module.InboundServerIntent{
				Metadata: module.Metadata{
					Name:	     "testinboundserverintent",
					Description: "test inbound server group intent",
					UserData1:   "some user data 1",
					UserData2:   "some user data 2",
				},
				Spec: module.InbondServerIntentSpec{
					AppName:	 "test-app",
					AppLabel:	 "test-applabel",
					ServiceName:	 "test-servicename",
					ExternalName:	 "test-externalname",
					Port:		 6666,
					Protocol:	 "TCP",
					ExternalSupport: false,
				},
			},
			mockError: pkgerrors.New("does not exist"),
			mockVal: module.InboundServerIntent{},
			client: &mocks.InboundServerIntentManager{},
		}),
		Entry("fails due to db error", testCase{
			inputName:    "testinboundserverintent",
			expectedCode: http.StatusInternalServerError,
			inputReader: bytes.NewBuffer([]byte(`{
				"metadata": {
					"name": "testinboundserverintent",
					"description": "test inbound server group intent",
					"userData1": "some user data 1",
					"userData2": "some user data 2"
				},
				"spec": {
					"appName":	   "test-app",
					"appLabel":	   "test-applabel",
					"serviceName":	   "test-servicename",
					"externalName":    "test-externalname",
					"port":		   6666,
					"protocol":	   "TCP",
					"externalSupport": false
				}
			}`)),
			inStruct: module.InboundServerIntent{
				Metadata: module.Metadata{
					Name:	     "testinboundserverintent",
					Description: "test inbound server group intent",
					UserData1:   "some user data 1",
					UserData2:   "some user data 2",
				},
				Spec: module.InbondServerIntentSpec{
					AppName:	 "test-app",
					AppLabel:	 "test-applabel",
					ServiceName:	 "test-servicename",
					ExternalName:	 "test-externalname",
					Port:		 6666,
					Protocol:	 "TCP",
					ExternalSupport: false,
				},
			},
			mockError: pkgerrors.New("Creating DB Entry"),
			mockVal: module.InboundServerIntent{},
			client: &mocks.InboundServerIntentManager{},
		}),
	)
	DescribeTable("Get List CreateServerInboundIntent tests",
		func(t testCase) {
			// set up client mock responses

			t.client.On("GetServerInboundIntents", "test-project", "test-compositeapp", "v1", "test-dig", "testtrafficgroupintent").Return(t.mockVals, t.mockError)

			// make HTTP request
			request := httptest.NewRequest("GET", "/v2/projects/test-project/composite-apps/test-compositeapp/v1/deployment-intent-groups/test-dig/traffic-group-intents/testtrafficgroupintent/inbound-intents", nil)
			resp := executeRequest(request, NewRouter(t.client))

			//Check returned code
			Expect(resp.StatusCode).To(Equal(t.expectedCode))

			//Check returned body
			got := []module.InboundServerIntent{}
			json.NewDecoder(resp.Body).Decode(&got)
			Expect(got).To(Equal(t.mockVals))
		},
		Entry("successful get", testCase{
			expectedCode: http.StatusOK,
			mockError: nil,
			mockVals: []module.InboundServerIntent{
				{
					Metadata: module.Metadata{
						Name:	     "testinboundserverintent1",
						Description: "test inbound server group intent",
						UserData1:   "some user data 1",
						UserData2:   "some user data 2",
					},
					Spec: module.InbondServerIntentSpec{
						AppName:	 "test-app1",
						AppLabel:	 "test-applabel1",
						ServiceName:	 "test-servicename1",
						ExternalName:	 "test-externalname1",
						Port:		 6666,
						Protocol:	 "TCP",
						ExternalSupport: false,
					},
				},
				{
					Metadata: module.Metadata{
						Name:	     "testinboundserverintent",
						Description: "test inbound server group intent",
						UserData1:   "some user data 1",
						UserData2:   "some user data 2",
					},
					Spec: module.InbondServerIntentSpec{
						AppName:	 "test-app2",
						AppLabel:	 "test-applabel2",
						ServiceName:	 "test-servicename2",
						ExternalName:	 "test-externalname2",
						Port:		 3333,
						Protocol:	 "TCP",
						ExternalSupport: false,
					},
				},
			},
			client: &mocks.InboundServerIntentManager{},
		}),
		Entry("fails due to not found", testCase{
			expectedCode: http.StatusNotFound,
			mockError: pkgerrors.New("db Find error"),
			mockVals:  []module.InboundServerIntent{},
			client:    &mocks.InboundServerIntentManager{},
		}),
		Entry("fails due to some other backend error", testCase{
			expectedCode: http.StatusInternalServerError,
			mockError: pkgerrors.New("backend error"),
			mockVals:  []module.InboundServerIntent{},
			client:    &mocks.InboundServerIntentManager{},
		}),
	)
	DescribeTable("Get CreateServerInboundIntent tests",
		func(t testCase) {
			// set up client mock responses

			t.client.On("GetServerInboundIntent", t.inputName, "test-project", "test-compositeapp", "v1", "test-dig", "testtrafficgroupintent").Return(t.mockVal, t.mockError)

			// make HTTP request
			request := httptest.NewRequest("GET", "/v2/projects/test-project/composite-apps/test-compositeapp/v1/deployment-intent-groups/test-dig/traffic-group-intents/testtrafficgroupintent/inbound-intents/"+t.inputName, nil)
			resp := executeRequest(request, NewRouter(t.client))

			//Check returned code
			Expect(resp.StatusCode).To(Equal(t.expectedCode))

			//Check returned body
			got := module.InboundServerIntent{}
			json.NewDecoder(resp.Body).Decode(&got)
			Expect(got).To(Equal(t.mockVal))
		},
		Entry("successful get", testCase{
			inputName:    "testinboundserverintent",
			expectedCode: http.StatusOK,
			mockError: nil,
			mockVal: module.InboundServerIntent{
				Metadata: module.Metadata{
					Name:	     "testinboundserverintent",
					Description: "test inbound server group intent",
					UserData1:   "some user data 1",
					UserData2:   "some user data 2",
				},
				Spec: module.InbondServerIntentSpec{
					AppName:	 "test-app",
					AppLabel:	 "test-applabel",
					ServiceName:	 "test-servicename",
					ExternalName:	 "test-externalname",
					Port:		 6666,
					Protocol:	 "TCP",
					ExternalSupport: false,
				},
			},
			client: &mocks.InboundServerIntentManager{},
		}),
		Entry("fails due to not found", testCase{
			inputName:    "testinboundserverintent",
			expectedCode: http.StatusNotFound,
			mockError: pkgerrors.New("db Find error"),
			mockVal:  module.InboundServerIntent{},
			client:    &mocks.InboundServerIntentManager{},
		}),
		Entry("fails due to some other backend error", testCase{
			inputName:    "testinboundserverintent",
			expectedCode: http.StatusInternalServerError,
			mockError: pkgerrors.New("backend error"),
			mockVal:  module.InboundServerIntent{},
			client:    &mocks.InboundServerIntentManager{},
		}),
	)

	DescribeTable("DELETE CreateServerInboundIntent tests",
		func(t testCase) {
			// set up client mock responses

			t.client.On("DeleteServerInboundIntent", t.inputName, "test-project", "test-compositeapp", "v1", "test-dig", "testtrafficgroupintent").Return(t.mockError)

			// make HTTP request
			request := httptest.NewRequest("DELETE", "/v2/projects/test-project/composite-apps/test-compositeapp/v1/deployment-intent-groups/test-dig/traffic-group-intents/testtrafficgroupintent/inbound-intents/"+t.inputName, nil)
			resp := executeRequest(request, NewRouter(t.client))

			//Check returned code
			Expect(resp.StatusCode).To(Equal(t.expectedCode))

			//Check returned body
			got := module.InboundServerIntent{}
			json.NewDecoder(resp.Body).Decode(&got)
			Expect(got).To(Equal(t.mockVal))
		},
		Entry("successful delete", testCase{
			inputName:    "testinboundserverintent",
			expectedCode: http.StatusNoContent,
			mockError:    nil,
			mockVal:      module.InboundServerIntent{},
			client:       &mocks.InboundServerIntentManager{},
		}),
		Entry("fails due to not found", testCase{
			inputName:    "testinboundserverintent",
			expectedCode: http.StatusNotFound,
			mockError:    pkgerrors.New("db Remove error - not found"),
			mockVal:      module.InboundServerIntent{},
			client:       &mocks.InboundServerIntentManager{},
		}),
		Entry("fails due to conflict", testCase{
			inputName:    "testinboundserverintent",
			expectedCode: http.StatusConflict,
			mockError:    pkgerrors.New("db Remove error - conflict"),
			mockVal:      module.InboundServerIntent{},
			client:       &mocks.InboundServerIntentManager{},
		}),
		Entry("fails due to some other backend error", testCase{
			inputName:    "testinboundserverintent",
			expectedCode: http.StatusInternalServerError,
			mockError:    pkgerrors.New("db Remove error - general"),
			mockVal:      module.InboundServerIntent{},
			client:       &mocks.InboundServerIntentManager{},
		}),
	)
})
