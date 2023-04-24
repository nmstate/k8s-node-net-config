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

package state

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	nmstate "github.com/nmstate/kubernetes-nmstate/api/shared"
)

var _ = Describe("FilterOut", func() {
	var (
		state, filteredState nmstate.State
	)

	Context("when there is a linux bridge with gc-timer and hello-timer", func() {
		BeforeEach(func() {
			state = nmstate.NewState(`
interfaces:
- name: eth1
  state: up
  type: ethernet
- name: br1
  bridge:
    options:
      gc-timer: 13715
      group-addr: 01:80:C2:00:00:00
      group-forward-mask: 0
      hash-max: 512
      hello-timer: 0
      stp:
        enabled: false
    port: []
  ipv4:
    address:
    - ip: 172.17.0.1
      prefix-length: 16
    dhcp: false
    enabled: true
  ipv6:
    address:
    - ip: 2001:db9:1::1
      prefix-length: 64
    - ip: fe80::1
      prefix-length: 64
    autoconf: false
    dhcp: false
    enabled: true
  lldp:
    enabled: false
  mac-address: 02:42:BB:10:B8:9F
  mtu: 1500
  name: br1
  state: up
  type: linux-bridge
routes:
  config: []
  running:
  - destination: 0.0.0.0/0
    metric: 102
    next-hop-address: 192.168.66.2
    next-hop-interface: eth1
    table-id: 254
`)
			filteredState = nmstate.NewState(`
interfaces:
- name: eth1
  state: up
  type: ethernet
- name: br1
  bridge:
    options:
      group-addr: 01:80:C2:00:00:00
      group-forward-mask: 0
      hash-max: 512
      stp:
        enabled: false
    port: []
  ipv4:
    address:
    - ip: 172.17.0.1
      prefix-length: 16
    dhcp: false
    enabled: true
  ipv6:
    address:
    - ip: 2001:db9:1::1
      prefix-length: 64
    - ip: fe80::1
      prefix-length: 64
    autoconf: false
    dhcp: false
    enabled: true
  lldp:
    enabled: false
  mac-address: 02:42:BB:10:B8:9F
  mtu: 1500
  name: br1
  state: up
  type: linux-bridge
routes:
  config: []
  running:
  - destination: 0.0.0.0/0
    metric: 102
    next-hop-address: 192.168.66.2
    next-hop-interface: eth1
    table-id: 254
`)
		})
		It("should remove dynamic attributes from linux-bridge interface", func() {
			returnedState, err := filterOut(state)
			Expect(err).ToNot(HaveOccurred())
			Expect(returnedState).To(MatchYAML(filteredState))
		})
	})

	Context("when there is managed veth interface", func() {
		BeforeEach(func() {
			state = nmstate.NewState(`interfaces:
- name: vethab6030bd
  state: down
  type: veth
  veth:
    peer: eth2
routes:
  config: []
  running:
  - destination: fd10:244::8c40/128
    metric: 1024
    next-hop-address: ""
    next-hop-interface: vethab6030bd
    table-id: 254
`)
			filteredState = nmstate.NewState(`interfaces:
- name: vethab6030bd
  state: down
  type: veth
  veth:
    peer: eth2
routes:
  config: []
  running:
  - destination: fd10:244::8c40/128
    metric: 1024
    next-hop-address: ""
    next-hop-interface: vethab6030bd
    table-id: 254
`)
		})

		It("should keep managed veth interface", func() {
			returnedState, err := filterOut(state)
			Expect(err).NotTo(HaveOccurred())
			Expect(returnedState).To(MatchYAML(filteredState))
		})
	})

	Context("when there is unmanaged veth interface", func() {
		BeforeEach(func() {
			state = nmstate.NewState(`interfaces:
- name: vethab6030bd
  state: ignore
  type: veth
  veth:
    peer: eth2
routes:
  config: []
  running:
  - destination: fd10:244::8c40/128
    metric: 1024
    next-hop-address: ""
    next-hop-interface: vethab6030bd
    table-id: 254
`)
			filteredState = nmstate.NewState(`interfaces: []
routes:
  config: []
  running: []
`)
		})

		It("should filter unmanaged veth interface", func() {
			returnedState, err := filterOut(state)
			Expect(err).NotTo(HaveOccurred())
			Expect(returnedState).To(MatchYAML(filteredState))
		})
	})

	Context("when there are multiple managed and unmanaged interfaces", func() {
		BeforeEach(func() {
			state = nmstate.NewState(`interfaces:
- name: eth1
  state: up
  type: ethernet
- name: veth101
  state: down
  type: veth
  veth:
    peer: eth2
- name: veth102
  state: ignore
  type: veth
  veth:
    peer: eth2
- name: vethjyuftrgv
  state: down
  type: veth
  veth:
    peer: eth2
- name: vethvasziovs
  state: ignore
  type: veth
  veth:
    peer: eth2
routes:
  config: []
  running:
  - destination: fd10:244::8c40/128
    metric: 1024
    next-hop-address: ""
    next-hop-interface: veth101
    table-id: 254
  - destination: fd10:244::8c40/128
    metric: 1024
    next-hop-address: ""
    next-hop-interface: veth102
    table-id: 254
  - destination: fd10:244::8c40/128
    metric: 1024
    next-hop-address: ""
    next-hop-interface: vethjyuftrgv
    table-id: 254
  - destination: fd10:244::8c40/128
    metric: 1024
    next-hop-address: ""
    next-hop-interface: vethvasziovs
    table-id: 254
  - destination: 0.0.0.0/0
    metric: 102
    next-hop-address: 192.168.66.2
    next-hop-interface: eth1
    table-id: 254
`)
			filteredState = nmstate.NewState(`interfaces:
- name: eth1
  state: up
  type: ethernet
- name: veth101
  state: down
  type: veth
  veth:
    peer: eth2
- name: vethjyuftrgv
  state: down
  type: veth
  veth:
    peer: eth2
routes:
  config: []
  running:
  - destination: fd10:244::8c40/128
    metric: 1024
    next-hop-address: ""
    next-hop-interface: veth101
    table-id: 254
  - destination: fd10:244::8c40/128
    metric: 1024
    next-hop-address: ""
    next-hop-interface: vethjyuftrgv
    table-id: 254
  - destination: 0.0.0.0/0
    metric: 102
    next-hop-address: 192.168.66.2
    next-hop-interface: eth1
    table-id: 254
`)
		})
		It("should filter out all unmanaged veth interfaces", func() {
			returnedState, err := filterOut(state)
			Expect(err).ToNot(HaveOccurred())
			Expect(returnedState).To(MatchYAML(filteredState))
		})
	})

	Context("With DNS Resolver populated", func() {
		BeforeEach(func() {
			state = nmstate.NewState(`interfaces:
  - name: eth1
    state: up
    type: ethernet
dns-resolver:
  config:
    search:
    - example.com
    - example.org
    server:
    - 2001:4860:4860::8888
    - 8.8.8.8
  running:
    search:
    - example.running.com
    - example.running.org
    server:
    - 8.8.4.4`)
		})

		It("Should keep the DNS Resolver intact", func() {
			returnedState, err := filterOut(state)
			Expect(err).ToNot(HaveOccurred())
			Expect(returnedState).To(MatchYAML(state))
		})
	})

	Context("when the interfaces have numeric characters", func() {
		BeforeEach(func() {
			state = nmstate.NewState(`interfaces:
  - name: eth0
    type: ethernet
  - name: '0'
    type: veth
    veth:
      peer: eth2
    state: ignore
  - name: '1101010'
    type: veth
    state: ignore
    veth:
      peer: eth2
  - name: '0.0'
    type: veth
    veth:
      peer: eth2
    state: ignore
  - name: '1.0'
    type: veth
    veth:
      peer: eth2
    state: ignore
  - name: '0xfe'
    type: veth
    veth:
      peer: eth2
    state: ignore
  - name: '60.e+02'
    type: veth
    veth:
      peer: eth2
    state: ignore
  - name: 10e+02
    type: veth
    veth:
      peer: eth2
    state: ignore
  - name: 70e+02
    type: veth
    veth:
      peer: eth2
    state: ignore
  - name: 94475496822e234
    type: veth
    veth:
      peer: eth2
    state: ignore
routes:
  config: []
  running:
  - destination: fd10:244::8c40/128
    metric: 1024
    next-hop-address: 10.21.21.10
    next-hop-interface: eth0
    table-id: 254
  - destination: fd10:244::8c40/128
    metric: 1024
    next-hop-address: 10.21.21.10
    next-hop-interface: 94475496822e234
    table-id: 254
  - destination: fd10:244::8c40/128
    metric: 1024
    next-hop-address: 10.21.21.10
    next-hop-interface: '94475496822e234'
    table-id: 254
  - destination: fd10:244::8c40/128
    metric: 1024
    next-hop-address: 10.21.21.10
    next-hop-interface: 70e+02
    table-id: 254
  - destination: fd10:244::8c40/128
    metric: 1024
    next-hop-address: 10.21.21.10
    next-hop-interface: 60.e+02
    table-id: 254
`)
			filteredState = nmstate.NewState(`interfaces:
- name: eth0
  type: ethernet
routes:
  config: []
  running:
  - destination: fd10:244::8c40/128
    metric: 1024
    next-hop-address: 10.21.21.10
    next-hop-interface: eth0
    table-id: 254
`)
		})

		It("should filter out interfaces correctly", func() {
			returnedState, err := filterOut(state)
			Expect(err).NotTo(HaveOccurred())
			Expect(returnedState).To(MatchYAML(filteredState))
		})
	})
})
