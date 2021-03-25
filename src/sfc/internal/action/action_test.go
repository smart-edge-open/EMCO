// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2021 Intel Corporation

package action_test

import (
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/open-ness/EMCO/src/orchestrator/pkg/appcontext"
	"github.com/open-ness/EMCO/src/orchestrator/pkg/infra/contextdb"
	"github.com/open-ness/EMCO/src/orchestrator/pkg/infra/db"
	orch "github.com/open-ness/EMCO/src/orchestrator/pkg/module"
	ovn "github.com/open-ness/EMCO/src/ovnaction/pkg/module"
	cacontext "github.com/open-ness/EMCO/src/rsync/pkg/context"
	catypes "github.com/open-ness/EMCO/src/rsync/pkg/types"
	"github.com/open-ness/EMCO/src/sfc/internal/action"
	"github.com/open-ness/EMCO/src/sfc/pkg/model"
	"github.com/open-ness/EMCO/src/sfc/pkg/module"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// For testing need:
// 1. A Mock AppContext
// 2. with a set of apps which have been place in a set of clusters
//    a) variation 1:  all clusters have all apps that are part of the chain
//    b) variation 2:  no cluster has all apps that are listed in the chain (is this ok?)
//    c) variation 3:  1 cluster does not have all app, another does (treat as ok)
// 3. A Mock Network Control Intent
//    a) variation 1:  No network control that matches the input network control intent
//    b) variation 2:  Network control intent matches the input network control intent
// 4. A Mock SFC Intent
//    a) variation 1:  Zero SFC Intents
//    b) variation 2:  One SFC Intent
//    c) variation 3:  Two SFC Intents

var TestCA1 catypes.CompositeApp = catypes.CompositeApp{
	CompMetadata: appcontext.CompositeAppMeta{
		Project:               "testp",
		CompositeApp:          "chainCA",
		Version:               "v1",
		Release:               "r1",
		DeploymentIntentGroup: "dig1",
		Namespace:             "default",
		Level:                 "0",
	},
	AppOrder: []string{"a1", "a2", "a3"},
	Apps: map[string]*catypes.App{
		"a1": &catypes.App{
			Name: "a1",
			Clusters: map[string]*catypes.Cluster{
				"provider1+cluster1": &catypes.Cluster{
					Name: "provider1+cluster1",
					Resources: map[string]*catypes.AppResource{
						"r1": &catypes.AppResource{Name: "r1", Data: "a1c1r1"},
						"r2": &catypes.AppResource{Name: "r2", Data: "a1c1r2"},
					},
					ResOrder: []string{"r1", "r2"}},
				"provider1+cluster2": &catypes.Cluster{
					Name: "provider1+cluster2",
					Resources: map[string]*catypes.AppResource{
						"r3": &catypes.AppResource{Name: "r3", Data: "a1c2r3"},
						"r4": &catypes.AppResource{Name: "r4", Data: "a1c2r4"},
					},
					ResOrder: []string{"r3", "r4"}}},
		},
		"a2": &catypes.App{
			Name: "a2",
			Clusters: map[string]*catypes.Cluster{
				"provider1+cluster1": &catypes.Cluster{
					Name: "provider1+cluster1",
					Resources: map[string]*catypes.AppResource{
						"r1": &catypes.AppResource{Name: "r3", Data: "a2c1r1"},
						"r2": &catypes.AppResource{Name: "r4", Data: "a2c1r2"},
					},
					ResOrder: []string{"r3", "r4"}},
				"provider1+cluster2": &catypes.Cluster{
					Name: "provider1+cluster2",
					Resources: map[string]*catypes.AppResource{
						"r3": &catypes.AppResource{Name: "r3", Data: "a2c2r3"},
						"r4": &catypes.AppResource{Name: "r4", Data: "a2c2r4"},
						"r5": &catypes.AppResource{Name: "r4", Data: "a2c2r5"},
					},
					ResOrder: []string{"r3", "r4", "r5"}}},
		},
		"a3": &catypes.App{
			Name: "a3",
			Clusters: map[string]*catypes.Cluster{
				"provider1+cluster2": &catypes.Cluster{
					Name: "provider1+cluster2",
					Resources: map[string]*catypes.AppResource{
						"r6": &catypes.AppResource{Name: "r3", Data: "a3c2r6"},
						"r7": &catypes.AppResource{Name: "r4", Data: "a3c2r7"},
					},
					ResOrder: []string{"r3", "r4"}}},
		},
	},
}

var _ = Describe("SFCAction", func() {
	var (
		// Mock AppContext variables
		cdb          *contextdb.MockConDb
		contextIdCA1 string

		// Mock DB variables
		proj       orch.Project
		projClient *orch.ProjectClient

		ca       orch.CompositeApp
		caClient *orch.CompositeAppClient

		dig       orch.DeploymentIntentGroup
		digClient *orch.DeploymentIntentGroupClient

		netCtl       ovn.NetControlIntent
		netCtlClient *ovn.NetControlIntentClient

		sfcIntent                     model.SfcIntent
		sfcIntent2                    model.SfcIntent
		sfcLeftClientSelectorIntent   model.SfcClientSelectorIntent
		sfcRightClientSelectorIntent  model.SfcClientSelectorIntent
		sfcLeftProviderNetworkIntent  model.SfcProviderNetworkIntent
		sfcRightProviderNetworkIntent model.SfcProviderNetworkIntent
		sfcClient                     *module.SfcIntentClient
		sfcClientSelectorClient       *module.SfcClientSelectorIntentClient
		sfcProviderNetworkClient      *module.SfcProviderNetworkIntentClient

		resultingCA catypes.CompositeApp

		mdb *db.NewMockDB
	)

	BeforeEach(func() {
		cdb = new(contextdb.MockConDb)
		cdb.Err = nil
		contextdb.Db = cdb

		// make an AppContext
		cid, _ := cacontext.CreateCompApp(TestCA1)
		con := cacontext.MockConnector{}
		con.Init(cid)
		contextIdCA1 = cid

		// setup the mock DB resources
		// (needs to match the mock AppContext)
		projClient = orch.NewProjectClient()
		proj = orch.Project{
			MetaData: orch.ProjectMetaData{
				Name: "testp",
			},
		}

		caClient = orch.NewCompositeAppClient()
		ca = orch.CompositeApp{
			Metadata: orch.CompositeAppMetaData{
				Name: "chainCA",
			},
			Spec: orch.CompositeAppSpec{
				Version: "v1",
			},
		}

		digClient = orch.NewDeploymentIntentGroupClient()
		dig = orch.DeploymentIntentGroup{
			MetaData: orch.DepMetaData{
				Name: "dig1",
			},
			Spec: orch.DepSpecData{
				Profile:      "profilename",
				Version:      "r1",
				LogicalCloud: "logCloud",
			},
		}

		netCtlClient = ovn.NewNetControlIntentClient()
		netCtl = ovn.NetControlIntent{
			Metadata: ovn.Metadata{
				Name: "netctl",
			},
		}

		sfcClient = module.NewSfcIntentClient()
		sfcIntent = model.SfcIntent{
			Metadata: model.Metadata{
				Name: "sfcIntentName",
			},
			Spec: model.SfcIntentSpec{
				ChainType:    model.RoutingChainType,
				NetworkChain: "net=left-virtual,app=a1,net=dyn1,app=a2,net=right-virtual",
				Namespace:    "chainspace",
			},
		}
		sfcIntent2 = model.SfcIntent{
			Metadata: model.Metadata{
				Name: "sfcIntentName2",
			},
			Spec: model.SfcIntentSpec{
				ChainType:    model.RoutingChainType,
				NetworkChain: "net=left-virtual,app=a3,net=right-virtual",
				Namespace:    "chainspace",
			},
		}

		sfcClientSelectorClient = module.NewSfcClientSelectorIntentClient()
		sfcLeftClientSelectorIntent = model.SfcClientSelectorIntent{
			Metadata: model.Metadata{
				Name: "sfcLeftClientSelectorIntentName",
			},
			Spec: model.SfcClientSelectorIntentSpec{
				ChainEnd: "left",
				PodSelector: metav1.LabelSelector{
					MatchLabels: map[string]string{
						"app": "leftapp",
					},
				},
			},
		}

		sfcRightClientSelectorIntent = model.SfcClientSelectorIntent{
			Metadata: model.Metadata{
				Name: "sfcRightClientSelectorIntentName",
			},
			Spec: model.SfcClientSelectorIntentSpec{
				ChainEnd: "right",
				PodSelector: metav1.LabelSelector{
					MatchLabels: map[string]string{
						"app": "rightapp",
					},
				},
			},
		}

		sfcProviderNetworkClient = module.NewSfcProviderNetworkIntentClient()
		sfcLeftProviderNetworkIntent = model.SfcProviderNetworkIntent{
			Metadata: model.Metadata{
				Name: "sfcLeftProviderNetworkIntentName",
			},
			Spec: model.SfcProviderNetworkIntentSpec{
				ChainEnd:    "left",
				NetworkName: "leftPNet",
				GatewayIp:   "10.10.10.1",
				Subnet:      "10.10.10.0/24",
			},
		}

		sfcRightProviderNetworkIntent = model.SfcProviderNetworkIntent{
			Metadata: model.Metadata{
				Name: "sfcRightProviderNetworkIntentName",
			},
			Spec: model.SfcProviderNetworkIntentSpec{
				ChainEnd:    "right",
				NetworkName: "rightPNet",
				GatewayIp:   "11.11.11.1",
				Subnet:      "11.11.11.0/24",
			},
		}

		mdb = new(db.NewMockDB)
		mdb.Err = nil
		db.DBconn = mdb

		// set up prerequisites
		_, err := (*projClient).CreateProject(proj, false)
		Expect(err).To(BeNil())
		_, err = (*caClient).CreateCompositeApp(ca, "testp", false)
		Expect(err).To(BeNil())
		_, err = (*digClient).CreateDeploymentIntentGroup(dig, "testp", "chainCA", "v1")
		Expect(err).To(BeNil())
		_, err = (*netCtlClient).CreateNetControlIntent(netCtl, "testp", "chainCA", "v1", "dig1", false)
		Expect(err).To(BeNil())
		_, err = (*sfcClient).CreateSfcIntent(sfcIntent, "testp", "chainCA", "v1", "dig1", "netctl", false)
		Expect(err).To(BeNil())
		_, err = (*sfcClientSelectorClient).CreateSfcClientSelectorIntent(sfcLeftClientSelectorIntent, "testp", "chainCA", "v1", "dig1", "netctl", "sfcIntentName", false)
		Expect(err).To(BeNil())
		_, err = (*sfcClientSelectorClient).CreateSfcClientSelectorIntent(sfcRightClientSelectorIntent, "testp", "chainCA", "v1", "dig1", "netctl", "sfcIntentName", false)
		Expect(err).To(BeNil())
		_, err = (*sfcProviderNetworkClient).CreateSfcProviderNetworkIntent(sfcLeftProviderNetworkIntent, "testp", "chainCA", "v1", "dig1", "netctl", "sfcIntentName", false)
		Expect(err).To(BeNil())
		_, err = (*sfcProviderNetworkClient).CreateSfcProviderNetworkIntent(sfcRightProviderNetworkIntent, "testp", "chainCA", "v1", "dig1", "netctl", "sfcIntentName", false)
		Expect(err).To(BeNil())
	})

	It("Missing Both Client Selector intents", func() {
		err := (*sfcClientSelectorClient).DeleteSfcClientSelectorIntent("sfcLeftClientSelectorIntentName", "testp", "chainCA", "v1", "dig1", "netctl", "sfcIntentName")
		Expect(err).To(BeNil())
		err = (*sfcClientSelectorClient).DeleteSfcClientSelectorIntent("sfcRightClientSelectorIntentName", "testp", "chainCA", "v1", "dig1", "netctl", "sfcIntentName")
		Expect(err).To(BeNil())

		err = action.UpdateAppContext("netctl", contextIdCA1)
		Expect(strings.Contains(err.Error(), "Missing left and right client selector intents")).To(Equal(true))
	})

	It("Missing Left Client Selector intent", func() {
		err := (*sfcClientSelectorClient).DeleteSfcClientSelectorIntent("sfcLeftClientSelectorIntentName", "testp", "chainCA", "v1", "dig1", "netctl", "sfcIntentName")
		Expect(err).To(BeNil())

		err = action.UpdateAppContext("netctl", contextIdCA1)
		Expect(strings.Contains(err.Error(), "Missing left client selector intent")).To(Equal(true))
	})

	It("Missing Right Client Selector intent", func() {
		err := (*sfcClientSelectorClient).DeleteSfcClientSelectorIntent("sfcRightClientSelectorIntentName", "testp", "chainCA", "v1", "dig1", "netctl", "sfcIntentName")
		Expect(err).To(BeNil())

		err = action.UpdateAppContext("netctl", contextIdCA1)
		Expect(strings.Contains(err.Error(), "Missing right client selector intent")).To(Equal(true))
	})

	It("Missing Both Provider Network intents", func() {
		err := (*sfcProviderNetworkClient).DeleteSfcProviderNetworkIntent("sfcLeftProviderNetworkIntentName", "testp", "chainCA", "v1", "dig1", "netctl", "sfcIntentName")
		Expect(err).To(BeNil())
		err = (*sfcProviderNetworkClient).DeleteSfcProviderNetworkIntent("sfcRightProviderNetworkIntentName", "testp", "chainCA", "v1", "dig1", "netctl", "sfcIntentName")
		Expect(err).To(BeNil())

		err = action.UpdateAppContext("netctl", contextIdCA1)
		Expect(strings.Contains(err.Error(), "Missing left and right provider network intent")).To(Equal(true))
	})

	It("Missing Left Provider Network intent", func() {
		err := (*sfcProviderNetworkClient).DeleteSfcProviderNetworkIntent("sfcLeftProviderNetworkIntentName", "testp", "chainCA", "v1", "dig1", "netctl", "sfcIntentName")
		Expect(err).To(BeNil())

		err = action.UpdateAppContext("netctl", contextIdCA1)
		Expect(strings.Contains(err.Error(), "Missing left provider network intent")).To(Equal(true))
	})

	It("Missing Right Provider Network intent", func() {
		err := (*sfcProviderNetworkClient).DeleteSfcProviderNetworkIntent("sfcRightProviderNetworkIntentName", "testp", "chainCA", "v1", "dig1", "netctl", "sfcIntentName")
		Expect(err).To(BeNil())

		err = action.UpdateAppContext("netctl", contextIdCA1)
		Expect(strings.Contains(err.Error(), "Missing right provider network intent")).To(Equal(true))
	})

	It("Successful Apply SFC to an App Context", func() {
		err := action.UpdateAppContext("netctl", contextIdCA1)
		Expect(err).To(BeNil())
	})

	It("Net Control Intent does not exist", func() {
		err := action.UpdateAppContext("netctlNot", contextIdCA1)
		Expect(strings.Contains(err.Error(), "Net Control Intent not found")).To(Equal(true))
	})

	It("No SFC intents", func() {
		// delete all the SFC intents
		err := (*sfcProviderNetworkClient).DeleteSfcProviderNetworkIntent("sfcLeftProviderNetworkIntentName", "testp", "chainCA", "v1", "dig1", "netctl", "sfcIntentName")
		Expect(err).To(BeNil())
		err = (*sfcProviderNetworkClient).DeleteSfcProviderNetworkIntent("sfcRightProviderNetworkIntentName", "testp", "chainCA", "v1", "dig1", "netctl", "sfcIntentName")
		Expect(err).To(BeNil())
		err = (*sfcClientSelectorClient).DeleteSfcClientSelectorIntent("sfcLeftClientSelectorIntentName", "testp", "chainCA", "v1", "dig1", "netctl", "sfcIntentName")
		Expect(err).To(BeNil())
		err = (*sfcClientSelectorClient).DeleteSfcClientSelectorIntent("sfcRightClientSelectorIntentName", "testp", "chainCA", "v1", "dig1", "netctl", "sfcIntentName")
		Expect(err).To(BeNil())
		err = (*sfcClient).DeleteSfcIntent("sfcIntentName", "testp", "chainCA", "v1", "dig1", "netctl")
		Expect(err).To(BeNil())

		resultingCA, err = cacontext.ReadAppContext(contextIdCA1)
		cacontext.PrintCompositeApp(resultingCA)

		err = action.UpdateAppContext("netctl", contextIdCA1)
		Expect(strings.Contains(err.Error(), "No SFC Intents are defined for the Network Control Intent")).To(Equal(true))
	})

	It("Successful Apply two SFCs to an App Context", func() {
		// set up second SFC
		_, err := (*sfcClient).CreateSfcIntent(sfcIntent2, "testp", "chainCA", "v1", "dig1", "netctl", false)
		Expect(err).To(BeNil())
		_, err = (*sfcClientSelectorClient).CreateSfcClientSelectorIntent(sfcLeftClientSelectorIntent, "testp", "chainCA", "v1", "dig1", "netctl", "sfcIntentName2", false)
		Expect(err).To(BeNil())
		_, err = (*sfcClientSelectorClient).CreateSfcClientSelectorIntent(sfcRightClientSelectorIntent, "testp", "chainCA", "v1", "dig1", "netctl", "sfcIntentName2", false)
		Expect(err).To(BeNil())
		_, err = (*sfcProviderNetworkClient).CreateSfcProviderNetworkIntent(sfcLeftProviderNetworkIntent, "testp", "chainCA", "v1", "dig1", "netctl", "sfcIntentName2", false)
		Expect(err).To(BeNil())
		_, err = (*sfcProviderNetworkClient).CreateSfcProviderNetworkIntent(sfcRightProviderNetworkIntent, "testp", "chainCA", "v1", "dig1", "netctl", "sfcIntentName2", false)
		Expect(err).To(BeNil())

		resultingCA, err = cacontext.ReadAppContext(contextIdCA1)
		cacontext.PrintCompositeApp(resultingCA)

		err = action.UpdateAppContext("netctl", contextIdCA1)
		Expect(err).To(BeNil())
	})
})
