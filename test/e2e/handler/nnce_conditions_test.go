package handler

import (
	"fmt"
	"sync"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	corev1 "k8s.io/api/core/v1"

	nmstatev1beta1 "github.com/nmstate/kubernetes-nmstate/pkg/apis/nmstate/v1beta1"
)

func invalidConfig(bridgeName string) nmstatev1beta1.State {
	return nmstatev1beta1.NewState(fmt.Sprintf(`interfaces:
  - name: %s
    type: linux-bridge
    state: invalid_state
`, bridgeName))
}

var _ = Describe("EnactmentCondition", func() {
	Context("when applying valid config", func() {
		BeforeEach(func() {
		})
		AfterEach(func() {
			By("Remove the bridge")
			updateDesiredStateAndWait(linuxBrAbsent(bridge1))

			By("Reset desired state at all nodes")
			resetDesiredStateForNodes()
		})
		It("should go from Progressing to Available", func() {
			progressConditions := []nmstatev1beta1.Condition{
				nmstatev1beta1.Condition{
					Type:   nmstatev1beta1.NodeNetworkConfigurationEnactmentConditionProgressing,
					Status: corev1.ConditionTrue,
				},
				nmstatev1beta1.Condition{
					Type:   nmstatev1beta1.NodeNetworkConfigurationEnactmentConditionAvailable,
					Status: corev1.ConditionUnknown,
				},
				nmstatev1beta1.Condition{
					Type:   nmstatev1beta1.NodeNetworkConfigurationEnactmentConditionFailing,
					Status: corev1.ConditionUnknown,
				},
				nmstatev1beta1.Condition{
					Type:   nmstatev1beta1.NodeNetworkConfigurationEnactmentConditionMatching,
					Status: corev1.ConditionTrue,
				},
			}
			availableConditions := []nmstatev1beta1.Condition{
				nmstatev1beta1.Condition{
					Type:   nmstatev1beta1.NodeNetworkConfigurationEnactmentConditionProgressing,
					Status: corev1.ConditionFalse,
				},
				nmstatev1beta1.Condition{
					Type:   nmstatev1beta1.NodeNetworkConfigurationEnactmentConditionAvailable,
					Status: corev1.ConditionTrue,
				},
				nmstatev1beta1.Condition{
					Type:   nmstatev1beta1.NodeNetworkConfigurationEnactmentConditionFailing,
					Status: corev1.ConditionFalse,
				},
				nmstatev1beta1.Condition{
					Type:   nmstatev1beta1.NodeNetworkConfigurationEnactmentConditionMatching,
					Status: corev1.ConditionTrue,
				},
			}
			var wg sync.WaitGroup
			wg.Add(len(nodes))
			for i, _ := range nodes {
				node := nodes[i]
				go func() {
					defer wg.Done()
					defer GinkgoRecover()
					By(fmt.Sprintf("Check %s progressing state is reached", node))
					enactmentConditionsStatusEventually(node).Should(ConsistOf(progressConditions))

					By(fmt.Sprintf("Check %s available state is the next condition", node))
					enactmentConditionsStatusEventually(node).Should(ConsistOf(availableConditions))

					By(fmt.Sprintf("Check %s available state is kept", node))
					enactmentConditionsStatusConsistently(node).Should(ConsistOf(availableConditions))
				}()
			}
			// Run the policy after we set the nnce conditions assert so we
			// make sure we catch progressing state.
			updateDesiredState(linuxBrUp(bridge1))

			wg.Wait()

			By("Check policy is at available state")
			waitForAvailableTestPolicy()
		})
	})

	Context("when applying invalid configuration", func() {
		BeforeEach(func() {
			updateDesiredState(invalidConfig(bridge1))

		})

		AfterEach(func() {
			By("Remove the bridge")
			updateDesiredStateAndWait(linuxBrAbsent(bridge1))
			By("Reset desired state at all nodes")
			resetDesiredStateForNodes()
		})

		It("should have Failing ConditionType set to true", func() {
			for _, node := range nodes {
				enactmentConditionsStatusEventually(node).Should(ConsistOf(
					nmstatev1beta1.Condition{
						Type:   nmstatev1beta1.NodeNetworkConfigurationEnactmentConditionFailing,
						Status: corev1.ConditionTrue,
					},
					nmstatev1beta1.Condition{
						Type:   nmstatev1beta1.NodeNetworkConfigurationEnactmentConditionAvailable,
						Status: corev1.ConditionFalse,
					},
					nmstatev1beta1.Condition{
						Type:   nmstatev1beta1.NodeNetworkConfigurationEnactmentConditionProgressing,
						Status: corev1.ConditionFalse,
					},
					nmstatev1beta1.Condition{
						Type:   nmstatev1beta1.NodeNetworkConfigurationEnactmentConditionMatching,
						Status: corev1.ConditionTrue,
					},
				))
			}
			By("Check policy is at degraded state")
			waitForDegradedTestPolicy()
		})
	})
})
