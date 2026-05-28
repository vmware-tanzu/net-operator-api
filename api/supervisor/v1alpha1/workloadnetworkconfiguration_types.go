// Copyright (c) 2026 Broadcom. All Rights Reserved.
// Broadcom Confidential. The term "Broadcom" refers to Broadcom Inc.
// and/or its subsidiaries.

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	netv1alpha1 "github.com/vmware-tanzu/net-operator-api/api/v1alpha1"
)

const (
	// WorkloadNetworkConfigurationName is the required name for the singleton
	// WorkloadNetworkConfiguration instance per cluster.
	WorkloadNetworkConfigurationName = "default"

	// WorkloadNetworkConditionReady is True when the WorkloadNetworkConfiguration
	// has been fully reconciled.
	WorkloadNetworkConditionReady = "Ready"

	// WorkloadNetworkConditionSystemReady is True when the system
	// NamespaceNetworkConfiguration managed by this resource has been
	// successfully created or updated and is Ready.
	WorkloadNetworkConditionSystemReady = "SystemNetworkConfigurationReady"
)

// NetworkProviderEntry declares a single network provider that should be
// installed and available in this cluster.
type NetworkProviderEntry struct {
	// type identifies the network provider. Each provider may appear at most
	// once in the providers list.
	//
	// +required
	Type netv1alpha1.NetworkProvider `json:"type,omitempty"`
}

// SystemVDSNetworkConfig holds the vSphere Distributed Switch (VDS)
// configuration for the system NamespaceNetworkConfiguration.
type SystemVDSNetworkConfig struct {
	// network is a reference to the VSphereDistributedNetwork resource that
	// backs the system NamespaceNetworkConfiguration.
	//
	// +required
	Network netv1alpha1.VSphereDistributedNetworkRef `json:"network,omitempty"`
}

// SystemNetworkConfiguration declares the system-level network configuration
// that drives the creation and management of the system
// NamespaceNetworkConfiguration for this cluster.
//
// +kubebuilder:validation:XValidation:rule="self.provider != 'vsphere-distributed' || has(self.vdsConfig)",message="vdsConfig must be set when provider is vsphere-distributed"
// +kubebuilder:validation:XValidation:rule="self.provider == 'vsphere-distributed' || !has(self.vdsConfig)",message="vdsConfig may only be set when provider is vsphere-distributed"
type SystemNetworkConfiguration struct {
	// provider selects the network provider for the system
	// NamespaceNetworkConfiguration. The corresponding provider-specific
	// config section must be populated.
	//
	// +required
	Provider netv1alpha1.NetworkProvider `json:"provider,omitempty"`

	// vdsConfig holds system-level configuration for the vsphere-distributed
	// provider. Required when provider is vsphere-distributed, and must not be
	// set otherwise.
	//
	// +optional
	VDSConfig SystemVDSNetworkConfig `json:"vdsConfig,omitempty,omitzero"`
}

// WorkloadNetworkConfigurationSpec defines the desired global network
// configuration for a Supervisor cluster.
//
// +kubebuilder:validation:XValidation:rule="self.providers.exists(p, p.type == self.systemConfiguration.provider)",message="systemConfiguration.provider must be listed in providers"
type WorkloadNetworkConfigurationSpec struct {
	// providers lists the network providers that should be installed and
	// available in this cluster. Each entry identifies a provider whose
	// corresponding operator and infrastructure are expected to be present.
	// Each provider may appear at most once.
	//
	// +required
	// +kubebuilder:validation:MinItems=1
	// +kubebuilder:validation:MaxItems=3
	// +listType=map
	// +listMapKey=type
	Providers []NetworkProviderEntry `json:"providers,omitempty"`

	// systemConfiguration declares the system-level network configuration
	// and drives the creation and lifecycle of the system
	// NamespaceNetworkConfiguration for the cluster.
	//
	// +required
	SystemConfiguration SystemNetworkConfiguration `json:"systemConfiguration,omitzero"`
}

// WorkloadNetworkConfigurationStatus describes the observed state of a
// WorkloadNetworkConfiguration.
type WorkloadNetworkConfigurationStatus struct {
	// conditions describe the current state of the WorkloadNetworkConfiguration.
	// Each condition's observedGeneration field records the spec generation it
	// was computed from.
	//
	// +optional
	// +kubebuilder:validation:MaxItems=8
	// +listType=map
	// +listMapKey=type
	Conditions []metav1.Condition `json:"conditions,omitempty"`
}

// +genclient
// +genclient:nonNamespaced
// +kubebuilder:object:root=true
// +kubebuilder:resource:scope=Cluster,shortName=wnc
// +kubebuilder:subresource:status
// +kubebuilder:validation:XValidation:rule="self.metadata.name == 'default'",message="WorkloadNetworkConfiguration must be named 'default'"

// WorkloadNetworkConfiguration is a cluster-scoped Supervisor resource that
// declares the global network configuration for a Supervisor. It
// identifies which network providers are active, holds system-level
// provider-specific settings, and drives the lifecycle of the system
// NamespaceNetworkConfiguration.
//
// Exactly one instance of this resource must exist per cluster, and it must be
// named "default".
type WorkloadNetworkConfiguration struct {
	metav1.TypeMeta `json:",inline"`

	// metadata is the standard object metadata.
	//
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// spec defines the desired global workload network configuration.
	//
	// +required
	Spec WorkloadNetworkConfigurationSpec `json:"spec,omitzero"`

	// status describes the observed state of the WorkloadNetworkConfiguration.
	//
	// +optional
	Status *WorkloadNetworkConfigurationStatus `json:"status,omitempty"`
}

// GetConditions returns the status conditions for this WorkloadNetworkConfiguration.
func (w *WorkloadNetworkConfiguration) GetConditions() []metav1.Condition {
	if w.Status == nil {
		return nil
	}
	return w.Status.Conditions
}

// SetConditions sets the status conditions for this WorkloadNetworkConfiguration.
func (w *WorkloadNetworkConfiguration) SetConditions(conditions []metav1.Condition) {
	if w.Status == nil {
		w.Status = &WorkloadNetworkConfigurationStatus{}
	}
	w.Status.Conditions = conditions
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
