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

package controllers

import (
	"context"
	"fmt"
	"reflect"
	"time"

	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	"k8s.io/client-go/util/retry"

	ctrl "sigs.k8s.io/controller-runtime"
	builder "sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"

	nmstateapi "github.com/nmstate/kubernetes-nmstate/api/shared"
	nmstatev1beta1 "github.com/nmstate/kubernetes-nmstate/api/v1beta1"
	"github.com/nmstate/kubernetes-nmstate/pkg/enactment"
	"github.com/nmstate/kubernetes-nmstate/pkg/enactmentstatus"
	enactmentconditions "github.com/nmstate/kubernetes-nmstate/pkg/enactmentstatus/conditions"
	"github.com/nmstate/kubernetes-nmstate/pkg/environment"
	nmstate "github.com/nmstate/kubernetes-nmstate/pkg/helper"
	"github.com/nmstate/kubernetes-nmstate/pkg/node"
	"github.com/nmstate/kubernetes-nmstate/pkg/policyconditions"
	"github.com/nmstate/kubernetes-nmstate/pkg/probe"
	"github.com/nmstate/kubernetes-nmstate/pkg/selectors"
)

var (
	nodeName                                string
	nodeRunningUpdateRetryTime              = 5 * time.Second
	onCreateOrUpdateWithDifferentGeneration = predicate.Funcs{
		CreateFunc: func(createEvent event.CreateEvent) bool {
			return true
		},
		DeleteFunc: func(deleteEvent event.DeleteEvent) bool {
			return false
		},
		UpdateFunc: func(updateEvent event.UpdateEvent) bool {
			// [1] https://blog.openshift.com/kubernetes-operators-best-practices/
			generationIsDifferent := updateEvent.ObjectNew.GetGeneration() != updateEvent.ObjectOld.GetGeneration()
			return generationIsDifferent
		},
	}

	onLabelsUpdatedForThisNode = predicate.Funcs{
		CreateFunc: func(createEvent event.CreateEvent) bool {
			return false
		},
		DeleteFunc: func(event.DeleteEvent) bool {
			return false
		},
		UpdateFunc: func(updateEvent event.UpdateEvent) bool {
			labelsChanged := !reflect.DeepEqual(updateEvent.ObjectOld.GetLabels(), updateEvent.ObjectNew.GetLabels())
			return labelsChanged && nmstate.EventIsForThisNode(updateEvent.ObjectNew)
		},
		GenericFunc: func(event.GenericEvent) bool {
			return false
		},
	}
)

// NodeNetworkConfigurationPolicyReconciler reconciles a NodeNetworkConfigurationPolicy object
type NodeNetworkConfigurationPolicyReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

func init() {
	if !environment.IsHandler() {
		return
	}

	nodeName = environment.NodeName()
	if len(nodeName) == 0 {
		panic("NODE_NAME is mandatory")
	}
}

// Reconcile reads that state of the cluster for a NodeNetworkConfigurationPolicy object and makes changes based on the state read
// and what is in the NodeNetworkConfigurationPolicy.Spec
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *NodeNetworkConfigurationPolicyReconciler) Reconcile(ctx context.Context, request ctrl.Request) (ctrl.Result, error) {
	_ = context.Background()
	log := r.Log.WithValues("nodenetworkconfigurationpolicy", request.NamespacedName)

	// Fetch the NodeNetworkConfigurationPolicy instance
	instance := &nmstatev1beta1.NodeNetworkConfigurationPolicy{}
	err := r.Client.Get(context.TODO(), request.NamespacedName, instance)
	if err != nil {
		if apierrors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			return ctrl.Result{}, nil
		}
		log.Error(err, "Error retrieving policy")
		// Error reading the object - requeue the request.
		return ctrl.Result{}, err
	}

	policyconditions.Reset(r.Client, request.NamespacedName)

	previousConditions, err := r.initializeEnactment(*instance)
	if err != nil {
		log.Error(err, "Error initializing enactment")
		return ctrl.Result{}, err
	}

	enactmentConditions := enactmentconditions.New(r.Client, nmstateapi.EnactmentKey(nodeName, instance.Name))

	// Policy conditions will be updated at the end so updating it
	// does not impact at applying state, it will increase just
	// reconcile time.
	defer policyconditions.Update(r.Client, request.NamespacedName)

	policySelectors := selectors.NewFromPolicy(r.Client, *instance)
	unmatchingNodeLabels, err := policySelectors.UnmatchedNodeLabels(nodeName)
	if err != nil {
		log.Error(err, "failed checking node selectors")
		enactmentConditions.NotifyNodeSelectorFailure(err)
	}
	if len(unmatchingNodeLabels) > 0 {
		log.Info("Policy node selectors does not match node")
		enactmentConditions.NotifyNodeSelectorNotMatching(unmatchingNodeLabels)
		return ctrl.Result{}, nil
	}

	enactmentInstance, err := r.enactmentForPolicy(instance)
	if err != nil {
		log.Error(err, "error getting enactment for policy")
		return ctrl.Result{}, err
	}

	enactmentConditions.NotifyMatching()

	enactmentCount, err := enactment.CountByPolicy(r.Client, instance)
	if err != nil {
		log.Error(err, "Error getting enactment counts")
		return ctrl.Result{}, err
	}
	if enactmentCount.Failed() > 0 {
		err = fmt.Errorf("policy has failing enactments, aborting")
		log.Error(err, "")
		enactmentConditions.NotifyAborted(err)
		return ctrl.Result{}, nil
	}

	if r.shouldIncrementUnavailableNodeCount(instance, previousConditions) {
		err = r.incrementUnavailableNodeCount(instance)
		if err != nil {
			if apierrors.IsConflict(err) {
				log.Info(err.Error())
				return ctrl.Result{RequeueAfter: nodeRunningUpdateRetryTime}, nil
			}
			return ctrl.Result{}, err
		}
	}
	defer r.decrementUnavailableNodeCount(instance)

	enactmentConditions.NotifyProgressing()
	nmstateOutput, err := nmstate.ApplyDesiredState(r.Client, enactmentInstance.Status.DesiredState)
	if err != nil {
		errmsg := fmt.Errorf("error reconciling NodeNetworkConfigurationPolicy at desired state apply: %s, %v", nmstateOutput, err)

		enactmentConditions.NotifyFailedToConfigure(errmsg)
		log.Error(errmsg, fmt.Sprintf("Rolling back network configuration, manual intervention needed: %s", nmstateOutput))
		return ctrl.Result{}, nil
	}
	log.Info("nmstate", "output", nmstateOutput)

	enactmentConditions.NotifySuccess()

	r.forceNNSRefresh(nodeName)

	return ctrl.Result{}, nil
}

func (r *NodeNetworkConfigurationPolicyReconciler) SetupWithManager(mgr ctrl.Manager) error {
	allPolicies := handler.MapFunc(
		func(client.Object) []reconcile.Request {
			log := r.Log.WithName("allPolicies")
			allPoliciesAsRequest := []reconcile.Request{}
			policyList := nmstatev1beta1.NodeNetworkConfigurationPolicyList{}
			err := r.Client.List(context.TODO(), &policyList)
			if err != nil {
				log.Error(err, "failed listing all NodeNetworkConfigurationPolicies to re-reconcile them after node created or updated")
				return []reconcile.Request{}
			}
			for _, policy := range policyList.Items {
				allPoliciesAsRequest = append(allPoliciesAsRequest, reconcile.Request{
					NamespacedName: types.NamespacedName{
						Name: policy.Name,
					}})
			}
			return allPoliciesAsRequest
		})

	// Reconcile NNCP if they are created or updated
	err := ctrl.NewControllerManagedBy(mgr).
		For(&nmstatev1beta1.NodeNetworkConfigurationPolicy{}).
		WithEventFilter(onCreateOrUpdateWithDifferentGeneration).
		Complete(r)
	if err != nil {
		return errors.Wrap(err, "failed to add controller to NNCP Reconciler listening NNCP events")
	}

	// Reconcile all NNCPs if Node is updated (for example labels are changed), node creation event
	// is not needed since all NNCPs are going to be Reconcile at node startup.
	err = ctrl.NewControllerManagedBy(mgr).
		For(&corev1.Node{}).
		Watches(&source.Kind{Type: &corev1.Node{}}, handler.EnqueueRequestsFromMapFunc(allPolicies), builder.WithPredicates(onLabelsUpdatedForThisNode)).
		Complete(r)
	if err != nil {
		return errors.Wrap(err, "failed to add controller to NNCP Reconciler listening Node events")
	}
	return nil
}

func (r *NodeNetworkConfigurationPolicyReconciler) initializeEnactment(policy nmstatev1beta1.NodeNetworkConfigurationPolicy) (*nmstateapi.ConditionList, error) {
	desiredStateWithDefaults, err := nmstate.ApplyDefaultVlanFiltering(policy.Spec.DesiredState)
	if err != nil {
		return nil, errors.Wrap(err, "error applying defaults to policy desiredState")
	}

	enactmentKey := nmstateapi.EnactmentKey(nodeName, policy.Name)
	log := r.Log.WithName("initializeEnactment").WithValues("policy", policy.Name, "enactment", enactmentKey.Name)
	// Return if it's already initialize or we cannot retrieve it
	enactment := nmstatev1beta1.NodeNetworkConfigurationEnactment{}
	err = r.Client.Get(context.TODO(), enactmentKey, &enactment)
	if err != nil && !apierrors.IsNotFound(err) {
		return nil, errors.Wrap(err, "failed getting enactment ")
	}
	if err != nil && apierrors.IsNotFound(err) {
		log.Info("creating enactment")
		enactment = nmstatev1beta1.NewEnactment(nodeName, policy)
		err = r.Client.Create(context.TODO(), &enactment)
		if err != nil {
			return nil, errors.Wrapf(err, "error creating NodeNetworkConfigurationEnactment: %+v", enactment)
		}
		err = r.waitEnactmentCreated(enactmentKey)
		if err != nil {
			return nil, errors.Wrapf(err, "error waitting for NodeNetworkConfigurationEnactment: %+v", enactment)
		}
	} else {
		enactmentConditions := enactmentconditions.New(r.Client, enactmentKey)
		enactmentConditions.Reset()
	}

	return &enactment.Status.Conditions, enactmentstatus.Update(r.Client, enactmentKey, func(status *nmstateapi.NodeNetworkConfigurationEnactmentStatus) {
		status.DesiredState = desiredStateWithDefaults
		status.PolicyGeneration = policy.Generation
	})
}

func (r *NodeNetworkConfigurationPolicyReconciler) enactmentForPolicy(policy *nmstatev1beta1.NodeNetworkConfigurationPolicy) (*nmstatev1beta1.NodeNetworkConfigurationEnactment, error) {
	enactmentKey := nmstateapi.EnactmentKey(nodeName, policy.Name)
	instance := &nmstatev1beta1.NodeNetworkConfigurationEnactment{}
	err := r.Client.Get(context.TODO(), enactmentKey, instance)
	if err != nil {
		return nil, errors.Wrap(err, "getting enactment failed")
	}
	return instance, nil
}

func (r *NodeNetworkConfigurationPolicyReconciler) waitEnactmentCreated(enactmentKey types.NamespacedName) error {
	var enactment nmstatev1beta1.NodeNetworkConfigurationEnactment
	pollErr := wait.PollImmediate(1*time.Second, 10*time.Second, func() (bool, error) {
		err := r.Client.Get(context.TODO(), enactmentKey, &enactment)
		if err != nil {
			if apierrors.IsNotFound(err) {
				// Let's retry after a while, sometimes it takes some time
				// for enactment to be created
				return false, nil
			}
			return false, err
		}
		return true, nil
	})

	return pollErr
}

func (r *NodeNetworkConfigurationPolicyReconciler) shouldIncrementUnavailableNodeCount(policy *nmstatev1beta1.NodeNetworkConfigurationPolicy, conditions *nmstateapi.ConditionList) bool {
	return !enactmentstatus.IsProgressing(conditions) && (policy.Status.LastUnavailableNodeCountUpdate == nil ||
		time.Since(policy.Status.LastUnavailableNodeCountUpdate.Time) < (nmstate.DesiredStateConfigurationTimeout+probe.ProbesTotalTimeout))
}

func (r *NodeNetworkConfigurationPolicyReconciler) incrementUnavailableNodeCount(policy *nmstatev1beta1.NodeNetworkConfigurationPolicy) error {
	policyKey := types.NamespacedName{Name: policy.GetName(), Namespace: policy.GetNamespace()}
	err := r.Client.Get(context.TODO(), policyKey, policy)
	if err != nil {
		return err
	}
	maxUnavailable, err := node.MaxUnavailableNodeCount(r.Client, policy)
	if err != nil {
		return err
	}
	if policy.Status.UnavailableNodeCount >= maxUnavailable {
		return apierrors.NewConflict(schema.GroupResource{Resource: "nodenetworkconfigurationpolicies"}, policy.Name, fmt.Errorf("maximal number of %d nodes are already processing policy configuration", policy.Status.UnavailableNodeCount))
	}
	policy.Status.LastUnavailableNodeCountUpdate = &metav1.Time{Time: time.Now()}
	policy.Status.UnavailableNodeCount += 1
	err = r.Client.Status().Update(context.TODO(), policy)
	if err != nil {
		return err
	}
	return nil
}

func (r *NodeNetworkConfigurationPolicyReconciler) decrementUnavailableNodeCount(policy *nmstatev1beta1.NodeNetworkConfigurationPolicy) {
	policyKey := types.NamespacedName{Name: policy.GetName(), Namespace: policy.GetNamespace()}
	instance := &nmstatev1beta1.NodeNetworkConfigurationPolicy{}
	err := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		err := r.Client.Get(context.TODO(), policyKey, instance)
		if err != nil {
			return err
		}
		if instance.Status.UnavailableNodeCount <= 0 {
			return fmt.Errorf("no unavailable nodes")
		}
		instance.Status.LastUnavailableNodeCountUpdate = &metav1.Time{Time: time.Now()}
		instance.Status.UnavailableNodeCount -= 1
		return r.Client.Status().Update(context.TODO(), instance)
	})
	if err != nil {
		r.Log.Error(err, "error decrementing unavailableNodeCount")
	}
}

func (r *NodeNetworkConfigurationPolicyReconciler) forceNNSRefresh(name string) {
	log := r.Log.WithName("forceNNSRefresh").WithValues("node", name)
	log.Info("forcing NodeNetworkState refresh after NNCP applied")
	nns := &nmstatev1beta1.NodeNetworkState{}
	err := r.Client.Get(context.TODO(), types.NamespacedName{Name: name}, nns)
	if err != nil {
		log.WithValues("error", err).Info("WARNING: failed retrieving NodeNetworkState to force refresh, it will be refreshed after regular period")
		return
	}
	if nns.Labels == nil {
		nns.Labels = map[string]string{}
	}
	nns.Labels[forceRefreshLabel] = fmt.Sprintf("%d", time.Now().UnixNano())

	err = r.Client.Update(context.Background(), nns)
	if err != nil {
		log.WithValues("error", err).Info("WARNING: failed forcing NNS refresh, it will be refreshed after regular period")
	}
}

func desiredState(object runtime.Object) (nmstateapi.State, error) {
	var state nmstateapi.State
	switch v := object.(type) {
	default:
		return nmstateapi.State{}, fmt.Errorf("unexpected type %T", v)
	case *nmstatev1beta1.NodeNetworkConfigurationPolicy:
		state = v.Spec.DesiredState
	}
	return state, nil
}
