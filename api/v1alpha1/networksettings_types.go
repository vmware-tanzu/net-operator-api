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
// +kubebuilder:validation:XValidation:rule="!has(self.previousProvider) || self.previousProvider != self.provider",message="previousProvider must differ from provider"
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

	// previousProvider is the network provider this namespace used immediately before
	// transitioning to the current provider. When set, the migration from previousProvider
	// to provider is either in progress or has recently completed; resources associated with
	// the previous provider (such as Network objects originally backed by VSphereDistributedNetwork)
	// may still exist in this namespace and require continued validation and reconciliation.
	//
	// Operators that serve resources for multiple providers should continue to validate and
	// reconcile resources associated with previousProvider's APIs until this field is cleared.
	// Net Operator clears this field once all resources belonging to the previous provider
	// have been removed from the namespace.
	//
	// This field is absent when the namespace has never undergone a provider transition, or
	// after a transition has fully completed and all prior-provider resources are gone.
	//
	// +optional
	PreviousProvider NetworkProvider `json:"previousProvider,omitempty"`
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
