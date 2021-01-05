package networkpolicy_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestNetworkpolicy(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Networkpolicy Suite")
}
