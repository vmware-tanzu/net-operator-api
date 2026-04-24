// Copyright (c) 2020-2026 Broadcom. All Rights Reserved.
// Broadcom Confidential. The term "Broadcom" refers to Broadcom Inc.
// and/or its subsidiaries.

package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	// NetworkProtectionFinalizer allows the Controller to clean up resources and ensure
	// that no NetworkInterfaces are actively using this Network before deletion.
	NetworkProtectionFinalizer = "network.netoperator.vmware.com/network-protection"
)

type NetworkConditionType string

const (
	// NetworkDeletionBlocked indicates that the Network cannot be deleted, because
	// there may be some consumers (NetworkInterface) still actively using it.
	NetworkDeletionBlocked NetworkConditionType = "DeletionBlocked"
)

type NetworkConditionReason string

const (
	// NetworkDeletionBlockedReasonInUse indicates that the Network deletion is blocked
	// because there are NetworkInterfaces still actively using this Network.
	NetworkDeletionBlockedReasonInUse NetworkConditionReason = "NetworkInUse"
)

// NetworkProviderReference contains info to locate a network provider object.
type NetworkProviderReference struct {
	// APIGroup is the group for the resource being referenced.
	//
	//nolint:kubeapilinter // Stable v1alpha1 retention: avoid MaxLength (would tighten validation). Avoid omitempty (requiredfields wire shape).
	APIGroup string `json:"apiGroup"`

	// Kind is the type of resource being referenced.
	//
	//nolint:kubeapilinter // Stable v1alpha1 retention: avoid MaxLength (would tighten validation). Avoid omitempty (requiredfields wire shape).
	Kind string `json:"kind"`

	// Name is the name of resource being referenced.
	//
	//nolint:kubeapilinter // Stable v1alpha1 retention: avoid MaxLength (would tighten validation). Avoid omitempty (requiredfields wire shape).
	Name string `json:"name"`

	// Namespace is the namespace of the resource being referenced. If empty, cluster scoped resource is assumed.
	//
	//nolint:kubeapilinter // Stable v1alpha1 retention: keep optional string without pointer (optionalfields). Retain default (forbiddenmarkers).
	Namespace string `json:"namespace,omitempty"`

	// APIVersion is the API version of the referent.
	//
	//nolint:kubeapilinter // Stable v1alpha1 retention: keep optional string without pointer (optionalfields).
	APIVersion string `json:"apiVersion,omitempty"`
}

// NetworkType is used to type the constants describing possible network types.
type NetworkType string

const (
	// NetworkTypeNSXT is the network type describing NSX-T.
	NetworkTypeNSXT = NetworkType("nsx-t")

	// NetworkTypeVDS is the network type describing VSphere Distributed Switch.
	NetworkTypeVDS = NetworkType("vsphere-distributed")

	// NetworkTypeNSXTVPC is the network type describing NSX-T VPC.
	NetworkTypeNSXTVPC = NetworkType("nsx-t_vpc")
)

// NetworkSpec defines the state of Network.
type NetworkSpec struct {
	// Type describes type of Network. Supported values are nsx-t, vsphere-distributed.
	//
	//nolint:kubeapilinter // Stable v1alpha1 retention: avoid MaxLength (would tighten validation). Keep type without omitempty (requiredfields wire shape).
	Type NetworkType `json:"type"`

	// ProviderRef is reference to a network provider object that provides this type of network.
	//
	//nolint:kubeapilinter // Stable v1alpha1 retention: avoid omitempty (requiredfields wire shape).
	ProviderRef NetworkProviderReference `json:"providerRef"`

	// DNS is a list of DNS server IPs to associate with network interfaces on this network.
	//
	//nolint:kubeapilinter // Stable v1alpha1 retention: avoid MaxLength (would tighten validation). Keep slice without omitempty (optionalfields wire shape).
	DNS []string `json:"dns,omitempty"`

	// DNSSearchDomains is a list of DNS search domains to associate with network interfaces on this network.
	//
	//nolint:kubeapilinter // Stable v1alpha1 retention: avoid MaxLength (would tighten validation). Keep slice without omitempty (optionalfields wire shape).
	DNSSearchDomains []string `json:"dnsSearchDomains,omitempty"`

	// NTP is a list of NTP server DNS names or IP addresses to use on this network.
	//
	//nolint:kubeapilinter // Stable v1alpha1 retention: avoid MaxLength (would tighten validation). Keep slice without omitempty (optionalfields wire shape).
	NTP []string `json:"ntp,omitempty"`
}

// NetworkCondition describes the state of a Network at a certain point.
type NetworkCondition struct {
	// Type is the type of network condition.
	//
	//nolint:kubeapilinter // Stable v1alpha1 retention: avoid MaxLength (would tighten validation). Keep condition type without omitempty (requiredfields wire shape).
	Type NetworkConditionType `json:"type"`

	// Status is the status of the condition.
	// Can be True, False, Unknown.
	//
	//nolint:kubeapilinter // Stable v1alpha1 retention: keep status without omitempty (requiredfields wire shape).
	Status corev1.ConditionStatus `json:"status"`

	// LastTransitionTime is the timestamp corresponding to the last status
	// change of this condition.
	//
	//nolint:kubeapilinter // Stable v1alpha1 retention: avoid MaxLength (would tighten validation). Avoid pointer (optionalfields).
	LastTransitionTime metav1.Time `json:"lastTransitionTime,omitempty"`

	// Reason is a machine understandable string that gives the reason for condition's last transition.
	//
	//nolint:kubeapilinter // Stable v1alpha1 retention: avoid MaxLength (would tighten validation). Avoid pointer (optionalfields).
	Reason NetworkConditionReason `json:"reason,omitempty"`

	// Message is a human-readable message indicating details about last transition.
	//
	//nolint:kubeapilinter // Stable v1alpha1 retention: avoid MaxLength (would tighten validation). Avoid pointer (optionalfields).
	Message string `json:"message,omitempty"`
}

// NetworkStatus defines the observed state of Network.
type NetworkStatus struct {
	// Conditions are an array of current observed network conditions.
	// +optional
	//
	//nolint:kubeapilinter // Stable v1alpha1 retention: keep custom NetworkCondition slice (not metav1.Condition).
	Conditions []NetworkCondition `json:"conditions,omitempty"`

	// SupportedIPFamilies list the IP families that are available on this network,
	// as determined by the backing network provider (e.g. the IP families of the
	// IPPools referenced by a VSphereDistributedNetwork). Users can inspect this field
	// to understand which IPFamilyPolicy values are valid when creating a NetworkInterface
	// on this network.
	// +optional
	//
	//nolint:kubeapilinter // Stable v1alpha1 retention: avoid MaxItems (would tighten validation).
	SupportedIPFamilies []corev1.IPFamily `json:"supportedIPFamilies,omitempty"`
}

// NetworkReference is an object that points to a Network.
type NetworkReference struct {
	// Kind is the type of resource being referenced.
	//
	//nolint:kubeapilinter // Stable v1alpha1 retention: avoid MaxLength (would tighten validation). Avoid omitempty (requiredfields wire shape).
	Kind string `json:"kind"`

	// Name is the name of resource being referenced.
	//
	//nolint:kubeapilinter // Stable v1alpha1 retention: avoid MaxLength (would tighten validation). Avoid omitempty (requiredfields wire shape).
	Name string `json:"name"`

	// APIVersion of the referent.
	//
	// +optional
	//
	//nolint:kubeapilinter // Stable v1alpha1 retention: keep optional string without pointer (optionalfields).
	APIVersion string `json:"apiVersion,omitempty"`
}

// +genclient
// +kubebuilder:object:root=true

// Network is the Schema for the networks API.
// A Network describes type, class and common attributes of a network available
// in a namespace. A NetworkInterface resource references a Network.
// +kubebuilder:subresource:status
type Network struct {
	metav1.TypeMeta `json:",inline"`

	// metadata is the standard object's metadata.
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// Spec describes the desired network configuration.
	//
	//nolint:kubeapilinter // Stable v1alpha1 retention: keep nested spec without omitzero (requiredfields).
	Spec NetworkSpec `json:"spec,omitempty"`

	// Status reflects the observed state of the network.
	//
	//nolint:kubeapilinter // Stable v1alpha1 retention: keep nested status without omitzero (requiredfields).
	Status NetworkStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// NetworkList contains a list of Network
type NetworkList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Network `json:"items"`
}

func init() {
	RegisterTypeWithScheme(&Network{}, &NetworkList{})
}
