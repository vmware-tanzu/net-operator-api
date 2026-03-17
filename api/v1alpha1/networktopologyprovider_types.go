// Copyright (c) 2026 Broadcom. All Rights Reserved.
// Broadcom Confidential. The term "Broadcom" refers to Broadcom Inc.
// and/or its subsidiaries.

package v1alpha1

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

type NetworkTopologyProviderReference struct {
	// APIGroup is the group for the resource being referenced.
	APIGroup string `json:"apiGroup"`
	// Kind is the type of resource being referenced.
	Kind string `json:"kind"`
	// Name is the name of resource being referenced.
	Name string `json:"name"`
	// API version of the referent.
	APIVersion string `json:"apiVersion,omitempty"`
}

type NetworkTopologyProviderType string

const (
	// NetworkTopologyProviderTypeNSXT is the provider type for NSX Container Plugin Tier1-per-namespace network
	// topology.
	NetworkTopologyProviderTypeNSXT NetworkTopologyProviderType = "nsx_container_plugin"
	// NetworkTopologyProviderTypeVDS is the provider type for vSphere Networking topology.
	NetworkTopologyProviderTypeVDS NetworkTopologyProviderType = "vsphere_network"
	// NetworkTopologyProviderTypeNSXTVPC is the provider type for NSX-T VPC network topology.
	NetworkTopologyProviderTypeNSXTVPC NetworkTopologyProviderType = "nsx_vpc"
)

type NetworkTopologyProviderSpec struct {
	// Type describes type of network topology provider.
	// +kubebuilder:validation:Enum=nsx_container_plugin;vsphere_network;nsx_vpc
	Type NetworkTopologyProviderType `json:"type"`
	// ProviderRef is reference to a network topology provider object that provides the details for this type of network topology provider
	ProviderRef NetworkTopologyProviderReference `json:"providerRef"`
}

// +genclient
// +kubebuilder:object:root=true

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
