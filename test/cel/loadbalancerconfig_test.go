// Copyright (c) 2024-2026 Broadcom. All Rights Reserved.
// Broadcom Confidential. The term "Broadcom" refers to Broadcom Inc.
// and/or its subsidiaries.

package cel_test

import (
	"testing"

	netv1alpha1 "github.com/vmware-tanzu/net-operator-api/api/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func lbcWithType(name string, lbType netv1alpha1.LoadBalancerConfigType, providerKind string) *netv1alpha1.LoadBalancerConfig {
	obj := &netv1alpha1.LoadBalancerConfig{
		ObjectMeta: metav1.ObjectMeta{Name: name},
		Spec: netv1alpha1.LoadBalancerConfigSpec{
			Type: lbType,
		},
	}
	if providerKind != "" {
		obj.Spec.ProviderRef = netv1alpha1.LoadBalancerConfigProviderReference{
			APIGroup:   "netoperator.vmware.com",
			Kind:       providerKind,
			Name:       "my-provider",
			APIVersion: "netoperator.vmware.com/v1alpha1",
		}
	}
	return obj
}

func TestLBC_ValidFoundation_Admitted(t *testing.T) {
	obj := lbcWithType("lbc-valid-foundation", netv1alpha1.LoadBalancerConfigTypeFoundation, "FoundationLoadBalancerConfig")
	if err := k8sClient.Create(testCtx, obj); err != nil {
		t.Fatalf("expected admission, got: %v", err)
	}
	_ = k8sClient.Delete(testCtx, obj)
}

func TestLBC_FoundationWithAviKind_Rejected(t *testing.T) {
	obj := lbcWithType("lbc-bad-kind", netv1alpha1.LoadBalancerConfigTypeFoundation, "AviLoadBalancerConfig")
	if err := k8sClient.Create(testCtx, obj); !isRejected(err) {
		t.Fatalf("expected rejection for type/kind mismatch, got: %v", err)
	}
}

func TestLBC_NSXWithNoProviderRef_Admitted(t *testing.T) {
	obj := lbcWithType("lbc-nsx-no-ref", netv1alpha1.LoadBalancerConfigTypeNSX, "")
	if err := k8sClient.Create(testCtx, obj); err != nil {
		t.Fatalf("expected admission for nsx without providerRef, got: %v", err)
	}
	_ = k8sClient.Delete(testCtx, obj)
}

func TestLBC_NSXRegisteredAvi_Admitted(t *testing.T) {
	obj := lbcWithType("lbc-nsx-reg-avi", netv1alpha1.LoadBalancerConfigTypeNSXRegisteredAvi, "")
	if err := k8sClient.Create(testCtx, obj); err != nil {
		t.Fatalf("expected admission for nsx-registered-avi, got: %v", err)
	}
	_ = k8sClient.Delete(testCtx, obj)
}

func TestLBC_TypeImmutable_FoundationToAvi_Rejected(t *testing.T) {
	obj := lbcWithType("lbc-immutable-type", netv1alpha1.LoadBalancerConfigTypeFoundation, "FoundationLoadBalancerConfig")
	if err := k8sClient.Create(testCtx, obj); err != nil {
		t.Fatalf("create: %v", err)
	}
	defer func() { _ = k8sClient.Delete(testCtx, obj) }()

	latest := &netv1alpha1.LoadBalancerConfig{}
	if err := k8sClient.Get(testCtx, client.ObjectKeyFromObject(obj), latest); err != nil {
		t.Fatalf("get: %v", err)
	}
	latest.Spec.Type = netv1alpha1.LoadBalancerConfigTypeAvi
	latest.Spec.ProviderRef = netv1alpha1.LoadBalancerConfigProviderReference{
		APIGroup: "netoperator.vmware.com",
		Kind:     "AviLoadBalancerConfig",
		Name:     "my-provider",
	}
	if err := k8sClient.Update(testCtx, latest); !isRejected(err) {
		t.Fatalf("expected rejection for foundation→avi type change, got: %v", err)
	}
}

func TestLBC_NSXToNSXRegisteredAvi_Admitted(t *testing.T) {
	obj := lbcWithType("lbc-nsx-carveout", netv1alpha1.LoadBalancerConfigTypeNSX, "")
	if err := k8sClient.Create(testCtx, obj); err != nil {
		t.Fatalf("create: %v", err)
	}
	defer func() { _ = k8sClient.Delete(testCtx, obj) }()

	latest := &netv1alpha1.LoadBalancerConfig{}
	if err := k8sClient.Get(testCtx, client.ObjectKeyFromObject(obj), latest); err != nil {
		t.Fatalf("get: %v", err)
	}
	latest.Spec.Type = netv1alpha1.LoadBalancerConfigTypeNSXRegisteredAvi
	if err := k8sClient.Update(testCtx, latest); err != nil {
		t.Fatalf("expected admission for nsx→nsx-registered-avi, got: %v", err)
	}
}

func TestLBC_NSXRegisteredAviToNSX_Admitted(t *testing.T) {
	obj := lbcWithType("lbc-nsx-back", netv1alpha1.LoadBalancerConfigTypeNSXRegisteredAvi, "")
	if err := k8sClient.Create(testCtx, obj); err != nil {
		t.Fatalf("create: %v", err)
	}
	defer func() { _ = k8sClient.Delete(testCtx, obj) }()

	latest := &netv1alpha1.LoadBalancerConfig{}
	if err := k8sClient.Get(testCtx, client.ObjectKeyFromObject(obj), latest); err != nil {
		t.Fatalf("get: %v", err)
	}
	latest.Spec.Type = netv1alpha1.LoadBalancerConfigTypeNSX
	if err := k8sClient.Update(testCtx, latest); err != nil {
		t.Fatalf("expected admission for nsx-registered-avi→nsx, got: %v", err)
	}
}

func TestLBC_NSXToFoundation_Rejected(t *testing.T) {
	obj := lbcWithType("lbc-nsx-to-foundation", netv1alpha1.LoadBalancerConfigTypeNSX, "")
	if err := k8sClient.Create(testCtx, obj); err != nil {
		t.Fatalf("create: %v", err)
	}
	defer func() { _ = k8sClient.Delete(testCtx, obj) }()

	latest := &netv1alpha1.LoadBalancerConfig{}
	if err := k8sClient.Get(testCtx, client.ObjectKeyFromObject(obj), latest); err != nil {
		t.Fatalf("get: %v", err)
	}
	latest.Spec.Type = netv1alpha1.LoadBalancerConfigTypeFoundation
	latest.Spec.ProviderRef = netv1alpha1.LoadBalancerConfigProviderReference{
		APIGroup: "netoperator.vmware.com",
		Kind:     "FoundationLoadBalancerConfig",
		Name:     "my-provider",
	}
	if err := k8sClient.Update(testCtx, latest); !isRejected(err) {
		t.Fatalf("expected rejection for nsx→foundation type change, got: %v", err)
	}
}

func TestLBC_ProviderRefImmutableOnceSet_Rejected(t *testing.T) {
	obj := lbcWithType("lbc-ref-immutable", netv1alpha1.LoadBalancerConfigTypeFoundation, "FoundationLoadBalancerConfig")
	if err := k8sClient.Create(testCtx, obj); err != nil {
		t.Fatalf("create: %v", err)
	}
	defer func() { _ = k8sClient.Delete(testCtx, obj) }()

	latest := &netv1alpha1.LoadBalancerConfig{}
	if err := k8sClient.Get(testCtx, client.ObjectKeyFromObject(obj), latest); err != nil {
		t.Fatalf("get: %v", err)
	}
	// Attempt to change the providerRef name after it has been set.
	latest.Spec.ProviderRef = netv1alpha1.LoadBalancerConfigProviderReference{
		APIGroup: "netoperator.vmware.com",
		Kind:     "FoundationLoadBalancerConfig",
		Name:     "different-provider",
	}
	if err := k8sClient.Update(testCtx, latest); !isRejected(err) {
		t.Fatalf("expected rejection for providerRef change once set, got: %v", err)
	}
}

// --- providerRef required for foundation / avi / haproxy ---

func TestLBC_ValidAvi_Admitted(t *testing.T) {
	obj := lbcWithType("lbc-valid-avi", netv1alpha1.LoadBalancerConfigTypeAvi, "AviLoadBalancerConfig")
	if err := k8sClient.Create(testCtx, obj); err != nil {
		t.Fatalf("expected admission, got: %v", err)
	}
	_ = k8sClient.Delete(testCtx, obj)
}

func TestLBC_ValidHAProxy_Admitted(t *testing.T) {
	obj := lbcWithType("lbc-valid-haproxy", netv1alpha1.LoadBalancerConfigTypeHAProxy, "HAProxyLoadBalancerConfig")
	if err := k8sClient.Create(testCtx, obj); err != nil {
		t.Fatalf("expected admission, got: %v", err)
	}
	_ = k8sClient.Delete(testCtx, obj)
}

func TestLBC_FoundationWithEmptyProviderRef_Rejected(t *testing.T) {
	obj := lbcWithType("lbc-foundation-no-ref", netv1alpha1.LoadBalancerConfigTypeFoundation, "")
	if err := k8sClient.Create(testCtx, obj); !isRejected(err) {
		t.Fatalf("expected rejection for foundation without providerRef, got: %v", err)
	}
}

func TestLBC_AviWithEmptyProviderRef_Rejected(t *testing.T) {
	obj := lbcWithType("lbc-avi-no-ref", netv1alpha1.LoadBalancerConfigTypeAvi, "")
	if err := k8sClient.Create(testCtx, obj); !isRejected(err) {
		t.Fatalf("expected rejection for avi without providerRef, got: %v", err)
	}
}

func TestLBC_HAProxyWithEmptyProviderRef_Rejected(t *testing.T) {
	obj := lbcWithType("lbc-haproxy-no-ref", netv1alpha1.LoadBalancerConfigTypeHAProxy, "")
	if err := k8sClient.Create(testCtx, obj); !isRejected(err) {
		t.Fatalf("expected rejection for haproxy without providerRef, got: %v", err)
	}
}

// --- providerRef must be zero for nsx / nsx-registered-avi ---

func TestLBC_NSXWithProviderRef_Rejected(t *testing.T) {
	obj := lbcWithType("lbc-nsx-with-ref", netv1alpha1.LoadBalancerConfigTypeNSX, "SomeLoadBalancerConfig")
	if err := k8sClient.Create(testCtx, obj); !isRejected(err) {
		t.Fatalf("expected rejection for nsx with non-empty providerRef, got: %v", err)
	}
}

func TestLBC_NSXRegisteredAviWithProviderRef_Rejected(t *testing.T) {
	obj := lbcWithType("lbc-nsx-reg-avi-with-ref", netv1alpha1.LoadBalancerConfigTypeNSXRegisteredAvi, "SomeLoadBalancerConfig")
	if err := k8sClient.Create(testCtx, obj); !isRejected(err) {
		t.Fatalf("expected rejection for nsx-registered-avi with non-empty providerRef, got: %v", err)
	}
}
