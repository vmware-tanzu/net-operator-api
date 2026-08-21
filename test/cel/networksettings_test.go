// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: Apache-2.0

package cel_test

import (
	"strings"
	"testing"

	netv1alpha1 "github.com/vmware-tanzu/net-operator-api/api/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// nsObj builds a NetworkSettings with only provider set (no prior transition).
func nsObj(name string, provider netv1alpha1.NetworkProvider) *netv1alpha1.NetworkSettings {
	return &netv1alpha1.NetworkSettings{
		ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: nsNamespace},
		Provider:   provider,
	}
}

// nsWithLegacy builds a NetworkSettings representing a completed provider transition.
func nsWithLegacy(name string, provider, legacy netv1alpha1.NetworkProvider) *netv1alpha1.NetworkSettings {
	return &netv1alpha1.NetworkSettings{
		ObjectMeta:     metav1.ObjectMeta{Name: name, Namespace: nsNamespace},
		Provider:       provider,
		LegacyProvider: legacy,
	}
}

// unstrNS builds an unstructured NetworkSettings. Fields are merged into the
// top-level object map, allowing tests to exercise values that Go encoding
// would otherwise reject or omit (e.g. missing required fields, bad enums).
func unstrNS(name string, fields map[string]interface{}) *unstructured.Unstructured {
	obj := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": nsAPIVersion,
			"kind":       nsKind,
			"metadata":   map[string]interface{}{"name": name, "namespace": nsNamespace},
		},
	}
	for k, v := range fields {
		obj.Object[k] = v
	}
	return obj
}

// -----------------------------------------------------------------------
// Provider field
// Schema enum: vsphere-distributed | nsx-tier1 | vpc
// provider is +required in the CRD schema
// -----------------------------------------------------------------------

func TestNetworkSettings_VSphereDistributedProvider_Admitted(t *testing.T) {
	ensureNamespace(t, nsNamespace)
	obj := nsObj("ns-vds", netv1alpha1.NetworkProviderVSphereDistributed)
	if err := k8sClient.Create(testCtx, obj); err != nil {
		t.Fatalf("expected admission, got: %v", err)
	}
	defer func() { _ = k8sClient.Delete(testCtx, obj) }()
}

func TestNetworkSettings_NSXTier1Provider_Admitted(t *testing.T) {
	ensureNamespace(t, nsNamespace)
	obj := nsObj("ns-nsx", netv1alpha1.NetworkProviderNSXTier1)
	if err := k8sClient.Create(testCtx, obj); err != nil {
		t.Fatalf("expected admission, got: %v", err)
	}
	defer func() { _ = k8sClient.Delete(testCtx, obj) }()
}

func TestNetworkSettings_VPCProvider_Admitted(t *testing.T) {
	ensureNamespace(t, nsNamespace)
	obj := nsObj("ns-vpc", netv1alpha1.NetworkProviderVPC)
	if err := k8sClient.Create(testCtx, obj); err != nil {
		t.Fatalf("expected admission, got: %v", err)
	}
	defer func() { _ = k8sClient.Delete(testCtx, obj) }()
}

func TestNetworkSettings_MissingProvider_Rejected(t *testing.T) {
	ensureNamespace(t, nsNamespace)
	obj := unstrNS("ns-no-provider", map[string]interface{}{})
	if err := k8sClient.Create(testCtx, obj); err == nil || !strings.Contains(err.Error(), "provider") {
		t.Fatalf("expected rejection containing %q, got: %v", "provider", err)
	}
}

func TestNetworkSettings_InvalidProviderEnum_Rejected(t *testing.T) {
	ensureNamespace(t, nsNamespace)
	obj := unstrNS("ns-bad-provider", map[string]interface{}{"provider": "unknown"})
	if err := k8sClient.Create(testCtx, obj); err == nil || !strings.Contains(err.Error(), "Unsupported value") {
		t.Fatalf("expected rejection containing %q, got: %v", "Unsupported value", err)
	}
}

// -----------------------------------------------------------------------
// legacyProvider field-level CEL
// Rule: self in ['vsphere-distributed', 'nsx-tier1']
// Rationale: vpc cannot be a legacy provider; transitions are always away
// from vSphere Distributed or NSX Tier-1, never away from VPC.
// -----------------------------------------------------------------------

func TestNetworkSettings_LegacyProviderVSphereDistributedWithVPCCurrent_Admitted(t *testing.T) {
	ensureNamespace(t, nsNamespace)
	obj := nsWithLegacy("ns-leg-vds",
		netv1alpha1.NetworkProviderVPC,
		netv1alpha1.NetworkProviderVSphereDistributed,
	)
	if err := k8sClient.Create(testCtx, obj); err != nil {
		t.Fatalf("expected admission, got: %v", err)
	}
	defer func() { _ = k8sClient.Delete(testCtx, obj) }()
}

func TestNetworkSettings_LegacyProviderNSXTier1WithVPCCurrent_Admitted(t *testing.T) {
	ensureNamespace(t, nsNamespace)
	obj := nsWithLegacy("ns-leg-nsx",
		netv1alpha1.NetworkProviderVPC,
		netv1alpha1.NetworkProviderNSXTier1,
	)
	if err := k8sClient.Create(testCtx, obj); err != nil {
		t.Fatalf("expected admission, got: %v", err)
	}
	defer func() { _ = k8sClient.Delete(testCtx, obj) }()
}

func TestNetworkSettings_LegacyProviderVPC_Rejected(t *testing.T) {
	ensureNamespace(t, nsNamespace)
	obj := nsWithLegacy("ns-leg-vpc",
		netv1alpha1.NetworkProviderVSphereDistributed,
		netv1alpha1.NetworkProviderVPC,
	)
	if err := k8sClient.Create(testCtx, obj); err == nil || !strings.Contains(err.Error(), "legacyProvider must be vsphere-distributed or nsx-tier1") {
		t.Fatalf("expected rejection containing %q, got: %v", "legacyProvider must be vsphere-distributed or nsx-tier1", err)
	}
}

func TestNetworkSettings_LegacyProviderInvalidEnum_Rejected(t *testing.T) {
	ensureNamespace(t, nsNamespace)
	obj := unstrNS("ns-leg-bad", map[string]interface{}{
		"provider":       providerVPC,
		"legacyProvider": "unknown",
	})
	if err := k8sClient.Create(testCtx, obj); err == nil || !strings.Contains(err.Error(), "Unsupported value") {
		t.Fatalf("expected rejection containing %q, got: %v", "Unsupported value", err)
	}
}

// -----------------------------------------------------------------------
// legacyProvider resource-level CEL
// Rule: !has(self.legacyProvider) || self.legacyProvider != self.provider
// Rationale: the legacy provider must be distinct from the current provider;
// a same-value pair indicates a misconfigured or no-op transition.
// -----------------------------------------------------------------------

func TestNetworkSettings_ProviderVPCWithLegacyVSphereDistributed_Admitted(t *testing.T) {
	ensureNamespace(t, nsNamespace)
	obj := nsWithLegacy("ns-rl-vpc-vds",
		netv1alpha1.NetworkProviderVPC,
		netv1alpha1.NetworkProviderVSphereDistributed,
	)
	if err := k8sClient.Create(testCtx, obj); err != nil {
		t.Fatalf("expected admission, got: %v", err)
	}
	defer func() { _ = k8sClient.Delete(testCtx, obj) }()
}

func TestNetworkSettings_ProviderVPCWithLegacyNSXTier1_Admitted(t *testing.T) {
	ensureNamespace(t, nsNamespace)
	obj := nsWithLegacy("ns-rl-vpc-nsx",
		netv1alpha1.NetworkProviderVPC,
		netv1alpha1.NetworkProviderNSXTier1,
	)
	if err := k8sClient.Create(testCtx, obj); err != nil {
		t.Fatalf("expected admission, got: %v", err)
	}
	defer func() { _ = k8sClient.Delete(testCtx, obj) }()
}

func TestNetworkSettings_ProviderVSphereDistributedWithLegacyNSXTier1_Admitted(t *testing.T) {
	ensureNamespace(t, nsNamespace)
	obj := nsWithLegacy("ns-rl-vds-nsx",
		netv1alpha1.NetworkProviderVSphereDistributed,
		netv1alpha1.NetworkProviderNSXTier1,
	)
	if err := k8sClient.Create(testCtx, obj); err != nil {
		t.Fatalf("expected admission, got: %v", err)
	}
	defer func() { _ = k8sClient.Delete(testCtx, obj) }()
}

func TestNetworkSettings_ProviderNSXTier1WithLegacyVSphereDistributed_Admitted(t *testing.T) {
	ensureNamespace(t, nsNamespace)
	obj := nsWithLegacy("ns-rl-nsx-vds",
		netv1alpha1.NetworkProviderNSXTier1,
		netv1alpha1.NetworkProviderVSphereDistributed,
	)
	if err := k8sClient.Create(testCtx, obj); err != nil {
		t.Fatalf("expected admission, got: %v", err)
	}
	defer func() { _ = k8sClient.Delete(testCtx, obj) }()
}

func TestNetworkSettings_NoLegacyProvider_Admitted(t *testing.T) {
	ensureNamespace(t, nsNamespace)
	obj := nsObj("ns-rl-no-leg", netv1alpha1.NetworkProviderVPC)
	if err := k8sClient.Create(testCtx, obj); err != nil {
		t.Fatalf("expected admission, got: %v", err)
	}
	defer func() { _ = k8sClient.Delete(testCtx, obj) }()
}

func TestNetworkSettings_ProviderVSphereDistributedWithLegacySame_Rejected(t *testing.T) {
	ensureNamespace(t, nsNamespace)
	obj := nsWithLegacy("ns-rl-vds-vds",
		netv1alpha1.NetworkProviderVSphereDistributed,
		netv1alpha1.NetworkProviderVSphereDistributed,
	)
	if err := k8sClient.Create(testCtx, obj); err == nil || !strings.Contains(err.Error(), "legacyProvider must differ from provider") {
		t.Fatalf("expected rejection containing %q, got: %v", "legacyProvider must differ from provider", err)
	}
}

func TestNetworkSettings_ProviderNSXTier1WithLegacySame_Rejected(t *testing.T) {
	ensureNamespace(t, nsNamespace)
	obj := nsWithLegacy("ns-rl-nsx-nsx",
		netv1alpha1.NetworkProviderNSXTier1,
		netv1alpha1.NetworkProviderNSXTier1,
	)
	if err := k8sClient.Create(testCtx, obj); err == nil || !strings.Contains(err.Error(), "legacyProvider must differ from provider") {
		t.Fatalf("expected rejection containing %q, got: %v", "legacyProvider must differ from provider", err)
	}
}

// -----------------------------------------------------------------------
// Status & WorkloadCapabilities CEL and schema validation
// -----------------------------------------------------------------------

func TestNetworkSettings_StatusWorkloadCapabilities_Admitted(t *testing.T) {
	ensureNamespace(t, nsNamespace)
	obj := nsObj("ns-status-valid", netv1alpha1.NetworkProviderVPC)
	if err := k8sClient.Create(testCtx, obj); err != nil {
		t.Fatalf("create: %v", err)
	}
	defer func() { _ = k8sClient.Delete(testCtx, obj) }()

	if err := k8sClient.Get(testCtx, client.ObjectKey{Name: obj.Name, Namespace: nsNamespace}, obj); err != nil {
		t.Fatalf("get: %v", err)
	}

	obj.Status = &netv1alpha1.NetworkSettingsStatus{
		WorkloadCapabilities: []netv1alpha1.WorkloadCapability{
			{
				Type:                netv1alpha1.WorkloadTypeVKSCluster,
				SupportedIPFamilies: []string{netv1alpha1.WorkloadIPFamilyIPv4, netv1alpha1.WorkloadIPFamilyIPv6, netv1alpha1.WorkloadIPFamilyDualStack},
				DefaultIPFamily:     netv1alpha1.WorkloadIPFamilyDualStack,
				Message:             "Full Dual-Stack, single-stack IPv4, and single-stack IPv6 are all supported.",
			},
			{
				Type:                netv1alpha1.WorkloadTypePodVM,
				SupportedIPFamilies: []string{netv1alpha1.WorkloadIPFamilyDualStack},
				DefaultIPFamily:     netv1alpha1.WorkloadIPFamilyDualStack,
			},
			{
				Type:                netv1alpha1.WorkloadTypeVirtualMachine,
				SupportedIPFamilies: []string{netv1alpha1.WorkloadIPFamilyIPv4, netv1alpha1.WorkloadIPFamilyIPv6},
				DefaultIPFamily:     netv1alpha1.WorkloadIPFamilyIPv4,
			},
		},
	}
	if err := k8sClient.Status().Update(testCtx, obj); err != nil {
		t.Fatalf("expected admission for status update, got: %v", err)
	}
}

func TestNetworkSettings_StatusWorkloadCapabilities_VDS_Admitted(t *testing.T) {
	ensureNamespace(t, nsNamespace)
	obj := nsObj("ns-status-vds", netv1alpha1.NetworkProviderVSphereDistributed)
	if err := k8sClient.Create(testCtx, obj); err != nil {
		t.Fatalf("create: %v", err)
	}
	defer func() { _ = k8sClient.Delete(testCtx, obj) }()

	if err := k8sClient.Get(testCtx, client.ObjectKey{Name: obj.Name, Namespace: nsNamespace}, obj); err != nil {
		t.Fatalf("get: %v", err)
	}

	obj.Status = &netv1alpha1.NetworkSettingsStatus{
		WorkloadCapabilities: []netv1alpha1.WorkloadCapability{
			{
				Type:                netv1alpha1.WorkloadTypeVKSCluster,
				SupportedIPFamilies: []string{netv1alpha1.WorkloadIPFamilyIPv4},
				DefaultIPFamily:     netv1alpha1.WorkloadIPFamilyIPv4,
				Message:             "VDS topology currently only supports IPv4.",
			},
			{
				Type:                netv1alpha1.WorkloadTypePodVM,
				SupportedIPFamilies: []string{netv1alpha1.WorkloadIPFamilyIPv4},
				DefaultIPFamily:     netv1alpha1.WorkloadIPFamilyIPv4,
				Message:             "VDS topology currently only supports IPv4.",
			},
			{
				Type:                netv1alpha1.WorkloadTypeVirtualMachine,
				SupportedIPFamilies: []string{netv1alpha1.WorkloadIPFamilyIPv4},
				DefaultIPFamily:     netv1alpha1.WorkloadIPFamilyIPv4,
				Message:             "VDS topology currently only supports IPv4.",
			},
		},
	}
	if err := k8sClient.Status().Update(testCtx, obj); err != nil {
		t.Fatalf("expected admission for status update, got: %v", err)
	}
}

func TestNetworkSettings_StatusWorkloadCapabilities_NSXTier1_Admitted(t *testing.T) {
	ensureNamespace(t, nsNamespace)
	obj := nsObj("ns-status-nsx", netv1alpha1.NetworkProviderNSXTier1)
	if err := k8sClient.Create(testCtx, obj); err != nil {
		t.Fatalf("create: %v", err)
	}
	defer func() { _ = k8sClient.Delete(testCtx, obj) }()

	if err := k8sClient.Get(testCtx, client.ObjectKey{Name: obj.Name, Namespace: nsNamespace}, obj); err != nil {
		t.Fatalf("get: %v", err)
	}

	obj.Status = &netv1alpha1.NetworkSettingsStatus{
		WorkloadCapabilities: []netv1alpha1.WorkloadCapability{
			{
				Type:                netv1alpha1.WorkloadTypeVKSCluster,
				SupportedIPFamilies: []string{netv1alpha1.WorkloadIPFamilyIPv4},
				DefaultIPFamily:     netv1alpha1.WorkloadIPFamilyIPv4,
				Message:             "NSX-T1 topology currently only supports IPv4.",
			},
		},
	}
	if err := k8sClient.Status().Update(testCtx, obj); err != nil {
		t.Fatalf("expected admission for status update, got: %v", err)
	}
}

func TestNetworkSettings_StatusWorkloadCapabilities_SingleFamilyOptions_Admitted(t *testing.T) {
	families := [][]string{
		{netv1alpha1.WorkloadIPFamilyIPv4},
		{netv1alpha1.WorkloadIPFamilyIPv6},
		{netv1alpha1.WorkloadIPFamilyDualStack},
	}
	for _, fam := range families {
		name := "ns-status-fam-" + strings.ToLower(fam[0])
		ensureNamespace(t, nsNamespace)
		obj := nsObj(name, netv1alpha1.NetworkProviderVPC)
		if err := k8sClient.Create(testCtx, obj); err != nil {
			t.Fatalf("create: %v", err)
		}
		defer func() { _ = k8sClient.Delete(testCtx, obj) }()

		if err := k8sClient.Get(testCtx, client.ObjectKey{Name: obj.Name, Namespace: nsNamespace}, obj); err != nil {
			t.Fatalf("get: %v", err)
		}

		obj.Status = &netv1alpha1.NetworkSettingsStatus{
			WorkloadCapabilities: []netv1alpha1.WorkloadCapability{
				{
					Type:                netv1alpha1.WorkloadTypeVKSCluster,
					SupportedIPFamilies: fam,
				},
			},
		}
		if err := k8sClient.Status().Update(testCtx, obj); err != nil {
			t.Fatalf("expected admission for family %v, got: %v", fam, err)
		}
	}
}

func TestNetworkSettings_StatusWorkloadCapabilities_InvalidIPFamily_Rejected(t *testing.T) {
	ensureNamespace(t, nsNamespace)
	obj := nsObj("ns-status-invalid-ipfam", netv1alpha1.NetworkProviderVPC)
	if err := k8sClient.Create(testCtx, obj); err != nil {
		t.Fatalf("create: %v", err)
	}
	defer func() { _ = k8sClient.Delete(testCtx, obj) }()

	if err := k8sClient.Get(testCtx, client.ObjectKey{Name: obj.Name, Namespace: nsNamespace}, obj); err != nil {
		t.Fatalf("get: %v", err)
	}

	obj.Status = &netv1alpha1.NetworkSettingsStatus{
		WorkloadCapabilities: []netv1alpha1.WorkloadCapability{
			{
				Type:                netv1alpha1.WorkloadTypeVKSCluster,
				SupportedIPFamilies: []string{"IPv5"},
			},
		},
	}
	if err := k8sClient.Status().Update(testCtx, obj); err == nil || !strings.Contains(err.Error(), "supportedIPFamilies must only contain 'IPv4', 'IPv6', or 'DualStack'") {
		t.Fatalf("expected rejection containing %q, got: %v", "supportedIPFamilies must only contain 'IPv4', 'IPv6', or 'DualStack'", err)
	}
}

func TestNetworkSettings_StatusWorkloadCapabilities_CaseSensitivity_Rejected(t *testing.T) {
	ensureNamespace(t, nsNamespace)
	obj := nsObj("ns-status-case-sens", netv1alpha1.NetworkProviderVPC)
	if err := k8sClient.Create(testCtx, obj); err != nil {
		t.Fatalf("create: %v", err)
	}
	defer func() { _ = k8sClient.Delete(testCtx, obj) }()

	if err := k8sClient.Get(testCtx, client.ObjectKey{Name: obj.Name, Namespace: nsNamespace}, obj); err != nil {
		t.Fatalf("get: %v", err)
	}

	obj.Status = &netv1alpha1.NetworkSettingsStatus{
		WorkloadCapabilities: []netv1alpha1.WorkloadCapability{
			{
				Type:                netv1alpha1.WorkloadTypeVKSCluster,
				SupportedIPFamilies: []string{"dualstack"},
			},
		},
	}
	if err := k8sClient.Status().Update(testCtx, obj); err == nil || !strings.Contains(err.Error(), "supportedIPFamilies must only contain 'IPv4', 'IPv6', or 'DualStack'") {
		t.Fatalf("expected rejection containing %q, got: %v", "supportedIPFamilies must only contain 'IPv4', 'IPv6', or 'DualStack'", err)
	}
}

func TestNetworkSettings_StatusWorkloadCapabilities_ExceedMaxItems_Rejected(t *testing.T) {
	ensureNamespace(t, nsNamespace)
	obj := nsObj("ns-status-max-items", netv1alpha1.NetworkProviderVPC)
	if err := k8sClient.Create(testCtx, obj); err != nil {
		t.Fatalf("create: %v", err)
	}
	defer func() { _ = k8sClient.Delete(testCtx, obj) }()

	if err := k8sClient.Get(testCtx, client.ObjectKey{Name: obj.Name, Namespace: nsNamespace}, obj); err != nil {
		t.Fatalf("get: %v", err)
	}

	obj.Status = &netv1alpha1.NetworkSettingsStatus{
		WorkloadCapabilities: []netv1alpha1.WorkloadCapability{
			{
				Type:                netv1alpha1.WorkloadTypeVKSCluster,
				SupportedIPFamilies: []string{netv1alpha1.WorkloadIPFamilyIPv4, netv1alpha1.WorkloadIPFamilyIPv6, netv1alpha1.WorkloadIPFamilyDualStack, netv1alpha1.WorkloadIPFamilyIPv4},
			},
		},
	}
	if err := k8sClient.Status().Update(testCtx, obj); err == nil || !isRejected(err) {
		t.Fatalf("expected rejection exceeding maxItems, got: %v", err)
	}
}

func TestNetworkSettings_StatusWorkloadCapabilities_DuplicateType_Rejected(t *testing.T) {
	ensureNamespace(t, nsNamespace)
	obj := nsObj("ns-status-dup-type", netv1alpha1.NetworkProviderVPC)
	if err := k8sClient.Create(testCtx, obj); err != nil {
		t.Fatalf("create: %v", err)
	}
	defer func() { _ = k8sClient.Delete(testCtx, obj) }()

	unstr := &unstructured.Unstructured{}
	unstr.SetAPIVersion(nsAPIVersion)
	unstr.SetKind(nsKind)
	unstr.SetName(obj.Name)
	unstr.SetNamespace(nsNamespace)

	if err := k8sClient.Get(testCtx, client.ObjectKey{Name: obj.Name, Namespace: nsNamespace}, unstr); err != nil {
		t.Fatalf("get: %v", err)
	}

	unstr.Object["status"] = map[string]interface{}{
		"workloadCapabilities": []interface{}{
			map[string]interface{}{
				"type":                "vksCluster",
				"supportedIPFamilies": []interface{}{"IPv4"},
			},
			map[string]interface{}{
				"type":                "vksCluster",
				"supportedIPFamilies": []interface{}{"IPv6"},
			},
		},
	}

	if err := k8sClient.Status().Update(testCtx, unstr); err == nil || !strings.Contains(err.Error(), "uplicate") {
		t.Fatalf("expected rejection for duplicate workload capability type containing 'uplicate', got: %v", err)
	}
}
