// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2020 Intel Corporation

package module

import (
	"reflect"
	"strings"
	"testing"

	"github.com/open-ness/EMCO/src/orchestrator/pkg/infra/db"
)


func TestCreateCompositeProfile(t *testing.T) {
	testCases := []struct {
		label               string
		compositeProfile    CompositeProfile
		projectName         string
		compositeApp        string
		compositeAppVersion string
		expectedError       string
		mockdb              *db.MockDB
		expected            CompositeProfile
	}{
		{
			label: "Create CompositeProfile",
			compositeProfile: CompositeProfile{
				Metadata: CompositeProfileMetadata{
					Name:        "testCompositeProfile",
					Description: "A sample Composite Profile for testing",
					UserData1:   "userData1",
					UserData2:   "userData2",
				},
			},
			projectName:         "testProject",
			compositeApp:        "testCompositeApp",
			compositeAppVersion: "v1",
			expected: CompositeProfile{
				Metadata: CompositeProfileMetadata{
					Name:        "testCompositeProfile",
					Description: "A sample Composite Profile for testing",
					UserData1:   "userData1",
					UserData2:   "userData2",
				},
			},
			expectedError: "",
			mockdb: &db.MockDB{
				Items: []map[string]map[string][]byte{{
					ProjectKey{ProjectName: "testProject"}.String(): {
						"projectmetadata": []byte(
							"{" +
								"\"metadata\" : {" +
								"\"Name\":\"testProject\"," +
								"\"Description\":\"Test project for unit testing\"," +
								"\"UserData1\": \"userData1\"," +
								"\"UserData2\":\"userData2\"}" +
								"}"),
					},
					CompositeAppKey{CompositeAppName: "testCompositeApp", Project: "testProject", Version: "v1"}.String(): {
						"compositeappmetadata": []byte(
							"{" +
								"\"metadata\" : {" +
								"\"Name\":\"testCompositeApp\"," +
								"\"Description\":\"Test Composite App for unit testing\"," +
								"\"UserData1\": \"userData1\"," +
								"\"UserData2\":\"userData2\"}," +
								"\"spec\": {" +
								"\"Version\": \"v1\"}" +
								"}"),
					},
				},},
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.label, func(t *testing.T) {
			db.DBconn = testCase.mockdb
			cprofCli := NewCompositeProfileClient()
			got, err := cprofCli.CreateCompositeProfile(testCase.compositeProfile, testCase.projectName, testCase.compositeApp, testCase.compositeAppVersion, false)
			if err != nil {
				if testCase.expectedError == "" {
					t.Fatalf("CreateCompositeProfile returned an unexpected error %s", err)
				}
				if strings.Contains(err.Error(), testCase.expectedError) == false {
					t.Fatalf("CreateCompositeProfile returned an unexpected error %s", err)
				}
			} else {
				if reflect.DeepEqual(testCase.expected, got) == false {
					t.Errorf("CreateCompositeProfile returned unexpected body: got %v; "+" expected %v", got, testCase.expected)
				}
			}
		})

	}
}

func TestGetCompositeProfile(t *testing.T) {

	testCases := []struct {
		label                string
		expectedError        string
		expected             CompositeProfile
		mockdb               *db.MockDB
		compositeProfileName string
		projectName          string
		compositeAppName     string
		compositeAppVersion  string
	}{
		{
			label:                "Get CompositeProfile",
			compositeProfileName: "testCompositeProfile",
			projectName:          "testProject",
			compositeAppName:     "testCompositeApp",
			compositeAppVersion:  "v1",
			expected: CompositeProfile{
				Metadata: CompositeProfileMetadata{
					Name:        "testCompositeProfile",
					Description: "A sample CompositeProfile for testing",
					UserData1:   "userData1",
					UserData2:   "userData2",
				},
			},
			expectedError: "",
			mockdb: &db.MockDB{
				Items: []map[string]map[string][]byte{{
					CompositeProfileKey{
						Name:         "testCompositeProfile",
						Project:      "testProject",
						CompositeApp: "testCompositeApp",
						Version:      "v1",
					}.String(): {
						"compositeprofilemetadata": []byte(
							"{\"metadata\":{\"Name\":\"testCompositeProfile\"," +
								"\"Description\":\"A sample CompositeProfile for testing\"," +
								"\"UserData1\": \"userData1\"," +
								"\"UserData2\": \"userData2\"}}"),
					},
				},},
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.label, func(t *testing.T) {
			db.DBconn = testCase.mockdb
			cprofCli := NewCompositeProfileClient()
			got, err := cprofCli.GetCompositeProfile(testCase.compositeProfileName, testCase.projectName, testCase.compositeAppName, testCase.compositeAppVersion)
			if err != nil {
				if testCase.expectedError == "" {
					t.Fatalf("GetCompositeProfile returned an unexpected error: %s", err)
				}
				if strings.Contains(err.Error(), testCase.expectedError) == false {
					t.Fatalf("GetCompositeProfile returned an unexpected error: %s", err)
				}
			} else {
				if reflect.DeepEqual(testCase.expected, got) == false {
					t.Errorf("GetCompositeProfile returned unexpected body: got %v;"+
						" expected %v", got, testCase.expected)
				}
			}

		})
	}

}

