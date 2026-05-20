// Copyright (c) 2026 Broadcom. All Rights Reserved.
// Broadcom Confidential. The term "Broadcom" refers to Broadcom Inc.
// and/or its subsidiaries.

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	// NamespaceNetworkLabelKey is the label placed on a Namespace to associate
	// it with a specific NamespaceNetworkConfiguration. The value is the name
	// of the NamespaceNetworkConfiguration resource.
	NamespaceNetworkLabelKey = "netoperator.vmware.com/network-configuration"

	// NamespaceNetworkProtectionFinalizer is attached to a
	// NamespaceNetworkConfiguration by Net Operator to prevent deletion while
	// any Namespace holds the NamespaceNetworkLabelKey label pointing to this
	// resource.
	NamespaceNetworkProtectionFinalizer = "netoperator.vmware.com/nnc-protection"

	// NamespaceNetworkConditionReady is True when all networking resources owned
	// by the NamespaceNetworkConfiguration have been created and every associated
	// Namespace has been fully reconciled. When no Namespaces are associated,
	// readiness reflects whether all cluster-scoped resources were created
	// successfully.
	NamespaceNetworkConditionReady = "Ready"
)

// NamespaceNetworkReconciliationStatus is the reconciliation state of a single
// Namespace within a NamespaceNetworkConfiguration.
//
// +kubebuilder:validation:Enum=Reconciling;Reconciled
type NamespaceNetworkReconciliationStatus string

const (
	// NamespaceNetworkReconciling indicates that the network configuration is
	// being applied to the Namespace and is not yet complete.
	NamespaceNetworkReconciling NamespaceNetworkReconciliationStatus = "Reconciling"

	// NamespaceNetworkReconciled indicates that the network configuration has
	// been fully applied to the Namespace.
	NamespaceNetworkReconciled NamespaceNetworkReconciliationStatus = "Reconciled"
)

// VSphereDistributedConfig specifies the vSphere Distributed (VDS) network
// configuration for a namespace.
//
// +kubebuilder:validation:XValidation:rule="self.networks.exists(n, n.name == self.defaultNetwork)",message="defaultNetwork must match the name of one of the entries in networks"
// +kubebuilder:validation:XValidation:rule="oldSelf.defaultNetwork == '' || self.defaultNetwork == oldSelf.defaultNetwork",message="defaultNetwork is immutable once set"
type VSphereDistributedConfig struct {
	// networks lists the VSphereDistributedNetwork resources to present inside
	// each associated Namespace. Each entry results in a corresponding Network
	// created inside each associated Namespace, with a reference to the
	// corresponding VSphereDistributedNetwork. When an entry is removed from
	// this list, the corresponding Network is deleted from all associated
	// Namespaces. At least one entry is required; the entry referenced by
	// defaultNetwork cannot be removed once Namespaces are associated
	// with this NamespaceNetworkConfiguration.
	//
	// +required
	// +kubebuilder:validation:MinItems=1
	// +kubebuilder:validation:MaxItems=32
	// +listType=map
	// +listMapKey=name
	Networks []VSphereDistributedNetworkRef `json:"networks,omitempty"`

	// defaultNetwork is the name of one of the entries in networks. The
	// generated Network corresponding to this entry is labeled
	// netoperator.vmware.com/is-default: "true", making it the network
	// resolved by workloads that do not explicitly select a network.
	//
	// The referenced VSphereDistributedNetwork must not have an IP assignment
	// mode of None; a network with no IP assignment cannot serve as a workload
	// default.
	//
	// This field is immutable once set and cannot be removed while Namespaces
	// are associated with this NamespaceNetworkConfiguration.
	//
	// +required
	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:validation:MaxLength=253
	// +kubebuilder:validation:Pattern=`^[a-z0-9]([-a-z0-9]*[a-z0-9])?(\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*$`
	DefaultNetwork string `json:"defaultNetwork,omitempty"`
}

// NamespaceNetworkSpec defines the desired network configuration
// for Namespaces associated with this NamespaceNetworkConfiguration.
//
// The type field selects the active network provider. For the vsphere-distributed
// provider, vsphereDistributedConfig must be populated.
//
// +kubebuilder:validation:XValidation:rule="self.type == 'vsphere-distributed'",message="only vsphere-distributed is currently supported; nsx-tier1 and vpc will be introduced in a future version"
// +kubebuilder:validation:XValidation:rule="self.type == 'vsphere-distributed' ? (has(self.vsphereDistributedConfig.networks) && self.vsphereDistributedConfig.networks.size() > 0) : true",message="vsphereDistributedConfig.networks must contain at least one entry when type is vsphere-distributed"
type NamespaceNetworkSpec struct {
	// type selects the network provider for this configuration and determines
	// which provider-specific config section must be populated.
	//
	// +required
	Type NetworkProvider `json:"type,omitempty"`

	// vsphereDistributedConfig contains the vSphere Distributed (VDS) network
	// configuration. Required when type is vsphere-distributed.
	//
	// +optional
	VSphereDistributedConfig VSphereDistributedConfig `json:"vsphereDistributedConfig,omitempty,omitzero"`
}

// NamespaceNetworkAssociation describes the reconciliation state of a
// single Namespace associated with a NamespaceNetworkConfiguration.
type NamespaceNetworkAssociation struct {
	// name is the name of the associated Namespace.
	//
	// +required
	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:validation:MaxLength=63
	// +kubebuilder:validation:Pattern=`^[a-z0-9]([-a-z0-9]*[a-z0-9])?$`
	Name string `json:"name,omitempty"`

	// status is the reconciliation state for this Namespace.
	//
	// +required
	Status NamespaceNetworkReconciliationStatus `json:"status,omitempty"`

	// message provides a human-readable explanation when the Namespace has not
	// yet reached the Reconciled state.
	//
	// +optional
	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:validation:MaxLength=2048
	Message string `json:"message,omitempty"`
}

// NamespaceNetworkStatus describes the observed state of the
// NamespaceNetworkConfiguration.
type NamespaceNetworkStatus struct {
	// conditions describe the current state of the NamespaceNetworkConfiguration.
	// The Ready condition is True when all networking resources owned by this
	// configuration have been created and all associated Namespaces have been
	// fully reconciled.
	//
	// +optional
	// +kubebuilder:validation:MaxItems=32
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

	// associatedNamespaces lists each Namespace currently associated with this
	// NamespaceNetworkConfiguration and its individual reconciliation state. Net
	// Operator updates this list as Namespaces are attached or detached via the
	// NamespaceNetworkLabelKey label.
	//
	// +optional
	// +kubebuilder:validation:MaxItems=2048
	// +listType=map
	// +listMapKey=name
	AssociatedNamespaces []NamespaceNetworkAssociation `json:"associatedNamespaces,omitempty"`
}

// +genclient
// +genclient:nonNamespaced
// +kubebuilder:object:root=true
// +kubebuilder:resource:scope=Cluster,shortName=nnc
// +kubebuilder:subresource:status

// NamespaceNetworkConfiguration is a cluster-scoped resource that declares the
// network configuration to be applied to one or more Namespaces.
//
// A Namespace is associated with a NamespaceNetworkConfiguration by setting the
// netoperator.vmware.com/network-configuration label on the Namespace to the
// name of this resource.
//
// The spec.type field selects the network provider; the corresponding
// provider-specific config section must be populated to match. Net Operator
// watches this resource and the label on Namespaces, reconciles networking
// resources into associated Namespaces, and creates a NetworkSettings CR in each
// Namespace to expose the active provider to network-aware operators.
//
// Deletion is blocked by the
// netoperator.vmware.com/namespace-network-configuration-protection finalizer
// while any Namespace holds the netoperator.vmware.com/network-configuration
// label pointing to this resource.
//
// +kubebuilder:validation:XValidation:rule="size(self.metadata.name) <= 63",message="name must be 63 characters or fewer to be usable as a Kubernetes label value"
type NamespaceNetworkConfiguration struct {
	metav1.TypeMeta `json:",inline"`

	// metadata is the standard object metadata.
	//
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// spec defines the desired network configuration. When unset, no network
	// resource reconciliation will occur until populated.
	//
	// +optional
	Spec NamespaceNetworkSpec `json:"spec,omitempty,omitzero"`

	// status describes the observed state of the NamespaceNetworkConfiguration.
	//
	// +optional
	Status *NamespaceNetworkStatus `json:"status,omitempty"`
}

// GetConditions returns the status conditions for this NamespaceNetworkConfiguration.
func (n *NamespaceNetworkConfiguration) GetConditions() []metav1.Condition {
	if n.Status == nil {
		return nil
	}
	return n.Status.Conditions
}

// SetConditions sets the status conditions for this NamespaceNetworkConfiguration.
func (n *NamespaceNetworkConfiguration) SetConditions(conditions []metav1.Condition) {
	if n.Status == nil {
		n.Status = &NamespaceNetworkStatus{}
	}
	n.Status.Conditions = conditions
}

// +kubebuilder:object:root=true

// NamespaceNetworkConfigurationList contains a list of NamespaceNetworkConfiguration.
type NamespaceNetworkConfigurationList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []NamespaceNetworkConfiguration `json:"items"`
}

func init() {
	RegisterTypeWithScheme(&NamespaceNetworkConfiguration{}, &NamespaceNetworkConfigurationList{})
}
