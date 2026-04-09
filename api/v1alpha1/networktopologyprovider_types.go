// Copyright (c) 2026 Broadcom. All Rights Reserved.
// Broadcom Confidential. The term "Broadcom" refers to Broadcom Inc.
// and/or its subsidiaries.

package v1alpha1

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

type TypedObjectReference struct {
	// APIGroup is the group for the resource being referenced.
	APIGroup string `json:"apiGroup"`
	// Kind is the type of resource being referenced.
	Kind string `json:"kind"`
	// Name is the name of resource being referenced.
	Name string `json:"name"`
}

type NetworkTopologyType string

const (
	// NetworkTopologyTypeNSXT is the type for NSX Container Plugin Tier1-per-namespace network topology.
	NetworkTopologyTypeNSXT NetworkTopologyType = "nsx-t"
	// NetworkTopologyTypeVDS is the type for vSphere Networking topology.
	NetworkTopologyTypeVDS NetworkTopologyType = "vsphere-distributed"
	// NetworkTopologyTypeNSXTVPC is the type for NSX-T VPC network topology.
	NetworkTopologyTypeNSXTVPC NetworkTopologyType = "nsx-t_vpc"
)

type NetworkTopologyProviderSpec struct {
	// Type describes type of network topology.
	// +kubebuilder:validation:Enum=nsx-t;vsphere-distributed;nsx-t_vpc
	Type NetworkTopologyType `json:"type"`
	// ProviderRef is reference to a network topology provider object that provides the details for this type of network topology provider.
	ProviderRef TypedObjectReference `json:"providerRef"`
}

// +kubebuilder:object:root=true
// +kubebuilder:resource:shortName=ntp,scope=Cluster

// NetworkTopologyProvider is the Schema for the networktopologyproviders API.
// A NetworkTopologyProvider represents a network topology provider configuration.
type NetworkTopologyProvider struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec NetworkTopologyProviderSpec `json:"spec,omitempty"`
}

// +kubebuilder:object:root=true

type NetworkTopologyProviderList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []NetworkTopologyProvider `json:"items"`
}

func init() {
	RegisterTypeWithScheme(&NetworkTopologyProvider{}, &NetworkTopologyProviderList{})
}
