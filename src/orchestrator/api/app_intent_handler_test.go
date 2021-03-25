// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2020 Intel Corporation

package api

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	gpic "github.com/open-ness/EMCO/src/orchestrator/pkg/gpic"
	moduleLib "github.com/open-ness/EMCO/src/orchestrator/pkg/module"
)

//Creating an embedded interface via anonymous variable
//This allows us to make mockDB satisfy the DatabaseConnection
//interface even if we are not implementing all the methods in it
type mockAppIntentManager struct {
	// Items and err will be used to customize each test
	// via a localized instantiation of mockAppIntentManager
	Items []moduleLib.AppIntent
	Err   error
}

func (m *mockAppIntentManager) CreateAppIntent(a moduleLib.AppIntent, p string, ca string, v string, i string, digName string) (moduleLib.AppIntent, error) {
	if m.Err != nil {
		return moduleLib.AppIntent{}, m.Err
	}
	return m.Items[0], nil
}

func (m *mockAppIntentManager) GetAppIntent(ai string, p string, ca string, v string, i string, digName string) (moduleLib.AppIntent, error) {
	if m.Err != nil {
		return moduleLib.AppIntent{}, m.Err
	}

	return moduleLib.AppIntent{}, nil
}

func (m *mockAppIntentManager) GetAllIntentsByApp(aN, p, ca, v, i, digName string) (moduleLib.SpecData, error) {
	if m.Err != nil {
		return moduleLib.SpecData{}, m.Err
	}
	return moduleLib.SpecData{}, nil
}

func (m *mockAppIntentManager) DeleteAppIntent(ai string, p string, ca string, v string, i string, digName string) error {
	return m.Err
}

func (m *mockAppIntentManager) GetAllAppIntents(p, ca, v, i, digName string) ([]moduleLib.AppIntent, error) {
	return []moduleLib.AppIntent{}, nil
}

func init() {
	appIntentJSONFile = "../json-schemas/generic-placement-intent-app.json"
}

func Test_appintent_createHandler(t *testing.T) {
	testCases := []struct {
		label            string
		reader           io.Reader
		expected         moduleLib.AppIntent
		expectedCode     int
		errorString      string
		cAppIntentClient *mockAppIntentManager
	}{

		{
			label:        "Metadata name is missing",
			expectedCode: http.StatusBadRequest,
			errorString:  "Missing name for the intent",
			reader: bytes.NewBuffer([]byte(`{   "metadata": {
				"description": "description of placement_intent"
			 },
			 "spec": {
				"app-name": "app",
				"intent": {
				   "allOf": [
					  {
						 "provider-name": "p",
						 "cluster-label-name": "c"
					  },
					  {
						"provider-name": "p",
						"cluster-label-name": "d"
					 }
				   ]
				}
			 }
		  }`)),
			cAppIntentClient: &mockAppIntentManager{Err: errors.New("1")},
		},
		{
			label:        "app name is missing",
			expectedCode: http.StatusBadRequest,
			errorString:  "Missing app-name for the intent",
			reader: bytes.NewBuffer([]byte(`{   "metadata": {
				"name": "Test1",
				"description": "description of placement_intent"

			 },
			 "spec": {
				
			 }
		  }`)),
		  cAppIntentClient: &mockAppIntentManager{Err: errors.New("1")},
		},
		{
			label:        "provider name is missing",
			expectedCode: http.StatusBadRequest,
			errorString:  "Missing provider-name in an intent",
			reader: bytes.NewBuffer([]byte(`{   "metadata": {
				"name": "Test1",
				"description": "description of placement_intent"

			 },
			 "spec": {
				"app-name": "app1",
				"intent": {
					"anyOf": [
					  {
						"cluster-label-name": "c"
					  }
					]
				}
			  }		
			 }
		  }`)),
		  cAppIntentClient: &mockAppIntentManager{Err: errors.New("1")},
		},
		{
			label:        "cluster label or name is missing",
			expectedCode: http.StatusBadRequest,
			errorString:  "Missing cluster-name or cluster-label-name",
			reader: bytes.NewBuffer([]byte(`{   "metadata": {
				"name": "Test1",
				"description": "description of placement_intent"

			 },
			 "spec": {
				"app-name": "app1",
				"intent": {
					"anyOf": [
					  {
						"provider-name": "p"
					  }
					]
				}
			  }		
			 }
		  }`)),
		  cAppIntentClient: &mockAppIntentManager{Err: errors.New("1")},
		},
		{
			label:        "duplicate input only one cluster label or name required",
			expectedCode: http.StatusBadRequest,
			errorString:  "Only one of cluster name or cluster label allowed",
			reader: bytes.NewBuffer([]byte(`{   "metadata": {
				"name": "Test1",
				"description": "description of placement_intent"

			 },
			 "spec": {
				"app-name": "app1",
				"intent": {
					"anyOf": [
					  {
						"provider-name": "p",
						"cluster-label-name": "d",
						"cluster-name": "e"
					  }
					]
				}
			  }		
			 }
		  }`)),
		  cAppIntentClient: &mockAppIntentManager{Err: errors.New("1")},
		},
		{
			label:        "allOf provider name missing",
			expectedCode: http.StatusBadRequest,
			errorString:  "Missing provider-name in an intent",
			reader: bytes.NewBuffer([]byte(`{   "metadata": {
				"name": "Test1"			 },
			 "spec": {
				"app-name": "app1",
				"intent": {
					"allOf": [{
							"name": "p",
							"cluster-label-name": "c"
						},
						{
							"anyOf": [{
								"provider-name": "p",
								"cluster-label-name": "d"
							}]
						}
					]
				}
			}}}`)),
			cAppIntentClient: &mockAppIntentManager{Err: errors.New("1")},
		},
		{
			label:        "allOf anyof provider name missing",
			expectedCode: http.StatusBadRequest,
			errorString:  "Missing provider-name in an intent",
			reader: bytes.NewBuffer([]byte(`{   "metadata": {
				"name": "Test1"			 },
			 "spec": {
				"app-name": "app1",
				"intent": {
					"allOf": [{
							"provider-name": "p",
							"cluster-label-name": "c"
						},
						{
							"anyOf": [{
								"name": "p",
								"cluster-label-name": "d"
							}]
						}
					]
				}
			}}}`)),
			cAppIntentClient: &mockAppIntentManager{Err: errors.New("1")},
		},
		{
			label:        "allOf duplicate input only one cluster label or name required",
			expectedCode: http.StatusBadRequest,
			errorString:  "Only one of cluster name or cluster label allowed",
			reader: bytes.NewBuffer([]byte(`{   "metadata": {
				"name": "Test1"			 },
			 "spec": {
				"app-name": "app1",
				"intent": {
					"allOf": [{
							"provider-name": "p",
							"cluster-label-name": "c",
							"cluster-name": "e"
						},
						{
							"anyOf": [{
								"provider-name": "p",
								"cluster-label-name": "d"
							}]
						}
					]
				}
			}}}`)),
			cAppIntentClient: &mockAppIntentManager{Err: errors.New("1")},
		},
		{
			label:        "allOf anyOf duplicate input only one cluster label or name required",
			expectedCode: http.StatusBadRequest,
			errorString:  "Only one of cluster name or cluster label allowed",
			reader: bytes.NewBuffer([]byte(`{   "metadata": {
				"name": "Test1"			 },
			 "spec": {
				"app-name": "app1",
				"intent": {
					"allOf": [{
							"provider-name": "p",
							"cluster-label-name": "c"
						},
						{
							"anyOf": [{
								"provider-name": "p",
								"cluster-label-name": "d",
								"cluster-name": "e"
							}]
						}
					]
				}
			}}}`)),
			cAppIntentClient: &mockAppIntentManager{Err: errors.New("1")},
		},
		{
			label:        "Success Case",
			expectedCode: http.StatusCreated,
			errorString:  "",
			reader: bytes.NewBuffer([]byte(`{   
				"metadata": {
				"name": "Test1"			 
				},
			 "spec": {
				"app-name": "app1",
				"intent": {
					"allOf": [{
							"provider-name": "aws",
							"cluster-name": "edge1"
						},
						{
							"provider-name": "aws",
							"cluster-label-name": "west-us1"
						}
					]
				}
			}
			}`)),
			cAppIntentClient: &mockAppIntentManager{
				//Items that will be returned by the mocked Client
				Err: nil,
				Items: []moduleLib.AppIntent{
					{
						MetaData: moduleLib.MetaData{
							Name: "Test1",
						},
						Spec: moduleLib.SpecData{
							AppName: "app1",
							Intent: gpic.IntentStruc{
								AllOfArray: []gpic.AllOf{
									{
										ProviderName: "aws",
										ClusterName:  "edge1",
									},
									{
										ProviderName:     "aws",
										ClusterLabelName: "west-us1",
									},
								},
							},
						},
					},
				},
			},
			expected: moduleLib.AppIntent{
				MetaData: moduleLib.MetaData{
					Name: "Test1",
				},
				Spec: moduleLib.SpecData{
					AppName: "app1",
					Intent: gpic.IntentStruc{
						AllOfArray: []gpic.AllOf{
							{
								ProviderName: "aws",
								ClusterName:  "edge1",
							},
							{
								ProviderName:     "aws",
								ClusterLabelName: "west-us1",
							},
						},
					},
				},
			},
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.label, func(t *testing.T) {
			request := httptest.NewRequest("POST", "/v2/projects/{project-name}/composite-apps/{composite-app-name}/{version}/deployment-intent-groups/{deployment-intent-group-name}/generic-placement-intents/{intent-name}/app-intents", testCase.reader)
			resp := executeRequestReturnWithBody(request, NewRouter(nil, nil, nil, nil, nil, testCase.cAppIntentClient, nil, nil, nil, nil, nil))

			b := string(resp.Body.Bytes())

			//Check returned code
			if resp.Code != testCase.expectedCode {
				t.Fatalf("Expected %d; Got: %d", testCase.expectedCode, resp.Code)
			}
			//Check returned body only if statusCreated
			if resp.Code == http.StatusCreated {
				got := moduleLib.AppIntent{}
				json.NewDecoder(resp.Body).Decode(&got)
				if reflect.DeepEqual(testCase.expected, got) == false {
					t.Errorf("createHandler returned unexpected body: got %+v;"+
						" expected %+v", got, testCase.expected)
				}
			} else {
				if !strings.Contains(b, testCase.errorString) {
					t.Fatal("Unexpected error found")
				}
			}
		})
	}
}
