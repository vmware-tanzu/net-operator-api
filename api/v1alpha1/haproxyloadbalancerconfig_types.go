// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: Apache-2.0

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// HAProxyLoadBalancerConfigSpec defines the configuration for an HAProxyLoadBalancerConfig instance.
// The spec is used to configure the HAProxyLoadBalancer instance to correctly route traffic to services.
// This spec supports HAProxyLoadBalancerConfig Dataplane API 2.0+ sidecar
//
// +kubebuilder:validation:XValidation:rule="!has(oldSelf.virtualServerIPPools) || (has(self.virtualServerIPPools) && oldSelf.virtualServerIPPools.all(x, self.virtualServerIPPools.exists(y, y.name == x.name)))",message="entries may not be removed from virtualServerIPPools"
// +kubebuilder:validation:XValidation:rule="!has(oldSelf.virtualServerIPRanges) || (has(self.virtualServerIPRanges) && oldSelf.virtualServerIPRanges.all(x, self.virtualServerIPRanges.exists(y, y.startingAddress == x.startingAddress)))",message="entries may not be removed from virtualServerIPRanges"
type HAProxyLoadBalancerConfigSpec struct {
	// EndPointURLs is a list of the addresses for the DataPlane API servers used
	// to configure HAProxy.
	// One or more DataPlane API endpoints are possible due to the following topologies:
	// Single Node Topology
	// Multi-Node Active/Passive Topology
	// The strings should include the host, port, and API version, ex.:
	// https://hostname:port/v1
	//
	// +kubebuilder:validation:MinItems=1
	// +kubebuilder:validation:items:MinLength=1
	// +kubebuilder:validation:XValidation:rule="self.all(u, size(u) > 0)",message="endPointURLs entries must be non-empty"
	EndPointURLs []string `json:"endPointURLs"`

	// ServerName is used to verify the hostname on the returned
	// certificates. It is also included
	// in the client's handshake to support virtual hosting unless it is
	// an IP address.
	// Defaults to the host part parsed from Server
	// +optional
	ServerName string `json:"serverName,omitempty"`

	// CredentialSecretRef is an object name of kind Secret.
	// It will be used to access and configure the HAProxy load balancer DataPlane API servers.
	// The following fields are optional:
	//
	// * certificateAuthorityData - CertificateAuthorityData contains PEM-encoded certificate authority certificates.
	//
	// * clientCertificateData - ClientCertificateData contains PEM-encoded data from a client cert file.
	//
	// * clientKeyData - ClientKeyData contains PEM-encoded data from a client key file for TLS.
	//
	// * username - Username is the username for basic authentication. Defaults to "client".
	//
	// * password - Password is the password for basic authentication. Defaults to "cert".
	//
	// Sample of a secret:
	//
	// apiVersion: v1
	// kind: Secret
	// metadata:
	// name: haproxy-lb-config
	// namespace: vmware-system-netop
	// data:
	// 	 certificateAuthorityData: <base64_Encoded>
	// 	 clientCertificateData: <base64_Encoded>
	// 	 clientKeyData: <base64_Encoded>
	//   username: <base64_Encoded>
	//   password: <base64_Encoded>
	// +optional
	CredentialSecretRef ClientSecretReference `json:"credentialSecretRef,omitempty"`

	// CertificateAuthorityData contains PEM-encoded certificate authority
	// certificates used to verify x509 certificates received from the DataPlane API server.
	// When specified, this takes precedence over the certificateAuthorityData in the
	// referenced credentialSecretRef Secret.
	//
	// +optional
	CertificateAuthorityData string `json:"certificateAuthorityData,omitempty"`

	// virtualServerIPPools is the list of IPPools that are used for load balancer IP addresses.
	// If this field is used, effectiveVirtualServerIPPools will be populated with entries of virtualServerIPPools
	// on a successful reconciliation.
	//
	// +optional
	// +listType=atomic
	// +kubebuilder:validation:MaxItems=256
	// +kubebuilder:validation:XValidation:rule="self.all(p, size(p.name) > 0)",message="virtualServerIPPools entries must have non-empty names"
	VirtualServerIPPools []IPPoolReference `json:"virtualServerIPPools,omitempty"`

	// virtualServerIPRanges are IP ranges from which Virtual Server IPs are allocated.
	// If this field is used, on successful reconciliation of virtualServerIPRanges, effectiveVirtualServerIPPools
	// will be populated with names of IP Pools reconciled from it.
	//
	// +optional
	// +listType=atomic
	// +kubebuilder:validation:MaxItems=256
	VirtualServerIPRanges []IPRange `json:"virtualServerIPRanges,omitempty"`
}

// HAProxyLoadBalancerConfigStatus describes the observed state of the HAProxy Load Balancer.
type HAProxyLoadBalancerConfigStatus struct {
	// Conditions describes states of the load balancer at specific points in time.
	//
	// +optional
	Conditions []metav1.Condition `json:"conditions,omitempty"`

	// effectiveVirtualServerIPPools is the union of explicitly referenced pools
	// (spec.virtualServerIPPools) and controller-managed pools derived
	// from spec.virtualServerIPRanges, as of the last successful reconcile.
	//
	// +optional
	// +listType=atomic
	// +kubebuilder:validation:MaxItems=1024
	// +kubebuilder:validation:items:MaxLength=253
	EffectiveVirtualServerIPPools []string `json:"effectiveVirtualServerIPPools,omitempty"`
}

// +genclient
// +genclient:nonNamespaced
// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster

// HAProxyLoadBalancerConfig is the Schema for the HAProxyLoadBalancerConfigs API
type HAProxyLoadBalancerConfig struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   HAProxyLoadBalancerConfigSpec   `json:"spec,omitempty"`
	Status HAProxyLoadBalancerConfigStatus `json:"status,omitempty"`
}

func (hac *HAProxyLoadBalancerConfig) GetConditions() []metav1.Condition {
	return hac.Status.Conditions
}

func (hac *HAProxyLoadBalancerConfig) SetConditions(conditions []metav1.Condition) {
	hac.Status.Conditions = conditions
}

// +kubebuilder:object:root=true

// HAProxyLoadBalancerConfigList contains a list of HAProxyLoadBalancerConfig
type HAProxyLoadBalancerConfigList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []HAProxyLoadBalancerConfig `json:"items"`
}

func init() {
	RegisterTypeWithScheme(&HAProxyLoadBalancerConfig{}, &HAProxyLoadBalancerConfigList{})
}
