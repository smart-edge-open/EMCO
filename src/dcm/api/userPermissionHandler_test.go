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

var _ = Describe("UserPermissionHandler", func() {
	type testCase struct {
		inputName    string
		inputReader  io.Reader
		inStruct     module.UserPermission
		mockError    error
		mockVal      module.UserPermission
		mockVals     []module.UserPermission
		expectedCode int
		lcClient     *mocks.LogicalCloudManager
		clClient     *mocks.ClusterManager
		upClient     *mocks.UserPermissionManager
		quotaClient  *mocks.QuotaManager
		kvClient     *mocks.KeyValueManager
	}

	DescribeTable("Create UserPermission tests",
		func(t testCase) {
			// set up client mock responses
			t.upClient.On("CreateUserPerm", "test-project", "test-lc", t.inStruct).Return(t.mockVal, t.mockError)

			// make HTTP request
			request := httptest.NewRequest("POST", "/v2/projects/test-project/logical-clouds/test-lc/user-permissions", t.inputReader)
			resp := executeRequest(request, NewRouter(t.lcClient, t.clClient, t.upClient, t.quotaClient, t.kvClient))

			// Check returned code
			Expect(resp.StatusCode).To(Equal(t.expectedCode))

			// Check returned body
			got := module.UserPermission{}
			json.NewDecoder(resp.Body).Decode(&got)
			Expect(got).To(Equal(t.mockVal))
		},

		Entry("successful create", testCase{
			expectedCode: http.StatusCreated,
			inputReader: bytes.NewBuffer([]byte(`{
				"metadata": {
					"name" : "testuserpermission"
				},
				"spec": {
					"namespace": "ns1",
					"apiGroups": ["", "apps"],
					"resources": ["secrets", "pods"],
					"verbs": ["get", "watch", "list", "create"]
				}
		}`)),
			inStruct: module.UserPermission{
				MetaData: module.UPMetaDataList{
					UserPermissionName: "testuserpermission",
				},
				Specification: module.UPSpec{
					Namespace: "ns1",
					APIGroups: []string{"", "apps"},
					Resources: []string{"secrets", "pods"},
					Verbs:     []string{"get", "watch", "list", "create"},
				},
			},
			mockError: nil,
			mockVal: module.UserPermission{
				MetaData: module.UPMetaDataList{
					UserPermissionName: "testuserpermission",
				},
				Specification: module.UPSpec{
					Namespace: "ns1",
					APIGroups: []string{"", "apps"},
					Resources: []string{"secrets", "pods"},
					Verbs:     []string{"get", "watch", "list", "create"},
				},
			},
			upClient: &mocks.UserPermissionManager{},
		}),

		Entry("fails due to empty body", testCase{
			expectedCode: http.StatusBadRequest,
			inStruct:     module.UserPermission{},
			mockError:    nil,
			mockVal:      module.UserPermission{},
			upClient:     &mocks.UserPermissionManager{},
		}),

		Entry("fails due missing name", testCase{
			expectedCode: http.StatusBadRequest,
			inputReader: bytes.NewBuffer([]byte(`{
				"metadata": {
					"name" : ""
				},
				"spec": {
					"namespace": "ns1",
					"apiGroups": ["", "apps"],
					"resources": ["secrets", "pods"],
					"verbs": ["get", "watch", "list", "create"]
				}
		}`)),
			inStruct:  module.UserPermission{},
			mockError: nil,
			upClient:  &mocks.UserPermissionManager{},
		}),

		// TODO: implement logic and then enable this test:
		// Entry("fails due to other json validation error", testCase{
		// 	// name field has an '=' character
		// 	expectedCode: http.StatusBadRequest,
		// 	inputReader: bytes.NewBuffer([]byte(`{
		// 		"name": "test=userpermission",
		// "apiGroups" : ["", "apps"],
		// "resources" : ["secrets", "pods"],
		// "verbs" : ["get", "watch", "list", "create"]
		// }`)),
		// 	inStruct:    module.UserPermission{},
		// 	mockError:   nil,
		// 	upClient: &mocks.UserPermissionManager{},
		// }),

		Entry("fails due to json body decoding error", testCase{
			// extra comma at the end
			expectedCode: http.StatusUnprocessableEntity,
			inputReader: bytes.NewBuffer([]byte(`{
				"metadata": {
					"name" : "testuserpermission"
				},
				"spec": {
					"namespace": "ns1",
					"apiGroups": ["", "apps"],
					"resources": ["secrets", "pods"],
					"verbs": ["get", "watch", "list", "create"]
				},
		}`)),
			inStruct:  module.UserPermission{},
			mockError: nil,
			upClient:  &mocks.UserPermissionManager{},
		}),

		// TODO: implement logic and then enable this test:
		// Entry("fails due to entry already exists", testCase{
		// 	expectedCode: http.StatusConflict,
		// 	inputReader: bytes.NewBuffer([]byte(`{
		// 	"metadata": {
		// 		"name": "testuserpermission",
		// "apiGroups" : ["", "apps"],
		// "resources" : ["secrets", "pods"],
		// "verbs" : ["get", "watch", "list", "create"]
		// 	}
		// }`)),
		// 	inStruct: module.UserPermission{
		// UserPermissionName: "testuserpermission",
		// APIGroups:          []string{"description"},
		// Resources:          []string{"some user data 1"},
		// Verbs:              []string{"some user data 2"},
		// 	},
		// 	mockVal:     module.UserPermission{},
		// 	mockError:   pkgerrors.New("UserPermission already exists"),
		// 	upClient: &mocks.UserPermissionManager{},
		// }),

		Entry("fails due to db error", testCase{
			expectedCode: http.StatusInternalServerError,
			inputReader: bytes.NewBuffer([]byte(`{
				"metadata": {
					"name" : "testuserpermission"
				},
				"spec": {
					"namespace": "ns1",
					"apiGroups": ["", "apps"],
					"resources": ["secrets", "pods"],
					"verbs": ["get", "watch", "list", "create"]
				}
		}`)),
			inStruct: module.UserPermission{
				MetaData: module.UPMetaDataList{
					UserPermissionName: "testuserpermission",
				},
				Specification: module.UPSpec{
					Namespace: "ns1",
					APIGroups: []string{"", "apps"},
					Resources: []string{"secrets", "pods"},
					Verbs:     []string{"get", "watch", "list", "create"},
				},
			},
			mockVal:   module.UserPermission{},
			mockError: pkgerrors.New("Creating DB Entry"),
			upClient:  &mocks.UserPermissionManager{},
		}),
	)

	// DCM PUT API currently disabled, so all tests commented out
	// DescribeTable("Put UserPermission tests",
	// 	func(t testCase) {
	// 		// set up client mock responses
	// 		t.upClient.On("UpdateUserPerm", "test-project", "test-lc", t.inputName, t.inStruct).Return(t.mockVal, t.mockError)

	// 		// make HTTP request
	// 		request := httptest.NewRequest("PUT", "/v2/projects/test-project/logical-clouds/test-lc/user-permissions/"+t.inputName, t.inputReader)
	// 		resp := executeRequest(request, NewRouter(t.lcClient, t.clClient,t.upClient, t.quotaClient, t.kvClient))

	// 		//Check returned code
	// 		Expect(resp.StatusCode).To(Equal(t.expectedCode))

	// 		//Check returned body
	// 		got := module.UserPermission{}
	// 		json.NewDecoder(resp.Body).Decode(&got)
	// 		Expect(got).To(Equal(t.mockVal))
	// 	},

	// 	Entry("successful put", testCase{
	// 		expectedCode: http.StatusOK, // TODO: change to StatusCreated?
	// 		inputName:    "userpermission",
	// 		inputReader: bytes.NewBuffer([]byte(`{
	// 			"name": "userpermission",
	// "apiGroups" : ["", "apps"],
	// "resources" : ["secrets", "pods"],
	// "verbs" : ["get", "watch", "list", "create"]
	// 		"spec" : {
	// 			"limits.cpu": "500",
	// 			"limits.memory": "2000Gi"
	// 		}
	// 	}`)),
	// 		inStruct: module.UserPermission{
	// UserPermissionName: "testuserpermission",
	// APIGroups:          []string{"", "apps"},
	// Resources:          []string{"secrets", "pods"},
	// Verbs:              []string{"get", "watch", "list", "create"},
	// 		},
	// 		mockError: nil,
	// 		mockVal: module.UserPermission{
	// UserPermissionName: "testuserpermission",
	// APIGroups:          []string{"", "apps"},
	// Resources:          []string{"secrets", "pods"},
	// Verbs:              []string{"get", "watch", "list", "create"},
	// 		},
	// 		upClient: &mocks.UserPermissionManager{},
	// 	}),

	// 	Entry("fails due to empty body", testCase{
	// 		inputName:    "userpermission",
	// 		expectedCode: http.StatusBadRequest,
	// 		inStruct:     module.UserPermission{},
	// 		mockError:    nil,
	// 		mockVal:      module.UserPermission{},
	// 		upClient:  &mocks.UserPermissionManager{},
	// 	}),

	// 	Entry("fails due missing name", testCase{
	// 		inputName:    "userpermission",
	// 		expectedCode: http.StatusBadRequest,
	// 		inputReader: bytes.NewBuffer([]byte(`{
	// "apiGroups" : ["", "apps"],
	// "resources" : ["secrets", "pods"],
	// "verbs" : ["get", "watch", "list", "create"]
	// 	}`)),
	// 		inStruct:    module.UserPermission{},
	// 		mockError:   nil,
	// 		upClient: &mocks.UserPermissionManager{},
	// 	}),

	// 	// TODO: implement logic and then enable this test:
	// 	// Entry("fails due to other json validation error", testCase{
	// 	// 	// name field in body has an '=' character
	// 	// 	inputName:    "userpermission",
	// 	// 	expectedCode: http.StatusBadRequest,
	// 	// 	inputReader: bytes.NewBuffer([]byte(`{
	// 	// 		"name": "test=userpermission",
	// "apiGroups" : ["", "apps"],
	// "resources" : ["secrets", "pods"],
	// "verbs" : ["get", "watch", "list", "create"]
	// 	// }`)),
	// 	// 	inStruct:    module.UserPermission{},
	// 	// 	mockError:   nil,
	// 	// 	upClient: &mocks.UserPermissionManager{},
	// 	// }),

	// 	Entry("fails due to json body decoding error", testCase{
	// 		// extra comma at the end of the userData2 line
	// 		inputName:    "userpermission",
	// 		expectedCode: http.StatusUnprocessableEntity,
	// 		inputReader: bytes.NewBuffer([]byte(`{
	// 			"name": "userpermission",
	// "apiGroups" : ["", "apps"],
	// "resources" : ["secrets", "pods"],
	// "verbs" : ["get", "watch", "list", "create"]
	// 	}`)),
	// 		inStruct:    module.UserPermission{},
	// 		mockError:   nil,
	// 		upClient: &mocks.UserPermissionManager{},
	// 	}),

	// 	// TODO: implement logic and then enable this test:
	// 	// Entry("fails due to mismatched name", testCase{
	// 	// 	inputName:    "userpermissionXYZ",
	// 	// 	expectedCode: http.StatusBadRequest,
	// 	// 	inputReader: bytes.NewBuffer([]byte(`{
	// 	// 		"name": "userpermission",
	// "apiGroups" : ["", "apps"],
	// "resources" : ["secrets", "pods"],
	// "verbs" : ["get", "watch", "list", "create"]
	// 	// }`)),
	// 	// 	inStruct: module.UserPermission{
	// UserPermissionName: "testuserpermission",
	// APIGroups:          []string{"", "apps"},
	// Resources:          []string{"secrets", "pods"},
	// Verbs:              []string{"get", "watch", "list", "create"},
	// 	// 	},
	// 	// 	mockVal:     module.UserPermission{},
	// 	// 	mockError:   pkgerrors.New("Creating DB Entry"),
	// 	// 	upClient: &mocks.UserPermissionManager{},
	// 	// }),

	// 	Entry("fails due to db error", testCase{
	// 		inputName:    "userpermission",
	// 		expectedCode: http.StatusInternalServerError,
	// 		inputReader: bytes.NewBuffer([]byte(`{
	// 			"name": "userpermission",
	// "apiGroups" : ["", "apps"],
	// "resources" : ["secrets", "pods"],
	// "verbs" : ["get", "watch", "list", "create"]
	// 	}`)),
	// 		inStruct: module.UserPermission{
	// UserPermissionName: "testuserpermission",
	// APIGroups:          []string{"", "apps"},
	// Resources:          []string{"secrets", "pods"},
	// Verbs:              []string{"get", "watch", "list", "create"},
	// 		},
	// 		mockVal:     module.UserPermission{},
	// 		mockError:   pkgerrors.New("Creating DB Entry"),
	// 		upClient: &mocks.UserPermissionManager{},
	// 	}),
	// )

	DescribeTable("Get List UserPermission tests",
		func(t testCase) {
			// set up client mock responses
			t.upClient.On("GetAllUserPerms", "test-project", "test-lc").Return(t.mockVals, t.mockError)

			// make HTTP request
			request := httptest.NewRequest("GET", "/v2/projects/test-project/logical-clouds/test-lc/user-permissions", nil)
			resp := executeRequest(request, NewRouter(t.lcClient, t.clClient, t.upClient, t.quotaClient, t.kvClient))

			// Check returned code
			Expect(resp.StatusCode).To(Equal(t.expectedCode))

			// Check returned body
			got := []module.UserPermission{}
			json.NewDecoder(resp.Body).Decode(&got)
			Expect(got).To(Equal(t.mockVals))
		},

		Entry("successful get", testCase{
			expectedCode: http.StatusOK,
			mockError:    nil,
			mockVals: []module.UserPermission{
				{
					MetaData: module.UPMetaDataList{
						UserPermissionName: "testuserpermission1",
					},
					Specification: module.UPSpec{
						Namespace: "ns1",
						APIGroups: []string{"", "apps"},
						Resources: []string{"secrets", "pods"},
						Verbs:     []string{"get", "watch", "list", "create"},
					},
				},
				{
					MetaData: module.UPMetaDataList{
						UserPermissionName: "testuserpermission2",
					},
					Specification: module.UPSpec{
						Namespace: "ns1",
						APIGroups: []string{"", "apps"},
						Resources: []string{"secrets", "pods"},
						Verbs:     []string{"get", "watch", "list", "create"},
					},
				},
			},
			upClient: &mocks.UserPermissionManager{},
		}),

		Entry("fails due to some other backend error", testCase{
			expectedCode: http.StatusInternalServerError,
			mockError:    pkgerrors.New("backend error"),
			mockVals:     []module.UserPermission{},
			upClient:     &mocks.UserPermissionManager{},
		}),
	)

	DescribeTable("Get UserPermission tests",
		func(t testCase) {
			// set up client mock responses
			t.upClient.On("GetUserPerm", "test-project", "test-lc", t.inputName).Return(t.mockVal, t.mockError)

			// make HTTP request
			request := httptest.NewRequest("GET", "/v2/projects/test-project/logical-clouds/test-lc/user-permissions/"+t.inputName, nil)
			resp := executeRequest(request, NewRouter(t.lcClient, t.clClient, t.upClient, t.quotaClient, t.kvClient))

			// Check returned code
			Expect(resp.StatusCode).To(Equal(t.expectedCode))

			// Check returned body
			got := module.UserPermission{}
			json.NewDecoder(resp.Body).Decode(&got)
			Expect(got).To(Equal(t.mockVal))
		},

		Entry("successful get", testCase{
			inputName:    "testuserpermission",
			expectedCode: http.StatusOK,
			mockError:    nil,
			mockVal: module.UserPermission{
				MetaData: module.UPMetaDataList{
					UserPermissionName: "testuserpermission",
				},
				Specification: module.UPSpec{
					Namespace: "ns1",
					APIGroups: []string{"", "apps"},
					Resources: []string{"secrets", "pods"},
					Verbs:     []string{"get", "watch", "list", "create"},
				},
			},
			upClient: &mocks.UserPermissionManager{},
		}),

		Entry("fails due to not found", testCase{
			inputName:    "testuserpermission",
			expectedCode: http.StatusNotFound,
			mockError:    pkgerrors.New("User Permission does not exist"),
			mockVal:      module.UserPermission{},
			upClient:     &mocks.UserPermissionManager{},
		}),

		Entry("fails due to some other backend error", testCase{
			inputName:    "testuserpermission",
			expectedCode: http.StatusInternalServerError,
			mockError:    pkgerrors.New("backend error"),
			mockVal:      module.UserPermission{},
			upClient:     &mocks.UserPermissionManager{},
		}),
	)

	DescribeTable("Delete UserPermission tests",
		func(t testCase) {
			// set up client mock responses
			t.upClient.On("DeleteUserPerm", "test-project", "test-lc", t.inputName).Return(t.mockError)

			// make HTTP request
			request := httptest.NewRequest("DELETE", "/v2/projects/test-project/logical-clouds/test-lc/user-permissions/"+t.inputName, nil)
			resp := executeRequest(request, NewRouter(t.lcClient, t.clClient, t.upClient, t.quotaClient, t.kvClient))

			// Check returned code
			Expect(resp.StatusCode).To(Equal(t.expectedCode))

			// Check returned body
			got := module.UserPermission{}
			json.NewDecoder(resp.Body).Decode(&got)
			Expect(got).To(Equal(t.mockVal))
		},

		Entry("successful delete", testCase{
			inputName:    "testuserpermission",
			expectedCode: http.StatusNoContent,
			mockError:    nil,
			upClient:     &mocks.UserPermissionManager{},
		}),

		// TODO: implement logic and then enable this test:
		// Entry("fails due to not found", testCase{
		// 	inputName:    "testuserpermission",
		// 	expectedCode: http.StatusNotFound,
		// 	mockError:    pkgerrors.New("db Remove error - not found"),
		// 	upClient:  &mocks.UserPermissionManager{},
		// }),

		// TODO: implement logic and then enable this test:
		// Entry("fails due to a conflict", testCase{
		// 	inputName:    "testuserpermission",
		// 	expectedCode: http.StatusConflict,
		// 	mockError:    pkgerrors.New("db Remove error - conflict"),
		// 	upClient:       &mocks.UserPermissionManager{},
		// }),

		Entry("fails due to other backend error", testCase{
			inputName:    "testuserpermission",
			expectedCode: http.StatusInternalServerError,
			mockError:    pkgerrors.New("db Remove error - general"),
			upClient:     &mocks.UserPermissionManager{},
		}),
	)
})
