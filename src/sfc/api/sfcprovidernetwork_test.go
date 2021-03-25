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
	sfcProviderNetworkJSONFile = "../json-schemas/sfc-provider-network.json"
}

var _ = Describe("SfcProviderNetworkintenthandler", func() {

	type testCase struct {
		inputName    string
		inputReader  io.Reader
		inStruct     model.SfcProviderNetworkIntent
		mockError    error
		mockVal      model.SfcProviderNetworkIntent
		mockVals     []model.SfcProviderNetworkIntent
		expectedCode int
		client       *mocks.SfcProviderNetworkIntentManager
	}

	DescribeTable("Create SfcProviderNetworkIntent tests",
		func(t testCase) {
			// set up client mock responses
			t.client.On("CreateSfcProviderNetworkIntent", t.inStruct, "test-project", "test-compositeapp", "v1", "test-dig", "net-ctl-intent", "sfc-intent", false).Return(t.mockVal, t.mockError)

			// make HTTP request
			request := httptest.NewRequest("POST", "/v2/projects/test-project/composite-apps/test-compositeapp/v1/deployment-intent-groups/test-dig/network-controller-intent/net-ctl-intent/network-chains/sfc-intent/provider-networks", t.inputReader)
			resp := executeRequest(request, NewRouter(t.client))

			//Check returned code
			Expect(resp.StatusCode).To(Equal(t.expectedCode))

			//Check returned body
			got := model.SfcProviderNetworkIntent{}
			json.NewDecoder(resp.Body).Decode(&got)
			Expect(got).To(Equal(t.mockVal))
		},

		Entry("successful create", testCase{
			expectedCode: http.StatusCreated,
			inputReader: bytes.NewBuffer([]byte(`{
				"metadata": {
					"name": "testsfcprovidernetworkintent",
					"description": "test sfc intent",
					"userData1": "some user data 1",
					"userData2": "some user data 2"
				},
				"spec": {
					"chainEnd": "right",
				    "networkName": "sfc-provider-net",
					"gatewayIp": "1.2.3.4",
					"subnet": "1.2.3.0/24"
				}
			}`)),
			inStruct: model.SfcProviderNetworkIntent{
				Metadata: model.Metadata{
					Name:        "testsfcprovidernetworkintent",
					Description: "test sfc intent",
					UserData1:   "some user data 1",
					UserData2:   "some user data 2",
				},
				Spec: model.SfcProviderNetworkIntentSpec{
					ChainEnd:    "right",
					NetworkName: "sfc-provider-net",
					GatewayIp:   "1.2.3.4",
					Subnet:      "1.2.3.0/24",
				},
			},
			mockError: nil,
			mockVal: model.SfcProviderNetworkIntent{
				Metadata: model.Metadata{
					Name:        "testsfcprovidernetworkintent",
					Description: "test sfc intent",
					UserData1:   "some user data 1",
					UserData2:   "some user data 2",
				},
				Spec: model.SfcProviderNetworkIntentSpec{
					ChainEnd:    "right",
					NetworkName: "sfc-provider-net",
					GatewayIp:   "1.2.3.4",
					Subnet:      "1.2.3.0/24",
				},
			},
			client: &mocks.SfcProviderNetworkIntentManager{},
		}),

		Entry("fails due to empty body", testCase{
			expectedCode: http.StatusBadRequest,
			inStruct:     model.SfcProviderNetworkIntent{},
			mockError:    nil,
			mockVal:      model.SfcProviderNetworkIntent{},
			client:       &mocks.SfcProviderNetworkIntentManager{},
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
						"chainEnd": "right",
						"networkName": "sfc-provider-net",
						"gatewayIp": "1.2.3.4",
						"subnet": "1.2.3.0/24"
					}
				}`)),
			inStruct:  model.SfcProviderNetworkIntent{},
			mockError: nil,
			client:    &mocks.SfcProviderNetworkIntentManager{},
		}),

		Entry("fails due to bad IP address", testCase{
			expectedCode: http.StatusBadRequest,
			inputReader: bytes.NewBuffer([]byte(`{
					"metadata": {
						"name": "testsfcprovidernetworkintent",
						"description": "test sfc intent",
						"userData1": "some user data 1",
						"userData2": "some user data 2"
					},
					"spec": {
						"chainEnd": "right",
						"networkName": "sfc-provider-net",
						"gatewayIp": "1.2.3.400",
						"subnet": "1.2.3.0/24"
					}
				}`)),
			inStruct:  model.SfcProviderNetworkIntent{},
			mockError: nil,
			client:    &mocks.SfcProviderNetworkIntentManager{},
		}),

		Entry("fails due to bad subnet", testCase{
			expectedCode: http.StatusBadRequest,
			inputReader: bytes.NewBuffer([]byte(`{
					"metadata": {
						"name": "testsfcprovidernetworkintent",
						"description": "test sfc intent",
						"userData1": "some user data 1",
						"userData2": "some user data 2"
					},
					"spec": {
						"chainEnd": "right",
						"networkName": "sfc-provider-net",
						"gatewayIp": "1.2.3.4",
						"subnet": "1.2.3.0/33"
					}
				}`)),
			inStruct:  model.SfcProviderNetworkIntent{},
			mockError: nil,
			client:    &mocks.SfcProviderNetworkIntentManager{},
		}),

		Entry("fails due missing chain end", testCase{
			expectedCode: http.StatusBadRequest,
			inputReader: bytes.NewBuffer([]byte(`{
					"metadata": {
						"name": "testsfcprovidernetworkintent",
						"description": "test sfc intent",
						"userData1": "some user data 1",
						"userData2": "some user data 2"
					},
					"spec": {
					    "networkName": "sfc-provider-net",
						"gatewayIp": "1.2.3.4",
						"subnet": "1.2.3.0/24"
					}
				}`)),
			inStruct:  model.SfcProviderNetworkIntent{},
			mockError: nil,
			client:    &mocks.SfcProviderNetworkIntentManager{},
		}),

		Entry("fails due missing network name", testCase{
			expectedCode: http.StatusBadRequest,
			inputReader: bytes.NewBuffer([]byte(`{
					"metadata": {
						"name": "testsfcprovidernetworkintent",
						"description": "test sfc intent",
						"userData1": "some user data 1",
						"userData2": "some user data 2"
					},
					"spec": {
						"chainEnd": "right",
						"gatewayIp": "1.2.3.4",
						"subnet": "1.2.3.0/24"
					}
				}`)),
			inStruct:  model.SfcProviderNetworkIntent{},
			mockError: nil,
			client:    &mocks.SfcProviderNetworkIntentManager{},
		}),

		Entry("fails due missing gateway IP", testCase{
			expectedCode: http.StatusBadRequest,
			inputReader: bytes.NewBuffer([]byte(`{
					"metadata": {
						"name": "testsfcprovidernetworkintent",
						"description": "test sfc intent",
						"userData1": "some user data 1",
						"userData2": "some user data 2"
					},
					"spec": {
						"chainEnd": "right",
					    "networkName": "sfc-provider-net",
						"subnet": "1.2.3.0/24"
					}
				}`)),
			inStruct:  model.SfcProviderNetworkIntent{},
			mockError: nil,
			client:    &mocks.SfcProviderNetworkIntentManager{},
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
						"chainEnd": "right",
					    "networkName": "sfc-provider-net",
						"gatewayIp": "1.2.3.4",
						"subnet": "1.2.3.0/24"
					}
				}`)),
			inStruct:  model.SfcProviderNetworkIntent{},
			mockError: nil,
			client:    &mocks.SfcProviderNetworkIntentManager{},
		}),

		Entry("fails due to invalid chainEnd content", testCase{
			// chainEnd has value 'lefty'
			expectedCode: http.StatusBadRequest,
			inputReader: bytes.NewBuffer([]byte(`{
					"metadata": {
						"name": "testsfcintent",
						"description": "test sfc intent",
						"userData1": "some user data 1",
						"userData2": "some user data 2"
					},
					"spec": {
						"chainEnd": "rightnow",
					    "networkName": "sfc-provider-net",
						"gatewayIp": "1.2.3.4",
						"subnet": "1.2.3.0/24"
					}
				}`)),
			inStruct:  model.SfcProviderNetworkIntent{},
			mockError: nil,
			client:    &mocks.SfcProviderNetworkIntentManager{},
		}),

		Entry("fails due to entry already exists", testCase{
			expectedCode: http.StatusConflict,
			inputReader: bytes.NewBuffer([]byte(`{
					"metadata": {
						"name": "testsfcprovidernetworkintent",
						"description": "test sfc intent",
						"userData1": "some user data 1",
						"userData2": "some user data 2"
					},
					"spec": {
						"chainEnd": "right",
						"networkName": "sfc-provider-net",
						"gatewayIp": "1.2.3.4",
						"subnet": "1.2.3.0/24"
					}
				}`)),
			inStruct: model.SfcProviderNetworkIntent{
				Metadata: model.Metadata{
					Name:        "testsfcprovidernetworkintent",
					Description: "test sfc intent",
					UserData1:   "some user data 1",
					UserData2:   "some user data 2",
				},
				Spec: model.SfcProviderNetworkIntentSpec{
					ChainEnd:    "right",
					NetworkName: "sfc-provider-net",
					GatewayIp:   "1.2.3.4",
					Subnet:      "1.2.3.0/24",
				},
			},
			mockVal:   model.SfcProviderNetworkIntent{},
			mockError: pkgerrors.New("SFC Provider Network already exists"),
			client:    &mocks.SfcProviderNetworkIntentManager{},
		}),

		Entry("fails due to sfc provider network intent does not exist", testCase{
			expectedCode: http.StatusNotFound,
			inputReader: bytes.NewBuffer([]byte(`{
					"metadata": {
						"name": "testsfcprovidernetworkintent",
						"description": "test sfc intent",
						"userData1": "some user data 1",
						"userData2": "some user data 2"
					},
					"spec": {
						"chainEnd": "right",
						"networkName": "sfc-provider-net",
						"gatewayIp": "1.2.3.4",
						"subnet": "1.2.3.0/24"
					}
				}`)),
			inStruct: model.SfcProviderNetworkIntent{
				Metadata: model.Metadata{
					Name:        "testsfcprovidernetworkintent",
					Description: "test sfc intent",
					UserData1:   "some user data 1",
					UserData2:   "some user data 2",
				},
				Spec: model.SfcProviderNetworkIntentSpec{
					ChainEnd:    "right",
					NetworkName: "sfc-provider-net",
					GatewayIp:   "1.2.3.4",
					Subnet:      "1.2.3.0/24",
				},
			},
			mockVal:   model.SfcProviderNetworkIntent{},
			mockError: pkgerrors.New("does not exist"),
			client:    &mocks.SfcProviderNetworkIntentManager{},
		}),

		Entry("fails due to db error", testCase{
			expectedCode: http.StatusInternalServerError,
			inputReader: bytes.NewBuffer([]byte(`{
					"metadata": {
						"name": "testsfcprovidernetworkintent",
						"description": "test sfc intent",
						"userData1": "some user data 1",
						"userData2": "some user data 2"
					},
					"spec": {
						"chainEnd": "right",
						"networkName": "sfc-provider-net",
						"gatewayIp": "1.2.3.4",
						"subnet": "1.2.3.0/24"
					}
				}`)),
			inStruct: model.SfcProviderNetworkIntent{
				Metadata: model.Metadata{
					Name:        "testsfcprovidernetworkintent",
					Description: "test sfc intent",
					UserData1:   "some user data 1",
					UserData2:   "some user data 2",
				},
				Spec: model.SfcProviderNetworkIntentSpec{
					ChainEnd:    "right",
					NetworkName: "sfc-provider-net",
					GatewayIp:   "1.2.3.4",
					Subnet:      "1.2.3.0/24",
				},
			},
			mockVal:   model.SfcProviderNetworkIntent{},
			mockError: pkgerrors.New("Creating DB Entry"),
			client:    &mocks.SfcProviderNetworkIntentManager{},
		}),
	)

	DescribeTable("Put SfcProviderNetworkIntent tests",
		func(t testCase) {
			// set up client mock responses
			t.client.On("CreateSfcProviderNetworkIntent", t.inStruct, "test-project", "test-compositeapp", "v1", "test-dig", "net-ctl-intent", "sfc-intent", true).Return(t.mockVal, t.mockError)

			// make HTTP request
			request := httptest.NewRequest("PUT", "/v2/projects/test-project/composite-apps/test-compositeapp/v1/deployment-intent-groups/test-dig/network-controller-intent/net-ctl-intent/network-chains/sfc-intent/provider-networks/"+t.inputName, t.inputReader)
			resp := executeRequest(request, NewRouter(t.client))

			//Check returned code
			Expect(resp.StatusCode).To(Equal(t.expectedCode))

			//Check returned body
			got := model.SfcProviderNetworkIntent{}
			json.NewDecoder(resp.Body).Decode(&got)
			Expect(got).To(Equal(t.mockVal))
		},

		Entry("successful put", testCase{
			expectedCode: http.StatusCreated,
			inputName:    "testsfcprovidernetworkintent",
			inputReader: bytes.NewBuffer([]byte(`{
				"metadata": {
					"name": "testsfcprovidernetworkintent",
					"description": "test sfc intent",
					"userData1": "some user data 1",
					"userData2": "some user data 2"
				},
				"spec": {
					"chainEnd": "right",
					"networkName": "sfc-provider-net",
					"gatewayIp": "1.2.3.4",
					"subnet": "1.2.3.0/24"
				}
			}`)),
			inStruct: model.SfcProviderNetworkIntent{
				Metadata: model.Metadata{
					Name:        "testsfcprovidernetworkintent",
					Description: "test sfc intent",
					UserData1:   "some user data 1",
					UserData2:   "some user data 2",
				},
				Spec: model.SfcProviderNetworkIntentSpec{
					ChainEnd:    "right",
					NetworkName: "sfc-provider-net",
					GatewayIp:   "1.2.3.4",
					Subnet:      "1.2.3.0/24",
				},
			},
			mockError: nil,
			mockVal: model.SfcProviderNetworkIntent{
				Metadata: model.Metadata{
					Name:        "testsfcprovidernetworkintent",
					Description: "test sfc intent",
					UserData1:   "some user data 1",
					UserData2:   "some user data 2",
				},
				Spec: model.SfcProviderNetworkIntentSpec{
					ChainEnd:    "right",
					NetworkName: "sfc-provider-net",
					GatewayIp:   "1.2.3.4",
					Subnet:      "1.2.3.0/24",
				},
			},
			client: &mocks.SfcProviderNetworkIntentManager{},
		}),

		Entry("fails due to empty body", testCase{
			expectedCode: http.StatusBadRequest,
			inputName:    "testsfcprovidernetworkintent",
			inStruct:     model.SfcProviderNetworkIntent{},
			mockError:    nil,
			mockVal:      model.SfcProviderNetworkIntent{},
			client:       &mocks.SfcProviderNetworkIntentManager{},
		}),

		Entry("fails due missing name", testCase{
			expectedCode: http.StatusBadRequest,
			inputName:    "testsfcprovidernetworkintent",
			inputReader: bytes.NewBuffer([]byte(`{
					"metadata": {
						"description": "test sfc intent",
						"userData1": "some user data 1",
						"userData2": "some user data 2"
					},
					"spec": {
						"chainEnd": "right",
						"networkName": "sfc-provider-net",
						"gatewayIp": "1.2.3.4",
						"subnet": "1.2.3.0/24"
					}
				}`)),
			inStruct:  model.SfcProviderNetworkIntent{},
			mockError: nil,
			client:    &mocks.SfcProviderNetworkIntentManager{},
		}),

		Entry("fails due missing chain end", testCase{
			expectedCode: http.StatusBadRequest,
			inputName:    "testsfcprovidernetworkintent",
			inputReader: bytes.NewBuffer([]byte(`{
					"metadata": {
						"name": "testsfcprovidernetworkintent",
						"description": "test sfc intent",
						"userData1": "some user data 1",
						"userData2": "some user data 2"
					},
					"spec": {
						"networkName": "sfc-provider-net",
						"gatewayIp": "1.2.3.4",
						"subnet": "1.2.3.0/24"
					}
				}`)),
			inStruct:  model.SfcProviderNetworkIntent{},
			mockError: nil,
			client:    &mocks.SfcProviderNetworkIntentManager{},
		}),

		Entry("fails due to json body decoding error", testCase{
			// extra comma at the end of the userData2 line
			expectedCode: http.StatusUnprocessableEntity,
			inputName:    "testsfcprovidernetworkintent",
			inputReader: bytes.NewBuffer([]byte(`{
					"metadata": {
						"name": "testsfcprovidernetworkintent",
						"description": "test sfc intent",
						"userData1": "some user data 1",
						"userData2": "some user data 2",
					},
					"spec": {
						"chainEnd": "right",
						"networkName": "sfc-provider-net",
						"gatewayIp": "1.2.3.4",
						"subnet": "1.2.3.0/24"
					}
				}`)),
			inStruct:  model.SfcProviderNetworkIntent{},
			mockError: nil,
			client:    &mocks.SfcProviderNetworkIntentManager{},
		}),

		Entry("fails due to name mismatch", testCase{
			expectedCode: http.StatusBadRequest,
			inputName:    "testsfcprovidernetworkintent",
			inputReader: bytes.NewBuffer([]byte(`{
					"metadata": {
						"name": "testsfcprovidernetworkintentABC",
						"description": "test sfc intent",
						"userData1": "some user data 1",
						"userData2": "some user data 2"
					},
					"spec": {
						"chainEnd": "right",
						"networkName": "sfc-provider-net",
						"gatewayIp": "1.2.3.4",
						"subnet": "1.2.3.0/24"
					}
				}`)),
			inStruct: model.SfcProviderNetworkIntent{
				Metadata: model.Metadata{
					Name:        "testsfcprovidernetworkintent",
					Description: "test sfc intent",
					UserData1:   "some user data 1",
					UserData2:   "some user data 2",
				},
				Spec: model.SfcProviderNetworkIntentSpec{
					ChainEnd:    "right",
					NetworkName: "sfc-provider-net",
					GatewayIp:   "1.2.3.4",
					Subnet:      "1.2.3.0/24",
				},
			},
			mockVal:   model.SfcProviderNetworkIntent{},
			mockError: pkgerrors.New("SfcProviderNetworkIntent already exists"),
			client:    &mocks.SfcProviderNetworkIntentManager{},
		}),

		Entry("fails due to sfc provider network intent does not exist", testCase{
			expectedCode: http.StatusNotFound,
			inputName:    "testsfcprovidernetworkintent",
			inputReader: bytes.NewBuffer([]byte(`{
					"metadata": {
						"name": "testsfcprovidernetworkintent",
						"description": "test sfc intent",
						"userData1": "some user data 1",
						"userData2": "some user data 2"
					},
					"spec": {
						"chainEnd": "right",
						"networkName": "sfc-provider-net",
						"gatewayIp": "1.2.3.4",
						"subnet": "1.2.3.0/24"
					}
				}`)),
			inStruct: model.SfcProviderNetworkIntent{
				Metadata: model.Metadata{
					Name:        "testsfcprovidernetworkintent",
					Description: "test sfc intent",
					UserData1:   "some user data 1",
					UserData2:   "some user data 2",
				},
				Spec: model.SfcProviderNetworkIntentSpec{
					ChainEnd:    "right",
					NetworkName: "sfc-provider-net",
					GatewayIp:   "1.2.3.4",
					Subnet:      "1.2.3.0/24",
				},
			},
			mockVal:   model.SfcProviderNetworkIntent{},
			mockError: pkgerrors.New("does not exist"),
			client:    &mocks.SfcProviderNetworkIntentManager{},
		}),

		Entry("fails due to db error", testCase{
			expectedCode: http.StatusInternalServerError,
			inputName:    "testsfcprovidernetworkintent",
			inputReader: bytes.NewBuffer([]byte(`{
					"metadata": {
						"name": "testsfcprovidernetworkintent",
						"description": "test sfc intent",
						"userData1": "some user data 1",
						"userData2": "some user data 2"
					},
					"spec": {
						"chainEnd": "right",
						"networkName": "sfc-provider-net",
						"gatewayIp": "1.2.3.4",
						"subnet": "1.2.3.0/24"
					}
				}`)),
			inStruct: model.SfcProviderNetworkIntent{
				Metadata: model.Metadata{
					Name:        "testsfcprovidernetworkintent",
					Description: "test sfc intent",
					UserData1:   "some user data 1",
					UserData2:   "some user data 2",
				},
				Spec: model.SfcProviderNetworkIntentSpec{
					ChainEnd:    "right",
					NetworkName: "sfc-provider-net",
					GatewayIp:   "1.2.3.4",
					Subnet:      "1.2.3.0/24",
				},
			},
			mockVal:   model.SfcProviderNetworkIntent{},
			mockError: pkgerrors.New("Creating DB Entry"),
			client:    &mocks.SfcProviderNetworkIntentManager{},
		}),
	)

	DescribeTable("Get List SfcProviderNetworkIntent tests",
		func(t testCase) {
			// set up client mock responses
			t.client.On("GetAllSfcProviderNetworkIntents", "test-project", "test-compositeapp", "v1", "test-dig", "net-ctl-intent", "sfc-intent").Return(t.mockVals, t.mockError)

			// make HTTP request
			request := httptest.NewRequest("GET", "/v2/projects/test-project/composite-apps/test-compositeapp/v1/deployment-intent-groups/test-dig/network-controller-intent/net-ctl-intent/network-chains/sfc-intent/provider-networks", nil)
			resp := executeRequest(request, NewRouter(t.client))

			//Check returned code
			Expect(resp.StatusCode).To(Equal(t.expectedCode))

			//Check returned body
			got := []model.SfcProviderNetworkIntent{}
			json.NewDecoder(resp.Body).Decode(&got)
			Expect(got).To(Equal(t.mockVals))
		},

		Entry("successful get", testCase{
			expectedCode: http.StatusOK,
			mockError:    nil,
			mockVals: []model.SfcProviderNetworkIntent{
				{
					Metadata: model.Metadata{
						Name:        "testsfcprovidernetworkintent1",
						Description: "test sfc intent",
						UserData1:   "some user data 1",
						UserData2:   "some user data 2",
					},
					Spec: model.SfcProviderNetworkIntentSpec{
						ChainEnd:    "right",
						NetworkName: "sfc-provider-net",
						GatewayIp:   "1.2.3.4",
						Subnet:      "1.2.3.0/24",
					},
				},
				{
					Metadata: model.Metadata{
						Name:        "testsfcprovidernetworkintent2",
						Description: "test sfc intent",
						UserData1:   "some user data 1",
						UserData2:   "some user data 2",
					},
					Spec: model.SfcProviderNetworkIntentSpec{
						ChainEnd:    "right",
						NetworkName: "sfc-provider-net",
						GatewayIp:   "1.2.3.4",
						Subnet:      "1.2.3.0/24",
					},
				},
			},
			client: &mocks.SfcProviderNetworkIntentManager{},
		}),

		Entry("fails due to db find error", testCase{
			expectedCode: http.StatusNotFound,
			mockError:    pkgerrors.New("db Find error"),
			mockVals:     []model.SfcProviderNetworkIntent{},
			client:       &mocks.SfcProviderNetworkIntentManager{},
		}),

		Entry("fails due to parent SFC Intent not found", testCase{
			expectedCode: http.StatusNotFound,
			mockError:    pkgerrors.New("SFC Provider Network Intent not found"),
			mockVals:     []model.SfcProviderNetworkIntent{},
			client:       &mocks.SfcProviderNetworkIntentManager{},
		}),

		Entry("fails due to some other backend error", testCase{
			expectedCode: http.StatusInternalServerError,
			mockError:    pkgerrors.New("backend error"),
			mockVals:     []model.SfcProviderNetworkIntent{},
			client:       &mocks.SfcProviderNetworkIntentManager{},
		}),
	)

	DescribeTable("Get SfcProviderNetworkIntent tests",
		func(t testCase) {
			// set up client mock responses
			t.client.On("GetSfcProviderNetworkIntent", t.inputName, "test-project", "test-compositeapp", "v1", "test-dig", "net-ctl-intent", "sfc-intent").Return(t.mockVal, t.mockError)

			// make HTTP request
			request := httptest.NewRequest("GET", "/v2/projects/test-project/composite-apps/test-compositeapp/v1/deployment-intent-groups/test-dig/network-controller-intent/net-ctl-intent/network-chains/sfc-intent/provider-networks/"+t.inputName, nil)
			resp := executeRequest(request, NewRouter(t.client))

			//Check returned code
			Expect(resp.StatusCode).To(Equal(t.expectedCode))

			//Check returned body
			got := model.SfcProviderNetworkIntent{}
			json.NewDecoder(resp.Body).Decode(&got)
			Expect(got).To(Equal(t.mockVal))
		},

		Entry("successful get", testCase{
			inputName:    "testsfcprovidernetworkintent",
			expectedCode: http.StatusOK,
			mockError:    nil,
			mockVal: model.SfcProviderNetworkIntent{
				Metadata: model.Metadata{
					Name:        "testsfcprovidernetworkintent",
					Description: "test sfc intent",
					UserData1:   "some user data 1",
					UserData2:   "some user data 2",
				},
				Spec: model.SfcProviderNetworkIntentSpec{
					ChainEnd:    "right",
					NetworkName: "sfc-provider-net",
					GatewayIp:   "1.2.3.4",
					Subnet:      "1.2.3.0/24",
				},
			},
			client: &mocks.SfcProviderNetworkIntentManager{},
		}),

		Entry("fails due to not found", testCase{
			inputName:    "testsfcprovidernetworkintent",
			expectedCode: http.StatusNotFound,
			mockError:    pkgerrors.New("db Find error"),
			mockVal:      model.SfcProviderNetworkIntent{},
			client:       &mocks.SfcProviderNetworkIntentManager{},
		}),

		Entry("fails due to not found", testCase{
			inputName:    "testsfcprovidernetworkintent",
			expectedCode: http.StatusNotFound,
			mockError:    pkgerrors.New("not found"),
			mockVal:      model.SfcProviderNetworkIntent{},
			client:       &mocks.SfcProviderNetworkIntentManager{},
		}),

		Entry("fails due to some other backend error", testCase{
			inputName:    "testsfcprovidernetworkintent",
			expectedCode: http.StatusInternalServerError,
			mockError:    pkgerrors.New("backend error"),
			mockVal:      model.SfcProviderNetworkIntent{},
			client:       &mocks.SfcProviderNetworkIntentManager{},
		}),
	)

	DescribeTable("Delete SfcProviderNetworkIntent tests",
		func(t testCase) {
			// set up client mock responses
			t.client.On("DeleteSfcProviderNetworkIntent", t.inputName, "test-project", "test-compositeapp", "v1", "test-dig", "net-ctl-intent", "sfc-intent").Return(t.mockError)

			// make HTTP request
			request := httptest.NewRequest("DELETE", "/v2/projects/test-project/composite-apps/test-compositeapp/v1/deployment-intent-groups/test-dig/network-controller-intent/net-ctl-intent/network-chains/sfc-intent/provider-networks/"+t.inputName, nil)
			resp := executeRequest(request, NewRouter(t.client))

			//Check returned code
			Expect(resp.StatusCode).To(Equal(t.expectedCode))

			//Check returned body
			got := model.SfcProviderNetworkIntent{}
			json.NewDecoder(resp.Body).Decode(&got)
			Expect(got).To(Equal(t.mockVal))
		},

		Entry("successful delete", testCase{
			inputName:    "testsfcprovidernetworkintent",
			expectedCode: http.StatusNoContent,
			mockError:    nil,
			client:       &mocks.SfcProviderNetworkIntentManager{},
		}),

		Entry("fails due to not found", testCase{
			inputName:    "testsfcprovidernetworkintent",
			expectedCode: http.StatusNotFound,
			mockError:    pkgerrors.New("db Remove error - not found"),
			client:       &mocks.SfcProviderNetworkIntentManager{},
		}),

		Entry("fails due to a conflict", testCase{
			inputName:    "testsfcprovidernetworkintent",
			expectedCode: http.StatusConflict,
			mockError:    pkgerrors.New("db Remove error - conflict"),
			client:       &mocks.SfcProviderNetworkIntentManager{},
		}),

		Entry("fails due to other backend error", testCase{
			inputName:    "testsfcprovidernetworkintent",
			expectedCode: http.StatusInternalServerError,
			mockError:    pkgerrors.New("db Remove error - general"),
			client:       &mocks.SfcProviderNetworkIntentManager{},
		}),
	)
})
