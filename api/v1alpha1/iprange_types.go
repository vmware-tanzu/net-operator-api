// Copyright (c) 2020-2026 Broadcom. All Rights Reserved.
// Broadcom Confidential. The term "Broadcom" refers to Broadcom Inc.
// and/or its subsidiaries.

package v1alpha1

// IPRange defines a contiguous block of IPv4 addresses.
type IPRange struct {
	// StartingAddress is the first IPv4 address of the range.
	//
	// +kubebuilder:validation:Format=ipv4
	StartingAddress string `json:"startingAddress"`

	// AddressCount is the number of IPv4 addresses in the range.
	//
	// +kubebuilder:validation:Minimum=1
	AddressCount int64 `json:"addressCount"`
}
