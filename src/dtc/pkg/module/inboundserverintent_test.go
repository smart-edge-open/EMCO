// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2020 Intel Corporation
package module_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/open-ness/EMCO/src/dtc/pkg/module"
	"github.com/open-ness/EMCO/src/orchestrator/pkg/infra/db"
	pkgerrors "github.com/pkg/errors"
)

var _ = Describe("Inboundserverintent", func() {

	var (
		TGI    module.TrafficGroupIntent
		TGIDBC *module.TrafficGroupIntentDbClient

		ISI    module.InboundServerIntent
		ISIDBC *module.InboundServerIntentDbClient

		mdb *db.MockDB
	)

	BeforeEach(func() {

		TGIDBC = module.NewTrafficGroupIntentClient()
		TGI = module.TrafficGroupIntent{
			Metadata: module.Metadata{
				Name:        "testtgi",
				Description: "traffic group intent",
				UserData1:   "user data1",
				UserData2:   "user data2",
			},
		}

		ISIDBC = module.NewServerInboundIntentClient()
		ISI = module.InboundServerIntent{
			Metadata: module.Metadata{
				Name:        "testisi",
				Description: "inbound server intent",
				UserData1:   "user data1",
				UserData2:   "user data2",
			},
		}
		mdb = new(db.MockDB)
		mdb.Err = nil
		db.DBconn = mdb

	})

	Describe("Create server intent", func() {
		It("with pre created traffic intent should return nil", func() {
			_, err := (*TGIDBC).CreateTrafficGroupIntent(TGI, "test", "capp1", "v1", "dig", false)
			Expect(err).To(BeNil())
			_, err = (*ISIDBC).CreateServerInboundIntent(ISI, "test", "capp1", "v1", "dig", "testtgi", false)
			Expect(err).To(BeNil())
		})
		It("should return error", func() {
			_, err := (*ISIDBC).CreateServerInboundIntent(ISI, "test", "capp1", "v1", "dig", "testtgi", false)
			Expect(err).To(HaveOccurred())
		})
		It("create again should return error", func() {
			_, err := (*TGIDBC).CreateTrafficGroupIntent(TGI, "test", "capp1", "v1", "dig", false)
			Expect(err).To(BeNil())
			_, err = (*ISIDBC).CreateServerInboundIntent(ISI, "test", "capp1", "v1", "dig", "testtgi", false)
			Expect(err).To(BeNil())
			_, err = (*ISIDBC).CreateServerInboundIntent(ISI, "test", "capp1", "v1", "dig", "testtgi", false)
			Expect(err).To(HaveOccurred())
		})
		It("followed by get server intent should return nil", func() {
			_, err := (*TGIDBC).CreateTrafficGroupIntent(TGI, "test", "capp1", "v1", "dig", false)
			Expect(err).To(BeNil())
			_, err = (*ISIDBC).CreateServerInboundIntent(ISI, "test", "capp1", "v1", "dig", "testtgi", false)
			Expect(err).To(BeNil())
			isi, err := (*ISIDBC).GetServerInboundIntent("testisi", "test", "capp1", "v1", "dig", "testtgi")
			Expect(err).To(BeNil())
			Expect(isi).Should(Equal(ISI))
		})
		It("followed by delete server intent should return nil", func() {
			_, err := (*TGIDBC).CreateTrafficGroupIntent(TGI, "test", "capp1", "v1", "dig", false)
			Expect(err).To(BeNil())
			_, err = (*ISIDBC).CreateServerInboundIntent(ISI, "test", "capp1", "v1", "dig", "testtgi", false)
			Expect(err).To(BeNil())
			err = (*ISIDBC).DeleteServerInboundIntent("testisi", "test", "capp1", "v1", "dig", "testtgi")
			Expect(err).To(BeNil())
		})

	})

	Describe("Get server intent", func() {
		It("should return error for non-existing record", func() {
			_, err := (*ISIDBC).GetServerInboundIntent("testisi", "test", "capp1", "v1", "dig", "testtgi")
			Expect(err).To(HaveOccurred())
		})

	})
	Describe("Get server intents", func() {
		It("should return error for non-existing record", func() {
			mdb.Err = pkgerrors.New("Error finding:")
			_, err := (*ISIDBC).GetServerInboundIntents("test", "capp1", "v1", "dig", "testtgi")
			Expect(err).To(HaveOccurred())
		})

	})
	Describe("Delete server intent", func() {
		It("should return error for non-existing record", func() {
			mdb.Err = pkgerrors.New("Error finding:")
			err := (*ISIDBC).DeleteServerInboundIntent("testisi", "test", "capp1", "v1", "dig", "testtgi")
			Expect(err).To(HaveOccurred())
		})
		It("should return error for deleting parent without deleting child", func() {
			mdb.Err = pkgerrors.New("Can't delete parent without deleting child")
			err := (*ISIDBC).DeleteServerInboundIntent("testisi", "test", "capp1", "v1", "dig", "testtgi")
			Expect(err).To(HaveOccurred())
		})
		It("should return error for general db error", func() {
			mdb.Err = pkgerrors.New("general error")
			err := (*ISIDBC).DeleteServerInboundIntent("testisi", "test", "capp1", "v1", "dig", "testtgi")
			Expect(err).To(HaveOccurred())
		})

	})
})
