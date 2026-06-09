// Copyright (c) 2026 Broadcom. All Rights Reserved.
// Broadcom Confidential. The term "Broadcom" refers to Broadcom Inc.
// and/or its subsidiaries.

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +kubebuilder:object:root=true
// +kubebuilder:resource:scope=Namespaced
// +kubebuilder:validation:XValidation:rule="!has(self.legacyProvider) || self.legacyProvider != self.provider",message="legacyProvider must differ from provider"
//
// NetworkSettings exposes information about the effective network configuration for a namespace.
// This is observed, realized state, and its contents may be updated by further network configuration
// mutations.
type NetworkSettings struct {
	metav1.TypeMeta `json:",inline"`

	// metadata is the standard object's metadata.
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// provider is the active network provider in this namespace. Workloads and network-aware
	// components should use this to determine which provider governs networking for the namespace,
	// including when choosing defaulting behavior or which provider-specific APIs to use when not
	// otherwise specified.
	//
	// +required
	Provider NetworkProvider `json:"provider,omitempty"`

	// legacyProvider is the network provider to which this namespace was previously affined.
	// When set, the namespace has transitioned from legacyProvider to the current provider.
	// APIs and resources associated with legacyProvider remain functional within this namespace
	// but are no longer the governing provider; new network resources will be created under
	// the current provider's APIs.
	//
	// This field is absent when the namespace has never undergone a provider transition.
	//
	// +optional
	// +kubebuilder:validation:XValidation:rule="self in ['vsphere-distributed', 'nsx-tier1']",message="legacyProvider must be vsphere-distributed or nsx-tier1"
	LegacyProvider NetworkProvider `json:"legacyProvider,omitempty"`
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
