// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2021 Intel Corporation

package api

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	"github.com/open-ness/EMCO/src/sfc/api/mocks"
	"github.com/open-ness/EMCO/src/sfc/pkg/model"
	pkgerrors "github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func init() {
	sfcClientSelectorJSONFile = "../json-schemas/sfc-client-selector.json"
}

var _ = Describe("SfcClientSelectorintenthandler", func() {

	type testCase struct {
		inputName    string
		inputReader  io.Reader
		inStruct     model.SfcClientSelectorIntent
		mockError    error
		mockVal      model.SfcClientSelectorIntent
		mockVals     []model.SfcClientSelectorIntent
		expectedCode int
		client       *mocks.SfcClientSelectorIntentManager
	}

	DescribeTable("Create SfcClientSelectorIntent tests",
		func(t testCase) {
			// set up client mock responses
			t.client.On("CreateSfcClientSelectorIntent", t.inStruct, "test-project", "test-compositeapp", "v1", "test-dig", "net-ctl-intent", "sfc-intent", false).Return(t.mockVal, t.mockError)

			// make HTTP request
			request := httptest.NewRequest("POST", "/v2/projects/test-project/composite-apps/test-compositeapp/v1/deployment-intent-groups/test-dig/network-controller-intent/net-ctl-intent/network-chains/sfc-intent/client-selectors", t.inputReader)
			resp := executeRequest(request, NewRouter(t.client))

			//Check returned code
			Expect(resp.StatusCode).To(Equal(t.expectedCode))

			//Check returned body
			got := model.SfcClientSelectorIntent{}
			json.NewDecoder(resp.Body).Decode(&got)
			Expect(got).To(Equal(t.mockVal))
		},

		Entry("successful create", testCase{
			expectedCode: http.StatusCreated,
			inputReader: bytes.NewBuffer([]byte(`{
				"metadata": {
					"name": "testsfcclientselectorintent",
					"description": "test sfc intent",
					"userData1": "some user data 1",
					"userData2": "some user data 2"
				},
				"spec": {
					"chainEnd": "left",
				    "podSelector": {
						"matchLabels": {
							"app": "leftapp"
						}
					},
				    "namespaceSelector": {
						"matchLabels": {
							"app": "chainspace"
						}
					}
				}
			}`)),
			inStruct: model.SfcClientSelectorIntent{
				Metadata: model.Metadata{
					Name:        "testsfcclientselectorintent",
					Description: "test sfc intent",
					UserData1:   "some user data 1",
					UserData2:   "some user data 2",
				},
				Spec: model.SfcClientSelectorIntentSpec{
					ChainEnd: "left",
					PodSelector: metav1.LabelSelector{
						MatchLabels: map[string]string{
							"app": "leftapp"},
					},
					NamespaceSelector: metav1.LabelSelector{
						MatchLabels: map[string]string{
							"app": "chainspace"},
					},
				},
			},
			mockError: nil,
			mockVal: model.SfcClientSelectorIntent{
				Metadata: model.Metadata{
					Name:        "testsfcclientselectorintent",
					Description: "test sfc intent",
					UserData1:   "some user data 1",
					UserData2:   "some user data 2",
				},
				Spec: model.SfcClientSelectorIntentSpec{
					ChainEnd: "left",
					PodSelector: metav1.LabelSelector{
						MatchLabels: map[string]string{
							"app": "leftapp"},
					},
					NamespaceSelector: metav1.LabelSelector{
						MatchLabels: map[string]string{
							"app": "chainspace"},
					},
				},
			},
			client: &mocks.SfcClientSelectorIntentManager{},
		}),

		Entry("fails due to empty body", testCase{
			expectedCode: http.StatusBadRequest,
			inStruct:     model.SfcClientSelectorIntent{},
			mockError:    nil,
			mockVal:      model.SfcClientSelectorIntent{},
			client:       &mocks.SfcClientSelectorIntentManager{},
		}),

		Entry("fails due missing name", testCase{
			expectedCode: http.StatusBadRequest,
			inputReader: bytes.NewBuffer([]byte(`{
					"metadata": {
						"description": "test sfc intent",
						"userData1": "some user data 1",
						"userData2": "some user data 2"
					},
					"spec": {
						"chainEnd": "left",
						"podSelector": {
						"matchLabels": {
							"app": "leftapp"
							}
						},
						"namespaceSelector": {
						"matchLabels": {
							"app": "chainspace"
							}
						}
					}
				}`)),
			inStruct:  model.SfcClientSelectorIntent{},
			mockError: nil,
			client:    &mocks.SfcClientSelectorIntentManager{},
		}),

		Entry("fails due missing chain end", testCase{
			expectedCode: http.StatusBadRequest,
			inputReader: bytes.NewBuffer([]byte(`{
					"metadata": {
						"name": "testsfcclientselectorintent",
						"description": "test sfc intent",
						"userData1": "some user data 1",
						"userData2": "some user data 2"
					},
					"spec": {
					    "podSelector": {
							"matchLabels": {
								"app": "leftapp"
							}
						},
					    "namespaceSelector": {
							"matchLabels": {
								"app": "chainspace"
							}
						}
					}
				}`)),
			inStruct:  model.SfcClientSelectorIntent{},
			mockError: nil,
			client:    &mocks.SfcClientSelectorIntentManager{},
		}),

		Entry("fails due missing pod selector", testCase{
			expectedCode: http.StatusBadRequest,
			inputReader: bytes.NewBuffer([]byte(`{
					"metadata": {
						"name": "testsfcclientselectorintent",
						"description": "test sfc intent",
						"userData1": "some user data 1",
						"userData2": "some user data 2"
					},
					"spec": {
						"chainEnd": "left",
						"namespaceSelector": {
							"matchLabels": {
								"app": "leftapp"
							}
						}
					}
				}`)),
			inStruct:  model.SfcClientSelectorIntent{},
			mockError: nil,
			client:    &mocks.SfcClientSelectorIntentManager{},
		}),

		Entry("fails due missing matchLabels in namespace selector", testCase{
			// matchedLabels instead of matchLabels
			expectedCode: http.StatusBadRequest,
			inputReader: bytes.NewBuffer([]byte(`{
					"metadata": {
						"name": "testsfcclientselectorintent",
						"description": "test sfc intent",
						"userData1": "some user data 1",
						"userData2": "some user data 2"
					},
					"spec": {
						"chainEnd": "left",
					    "podSelector": {
							"matchLabels": {
								"app": "leftapp"
							}
						},
						"namespaceSelector": {
							"matchedLabels": {
								"app": "leftapp"
							}
						}
					}
				}`)),
			inStruct:  model.SfcClientSelectorIntent{},
			mockError: nil,
			client:    &mocks.SfcClientSelectorIntentManager{},
		}),

		Entry("fails due to json validation error", testCase{
			// name field has an '=' character
			expectedCode: http.StatusBadRequest,
			inputReader: bytes.NewBuffer([]byte(`{
					"metadata": {
						"name": "test=sfcintent",
						"description": "test sfc intent",
						"userData1": "some user data 1",
						"userData2": "some user data 2"
					},
					"spec": {
						"chainEnd": "left",
					    "podSelector": {
							"matchLabels": {
								"app": "leftapp"
							}
						},
					    "namespaceSelector": {
							"matchLabels": {
								"app": "chainspace"
							}
						}
					}
				}`)),
			inStruct:  model.SfcClientSelectorIntent{},
			mockError: nil,
			client:    &mocks.SfcClientSelectorIntentManager{},
		}),

		Entry("fails due to invalid chainEnd content", testCase{
			// chainEnd has value 'lefty'
			expectedCode: http.StatusBadRequest,
			inputReader: bytes.NewBuffer([]byte(`{
					"metadata": {
						"name": "testsfcintent",
						"description": "test sfc intent",
						"userData1": "some user data 1",
						"userData2": "some user data 2"
					},
					"spec": {
						"chainEnd": "lefty",
					    "podSelector": {
							"matchLabels": {
								"app": "leftapp"
							}
						},
					    "namespaceSelector": {
							"matchLabels": {
								"app": "chainspace"
							}
						}
					}
				}`)),
			inStruct:  model.SfcClientSelectorIntent{},
			mockError: nil,
			client:    &mocks.SfcClientSelectorIntentManager{},
		}),

		Entry("fails due to entry already exists", testCase{
			expectedCode: http.StatusConflict,
			inputReader: bytes.NewBuffer([]byte(`{
					"metadata": {
						"name": "testsfcclientselectorintent",
						"description": "test sfc intent",
						"userData1": "some user data 1",
						"userData2": "some user data 2"
					},
					"spec": {
						"chainEnd": "left",
						"podSelector": {
						"matchLabels": {
							"app": "leftapp"
							}
						},
						"namespaceSelector": {
						"matchLabels": {
							"app": "chainspace"
							}
						}
					}
				}`)),
			inStruct: model.SfcClientSelectorIntent{
				Metadata: model.Metadata{
					Name:        "testsfcclientselectorintent",
					Description: "test sfc intent",
					UserData1:   "some user data 1",
					UserData2:   "some user data 2",
				},
				Spec: model.SfcClientSelectorIntentSpec{
					ChainEnd: "left",
					PodSelector: metav1.LabelSelector{
						MatchLabels: map[string]string{
							"app": "leftapp"},
					},
					NamespaceSelector: metav1.LabelSelector{
						MatchLabels: map[string]string{
							"app": "chainspace"},
					},
				},
			},
			mockVal:   model.SfcClientSelectorIntent{},
			mockError: pkgerrors.New("SFC Client Selector already exists"),
			client:    &mocks.SfcClientSelectorIntentManager{},
		}),

		Entry("fails due to sfc client selector intent does not exist", testCase{
			expectedCode: http.StatusNotFound,
			inputReader: bytes.NewBuffer([]byte(`{
					"metadata": {
						"name": "testsfcclientselectorintent",
						"description": "test sfc intent",
						"userData1": "some user data 1",
						"userData2": "some user data 2"
					},
					"spec": {
						"chainEnd": "left",
						"podSelector": {
						"matchLabels": {
							"app": "leftapp"
							}
						},
						"namespaceSelector": {
						"matchLabels": {
							"app": "chainspace"
							}
						}
					}
				}`)),
			inStruct: model.SfcClientSelectorIntent{
				Metadata: model.Metadata{
					Name:        "testsfcclientselectorintent",
					Description: "test sfc intent",
					UserData1:   "some user data 1",
					UserData2:   "some user data 2",
				},
				Spec: model.SfcClientSelectorIntentSpec{
					ChainEnd: "left",
					PodSelector: metav1.LabelSelector{
						MatchLabels: map[string]string{
							"app": "leftapp"},
					},
					NamespaceSelector: metav1.LabelSelector{
						MatchLabels: map[string]string{
							"app": "chainspace"},
					},
				},
			},
			mockVal:   model.SfcClientSelectorIntent{},
			mockError: pkgerrors.New("does not exist"),
			client:    &mocks.SfcClientSelectorIntentManager{},
		}),

		Entry("fails due to db error", testCase{
			expectedCode: http.StatusInternalServerError,
			inputReader: bytes.NewBuffer([]byte(`{
					"metadata": {
						"name": "testsfcclientselectorintent",
						"description": "test sfc intent",
						"userData1": "some user data 1",
						"userData2": "some user data 2"
					},
					"spec": {
						"chainEnd": "left",
						"podSelector": {
						"matchLabels": {
							"app": "leftapp"
							}
						},
						"namespaceSelector": {
						"matchLabels": {
							"app": "chainspace"
							}
						}
					}
				}`)),
			inStruct: model.SfcClientSelectorIntent{
				Metadata: model.Metadata{
					Name:        "testsfcclientselectorintent",
					Description: "test sfc intent",
					UserData1:   "some user data 1",
					UserData2:   "some user data 2",
				},
				Spec: model.SfcClientSelectorIntentSpec{
					ChainEnd: "left",
					PodSelector: metav1.LabelSelector{
						MatchLabels: map[string]string{
							"app": "leftapp"},
					},
					NamespaceSelector: metav1.LabelSelector{
						MatchLabels: map[string]string{
							"app": "chainspace"},
					},
				},
			},
			mockVal:   model.SfcClientSelectorIntent{},
			mockError: pkgerrors.New("Creating DB Entry"),
			client:    &mocks.SfcClientSelectorIntentManager{},
		}),
	)

	DescribeTable("Put SfcClientSelectorIntent tests",
		func(t testCase) {
			// set up client mock responses
			t.client.On("CreateSfcClientSelectorIntent", t.inStruct, "test-project", "test-compositeapp", "v1", "test-dig", "net-ctl-intent", "sfc-intent", true).Return(t.mockVal, t.mockError)

			// make HTTP request
			request := httptest.NewRequest("PUT", "/v2/projects/test-project/composite-apps/test-compositeapp/v1/deployment-intent-groups/test-dig/network-controller-intent/net-ctl-intent/network-chains/sfc-intent/client-selectors/"+t.inputName, t.inputReader)
			resp := executeRequest(request, NewRouter(t.client))

			//Check returned code
			Expect(resp.StatusCode).To(Equal(t.expectedCode))

			//Check returned body
			got := model.SfcClientSelectorIntent{}
			json.NewDecoder(resp.Body).Decode(&got)
			Expect(got).To(Equal(t.mockVal))
		},

		Entry("successful put", testCase{
			expectedCode: http.StatusCreated,
			inputName:    "testsfcclientselectorintent",
			inputReader: bytes.NewBuffer([]byte(`{
				"metadata": {
					"name": "testsfcclientselectorintent",
					"description": "test sfc intent",
					"userData1": "some user data 1",
					"userData2": "some user data 2"
				},
				"spec": {
					"chainEnd": "left",
					"podSelector": {
						"matchLabels": {
							"app": "leftapp"
						}
					},
					"namespaceSelector": {
						"matchLabels": {
							"app": "chainspace"
						}
					}
				}
			}`)),
			inStruct: model.SfcClientSelectorIntent{
				Metadata: model.Metadata{
					Name:        "testsfcclientselectorintent",
					Description: "test sfc intent",
					UserData1:   "some user data 1",
					UserData2:   "some user data 2",
				},
				Spec: model.SfcClientSelectorIntentSpec{
					ChainEnd: "left",
					PodSelector: metav1.LabelSelector{
						MatchLabels: map[string]string{
							"app": "leftapp"},
					},
					NamespaceSelector: metav1.LabelSelector{
						MatchLabels: map[string]string{
							"app": "chainspace"},
					},
				},
			},
			mockError: nil,
			mockVal: model.SfcClientSelectorIntent{
				Metadata: model.Metadata{
					Name:        "testsfcclientselectorintent",
					Description: "test sfc intent",
					UserData1:   "some user data 1",
					UserData2:   "some user data 2",
				},
				Spec: model.SfcClientSelectorIntentSpec{
					ChainEnd: "left",
					PodSelector: metav1.LabelSelector{
						MatchLabels: map[string]string{
							"app": "leftapp"},
					},
					NamespaceSelector: metav1.LabelSelector{
						MatchLabels: map[string]string{
							"app": "chainspace"},
					},
				},
			},
			client: &mocks.SfcClientSelectorIntentManager{},
		}),

		Entry("fails due to empty body", testCase{
			expectedCode: http.StatusBadRequest,
			inputName:    "testsfcclientselectorintent",
			inStruct:     model.SfcClientSelectorIntent{},
			mockError:    nil,
			mockVal:      model.SfcClientSelectorIntent{},
			client:       &mocks.SfcClientSelectorIntentManager{},
		}),

		Entry("fails due missing name", testCase{
			expectedCode: http.StatusBadRequest,
			inputName:    "testsfcclientselectorintent",
			inputReader: bytes.NewBuffer([]byte(`{
					"metadata": {
						"description": "test sfc intent",
						"userData1": "some user data 1",
						"userData2": "some user data 2"
					},
					"spec": {
						"chainEnd": "left",
						"podSelector": {
							"matchLabels": {
								"app": "leftapp"
							}
						},
						"namespaceSelector": {
							"matchLabels": {
								"app": "chainspace"
							}
						}
					}
				}`)),
			inStruct:  model.SfcClientSelectorIntent{},
			mockError: nil,
			client:    &mocks.SfcClientSelectorIntentManager{},
		}),

		Entry("fails due missing chain end", testCase{
			expectedCode: http.StatusBadRequest,
			inputName:    "testsfcclientselectorintent",
			inputReader: bytes.NewBuffer([]byte(`{
					"metadata": {
						"name": "testsfcclientselectorintent",
						"description": "test sfc intent",
						"userData1": "some user data 1",
						"userData2": "some user data 2"
					},
					"spec": {
						"podSelector": {
							"matchLabels": {
								"app": "leftapp"
							}
						},
						"namespaceSelector": {
							"matchLabels": {
								"app": "chainspace"
							}
						}
					}
				}`)),
			inStruct:  model.SfcClientSelectorIntent{},
			mockError: nil,
			client:    &mocks.SfcClientSelectorIntentManager{},
		}),

		Entry("fails due to unparsable json body", testCase{
			// missing comma after podSelector object
			expectedCode: http.StatusUnprocessableEntity,
			inputName:    "testsfcclientselectorintent",
			inputReader: bytes.NewBuffer([]byte(`{
					"metadata": {
						"name": "testsfcintent",
						"description": "test sfc intent",
						"userData1": "some user data 1",
						"userData2": "some user data 2"
					},
					"spec": {
						"chainEnd": "left",
						"podSelector": {
							"matchLabels": {
								"app": "leftapp"
							}
						}
						"namespaceSelector": {
							"matchLabels": {
								"app": "chainspace"
							}
						}
					}
				}`)),
			inStruct:  model.SfcClientSelectorIntent{},
			mockError: nil,
			client:    &mocks.SfcClientSelectorIntentManager{},
		}),

		Entry("fails due to json body decoding error", testCase{
			// extra comma at the end of the userData2 line
			expectedCode: http.StatusUnprocessableEntity,
			inputName:    "testsfcclientselectorintent",
			inputReader: bytes.NewBuffer([]byte(`{
					"metadata": {
						"name": "testsfcclientselectorintent",
						"description": "test sfc intent",
						"userData1": "some user data 1",
						"userData2": "some user data 2",
					},
					"spec": {
						"chainEnd": "left",
						"podSelector": {
							"matchLabels": {
								"app": "leftapp"
							}
						},
						"namespaceSelector": {
							"matchLabels": {
								"app": "chainspace"
							}
						}
					}
				}`)),
			inStruct:  model.SfcClientSelectorIntent{},
			mockError: nil,
			client:    &mocks.SfcClientSelectorIntentManager{},
		}),

		Entry("fails due to name mismatch", testCase{
			expectedCode: http.StatusBadRequest,
			inputName:    "testsfcclientselectorintent",
			inputReader: bytes.NewBuffer([]byte(`{
					"metadata": {
						"name": "testsfcclientselectorintentABC",
						"description": "test sfc intent",
						"userData1": "some user data 1",
						"userData2": "some user data 2"
					},
					"spec": {
						"chainEnd": "left",
						"podSelector": {
							"matchLabels": {
								"app": "leftapp"
							}
						},
						"namespaceSelector": {
							"matchLabels": {
								"app": "chainspace"
							}
						}
					}
				}`)),
			inStruct: model.SfcClientSelectorIntent{
				Metadata: model.Metadata{
					Name:        "testsfcclientselectorintent",
					Description: "test sfc intent",
					UserData1:   "some user data 1",
					UserData2:   "some user data 2",
				},
				Spec: model.SfcClientSelectorIntentSpec{
					ChainEnd: "left",
					PodSelector: metav1.LabelSelector{
						MatchLabels: map[string]string{
							"app": "leftapp"},
					},
					NamespaceSelector: metav1.LabelSelector{
						MatchLabels: map[string]string{
							"app": "chainspace"},
					},
				},
			},
			mockVal:   model.SfcClientSelectorIntent{},
			mockError: pkgerrors.New("SfcClientSelectorIntent already exists"),
			client:    &mocks.SfcClientSelectorIntentManager{},
		}),

		Entry("fails due to sfc client selectorintent does not exist", testCase{
			expectedCode: http.StatusNotFound,
			inputName:    "testsfcclientselectorintent",
			inputReader: bytes.NewBuffer([]byte(`{
					"metadata": {
						"name": "testsfcclientselectorintent",
						"description": "test sfc intent",
						"userData1": "some user data 1",
						"userData2": "some user data 2"
					},
					"spec": {
						"chainEnd": "left",
						"podSelector": {
							"matchLabels": {
								"app": "leftapp"
							}
						},
						"namespaceSelector": {
							"matchLabels": {
								"app": "chainspace"
							}
						}
					}
				}`)),
			inStruct: model.SfcClientSelectorIntent{
				Metadata: model.Metadata{
					Name:        "testsfcclientselectorintent",
					Description: "test sfc intent",
					UserData1:   "some user data 1",
					UserData2:   "some user data 2",
				},
				Spec: model.SfcClientSelectorIntentSpec{
					ChainEnd: "left",
					PodSelector: metav1.LabelSelector{
						MatchLabels: map[string]string{
							"app": "leftapp"},
					},
					NamespaceSelector: metav1.LabelSelector{
						MatchLabels: map[string]string{
							"app": "chainspace"},
					},
				},
			},
			mockVal:   model.SfcClientSelectorIntent{},
			mockError: pkgerrors.New("does not exist"),
			client:    &mocks.SfcClientSelectorIntentManager{},
		}),

		Entry("fails due to db error", testCase{
			expectedCode: http.StatusInternalServerError,
			inputName:    "testsfcclientselectorintent",
			inputReader: bytes.NewBuffer([]byte(`{
					"metadata": {
						"name": "testsfcclientselectorintent",
						"description": "test sfc intent",
						"userData1": "some user data 1",
						"userData2": "some user data 2"
					},
					"spec": {
						"chainEnd": "left",
						"podSelector": {
							"matchLabels": {
								"app": "leftapp"
							}
						},
						"namespaceSelector": {
							"matchLabels": {
								"app": "chainspace"
							}
						}
					}
				}`)),
			inStruct: model.SfcClientSelectorIntent{
				Metadata: model.Metadata{
					Name:        "testsfcclientselectorintent",
					Description: "test sfc intent",
					UserData1:   "some user data 1",
					UserData2:   "some user data 2",
				},
				Spec: model.SfcClientSelectorIntentSpec{
					ChainEnd: "left",
					PodSelector: metav1.LabelSelector{
						MatchLabels: map[string]string{
							"app": "leftapp"},
					},
					NamespaceSelector: metav1.LabelSelector{
						MatchLabels: map[string]string{
							"app": "chainspace"},
					},
				},
			},
			mockVal:   model.SfcClientSelectorIntent{},
			mockError: pkgerrors.New("Creating DB Entry"),
			client:    &mocks.SfcClientSelectorIntentManager{},
		}),
	)

	DescribeTable("Get List SfcClientSelectorIntent tests",
		func(t testCase) {
			// set up client mock responses
			t.client.On("GetAllSfcClientSelectorIntents", "test-project", "test-compositeapp", "v1", "test-dig", "net-ctl-intent", "sfc-intent").Return(t.mockVals, t.mockError)

			// make HTTP request
			request := httptest.NewRequest("GET", "/v2/projects/test-project/composite-apps/test-compositeapp/v1/deployment-intent-groups/test-dig/network-controller-intent/net-ctl-intent/network-chains/sfc-intent/client-selectors", nil)
			resp := executeRequest(request, NewRouter(t.client))

			//Check returned code
			Expect(resp.StatusCode).To(Equal(t.expectedCode))

			//Check returned body
			got := []model.SfcClientSelectorIntent{}
			json.NewDecoder(resp.Body).Decode(&got)
			Expect(got).To(Equal(t.mockVals))
		},

		Entry("successful get", testCase{
			expectedCode: http.StatusOK,
			mockError:    nil,
			mockVals: []model.SfcClientSelectorIntent{
				{
					Metadata: model.Metadata{
						Name:        "testsfcclientselectorintent1",
						Description: "test sfc intent",
						UserData1:   "some user data 1",
						UserData2:   "some user data 2",
					},
					Spec: model.SfcClientSelectorIntentSpec{
						ChainEnd: "left",
						PodSelector: metav1.LabelSelector{
							MatchLabels: map[string]string{
								"app": "leftapp"},
						},
						NamespaceSelector: metav1.LabelSelector{
							MatchLabels: map[string]string{
								"app": "chainspace"},
						},
					},
				},
				{
					Metadata: model.Metadata{
						Name:        "testsfcclientselectorintent2",
						Description: "test sfc intent",
						UserData1:   "some user data 1",
						UserData2:   "some user data 2",
					},
					Spec: model.SfcClientSelectorIntentSpec{
						ChainEnd: "left",
						PodSelector: metav1.LabelSelector{
							MatchLabels: map[string]string{
								"app": "leftapp"},
						},
						NamespaceSelector: metav1.LabelSelector{
							MatchLabels: map[string]string{
								"app": "chainspace"},
						},
					},
				},
			},
			client: &mocks.SfcClientSelectorIntentManager{},
		}),

		Entry("fails due to db find error", testCase{
			expectedCode: http.StatusNotFound,
			mockError:    pkgerrors.New("db Find error"),
			mockVals:     []model.SfcClientSelectorIntent{},
			client:       &mocks.SfcClientSelectorIntentManager{},
		}),

		Entry("fails due to parent SFC Intent not found", testCase{
			expectedCode: http.StatusNotFound,
			mockError:    pkgerrors.New("SFC Intent not found"),
			mockVals:     []model.SfcClientSelectorIntent{},
			client:       &mocks.SfcClientSelectorIntentManager{},
		}),

		Entry("fails due to some other backend error", testCase{
			expectedCode: http.StatusInternalServerError,
			mockError:    pkgerrors.New("backend error"),
			mockVals:     []model.SfcClientSelectorIntent{},
			client:       &mocks.SfcClientSelectorIntentManager{},
		}),
	)

	DescribeTable("Get SfcClientSelectorIntent tests",
		func(t testCase) {
			// set up client mock responses
			t.client.On("GetSfcClientSelectorIntent", t.inputName, "test-project", "test-compositeapp", "v1", "test-dig", "net-ctl-intent", "sfc-intent").Return(t.mockVal, t.mockError)

			// make HTTP request
			request := httptest.NewRequest("GET", "/v2/projects/test-project/composite-apps/test-compositeapp/v1/deployment-intent-groups/test-dig/network-controller-intent/net-ctl-intent/network-chains/sfc-intent/client-selectors/"+t.inputName, nil)
			resp := executeRequest(request, NewRouter(t.client))

			//Check returned code
			Expect(resp.StatusCode).To(Equal(t.expectedCode))

			//Check returned body
			got := model.SfcClientSelectorIntent{}
			json.NewDecoder(resp.Body).Decode(&got)
			Expect(got).To(Equal(t.mockVal))
		},

		Entry("successful get", testCase{
			inputName:    "testsfcclientselectorintent",
			expectedCode: http.StatusOK,
			mockError:    nil,
			mockVal: model.SfcClientSelectorIntent{
				Metadata: model.Metadata{
					Name:        "testsfcclientselectorintent",
					Description: "test sfc intent",
					UserData1:   "some user data 1",
					UserData2:   "some user data 2",
				},
				Spec: model.SfcClientSelectorIntentSpec{
					ChainEnd: "left",
					PodSelector: metav1.LabelSelector{
						MatchLabels: map[string]string{
							"app": "leftapp"},
					},
					NamespaceSelector: metav1.LabelSelector{
						MatchLabels: map[string]string{
							"app": "chainspace"},
					},
				},
			},
			client: &mocks.SfcClientSelectorIntentManager{},
		}),

		Entry("fails due to not found", testCase{
			inputName:    "testsfcclientselectorintent",
			expectedCode: http.StatusNotFound,
			mockError:    pkgerrors.New("db Find error"),
			mockVal:      model.SfcClientSelectorIntent{},
			client:       &mocks.SfcClientSelectorIntentManager{},
		}),

		Entry("fails due to not found II", testCase{
			inputName:    "testsfcclientselectorintent",
			expectedCode: http.StatusNotFound,
			mockError:    pkgerrors.New("not found"),
			mockVal:      model.SfcClientSelectorIntent{},
			client:       &mocks.SfcClientSelectorIntentManager{},
		}),

		Entry("fails due to some other backend error", testCase{
			inputName:    "testsfcclientselectorintent",
			expectedCode: http.StatusInternalServerError,
			mockError:    pkgerrors.New("backend error"),
			mockVal:      model.SfcClientSelectorIntent{},
			client:       &mocks.SfcClientSelectorIntentManager{},
		}),
	)

	DescribeTable("Delete SfcClientSelectorIntent tests",
		func(t testCase) {
			// set up client mock responses
			t.client.On("DeleteSfcClientSelectorIntent", t.inputName, "test-project", "test-compositeapp", "v1", "test-dig", "net-ctl-intent", "sfc-intent").Return(t.mockError)

			// make HTTP request
			request := httptest.NewRequest("DELETE", "/v2/projects/test-project/composite-apps/test-compositeapp/v1/deployment-intent-groups/test-dig/network-controller-intent/net-ctl-intent/network-chains/sfc-intent/client-selectors/"+t.inputName, nil)
			resp := executeRequest(request, NewRouter(t.client))

			//Check returned code
			Expect(resp.StatusCode).To(Equal(t.expectedCode))

			//Check returned body
			got := model.SfcClientSelectorIntent{}
			json.NewDecoder(resp.Body).Decode(&got)
			Expect(got).To(Equal(t.mockVal))
		},

		Entry("successful delete", testCase{
			inputName:    "testsfcclientselectorintent",
			expectedCode: http.StatusNoContent,
			mockError:    nil,
			client:       &mocks.SfcClientSelectorIntentManager{},
		}),

		Entry("fails due to not found", testCase{
			inputName:    "testsfcclientselectorintent",
			expectedCode: http.StatusNotFound,
			mockError:    pkgerrors.New("db Remove error - not found"),
			client:       &mocks.SfcClientSelectorIntentManager{},
		}),

		Entry("fails due to a conflict", testCase{
			inputName:    "testsfcclientselectorintent",
			expectedCode: http.StatusConflict,
			mockError:    pkgerrors.New("db Remove error - conflict"),
			client:       &mocks.SfcClientSelectorIntentManager{},
		}),

		Entry("fails due to other backend error", testCase{
			inputName:    "testsfcclientselectorintent",
			expectedCode: http.StatusInternalServerError,
			mockError:    pkgerrors.New("db Remove error - general"),
			client:       &mocks.SfcClientSelectorIntentManager{},
		}),
	)
})
