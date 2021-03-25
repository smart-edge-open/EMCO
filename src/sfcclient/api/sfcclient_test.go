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
	"github.com/open-ness/EMCO/src/sfcclient/api/mocks"
	"github.com/open-ness/EMCO/src/sfcclient/pkg/model"
	pkgerrors "github.com/pkg/errors"
)

func init() {
	sfcClientJSONFile = "../json-schemas/sfc-client.json"
}

var _ = Describe("Sfcintenthandler", func() {

	type testCase struct {
		inputName    string
		inputReader  io.Reader
		inStruct     model.SfcClientIntent
		mockError    error
		mockVal      model.SfcClientIntent
		mockVals     []model.SfcClientIntent
		expectedCode int
		client       *mocks.SfcManager
	}

	DescribeTable("Create SfcClientIntent tests",
		func(t testCase) {
			// set up client mock responses
			t.client.On("CreateSfcClientIntent", t.inStruct, "test-project", "test-compositeapp", "v1", "test-dig", "net-ctl-intent", false).Return(t.mockVal, t.mockError)

			// make HTTP request
			request := httptest.NewRequest("POST", "/v2/projects/test-project/composite-apps/test-compositeapp/v1/deployment-intent-groups/test-dig/network-controller-intent/net-ctl-intent/sfc-clients", t.inputReader)
			resp := executeRequest(request, NewRouter(t.client))

			//Check returned code
			Expect(resp.StatusCode).To(Equal(t.expectedCode))

			//Check returned body
			got := model.SfcClientIntent{}
			json.NewDecoder(resp.Body).Decode(&got)
			Expect(got).To(Equal(t.mockVal))
		},

		Entry("successful create", testCase{
			expectedCode: http.StatusCreated,
			inputReader: bytes.NewBuffer([]byte(`{
				"metadata": {
					"name": "testsfcclientintent",
					"description": "test sfc intent",
					"userData1": "some user data 1",
					"userData2": "some user data 2"
				},
				"spec": {
					"chainEnd": "left",
					"chainName": "testChain",
					"chainCompositeApp": "chainCA",
					"chainCompositeAppVersion": "v1",
					"chainDeploymentIntentGroup": "chainDig",
					"chainNetControlIntent": "chainNetCtlIntent",
						"appName": "chainApp",
					"workloadResource": "chainDep",
					"resourceType": "Deployment"
				}
			}`)),
			inStruct: model.SfcClientIntent{
				Metadata: model.Metadata{
					Name:        "testsfcclientintent",
					Description: "test sfc intent",
					UserData1:   "some user data 1",
					UserData2:   "some user data 2",
				},
				Spec: model.SfcClientIntentSpec{
					ChainEnd:                   "left",
					ChainName:                  "testChain",
					ChainCompositeApp:          "chainCA",
					ChainCompositeAppVersion:   "v1",
					ChainDeploymentIntentGroup: "chainDig",
					ChainNetControlIntent:      "chainNetCtlIntent",
					AppName:                    "chainApp",
					WorkloadResource:           "chainDep",
					ResourceType:               "Deployment",
				},
			},
			mockError: nil,
			mockVal: model.SfcClientIntent{
				Metadata: model.Metadata{
					Name:        "testsfcclientintent",
					Description: "test sfc intent",
					UserData1:   "some user data 1",
					UserData2:   "some user data 2",
				},
				Spec: model.SfcClientIntentSpec{
					ChainEnd:                   "left",
					ChainName:                  "testChain",
					ChainCompositeApp:          "chainCA",
					ChainCompositeAppVersion:   "v1",
					ChainDeploymentIntentGroup: "chainDig",
					ChainNetControlIntent:      "chainNetCtlIntent",
					AppName:                    "chainApp",
					WorkloadResource:           "chainDep",
					ResourceType:               "Deployment",
				},
			},
			client: &mocks.SfcManager{},
		}),

		Entry("fails due to empty body", testCase{
			expectedCode: http.StatusBadRequest,
			inStruct:     model.SfcClientIntent{},
			mockError:    nil,
			mockVal:      model.SfcClientIntent{},
			client:       &mocks.SfcManager{},
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
						"chainEnd": "left",
						"chainName": "testChain",
						"chainCompositeApp": "chainCA",
						"chainCompositeAppVersion": "v1",
						"chainDeploymentIntentGroup": "chainDig",
						"chainNetControlIntent": "chainNetCtlIntent",
						"appName": "chainApp",
						"workloadResource": "chainDep",
						"resourceType": "Deployment"
					}
				}`)),
			inStruct:  model.SfcClientIntent{},
			mockError: nil,
			client:    &mocks.SfcManager{},
		}),

		Entry("fails due missing chain end", testCase{
			expectedCode: http.StatusBadRequest,
			inputReader: bytes.NewBuffer([]byte(`{
					"metadata": {
						"name": "testsfcclientintent",
						"description": "test sfc intent",
						"userData1": "some user data 1",
						"userData2": "some user data 2"
					},
					"spec": {
						"chainName": "testChain",
						"chainCompositeApp": "chainCA",
						"chainCompositeAppVersion": "v1",
						"chainDeploymentIntentGroup": "chainDig",
						"chainNetControlIntent": "chainNetCtlIntent",
						"appName": "chainApp",
						"workloadResource": "chainDep",
						"resourceType": "Deployment"
					}
				}`)),
			inStruct:  model.SfcClientIntent{},
			mockError: nil,
			client:    &mocks.SfcManager{},
		}),

		Entry("fails due missing chain name", testCase{
			expectedCode: http.StatusBadRequest,
			inputReader: bytes.NewBuffer([]byte(`{
					"metadata": {
						"name": "testsfcclientintent",
						"description": "test sfc intent",
						"userData1": "some user data 1",
						"userData2": "some user data 2"
					},
					"spec": {
						"chainEnd": "left",
						"chainCompositeApp": "chainCA",
						"chainCompositeAppVersion": "v1",
						"chainDeploymentIntentGroup": "chainDig",
						"chainNetControlIntent": "chainNetCtlIntent",
						"appName": "chainApp",
						"workloadResource": "chainDep",
						"resourceType": "Deployment"
					}
				}`)),
			inStruct:  model.SfcClientIntent{},
			mockError: nil,
			client:    &mocks.SfcManager{},
		}),

		Entry("fails due missing composite app", testCase{
			expectedCode: http.StatusBadRequest,
			inputReader: bytes.NewBuffer([]byte(`{
					"metadata": {
						"name": "testsfcclientintent",
						"description": "test sfc intent",
						"userData1": "some user data 1",
						"userData2": "some user data 2"
					},
					"spec": {
						"chainEnd": "left",
						"chainName": "testChain",
						"chainCompositeAppVersion": "v1",
						"chainDeploymentIntentGroup": "chainDig",
						"chainNetControlIntent": "chainNetCtlIntent",
						"appName": "chainApp",
						"workloadResource": "chainDep",
						"resourceType": "Deployment"
					}
				}`)),
			inStruct:  model.SfcClientIntent{},
			mockError: nil,
			client:    &mocks.SfcManager{},
		}),

		Entry("fails due to missing composite app version", testCase{
			expectedCode: http.StatusBadRequest,
			inputReader: bytes.NewBuffer([]byte(`{
					"metadata": {
						"name": "testsfcintent",
						"description": "test sfc intent",
						"userData1": "some user data 1",
						"userData2": "some user data 2"
					},
					"spec": {
						"chainEnd": "lefty",
						"chainName": "testChain",
						"chainCompositeApp": "chainCA",
						"chainDeploymentIntentGroup": "chainDig",
						"chainNetControlIntent": "chainNetCtlIntent",
						"appName": "chainApp",
						"workloadResource": "chainDep",
						"resourceType": "Deployment"
					}
				}`)),
			inStruct:  model.SfcClientIntent{},
			mockError: nil,
			client:    &mocks.SfcManager{},
		}),

		Entry("fails due to missing deployment intent group", testCase{
			expectedCode: http.StatusBadRequest,
			inputReader: bytes.NewBuffer([]byte(`{
					"metadata": {
						"name": "testsfcintent",
						"description": "test sfc intent",
						"userData1": "some user data 1",
						"userData2": "some user data 2"
					},
					"spec": {
						"chainEnd": "lefty",
						"chainName": "testChain",
						"chainCompositeApp": "chainCA",
						"chainCompositeAppVersion": "v1",
						"chainNetControlIntent": "chainNetCtlIntent",
						"appName": "chainApp",
						"workloadResource": "chainDep",
						"resourceType": "Deployment"
					}
				}`)),
			inStruct:  model.SfcClientIntent{},
			mockError: nil,
			client:    &mocks.SfcManager{},
		}),

		Entry("fails due to net control intent", testCase{
			expectedCode: http.StatusBadRequest,
			inputReader: bytes.NewBuffer([]byte(`{
					"metadata": {
						"name": "testsfcintent",
						"description": "test sfc intent",
						"userData1": "some user data 1",
						"userData2": "some user data 2"
					},
					"spec": {
						"chainEnd": "lefty",
						"chainName": "testChain",
						"chainCompositeApp": "chainCA",
						"chainCompositeAppVersion": "v1",
						"chainDeploymentIntentGroup": "chainDig",
						"appName": "chainApp",
						"workloadResource": "chainDep",
						"resourceType": "Deployment"
					}
				}`)),
			inStruct:  model.SfcClientIntent{},
			mockError: nil,
			client:    &mocks.SfcManager{},
		}),

		Entry("fails due to missing app name", testCase{
			expectedCode: http.StatusBadRequest,
			inputReader: bytes.NewBuffer([]byte(`{
					"metadata": {
						"name": "testsfcintent",
						"description": "test sfc intent",
						"userData1": "some user data 1",
						"userData2": "some user data 2"
					},
					"spec": {
						"chainEnd": "lefty",
						"chainName": "testChain",
						"chainCompositeApp": "chainCA",
						"chainCompositeAppVersion": "v1",
						"chainDeploymentIntentGroup": "chainDig",
						"chainNetControlIntent": "chainNetCtlIntent",
						"workloadResource": "chainDep",
						"resourceType": "Deployment"
					}
				}`)),
			inStruct:  model.SfcClientIntent{},
			mockError: nil,
			client:    &mocks.SfcManager{},
		}),

		Entry("fails due to missing workload resource", testCase{
			expectedCode: http.StatusBadRequest,
			inputReader: bytes.NewBuffer([]byte(`{
					"metadata": {
						"name": "testsfcintent",
						"description": "test sfc intent",
						"userData1": "some user data 1",
						"userData2": "some user data 2"
					},
					"spec": {
						"chainEnd": "lefty",
						"chainName": "testChain",
						"chainCompositeApp": "chainCA",
						"chainCompositeAppVersion": "v1",
						"chainDeploymentIntentGroup": "chainDig",
						"chainNetControlIntent": "chainNetCtlIntent",
						"appName": "chainApp",
						"resourceType": "Deployment"
					}
				}`)),
			inStruct:  model.SfcClientIntent{},
			mockError: nil,
			client:    &mocks.SfcManager{},
		}),

		Entry("fails due to missing resource type", testCase{
			expectedCode: http.StatusBadRequest,
			inputReader: bytes.NewBuffer([]byte(`{
					"metadata": {
						"name": "testsfcintent",
						"description": "test sfc intent",
						"userData1": "some user data 1",
						"userData2": "some user data 2"
					},
					"spec": {
						"chainEnd": "lefty",
						"chainName": "testChain",
						"chainCompositeApp": "chainCA",
						"chainCompositeAppVersion": "v1",
						"chainDeploymentIntentGroup": "chainDig",
						"chainNetControlIntent": "chainNetCtlIntent",
						"appName": "chainApp",
						"workloadResource": "chainDep"
					}
				}`)),
			inStruct:  model.SfcClientIntent{},
			mockError: nil,
			client:    &mocks.SfcManager{},
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
						"chainEnd": "lefty",
						"chainName": "testChain",
						"chainCompositeApp": "chainCA",
						"chainCompositeAppVersion": "v1",
						"chainDeploymentIntentGroup": "chainDig",
						"chainNetControlIntent": "chainNetCtlIntent",
						"appName": "chainApp",
						"workloadResource": "chainDep",
						"resourceType": "Deployment"
					}
				}`)),
			inStruct:  model.SfcClientIntent{},
			mockError: nil,
			client:    &mocks.SfcManager{},
		}),

		Entry("fails due to entry already exists", testCase{
			expectedCode: http.StatusConflict,
			inputReader: bytes.NewBuffer([]byte(`{
					"metadata": {
						"name": "testsfcclientintent",
						"description": "test sfc intent",
						"userData1": "some user data 1",
						"userData2": "some user data 2"
					},
					"spec": {
						"chainEnd": "left",
						"chainName": "testChain",
						"chainCompositeApp": "chainCA",
						"chainCompositeAppVersion": "v1",
						"chainDeploymentIntentGroup": "chainDig",
						"chainNetControlIntent": "chainNetCtlIntent",
						"appName": "chainApp",
						"workloadResource": "chainDep",
						"resourceType": "Deployment"
					}
				}`)),
			inStruct: model.SfcClientIntent{
				Metadata: model.Metadata{
					Name:        "testsfcclientintent",
					Description: "test sfc intent",
					UserData1:   "some user data 1",
					UserData2:   "some user data 2",
				},
				Spec: model.SfcClientIntentSpec{
					ChainEnd:                   "left",
					ChainName:                  "testChain",
					ChainCompositeApp:          "chainCA",
					ChainCompositeAppVersion:   "v1",
					ChainDeploymentIntentGroup: "chainDig",
					ChainNetControlIntent:      "chainNetCtlIntent",
					AppName:                    "chainApp",
					WorkloadResource:           "chainDep",
					ResourceType:               "Deployment",
				},
			},
			mockVal:   model.SfcClientIntent{},
			mockError: pkgerrors.New("SFC Client Intent already exists"),
			client:    &mocks.SfcManager{},
		}),

		Entry("fails due to network control intent does not exist", testCase{
			expectedCode: http.StatusNotFound,
			inputReader: bytes.NewBuffer([]byte(`{
					"metadata": {
						"name": "testsfcclientintent",
						"description": "test sfc intent",
						"userData1": "some user data 1",
						"userData2": "some user data 2"
					},
					"spec": {
						"chainEnd": "left",
						"chainName": "testChain",
						"chainCompositeApp": "chainCA",
						"chainCompositeAppVersion": "v1",
						"chainDeploymentIntentGroup": "chainDig",
						"chainNetControlIntent": "chainNetCtlIntent",
						"appName": "chainApp",
						"workloadResource": "chainDep",
						"resourceType": "Deployment"
					}
				}`)),
			inStruct: model.SfcClientIntent{
				Metadata: model.Metadata{
					Name:        "testsfcclientintent",
					Description: "test sfc intent",
					UserData1:   "some user data 1",
					UserData2:   "some user data 2",
				},
				Spec: model.SfcClientIntentSpec{
					ChainEnd:                   "left",
					ChainName:                  "testChain",
					ChainCompositeApp:          "chainCA",
					ChainCompositeAppVersion:   "v1",
					ChainDeploymentIntentGroup: "chainDig",
					ChainNetControlIntent:      "chainNetCtlIntent",
					AppName:                    "chainApp",
					WorkloadResource:           "chainDep",
					ResourceType:               "Deployment",
				},
			},
			mockVal:   model.SfcClientIntent{},
			mockError: pkgerrors.New("does not exist"),
			client:    &mocks.SfcManager{},
		}),

		Entry("fails due to db error", testCase{
			expectedCode: http.StatusInternalServerError,
			inputReader: bytes.NewBuffer([]byte(`{
					"metadata": {
						"name": "testsfcclientintent",
						"description": "test sfc intent",
						"userData1": "some user data 1",
						"userData2": "some user data 2"
					},
					"spec": {
						"chainEnd": "left",
						"chainName": "testChain",
						"chainCompositeApp": "chainCA",
						"chainCompositeAppVersion": "v1",
						"chainDeploymentIntentGroup": "chainDig",
						"chainNetControlIntent": "chainNetCtlIntent",
						"appName": "chainApp",
						"workloadResource": "chainDep",
						"resourceType": "Deployment"
					}
				}`)),
			inStruct: model.SfcClientIntent{
				Metadata: model.Metadata{
					Name:        "testsfcclientintent",
					Description: "test sfc intent",
					UserData1:   "some user data 1",
					UserData2:   "some user data 2",
				},
				Spec: model.SfcClientIntentSpec{
					ChainEnd:                   "left",
					ChainName:                  "testChain",
					ChainCompositeApp:          "chainCA",
					ChainCompositeAppVersion:   "v1",
					ChainDeploymentIntentGroup: "chainDig",
					ChainNetControlIntent:      "chainNetCtlIntent",
					AppName:                    "chainApp",
					WorkloadResource:           "chainDep",
					ResourceType:               "Deployment",
				},
			},
			mockVal:   model.SfcClientIntent{},
			mockError: pkgerrors.New("Creating DB Entry"),
			client:    &mocks.SfcManager{},
		}),
	)

	DescribeTable("Put SfcClientIntent tests",
		func(t testCase) {
			// set up client mock responses
			t.client.On("CreateSfcClientIntent", t.inStruct, "test-project", "test-compositeapp", "v1", "test-dig", "net-ctl-intent", true).Return(t.mockVal, t.mockError)

			// make HTTP request
			request := httptest.NewRequest("PUT", "/v2/projects/test-project/composite-apps/test-compositeapp/v1/deployment-intent-groups/test-dig/network-controller-intent/net-ctl-intent/sfc-clients/"+t.inputName, t.inputReader)
			resp := executeRequest(request, NewRouter(t.client))

			//Check returned code
			Expect(resp.StatusCode).To(Equal(t.expectedCode))

			//Check returned body
			got := model.SfcClientIntent{}
			json.NewDecoder(resp.Body).Decode(&got)
			Expect(got).To(Equal(t.mockVal))
		},

		Entry("successful put", testCase{
			expectedCode: http.StatusCreated,
			inputName:    "testsfcclientintent",
			inputReader: bytes.NewBuffer([]byte(`{
				"metadata": {
					"name": "testsfcclientintent",
					"description": "test sfc intent",
					"userData1": "some user data 1",
					"userData2": "some user data 2"
				},
				"spec": {
					"chainEnd": "left",
					"chainName": "testChain",
					"chainCompositeApp": "chainCA",
					"chainCompositeAppVersion": "v1",
					"chainDeploymentIntentGroup": "chainDig",
					"chainNetControlIntent": "chainNetCtlIntent",
						"appName": "chainApp",
					"workloadResource": "chainDep",
					"resourceType": "Deployment"
				}
			}`)),
			inStruct: model.SfcClientIntent{
				Metadata: model.Metadata{
					Name:        "testsfcclientintent",
					Description: "test sfc intent",
					UserData1:   "some user data 1",
					UserData2:   "some user data 2",
				},
				Spec: model.SfcClientIntentSpec{
					ChainEnd:                   "left",
					ChainName:                  "testChain",
					ChainCompositeApp:          "chainCA",
					ChainCompositeAppVersion:   "v1",
					ChainDeploymentIntentGroup: "chainDig",
					ChainNetControlIntent:      "chainNetCtlIntent",
					AppName:                    "chainApp",
					WorkloadResource:           "chainDep",
					ResourceType:               "Deployment",
				},
			},
			mockError: nil,
			mockVal: model.SfcClientIntent{
				Metadata: model.Metadata{
					Name:        "testsfcclientintent",
					Description: "test sfc intent",
					UserData1:   "some user data 1",
					UserData2:   "some user data 2",
				},
				Spec: model.SfcClientIntentSpec{
					ChainEnd:                   "left",
					ChainName:                  "testChain",
					ChainCompositeApp:          "chainCA",
					ChainCompositeAppVersion:   "v1",
					ChainDeploymentIntentGroup: "chainDig",
					ChainNetControlIntent:      "chainNetCtlIntent",
					AppName:                    "chainApp",
					WorkloadResource:           "chainDep",
					ResourceType:               "Deployment",
				},
			},
			client: &mocks.SfcManager{},
		}),

		Entry("fails due to empty body", testCase{
			expectedCode: http.StatusBadRequest,
			inputName:    "testsfcclientintent",
			inStruct:     model.SfcClientIntent{},
			mockError:    nil,
			mockVal:      model.SfcClientIntent{},
			client:       &mocks.SfcManager{},
		}),

		Entry("fails due missing name", testCase{
			expectedCode: http.StatusBadRequest,
			inputName:    "testsfcclientintent",
			inputReader: bytes.NewBuffer([]byte(`{
					"metadata": {
						"description": "test sfc intent",
						"userData1": "some user data 1",
						"userData2": "some user data 2"
					},
					"spec": {
						"chainEnd": "left",
						"chainName": "testChain",
						"chainCompositeApp": "chainCA",
						"chainCompositeAppVersion": "v1",
						"chainDeploymentIntentGroup": "chainDig",
						"chainNetControlIntent": "chainNetCtlIntent",
						"appName": "chainApp",
						"workloadResource": "chainDep",
						"resourceType": "Deployment"
					}
				}`)),
			inStruct:  model.SfcClientIntent{},
			mockError: nil,
			client:    &mocks.SfcManager{},
		}),

		Entry("fails due to other json validation error", testCase{
			// name field has an '=' character
			expectedCode: http.StatusBadRequest,
			inputName:    "testsfcclientintent",
			inputReader: bytes.NewBuffer([]byte(`{
					"metadata": {
						"name": "test=sfcintent",
						"description": "test sfc intent",
						"userData1": "some user data 1",
						"userData2": "some user data 2"
					},
					"spec": {
						"chainEnd": "left",
						"chainName": "testChain",
						"chainCompositeApp": "chainCA",
						"chainCompositeAppVersion": "v1",
						"chainDeploymentIntentGroup": "chainDig",
						"chainNetControlIntent": "chainNetCtlIntent",
						"appName": "chainApp",
						"workloadResource": "chainDep",
						"resourceType": "Deployment"
					}
				}`)),
			inStruct:  model.SfcClientIntent{},
			mockError: nil,
			client:    &mocks.SfcManager{},
		}),

		Entry("fails due to json body decoding error", testCase{
			// extra comma at the end of the userData2 line
			expectedCode: http.StatusUnprocessableEntity,
			inputName:    "testsfcclientintent",
			inputReader: bytes.NewBuffer([]byte(`{
					"metadata": {
						"name": "testsfcclientintent",
						"description": "test sfc intent",
						"userData1": "some user data 1",
						"userData2": "some user data 2",
					},
					"spec": {
						"chainEnd": "left",
						"chainName": "testChain",
						"chainCompositeApp": "chainCA",
						"chainCompositeAppVersion": "v1",
						"chainDeploymentIntentGroup": "chainDig",
						"chainNetControlIntent": "chainNetCtlIntent",
						"appName": "chainApp",
						"workloadResource": "chainDep",
						"resourceType": "Deployment"
					}
				}`)),
			inStruct:  model.SfcClientIntent{},
			mockError: nil,
			client:    &mocks.SfcManager{},
		}),

		Entry("fails due to name mismatch", testCase{
			expectedCode: http.StatusBadRequest,
			inputName:    "testsfcclientintent",
			inputReader: bytes.NewBuffer([]byte(`{
					"metadata": {
						"name": "testsfcclientintentABC",
						"description": "test sfc intent",
						"userData1": "some user data 1",
						"userData2": "some user data 2"
					},
					"spec": {
						"chainEnd": "left",
						"chainName": "testChain",
						"chainCompositeApp": "chainCA",
						"chainCompositeAppVersion": "v1",
						"chainDeploymentIntentGroup": "chainDig",
						"chainNetControlIntent": "chainNetCtlIntent",
						"appName": "chainApp",
						"workloadResource": "chainDep",
						"resourceType": "Deployment"
					}
				}`)),
			inStruct: model.SfcClientIntent{
				Metadata: model.Metadata{
					Name:        "testsfcclientintent",
					Description: "test sfc intent",
					UserData1:   "some user data 1",
					UserData2:   "some user data 2",
				},
				Spec: model.SfcClientIntentSpec{
					ChainEnd:                   "left",
					ChainName:                  "testChain",
					ChainCompositeApp:          "chainCA",
					ChainCompositeAppVersion:   "v1",
					ChainDeploymentIntentGroup: "chainDig",
					ChainNetControlIntent:      "chainNetCtlIntent",
					AppName:                    "chainApp",
					WorkloadResource:           "chainDep",
					ResourceType:               "Deployment",
				},
			},
			mockVal:   model.SfcClientIntent{},
			mockError: pkgerrors.New("SfcClientIntent already exists"),
			client:    &mocks.SfcManager{},
		}),

		Entry("fails due to network control intent does not exist", testCase{
			expectedCode: http.StatusNotFound,
			inputName:    "testsfcclientintent",
			inputReader: bytes.NewBuffer([]byte(`{
					"metadata": {
						"name": "testsfcclientintent",
						"description": "test sfc intent",
						"userData1": "some user data 1",
						"userData2": "some user data 2"
					},
					"spec": {
						"chainEnd": "left",
						"chainName": "testChain",
						"chainCompositeApp": "chainCA",
						"chainCompositeAppVersion": "v1",
						"chainDeploymentIntentGroup": "chainDig",
						"chainNetControlIntent": "chainNetCtlIntent",
						"appName": "chainApp",
						"workloadResource": "chainDep",
						"resourceType": "Deployment"
					}
				}`)),
			inStruct: model.SfcClientIntent{
				Metadata: model.Metadata{
					Name:        "testsfcclientintent",
					Description: "test sfc intent",
					UserData1:   "some user data 1",
					UserData2:   "some user data 2",
				},
				Spec: model.SfcClientIntentSpec{
					ChainEnd:                   "left",
					ChainName:                  "testChain",
					ChainCompositeApp:          "chainCA",
					ChainCompositeAppVersion:   "v1",
					ChainDeploymentIntentGroup: "chainDig",
					ChainNetControlIntent:      "chainNetCtlIntent",
					AppName:                    "chainApp",
					WorkloadResource:           "chainDep",
					ResourceType:               "Deployment",
				},
			},
			mockVal:   model.SfcClientIntent{},
			mockError: pkgerrors.New("does not exist"),
			client:    &mocks.SfcManager{},
		}),

		Entry("fails due to db error", testCase{
			expectedCode: http.StatusInternalServerError,
			inputName:    "testsfcclientintent",
			inputReader: bytes.NewBuffer([]byte(`{
					"metadata": {
						"name": "testsfcclientintent",
						"description": "test sfc intent",
						"userData1": "some user data 1",
						"userData2": "some user data 2"
					},
					"spec": {
						"chainEnd": "left",
						"chainName": "testChain",
						"chainCompositeApp": "chainCA",
						"chainCompositeAppVersion": "v1",
						"chainDeploymentIntentGroup": "chainDig",
						"chainNetControlIntent": "chainNetCtlIntent",
						"appName": "chainApp",
						"workloadResource": "chainDep",
						"resourceType": "Deployment"
					}
				}`)),
			inStruct: model.SfcClientIntent{
				Metadata: model.Metadata{
					Name:        "testsfcclientintent",
					Description: "test sfc intent",
					UserData1:   "some user data 1",
					UserData2:   "some user data 2",
				},
				Spec: model.SfcClientIntentSpec{
					ChainEnd:                   "left",
					ChainName:                  "testChain",
					ChainCompositeApp:          "chainCA",
					ChainCompositeAppVersion:   "v1",
					ChainDeploymentIntentGroup: "chainDig",
					ChainNetControlIntent:      "chainNetCtlIntent",
					AppName:                    "chainApp",
					WorkloadResource:           "chainDep",
					ResourceType:               "Deployment",
				},
			},
			mockVal:   model.SfcClientIntent{},
			mockError: pkgerrors.New("Creating DB Entry"),
			client:    &mocks.SfcManager{},
		}),
	)

	DescribeTable("Get List SfcClientIntent tests",
		func(t testCase) {
			// set up client mock responses
			t.client.On("GetAllSfcClientIntents", "test-project", "test-compositeapp", "v1", "test-dig", "net-ctl-intent").Return(t.mockVals, t.mockError)

			// make HTTP request
			request := httptest.NewRequest("GET", "/v2/projects/test-project/composite-apps/test-compositeapp/v1/deployment-intent-groups/test-dig/network-controller-intent/net-ctl-intent/sfc-clients", nil)
			resp := executeRequest(request, NewRouter(t.client))

			//Check returned code
			Expect(resp.StatusCode).To(Equal(t.expectedCode))

			//Check returned body
			got := []model.SfcClientIntent{}
			json.NewDecoder(resp.Body).Decode(&got)
			Expect(got).To(Equal(t.mockVals))
		},

		Entry("successful get", testCase{
			expectedCode: http.StatusOK,
			mockError:    nil,
			mockVals: []model.SfcClientIntent{
				{
					Metadata: model.Metadata{
						Name:        "testsfcclientintent1",
						Description: "test sfc intent",
						UserData1:   "some user data 1",
						UserData2:   "some user data 2",
					},
					Spec: model.SfcClientIntentSpec{
						ChainEnd:                   "left",
						ChainName:                  "testChain",
						ChainCompositeApp:          "chainCA",
						ChainCompositeAppVersion:   "v1",
						ChainDeploymentIntentGroup: "chainDig",
						ChainNetControlIntent:      "chainNetCtlIntent",
						AppName:                    "chainApp",
						WorkloadResource:           "chainDep",
						ResourceType:               "Deployment",
					},
				},
				{
					Metadata: model.Metadata{
						Name:        "testsfcclientintent2",
						Description: "test sfc intent",
						UserData1:   "some user data 1",
						UserData2:   "some user data 2",
					},
					Spec: model.SfcClientIntentSpec{
						ChainEnd:                   "left",
						ChainName:                  "testChain",
						ChainCompositeApp:          "chainCA",
						ChainCompositeAppVersion:   "v1",
						ChainDeploymentIntentGroup: "chainDig",
						ChainNetControlIntent:      "chainNetCtlIntent",
						AppName:                    "chainApp",
						WorkloadResource:           "chainDep",
						ResourceType:               "Deployment",
					},
				},
			},
			client: &mocks.SfcManager{},
		}),

		Entry("fails due to not found", testCase{
			expectedCode: http.StatusNotFound,
			mockError:    pkgerrors.New("db Find error"),
			mockVals:     []model.SfcClientIntent{},
			client:       &mocks.SfcManager{},
		}),

		Entry("fails due to some other backend error", testCase{
			expectedCode: http.StatusInternalServerError,
			mockError:    pkgerrors.New("backend error"),
			mockVals:     []model.SfcClientIntent{},
			client:       &mocks.SfcManager{},
		}),
	)

	DescribeTable("Get SfcClientIntent tests",
		func(t testCase) {
			// set up client mock responses
			t.client.On("GetSfcClientIntent", t.inputName, "test-project", "test-compositeapp", "v1", "test-dig", "net-ctl-intent").Return(t.mockVal, t.mockError)

			// make HTTP request
			request := httptest.NewRequest("GET", "/v2/projects/test-project/composite-apps/test-compositeapp/v1/deployment-intent-groups/test-dig/network-controller-intent/net-ctl-intent/sfc-clients/"+t.inputName, nil)
			resp := executeRequest(request, NewRouter(t.client))

			//Check returned code
			Expect(resp.StatusCode).To(Equal(t.expectedCode))

			//Check returned body
			got := model.SfcClientIntent{}
			json.NewDecoder(resp.Body).Decode(&got)
			Expect(got).To(Equal(t.mockVal))
		},

		Entry("successful get", testCase{
			inputName:    "testsfcclientintent",
			expectedCode: http.StatusOK,
			mockError:    nil,
			mockVal: model.SfcClientIntent{
				Metadata: model.Metadata{
					Name:        "testsfcclientintent",
					Description: "test sfc intent",
					UserData1:   "some user data 1",
					UserData2:   "some user data 2",
				},
				Spec: model.SfcClientIntentSpec{
					ChainEnd:                   "left",
					ChainName:                  "testChain",
					ChainCompositeApp:          "chainCA",
					ChainCompositeAppVersion:   "v1",
					ChainDeploymentIntentGroup: "chainDig",
					ChainNetControlIntent:      "chainNetCtlIntent",
					AppName:                    "chainApp",
					WorkloadResource:           "chainDep",
					ResourceType:               "Deployment",
				},
			},
			client: &mocks.SfcManager{},
		}),

		Entry("fails due to not found", testCase{
			inputName:    "testsfcclientintent",
			expectedCode: http.StatusNotFound,
			mockError:    pkgerrors.New("db Find error"),
			mockVal:      model.SfcClientIntent{},
			client:       &mocks.SfcManager{},
		}),

		Entry("fails due to not found II", testCase{
			inputName:    "testsfcclientintent",
			expectedCode: http.StatusNotFound,
			mockError:    pkgerrors.New("not found"),
			mockVal:      model.SfcClientIntent{},
			client:       &mocks.SfcManager{},
		}),

		Entry("fails due to some other backend error", testCase{
			inputName:    "testsfcclientintent",
			expectedCode: http.StatusInternalServerError,
			mockError:    pkgerrors.New("backend error"),
			mockVal:      model.SfcClientIntent{},
			client:       &mocks.SfcManager{},
		}),
	)

	DescribeTable("Delete SfcClientIntent tests",
		func(t testCase) {
			// set up client mock responses
			t.client.On("DeleteSfcClientIntent", t.inputName, "test-project", "test-compositeapp", "v1", "test-dig", "net-ctl-intent").Return(t.mockError)

			// make HTTP request
			request := httptest.NewRequest("DELETE", "/v2/projects/test-project/composite-apps/test-compositeapp/v1/deployment-intent-groups/test-dig/network-controller-intent/net-ctl-intent/sfc-clients/"+t.inputName, nil)
			resp := executeRequest(request, NewRouter(t.client))

			//Check returned code
			Expect(resp.StatusCode).To(Equal(t.expectedCode))

			//Check returned body
			got := model.SfcClientIntent{}
			json.NewDecoder(resp.Body).Decode(&got)
			Expect(got).To(Equal(t.mockVal))
		},

		Entry("successful delete", testCase{
			inputName:    "testsfcclientintent",
			expectedCode: http.StatusNoContent,
			mockError:    nil,
			client:       &mocks.SfcManager{},
		}),

		Entry("fails due to not found", testCase{
			inputName:    "testsfcclientintent",
			expectedCode: http.StatusNotFound,
			mockError:    pkgerrors.New("db Remove error - not found"),
			client:       &mocks.SfcManager{},
		}),

		Entry("fails due to a conflict", testCase{
			inputName:    "testsfcclientintent",
			expectedCode: http.StatusConflict,
			mockError:    pkgerrors.New("db Remove error - conflict"),
			client:       &mocks.SfcManager{},
		}),

		Entry("fails due to other backend error", testCase{
			inputName:    "testsfcclientintent",
			expectedCode: http.StatusInternalServerError,
			mockError:    pkgerrors.New("db Remove error - general"),
			client:       &mocks.SfcManager{},
		}),
	)
})
