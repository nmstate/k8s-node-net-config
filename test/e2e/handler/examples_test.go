package handler

import (
	"fmt"

	. "github.com/onsi/ginkgo"
)

type exampleSpec struct {
	name       string
	fileName   string
	policyName string
	ifaceNames []string
}

// This suite checks that all examples in our docs can be successfully applied.
// It only checks the top level API, hence it does not verify that the
// configuration is indeed applied on nodes. That should be tested by dedicated
// test suites for each feature.
var _ = Describe("Examples", func() {
	beforeTestIfaceExample := func(fileName string) {
		kubectlAndCheck("apply", "-f", fmt.Sprintf("docs/examples/%s", fileName))
	}

	testIfaceExample := func(policyName string) {
		kubectlAndCheck("wait", "nncp", policyName, "--for", "condition=Available", "--timeout", "2m")
	}

	afterIfaceExample := func(policyName string, ifaceNames []string) {
		deletePolicy(policyName)

		for _, ifaceName := range ifaceNames {
			updateDesiredState(interfaceAbsent(ifaceName))
			waitForAvailableTestPolicy()
		}

		resetDesiredStateForNodes()
	}

	examples := []exampleSpec{
		exampleSpec{
			name:       "Ethernet",
			fileName:   "ethernet.yaml",
			policyName: "ethernet",
			ifaceNames: []string{"eth1"},
		},
		exampleSpec{
			name:       "Linux bridge",
			fileName:   "linux-bridge.yaml",
			policyName: "linux-bridge",
			ifaceNames: []string{"br1"},
		},
		exampleSpec{
			name:       "OVS bridge",
			fileName:   "ovs-bridge.yaml",
			policyName: "ovs-bridge",
			ifaceNames: []string{"br1"},
		},
		exampleSpec{
			name:       "OVS bridge with interface",
			fileName:   "ovs-bridge-iface.yaml",
			policyName: "ovs-bridge-iface",
			ifaceNames: []string{"br1"},
		},
		exampleSpec{
			name:       "Linux bonding",
			fileName:   "bond.yaml",
			policyName: "bond",
			ifaceNames: []string{"bond0"},
		},
		exampleSpec{
			name:       "Linux bonding and VLAN",
			fileName:   "bond-vlan.yaml",
			policyName: "bond-vlan",
			ifaceNames: []string{"bond0.102", "bond0"},
		},
		exampleSpec{
			name:       "VLAN",
			fileName:   "vlan.yaml",
			policyName: "vlan",
			ifaceNames: []string{"eth1.102", "eth1"},
		},
		exampleSpec{
			name:       "DHCP",
			fileName:   "dhcp.yaml",
			policyName: "dhcp",
			ifaceNames: []string{"eth1"},
		},
		exampleSpec{
			name:       "Static IP",
			fileName:   "static-ip.yaml",
			policyName: "static-ip",
			ifaceNames: []string{"eth1"},
		},
		exampleSpec{
			name:       "Route",
			fileName:   "route.yaml",
			policyName: "route",
			ifaceNames: []string{"eth1"},
		},
		exampleSpec{
			name:       "DNS",
			fileName:   "dns.yaml",
			policyName: "dns",
			ifaceNames: []string{},
		},
		exampleSpec{
			name:       "Worker selector",
			fileName:   "worker-selector.yaml",
			policyName: "worker-selector",
			ifaceNames: []string{"eth1"},
		},
	}

	for _, example := range examples {
		Context(example.name, func() {
			BeforeEach(func() {
				beforeTestIfaceExample(example.fileName)
			})

			AfterEach(func() {
				afterIfaceExample(example.policyName, example.ifaceNames)
			})

			It("should succeed applying the policy", func() {
				testIfaceExample(example.policyName)
			})
		})
	}
})
