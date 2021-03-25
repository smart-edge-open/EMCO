package module_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/open-ness/EMCO/src/orchestrator/pkg/infra/db"
	"github.com/open-ness/EMCO/src/ovnaction/pkg/module"
	pkgerrors "github.com/pkg/errors"
)

var _ = Describe("Netcontrolintent", func() {
	var (
		NCI       module.NetControlIntent
		OTHER_NCI module.NetControlIntent
		NCIDBC    *module.NetControlIntentClient
		mdb       *db.NewMockDB
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
		OTHER_NCI = module.NetControlIntent{
			Metadata: module.Metadata{
				Name:        "Name",
				Description: "net control intent",
				UserData1:   "user data1",
				UserData2:   "user data2",
			},
		}
		mdb = new(db.NewMockDB)
		mdb.Err = nil
		db.DBconn = mdb
	})

	Describe("Create net intent", func() {
		It("should return nil", func() {
			_, err := (*NCIDBC).CreateNetControlIntent(NCI, "test", "capp1", "v1", "dig", false)
			Expect(err).To(BeNil())
		})
		It("followed by create again should return error", func() {
			_, err := (*NCIDBC).CreateNetControlIntent(NCI, "test", "capp1", "v1", "dig", false)
			Expect(err).To(BeNil())
			_, err = (*NCIDBC).CreateNetControlIntent(NCI, "test", "capp1", "v1", "dig", false)
			Expect(err).To(HaveOccurred())
		})
		It("followed by get should return nil", func() {
			_, err := (*NCIDBC).CreateNetControlIntent(NCI, "test", "capp1", "v1", "dig", false)
			Expect(err).To(BeNil())
			nci, err := (*NCIDBC).GetNetControlIntent("theName", "test", "capp1", "v1", "dig")
			Expect(nci).Should(Equal(NCI))
		})
		It("followed by delete should return nil", func() {
			_, err := (*NCIDBC).CreateNetControlIntent(NCI, "test", "capp1", "v1", "dig", false)
			Expect(err).To(BeNil())
			err = (*NCIDBC).DeleteNetControlIntent("testnci", "test", "capp1", "v1", "dig")
			Expect(err).To(BeNil())
		})
	})

	Describe("Create net intent", func() {
		It("followed by create,get,delete,get net intent should return an error", func() {
			_, err := (*NCIDBC).CreateNetControlIntent(NCI, "test", "capp1", "v1", "dig", false)
			Expect(err).To(BeNil())
			_, err = (*NCIDBC).GetNetControlIntent("theName", "test", "capp1", "v1", "dig")
			Expect(err).To(BeNil())
			err = (*NCIDBC).DeleteNetControlIntent("theName", "test", "capp1", "v1", "dig")
			Expect(err).To(BeNil())
			_, err = (*NCIDBC).GetNetControlIntent("theName", "test", "capp1", "v1", "dig")
			Expect(err).To(HaveOccurred())
		})
	})

	Describe("Get net control intents", func() {
		It("should return error for non-existing record", func() {
			mdb.Err = pkgerrors.New("Error finding:")
			_, err := (*NCIDBC).GetNetControlIntents("test", "capp1", "v1", "dig")
			Expect(err).To(HaveOccurred())
		})
	})

	Describe("Delete net control intent", func() {
		It("should return error for non-existing record", func() {
			mdb.Err = pkgerrors.New("Error finding:")
			err := (*NCIDBC).DeleteNetControlIntent("testtgi", "test", "capp1", "v1", "dig")
			Expect(err).To(HaveOccurred())
		})
		It("should return error for deleting parent without deleting child", func() {
			mdb.Err = pkgerrors.New("Can't delete parent without deleting child")
			err := (*NCIDBC).DeleteNetControlIntent("testtgi", "test", "capp1", "v1", "dig")
			Expect(err).To(HaveOccurred())
		})
		It("should return error for general db error", func() {
			mdb.Err = pkgerrors.New("general error")
			err := (*NCIDBC).DeleteNetControlIntent("testtgi", "test", "capp1", "v1", "dig")
			Expect(err).To(HaveOccurred())
		})
	})
	Describe("Create 2 net control intents", func() {
		It("should get all the net control intents for the project", func() {
			_, err := (*NCIDBC).CreateNetControlIntent(NCI, "test", "capp1", "v1", "dig", false)
			Expect(err).To(BeNil())
			_, err = (*NCIDBC).CreateNetControlIntent(OTHER_NCI, "test", "capp1", "v1", "dig", false)
			Expect(err).To(BeNil())
			rval, err := (*NCIDBC).GetNetControlIntents("test", "capp1", "v1", "dig")
			Expect(err).To(BeNil())
			Expect(len(rval)).To(Equal(2))
		})
	})
})
