package module_test

import (
	"fmt"
	"strings"

	"github.com/golang/mock/gomock"
	"github.com/golang/protobuf/proto"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	dcm "github.com/open-ness/EMCO/src/dcm/pkg/module"
	"github.com/open-ness/EMCO/src/orchestrator/pkg/infra/contextdb"
	"github.com/open-ness/EMCO/src/orchestrator/pkg/infra/db"
	orch "github.com/open-ness/EMCO/src/orchestrator/pkg/module"
	"github.com/open-ness/EMCO/src/orchestrator/pkg/module/controller"
	"github.com/open-ness/EMCO/src/orchestrator/pkg/module/types"
	rsync "github.com/open-ness/EMCO/src/rsync/pkg/db"
	"github.com/open-ness/EMCO/src/rsync/pkg/grpc/installapp"
)

var _ = Describe("Logicalcloud", func() {

	var (
		mdb    *db.MockDB              // for MongoDB/database mocking
		edb    *contextdb.MockConDb    // for etcd/appcontext mocking
		client *dcm.LogicalCloudClient // for DCM operations
	)

	BeforeEach(func() {
		mdb = new(db.MockDB)
		mdb.Err = nil
		mdb.Items = []map[string]map[string][]byte{}
		db.DBconn = mdb
		edb = new(contextdb.MockConDb)
		edb.Err = nil
		contextdb.Db = edb
		client = dcm.NewLogicalCloudClient()
	})
	Describe("Logical Cloud operations", func() {
		Context("from having a L1 logical cloud already created", func() {
			BeforeEach(func() {
				_createExistingLogicalCloud(mdb, "1", true, false)
			})
			It("instantiation (non-privileged) should be successful", func() {

				// Mock gRPC InstallApp()
				var gsia = &grpcSignature{}
				gsia.grpcReq = nil
				gsia.grpcRsp = []interface{}{&installapp.InstallAppResponse{
					AppContextInstalled: true,
				}, nil}
				testMockGrpc(mockinstallapp.EXPECT().InstallApp, gsia.grpcReq, gsia.grpcRsp)

				// Mock gRPC ReadyNotify()
				var gsrn = &grpcSignature{}
				gsrn.grpcReq = nil
				testMockGrpcRN(mockreadynotify.EXPECT().Alert, gsrn.grpcReq)

				// these are used for specifying expected keys and bodies for appcontext
				var mockedKeys []string
				var mockedValues []string

				// create these outside of BeforeEach with the purpose of being passed to the tested functions
				lc := _createTestLogicalCloud("testlc", "1")
				cl := _createTestClusterReference("testcp", "testcl")
				quota := _createTestQuota("testquota")
				up := _createTestUserPermission("testup", "testns")
				err := dcm.Instantiate("project", lc, []dcm.Cluster{cl}, []dcm.Quota{quota}, []dcm.UserPermission{up})

				etcdKeys, _ := contextdb.Db.GetAllKeys("/")
				appcontextId := strings.Split(etcdKeys[0], "/")[2]
				var val string

				// check that the instantiation generated the expected keys

				Expect(err).ShouldNot(HaveOccurred())

				Expect(len(etcdKeys)).To(Equal(15))

				mockedKeys = append(mockedKeys, fmt.Sprintf("/context/%v/", appcontextId))
				mockedKeys = append(mockedKeys, fmt.Sprintf("/context/%v/app/logical-cloud/", appcontextId))
				mockedKeys = append(mockedKeys, fmt.Sprintf("/context/%v/app/logical-cloud/cluster/testcp+testcl/", appcontextId))
				mockedKeys = append(mockedKeys, fmt.Sprintf("/context/%v/app/logical-cloud/cluster/testcp+testcl/resource/testns+Namespace/", appcontextId))
				mockedKeys = append(mockedKeys, fmt.Sprintf("/context/%v/app/logical-cloud/cluster/testcp+testcl/resource/testlc-user-csr+CertificateSigningRequest/", appcontextId))
				mockedKeys = append(mockedKeys, fmt.Sprintf("/context/%v/app/logical-cloud/cluster/testcp+testcl/resource/testlc-user-csr+CertificateSigningRequest/subresource/approval/", appcontextId))
				mockedKeys = append(mockedKeys, fmt.Sprintf("/context/%v/app/logical-cloud/cluster/testcp+testcl/resource/testlc-role0+Role/", appcontextId))
				mockedKeys = append(mockedKeys, fmt.Sprintf("/context/%v/app/logical-cloud/cluster/testcp+testcl/resource/testlc-roleBinding0+RoleBinding/", appcontextId))
				mockedKeys = append(mockedKeys, fmt.Sprintf("/context/%v/app/logical-cloud/cluster/testcp+testcl/resource/testquota+ResourceQuota/", appcontextId))
				mockedKeys = append(mockedKeys, fmt.Sprintf("/context/%v/app/logical-cloud/cluster/testcp+testcl/resource/instruction/order/", appcontextId))
				mockedKeys = append(mockedKeys, fmt.Sprintf("/context/%v/app/logical-cloud/cluster/testcp+testcl/resource/instruction/dependency/", appcontextId))
				mockedKeys = append(mockedKeys, fmt.Sprintf("/context/%v/app/logical-cloud/cluster/testcp+testcl/resource/testlc-user-csr+CertificateSigningRequest/subresource/instruction/order/", appcontextId))
				mockedKeys = append(mockedKeys, fmt.Sprintf("/context/%v/app/logical-cloud/cluster/testcp+testcl/resource/testlc-user-csr+CertificateSigningRequest/subresource/instruction/dependency/", appcontextId))
				mockedKeys = append(mockedKeys, fmt.Sprintf("/context/%v/app/instruction/order/", appcontextId))
				mockedKeys = append(mockedKeys, fmt.Sprintf("/context/%v/app/instruction/dependency/", appcontextId))

				for k, _ := range etcdKeys {
					Expect(etcdKeys[k]).To(Equal(mockedKeys[k]))
				}

				// check that the content of the keys is also the one expected
				mockedValues = append(mockedValues, appcontextId) //[0]
				mockedValues = append(mockedValues, "logical-cloud")
				mockedValues = append(mockedValues, "testcp+testcl")
				mockedValues = append(mockedValues, nsmock())
				mockedValues = append(mockedValues, "") // [4] not testing full equality
				mockedValues = append(mockedValues, "") // [5] not testing full equality
				mockedValues = append(mockedValues, rolemock())
				mockedValues = append(mockedValues, rolebindingmock())
				mockedValues = append(mockedValues, quotamock())
				mockedValues = append(mockedValues, resordermock()) // [9]
				mockedValues = append(mockedValues, resdepmock())
				mockedValues = append(mockedValues, `{"subresorder":["approval"]}`)
				mockedValues = append(mockedValues, `{"subresdependency":{"approval":"go"}}`)
				mockedValues = append(mockedValues, `{"apporder":["logical-cloud"]}`)
				mockedValues = append(mockedValues, `{"appdependency":{"logical-cloud":"go"}}`) // [14]

				for _, k := range [13]int{0, 1, 2, 3, 6, 7, 8, 9, 10, 11, 12, 13, 14} {
					contextdb.Db.Get(mockedKeys[k], &val)
					Expect(val).To(Equal(mockedValues[k]))
				}

				// check the untested values separately (these change every run due to crypto or timestamps)
				contextdb.Db.Get(mockedKeys[4], &val)
				Expect(strings.Contains(val, "kind: CertificateSigningRequest")).Should(BeTrue())
				Expect(strings.Contains(val, "name: testlc-user-csr")).Should(BeTrue())
				Expect(strings.Contains(val, "request: LS0")).Should(BeTrue())
				contextdb.Db.Get(mockedKeys[5], &val)
				Expect(strings.Contains(val, `"message":"Approved for Logical Cloud authentication","reason":"LogicalCloud","type":"Approved"}`)).Should(BeTrue())
			})
		})
		Context("from having a Privileged L1 logical cloud already created", func() {
			BeforeEach(func() {
				_createExistingLogicalCloud(mdb, "1", true, true)
			})
			It("instantiation (privileged) should be successful", func() {

				// Mock gRPC InstallApp()
				var gsia = &grpcSignature{}
				gsia.grpcReq = nil
				gsia.grpcRsp = []interface{}{&installapp.InstallAppResponse{
					AppContextInstalled: true,
				}, nil}
				testMockGrpc(mockinstallapp.EXPECT().InstallApp, gsia.grpcReq, gsia.grpcRsp)

				// Mock gRPC ReadyNotify()
				var gsrn = &grpcSignature{}
				gsrn.grpcReq = nil
				testMockGrpcRN(mockreadynotify.EXPECT().Alert, gsrn.grpcReq)

				// these are used for specifying expected keys and bodies for appcontext
				var mockedKeys []string
				var mockedValues []string

				// create these outside of BeforeEach with the purpose of being passed to the tested functions
				lc := _createTestLogicalCloud("testlc", "1")
				cl := _createTestClusterReference("testcp", "testcl")
				quota := _createTestQuota("testquota")
				up1 := _createTestUserPermission("testup", "testns")
				up2 := _createTestUserPermission("testup", "")
				err := dcm.Instantiate("project", lc, []dcm.Cluster{cl}, []dcm.Quota{quota}, []dcm.UserPermission{up1, up2})

				etcdKeys, _ := contextdb.Db.GetAllKeys("/")
				appcontextId := strings.Split(etcdKeys[0], "/")[2]
				var val string

				// check that the instantiation generated the expected keys

				Expect(err).ShouldNot(HaveOccurred())

				Expect(len(etcdKeys)).To(Equal(17))

				mockedKeys = append(mockedKeys, fmt.Sprintf("/context/%v/", appcontextId))
				mockedKeys = append(mockedKeys, fmt.Sprintf("/context/%v/app/logical-cloud/", appcontextId))
				mockedKeys = append(mockedKeys, fmt.Sprintf("/context/%v/app/logical-cloud/cluster/testcp+testcl/", appcontextId))
				mockedKeys = append(mockedKeys, fmt.Sprintf("/context/%v/app/logical-cloud/cluster/testcp+testcl/resource/testns+Namespace/", appcontextId))
				mockedKeys = append(mockedKeys, fmt.Sprintf("/context/%v/app/logical-cloud/cluster/testcp+testcl/resource/testlc-user-csr+CertificateSigningRequest/", appcontextId))
				mockedKeys = append(mockedKeys, fmt.Sprintf("/context/%v/app/logical-cloud/cluster/testcp+testcl/resource/testlc-user-csr+CertificateSigningRequest/subresource/approval/", appcontextId))
				mockedKeys = append(mockedKeys, fmt.Sprintf("/context/%v/app/logical-cloud/cluster/testcp+testcl/resource/testlc-role0+Role/", appcontextId))
				mockedKeys = append(mockedKeys, fmt.Sprintf("/context/%v/app/logical-cloud/cluster/testcp+testcl/resource/testlc-clusterRole1+ClusterRole/", appcontextId))
				mockedKeys = append(mockedKeys, fmt.Sprintf("/context/%v/app/logical-cloud/cluster/testcp+testcl/resource/testlc-roleBinding0+RoleBinding/", appcontextId))
				mockedKeys = append(mockedKeys, fmt.Sprintf("/context/%v/app/logical-cloud/cluster/testcp+testcl/resource/testlc-clusterRoleBinding1+ClusterRoleBinding/", appcontextId))
				mockedKeys = append(mockedKeys, fmt.Sprintf("/context/%v/app/logical-cloud/cluster/testcp+testcl/resource/testquota+ResourceQuota/", appcontextId))
				mockedKeys = append(mockedKeys, fmt.Sprintf("/context/%v/app/logical-cloud/cluster/testcp+testcl/resource/instruction/order/", appcontextId))
				mockedKeys = append(mockedKeys, fmt.Sprintf("/context/%v/app/logical-cloud/cluster/testcp+testcl/resource/instruction/dependency/", appcontextId))
				mockedKeys = append(mockedKeys, fmt.Sprintf("/context/%v/app/logical-cloud/cluster/testcp+testcl/resource/testlc-user-csr+CertificateSigningRequest/subresource/instruction/order/", appcontextId))
				mockedKeys = append(mockedKeys, fmt.Sprintf("/context/%v/app/logical-cloud/cluster/testcp+testcl/resource/testlc-user-csr+CertificateSigningRequest/subresource/instruction/dependency/", appcontextId))
				mockedKeys = append(mockedKeys, fmt.Sprintf("/context/%v/app/instruction/order/", appcontextId))
				mockedKeys = append(mockedKeys, fmt.Sprintf("/context/%v/app/instruction/dependency/", appcontextId))

				for k, _ := range etcdKeys {
					Expect(etcdKeys[k]).To(Equal(mockedKeys[k]))
				}

				// check that the content of the keys is also the one expected
				mockedValues = append(mockedValues, appcontextId) //[0]
				mockedValues = append(mockedValues, "logical-cloud")
				mockedValues = append(mockedValues, "testcp+testcl")
				mockedValues = append(mockedValues, nsmock())
				mockedValues = append(mockedValues, "") // [4] not testing full equality
				mockedValues = append(mockedValues, "") // [5] not testing full equality
				mockedValues = append(mockedValues, rolemock())
				mockedValues = append(mockedValues, clusterrolemock())
				mockedValues = append(mockedValues, rolebindingmock())
				mockedValues = append(mockedValues, clusterrolebindingmock())
				mockedValues = append(mockedValues, quotamock())
				mockedValues = append(mockedValues, resordermockPrivileged()) // [11]
				mockedValues = append(mockedValues, resdepmockPrivileged())
				mockedValues = append(mockedValues, `{"subresorder":["approval"]}`)
				mockedValues = append(mockedValues, `{"subresdependency":{"approval":"go"}}`)
				mockedValues = append(mockedValues, `{"apporder":["logical-cloud"]}`)
				mockedValues = append(mockedValues, `{"appdependency":{"logical-cloud":"go"}}`) // [16]

				for _, k := range [15]int{0, 1, 2, 3, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16} {
					contextdb.Db.Get(mockedKeys[k], &val)
					Expect(val).To(Equal(mockedValues[k]))
				}

				// check the untested values separately (these change every run due to crypto or timestamps)
				contextdb.Db.Get(mockedKeys[4], &val)
				Expect(strings.Contains(val, "kind: CertificateSigningRequest")).Should(BeTrue())
				Expect(strings.Contains(val, "name: testlc-user-csr")).Should(BeTrue())
				Expect(strings.Contains(val, "request: LS0")).Should(BeTrue())
				contextdb.Db.Get(mockedKeys[5], &val)
				Expect(strings.Contains(val, `"message":"Approved for Logical Cloud authentication","reason":"LogicalCloud","type":"Approved"}`)).Should(BeTrue())
			})
		})
		Context("from having a L1 logical cloud already created (but without a primary namespace user permission)", func() {
			BeforeEach(func() {
				_createExistingLogicalCloud(mdb, "1", false, false)
			})
			It("instantiation (non-privileged) should be successful", func() {

				// Mock gRPC InstallApp()
				var gsia = &grpcSignature{}
				gsia.grpcReq = nil
				gsia.grpcRsp = []interface{}{&installapp.InstallAppResponse{
					AppContextInstalled: true,
				}, nil}
				testMockGrpc(mockinstallapp.EXPECT().InstallApp, gsia.grpcReq, gsia.grpcRsp)

				// Mock gRPC ReadyNotify()
				var gsrn = &grpcSignature{}
				gsrn.grpcReq = nil
				testMockGrpcRN(mockreadynotify.EXPECT().Alert, gsrn.grpcReq)

				// create these outside of BeforeEach with the purpose of being passed to the tested functions
				lc := _createTestLogicalCloud("testlc", "1")
				cl := _createTestClusterReference("testcp", "testcl")
				quota := _createTestQuota("testquota")
				err := dcm.Instantiate("project", lc, []dcm.Cluster{cl}, []dcm.Quota{quota}, []dcm.UserPermission{})

				// check that the instantiation failed
				Expect(err).Should(HaveOccurred())
			})
		})
		// Temporarily disabled until issue with etcd keys is resolved
		Context("from having a L0 logical cloud already created", func() {
			BeforeEach(func() {
				rclient := rsync.NewCloudConfigClient()
				_createExistingLogicalCloud(mdb, "0", false, false)
				_, _ = rclient.CreateCloudConfig("testcp", "testcl", "0", "default", "123")
			})
			// 	It("instantiation should be successful", func() {

			// 		// Mock gRPC InstallApp()
			// 		var gsia = &grpcSignature{}
			// 		gsia.grpcReq = nil
			// 		gsia.grpcRsp = []interface{}{&installapp.InstallAppResponse{
			// 			AppContextInstalled: true,
			// 		}, nil}
			// 		testMockGrpc(mockinstallapp.EXPECT().InstallApp, gsia.grpcReq, gsia.grpcRsp)

			// 		// Mock gRPC ReadyNotify()
			// 		var gsrn = &grpcSignature{}
			// 		gsrn.grpcReq = nil
			// 		testMockGrpcRN(mockreadynotify.EXPECT().Alert, gsrn.grpcReq)

			// 		// these are used for specifying expected keys and bodies for appcontext
			// 		var mockedKeys []string
			// 		var mockedValues []string
			// 		// create these outside of BeforeEach with the purpose of being passed to the tested functions
			// 		lc := _createTestLogicalCloud("testlc", "0")
			// 		cl := _createTestClusterReference("testcp", "testcl")
			// 		quota := _createTestQuota("testquota")
			// 		err := dcm.Instantiate("project", lc, []dcm.Cluster{cl}, []dcm.Quota{quota})

			// 		etcdKeys, _ := contextdb.Db.GetAllKeys("/")
			// 		appcontextId := strings.Split(etcdKeys[0], "/")[2]
			// 		var val string

			// 		// check that the instantiation generated the expected keys

			// 		Expect(err).ShouldNot(HaveOccurred())

			// 		Expect(len(etcdKeys)).To(Equal(5))

			// 		// TODO fix apply.go then uncomment keys below
			// 		mockedKeys = append(mockedKeys, fmt.Sprintf("/context/%v/", appcontextId))
			// 		mockedKeys = append(mockedKeys, fmt.Sprintf("/context/%v/app/logical-cloud/", appcontextId))
			// 		mockedKeys = append(mockedKeys, fmt.Sprintf("/context/%v/app/logical-cloud/cluster/testcp+testcl/", appcontextId))
			// 		mockedKeys = append(mockedKeys, fmt.Sprintf("/context/%v/app/logical-cloud/cluster/testcp+testcl/resource/instruction/order/", appcontextId))
			// 		// mockedKeys = append(mockedKeys, fmt.Sprintf("/context/%v/app/logical-cloud/cluster/testcp+testcl/resource/instruction/dependency/", appcontextId))
			// 		mockedKeys = append(mockedKeys, fmt.Sprintf("/context/%v/app/instruction/order/", appcontextId))
			// 		// mockedKeys = append(mockedKeys, fmt.Sprintf("/context/%v/app/instruction/dependency/", appcontextId))

			// 		for k, _ := range etcdKeys {
			// 			Expect(etcdKeys[k]).To(Equal(mockedKeys[k]))
			// 		}

			// 		// check that the content of the keys is also the one expected

			// 		mockedValues = append(mockedValues, appcontextId)
			// 		mockedValues = append(mockedValues, "logical-cloud")
			// 		mockedValues = append(mockedValues, "testcp+testcl")
			// 		mockedValues = append(mockedValues, `{"resorder":[]}`)
			// 		// mockedValues = append(mockedValues, resdepmock())
			// 		mockedValues = append(mockedValues, `{"apporder":["logical-cloud"]}`)
			// 		// mockedValues = append(mockedValues, `{"appdependency":{"logical-cloud":"go"}}`) // [14]

			// 		for _, k := range [5]int{0, 1, 2, 3, 4} { // TODO set to the 7 elements once issue fixed
			// 			contextdb.Db.Get(mockedKeys[k], &val)
			// 			Expect(val).To(Equal(mockedValues[k]))
			// 		}
			// 	})

			// TODO: uncomment/finalize once lcc.LoadAppContext(ctxVal) can be mocked
			// It("termination should be successful when instantiated", func() {

			// 	// Mock gRPC InstallApp()
			// 	var gsia = &grpcSignature{}
			// 	gsia.grpcReq = nil
			// 	gsia.grpcRsp = []interface{}{&installapp.InstallAppResponse{
			// 		AppContextInstalled: true,
			// 	}, nil}
			// 	testMockGrpc(mockinstallapp.EXPECT().InstallApp, gsia.grpcReq, gsia.grpcRsp)

			// 	// Mock gRPC ReadyNotify()
			// 	var gsrn = &grpcSignature{}
			// 	gsrn.grpcReq = nil
			// 	testMockGrpcRN(mockreadynotify.EXPECT().Alert, gsrn.grpcReq)

			// 	// create these outside of BeforeEach with the purpose of being passed to the tested functions
			// 	lc := _createTestLogicalCloud("testlc", "0")
			// 	cl := _createTestClusterReference("testcp", "testcl")
			// 	quota := _createTestQuota("testquota")

			// 	// set the Logical Cloud as instantiated
			// 	lckey := dcm.LogicalCloudKey{
			// 		LogicalCloudName: "testlc",
			// 		Project:          "project",
			// 	}
			// 	mdb.Insert("orchestrator", lckey, nil, "lccontext", "4714860942153963991")

			// 	err := dcm.Terminate("project", lc, []dcm.Cluster{cl}, []dcm.Quota{quota})

			// 	etcdKeys, _ := contextdb.Db.GetAllKeys("/")
			// 	Expect(len(etcdKeys)).To(Equal(0))

			// 	Expect(err).ShouldNot(HaveOccurred())
			// })
		})
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
			})
			It("creation should succeed and return the resource created (2x - level 1 and level 0)", func() {
				originalLogicalCloud := _createTestLogicalCloud("testlogicalCloudL1", "1")
				logicalCloud, err := client.Create("project", originalLogicalCloud)
				Expect(err).ShouldNot(HaveOccurred())
				Expect(originalLogicalCloud).To(Equal(logicalCloud))
				originalLogicalCloud = _createTestLogicalCloud("testlogicalCloudL0", "0")
				logicalCloud, err = client.Create("project", originalLogicalCloud)
				Expect(err).ShouldNot(HaveOccurred())
				Expect(originalLogicalCloud).To(Equal(logicalCloud))
			})
			It("creation should succeed and return the resource created (level not specified)", func() {
				originalLogicalCloud := _createTestLogicalCloud("testlogicalCloud", "")
				logicalCloud, err := client.Create("project", originalLogicalCloud)
				Expect(err).ShouldNot(HaveOccurred())
				originalLogicalCloud.Specification.Level = "1" // created LC should default to 1
				Expect(originalLogicalCloud).To(Equal(logicalCloud))
			})
			It("get should fail and not return anything", func() {
				logicalCloud, err := client.Get("project", "testlogicalCloud")
				Expect(err).Should(HaveOccurred())
				Expect(logicalCloud).To(Equal(dcm.LogicalCloud{}))
			})
			It("create followed by get should return what was created", func() {
				logicalCloud := _createTestLogicalCloud("testlogicalCloud", "1")
				_, _ = client.Create("project", logicalCloud)
				logicalCloud, err := client.Get("project", "testlogicalCloud")
				Expect(err).ShouldNot(HaveOccurred())
				Expect(logicalCloud).To(Equal(logicalCloud))
			})
			It("create followed by get-all should return only what was created", func() {
				logicalCloud := _createTestLogicalCloud("testlogicalCloud", "1")
				_, _ = client.Create("project", logicalCloud)
				logicalClouds, err := client.GetAll("project")
				Expect(err).ShouldNot(HaveOccurred())
				Expect(len(logicalClouds)).To(Equal(1))
				Expect(logicalClouds[0]).To(Equal(logicalCloud))
			})
			It("three creates followed by get-all should return all that was created", func() {
				logicalCloud1 := _createTestLogicalCloud("testlogicalCloud1", "1")
				logicalCloud2 := _createTestLogicalCloud("testlogicalCloud2", "1")
				logicalCloud3 := _createTestLogicalCloud("testlogicalCloud3", "1")
				_, _ = client.Create("project", logicalCloud1)
				_, _ = client.Create("project", logicalCloud2)
				_, _ = client.Create("project", logicalCloud3)
				logicalClouds, err := client.GetAll("project")
				Expect(err).ShouldNot(HaveOccurred())
				Expect(len(logicalClouds)).To(Equal(3))
				Expect(logicalClouds[0]).To(Equal(logicalCloud1))
				Expect(logicalClouds[1]).To(Equal(logicalCloud2))
				Expect(logicalClouds[2]).To(Equal(logicalCloud3))
			})
			It("delete after creation should succeed and database remain empty", func() {
				logicalCloud := _createTestLogicalCloud("testlogicalCloud", "1")
				_, _ = client.Create("project", logicalCloud)
				err := client.Delete("project", "testlogicalCloud")
				Expect(err).ShouldNot(HaveOccurred())
				logicalClouds, err := client.GetAll("project")
				Expect(len(logicalClouds)).To(Equal(0))
			})
			// will uncomment after general mockdb issues resolved
			// It("delete when nothing exists should fail", func() {
			// 	err := client.Delete("project", "testlogicalCloud")
			// 	Expect(err).Should(HaveOccurred())
			// })
			It("update after creation should succeed and return updated resource", func() {
				logicalCloud := _createTestLogicalCloud("testlogicalCloud", "1")
				_, _ = client.Create("project", logicalCloud)
				logicalCloud.MetaData.UserData1 = "new user data"
				logicalCloud, err := client.Update("project", "testlogicalCloud", logicalCloud)
				Expect(err).ShouldNot(HaveOccurred())
				Expect(logicalCloud.MetaData.LogicalCloudName).To(Equal("testlogicalCloud"))
				Expect(logicalCloud.MetaData.Description).To(Equal(""))
				Expect(logicalCloud.MetaData.UserData1).To(Equal("new user data"))
				Expect(logicalCloud.MetaData.UserData2).To(Equal(""))
			})
			It("create followed by updating the name is disallowed and should fail", func() {
				logicalCloud := _createTestLogicalCloud("testlogicalCloud", "1")
				_, _ = client.Create("project", logicalCloud)
				logicalCloud.MetaData.LogicalCloudName = "updated"
				logicalCloud, err := client.Update("project", "testlogicalCloud", logicalCloud)
				Expect(err).Should(HaveOccurred())
				Expect(logicalCloud).To(Equal(dcm.LogicalCloud{}))
			})
		})
	})
})

// helper functions below

func _createTestLogicalCloud(name string, level string) dcm.LogicalCloud {
	lc := dcm.LogicalCloud{}
	lc.MetaData = dcm.MetaDataList{
		LogicalCloudName: name,
		Description:      "",
		UserData1:        "",
		UserData2:        "",
	}
	lc.Specification.NameSpace = "testns"
	lc.Specification.User = dcm.UserData{
		UserName: "lcuser",
		Type:     "certificate",
	}
	if level == "0" {
		lc.Specification.Level = level
	} else if level == "1" {
		lc.Specification.Level = level
		lc.Specification.NameSpace = "testns"
	}
	return lc
}

// TODO: merge with cluster_test.go _createTestCluster()
func _createTestClusterReference(provider string, cluster string) dcm.Cluster {
	cl := dcm.Cluster{}
	cl.MetaData = dcm.ClusterMeta{
		ClusterReference: cluster,
		Description:      "",
		UserData1:        "",
		UserData2:        "",
	}
	cl.Specification = dcm.ClusterSpec{
		ClusterProvider: provider,
		ClusterName:     cluster,
		LoadBalancerIP:  "10.10.10.10",
		Certificate:     "abcdef",
	}
	return cl
}

func _createExistingLogicalCloud(mdb *db.MockDB, level string, standard bool, privileged bool) {
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
		LogicalCloudName: "testlc",
	}
	lc := _createTestLogicalCloud("testlc", level)
	mdb.Insert("orchestrator", lkey, nil, "logicalcloud", lc)
	// create cluster reference in mocked db
	ckey := dcm.ClusterKey{
		Project:          "project",
		LogicalCloudName: "testlc",
		ClusterReference: "testcl",
	}
	cl := _createTestClusterReference("testcp", "testcl")
	mdb.Insert("orchestrator", ckey, nil, "cluster", cl)
	// create quota in mocked db
	qkey := dcm.QuotaKey{
		Project:          "project",
		LogicalCloudName: "testlc",
		QuotaName:        "testquota",
	}
	quota := _createTestQuota("testquota")
	mdb.Insert("orchestrator", qkey, nil, "quota", quota)
	upkey := dcm.UserPermissionKey{
		Project:            "project",
		LogicalCloudName:   "testlc",
		UserPermissionName: "testup",
	}
	// the standard flag indicates whther this logical cloud is, at least, a standard logical cloud.
	// for the sake of testing, if this flag isn't specified it means this logical cloud is invalid as
	// it doesn't specify its primary-namespace user permission, which is a requirement to any L1 LC.
	if standard {
		mdb.Insert("orchestrator", upkey, nil, "userpermission", _createTestUserPermission("testup", "testns"))
	}
	// the privileged flag indicates whether this logical cloud is privileged and thus,
	// for the sake of testing, a clusterwide user permission should be created
	if privileged {
		mdb.Insert("orchestrator", upkey, nil, "userpermission", _createTestUserPermission("testup", ""))
	}
	// create rsync controller in database (for grpc calls)
	ctrlkey := controller.ControllerKey{
		ControllerName: "rsync",
	}
	ctrl := controller.Controller{}
	ctrl.Metadata = types.Metadata{
		Name:        "rsync",
		Description: "",
		UserData1:   "",
		UserData2:   "",
	}
	ctrl.Spec = controller.ControllerSpec{
		Host: "localhost",
		Port: 9031,
		Type: "",
	}
	mdb.Insert("controller", ctrlkey, nil, "controllermetadata", ctrl)
}

// functions that mock strings for appcontext K8s resources

func nsmock() string {
	return `apiVersion: v1
kind: Namespace
metadata:
  name: testns
`
}

func rolemock() string {
	return `apiVersion: rbac.authorization.k8s.io/v1beta1
kind: Role
metadata:
  name: testlc-role0
  namespace: testns
rules:
- apiGroups:
  - ""
  - apps
  resources:
  - deployments
  - pods
  verbs:
  - get
  - list
`
}

func rolebindingmock() string {
	return `apiVersion: rbac.authorization.k8s.io/v1beta1
kind: RoleBinding
metadata:
  name: testlc-roleBinding0
  namespace: testns
subjects:
- kind: User
  name: lcuser
  apiGroup: ""
roleRef:
  kind: Role
  name: testlc-role0
  apiGroup: ""
`
}

func clusterrolemock() string {
	return `apiVersion: rbac.authorization.k8s.io/v1beta1
kind: ClusterRole
metadata:
  name: testlc-clusterRole1
rules:
- apiGroups:
  - ""
  - apps
  resources:
  - deployments
  - pods
  verbs:
  - get
  - list
`
}

func clusterrolebindingmock() string {
	return `apiVersion: rbac.authorization.k8s.io/v1beta1
kind: ClusterRoleBinding
metadata:
  name: testlc-clusterRoleBinding1
subjects:
- kind: User
  name: lcuser
  apiGroup: ""
roleRef:
  kind: ClusterRole
  name: testlc-clusterRole1
  apiGroup: ""
`
}

func quotamock() string {
	return `apiVersion: v1
kind: ResourceQuota
metadata:
  name: testquota
  namespace: testns
spec:
  hard:
    limits.cpu: "4"
    limits.memory: "4096"
`
}

func resordermock() string {
	return `{"resorder":["testns+Namespace","testquota+ResourceQuota","testlc-user-csr+CertificateSigningRequest","testlc-role0+Role","testlc-roleBinding0+RoleBinding"]}`
}

func resdepmock() string {
	return `{"resdependency":{"testlc-role0+Role":"wait on testlc-user-csr+CertificateSigningRequest","testlc-roleBinding0+RoleBinding":"wait on testlc-role0+Role","testlc-user-csr+CertificateSigningRequest":"wait on testquota+ResourceQuota","testns+Namespace":"go","testquota+ResourceQuota":"wait on testns+Namespace"}}`
}

func resordermockPrivileged() string {
	return `{"resorder":["testns+Namespace","testquota+ResourceQuota","testlc-user-csr+CertificateSigningRequest","testlc-role0+Role","testlc-clusterRole1+ClusterRole","testlc-roleBinding0+RoleBinding","testlc-clusterRoleBinding1+ClusterRoleBinding"]}`
}

func resdepmockPrivileged() string {
	return `{"resdependency":{"testlc-clusterRole1+ClusterRole":"wait on testlc-user-csr+CertificateSigningRequest","testlc-clusterRoleBinding1+ClusterRoleBinding":"wait on testlc-clusterRole1+ClusterRole","testlc-role0+Role":"wait on testlc-user-csr+CertificateSigningRequest","testlc-roleBinding0+RoleBinding":"wait on testlc-role0+Role","testlc-user-csr+CertificateSigningRequest":"wait on testquota+ResourceQuota","testns+Namespace":"go","testquota+ResourceQuota":"wait on testns+Namespace"}}`
}

// credit to Vinod
type rpcMsg struct {
	msg proto.Message
}
type _grpcReq proto.Message
type _grpcRsp []interface{}
type grpcSignature struct {
	grpcReq _grpcReq
	grpcRsp _grpcRsp
}
type op func(arg0, arg1 interface{}, arg2 ...interface{}) *gomock.Call

func (r *rpcMsg) Matches(msg interface{}) bool {
	m, ok := msg.(proto.Message)
	if !ok {
		return false
	}
	return proto.Equal(m, r.msg)
}

func (r *rpcMsg) String() string {
	return fmt.Sprintf("is %s", r.msg)
}
func testMockGrpc(f op, grpcReq _grpcReq, grpcResp _grpcRsp) {
	arg2 := gomock.Any()

	if grpcReq != nil {
		arg2 = &rpcMsg{msg: grpcReq}
	}

	f(gomock.Any(), arg2).Return(grpcResp...).AnyTimes()

}
func testMockGrpcRN(f op, grpcReq _grpcReq) {
	arg2 := gomock.Any()

	if grpcReq != nil {
		arg2 = &rpcMsg{msg: grpcReq}
	}

	f(gomock.Any(), arg2).Return(nil, nil).AnyTimes()

}
