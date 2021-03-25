package module_test

import (
	"bytes"
	"testing"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	readynotifyclient "github.com/open-ness/EMCO/src/dcm/pkg/module"
	installappclient "github.com/open-ness/EMCO/src/orchestrator/pkg/grpc/installappclient"
	mockinstallpb "github.com/open-ness/EMCO/src/rsync/pkg/grpc/mock_installapp"
	mockreadynotifypb "github.com/open-ness/EMCO/src/rsync/pkg/grpc/mock_readynotify"
)

var mockinstallapp *mockinstallpb.MockInstallappClient       // for gRPC communication
var mockreadynotify *mockreadynotifypb.MockReadyNotifyClient // for gRPC communication

var buf bytes.Buffer

func TestModule(t *testing.T) {
	RegisterFailHandler(Fail)
	buf.Reset()
	ctrl := gomock.NewController(t)
	mockinstallapp = mockinstallpb.NewMockInstallappClient(ctrl)
	installappclient.Testvars.UseGrpcMock = true
	installappclient.Testvars.InstallClient = mockinstallapp
	mockreadynotify = mockreadynotifypb.NewMockReadyNotifyClient(ctrl)
	readynotifyclient.Testvars.UseGrpcMock = true
	readynotifyclient.Testvars.ReadyNotifyClient = mockreadynotify
	RunSpecs(t, "Module Suite")
	ctrl.Finish()

}
