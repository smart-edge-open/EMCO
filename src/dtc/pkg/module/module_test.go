// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2020 Intel Corporation

package module_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/open-ness/EMCO/src/dtc/pkg/module"
)

var _ = Describe("Trafficgroupintent", func() {

	var (
		client *module.Client
	)

	BeforeEach(func() {
		client = &module.Client{}
		client.TrafficGroupIntent = module.NewTrafficGroupIntentClient()
		client.ServerInboundIntent = module.NewServerInboundIntentClient()
		client.ClientsInboundIntent = module.NewClientsInboundIntentClient()
	})

	Describe("Create new client", func() {
		It("should return client", func() {
			c := module.NewClient()
			Expect(c).Should(Equal(client))
		})
	})
})
