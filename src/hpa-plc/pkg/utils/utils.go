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

// IsInSlice .. Check whether a string is part of a list
func IsInSlice(s string, list []string) bool {
	for _, i := range list {
		if i == s {
			return true
		}
	}
	return false
}

// GetSliceIntersect .. intersect slices
func GetSliceIntersect(c1 []string, c2 []string) []string {
	var set = make([]string, 0)
	for _, v1 := range c1 {
		for _, v2 := range c2 {
			if v1 == v2 {
				set = append(set, v1)
			}
		}
	}
	return set
}

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
