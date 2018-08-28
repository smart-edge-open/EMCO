package jsonapihelper

// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2020 Intel Corporation

import (
	"encoding/json"
	"strings"

	jsonpatch "github.com/evanphx/json-patch"
	log "github.com/open-ness/EMCO/src/orchestrator/pkg/infra/logutils"
	pkgerrors "github.com/pkg/errors"
	yamlV2 "gopkg.in/yaml.v2"
	"sigs.k8s.io/yaml"
)

// ConfigMapResource consists of ApiVersion, Kind, MetaData and Data map
type ConfigMapResource struct {
	APIVersion string            `yaml:"apiVersion"`
	Kind       string            `yaml:"kind"`
	MetaData   MetaDataStr       `yaml:"metadata"`
	Data       map[string]string `yaml:"data"`
}

// SecretResource consists of ApiVersion, Kind, MetaData, type and Data map
type SecretResource struct {
	APIVersion string            `yaml:"apiVersion"`
	Kind       string            `yaml:"kind"`
	MetaData   MetaDataStr       `yaml:"metadata"`
	Type       string            `yaml:"type"`
	Data       map[string]string `yaml:"data"`
}

// MetaDataStr consists of Name and Namespace. Namespace is optional
type MetaDataStr struct {
	Name      string `yaml:"name"`
	Namespace string `yaml:"namespace,omitempty"`
}

/*
GenerateModifiedConfigFile takes in two JSON bytes arrays.
First JSON byte array consists of data which the generated configMap
shall contain or the raw data that the files contain.
Second JSON byte array consists of the patch JSON.
Example of both JSONs
configMap.json
[{
    "scenario": "traffic",
    "address": "Dawson Creek",
    "location": {
        "lat": 45.539626,
        "lon": -122.929569
    },
    "sensors": [{
        "address": "Shute & Dawson Creek",
        "location": {
            "lat": 45.544223,
            "lon": -122.926128
        },
        "algorithm": "object-detection",
        "mnth": 75.0,
        "alpha": 45.0,
        "fovh": 90.0,
        "fovv": 68.0,
        "theta": 0.0,
        "simsn": "cams1o1c0",
        "simfile": "_traffic.mp4$"
    }]
}]

PatchJSON:
[
    {
        "op": "replace",
        "path": "/data/config.json",
        "value":  "1"
    }
]

*/
func GenerateModifiedConfigFile(dataFiles [][]byte, patchData []byte, fns []string, rsName string, rsKind string) ([]byte, error) {
	if len(dataFiles) == 0 || len(patchData) == 0 {
		return []byte{}, pkgerrors.Errorf("Either configData or PatchData empty")
	}

	// Modify the patchFile
	var patchJSON []map[string]interface{}
	json.Unmarshal(patchData, &patchJSON)
	lookupKey := "value"
	for i, patchItem := range patchJSON {
		// replacing the value key in patch file with the data file.
		patchItem[lookupKey] = string((dataFiles[i]))
	}
	modifiedPatchBytes, _ := json.MarshalIndent(patchJSON, "", " ")

	// Decode patch file
	decodedPatch, err := jsonpatch.DecodePatch([]byte(modifiedPatchBytes))
	if err != nil {
		log.Error("Error during decoding Patch file", log.Fields{
			"Error": err.Error(),
		})
		return []byte{}, pkgerrors.Errorf("Internal error")
	}

	var baseResourceBytes []byte
	if strings.ToLower(rsKind) == "configmap" {
		baseResourceBytes, err = SetBaseConfigMap(fns, rsName)
		if err != nil {
			log.Error("Error during SetBaseConfigMap", log.Fields{
				"Error":  err.Error(),
				"cmName": rsName,
			})
			return []byte{}, pkgerrors.Errorf("Internal error")
		}
		log.Info("Set the base configMap", log.Fields{})
	} else if strings.ToLower(rsKind) == "secret" {
		baseResourceBytes, err = SetBaseSecret(fns, rsName)
		if err != nil {
			log.Error("Error during SetBaseSecret", log.Fields{
				"Error":      err.Error(),
				"SecretName": rsName,
			})
			return []byte{}, pkgerrors.Errorf("Internal error")
		}
		log.Info("Set the base Secret", log.Fields{})
	}

	baseResourceJSON, err := yaml.YAMLToJSON(baseResourceBytes)
	if err != nil {
		log.Error("Error in YAML to JSON conversion", log.Fields{})
		return []byte{}, pkgerrors.Errorf("Internal error")
	}

	// Apply the patch
	modified, err := decodedPatch.Apply(baseResourceJSON)
	if err != nil {
		log.Error("Error during applying patch :: GenerateModifiedConfigFile", log.Fields{
			"Error":  err.Error(),
			"cmName": rsName,
		})
		return []byte{}, pkgerrors.Errorf("Internal error")
	}
	log.Info("Successfully patched JSON ...", log.Fields{})
	modifiedYaml, err := yaml.JSONToYAML(modified)
	if err != nil {
		return []byte{}, pkgerrors.Errorf("Internal error")
	}
	return modifiedYaml, nil
}

// SetBaseConfigMap returns a base configMap JSON in form of bytes
func SetBaseConfigMap(fns []string, cmName string) ([]byte, error) {
	cmData := make(map[string]string)
	configJSONStr := `
	{

        "env" : "dev"

    }
	`

	for _, fn := range fns {
		cmData[fn] = configJSONStr
	}

	configMap := ConfigMapResource{
		APIVersion: "v1",
		Kind:       "ConfigMap",
		MetaData: MetaDataStr{
			Name: cmName,
		},
		Data: cmData,
	}

	configMapBytes, err := yamlV2.Marshal(&configMap)
	if err != nil {
		log.Error("error", log.Fields{"error ": err.Error()})
		return []byte{}, err
	}
	return configMapBytes, nil

}

// SetBaseSecret returns a base secret resource in bytes format
func SetBaseSecret(fns []string, cmName string) ([]byte, error) {
	secretData := make(map[string]string)
	secretJSONStr := `
	{

        "SecretKey" : "SecretValue"

    }
	`
	for _, fn := range fns {
		secretData[fn] = secretJSONStr
	}

	secret := SecretResource{
		APIVersion: "v1",
		Kind:       "ConfigMap",
		MetaData: MetaDataStr{
			Name: cmName,
		},
		Type: "kubernetes.io/ssh-auth",
		Data: secretData,
	}

	secretBytes, err := yamlV2.Marshal(&secret)
	if err != nil {
		log.Error("error", log.Fields{"err1 ": err.Error()})
		return []byte{}, err
	}
	return secretBytes, nil

}

// GetPatchFromFile generates patch files. Patch files are array of JSON, eg:
/*
	[
    	{
        	"op": "replace",
        	"path": "/Data/config1.json",
        	"value":  "1"
		},
		{
        	"op": "replace",
        	"path": "/Data/config2.json",
        	"value":  "1"
    	}
	]
	Here "config.json" shall be replaced by the fileNames.
*/
func GetPatchFromFile(fns []string) ([]byte, error) {
	var patch []map[string]interface{}
	// Now, iterate through fns , and fill in the above patch
	for _, fn := range fns {
		eachOp := make(map[string]interface{})
		eachOp["op"] = "replace"
		eachOp["path"] = "/data/" + fn
		eachOp["value"] = "1" // this can be any string value
		patch = append(patch, eachOp)
	}
	return json.Marshal(&patch)
}

// GetPatchFromPatchJSON generates patch files in bytes. This is used when the user requests
/* the patchType as "json" and patchJSON as below:

	"patchType": "json",
    "patchJson": [
      {
        "op": "replace",
        "path": "/spec/replicas",
        "value": "1"
      }
    ]
*/
func GetPatchFromPatchJSON(p []map[string]interface{}) ([]byte, error) {

	modifiedPatchBytes, err := json.MarshalIndent(p, "", " ")
	if err != nil {
		log.Error("Marshal error in GetPatchFromPatchJSON", log.Fields{"Error": err.Error()})
		return nil, err
	}
	return modifiedPatchBytes, err
}

// GenerateModifiedYamlFileForExistingResources takes in the patchData and the existing resource's yaml file and returns the modified yaml file for the resource
func GenerateModifiedYamlFileForExistingResources(patchData []byte, existingResData []byte, resName string) ([]byte, error) {

	// Decode patch file
	decodedPatch, err := jsonpatch.DecodePatch([]byte(patchData))
	if err != nil {
		log.Error("Error during decoding Patch file", log.Fields{
			"Error": err.Error(),
		})
		return []byte{}, pkgerrors.Errorf("Internal error")
	}

	existingResDataJSON, err := yaml.YAMLToJSON(existingResData)
	if err != nil {
		log.Error("Error in YAML to JSON conversion", log.Fields{})
		return []byte{}, pkgerrors.Errorf("Internal error")
	}

	// Apply the patch
	modified, err := decodedPatch.Apply(existingResDataJSON)
	if err != nil {
		log.Error("GenerateModifiedYamlFileForExistingResources::Error during applying patch ", log.Fields{
			"Error":        err.Error(),
			"ResourceName": resName,
		})
		return []byte{}, pkgerrors.Errorf("Internal error")
	}
	log.Info("Successfully patched JSON ...", log.Fields{})
	modifiedYaml, err := yaml.JSONToYAML(modified)
	if err != nil {
		return []byte{}, pkgerrors.Errorf("Internal error")
	}
	return modifiedYaml, nil
}
