// Copyright (c) 2026 Broadcom. All Rights Reserved.
// Broadcom Confidential. The term "Broadcom" refers to Broadcom Inc.
// and/or its subsidiaries.

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +kubebuilder:resource:path=vspherenetworktopologyproviders,singular=vspherenetworktopologyprovider,shortName=vsntp
type VSphereNetworkTopologyProviderSpec struct {
	// TODO: placeholder. Unclear if anything is needed here.
}

// +genclient
// +kubebuilder:object:root=true

// VSphereNetworkTopologyProvider is the Schema for the vspherenetworktopologyproviders API.
// A VSphereNetworkTopologyProvider represents a topology provider for vSphere distributed networks for a Supervisor.
type VSphereNetworkTopologyProvider struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec VSphereNetworkTopologyProviderSpec `json:"spec,omitempty"`
}

// +kubebuilder:object:root=true

// VSphereNetworkTopologyProviderList contains a list of VSphereNetworkTopologyProvider.
type VSphereNetworkTopologyProviderList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []VSphereNetworkTopologyProvider `json:"items"`
}

func init() {
	RegisterTypeWithScheme(&VSphereNetworkTopologyProvider{}, &VSphereNetworkTopologyProviderList{})
}
