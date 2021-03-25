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
	"github.com/open-ness/EMCO/src/dtc/api/mocks"
	controller "github.com/open-ness/EMCO/src/orchestrator/pkg/module/controller"
	mtypes "github.com/open-ness/EMCO/src/orchestrator/pkg/module/types"
	pkgerrors "github.com/pkg/errors"

)

func init() {
	controllerJSONFile = "../json-schemas/controller.json"
}

var _ = Describe("Controllerhandler", func() {

	type testCase struct {
		inputName    string
		inputReader  io.Reader
		inStruct     controller.Controller
		mockError    error
		mockVal      controller.Controller
		mockVals     []controller.Controller
		expectedCode int
		client	     *mocks.ControllerManager
	}

	DescribeTable("Create Controller tests",
		func(t testCase) {
			// set up client mock responses

			t.client.On("CreateController", t.inStruct, false).Return(t.mockVal, t.mockError)

			// make HTTP request
			request := httptest.NewRequest("POST", "/v2/dtc-controllers", t.inputReader)
			resp := executeRequest(request, NewRouter(t.client))

			//Check returned code
			Expect(resp.StatusCode).To(Equal(t.expectedCode))

			//Check returned body
			got := controller.Controller{}
			json.NewDecoder(resp.Body).Decode(&got)
			Expect(got).To(Equal(t.mockVal))
		},

		Entry("successful create", testCase{
			expectedCode: http.StatusCreated,
			inputReader: bytes.NewBuffer([]byte(`{
				"metadata": {
					"name": "dtccontroller",
					"description": "test dtc controller api",
					"userData1": "some user data 1",
					"userData2": "some user data 2"
					},
				"spec": {
					"host":     "test-host",
					"port":     8888,
					"type":     "action",
					"priority": 1
				}
			}`)),
			inStruct: controller.Controller{
				Metadata: mtypes.Metadata{
					Name:	     "dtccontroller",
					Description: "test dtc controller api",
					UserData1:   "some user data 1",
					UserData2:   "some user data 2",
				},
				Spec: controller.ControllerSpec{
					Host:	  "test-host",
					Port:	  8888,
					Type:	  "action",
					Priority: 1,
				},
			},
			mockError: nil,
			mockVal: controller.Controller{
				Metadata: mtypes.Metadata{
					Name:	     "dtccontroller",
					Description: "test dtc controller api",
					UserData1:   "some user data 1",
					UserData2:   "some user data 2",
				},
				Spec: controller.ControllerSpec{
					Host:	  "test-host",
					Port:	  8888,
					Type:	  "action",
					Priority: 1,
				},
			},
			client: &mocks.ControllerManager{},
		}),
		Entry("fails due to empty body", testCase{
			expectedCode: http.StatusBadRequest,
			inStruct:     controller.Controller{},
			mockError:    nil,
			mockVal:      controller.Controller{},
			client:       &mocks.ControllerManager{},
		}),

		Entry("fails due missing name", testCase{
			expectedCode: http.StatusBadRequest,
			inputReader: bytes.NewBuffer([]byte(`{
				"metadata": {
					"description": "test dtc controller api",
					"userData1": "some user data 1",
					"userData2": "some user data 2"
					},
				"spec": {
					"host":     "test-host",
					"port":     8888,
					"type":     "action",
					"priority": 1
				}
			}`)),
			inStruct:  controller.Controller{},
			mockError: nil,
			mockVal:      controller.Controller{},
			client:    &mocks.ControllerManager{},
		}),
		Entry("fails due missing port", testCase{
			expectedCode: http.StatusBadRequest,
			inputReader: bytes.NewBuffer([]byte(`{
				"metadata": {
					"name": "dtccontroller",
					"description": "test dtc controller api",
					"userData1": "some user data 1",
					"userData2": "some user data 2"
					},
				"spec": {
					"host":     "test-host",
					"type":     "action",
					"priority": 1
				}
			}`)),
			inStruct:  controller.Controller{},
			mockError: nil,
			mockVal:      controller.Controller{},
			client:    &mocks.ControllerManager{},
		}),

		Entry("fails due to invalid port", testCase{
			expectedCode: http.StatusBadRequest,
			inputReader: bytes.NewBuffer([]byte(`{
				"metadata": {
					"name": "dtccontroller",
					"description": "test dtc controller api",
					"userData1": "some user data 1",
					"userData2": "some user data 2"
					},
				"spec": {
					"host":     "test-host",
					"port":     760000,
					"type":     "action",
					"priority": 1
				}
			}`)),
			inStruct:  controller.Controller{},
			mockError: nil,
			mockVal:      controller.Controller{},
			client:    &mocks.ControllerManager{},
		}),

		Entry("fails due to invalid priority", testCase{
			expectedCode: http.StatusBadRequest,
			inputReader: bytes.NewBuffer([]byte(`{
				"metadata": {
					"name": "dtccontroller",
					"description": "test dtc controller api",
					"userData1": "some user data 1",
					"userData2": "some user data 2"
					},
				"spec": {
					"host":     "test-host",
					"port":     8888,
					"type":     "action",
					"priority": 0
				}
			}`)),
			inStruct:  controller.Controller{},
			mockError: nil,
			client:    &mocks.ControllerManager{},
		}),
		Entry("fails due to other json validation error", testCase{
			// name field has an '=' character
			expectedCode: http.StatusBadRequest,
			inputReader: bytes.NewBuffer([]byte(`{
				"metadata": {
					"name": "test=dtccontroller",
					"description": "test dtc controller api",
					"userData1": "some user data 1",
					"userData2": "some user data 2"
					},
				"spec": {
					"host":     "test-host",
					"port":     8888,
					"type":     "action",
					"priority": 1
				}
			}`)),
			inStruct:  controller.Controller{},
			mockError: nil,
			mockVal:   controller.Controller{},
			client:    &mocks.ControllerManager{},
		}),
		Entry("fails due to entry already exists", testCase{
			expectedCode: http.StatusConflict,
			inputReader: bytes.NewBuffer([]byte(`{
				"metadata": {
					"name": "dtccontroller",
					"description": "test dtc controller api",
					"userData1": "some user data 1",
					"userData2": "some user data 2"
					},
				"spec": {
					"host":     "test-host",
					"port":     8888,
					"type":     "action",
					"priority": 1
				}
			}`)),
			inStruct: controller.Controller{
				Metadata: mtypes.Metadata{
					Name:	     "dtccontroller",
					Description: "test dtc controller api",
					UserData1:   "some user data 1",
					UserData2:   "some user data 2",
				},
				Spec: controller.ControllerSpec{
					Host:	  "test-host",
					Port:	  8888,
					Type:	  "action",
					Priority: 1,
				},
			},
			mockError: pkgerrors.New("Controller already exists"),
			mockVal: controller.Controller{},
			client: &mocks.ControllerManager{},
		}),


	)
	DescribeTable("Put Controller tests",
		func(t testCase) {
			// set up client mock responses

			t.client.On("CreateController", t.inStruct, true).Return(t.mockVal, t.mockError)

			// make HTTP request
			request := httptest.NewRequest("PUT", "/v2/dtc-controllers/" + t.inputName, t.inputReader)
			resp := executeRequest(request, NewRouter(t.client))

			//Check returned code
			Expect(resp.StatusCode).To(Equal(t.expectedCode))

			//Check returned body
			got := controller.Controller{}
			json.NewDecoder(resp.Body).Decode(&got)
			Expect(got).To(Equal(t.mockVal))
		},

		Entry("successful put", testCase{
			inputName: "dtccontroller",
			expectedCode: http.StatusCreated,
			inputReader: bytes.NewBuffer([]byte(`{
				"metadata": {
					"name": "dtccontroller",
					"description": "test dtc controller api",
					"userData1": "some user data 1",
					"userData2": "some user data 2"
					},
				"spec": {
					"host":     "test-host",
					"port":     8888,
					"type":     "action",
					"priority": 1
				}
			}`)),
			inStruct: controller.Controller{
				Metadata: mtypes.Metadata{
					Name:	     "dtccontroller",
					Description: "test dtc controller api",
					UserData1:   "some user data 1",
					UserData2:   "some user data 2",
				},
				Spec: controller.ControllerSpec{
					Host:	  "test-host",
					Port:	  8888,
					Type:	  "action",
					Priority: 1,
				},
			},
			mockError: nil,
			mockVal: controller.Controller{
				Metadata: mtypes.Metadata{
					Name:	     "dtccontroller",
					Description: "test dtc controller api",
					UserData1:   "some user data 1",
					UserData2:   "some user data 2",
				},
				Spec: controller.ControllerSpec{
					Host:	  "test-host",
					Port:	  8888,
					Type:	  "action",
					Priority: 1,
				},
			},
			client: &mocks.ControllerManager{},
		}),
		Entry("fails due to empty body", testCase{
			inputName: "dtccontroller",
			expectedCode: http.StatusBadRequest,
			inStruct:     controller.Controller{},
			mockError:    nil,
			mockVal:      controller.Controller{},
			client:       &mocks.ControllerManager{},
		}),

		Entry("fails due missing name", testCase{
			inputName: "dtccontroller",
			expectedCode: http.StatusBadRequest,
			inputReader: bytes.NewBuffer([]byte(`{
				"metadata": {
					"description": "test dtc controller api",
					"userData1": "some user data 1",
					"userData2": "some user data 2"
					},
				"spec": {
					"host":     "test-host",
					"port":     8888,
					"type":     "action",
					"priority": 1
				}
			}`)),
			inStruct:  controller.Controller{},
			mockError: nil,
			mockVal:      controller.Controller{},
			client:    &mocks.ControllerManager{},
		}),

		Entry("fails due to invalid port", testCase{
			inputName: "dtccontroller",
			expectedCode: http.StatusBadRequest,
			inputReader: bytes.NewBuffer([]byte(`{
				"metadata": {
					"name": "dtccontroller",
					"description": "test dtc controller api",
					"userData1": "some user data 1",
					"userData2": "some user data 2"
					},
				"spec": {
					"host":     "test-host",
					"port":     760000,
					"type":     "action",
					"priority": 1
				}
			}`)),
			inStruct:  controller.Controller{},
			mockError: nil,
			mockVal:      controller.Controller{},
			client:    &mocks.ControllerManager{},
		}),

		Entry("fails due to invalid priority", testCase{
			inputName: "dtccontroller",
			expectedCode: http.StatusBadRequest,
			inputReader: bytes.NewBuffer([]byte(`{
				"metadata": {
					"name": "dtccontroller",
					"description": "test dtc controller api",
					"userData1": "some user data 1",
					"userData2": "some user data 2"
					},
				"spec": {
					"host":     "test-host",
					"port":     8888,
					"type":     "action",
					"priority": 0
				}
			}`)),
			inStruct:  controller.Controller{},
			mockError: nil,
			client:    &mocks.ControllerManager{},
		}),
		Entry("fails due to other json validation error", testCase{
			// name field has an '=' character
			inputName: "dtccontroller",
			expectedCode: http.StatusBadRequest,
			inputReader: bytes.NewBuffer([]byte(`{
				"metadata": {
					"name": "test=dtccontroller",
					"description": "test dtc controller api",
					"userData1": "some user data 1",
					"userData2": "some user data 2"
					},
				"spec": {
					"host":     "test-host",
					"port":     8888,
					"type":     "action",
					"priority": 1
				}
			}`)),
			inStruct:  controller.Controller{},
			mockError: nil,
			mockVal:   controller.Controller{},
			client:    &mocks.ControllerManager{},
		}),
		Entry("fails due to other json body decoding error", testCase{
			// extra comma at the end of the userData2 line
			inputName: "dtccontroller",
			expectedCode: http.StatusUnprocessableEntity,
			inputReader: bytes.NewBuffer([]byte(`{
				"metadata": {
					"name": "dtccontroller",
					"description": "test dtc controller api",
					"userData1": "some user data 1",
					"userData2": "some user data 2",
					},
				"spec": {
					"host":     "test-host",
					"port":     8888,
					"type":     "action",
					"priority": 1
				}
			}`)),
			inStruct:  controller.Controller{},
			mockError: nil,
			mockVal:   controller.Controller{},
			client:    &mocks.ControllerManager{},
		}),
		Entry("fails due to mismatched name", testCase{
			inputName: "dtccontroller",
			expectedCode: http.StatusBadRequest,
			inputReader: bytes.NewBuffer([]byte(`{
				"metadata": {
					"name": "dtccontrollerXYZ",
					"description": "test dtc controller api",
					"userData1": "some user data 1",
					"userData2": "some user data 2"
					},
				"spec": {
					"host":     "test-host",
					"port":     8888,
					"type":     "action",
					"priority": 1
				}
			}`)),
			inStruct:  controller.Controller{},
			mockError: nil,
			mockVal:   controller.Controller{},
			client:    &mocks.ControllerManager{},
		}),


	)
	DescribeTable("Get Controllers tests",
		func(t testCase) {
			// set up client mock responses

			t.client.On("GetControllers").Return(t.mockVals, t.mockError)

			// make HTTP request
			request := httptest.NewRequest("GET", "/v2/dtc-controllers", nil)
			resp := executeRequest(request, NewRouter(t.client))

			//Check returned code
			Expect(resp.StatusCode).To(Equal(t.expectedCode))

			//Check returned body
			got := []controller.Controller{}
			json.NewDecoder(resp.Body).Decode(&got)
			Expect(got).To(Equal(t.mockVals))
		},

		Entry("successful get", testCase{
			expectedCode: http.StatusOK,
			mockError: nil,
			mockVals: []controller.Controller{
				{
					Metadata: mtypes.Metadata{
						Name:	     "dtccontroller1",
						Description: "test dtc controller api",
						UserData1:   "some user data 1",
						UserData2:   "some user data 2",
					},
					Spec: controller.ControllerSpec{
						Host:	  "test-host1",
						Port:	  8888,
						Type:	  "action",
						Priority: 1,
					},
				},
				{
					Metadata: mtypes.Metadata{
						Name:	     "dtccontroller2",
						Description: "test dtc controller api",
						UserData1:   "some user data 1",
						UserData2:   "some user data 2",
					},
					Spec: controller.ControllerSpec{
						Host:	  "test-hosti2",
						Port:	  4444,
						Type:	  "action",
						Priority: 2,
					},

				},
			},
			client: &mocks.ControllerManager{},
		}),
		Entry("fails due to not found", testCase{
			expectedCode: http.StatusNotFound,
			mockError: pkgerrors.New("db Find error"),
			mockVals:  []controller.Controller{},
			client:    &mocks.ControllerManager{},
		}),
		Entry("fails due to some other backend error", testCase{
			expectedCode: http.StatusInternalServerError,
			mockError: pkgerrors.New("backend error"),
			mockVals:  []controller.Controller{},
			client:    &mocks.ControllerManager{},
		}),
	)
	DescribeTable("Get Controller tests",
		func(t testCase) {
			// set up client mock responses

			t.client.On("GetController", "dtccontroller").Return(t.mockVal, t.mockError)

			// make HTTP request
			request := httptest.NewRequest("GET", "/v2/dtc-controllers/"+t.inputName, nil)
			resp := executeRequest(request, NewRouter(t.client))

			//Check returned code
			Expect(resp.StatusCode).To(Equal(t.expectedCode))

			//Check returned body
			got := controller.Controller{}
			json.NewDecoder(resp.Body).Decode(&got)
			Expect(got).To(Equal(t.mockVal))
		},

		Entry("successful get", testCase{
			inputName:    "dtccontroller",
			expectedCode: http.StatusOK,
			mockError: nil,
			mockVal: controller.Controller{
				Metadata: mtypes.Metadata{
					Name:	     "dtccontroller",
					Description: "test dtc controller api",
					UserData1:   "some user data 1",
					UserData2:   "some user data 2",
				},
				Spec: controller.ControllerSpec{
					Host:	  "test-host",
					Port:	  8888,
					Type:	  "action",
					Priority: 1,
				},
			},
			client: &mocks.ControllerManager{},
		}),
		Entry("fails due to not found", testCase{
			inputName:    "dtccontroller",
			expectedCode: http.StatusNotFound,
			mockError: pkgerrors.New("db Find error"),
			mockVal:  controller.Controller{},
			client:    &mocks.ControllerManager{},
		}),
		Entry("fails due to some other backend error", testCase{
			inputName:    "dtccontroller",
			expectedCode: http.StatusInternalServerError,
			mockError: pkgerrors.New("backend error"),
			mockVal:  controller.Controller{},
			client:    &mocks.ControllerManager{},
		}),
	)
       DescribeTable("DELETE Controller tests",
		func(t testCase) {
			// set up client mock responses

			t.client.On("DeleteController", t.inputName).Return(t.mockError)

			// make HTTP request
			request := httptest.NewRequest("DELETE", "/v2/dtc-controllers/"+t.inputName, nil)
			resp := executeRequest(request, NewRouter(t.client))

			//Check returned code
			Expect(resp.StatusCode).To(Equal(t.expectedCode))

			//Check returned body
			got := controller.Controller{}
			json.NewDecoder(resp.Body).Decode(&got)
			Expect(got).To(Equal(t.mockVal))
		},
		Entry("successful delete", testCase{
			inputName:    "dtccontroller",
			expectedCode: http.StatusNoContent,
			mockError:    nil,
			mockVal:      controller.Controller{},
			client:       &mocks.ControllerManager{},
		}),
		Entry("fails due to not found", testCase{
			inputName:    "dtccontroller",
			expectedCode: http.StatusNotFound,
			mockError:    pkgerrors.New("db Remove error - not found"),
			mockVal:      controller.Controller{},
			client:       &mocks.ControllerManager{},
		}),
		Entry("fails due to conflict", testCase{
			inputName:    "dtccontroller",
			expectedCode: http.StatusConflict,
			mockError:    pkgerrors.New("db Remove error - conflict"),
			mockVal:      controller.Controller{},
			client:       &mocks.ControllerManager{},
		}),
		Entry("fails due to some other backend error", testCase{
			inputName:    "dtccontroller",
			expectedCode: http.StatusInternalServerError,
			mockError:    pkgerrors.New("db Remove error - general"),
			mockVal:      controller.Controller{},
			client:       &mocks.ControllerManager{},
		}),
	)
})
