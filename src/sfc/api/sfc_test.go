// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2021 Intel Corporation

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
	"github.com/open-ness/EMCO/src/sfc/api/mocks"
	"github.com/open-ness/EMCO/src/sfc/pkg/model"
	pkgerrors "github.com/pkg/errors"
)

func init() {
	sfcJSONFile = "../json-schemas/sfc.json"
}

var _ = Describe("Sfcintenthandler", func() {

	type testCase struct {
		inputName    string
		inputReader  io.Reader
		inStruct     model.SfcIntent
		mockError    error
		mockVal      model.SfcIntent
		mockVals     []model.SfcIntent
		expectedCode int
		client       *mocks.SfcIntentManager
	}

	DescribeTable("Create SfcIntent tests",
		func(t testCase) {
			// set up client mock responses
			t.client.On("CreateSfcIntent", t.inStruct, "test-project", "test-compositeapp", "v1", "test-dig", "net-ctl-intent", false).Return(t.mockVal, t.mockError)

			// make HTTP request
			request := httptest.NewRequest("POST", "/v2/projects/test-project/composite-apps/test-compositeapp/v1/deployment-intent-groups/test-dig/network-controller-intent/net-ctl-intent/network-chains", t.inputReader)
			resp := executeRequest(request, NewRouter(t.client))

			//Check returned code
			Expect(resp.StatusCode).To(Equal(t.expectedCode))

			//Check returned body
			got := model.SfcIntent{}
			json.NewDecoder(resp.Body).Decode(&got)
			Expect(got).To(Equal(t.mockVal))
		},

		Entry("successful create", testCase{
			expectedCode: http.StatusCreated,
			inputReader: bytes.NewBuffer([]byte(`{
				"metadata": {
					"name": "testsfcintent",
					"description": "test sfc intent",
					"userData1": "some user data 1",
					"userData2": "some user data 2"
				},
				"spec": {
					"chainType": "Routing",
				    "namespace": "chainspace",
				    "networkChain": "net=net0,app=a1,net=n1,app=a2,net=n2,app=a3,net=n3"
				}
			}`)),
			inStruct: model.SfcIntent{
				Metadata: model.Metadata{
					Name:        "testsfcintent",
					Description: "test sfc intent",
					UserData1:   "some user data 1",
					UserData2:   "some user data 2",
				},
				Spec: model.SfcIntentSpec{
					ChainType:    "Routing",
					Namespace:    "chainspace",
					NetworkChain: "net=net0,app=a1,net=n1,app=a2,net=n2,app=a3,net=n3",
				},
			},
			mockError: nil,
			mockVal: model.SfcIntent{
				Metadata: model.Metadata{
					Name:        "testsfcintent",
					Description: "test sfc intent",
					UserData1:   "some user data 1",
					UserData2:   "some user data 2",
				},
				Spec: model.SfcIntentSpec{
					ChainType:    "Routing",
					Namespace:    "chainspace",
					NetworkChain: "net=net0,app=a1,net=n1,app=a2,net=n2,app=a3,net=n3",
				},
			},
			client: &mocks.SfcIntentManager{},
		}),

		Entry("fails due to empty body", testCase{
			expectedCode: http.StatusBadRequest,
			inStruct:     model.SfcIntent{},
			mockError:    nil,
			mockVal:      model.SfcIntent{},
			client:       &mocks.SfcIntentManager{},
		}),

		Entry("fails due missing name", testCase{
			expectedCode: http.StatusBadRequest,
			inputReader: bytes.NewBuffer([]byte(`{
					"metadata": {
						"description": "test sfc intent",
						"userData1": "some user data 1",
						"userData2": "some user data 2"
					},
					"spec": {
						"chainType": "Routing",
						"namespace": "chainspace",
						"networkChain": "net=net0,app=a1,net=n1,app=a2,net=n2,app=a3,net=n3"
					}
				}`)),
			inStruct:  model.SfcIntent{},
			mockError: nil,
			client:    &mocks.SfcIntentManager{},
		}),

		Entry("fails due missing chain type", testCase{
			expectedCode: http.StatusBadRequest,
			inputReader: bytes.NewBuffer([]byte(`{
					"metadata": {
						"name": "testsfcintent",
						"description": "test sfc intent",
						"userData1": "some user data 1",
						"userData2": "some user data 2"
					},
					"spec": {
					    "namespace": "chainspace",
					    "networkChain": "net=net0,app=a1,net=n1,app=a2,net=n2,app=a3,net=n3"
					}
				}`)),
			inStruct:  model.SfcIntent{},
			mockError: nil,
			client:    &mocks.SfcIntentManager{},
		}),

		Entry("fails due missing network chain", testCase{
			expectedCode: http.StatusBadRequest,
			inputReader: bytes.NewBuffer([]byte(`{
					"metadata": {
						"name": "testsfcintent",
						"description": "test sfc intent",
						"userData1": "some user data 1",
						"userData2": "some user data 2"
					},
					"spec": {
						"chainType": "Routing",
						"namespace": "chainspace"
					}
				}`)),
			inStruct:  model.SfcIntent{},
			mockError: nil,
			client:    &mocks.SfcIntentManager{},
		}),

		Entry("fails due to json validation error", testCase{
			// name field has an '=' character
			expectedCode: http.StatusBadRequest,
			inputReader: bytes.NewBuffer([]byte(`{
					"metadata": {
						"name": "test=sfcintent",
						"description": "test sfc intent",
						"userData1": "some user data 1",
						"userData2": "some user data 2"
					},
					"spec": {
						"chainType": "Routing",
					    "namespace": "chainspace",
					    "networkChain": "net=net0,app=a1,net=n1,app=a2,net=n2,app=a3,net=n3"
					}
				}`)),
			inStruct:  model.SfcIntent{},
			mockError: nil,
			client:    &mocks.SfcIntentManager{},
		}),

		Entry("fails due to invalid networkChain content", testCase{
			// name field has an '=' character
			expectedCode: http.StatusBadRequest,
			inputReader: bytes.NewBuffer([]byte(`{
					"metadata": {
						"name": "test=sfcintent",
						"description": "test sfc intent",
						"userData1": "some user data 1",
						"userData2": "some user data 2"
					},
					"spec": {
						"chainType": "Routing",
					    "namespace": "chainspace",
					    "networkChain": "net=n1,app=a1"
					}
				}`)),
			inStruct:  model.SfcIntent{},
			mockError: nil,
			client:    &mocks.SfcIntentManager{},
		}),

		Entry("fails due to entry already exists", testCase{
			expectedCode: http.StatusConflict,
			inputReader: bytes.NewBuffer([]byte(`{
					"metadata": {
						"name": "testsfcintent",
						"description": "test sfc intent",
						"userData1": "some user data 1",
						"userData2": "some user data 2"
					},
					"spec": {
						"chainType": "Routing",
						"namespace": "chainspace",
						"networkChain": "net=net0,app=a1,net=n1,app=a2,net=n2,app=a3,net=n3"
					}
				}`)),
			inStruct: model.SfcIntent{
				Metadata: model.Metadata{
					Name:        "testsfcintent",
					Description: "test sfc intent",
					UserData1:   "some user data 1",
					UserData2:   "some user data 2",
				},
				Spec: model.SfcIntentSpec{
					ChainType:    "Routing",
					Namespace:    "chainspace",
					NetworkChain: "net=net0,app=a1,net=n1,app=a2,net=n2,app=a3,net=n3",
				},
			},
			mockVal:   model.SfcIntent{},
			mockError: pkgerrors.New("SFC Intent already exists"),
			client:    &mocks.SfcIntentManager{},
		}),

		Entry("fails due to network control intent does not exist", testCase{
			expectedCode: http.StatusNotFound,
			inputReader: bytes.NewBuffer([]byte(`{
					"metadata": {
						"name": "testsfcintent",
						"description": "test sfc intent",
						"userData1": "some user data 1",
						"userData2": "some user data 2"
					},
					"spec": {
						"chainType": "Routing",
						"namespace": "chainspace",
						"networkChain": "net=net0,app=a1,net=n1,app=a2,net=n2,app=a3,net=n3"
					}
				}`)),
			inStruct: model.SfcIntent{
				Metadata: model.Metadata{
					Name:        "testsfcintent",
					Description: "test sfc intent",
					UserData1:   "some user data 1",
					UserData2:   "some user data 2",
				},
				Spec: model.SfcIntentSpec{
					ChainType:    "Routing",
					Namespace:    "chainspace",
					NetworkChain: "net=net0,app=a1,net=n1,app=a2,net=n2,app=a3,net=n3",
				},
			},
			mockVal:   model.SfcIntent{},
			mockError: pkgerrors.New("does not exist"),
			client:    &mocks.SfcIntentManager{},
		}),

		Entry("fails due to db error", testCase{
			expectedCode: http.StatusInternalServerError,
			inputReader: bytes.NewBuffer([]byte(`{
					"metadata": {
						"name": "testsfcintent",
						"description": "test sfc intent",
						"userData1": "some user data 1",
						"userData2": "some user data 2"
					},
					"spec": {
						"chainType": "Routing",
						"namespace": "chainspace",
						"networkChain": "net=net0,app=a1,net=n1,app=a2,net=n2,app=a3,net=n3"
					}
				}`)),
			inStruct: model.SfcIntent{
				Metadata: model.Metadata{
					Name:        "testsfcintent",
					Description: "test sfc intent",
					UserData1:   "some user data 1",
					UserData2:   "some user data 2",
				},
				Spec: model.SfcIntentSpec{
					ChainType:    "Routing",
					Namespace:    "chainspace",
					NetworkChain: "net=net0,app=a1,net=n1,app=a2,net=n2,app=a3,net=n3",
				},
			},
			mockVal:   model.SfcIntent{},
			mockError: pkgerrors.New("Creating DB Entry"),
			client:    &mocks.SfcIntentManager{},
		}),
	)

	DescribeTable("Put SfcIntent tests",
		func(t testCase) {
			// set up client mock responses
			t.client.On("CreateSfcIntent", t.inStruct, "test-project", "test-compositeapp", "v1", "test-dig", "net-ctl-intent", true).Return(t.mockVal, t.mockError)

			// make HTTP request
			request := httptest.NewRequest("PUT", "/v2/projects/test-project/composite-apps/test-compositeapp/v1/deployment-intent-groups/test-dig/network-controller-intent/net-ctl-intent/network-chains/"+t.inputName, t.inputReader)
			resp := executeRequest(request, NewRouter(t.client))

			//Check returned code
			Expect(resp.StatusCode).To(Equal(t.expectedCode))

			//Check returned body
			got := model.SfcIntent{}
			json.NewDecoder(resp.Body).Decode(&got)
			Expect(got).To(Equal(t.mockVal))
		},

		Entry("successful put", testCase{
			expectedCode: http.StatusCreated,
			inputName:    "testsfcintent",
			inputReader: bytes.NewBuffer([]byte(`{
				"metadata": {
					"name": "testsfcintent",
					"description": "test sfc intent",
					"userData1": "some user data 1",
					"userData2": "some user data 2"
				},
				"spec": {
					"chainType": "Routing",
					"namespace": "chainspace",
					"networkChain": "net=net0,app=a1,net=n1,app=a2,net=n2,app=a3,net=n3"
				}
			}`)),
			inStruct: model.SfcIntent{
				Metadata: model.Metadata{
					Name:        "testsfcintent",
					Description: "test sfc intent",
					UserData1:   "some user data 1",
					UserData2:   "some user data 2",
				},
				Spec: model.SfcIntentSpec{
					ChainType:    "Routing",
					Namespace:    "chainspace",
					NetworkChain: "net=net0,app=a1,net=n1,app=a2,net=n2,app=a3,net=n3",
				},
			},
			mockError: nil,
			mockVal: model.SfcIntent{
				Metadata: model.Metadata{
					Name:        "testsfcintent",
					Description: "test sfc intent",
					UserData1:   "some user data 1",
					UserData2:   "some user data 2",
				},
				Spec: model.SfcIntentSpec{
					ChainType:    "Routing",
					Namespace:    "chainspace",
					NetworkChain: "net=net0,app=a1,net=n1,app=a2,net=n2,app=a3,net=n3",
				},
			},
			client: &mocks.SfcIntentManager{},
		}),

		Entry("fails due to empty body", testCase{
			expectedCode: http.StatusBadRequest,
			inputName:    "testsfcintent",
			inStruct:     model.SfcIntent{},
			mockError:    nil,
			mockVal:      model.SfcIntent{},
			client:       &mocks.SfcIntentManager{},
		}),

		Entry("fails due missing name", testCase{
			expectedCode: http.StatusBadRequest,
			inputName:    "testsfcintent",
			inputReader: bytes.NewBuffer([]byte(`{
					"metadata": {
						"description": "test sfc intent",
						"userData1": "some user data 1",
						"userData2": "some user data 2"
					},
					"spec": {
						"chainType": "Routing",
						"namespace": "chainspace",
						"networkChain": "net=net0,app=a1,net=n1,app=a2,net=n2,app=a3,net=n3"
					}
				}`)),
			inStruct:  model.SfcIntent{},
			mockError: nil,
			client:    &mocks.SfcIntentManager{},
		}),

		Entry("fails due missing type", testCase{
			expectedCode: http.StatusBadRequest,
			inputName:    "testsfcintent",
			inputReader: bytes.NewBuffer([]byte(`{
					"metadata": {
						"name": "testsfcintent",
						"description": "test sfc intent",
						"userData1": "some user data 1",
						"userData2": "some user data 2"
					},
					"spec": {
						"namespace": "chainspace",
						"networkChain": "net=net0,app=a1,net=n1,app=a2,net=n2,app=a3,net=n3"
					}
				}`)),
			inStruct:  model.SfcIntent{},
			mockError: nil,
			client:    &mocks.SfcIntentManager{},
		}),

		Entry("fails due to other json validation error", testCase{
			// name field has an '=' character
			expectedCode: http.StatusBadRequest,
			inputName:    "testsfcintent",
			inputReader: bytes.NewBuffer([]byte(`{
					"metadata": {
						"name": "test=sfcintent",
						"description": "test sfc intent",
						"userData1": "some user data 1",
						"userData2": "some user data 2"
					},
					"spec": {
						"chainType": "Routing",
						"namespace": "chainspace",
						"networkChain": "net=net0,app=a1,net=n1,app=a2,net=n2,app=a3,net=n3"
					}
				}`)),
			inStruct:  model.SfcIntent{},
			mockError: nil,
			client:    &mocks.SfcIntentManager{},
		}),

		Entry("fails due to json body decoding error", testCase{
			// extra comma at the end of the userData2 line
			expectedCode: http.StatusUnprocessableEntity,
			inputName:    "testsfcintent",
			inputReader: bytes.NewBuffer([]byte(`{
					"metadata": {
						"name": "testsfcintent",
						"description": "test sfc intent",
						"userData1": "some user data 1",
						"userData2": "some user data 2",
					},
					"spec": {
						"chainType": "Routing",
						"namespace": "chainspace",
						"networkChain": "net=net0,app=a1,net=n1,app=a2,net=n2,app=a3,net=n3"
					}
				}`)),
			inStruct:  model.SfcIntent{},
			mockError: nil,
			client:    &mocks.SfcIntentManager{},
		}),

		Entry("fails due to name mismatch", testCase{
			expectedCode: http.StatusBadRequest,
			inputName:    "testsfcintent",
			inputReader: bytes.NewBuffer([]byte(`{
					"metadata": {
						"name": "testsfcintentABC",
						"description": "test sfc intent",
						"userData1": "some user data 1",
						"userData2": "some user data 2"
					},
					"spec": {
						"chainType": "Routing",
						"namespace": "chainspace",
						"networkChain": "net=net0,app=a1,net=n1,app=a2,net=n2,app=a3,net=n3"
					}
				}`)),
			inStruct: model.SfcIntent{
				Metadata: model.Metadata{
					Name:        "testsfcintent",
					Description: "test sfc intent",
					UserData1:   "some user data 1",
					UserData2:   "some user data 2",
				},
				Spec: model.SfcIntentSpec{
					ChainType:    "Routing",
					Namespace:    "chainspace",
					NetworkChain: "net=net0,app=a1,net=n1,app=a2,net=n2,app=a3,net=n3",
				},
			},
			mockVal:   model.SfcIntent{},
			mockError: pkgerrors.New("SfcIntent already exists"),
			client:    &mocks.SfcIntentManager{},
		}),

		Entry("fails due to network control intent does not exist", testCase{
			expectedCode: http.StatusNotFound,
			inputName:    "testsfcintent",
			inputReader: bytes.NewBuffer([]byte(`{
					"metadata": {
						"name": "testsfcintent",
						"description": "test sfc intent",
						"userData1": "some user data 1",
						"userData2": "some user data 2"
					},
					"spec": {
						"chainType": "Routing",
						"namespace": "chainspace",
						"networkChain": "net=net0,app=a1,net=n1,app=a2,net=n2,app=a3,net=n3"
					}
				}`)),
			inStruct: model.SfcIntent{
				Metadata: model.Metadata{
					Name:        "testsfcintent",
					Description: "test sfc intent",
					UserData1:   "some user data 1",
					UserData2:   "some user data 2",
				},
				Spec: model.SfcIntentSpec{
					ChainType:    "Routing",
					Namespace:    "chainspace",
					NetworkChain: "net=net0,app=a1,net=n1,app=a2,net=n2,app=a3,net=n3",
				},
			},
			mockVal:   model.SfcIntent{},
			mockError: pkgerrors.New("does not exist"),
			client:    &mocks.SfcIntentManager{},
		}),

		Entry("fails due to db error", testCase{
			expectedCode: http.StatusInternalServerError,
			inputName:    "testsfcintent",
			inputReader: bytes.NewBuffer([]byte(`{
					"metadata": {
						"name": "testsfcintent",
						"description": "test sfc intent",
						"userData1": "some user data 1",
						"userData2": "some user data 2"
					},
					"spec": {
						"chainType": "Routing",
						"namespace": "chainspace",
						"networkChain": "net=net0,app=a1,net=n1,app=a2,net=n2,app=a3,net=n3"
					}
				}`)),
			inStruct: model.SfcIntent{
				Metadata: model.Metadata{
					Name:        "testsfcintent",
					Description: "test sfc intent",
					UserData1:   "some user data 1",
					UserData2:   "some user data 2",
				},
				Spec: model.SfcIntentSpec{
					ChainType:    "Routing",
					Namespace:    "chainspace",
					NetworkChain: "net=net0,app=a1,net=n1,app=a2,net=n2,app=a3,net=n3",
				},
			},
			mockVal:   model.SfcIntent{},
			mockError: pkgerrors.New("Creating DB Entry"),
			client:    &mocks.SfcIntentManager{},
		}),
	)

	DescribeTable("Get List SfcIntent tests",
		func(t testCase) {
			// set up client mock responses
			t.client.On("GetAllSfcIntents", "test-project", "test-compositeapp", "v1", "test-dig", "net-ctl-intent").Return(t.mockVals, t.mockError)

			// make HTTP request
			request := httptest.NewRequest("GET", "/v2/projects/test-project/composite-apps/test-compositeapp/v1/deployment-intent-groups/test-dig/network-controller-intent/net-ctl-intent/network-chains", nil)
			resp := executeRequest(request, NewRouter(t.client))

			//Check returned code
			Expect(resp.StatusCode).To(Equal(t.expectedCode))

			//Check returned body
			got := []model.SfcIntent{}
			json.NewDecoder(resp.Body).Decode(&got)
			Expect(got).To(Equal(t.mockVals))
		},

		Entry("successful get", testCase{
			expectedCode: http.StatusOK,
			mockError:    nil,
			mockVals: []model.SfcIntent{
				{
					Metadata: model.Metadata{
						Name:        "testsfcintent1",
						Description: "test sfc intent",
						UserData1:   "some user data 1",
						UserData2:   "some user data 2",
					},
					Spec: model.SfcIntentSpec{
						ChainType:    "Routing",
						Namespace:    "chainspace",
						NetworkChain: "net=net0,app=a1,net=n1,app=a2,net=n2,app=a3,net=n3",
					},
				},
				{
					Metadata: model.Metadata{
						Name:        "testsfcintent2",
						Description: "test sfc intent",
						UserData1:   "some user data 1",
						UserData2:   "some user data 2",
					},
					Spec: model.SfcIntentSpec{
						ChainType:    "Routing",
						Namespace:    "chainspace",
						NetworkChain: "net=net0,app=a1,net=n1,app=a2,net=n2,app=a3,net=n3",
					},
				},
			},
			client: &mocks.SfcIntentManager{},
		}),

		Entry("fails due to not found", testCase{
			expectedCode: http.StatusNotFound,
			mockError:    pkgerrors.New("db Find error"),
			mockVals:     []model.SfcIntent{},
			client:       &mocks.SfcIntentManager{},
		}),

		Entry("fails due to some other backend error", testCase{
			expectedCode: http.StatusInternalServerError,
			mockError:    pkgerrors.New("backend error"),
			mockVals:     []model.SfcIntent{},
			client:       &mocks.SfcIntentManager{},
		}),
	)

	DescribeTable("Get SfcIntent tests",
		func(t testCase) {
			// set up client mock responses
			t.client.On("GetSfcIntent", t.inputName, "test-project", "test-compositeapp", "v1", "test-dig", "net-ctl-intent").Return(t.mockVal, t.mockError)

			// make HTTP request
			request := httptest.NewRequest("GET", "/v2/projects/test-project/composite-apps/test-compositeapp/v1/deployment-intent-groups/test-dig/network-controller-intent/net-ctl-intent/network-chains/"+t.inputName, nil)
			resp := executeRequest(request, NewRouter(t.client))

			//Check returned code
			Expect(resp.StatusCode).To(Equal(t.expectedCode))

			//Check returned body
			got := model.SfcIntent{}
			json.NewDecoder(resp.Body).Decode(&got)
			Expect(got).To(Equal(t.mockVal))
		},

		Entry("successful get", testCase{
			inputName:    "testsfcintent",
			expectedCode: http.StatusOK,
			mockError:    nil,
			mockVal: model.SfcIntent{
				Metadata: model.Metadata{
					Name:        "testsfcintent",
					Description: "test sfc intent",
					UserData1:   "some user data 1",
					UserData2:   "some user data 2",
				},
				Spec: model.SfcIntentSpec{
					ChainType:    "Routing",
					Namespace:    "chainspace",
					NetworkChain: "net=net0,app=a1,net=n1,app=a2,net=n2,app=a3,net=n3",
				},
			},
			client: &mocks.SfcIntentManager{},
		}),

		Entry("fails due to not found", testCase{
			inputName:    "testsfcintent",
			expectedCode: http.StatusNotFound,
			mockError:    pkgerrors.New("db Find error"),
			mockVal:      model.SfcIntent{},
			client:       &mocks.SfcIntentManager{},
		}),

		Entry("fails due to not found II", testCase{
			inputName:    "testsfcintent",
			expectedCode: http.StatusNotFound,
			mockError:    pkgerrors.New("not found"),
			mockVal:      model.SfcIntent{},
			client:       &mocks.SfcIntentManager{},
		}),

		Entry("fails due to some other backend error", testCase{
			inputName:    "testsfcintent",
			expectedCode: http.StatusInternalServerError,
			mockError:    pkgerrors.New("backend error"),
			mockVal:      model.SfcIntent{},
			client:       &mocks.SfcIntentManager{},
		}),
	)

	DescribeTable("Delete SfcIntent tests",
		func(t testCase) {
			// set up client mock responses
			t.client.On("DeleteSfcIntent", t.inputName, "test-project", "test-compositeapp", "v1", "test-dig", "net-ctl-intent").Return(t.mockError)

			// make HTTP request
			request := httptest.NewRequest("DELETE", "/v2/projects/test-project/composite-apps/test-compositeapp/v1/deployment-intent-groups/test-dig/network-controller-intent/net-ctl-intent/network-chains/"+t.inputName, nil)
			resp := executeRequest(request, NewRouter(t.client))

			//Check returned code
			Expect(resp.StatusCode).To(Equal(t.expectedCode))

			//Check returned body
			got := model.SfcIntent{}
			json.NewDecoder(resp.Body).Decode(&got)
			Expect(got).To(Equal(t.mockVal))
		},

		Entry("successful delete", testCase{
			inputName:    "testsfcintent",
			expectedCode: http.StatusNoContent,
			mockError:    nil,
			client:       &mocks.SfcIntentManager{},
		}),

		Entry("fails due to not found", testCase{
			inputName:    "testsfcintent",
			expectedCode: http.StatusNotFound,
			mockError:    pkgerrors.New("db Remove error - not found"),
			client:       &mocks.SfcIntentManager{},
		}),

		Entry("fails due to a conflict", testCase{
			inputName:    "testsfcintent",
			expectedCode: http.StatusConflict,
			mockError:    pkgerrors.New("db Remove error - conflict"),
			client:       &mocks.SfcIntentManager{},
		}),

		Entry("fails due to other backend error", testCase{
			inputName:    "testsfcintent",
			expectedCode: http.StatusInternalServerError,
			mockError:    pkgerrors.New("db Remove error - general"),
			client:       &mocks.SfcIntentManager{},
		}),
	)
})
