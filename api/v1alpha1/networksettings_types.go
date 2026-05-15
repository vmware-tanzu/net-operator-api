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
