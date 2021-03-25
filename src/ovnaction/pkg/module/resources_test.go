package module_test

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strings"

	jyaml "github.com/ghodss/yaml"
	nettypes "github.com/k8snetworkplumbingwg/network-attachment-definition-client/pkg/apis/k8s.cni.cncf.io/v1"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	v1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/scheme"

	"github.com/open-ness/EMCO/src/ovnaction/pkg/module"
)

//k8s.v1.cni.cncf.io/networks: ns1/net1@if1, ns2/net2@if2, net3, net4@if4
var deployment = `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: mock-deployment
spec:
  template:
    metadata:
      annotations:
        abc: abc123
        k8s.v1.cni.cncf.io/networks: ns1/net1@if1, ns2/net2@if2, net3, net4@if4
        junk: blahblahblah
      labels:
        app: sise
    spec:
      containers:
      - name: sise
        image: mhausenblas/simpleservice:0.5.0
`

var temp, _ = ioutil.ReadFile("test/elastic-ds.yaml")
var daemonset = string(temp)

var temp2, _ = ioutil.ReadFile("test/frontend-rs.yaml")
var replicaset = string(temp2)

var temp3, _ = ioutil.ReadFile("test/hello-cj.yaml")
var cronjob = string(temp3)

var temp4, _ = ioutil.ReadFile("test/pod.yaml")
var pod = string(temp4)

var networkAnnotation []nettypes.NetworkSelectionElement

func convert(r string, spec_array []module.WorkloadIfIntentSpec) string {
	obj1, err := runtime.Decode(scheme.Codecs.UniversalDeserializer(), []byte(r))
	module.AddNfnAnnotation(obj1, spec_array)
	Expect(err).To(BeNil())
	obj2, err := json.Marshal(obj1)
	Expect(err).To(BeNil())
	obj3, err := jyaml.JSONToYAML(obj2)
	Expect(err).To(BeNil())
	temp := string(obj3)

	check := strings.Contains(temp, "name1")
	Expect(check).To(Equal(true))
	check = strings.Contains(temp, "k8s.plugin.opnfv.org/nfn-network: '{\"type\":\"ovn4nfv\",\"interface\":[{\"interface\":\"name1\",\"name\":\"network_name\",\"defaultGateway\":\"\"}]}'")
	check1 := strings.Contains(temp, "k8s.plugin.opnfv.org/nfn-network: '{\"type\":\"ovn4nfv\",\"interface\":[{\"interface\":\"eth3\",\"name\":\"network3\",\"defaultGateway\":\"\"},{\"interface\":\"name1\",\"name\":\"network_name\",\"defaultGateway\":\"\"}]}'")
	fmt.Printf("TEMP = %s\n", temp)

	if check1 {
		Expect(check1).To(Equal(true))
		return temp
	}
	Expect(check).To(Equal(true))
	return temp
}

var _ = Describe("Resources", func() {
	var (
		WLFI module.WorkloadIfIntent

		nfn module.NfnAnnotation

		spec_array  []module.WorkloadIfIntentSpec
		spec_array2 []module.WorkloadIfIntentSpec
	)

	BeforeEach(func() {
		WLFI = module.WorkloadIfIntent{
			Metadata: module.Metadata{
				Name:        "theThirdName",
				Description: "work load if intent",
				UserData1:   "user data1",
				UserData2:   "user data2",
			},
		}
		WLFI.Spec.IfName = "name1"
		WLFI.Spec.NetworkName = "network_name"
		spec_array = append(spec_array, WLFI.Spec)
		spec_array = spec_array[:0]
		spec_array = append(spec_array, WLFI.Spec)

		WLFI.Spec.IfName = "name2"
		WLFI.Spec.NetworkName = "network_name2"
		spec_array2 = append(spec_array2, WLFI.Spec)
		spec_array2 = spec_array2[:0]
		spec_array2 = append(spec_array2, WLFI.Spec)

		WLFI.Spec.IfName = "name1"
		WLFI.Spec.NetworkName = "network_name"

		network_selection_obj := nettypes.NetworkSelectionElement{
			Name:      "selection_name",
			Namespace: "namespace_name",
		}
		networkAnnotation = append(networkAnnotation, network_selection_obj)

		nfn.CniType = "ovn4nfv"
		nfn.Interface = spec_array
	})

	It("should add NfnAnnotation to the deployment", func() {
		convert(deployment, spec_array)
	})
	It("should add NfnAnnotation to the daemonset", func() {
		convert(daemonset, spec_array)
	})
	It("should add NfnAnnotation to the replicaset", func() {
		convert(replicaset, spec_array)
	})
	It("should add NfnAnnotation to the replicaset", func() {
		convert(cronjob, spec_array)
	})
	It("should add NfnAnnotation to the pod", func() {
		obj1, err := runtime.Decode(scheme.Codecs.UniversalDeserializer(), []byte(pod))
		module.AddNfnAnnotation(obj1, spec_array)
		Expect(err).To(BeNil())
		obj2, err := json.Marshal(obj1)
		Expect(err).To(BeNil())
		obj3, err := jyaml.JSONToYAML(obj2)
		Expect(err).To(BeNil())
		temp := string(obj3)

		check := strings.Contains(temp, "name1")
		Expect(check).To(Equal(true))
	})
	It("should add network annotation to daemonset", func() {
		obj1, _ := runtime.Decode(scheme.Codecs.UniversalDeserializer(), []byte(daemonset))
		module.AddNetworkAnnotation(obj1, networkAnnotation[0])
		switch o := obj1.(type) {
		case *v1.DaemonSet:
			_, err := module.ParsePodTemplateNetworkAnnotation(&o.Spec.Template)
			Expect(err).To(BeNil())
		}
	})
	It("should add network annotation to replicaset", func() {
		obj1, _ := runtime.Decode(scheme.Codecs.UniversalDeserializer(), []byte(replicaset))
		module.AddNetworkAnnotation(obj1, networkAnnotation[0])

		switch o := obj1.(type) {
		case *v1.ReplicaSet:
			_, err := module.ParsePodTemplateNetworkAnnotation(&o.Spec.Template)
			Expect(err).To(BeNil())
		}
	})
	It("should add networkAnnotation then parse pod annotation", func() {
		obj1, _ := runtime.Decode(scheme.Codecs.UniversalDeserializer(), []byte(deployment))
		module.AddNfnAnnotation(obj1, spec_array)
		module.AddNetworkAnnotation(obj1, networkAnnotation[0])
		switch o := obj1.(type) {
		case *v1.Deployment:
			_, err := module.ParsePodTemplateNetworkAnnotation(&o.Spec.Template)
			Expect(err).To(BeNil())
		}
	})
	It("should get pod template nfn annotation", func() {
		obj1, _ := runtime.Decode(scheme.Codecs.UniversalDeserializer(), []byte(deployment))
		module.AddNfnAnnotation(obj1, spec_array)
		module.AddNetworkAnnotation(obj1, networkAnnotation[0])

		switch o := obj1.(type) {
		case *v1.Deployment:
			_, err := module.ParsePodTemplateNetworkAnnotation(&o.Spec.Template)
			Expect(err).To(BeNil())
			nfnAnnotation := module.GetPodTemplateNfnAnnotation(&o.Spec.Template)
			nfn.CniType = "ovn4nfv"
			nfn.Interface = spec_array
			Expect(nfnAnnotation).To(Equal(nfn))
		}
	})
	It("should add networkAnnotation for pod", func() {
		obj1, _ := runtime.Decode(scheme.Codecs.UniversalDeserializer(), []byte(pod))
		module.AddNfnAnnotation(obj1, spec_array)
		module.AddNetworkAnnotation(obj1, networkAnnotation[0])
	})
})
