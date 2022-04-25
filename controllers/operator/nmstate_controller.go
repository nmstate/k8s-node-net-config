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
	"os"
	"path/filepath"
	goruntime "runtime"
	"sort"

	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	"github.com/openshift/cluster-network-operator/pkg/apply"
	"github.com/openshift/cluster-network-operator/pkg/render"

	"github.com/nmstate/kubernetes-nmstate/api/names"
	nmstatev1 "github.com/nmstate/kubernetes-nmstate/api/v1"
	nmstaterenderer "github.com/nmstate/kubernetes-nmstate/pkg/render"
)

// NMStateReconciler reconciles a NMState object
type NMStateReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups="",resources=services;endpoints;persistentvolumeclaims;events;configmaps;secrets;pods,verbs="*"
// ,namespace="{{ .OperatorNamespace }}"
// +kubebuilder:rbac:groups=apps,resources=deployments;daemonsets;replicasets;statefulsets,verbs="*",namespace="{{ .OperatorNamespace }}"
// +kubebuilder:rbac:groups=policy,resources=poddisruptionbudgets,verbs="*",namespace="{{ .OperatorNamespace }}"
// +kubebuilder:rbac:groups=admissionregistration.k8s.io,resources=mutatingwebhookconfigurations,verbs="*"
// +kubebuilder:rbac:groups=rbac.authorization.k8s.io,resources=clusterroles;clusterrolebindings;rolebindings;roles,verbs="*"
// +kubebuilder:rbac:groups=nmstate.io,resources="*",verbs="*"
// +kubebuilder:rbac:groups=apiextensions.k8s.io,resources="*",verbs="*"
// +kubebuilder:rbac:groups=apps,resources=deployments;daemonsets;replicasets;statefulsets,verbs="*"
// +kubebuilder:rbac:groups="",resources=serviceaccounts;configmaps;namespaces,verbs="*"

func (r *NMStateReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = context.Background()
	_ = r.Log.WithValues("nmstate", req.NamespacedName)

	// Fetch the NMState instance
	instanceList := &nmstatev1.NMStateList{}
	err := r.Client.List(context.TODO(), instanceList, &client.ListOptions{})
	if err != nil {
		return ctrl.Result{}, errors.Wrap(err, "failed listing all NMState instances")
	}
	instance := &nmstatev1.NMState{}
	err = r.Client.Get(context.TODO(), req.NamespacedName, instance)
	if err != nil {
		if apierrors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile req.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			return ctrl.Result{}, nil
		}
		// Error reading the object - requeue the req.
		return ctrl.Result{}, err
	}

	// We only want one instance of NMState. Ignore anything after that.
	if len(instanceList.Items) > 0 {
		if len(instanceList.Items) > 1 {
			sort.Slice(instanceList.Items, func(i, j int) bool {
				return instanceList.Items[j].CreationTimestamp.After(instanceList.Items[i].CreationTimestamp.Time)
			})
		}
		if instanceList.Items[0].Name != req.Name {
			r.Log.Info("Ignoring NMState.nmstate.io because one already exists and does not match existing name")
			err = r.Client.Delete(context.TODO(), instance, &client.DeleteOptions{})
			if err != nil {
				r.Log.Error(err, "failed to remove NMState.nmstate.io instance")
			}
			return ctrl.Result{}, nil
		}
	}

	err = r.applyCRDs(instance)
	if err != nil {
		errors.Wrap(err, "failed applying CRDs")
		return ctrl.Result{}, err
	}

	err = r.applyNamespace(instance)
	if err != nil {
		errors.Wrap(err, "failed applying Namespace")
		return ctrl.Result{}, err
	}

	err = r.applyRBAC(instance)
	if err != nil {
		errors.Wrap(err, "failed applying RBAC")
		return ctrl.Result{}, err
	}

	err = r.applyHandler(instance)
	if err != nil {
		errors.Wrap(err, "failed applying Handler")
		return ctrl.Result{}, err
	}

	r.Log.Info("Reconcile complete.")
	return ctrl.Result{}, nil
}

func (r *NMStateReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&nmstatev1.NMState{}).
		Complete(r)
}

func (r *NMStateReconciler) applyCRDs(instance *nmstatev1.NMState) error {
	data := render.MakeRenderData()
	return r.renderAndApply(instance, data, "crds", false)
}

func (r *NMStateReconciler) applyNamespace(instance *nmstatev1.NMState) error {
	data := render.MakeRenderData()
	data.Data["HandlerNamespace"] = os.Getenv("HANDLER_NAMESPACE")
	data.Data["HandlerPrefix"] = os.Getenv("HANDLER_PREFIX")
	return r.renderAndApply(instance, data, "namespace", false)
}

func (r *NMStateReconciler) applyRBAC(instance *nmstatev1.NMState) error {
	data := render.MakeRenderData()
	data.Data["HandlerNamespace"] = os.Getenv("HANDLER_NAMESPACE")
	data.Data["HandlerImage"] = os.Getenv("RELATED_IMAGE_HANDLER_IMAGE")
	data.Data["HandlerPullPolicy"] = os.Getenv("HANDLER_IMAGE_PULL_POLICY")
	data.Data["HandlerPrefix"] = os.Getenv("HANDLER_PREFIX")
	return r.renderAndApply(instance, data, "rbac", true)
}

func (r *NMStateReconciler) applyHandler(instance *nmstatev1.NMState) error {
	data := render.MakeRenderData()
	// Register ToYaml template method
	data.Funcs["toYaml"] = nmstaterenderer.ToYaml
	// Prepare defaults
	masterExistsNoScheduleToleration := corev1.Toleration{
		Key:      "node-role.kubernetes.io/master",
		Operator: corev1.TolerationOpExists,
		Effect:   corev1.TaintEffectNoSchedule,
	}
	operatorExistsToleration := corev1.Toleration{
		Key:      "",
		Operator: corev1.TolerationOpExists,
	}
	archOnMasterNodeSelector := map[string]string{
		"kubernetes.io/arch":             goruntime.GOARCH,
		"node-role.kubernetes.io/master": "",
	}
	archAndCRNodeSelector := instance.Spec.NodeSelector
	if archAndCRNodeSelector == nil {
		archAndCRNodeSelector = map[string]string{}
	}
	archAndCRNodeSelector["kubernetes.io/arch"] = goruntime.GOARCH
	archAndCRNodeSelector["kubernetes.io/os"] = "linux"

	handlerTolerations := instance.Spec.Tolerations
	if handlerTolerations == nil {
		handlerTolerations = []corev1.Toleration{operatorExistsToleration}
	}
	handlerAffinity := instance.Spec.Affinity
	if handlerAffinity == nil {
		handlerAffinity = &corev1.Affinity{}
	}

	const (
		webhookMinReplicas int32 = 1
		webhookReplicas    int32 = 2
	)
	archAndCRInfraNodeSelector := instance.Spec.InfraNodeSelector
	if archAndCRInfraNodeSelector == nil {
		archAndCRInfraNodeSelector = archOnMasterNodeSelector
	} else {
		archAndCRInfraNodeSelector["kubernetes.io/arch"] = goruntime.GOARCH
	}

	infraTolerations := instance.Spec.InfraTolerations
	if infraTolerations == nil {
		infraTolerations = []corev1.Toleration{masterExistsNoScheduleToleration}
	}
	infraAffinity := instance.Spec.InfraAffinity
	if infraAffinity == nil {
		infraAffinity = &corev1.Affinity{}
	}

	selfSignConfiguration := instance.Spec.SelfSignConfiguration
	if selfSignConfiguration == nil {
		selfSignConfiguration = &nmstatev1.SelfSignConfiguration{
			CARotateInterval:    "8760h0m0s",
			CAOverlapInterval:   "24h0m0s",
			CertRotateInterval:  "4380h0m0s",
			CertOverlapInterval: "24h0m0s",
		}
	}

	data.Data["HandlerNamespace"] = os.Getenv("HANDLER_NAMESPACE")
	data.Data["HandlerImage"] = os.Getenv("RELATED_IMAGE_HANDLER_IMAGE")
	data.Data["HandlerPullPolicy"] = os.Getenv("HANDLER_IMAGE_PULL_POLICY")
	data.Data["HandlerPrefix"] = os.Getenv("HANDLER_PREFIX")
	data.Data["InfraNodeSelector"] = archAndCRInfraNodeSelector
	data.Data["InfraTolerations"] = infraTolerations
	data.Data["WebhookAffinity"] = infraAffinity
	data.Data["WebhookReplicas"] = webhookReplicas
	data.Data["WebhookMinReplicas"] = webhookMinReplicas
	data.Data["HandlerNodeSelector"] = archAndCRNodeSelector
	data.Data["HandlerTolerations"] = handlerTolerations
	data.Data["HandlerAffinity"] = handlerAffinity
	data.Data["SelfSignConfiguration"] = selfSignConfiguration

	return r.renderAndApply(instance, data, "handler", true)
}

func (r *NMStateReconciler) renderAndApply(
	instance *nmstatev1.NMState,
	data render.RenderData,
	sourceDirectory string,
	setControllerReference bool,
) error {
	var err error

	sourceFullDirectory := filepath.Join(names.ManifestDir, "kubernetes-nmstate", sourceDirectory)
	objs, err := render.RenderDir(sourceFullDirectory, &data)
	if err != nil {
		return errors.Wrapf(err, "failed to render kubernetes-nmstate %s", sourceDirectory)
	}

	// If no file found in directory - return error
	if len(objs) == 0 {
		return fmt.Errorf("no manifests rendered from %s", sourceFullDirectory)
	}

	for _, obj := range objs {
		// RenderDir seems to add an extra null entry to the list. It appears to be because of the
		// nested templates. This just makes sure we don't try to apply an empty obj.
		if obj.GetName() == "" {
			continue
		}
		if setControllerReference {
			// Set the controller reference. When the CR is removed, it will remove the CRDs as well
			err = controllerutil.SetControllerReference(instance, obj, r.Scheme)
			if err != nil {
				return errors.Wrap(err, "failed to set owner reference")
			}
		}

		// Now apply the object
		err = apply.ApplyObject(context.TODO(), r.Client, obj)
		if err != nil {
			return errors.Wrapf(err, "failed to apply object %v", obj)
		}
	}
	return nil
}
