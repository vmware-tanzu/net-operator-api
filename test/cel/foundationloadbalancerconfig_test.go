// Copyright (c) 2026 Broadcom. All Rights Reserved.
// Broadcom Confidential. The term "Broadcom" refers to Broadcom Inc.
// and/or its subsidiaries.

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

const flbcNS = "flbc-cel-test"

func validFLBC(name string) *netv1alpha1.FoundationLoadBalancerConfig {
	return &netv1alpha1.FoundationLoadBalancerConfig{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: flbcNS,
		},
		Spec: netv1alpha1.FoundationLoadBalancerConfigSpec{
			DeploymentSpec: netv1alpha1.FoundationLoadBalancerDeploymentSpec{
				Size:             netv1alpha1.FoundationLoadBalancerSizeSmall,
				StoragePolicy:    "default",
				AvailabilityMode: netv1alpha1.FoundationAvailabilityModeActivePassive,
				Zones:            []string{"zone-a"},
				ActivePassiveAvailabilityMode: &netv1alpha1.ActivePassiveAvailabilityMode{
					Replicas: 2,
				},
			},
			VirtualIPNetwork: netv1alpha1.NetworkReference{Name: "vip-net"},
			NetworkSpec: netv1alpha1.FoundationLoadBalancerNetworkConfigSpec{
				VirtualServerIPPools: []netv1alpha1.IPPoolReference{
					{Name: "pool-1"},
				},
			},
		},
	}
}

// --- availability mode, zones, DNS/NTP, storage policy ---

func TestFoundationLoadBalancerConfig_InvalidAvailabilityMode_Rejected(t *testing.T) {
	ensureNamespace(t, flbcNS)
	obj := validFLBC("flbc-bad-mode")
	obj.Spec.DeploymentSpec.AvailabilityMode = "invalid-mode"
	if err := k8sClient.Create(testCtx, obj); !isRejected(err) {
		t.Fatalf("expected rejection for invalid availabilityMode, got: %v", err)
	}
}

func TestFoundationLoadBalancerConfig_ValidActivePassive_Admitted(t *testing.T) {
	ensureNamespace(t, flbcNS)
	obj := validFLBC("flbc-valid-ap")
	if err := k8sClient.Create(testCtx, obj); err != nil {
		t.Fatalf("expected admission, got: %v", err)
	}
	_ = k8sClient.Delete(testCtx, obj)
}

func TestFoundationLoadBalancerConfig_ValidSingleNode_Admitted(t *testing.T) {
	ensureNamespace(t, flbcNS)
	obj := validFLBC("flbc-valid-sn")
	obj.Spec.DeploymentSpec.AvailabilityMode = netv1alpha1.FoundationAvailabilityModeSingleNode
	obj.Spec.DeploymentSpec.ActivePassiveAvailabilityMode = nil
	obj.Spec.DeploymentSpec.SingleNodeAvailabilityMode = &netv1alpha1.SingleNodeAvailabilityMode{
		Replicas: 1,
	}
	if err := k8sClient.Create(testCtx, obj); err != nil {
		t.Fatalf("expected admission, got: %v", err)
	}
	_ = k8sClient.Delete(testCtx, obj)
}

func TestFoundationLoadBalancerConfig_DowngradeAvailabilityMode_Rejected(t *testing.T) {
	ensureNamespace(t, flbcNS)
	obj := validFLBC("flbc-downgrade")
	if err := k8sClient.Create(testCtx, obj); err != nil {
		t.Fatalf("create: %v", err)
	}
	defer func() { _ = k8sClient.Delete(testCtx, obj) }()

	latest := &netv1alpha1.FoundationLoadBalancerConfig{}
	if err := k8sClient.Get(testCtx, client.ObjectKeyFromObject(obj), latest); err != nil {
		t.Fatalf("get: %v", err)
	}
	latest.Spec.DeploymentSpec.AvailabilityMode = netv1alpha1.FoundationAvailabilityModeSingleNode
	latest.Spec.DeploymentSpec.ActivePassiveAvailabilityMode = nil
	latest.Spec.DeploymentSpec.SingleNodeAvailabilityMode = &netv1alpha1.SingleNodeAvailabilityMode{Replicas: 1}

	if err := k8sClient.Update(testCtx, latest); !isRejected(err) {
		t.Fatalf("expected rejection for downgrade active-passive→single-node, got: %v", err)
	}
}

func TestFoundationLoadBalancerConfig_BothSingleNodeAndActivePassive_Rejected(t *testing.T) {
	ensureNamespace(t, flbcNS)
	obj := validFLBC("flbc-both-specs")
	obj.Spec.DeploymentSpec.SingleNodeAvailabilityMode = &netv1alpha1.SingleNodeAvailabilityMode{Replicas: 1}
	if err := k8sClient.Create(testCtx, obj); !isRejected(err) {
		t.Fatalf("expected rejection for both singleNodeSpec and activePassiveSpec, got: %v", err)
	}
}

// makeZones returns a slice of n zone names, each the single character "z".
func makeZones(n int) []string {
	z := make([]string, n)
	for i := range z {
		z[i] = "z"
	}
	return z
}

func TestFoundationLoadBalancerConfig_TooManyZones_Rejected(t *testing.T) {
	ensureNamespace(t, flbcNS)
	obj := validFLBC("flbc-zones-overflow")
	obj.Spec.DeploymentSpec.Zones = makeZones(257)
	if err := k8sClient.Create(testCtx, obj); !isRejected(err) {
		t.Fatalf("expected rejection for >256 zones, got: %v", err)
	}
}

func TestFoundationLoadBalancerConfig_ZonesAtMaxCount_Admitted(t *testing.T) {
	ensureNamespace(t, flbcNS)
	obj := validFLBC("flbc-zones-maxcount")
	obj.Spec.DeploymentSpec.Zones = makeZones(256)
	if err := k8sClient.Create(testCtx, obj); err != nil {
		t.Fatalf("expected admission for 256 zones, got: %v", err)
	}
	_ = k8sClient.Delete(testCtx, obj)
}

// TestFoundationLoadBalancerConfig_LongZoneName_Rejected verifies that zone names longer than 253
// characters are rejected. Per-item length validation is deferred to the FLBC webhook.
func TestFoundationLoadBalancerConfig_LongZoneName_Rejected(t *testing.T) {
	ensureNamespace(t, flbcNS)
	obj := validFLBC("flbc-zone-long")
	obj.Spec.DeploymentSpec.Zones = []string{strings.Repeat("z", 254)}
	if err := k8sClient.Create(testCtx, obj); !isRejected(err) {
		t.Fatalf("expected rejection for long zone name (no items:MaxLength in CRD), got: %v", err)
	}
	_ = k8sClient.Delete(testCtx, obj)
}

func TestFoundationLoadBalancerConfig_NoZones_Admitted(t *testing.T) {
	ensureNamespace(t, flbcNS)
	obj := validFLBC("flbc-no-zones")
	obj.Spec.DeploymentSpec.Zones = nil
	if err := k8sClient.Create(testCtx, obj); err != nil {
		t.Fatalf("expected admission without zones, got: %v", err)
	}
	_ = k8sClient.Delete(testCtx, obj)
}

func TestFoundationLoadBalancerConfig_ExplicitZones_Admitted(t *testing.T) {
	ensureNamespace(t, flbcNS)
	obj := validFLBC("flbc-explicit-zones")
	obj.Spec.DeploymentSpec.Zones = []string{"zone-a", "zone-b"}
	if err := k8sClient.Create(testCtx, obj); err != nil {
		t.Fatalf("expected admission with explicit zones, got: %v", err)
	}
	_ = k8sClient.Delete(testCtx, obj)
}

// TestFoundationLoadBalancerConfig_EmptyDNSServer_Admitted verifies that an empty string item in
// dnsServers is admitted.  The +kubebuilder:validation:items:MinLength=1 marker is present in the
// Go types but is silently ignored by controller-gen v0.14 — no minLength constraint appears on
// dnsServers items in the generated CRD.  Per-item empty-string rejection is deferred to the FLBC
// validating webhook (#20).
func TestFoundationLoadBalancerConfig_EmptyDNSServer_Admitted(t *testing.T) {
	ensureNamespace(t, flbcNS)
	obj := makeFlbcUnstructured("flbc-empty-dns", map[string]interface{}{
		"dnsServers": []interface{}{""},
	})
	if err := k8sClient.Create(testCtx, obj); err != nil {
		t.Fatalf("expected admission for empty DNS server (no items:minLength in CRD), got: %v", err)
	}
	_ = k8sClient.Delete(testCtx, obj)
}

// TestFoundationLoadBalancerConfig_NonEmptyDNSServer_Admitted verifies that a non-empty DNS entry
// is admitted — format validation (valid IP) is deferred to the FLBC webhook (#20).
func TestFoundationLoadBalancerConfig_NonEmptyDNSServer_Admitted(t *testing.T) {
	ensureNamespace(t, flbcNS)
	obj := validFLBC("flbc-ipv6-dns")
	obj.Spec.NetworkSpec.DNSServers = []string{"2001:db8::1"}
	if err := k8sClient.Create(testCtx, obj); err != nil {
		t.Fatalf("expected admission for non-IPv4 DNS server, got: %v", err)
	}
	_ = k8sClient.Delete(testCtx, obj)
}

// TestFoundationLoadBalancerConfig_EmptyVIPSubnet_Admitted verifies that an empty string item in
// virtualServerSubnets is admitted.  Same controller-gen v0.14 items: marker limitation as DNS/NTP
// — no minLength on subnet items in the generated CRD.
func TestFoundationLoadBalancerConfig_EmptyVIPSubnet_Admitted(t *testing.T) {
	ensureNamespace(t, flbcNS)
	obj := makeFlbcUnstructured("flbc-empty-cidr", map[string]interface{}{
		"virtualServerSubnets": []interface{}{""},
	})
	if err := k8sClient.Create(testCtx, obj); err != nil {
		t.Fatalf("expected admission for empty VIP subnet (no items:minLength in CRD), got: %v", err)
	}
	_ = k8sClient.Delete(testCtx, obj)
}

// TestFoundationLoadBalancerConfig_EmptyNTPServer_Admitted verifies that an empty string item in
// ntpServers is admitted for the same reason as dnsServers — items:MinLength=1 is ignored by
// controller-gen v0.14.
func TestFoundationLoadBalancerConfig_EmptyNTPServer_Admitted(t *testing.T) {
	ensureNamespace(t, flbcNS)
	obj := makeFlbcUnstructured("flbc-empty-ntp", map[string]interface{}{
		"ntpServers": []interface{}{""},
	})
	if err := k8sClient.Create(testCtx, obj); err != nil {
		t.Fatalf("expected admission for empty NTP server (no items:minLength in CRD), got: %v", err)
	}
	_ = k8sClient.Delete(testCtx, obj)
}

// TestFoundationLoadBalancerConfig_NonEmptyNTPServer_Admitted verifies a valid NTP hostname is
// admitted.
func TestFoundationLoadBalancerConfig_NonEmptyNTPServer_Admitted(t *testing.T) {
	ensureNamespace(t, flbcNS)
	obj := validFLBC("flbc-valid-ntp")
	obj.Spec.NetworkSpec.NTPServers = []string{"ntp.example.com"}
	if err := k8sClient.Create(testCtx, obj); err != nil {
		t.Fatalf("expected admission for non-empty NTP server, got: %v", err)
	}
	_ = k8sClient.Delete(testCtx, obj)
}

// TestFoundationLoadBalancerConfig_StoragePolicyZeroValue_Admitted verifies that a typed Go client
// that sets StoragePolicy to "" is admitted: omitempty drops the field before it reaches the wire,
// so MinLength=1 is never evaluated.
func TestFoundationLoadBalancerConfig_StoragePolicyZeroValue_Admitted(t *testing.T) {
	ensureNamespace(t, flbcNS)
	obj := validFLBC("flbc-storage-zero")
	obj.Spec.DeploymentSpec.StoragePolicy = ""
	if err := k8sClient.Create(testCtx, obj); err != nil {
		t.Fatalf("expected admission for zero-value StoragePolicy (omitted by omitempty), got: %v", err)
	}
	_ = k8sClient.Delete(testCtx, obj)
}

// TestFoundationLoadBalancerConfig_EmptyStoragePolicy_Rejected_ViaRawClient verifies that an
// unstructured client that explicitly sends "storagePolicy": "" is rejected by MinLength=1.
// This is the only path that can reach the constraint — the typed client can never produce this
// payload because omitempty drops "" before serialization.
func TestFoundationLoadBalancerConfig_EmptyStoragePolicy_Rejected_ViaRawClient(t *testing.T) {
	ensureNamespace(t, flbcNS)
	obj := makeFlbcUnstructured("flbc-empty-storage", nil)
	// storagePolicy lives in deploymentSpec, not networkSpec — set it directly.
	spec := obj.Object["spec"].(map[string]interface{})
	deploymentSpec := spec["deploymentSpec"].(map[string]interface{})
	deploymentSpec["storagePolicy"] = ""
	if err := k8sClient.Create(testCtx, obj); !isRejected(err) {
		t.Fatalf("expected rejection for explicit empty storagePolicy (MinLength=1), got: %v", err)
		_ = k8sClient.Delete(testCtx, obj)
	}
}

// --- VirtualServerIPRanges + relaxed VirtualServerIPPools ---

// TestFoundationLoadBalancerConfig_VirtualServerIPRangesOnly_Admitted verifies that an FLBC with
// only virtualServerIPRanges set (no virtualServerIPPools) satisfies the at-least-one XValidation.
func TestFoundationLoadBalancerConfig_VirtualServerIPRangesOnly_Admitted(t *testing.T) {
	ensureNamespace(t, flbcNS)
	obj := validFLBC("flbc-ranges-only")
	obj.Spec.NetworkSpec.VirtualServerIPPools = nil
	obj.Spec.NetworkSpec.VirtualServerIPRanges = []netv1alpha1.IPRange{
		{StartingAddress: "10.0.0.1", AddressCount: 64},
	}
	if err := k8sClient.Create(testCtx, obj); err != nil {
		t.Fatalf("expected admission with only virtualServerIPRanges, got: %v", err)
	}
	_ = k8sClient.Delete(testCtx, obj)
}

// TestFoundationLoadBalancerConfig_BothIPRangesAndIPPools_Rejected verifies that providing both
// fields together is rejected by the mutual-exclusivity XValidation rule.
func TestFoundationLoadBalancerConfig_BothIPRangesAndIPPools_Rejected(t *testing.T) {
	ensureNamespace(t, flbcNS)
	obj := validFLBC("flbc-ranges-and-pools")
	obj.Spec.NetworkSpec.VirtualServerIPRanges = []netv1alpha1.IPRange{
		{StartingAddress: "10.0.1.0", AddressCount: 32},
	}
	if err := k8sClient.Create(testCtx, obj); !isRejected(err) {
		t.Fatalf("expected rejection when both virtualServerIPPools and virtualServerIPRanges set, got: %v", err)
		_ = k8sClient.Delete(testCtx, obj)
	}
}

// TestFoundationLoadBalancerConfig_NeitherIPRangesNorIPPools_Rejected verifies that the at-least-one
// XValidation rejects an FLBC where both virtualServerIPPools and virtualServerIPRanges are absent.
func TestFoundationLoadBalancerConfig_NeitherIPRangesNorIPPools_Rejected(t *testing.T) {
	ensureNamespace(t, flbcNS)
	obj := validFLBC("flbc-no-pools-no-ranges")
	obj.Spec.NetworkSpec.VirtualServerIPPools = nil
	obj.Spec.NetworkSpec.VirtualServerIPRanges = nil
	if err := k8sClient.Create(testCtx, obj); !isRejected(err) {
		t.Fatalf("expected rejection when both virtualServerIPPools and virtualServerIPRanges absent, got: %v", err)
		_ = k8sClient.Delete(testCtx, obj)
	}
}

// TestFoundationLoadBalancerConfig_EmptyVirtualServerIPPools_Rejected verifies that setting
// virtualServerIPPools to an empty slice (with no ranges) triggers the at-least-one XValidation.
// The at-least-one rule replaced the old required+minItems:1 constraint.
func TestFoundationLoadBalancerConfig_EmptyVirtualServerIPPools_Rejected(t *testing.T) {
	ensureNamespace(t, flbcNS)
	obj := validFLBC("flbc-empty-pools")
	obj.Spec.NetworkSpec.VirtualServerIPPools = []netv1alpha1.IPPoolReference{}
	if err := k8sClient.Create(testCtx, obj); !isRejected(err) {
		t.Fatalf("expected rejection when virtualServerIPPools empty and ranges absent, got: %v", err)
	}
}

// TestFoundationLoadBalancerConfig_BothIPRangesAndIPPools_UpdateRejected verifies that adding
// virtualServerIPRanges to an existing FLBC that already has virtualServerIPPools is rejected.
func TestFoundationLoadBalancerConfig_BothIPRangesAndIPPools_UpdateRejected(t *testing.T) {
	ensureNamespace(t, flbcNS)
	obj := validFLBC("flbc-ranges-and-pools-update")
	if err := k8sClient.Create(testCtx, obj); err != nil {
		t.Fatalf("create: %v", err)
	}
	defer func() { _ = k8sClient.Delete(testCtx, obj) }()

	latest := &netv1alpha1.FoundationLoadBalancerConfig{}
	if err := k8sClient.Get(testCtx, client.ObjectKeyFromObject(obj), latest); err != nil {
		t.Fatalf("get: %v", err)
	}
	latest.Spec.NetworkSpec.VirtualServerIPRanges = []netv1alpha1.IPRange{
		{StartingAddress: "10.0.1.0", AddressCount: 32},
	}
	if err := k8sClient.Update(testCtx, latest); !isRejected(err) {
		t.Fatalf("expected rejection when adding virtualServerIPRanges to an FLBC with existing virtualServerIPPools, got: %v", err)
	}
}

// TestFoundationLoadBalancerConfig_IPRangeAddressCountZero_Rejected verifies that addressCount: 0
// is rejected by the Minimum=1 constraint on IPRange.AddressCount.
func TestFoundationLoadBalancerConfig_IPRangeAddressCountZero_Rejected(t *testing.T) {
	ensureNamespace(t, flbcNS)
	obj := validFLBC("flbc-range-zero-count")
	obj.Spec.NetworkSpec.VirtualServerIPPools = nil
	obj.Spec.NetworkSpec.VirtualServerIPRanges = []netv1alpha1.IPRange{
		{StartingAddress: "10.0.0.1", AddressCount: 0},
	}
	if err := k8sClient.Create(testCtx, obj); !isRejected(err) {
		t.Fatalf("expected rejection for addressCount 0, got: %v", err)
	}
}

// TestFoundationLoadBalancerConfig_IPRangeValidAddressCount_Admitted verifies that addressCount: 1
// (minimum boundary) is admitted.
func TestFoundationLoadBalancerConfig_IPRangeValidAddressCount_Admitted(t *testing.T) {
	ensureNamespace(t, flbcNS)
	obj := validFLBC("flbc-range-count-one")
	obj.Spec.NetworkSpec.VirtualServerIPPools = nil
	obj.Spec.NetworkSpec.VirtualServerIPRanges = []netv1alpha1.IPRange{
		{StartingAddress: "10.0.0.1", AddressCount: 1},
	}
	if err := k8sClient.Create(testCtx, obj); err != nil {
		t.Fatalf("expected admission for addressCount 1, got: %v", err)
	}
	_ = k8sClient.Delete(testCtx, obj)
}

// TestFoundationLoadBalancerConfig_IPRangeInvalidStartingAddress_Rejected verifies that a
// startingAddress that is not a valid IPv4 address is rejected by Format=ipv4.
func TestFoundationLoadBalancerConfig_IPRangeInvalidStartingAddress_Rejected(t *testing.T) {
	ensureNamespace(t, flbcNS)
	obj := validFLBC("flbc-range-bad-ip")
	obj.Spec.NetworkSpec.VirtualServerIPPools = nil
	obj.Spec.NetworkSpec.VirtualServerIPRanges = []netv1alpha1.IPRange{
		{StartingAddress: "not-an-ip", AddressCount: 8},
	}
	if err := k8sClient.Create(testCtx, obj); !isRejected(err) {
		t.Fatalf("expected rejection for non-IPv4 startingAddress, got: %v", err)
	}
}

// TestFoundationLoadBalancerConfig_IPRangeTooManyRanges_Rejected verifies that providing more
// than 256 ranges is rejected by MaxItems=256.
func TestFoundationLoadBalancerConfig_IPRangeTooManyRanges_Rejected(t *testing.T) {
	ensureNamespace(t, flbcNS)
	obj := validFLBC("flbc-ranges-overflow")
	obj.Spec.NetworkSpec.VirtualServerIPPools = nil
	ranges := make([]netv1alpha1.IPRange, 257)
	for i := range ranges {
		// 10.0.0.0–10.0.0.255 then 10.0.1.0 — 257 unique IPv4 starting addresses.
		ranges[i] = netv1alpha1.IPRange{StartingAddress: fmt.Sprintf("10.0.%d.%d", i/256, i%256), AddressCount: 1}
	}
	obj.Spec.NetworkSpec.VirtualServerIPRanges = ranges
	if err := k8sClient.Create(testCtx, obj); !isRejected(err) {
		t.Fatalf("expected rejection for >256 IP ranges, got: %v", err)
	}
}

// TestFoundationLoadBalancerConfig_IPRangeExactly256_Admitted verifies that exactly 256 ranges
// (the MaxItems boundary) is admitted.
func TestFoundationLoadBalancerConfig_IPRangeExactly256_Admitted(t *testing.T) {
	ensureNamespace(t, flbcNS)
	obj := validFLBC("flbc-ranges-max")
	obj.Spec.NetworkSpec.VirtualServerIPPools = nil
	ranges := make([]netv1alpha1.IPRange, 256)
	for i := range ranges {
		// 10.0.0.0 through 10.0.0.255 — 256 unique IPv4 starting addresses.
		ranges[i] = netv1alpha1.IPRange{StartingAddress: fmt.Sprintf("10.0.0.%d", i), AddressCount: 1}
	}
	obj.Spec.NetworkSpec.VirtualServerIPRanges = ranges
	if err := k8sClient.Create(testCtx, obj); err != nil {
		t.Fatalf("expected admission for exactly 256 IP ranges, got: %v", err)
	}
	_ = k8sClient.Delete(testCtx, obj)
}

// TestFoundationLoadBalancerConfig_IPRangeNonOverlapping_Admitted verifies that two IPRange
// entries with distinct, non-overlapping addresses are admitted. Overlap and uniqueness
// validation is handled by the FLBC admission webhook, not the CRD schema.
func TestFoundationLoadBalancerConfig_IPRangeNonOverlapping_Admitted(t *testing.T) {
	ensureNamespace(t, flbcNS)
	obj := validFLBC("flbc-ranges-nooverlap")
	obj.Spec.NetworkSpec.VirtualServerIPPools = nil
	obj.Spec.NetworkSpec.VirtualServerIPRanges = []netv1alpha1.IPRange{
		{StartingAddress: "10.0.0.1", AddressCount: 4},
		{StartingAddress: "10.0.1.1", AddressCount: 4},
	}
	if err := k8sClient.Create(testCtx, obj); err != nil {
		t.Fatalf("expected admission for non-overlapping virtualServerIPRanges, got: %v", err)
	}
	_ = k8sClient.Delete(testCtx, obj)
}

// --- EffectiveVirtualServerIPPools status field ---

// TestFoundationLoadBalancerConfig_EffectiveVirtualServerIPPools_StatusRoundtrip verifies that the
// effectiveVirtualServerIPPools status field can be written via the status subresource and read
// back correctly.
func TestFoundationLoadBalancerConfig_EffectiveVirtualServerIPPools_StatusRoundtrip(t *testing.T) {
	ensureNamespace(t, flbcNS)
	obj := validFLBC("flbc-effective-pools")
	if err := k8sClient.Create(testCtx, obj); err != nil {
		t.Fatalf("create: %v", err)
	}
	defer func() { _ = k8sClient.Delete(testCtx, obj) }()

	latest := &netv1alpha1.FoundationLoadBalancerConfig{}
	if err := k8sClient.Get(testCtx, client.ObjectKeyFromObject(obj), latest); err != nil {
		t.Fatalf("get: %v", err)
	}

	latest.Status.EffectiveVirtualServerIPPools = []string{"pool-managed-1", "pool-managed-2"}
	if err := k8sClient.Status().Update(testCtx, latest); err != nil {
		t.Fatalf("status update: %v", err)
	}

	updated := &netv1alpha1.FoundationLoadBalancerConfig{}
	if err := k8sClient.Get(testCtx, client.ObjectKeyFromObject(obj), updated); err != nil {
		t.Fatalf("get after status update: %v", err)
	}
	if len(updated.Status.EffectiveVirtualServerIPPools) != 2 {
		t.Fatalf("expected 2 effective pools, got %d", len(updated.Status.EffectiveVirtualServerIPPools))
	}
	if updated.Status.EffectiveVirtualServerIPPools[0] != "pool-managed-1" {
		t.Errorf("pool[0]: got %q, want %q", updated.Status.EffectiveVirtualServerIPPools[0], "pool-managed-1")
	}
	if updated.Status.EffectiveVirtualServerIPPools[1] != "pool-managed-2" {
		t.Errorf("pool[1]: got %q, want %q", updated.Status.EffectiveVirtualServerIPPools[1], "pool-managed-2")
	}
}

// TestFoundationLoadBalancerConfig_EffectiveVirtualServerIPPools_AbsentByDefault verifies that
// effectiveVirtualServerIPPools is absent from a freshly created FLBC (omitempty).
func TestFoundationLoadBalancerConfig_EffectiveVirtualServerIPPools_AbsentByDefault(t *testing.T) {
	ensureNamespace(t, flbcNS)
	obj := validFLBC("flbc-effective-absent")
	if err := k8sClient.Create(testCtx, obj); err != nil {
		t.Fatalf("create: %v", err)
	}
	defer func() { _ = k8sClient.Delete(testCtx, obj) }()

	got := &netv1alpha1.FoundationLoadBalancerConfig{}
	if err := k8sClient.Get(testCtx, client.ObjectKeyFromObject(obj), got); err != nil {
		t.Fatalf("get: %v", err)
	}
	if len(got.Status.EffectiveVirtualServerIPPools) != 0 {
		t.Errorf("expected empty effectiveVirtualServerIPPools on fresh FLBC, got %v",
			got.Status.EffectiveVirtualServerIPPools)
	}
}

// --- append-only enforcement (oldSelf CEL rules) ---

// TestFoundationLoadBalancerConfig_VirtualServerIPPools_RemovalRejected verifies that once a
// pool entry exists it cannot be removed by any client — the oldSelf XValidation rule rejects
// updates that drop a previously-present name.
func TestFoundationLoadBalancerConfig_VirtualServerIPPools_RemovalRejected(t *testing.T) {
	ensureNamespace(t, flbcNS)
	obj := validFLBC("flbc-pool-removal")
	obj.Spec.NetworkSpec.VirtualServerIPPools = []netv1alpha1.IPPoolReference{
		{Name: "pool-keep"},
		{Name: "pool-drop"},
	}
	if err := k8sClient.Create(testCtx, obj); err != nil {
		t.Fatalf("create: %v", err)
	}
	defer func() { _ = k8sClient.Delete(testCtx, obj) }()

	latest := &netv1alpha1.FoundationLoadBalancerConfig{}
	if err := k8sClient.Get(testCtx, client.ObjectKeyFromObject(obj), latest); err != nil {
		t.Fatalf("get: %v", err)
	}
	latest.Spec.NetworkSpec.VirtualServerIPPools = []netv1alpha1.IPPoolReference{
		{Name: "pool-keep"},
	}
	if err := k8sClient.Update(testCtx, latest); !isRejected(err) {
		t.Fatalf("expected rejection when removing pool-drop from virtualServerIPPools, got: %v", err)
	}
}

// TestFoundationLoadBalancerConfig_VirtualServerIPPools_AppendAdmitted verifies that adding a
// new pool entry to an existing list is permitted by the append-only rule.
func TestFoundationLoadBalancerConfig_VirtualServerIPPools_AppendAdmitted(t *testing.T) {
	ensureNamespace(t, flbcNS)
	obj := validFLBC("flbc-pool-append")
	obj.Spec.NetworkSpec.VirtualServerIPPools = []netv1alpha1.IPPoolReference{
		{Name: "pool-original"},
	}
	if err := k8sClient.Create(testCtx, obj); err != nil {
		t.Fatalf("create: %v", err)
	}
	defer func() { _ = k8sClient.Delete(testCtx, obj) }()

	latest := &netv1alpha1.FoundationLoadBalancerConfig{}
	if err := k8sClient.Get(testCtx, client.ObjectKeyFromObject(obj), latest); err != nil {
		t.Fatalf("get: %v", err)
	}
	latest.Spec.NetworkSpec.VirtualServerIPPools = []netv1alpha1.IPPoolReference{
		{Name: "pool-original"},
		{Name: "pool-new"},
	}
	if err := k8sClient.Update(testCtx, latest); err != nil {
		t.Fatalf("expected admission when appending to virtualServerIPPools, got: %v", err)
	}
}

// TestFoundationLoadBalancerConfig_VirtualServerIPRanges_RemovalRejected verifies that once a
// range entry exists it cannot be removed — the oldSelf XValidation rule rejects updates that
// drop a previously-present startingAddress.
func TestFoundationLoadBalancerConfig_VirtualServerIPRanges_RemovalRejected(t *testing.T) {
	ensureNamespace(t, flbcNS)
	obj := validFLBC("flbc-range-removal")
	obj.Spec.NetworkSpec.VirtualServerIPPools = nil
	obj.Spec.NetworkSpec.VirtualServerIPRanges = []netv1alpha1.IPRange{
		{StartingAddress: "10.1.0.1", AddressCount: 4},
		{StartingAddress: "10.1.1.1", AddressCount: 4},
	}
	if err := k8sClient.Create(testCtx, obj); err != nil {
		t.Fatalf("create: %v", err)
	}
	defer func() { _ = k8sClient.Delete(testCtx, obj) }()

	latest := &netv1alpha1.FoundationLoadBalancerConfig{}
	if err := k8sClient.Get(testCtx, client.ObjectKeyFromObject(obj), latest); err != nil {
		t.Fatalf("get: %v", err)
	}
	latest.Spec.NetworkSpec.VirtualServerIPRanges = []netv1alpha1.IPRange{
		{StartingAddress: "10.1.0.1", AddressCount: 4},
	}
	if err := k8sClient.Update(testCtx, latest); !isRejected(err) {
		t.Fatalf("expected rejection when removing 10.1.1.1 from virtualServerIPRanges, got: %v", err)
	}
}

// TestFoundationLoadBalancerConfig_VirtualServerIPRanges_AppendAdmitted verifies that adding a
// new range entry to an existing list is permitted by the append-only rule.
func TestFoundationLoadBalancerConfig_VirtualServerIPRanges_AppendAdmitted(t *testing.T) {
	ensureNamespace(t, flbcNS)
	obj := validFLBC("flbc-range-append")
	obj.Spec.NetworkSpec.VirtualServerIPPools = nil
	obj.Spec.NetworkSpec.VirtualServerIPRanges = []netv1alpha1.IPRange{
		{StartingAddress: "10.2.0.1", AddressCount: 4},
	}
	if err := k8sClient.Create(testCtx, obj); err != nil {
		t.Fatalf("create: %v", err)
	}
	defer func() { _ = k8sClient.Delete(testCtx, obj) }()

	latest := &netv1alpha1.FoundationLoadBalancerConfig{}
	if err := k8sClient.Get(testCtx, client.ObjectKeyFromObject(obj), latest); err != nil {
		t.Fatalf("get: %v", err)
	}
	latest.Spec.NetworkSpec.VirtualServerIPRanges = []netv1alpha1.IPRange{
		{StartingAddress: "10.2.0.1", AddressCount: 4},
		{StartingAddress: "10.2.1.1", AddressCount: 4},
	}
	if err := k8sClient.Update(testCtx, latest); err != nil {
		t.Fatalf("expected admission when appending to virtualServerIPRanges, got: %v", err)
	}
}

// makeFlbcUnstructured builds a minimal valid FLBC as an Unstructured object, merging extra
// networkSpec fields.  This allows tests to send exact wire values (e.g. empty strings) that the
// typed Go client would silently drop via omitempty.
func makeFlbcUnstructured(name string, extraNetworkSpec map[string]interface{}) *unstructured.Unstructured {
	networkSpec := map[string]interface{}{
		"virtualServerIPPools": []interface{}{
			map[string]interface{}{"name": "pool-1"},
		},
	}
	for k, v := range extraNetworkSpec {
		networkSpec[k] = v
	}
	return &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "netoperator.vmware.com/v1alpha1",
			"kind":       "FoundationLoadBalancerConfig",
			"metadata": map[string]interface{}{
				"name":      name,
				"namespace": flbcNS,
			},
			"spec": map[string]interface{}{
				"deploymentSpec": map[string]interface{}{
					"size":             "small",
					"availabilityMode": "active-passive",
					"storagePolicy":    "default",
					"zones":            []interface{}{"zone-a"},
					"activePassiveSpec": map[string]interface{}{
						"replicas": int64(2),
					},
				},
				"virtualIPNetwork": map[string]interface{}{
					"kind": "Network",
					"name": "vip-net",
				},
				"networkSpec": networkSpec,
			},
		},
	}
}
