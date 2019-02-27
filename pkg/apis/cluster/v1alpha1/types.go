/*
Copyright 2018 The Rook Authors. All rights reserved.

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
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ***************************************************************************
// IMPORTANT FOR CODE GENERATION
// If the types in this file are updated, you will need to run
// `make codegen` to generate the new types under the client/clientset folder.
// ***************************************************************************

// +genclient
// +genclient:noStatus
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type Cluster struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`
	Spec              ClusterSpec   `json:"spec"`
	Status            ClusterStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type ClusterList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`
	Items           []Cluster `json:"items"`
}

type ClusterSpec struct {
	// The version information that instructs Rook to orchestrate a particular version of Ceph.
	SlurmVersion SlurmVersionSpec `json:"slurmVersion,omitempty"`

	// A spec for available storage in the cluster and how it should be used
	Pool PoolSpec `json:"pool,omitempty"`

	// Resources set resource requests and limits
	Resource v1.ResourceRequirements `json:"resource,omitempty"`
}

// VersionSpec represents the settings for the Ceph version that Rook is orchestrating.
type SlurmVersionSpec struct {
	// Image is the container image used to launch the ceph daemons, such as ceph/ceph:v12.2.7 or ceph/ceph:v13.2.1
	Image string `json:"image,omitempty"`

	// The name of the major release of Ceph: luminous, mimic, or nautilus
	Name string `json:"name,omitempty"`
}

type ClusterStatus struct {
	State   ClusterState `json:"state,omitempty"`
	Message string       `json:"message,omitempty"`
}

type ClusterState string

const (
	ClusterStateCreating ClusterState = "Creating"
	ClusterStateCreated  ClusterState = "Created"
	ClusterStateUpdating ClusterState = "Updating"
	ClusterStateError    ClusterState = "Error"
)

type PoolSpec struct {
	Count              int32 `json:"count"`
	Template v1.PodTemplateSpec `json:"slave,omitempty"`
}

