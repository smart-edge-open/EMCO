package resources_test

import (
	"context"
	"fmt"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	hpaModel "github.com/open-ness/EMCO/src/hpa-plc/pkg/model"
	orchLog "github.com/open-ness/EMCO/src/orchestrator/pkg/infra/logutils"

	"github.com/open-ness/EMCO/src/hpa-plc/pkg/resources"
	"github.com/open-ness/EMCO/src/rsync/pkg/connector"
	"github.com/sirupsen/logrus"
)

func TestResources(t *testing.T) {

	fmt.Printf("\n================== TestResources .. start ==================\n")

	orchLog.SetLoglevel(logrus.InfoLevel)
	RegisterFailHandler(Fail)
	RunSpecs(t, "TestResources")

	fmt.Printf("\n================== TestResources .. end ==================\n")
}

var _ = Describe("TestResources", func() {

	It("successful allocatable-resource filter-clusters PopulateResourceInfo cpu", func() {

		// Use Kube Fake client for unit-testing
		connector.IsTestKubeClient = true

		var hpaResource hpaModel.HpaResourceRequirement
		hpaResource.MetaData.Name = "hap-resource-1"
		hpaResource.Spec.Resource.AllocatableResources.Name = "cpu"
		hpaResource.Spec.Resource.AllocatableResources.Requests = 1
		hpaResource.Spec.Resource.AllocatableResources.Limits = 1

		var rsAllocatable resources.GenericResource
		_, _, err := rsAllocatable.PopulateResourceInfo(context.TODO(), "provider1-cluster1", hpaResource)
		Expect(err).To(BeNil())
		//Expect(err).To(HaveOccurred())
	})

	It("successful allocatable-resource filter-clusters Qualified cpu", func() {

		// Use Kube Fake client for unit-testing
		connector.IsTestKubeClient = true

		var hpaResource hpaModel.HpaResourceRequirement
		hpaResource.MetaData.Name = "hap-resource-1"
		hpaResource.Spec.Resource.AllocatableResources.Name = "cpu"
		hpaResource.Spec.Resource.AllocatableResources.Requests = 1
		hpaResource.Spec.Resource.AllocatableResources.Limits = 1

		var rsAllocatable resources.GenericResource
		status := rsAllocatable.Qualified(context.TODO(), "provider1-cluster1", hpaResource)

		Expect(status).To(Equal(true))
		//Expect(err).To(HaveOccurred())
	})

	It("successful allocatable-resource filter-clusters Qualified cpu even if limits is zero", func() {

		// Use Kube Fake client for unit-testing
		connector.IsTestKubeClient = true

		var hpaResource hpaModel.HpaResourceRequirement
		hpaResource.MetaData.Name = "hap-resource-1"
		hpaResource.Spec.Resource.AllocatableResources.Name = "cpu"
		hpaResource.Spec.Resource.AllocatableResources.Requests = 1
		//hpaResource.Spec.Resource.AllocatableResources.Limits = 1

		var rsAllocatable resources.GenericResource
		status := rsAllocatable.Qualified(context.TODO(), "provider1-cluster1", hpaResource)

		Expect(status).To(Equal(true))
		//Expect(status).To(Equal(false))
	})

	It("successful allocatable-resource filter-clusters PopulateResourceInfo memory", func() {

		// Use Kube Fake client for unit-testing
		connector.IsTestKubeClient = true

		var hpaResource hpaModel.HpaResourceRequirement
		hpaResource.MetaData.Name = "hap-resource-1"
		hpaResource.Spec.Resource.AllocatableResources.Name = "memory"
		hpaResource.Spec.Resource.AllocatableResources.Requests = 1
		hpaResource.Spec.Resource.AllocatableResources.Limits = 1

		var rsAllocatable resources.GenericResource
		_, _, err := rsAllocatable.PopulateResourceInfo(context.TODO(), "provider1-cluster1", hpaResource)
		Expect(err).To(BeNil())
		//Expect(err).To(HaveOccurred())
	})

	It("successful allocatable-resource filter-clusters Qualified memory", func() {

		// Use Kube Fake client for unit-testing
		connector.IsTestKubeClient = true

		var hpaResource hpaModel.HpaResourceRequirement
		hpaResource.MetaData.Name = "hap-resource-1"
		hpaResource.Spec.Resource.AllocatableResources.Name = "memory"
		hpaResource.Spec.Resource.AllocatableResources.Requests = 1000
		hpaResource.Spec.Resource.AllocatableResources.Limits = 1000

		var rsAllocatable resources.GenericResource
		status := rsAllocatable.Qualified(context.TODO(), "provider1-cluster1", hpaResource)
		Expect(status).To(Equal(true))
	})

	It("unsuccessful allocatable-resource filter-clusters non-existing resource", func() {

		// Use Kube Fake client for unit-testing
		connector.IsTestKubeClient = true

		var hpaResource hpaModel.HpaResourceRequirement
		hpaResource.MetaData.Name = "hap-resource-1"
		hpaResource.Spec.Resource.AllocatableResources.Name = "cpu1"
		hpaResource.Spec.Resource.AllocatableResources.Requests = 1
		hpaResource.Spec.Resource.AllocatableResources.Limits = 1

		var rsAllocatable resources.GenericResource
		status := rsAllocatable.Qualified(context.TODO(), "provider1-cluster1", hpaResource)

		Expect(status).To(Equal(false))
	})

	// non-allocatable resources
	It("successful non-allocatable-resource filter-clusters PopulateResourceInfo", func() {

		// Use Kube Fake client for unit-testing
		connector.IsTestKubeClient = true

		var rsNonAllocatable resources.NFDResource
		err := rsNonAllocatable.PopulateResourceInfo(context.TODO(), "provider1-cluster1")
		Expect(err).To(BeNil())
		//Expect(err).To(HaveOccurred())
	})

	It("successful non-allocatable-resource filter-clusters Qualified", func() {

		// Use Kube Fake client for unit-testing
		connector.IsTestKubeClient = true

		var hpaResource hpaModel.HpaResourceRequirement
		hpaResource.MetaData.Name = "hap-resource-1"
		hpaResource.Spec.Resource.NonAllocatableResources.Key = "feature.node.kubernetes.io/intel_qat"
		hpaResource.Spec.Resource.NonAllocatableResources.Value = "true"

		var rsNonAllocatable resources.NFDResource
		status := rsNonAllocatable.Qualified(context.TODO(), "provider1-cluster1", hpaResource)
		Expect(status).To(Equal(true))
	})

	It("unsuccessful non-allocatable-resource filter-clusters non-existing resource label", func() {

		// Use Kube Fake client for unit-testing
		connector.IsTestKubeClient = true

		var hpaResource hpaModel.HpaResourceRequirement
		hpaResource.MetaData.Name = "hap-resource-1"
		hpaResource.Spec.Resource.NonAllocatableResources.Key = "cpuLabel"
		hpaResource.Spec.Resource.NonAllocatableResources.Value = "true"

		var rsNonAllocatable resources.NFDResource
		status := rsNonAllocatable.Qualified(context.TODO(), "provider1-cluster1", hpaResource)
		Expect(status).To(Equal(false))
	})
})
