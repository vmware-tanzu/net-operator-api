// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: Apache-2.0

package v1alpha1_test

import (
	"reflect"
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

func legacyVDN() *v1alpha1.VSphereDistributedNetwork {
	return &v1alpha1.VSphereDistributedNetwork{
		ObjectMeta: metav1.ObjectMeta{Name: "primary", Labels: map[string]string{"a": "b"}},
		Spec: v1alpha1.VSphereDistributedNetworkSpec{
			PortGroupID:        "dvportgroup-1",
			PortGroupIDs:       []string{"dvportgroup-1", "dvportgroup-2"},
			IPAssignmentMode:   v1alpha1.IPAssignmentModeStaticPool,
			IPv6AssignmentMode: v1alpha1.IPAssignmentModeStaticPool,
			IPPools:            []v1alpha1.IPPoolReference{{Name: "primary-10-0-0-10", APIVersion: "v1alpha1"}},
			Gateway:            "10.0.0.1",
			SubnetMask:         "255.255.255.0",
			IPv6Gateway:        "fd00::1",
			IPv6Prefix:         new(int32(64)),
			AddressRanges: []v1alpha1.VSphereDistributedNetworkIPRange{
				{Address: "10.0.0.10", Count: 4},
				{Address: "fd00::10", Count: 8},
			},
		},
		Status: v1alpha1.VSphereDistributedNetworkStatus{
			Conditions: []v1alpha1.VSphereDistributedNetworkCondition{{
				Type:   v1alpha1.VsphereDistributedNetworkIPPoolPressure,
				Status: "False",
				Reason: "Ok",
			}},
		},
	}
}

func TestConvertTo_IPPoolsMoveToStatus(t *testing.T) {
	in := legacyVDN()
	hub := toHub(t, in)

	want := []v1alpha2.IPPoolReference{{Name: "primary-10-0-0-10", APIVersion: "v1alpha1"}}
	if !reflect.DeepEqual(hub.Status.IPPools, want) {
		t.Errorf("status.ipPools = %v, want %v", hub.Status.IPPools, want)
	}
	if hub.Spec.Gateway != "10.0.0.1" || hub.Spec.SubnetMask != "255.255.255.0" || hub.Spec.IPv6Gateway != "fd00::1" {
		t.Errorf("spec not copied: %+v", hub.Spec)
	}
	if len(hub.Spec.AddressRanges) != 2 || hub.Spec.AddressRanges[1].Address != "fd00::10" {
		t.Errorf("addressRanges not copied: %+v", hub.Spec.AddressRanges)
	}
	if len(hub.Status.Conditions) != 1 || hub.Status.Conditions[0].Reason != "Ok" {
		t.Errorf("conditions not copied: %+v", hub.Status.Conditions)
	}
	if hub.Annotations != nil {
		t.Errorf("annotations = %v, want none: the two versions need no conversion data", hub.Annotations)
	}
}

func TestConvertFrom_StatusIPPoolsBecomeSpec(t *testing.T) {
	hub := toHub(t, legacyVDN())
	hub.Status.IPPools = []v1alpha2.IPPoolReference{{Name: "managed-1"}, {Name: "managed-2", APIVersion: "v1alpha1"}}

	out := fromHub(t, hub)
	want := []v1alpha1.IPPoolReference{{Name: "managed-1"}, {Name: "managed-2", APIVersion: "v1alpha1"}}
	if !reflect.DeepEqual(out.Spec.IPPools, want) {
		t.Errorf("spec.ipPools = %v, want %v", out.Spec.IPPools, want)
	}
}

// Legacy writers compare specs with reflect.DeepEqual, and v1alpha1 spec.ipPools defaults to an
// empty list, so a network with no pools must read back with an empty, not nil, list.
func TestConvertFrom_NoPoolsReadsBackEmptyNotNil(t *testing.T) {
	in := legacyVDN()
	in.Spec.IPPools = []v1alpha1.IPPoolReference{}
	hub := toHub(t, in)
	if hub.Status.IPPools != nil {
		t.Errorf("status.ipPools = %v, want nil", hub.Status.IPPools)
	}
	out := fromHub(t, hub)
	if out.Spec.IPPools == nil || len(out.Spec.IPPools) != 0 {
		t.Errorf("spec.ipPools = %#v, want an empty non-nil list", out.Spec.IPPools)
	}
	if !reflect.DeepEqual(in.Spec, out.Spec) {
		t.Errorf("spec changed:\n%s", diff.Diff(in.Spec, out.Spec))
	}
}

// otherHub is a hub of some other kind, to check that a conversion refuses it.
type otherHub struct{ v1alpha1.IPPool }

func (*otherHub) Hub() {}

func TestConvert_UnexpectedHubType(t *testing.T) {
	if err := legacyVDN().ConvertTo(&otherHub{}); err == nil {
		t.Errorf("ConvertTo accepted a hub of another kind")
	}
	if err := (&v1alpha1.VSphereDistributedNetwork{}).ConvertFrom(&otherHub{}); err == nil {
		t.Errorf("ConvertFrom accepted a hub of another kind")
	}
}

func newFiller(seed int64) *randfill.Filler {
	return randfill.NewWithSeed(seed).NilChance(0.3).NumElements(0, 4)
}

// TestConvert_FuzzSpokeRoundTrip checks that any v1alpha1 object survives v1alpha1 -> v1alpha2 ->
// v1alpha1 unchanged. It also fails when a field is added to v1alpha1 only, which the field copy
// would otherwise drop silently.
func TestConvert_FuzzSpokeRoundTrip(t *testing.T) {
	f := newFiller(1)
	for i := range 2000 {
		in := &v1alpha1.VSphereDistributedNetwork{}
		f.Fill(in)
		in.TypeMeta = metav1.TypeMeta{}
		out := fromHub(t, toHub(t, in))
		if !apiequality.Semantic.DeepEqual(in, out) {
			t.Fatalf("iteration %d: v1alpha1 round trip changed the object:\n%s", i, diff.Diff(in, out))
		}
	}
}

// TestConvert_FuzzHubRoundTrip checks the same from v1alpha2, which also fails when a field is
// added to v1alpha2 only.
func TestConvert_FuzzHubRoundTrip(t *testing.T) {
	f := newFiller(2)
	for i := range 2000 {
		hub := &v1alpha2.VSphereDistributedNetwork{}
		f.Fill(hub)
		hub.TypeMeta = metav1.TypeMeta{}
		got := toHub(t, fromHub(t, hub))
		if !apiequality.Semantic.DeepEqual(hub, got) {
			t.Fatalf("iteration %d: v1alpha2 round trip changed the object:\n%s", i, diff.Diff(hub, got))
		}
	}
}
