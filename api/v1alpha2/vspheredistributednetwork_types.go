// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: Apache-2.0

package v1alpha2

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	// VSphereDistributedNetworkPortGroupFailure is the condition type set when a portgroup named by
	// spec.portGroupID or spec.portGroupIDs doesn't exist, or vCenter Server can't be reached.
	VSphereDistributedNetworkPortGroupFailure = "PortGroupFailure"
	// VSphereDistributedNetworkIPPoolInvalid is the condition type set when the network has no
	// usable IPPools.
	VSphereDistributedNetworkIPPoolInvalid = "IPPoolInvalid"
	// VSphereDistributedNetworkIPPoolPressure is the condition type set to True when the network's
	// IPPools are low on free addresses.
	VSphereDistributedNetworkIPPoolPressure = "IPPoolPressure"
)

const (
	// VSphereDistributedNetworkReasonPortGroupNotFound is used with VSphereDistributedNetworkPortGroupFailure
	// when a portgroup named by portGroupID or portGroupIDs does not exist.
	VSphereDistributedNetworkReasonPortGroupNotFound = "PortGroupNotFound"
	// VSphereDistributedNetworkReasonPortGroupConfigMismatch is used with VSphereDistributedNetworkPortGroupFailure
	// when the portgroups named by portGroupIDs do not share the same default port configuration.
	VSphereDistributedNetworkReasonPortGroupConfigMismatch = "PortGroupConfigMismatch"
	// VSphereDistributedNetworkReasonIPPoolReconcileFailed is used with VSphereDistributedNetworkIPPoolInvalid
	// when the IPPools for the address ranges could not be created, updated, or deleted.
	VSphereDistributedNetworkReasonIPPoolReconcileFailed = "IPPoolReconcileFailed"
)

// IPAssignmentModeType is how addresses of one IP family are assigned on a network.
//
// +kubebuilder:validation:Enum=dhcp;staticpool;none
type IPAssignmentModeType string

const (
	// IPAssignmentModeDHCP indicates addresses are assigned dynamically using DHCP.
	IPAssignmentModeDHCP IPAssignmentModeType = "dhcp"
	// IPAssignmentModeStaticPool indicates addresses are assigned from the network's address ranges.
	IPAssignmentModeStaticPool IPAssignmentModeType = "staticpool"
	// IPAssignmentModeNone indicates no address is assigned and no DHCP client is configured.
	IPAssignmentModeNone IPAssignmentModeType = "none"
)

// VSphereDistributedNetworkIPRange is a contiguous block of addresses that a network assigns from.
type VSphereDistributedNetworkIPRange struct {
	// startingAddress is the first address of the range, in canonical form. Its IP family must
	// match the block (ipv4 or ipv6) the range is listed in.
	//
	// +required
	// +kubebuilder:validation:MinLength=2
	// +kubebuilder:validation:MaxLength=39
	// +kubebuilder:validation:XValidation:rule="isIP(self) && ip.isCanonical(self)",message="startingAddress must be a valid IP address in canonical form"
	StartingAddress string `json:"startingAddress,omitempty"`

	// addressCount is the number of addresses in the range.
	//
	// +required
	// +kubebuilder:validation:Minimum=1
	AddressCount int64 `json:"addressCount,omitempty"`
}

// VSphereDistributedNetworkIPv4Config is the IPv4 configuration of a network.
//
// +kubebuilder:validation:XValidation:rule="self.assignmentMode == 'staticpool' ? (has(self.gateway) && has(self.prefixLength)) : true",message="gateway and prefixLength are required when assignmentMode is staticpool"
// +kubebuilder:validation:XValidation:rule="self.assignmentMode != 'staticpool' ? (!has(self.gateway) && !has(self.prefixLength) && !has(self.addressRanges)) : true",message="gateway, prefixLength and addressRanges must be unset when assignmentMode is dhcp or none"
// +kubebuilder:validation:XValidation:rule="!has(self.addressRanges) || self.addressRanges.all(r, !isIP(r.startingAddress) || ip(r.startingAddress).family() == 4)",message="addressRanges must contain IPv4 addresses"
type VSphereDistributedNetworkIPv4Config struct {
	// assignmentMode is how IPv4 addresses are assigned on the network. It can't be changed.
	//
	// +required
	// +kubebuilder:validation:XValidation:rule="self == oldSelf",message="assignmentMode is immutable"
	AssignmentMode IPAssignmentModeType `json:"assignmentMode,omitempty"`

	// gateway is the IPv4 gateway for network interfaces. Set it only when assignmentMode is
	// staticpool.
	//
	// +optional
	// +kubebuilder:validation:MinLength=7
	// +kubebuilder:validation:MaxLength=15
	// +kubebuilder:validation:XValidation:rule="isIP(self) && ip(self).family() == 4 && ip.isCanonical(self)",message="gateway must be a valid IPv4 address in canonical form"
	Gateway string `json:"gateway,omitempty"`

	// prefixLength is the length of the IPv4 subnet prefix for network interfaces, for example 24
	// for a /24 subnet. Set it only when assignmentMode is staticpool.
	//
	// +optional
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:validation:Maximum=32
	PrefixLength *int32 `json:"prefixLength,omitempty"`

	// addressRanges are the IPv4 ranges addresses are assigned from when assignmentMode is
	// staticpool. The operator creates one IPPool per range, reports them in status.ipPools, and
	// deletes the IPPool for a range when the range is removed.
	//
	// +optional
	// +listType=atomic
	// +kubebuilder:validation:MaxItems=1024
	AddressRanges []VSphereDistributedNetworkIPRange `json:"addressRanges,omitempty"`
}

// VSphereDistributedNetworkIPv6Config is the IPv6 configuration of a network.
//
// +kubebuilder:validation:XValidation:rule="self.assignmentMode == 'staticpool' ? (has(self.gateway) && has(self.prefixLength)) : true",message="gateway and prefixLength are required when assignmentMode is staticpool"
// +kubebuilder:validation:XValidation:rule="self.assignmentMode != 'staticpool' ? (!has(self.gateway) && !has(self.prefixLength) && !has(self.addressRanges)) : true",message="gateway, prefixLength and addressRanges must be unset when assignmentMode is dhcp or none"
// +kubebuilder:validation:XValidation:rule="!has(self.addressRanges) || self.addressRanges.all(r, !isIP(r.startingAddress) || ip(r.startingAddress).family() == 6)",message="addressRanges must contain IPv6 addresses"
type VSphereDistributedNetworkIPv6Config struct {
	// assignmentMode is how IPv6 addresses are assigned on the network.
	//
	// +required
	AssignmentMode IPAssignmentModeType `json:"assignmentMode,omitempty"`

	// gateway is the IPv6 gateway for network interfaces. Set it only when assignmentMode is
	// staticpool.
	//
	// +optional
	// +kubebuilder:validation:MinLength=2
	// +kubebuilder:validation:MaxLength=39
	// +kubebuilder:validation:XValidation:rule="isIP(self) && ip(self).family() == 6 && ip.isCanonical(self)",message="gateway must be a valid IPv6 address in canonical form"
	Gateway string `json:"gateway,omitempty"`

	// prefixLength is the length of the IPv6 subnet prefix for network interfaces, for example 64
	// for a /64 subnet. Set it only when assignmentMode is staticpool.
	//
	// +optional
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:validation:Maximum=128
	PrefixLength *int32 `json:"prefixLength,omitempty"`

	// addressRanges are the IPv6 ranges addresses are assigned from when assignmentMode is
	// staticpool. The operator creates one IPPool per range, reports them in status.ipPools, and
	// deletes the IPPool for a range when the range is removed.
	//
	// +optional
	// +listType=atomic
	// +kubebuilder:validation:MaxItems=1024
	AddressRanges []VSphereDistributedNetworkIPRange `json:"addressRanges,omitempty"`
}

// VSphereDistributedNetworkSpec defines the desired state of VSphereDistributedNetwork.
//
// +kubebuilder:validation:XValidation:rule="!has(self.portGroupIDs) || size(self.portGroupIDs) == 0 || self.portGroupIDs[0] == self.portGroupID",message="portGroupIDs[0] must equal portGroupID when portGroupIDs is set"
type VSphereDistributedNetworkSpec struct {
	// portGroupID is an existing vSphere Distributed PortGroup identifier.
	//
	// +required
	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:validation:MaxLength=64
	PortGroupID string `json:"portGroupID,omitempty"`

	// portGroupIDs is the ordered list of vSphere portgroup identifiers that back this network.
	// All portgroups must be on the same subnet. Order is significant: portGroupIDs[0] is the
	// primary portgroup and must equal portGroupID. When unset or empty, the network is backed
	// by the single portgroup named by portGroupID. Use EffectivePortGroupIDs to read the
	// resolved list.
	//
	// +optional
	// +listType=atomic
	// +kubebuilder:validation:MaxItems=32
	// +kubebuilder:validation:items:MaxLength=64
	// +kubebuilder:validation:XValidation:rule="self.all(x, self.exists_one(y, y == x))",message="portGroupIDs must not contain duplicate entries"
	PortGroupIDs []string `json:"portGroupIDs,omitempty"`

	// ipv4 is the IPv4 configuration of the network. Set assignmentMode to none for a network
	// without IPv4.
	//
	// +required
	IPv4 VSphereDistributedNetworkIPv4Config `json:"ipv4,omitzero"`

	// ipv6 is the IPv6 configuration of the network. When unset, the network has no IPv6.
	//
	// +optional
	IPv6 VSphereDistributedNetworkIPv6Config `json:"ipv6,omitzero"`
}

// EffectivePortGroupIDs returns the portgroups that back the network, in order. It returns a copy
// of PortGroupIDs when it is non-empty, else a single-element list holding PortGroupID. The caller
// may modify the returned slice.
func (s *VSphereDistributedNetworkSpec) EffectivePortGroupIDs() []string {
	if len(s.PortGroupIDs) > 0 {
		return append([]string(nil), s.PortGroupIDs...)
	}
	return []string{s.PortGroupID}
}

// IPPoolReference is a reference to a netoperator.vmware.com IPPool.
type IPPoolReference struct {
	// name is the name of the IPPool.
	//
	// +required
	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:validation:MaxLength=253
	Name string `json:"name,omitempty"`

	// apiVersion is the API version of the IPPool.
	//
	// +optional
	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:validation:MaxLength=64
	APIVersion string `json:"apiVersion,omitempty"`
}

// VLANType represents the type of VLAN configuration
type VLANType string

const (
	// VLANTypeStandard represents a standard VLAN configuration with a single VLAN ID
	VLANTypeStandard VLANType = "standard"
	// VLANTypeTrunk represents a VLAN trunk configuration that allows multiple VLANs
	VLANTypeTrunk VLANType = "trunk"
	// VLANTypePrivate represents a private VLAN configuration
	VLANTypePrivate VLANType = "private"
)

// VLANTrunkRange represents a range of VLAN IDs for trunk configuration
//
//nolint:kubeapilinter // status-only; identical to v1alpha1 so conversion is a straight copy
type VLANTrunkRange struct {
	// start represents the beginning of the VLAN ID range (inclusive).
	//
	// +required
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:validation:Maximum=4094
	Start int32 `json:"start"`

	// end represents the end of the VLAN ID range (inclusive).
	//
	// +required
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:validation:Maximum=4094
	End int32 `json:"end"`
}

// VlanSpec represents the VLAN configuration.
//
//nolint:kubeapilinter // status-only; identical to v1alpha1 so conversion is a straight copy
type VlanSpec struct {
	// type indicates the type of VLAN configuration (standard, trunk, or private).
	//
	// +required
	// +kubebuilder:validation:Enum=standard;trunk;private
	Type VLANType `json:"type"`

	// vlanID specifies the VLAN ID when Type is VLANTypeStandard.
	// This field is ignored for other VLAN types.
	// Possible values:
	// - A value of 0 indicates there is no VLAN configuration for the port.
	// - A value from 1 to 4094 specifies a VLAN ID for the port.
	//
	// +optional
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:validation:Maximum=4094
	VlanID *int32 `json:"vlanID,omitempty"`

	// trunkRange specifies the ranges of allowed VLANs when Type is VLANTypeTrunk.
	// This field is ignored for other VLAN types.
	// Each range's Start and End values must be between 0 and 4094 inclusive.
	// Overlapping ranges are allowed.
	//
	// +optional
	// +listType=atomic
	// +kubebuilder:validation:MaxItems=4095
	TrunkRange []VLANTrunkRange `json:"trunkRange,omitempty"`

	// privateVlanID specifies the private VLAN ID when Type is VLANTypePrivate.
	// This field is ignored for other VLAN types.
	//
	// +optional
	PrivateVlanID *int32 `json:"privateVlanID,omitempty"`
}

// MacLimitPolicyType represents the policy type to be used when the MAC address limit is exceeded.
type MacLimitPolicyType string

const (
	// MacLimitPolicyAllow indicates that new MAC addresses should still be allowed when the limit is exceeded.
	MacLimitPolicyAllow MacLimitPolicyType = "allow"
	// MacLimitPolicyDrop indicates that new MAC addresses should be dropped when the limit is exceeded.
	MacLimitPolicyDrop MacLimitPolicyType = "drop"
)

// MacLearningPolicy represents the MAC learning policy configuration. It is reported in status
// only and matches v1alpha1 field for field, so conversion is a straight copy.
//
//nolint:kubeapilinter // status-only; identical to v1alpha1 so conversion is a straight copy
type MacLearningPolicy struct {
	// enabled indicates whether MAC learning is enabled.
	//
	// +required
	Enabled bool `json:"enabled"`

	// allowUnicastFlooding indicates whether to allow flooding of unlearned MAC for ingress traffic.
	//
	// +optional
	AllowUnicastFlooding *bool `json:"allowUnicastFlooding,omitempty"`

	// limit represents the maximum number of MAC addresses that can be learned.
	//
	// +optional
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:validation:Maximum=4096
	Limit *int32 `json:"limit,omitempty"`

	// limitPolicy represents the policy to be used when the limit is exceeded.
	//
	// +optional
	// +kubebuilder:validation:Enum=allow;drop
	LimitPolicy *MacLimitPolicyType `json:"limitPolicy,omitempty"`
}

// MacManagementPolicy represents the MAC management policy configuration. It is reported in
// status only and matches v1alpha1 field for field, so conversion is a straight copy.
//
//nolint:kubeapilinter // status-only; identical to v1alpha1 so conversion is a straight copy
type MacManagementPolicy struct {
	// allowPromiscuous indicates whether promiscuous mode is enabled. Determines whether or not all
	// traffic is seen on the port.
	//
	// +optional
	AllowPromiscuous *bool `json:"allowPromiscuous,omitempty"`

	// macChanges specifies whether virtual machines can receive frames with a Mac Address that is different from the one configured in the VMX.
	//
	// +optional
	MacChanges *bool `json:"macChanges,omitempty"`

	// forgedTransmits indicates whether or not the virtual network adapter should be allowed to send
	// network traffic with a different MAC address than the one assigned to it.
	//
	// +optional
	ForgedTransmits *bool `json:"forgedTransmits,omitempty"`

	// macLearningPolicy represents the MAC learning policy configuration.
	//
	// +optional
	MacLearningPolicy *MacLearningPolicy `json:"macLearningPolicy,omitempty"`
}

// VSphereDistributedPortConfig represents the port-level configuration for a vSphere Distributed Network's ports.
//
//nolint:kubeapilinter // status-only; identical to v1alpha1 so conversion is a straight copy
type VSphereDistributedPortConfig struct {
	// vlan represents the VLAN configuration.
	//
	// +optional
	Vlan *VlanSpec `json:"vlan,omitempty"`

	// macManagementPolicy represents the MAC management policy configuration.
	//
	// +optional
	MacManagementPolicy *MacManagementPolicy `json:"macManagementPolicy,omitempty"`
}

// VSphereDistributedNetworkStatus defines the observed state of VSphereDistributedNetwork.
type VSphereDistributedNetworkStatus struct {
	// conditions describe the current state of the network.
	//
	// +optional
	// +listType=map
	// +listMapKey=type
	// +kubebuilder:validation:MaxItems=32
	Conditions []metav1.Condition `json:"conditions,omitempty"`

	// ipPools are the IPPools that addresses on this network are allocated from. When the network
	// has address ranges, the operator sets it to the IPPools it created for them, in range order
	// (IPv4 ranges first). Networks created through v1alpha1 without address ranges carry the
	// IPPools named in v1alpha1 spec.ipPools here.
	//
	// +optional
	// +listType=atomic
	// +kubebuilder:validation:MaxItems=1024
	IPPools []IPPoolReference `json:"ipPools,omitempty"`

	// defaultPortConfig represents the default port-level configuration that applies to all ports
	// unless overridden at the individual port level.
	// If unset, indicates that no default port-level configuration has been retrieved yet for this network.
	//
	// +optional
	DefaultPortConfig *VSphereDistributedPortConfig `json:"defaultPortConfig,omitempty"`
}

// VSphereDistributedNetwork represents schema for a network backed by a vSphere Distributed PortGroup on vSphere
// Distributed switch.
//
// The status subresource is deliberately not enabled: v1alpha1 clients set status.ipPools through
// v1alpha1 spec.ipPools with full-object updates.
//
// +genclient
// +genclient:nonNamespaced
// +kubebuilder:object:root=true
// +kubebuilder:resource:scope=Cluster
// +kubebuilder:storageversion
// +kubebuilder:validation:XValidation:rule="size(self.metadata.name) <= 253 && self.metadata.name.matches('^[a-z0-9]([-a-z0-9]*[a-z0-9])?(\\\\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*$')",message="metadata.name must be a lowercase RFC 1123 DNS subdomain (alphanumeric or '-' or '.', each segment starting/ending with alphanumeric; max 253 characters)"
type VSphereDistributedNetwork struct { //nolint:kubeapilinter // no status subresource; see the type comment
	metav1.TypeMeta `json:",inline"`
	// metadata is the standard object metadata.
	//
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// spec is the desired state of the network.
	//
	// +required
	Spec VSphereDistributedNetworkSpec `json:"spec,omitzero"`

	// status is the observed state of the network.
	//
	// +optional
	Status *VSphereDistributedNetworkStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// VSphereDistributedNetworkList contains a list of VSphereDistributedNetwork
type VSphereDistributedNetworkList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []VSphereDistributedNetwork `json:"items"`
}

func init() {
	RegisterTypeWithScheme(&VSphereDistributedNetwork{}, &VSphereDistributedNetworkList{})
}
