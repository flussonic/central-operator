/*
Copyright 2024.

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
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// CentralSpec defines the desired state of Central
type CentralSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Central Docker Image (https://hub.docker.com/r/flussonic/central)
	Image string `json:"image"`

	// Postgresql Database connection string
	// You must setup it by yourself
	Database string `json:"database"`

	// (Optional) PodEnvVariables is a slice of environment variables that are added to the pods
	// Default: (empty list)
	// +kubebuilder:validation:Optional
	PodEnvVariables []corev1.EnvVar `json:"env,omitempty"`

	// Selector for nodes with running central worker instances
	NodeSelector map[string]string `json:"nodeSelector,omitempty" protobuf:"bytes,7,rep,name=nodeSelector"`

	// API_URL used for setting the hostname and port under which Central
	// is accessible by Flussonic for CONFIG_EXTERNAL and http_proxy requests
	APIURL string `json:"apiUrl"`

	// API_KEY is used to access Central API
	APIKey string `json:"apiKey"`

	// Credentials for administrator access to the Central Admin UI.
	// +kubebuilder:validation:Optional
	EditAuth string `json:"editAuth"`

	// Enables logging HTTP-requests
	// +kubebuilder:validation:Optional
	LogRequests bool `json:"logRequests"`

	// Logging level
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Enum:=debug;info;error
	LogLevel string `json:"logLevel,omitempty"`

	// Token for accessing dynamic streams
	// +kubebuilder:validation:Optional
	DynamicStreamsAuthToken string `json:"dynamicStreamsAuthToken"`

	// Pod selector for locating mediaserver instances.
	// If not empty - Central will automatically provision streamers to cluster from k8s API
	// +kubebuilder:validation:Optional
	ProvisionerSelector string `json:"provisionerSelector"`

	// API key for accessing mediaserver instances provisioned to Central
	// +kubebuilder:validation:Optional
	ProvisionerClusterKey string `json:"provisionerClusterKey"`
}

// CentralStatus defines the observed state of Central
type CentralStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// Central is the Schema for the centrals API
type Central struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   CentralSpec   `json:"spec,omitempty"`
	Status CentralStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// CentralList contains a list of Central
type CentralList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Central `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Central{}, &CentralList{})
}
