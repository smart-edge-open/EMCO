package module

// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2020 Intel Corporation

import (
	"github.com/open-ness/EMCO/src/orchestrator/pkg/infra/db"
	moduleLib "github.com/open-ness/EMCO/src/orchestrator/pkg/module"
	"reflect"
	"strings"
	"testing"
)

func TestCreateCustomization(t *testing.T) {
	testCases := []struct {
		label                        string
		inputCustomization           Customization
		inputSpecFileContent         SpecFileContent
		inputProject                 string
		inputCompositeApp            string
		inputCompositeAppVersion     string
		inputDeploymentIntentGrpName string
		inputGenericK8sIntent        string
		inputResource                string
		inputExists                  bool
		expectedError                string
		mockdb                       *db.MockDB
		expected                     Customization
	}{
		{
			label: "Create customization",
			inputCustomization: Customization{
				Metadata: Metadata{
					Name:        "testCustomization",
					Description: "testCustomization",
					UserData1:   "userData1",
					UserData2:   "userData2",
				},
				Spec: CustomizeSpec{
					ClusterSpecific: "true",
					ClusterInfo: ClusterInfo{
						Scope:           "label",
						ClusterProvider: "testClusterProvider",
						ClusterName:     "",
						ClusterLabel:    "testLabel",
						Mode:            "allow",
					},
					PatchType: "json",
					PatchJSON: []map[string]interface{}{
						{
							"op":    "replace",
							"path":  "/name",
							"value": "test",
						},
					},
				},
			},
			inputSpecFileContent: SpecFileContent{
				FileContents: []string{
					"This is testFile1",
					"This is testFile2",
				},
				FileNames: []string{
					"testFile1",
					"testFile2",
				},
			},
			inputProject:                 "testProject",
			inputCompositeApp:            "testCompositeApp",
			inputCompositeAppVersion:     "testCompositeAppVersion",
			inputDeploymentIntentGrpName: "testDeploymentIntentGroup",
			inputGenericK8sIntent:        "testGenK8sIntent",
			inputResource:                "testResource",
			inputExists:                  false,
			expected: Customization{
				Metadata: Metadata{
					Name:        "testCustomization",
					Description: "testCustomization",
					UserData1:   "userData1",
					UserData2:   "userData2",
				},
				Spec: CustomizeSpec{
					ClusterSpecific: "true",
					ClusterInfo: ClusterInfo{
						Scope:           "label",
						ClusterProvider: "testClusterProvider",
						ClusterName:     "",
						ClusterLabel:    "testLabel",
						Mode:            "allow",
					},
					PatchType: "json",
					PatchJSON: []map[string]interface{}{
						{
							"op":    "replace",
							"path":  "/name",
							"value": "test",
						},
					},
				},
			},
			expectedError: "",
			mockdb: &db.MockDB{
				Items: []map[string]map[string][]byte{
					{
						moduleLib.ProjectKey{ProjectName: "testProject"}.String(): {
							"projectmetadata": []byte(
								"{\"project-name\":\"testProject\"," +
									"\"description\":\"Test project for unit testing\"}"),
						},
						moduleLib.CompositeAppKey{CompositeAppName: "testCompositeApp",
							Version: "testCompositeAppVersion", Project: "testProject"}.String(): {
							"compositeappmetadata": []byte(
								"{\"metadata\":{" +
									"\"name\":\"testCompositeApp\"," +
									"\"description\":\"description\"," +
									"\"userData1\":\"user data\"," +
									"\"userData2\":\"user data\"" +
									"}," +
									"\"spec\":{" +
									"\"version\":\"version of the composite app\"}}"),
						},
						moduleLib.DeploymentIntentGroupKey{
							Name:         "testDeploymentIntentGroup",
							Project:      "testProject",
							CompositeApp: "testCompositeApp",
							Version:      "testCompositeAppVersion",
						}.String(): {
							"deploymentintentgroupmetadata": []byte(
								"{\"metadata\":{\"name\":\"testDeploymentIntentGroup\"," +
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
						GenericK8sIntentKey{
							GenericK8sIntent:    "testGenK8sIntent",
							Project:             "testProject",
							CompositeApp:        "testCompositeApp",
							CompositeAppVersion: "testCompositeAppVersion",
							DigName:             "testDeploymentIntentGroup",
						}.String(): {
							"generick8sintentmetadata": []byte(
								"{\"metadata\":{\"Name\":\"testGenK8sIntent\"," +
									"\"Description\":\"testGenK8sIntent\"," +
									"\"UserData1\": \"userData1\"," +
									"\"UserData2\": \"userData2\"}}"),
						},
						ResourceKey{
							Resource:            "testResource",
							Project:             "testProject",
							CompositeApp:        "testCompositeApp",
							CompositeAppVersion: "testCompositeAppVersion",
							DigName:             "testDeploymentIntentGroup",
							GenericK8sIntent:    "testGenK8sIntent",
						}.String(): {
							"resourcemetadata": []byte(
								"{\"metadata\":{\"name\":\"testResource\",\"description\":\"testResource\",\"userData1\":\"userData1\",\"userData2\":\"userData2\"},\"spec\":{\"appname\":\"testApp\",\"newobject\":\"True\",\"resourcegvk\":{\"apiversion\":\"v1\",\"kind\":\"configMap\",\"name\":\"TestCM\"}}}"),
						},
					},
				},
			},
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.label, func(t *testing.T) {
			db.DBconn = testCase.mockdb
			customCli := NewCustomizationClient()
			got, err := customCli.CreateCustomization(testCase.inputCustomization, testCase.inputSpecFileContent, testCase.inputProject, testCase.inputCompositeApp, testCase.inputCompositeAppVersion, testCase.inputDeploymentIntentGrpName, testCase.inputGenericK8sIntent, testCase.inputResource, testCase.inputExists)
			if err != nil {
				if testCase.expectedError == "" {
					t.Fatalf("CreateCustomization returned an unexpected error: %s", err)
				}
				if strings.Contains(err.Error(), testCase.expectedError) == false {
					t.Fatalf("CreateCustomization returned an unexpected error: %s", err)
				}
			} else {
				if reflect.DeepEqual(testCase.expected, got) == false {
					t.Errorf("CreateCustomization returned unexpected body: got %v;"+
						" expected %v", got, testCase.expected)
				}
			}
		})
	}
}

func TestGetCustomization(t *testing.T) {
	testCases := []struct {
		label                        string
		inputCustomization           string
		inputProject                 string
		inputCompositeApp            string
		inputCompositeAppVersion     string
		inputDeploymentIntentGrpName string
		inputGenericK8sIntent        string
		inputResource                string
		expectedError                string
		mockdb                       *db.MockDB
		expected                     Customization
	}{
		{
			label:                        "Get customization",
			inputCustomization:           "testCustomization",
			inputProject:                 "testProject",
			inputCompositeApp:            "testCompositeApp",
			inputCompositeAppVersion:     "testCompositeAppVersion",
			inputDeploymentIntentGrpName: "testDeploymentIntentGroup",
			inputGenericK8sIntent:        "testGenK8sIntent",
			inputResource:                "testResource",
			expected: Customization{

				Metadata: Metadata{
					Name:        "testCustomization",
					Description: "testCustomization",
					UserData1:   "userData1",
					UserData2:   "userData2",
				},
				Spec: CustomizeSpec{
					ClusterSpecific: "true",
					ClusterInfo: ClusterInfo{
						Scope:           "label",
						ClusterProvider: "testClusterProvider",
						ClusterName:     "",
						ClusterLabel:    "testLabel",
						Mode:            "allow",
					},
					PatchType: "json",
					PatchJSON: []map[string]interface{}{
						{
							"op":    "replace",
							"path":  "/name",
							"value": "test",
						},
					},
				},
			},
			expectedError: "",
			mockdb: &db.MockDB{
				Items: []map[string]map[string][]byte{
					{
						CustomizationKey{
							Customization:       "testCustomization",
							Project:             "testProject",
							CompositeApp:        "testCompositeApp",
							CompositeAppVersion: "testCompositeAppVersion",
							DigName:             "testDeploymentIntentGroup",
							GenericK8sIntent:    "testGenK8sIntent",
							Resource:            "testResource",
						}.String(): {
							"customizationmetadata": []byte(
								"{\"metadata\":{\"name\":\"testCustomization\",\"description\":\"testCustomization\",\"userData1\":\"userData1\",\"userData2\":\"userData2\"},\"spec\":{\"clusterSpecific\":\"true\",\"clusterInfo\":{\"scope\":\"label\",\"clusterProvider\":\"testClusterProvider\",\"clusterName\":\"\",\"clusterLabel\":\"testLabel\",\"mode\":\"allow\"},\"patchType\":\"json\",\"patchJson\":[{\"op\":\"replace\",\"path\":\"\\/name\",\"value\":\"test\"}]}}"),
						},
					},
				},
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.label, func(t *testing.T) {
			db.DBconn = testCase.mockdb
			customCli := NewCustomizationClient()
			got, err := customCli.GetCustomization(testCase.inputCustomization, testCase.inputProject, testCase.inputCompositeApp, testCase.inputCompositeAppVersion, testCase.inputDeploymentIntentGrpName, testCase.inputGenericK8sIntent, testCase.inputResource)
			if err != nil {
				if testCase.expectedError == "" {
					t.Fatalf("GetCustomization returned an unexpected error: %s", err)
				}
				if strings.Contains(err.Error(), testCase.expectedError) == false {
					t.Fatalf("GetCustomization returned an unexpected error: %s", err)
				}
			} else {
				if reflect.DeepEqual(testCase.expected, got) == false {
					t.Errorf("GetCustomization returned unexpected body: got %v;"+
						" expected %v", got, testCase.expected)
				}
			}
		})
	}
}

func TestGetAllCustomization(t *testing.T) {
	testCases := []struct {
		label                        string
		inputProject                 string
		inputCompositeApp            string
		inputCompositeAppVersion     string
		inputDeploymentIntentGrpName string
		inputGenericK8sIntent        string
		inputResource                string
		inputExists                  bool
		expectedError                string
		mockdb                       *db.MockDB
		expected                     []Customization
	}{
		{
			label:                        "Get All customizations",
			inputProject:                 "testProject",
			inputCompositeApp:            "testCompositeApp",
			inputCompositeAppVersion:     "testCompositeAppVersion",
			inputDeploymentIntentGrpName: "testDeploymentIntentGroup",
			inputGenericK8sIntent:        "testGenK8sIntent",
			inputResource:                "testResource",
			expected: []Customization{
				{
					Metadata: Metadata{
						Name:        "testCustomization",
						Description: "testCustomization",
						UserData1:   "userData1",
						UserData2:   "userData2",
					},
					Spec: CustomizeSpec{
						ClusterSpecific: "true",
						ClusterInfo: ClusterInfo{
							Scope:           "label",
							ClusterProvider: "testClusterProvider",
							ClusterName:     "",
							ClusterLabel:    "testLabel",
							Mode:            "allow",
						},
						PatchType: "json",
						PatchJSON: []map[string]interface{}{
							{
								"op":    "replace",
								"path":  "/name",
								"value": "test",
							},
						},
					},
				},
			},
			expectedError: "",
			mockdb: &db.MockDB{
				Items: []map[string]map[string][]byte{
					{
						CustomizationKey{
							Customization:       "",
							Project:             "testProject",
							CompositeApp:        "testCompositeApp",
							CompositeAppVersion: "testCompositeAppVersion",
							DigName:             "testDeploymentIntentGroup",
							GenericK8sIntent:    "testGenK8sIntent",
							Resource:            "testResource",
						}.String(): {
							"customizationmetadata": []byte(
								"{\"metadata\":{\"name\":\"testCustomization\",\"description\":\"testCustomization\",\"userData1\":\"userData1\",\"userData2\":\"userData2\"},\"spec\":{\"clusterSpecific\":\"true\",\"clusterInfo\":{\"scope\":\"label\",\"clusterProvider\":\"testClusterProvider\",\"clusterName\":\"\",\"clusterLabel\":\"testLabel\",\"mode\":\"allow\"},\"patchType\":\"json\",\"patchJson\":[{\"op\":\"replace\",\"path\":\"\\/name\",\"value\":\"test\"}]}}"),
						},
					},
				},
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.label, func(t *testing.T) {
			db.DBconn = testCase.mockdb
			customCli := NewCustomizationClient()
			got, err := customCli.GetAllCustomization(testCase.inputProject, testCase.inputCompositeApp, testCase.inputCompositeAppVersion, testCase.inputDeploymentIntentGrpName, testCase.inputGenericK8sIntent, testCase.inputResource)
			if err != nil {
				if testCase.expectedError == "" {
					t.Fatalf("GetAllCustomization returned an unexpected error: %s", err)
				}
				if strings.Contains(err.Error(), testCase.expectedError) == false {
					t.Fatalf("GetAllCustomization returned an unexpected error: %s", err)
				}
			} else {
				if reflect.DeepEqual(testCase.expected, got) == false {
					t.Errorf("GetAllCustomization returned unexpected body: got %v;"+
						" expected %v", got, testCase.expected)
				}
			}
		})
	}
}

func TestGetCustomizationContent(t *testing.T) {
	testCases := []struct {
		label                        string
		inputCustomization           string
		inputProject                 string
		inputCompositeApp            string
		inputCompositeAppVersion     string
		inputDeploymentIntentGrpName string
		inputGenericK8sIntent        string
		inputResourceName            string
		expectedError                string
		mockdb                       *db.MockDB
		expected                     SpecFileContent
	}{
		{
			label:                        "Get ResourceContent",
			inputCustomization:           "testCustomization",
			inputProject:                 "testProject",
			inputCompositeApp:            "testCompositeApp",
			inputCompositeAppVersion:     "testCompositeAppVersion",
			inputDeploymentIntentGrpName: "testDeploymentIntentGroup",
			inputGenericK8sIntent:        "testGenK8sIntent",
			inputResourceName:            "testResource",
			expected: SpecFileContent{
				FileContents: []string{
					"This is testFile1",
					"This is testFile2",
				},
				FileNames: []string{
					"testFile1",
					"testFile2",
				},
			},
			expectedError: "",
			mockdb: &db.MockDB{
				Items: []map[string]map[string][]byte{
					{
						CustomizationKey{
							Customization:       "testCustomization",
							Project:             "testProject",
							CompositeApp:        "testCompositeApp",
							CompositeAppVersion: "testCompositeAppVersion",
							DigName:             "testDeploymentIntentGroup",
							GenericK8sIntent:    "testGenK8sIntent",
							Resource:            "testResource",
						}.String(): {
							"customizationmetadata": []byte(
								"{\"metadata\":{\"name\":\"testCustomization\",\"description\":\"testCustomization\",\"userData1\":\"userData1\",\"userData2\":\"userData2\"},\"spec\":{\"clusterSpecific\":\"true\",\"clusterInfo\":{\"scope\":\"label\",\"clusterProvider\":\"testClusterProvider\",\"clusterName\":\"\",\"clusterLabel\":\"testLabel\",\"mode\":\"allow\"},\"patchType\":\"json\",\"patchJson\":[{\"op\":\"replace\",\"path\":\"\\/name\",\"value\":\"test\"}]}}"),
							"customizationcontent": []byte(
							"{\"FileContents\":[\"This is testFile1\",\"This is testFile2\"],\"FileNames\":[\"testFile1\",\"testFile2\"]}"),
						},
					},
				},
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.label, func(t *testing.T) {
			db.DBconn = testCase.mockdb
			customCli := NewCustomizationClient()
			got, err := customCli.GetCustomizationContent(testCase.inputCustomization, testCase.inputProject, testCase.inputCompositeApp, testCase.inputCompositeAppVersion, testCase.inputDeploymentIntentGrpName, testCase.inputGenericK8sIntent, testCase.inputResourceName)
			if err != nil {
				if testCase.expectedError == "" {
					t.Fatalf("GetCustomizationContent returned an unexpected error: %s", err)
				}
				if strings.Contains(err.Error(), testCase.expectedError) == false {
					t.Fatalf("GetCustomizationContent returned an unexpected error: %s", err)
				}
			} else {
				if reflect.DeepEqual(testCase.expected, got) == false {
					t.Errorf("GetCustomizationContent returned unexpected body: got %v;"+
						" expected %v", got, testCase.expected)
				}
			}
		})
	}
}

func TestDeleteCustomization(t *testing.T) {
	testCases := []struct {
		label                        string
		inputCustomization           string
		inputProject                 string
		inputCompositeApp            string
		inputCompositeAppVersion     string
		inputDeploymentIntentGrpName string
		inputGenericK8sIntent        string
		inputResource                string
		expectedError                string
		mockdb                       *db.MockDB
	}{
		{
			label:                        "Delete customization",
			inputCustomization:           "testCustomization",
			inputProject:                 "testProject",
			inputCompositeApp:            "testCompositeApp",
			inputCompositeAppVersion:     "testCompositeAppVersion",
			inputDeploymentIntentGrpName: "testDeploymentIntentGroup",
			inputGenericK8sIntent:        "testGenK8sIntent",
			inputResource:                "testResource",
			expectedError:                "",
			mockdb: &db.MockDB{
				Items: []map[string]map[string][]byte{
					{
						CustomizationKey{
							Customization:       "testCustomization",
							Project:             "testProject",
							CompositeApp:        "testCompositeApp",
							CompositeAppVersion: "testCompositeAppVersion",
							DigName:             "testDeploymentIntentGroup",
							GenericK8sIntent:    "testGenK8sIntent",
							Resource:            "testResource",
						}.String(): {
							"customizationmetadata": []byte(
								"{\"metadata\":{\"name\":\"testCustomization\",\"description\":\"testCustomization\",\"userData1\":\"userData1\",\"userData2\":\"userData2\"},\"spec\":{\"clusterSpecific\":\"true\",\"clusterInfo\":{\"scope\":\"label\",\"clusterProvider\":\"testClusterProvider\",\"clusterName\":\"\",\"clusterLabel\":\"testLabel\",\"mode\":\"allow\"},\"patchType\":\"json\",\"patchJson\":[{\"op\":\"replace\",\"path\":\"\\/name\",\"value\":\"test\"}]}}"),
						},
					},
				},
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.label, func(t *testing.T) {
			db.DBconn = testCase.mockdb
			customCli := NewCustomizationClient()
			err := customCli.DeleteCustomization(testCase.inputCustomization, testCase.inputProject, testCase.inputCompositeApp, testCase.inputCompositeAppVersion, testCase.inputDeploymentIntentGrpName, testCase.inputGenericK8sIntent, testCase.inputResource)
			if err != nil {
				if testCase.expectedError == "" {
					t.Fatalf("DeleteCustomization returned an unexpected error: %s", err)
				}
				if strings.Contains(err.Error(), testCase.expectedError) == false {
					t.Fatalf("DeleteCustomization returned an unexpected error: %s", err)
				}
			}
		})
	}
}
