package helper

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	nmstatev1alpha1 "github.com/nmstate/kubernetes-nmstate/pkg/apis/nmstate/v1alpha1"
)

var _ = Describe("FilterOut", func() {
	var (
		state, filteredState nmstatev1alpha1.State
	)

	Context("when given empty interface", func() {
		BeforeEach(func() {
			state = nmstatev1alpha1.State(`interfaces:
- name: eth1
  state: up
  type: ethernet
- name: vethab6030bd
  state: down
  type: ethernet
`)
		})

		It("should return same state", func() {
			returnedState, err := filterOut(state, "")

			Expect(err).ToNot(HaveOccurred())
			Expect(returnedState).To(Equal(state))
		})
	})

	Context("when given invalid yaml", func() {
		BeforeEach(func() {
			state = nmstatev1alpha1.State(`invalid yaml`)
		})

		It("should return err", func() {
			_, err := filterOut(state, "veth*")

			Expect(err).To(HaveOccurred())
		})
	})

	Context("when given 2 interfaces and 1 is veth", func() {
		BeforeEach(func() {
			state = nmstatev1alpha1.State(`interfaces:
- name: eth1
  state: up
  type: ethernet
- name: vethab6030bd
  state: down
  type: ethernet
`)
			filteredState = nmstatev1alpha1.State(`interfaces:
- name: eth1
  state: up
  type: ethernet
`)
		})

		It("should return filtered 1 interface without veth", func() {
			returnedState, err := filterOut(state, "veth*")

			Expect(err).NotTo(HaveOccurred())
			Expect(returnedState).To(Equal(filteredState))
		})
	})

	Context("when given 3 interfaces and 2 are veths", func() {
		BeforeEach(func() {
			state = nmstatev1alpha1.State(`interfaces:
- name: eth1
  state: up
  type: ethernet
- name: vethab6030bd
  state: down
  type: ethernet
- name: vethjyuftrgv
  state: down
  type: ethernet
`)
			filteredState = nmstatev1alpha1.State(`interfaces:
- name: eth1
  state: up
  type: ethernet
`)
		})

		It("should return filtered 1 interface without veth", func() {
			returnedState, err := filterOut(state, "veth*")

			Expect(err).ToNot(HaveOccurred())
			Expect(returnedState).To(Equal(filteredState))
		})
	})
})
