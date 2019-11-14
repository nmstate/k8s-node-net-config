// +build !ignore_autogenerated

// Code generated by operator-sdk. DO NOT EDIT.

package v1alpha1

import (
	runtime "k8s.io/apimachinery/pkg/runtime"
)

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Condition) DeepCopyInto(out *Condition) {
	*out = *in
	in.LastHeartbeatTime.DeepCopyInto(&out.LastHeartbeatTime)
	in.LastTransitionTime.DeepCopyInto(&out.LastTransitionTime)
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Condition.
func (in *Condition) DeepCopy() *Condition {
	if in == nil {
		return nil
	}
	out := new(Condition)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in ConditionList) DeepCopyInto(out *ConditionList) {
	{
		in := &in
		*out = make(ConditionList, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
		return
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ConditionList.
func (in ConditionList) DeepCopy() ConditionList {
	if in == nil {
		return nil
	}
	out := new(ConditionList)
	in.DeepCopyInto(out)
	return *out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Enactment) DeepCopyInto(out *Enactment) {
	*out = *in
	if in.Conditions != nil {
		in, out := &in.Conditions, &out.Conditions
		*out = make(ConditionList, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Enactment.
func (in *Enactment) DeepCopy() *Enactment {
	if in == nil {
		return nil
	}
	out := new(Enactment)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in EnactmentList) DeepCopyInto(out *EnactmentList) {
	{
		in := &in
		*out = make(EnactmentList, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
		return
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new EnactmentList.
func (in EnactmentList) DeepCopy() EnactmentList {
	if in == nil {
		return nil
	}
	out := new(EnactmentList)
	in.DeepCopyInto(out)
	return *out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *NodeNetworkConfigurationEnactment) DeepCopyInto(out *NodeNetworkConfigurationEnactment) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	out.Spec = in.Spec
	out.Status = in.Status
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new NodeNetworkConfigurationEnactment.
func (in *NodeNetworkConfigurationEnactment) DeepCopy() *NodeNetworkConfigurationEnactment {
	if in == nil {
		return nil
	}
	out := new(NodeNetworkConfigurationEnactment)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *NodeNetworkConfigurationEnactment) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *NodeNetworkConfigurationEnactmentList) DeepCopyInto(out *NodeNetworkConfigurationEnactmentList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]NodeNetworkConfigurationEnactment, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new NodeNetworkConfigurationEnactmentList.
func (in *NodeNetworkConfigurationEnactmentList) DeepCopy() *NodeNetworkConfigurationEnactmentList {
	if in == nil {
		return nil
	}
	out := new(NodeNetworkConfigurationEnactmentList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *NodeNetworkConfigurationEnactmentList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *NodeNetworkConfigurationEnactmentSpec) DeepCopyInto(out *NodeNetworkConfigurationEnactmentSpec) {
	*out = *in
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new NodeNetworkConfigurationEnactmentSpec.
func (in *NodeNetworkConfigurationEnactmentSpec) DeepCopy() *NodeNetworkConfigurationEnactmentSpec {
	if in == nil {
		return nil
	}
	out := new(NodeNetworkConfigurationEnactmentSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *NodeNetworkConfigurationEnactmentStatus) DeepCopyInto(out *NodeNetworkConfigurationEnactmentStatus) {
	*out = *in
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new NodeNetworkConfigurationEnactmentStatus.
func (in *NodeNetworkConfigurationEnactmentStatus) DeepCopy() *NodeNetworkConfigurationEnactmentStatus {
	if in == nil {
		return nil
	}
	out := new(NodeNetworkConfigurationEnactmentStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *NodeNetworkConfigurationPolicy) DeepCopyInto(out *NodeNetworkConfigurationPolicy) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	in.Status.DeepCopyInto(&out.Status)
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new NodeNetworkConfigurationPolicy.
func (in *NodeNetworkConfigurationPolicy) DeepCopy() *NodeNetworkConfigurationPolicy {
	if in == nil {
		return nil
	}
	out := new(NodeNetworkConfigurationPolicy)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *NodeNetworkConfigurationPolicy) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *NodeNetworkConfigurationPolicyList) DeepCopyInto(out *NodeNetworkConfigurationPolicyList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]NodeNetworkConfigurationPolicy, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new NodeNetworkConfigurationPolicyList.
func (in *NodeNetworkConfigurationPolicyList) DeepCopy() *NodeNetworkConfigurationPolicyList {
	if in == nil {
		return nil
	}
	out := new(NodeNetworkConfigurationPolicyList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *NodeNetworkConfigurationPolicyList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *NodeNetworkConfigurationPolicySpec) DeepCopyInto(out *NodeNetworkConfigurationPolicySpec) {
	*out = *in
	if in.NodeSelector != nil {
		in, out := &in.NodeSelector, &out.NodeSelector
		*out = make(map[string]string, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
	if in.DesiredState != nil {
		in, out := &in.DesiredState, &out.DesiredState
		*out = make(State, len(*in))
		copy(*out, *in)
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new NodeNetworkConfigurationPolicySpec.
func (in *NodeNetworkConfigurationPolicySpec) DeepCopy() *NodeNetworkConfigurationPolicySpec {
	if in == nil {
		return nil
	}
	out := new(NodeNetworkConfigurationPolicySpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *NodeNetworkConfigurationPolicyStatus) DeepCopyInto(out *NodeNetworkConfigurationPolicyStatus) {
	*out = *in
	if in.Enactments != nil {
		in, out := &in.Enactments, &out.Enactments
		*out = make(EnactmentList, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new NodeNetworkConfigurationPolicyStatus.
func (in *NodeNetworkConfigurationPolicyStatus) DeepCopy() *NodeNetworkConfigurationPolicyStatus {
	if in == nil {
		return nil
	}
	out := new(NodeNetworkConfigurationPolicyStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *NodeNetworkState) DeepCopyInto(out *NodeNetworkState) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Status.DeepCopyInto(&out.Status)
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new NodeNetworkState.
func (in *NodeNetworkState) DeepCopy() *NodeNetworkState {
	if in == nil {
		return nil
	}
	out := new(NodeNetworkState)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *NodeNetworkState) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *NodeNetworkStateList) DeepCopyInto(out *NodeNetworkStateList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]NodeNetworkState, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new NodeNetworkStateList.
func (in *NodeNetworkStateList) DeepCopy() *NodeNetworkStateList {
	if in == nil {
		return nil
	}
	out := new(NodeNetworkStateList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *NodeNetworkStateList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *NodeNetworkStateStatus) DeepCopyInto(out *NodeNetworkStateStatus) {
	*out = *in
	if in.CurrentState != nil {
		in, out := &in.CurrentState, &out.CurrentState
		*out = make(State, len(*in))
		copy(*out, *in)
	}
	if in.Conditions != nil {
		in, out := &in.Conditions, &out.Conditions
		*out = make(ConditionList, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new NodeNetworkStateStatus.
func (in *NodeNetworkStateStatus) DeepCopy() *NodeNetworkStateStatus {
	if in == nil {
		return nil
	}
	out := new(NodeNetworkStateStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in State) DeepCopyInto(out *State) {
	{
		in := &in
		*out = make(State, len(*in))
		copy(*out, *in)
		return
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new State.
func (in State) DeepCopy() State {
	if in == nil {
		return nil
	}
	out := new(State)
	in.DeepCopyInto(out)
	return *out
}
