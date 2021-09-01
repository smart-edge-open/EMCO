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
	hpaConsumerJSONFile = "../json-schemas/placement-hpa-consumer.json"
	orchLog.SetLoglevel(logrus.InfoLevel)
}

func TestConsumerCreateHandler(t *testing.T) {
	testCases := []struct {
		label          string
		reader         io.Reader
		expected       hpaModel.HpaResourceConsumer
		expectedCode   int
		ConsumerClient *mockIntentManager
	}{
		{
			label:        "Create Consumer",
			expectedCode: http.StatusCreated,
			reader: bytes.NewBuffer([]byte(`{
				"metadata" : {
					"name": "testConsumer",
    				"description": "Test Consumer used for unit testing",
    				"userData1": "data1",
    				"userData2": "data2"
				},
				"spec" : {
					"replicas": 1,
					"name":          "deployment-1",
					"container-name": "container-1"
				}
			}`)),
			expected: hpaModel.HpaResourceConsumer{
				MetaData: mtypes.Metadata{
					Name:        "testConsumer",
					Description: "Test Consumer used for unit testing",
					UserData1:   "data1",
					UserData2:   "data2",
				},
				Spec: hpaModel.HpaResourceConsumerSpec{
					Replicas:      1,
					Name:          "deployment-1",
					ContainerName: "container-1",
				},
			},
			ConsumerClient: &mockIntentManager{
				//Items that will be returned by the mocked Client
				ConsumerItems: []hpaModel.HpaResourceConsumer{
					{
						MetaData: mtypes.Metadata{
							Name:        "testConsumer",
							Description: "Test Consumer used for unit testing",
							UserData1:   "data1",
							UserData2:   "data2",
						},
						Spec: hpaModel.HpaResourceConsumerSpec{
							Replicas:      1,
							Name:          "deployment-1",
							ContainerName: "container-1",
						},
					},
				},
			},
		},
		{
			label:        "Create Consumer with replicas",
			expectedCode: http.StatusCreated,
			reader: bytes.NewBuffer([]byte(`{
				"metadata" : {
					"name": "testConsumer",
    				"description": "Test Consumer used for unit testing",
    				"userData1": "data1",
    				"userData2": "data2"
				},
				"spec" : {
					"replicas": 100,
					"name":          "deployment-1",
					"container-name": "container-1"
				}
			}`)),
			expected: hpaModel.HpaResourceConsumer{
				MetaData: mtypes.Metadata{
					Name:        "testConsumer",
					Description: "Test Consumer used for unit testing",
					UserData1:   "data1",
					UserData2:   "data2",
				},
				Spec: hpaModel.HpaResourceConsumerSpec{
					Replicas:      100,
					Name:          "deployment-1",
					ContainerName: "container-1",
				},
			},
			ConsumerClient: &mockIntentManager{
				//Items that will be returned by the mocked Client
				ConsumerItems: []hpaModel.HpaResourceConsumer{
					{
						MetaData: mtypes.Metadata{
							Name:        "testConsumer",
							Description: "Test Consumer used for unit testing",
							UserData1:   "data1",
							UserData2:   "data2",
						},
						Spec: hpaModel.HpaResourceConsumerSpec{
							Replicas:      100,
							Name:          "deployment-1",
							ContainerName: "container-1",
						},
					},
				},
			},
		},
		{
			label:        "Failed Create Consumer with NO replicas specified",
			expectedCode: http.StatusCreated,
			reader: bytes.NewBuffer([]byte(`{
				"metadata" : {
					"name": "testConsumer",
    				"description": "Test Consumer used for unit testing",
    				"userData1": "data1",
    				"userData2": "data2"
				},
				"spec" : {
					"name":          "deployment-1",
					"container-name": "container-1"
				}
			}`)),
			expected: hpaModel.HpaResourceConsumer{
				MetaData: mtypes.Metadata{
					Name:        "testConsumer",
					Description: "Test Consumer used for unit testing",
					UserData1:   "data1",
					UserData2:   "data2",
				},
				Spec: hpaModel.HpaResourceConsumerSpec{
					Name:          "deployment-1",
					ContainerName: "container-1",
				},
			},
			ConsumerClient: &mockIntentManager{
				//Items that will be returned by the mocked Client
				ConsumerItems: []hpaModel.HpaResourceConsumer{
					{
						MetaData: mtypes.Metadata{
							Name:        "testConsumer",
							Description: "Test Consumer used for unit testing",
							UserData1:   "data1",
							UserData2:   "data2",
						},
						Spec: hpaModel.HpaResourceConsumerSpec{
							Name:          "deployment-1",
							ContainerName: "container-1",
						},
					},
				},
			},
		},
		{
			label:          "Missing Body Failure",
			expectedCode:   http.StatusBadRequest,
			ConsumerClient: &mockIntentManager{},
		},
		{
			label:        "Failed Create Consumer due to bad request body",
			expectedCode: http.StatusUnprocessableEntity,
			reader: bytes.NewBuffer([]byte(`{
				"metadata" : {
					"name": "testConsumer",
    				"description": "Test Consumer used for unit testing",
    				"userData1": "data1",
    				"userData2": "data2"
				},
				"spec" : {
					"name":          "deployment-1
					"container-name": "container-1"
				}
			}`)),
		},
		{
			label:        "Failed Create Consumer due to not found status",
			expectedCode: http.StatusNotFound,
			reader: bytes.NewBuffer([]byte(`{
				"metadata" : {
					"name": "testConsumer",
    				"description": "Test Consumer used for unit testing",
    				"userData1": "data1",
    				"userData2": "data2"
				},
				"spec" : {
					"replicas": 1,
					"name":          "deployment-1",
					"container-name": "container-1"
				}
			}`)),
			ConsumerClient: &mockIntentManager{
				ConsumerItems: []hpaModel.HpaResourceConsumer{},
				Err:           pkgerrors.New("not found"),
			},
		},
		{
			label:        "Failed Create Consumer due to conflict status",
			expectedCode: http.StatusConflict,
			reader: bytes.NewBuffer([]byte(`{
				"metadata" : {
					"name": "testConsumer",
    				"description": "Test Consumer used for unit testing",
    				"userData1": "data1",
    				"userData2": "data2"
				},
				"spec" : {
					"replicas": 1,
					"name":          "deployment-1",
					"container-name": "container-1"
				}
			}`)),
			ConsumerClient: &mockIntentManager{
				ConsumerItems: []hpaModel.HpaResourceConsumer{},
				Err:           pkgerrors.New("conflict"),
			},
		},
		{
			label:        "Failed Create Consumer due to internal error status",
			expectedCode: http.StatusInternalServerError,
			reader: bytes.NewBuffer([]byte(`{
				"metadata" : {
					"name": "testConsumer",
    				"description": "Test Consumer used for unit testing",
    				"userData1": "data1",
    				"userData2": "data2"
				},
				"spec" : {
					"replicas": 1,
					"name":          "deployment-1",
					"container-name": "container-1"
				}
			}`)),
			ConsumerClient: &mockIntentManager{
				ConsumerItems: []hpaModel.HpaResourceConsumer{},
				Err:           pkgerrors.New("internal"),
			},
		},
		{
			label: "Missing Consumer Name in Request Body",
			reader: bytes.NewBuffer([]byte(`{
                "description":"test description"
                }`)),
			expectedCode:   http.StatusBadRequest,
			ConsumerClient: &mockIntentManager{},
		},
		{
			label: "Empty Consumer Name in Request Body",
			reader: bytes.NewBuffer([]byte(`{
				"name": "",
                "description":"test description"
                }`)),
			expectedCode:   http.StatusBadRequest,
			ConsumerClient: &mockIntentManager{},
		},
		{
			label:        "Missing Deployment name in Create Consumer",
			expectedCode: http.StatusBadRequest,
			reader: bytes.NewBuffer([]byte(`{
				"metadata" : {
					"name": "testConsumer",
    				"description": "Test Consumer used for unit testing",
    				"userData1": "data1",
    				"userData2": "data2"
				},
				"spec" : {
					"replicas": 1,
					"container-name": "container-1"
				}
			}`)),
			expected: hpaModel.HpaResourceConsumer{
				MetaData: mtypes.Metadata{
					Name:        "testConsumer",
					Description: "Test Consumer used for unit testing",
					UserData1:   "data1",
					UserData2:   "data2",
				},
				Spec: hpaModel.HpaResourceConsumerSpec{
					Replicas:      1,
					ContainerName: "container-1",
				},
			},
			ConsumerClient: &mockIntentManager{
				//Items that will be returned by the mocked Client
				ConsumerItems: []hpaModel.HpaResourceConsumer{
					{
						MetaData: mtypes.Metadata{
							Name:        "testConsumer",
							Description: "Test Consumer used for unit testing",
							UserData1:   "data1",
							UserData2:   "data2",
						},
						Spec: hpaModel.HpaResourceConsumerSpec{
							Replicas:      1,
							ContainerName: "container-1",
						},
					},
				},
			},
		},
		{
			label:        "Empty Deployment name in Create Consumer",
			expectedCode: http.StatusBadRequest,
			reader: bytes.NewBuffer([]byte(`{
				"metadata" : {
					"name": "testConsumer",
    				"description": "Test Consumer used for unit testing",
    				"userData1": "data1",
    				"userData2": "data2"
				},
				"spec" : {
					"replicas": 1,
					"name":          "",
					"container-name": "container-1"
				}
			}`)),
			expected: hpaModel.HpaResourceConsumer{
				MetaData: mtypes.Metadata{
					Name:        "testConsumer",
					Description: "Test Consumer used for unit testing",
					UserData1:   "data1",
					UserData2:   "data2",
				},
				Spec: hpaModel.HpaResourceConsumerSpec{
					Replicas:      1,
					Name:          "",
					ContainerName: "container-1",
				},
			},
			ConsumerClient: &mockIntentManager{
				//Items that will be returned by the mocked Client
				ConsumerItems: []hpaModel.HpaResourceConsumer{
					{
						MetaData: mtypes.Metadata{
							Name:        "testConsumer",
							Description: "Test Consumer used for unit testing",
							UserData1:   "data1",
							UserData2:   "data2",
						},
						Spec: hpaModel.HpaResourceConsumerSpec{
							Replicas:      1,
							Name:          "",
							ContainerName: "container-1",
						},
					},
				},
			},
		},
	}

	fmt.Printf("\n================== TestConsumerCreateHandler .. total_testcase_count[%d] ==================\n", len(testCases))
	for i, testCase := range testCases {
		fmt.Printf("\n================== TestConsumerCreateHandler .. testcase_count[%d] testcase_name[%s] ==================\n", i, testCase.label)
		t.Run(testCase.label, func(t *testing.T) {
			request := httptest.NewRequest("POST", "/v2/projects/project1/composite-apps/compositeapp1/v2/deployment-intent-groups/digroup/hpa-intents/hpaintent1/hpa-resource-consumers", testCase.reader)

			resp := executeRequest(request, NewRouter(testCase.ConsumerClient))

			//Check returned code
			if resp.StatusCode != testCase.expectedCode {
				t.Fatalf("Expected %d; Got: %d", testCase.expectedCode, resp.StatusCode)
			}

			//Check returned body only if statusCreated
			if resp.StatusCode == http.StatusCreated {
				got := hpaModel.HpaResourceConsumer{}
				json.NewDecoder(resp.Body).Decode(&got)

				if reflect.DeepEqual(testCase.expected, got) == false {
					t.Errorf("createHandler returned unexpected body: got %v;"+
						" expected %v", got, testCase.expected)
				}
			}
		})
	}
}

func TestConsumerUpdateHandler(t *testing.T) {
	testCases := []struct {
		label, name    string
		reader         io.Reader
		expected       hpaModel.HpaResourceConsumer
		expectedCode   int
		ConsumerClient *mockIntentManager
	}{
		{
			label: "Missing Consumer Name in Request Body",
			name:  "testConsumer",
			reader: bytes.NewBuffer([]byte(`{
				"description":"test description"
				}`)),
			expectedCode:   http.StatusBadRequest,
			ConsumerClient: &mockIntentManager{},
		},
		{
			label:          "Missing Body Failure",
			name:           "testConsumer",
			expectedCode:   http.StatusBadRequest,
			ConsumerClient: &mockIntentManager{},
		},
		{
			label:        "Mismatched Name Failure",
			name:         "testConsumer",
			expectedCode: http.StatusBadRequest,
			reader: bytes.NewBuffer([]byte(`{
				"metadata" : {
					"name": "testConsumerNameMismatch",
					"description": "Test Consumer used for unit testing"
				}
			}`)),
			ConsumerClient: &mockIntentManager{},
		},
		{
			label:        "Update Consumer",
			name:         "testConsumer",
			expectedCode: http.StatusOK,
			reader: bytes.NewBuffer([]byte(`{
				"metadata" : {
					"name": "testConsumer",
    				"description": "Test Consumer used for unit testing",
    				"userData1": "update data1",
    				"userData2": "update data2"
				},
				"spec" : {
					"replicas": 1,
					"name":          "deployment-1",
					"container-name": "container-1"
				}
			}`)),
			expected: hpaModel.HpaResourceConsumer{
				MetaData: mtypes.Metadata{
					Name:        "testConsumer",
					Description: "Test Consumer used for unit testing 2",
					UserData1:   "update data1",
					UserData2:   "update data2",
				},
				Spec: hpaModel.HpaResourceConsumerSpec{
					Replicas:      1,
					Name:          "deployment-1",
					ContainerName: "container-1",
				},
			},
			ConsumerClient: &mockIntentManager{
				//Items that will be returned by the mocked Client
				ConsumerItems: []hpaModel.HpaResourceConsumer{
					{
						MetaData: mtypes.Metadata{
							Name:        "testConsumer",
							Description: "Test Consumer used for unit testing 2",
							UserData1:   "update data1",
							UserData2:   "update data2",
						},
						Spec: hpaModel.HpaResourceConsumerSpec{
							Replicas:      1,
							Name:          "deployment-1",
							ContainerName: "container-1",
						},
					},
				},
			},
		},
		{
			label:        "Failed Update Consumer due to internal server error",
			name:         "testConsumer",
			expectedCode: http.StatusInternalServerError,
			reader: bytes.NewBuffer([]byte(`{
				"metadata" : {
					"name": "testConsumer",
    				"description": "Test Consumer used for unit testing",
    				"userData1": "update data1",
    				"userData2": "update data2"
				},
				"spec" : {
					"replicas": 1,
					"name":          "deployment-1",
					"container-name": "container-1"
				}
			}`)),
			ConsumerClient: &mockIntentManager{
				//Items that will be returned by the mocked Client
				ConsumerItems: []hpaModel.HpaResourceConsumer{},
				Err:           pkgerrors.New("inernal"),
			},
		},
		{
			label:        "Failed Update Consumer due to not found error",
			name:         "testConsumer",
			expectedCode: http.StatusNotFound,
			reader: bytes.NewBuffer([]byte(`{
				"metadata" : {
					"name": "testConsumer",
    				"description": "Test Consumer used for unit testing",
    				"userData1": "update data1",
    				"userData2": "update data2"
				},
				"spec" : {
					"replicas": 1,
					"name":          "deployment-1",
					"container-name": "container-1"
				}
			}`)),
			ConsumerClient: &mockIntentManager{
				//Items that will be returned by the mocked Client
				ConsumerItems: []hpaModel.HpaResourceConsumer{},
				Err:           pkgerrors.New("not found"),
			},
		},
		{
			label:        "Update Consumer with consumer-name mismatch",
			name:         "testConsumer1",
			expectedCode: http.StatusBadRequest,
			reader: bytes.NewBuffer([]byte(`{
				"metadata" : {
					"name": "testConsumer",
    				"description": "Test Consumer used for unit testing",
    				"userData1": "update data1",
    				"userData2": "update data2"
				},
				"spec" : {
					"replicas": 1,
					"name":          "deployment-1",
					"container-name": "container-1"
				}
			}`)),
		},
		{
			label:        "Update Consumer with req empty consumer-name",
			name:         "testConsumer",
			expectedCode: http.StatusBadRequest,
			reader: bytes.NewBuffer([]byte(`{
				"metadata" : {
					"name": "",
    				"description": "Test Consumer used for unit testing",
    				"userData1": "update data1",
    				"userData2": "update data2"
				},
				"spec" : {
					"replicas": 1,
					"name":          "deployment-1",
					"container-name": "container-1"
				}
			}`)),
		},
		{
			label:        "Update Consumer with empty consumer-name",
			name:         "",
			expectedCode: http.StatusNotFound,
			reader: bytes.NewBuffer([]byte(`{
				"metadata" : {
					"name": "testConsumer",
    				"description": "Test Consumer used for unit testing",
    				"userData1": "update data1",
    				"userData2": "update data2"
				},
				"spec" : {
					"replicas": 1,
					"name":          "deployment-1",
					"container-name": "container-1"
				}
			}`)),
		},
	}

	fmt.Printf("\n================== TestConsumerUpdateHandler .. total_testcase_count[%d] ==================\n", len(testCases))
	for i, testCase := range testCases {
		fmt.Printf("\n================== TestConsumerUpdateHandler .. testcase_count[%d] testcase_name[%s] ==================\n", i, testCase.label)
		t.Run(testCase.label, func(t *testing.T) {
			request := httptest.NewRequest("PUT", "/v2/projects/project1/composite-apps/compositeapp1/v2/deployment-intent-groups/digroup/hpa-intents/hpaintent1/hpa-resource-consumers/"+testCase.name, testCase.reader)
			resp := executeRequest(request, NewRouter(testCase.ConsumerClient))

			//Check returned code
			if resp.StatusCode != testCase.expectedCode {
				t.Fatalf("Expected %d; Got: %d", testCase.expectedCode, resp.StatusCode)
			}

			//Check returned body only if statusOK
			if resp.StatusCode == http.StatusOK {
				got := hpaModel.HpaResourceConsumer{}
				json.NewDecoder(resp.Body).Decode(&got)

				if reflect.DeepEqual(testCase.expected, got) == false {
					t.Errorf("updateHandler returned unexpected body: got %v;"+
						" expected %v", got, testCase.expected)
				}
			}
		})
	}
}

func TestConsumerGetHandler(t *testing.T) {

	testCases := []struct {
		label          string
		expected       hpaModel.HpaResourceConsumer
		name, version  string
		expectedCode   int
		ConsumerClient *mockIntentManager
	}{
		{
			label:        "Get Consumer",
			expectedCode: http.StatusOK,
			expected: hpaModel.HpaResourceConsumer{
				MetaData: mtypes.Metadata{
					Name:        "testConsumer",
					Description: "Test Consumer used for unit testing",
					UserData1:   "data1",
					UserData2:   "data2",
				},
				Spec: hpaModel.HpaResourceConsumerSpec{
					Name:          "deployment-1",
					ContainerName: "container-1",
				},
			},
			name: "testConsumer",
			ConsumerClient: &mockIntentManager{
				ConsumerItems: []hpaModel.HpaResourceConsumer{
					{
						MetaData: mtypes.Metadata{
							Name:        "testConsumer",
							Description: "Test Consumer used for unit testing",
							UserData1:   "data1",
							UserData2:   "data2",
						},
						Spec: hpaModel.HpaResourceConsumerSpec{
							Name:          "deployment-1",
							ContainerName: "container-1",
						},
					},
				},
			},
		},
		{
			label:        "Get Non-Exiting Consumer",
			expectedCode: http.StatusNotFound,
			name:         "nonexistingConsumer",
			ConsumerClient: &mockIntentManager{
				ConsumerItems: []hpaModel.HpaResourceConsumer{},
				Err:           pkgerrors.New("Internal Error"),
			},
		},
	}

	fmt.Printf("\n================== TestConsumerGetHandler .. total_testcase_count[%d] ==================\n", len(testCases))
	for i, testCase := range testCases {
		fmt.Printf("\n================== TestConsumerGetHandler .. testcase_count[%d] testcase_name[%s] ==================\n", i, testCase.label)
		t.Run(testCase.label, func(t *testing.T) {
			request := httptest.NewRequest("GET", "/v2/projects/project1/composite-apps/compositeapp1/v2/deployment-intent-groups/digroup/hpa-intents/hpaintent1/hpa-resource-consumers/"+testCase.name, nil)
			resp := executeRequest(request, NewRouter(testCase.ConsumerClient))

			//Check returned code
			if resp.StatusCode != testCase.expectedCode {
				t.Fatalf("Expected %d; Got: %d", testCase.expectedCode, resp.StatusCode)
			}

			//Check returned body only if statusOK
			if resp.StatusCode == http.StatusOK {
				got := hpaModel.HpaResourceConsumer{}
				json.NewDecoder(resp.Body).Decode(&got)

				if reflect.DeepEqual(testCase.expected, got) == false {
					t.Errorf("TestConsumerGetHandler returned unexpected body: got %v;"+
						" expected %v", got, testCase.expected)
				}
			}
		})
	}
}

func TestConsumerGetHandlerByName(t *testing.T) {

	testCases := []struct {
		label          string
		expected       hpaModel.HpaResourceConsumer
		name, version  string
		expectedCode   int
		ConsumerClient *mockIntentManager
	}{
		{
			label:        "GetConsumerByName Consumer",
			expectedCode: http.StatusOK,
			expected: hpaModel.HpaResourceConsumer{
				MetaData: mtypes.Metadata{
					Name:        "testConsumer",
					Description: "Test Consumer used for unit testing",
					UserData1:   "data1",
					UserData2:   "data2",
				},
				Spec: hpaModel.HpaResourceConsumerSpec{
					Name:          "deployment-1",
					ContainerName: "container-1",
				},
			},
			name: "testConsumer",
			ConsumerClient: &mockIntentManager{
				ConsumerItems: []hpaModel.HpaResourceConsumer{
					{
						MetaData: mtypes.Metadata{
							Name:        "testConsumer",
							Description: "Test Consumer used for unit testing",
							UserData1:   "data1",
							UserData2:   "data2",
						},
						Spec: hpaModel.HpaResourceConsumerSpec{
							Name:          "deployment-1",
							ContainerName: "container-1",
						},
					},
				},
			},
		},
		{
			label:        "GetConsumerByName Non-Exiting Consumer",
			expectedCode: http.StatusNotFound,
			name:         "nonexistingConsumer",
			ConsumerClient: &mockIntentManager{
				ConsumerItemsSpec: []hpaModel.HpaResourceConsumerSpec{},
				Err:               pkgerrors.New("Internal Error"),
			},
		},
		{
			label:        "GetConsumerByName Non-Exiting empty Consumer",
			expectedCode: http.StatusBadRequest,
			name:         "",
			ConsumerClient: &mockIntentManager{
				ConsumerItemsSpec: []hpaModel.HpaResourceConsumerSpec{},
				Err:               pkgerrors.New("Internal Error"),
			},
		},
	}

	fmt.Printf("\n================== TestConsumerGetHandler .. total_testcase_count[%d] ==================\n", len(testCases))
	for i, testCase := range testCases {
		fmt.Printf("\n================== TestConsumerGetHandler .. testcase_count[%d] testcase_name[%s] ==================\n", i, testCase.label)
		t.Run(testCase.label, func(t *testing.T) {
			request := httptest.NewRequest("QUERY", "/v2/projects/project1/composite-apps/compositeapp1/v2/deployment-intent-groups/digroup/hpa-intents/hpaintent1/hpa-resource-consumers?consumer="+testCase.name, nil)
			resp := executeRequest(request, NewRouter(testCase.ConsumerClient))

			//Check returned code
			if resp.StatusCode != testCase.expectedCode {
				t.Fatalf("Expected %d; Got: %d", testCase.expectedCode, resp.StatusCode)
			}

			//Check returned body only if statusOK
			if resp.StatusCode == http.StatusOK {
				got := hpaModel.HpaResourceConsumer{}
				json.NewDecoder(resp.Body).Decode(&got)

				if reflect.DeepEqual(testCase.expected, got) == false {
					t.Errorf("TestConsumerGetHandler returned unexpected body: got %v;"+
						" expected %v", got, testCase.expected)
				}
			}
		})
	}
}
func TestConsumerGetAllHandler(t *testing.T) {

	testCases := []struct {
		label          string
		expected       []hpaModel.HpaResourceConsumer
		name, version  string
		expectedCode   int
		ConsumerClient *mockIntentManager
	}{
		{
			label:        "GetAll Consumer",
			expectedCode: http.StatusOK,
			expected: []hpaModel.HpaResourceConsumer{
				{
					MetaData: mtypes.Metadata{
						Name:        "testConsumer",
						Description: "Test Consumer used for unit testing",
						UserData1:   "data1",
						UserData2:   "data2",
					},
					Spec: hpaModel.HpaResourceConsumerSpec{
						Name:          "deployment-1",
						ContainerName: "container-1",
					},
				},
			},
			ConsumerClient: &mockIntentManager{
				ConsumerItems: []hpaModel.HpaResourceConsumer{
					{
						MetaData: mtypes.Metadata{
							Name:        "testConsumer",
							Description: "Test Consumer used for unit testing",
							UserData1:   "data1",
							UserData2:   "data2",
						},
						Spec: hpaModel.HpaResourceConsumerSpec{
							Name:          "deployment-1",
							ContainerName: "container-1",
						},
					},
				},
			},
		},
		{
			label:        "GetAll Non-Exiting Consumers",
			expectedCode: http.StatusNotFound,
			ConsumerClient: &mockIntentManager{
				ConsumerItems: []hpaModel.HpaResourceConsumer{},
				Err:           pkgerrors.New("Internal Error"),
			},
		},
		{
			label:        "GetAll No Consumers",
			expectedCode: http.StatusOK,
			ConsumerClient: &mockIntentManager{
				ConsumerItems: []hpaModel.HpaResourceConsumer{},
			},
			expected: []hpaModel.HpaResourceConsumer{},
		},
	}

	fmt.Printf("\n================== TestConsumerGetAllHandler .. total_testcase_count[%d] ==================\n", len(testCases))
	for i, testCase := range testCases {
		fmt.Printf("\n================== TestConsumerGetAllHandler .. testcase_count[%d] testcase_name[%s] ==================\n", i, testCase.label)
		t.Run(testCase.label, func(t *testing.T) {
			request := httptest.NewRequest("GET", "/v2/projects/project1/composite-apps/compositeapp1/v2/deployment-intent-groups/digroup/hpa-intents/hpaintent1/hpa-resource-consumers", nil)
			resp := executeRequest(request, NewRouter(testCase.ConsumerClient))

			//Check returned code
			if resp.StatusCode != testCase.expectedCode {
				t.Fatalf("Expected %d; Got: %d", testCase.expectedCode, resp.StatusCode)
			}

			//Check returned body only if statusOK
			if resp.StatusCode == http.StatusOK {
				got := []hpaModel.HpaResourceConsumer{}
				json.NewDecoder(resp.Body).Decode(&got)

				if reflect.DeepEqual(testCase.expected, got) == false {
					t.Errorf("TestConsumerGetAllHandler returned unexpected body: got %v;"+
						" expected %v", got, testCase.expected)
				}
			}
		})
	}
}

func TestConsumerDeleteHandler(t *testing.T) {

	testCases := []struct {
		label          string
		name           string
		version        string
		expectedCode   int
		ConsumerClient *mockIntentManager
	}{
		{
			label:        "Delete Consumer",
			expectedCode: http.StatusNoContent,
			name:         "testConsumer",
			ConsumerClient: &mockIntentManager{
				//Items that will be returned by the mocked Client
				ConsumerItems: []hpaModel.HpaResourceConsumer{
					{
						MetaData: mtypes.Metadata{
							Name:        "testConsumer",
							Description: "Test Consumer used for unit testing",
							UserData1:   "data1",
							UserData2:   "data2",
						},
					},
				},
			},
		},
		{
			label:        "Delete Non-Exiting Consumer",
			expectedCode: http.StatusNotFound,
			name:         "testConsumer",
			ConsumerClient: &mockIntentManager{
				Err: pkgerrors.New("Internal Error"),
			},
		},
		{
			label:        "Delete Non-Exiting empty Consumer",
			expectedCode: http.StatusNotFound,
			name:         "",
			ConsumerClient: &mockIntentManager{
				Err: pkgerrors.New("Internal Error"),
			},
		},
	}

	fmt.Printf("\n================== TestConsumerDeleteHandler .. total_testcase_count[%d] ==================\n", len(testCases))
	for i, testCase := range testCases {
		fmt.Printf("\n================== TestConsumerDeleteHandler .. testcase_count[%d] testcase_name[%s] ==================\n", i, testCase.label)
		t.Run(testCase.label, func(t *testing.T) {
			request := httptest.NewRequest("DELETE", "/v2/projects/project1/composite-apps/compositeapp1/v2/deployment-intent-groups/digroup/hpa-intents/hpaintent1/hpa-resource-consumers/"+testCase.name, nil)
			resp := executeRequest(request, NewRouter(testCase.ConsumerClient))

			//Check returned code
			if resp.StatusCode != testCase.expectedCode {
				t.Fatalf("Expected %d; Got: %d", testCase.expectedCode, resp.StatusCode)
			}
		})
	}
}

func TestConsumerDeleteAllHandler(t *testing.T) {

	testCases := []struct {
		label          string
		name           string
		version        string
		expectedCode   int
		ConsumerClient *mockIntentManager
	}{
		{
			label:        "Delete All Consumers",
			expectedCode: http.StatusNoContent,
			ConsumerClient: &mockIntentManager{
				//Items that will be returned by the mocked Client
				ConsumerItems: []hpaModel.HpaResourceConsumer{
					{
						MetaData: mtypes.Metadata{
							Name:        "testConsumer",
							Description: "Test Consumer used for unit testing",
							UserData1:   "data1",
							UserData2:   "data2",
						},
					},
				},
			},
		},
		{
			label:        "Delete Non-Exiting Consumer",
			expectedCode: http.StatusNotFound,
			ConsumerClient: &mockIntentManager{
				Err: pkgerrors.New("Internal Error"),
			},
		},
	}

	fmt.Printf("\n================== TestConsumerDeleteAllHandler .. total_testcase_count[%d] ==================\n", len(testCases))
	for i, testCase := range testCases {
		fmt.Printf("\n================== TestConsumerDeleteAllHandler .. testcase_count[%d] testcase_name[%s] ==================\n", i, testCase.label)
		t.Run(testCase.label, func(t *testing.T) {
			request := httptest.NewRequest("DELETE", "/v2/projects/project1/composite-apps/compositeapp1/v2/deployment-intent-groups/digroup/hpa-intents/hpaintent1/hpa-resource-consumers", nil)
			resp := executeRequest(request, NewRouter(testCase.ConsumerClient))

			//Check returned code
			if resp.StatusCode != testCase.expectedCode {
				t.Fatalf("Expected %d; Got: %d", testCase.expectedCode, resp.StatusCode)
			}
		})
	}
}
