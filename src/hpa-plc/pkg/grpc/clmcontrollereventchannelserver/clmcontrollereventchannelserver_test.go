package clmcontrollereventchannel_test

import (
	"context"
	"fmt"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/sirupsen/logrus"

	"github.com/open-ness/EMCO/src/rsync/pkg/connector"

	clmcontrollerpb "github.com/open-ness/EMCO/src/clm/pkg/grpc/controller-eventchannel"
	clmControllerserver "github.com/open-ness/EMCO/src/hpa-plc/pkg/grpc/clmcontrollereventchannelserver"
	"github.com/open-ness/EMCO/src/orchestrator/pkg/infra/db"
	orchLog "github.com/open-ness/EMCO/src/orchestrator/pkg/infra/logutils"
	orchMod "github.com/open-ness/EMCO/src/orchestrator/pkg/module"
)

func TestClmcontrollereventchannelserver(t *testing.T) {

	fmt.Printf("\n================== TestClmcontrollereventchannelserver .. start ==================\n")

	orchLog.SetLoglevel(logrus.InfoLevel)
	RegisterFailHandler(Fail)
	RunSpecs(t, "Clmcontrollereventchannelserver")

	fmt.Printf("\n================== TestClmcontrollereventchannelserver .. end ==================\n")
}

var _ = Describe("Clmcontrollereventchannelserver", func() {

	var (
		project string = "p"
		compApp string = "ca"
		version string = "v1"
		dig     string = "dig"
		app1    string = "client"
	)

	BeforeEach(func() {
		// mongo mockdb
		mdb := &db.MockDB{
			Items: []map[string]map[string][]byte{
				{
					orchMod.ProjectKey{ProjectName: project}.String(): {
						"projectmetadata": []byte(
							"{\"project-name\":\"" + project + "\"," +
								"\"description\":\"Test project for unit testing\"}"),
					},
					orchMod.CompositeAppKey{CompositeAppName: compApp,
						Version: version, Project: project}.String(): {
						"compositeappmetadata": []byte(
							"{\"metadata\":{" +
								"\"name\":\"" + compApp + "\"," +
								"\"description\":\"description\"," +
								"\"userData1\":\"user data\"," +
								"\"userData2\":\"user data\"" +
								"}," +
								"\"spec\":{" +
								"\"version\":\"version of the composite app\"}}"),
					},
					orchMod.AppKey{App: "", Project: project, CompositeApp: compApp, CompositeAppVersion: version}.String(): {
						"appmetadata": []byte(
							"{" +
								"\"metadata\": {" +
								"\"name\":\"" + app1 + "\"," +
								"\"description\": \"Test App for unit testing\"," +
								"\"userData1\": \"userData1\"," +
								"\"userData2\": \"userData2\"}" +
								"}"),
						"appcontent": []byte(
							"{" +
								"\"FileContent\": \"sample file content\"" +
								"}"),
					},
					orchMod.DeploymentIntentGroupKey{
						Name:         dig,
						Project:      project,
						CompositeApp: compApp,
						Version:      version,
					}.String(): {
						"deploymentintentgroupmetadata": []byte(
							"{" +
								"\"metadata\":{" +
								"\"name\":\"" + dig + "\"," +
								"\"description\":\"DescriptionTestDeploymentIntentGroup\"," +
								"\"userData1\": \"userData1\"," +
								"\"userData2\": \"userData2\"}" +
								"}"),
					},
				},
			},
		}
		mdb.Err = nil
		db.DBconn = mdb
	})

	It("unsuccessful Publish", func() {
		clmControllerServer := clmControllerserver.NewControllerEventchannelServer()
		resp, _ := clmControllerServer.Publish(context.TODO(), nil)
		Expect(resp.Status).To(Equal(false))
	})

	It("unsuccessful Publish request both provideName & clusterName empty", func() {
		var req clmcontrollerpb.ClmControllerEventRequest
		req.ProviderName = ""
		req.ClusterName = ""

		clmControllerServer := clmControllerserver.NewControllerEventchannelServer()
		resp, _ := clmControllerServer.Publish(context.TODO(), &req)
		Expect(resp.Status).To(Equal(false))
	})

	It("Successful Publish request unknown event", func() {
		var req clmcontrollerpb.ClmControllerEventRequest
		req.ProviderName = ""
		req.ClusterName = "cluster"
		req.Event = 8

		clmControllerServer := clmControllerserver.NewControllerEventchannelServer()
		resp, _ := clmControllerServer.Publish(context.TODO(), &req)
		Expect(resp.Status).To(Equal(true))
	})

	It("Successful Publish request create event", func() {
		var req clmcontrollerpb.ClmControllerEventRequest
		req.ProviderName = ""
		req.ClusterName = "cluster"
		req.Event = clmcontrollerpb.ClmControllerEventType_CLUSTER_CREATED

		// Use Kube Fake client for unit-testing
		connector.IsTestKubeClient = true

		clmControllerServer := clmControllerserver.NewControllerEventchannelServer()
		resp, _ := clmControllerServer.Publish(context.TODO(), &req)
		Expect(resp.Status).To(Equal(true))
	})

})
