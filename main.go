/*

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
	"flag"
	"os"

	cndev1alpha1 "cnde-operator.cloud-native-coding.dev/api/v1alpha1"
	"cnde-operator.cloud-native-coding.dev/controllers"
	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	// +kubebuilder:scaffold:imports
)

var (
	scheme   = runtime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")
)

func init() {
	_ = clientgoscheme.AddToScheme(scheme)

	_ = cndev1alpha1.AddToScheme(scheme)
	// +kubebuilder:scaffold:scheme
}

func main() {
	var metricsAddr string
	var enableLeaderElection bool
	flag.StringVar(&metricsAddr, "metrics-addr", ":8080", "The address the metric endpoint binds to.")
	flag.BoolVar(&enableLeaderElection, "enable-leader-election", false,
		"Enable leader election for controller manager. Enabling this will ensure there is only one active controller manager.")
	flag.Parse()

	ctrl.SetLogger(zap.New(func(o *zap.Options) {
		o.Development = true
	}))

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:             scheme,
		MetricsBindAddress: metricsAddr,
		Port:               9443,
		LeaderElection:     enableLeaderElection,
		LeaderElectionID:   "b07929c2.cnde-operator.cloud-native-coding.dev",
	})
	if err != nil {
		setupLog.Error(err, "unable to start manager")
		os.Exit(1)
	}

	if err = (&controllers.DevEnvReconciler{
		Client: mgr.GetClient(),
		Log:    ctrl.Log.WithName("controllers").WithName("DevEnv"),
		Scheme: mgr.GetScheme(),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "DevEnv")
		os.Exit(1)
	}

	setupLog.Info("starting manager with the following settings:",
		"Oauth Admin Name", os.Getenv("CNDE_OAUTH_ADMIN_NAME"), "Oauth Admin Realm", os.Getenv("CNDE_OAUTH_ADMIN_REALM"),
		"Oauth URL", os.Getenv("CNDE_OAUTH_URL"), "Manager Namespace", os.Getenv("CNDE_MANAGER_NAMESPACE"))

	if _, b := os.LookupEnv("CNDE_OAUTH_ADMIN_PASSWORD"); b {
		setupLog.Info("CNDE_OAUTH_ADMIN_PASSWORD is set")
	} else {
		setupLog.Info("CNDE_OAUTH_ADMIN_PASSWORD unset, but should be provided")
	}

	if _, b := os.LookupEnv("CNDE_OAUTH_INITIAL_PW"); b {
		setupLog.Info("CNDE_OAUTH_INITIAL_PW is set")
	} else {
		setupLog.Info("CNDE_OAUTH_INITIAL_PW unset, but should be provided")
	}

	// +kubebuilder:scaffold:builder

	setupLog.Info("starting manager")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		setupLog.Error(err, "problem running manager")
		os.Exit(1)
	}
}
