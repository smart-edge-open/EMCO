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
	TrGroupIntJSONFile = "../json-schemas/metadata.json"
}

var _ = Describe("Trafficgroupintenthandler", func() {

	var (
		mdb *db.MockDB
	)
	type testCase struct {
		inputName    string
		inputReader  io.Reader
		inStruct     module.TrafficGroupIntent
		mockError    error
		mockVal      module.TrafficGroupIntent
		mockVals     []module.TrafficGroupIntent
		expectedCode int
		client	     *mocks.TrafficGroupIntentManager
		digmockVal   orcmod.DeploymentIntentGroup
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

	DescribeTable("Create TrafficGroupIntent tests",
		func(t testCase) {
			// set up client mock responses

			t.client.On("CreateTrafficGroupIntent", t.inStruct, "test-project", "test-compositeapp", "v1", "test-dig", false).Return(t.mockVal, t.mockError)

			// make HTTP request
			request := httptest.NewRequest("POST", "/v2/projects/test-project/composite-apps/test-compositeapp/v1/deployment-intent-groups/test-dig/traffic-group-intents", t.inputReader)
			resp := executeRequest(request, NewRouter(t.client))

			//Check returned code
			Expect(resp.StatusCode).To(Equal(t.expectedCode))

			//Check returned body
			got := module.TrafficGroupIntent{}
			json.NewDecoder(resp.Body).Decode(&got)
			Expect(got).To(Equal(t.mockVal))
		},

		Entry("successful create", testCase{
			expectedCode: http.StatusCreated,
			inputReader: bytes.NewBuffer([]byte(`{
				"metadata": {
					"name": "testtrafficgroupintent",
					"description": "test traffic group intent",
					"userData1": "some user data 1",
					"userData2": "some user data 2"
				}
			}`)),
			inStruct: module.TrafficGroupIntent{
				Metadata: module.Metadata{
					Name:	     "testtrafficgroupintent",
					Description: "test traffic group intent",
					UserData1:   "some user data 1",
					UserData2:   "some user data 2",
				},
			},
			mockError: nil,
			mockVal: module.TrafficGroupIntent{
				Metadata: module.Metadata{
					Name:	     "testtrafficgroupintent",
					Description: "test traffic group intent",
					UserData1:   "some user data 1",
					UserData2:   "some user data 2",
				},
			},
			client: &mocks.TrafficGroupIntentManager{},
		}),

		Entry("fails due to empty body", testCase{
			expectedCode: http.StatusBadRequest,
			inStruct:     module.TrafficGroupIntent{},
			mockError:    nil,
			mockVal:      module.TrafficGroupIntent{},
			client:       &mocks.TrafficGroupIntentManager{},
		}),
		Entry("fails due to missing name", testCase{
			expectedCode: http.StatusBadRequest,
			inStruct:     module.TrafficGroupIntent{},
			inputReader: bytes.NewBuffer([]byte(`{
				"metadata": {
					"description": "test traffic group intent",
					"userData1": "some user data 1",
					"userData2": "some user data 2"
				}
			}`)),
			mockError:    nil,
			mockVal:      module.TrafficGroupIntent{},
			client:       &mocks.TrafficGroupIntentManager{},
		}),
		Entry("fails due to other json validation error", testCase{
			// name field has an '=' character
			expectedCode: http.StatusBadRequest,
			inputReader: bytes.NewBuffer([]byte(`{
				"metadata": {
					"name": "test=testtrafficgroupintent",
					"description": "test traffic group intent",
					"userData1": "some user data 1",
					"userData2": "some user data 2"
				}
			}`)),
			inStruct:  module.TrafficGroupIntent{},
			mockError: nil,
			mockVal:      module.TrafficGroupIntent{},
			client:    &mocks.TrafficGroupIntentManager{},
		}),
		Entry("fails due to json body decoding error", testCase{
			// extra comma at the end of the userData2 line 
			expectedCode: http.StatusUnprocessableEntity,
			inputReader: bytes.NewBuffer([]byte(`{
				"metadata": {
					"name": "testtrafficgroupintent",
					"description": "test traffic group intent",
					"userData1": "some user data 1",
					"userData2": "some user data 2",
				}
			}`)),
			inStruct:  module.TrafficGroupIntent{},
			mockError: nil,
			mockVal:      module.TrafficGroupIntent{},
			client:    &mocks.TrafficGroupIntentManager{},
		}),
		Entry("fails due to entry already exists", testCase{
			expectedCode: http.StatusConflict,
			inputReader: bytes.NewBuffer([]byte(`{
				"metadata": {
					"name": "testtrafficgroupintent",
					"description": "test traffic group intent",
					"userData1": "some user data 1",
					"userData2": "some user data 2"
				}
			}`)),
			inStruct: module.TrafficGroupIntent{
				Metadata: module.Metadata{
					Name:	     "testtrafficgroupintent",
					Description: "test traffic group intent",
					UserData1:   "some user data 1",
					UserData2:   "some user data 2",
				},
			},
			mockError: pkgerrors.New("TrafficGroupIntent already exists"),
			mockVal: module.TrafficGroupIntent{},
			client: &mocks.TrafficGroupIntentManager{},
		}),
		Entry("fails due to db error", testCase{
			expectedCode: http.StatusInternalServerError,
			inputReader: bytes.NewBuffer([]byte(`{
				"metadata": {
					"name": "testtrafficgroupintent",
					"description": "test traffic group intent",
					"userData1": "some user data 1",
					"userData2": "some user data 2"
				}
			}`)),
			inStruct: module.TrafficGroupIntent{
				Metadata: module.Metadata{
					Name:	     "testtrafficgroupintent",
					Description: "test traffic group intent",
					UserData1:   "some user data 1",
					UserData2:   "some user data 2",
				},
			},
			mockError: pkgerrors.New("Creating DB Entry"),
			mockVal: module.TrafficGroupIntent{},
			client: &mocks.TrafficGroupIntentManager{},
		}),
	)
	DescribeTable("Put TrafficGroupIntent tests",
		func(t testCase) {
			// set up client mock responses

			t.client.On("CreateTrafficGroupIntent", t.inStruct, "test-project", "test-compositeapp", "v1", "test-dig", true).Return(t.mockVal, t.mockError)

			// make HTTP request
			request := httptest.NewRequest("PUT", "/v2/projects/test-project/composite-apps/test-compositeapp/v1/deployment-intent-groups/test-dig/traffic-group-intents/"+t.inputName, t.inputReader)
			resp := executeRequest(request, NewRouter(t.client))

			//Check returned code
			Expect(resp.StatusCode).To(Equal(t.expectedCode))

			//Check returned body
			got := module.TrafficGroupIntent{}
			json.NewDecoder(resp.Body).Decode(&got)
			Expect(got).To(Equal(t.mockVal))
		},
		Entry("successful create", testCase{
			expectedCode: http.StatusCreated,
			inputName:    "testtrafficgroupintent",
			inputReader: bytes.NewBuffer([]byte(`{
				"metadata": {
					"name": "testtrafficgroupintent",
					"description": "test traffic group intent",
					"userData1": "some user data 1",
					"userData2": "some user data 2"
				}
			}`)),
			inStruct: module.TrafficGroupIntent{
				Metadata: module.Metadata{
					Name:	     "testtrafficgroupintent",
					Description: "test traffic group intent",
					UserData1:   "some user data 1",
					UserData2:   "some user data 2",
				},
			},
			mockError: nil,
			mockVal: module.TrafficGroupIntent{
				Metadata: module.Metadata{
					Name:	     "testtrafficgroupintent",
					Description: "test traffic group intent",
					UserData1:   "some user data 1",
					UserData2:   "some user data 2",
				},
			},
			client: &mocks.TrafficGroupIntentManager{},
		}),
		Entry("fails due to empty body", testCase{
			inputName:    "testtrafficgroupintent",
			expectedCode: http.StatusBadRequest,
			inStruct:     module.TrafficGroupIntent{},
			mockError:    nil,
			mockVal:      module.TrafficGroupIntent{},
			client:       &mocks.TrafficGroupIntentManager{},
		}),
		Entry("fails due to missing name", testCase{
			inputName:    "testtrafficgroupintent",
			expectedCode: http.StatusBadRequest,
			inStruct:     module.TrafficGroupIntent{},
			inputReader: bytes.NewBuffer([]byte(`{
				"metadata": {
					"description": "test traffic group intent",
					"userData1": "some user data 1",
					"userData2": "some user data 2"
				}
			}`)),
			mockError:    nil,
			mockVal:      module.TrafficGroupIntent{},
			client:       &mocks.TrafficGroupIntentManager{},
		}),
		Entry("fails due to other json validation error", testCase{
			// name field has an '=' character
			inputName:    "testtrafficgroupintent",
			expectedCode: http.StatusBadRequest,
			inputReader: bytes.NewBuffer([]byte(`{
				"metadata": {
					"name": "test=testtrafficgroupintent",
					"description": "test traffic group intent",
					"userData1": "some user data 1",
					"userData2": "some user data 2"
				}
			}`)),
			inStruct:  module.TrafficGroupIntent{},
			mockError: nil,
			mockVal:      module.TrafficGroupIntent{},
			client:    &mocks.TrafficGroupIntentManager{},
		}),
		Entry("fails due to json body decoding error", testCase{
			// extra comma at the end of the userData2 line
			inputName:    "testtrafficgroupintent",
			expectedCode: http.StatusUnprocessableEntity,
			inputReader: bytes.NewBuffer([]byte(`{
				"metadata": {
					"name": "testtrafficgroupintent",
					"description": "test traffic group intent",
					"userData1": "some user data 1",
					"userData2": "some user data 2",
				}
			}`)),
			inStruct:  module.TrafficGroupIntent{},
			mockError: nil,
			mockVal:      module.TrafficGroupIntent{},
			client:    &mocks.TrafficGroupIntentManager{},
		}),
		Entry("fails due to mismatched name", testCase{
			inputName:    "testtrafficgroupintentXYZ",
			expectedCode: http.StatusBadRequest,
			inputReader: bytes.NewBuffer([]byte(`{
				"metadata": {
					"name": "testtrafficgroupintent",
					"description": "test traffic group intent",
					"userData1": "some user data 1",
					"userData2": "some user data 2"
				}
			}`)),
			inStruct: module.TrafficGroupIntent{
				Metadata: module.Metadata{
					Name:	     "testtrafficgroupintent",
					Description: "test traffic group intent",
					UserData1:   "some user data 1",
					UserData2:   "some user data 2",
				},
			},
			mockError: pkgerrors.New("Creating DB Entry"),
			mockVal: module.TrafficGroupIntent{},
			client: &mocks.TrafficGroupIntentManager{},
		}),
		Entry("fails due to db error", testCase{
			inputName:    "testtrafficgroupintent",
			expectedCode: http.StatusInternalServerError,
			inputReader: bytes.NewBuffer([]byte(`{
				"metadata": {
					"name": "testtrafficgroupintent",
					"description": "test traffic group intent",
					"userData1": "some user data 1",
					"userData2": "some user data 2"
				}
			}`)),
			inStruct: module.TrafficGroupIntent{
				Metadata: module.Metadata{
					Name:	     "testtrafficgroupintent",
					Description: "test traffic group intent",
					UserData1:   "some user data 1",
					UserData2:   "some user data 2",
				},
			},
			mockError: pkgerrors.New("Creating DB Entry"),
			mockVal: module.TrafficGroupIntent{},
			client: &mocks.TrafficGroupIntentManager{},
		}),

	)
	DescribeTable("Get List TrafficGroupIntent tests",
		func(t testCase) {
			// set up client mock responses

			t.client.On("GetTrafficGroupIntents", "test-project", "test-compositeapp", "v1", "test-dig").Return(t.mockVals, t.mockError)

			// make HTTP request
			request := httptest.NewRequest("GET", "/v2/projects/test-project/composite-apps/test-compositeapp/v1/deployment-intent-groups/test-dig/traffic-group-intents", nil)
			resp := executeRequest(request, NewRouter(t.client))

			//Check returned code
			Expect(resp.StatusCode).To(Equal(t.expectedCode))

			//Check returned body
			got := []module.TrafficGroupIntent{}
			json.NewDecoder(resp.Body).Decode(&got)
			Expect(got).To(Equal(t.mockVals))
		},
		Entry("successful get", testCase{
			expectedCode: http.StatusOK,
			mockError: nil,
			mockVals: []module.TrafficGroupIntent{
				{
					Metadata: module.Metadata{
						Name:	     "testtrafficgroupintent1",
						Description: "test traffic group intent",
						UserData1:   "some user data 1",
						UserData2:   "some user data 2",
					},
				},
				{
					Metadata: module.Metadata{
						Name:	     "testtrafficgroupintent2",
						Description: "test traffic group intent",
						UserData1:   "some user data 1",
						UserData2:   "some user data 2",
					},
				},
			},
			client: &mocks.TrafficGroupIntentManager{},
		}),
		Entry("fails due to not found", testCase{
			expectedCode: http.StatusNotFound,
			mockError: pkgerrors.New("db Find error"),
			mockVals:  []module.TrafficGroupIntent{},
			client:    &mocks.TrafficGroupIntentManager{},
		}),
		Entry("fails due to some other backend error", testCase{
			expectedCode: http.StatusInternalServerError,
			mockError: pkgerrors.New("backend error"),
			mockVals:  []module.TrafficGroupIntent{},
			client:    &mocks.TrafficGroupIntentManager{},
		}),
	)
	DescribeTable("Get TrafficGroupIntent tests",
		func(t testCase) {
			// set up client mock responses

			t.client.On("GetTrafficGroupIntent", t.inputName, "test-project", "test-compositeapp", "v1", "test-dig").Return(t.mockVal, t.mockError)

			// make HTTP request
			request := httptest.NewRequest("GET", "/v2/projects/test-project/composite-apps/test-compositeapp/v1/deployment-intent-groups/test-dig/traffic-group-intents/"+t.inputName, nil)
			resp := executeRequest(request, NewRouter(t.client))

			//Check returned code
			Expect(resp.StatusCode).To(Equal(t.expectedCode))

			//Check returned body
			got := module.TrafficGroupIntent{}
			json.NewDecoder(resp.Body).Decode(&got)
			Expect(got).To(Equal(t.mockVal))
		},
		Entry("successful get", testCase{
			inputName:    "testtrafficgroupintent",
			expectedCode: http.StatusOK,
			mockError: nil,
			mockVal: module.TrafficGroupIntent{
				Metadata: module.Metadata{
					Name:	     "testtrafficgroupintent",
					Description: "test traffic group intent",
					UserData1:   "some user data 1",
					UserData2:   "some user data 2",
				},
			},
			client: &mocks.TrafficGroupIntentManager{},
		}),
		Entry("fails due to not found", testCase{
			inputName:    "testtrafficgroupintent",
			expectedCode: http.StatusNotFound,
			mockError:    pkgerrors.New("db Find error"),
			mockVal:     module.TrafficGroupIntent{},
			client:       &mocks.TrafficGroupIntentManager{},
		}),
		Entry("fails due to some other backend error", testCase{
			inputName:    "testtrafficgroupintent",
			expectedCode: http.StatusInternalServerError,
			mockError:    pkgerrors.New("backend error"),
			mockVal:     module.TrafficGroupIntent{},
			client:       &mocks.TrafficGroupIntentManager{},
		}),
	)
	DescribeTable("DELETE TrafficGroupIntent tests",
		func(t testCase) {
			// set up client mock responses

			t.client.On("DeleteTrafficGroupIntent", t.inputName, "test-project", "test-compositeapp", "v1", "test-dig").Return(t.mockError)

			// make HTTP request
			request := httptest.NewRequest("DELETE", "/v2/projects/test-project/composite-apps/test-compositeapp/v1/deployment-intent-groups/test-dig/traffic-group-intents/"+t.inputName, nil)
			resp := executeRequest(request, NewRouter(t.client))

			//Check returned code
			Expect(resp.StatusCode).To(Equal(t.expectedCode))

			//Check returned body
			got := module.TrafficGroupIntent{}
			json.NewDecoder(resp.Body).Decode(&got)
			Expect(got).To(Equal(t.mockVal))
		},
		Entry("successful delete", testCase{
			inputName:    "testtrafficgroupintent",
			expectedCode: http.StatusNoContent,
			mockError:    nil,
			mockVal:      module.TrafficGroupIntent{},
			client:       &mocks.TrafficGroupIntentManager{},
		}),
		Entry("fails due to not found", testCase{
			inputName:    "testtrafficgroupintent",
			expectedCode: http.StatusNotFound,
			mockError:    pkgerrors.New("db Remove error - not found"),
			mockVal:      module.TrafficGroupIntent{},
			client:       &mocks.TrafficGroupIntentManager{},
		}),
		Entry("fails due to conflict", testCase{
			inputName:    "testtrafficgroupintent",
			expectedCode: http.StatusConflict,
			mockError:    pkgerrors.New("db Remove error - conflict"),
			mockVal:      module.TrafficGroupIntent{},
			client:       &mocks.TrafficGroupIntentManager{},
		}),
		Entry("fails due to some other backend error", testCase{
			inputName:    "testtrafficgroupintent",
			expectedCode: http.StatusInternalServerError,
			mockError:    pkgerrors.New("db Remove error - general"),
			mockVal:      module.TrafficGroupIntent{},
			client:       &mocks.TrafficGroupIntentManager{},
		}),
	)

})
