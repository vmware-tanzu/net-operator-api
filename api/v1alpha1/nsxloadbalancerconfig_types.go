// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: Apache-2.0

package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	// NSXEdgeIDAnnotation identifies the discovered NSX integration edge.
	NSXEdgeIDAnnotation = "netoperator.vmware.com/nsx-edge-id"
	// NSXEdgeNameAnnotation names the discovered NSX integration edge.
	NSXEdgeNameAnnotation = "netoperator.vmware.com/nsx-edge-name"
	// NSXIngressTLSExternallyManagedAnnotation preserves a user-provided ingress certificate.
	NSXIngressTLSExternallyManagedAnnotation = "netoperator.vmware.com/ingress-tls-externally-managed"
	// NSXLoadBalancerCertificateSynced reports whether the certificate has been published for consumption.
	NSXLoadBalancerCertificateSynced = "CertificateSynced"
)

// NSXLoadBalancerConfigSpec configures native NSX load balancing.
type NSXLoadBalancerConfigSpec struct {
	// certificateSecretRef identifies a TLS Secret containing the default HTTPS
	// ingress certificate chain in tls.crt and its matching private key in tls.key.
	// Both the name and namespace must be specified. Replacing the Secret updates
	// the desired certificate; publication does not imply application by NSX.
	//
	// +required
	// +kubebuilder:validation:XValidation:rule="has(self.name) && self.name != '' && has(self.namespace) && self.namespace != ''",message="certificateSecretRef.name and namespace are required"
	CertificateSecretRef corev1.SecretReference `json:"certificateSecretRef,omitempty"`
}

// NSXLoadBalancerConfigStatus reports publication of the native ingress certificate.
type NSXLoadBalancerConfigStatus struct {
	// conditions describes certificate publication. CertificateSynced does not
	// indicate whether NSX has applied the published certificate.
	//
	// +optional
	// +listType=map
	// +listMapKey=type
	// +kubebuilder:validation:MaxItems=32
	Conditions []metav1.Condition `json:"conditions,omitempty"`
}

// +genclient
// +genclient:nonNamespaced
// +kubebuilder:object:root=true
// +kubebuilder:resource:scope=Cluster
// +kubebuilder:subresource:status

// NSXLoadBalancerConfig is the configuration for a native NSX load balancer.
type NSXLoadBalancerConfig struct {
	metav1.TypeMeta   `json:",inline"`
	// metadata identifies this configuration.
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// spec defines the desired native load balancer configuration.
	// +required
	Spec NSXLoadBalancerConfigSpec `json:"spec,omitempty,omitzero"`
	// status describes the observed certificate publication state.
	// +optional
	Status *NSXLoadBalancerConfigStatus `json:"status,omitempty,omitzero"`
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
