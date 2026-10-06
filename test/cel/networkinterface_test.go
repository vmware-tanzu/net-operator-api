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

// nifObj builds a NetworkInterface with only networkName set.
func nifObj(name string) *netv1alpha1.NetworkInterface {
	return &netv1alpha1.NetworkInterface{
		ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: nifNamespace},
		Spec:       netv1alpha1.NetworkInterfaceSpec{NetworkName: testNetName},
	}
}

// nifWithRequestedIPs builds a NetworkInterface with requestedIPs set.
func nifWithRequestedIPs(name string, requestedIPs ...string) *netv1alpha1.NetworkInterface {
	obj := nifObj(name)
	obj.Spec.RequestedIPs = requestedIPs
	return obj
}

// nifWithPolicyAndRequestedIPs builds a NetworkInterface with both ipFamilyPolicy and
// requestedIPs set.
func nifWithPolicyAndRequestedIPs(name string, policy netv1alpha1.NetworkInterfaceIPFamilyPolicy, requestedIPs ...string) *netv1alpha1.NetworkInterface {
	obj := nifWithRequestedIPs(name, requestedIPs...)
	obj.Spec.IPFamilyPolicy = policy
	return obj
}

// unstrNIF builds an unstructured NetworkInterface. Fields are merged into
// spec, allowing tests to exercise values that Go encoding would otherwise
// reject or omit.
func unstrNIF(name string, specFields map[string]interface{}) *unstructured.Unstructured {
	spec := map[string]interface{}{"networkName": testNetName}
	for k, v := range specFields {
		spec[k] = v
	}
	return &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": nifAPIVersion,
			"kind":       nifKind,
			"metadata":   map[string]interface{}{"name": name, "namespace": nifNamespace},
			"spec":       spec,
		},
	}
}

// -----------------------------------------------------------------------
// requestedIPs field
// Schema: optional list, maxItems=2, items minLength=3/maxLength=39
// Item-level CEL: isIP(self)
// List-level CEL: at most one IPv4 and one IPv6 (no same-family pair)
// -----------------------------------------------------------------------

func TestNetworkInterface_NoRequestedIPs_Admitted(t *testing.T) {
	ensureNamespace(t, nifNamespace)
	obj := nifObj("nif-no-requested-ips")
	if err := k8sClient.Create(testCtx, obj); err != nil {
		t.Fatalf("expected admission, got: %v", err)
	}
	defer func() { _ = k8sClient.Delete(testCtx, obj) }()
}

func TestNetworkInterface_RequestedIPv4Only_Admitted(t *testing.T) {
	ensureNamespace(t, nifNamespace)
	obj := nifWithRequestedIPs("nif-requested-ipv4", "10.0.0.5")
	if err := k8sClient.Create(testCtx, obj); err != nil {
		t.Fatalf("expected admission, got: %v", err)
	}
	defer func() { _ = k8sClient.Delete(testCtx, obj) }()
}

func TestNetworkInterface_RequestedIPv6Only_Admitted(t *testing.T) {
	ensureNamespace(t, nifNamespace)
	obj := nifWithRequestedIPs("nif-requested-ipv6", "2001:db8::5")
	if err := k8sClient.Create(testCtx, obj); err != nil {
		t.Fatalf("expected admission, got: %v", err)
	}
	defer func() { _ = k8sClient.Delete(testCtx, obj) }()
}

func TestNetworkInterface_RequestedDualStack_Admitted(t *testing.T) {
	ensureNamespace(t, nifNamespace)
	obj := nifWithRequestedIPs("nif-requested-dual", "10.0.0.5", "2001:db8::5")
	if err := k8sClient.Create(testCtx, obj); err != nil {
		t.Fatalf("expected admission, got: %v", err)
	}
	defer func() { _ = k8sClient.Delete(testCtx, obj) }()
}

func TestNetworkInterface_RequestedIPsSameFamily_Rejected(t *testing.T) {
	ensureNamespace(t, nifNamespace)
	obj := nifWithRequestedIPs("nif-requested-same-family", "10.0.0.5", "10.0.0.6")
	if err := k8sClient.Create(testCtx, obj); err == nil || !strings.Contains(err.Error(), "requestedIPs must not contain two addresses of the same IP family") {
		t.Fatalf("expected rejection containing %q, got: %v", "requestedIPs must not contain two addresses of the same IP family", err)
	}
}

func TestNetworkInterface_RequestedIPInvalidFormat_Rejected(t *testing.T) {
	ensureNamespace(t, nifNamespace)
	obj := unstrNIF("nif-requested-ip-bad", map[string]interface{}{"requestedIPs": []interface{}{"not-an-ip"}})
	if err := k8sClient.Create(testCtx, obj); err == nil || !strings.Contains(err.Error(), "each requestedIP must be a valid IPv4 or IPv6 address") {
		t.Fatalf("expected rejection containing %q, got: %v", "each requestedIP must be a valid IPv4 or IPv6 address", err)
	}
}

func TestNetworkInterface_RequestedIPTooShort_Rejected(t *testing.T) {
	ensureNamespace(t, nifNamespace)
	obj := unstrNIF("nif-requested-ip-short", map[string]interface{}{"requestedIPs": []interface{}{"::"}})
	if err := k8sClient.Create(testCtx, obj); err == nil {
		t.Fatalf("expected rejection, got admission")
	}
}

// -----------------------------------------------------------------------
// requestedIPs tied to ipFamilyPolicy
// Resource-level CEL: ipFamilyPolicy IPv4Only/IPv6Only forces all
// requestedIPs entries to match that family; DualStack/unset allow either.
// -----------------------------------------------------------------------

func TestNetworkInterface_IPv4OnlyWithIPv4Requested_Admitted(t *testing.T) {
	ensureNamespace(t, nifNamespace)
	obj := nifWithPolicyAndRequestedIPs("nif-v4only-v4", netv1alpha1.NetworkInterfaceIPFamilyPolicyIPv4Only, "10.0.0.5")
	if err := k8sClient.Create(testCtx, obj); err != nil {
		t.Fatalf("expected admission, got: %v", err)
	}
	defer func() { _ = k8sClient.Delete(testCtx, obj) }()
}

func TestNetworkInterface_IPv4OnlyWithIPv6Requested_Rejected(t *testing.T) {
	ensureNamespace(t, nifNamespace)
	obj := nifWithPolicyAndRequestedIPs("nif-v4only-v6", netv1alpha1.NetworkInterfaceIPFamilyPolicyIPv4Only, "2001:db8::5")
	if err := k8sClient.Create(testCtx, obj); err == nil || !strings.Contains(err.Error(), "requestedIPs must only contain IPv4 addresses when ipFamilyPolicy is IPv4Only") {
		t.Fatalf("expected rejection containing %q, got: %v", "requestedIPs must only contain IPv4 addresses when ipFamilyPolicy is IPv4Only", err)
	}
}

func TestNetworkInterface_IPv6OnlyWithIPv6Requested_Admitted(t *testing.T) {
	ensureNamespace(t, nifNamespace)
	obj := nifWithPolicyAndRequestedIPs("nif-v6only-v6", netv1alpha1.NetworkInterfaceIPFamilyPolicyIPv6Only, "2001:db8::5")
	if err := k8sClient.Create(testCtx, obj); err != nil {
		t.Fatalf("expected admission, got: %v", err)
	}
	defer func() { _ = k8sClient.Delete(testCtx, obj) }()
}

func TestNetworkInterface_IPv6OnlyWithIPv4Requested_Rejected(t *testing.T) {
	ensureNamespace(t, nifNamespace)
	obj := nifWithPolicyAndRequestedIPs("nif-v6only-v4", netv1alpha1.NetworkInterfaceIPFamilyPolicyIPv6Only, "10.0.0.5")
	if err := k8sClient.Create(testCtx, obj); err == nil || !strings.Contains(err.Error(), "requestedIPs must only contain IPv6 addresses when ipFamilyPolicy is IPv6Only") {
		t.Fatalf("expected rejection containing %q, got: %v", "requestedIPs must only contain IPv6 addresses when ipFamilyPolicy is IPv6Only", err)
	}
}

func TestNetworkInterface_DualStackWithDualStackRequested_Admitted(t *testing.T) {
	ensureNamespace(t, nifNamespace)
	obj := nifWithPolicyAndRequestedIPs("nif-dual-dual", netv1alpha1.NetworkInterfaceIPFamilyPolicyDualStack, "10.0.0.5", "2001:db8::5")
	if err := k8sClient.Create(testCtx, obj); err != nil {
		t.Fatalf("expected admission, got: %v", err)
	}
	defer func() { _ = k8sClient.Delete(testCtx, obj) }()
}

func TestNetworkInterface_UnsetPolicyWithDualStackRequested_Admitted(t *testing.T) {
	ensureNamespace(t, nifNamespace)
	obj := nifWithRequestedIPs("nif-unset-dual", "10.0.0.5", "2001:db8::5")
	if err := k8sClient.Create(testCtx, obj); err != nil {
		t.Fatalf("expected admission, got: %v", err)
	}
	defer func() { _ = k8sClient.Delete(testCtx, obj) }()
}

func TestNetworkInterface_RequestedIPsTooMany_Rejected(t *testing.T) {
	ensureNamespace(t, nifNamespace)
	obj := unstrNIF("nif-requested-ips-too-many", map[string]interface{}{
		"requestedIPs": []interface{}{"10.0.0.5", "2001:db8::5", "10.0.0.6"},
	})
	if err := k8sClient.Create(testCtx, obj); err == nil || !strings.Contains(err.Error(), "Too many") {
		t.Fatalf("expected rejection containing %q, got: %v", "Too many", err)
	}
}

// -----------------------------------------------------------------------
// filter field
// Schema: optional struct with one optional field clusterMoID (minLength=1, maxLength=64).
// No transition rule: filter is mutable.
// -----------------------------------------------------------------------

func TestNetworkInterface_FilterUnset_Admitted(t *testing.T) {
	ensureNamespace(t, nifNamespace)
	obj := nifObj("nif-filter-unset")
	if err := k8sClient.Create(testCtx, obj); err != nil {
		t.Fatalf("expected admission, got: %v", err)
	}
	defer func() { _ = k8sClient.Delete(testCtx, obj) }()
}

func TestNetworkInterface_FilterClusterMoIDSet_Admitted(t *testing.T) {
	ensureNamespace(t, nifNamespace)
	obj := nifObj("nif-filter-cluster-set")
	obj.Spec.Filter = &netv1alpha1.NetworkInterfaceFilter{ClusterMoID: "domain-c1"}
	if err := k8sClient.Create(testCtx, obj); err != nil {
		t.Fatalf("expected admission, got: %v", err)
	}
	defer func() { _ = k8sClient.Delete(testCtx, obj) }()
}

func TestNetworkInterface_FilterClusterMoIDChanged_Admitted(t *testing.T) {
	ensureNamespace(t, nifNamespace)
	obj := nifObj("nif-filter-cluster-changed")
	obj.Spec.Filter = &netv1alpha1.NetworkInterfaceFilter{ClusterMoID: "domain-c1"}
	if err := k8sClient.Create(testCtx, obj); err != nil {
		t.Fatalf("create: %v", err)
	}
	defer func() { _ = k8sClient.Delete(testCtx, obj) }()

	latest := &netv1alpha1.NetworkInterface{}
	if err := k8sClient.Get(testCtx, client.ObjectKeyFromObject(obj), latest); err != nil {
		t.Fatalf("get: %v", err)
	}
	latest.Spec.Filter.ClusterMoID = "domain-c2"
	if err := k8sClient.Update(testCtx, latest); err != nil {
		t.Fatalf("expected update admission, got: %v", err)
	}
}

func TestNetworkInterface_FilterClusterMoIDTooLong_Rejected(t *testing.T) {
	ensureNamespace(t, nifNamespace)
	obj := unstrNIF("nif-filter-cluster-too-long", map[string]interface{}{
		"filter": map[string]interface{}{"clusterMoID": strings.Repeat("a", 65)},
	})
	if err := k8sClient.Create(testCtx, obj); !isRejected(err) || !strings.Contains(err.Error(), "Too long") {
		t.Fatalf("expected rejection containing %q, got: %v", "Too long", err)
	}
}

func TestNetworkInterface_FilterClusterMoIDEmpty_Rejected(t *testing.T) {
	ensureNamespace(t, nifNamespace)
	obj := unstrNIF("nif-filter-cluster-empty", map[string]interface{}{
		"filter": map[string]interface{}{"clusterMoID": ""},
	})
	if err := k8sClient.Create(testCtx, obj); !isRejected(err) {
		t.Fatalf("expected rejection for empty clusterMoID, got: %v", err)
	}
}
