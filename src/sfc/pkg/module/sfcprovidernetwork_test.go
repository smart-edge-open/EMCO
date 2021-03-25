// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2021 Intel Corporation

package module_test

import (
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	pkgerrors "github.com/pkg/errors"

	"github.com/open-ness/EMCO/src/orchestrator/pkg/infra/db"
	orch "github.com/open-ness/EMCO/src/orchestrator/pkg/module"
	ovn "github.com/open-ness/EMCO/src/ovnaction/pkg/module"
	"github.com/open-ness/EMCO/src/sfc/pkg/model"
	"github.com/open-ness/EMCO/src/sfc/pkg/module"
)

var _ = Describe("SFCProviderNetworkIntent", func() {
	var (
		proj       orch.Project
		projClient *orch.ProjectClient

		ca       orch.CompositeApp
		caClient *orch.CompositeAppClient

		dig       orch.DeploymentIntentGroup
		digClient *orch.DeploymentIntentGroupClient

		netCtl       ovn.NetControlIntent
		netCtlClient *ovn.NetControlIntentClient

		sfcIntent                model.SfcIntent
		sfcIntentClient          *module.SfcIntentClient
		sfcProviderNetworkIntent model.SfcProviderNetworkIntent
		sfcClient                *module.SfcProviderNetworkIntentClient

		mdb *db.NewMockDB
	)

	BeforeEach(func() {
		projClient = orch.NewProjectClient()
		proj = orch.Project{
			MetaData: orch.ProjectMetaData{
				Name: "testproject",
			},
		}

		caClient = orch.NewCompositeAppClient()
		ca = orch.CompositeApp{
			Metadata: orch.CompositeAppMetaData{
				Name: "ca",
			},
			Spec: orch.CompositeAppSpec{
				Version: "v1",
			},
		}

		digClient = orch.NewDeploymentIntentGroupClient()
		dig = orch.DeploymentIntentGroup{
			MetaData: orch.DepMetaData{
				Name: "dig",
			},
			Spec: orch.DepSpecData{
				Profile:      "profilename",
				Version:      "testver",
				LogicalCloud: "logCloud",
			},
		}

		netCtlClient = ovn.NewNetControlIntentClient()
		netCtl = ovn.NetControlIntent{
			Metadata: ovn.Metadata{
				Name: "netctl",
			},
		}

		sfcIntentClient = module.NewSfcIntentClient()
		sfcIntent = model.SfcIntent{
			Metadata: model.Metadata{
				Name: "sfcIntentName",
			},
		}

		sfcClient = module.NewSfcProviderNetworkIntentClient()
		sfcProviderNetworkIntent = model.SfcProviderNetworkIntent{
			Metadata: model.Metadata{
				Name: "sfcProviderNetworkIntentName",
			},
		}

		mdb = new(db.NewMockDB)
		mdb.Err = nil
		db.DBconn = mdb
	})

	Describe("Create SFC provider network intent", func() {
		It("successful creation of sfc provider network intent", func() {
			// set up prerequisites
			_, err := (*projClient).CreateProject(proj, false)
			Expect(err).To(BeNil())
			_, err = (*caClient).CreateCompositeApp(ca, "testproject", false)
			Expect(err).To(BeNil())
			_, err = (*digClient).CreateDeploymentIntentGroup(dig, "testproject", "ca", "v1")
			Expect(err).To(BeNil())
			_, err = (*netCtlClient).CreateNetControlIntent(netCtl, "testproject", "ca", "v1", "dig", false)
			Expect(err).To(BeNil())
			_, err = (*sfcIntentClient).CreateSfcIntent(sfcIntent, "testproject", "ca", "v1", "dig", "netctl", false)
			Expect(err).To(BeNil())

			// test SFC provider network intent creation
			_, err = (*sfcClient).CreateSfcProviderNetworkIntent(sfcProviderNetworkIntent, "testproject", "ca", "v1", "dig", "netctl", "sfcIntentName", false)
			Expect(err).To(BeNil())
		})
		It("followed by create again should return error", func() {
			// set up prerequisites
			_, err := (*projClient).CreateProject(proj, false)
			Expect(err).To(BeNil())
			_, err = (*caClient).CreateCompositeApp(ca, "testproject", false)
			Expect(err).To(BeNil())
			_, err = (*digClient).CreateDeploymentIntentGroup(dig, "testproject", "ca", "v1")
			Expect(err).To(BeNil())
			_, err = (*netCtlClient).CreateNetControlIntent(netCtl, "testproject", "ca", "v1", "dig", false)
			Expect(err).To(BeNil())
			_, err = (*sfcIntentClient).CreateSfcIntent(sfcIntent, "testproject", "ca", "v1", "dig", "netctl", false)
			Expect(err).To(BeNil())

			// test SFC intent creation
			_, err = (*sfcClient).CreateSfcProviderNetworkIntent(sfcProviderNetworkIntent, "testproject", "ca", "v1", "dig", "netctl", "sfcIntentName", false)
			Expect(err).To(BeNil())
			// test SFC intent creation
			_, err = (*sfcClient).CreateSfcProviderNetworkIntent(sfcProviderNetworkIntent, "testproject", "ca", "v1", "dig", "netctl", "sfcIntentName", false)
			Expect(strings.Contains(err.Error(), "SFC Provider Network Intent already exists")).To(Equal(true))
		})
		It("Parent resource does not exist", func() {
			// set up prerequisites
			_, err := (*projClient).CreateProject(proj, false)
			Expect(err).To(BeNil())
			_, err = (*caClient).CreateCompositeApp(ca, "testproject", false)
			Expect(err).To(BeNil())
			_, err = (*digClient).CreateDeploymentIntentGroup(dig, "testproject", "ca", "v1")
			Expect(err).To(BeNil())
			_, err = (*netCtlClient).CreateNetControlIntent(netCtl, "testproject", "ca", "v1", "dig", false)
			Expect(err).To(BeNil())
			_, err = (*sfcIntentClient).CreateSfcIntent(sfcIntent, "testproject", "ca", "v1", "dig", "netctl", false)
			Expect(err).To(BeNil())

			// test SFC provider network intent creation, but give SFC Intent  name that does not exist
			_, err = (*sfcClient).CreateSfcProviderNetworkIntent(sfcProviderNetworkIntent, "testproject", "ca", "v1", "dig", "netctl", "sfcIntentNameXYZ", false)
			Expect(strings.Contains(err.Error(), "Parent SFC Intent does not exist")).To(Equal(true))
		})
		It("successful creation of sfc provider network intent with update version of call", func() {
			// set up prerequisites
			_, err := (*projClient).CreateProject(proj, false)
			Expect(err).To(BeNil())
			_, err = (*caClient).CreateCompositeApp(ca, "testproject", false)
			Expect(err).To(BeNil())
			_, err = (*digClient).CreateDeploymentIntentGroup(dig, "testproject", "ca", "v1")
			Expect(err).To(BeNil())
			_, err = (*netCtlClient).CreateNetControlIntent(netCtl, "testproject", "ca", "v1", "dig", false)
			Expect(err).To(BeNil())
			_, err = (*sfcIntentClient).CreateSfcIntent(sfcIntent, "testproject", "ca", "v1", "dig", "netctl", false)
			Expect(err).To(BeNil())

			// test SFC provider network intent creation, with update form of call (exists bool == true)
			_, err = (*sfcClient).CreateSfcProviderNetworkIntent(sfcProviderNetworkIntent, "testproject", "ca", "v1", "dig", "netctl", "sfcIntentName", true)
			Expect(err).To(BeNil())
		})
		It("successful creation of sfc provider network intent with update version of call", func() {
			// set up prerequisites
			_, err := (*projClient).CreateProject(proj, false)
			Expect(err).To(BeNil())
			_, err = (*caClient).CreateCompositeApp(ca, "testproject", false)
			Expect(err).To(BeNil())
			_, err = (*digClient).CreateDeploymentIntentGroup(dig, "testproject", "ca", "v1")
			Expect(err).To(BeNil())
			_, err = (*netCtlClient).CreateNetControlIntent(netCtl, "testproject", "ca", "v1", "dig", false)
			Expect(err).To(BeNil())
			_, err = (*sfcIntentClient).CreateSfcIntent(sfcIntent, "testproject", "ca", "v1", "dig", "netctl", false)
			Expect(err).To(BeNil())

			// test SFC provider network intent creation
			_, err = (*sfcClient).CreateSfcProviderNetworkIntent(sfcProviderNetworkIntent, "testproject", "ca", "v1", "dig", "netctl", "sfcIntentName", false)
			Expect(err).To(BeNil())
			_, err = (*sfcClient).CreateSfcProviderNetworkIntent(sfcProviderNetworkIntent, "testproject", "ca", "v1", "dig", "netctl", "sfcIntentName", true)
			Expect(err).To(BeNil())
		})
	})

	Describe("Get all sfc provider network intents", func() {
		It("Parent SFC Intent does not exist - return not found error", func() {
			_, err := (*sfcClient).GetAllSfcProviderNetworkIntents("testproject", "ca", "v1", "dig", "netctl", "sfcIntentName")
			Expect(strings.Contains(err.Error(), "not found")).To(Equal(true))
		})
		It("Parent SFC Intent does exist - No SFC provider network Intents - should return empty list", func() {
			// set up prerequisites
			_, err := (*projClient).CreateProject(proj, false)
			Expect(err).To(BeNil())
			_, err = (*caClient).CreateCompositeApp(ca, "testproject", false)
			Expect(err).To(BeNil())
			_, err = (*digClient).CreateDeploymentIntentGroup(dig, "testproject", "ca", "v1")
			Expect(err).To(BeNil())
			_, err = (*netCtlClient).CreateNetControlIntent(netCtl, "testproject", "ca", "v1", "dig", false)
			Expect(err).To(BeNil())
			_, err = (*sfcIntentClient).CreateSfcIntent(sfcIntent, "testproject", "ca", "v1", "dig", "netctl", false)
			Expect(err).To(BeNil())

			list, err := (*sfcClient).GetAllSfcProviderNetworkIntents("testproject", "ca", "v1", "dig", "netctl", "sfcIntentName")
			Expect(len(list)).To(Equal(0))
		})
		It("Parent SFC Intent does exist - 2 SFC Intents created - should return list of len 2", func() {
			// set up prerequisites
			_, err := (*projClient).CreateProject(proj, false)
			Expect(err).To(BeNil())
			_, err = (*caClient).CreateCompositeApp(ca, "testproject", false)
			Expect(err).To(BeNil())
			_, err = (*digClient).CreateDeploymentIntentGroup(dig, "testproject", "ca", "v1")
			Expect(err).To(BeNil())
			_, err = (*netCtlClient).CreateNetControlIntent(netCtl, "testproject", "ca", "v1", "dig", false)
			Expect(err).To(BeNil())
			_, err = (*sfcIntentClient).CreateSfcIntent(sfcIntent, "testproject", "ca", "v1", "dig", "netctl", false)
			Expect(err).To(BeNil())

			// SFC provider network intent creation - make 2 of them
			_, err = (*sfcClient).CreateSfcProviderNetworkIntent(sfcProviderNetworkIntent, "testproject", "ca", "v1", "dig", "netctl", "sfcIntentName", true)
			Expect(err).To(BeNil())
			sfcProviderNetworkIntent.Metadata.Name = "2nd_name"
			_, err = (*sfcClient).CreateSfcProviderNetworkIntent(sfcProviderNetworkIntent, "testproject", "ca", "v1", "dig", "netctl", "sfcIntentName", true)
			Expect(err).To(BeNil())

			list, err := (*sfcClient).GetAllSfcProviderNetworkIntents("testproject", "ca", "v1", "dig", "netctl", "sfcIntentName")
			Expect(len(list)).To(Equal(2))

		})
		It("should return error for general db error", func() {
			mdb.Err = pkgerrors.New("db Find error")
			_, err := (*sfcClient).GetAllSfcProviderNetworkIntents("testproject", "ca", "v1", "dig", "netctl", "sfcIntentName")
			Expect(strings.Contains(err.Error(), "db Find error")).To(Equal(true))
		})
		It("should return error for unmarshalling db error", func() {
			// set up prerequisites
			_, err := (*projClient).CreateProject(proj, false)
			Expect(err).To(BeNil())
			_, err = (*caClient).CreateCompositeApp(ca, "testproject", false)
			Expect(err).To(BeNil())
			_, err = (*digClient).CreateDeploymentIntentGroup(dig, "testproject", "ca", "v1")
			Expect(err).To(BeNil())
			_, err = (*netCtlClient).CreateNetControlIntent(netCtl, "testproject", "ca", "v1", "dig", false)
			Expect(err).To(BeNil())
			_, err = (*sfcIntentClient).CreateSfcIntent(sfcIntent, "testproject", "ca", "v1", "dig", "netctl", false)
			Expect(err).To(BeNil())

			mdb.MarshalErr = pkgerrors.New("Unmarshaling bson")
			_, err = (*sfcClient).GetAllSfcProviderNetworkIntents("testproject", "ca", "v1", "dig", "netctl", "sfcIntentName")
			Expect(strings.Contains(err.Error(), "Unmarshalling Value")).To(Equal(true))
		})
	})

	Describe("Get sfc provider network intent", func() {
		It("Parent SFC Intent does not exist - return not found error", func() {
			_, err := (*sfcClient).GetSfcProviderNetworkIntent("sfcIntentName", "testproject", "ca", "v1", "dig", "netctl", "sfcIntentName")
			Expect(strings.Contains(err.Error(), "not found")).To(Equal(true))
		})
		It("Successful get of sfcProviderNetworkIntent", func() {
			// set up prerequisites
			_, err := (*projClient).CreateProject(proj, false)
			Expect(err).To(BeNil())
			_, err = (*caClient).CreateCompositeApp(ca, "testproject", false)
			Expect(err).To(BeNil())
			_, err = (*digClient).CreateDeploymentIntentGroup(dig, "testproject", "ca", "v1")
			Expect(err).To(BeNil())
			_, err = (*netCtlClient).CreateNetControlIntent(netCtl, "testproject", "ca", "v1", "dig", false)
			Expect(err).To(BeNil())
			_, err = (*sfcIntentClient).CreateSfcIntent(sfcIntent, "testproject", "ca", "v1", "dig", "netctl", false)
			Expect(err).To(BeNil())

			// test SFC intent creation
			_, err = (*sfcClient).CreateSfcProviderNetworkIntent(sfcProviderNetworkIntent, "testproject", "ca", "v1", "dig", "netctl", "sfcIntentName", false)
			Expect(err).To(BeNil())

			_, err = (*sfcClient).GetSfcProviderNetworkIntent("sfcProviderNetworkIntentName", "testproject", "ca", "v1", "dig", "netctl", "sfcIntentName")
			Expect(err).To(BeNil())
		})
		It("should return error for general db error", func() {
			mdb.Err = pkgerrors.New("db Find error")
			_, err := (*sfcClient).GetSfcProviderNetworkIntent("sfcProviderNetworkIntentName", "testproject", "ca", "v1", "dig", "netctl", "sfcIntentName")
			Expect(strings.Contains(err.Error(), "db Find error")).To(Equal(true))
		})
		It("should return error for unmarshalling db error", func() {
			// set up prerequisites
			_, err := (*projClient).CreateProject(proj, false)
			Expect(err).To(BeNil())
			_, err = (*caClient).CreateCompositeApp(ca, "testproject", false)
			Expect(err).To(BeNil())
			_, err = (*digClient).CreateDeploymentIntentGroup(dig, "testproject", "ca", "v1")
			Expect(err).To(BeNil())
			_, err = (*netCtlClient).CreateNetControlIntent(netCtl, "testproject", "ca", "v1", "dig", false)
			Expect(err).To(BeNil())
			_, err = (*sfcIntentClient).CreateSfcIntent(sfcIntent, "testproject", "ca", "v1", "dig", "netctl", false)
			Expect(err).To(BeNil())

			_, err = (*sfcClient).CreateSfcProviderNetworkIntent(sfcProviderNetworkIntent, "testproject", "ca", "v1", "dig", "netctl", "sfcIntentName", false)
			Expect(err).To(BeNil())
			mdb.MarshalErr = pkgerrors.New("Unmarshaling bson")
			_, err = (*sfcClient).GetSfcProviderNetworkIntent("sfcProviderNetworkIntentName", "testproject", "ca", "v1", "dig", "netctl", "sfcIntentName")
			Expect(strings.Contains(err.Error(), "Unmarshalling Value")).To(Equal(true))
		})
	})

	Describe("Delete SFC provider network intent", func() {
		It("successful delete", func() {
			// set up prerequisites
			_, err := (*projClient).CreateProject(proj, false)
			Expect(err).To(BeNil())
			_, err = (*caClient).CreateCompositeApp(ca, "testproject", false)
			Expect(err).To(BeNil())
			_, err = (*digClient).CreateDeploymentIntentGroup(dig, "testproject", "ca", "v1")
			Expect(err).To(BeNil())
			_, err = (*netCtlClient).CreateNetControlIntent(netCtl, "testproject", "ca", "v1", "dig", false)
			Expect(err).To(BeNil())
			_, err = (*sfcIntentClient).CreateSfcIntent(sfcIntent, "testproject", "ca", "v1", "dig", "netctl", false)
			Expect(err).To(BeNil())

			// test SFC intent creation
			_, err = (*sfcClient).CreateSfcProviderNetworkIntent(sfcProviderNetworkIntent, "testproject", "ca", "v1", "dig", "netctl", "sfcIntentName", false)
			Expect(err).To(BeNil())

			err = (*sfcClient).DeleteSfcProviderNetworkIntent("sfcProviderNetworkIntentName", "testproject", "ca", "v1", "dig", "netctl", "sfcIntentName")
			Expect(err).To(BeNil())
		})
		It("should return not found error for non-existing record", func() {
			mdb.Err = pkgerrors.New("Error finding:")
			err := (*sfcClient).DeleteSfcProviderNetworkIntent("sfcProviderNetworkIntentName", "testproject", "ca", "v1", "dig", "netctl", "sfcIntentName")
			Expect(strings.Contains(err.Error(), "db Remove error - not found")).To(Equal(true))
		})
		It("should return error for deleting parent without deleting child", func() {
			mdb.Err = pkgerrors.New("Can't delete parent without deleting child references first")
			err := (*sfcClient).DeleteSfcProviderNetworkIntent("sfcProviderNetworkIntentName", "testproject", "ca", "v1", "dig", "netctl", "sfcIntentName")
			Expect(strings.Contains(err.Error(), "db Remove error - conflict")).To(Equal(true))
		})
		It("should return error for general db error", func() {
			mdb.Err = pkgerrors.New("any other error")
			err := (*sfcClient).DeleteSfcProviderNetworkIntent("sfcProviderNetworkIntentName", "testproject", "ca", "v1", "dig", "netctl", "sfcIntentName")
			Expect(strings.Contains(err.Error(), "db Remove error - general")).To(Equal(true))
		})
	})
})
