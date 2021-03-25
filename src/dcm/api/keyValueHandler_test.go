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
	"github.com/open-ness/EMCO/src/dcm/api/mocks"
	"github.com/open-ness/EMCO/src/dcm/pkg/module"
	pkgerrors "github.com/pkg/errors"
)

var _ = Describe("KeyValueHandler", func() {
	type testCase struct {
		inputName    string
		inputReader  io.Reader
		inStruct     module.KeyValue
		mockError    error
		mockVal      module.KeyValue
		mockVals     []module.KeyValue
		expectedCode int
		lcClient     *mocks.LogicalCloudManager
		clClient     *mocks.ClusterManager
		upClient     *mocks.UserPermissionManager
		quotaClient  *mocks.QuotaManager
		kvClient     *mocks.KeyValueManager
	}

	DescribeTable("Create KeyValue tests",
		func(t testCase) {
			// set up client mock responses
			t.kvClient.On("CreateKVPair", "test-project", "test-lc", t.inStruct).Return(t.mockVal, t.mockError)

			// make HTTP request
			request := httptest.NewRequest("POST", "/v2/projects/test-project/logical-clouds/test-lc/kv-pairs", t.inputReader)
			resp := executeRequest(request, NewRouter(t.lcClient, t.clClient, t.upClient, t.quotaClient, t.kvClient))

			// Check returned code
			Expect(resp.StatusCode).To(Equal(t.expectedCode))

			// Check returned body
			got := module.KeyValue{}
			json.NewDecoder(resp.Body).Decode(&got)
			Expect(got).To(Equal(t.mockVal))
		},

		Entry("successful create", testCase{
			expectedCode: http.StatusCreated,
			inputReader: bytes.NewBuffer([]byte(`{
			"metadata": {
				"name": "testkvpair",
				"description": "description",
				"userData1": "some user data 1",
				"userData2": "some user data 2"
			}
		}`)),
			inStruct: module.KeyValue{
				MetaData: module.KVMetaDataList{
					KeyValueName: "testkvpair",
					Description:  "description",
					UserData1:    "some user data 1",
					UserData2:    "some user data 2",
				},
			},
			mockError: nil,
			mockVal: module.KeyValue{
				MetaData: module.KVMetaDataList{
					KeyValueName: "testkvpair",
					Description:  "description",
					UserData1:    "some user data 1",
					UserData2:    "some user data 2",
				},
			},
			kvClient: &mocks.KeyValueManager{},
		}),

		Entry("fails due to empty body", testCase{
			expectedCode: http.StatusBadRequest,
			inStruct:     module.KeyValue{},
			mockError:    nil,
			mockVal:      module.KeyValue{},
			kvClient:     &mocks.KeyValueManager{},
		}),

		Entry("fails due missing name", testCase{
			expectedCode: http.StatusBadRequest,
			inputReader: bytes.NewBuffer([]byte(`{
				"metadata": {
					"description": "description",
					"userData1": "some user data 1",
					"userData2": "some user data 2"
				}
		}`)),
			inStruct:  module.KeyValue{},
			mockError: nil,
			kvClient:  &mocks.KeyValueManager{},
		}),

		// TODO: implement logic and then enable this test:
		// Entry("fails due to other json validation error", testCase{
		// 	// name field has an '=' character
		// 	expectedCode: http.StatusBadRequest,
		// 	inputReader: bytes.NewBuffer([]byte(`{
		// 	"metadata": {
		// 		"name": "kv=pair",
		// 		"description": "description",
		// 		"userData1": "some user data 1",
		// 		"userData2": "some user data 2"
		// 	}
		// }`)),
		// 	inStruct:    module.KeyValue{},
		// 	mockError:   nil,
		// 	kvClient:&mocks.KeyValueManager{},
		// }),

		Entry("fails due to json body decoding error", testCase{
			// extra comma at the end of the userData2 line
			expectedCode: http.StatusUnprocessableEntity,
			inputReader: bytes.NewBuffer([]byte(`{
			"metadata": {
				"name": "testkvpair",
				"description": "description",
				"userData1": "some user data 1",
				"userData2": "some user data 2",
			}
		}`)),
			inStruct:  module.KeyValue{},
			mockError: nil,
			kvClient:  &mocks.KeyValueManager{},
		}),

		// TODO: implement logic and then enable this test:
		// Entry("fails due to entry already exists", testCase{
		// 	expectedCode: http.StatusConflict,
		// 	inputReader: bytes.NewBuffer([]byte(`{
		// 	"metadata": {
		// 		"name": "testkvpair",
		// 		"description": "description",
		// 		"userData1": "some user data 1",
		// 		"userData2": "some user data 2"
		// 	}
		// }`)),
		// 	inStruct: module.KeyValue{
		// 		MetaData: module.KVMetaDataList{
		// 			KeyValueName:   "testkvpair",
		// 			Description: "description",
		// 			UserData1:   "some user data 1",
		// 			UserData2:   "some user data 2",
		// 		},
		// 	},
		// 	mockVal:     module.KeyValue{},
		// 	mockError:   pkgerrors.New("KeyValue already exists"),
		// 	kvClient:&mocks.KeyValueManager{},
		// }),

		Entry("fails due to db error", testCase{
			expectedCode: http.StatusInternalServerError,
			inputReader: bytes.NewBuffer([]byte(`{
			"metadata": {
				"name": "testkvpair",
				"description": "description",
				"userData1": "some user data 1",
				"userData2": "some user data 2"
			}
		}`)),
			inStruct: module.KeyValue{
				MetaData: module.KVMetaDataList{
					KeyValueName: "testkvpair",
					Description:  "description",
					UserData1:    "some user data 1",
					UserData2:    "some user data 2",
				},
			},
			mockVal:   module.KeyValue{},
			mockError: pkgerrors.New("Creating DB Entry"),
			kvClient:  &mocks.KeyValueManager{},
		}),
	)

	// DCM PUT API currently disabled, so all tests commented out
	// DescribeTable("Put KeyValue tests",
	// 	func(t testCase) {
	// 		// set up client mock responses
	// 		t.kvClient.On("UpdateKVPair", "test-project", "test-lc", t.inputName, t.inStruct).Return(t.mockVal, t.mockError)

	// 		// make HTTP request
	// 		request := httptest.NewRequest("PUT", "/v2/projects/test-project/logical-clouds/test-lc/kv-pairs/"+t.inputName, t.inputReader)
	// 		resp := executeRequest(request, NewRouter(t.lcClient, t.clClient, t.upClient, t.quotaClient, t.kvClient))

	// 		//Check returned code
	// 		Expect(resp.StatusCode).To(Equal(t.expectedCode))

	// 		//Check returned body
	// 		got := module.KeyValue{}
	// 		json.NewDecoder(resp.Body).Decode(&got)
	// 		Expect(got).To(Equal(t.mockVal))
	// 	},

	// 	Entry("successful put", testCase{
	// 		expectedCode: http.StatusOK, // TODO: change to StatusCreated?
	// 		inputName:    "kvpair",
	// 		inputReader: bytes.NewBuffer([]byte(`{
	// 		"metadata": {
	// 			"name": "kvpair",
	// 			"description": "description",
	// 			"userData1": "some user data 1",
	// 			"userData2": "some user data 2"
	// 		},
	// 		"spec" : {
	// 			"limits.cpu": "500",
	// 			"limits.memory": "2000Gi"
	// 		}
	// 	}`)),
	// 		inStruct: module.KeyValue{
	// 			MetaData: module.KVMetaDataList{
	// 				KeyValueName:   "kvpair",
	// 				Description: "description",
	// 				UserData1:   "some user data 1",
	// 				UserData2:   "some user data 2",
	// 			},
	// 			Specification: map[string]string{
	// 				"limits.cpu":    "500",
	// 				"limits.memory": "2000Gi",
	// 			},
	// 		},
	// 		mockError: nil,
	// 		mockVal: module.KeyValue{
	// 			MetaData: module.KVMetaDataList{
	// 				KeyValueName:   "kvpair",
	// 				Description: "description",
	// 				UserData1:   "some user data 1",
	// 				UserData2:   "some user data 2",
	// 			},
	// 			Specification: map[string]string{
	// 				"limits.cpu":    "400",
	// 				"limits.memory": "1000Gi",
	// 			},
	// 		},
	// 		kvClient:&mocks.KeyValueManager{},
	// 	}),

	// 	Entry("fails due to empty body", testCase{
	// 		inputName:    "kvpair",
	// 		expectedCode: http.StatusBadRequest,
	// 		inStruct:     module.KeyValue{},
	// 		mockError:    nil,
	// 		mockVal:      module.KeyValue{},
	// 		kvClient: &mocks.KeyValueManager{},
	// 	}),

	// 	Entry("fails due missing name", testCase{
	// 		inputName:    "kvpair",
	// 		expectedCode: http.StatusBadRequest,
	// 		inputReader: bytes.NewBuffer([]byte(`{
	// 		"metadata": {
	// 			"description": "description",
	// 			"userData1": "some user data 1",
	// 			"userData2": "some user data 2"
	// 		}
	// 	}`)),
	// 		inStruct:    module.KeyValue{},
	// 		mockError:   nil,
	// 		kvClient:&mocks.KeyValueManager{},
	// 	}),

	// 	// TODO: implement logic and then enable this test:
	// 	// Entry("fails due to other json validation error", testCase{
	// 	// 	// name field in body has an '=' character
	// 	// 	inputName:    "kvpair",
	// 	// 	expectedCode: http.StatusBadRequest,
	// 	// 	inputReader: bytes.NewBuffer([]byte(`{
	// 	// 	"metadata": {
	// 	// 		"name": "test=quota",
	// 	// 		"description": "description",
	// 	// 		"userData1": "some user data 1",
	// 	// 		"userData2": "some user data 2"
	// 	// 	}
	// 	// }`)),
	// 	// 	inStruct:    module.KeyValue{},
	// 	// 	mockError:   nil,
	// 	// 	kvClient:&mocks.KeyValueManager{},
	// 	// }),

	// 	Entry("fails due to json body decoding error", testCase{
	// 		// extra comma at the end of the userData2 line
	// 		inputName:    "kvpair",
	// 		expectedCode: http.StatusUnprocessableEntity,
	// 		inputReader: bytes.NewBuffer([]byte(`{
	// 		"metadata": {
	// 			"name": "kvpair",
	// 			"description": "description",
	// 			"userData1": "some user data 1",
	// 			"userData2": "some user data 2",
	// 		}
	// 	}`)),
	// 		inStruct:    module.KeyValue{},
	// 		mockError:   nil,
	// 		kvClient:&mocks.KeyValueManager{},
	// 	}),

	// 	// TODO: implement logic and then enable this test:
	// 	// Entry("fails due to mismatched name", testCase{
	// 	// 	inputName:    "quotaXYZ",
	// 	// 	expectedCode: http.StatusBadRequest,
	// 	// 	inputReader: bytes.NewBuffer([]byte(`{
	// 	// 	"metadata": {
	// 	// 		"name": "kvpair",
	// 	// 		"description": "description",
	// 	// 		"userData1": "some user data 1",
	// 	// 		"userData2": "some user data 2"
	// 	// 	}
	// 	// }`)),
	// 	// 	inStruct: module.KeyValue{
	// 	// 		MetaData: module.KVMetaDataList{
	// 	// 			KeyValueName:   "kvpair",
	// 	// 			Description: "description",
	// 	// 			UserData1:   "some user data 1",
	// 	// 			UserData2:   "some user data 2",
	// 	// 		},
	// 	// 	},
	// 	// 	mockVal:     module.KeyValue{},
	// 	// 	mockError:   pkgerrors.New("Creating DB Entry"),
	// 	// 	kvClient:&mocks.KeyValueManager{},
	// 	// }),

	// 	Entry("fails due to db error", testCase{
	// 		inputName:    "kvpair",
	// 		expectedCode: http.StatusInternalServerError,
	// 		inputReader: bytes.NewBuffer([]byte(`{
	// 		"metadata": {
	// 			"name": "kvpair",
	// 			"description": "description",
	// 			"userData1": "some user data 1",
	// 			"userData2": "some user data 2"
	// 		}
	// 	}`)),
	// 		inStruct: module.KeyValue{
	// 			MetaData: module.KVMetaDataList{
	// 				KeyValueName:   "kvpair",
	// 				Description: "description",
	// 				UserData1:   "some user data 1",
	// 				UserData2:   "some user data 2",
	// 			},
	// 		},
	// 		mockVal:     module.KeyValue{},
	// 		mockError:   pkgerrors.New("Creating DB Entry"),
	// 		kvClient:&mocks.KeyValueManager{},
	// 	}),
	// )

	DescribeTable("Get List KeyValue tests",
		func(t testCase) {
			// set up client mock responses
			t.kvClient.On("GetAllKVPairs", "test-project", "test-lc").Return(t.mockVals, t.mockError)

			// make HTTP request
			request := httptest.NewRequest("GET", "/v2/projects/test-project/logical-clouds/test-lc/kv-pairs", nil)
			resp := executeRequest(request, NewRouter(t.lcClient, t.clClient, t.upClient, t.quotaClient, t.kvClient))

			// Check returned code
			Expect(resp.StatusCode).To(Equal(t.expectedCode))

			// Check returned body
			got := []module.KeyValue{}
			json.NewDecoder(resp.Body).Decode(&got)
			Expect(got).To(Equal(t.mockVals))
		},

		Entry("successful get", testCase{
			expectedCode: http.StatusOK,
			mockError:    nil,
			mockVals: []module.KeyValue{
				{
					MetaData: module.KVMetaDataList{
						KeyValueName: "testkvpair1",
						Description:  "description",
						UserData1:    "some user data 1",
						UserData2:    "some user data 2",
					},
				},
				{
					MetaData: module.KVMetaDataList{
						KeyValueName: "testkvpair2",
						Description:  "description",
						UserData1:    "some user data 1",
						UserData2:    "some user data 2",
					},
				},
			},
			kvClient: &mocks.KeyValueManager{},
		}),

		Entry("fails due to some other backend error", testCase{
			expectedCode: http.StatusInternalServerError,
			mockError:    pkgerrors.New("backend error"),
			mockVals:     []module.KeyValue{},
			kvClient:     &mocks.KeyValueManager{},
		}),
	)

	DescribeTable("Get KeyValue tests",
		func(t testCase) {
			// set up client mock responses
			t.kvClient.On("GetKVPair", "test-project", "test-lc", t.inputName).Return(t.mockVal, t.mockError)

			// make HTTP request
			request := httptest.NewRequest("GET", "/v2/projects/test-project/logical-clouds/test-lc/kv-pairs/"+t.inputName, nil)
			resp := executeRequest(request, NewRouter(t.lcClient, t.clClient, t.upClient, t.quotaClient, t.kvClient))

			// Check returned code
			Expect(resp.StatusCode).To(Equal(t.expectedCode))

			// Check returned body
			got := module.KeyValue{}
			json.NewDecoder(resp.Body).Decode(&got)
			Expect(got).To(Equal(t.mockVal))
		},

		Entry("successful get", testCase{
			inputName:    "testkvpair",
			expectedCode: http.StatusOK,
			mockError:    nil,
			mockVal: module.KeyValue{
				MetaData: module.KVMetaDataList{
					KeyValueName: "testkvpair",
					Description:  "description",
					UserData1:    "some user data 1",
					UserData2:    "some user data 2",
				},
			},
			kvClient: &mocks.KeyValueManager{},
		}),

		Entry("fails due to not found", testCase{
			inputName:    "testkvpair",
			expectedCode: http.StatusNotFound,
			mockError:    pkgerrors.New("KV Pair does not exist"),
			mockVal:      module.KeyValue{},
			kvClient:     &mocks.KeyValueManager{},
		}),

		Entry("fails due to some other backend error", testCase{
			inputName:    "testkvpair",
			expectedCode: http.StatusInternalServerError,
			mockError:    pkgerrors.New("backend error"),
			mockVal:      module.KeyValue{},
			kvClient:     &mocks.KeyValueManager{},
		}),
	)

	DescribeTable("Delete KeyValue tests",
		func(t testCase) {
			// set up client mock responses
			t.kvClient.On("DeleteKVPair", "test-project", "test-lc", t.inputName).Return(t.mockError)

			// make HTTP request
			request := httptest.NewRequest("DELETE", "/v2/projects/test-project/logical-clouds/test-lc/kv-pairs/"+t.inputName, nil)
			resp := executeRequest(request, NewRouter(t.lcClient, t.clClient, t.upClient, t.quotaClient, t.kvClient))

			// Check returned code
			Expect(resp.StatusCode).To(Equal(t.expectedCode))

			// Check returned body
			got := module.KeyValue{}
			json.NewDecoder(resp.Body).Decode(&got)
			Expect(got).To(Equal(t.mockVal))
		},

		Entry("successful delete", testCase{
			inputName:    "testkvpair",
			expectedCode: http.StatusNoContent,
			mockError:    nil,
			kvClient:     &mocks.KeyValueManager{},
		}),

		// TODO: implement logic and then enable this test:
		// Entry("fails due to not found", testCase{
		// 	inputName:    "testkvpair",
		// 	expectedCode: http.StatusNotFound,
		// 	mockError:    pkgerrors.New("db Remove error - not found"),
		// 	kvClient: &mocks.KeyValueManager{},
		// }),

		// TODO: implement logic and then enable this test:
		// Entry("fails due to a conflict", testCase{
		// 	inputName:    "testkvpair",
		// 	expectedCode: http.StatusConflict,
		// 	mockError:    pkgerrors.New("db Remove error - conflict"),
		// 	kvClient:      &mocks.KeyValueManager{},
		// }),

		Entry("fails due to other backend error", testCase{
			inputName:    "testkvpair",
			expectedCode: http.StatusInternalServerError,
			mockError:    pkgerrors.New("db Remove error - general"),
			kvClient:     &mocks.KeyValueManager{},
		}),
	)
})
