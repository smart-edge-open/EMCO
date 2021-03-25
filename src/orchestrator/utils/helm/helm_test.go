// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2020 Intel Corporation

package helm

import (
	"crypto/sha256"
	"fmt"

	"io/ioutil"
	"path/filepath"
	"strings"
	"gopkg.in/yaml.v2"

	"testing"
)

func TestProcessValues(t *testing.T) {

	chartDir := "../../mock_files/mock_charts/testchart2"
	profileDir := "../../mock_files/mock_profiles/profile1"

	testCases := []struct {
		label         string
		valueFiles    []string
		values        []string
		expectedHash  string
		expectedError string
	}{
		{
			label: "Process Values with Value Files Override",
			valueFiles: []string{
				filepath.Join(chartDir, "values.yaml"),
				filepath.Join(profileDir, "override_values.yaml"),
			},
			//Hash of a combined values.yaml file that is expected
			expectedHash:  "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
			expectedError: "",
		},
		{
			label: "Process Values with Values Pair Override",
			valueFiles: []string{
				filepath.Join(chartDir, "values.yaml"),
			},
			//Use the same convention as specified in helm template --set
			values: []string{
				"service.externalPort=82",
			},
			//Hash of a combined values.yaml file that is expected
			expectedHash:  "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
			expectedError: "",
		},
		{
			label: "Process Values with Both Overrides",
			valueFiles: []string{
				filepath.Join(chartDir, "values.yaml"),
				filepath.Join(profileDir, "override_values.yaml"),
			},
			//Use the same convention as specified in helm template --set
			//Key takes precedence over the value from override_values.yaml
			values: []string{
				"service.externalPort=82",
			},
			//Hash of a combined values.yaml file that is expected
			expectedHash:  "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
			expectedError: "",
		},
		{
			label: "Process complex Pair Override",
			values: []string{
				"name={a,b,c}",
				"servers[0].port=80",
			},
			expectedError: "",
			expectedHash:  "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
		},
	}

	h := sha256.New()

	for _, testCase := range testCases {
		t.Run(testCase.label, func(t *testing.T) {
			tc := NewTemplateClient("1.12.3", "testnamespace", "testreleasename", "manifest.yaml")
			out, err := tc.processValues(testCase.valueFiles, testCase.values)
			if err != nil {
				if testCase.expectedError == "" {
					t.Fatalf("Got an error %s", err)
				}
				if strings.Contains(err.Error(), testCase.expectedError) == false {
					t.Fatalf("Got unexpected error message %s", err)
				}
			} else {
				//Compute the hash of returned data and compare
				gotHash := fmt.Sprintf("%x", h.Sum(nil))
				h.Reset()
				if gotHash != testCase.expectedHash {
					mout, _ := yaml.Marshal(&out)
					t.Fatalf("Got unexpected hash '%s' of values.yaml:\n%v", gotHash, string(mout))
				}
			}
		})
	}
}

func TestGenerateKubernetesArtifacts(t *testing.T) {

	chartDir := "../../mock_files/mock_charts/testchart2"
	profileDir := "../../mock_files/mock_profiles/profile1"

	testCases := []struct {
		label           string
		chartPath       string
		valueFiles      []string
		values          []string
		expectedHashMap map[string]string
		expectedError   string
	}{
		{
			label:      "Generate artifacts without any overrides",
			chartPath:  chartDir,
			valueFiles: []string{},
			values:     []string{},
			//sha256 hash of the evaluated templates in each chart
			expectedHashMap: map[string]string{
				"/tmp/helm-tmpl-766285534/manifest-0": "fcc1083ace82b633e3a0a687d50f532c07e1212b7a42b2c178b65e5768fffcfe",
				"/tmp/helm-tmpl-490085794/manifest-2": "eefeac6ff5430a16a32ae3974857cbe5ff516a1a68566e5edcddd410d60397c0",
				"/tmp/helm-tmpl-522092734/manifest-1": "b88aa963ee3afb9676e9930519d7caa103df1251da48a9351ab4ac0c5730d2af",
			},
			expectedError: "",
		},
		{
			label:     "Generate artifacts with overrides",
			chartPath: chartDir,
			valueFiles: []string{
				filepath.Join(profileDir, "override_values.yaml"),
			},
			values: []string{
				"service.externalPort=82",
			},
			//sha256 hash of the evaluated templates in each chart
			expectedHashMap: map[string]string{
				"/tmp/helm-tmpl-766285534/manifest-0": "fcc1083ace82b633e3a0a687d50f532c07e1212b7a42b2c178b65e5768fffcfe",
				"/tmp/helm-tmpl-562098139/manifest-2": "03ae530e49071d005be78f581b7c06c59119f91f572b28c0c0c06ced8e37bf6e",
				"/tmp/helm-tmpl-522092734/manifest-1": "b88aa963ee3afb9676e9930519d7caa103df1251da48a9351ab4ac0c5730d2af",
			},
			expectedError: "",
		},
		{
			label:      "Generate artifacts without any overrides http-server",
			chartPath:  "../../../../kud/tests/helm_charts/dtc/http-server",
			valueFiles: []string{},
			values:     []string{},
			//sha256 hash of the evaluated templates in each chart
			expectedHashMap: map[string]string{
				"/tmp/helm-tmpl-766285534/manifest-0": "81ef115271f6579f6346c5bf909553e139864d9938e3eea82ad50cf6dedc1ab9",
				"/tmp/helm-tmpl-490085794/manifest-2": "0f7ac458db24f2bdf7bd3ad8df3b0cb3e19e61c0db96d73df617d2ca15bf936d",
				"/tmp/helm-tmpl-522092734/manifest-1": "4ba5336b0cdd3c8d23ab60fc3e4680588bf5101dc774c177ef281485ddf0790c",
			},
			expectedError: "",
		},
		{
			label:      "Generate artifacts without any overrides prometheus",
			chartPath:  "../../../../kud/tests/vnfs/comp-app/collection/app2/helm/prometheus-operator",
			valueFiles: []string{},
			values:     []string{},
			//sha256 hash of the evaluated templates in each chart
			expectedHashMap: map[string]string{
				"/tmp/helm-tmpl-766285534/manifest-0": "4a24e02bf57db191719ef54ec08b0fc5b9716e7090a8f51e00723903b60fa6cb",
				"/tmp/helm-tmpl-490085794/manifest-2": "fda4f06ac6129819613011875734d1405da306a4e2397fe070c082ade78b8d07",
				"/tmp/helm-tmpl-522092734/manifest-1": "baa2cde5c311128498eb16c6d341bf7aa2308209d159876a19c3f8b16025c9d9",
			},
			expectedError: "",
		},
	}

	h := sha256.New()

	for _, testCase := range testCases {
		t.Run(testCase.label, func(t *testing.T) {
			tc := NewTemplateClient("1.12.3", "testnamespace", "testreleasename", "manifest.yaml")
			out, err := tc.GenerateKubernetesArtifacts(testCase.chartPath, testCase.valueFiles,
				testCase.values)
			if err != nil {
				if testCase.expectedError == "" {
					t.Fatalf("Got an error %s", err)
				}
				if strings.Contains(err.Error(), testCase.expectedError) == false {
					t.Fatalf("Got unexpected error message %s", err)
				}
			} else {
				exists := false
				//Compute the hash of returned data and compare
				for _, v := range out {
					f := v.FilePath
					data, err := ioutil.ReadFile(f)
					if err != nil {
						t.Errorf("Unable to read file %s", v)
					}
					h.Write(data)
					gotHash := fmt.Sprintf("%x", h.Sum(nil))
					h.Reset()

					//Find the right hash from expectedHashMap
					expectedHash := ""
					found := false
					for k1, v1 := range testCase.expectedHashMap {
						// Split filename and use digits after last -
						sp := strings.Split(k1, "-")
						ap := strings.Split(f, "-")
						if len(sp) < 4 || len(ap) < 4 {
							t.Fatalf("Unexpected filenames")
						}
						if sp[3] == ap[3] {
							expectedHash = v1
							found = true
							exists = true
							break
						}
					}
					if found && gotHash != expectedHash {
						t.Fatalf("Got unexpected hash %s for %s", gotHash, f)
					}
				}
				if !exists {
					t.Fatalf("Resources not found in output - GenerateKubernetesArtifacts")
				}
			}
		})
	}
}
