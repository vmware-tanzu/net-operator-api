// Copyright (c) 2020-2026 Broadcom. All Rights Reserved.
// Broadcom Confidential. The term "Broadcom" refers to Broadcom Inc.
// and/or its subsidiaries.

package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// IPAddressAllocationFinalizer is a finalizer that allows the controller to perform cleanup
// of resources associated with an IPAddressAllocation before it is removed from the API Server.
const IPAddressAllocationFinalizer = "ipaddressallocation.netoperator.vmware.com"

// IPAddressAllocationConditionType is a string type for the condition types of an IPAddressAllocation.
type IPAddressAllocationConditionType string

const (
	// IPAddressAllocationReady indicates the IP has been successfully allocated.
	IPAddressAllocationReady IPAddressAllocationConditionType = "Ready"
	// IPAddressAllocationFail indicates an error was encountered during allocation.
	IPAddressAllocationFail IPAddressAllocationConditionType = "Failure"
)

// IPAddressAllocationConditionReason describes the reason for the last transition of a condition.
type IPAddressAllocationConditionReason string

const (
	// IPAddressAllocationConditionInvalidRequestedIP is used when the IPAddressAllocation fails due to an invalid RequestedIP.
	IPAddressAllocationConditionInvalidRequestedIP IPAddressAllocationConditionReason = "InvalidRequestedIP"
	// IPAddressAllocationConditionFailureReasonCannotAllocIP is used when the IPAddressAllocation fails because an IP cannot be allocated.
	IPAddressAllocationConditionFailureReasonCannotAllocIP IPAddressAllocationConditionReason = "CannotAllocIP"
	// IPAddressAllocationConditionFailureReasonIPPoolRefRetrievalFailed is used when retrieval of the IPPoolRef has failed.
	IPAddressAllocationConditionFailureReasonIPPoolRefRetrievalFailed IPAddressAllocationConditionReason = "IPPoolRefRetrievalFailed"
)

// IPAddressAllocationCondition describes the state of an IPAddressAllocation at a specific point in time.
type IPAddressAllocationCondition struct {
	// Type is the type of the condition.
	//
	//nolint:kubeapilinter // Stable v1alpha1 retention: avoid MaxLength (would tighten validation). Keep condition type without omitempty (requiredfields wire shape).
	Type IPAddressAllocationConditionType `json:"type"`

	// Status reflects whether the condition is True, False, or Unknown.
	//
	//nolint:kubeapilinter // Stable v1alpha1 retention: keep status without omitempty (requiredfields wire shape).
	Status corev1.ConditionStatus `json:"status"`

	// LastTransitionTime is the timestamp of the last change to the condition's status.
	//
	//nolint:kubeapilinter // Stable v1alpha1 retention: avoid MaxLength (would tighten validation). Avoid pointer (optionalfields).
	LastTransitionTime metav1.Time `json:"lastTransitionTime,omitempty"`

	// Reason provides a machine-readable explanation for the last status transition.
	//
	//nolint:kubeapilinter // Stable v1alpha1 retention: avoid MaxLength (would tighten validation). Avoid pointer (optionalfields).
	Reason IPAddressAllocationConditionReason `json:"reason,omitempty"`

	// Message provides a human-readable explanation for the last status transition.
	//
	//nolint:kubeapilinter // Stable v1alpha1 retention: avoid MaxLength (would tighten validation). Avoid pointer (optionalfields).
	Message string `json:"message,omitempty"`
}

// IPAddressAllocationSpec defines the desired state of an IPAddressAllocation, including the pool reference and an optional requested IP.
type IPAddressAllocationSpec struct {
	// PoolRef is the reference to the network's IP pool within the namespace.
	// It currently only supports reference to a Network.
	//
	//nolint:kubeapilinter // Stable v1alpha1 retention: avoid omitempty (requiredfields wire shape).
	PoolRef corev1.TypedLocalObjectReference `json:"poolRef"`

	// RequestedIP is an optional field for a user to specify a particular IP they want to request.
	// If omitted, the system will allocate a single IP address.
	//
	//nolint:kubeapilinter // Stable v1alpha1 retention: keep value-typed optional string without pointer (optionalfields).
	RequestedIP string `json:"requestedIP,omitempty"`
}

// IPAddressAllocationStatus contains the current status of an IPAddressAllocation, including the allocated IP address and conditions.
type IPAddressAllocationStatus struct {
	// IPAddress is the actually allocated IP address.
	//
	//nolint:kubeapilinter // Stable v1alpha1 retention: keep value-typed optional string without pointer (optionalfields).
	IPAddress string `json:"ipAddress,omitempty"`

	// Conditions provide detailed information about the status of the allocation.
	//
	//nolint:kubeapilinter // Stable v1alpha1 retention: keep custom IPAddressAllocationCondition slice (not metav1.Condition).
	Conditions []IPAddressAllocationCondition `json:"conditions,omitempty"`
}

// +genclient
// +kubebuilder:object:root=true

// IPAddressAllocation represents a request for IP address allocation, including the desired state and current status.
//
//nolint:kubeapilinter // Stable v1alpha1 retention: ignore kubebuilder:subresource:status marker.
type IPAddressAllocation struct {
	metav1.TypeMeta `json:",inline"`

	// metadata is the standard object's metadata.
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// Spec describes the desired IP address allocation.
	//
	//nolint:kubeapilinter // Stable v1alpha1 retention: keep nested spec without omitzero (requiredfields).
	Spec IPAddressAllocationSpec `json:"spec,omitempty"`

	// Status reflects the observed state of the IP address allocation.
	//
	//nolint:kubeapilinter // Stable v1alpha1 retention: keep nested status without omitzero (requiredfields).
	Status IPAddressAllocationStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// IPAddressAllocationList is a list of IPAddressAllocation objects.
type IPAddressAllocationList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []IPAddressAllocation `json:"items"`
}

// init function registers the IPAddressAllocation type with the scheme.
func init() {
	RegisterTypeWithScheme(&IPAddressAllocation{}, &IPAddressAllocationList{})
}
