package contextupdateserver_test

import (
	"context"
	"fmt"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	pkgerrors "github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"github.com/open-ness/EMCO/src/hpa-ac/pkg/grpc/contextupdateserver"
	"github.com/open-ness/EMCO/src/orchestrator/pkg/appcontext"
	contextpb "github.com/open-ness/EMCO/src/orchestrator/pkg/grpc/contextupdate"
	"github.com/open-ness/EMCO/src/orchestrator/pkg/infra/contextdb"
	orchLog "github.com/open-ness/EMCO/src/orchestrator/pkg/infra/logutils"
)

func TestHpaContextupdateserver(t *testing.T) {

	fmt.Printf("\n================== TestHpaContextupdateserver .. start ==================\n")

	orchLog.SetLoglevel(logrus.InfoLevel)
	RegisterFailHandler(Fail)
	RunSpecs(t, "HpaContextupdateserver")

	fmt.Printf("\n================== TestHpaContextupdateserver .. end ==================\n")
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

var _ = Describe("HpaContextupdateserver", func() {

	var (
		project    string = "p"
		compApp    string = "ca"
		version    string = "v1"
		dig        string = "dig"
		logicCloud string = "default"
		release    string = "r1"
		namespace  string = "n1"
	)

	It("unsuccessful update context with nil request", func() {
		hpaContextUpdateServer := contextupdateserver.NewContextupdateServer()
		resp, _ := hpaContextUpdateServer.UpdateAppContext(context.TODO(), nil)
		Expect(resp.AppContextUpdated).To(Equal(false))
	})

	It("unsuccessful update context with empty appcontext", func() {
		hpaContextUpdateServer := contextupdateserver.NewContextupdateServer()

		var updateReq contextpb.ContextUpdateRequest
		updateReq.AppContext = ""

		resp, _ := hpaContextUpdateServer.UpdateAppContext(context.TODO(), &updateReq)
		Expect(resp.AppContextUpdated).To(Equal(false))
	})

	It("unsuccessful update with invalid context", func() {

		// etcd mockdb
		edb := new(contextdb.MockConDb)
		edb.Err = nil
		contextdb.Db = edb

		// Initialize etcd with default values
		var err error
		_, err = makeAppContextForCompositeApp(project, compApp, version, release, dig, namespace, logicCloud)
		Expect(err).To(BeNil())

		hpaContextUpdateServer := contextupdateserver.NewContextupdateServer()

		var updateReq contextpb.ContextUpdateRequest
		updateReq.AppContext = "1234"
		updateReq.IntentName = "testIntent"

		resp, _ := hpaContextUpdateServer.UpdateAppContext(context.TODO(), &updateReq)
		Expect(resp.AppContextUpdated).To(Equal(false))
	})

})
