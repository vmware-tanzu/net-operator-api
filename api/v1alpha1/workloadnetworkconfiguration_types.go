// Copyright (c) 2026 Broadcom. All Rights Reserved.
// Broadcom Confidential. The term "Broadcom" refers to Broadcom Inc.
// and/or its subsidiaries.

package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	// WorkloadNetworkConditionReady is True when the WorkloadNetworkConfiguration
	// has been fully reconciled.
	WorkloadNetworkConditionReady = "Ready"

	// WorkloadNetworkConditionSystemReady is True when the system
	// NamespaceNetworkConfiguration managed by this resource has been
	// successfully created or updated and is Ready.
	WorkloadNetworkConditionSystemReady = "SystemNetworkConfigurationReady"
)

// VDSNetworkConfig holds the common configuration for the vSphere Distributed
// Switch (VDS) network provider.
type VDSNetworkConfig struct {
	// primaryNetwork is a reference to the VSphereDistributedNetwork resource
	// that backs this configuration. It is used to populate the networks and
	// defaultNetwork fields of the system NamespaceNetworkConfiguration, and
	// the networks and defaultNetwork fields of the default namespace and
	// supervisor service NamespaceNetworkConfiguration specs.
	//
	// +required
	PrimaryNetwork VSphereDistributedNetworkRef `json:"primaryNetwork,omitempty,omitzero"`
}

// WorkloadNetworkConfigurationSpec defines the desired global network
// configuration for a Kubernetes cluster.
//
// +kubebuilder:validation:XValidation:rule="self.providers.exists(p, p == 'vsphere-distributed') || !has(self.vdsConfiguration)",message="vdsConfiguration may only be set when vsphere-distributed is listed in providers"
type WorkloadNetworkConfigurationSpec struct {
	// providers declares the network providers that should be installed and
	// available in this cluster. Each entry indicates a network provider whose
	// corresponding operator and infrastructure are expected to be present.
	//
	// +required
	// +kubebuilder:validation:MinItems=1
	// +kubebuilder:validation:MaxItems=3
	// +listType=set
	Providers []NetworkProvider `json:"providers,omitempty"`

	// vdsConfiguration holds global configuration for the vSphere Distributed
	// Switch network provider. May only be set when vsphere-distributed is
	// listed in providers.
	//
	// +optional
	VDSConfiguration VDSNetworkConfig `json:"vdsConfiguration,omitempty,omitzero"`
}

// WorkloadNetworkConfigurationStatus describes the observed state of a
// WorkloadNetworkConfiguration.
type WorkloadNetworkConfigurationStatus struct {
	// conditions describe the current state of the WorkloadNetworkConfiguration.
	//
	// +optional
	// +kubebuilder:validation:MaxItems=8
	// +listType=map
	// +listMapKey=type
	Conditions []metav1.Condition `json:"conditions,omitempty"`

	// observedGeneration is the metadata.generation that this status was
	// computed from. Clients must compare this against metadata.generation
	// before relying on any field in status; if they differ, the status
	// reflects a prior spec version and should be treated as stale.
	//
	// +optional
	ObservedGeneration *int64 `json:"observedGeneration,omitempty"`

	// systemNetworkConfigurationRef is an object reference to the system
	// NamespaceNetworkConfiguration that this controller creates and manages.
	// Populated once the system NamespaceNetworkConfiguration has been
	// successfully reconciled.
	//
	// +optional
	SystemNetworkConfigurationRef *corev1.ObjectReference `json:"systemNetworkConfigurationRef,omitempty"`

	// defaultNamespaceTemplate reflects the effective network configuration that
	// should be applied to namespaces by default.
	//
	// +optional
	DefaultNamespaceTemplate NamespaceNetworkSpec `json:"defaultNamespaceTemplate,omitempty,omitzero"`

	// supervisorServiceTemplate reflects the effective network configuration that
	// should be applied to Supervisor Service namespaces by default.
	//
	// +optional
	SupervisorServiceTemplate NamespaceNetworkSpec `json:"supervisorServiceTemplate,omitempty,omitzero"`
}

// +genclient
// +genclient:nonNamespaced
// +kubebuilder:object:root=true
// +kubebuilder:resource:scope=Cluster,shortName=wnc
// +kubebuilder:subresource:status

// WorkloadNetworkConfiguration is a cluster-scoped resource that declares the
// global network configuration for a Kubernetes cluster. It identifies which
// network providers are active, holds provider-specific global settings, and
// drives lifecycle management of the system NamespaceNetworkConfiguration. It
// also reflects derived configuration data for default-namespace and
// supervisor-services networking.
//
// Only one instance of this resource, named "default", is expected per cluster.
type WorkloadNetworkConfiguration struct {
	metav1.TypeMeta `json:",inline"`

	// metadata is the standard object metadata.
	//
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// spec defines the desired global workload network configuration. When
	// unset, no network resource reconciliation will occur until populated.
	//
	// +optional
	Spec WorkloadNetworkConfigurationSpec `json:"spec,omitempty,omitzero"`

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
