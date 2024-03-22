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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// CentralSpec defines the desired state of Central
type CentralSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	Image string `json:"image"`
	// Postgresql url
	Database string `json:"database"`

	// Selector for nodes with running central worker instances
	NodeSelector map[string]string `json:"nodeSelector,omitempty" protobuf:"bytes,7,rep,name=nodeSelector"`

	// url of this installation of central
	APIURL string `json:"api_url"`
	// API key for connecting to this central node
	APIKey string `json:"api_key"`

	// Credentials for modifying installation of media server
	EditAuth string `json:"edit_auth"`

	// Enables logging HTTP-requests
	LogRequests bool `json:"log_requests,omitempty"`

	HTTPPort int `json:"http_port"`

	// API key for accessing mediaserver instances
	ProvisionerClusterKey string `json:"provisioner_cluster_key"`
	// Pod selector for locating mediaserver instances
	ProvisionerSelector string `json:"provisioner_selector"`
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
