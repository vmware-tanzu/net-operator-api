// Copyright (c) 2026 Broadcom. All Rights Reserved.
// Broadcom Confidential. The term "Broadcom" refers to Broadcom Inc.
// and/or its subsidiaries.

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// NetworkSettingsProvider is the active network provider type for a namespace.
type NetworkSettingsProvider string

const (
	// NetworkSettingsProviderVSphereDistributed is vSphere Distributed (VDS) network backing.
	NetworkSettingsProviderVSphereDistributed NetworkSettingsProvider = "vsphere-distributed"
	// NetworkSettingsProviderNSXTier1 is NSX Tier-1 network backing.
	NetworkSettingsProviderNSXTier1 NetworkSettingsProvider = "nsx-tier1"
	// NetworkSettingsProviderVPC is VPC (NSX) network backing.
	NetworkSettingsProviderVPC NetworkSettingsProvider = "vpc"
)

// +genclient
// +kubebuilder:object:root=true
// +kubebuilder:resource:scope=Namespaced
//
// NetworkSettings exposes read-only, operator-relevant information about the effective network
// configuration for a namespace.
//
// Consumers should treat it as observed, realized state, and expect it to track the network topology
// backing the namespace.
type NetworkSettings struct {
	metav1.TypeMeta `json:",inline"`

	// metadata is the standard object's metadata.
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// provider is the active network provider in this namespace. Workloads and network-aware
	// components should use this to determine the network backing that is in effect, including
	// when choosing defaulting behavior or which provider-specific APIs to use when not specified
	// elsewhere.
	//
	// +required
	// +kubebuilder:validation:Enum=vsphere-distributed;nsx-tier1;vpc
	Provider NetworkSettingsProvider `json:"provider,omitempty"`
}

// +kubebuilder:object:root=true

// NetworkSettingsList is a list of NetworkSettings.
type NetworkSettingsList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []NetworkSettings `json:"items"`
}

func init() {
	RegisterTypeWithScheme(&NetworkSettings{}, &NetworkSettingsList{})
}
