// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: Apache-2.0

package cel_test

import (
	"fmt"
	"strings"
	"testing"

	netv1alpha1 "github.com/vmware-tanzu/net-operator-api/api/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func validHAProxyLoadBalancerConfig(name string) *netv1alpha1.HAProxyLoadBalancerConfig {
	return &netv1alpha1.HAProxyLoadBalancerConfig{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Spec: netv1alpha1.HAProxyLoadBalancerConfigSpec{
			EndPointURLs: []string{"https://10.0.0.1:5555/v2"},
			ServerName:   "haproxy-server",
			CredentialSecretRef: netv1alpha1.ClientSecretReference{
				Name:      "haproxy-creds",
				Namespace: "vmware-system-netop",
			},
		},
	}
}

// --- Basic Validity ---

func TestHAProxyLoadBalancerConfig_Valid_Admitted(t *testing.T) {
	obj := validHAProxyLoadBalancerConfig("hac-valid")
	if err := k8sClient.Create(testCtx, obj); err != nil {
		t.Fatalf("expected admission, got: %v", err)
	}
	_ = k8sClient.Delete(testCtx, obj)
}

func TestHAProxyLoadBalancerConfig_MinimalValid_Admitted(t *testing.T) {
	obj := &netv1alpha1.HAProxyLoadBalancerConfig{
		ObjectMeta: metav1.ObjectMeta{Name: "hac-minimal"},
		Spec: netv1alpha1.HAProxyLoadBalancerConfigSpec{
			EndPointURLs: []string{"https://10.0.0.1:5555/v2"},
			CredentialSecretRef: netv1alpha1.ClientSecretReference{
				Name: "haproxy-creds",
			},
		},
	}
	if err := k8sClient.Create(testCtx, obj); err != nil {
		t.Fatalf("expected admission for minimal HAC, got: %v", err)
	}
	_ = k8sClient.Delete(testCtx, obj)
}

func TestHAProxyLoadBalancerConfig_OmittedCredentialSecretRef_Admitted(t *testing.T) {
	u := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "netoperator.vmware.com/v1alpha1",
			"kind":       "HAProxyLoadBalancerConfig",
			"metadata": map[string]interface{}{
				"name": "hac-omitted-creds",
			},
			"spec": map[string]interface{}{
				"endPointURLs": []interface{}{"https://10.0.0.1:5555/v2"},
			},
		},
	}
	if err := k8sClient.Create(testCtx, u); err != nil {
		t.Fatalf("expected admission for HAC with omitted credentialSecretRef, got: %v", err)
	}
	_ = k8sClient.Delete(testCtx, u)
}

// --- EndPointURLs & CredentialSecretRef Validation ---

func TestHAProxyLoadBalancerConfig_EmptyEndPointURLs_Rejected(t *testing.T) {
	obj := validHAProxyLoadBalancerConfig("hac-empty-endpoints")
	obj.Spec.EndPointURLs = []string{}
	if err := k8sClient.Create(testCtx, obj); !isRejected(err) {
		t.Fatalf("expected rejection for empty endPointURLs (MinItems=1), got: %v", err)
	}
}

func TestHAProxyLoadBalancerConfig_EmptyStringEndPointURL_Rejected(t *testing.T) {
	obj := validHAProxyLoadBalancerConfig("hac-empty-endpoint-str")
	obj.Spec.EndPointURLs = []string{""}
	if err := k8sClient.Create(testCtx, obj); !isRejected(err) {
		t.Fatalf("expected rejection for empty string in endPointURLs, got: %v", err)
	}
}

func TestHAProxyLoadBalancerConfig_MultipleEndPointURLs_Admitted(t *testing.T) {
	obj := validHAProxyLoadBalancerConfig("hac-multi-endpoints")
	obj.Spec.EndPointURLs = []string{
		"https://10.0.0.1:5555/v2",
		"https://10.0.0.2:5555/v2",
	}
	if err := k8sClient.Create(testCtx, obj); err != nil {
		t.Fatalf("expected admission for multiple endpoints, got: %v", err)
	}
	_ = k8sClient.Delete(testCtx, obj)
}

// --- VirtualServerIPPools & VirtualServerIPRanges Validation ---

func TestHAProxyLoadBalancerConfig_ValidVirtualServerIPPools_Admitted(t *testing.T) {
	obj := validHAProxyLoadBalancerConfig("hac-valid-pools")
	obj.Spec.VirtualServerIPPools = []netv1alpha1.IPPoolReference{
		{Name: "pool-1"},
		{Name: "pool-2"},
	}
	if err := k8sClient.Create(testCtx, obj); err != nil {
		t.Fatalf("expected admission for valid virtualServerIPPools, got: %v", err)
	}
	_ = k8sClient.Delete(testCtx, obj)
}

func TestHAProxyLoadBalancerConfig_EmptyPoolName_Rejected(t *testing.T) {
	obj := validHAProxyLoadBalancerConfig("hac-empty-pool-name")
	obj.Spec.VirtualServerIPPools = []netv1alpha1.IPPoolReference{
		{Name: ""},
	}
	if err := k8sClient.Create(testCtx, obj); !isRejected(err) {
		t.Fatalf("expected rejection for empty pool name, got: %v", err)
	}
}

func TestHAProxyLoadBalancerConfig_TooManyVirtualServerIPPools_Rejected(t *testing.T) {
	obj := validHAProxyLoadBalancerConfig("hac-pools-overflow")
	pools := make([]netv1alpha1.IPPoolReference, 257)
	for i := range pools {
		pools[i] = netv1alpha1.IPPoolReference{Name: fmt.Sprintf("pool-%d", i)}
	}
	obj.Spec.VirtualServerIPPools = pools
	if err := k8sClient.Create(testCtx, obj); !isRejected(err) {
		t.Fatalf("expected rejection for >256 virtualServerIPPools, got: %v", err)
	}
}

func TestHAProxyLoadBalancerConfig_MaxVirtualServerIPPools_Admitted(t *testing.T) {
	obj := validHAProxyLoadBalancerConfig("hac-pools-max")
	pools := make([]netv1alpha1.IPPoolReference, 256)
	for i := range pools {
		pools[i] = netv1alpha1.IPPoolReference{Name: fmt.Sprintf("pool-%d", i)}
	}
	obj.Spec.VirtualServerIPPools = pools
	if err := k8sClient.Create(testCtx, obj); err != nil {
		t.Fatalf("expected admission for 256 virtualServerIPPools, got: %v", err)
	}
	_ = k8sClient.Delete(testCtx, obj)
}

func TestHAProxyLoadBalancerConfig_ValidVirtualServerIPRanges_Admitted(t *testing.T) {
	obj := validHAProxyLoadBalancerConfig("hac-valid-ranges")
	obj.Spec.VirtualServerIPRanges = []netv1alpha1.IPRange{
		{StartingAddress: "10.0.0.1", AddressCount: 64},
		{StartingAddress: "fd00::1", AddressCount: 128},
	}
	if err := k8sClient.Create(testCtx, obj); err != nil {
		t.Fatalf("expected admission for valid virtualServerIPRanges, got: %v", err)
	}
	_ = k8sClient.Delete(testCtx, obj)
}

func TestHAProxyLoadBalancerConfig_InvalidStartingAddress_Rejected(t *testing.T) {
	obj := validHAProxyLoadBalancerConfig("hac-range-bad-ip")
	obj.Spec.VirtualServerIPRanges = []netv1alpha1.IPRange{
		{StartingAddress: "not-an-ip", AddressCount: 8},
	}
	if err := k8sClient.Create(testCtx, obj); !isRejected(err) {
		t.Fatalf("expected rejection for invalid startingAddress, got: %v", err)
	}
}

func TestHAProxyLoadBalancerConfig_AddressCountZero_Rejected(t *testing.T) {
	obj := validHAProxyLoadBalancerConfig("hac-range-zero-count")
	obj.Spec.VirtualServerIPRanges = []netv1alpha1.IPRange{
		{StartingAddress: "10.0.0.1", AddressCount: 0},
	}
	if err := k8sClient.Create(testCtx, obj); !isRejected(err) {
		t.Fatalf("expected rejection for addressCount 0, got: %v", err)
	}
}

func TestHAProxyLoadBalancerConfig_TooManyVirtualServerIPRanges_Rejected(t *testing.T) {
	obj := validHAProxyLoadBalancerConfig("hac-ranges-overflow")
	ranges := make([]netv1alpha1.IPRange, 257)
	for i := range ranges {
		ranges[i] = netv1alpha1.IPRange{StartingAddress: fmt.Sprintf("10.0.%d.%d", i/256, i%256), AddressCount: 1}
	}
	obj.Spec.VirtualServerIPRanges = ranges
	if err := k8sClient.Create(testCtx, obj); !isRejected(err) {
		t.Fatalf("expected rejection for >256 virtualServerIPRanges, got: %v", err)
	}
}

func TestHAProxyLoadBalancerConfig_MaxVirtualServerIPRanges_Admitted(t *testing.T) {
	obj := validHAProxyLoadBalancerConfig("hac-ranges-max")
	ranges := make([]netv1alpha1.IPRange, 256)
	for i := range ranges {
		ranges[i] = netv1alpha1.IPRange{StartingAddress: fmt.Sprintf("10.0.0.%d", i), AddressCount: 1}
	}
	obj.Spec.VirtualServerIPRanges = ranges
	if err := k8sClient.Create(testCtx, obj); err != nil {
		t.Fatalf("expected admission for 256 virtualServerIPRanges, got: %v", err)
	}
	_ = k8sClient.Delete(testCtx, obj)
}

func TestHAProxyLoadBalancerConfig_BothIPPoolsAndIPRanges_Admitted(t *testing.T) {
	obj := validHAProxyLoadBalancerConfig("hac-pools-and-ranges")
	obj.Spec.VirtualServerIPPools = []netv1alpha1.IPPoolReference{
		{Name: "pool-explicit"},
	}
	obj.Spec.VirtualServerIPRanges = []netv1alpha1.IPRange{
		{StartingAddress: "10.0.0.1", AddressCount: 32},
	}
	if err := k8sClient.Create(testCtx, obj); err != nil {
		t.Fatalf("expected admission when both pools and ranges provided, got: %v", err)
	}
	_ = k8sClient.Delete(testCtx, obj)
}

func TestHAProxyLoadBalancerConfig_NeitherIPPoolsNorIPRanges_Admitted(t *testing.T) {
	obj := validHAProxyLoadBalancerConfig("hac-no-pools-no-ranges")
	obj.Spec.VirtualServerIPPools = nil
	obj.Spec.VirtualServerIPRanges = nil
	if err := k8sClient.Create(testCtx, obj); err != nil {
		t.Fatalf("expected admission when neither pools nor ranges provided, got: %v", err)
	}
	_ = k8sClient.Delete(testCtx, obj)
}

func TestHAProxyLoadBalancerConfig_CertificateAuthorityData_Admitted(t *testing.T) {
	obj := validHAProxyLoadBalancerConfig("hac-ca-data")
	obj.Spec.CertificateAuthorityData = "-----BEGIN CERTIFICATE-----\nMIIB...\n-----END CERTIFICATE-----"
	if err := k8sClient.Create(testCtx, obj); err != nil {
		t.Fatalf("expected admission with certificateAuthorityData, got: %v", err)
	}
	_ = k8sClient.Delete(testCtx, obj)
}

func TestHAProxyLoadBalancerConfig_CertificateAuthorityData_TooLong_Rejected(t *testing.T) {
	obj := validHAProxyLoadBalancerConfig("hac-ca-too-long")
	obj.Spec.CertificateAuthorityData = strings.Repeat("a", 65537)
	if err := k8sClient.Create(testCtx, obj); !isRejected(err) {
		t.Fatalf("expected rejection for certificateAuthorityData >65536 bytes, got: %v", err)
	}
}

func TestHAProxyLoadBalancerConfig_CertificateAuthorityData_EmptyString_Rejected(t *testing.T) {
	u := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "netoperator.vmware.com/v1alpha1",
			"kind":       "HAProxyLoadBalancerConfig",
			"metadata": map[string]interface{}{
				"name": "hac-ca-empty-str",
			},
			"spec": map[string]interface{}{
				"endPointURLs":             []interface{}{"https://10.0.0.1:5555/v2"},
				"certificateAuthorityData": "",
			},
		},
	}
	if err := k8sClient.Create(testCtx, u); !isRejected(err) {
		t.Fatalf("expected rejection for empty certificateAuthorityData string, got: %v", err)
	}
}

func TestHAProxyLoadBalancerConfig_UnstructuredEmptyEndPointURL_Rejected(t *testing.T) {
	u := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "netoperator.vmware.com/v1alpha1",
			"kind":       "HAProxyLoadBalancerConfig",
			"metadata": map[string]interface{}{
				"name": "hac-unstructured-empty-endpoint",
			},
			"spec": map[string]interface{}{
				"endPointURLs": []interface{}{""},
			},
		},
	}
	if err := k8sClient.Create(testCtx, u); !isRejected(err) {
		t.Fatalf("expected rejection for unstructured empty endpoint string, got: %v", err)
		_ = k8sClient.Delete(testCtx, u)
	}
}

// --- EffectiveVirtualServerIPPools Status Field ---

func TestHAProxyLoadBalancerConfig_EffectiveVirtualServerIPPools_StatusRoundtrip(t *testing.T) {
	obj := validHAProxyLoadBalancerConfig("hac-effective-pools")
	if err := k8sClient.Create(testCtx, obj); err != nil {
		t.Fatalf("create: %v", err)
	}
	defer func() { _ = k8sClient.Delete(testCtx, obj) }()

	latest := &netv1alpha1.HAProxyLoadBalancerConfig{}
	if err := k8sClient.Get(testCtx, client.ObjectKeyFromObject(obj), latest); err != nil {
		t.Fatalf("get: %v", err)
	}

	latest.Status.EffectiveVirtualServerIPPools = []string{"pool-1", "pool-2", "pool-3"}
	if err := k8sClient.Status().Update(testCtx, latest); err != nil {
		t.Fatalf("status update: %v", err)
	}

	updated := &netv1alpha1.HAProxyLoadBalancerConfig{}
	if err := k8sClient.Get(testCtx, client.ObjectKeyFromObject(obj), updated); err != nil {
		t.Fatalf("get after status update: %v", err)
	}
	if len(updated.Status.EffectiveVirtualServerIPPools) != 3 {
		t.Fatalf("expected 3 effective pools, got %d", len(updated.Status.EffectiveVirtualServerIPPools))
	}
	expected := []string{"pool-1", "pool-2", "pool-3"}
	for i, exp := range expected {
		if updated.Status.EffectiveVirtualServerIPPools[i] != exp {
			t.Errorf("pool[%d]: got %q, want %q", i, updated.Status.EffectiveVirtualServerIPPools[i], exp)
		}
	}
}

func TestHAProxyLoadBalancerConfig_EffectiveVirtualServerIPPools_AbsentByDefault(t *testing.T) {
	obj := validHAProxyLoadBalancerConfig("hac-effective-absent")
	if err := k8sClient.Create(testCtx, obj); err != nil {
		t.Fatalf("create: %v", err)
	}
	defer func() { _ = k8sClient.Delete(testCtx, obj) }()

	got := &netv1alpha1.HAProxyLoadBalancerConfig{}
	if err := k8sClient.Get(testCtx, client.ObjectKeyFromObject(obj), got); err != nil {
		t.Fatalf("get: %v", err)
	}
	if len(got.Status.EffectiveVirtualServerIPPools) != 0 {
		t.Errorf("expected empty effectiveVirtualServerIPPools on fresh HAC, got %v",
			got.Status.EffectiveVirtualServerIPPools)
	}
}

// --- Status Conditions ---

func TestHAProxyLoadBalancerConfig_Conditions_StatusRoundtrip(t *testing.T) {
	obj := validHAProxyLoadBalancerConfig("hac-status-conditions")
	if err := k8sClient.Create(testCtx, obj); err != nil {
		t.Fatalf("create: %v", err)
	}
	defer func() { _ = k8sClient.Delete(testCtx, obj) }()

	latest := &netv1alpha1.HAProxyLoadBalancerConfig{}
	if err := k8sClient.Get(testCtx, client.ObjectKeyFromObject(obj), latest); err != nil {
		t.Fatalf("get: %v", err)
	}

	cond := metav1.Condition{
		Type:               "IPPoolReconciled",
		Status:             metav1.ConditionTrue,
		Reason:             "Reconciled",
		Message:            "IPPools reconciled successfully",
		LastTransitionTime: metav1.Now(),
	}
	latest.SetConditions([]metav1.Condition{cond})
	if err := k8sClient.Status().Update(testCtx, latest); err != nil {
		t.Fatalf("status update with conditions: %v", err)
	}

	updated := &netv1alpha1.HAProxyLoadBalancerConfig{}
	if err := k8sClient.Get(testCtx, client.ObjectKeyFromObject(obj), updated); err != nil {
		t.Fatalf("get after status update: %v", err)
	}
	gotConditions := updated.GetConditions()
	if len(gotConditions) != 1 {
		t.Fatalf("expected 1 condition, got %d", len(gotConditions))
	}
	if gotConditions[0].Type != "IPPoolReconciled" || gotConditions[0].Status != metav1.ConditionTrue {
		t.Errorf("unexpected condition: %+v", gotConditions[0])
	}
}

func TestHAProxyLoadBalancerConfig_Conditions_InvalidStatus_Rejected(t *testing.T) {
	obj := validHAProxyLoadBalancerConfig("hac-bad-condition-status")
	if err := k8sClient.Create(testCtx, obj); err != nil {
		t.Fatalf("create: %v", err)
	}
	defer func() { _ = k8sClient.Delete(testCtx, obj) }()

	latest := &netv1alpha1.HAProxyLoadBalancerConfig{}
	if err := k8sClient.Get(testCtx, client.ObjectKeyFromObject(obj), latest); err != nil {
		t.Fatalf("get: %v", err)
	}

	latest.Status.Conditions = []metav1.Condition{
		{
			Type:               "IPPoolReconciled",
			Status:             "InvalidStatus",
			Reason:             "Reconciled",
			Message:            "testing invalid status",
			LastTransitionTime: metav1.Now(),
		},
	}
	if err := k8sClient.Status().Update(testCtx, latest); err == nil || !strings.Contains(err.Error(), "Unsupported value") {
		t.Fatalf("expected rejection for invalid condition status, got: %v", err)
	}
}

func TestHAProxyLoadBalancerConfig_Conditions_DuplicateTypes_Rejected(t *testing.T) {
	obj := validHAProxyLoadBalancerConfig("hac-dup-conditions")
	if err := k8sClient.Create(testCtx, obj); err != nil {
		t.Fatalf("create: %v", err)
	}
	defer func() { _ = k8sClient.Delete(testCtx, obj) }()

	latest := &netv1alpha1.HAProxyLoadBalancerConfig{}
	if err := k8sClient.Get(testCtx, client.ObjectKeyFromObject(obj), latest); err != nil {
		t.Fatalf("get: %v", err)
	}

	latest.Status.Conditions = []metav1.Condition{
		{
			Type:               "IPPoolReconciled",
			Status:             metav1.ConditionTrue,
			Reason:             "Reconciled",
			Message:            "first",
			LastTransitionTime: metav1.Now(),
		},
		{
			Type:               "IPPoolReconciled",
			Status:             metav1.ConditionFalse,
			Reason:             "Failed",
			Message:            "duplicate",
			LastTransitionTime: metav1.Now(),
		},
	}
	if err := k8sClient.Status().Update(testCtx, latest); !isRejected(err) {
		t.Fatalf("expected rejection for duplicate condition types, got: %v", err)
	}
}

func TestHAProxyLoadBalancerConfig_Conditions_TooMany_Rejected(t *testing.T) {
	obj := validHAProxyLoadBalancerConfig("hac-conditions-overflow")
	if err := k8sClient.Create(testCtx, obj); err != nil {
		t.Fatalf("create: %v", err)
	}
	defer func() { _ = k8sClient.Delete(testCtx, obj) }()

	latest := &netv1alpha1.HAProxyLoadBalancerConfig{}
	if err := k8sClient.Get(testCtx, client.ObjectKeyFromObject(obj), latest); err != nil {
		t.Fatalf("get: %v", err)
	}

	conditions := make([]metav1.Condition, 9)
	for i := range conditions {
		conditions[i] = metav1.Condition{
			Type:               fmt.Sprintf("Cond%d", i),
			Status:             metav1.ConditionTrue,
			Reason:             "Reconciled",
			Message:            "ok",
			LastTransitionTime: metav1.Now(),
		}
	}
	latest.Status.Conditions = conditions
	if err := k8sClient.Status().Update(testCtx, latest); !isRejected(err) {
		t.Fatalf("expected rejection for >8 conditions, got: %v", err)
	}
}

// --- virtualServerIPPools & virtualServerIPRanges additive-only rules ---

func TestHAProxyLoadBalancerConfig_VirtualServerIPPools_RemovalRejected(t *testing.T) {
	obj := validHAProxyLoadBalancerConfig("hac-pool-removal")
	obj.Spec.VirtualServerIPPools = []netv1alpha1.IPPoolReference{
		{Name: "pool-1"},
		{Name: "pool-2"},
	}
	if err := k8sClient.Create(testCtx, obj); err != nil {
		t.Fatalf("create: %v", err)
	}
	defer func() { _ = k8sClient.Delete(testCtx, obj) }()

	latest := &netv1alpha1.HAProxyLoadBalancerConfig{}
	if err := k8sClient.Get(testCtx, client.ObjectKeyFromObject(obj), latest); err != nil {
		t.Fatalf("get: %v", err)
	}
	latest.Spec.VirtualServerIPPools = []netv1alpha1.IPPoolReference{
		{Name: "pool-2"},
	}
	if err := k8sClient.Update(testCtx, latest); !isRejected(err) {
		t.Fatalf("expected rejection when removing pool-1 from virtualServerIPPools, got: %v", err)
	}
}

func TestHAProxyLoadBalancerConfig_VirtualServerIPPools_AppendAdmitted(t *testing.T) {
	obj := validHAProxyLoadBalancerConfig("hac-pool-append")
	obj.Spec.VirtualServerIPPools = []netv1alpha1.IPPoolReference{
		{Name: "pool-1"},
	}
	if err := k8sClient.Create(testCtx, obj); err != nil {
		t.Fatalf("create: %v", err)
	}
	defer func() { _ = k8sClient.Delete(testCtx, obj) }()

	latest := &netv1alpha1.HAProxyLoadBalancerConfig{}
	if err := k8sClient.Get(testCtx, client.ObjectKeyFromObject(obj), latest); err != nil {
		t.Fatalf("get: %v", err)
	}
	latest.Spec.VirtualServerIPPools = []netv1alpha1.IPPoolReference{
		{Name: "pool-1"},
		{Name: "pool-2"},
	}
	if err := k8sClient.Update(testCtx, latest); err != nil {
		t.Fatalf("expected admission when appending to virtualServerIPPools, got: %v", err)
	}
}

func TestHAProxyLoadBalancerConfig_VirtualServerIPRanges_RemovalRejected(t *testing.T) {
	obj := validHAProxyLoadBalancerConfig("hac-range-removal")
	obj.Spec.VirtualServerIPRanges = []netv1alpha1.IPRange{
		{StartingAddress: "10.1.0.1", AddressCount: 4},
		{StartingAddress: "10.1.1.1", AddressCount: 4},
	}
	if err := k8sClient.Create(testCtx, obj); err != nil {
		t.Fatalf("create: %v", err)
	}
	defer func() { _ = k8sClient.Delete(testCtx, obj) }()

	latest := &netv1alpha1.HAProxyLoadBalancerConfig{}
	if err := k8sClient.Get(testCtx, client.ObjectKeyFromObject(obj), latest); err != nil {
		t.Fatalf("get: %v", err)
	}
	latest.Spec.VirtualServerIPRanges = []netv1alpha1.IPRange{
		{StartingAddress: "10.1.0.1", AddressCount: 4},
	}
	if err := k8sClient.Update(testCtx, latest); !isRejected(err) {
		t.Fatalf("expected rejection when removing 10.1.1.1 from virtualServerIPRanges, got: %v", err)
	}
}

func TestHAProxyLoadBalancerConfig_VirtualServerIPRanges_AppendAdmitted(t *testing.T) {
	obj := validHAProxyLoadBalancerConfig("hac-range-append")
	obj.Spec.VirtualServerIPRanges = []netv1alpha1.IPRange{
		{StartingAddress: "10.2.0.1", AddressCount: 4},
	}
	if err := k8sClient.Create(testCtx, obj); err != nil {
		t.Fatalf("create: %v", err)
	}
	defer func() { _ = k8sClient.Delete(testCtx, obj) }()

	latest := &netv1alpha1.HAProxyLoadBalancerConfig{}
	if err := k8sClient.Get(testCtx, client.ObjectKeyFromObject(obj), latest); err != nil {
		t.Fatalf("get: %v", err)
	}
	latest.Spec.VirtualServerIPRanges = []netv1alpha1.IPRange{
		{StartingAddress: "10.2.0.1", AddressCount: 4},
		{StartingAddress: "10.2.1.1", AddressCount: 4},
	}
	if err := k8sClient.Update(testCtx, latest); err != nil {
		t.Fatalf("expected admission when appending to virtualServerIPRanges, got: %v", err)
	}
}

