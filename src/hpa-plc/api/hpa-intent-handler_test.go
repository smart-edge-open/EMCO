// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2020 Intel Corporation

package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	hpaModel "github.com/open-ness/EMCO/src/hpa-plc/pkg/model"
	orchLog "github.com/open-ness/EMCO/src/orchestrator/pkg/infra/logutils"
	mtypes "github.com/open-ness/EMCO/src/orchestrator/pkg/module/types"
	pkgerrors "github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

func init() {
	hpaIntentJSONFile = "../json-schemas/placement-hpa-intent.json"
	orchLog.SetLoglevel(logrus.InfoLevel)
}

func TestIntentCreateHandler(t *testing.T) {
	testCases := []struct {
		label        string
		reader       io.Reader
		expected     hpaModel.DeploymentHpaIntent
		expectedCode int
		IntentClient *mockIntentManager
	}{
		{
			label:        "Create Intent",
			expectedCode: http.StatusCreated,
			reader: bytes.NewBuffer([]byte(`{
				"metadata" : {
					"name": "testIntent",
    				"description": "Test Intent used for unit testing",
    				"userData1": "data1",
    				"userData2": "data2"
				},
				"spec" : {
					"app-name":"app1"
				}
			}`)),
			expected: hpaModel.DeploymentHpaIntent{
				MetaData: mtypes.Metadata{
					Name:        "testIntent",
					Description: "Test Intent used for unit testing",
					UserData1:   "data1",
					UserData2:   "data2",
				},
				Spec: hpaModel.DeploymentHpaIntentSpec{
					AppName: "app1",
				},
			},
			IntentClient: &mockIntentManager{
				//Items that will be returned by the mocked Client
				Items: []hpaModel.DeploymentHpaIntent{
					{
						MetaData: mtypes.Metadata{
							Name:        "testIntent",
							Description: "Test Intent used for unit testing",
							UserData1:   "data1",
							UserData2:   "data2",
						},
						Spec: hpaModel.DeploymentHpaIntentSpec{
							AppName: "app1",
						},
					},
				},
			},
		},
		{
			label:        "Missing Body Failure",
			expectedCode: http.StatusBadRequest,
			IntentClient: &mockIntentManager{},
		},
		{
			label: "Failed Create Intent due to not found status",
			reader: bytes.NewBuffer([]byte(`{
				"metadata" : {
					"name": "testIntent",
    				"description": "Test Intent used for unit testing",
    				"userData1": "data1",
    				"userData2": "data2"
				},
				"spec" : {
					"app-name":"app1"
				}
			}`)),
			expectedCode: http.StatusNotFound,
			IntentClient: &mockIntentManager{
				Items: []hpaModel.DeploymentHpaIntent{},
				Err:   pkgerrors.New("not found"),
			},
		},
		{
			label: "Failed Create Intent due to conflict status",
			reader: bytes.NewBuffer([]byte(`{
				"metadata" : {
					"name": "testIntent",
    				"description": "Test Intent used for unit testing",
    				"userData1": "data1",
    				"userData2": "data2"
				},
				"spec" : {
					"app-name":"app1"
				}
			}`)),
			expectedCode: http.StatusConflict,
			IntentClient: &mockIntentManager{
				Items: []hpaModel.DeploymentHpaIntent{},
				Err:   pkgerrors.New("conflict"),
			},
		},
		{
			label: "Failed Create Intent due to inernal error status",
			reader: bytes.NewBuffer([]byte(`{
				"metadata" : {
					"name": "testIntent",
    				"description": "Test Intent used for unit testing",
    				"userData1": "data1",
    				"userData2": "data2"
				},
				"spec" : {
					"app-name":"app1"
				}
			}`)),
			expectedCode: http.StatusInternalServerError,
			IntentClient: &mockIntentManager{
				Items: []hpaModel.DeploymentHpaIntent{},
				Err:   pkgerrors.New("inernal"),
			},
		},
		{
			label:        "Failed Create Intent due to bad request body",
			expectedCode: http.StatusUnprocessableEntity,
			reader: bytes.NewBuffer([]byte(`{
				"metadata" : {
					"name": "testIntent",
    				"description": "Test Intent used for unit testing",
    				"userData1": "data1",
    				"userData2": "data2"
				},
				"spec" : {
					"app-name":"app1
				}
			}`)),
		},
		{
			label:        "Missing metadata in Creating Intent",
			expectedCode: http.StatusBadRequest,
			reader: bytes.NewBuffer([]byte(`{
				"spec" : {
					"app-name":"app1"
				}
			}`)),
			expected: hpaModel.DeploymentHpaIntent{
				Spec: hpaModel.DeploymentHpaIntentSpec{
					AppName: "app1",
				},
			},
			IntentClient: &mockIntentManager{
				//Items that will be returned by the mocked Client
				Items: []hpaModel.DeploymentHpaIntent{
					{
						Spec: hpaModel.DeploymentHpaIntentSpec{
							AppName: "app1",
						},
					},
				},
			},
		},
		{
			label: "Missing Intent Name in Request Body",
			reader: bytes.NewBuffer([]byte(`{
                "description":"test description"
                }`)),
			expectedCode: http.StatusBadRequest,
			IntentClient: &mockIntentManager{},
		},
		{
			label: "Empty Intent Name in Request Body",
			reader: bytes.NewBuffer([]byte(`{
				"name": "",
                "description":"test description"
                }`)),
			expectedCode: http.StatusBadRequest,
			IntentClient: &mockIntentManager{},
		},
		{
			label: "Missing  App Name in Request Body",
			reader: bytes.NewBuffer([]byte(`{
				"metadata" : {
					"name": "testIntent",
    				"description": "Test Intent used for unit testing",
    				"userData1": "data1",
    				"userData2": "data2"
				},
				"spec" : {
				}
			}`)),
			expectedCode: http.StatusBadRequest,
			IntentClient: &mockIntentManager{},
		},
		{
			label: "Empty App Name in Request Body",
			reader: bytes.NewBuffer([]byte(`{
				"metadata" : {
					"name": "testIntent",
    				"description": "Test Intent used for unit testing",
    				"userData1": "data1",
    				"userData2": "data2"
				},
				"spec" : {
					"app-name":""
				}
			}`)),
			expectedCode: http.StatusBadRequest,
			IntentClient: &mockIntentManager{},
		},
	}

	fmt.Printf("\n================== TestIntentCreateHandler .. total_testcase_count[%d] ==================\n", len(testCases))
	for i, testCase := range testCases {
		fmt.Printf("\n================== TestIntentCreateHandler .. testcase_count[%d] testcase_name[%s] ==================\n", i, testCase.label)
		t.Run(testCase.label, func(t *testing.T) {
			request := httptest.NewRequest("POST", "/v2/projects/project1/composite-apps/compositeapp1/v2/deployment-intent-groups/digroup/hpa-intents", testCase.reader)
			resp := executeRequest(request, NewRouter(testCase.IntentClient))

			//Check returned code
			if resp.StatusCode != testCase.expectedCode {
				t.Fatalf("Expected %d; Got: %d", testCase.expectedCode, resp.StatusCode)
			}

			//Check returned body only if statusCreated
			if resp.StatusCode == http.StatusCreated {
				got := hpaModel.DeploymentHpaIntent{}
				json.NewDecoder(resp.Body).Decode(&got)

				if reflect.DeepEqual(testCase.expected, got) == false {
					t.Errorf("createHandler returned unexpected body: got %v;"+
						" expected %v", got, testCase.expected)
				}
			}
		})
	}
}

func TestIntentUpdateHandler(t *testing.T) {
	testCases := []struct {
		label, name  string
		reader       io.Reader
		expected     hpaModel.DeploymentHpaIntent
		expectedCode int
		IntentClient *mockIntentManager
	}{
		{
			label: "Missing Intent Name in Request Body",
			name:  "testIntent",
			reader: bytes.NewBuffer([]byte(`{
				"description":"test description"
				}`)),
			expectedCode: http.StatusBadRequest,
			IntentClient: &mockIntentManager{},
		},
		{
			label:        "Missing Body Failure",
			name:         "testIntent",
			expectedCode: http.StatusBadRequest,
			IntentClient: &mockIntentManager{},
		},
		{
			label:        "Failed Update Intent due to bad request body",
			name:         "testIntent",
			expectedCode: http.StatusUnprocessableEntity,
			reader: bytes.NewBuffer([]byte(`{
				"metadata" : {
					"name": "testIntent",
    				"description": "Test Intent used for unit testing",
    				"userData1": "data1",
    				"userData2": "data2
				},
				"spec" : {
					"app-name":"app1
				}
			}`)),
		},
		{
			label:        "Missing metadata in updating Intent",
			name:         "testIntent",
			expectedCode: http.StatusBadRequest,
			reader: bytes.NewBuffer([]byte(`{
				"spec" : {
					"app-name":"app1"
				}
			}`)),
		},
		{
			label:        "Mismatched Name Failure",
			name:         "testIntent",
			expectedCode: http.StatusBadRequest,
			reader: bytes.NewBuffer([]byte(`{
				"metadata" : {
					"name": "testIntentNameMismatch",
					"description": "Test Intent used for unit testing"
				}
			}`)),
			IntentClient: &mockIntentManager{},
		},
		{
			label:        "Update Intent",
			name:         "testIntent",
			expectedCode: http.StatusOK,
			reader: bytes.NewBuffer([]byte(`{
				"metadata" : {
					"name": "testIntent",
					"description": "Test Intent used for unit testing 2",
					"userData1": "update data1",
    				"userData2": "update data2"
				},
				"spec" : {
					"app-name":"app1"
				}
			}`)),
			expected: hpaModel.DeploymentHpaIntent{
				MetaData: mtypes.Metadata{
					Name:        "testIntent",
					Description: "Test Intent used for unit testing 2",
					UserData1:   "update data1",
					UserData2:   "update data2",
				},
				Spec: hpaModel.DeploymentHpaIntentSpec{
					AppName: "app1",
				},
			},
			IntentClient: &mockIntentManager{
				//Items that will be returned by the mocked Client
				Items: []hpaModel.DeploymentHpaIntent{
					{
						MetaData: mtypes.Metadata{
							Name:        "testIntent",
							Description: "Test Intent used for unit testing 2",
							UserData1:   "update data1",
							UserData2:   "update data2",
						},
						Spec: hpaModel.DeploymentHpaIntentSpec{
							AppName: "app1",
						},
					},
				},
			},
		},
		{
			label:        "Failed Update Intent due to intent-name mismatch",
			name:         "testIntent",
			expectedCode: http.StatusBadRequest,
			reader: bytes.NewBuffer([]byte(`{
				"metadata" : {
					"name": "testIntent1",
					"description": "Test Intent used for unit testing 2",
					"userData1": "update data1",
    				"userData2": "update data2"
				},
				"spec" : {
					"app-name":"app1"
				}
			}`)),
			expected: hpaModel.DeploymentHpaIntent{
				MetaData: mtypes.Metadata{
					Name:        "testIntent",
					Description: "Test Intent used for unit testing 2",
					UserData1:   "update data1",
					UserData2:   "update data2",
				},
				Spec: hpaModel.DeploymentHpaIntentSpec{
					AppName: "app1",
				},
			},
			IntentClient: &mockIntentManager{
				//Items that will be returned by the mocked Client
				Items: []hpaModel.DeploymentHpaIntent{
					{
						MetaData: mtypes.Metadata{
							Name:        "testIntent",
							Description: "Test Intent used for unit testing 2",
							UserData1:   "update data1",
							UserData2:   "update data2",
						},
						Spec: hpaModel.DeploymentHpaIntentSpec{
							AppName: "app1",
						},
					},
				},
			},
		},
		{
			label:        "Empty req intentname in Update Intent",
			name:         "testIntent",
			expectedCode: http.StatusBadRequest,
			reader: bytes.NewBuffer([]byte(`{
				"metadata" : {
					"name": "",
					"description": "Test Intent used for unit testing 2",
					"userData1": "update data1",
    				"userData2": "update data2"
				},
				"spec" : {
					"app-name":"app1"
				}
			}`)),
		},
		{
			label:        "Empty intentname in Update Intent",
			name:         "",
			expectedCode: http.StatusNotFound,
			reader: bytes.NewBuffer([]byte(`{
				"metadata" : {
					"name": "testIntent",
					"description": "Test Intent used for unit testing 2",
					"userData1": "update data1",
    				"userData2": "update data2"
				},
				"spec" : {
					"app-name":"app1"
				}
			}`)),
		},
		{
			label: "Empty App Name in Request Body",
			name:  "testIntent",
			reader: bytes.NewBuffer([]byte(`{
				"metadata" : {
					"name": "testIntent",
    				"description": "Test Intent used for unit testing",
    				"userData1": "data1",
    				"userData2": "data2"
				},
				"spec" : {
					"app-name":""
				}
			}`)),
			expectedCode: http.StatusBadRequest,
			IntentClient: &mockIntentManager{},
		},
		{
			label:        "Failed Update Intent due to internal server error",
			name:         "testIntent",
			expectedCode: http.StatusInternalServerError,
			reader: bytes.NewBuffer([]byte(`{
				"metadata" : {
					"name": "testIntent",
					"description": "Test Intent used for unit testing 2",
					"userData1": "update data1",
    				"userData2": "update data2"
				},
				"spec" : {
					"app-name":"app1"
				}
			}`)),
			IntentClient: &mockIntentManager{
				Items: []hpaModel.DeploymentHpaIntent{},
				Err:   pkgerrors.New("inernal"),
			},
		},
		{
			label:        "Failed Update Intent due to not found error",
			name:         "testIntent",
			expectedCode: http.StatusNotFound,
			reader: bytes.NewBuffer([]byte(`{
				"metadata" : {
					"name": "testIntent",
					"description": "Test Intent used for unit testing 2",
					"userData1": "update data1",
    				"userData2": "update data2"
				},
				"spec" : {
					"app-name":"app1"
				}
			}`)),
			IntentClient: &mockIntentManager{
				Items: []hpaModel.DeploymentHpaIntent{},
				Err:   pkgerrors.New("not found"),
			},
		},
	}

	fmt.Printf("\n================== TestIntentUpdateHandler .. total_testcase_count[%d] ==================\n", len(testCases))
	for i, testCase := range testCases {
		fmt.Printf("\n================== TestIntentUpdateHandler .. testcase_count[%d] testcase_name[%s] ==================\n", i, testCase.label)
		t.Run(testCase.label, func(t *testing.T) {
			request := httptest.NewRequest("PUT", "/v2/projects/project1/composite-apps/compositeapp1/v2/deployment-intent-groups/digroup/hpa-intents/"+testCase.name, testCase.reader)
			resp := executeRequest(request, NewRouter(testCase.IntentClient))

			//Check returned code
			if resp.StatusCode != testCase.expectedCode {
				t.Fatalf("Expected %d; Got: %d", testCase.expectedCode, resp.StatusCode)
			}

			//Check returned body only if statusOK
			if resp.StatusCode == http.StatusOK {
				got := hpaModel.DeploymentHpaIntent{}
				json.NewDecoder(resp.Body).Decode(&got)

				if reflect.DeepEqual(testCase.expected, got) == false {
					t.Errorf("updateHandler returned unexpected body: got %v;"+
						" expected %v", got, testCase.expected)
				}
			}
		})
	}
}

func TestIntentGetHandler(t *testing.T) {

	testCases := []struct {
		label         string
		expected      hpaModel.DeploymentHpaIntent
		name, version string
		expectedCode  int
		IntentClient  *mockIntentManager
	}{
		{
			label:        "Get Intent metadata",
			expectedCode: http.StatusOK,
			expected: hpaModel.DeploymentHpaIntent{
				MetaData: mtypes.Metadata{
					Name:        "testIntent",
					Description: "Test Intent used for unit testing",
					UserData1:   "data1",
					UserData2:   "data2",
				},
			},
			name: "testIntent",
			IntentClient: &mockIntentManager{
				Items: []hpaModel.DeploymentHpaIntent{
					{
						MetaData: mtypes.Metadata{
							Name:        "testIntent",
							Description: "Test Intent used for unit testing",
							UserData1:   "data1",
							UserData2:   "data2",
						},
					},
				},
			},
		},
		{
			label:        "Get Intent request",
			expectedCode: http.StatusOK,
			expected: hpaModel.DeploymentHpaIntent{
				MetaData: mtypes.Metadata{
					Name:        "testIntent",
					Description: "Test Intent used for unit testing",
					UserData1:   "data1",
					UserData2:   "data2",
				},
				Spec: hpaModel.DeploymentHpaIntentSpec{AppName: "app1"},
			},
			name: "testIntent",
			IntentClient: &mockIntentManager{
				Items: []hpaModel.DeploymentHpaIntent{
					{
						MetaData: mtypes.Metadata{
							Name:        "testIntent",
							Description: "Test Intent used for unit testing",
							UserData1:   "data1",
							UserData2:   "data2",
						},
						Spec: hpaModel.DeploymentHpaIntentSpec{AppName: "app1"},
					},
				},
			},
		},
		{
			label:        "Get Non-Exiting empty Intent",
			expectedCode: http.StatusNotFound,
			name:         "testIntentBad",
			IntentClient: &mockIntentManager{
				Items: []hpaModel.DeploymentHpaIntent{},
				Err:   pkgerrors.New("Internal Error"),
			},
		},
		{
			label:        "Get empty Intent request",
			expectedCode: http.StatusNotFound,
			expected: hpaModel.DeploymentHpaIntent{
				MetaData: mtypes.Metadata{
					Name:        "testIntent",
					Description: "Test Intent used for unit testing",
					UserData1:   "data1",
					UserData2:   "data2",
				},
				Spec: hpaModel.DeploymentHpaIntentSpec{AppName: "app1"},
			},
			name: "",
			IntentClient: &mockIntentManager{
				Items: []hpaModel.DeploymentHpaIntent{
					{
						MetaData: mtypes.Metadata{
							Name:        "testIntent",
							Description: "Test Intent used for unit testing",
							UserData1:   "data1",
							UserData2:   "data2",
						},
						Spec: hpaModel.DeploymentHpaIntentSpec{AppName: "app1"},
					},
				},
			},
		},
	}

	fmt.Printf("\n================== TestIntentGetHandler .. total_testcase_count[%d] ==================\n", len(testCases))
	for i, testCase := range testCases {
		fmt.Printf("\n================== TestIntentGetHandler .. testcase_count[%d] testcase_name[%s] ==================\n", i, testCase.label)
		t.Run(testCase.label, func(t *testing.T) {
			request := httptest.NewRequest("GET", "/v2/projects/project1/composite-apps/compositeapp1/v2/deployment-intent-groups/digroup/hpa-intents/"+testCase.name, nil)
			resp := executeRequest(request, NewRouter(testCase.IntentClient))

			//Check returned code
			if resp.StatusCode != testCase.expectedCode {
				t.Fatalf("Expected %d; Got: %d", testCase.expectedCode, resp.StatusCode)
			}

			//Check returned body only if statusOK
			if resp.StatusCode == http.StatusOK {
				got := hpaModel.DeploymentHpaIntent{}
				json.NewDecoder(resp.Body).Decode(&got)

				if reflect.DeepEqual(testCase.expected, got) == false {
					t.Errorf("TestIntentGetHandler returned unexpected body: got %v;"+
						" expected %v", got, testCase.expected)
				}
			}
		})
	}
}

func TestIntentGetHandlerByName(t *testing.T) {

	testCases := []struct {
		label         string
		expected      hpaModel.DeploymentHpaIntent
		name, version string
		expectedCode  int
		IntentClient  *mockIntentManager
	}{
		{
			label:        "GetIntentByName metadata",
			expectedCode: http.StatusOK,
			expected: hpaModel.DeploymentHpaIntent{
				MetaData: mtypes.Metadata{
					Name:        "testIntent",
					Description: "Test Intent used for unit testing",
					UserData1:   "data1",
					UserData2:   "data2",
				},
			},
			name: "testIntent",
			IntentClient: &mockIntentManager{
				Items: []hpaModel.DeploymentHpaIntent{
					{
						MetaData: mtypes.Metadata{
							Name:        "testIntent",
							Description: "Test Intent used for unit testing",
							UserData1:   "data1",
							UserData2:   "data2",
						},
					},
				},
			},
		},
		{
			label:        "GetIntentByName request",
			expectedCode: http.StatusOK,
			expected: hpaModel.DeploymentHpaIntent{
				MetaData: mtypes.Metadata{
					Name:        "testIntent",
					Description: "Test Intent used for unit testing",
					UserData1:   "data1",
					UserData2:   "data2",
				},
				Spec: hpaModel.DeploymentHpaIntentSpec{AppName: "app1"},
			},
			name: "testIntent",
			IntentClient: &mockIntentManager{
				Items: []hpaModel.DeploymentHpaIntent{
					{
						MetaData: mtypes.Metadata{
							Name:        "testIntent",
							Description: "Test Intent used for unit testing",
							UserData1:   "data1",
							UserData2:   "data2",
						},
						Spec: hpaModel.DeploymentHpaIntentSpec{AppName: "app1"},
					},
				},
			},
		},
		{
			label:        "GetIntentByName Non-Exiting empty Intent",
			expectedCode: http.StatusNotFound,
			name:         "testIntentBad",
			IntentClient: &mockIntentManager{
				Items: []hpaModel.DeploymentHpaIntent{},
				Err:   pkgerrors.New("Internal Error"),
			},
		},
	}

	fmt.Printf("\n================== TestIntentGetHandlerByName .. total_testcase_count[%d] ==================\n", len(testCases))
	for i, testCase := range testCases {
		fmt.Printf("\n================== TestIntentGetHandlerByName .. testcase_count[%d] testcase_name[%s] ==================\n", i, testCase.label)
		t.Run(testCase.label, func(t *testing.T) {
			request := httptest.NewRequest("QUERY", "/v2/projects/project1/composite-apps/compositeapp1/v2/deployment-intent-groups/digroup/hpa-intents?intent="+testCase.name, nil)
			resp := executeRequest(request, NewRouter(testCase.IntentClient))

			//Check returned code
			if resp.StatusCode != testCase.expectedCode {
				t.Fatalf("Expected %d; Got: %d", testCase.expectedCode, resp.StatusCode)
			}

			//Check returned body only if statusOK
			if resp.StatusCode == http.StatusOK {
				got := hpaModel.DeploymentHpaIntent{}
				json.NewDecoder(resp.Body).Decode(&got)

				if reflect.DeepEqual(testCase.expected, got) == false {
					t.Errorf("TestIntentGetHandler returned unexpected body: got %v;"+
						" expected %v", got, testCase.expected)
				}
			}
		})
	}
}

func TestIntentGetAllHandler(t *testing.T) {

	testCases := []struct {
		label         string
		expected      []hpaModel.DeploymentHpaIntent
		name, version string
		expectedCode  int
		IntentClient  *mockIntentManager
	}{
		{
			label:        "GetAll Intent metadata",
			expectedCode: http.StatusOK,
			expected: []hpaModel.DeploymentHpaIntent{
				{
					MetaData: mtypes.Metadata{
						Name:        "testIntent",
						Description: "Test Intent used for unit testing",
						UserData1:   "data1",
						UserData2:   "data2",
					},
				},
			},
			IntentClient: &mockIntentManager{
				Items: []hpaModel.DeploymentHpaIntent{
					{
						MetaData: mtypes.Metadata{
							Name:        "testIntent",
							Description: "Test Intent used for unit testing",
							UserData1:   "data1",
							UserData2:   "data2",
						},
					},
				},
			},
		},
		{
			label:        "GetAll Intent request",
			expectedCode: http.StatusOK,
			expected: []hpaModel.DeploymentHpaIntent{
				{
					MetaData: mtypes.Metadata{
						Name:        "testIntent",
						Description: "Test Intent used for unit testing",
						UserData1:   "data1",
						UserData2:   "data2",
					},
					Spec: hpaModel.DeploymentHpaIntentSpec{AppName: "app1"},
				},
				{
					MetaData: mtypes.Metadata{
						Name:        "testIntent2",
						Description: "Test Intent2 used for unit testing",
						UserData1:   "data1",
						UserData2:   "data2",
					},
					Spec: hpaModel.DeploymentHpaIntentSpec{AppName: "app2"},
				},
			},
			IntentClient: &mockIntentManager{
				Items: []hpaModel.DeploymentHpaIntent{
					{
						MetaData: mtypes.Metadata{
							Name:        "testIntent",
							Description: "Test Intent used for unit testing",
							UserData1:   "data1",
							UserData2:   "data2",
						},
						Spec: hpaModel.DeploymentHpaIntentSpec{AppName: "app1"},
					},
					{
						MetaData: mtypes.Metadata{
							Name:        "testIntent2",
							Description: "Test Intent2 used for unit testing",
							UserData1:   "data1",
							UserData2:   "data2",
						},
						Spec: hpaModel.DeploymentHpaIntentSpec{AppName: "app2"},
					},
				},
			},
		},
		{
			label:        "GetAll Non-Exiting Intent",
			expectedCode: http.StatusNotFound,
			name:         "nonexistingIntent",
			IntentClient: &mockIntentManager{
				Items: []hpaModel.DeploymentHpaIntent{},
				Err:   pkgerrors.New("Internal Error"),
			},
		},
		{
			label:        "GetAll No intents",
			expectedCode: http.StatusOK,
			IntentClient: &mockIntentManager{
				Items: []hpaModel.DeploymentHpaIntent{},
			},
			expected: []hpaModel.DeploymentHpaIntent{},
		},
	}

	fmt.Printf("\n================== TestIntentGetAllHandler .. total_testcase_count[%d] ==================\n", len(testCases))
	for i, testCase := range testCases {
		fmt.Printf("\n================== TestIntentGetAllHandler .. testcase_count[%d] testcase_name[%s] ==================\n", i, testCase.label)
		t.Run(testCase.label, func(t *testing.T) {
			request := httptest.NewRequest("GET", "/v2/projects/project1/composite-apps/compositeapp1/v2/deployment-intent-groups/digroup/hpa-intents", nil)
			resp := executeRequest(request, NewRouter(testCase.IntentClient))

			//Check returned code
			if resp.StatusCode != testCase.expectedCode {
				t.Fatalf("Expected [%d]; Got: [%d]", testCase.expectedCode, resp.StatusCode)
			}

			//Check returned body only if statusOK
			if resp.StatusCode == http.StatusOK {
				got := []hpaModel.DeploymentHpaIntent{}
				json.NewDecoder(resp.Body).Decode(&got)

				if reflect.DeepEqual(testCase.expected, got) == false {
					t.Errorf("TestIntentGetAllHandler returned unexpected body: got [%v];"+
						" expected [%v]", got, testCase.expected)
				}
			}
		})
	}
}

func TestIntentDeleteHandler(t *testing.T) {

	testCases := []struct {
		label        string
		name         string
		version      string
		expectedCode int
		IntentClient *mockIntentManager
	}{
		{
			label:        "Delete Intent",
			expectedCode: http.StatusNoContent,
			name:         "testIntent",
			IntentClient: &mockIntentManager{
				//Items that will be returned by the mocked Client
				Items: []hpaModel.DeploymentHpaIntent{
					{
						MetaData: mtypes.Metadata{
							Name:        "testIntent",
							Description: "Test Intent used for unit testing",
							UserData1:   "data1",
							UserData2:   "data2",
						},
					},
				},
			},
		},
		{
			label:        "Delete exiting Intent & internal error",
			expectedCode: http.StatusNotFound,
			name:         "testIntent",
			IntentClient: &mockIntentManager{
				//Items that will be returned by the mocked Client
				Items: []hpaModel.DeploymentHpaIntent{
					{
						MetaData: mtypes.Metadata{
							Name:        "testIntent",
							Description: "Test Intent used for unit testing",
							UserData1:   "data1",
							UserData2:   "data2",
						},
					},
				},
				Err: pkgerrors.New("Internal Error"),
			},
		},
		{
			label:        "Delete Non-Exiting Intent",
			expectedCode: http.StatusNotFound,
			name:         "testIntent",
			IntentClient: &mockIntentManager{
				Err: pkgerrors.New("Internal Error"),
			},
		},
		{
			label:        "Delete Non-Exiting emoty Intent",
			expectedCode: http.StatusNotFound,
			name:         "",
			IntentClient: &mockIntentManager{
				Err: pkgerrors.New("Internal Error"),
			},
		},
	}

	fmt.Printf("\n================== TestIntentDeleteHandler .. total_testcase_count[%d] ==================\n", len(testCases))
	for i, testCase := range testCases {
		fmt.Printf("\n================== TestIntentDeleteHandler .. testcase_count[%d] testcase_name[%s] ==================\n", i, testCase.label)
		t.Run(testCase.label, func(t *testing.T) {
			request := httptest.NewRequest("DELETE", "/v2/projects/project1/composite-apps/compositeapp1/v2/deployment-intent-groups/digroup/hpa-intents/"+testCase.name, nil)
			resp := executeRequest(request, NewRouter(testCase.IntentClient))

			//Check returned code
			if resp.StatusCode != testCase.expectedCode {
				t.Fatalf("Expected %d; Got: %d", testCase.expectedCode, resp.StatusCode)
			}
		})
	}
}

func TestIntentDeleteAllHandler(t *testing.T) {

	testCases := []struct {
		label        string
		name         string
		version      string
		expectedCode int
		IntentClient *mockIntentManager
	}{
		{
			label:        "Delete All Intent",
			expectedCode: http.StatusNoContent,
			IntentClient: &mockIntentManager{
				//Items that will be returned by the mocked Client
				Items: []hpaModel.DeploymentHpaIntent{
					{
						MetaData: mtypes.Metadata{
							Name:        "testIntent",
							Description: "Test Intent used for unit testing",
							UserData1:   "data1",
							UserData2:   "data2",
						},
					},
				},
			},
		},
		{
			label:        "Delete All Non-Exiting Intents",
			expectedCode: http.StatusNotFound,
			IntentClient: &mockIntentManager{
				Err: pkgerrors.New("Internal Error"),
			},
		},
	}

	fmt.Printf("\n================== TestIntentDeleteAllHandler .. total_testcase_count[%d] ==================\n", len(testCases))
	for i, testCase := range testCases {
		fmt.Printf("\n================== TestIntentDeleteAllHandler .. testcase_count[%d] testcase_name[%s] ==================\n", i, testCase.label)
		t.Run(testCase.label, func(t *testing.T) {
			request := httptest.NewRequest("DELETE", "/v2/projects/project1/composite-apps/compositeapp1/v2/deployment-intent-groups/digroup/hpa-intents", nil)
			resp := executeRequest(request, NewRouter(testCase.IntentClient))

			//Check returned code
			if resp.StatusCode != testCase.expectedCode {
				t.Fatalf("Expected %d; Got: %d", testCase.expectedCode, resp.StatusCode)
			}
		})
	}
}
