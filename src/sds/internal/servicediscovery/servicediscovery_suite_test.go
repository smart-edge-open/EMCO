package servicediscovery_test

import (
	"bytes"
	"testing"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	installappclient "github.com/open-ness/EMCO/src/orchestrator/pkg/grpc/installappclient"
	mockinstallpb "github.com/open-ness/EMCO/src/rsync/pkg/grpc/mock_installapp"
)

var mockApplication *mockinstallpb.MockInstallappClient
var buf bytes.Buffer

func TestServiceDiscovery(t *testing.T) {
	RegisterFailHandler(Fail)
	buf.Reset()
	ctrl := gomock.NewController(t)
	mockApplication = mockinstallpb.NewMockInstallappClient(ctrl)
	installappclient.Testvars.UseGrpcMock = true
	installappclient.Testvars.InstallClient = mockApplication
	RunSpecs(t, "Servicediscovery Suite")
	ctrl.Finish()
}
