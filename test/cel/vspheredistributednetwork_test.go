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

const vdsAPIVersion = "netoperator.vmware.com/v1alpha1"
const vdsKind = "VSphereDistributedNetwork"

// validVDS returns a minimal valid VSphereDistributedNetwork using
// IPAssignmentModeStaticPool, the mode that requires gateway/subnetMask.
func validVDS(name string) *netv1alpha1.VSphereDistributedNetwork {
	return &netv1alpha1.VSphereDistributedNetwork{
		ObjectMeta: metav1.ObjectMeta{Name: name},
		Spec: netv1alpha1.VSphereDistributedNetworkSpec{
			PortGroupID:      "dvportgroup-1",
			IPAssignmentMode: netv1alpha1.IPAssignmentModeStaticPool,
			Gateway:          "10.0.0.1",
			SubnetMask:       "255.255.255.0",
		},
	}
}

// unstrVDS builds an unstructured VSphereDistributedNetwork, merging extra
// spec fields onto a minimal valid base (portGroupID only).
func unstrVDS(name string, extraSpec map[string]interface{}) *unstructured.Unstructured {
	spec := map[string]interface{}{
		"portGroupID": "dvportgroup-1",
	}
	for k, v := range extraSpec {
		spec[k] = v
	}
	return &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": vdsAPIVersion,
			"kind":       vdsKind,
			"metadata": map[string]interface{}{
				"name": name,
			},
			"spec": spec,
		},
	}
}

func TestVSphereDistributedNetwork_MinimalValid_Admitted(t *testing.T) {
	obj := validVDS("vds-minimal")
	if err := k8sClient.Create(testCtx, obj); err != nil {
		t.Fatalf("expected admission, got: %v", err)
	}
	defer func() { _ = k8sClient.Delete(testCtx, obj) }()
}

func TestVSphereDistributedNetwork_MissingPortGroupID_Rejected(t *testing.T) {
	obj := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": vdsAPIVersion,
			"kind":       vdsKind,
			"metadata":   map[string]interface{}{"name": "vds-no-portgroup"},
			"spec":       map[string]interface{}{},
		},
	}
	if err := k8sClient.Create(testCtx, obj); !isRejected(err) {
		t.Fatalf("expected rejection for missing portGroupID, got: %v", err)
	}
}

// --- name validation ---

func TestVSphereDistributedNetwork_NameTooLong_Rejected(t *testing.T) {
	obj := validVDS(strings.Repeat("a", 254))
	if err := k8sClient.Create(testCtx, obj); !isRejected(err) {
		t.Fatalf("expected rejection for name exceeding 253 characters, got: %v", err)
	}
}

func TestVSphereDistributedNetwork_NameUppercase_Rejected(t *testing.T) {
	obj := validVDS("VDS-Uppercase")
	if err := k8sClient.Create(testCtx, obj); !isRejected(err) {
		t.Fatalf("expected rejection for uppercase name, got: %v", err)
	}
}

// --- ipAssignmentMode enum and immutability ---

func TestVSphereDistributedNetwork_InvalidIPAssignmentMode_Rejected(t *testing.T) {
	obj := validVDS("vds-bad-mode")
	obj.Spec.IPAssignmentMode = "bogus-mode"
	if err := k8sClient.Create(testCtx, obj); !isRejected(err) {
		t.Fatalf("expected rejection for invalid ipAssignmentMode, got: %v", err)
	}
}

func TestVSphereDistributedNetwork_IPAssignmentModeChange_Rejected(t *testing.T) {
	obj := validVDS("vds-mode-immutable")
	if err := k8sClient.Create(testCtx, obj); err != nil {
		t.Fatalf("create: %v", err)
	}
	defer func() { _ = k8sClient.Delete(testCtx, obj) }()

	latest := &netv1alpha1.VSphereDistributedNetwork{}
	if err := k8sClient.Get(testCtx, client.ObjectKeyFromObject(obj), latest); err != nil {
		t.Fatalf("get: %v", err)
	}
	latest.Spec.IPAssignmentMode = netv1alpha1.IPAssignmentModeDHCP
	latest.Spec.Gateway = ""
	latest.Spec.SubnetMask = ""
	if err := k8sClient.Update(testCtx, latest); !isRejected(err) {
		t.Fatalf("expected rejection for changing immutable ipAssignmentMode, got: %v", err)
	}
}

// --- dhcp/none mode: gateway/subnetMask/addressRanges/ipPools must be empty ---

func TestVSphereDistributedNetwork_DHCPWithGateway_Rejected(t *testing.T) {
	obj := unstrVDS("vds-dhcp-gateway", map[string]interface{}{
		"ipAssignmentMode": string(netv1alpha1.IPAssignmentModeDHCP),
		"gateway":          "10.0.0.1",
	})
	if err := k8sClient.Create(testCtx, obj); !isRejected(err) {
		t.Fatalf("expected rejection for gateway set under dhcp mode, got: %v", err)
	}
}

func TestVSphereDistributedNetwork_DHCPWithSubnetMask_Rejected(t *testing.T) {
	obj := unstrVDS("vds-dhcp-subnetmask", map[string]interface{}{
		"ipAssignmentMode": string(netv1alpha1.IPAssignmentModeDHCP),
		"subnetMask":       "255.255.255.0",
	})
	if err := k8sClient.Create(testCtx, obj); !isRejected(err) {
		t.Fatalf("expected rejection for subnetMask set under dhcp mode, got: %v", err)
	}
}

func TestVSphereDistributedNetwork_DHCPWithAddressRanges_Rejected(t *testing.T) {
	obj := unstrVDS("vds-dhcp-ranges", map[string]interface{}{
		"ipAssignmentMode": string(netv1alpha1.IPAssignmentModeDHCP),
		"addressRanges": []interface{}{
			map[string]interface{}{"address": "10.0.0.10", "count": int64(4)},
		},
	})
	if err := k8sClient.Create(testCtx, obj); !isRejected(err) {
		t.Fatalf("expected rejection for addressRanges set under dhcp mode, got: %v", err)
	}
}

func TestVSphereDistributedNetwork_DHCPWithIPPools_Rejected(t *testing.T) {
	obj := unstrVDS("vds-dhcp-ippools", map[string]interface{}{
		"ipAssignmentMode": string(netv1alpha1.IPAssignmentModeDHCP),
		"ipPools": []interface{}{
			map[string]interface{}{"name": "pool-1"},
		},
	})
	if err := k8sClient.Create(testCtx, obj); !isRejected(err) {
		t.Fatalf("expected rejection for ipPools set under dhcp mode, got: %v", err)
	}
}

func TestVSphereDistributedNetwork_NoneMode_Admitted(t *testing.T) {
	obj := unstrVDS("vds-none-mode", map[string]interface{}{
		"ipAssignmentMode": string(netv1alpha1.IPAssignmentModeNone),
	})
	if err := k8sClient.Create(testCtx, obj); err != nil {
		t.Fatalf("expected admission, got: %v", err)
	}
	defer func() { _ = k8sClient.Delete(testCtx, obj) }()
}

// --- staticpool/unset mode: gateway/subnetMask are required ---

func TestVSphereDistributedNetwork_StaticPoolMissingGateway_Rejected(t *testing.T) {
	obj := unstrVDS("vds-static-no-gateway", map[string]interface{}{
		"ipAssignmentMode": string(netv1alpha1.IPAssignmentModeStaticPool),
		"subnetMask":       "255.255.255.0",
	})
	if err := k8sClient.Create(testCtx, obj); !isRejected(err) {
		t.Fatalf("expected rejection for missing gateway under staticpool mode, got: %v", err)
	}
}

func TestVSphereDistributedNetwork_StaticPoolMissingSubnetMask_Rejected(t *testing.T) {
	obj := unstrVDS("vds-static-no-subnetmask", map[string]interface{}{
		"ipAssignmentMode": string(netv1alpha1.IPAssignmentModeStaticPool),
		"gateway":          "10.0.0.1",
	})
	if err := k8sClient.Create(testCtx, obj); !isRejected(err) {
		t.Fatalf("expected rejection for missing subnetMask under staticpool mode, got: %v", err)
	}
}

func TestVSphereDistributedNetwork_UnsetModeRequiresGatewayAndSubnetMask_Rejected(t *testing.T) {
	// ipAssignmentMode unset behaves like staticpool for this rule.
	obj := unstrVDS("vds-unset-mode-no-gw", map[string]interface{}{})
	if err := k8sClient.Create(testCtx, obj); !isRejected(err) {
		t.Fatalf("expected rejection for missing gateway/subnetMask with ipAssignmentMode unset, got: %v", err)
	}
}

func TestVSphereDistributedNetwork_StaticPoolValid_Admitted(t *testing.T) {
	obj := validVDS("vds-static-valid")
	if err := k8sClient.Create(testCtx, obj); err != nil {
		t.Fatalf("expected admission, got: %v", err)
	}
	defer func() { _ = k8sClient.Delete(testCtx, obj) }()
}

// --- addressRanges field validation ---

func TestVSphereDistributedNetwork_AddressRangeInvalidAddress_Rejected(t *testing.T) {
	obj := validVDS("vds-bad-range-address")
	obj.Spec.AddressRanges = []netv1alpha1.VSphereDistributedNetworkIPRange{
		{Address: "not-an-ip", Count: 4},
	}
	if err := k8sClient.Create(testCtx, obj); !isRejected(err) {
		t.Fatalf("expected rejection for invalid address range address, got: %v", err)
	}
}

func TestVSphereDistributedNetwork_AddressRangeZeroCount_Rejected(t *testing.T) {
	obj := validVDS("vds-zero-count")
	obj.Spec.AddressRanges = []netv1alpha1.VSphereDistributedNetworkIPRange{
		{Address: "10.0.0.10", Count: 0},
	}
	if err := k8sClient.Create(testCtx, obj); !isRejected(err) {
		t.Fatalf("expected rejection for addressRange count of 0, got: %v", err)
	}
}

func TestVSphereDistributedNetwork_AddressRangeValid_Admitted(t *testing.T) {
	obj := validVDS("vds-valid-range")
	obj.Spec.AddressRanges = []netv1alpha1.VSphereDistributedNetworkIPRange{
		{Address: "10.0.0.10", Count: 4},
		{Address: "fe80::1", Count: 8},
	}
	if err := k8sClient.Create(testCtx, obj); err != nil {
		t.Fatalf("expected admission, got: %v", err)
	}
	defer func() { _ = k8sClient.Delete(testCtx, obj) }()
}

func TestVSphereDistributedNetwork_TooManyAddressRanges_Rejected(t *testing.T) {
	ranges := make([]netv1alpha1.VSphereDistributedNetworkIPRange, 1025)
	for i := range ranges {
		ranges[i] = netv1alpha1.VSphereDistributedNetworkIPRange{
			Address: fmt.Sprintf("10.%d.%d.1", i/256, i%256),
			Count:   1,
		}
	}
	obj := validVDS("vds-too-many-ranges")
	obj.Spec.AddressRanges = ranges
	if err := k8sClient.Create(testCtx, obj); !isRejected(err) {
		t.Fatalf("expected rejection for 1025 addressRanges (maxItems=1024), got: %v", err)
	}
}

// --- DefaultPortConfig / VlanSpec / MacManagementPolicy ---

func TestVSphereDistributedNetwork_VlanInvalidType_Rejected(t *testing.T) {
	obj := unstrVDS("vds-bad-vlan-type", map[string]interface{}{
		"ipAssignmentMode": string(netv1alpha1.IPAssignmentModeNone),
	})
	obj.Object["status"] = map[string]interface{}{
		"defaultPortConfig": map[string]interface{}{
			"vlan": map[string]interface{}{"type": "bogus-vlan-type"},
		},
	}
	if err := k8sClient.Create(testCtx, obj); !isRejected(err) {
		t.Fatalf("expected rejection for invalid vlan type, got: %v", err)
	}
}

func TestVSphereDistributedNetwork_StatusValidVlanAndMacPolicy_Admitted(t *testing.T) {
	obj := validVDS("vds-status-vlan-mac")
	if err := k8sClient.Create(testCtx, obj); err != nil {
		t.Fatalf("create: %v", err)
	}
	defer func() { _ = k8sClient.Delete(testCtx, obj) }()

	latest := &netv1alpha1.VSphereDistributedNetwork{}
	if err := k8sClient.Get(testCtx, client.ObjectKeyFromObject(obj), latest); err != nil {
		t.Fatalf("get: %v", err)
	}
	vlanID := int32(100)
	latest.Status = netv1alpha1.VSphereDistributedNetworkStatus{
		DefaultPortConfig: &netv1alpha1.VSphereDistributedPortConfig{
			Vlan: &netv1alpha1.VlanSpec{
				Type:   netv1alpha1.VLANTypeStandard,
				VlanID: &vlanID,
			},
			MacManagementPolicy: &netv1alpha1.MacManagementPolicy{
				MacLearningPolicy: &netv1alpha1.MacLearningPolicy{
					Enabled: true,
				},
			},
		},
	}
	if err := k8sClient.Update(testCtx, latest); err != nil {
		t.Fatalf("expected admission for valid status vlan/mac policy, got: %v", err)
	}
}

func TestVSphereDistributedNetwork_VlanIDOutOfRange_Rejected(t *testing.T) {
	vlanID := int32(4095)
	obj := validVDS("vds-vlan-out-of-range")
	obj.Status = netv1alpha1.VSphereDistributedNetworkStatus{
		DefaultPortConfig: &netv1alpha1.VSphereDistributedPortConfig{
			Vlan: &netv1alpha1.VlanSpec{
				Type:   netv1alpha1.VLANTypeStandard,
				VlanID: &vlanID,
			},
		},
	}
	if err := k8sClient.Create(testCtx, obj); !isRejected(err) {
		t.Fatalf("expected rejection for vlanID above maximum (4094), got: %v", err)
	}
}

func TestVSphereDistributedNetwork_MacLearningPolicyInvalidLimitPolicy_Rejected(t *testing.T) {
	badPolicy := netv1alpha1.MacLimitPolicyType("bogus-policy")
	obj := validVDS("vds-bad-mac-limit-policy")
	obj.Status = netv1alpha1.VSphereDistributedNetworkStatus{
		DefaultPortConfig: &netv1alpha1.VSphereDistributedPortConfig{
			MacManagementPolicy: &netv1alpha1.MacManagementPolicy{
				MacLearningPolicy: &netv1alpha1.MacLearningPolicy{
					Enabled:     true,
					LimitPolicy: &badPolicy,
				},
			},
		},
	}
	if err := k8sClient.Create(testCtx, obj); !isRejected(err) {
		t.Fatalf("expected rejection for invalid macLearningPolicy.limitPolicy, got: %v", err)
	}
}
