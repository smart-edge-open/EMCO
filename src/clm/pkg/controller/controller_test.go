// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2020 Intel Corporation

package controller

import (
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/open-ness/EMCO/src/orchestrator/pkg/infra/db"
	"github.com/open-ness/EMCO/src/orchestrator/pkg/module/types"

	clmModel "github.com/open-ness/EMCO/src/clm/pkg/model"
	pkgerrors "github.com/pkg/errors"
)

func TestCreateController(t *testing.T) {
	testCases := []struct {
		label         string
		inp           clmModel.Controller
		expectedError string
		mockdb        *db.MockDB
		expected      clmModel.Controller
	}{
		{
			label: "Create Controller",
			inp: clmModel.Controller{
				Metadata: types.Metadata{
					Name: "testController",
				},
				Spec: clmModel.ControllerSpec{
					Host: "132.156.0.10",
					Port: 8080,
				},
			},
			expected: clmModel.Controller{
				Metadata: types.Metadata{
					Name: "testController",
				},
				Spec: clmModel.ControllerSpec{
					Host: "132.156.0.10",
					Port: 8080,
				},
			},
			expectedError: "",
			mockdb:        &db.MockDB{},
		},
		{
			label:         "Failed Create Controller",
			expectedError: "Error Creating Controller",
			mockdb: &db.MockDB{
				Err: pkgerrors.New("Error Creating Controller"),
			},
		},
	}

	fmt.Printf("\n================== TestCreateController .. total_testcase_count[%d] ==================\n", len(testCases))
	for i, testCase := range testCases {
		t.Run(testCase.label, func(t *testing.T) {
			fmt.Printf("\n================== TestCreateController .. testcase_count[%d] testcase_name[%s] ==================\n", i, testCase.label)
			db.DBconn = testCase.mockdb
			impl := NewControllerClient()
			got, err := impl.CreateController(testCase.inp, false)
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

func TestGetController(t *testing.T) {

	testCases := []struct {
		label         string
		name          string
		expectedError string
		mockdb        *db.MockDB
		inp           string
		expected      clmModel.Controller
	}{
		{
			label:         "Get Error",
			expectedError: "DB Error",
			mockdb: &db.MockDB{
				Err: pkgerrors.New("DB Error"),
			},
		},
		{
			label: "Get Controller",
			name:  "testController",
			expected: clmModel.Controller{
				Metadata: types.Metadata{
					Name: "testController",
				},
				Spec: clmModel.ControllerSpec{
					Host: "132.156.0.10",
					Port: 8080,
				},
			},
			expectedError: "",
			mockdb: &db.MockDB{
				Items: []map[string]map[string][]byte{
					{
						clmModel.ControllerKey{ControllerName: "testController"}.String(): {
							"controllermetadata": []byte(
								"{\"metadata\":{" +
									"\"name\":\"testController\"" +
									"}," +
									"\"spec\":{" +
									"\"host\":\"132.156.0.10\"," +
									"\"port\": 8080 }}"),
						},
					},
				},
			},
		},
	}

	fmt.Printf("\n================== TestGetController .. total_testcase_count[%d] ==================\n", len(testCases))
	for i, testCase := range testCases {
		fmt.Printf("\n================== TestGetController .. testcase_count[%d] testcase_name[%s] ==================\n", i, testCase.label)
		t.Run(testCase.label, func(t *testing.T) {
			db.DBconn = testCase.mockdb
			impl := NewControllerClient()
			got, err := impl.GetController(testCase.name)
			if err != nil {
				if testCase.expectedError == "" {
					t.Fatalf("Get returned an unexpected-error[%s] expected[%s]", err, testCase.expectedError)
				}
				if strings.Contains(strings.ToLower(err.Error()), strings.ToLower(testCase.expectedError)) == false {
					t.Fatalf("Get returned an unexpected-error[%s] expected[%s]", err, testCase.expectedError)
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

func TestDeleteController(t *testing.T) {

	testCases := []struct {
		label         string
		name          string
		expectedError string
		mockdb        *db.MockDB
	}{
		{
			label:  "Delete Controller",
			name:   "testController",
			mockdb: &db.MockDB{},
		},
		{
			label:         "Delete Error",
			expectedError: "DB Error",
			mockdb: &db.MockDB{
				Err: pkgerrors.New("DB Error"),
			},
		},
	}

	fmt.Printf("\n================== TestDeleteController .. total_testcase_count[%d] ==================\n", len(testCases))
	for i, testCase := range testCases {
		fmt.Printf("\n================== TestDeleteController .. testcase_count[%d] testcase_name[%s] ==================\n", i, testCase.label)
		t.Run(testCase.label, func(t *testing.T) {
			db.DBconn = testCase.mockdb
			impl := NewControllerClient()
			err := impl.DeleteController(testCase.name)
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
