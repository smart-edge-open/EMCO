package api

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	"github.com/open-ness/EMCO/src/ovnaction/api/mocks"
	"github.com/open-ness/EMCO/src/ovnaction/pkg/module"
	pkgerrors "github.com/pkg/errors"
)

func init() {
	workloadIntJSONFile = "../json-schemas/network-workload.json"
}

var _ = Describe("Workloadintenthandler", func() {

	type testCase struct {
		inputName    string
		inputReader  io.Reader
		inStruct     module.WorkloadIntent
		mockError    error
		mockVal      module.WorkloadIntent
		mockVals     []module.WorkloadIntent
		expectedCode int
		client       *mocks.WorkloadIntentManager
	}

	DescribeTable("Create WorkloadIntent tests",
		func(t testCase) {
			// set up client mock responses
			t.client.On("CreateWorkloadIntent", t.inStruct, "test-project", "test-compositeapp", "v1", "test-dig", "net-ctl-intent", false).Return(t.mockVal, t.mockError)

			// make HTTP request
			request := httptest.NewRequest("POST", "/v2/projects/test-project/composite-apps/test-compositeapp/v1/deployment-intent-groups/test-dig/network-controller-intent/net-ctl-intent/workload-intents", t.inputReader)
			resp := executeRequest(request, NewRouter(t.client))

			//Check returned code
			Expect(resp.StatusCode).To(Equal(t.expectedCode))

			//Check returned body
			got := module.WorkloadIntent{}
			json.NewDecoder(resp.Body).Decode(&got)
			Expect(got).To(Equal(t.mockVal))
		},

		Entry("successful create", testCase{
			expectedCode: http.StatusCreated,
			inputReader: bytes.NewBuffer([]byte(`{
				"metadata": {
					"name": "testworkloadintent",
					"description": "test workload intent",
					"userData1": "some user data 1",
					"userData2": "some user data 2"
				},
				"spec": {
					"application-name": "test-app",
					"workload-resource": "release-test-app",
					"type": "Deployment"
				}
			}`)),
			inStruct: module.WorkloadIntent{
				Metadata: module.Metadata{
					Name:        "testworkloadintent",
					Description: "test workload intent",
					UserData1:   "some user data 1",
					UserData2:   "some user data 2",
				},
				Spec: module.WorkloadIntentSpec{
					AppName:          "test-app",
					WorkloadResource: "release-test-app",
					Type:             "Deployment",
				},
			},
			mockError: nil,
			mockVal: module.WorkloadIntent{
				Metadata: module.Metadata{
					Name:        "testworkloadintent",
					Description: "test workload intent",
					UserData1:   "some user data 1",
					UserData2:   "some user data 2",
				},
				Spec: module.WorkloadIntentSpec{
					AppName:          "test-app",
					WorkloadResource: "release-test-app",
					Type:             "Deployment",
				},
			},
			client: &mocks.WorkloadIntentManager{},
		}),

		Entry("fails due to empty body", testCase{
			expectedCode: http.StatusBadRequest,
			inStruct:     module.WorkloadIntent{},
			mockError:    nil,
			mockVal:      module.WorkloadIntent{},
			client:       &mocks.WorkloadIntentManager{},
		}),

		Entry("fails due missing name", testCase{
			expectedCode: http.StatusBadRequest,
			inputReader: bytes.NewBuffer([]byte(`{
					"metadata": {
						"description": "test workload intent",
						"userData1": "some user data 1",
						"userData2": "some user data 2"
					},
					"spec": {
						"application-name": "test-app",
						"workload-resource": "release-test-app",
						"type": "Deployment"
					}
				}`)),
			inStruct:  module.WorkloadIntent{},
			mockError: nil,
			client:    &mocks.WorkloadIntentManager{},
		}),

		Entry("fails due missing app name", testCase{
			expectedCode: http.StatusBadRequest,
			inputReader: bytes.NewBuffer([]byte(`{
					"metadata": {
						"name": "testworkloadintent",
						"description": "test workload intent",
						"userData1": "some user data 1",
						"userData2": "some user data 2"
					},
					"spec": {
						"workload-resource": "release-test-app",
						"type": "Deployment"
					}
				}`)),
			inStruct:  module.WorkloadIntent{},
			mockError: nil,
			client:    &mocks.WorkloadIntentManager{},
		}),

		Entry("fails due missing workload resource", testCase{
			expectedCode: http.StatusBadRequest,
			inputReader: bytes.NewBuffer([]byte(`{
					"metadata": {
						"name": "testworkloadintent",
						"description": "test workload intent",
						"userData1": "some user data 1",
						"userData2": "some user data 2"
					},
					"spec": {
						"application-name": "test-app",
						"type": "Deployment"
					}
				}`)),
			inStruct:  module.WorkloadIntent{},
			mockError: nil,
			client:    &mocks.WorkloadIntentManager{},
		}),

		Entry("fails due missing type", testCase{
			expectedCode: http.StatusBadRequest,
			inputReader: bytes.NewBuffer([]byte(`{
					"metadata": {
						"name": "testworkloadintent",
						"description": "test workload intent",
						"userData1": "some user data 1",
						"userData2": "some user data 2"
					},
					"spec": {
						"application-name": "test-app",
						"workload-resource": "release-test-app"
					}
				}`)),
			inStruct:  module.WorkloadIntent{},
			mockError: nil,
			client:    &mocks.WorkloadIntentManager{},
		}),

		Entry("fails due to other json validation error", testCase{
			// name field has an '=' character
			expectedCode: http.StatusBadRequest,
			inputReader: bytes.NewBuffer([]byte(`{
					"metadata": {
						"name": "test=workloadintent",
						"description": "test workload intent",
						"userData1": "some user data 1",
						"userData2": "some user data 2"
					},
					"spec": {
						"application-name": "test-app",
						"workload-resource": "release-test-app",
						"type": "Deployment"
					}
				}`)),
			inStruct:  module.WorkloadIntent{},
			mockError: nil,
			client:    &mocks.WorkloadIntentManager{},
		}),

		Entry("fails due to json body decoding error", testCase{
			// extra comma at the end of the userData2 line
			expectedCode: http.StatusUnprocessableEntity,
			inputReader: bytes.NewBuffer([]byte(`{
					"metadata": {
						"name": "testworkloadintent",
						"description": "test workload intent",
						"userData1": "some user data 1",
						"userData2": "some user data 2",
					},
					"spec": {
						"application-name": "test-app",
						"workload-resource": "release-test-app",
						"type": "Deployment"
					}
				}`)),
			inStruct:  module.WorkloadIntent{},
			mockError: nil,
			client:    &mocks.WorkloadIntentManager{},
		}),

		Entry("fails due to entry already exists", testCase{
			expectedCode: http.StatusConflict,
			inputReader: bytes.NewBuffer([]byte(`{
					"metadata": {
						"name": "testworkloadintent",
						"description": "test workload intent",
						"userData1": "some user data 1",
						"userData2": "some user data 2"
					},
					"spec": {
						"application-name": "test-app",
						"workload-resource": "release-test-app",
						"type": "Deployment"
					}
				}`)),
			inStruct: module.WorkloadIntent{
				Metadata: module.Metadata{
					Name:        "testworkloadintent",
					Description: "test workload intent",
					UserData1:   "some user data 1",
					UserData2:   "some user data 2",
				},
				Spec: module.WorkloadIntentSpec{
					AppName:          "test-app",
					WorkloadResource: "release-test-app",
					Type:             "Deployment",
				},
			},
			mockVal:   module.WorkloadIntent{},
			mockError: pkgerrors.New("WorkloadIntent already exists"),
			client:    &mocks.WorkloadIntentManager{},
		}),

		Entry("fails due to network control intent does not exist", testCase{
			expectedCode: http.StatusNotFound,
			inputReader: bytes.NewBuffer([]byte(`{
					"metadata": {
						"name": "testworkloadintent",
						"description": "test workload intent",
						"userData1": "some user data 1",
						"userData2": "some user data 2"
					},
					"spec": {
						"application-name": "test-app",
						"workload-resource": "release-test-app",
						"type": "Deployment"
					}
				}`)),
			inStruct: module.WorkloadIntent{
				Metadata: module.Metadata{
					Name:        "testworkloadintent",
					Description: "test workload intent",
					UserData1:   "some user data 1",
					UserData2:   "some user data 2",
				},
				Spec: module.WorkloadIntentSpec{
					AppName:          "test-app",
					WorkloadResource: "release-test-app",
					Type:             "Deployment",
				},
			},
			mockVal:   module.WorkloadIntent{},
			mockError: pkgerrors.New("does not exist"),
			client:    &mocks.WorkloadIntentManager{},
		}),

		Entry("fails due to db error", testCase{
			expectedCode: http.StatusInternalServerError,
			inputReader: bytes.NewBuffer([]byte(`{
					"metadata": {
						"name": "testworkloadintent",
						"description": "test workload intent",
						"userData1": "some user data 1",
						"userData2": "some user data 2"
					},
					"spec": {
						"application-name": "test-app",
						"workload-resource": "release-test-app",
						"type": "Deployment"
					}
				}`)),
			inStruct: module.WorkloadIntent{
				Metadata: module.Metadata{
					Name:        "testworkloadintent",
					Description: "test workload intent",
					UserData1:   "some user data 1",
					UserData2:   "some user data 2",
				},
				Spec: module.WorkloadIntentSpec{
					AppName:          "test-app",
					WorkloadResource: "release-test-app",
					Type:             "Deployment",
				},
			},
			mockVal:   module.WorkloadIntent{},
			mockError: pkgerrors.New("Creating DB Entry"),
			client:    &mocks.WorkloadIntentManager{},
		}),
	)

	DescribeTable("Put WorkloadIntent tests",
		func(t testCase) {
			// set up client mock responses
			t.client.On("CreateWorkloadIntent", t.inStruct, "test-project", "test-compositeapp", "v1", "test-dig", "net-ctl-intent", true).Return(t.mockVal, t.mockError)

			// make HTTP request
			request := httptest.NewRequest("PUT", "/v2/projects/test-project/composite-apps/test-compositeapp/v1/deployment-intent-groups/test-dig/network-controller-intent/net-ctl-intent/workload-intents/"+t.inputName, t.inputReader)
			resp := executeRequest(request, NewRouter(t.client))

			//Check returned code
			Expect(resp.StatusCode).To(Equal(t.expectedCode))

			//Check returned body
			got := module.WorkloadIntent{}
			json.NewDecoder(resp.Body).Decode(&got)
			Expect(got).To(Equal(t.mockVal))
		},

		Entry("successful put", testCase{
			expectedCode: http.StatusCreated,
			inputName:    "testworkloadintent",
			inputReader: bytes.NewBuffer([]byte(`{
				"metadata": {
					"name": "testworkloadintent",
					"description": "test workload intent",
					"userData1": "some user data 1",
					"userData2": "some user data 2"
				},
				"spec": {
					"application-name": "test-app",
					"workload-resource": "release-test-app",
					"type": "Deployment"
				}
			}`)),
			inStruct: module.WorkloadIntent{
				Metadata: module.Metadata{
					Name:        "testworkloadintent",
					Description: "test workload intent",
					UserData1:   "some user data 1",
					UserData2:   "some user data 2",
				},
				Spec: module.WorkloadIntentSpec{
					AppName:          "test-app",
					WorkloadResource: "release-test-app",
					Type:             "Deployment",
				},
			},
			mockError: nil,
			mockVal: module.WorkloadIntent{
				Metadata: module.Metadata{
					Name:        "testworkloadintent",
					Description: "test workload intent",
					UserData1:   "some user data 1",
					UserData2:   "some user data 2",
				},
				Spec: module.WorkloadIntentSpec{
					AppName:          "test-app",
					WorkloadResource: "release-test-app",
					Type:             "Deployment",
				},
			},
			client: &mocks.WorkloadIntentManager{},
		}),

		Entry("fails due to empty body", testCase{
			expectedCode: http.StatusBadRequest,
			inputName:    "testworkloadintent",
			inStruct:     module.WorkloadIntent{},
			mockError:    nil,
			mockVal:      module.WorkloadIntent{},
			client:       &mocks.WorkloadIntentManager{},
		}),

		Entry("fails due missing name", testCase{
			expectedCode: http.StatusBadRequest,
			inputName:    "testworkloadintent",
			inputReader: bytes.NewBuffer([]byte(`{
					"metadata": {
						"description": "test workload intent",
						"userData1": "some user data 1",
						"userData2": "some user data 2"
					},
					"spec": {
						"application-name": "test-app",
						"workload-resource": "release-test-app",
						"type": "Deployment"
					}
				}`)),
			inStruct:  module.WorkloadIntent{},
			mockError: nil,
			client:    &mocks.WorkloadIntentManager{},
		}),

		Entry("fails due missing app name", testCase{
			expectedCode: http.StatusBadRequest,
			inputName:    "testworkloadintent",
			inputReader: bytes.NewBuffer([]byte(`{
					"metadata": {
						"name": "testworkloadintent",
						"description": "test workload intent",
						"userData1": "some user data 1",
						"userData2": "some user data 2"
					},
					"spec": {
						"workload-resource": "release-test-app",
						"type": "Deployment"
					}
				}`)),
			inStruct:  module.WorkloadIntent{},
			mockError: nil,
			client:    &mocks.WorkloadIntentManager{},
		}),

		Entry("fails due missing workload resource", testCase{
			expectedCode: http.StatusBadRequest,
			inputName:    "testworkloadintent",
			inputReader: bytes.NewBuffer([]byte(`{
					"metadata": {
						"name": "testworkloadintent",
						"description": "test workload intent",
						"userData1": "some user data 1",
						"userData2": "some user data 2"
					},
					"spec": {
						"application-name": "test-app",
						"type": "Deployment"
					}
				}`)),
			inStruct:  module.WorkloadIntent{},
			mockError: nil,
			client:    &mocks.WorkloadIntentManager{},
		}),

		Entry("fails due missing type", testCase{
			expectedCode: http.StatusBadRequest,
			inputName:    "testworkloadintent",
			inputReader: bytes.NewBuffer([]byte(`{
					"metadata": {
						"name": "testworkloadintent",
						"description": "test workload intent",
						"userData1": "some user data 1",
						"userData2": "some user data 2"
					},
					"spec": {
						"application-name": "test-app",
						"workload-resource": "release-test-app"
					}
				}`)),
			inStruct:  module.WorkloadIntent{},
			mockError: nil,
			client:    &mocks.WorkloadIntentManager{},
		}),

		Entry("fails due to other json validation error", testCase{
			// name field has an '=' character
			expectedCode: http.StatusBadRequest,
			inputName:    "testworkloadintent",
			inputReader: bytes.NewBuffer([]byte(`{
					"metadata": {
						"name": "test=workloadintent",
						"description": "test workload intent",
						"userData1": "some user data 1",
						"userData2": "some user data 2"
					},
					"spec": {
						"application-name": "test-app",
						"workload-resource": "release-test-app",
						"type": "Deployment"
					}
				}`)),
			inStruct:  module.WorkloadIntent{},
			mockError: nil,
			client:    &mocks.WorkloadIntentManager{},
		}),

		Entry("fails due to json body decoding error", testCase{
			// extra comma at the end of the userData2 line
			expectedCode: http.StatusUnprocessableEntity,
			inputName:    "testworkloadintent",
			inputReader: bytes.NewBuffer([]byte(`{
					"metadata": {
						"name": "testworkloadintent",
						"description": "test workload intent",
						"userData1": "some user data 1",
						"userData2": "some user data 2",
					},
					"spec": {
						"application-name": "test-app",
						"workload-resource": "release-test-app",
						"type": "Deployment"
					}
				}`)),
			inStruct:  module.WorkloadIntent{},
			mockError: nil,
			client:    &mocks.WorkloadIntentManager{},
		}),

		Entry("fails due to name mismatch", testCase{
			expectedCode: http.StatusBadRequest,
			inputName:    "testworkloadintent",
			inputReader: bytes.NewBuffer([]byte(`{
					"metadata": {
						"name": "testworkloadintentABC",
						"description": "test workload intent",
						"userData1": "some user data 1",
						"userData2": "some user data 2"
					},
					"spec": {
						"application-name": "test-app",
						"workload-resource": "release-test-app",
						"type": "Deployment"
					}
				}`)),
			inStruct: module.WorkloadIntent{
				Metadata: module.Metadata{
					Name:        "testworkloadintent",
					Description: "test workload intent",
					UserData1:   "some user data 1",
					UserData2:   "some user data 2",
				},
				Spec: module.WorkloadIntentSpec{
					AppName:          "test-app",
					WorkloadResource: "release-test-app",
					Type:             "Deployment",
				},
			},
			mockVal:   module.WorkloadIntent{},
			mockError: pkgerrors.New("WorkloadIntent already exists"),
			client:    &mocks.WorkloadIntentManager{},
		}),

		Entry("fails due to network control intent does not exist", testCase{
			expectedCode: http.StatusNotFound,
			inputName:    "testworkloadintent",
			inputReader: bytes.NewBuffer([]byte(`{
					"metadata": {
						"name": "testworkloadintent",
						"description": "test workload intent",
						"userData1": "some user data 1",
						"userData2": "some user data 2"
					},
					"spec": {
						"application-name": "test-app",
						"workload-resource": "release-test-app",
						"type": "Deployment"
					}
				}`)),
			inStruct: module.WorkloadIntent{
				Metadata: module.Metadata{
					Name:        "testworkloadintent",
					Description: "test workload intent",
					UserData1:   "some user data 1",
					UserData2:   "some user data 2",
				},
				Spec: module.WorkloadIntentSpec{
					AppName:          "test-app",
					WorkloadResource: "release-test-app",
					Type:             "Deployment",
				},
			},
			mockVal:   module.WorkloadIntent{},
			mockError: pkgerrors.New("does not exist"),
			client:    &mocks.WorkloadIntentManager{},
		}),

		Entry("fails due to db error", testCase{
			expectedCode: http.StatusInternalServerError,
			inputName:    "testworkloadintent",
			inputReader: bytes.NewBuffer([]byte(`{
					"metadata": {
						"name": "testworkloadintent",
						"description": "test workload intent",
						"userData1": "some user data 1",
						"userData2": "some user data 2"
					},
					"spec": {
						"application-name": "test-app",
						"workload-resource": "release-test-app",
						"type": "Deployment"
					}
				}`)),
			inStruct: module.WorkloadIntent{
				Metadata: module.Metadata{
					Name:        "testworkloadintent",
					Description: "test workload intent",
					UserData1:   "some user data 1",
					UserData2:   "some user data 2",
				},
				Spec: module.WorkloadIntentSpec{
					AppName:          "test-app",
					WorkloadResource: "release-test-app",
					Type:             "Deployment",
				},
			},
			mockVal:   module.WorkloadIntent{},
			mockError: pkgerrors.New("Creating DB Entry"),
			client:    &mocks.WorkloadIntentManager{},
		}),
	)

	DescribeTable("Get List WorkloadIntent tests",
		func(t testCase) {
			// set up client mock responses
			t.client.On("GetWorkloadIntents", "test-project", "test-compositeapp", "v1", "test-dig", "net-ctl-intent").Return(t.mockVals, t.mockError)

			// make HTTP request
			request := httptest.NewRequest("GET", "/v2/projects/test-project/composite-apps/test-compositeapp/v1/deployment-intent-groups/test-dig/network-controller-intent/net-ctl-intent/workload-intents", nil)
			resp := executeRequest(request, NewRouter(t.client))

			//Check returned code
			Expect(resp.StatusCode).To(Equal(t.expectedCode))

			//Check returned body
			got := []module.WorkloadIntent{}
			json.NewDecoder(resp.Body).Decode(&got)
			Expect(got).To(Equal(t.mockVals))
		},

		Entry("successful get", testCase{
			expectedCode: http.StatusOK,
			mockError:    nil,
			mockVals: []module.WorkloadIntent{
				{
					Metadata: module.Metadata{
						Name:        "testworkloadintent1",
						Description: "test workload intent",
						UserData1:   "some user data 1",
						UserData2:   "some user data 2",
					},
					Spec: module.WorkloadIntentSpec{
						AppName:          "test-app1",
						WorkloadResource: "release-test-app1",
						Type:             "Deployment",
					},
				},
				{
					Metadata: module.Metadata{
						Name:        "testworkloadintent2",
						Description: "test workload intent",
						UserData1:   "some user data 1",
						UserData2:   "some user data 2",
					},
					Spec: module.WorkloadIntentSpec{
						AppName:          "test-app2",
						WorkloadResource: "release-test-app2",
						Type:             "Deployment",
					},
				},
			},
			client: &mocks.WorkloadIntentManager{},
		}),

		Entry("fails due to not found", testCase{
			expectedCode: http.StatusNotFound,
			mockError:    pkgerrors.New("db Find error"),
			mockVals:     []module.WorkloadIntent{},
			client:       &mocks.WorkloadIntentManager{},
		}),

		Entry("fails due to some other backend error", testCase{
			expectedCode: http.StatusInternalServerError,
			mockError:    pkgerrors.New("backend error"),
			mockVals:     []module.WorkloadIntent{},
			client:       &mocks.WorkloadIntentManager{},
		}),
	)

	DescribeTable("Get WorkloadIntent tests",
		func(t testCase) {
			// set up client mock responses
			t.client.On("GetWorkloadIntent", t.inputName, "test-project", "test-compositeapp", "v1", "test-dig", "net-ctl-intent").Return(t.mockVal, t.mockError)

			// make HTTP request
			request := httptest.NewRequest("GET", "/v2/projects/test-project/composite-apps/test-compositeapp/v1/deployment-intent-groups/test-dig/network-controller-intent/net-ctl-intent/workload-intents/"+t.inputName, nil)
			resp := executeRequest(request, NewRouter(t.client))

			//Check returned code
			Expect(resp.StatusCode).To(Equal(t.expectedCode))

			//Check returned body
			got := module.WorkloadIntent{}
			json.NewDecoder(resp.Body).Decode(&got)
			Expect(got).To(Equal(t.mockVal))
		},

		Entry("successful get", testCase{
			inputName:    "testworkloadintent",
			expectedCode: http.StatusOK,
			mockError:    nil,
			mockVal: module.WorkloadIntent{
				Metadata: module.Metadata{
					Name:        "testworkloadintent",
					Description: "test workload intent",
					UserData1:   "some user data 1",
					UserData2:   "some user data 2",
				},
				Spec: module.WorkloadIntentSpec{
					AppName:          "test-app",
					WorkloadResource: "release-test-app",
					Type:             "Deployment",
				},
			},
			client: &mocks.WorkloadIntentManager{},
		}),

		Entry("fails due to not found", testCase{
			inputName:    "testworkloadintent",
			expectedCode: http.StatusNotFound,
			mockError:    pkgerrors.New("db Find error"),
			mockVal:      module.WorkloadIntent{},
			client:       &mocks.WorkloadIntentManager{},
		}),

		Entry("fails due to some other backend error", testCase{
			inputName:    "testworkloadintent",
			expectedCode: http.StatusInternalServerError,
			mockError:    pkgerrors.New("backend error"),
			mockVal:      module.WorkloadIntent{},
			client:       &mocks.WorkloadIntentManager{},
		}),
	)

	DescribeTable("Delete WorkloadIntent tests",
		func(t testCase) {
			// set up client mock responses
			t.client.On("DeleteWorkloadIntent", t.inputName, "test-project", "test-compositeapp", "v1", "test-dig", "net-ctl-intent").Return(t.mockError)

			// make HTTP request
			request := httptest.NewRequest("DELETE", "/v2/projects/test-project/composite-apps/test-compositeapp/v1/deployment-intent-groups/test-dig/network-controller-intent/net-ctl-intent/workload-intents/"+t.inputName, nil)
			resp := executeRequest(request, NewRouter(t.client))

			//Check returned code
			Expect(resp.StatusCode).To(Equal(t.expectedCode))

			//Check returned body
			got := module.WorkloadIntent{}
			json.NewDecoder(resp.Body).Decode(&got)
			Expect(got).To(Equal(t.mockVal))
		},

		Entry("successful delete", testCase{
			inputName:    "testworkloadintent",
			expectedCode: http.StatusNoContent,
			mockError:    nil,
			client:       &mocks.WorkloadIntentManager{},
		}),

		Entry("fails due to not found", testCase{
			inputName:    "testworkloadintent",
			expectedCode: http.StatusNotFound,
			mockError:    pkgerrors.New("db Remove error - not found"),
			client:       &mocks.WorkloadIntentManager{},
		}),

		Entry("fails due to a conflict", testCase{
			inputName:    "testworkloadintent",
			expectedCode: http.StatusConflict,
			mockError:    pkgerrors.New("db Remove error - conflict"),
			client:       &mocks.WorkloadIntentManager{},
		}),

		Entry("fails due to other backend error", testCase{
			inputName:    "testworkloadintent",
			expectedCode: http.StatusInternalServerError,
			mockError:    pkgerrors.New("db Remove error - general"),
			client:       &mocks.WorkloadIntentManager{},
		}),
	)
})
