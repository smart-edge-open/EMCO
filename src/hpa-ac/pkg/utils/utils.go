// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2020 Intel Corporation

package utils

import (
	"io/ioutil"
	"os"

	pkgerrors "github.com/pkg/errors"
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
