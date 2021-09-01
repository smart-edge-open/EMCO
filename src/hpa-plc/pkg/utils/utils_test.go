package utils_test

import (
	"fmt"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/open-ness/EMCO/src/hpa-plc/pkg/utils"
	"github.com/sirupsen/logrus"

	orchLog "github.com/open-ness/EMCO/src/orchestrator/pkg/infra/logutils"
)

func TestUtils(t *testing.T) {

	fmt.Printf("\n================== TestUtils .. start ==================\n")

	orchLog.SetLoglevel(logrus.InfoLevel)
	RegisterFailHandler(Fail)
	RunSpecs(t, "Utils")

	fmt.Printf("\n================== TestUtils .. end ==================\n")
}

var _ = Describe("Utils", func() {

	It("successful IsInSlice", func() {

		list := []string{"test", "mystr"}
		status := utils.IsInSlice("test", list)
		Expect(status).To(Equal(true))
	})

	It("unsuccessful IsInSlice", func() {

		list := []string{"test", "mystr"}
		status := utils.IsInSlice("test1", list)
		Expect(status).To(Equal(false))
	})

	It("successful GetSliceIntersect", func() {
		list1 := []string{"test", "mystr"}
		list2 := []string{"test", "mystr2"}
		status := utils.GetSliceIntersect(list1, list2)
		Expect(1).To(Equal(len(status)))
	})

	It("successful GetSliceIntersect with empty slice returned", func() {
		list1 := []string{"test", "mystr"}
		list2 := []string{"test1", "mystr1"}
		status := utils.GetSliceIntersect(list1, list2)
		Expect(0).To(Equal(len(status)))
	})

})
