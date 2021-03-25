package module_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/open-ness/EMCO/src/orchestrator/pkg/infra/db"
	"github.com/open-ness/EMCO/src/ovnaction/pkg/module"
	pkgerrors "github.com/pkg/errors"
)

var _ = Describe("Workloadintent", func() {
	var (
		NCI    module.NetControlIntent
		NCIDBC *module.NetControlIntentClient

		WLI    module.WorkloadIntent
		WLIDBC *module.WorkloadIntentClient

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

		mdb = new(db.MockDB)
		mdb.Err = nil
		db.DBconn = mdb
	})

	Describe("Create client intent", func() {
		It("with pre created netcontrolintent should return nil", func() {
			_, err := (*NCIDBC).CreateNetControlIntent(NCI, "test", "capp1", "v1", "dig", false)
			Expect(err).To(BeNil())
			_, err = (*WLIDBC).CreateWorkloadIntent(WLI, "test", "capp1", "v1", "dig", "theName", false)
			Expect(err).To(BeNil())
		})
		It("followed by create again should return error", func() {
			_, err := (*NCIDBC).CreateNetControlIntent(NCI, "test", "capp1", "v1", "dig", false)
			Expect(err).To(BeNil())
			_, err = (*WLIDBC).CreateWorkloadIntent(WLI, "test", "capp1", "v1", "dig", "theName", false)
			Expect(err).To(BeNil())
			mdb.Err = pkgerrors.New("Already exists:")
			_, err = (*WLIDBC).CreateWorkloadIntent(WLI, "test", "capp1", "v1", "dig", "theName", false)
			Expect(err).To(HaveOccurred())
		})
		It("followed by get should return nil", func() {
			_, err := (*NCIDBC).CreateNetControlIntent(NCI, "test", "capp1", "v1", "dig", false)
			Expect(err).To(BeNil())
			_, err = (*WLIDBC).CreateWorkloadIntent(WLI, "test", "capp1", "v1", "dig", "theName", false)
			Expect(err).To(BeNil())
			wli, err := (*WLIDBC).GetWorkloadIntent("theSecondName", "test", "capp1", "v1", "dig", "theName")
			Expect(wli).Should(Equal(WLI))
		})
		It("followed by delete should return nil", func() {
			_, err := (*NCIDBC).CreateNetControlIntent(NCI, "test", "capp1", "v1", "dig", false)
			Expect(err).To(BeNil())
			_, err = (*WLIDBC).CreateWorkloadIntent(WLI, "test", "capp1", "v1", "dig", "theName", false)
			Expect(err).To(BeNil())
			err = (*WLIDBC).DeleteWorkloadIntent("theSecondName", "test", "capp1", "v1", "dig", "theName")
			Expect(err).To(BeNil())
		})
	})

	Describe("Create workload intent", func() {
		It("followed by create,get,delete,get should return an error", func() {
			_, err := (*NCIDBC).CreateNetControlIntent(NCI, "test", "capp1", "v1", "dig", false)
			Expect(err).To(BeNil())
			_, err = (*WLIDBC).CreateWorkloadIntent(WLI, "test", "capp1", "v1", "dig", "theName", false)
			Expect(err).To(BeNil())
			wli, err := (*WLIDBC).GetWorkloadIntent("theSecondName", "test", "capp1", "v1", "dig", "theName")
			Expect(wli).Should(Equal(WLI))
			err = (*WLIDBC).DeleteWorkloadIntent("theSecondName", "test", "capp1", "v1", "dig", "theName")
			Expect(err).To(BeNil())
			wli, err = (*WLIDBC).GetWorkloadIntent("theSecondName", "test", "capp1", "v1", "dig", "theName")
			Expect(err).To(HaveOccurred())
		})
	})

	Describe("Get workload intents", func() {
		It("should return error for non-existing record", func() {
			mdb.Err = pkgerrors.New("Error finding:")
			_, err := (*WLIDBC).GetWorkloadIntents("test", "capp1", "v1", "dig", "theName")
			Expect(err).To(HaveOccurred())
		})
	})

	Describe("Delete workload intent", func() {
		It("should return error for non-existing record", func() {
			mdb.Err = pkgerrors.New("Error finding:")
			err := (*WLIDBC).DeleteWorkloadIntent("testtgi", "test", "capp1", "v1", "dig", "theName")
			Expect(err).To(HaveOccurred())
		})
		It("should return error for deleting parent without deleting child", func() {
			mdb.Err = pkgerrors.New("Can't delete parent without deleting child")
			err := (*WLIDBC).DeleteWorkloadIntent("testtgi", "test", "capp1", "v1", "dig", "theName")
			Expect(err).To(HaveOccurred())
		})
		It("should return error for general db error", func() {
			mdb.Err = pkgerrors.New("general error")
			err := (*WLIDBC).DeleteWorkloadIntent("testtgi", "test", "capp1", "v1", "dig", "theName")
			Expect(err).To(HaveOccurred())
		})
	})
})
