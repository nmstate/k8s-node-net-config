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

package nmpolicy

import (
	"fmt"
	"time"

	nmstateapi "github.com/nmstate/kubernetes-nmstate/api/shared"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	nmstateapiv2 "github.com/nmstate/nmstate/rust/src/go/api/v2"

	nmpolicytypes "github.com/nmstate/nmpolicy/nmpolicy/types"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("NMPolicy GenerateState", func() {
	When("fails", func() {
		It("Should return an error", func() {
			capturedState, desiredState, err := GenerateStateWithStateGenerator(
				nmpolicyStub{shouldFail: true},
				nmstateapi.State{},
				nmstateapi.NodeNetworkConfigurationPolicySpec{},
				nmstateapi.State{},
				map[string]nmstateapi.NodeNetworkConfigurationEnactmentCapturedState{},
			)
			Expect(err).To(HaveOccurred())
			Expect(capturedState).To(Equal(map[string]nmstateapi.NodeNetworkConfigurationEnactmentCapturedState{}))
			Expect(desiredState).To(Equal(nmstateapiv2.NetworkState{}))
		})

	})

	When("succeeds", func() {
		It("should generated a desired state", func() {
			const desiredStateYaml = `
interfaces:
- name: eth1
  type: ethernet
  state: up
`
			const captureYaml1 = `default-gw expression`
			const captureYaml2 = `base-iface expression`
			const metaVersion = "5"
			metaTime := time.Now()

			nmpolicyMetaInfo := nmpolicytypes.MetaInfo{
				Version:   metaVersion,
				TimeStamp: metaTime,
			}

			generatedState := nmpolicytypes.GeneratedState{
				DesiredState: []byte(desiredStateYaml),
				Cache: nmpolicytypes.CachedState{
					Capture: map[string]nmpolicytypes.CaptureState{
						"default-gw": {State: []byte(captureYaml1), MetaInfo: nmpolicyMetaInfo},
						"base-iface": {State: []byte(captureYaml2)},
					},
				},
			}

			capturedStates, desiredState, err := GenerateStateWithStateGenerator(
				nmpolicyStub{output: generatedState},
				nmstateapi.State{},
				nmstateapi.NodeNetworkConfigurationPolicySpec{},
				nmstateapi.State{},
				map[string]nmstateapi.NodeNetworkConfigurationEnactmentCapturedState{},
			)

			Expect(err).NotTo(HaveOccurred())

			expectedMetaInfo := nmstateapi.NodeNetworkConfigurationEnactmentMetaInfo{
				Version:   metaVersion,
				TimeStamp: metav1.NewTime(metaTime),
			}

			expectedcCaptureCache := map[string]nmstateapi.NodeNetworkConfigurationEnactmentCapturedState{
				"default-gw": {State: nmstateapi.State{Raw: []byte(captureYaml1)}, MetaInfo: expectedMetaInfo},
				"base-iface": {State: nmstateapi.State{Raw: []byte(captureYaml2)}},
			}

			Expect(capturedStates).To(Equal(expectedcCaptureCache))
			Expect(desiredState).To(Equal(nmstateapiv2.NetworkState{
				Interfaces: []nmstateapiv2.Interface{{
					BaseInterface: nmstateapiv2.BaseInterface{
						Name:  "eth1",
						Type:  nmstateapiv2.InterfaceTypeEthernet,
						State: nmstateapiv2.InterfaceStateUp,
					},
				}},
			}))
		})
	})
})

type nmpolicyStub struct {
	shouldFail bool
	output     nmpolicytypes.GeneratedState
}

func (n nmpolicyStub) GenerateState(
	nmpolicySpec nmpolicytypes.PolicySpec,
	currentState []byte,
	cache nmpolicytypes.CachedState,
) (nmpolicytypes.GeneratedState, error) {
	if n.shouldFail {
		return nmpolicytypes.GeneratedState{}, fmt.Errorf("GenerateStateFailed")
	}
	return n.output, nil
}
