// Copyright (c) 2026 Broadcom. All Rights Reserved.
// Broadcom Confidential. The term "Broadcom" refers to Broadcom Inc.
// and/or its subsidiaries.

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	// WorkloadNetworkConfigurationName is the singleton resource name.
	WorkloadNetworkConfigurationName = "default"

	// WorkloadNetworkConditionReady is the top-level aggregate condition.
	// It is True when all sub-conditions are True, giving operators and tooling
	// a single signal to wait on (e.g. kubectl wait --for=condition=Ready).
	WorkloadNetworkConditionReady = "Ready"

	// WorkloadNetworkConditionSystemReady indicates whether the system
	// NamespaceNetworkConfiguration derived from the active provider has been
	// successfully reconciled.
	WorkloadNetworkConditionSystemReady = "SystemNetworkConfigurationReady"

	// WorkloadNetworkReasonPending is set on a condition with status False when
	// reconciliation has not yet completed (initial state or in progress).
	WorkloadNetworkReasonPending = "Pending"

	// WorkloadNetworkReasonFailed is set on a condition with status False when
	// the controller encountered an error during reconciliation.
	WorkloadNetworkReasonFailed = "Failed"
)

// +kubebuilder:validation:XValidation:rule="self.type == 'vsphere-distributed' || self.type == 'vpc'",message="only vsphere-distributed and vpc are currently supported; nsx-tier1 will be introduced in a future version"
// +kubebuilder:validation:XValidation:rule="self.type != 'vsphere-distributed' || has(self.systemConfiguration.vsphereDistributedConfig)",message="systemConfiguration.vsphereDistributedConfig must be set when type is vsphere-distributed"
// +kubebuilder:validation:XValidation:rule="self.type == 'vsphere-distributed' || !has(self.systemConfiguration.vsphereDistributedConfig)",message="systemConfiguration.vsphereDistributedConfig may only be set when type is vsphere-distributed"
// +kubebuilder:validation:XValidation:rule="self.type != 'vpc' || has(self.systemConfiguration.vpcConfig)",message="systemConfiguration.vpcConfig must be set when type is vpc"
// +kubebuilder:validation:XValidation:rule="self.type == 'vpc' || !has(self.systemConfiguration.vpcConfig)",message="systemConfiguration.vpcConfig may only be set when type is vpc"
// +kubebuilder:validation:XValidation:rule="self.type == 'vpc' || !has(self.defaultNamespaceConfiguration) || !has(self.defaultNamespaceConfiguration.vpcConfig)",message="defaultNamespaceConfiguration.vpcConfig may only be set when type is vpc"

// NetworkProviderEntry pairs a network provider type with its system-level
// NamespaceNetworkConfiguration template. Exactly one entry per type is allowed
// (enforced via listType=map on the providers field).
type NetworkProviderEntry struct {
	// type identifies the network provider for this entry.
	//
	// +required
	Type NetworkProvider `json:"type,omitempty"`

	// systemConfiguration holds the provider-specific NNC template for this provider.
	//
	// +required
	SystemConfiguration *NamespaceNetworkConfig `json:"systemConfiguration,omitempty"`

	// defaultNamespaceConfiguration optionally overrides configuration values
	// applied when provisioning new namespaces under this provider, going
	// forward. It does not alter the configuration of namespaces already
	// associated under this provider. When unset, or when a given sub-field is
	// unset, the equivalent value from systemConfiguration is used instead.
	//
	// Values here follow different mutability rules than systemConfiguration:
	// they may be replaced freely and take effect only for namespaces onboarded
	// after the change.
	//
	// +optional
	DefaultNamespaceConfiguration NetworkProviderDefaultConfig `json:"defaultNamespaceConfiguration,omitempty,omitzero"`
}

// NetworkProviderDefaultConfig holds the subset of provider configuration that
// may be independently overridden for new namespaces, distinct from and
// unconstrained by the append-only/immutable rules that apply to
// systemConfiguration.
//
// +kubebuilder:validation:MinProperties=1
type NetworkProviderDefaultConfig struct {
	// vpcConfig overrides default values used when auto-creating VPCs for new
	// namespaces under the vpc provider.
	//
	// +optional
	VPCConfig *DefaultVPCConfig `json:"vpcConfig,omitempty"`
}

// DefaultVPCConfig holds VPC default values that may be changed independently
// of the vpc provider's systemConfiguration.
type DefaultVPCConfig struct {
	// privateCIDRs sets the CIDR blocks used as private-subnet/private-pod-IP
	// defaults for VPCs auto-created for namespaces onboarding after this is
	// set. Already-created namespace VPCs are unaffected.
	//
	// This value fully replaces (not appends to) the previous default and is
	// not subject to the append-only constraint that applies to
	// systemConfiguration.vpcConfig.autoCreateConfig.privateCIDRs. When unset,
	// new namespaces use systemConfiguration's privateCIDRs instead.
	//
	// +optional
	// +kubebuilder:validation:MinItems=1
	// +kubebuilder:validation:MaxItems=16
	// +kubebuilder:validation:items:MaxLength=64
	// +kubebuilder:validation:items:Pattern=`^([0-9]{1,3}\.){3}[0-9]{1,3}/[0-9]{1,2}$`
	// +listType=atomic
	PrivateCIDRs []string `json:"privateCIDRs,omitempty"`
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
	ActiveSystemProvider NetworkProvider `json:"activeSystemProvider,omitempty"`
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
	Spec WorkloadNetworkConfigurationSpec `json:"spec,omitzero"`

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
