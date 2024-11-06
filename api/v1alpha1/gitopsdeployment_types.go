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

// GitOpsDeploymentSpec defines the desired state of GitOpsDeployment.
type GitOpsDeploymentSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	RepoURL  string `json:"repoURL"`
	Branch   string `json:"branch"`
	Path     string `json:"path"`     // Directory in the repo with manifests
	Interval string `json:"interval"` // How often to check for updates
}

// GitOpsDeploymentStatus defines the observed state of GitOpsDeployment.
type GitOpsDeploymentStatus struct {
	Synced       bool               `json:"synced"`
	LastSyncTime metav1.Time        `json:"lastSyncTime"`
	Conditions   []metav1.Condition `json:"conditions"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// GitOpsDeployment is the Schema for the gitopsdeployments API.
type GitOpsDeployment struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   GitOpsDeploymentSpec   `json:"spec,omitempty"`
	Status GitOpsDeploymentStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// GitOpsDeploymentList contains a list of GitOpsDeployment.
type GitOpsDeploymentList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []GitOpsDeployment `json:"items"`
}

func init() {
	SchemeBuilder.Register(&GitOpsDeployment{}, &GitOpsDeploymentList{})
}
