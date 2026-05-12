// Copyright (c) 2026 Broadcom. All Rights Reserved.
// Broadcom Confidential. The term "Broadcom" refers to Broadcom Inc.
// and/or its subsidiaries.

package v1alpha1

// CIDR is an IP network address range in CIDR notation, e.g. "192.168.0.0/16"
// or "2001:db8::/32". The address must be a network address (no host bits
// set): "192.168.1.5/16" is invalid because host bits are set in the host
// portion.
//
// +kubebuilder:validation:MinLength=1
// +kubebuilder:validation:MaxLength=64
// +kubebuilder:validation:XValidation:rule="isIPPrefix(self)",message="must be a valid CIDR (e.g. \"192.168.0.0/16\" or \"2001:db8::/32\")"
type CIDR string
