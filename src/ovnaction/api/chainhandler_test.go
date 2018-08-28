// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2020 Intel Corporation

package api

import (
	"testing"
)

func TestIsValidNetworkChain(t *testing.T) {
	t.Run("Valid Chains", func(t *testing.T) {
		validchains := []string{
			"app=abc,net1,app=xyz",
			"app=abc, net1, app=xyz",
			" app=abc , net1 , app=xyz ",
			"app.kubernets.io/name=abc,net1,app.kubernets.io/name=xyz",
			"app.kubernets.io/name=abc,net1,app.kubernets.io/name=xyz, net2, anotherlabel=wex",
		}
		for _, chain := range validchains {
			err := validateNetworkChain(chain)
			if err != nil {
				t.Errorf("Valid network chain failed to pass: %v %v", chain, err)
			}
		}
	})

	t.Run("Invalid Chains", func(t *testing.T) {
		invalidchains := []string{
			"",
			"app=abc,net1,app= xyz",
			"app=abc,net1,xyz",
			"app=abc,net1",
			"app.kubernets.io/name=abc,net1,=xyz",
			"abcdefg",
		}
		for _, chain := range invalidchains {
			err := validateNetworkChain(chain)
			if err == nil {
				t.Errorf("Invalid network chain passed: %v", chain)
			}
		}
	})
}
