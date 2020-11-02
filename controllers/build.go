package controllers

import (
	"strings"

	cndev1alpha1 "cnde-operator.cloud-native-coding.dev/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (r *DevEnvReconciler) buildPodForDevEnv(cr *cndev1alpha1.DevEnv, b *cndev1alpha1.Builder) *corev1.Pod {
	labels := labelsForDevEnv(cr.Name)

	podSpec := b.Spec.Template

	podSpec.RestartPolicy = corev1.RestartPolicyNever

	podSpec.Containers[0].Env = append(podSpec.Containers[0].Env, corev1.EnvVar{
		Name:  imageTagName,
		Value: r.devEnvImg,
	})

	for i, arg := range podSpec.Containers[0].Args {
		podSpec.Containers[0].Args[i] = strings.Replace(arg, "$"+imageTagName, r.devEnvImg, -1)
	}

	for i, cmd := range podSpec.Containers[0].Command {
		podSpec.Containers[0].Command[i] = strings.Replace(cmd, "$"+imageTagName, r.devEnvImg, -1)
	}

	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      r.buildName,
			Namespace: r.ManagerNamespace,
			Labels:    labels,
		},
		Spec: podSpec,
	}

	return pod
}
