package module_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/open-ness/EMCO/src/ovnaction/pkg/module"
	"github.com/open-ness/EMCO/src/orchestrator/pkg/infra/db"
	pkgerrors "github.com/pkg/errors"
)

var _ = Describe("Chaining", func() {

	var (
		NCI module.NetControlIntent
		NCIDBC *module.NetControlIntentClient

		CHAIN module.Chain
		CHAINDBC *module.ChainClient

		mdb *db.MockDB
		)

		BeforeEach(func() {
			NCIDBC = module.NewNetControlIntentClient()
			NCI = module.NetControlIntent{
			Metadata: module.Metadata {
			Name: "theName",
			Description: "net control intent",
			UserData1: "user data1",
			UserData2: "user data2",
				},
			}

			CHAINDBC = module.NewChainClient()
			CHAIN = module.Chain{
	  		Metadata: module.Metadata {
				Name: "theFourthName",
				Description: "chain",
				UserData1: "user data1",
				UserData2: "user data2",
			},
		}

		mdb = new(db.MockDB)
		mdb.Err = nil
		db.DBconn = mdb
	})

	Describe("Create client intent", func() {
		It("with pre created net control intent should return nil", func() {
			_,err := (*NCIDBC).CreateNetControlIntent(NCI, "test", "capp1", "v1", "dig", false)
			Expect(err).To(BeNil())
			_,err = (*CHAINDBC).CreateChain(CHAIN, "test", "capp1", "v1", "dig", "theName", false)
			Expect(err).To(BeNil())
		})
		It("followed by create again should return error", func() {
			_,err := (*NCIDBC).CreateNetControlIntent(NCI, "test", "capp1", "v1", "dig", false)
			Expect(err).To(BeNil())
			_,err = (*CHAINDBC).CreateChain(CHAIN, "test", "capp1", "v1", "dig", "theName", false)
			Expect(err).To(BeNil())
			_,err = (*CHAINDBC).CreateChain(CHAIN, "test", "capp1", "v1", "dig", "theName", false)
			Expect(err).To(HaveOccurred())
		})
		It("followed by get should return nil", func() {
			_,err := (*NCIDBC).CreateNetControlIntent(NCI, "test", "capp1", "v1", "dig", false)
			Expect(err).To(BeNil())
			_,err = (*CHAINDBC).CreateChain(CHAIN, "test", "capp1", "v1", "dig", "theName", false)
			Expect(err).To(BeNil())
			chain,err := (*CHAINDBC).GetChain("theFourthName", "test", "capp1", "v1", "dig", "theName")
			Expect(chain).Should(Equal(CHAIN))
		})
		It("followed by delete should return nil", func() {
			_,err := (*NCIDBC).CreateNetControlIntent(NCI, "test", "capp1", "v1", "dig", false)
			Expect(err).To(BeNil())
			_,err = (*CHAINDBC).CreateChain(CHAIN, "test", "capp1", "v1", "dig", "theName", false)
			Expect(err).To(BeNil())
			err = (*CHAINDBC).DeleteChain("testnci", "test", "capp1", "v1", "dig", "theName")
			Expect(err).To(BeNil())
		})
	})

	Describe("Create net intent", func() {
		It("followed by create,get,delete,get chain should return an error", func() {
			_,err := (*NCIDBC).CreateNetControlIntent(NCI, "test", "capp1", "v1", "dig", false)
			Expect(err).To(BeNil())
			_,err = (*CHAINDBC).CreateChain(CHAIN, "test", "capp1", "v1", "dig", "theName", false)
			Expect(err).To(BeNil())
			chain,err := (*CHAINDBC).GetChain("theFourthName", "test", "capp1", "v1", "dig", "theName")
			Expect(chain).Should(Equal(CHAIN))
			err = (*CHAINDBC).DeleteChain("theFourthName", "test", "capp1", "v1", "dig", "theName")
			Expect(err).To(BeNil())
			chain,err = (*CHAINDBC).GetChain("theFourthName", "test", "capp1", "v1", "dig", "theName")
			Expect(err).To(HaveOccurred())
		})
	})

	Describe("Get chains", func() {
		It("should return error for non-existing record", func() {
			mdb.Err = pkgerrors.New("Error finding:")
			_,err := (*CHAINDBC).GetChains("test", "capp1", "v1", "dig", "theName")
			Expect(err).To(HaveOccurred())
		})
	})
})
