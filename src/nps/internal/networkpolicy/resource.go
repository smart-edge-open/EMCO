// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2020 Intel Corporation

package networkpolicy

import (
	"gopkg.in/yaml.v2"
	pkgerrors "github.com/pkg/errors"
	log "github.com/open-ness/EMCO/src/orchestrator/pkg/infra/logutils"
)

const NETWORKING_APIVERSION = "networking.k8s.io/v1"
const NETWORKING_KIND = "NetworkPolicy"

type K8sNeworkPolicyResource struct {
	ApiVersion string   `yaml:"apiVersion"`
	Kind       string   `yaml:"kind"`
	Meta       Metadata `yaml:"metadata"`
	Sp         Spec     `yaml:"spec,omitempty"`
}
type Metadata struct {
	Name        string `yaml:"name"`
	Namespace   string `yaml:"namespace,omitempty"`
	Description string `yaml:"description,omitempty"`
}
type Spec struct {
	Podsel PodSelector    `yaml:"podSelector"`
	Policytypes []string  `yaml:"policyTypes,omitempty"`
	Ingress []IngressType `yaml:"ingress,omitempty"`
	Egress []EgressType   `yaml:"egress,omitempty"`
}
type PodSelector struct {
	MatchLabels map[string]string `yaml:"matchLabels"`
}
type NamespaceSelector struct {
	MatchLabels map[string]string `yaml:"matchLabels"`
}
type NetworkSelector struct {
	Cidr string     `yaml:"cidr,omitempty"`
	Except []string `yaml:"except,omitempty"`
}

type IngressType struct {
	Fm []interface{} `yaml:"from,omitempty"`
	Po []interface{} `yaml:"ports,omitempty"`
}
type EgressType struct {
	To []interface{} `yaml:"to,omitempty"`
	Po []interface{} `yaml:"ports,omitempty"`
}
type Pods struct {
	Pod PodSelector `yaml:"podSelector,omitempty"`
}
type Namespace struct {
	Names NamespaceSelector `yaml:"namespaceSelector,omitempty"`
}
type Network struct {
	Net NetworkSelector `yaml:"ipBlock,omitempty"`
}

type Ports struct {
	P []Protocol `yaml:"ports, omitempty"`
}
type Protocol struct {
	Proto string `yaml:"protocol,omitempty"`
}
type Port struct {
	P string `yaml:"namespaceSelector,omitempty"`
}

func createResource(meta Metadata, policytype []string, podselector map[string]string,
	fromselector []interface{}, inports []interface{}, toselector []interface{}, eports []interface{})([]byte, error) {

	var ing []IngressType = nil
	var eg []EgressType = nil

	for _, ptype := range policytype {
		switch ptype {
		case "Ingress":
			ing = []IngressType { {Fm: fromselector, Po: inports,}}
			break
		case "Egress":
			eg = []EgressType { {To: toselector, Po: eports,}}
			break
		default:
			log.Error("Unknow policy type", log.Fields{
				"policy type":  ptype,
			})
			var err error
			return nil, pkgerrors.Wrapf(err, "Unknow policy type: %v", ptype)
		}

	}
	var npo = K8sNeworkPolicyResource {
			ApiVersion: NETWORKING_APIVERSION,
			Kind: NETWORKING_KIND,
			Meta: meta,
			Sp: Spec {
				Podsel: PodSelector {
					MatchLabels: podselector,
				},
				Policytypes: policytype,
				Ingress: ing,
				Egress: eg,

			},
	}

	y, err := yaml.Marshal(&npo)
	//fmt.Println(string(y))
	return y, err
}
