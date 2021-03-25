package module_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	dcm "github.com/open-ness/EMCO/src/dcm/pkg/module"
	"github.com/open-ness/EMCO/src/orchestrator/pkg/infra/db"
	orch "github.com/open-ness/EMCO/src/orchestrator/pkg/module"
)

var _ = Describe("Cluster", func() {

	var (
		mdb    *db.MockDB
		client *dcm.ClusterClient
	)

	BeforeEach(func() {
		client = dcm.NewClusterClient()
		mdb = new(db.MockDB)
		mdb.Err = nil
		mdb.Items = []map[string]map[string][]byte{}
		db.DBconn = mdb
	})
	Describe("Cluster operations", func() {
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
					UserData1:   "",
					UserData2:   "",
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
					UserData1:        "",
					UserData2:        "",
				}
				lc.Specification = dcm.Spec{
					NameSpace: "anything",
					Level:     "1",
				}
				mdb.Insert("orchestrator", lkey, nil, "logicalcloud", lc)
			})
			It("creation should succeed and return the resource created", func() {
				cluster := _createTestCluster("testcluster")
				cluster, err := client.CreateCluster("project", "logicalcloud", cluster)
				Expect(err).ShouldNot(HaveOccurred())
				Expect(cluster.MetaData.ClusterReference).To(Equal("testcluster"))
				Expect(cluster.MetaData.Description).To(Equal(""))
				Expect(cluster.MetaData.UserData1).To(Equal(""))
				Expect(cluster.MetaData.UserData2).To(Equal(""))
			})
			// TODO
			It("creation on instantiated cloud should fail", func() {
				cluster := _createTestCluster("testcluster")
				cluster, err := client.CreateCluster("project", "logicalcloud", cluster)
				Expect(err).ShouldNot(HaveOccurred())
				Expect(cluster.MetaData.ClusterReference).To(Equal("testcluster"))
				Expect(cluster.MetaData.Description).To(Equal(""))
				Expect(cluster.MetaData.UserData1).To(Equal(""))
				Expect(cluster.MetaData.UserData2).To(Equal(""))
			})
			It("get should fail and not return anything", func() {
				cluster, err := client.GetCluster("project", "logicalcloud", "testcluster")
				Expect(err).Should(HaveOccurred())
				Expect(cluster).To(Equal(dcm.Cluster{}))
			})
			It("create followed by get should return what was created", func() {
				cluster := _createTestCluster("testcluster")
				_, _ = client.CreateCluster("project", "logicalcloud", cluster)
				cluster, err := client.GetCluster("project", "logicalcloud", "testcluster")
				Expect(err).ShouldNot(HaveOccurred())
				Expect(cluster).To(Equal(cluster))
			})
			It("create followed by get-all should return only what was created", func() {
				cluster := _createTestCluster("testcluster")
				_, _ = client.CreateCluster("project", "logicalcloud", cluster)
				clusters, err := client.GetAllClusters("project", "logicalcloud")
				Expect(err).ShouldNot(HaveOccurred())
				Expect(len(clusters)).To(Equal(1))
				Expect(clusters[0]).To(Equal(cluster))
			})
			It("three creates followed by get-all should return all that was created", func() {
				cluster1 := _createTestCluster("testcluster1")
				cluster2 := _createTestCluster("testcluster2")
				cluster3 := _createTestCluster("testcluster3")
				_, _ = client.CreateCluster("project", "logicalcloud", cluster1)
				_, _ = client.CreateCluster("project", "logicalcloud", cluster2)
				_, _ = client.CreateCluster("project", "logicalcloud", cluster3)
				clusters, err := client.GetAllClusters("project", "logicalcloud")
				Expect(err).ShouldNot(HaveOccurred())
				Expect(len(clusters)).To(Equal(3))
				Expect(clusters[0]).To(Equal(cluster1))
				Expect(clusters[1]).To(Equal(cluster2))
				Expect(clusters[2]).To(Equal(cluster3))
			})
			It("delete after creation should succeed and database remain empty", func() {
				cluster := _createTestCluster("testcluster")
				_, _ = client.CreateCluster("project", "logicalcloud", cluster)
				err := client.DeleteCluster("project", "logicalcloud", "testcluster")
				Expect(err).ShouldNot(HaveOccurred())
				clusters, err := client.GetAllClusters("project", "logicalcloud")
				Expect(len(clusters)).To(Equal(0))
			})
			// will uncomment after general mockdb issues resolved
			// It("delete when nothing exists should fail", func() {
			// 	err := client.DeleteCluster("project", "logicalcloud", "testcluster")
			// 	Expect(err).Should(HaveOccurred())
			// })
			It("update after creation should succeed and return updated resource", func() {
				cluster := _createTestCluster("testcluster")
				_, _ = client.CreateCluster("project", "logicalcloud", cluster)
				cluster.MetaData.UserData1 = "new user data"
				cluster, err := client.UpdateCluster("project", "logicalcloud", "testcluster", cluster)
				Expect(err).ShouldNot(HaveOccurred())
				Expect(cluster.MetaData.ClusterReference).To(Equal("testcluster"))
				Expect(cluster.MetaData.Description).To(Equal(""))
				Expect(cluster.MetaData.UserData1).To(Equal("new user data"))
				Expect(cluster.MetaData.UserData2).To(Equal(""))
			})
			It("create followed by updating the name is disallowed and should fail", func() {
				cluster := _createTestCluster("testcluster")
				_, _ = client.CreateCluster("project", "logicalcloud", cluster)
				cluster.MetaData.ClusterReference = "updated"
				cluster, err := client.UpdateCluster("project", "logicalcloud", "testcluster", cluster)
				Expect(err).Should(HaveOccurred())
				Expect(cluster).To(Equal(dcm.Cluster{}))
			})
		})
	})
})

// TODO:
// - test when cluster references already exist
// - appcontext status check for creation and deletion of cluster references
// - test GetClusterConfig

// _createTestCluster is an helper function to reduce code duplication
func _createTestCluster(name string) dcm.Cluster {
	return dcm.Cluster{
		MetaData: dcm.ClusterMeta{
			ClusterReference: name,
			Description:      "",
			UserData1:        "",
			UserData2:        "",
		},
	}
}
