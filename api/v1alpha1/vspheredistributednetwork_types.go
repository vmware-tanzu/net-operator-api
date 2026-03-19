// Copyright (c) 2020-2025 Broadcom. All Rights Reserved.
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

// VSphereDistributedNetworkCondition describes the state of a VSphereDistributedNetwork at a certain point.
type VSphereDistributedNetworkCondition struct {
	// Type is the type of VSphereDistributedNetwork condition.
	Type VSphereDistributedNetworkConditionType `json:"type"`
	// Status is the status of the condition.
	// Can be True, False, Unknown.
	Status corev1.ConditionStatus `json:"status"`
	// Machine understandable string that gives the reason for condition's last transition.
	Reason string `json:"reason,omitempty"`
	// Human-readable message indicating details about last transition.
	Message string `json:"message,omitempty"`
	// Provides a timestamp for when the VSphereDistributedNetwork object last transitioned from one status to another.
	LastTransitionTime metav1.Time `json:"lastTransitionTime,omitempty" patchStrategy:"replace"`
}

// VSphereDistributedNetworkSpec defines the desired state of VSphereDistributedNetwork.
type VSphereDistributedNetworkSpec struct {
	// PortGroupID is an existing vSphere Distributed PortGroup identifier.
	PortGroupID string `json:"portGroupID"`

	// IPAssignmentMode to use for IPv4 addresses on network interfaces. If unset, defaults to IPAssignmentModeStaticPool.
	// For IPAssignmentModeDHCP and IPAssignmentModeNone, the IPPools, Gateway and SubnetMask
	// fields should be empty/unset. When using IPAssignmentModeNone, no IP will be assigned
	// and no DHCP client will be configured.
	// Note: For IPv6 address assignment, see IPv6AssignmentMode.
	// +optional
	IPAssignmentMode IPAssignmentModeType `json:"ipAssignmentMode,omitempty"`

	// IPv6AssignmentMode to use for IPv6 addresses on network interfaces. If unset, defaults to
	// IPAssignmentModeNone (IPv6 disabled) for backward compatibility with existing IPv4-only
	// deployments. To enable IPv6 support, explicitly set this field to IPAssignmentModeStaticPool
	// or IPAssignmentModeDHCP. This allows different assignment modes for IPv4 and IPv6, for
	// example static IPv4 pool assignment combined with DHCPv6 for IPv6.
	// +optional
	IPv6AssignmentMode IPAssignmentModeType `json:"ipv6AssignmentMode,omitempty"`

	// IPPools references list of IPPool objects. This field should only be set when using
	// IPAssignmentModeStaticPool. For all other modes (IPAssignmentModeDHCP, IPAssignmentModeNone), this should be set
	// to an empty list.
	IPPools []IPPoolReference `json:"ipPools"`

	// Gateway setting to use for network interfaces. This field should only be set when using
	// IPAssignmentModeStaticPool. For all other modes (IPAssignmentModeDHCP, IPAssignmentModeNone), this should be set
	// to an empty string.
	Gateway string `json:"gateway"`

	// SubnetMask setting to use for network interfaces. This field should only be set when using
	// IPAssignmentModeStaticPool. For all other modes (IPAssignmentModeDHCP, IPAssignmentModeNone), this should be set
	// to an empty string.
	SubnetMask string `json:"subnetMask"`

	// IPv6Gateway setting to use for IPv6 addresses on network interfaces. This field should only
	// be set when using IPv6AssignmentMode IPAssignmentModeStaticPool. For all other modes
	// (IPAssignmentModeDHCP, IPAssignmentModeNone), this should be empty/unset.
	// +optional
	IPv6Gateway string `json:"ipv6Gateway,omitempty"`

	// IPv6Prefix is the prefix length for IPv6 addresses assigned to network interfaces (e.g. 64
	// for a /64 network). This field should only be set when using IPv6AssignmentMode
	// IPAssignmentModeStaticPool. For all other modes (IPAssignmentModeDHCP, IPAssignmentModeNone),
	// this should be unset.
	// +optional
	IPv6Prefix *int32 `json:"ipv6Prefix,omitempty"`
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
	Start int32 `json:"start"`

	// End represents the end of the VLAN ID range (inclusive).
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:validation:Maximum=4094
	End int32 `json:"end"`
}

// VlanSpec represents the VLAN configuration.
type VlanSpec struct {
	// Type indicates the type of VLAN configuration (standard, trunk, or private).
	// +kubebuilder:validation:Enum=standard;trunk;private
	Type VLANType `json:"type"`

	// VlanID specifies the VLAN ID when Type is VLANTypeStandard.
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
	TrunkRange []VLANTrunkRange `json:"trunkRange,omitempty"`

	// PrivateVlanID specifies the private VLAN ID when Type is VLANTypePrivate.
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
	Enabled bool `json:"enabled"`

	// AllowUnicastFlooding indicates whether to allow flooding of unlearned MAC for ingress traffic.
	// +optional
	AllowUnicastFlooding *bool `json:"allowUnicastFlooding,omitempty"`

	// Limit represents the maximum number of MAC addresses that can be learned.
	// +optional
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:validation:Maximum=4096
	Limit *int32 `json:"limit,omitempty"`

	// LimitPolicy represents the policy to be used when the limit is exceeded.
	// +optional
	// +kubebuilder:validation:Enum=allow;drop
	LimitPolicy *MacLimitPolicyType `json:"limitPolicy,omitempty"`
}

// MacManagementPolicy represents the MAC management policy configuration.
type MacManagementPolicy struct {
	// AllowPromiscuous indicates whether promiscuous mode is enabled. Determines whether or not all
	// traffic is seen on the port.
	// +optional
	AllowPromiscuous *bool `json:"allowPromiscuous,omitempty"`

	// MacChanges specifies whether virtual machines can receive frames with a Mac Address that is different from the one configured in the VMX.
	// +optional
	MacChanges *bool `json:"macChanges,omitempty"`

	// ForgedTransmits indicates whether or not the virtual network adapter should be allowed to send
	// network traffic with a different MAC address than the one assigned to it.
	// +optional
	ForgedTransmits *bool `json:"forgedTransmits,omitempty"`

	// MacLearningPolicy represents the MAC learning policy configuration.
	// +optional
	MacLearningPolicy *MacLearningPolicy `json:"macLearningPolicy,omitempty"`
}

// VSphereDistributedPortConfig represents the port-level configuration for a vSphere Distributed Network's ports.
type VSphereDistributedPortConfig struct {
	// Vlan represents the VLAN configuration.
	// +optional
	Vlan *VlanSpec `json:"vlan,omitempty"`

	// MacManagementPolicy represents the MAC management policy configuration.
	// +optional
	MacManagementPolicy *MacManagementPolicy `json:"macManagementPolicy,omitempty"`
}

// VSphereDistributedNetworkStatus defines the observed state of VSphereDistributedNetwork.
type VSphereDistributedNetworkStatus struct {
	// Conditions is an array of current observed vSphere Distributed network conditions.
	Conditions []VSphereDistributedNetworkCondition `json:"conditions,omitempty"`

	// DefaultPortConfig represents the default port-level configuration that applies to all ports
	// unless overridden at the individual port level.
	// If unset, indicates that no default port-level configuration has been retrieved yet for this network.
	// +optional
	DefaultPortConfig *VSphereDistributedPortConfig `json:"defaultPortConfig,omitempty"`
}

// +genclient
// +genclient:nonNamespaced
// +kubebuilder:object:root=true
// +kubebuilder:resource:scope=Cluster

// VSphereDistributedNetwork represents schema for a network backed by a vSphere Distributed PortGroup on vSphere
// Distributed switch.
type VSphereDistributedNetwork struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   VSphereDistributedNetworkSpec   `json:"spec,omitempty"`
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
