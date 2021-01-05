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
	netIfJSONFile = "../json-schemas/network-load-interface.json"
}

var _ = Describe("Workloadifintenthandler", func() {

	type testCase struct {
		inputName    string
		inputReader  io.Reader
		inStruct     module.WorkloadIfIntent
		mockError    error
		mockVal      module.WorkloadIfIntent
		mockVals     []module.WorkloadIfIntent
		expectedCode int
		client       *mocks.WorkloadIfIntentManager
	}

	DescribeTable("Create WorkloadIfIntent tests",
		func(t testCase) {
			// set up client mock responses
			t.client.On("CreateWorkloadIfIntent", t.inStruct, "test-project", "test-compositeapp", "v1", "test-dig", "net-ctl-intent", "testworkloadintent", false).Return(t.mockVal, t.mockError)

			// make HTTP request
			request := httptest.NewRequest("POST", "/v2/projects/test-project/composite-apps/test-compositeapp/v1/deployment-intent-groups/test-dig/network-controller-intent/net-ctl-intent/workload-intents/testworkloadintent/interfaces", t.inputReader)
			resp := executeRequest(request, NewRouter(t.client))

			//Check returned code
			Expect(resp.StatusCode).To(Equal(t.expectedCode))

			//Check returned body
			got := module.WorkloadIfIntent{}
			json.NewDecoder(resp.Body).Decode(&got)
			Expect(got).To(Equal(t.mockVal))
		},

		Entry("successful create", testCase{
			expectedCode: http.StatusCreated,
			inputReader: bytes.NewBuffer([]byte(`{
				"metadata": {
					"name": "testinterfaceintent",
					"description": "test interface intent",
					"userData1": "some user data 1",
					"userData2": "some user data 2"
				},
				"spec": {
					"interface": "eth1",
					"name": "networkA",
					"defaultGateway": "false",
					"ipAddress": "10.10.10.10"
				}
			}`)),
			inStruct: module.WorkloadIfIntent{
				Metadata: module.Metadata{
					Name:        "testinterfaceintent",
					Description: "test interface intent",
					UserData1:   "some user data 1",
					UserData2:   "some user data 2",
				},
				Spec: module.WorkloadIfIntentSpec{
					IfName:         "eth1",
					NetworkName:    "networkA",
					DefaultGateway: "false",
					IpAddr:         "10.10.10.10",
				},
			},
			mockError: nil,
			mockVal: module.WorkloadIfIntent{
				Metadata: module.Metadata{
					Name:        "testinterfaceintent",
					Description: "test interface intent",
					UserData1:   "some user data 1",
					UserData2:   "some user data 2",
				},
				Spec: module.WorkloadIfIntentSpec{
					IfName:         "eth1",
					NetworkName:    "networkA",
					DefaultGateway: "false",
					IpAddr:         "10.10.10.10",
				},
			},
			client: &mocks.WorkloadIfIntentManager{},
		}),

		Entry("successful create - default gateway", testCase{
			expectedCode: http.StatusCreated,
			inputReader: bytes.NewBuffer([]byte(`{
				"metadata": {
					"name": "testinterfaceintent",
					"description": "test interface intent",
					"userData1": "some user data 1",
					"userData2": "some user data 2"
				},
				"spec": {
					"interface": "eth1",
					"name": "networkA",
					"ipAddress": "10.10.10.10"
				}
			}`)),
			inStruct: module.WorkloadIfIntent{
				Metadata: module.Metadata{
					Name:        "testinterfaceintent",
					Description: "test interface intent",
					UserData1:   "some user data 1",
					UserData2:   "some user data 2",
				},
				Spec: module.WorkloadIfIntentSpec{
					IfName:         "eth1",
					NetworkName:    "networkA",
					DefaultGateway: "false",
					IpAddr:         "10.10.10.10",
				},
			},
			mockError: nil,
			mockVal: module.WorkloadIfIntent{
				Metadata: module.Metadata{
					Name:        "testinterfaceintent",
					Description: "test interface intent",
					UserData1:   "some user data 1",
					UserData2:   "some user data 2",
				},
				Spec: module.WorkloadIfIntentSpec{
					IfName:         "eth1",
					NetworkName:    "networkA",
					DefaultGateway: "false",
					IpAddr:         "10.10.10.10",
				},
			},
			client: &mocks.WorkloadIfIntentManager{},
		}),

		Entry("fails due to empty body", testCase{
			expectedCode: http.StatusBadRequest,
			inStruct:     module.WorkloadIfIntent{},
			mockError:    nil,
			mockVal:      module.WorkloadIfIntent{},
			client:       &mocks.WorkloadIfIntentManager{},
		}),

		Entry("fails due missing name", testCase{
			expectedCode: http.StatusBadRequest,
			inputReader: bytes.NewBuffer([]byte(`{
					"metadata": {
						"description": "test interface intent",
						"userData1": "some user data 1",
						"userData2": "some user data 2"
					},
					"spec": {
						"interface": "eth1",
						"name": "networkA",
						"defaultGateway": "false",
						"ipAddress": "10.10.10.10"
						}
				}`)),
			inStruct:  module.WorkloadIfIntent{},
			mockError: nil,
			client:    &mocks.WorkloadIfIntentManager{},
		}),

		Entry("fails due missing interface resource", testCase{
			expectedCode: http.StatusBadRequest,
			inputReader: bytes.NewBuffer([]byte(`{
					"metadata": {
						"name": "testinterfaceintent",
						"description": "test interface intent",
						"userData1": "some user data 1",
						"userData2": "some user data 2"
					},
					"spec": {
						"name": "networkA",
						"defaultGateway": "false",
						"ipAddress": "10.10.10.10"
					}
				}`)),
			inStruct:  module.WorkloadIfIntent{},
			mockError: nil,
			client:    &mocks.WorkloadIfIntentManager{},
		}),

		Entry("fails due missing spec name", testCase{
			expectedCode: http.StatusBadRequest,
			inputReader: bytes.NewBuffer([]byte(`{
					"metadata": {
						"name": "testinterfaceintent",
						"description": "test interface intent",
						"userData1": "some user data 1",
						"userData2": "some user data 2"
					},
					"spec": {
						"interface": "eth1",
						"defaultGateway": "false",
						"ipAddress": "10.10.10.10"
					}
				}`)),
			inStruct:  module.WorkloadIfIntent{},
			mockError: nil,
			client:    &mocks.WorkloadIfIntentManager{},
		}),

		Entry("fails due to other json validation error", testCase{
			// name field has an '=' character
			expectedCode: http.StatusBadRequest,
			inputReader: bytes.NewBuffer([]byte(`{
					"metadata": {
						"name": "test=interfaceintent",
						"description": "test interface intent",
						"userData1": "some user data 1",
						"userData2": "some user data 2"
					},
					"spec": {
						"interface": "eth1",
						"name": "networkA",
						"defaultGateway": "false",
						"ipAddress": "10.10.10.10"
					}
				}`)),
			inStruct:  module.WorkloadIfIntent{},
			mockError: nil,
			client:    &mocks.WorkloadIfIntentManager{},
		}),

		Entry("fails due to json body decoding error", testCase{
			// extra comma at the end of the userData2 line
			expectedCode: http.StatusUnprocessableEntity,
			inputReader: bytes.NewBuffer([]byte(`{
					"metadata": {
						"name": "testinterfaceintent",
						"description": "test interface intent",
						"userData1": "some user data 1",
						"userData2": "some user data 2",
					},
					"spec": {
						"interface": "eth1",
						"name": "networkA",
						"defaultGateway": "false",
						"ipAddress": "10.10.10.10"
					}
				}`)),
			inStruct:  module.WorkloadIfIntent{},
			mockError: nil,
			client:    &mocks.WorkloadIfIntentManager{},
		}),

		Entry("fails due to entry already exists", testCase{
			expectedCode: http.StatusConflict,
			inputReader: bytes.NewBuffer([]byte(`{
					"metadata": {
						"name": "testinterfaceintent",
						"description": "test interface intent",
						"userData1": "some user data 1",
						"userData2": "some user data 2"
					},
					"spec": {
						"interface": "eth1",
						"name": "networkA",
						"defaultGateway": "false",
						"ipAddress": "10.10.10.10"
					}
				}`)),
			inStruct: module.WorkloadIfIntent{
				Metadata: module.Metadata{
					Name:        "testinterfaceintent",
					Description: "test interface intent",
					UserData1:   "some user data 1",
					UserData2:   "some user data 2",
				},
				Spec: module.WorkloadIfIntentSpec{
					IfName:         "eth1",
					NetworkName:    "networkA",
					DefaultGateway: "false",
					IpAddr:         "10.10.10.10",
				},
			},
			mockVal:   module.WorkloadIfIntent{},
			mockError: pkgerrors.New("WorkloadIfIntent already exists"),
			client:    &mocks.WorkloadIfIntentManager{},
		}),

		Entry("fails due to network control intent does not exist", testCase{
			expectedCode: http.StatusNotFound,
			inputReader: bytes.NewBuffer([]byte(`{
					"metadata": {
						"name": "testinterfaceintent",
						"description": "test interface intent",
						"userData1": "some user data 1",
						"userData2": "some user data 2"
					},
					"spec": {
						"interface": "eth1",
						"name": "networkA",
						"defaultGateway": "false",
						"ipAddress": "10.10.10.10"
					}
				}`)),
			inStruct: module.WorkloadIfIntent{
				Metadata: module.Metadata{
					Name:        "testinterfaceintent",
					Description: "test interface intent",
					UserData1:   "some user data 1",
					UserData2:   "some user data 2",
				},
				Spec: module.WorkloadIfIntentSpec{
					IfName:         "eth1",
					NetworkName:    "networkA",
					DefaultGateway: "false",
					IpAddr:         "10.10.10.10",
				},
			},
			mockVal:   module.WorkloadIfIntent{},
			mockError: pkgerrors.New("does not exist"),
			client:    &mocks.WorkloadIfIntentManager{},
		}),

		Entry("fails due to db error", testCase{
			expectedCode: http.StatusInternalServerError,
			inputReader: bytes.NewBuffer([]byte(`{
					"metadata": {
						"name": "testinterfaceintent",
						"description": "test interface intent",
						"userData1": "some user data 1",
						"userData2": "some user data 2"
					},
					"spec": {
						"interface": "eth1",
						"name": "networkA",
						"defaultGateway": "false",
						"ipAddress": "10.10.10.10"
					}
				}`)),
			inStruct: module.WorkloadIfIntent{
				Metadata: module.Metadata{
					Name:        "testinterfaceintent",
					Description: "test interface intent",
					UserData1:   "some user data 1",
					UserData2:   "some user data 2",
				},
				Spec: module.WorkloadIfIntentSpec{
					IfName:         "eth1",
					NetworkName:    "networkA",
					DefaultGateway: "false",
					IpAddr:         "10.10.10.10",
				},
			},
			mockVal:   module.WorkloadIfIntent{},
			mockError: pkgerrors.New("Creating DB Entry"),
			client:    &mocks.WorkloadIfIntentManager{},
		}),
	)

	DescribeTable("Put WorkloadIfIntent tests",
		func(t testCase) {
			// set up client mock responses
			t.client.On("CreateWorkloadIfIntent", t.inStruct, "test-project", "test-compositeapp", "v1", "test-dig", "net-ctl-intent", "testworkloadintent", true).Return(t.mockVal, t.mockError)

			// make HTTP request
			request := httptest.NewRequest("PUT", "/v2/projects/test-project/composite-apps/test-compositeapp/v1/deployment-intent-groups/test-dig/network-controller-intent/net-ctl-intent/workload-intents/testworkloadintent/interfaces/"+t.inputName, t.inputReader)
			resp := executeRequest(request, NewRouter(t.client))

			//Check returned code
			Expect(resp.StatusCode).To(Equal(t.expectedCode))

			//Check returned body
			got := module.WorkloadIfIntent{}
			json.NewDecoder(resp.Body).Decode(&got)
			Expect(got).To(Equal(t.mockVal))
		},

		Entry("successful put", testCase{
			expectedCode: http.StatusCreated,
			inputName:    "testinterfaceintent",
			inputReader: bytes.NewBuffer([]byte(`{
				"metadata": {
					"name": "testinterfaceintent",
					"description": "test interface intent",
					"userData1": "some user data 1",
					"userData2": "some user data 2"
				},
				"spec": {
					"interface": "eth1",
					"name": "networkA",
					"defaultGateway": "false",
					"ipAddress": "10.10.10.10"
				}
			}`)),
			inStruct: module.WorkloadIfIntent{
				Metadata: module.Metadata{
					Name:        "testinterfaceintent",
					Description: "test interface intent",
					UserData1:   "some user data 1",
					UserData2:   "some user data 2",
				},
				Spec: module.WorkloadIfIntentSpec{
					IfName:         "eth1",
					NetworkName:    "networkA",
					DefaultGateway: "false",
					IpAddr:         "10.10.10.10",
				},
			},
			mockError: nil,
			mockVal: module.WorkloadIfIntent{
				Metadata: module.Metadata{
					Name:        "testinterfaceintent",
					Description: "test interface intent",
					UserData1:   "some user data 1",
					UserData2:   "some user data 2",
				},
				Spec: module.WorkloadIfIntentSpec{
					IfName:         "eth1",
					NetworkName:    "networkA",
					DefaultGateway: "false",
					IpAddr:         "10.10.10.10",
				},
			},
			client: &mocks.WorkloadIfIntentManager{},
		}),

		Entry("fails due to empty body", testCase{
			expectedCode: http.StatusBadRequest,
			inputName:    "testinterfaceintent",
			inStruct:     module.WorkloadIfIntent{},
			mockError:    nil,
			mockVal:      module.WorkloadIfIntent{},
			client:       &mocks.WorkloadIfIntentManager{},
		}),

		Entry("fails due missing name", testCase{
			expectedCode: http.StatusBadRequest,
			inputName:    "testinterfaceintent",
			inputReader: bytes.NewBuffer([]byte(`{
					"metadata": {
						"description": "test interface intent",
						"userData1": "some user data 1",
						"userData2": "some user data 2"
					},
					"spec": {
						"interface": "eth1",
						"name": "networkA",
						"defaultGateway": "false",
						"ipAddress": "10.10.10.10"
					}
				}`)),
			inStruct:  module.WorkloadIfIntent{},
			mockError: nil,
			client:    &mocks.WorkloadIfIntentManager{},
		}),

		Entry("fails due to other json validation error", testCase{
			// name field has an '=' character
			expectedCode: http.StatusBadRequest,
			inputName:    "testinterfaceintent",
			inputReader: bytes.NewBuffer([]byte(`{
					"metadata": {
						"name": "test=interfaceintent",
						"description": "test interface intent",
						"userData1": "some user data 1",
						"userData2": "some user data 2"
					},
					"spec": {
						"interface": "eth1",
						"name": "networkA",
						"defaultGateway": "false",
						"ipAddress": "10.10.10.10"
					}
				}`)),
			inStruct:  module.WorkloadIfIntent{},
			mockError: nil,
			client:    &mocks.WorkloadIfIntentManager{},
		}),

		Entry("fails due to json body decoding error", testCase{
			// extra comma at the end of the userData2 line
			expectedCode: http.StatusUnprocessableEntity,
			inputName:    "testinterfaceintent",
			inputReader: bytes.NewBuffer([]byte(`{
					"metadata": {
						"name": "testinterfaceintent",
						"description": "test interface intent",
						"userData1": "some user data 1",
						"userData2": "some user data 2",
					},
					"spec": {
						"interface": "eth1",
						"name": "networkA",
						"defaultGateway": "false",
						"ipAddress": "10.10.10.10"
					}
				}`)),
			inStruct:  module.WorkloadIfIntent{},
			mockError: nil,
			client:    &mocks.WorkloadIfIntentManager{},
		}),

		Entry("fails due to name mismatch", testCase{
			expectedCode: http.StatusBadRequest,
			inputName:    "testinterfaceintent",
			inputReader: bytes.NewBuffer([]byte(`{
					"metadata": {
						"name": "testinterfaceintentABC",
						"description": "test interface intent",
						"userData1": "some user data 1",
						"userData2": "some user data 2"
					},
					"spec": {
						"interface": "eth1",
						"name": "networkA",
						"defaultGateway": "false",
						"ipAddress": "10.10.10.10"
					}
				}`)),
			inStruct: module.WorkloadIfIntent{
				Metadata: module.Metadata{
					Name:        "testinterfaceintent",
					Description: "test interface intent",
					UserData1:   "some user data 1",
					UserData2:   "some user data 2",
				},
				Spec: module.WorkloadIfIntentSpec{
					IfName:         "eth1",
					NetworkName:    "networkA",
					DefaultGateway: "false",
					IpAddr:         "10.10.10.10",
				},
			},
			mockVal:   module.WorkloadIfIntent{},
			mockError: pkgerrors.New("WorkloadIfIntent already exists"),
			client:    &mocks.WorkloadIfIntentManager{},
		}),

		Entry("fails due to network control intent does not exist", testCase{
			expectedCode: http.StatusNotFound,
			inputName:    "testinterfaceintent",
			inputReader: bytes.NewBuffer([]byte(`{
					"metadata": {
						"name": "testinterfaceintent",
						"description": "test interface intent",
						"userData1": "some user data 1",
						"userData2": "some user data 2"
					},
					"spec": {
						"interface": "eth1",
						"name": "networkA",
						"defaultGateway": "false",
						"ipAddress": "10.10.10.10"
					}
				}`)),
			inStruct: module.WorkloadIfIntent{
				Metadata: module.Metadata{
					Name:        "testinterfaceintent",
					Description: "test interface intent",
					UserData1:   "some user data 1",
					UserData2:   "some user data 2",
				},
				Spec: module.WorkloadIfIntentSpec{
					IfName:         "eth1",
					NetworkName:    "networkA",
					DefaultGateway: "false",
					IpAddr:         "10.10.10.10",
				},
			},
			mockVal:   module.WorkloadIfIntent{},
			mockError: pkgerrors.New("does not exist"),
			client:    &mocks.WorkloadIfIntentManager{},
		}),

		Entry("fails due to db error", testCase{
			expectedCode: http.StatusInternalServerError,
			inputName:    "testinterfaceintent",
			inputReader: bytes.NewBuffer([]byte(`{
					"metadata": {
						"name": "testinterfaceintent",
						"description": "test interface intent",
						"userData1": "some user data 1",
						"userData2": "some user data 2"
					},
					"spec": {
						"interface": "eth1",
						"name": "networkA",
						"defaultGateway": "false",
						"ipAddress": "10.10.10.10"
					}
				}`)),
			inStruct: module.WorkloadIfIntent{
				Metadata: module.Metadata{
					Name:        "testinterfaceintent",
					Description: "test interface intent",
					UserData1:   "some user data 1",
					UserData2:   "some user data 2",
				},
				Spec: module.WorkloadIfIntentSpec{
					IfName:         "eth1",
					NetworkName:    "networkA",
					DefaultGateway: "false",
					IpAddr:         "10.10.10.10",
				},
			},
			mockVal:   module.WorkloadIfIntent{},
			mockError: pkgerrors.New("Creating DB Entry"),
			client:    &mocks.WorkloadIfIntentManager{},
		}),
	)

	DescribeTable("Get List WorkloadIfIntent tests",
		func(t testCase) {
			// set up client mock responses
			t.client.On("GetWorkloadIfIntents", "test-project", "test-compositeapp", "v1", "test-dig", "net-ctl-intent", "testworkloadintent").Return(t.mockVals, t.mockError)

			// make HTTP request
			request := httptest.NewRequest("GET", "/v2/projects/test-project/composite-apps/test-compositeapp/v1/deployment-intent-groups/test-dig/network-controller-intent/net-ctl-intent/workload-intents/testworkloadintent/interfaces", nil)
			resp := executeRequest(request, NewRouter(t.client))

			//Check returned code
			Expect(resp.StatusCode).To(Equal(t.expectedCode))

			//Check returned body
			got := []module.WorkloadIfIntent{}
			json.NewDecoder(resp.Body).Decode(&got)
			Expect(got).To(Equal(t.mockVals))
		},

		Entry("successful get all", testCase{
			expectedCode: http.StatusOK,
			mockError:    nil,
			mockVals: []module.WorkloadIfIntent{
				{
					Metadata: module.Metadata{
						Name:        "testinterfaceintent1",
						Description: "test interface intent",
						UserData1:   "some user data 1",
						UserData2:   "some user data 2",
					},
					Spec: module.WorkloadIfIntentSpec{
						IfName:         "eth1",
						NetworkName:    "networkA",
						DefaultGateway: "false",
						IpAddr:         "10.10.10.10",
					},
				},
				{
					Metadata: module.Metadata{
						Name:        "testinterfaceintent2",
						Description: "test interface intent",
						UserData1:   "some user data 1",
						UserData2:   "some user data 2",
					},
					Spec: module.WorkloadIfIntentSpec{
						IfName:         "eth2",
						NetworkName:    "networkB",
						DefaultGateway: "false",
						IpAddr:         "10.10.20.10",
					},
				},
			},
			client: &mocks.WorkloadIfIntentManager{},
		}),

		Entry("fails due to not found", testCase{
			expectedCode: http.StatusNotFound,
			mockError:    pkgerrors.New("db Find error"),
			mockVals:     []module.WorkloadIfIntent{},
			client:       &mocks.WorkloadIfIntentManager{},
		}),

		Entry("fails due to some other backend error", testCase{
			expectedCode: http.StatusInternalServerError,
			mockError:    pkgerrors.New("backend error"),
			mockVals:     []module.WorkloadIfIntent{},
			client:       &mocks.WorkloadIfIntentManager{},
		}),
	)

	DescribeTable("Get WorkloadIfIntent tests",
		func(t testCase) {
			// set up client mock responses
			t.client.On("GetWorkloadIfIntent", t.inputName, "test-project", "test-compositeapp", "v1", "test-dig", "net-ctl-intent", "testworkloadintent").Return(t.mockVal, t.mockError)

			// make HTTP request
			request := httptest.NewRequest("GET", "/v2/projects/test-project/composite-apps/test-compositeapp/v1/deployment-intent-groups/test-dig/network-controller-intent/net-ctl-intent/workload-intents/testworkloadintent/interfaces/"+t.inputName, nil)
			resp := executeRequest(request, NewRouter(t.client))

			//Check returned code
			Expect(resp.StatusCode).To(Equal(t.expectedCode))

			//Check returned body
			got := module.WorkloadIfIntent{}
			json.NewDecoder(resp.Body).Decode(&got)
			Expect(got).To(Equal(t.mockVal))
		},

		Entry("successful get", testCase{
			inputName:    "testworkloadintent",
			expectedCode: http.StatusOK,
			mockError:    nil,
			mockVal: module.WorkloadIfIntent{
				Metadata: module.Metadata{
					Name:        "testinterfaceintent",
					Description: "test interface intent",
					UserData1:   "some user data 1",
					UserData2:   "some user data 2",
				},
				Spec: module.WorkloadIfIntentSpec{
					IfName:         "eth1",
					NetworkName:    "networkA",
					DefaultGateway: "false",
					IpAddr:         "10.10.10.10",
				},
			},
			client: &mocks.WorkloadIfIntentManager{},
		}),

		Entry("fails due to not found", testCase{
			inputName:    "testworkloadintent",
			expectedCode: http.StatusNotFound,
			mockError:    pkgerrors.New("db Find error"),
			mockVal:      module.WorkloadIfIntent{},
			client:       &mocks.WorkloadIfIntentManager{},
		}),

		Entry("fails due to some other backend error", testCase{
			inputName:    "testworkloadintent",
			expectedCode: http.StatusInternalServerError,
			mockError:    pkgerrors.New("backend error"),
			mockVal:      module.WorkloadIfIntent{},
			client:       &mocks.WorkloadIfIntentManager{},
		}),
	)

	DescribeTable("Delete WorkloadIfIntent tests",
		func(t testCase) {
			// set up client mock responses
			t.client.On("DeleteWorkloadIfIntent", t.inputName, "test-project", "test-compositeapp", "v1", "test-dig", "net-ctl-intent", "testworkloadintent").Return(t.mockError)

			// make HTTP request
			request := httptest.NewRequest("DELETE", "/v2/projects/test-project/composite-apps/test-compositeapp/v1/deployment-intent-groups/test-dig/network-controller-intent/net-ctl-intent/workload-intents/testworkloadintent/interfaces/"+t.inputName, nil)
			resp := executeRequest(request, NewRouter(t.client))

			//Check returned code
			Expect(resp.StatusCode).To(Equal(t.expectedCode))

			//Check returned body
			got := module.WorkloadIfIntent{}
			json.NewDecoder(resp.Body).Decode(&got)
			Expect(got).To(Equal(t.mockVal))
		},

		Entry("successful delete", testCase{
			inputName:    "testinterfaceintent",
			expectedCode: http.StatusNoContent,
			mockError:    nil,
			client:       &mocks.WorkloadIfIntentManager{},
		}),

		Entry("fails due to not found", testCase{
			inputName:    "testinterfaceintent",
			expectedCode: http.StatusNotFound,
			mockError:    pkgerrors.New("db Remove error - not found"),
			client:       &mocks.WorkloadIfIntentManager{},
		}),

		Entry("fails due to a conflict", testCase{
			inputName:    "testinterfaceintent",
			expectedCode: http.StatusConflict,
			mockError:    pkgerrors.New("db Remove error - conflict"),
			client:       &mocks.WorkloadIfIntentManager{},
		}),

		Entry("fails due to other backend error", testCase{
			inputName:    "testinterfaceintent",
			expectedCode: http.StatusInternalServerError,
			mockError:    pkgerrors.New("db Remove error - general"),
			client:       &mocks.WorkloadIfIntentManager{},
		}),
	)
})
