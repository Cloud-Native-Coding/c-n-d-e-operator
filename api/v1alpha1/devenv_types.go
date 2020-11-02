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

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// DevEnvSpec defines the desired state of DevEnv
type DevEnvSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Volume settings
	DockerVolumeSize string `json:"dockerVolumeSize"`
	HomeVolumeSize   string `json:"homeVolumeSize"`
	DeleteVolumes    bool   `json:"deleteVolumes"`

	// Operator environment
	UserEnvDomain string `json:"userEnvDomain"`
	KeycloakHost  string `json:"keycloakHost"`
	UserEmail     string `json:"userEmail"`

	// Definition of Container Images
	DockerImg     string `json:"dockerImg,omitempty"`
	DevEnvImg     string `json:"devEnvImg,omitempty"`
	KubeConfigImg string `json:"kubeConfigImg,omitempty"`
	ConfigureImg  string `json:"configureImg,omitempty"`
	OauthProxyImg string `json:"oauthProxyImg,omitempty"`

	// DevEnv configuration
	SSHSecret       string `json:"sshSecret,omitempty"`
	ClusterRoleName string `json:"clusterRoleName"`
	RoleName        string `json:"roleName"`

	BuilderName string `json:"builderName,omitempty"`
}

// BuildPhase is the status of build phases
type BuildPhase string

const (
	// BuildPhaseInitial cr just created
	BuildPhaseInitial BuildPhase = ""
	// BuildPhaseBuilding waiting for build pod
	BuildPhaseBuilding = "Building"
	// BuildPhaseWaitForInitializion waiting for init volume Pod
	BuildPhaseWaitForInitializion = "WaitForInitializion"
	// BuildPhaseInitializing waiting for init volume Pod
	BuildPhaseInitializing = "Initializing"
	// BuildPhaseRunning DevEnv pod is started
	BuildPhaseRunning = "Running"
)

// DevEnvStatus defines the observed state of DevEnv
type DevEnvStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	Realm string     `json:"realm"`
	User  string     `json:"user"`
	Build BuildPhase `json:"build"`
}

// +kubebuilder:object:root=true
// +kubebuilder:resource:scope=Cluster
// +kubebuilder:subresource:status

// DevEnv is the Schema for the devenvs API
type DevEnv struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   DevEnvSpec   `json:"spec,omitempty"`
	Status DevEnvStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// DevEnvList contains a list of DevEnv
type DevEnvList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []DevEnv `json:"items"`
}

func init() {
	SchemeBuilder.Register(&DevEnv{}, &DevEnvList{})
}
