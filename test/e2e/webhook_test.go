package e2e

import (
	"strconv"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	nmstatev1alpha1 "github.com/nmstate/kubernetes-nmstate/pkg/apis/nmstate/v1alpha1"
	nncpwebhook "github.com/nmstate/kubernetes-nmstate/pkg/webhook/nodenetworkconfigurationpolicy"
)

// We just check the labe at CREATE/UPDATE events since mutated data is already
// check at unit test.
var _ = Describe("Mutating Admission Webhook", func() {
	Context("when policy is created", func() {
		BeforeEach(func() {
			// Make sure test policy is not there so
			// we exercise CREATE event
			resetDesiredStateForNodes()
			updateDesiredState(linuxBrUp(bridge1))
			waitForAvailableTestPolicy()
		})
		AfterEach(func() {
			updateDesiredState(linuxBrAbsent(bridge1))
			waitForAvailableTestPolicy()
			resetDesiredStateForNodes()
		})

		It("should add an annotation with mutation timestamp", func() {
			policy := nodeNetworkConfigurationPolicy(TestPolicy)
			Expect(policy.ObjectMeta.Annotations).To(HaveKey(nncpwebhook.TimestampLabelKey))
		})
		Context("and we updated it", func() {
			var (
				oldPolicy nmstatev1alpha1.NodeNetworkConfigurationPolicy
			)
			BeforeEach(func() {
				oldPolicy = nodeNetworkConfigurationPolicy(TestPolicy)
				updateDesiredState(linuxBrAbsent(bridge1))
				waitForAvailableTestPolicy()
			})
			It("should update annotation with newer mutation timestamp", func() {
				newPolicy := nodeNetworkConfigurationPolicy(TestPolicy)
				Expect(newPolicy.ObjectMeta.Annotations).To(HaveKey(nncpwebhook.TimestampLabelKey))

				oldAnnotation := oldPolicy.ObjectMeta.Annotations[nncpwebhook.TimestampLabelKey]
				oldConditionsMutation, err := strconv.ParseInt(oldAnnotation, 10, 64)
				Expect(err).ToNot(HaveOccurred())
				newAnnotation := newPolicy.ObjectMeta.Annotations[nncpwebhook.TimestampLabelKey]
				newConditionsMutation, err := strconv.ParseInt(newAnnotation, 10, 64)
				Expect(err).ToNot(HaveOccurred())

				Expect(newConditionsMutation).To(BeNumerically(">", oldConditionsMutation), "mutation timestamp not updated")
			})
		})
	})
})
