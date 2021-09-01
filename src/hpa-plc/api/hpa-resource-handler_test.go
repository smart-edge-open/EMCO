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
	hpaResourceJSONFile = "../json-schemas/placement-hpa-resource.json"
	orchLog.SetLoglevel(logrus.InfoLevel)
}

var allocatable_true = true
var allocatable_false = false

func TestResourceCreateHandler(t *testing.T) {
	testCases := []struct {
		label          string
		reader         io.Reader
		expected       hpaModel.HpaResourceRequirement
		expectedCode   int
		ResourceClient *mockIntentManager
	}{
		{
			label:        "Create non-allocatable Resource",
			expectedCode: http.StatusCreated,
			reader: bytes.NewBuffer([]byte(`{
				"metadata" : {
					"name": "testResource",
    				"description": "Test Resource used for unit testing",
    				"userData1": "data1",
    				"userData2": "data2"
				},
				"spec" : {
					"allocatable":          false,
					"resource" : {"key":"vpu", "value":"yes"}
				}
			}`)),
			expected: hpaModel.HpaResourceRequirement{
				MetaData: mtypes.Metadata{
					Name:        "testResource",
					Description: "Test Resource used for unit testing",
					UserData1:   "data1",
					UserData2:   "data2",
				},
				Spec: hpaModel.HpaResourceRequirementSpec{
					Allocatable: &allocatable_false,
					Resource:    hpaModel.HpaResourceRequirementDetails{hpaModel.AllocatableResources{}, hpaModel.NonAllocatableResources{Key: "vpu", Value: "yes"}},
				},
			},
			ResourceClient: &mockIntentManager{
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
				//Items that will be returned by the mocked Client
				ResourceItems: []hpaModel.HpaResourceRequirement{
					{
						MetaData: mtypes.Metadata{
							Name:        "testResource",
							Description: "Test Resource used for unit testing",
							UserData1:   "data1",
							UserData2:   "data2",
						},
						Spec: hpaModel.HpaResourceRequirementSpec{
							Allocatable: &allocatable_false,
							Resource:    hpaModel.HpaResourceRequirementDetails{hpaModel.AllocatableResources{}, hpaModel.NonAllocatableResources{Key: "vpu", Value: "yes"}},
						},
					},
				},
			},
		},
		{
			label:        "Create allocatable Resource when requests = limits",
			expectedCode: http.StatusCreated,
			reader: bytes.NewBuffer([]byte(`{
				"metadata" : {
					"name": "testResource",
    				"description": "Test Resource used for unit testing",
    				"userData1": "data1",
    				"userData2": "data2"
				},
				"spec" : {
					"allocatable":          true,
					"resource" : {"name":"cpu", "requests":1,"limits":1}
				}
			}`)),
			expected: hpaModel.HpaResourceRequirement{
				MetaData: mtypes.Metadata{
					Name:        "testResource",
					Description: "Test Resource used for unit testing",
					UserData1:   "data1",
					UserData2:   "data2",
				},
				Spec: hpaModel.HpaResourceRequirementSpec{
					Allocatable: &allocatable_true,
					Resource:    hpaModel.HpaResourceRequirementDetails{hpaModel.AllocatableResources{Name: "cpu", Requests: 1, Limits: 1}, hpaModel.NonAllocatableResources{}},
				},
			},
			ResourceClient: &mockIntentManager{
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
				//Items that will be returned by the mocked Client
				ResourceItems: []hpaModel.HpaResourceRequirement{
					{
						MetaData: mtypes.Metadata{
							Name:        "testResource",
							Description: "Test Resource used for unit testing",
							UserData1:   "data1",
							UserData2:   "data2",
						},
						Spec: hpaModel.HpaResourceRequirementSpec{
							Allocatable: &allocatable_true,
							Resource:    hpaModel.HpaResourceRequirementDetails{hpaModel.AllocatableResources{Name: "cpu", Requests: 1, Limits: 1}, hpaModel.NonAllocatableResources{}},
						},
					},
				},
			},
		},
		{
			label:        "Create allocatable Resource requests < limits",
			expectedCode: http.StatusCreated,
			reader: bytes.NewBuffer([]byte(`{
				"metadata" : {
					"name": "testResource",
    				"description": "Test Resource used for unit testing",
    				"userData1": "data1",
    				"userData2": "data2"
				},
				"spec" : {
					"allocatable":          true,
					"resource" : {"name":"cpu", "requests":2,"limits":3}
				}
			}`)),
			expected: hpaModel.HpaResourceRequirement{
				MetaData: mtypes.Metadata{
					Name:        "testResource",
					Description: "Test Resource used for unit testing",
					UserData1:   "data1",
					UserData2:   "data2",
				},
				Spec: hpaModel.HpaResourceRequirementSpec{
					Allocatable: &allocatable_true,
					Resource:    hpaModel.HpaResourceRequirementDetails{hpaModel.AllocatableResources{Name: "cpu", Requests: 1, Limits: 1}, hpaModel.NonAllocatableResources{}},
				},
			},
			ResourceClient: &mockIntentManager{
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
				//Items that will be returned by the mocked Client
				ResourceItems: []hpaModel.HpaResourceRequirement{
					{
						MetaData: mtypes.Metadata{
							Name:        "testResource",
							Description: "Test Resource used for unit testing",
							UserData1:   "data1",
							UserData2:   "data2",
						},
						Spec: hpaModel.HpaResourceRequirementSpec{
							Allocatable: &allocatable_true,
							Resource:    hpaModel.HpaResourceRequirementDetails{hpaModel.AllocatableResources{Name: "cpu", Requests: 1, Limits: 1}, hpaModel.NonAllocatableResources{}},
						},
					},
				},
			},
		},
		{
			label:        "Create allocatable Resource even if limits is zero",
			expectedCode: http.StatusCreated,
			reader: bytes.NewBuffer([]byte(`{
				"metadata" : {
					"name": "testResource",
    				"description": "Test Resource used for unit testing",
    				"userData1": "data1",
    				"userData2": "data2"
				},
				"spec" : {
					"allocatable":          true,
					"resource" : {"name":"cpu", "requests":1}
				}
			}`)),
			expected: hpaModel.HpaResourceRequirement{
				MetaData: mtypes.Metadata{
					Name:        "testResource",
					Description: "Test Resource used for unit testing",
					UserData1:   "data1",
					UserData2:   "data2",
				},
				Spec: hpaModel.HpaResourceRequirementSpec{
					Allocatable: &allocatable_true,
					Resource:    hpaModel.HpaResourceRequirementDetails{hpaModel.AllocatableResources{Name: "cpu", Requests: 1, Limits: 1}, hpaModel.NonAllocatableResources{}},
				},
			},
			ResourceClient: &mockIntentManager{
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
				//Items that will be returned by the mocked Client
				ResourceItems: []hpaModel.HpaResourceRequirement{
					{
						MetaData: mtypes.Metadata{
							Name:        "testResource",
							Description: "Test Resource used for unit testing",
							UserData1:   "data1",
							UserData2:   "data2",
						},
						Spec: hpaModel.HpaResourceRequirementSpec{
							Allocatable: &allocatable_true,
							Resource:    hpaModel.HpaResourceRequirementDetails{hpaModel.AllocatableResources{Name: "cpu", Requests: 1, Limits: 1}, hpaModel.NonAllocatableResources{}},
						},
					},
				},
			},
		},
		{
			label:          "Missing Body Failure",
			expectedCode:   http.StatusBadRequest,
			ResourceClient: &mockIntentManager{},
		},
		{
			label:        "Failed Create non-allocatable Resource due to not specifying allocatable",
			expectedCode: http.StatusBadRequest,
			reader: bytes.NewBuffer([]byte(`{
				"metadata" : {
					"name": "testResource",
    				"description": "Test Resource used for unit testing",
    				"userData1": "data1",
    				"userData2": "data2"
				},
				"spec" : {
					"resource" : {"key":"vpu", "value":"yes"}
				}
			}`)),
		},
		{
			label:        "Failed to Create allocatable Resource when not specifying allocatable field",
			expectedCode: http.StatusBadRequest,
			reader: bytes.NewBuffer([]byte(`{
				"metadata" : {
					"name": "testResource",
    				"description": "Test Resource used for unit testing",
    				"userData1": "data1",
    				"userData2": "data2"
				},
				"spec" : {
					"resource" : {"name":"cpu", "requests":1,"limits":1}
				}
			}`)),
		},
		{
			label:        "Create Resource Failed due to allocatable Resource requests > limits",
			expectedCode: http.StatusBadRequest,
			reader: bytes.NewBuffer([]byte(`{
				"metadata" : {
					"name": "testResource",
    				"description": "Test Resource used for unit testing",
    				"userData1": "data1",
    				"userData2": "data2"
				},
				"spec" : {
					"allocatable":          true,
					"resource" : {"name":"cpu", "requests":3,"limits":2}
				}
			}`)),
			ResourceClient: &mockIntentManager{
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
				//Items that will be returned by the mocked Client
				ResourceItems: []hpaModel.HpaResourceRequirement{
					{
						MetaData: mtypes.Metadata{
							Name:        "testResource",
							Description: "Test Resource used for unit testing",
							UserData1:   "data1",
							UserData2:   "data2",
						},
						Spec: hpaModel.HpaResourceRequirementSpec{
							Allocatable: &allocatable_true,
							Resource:    hpaModel.HpaResourceRequirementDetails{hpaModel.AllocatableResources{Name: "cpu", Requests: 2, Limits: 3}, hpaModel.NonAllocatableResources{}},
						},
					},
				},
			},
		},
		{
			label:        "Failed Create Resource due to not found error status",
			expectedCode: http.StatusNotFound,
			reader: bytes.NewBuffer([]byte(`{
				"metadata" : {
					"name": "testResource",
    				"description": "Test Resource used for unit testing",
    				"userData1": "data1",
    				"userData2": "data2"
				},
				"spec" : {
					"allocatable":          true,
					"resource" : {"name":"cpu", "requests":1,"limits":1}
				}
			}`)),
			ResourceClient: &mockIntentManager{
				//Items that will be returned by the mocked Client
				ConsumerItems: []hpaModel.HpaResourceConsumer{
					{
						MetaData: mtypes.Metadata{
							Name:        "consumer1",
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
				ResourceItems: []hpaModel.HpaResourceRequirement{},
				Err:           pkgerrors.New("internal"),
			},
		},
		{
			label:        "Failed Create non-allocatable Resource due to bad request body",
			expectedCode: http.StatusUnprocessableEntity,
			reader: bytes.NewBuffer([]byte(`{
				"metadata" : {
					"name": "testResource",
    				"description": "Test Resource used for unit testing",
    				"userData1": "data1",
    				"userData2": "data2"
				},
				"spec" : {
					"allocatable":          false,
					"resource" : {"key":"vpu", "value":"yes"
				}
			}`)),
		},
		{
			label:        "Missing limits in allocatable Resource",
			expectedCode: http.StatusCreated,
			reader: bytes.NewBuffer([]byte(`{
				"metadata" : {
					"name": "testResource",
    				"description": "Test Resource used for unit testing",
    				"userData1": "data1",
    				"userData2": "data2"
				},
				"spec" : {
					"allocatable":          true,
					"resource" : {"name":"cpu", "requests":1}
				}
			}`)),
			expected: hpaModel.HpaResourceRequirement{
				MetaData: mtypes.Metadata{
					Name:        "testResource",
					Description: "Test Resource used for unit testing",
					UserData1:   "data1",
					UserData2:   "data2",
				},
				Spec: hpaModel.HpaResourceRequirementSpec{
					Allocatable: &allocatable_true,
					Resource:    hpaModel.HpaResourceRequirementDetails{hpaModel.AllocatableResources{Name: "cpu", Requests: 1}, hpaModel.NonAllocatableResources{}},
				},
			},
			ResourceClient: &mockIntentManager{
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
				//Items that will be returned by the mocked Client
				ResourceItems: []hpaModel.HpaResourceRequirement{
					{
						MetaData: mtypes.Metadata{
							Name:        "testResource",
							Description: "Test Resource used for unit testing",
							UserData1:   "data1",
							UserData2:   "data2",
						},
						Spec: hpaModel.HpaResourceRequirementSpec{
							Allocatable: &allocatable_true,
							Resource:    hpaModel.HpaResourceRequirementDetails{hpaModel.AllocatableResources{Name: "cpu", Requests: 1}, hpaModel.NonAllocatableResources{}},
						},
					},
				},
			},
		},
		{
			label: "Missing Resource Name in Request Body",
			reader: bytes.NewBuffer([]byte(`{
                "description":"test description"
                }`)),
			expectedCode:   http.StatusBadRequest,
			ResourceClient: &mockIntentManager{},
		},
		{
			label: "Empty Resource Name in Request Body",
			reader: bytes.NewBuffer([]byte(`{
				"name": "",
                "description":"test description"
                }`)),
			expectedCode:   http.StatusBadRequest,
			ResourceClient: &mockIntentManager{},
		},
		{
			label:        "Missing deployment name in consumer spec while creating non-allocatable Resource",
			expectedCode: http.StatusBadRequest,
			reader: bytes.NewBuffer([]byte(`{
				"metadata" : {
					"name": "testResource",
    				"description": "Test Resource used for unit testing",
    				"userData1": "data1",
    				"userData2": "data2"
				},
				"spec" : {
					"allocatable":          true,
					"resource" : {"key":"vpu", "value":"yes"}
				}
			}`)),
			ResourceClient: &mockIntentManager{
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
							ContainerName: "container-1",
						},
					},
				},
				//Items that will be returned by the mocked Client
				ResourceItems: []hpaModel.HpaResourceRequirement{
					{
						MetaData: mtypes.Metadata{
							Name:        "testResource",
							Description: "Test Resource used for unit testing",
							UserData1:   "data1",
							UserData2:   "data2",
						},
						Spec: hpaModel.HpaResourceRequirementSpec{
							Allocatable: &allocatable_false,
							Resource:    hpaModel.HpaResourceRequirementDetails{hpaModel.AllocatableResources{}, hpaModel.NonAllocatableResources{Key: "vpu", Value: "yes"}},
						},
					},
				},
			},
		},
		{
			label:        "Missing deployment name in consumer spec while creating allocatable Resource",
			expectedCode: http.StatusBadRequest,
			reader: bytes.NewBuffer([]byte(`{
				"metadata" : {
					"name": "testResource",
    				"description": "Test Resource used for unit testing",
    				"userData1": "data1",
    				"userData2": "data2"
				},
				"spec" : {
					"allocatable":          true,
					"resource" : {"name":"cpu", "requests":1,"limits":1}
				}
			}`)),
			ResourceClient: &mockIntentManager{
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
							ContainerName: "container-1",
						},
					},
				},
				//Items that will be returned by the mocked Client
				ResourceItems: []hpaModel.HpaResourceRequirement{
					{
						MetaData: mtypes.Metadata{
							Name:        "testResource",
							Description: "Test Resource used for unit testing",
							UserData1:   "data1",
							UserData2:   "data2",
						},
						Spec: hpaModel.HpaResourceRequirementSpec{
							Allocatable: &allocatable_true,
							Resource:    hpaModel.HpaResourceRequirementDetails{hpaModel.AllocatableResources{Name: "cpu", Requests: 1, Limits: 1}, hpaModel.NonAllocatableResources{}},
						},
					},
				},
			},
		},
		{
			label:        "Missing container name in consumer spec while creating allocatable Resource",
			expectedCode: http.StatusBadRequest,
			reader: bytes.NewBuffer([]byte(`{
				"metadata" : {
					"name": "testResource",
    				"description": "Test Resource used for unit testing",
    				"userData1": "data1",
    				"userData2": "data2"
				},
				"spec" : {
					"allocatable":          true,
					"resource" : {"name":"cpu", "requests":1,"limits":1}
				}
			}`)),
			ResourceClient: &mockIntentManager{
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
							Name: "deployment-1",
						},
					},
				},
				//Items that will be returned by the mocked Client
				ResourceItems: []hpaModel.HpaResourceRequirement{
					{
						MetaData: mtypes.Metadata{
							Name:        "testResource",
							Description: "Test Resource used for unit testing",
							UserData1:   "data1",
							UserData2:   "data2",
						},
						Spec: hpaModel.HpaResourceRequirementSpec{
							Allocatable: &allocatable_true,
							Resource:    hpaModel.HpaResourceRequirementDetails{hpaModel.AllocatableResources{Name: "cpu", Requests: 1, Limits: 1}, hpaModel.NonAllocatableResources{}},
						},
					},
				},
			},
		},
		{
			label:        "Missing value(of key-value pair) of non-allocatable resource",
			expectedCode: http.StatusBadRequest,
			reader: bytes.NewBuffer([]byte(`{
				"metadata" : {
					"name": "testResource",
    				"description": "Test Resource used for unit testing",
    				"userData1": "data1",
    				"userData2": "data2"
				},
				"spec" : {
					"allocatable":          true,
					"resource" : {"key":"yes"}
				}
			}`)),
			ResourceClient: &mockIntentManager{
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
				//Items that will be returned by the mocked Client
				ResourceItems: []hpaModel.HpaResourceRequirement{
					{
						MetaData: mtypes.Metadata{
							Name:        "testResource",
							Description: "Test Resource used for unit testing",
							UserData1:   "data1",
							UserData2:   "data2",
						},
						Spec: hpaModel.HpaResourceRequirementSpec{
							Allocatable: &allocatable_false,
							Resource:    hpaModel.HpaResourceRequirementDetails{hpaModel.AllocatableResources{}, hpaModel.NonAllocatableResources{Key: "yes"}},
						},
					},
				},
			},
		},
		{
			label:        "Missing key-value pair of non-allocatable resource",
			expectedCode: http.StatusBadRequest,
			reader: bytes.NewBuffer([]byte(`{
				"metadata" : {
					"name": "testResource",
    				"description": "Test Resource used for unit testing",
    				"userData1": "data1",
    				"userData2": "data2"
				},
				"spec" : {
					"allocatable":          true,
					"resource" : {}
				}
			}`)),
			ResourceClient: &mockIntentManager{
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
				//Items that will be returned by the mocked Client
				ResourceItems: []hpaModel.HpaResourceRequirement{
					{
						MetaData: mtypes.Metadata{
							Name:        "testResource",
							Description: "Test Resource used for unit testing",
							UserData1:   "data1",
							UserData2:   "data2",
						},
						Spec: hpaModel.HpaResourceRequirementSpec{
							Allocatable: &allocatable_false,
							Resource:    hpaModel.HpaResourceRequirementDetails{hpaModel.AllocatableResources{}, hpaModel.NonAllocatableResources{}},
						},
					},
				},
			},
		},
		{
			label:        "Missing name in allocatable resource",
			expectedCode: http.StatusBadRequest,
			reader: bytes.NewBuffer([]byte(`{
				"metadata" : {
					"name": "testResource",
    				"description": "Test Resource used for unit testing",
    				"userData1": "data1",
    				"userData2": "data2"
				},
				"spec" : {
					"allocatable":          true,
					"resource" : {"requests":1,"limits":1}
				}
			}`)),
			expected: hpaModel.HpaResourceRequirement{
				MetaData: mtypes.Metadata{
					Name:        "testResource",
					Description: "Test Resource used for unit testing",
					UserData1:   "data1",
					UserData2:   "data2",
				},
				Spec: hpaModel.HpaResourceRequirementSpec{
					Allocatable: &allocatable_true,
					Resource:    hpaModel.HpaResourceRequirementDetails{hpaModel.AllocatableResources{Requests: 1, Limits: 1}, hpaModel.NonAllocatableResources{}},
				},
			},
			ResourceClient: &mockIntentManager{
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
				//Items that will be returned by the mocked Client
				ResourceItems: []hpaModel.HpaResourceRequirement{
					{
						MetaData: mtypes.Metadata{
							Name:        "testResource",
							Description: "Test Resource used for unit testing",
							UserData1:   "data1",
							UserData2:   "data2",
						},
						Spec: hpaModel.HpaResourceRequirementSpec{
							Allocatable: &allocatable_true,
							Resource:    hpaModel.HpaResourceRequirementDetails{hpaModel.AllocatableResources{Requests: 1, Limits: 1}, hpaModel.NonAllocatableResources{}},
						},
					},
				},
			},
		},
		{
			label:        "Missing requests in allocatable resource",
			expectedCode: http.StatusBadRequest,
			reader: bytes.NewBuffer([]byte(`{
				"metadata" : {
					"name": "testResource",
    				"description": "Test Resource used for unit testing",
    				"userData1": "data1",
    				"userData2": "data2"
				},
				"spec" : {
					"allocatable":          true,
					"resource" : {"name":"cpu", "limits":1}
				}
			}`)),
			expected: hpaModel.HpaResourceRequirement{
				MetaData: mtypes.Metadata{
					Name:        "testResource",
					Description: "Test Resource used for unit testing",
					UserData1:   "data1",
					UserData2:   "data2",
				},
				Spec: hpaModel.HpaResourceRequirementSpec{
					Allocatable: &allocatable_true,
					Resource:    hpaModel.HpaResourceRequirementDetails{hpaModel.AllocatableResources{Name: "cpu", Limits: 1}, hpaModel.NonAllocatableResources{}},
				},
			},
			ResourceClient: &mockIntentManager{
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
				//Items that will be returned by the mocked Client
				ResourceItems: []hpaModel.HpaResourceRequirement{
					{
						MetaData: mtypes.Metadata{
							Name:        "testResource",
							Description: "Test Resource used for unit testing",
							UserData1:   "data1",
							UserData2:   "data2",
						},
						Spec: hpaModel.HpaResourceRequirementSpec{
							Allocatable: &allocatable_true,
							Resource:    hpaModel.HpaResourceRequirementDetails{hpaModel.AllocatableResources{Name: "cpu", Limits: 1}, hpaModel.NonAllocatableResources{}},
						},
					},
				},
			},
		},
		{
			label:        "non-allocatable resource-spec allocatable and resource field value mismatch",
			expectedCode: http.StatusBadRequest,
			reader: bytes.NewBuffer([]byte(`{
				"metadata" : {
					"name": "testResource",
    				"description": "Test Resource used for unit testing",
    				"userData1": "data1",
    				"userData2": "data2"
				},
				"spec" : {
					"allocatable":          false,
					"resource" : {"name":"cpu", "requests":1, "limits":1}
				}
			}`)),
			expected: hpaModel.HpaResourceRequirement{
				MetaData: mtypes.Metadata{
					Name:        "testResource",
					Description: "Test Resource used for unit testing",
					UserData1:   "data1",
					UserData2:   "data2",
				},
				Spec: hpaModel.HpaResourceRequirementSpec{
					Allocatable: &allocatable_true,
					Resource:    hpaModel.HpaResourceRequirementDetails{hpaModel.AllocatableResources{Name: "cpu", Limits: 1}, hpaModel.NonAllocatableResources{}},
				},
			},
			ResourceClient: &mockIntentManager{
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
				//Items that will be returned by the mocked Client
				ResourceItems: []hpaModel.HpaResourceRequirement{
					{
						MetaData: mtypes.Metadata{
							Name:        "testResource",
							Description: "Test Resource used for unit testing",
							UserData1:   "data1",
							UserData2:   "data2",
						},
						Spec: hpaModel.HpaResourceRequirementSpec{
							Allocatable: &allocatable_true,
							Resource:    hpaModel.HpaResourceRequirementDetails{hpaModel.AllocatableResources{Name: "cpu", Limits: 1}, hpaModel.NonAllocatableResources{}},
						},
					},
				},
			},
		},
		{
			label:        "allocatable resource-spec allocatable and resource field value mismatch",
			expectedCode: http.StatusBadRequest,
			reader: bytes.NewBuffer([]byte(`{
				"metadata" : {
					"name": "testResource",
    				"description": "Test Resource used for unit testing",
    				"userData1": "data1",
    				"userData2": "data2"
				},
				"spec" : {
					"allocatable":          true,
					"resource" : {"key":"cpu", "value":"cpu-value"}
				}
			}`)),
			expected: hpaModel.HpaResourceRequirement{
				MetaData: mtypes.Metadata{
					Name:        "testResource",
					Description: "Test Resource used for unit testing",
					UserData1:   "data1",
					UserData2:   "data2",
				},
				Spec: hpaModel.HpaResourceRequirementSpec{
					Allocatable: &allocatable_true,
					Resource:    hpaModel.HpaResourceRequirementDetails{hpaModel.AllocatableResources{Name: "cpu", Limits: 1}, hpaModel.NonAllocatableResources{}},
				},
			},
			ResourceClient: &mockIntentManager{
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
				//Items that will be returned by the mocked Client
				ResourceItems: []hpaModel.HpaResourceRequirement{
					{
						MetaData: mtypes.Metadata{
							Name:        "testResource",
							Description: "Test Resource used for unit testing",
							UserData1:   "data1",
							UserData2:   "data2",
						},
						Spec: hpaModel.HpaResourceRequirementSpec{
							Allocatable: &allocatable_true,
							Resource:    hpaModel.HpaResourceRequirementDetails{hpaModel.AllocatableResources{Name: "cpu", Limits: 1}, hpaModel.NonAllocatableResources{}},
						},
					},
				},
			},
		},
	}

	fmt.Printf("\n================== TestResourceCreateHandler .. total_testcase_count[%d] ==================\n", len(testCases))
	testcasesFailed := make([]string, 0)
	for i, testCase := range testCases {
		fmt.Printf("\n================== TestResourceCreateHandler .. testcase_count[%d] testcase_name[%s] ==================\n", i, testCase.label)
		t.Run(testCase.label, func(t *testing.T) {

			request := httptest.NewRequest("POST", "/v2/projects/project1/composite-apps/compositeapp1/v2/deployment-intent-groups/digroup/hpa-intents/hpaintent1/hpa-resource-consumers/consumer1/resource-requirements", testCase.reader)
			resp := executeRequest(request, NewRouter(testCase.ResourceClient))

			//Check returned code
			if resp.StatusCode != testCase.expectedCode {
				error := fmt.Sprintf("TestCase Failed[%v] => index[%d] Expected %d; Got: %d", testCase.label, i, testCase.expectedCode, resp.StatusCode)
				testcasesFailed = append(testcasesFailed, error)
				t.Fatalf(error)
			}

			//Check returned body only if statusCreated
			if resp.StatusCode == http.StatusCreated {
				got := hpaModel.HpaResourceRequirement{}
				json.NewDecoder(resp.Body).Decode(&got)

				if reflect.DeepEqual(testCase.expected, got) == false {
					t.Errorf("createHandler returned unexpected body: got %v;"+
						" expected %v", got, testCase.expected)
				}
			}
		})
	}
	for _, testCaseFailed := range testcasesFailed {
		fmt.Printf("\n================== TestResourceCreateHandler .. testcase-failed[%v] ==================\n", testCaseFailed)
	}
}

func TestResourceUpdateHandler(t *testing.T) {
	testCases := []struct {
		label, name    string
		reader         io.Reader
		expected       hpaModel.HpaResourceRequirement
		expectedCode   int
		ResourceClient *mockIntentManager
	}{
		{
			label: "Missing Resource Name in Request Body",
			name:  "testResource",
			reader: bytes.NewBuffer([]byte(`{
				"description":"test description"
				}`)),
			expectedCode:   http.StatusBadRequest,
			ResourceClient: &mockIntentManager{},
		},
		{
			label:          "Missing Body Failure",
			name:           "testResource",
			expectedCode:   http.StatusBadRequest,
			ResourceClient: &mockIntentManager{},
		},
		{
			label:        "Mismatched Name Failure",
			name:         "testResource",
			expectedCode: http.StatusBadRequest,
			reader: bytes.NewBuffer([]byte(`{
				"metadata" : {
					"name": "testResourceNameMismatch",
					"description": "Test Resource used for unit testing"
				}
			}`)),
			ResourceClient: &mockIntentManager{},
		},
		{
			label:        "Update Resource when requests = limits",
			name:         "testResource",
			expectedCode: http.StatusOK,
			reader: bytes.NewBuffer([]byte(`{
				"metadata" : {
					"name": "testResource",
    				"description": "Test Resource used for unit testing",
    				"userData1": "update data1",
    				"userData2": "update data2"
				},
				"spec" : {
					"allocatable":          true,
					"resource" : {"name":"cpu", "requests":1,"limits":1}
				}
			}`)),
			expected: hpaModel.HpaResourceRequirement{
				MetaData: mtypes.Metadata{
					Name:        "testResource",
					Description: "Test Resource used for unit testing 2",
					UserData1:   "update data1",
					UserData2:   "update data2",
				},
				Spec: hpaModel.HpaResourceRequirementSpec{
					Allocatable: &allocatable_true,
					Resource:    hpaModel.HpaResourceRequirementDetails{hpaModel.AllocatableResources{Name: "cpu", Requests: 1, Limits: 1}, hpaModel.NonAllocatableResources{}},
				},
			},
			ResourceClient: &mockIntentManager{
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
				//Items that will be returned by the mocked Client
				ResourceItems: []hpaModel.HpaResourceRequirement{
					{
						MetaData: mtypes.Metadata{
							Name:        "testResource",
							Description: "Test Resource used for unit testing 2",
							UserData1:   "update data1",
							UserData2:   "update data2",
						},
						Spec: hpaModel.HpaResourceRequirementSpec{
							Allocatable: &allocatable_true,
							Resource:    hpaModel.HpaResourceRequirementDetails{hpaModel.AllocatableResources{Name: "cpu", Requests: 1, Limits: 1}, hpaModel.NonAllocatableResources{}},
						},
					},
				},
			},
		},
		{
			label:        "Update Resource when requests < limits",
			name:         "testResource",
			expectedCode: http.StatusOK,
			reader: bytes.NewBuffer([]byte(`{
				"metadata" : {
					"name": "testResource",
    				"description": "Test Resource used for unit testing",
    				"userData1": "update data1",
    				"userData2": "update data2"
				},
				"spec" : {
					"allocatable":          true,
					"resource" : {"name":"cpu", "requests":2,"limits":3}
				}
			}`)),
			expected: hpaModel.HpaResourceRequirement{
				MetaData: mtypes.Metadata{
					Name:        "testResource",
					Description: "Test Resource used for unit testing 2",
					UserData1:   "update data1",
					UserData2:   "update data2",
				},
				Spec: hpaModel.HpaResourceRequirementSpec{
					Allocatable: &allocatable_true,
					Resource:    hpaModel.HpaResourceRequirementDetails{hpaModel.AllocatableResources{Name: "cpu", Requests: 1, Limits: 1}, hpaModel.NonAllocatableResources{}},
				},
			},
			ResourceClient: &mockIntentManager{
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
				//Items that will be returned by the mocked Client
				ResourceItems: []hpaModel.HpaResourceRequirement{
					{
						MetaData: mtypes.Metadata{
							Name:        "testResource",
							Description: "Test Resource used for unit testing 2",
							UserData1:   "update data1",
							UserData2:   "update data2",
						},
						Spec: hpaModel.HpaResourceRequirementSpec{
							Allocatable: &allocatable_true,
							Resource:    hpaModel.HpaResourceRequirementDetails{hpaModel.AllocatableResources{Name: "cpu", Requests: 1, Limits: 1}, hpaModel.NonAllocatableResources{}},
						},
					},
				},
			},
		},
		{
			label:        "Update Resource when limits is not specified",
			name:         "testResource",
			expectedCode: http.StatusOK,
			reader: bytes.NewBuffer([]byte(`{
				"metadata" : {
					"name": "testResource",
    				"description": "Test Resource used for unit testing",
    				"userData1": "update data1",
    				"userData2": "update data2"
				},
				"spec" : {
					"allocatable":          true,
					"resource" : {"name":"cpu", "requests":2}
				}
			}`)),
			expected: hpaModel.HpaResourceRequirement{
				MetaData: mtypes.Metadata{
					Name:        "testResource",
					Description: "Test Resource used for unit testing 2",
					UserData1:   "update data1",
					UserData2:   "update data2",
				},
				Spec: hpaModel.HpaResourceRequirementSpec{
					Allocatable: &allocatable_true,
					Resource:    hpaModel.HpaResourceRequirementDetails{hpaModel.AllocatableResources{Name: "cpu", Requests: 1, Limits: 1}, hpaModel.NonAllocatableResources{}},
				},
			},
			ResourceClient: &mockIntentManager{
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
				//Items that will be returned by the mocked Client
				ResourceItems: []hpaModel.HpaResourceRequirement{
					{
						MetaData: mtypes.Metadata{
							Name:        "testResource",
							Description: "Test Resource used for unit testing 2",
							UserData1:   "update data1",
							UserData2:   "update data2",
						},
						Spec: hpaModel.HpaResourceRequirementSpec{
							Allocatable: &allocatable_true,
							Resource:    hpaModel.HpaResourceRequirementDetails{hpaModel.AllocatableResources{Name: "cpu", Requests: 1, Limits: 1}, hpaModel.NonAllocatableResources{}},
						},
					},
				},
			},
		},
		{
			label:        "Update Resource Failed due to allocatable Resource requests > limits",
			name:         "testResource",
			expectedCode: http.StatusBadRequest,
			reader: bytes.NewBuffer([]byte(`{
				"metadata" : {
					"name": "testResource",
    				"description": "Test Resource used for unit testing",
    				"userData1": "update data1",
    				"userData2": "update data2"
				},
				"spec" : {
					"allocatable":          true,
					"resource" : {"name":"cpu", "requests":3,"limits":2}
				}
			}`)),
			ResourceClient: &mockIntentManager{
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
				//Items that will be returned by the mocked Client
				ResourceItems: []hpaModel.HpaResourceRequirement{
					{
						MetaData: mtypes.Metadata{
							Name:        "testResource",
							Description: "Test Resource used for unit testing 2",
							UserData1:   "update data1",
							UserData2:   "update data2",
						},
						Spec: hpaModel.HpaResourceRequirementSpec{
							Allocatable: &allocatable_true,
							Resource:    hpaModel.HpaResourceRequirementDetails{hpaModel.AllocatableResources{Name: "cpu", Requests: 1, Limits: 1}, hpaModel.NonAllocatableResources{}},
						},
					},
				},
			},
		},
		{
			label:        "Failed Update Resource due to resouce-name mismatch",
			name:         "testResource",
			expectedCode: http.StatusBadRequest,
			reader: bytes.NewBuffer([]byte(`{
				"metadata" : {
					"name": "testResource1",
    				"description": "Test Resource used for unit testing",
    				"userData1": "update data1",
    				"userData2": "update data2"
				},
				"spec" : {
					"allocatable":          true,
					"resource" : {"name":"cpu", "requests":1,"limits":1}
				}
			}`)),
			ResourceClient: &mockIntentManager{
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
			label:        "Failed Update Resource due to not found error",
			name:         "testResource",
			expectedCode: http.StatusNotFound,
			reader: bytes.NewBuffer([]byte(`{
				"metadata" : {
					"name": "testResource",
    				"description": "Test Resource used for unit testing",
    				"userData1": "update data1",
    				"userData2": "update data2"
				},
				"spec" : {
					"allocatable":          true,
					"resource" : {"name":"cpu", "requests":1,"limits":1}
				}
			}`)),
			ResourceClient: &mockIntentManager{
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
				//Items that will be returned by the mocked Client
				ResourceItems: []hpaModel.HpaResourceRequirement{
					{
						MetaData: mtypes.Metadata{
							Name:        "testResource",
							Description: "Test Resource used for unit testing 2",
							UserData1:   "update data1",
							UserData2:   "update data2",
						},
						Spec: hpaModel.HpaResourceRequirementSpec{
							Allocatable: &allocatable_true,
							Resource:    hpaModel.HpaResourceRequirementDetails{hpaModel.AllocatableResources{Name: "cpu", Requests: 1, Limits: 1}, hpaModel.NonAllocatableResources{}},
						},
					},
				},
				Err: pkgerrors.New("internal"),
			},
		},
		{
			label:        "Failed Update Resource due to not found error",
			name:         "testResource",
			expectedCode: http.StatusNotFound,
			reader: bytes.NewBuffer([]byte(`{
				"metadata" : {
					"name": "testResource",
    				"description": "Test Resource used for unit testing",
    				"userData1": "update data1",
    				"userData2": "update data2"
				},
				"spec" : {
					"allocatable":          true,
					"resource" : {"name":"cpu", "requests":1,"limits":1}
				}
			}`)),
			ResourceClient: &mockIntentManager{
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
				//Items that will be returned by the mocked Client
				ResourceItems: []hpaModel.HpaResourceRequirement{
					{
						MetaData: mtypes.Metadata{
							Name:        "testResource",
							Description: "Test Resource used for unit testing 2",
							UserData1:   "update data1",
							UserData2:   "update data2",
						},
						Spec: hpaModel.HpaResourceRequirementSpec{
							Allocatable: &allocatable_true,
							Resource:    hpaModel.HpaResourceRequirementDetails{hpaModel.AllocatableResources{Name: "cpu", Requests: 1, Limits: 1}, hpaModel.NonAllocatableResources{}},
						},
					},
				},
				Err: pkgerrors.New("not found"),
			},
		},
		{
			label:        "Failed Update Resource due to empty req resource-name",
			name:         "testResource",
			expectedCode: http.StatusBadRequest,
			reader: bytes.NewBuffer([]byte(`{
				"metadata" : {
					"name": "",
    				"description": "Test Resource used for unit testing",
    				"userData1": "update data1",
    				"userData2": "update data2"
				},
				"spec" : {
					"allocatable":          true,
					"resource" : {"name":"cpu", "requests":1,"limits":1}
				}
			}`)),
			ResourceClient: &mockIntentManager{
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
	}

	fmt.Printf("\n================== TestResourceUpdateHandler .. total_testcase_count[%d] ==================\n", len(testCases))
	testcasesFailed := make([]string, 0)
	for i, testCase := range testCases {
		fmt.Printf("\n================== TestResourceUpdateHandler .. testcase_count[%d] testcase_name[%s] ==================\n", i, testCase.label)
		t.Run(testCase.label, func(t *testing.T) {
			request := httptest.NewRequest("PUT", "/v2/projects/project1/composite-apps/compositeapp1/v2/deployment-intent-groups/digroup/hpa-intents/hpaintent1/hpa-resource-consumers/consumer1/resource-requirements/"+testCase.name, testCase.reader)
			resp := executeRequest(request, NewRouter(testCase.ResourceClient))

			//Check returned code
			if resp.StatusCode != testCase.expectedCode {
				error := fmt.Sprintf("TestCase Failed[%v] => index[%d] Expected %d; Got: %d", testCase.label, i, testCase.expectedCode, resp.StatusCode)
				testcasesFailed = append(testcasesFailed, error)
				t.Fatalf(error)
			}

			//Check returned body only if statusOK
			if resp.StatusCode == http.StatusOK {
				got := hpaModel.HpaResourceRequirement{}
				json.NewDecoder(resp.Body).Decode(&got)

				if reflect.DeepEqual(testCase.expected, got) == false {
					t.Errorf("updateHandler returned unexpected body: got %v;"+
						" expected %v", got, testCase.expected)
				}
			}
		})
	}
	for _, testCaseFailed := range testcasesFailed {
		fmt.Printf("\n================== TestResourceUpdateHandler .. testcase-failed[%v] ==================\n", testCaseFailed)
	}
}

func TestResourceGetHandler(t *testing.T) {

	testCases := []struct {
		label          string
		expected       hpaModel.HpaResourceRequirement
		name, version  string
		expectedCode   int
		ResourceClient *mockIntentManager
	}{
		{
			label:        "Get non-allocatable Resource",
			expectedCode: http.StatusOK,
			expected: hpaModel.HpaResourceRequirement{
				MetaData: mtypes.Metadata{
					Name:        "testResource",
					Description: "Test Resource used for unit testing",
					UserData1:   "data1",
					UserData2:   "data2",
				},
				Spec: hpaModel.HpaResourceRequirementSpec{
					Allocatable: &allocatable_false,
					Resource:    hpaModel.HpaResourceRequirementDetails{hpaModel.AllocatableResources{}, hpaModel.NonAllocatableResources{Key: "vpu", Value: "yes"}},
				},
			},
			name: "testResource",
			ResourceClient: &mockIntentManager{
				ResourceItems: []hpaModel.HpaResourceRequirement{
					{
						MetaData: mtypes.Metadata{
							Name:        "testResource",
							Description: "Test Resource used for unit testing",
							UserData1:   "data1",
							UserData2:   "data2",
						},
						Spec: hpaModel.HpaResourceRequirementSpec{
							Allocatable: &allocatable_false,
							Resource:    hpaModel.HpaResourceRequirementDetails{hpaModel.AllocatableResources{}, hpaModel.NonAllocatableResources{Key: "vpu", Value: "yes"}},
						},
					},
				},
			},
		},
		{
			label:        "Get allocatable Resource",
			expectedCode: http.StatusOK,
			expected: hpaModel.HpaResourceRequirement{
				MetaData: mtypes.Metadata{
					Name:        "testResource",
					Description: "Test Resource used for unit testing",
					UserData1:   "data1",
					UserData2:   "data2",
				},
				Spec: hpaModel.HpaResourceRequirementSpec{
					Allocatable: &allocatable_true,
					Resource:    hpaModel.HpaResourceRequirementDetails{hpaModel.AllocatableResources{Name: "cpu", Requests: 1, Limits: 1}, hpaModel.NonAllocatableResources{}},
				},
			},
			name: "testResource",
			ResourceClient: &mockIntentManager{
				ResourceItems: []hpaModel.HpaResourceRequirement{
					{
						MetaData: mtypes.Metadata{
							Name:        "testResource",
							Description: "Test Resource used for unit testing",
							UserData1:   "data1",
							UserData2:   "data2",
						},
						Spec: hpaModel.HpaResourceRequirementSpec{
							Allocatable: &allocatable_true,
							Resource:    hpaModel.HpaResourceRequirementDetails{hpaModel.AllocatableResources{Name: "cpu", Requests: 1, Limits: 1}, hpaModel.NonAllocatableResources{}},
						},
					},
				},
			},
		},
		{
			label:        "Get Non-Exiting Resource",
			expectedCode: http.StatusNotFound,
			name:         "nonexistingResource",
			ResourceClient: &mockIntentManager{
				ResourceItems: []hpaModel.HpaResourceRequirement{},
				Err:           pkgerrors.New("Internal Error"),
			},
		},
	}

	fmt.Printf("\n================== TestResourceGetHandler .. total_testcase_count[%d] ==================\n", len(testCases))
	testcasesFailed := make([]string, 0)
	for i, testCase := range testCases {
		fmt.Printf("\n================== TestResourceGetHandler .. testcase_count[%d] testcase_name[%s] ==================\n", i, testCase.label)
		t.Run(testCase.label, func(t *testing.T) {
			request := httptest.NewRequest("GET", "/v2/projects/project1/composite-apps/compositeapp1/v2/deployment-intent-groups/digroup/hpa-intents/hpaintent1/hpa-resource-consumers/consumer1/resource-requirements/"+testCase.name, nil)
			resp := executeRequest(request, NewRouter(testCase.ResourceClient))

			//Check returned code
			if resp.StatusCode != testCase.expectedCode {
				error := fmt.Sprintf("TestCase Failed[%v] => index[%d] Expected %d; Got: %d", testCase.label, i, testCase.expectedCode, resp.StatusCode)
				testcasesFailed = append(testcasesFailed, error)
				t.Fatalf(error)
			}

			//Check returned body only if statusOK
			if resp.StatusCode == http.StatusOK {
				got := hpaModel.HpaResourceRequirement{}
				json.NewDecoder(resp.Body).Decode(&got)

				if reflect.DeepEqual(testCase.expected, got) == false {
					t.Errorf("TestResourceGetHandler returned unexpected body: got %v;"+
						" expected %v", got, testCase.expected)
				}
			}
		})
	}
	for _, testCaseFailed := range testcasesFailed {
		fmt.Printf("\n================== TestResourceGetHandler .. testcase-failed[%v] ==================\n", testCaseFailed)
	}
}

func TestResourceGetHandlerByName(t *testing.T) {

	testCases := []struct {
		label          string
		expected       hpaModel.HpaResourceRequirement
		name, version  string
		expectedCode   int
		ResourceClient *mockIntentManager
	}{
		{
			label:        "GetResourceByName non-allocatable Resource",
			expectedCode: http.StatusOK,
			expected: hpaModel.HpaResourceRequirement{
				MetaData: mtypes.Metadata{
					Name:        "testResource",
					Description: "Test Resource used for unit testing",
					UserData1:   "data1",
					UserData2:   "data2",
				},
				Spec: hpaModel.HpaResourceRequirementSpec{
					Allocatable: &allocatable_false,
					Resource:    hpaModel.HpaResourceRequirementDetails{hpaModel.AllocatableResources{}, hpaModel.NonAllocatableResources{Key: "vpu", Value: "yes"}},
				},
			},
			name: "testResource",
			ResourceClient: &mockIntentManager{
				ResourceItems: []hpaModel.HpaResourceRequirement{
					{
						MetaData: mtypes.Metadata{
							Name:        "testResource",
							Description: "Test Resource used for unit testing",
							UserData1:   "data1",
							UserData2:   "data2",
						},
						Spec: hpaModel.HpaResourceRequirementSpec{
							Allocatable: &allocatable_false,
							Resource:    hpaModel.HpaResourceRequirementDetails{hpaModel.AllocatableResources{}, hpaModel.NonAllocatableResources{Key: "vpu", Value: "yes"}},
						},
					},
				},
			},
		},
		{
			label:        "GetResourceByName allocatable Resource",
			expectedCode: http.StatusOK,
			expected: hpaModel.HpaResourceRequirement{
				MetaData: mtypes.Metadata{
					Name:        "testResource",
					Description: "Test Resource used for unit testing",
					UserData1:   "data1",
					UserData2:   "data2",
				},
				Spec: hpaModel.HpaResourceRequirementSpec{
					Allocatable: &allocatable_true,
					Resource:    hpaModel.HpaResourceRequirementDetails{hpaModel.AllocatableResources{Name: "cpu", Requests: 1, Limits: 1}, hpaModel.NonAllocatableResources{}},
				},
			},
			name: "testResource",
			ResourceClient: &mockIntentManager{
				ResourceItems: []hpaModel.HpaResourceRequirement{
					{
						MetaData: mtypes.Metadata{
							Name:        "testResource",
							Description: "Test Resource used for unit testing",
							UserData1:   "data1",
							UserData2:   "data2",
						},
						Spec: hpaModel.HpaResourceRequirementSpec{
							Allocatable: &allocatable_true,
							Resource:    hpaModel.HpaResourceRequirementDetails{hpaModel.AllocatableResources{Name: "cpu", Requests: 1, Limits: 1}, hpaModel.NonAllocatableResources{}},
						},
					},
				},
			},
		},
		{
			label:        "GetResourceByName Non-Exiting Resource",
			expectedCode: http.StatusNotFound,
			name:         "nonexistingResource",
			ResourceClient: &mockIntentManager{
				ResourceItemsSpec: []hpaModel.HpaResourceRequirementSpec{},
				Err:               pkgerrors.New("Internal Error"),
			},
		},
		{
			label:        "GetResourceByName Non-Exiting empty Resource",
			expectedCode: http.StatusBadRequest,
			name:         "",
			ResourceClient: &mockIntentManager{
				ResourceItemsSpec: []hpaModel.HpaResourceRequirementSpec{},
				Err:               pkgerrors.New("Internal Error"),
			},
		},
	}

	fmt.Printf("\n================== TestResourceGetHandlerByName .. total_testcase_count[%d] ==================\n", len(testCases))
	testcasesFailed := make([]string, 0)
	for i, testCase := range testCases {
		fmt.Printf("\n================== TestResourceGetHandlerByName .. testcase_count[%d] testcase_name[%s] ==================\n", i, testCase.label)
		t.Run(testCase.label, func(t *testing.T) {
			request := httptest.NewRequest("QUERY", "/v2/projects/project1/composite-apps/compositeapp1/v2/deployment-intent-groups/digroup/hpa-intents/hpaintent1/hpa-resource-consumers/consumer1/resource-requirements?resource="+testCase.name, nil)
			resp := executeRequest(request, NewRouter(testCase.ResourceClient))

			//Check returned code
			if resp.StatusCode != testCase.expectedCode {
				error := fmt.Sprintf("TestCase Failed[%v] => index[%d] Expected %d; Got: %d", testCase.label, i, testCase.expectedCode, resp.StatusCode)
				testcasesFailed = append(testcasesFailed, error)
				t.Fatalf(error)
			}

			//Check returned body only if statusOK
			if resp.StatusCode == http.StatusOK {
				got := hpaModel.HpaResourceRequirement{}
				json.NewDecoder(resp.Body).Decode(&got)

				if reflect.DeepEqual(testCase.expected, got) == false {
					t.Errorf("TestResourceGetHandlerByName returned unexpected body: got %v;"+
						" expected %v", got, testCase.expected)
				}
			}
		})
	}
	for _, testCaseFailed := range testcasesFailed {
		fmt.Printf("\n================== TestResourceGetHandlerByName .. testcase-failed[%v] ==================\n", testCaseFailed)
	}
}

func TestResourceGetAllHandler(t *testing.T) {

	testCases := []struct {
		label          string
		expected       []hpaModel.HpaResourceRequirement
		name, version  string
		expectedCode   int
		ResourceClient *mockIntentManager
	}{
		{
			label:        "Get non-allocatable Resource",
			expectedCode: http.StatusOK,
			expected: []hpaModel.HpaResourceRequirement{
				{
					MetaData: mtypes.Metadata{
						Name:        "testResource",
						Description: "Test Resource used for unit testing",
						UserData1:   "data1",
						UserData2:   "data2",
					},
					Spec: hpaModel.HpaResourceRequirementSpec{
						Allocatable: &allocatable_false,
						Resource:    hpaModel.HpaResourceRequirementDetails{hpaModel.AllocatableResources{}, hpaModel.NonAllocatableResources{Key: "vpu", Value: "yes"}},
					},
				},
			},
			ResourceClient: &mockIntentManager{
				ResourceItems: []hpaModel.HpaResourceRequirement{
					{
						MetaData: mtypes.Metadata{
							Name:        "testResource",
							Description: "Test Resource used for unit testing",
							UserData1:   "data1",
							UserData2:   "data2",
						},
						Spec: hpaModel.HpaResourceRequirementSpec{
							Allocatable: &allocatable_false,
							Resource:    hpaModel.HpaResourceRequirementDetails{hpaModel.AllocatableResources{}, hpaModel.NonAllocatableResources{Key: "vpu", Value: "yes"}},
						},
					},
				},
			},
		},
		{
			label:        "Get allocatable Resource",
			expectedCode: http.StatusOK,
			expected: []hpaModel.HpaResourceRequirement{
				{
					MetaData: mtypes.Metadata{
						Name:        "testResource",
						Description: "Test Resource used for unit testing",
						UserData1:   "data1",
						UserData2:   "data2",
					},
					Spec: hpaModel.HpaResourceRequirementSpec{
						Allocatable: &allocatable_true,
						Resource:    hpaModel.HpaResourceRequirementDetails{hpaModel.AllocatableResources{Name: "cpu", Requests: 1, Limits: 1}, hpaModel.NonAllocatableResources{}},
					},
				},
			},
			name: "testResource",
			ResourceClient: &mockIntentManager{
				ResourceItems: []hpaModel.HpaResourceRequirement{
					{
						MetaData: mtypes.Metadata{
							Name:        "testResource",
							Description: "Test Resource used for unit testing",
							UserData1:   "data1",
							UserData2:   "data2",
						},
						Spec: hpaModel.HpaResourceRequirementSpec{
							Allocatable: &allocatable_true,
							Resource:    hpaModel.HpaResourceRequirementDetails{hpaModel.AllocatableResources{Name: "cpu", Requests: 1, Limits: 1}, hpaModel.NonAllocatableResources{}},
						},
					},
				},
			},
		},
		{
			label:        "Get Non-Exiting Resource",
			expectedCode: http.StatusNotFound,
			name:         "nonexistingResource",
			ResourceClient: &mockIntentManager{
				ResourceItems: []hpaModel.HpaResourceRequirement{},
				Err:           pkgerrors.New("Internal Error"),
			},
		},
		{
			label:        "GetAll No Resources",
			expectedCode: http.StatusOK,
			ResourceClient: &mockIntentManager{
				ResourceItems: []hpaModel.HpaResourceRequirement{},
			},
			expected: []hpaModel.HpaResourceRequirement{},
		},
	}

	fmt.Printf("\n================== TestResourceGetAllHandler .. total_testcase_count[%d] ==================\n", len(testCases))
	for i, testCase := range testCases {
		fmt.Printf("\n================== TestResourceGetAllHandler .. testcase_count[%d] testcase_name[%s] ==================\n", i, testCase.label)
		t.Run(testCase.label, func(t *testing.T) {
			request := httptest.NewRequest("GET", "/v2/projects/project1/composite-apps/compositeapp1/v2/deployment-intent-groups/digroup/hpa-intents/hpaintent1/hpa-resource-consumers/consumer1/resource-requirements", nil)
			resp := executeRequest(request, NewRouter(testCase.ResourceClient))

			//Check returned code
			if resp.StatusCode != testCase.expectedCode {
				t.Fatalf("Expected %d; Got: %d", testCase.expectedCode, resp.StatusCode)
			}

			//Check returned body only if statusOK
			if resp.StatusCode == http.StatusOK {
				got := []hpaModel.HpaResourceRequirement{}
				json.NewDecoder(resp.Body).Decode(&got)

				if reflect.DeepEqual(testCase.expected, got) == false {
					t.Errorf("TestResourceGetAllHandler returned unexpected body: got %v;"+
						" expected %v", got, testCase.expected)
				}
			}
		})
	}
}

func TestResourceDeleteHandler(t *testing.T) {

	testCases := []struct {
		label          string
		name           string
		version        string
		expectedCode   int
		ResourceClient *mockIntentManager
	}{
		{
			label:        "Delete Resource",
			expectedCode: http.StatusNoContent,
			name:         "testResource",
			ResourceClient: &mockIntentManager{
				//Items that will be returned by the mocked Client
				ResourceItems: []hpaModel.HpaResourceRequirement{
					{
						MetaData: mtypes.Metadata{
							Name:        "testResource",
							Description: "Test Resource used for unit testing",
							UserData1:   "data1",
							UserData2:   "data2",
						},
					},
				},
			},
		},
		{
			label:        "Delete Non-Exiting Resource",
			expectedCode: http.StatusNotFound,
			name:         "testResource",
			ResourceClient: &mockIntentManager{
				Err: pkgerrors.New("Internal Error"),
			},
		},
		{
			label:        "Delete Non-Exiting empty Resource",
			expectedCode: http.StatusNotFound,
			name:         "",
			ResourceClient: &mockIntentManager{
				Err: pkgerrors.New("Internal Error"),
			},
		},
	}

	fmt.Printf("\n================== TestResourceDeleteHandler .. total_testcase_count[%d] ==================\n", len(testCases))
	testcasesFailed := make([]string, 0)
	for i, testCase := range testCases {
		fmt.Printf("\n================== TestResourceDeleteHandler .. testcase_count[%d] testcase_name[%s] ==================\n", i, testCase.label)
		t.Run(testCase.label, func(t *testing.T) {
			request := httptest.NewRequest("DELETE", "/v2/projects/project1/composite-apps/compositeapp1/v2/deployment-intent-groups/digroup/hpa-intents/hpaintent1/hpa-resource-consumers/consumer1/resource-requirements/"+testCase.name, nil)
			resp := executeRequest(request, NewRouter(testCase.ResourceClient))

			//Check returned code
			if resp.StatusCode != testCase.expectedCode {
				error := fmt.Sprintf("TestCase Failed[%v] => index[%d] Expected %d; Got: %d", testCase.label, i, testCase.expectedCode, resp.StatusCode)
				testcasesFailed = append(testcasesFailed, error)
				t.Fatalf(error)
			}

		})
	}
	for _, testCaseFailed := range testcasesFailed {
		fmt.Printf("\n================== TestResourceDeleteHandler .. testcase-failed[%v] ==================\n", testCaseFailed)
	}
}

func TestResourceDeleteAllHandler(t *testing.T) {

	testCases := []struct {
		label          string
		name           string
		version        string
		expectedCode   int
		ResourceClient *mockIntentManager
	}{
		{
			label:        "Delete Resource",
			expectedCode: http.StatusNoContent,
			ResourceClient: &mockIntentManager{
				//Items that will be returned by the mocked Client
				ResourceItems: []hpaModel.HpaResourceRequirement{
					{
						MetaData: mtypes.Metadata{
							Name:        "testResource",
							Description: "Test Resource used for unit testing",
							UserData1:   "data1",
							UserData2:   "data2",
						},
					},
				},
			},
		},
		{
			label:        "Delete Non-Exiting Resource",
			expectedCode: http.StatusNotFound,
			ResourceClient: &mockIntentManager{
				Err: pkgerrors.New("Internal Error"),
			},
		},
	}

	fmt.Printf("\n================== TestResourceDeleteAllHandler .. total_testcase_count[%d] ==================\n", len(testCases))
	for i, testCase := range testCases {
		fmt.Printf("\n================== TestResourceDeleteAllHandler .. testcase_count[%d] testcase_name[%s] ==================\n", i, testCase.label)
		t.Run(testCase.label, func(t *testing.T) {
			request := httptest.NewRequest("DELETE", "/v2/projects/project1/composite-apps/compositeapp1/v2/deployment-intent-groups/digroup/hpa-intents/hpaintent1/hpa-resource-consumers/consumer1/resource-requirements", nil)
			resp := executeRequest(request, NewRouter(testCase.ResourceClient))

			//Check returned code
			if resp.StatusCode != testCase.expectedCode {
				t.Fatalf("Expected %d; Got: %d", testCase.expectedCode, resp.StatusCode)
			}
		})
	}
}
