// Copyright (c) 2026 Broadcom. All Rights Reserved.
// Broadcom Confidential. The term "Broadcom" refers to Broadcom Inc.
// and/or its subsidiaries.

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +kubebuilder:resource:path=vspheredistributednetworktopologyproviders,singular=vspheredistributednetworktopologyprovider,shortName=vsdntp
type VSphereDistributedNetworkTopologyProviderSpec struct {
	// TODO: placeholder. Unclear if anything is needed here.
}

// +genclient
// +kubebuilder:object:root=true

// VSphereDistributedNetworkTopologyProvider is the Schema for the vspheredistributednetworktopologyproviders API.
// A VSphereDistributedNetworkTopologyProvider represents a topology provider for vSphere distributed networks for a Supervisor.
type VSphereDistributedNetworkTopologyProvider struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec VSphereDistributedNetworkTopologyProviderSpec `json:"spec,omitempty"`
}

// +kubebuilder:object:root=true

// VSphereDistributedNetworkTopologyProviderList contains a list of VSphereDistributedNetworkTopologyProvider.
type VSphereDistributedNetworkTopologyProviderList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []VSphereDistributedNetworkTopologyProvider `json:"items"`
}

func init() {
	RegisterTypeWithScheme(&VSphereDistributedNetworkTopologyProvider{}, &VSphereDistributedNetworkTopologyProviderList{})
}
