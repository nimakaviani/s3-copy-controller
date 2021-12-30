/*
Copyright 2021.

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

// A SecretReference is a reference to a secret in an arbitrary namespace.
type SecretReference struct {
	// Name of the secret.
	Name string `json:"name,required"`

	// Namespace of the secret.
	Namespace string `json:"namespace,required"`
}

// A SecretKeySelector is a reference to a secret key in an arbitrary namespace.
type SecretKeySelector struct {
	SecretReference `json:",inline"`

	// The key to select.
	Key string `json:"key,required"`
}

type Credentials struct {
	Source          string            `json:"source,omitempty"`
	SecretReference SecretKeySelector `json:"secretRef"`
}

// An ObjectSource refers to the location to get the object from
type ObjectSource struct {
	// sourcetype: local / configmap
	// +kubebuilder:default:=local
	Reference string `json:"reference,omitempty"`
	// namespace for configmap
	Namespace string `json:"namespace,omitempty"`
	// name for configmap
	Name string `json:"name,omitempty"`
	// The key to select.
	Key string `json:"key,omitempty"`
	// raw content for the object
	Data string `json:"data,omitempty"`
}

// An ObjectTarget refers to the object store reference to store the object into
type ObjectTarget struct {
	// reference to where the object will be stored
	Bucket string `json:"bucket,required"`
	// region to be used for creds
	Region string `json:"region,required"`
	// object key
	Key string `json:"key,required"`
}

// ObjectSpec defines the desired state of Object
type ObjectSpec struct {
	DeletionPolicy string       `json:"deletionPolicy"`
	Credentials    Credentials  `json:"credentials,required"`
	Source         ObjectSource `json:"source,required"`
	Target         ObjectTarget `json:"target,required"`
}

// ObjectStatus defines the observed state of Object
type ObjectStatus struct {
	// +kubebuilder:default:=false
	Synced    bool   `json:"synced"`
	Reference string `json:"reference"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:printcolumn:name="Synced",type="string",JSONPath=".status.synced",description="Whether or not the sync succeeded"
//+kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"
//+kubebuilder:printcolumn:name="Reference",type="string",JSONPath=".status.reference",description="Object reference in the target object store"

// Object is the Schema for the objects API
type Object struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ObjectSpec   `json:"spec,omitempty"`
	Status ObjectStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// ObjectList contains a list of Object
type ObjectList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Object `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Object{}, &ObjectList{})
}
