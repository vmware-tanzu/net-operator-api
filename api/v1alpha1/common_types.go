// Copyright (c) 2026 Broadcom. All Rights Reserved.
// Broadcom Confidential. The term "Broadcom" refers to Broadcom Inc.
// and/or its subsidiaries.

package v1alpha1

// NetworkProvider identifies the network provider for a namespace-level
// network configuration. The same type is used by both NamespaceNetworkConfiguration
// (to declare the desired provider) and NetworkSettings (to reflect the active
// provider realized in a namespace).
//
// +kubebuilder:validation:Enum=vsphere-distributed;nsx-tier1;vpc
type NetworkProvider string

const (
	// NetworkProviderVSphereDistributed indicates that the namespace uses vSphere
	// Distributed (VDS) networking.
	NetworkProviderVSphereDistributed NetworkProvider = "vsphere-distributed"

	// NetworkProviderNSXTier1 indicates that the namespace uses NSX Tier-1 networking.
	NetworkProviderNSXTier1 NetworkProvider = "nsx-tier1"

	// NetworkProviderVPC indicates that the namespace uses NSX VPC networking.
	NetworkProviderVPC NetworkProvider = "vpc"
)
