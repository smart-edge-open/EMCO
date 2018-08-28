// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2020 Intel Corporation

package networkpolicy

import (
	"encoding/json"
	"strings"

	"github.com/open-ness/EMCO/src/clm/pkg/cluster"
	"github.com/open-ness/EMCO/src/dtc/pkg/module"
	"github.com/open-ness/EMCO/src/orchestrator/pkg/appcontext"
	log "github.com/open-ness/EMCO/src/orchestrator/pkg/infra/logutils"
	pkgerrors "github.com/pkg/errors"
)

// Action applies the supplied intent against the given AppContext ID
func UpdateAppContext(intentName, appContextId string) error {
	var ac appcontext.AppContext
	_, err := ac.LoadAppContext(appContextId)
	if err != nil {
		log.Error("Error loading AppContext", log.Fields{
			"error": err,
		})
		return pkgerrors.Wrapf(err, "Error loading AppContext with Id: %v", appContextId)
	}

	caMeta, err := ac.GetCompositeAppMeta()
	if err != nil {
		log.Error("Error getting metadata from AppContext", log.Fields{
			"error": err,
		})
		return pkgerrors.Wrapf(err, "Error getting metadata from AppContext with Id: %v", appContextId)
	}

	project := caMeta.Project
	compositeapp := caMeta.CompositeApp
	compositeappversion := caMeta.Version
	deployIntentGroup := caMeta.DeploymentIntentGroup

	// Get all server inbound intents
	iss, err := module.NewServerInboundIntentClient().GetServerInboundIntents(project, compositeapp, compositeappversion, deployIntentGroup, intentName)
	if err != nil {
		log.Error("Error getting server inbound intents", log.Fields{
			"error": err,
		})
		return pkgerrors.Wrapf(err, "Error getting server inbound intents %v for %v/%v%v/%v not found", intentName, project, compositeapp, deployIntentGroup, compositeappversion)
	}

	for _, is := range iss {
		policytypes := []string{"Ingress"}
		meta := Metadata{
			Name:        intentName + "-" + is.Metadata.Name,
			Namespace:   "",
			Description: "",
		}

		ps := make(map[string]string)
		ps["app"] = is.Spec.AppName

		inports := []interface{}{}
		proto := Protocol{Proto: is.Spec.Protocol}
		inports = append(inports, proto)
		port := make(map[string]int)
		port["port"] = is.Spec.Port
		inports = append(inports, port)

		ics, err := module.NewClientsInboundIntentClient().GetClientsInboundIntents(project,
			compositeapp,
			compositeappversion,
			deployIntentGroup,
			intentName,
			is.Metadata.Name)
		if err != nil {
			log.Error("Error getting clients inbound intents", log.Fields{
				"error": err,
			})
			return pkgerrors.Wrapf(err,
				"Error getting clients inbound intents %v under server inbound intent %v for %v/%v%v/%v not found",
				is.Metadata.Name, intentName, project, compositeapp, compositeappversion, deployIntentGroup)
		}

		flist := []interface{}{}
		for _, ic := range ics {
			if ic.Spec.AppName != "" {
				flist = append(flist, Pods{Pod: PodSelector{MatchLabels: map[string]string{"app": ic.Spec.AppName}}})
			}
			if is.Spec.ExternalSupport {
				for _, cidr := range ic.Spec.IpRange {
					flist = append(flist, Network{Net: NetworkSelector{Cidr: cidr, Except: nil}})
				}
			}

		}
		if !is.Spec.ExternalSupport {
			fromnetwork := Network{Net: NetworkSelector{Cidr: "0.0.0.0/0", Except: nil}}
			flist = append(flist, fromnetwork)

		}

		r, err := createResource(meta, policytypes, ps, flist, inports, nil, nil)
		if err != nil {
			log.Error("Error creating resource from App Context", log.Fields{
				"error":    err,
				"app name": is.Spec.AppName,
			})
			return pkgerrors.Wrapf(err,
				"Error creating resource from App Context for app %v", is.Spec.AppName)

		}

		// create resource using is and ics
		// Get all clusters for the current App from the AppContext
		clusters, err := ac.GetClusterNames(is.Spec.AppName)
		if err != nil {
			log.Error("Error retrieving clusters from App Context", log.Fields{
				"error":    err,
				"app name": is.Spec.AppName,
			})
			return pkgerrors.Wrapf(err,
				"Error retrieving clusters from App Context for app %v", is.Spec.AppName)
		}

		for _, c := range clusters {

			// check if the cluster supports networkpolicy
			client := cluster.NewClusterClient()
			parts := strings.Split(c, "+")
			if len(parts) != 2 {
				log.Error("Not a valid cluster name", log.Fields{
					"cluster name": c,
				})
				return pkgerrors.New("Not a valid cluster name")
			}
			cl, err := client.GetClusterLabel(parts[0], parts[1], "networkpolicy-supported")
			if err != nil || cl.LabelName != "networkpolicy-supported" {
				continue
			}

			//put the resource in all the clusters
			ch, err := ac.GetClusterHandle(is.Spec.AppName, c)
			if err != nil {
				log.Error("Error getting clusters handle App Context", log.Fields{
					"error":        err,
					"app name":     is.Spec.AppName,
					"cluster name": c,
				})
				return pkgerrors.Wrapf(err,
					"Error getting clusters from App Context for app %v and cluster %v", is.Spec.AppName, c)
			}
			// Add resource to the cluster
			resname := intentName + "-" + is.Metadata.Name
			_, err = ac.AddResource(ch, resname, string(r))
			if err != nil {
				log.Error("Error adding Resource to AppContext", log.Fields{
					"error":        err,
					"app name":     is.Spec.AppName,
					"cluster name": c,
				})
				return pkgerrors.Wrap(err, "Error adding Resource to AppContext")
			}
			resorder, err := ac.GetResourceInstruction(is.Spec.AppName, c, "order")
			if err != nil {
				log.Error("Error getting Resource order", log.Fields{
					"error":        err,
					"app name":     is.Spec.AppName,
					"cluster name": c,
				})
				return pkgerrors.Wrap(err, "Error getting Resource order")
			}
			aov := make(map[string][]string)
			json.Unmarshal([]byte(resorder.(string)), &aov)
			aov["resorder"] = append(aov["resorder"], resname)
			jresord, _ := json.Marshal(aov)

			_, err = ac.AddInstruction(ch, "resource", "order", string(jresord))
			if err != nil {
				log.Error("Error updating Resource order", log.Fields{
					"error":        err,
					"app name":     is.Spec.AppName,
					"cluster name": c,
				})
				return pkgerrors.Wrap(err, "Error updating Resource order")
			}
		}
	}

	return nil
}
