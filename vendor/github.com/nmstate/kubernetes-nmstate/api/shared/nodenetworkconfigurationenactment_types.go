package shared

import (
	"fmt"

	"k8s.io/apimachinery/pkg/types"
)

// NodeNetworkConfigurationEnactmentStatus defines the observed state of NodeNetworkConfigurationEnactment
type NodeNetworkConfigurationEnactmentStatus struct {
	// +kubebuilder:validation:XPreserveUnknownFields
	// The desired state rendered for the enactment's node using
	// the policy desiredState as template
	DesiredState State `json:"desiredState,omitempty"`

	// The generation from policy needed to check if an enactment
	// condition status belongs to the same policy version
	PolicyGeneration int64         `json:"policyGeneration,omitempty"`
	Conditions       ConditionList `json:"conditions,omitempty"`
}

const (
	EnactmentPolicyLabel                                                = "nmstate.io/policy"
	EnactmentNodeLabel                                                  = "nmstate.io/node"
	NodeNetworkConfigurationEnactmentConditionAvailable   ConditionType = "Available"
	NodeNetworkConfigurationEnactmentConditionFailing     ConditionType = "Failing"
	NodeNetworkConfigurationEnactmentConditionPending     ConditionType = "Pending"
	NodeNetworkConfigurationEnactmentConditionProgressing ConditionType = "Progressing"
	NodeNetworkConfigurationEnactmentConditionAborted     ConditionType = "Aborted"
)

var NodeNetworkConfigurationEnactmentConditionTypes = [...]ConditionType{
	NodeNetworkConfigurationEnactmentConditionAvailable,
	NodeNetworkConfigurationEnactmentConditionFailing,
	NodeNetworkConfigurationEnactmentConditionProgressing,
	NodeNetworkConfigurationEnactmentConditionPending,
	NodeNetworkConfigurationEnactmentConditionAborted,
}

const (
	NodeNetworkConfigurationEnactmentConditionFailedToConfigure          ConditionReason = "FailedToConfigure"
	NodeNetworkConfigurationEnactmentConditionSuccessfullyConfigured     ConditionReason = "SuccessfullyConfigured"
	NodeNetworkConfigurationEnactmentConditionMaxUnavailableLimitReached ConditionReason = "MaxUnavailableLimitReached"
	NodeNetworkConfigurationEnactmentConditionConfigurationProgressing   ConditionReason = "ConfigurationProgressing"
	NodeNetworkConfigurationEnactmentConditionConfigurationAborted       ConditionReason = "ConfigurationAborted"
)

func EnactmentKey(node string, policy string) types.NamespacedName {
	return types.NamespacedName{Name: fmt.Sprintf("%s.%s", node, policy)}
}
