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
	inClientsIntJSONFile = "../json-schemas/inbound-clients.json"
}

var _ = Describe("Inboundclientsintenthandler", func() {

	var (
		mdb *db.MockDB
	)
	type testCase struct {
		inputName    string
		inputReader  io.Reader
		inStruct     module.InboundClientsIntent
		mockError    error
		mockVal      module.InboundClientsIntent
		mockVals     []module.InboundClientsIntent
		expectedCode int
		client	     *mocks.InboundClientsIntentManager
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

	DescribeTable("Create InboundClientsIntent tests",
		func(t testCase) {
			// set up client mock responses

			t.client.On("CreateClientsInboundIntent", t.inStruct, "test-project", "test-compositeapp", "v1", "test-dig", "testtrafficgroupintent", "testinboundintentname", false).Return(t.mockVal, t.mockError)

			// make HTTP request
			request := httptest.NewRequest("POST", "/v2/projects/test-project/composite-apps/test-compositeapp/v1/deployment-intent-groups/test-dig/traffic-group-intents/testtrafficgroupintent/inbound-intents/testinboundintentname/clients", t.inputReader)
			resp := executeRequest(request, NewRouter(t.client))

			//Check returned code
			Expect(resp.StatusCode).To(Equal(t.expectedCode))

			//Check returned body
			got := module.InboundClientsIntent{}
			json.NewDecoder(resp.Body).Decode(&got)
			Expect(got).To(Equal(t.mockVal))
		},

		Entry("successful create", testCase{
			expectedCode: http.StatusCreated,
			inputReader: bytes.NewBuffer([]byte(`{
				"metadata": {
					"name": "testinboundclientsintent",
					"description": "test inbound clients group intent",
					"userData1": "some user data 1",
					"userData2": "some user data 2"
				},
				"spec": {
					"appName":	   "test-app",
					"appLabel":	   "test-applabel",
					"serviceName":	   "test-servicename",
					"namespaces":	   [],
					"cidrs":	   []
				}
			}`)),
			inStruct: module.InboundClientsIntent{
				Metadata: module.Metadata{
					Name:	     "testinboundclientsintent",
					Description: "test inbound clients group intent",
					UserData1:   "some user data 1",
					UserData2:   "some user data 2",
				},
				Spec: module.InboundClientsIntentSpec{
					AppName:      "test-app",
					AppLabel:     "test-applabel",
					ServiceName:  "test-servicename",
					Namespaces:   []string{},
					IpRange:      []string{},
				},
			},
			mockError: nil,
			mockVal: module.InboundClientsIntent{
				Metadata: module.Metadata{
					Name:	     "testinboundclientsintent",
					Description: "test inbound clients group intent",
					UserData1:   "some user data 1",
					UserData2:   "some user data 2",
				},
				Spec: module.InboundClientsIntentSpec{
					AppName:      "test-app",
					AppLabel:     "test-applabel",
					ServiceName:  "test-servicename",
					Namespaces:   []string{},
					IpRange:      []string{},
				},
			},
			client: &mocks.InboundClientsIntentManager{},
		}),

		Entry("fails due to empty body", testCase{
			expectedCode: http.StatusBadRequest,
			inStruct:     module.InboundClientsIntent{},
			mockError:    nil,
			mockVal:      module.InboundClientsIntent{},
			client:       &mocks.InboundClientsIntentManager{},
		}),
		Entry("fails due to missing name", testCase{
			expectedCode: http.StatusBadRequest,
			inStruct:     module.InboundClientsIntent{},
			inputReader: bytes.NewBuffer([]byte(`{
				"metadata": {
					"description": "test inbound clients group intent",
					"userData1": "some user data 1",
					"userData2": "some user data 2"
				},
				"spec": {
					"appName":	   "test-app",
					"appLabel":	   "test-applabel",
					"serviceName":	   "test-servicename",
					"namespaces":	   [],
					"cidrs":	   []
				}
			}`)),
			mockError:    nil,
			mockVal:      module.InboundClientsIntent{},
			client:       &mocks.InboundClientsIntentManager{},
		}),
		Entry("fails due to missing app name", testCase{
			expectedCode: http.StatusBadRequest,
			inStruct:     module.InboundClientsIntent{},
			inputReader: bytes.NewBuffer([]byte(`{
				"metadata": {
					"name": "testinboundclientsintent",
					"description": "test inbound clients group intent",
					"userData1": "some user data 1",
					"userData2": "some user data 2"
				},
				"spec": {
					"appLabel":	   "test-applabel",
					"serviceName":	   "test-servicename",
					"namespaces":	   [],
					"cidrs":	   []
				}
			}`)),
			mockError:    nil,
			mockVal:      module.InboundClientsIntent{},
			client:       &mocks.InboundClientsIntentManager{},
		}),
		Entry("fails due to missing app label", testCase{
			expectedCode: http.StatusBadRequest,
			inStruct:     module.InboundClientsIntent{},
			inputReader: bytes.NewBuffer([]byte(`{
				"metadata": {
					"name": "testinboundclientsintent",
					"description": "test inbound clients group intent",
					"userData1": "some user data 1",
					"userData2": "some user data 2"
				},
				"spec": {
					"appName":	   "test-app",
					"serviceName":	   "test-servicename",
					"namespaces":	   [],
					"cidrs":	   []
				}
			}`)),
			mockError:    nil,
			mockVal:      module.InboundClientsIntent{},
			client:       &mocks.InboundClientsIntentManager{},
		}),
		Entry("fails due to missing namespaces", testCase{
			expectedCode: http.StatusBadRequest,
			inStruct:     module.InboundClientsIntent{},
			inputReader: bytes.NewBuffer([]byte(`{
				"metadata": {
					"name": "testinboundclientsintent",
					"description": "test inbound clients group intent",
					"userData1": "some user data 1",
					"userData2": "some user data 2"
				},
				"spec": {
					"appName":	   "test-app",
					"appLabel":	   "test-applabel",
					"serviceName":	   "test-servicename",
					"cidrs":	   []
				}
			}`)),
			mockError:    nil,
			mockVal:      module.InboundClientsIntent{},
			client:       &mocks.InboundClientsIntentManager{},
		}),
		Entry("fails due to other json validation error", testCase{
			// name field has an '=' character
			expectedCode: http.StatusBadRequest,
			inputReader: bytes.NewBuffer([]byte(`{
				"metadata": {
					"name": "test=testinboundclientsintent",
					"description": "test inbound clients group intent",
					"userData1": "some user data 1",
					"userData2": "some user data 2"
				},
				"spec": {
					"appName":	   "test-app",
					"appLabel":	   "test-applabel",
					"serviceName":	   "test-servicename",
					"namespaces":	   [],
					"cidrs":	   []
				}
			}`)),
			inStruct:  module.InboundClientsIntent{},
			mockError: nil,
			mockVal:   module.InboundClientsIntent{},
			client:    &mocks.InboundClientsIntentManager{},
		}),
		Entry("fails due to entry already exists", testCase{
			expectedCode: http.StatusConflict,
			inputReader: bytes.NewBuffer([]byte(`{
				"metadata": {
					"name": "testinboundclientsintent",
					"description": "test inbound clients group intent",
					"userData1": "some user data 1",
					"userData2": "some user data 2"
				},
				"spec": {
					"appName":	   "test-app",
					"appLabel":	   "test-applabel",
					"serviceName":	   "test-servicename",
					"namespaces":	   [],
					"cidrs":	   []
				}
			}`)),
			inStruct: module.InboundClientsIntent{
				Metadata: module.Metadata{
					Name:	     "testinboundclientsintent",
					Description: "test inbound clients group intent",
					UserData1:   "some user data 1",
					UserData2:   "some user data 2",
				},
				Spec: module.InboundClientsIntentSpec{
					AppName:      "test-app",
					AppLabel:     "test-applabel",
					ServiceName:  "test-servicename",
					Namespaces:   []string{},
					IpRange:      []string{},
				},
			},
			mockError: pkgerrors.New("InboundServerIntent already exists"),
			mockVal: module.InboundClientsIntent{},
			client: &mocks.InboundClientsIntentManager{},
		}),
	       Entry("fails due to traffic group does not exist error", testCase{
			expectedCode: http.StatusNotFound,
			inputReader: bytes.NewBuffer([]byte(`{
				"metadata": {
					"name": "testinboundclientsintent",
					"description": "test inbound clients group intent",
					"userData1": "some user data 1",
					"userData2": "some user data 2"
				},
				"spec": {
					"appName":	   "test-app",
					"appLabel":	   "test-applabel",
					"serviceName":	   "test-servicename",
					"namespaces":	   [],
					"cidrs":	   []
				}
			}`)),
			inStruct: module.InboundClientsIntent{
				Metadata: module.Metadata{
					Name:	     "testinboundclientsintent",
					Description: "test inbound clients group intent",
					UserData1:   "some user data 1",
					UserData2:   "some user data 2",
				},
				Spec: module.InboundClientsIntentSpec{
					AppName:      "test-app",
					AppLabel:     "test-applabel",
					ServiceName:  "test-servicename",
					Namespaces:   []string{},
					IpRange:      []string{},
				},
			},
			mockError: pkgerrors.New("does not exist"),
			mockVal: module.InboundClientsIntent{},
			client: &mocks.InboundClientsIntentManager{},
		}),
		Entry("fails due to db error", testCase{
			expectedCode: http.StatusInternalServerError,
			inputReader: bytes.NewBuffer([]byte(`{
				"metadata": {
					"name": "testinboundclientsintent",
					"description": "test inbound clients group intent",
					"userData1": "some user data 1",
					"userData2": "some user data 2"
				},
				"spec": {
					"appName":	   "test-app",
					"appLabel":	   "test-applabel",
					"serviceName":	   "test-servicename",
					"namespaces":	   [],
					"cidrs":	   []
				}
			}`)),
			inStruct: module.InboundClientsIntent{
				Metadata: module.Metadata{
					Name:	     "testinboundclientsintent",
					Description: "test inbound clients group intent",
					UserData1:   "some user data 1",
					UserData2:   "some user data 2",
				},
				Spec: module.InboundClientsIntentSpec{
					AppName:      "test-app",
					AppLabel:     "test-applabel",
					ServiceName:  "test-servicename",
					Namespaces:   []string{},
					IpRange:      []string{},
				},
			},
			mockError: pkgerrors.New("Creating DB Entry"),
			mockVal: module.InboundClientsIntent{},
			client: &mocks.InboundClientsIntentManager{},
		}),
	)
	DescribeTable("Put InboundClientsIntent tests",
		func(t testCase) {
			// set up client mock responses

			t.client.On("CreateClientsInboundIntent", t.inStruct, "test-project", "test-compositeapp", "v1", "test-dig", "testtrafficgroupintent", "testinboundintentname", true).Return(t.mockVal, t.mockError)

			// make HTTP request
			request := httptest.NewRequest("PUT", "/v2/projects/test-project/composite-apps/test-compositeapp/v1/deployment-intent-groups/test-dig/traffic-group-intents/testtrafficgroupintent/inbound-intents/testinboundintentname/clients/"+t.inputName, t.inputReader)
			resp := executeRequest(request, NewRouter(t.client))

			//Check returned code
			Expect(resp.StatusCode).To(Equal(t.expectedCode))

			//Check returned body
			got := module.InboundClientsIntent{}
			json.NewDecoder(resp.Body).Decode(&got)
			Expect(got).To(Equal(t.mockVal))
		},

		Entry("successful put", testCase{
			inputName: "testinboundclientsintent",
			expectedCode: http.StatusCreated,
			inputReader: bytes.NewBuffer([]byte(`{
				"metadata": {
					"name": "testinboundclientsintent",
					"description": "test inbound clients group intent",
					"userData1": "some user data 1",
					"userData2": "some user data 2"
				},
				"spec": {
					"appName":	   "test-app",
					"appLabel":	   "test-applabel",
					"serviceName":	   "test-servicename",
					"namespaces":	   [],
					"cidrs":	   []
				}
			}`)),
			inStruct: module.InboundClientsIntent{
				Metadata: module.Metadata{
					Name:	     "testinboundclientsintent",
					Description: "test inbound clients group intent",
					UserData1:   "some user data 1",
					UserData2:   "some user data 2",
				},
				Spec: module.InboundClientsIntentSpec{
					AppName:      "test-app",
					AppLabel:     "test-applabel",
					ServiceName:  "test-servicename",
					Namespaces:   []string{},
					IpRange:      []string{},
				},
			},
			mockError: nil,
			mockVal: module.InboundClientsIntent{
				Metadata: module.Metadata{
					Name:	     "testinboundclientsintent",
					Description: "test inbound clients group intent",
					UserData1:   "some user data 1",
					UserData2:   "some user data 2",
				},
				Spec: module.InboundClientsIntentSpec{
					AppName:      "test-app",
					AppLabel:     "test-applabel",
					ServiceName:  "test-servicename",
					Namespaces:   []string{},
					IpRange:      []string{},
				},
			},
			client: &mocks.InboundClientsIntentManager{},
		}),

		Entry("fails due to empty body", testCase{
			inputName: "testinboundclientsintent",
			expectedCode: http.StatusBadRequest,
			inStruct:     module.InboundClientsIntent{},
			mockError:    nil,
			mockVal:      module.InboundClientsIntent{},
			client:       &mocks.InboundClientsIntentManager{},
		}),
		Entry("fails due to missing name", testCase{
			inputName: "testinboundclientsintent",
			expectedCode: http.StatusBadRequest,
			inStruct:     module.InboundClientsIntent{},
			inputReader: bytes.NewBuffer([]byte(`{
				"metadata": {
					"description": "test inbound clients group intent",
					"userData1": "some user data 1",
					"userData2": "some user data 2"
				},
				"spec": {
					"appName":	   "test-app",
					"appLabel":	   "test-applabel",
					"serviceName":	   "test-servicename",
					"namespaces":	   [],
					"cidrs":	   []
				}
			}`)),
			mockError:    nil,
			mockVal:      module.InboundClientsIntent{},
			client:       &mocks.InboundClientsIntentManager{},
		}),
		Entry("fails due to other json validation error", testCase{
			inputName: "testinboundclientsintent",
			// name field has an '=' character
			expectedCode: http.StatusBadRequest,
			inputReader: bytes.NewBuffer([]byte(`{
				"metadata": {
					"name": "test=testinboundclientsintent",
					"description": "test inbound clients group intent",
					"userData1": "some user data 1",
					"userData2": "some user data 2"
				},
				"spec": {
					"appName":	   "test-app",
					"appLabel":	   "test-applabel",
					"serviceName":	   "test-servicename",
					"namespaces":	   [],
					"cidrs":	   []
				}
			}`)),
			inStruct:  module.InboundClientsIntent{},
			mockError: nil,
			mockVal:   module.InboundClientsIntent{},
			client:    &mocks.InboundClientsIntentManager{},
		}),
		Entry("fails due to mismatched name", testCase{
			inputName: "testinboundclientsintentXYZ",
			expectedCode: http.StatusBadRequest,
			inputReader: bytes.NewBuffer([]byte(`{
				"metadata": {
					"name": "testinboundclientsintent",
					"description": "test inbound clients group intent",
					"userData1": "some user data 1",
					"userData2": "some user data 2"
				},
				"spec": {
					"appName":	   "test-app",
					"appLabel":	   "test-applabel",
					"serviceName":	   "test-servicename",
					"namespaces":	   [],
					"cidrs":	   []
				}
			}`)),
			inStruct: module.InboundClientsIntent{
				Metadata: module.Metadata{
					Name:	     "testinboundclientsintent",
					Description: "test inbound clients group intent",
					UserData1:   "some user data 1",
					UserData2:   "some user data 2",
				},
				Spec: module.InboundClientsIntentSpec{
					AppName:      "test-app",
					AppLabel:     "test-applabel",
					ServiceName:  "test-servicename",
					Namespaces:   []string{},
					IpRange:      []string{},
				},
			},
			mockError: pkgerrors.New("Creating DB Entry"),
			mockVal: module.InboundClientsIntent{},
			client: &mocks.InboundClientsIntentManager{},
		}),
	       Entry("fails due to traffic group does not exist error", testCase{
			inputName: "testinboundclientsintent",
			expectedCode: http.StatusNotFound,
			inputReader: bytes.NewBuffer([]byte(`{
				"metadata": {
					"name": "testinboundclientsintent",
					"description": "test inbound clients group intent",
					"userData1": "some user data 1",
					"userData2": "some user data 2"
				},
				"spec": {
					"appName":	   "test-app",
					"appLabel":	   "test-applabel",
					"serviceName":	   "test-servicename",
					"namespaces":	   [],
					"cidrs":	   []
				}
			}`)),
			inStruct: module.InboundClientsIntent{
				Metadata: module.Metadata{
					Name:	     "testinboundclientsintent",
					Description: "test inbound clients group intent",
					UserData1:   "some user data 1",
					UserData2:   "some user data 2",
				},
				Spec: module.InboundClientsIntentSpec{
					AppName:      "test-app",
					AppLabel:     "test-applabel",
					ServiceName:  "test-servicename",
					Namespaces:   []string{},
					IpRange:      []string{},
				},
			},
			mockError: pkgerrors.New("does not exist"),
			mockVal: module.InboundClientsIntent{},
			client: &mocks.InboundClientsIntentManager{},
		}),
		Entry("fails due to db error", testCase{
			inputName: "testinboundclientsintent",
			expectedCode: http.StatusInternalServerError,
			inputReader: bytes.NewBuffer([]byte(`{
				"metadata": {
					"name": "testinboundclientsintent",
					"description": "test inbound clients group intent",
					"userData1": "some user data 1",
					"userData2": "some user data 2"
				},
				"spec": {
					"appName":	   "test-app",
					"appLabel":	   "test-applabel",
					"serviceName":	   "test-servicename",
					"namespaces":	   [],
					"cidrs":	   []
				}
			}`)),
			inStruct: module.InboundClientsIntent{
				Metadata: module.Metadata{
					Name:	     "testinboundclientsintent",
					Description: "test inbound clients group intent",
					UserData1:   "some user data 1",
					UserData2:   "some user data 2",
				},
				Spec: module.InboundClientsIntentSpec{
					AppName:      "test-app",
					AppLabel:     "test-applabel",
					ServiceName:  "test-servicename",
					Namespaces:   []string{},
					IpRange:      []string{},
				},
			},
			mockError: pkgerrors.New("Creating DB Entry"),
			mockVal: module.InboundClientsIntent{},
			client: &mocks.InboundClientsIntentManager{},
		}),
	)
	DescribeTable("Get List GetClientsInboundIntent tests",
		func(t testCase) {
			// set up client mock responses

			t.client.On("GetClientsInboundIntents", "test-project", "test-compositeapp", "v1", "test-dig", "testtrafficgroupintent", "testinboundintentname").Return(t.mockVals, t.mockError)

			// make HTTP request
			request := httptest.NewRequest("GET", "/v2/projects/test-project/composite-apps/test-compositeapp/v1/deployment-intent-groups/test-dig/traffic-group-intents/testtrafficgroupintent/inbound-intents/testinboundintentname/clients", nil)
			resp := executeRequest(request, NewRouter(t.client))

			//Check returned code
			Expect(resp.StatusCode).To(Equal(t.expectedCode))

			//Check returned body
			got := []module.InboundClientsIntent{}
			json.NewDecoder(resp.Body).Decode(&got)
			Expect(got).To(Equal(t.mockVals))
	       },
	       Entry("successful get", testCase{
			expectedCode: http.StatusOK,
			mockError: nil,
			mockVals: []module.InboundClientsIntent{
				{
					Metadata: module.Metadata{
						Name:	     "testinboundclientsintent1",
						Description: "test inbound clients group intent",
						UserData1:   "some user data 1",
						UserData2:   "some user data 2",
					},
					Spec: module.InboundClientsIntentSpec{
						AppName:      "test-app1",
						AppLabel:     "test-applabel1",
						ServiceName:  "test-servicename1",
						Namespaces:   []string{},
						IpRange:      []string{},
					},
				},
				{
					Metadata: module.Metadata{
						Name:	     "testinboundclientsintent2",
						Description: "test inbound clients group intent",
						UserData1:   "some user data 1",
						UserData2:   "some user data 2",
					},
					Spec: module.InboundClientsIntentSpec{
						AppName:      "test-appo2",
						AppLabel:     "test-applabel2",
						ServiceName:  "test-servicename2",
						Namespaces:   []string{},
						IpRange:      []string{},
					},
				},
			},
			client: &mocks.InboundClientsIntentManager{},
		}),
		Entry("fails due to not found", testCase{
			expectedCode: http.StatusNotFound,
			mockError: pkgerrors.New("db Find error"),
			mockVals:  []module.InboundClientsIntent{},
			client:    &mocks.InboundClientsIntentManager{},
		}),
		Entry("fails due to some other backend error", testCase{
			expectedCode: http.StatusInternalServerError,
			mockError: pkgerrors.New("backend error"),
			mockVals:  []module.InboundClientsIntent{},
			client:    &mocks.InboundClientsIntentManager{},
		}),
	)
	DescribeTable("Get GetClientsInboundIntent tests",
		func(t testCase) {
			// set up client mock responses

			t.client.On("GetClientsInboundIntent", t.inputName,"test-project", "test-compositeapp", "v1", "test-dig", "testtrafficgroupintent", "testinboundintentname").Return(t.mockVal, t.mockError)

			// make HTTP request
			request := httptest.NewRequest("GET", "/v2/projects/test-project/composite-apps/test-compositeapp/v1/deployment-intent-groups/test-dig/traffic-group-intents/testtrafficgroupintent/inbound-intents/testinboundintentname/clients/"+t.inputName, nil)
			resp := executeRequest(request, NewRouter(t.client))

			//Check returned code
			Expect(resp.StatusCode).To(Equal(t.expectedCode))

			//Check returned body
			got := module.InboundClientsIntent{}
			json.NewDecoder(resp.Body).Decode(&got)
			Expect(got).To(Equal(t.mockVal))
		},
		Entry("successful get", testCase{
			inputName:    "testinboundclientsintent",
			expectedCode: http.StatusOK,
			mockError: nil,
			mockVal: module.InboundClientsIntent{
				Metadata: module.Metadata{
					Name:	     "testinboundclientsintent",
					Description: "test inbound clients group intent",
					UserData1:   "some user data 1",
					UserData2:   "some user data 2",
				},
				Spec: module.InboundClientsIntentSpec{
					AppName:      "test-app",
					AppLabel:     "test-applabel",
					ServiceName:  "test-servicename",
					Namespaces:   []string{},
					IpRange:      []string{},
				},
			},
			client: &mocks.InboundClientsIntentManager{},
		}),
		Entry("fails due to not found", testCase{
			inputName:    "testinboundclientsintent",
			expectedCode: http.StatusNotFound,
			mockError: pkgerrors.New("db Find error"),
			mockVal:  module.InboundClientsIntent{},
			client:    &mocks.InboundClientsIntentManager{},
		}),
		Entry("fails due to some other backend error", testCase{
			inputName:    "testinboundclientsintent",
			expectedCode: http.StatusInternalServerError,
			mockError: pkgerrors.New("backend error"),
			mockVal:  module.InboundClientsIntent{},
			client:    &mocks.InboundClientsIntentManager{},
		}),
	)
	DescribeTable("DELETE GetClientsInboundIntent tests",
		func(t testCase) {
			// set up client mock responses

			t.client.On("DeleteClientsInboundIntent", t.inputName,"test-project", "test-compositeapp", "v1", "test-dig", "testtrafficgroupintent", "testinboundintentname").Return(t.mockError)

			// make HTTP request
			request := httptest.NewRequest("DELETE", "/v2/projects/test-project/composite-apps/test-compositeapp/v1/deployment-intent-groups/test-dig/traffic-group-intents/testtrafficgroupintent/inbound-intents/testinboundintentname/clients/"+t.inputName, nil)
			resp := executeRequest(request, NewRouter(t.client))

			//Check returned code
			Expect(resp.StatusCode).To(Equal(t.expectedCode))

			//Check returned body
			got := module.InboundClientsIntent{}
			json.NewDecoder(resp.Body).Decode(&got)
			Expect(got).To(Equal(t.mockVal))
		},
		Entry("successful delete", testCase{
			inputName:    "testinboundclientsintent",
			expectedCode: http.StatusNoContent,
			mockError: nil,
			mockVal: module.InboundClientsIntent{},
			client: &mocks.InboundClientsIntentManager{},
		}),
		Entry("fails due to not found", testCase{
			inputName:    "testinboundclientsintent",
			expectedCode: http.StatusNotFound,
			mockError: pkgerrors.New("db Remove error - not found"),
			mockVal:  module.InboundClientsIntent{},
			client:    &mocks.InboundClientsIntentManager{},
		}),
		Entry("fails due to conflict", testCase{
			inputName:    "testinboundclientsintent",
			expectedCode: http.StatusConflict,
			mockError: pkgerrors.New("db Remove error - conflict"),
			mockVal:  module.InboundClientsIntent{},
			client:    &mocks.InboundClientsIntentManager{},
		}),
		Entry("fails due to some other backend error", testCase{
			inputName:    "testinboundclientsintent",
			expectedCode: http.StatusInternalServerError,
			mockError: pkgerrors.New("db Remove error - general"),
			mockVal:  module.InboundClientsIntent{},
			client:    &mocks.InboundClientsIntentManager{},
		}),
	)
})
