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

package controllers

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	extv1beta1 "k8s.io/api/extensions/v1beta1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"cnde-operator.cloud-native-coding.dev/api/v1alpha1"
	cndev1alpha1 "cnde-operator.cloud-native-coding.dev/api/v1alpha1"
	"cnde-operator.cloud-native-coding.dev/oauth"
	"cnde-operator.cloud-native-coding.dev/oauth/keycloak"
	apierrs "k8s.io/apimachinery/pkg/api/errors"
)

const (
	finalizerName = "realm.devenvs.c-n-d-e.kube-platform.dev"

	// used by Reconciler to identify where to find CR
	namespaceLabel   = "user-env-ns"
	userenvnameLabel = "user-env-name"
	imageTagName     = "IMAGE_TAG"
)

// DevEnvReconciler reconciles a DevEnv object
type DevEnvReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme

	serviceAccountName string
	resourceName       string
	buildName          string
	initName           string
	homeVolumeName     string
	vmVolumeName       string
	dockerVolumeName   string
	dockerVolumeSize   string
	homeVolumeSize     string
	ingressHost        string
	ingressUIName      string
	ingressOauthName   string
	proxyPodName       string
	DevEnvPodIP        string
	DevEnvNamespace    string
	ManagerNamespace   string

	dockerImg     string
	devEnvImg     string
	kubeConfigImg string
	configureImg  string
	oauthProxyImg string
	alpineImage   string

	memRequestIDE    resource.Quantity
	memRequestDocker resource.Quantity

	oauth             oauth.OAUTHProvider
	oauthClientSecret string
	oauthClientID     string

	oauthProviderName string

	hasBuilder bool
}

func ignoreNotFound(err error) error {
	if apierrs.IsNotFound(err) {
		return nil
	}
	return err
}

// +kubebuilder:rbac:groups=c-n-d-e.kube-platform.dev,resources=devenvs,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=c-n-d-e.kube-platform.dev,resources=devenvs/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=c-n-d-e.kube-platform.dev,resources=builders,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="",resources=namespaces,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="",resources=pods,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="",resources=serviceaccounts,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="rbac.authorization.k8s.io",resources=rolebindings,verbs=*
// +kubebuilder:rbac:groups="rbac.authorization.k8s.io",resources=clusterrolebindings,verbs=*

func (r *DevEnvReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	_ = r.Log.WithValues("DevEnv", req.NamespacedName)

	// Fetch the DevEnv instance
	devenv := &cndev1alpha1.DevEnv{}
	err := r.Get(ctx, req.NamespacedName, devenv)
	if err != nil {
		return ctrl.Result{}, ignoreNotFound(err)
	}

	r.initStruct(devenv)

	// --------------------------------------------
	// Check if the APP CR was marked to be deleted
	// --------------------------------------------
	isUserEnvMarkedToBeDeleted := devenv.GetDeletionTimestamp() != nil
	if isUserEnvMarkedToBeDeleted {

		err = r.oauth.DeleteRealm(devenv)
		if err != nil && !errors.IsNotFound(err) {
			return ctrl.Result{}, err
		}

		buildPod := &corev1.Pod{}
		if err = r.Get(ctx, types.NamespacedName{Name: r.buildName, Namespace: r.ManagerNamespace}, buildPod); err == nil {
			r.Log.Info("Deleting Build Pod")
			err = r.Delete(ctx, buildPod)
			if err != nil {
				r.Log.Error(err, "Failed to delete Build Pod")
			}
		}

		devenv.SetFinalizers(nil)
		err := r.Update(ctx, devenv)
		if err != nil {
			return ctrl.Result{}, err
		}
		return ctrl.Result{}, nil
	}

	r.hasBuilder = false
	builder := &cndev1alpha1.Builder{}
	if devenv.Spec.BuilderName != "" {
		err = r.Get(ctx, types.NamespacedName{Name: devenv.Spec.BuilderName, Namespace: r.ManagerNamespace}, builder)
		if err != nil {
			if !errors.IsNotFound(err) {
				return ctrl.Result{}, err
			}
			if devenv.Spec.BuilderName != "" {
				return ctrl.Result{}, fmt.Errorf("BuilderName: %v configured but not found", devenv.Spec.BuilderName)
			}
		} else {
			r.hasBuilder = true
		}
	}

	// -----------------------------------
	// keycloak stuff
	// -----------------------------------

	if devenv.Status.Realm == "" {
		name, err := r.oauth.NewRealm(devenv)
		if err != nil {
			r.Log.Error(err, "Failed to create new Realm.")
			return ctrl.Result{}, err
		}

		devenv.Status.Realm = name
		err = r.Status().Update(ctx, devenv)
		if err != nil {
			r.Log.Error(err, "Failed to update User Environment Status")
			return ctrl.Result{}, err
		}

		devenv.SetFinalizers([]string{finalizerName})
		err = r.Update(ctx, devenv)
		if err != nil {
			r.Log.Error(err, "Failed to update User Environment Finalizers")
			return ctrl.Result{}, err
		}
	}

	r.oauthClientSecret, err = r.oauth.CreateClient(devenv)
	if err != nil {
		return ctrl.Result{}, err
	}
	if r.oauthClientSecret == "" {
		return ctrl.Result{RequeueAfter: 2 * time.Second}, nil // some time for keycloak
	}

	if devenv.Status.User == "" {
		name, err := r.oauth.CreateUser(devenv)
		if err != nil {
			r.Log.Error(err, "Failed to create new User.")
			return ctrl.Result{}, err
		}

		devenv.Status.User = name
		err = r.Status().Update(ctx, devenv)
		if err != nil {
			r.Log.Error(err, "Failed to update User Environment User")
			return ctrl.Result{}, err
		}
		return ctrl.Result{Requeue: true}, nil // wait for status udpate
	}

	// --------------------------------------------
	// --------------------------------------------

	namespace := &corev1.Namespace{}
	err = r.Get(ctx, types.NamespacedName{Name: r.resourceName}, namespace)
	if err != nil && errors.IsNotFound(err) {
		ns := r.createNamespaceForDevEnv(devenv)
		r.Log.Info("Creating a new Namespace.", "Namespace.Name", ns.Name)
		err = r.Create(ctx, ns)
		if err != nil {
			if errors.IsForbidden(err) {
				r.Log.Info("Forbidden to create Namespace, may be terminating? - waiting")
				return ctrl.Result{RequeueAfter: 2 * time.Second}, nil
			}
			r.Log.Error(err, "Failed to create new Namespace.", "Namespace.Name", ns.Name)
			return ctrl.Result{}, err
		}
		// resetting status if new Namespace
		if r, err := r.setDevEnvStatus(ctx, devenv, v1alpha1.BuildPhaseInitial); err != nil {
			return r, err
		}
		return ctrl.Result{RequeueAfter: 5 * time.Second}, nil
	} else if err != nil {
		r.Log.Error(err, "Failed to get Namespace.")
		return ctrl.Result{}, err
	}

	serviceaccount := &corev1.ServiceAccount{}
	err = r.Get(ctx, types.NamespacedName{Name: r.serviceAccountName, Namespace: r.DevEnvNamespace}, serviceaccount)
	if err != nil && errors.IsNotFound(err) {
		sa := r.serviceAccountForDevEnv(devenv)
		r.Log.Info("Creating a new ServiceAccount.", "ServiceAccount.Namespace", sa.Namespace, "ServiceAccount.Name", sa.Name)
		err = r.Create(ctx, sa)
		if err != nil {
			r.Log.Error(err, "Failed to create new ServiceAccount.", "ServiceAccount.Namespace", sa.Namespace, "ServiceAccount.Name", sa.Name)
			return ctrl.Result{}, err
		}
	} else if err != nil {
		r.Log.Error(err, "Failed to get ServiceAccount.")
		return ctrl.Result{}, err
	}

	roleBinding := &rbacv1.RoleBinding{}
	err = r.Get(ctx, types.NamespacedName{Name: r.resourceName, Namespace: r.DevEnvNamespace}, roleBinding)
	if err != nil && errors.IsNotFound(err) {
		rb := r.rbacRBForDevEnv(devenv)
		r.Log.Info("Creating a new rb for User.", "rb.Namespace", rb.Namespace, "rb.Name", rb.Name)
		err = r.Create(ctx, rb)
		if err != nil {
			r.Log.Error(err, "Failed to create new rb.", "rb.Namespace", rb.Namespace, "rb.Name", rb.Name)
			return ctrl.Result{}, err
		}
	} else if err != nil {
		r.Log.Error(err, "Failed to get rb.")
		return ctrl.Result{}, err
	}

	clusterRoleBinding := &rbacv1.ClusterRoleBinding{}
	err = r.Get(ctx, types.NamespacedName{Name: r.resourceName}, clusterRoleBinding)
	if err != nil && errors.IsNotFound(err) {
		crb := r.rbacCRBForDevEnv(devenv)
		r.Log.Info("Creating a new crb for User.", "crb.Namespace", crb.Namespace, "crb.Name", crb.Name)
		err = r.Create(ctx, crb)
		if err != nil {
			r.Log.Error(err, "Failed to create new crb.", "crb.Namespace", crb.Namespace, "crb.Name", crb.Name)
			return ctrl.Result{}, err
		}
	} else if err != nil {
		r.Log.Error(err, "Failed to get crb.")
		return ctrl.Result{}, err
	}

	persistenceVM := &corev1.PersistentVolumeClaim{}
	err = r.Get(ctx, types.NamespacedName{Name: r.vmVolumeName, Namespace: r.DevEnvNamespace}, persistenceVM)
	if err != nil && errors.IsNotFound(err) {
		pvcVM := r.pvcVMForDevEnv(devenv)
		r.Log.Info("Creating a new VM pvc.", "pvc.Namespace", pvcVM.Namespace, "pvc.Name", pvcVM.Name)
		err = r.Create(ctx, pvcVM)
		if err != nil {
			r.Log.Error(err, "Failed to create new VM pvc.", "pvc.Namespace", pvcVM.Namespace, "pvc.Name", pvcVM.Name)
			return ctrl.Result{}, err
		}
	} else if err != nil {
		r.Log.Error(err, "Failed to get VM pvc.")
		return ctrl.Result{}, err
	}

	persistenceHome := &corev1.PersistentVolumeClaim{}
	err = r.Get(ctx, types.NamespacedName{Name: r.homeVolumeName, Namespace: r.DevEnvNamespace}, persistenceHome)
	if err != nil && errors.IsNotFound(err) {
		pvcHome := r.pvcHomeForDevEnv(devenv)
		r.Log.Info("Creating a new Home pvc.", "pvc.Namespace", pvcHome.Namespace, "pvc.Name", pvcHome.Name)
		err = r.Create(ctx, pvcHome)
		if err != nil {
			r.Log.Error(err, "Failed to create new Home pvc.", "pvc.Namespace", pvcHome.Namespace, "pvc.Name", pvcHome.Name)
			return ctrl.Result{}, err
		}
	} else if err != nil {
		r.Log.Error(err, "Failed to get Home pvc.")
		return ctrl.Result{}, err
	}

	persistenceDocker := &corev1.PersistentVolumeClaim{}
	err = r.Get(ctx, types.NamespacedName{Name: r.dockerVolumeName, Namespace: r.DevEnvNamespace}, persistenceDocker)
	if err != nil && errors.IsNotFound(err) {
		pvcDocker := r.pvcDockerForDevEnv(devenv)
		r.Log.Info("Creating a new Docker pvc.", "pvc.Namespace", pvcDocker.Namespace, "pvc.Name", pvcDocker.Name)
		err = r.Create(ctx, pvcDocker)
		if err != nil {
			r.Log.Error(err, "Failed to create new Docker pvc.", "pvc.Namespace", pvcDocker.Namespace, "pvc.Name", pvcDocker.Name)
			return ctrl.Result{}, err
		}
	} else if err != nil {
		r.Log.Error(err, "Failed to get Docker pvc.")
		return ctrl.Result{}, err
	}

	/**
	*** Processing Build
	**/

	if r.hasBuilder {
		switch devenv.Status.Build {

		case v1alpha1.BuildPhaseInitial:
			buildPod := r.buildPodForDevEnv(devenv, builder)
			err = r.Create(ctx, buildPod)
			if err != nil {
				if errors.IsAlreadyExists(err) {
					r.Log.Info("Build Pod already there, trying to delete it")
					return ctrl.Result{RequeueAfter: 5 * time.Second}, r.Delete(ctx, buildPod)
				}
				r.Log.Error(err, "Failed to create Build Pod.")
				return ctrl.Result{}, err
			}

			if r, err := r.setDevEnvStatus(ctx, devenv, v1alpha1.BuildPhaseBuilding); err != nil {
				return r, err
			}
			return ctrl.Result{RequeueAfter: 10 * time.Second}, nil

		case v1alpha1.BuildPhaseBuilding:
			buildPod := &corev1.Pod{}
			err = r.Get(ctx, types.NamespacedName{Name: r.buildName, Namespace: r.ManagerNamespace}, buildPod)
			if err != nil {
				r.Log.Error(err, "Failed to find Build Pod in state Building. Resetting DevEnv status")
				if r, err := r.setDevEnvStatus(ctx, devenv, v1alpha1.BuildPhaseInitial); err != nil {
					return r, err
				}
				return ctrl.Result{RequeueAfter: 5 * time.Second}, nil
			}

			switch buildPod.Status.Phase {
			case corev1.PodSucceeded:
				if r, err := r.setDevEnvStatus(ctx, devenv, v1alpha1.BuildPhaseWaitForInitializion); err != nil {
					return r, err
				}
				err = r.Delete(ctx, buildPod)
				if err != nil {
					r.Log.Error(err, "Failed to Delete Build Pod.")
					return ctrl.Result{}, err
				}
				r.Log.Info("Build succeeded, creating a new DevEnv Pod.")
			case corev1.PodFailed:
				r.Log.Info("Build failed")
				return ctrl.Result{}, nil
			}
			return ctrl.Result{RequeueAfter: 5 * time.Second}, nil // wait for build POD to finish
		}
	} else {
		switch devenv.Status.Build {
		case v1alpha1.BuildPhaseInitial:
			if r, err := r.setDevEnvStatus(ctx, devenv, v1alpha1.BuildPhaseWaitForInitializion); err != nil {
				return r, err
			}
			return ctrl.Result{Requeue: true}, nil
		}
	}

	/**
	*** Initialization
	**/

	switch devenv.Status.Build {

	case v1alpha1.BuildPhaseWaitForInitializion:
		initPod := r.podForInitializingDevEnv(devenv)
		err = r.Create(ctx, initPod)
		if err != nil {
			if errors.IsAlreadyExists(err) {
				r.Log.Info("Initializing Pod already there, trying to delete it")
				return ctrl.Result{RequeueAfter: 5 * time.Second}, r.Delete(ctx, initPod)
			}
			r.Log.Error(err, "Failed to create Initializing Pod.")
			return ctrl.Result{}, err
		}
		if r, err := r.setDevEnvStatus(ctx, devenv, v1alpha1.BuildPhaseInitializing); err != nil {
			return r, err
		}
		return ctrl.Result{RequeueAfter: 2 * time.Second}, nil

	case v1alpha1.BuildPhaseInitializing:
		initPod := &corev1.Pod{}
		err = r.Get(ctx, types.NamespacedName{Name: r.initName, Namespace: r.DevEnvNamespace}, initPod)
		if err != nil {
			r.Log.Error(err, "Failed to find Initialization Pod in state Initializing. Resetting DevEnv status")
			if r, err := r.setDevEnvStatus(ctx, devenv, v1alpha1.BuildPhaseWaitForInitializion); err != nil {
				return r, err
			}
			return ctrl.Result{RequeueAfter: 2 * time.Second}, nil
		}
		switch initPod.Status.Phase {
		case corev1.PodSucceeded:
			if r, err := r.setDevEnvStatus(ctx, devenv, v1alpha1.BuildPhaseRunning); err != nil {
				return r, err
			}
			err = r.Delete(ctx, initPod)
			if err != nil {
				r.Log.Error(err, "Failed to Delete Initialization Pod.")
				return ctrl.Result{}, err
			}
			r.Log.Info("Initialization succeeded, creating a new DevEnv Pod.")
		case corev1.PodFailed:
			r.Log.Info("Build failed")
			return ctrl.Result{Requeue: false}, nil
		}
		return ctrl.Result{RequeueAfter: 5 * time.Second}, nil
	}

	found := &corev1.Pod{}
	err = r.Get(ctx, types.NamespacedName{Name: r.resourceName, Namespace: r.DevEnvNamespace}, found)
	if err != nil && errors.IsNotFound(err) {
		pod := r.podForDevEnv(devenv)
		r.Log.Info("Creating a new DevEnv Pod.", "Pod.Namespace", pod.Namespace, "Pod.Name", pod.Name)
		err = r.Create(ctx, pod)
		if err != nil {
			return ctrl.Result{}, err
		}
		// Requeueing because of PodIP, that is needed below
		return ctrl.Result{Requeue: true}, nil
	} else if err != nil {
		r.Log.Error(err, "Failed to get DevEnv Pod.")
		return ctrl.Result{}, err
	}

	// check if POD is created and has an IP that is needed for creating an instance of Endpoint
	r.DevEnvPodIP = found.Status.PodIP
	if r.DevEnvPodIP == "" {
		return ctrl.Result{Requeue: true}, nil
	}

	proxyService := &corev1.Service{}
	err = r.Get(ctx, types.NamespacedName{Name: r.resourceName, Namespace: r.ManagerNamespace}, proxyService)
	if err != nil && errors.IsNotFound(err) {
		ser := r.serviceProxyForDevEnv(devenv)
		r.Log.Info("Creating a new Proxy Service.", "Service.Namespace", ser.Namespace, "Service.Name", ser.Name)
		err = r.Create(ctx, ser)
		if err != nil {
			r.Log.Error(err, "Failed to create new Service.", "Service.Namespace", ser.Namespace, "Service.Name", ser.Name)
			return ctrl.Result{}, err
		}
	} else if err != nil {
		r.Log.Error(err, "Failed to get Service.")
		return ctrl.Result{}, err
	}

	// TODO Endpoint has to be recreated if the Pod is reconciled
	// **********************************************************

	proxyEndpoint := &corev1.Endpoints{}
	err = r.Get(ctx, types.NamespacedName{Name: r.resourceName, Namespace: r.ManagerNamespace}, proxyEndpoint)
	if err != nil && errors.IsNotFound(err) {
		ep := r.endpointProxyForDevEnv(devenv)
		r.Log.Info("Creating a new Proxy Endpoint.", "Endpoint.Namespace", ep.Namespace, "Endpoint.Name", ep.Name)
		err = r.Create(ctx, ep)
		if err != nil {
			r.Log.Error(err, "Failed to create new Proxy Endpoint.", "Service.Endpoint", ep.Namespace, "Endpoint.Name", ep.Name)
			return ctrl.Result{}, err
		}
	} else if err != nil {
		r.Log.Error(err, "Failed to get Proxy Endpoint.")
		return ctrl.Result{}, err
	} else if proxyEndpoint.Subsets[0].Addresses[0].IP != r.DevEnvPodIP {
		ep := r.endpointProxyForDevEnv(devenv)
		r.Log.Info("Updating Proxy Endpoint.", "Endpoint.Namespace", ep.Namespace, "Endpoint.Name", ep.Name)
		err = r.Update(ctx, ep)
		if err != nil {
			r.Log.Error(err, "Failed to update Proxy Endpoint.")
		}
	}

	uiIngress := &extv1beta1.Ingress{}
	err = r.Get(ctx, types.NamespacedName{Name: r.ingressUIName, Namespace: r.ManagerNamespace}, uiIngress)
	if err != nil && errors.IsNotFound(err) {
		ingUI := r.ingressUIForDevEnv(devenv)
		r.Log.Info("Creating a new UI Ingress.", "Ingress.Namespace", ingUI.Namespace, "Ingress.Name", ingUI.Name)
		err = r.Create(ctx, ingUI)
		if err != nil {
			r.Log.Error(err, "Failed to create new UI Ingress.", "Ingress.Namespace", ingUI.Namespace, "Ingress.Name", ingUI.Name)
			return ctrl.Result{}, err
		}
	} else if err != nil {
		r.Log.Error(err, "Failed to get UI Ingress.")
		return ctrl.Result{}, err
	}

	oauthIngress := &extv1beta1.Ingress{}
	err = r.Get(ctx, types.NamespacedName{Name: r.ingressOauthName, Namespace: r.ManagerNamespace}, oauthIngress)
	if err != nil && errors.IsNotFound(err) {
		ingOauth := r.ingressOauthForDevEnv(devenv)
		r.Log.Info("Creating a new OAUTH Ingress.", "Ingress.Namespace", ingOauth.Namespace, "Ingress.Name", ingOauth.Name)
		err = r.Create(ctx, ingOauth)
		if err != nil {
			r.Log.Error(err, "Failed to create new OAUTH Ingress.", "Ingress.Namespace", ingOauth.Namespace, "Ingress.Name", ingOauth.Name)
			return ctrl.Result{}, err
		}
	} else if err != nil {
		r.Log.Error(err, "Failed to get OAUTH Ingress.")
		return ctrl.Result{}, err
	}

	proxyPod := &corev1.Pod{}
	err = r.Get(ctx, types.NamespacedName{Name: r.proxyPodName, Namespace: r.ManagerNamespace}, proxyPod)
	if err != nil && errors.IsNotFound(err) {
		ppod := r.podOauthProxyForDevEnv(devenv)
		r.Log.Info("Creating a new OAUTH Proxy Pod.", "Pod.Namespace", ppod.Namespace, "Pod.Name", ppod.Name)
		err = r.Create(ctx, ppod)
		if err != nil {
			r.Log.Error(err, "Failed to create new OAUTH Proxy Pod.", "Pod.Namespace", ppod.Namespace, "Pod.Name", ppod.Name)
			return ctrl.Result{}, err
		}
	} else if err != nil {
		r.Log.Error(err, "Failed to get OAUTH Proxy Pod.")
		return ctrl.Result{}, err
	}

	proxySecret := &corev1.Secret{}
	err = r.Get(ctx, types.NamespacedName{Name: r.proxyPodName, Namespace: r.ManagerNamespace}, proxySecret)
	if err != nil && errors.IsNotFound(err) {
		proxysec := r.secretOauthProxyForDevEnv(devenv)
		r.Log.Info("Creating a new OAUTH Proxy Secret.", "Secret.Namespace", proxysec.Namespace, "Secret.Name", proxysec.Name)
		err = r.Create(ctx, proxysec)
		if err != nil {
			r.Log.Error(err, "Failed to create new OAUTH Proxy Secret.", "Secret.Namespace", proxysec.Namespace, "Secret.Name", proxysec.Name)
			return ctrl.Result{}, err
		}
	} else if err != nil {
		r.Log.Error(err, "Failed to get OAUTH Proxy Secret.")
		return ctrl.Result{}, err
	}

	oauthProxyService := &corev1.Service{}
	err = r.Get(ctx, types.NamespacedName{Name: r.proxyPodName, Namespace: r.ManagerNamespace}, oauthProxyService)
	if err != nil && errors.IsNotFound(err) {
		psrv := r.serviceOauthProxyForDevEnv(devenv)
		r.Log.Info("Creating a new OAUTH Proxy Service.", "Service.Namespace", psrv.Namespace, "Service.Name", psrv.Name)
		err = r.Create(ctx, psrv)
		if err != nil {
			r.Log.Error(err, "Failed to create new OAUTH Proxy Service.", "Service.Namespace", psrv.Namespace, "Service.Name", psrv.Name)
			return ctrl.Result{}, err
		}
	} else if err != nil {
		r.Log.Error(err, "Failed to get OAUTH Proxy Service.")
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

func (r *DevEnvReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&cndev1alpha1.DevEnv{}).
		Owns(&corev1.Pod{}).
		Owns(&corev1.Service{}).
		Owns(&corev1.Endpoints{}).
		Owns(&corev1.Secret{}).
		Owns(&corev1.ServiceAccount{}).
		Owns(&corev1.PersistentVolumeClaim{}).
		Owns(&corev1.Namespace{}).
		Owns(&rbacv1.RoleBinding{}).
		Owns(&rbacv1.ClusterRoleBinding{}).
		Owns(&extv1beta1.Ingress{}).
		Complete(r)
}

// init local structure
func (r *DevEnvReconciler) initStruct(userenv *cndev1alpha1.DevEnv) {

	r.ManagerNamespace = os.Getenv("CNDE_MANAGER_NAMESPACE")

	r.serviceAccountName = "cnde"
	r.resourceName = "cnde-" + userenv.Name
	r.ingressHost = userenv.Name + "." + userenv.Spec.UserEnvDomain

	if subDomain, exists := os.LookupEnv("CNDE_SUBDOMAIN"); exists {
		r.resourceName = "cnde-" + subDomain + "-" + userenv.Name
		r.ingressHost = userenv.Name + "." + subDomain + "." + userenv.Spec.UserEnvDomain
	}

	if memRequestIDE, exists := os.LookupEnv("CNDE_IDE_MEM_REQUEST"); exists {
		r.memRequestIDE = resource.MustParse(memRequestIDE)
	} else {
		r.memRequestIDE = resource.MustParse("512Mi")
	}

	if memRequestDocker, exists := os.LookupEnv("CNDE_DOCKER_MEM_REQUEST"); exists {
		r.memRequestDocker = resource.MustParse(memRequestDocker)
	} else {
		r.memRequestDocker = resource.MustParse("512Mi")
	}

	if oauthProviderName, exists := os.LookupEnv("CNDE_OAUTH_PROVIDERNAME"); exists {
		r.oauthProviderName = oauthProviderName
	} else {
		r.oauthProviderName = "keycloak"
	}

	r.oauthClientID = "c-n-d-e"

	oauthConfig := &oauth.OAUTHProviderConfig{
		Log:                  r.Log,
		OauthClientID:        r.oauthClientID,
		OauthAdminName:       os.Getenv("CNDE_OAUTH_ADMIN_NAME"),
		OauthAdminPassword:   os.Getenv("CNDE_OAUTH_ADMIN_PASSWORD"),
		OauthAdminRealm:      os.Getenv("CNDE_OAUTH_ADMIN_REALM"),
		OauthInitialPassword: os.Getenv("CNDE_OAUTH_INITIAL_PW"),
		OauthURL:             os.Getenv("CNDE_OAUTH_URL"),
		IngressHost:          r.ingressHost,
		ResourceName:         r.resourceName,
	}

	switch r.oauthProviderName {
	case "keycloak":
		r.oauth = keycloak.NewKeycloakOAUTHProvider(oauthConfig)
	}

	r.DevEnvNamespace = r.resourceName
	r.homeVolumeName = r.resourceName + "-home-storage"
	r.vmVolumeName = r.resourceName + "-vm-storage"
	r.dockerVolumeName = r.resourceName + "-docker-storage"
	r.buildName = r.resourceName + "-build"
	r.initName = r.resourceName + "-init"

	r.proxyPodName = r.resourceName + "-oauth-proxy"

	r.ingressUIName = r.resourceName + "-ui"
	r.ingressOauthName = r.resourceName + "-oauth"

	r.dockerVolumeSize = userenv.Spec.HomeVolumeSize
	r.homeVolumeSize = userenv.Spec.DockerVolumeSize

	//r.ingressHost = userenv.Name + "." + subdomainDot + userenv.Spec.UserEnvDomain

	if r.dockerImg = userenv.Spec.DockerImg; r.dockerImg == "" {
		r.dockerImg = "docker:19-dind"
	}
	if r.devEnvImg = userenv.Spec.DevEnvImg; r.devEnvImg == "" {
		r.devEnvImg = "eu.gcr.io/cloud-native-coding/code-server-example"
	}
	if r.kubeConfigImg = userenv.Spec.KubeConfigImg; r.kubeConfigImg == "" {
		r.kubeConfigImg = "eu.gcr.io/cloud-native-coding/create-kubeconfig"
	}
	if r.configureImg = userenv.Spec.ConfigureImg; r.configureImg == "" {
		r.configureImg = "eu.gcr.io/cloud-native-coding/code-server-example"
	}
	if r.oauthProxyImg = userenv.Spec.OauthProxyImg; r.oauthProxyImg == "" {
		r.oauthProxyImg = "bitnami/oauth2-proxy:5"
	}

	r.alpineImage = "alpine:3"

	// r.Log.Info("creating DevEnv-controller with the following settings:", "DevEnv Name", userenv.Name)
	// r.Log.Info("derived names:", "resource Name", r.resourceName, "ingress Host", r.ingressHost,
	// 	"home Volume Name", r.homeVolumeName, "VM Volume Name", r.vmVolumeName, "docker Volume Name", r.dockerVolumeName,
	// 	"build Name", r.buildName, "init Name", r.initName, "proxy Pod Name", r.proxyPodName,
	// 	"ingress UI Name", r.ingressUIName, "ingress Oauth Name", r.ingressOauthName)

	// r.Log.Info("Images:", "Docker Img", r.DockerImg, "Kube Config Img", r.KubeConfigImg, "Configure Img", r.ConfigureImg,
	// 	"Oauth Proxy Img", r.OauthProxyImg, "Alpine Image", r.AlpineImage)
}

func (r *DevEnvReconciler) setDevEnvStatus(ctx context.Context, devenv *cndev1alpha1.DevEnv, phase v1alpha1.BuildPhase) (ctrl.Result, error) {
	devenv.Status.Build = phase
	err := r.Status().Update(ctx, devenv)
	if err != nil {
		r.Log.Error(err, "Failed to Update UserEnv Status to:", "devenv.Status.Build", devenv.Status.Build)
		return ctrl.Result{}, err
	}
	return ctrl.Result{}, nil
}
