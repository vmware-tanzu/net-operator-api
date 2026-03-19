// Copyright (c) 2026 Broadcom. All Rights Reserved.
// Broadcom Confidential. The term "Broadcom" refers to Broadcom Inc.
// and/or its subsidiaries.

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +kubebuilder:resource:path=ncpnetworktopologyproviders,singular=ncpnetworktopologyprovider,shortName=ncpntp
type NCPNetworkTopologyProviderSpec struct {
	// PodCidrs specifies the CIDR blocks for Pod networking.
	PodCidrs []string `json:"podCidrs,omitempty"`
	// IngressCidrs specifies the CIDR blocks for ingress traffic.
	IngressCidrs []string `json:"ingressCidrs,omitempty"`
	// EgressCidrs specifies the CIDR blocks for egress traffic.
	EgressCidrs []string `json:"egressCidrs,omitempty"`
	// ClusterDistributedSwitch is the vSphere Distributed Switch used for the cluster.
	ClusterDistributedSwitch *string `json:"clusterDistributedSwitch,omitempty"`
	// NsxEdgeCluster is the NSX-T Edge Cluster ID.
	NsxEdgeCluster *string `json:"nsxEdgeCluster,omitempty"`
	// NsxTier0Gateway is the NSX-T Tier-0 gateway path used for the Supervisor's Tier-1 gateway uplink.
	NsxTier0Gateway *string `json:"nsxTier0Gateway,omitempty"`
	// NamespaceSubnetPrefix is the subnet prefix size for namespaces.
	NamespaceSubnetPrefix *int32 `json:"namespaceSubnetPrefix,omitempty"`
	// RoutedMode indicates whether routed mode is enabled.
	RoutedMode *bool `json:"routedMode,omitempty"`
}

// +genclient
// +kubebuilder:object:root=true

// NCPNetworkTopologyProvider is the Schema for the ncpnetworktopologyproviders API.
// A NCPNetworkTopologyProvider represents a topology provider for NSX Container Plugin networks for a Supervisor.
type NCPNetworkTopologyProvider struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec NCPNetworkTopologyProviderSpec `json:"spec,omitempty"`
}

// +kubebuilder:object:root=true

// NCPNetworkTopologyProviderList contains a list of NCPNetworkTopologyProvider.
type NCPNetworkTopologyProviderList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []NCPNetworkTopologyProvider `json:"items"`
}

func init() {
	RegisterTypeWithScheme(&NCPNetworkTopologyProvider{}, &NCPNetworkTopologyProviderList{})
}
