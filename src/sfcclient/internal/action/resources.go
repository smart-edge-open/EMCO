// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2021 Intel Corporation

package action

import (
	"fmt"

	log "github.com/open-ness/EMCO/src/orchestrator/pkg/infra/logutils"
	v1 "k8s.io/api/apps/v1"
	batch "k8s.io/api/batch/v1"
	batchv1beta1 "k8s.io/api/batch/v1beta1"
	v1core "k8s.io/api/core/v1"
	_ "k8s.io/kubernetes/pkg/apis/apps/install"
	_ "k8s.io/kubernetes/pkg/apis/batch/install"
	_ "k8s.io/kubernetes/pkg/apis/core/install"
	_ "k8s.io/kubernetes/pkg/apis/extensions/install"
)

func updatePodTemplateLabels(pt *v1core.PodTemplateSpec, matchLabels map[string]string) {
	for k, v := range matchLabels {
		pt.Labels[k] = v
	}
}

// AddLabelsToPodTemplates adds the labels in matchLabels to the labels in the pod template
// of the resource r.
func AddLabelsToPodTemplates(r interface{}, matchLabels map[string]string) {

	switch o := r.(type) {
	case *batch.Job:
		updatePodTemplateLabels(&o.Spec.Template, matchLabels)
	case *batchv1beta1.CronJob:
		updatePodTemplateLabels(&o.Spec.JobTemplate.Spec.Template, matchLabels)
	case *v1.DaemonSet:
		updatePodTemplateLabels(&o.Spec.Template, matchLabels)
		return
	case *v1.Deployment:
		updatePodTemplateLabels(&o.Spec.Template, matchLabels)
		return
	case *v1.ReplicaSet:
		updatePodTemplateLabels(&o.Spec.Template, matchLabels)
	case *v1.StatefulSet:
		updatePodTemplateLabels(&o.Spec.Template, matchLabels)
	case *v1core.Pod:
		for k, v := range matchLabels {
			o.Labels[k] = v
		}
		return
	case *v1core.ReplicationController:
		updatePodTemplateLabels(o.Spec.Template, matchLabels)
		return
	default:
		typeStr := fmt.Sprintf("%T", o)
		log.Warn("Resource type does not have pod template", log.Fields{
			"resource type": typeStr,
		})
	}
}
