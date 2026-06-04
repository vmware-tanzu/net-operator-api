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

// SharedSubnetDefault is a string enum used to mark a SharedSubnet as the
// default network for a workload type. A nil or absent value means this Subnet
// is not the default for that workload type.
//
// +kubebuilder:validation:Enum=True;False
type SharedSubnetDefault string

const (
	SharedSubnetDefaultTrue  SharedSubnetDefault = "True"
	SharedSubnetDefaultFalse SharedSubnetDefault = "False"
)

// SharedSubnet defines a pre-created Subnet to be associated with a Namespace.
//
// +kubebuilder:validation:XValidation:rule="oldSelf.path == '' || self.path == oldSelf.path",message="path is immutable once set"
// +kubebuilder:validation:XValidation:rule="oldSelf.name == '' || self.name == oldSelf.name",message="name is immutable once set"
type SharedSubnet struct {
	// path is the NSX policy path of the shared Subnet to be associated with
	// this Namespace. This field is immutable once set.
	//
	// +required
	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:validation:MaxLength=2048
	Path string `json:"path,omitempty"`

	// name is the name of the NSX Operator Subnet CR that will be created in
	// each associated Namespace to represent this shared Subnet. It serves as
	// the unique identifier for this entry in the sharedSubnets list and must
	// be unique across all entries. This field is immutable.
	//
	// +required
	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:validation:MaxLength=253
	// +kubebuilder:validation:Pattern=`^[a-z0-9]([-a-z0-9]*[a-z0-9])?(\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*$`
	Name string `json:"name,omitempty"`

	// podDefault indicates this Subnet is the default network for Pod workloads
	// in this Namespace. At most one entry in sharedSubnets may be set to True.
	// When unset, this Subnet is not a Pod default. When set to True, vpc must
	// be set on the parent VPCConfig, as the Subnet must reside in the
	// Namespace's VPC to support resources such as load balancer virtual
	// services and static routes.
	//
	// If no shared Subnet has podDefault set to True and the Namespace VPC has
	// sufficient PrivateTGW IP block space, Pods use Subnets generated from
	// that block. Otherwise, Pods are not assigned a default network.
	//
	// +optional
	PodDefault SharedSubnetDefault `json:"podDefault,omitempty"`

	// vmDefault indicates this Subnet is the default network for VM workloads
	// in this Namespace. At most one entry in sharedSubnets may be set to True.
	// When unset, this Subnet is not a VM default. When set to True, vpc must
	// be set on the parent VPCConfig, as the Subnet must reside in the
	// Namespace's VPC to support resources such as load balancer virtual
	// services and static routes.
	//
	// If no shared Subnet has vmDefault set to True and the Namespace VPC has
	// sufficient privateCIDR space, VMs use Subnets generated from those CIDRs.
	// Otherwise, there is no VM default network.
	//
	// +optional
	VMDefault SharedSubnetDefault `json:"vmDefault,omitempty"`
}

// AutoCreateVPCConfig specifies the configuration used to automatically create
// a namespace-scoped VPC.
//
// +kubebuilder:validation:XValidation:rule="oldSelf.nsxProject == '' || self.nsxProject == oldSelf.nsxProject",message="nsxProject is immutable once set"
// +kubebuilder:validation:XValidation:rule="oldSelf.vpcConnectivityProfile == '' || self.vpcConnectivityProfile == oldSelf.vpcConnectivityProfile",message="vpcConnectivityProfile is immutable once set"
// +kubebuilder:validation:XValidation:rule="!has(oldSelf.privateCIDRs) || oldSelf.privateCIDRs.all(cidr, self.privateCIDRs.exists(c, c == cidr))",message="privateCIDRs is append-only; existing entries cannot be removed"
type AutoCreateVPCConfig struct {
	// nsxProject is the NSX policy path of the Project the namespace is
	// associated with. This field is immutable once set.
	//
	// NSX Projects provide multi-tenancy by partitioning networking and
	// security configurations within a single NSX deployment.
	//
	// +optional
	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:validation:MaxLength=2048
	NSXProject string `json:"nsxProject,omitempty"`

	// vpcConnectivityProfile is the NSX policy path of the VPC Connectivity
	// Profile. This profile defines northbound connectivity configuration
	// for VPCs including:
	//   - Transit Gateway attachment
	//   - External IP blocks (for public subnets and external IP bindings)
	//   - Private Transit Gateway IP blocks (for inter-VPC communication)
	//
	// This field is immutable once set.
	//
	// +optional
	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:validation:MaxLength=2048
	VPCConnectivityProfile string `json:"vpcConnectivityProfile,omitempty"`

	// privateCIDRs specifies CIDR blocks from which private Subnets are
	// allocated for this namespace. These ranges should not overlap with:
	//   - CIDRs in the VPC connectivity profile
	//   - Kubernetes service CIDRs
	//   - Other services running in the datacenter
	//
	// This field is append-only; existing entries may not be removed.
	//
	// +optional
	// +kubebuilder:validation:MaxItems=16
	// +kubebuilder:validation:items:MaxLength=64
	// +listType=atomic
	PrivateCIDRs []string `json:"privateCIDRs,omitempty"`
}

// VPCConfig specifies the VPC network configuration for a namespace.
//
// There are two mutually exclusive modes:
//
//  1. Pre-created VPC mode: Set vpc to reference an existing VPC. Only
//     defaultSubnetSize and sharedSubnets take effect alongside vpc.
//
//  2. Auto-create VPC mode: Set autoCreateConfig to have a VPC automatically
//     created and scoped to this namespace.
//
// +kubebuilder:validation:MinProperties=1
// +kubebuilder:validation:XValidation:rule="!(has(self.vpc) && self.vpc != '' && has(self.autoCreateConfig))",message="vpc and autoCreateConfig are mutually exclusive; set vpc for pre-created VPC mode or autoCreateConfig for auto-create VPC mode"
// +kubebuilder:validation:XValidation:rule="oldSelf.vpc == '' || self.vpc == oldSelf.vpc",message="vpc is immutable once set"
// +kubebuilder:validation:XValidation:rule="!has(self.sharedSubnets) || self.sharedSubnets.filter(s, has(s.podDefault) && s.podDefault == 'True').size() <= 1",message="at most one sharedSubnet may have podDefault set to True"
// +kubebuilder:validation:XValidation:rule="!has(self.sharedSubnets) || self.sharedSubnets.filter(s, has(s.vmDefault) && s.vmDefault == 'True').size() <= 1",message="at most one sharedSubnet may have vmDefault set to True"
// +kubebuilder:validation:XValidation:rule="!has(self.sharedSubnets) || self.sharedSubnets.filter(s, has(s.podDefault) && s.podDefault == 'True').size() == 0 || (has(self.vpc) && self.vpc != '')",message="vpc must be set when any sharedSubnet has podDefault set to True"
// +kubebuilder:validation:XValidation:rule="!has(self.sharedSubnets) || self.sharedSubnets.filter(s, has(s.vmDefault) && s.vmDefault == 'True').size() == 0 || (has(self.vpc) && self.vpc != '')",message="vpc must be set when any sharedSubnet has vmDefault set to True"
type VPCConfig struct {
	// vpc is the NSX policy path of an existing VPC the namespace is associated
	// with. When set, the namespace uses this pre-created VPC and
	// autoCreateConfig must not be set. This field is immutable once set.
	//
	// +optional
	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:validation:MaxLength=2048
	VPC string `json:"vpc,omitempty"`

	// autoCreateConfig holds the configuration for automatically creating a
	// namespace-scoped VPC. Mutually exclusive with vpc.
	//
	// +optional
	AutoCreateConfig *AutoCreateVPCConfig `json:"autoCreateConfig,omitempty"`

	// sharedSubnets lists pre-created Subnets to be associated with this
	// Namespace. At most one entry may have podDefault set to True, and at
	// most one entry may have vmDefault set to True. These constraints are
	// enforced by validation rules; the API server will reject any update
	// that sets podDefault or vmDefault to True on more than one entry.
	//
	// A Subnet that is currently in use cannot be removed. If all shared Subnets
	// acting as a Pod or VM default are removed, the default network falls back
	// to Subnets generated from the Namespace VPC's available address space. If
	// no such space exists, the affected workload type is not assigned a default
	// network.
	//
	// +optional
	// +kubebuilder:validation:MaxItems=32
	// +listType=map
	// +listMapKey=name
	SharedSubnets []SharedSubnet `json:"sharedSubnets,omitempty"`

	// defaultSubnetSize is the default size of Namespace Subnets, specified as
	// the number of IP addresses. Must be a power of 2 (e.g. 16, 32, 64, 128).
	// When not set, defaults to 32 (equivalent to a /27 subnet).
	//
	// +optional
	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:validation:Maximum=65536
	// +kubebuilder:validation:XValidation:rule="self == 0 || self in [1, 2, 4, 8, 16, 32, 64, 128, 256, 512, 1024, 2048, 4096, 8192, 16384, 32768, 65536]",message="defaultSubnetSize must be a power of 2 (e.g. 1, 2, 4, 8, 16, 32, 64, 128, 256, 512, 1024, 2048, 4096, 8192, 16384, 32768, 65536)"
	DefaultSubnetSize int32 `json:"defaultSubnetSize,omitempty"`
}

// NamespaceNetworkSpec defines the desired network configuration
// for Namespaces associated with this NamespaceNetworkConfiguration.
//
// The type field selects the active network provider. For the vsphere-distributed
// provider, vsphereDistributedConfig must be populated. For the vpc provider,
// vpcConfig must be populated. Only the config section corresponding to the
// selected type may be populated; setting both vsphereDistributedConfig and
// vpcConfig is invalid and will be rejected by the API server.
//
// +kubebuilder:validation:XValidation:rule="self.type == 'vsphere-distributed' || self.type == 'vpc'",message="only vsphere-distributed and vpc are currently supported; nsx-tier1 will be introduced in a future version"
// +kubebuilder:validation:XValidation:rule="self.type == 'vsphere-distributed' ? (has(self.vsphereDistributedConfig.networks) && self.vsphereDistributedConfig.networks.size() > 0) : true",message="vsphereDistributedConfig.networks must contain at least one entry when type is vsphere-distributed"
// +kubebuilder:validation:XValidation:rule="self.type == 'vpc' ? (has(self.vpcConfig) && (self.vpcConfig.vpc != '' || has(self.vpcConfig.autoCreateConfig))) : true",message="vpcConfig must have either vpc (pre-created VPC mode) or autoCreateConfig (auto-create VPC mode) set when type is vpc"
// +kubebuilder:validation:XValidation:rule="!(has(self.vsphereDistributedConfig.networks) && self.vsphereDistributedConfig.networks.size() > 0 && has(self.vpcConfig) && ((has(self.vpcConfig.vpc) && self.vpcConfig.vpc != '') || has(self.vpcConfig.autoCreateConfig)))",message="vsphereDistributedConfig and vpcConfig are mutually exclusive; only the config section matching the selected type may be populated"
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

	// vpcConfig contains the VPC network configuration.
	// Required when type is vpc.
	//
	// When specified, network configuration is delegated to NSX Operator,
	// which constructs a VPCNetworkConfiguration from this spec and utilizes
	// the NSX Operator APIs to provision and manage the Namespace's VPC resources.
	//
	// +optional
	VPCConfig VPCConfig `json:"vpcConfig,omitempty,omitzero"`
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
// Deletion is blocked by the netoperator.vmware.com/nnc-protection finalizer
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
