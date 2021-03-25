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

var _ = Describe("ClusterHandler", func() {
	type testCase struct {
		inputName    string
		inputReader  io.Reader
		inStruct     module.Cluster
		mockError    error
		mockVal      module.Cluster
		mockVals     []module.Cluster
		expectedCode int
		lcClient     *mocks.LogicalCloudManager
		clClient     *mocks.ClusterManager
		upClient     *mocks.UserPermissionManager
		quotaClient  *mocks.QuotaManager
		kvClient     *mocks.KeyValueManager
	}

	DescribeTable("Create Cluster tests",
		func(t testCase) {
			// set up client mock responses
			t.clClient.On("CreateCluster", "test-project", "test-lc", t.inStruct).Return(t.mockVal, t.mockError)

			// make HTTP request
			request := httptest.NewRequest("POST", "/v2/projects/test-project/logical-clouds/test-lc/cluster-references", t.inputReader)
			resp := executeRequest(request, NewRouter(t.lcClient, t.clClient, t.upClient, t.quotaClient, t.kvClient))

			// Check returned code
			Expect(resp.StatusCode).To(Equal(t.expectedCode))

			// Check returned body
			got := module.Cluster{}
			json.NewDecoder(resp.Body).Decode(&got)
			Expect(got).To(Equal(t.mockVal))
		},

		Entry("successful create", testCase{
			expectedCode: http.StatusCreated,
			inputReader: bytes.NewBuffer([]byte(`{
			"metadata": {
				"name": "testcluster",
				"description": "description",
				"userData1": "some user data 1",
				"userData2": "some user data 2"
			}
		}`)),
			inStruct: module.Cluster{
				MetaData: module.ClusterMeta{
					ClusterReference: "testcluster",
					Description:      "description",
					UserData1:        "some user data 1",
					UserData2:        "some user data 2",
				},
			},
			mockError: nil,
			mockVal: module.Cluster{
				MetaData: module.ClusterMeta{
					ClusterReference: "testcluster",
					Description:      "description",
					UserData1:        "some user data 1",
					UserData2:        "some user data 2",
				},
			},
			clClient: &mocks.ClusterManager{},
		}),

		Entry("fails due to empty body", testCase{
			expectedCode: http.StatusBadRequest,
			inStruct:     module.Cluster{},
			mockError:    nil,
			mockVal:      module.Cluster{},
			clClient:     &mocks.ClusterManager{},
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
			inStruct:  module.Cluster{},
			mockError: nil,
			clClient:  &mocks.ClusterManager{},
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
		// 	inStruct:    module.Cluster{},
		// 	mockError:   nil,
		// 	clClient:&mocks.ClusterManager{},
		// }),

		Entry("fails due to json body decoding error", testCase{
			// extra comma at the end of the userData2 line
			expectedCode: http.StatusUnprocessableEntity,
			inputReader: bytes.NewBuffer([]byte(`{
			"metadata": {
				"name": "testcluster",
				"description": "description",
				"userData1": "some user data 1",
				"userData2": "some user data 2",
			}
		}`)),
			inStruct:  module.Cluster{},
			mockError: nil,
			clClient:  &mocks.ClusterManager{},
		}),

		// TODO: implement logic and then enable this test:
		// Entry("fails due to entry already exists", testCase{
		// 	expectedCode: http.StatusConflict,
		// 	inputReader: bytes.NewBuffer([]byte(`{
		// 	"metadata": {
		// 		"name": "testcluster",
		// 		"description": "description",
		// 		"userData1": "some user data 1",
		// 		"userData2": "some user data 2"
		// 	}
		// }`)),
		// 	inStruct: module.Cluster{
		// 		MetaData: module.ClusterMeta{
		// 			ClusterReference:   "testcluster",
		// 			Description: "description",
		// 			UserData1:   "some user data 1",
		// 			UserData2:   "some user data 2",
		// 		},
		// 	},
		// 	mockVal:     module.Cluster{},
		// 	mockError:   pkgerrors.New("KeyValue already exists"),
		// 	clClient:&mocks.ClusterManager{},
		// }),

		Entry("fails due to db error", testCase{
			expectedCode: http.StatusInternalServerError,
			inputReader: bytes.NewBuffer([]byte(`{
			"metadata": {
				"name": "testcluster",
				"description": "description",
				"userData1": "some user data 1",
				"userData2": "some user data 2"
			}
		}`)),
			inStruct: module.Cluster{
				MetaData: module.ClusterMeta{
					ClusterReference: "testcluster",
					Description:      "description",
					UserData1:        "some user data 1",
					UserData2:        "some user data 2",
				},
			},
			mockVal:   module.Cluster{},
			mockError: pkgerrors.New("Creating DB Entry"),
			clClient:  &mocks.ClusterManager{},
		}),
	)

	// DCM PUT API currently disabled, so all tests commented out
	// DescribeTable("Put Cluster tests",
	// 	func(t testCase) {
	// 		// set up client mock responses
	// 		t.clClient.On("UpdateCluster", "test-project", "test-lc", t.inputName, t.inStruct).Return(t.mockVal, t.mockError)

	// 		// make HTTP request
	// 		request := httptest.NewRequest("PUT", "/v2/projects/test-project/logical-clouds/test-lc/cluster-references/"+t.inputName, t.inputReader)
	// 		resp := executeRequest(request, NewRouter(t.lcClient, t.clClient, t.upClient, t.quotaClient, t.kvClient))

	// 		//Check returned code
	// 		Expect(resp.StatusCode).To(Equal(t.expectedCode))

	// 		//Check returned body
	// 		got := module.Cluster{}
	// 		json.NewDecoder(resp.Body).Decode(&got)
	// 		Expect(got).To(Equal(t.mockVal))
	// 	},

	// 	Entry("successful put", testCase{
	// 		expectedCode: http.StatusOK, // TODO: change to StatusCreated?
	// 		inputName:    "cluster",
	// 		inputReader: bytes.NewBuffer([]byte(`{
	// 		"metadata": {
	// 			"name": "cluster",
	// 			"description": "description",
	// 			"userData1": "some user data 1",
	// 			"userData2": "some user data 2"
	// 		},
	// 		"spec" : {
	// 			"limits.cpu": "500",
	// 			"limits.memory": "2000Gi"
	// 		}
	// 	}`)),
	// 		inStruct: module.Cluster{
	// 			MetaData: module.ClusterMeta{
	// 				ClusterReference:   "cluster",
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
	// 		mockVal: module.Cluster{
	// 			MetaData: module.ClusterMeta{
	// 				ClusterReference:   "cluster",
	// 				Description: "description",
	// 				UserData1:   "some user data 1",
	// 				UserData2:   "some user data 2",
	// 			},
	// 			Specification: map[string]string{
	// 				"limits.cpu":    "400",
	// 				"limits.memory": "1000Gi",
	// 			},
	// 		},
	// 		clClient:&mocks.ClusterManager{},
	// 	}),

	// 	Entry("fails due to empty body", testCase{
	// 		inputName:    "cluster",
	// 		expectedCode: http.StatusBadRequest,
	// 		inStruct:     module.Cluster{},
	// 		mockError:    nil,
	// 		mockVal:      module.Cluster{},
	// 		clClient: &mocks.ClusterManager{},
	// 	}),

	// 	Entry("fails due missing name", testCase{
	// 		inputName:    "cluster",
	// 		expectedCode: http.StatusBadRequest,
	// 		inputReader: bytes.NewBuffer([]byte(`{
	// 		"metadata": {
	// 			"description": "description",
	// 			"userData1": "some user data 1",
	// 			"userData2": "some user data 2"
	// 		}
	// 	}`)),
	// 		inStruct:    module.Cluster{},
	// 		mockError:   nil,
	// 		clClient:&mocks.ClusterManager{},
	// 	}),

	// 	// TODO: implement logic and then enable this test:
	// 	// Entry("fails due to other json validation error", testCase{
	// 	// 	// name field in body has an '=' character
	// 	// 	inputName:    "cluster",
	// 	// 	expectedCode: http.StatusBadRequest,
	// 	// 	inputReader: bytes.NewBuffer([]byte(`{
	// 	// 	"metadata": {
	// 	// 		"name": "test=cluster",
	// 	// 		"description": "description",
	// 	// 		"userData1": "some user data 1",
	// 	// 		"userData2": "some user data 2"
	// 	// 	}
	// 	// }`)),
	// 	// 	inStruct:    module.Cluster{},
	// 	// 	mockError:   nil,
	// 	// 	clClient:&mocks.ClusterManager{},
	// 	// }),

	// 	Entry("fails due to json body decoding error", testCase{
	// 		// extra comma at the end of the userData2 line
	// 		inputName:    "cluster",
	// 		expectedCode: http.StatusUnprocessableEntity,
	// 		inputReader: bytes.NewBuffer([]byte(`{
	// 		"metadata": {
	// 			"name": "cluster",
	// 			"description": "description",
	// 			"userData1": "some user data 1",
	// 			"userData2": "some user data 2",
	// 		}
	// 	}`)),
	// 		inStruct:    module.Cluster{},
	// 		mockError:   nil,
	// 		clClient:&mocks.ClusterManager{},
	// 	}),

	// 	// TODO: implement logic and then enable this test:
	// 	// Entry("fails due to mismatched name", testCase{
	// 	// 	inputName:    "quotaXYZ",
	// 	// 	expectedCode: http.StatusBadRequest,
	// 	// 	inputReader: bytes.NewBuffer([]byte(`{
	// 	// 	"metadata": {
	// 	// 		"name": "cluster",
	// 	// 		"description": "description",
	// 	// 		"userData1": "some user data 1",
	// 	// 		"userData2": "some user data 2"
	// 	// 	}
	// 	// }`)),
	// 	// 	inStruct: module.Cluster{
	// 	// 		MetaData: module.ClusterMeta{
	// 	// 			ClusterReference:   "cluster",
	// 	// 			Description: "description",
	// 	// 			UserData1:   "some user data 1",
	// 	// 			UserData2:   "some user data 2",
	// 	// 		},
	// 	// 	},
	// 	// 	mockVal:     module.Cluster{},
	// 	// 	mockError:   pkgerrors.New("Creating DB Entry"),
	// 	// 	clClient:&mocks.ClusterManager{},
	// 	// }),

	// 	Entry("fails due to db error", testCase{
	// 		inputName:    "cluster",
	// 		expectedCode: http.StatusInternalServerError,
	// 		inputReader: bytes.NewBuffer([]byte(`{
	// 		"metadata": {
	// 			"name": "cluster",
	// 			"description": "description",
	// 			"userData1": "some user data 1",
	// 			"userData2": "some user data 2"
	// 		}
	// 	}`)),
	// 		inStruct: module.Cluster{
	// 			MetaData: module.ClusterMeta{
	// 				ClusterReference:   "cluster",
	// 				Description: "description",
	// 				UserData1:   "some user data 1",
	// 				UserData2:   "some user data 2",
	// 			},
	// 		},
	// 		mockVal:     module.Cluster{},
	// 		mockError:   pkgerrors.New("Creating DB Entry"),
	// 		clClient:&mocks.ClusterManager{},
	// 	}),
	// )

	DescribeTable("Get List Cluster tests",
		func(t testCase) {
			// set up client mock responses
			t.clClient.On("GetAllClusters", "test-project", "test-lc").Return(t.mockVals, t.mockError)

			// make HTTP request
			request := httptest.NewRequest("GET", "/v2/projects/test-project/logical-clouds/test-lc/cluster-references", nil)
			resp := executeRequest(request, NewRouter(t.lcClient, t.clClient, t.upClient, t.quotaClient, t.kvClient))

			// Check returned code
			Expect(resp.StatusCode).To(Equal(t.expectedCode))

			// Check returned body
			got := []module.Cluster{}
			json.NewDecoder(resp.Body).Decode(&got)
			Expect(got).To(Equal(t.mockVals))
		},

		Entry("successful get", testCase{
			expectedCode: http.StatusOK,
			mockError:    nil,
			mockVals: []module.Cluster{
				{
					MetaData: module.ClusterMeta{
						ClusterReference: "testcluster1",
						Description:      "description",
						UserData1:        "some user data 1",
						UserData2:        "some user data 2",
					},
				},
				{
					MetaData: module.ClusterMeta{
						ClusterReference: "testcluster2",
						Description:      "description",
						UserData1:        "some user data 1",
						UserData2:        "some user data 2",
					},
				},
			},
			clClient: &mocks.ClusterManager{},
		}),

		Entry("fails due to some other backend error", testCase{
			expectedCode: http.StatusInternalServerError,
			mockError:    pkgerrors.New("backend error"),
			mockVals:     []module.Cluster{},
			clClient:     &mocks.ClusterManager{},
		}),
	)

	DescribeTable("Get Cluster tests",
		func(t testCase) {
			// set up client mock responses
			t.clClient.On("GetCluster", "test-project", "test-lc", t.inputName).Return(t.mockVal, t.mockError)

			// make HTTP request
			request := httptest.NewRequest("GET", "/v2/projects/test-project/logical-clouds/test-lc/cluster-references/"+t.inputName, nil)
			resp := executeRequest(request, NewRouter(t.lcClient, t.clClient, t.upClient, t.quotaClient, t.kvClient))

			// Check returned code
			Expect(resp.StatusCode).To(Equal(t.expectedCode))

			// Check returned body
			got := module.Cluster{}
			json.NewDecoder(resp.Body).Decode(&got)
			Expect(got).To(Equal(t.mockVal))
		},

		Entry("successful get", testCase{
			inputName:    "testcluster",
			expectedCode: http.StatusOK,
			mockError:    nil,
			mockVal: module.Cluster{
				MetaData: module.ClusterMeta{
					ClusterReference: "testcluster",
					Description:      "description",
					UserData1:        "some user data 1",
					UserData2:        "some user data 2",
				},
			},
			clClient: &mocks.ClusterManager{},
		}),

		Entry("fails due to not found", testCase{
			inputName:    "testcluster",
			expectedCode: http.StatusNotFound,
			mockError:    pkgerrors.New("Cluster Reference does not exist"),
			mockVal:      module.Cluster{},
			clClient:     &mocks.ClusterManager{},
		}),

		Entry("fails due to some other backend error", testCase{
			inputName:    "testcluster",
			expectedCode: http.StatusInternalServerError,
			mockError:    pkgerrors.New("backend error"),
			mockVal:      module.Cluster{},
			clClient:     &mocks.ClusterManager{},
		}),
	)

	DescribeTable("Delete Cluster tests",
		func(t testCase) {
			// set up client mock responses
			t.clClient.On("DeleteCluster", "test-project", "test-lc", t.inputName).Return(t.mockError)

			// make HTTP request
			request := httptest.NewRequest("DELETE", "/v2/projects/test-project/logical-clouds/test-lc/cluster-references/"+t.inputName, nil)
			resp := executeRequest(request, NewRouter(t.lcClient, t.clClient, t.upClient, t.quotaClient, t.kvClient))

			// Check returned code
			Expect(resp.StatusCode).To(Equal(t.expectedCode))

			// Check returned body
			got := module.Cluster{}
			json.NewDecoder(resp.Body).Decode(&got)
			Expect(got).To(Equal(t.mockVal))
		},

		Entry("successful delete", testCase{
			inputName:    "testcluster",
			expectedCode: http.StatusNoContent,
			mockError:    nil,
			clClient:     &mocks.ClusterManager{},
		}),

		// TODO: implement logic and then enable this test:
		// Entry("fails due to not found", testCase{
		// 	inputName:    "testcluster",
		// 	expectedCode: http.StatusNotFound,
		// 	mockError:    pkgerrors.New("db Remove error - not found"),
		// 	clClient: &mocks.ClusterManager{},
		// }),

		// TODO: implement logic and then enable this test:
		// Entry("fails due to a conflict", testCase{
		// 	inputName:    "testcluster",
		// 	expectedCode: http.StatusConflict,
		// 	mockError:    pkgerrors.New("db Remove error - conflict"),
		// 	clClient:      &mocks.ClusterManager{},
		// }),

		Entry("fails due to other backend error", testCase{
			inputName:    "testcluster",
			expectedCode: http.StatusInternalServerError,
			mockError:    pkgerrors.New("db Remove error - general"),
			clClient:     &mocks.ClusterManager{},
		}),
	)
})
