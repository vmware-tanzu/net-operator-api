// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: Apache-2.0

package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	// NetworkInterfaceFinalizer allows the Controller to clean up resources associated
	// with a NetworkInterface before removing it from the API Server.
	NetworkInterfaceFinalizer = "networkinterface.netoperator.vmware.com"

	// NetworkInterfaceClientManagedAnnotation annotations means the NetworkInterface is
	// client managed and the Controller will not reconcile it. The value does not need
	// to be truthy; the presence of the key is what disables reconciliation.
	NetworkInterfaceClientManagedAnnotation = "networkinterface.netoperator.vmware.com/client-managed"
)

// IPConfig represents an IP configuration.
type IPConfig struct {
	// IP setting.
	IP string `json:"ip"`
	// IPFamily specifies the IP family (IPv4 vs IPv6) the IP belongs to.
	IPFamily corev1.IPFamily `json:"ipFamily"`
	// Gateway setting.
	Gateway string `json:"gateway"`
	// SubnetMask setting.
	// Deprecated: Use Prefix instead. If Prefix is set, SubnetMask is ignored.
	SubnetMask string `json:"subnetMask"`
	// Prefix is the prefix length for the IP address (e.g. 24 for a /24 IPv4 network,
	// 64 for a /64 IPv6 network). If set, this field takes precedence over SubnetMask
	// for both IPv4 and IPv6 addresses.
	// +optional
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:validation:Maximum=128
	Prefix *int32 `json:"prefix,omitempty"`
}

// NetworkInterfaceProviderReference contains info to locate a network interface provider object.
type NetworkInterfaceProviderReference struct {
	// APIGroup is the group for the resource being referenced.
	APIGroup string `json:"apiGroup"`
	// Kind is the type of resource being referenced
	Kind string `json:"kind"`
	// Name is the name of resource being referenced
	Name string `json:"name"`
	// API version of the referent.
	APIVersion string `json:"apiVersion,omitempty"`
}

type NetworkInterfaceConditionType string

const (
	// NetworkInterfaceReady is added when all network settings have been updated and the network
	// interface is ready to be used.
	NetworkInterfaceReady NetworkInterfaceConditionType = "Ready"
	// NetworkInterfaceFailure is added when network provider plugin returns an error.
	NetworkInterfaceFailure NetworkInterfaceConditionType = "Failure"
)

type NetworkInterfaceConditionReason string

const (
	// NetworkInterfaceFailureReasonCannotAllocIP indicates NetworkInterface is in failed state because an
	// IPConfig cannot be allocated.
	NetworkInterfaceFailureReasonCannotAllocIP NetworkInterfaceConditionReason = "CannotAllocIP"
	// NetworkInterfaceFailureReasonCannotAllocPort indicates NetworkInterface is in failed state because
	// port cannot be allocated for network interface on the network.
	NetworkInterfaceFailureReasonCannotAllocPort NetworkInterfaceConditionReason = "CannotAllocPort"
	// NetworkInterfaceFailureReasonNetworkDeleted indicates NetworkInterface is in failed state because
	// the underlying Network resource has been deleted.
	NetworkInterfaceFailureReasonNetworkDeleted NetworkInterfaceConditionReason = "NetworkDeleted"
	// NetworkInterfaceFailureReasonUnsupportedIPFamilyPolicy indicates NetworkInterface is in failed state
	// because the requested IPFamilyPolicy is not supported by the Network's SupportedIPFamilies.
	NetworkInterfaceFailureReasonUnsupportedIPFamilyPolicy NetworkInterfaceConditionReason = "UnsupportedIPFamilyPolicy"
	// NetworkInterfaceFailureReasonNoCandidateNetwork indicates NetworkInterface is in failed state
	// because no portgroup of the backing network is available on the cluster named by
	// filter.clusterMoID.
	NetworkInterfaceFailureReasonNoCandidateNetwork NetworkInterfaceConditionReason = "NoCandidateNetwork"
	// NetworkInterfaceFailureReasonAmbiguousPlacement indicates NetworkInterface is in failed state
	// because the backing network is made up of more than one portgroup and filter does not
	// give the network provider enough information to choose one.
	NetworkInterfaceFailureReasonAmbiguousPlacement NetworkInterfaceConditionReason = "AmbiguousPlacement"
)

// NetworkInterfaceCondition describes the state of a NetworkInterface at a certain point.
type NetworkInterfaceCondition struct {
	// Type is the type of network interface condition.
	Type NetworkInterfaceConditionType `json:"type"`
	// Status is the status of the condition.
	// Can be True, False, Unknown.
	Status corev1.ConditionStatus `json:"status"`
	// LastTransitionTime is the timestamp corresponding to the last status
	// change of this condition.
	LastTransitionTime metav1.Time `json:"lastTransitionTime,omitempty"`
	// Machine understandable string that gives the reason for condition's last transition.
	Reason NetworkInterfaceConditionReason `json:"reason,omitempty"`
	// Human-readable message indicating details about last transition.
	Message string `json:"message,omitempty"`
}

// NetworkInterfaceStatus defines the observed state of NetworkInterface.
// Once NetworkInterfaceReady condition is True, it should contain configuration to use to place
// a VM/Pod/Container's nic on the specified network.
type NetworkInterfaceStatus struct {
	// Conditions is an array of current observed network interface conditions.
	Conditions []NetworkInterfaceCondition `json:"conditions,omitempty"`
	// IPConfigs is an array of IP configurations for the network interface.
	IPConfigs []IPConfig `json:"ipConfigs,omitempty"`
	// MacAddress setting for the network interface.
	MacAddress string `json:"macAddress,omitempty"`
	// ExternalID is a network provider specific identifier assigned to the network interface.
	ExternalID string `json:"externalID,omitempty"`
	// NetworkID is an network provider specific identifier for the network backing the network
	// interface.
	NetworkID string `json:"networkID,omitempty"`
	// PortID is a network provider specific port identifier allocated for this network interface on
	// the backing network. It is only valid on requested node and is set only if port allocation
	// was requested.
	PortID string `json:"portID,omitempty"`
	// ConnectionID is a network provider specific port connection identifier allocated for this
	// network interface on the backing network. It is only valid on requested node and is set
	// only if port allocation was requested.
	ConnectionID string `json:"connectionID,omitempty"`
	// IPAssignmentMode indicates how IPv4 addresses are assigned to this interface.
	// When unset:
	// - If IP is assigned, it is assumed to be NetworkInterfaceIPAssignmentModeStaticPool.
	// - If IP is unassigned, it is assumed to be NetworkInterfaceIPAssignmentModeDHCP.
	// When set to NetworkInterfaceIPAssignmentModeStaticPool, indicates IP is assigned from a static pool.
	// When set to NetworkInterfaceIPAssignmentModeDHCP, indicates IP should be obtained via DHCP.
	// When set to NetworkInterfaceIPAssignmentModeNone, indicates no IP assignment should be performed.
	// +optional
	IPAssignmentMode NetworkInterfaceIPAssignmentMode `json:"ipAssignmentMode,omitempty"`
	// IPv6AssignmentMode indicates how IPv6 addresses are assigned to this interface.
	// This field is independent of IPAssignmentMode, allowing different assignment modes for IPv4
	// and IPv6 (e.g., IPv4 uses DHCP while IPv6 uses static pool).
	// When unset, defaults to IPAssignmentModeNone (IPv6 disabled) for backward compatibility with existing IPv4-only
	// deployments.
	// When set to NetworkInterfaceIPAssignmentModeStaticPool, indicates IPv6 is assigned from a static pool.
	// When set to NetworkInterfaceIPAssignmentModeDHCP, indicates IPv6 should be obtained via DHCPv6.
	// When set to NetworkInterfaceIPAssignmentModeNone, indicates no IPv6 assignment should be performed.
	// +optional
	IPv6AssignmentMode NetworkInterfaceIPAssignmentMode `json:"ipv6AssignmentMode,omitempty"`
}

type NetworkInterfaceType string

const (
	// NetworkInterfaceTypeVMXNet3 is for a VMXNET3 device.
	NetworkInterfaceTypeVMXNet3 = NetworkInterfaceType("vmxnet3")
)

// NetworkInterfaceIPAssignmentMode defines how IP addresses are assigned to a network interface
type NetworkInterfaceIPAssignmentMode string

const (
	// NetworkInterfaceIPAssignmentModeStaticPool indicates IP address is assigned from a static pool.
	NetworkInterfaceIPAssignmentModeStaticPool NetworkInterfaceIPAssignmentMode = "staticpool"

	// NetworkInterfaceIPAssignmentModeDHCP indicates IP address should be obtained via DHCP.
	NetworkInterfaceIPAssignmentModeDHCP NetworkInterfaceIPAssignmentMode = "dhcp"

	// NetworkInterfaceIPAssignmentModeNone indicates no IP assignment should be performed.
	NetworkInterfaceIPAssignmentModeNone NetworkInterfaceIPAssignmentMode = "none"
)

// NetworkInterfaceIPFamilyPolicy defines the IP family policy for a network interface.
type NetworkInterfaceIPFamilyPolicy string

const (
	// NetworkInterfaceIPFamilyPolicyIPv4Only indicates only IPv4 addresses will be allocated.
	NetworkInterfaceIPFamilyPolicyIPv4Only NetworkInterfaceIPFamilyPolicy = "IPv4Only"
	// NetworkInterfaceIPFamilyPolicyIPv6Only indicates only IPv6 addresses will be allocated.
	NetworkInterfaceIPFamilyPolicyIPv6Only NetworkInterfaceIPFamilyPolicy = "IPv6Only"
	// NetworkInterfaceIPFamilyPolicyDualStack indicates both IPv4 and IPv6 addresses will be allocated.
	NetworkInterfaceIPFamilyPolicyDualStack NetworkInterfaceIPFamilyPolicy = "DualStack"
)

// NetworkInterfacePortAllocation describes the settings for network interface port allocation request.
type NetworkInterfacePortAllocation struct {
	// NodeName is the node where port must be allocated for this network interface.
	NodeName string `json:"nodeName"`
}

// NetworkInterfaceSpec defines the desired state of NetworkInterface.
//
// +kubebuilder:validation:XValidation:rule="!has(self.ipFamilyPolicy) || self.ipFamilyPolicy != 'IPv4Only' || !has(self.requestedIPs) || self.requestedIPs.all(x, !isIP(x) || ip(x).family() == 4)",message="requestedIPs must only contain IPv4 addresses when ipFamilyPolicy is IPv4Only"
// +kubebuilder:validation:XValidation:rule="!has(self.ipFamilyPolicy) || self.ipFamilyPolicy != 'IPv6Only' || !has(self.requestedIPs) || self.requestedIPs.all(x, !isIP(x) || ip(x).family() == 6)",message="requestedIPs must only contain IPv6 addresses when ipFamilyPolicy is IPv6Only"
type NetworkInterfaceSpec struct {
	// NetworkName refers to a NetworkObject in the same namespace.
	NetworkName string `json:"networkName,omitempty"`
	// Type is the type of NetworkInterface. Supported values are vmxnet3.
	Type NetworkInterfaceType `json:"type,omitempty"`
	// ProviderRef is a reference to a provider specific network interface object
	// that specifies the network interface configuration.
	// If unset, default configuration is assumed.
	ProviderRef *NetworkInterfaceProviderReference `json:"providerRef,omitempty"`
	// PortAllocation is a request to allocate a port for this network interface on the backing network.
	// This feature is currently supported only if backing network type is NetworkTypeVDS. In all other
	// cases this field is ignored. Typically this is done implicitly by vCenter Server at the time
	// of attaching a network interface to a network and should be left unset. This is used primarily when
	// attachment of network interface to the network is done without vCenter Server's knowledge.
	PortAllocation *NetworkInterfacePortAllocation `json:"portAllocation,omitempty"`
	// ExternalID describes a value that will be surfaced as status.externalID.
	// If this field is omitted, then it is up to the underlying network
	// provider to surface any information in status.externalID.
	// +optional
	ExternalID string `json:"externalID,omitempty"`
	// IPFamilyPolicy specifies the IP family policy for this network interface.
	// Values: IPv4Only, IPv6Only, DualStack.
	// When set to IPv4Only, only an IPv4 address will be allocated.
	// When set to IPv6Only, only an IPv6 address will be allocated.
	// When set to DualStack, both IPv4 and IPv6 addresses will be allocated.
	// If not specified, the allocation is determined by the IP families available in the
	// IPPools referenced by the backing Network: if both IPv4 and IPv6 pools are present,
	// one address per IP family will be allocated (equivalent to DualStack); if only a
	// single IP family is available, only one address of that family will be allocated.
	// Users can discover the supported IP families by inspecting the SupportedIPFamilies
	// field on the Network object.
	// +optional
	// +kubebuilder:validation:Enum=IPv4Only;IPv6Only;DualStack
	IPFamilyPolicy NetworkInterfaceIPFamilyPolicy `json:"ipFamilyPolicy,omitempty"`
	// requestedIPs is an optional list of specific IP addresses to allocate to this network
	// interface. At most one IPv4 and one IPv6 address may be requested. When IPFamilyPolicy
	// is IPv4Only, requestedIPs must only contain IPv4 addresses; when IPFamilyPolicy is
	// IPv6Only, requestedIPs must only contain IPv6 addresses. If omitted, IP addresses are
	// allocated automatically.
	//
	// requestedIPs is only honored for IP families whose IP allocation mode on the backing
	// Network is static pool. If an entry's family uses DHCP, the Network has no IPPool
	// configured for that family, the entry falls outside every configured IPPool's range, or
	// the entry is already reserved by another IPAM consumer, the NetworkInterface fails to
	// become Ready (Failure condition reason CannotAllocIP).
	// +optional
	// +listType=atomic
	// +kubebuilder:validation:MaxItems=2
	// +kubebuilder:validation:items:MinLength=3
	// +kubebuilder:validation:items:MaxLength=39
	// +kubebuilder:validation:items:XValidation:rule="isIP(self)",message="each requestedIP must be a valid IPv4 or IPv6 address"
	// +kubebuilder:validation:XValidation:rule="!self.all(x, isIP(x)) || size(self) <= 1 || ip(self[0]).family() != ip(self[1]).family()",message="requestedIPs must not contain two addresses of the same IP family"
	RequestedIPs []string `json:"requestedIPs,omitempty"`
	// filter constrains which backing network the network provider selects for this network
	// interface, when the Network is backed by more than one portgroup. If unset and the Network
	// is backed by more than one portgroup, the NetworkInterface does not become Ready: its Ready
	// condition is False and its Failure condition has reason AmbiguousPlacement. If unset and the
	// Network is backed by a single portgroup, that portgroup is used.
	// +optional
	Filter *NetworkInterfaceFilter `json:"filter,omitempty"`
}

// NetworkInterfaceFilter describes constraints on the backing network that the network provider
// selects for a NetworkInterface.
type NetworkInterfaceFilter struct {
	// clusterMoID is the managed object ID of the vSphere cluster (for example domain-c8) on which
	// the selected portgroup must be available. Changing it has no effect once status.networkID
	// is set.
	// +optional
	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:validation:MaxLength=64
	ClusterMoID string `json:"clusterMoID,omitempty"`
}

// NetworkInterfaceReference is an object that points to a NetworkInterface.
type NetworkInterfaceReference struct {
	// Kind is the type of resource being referenced.
	Kind string `json:"kind"`
	// Name is the name of resource being referenced.
	Name string `json:"name"`
	// APIVersion of the referent.
	//
	// +optional
	APIVersion string `json:"apiVersion,omitempty"`
}

// +genclient
// +kubebuilder:object:root=true

// NetworkInterface is the Schema for the networkinterfaces API.
// A NetworkInterface represents a user's request for network configuration to use to place a
// VM/Pod/Container's nic on a specified network.
type NetworkInterface struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   NetworkInterfaceSpec   `json:"spec,omitempty"`
	Status NetworkInterfaceStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// NetworkInterfaceList contains a list of NetworkInterface
type NetworkInterfaceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []NetworkInterface `json:"items"`
}

func init() {
	RegisterTypeWithScheme(&NetworkInterface{}, &NetworkInterfaceList{})
}
