package module_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	dcm "github.com/open-ness/EMCO/src/dcm/pkg/module"
	"github.com/open-ness/EMCO/src/orchestrator/pkg/infra/db"
	orch "github.com/open-ness/EMCO/src/orchestrator/pkg/module"
)

var _ = Describe("Userpermissions", func() {

	var (
		mdb    *db.MockDB
		client *dcm.UserPermissionClient
	)

	BeforeEach(func() {
		client = dcm.NewUserPermissionClient()
		mdb = new(db.MockDB)
		mdb.Err = nil
		mdb.Items = []map[string]map[string][]byte{}
		db.DBconn = mdb
	})
	Describe("User permission operations", func() {
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
					NameSpace: "testns",
					Level:     "1",
				}
				mdb.Insert("orchestrator", lkey, nil, "logicalcloud", lc)
			})
			It("creation should succeed and return the resource created", func() {
				up := _createTestUserPermission("testup", "testns")
				userPermission, err := client.CreateUserPerm("project", "logicalcloud", up)
				Expect(err).ShouldNot(HaveOccurred())
				Expect(userPermission.MetaData.UserPermissionName).To(Equal("testup"))
				Expect(userPermission.Specification.Namespace).To(Equal("testns"))
				Expect(userPermission.Specification.APIGroups).To(Equal([]string{"", "apps"}))
				Expect(userPermission.Specification.Resources).To(Equal([]string{"deployments", "pods"}))
				Expect(userPermission.Specification.Verbs).To(Equal([]string{"get", "list"}))
			})
			It("creation should succeed and return the resource created (cluster-wide)", func() {
				up := _createTestUserPermission("testup", "")
				userPermission, err := client.CreateUserPerm("project", "logicalcloud", up)
				Expect(err).ShouldNot(HaveOccurred())
				Expect(userPermission.MetaData.UserPermissionName).To(Equal("testup"))
				Expect(userPermission.Specification.Namespace).To(Equal(""))
				Expect(userPermission.Specification.APIGroups).To(Equal([]string{"", "apps"}))
				Expect(userPermission.Specification.Resources).To(Equal([]string{"deployments", "pods"}))
				Expect(userPermission.Specification.Verbs).To(Equal([]string{"get", "list"}))
			})
			It("get should fail and not return anything", func() {
				userPermission, err := client.GetUserPerm("project", "logicalcloud", "testup")
				Expect(err).Should(HaveOccurred())
				Expect(userPermission).To(Equal(dcm.UserPermission{}))
			})
			It("create followed by get should return what was created", func() {
				up := _createTestUserPermission("testup", "testns")
				_, _ = client.CreateUserPerm("project", "logicalcloud", up)
				userPermission, err := client.GetUserPerm("project", "logicalcloud", "testup")
				Expect(err).ShouldNot(HaveOccurred())
				Expect(userPermission).To(Equal(up))
			})
			It("create followed by get-all should return only what was created", func() {
				up := _createTestUserPermission("testup", "testns")
				_, _ = client.CreateUserPerm("project", "logicalcloud", up)
				userPermissions, err := client.GetAllUserPerms("project", "logicalcloud")
				Expect(err).ShouldNot(HaveOccurred())
				Expect(len(userPermissions)).To(Equal(1))
				Expect(userPermissions[0]).To(Equal(up))
			})
			It("three creates followed by get-all should return all that was created", func() {
				up1 := _createTestUserPermission("testup1", "testns")
				up2 := _createTestUserPermission("testup2", "testns")
				up3 := _createTestUserPermission("testup3", "testns")
				_, _ = client.CreateUserPerm("project", "logicalcloud", up1)
				_, _ = client.CreateUserPerm("project", "logicalcloud", up2)
				_, _ = client.CreateUserPerm("project", "logicalcloud", up3)
				userPermissions, err := client.GetAllUserPerms("project", "logicalcloud")
				Expect(err).ShouldNot(HaveOccurred())
				Expect(len(userPermissions)).To(Equal(3))
				Expect(userPermissions[0]).To(Equal(up1))
				Expect(userPermissions[1]).To(Equal(up2))
				Expect(userPermissions[2]).To(Equal(up3))
			})
			It("delete after creation should succeed and database remain empty", func() {
				up := _createTestUserPermission("testup", "testns")
				_, _ = client.CreateUserPerm("project", "logicalcloud", up)
				err := client.DeleteUserPerm("project", "logicalcloud", "testup")
				Expect(err).ShouldNot(HaveOccurred())
				userPermissions, err := client.GetAllUserPerms("project", "logicalcloud")
				Expect(len(userPermissions)).To(Equal(0))
			})
			// will uncomment after general mockdb issues resolved
			// It("delete when nothing exists should fail", func() {
			// 	err := client.DeleteUserPerm("project", "logicalcloud", "testup")
			// 	Expect(err).Should(HaveOccurred())
			// })
			It("update after creation should succeed and return updated resource", func() {
				up := _createTestUserPermission("testup", "testns")
				_, _ = client.CreateUserPerm("project", "logicalcloud", up)
				up.Specification.APIGroups = []string{"", "apps", "k8splugin.io"}
				userPermission, err := client.UpdateUserPerm("project", "logicalcloud", "testup", up)
				Expect(err).ShouldNot(HaveOccurred())
				Expect(userPermission.MetaData.UserPermissionName).To(Equal("testup"))
				Expect(userPermission.Specification.APIGroups).To(Equal([]string{"", "apps", "k8splugin.io"}))
				Expect(userPermission.Specification.Resources).To(Equal([]string{"deployments", "pods"}))
				Expect(userPermission.Specification.Verbs).To(Equal([]string{"get", "list"}))
			})
			It("create followed by updating the name is disallowed and should fail", func() {
				up := _createTestUserPermission("testup", "testns")
				_, _ = client.CreateUserPerm("project", "logicalcloud", up)
				up.MetaData.UserPermissionName = "updated"
				userPermission, err := client.UpdateUserPerm("project", "logicalcloud", "testup", up)
				Expect(err).Should(HaveOccurred())
				Expect(userPermission).To(Equal(dcm.UserPermission{}))
			})
		})
	})
})

// _createTestUserPermission is an helper function to reduce code duplication
func _createTestUserPermission(name string, namespace string) dcm.UserPermission {

	up := dcm.UserPermission{}

	up.MetaData = dcm.UPMetaDataList{
		UserPermissionName: name,
		Description:        "",
		UserData1:          "",
		UserData2:          "",
	}
	up.Specification = dcm.UPSpec{
		Namespace: namespace,
		APIGroups: []string{"", "apps"},
		Resources: []string{"deployments", "pods"},
		Verbs:     []string{"get", "list"},
	}

	return up
}
