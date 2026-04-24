// Copyright (c) 2020-2026 Broadcom. All Rights Reserved.
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
	//
	//nolint:kubeapilinter // Stable v1alpha1 retention: avoid MaxLength (would tighten validation).
	Name string `json:"name"`

	// Namespace is the namespace of the resource being referenced. If empty, cluster scoped resource is assumed.
	// +kubebuilder:default:=default
	//
	//nolint:kubeapilinter // Stable v1alpha1 retention: keep optional string without pointer (optionalfields). Retain default (forbiddenmarkers).
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
	//
	//nolint:kubeapilinter // Stable v1alpha1 retention: avoid MaxLength (would tighten validation).
	Type LoadBalancerConfigConditionType `json:"type"`

	// Status is the status of the condition
	// Can be True, False, Unknown
	//
	//nolint:kubeapilinter // Stable v1alpha1 retention: keep status without omitempty (requiredfields wire shape).
	Status corev1.ConditionStatus `json:"status"`

	// Reason is a machine understandable string that gives the reason for the condition's last transition.
	// +optional
	//
	//nolint:kubeapilinter // Stable v1alpha1 retention: avoid MaxLength (would tighten validation). Avoid pointer (optionalfields).
	Reason string `json:"reason,omitempty"`

	// Message is a human-readable message indicating details about last transition.
	// +optional
	//
	//nolint:kubeapilinter // Stable v1alpha1 retention: avoid MaxLength (would tighten validation). Avoid pointer (optionalfields).
	Message string `json:"message,omitempty"`

	// LastTransitionTime is the timestamp for when the LoadBalancerConfig object last transitioned from one status to another.
	//
	//nolint:kubeapilinter // Stable v1alpha1 retention: avoid MaxLength (would tighten validation). Avoid pointer (optionalfields).
	LastTransitionTime metav1.Time `json:"lastTransitionTime,omitempty" patchStrategy:"replace"`
}

// LoadBalancerConfigProviderReference represents the specific load balancer instance that needs to be configured
type LoadBalancerConfigProviderReference struct {
	// APIGroup is the group for the resource being referenced
	//
	//nolint:kubeapilinter // Stable v1alpha1 retention: avoid MaxLength (would tighten validation).
	APIGroup string `json:"apiGroup"`

	// Kind is the type of resource being referenced
	//
	//nolint:kubeapilinter // Stable v1alpha1 retention: avoid MaxLength (would tighten validation). Avoid omitempty (requiredfields wire shape).
	Kind string `json:"kind"`

	// Name is the name of resource being referenced
	//
	//nolint:kubeapilinter // Stable v1alpha1 retention: avoid MaxLength (would tighten validation). Avoid omitempty (requiredfields wire shape).
	Name string `json:"name"`

	// APIVersion is the API version of the referent.
	//
	//nolint:kubeapilinter // Stable v1alpha1 retention: keep optional string without pointer (optionalfields).
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
)

// LoadBalancerConfigSpec defines the desired state of LoadBalancerConfig
type LoadBalancerConfigSpec struct {
	// Type describes type of load balancer.
	// +kubebuilder:validation:Enum=haproxy;avi;foundation
	//
	//nolint:kubeapilinter // Stable v1alpha1 retention: avoid MaxLength (would tighten validation). Keep type without omitempty (requiredfields wire shape).
	Type LoadBalancerConfigType `json:"type"`

	// ProviderRef is reference to a load balancer provider object that provides the details for this type of load balancer
	//
	//nolint:kubeapilinter // Stable v1alpha1 retention: avoid MaxLength (would tighten validation). Avoid omitempty (requiredfields wire shape).
	ProviderRef LoadBalancerConfigProviderReference `json:"providerRef"`
}

// LoadBalancerConfigStatus defines the observed state of LoadBalancerConfig
type LoadBalancerConfigStatus struct {
	// Conditions are an array of current observed load balancer conditions
	//
	//nolint:kubeapilinter // Stable v1alpha1 retention: keep custom LoadBalancerConfigCondition slice (not metav1.Condition).
	Conditions []LoadBalancerConfigCondition `json:"conditions,omitempty"`
}

// +genclient
// +genclient:nonNamespaced
// +kubebuilder:object:root=true
// +kubebuilder:resource:scope=Cluster

// LoadBalancerConfig is the Schema for the LoadBalancerConfigs API
//
//nolint:kubeapilinter // Stable v1alpha1 retention: ignore kubebuilder:subresource:status marker.
type LoadBalancerConfig struct {
	metav1.TypeMeta `json:",inline"`

	// metadata is the standard object's metadata.
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// Spec describes the desired load balancer configuration.
	//
	//nolint:kubeapilinter // Stable v1alpha1 retention: keep nested spec without omitzero (requiredfields).
	Spec LoadBalancerConfigSpec `json:"spec,omitempty"`

	// Status reflects the observed state of the load balancer configuration.
	//
	//nolint:kubeapilinter // Stable v1alpha1 retention: keep nested status without omitzero (requiredfields).
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
