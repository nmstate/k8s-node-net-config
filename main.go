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

package main

import (
	"bytes"
	"flag"
	"fmt"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/exec"
	"time"

	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	// +kubebuilder:scaffold:imports

	"github.com/kelseyhightower/envconfig"
	"github.com/nightlyone/lockfile"
	"github.com/pkg/errors"
	"github.com/qinqon/kube-admission-webhook/pkg/certificate"
	"k8s.io/apimachinery/pkg/util/wait"

	nmstatev1alpha1 "github.com/nmstate/kubernetes-nmstate/api/v1alpha1"
	nmstatev1beta1 "github.com/nmstate/kubernetes-nmstate/api/v1beta1"
	"github.com/nmstate/kubernetes-nmstate/controllers"
	"github.com/nmstate/kubernetes-nmstate/pkg/environment"
	"github.com/nmstate/kubernetes-nmstate/pkg/webhook"
)

const unmanagedVethCommand = "unmanaged-veth"

type ProfilerConfig struct {
	EnableProfiler bool   `envconfig:"ENABLE_PROFILER"`
	ProfilerPort   string `envconfig:"PROFILER_PORT" default:"6060"`
}

var (
	scheme   = runtime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")
)

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))

	utilruntime.Must(nmstatev1beta1.AddToScheme(scheme))
	utilruntime.Must(nmstatev1alpha1.AddToScheme(scheme))
	// +kubebuilder:scaffold:scheme
}

func main() {
	var logType string
	flag.StringVar(&logType, "v", "production", "Log type (debug/production).")
	flag.Parse()

	ctrl.SetLogger(zap.New(zap.UseDevMode(logType != "production")))

	// Lock only for handler, we can run old and new version of
	// webhook without problems, policy status will be updated
	// by multiple instances.
	if environment.IsHandler() {
		handlerLock, err := lockHandler()
		if err != nil {
			setupLog.Error(err, "Failed to run lockHandler")
			os.Exit(1)
		}
		defer handlerLock.Unlock()
		setupLog.Info("Successfully took nmstate exclusive lock")

		setVethInterfacesAsUnmanaged()
	}

	ctrlOptions := ctrl.Options{
		Scheme:             scheme,
		MetricsBindAddress: "0", // disable metrics
	}

	// We need to add LeaerElection for the webhook
	// cert-manager the LeaderElectionID was generated by operator-sdk
	if environment.IsWebhook() {
		ctrlOptions.LeaderElection = true
		ctrlOptions.LeaderElectionID = "5d2e944a.nmstate.io"
	}

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrlOptions)
	if err != nil {
		setupLog.Error(err, "unable to start manager")
		os.Exit(1)
	}

	// Runs only webhook controllers if it's specified
	if environment.IsWebhook() {

		webhookOpts := certificate.Options{
			Namespace:   os.Getenv("POD_NAMESPACE"),
			WebhookName: "nmstate",
			WebhookType: certificate.MutatingWebhook,
		}

		webhookOpts.CARotateInterval, err = environment.LookupAsDuration("CA_ROTATE_INTERVAL")
		if err != nil {
			setupLog.Error(err, "Failed retrieving ca rotate interval")
			os.Exit(1)
		}

		webhookOpts.CAOverlapInterval, err = environment.LookupAsDuration("CA_OVERLAP_INTERVAL")
		if err != nil {
			setupLog.Error(err, "Failed retrieving ca overlap interval")
			os.Exit(1)
		}

		webhookOpts.CertRotateInterval, err = environment.LookupAsDuration("CERT_ROTATE_INTERVAL")
		if err != nil {
			setupLog.Error(err, "Failed retrieving cert rotate interval")
			os.Exit(1)
		}

		if err := webhook.AddToManager(mgr, webhookOpts); err != nil {
			setupLog.Error(err, "Cannot initialize webhook")
			os.Exit(1)
		}
	} else if environment.IsOperator() {
		if err = (&controllers.NMStateReconciler{
			Client: mgr.GetClient(),
			Log:    ctrl.Log.WithName("controllers").WithName("NMState"),
			Scheme: mgr.GetScheme(),
		}).SetupWithManager(mgr); err != nil {
			setupLog.Error(err, "unable to create NMState controller", "controller", "NMState")
			os.Exit(1)
		}
	} else if environment.IsHandler() {
		if err = (&controllers.NodeReconciler{
			Client: mgr.GetClient(),
			Log:    ctrl.Log.WithName("controllers").WithName("Node"),
			Scheme: mgr.GetScheme(),
		}).SetupWithManager(mgr); err != nil {
			setupLog.Error(err, "unable to create Node controller", "controller", "NMState")
			os.Exit(1)
		}
		if err = (&controllers.NodeNetworkConfigurationPolicyReconciler{
			Client: mgr.GetClient(),
			Log:    ctrl.Log.WithName("controllers").WithName("NodeNetworkConfigurationPolicy"),
			Scheme: mgr.GetScheme(),
		}).SetupWithManager(mgr); err != nil {
			setupLog.Error(err, "unable to create NodeNetworkConfigurationPolicy controller", "controller", "NMState")
			os.Exit(1)
		}
		if err = (&controllers.NodeNetworkStateReconciler{
			Client: mgr.GetClient(),
			Log:    ctrl.Log.WithName("controllers").WithName("NodeNetworkState"),
			Scheme: mgr.GetScheme(),
		}).SetupWithManager(mgr); err != nil {
			setupLog.Error(err, "unable to create NodeNetworkState controller", "controller", "NMState")
			os.Exit(1)
		}
	}

	setProfiler()

	setupLog.Info("starting manager")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		setupLog.Error(err, "problem running manager")
		os.Exit(1)
	}
}

// Start profiler on given port if ENABLE_PROFILER is True
func setProfiler() {
	cfg := ProfilerConfig{}
	envconfig.Process("", &cfg)
	if cfg.EnableProfiler {
		setupLog.Info("Starting profiler")
		go func() {
			profilerAddress := fmt.Sprintf("0.0.0.0:%s", cfg.ProfilerPort)
			setupLog.Info(fmt.Sprintf("Starting Profiler Server! \t Go to http://%s/debug/pprof/\n", profilerAddress))
			err := http.ListenAndServe(profilerAddress, nil)
			if err != nil {
				setupLog.Info("Failed to start the server! Error: %v", err)
			}
		}()
	}
}

func lockHandler() (lockfile.Lockfile, error) {
	lockFilePath, ok := os.LookupEnv("NMSTATE_INSTANCE_NODE_LOCK_FILE")
	if !ok {
		return "", errors.New("Failed to find NMSTATE_INSTANCE_NODE_LOCK_FILE ENV var")
	}
	setupLog.Info(fmt.Sprintf("Try to take exclusive lock on file: %s", lockFilePath))
	handlerLock, err := lockfile.New(lockFilePath)
	if err != nil {
		return handlerLock, errors.Wrapf(err, "failed to create lockFile for %s", lockFilePath)
	}
	err = wait.PollImmediateInfinite(5*time.Second, func() (done bool, err error) {
		err = handlerLock.TryLock()
		if err != nil {
			setupLog.Error(err, "retrying to lock handler")
			return false, nil // Don't return the error here, it will not re-poll if we do
		}
		return true, nil
	})
	return handlerLock, err
}

func setVethInterfacesAsUnmanaged() {
	cmd := exec.Command(unmanagedVethCommand)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		setupLog.Info(fmt.Sprintf("failed to execute %s: '%v', '%s', '%s'", unmanagedVethCommand, err, stdout.String(), stderr.String()))
	}
}
