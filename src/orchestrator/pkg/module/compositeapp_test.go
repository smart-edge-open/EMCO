// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2020 Intel Corporation

package module

import (
	"reflect"
	"strings"
	"testing"

	"github.com/open-ness/EMCO/src/orchestrator/pkg/infra/db"

	pkgerrors "github.com/pkg/errors"
)

func TestCreateCompositeApp(t *testing.T) {
	testCases := []struct {
		label         string
		inpCompApp    CompositeApp
		inpProject    string
		expectedError string
		mockdb        *db.MockDB
		expected      CompositeApp
	}{
		{
			label: "Create Composite App",
			inpCompApp: CompositeApp{
				Metadata: CompositeAppMetaData{
					Name:        "testCompositeApp",
					Description: "A sample composite app used for unit testing",
					UserData1:   "userData1",
					UserData2:   "userData2",
				},
				Spec: CompositeAppSpec{
					Version: "v1",
				},
			},

			inpProject: "testProject",
			expected: CompositeApp{
				Metadata: CompositeAppMetaData{
					Name:        "testCompositeApp",
					Description: "A sample composite app used for unit testing",
					UserData1:   "userData1",
					UserData2:   "userData2",
				},
				Spec: CompositeAppSpec{
					Version: "v1",
				},
			},
			expectedError: "",
			mockdb: &db.MockDB{
				Items: []map[string]map[string][]byte{
					{
						ProjectKey{ProjectName: "testProject"}.String(): {
							"projectmetadata": []byte(
								"{" +
									"\"metadata\": {" +
									"\"Name\": \"testProject\"," +
									"\"Description\": \"Test project for unit testing\"," +
									"\"UserData1\": \"userData1\"," +
									"\"UserData2\": \"userData2\"}" +
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
			impl := NewCompositeAppClient()
			got, err := impl.CreateCompositeApp(testCase.inpCompApp, testCase.inpProject, false)
			if err != nil {
				if testCase.expectedError == "" {
					t.Fatalf("Create returned an unexpected error %s", err)
				}
				if strings.Contains(err.Error(), testCase.expectedError) == false {
					t.Fatalf("Create returned an unexpected error %s", err)
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

func TestGetCompositeApp(t *testing.T) {

	testCases := []struct {
		label         string
		inpName       string
		inpVersion    string
		inpProject    string
		expectedError string
		mockdb        *db.MockDB
		expected      CompositeApp
	}{
		{
			label:      "Get Composite App",
			inpName:    "testCompositeApp",
			inpVersion: "v1",
			inpProject: "testProject",
			expected: CompositeApp{
				Metadata: CompositeAppMetaData{
					Name:        "testCompositeApp",
					Description: "Test CompositeApp for unit testing",
					UserData1:   "userData1",
					UserData2:   "userData2",
				},
				Spec: CompositeAppSpec{
					Version: "v1",
				},
			},
			expectedError: "",
			mockdb: &db.MockDB{
				Items: []map[string]map[string][]byte{
					{
						CompositeAppKey{CompositeAppName: "testCompositeApp", Version: "v1", Project: "testProject"}.String(): {
							"compositeappmetadata": []byte(
								"{" +
									"\"metadata\":{" +
									"\"Name\":\"testCompositeApp\"," +
									"\"Description\":\"Test CompositeApp for unit testing\"," +
									"\"UserData1\":\"userData1\"," +
									"\"UserData2\":\"userData2\"}," +
									"\"spec\":{" +
									"\"Version\":\"v1\"}" +
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

	for _, testCase := range testCases {
		t.Run(testCase.label, func(t *testing.T) {
			db.DBconn = testCase.mockdb
			impl := NewCompositeAppClient()
			got, err := impl.GetCompositeApp(testCase.inpName, testCase.inpVersion, testCase.inpProject)
			if err != nil {
				if testCase.expectedError == "" {
					t.Fatalf("Get returned an unexpected error: %s", err)
				}
				if strings.Contains(err.Error(), testCase.expectedError) == false {
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

func TestDeleteCompositeApp(t *testing.T) {

	testCases := []struct {
		label         string
		inpName       string
		inpVersion    string
		inpProject    string
		expectedError string
		mockdb        *db.MockDB
	}{
		{
			label:      "Delete Composite app",
			inpName:    "testCompositeApp",
			inpVersion: "v1",
			inpProject: "testProject",
			mockdb: &db.MockDB{
				Items: []map[string]map[string][]byte{
					{
						CompositeAppKey{CompositeAppName: "testCompositeApp", Version: "v1", Project: "testProject"}.String(): {
							"compositeappmetadata": []byte(
								"{" +
									"\"metadata\":{" +
									"\"Name\":\"testCompositeApp\"," +
									"\"Description\":\"Test CompositeApp for unit testing\"," +
									"\"UserData1\":\"userData1\"," +
									"\"UserData2\":\"userData2\"}," +
									"\"spec\":{" +
									"\"Version\":\"v1\"}" +
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

	for _, testCase := range testCases {
		t.Run(testCase.label, func(t *testing.T) {
			db.DBconn = testCase.mockdb
			impl := NewCompositeAppClient()
			err := impl.DeleteCompositeApp(testCase.inpName, testCase.inpVersion, testCase.inpProject)
			if err != nil {
				if testCase.expectedError == "" {
					t.Fatalf("Delete returned an unexpected error %s", err)
				}
				if strings.Contains(err.Error(), testCase.expectedError) == false {
					t.Fatalf("Delete returned an unexpected error %s", err)
				}
			}
		})
	}
}
