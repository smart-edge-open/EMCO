package module

import (
	"github.com/open-ness/EMCO/src/orchestrator/pkg/infra/db"
	moduleLib "github.com/open-ness/EMCO/src/orchestrator/pkg/module"
	"reflect"
	"strings"
	"testing"
)

func TestCreateGenericK8sIntent(t *testing.T) {
	testCases := []struct {
		label                        string
		inputGenericK8sIntent        GenericK8sIntent
		inputProject                 string
		inputCompositeApp            string
		inputCompositeAppVersion     string
		inputDeploymentIntentGrpName string
		inputExists                  bool
		expectedError                string
		mockdb                       *db.MockDB
		expected                     GenericK8sIntent
	}{
		{
			label: "Create GenericK8sIntent",
			inputGenericK8sIntent: GenericK8sIntent{
				Metadata: Metadata{
					Name:        "testGenK8sIntent",
					Description: "testGenK8sIntent",
					UserData1:   "userData1",
					UserData2:   "userData2",
				},
			},
			inputProject:                 "testProject",
			inputCompositeApp:            "testCompositeApp",
			inputCompositeAppVersion:     "testCompositeAppVersion",
			inputDeploymentIntentGrpName: "testDeploymentIntentGroup",
			inputExists:                  false,
			expected: GenericK8sIntent{
				Metadata: Metadata{
					Name:        "testGenK8sIntent",
					Description: "testGenK8sIntent",
					UserData1:   "userData1",
					UserData2:   "userData2",
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
					},
				},
			},
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.label, func(t *testing.T) {
			db.DBconn = testCase.mockdb
			genK8sCli := NewGenericK8sIntentClient()
			got, err := genK8sCli.CreateGenericK8sIntent(testCase.inputGenericK8sIntent, testCase.inputProject, testCase.inputCompositeApp, testCase.inputCompositeAppVersion, testCase.inputDeploymentIntentGrpName, testCase.inputExists)
			if err != nil {
				if testCase.expectedError == "" {
					t.Fatalf("CreateGenericK8sIntent returned an unexpected error %s, ", err)
				}
				if strings.Contains(err.Error(), testCase.expectedError) == false {
					t.Fatalf("CreateGenericK8sIntent returned an unexpected error %s", err)
				}
			} else {
				if reflect.DeepEqual(testCase.expected, got) == false {
					t.Errorf("CreateGenericK8sIntent returned unexpected body: got %v; "+" expected %v", got, testCase.expected)
				}
			}
		})
	}
}

func TestGetGenericK8sIntent(t *testing.T) {
	testCases := []struct {
		label                        string
		inputGenericK8sIntent        string
		inputProject                 string
		inputCompositeApp            string
		inputCompositeAppVersion     string
		inputDeploymentIntentGrpName string
		expectedError                string
		mockdb                       *db.MockDB
		expected                     GenericK8sIntent
	}{
		{
			label:                        "Get Intent",
			inputGenericK8sIntent:        "testGenK8sIntent",
			inputProject:                 "testProject",
			inputCompositeApp:            "testCompositeApp",
			inputCompositeAppVersion:     "testCompositeAppVersion",
			inputDeploymentIntentGrpName: "testDeploymentIntentGroup",
			expected: GenericK8sIntent{
				Metadata: Metadata{
					Name:        "testGenK8sIntent",
					Description: "testGenK8sIntent",
					UserData1:   "userData1",
					UserData2:   "userData2",
				},
			},
			expectedError: "",
			mockdb: &db.MockDB{
				Items: []map[string]map[string][]byte{
					{
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
			genK8sCli := NewGenericK8sIntentClient()
			got, err := genK8sCli.GetGenericK8sIntent(testCase.inputGenericK8sIntent, testCase.inputProject, testCase.inputCompositeApp, testCase.inputCompositeAppVersion, testCase.inputDeploymentIntentGrpName)
			if err != nil {
				if testCase.expectedError == "" {
					t.Fatalf("GetGenericK8sIntent returned an unexpected error: %s", err)
				}
				if strings.Contains(err.Error(), testCase.expectedError) == false {
					t.Fatalf("GetGenericK8sIntent returned an unexpected error: %s", err)
				}
			} else {
				if reflect.DeepEqual(testCase.expected, got) == false {
					t.Errorf("GetGenericK8sIntent returned unexpected body: got %v;"+
						" expected %v", got, testCase.expected)
				}
			}
		})
	}
}

func TestGetAllGenericK8sIntents(t *testing.T) {
	testCases := []struct {
		label                        string
		inputGenericK8sIntent        string
		inputProject                 string
		inputCompositeApp            string
		inputCompositeAppVersion     string
		inputDeploymentIntentGrpName string
		expectedError                string
		mockdb                       *db.MockDB
		expected                     []GenericK8sIntent
	}{
		{
			label:                        "Get All Intents",
			inputProject:                 "testProject",
			inputCompositeApp:            "testCompositeApp",
			inputCompositeAppVersion:     "testCompositeAppVersion",
			inputDeploymentIntentGrpName: "testDeploymentIntentGroup",
			expected: []GenericK8sIntent{
				{
					Metadata: Metadata{
						Name:        "testGenK8sIntent",
						Description: "testGenK8sIntent",
						UserData1:   "userData1",
						UserData2:   "userData2",
					},
				},
			},
			expectedError: "",
			mockdb: &db.MockDB{
				Items: []map[string]map[string][]byte{
					{
						GenericK8sIntentKey{
							GenericK8sIntent:    "",
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
			genK8sCli := NewGenericK8sIntentClient()
			got, err := genK8sCli.GetAllGenericK8sIntents(testCase.inputProject, testCase.inputCompositeApp, testCase.inputCompositeAppVersion, testCase.inputDeploymentIntentGrpName)
			if err != nil {
				if testCase.expectedError == "" {
					t.Fatalf("GetGenericK8sIntent returned an unexpected error: %s", err)
				}
				if strings.Contains(err.Error(), testCase.expectedError) == false {
					t.Fatalf("GetGenericK8sIntent returned an unexpected error: %s", err)
				}
			} else {
				if reflect.DeepEqual(testCase.expected, got) == false {
					t.Errorf("GetGenericK8sIntent returned unexpected body: got %v;"+
						" expected %v", got, testCase.expected)
				}
			}
		})
	}
}

func TestDeleteGenericK8sIntent(t *testing.T) {
	testCases := []struct {
		label                        string
		inputGenericK8sIntent        string
		inputProject                 string
		inputCompositeApp            string
		inputCompositeAppVersion     string
		inputDeploymentIntentGrpName string
		expectedError                string
		mockdb                       *db.MockDB
	}{
		{
			label:                        "Delete Intent",
			inputGenericK8sIntent:        "testGenK8sIntent",
			inputProject:                 "testProject",
			inputCompositeApp:            "testCompositeApp",
			inputCompositeAppVersion:     "testCompositeAppVersion",
			inputDeploymentIntentGrpName: "testDeploymentIntentGroup",
			expectedError:                "",
			mockdb: &db.MockDB{
				Items: []map[string]map[string][]byte{
					{
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
			genK8sCli := NewGenericK8sIntentClient()
			err := genK8sCli.DeleteGenericK8sIntent(testCase.inputGenericK8sIntent, testCase.inputProject, testCase.inputCompositeApp, testCase.inputCompositeAppVersion, testCase.inputDeploymentIntentGrpName)
			if err != nil {
				if testCase.expectedError == "" {
					t.Fatalf("DeleteGenericK8sIntent returned an unexpected error: %s", err)
				}
				if strings.Contains(err.Error(), testCase.expectedError) == false {
					t.Fatalf("DeleteGenericK8sIntent returned an unexpected error: %s", err)
				}
			}
		})
	}
}
