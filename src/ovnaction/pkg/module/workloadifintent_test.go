package module_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/open-ness/EMCO/src/orchestrator/pkg/infra/db"
	"github.com/open-ness/EMCO/src/ovnaction/pkg/module"
	pkgerrors "github.com/pkg/errors"
)

var _ = Describe("Workloadifintent", func() {
	var (
		NCI    module.NetControlIntent
		NCIDBC *module.NetControlIntentClient

		WLI    module.WorkloadIntent
		WLIDBC *module.WorkloadIntentClient

		WLFI    module.WorkloadIfIntent
		WLFIDBC *module.WorkloadIfIntentClient

		mdb *db.MockDB
	)

	BeforeEach(func() {
		NCIDBC = module.NewNetControlIntentClient()
		NCI = module.NetControlIntent{
			Metadata: module.Metadata{
				Name:        "theName",
				Description: "net control intent",
				UserData1:   "user data1",
				UserData2:   "user data2",
			},
		}

		WLIDBC = module.NewWorkloadIntentClient()
		WLI = module.WorkloadIntent{
			Metadata: module.Metadata{
				Name:        "theSecondName",
				Description: "work load intent",
				UserData1:   "user data1",
				UserData2:   "user data2",
			},
		}

		WLFIDBC = module.NewWorkloadIfIntentClient()
		WLFI = module.WorkloadIfIntent{
			Metadata: module.Metadata{
				Name:        "theThirdName",
				Description: "work load if intent",
				UserData1:   "user data1",
				UserData2:   "user data2",
			},
		}

		mdb = new(db.MockDB)
		mdb.Err = nil
		db.DBconn = mdb
	})

	Describe("Create client intent", func() {
		It("with pre created net control intent should return nil", func() {
			_, err := (*NCIDBC).CreateNetControlIntent(NCI, "test", "capp1", "v1", "dig", false)
			Expect(err).To(BeNil())
			_, err = (*WLIDBC).CreateWorkloadIntent(WLI, "test", "capp1", "v1", "dig", "theName", false)
			Expect(err).To(BeNil())
			_, err = (*WLFIDBC).CreateWorkloadIfIntent(WLFI, "test", "capp1", "v1", "dig", "theName", "theSecondName", false)
			Expect(err).To(BeNil())
		})
		It("followed by create again should return error", func() {
			_, err := (*NCIDBC).CreateNetControlIntent(NCI, "test", "capp1", "v1", "dig", false)
			Expect(err).To(BeNil())
			_, err = (*WLIDBC).CreateWorkloadIntent(WLI, "test", "capp1", "v1", "dig", "theName", false)
			Expect(err).To(BeNil())
			_, err = (*WLFIDBC).CreateWorkloadIfIntent(WLFI, "test", "capp1", "v1", "dig", "theName", "theSecondName", false)
			Expect(err).To(BeNil())
			_, err = (*WLFIDBC).CreateWorkloadIfIntent(WLFI, "test", "capp1", "v1", "dig", "theName", "theSecondName", false)
			Expect(err).To(HaveOccurred())
		})
		It("followed by get should return nil", func() {
			_, err := (*NCIDBC).CreateNetControlIntent(NCI, "test", "capp1", "v1", "dig", false)
			Expect(err).To(BeNil())
			_, err = (*WLIDBC).CreateWorkloadIntent(WLI, "test", "capp1", "v1", "dig", "theName", false)
			Expect(err).To(BeNil())
			_, err = (*WLFIDBC).CreateWorkloadIfIntent(WLFI, "test", "capp1", "v1", "dig", "theName", "theSecondName", false)
			Expect(err).To(BeNil())
			wlfi, err := (*WLFIDBC).GetWorkloadIfIntent("theThirdName", "test", "capp1", "v1", "dig", "theName", "theSecondName")
			Expect(wlfi).Should(Equal(WLFI))
		})
		It("followed by delete should return nil", func() {
			_, err := (*NCIDBC).CreateNetControlIntent(NCI, "test", "capp1", "v1", "dig", false)
			Expect(err).To(BeNil())
			_, err = (*WLIDBC).CreateWorkloadIntent(WLI, "test", "capp1", "v1", "dig", "theName", false)
			Expect(err).To(BeNil())
			_, err = (*WLFIDBC).CreateWorkloadIfIntent(WLFI, "test", "capp1", "v1", "dig", "theName", "theSecondName", false)
			Expect(err).To(BeNil())
			err = (*WLFIDBC).DeleteWorkloadIfIntent("theSecondName", "test", "capp1", "v1", "dig", "theName", "theSecondName")
			Expect(err).To(BeNil())
		})
	})

	Describe("Create workload if intent", func() {
		It("followed by create,get,delete,get workload if intent should return an error", func() {
			_, err := (*NCIDBC).CreateNetControlIntent(NCI, "test", "capp1", "v1", "dig", false)
			Expect(err).To(BeNil())
			_, err = (*WLIDBC).CreateWorkloadIntent(WLI, "test", "capp1", "v1", "dig", "theName", false)
			Expect(err).To(BeNil())
			_, err = (*WLFIDBC).CreateWorkloadIfIntent(WLFI, "test", "capp1", "v1", "dig", "theName", "theSecondName", false)
			Expect(err).To(BeNil())
			wlfi, err := (*WLFIDBC).GetWorkloadIfIntent("theThirdName", "test", "capp1", "v1", "dig", "theName", "theSecondName")
			Expect(wlfi).Should(Equal(WLFI))
			err = (*WLFIDBC).DeleteWorkloadIfIntent("theThirdName", "test", "capp1", "v1", "dig", "theName", "theSecondName")
			Expect(err).To(BeNil())
			wlfi, err = (*WLFIDBC).GetWorkloadIfIntent("theThirdName", "test", "capp1", "v1", "dig", "theName", "theSecondName")
			Expect(err).To(HaveOccurred())
		})
	})

	Describe("Get workload if intents", func() {
		It("should return error for non-existing record", func() {
			mdb.Err = pkgerrors.New("Error finding:")
			_, err := (*WLFIDBC).GetWorkloadIfIntents("test", "capp1", "v1", "dig", "theName", "theSecondName")
			Expect(err).To(HaveOccurred())
		})
	})

})
