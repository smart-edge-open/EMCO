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

var allocatable_true = true

func TestCreateResource(t *testing.T) {
	testCases := []struct {
		label         string
		inp           hpaModel.HpaResourceRequirement
		expectedError string
		mockdb        *db.MockDB
		expected      hpaModel.HpaResourceRequirement
	}{
		{
			label: "Create Resource",
			inp: hpaModel.HpaResourceRequirement{
				MetaData: mtypes.Metadata{
					Name:        "resource1",
					Description: "A sample Resource used for unit testing",
					UserData1:   "data1",
					UserData2:   "data2",
				},
			},
			expected: hpaModel.HpaResourceRequirement{
				MetaData: mtypes.Metadata{
					Name:        "resource1",
					Description: "A sample Resource used for unit testing",
					UserData1:   "data1",
					UserData2:   "data2",
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
			label:         "Failed Create Resource",
			expectedError: "dependency not found",
			mockdb: &db.MockDB{
				Err: pkgerrors.New("Error Creating Resource"),
			},
		},
		{
			label: "Create Existing Resource",
			inp: hpaModel.HpaResourceRequirement{
				MetaData: mtypes.Metadata{
					Name:        "resource1",
					Description: "A sample Resource used for unit testing",
					UserData1:   "data1",
					UserData2:   "data2",
				},
			},
			expectedError: "Resource already exists",
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
						HpaResourceKey{ResourceName: "resource1", ConsumerName: "consumer1", IntentName: "intent1",
							Project: "project1", CompositeApp: "compositeapp1",
							Version: "version1", DeploymentIntentGroup: "dgroup1"}.String(): {
							"HpaPlacementControllerMetadata": []byte(
								"{" +
									"\"metadata\" : {" +
									"\"Name\":\"resource1\"," +
									"\"Description\":\"Test Resource for unit testing\"," +
									"\"UserData1\":\"userData1\"," +
									"\"UserData2\":\"userData2\"}" +
									"}"),
						},
					},
				},
			},
		},
	}

	fmt.Printf("\n================== TestCreateResource .. total_testcase_count[%d] ==================\n", len(testCases))
	for i, testCase := range testCases {
		t.Run(testCase.label, func(t *testing.T) {
			fmt.Printf("\n================== TestCreateResource .. testcase_count[%d] testcase_name[%s] ==================\n", i, testCase.label)
			db.DBconn = testCase.mockdb
			impl := NewHpaPlacementClient()
			got, err := impl.AddResource(testCase.inp, "project1", "compositeapp1", "version1", "dgroup1", "intent1", "consumer1", false)
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

func TestUpdateResource(t *testing.T) {
	testCases := []struct {
		label         string
		inp           hpaModel.HpaResourceRequirement
		expectedError string
		mockdb        *db.MockDB
		expected      hpaModel.HpaResourceRequirement
	}{
		{
			label: "Update Resource",
			inp: hpaModel.HpaResourceRequirement{
				MetaData: mtypes.Metadata{
					Name:        "resource1",
					Description: "Test Resource for unit testing",
					UserData1:   "update userData1",
					UserData2:   "update userData2",
				},
			},
			expected: hpaModel.HpaResourceRequirement{
				MetaData: mtypes.Metadata{
					Name:        "resource1",
					Description: "Test Resource for unit testing",
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
						HpaResourceKey{ResourceName: "resource1", ConsumerName: "consumer1", IntentName: "intent1",
							Project: "project1", CompositeApp: "compositeapp1",
							Version: "version1", DeploymentIntentGroup: "dgroup1"}.String(): {
							"HpaPlacementControllerMetadata": []byte(
								"{" +
									"\"metadata\" : {" +
									"\"Name\":\"resource1\"," +
									"\"Description\":\"Test Resource for unit testing\"," +
									"\"UserData1\":\"userData1\"," +
									"\"UserData2\":\"userData2\"}" +
									"}"),
						},
					},
				},
			},
		},
		{
			label: "Failed Update Resource",
			inp: hpaModel.HpaResourceRequirement{
				MetaData: mtypes.Metadata{
					Name:        "unknownResource",
					Description: "Unknown Resource for unit testing",
				},
			},
			expectedError: "dependency not found",
			mockdb: &db.MockDB{
				Err: pkgerrors.New("Error Updating Resource"),
			},
		},
	}

	fmt.Printf("\n================== TestUpdateResource .. total_testcase_count[%d] ==================\n", len(testCases))
	for i, testCase := range testCases {
		t.Run(testCase.label, func(t *testing.T) {
			fmt.Printf("\n================== TestUpdateResource .. testcase_count[%d] testcase_name[%s] ==================\n", i, testCase.label)
			db.DBconn = testCase.mockdb
			impl := NewHpaPlacementClient()
			got, err := impl.AddResource(testCase.inp, "project1", "compositeapp1", "version1", "dgroup1", "intent1", "consumer1", true)
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

func TestGetAllResources(t *testing.T) {

	testCases := []struct {
		label         string
		expectedError string
		mockdb        *db.MockDB
		inp           string
		expected      []hpaModel.HpaResourceRequirement
	}{
		{
			label: "GetAll Resource",
			expected: []hpaModel.HpaResourceRequirement{
				{
					MetaData: mtypes.Metadata{
						Name:        "resource1",
						Description: "Test Resource for unit testing",
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
						HpaResourceKey{ResourceName: "", ConsumerName: "consumer1", IntentName: "intent1",
							Project: "project1", CompositeApp: "compositeapp1",
							Version: "version1", DeploymentIntentGroup: "dgroup1"}.String(): {
							"HpaPlacementControllerMetadata": []byte(
								"{" +
									"\"metadata\" : {" +
									"\"Name\":\"resource1\"," +
									"\"Description\":\"Test Resource for unit testing\"," +
									"\"UserData1\":\"userData1\"," +
									"\"UserData2\":\"userData2\"}" +
									"}"),
						},
					},
				},
			},
		},
		{
			label: "GetAll Resource",
			expected: []hpaModel.HpaResourceRequirement{
				{
					MetaData: mtypes.Metadata{
						Name:        "resource1",
						Description: "Test Resource for unit testing",
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
						HpaResourceKey{ResourceName: "", ConsumerName: "consumer1", IntentName: "intent1",
							Project: "project1", CompositeApp: "compositeapp1",
							Version: "version1", DeploymentIntentGroup: "dgroup1"}.String(): {
							"HpaPlacementControllerMetadata": []byte(
								"{" +
									"\"metadata\" : {" +
									"\"Name\":\"resource1\"," +
									"\"Description\":\"Test Resource for unit testing\"," +
									"\"UserData1\":\"userData1\"," +
									"\"UserData2\":\"userData2\"}" +
									"}"),
						},
					},
				},
			},
		},
		{
			label: "GetAll Resource Unable to find the consumer-name",
			expected: []hpaModel.HpaResourceRequirement{
				{
					MetaData: mtypes.Metadata{
						Name:        "resource1",
						Description: "Test Resource for unit testing",
						UserData1:   "userData1",
						UserData2:   "userData2",
					},
				},
			},
			expectedError: "Unable to find the consumer-name",
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
						HpaResourceKey{ResourceName: "", ConsumerName: "consumer1", IntentName: "intent1",
							Project: "project1", CompositeApp: "compositeapp1",
							Version: "version1", DeploymentIntentGroup: "dgroup1"}.String(): {
							"HpaPlacementControllerMetadata": []byte(
								"{" +
									"\"metadata\" : {" +
									"\"Name\":\"resource1\"," +
									"\"Description\":\"Test Resource for unit testing\"," +
									"\"UserData1\":\"userData1\"," +
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

	fmt.Printf("\n================== TestGetAllResourcess .. total_testcase_count[%d] ==================\n", len(testCases))
	for i, testCase := range testCases {
		fmt.Printf("\n================== TestGetAllResourcess .. testcase_count[%d] testcase_name[%s] ==================\n", i, testCase.label)
		t.Run(testCase.label, func(t *testing.T) {
			db.DBconn = testCase.mockdb
			impl := NewHpaPlacementClient()
			got, err := impl.GetAllResources("project1", "compositeapp1", "version1", "dgroup1", "intent1", "consumer1")
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

func TestGetResource(t *testing.T) {

	testCases := []struct {
		label         string
		name          string
		expectedError string
		mockdb        *db.MockDB
		inp           string
		expected      hpaModel.HpaResourceRequirement
	}{
		{
			label: "Get Resource",
			name:  "resource1",
			expected: hpaModel.HpaResourceRequirement{
				MetaData: mtypes.Metadata{
					Name:        "resource1",
					Description: "Test Resource for unit testing",
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
									"\"UserData1\":\"userData1\"," +
									"\"UserData2\":\"userData2\"}" +
									"}"),
						},
						HpaResourceKey{ResourceName: "resource1", ConsumerName: "consumer1", IntentName: "intent1",
							Project: "project1", CompositeApp: "compositeapp1",
							Version: "version1", DeploymentIntentGroup: "dgroup1"}.String(): {
							"HpaPlacementControllerMetadata": []byte(
								"{" +
									"\"metadata\" : {" +
									"\"Name\":\"resource1\"," +
									"\"Description\":\"Test Resource for unit testing\"," +
									"\"UserData1\":\"userData1\"," +
									"\"UserData2\":\"userData2\"}" +
									"}"),
						},
					},
				},
			},
		},
		{
			label: "Get Resource",
			name:  "resource1",
			expected: hpaModel.HpaResourceRequirement{
				MetaData: mtypes.Metadata{
					Name:        "resource1",
					Description: "Test Resource for unit testing",
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
									"\"UserData1\":\"userData1\"," +
									"\"UserData2\":\"userData2\"}" +
									"}"),
						},
						HpaResourceKey{ResourceName: "resource1", ConsumerName: "consumer1", IntentName: "intent1",
							Project: "project1", CompositeApp: "compositeapp1",
							Version: "version1", DeploymentIntentGroup: "dgroup1"}.String(): {
							"HpaPlacementControllerMetadata": []byte(
								"{" +
									"\"metadata\" : {" +
									"\"Name\":\"resource1\"," +
									"\"Description\":\"Test Resource for unit testing\"," +
									"\"UserData1\":\"userData1\"," +
									"\"UserData2\":\"userData2\"}" +
									"}"),
						},
					},
				},
			},
		},
		{
			label: "Get Resource Unable to find the consumer-name",
			name:  "resource1",
			expected: hpaModel.HpaResourceRequirement{
				MetaData: mtypes.Metadata{
					Name:        "resource1",
					Description: "Test Resource for unit testing",
					UserData1:   "userData1",
					UserData2:   "userData2",
				},
			},
			expectedError: "Unable to find the consumer-name",
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
						HpaResourceKey{ResourceName: "resource1", ConsumerName: "consumer1", IntentName: "intent1",
							Project: "project1", CompositeApp: "compositeapp1",
							Version: "version1", DeploymentIntentGroup: "dgroup1"}.String(): {
							"HpaPlacementControllerMetadata": []byte(
								"{" +
									"\"metadata\" : {" +
									"\"Name\":\"resource1\"," +
									"\"Description\":\"Test Resource for unit testing\"," +
									"\"UserData1\":\"userData1\"," +
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

	fmt.Printf("\n================== TestGetResource .. total_testcase_count[%d] ==================\n", len(testCases))
	for i, testCase := range testCases {
		fmt.Printf("\n================== TestGetResource .. testcase_count[%d] testcase_name[%s] ==================\n", i, testCase.label)
		t.Run(testCase.label, func(t *testing.T) {
			db.DBconn = testCase.mockdb
			impl := NewHpaPlacementClient()
			got, _, err := impl.GetResource(testCase.name, "project1", "compositeapp1", "version1", "dgroup1", "intent1", "consumer1")
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

func TestGetResourceByName(t *testing.T) {

	testCases := []struct {
		label         string
		name          string
		expectedError string
		mockdb        *db.MockDB
		inp           string
		expected      hpaModel.HpaResourceRequirement
	}{
		{
			label: "GetResourceByName Resource",
			name:  "resource1",
			expected: hpaModel.HpaResourceRequirement{
				MetaData: mtypes.Metadata{
					Name:        "resource1",
					Description: "Test Resource for unit testing",
					UserData1:   "userData1",
					UserData2:   "userData2",
				},
				Spec: hpaModel.HpaResourceRequirementSpec{
					Allocatable: &allocatable_true,
					Resource:    hpaModel.HpaResourceRequirementDetails{hpaModel.AllocatableResources{Name: "cpu", Requests: 1, Limits: 1}, hpaModel.NonAllocatableResources{}}},
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
						HpaResourceKey{ResourceName: "resource1", ConsumerName: "consumer1", IntentName: "intent1",
							Project: "project1", CompositeApp: "compositeapp1",
							Version: "version1", DeploymentIntentGroup: "dgroup1"}.String(): {
							"HpaPlacementControllerMetadata": []byte(
								"{" +
									"\"metadata\" : {" +
									"\"Name\":\"resource1\"," +
									"\"Description\":\"Test Resource for unit testing\"," +
									"\"UserData1\":\"userData1\"," +
									"\"UserData2\":\"userData2\"}," +
									"\"spec\" : {" +
									"\"allocatable\":true," +
									"\"resource\":{\"name\":\"cpu\", \"requests\":1, \"limits\":1}" +
									"}" +
									"}"),
						},
					},
				},
			},
		},
		{
			label:         "GetResourceByName Resource Unable to find the consumer-name",
			name:          "resource1",
			expectedError: "Unable to find the consumer-name",
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
						HpaResourceKey{ResourceName: "resource1", ConsumerName: "consumer1", IntentName: "intent1",
							Project: "project1", CompositeApp: "compositeapp1",
							Version: "version1", DeploymentIntentGroup: "dgroup1"}.String(): {
							"HpaPlacementControllerMetadata": []byte(
								"{" +
									"\"metadata\" : {" +
									"\"Name\":\"resource1\"," +
									"\"Description\":\"Test Resource for unit testing\"," +
									"\"UserData1\":\"userData1\"," +
									"\"UserData2\":\"userData2\"}" +
									"}"),
						},
					},
				},
			},
		},
		{
			label:         "GetResourceByName Error",
			expectedError: "DB Error",
			mockdb: &db.MockDB{
				Err: pkgerrors.New("DB Error"),
			},
		},
	}

	fmt.Printf("\n================== TestGetResourceByName .. total_testcase_count[%d] ==================\n", len(testCases))
	for i, testCase := range testCases {
		fmt.Printf("\n================== TestGetResourceByName .. testcase_count[%d] testcase_name[%s] ==================\n", i, testCase.label)
		t.Run(testCase.label, func(t *testing.T) {
			db.DBconn = testCase.mockdb
			impl := NewHpaPlacementClient()
			got, err := impl.GetResourceByName(testCase.name, "project1", "compositeapp1", "version1", "dgroup1", "intent1", "consumer1")
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
func TestDeleteResource(t *testing.T) {

	testCases := []struct {
		label         string
		name          string
		expectedError string
		mockdb        *db.MockDB
	}{
		{
			label:         "Delete Resource",
			expectedError: "",
			name:          "resource1",
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
						HpaResourceKey{ResourceName: "resource1", ConsumerName: "consumer1", IntentName: "intent1",
							Project: "project1", CompositeApp: "compositeapp1",
							Version: "version1", DeploymentIntentGroup: "dgroup1"}.String(): {
							"HpaPlacementControllerMetadata": []byte(
								"{" +
									"\"metadata\" : {" +
									"\"Name\":\"resource1\"," +
									"\"Description\":\"Test Resource for unit testing\"," +
									"\"UserData1\":\"userData1\"," +
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
	}

	fmt.Printf("\n================== TestDeleteResource .. total_testcase_count[%d] ==================\n", len(testCases))
	for i, testCase := range testCases {
		fmt.Printf("\n================== TestDeleteResource .. testcase_count[%d] testcase_name[%s] ==================\n", i, testCase.label)
		t.Run(testCase.label, func(t *testing.T) {
			db.DBconn = testCase.mockdb
			impl := NewHpaPlacementClient()
			err := impl.DeleteResource(testCase.name, "project1", "compositeapp1", "version1", "dgroup1", "intent1", "consumer1")
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
