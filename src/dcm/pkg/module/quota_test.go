package module_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	dcm "github.com/open-ness/EMCO/src/dcm/pkg/module"
	"github.com/open-ness/EMCO/src/orchestrator/pkg/infra/db"
	orch "github.com/open-ness/EMCO/src/orchestrator/pkg/module"
)

var _ = Describe("Quota", func() {

	var (
		mdb    *db.MockDB
		client *dcm.QuotaClient
	)

	BeforeEach(func() {
		client = dcm.NewQuotaClient()
		mdb = new(db.MockDB)
		mdb.Err = nil
		mdb.Items = []map[string]map[string][]byte{}
		db.DBconn = mdb
	})
	Describe("Quota operations", func() {
		Context("from an empty database", func() {
			BeforeEach(func() {
				// create project in mocked db
				okey := orch.ProjectKey{
					ProjectName: "project",
				}
				p := orch.Project{}
				p.MetaData = orch.ProjectMetaData{
					Name:        "project",
					Description: "",
				}
				mdb.Insert("orchestrator", okey, nil, "projectmetadata", p)
				// create logical cloud in mocked db
				lkey := dcm.LogicalCloudKey{
					Project:          "project",
					LogicalCloudName: "logicalcloud",
				}
				lc := dcm.LogicalCloud{}
				lc.MetaData = dcm.MetaDataList{
					LogicalCloudName: "logicalcloud",
					Description:      "",
				}
				lc.Specification = dcm.Spec{
					NameSpace: "anything",
					Level:     "1",
				}
				mdb.Insert("orchestrator", lkey, nil, "logicalcloud", lc)
			})
			It("creation should succeed and return the resource created", func() {
				quota := _createTestQuota("testquota")
				quota, err := client.CreateQuota("project", "logicalcloud", quota)
				Expect(err).ShouldNot(HaveOccurred())
				Expect(quota.MetaData.QuotaName).To(Equal("testquota"))
				Expect(quota.MetaData.Description).To(Equal(""))
			})
			It("get should fail and not return anything", func() {
				quota, err := client.GetQuota("project", "logicalcloud", "testquota")
				Expect(err).Should(HaveOccurred())
				Expect(quota).To(Equal(dcm.Quota{}))
			})
			It("create followed by get should return what was created", func() {
				quota := _createTestQuota("testquota")
				_, _ = client.CreateQuota("project", "logicalcloud", quota)
				quota, err := client.GetQuota("project", "logicalcloud", "testquota")
				Expect(err).ShouldNot(HaveOccurred())
				Expect(quota).To(Equal(quota))
			})
			It("create followed by get-all should return only what was created", func() {
				quota := _createTestQuota("testquota")
				_, _ = client.CreateQuota("project", "logicalcloud", quota)
				quotas, err := client.GetAllQuotas("project", "logicalcloud")
				Expect(err).ShouldNot(HaveOccurred())
				Expect(len(quotas)).To(Equal(1))
				Expect(quotas[0]).To(Equal(quota))
			})
			It("three creates followed by get-all should return all that was created", func() {
				quota1 := _createTestQuota("testquota1")
				quota2 := _createTestQuota("testquota2")
				quota3 := _createTestQuota("testquota3")
				_, _ = client.CreateQuota("project", "logicalcloud", quota1)
				_, _ = client.CreateQuota("project", "logicalcloud", quota2)
				_, _ = client.CreateQuota("project", "logicalcloud", quota3)
				quotas, err := client.GetAllQuotas("project", "logicalcloud")
				Expect(err).ShouldNot(HaveOccurred())
				Expect(len(quotas)).To(Equal(3))
				Expect(quotas[0]).To(Equal(quota1))
				Expect(quotas[1]).To(Equal(quota2))
				Expect(quotas[2]).To(Equal(quota3))
			})
			It("delete after creation should succeed and database remain empty", func() {
				quota := _createTestQuota("testquota")
				_, _ = client.CreateQuota("project", "logicalcloud", quota)
				err := client.DeleteQuota("project", "logicalcloud", "testquota")
				Expect(err).ShouldNot(HaveOccurred())
				quotas, err := client.GetAllQuotas("project", "logicalcloud")
				Expect(len(quotas)).To(Equal(0))
			})
			// will uncomment after general mockdb issues resolved
			// It("delete when nothing exists should fail", func() {
			// 	err := client.DeleteQuota("project", "logicalcloud", "testquota")
			// 	Expect(err).Should(HaveOccurred())
			// })
			It("update after creation should succeed and return updated resource", func() {
				quota := _createTestQuota("testquota")
				_, _ = client.CreateQuota("project", "logicalcloud", quota)
				quota.MetaData.Description = "new description"
				quota, err := client.UpdateQuota("project", "logicalcloud", "testquota", quota)
				Expect(err).ShouldNot(HaveOccurred())
				Expect(quota.MetaData.QuotaName).To(Equal("testquota"))
				Expect(quota.MetaData.Description).To(Equal("new description"))
			})
			It("create followed by updating the name is disallowed and should fail", func() {
				quota := _createTestQuota("testquota")
				_, _ = client.CreateQuota("project", "logicalcloud", quota)
				quota.MetaData.QuotaName = "updated"
				quota, err := client.UpdateQuota("project", "logicalcloud", "testquota", quota)
				Expect(err).Should(HaveOccurred())
				Expect(quota).To(Equal(dcm.Quota{}))
			})
		})
	})
})

func _createTestQuota(name string) dcm.Quota {
	quota := dcm.Quota{}
	quota.MetaData = dcm.QMetaDataList{
		QuotaName:   name,
		Description: "",
	}
	quota.Specification = map[string]string{}
	quota.Specification["limits.cpu"] = "4"
	quota.Specification["limits.memory"] = "4096"
	return quota
}
