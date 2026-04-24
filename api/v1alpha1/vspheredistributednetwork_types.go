// Copyright (c) 2020-2026 Broadcom. All Rights Reserved.
// Broadcom Confidential. The term "Broadcom" refers to Broadcom Inc.
// and/or its subsidiaries.

package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type VSphereDistributedNetworkConditionType string

const (
	// VSphereDistributedNetworkPortGroupFailure is added when PortGroupID specified either doesn't exist, or
	// there was an error in communicating with vCenter Server.
	VSphereDistributedNetworkPortGroupFailure VSphereDistributedNetworkConditionType = "PortGroupFailure"

	// VSphereDistributedNetworkIPPoolInvalid is added when no valid IPPool references exists.
	VSphereDistributedNetworkIPPoolInvalid VSphereDistributedNetworkConditionType = "IPPoolInvalid"

	// VsphereDistributedNetworkIPPoolPressure condition status is set to True when IPPool is low on free IPs.
	VsphereDistributedNetworkIPPoolPressure VSphereDistributedNetworkConditionType = "IPPoolPressure"
)

type IPAssignmentModeType string

const (
	// IPAssignmentModeDHCP indicates IP address is assigned dynamically using DHCP.
	IPAssignmentModeDHCP IPAssignmentModeType = "dhcp"

	// IPAssignmentModeStaticPool indicates IP address is assigned from a static pool of IP addresses.
	IPAssignmentModeStaticPool IPAssignmentModeType = "staticpool"

	// IPAssignmentModeNone indicates that no IP assignment will be performed.
	// The operator will not assign an IP and no DHCP client will be configured.
	IPAssignmentModeNone IPAssignmentModeType = "none"
)

// VSphereDistributedNetworkIPRange is the static IP range for a VSphereDistributedNetwork.
type VSphereDistributedNetworkIPRange struct {
	// address is the starting IPv4 address of the range.
	// +kubebuilder:validation:Format=ipv4
	// +kubebuilder:validation:MinLength=7
	// +kubebuilder:validation:MaxLength=15
	// +required
	Address string `json:"address,omitempty"`

	// count is the number of addresses in the range when using static range assignment.
	// +kubebuilder:validation:Minimum=1
	// +required
	Count int64 `json:"count,omitempty"`
}

// VSphereDistributedNetworkCondition describes the state of a VSphereDistributedNetwork at a certain point.
type VSphereDistributedNetworkCondition struct {
	// Type is the type of VSphereDistributedNetwork condition.
	//
	//nolint:kubeapilinter // Stable v1alpha1 retention: avoid MaxLength (would tighten validation). Keep condition type without omitempty (requiredfields wire shape).
	Type VSphereDistributedNetworkConditionType `json:"type"`

	// Status is the status of the condition.
	// Can be True, False, Unknown.
	//
	//nolint:kubeapilinter // Stable v1alpha1 retention: keep status without omitempty (requiredfields wire shape).
	Status corev1.ConditionStatus `json:"status"`

	// Reason is a machine understandable string that gives the reason for condition's last transition.
	//
	//nolint:kubeapilinter // Stable v1alpha1 retention: avoid MaxLength (would tighten validation). Avoid pointer (optionalfields).
	Reason string `json:"reason,omitempty"`

	// Message is a human-readable message indicating details about last transition.
	//
	//nolint:kubeapilinter // Stable v1alpha1 retention: avoid MaxLength (would tighten validation). Avoid pointer (optionalfields).
	Message string `json:"message,omitempty"`

	// lastTransitionTime provides a timestamp for when the VSphereDistributedNetwork object last transitioned from one status to another.
	// +optional
	LastTransitionTime metav1.Time `json:"lastTransitionTime,omitempty" patchStrategy:"replace"`
}

// +kubebuilder:validation:XValidation:rule="(has(self.ipAssignmentMode) && (self.ipAssignmentMode == 'dhcp' || self.ipAssignmentMode == 'none')) ? (!has(self.gateway) || self.gateway == ”) : true",message="Gateway must be empty when IpAssignmentMode is dhcp or none"
// +kubebuilder:validation:XValidation:rule="(has(self.ipAssignmentMode) && (self.ipAssignmentMode == 'dhcp' || self.ipAssignmentMode == 'none')) ? (!has(self.subnetMask) || self.subnetMask == ”) : true",message="SubnetMask must be empty when IpAssignmentMode is dhcp or none"
// +kubebuilder:validation:XValidation:rule="(has(self.ipAssignmentMode) && (self.ipAssignmentMode == 'dhcp' || self.ipAssignmentMode == 'none')) ? (!has(self.addressRanges) || size(self.addressRanges) == 0) : true",message="AddressRanges must be empty when IpAssignmentMode is dhcp or none"
// +kubebuilder:validation:XValidation:rule="(has(self.ipAssignmentMode) && (self.ipAssignmentMode == 'dhcp' || self.ipAssignmentMode == 'none')) ? (!has(self.ipPools) || size(self.ipPools) == 0) : true",message="IPPools must be empty when IpAssignmentMode is dhcp or none"
// +kubebuilder:validation:XValidation:rule="(!has(self.ipAssignmentMode) || self.ipAssignmentMode == 'staticpool') ? (has(self.gateway) && self.gateway != ”) : true",message="Gateway is required when IpAssignmentMode is staticpool"
// +kubebuilder:validation:XValidation:rule="(!has(self.ipAssignmentMode) || self.ipAssignmentMode == 'staticpool') ? (has(self.subnetMask) && self.subnetMask != ”) : true",message="SubnetMask is required when IpAssignmentMode is staticpool"
// VSphereDistributedNetworkSpec defines the desired state of VSphereDistributedNetwork.
type VSphereDistributedNetworkSpec struct {
	// PortGroupID is an existing vSphere Distributed PortGroup identifier.
	//
	//nolint:kubeapilinter // Stable v1alpha1 retention: avoid MaxLength (would tighten validation). Avoid omitempty (requiredfields wire shape).
	PortGroupID string `json:"portGroupID"`

	// IPAssignmentMode selects IPv4 assignment for network interfaces. If unset, defaults to IPAssignmentModeStaticPool.
	// For IPAssignmentModeDHCP and IPAssignmentModeNone, the IPv4 IPPools, Gateway and SubnetMask
	// fields should be empty/unset. When using IPAssignmentModeNone, no IPv4 IP will be assigned
	// and no DHCP client will be configured.
	// Note: For IPv6 address assignment, see IPv6AssignmentMode.
	// +kubebuilder:validation:Enum=dhcp;staticpool;none
	// +kubebuilder:default:=staticpool
	// +kubebuilder:validation:XValidation:rule="self == oldSelf",message="ipAssignmentMode is immutable"
	// +optional
	//
	//nolint:kubeapilinter // Stable v1alpha1 retention: avoid MaxLength (would tighten validation). Avoid pointer (optionalfields).
	IPAssignmentMode IPAssignmentModeType `json:"ipAssignmentMode,omitempty"`

	// IPv6AssignmentMode selects IPv6 assignment for network interfaces. If unset, defaults to
	// IPAssignmentModeNone (IPv6 disabled) for backward compatibility with existing IPv4-only
	// deployments. To enable IPv6 support, explicitly set this field to IPAssignmentModeStaticPool
	// or IPAssignmentModeDHCP. This allows different assignment modes for IPv4 and IPv6, for
	// example static IPv4 pool assignment combined with DHCPv6 for IPv6.
	// +optional
	//
	//nolint:kubeapilinter // Stable v1alpha1 retention: avoid MaxLength (would tighten validation). Avoid pointer (optionalfields).
	IPv6AssignmentMode IPAssignmentModeType `json:"ipv6AssignmentMode,omitempty"`

	// IPPools references list of IPPool objects. This field should only be set when using
	// IPAssignmentModeStaticPool. For all other modes (IPAssignmentModeDHCP, IPAssignmentModeNone), this should be set
	// 	to an empty list.
	//
	//nolint:kubeapilinter // Stable v1alpha1 retention: avoid MaxItems (would tighten validation). Avoid omitempty (requiredfields wire shape).
	IPPools []IPPoolReference `json:"ipPools"`

	// Gateway is the gateway to use for network interfaces. This field should only be set when using
	// IPAssignmentModeStaticPool. For all other modes (IPAssignmentModeDHCP, IPAssignmentModeNone), this should be set
	// 	to an empty string.
	// Note: The regex pattern performs IPv4 validation but also allows an empty string for backward compatibility.
	// +kubebuilder:validation:Pattern="^(|((25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\\.){3}(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?))$"
	// +optional
	//
	//nolint:kubeapilinter // Stable v1alpha1 retention: avoid MaxLength (would tighten validation). Avoid omitempty (requiredfields wire shape).
	Gateway string `json:"gateway,omitempty"`

	// SubnetMask is the subnet mask to use for network interfaces. This field should only be set when using
	// IPAssignmentModeStaticPool. For all other modes (IPAssignmentModeDHCP, IPAssignmentModeNone), this should be set
	// 	to an empty string.
	// Note: The regex pattern performs IPv4 validation but also allows an empty string for backward compatibility.
	// +kubebuilder:validation:Pattern="^(|((25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\\.){3}(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?))$"
	// +optional
	//
	//nolint:kubeapilinter // Stable v1alpha1 retention: avoid MaxLength (would tighten validation). Avoid omitempty (requiredfields wire shape).
	SubnetMask string `json:"subnetMask,omitempty"`

	// IPv6Gateway is the IPv6 gateway to use for network interfaces. This field should only
	// be set when using IPv6AssignmentMode IPAssignmentModeStaticPool. For all other modes
	// (IPAssignmentModeDHCP, IPAssignmentModeNone), this should be empty/unset.
	// +optional
	//
	//nolint:kubeapilinter // Stable v1alpha1 retention: avoid MaxLength (would tighten validation). Avoid pointer (optionalfields).
	IPv6Gateway string `json:"ipv6Gateway,omitempty"`

	// ipv6Prefix is the prefix length for IPv6 addresses assigned to network interfaces (e.g. 64
	// for a /64 network). This field should only be set when using IPv6AssignmentMode
	// IPAssignmentModeStaticPool. For all other modes (IPAssignmentModeDHCP, IPAssignmentModeNone),
	// this should be unset.
	// +optional
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:validation:Maximum=128
	IPv6Prefix *int32 `json:"ipv6Prefix,omitempty"`

	// addressRanges is a list of IP ranges for static IP assignment.
	// +optional
	// +kubebuilder:validation:MaxItems=32
	// +listType=atomic
	AddressRanges []VSphereDistributedNetworkIPRange `json:"addressRanges,omitempty"`
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
type VLANTrunkRange struct {
	// Start represents the beginning of the VLAN ID range (inclusive).
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:validation:Maximum=4094
	//
	//nolint:kubeapilinter // Stable v1alpha1 retention: keep int32 range endpoints without pointers (requiredfields).
	Start int32 `json:"start"`

	// End represents the end of the VLAN ID range (inclusive).
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:validation:Maximum=4094
	//
	//nolint:kubeapilinter // Stable v1alpha1 retention: keep int32 range endpoints without pointers (requiredfields).
	End int32 `json:"end"`
}

// VlanSpec represents the VLAN configuration.
type VlanSpec struct {
	// Type indicates the type of VLAN configuration (standard, trunk, or private).
	// +kubebuilder:validation:Enum=standard;trunk;private
	//
	//nolint:kubeapilinter // Stable v1alpha1 retention: keep type without omitempty (requiredfields wire shape).
	Type VLANType `json:"type"`

	// vlanID specifies the VLAN ID when Type is VLANTypeStandard.
	// This field is ignored for other VLAN types.
	// Possible values:
	// - A value of 0 indicates there is no VLAN configuration for the port.
	// - A value from 1 to 4094 specifies a VLAN ID for the port.
	// +optional
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:validation:Maximum=4094
	VlanID *int32 `json:"vlanID,omitempty"`

	// TrunkRange specifies the ranges of allowed VLANs when Type is VLANTypeTrunk.
	// This field is ignored for other VLAN types.
	// Each range's Start and End values must be between 0 and 4094 inclusive.
	// Overlapping ranges are allowed.
	// +optional
	//
	//nolint:kubeapilinter // Stable v1alpha1 retention: avoid MaxItems (would tighten validation).
	TrunkRange []VLANTrunkRange `json:"trunkRange,omitempty"`

	// privateVlanID specifies the private VLAN ID when Type is VLANTypePrivate.
	// This field is ignored for other VLAN types.
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

// MacLearningPolicy represents the MAC learning policy configuration.
type MacLearningPolicy struct {
	// Enabled indicates whether MAC learning is enabled.
	//
	//nolint:kubeapilinter // Stable v1alpha1 retention: keep bool flags (nobools); enum would be wire-incompatible.
	Enabled bool `json:"enabled"`

	// AllowUnicastFlooding indicates whether to allow flooding of unlearned MAC for ingress traffic.
	// +optional
	//
	//nolint:kubeapilinter // Stable v1alpha1 retention: keep optional bool pointer (nobools).
	AllowUnicastFlooding *bool `json:"allowUnicastFlooding,omitempty"`

	// limit represents the maximum number of MAC addresses that can be learned.
	// +optional
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:validation:Maximum=4096
	Limit *int32 `json:"limit,omitempty"`

	// LimitPolicy represents the policy to be used when the limit is exceeded.
	// +optional
	// +kubebuilder:validation:Enum=allow;drop
	//
	//nolint:kubeapilinter // Stable v1alpha1 retention: keep optional pointer for unset vs zero (optionalfields).
	LimitPolicy *MacLimitPolicyType `json:"limitPolicy,omitempty"`
}

// MacManagementPolicy represents the MAC management policy configuration.
type MacManagementPolicy struct {
	// AllowPromiscuous indicates whether promiscuous mode is enabled. Determines whether or not all
	// traffic is seen on the port.
	// +optional
	//
	//nolint:kubeapilinter // Stable v1alpha1 retention: keep optional bool pointer (nobools).
	AllowPromiscuous *bool `json:"allowPromiscuous,omitempty"`

	// MacChanges specifies whether virtual machines can receive frames with a Mac Address that is different from the one configured in the VMX.
	// +optional
	//
	//nolint:kubeapilinter // Stable v1alpha1 retention: keep optional bool pointer (nobools).
	MacChanges *bool `json:"macChanges,omitempty"`

	// ForgedTransmits indicates whether or not the virtual network adapter should be allowed to send
	// network traffic with a different MAC address than the one assigned to it.
	// +optional
	//
	//nolint:kubeapilinter // Stable v1alpha1 retention: keep optional bool pointer (nobools).
	ForgedTransmits *bool `json:"forgedTransmits,omitempty"`

	// macLearningPolicy represents the MAC learning policy configuration.
	// +optional
	MacLearningPolicy *MacLearningPolicy `json:"macLearningPolicy,omitempty"`
}

// VSphereDistributedPortConfig represents the port-level configuration for a vSphere Distributed Network's ports.
type VSphereDistributedPortConfig struct {
	// Vlan represents the VLAN configuration.
	// +optional
	//
	//nolint:kubeapilinter // Stable v1alpha1 retention: keep json tag without omitzero (optionalfields wire shape).
	Vlan *VlanSpec `json:"vlan,omitempty"`

	// macManagementPolicy represents the MAC management policy configuration.
	// +optional
	MacManagementPolicy *MacManagementPolicy `json:"macManagementPolicy,omitempty"`
}

// VSphereDistributedNetworkStatus defines the observed state of VSphereDistributedNetwork.
type VSphereDistributedNetworkStatus struct {
	// Conditions are an array of current observed vSphere Distributed network conditions.
	//
	//nolint:kubeapilinter // Stable v1alpha1 retention: keep custom VSphereDistributedNetworkCondition slice (conditions).
	Conditions []VSphereDistributedNetworkCondition `json:"conditions,omitempty"`

	// defaultPortConfig represents the default port-level configuration that applies to all ports
	// unless overridden at the individual port level.
	// If unset, indicates that no default port-level configuration has been retrieved yet for this network.
	// +optional
	DefaultPortConfig *VSphereDistributedPortConfig `json:"defaultPortConfig,omitempty"`
}

// +genclient
// +genclient:nonNamespaced
// +kubebuilder:object:root=true
// +kubebuilder:resource:scope=Cluster
// +kubebuilder:validation:XValidation:rule="size(self.metadata.name) <= 253 && self.metadata.name.matches('^[a-z0-9]([-a-z0-9]*[a-z0-9])?(\\\\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*$')",message="metadata.name must be a lowercase RFC 1123 DNS subdomain (alphanumeric or '-' or '.', each segment starting/ending with alphanumeric; max 253 characters)"

// VSphereDistributedNetwork represents schema for a network backed by a vSphere Distributed PortGroup on vSphere
// Distributed switch.
//
//nolint:kubeapilinter // Stable v1alpha1 retention: ignore kubebuilder:subresource:status marker.
type VSphereDistributedNetwork struct {
	metav1.TypeMeta `json:",inline"`

	// metadata is standard Kubernetes object metadata.
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// Spec defines the desired state of the VSphereDistributedNetwork.
	//
	//nolint:kubeapilinter // Stable v1alpha1 retention: keep spec value type (optionalfields pointer churn).
	Spec VSphereDistributedNetworkSpec `json:"spec,omitempty"`

	// Status defines the observed state of the VSphereDistributedNetwork.
	//
	//nolint:kubeapilinter // Stable v1alpha1 retention: keep nested status without omitzero (requiredfields).
	Status VSphereDistributedNetworkStatus `json:"status,omitempty"`
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
