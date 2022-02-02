/*
Copyright The Kubernetes NMState Authors.


Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package selectors

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
)

func unmatchingLabels(nodeSelector, labels map[string]string) map[string]string {
	unmatchingLabels := map[string]string{}
	for key, value := range nodeSelector {
		if foundValue, hasKey := labels[key]; !hasKey || foundValue != value {
			unmatchingLabels[key] = value
		}
	}
	return unmatchingLabels
}

func (s *Selectors) UnmatchedNodeLabels(nodeName string) (map[string]string, error) {
	logger := s.logger.WithValues("node", nodeName)
	node := corev1.Node{}
	err := s.client.Get(context.TODO(), types.NamespacedName{Name: nodeName}, &node)
	if err != nil {
		logger.Info("Cannot find corev1.Node")
		return map[string]string{}, err
	}

	return unmatchingLabels(s.policy.Spec.NodeSelector, node.ObjectMeta.Labels), nil
}
