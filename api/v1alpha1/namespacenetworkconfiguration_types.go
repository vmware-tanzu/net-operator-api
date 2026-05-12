// Copyright (c) 2026 Broadcom. All Rights Reserved.
// Broadcom Confidential. The term "Broadcom" refers to Broadcom Inc.
// and/or its subsidiaries.

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	// NamespaceNetworkConfigurationAnnotationKey is the annotation placed on a
	// Namespace by a privileged service account to associate the Namespace with a
	// specific NamespaceNetworkConfiguration.
	NamespaceNetworkConfigurationAnnotationKey = "netoperator.vmware.com/network-configuration"

	// NamespaceNetworkConfigurationDefaultLabelKey is the label applied to a
	// NamespaceNetworkConfiguration to mark it as the cluster-wide default for new
	// Namespaces.
	NamespaceNetworkConfigurationDefaultLabelKey = "netoperator.vmware.com/default"

	// NamespaceNetworkConfigurationProtectionFinalizer is attached to a
	// NamespaceNetworkConfiguration by Net Operator to block deletion while any
	// Namespace still holds the NamespaceNetworkConfigurationAnnotationKey
	// annotation pointing to this object.
	NamespaceNetworkConfigurationProtectionFinalizer = "netoperator.vmware.com/namespace-network-configuration-protection"

	// NamespaceNetworkConfigurationConditionReady is True when all cluster-scoped
	// resources owned by the NamespaceNetworkConfiguration have been created and
	// every associated Namespace has been fully reconciled.
	NamespaceNetworkConfigurationConditionReady = "Ready"
)

// NamespaceNetworkConfigurationReconciliationStatus is the reconciliation state
// of a single Namespace within a NamespaceNetworkConfiguration.
//
// +kubebuilder:validation:Enum=Reconciling;Reconciled
type NamespaceNetworkConfigurationReconciliationStatus string

const (
	// NamespaceNetworkConfigurationReconciling indicates that the network
	// configuration is being applied to the Namespace and is not yet complete.
	NamespaceNetworkConfigurationReconciling NamespaceNetworkConfigurationReconciliationStatus = "Reconciling"

	// NamespaceNetworkConfigurationReconciled indicates that the network
	// configuration has been fully applied to the Namespace.
	NamespaceNetworkConfigurationReconciled NamespaceNetworkConfigurationReconciliationStatus = "Reconciled"
)

// VPCSharedSubnetWorkloadType identifies a category of workloads for which a
// shared Subnet can be designated as the default.
//
// +kubebuilder:validation:Enum=Pods;VMs
type VPCSharedSubnetWorkloadType string

const (
	// VPCSharedSubnetWorkloadTypePods designates the Subnet as the default for
	// Pods in the namespace.
	VPCSharedSubnetWorkloadTypePods VPCSharedSubnetWorkloadType = "Pods"

	// VPCSharedSubnetWorkloadTypeVMs designates the Subnet as the default for
	// VirtualMachines in the namespace.
	VPCSharedSubnetWorkloadTypeVMs VPCSharedSubnetWorkloadType = "VMs"
)

// NSXTier1Mode specifies the traffic routing mode for an NSX Tier-1 namespace.
//
// +kubebuilder:validation:Enum=Routed;NAT
type NSXTier1Mode string

const (
	// NSXTier1ModeRouted indicates that traffic is routed without NAT.
	NSXTier1ModeRouted NSXTier1Mode = "Routed"

	// NSXTier1ModeNAT indicates that traffic is NATed. This is the default
	// when mode is unset.
	NSXTier1ModeNAT NSXTier1Mode = "NAT"
)

// NSXTier1LoadBalancerSize specifies the size of the NSX Load Balancer
// provisioned for a namespace.
//
// +kubebuilder:validation:Enum=Small;Medium;Large
type NSXTier1LoadBalancerSize string

const (
	// NSXTier1LoadBalancerSizeSmall is the default load balancer size.
	NSXTier1LoadBalancerSizeSmall NSXTier1LoadBalancerSize = "Small"

	// NSXTier1LoadBalancerSizeMedium is a medium load balancer.
	NSXTier1LoadBalancerSizeMedium NSXTier1LoadBalancerSize = "Medium"

	// NSXTier1LoadBalancerSizeLarge is the largest available load balancer.
	NSXTier1LoadBalancerSizeLarge NSXTier1LoadBalancerSize = "Large"
)

// VSphereDistributedNetworkRef is a reference to a VSphereDistributedNetwork
// resource.
type VSphereDistributedNetworkRef struct {
	// name is the name of the VSphereDistributedNetwork resource.
	//
	// +required
	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:validation:MaxLength=256
	Name string `json:"name,omitempty"`
}

// VSphereDistributedConfig specifies vSphere Distributed (VDS) networking for a
// namespace. Net Operator creates a Network CR in each associated Namespace for
// every entry in networks, and labels the Network corresponding to defaultNetwork
// with netoperator.vmware.com/is-default: "true" so that workloads without an
// explicit network selection resolve to the correct backing.
type VSphereDistributedConfig struct {
	// networks lists the VSphereDistributedNetwork resources to present inside
	// each associated Namespace. Each entry results in a corresponding Network CR
	// created and managed by Net Operator. When a Network is removed from this
	// list, Net Operator issues a DELETE to the corresponding Network CR from all
	// associated Namespaces.
	//
	// +optional
	// +kubebuilder:validation:MaxItems=32
	// +listType=map
	// +listMapKey=name
	Networks []VSphereDistributedNetworkRef `json:"networks,omitempty"`

	// defaultNetwork is the name of one of the VSphereDistributedNetwork
	// resources listed in networks. The Network CR for this entry is labeled
	// netoperator.vmware.com/is-default: "true", making it the network resolved
	// by workloads that do not explicitly specify a network. The default network
	// cannot be removed after Namespaces are associated with this NamespaceNetworkConfiguration.
	//
	// +optional
	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:validation:MaxLength=256
	DefaultNetwork string `json:"defaultNetwork,omitempty"`
}

// NSXTier1Config specifies the explicit NSX Tier-1 network configuration for a
// namespace, delegated by Net Operator to NCP.
//
// There are two NSX Tier-1 provisioning flows:
//
//   - Implicit T1 (NCP default): set spec.type to nsx-tier1 and omit
//     nsxTier1Config. NCP provisions networking using its default settings
//     inherited from the NCP ConfigMap.
//   - Explicit namespace network: populate the CIDR and gateway fields below.
//     Net Operator creates an NSXNetworkConfiguration CR owned by this
//     NamespaceNetworkConfiguration with these values. Omitted fields
//     inherit their NCP defaults.
//
// +kubebuilder:validation:XValidation:rule="self.mode != 'Routed' || !has(self.egressCIDRs) || self.egressCIDRs.size() == 0",message="egressCIDRs must not be set when mode is Routed"
type NSXTier1Config struct {
	// subnetCIDRs are the CIDR blocks from which Kubernetes allocates IP
	// addresses for all workloads in the namespace. These ranges must not overlap
	// with ingressCIDRs, egressCIDRs, or other services running in the Kubernetes cluster.
	//
	// Required when nsxTier0Gateway, ingressCIDRs, or egressCIDRs is set.
	// Updates may only append new CIDR blocks; existing entries cannot be removed.
	//
	// +optional
	// +kubebuilder:validation:MaxItems=16
	// +listType=atomic
	SubnetCIDRs []CIDR `json:"subnetCIDRs,omitempty"`

	// ingressCIDRs are the CIDR blocks from which NSX assigns IP addresses for
	// Kubernetes Ingresses and Services of type LoadBalancer. These ranges must
	// not overlap with subnetCIDRs, egressCIDRs, or other services running in
	// the Kubernetes cluster.
	//
	// Required when nsxTier0Gateway, subnetCIDRs, or egressCIDRs is set.
	// Updates may only append new CIDR blocks; existing entries cannot be removed.
	//
	// +optional
	// +kubebuilder:validation:MaxItems=16
	// +listType=atomic
	IngressCIDRs []CIDR `json:"ingressCIDRs,omitempty"`

	// egressCIDRs are the CIDR blocks from which NSX assigns IP addresses for
	// performing SNAT from container IPs to external IPs. These ranges must not
	// overlap with subnetCIDRs, ingressCIDRs, or other services running in the
	// Kubernetes cluster.
	//
	// Required when mode is NAT (or unset) and nsxTier0Gateway, subnetCIDRs,
	// or ingressCIDRs is set. Must not be set when mode is Routed.
	// Updates may only append new CIDR blocks; existing entries cannot be removed.
	//
	// +optional
	// +kubebuilder:validation:MaxItems=16
	// +listType=atomic
	EgressCIDRs []CIDR `json:"egressCIDRs,omitempty"`

	// nsxTier0Gateway is the NSX API path of the Tier-0 Gateway to use for this
	// namespace. If unset, the Tier-0 gateway configured for NCP in this Kubernetes
	// environment is used. This field cannot be changed after the namespace is created.
	//
	// +optional
	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:validation:MaxLength=256
	NSXTier0Gateway string `json:"nsxTier0Gateway,omitempty"`

	// subnetPrefixLength is the size of the subnet reserved for namespace
	// segments. If unset, the namespace subnet prefix configured for NCP in this
	// Kubernetes environment is used. This field cannot be changed after the namespace is created.
	//
	// +optional
	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:validation:Maximum=32
	SubnetPrefixLength int32 `json:"subnetPrefixLength,omitempty"`

	// mode specifies the traffic routing mode for the namespace. NAT (the
	// default when unset) enables SNAT translation from container IPs to
	// external IPs. Routed disables NAT, routing traffic without address
	// translation. This field cannot be changed after the namespace is created.
	//
	// +optional
	Mode NSXTier1Mode `json:"mode,omitempty"`

	// loadBalancerSize is the size of the NSX Load Balancer for this namespace.
	// If unset, defaults to Small. This field cannot be changed after the
	// namespace is created.
	//
	// +optional
	LoadBalancerSize NSXTier1LoadBalancerSize `json:"loadBalancerSize,omitempty"`
}

// VPCSharedSubnet describes a pre-existing NSX Subnet to associate with a
// namespace.
type VPCSharedSubnet struct {
	// path is the NSX API path of the shared Subnet as it appears on the NSX
	// Manager.
	//
	// +required
	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:validation:MaxLength=2048
	Path string `json:"path,omitempty"`

	// defaultFor lists the workload types (Pods, VMs) for which this shared
	// Subnet is the default in the namespace. vpc must be set on the parent
	// vpcConfig and this Subnet must belong to that VPC for any entry to take
	// effect.
	//
	// +optional
	// +kubebuilder:validation:MaxItems=2
	// +listType=set
	DefaultFor []VPCSharedSubnetWorkloadType `json:"defaultFor,omitempty"`

	// name is an optional stable identifier for this shared Subnet within the
	// namespace. Must be a valid RFC 1123 DNS subdomain. If unset, a unique name
	// is generated by Net Operator.
	//
	// +optional
	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:validation:MaxLength=253
	// +kubebuilder:validation:Pattern=`^[a-z0-9]([a-z0-9\-.]*[a-z0-9])?$`
	Name string `json:"name,omitempty"`
}

// VPCConfig specifies VPC networking for a namespace. Net Operator creates
// a VPCNetworkConfiguration CR owned by this NamespaceNetworkConfiguration and
// annotates each associated Namespace with nsx.vmware.com/vpc_network_config:
// <name> so that NSX Operator can realize VPC networking resources within the
// namespace.
//
// +kubebuilder:validation:XValidation:rule="self.vpc == ” || !has(self.privateCIDRs) || self.privateCIDRs.size() == 0",message="privateCIDRs may only be set for auto-created VPCs (vpc must be unset)"
type VPCConfig struct {
	// vpc is the NSX API path of a pre-created VPC. When set, Net Operator
	// uses this VPC directly without auto-creating one. Mutually exclusive
	// with privateCIDRs. When unset, NSX Operator auto-creates a private VPC
	// for this namespace scoped to nsxProject and vpcConnectivityProfile.
	//
	// +optional
	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:validation:MaxLength=2048
	VPC string `json:"vpc,omitempty"`

	// nsxProject is the NSX API path of the NSX Project this namespace should
	// be associated with. Used to determine the project for an auto-created
	// VPC. If both vpc and nsxProject are set, the vpc must belong to this
	// project.
	//
	// +optional
	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:validation:MaxLength=2048
	NSXProject string `json:"nsxProject,omitempty"`

	// vpcConnectivityProfile is the NSX API path of the VPC Connectivity
	// Profile for this namespace. Used when auto-creating a VPC. If both vpc
	// and vpcConnectivityProfile are set, the vpc must have been constructed
	// with this connectivity profile.
	//
	// +optional
	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:validation:MaxLength=2048
	VPCConnectivityProfile string `json:"vpcConnectivityProfile,omitempty"`

	// privateCIDRs specifies the CIDR blocks from which private Subnets are
	// allocated in an auto-created VPC. Must not be set when vpc is
	// specified.
	//
	// +optional
	// +kubebuilder:validation:MaxItems=16
	// +listType=atomic
	PrivateCIDRs []CIDR `json:"privateCIDRs,omitempty"`

	// defaultIPv4SubnetSize is the default address size for IPv4 Subnets
	// in this namespace. If unset, defaults to 32 addresses (/27).
	//
	// +optional
	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:validation:Maximum=65536
	DefaultIPv4SubnetSize int32 `json:"defaultIPv4SubnetSize,omitempty"`

	// sharedSubnets lists pre-existing NSX Subnets to associate with this
	// namespace.
	//
	// +optional
	// +kubebuilder:validation:MaxItems=32
	// +listType=map
	// +listMapKey=path
	SharedSubnets []VPCSharedSubnet `json:"sharedSubnets,omitempty"`
}

// NamespaceNetworkConfigurationSpec defines the desired network configuration
// for Namespaces associated with this NamespaceNetworkConfiguration.
//
// Exactly one provider-specific config section must be set and must correspond
// to the value of type. The one exception is nsx-tier1: nsxTier1Config may be
// omitted to indicate implicit Tier-1 provisioning by NCP.
//
// +kubebuilder:validation:XValidation:rule="(self.type == 'vsphere-distributed') ? (has(self.vsphereDistributedConfig) && !has(self.nsxTier1Config) && !has(self.vpcConfig)) : true",message="vsphereDistributedConfig must be set—and nsxTier1Config and vpcConfig must not be set—when type is vsphere-distributed"
// +kubebuilder:validation:XValidation:rule="!has(self.vsphereDistributedConfig) || (has(self.vsphereDistributedConfig.networks) && self.vsphereDistributedConfig.networks.size() > 0)",message="vsphereDistributedConfig.networks must contain at least one entry"
// +kubebuilder:validation:XValidation:rule="!has(self.vsphereDistributedConfig) || !has(self.vsphereDistributedConfig.defaultNetwork) || self.vsphereDistributedConfig.networks.exists(n, n.name == self.vsphereDistributedConfig.defaultNetwork)",message="vsphereDistributedConfig.defaultNetwork must reference one of the names listed in vsphereDistributedConfig.networks"
// +kubebuilder:validation:XValidation:rule="(self.type == 'nsx-tier1') ? (!has(self.vsphereDistributedConfig) && !has(self.vpcConfig)) : true",message="vsphereDistributedConfig and vpcConfig must not be set when type is nsx-tier1"
// +kubebuilder:validation:XValidation:rule="(self.type == 'vpc') ? (has(self.vpcConfig) && !has(self.vsphereDistributedConfig) && !has(self.nsxTier1Config)) : true",message="vpcConfig must be set—and vsphereDistributedConfig and nsxTier1Config must not be set—when type is vpc"
type NamespaceNetworkConfigurationSpec struct {
	// type selects the network provider for this configuration and determines
	// which provider-specific config section must be populated. Workloads and
	// network-aware operators should consult the NetworkSettings CR in the
	// namespace rather than this field to determine the active provider at runtime.
	//
	// +required
	// +kubebuilder:validation:Enum=vsphere-distributed;nsx-tier1;vpc
	Type NetworkSettingsProvider `json:"type,omitempty"`

	// vsphereDistributedConfig contains the vSphere Distributed (VDS) network
	// configuration. Required when type is vsphere-distributed; must not be set
	// for any other type.
	//
	// +optional
	VSphereDistributedConfig *VSphereDistributedConfig `json:"vsphereDistributedConfig,omitempty"`

	// nsxTier1Config contains the NSX Tier-1 network configuration. When type is
	// nsx-tier1, this field may be omitted (implicit NCP provisioning) or
	// populated with explicit CIDR and gateway overrides. Must not be set for any
	// other type.
	//
	// +optional
	NSXTier1Config *NSXTier1Config `json:"nsxTier1Config,omitempty"`

	// vpcConfig contains the NSX VPC network configuration. Required when type is
	// vpc; must not be set for any other type.
	//
	// +optional
	VPCConfig *VPCConfig `json:"vpcConfig,omitempty"`
}

// NamespaceNetworkConfigurationAppliedNamespace describes the reconciliation
// state of a single Namespace associated with this NamespaceNetworkConfiguration.
type NamespaceNetworkConfigurationAppliedNamespace struct {
	// name is the name of the associated Namespace.
	//
	// +required
	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:validation:MaxLength=253
	Name string `json:"name,omitempty"`

	// status is the reconciliation state for this Namespace.
	//
	// +required
	Status NamespaceNetworkConfigurationReconciliationStatus `json:"status,omitempty"`

	// message provides a human-readable explanation when the Namespace has not
	// yet reached the Reconciled state.
	//
	// +optional
	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:validation:MaxLength=1024
	Message string `json:"message,omitempty"`
}

// NamespaceNetworkConfigurationStatus describes the observed state of the
// NamespaceNetworkConfiguration.
type NamespaceNetworkConfigurationStatus struct {
	// conditions describe the current state of the NamespaceNetworkConfiguration.
	// The Ready condition is True when all cluster-scoped resources owned by this
	// configuration have been created and all associated Namespaces have been
	// reconciled. If no Namespaces are associated, readiness is determined by
	// whether all cluster-scoped resources have been created successfully.
	//
	// +optional
	// +kubebuilder:validation:MaxItems=8
	// +listType=map
	// +listMapKey=type
	Conditions []metav1.Condition `json:"conditions,omitempty"`

	// appliedToNamespaces lists each Namespace currently associated with this
	// NamespaceNetworkConfiguration and its individual reconciliation state. Net
	// Operator updates this list as Namespaces are attached or detached via the
	// NamespaceNetworkConfigurationAnnotationKey annotation.
	//
	// +optional
	// +kubebuilder:validation:MaxItems=1024
	// +listType=map
	// +listMapKey=name
	AppliedToNamespaces []NamespaceNetworkConfigurationAppliedNamespace `json:"appliedToNamespaces,omitempty"`
}

// +genclient
// +genclient:nonNamespaced
// +kubebuilder:object:root=true
// +kubebuilder:resource:scope=Cluster,shortName=nnc
// +kubebuilder:subresource:status

// NamespaceNetworkConfiguration is a cluster-scoped resource that declares the
// network configuration to be applied by Net Operator to one or more Namespaces.
//
// A Namespace is associated with a NamespaceNetworkConfiguration by annotating
// the Namespace with netoperator.vmware.com/network-configuration: <name>. Only
// Cluster Admin service accounts may set this annotation.
//
// The spec.type field selects the network provider (vsphere-distributed,
// nsx-tier1, or vpc); the corresponding config section in spec must be populated
// to match. Net Operator watches this CR and Namespace annotations, reconciles
// networking resources into associated Namespaces, and creates a NetworkSettings
// CR in each Namespace to expose the active provider to network-aware operators.
//
// Deletion is blocked by the
// netoperator.vmware.com/namespace-network-configuration-protection finalizer
// while any Namespace holds the netoperator.vmware.com/network-configuration
// annotation pointing to this object.
type NamespaceNetworkConfiguration struct {
	metav1.TypeMeta `json:",inline"`

	// metadata is the standard object's metadata.
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// spec defines the desired network configuration.
	//
	// +optional
	Spec NamespaceNetworkConfigurationSpec `json:"spec,omitempty,omitzero"`

	// status describes the observed state of the NamespaceNetworkConfiguration.
	//
	// +optional
	Status *NamespaceNetworkConfigurationStatus `json:"status,omitempty"`
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
		n.Status = &NamespaceNetworkConfigurationStatus{}
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
