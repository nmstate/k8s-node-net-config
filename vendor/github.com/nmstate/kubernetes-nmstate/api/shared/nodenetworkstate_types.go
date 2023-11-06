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

package shared

import (
	nmstateapiv2 "github.com/nmstate/nmstate/rust/src/go/api/v2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// NodeNetworkStateStatus is the status of the NodeNetworkState of a specific node
type NodeNetworkStateStatus struct {
	CurrentState                 nmstateapiv2.NetworkState `json:"currentState,omitempty"`
	LastSuccessfulUpdateTime     metav1.Time               `json:"lastSuccessfulUpdateTime,omitempty"`
	HostNetworkManagerVersion    string                    `json:"hostNetworkManagerVersion,omitempty"`
	HandlerNetworkManagerVersion string                    `json:"handlerNetworkManagerVersion,omitempty"`
	HandlerNmstateVersion        string                    `json:"handlerNmstateVersion,omitempty"`

	Conditions ConditionList `json:"conditions,omitempty" optional:"true"`
}

const (
	NodeNetworkStateConditionAvailable ConditionType = "Available"
	NodeNetworkStateConditionFailing   ConditionType = "Failing"
)

const (
	NodeNetworkStateConditionFailedToConfigure      ConditionReason = "FailedToConfigure"
	NodeNetworkStateConditionSuccessfullyConfigured ConditionReason = "SuccessfullyConfigured"
)
