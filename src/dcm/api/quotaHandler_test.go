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

var _ = Describe("QuotaHandler", func() {
	type testCase struct {
		inputName    string
		inputReader  io.Reader
		inStruct     module.Quota
		mockError    error
		mockVal      module.Quota
		mockVals     []module.Quota
		expectedCode int
		lcClient     *mocks.LogicalCloudManager
		clClient     *mocks.ClusterManager
		upClient     *mocks.UserPermissionManager
		quotaClient  *mocks.QuotaManager
		kvClient     *mocks.KeyValueManager
	}

	DescribeTable("Create Quota tests",
		func(t testCase) {
			// set up client mock responses
			t.quotaClient.On("CreateQuota", "test-project", "test-lc", t.inStruct).Return(t.mockVal, t.mockError)

			// make HTTP request
			request := httptest.NewRequest("POST", "/v2/projects/test-project/logical-clouds/test-lc/cluster-quotas", t.inputReader)
			resp := executeRequest(request, NewRouter(t.lcClient, t.clClient, t.upClient, t.quotaClient, t.kvClient))

			// Check returned code
			Expect(resp.StatusCode).To(Equal(t.expectedCode))

			// Check returned body
			got := module.Quota{}
			json.NewDecoder(resp.Body).Decode(&got)
			Expect(got).To(Equal(t.mockVal))
		},

		Entry("successful create", testCase{
			expectedCode: http.StatusCreated,
			inputReader: bytes.NewBuffer([]byte(`{
			"metadata": {
				"name": "testquota",
				"description": "description",
				"userData1": "some user data 1",
				"userData2": "some user data 2"
			},
			"spec" : {
				"limits.cpu": "400",
				"limits.memory": "1000Gi",
				"requests.cpu": "300",
				"requests.memory": "900Gi",
				"requests.storage" : "500Gi",
				"requests.ephemeral-storage": "500",
				"limits.ephemeral-storage": "500",
				"persistentvolumeclaims" : "500",
				"pods": "500",
				"configmaps" : "1000",
				"replicationcontrollers": "500",
				"resourcequotas" : "500",
				"services": "500",
				"services.loadbalancers" : "500",
				"services.nodeports" : "500",
				"secrets" : "500",
				"count/replicationcontrollers" : "500",
				"count/deployments.apps" : "500",
				"count/replicasets.apps" : "500",
				"count/statefulsets.apps" : "500",
				"count/jobs.batch" : "500",
				"count/cronjobs.batch" : "500",
				"count/deployments.extensions" : "500"
			  }
		}`)),
			inStruct: module.Quota{
				MetaData: module.QMetaDataList{
					QuotaName:   "testquota",
					Description: "description",
					UserData1:   "some user data 1",
					UserData2:   "some user data 2",
				},
				Specification: map[string]string{
					"limits.cpu":                   "400",
					"limits.memory":                "1000Gi",
					"requests.cpu":                 "300",
					"requests.memory":              "900Gi",
					"requests.storage":             "500Gi",
					"requests.ephemeral-storage":   "500",
					"limits.ephemeral-storage":     "500",
					"persistentvolumeclaims":       "500",
					"pods":                         "500",
					"configmaps":                   "1000",
					"replicationcontrollers":       "500",
					"resourcequotas":               "500",
					"services":                     "500",
					"services.loadbalancers":       "500",
					"services.nodeports":           "500",
					"secrets":                      "500",
					"count/replicationcontrollers": "500",
					"count/deployments.apps":       "500",
					"count/replicasets.apps":       "500",
					"count/statefulsets.apps":      "500",
					"count/jobs.batch":             "500",
					"count/cronjobs.batch":         "500",
					"count/deployments.extensions": "500",
				},
			},
			mockError: nil,
			mockVal: module.Quota{
				MetaData: module.QMetaDataList{
					QuotaName:   "testquota",
					Description: "description",
					UserData1:   "some user data 1",
					UserData2:   "some user data 2",
				},
			},
			quotaClient: &mocks.QuotaManager{},
		}),

		Entry("fails due to empty body", testCase{
			expectedCode: http.StatusBadRequest,
			inStruct:     module.Quota{},
			mockError:    nil,
			mockVal:      module.Quota{},
			quotaClient:  &mocks.QuotaManager{},
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
			inStruct:    module.Quota{},
			mockError:   nil,
			quotaClient: &mocks.QuotaManager{},
		}),

		// TODO: implement logic and then enable this test:
		// Entry("fails due to other json validation error", testCase{
		// 	// name field has an '=' character
		// 	expectedCode: http.StatusBadRequest,
		// 	inputReader: bytes.NewBuffer([]byte(`{
		// 	"metadata": {
		// 		"name": "test=quota",
		// 		"description": "description",
		// 		"userData1": "some user data 1",
		// 		"userData2": "some user data 2"
		// 	}
		// }`)),
		// 	inStruct:    module.Quota{},
		// 	mockError:   nil,
		// 	quotaClient: &mocks.QuotaManager{},
		// }),

		Entry("fails due to json body decoding error", testCase{
			// extra comma at the end of the userData2 line
			expectedCode: http.StatusUnprocessableEntity,
			inputReader: bytes.NewBuffer([]byte(`{
			"metadata": {
				"name": "testquota",
				"description": "description",
				"userData1": "some user data 1",
				"userData2": "some user data 2",
			}
		}`)),
			inStruct:    module.Quota{},
			mockError:   nil,
			quotaClient: &mocks.QuotaManager{},
		}),

		// TODO: implement logic and then enable this test:
		// Entry("fails due to entry already exists", testCase{
		// 	expectedCode: http.StatusConflict,
		// 	inputReader: bytes.NewBuffer([]byte(`{
		// 	"metadata": {
		// 		"name": "testquota",
		// 		"description": "description",
		// 		"userData1": "some user data 1",
		// 		"userData2": "some user data 2"
		// 	}
		// }`)),
		// 	inStruct: module.Quota{
		// 		MetaData: module.QMetaDataList{
		// 			QuotaName:   "testquota",
		// 			Description: "description",
		// 			UserData1:   "some user data 1",
		// 			UserData2:   "some user data 2",
		// 		},
		// 	},
		// 	mockVal:     module.Quota{},
		// 	mockError:   pkgerrors.New("Quota already exists"),
		// 	quotaClient: &mocks.QuotaManager{},
		// }),

		Entry("fails due to db error", testCase{
			expectedCode: http.StatusInternalServerError,
			inputReader: bytes.NewBuffer([]byte(`{
			"metadata": {
				"name": "testquota",
				"description": "description",
				"userData1": "some user data 1",
				"userData2": "some user data 2"
			}
		}`)),
			inStruct: module.Quota{
				MetaData: module.QMetaDataList{
					QuotaName:   "testquota",
					Description: "description",
					UserData1:   "some user data 1",
					UserData2:   "some user data 2",
				},
			},
			mockVal:     module.Quota{},
			mockError:   pkgerrors.New("Creating DB Entry"),
			quotaClient: &mocks.QuotaManager{},
		}),
	)

	// DCM PUT API currently disabled, so all tests commented out
	// DescribeTable("Put Quota tests",
	// 	func(t testCase) {
	// 		// set up client mock responses
	// 		t.quotaClient.On("UpdateQuota", "test-project", "test-lc", t.inputName, t.inStruct).Return(t.mockVal, t.mockError)

	// 		// make HTTP request
	// 		request := httptest.NewRequest("PUT", "/v2/projects/test-project/logical-clouds/test-lc/cluster-quotas/"+t.inputName, t.inputReader)
	// 		resp := executeRequest(request, NewRouter(t.lcClient, t.clClient, t.upClient, t.quotaClient, t.kvClient))

	// 		//Check returned code
	// 		Expect(resp.StatusCode).To(Equal(t.expectedCode))

	// 		//Check returned body
	// 		got := module.Quota{}
	// 		json.NewDecoder(resp.Body).Decode(&got)
	// 		Expect(got).To(Equal(t.mockVal))
	// 	},

	// 	Entry("successful put", testCase{
	// 		expectedCode: http.StatusOK, // TODO: change to StatusCreated?
	// 		inputName:    "quota",
	// 		inputReader: bytes.NewBuffer([]byte(`{
	// 		"metadata": {
	// 			"name": "quota",
	// 			"description": "description",
	// 			"userData1": "some user data 1",
	// 			"userData2": "some user data 2"
	// 		},
	// 		"spec" : {
	// 			"limits.cpu": "500",
	// 			"limits.memory": "2000Gi"
	// 		}
	// 	}`)),
	// 		inStruct: module.Quota{
	// 			MetaData: module.QMetaDataList{
	// 				QuotaName:   "quota",
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
	// 		mockVal: module.Quota{
	// 			MetaData: module.QMetaDataList{
	// 				QuotaName:   "quota",
	// 				Description: "description",
	// 				UserData1:   "some user data 1",
	// 				UserData2:   "some user data 2",
	// 			},
	// 			Specification: map[string]string{
	// 				"limits.cpu":    "400",
	// 				"limits.memory": "1000Gi",
	// 			},
	// 		},
	// 		quotaClient: &mocks.QuotaManager{},
	// 	}),

	// 	Entry("fails due to empty body", testCase{
	// 		inputName:    "quota",
	// 		expectedCode: http.StatusBadRequest,
	// 		inStruct:     module.Quota{},
	// 		mockError:    nil,
	// 		mockVal:      module.Quota{},
	// 		quotaClient:  &mocks.QuotaManager{},
	// 	}),

	// 	Entry("fails due missing name", testCase{
	// 		inputName:    "quota",
	// 		expectedCode: http.StatusBadRequest,
	// 		inputReader: bytes.NewBuffer([]byte(`{
	// 		"metadata": {
	// 			"description": "description",
	// 			"userData1": "some user data 1",
	// 			"userData2": "some user data 2"
	// 		}
	// 	}`)),
	// 		inStruct:    module.Quota{},
	// 		mockError:   nil,
	// 		quotaClient: &mocks.QuotaManager{},
	// 	}),

	// 	// TODO: implement logic and then enable this test:
	// 	// Entry("fails due to other json validation error", testCase{
	// 	// 	// name field in body has an '=' character
	// 	// 	inputName:    "quota",
	// 	// 	expectedCode: http.StatusBadRequest,
	// 	// 	inputReader: bytes.NewBuffer([]byte(`{
	// 	// 	"metadata": {
	// 	// 		"name": "test=quota",
	// 	// 		"description": "description",
	// 	// 		"userData1": "some user data 1",
	// 	// 		"userData2": "some user data 2"
	// 	// 	}
	// 	// }`)),
	// 	// 	inStruct:    module.Quota{},
	// 	// 	mockError:   nil,
	// 	// 	quotaClient: &mocks.QuotaManager{},
	// 	// }),

	// 	Entry("fails due to json body decoding error", testCase{
	// 		// extra comma at the end of the userData2 line
	// 		inputName:    "quota",
	// 		expectedCode: http.StatusUnprocessableEntity,
	// 		inputReader: bytes.NewBuffer([]byte(`{
	// 		"metadata": {
	// 			"name": "quota",
	// 			"description": "description",
	// 			"userData1": "some user data 1",
	// 			"userData2": "some user data 2",
	// 		}
	// 	}`)),
	// 		inStruct:    module.Quota{},
	// 		mockError:   nil,
	// 		quotaClient: &mocks.QuotaManager{},
	// 	}),

	// 	// TODO: implement logic and then enable this test:
	// 	// Entry("fails due to mismatched name", testCase{
	// 	// 	inputName:    "quotaXYZ",
	// 	// 	expectedCode: http.StatusBadRequest,
	// 	// 	inputReader: bytes.NewBuffer([]byte(`{
	// 	// 	"metadata": {
	// 	// 		"name": "quota",
	// 	// 		"description": "description",
	// 	// 		"userData1": "some user data 1",
	// 	// 		"userData2": "some user data 2"
	// 	// 	}
	// 	// }`)),
	// 	// 	inStruct: module.Quota{
	// 	// 		MetaData: module.QMetaDataList{
	// 	// 			QuotaName:   "quota",
	// 	// 			Description: "description",
	// 	// 			UserData1:   "some user data 1",
	// 	// 			UserData2:   "some user data 2",
	// 	// 		},
	// 	// 	},
	// 	// 	mockVal:     module.Quota{},
	// 	// 	mockError:   pkgerrors.New("Creating DB Entry"),
	// 	// 	quotaClient: &mocks.QuotaManager{},
	// 	// }),

	// 	Entry("fails due to db error", testCase{
	// 		inputName:    "quota",
	// 		expectedCode: http.StatusInternalServerError,
	// 		inputReader: bytes.NewBuffer([]byte(`{
	// 		"metadata": {
	// 			"name": "quota",
	// 			"description": "description",
	// 			"userData1": "some user data 1",
	// 			"userData2": "some user data 2"
	// 		}
	// 	}`)),
	// 		inStruct: module.Quota{
	// 			MetaData: module.QMetaDataList{
	// 				QuotaName:   "quota",
	// 				Description: "description",
	// 				UserData1:   "some user data 1",
	// 				UserData2:   "some user data 2",
	// 			},
	// 		},
	// 		mockVal:     module.Quota{},
	// 		mockError:   pkgerrors.New("Creating DB Entry"),
	// 		quotaClient: &mocks.QuotaManager{},
	// 	}),
	// )

	DescribeTable("Get List Quota tests",
		func(t testCase) {
			// set up client mock responses
			t.quotaClient.On("GetAllQuotas", "test-project", "test-lc").Return(t.mockVals, t.mockError)

			// make HTTP request
			request := httptest.NewRequest("GET", "/v2/projects/test-project/logical-clouds/test-lc/cluster-quotas", nil)
			resp := executeRequest(request, NewRouter(t.lcClient, t.clClient, t.upClient, t.quotaClient, t.kvClient))

			// Check returned code
			Expect(resp.StatusCode).To(Equal(t.expectedCode))

			// Check returned body
			got := []module.Quota{}
			json.NewDecoder(resp.Body).Decode(&got)
			Expect(got).To(Equal(t.mockVals))
		},

		Entry("successful get", testCase{
			expectedCode: http.StatusOK,
			mockError:    nil,
			mockVals: []module.Quota{
				{
					MetaData: module.QMetaDataList{
						QuotaName:   "testquota1",
						Description: "description",
						UserData1:   "some user data 1",
						UserData2:   "some user data 2",
					},
				},
				{
					MetaData: module.QMetaDataList{
						QuotaName:   "testquota2",
						Description: "description",
						UserData1:   "some user data 1",
						UserData2:   "some user data 2",
					},
				},
			},
			quotaClient: &mocks.QuotaManager{},
		}),

		Entry("fails due to some other backend error", testCase{
			expectedCode: http.StatusInternalServerError,
			mockError:    pkgerrors.New("backend error"),
			mockVals:     []module.Quota{},
			quotaClient:  &mocks.QuotaManager{},
		}),
	)

	DescribeTable("Get Quota tests",
		func(t testCase) {
			// set up client mock responses
			t.quotaClient.On("GetQuota", "test-project", "test-lc", t.inputName).Return(t.mockVal, t.mockError)

			// make HTTP request
			request := httptest.NewRequest("GET", "/v2/projects/test-project/logical-clouds/test-lc/cluster-quotas/"+t.inputName, nil)
			resp := executeRequest(request, NewRouter(t.lcClient, t.clClient, t.upClient, t.quotaClient, t.kvClient))

			// Check returned code
			Expect(resp.StatusCode).To(Equal(t.expectedCode))

			// Check returned body
			got := module.Quota{}
			json.NewDecoder(resp.Body).Decode(&got)
			Expect(got).To(Equal(t.mockVal))
		},

		Entry("successful get", testCase{
			inputName:    "testquota",
			expectedCode: http.StatusOK,
			mockError:    nil,
			mockVal: module.Quota{
				MetaData: module.QMetaDataList{
					QuotaName:   "testquota",
					Description: "description",
					UserData1:   "some user data 1",
					UserData2:   "some user data 2",
				},
			},
			quotaClient: &mocks.QuotaManager{},
		}),

		Entry("fails due to not found", testCase{
			inputName:    "testquota",
			expectedCode: http.StatusNotFound,
			mockError:    pkgerrors.New("Cluster Quota does not exist"),
			mockVal:      module.Quota{},
			quotaClient:  &mocks.QuotaManager{},
		}),

		Entry("fails due to some other backend error", testCase{
			inputName:    "testquota",
			expectedCode: http.StatusInternalServerError,
			mockError:    pkgerrors.New("backend error"),
			mockVal:      module.Quota{},
			quotaClient:  &mocks.QuotaManager{},
		}),
	)

	DescribeTable("Delete Quota tests",
		func(t testCase) {
			// set up client mock responses
			t.quotaClient.On("DeleteQuota", "test-project", "test-lc", t.inputName).Return(t.mockError)

			// make HTTP request
			request := httptest.NewRequest("DELETE", "/v2/projects/test-project/logical-clouds/test-lc/cluster-quotas/"+t.inputName, nil)
			resp := executeRequest(request, NewRouter(t.lcClient, t.clClient, t.upClient, t.quotaClient, t.kvClient))

			// Check returned code
			Expect(resp.StatusCode).To(Equal(t.expectedCode))

			// Check returned body
			got := module.Quota{}
			json.NewDecoder(resp.Body).Decode(&got)
			Expect(got).To(Equal(t.mockVal))
		},

		Entry("successful delete", testCase{
			inputName:    "testquota",
			expectedCode: http.StatusNoContent,
			mockError:    nil,
			quotaClient:  &mocks.QuotaManager{},
		}),

		// TODO: implement logic and then enable this test:
		// Entry("fails due to not found", testCase{
		// 	inputName:    "testquota",
		// 	expectedCode: http.StatusNotFound,
		// 	mockError:    pkgerrors.New("db Remove error - not found"),
		// 	quotaClient:  &mocks.QuotaManager{},
		// }),

		// TODO: implement logic and then enable this test:
		// Entry("fails due to a conflict", testCase{
		// 	inputName:    "testquota",
		// 	expectedCode: http.StatusConflict,
		// 	mockError:    pkgerrors.New("db Remove error - conflict"),
		// 	quotaClient:       &mocks.QuotaManager{},
		// }),

		Entry("fails due to other backend error", testCase{
			inputName:    "testquota",
			expectedCode: http.StatusInternalServerError,
			mockError:    pkgerrors.New("db Remove error - general"),
			quotaClient:  &mocks.QuotaManager{},
		}),
	)
})
