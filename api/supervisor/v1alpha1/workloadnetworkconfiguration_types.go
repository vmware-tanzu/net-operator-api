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

// WorkloadNetworkConfigurationSpec defines the desired global network
// configuration for a Supervisor cluster.
type WorkloadNetworkConfigurationSpec struct {
	// systemNetworkConfiguration declares the network configuration used to
	// create and manage the cluster's system NamespaceNetworkConfiguration.
	// Net Operator reconciles a cluster-scoped NamespaceNetworkConfiguration
	// whose spec mirrors this field.
	//
	// +required
	SystemNetworkConfiguration netv1alpha1.NamespaceNetworkSpec `json:"systemNetworkConfiguration,omitempty,omitzero"`
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
// declares the global network configuration for a Supervisor cluster. Its
// spec drives the creation and lifecycle of the cluster's system
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
