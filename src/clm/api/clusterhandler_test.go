// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2020 Intel Corporation

package api

import (
	"bytes"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/textproto"
	"reflect"
	"testing"

	"github.com/open-ness/EMCO/src/clm/pkg/cluster"
	types "github.com/open-ness/EMCO/src/orchestrator/pkg/module/types"
	"github.com/open-ness/EMCO/src/orchestrator/pkg/state"

	pkgerrors "github.com/pkg/errors"
)

//Creating an embedded interface via anonymous variable
//This allows us to make mockDB satisfy the DatabaseConnection
//interface even if we are not implementing all the methods in it
type mockClusterManager struct {
	// Items and err will be used to customize each test
	// via a localized instantiation of mockClusterManager
	ClusterProviderItems []cluster.ClusterProvider
	ClusterItems         []cluster.Cluster
	ClusterContentItems  []cluster.ClusterContent
	ClusterStateInfo     []state.StateInfo
	ClusterLabelItems    []cluster.ClusterLabel
	ClusterKvPairsItems  []cluster.ClusterKvPairs
	ClusterList          []string
	Err                  error
}

func (m *mockClusterManager) CreateClusterProvider(inp cluster.ClusterProvider, exists bool) (cluster.ClusterProvider, error) {
	if m.Err != nil {
		return cluster.ClusterProvider{}, m.Err
	}

	return m.ClusterProviderItems[0], nil
}

func (m *mockClusterManager) GetClusterProvider(name string) (cluster.ClusterProvider, error) {
	if m.Err != nil {
		return cluster.ClusterProvider{}, m.Err
	}

	return m.ClusterProviderItems[0], nil
}

func (m *mockClusterManager) GetClusterProviders() ([]cluster.ClusterProvider, error) {
	if m.Err != nil {
		return []cluster.ClusterProvider{}, m.Err
	}

	return m.ClusterProviderItems, nil
}

func (m *mockClusterManager) DeleteClusterProvider(name string) error {
	return m.Err
}

func (m *mockClusterManager) CreateCluster(provider string, inp cluster.Cluster, inq cluster.ClusterContent) (cluster.Cluster, error) {
	if m.Err != nil {
		return cluster.Cluster{}, m.Err
	}

	return m.ClusterItems[0], nil
}

func (m *mockClusterManager) GetCluster(provider, name string) (cluster.Cluster, error) {
	if m.Err != nil {
		return cluster.Cluster{}, m.Err
	}

	return m.ClusterItems[0], nil
}

func (m *mockClusterManager) GetClusterContent(provider, name string) (cluster.ClusterContent, error) {
	if m.Err != nil {
		return cluster.ClusterContent{}, m.Err
	}

	return m.ClusterContentItems[0], nil
}

func (m *mockClusterManager) GetClusterState(provider, name string) (state.StateInfo, error) {
	if m.Err != nil {
		return state.StateInfo{}, m.Err
	}

	return m.ClusterStateInfo[0], nil
}

func (m *mockClusterManager) GetClusters(provider string) ([]cluster.Cluster, error) {
	if m.Err != nil {
		return []cluster.Cluster{}, m.Err
	}

	return m.ClusterItems, nil
}

func (m *mockClusterManager) GetClustersWithLabel(provider, label string) ([]string, error) {
	if m.Err != nil {
		return []string{}, m.Err
	}

	return m.ClusterList, nil
}

func (m *mockClusterManager) DeleteCluster(provider, name string) error {
	return m.Err
}

func (m *mockClusterManager) CreateClusterLabel(provider, clusterName string, inp cluster.ClusterLabel, exists bool) (cluster.ClusterLabel, error) {
	if m.Err != nil {
		return cluster.ClusterLabel{}, m.Err
	}

	return m.ClusterLabelItems[0], nil
}

func (m *mockClusterManager) GetClusterLabel(provider, clusterName, label string) (cluster.ClusterLabel, error) {
	if m.Err != nil {
		return cluster.ClusterLabel{}, m.Err
	}

	return m.ClusterLabelItems[0], nil
}

func (m *mockClusterManager) GetClusterLabels(provider, clusterName string) ([]cluster.ClusterLabel, error) {
	if m.Err != nil {
		return []cluster.ClusterLabel{}, m.Err
	}

	return m.ClusterLabelItems, nil
}

func (m *mockClusterManager) DeleteClusterLabel(provider, clusterName, label string) error {
	return m.Err
}

func (m *mockClusterManager) CreateClusterKvPairs(provider, clusterName string, inp cluster.ClusterKvPairs, exists bool) (cluster.ClusterKvPairs, error) {
	if m.Err != nil {
		return cluster.ClusterKvPairs{}, m.Err
	}

	return m.ClusterKvPairsItems[0], nil
}

func (m *mockClusterManager) GetClusterKvPairs(provider, clusterName, kvpair string) (cluster.ClusterKvPairs, error) {
	if m.Err != nil {
		return cluster.ClusterKvPairs{}, m.Err
	}

	return m.ClusterKvPairsItems[0], nil
}

func (m *mockClusterManager) GetClusterKvPairsValue(provider, cluster, kvpair, kvkey string) (interface{}, error) {
	if m.Err != nil {
		return nil, m.Err
	}

	for _, kvMap := range m.ClusterKvPairsItems[0].Spec.Kv {
		if val, ok := kvMap[kvkey]; ok {
			return val, nil
		}
	}
	return nil, m.Err

}

func (m *mockClusterManager) GetAllClusterKvPairs(provider, clusterName string) ([]cluster.ClusterKvPairs, error) {
	if m.Err != nil {
		return []cluster.ClusterKvPairs{}, m.Err
	}

	return m.ClusterKvPairsItems, nil
}

func (m *mockClusterManager) DeleteClusterKvPairs(provider, clusterName, kvpair string) error {
	return m.Err
}

func init() {
	cpJSONFile = "../json-schemas/metadata.json"
	ckvJSONFile = "../json-schemas/cluster-kv.json"
	clJSONFile = "../json-schemas/cluster-label.json"
}

func TestClusterProviderCreateHandler(t *testing.T) {
	testCases := []struct {
		label         string
		reader        io.Reader
		expected      cluster.ClusterProvider
		expectedCode  int
		clusterClient *mockClusterManager
	}{
		{
			label:         "Missing Cluster Provider Body Failure",
			expectedCode:  http.StatusBadRequest,
			clusterClient: &mockClusterManager{},
		},
		{
			label:        "Create Cluster Provider",
			expectedCode: http.StatusCreated,
			reader: bytes.NewBuffer([]byte(`{
					"metadata": {
						"name": "clusterProviderTest",
						"description": "testClusterProvider",
						"userData1": "some user data 1",
						"userData2": "some user data 2"
					}
				}`)),
			expected: cluster.ClusterProvider{
				Metadata: types.Metadata{
					Name:        "clusterProviderTest",
					Description: "testClusterProvider",
					UserData1:   "some user data 1",
					UserData2:   "some user data 2",
				},
			},
			clusterClient: &mockClusterManager{
				//Items that will be returned by the mocked Client
				ClusterProviderItems: []cluster.ClusterProvider{
					{
						Metadata: types.Metadata{
							Name:        "clusterProviderTest",
							Description: "testClusterProvider",
							UserData1:   "some user data 1",
							UserData2:   "some user data 2",
						},
					},
				},
			},
		},
		{
			label: "Missing ClusterProvider Name in Request Body",
			reader: bytes.NewBuffer([]byte(`{
					"metadata": {
						"description": "this is a test cluster provider",
						"userData1": "some user data 1",
						"userData2": "some user data 2"
					}
				}`)),
			expectedCode:  http.StatusBadRequest,
			clusterClient: &mockClusterManager{},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.label, func(t *testing.T) {
			request := httptest.NewRequest("POST", "/v2/cluster-providers", testCase.reader)
			resp := executeRequest(request, NewRouter(testCase.clusterClient))

			//Check returned code
			if resp.StatusCode != testCase.expectedCode {
				t.Fatalf("Expected %d; Got: %d", testCase.expectedCode, resp.StatusCode)
			}

			//Check returned body only if statusCreated
			if resp.StatusCode == http.StatusCreated {
				got := cluster.ClusterProvider{}
				json.NewDecoder(resp.Body).Decode(&got)

				if reflect.DeepEqual(testCase.expected, got) == false {
					t.Errorf("createHandler returned unexpected body: got %v;"+
						" expected %v", got, testCase.expected)
				}
			}
		})
	}
}

func TestClusterProviderPutHandler(t *testing.T) {
	testCases := []struct {
		label         string
		name          string
		reader        io.Reader
		expected      cluster.ClusterProvider
		expectedCode  int
		clusterClient *mockClusterManager
	}{
		{
			label:        "Create Cluster Provider via PUT",
			name:         "clusterProviderTest",
			expectedCode: http.StatusCreated,
			reader: bytes.NewBuffer([]byte(`{
					"metadata": {
						"name": "clusterProviderTest",
						"description": "testClusterProvider",
						"userData1": "some user data 1",
						"userData2": "some user data 2"
					}
				}`)),
			expected: cluster.ClusterProvider{
				Metadata: types.Metadata{
					Name:        "clusterProviderTest",
					Description: "testClusterProvider",
					UserData1:   "some user data 1",
					UserData2:   "some user data 2",
				},
			},
			clusterClient: &mockClusterManager{
				//Items that will be returned by the mocked Client
				ClusterProviderItems: []cluster.ClusterProvider{
					{
						Metadata: types.Metadata{
							Name:        "clusterProviderTest",
							Description: "testClusterProvider",
							UserData1:   "some user data 1",
							UserData2:   "some user data 2",
						},
					},
				},
			},
		},
		{
			label: "Missing ClusterProvider Name in Request Body",
			name:  "clusterProviderTest",
			reader: bytes.NewBuffer([]byte(`{
					"metadata": {
						"description": "this is a test cluster provider",
						"userData1": "some user data 1",
						"userData2": "some user data 2"
					}
				}`)),
			expectedCode:  http.StatusBadRequest,
			clusterClient: &mockClusterManager{},
		},
		{
			label: "Mismatched Cluster Provider Name in Request",
			name:  "clusterProviderTestABC",
			reader: bytes.NewBuffer([]byte(`{
					"metadata": {
						"name": "clusterProviderTest",
						"description": "testClusterProvider",
						"userData1": "some user data 1",
						"userData2": "some user data 2"
					}
				}`)),
			expectedCode:  http.StatusBadRequest,
			clusterClient: &mockClusterManager{},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.label, func(t *testing.T) {
			request := httptest.NewRequest("PUT", "/v2/cluster-providers/"+testCase.name, testCase.reader)
			resp := executeRequest(request, NewRouter(testCase.clusterClient))

			//Check returned code
			if resp.StatusCode != testCase.expectedCode {
				t.Fatalf("Expected %d; Got: %d", testCase.expectedCode, resp.StatusCode)
			}

			//Check returned body only if statusCreated
			if resp.StatusCode == http.StatusCreated {
				got := cluster.ClusterProvider{}
				json.NewDecoder(resp.Body).Decode(&got)

				if reflect.DeepEqual(testCase.expected, got) == false {
					t.Errorf("createHandler returned unexpected body: got %v;"+
						" expected %v", got, testCase.expected)
				}
			}
		})
	}
}

func TestClusterProviderGetAllHandler(t *testing.T) {

	testCases := []struct {
		label         string
		expected      []cluster.ClusterProvider
		name, version string
		expectedCode  int
		clusterClient *mockClusterManager
	}{
		{
			label:        "Get Cluster Provider",
			expectedCode: http.StatusOK,
			expected: []cluster.ClusterProvider{
				{
					Metadata: types.Metadata{
						Name:        "testClusterProvider1",
						Description: "testClusterProvider 1 description",
						UserData1:   "some user data 1",
						UserData2:   "some user data 2",
					},
				},
				{
					Metadata: types.Metadata{
						Name:        "testClusterProvider2",
						Description: "testClusterProvider 2 description",
						UserData1:   "some user data A",
						UserData2:   "some user data B",
					},
				},
			},
			clusterClient: &mockClusterManager{
				//Items that will be returned by the mocked Client
				ClusterProviderItems: []cluster.ClusterProvider{
					{
						Metadata: types.Metadata{
							Name:        "testClusterProvider1",
							Description: "testClusterProvider 1 description",
							UserData1:   "some user data 1",
							UserData2:   "some user data 2",
						},
					},
					{
						Metadata: types.Metadata{
							Name:        "testClusterProvider2",
							Description: "testClusterProvider 2 description",
							UserData1:   "some user data A",
							UserData2:   "some user data B",
						},
					},
				},
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.label, func(t *testing.T) {
			request := httptest.NewRequest("GET", "/v2/cluster-providers", nil)
			resp := executeRequest(request, NewRouter(testCase.clusterClient))

			//Check returned code
			if resp.StatusCode != testCase.expectedCode {
				t.Fatalf("Expected %d; Got: %d", testCase.expectedCode, resp.StatusCode)
			}

			//Check returned body only if statusOK
			if resp.StatusCode == http.StatusOK {
				got := []cluster.ClusterProvider{}
				json.NewDecoder(resp.Body).Decode(&got)

				if reflect.DeepEqual(testCase.expected, got) == false {
					t.Errorf("listHandler returned unexpected body: got %v;"+
						" expected %v", got, testCase.expected)
				}
			}
		})
	}
}

func TestClusterProviderGetHandler(t *testing.T) {

	testCases := []struct {
		label         string
		expected      cluster.ClusterProvider
		name, version string
		expectedCode  int
		clusterClient *mockClusterManager
	}{
		{
			label:        "Get Cluster Provider",
			expectedCode: http.StatusOK,
			expected: cluster.ClusterProvider{
				Metadata: types.Metadata{
					Name:        "testClusterProvider",
					Description: "testClusterProvider description",
					UserData1:   "some user data 1",
					UserData2:   "some user data 2",
				},
			},
			name: "testClusterProvider",
			clusterClient: &mockClusterManager{
				//Items that will be returned by the mocked Client
				ClusterProviderItems: []cluster.ClusterProvider{
					{
						Metadata: types.Metadata{
							Name:        "testClusterProvider",
							Description: "testClusterProvider description",
							UserData1:   "some user data 1",
							UserData2:   "some user data 2",
						},
					},
				},
			},
		},
		{
			label:        "Get Non-Existing Cluster Provider",
			expectedCode: http.StatusNotFound,
			name:         "nonexistingclusterprovider",
			clusterClient: &mockClusterManager{
				ClusterProviderItems: []cluster.ClusterProvider{},
				Err:                  pkgerrors.New("db Find error"),
			},
		},
		{
			label:        "Get Non-Existing Cluster Provider ",
			expectedCode: http.StatusNotFound,
			name:         "nonexistingclusterprovider",
			clusterClient: &mockClusterManager{
				ClusterProviderItems: []cluster.ClusterProvider{},
				Err:                  pkgerrors.New("Cluster provider not found"),
			},
		},
		{
			label:        "Get Cluster Provider internal error",
			expectedCode: http.StatusInternalServerError,
			name:         "nonexistingclusterprovider",
			clusterClient: &mockClusterManager{
				ClusterProviderItems: []cluster.ClusterProvider{},
				Err:                  pkgerrors.New("some internal error"),
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.label, func(t *testing.T) {
			request := httptest.NewRequest("GET", "/v2/cluster-providers/"+testCase.name, nil)
			resp := executeRequest(request, NewRouter(testCase.clusterClient))

			//Check returned code
			if resp.StatusCode != testCase.expectedCode {
				t.Fatalf("Expected %d; Got: %d", testCase.expectedCode, resp.StatusCode)
			}

			//Check returned body only if statusOK
			if resp.StatusCode == http.StatusOK {
				got := cluster.ClusterProvider{}
				json.NewDecoder(resp.Body).Decode(&got)

				if reflect.DeepEqual(testCase.expected, got) == false {
					t.Errorf("listHandler returned unexpected body: got %v;"+
						" expected %v", got, testCase.expected)
				}
			}
		})
	}
}

func TestClusterProviderDeleteHandler(t *testing.T) {

	testCases := []struct {
		label         string
		name          string
		version       string
		expectedCode  int
		clusterClient *mockClusterManager
	}{
		{
			label:         "Delete Cluster Provider",
			expectedCode:  http.StatusNoContent,
			name:          "testClusterProvider",
			clusterClient: &mockClusterManager{},
		},
		{
			label:        "Delete Cluster Provider internal error",
			expectedCode: http.StatusInternalServerError,
			name:         "testClusterProvider",
			clusterClient: &mockClusterManager{
				Err: pkgerrors.New("Internal Error"),
			},
		},
		{
			label:        "Delete Non-Existing Cluster Provider",
			expectedCode: http.StatusNotFound,
			name:         "testClusterProvider",
			clusterClient: &mockClusterManager{
				Err: pkgerrors.New("cluster provider not found"),
			},
		},
		{
			label:        "Delete Cluster Provider with children",
			expectedCode: http.StatusConflict,
			name:         "testClusterProvider",
			clusterClient: &mockClusterManager{
				Err: pkgerrors.New("db Remove error - conflict"),
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.label, func(t *testing.T) {
			request := httptest.NewRequest("DELETE", "/v2/cluster-providers/"+testCase.name, nil)
			resp := executeRequest(request, NewRouter(testCase.clusterClient))

			//Check returned code
			if resp.StatusCode != testCase.expectedCode {
				t.Fatalf("Expected %d; Got: %d", testCase.expectedCode, resp.StatusCode)
			}
		})
	}
}

func TestClusterCreateHandler(t *testing.T) {
	testCases := []struct {
		label         string
		metadata      string
		kubeconfig    string
		expected      cluster.Cluster
		expectedCode  int
		clusterClient *mockClusterManager
	}{
		{
			label:         "Missing Cluster Body Failure",
			expectedCode:  http.StatusBadRequest,
			clusterClient: &mockClusterManager{},
		},
		{
			label:        "Create Cluster",
			expectedCode: http.StatusCreated,
			metadata: `
{
 "metadata": {
  "name": "clusterTest",
  "description": "this is test cluster",
  "userData1": "some cluster data abc",
  "userData2": "some cluster data def"
 }
}`,
			kubeconfig: `test contents
of a file attached
to the creation
of clusterTest
`,
			expected: cluster.Cluster{
				Metadata: types.Metadata{
					Name:        "clusterTest",
					Description: "testCluster",
					UserData1:   "some user data 1",
					UserData2:   "some user data 2",
				},
			},
			clusterClient: &mockClusterManager{
				//Items that will be returned by the mocked Client
				ClusterProviderItems: []cluster.ClusterProvider{
					{
						Metadata: types.Metadata{
							Name:        "clusterProvider1",
							Description: "ClusterProvider 1 description",
							UserData1:   "some user data 1",
							UserData2:   "some user data 2",
						},
					},
				},
				ClusterItems: []cluster.Cluster{
					{
						Metadata: types.Metadata{
							Name:        "clusterTest",
							Description: "testCluster",
							UserData1:   "some user data 1",
							UserData2:   "some user data 2",
						},
					},
				},
				ClusterContentItems: []cluster.ClusterContent{
					{
						Kubeconfig: "dGVzdCBjb250ZW50cwpvZiBhIGZpbGUgYXR0YWNoZWQKdG8gdGhlIGNyZWF0aW9uCm9mIGNsdXN0ZXJUZXN0Cg==",
					},
				},
			},
		},
		{
			label:        "Missing Cluster Name in Request Body",
			expectedCode: http.StatusBadRequest,
			metadata: `
{
 "metadata": {
  "description": "this is test cluster",
  "userData1": "some cluster data abc",
  "userData2": "some cluster data def"
 }
}`,
			kubeconfig: `test contents
of a file attached
to the creation
of clusterTest
`,
			clusterClient: &mockClusterManager{},
		},
		{
			label:        "Cluster already exists",
			expectedCode: http.StatusConflict,
			metadata: `
{
 "metadata": {
  "name": "alreadyExists",
  "description": "this is test cluster",
  "userData1": "some cluster data abc",
  "userData2": "some cluster data def"
 }
}`,
			kubeconfig: `test contents
of a file attached
to the creation
of clusterTest
`,
			clusterClient: &mockClusterManager{
				//Items that will be returned by the mocked Client
				ClusterItems:        []cluster.Cluster{},
				ClusterContentItems: []cluster.ClusterContent{},
				Err:                 pkgerrors.New("Cluster already exists"),
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.label, func(t *testing.T) {
			// Create the multipart test Request body
			body := new(bytes.Buffer)
			multiwr := multipart.NewWriter(body)
			multiwr.SetBoundary("------------------------f77f80a7092eb312")
			pw, _ := multiwr.CreatePart(textproto.MIMEHeader{"Content-Type": {"application/json"}, "Content-Disposition": {"form-data; name=metadata"}})
			pw.Write([]byte(testCase.metadata))
			pw, _ = multiwr.CreateFormFile("file", "kubeconfig")
			pw.Write([]byte(testCase.kubeconfig))
			multiwr.Close()

			request := httptest.NewRequest("POST", "/v2/cluster-providers/clusterProvider1/clusters", bytes.NewBuffer(body.Bytes()))
			request.Header.Set("Content-Type", multiwr.FormDataContentType())
			resp := executeRequest(request, NewRouter(testCase.clusterClient))

			//Check returned code
			if resp.StatusCode != testCase.expectedCode {
				t.Fatalf("Expected %d; Got: %d", testCase.expectedCode, resp.StatusCode)
			}

			//Check returned body only if statusCreated
			if resp.StatusCode == http.StatusCreated {
				got := cluster.Cluster{}
				json.NewDecoder(resp.Body).Decode(&got)

				if reflect.DeepEqual(testCase.expected, got) == false {
					t.Errorf("createHandler returned unexpected body: got %v;"+
						" expected %v", got, testCase.expected)
				}
			}
		})
	}
}

func TestClusterGetAllHandler(t *testing.T) {

	testCases := []struct {
		label         string
		expected      []cluster.Cluster
		name, version string
		expectedCode  int
		clusterClient *mockClusterManager
	}{
		{
			label:        "Get Clusters",
			expectedCode: http.StatusOK,
			expected: []cluster.Cluster{
				{
					Metadata: types.Metadata{
						Name:        "testCluster1",
						Description: "testCluster 1 description",
						UserData1:   "some user data 1",
						UserData2:   "some user data 2",
					},
				},
				{
					Metadata: types.Metadata{
						Name:        "testCluster2",
						Description: "testCluster 2 description",
						UserData1:   "some user data A",
						UserData2:   "some user data B",
					},
				},
			},
			clusterClient: &mockClusterManager{
				//Items that will be returned by the mocked Client
				ClusterItems: []cluster.Cluster{
					{
						Metadata: types.Metadata{
							Name:        "testCluster1",
							Description: "testCluster 1 description",
							UserData1:   "some user data 1",
							UserData2:   "some user data 2",
						},
					},
					{
						Metadata: types.Metadata{
							Name:        "testCluster2",
							Description: "testCluster 2 description",
							UserData1:   "some user data A",
							UserData2:   "some user data B",
						},
					},
				},
				ClusterContentItems: []cluster.ClusterContent{
					// content here doesn't matter - just needs to be present
					{
						Kubeconfig: "dGVzdCBjb250ZW50cwpvZiBhIGZpbGUgYXR0YWNoZWQKdG8gdGhlIGNyZWF0aW9uCm9mIGNsdXN0ZXJUZXN0Cg==",
					},
					{
						Kubeconfig: "dGVzdCBjb250ZW50cwpvZiBhIGZpbGUgYXR0YWNoZWQKdG8gdGhlIGNyZWF0aW9uCm9mIGNsdXN0ZXJUZXN0Cg==",
					},
				},
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.label, func(t *testing.T) {
			request := httptest.NewRequest("GET", "/v2/cluster-providers/clusterProvder1/clusters", nil)
			resp := executeRequest(request, NewRouter(testCase.clusterClient))

			//Check returned code
			if resp.StatusCode != testCase.expectedCode {
				t.Fatalf("Expected %d; Got: %d", testCase.expectedCode, resp.StatusCode)
			}

			//Check returned body only if statusOK
			if resp.StatusCode == http.StatusOK {
				got := []cluster.Cluster{}
				json.NewDecoder(resp.Body).Decode(&got)

				if reflect.DeepEqual(testCase.expected, got) == false {
					t.Errorf("listHandler returned unexpected body: got %v;"+
						" expected %v", got, testCase.expected)
				}
			}
		})
	}
}

func TestClusterGetHandler(t *testing.T) {

	testCases := []struct {
		label         string
		expected      cluster.Cluster
		name, version string
		accept        string
		expectedCode  int
		clusterClient *mockClusterManager
	}{
		{
			label:        "Get Cluster with Accept: application/json",
			accept:       "application/json",
			expectedCode: http.StatusOK,
			expected: cluster.Cluster{
				Metadata: types.Metadata{
					Name:        "testCluster",
					Description: "testCluster description",
					UserData1:   "some user data 1",
					UserData2:   "some user data 2",
				},
			},
			name: "testCluster",
			clusterClient: &mockClusterManager{
				//Items that will be returned by the mocked Client
				ClusterItems: []cluster.Cluster{
					{
						Metadata: types.Metadata{
							Name:        "testCluster",
							Description: "testCluster description",
							UserData1:   "some user data 1",
							UserData2:   "some user data 2",
						},
					},
				},
				ClusterContentItems: []cluster.ClusterContent{
					{
						Kubeconfig: "dGVzdCBjb250ZW50cwpvZiBhIGZpbGUgYXR0YWNoZWQKdG8gdGhlIGNyZWF0aW9uCm9mIGNsdXN0ZXJUZXN0Cg==",
					},
				},
			},
		},
		{
			label:        "Get Non-Existing Cluster",
			accept:       "application/json",
			expectedCode: http.StatusNotFound,
			name:         "nonexistingcluster",
			clusterClient: &mockClusterManager{
				ClusterItems: []cluster.Cluster{},
				Err:          pkgerrors.New("cluster not found"),
			},
		},
		{
			label:        "Get Non-Existing Cluster",
			accept:       "application/json",
			expectedCode: http.StatusNotFound,
			name:         "nonexistingcluster",
			clusterClient: &mockClusterManager{
				ClusterItems: []cluster.Cluster{},
				Err:          pkgerrors.New("db Find error"),
			},
		},
		{
			label:        "Get Cluster internal error",
			accept:       "application/json",
			expectedCode: http.StatusInternalServerError,
			name:         "nonexistingcluster",
			clusterClient: &mockClusterManager{
				ClusterItems: []cluster.Cluster{},
				Err:          pkgerrors.New("Internal Error"),
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.label, func(t *testing.T) {
			request := httptest.NewRequest("GET", "/v2/cluster-providers/clusterProvider1/clusters/"+testCase.name, nil)
			if len(testCase.accept) > 0 {
				request.Header.Set("Accept", testCase.accept)
			}
			resp := executeRequest(request, NewRouter(testCase.clusterClient))

			//Check returned code
			if resp.StatusCode != testCase.expectedCode {
				t.Fatalf("Expected %d; Got: %d", testCase.expectedCode, resp.StatusCode)
			}

			//Check returned body only if statusOK
			if resp.StatusCode == http.StatusOK {
				got := cluster.Cluster{}
				json.NewDecoder(resp.Body).Decode(&got)

				if reflect.DeepEqual(testCase.expected, got) == false {
					t.Errorf("listHandler returned unexpected body: got %v;"+
						" expected %v", got, testCase.expected)
				}
			}
		})
	}
}

func TestClusterGetByLabelHandler(t *testing.T) {

	testCases := []struct {
		label         string
		expected      []string
		matchLabel    string
		expectedCode  int
		clusterClient *mockClusterManager
	}{
		{
			label:        "Get Clusters by label",
			expectedCode: http.StatusOK,
			expected:     []string{"cluster1", "cluster2"},
			matchLabel:   "labelA",
			clusterClient: &mockClusterManager{
				//Items that will be returned by the mocked Client
				ClusterList: []string{"cluster1", "cluster2"},
			},
		},
		{
			label:        "Get Clusters by Label, Non-Existing Cluster provider",
			expectedCode: http.StatusNotFound,
			matchLabel:   "labelA",
			clusterClient: &mockClusterManager{
				ClusterList: []string{},
				Err:         pkgerrors.New("cluster provider not found"),
			},
		},
		{
			label:        "Get Clusters by label internal error",
			expectedCode: http.StatusInternalServerError,
			matchLabel:   "labelA",
			clusterClient: &mockClusterManager{
				ClusterItems: []cluster.Cluster{},
				Err:          pkgerrors.New("Internal Error"),
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.label, func(t *testing.T) {
			request := httptest.NewRequest("GET", "/v2/cluster-providers/clusterProvider1/clusters?label="+testCase.matchLabel, nil)
			resp := executeRequest(request, NewRouter(testCase.clusterClient))

			//Check returned code
			if resp.StatusCode != testCase.expectedCode {
				t.Fatalf("Expected %d; Got: %d", testCase.expectedCode, resp.StatusCode)
			}

			//Check returned body only if statusOK
			if resp.StatusCode == http.StatusOK {
				var got []string
				json.NewDecoder(resp.Body).Decode(&got)

				if reflect.DeepEqual(testCase.expected, got) == false {
					t.Errorf("listHandler returned unexpected body: got %v;"+
						" expected %v", got, testCase.expected)
				}
			}
		})
	}
}

func TestClusterGetContentHandler(t *testing.T) {

	testCases := []struct {
		label         string
		expected      string
		name, version string
		accept        string
		expectedCode  int
		clusterClient *mockClusterManager
	}{
		{
			label:        "Get Cluster Content with Accept: application/octet-stream",
			accept:       "application/octet-stream",
			expectedCode: http.StatusOK,
			expected: `test contents
of a file attached
to the creation
of clusterTest
`,
			name: "testCluster",
			clusterClient: &mockClusterManager{
				//Items that will be returned by the mocked Client
				ClusterItems: []cluster.Cluster{
					{
						Metadata: types.Metadata{
							Name:        "testCluster",
							Description: "testCluster description",
							UserData1:   "some user data 1",
							UserData2:   "some user data 2",
						},
					},
				},
				ClusterContentItems: []cluster.ClusterContent{
					{
						Kubeconfig: "dGVzdCBjb250ZW50cwpvZiBhIGZpbGUgYXR0YWNoZWQKdG8gdGhlIGNyZWF0aW9uCm9mIGNsdXN0ZXJUZXN0Cg==",
					},
				},
			},
		},
		{
			label:        "Get Non-Existing Cluster",
			accept:       "application/octet-stream",
			expectedCode: http.StatusNotFound,
			name:         "nonexistingcluster",
			clusterClient: &mockClusterManager{
				ClusterItems: []cluster.Cluster{},
				Err:          pkgerrors.New("content not found"),
			},
		},
		{
			label:        "Get Cluster Content internal error",
			accept:       "application/octet-stream",
			expectedCode: http.StatusInternalServerError,
			name:         "nonexistingcluster",
			clusterClient: &mockClusterManager{
				ClusterItems: []cluster.Cluster{},
				Err:          pkgerrors.New("some internal error"),
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.label, func(t *testing.T) {
			request := httptest.NewRequest("GET", "/v2/cluster-providers/clusterProvider1/clusters/"+testCase.name, nil)
			if len(testCase.accept) > 0 {
				request.Header.Set("Accept", testCase.accept)
			}
			resp := executeRequest(request, NewRouter(testCase.clusterClient))

			//Check returned code
			if resp.StatusCode != testCase.expectedCode {
				t.Fatalf("Expected %d; Got: %d", testCase.expectedCode, resp.StatusCode)
			}

			//Check returned body only if statusOK
			if resp.StatusCode == http.StatusOK {
				body := new(bytes.Buffer)
				body.ReadFrom(resp.Body)
				got := body.String()

				if reflect.DeepEqual(testCase.expected, got) == false {
					t.Errorf("listHandler returned unexpected body: got %v;"+
						" expected %v", got, testCase.expected)
				}
			}
		})
	}
}

func TestClusterDeleteHandler(t *testing.T) {

	testCases := []struct {
		label         string
		name          string
		version       string
		expectedCode  int
		clusterClient *mockClusterManager
	}{
		{
			label:         "Delete Cluster",
			expectedCode:  http.StatusNoContent,
			name:          "testCluster",
			clusterClient: &mockClusterManager{},
		},
		{
			label:        "Delete Cluster with children",
			expectedCode: http.StatusConflict,
			name:         "testCluster",
			clusterClient: &mockClusterManager{
				Err: pkgerrors.New("db Remove error - conflict"),
			},
		},
		{
			label:        "Delete Cluster internal error",
			expectedCode: http.StatusInternalServerError,
			name:         "testCluster",
			clusterClient: &mockClusterManager{
				Err: pkgerrors.New("Internal Error"),
			},
		},
		{
			label:        "Delete Non-Existing Cluster",
			expectedCode: http.StatusNotFound,
			name:         "testCluster",
			clusterClient: &mockClusterManager{
				Err: pkgerrors.New("cluster not found"),
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.label, func(t *testing.T) {
			request := httptest.NewRequest("DELETE", "/v2/cluster-providers/clusterProvider1/clusters/"+testCase.name, nil)
			resp := executeRequest(request, NewRouter(testCase.clusterClient))

			//Check returned code
			if resp.StatusCode != testCase.expectedCode {
				t.Fatalf("Expected %d; Got: %d", testCase.expectedCode, resp.StatusCode)
			}
		})
	}
}

func TestClusterLabelCreateHandler(t *testing.T) {
	testCases := []struct {
		label         string
		reader        io.Reader
		expected      cluster.ClusterLabel
		expectedCode  int
		clusterClient *mockClusterManager
	}{
		{
			label:         "Missing Cluster Label Body Failure",
			expectedCode:  http.StatusBadRequest,
			clusterClient: &mockClusterManager{},
		},
		{
			label:        "Create Cluster Label cluster does not exist",
			expectedCode: http.StatusNotFound,
			reader: bytes.NewBuffer([]byte(`{
					"label-name": "test-label"
				}`)),
			clusterClient: &mockClusterManager{
				Err: pkgerrors.New("Cluster does not exist"),
			},
		},
		{
			label:        "Create Cluster Label",
			expectedCode: http.StatusCreated,
			reader: bytes.NewBuffer([]byte(`{
					"label-name": "test-label"
				}`)),
			expected: cluster.ClusterLabel{
				LabelName: "test-label",
			},
			clusterClient: &mockClusterManager{
				//Items that will be returned by the mocked Client
				ClusterLabelItems: []cluster.ClusterLabel{
					{
						LabelName: "test-label",
					},
				},
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.label, func(t *testing.T) {
			request := httptest.NewRequest("POST", "/v2/cluster-providers/cp1/clusters/cl1/labels", testCase.reader)
			resp := executeRequest(request, NewRouter(testCase.clusterClient))

			//Check returned code
			if resp.StatusCode != testCase.expectedCode {
				t.Fatalf("Expected %d; Got: %d", testCase.expectedCode, resp.StatusCode)
			}

			//Check returned body only if statusCreated
			if resp.StatusCode == http.StatusCreated {
				got := cluster.ClusterLabel{}
				json.NewDecoder(resp.Body).Decode(&got)

				if reflect.DeepEqual(testCase.expected, got) == false {
					t.Errorf("createHandler returned unexpected body: got %v;"+
						" expected %v", got, testCase.expected)
				}
			}
		})
	}
}

func TestClusterLabelPutHandler(t *testing.T) {
	testCases := []struct {
		name          string
		label         string
		reader        io.Reader
		expected      cluster.ClusterLabel
		expectedCode  int
		clusterClient *mockClusterManager
	}{
		{
			name:          "test-label",
			label:         "Missing Cluster Label Body Failure",
			expectedCode:  http.StatusBadRequest,
			clusterClient: &mockClusterManager{},
		},
		{
			name:         "test-label",
			label:        "Cluster not found",
			expectedCode: http.StatusNotFound,
			reader: bytes.NewBuffer([]byte(`{
					"label-name": "test-label"
				}`)),
			expected: cluster.ClusterLabel{
				LabelName: "test-label",
			},
			clusterClient: &mockClusterManager{
				Err: pkgerrors.New("Cluster does not exist"),
			},
		},
		{
			name:         "test-labelABC",
			label:        "Mismatched Cluster Label Update",
			expectedCode: http.StatusBadRequest,
			reader: bytes.NewBuffer([]byte(`{
					"label-name": "test-label"
				}`)),
			expected: cluster.ClusterLabel{
				LabelName: "test-label",
			},
			clusterClient: &mockClusterManager{
				//Items that will be returned by the mocked Client
				ClusterLabelItems: []cluster.ClusterLabel{
					{
						LabelName: "test-label",
					},
				},
			},
		},
		{
			name:         "test-label",
			label:        "Update Cluster Label",
			expectedCode: http.StatusCreated,
			reader: bytes.NewBuffer([]byte(`{
					"label-name": "test-label"
				}`)),
			expected: cluster.ClusterLabel{
				LabelName: "test-label",
			},
			clusterClient: &mockClusterManager{
				//Items that will be returned by the mocked Client
				ClusterLabelItems: []cluster.ClusterLabel{
					{
						LabelName: "test-label",
					},
				},
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.label, func(t *testing.T) {
			request := httptest.NewRequest("PUT", "/v2/cluster-providers/cp1/clusters/cl1/labels/"+testCase.name, testCase.reader)
			resp := executeRequest(request, NewRouter(testCase.clusterClient))

			//Check returned code
			if resp.StatusCode != testCase.expectedCode {
				t.Fatalf("Expected %d; Got: %d", testCase.expectedCode, resp.StatusCode)
			}

			//Check returned body only if statusCreated
			if resp.StatusCode == http.StatusCreated {
				got := cluster.ClusterLabel{}
				json.NewDecoder(resp.Body).Decode(&got)

				if reflect.DeepEqual(testCase.expected, got) == false {
					t.Errorf("createHandler returned unexpected body: got %v;"+
						" expected %v", got, testCase.expected)
				}
			}
		})
	}
}

func TestClusterLabelsGetHandler(t *testing.T) {

	testCases := []struct {
		label         string
		expected      []cluster.ClusterLabel
		name, version string
		expectedCode  int
		clusterClient *mockClusterManager
	}{
		{
			label:        "Get Cluster Labels",
			expectedCode: http.StatusOK,
			expected: []cluster.ClusterLabel{
				{
					LabelName: "test-label1",
				},
				{
					LabelName: "test-label-two",
				},
				{
					LabelName: "test-label-3",
				},
			},
			clusterClient: &mockClusterManager{
				//Items that will be returned by the mocked Client
				ClusterLabelItems: []cluster.ClusterLabel{
					{
						LabelName: "test-label1",
					},
					{
						LabelName: "test-label-two",
					},
					{
						LabelName: "test-label-3",
					},
				},
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.label, func(t *testing.T) {
			request := httptest.NewRequest("GET", "/v2/cluster-providers/cp1/clusters/cl1/labels", nil)
			resp := executeRequest(request, NewRouter(testCase.clusterClient))

			//Check returned code
			if resp.StatusCode != testCase.expectedCode {
				t.Fatalf("Expected %d; Got: %d", testCase.expectedCode, resp.StatusCode)
			}

			//Check returned body only if statusOK
			if resp.StatusCode == http.StatusOK {
				got := []cluster.ClusterLabel{}
				json.NewDecoder(resp.Body).Decode(&got)

				if reflect.DeepEqual(testCase.expected, got) == false {
					t.Errorf("listHandler returned unexpected body: got %v;"+
						" expected %v", got, testCase.expected)
				}
			}
		})
	}
}

func TestClusterLabelGetHandler(t *testing.T) {

	testCases := []struct {
		label         string
		expected      cluster.ClusterLabel
		name, version string
		expectedCode  int
		clusterClient *mockClusterManager
	}{
		{
			label:        "Get Cluster Label",
			expectedCode: http.StatusOK,
			expected: cluster.ClusterLabel{
				LabelName: "testlabel",
			},
			name: "testlabel",
			clusterClient: &mockClusterManager{
				//Items that will be returned by the mocked Client
				ClusterLabelItems: []cluster.ClusterLabel{
					{
						LabelName: "testlabel",
					},
				},
			},
		},
		{
			label:        "Get Non-Existing Cluster Label",
			expectedCode: http.StatusNotFound,
			name:         "nonexistingclusterlabel",
			clusterClient: &mockClusterManager{
				ClusterLabelItems: []cluster.ClusterLabel{},
				Err:               pkgerrors.New("Label not found"),
			},
		},
		{
			label:        "Get Non-Existing Cluster Label",
			expectedCode: http.StatusNotFound,
			name:         "nonexistingclusterlabel",
			clusterClient: &mockClusterManager{
				ClusterLabelItems: []cluster.ClusterLabel{},
				Err:               pkgerrors.New("db Find error"),
			},
		},
		{
			label:        "Get Cluster Label server error",
			expectedCode: http.StatusInternalServerError,
			name:         "nonexistingclusterlabel",
			clusterClient: &mockClusterManager{
				ClusterLabelItems: []cluster.ClusterLabel{},
				Err:               pkgerrors.New("Internal Error"),
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.label, func(t *testing.T) {
			request := httptest.NewRequest("GET", "/v2/cluster-providers/clusterProvider1/clusters/cl1/labels/"+testCase.name, nil)
			resp := executeRequest(request, NewRouter(testCase.clusterClient))

			//Check returned code
			if resp.StatusCode != testCase.expectedCode {
				t.Fatalf("Expected %d; Got: %d", testCase.expectedCode, resp.StatusCode)
			}

			//Check returned body only if statusOK
			if resp.StatusCode == http.StatusOK {
				got := cluster.ClusterLabel{}
				json.NewDecoder(resp.Body).Decode(&got)

				if reflect.DeepEqual(testCase.expected, got) == false {
					t.Errorf("listHandler returned unexpected body: got %v;"+
						" expected %v", got, testCase.expected)
				}
			}
		})
	}
}

func TestClusterLabelDeleteHandler(t *testing.T) {

	testCases := []struct {
		label         string
		name          string
		version       string
		expectedCode  int
		clusterClient *mockClusterManager
	}{
		{
			label:         "Delete Cluster Label",
			expectedCode:  http.StatusNoContent,
			name:          "testClusterLabel",
			clusterClient: &mockClusterManager{},
		},
		{
			label:        "Delete Non-Existing Cluster Label",
			expectedCode: http.StatusNotFound,
			name:         "testClusterLabel",
			clusterClient: &mockClusterManager{
				Err: pkgerrors.New("cluster label not found"),
			},
		},
		{
			label:        "Delete Cluster Label internal error",
			expectedCode: http.StatusInternalServerError,
			name:         "testClusterLabel",
			clusterClient: &mockClusterManager{
				Err: pkgerrors.New("Internal Error"),
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.label, func(t *testing.T) {
			request := httptest.NewRequest("DELETE", "/v2/cluster-providers/cp1/clusters/cl1/labels/"+testCase.name, nil)
			resp := executeRequest(request, NewRouter(testCase.clusterClient))

			//Check returned code
			if resp.StatusCode != testCase.expectedCode {
				t.Fatalf("Expected %d; Got: %d", testCase.expectedCode, resp.StatusCode)
			}
		})
	}
}

func TestClusterKvPairsCreateHandler(t *testing.T) {
	testCases := []struct {
		label         string
		reader        io.Reader
		expected      cluster.ClusterKvPairs
		expectedCode  int
		clusterClient *mockClusterManager
	}{
		{
			label:         "Missing Cluster KvPairs Body Failure",
			expectedCode:  http.StatusBadRequest,
			clusterClient: &mockClusterManager{},
		},
		{
			label:        "Create Cluster KvPairs",
			expectedCode: http.StatusCreated,
			reader: bytes.NewBuffer([]byte(`{
					"metadata": {
						"name": "ClusterKvPair1",
						"description": "test cluster kv pairs",
						"userData1": "some user data 1",
						"userData2": "some user data 2"
					},
					"spec": {
						"kv": [
							{
								"key1": "value1"
							},
							{
								"key2": "value2"
							}
						]
					}
				}`)),
			expected: cluster.ClusterKvPairs{
				Metadata: types.Metadata{
					Name:        "ClusterKvPair1",
					Description: "test cluster kv pairs",
					UserData1:   "some user data 1",
					UserData2:   "some user data 2",
				},
				Spec: cluster.ClusterKvSpec{
					Kv: []map[string]interface{}{
						{
							"key1": "value1",
						},
						{
							"key2": "value2",
						},
					},
				},
			},
			clusterClient: &mockClusterManager{
				//Items that will be returned by the mocked Client
				ClusterKvPairsItems: []cluster.ClusterKvPairs{
					{
						Metadata: types.Metadata{
							Name:        "ClusterKvPair1",
							Description: "test cluster kv pairs",
							UserData1:   "some user data 1",
							UserData2:   "some user data 2",
						},
						Spec: cluster.ClusterKvSpec{
							Kv: []map[string]interface{}{
								{
									"key1": "value1",
								},
								{
									"key2": "value2",
								},
							},
						},
					},
				},
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.label, func(t *testing.T) {
			request := httptest.NewRequest("POST", "/v2/cluster-providers/cp1/clusters/cl1/kv-pairs", testCase.reader)
			resp := executeRequest(request, NewRouter(testCase.clusterClient))

			//Check returned code
			if resp.StatusCode != testCase.expectedCode {
				t.Fatalf("Expected %d; Got: %d", testCase.expectedCode, resp.StatusCode)
			}

			//Check returned body only if statusCreated
			if resp.StatusCode == http.StatusCreated {
				got := cluster.ClusterKvPairs{}
				json.NewDecoder(resp.Body).Decode(&got)

				if reflect.DeepEqual(testCase.expected, got) == false {
					t.Errorf("createHandler returned unexpected body: got %v;"+
						" expected %v", got, testCase.expected)
				}
			}
		})
	}
}

func TestClusterKvPairsPutHandler(t *testing.T) {
	testCases := []struct {
		name          string
		label         string
		reader        io.Reader
		expected      cluster.ClusterKvPairs
		expectedCode  int
		clusterClient *mockClusterManager
	}{
		{
			name:          "ClusterKvPair1",
			label:         "Missing Cluster KvPairs Body Failure",
			expectedCode:  http.StatusBadRequest,
			clusterClient: &mockClusterManager{},
		},
		{
			label: "Missing Cluster KV Pair Name in Request Body",
			name:  "ClusterKvPair1",
			reader: bytes.NewBuffer([]byte(`{
					"metadata": {
						"description": "this is a test cluster kvpair",
						"userData1": "some user data 1",
						"userData2": "some user data 2"
					}
				}`)),
			expectedCode:  http.StatusBadRequest,
			clusterClient: &mockClusterManager{},
		},
		{
			label:        "Mismatched Cluster KV Pair Name in Request",
			name:         "ClusterKvPair1ABC",
			expectedCode: http.StatusBadRequest,
			reader: bytes.NewBuffer([]byte(`{
					"metadata": {
						"name": "ClusterKvPair1",
						"description": "test cluster kv pairs",
						"userData1": "some user data 1",
						"userData2": "some user data 2"
					},
					"spec": {
						"kv": [
							{
								"key1": "value1"
							},
							{
								"key2": "value2"
							}
						]
					}
				}`)),
			clusterClient: &mockClusterManager{},
		},
		{
			name:         "ClusterKvPair1",
			label:        "Update Cluster KvPairs",
			expectedCode: http.StatusCreated,
			reader: bytes.NewBuffer([]byte(`{
					"metadata": {
						"name": "ClusterKvPair1",
						"description": "test cluster kv pairs",
						"userData1": "some user data 1",
						"userData2": "some user data 2"
					},
					"spec": {
						"kv": [
							{
								"key1": "value1"
							},
							{
								"key2": "value2"
							}
						]
					}
				}`)),
			expected: cluster.ClusterKvPairs{
				Metadata: types.Metadata{
					Name:        "ClusterKvPair1",
					Description: "test cluster kv pairs",
					UserData1:   "some user data 1",
					UserData2:   "some user data 2",
				},
				Spec: cluster.ClusterKvSpec{
					Kv: []map[string]interface{}{
						{
							"key1": "value1",
						},
						{
							"key2": "value2",
						},
					},
				},
			},
			clusterClient: &mockClusterManager{
				//Items that will be returned by the mocked Client
				ClusterKvPairsItems: []cluster.ClusterKvPairs{
					{
						Metadata: types.Metadata{
							Name:        "ClusterKvPair1",
							Description: "test cluster kv pairs",
							UserData1:   "some user data 1",
							UserData2:   "some user data 2",
						},
						Spec: cluster.ClusterKvSpec{
							Kv: []map[string]interface{}{
								{
									"key1": "value1",
								},
								{
									"key2": "value2",
								},
							},
						},
					},
				},
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.label, func(t *testing.T) {
			request := httptest.NewRequest("PUT", "/v2/cluster-providers/cp1/clusters/cl1/kv-pairs/"+testCase.name, testCase.reader)
			resp := executeRequest(request, NewRouter(testCase.clusterClient))

			//Check returned code
			if resp.StatusCode != testCase.expectedCode {
				t.Fatalf("Expected %d; Got: %d", testCase.expectedCode, resp.StatusCode)
			}

			//Check returned body only if statusCreated
			if resp.StatusCode == http.StatusCreated {
				got := cluster.ClusterKvPairs{}
				json.NewDecoder(resp.Body).Decode(&got)

				if reflect.DeepEqual(testCase.expected, got) == false {
					t.Errorf("createHandler returned unexpected body: got %v;"+
						" expected %v", got, testCase.expected)
				}
			}
		})
	}
}

func TestClusterKvPairsGetAllHandler(t *testing.T) {

	testCases := []struct {
		label         string
		expected      []cluster.ClusterKvPairs
		name, version string
		expectedCode  int
		clusterClient *mockClusterManager
	}{
		{
			label:        "Get Cluster KvPairs",
			expectedCode: http.StatusOK,
			expected: []cluster.ClusterKvPairs{
				{
					Metadata: types.Metadata{
						Name:        "ClusterKvPair1",
						Description: "test cluster kv pairs",
						UserData1:   "some user data 1",
						UserData2:   "some user data 2",
					},
					Spec: cluster.ClusterKvSpec{
						Kv: []map[string]interface{}{
							{
								"key1": "value1",
							},
							{
								"key2": "value2",
							},
						},
					},
				},
				{
					Metadata: types.Metadata{
						Name:        "ClusterKvPair2",
						Description: "test cluster kv pairs",
						UserData1:   "some user data A",
						UserData2:   "some user data B",
					},
					Spec: cluster.ClusterKvSpec{
						Kv: []map[string]interface{}{
							{
								"keyA": "valueA",
							},
							{
								"keyB": "valueB",
							},
						},
					},
				},
			},
			clusterClient: &mockClusterManager{
				//Items that will be returned by the mocked Client
				ClusterKvPairsItems: []cluster.ClusterKvPairs{
					{
						Metadata: types.Metadata{
							Name:        "ClusterKvPair1",
							Description: "test cluster kv pairs",
							UserData1:   "some user data 1",
							UserData2:   "some user data 2",
						},
						Spec: cluster.ClusterKvSpec{
							Kv: []map[string]interface{}{
								{
									"key1": "value1",
								},
								{
									"key2": "value2",
								},
							},
						},
					},
					{
						Metadata: types.Metadata{
							Name:        "ClusterKvPair2",
							Description: "test cluster kv pairs",
							UserData1:   "some user data A",
							UserData2:   "some user data B",
						},
						Spec: cluster.ClusterKvSpec{
							Kv: []map[string]interface{}{
								{
									"keyA": "valueA",
								},
								{
									"keyB": "valueB",
								},
							},
						},
					},
				},
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.label, func(t *testing.T) {
			request := httptest.NewRequest("GET", "/v2/cluster-providers/cp1/clusters/cl1/kv-pairs", nil)
			resp := executeRequest(request, NewRouter(testCase.clusterClient))

			//Check returned code
			if resp.StatusCode != testCase.expectedCode {
				t.Fatalf("Expected %d; Got: %d", testCase.expectedCode, resp.StatusCode)
			}

			//Check returned body only if statusOK
			if resp.StatusCode == http.StatusOK {
				got := []cluster.ClusterKvPairs{}
				json.NewDecoder(resp.Body).Decode(&got)

				if reflect.DeepEqual(testCase.expected, got) == false {
					t.Errorf("listHandler returned unexpected body: got %v;"+
						" expected %v", got, testCase.expected)
				}
			}
		})
	}
}

func TestClusterKvPairsGetValueHandler(t *testing.T) {

	testCases := []struct {
		label         string
		expected      string
		name, version string
		key           string
		expectedCode  int
		clusterClient *mockClusterManager
	}{
		{
			label:        "Get Cluster KV Pairs key value",
			expectedCode: http.StatusOK,
			expected:     "valueB",
			name:         "ClusterKvPair2",
			key:          "keyB",
			clusterClient: &mockClusterManager{
				//Items that will be returned by the mocked Client
				ClusterKvPairsItems: []cluster.ClusterKvPairs{
					{
						Metadata: types.Metadata{
							Name:        "ClusterKvPair2",
							Description: "test cluster kv pairs",
							UserData1:   "some user data A",
							UserData2:   "some user data B",
						},
						Spec: cluster.ClusterKvSpec{
							Kv: []map[string]interface{}{
								{
									"keyA": "valueA",
								},
								{
									"keyB": "valueB",
								},
							},
						},
					},
				},
			},
		},
		{
			label:        "Get Non-Existing Cluster KV Pairs",
			expectedCode: http.StatusNotFound,
			name:         "nonexistingclusterkvpairs",
			key:          "keyB",
			clusterClient: &mockClusterManager{
				ClusterKvPairsItems: []cluster.ClusterKvPairs{},
				Err:                 pkgerrors.New("Cluster key value pair not found"),
			},
		},
		{
			label:        "Get Non-Existing Cluster KV Pairs",
			expectedCode: http.StatusNotFound,
			name:         "nonexistingclusterkvpairs",
			key:          "keyB",
			clusterClient: &mockClusterManager{
				ClusterKvPairsItems: []cluster.ClusterKvPairs{},
				Err:                 pkgerrors.New("db Find error"),
			},
		},
		{
			label:        "Get Cluster KV Pairs internal error",
			expectedCode: http.StatusInternalServerError,
			name:         "nonexistingclusterkvpairs",
			key:          "keyB",
			clusterClient: &mockClusterManager{
				ClusterKvPairsItems: []cluster.ClusterKvPairs{},
				Err:                 pkgerrors.New("Internal error"),
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.label, func(t *testing.T) {
			request := httptest.NewRequest("GET", "/v2/cluster-providers/clusterProvider1/clusters/cl1/kv-pairs/"+testCase.name+"?key="+testCase.key, nil)
			resp := executeRequest(request, NewRouter(testCase.clusterClient))

			//Check returned code
			if resp.StatusCode != testCase.expectedCode {
				t.Fatalf("Expected %d; Got: %d", testCase.expectedCode, resp.StatusCode)
			}

			//Check returned body only if statusOK
			if resp.StatusCode == http.StatusOK {
				var got string
				json.NewDecoder(resp.Body).Decode(&got)

				if reflect.DeepEqual(testCase.expected, got) == false {
					t.Errorf("listHandler returned unexpected body: got %v;"+
						" expected %v", got, testCase.expected)
				}
			}
		})
	}
}

func TestClusterKvPairsGetHandler(t *testing.T) {

	testCases := []struct {
		label         string
		expected      cluster.ClusterKvPairs
		name, version string
		expectedCode  int
		clusterClient *mockClusterManager
	}{
		{
			label:        "Get Cluster KV Pairs",
			expectedCode: http.StatusOK,
			expected: cluster.ClusterKvPairs{
				Metadata: types.Metadata{
					Name:        "ClusterKvPair2",
					Description: "test cluster kv pairs",
					UserData1:   "some user data A",
					UserData2:   "some user data B",
				},
				Spec: cluster.ClusterKvSpec{
					Kv: []map[string]interface{}{
						{
							"keyA": "valueA",
						},
						{
							"keyB": "valueB",
						},
					},
				},
			},
			name: "ClusterKvPair2",
			clusterClient: &mockClusterManager{
				//Items that will be returned by the mocked Client
				ClusterKvPairsItems: []cluster.ClusterKvPairs{
					{
						Metadata: types.Metadata{
							Name:        "ClusterKvPair2",
							Description: "test cluster kv pairs",
							UserData1:   "some user data A",
							UserData2:   "some user data B",
						},
						Spec: cluster.ClusterKvSpec{
							Kv: []map[string]interface{}{
								{
									"keyA": "valueA",
								},
								{
									"keyB": "valueB",
								},
							},
						},
					},
				},
			},
		},
		{
			label:        "Get Non-Existing Cluster KV Pairs",
			expectedCode: http.StatusNotFound,
			name:         "nonexistingclusterkvpairs",
			clusterClient: &mockClusterManager{
				ClusterKvPairsItems: []cluster.ClusterKvPairs{},
				Err:                 pkgerrors.New("not found"),
			},
		},
		{
			label:        "Get Non-Existing Cluster KV Pairs - part II",
			expectedCode: http.StatusNotFound,
			name:         "nonexistingclusterkvpairs",
			clusterClient: &mockClusterManager{
				ClusterKvPairsItems: []cluster.ClusterKvPairs{},
				Err:                 pkgerrors.New("db Find error"),
			},
		},
		{
			label:        "Get Cluster KV Pairs internal error",
			expectedCode: http.StatusInternalServerError,
			name:         "nonexistingclusterkvpairs",
			clusterClient: &mockClusterManager{
				ClusterKvPairsItems: []cluster.ClusterKvPairs{},
				Err:                 pkgerrors.New("some other error"),
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.label, func(t *testing.T) {
			request := httptest.NewRequest("GET", "/v2/cluster-providers/clusterProvider1/clusters/cl1/kv-pairs/"+testCase.name, nil)
			resp := executeRequest(request, NewRouter(testCase.clusterClient))

			//Check returned code
			if resp.StatusCode != testCase.expectedCode {
				t.Fatalf("Expected %d; Got: %d", testCase.expectedCode, resp.StatusCode)
			}

			//Check returned body only if statusOK
			if resp.StatusCode == http.StatusOK {
				got := cluster.ClusterKvPairs{}
				json.NewDecoder(resp.Body).Decode(&got)

				if reflect.DeepEqual(testCase.expected, got) == false {
					t.Errorf("listHandler returned unexpected body: got %v;"+
						" expected %v", got, testCase.expected)
				}
			}
		})
	}
}

func TestClusterKvPairsDeleteHandler(t *testing.T) {

	testCases := []struct {
		label         string
		name          string
		version       string
		expectedCode  int
		clusterClient *mockClusterManager
	}{
		{
			label:         "Delete Cluster KV Pairs",
			expectedCode:  http.StatusNoContent,
			name:          "testClusterKvPairs",
			clusterClient: &mockClusterManager{},
		},
		{
			label:        "Delete Non-Existing Cluster KV Pairs",
			expectedCode: http.StatusNotFound,
			name:         "testClusterKvPairs",
			clusterClient: &mockClusterManager{
				Err: pkgerrors.New("kv pair not found"),
			},
		},
		{
			label:        "Delete Cluster KV Pairs internal error",
			expectedCode: http.StatusInternalServerError,
			name:         "testClusterKvPairs",
			clusterClient: &mockClusterManager{
				Err: pkgerrors.New("Internal Error"),
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.label, func(t *testing.T) {
			request := httptest.NewRequest("DELETE", "/v2/cluster-providers/cp1/clusters/cl1/kv-pairs/"+testCase.name, nil)
			resp := executeRequest(request, NewRouter(testCase.clusterClient))

			//Check returned code
			if resp.StatusCode != testCase.expectedCode {
				t.Fatalf("Expected %d; Got: %d", testCase.expectedCode, resp.StatusCode)
			}
		})
	}
}
