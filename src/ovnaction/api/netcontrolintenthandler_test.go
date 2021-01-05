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
	"github.com/open-ness/EMCO/src/ovnaction/api/mocks"
	"github.com/open-ness/EMCO/src/ovnaction/pkg/module"
	pkgerrors "github.com/pkg/errors"
)

func init() {
	netCntIntJSONFile = "../json-schemas/metadata.json"
}

var _ = Describe("Netcontrolintenthandler", func() {

	type testCase struct {
		inputName    string
		inputReader  io.Reader
		inStruct     module.NetControlIntent
		mockError    error
		mockVal      module.NetControlIntent
		mockVals     []module.NetControlIntent
		expectedCode int
		client       *mocks.NetControlIntentManager
	}

	DescribeTable("Create NetControlIntent tests",
		func(t testCase) {
			// set up client mock responses
			t.client.On("CreateNetControlIntent", t.inStruct, "test-project", "test-compositeapp", "v1", "test-dig", false).Return(t.mockVal, t.mockError)

			// make HTTP request
			request := httptest.NewRequest("POST", "/v2/projects/test-project/composite-apps/test-compositeapp/v1/deployment-intent-groups/test-dig/network-controller-intent", t.inputReader)
			resp := executeRequest(request, NewRouter(t.client))

			//Check returned code
			Expect(resp.StatusCode).To(Equal(t.expectedCode))

			//Check returned body
			got := module.NetControlIntent{}
			json.NewDecoder(resp.Body).Decode(&got)
			Expect(got).To(Equal(t.mockVal))
		},

		Entry("successful create", testCase{
			expectedCode: http.StatusCreated,
			inputReader: bytes.NewBuffer([]byte(`{
				"metadata": {
					"name": "testnetcontrolintent",
					"description": "test network control intent",
					"userData1": "some user data 1",
					"userData2": "some user data 2"
				}
			}`)),
			inStruct: module.NetControlIntent{
				Metadata: module.Metadata{
					Name:        "testnetcontrolintent",
					Description: "test network control intent",
					UserData1:   "some user data 1",
					UserData2:   "some user data 2",
				},
			},
			mockError: nil,
			mockVal: module.NetControlIntent{
				Metadata: module.Metadata{
					Name:        "testnetcontrolintent",
					Description: "test network control intent",
					UserData1:   "some user data 1",
					UserData2:   "some user data 2",
				},
			},
			client: &mocks.NetControlIntentManager{},
		}),

		Entry("fails due to empty body", testCase{
			expectedCode: http.StatusBadRequest,
			inStruct:     module.NetControlIntent{},
			mockError:    nil,
			mockVal:      module.NetControlIntent{},
			client:       &mocks.NetControlIntentManager{},
		}),

		Entry("fails due missing name", testCase{
			expectedCode: http.StatusBadRequest,
			inputReader: bytes.NewBuffer([]byte(`{
				"metadata": {
					"description": "test network control intent",
					"userData1": "some user data 1",
					"userData2": "some user data 2"
				}
			}`)),
			inStruct:  module.NetControlIntent{},
			mockError: nil,
			client:    &mocks.NetControlIntentManager{},
		}),

		Entry("fails due to other json validation error", testCase{
			// name field has an '=' character
			expectedCode: http.StatusBadRequest,
			inputReader: bytes.NewBuffer([]byte(`{
				"metadata": {
					"name": "test=netcontrolintent",
					"description": "test network control intent",
					"userData1": "some user data 1",
					"userData2": "some user data 2"
				}
			}`)),
			inStruct:  module.NetControlIntent{},
			mockError: nil,
			client:    &mocks.NetControlIntentManager{},
		}),

		Entry("fails due to json body decoding error", testCase{
			// extra comma at the end of the userData2 line
			expectedCode: http.StatusUnprocessableEntity,
			inputReader: bytes.NewBuffer([]byte(`{
				"metadata": {
					"name": "testnetcontrolintent",
					"description": "test network control intent",
					"userData1": "some user data 1",
					"userData2": "some user data 2",
				}
			}`)),
			inStruct:  module.NetControlIntent{},
			mockError: nil,
			client:    &mocks.NetControlIntentManager{},
		}),

		Entry("fails due to entry already exists", testCase{
			expectedCode: http.StatusConflict,
			inputReader: bytes.NewBuffer([]byte(`{
				"metadata": {
					"name": "testnetcontrolintent",
					"description": "test network control intent",
					"userData1": "some user data 1",
					"userData2": "some user data 2"
				}
			}`)),
			inStruct: module.NetControlIntent{
				Metadata: module.Metadata{
					Name:        "testnetcontrolintent",
					Description: "test network control intent",
					UserData1:   "some user data 1",
					UserData2:   "some user data 2",
				},
			},
			mockVal:   module.NetControlIntent{},
			mockError: pkgerrors.New("NetControlIntent already exists"),
			client:    &mocks.NetControlIntentManager{},
		}),

		Entry("fails due to db error", testCase{
			expectedCode: http.StatusInternalServerError,
			inputReader: bytes.NewBuffer([]byte(`{
				"metadata": {
					"name": "testnetcontrolintent",
					"description": "test network control intent",
					"userData1": "some user data 1",
					"userData2": "some user data 2"
				}
			}`)),
			inStruct: module.NetControlIntent{
				Metadata: module.Metadata{
					Name:        "testnetcontrolintent",
					Description: "test network control intent",
					UserData1:   "some user data 1",
					UserData2:   "some user data 2",
				},
			},
			mockVal:   module.NetControlIntent{},
			mockError: pkgerrors.New("Creating DB Entry"),
			client:    &mocks.NetControlIntentManager{},
		}),
	)

	DescribeTable("Put NetControlIntent tests",
		func(t testCase) {
			// set up client mock responses
			t.client.On("CreateNetControlIntent", t.inStruct, "test-project", "test-compositeapp", "v1", "test-dig", true).Return(t.mockVal, t.mockError)

			// make HTTP request
			request := httptest.NewRequest("PUT", "/v2/projects/test-project/composite-apps/test-compositeapp/v1/deployment-intent-groups/test-dig/network-controller-intent/"+t.inputName, t.inputReader)
			resp := executeRequest(request, NewRouter(t.client))

			//Check returned code
			Expect(resp.StatusCode).To(Equal(t.expectedCode))

			//Check returned body
			got := module.NetControlIntent{}
			json.NewDecoder(resp.Body).Decode(&got)
			Expect(got).To(Equal(t.mockVal))
		},

		Entry("successful put", testCase{
			expectedCode: http.StatusCreated,
			inputName:    "testnetcontrolintent",
			inputReader: bytes.NewBuffer([]byte(`{
				"metadata": {
					"name": "testnetcontrolintent",
					"description": "test network control intent",
					"userData1": "some user data 1",
					"userData2": "some user data 2"
				}
			}`)),
			inStruct: module.NetControlIntent{
				Metadata: module.Metadata{
					Name:        "testnetcontrolintent",
					Description: "test network control intent",
					UserData1:   "some user data 1",
					UserData2:   "some user data 2",
				},
			},
			mockError: nil,
			mockVal: module.NetControlIntent{
				Metadata: module.Metadata{
					Name:        "testnetcontrolintent",
					Description: "test network control intent",
					UserData1:   "some user data 1",
					UserData2:   "some user data 2",
				},
			},
			client: &mocks.NetControlIntentManager{},
		}),

		Entry("fails due to empty body", testCase{
			inputName:    "testnetcontrolintent",
			expectedCode: http.StatusBadRequest,
			inStruct:     module.NetControlIntent{},
			mockError:    nil,
			mockVal:      module.NetControlIntent{},
			client:       &mocks.NetControlIntentManager{},
		}),

		Entry("fails due missing name", testCase{
			inputName:    "testnetcontrolintent",
			expectedCode: http.StatusBadRequest,
			inputReader: bytes.NewBuffer([]byte(`{
				"metadata": {
					"description": "test network control intent",
					"userData1": "some user data 1",
					"userData2": "some user data 2"
				}
			}`)),
			inStruct:  module.NetControlIntent{},
			mockError: nil,
			client:    &mocks.NetControlIntentManager{},
		}),

		Entry("fails due to other json validation error", testCase{
			// name field in body has an '=' character
			inputName:    "testnetcontrolintent",
			expectedCode: http.StatusBadRequest,
			inputReader: bytes.NewBuffer([]byte(`{
				"metadata": {
					"name": "test=netcontrolintent",
					"description": "test network control intent",
					"userData1": "some user data 1",
					"userData2": "some user data 2"
				}
			}`)),
			inStruct:  module.NetControlIntent{},
			mockError: nil,
			client:    &mocks.NetControlIntentManager{},
		}),

		Entry("fails due to json body decoding error", testCase{
			// extra comma at the end of the userData2 line
			inputName:    "testnetcontrolintent",
			expectedCode: http.StatusUnprocessableEntity,
			inputReader: bytes.NewBuffer([]byte(`{
				"metadata": {
					"name": "testnetcontrolintent",
					"description": "test network control intent",
					"userData1": "some user data 1",
					"userData2": "some user data 2",
				}
			}`)),
			inStruct:  module.NetControlIntent{},
			mockError: nil,
			client:    &mocks.NetControlIntentManager{},
		}),

		Entry("fails due to mismatched name", testCase{
			inputName:    "testnetcontrolintentXYZ",
			expectedCode: http.StatusBadRequest,
			inputReader: bytes.NewBuffer([]byte(`{
				"metadata": {
					"name": "testnetcontrolintent",
					"description": "test network control intent",
					"userData1": "some user data 1",
					"userData2": "some user data 2"
				}
			}`)),
			inStruct: module.NetControlIntent{
				Metadata: module.Metadata{
					Name:        "testnetcontrolintent",
					Description: "test network control intent",
					UserData1:   "some user data 1",
					UserData2:   "some user data 2",
				},
			},
			mockVal:   module.NetControlIntent{},
			mockError: pkgerrors.New("Creating DB Entry"),
			client:    &mocks.NetControlIntentManager{},
		}),

		Entry("fails due to db error", testCase{
			inputName:    "testnetcontrolintent",
			expectedCode: http.StatusInternalServerError,
			inputReader: bytes.NewBuffer([]byte(`{
				"metadata": {
					"name": "testnetcontrolintent",
					"description": "test network control intent",
					"userData1": "some user data 1",
					"userData2": "some user data 2"
				}
			}`)),
			inStruct: module.NetControlIntent{
				Metadata: module.Metadata{
					Name:        "testnetcontrolintent",
					Description: "test network control intent",
					UserData1:   "some user data 1",
					UserData2:   "some user data 2",
				},
			},
			mockVal:   module.NetControlIntent{},
			mockError: pkgerrors.New("Creating DB Entry"),
			client:    &mocks.NetControlIntentManager{},
		}),
	)

	DescribeTable("Get List NetControlIntent tests",
		func(t testCase) {
			// set up client mock responses
			t.client.On("GetNetControlIntents", "test-project", "test-compositeapp", "v1", "test-dig").Return(t.mockVals, t.mockError)

			// make HTTP request
			request := httptest.NewRequest("GET", "/v2/projects/test-project/composite-apps/test-compositeapp/v1/deployment-intent-groups/test-dig/network-controller-intent", nil)
			resp := executeRequest(request, NewRouter(t.client))

			//Check returned code
			Expect(resp.StatusCode).To(Equal(t.expectedCode))

			//Check returned body
			got := []module.NetControlIntent{}
			json.NewDecoder(resp.Body).Decode(&got)
			Expect(got).To(Equal(t.mockVals))
		},

		Entry("successful get", testCase{
			expectedCode: http.StatusOK,
			mockError:    nil,
			mockVals: []module.NetControlIntent{
				{
					Metadata: module.Metadata{
						Name:        "testnetcontrolintent1",
						Description: "test network control intent",
						UserData1:   "some user data 1",
						UserData2:   "some user data 2",
					},
				},
				{
					Metadata: module.Metadata{
						Name:        "testnetcontrolintent2",
						Description: "test network control intent",
						UserData1:   "some user data 1",
						UserData2:   "some user data 2",
					},
				},
			},
			client: &mocks.NetControlIntentManager{},
		}),

		Entry("fails due to not found", testCase{
			expectedCode: http.StatusNotFound,
			mockError:    pkgerrors.New("db Find error"),
			mockVals:     []module.NetControlIntent{},
			client:       &mocks.NetControlIntentManager{},
		}),

		Entry("fails due to some other backend error", testCase{
			expectedCode: http.StatusInternalServerError,
			mockError:    pkgerrors.New("backend error"),
			mockVals:     []module.NetControlIntent{},
			client:       &mocks.NetControlIntentManager{},
		}),
	)

	DescribeTable("Get NetControlIntent tests",
		func(t testCase) {
			// set up client mock responses
			t.client.On("GetNetControlIntent", t.inputName, "test-project", "test-compositeapp", "v1", "test-dig").Return(t.mockVal, t.mockError)

			// make HTTP request
			request := httptest.NewRequest("GET", "/v2/projects/test-project/composite-apps/test-compositeapp/v1/deployment-intent-groups/test-dig/network-controller-intent/"+t.inputName, nil)
			resp := executeRequest(request, NewRouter(t.client))

			//Check returned code
			Expect(resp.StatusCode).To(Equal(t.expectedCode))

			//Check returned body
			got := module.NetControlIntent{}
			json.NewDecoder(resp.Body).Decode(&got)
			Expect(got).To(Equal(t.mockVal))
		},

		Entry("successful get", testCase{
			inputName:    "testnetcontrolintent",
			expectedCode: http.StatusOK,
			mockError:    nil,
			mockVal: module.NetControlIntent{
				Metadata: module.Metadata{
					Name:        "testnetcontrolintent",
					Description: "test network control intent",
					UserData1:   "some user data 1",
					UserData2:   "some user data 2",
				},
			},
			client: &mocks.NetControlIntentManager{},
		}),

		Entry("fails due to not found", testCase{
			inputName:    "testnetcontrolintent",
			expectedCode: http.StatusNotFound,
			mockError:    pkgerrors.New("db Find error"),
			mockVal:      module.NetControlIntent{},
			client:       &mocks.NetControlIntentManager{},
		}),

		Entry("fails due to some other backend error", testCase{
			inputName:    "testnetcontrolintent",
			expectedCode: http.StatusInternalServerError,
			mockError:    pkgerrors.New("backend error"),
			mockVal:      module.NetControlIntent{},
			client:       &mocks.NetControlIntentManager{},
		}),
	)

	DescribeTable("Delete NetControlIntent tests",
		func(t testCase) {
			// set up client mock responses
			t.client.On("DeleteNetControlIntent", t.inputName, "test-project", "test-compositeapp", "v1", "test-dig").Return(t.mockError)

			// make HTTP request
			request := httptest.NewRequest("DELETE", "/v2/projects/test-project/composite-apps/test-compositeapp/v1/deployment-intent-groups/test-dig/network-controller-intent/"+t.inputName, nil)
			resp := executeRequest(request, NewRouter(t.client))

			//Check returned code
			Expect(resp.StatusCode).To(Equal(t.expectedCode))

			//Check returned body
			got := module.NetControlIntent{}
			json.NewDecoder(resp.Body).Decode(&got)
			Expect(got).To(Equal(t.mockVal))
		},

		Entry("successful delete", testCase{
			inputName:    "testnetcontrolintent",
			expectedCode: http.StatusNoContent,
			mockError:    nil,
			client:       &mocks.NetControlIntentManager{},
		}),

		Entry("fails due to not found", testCase{
			inputName:    "testnetcontrolintent",
			expectedCode: http.StatusNotFound,
			mockError:    pkgerrors.New("db Remove error - not found"),
			client:       &mocks.NetControlIntentManager{},
		}),

		Entry("fails due to a conflict", testCase{
			inputName:    "testnetcontrolintent",
			expectedCode: http.StatusConflict,
			mockError:    pkgerrors.New("db Remove error - conflict"),
			client:       &mocks.NetControlIntentManager{},
		}),

		Entry("fails due to other backend error", testCase{
			inputName:    "testnetcontrolintent",
			expectedCode: http.StatusInternalServerError,
			mockError:    pkgerrors.New("db Remove error - general"),
			client:       &mocks.NetControlIntentManager{},
		}),
	)
})
