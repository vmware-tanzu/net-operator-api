// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: Apache-2.0

package v1alpha1_test

import (
	"testing"

	apiequality "k8s.io/apimachinery/pkg/api/equality"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/diff"
	"sigs.k8s.io/randfill"

	"github.com/vmware-tanzu/net-operator-api/api/v1alpha1"
	"github.com/vmware-tanzu/net-operator-api/api/v1alpha2"
)

func toHub(t *testing.T, in *v1alpha1.VSphereDistributedNetwork) *v1alpha2.VSphereDistributedNetwork {
	t.Helper()
	hub := &v1alpha2.VSphereDistributedNetwork{}
	if err := in.DeepCopy().ConvertTo(hub); err != nil {
		t.Fatalf("ConvertTo: %v", err)
	}
	return hub
}

func fromHub(t *testing.T, hub *v1alpha2.VSphereDistributedNetwork) *v1alpha1.VSphereDistributedNetwork {
	t.Helper()
	out := &v1alpha1.VSphereDistributedNetwork{}
	if err := out.ConvertFrom(hub.DeepCopy()); err != nil {
		t.Fatalf("ConvertFrom: %v", err)
	}
	return out
}

// requireSpokeRoundTrip checks that v1alpha1 -> v1alpha2 -> v1alpha1 returns exactly the input.
// Legacy writers compare specs with reflect.DeepEqual, so even nil versus empty ipPools matters.
func requireSpokeRoundTrip(t *testing.T, in *v1alpha1.VSphereDistributedNetwork) {
	t.Helper()
	out := fromHub(t, toHub(t, in))
	if !apiequality.Semantic.DeepEqual(in, out) {
		t.Fatalf("v1alpha1 round trip changed the object:\n%s", diff.Diff(in, out))
	}
}

func requireHubRoundTrip(t *testing.T, hub *v1alpha2.VSphereDistributedNetwork) {
	t.Helper()
	got := toHub(t, fromHub(t, hub))
	if !apiequality.Semantic.DeepEqual(hub, got) {
		t.Fatalf("v1alpha2 round trip changed the object:\n%s", diff.Diff(hub, got))
	}
}

func legacyStaticPoolVDN() *v1alpha1.VSphereDistributedNetwork {
	return &v1alpha1.VSphereDistributedNetwork{
		ObjectMeta: metav1.ObjectMeta{Name: "primary"},
		Spec: v1alpha1.VSphereDistributedNetworkSpec{
			PortGroupID: "dvportgroup-1",
			IPPools:     []v1alpha1.IPPoolReference{{Name: "primary-10-0-0-10", APIVersion: "v1alpha1"}},
			Gateway:     "10.0.0.1",
			SubnetMask:  "255.255.255.0",
		},
	}
}

func rangedHub() *v1alpha2.VSphereDistributedNetwork {
	return &v1alpha2.VSphereDistributedNetwork{
		ObjectMeta: metav1.ObjectMeta{Name: "workload", Annotations: map[string]string{"keep": "me"}},
		Spec: v1alpha2.VSphereDistributedNetworkSpec{
			PortGroupID: "dvportgroup-1",
			IPv4: v1alpha2.VSphereDistributedNetworkIPv4Config{
				AssignmentMode: v1alpha2.IPAssignmentModeStaticPool,
				Gateway:        "10.0.0.1",
				PrefixLength:   new(int32(24)),
				AddressRanges:  []v1alpha2.VSphereDistributedNetworkIPRange{{StartingAddress: "10.0.0.10", AddressCount: 5}},
			},
		},
		Status: &v1alpha2.VSphereDistributedNetworkStatus{
			Conditions: []metav1.Condition{{
				Type:               v1alpha2.VSphereDistributedNetworkIPPoolInvalid,
				Status:             metav1.ConditionFalse,
				Reason:             "IPPoolsReady",
				LastTransitionTime: metav1.Now().Rfc3339Copy(),
			}},
			IPPools: []v1alpha2.IPPoolReference{{Name: "workload-10-0-0-10", APIVersion: "v1alpha1"}},
		},
	}
}

func TestConvert_LegacyRefsOnly(t *testing.T) {
	in := legacyStaticPoolVDN()
	hub := toHub(t, in)

	if hub.Spec.IPv4.AssignmentMode != v1alpha2.IPAssignmentModeStaticPool {
		t.Errorf("ipv4.assignmentMode = %q, want staticpool", hub.Spec.IPv4.AssignmentMode)
	}
	if hub.Spec.IPv4.PrefixLength == nil || *hub.Spec.IPv4.PrefixLength != 24 {
		t.Errorf("ipv4.prefixLength = %v, want 24", hub.Spec.IPv4.PrefixLength)
	}
	want := []v1alpha2.IPPoolReference{{Name: "primary-10-0-0-10", APIVersion: "v1alpha1"}}
	if hub.Status == nil || !apiequality.Semantic.DeepEqual(hub.Status.IPPools, want) {
		t.Errorf("status = %+v, want ipPools %v", hub.Status, want)
	}
	if _, ok := hub.Annotations[v1alpha1.VSphereDistributedNetworkV1alpha1DataAnnotation]; !ok {
		t.Errorf("expected v1alpha1 data for the unset ipAssignmentMode")
	}
	requireSpokeRoundTrip(t, in)

	out := fromHub(t, hub)
	if _, ok := out.Annotations[v1alpha1.VSphereDistributedNetworkV1alpha1DataAnnotation]; ok {
		t.Errorf("v1alpha1 data visible through v1alpha1: %v", out.Annotations)
	}
	if _, ok := out.Annotations[v1alpha1.VSphereDistributedNetworkV1alpha2DataAnnotation]; ok {
		t.Errorf("a network without v1alpha2-only fields got v1alpha2 data: %v", out.Annotations)
	}
}

func TestConvert_DHCPKeepsEmptyIPPools(t *testing.T) {
	in := &v1alpha1.VSphereDistributedNetwork{
		ObjectMeta: metav1.ObjectMeta{Name: "dhcp"},
		Spec: v1alpha1.VSphereDistributedNetworkSpec{
			PortGroupID:      "dvportgroup-1",
			IPAssignmentMode: v1alpha1.IPAssignmentModeDHCP,
			IPPools:          []v1alpha1.IPPoolReference{},
		},
	}
	hub := toHub(t, in)
	if hub.Annotations != nil || hub.Status != nil {
		t.Errorf("hub = %+v, want no annotations and no status", hub)
	}
	if out := fromHub(t, hub); out.Spec.IPPools == nil {
		t.Errorf("spec.ipPools is nil, want an empty list")
	}
	requireSpokeRoundTrip(t, in)
}

func TestConvert_NonCanonicalSubnetMask(t *testing.T) {
	in := legacyStaticPoolVDN()
	in.Spec.SubnetMask = "255.255.255.000"
	requireSpokeRoundTrip(t, in)

	// Once a v1alpha2 client changes the prefix length, the stashed mask no longer applies.
	hub := toHub(t, in)
	hub.Spec.IPv4.PrefixLength = new(int32(16))
	if got := fromHub(t, hub).Spec.SubnetMask; got != "255.255.0.0" {
		t.Errorf("subnetMask = %q, want 255.255.0.0", got)
	}
}

func TestConvert_RangedNetworkThroughV1alpha1(t *testing.T) {
	hub := rangedHub()
	out := fromHub(t, hub)

	if out.Spec.IPAssignmentMode != v1alpha1.IPAssignmentModeStaticPool || out.Spec.SubnetMask != "255.255.255.0" {
		t.Errorf("unexpected v1alpha1 spec: %+v", out.Spec)
	}
	if len(out.Spec.IPPools) != 1 || out.Spec.IPPools[0].Name != "workload-10-0-0-10" {
		t.Errorf("spec.ipPools = %v, want status.ipPools", out.Spec.IPPools)
	}
	if _, ok := out.Annotations[v1alpha1.VSphereDistributedNetworkV1alpha2DataAnnotation]; !ok {
		t.Errorf("expected v1alpha2 data holding the address ranges")
	}
	requireHubRoundTrip(t, hub)

	if back := toHub(t, out); back.Annotations[v1alpha1.VSphereDistributedNetworkV1alpha2DataAnnotation] != "" {
		t.Errorf("v1alpha2 data stored on the hub: %v", back.Annotations)
	}
}

// TestConvert_V1alpha1UpdateKeepsRanges covers a v1alpha1 client that reads a network, changes
// a field it knows about, and writes it back: the ranges it can't see must survive.
func TestConvert_V1alpha1UpdateKeepsRanges(t *testing.T) {
	v1 := fromHub(t, rangedHub())
	v1.Spec.Gateway = "10.0.0.254"
	hub := toHub(t, v1)

	if hub.Spec.IPv4.Gateway != "10.0.0.254" {
		t.Errorf("gateway = %q, want the v1alpha1 change", hub.Spec.IPv4.Gateway)
	}
	if want := rangedHub().Spec.IPv4.AddressRanges; !apiequality.Semantic.DeepEqual(hub.Spec.IPv4.AddressRanges, want) {
		t.Errorf("addressRanges = %v, want %v", hub.Spec.IPv4.AddressRanges, want)
	}
}

func TestConvert_IPv6OnlyHub(t *testing.T) {
	hub := rangedHub()
	hub.Spec.IPv4 = v1alpha2.VSphereDistributedNetworkIPv4Config{AssignmentMode: v1alpha2.IPAssignmentModeNone}
	hub.Spec.IPv6 = v1alpha2.VSphereDistributedNetworkIPv6Config{
		AssignmentMode: v1alpha2.IPAssignmentModeStaticPool,
		Gateway:        "fd00::1",
		PrefixLength:   new(int32(64)),
		AddressRanges:  []v1alpha2.VSphereDistributedNetworkIPRange{{StartingAddress: "fd00::10", AddressCount: 16}},
	}
	hub.Status.IPPools = []v1alpha2.IPPoolReference{{Name: "workload-fd00-10", APIVersion: "v1alpha1"}}

	out := fromHub(t, hub)
	// v1alpha1 requires ipPools to be empty when ipAssignmentMode is none.
	if out.Spec.IPAssignmentMode != v1alpha1.IPAssignmentModeNone || out.Spec.IPPools == nil || len(out.Spec.IPPools) != 0 {
		t.Errorf("unexpected v1alpha1 spec: %+v", out.Spec)
	}
	requireHubRoundTrip(t, hub)
}

func TestConvert_MalformedV1alpha1DataIgnored(t *testing.T) {
	hub := toHub(t, legacyStaticPoolVDN())
	hub.Annotations[v1alpha1.VSphereDistributedNetworkV1alpha1DataAnnotation] = "{not json"
	if out := fromHub(t, hub); out.Spec.IPAssignmentMode != v1alpha1.IPAssignmentModeStaticPool {
		t.Errorf("ipAssignmentMode = %q, want staticpool when the data can't be read", out.Spec.IPAssignmentMode)
	}
}

func TestConvert_MalformedV1alpha2DataFailsWrite(t *testing.T) {
	v1 := fromHub(t, rangedHub())
	v1.Annotations[v1alpha1.VSphereDistributedNetworkV1alpha2DataAnnotation] = "{not json"
	if err := v1.ConvertTo(&v1alpha2.VSphereDistributedNetwork{}); err == nil {
		t.Fatalf("expected an error rather than dropping the address ranges")
	}
}

func newFiller(seed int64) *randfill.Filler {
	return randfill.NewWithSeed(seed).NilChance(0.3).NumElements(0, 4).Funcs(
		func(s *string, c randfill.Continue) {
			// Mix realistic values in with random ones so the address and mask paths are exercised.
			switch c.Intn(6) {
			case 0:
				*s = ""
			case 1:
				*s = "10.0.0.1"
			case 2:
				*s = "fd00::1"
			case 3:
				*s = "255.255.255.0"
			default:
				*s = c.String(0)
			}
		},
	)
}

// TestConvert_FuzzSpokeRoundTrip checks that any v1alpha1 object, valid or not, survives
// v1alpha1 -> v1alpha2 -> v1alpha1 unchanged.
func TestConvert_FuzzSpokeRoundTrip(t *testing.T) {
	f := newFiller(1)
	for i := range 2000 {
		in := &v1alpha1.VSphereDistributedNetwork{}
		f.Fill(in)
		in.TypeMeta = metav1.TypeMeta{}
		delete(in.Annotations, v1alpha1.VSphereDistributedNetworkV1alpha1DataAnnotation)
		delete(in.Annotations, v1alpha1.VSphereDistributedNetworkV1alpha2DataAnnotation)
		out := fromHub(t, toHub(t, in))
		if !apiequality.Semantic.DeepEqual(in, out) {
			t.Fatalf("iteration %d: v1alpha1 round trip changed the object:\n%s", i, diff.Diff(in, out))
		}
	}
}

// TestConvert_FuzzHubRoundTrip checks that v1alpha2 -> v1alpha1 -> v1alpha2 is unchanged for
// v1alpha2 objects within the bounds v1alpha2 validation enforces.
func TestConvert_FuzzHubRoundTrip(t *testing.T) {
	f := newFiller(2)
	for i := range 2000 {
		hub := &v1alpha2.VSphereDistributedNetwork{}
		f.Fill(hub)
		hub.TypeMeta = metav1.TypeMeta{}
		delete(hub.Annotations, v1alpha1.VSphereDistributedNetworkV1alpha1DataAnnotation)
		delete(hub.Annotations, v1alpha1.VSphereDistributedNetworkV1alpha2DataAnnotation)
		// v1alpha2 requires ipv4.assignmentMode and bounds the IPv4 prefix length.
		if hub.Spec.IPv4.AssignmentMode == "" {
			hub.Spec.IPv4.AssignmentMode = v1alpha2.IPAssignmentModeStaticPool
		}
		if p := hub.Spec.IPv4.PrefixLength; p != nil && (*p < 0 || *p > 32) {
			hub.Spec.IPv4.PrefixLength = new(int32(24))
		}
		if hub.Status != nil {
			// v1alpha1 conditions have no observedGeneration.
			for j := range hub.Status.Conditions {
				hub.Status.Conditions[j].ObservedGeneration = 0
			}
			if apiequality.Semantic.DeepEqual(*hub.Status, v1alpha2.VSphereDistributedNetworkStatus{}) {
				hub.Status = nil
			}
		}
		got := toHub(t, fromHub(t, hub))
		if !apiequality.Semantic.DeepEqual(hub, got) {
			t.Fatalf("iteration %d: v1alpha2 round trip changed the object:\n%s", i, diff.Diff(hub, got))
		}
	}
}
