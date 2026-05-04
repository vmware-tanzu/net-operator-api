// Copyright (c) 2026 Broadcom. All Rights Reserved.
// Broadcom Confidential. The term "Broadcom" refers to Broadcom Inc.
// and/or its subsidiaries.

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type NSXTVPCNetworkTopologyProviderSpec struct {
	// NsxProject is the default Project for VPCs in the Supervisor, including the System VPC, and Supervisor Services
	// VPC. It needs to be NSX path of Project.
	NsxProject *string `json:"nsxProject,omitempty"`
	// VpcConnectivityProfile is the configuration for how a VPC is constructed, including it's Transit Gateway
	// Attachments, IP blocks, and other settings on NSX. It needs to be NSX path of VPC Connectivity Profile.
	VpcConnectivityProfile *string `json:"vpcConnectivityProfile,omitempty"`
	// DefaultPrivateCidrs specifies CIDR blocks from which private subnets are allocated. This range must not overlap
	// with those in VpcConnectivityProfile, the Supervisor's Service CIDR, or other services running in the datacenter.
	// You must have at least one CIDR of size 16 or larger to enable Supervisor with VPC networking.
	// If Avi is used, another CIDR of size 64 is needed.
	DefaultPrivateCidrs []string `json:"defaultPrivateCidrs,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:resource:shortName=nsxtvpcntp,scope=Cluster

// NSXTVPCNetworkTopologyProvider is the Schema for the nsxtvpcnetworktopologyproviders API.
// A NSXTVPCNetworkTopologyProvider represents a topology provider for NSX VPC networks for a Supervisor.
type NSXTVPCNetworkTopologyProvider struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec NSXTVPCNetworkTopologyProviderSpec `json:"spec,omitempty"`
}

// +kubebuilder:object:root=true

// NSXTVPCNetworkTopologyProviderList contains a list of NSXTVPCNetworkTopologyProvider.
type NSXTVPCNetworkTopologyProviderList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []NSXTVPCNetworkTopologyProvider `json:"items"`
}

func init() {
	RegisterTypeWithScheme(&NSXTVPCNetworkTopologyProvider{}, &NSXTVPCNetworkTopologyProviderList{})
}
