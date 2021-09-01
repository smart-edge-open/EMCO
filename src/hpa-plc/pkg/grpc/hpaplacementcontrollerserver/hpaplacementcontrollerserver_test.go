package placementcontroller_test

import (
	"fmt"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/sirupsen/logrus"
	"golang.org/x/net/context"

	placementcontrollerserver "github.com/open-ness/EMCO/src/hpa-plc/pkg/grpc/hpaplacementcontrollerserver"
	"github.com/open-ness/EMCO/src/orchestrator/pkg/appcontext"
	placementcontrollerpb "github.com/open-ness/EMCO/src/orchestrator/pkg/grpc/placementcontroller"
	"github.com/open-ness/EMCO/src/orchestrator/pkg/infra/contextdb"
	orchLog "github.com/open-ness/EMCO/src/orchestrator/pkg/infra/logutils"
	pkgerrors "github.com/pkg/errors"
)

func TestHpaPlacementControllerServer(t *testing.T) {

	fmt.Printf("\n================== TestHpaPlacementControllerServer .. start ==================\n")

	orchLog.SetLoglevel(logrus.InfoLevel)
	RegisterFailHandler(Fail)
	RunSpecs(t, "HpaPlacementControllerServer")

	fmt.Printf("\n================== TestHpaPlacementControllerServer .. end ==================\n")
}

type contextForCompositeApp struct {
	context            appcontext.AppContext
	ctxval             interface{}
	compositeAppHandle interface{}
}

func makeAppContextForCompositeApp(p, ca, v, rName, dig string, namespace string, level string) (contextForCompositeApp, error) {
	context := appcontext.AppContext{}
	ctxval, err := context.InitAppContext()
	if err != nil {
		return contextForCompositeApp{}, pkgerrors.Wrap(err, "Error creating AppContext CompositeApp")
	}
	compositeHandle, err := context.CreateCompositeApp()
	if err != nil {
		return contextForCompositeApp{}, pkgerrors.Wrap(err, "Error creating CompositeApp handle")
	}
	compMetadata := appcontext.CompositeAppMeta{Project: p, CompositeApp: ca, Version: v, Release: rName, DeploymentIntentGroup: dig, Namespace: namespace, Level: level}
	err = context.AddCompositeAppMeta(compMetadata)
	if err != nil {
		return contextForCompositeApp{}, pkgerrors.Wrap(err, "Error Adding CompositeAppMeta")
	}

	cca := contextForCompositeApp{context: context, ctxval: ctxval, compositeAppHandle: compositeHandle}

	return cca, nil
}

var _ = Describe("HpaPlacementControllerServer", func() {

	var (
		edb *contextdb.MockConDb

		project    string = "p"
		compApp    string = "ca"
		version    string = "v1"
		dig        string = "dig"
		logicCloud string = "default"
		release    string = "r1"
		namespace  string = "n1"
	)

	BeforeEach(func() {

		// etcd mockdb
		edb = new(contextdb.MockConDb)
		edb.Err = nil
		contextdb.Db = edb

		// Initialize etcd with default values
		var err error
		_, err = makeAppContextForCompositeApp(project, compApp, version, release, dig, namespace, logicCloud)
		Expect(err).To(BeNil())

	})

	It("unsuccessful FilterClusters", func() {
		hpaPlacementcontrollerServer := placementcontrollerserver.NewHpaPlacementControllerServer()
		resp, _ := hpaPlacementcontrollerServer.FilterClusters(context.TODO(), nil)
		Expect(resp.Status).To(Equal(false))
	})

	It("unsuccessful FilterClusters request AppContext is empty", func() {
		var req placementcontrollerpb.ResourceRequest
		req.AppContext = ""

		hpaPlacementcontrollerServer := placementcontrollerserver.NewHpaPlacementControllerServer()
		resp, _ := hpaPlacementcontrollerServer.FilterClusters(context.TODO(), &req)
		Expect(resp.Status).To(Equal(false))
	})

	It("unsuccessful FilterClusters request AppContext is invalid", func() {
		var req placementcontrollerpb.ResourceRequest
		req.AppContext = "1234"

		hpaPlacementcontrollerServer := placementcontrollerserver.NewHpaPlacementControllerServer()
		resp, _ := hpaPlacementcontrollerServer.FilterClusters(context.TODO(), &req)
		Expect(resp.Status).To(Equal(false))
	})
})
