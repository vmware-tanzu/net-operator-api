// Copyright (c) 2020-2024 Broadcom. All Rights Reserved.
// Broadcom Confidential. The term "Broadcom" refers to Broadcom Inc.
// and/or its subsidiaries.

package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ClientSecretReference contains info to locate an object of Kind Secret
// which contains credential specifications for a load balancer.
type ClientSecretReference struct {
	// Name is the name of resource being referenced.
	// It must conform to DNS-1123 subdomain format.
	//
	// +kubebuilder:validation:Pattern=`^[a-z0-9]([a-z0-9.\-]{0,251}[a-z0-9])?$`
	Name string `json:"name"`
	// Namespace of the resource being referenced. If empty, cluster scoped resource is assumed.
	// +kubebuilder:default:=default
	Namespace string `json:"namespace,omitempty"`
}

// LoadBalancerConfigConditionType is used as a typed string for representing
// LoadBalancerConfig.Status.Conditions.
type LoadBalancerConfigConditionType string

const (
	// LoadBalancerConfigReady is added when the LoadBalancerConfig object has been successfully realized
	LoadBalancerConfigReady LoadBalancerConfigConditionType = "Ready"
	// LoadBalancerConfigFailure is added if any failure is encountered while realizing LoadBalancerConfig object
	LoadBalancerConfigFailure LoadBalancerConfigConditionType = "Failure"
	// LoadBalancerConfigIPPoolPressure condition status is set to True when IPPool is low on free IPs.
	LoadBalancerConfigIPPoolPressure LoadBalancerConfigConditionType = "IPPoolPressure"
)

// LoadBalancerConfigCondition describes the state of a LoadBalancerConfig at a certain point
type LoadBalancerConfigCondition struct {
	// Type is the type of load balancer condition
	// Can be Ready or Failure
	Type LoadBalancerConfigConditionType `json:"type"`
	// Status is the status of the condition
	// Can be True, False, Unknown
	Status corev1.ConditionStatus `json:"status"`
	// Machine understandable string that gives the reason for the condition's last transition
	// +optional
	Reason string `json:"reason,omitempty"`
	// Human-readable message indicating details about last transition
	// +optional
	Message string `json:"message,omitempty"`
	// Provides a timestamp for when the LoadBalancerConfig object last transitioned from one status to another
	LastTransitionTime metav1.Time `json:"lastTransitionTime,omitempty" patchStrategy:"replace"`
}

// LoadBalancerConfigProviderReference represents the specific load balancer instance that needs to be configured
type LoadBalancerConfigProviderReference struct {
	// APIGroup is the group for the resource being referenced
	APIGroup string `json:"apiGroup"`
	// Kind is the type of resource being referenced
	Kind string `json:"kind"`
	// Name is the name of resource being referenced
	Name string `json:"name"`
	// API version of the referent
	APIVersion string `json:"apiVersion,omitempty"`
}

type LoadBalancerConfigType string

const (
	// LoadBalancerConfigTypeHAProxy is the LoadBalancerConfigType for HAProxy.
	LoadBalancerConfigTypeHAProxy LoadBalancerConfigType = "haproxy"

	// LoadBalancerConfigTypeAvi is the LoadBalancerConfigType for Avi.
	LoadBalancerConfigTypeAvi LoadBalancerConfigType = "avi"

	// LoadBalancerConfigTypeFoundation is the FoundationLoadBalancerConfigType for VCF Foundation Load Balancer.
	LoadBalancerConfigTypeFoundation LoadBalancerConfigType = "foundation"

	// LoadBalancerConfigTypeNSX is the LoadBalancerConfigType for VMware NSX.
	// NSX-type configs have no providerRef.
	LoadBalancerConfigTypeNSX LoadBalancerConfigType = "nsx"

	// LoadBalancerConfigTypeNSXRegisteredAvi is the LoadBalancerConfigType for
	// an AVI controller registered and managed by NSX.
	// NSX-registered-AVI configs have no providerRef and are mutable to/from nsx.
	LoadBalancerConfigTypeNSXRegisteredAvi LoadBalancerConfigType = "nsx-registered-avi"
)

// LoadBalancerConfigSpec defines the desired state of LoadBalancerConfig.
//
// +kubebuilder:validation:XValidation:rule="!has(oldSelf.providerRef) || self.providerRef == oldSelf.providerRef",message="spec.providerRef is immutable once set"
// +kubebuilder:validation:XValidation:rule="(self.type == 'foundation' && self.providerRef.kind == 'FoundationLoadBalancerConfig') || (self.type == 'avi' && self.providerRef.kind == 'AviLoadBalancerConfig') || (self.type == 'haproxy' && self.providerRef.kind == 'HAProxyLoadBalancerConfig') || self.type in ['nsx', 'nsx-registered-avi']",message="spec.providerRef.kind must match spec.type"
type LoadBalancerConfigSpec struct {
	// Type describes type of load balancer.
	//
	// +kubebuilder:validation:Enum=haproxy;avi;foundation;nsx;nsx-registered-avi
	// +kubebuilder:validation:XValidation:rule="self == oldSelf || (oldSelf in ['nsx', 'nsx-registered-avi'] && self in ['nsx', 'nsx-registered-avi'])",message="spec.type is immutable except for transitions between nsx and nsx-registered-avi"
	Type LoadBalancerConfigType `json:"type"`

	// providerRef is a reference to a load balancer provider object that provides the details for this type of load balancer.
	// Not set for nsx and nsx-registered-avi types.
	//
	// +optional
	ProviderRef *LoadBalancerConfigProviderReference `json:"providerRef,omitempty"`
}

// LoadBalancerConfigStatus defines the observed state of LoadBalancerConfig
type LoadBalancerConfigStatus struct {
	// Conditions is an array of current observed load balancer conditions
	Conditions []LoadBalancerConfigCondition `json:"conditions,omitempty"`
}

// +genclient
// +genclient:nonNamespaced
// +kubebuilder:object:root=true
// +kubebuilder:resource:scope=Cluster

// LoadBalancerConfig is the Schema for the LoadBalancerConfigs API
type LoadBalancerConfig struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   LoadBalancerConfigSpec   `json:"spec,omitempty"`
	Status LoadBalancerConfigStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// LoadBalancerConfigList contains a list of LoadBalancerConfig
type LoadBalancerConfigList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []LoadBalancerConfig `json:"items"`
}

func init() {
	RegisterTypeWithScheme(&LoadBalancerConfig{}, &LoadBalancerConfigList{})
}
