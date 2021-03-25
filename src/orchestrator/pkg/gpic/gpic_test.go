package gpic

import (
	"reflect"
	"testing"
)

func TestGpic(t *testing.T) {
	intentResolverHelper = func(pn, cn, cln string, clusters []ClusterWithName) ([]ClusterWithName, error) {
		if cln == "" && cn != "" {
			eachClusterWithName := ClusterWithName{pn, cn}
			clusters = append(clusters, eachClusterWithName)
		}
		if cn == "" && cln != "" {
			if cln == "east-us1" {
				eachClusterWithName := ClusterWithName{pn, "edge1"}
				clusters = append(clusters, eachClusterWithName)
				eachClusterWithName = ClusterWithName{pn, "edge2"}
				clusters = append(clusters, eachClusterWithName)
				eachClusterWithName = ClusterWithName{pn, "edge3"}
				clusters = append(clusters, eachClusterWithName)

			}
			if cln == "east-us2" {
				eachClusterWithName := ClusterWithName{pn, "edge4"}
				clusters = append(clusters, eachClusterWithName)
				eachClusterWithName = ClusterWithName{pn, "edge5"}
				clusters = append(clusters, eachClusterWithName)

			}
			if cln == "east-us3" {
				eachClusterWithName := ClusterWithName{pn, "edge6"}
				clusters = append(clusters, eachClusterWithName)
				eachClusterWithName = ClusterWithName{pn, "edge7"}
				clusters = append(clusters, eachClusterWithName)
			}
		}
		return clusters, nil
	}
	testCases := []struct {
		label          string
		intent         IntentStruc
		expectedOutput map[string][]string
		expectedError  error
	}{
		{
			intent: IntentStruc{
				AllOfArray: []AllOf{
					{
						ProviderName: "aws",
						ClusterName:  "edge10",
					},
					{
						ProviderName: "aws",
						ClusterName:  "edge11",
					},
					{
						ProviderName:     "aws",
						ClusterLabelName: "east-us3",
					},
					{
						AnyOfArray: []AnyOf{
							{ProviderName: "aws",
								ClusterLabelName: "east-us1"},
							{ProviderName: "aws",
								ClusterLabelName: "east-us2"},
						},
					},
					{
						AnyOfArray: []AnyOf{
							{ProviderName: "aws",
								ClusterName: "edge8"},
							{ProviderName: "aws",
								ClusterName: "edge9"},
						},
					},
				},
				AnyOfArray: []AnyOf{},
			},
			expectedOutput: map[string][]string{"1": {"awsedge10"},
				"2": {"awsedge11"},
				"3": {"awsedge6"},
				"4": {"awsedge7"},
				"5": {"awsedge1", "awsedge2", "awsedge3", "awsedge4", "awsedge5"},
				"6": {"awsedge8", "awsedge9"}},
			expectedError: nil,
			label:         "Resolve clusters",
		},
		{
			intent: IntentStruc{
				AnyOfArray: []AnyOf{
					{ProviderName: "aws",
						ClusterName: "edge8"},
					{ProviderName: "aws",
						ClusterName: "edge9"},
				},
			},
			expectedOutput: map[string][]string{"1": {"awsedge8", "awsedge9"}},
			expectedError:  nil,
			label:          "Resolve Anyof clusters",
		},
		{
			intent: IntentStruc{
				AnyOfArray: []AnyOf{
					{ProviderName: "aws",
								ClusterLabelName: "east-us1"},
				},
			},
			expectedOutput: map[string][]string{"1": {"awsedge1", "awsedge2", "awsedge3"}},
			expectedError:  nil,
			label:          "Resolve Anyof clusters with labels",
		},
		{
			intent: IntentStruc{
				AnyOfArray: []AnyOf{
					{ProviderName: "aws",
					ClusterLabelName: "east-us1",
				    },
					{ProviderName: "aws",
						ClusterName: "edge8"},
				},
			},
			expectedOutput: map[string][]string{"1": {"awsedge1", "awsedge2", "awsedge3", "awsedge8"}},
			expectedError:  nil,
			label:          "Resolve Anyof clusters with labels and names",
		},
		{
			intent: IntentStruc{
				AllOfArray: []AllOf{
					{
						ProviderName: "aws",
						ClusterName:  "edge10",
					},
					{
						ProviderName: "aws",
						ClusterName:  "edge11",
					},
					{
						ProviderName:     "aws",
						ClusterLabelName: "east-us1",
					},
					{
						AnyOfArray: []AnyOf{
							{ProviderName: "aws",
							ClusterLabelName: "east-us3",
								},
							{ProviderName: "aws",
								ClusterName: "edge12"},
						},
					},
					{
						AnyOfArray: []AnyOf{
							{ProviderName: "aws",
								ClusterLabelName: "east-us2"},
							{ProviderName: "aws",
								ClusterName: "edge8"},
							{ProviderName: "aws",
								ClusterName: "edge9"},
						},
					},
				},
				AnyOfArray: []AnyOf{},
			},
			expectedOutput: map[string][]string{"1": {"awsedge10"},
				"2": {"awsedge11"},
				"3": {"awsedge1"},
				"4": {"awsedge2"},
				"5": {"awsedge3"},
				"6": {"awsedge6", "awsedge7", "awsedge12"},
				"7": {"awsedge4", "awsedge5", "awsedge8", "awsedge9"}},
			expectedError: nil,
			label:         "Resolve clusters with labels and names",
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.label, func(t *testing.T) {
			l, err := IntentResolver(testCase.intent)
			if err != testCase.expectedError {
				t.Fatalf("Expected %d; Got: %d", testCase.expectedError, err)
			}
			got := make(map[string][]string)
			for _, cg := range l.OptionalClusters {
				gn := cg.GroupNumber
				for _, eachCluster := range cg.Clusters {
					n := eachCluster.ProviderName + eachCluster.ClusterName
					got[gn] = append(got[gn], n)
				}
			}
			for _, cg := range l.MandatoryClusters {
				gn := cg.GroupNumber
				for _, eachCluster := range cg.Clusters {
					n := eachCluster.ProviderName + eachCluster.ClusterName
					got[gn] = append(got[gn], n)
				}
			}
			if reflect.DeepEqual(testCase.expectedOutput, got) == false {
				t.Errorf("createHandler returned unexpected body: got %+v;"+
					" expected %+v", got, testCase.expectedOutput)
			}
		})
	}
}
