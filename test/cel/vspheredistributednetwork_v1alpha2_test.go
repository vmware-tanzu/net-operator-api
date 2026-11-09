// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: Apache-2.0

package cel_test

import (
	"fmt"
	"reflect"
	"testing"

	netv1alpha1 "github.com/vmware-tanzu/net-operator-api/api/v1alpha1"
	netv1alpha2 "github.com/vmware-tanzu/net-operator-api/api/v1alpha2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func validVDSv2(name string) *netv1alpha2.VSphereDistributedNetwork {
	return &netv1alpha2.VSphereDistributedNetwork{
		ObjectMeta: metav1.ObjectMeta{Name: name},
		Spec: netv1alpha2.VSphereDistributedNetworkSpec{
			PortGroupID: "dvportgroup-1",
			IPv4: netv1alpha2.VSphereDistributedNetworkIPv4Config{
				AssignmentMode: netv1alpha2.IPAssignmentModeStaticPool,
				Gateway:        "10.0.0.1",
				PrefixLength:   new(int32(24)),
				AddressRanges:  []netv1alpha2.VSphereDistributedNetworkIPRange{{StartingAddress: "10.0.0.10", AddressCount: 10}},
			},
		},
	}
}

func createOrFail(t *testing.T, obj client.Object) {
	t.Helper()
	if err := k8sClient.Create(testCtx, obj); err != nil {
		t.Fatalf("expected admission, got: %v", err)
	}
	t.Cleanup(func() { _ = k8sClient.Delete(testCtx, obj) })
}

func TestVSphereDistributedNetworkV1alpha2_StaticIPv4_Admitted(t *testing.T) {
	createOrFail(t, validVDSv2("vds2-static-ipv4"))
}

func TestVSphereDistributedNetworkV1alpha2_IPv6Only_Admitted(t *testing.T) {
	obj := validVDSv2("vds2-ipv6-only")
	obj.Spec.IPv4 = netv1alpha2.VSphereDistributedNetworkIPv4Config{AssignmentMode: netv1alpha2.IPAssignmentModeNone}
	obj.Spec.IPv6 = netv1alpha2.VSphereDistributedNetworkIPv6Config{
		AssignmentMode: netv1alpha2.IPAssignmentModeStaticPool,
		Gateway:        "fd00::1",
		PrefixLength:   new(int32(64)),
		AddressRanges:  []netv1alpha2.VSphereDistributedNetworkIPRange{{StartingAddress: "fd00::10", AddressCount: 16}},
	}
	createOrFail(t, obj)
}

func TestVSphereDistributedNetworkV1alpha2_Rejected(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(*netv1alpha2.VSphereDistributedNetwork)
	}{
		{"staticpool without gateway", func(o *netv1alpha2.VSphereDistributedNetwork) { o.Spec.IPv4.Gateway = "" }},
		{"staticpool without prefixLength", func(o *netv1alpha2.VSphereDistributedNetwork) { o.Spec.IPv4.PrefixLength = nil }},
		{"dhcp with gateway", func(o *netv1alpha2.VSphereDistributedNetwork) {
			o.Spec.IPv4.AssignmentMode = netv1alpha2.IPAssignmentModeDHCP
			o.Spec.IPv4.PrefixLength = nil
			o.Spec.IPv4.AddressRanges = nil
		}},
		{"none with ranges", func(o *netv1alpha2.VSphereDistributedNetwork) {
			o.Spec.IPv4.AssignmentMode = netv1alpha2.IPAssignmentModeNone
			o.Spec.IPv4.Gateway = ""
			o.Spec.IPv4.PrefixLength = nil
		}},
		{"IPv6 range in ipv4", func(o *netv1alpha2.VSphereDistributedNetwork) {
			o.Spec.IPv4.AddressRanges = []netv1alpha2.VSphereDistributedNetworkIPRange{{StartingAddress: "fd00::10", AddressCount: 1}}
		}},
		{"IPv6 gateway in ipv4", func(o *netv1alpha2.VSphereDistributedNetwork) { o.Spec.IPv4.Gateway = "fd00::1" }},
		{"ipv4 prefixLength over 32", func(o *netv1alpha2.VSphereDistributedNetwork) { o.Spec.IPv4.PrefixLength = new(int32(33)) }},
		{"ipv4 address with leading zeros", func(o *netv1alpha2.VSphereDistributedNetwork) {
			o.Spec.IPv4.AddressRanges[0].StartingAddress = "10.0.0.010"
		}},
		{"non-canonical IPv6 address", func(o *netv1alpha2.VSphereDistributedNetwork) {
			o.Spec.IPv6 = netv1alpha2.VSphereDistributedNetworkIPv6Config{
				AssignmentMode: netv1alpha2.IPAssignmentModeStaticPool,
				Gateway:        "fd00::1",
				PrefixLength:   new(int32(64)),
				AddressRanges:  []netv1alpha2.VSphereDistributedNetworkIPRange{{StartingAddress: "fd00:0::10", AddressCount: 1}},
			}
		}},
		{"IPv4 range in ipv6", func(o *netv1alpha2.VSphereDistributedNetwork) {
			o.Spec.IPv6 = netv1alpha2.VSphereDistributedNetworkIPv6Config{
				AssignmentMode: netv1alpha2.IPAssignmentModeStaticPool,
				Gateway:        "fd00::1",
				PrefixLength:   new(int32(64)),
				AddressRanges:  []netv1alpha2.VSphereDistributedNetworkIPRange{{StartingAddress: "10.0.1.10", AddressCount: 1}},
			}
		}},
		{"range with zero addressCount", func(o *netv1alpha2.VSphereDistributedNetwork) {
			o.Spec.IPv4.AddressRanges[0].AddressCount = 0
		}},
		{"more than 1024 ranges", func(o *netv1alpha2.VSphereDistributedNetwork) {
			ranges := make([]netv1alpha2.VSphereDistributedNetworkIPRange, 1025)
			for i := range ranges {
				ranges[i] = netv1alpha2.VSphereDistributedNetworkIPRange{StartingAddress: fmt.Sprintf("10.%d.%d.1", i/256, i%256), AddressCount: 1}
			}
			o.Spec.IPv4.AddressRanges = ranges
		}},
		{"missing ipv4 assignmentMode", func(o *netv1alpha2.VSphereDistributedNetwork) {
			o.Spec.IPv4 = netv1alpha2.VSphereDistributedNetworkIPv4Config{}
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			obj := validVDSv2("vds2-rejected")
			tt.mutate(obj)
			err := k8sClient.Create(testCtx, obj)
			if err == nil {
				_ = k8sClient.Delete(testCtx, obj)
			}
			if !isRejected(err) {
				t.Fatalf("expected rejection, got: %v", err)
			}
		})
	}
}

func TestVSphereDistributedNetworkV1alpha2_IPv4AssignmentModeImmutable(t *testing.T) {
	obj := validVDSv2("vds2-mode-immutable")
	createOrFail(t, obj)
	obj.Spec.IPv4 = netv1alpha2.VSphereDistributedNetworkIPv4Config{AssignmentMode: netv1alpha2.IPAssignmentModeDHCP}
	if err := k8sClient.Update(testCtx, obj); !isRejected(err) {
		t.Fatalf("expected rejection for changing ipv4.assignmentMode, got: %v", err)
	}
}

// legacyVDS is what wcpsvc's legacy path writes: ipAssignmentMode unset and IPPool references.
func legacyVDS(name string) *netv1alpha1.VSphereDistributedNetwork {
	return &netv1alpha1.VSphereDistributedNetwork{
		ObjectMeta: metav1.ObjectMeta{Name: name},
		Spec: netv1alpha1.VSphereDistributedNetworkSpec{
			PortGroupID: "dvportgroup-1",
			IPPools:     []netv1alpha1.IPPoolReference{{Name: name + "-10-0-0-10", APIVersion: "v1alpha1"}},
			Gateway:     "10.0.0.1",
			SubnetMask:  "255.255.255.0",
		},
	}
}

func TestVSphereDistributedNetworkConversion_LegacySpecReadsBackUnchanged(t *testing.T) {
	obj := legacyVDS("vds-legacy-readback")
	want := obj.Spec.DeepCopy()
	createOrFail(t, obj)

	got := &netv1alpha1.VSphereDistributedNetwork{}
	if err := k8sClient.Get(testCtx, client.ObjectKeyFromObject(obj), got); err != nil {
		t.Fatalf("get v1alpha1: %v", err)
	}
	// wcpsvc compares specs with reflect.DeepEqual; any difference would make it update forever.
	if !reflect.DeepEqual(*want, got.Spec) {
		t.Fatalf("v1alpha1 spec changed:\nwant %+v\ngot  %+v", *want, got.Spec)
	}
	if _, ok := got.Annotations[netv1alpha1.VSphereDistributedNetworkV1alpha1DataAnnotation]; ok {
		t.Errorf("v1alpha1 data visible through v1alpha1")
	}

	hub := &netv1alpha2.VSphereDistributedNetwork{}
	if err := k8sClient.Get(testCtx, client.ObjectKeyFromObject(obj), hub); err != nil {
		t.Fatalf("get v1alpha2: %v", err)
	}
	if hub.Spec.IPv4.AssignmentMode != netv1alpha2.IPAssignmentModeStaticPool ||
		hub.Spec.IPv4.PrefixLength == nil || *hub.Spec.IPv4.PrefixLength != 24 {
		t.Errorf("unexpected v1alpha2 ipv4: %+v", hub.Spec.IPv4)
	}
	if len(hub.Status.IPPools) != 1 || hub.Status.IPPools[0].Name != "vds-legacy-readback-10-0-0-10" {
		t.Errorf("status.ipPools = %v, want the v1alpha1 spec.ipPools", hub.Status.IPPools)
	}
}

func TestVSphereDistributedNetworkConversion_StatusIPPoolsVisibleAsV1alpha1Spec(t *testing.T) {
	obj := validVDSv2("vds2-status-ippools")
	createOrFail(t, obj)
	obj.Status = &netv1alpha2.VSphereDistributedNetworkStatus{}
	obj.Status.IPPools = []netv1alpha2.IPPoolReference{{Name: "vds2-status-ippools-10-0-0-10", APIVersion: "v1alpha1"}}
	if err := k8sClient.Update(testCtx, obj); err != nil {
		t.Fatalf("update v1alpha2 status: %v", err)
	}

	got := &netv1alpha1.VSphereDistributedNetwork{}
	if err := k8sClient.Get(testCtx, client.ObjectKeyFromObject(obj), got); err != nil {
		t.Fatalf("get v1alpha1: %v", err)
	}
	if len(got.Spec.IPPools) != 1 || got.Spec.IPPools[0].Name != "vds2-status-ippools-10-0-0-10" {
		t.Errorf("v1alpha1 spec.ipPools = %v, want v1alpha2 status.ipPools", got.Spec.IPPools)
	}
}

// TestVSphereDistributedNetworkConversion_V1alpha2WriteOfLegacyObject checks that validation
// ratcheting lets a v1alpha2 client update an object whose spec, written through v1alpha1, would
// fail v1alpha2 validation. net-operator writes status with full-object updates, so without this
// it couldn't report status on such networks.
func TestVSphereDistributedNetworkConversion_V1alpha2WriteOfLegacyObject(t *testing.T) {
	obj := legacyVDS("vds-legacy-ratchet")
	obj.Spec.Gateway = "10.0.0.001" // accepted by v1alpha1's pattern, rejected by v1alpha2
	createOrFail(t, obj)

	hub := &netv1alpha2.VSphereDistributedNetwork{}
	if err := k8sClient.Get(testCtx, client.ObjectKeyFromObject(obj), hub); err != nil {
		t.Fatalf("get v1alpha2: %v", err)
	}
	if hub.Status == nil {
		hub.Status = &netv1alpha2.VSphereDistributedNetworkStatus{}
	}
	hub.Status.Conditions = []metav1.Condition{{
		Type:               netv1alpha2.VSphereDistributedNetworkIPPoolInvalid,
		Status:             metav1.ConditionFalse,
		Reason:             "IPPoolsReady",
		LastTransitionTime: metav1.Now(),
	}}
	if err := k8sClient.Update(testCtx, hub); err != nil {
		t.Fatalf("expected the v1alpha2 status write to be admitted, got: %v", err)
	}

	got := &netv1alpha1.VSphereDistributedNetwork{}
	if err := k8sClient.Get(testCtx, client.ObjectKeyFromObject(obj), got); err != nil {
		t.Fatalf("get v1alpha1: %v", err)
	}
	if got.Spec.Gateway != "10.0.0.001" {
		t.Errorf("gateway = %q, want it unchanged", got.Spec.Gateway)
	}
}

// TestVSphereDistributedNetworkConversion_V1alpha1UpdateKeepsRanges checks, through the API server,
// that a v1alpha1 client updating a network with address ranges, which it can't see, keeps them.
func TestVSphereDistributedNetworkConversion_V1alpha1UpdateKeepsRanges(t *testing.T) {
	obj := validVDSv2("vds2-v1alpha1-update")
	createOrFail(t, obj)

	v1 := &netv1alpha1.VSphereDistributedNetwork{}
	if err := k8sClient.Get(testCtx, client.ObjectKeyFromObject(obj), v1); err != nil {
		t.Fatalf("get v1alpha1: %v", err)
	}
	v1.Spec.Gateway = "10.0.0.254"
	if err := k8sClient.Update(testCtx, v1); err != nil {
		t.Fatalf("update v1alpha1: %v", err)
	}

	got := &netv1alpha2.VSphereDistributedNetwork{}
	if err := k8sClient.Get(testCtx, client.ObjectKeyFromObject(obj), got); err != nil {
		t.Fatalf("get v1alpha2: %v", err)
	}
	if got.Spec.IPv4.Gateway != "10.0.0.254" {
		t.Errorf("gateway = %q, want the v1alpha1 change", got.Spec.IPv4.Gateway)
	}
	if !reflect.DeepEqual(got.Spec.IPv4.AddressRanges, obj.Spec.IPv4.AddressRanges) {
		t.Errorf("addressRanges = %v, want %v", got.Spec.IPv4.AddressRanges, obj.Spec.IPv4.AddressRanges)
	}
	if _, ok := got.Annotations[netv1alpha1.VSphereDistributedNetworkV1alpha2DataAnnotation]; ok {
		t.Errorf("v1alpha2 data stored on the object: %v", got.Annotations)
	}
}
