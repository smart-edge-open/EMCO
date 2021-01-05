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

func TestCreateResource(t *testing.T) {
	testCases := []struct {
		label                        string
		inputResource                Resource
		inputResourceContent         ResourceFileContent
		inputProject                 string
		inputCompositeApp            string
		inputCompositeAppVersion     string
		inputDeploymentIntentGrpName string
		inputGenericK8sIntent        string
		inputExists                  bool
		expectedError                string
		mockdb                       *db.MockDB
		expected                     Resource
	}{
		{
			label: "Create Resource",
			inputResource: Resource{
				Metadata: Metadata{
					Name:        "testResource",
					Description: "testResource",
					UserData1:   "userData1",
					UserData2:   "userData2",
				},
				Spec: ResourceSpec{
					AppName:   "testApp",
					NewObject: "True",
					ResourceGVK: ResourceGVK{
						APIVersion: "v1",
						Kind:       "configMap",
						Name:       "TestCM",
					},
				},
			},
			inputResourceContent: ResourceFileContent{
				FileContent: "TestResourceContent",
			},
			inputProject:                 "testProject",
			inputCompositeApp:            "testCompositeApp",
			inputCompositeAppVersion:     "testCompositeAppVersion",
			inputDeploymentIntentGrpName: "testDeploymentIntentGroup",
			inputGenericK8sIntent:        "testGenK8sIntent",
			inputExists:                  false,
			expected: Resource{
				Metadata: Metadata{
					Name:        "testResource",
					Description: "testResource",
					UserData1:   "userData1",
					UserData2:   "userData2",
				},
				Spec: ResourceSpec{
					AppName:   "testApp",
					NewObject: "True",
					ResourceGVK: ResourceGVK{
						APIVersion: "v1",
						Kind:       "configMap",
						Name:       "TestCM",
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
					},
				},
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.label, func(t *testing.T) {
			db.DBconn = testCase.mockdb
			resourceCli := NewResourceClient()
			got, err := resourceCli.CreateResource(testCase.inputResource, testCase.inputResourceContent, testCase.inputProject, testCase.inputCompositeApp, testCase.inputCompositeAppVersion, testCase.inputDeploymentIntentGrpName, testCase.inputGenericK8sIntent, testCase.inputExists)
			if err != nil {
				if testCase.expectedError == "" {
					t.Fatalf("CreateResource returned an unexpected error: %s", err)
				}
				if strings.Contains(err.Error(), testCase.expectedError) == false {
					t.Fatalf("CreateResource returned an unexpected error: %s", err)
				}
			} else {
				if reflect.DeepEqual(testCase.expected, got) == false {
					t.Errorf("CreateResource returned unexpected body: got %v;"+
						" expected %v", got, testCase.expected)
				}
			}
		})
	}

}

func TestGetResource(t *testing.T) {
	testCases := []struct {
		label                        string
		inputResourceName            string
		inputProject                 string
		inputCompositeApp            string
		inputCompositeAppVersion     string
		inputDeploymentIntentGrpName string
		inputGenericK8sIntent        string
		expectedError                string
		mockdb                       *db.MockDB
		expected                     Resource
	}{
		{
			label:                        "Get All resources",
			inputResourceName:            "testResource",
			inputProject:                 "testProject",
			inputCompositeApp:            "testCompositeApp",
			inputCompositeAppVersion:     "testCompositeAppVersion",
			inputDeploymentIntentGrpName: "testDeploymentIntentGroup",
			inputGenericK8sIntent:        "testGenK8sIntent",
			expected: Resource{

				Metadata: Metadata{
					Name:        "testResource",
					Description: "testResource",
					UserData1:   "userData1",
					UserData2:   "userData2",
				},
				Spec: ResourceSpec{
					AppName:   "testApp",
					NewObject: "True",
					ResourceGVK: ResourceGVK{
						APIVersion: "v1",
						Kind:       "configMap",
						Name:       "TestCM",
					},
				},
			},
			expectedError: "",
			mockdb: &db.MockDB{
				Items: []map[string]map[string][]byte{
					{
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
			resourceCli := NewResourceClient()
			got, err := resourceCli.GetResource(testCase.inputResourceName, testCase.inputProject, testCase.inputCompositeApp, testCase.inputCompositeAppVersion, testCase.inputDeploymentIntentGrpName, testCase.inputGenericK8sIntent)
			if err != nil {
				if testCase.expectedError == "" {
					t.Fatalf("GetResource returned an unexpected error: %s", err)
				}
				if strings.Contains(err.Error(), testCase.expectedError) == false {
					t.Fatalf("GetResource returned an unexpected error: %s", err)
				}
			} else {
				if reflect.DeepEqual(testCase.expected, got) == false {
					t.Errorf("GetResource returned unexpected body: got %v;"+
						" expected %v", got, testCase.expected)
				}
			}
		})
	}
}

func TestGetAllResources(t *testing.T) {
	testCases := []struct {
		label                        string
		inputProject                 string
		inputCompositeApp            string
		inputCompositeAppVersion     string
		inputDeploymentIntentGrpName string
		inputGenericK8sIntent        string
		expectedError                string
		mockdb                       *db.MockDB
		expected                     []Resource
	}{
		{
			label:                        "Get All resources",
			inputProject:                 "testProject",
			inputCompositeApp:            "testCompositeApp",
			inputCompositeAppVersion:     "testCompositeAppVersion",
			inputDeploymentIntentGrpName: "testDeploymentIntentGroup",
			inputGenericK8sIntent:        "testGenK8sIntent",
			expected: []Resource{
				{
					Metadata: Metadata{
						Name:        "testResource",
						Description: "testResource",
						UserData1:   "userData1",
						UserData2:   "userData2",
					},
					Spec: ResourceSpec{
						AppName:   "testApp",
						NewObject: "True",
						ResourceGVK: ResourceGVK{
							APIVersion: "v1",
							Kind:       "configMap",
							Name:       "TestCM",
						},
					},
				},
			},
			expectedError: "",
			mockdb: &db.MockDB{
				Items: []map[string]map[string][]byte{
					{
						ResourceKey{
							Resource:            "",
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
			resourceCli := NewResourceClient()
			got, err := resourceCli.GetAllResources(testCase.inputProject, testCase.inputCompositeApp, testCase.inputCompositeAppVersion, testCase.inputDeploymentIntentGrpName, testCase.inputGenericK8sIntent)
			if err != nil {
				if testCase.expectedError == "" {
					t.Fatalf("GetAllResources returned an unexpected error: %s", err)
				}
				if strings.Contains(err.Error(), testCase.expectedError) == false {
					t.Fatalf("GetAllResources returned an unexpected error: %s", err)
				}
			} else {
				if reflect.DeepEqual(testCase.expected, got) == false {
					t.Errorf("GetAllResources returned unexpected body: got %v;"+
						" expected %v", got, testCase.expected)
				}
			}
		})
	}
}

func TestGetResourceContent(t *testing.T) {
	testCases := []struct {
		label                        string
		inputResourceName            string
		inputProject                 string
		inputCompositeApp            string
		inputCompositeAppVersion     string
		inputDeploymentIntentGrpName string
		inputGenericK8sIntent        string
		expectedError                string
		mockdb                       *db.MockDB
		expected                     ResourceFileContent
	}{
		{
			label:                        "Get ResourceContent",
			inputResourceName:            "testResource",
			inputProject:                 "testProject",
			inputCompositeApp:            "testCompositeApp",
			inputCompositeAppVersion:     "testCompositeAppVersion",
			inputDeploymentIntentGrpName: "testDeploymentIntentGroup",
			inputGenericK8sIntent:        "testGenK8sIntent",
			expected: ResourceFileContent{
				FileContent: "testFileContent",
			},
			expectedError: "",
			mockdb: &db.MockDB{
				Items: []map[string]map[string][]byte{
					{
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
							"resourcecontent": []byte(
								"{\"FileContent\":\"testFileContent\"}"),
						},
					},
				},
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.label, func(t *testing.T) {
			db.DBconn = testCase.mockdb
			resourceCli := NewResourceClient()
			got, err := resourceCli.GetResourceContent(testCase.inputResourceName, testCase.inputProject, testCase.inputCompositeApp, testCase.inputCompositeAppVersion, testCase.inputDeploymentIntentGrpName, testCase.inputGenericK8sIntent)
			if err != nil {
				if testCase.expectedError == "" {
					t.Fatalf("GetResourceContent returned an unexpected error: %s", err)
				}
				if strings.Contains(err.Error(), testCase.expectedError) == false {
					t.Fatalf("GetResourceContent returned an unexpected error: %s", err)
				}
			} else {
				if reflect.DeepEqual(testCase.expected, got) == false {
					t.Errorf("GetResourceContent returned unexpected body: got %v;"+
						" expected %v", got, testCase.expected)
				}
			}
		})
	}
}

func TestDeleteResource(t *testing.T) {
	testCases := []struct {
		label                        string
		inputResourceName            string
		inputProject                 string
		inputCompositeApp            string
		inputCompositeAppVersion     string
		inputDeploymentIntentGrpName string
		inputGenericK8sIntent        string
		expectedError                string
		mockdb                       *db.MockDB
	}{
		{
			label:                        "Delete resource",
			inputResourceName:            "testResource",
			inputProject:                 "testProject",
			inputCompositeApp:            "testCompositeApp",
			inputCompositeAppVersion:     "testCompositeAppVersion",
			inputDeploymentIntentGrpName: "testDeploymentIntentGroup",
			inputGenericK8sIntent:        "testGenK8sIntent",
			expectedError:                "",
			mockdb: &db.MockDB{
				Items: []map[string]map[string][]byte{
					{
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
			resourceCli := NewResourceClient()
			err := resourceCli.DeleteResource(testCase.inputResourceName, testCase.inputProject, testCase.inputCompositeApp, testCase.inputCompositeAppVersion, testCase.inputDeploymentIntentGrpName, testCase.inputGenericK8sIntent)
			if err != nil {
				if testCase.expectedError == "" {
					t.Fatalf("DeleteResource returned an unexpected error: %s", err)
				}
				if strings.Contains(err.Error(), testCase.expectedError) == false {
					t.Fatalf("DeleteResource returned an unexpected error: %s", err)
				}
			}
		})
	}
}
