// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: Apache-2.0

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

// +kubebuilder:validation:XValidation:rule="self.type == 'vsphere-distributed' || self.type == 'vpc' || self.type == 'nsx-tier1'",message="type must be one of: vsphere-distributed, vpc, nsx-tier1"
// +kubebuilder:validation:XValidation:rule="self.type != 'vsphere-distributed' || has(self.systemConfiguration.vsphereDistributedConfig)",message="systemConfiguration.vsphereDistributedConfig must be set when type is vsphere-distributed"
// +kubebuilder:validation:XValidation:rule="self.type == 'vsphere-distributed' || !has(self.systemConfiguration.vsphereDistributedConfig)",message="systemConfiguration.vsphereDistributedConfig may only be set when type is vsphere-distributed"
// +kubebuilder:validation:XValidation:rule="self.type != 'vpc' || has(self.systemConfiguration.vpcConfig)",message="systemConfiguration.vpcConfig must be set when type is vpc"
// +kubebuilder:validation:XValidation:rule="self.type == 'vpc' || !has(self.systemConfiguration.vpcConfig)",message="systemConfiguration.vpcConfig may only be set when type is vpc"
// +kubebuilder:validation:XValidation:rule="self.type == 'nsx-tier1' || !has(self.systemConfiguration.nsxTier1Config)",message="systemConfiguration.nsxTier1Config may only be set when type is nsx-tier1"

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

	// services defines cluster-wide workload network services including workload network DNS and NTP.
	//
	// +optional
	Services *WorkloadNetworkServicesConfig `json:"services,omitempty"`
}

// +kubebuilder:validation:XValidation:rule="(has(self.dns) && ((has(self.dns.servers) && size(self.dns.servers) > 0) || has(self.dns.maxConcurrentForwards))) || (has(self.ntp) && has(self.ntp.servers) && size(self.ntp.servers) > 0)",message="at least one of dns or ntp must be configured with valid settings"

// WorkloadNetworkServicesConfig defines configuration for workload DNS, forwarding, and NTP.
type WorkloadNetworkServicesConfig struct {
	// dns defines configuration for workload network DNS resolution and CoreDNS upstream forwarding.
	//
	// +optional
	DNS *WorkloadDNSConfig `json:"dns,omitempty"`

	// ntp defines configuration for default workload time synchronization.
	//
	// +optional
	NTP *WorkloadNTPConfig `json:"ntp,omitempty"`
}

// +kubebuilder:validation:XValidation:rule="(has(self.servers) && size(self.servers) > 0) || has(self.maxConcurrentForwards)",message="at least one of servers or maxConcurrentForwards must be specified"

// WorkloadDNSConfig specifies upstream nameservers and CoreDNS concurrency tuning.
type WorkloadDNSConfig struct {
	// servers is a list of upstream DNS server IP addresses used by CoreDNS and guest workloads.
	//
	// +optional
	// +kubebuilder:validation:MinItems=1
	// +kubebuilder:validation:MaxItems=10
	// +kubebuilder:validation:items:MinLength=1
	// +kubebuilder:validation:items:MaxLength=45
	// +kubebuilder:validation:items:XValidation:rule="isIP(self)",message="each server must be a valid IPv4 or IPv6 address"
	// +listType=set
	Servers []string `json:"servers,omitempty"`

	// maxConcurrentForwards specifies the maximum number of concurrent requests forwarded upstream by CoreDNS.
	// When unset, defaults to 1000. 0 indicates unconstrained concurrency.
	//
	// +optional
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:validation:Maximum=65535
	MaxConcurrentForwards *int32 `json:"maxConcurrentForwards,omitempty"`
}

// +kubebuilder:validation:XValidation:rule="has(self.servers) && size(self.servers) > 0",message="servers must be specified"

// WorkloadNTPConfig specifies workload NTP time synchronization sources.
type WorkloadNTPConfig struct {
	// servers is a list of default NTP server hostnames (FQDN) or IP addresses for VM workloads.
	//
	// +optional
	// +kubebuilder:validation:MinItems=1
	// +kubebuilder:validation:MaxItems=10
	// +kubebuilder:validation:items:MinLength=1
	// +kubebuilder:validation:items:MaxLength=253
	// +listType=set
	Servers []string `json:"servers,omitempty"`
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
