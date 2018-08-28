// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2020 Intel Corporation

package helm

import (
	"io/ioutil"
	"log"
	"path/filepath"

	"github.com/ghodss/yaml"
	pkgerrors "github.com/pkg/errors"
)

/*
#Sample Yaml format for profile manifest.yaml
---
version: v1
type:
  values: "values_override.yaml"
  configresource:
    - filepath: config.yaml
      chartpath: chart/config/resources/config.yaml
    - filepath: config2.yaml
      chartpath: chart/config/resources/config2.yaml
*/

type overrideFiles struct {
	FilePath  string `yaml:"filepath"`
	ChartPath string `yaml:"chartpath"`
}

type supportedOverrides struct {
	ConfigResource []overrideFiles `yaml:"configresource"`
	Values         string          `yaml:"values"`
}

type profileOverride struct {
	Version string             `yaml:"version"`
	Type    supportedOverrides `yaml:"type"`
}

type ProfileYamlClient struct {
	path     string
	override profileOverride
}

func (p ProfileYamlClient) Print() {
	log.Println(p.override)
}

//GetValues returns a path to the override values.yam
//that was part of the profile
func (p ProfileYamlClient) GetValues() string {
	return filepath.Join(p.path, p.override.Type.Values)
}

//CopyConfigurationOverrides copies the various files that are
//provided as overrides to their corresponding locations within
//the destination chart.
func (p ProfileYamlClient) CopyConfigurationOverrides(chartPath string) error {

	//Iterate over each configresource and copy that file into
	//the respective path in the chart.
	for _, v := range p.override.Type.ConfigResource {
		data, err := ioutil.ReadFile(filepath.Join(p.path, v.FilePath))
		if err != nil {
			return pkgerrors.Wrap(err, "Reading configuration file")
		}
		err = ioutil.WriteFile(filepath.Join(chartPath, v.ChartPath), data, 0600)
		if err != nil {
			return pkgerrors.Wrap(err, "Writing configuration file into chartpath")
		}
	}

	return nil
}

//ProcessProfileYaml parses the manifest.yaml file that is part of the profile
//package and creates the appropriate structures out of it.
func ProcessProfileYaml(fpath string, manifestFileName string) (ProfileYamlClient, error) {

	p := filepath.Join(fpath, manifestFileName)
	data, err := ioutil.ReadFile(p)
	if err != nil {
		return ProfileYamlClient{}, pkgerrors.Wrap(err, "Reading manifest file")
	}

	out := profileOverride{}
	err = yaml.Unmarshal(data, &out)
	if err != nil {
		return ProfileYamlClient{}, pkgerrors.Wrap(err, "Marshaling manifest yaml file")
	}

	return ProfileYamlClient{path: fpath, override: out}, nil
}
