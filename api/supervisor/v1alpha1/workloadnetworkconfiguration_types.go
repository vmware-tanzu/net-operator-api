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

// SystemVSphereDistributedConfig mirrors VSphereDistributedConfig from
// NamespaceNetworkSpec for use as a per-provider system-level NNC template.
// Defined here (rather than referenced from api/v1alpha1) so that the
// supervisor API can evolve independently of the per-namespace API.
type SystemVSphereDistributedConfig struct {
	// networks lists the VSphereDistributedNetwork resources that back the
	// system NamespaceNetworkConfiguration. At least one entry is required.
	//
	// +required
	// +kubebuilder:validation:MinItems=1
	// +kubebuilder:validation:MaxItems=32
	// +listType=map
	// +listMapKey=name
	Networks []netv1alpha1.VSphereDistributedNetworkRef `json:"networks,omitempty"`

	// defaultNetwork is the name of one of the entries in networks. The
	// generated Network corresponding to this entry is labeled
	// netoperator.vmware.com/is-default: "true".
	//
	// +required
	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:validation:MaxLength=253
	// +kubebuilder:validation:Pattern=`^[a-z0-9]([-a-z0-9]*[a-z0-9])?(\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*$`
	DefaultNetwork string `json:"defaultNetwork,omitempty"`
}

// NetworkProviderSystemConfig mirrors NamespaceNetworkSpec (without the type
// field, which is carried by the parent NetworkProviderEntry) as the
// system-level NNC template for a given provider. Only the field corresponding
// to the parent provider type may be populated; the per-entry CEL rules on
// NetworkProviderEntry enforce this constraint.
type NetworkProviderSystemConfig struct {
	// vsphereDistributedConfig holds the vSphere Distributed network
	// configuration for the system NamespaceNetworkConfiguration.
	// Must be set when the parent provider entry type is vsphere-distributed;
	// must be absent for all other provider types.
	//
	// +optional
	VSphereDistributedConfig SystemVSphereDistributedConfig `json:"vsphereDistributedConfig,omitempty,omitzero"`
}

// +kubebuilder:validation:XValidation:rule="self.type != 'vsphere-distributed' || has(self.systemConfiguration.vsphereDistributedConfig)",message="systemConfiguration.vsphereDistributedConfig must be set when type is vsphere-distributed"
// +kubebuilder:validation:XValidation:rule="self.type == 'vsphere-distributed' || !has(self.systemConfiguration.vsphereDistributedConfig)",message="systemConfiguration.vsphereDistributedConfig may only be set when type is vsphere-distributed"

// NetworkProviderEntry pairs a network provider type with its system-level configuration.
// Exactly one entry per type is allowed (enforced via listType=map on the providers field).
type NetworkProviderEntry struct {
	// type identifies the network provider for this entry.
	//
	// +required
	Type netv1alpha1.NetworkProvider `json:"type,omitempty"`

	// systemConfiguration holds the system-level NNC template for this provider,
	// mirroring NamespaceNetworkSpec without the redundant type field. For providers
	// that require no system config (e.g. nsx-tier1), an empty object is valid.
	//
	// +required
	SystemConfiguration *NetworkProviderSystemConfig `json:"systemConfiguration,omitempty"`
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
	metav1.TypeMeta `json:",inline"`
	// metadata carries standard Kubernetes object metadata.
	//
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// spec defines the desired state of this WorkloadNetworkConfiguration.
	//
	// +required
	Spec WorkloadNetworkConfigurationSpec `json:"spec,omitempty,omitzero"`

	// status describes the observed state of this WorkloadNetworkConfiguration.
	//
	// +optional
	Status *WorkloadNetworkConfigurationStatus `json:"status,omitempty"`
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
