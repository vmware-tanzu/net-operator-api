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

	// certificateAuthorityData contains PEM-encoded certificate authority
	// certificates used to verify x509 certificates received from the DataPlane API server.
	// When specified, this takes precedence over the certificateAuthorityData in the
	// referenced credentialSecretRef Secret.
	//
	// +optional
	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:validation:MaxLength=65536
	CertificateAuthorityData string `json:"certificateAuthorityData,omitempty"`

	// virtualServerIPPools is the list of IPPools that are used for load balancer IP addresses.
	// When specified, entries of virtualServerIPPools are included in status.effectiveVirtualServerIPPools
	// on successful reconciliation. If omitted or empty, status.effectiveVirtualServerIPPools will only
	// contain pools derived from virtualServerIPRanges (or remain empty if ranges are also omitted).
	//
	// +optional
	// +listType=atomic
	// +kubebuilder:validation:MaxItems=256
	// +kubebuilder:validation:XValidation:rule="self.all(p, size(p.name) > 0)",message="virtualServerIPPools entries must have non-empty names"
	VirtualServerIPPools []IPPoolReference `json:"virtualServerIPPools,omitempty"`

	// virtualServerIPRanges are IP ranges from which Virtual Server IPs are allocated.
	// When specified, controller-managed IPPools reconciled from virtualServerIPRanges are included
	// in status.effectiveVirtualServerIPPools on successful reconciliation. If omitted or empty,
	// status.effectiveVirtualServerIPPools will only contain pools from virtualServerIPPools
	// (or remain empty if pools are also omitted).
	//
	// +optional
	// +listType=atomic
	// +kubebuilder:validation:MaxItems=256
	VirtualServerIPRanges []IPRange `json:"virtualServerIPRanges,omitempty"`
}

// HAProxyLoadBalancerConfigStatus describes the observed state of the HAProxy Load Balancer.
type HAProxyLoadBalancerConfigStatus struct {
	// conditions describes states of the load balancer at specific points in time.
	//
	// +optional
	// +listType=map
	// +listMapKey=type
	// +kubebuilder:validation:MaxItems=8
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

// GetConditions returns the status conditions for this HAProxyLoadBalancerConfig.
func (hac *HAProxyLoadBalancerConfig) GetConditions() []metav1.Condition {
	return hac.Status.Conditions
}

// SetConditions sets the status conditions for this HAProxyLoadBalancerConfig.
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
