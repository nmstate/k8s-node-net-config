package handler

import (
	"fmt"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("NodeNetworkState", func() {
	Context("with multiple policies configured", func() {
		var (
			staticIpPolicy = "static-ip-policy"
			vlanPolicy     = "vlan-policy"
			ipAddress      = "10.244.0.1"
			vlanId         = "102"
			prefixLen      = "24"
		)

		BeforeEach(func() {
			setDesiredStateWithPolicy(staticIpPolicy, ifaceUpWithStaticIP(firstSecondaryNic, ipAddress, prefixLen))
			waitForAvailablePolicy(staticIpPolicy)
			setDesiredStateWithPolicy(vlanPolicy, ifaceUpWithVlanUp(firstSecondaryNic, vlanId))
			waitForAvailablePolicy(vlanPolicy)
		})

		AfterEach(func() {
			setDesiredStateWithPolicy(staticIpPolicy, ifaceDownIPv4Disabled(firstSecondaryNic))
			waitForAvailablePolicy(staticIpPolicy)
			setDesiredStateWithPolicy(vlanPolicy, vlanAbsent(firstSecondaryNic, vlanId))
			waitForAvailablePolicy(vlanPolicy)
			deletePolicy(staticIpPolicy)
			deletePolicy(vlanPolicy)
			resetDesiredStateForNodes()
		})

		It("should have the IP and vlan interface configured", func() {
			for _, node := range nodes {
				ipAddressForNodeInterfaceEventually(node, firstSecondaryNic).Should(Equal(ipAddress))
				interfacesNameForNodeEventually(node).Should(ContainElement(fmt.Sprintf(`%s.%s`, firstSecondaryNic, vlanId)))
				vlanForNodeInterfaceEventually(node, fmt.Sprintf(`%s.%s`, firstSecondaryNic, vlanId)).Should(Equal(vlanId))
			}
		})
	})
})
