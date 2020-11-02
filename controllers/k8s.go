package controllers

import (
	cndev1alpha1 "cnde-operator.cloud-native-coding.dev/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	extv1beta1 "k8s.io/api/extensions/v1beta1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

const (
	idePort  = 8080
	ttydPort = 7681
)

func labelsForDevEnv(name string) map[string]string {
	return map[string]string{userenvnameLabel: name, "app": "code-server"}
}

func (r *DevEnvReconciler) rbacCRBForDevEnv(cr *cndev1alpha1.DevEnv) *rbacv1.ClusterRoleBinding {
	labels := labelsForDevEnv(cr.Name)

	crb := &rbacv1.ClusterRoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name:   r.resourceName,
			Labels: labels,
		},
		RoleRef: rbacv1.RoleRef{
			Name:     cr.Spec.ClusterRoleName,
			APIGroup: "rbac.authorization.k8s.io",
			Kind:     "ClusterRole",
		},
		Subjects: []rbacv1.Subject{
			{
				Kind:      "ServiceAccount",
				Name:      r.serviceAccountName,
				Namespace: r.DevEnvNamespace,
			},
		},
	}

	controllerutil.SetControllerReference(cr, crb, r.Scheme)
	return crb
}

func (r *DevEnvReconciler) rbacRBForDevEnv(cr *cndev1alpha1.DevEnv) *rbacv1.RoleBinding {
	labels := labelsForDevEnv(cr.Name)

	rb := &rbacv1.RoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name:      r.resourceName,
			Namespace: r.DevEnvNamespace,
			Labels:    labels,
		},
		RoleRef: rbacv1.RoleRef{
			Name:     cr.Spec.RoleName,
			APIGroup: "rbac.authorization.k8s.io",
			Kind:     "ClusterRole",
		},
		Subjects: []rbacv1.Subject{
			{
				Kind: "ServiceAccount",
				Name: r.serviceAccountName,
			},
		},
	}

	controllerutil.SetControllerReference(cr, rb, r.Scheme)
	return rb
}

func (r *DevEnvReconciler) pvcDockerForDevEnv(cr *cndev1alpha1.DevEnv) *corev1.PersistentVolumeClaim {
	labels := labelsForDevEnv(cr.Name)

	pvc := &corev1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:      r.dockerVolumeName,
			Namespace: r.DevEnvNamespace,
			Labels:    labels,
		},
		Spec: corev1.PersistentVolumeClaimSpec{
			AccessModes: []corev1.PersistentVolumeAccessMode{
				corev1.ReadWriteOnce,
			},
			Resources: corev1.ResourceRequirements{
				Requests: corev1.ResourceList{
					corev1.ResourceName(corev1.ResourceStorage): resource.MustParse(r.dockerVolumeSize),
				},
			},
		},
	}
	if cr.Spec.DeleteVolumes {
		controllerutil.SetControllerReference(cr, pvc, r.Scheme)
	}
	return pvc
}

func (r *DevEnvReconciler) pvcHomeForDevEnv(cr *cndev1alpha1.DevEnv) *corev1.PersistentVolumeClaim {
	labels := labelsForDevEnv(cr.Name)

	pvc := &corev1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:      r.homeVolumeName,
			Namespace: r.DevEnvNamespace,
			Labels:    labels,
		},
		Spec: corev1.PersistentVolumeClaimSpec{
			AccessModes: []corev1.PersistentVolumeAccessMode{
				corev1.ReadWriteOnce,
			},
			Resources: corev1.ResourceRequirements{
				Requests: corev1.ResourceList{
					corev1.ResourceName(corev1.ResourceStorage): resource.MustParse(r.homeVolumeSize),
				},
			},
		},
	}
	if cr.Spec.DeleteVolumes {
		controllerutil.SetControllerReference(cr, pvc, r.Scheme)
	}
	return pvc
}

func (r *DevEnvReconciler) pvcVMForDevEnv(cr *cndev1alpha1.DevEnv) *corev1.PersistentVolumeClaim {
	labels := labelsForDevEnv(cr.Name)

	pvc := &corev1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:      r.vmVolumeName,
			Namespace: r.DevEnvNamespace,
			Labels:    labels,
		},
		Spec: corev1.PersistentVolumeClaimSpec{
			AccessModes: []corev1.PersistentVolumeAccessMode{
				corev1.ReadWriteOnce,
			},
			Resources: corev1.ResourceRequirements{
				Requests: corev1.ResourceList{
					corev1.ResourceName(corev1.ResourceStorage): resource.MustParse(r.homeVolumeSize),
				},
			},
		},
	}
	if cr.Spec.DeleteVolumes {
		controllerutil.SetControllerReference(cr, pvc, r.Scheme)
	}
	return pvc
}

//
func (r *DevEnvReconciler) serviceAccountForDevEnv(cr *cndev1alpha1.DevEnv) *corev1.ServiceAccount {
	labels := labelsForDevEnv(cr.Name)

	sa := &corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      r.serviceAccountName,
			Namespace: r.DevEnvNamespace,
			Labels:    labels,
		},
	}
	controllerutil.SetControllerReference(cr, sa, r.Scheme)
	return sa
}

//
func (r *DevEnvReconciler) podForInitializingDevEnv(cr *cndev1alpha1.DevEnv) *corev1.Pod {
	TRUE := true

	labels := labelsForDevEnv(cr.Name)
	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      r.initName,
			Namespace: r.DevEnvNamespace,
			Labels:    labels,
		},
		Spec: corev1.PodSpec{
			RestartPolicy:      corev1.RestartPolicyNever,
			ServiceAccountName: r.serviceAccountName,
			InitContainers: []corev1.Container{
				{
					Name:    "pre-pull-images",
					Image:   r.devEnvImg,
					Command: []string{"/bin/sh", "-c"},
					Args:    []string{"echo this step is for pulling the image to the node and does actually nothing else"},
					Resources: corev1.ResourceRequirements{
						Requests: map[corev1.ResourceName]resource.Quantity{
							corev1.ResourceMemory: resource.MustParse("8Mi"),
						},
					},
				},
			},
			Containers: []corev1.Container{
				{
					Name:    "init-chroot",
					Image:   r.dockerImg,
					Command: []string{"/bin/sh", "-c"},
					Args: []string{
						`if [ ! -f '/cnde/.cnde' ]; then 
						echo Creating new volume from $DEVENV_IMAGE; 
						docker export $(docker create $DEVENV_IMAGE) | tar -C cnde -xf -; 
						docker inspect --format='{{.Config.Entrypoint}}' $DEVENV_IMAGE > /cnde/.cnde; 
						chown -R 1000.1000 /cnde/home/cnde; 
						else echo File .cnde found. Keeping volume as it is; 
						fi`,
					},
					Env: []corev1.EnvVar{
						{
							Name:  "DEVENV_IMAGE",
							Value: r.devEnvImg,
						},
					},
					SecurityContext: &corev1.SecurityContext{
						Privileged: &TRUE,
					},
					Resources: corev1.ResourceRequirements{
						Requests: map[corev1.ResourceName]resource.Quantity{
							corev1.ResourceMemory: resource.MustParse("128Mi"),
						},
					},
					VolumeMounts: []corev1.VolumeMount{
						{
							Name:      "vm-storage",
							MountPath: "/cnde",
						},
						{
							Name:      "home-storage",
							MountPath: "/cnde/home/cnde",
						},
						{
							Name:      "host-docker-sock",
							MountPath: "/var/run/docker.sock",
						},
					},
				},
			},
			Volumes: []corev1.Volume{
				{
					Name:         "vm-storage",
					VolumeSource: corev1.VolumeSource{PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{ClaimName: r.vmVolumeName}},
				},
				{
					Name:         "home-storage",
					VolumeSource: corev1.VolumeSource{PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{ClaimName: r.homeVolumeName}},
				},
				{
					Name:         "host-docker-sock",
					VolumeSource: corev1.VolumeSource{HostPath: &corev1.HostPathVolumeSource{Path: "/var/run/docker.sock"}},
				},
			},
		},
	}
	return pod
}

func (r *DevEnvReconciler) podForDevEnv(cr *cndev1alpha1.DevEnv) *corev1.Pod {
	TRUE := true

	labels := labelsForDevEnv(cr.Name)
	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      r.resourceName,
			Namespace: r.DevEnvNamespace,
			Labels:    labels,
		},
		Spec: corev1.PodSpec{
			ServiceAccountName: r.serviceAccountName,
			InitContainers: []corev1.Container{
				{
					Name:    "create-kubeconfig",
					Image:   r.kubeConfigImg,
					Command: []string{"/bin/sh", "-c"},
					Args:    []string{"/create_kubeconfig.sh; chown -R 1000.1000 /kube"},
					SecurityContext: &corev1.SecurityContext{
						Privileged: &TRUE,
					},
					Resources: corev1.ResourceRequirements{
						Requests: map[corev1.ResourceName]resource.Quantity{
							corev1.ResourceMemory: resource.MustParse("8Mi"),
						},
					},
					VolumeMounts: []corev1.VolumeMount{
						{
							Name:      "home-storage",
							MountPath: "/kube",
							SubPath:   ".kube",
						},
					},
				},
			},
			Containers: []corev1.Container{
				{
					Name:  "docker-daemon",
					Image: r.dockerImg,
					Env: []corev1.EnvVar{
						{
							Name: "DOCKER_TLS_CERTDIR",
						},
					},
					SecurityContext: &corev1.SecurityContext{
						Privileged: &TRUE,
					},
					Resources: corev1.ResourceRequirements{
						Requests: map[corev1.ResourceName]resource.Quantity{
							corev1.ResourceMemory: r.memRequestDocker,
						},
					},
					VolumeMounts: []corev1.VolumeMount{
						{
							Name:      "docker-storage",
							MountPath: "/var/lib/docker",
						},
					},
				},
				{
					Name:    "code-server",
					Image:   r.alpineImage,
					Command: []string{"/bin/sh", "-c"},
					Args: []string{
						`mount -t proc /proc /home/cnde/proc/; 
						mount --rbind /sys /home/cnde/sys/; 
						mount --rbind /dev /home/cnde/dev/; 
						cp /etc/resolv.conf /home/cnde/etc/; 
						cp /etc/hosts /home/cnde/etc/; 
						ENTRYPOINT=$(cat /home/cnde/.cnde); 
						ENTRYPOINT=${ENTRYPOINT:1:-1}; 
						echo "export ENTRYPOINT=$ENTRYPOINT" >> /home/cnde/etc/environment; 
						echo starting application with: $ENTRYPOINT; 
						exec chroot /home/cnde su cnde -c 'cd /home/cnde; $ENTRYPOINT'`,
					},
					SecurityContext: &corev1.SecurityContext{
						Privileged: &TRUE,
					},
					Resources: corev1.ResourceRequirements{
						Requests: map[corev1.ResourceName]resource.Quantity{
							corev1.ResourceMemory: r.memRequestIDE,
						},
					},
					VolumeMounts: []corev1.VolumeMount{
						{
							Name:      "vm-storage",
							MountPath: "/home/cnde",
						},
						{
							Name:      "home-storage",
							MountPath: "/home/cnde/home/cnde",
						},
					},
				},
			},
			Volumes: []corev1.Volume{
				{
					Name:         "vm-storage",
					VolumeSource: corev1.VolumeSource{PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{ClaimName: r.vmVolumeName}},
				},
				{
					Name:         "home-storage",
					VolumeSource: corev1.VolumeSource{PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{ClaimName: r.homeVolumeName}},
				},
				{
					Name:         "docker-storage",
					VolumeSource: corev1.VolumeSource{PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{ClaimName: r.dockerVolumeName}},
				},
			},
		},
	}

	controllerutil.SetControllerReference(cr, pod, r.Scheme)
	return pod
}

// func (r *DevEnvReconciler) podForDevEnvLimited(cr *cndev1alpha1.DevEnv) *corev1.Pod {
// 	TRUE := true
// 	uid := int64(1000)
// 	sshsecretName := cr.Spec.SSHSecret
// 	if sshsecretName == "" {
// 		sshsecretName = "non-provided-ssh-secret"
// 	}

// 	labels := labelsForDevEnv(cr.Name)
// 	pod := &corev1.Pod{
// 		ObjectMeta: metav1.ObjectMeta{
// 			Name:      r.resourceName,
// 			Namespace: r.DevEnvNamespace,
// 			Labels:    labels,
// 		},
// 		Spec: corev1.PodSpec{
// 			ServiceAccountName: r.resourceName,
// 			InitContainers: []corev1.Container{
// 				{
// 					Name:  "configure",
// 					Image: r.ConfigureImg,
// 					Command: []string{
// 						"/bin/sh",
// 						"-c",
// 						"if [ ! -f '/cnde/.cnde' ]; then " +
// 							"sudo mkdir -p /cnde/project; sudo mkdir -p /cnde/.kube;" +
// 							"sudo cp -aR /home/cnde/. /cnde/; sudo chown -R cnde:cnde /cnde/.; " +
// 							"touch /cnde/.cnde; " +
// 							"if [ -f '/ssh/id_rsa' ]; then " +
// 							"	echo --> copy ssh keys from secret to .ssh; " +
// 							"	mkdir /cnde/.ssh; cp /ssh/* /cnde/.ssh/; chmod 0600 /cnde/.ssh/id_rsa; fi; " +
// 							"fi",
// 					},
// 					VolumeMounts: []corev1.VolumeMount{
// 						{
// 							Name:      "ssh-key-volume",
// 							MountPath: "/ssh",
// 							ReadOnly:  true,
// 						},
// 						{
// 							Name:      "home-storage",
// 							MountPath: "/cnde",
// 						},
// 					},
// 				},
// 				{
// 					Name:  "create-kubeconfig",
// 					Image: r.KubeConfigImg,
// 					SecurityContext: &corev1.SecurityContext{
// 						RunAsUser: &uid,
// 					},
// 					VolumeMounts: []corev1.VolumeMount{
// 						{
// 							Name:      "home-storage",
// 							MountPath: "/kube",
// 							SubPath:   ".kube",
// 						},
// 					},
// 				},
// 			},
// 			Containers: []corev1.Container{
// 				{
// 					Name:  "docker-daemon",
// 					Image: r.DockerImg,
// 					Env: []corev1.EnvVar{
// 						{
// 							Name: "DOCKER_TLS_CERTDIR",
// 						},
// 					},
// 					SecurityContext: &corev1.SecurityContext{
// 						Privileged: &TRUE,
// 					},
// 					VolumeMounts: []corev1.VolumeMount{
// 						{
// 							Name:      "docker-storage",
// 							MountPath: "/var/lib/docker",
// 						},
// 					},
// 				},
// 				{
// 					Name:  "code-server",
// 					Image: r.DevEnvImg,
// 					Command: []string{"/bin/sh",
// 						"-c",
// 						"exec dumb-init code-server --auth none",
// 					},
// 					SecurityContext: &corev1.SecurityContext{
// 						Privileged: &TRUE,
// 					},
// 					VolumeMounts: []corev1.VolumeMount{
// 						{
// 							Name:      "home-storage",
// 							MountPath: "/home/cnde",
// 						},
// 					},
// 				},
// 			},
// 			Volumes: []corev1.Volume{
// 				{
// 					Name: "ssh-key-volume",
// 					VolumeSource: corev1.VolumeSource{Secret: &corev1.SecretVolumeSource{
// 						SecretName: sshsecretName,
// 						Optional:   &TRUE,
// 					}},
// 				},
// 				{
// 					Name:         "home-storage",
// 					VolumeSource: corev1.VolumeSource{PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{ClaimName: r.homeVolumeName}},
// 				},
// 				{
// 					Name:         "docker-storage",
// 					VolumeSource: corev1.VolumeSource{PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{ClaimName: r.dockerVolumeName}},
// 				},
// 			},
// 		},
// 	}

// 	controllerutil.SetControllerReference(cr, pod, r.Scheme)
// 	return pod
// }

//------------------------------------------------------------------------
//------------------------------------------------------------------------

func (r *DevEnvReconciler) ingressOauthForDevEnv(cr *cndev1alpha1.DevEnv) *extv1beta1.Ingress {
	labels := labelsForDevEnv(cr.Name)
	ingOauth := &extv1beta1.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Name:      r.ingressOauthName,
			Namespace: r.ManagerNamespace,
			Labels:    labels,
			Annotations: map[string]string{
				"kubernetes.io/ingress.class":                   "nginx",
				"kubernetes.io/tls-acme":                        "true",
				"nginx.ingress.kubernetes.io/proxy-buffer-size": "16k",
			},
		},
		Spec: extv1beta1.IngressSpec{
			Rules: []extv1beta1.IngressRule{
				{
					Host: r.ingressHost,
					IngressRuleValue: extv1beta1.IngressRuleValue{
						HTTP: &extv1beta1.HTTPIngressRuleValue{
							Paths: []extv1beta1.HTTPIngressPath{
								{
									Path: "/oauth2",
									Backend: extv1beta1.IngressBackend{
										ServiceName: r.proxyPodName,
										ServicePort: intstr.FromInt(4180),
									},
								},
							},
						},
					},
				},
			},
			TLS: []extv1beta1.IngressTLS{
				{
					Hosts: []string{
						r.ingressHost,
					},
					SecretName: r.resourceName + "-tls",
				},
			},
		},
	}

	controllerutil.SetControllerReference(cr, ingOauth, r.Scheme)
	return ingOauth
}

func (r *DevEnvReconciler) ingressUIForDevEnv(cr *cndev1alpha1.DevEnv) *extv1beta1.Ingress {
	labels := labelsForDevEnv(cr.Name)
	ingUI := &extv1beta1.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Name:      r.ingressUIName,
			Namespace: r.ManagerNamespace,
			Labels:    labels,
			Annotations: map[string]string{
				"kubernetes.io/ingress.class":                       "nginx",
				"nginx.ingress.kubernetes.io/auth-response-headers": "X-Auth-Request-User, X-Auth-Request-Email",
				"nginx.ingress.kubernetes.io/auth-signin":           "https://$host/oauth2/start?rd=$request_uri",
				"nginx.ingress.kubernetes.io/auth-url":              "https://$host/oauth2/auth",
				"nginx.ingress.kubernetes.io/proxy-buffer-size":     "16k",
			},
		},
		Spec: extv1beta1.IngressSpec{
			Rules: []extv1beta1.IngressRule{
				{
					Host: r.ingressHost,
					IngressRuleValue: extv1beta1.IngressRuleValue{
						HTTP: &extv1beta1.HTTPIngressRuleValue{
							Paths: []extv1beta1.HTTPIngressPath{
								{
									Path: "/",
									Backend: extv1beta1.IngressBackend{
										ServiceName: r.resourceName,
										ServicePort: intstr.FromInt(idePort),
									},
								},
								{
									Path: "/terminal/",
									Backend: extv1beta1.IngressBackend{
										ServiceName: r.resourceName,
										ServicePort: intstr.FromInt(ttydPort),
									},
								},
							},
						},
					},
				},
			},
		},
	}

	controllerutil.SetControllerReference(cr, ingUI, r.Scheme)
	return ingUI
}

func (r *DevEnvReconciler) podOauthProxyForDevEnv(cr *cndev1alpha1.DevEnv) *corev1.Pod {
	labels := map[string]string{"user-env-name": cr.Name, "app": "oauth2-proxy"}
	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      r.proxyPodName,
			Namespace: r.ManagerNamespace,
			Labels:    labels,
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name:  "oauth2-proxy",
					Image: r.oauthProxyImg,
					Args: []string{
						"--cookie-name=auth",
						"--cookie-refresh=23h",
						"--cookie-secure=true",
						"--email-domain=*",
						"--http-address=0.0.0.0:4180",
						"--oidc-issuer-url=https://keycloak." + cr.Spec.UserEnvDomain + "/auth/realms/" + r.resourceName,
						"--pass-access-token=true",
						"--provider=oidc",
						"--set-xauthrequest=true",
						"--tls-cert-file=",
						"--upstream=file:///dev/null",
						// remove for prod cluster issuer
						"--ssl-insecure-skip-verify=true",
					},
					Ports: []corev1.ContainerPort{
						{
							Name:          "http",
							ContainerPort: 4180,
							Protocol:      corev1.ProtocolTCP,
						},
					},
					Resources: corev1.ResourceRequirements{
						Requests: corev1.ResourceList{
							corev1.ResourceCPU:    resource.MustParse("10m"),
							corev1.ResourceMemory: resource.MustParse("8Mi"),
						},
					},
					Env: []corev1.EnvVar{
						{
							Name: "OAUTH2_PROXY_CLIENT_ID",
							ValueFrom: &corev1.EnvVarSource{
								SecretKeyRef: &corev1.SecretKeySelector{
									Key: "client_id",
									LocalObjectReference: corev1.LocalObjectReference{
										Name: r.proxyPodName,
									},
								},
							},
						},
						{
							Name: "OAUTH2_PROXY_CLIENT_SECRET",
							ValueFrom: &corev1.EnvVarSource{
								SecretKeyRef: &corev1.SecretKeySelector{
									Key: "client_secret",
									LocalObjectReference: corev1.LocalObjectReference{
										Name: r.proxyPodName,
									},
								},
							},
						},
						{
							Name: "OAUTH2_PROXY_COOKIE_SECRET",
							ValueFrom: &corev1.EnvVarSource{
								SecretKeyRef: &corev1.SecretKeySelector{
									Key: "cookie_secret",
									LocalObjectReference: corev1.LocalObjectReference{
										Name: r.proxyPodName,
									},
								},
							},
						},
					},
					LivenessProbe: &corev1.Probe{
						Handler: corev1.Handler{
							HTTPGet: &corev1.HTTPGetAction{
								Path:   "/ping",
								Port:   intstr.FromString("http"),
								Scheme: corev1.URISchemeHTTP,
							},
						},
						InitialDelaySeconds: 30,
					},
					ReadinessProbe: &corev1.Probe{
						Handler: corev1.Handler{
							HTTPGet: &corev1.HTTPGetAction{
								Path:   "/ping",
								Port:   intstr.FromString("http"),
								Scheme: corev1.URISchemeHTTP,
							},
						},
					},
				},
			},
		},
	}

	controllerutil.SetControllerReference(cr, pod, r.Scheme)
	return pod
}

func (r *DevEnvReconciler) serviceOauthProxyForDevEnv(cr *cndev1alpha1.DevEnv) *corev1.Service {
	labels := map[string]string{"user-env-name": cr.Name, "app": "oauth2-proxy"}
	ser := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      r.proxyPodName,
			Namespace: r.ManagerNamespace,
			Labels:    labels,
		},
		Spec: corev1.ServiceSpec{
			Selector: labels,
			Ports: []corev1.ServicePort{
				{
					Port: 4180,
					Name: "http",
				},
			},
		},
	}

	controllerutil.SetControllerReference(cr, ser, r.Scheme)
	return ser
}

func (r *DevEnvReconciler) secretOauthProxyForDevEnv(cr *cndev1alpha1.DevEnv) *corev1.Secret {
	labels := labelsForDevEnv(cr.Name)
	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      r.proxyPodName,
			Namespace: r.ManagerNamespace,
			Labels:    labels,
		},
		StringData: map[string]string{
			"client_id":     r.oauthClientID,
			"client_secret": r.oauthClientSecret,
			"cookie_secret": `WhatEver123456888888`,
		},
	}

	controllerutil.SetControllerReference(cr, secret, r.Scheme)
	return secret
}

func (r *DevEnvReconciler) createNamespaceForDevEnv(cr *cndev1alpha1.DevEnv) *corev1.Namespace {
	labels := labelsForDevEnv(cr.Name)
	labels[namespaceLabel] = r.ManagerNamespace

	ns := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name:   r.resourceName,
			Labels: labels,
		},
		Spec: corev1.NamespaceSpec{},
	}

	if cr.Spec.DeleteVolumes {
		controllerutil.SetControllerReference(cr, ns, r.Scheme)
	}
	return ns
}

func (r *DevEnvReconciler) serviceProxyForDevEnv(cr *cndev1alpha1.DevEnv) *corev1.Service {
	labels := labelsForDevEnv(cr.Name)
	ser := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      r.resourceName,
			Namespace: r.ManagerNamespace,
			Labels:    labels,
		},
		Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{
				{
					Name: "ide",
					Port: idePort,
				},
				{
					Name: "terminal",
					Port: ttydPort,
				},
			},
		},
	}
	controllerutil.SetControllerReference(cr, ser, r.Scheme)
	return ser
}

func (r *DevEnvReconciler) endpointProxyForDevEnv(cr *cndev1alpha1.DevEnv) *corev1.Endpoints {
	labels := labelsForDevEnv(cr.Name)
	endpoint := &corev1.Endpoints{
		ObjectMeta: metav1.ObjectMeta{
			Name:      r.resourceName,
			Namespace: r.ManagerNamespace,
			Labels:    labels,
		},
		Subsets: []corev1.EndpointSubset{
			{
				Addresses: []corev1.EndpointAddress{
					{
						IP: r.DevEnvPodIP,
					},
				},
				Ports: []corev1.EndpointPort{
					{
						Name: "ide",
						Port: idePort,
					},
					{
						Name: "terminal",
						Port: ttydPort,
					},
				},
			},
		},
	}
	controllerutil.SetControllerReference(cr, endpoint, r.Scheme)

	return endpoint
}
