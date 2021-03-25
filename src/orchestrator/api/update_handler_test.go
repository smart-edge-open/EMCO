// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2020 Intel Corporation

package api

import (
	"bytes"
	moduleLib "github.com/open-ness/EMCO/src/orchestrator/pkg/module"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

//Creating an embedded interface via anonymous variable
//This allows us to make mockDB satisfy the DatabaseConnection
//interface even if we are not implementing all the methods in it
type mockInstantiationManager struct {
	// Items and err will be used to customize each test
	// via a localized instantiation of mockInstantiationManager
	moduleLib.InstantiationClient
	Err   error
}

func (m mockInstantiationManager) Migrate(p string, ca string, v string, tCav string, di string, tDi string) error {
	if m.Err != nil {
		return m.Err
	}

	return  nil
}


func (m mockInstantiationManager) Update(p string, ca string, v string, di string) (int64, error) {
	if m.Err != nil {
		return -1,m.Err
	}

	return  0,nil
}

func (m mockInstantiationManager) Rollback(p string, ca string, v string, di string, rbRev string) error {
	if m.Err != nil {
		return m.Err
	}

	return  nil
}


func init() {
	migrateJSONFile = "../json-schemas/migrate.json"
	rollbackJSONFile = "../json-schemas/rollback.json"
}

func Test_updateHandler_migrate(t *testing.T) {
	testCases := []struct {
		label        string
		reader       io.Reader
		expectedCode int
		uClient  mockInstantiationManager
	}{
		{
			label: "Missing Body Failure",
			expectedCode: http.StatusBadRequest,
			reader: bytes.NewBuffer([]byte(``)),
			uClient:  mockInstantiationManager{},
		},
		{
			label: "Migrate Source DIG to Target DIG",
			expectedCode: http.StatusAccepted,
			reader: bytes.NewBuffer([]byte(`{
				"metadata" : {
    				"description": "Upgrade DIG1 from CA v1 to CA v3"
				},
				"spec" : {
					"target-composite-app-version": "v3",
					"target-dig-name": "test3"
				}
			}`)),
			uClient:  mockInstantiationManager{},
		},
		{
			label: "Missing target composite app version in Request Body",
			reader: bytes.NewBuffer([]byte(`{
				"metadata" : {
    				"description": "Upgrade DIG1 from CA v1 to CA v3"
				},
				"spec" : {
					"target-dig-name": "test3"
				}
			}`)),
			expectedCode: http.StatusBadRequest,
			uClient:  mockInstantiationManager{},
		},
		{
			label: "Missing target DIG name in Request Body",
			reader: bytes.NewBuffer([]byte(`{
				"metadata" : {
    				"description": "Upgrade DIG1 from CA v1 to CA v3"
				},
				"spec" : {
					"target-composite-app-version": "v3"
				}
			}`)),
			expectedCode: http.StatusBadRequest,
			uClient:  mockInstantiationManager{},
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.label, func(t *testing.T) {
			request := httptest.NewRequest("POST", "/v2/projects/{project-name}/composite-apps/{composite-app-name}/{composite-app-version}/deployment-intent-groups/{deployment-intent-group-name}/migrate", testCase.reader)
			resp := executeRequest(request, NewRouter(nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, testCase.uClient))

			//Check returned code
			if resp.StatusCode != testCase.expectedCode {
				t.Fatalf("Expected %d; Got: %d", testCase.expectedCode, resp.StatusCode)
			}
		})
	}

}

func Test_updateHandler_update(t *testing.T) {
	testCases := []struct {
		label        string
		reader       io.Reader
		expectedCode int
		uClient  mockInstantiationManager
	}{
		{
			label: "Update DIG",
			expectedCode: http.StatusAccepted,
			uClient:  mockInstantiationManager{},
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.label, func(t *testing.T) {
			request := httptest.NewRequest("POST", "/v2/projects/{project-name}/composite-apps/{composite-app-name}/{composite-app-version}/deployment-intent-groups/{deployment-intent-group-name}/update", testCase.reader)
			resp := executeRequest(request, NewRouter(nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, testCase.uClient))

			//Check returned code
			if resp.StatusCode != testCase.expectedCode {
				t.Fatalf("Expected %d; Got: %d", testCase.expectedCode, resp.StatusCode)
			}
		})
	}
}


func Test_updateHandler_rollback(t *testing.T) {
	testCases := []struct {
		label        string
		reader       io.Reader
		expectedCode int
		uClient  mockInstantiationManager
	}{
		{
			label: "Missing Body Failure",
			expectedCode: http.StatusBadRequest,
			reader: bytes.NewBuffer([]byte(``)),
			uClient:  mockInstantiationManager{},
		},
		{
			label: "Rollback DIG to given revision",
			expectedCode: http.StatusAccepted,
			reader: bytes.NewBuffer([]byte(`{
				"metadata" : {
    				"description": "rollback to revision 1"
				},
				"spec" : {
					"revision": "1"
				}
			}`)),
			uClient:  mockInstantiationManager{},
		},
		{
			label: "Missing revision number in Request Body",
			reader: bytes.NewBuffer([]byte(`{
				"metadata" : {
    				"description": "rollback to revision 1"
				}
			}`)),
			expectedCode: http.StatusBadRequest,
			uClient:  mockInstantiationManager{},
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.label, func(t *testing.T) {
			request := httptest.NewRequest("POST", "/v2/projects/{project-name}/composite-apps/{composite-app-name}/{composite-app-version}/deployment-intent-groups/{deployment-intent-group-name}/rollback", testCase.reader)
			resp := executeRequest(request, NewRouter(nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, testCase.uClient))

			//Check returned code
			if resp.StatusCode != testCase.expectedCode {
				t.Fatalf("Expected %d; Got: %d", testCase.expectedCode, resp.StatusCode)
			}
		})
	}

}