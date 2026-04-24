// Copyright (c) 2020-2026 Broadcom. All Rights Reserved.
// Broadcom Confidential. The term "Broadcom" refers to Broadcom Inc.
// and/or its subsidiaries.

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// VMXNET3NetworkInterfaceSpec defines the desired state of VMXNET3NetworkInterface.
type VMXNET3NetworkInterfaceSpec struct {
	// UPTCompatibilityEnabled indicates whether UPT(Universal Pass-through) compatibility is enabled
	// on this network interface.
	//
	//nolint:kubeapilinter // Stable v1alpha1 retention: keep optional bool without pointer (optionalfields).
	UPTCompatibilityEnabled bool `json:"uptCompatibilityEnabled,omitempty"`

	// WakeOnLanEnabled indicates whether wake-on-LAN is enabled on this network interface. Clients
	// can set this property to selectively enable or disable wake-on-LAN.
	//
	//nolint:kubeapilinter // Stable v1alpha1 retention: keep optional bool without pointer (optionalfields).
	WakeOnLanEnabled bool `json:"wakeOnLanEnabled,omitempty"`
}

// VMXNET3NetworkInterfaceStatus is unused. VMXNET3NetworkInterface is a configuration only resource.
type VMXNET3NetworkInterfaceStatus struct {
}

// +genclient
// +kubebuilder:object:root=true

// VMXNET3NetworkInterface is the Schema for the vmxnet3networkinterfaces API.
// It represents configuration of a vSphere VMXNET3 type  network interface card.
//
//nolint:kubeapilinter // Stable v1alpha1 retention: ignore kubebuilder:subresource:status marker.
type VMXNET3NetworkInterface struct {
	metav1.TypeMeta `json:",inline"`

	// metadata is the standard object's metadata.
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// Spec describes the desired VMXNET3 network interface configuration.
	//
	//nolint:kubeapilinter // Stable v1alpha1 retention: keep nested spec without omitzero (requiredfields).
	Spec VMXNET3NetworkInterfaceSpec `json:"spec,omitempty"`

	// Status reflects the observed state of the VMXNET3 network interface configuration.
	//
	//nolint:kubeapilinter // Stable v1alpha1 retention: keep nested status without omitzero (requiredfields).
	Status VMXNET3NetworkInterfaceStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// VMXNET3NetworkInterfaceList contains a list of VMXNET3NetworkInterface
type VMXNET3NetworkInterfaceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []VMXNET3NetworkInterface `json:"items"`
}

func init() {
	RegisterTypeWithScheme(&VMXNET3NetworkInterface{}, &VMXNET3NetworkInterfaceList{})
}
