// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2020 Intel Corporation

package module

import (
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/open-ness/EMCO/src/orchestrator/pkg/infra/db"
	orchMod "github.com/open-ness/EMCO/src/orchestrator/pkg/module"

	hpaModel "github.com/open-ness/EMCO/src/hpa-plc/pkg/model"
	mtypes "github.com/open-ness/EMCO/src/orchestrator/pkg/module/types"
	pkgerrors "github.com/pkg/errors"
)

func TestCreateConsumer(t *testing.T) {
	testCases := []struct {
		label         string
		inp           hpaModel.HpaResourceConsumer
		expectedError string
		mockdb        *db.MockDB
		expected      hpaModel.HpaResourceConsumer
	}{
		{
			label: "Create Consumer",
			inp: hpaModel.HpaResourceConsumer{
				MetaData: mtypes.Metadata{
					Name:        "consumer1",
					Description: "A sample Consumer used for unit testing",
					UserData1:   "data1",
					UserData2:   "data2",
				},
				Spec: hpaModel.HpaResourceConsumerSpec{
					Name:          "deployment-1",
					ContainerName: "container-1"},
			},
			expected: hpaModel.HpaResourceConsumer{
				MetaData: mtypes.Metadata{
					Name:        "consumer1",
					Description: "A sample Consumer used for unit testing",
					UserData1:   "data1",
					UserData2:   "data2",
				},
				Spec: hpaModel.HpaResourceConsumerSpec{
					Name:          "deployment-1",
					ContainerName: "container-1"},
			},
			expectedError: "",
			mockdb: &db.MockDB{
				Items: []map[string]map[string][]byte{
					{
						orchMod.ProjectKey{ProjectName: "project1"}.String(): {
							"projectmetadata": []byte(
								"{\"project-name\":\"project1\"," +
									"\"description\":\"Test project for unit testing\"}"),
						},
						orchMod.CompositeAppKey{CompositeAppName: "compositeapp1",
							Version: "version1", Project: "project1"}.String(): {
							"compositeappmetadata": []byte(
								"{\"metadata\":{" +
									"\"name\":\"compositeapp1\"," +
									"\"description\":\"description\"," +
									"\"userData1\":\"user data\"," +
									"\"userData2\":\"user data\"" +
									"}," +
									"\"spec\":{" +
									"\"version\":\"version of the composite app\"}}"),
						},
						orchMod.DeploymentIntentGroupKey{
							Name:         "dgroup1",
							Project:      "project1",
							CompositeApp: "compositeapp1",
							Version:      "version1",
						}.String(): {
							"deploymentintentgroupmetadata": []byte(
								"{\"metadata\":{\"name\":\"dgroup1\"," +
									"\"description\":\"DescriptionTestDeploymentIntentGroup\"," +
									"\"userData1\": \"userData1\"," +
									"\"userData2\": \"userData2\"}," +
									"\"spec\":{\"profile\": \"Testprofile\"," +
									"\"version\": \"version of deployment\"," +
									"\"override-values\":[" +
									"{" +
									"\"app-name\": \"TestAppName\"," +
									"\"values\": " +
									"{" +
									"\"imageRepository\":\"registry.hub.docker.com\"" +
									"}" +
									"}," +
									"{" +
									"\"app-name\": \"TestAppName\"," +
									"\"values\": " +
									"{" +
									"\"imageRepository\":\"registry.hub.docker.com\"" +
									"}" +
									"}" +
									"]," +
									"\"logical-cloud\": \"cloud1\"" +
									"}" +
									"}"),
						},
						HpaIntentKey{IntentName: "intent1",
							Project: "project1", CompositeApp: "compositeapp1",
							Version: "version1", DeploymentIntentGroup: "dgroup1"}.String(): {
							"HpaPlacementControllerMetadata": []byte(
								"{" +
									"\"metadata\" : {" +
									"\"Name\":\"intent1\"," +
									"\"Description\":\"Test Intent for unit testing\"," +
									"\"UserData1\": \"userData1\"," +
									"\"UserData2\":\"userData2\"}" +
									"}"),
						},
					},
				},
			},
		},
		{
			label: "Create Consumer with replicas",
			inp: hpaModel.HpaResourceConsumer{
				MetaData: mtypes.Metadata{
					Name:        "consumer1",
					Description: "A sample Consumer used for unit testing",
					UserData1:   "data1",
					UserData2:   "data2",
				},
				Spec: hpaModel.HpaResourceConsumerSpec{
					Replicas:      2,
					Name:          "deployment-1",
					ContainerName: "container-1"},
			},
			expected: hpaModel.HpaResourceConsumer{
				MetaData: mtypes.Metadata{
					Name:        "consumer1",
					Description: "A sample Consumer used for unit testing",
					UserData1:   "data1",
					UserData2:   "data2",
				},
				Spec: hpaModel.HpaResourceConsumerSpec{
					Replicas:      2,
					Name:          "deployment-1",
					ContainerName: "container-1"},
			},
			expectedError: "",
			mockdb: &db.MockDB{
				Items: []map[string]map[string][]byte{
					{
						orchMod.ProjectKey{ProjectName: "project1"}.String(): {
							"projectmetadata": []byte(
								"{\"project-name\":\"project1\"," +
									"\"description\":\"Test project for unit testing\"}"),
						},
						orchMod.CompositeAppKey{CompositeAppName: "compositeapp1",
							Version: "version1", Project: "project1"}.String(): {
							"compositeappmetadata": []byte(
								"{\"metadata\":{" +
									"\"name\":\"compositeapp1\"," +
									"\"description\":\"description\"," +
									"\"userData1\":\"user data\"," +
									"\"userData2\":\"user data\"" +
									"}," +
									"\"spec\":{" +
									"\"version\":\"version of the composite app\"}}"),
						},
						orchMod.DeploymentIntentGroupKey{
							Name:         "dgroup1",
							Project:      "project1",
							CompositeApp: "compositeapp1",
							Version:      "version1",
						}.String(): {
							"deploymentintentgroupmetadata": []byte(
								"{\"metadata\":{\"name\":\"dgroup1\"," +
									"\"description\":\"DescriptionTestDeploymentIntentGroup\"," +
									"\"userData1\": \"userData1\"," +
									"\"userData2\": \"userData2\"}," +
									"\"spec\":{\"profile\": \"Testprofile\"," +
									"\"version\": \"version of deployment\"," +
									"\"override-values\":[" +
									"{" +
									"\"app-name\": \"TestAppName\"," +
									"\"values\": " +
									"{" +
									"\"imageRepository\":\"registry.hub.docker.com\"" +
									"}" +
									"}," +
									"{" +
									"\"app-name\": \"TestAppName\"," +
									"\"values\": " +
									"{" +
									"\"imageRepository\":\"registry.hub.docker.com\"" +
									"}" +
									"}" +
									"]," +
									"\"logical-cloud\": \"cloud1\"" +
									"}" +
									"}"),
						},
						HpaIntentKey{IntentName: "intent1",
							Project: "project1", CompositeApp: "compositeapp1",
							Version: "version1", DeploymentIntentGroup: "dgroup1"}.String(): {
							"HpaPlacementControllerMetadata": []byte(
								"{" +
									"\"metadata\" : {" +
									"\"Name\":\"intent1\"," +
									"\"Description\":\"Test Intent for unit testing\"," +
									"\"UserData1\": \"userData1\"," +
									"\"UserData2\":\"userData2\"}" +
									"}"),
						},
					},
				},
			},
		},
		{
			label:         "Failed Create Consumer",
			expectedError: "dependency not found",
			mockdb: &db.MockDB{
				Err: pkgerrors.New("Error Creating Consumer"),
			},
		},
		{
			label: "Create Existing Consumer",
			inp: hpaModel.HpaResourceConsumer{
				MetaData: mtypes.Metadata{
					Name:        "consumer1",
					Description: "A sample Consumer used for unit testing",
					UserData1:   "data1",
					UserData2:   "data2",
				},
			},
			expectedError: "Consumer already exists",
			mockdb: &db.MockDB{
				Items: []map[string]map[string][]byte{
					{
						orchMod.ProjectKey{ProjectName: "project1"}.String(): {
							"projectmetadata": []byte(
								"{\"project-name\":\"project1\"," +
									"\"description\":\"Test project for unit testing\"}"),
						},
						orchMod.CompositeAppKey{CompositeAppName: "compositeapp1",
							Version: "version1", Project: "project1"}.String(): {
							"compositeappmetadata": []byte(
								"{\"metadata\":{" +
									"\"name\":\"compositeapp1\"," +
									"\"description\":\"description\"," +
									"\"userData1\":\"user data\"," +
									"\"userData2\":\"user data\"" +
									"}," +
									"\"spec\":{" +
									"\"version\":\"version of the composite app\"}}"),
						},
						orchMod.DeploymentIntentGroupKey{
							Name:         "dgroup1",
							Project:      "project1",
							CompositeApp: "compositeapp1",
							Version:      "version1",
						}.String(): {
							"deploymentintentgroupmetadata": []byte(
								"{\"metadata\":{\"name\":\"dgroup1\"," +
									"\"description\":\"DescriptionTestDeploymentIntentGroup\"," +
									"\"userData1\": \"userData1\"," +
									"\"userData2\": \"userData2\"}," +
									"\"spec\":{\"profile\": \"Testprofile\"," +
									"\"version\": \"version of deployment\"," +
									"\"override-values\":[" +
									"{" +
									"\"app-name\": \"TestAppName\"," +
									"\"values\": " +
									"{" +
									"\"imageRepository\":\"registry.hub.docker.com\"" +
									"}" +
									"}," +
									"{" +
									"\"app-name\": \"TestAppName\"," +
									"\"values\": " +
									"{" +
									"\"imageRepository\":\"registry.hub.docker.com\"" +
									"}" +
									"}" +
									"]," +
									"\"logical-cloud\": \"cloud1\"" +
									"}" +
									"}"),
						},
						HpaIntentKey{IntentName: "intent1",
							Project: "project1", CompositeApp: "compositeapp1",
							Version: "version1", DeploymentIntentGroup: "dgroup1"}.String(): {
							"HpaPlacementControllerMetadata": []byte(
								"{" +
									"\"metadata\" : {" +
									"\"Name\":\"intent1\"," +
									"\"Description\":\"Test Intent for unit testing\"," +
									"\"UserData1\": \"userData1\"," +
									"\"UserData2\":\"userData2\"}" +
									"}"),
						},
						HpaConsumerKey{ConsumerName: "consumer1", IntentName: "intent1",
							Project: "project1", CompositeApp: "compositeapp1",
							Version: "version1", DeploymentIntentGroup: "dgroup1"}.String(): {
							"HpaPlacementControllerMetadata": []byte(
								"{" +
									"\"metadata\" : {" +
									"\"Name\":\"consumer1\"," +
									"\"Description\":\"Test Consumer for unit testing\"," +
									"\"UserData1\":\"userData1\"," +
									"\"UserData2\":\"userData2\"}" +
									"}"),
						},
					},
				},
			},
		},
	}

	fmt.Printf("\n================== TestCreateConsumer .. total_testcase_count[%d] ==================\n", len(testCases))
	for i, testCase := range testCases {
		t.Run(testCase.label, func(t *testing.T) {
			fmt.Printf("\n================== TestCreateConsumer .. testcase_count[%d] testcase_name[%s] ==================\n", i, testCase.label)
			db.DBconn = testCase.mockdb
			impl := NewHpaPlacementClient()
			got, err := impl.AddConsumer(testCase.inp, "project1", "compositeapp1", "version1", "dgroup1", "intent1", false)
			if err != nil {
				if testCase.expectedError == "" {
					t.Fatalf("Create returned an unexpected-error[%s] expected[%s]", err, testCase.expectedError)
				}
				if strings.Contains(strings.ToLower(err.Error()), strings.ToLower(testCase.expectedError)) == false {
					t.Fatalf("Create returned an unexpected-error[%s] expected[%s]", err, testCase.expectedError)
				}
			} else {
				if reflect.DeepEqual(testCase.expected, got) == false {
					t.Errorf("Create returned unexpected body: got %v;"+
						" expected %v", got, testCase.expected)
				}
			}
		})
	}
}

func TestUpdateConsumer(t *testing.T) {
	testCases := []struct {
		label         string
		inp           hpaModel.HpaResourceConsumer
		expectedError string
		mockdb        *db.MockDB
		expected      hpaModel.HpaResourceConsumer
	}{
		{
			label: "Update Consumer",
			inp: hpaModel.HpaResourceConsumer{
				MetaData: mtypes.Metadata{
					Name:        "consumer1",
					Description: "Test Consumer for unit testing",
					UserData1:   "update userData1",
					UserData2:   "update userData2",
				},
			},
			expected: hpaModel.HpaResourceConsumer{
				MetaData: mtypes.Metadata{
					Name:        "consumer1",
					Description: "Test Consumer for unit testing",
					UserData1:   "update userData1",
					UserData2:   "update userData2",
				},
			},
			expectedError: "",
			mockdb: &db.MockDB{
				Items: []map[string]map[string][]byte{
					{
						orchMod.ProjectKey{ProjectName: "project1"}.String(): {
							"projectmetadata": []byte(
								"{\"project-name\":\"project1\"," +
									"\"description\":\"Test project for unit testing\"}"),
						},
						orchMod.CompositeAppKey{CompositeAppName: "compositeapp1",
							Version: "version1", Project: "project1"}.String(): {
							"compositeappmetadata": []byte(
								"{\"metadata\":{" +
									"\"name\":\"compositeapp1\"," +
									"\"description\":\"description\"," +
									"\"userData1\":\"user data\"," +
									"\"userData2\":\"user data\"" +
									"}," +
									"\"spec\":{" +
									"\"version\":\"version of the composite app\"}}"),
						},
						orchMod.DeploymentIntentGroupKey{
							Name:         "dgroup1",
							Project:      "project1",
							CompositeApp: "compositeapp1",
							Version:      "version1",
						}.String(): {
							"deploymentintentgroupmetadata": []byte(
								"{\"metadata\":{\"name\":\"dgroup1\"," +
									"\"description\":\"DescriptionTestDeploymentIntentGroup\"," +
									"\"userData1\": \"userData1\"," +
									"\"userData2\": \"userData2\"}," +
									"\"spec\":{\"profile\": \"Testprofile\"," +
									"\"version\": \"version of deployment\"," +
									"\"override-values\":[" +
									"{" +
									"\"app-name\": \"TestAppName\"," +
									"\"values\": " +
									"{" +
									"\"imageRepository\":\"registry.hub.docker.com\"" +
									"}" +
									"}," +
									"{" +
									"\"app-name\": \"TestAppName\"," +
									"\"values\": " +
									"{" +
									"\"imageRepository\":\"registry.hub.docker.com\"" +
									"}" +
									"}" +
									"]," +
									"\"logical-cloud\": \"cloud1\"" +
									"}" +
									"}"),
						},
						HpaIntentKey{IntentName: "intent1",
							Project: "project1", CompositeApp: "compositeapp1",
							Version: "version1", DeploymentIntentGroup: "dgroup1"}.String(): {
							"HpaPlacementControllerMetadata": []byte(
								"{" +
									"\"metadata\" : {" +
									"\"Name\":\"intent1\"," +
									"\"Description\":\"Test Intent for unit testing\"," +
									"\"UserData1\": \"userData1\"," +
									"\"UserData2\":\"userData2\"}" +
									"}"),
						},
						HpaConsumerKey{ConsumerName: "consumer1", IntentName: "intent1",
							Project: "project1", CompositeApp: "compositeapp1",
							Version: "version1", DeploymentIntentGroup: "dgroup1"}.String(): {
							"HpaPlacementControllerMetadata": []byte(
								"{" +
									"\"metadata\" : {" +
									"\"Name\":\"consumer1\"," +
									"\"Description\":\"Test Consumer for unit testing\"," +
									"\"UserData1\":\"userData1\"," +
									"\"UserData2\":\"userData2\"}" +
									"}"),
						},
					},
				},
			},
		},
		{
			label: "Update Consumer non-existing consumer",
			inp: hpaModel.HpaResourceConsumer{
				MetaData: mtypes.Metadata{
					Name:        "consumer1",
					Description: "Test Consumer for unit testing",
					UserData1:   "update userData1",
					UserData2:   "update userData2",
				},
			},
			expectedError: "Error",
			mockdb: &db.MockDB{
				Items: []map[string]map[string][]byte{
					{
						orchMod.ProjectKey{ProjectName: "project1"}.String(): {
							"projectmetadata": []byte(
								"{\"project-name\":\"project1\"," +
									"\"description\":\"Test project for unit testing\"}"),
						},
						orchMod.CompositeAppKey{CompositeAppName: "compositeapp1",
							Version: "version1", Project: "project1"}.String(): {
							"compositeappmetadata": []byte(
								"{\"metadata\":{" +
									"\"name\":\"compositeapp1\"," +
									"\"description\":\"description\"," +
									"\"userData1\":\"user data\"," +
									"\"userData2\":\"user data\"" +
									"}," +
									"\"spec\":{" +
									"\"version\":\"version of the composite app\"}}"),
						},
						orchMod.DeploymentIntentGroupKey{
							Name:         "dgroup1",
							Project:      "project1",
							CompositeApp: "compositeapp1",
							Version:      "version1",
						}.String(): {
							"deploymentintentgroupmetadata": []byte(
								"{\"metadata\":{\"name\":\"dgroup1\"," +
									"\"description\":\"DescriptionTestDeploymentIntentGroup\"," +
									"\"userData1\": \"userData1\"," +
									"\"userData2\": \"userData2\"}," +
									"\"spec\":{\"profile\": \"Testprofile\"," +
									"\"version\": \"version of deployment\"," +
									"\"override-values\":[" +
									"{" +
									"\"app-name\": \"TestAppName\"," +
									"\"values\": " +
									"{" +
									"\"imageRepository\":\"registry.hub.docker.com\"" +
									"}" +
									"}," +
									"{" +
									"\"app-name\": \"TestAppName\"," +
									"\"values\": " +
									"{" +
									"\"imageRepository\":\"registry.hub.docker.com\"" +
									"}" +
									"}" +
									"]," +
									"\"logical-cloud\": \"cloud1\"" +
									"}" +
									"}"),
						},
					},
				},
			},
		},
		{
			label: "Failed Update Consumer",
			inp: hpaModel.HpaResourceConsumer{
				MetaData: mtypes.Metadata{
					Name:        "unknownConsumer",
					Description: "Unknown Consumer for unit testing",
				},
			},
			expectedError: "dependency not found",
			mockdb: &db.MockDB{
				Err: pkgerrors.New("Error Updating Consumer"),
			},
		},
	}

	fmt.Printf("\n================== TestUpdateConsumer .. total_testcase_count[%d] ==================\n", len(testCases))
	for i, testCase := range testCases {
		t.Run(testCase.label, func(t *testing.T) {
			fmt.Printf("\n================== TestUpdateConsumer .. testcase_count[%d] testcase_name[%s] ==================\n", i, testCase.label)
			db.DBconn = testCase.mockdb
			impl := NewHpaPlacementClient()
			got, err := impl.AddConsumer(testCase.inp, "project1", "compositeapp1", "version1", "dgroup1", "intent1", true)
			if err != nil {
				if testCase.expectedError == "" {
					t.Fatalf("Update returned an unexpected error [%s]", err)
				}
				if strings.Contains(strings.ToLower(err.Error()), strings.ToLower(testCase.expectedError)) == false {
					t.Fatalf("Update returned an unexpected-error[%s] expected[%s]", err, testCase.expectedError)
				}
			} else {
				if reflect.DeepEqual(testCase.expected, got) == false {
					t.Errorf("Update returned unexpected body: got %v;"+
						" expected %v", got, testCase.expected)
				}
			}
		})
	}
}

func TestGetAllConsumers(t *testing.T) {

	testCases := []struct {
		label         string
		expectedError string
		mockdb        *db.MockDB
		inp           string
		expected      []hpaModel.HpaResourceConsumer
	}{
		{
			label: "GetAll Consumer",
			expected: []hpaModel.HpaResourceConsumer{
				{
					MetaData: mtypes.Metadata{
						Name:        "consumer1",
						Description: "Test Consumer for unit testing",
						UserData1:   "userData1",
						UserData2:   "userData2",
					},
				},
			},
			expectedError: "",
			mockdb: &db.MockDB{
				Items: []map[string]map[string][]byte{
					{
						orchMod.ProjectKey{ProjectName: "project1"}.String(): {
							"projectmetadata": []byte(
								"{\"project-name\":\"project1\"," +
									"\"description\":\"Test project for unit testing\"}"),
						},
						orchMod.CompositeAppKey{CompositeAppName: "compositeapp1",
							Version: "version1", Project: "project1"}.String(): {
							"compositeappmetadata": []byte(
								"{\"metadata\":{" +
									"\"name\":\"compositeapp1\"," +
									"\"description\":\"description\"," +
									"\"userData1\":\"user data\"," +
									"\"userData2\":\"user data\"" +
									"}," +
									"\"spec\":{" +
									"\"version\":\"version of the composite app\"}}"),
						},
						orchMod.DeploymentIntentGroupKey{
							Name:         "dgroup1",
							Project:      "project1",
							CompositeApp: "compositeapp1",
							Version:      "version1",
						}.String(): {
							"deploymentintentgroupmetadata": []byte(
								"{\"metadata\":{\"name\":\"dgroup1\"," +
									"\"description\":\"DescriptionTestDeploymentIntentGroup\"," +
									"\"userData1\": \"userData1\"," +
									"\"userData2\": \"userData2\"}," +
									"\"spec\":{\"profile\": \"Testprofile\"," +
									"\"version\": \"version of deployment\"," +
									"\"override-values\":[" +
									"{" +
									"\"app-name\": \"TestAppName\"," +
									"\"values\": " +
									"{" +
									"\"imageRepository\":\"registry.hub.docker.com\"" +
									"}" +
									"}," +
									"{" +
									"\"app-name\": \"TestAppName\"," +
									"\"values\": " +
									"{" +
									"\"imageRepository\":\"registry.hub.docker.com\"" +
									"}" +
									"}" +
									"]," +
									"\"logical-cloud\": \"cloud1\"" +
									"}" +
									"}"),
						},
						HpaIntentKey{IntentName: "intent1",
							Project: "project1", CompositeApp: "compositeapp1",
							Version: "version1", DeploymentIntentGroup: "dgroup1"}.String(): {
							"HpaPlacementControllerMetadata": []byte(
								"{" +
									"\"metadata\" : {" +
									"\"Name\":\"intent1\"," +
									"\"Description\":\"Test Intent for unit testing\"," +
									"\"UserData1\": \"userData1\"," +
									"\"UserData2\":\"userData2\"}" +
									"}"),
						},
						HpaConsumerKey{ConsumerName: "", IntentName: "intent1",
							Project: "project1", CompositeApp: "compositeapp1",
							Version: "version1", DeploymentIntentGroup: "dgroup1"}.String(): {
							"HpaPlacementControllerMetadata": []byte(
								"{" +
									"\"metadata\" : {" +
									"\"Name\":\"consumer1\"," +
									"\"Description\":\"Test Consumer for unit testing\"," +
									"\"UserData1\": \"userData1\"," +
									"\"UserData2\":\"userData2\"}" +
									"}"),
						},
					},
				},
			},
		},
		{
			label: "GetAll Consumer Unable to find the project",
			expected: []hpaModel.HpaResourceConsumer{
				{
					MetaData: mtypes.Metadata{
						Name:        "consumer1",
						Description: "Test Consumer for unit testing",
						UserData1:   "userData1",
						UserData2:   "userData2",
					},
				},
			},
			expectedError: "Unable to find the project",
			mockdb: &db.MockDB{
				Items: []map[string]map[string][]byte{
					{
						orchMod.CompositeAppKey{CompositeAppName: "compositeapp1",
							Version: "version1", Project: "project1"}.String(): {
							"compositeappmetadata": []byte(
								"{\"metadata\":{" +
									"\"name\":\"compositeapp1\"," +
									"\"description\":\"description\"," +
									"\"userData1\":\"user data\"," +
									"\"userData2\":\"user data\"" +
									"}," +
									"\"spec\":{" +
									"\"version\":\"version of the composite app\"}}"),
						},
						orchMod.DeploymentIntentGroupKey{
							Name:         "dgroup1",
							Project:      "project1",
							CompositeApp: "compositeapp1",
							Version:      "version1",
						}.String(): {
							"deploymentintentgroupmetadata": []byte(
								"{\"metadata\":{\"name\":\"dgroup1\"," +
									"\"description\":\"DescriptionTestDeploymentIntentGroup\"," +
									"\"userData1\": \"userData1\"," +
									"\"userData2\": \"userData2\"}," +
									"\"spec\":{\"profile\": \"Testprofile\"," +
									"\"version\": \"version of deployment\"," +
									"\"override-values\":[" +
									"{" +
									"\"app-name\": \"TestAppName\"," +
									"\"values\": " +
									"{" +
									"\"imageRepository\":\"registry.hub.docker.com\"" +
									"}" +
									"}," +
									"{" +
									"\"app-name\": \"TestAppName\"," +
									"\"values\": " +
									"{" +
									"\"imageRepository\":\"registry.hub.docker.com\"" +
									"}" +
									"}" +
									"]," +
									"\"logical-cloud\": \"cloud1\"" +
									"}" +
									"}"),
						},
						HpaIntentKey{IntentName: "intent1",
							Project: "project1", CompositeApp: "compositeapp1",
							Version: "version1", DeploymentIntentGroup: "dgroup1"}.String(): {
							"HpaPlacementControllerMetadata": []byte(
								"{" +
									"\"metadata\" : {" +
									"\"Name\":\"intent1\"," +
									"\"Description\":\"Test Intent for unit testing\"," +
									"\"UserData1\": \"userData1\"," +
									"\"UserData2\":\"userData2\"}" +
									"}"),
						},
						HpaConsumerKey{ConsumerName: "", IntentName: "intent1",
							Project: "project1", CompositeApp: "compositeapp1",
							Version: "version1", DeploymentIntentGroup: "dgroup1"}.String(): {
							"HpaPlacementControllerMetadata": []byte(
								"{" +
									"\"metadata\" : {" +
									"\"Name\":\"consumer1\"," +
									"\"Description\":\"Test Consumer for unit testing\"," +
									"\"UserData1\": \"userData1\"," +
									"\"UserData2\":\"userData2\"}" +
									"}"),
						},
					},
				},
			},
		},
		{
			label: "GetAll Consumer Unable to find the intent-name",
			expected: []hpaModel.HpaResourceConsumer{
				{
					MetaData: mtypes.Metadata{
						Name:        "consumer1",
						Description: "Test Consumer for unit testing",
						UserData1:   "userData1",
						UserData2:   "userData2",
					},
				},
			},
			expectedError: "Unable to find the intent-name",
			mockdb: &db.MockDB{
				Items: []map[string]map[string][]byte{
					{
						orchMod.ProjectKey{ProjectName: "project1"}.String(): {
							"projectmetadata": []byte(
								"{\"project-name\":\"project1\"," +
									"\"description\":\"Test project for unit testing\"}"),
						},
						orchMod.CompositeAppKey{CompositeAppName: "compositeapp1",
							Version: "version1", Project: "project1"}.String(): {
							"compositeappmetadata": []byte(
								"{\"metadata\":{" +
									"\"name\":\"compositeapp1\"," +
									"\"description\":\"description\"," +
									"\"userData1\":\"user data\"," +
									"\"userData2\":\"user data\"" +
									"}," +
									"\"spec\":{" +
									"\"version\":\"version of the composite app\"}}"),
						},
						orchMod.DeploymentIntentGroupKey{
							Name:         "dgroup1",
							Project:      "project1",
							CompositeApp: "compositeapp1",
							Version:      "version1",
						}.String(): {
							"deploymentintentgroupmetadata": []byte(
								"{\"metadata\":{\"name\":\"dgroup1\"," +
									"\"description\":\"DescriptionTestDeploymentIntentGroup\"," +
									"\"userData1\": \"userData1\"," +
									"\"userData2\": \"userData2\"}," +
									"\"spec\":{\"profile\": \"Testprofile\"," +
									"\"version\": \"version of deployment\"," +
									"\"override-values\":[" +
									"{" +
									"\"app-name\": \"TestAppName\"," +
									"\"values\": " +
									"{" +
									"\"imageRepository\":\"registry.hub.docker.com\"" +
									"}" +
									"}," +
									"{" +
									"\"app-name\": \"TestAppName\"," +
									"\"values\": " +
									"{" +
									"\"imageRepository\":\"registry.hub.docker.com\"" +
									"}" +
									"}" +
									"]," +
									"\"logical-cloud\": \"cloud1\"" +
									"}" +
									"}"),
						},
						HpaConsumerKey{ConsumerName: "", IntentName: "intent1",
							Project: "project1", CompositeApp: "compositeapp1",
							Version: "version1", DeploymentIntentGroup: "dgroup1"}.String(): {
							"HpaPlacementControllerMetadata": []byte(
								"{" +
									"\"metadata\" : {" +
									"\"Name\":\"consumer1\"," +
									"\"Description\":\"Test Consumer for unit testing\"," +
									"\"UserData1\": \"userData1\"," +
									"\"UserData2\":\"userData2\"}" +
									"}"),
						},
					},
				},
			},
		},
		{
			label:         "GetAll Error",
			expectedError: "DB Error",
			mockdb: &db.MockDB{
				Err: pkgerrors.New("DB Error"),
			},
		},
	}

	fmt.Printf("\n================== TestGetAllConsumers .. total_testcase_count[%d] ==================\n", len(testCases))
	for i, testCase := range testCases {
		fmt.Printf("\n================== TestGetAllConsumers .. testcase_count[%d] testcase_name[%s] ==================\n", i, testCase.label)
		t.Run(testCase.label, func(t *testing.T) {
			db.DBconn = testCase.mockdb
			impl := NewHpaPlacementClient()
			got, err := impl.GetAllConsumers("project1", "compositeapp1", "version1", "dgroup1", "intent1")
			if err != nil {
				if testCase.expectedError == "" {
					t.Fatalf("Get returned an unexpected error: %s", err)
				}
				if strings.Contains(strings.ToLower(err.Error()), strings.ToLower(testCase.expectedError)) == false {
					t.Fatalf("Get returned an unexpected error: err[%s] expectedError[%s]", err, testCase.expectedError)
				}
			} else {
				if reflect.DeepEqual(testCase.expected, got) == false {
					t.Errorf("Get returned unexpected body: got [%v];"+
						" expected [%v]", got, testCase.expected)
				}
			}
		})
	}
}
func TestGetConsumer(t *testing.T) {

	testCases := []struct {
		label         string
		name          string
		expectedError string
		mockdb        *db.MockDB
		inp           string
		expected      hpaModel.HpaResourceConsumer
	}{
		{
			label: "Get Consumer",
			name:  "consumer1",
			expected: hpaModel.HpaResourceConsumer{
				MetaData: mtypes.Metadata{
					Name:        "consumer1",
					Description: "Test Consumer for unit testing",
					UserData1:   "userData1",
					UserData2:   "userData2",
				},
			},
			expectedError: "",
			mockdb: &db.MockDB{
				Items: []map[string]map[string][]byte{
					{
						orchMod.ProjectKey{ProjectName: "project1"}.String(): {
							"projectmetadata": []byte(
								"{\"project-name\":\"project1\"," +
									"\"description\":\"Test project for unit testing\"}"),
						},
						orchMod.CompositeAppKey{CompositeAppName: "compositeapp1",
							Version: "version1", Project: "project1"}.String(): {
							"compositeappmetadata": []byte(
								"{\"metadata\":{" +
									"\"name\":\"compositeapp1\"," +
									"\"description\":\"description\"," +
									"\"userData1\":\"user data\"," +
									"\"userData2\":\"user data\"" +
									"}," +
									"\"spec\":{" +
									"\"version\":\"version of the composite app\"}}"),
						},
						orchMod.DeploymentIntentGroupKey{
							Name:         "dgroup1",
							Project:      "project1",
							CompositeApp: "compositeapp1",
							Version:      "version1",
						}.String(): {
							"deploymentintentgroupmetadata": []byte(
								"{\"metadata\":{\"name\":\"dgroup1\"," +
									"\"description\":\"DescriptionTestDeploymentIntentGroup\"," +
									"\"userData1\": \"userData1\"," +
									"\"userData2\": \"userData2\"}," +
									"\"spec\":{\"profile\": \"Testprofile\"," +
									"\"version\": \"version of deployment\"," +
									"\"override-values\":[" +
									"{" +
									"\"app-name\": \"TestAppName\"," +
									"\"values\": " +
									"{" +
									"\"imageRepository\":\"registry.hub.docker.com\"" +
									"}" +
									"}," +
									"{" +
									"\"app-name\": \"TestAppName\"," +
									"\"values\": " +
									"{" +
									"\"imageRepository\":\"registry.hub.docker.com\"" +
									"}" +
									"}" +
									"]," +
									"\"logical-cloud\": \"cloud1\"" +
									"}" +
									"}"),
						},
						HpaIntentKey{IntentName: "intent1",
							Project: "project1", CompositeApp: "compositeapp1",
							Version: "version1", DeploymentIntentGroup: "dgroup1"}.String(): {
							"HpaPlacementControllerMetadata": []byte(
								"{" +
									"\"metadata\" : {" +
									"\"Name\":\"intent1\"," +
									"\"Description\":\"Test Intent for unit testing\"," +
									"\"UserData1\": \"userData1\"," +
									"\"UserData2\":\"userData2\"}" +
									"}"),
						},
						HpaConsumerKey{ConsumerName: "consumer1", IntentName: "intent1",
							Project: "project1", CompositeApp: "compositeapp1",
							Version: "version1", DeploymentIntentGroup: "dgroup1"}.String(): {
							"HpaPlacementControllerMetadata": []byte(
								"{" +
									"\"metadata\" : {" +
									"\"Name\":\"consumer1\"," +
									"\"Description\":\"Test Consumer for unit testing\"," +
									"\"UserData1\": \"userData1\"," +
									"\"UserData2\":\"userData2\"}" +
									"}"),
						},
					},
				},
			},
		},
		{
			label: "Get Consumer Unable to find the project",
			name:  "consumer1",
			expected: hpaModel.HpaResourceConsumer{
				MetaData: mtypes.Metadata{
					Name:        "consumer1",
					Description: "Test Consumer for unit testing",
					UserData1:   "userData1",
					UserData2:   "userData2",
				},
			},
			expectedError: "Unable to find the project",
			mockdb: &db.MockDB{
				Items: []map[string]map[string][]byte{
					{
						orchMod.CompositeAppKey{CompositeAppName: "compositeapp1",
							Version: "version1", Project: "project1"}.String(): {
							"compositeappmetadata": []byte(
								"{\"metadata\":{" +
									"\"name\":\"compositeapp1\"," +
									"\"description\":\"description\"," +
									"\"userData1\":\"user data\"," +
									"\"userData2\":\"user data\"" +
									"}," +
									"\"spec\":{" +
									"\"version\":\"version of the composite app\"}}"),
						},
						orchMod.DeploymentIntentGroupKey{
							Name:         "dgroup1",
							Project:      "project1",
							CompositeApp: "compositeapp1",
							Version:      "version1",
						}.String(): {
							"deploymentintentgroupmetadata": []byte(
								"{\"metadata\":{\"name\":\"dgroup1\"," +
									"\"description\":\"DescriptionTestDeploymentIntentGroup\"," +
									"\"userData1\": \"userData1\"," +
									"\"userData2\": \"userData2\"}," +
									"\"spec\":{\"profile\": \"Testprofile\"," +
									"\"version\": \"version of deployment\"," +
									"\"override-values\":[" +
									"{" +
									"\"app-name\": \"TestAppName\"," +
									"\"values\": " +
									"{" +
									"\"imageRepository\":\"registry.hub.docker.com\"" +
									"}" +
									"}," +
									"{" +
									"\"app-name\": \"TestAppName\"," +
									"\"values\": " +
									"{" +
									"\"imageRepository\":\"registry.hub.docker.com\"" +
									"}" +
									"}" +
									"]," +
									"\"logical-cloud\": \"cloud1\"" +
									"}" +
									"}"),
						},
						HpaIntentKey{IntentName: "intent1",
							Project: "project1", CompositeApp: "compositeapp1",
							Version: "version1", DeploymentIntentGroup: "dgroup1"}.String(): {
							"HpaPlacementControllerMetadata": []byte(
								"{" +
									"\"metadata\" : {" +
									"\"Name\":\"intent1\"," +
									"\"Description\":\"Test Intent for unit testing\"," +
									"\"UserData1\": \"userData1\"," +
									"\"UserData2\":\"userData2\"}" +
									"}"),
						},
						HpaConsumerKey{ConsumerName: "consumer1", IntentName: "intent1",
							Project: "project1", CompositeApp: "compositeapp1",
							Version: "version1", DeploymentIntentGroup: "dgroup1"}.String(): {
							"HpaPlacementControllerMetadata": []byte(
								"{" +
									"\"metadata\" : {" +
									"\"Name\":\"consumer1\"," +
									"\"Description\":\"Test Consumer for unit testing\"," +
									"\"UserData1\": \"userData1\"," +
									"\"UserData2\":\"userData2\"}" +
									"}"),
						},
					},
				},
			},
		},
		{
			label: "Get Consumer Unable to find the intent-name",
			name:  "consumer1",
			expected: hpaModel.HpaResourceConsumer{
				MetaData: mtypes.Metadata{
					Name:        "consumer1",
					Description: "Test Consumer for unit testing",
					UserData1:   "userData1",
					UserData2:   "userData2",
				},
			},
			expectedError: "Unable to find the intent-name",
			mockdb: &db.MockDB{
				Items: []map[string]map[string][]byte{
					{
						orchMod.ProjectKey{ProjectName: "project1"}.String(): {
							"projectmetadata": []byte(
								"{\"project-name\":\"project1\"," +
									"\"description\":\"Test project for unit testing\"}"),
						},
						orchMod.CompositeAppKey{CompositeAppName: "compositeapp1",
							Version: "version1", Project: "project1"}.String(): {
							"compositeappmetadata": []byte(
								"{\"metadata\":{" +
									"\"name\":\"compositeapp1\"," +
									"\"description\":\"description\"," +
									"\"userData1\":\"user data\"," +
									"\"userData2\":\"user data\"" +
									"}," +
									"\"spec\":{" +
									"\"version\":\"version of the composite app\"}}"),
						},
						orchMod.DeploymentIntentGroupKey{
							Name:         "dgroup1",
							Project:      "project1",
							CompositeApp: "compositeapp1",
							Version:      "version1",
						}.String(): {
							"deploymentintentgroupmetadata": []byte(
								"{\"metadata\":{\"name\":\"dgroup1\"," +
									"\"description\":\"DescriptionTestDeploymentIntentGroup\"," +
									"\"userData1\": \"userData1\"," +
									"\"userData2\": \"userData2\"}," +
									"\"spec\":{\"profile\": \"Testprofile\"," +
									"\"version\": \"version of deployment\"," +
									"\"override-values\":[" +
									"{" +
									"\"app-name\": \"TestAppName\"," +
									"\"values\": " +
									"{" +
									"\"imageRepository\":\"registry.hub.docker.com\"" +
									"}" +
									"}," +
									"{" +
									"\"app-name\": \"TestAppName\"," +
									"\"values\": " +
									"{" +
									"\"imageRepository\":\"registry.hub.docker.com\"" +
									"}" +
									"}" +
									"]," +
									"\"logical-cloud\": \"cloud1\"" +
									"}" +
									"}"),
						},
						HpaConsumerKey{ConsumerName: "consumer1", IntentName: "intent1",
							Project: "project1", CompositeApp: "compositeapp1",
							Version: "version1", DeploymentIntentGroup: "dgroup1"}.String(): {
							"HpaPlacementControllerMetadata": []byte(
								"{" +
									"\"metadata\" : {" +
									"\"Name\":\"consumer1\"," +
									"\"Description\":\"Test Consumer for unit testing\"," +
									"\"UserData1\": \"userData1\"," +
									"\"UserData2\":\"userData2\"}" +
									"}"),
						},
					},
				},
			},
		},
		{
			label:         "Get Error",
			expectedError: "DB Error",
			mockdb: &db.MockDB{
				Err: pkgerrors.New("DB Error"),
			},
		},
	}

	fmt.Printf("\n================== TestGetConsumer .. total_testcase_count[%d] ==================\n", len(testCases))
	for i, testCase := range testCases {
		fmt.Printf("\n================== TestGetConsumer .. testcase_count[%d] testcase_name[%s] ==================\n", i, testCase.label)
		t.Run(testCase.label, func(t *testing.T) {
			db.DBconn = testCase.mockdb
			impl := NewHpaPlacementClient()
			got, _, err := impl.GetConsumer(testCase.name, "project1", "compositeapp1", "version1", "dgroup1", "intent1")
			if err != nil {
				if testCase.expectedError == "" {
					t.Fatalf("Get returned an unexpected error: %s", err)
				}
				if strings.Contains(strings.ToLower(err.Error()), strings.ToLower(testCase.expectedError)) == false {
					t.Fatalf("Get returned an unexpected error: %s", err)
				}
			} else {
				if reflect.DeepEqual(testCase.expected, got) == false {
					t.Errorf("Get returned unexpected body: got %v;"+
						" expected %v", got, testCase.expected)
				}
			}
		})
	}
}

func TestGetConsumerByName(t *testing.T) {

	testCases := []struct {
		label         string
		name          string
		expectedError string
		mockdb        *db.MockDB
		inp           string
		expected      hpaModel.HpaResourceConsumer
	}{
		{
			label: "GetConsumerByName Consumer",
			name:  "consumer1",
			expected: hpaModel.HpaResourceConsumer{
				MetaData: mtypes.Metadata{
					Name:        "consumer1",
					Description: "Test Consumer for unit testing",
					UserData1:   "user data",
					UserData2:   "user data",
				},
				Spec: hpaModel.HpaResourceConsumerSpec{Name: "deployment-1",
					ContainerName: "container-1"},
			},
			expectedError: "",
			mockdb: &db.MockDB{
				Items: []map[string]map[string][]byte{
					{
						orchMod.ProjectKey{ProjectName: "project1"}.String(): {
							"projectmetadata": []byte(
								"{\"project-name\":\"project1\"," +
									"\"description\":\"Test project for unit testing\"}"),
						},
						orchMod.CompositeAppKey{CompositeAppName: "compositeapp1",
							Version: "version1", Project: "project1"}.String(): {
							"compositeappmetadata": []byte(
								"{\"metadata\":{" +
									"\"name\":\"compositeapp1\"," +
									"\"description\":\"description\"," +
									"\"userData1\":\"user data\"," +
									"\"userData2\":\"user data\"" +
									"}," +
									"\"spec\":{" +
									"\"version\":\"version of the composite app\"}}"),
						},
						orchMod.DeploymentIntentGroupKey{
							Name:         "dgroup1",
							Project:      "project1",
							CompositeApp: "compositeapp1",
							Version:      "version1",
						}.String(): {
							"deploymentintentgroupmetadata": []byte(
								"{\"metadata\":{\"name\":\"dgroup1\"," +
									"\"description\":\"DescriptionTestDeploymentIntentGroup\"," +
									"\"userData1\": \"userData1\"," +
									"\"userData2\": \"userData2\"}," +
									"\"spec\":{\"profile\": \"Testprofile\"," +
									"\"version\": \"version of deployment\"," +
									"\"override-values\":[" +
									"{" +
									"\"app-name\": \"TestAppName\"," +
									"\"values\": " +
									"{" +
									"\"imageRepository\":\"registry.hub.docker.com\"" +
									"}" +
									"}," +
									"{" +
									"\"app-name\": \"TestAppName\"," +
									"\"values\": " +
									"{" +
									"\"imageRepository\":\"registry.hub.docker.com\"" +
									"}" +
									"}" +
									"]," +
									"\"logical-cloud\": \"cloud1\"" +
									"}" +
									"}"),
						},
						HpaIntentKey{IntentName: "intent1",
							Project: "project1", CompositeApp: "compositeapp1",
							Version: "version1", DeploymentIntentGroup: "dgroup1"}.String(): {
							"HpaPlacementControllerMetadata": []byte(
								"{" +
									"\"metadata\" : {" +
									"\"Name\":\"intent1\"," +
									"\"Description\":\"Test Intent for unit testing\"," +
									"\"UserData1\": \"userData1\"," +
									"\"UserData2\":\"userData2\"}" +
									"}"),
						},
						HpaConsumerKey{ConsumerName: "consumer1", IntentName: "intent1",
							Project: "project1", CompositeApp: "compositeapp1",
							Version: "version1", DeploymentIntentGroup: "dgroup1"}.String(): {
							"HpaPlacementControllerMetadata": []byte(
								"{\"metadata\":{" +
									"\"name\":\"consumer1\"," +
									"\"description\":\"Test Consumer for unit testing\"," +
									"\"userData1\":\"user data\"," +
									"\"userData2\":\"user data\"" +
									"}," +
									"\"spec\":{" +
									"\"name\":\"deployment-1\"," +
									"\"container-name\":\"container-1\"}}"),
						},
					},
				},
			},
		},
		{
			label:         "GetConsumerByName Consumer Unable to find the project",
			name:          "consumer1",
			expectedError: "Unable to find the project",
			mockdb: &db.MockDB{
				Items: []map[string]map[string][]byte{
					{
						orchMod.CompositeAppKey{CompositeAppName: "compositeapp1",
							Version: "version1", Project: "project1"}.String(): {
							"compositeappmetadata": []byte(
								"{\"metadata\":{" +
									"\"name\":\"compositeapp1\"," +
									"\"description\":\"description\"," +
									"\"userData1\":\"user data\"," +
									"\"userData2\":\"user data\"" +
									"}," +
									"\"spec\":{" +
									"\"version\":\"version of the composite app\"}}"),
						},
						orchMod.DeploymentIntentGroupKey{
							Name:         "dgroup1",
							Project:      "project1",
							CompositeApp: "compositeapp1",
							Version:      "version1",
						}.String(): {
							"deploymentintentgroupmetadata": []byte(
								"{\"metadata\":{\"name\":\"dgroup1\"," +
									"\"description\":\"DescriptionTestDeploymentIntentGroup\"," +
									"\"userData1\": \"userData1\"," +
									"\"userData2\": \"userData2\"}," +
									"\"spec\":{\"profile\": \"Testprofile\"," +
									"\"version\": \"version of deployment\"," +
									"\"override-values\":[" +
									"{" +
									"\"app-name\": \"TestAppName\"," +
									"\"values\": " +
									"{" +
									"\"imageRepository\":\"registry.hub.docker.com\"" +
									"}" +
									"}," +
									"{" +
									"\"app-name\": \"TestAppName\"," +
									"\"values\": " +
									"{" +
									"\"imageRepository\":\"registry.hub.docker.com\"" +
									"}" +
									"}" +
									"]," +
									"\"logical-cloud\": \"cloud1\"" +
									"}" +
									"}"),
						},
						HpaIntentKey{IntentName: "intent1",
							Project: "project1", CompositeApp: "compositeapp1",
							Version: "version1", DeploymentIntentGroup: "dgroup1"}.String(): {
							"HpaPlacementControllerMetadata": []byte(
								"{" +
									"\"metadata\" : {" +
									"\"Name\":\"intent1\"," +
									"\"Description\":\"Test Intent for unit testing\"," +
									"\"UserData1\": \"userData1\"," +
									"\"UserData2\":\"userData2\"}" +
									"}"),
						},
						HpaConsumerKey{ConsumerName: "consumer1", IntentName: "intent1",
							Project: "project1", CompositeApp: "compositeapp1",
							Version: "version1", DeploymentIntentGroup: "dgroup1"}.String(): {
							"HpaPlacementControllerMetadata": []byte(
								"{" +
									"\"metadata\" : {" +
									"\"Name\":\"consumer1\"," +
									"\"Description\":\"Test Consumer for unit testing\"," +
									"\"UserData1\": \"userData1\"," +
									"\"UserData2\":\"userData2\"}" +
									"}"),
						},
					},
				},
			},
		},
		{
			label:         "GetConsumerByName Consumer Unable to find the intent-name",
			name:          "consumer1",
			expectedError: "Unable to find the intent-name",
			mockdb: &db.MockDB{
				Items: []map[string]map[string][]byte{
					{
						orchMod.ProjectKey{ProjectName: "project1"}.String(): {
							"projectmetadata": []byte(
								"{\"project-name\":\"project1\"," +
									"\"description\":\"Test project for unit testing\"}"),
						},
						orchMod.CompositeAppKey{CompositeAppName: "compositeapp1",
							Version: "version1", Project: "project1"}.String(): {
							"compositeappmetadata": []byte(
								"{\"metadata\":{" +
									"\"name\":\"compositeapp1\"," +
									"\"description\":\"description\"," +
									"\"userData1\":\"user data\"," +
									"\"userData2\":\"user data\"" +
									"}," +
									"\"spec\":{" +
									"\"version\":\"version of the composite app\"}}"),
						},
						orchMod.DeploymentIntentGroupKey{
							Name:         "dgroup1",
							Project:      "project1",
							CompositeApp: "compositeapp1",
							Version:      "version1",
						}.String(): {
							"deploymentintentgroupmetadata": []byte(
								"{\"metadata\":{\"name\":\"dgroup1\"," +
									"\"description\":\"DescriptionTestDeploymentIntentGroup\"," +
									"\"userData1\": \"userData1\"," +
									"\"userData2\": \"userData2\"}," +
									"\"spec\":{\"profile\": \"Testprofile\"," +
									"\"version\": \"version of deployment\"," +
									"\"override-values\":[" +
									"{" +
									"\"app-name\": \"TestAppName\"," +
									"\"values\": " +
									"{" +
									"\"imageRepository\":\"registry.hub.docker.com\"" +
									"}" +
									"}," +
									"{" +
									"\"app-name\": \"TestAppName\"," +
									"\"values\": " +
									"{" +
									"\"imageRepository\":\"registry.hub.docker.com\"" +
									"}" +
									"}" +
									"]," +
									"\"logical-cloud\": \"cloud1\"" +
									"}" +
									"}"),
						},
						HpaConsumerKey{ConsumerName: "consumer1", IntentName: "intent1",
							Project: "project1", CompositeApp: "compositeapp1",
							Version: "version1", DeploymentIntentGroup: "dgroup1"}.String(): {
							"HpaPlacementControllerMetadata": []byte(
								"{" +
									"\"metadata\" : {" +
									"\"Name\":\"consumer1\"," +
									"\"Description\":\"Test Consumer for unit testing\"," +
									"\"UserData1\": \"userData1\"," +
									"\"UserData2\":\"userData2\"}" +
									"}"),
						},
					},
				},
			},
		},
		{
			label:         "GetConsumerByName Error",
			expectedError: "DB Error",
			mockdb: &db.MockDB{
				Err: pkgerrors.New("DB Error"),
			},
		},
	}

	fmt.Printf("\n================== TestGetConsumerByName .. total_testcase_count[%d] ==================\n", len(testCases))
	for i, testCase := range testCases {
		fmt.Printf("\n================== TestGetConsumerByName .. testcase_count[%d] testcase_name[%s] ==================\n", i, testCase.label)
		t.Run(testCase.label, func(t *testing.T) {
			db.DBconn = testCase.mockdb
			impl := NewHpaPlacementClient()
			got, err := impl.GetConsumerByName(testCase.name, "project1", "compositeapp1", "version1", "dgroup1", "intent1")
			if err != nil {
				if testCase.expectedError == "" {
					t.Fatalf("Get returned an unexpected error: %s", err)
				}
				if strings.Contains(strings.ToLower(err.Error()), strings.ToLower(testCase.expectedError)) == false {
					t.Fatalf("Get returned an unexpected error: %s", err)
				}
			} else {
				if reflect.DeepEqual(testCase.expected, got) == false {
					t.Errorf("Get returned unexpected body: got %v;"+
						" expected %v", got, testCase.expected)
				}
			}
		})
	}
}
func TestDeleteConsumer(t *testing.T) {

	testCases := []struct {
		label         string
		name          string
		expectedError string
		mockdb        *db.MockDB
	}{
		{
			label:         "Delete Consumer",
			expectedError: "",
			name:          "consumer1",
			mockdb: &db.MockDB{
				Items: []map[string]map[string][]byte{
					{
						orchMod.ProjectKey{ProjectName: "project1"}.String(): {
							"projectmetadata": []byte(
								"{\"project-name\":\"project1\"," +
									"\"description\":\"Test project for unit testing\"}"),
						},
						orchMod.CompositeAppKey{CompositeAppName: "compositeapp1",
							Version: "version1", Project: "project1"}.String(): {
							"compositeappmetadata": []byte(
								"{\"metadata\":{" +
									"\"name\":\"compositeapp1\"," +
									"\"description\":\"description\"," +
									"\"userData1\":\"user data\"," +
									"\"userData2\":\"user data\"" +
									"}," +
									"\"spec\":{" +
									"\"version\":\"version of the composite app\"}}"),
						},
						orchMod.DeploymentIntentGroupKey{
							Name:         "dgroup1",
							Project:      "project1",
							CompositeApp: "compositeapp1",
							Version:      "version1",
						}.String(): {
							"deploymentintentgroupmetadata": []byte(
								"{\"metadata\":{\"name\":\"dgroup1\"," +
									"\"description\":\"DescriptionTestDeploymentIntentGroup\"," +
									"\"userData1\": \"userData1\"," +
									"\"userData2\": \"userData2\"}," +
									"\"spec\":{\"profile\": \"Testprofile\"," +
									"\"version\": \"version of deployment\"," +
									"\"override-values\":[" +
									"{" +
									"\"app-name\": \"TestAppName\"," +
									"\"values\": " +
									"{" +
									"\"imageRepository\":\"registry.hub.docker.com\"" +
									"}" +
									"}," +
									"{" +
									"\"app-name\": \"TestAppName\"," +
									"\"values\": " +
									"{" +
									"\"imageRepository\":\"registry.hub.docker.com\"" +
									"}" +
									"}" +
									"]," +
									"\"logical-cloud\": \"cloud1\"" +
									"}" +
									"}"),
						},
						HpaIntentKey{IntentName: "intent1",
							Project: "project1", CompositeApp: "compositeapp1",
							Version: "version1", DeploymentIntentGroup: "dgroup1"}.String(): {
							"HpaPlacementControllerMetadata": []byte(
								"{" +
									"\"metadata\" : {" +
									"\"Name\":\"intent1\"," +
									"\"Description\":\"Test Intent for unit testing\"," +
									"\"UserData1\": \"userData1\"," +
									"\"UserData2\":\"userData2\"}" +
									"}"),
						},
						HpaConsumerKey{ConsumerName: "consumer1", IntentName: "intent1",
							Project: "project1", CompositeApp: "compositeapp1",
							Version: "version1", DeploymentIntentGroup: "dgroup1"}.String(): {
							"HpaPlacementControllerMetadata": []byte(
								"{" +
									"\"metadata\" : {" +
									"\"Name\":\"consumer1\"," +
									"\"Description\":\"Test Consumer for unit testing\"," +
									"\"UserData1\": \"userData1\"," +
									"\"UserData2\":\"userData2\"}" +
									"}"),
						},
					},
				},
			},
		},
		{
			label:         "Delete Error",
			expectedError: "DB Error",
			mockdb: &db.MockDB{
				Err: pkgerrors.New("DB Error"),
			},
		},
		{
			label:         "Delete Consumer",
			expectedError: "Error",
			name:          "consumer1",
			mockdb: &db.MockDB{
				Items: []map[string]map[string][]byte{
					{
						orchMod.ProjectKey{ProjectName: "project1"}.String(): {
							"projectmetadata": []byte(
								"{\"project-name\":\"project1\"," +
									"\"description\":\"Test project for unit testing\"}"),
						},
						orchMod.CompositeAppKey{CompositeAppName: "compositeapp1",
							Version: "version1", Project: "project1"}.String(): {
							"compositeappmetadata": []byte(
								"{\"metadata\":{" +
									"\"name\":\"compositeapp1\"," +
									"\"description\":\"description\"," +
									"\"userData1\":\"user data\"," +
									"\"userData2\":\"user data\"" +
									"}," +
									"\"spec\":{" +
									"\"version\":\"version of the composite app\"}}"),
						},
						orchMod.DeploymentIntentGroupKey{
							Name:         "dgroup1",
							Project:      "project1",
							CompositeApp: "compositeapp1",
							Version:      "version1",
						}.String(): {
							"deploymentintentgroupmetadata": []byte(
								"{\"metadata\":{\"name\":\"dgroup1\"," +
									"\"description\":\"DescriptionTestDeploymentIntentGroup\"," +
									"\"userData1\": \"userData1\"," +
									"\"userData2\": \"userData2\"}," +
									"\"spec\":{\"profile\": \"Testprofile\"," +
									"\"version\": \"version of deployment\"," +
									"\"override-values\":[" +
									"{" +
									"\"app-name\": \"TestAppName\"," +
									"\"values\": " +
									"{" +
									"\"imageRepository\":\"registry.hub.docker.com\"" +
									"}" +
									"}," +
									"{" +
									"\"app-name\": \"TestAppName\"," +
									"\"values\": " +
									"{" +
									"\"imageRepository\":\"registry.hub.docker.com\"" +
									"}" +
									"}" +
									"]," +
									"\"logical-cloud\": \"cloud1\"" +
									"}" +
									"}"),
						},
						HpaIntentKey{IntentName: "intent1",
							Project: "project1", CompositeApp: "compositeapp1",
							Version: "version1", DeploymentIntentGroup: "dgroup1"}.String(): {
							"HpaPlacementControllerMetadata": []byte(
								"{" +
									"\"metadata\" : {" +
									"\"Name\":\"intent1\"," +
									"\"Description\":\"Test Intent for unit testing\"," +
									"\"UserData1\": \"userData1\"," +
									"\"UserData2\":\"userData2\"}" +
									"}"),
						},
					},
				},
			},
		},
	}

	fmt.Printf("\n================== TestDeleteConsumer .. total_testcase_count[%d] ==================\n", len(testCases))
	for i, testCase := range testCases {
		fmt.Printf("\n================== TestDeleteConsumer .. testcase_count[%d] testcase_name[%s] ==================\n", i, testCase.label)
		t.Run(testCase.label, func(t *testing.T) {
			db.DBconn = testCase.mockdb
			impl := NewHpaPlacementClient()
			err := impl.DeleteConsumer(testCase.name, "project1", "compositeapp1", "version1", "dgroup1", "intent1")
			if err != nil {
				if testCase.expectedError == "" {
					t.Fatalf("Delete returned an unexpected-error[%s] expected[%s]", err, testCase.expectedError)
				}
				if strings.Contains(strings.ToLower(err.Error()), strings.ToLower(testCase.expectedError)) == false {
					t.Fatalf("Delete returned an unexpected-error[%s] expected[%s]", err, testCase.expectedError)
				}
			}
		})
	}
}
