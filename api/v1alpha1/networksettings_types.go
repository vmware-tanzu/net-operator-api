// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: Apache-2.0

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +kubebuilder:object:root=true
// +kubebuilder:resource:scope=Namespaced
// +kubebuilder:subresource:status
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

	// status defines the observed, realized state of NetworkSettings.
	//
	// +optional
	Status *NetworkSettingsStatus `json:"status,omitempty"`
}

// NetworkSettingsStatus defines the observed state of NetworkSettings.
type NetworkSettingsStatus struct {
	// workloadCapabilities summarizes the supported IP families and status messages per workload type.
	// +optional
	// +kubebuilder:validation:MaxItems=32
	// +listType=map
	// +listMapKey=type
	WorkloadCapabilities []WorkloadCapability `json:"workloadCapabilities,omitempty"`
}

// WorkloadCapability defines the supported IP families, default IP family, and messages for a specific workload type.
// +kubebuilder:validation:XValidation:rule="self.supportedIPFamilies.all(f, f in ['IPv4', 'IPv6', 'DualStack'])",message="supportedIPFamilies must only contain 'IPv4', 'IPv6', or 'DualStack'"
type WorkloadCapability struct {
	// type identifies the workload category (e.g., "vksCluster", "podVM", "virtualMachine").
	// Custom or future workload types can be added without CRD breaking changes.
	//
	// +required
	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:validation:MaxLength=63
	Type string `json:"type,omitempty"`

	// supportedIPFamilies lists the IP family modes (e.g. "IPv4", "IPv6", "DualStack") supported for this workload type.
	// For vksCluster, this can contain up to ["IPv4", "IPv6", "DualStack"].
	//
	// +required
	// +listType=set
	// +kubebuilder:validation:MaxItems=3
	// +kubebuilder:validation:items:MinLength=1
	// +kubebuilder:validation:items:MaxLength=16
	SupportedIPFamilies []string `json:"supportedIPFamilies,omitempty"`

	// defaultIPFamily specifies the default IP family assigned if none is explicitly specified by the workload (e.g. "IPv4", "IPv6", "DualStack").
	//
	// +optional
	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:validation:MaxLength=16
	DefaultIPFamily string `json:"defaultIPFamily,omitempty"`

	// message provides human-readable actionable warnings or reasons when capabilities are constrained.
	//
	// +optional
	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:validation:MaxLength=2048
	Message string `json:"message,omitempty"`
}

const (
	// WorkloadTypeVKSCluster identifies the Virtual Kubernetes Service cluster workload type.
	WorkloadTypeVKSCluster = "vksCluster"

	// WorkloadTypePodVM identifies the PodVM workload type.
	WorkloadTypePodVM = "podVM"

	// WorkloadTypeVirtualMachine identifies the standalone Virtual Machine workload type.
	WorkloadTypeVirtualMachine = "virtualMachine"

	// WorkloadIPFamilyIPv4 indicates IPv4 single-stack support.
	WorkloadIPFamilyIPv4 = "IPv4"

	// WorkloadIPFamilyIPv6 indicates IPv6 single-stack support.
	WorkloadIPFamilyIPv6 = "IPv6"

	// WorkloadIPFamilyDualStack indicates DualStack support.
	WorkloadIPFamilyDualStack = "DualStack"
)

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
