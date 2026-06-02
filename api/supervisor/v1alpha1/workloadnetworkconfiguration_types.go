// Copyright (c) 2026 Broadcom. All Rights Reserved.
// Broadcom Confidential. The term "Broadcom" refers to Broadcom Inc.
// and/or its subsidiaries.

package v1alpha1

import (
	netv1alpha1 "github.com/vmware-tanzu/net-operator-api/api/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	// WorkloadNetworkConfigurationName is the singleton resource name.
	WorkloadNetworkConfigurationName = "default"

	// WorkloadNetworkConditionReady indicates whether the workload network
	// configuration as a whole is ready.
	WorkloadNetworkConditionReady = "Ready"

	// WorkloadNetworkConditionSystemReady indicates whether the system
	// NNC derived from the active provider is ready.
	WorkloadNetworkConditionSystemReady = "SystemNetworkConfigurationReady"
)

// SystemVDSNetworkConfig holds the VDS-specific configuration that the net-operator
// uses to derive the system NamespaceNetworkConfiguration for the vsphere-distributed
// provider.
type SystemVDSNetworkConfig struct {
	// network identifies the VSphereDistributedNetwork resource backing the system NNC.
	//
	// +required
	Network netv1alpha1.VSphereDistributedNetworkRef `json:"network"`
}

// NetworkProviderSystemConfig carries provider-specific system-level NNC configuration.
// Only the field corresponding to the parent provider type may be set; the per-entry CEL
// rules on NetworkProviderEntry enforce this constraint.
type NetworkProviderSystemConfig struct {
	// vdsConfig holds system-level configuration for the vsphere-distributed provider.
	// Must be set when the parent provider entry type is vsphere-distributed; must be
	// absent for all other provider types.
	//
	// +optional
	VDSConfig SystemVDSNetworkConfig `json:"vdsConfig,omitempty,omitzero"`
}

// +kubebuilder:validation:XValidation:rule="self.type != 'vsphere-distributed' || has(self.systemConfiguration.vdsConfig)",message="systemConfiguration.vdsConfig must be set when type is vsphere-distributed"
// +kubebuilder:validation:XValidation:rule="self.type == 'vsphere-distributed' || !has(self.systemConfiguration.vdsConfig)",message="systemConfiguration.vdsConfig may only be set when type is vsphere-distributed"

// NetworkProviderEntry pairs a network provider type with its system-level configuration.
// Exactly one entry per type is allowed (enforced via listType=map on the providers field).
type NetworkProviderEntry struct {
	// type identifies the network provider for this entry.
	//
	// +required
	Type netv1alpha1.NetworkProvider `json:"type,omitempty"`

	// systemConfiguration holds the system-level network configuration for this provider.
	// For provider types that require no system config (e.g. nsx-tier1), set this to an
	// empty object.
	//
	// +required
	SystemConfiguration NetworkProviderSystemConfig `json:"systemConfiguration"`
}

// +kubebuilder:validation:XValidation:rule="self.providers.exists(p, p.type == self.activeSystemProvider)",message="activeSystemProvider must reference a provider type declared in providers"

// WorkloadNetworkConfigurationSpec defines the desired state of the WorkloadNetworkConfiguration.
type WorkloadNetworkConfigurationSpec struct {
	// providers declares the set of network providers and their system-level configurations.
	// Each entry must have a unique type. At least one provider must be declared.
	//
	// +required
	// +kubebuilder:validation:MinItems=1
	// +kubebuilder:validation:MaxItems=3
	// +listType=map
	// +listMapKey=type
	Providers []NetworkProviderEntry `json:"providers,omitempty"`

	// activeSystemProvider identifies which provider in the providers list is currently
	// authoritative for deriving the system NamespaceNetworkConfiguration. Changing this
	// field triggers a transition of the system NNC to the newly active provider.
	//
	// +required
	ActiveSystemProvider netv1alpha1.NetworkProvider `json:"activeSystemProvider,omitempty"`
}

// WorkloadNetworkConfigurationStatus defines the observed state of the WorkloadNetworkConfiguration.
type WorkloadNetworkConfigurationStatus struct {
	// conditions represents the latest available observations of the
	// WorkloadNetworkConfiguration's current state.
	//
	// +optional
	// +listType=map
	// +listMapKey=type
	// +kubebuilder:validation:MaxItems=8
	Conditions []metav1.Condition `json:"conditions,omitempty"`
}

// +genclient
// +genclient:nonNamespaced
// +kubebuilder:object:root=true
// +kubebuilder:resource:scope=Cluster
// +kubebuilder:subresource:status
// +kubebuilder:validation:XValidation:rule="self.metadata.name == 'default'",message="WorkloadNetworkConfiguration must be named 'default'"

// WorkloadNetworkConfiguration is a singleton cluster-scoped resource that describes the
// network providers available in this Supervisor and which provider is currently active for
// system-level networking.
type WorkloadNetworkConfiguration struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// +required
	Spec   WorkloadNetworkConfigurationSpec   `json:"spec,omitempty"`
	Status WorkloadNetworkConfigurationStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// WorkloadNetworkConfigurationList contains a list of WorkloadNetworkConfiguration.
type WorkloadNetworkConfigurationList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []WorkloadNetworkConfiguration `json:"items"`
}

func init() {
	RegisterTypeWithScheme(&WorkloadNetworkConfiguration{}, &WorkloadNetworkConfigurationList{})
}
