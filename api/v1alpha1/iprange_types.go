// Copyright (c) 2020-2026 Broadcom. All Rights Reserved.
// Broadcom Confidential. The term "Broadcom" refers to Broadcom Inc.
// and/or its subsidiaries.

package v1alpha1

// IPRange defines a contiguous block of IP addresses.
//
// +kubebuilder:validation:XValidation:rule="self.addressCount >= 1",message="addressCount must be at least 1"
type IPRange struct {
	// startingAddress is the first IP address of the range. Accepts both IPv4 and IPv6 addresses.
	//
	// +required
	// +kubebuilder:validation:MinLength=3
	// +kubebuilder:validation:MaxLength=39
	// +kubebuilder:validation:XValidation:rule="isIP(self)",message="startingAddress must be a valid IPv4 or IPv6 address"
	StartingAddress string `json:"startingAddress,omitempty"`

	// addressCount is the number of IP addresses in the range.
	//
	// +required
	// +kubebuilder:validation:Minimum=1
	AddressCount int64 `json:"addressCount,omitempty"`
}
