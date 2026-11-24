// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: Apache-2.0

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// NSXLoadBalancerConfigSpec defines the configuration for NSX Load Balancer.
type NSXLoadBalancerConfigSpec struct {
	// certificateSecretRef references a TLS Secret containing the default ingress
	// certificate chain (tls.crt) and matching private key (tls.key).
	//
	// +required
	// +kubebuilder:validation:XValidation:rule="has(self.name) && self.name != '' && has(self.namespace) && self.namespace != ''",message="certificateSecretRef.name and namespace must be non-empty"
	CertificateSecretRef ClientSecretReference `json:"certificateSecretRef,omitempty"` //nolint:kubeapilinter // requiredfields: CEL validates the value-typed reference.
}

// +genclient
// +genclient:nonNamespaced
// +kubebuilder:object:root=true
// +kubebuilder:resource:scope=Cluster

// NSXLoadBalancerConfig configures NSX Load Balancer.
type NSXLoadBalancerConfig struct {
	metav1.TypeMeta `json:",inline"`
	// metadata contains the resource name, labels and annotations.
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// spec defines the desired NSX load balancer configuration.
	// +required
	Spec NSXLoadBalancerConfigSpec `json:"spec,omitempty,omitzero"`
}

// +kubebuilder:object:root=true

// NSXLoadBalancerConfigList contains native NSX load balancer configurations.
type NSXLoadBalancerConfigList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []NSXLoadBalancerConfig `json:"items"`
}

func init() {
	RegisterTypeWithScheme(&NSXLoadBalancerConfig{}, &NSXLoadBalancerConfigList{})
}
