// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2020 Intel Corporation

package utils

import (
	"io/ioutil"
	"os"
	"path"

	log "github.com/open-ness/EMCO/src/orchestrator/pkg/infra/logutils"
	pkgerrors "github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/scheme"
)

// DecodeYAMLFile reads a YAMl file to extract the Kubernetes object definition
func DecodeYAMLFile(path string, into runtime.Object) (runtime.Object, error) {
	if _, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			return nil, pkgerrors.New("File " + path + " not found")
		} else {
			return nil, pkgerrors.Wrap(err, "Stat file error")
		}
	}

	rawBytes, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, pkgerrors.Wrap(err, "Read YAML file error")
	}

	decode := scheme.Codecs.UniversalDeserializer().Decode
	obj, _, err := decode(rawBytes, nil, into)
	if err != nil {
		return nil, pkgerrors.Wrap(err, "Deserialize YAML error")
	}

	return obj, nil
}

// DecodeYAMLData reads a string to extract the Kubernetes object definition
func DecodeYAMLData(data string, into runtime.Object) (runtime.Object, error) {
	decode := scheme.Codecs.UniversalDeserializer().Decode
	obj, _, err := decode([]byte(data), nil, into)
	if err != nil {
		return nil, pkgerrors.Wrap(err, "Deserialize YAML error")
	}

	return obj, nil
}

//EnsureDirectory makes sure that the directories specified in the path exist
//If not, it will create them, if possible.
func EnsureDirectory(f string) error {
	base := path.Dir(f)
	_, err := os.Stat(base)
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	return os.MkdirAll(base, 0700)
}

// TagPodsIfPresent finds the TemplateSpec from any workload
// object that contains it and changes the spec to include the tag label
func TagPodsIfPresent(unstruct *unstructured.Unstructured, tag string) {
	_, found, err := unstructured.NestedMap(unstruct.Object, "spec", "template")
	if err != nil || !found {
		return
	}
	// extract spec template labels
	labels, found, err := unstructured.NestedMap(unstruct.Object, "spec", "template", "metadata", "labels")
	if err != nil {
		log.Error("TagPodsIfPresent: Error reading the NestMap for template", log.Fields{"unstruct": unstruct, "err": err})
		return
	}
	if labels == nil || !found {
		labels = make(map[string]interface{})
	}
	labels["emco/deployment-id"] = tag
	if err := unstructured.SetNestedMap(unstruct.Object, labels, "spec", "template", "metadata", "labels"); err != nil {
		log.Error("Error tagging template with emco label", log.Fields{"err": err})
	}
}
