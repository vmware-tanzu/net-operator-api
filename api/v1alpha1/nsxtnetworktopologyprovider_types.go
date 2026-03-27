// Copyright (c) 2026 Broadcom. All Rights Reserved.
// Broadcom Confidential. The term "Broadcom" refers to Broadcom Inc.
// and/or its subsidiaries.

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type NSXTNetworkTopologyProviderSpec struct {
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

// +kubebuilder:object:root=true
// +kubebuilder:resource:shortName=nsxtntp,scope=Cluster

// NSXTNetworkTopologyProvider is the Schema for the nsxtnetworktopologyproviders API.
// A NSXTNetworkTopologyProvider represents a topology provider for NSX-T networks for a Supervisor.
type NSXTNetworkTopologyProvider struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec NSXTNetworkTopologyProviderSpec `json:"spec,omitempty"`
}

// +kubebuilder:object:root=true

// NSXTNetworkTopologyProviderList contains a list of NSXTNetworkTopologyProvider.
type NSXTNetworkTopologyProviderList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []NSXTNetworkTopologyProvider `json:"items"`
}

func init() {
	RegisterTypeWithScheme(&NSXTNetworkTopologyProvider{}, &NSXTNetworkTopologyProviderList{})
}
