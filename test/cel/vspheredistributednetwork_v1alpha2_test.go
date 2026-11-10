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
			PortGroupID:      "dvportgroup-1",
			IPAssignmentMode: netv1alpha2.IPAssignmentModeStaticPool,
			Gateway:          "10.0.0.1",
			SubnetMask:       "255.255.255.0",
			AddressRanges:    []netv1alpha2.VSphereDistributedNetworkIPRange{{Address: "10.0.0.10", Count: 10}},
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

func TestVSphereDistributedNetworkV1alpha2_StaticPoolWithRanges_Admitted(t *testing.T) {
	createOrFail(t, validVDSv2("vds2-static"))
}

func TestVSphereDistributedNetworkV1alpha2_IPv6Fields_Admitted(t *testing.T) {
	obj := validVDSv2("vds2-ipv6")
	obj.Spec.IPv6AssignmentMode = netv1alpha2.IPAssignmentModeStaticPool
	obj.Spec.IPv6Gateway = "fd00::1"
	obj.Spec.IPv6Prefix = new(int32(64))
	obj.Spec.AddressRanges = append(obj.Spec.AddressRanges, netv1alpha2.VSphereDistributedNetworkIPRange{Address: "fd00::10", Count: 16})
	createOrFail(t, obj)
}

func TestVSphereDistributedNetworkV1alpha2_Rejected(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(*netv1alpha2.VSphereDistributedNetwork)
	}{
		{"staticpool without gateway", func(o *netv1alpha2.VSphereDistributedNetwork) { o.Spec.Gateway = "" }},
		{"staticpool without subnetMask", func(o *netv1alpha2.VSphereDistributedNetwork) { o.Spec.SubnetMask = "" }},
		{"unset mode without gateway", func(o *netv1alpha2.VSphereDistributedNetwork) {
			o.Spec.IPAssignmentMode = ""
			o.Spec.Gateway = ""
		}},
		{"dhcp with gateway", func(o *netv1alpha2.VSphereDistributedNetwork) {
			o.Spec.IPAssignmentMode = netv1alpha2.IPAssignmentModeDHCP
			o.Spec.SubnetMask = ""
			o.Spec.AddressRanges = nil
		}},
		{"none with address ranges", func(o *netv1alpha2.VSphereDistributedNetwork) {
			o.Spec.IPAssignmentMode = netv1alpha2.IPAssignmentModeNone
			o.Spec.Gateway = ""
			o.Spec.SubnetMask = ""
		}},
		{"invalid mode", func(o *netv1alpha2.VSphereDistributedNetwork) { o.Spec.IPAssignmentMode = "bogus" }},
		{"range with an invalid address", func(o *netv1alpha2.VSphereDistributedNetwork) {
			o.Spec.AddressRanges[0].Address = "not-an-ip"
		}},
		{"range with a zero count", func(o *netv1alpha2.VSphereDistributedNetwork) { o.Spec.AddressRanges[0].Count = 0 }},
		{"more than 1024 ranges", func(o *netv1alpha2.VSphereDistributedNetwork) {
			ranges := make([]netv1alpha2.VSphereDistributedNetworkIPRange, 1025)
			for i := range ranges {
				ranges[i] = netv1alpha2.VSphereDistributedNetworkIPRange{Address: fmt.Sprintf("10.%d.%d.1", i/256, i%256), Count: 1}
			}
			o.Spec.AddressRanges = ranges
		}},
		{"portGroupIDs[0] not the primary", func(o *netv1alpha2.VSphereDistributedNetwork) {
			o.Spec.PortGroupIDs = []string{"dvportgroup-2", "dvportgroup-1"}
		}},
		{"duplicate portGroupIDs", func(o *netv1alpha2.VSphereDistributedNetwork) {
			o.Spec.PortGroupIDs = []string{"dvportgroup-1", "dvportgroup-1"}
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

func TestVSphereDistributedNetworkV1alpha2_AssignmentModeImmutable(t *testing.T) {
	obj := validVDSv2("vds2-mode-immutable")
	createOrFail(t, obj)
	obj.Spec.IPAssignmentMode = netv1alpha2.IPAssignmentModeDHCP
	obj.Spec.Gateway = ""
	obj.Spec.SubnetMask = ""
	obj.Spec.AddressRanges = nil
	if err := k8sClient.Update(testCtx, obj); !isRejected(err) {
		t.Fatalf("expected rejection for changing ipAssignmentMode, got: %v", err)
	}
}

// legacyVDS is what the legacy enablement path writes through v1alpha1: pool references next to
// the address ranges they were created from.
func legacyVDS(name string) *netv1alpha1.VSphereDistributedNetwork {
	return &netv1alpha1.VSphereDistributedNetwork{
		ObjectMeta: metav1.ObjectMeta{Name: name},
		Spec: netv1alpha1.VSphereDistributedNetworkSpec{
			PortGroupID:   "dvportgroup-1",
			IPPools:       []netv1alpha1.IPPoolReference{{Name: name + "-10-0-0-10", APIVersion: "v1alpha1"}},
			Gateway:       "10.0.0.1",
			SubnetMask:    "255.255.255.0",
			AddressRanges: []netv1alpha1.VSphereDistributedNetworkIPRange{{Address: "10.0.0.10", Count: 10}},
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
	// The legacy writer compares specs with reflect.DeepEqual; any difference would make it update
	// the network on every pass.
	if !reflect.DeepEqual(*want, got.Spec) {
		t.Fatalf("v1alpha1 spec changed:\nwant %+v\ngot  %+v", *want, got.Spec)
	}

	hub := &netv1alpha2.VSphereDistributedNetwork{}
	if err := k8sClient.Get(testCtx, client.ObjectKeyFromObject(obj), hub); err != nil {
		t.Fatalf("get v1alpha2: %v", err)
	}
	if len(hub.Spec.AddressRanges) != 1 || hub.Spec.Gateway != "10.0.0.1" {
		t.Errorf("unexpected v1alpha2 spec: %+v", hub.Spec)
	}
	if len(hub.Status.IPPools) != 1 || hub.Status.IPPools[0].Name != "vds-legacy-readback-10-0-0-10" {
		t.Errorf("status.ipPools = %v, want the v1alpha1 spec.ipPools", hub.Status.IPPools)
	}
}

func TestVSphereDistributedNetworkConversion_StatusIPPoolsVisibleAsV1alpha1Spec(t *testing.T) {
	obj := validVDSv2("vds2-status-ippools")
	createOrFail(t, obj)
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

// TestVSphereDistributedNetworkConversion_V1alpha1UpdateKeepsRangesAndPools checks that a v1alpha1
// client that reads a network, changes a field and writes it back through the API server leaves the
// address ranges and the pools as they were.
func TestVSphereDistributedNetworkConversion_V1alpha1UpdateKeepsRangesAndPools(t *testing.T) {
	obj := legacyVDS("vds-v1alpha1-update")
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
	if got.Spec.Gateway != "10.0.0.254" {
		t.Errorf("gateway = %q, want the v1alpha1 change", got.Spec.Gateway)
	}
	if len(got.Spec.AddressRanges) != 1 || got.Spec.AddressRanges[0].Address != "10.0.0.10" {
		t.Errorf("addressRanges = %v, want them unchanged", got.Spec.AddressRanges)
	}
	if len(got.Status.IPPools) != 1 {
		t.Errorf("status.ipPools = %v, want it unchanged", got.Status.IPPools)
	}
}

// TestVSphereDistributedNetworkConversion_V1alpha2SpecUpdateKeepsPools checks that a v1alpha2
// client changing the spec, with the status it read, does not drop the pools.
func TestVSphereDistributedNetworkConversion_V1alpha2SpecUpdateKeepsPools(t *testing.T) {
	obj := legacyVDS("vds-v1alpha2-update")
	createOrFail(t, obj)

	hub := &netv1alpha2.VSphereDistributedNetwork{}
	if err := k8sClient.Get(testCtx, client.ObjectKeyFromObject(obj), hub); err != nil {
		t.Fatalf("get v1alpha2: %v", err)
	}
	hub.Spec.AddressRanges = append(hub.Spec.AddressRanges, netv1alpha2.VSphereDistributedNetworkIPRange{Address: "10.0.0.50", Count: 5})
	if err := k8sClient.Update(testCtx, hub); err != nil {
		t.Fatalf("update v1alpha2: %v", err)
	}

	got := &netv1alpha1.VSphereDistributedNetwork{}
	if err := k8sClient.Get(testCtx, client.ObjectKeyFromObject(obj), got); err != nil {
		t.Fatalf("get v1alpha1: %v", err)
	}
	if len(got.Spec.AddressRanges) != 2 || len(got.Spec.IPPools) != 1 {
		t.Errorf("v1alpha1 view = %+v, want 2 ranges and the original pool", got.Spec)
	}
}
