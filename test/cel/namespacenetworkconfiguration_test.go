// Copyright (c) 2026 Broadcom. All Rights Reserved.
// Broadcom Confidential. The term "Broadcom" refers to Broadcom Inc.
// and/or its subsidiaries.

package cel_test

import (
	"fmt"
	"strings"
	"testing"
	"time"

	netv1alpha1 "github.com/vmware-tanzu/net-operator-api/api/v1alpha1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

// -----------------------------------------------------------------------
// Builders (ported from local closures in the source Ginkgo spec)
// -----------------------------------------------------------------------

// unstrNNC creates a bare unstructured NNC. Pass nil for spec to omit
// the spec key entirely; pass a non-nil map to include it verbatim.
func unstrNNC(name string, spec map[string]interface{}) *unstructured.Unstructured {
	obj := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "netoperator.vmware.com/v1alpha1",
			"kind":       "NamespaceNetworkConfiguration",
			"metadata":   map[string]interface{}{"name": name},
		},
	}
	if spec != nil {
		obj.Object["spec"] = spec
	}
	return obj
}

// vdsNNC builds a minimal valid vsphere-distributed NNC. The first name in
// nets becomes the defaultNetwork.
func vdsNNC(name string, nets ...string) *netv1alpha1.NamespaceNetworkConfiguration {
	refs := make([]netv1alpha1.VSphereDistributedNetworkRef, 0, len(nets))
	for _, n := range nets {
		refs = append(refs, netv1alpha1.VSphereDistributedNetworkRef{Name: n})
	}
	defaultNet := ""
	if len(nets) > 0 {
		defaultNet = nets[0]
	}
	return &netv1alpha1.NamespaceNetworkConfiguration{
		ObjectMeta: metav1.ObjectMeta{Name: name},
		Spec: netv1alpha1.NamespaceNetworkSpec{
			Type: netv1alpha1.NetworkProviderVSphereDistributed,
			NamespaceNetworkConfig: netv1alpha1.NamespaceNetworkConfig{
				VSphereDistributedConfig: netv1alpha1.VSphereDistributedConfig{
					DefaultNetwork: defaultNet,
					Networks:       refs,
				},
			},
		},
	}
}

// makeCondition returns a complete, valid metav1.Condition.
func makeCondition(condType string, status metav1.ConditionStatus) metav1.Condition {
	return metav1.Condition{
		Type:               condType,
		Status:             status,
		Reason:             testConditionReason,
		Message:            "test message",
		LastTransitionTime: metav1.NewTime(time.Now()),
	}
}

// vpcNNC builds a minimal valid pre-created-VPC NNC.
func vpcNNC(name, vpcPath string) *netv1alpha1.NamespaceNetworkConfiguration {
	return &netv1alpha1.NamespaceNetworkConfiguration{
		ObjectMeta: metav1.ObjectMeta{Name: name},
		Spec: netv1alpha1.NamespaceNetworkSpec{
			Type: netv1alpha1.NetworkProviderVPC,
			NamespaceNetworkConfig: netv1alpha1.NamespaceNetworkConfig{
				VPCConfig: netv1alpha1.VPCConfig{VPC: vpcPath},
			},
		},
	}
}

// autoVpcNNC builds a minimal valid auto-create VPC NNC.
func autoVpcNNC(name, nsxProject, vpcConnProfile string) *netv1alpha1.NamespaceNetworkConfiguration {
	return &netv1alpha1.NamespaceNetworkConfiguration{
		ObjectMeta: metav1.ObjectMeta{Name: name},
		Spec: netv1alpha1.NamespaceNetworkSpec{
			Type: netv1alpha1.NetworkProviderVPC,
			NamespaceNetworkConfig: netv1alpha1.NamespaceNetworkConfig{
				VPCConfig: netv1alpha1.VPCConfig{
					AutoCreateConfig: netv1alpha1.AutoCreateVPCConfig{
						NSXProject:             nsxProject,
						VPCConnectivityProfile: vpcConnProfile,
					},
				},
			},
		},
	}
}

// -----------------------------------------------------------------------
// Name Validations
// Rule: size(self.metadata.name) <= 63
// -----------------------------------------------------------------------

func TestNamespaceNetworkConfiguration_NameTooLong_Rejected(t *testing.T) {
	obj := vdsNNC(strings.Repeat("a", 64), testNetName)
	if err := k8sClient.Create(testCtx, obj); err == nil || !strings.Contains(err.Error(), "name must be 63 characters or fewer") {
		t.Fatalf("expected rejection containing %q, got: %v", "name must be 63 characters or fewer", err)
	}
}

func TestNamespaceNetworkConfiguration_Name63Chars_Admitted(t *testing.T) {
	obj := vdsNNC(strings.Repeat("a", 63), testNetName)
	if err := k8sClient.Create(testCtx, obj); err != nil {
		t.Fatalf("expected admission, got: %v", err)
	}
	defer func() { _ = k8sClient.Delete(testCtx, obj) }()
}

func TestNamespaceNetworkConfiguration_SingleCharName_Admitted(t *testing.T) {
	obj := vdsNNC("a", testNetName)
	if err := k8sClient.Create(testCtx, obj); err != nil {
		t.Fatalf("expected admission, got: %v", err)
	}
	defer func() { _ = k8sClient.Delete(testCtx, obj) }()
}

// -----------------------------------------------------------------------
// Spec Presence
// -----------------------------------------------------------------------

func TestNamespaceNetworkConfiguration_NoSpec_Admitted(t *testing.T) {
	obj := unstrNNC("test-no-spec", nil)
	if err := k8sClient.Create(testCtx, obj); err != nil {
		t.Fatalf("expected admission, got: %v", err)
	}
	defer func() { _ = k8sClient.Delete(testCtx, obj) }()
}

func TestNamespaceNetworkConfiguration_EmptySpec_Rejected(t *testing.T) {
	obj := unstrNNC("test-empty-spec", map[string]interface{}{})
	if err := k8sClient.Create(testCtx, obj); err == nil || !strings.Contains(err.Error(), "type") {
		t.Fatalf("expected rejection containing %q, got: %v", "type", err)
	}
}

// -----------------------------------------------------------------------
// Type Field Validations
// -----------------------------------------------------------------------

func TestNamespaceNetworkConfiguration_TypeNSXTier1Inherit_Admitted(t *testing.T) {
	nnc := &netv1alpha1.NamespaceNetworkConfiguration{
		ObjectMeta: metav1.ObjectMeta{Name: "test-type-nsx-tier1-inherit"},
		Spec:       netv1alpha1.NamespaceNetworkSpec{Type: netv1alpha1.NetworkProviderNSXTier1},
	}
	if err := k8sClient.Create(testCtx, nnc); err != nil {
		t.Fatalf("expected admission, got: %v", err)
	}
	defer func() { _ = k8sClient.Delete(testCtx, nnc) }()
}

func TestNamespaceNetworkConfiguration_TypeVPCValid_Admitted(t *testing.T) {
	obj := vpcNNC("test-type-vpc-valid", testVPCPathFull)
	if err := k8sClient.Create(testCtx, obj); err != nil {
		t.Fatalf("expected admission, got: %v", err)
	}
	defer func() { _ = k8sClient.Delete(testCtx, obj) }()
}

func TestNamespaceNetworkConfiguration_TypeVPCNoConfig_Rejected(t *testing.T) {
	nnc := &netv1alpha1.NamespaceNetworkConfiguration{
		ObjectMeta: metav1.ObjectMeta{Name: "test-type-vpc-no-config"},
		Spec:       netv1alpha1.NamespaceNetworkSpec{Type: netv1alpha1.NetworkProviderVPC},
	}
	if err := k8sClient.Create(testCtx, nnc); err == nil || !strings.Contains(err.Error(), "vpcConfig must have either vpc") {
		t.Fatalf("expected rejection containing %q, got: %v", "vpcConfig must have either vpc", err)
	}
}

func TestNamespaceNetworkConfiguration_TypeVPCWithVDS_Rejected(t *testing.T) {
	obj := unstrNNC("test-type-vpc-with-vds", map[string]interface{}{
		"type": providerVPC,
		"vsphereDistributedConfig": map[string]interface{}{
			"defaultNetwork": testNetName,
			"networks":       []interface{}{map[string]interface{}{"name": testNetName}},
		},
		"vpcConfig": map[string]interface{}{
			providerVPC: testVPCPathFull,
		},
	})
	if err := k8sClient.Create(testCtx, obj); err == nil || !strings.Contains(err.Error(), "vsphereDistributedConfig must not be populated when type is not vsphere-distributed") {
		t.Fatalf("expected rejection containing %q, got: %v", "vsphereDistributedConfig must not be populated when type is not vsphere-distributed", err)
	}
}

func TestNamespaceNetworkConfiguration_TypeVDSWithVPC_Rejected(t *testing.T) {
	obj := unstrNNC("test-type-vds-with-vpc", map[string]interface{}{
		"type": "vsphere-distributed",
		"vsphereDistributedConfig": map[string]interface{}{
			"defaultNetwork": testNetName,
			"networks":       []interface{}{map[string]interface{}{"name": testNetName}},
		},
		"vpcConfig": map[string]interface{}{
			providerVPC: testVPCPathFull,
		},
	})
	if err := k8sClient.Create(testCtx, obj); err == nil || !strings.Contains(err.Error(), "vpcConfig must not be populated when type is not vpc") {
		t.Fatalf("expected rejection containing %q, got: %v", "vpcConfig must not be populated when type is not vpc", err)
	}
}

func TestNamespaceNetworkConfiguration_TypeUnknownEnum_Rejected(t *testing.T) {
	// NetworkProvider is a plain string alias, so any value can be sent.
	nnc := &netv1alpha1.NamespaceNetworkConfiguration{
		ObjectMeta: metav1.ObjectMeta{Name: "test-type-unknown"},
		Spec:       netv1alpha1.NamespaceNetworkSpec{Type: netv1alpha1.NetworkProvider("unknown-provider")},
	}
	if err := k8sClient.Create(testCtx, nnc); err == nil || !strings.Contains(err.Error(), "Unsupported value") {
		t.Fatalf("expected rejection containing %q, got: %v", "Unsupported value", err)
	}
}

func TestNamespaceNetworkConfiguration_ValidVSphereDistributed_Admitted(t *testing.T) {
	obj := vdsNNC("test-type-valid", testNetName)
	if err := k8sClient.Create(testCtx, obj); err != nil {
		t.Fatalf("expected admission, got: %v", err)
	}
	defer func() { _ = k8sClient.Delete(testCtx, obj) }()
}

func TestNamespaceNetworkConfiguration_VSphereDistributedNoConfig_Rejected(t *testing.T) {
	// CEL: has(self.vsphereDistributedConfig.networks) evaluates to false
	// when the whole vsphereDistributedConfig section is omitted.
	obj := unstrNNC("test-type-no-config", map[string]interface{}{
		"type": "vsphere-distributed",
	})
	if err := k8sClient.Create(testCtx, obj); err == nil || !strings.Contains(err.Error(), "must contain at least one entry") {
		t.Fatalf("expected rejection containing %q, got: %v", "must contain at least one entry", err)
	}
}

// -----------------------------------------------------------------------
// vSphere Distributed Config — field validations
// -----------------------------------------------------------------------

func TestNamespaceNetworkConfiguration_VDSEmptyNetworks_Rejected(t *testing.T) {
	nnc := &netv1alpha1.NamespaceNetworkConfiguration{
		ObjectMeta: metav1.ObjectMeta{Name: "test-vds-empty-nets"},
		Spec: netv1alpha1.NamespaceNetworkSpec{
			Type: netv1alpha1.NetworkProviderVSphereDistributed,
			NamespaceNetworkConfig: netv1alpha1.NamespaceNetworkConfig{
				VSphereDistributedConfig: netv1alpha1.VSphereDistributedConfig{
					DefaultNetwork: testNetName,
					Networks:       []netv1alpha1.VSphereDistributedNetworkRef{},
				},
			},
		},
	}
	// An empty Go slice serializes as absent (omitempty), so the schema
	// required constraint fires before CEL rules are evaluated.
	err := k8sClient.Create(testCtx, nnc)
	if err == nil || (!strings.Contains(err.Error(), "must contain at least one entry") && !strings.Contains(err.Error(), "networks: Required value")) {
		t.Fatalf("expected rejection containing %q or %q, got: %v", "must contain at least one entry", "networks: Required value", err)
	}
}

func TestNamespaceNetworkConfiguration_VDS32Networks_Admitted(t *testing.T) {
	nets := make([]string, 32)
	for i := range nets {
		nets[i] = fmt.Sprintf("net-%02d", i)
	}
	obj := vdsNNC("test-vds-32-nets", nets...)
	if err := k8sClient.Create(testCtx, obj); err != nil {
		t.Fatalf("expected admission, got: %v", err)
	}
	defer func() { _ = k8sClient.Delete(testCtx, obj) }()
}

func TestNamespaceNetworkConfiguration_VDS33Networks_Rejected(t *testing.T) {
	nets := make([]string, 33)
	for i := range nets {
		nets[i] = fmt.Sprintf("net-%02d", i)
	}
	obj := vdsNNC("test-vds-33-nets", nets...)
	if err := k8sClient.Create(testCtx, obj); err == nil || !strings.Contains(err.Error(), "Too many") {
		t.Fatalf("expected rejection containing %q, got: %v", "Too many", err)
	}
}

func TestNamespaceNetworkConfiguration_VDSNetworkRefUppercaseName_Rejected(t *testing.T) {
	nnc := &netv1alpha1.NamespaceNetworkConfiguration{
		ObjectMeta: metav1.ObjectMeta{Name: "test-vds-bad-net-name"},
		Spec: netv1alpha1.NamespaceNetworkSpec{
			Type: netv1alpha1.NetworkProviderVSphereDistributed,
			NamespaceNetworkConfig: netv1alpha1.NamespaceNetworkConfig{
				VSphereDistributedConfig: netv1alpha1.VSphereDistributedConfig{
					DefaultNetwork: testUppercaseName,
					Networks:       []netv1alpha1.VSphereDistributedNetworkRef{{Name: testUppercaseName}},
				},
			},
		},
	}
	if err := k8sClient.Create(testCtx, nnc); err == nil || !strings.Contains(err.Error(), "Invalid value") {
		t.Fatalf("expected rejection containing %q, got: %v", "Invalid value", err)
	}
}

func TestNamespaceNetworkConfiguration_VDSNetworkRefNameTooLong_Rejected(t *testing.T) {
	longName := strings.Repeat("a", 254)
	nnc := &netv1alpha1.NamespaceNetworkConfiguration{
		ObjectMeta: metav1.ObjectMeta{Name: "test-vds-long-net-name"},
		Spec: netv1alpha1.NamespaceNetworkSpec{
			Type: netv1alpha1.NetworkProviderVSphereDistributed,
			NamespaceNetworkConfig: netv1alpha1.NamespaceNetworkConfig{
				VSphereDistributedConfig: netv1alpha1.VSphereDistributedConfig{
					DefaultNetwork: longName,
					Networks:       []netv1alpha1.VSphereDistributedNetworkRef{{Name: longName}},
				},
			},
		},
	}
	if err := k8sClient.Create(testCtx, nnc); err == nil || !strings.Contains(err.Error(), "Too long") {
		t.Fatalf("expected rejection containing %q, got: %v", "Too long", err)
	}
}

func TestNamespaceNetworkConfiguration_VDSDefaultNetworkStartsWithDash_Rejected(t *testing.T) {
	nnc := &netv1alpha1.NamespaceNetworkConfiguration{
		ObjectMeta: metav1.ObjectMeta{Name: "test-vds-bad-default"},
		Spec: netv1alpha1.NamespaceNetworkSpec{
			Type: netv1alpha1.NetworkProviderVSphereDistributed,
			NamespaceNetworkConfig: netv1alpha1.NamespaceNetworkConfig{
				VSphereDistributedConfig: netv1alpha1.VSphereDistributedConfig{
					DefaultNetwork: "-bad-start",
					Networks:       []netv1alpha1.VSphereDistributedNetworkRef{{Name: "good-name"}},
				},
			},
		},
	}
	if err := k8sClient.Create(testCtx, nnc); err == nil || !strings.Contains(err.Error(), "Invalid value") {
		t.Fatalf("expected rejection containing %q, got: %v", "Invalid value", err)
	}
}

func TestNamespaceNetworkConfiguration_VDSDefaultNetworkTooLong_Rejected(t *testing.T) {
	longName := strings.Repeat("a", 254)
	nnc := &netv1alpha1.NamespaceNetworkConfiguration{
		ObjectMeta: metav1.ObjectMeta{Name: "test-vds-long-default"},
		Spec: netv1alpha1.NamespaceNetworkSpec{
			Type: netv1alpha1.NetworkProviderVSphereDistributed,
			NamespaceNetworkConfig: netv1alpha1.NamespaceNetworkConfig{
				VSphereDistributedConfig: netv1alpha1.VSphereDistributedConfig{
					DefaultNetwork: longName,
					Networks:       []netv1alpha1.VSphereDistributedNetworkRef{{Name: "good-name"}},
				},
			},
		},
	}
	if err := k8sClient.Create(testCtx, nnc); err == nil || !strings.Contains(err.Error(), "Too long") {
		t.Fatalf("expected rejection containing %q, got: %v", "Too long", err)
	}
}

func TestNamespaceNetworkConfiguration_VDSDefaultNetworkMismatch_Rejected(t *testing.T) {
	nnc := &netv1alpha1.NamespaceNetworkConfiguration{
		ObjectMeta: metav1.ObjectMeta{Name: "test-vds-default-mismatch"},
		Spec: netv1alpha1.NamespaceNetworkSpec{
			Type: netv1alpha1.NetworkProviderVSphereDistributed,
			NamespaceNetworkConfig: netv1alpha1.NamespaceNetworkConfig{
				VSphereDistributedConfig: netv1alpha1.VSphereDistributedConfig{
					DefaultNetwork: "non-existent",
					Networks:       []netv1alpha1.VSphereDistributedNetworkRef{{Name: "actual-network"}},
				},
			},
		},
	}
	if err := k8sClient.Create(testCtx, nnc); err == nil || !strings.Contains(err.Error(), "defaultNetwork must match") {
		t.Fatalf("expected rejection containing %q, got: %v", "defaultNetwork must match", err)
	}
}

// -----------------------------------------------------------------------
// vSphere Distributed Config — updates and immutability
// -----------------------------------------------------------------------

func TestNamespaceNetworkConfiguration_VDSUpdateChangeDefaultNetwork_Rejected(t *testing.T) {
	obj := vdsNNC("test-upd-change-default", testNamespaceA, testNamespaceB)
	if err := k8sClient.Create(testCtx, obj); err != nil {
		t.Fatalf("create: %v", err)
	}
	defer func() { _ = k8sClient.Delete(testCtx, obj) }()

	fetched := &netv1alpha1.NamespaceNetworkConfiguration{}
	if err := k8sClient.Get(testCtx, client.ObjectKey{Name: "test-upd-change-default"}, fetched); err != nil {
		t.Fatalf("get: %v", err)
	}

	fetched.Spec.VSphereDistributedConfig.DefaultNetwork = testNamespaceB
	if err := k8sClient.Update(testCtx, fetched); err == nil || !strings.Contains(err.Error(), "defaultNetwork is immutable") {
		t.Fatalf("expected rejection containing %q, got: %v", "defaultNetwork is immutable", err)
	}
}

func TestNamespaceNetworkConfiguration_VDSUpdateKeepDefaultNetwork_Admitted(t *testing.T) {
	obj := vdsNNC("test-upd-keep-default", testNamespaceA)
	if err := k8sClient.Create(testCtx, obj); err != nil {
		t.Fatalf("create: %v", err)
	}
	defer func() { _ = k8sClient.Delete(testCtx, obj) }()

	fetched := &netv1alpha1.NamespaceNetworkConfiguration{}
	if err := k8sClient.Get(testCtx, client.ObjectKey{Name: "test-upd-keep-default"}, fetched); err != nil {
		t.Fatalf("get: %v", err)
	}

	// Trigger a meaningful update without changing spec.
	if fetched.Labels == nil {
		fetched.Labels = map[string]string{}
	}
	fetched.Labels["updated"] = testLabelTrue
	if err := k8sClient.Update(testCtx, fetched); err != nil {
		t.Fatalf("expected admission, got: %v", err)
	}
}

func TestNamespaceNetworkConfiguration_VDSUpdateAddNetwork_Admitted(t *testing.T) {
	obj := vdsNNC("test-upd-add-net", testNamespaceA)
	if err := k8sClient.Create(testCtx, obj); err != nil {
		t.Fatalf("create: %v", err)
	}
	defer func() { _ = k8sClient.Delete(testCtx, obj) }()

	fetched := &netv1alpha1.NamespaceNetworkConfiguration{}
	if err := k8sClient.Get(testCtx, client.ObjectKey{Name: "test-upd-add-net"}, fetched); err != nil {
		t.Fatalf("get: %v", err)
	}

	fetched.Spec.VSphereDistributedConfig.Networks = append(
		fetched.Spec.VSphereDistributedConfig.Networks,
		netv1alpha1.VSphereDistributedNetworkRef{Name: testNamespaceB},
	)
	if err := k8sClient.Update(testCtx, fetched); err != nil {
		t.Fatalf("expected admission, got: %v", err)
	}
}

func TestNamespaceNetworkConfiguration_VDSUpdateRemoveDefaultNetwork_Rejected(t *testing.T) {
	obj := vdsNNC("test-upd-remove-default", testNamespaceA, testNamespaceB)
	if err := k8sClient.Create(testCtx, obj); err != nil {
		t.Fatalf("create: %v", err)
	}
	defer func() { _ = k8sClient.Delete(testCtx, obj) }()

	fetched := &netv1alpha1.NamespaceNetworkConfiguration{}
	if err := k8sClient.Get(testCtx, client.ObjectKey{Name: "test-upd-remove-default"}, fetched); err != nil {
		t.Fatalf("get: %v", err)
	}

	// Remove net-a (the defaultNetwork), leaving only net-b.
	fetched.Spec.VSphereDistributedConfig.Networks = []netv1alpha1.VSphereDistributedNetworkRef{
		{Name: testNamespaceB},
	}
	if err := k8sClient.Update(testCtx, fetched); err == nil || !strings.Contains(err.Error(), "defaultNetwork must match") {
		t.Fatalf("expected rejection containing %q, got: %v", "defaultNetwork must match", err)
	}
}

// -----------------------------------------------------------------------
// VPC Config — field validations (pre-created VPC)
// -----------------------------------------------------------------------

func TestNamespaceNetworkConfiguration_VPCBothModes_Rejected(t *testing.T) {
	obj := unstrNNC("test-vpc-both-modes", map[string]interface{}{
		"type": providerVPC,
		"vpcConfig": map[string]interface{}{
			providerVPC: testVPCPathFull,
			"autoCreateConfig": map[string]interface{}{
				"nsxProject":             testNSXProject,
				"vpcConnectivityProfile": testVPCConnProfile,
			},
		},
	})
	if err := k8sClient.Create(testCtx, obj); err == nil || !strings.Contains(err.Error(), "vpc and autoCreateConfig are mutually exclusive") {
		t.Fatalf("expected rejection containing %q, got: %v", "vpc and autoCreateConfig are mutually exclusive", err)
	}
}

func TestNamespaceNetworkConfiguration_VPCDefaultSubnetSize1_Admitted(t *testing.T) {
	obj := vpcNNC("test-vpc-dss-1", testVPCPath)
	obj.Spec.VPCConfig.DefaultSubnetSize = 1
	if err := k8sClient.Create(testCtx, obj); err != nil {
		t.Fatalf("expected admission, got: %v", err)
	}
	defer func() { _ = k8sClient.Delete(testCtx, obj) }()
}

func TestNamespaceNetworkConfiguration_VPCDefaultSubnetSize65536_Admitted(t *testing.T) {
	obj := vpcNNC("test-vpc-dss-65536", testVPCPath)
	obj.Spec.VPCConfig.DefaultSubnetSize = 65536
	if err := k8sClient.Create(testCtx, obj); err != nil {
		t.Fatalf("expected admission, got: %v", err)
	}
	defer func() { _ = k8sClient.Delete(testCtx, obj) }()
}

func TestNamespaceNetworkConfiguration_VPCDefaultSubnetSize32_Admitted(t *testing.T) {
	obj := vpcNNC("test-vpc-dss-32", testVPCPath)
	obj.Spec.VPCConfig.DefaultSubnetSize = 32
	if err := k8sClient.Create(testCtx, obj); err != nil {
		t.Fatalf("expected admission, got: %v", err)
	}
	defer func() { _ = k8sClient.Delete(testCtx, obj) }()
}

func TestNamespaceNetworkConfiguration_VPCDefaultSubnetSize0_Rejected(t *testing.T) {
	// Use unstructured: Go int32 zero-value is omitted by omitempty,
	// so we send it explicitly via the unstructured path.
	obj := unstrNNC("test-vpc-dss-0", map[string]interface{}{
		"type": providerVPC,
		"vpcConfig": map[string]interface{}{
			providerVPC:         testVPCPath,
			"defaultSubnetSize": int64(0),
		},
	})
	if err := k8sClient.Create(testCtx, obj); err == nil || !strings.Contains(err.Error(), "defaultSubnetSize") {
		t.Fatalf("expected rejection containing %q, got: %v", "defaultSubnetSize", err)
	}
}

func TestNamespaceNetworkConfiguration_VPCDefaultSubnetSize65537_Rejected(t *testing.T) {
	obj := unstrNNC("test-vpc-dss-65537", map[string]interface{}{
		"type": providerVPC,
		"vpcConfig": map[string]interface{}{
			providerVPC:         testVPCPath,
			"defaultSubnetSize": int64(65537),
		},
	})
	if err := k8sClient.Create(testCtx, obj); err == nil || !strings.Contains(err.Error(), "defaultSubnetSize") {
		t.Fatalf("expected rejection containing %q, got: %v", "defaultSubnetSize", err)
	}
}

func TestNamespaceNetworkConfiguration_VPCDefaultSubnetSize3_Rejected(t *testing.T) {
	obj := unstrNNC("test-vpc-dss-3", map[string]interface{}{
		"type": providerVPC,
		"vpcConfig": map[string]interface{}{
			providerVPC:         testVPCPath,
			"defaultSubnetSize": int64(3),
		},
	})
	if err := k8sClient.Create(testCtx, obj); err == nil || !strings.Contains(err.Error(), "defaultSubnetSize must be a power of 2") {
		t.Fatalf("expected rejection containing %q, got: %v", "defaultSubnetSize must be a power of 2", err)
	}
}

func TestNamespaceNetworkConfiguration_VPCDefaultSubnetSize1000_Rejected(t *testing.T) {
	obj := unstrNNC("test-vpc-dss-1000", map[string]interface{}{
		"type": providerVPC,
		"vpcConfig": map[string]interface{}{
			providerVPC:         testVPCPath,
			"defaultSubnetSize": int64(1000),
		},
	})
	if err := k8sClient.Create(testCtx, obj); err == nil || !strings.Contains(err.Error(), "defaultSubnetSize must be a power of 2") {
		t.Fatalf("expected rejection containing %q, got: %v", "defaultSubnetSize must be a power of 2", err)
	}
}

func TestNamespaceNetworkConfiguration_VPCDefaultSubnetSize131072_Rejected(t *testing.T) {
	obj := unstrNNC("test-vpc-dss-131072", map[string]interface{}{
		"type": providerVPC,
		"vpcConfig": map[string]interface{}{
			providerVPC:         testVPCPath,
			"defaultSubnetSize": int64(131072),
		},
	})
	if err := k8sClient.Create(testCtx, obj); err == nil || !strings.Contains(err.Error(), "defaultSubnetSize") {
		t.Fatalf("expected rejection containing %q, got: %v", "defaultSubnetSize", err)
	}
}

func TestNamespaceNetworkConfiguration_VPC32SharedSubnets_Admitted(t *testing.T) {
	obj := vpcNNC("test-vpc-32-subnets", testVPCPath)
	subnets := make([]netv1alpha1.SharedSubnet, 32)
	for i := range subnets {
		subnets[i] = netv1alpha1.SharedSubnet{
			Path: fmt.Sprintf("/infra/subnets/s%02d", i),
			Name: fmt.Sprintf("subnet-%02d", i),
		}
	}
	obj.Spec.VPCConfig.SharedSubnets = subnets
	if err := k8sClient.Create(testCtx, obj); err != nil {
		t.Fatalf("expected admission, got: %v", err)
	}
	defer func() { _ = k8sClient.Delete(testCtx, obj) }()
}

func TestNamespaceNetworkConfiguration_VPC33SharedSubnets_Rejected(t *testing.T) {
	obj := vpcNNC("test-vpc-33-subnets", testVPCPath)
	subnets := make([]netv1alpha1.SharedSubnet, 33)
	for i := range subnets {
		subnets[i] = netv1alpha1.SharedSubnet{
			Path: fmt.Sprintf("/infra/subnets/s%02d", i),
			Name: fmt.Sprintf("subnet-%02d", i),
		}
	}
	obj.Spec.VPCConfig.SharedSubnets = subnets
	if err := k8sClient.Create(testCtx, obj); err == nil || !strings.Contains(err.Error(), "Too many") {
		t.Fatalf("expected rejection containing %q, got: %v", "Too many", err)
	}
}

// -----------------------------------------------------------------------
// VPC Config — SharedSubnet field validations
// -----------------------------------------------------------------------

func TestNamespaceNetworkConfiguration_SharedSubnetUppercaseName_Rejected(t *testing.T) {
	obj := vpcNNC("test-vpc-subnet-bad-name", testVPCPath)
	obj.Spec.VPCConfig.SharedSubnets = []netv1alpha1.SharedSubnet{
		{Path: testSubnetPath1, Name: testUppercaseName},
	}
	if err := k8sClient.Create(testCtx, obj); err == nil || !strings.Contains(err.Error(), "Invalid value") {
		t.Fatalf("expected rejection containing %q, got: %v", "Invalid value", err)
	}
}

func TestNamespaceNetworkConfiguration_SharedSubnetNameTooLong_Rejected(t *testing.T) {
	obj := vpcNNC("test-vpc-subnet-long-name", testVPCPath)
	obj.Spec.VPCConfig.SharedSubnets = []netv1alpha1.SharedSubnet{
		{Path: testSubnetPath1, Name: strings.Repeat("a", 254)},
	}
	if err := k8sClient.Create(testCtx, obj); err == nil || !strings.Contains(err.Error(), "Too long") {
		t.Fatalf("expected rejection containing %q, got: %v", "Too long", err)
	}
}

func TestNamespaceNetworkConfiguration_SharedSubnetName253Chars_Admitted(t *testing.T) {
	// 63+1+63+1+63+1+61 = 253, all valid DNS-1123 segments
	longName := strings.Repeat("a", 63) + "." + strings.Repeat("b", 63) + "." +
		strings.Repeat("c", 63) + "." + strings.Repeat("d", 61)
	obj := vpcNNC("test-vpc-subnet-253-name", testVPCPath)
	obj.Spec.VPCConfig.SharedSubnets = []netv1alpha1.SharedSubnet{
		{Path: testSubnetPath1, Name: longName},
	}
	if err := k8sClient.Create(testCtx, obj); err != nil {
		t.Fatalf("expected admission, got: %v", err)
	}
	defer func() { _ = k8sClient.Delete(testCtx, obj) }()
}

func TestNamespaceNetworkConfiguration_SharedSubnetPathTooLong_Rejected(t *testing.T) {
	obj := vpcNNC("test-vpc-subnet-long-path", testVPCPath)
	obj.Spec.VPCConfig.SharedSubnets = []netv1alpha1.SharedSubnet{
		{Path: "/" + strings.Repeat("a", 2048), Name: testSubnetNameA},
	}
	if err := k8sClient.Create(testCtx, obj); err == nil || !strings.Contains(err.Error(), "Too long") {
		t.Fatalf("expected rejection containing %q, got: %v", "Too long", err)
	}
}

func TestNamespaceNetworkConfiguration_SharedSubnetPath2048Chars_Admitted(t *testing.T) {
	obj := vpcNNC("test-vpc-subnet-2048-path", testVPCPath)
	obj.Spec.VPCConfig.SharedSubnets = []netv1alpha1.SharedSubnet{
		{Path: "/" + strings.Repeat("a", 2047), Name: testSubnetNameA},
	}
	if err := k8sClient.Create(testCtx, obj); err != nil {
		t.Fatalf("expected admission, got: %v", err)
	}
	defer func() { _ = k8sClient.Delete(testCtx, obj) }()
}

func TestNamespaceNetworkConfiguration_DuplicateSharedSubnetNames_Rejected(t *testing.T) {
	// Must use unstructured — Go deduplicates by list-map key on encoding.
	obj := unstrNNC("test-vpc-dup-subnet-name", map[string]interface{}{
		"type": providerVPC,
		"vpcConfig": map[string]interface{}{
			providerVPC: testVPCPath,
			"sharedSubnets": []interface{}{
				map[string]interface{}{"path": testSubnetPath1, "name": testSubnetNameA},
				map[string]interface{}{"path": testSubnetPath2, "name": testSubnetNameA},
			},
		},
	})
	if err := k8sClient.Create(testCtx, obj); err == nil || !strings.Contains(err.Error(), "uplicate") {
		t.Fatalf("expected rejection containing %q, got: %v", "uplicate", err)
	}
}

func TestNamespaceNetworkConfiguration_SharedSubnetPodDefaultTrueVMDefaultFalse_Admitted(t *testing.T) {
	obj := vpcNNC("test-vpc-subnet-defaults-valid", testVPCPath)
	obj.Spec.VPCConfig.SharedSubnets = []netv1alpha1.SharedSubnet{
		{
			Path:       testSubnetPath1,
			Name:       testSubnetNameA,
			PodDefault: netv1alpha1.SharedSubnetDefaultTrue,
			VMDefault:  netv1alpha1.SharedSubnetDefaultFalse,
		},
	}
	if err := k8sClient.Create(testCtx, obj); err != nil {
		t.Fatalf("expected admission, got: %v", err)
	}
	defer func() { _ = k8sClient.Delete(testCtx, obj) }()
}

func TestNamespaceNetworkConfiguration_SharedSubnetInvalidPodDefaultEnum_Rejected(t *testing.T) {
	obj := unstrNNC("test-vpc-subnet-bad-pod-default", map[string]interface{}{
		"type": providerVPC,
		"vpcConfig": map[string]interface{}{
			providerVPC: testVPCPath,
			"sharedSubnets": []interface{}{
				map[string]interface{}{
					"path":       testSubnetPath1,
					"name":       testSubnetNameA,
					"podDefault": "yes",
				},
			},
		},
	})
	if err := k8sClient.Create(testCtx, obj); err == nil || !strings.Contains(err.Error(), "Unsupported value") {
		t.Fatalf("expected rejection containing %q, got: %v", "Unsupported value", err)
	}
}

func TestNamespaceNetworkConfiguration_SharedSubnetInvalidVMDefaultEnum_Rejected(t *testing.T) {
	obj := unstrNNC("test-vpc-subnet-bad-vm-default", map[string]interface{}{
		"type": providerVPC,
		"vpcConfig": map[string]interface{}{
			providerVPC: testVPCPath,
			"sharedSubnets": []interface{}{
				map[string]interface{}{
					"path":      testSubnetPath1,
					"name":      testSubnetNameA,
					"vmDefault": testLabelTrue,
				},
			},
		},
	})
	if err := k8sClient.Create(testCtx, obj); err == nil || !strings.Contains(err.Error(), "Unsupported value") {
		t.Fatalf("expected rejection containing %q, got: %v", "Unsupported value", err)
	}
}

func TestNamespaceNetworkConfiguration_TwoSharedSubnetsPodDefaultTrue_Rejected(t *testing.T) {
	obj := vpcNNC("test-vpc-two-pod-defaults", testVPCPath)
	obj.Spec.VPCConfig.SharedSubnets = []netv1alpha1.SharedSubnet{
		{Path: testSubnetPath1, Name: testSubnetNameA, PodDefault: netv1alpha1.SharedSubnetDefaultTrue},
		{Path: testSubnetPath2, Name: testSubnetNameB, PodDefault: netv1alpha1.SharedSubnetDefaultTrue},
	}
	if err := k8sClient.Create(testCtx, obj); err == nil || !strings.Contains(err.Error(), "at most one sharedSubnet may have podDefault set to True") {
		t.Fatalf("expected rejection containing %q, got: %v", "at most one sharedSubnet may have podDefault set to True", err)
	}
}

func TestNamespaceNetworkConfiguration_TwoSharedSubnetsVMDefaultTrue_Rejected(t *testing.T) {
	obj := vpcNNC("test-vpc-two-vm-defaults", testVPCPath)
	obj.Spec.VPCConfig.SharedSubnets = []netv1alpha1.SharedSubnet{
		{Path: testSubnetPath1, Name: testSubnetNameA, VMDefault: netv1alpha1.SharedSubnetDefaultTrue},
		{Path: testSubnetPath2, Name: testSubnetNameB, VMDefault: netv1alpha1.SharedSubnetDefaultTrue},
	}
	if err := k8sClient.Create(testCtx, obj); err == nil || !strings.Contains(err.Error(), "at most one sharedSubnet may have vmDefault set to True") {
		t.Fatalf("expected rejection containing %q, got: %v", "at most one sharedSubnet may have vmDefault set to True", err)
	}
}

func TestNamespaceNetworkConfiguration_OneEachSharedSubnetDefault_Admitted(t *testing.T) {
	obj := vpcNNC("test-vpc-one-each-default", testVPCPath)
	obj.Spec.VPCConfig.SharedSubnets = []netv1alpha1.SharedSubnet{
		{Path: testSubnetPath1, Name: testSubnetNameA, PodDefault: netv1alpha1.SharedSubnetDefaultTrue},
		{Path: testSubnetPath2, Name: testSubnetNameB, VMDefault: netv1alpha1.SharedSubnetDefaultTrue},
	}
	if err := k8sClient.Create(testCtx, obj); err != nil {
		t.Fatalf("expected admission, got: %v", err)
	}
	defer func() { _ = k8sClient.Delete(testCtx, obj) }()
}

func TestNamespaceNetworkConfiguration_OneSharedSubnetBothDefaults_Admitted(t *testing.T) {
	obj := vpcNNC("test-vpc-both-defaults-one", testVPCPath)
	obj.Spec.VPCConfig.SharedSubnets = []netv1alpha1.SharedSubnet{
		{
			Path:       testSubnetPath1,
			Name:       testSubnetNameA,
			PodDefault: netv1alpha1.SharedSubnetDefaultTrue,
			VMDefault:  netv1alpha1.SharedSubnetDefaultTrue,
		},
	}
	if err := k8sClient.Create(testCtx, obj); err != nil {
		t.Fatalf("expected admission, got: %v", err)
	}
	defer func() { _ = k8sClient.Delete(testCtx, obj) }()
}

func TestNamespaceNetworkConfiguration_AutoCreatePodDefaultTrue_Rejected(t *testing.T) {
	nnc := autoVpcNNC("test-vpc-auto-pod-default", testNSXProject, testVPCConnProfile)
	nnc.Spec.VPCConfig.SharedSubnets = []netv1alpha1.SharedSubnet{
		{Path: testSubnetPath1, Name: testSubnetNameA, PodDefault: netv1alpha1.SharedSubnetDefaultTrue},
	}
	if err := k8sClient.Create(testCtx, nnc); err == nil || !strings.Contains(err.Error(), "vpc must be set when any sharedSubnet has podDefault set to True") {
		t.Fatalf("expected rejection containing %q, got: %v", "vpc must be set when any sharedSubnet has podDefault set to True", err)
	}
}

func TestNamespaceNetworkConfiguration_AutoCreateVMDefaultTrue_Rejected(t *testing.T) {
	nnc := autoVpcNNC("test-vpc-auto-vm-default", testNSXProject, testVPCConnProfile)
	nnc.Spec.VPCConfig.SharedSubnets = []netv1alpha1.SharedSubnet{
		{Path: testSubnetPath1, Name: testSubnetNameA, VMDefault: netv1alpha1.SharedSubnetDefaultTrue},
	}
	if err := k8sClient.Create(testCtx, nnc); err == nil || !strings.Contains(err.Error(), "vpc must be set when any sharedSubnet has vmDefault set to True") {
		t.Fatalf("expected rejection containing %q, got: %v", "vpc must be set when any sharedSubnet has vmDefault set to True", err)
	}
}

func TestNamespaceNetworkConfiguration_AutoCreateSharedSubnetsNoDefaults_Admitted(t *testing.T) {
	nnc := autoVpcNNC("test-vpc-auto-no-defaults", testNSXProject, testVPCConnProfile)
	nnc.Spec.VPCConfig.SharedSubnets = []netv1alpha1.SharedSubnet{
		{Path: testSubnetPath1, Name: testSubnetNameA},
		{Path: testSubnetPath2, Name: testSubnetNameB},
	}
	if err := k8sClient.Create(testCtx, nnc); err != nil {
		t.Fatalf("expected admission, got: %v", err)
	}
	defer func() { _ = k8sClient.Delete(testCtx, nnc) }()
}

func TestNamespaceNetworkConfiguration_AutoCreateSharedSubnetExplicitFalseDefaults_Admitted(t *testing.T) {
	nnc := autoVpcNNC("test-vpc-auto-explicit-false", testNSXProject, testVPCConnProfile)
	nnc.Spec.VPCConfig.SharedSubnets = []netv1alpha1.SharedSubnet{
		{
			Path:       testSubnetPath1,
			Name:       testSubnetNameA,
			PodDefault: netv1alpha1.SharedSubnetDefaultFalse,
			VMDefault:  netv1alpha1.SharedSubnetDefaultFalse,
		},
	}
	if err := k8sClient.Create(testCtx, nnc); err != nil {
		t.Fatalf("expected admission, got: %v", err)
	}
	defer func() { _ = k8sClient.Delete(testCtx, nnc) }()
}

// -----------------------------------------------------------------------
// VPC Config — AutoCreateVPCConfig field validations
// -----------------------------------------------------------------------

func TestNamespaceNetworkConfiguration_AutoCreateMinimal_Admitted(t *testing.T) {
	obj := autoVpcNNC("test-vpc-auto-minimal", testNSXProject, testVPCConnProfile)
	if err := k8sClient.Create(testCtx, obj); err != nil {
		t.Fatalf("expected admission, got: %v", err)
	}
	defer func() { _ = k8sClient.Delete(testCtx, obj) }()
}

func TestNamespaceNetworkConfiguration_AutoCreateNoNSXProject_Rejected(t *testing.T) {
	obj := unstrNNC("test-vpc-auto-no-project", map[string]interface{}{
		"type": providerVPC,
		"vpcConfig": map[string]interface{}{
			"autoCreateConfig": map[string]interface{}{
				"vpcConnectivityProfile": testVPCConnProfile,
			},
		},
	})
	if err := k8sClient.Create(testCtx, obj); err == nil || !strings.Contains(err.Error(), "nsxProject") {
		t.Fatalf("expected rejection containing %q, got: %v", "nsxProject", err)
	}
}

func TestNamespaceNetworkConfiguration_AutoCreateNoVPCConnProfile_Rejected(t *testing.T) {
	obj := unstrNNC("test-vpc-auto-no-profile", map[string]interface{}{
		"type": providerVPC,
		"vpcConfig": map[string]interface{}{
			"autoCreateConfig": map[string]interface{}{
				"nsxProject": testNSXProject,
			},
		},
	})
	if err := k8sClient.Create(testCtx, obj); err == nil || !strings.Contains(err.Error(), "vpcConnectivityProfile") {
		t.Fatalf("expected rejection containing %q, got: %v", "vpcConnectivityProfile", err)
	}
}

func TestNamespaceNetworkConfiguration_AutoCreate16PrivateCIDRs_Admitted(t *testing.T) {
	nnc := autoVpcNNC("test-vpc-auto-16-cidrs", testNSXProject, testVPCConnProfile)
	cidrs := make([]string, 16)
	for i := range cidrs {
		cidrs[i] = fmt.Sprintf("10.%d.0.0/24", i)
	}
	nnc.Spec.VPCConfig.AutoCreateConfig.PrivateCIDRs = cidrs
	if err := k8sClient.Create(testCtx, nnc); err != nil {
		t.Fatalf("expected admission, got: %v", err)
	}
	defer func() { _ = k8sClient.Delete(testCtx, nnc) }()
}

func TestNamespaceNetworkConfiguration_AutoCreate17PrivateCIDRs_Rejected(t *testing.T) {
	nnc := autoVpcNNC("test-vpc-auto-17-cidrs", testNSXProject, testVPCConnProfile)
	cidrs := make([]string, 17)
	for i := range cidrs {
		cidrs[i] = fmt.Sprintf("10.%d.0.0/24", i)
	}
	nnc.Spec.VPCConfig.AutoCreateConfig.PrivateCIDRs = cidrs
	if err := k8sClient.Create(testCtx, nnc); err == nil || !strings.Contains(err.Error(), "Too many") {
		t.Fatalf("expected rejection containing %q, got: %v", "Too many", err)
	}
}

func TestNamespaceNetworkConfiguration_AutoCreatePrivateCIDRNotCIDR_Rejected(t *testing.T) {
	obj := unstrNNC("test-vpc-auto-bad-cidr", map[string]interface{}{
		"type": providerVPC,
		"vpcConfig": map[string]interface{}{
			"autoCreateConfig": map[string]interface{}{
				"nsxProject":             testNSXProject,
				"vpcConnectivityProfile": testVPCConnProfile,
				"privateCIDRs":           []interface{}{"not-a-cidr"},
			},
		},
	})
	if err := k8sClient.Create(testCtx, obj); err == nil || !strings.Contains(err.Error(), "privateCIDRs") {
		t.Fatalf("expected rejection containing %q, got: %v", "privateCIDRs", err)
	}
}

// -----------------------------------------------------------------------
// VPC Config — updates and immutability
// -----------------------------------------------------------------------

func TestNamespaceNetworkConfiguration_VPCPathUpdate_Rejected(t *testing.T) {
	obj := vpcNNC("test-vpc-imm-vpc-path", testVPCPath)
	if err := k8sClient.Create(testCtx, obj); err != nil {
		t.Fatalf("create: %v", err)
	}
	defer func() { _ = k8sClient.Delete(testCtx, obj) }()

	fetched := &netv1alpha1.NamespaceNetworkConfiguration{}
	if err := k8sClient.Get(testCtx, client.ObjectKey{Name: "test-vpc-imm-vpc-path"}, fetched); err != nil {
		t.Fatalf("get: %v", err)
	}

	fetched.Spec.VPCConfig.VPC = "/infra/vpcs/v2"
	if err := k8sClient.Update(testCtx, fetched); err == nil || !strings.Contains(err.Error(), "vpc is immutable once set") {
		t.Fatalf("expected rejection containing %q, got: %v", "vpc is immutable once set", err)
	}
}

func TestNamespaceNetworkConfiguration_VPCPathUnchangedUpdate_Admitted(t *testing.T) {
	obj := vpcNNC("test-vpc-imm-vpc-same", testVPCPath)
	if err := k8sClient.Create(testCtx, obj); err != nil {
		t.Fatalf("create: %v", err)
	}
	defer func() { _ = k8sClient.Delete(testCtx, obj) }()

	fetched := &netv1alpha1.NamespaceNetworkConfiguration{}
	if err := k8sClient.Get(testCtx, client.ObjectKey{Name: "test-vpc-imm-vpc-same"}, fetched); err != nil {
		t.Fatalf("get: %v", err)
	}

	if fetched.Labels == nil {
		fetched.Labels = map[string]string{}
	}
	fetched.Labels["touched"] = testLabelTrue
	if err := k8sClient.Update(testCtx, fetched); err != nil {
		t.Fatalf("expected admission, got: %v", err)
	}
}

func TestNamespaceNetworkConfiguration_VPCModeSwitchToAutoCreate_Rejected(t *testing.T) {
	obj := vpcNNC("test-vpc-mode-switch", testVPCPath)
	if err := k8sClient.Create(testCtx, obj); err != nil {
		t.Fatalf("create: %v", err)
	}
	defer func() { _ = k8sClient.Delete(testCtx, obj) }()

	fetched := &netv1alpha1.NamespaceNetworkConfiguration{}
	if err := k8sClient.Get(testCtx, client.ObjectKey{Name: "test-vpc-mode-switch"}, fetched); err != nil {
		t.Fatalf("get: %v", err)
	}

	// Clear vpc and set autoCreateConfig — immutability rule fires because oldSelf.vpc was set.
	fetched.Spec.VPCConfig.VPC = ""
	fetched.Spec.VPCConfig.AutoCreateConfig = netv1alpha1.AutoCreateVPCConfig{
		NSXProject:             testNSXProject,
		VPCConnectivityProfile: testVPCConnProfile,
	}
	if err := k8sClient.Update(testCtx, fetched); err == nil || !strings.Contains(err.Error(), "vpc is immutable once set") {
		t.Fatalf("expected rejection containing %q, got: %v", "vpc is immutable once set", err)
	}
}

func TestNamespaceNetworkConfiguration_AutoCreateNSXProjectUpdate_Rejected(t *testing.T) {
	obj := autoVpcNNC("test-vpc-imm-project", testNSXProject, testVPCConnProfile)
	if err := k8sClient.Create(testCtx, obj); err != nil {
		t.Fatalf("create: %v", err)
	}
	defer func() { _ = k8sClient.Delete(testCtx, obj) }()

	fetched := &netv1alpha1.NamespaceNetworkConfiguration{}
	if err := k8sClient.Get(testCtx, client.ObjectKey{Name: "test-vpc-imm-project"}, fetched); err != nil {
		t.Fatalf("get: %v", err)
	}

	fetched.Spec.VPCConfig.AutoCreateConfig.NSXProject = "/infra/orgs/default/projects/p2"
	if err := k8sClient.Update(testCtx, fetched); err == nil || !strings.Contains(err.Error(), "nsxProject is immutable once set") {
		t.Fatalf("expected rejection containing %q, got: %v", "nsxProject is immutable once set", err)
	}
}

func TestNamespaceNetworkConfiguration_AutoCreateVPCConnProfileUpdate_Rejected(t *testing.T) {
	obj := autoVpcNNC("test-vpc-imm-profile", testNSXProject, testVPCConnProfile)
	if err := k8sClient.Create(testCtx, obj); err != nil {
		t.Fatalf("create: %v", err)
	}
	defer func() { _ = k8sClient.Delete(testCtx, obj) }()

	fetched := &netv1alpha1.NamespaceNetworkConfiguration{}
	if err := k8sClient.Get(testCtx, client.ObjectKey{Name: "test-vpc-imm-profile"}, fetched); err != nil {
		t.Fatalf("get: %v", err)
	}

	fetched.Spec.VPCConfig.AutoCreateConfig.VPCConnectivityProfile = "/infra/vpc-conn-profiles/other"
	if err := k8sClient.Update(testCtx, fetched); err == nil || !strings.Contains(err.Error(), "vpcConnectivityProfile is immutable once set") {
		t.Fatalf("expected rejection containing %q, got: %v", "vpcConnectivityProfile is immutable once set", err)
	}
}

func TestNamespaceNetworkConfiguration_PrivateCIDRsAppend_Admitted(t *testing.T) {
	nnc := autoVpcNNC("test-vpc-cidr-append", testNSXProject, testVPCConnProfile)
	nnc.Spec.VPCConfig.AutoCreateConfig.PrivateCIDRs = []string{testCIDR1}
	if err := k8sClient.Create(testCtx, nnc); err != nil {
		t.Fatalf("create: %v", err)
	}
	defer func() { _ = k8sClient.Delete(testCtx, nnc) }()

	fetched := &netv1alpha1.NamespaceNetworkConfiguration{}
	if err := k8sClient.Get(testCtx, client.ObjectKey{Name: "test-vpc-cidr-append"}, fetched); err != nil {
		t.Fatalf("get: %v", err)
	}

	fetched.Spec.VPCConfig.AutoCreateConfig.PrivateCIDRs = append(
		fetched.Spec.VPCConfig.AutoCreateConfig.PrivateCIDRs,
		"10.1.0.0/24",
	)
	if err := k8sClient.Update(testCtx, fetched); err != nil {
		t.Fatalf("expected admission, got: %v", err)
	}
}

func TestNamespaceNetworkConfiguration_PrivateCIDRsRemoval_Rejected(t *testing.T) {
	nnc := autoVpcNNC("test-vpc-cidr-remove", testNSXProject, testVPCConnProfile)
	nnc.Spec.VPCConfig.AutoCreateConfig.PrivateCIDRs = []string{testCIDR1, "10.1.0.0/24"}
	if err := k8sClient.Create(testCtx, nnc); err != nil {
		t.Fatalf("create: %v", err)
	}
	defer func() { _ = k8sClient.Delete(testCtx, nnc) }()

	fetched := &netv1alpha1.NamespaceNetworkConfiguration{}
	if err := k8sClient.Get(testCtx, client.ObjectKey{Name: "test-vpc-cidr-remove"}, fetched); err != nil {
		t.Fatalf("get: %v", err)
	}

	fetched.Spec.VPCConfig.AutoCreateConfig.PrivateCIDRs = []string{"10.1.0.0/24"}
	if err := k8sClient.Update(testCtx, fetched); err == nil || !strings.Contains(err.Error(), "privateCIDRs is append-only") {
		t.Fatalf("expected rejection containing %q, got: %v", "privateCIDRs is append-only", err)
	}
}

func TestNamespaceNetworkConfiguration_PrivateCIDRsReplace_Rejected(t *testing.T) {
	nnc := autoVpcNNC("test-vpc-cidr-replace", testNSXProject, testVPCConnProfile)
	nnc.Spec.VPCConfig.AutoCreateConfig.PrivateCIDRs = []string{testCIDR1}
	if err := k8sClient.Create(testCtx, nnc); err != nil {
		t.Fatalf("create: %v", err)
	}
	defer func() { _ = k8sClient.Delete(testCtx, nnc) }()

	fetched := &netv1alpha1.NamespaceNetworkConfiguration{}
	if err := k8sClient.Get(testCtx, client.ObjectKey{Name: "test-vpc-cidr-replace"}, fetched); err != nil {
		t.Fatalf("get: %v", err)
	}

	fetched.Spec.VPCConfig.AutoCreateConfig.PrivateCIDRs = []string{"192.168.0.0/24"}
	if err := k8sClient.Update(testCtx, fetched); err == nil || !strings.Contains(err.Error(), "privateCIDRs is append-only") {
		t.Fatalf("expected rejection containing %q, got: %v", "privateCIDRs is append-only", err)
	}
}

func TestNamespaceNetworkConfiguration_SharedSubnetPathUpdate_Rejected(t *testing.T) {
	nnc := vpcNNC("test-vpc-subnet-imm-path", testVPCPath)
	nnc.Spec.VPCConfig.SharedSubnets = []netv1alpha1.SharedSubnet{
		{Path: testSubnetPath1, Name: testSubnetNameA},
	}
	if err := k8sClient.Create(testCtx, nnc); err != nil {
		t.Fatalf("create: %v", err)
	}
	defer func() { _ = k8sClient.Delete(testCtx, nnc) }()

	fetched := &netv1alpha1.NamespaceNetworkConfiguration{}
	if err := k8sClient.Get(testCtx, client.ObjectKey{Name: "test-vpc-subnet-imm-path"}, fetched); err != nil {
		t.Fatalf("get: %v", err)
	}

	fetched.Spec.VPCConfig.SharedSubnets[0].Path = "/infra/subnets/s99"
	if err := k8sClient.Update(testCtx, fetched); err == nil || !strings.Contains(err.Error(), "path is immutable once set") {
		t.Fatalf("expected rejection containing %q, got: %v", "path is immutable once set", err)
	}
}

func TestNamespaceNetworkConfiguration_SharedSubnetRenameViaListMapKey_Admitted(t *testing.T) {
	// listType=map means changing 'name' (the map key) is treated as deleting
	// the old entry and inserting a new one, not as mutating the existing entry.
	// The per-entry CEL rule '!has(oldSelf.name) || self.name == oldSelf.name'
	// therefore never fires — the new entry has no oldSelf. This test documents
	// this behavior explicitly so future readers know the rename is intentionally
	// allowed at the CRD layer (business logic must enforce it elsewhere).
	nnc := vpcNNC("test-vpc-subnet-rename", testVPCPath)
	nnc.Spec.VPCConfig.SharedSubnets = []netv1alpha1.SharedSubnet{
		{Path: testSubnetPath1, Name: testSubnetNameA},
	}
	if err := k8sClient.Create(testCtx, nnc); err != nil {
		t.Fatalf("create: %v", err)
	}
	defer func() { _ = k8sClient.Delete(testCtx, nnc) }()

	fetched := &netv1alpha1.NamespaceNetworkConfiguration{}
	if err := k8sClient.Get(testCtx, client.ObjectKey{Name: "test-vpc-subnet-rename"}, fetched); err != nil {
		t.Fatalf("get: %v", err)
	}

	fetched.Spec.VPCConfig.SharedSubnets[0].Name = testSubnetNameB
	if err := k8sClient.Update(testCtx, fetched); err != nil {
		t.Fatalf("expected admission, got: %v", err)
	}
}

func TestNamespaceNetworkConfiguration_SharedSubnetAdd_Admitted(t *testing.T) {
	nnc := vpcNNC("test-vpc-subnet-add", testVPCPath)
	nnc.Spec.VPCConfig.SharedSubnets = []netv1alpha1.SharedSubnet{
		{Path: testSubnetPath1, Name: testSubnetNameA},
	}
	if err := k8sClient.Create(testCtx, nnc); err != nil {
		t.Fatalf("create: %v", err)
	}
	defer func() { _ = k8sClient.Delete(testCtx, nnc) }()

	fetched := &netv1alpha1.NamespaceNetworkConfiguration{}
	if err := k8sClient.Get(testCtx, client.ObjectKey{Name: "test-vpc-subnet-add"}, fetched); err != nil {
		t.Fatalf("get: %v", err)
	}

	fetched.Spec.VPCConfig.SharedSubnets = append(
		fetched.Spec.VPCConfig.SharedSubnets,
		netv1alpha1.SharedSubnet{Path: testSubnetPath2, Name: testSubnetNameB},
	)
	if err := k8sClient.Update(testCtx, fetched); err != nil {
		t.Fatalf("expected admission, got: %v", err)
	}
}

func TestNamespaceNetworkConfiguration_SharedSubnetPodDefaultUpdate_Admitted(t *testing.T) {
	nnc := vpcNNC("test-vpc-subnet-upd-default", testVPCPath)
	nnc.Spec.VPCConfig.SharedSubnets = []netv1alpha1.SharedSubnet{
		{Path: testSubnetPath1, Name: testSubnetNameA},
	}
	if err := k8sClient.Create(testCtx, nnc); err != nil {
		t.Fatalf("create: %v", err)
	}
	defer func() { _ = k8sClient.Delete(testCtx, nnc) }()

	fetched := &netv1alpha1.NamespaceNetworkConfiguration{}
	if err := k8sClient.Get(testCtx, client.ObjectKey{Name: "test-vpc-subnet-upd-default"}, fetched); err != nil {
		t.Fatalf("get: %v", err)
	}

	fetched.Spec.VPCConfig.SharedSubnets[0].PodDefault = netv1alpha1.SharedSubnetDefaultTrue
	if err := k8sClient.Update(testCtx, fetched); err != nil {
		t.Fatalf("expected admission, got: %v", err)
	}
}

// -----------------------------------------------------------------------
// Status Subresource
//
// The source Ginkgo spec shared one base NNC per Context via BeforeEach,
// relying on an outer AfterEach sweep to delete it between Its. Since that
// sweep is dropped here, each test below creates and fetches its own
// freshly-named NNC to preserve the same per-test isolation.
// -----------------------------------------------------------------------

func TestNamespaceNetworkConfiguration_StatusValidReadyCondition_Admitted(t *testing.T) {
	nnc := vdsNNC("test-status-ready-cond", testNetName)
	if err := k8sClient.Create(testCtx, nnc); err != nil {
		t.Fatalf("create: %v", err)
	}
	defer func() { _ = k8sClient.Delete(testCtx, nnc) }()
	if err := k8sClient.Get(testCtx, client.ObjectKey{Name: nnc.Name}, nnc); err != nil {
		t.Fatalf("get: %v", err)
	}

	nnc.Status = &netv1alpha1.NamespaceNetworkStatus{
		Conditions: []metav1.Condition{
			makeCondition(netv1alpha1.NamespaceNetworkConditionReady, metav1.ConditionTrue),
		},
	}
	if err := k8sClient.Status().Update(testCtx, nnc); err != nil {
		t.Fatalf("expected admission, got: %v", err)
	}
}

func TestNamespaceNetworkConfiguration_StatusAssociatedNamespacesBothStatusValues_Admitted(t *testing.T) {
	nnc := vdsNNC("test-status-assoc-ns", testNetName)
	if err := k8sClient.Create(testCtx, nnc); err != nil {
		t.Fatalf("create: %v", err)
	}
	defer func() { _ = k8sClient.Delete(testCtx, nnc) }()
	if err := k8sClient.Get(testCtx, client.ObjectKey{Name: nnc.Name}, nnc); err != nil {
		t.Fatalf("get: %v", err)
	}

	nnc.Status = &netv1alpha1.NamespaceNetworkStatus{
		AssociatedNamespaces: []netv1alpha1.NamespaceNetworkAssociation{
			{Name: testNamespaceA, Status: netv1alpha1.NamespaceNetworkReconciling, Message: "in progress"},
			{Name: testNamespaceB, Status: netv1alpha1.NamespaceNetworkReconciled},
		},
	}
	if err := k8sClient.Status().Update(testCtx, nnc); err != nil {
		t.Fatalf("expected admission, got: %v", err)
	}
}

func TestNamespaceNetworkConfiguration_StatusDuplicateConditionTypes_Rejected(t *testing.T) {
	nnc := vdsNNC("test-status-dup-cond", testNetName)
	if err := k8sClient.Create(testCtx, nnc); err != nil {
		t.Fatalf("create: %v", err)
	}
	defer func() { _ = k8sClient.Delete(testCtx, nnc) }()
	if err := k8sClient.Get(testCtx, client.ObjectKey{Name: nnc.Name}, nnc); err != nil {
		t.Fatalf("get: %v", err)
	}

	nnc.Status = &netv1alpha1.NamespaceNetworkStatus{
		Conditions: []metav1.Condition{
			makeCondition("Ready", metav1.ConditionTrue),
			makeCondition("Ready", metav1.ConditionFalse),
		},
	}
	if err := k8sClient.Status().Update(testCtx, nnc); err == nil || !strings.Contains(err.Error(), "uplicate") {
		t.Fatalf("expected rejection containing %q, got: %v", "uplicate", err)
	}
}

func TestNamespaceNetworkConfiguration_StatusConditionReasonBadPattern_Rejected(t *testing.T) {
	nnc := vdsNNC("test-status-bad-reason", testNetName)
	if err := k8sClient.Create(testCtx, nnc); err != nil {
		t.Fatalf("create: %v", err)
	}
	defer func() { _ = k8sClient.Delete(testCtx, nnc) }()
	if err := k8sClient.Get(testCtx, client.ObjectKey{Name: nnc.Name}, nnc); err != nil {
		t.Fatalf("get: %v", err)
	}

	cond := makeCondition("Ready", metav1.ConditionTrue)
	cond.Reason = "not-camel-case" // hyphens are not allowed in reason
	nnc.Status = &netv1alpha1.NamespaceNetworkStatus{
		Conditions: []metav1.Condition{cond},
	}
	if err := k8sClient.Status().Update(testCtx, nnc); err == nil || !strings.Contains(err.Error(), "Invalid value") {
		t.Fatalf("expected rejection containing %q, got: %v", "Invalid value", err)
	}
}

func TestNamespaceNetworkConfiguration_StatusConditionStatusOutsideEnum_Rejected(t *testing.T) {
	nnc := vdsNNC("test-status-bad-status", testNetName)
	if err := k8sClient.Create(testCtx, nnc); err != nil {
		t.Fatalf("create: %v", err)
	}
	defer func() { _ = k8sClient.Delete(testCtx, nnc) }()
	if err := k8sClient.Get(testCtx, client.ObjectKey{Name: nnc.Name}, nnc); err != nil {
		t.Fatalf("get: %v", err)
	}

	nnc.Status = &netv1alpha1.NamespaceNetworkStatus{
		Conditions: []metav1.Condition{
			{
				Type:               "Ready",
				Status:             "bad-status",
				Reason:             testConditionReason,
				Message:            "msg",
				LastTransitionTime: metav1.NewTime(time.Now()),
			},
		},
	}
	if err := k8sClient.Status().Update(testCtx, nnc); err == nil || !strings.Contains(err.Error(), "Unsupported value") {
		t.Fatalf("expected rejection containing %q, got: %v", "Unsupported value", err)
	}
}

func TestNamespaceNetworkConfiguration_StatusAssociatedNamespaceUppercaseName_Rejected(t *testing.T) {
	nnc := vdsNNC("test-status-uppercase-ns", testNetName)
	if err := k8sClient.Create(testCtx, nnc); err != nil {
		t.Fatalf("create: %v", err)
	}
	defer func() { _ = k8sClient.Delete(testCtx, nnc) }()
	if err := k8sClient.Get(testCtx, client.ObjectKey{Name: nnc.Name}, nnc); err != nil {
		t.Fatalf("get: %v", err)
	}

	nnc.Status = &netv1alpha1.NamespaceNetworkStatus{
		AssociatedNamespaces: []netv1alpha1.NamespaceNetworkAssociation{
			{Name: "UPPERCASE-NS", Status: netv1alpha1.NamespaceNetworkReconciled},
		},
	}
	if err := k8sClient.Status().Update(testCtx, nnc); err == nil || !strings.Contains(err.Error(), "Invalid value") {
		t.Fatalf("expected rejection containing %q, got: %v", "Invalid value", err)
	}
}

func TestNamespaceNetworkConfiguration_StatusAssociatedNamespaceInvalidStatusEnum_Rejected(t *testing.T) {
	nnc := vdsNNC("test-status-bad-assoc-status", testNetName)
	if err := k8sClient.Create(testCtx, nnc); err != nil {
		t.Fatalf("create: %v", err)
	}
	defer func() { _ = k8sClient.Delete(testCtx, nnc) }()
	if err := k8sClient.Get(testCtx, client.ObjectKey{Name: nnc.Name}, nnc); err != nil {
		t.Fatalf("get: %v", err)
	}

	nnc.Status = &netv1alpha1.NamespaceNetworkStatus{
		AssociatedNamespaces: []netv1alpha1.NamespaceNetworkAssociation{
			{Name: testNamespaceA, Status: "BadState"},
		},
	}
	if err := k8sClient.Status().Update(testCtx, nnc); err == nil || !strings.Contains(err.Error(), "Unsupported value") {
		t.Fatalf("expected rejection containing %q, got: %v", "Unsupported value", err)
	}
}

func TestNamespaceNetworkConfiguration_StatusDuplicateAssociatedNamespaceNames_Rejected(t *testing.T) {
	nnc := vdsNNC("test-status-dup-assoc-ns", testNetName)
	if err := k8sClient.Create(testCtx, nnc); err != nil {
		t.Fatalf("create: %v", err)
	}
	defer func() { _ = k8sClient.Delete(testCtx, nnc) }()
	if err := k8sClient.Get(testCtx, client.ObjectKey{Name: nnc.Name}, nnc); err != nil {
		t.Fatalf("get: %v", err)
	}

	nnc.Status = &netv1alpha1.NamespaceNetworkStatus{
		AssociatedNamespaces: []netv1alpha1.NamespaceNetworkAssociation{
			{Name: testNamespaceA, Status: netv1alpha1.NamespaceNetworkReconciled},
			{Name: testNamespaceA, Status: netv1alpha1.NamespaceNetworkReconciling},
		},
	}
	if err := k8sClient.Status().Update(testCtx, nnc); err == nil || !strings.Contains(err.Error(), "uplicate") {
		t.Fatalf("expected rejection containing %q, got: %v", "uplicate", err)
	}
}

// -----------------------------------------------------------------------
// Finalizers
// -----------------------------------------------------------------------

func TestNamespaceNetworkConfiguration_AddProtectionFinalizer_Admitted(t *testing.T) {
	obj := vdsNNC("test-fin-add", testNetName)
	if err := k8sClient.Create(testCtx, obj); err != nil {
		t.Fatalf("create: %v", err)
	}
	defer func() { _ = k8sClient.Delete(testCtx, obj) }()

	fetched := &netv1alpha1.NamespaceNetworkConfiguration{}
	if err := k8sClient.Get(testCtx, client.ObjectKey{Name: "test-fin-add"}, fetched); err != nil {
		t.Fatalf("get: %v", err)
	}

	controllerutil.AddFinalizer(fetched, netv1alpha1.NamespaceNetworkProtectionFinalizer)
	if err := k8sClient.Update(testCtx, fetched); err != nil {
		t.Fatalf("update: %v", err)
	}

	if err := k8sClient.Get(testCtx, client.ObjectKey{Name: "test-fin-add"}, fetched); err != nil {
		t.Fatalf("get: %v", err)
	}
	found := false
	for _, f := range fetched.Finalizers {
		if f == netv1alpha1.NamespaceNetworkProtectionFinalizer {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected finalizer %q on object, got: %v", netv1alpha1.NamespaceNetworkProtectionFinalizer, fetched.Finalizers)
	}

	// Cleanup: remove finalizer so the deferred delete is not blocked.
	controllerutil.RemoveFinalizer(fetched, netv1alpha1.NamespaceNetworkProtectionFinalizer)
	if err := k8sClient.Update(testCtx, fetched); err != nil {
		t.Fatalf("remove finalizer: %v", err)
	}
}

func TestNamespaceNetworkConfiguration_DeleteBlockedByFinalizer(t *testing.T) {
	obj := vdsNNC("test-fin-block", testNetName)
	if err := k8sClient.Create(testCtx, obj); err != nil {
		t.Fatalf("create: %v", err)
	}

	fetched := &netv1alpha1.NamespaceNetworkConfiguration{}
	if err := k8sClient.Get(testCtx, client.ObjectKey{Name: "test-fin-block"}, fetched); err != nil {
		t.Fatalf("get: %v", err)
	}

	controllerutil.AddFinalizer(fetched, netv1alpha1.NamespaceNetworkProtectionFinalizer)
	if err := k8sClient.Update(testCtx, fetched); err != nil {
		t.Fatalf("update: %v", err)
	}

	if err := k8sClient.Delete(testCtx, fetched); err != nil {
		t.Fatalf("delete: %v", err)
	}

	// The object must still exist with DeletionTimestamp set.
	if err := k8sClient.Get(testCtx, client.ObjectKey{Name: "test-fin-block"}, fetched); err != nil {
		t.Fatalf("get after delete: %v", err)
	}
	if fetched.DeletionTimestamp == nil {
		t.Fatalf("expected DeletionTimestamp to be set while finalizer is present")
	}

	// Cleanup: remove finalizer so the object is actually deleted.
	controllerutil.RemoveFinalizer(fetched, netv1alpha1.NamespaceNetworkProtectionFinalizer)
	if err := k8sClient.Update(testCtx, fetched); err != nil {
		t.Fatalf("remove finalizer: %v", err)
	}
}

func TestNamespaceNetworkConfiguration_FullyDeletedAfterFinalizerRemoved(t *testing.T) {
	obj := vdsNNC(testFinalizer, testNetName)
	if err := k8sClient.Create(testCtx, obj); err != nil {
		t.Fatalf("create: %v", err)
	}

	fetched := &netv1alpha1.NamespaceNetworkConfiguration{}
	if err := k8sClient.Get(testCtx, client.ObjectKey{Name: testFinalizer}, fetched); err != nil {
		t.Fatalf("get: %v", err)
	}

	controllerutil.AddFinalizer(fetched, netv1alpha1.NamespaceNetworkProtectionFinalizer)
	if err := k8sClient.Update(testCtx, fetched); err != nil {
		t.Fatalf("update: %v", err)
	}

	// Issue a delete — object lingers because of the finalizer.
	if err := k8sClient.Delete(testCtx, fetched); err != nil {
		t.Fatalf("delete: %v", err)
	}

	// Remove the finalizer: the API server can now purge the object.
	if err := k8sClient.Get(testCtx, client.ObjectKey{Name: testFinalizer}, fetched); err != nil {
		t.Fatalf("get: %v", err)
	}
	controllerutil.RemoveFinalizer(fetched, netv1alpha1.NamespaceNetworkProtectionFinalizer)
	if err := k8sClient.Update(testCtx, fetched); err != nil {
		t.Fatalf("remove finalizer: %v", err)
	}

	err := k8sClient.Get(testCtx, client.ObjectKey{Name: testFinalizer}, fetched)
	if !apierrors.IsNotFound(err) {
		t.Fatalf("expected IsNotFound after finalizer removal and delete, got: %v", err)
	}
}

// -----------------------------------------------------------------------
// NSX Tier-1 Config — field validations and mutability.
// -----------------------------------------------------------------------

const (
	testNamespaceCIDR = "192.168.1.0/24"
	testIngressCIDR   = "192.168.2.0/24"
	testEgressCIDR    = "192.168.3.0/24"
	testTier0Gateway  = "/infra/tier-0s/my-gw"
)

func TestNamespaceNetworkConfiguration_NSXTier1ValidOverride_Admitted(t *testing.T) {
	nnc := &netv1alpha1.NamespaceNetworkConfiguration{
		ObjectMeta: metav1.ObjectMeta{Name: "test-t1-valid-override"},
		Spec: netv1alpha1.NamespaceNetworkSpec{
			Type: netv1alpha1.NetworkProviderNSXTier1,
			NamespaceNetworkConfig: netv1alpha1.NamespaceNetworkConfig{
				NSXTier1Config: &netv1alpha1.NSXTier1Config{
					NamespaceCIDRs:   []string{testNamespaceCIDR},
					IngressCIDRs:     []string{testIngressCIDR},
					EgressCIDRs:      []string{testEgressCIDR},
					Tier0Gateway:     testTier0Gateway,
					RoutingMode:      netv1alpha1.NSXTier1RoutingModeNAT,
					LoadBalancerSize: netv1alpha1.NSXLoadBalancerSizeSmall,
				},
			},
		},
	}
	if err := k8sClient.Create(testCtx, nnc); err != nil {
		t.Fatalf("expected admission, got: %v", err)
	}
	defer func() { _ = k8sClient.Delete(testCtx, nnc) }()
}

func TestNamespaceNetworkConfiguration_NSXTier1RoutedWithEgress_Rejected(t *testing.T) {
	nnc := &netv1alpha1.NamespaceNetworkConfiguration{
		ObjectMeta: metav1.ObjectMeta{Name: "test-t1-routed-with-egress"},
		Spec: netv1alpha1.NamespaceNetworkSpec{
			Type: netv1alpha1.NetworkProviderNSXTier1,
			NamespaceNetworkConfig: netv1alpha1.NamespaceNetworkConfig{
				NSXTier1Config: &netv1alpha1.NSXTier1Config{
					NamespaceCIDRs: []string{testNamespaceCIDR},
					IngressCIDRs:   []string{testIngressCIDR},
					EgressCIDRs:    []string{testEgressCIDR},
					RoutingMode:    netv1alpha1.NSXTier1RoutingModeRouted,
				},
			},
		},
	}
	err := k8sClient.Create(testCtx, nnc)
	if err == nil || !strings.Contains(err.Error(), "egressCIDRs must not be set when routingMode is Routed") {
		t.Fatalf("expected rejection containing %q, got: %v", "egressCIDRs must not be set when routingMode is Routed", err)
	}
}

func TestNamespaceNetworkConfiguration_NSXTier1MissingNamespaceCIDR_Rejected(t *testing.T) {
	nnc := &netv1alpha1.NamespaceNetworkConfiguration{
		ObjectMeta: metav1.ObjectMeta{Name: "test-t1-missing-namespace-cidr"},
		Spec: netv1alpha1.NamespaceNetworkSpec{
			Type: netv1alpha1.NetworkProviderNSXTier1,
			NamespaceNetworkConfig: netv1alpha1.NamespaceNetworkConfig{
				NSXTier1Config: &netv1alpha1.NSXTier1Config{
					IngressCIDRs: []string{testIngressCIDR},
					Tier0Gateway: testTier0Gateway,
				},
			},
		},
	}
	err := k8sClient.Create(testCtx, nnc)
	if err == nil || !strings.Contains(err.Error(), "namespaceCIDRs must be set when tier0Gateway, ingressCIDRs, or egressCIDRs are specified") {
		t.Fatalf("expected rejection, got: %v", err)
	}
}

func TestNamespaceNetworkConfiguration_NSXTier1MissingIngressCIDR_Rejected(t *testing.T) {
	nnc := &netv1alpha1.NamespaceNetworkConfiguration{
		ObjectMeta: metav1.ObjectMeta{Name: "test-t1-missing-ingress-cidr"},
		Spec: netv1alpha1.NamespaceNetworkSpec{
			Type: netv1alpha1.NetworkProviderNSXTier1,
			NamespaceNetworkConfig: netv1alpha1.NamespaceNetworkConfig{
				NSXTier1Config: &netv1alpha1.NSXTier1Config{
					NamespaceCIDRs: []string{testNamespaceCIDR},
					Tier0Gateway:   testTier0Gateway,
				},
			},
		},
	}
	err := k8sClient.Create(testCtx, nnc)
	if err == nil || !strings.Contains(err.Error(), "ingressCIDRs must be set when tier0Gateway, namespaceCIDRs, or egressCIDRs are specified") {
		t.Fatalf("expected rejection, got: %v", err)
	}
}

func TestNamespaceNetworkConfiguration_NSXTier1MissingEgressCIDR_Rejected(t *testing.T) {
	nnc := &netv1alpha1.NamespaceNetworkConfiguration{
		ObjectMeta: metav1.ObjectMeta{Name: "test-t1-missing-egress-cidr"},
		Spec: netv1alpha1.NamespaceNetworkSpec{
			Type: netv1alpha1.NetworkProviderNSXTier1,
			NamespaceNetworkConfig: netv1alpha1.NamespaceNetworkConfig{
				NSXTier1Config: &netv1alpha1.NSXTier1Config{
					NamespaceCIDRs: []string{testNamespaceCIDR},
					IngressCIDRs:   []string{testIngressCIDR},
					Tier0Gateway:   testTier0Gateway,
					RoutingMode:    netv1alpha1.NSXTier1RoutingModeNAT,
				},
			},
		},
	}
	err := k8sClient.Create(testCtx, nnc)
	if err == nil || !strings.Contains(err.Error(), "egressCIDRs must be set when routingMode is NAT and tier0Gateway, namespaceCIDRs, or ingressCIDRs are specified") {
		t.Fatalf("expected rejection, got: %v", err)
	}
}

func TestNamespaceNetworkConfiguration_NSXTier1UpdatesAndImmutability(t *testing.T) {
	nnc := &netv1alpha1.NamespaceNetworkConfiguration{
		ObjectMeta: metav1.ObjectMeta{Name: "test-t1-upd-immutability"},
		Spec: netv1alpha1.NamespaceNetworkSpec{
			Type: netv1alpha1.NetworkProviderNSXTier1,
			NamespaceNetworkConfig: netv1alpha1.NamespaceNetworkConfig{
				NSXTier1Config: &netv1alpha1.NSXTier1Config{
					NamespaceCIDRs:   []string{testNamespaceCIDR},
					IngressCIDRs:     []string{testIngressCIDR},
					EgressCIDRs:      []string{testEgressCIDR},
					Tier0Gateway:     testTier0Gateway,
					RoutingMode:      netv1alpha1.NSXTier1RoutingModeNAT,
					LoadBalancerSize: netv1alpha1.NSXLoadBalancerSizeSmall,
				},
			},
		},
	}
	if err := k8sClient.Create(testCtx, nnc); err != nil {
		t.Fatalf("create: %v", err)
	}
	defer func() { _ = k8sClient.Delete(testCtx, nnc) }()

	fetched := &netv1alpha1.NamespaceNetworkConfiguration{}
	if err := k8sClient.Get(testCtx, client.ObjectKey{Name: nnc.Name}, fetched); err != nil {
		t.Fatalf("get: %v", err)
	}

	// 1. Reject updating tier0Gateway
	fetched.Spec.NSXTier1Config.Tier0Gateway = "/infra/tier-0s/other-gw"
	if err := k8sClient.Update(testCtx, fetched); err == nil || !strings.Contains(err.Error(), "tier0Gateway is immutable once set") {
		t.Fatalf("expected rejection for tier0Gateway update, got: %v", err)
	}

	// Reset tier0Gateway
	if err := k8sClient.Get(testCtx, client.ObjectKey{Name: nnc.Name}, fetched); err != nil {
		t.Fatalf("get: %v", err)
	}

	// 2. Reject updating routingMode
	fetched.Spec.NSXTier1Config.RoutingMode = netv1alpha1.NSXTier1RoutingModeRouted
	if err := k8sClient.Update(testCtx, fetched); err == nil || !strings.Contains(err.Error(), "routingMode is immutable once set") {
		t.Fatalf("expected rejection for routingMode update, got: %v", err)
	}

	// Reset routingMode
	if err := k8sClient.Get(testCtx, client.ObjectKey{Name: nnc.Name}, fetched); err != nil {
		t.Fatalf("get: %v", err)
	}

	// 3. Reject updating loadBalancerSize
	fetched.Spec.NSXTier1Config.LoadBalancerSize = netv1alpha1.NSXLoadBalancerSizeMedium
	if err := k8sClient.Update(testCtx, fetched); err == nil || !strings.Contains(err.Error(), "loadBalancerSize is immutable once set") {
		t.Fatalf("expected rejection for loadBalancerSize update, got: %v", err)
	}

	// Reset loadBalancerSize
	if err := k8sClient.Get(testCtx, client.ObjectKey{Name: nnc.Name}, fetched); err != nil {
		t.Fatalf("get: %v", err)
	}

	// 4. Reject updating namespaceCIDRs (removing an entry)
	fetched.Spec.NSXTier1Config.NamespaceCIDRs = []string{}
	if err := k8sClient.Update(testCtx, fetched); err == nil || !strings.Contains(err.Error(), "namespaceCIDRs is append-only") {
		t.Fatalf("expected rejection for namespaceCIDRs removal, got: %v", err)
	}

	// Reset namespaceCIDRs
	if err := k8sClient.Get(testCtx, client.ObjectKey{Name: nnc.Name}, fetched); err != nil {
		t.Fatalf("get: %v", err)
	}

	// 5. Reject updating ingressCIDRs (removing an entry)
	fetched.Spec.NSXTier1Config.IngressCIDRs = []string{}
	if err := k8sClient.Update(testCtx, fetched); err == nil || !strings.Contains(err.Error(), "ingressCIDRs is append-only") {
		t.Fatalf("expected rejection for ingressCIDRs removal, got: %v", err)
	}

	// Reset ingressCIDRs
	if err := k8sClient.Get(testCtx, client.ObjectKey{Name: nnc.Name}, fetched); err != nil {
		t.Fatalf("get: %v", err)
	}

	// 6. Reject updating egressCIDRs (removing an entry)
	fetched.Spec.NSXTier1Config.EgressCIDRs = []string{}
	if err := k8sClient.Update(testCtx, fetched); err == nil || !strings.Contains(err.Error(), "egressCIDRs is append-only") {
		t.Fatalf("expected rejection for egressCIDRs removal, got: %v", err)
	}

	// Reset egressCIDRs
	if err := k8sClient.Get(testCtx, client.ObjectKey{Name: nnc.Name}, fetched); err != nil {
		t.Fatalf("get: %v", err)
	}

	// 7. Accept appending new CIDRs
	fetched.Spec.NSXTier1Config.NamespaceCIDRs = append(fetched.Spec.NSXTier1Config.NamespaceCIDRs, "192.168.10.0/24")
	fetched.Spec.NSXTier1Config.IngressCIDRs = append(fetched.Spec.NSXTier1Config.IngressCIDRs, "192.168.20.0/24")
	fetched.Spec.NSXTier1Config.EgressCIDRs = append(fetched.Spec.NSXTier1Config.EgressCIDRs, "192.168.30.0/24")
	if err := k8sClient.Update(testCtx, fetched); err != nil {
		t.Fatalf("expected successful CIDR append, got: %v", err)
	}

	// Reset fetched
	if err := k8sClient.Get(testCtx, client.ObjectKey{Name: nnc.Name}, fetched); err != nil {
		t.Fatalf("get: %v", err)
	}

	// 8. Reject updating subnetPrefixLength once set
	fetched.Spec.NSXTier1Config.SubnetPrefixLength = 24
	if err := k8sClient.Update(testCtx, fetched); err != nil {
		t.Fatalf("expected setting subnetPrefixLength first time to succeed, got: %v", err)
	}
	if err := k8sClient.Get(testCtx, client.ObjectKey{Name: nnc.Name}, fetched); err != nil {
		t.Fatalf("get: %v", err)
	}
	fetched.Spec.NSXTier1Config.SubnetPrefixLength = 28
	if err := k8sClient.Update(testCtx, fetched); err == nil || !strings.Contains(err.Error(), "subnetPrefixLength is immutable once set") {
		t.Fatalf("expected rejection for subnetPrefixLength update, got: %v", err)
	}

	// Reset fetched
	if err := k8sClient.Get(testCtx, client.ObjectKey{Name: nnc.Name}, fetched); err != nil {
		t.Fatalf("get: %v", err)
	}

	// 9. Reject removing nsxTier1Config once set
	fetched.Spec.NSXTier1Config = nil
	if err := k8sClient.Update(testCtx, fetched); err == nil || !strings.Contains(err.Error(), "nsxTier1Config cannot be added or removed once set") {
		t.Fatalf("expected rejection for removing nsxTier1Config, got: %v", err)
	}
}

func TestNamespaceNetworkConfiguration_NSXTier1InheritToOverrideTransition_Rejected(t *testing.T) {
	inheritNNC := &netv1alpha1.NamespaceNetworkConfiguration{
		ObjectMeta: metav1.ObjectMeta{Name: "test-t1-inherit-to-override"},
		Spec: netv1alpha1.NamespaceNetworkSpec{
			Type: netv1alpha1.NetworkProviderNSXTier1,
		},
	}
	if err := k8sClient.Create(testCtx, inheritNNC); err != nil {
		t.Fatalf("create: %v", err)
	}
	defer func() { _ = k8sClient.Delete(testCtx, inheritNNC) }()

	fetched := &netv1alpha1.NamespaceNetworkConfiguration{}
	if err := k8sClient.Get(testCtx, client.ObjectKey{Name: inheritNNC.Name}, fetched); err != nil {
		t.Fatalf("get: %v", err)
	}

	// Transition to override mode: set nsxTier1Config (should be rejected by the strict presence immutability rule)
	fetched.Spec.NSXTier1Config = &netv1alpha1.NSXTier1Config{
		NamespaceCIDRs: []string{testNamespaceCIDR},
		IngressCIDRs:   []string{testIngressCIDR},
		EgressCIDRs:    []string{testEgressCIDR},
	}
	err := k8sClient.Update(testCtx, fetched)
	if err == nil || !strings.Contains(err.Error(), "nsxTier1Config cannot be added or removed once set") {
		t.Fatalf("expected rejection for adding nsxTier1Config online, got: %v", err)
	}
}

func TestNamespaceNetworkConfiguration_NSXTier1MutualExclusion(t *testing.T) {
	// 1. Populating vsphereDistributedConfig when type is nsx-tier1
	nnc1 := &netv1alpha1.NamespaceNetworkConfiguration{
		ObjectMeta: metav1.ObjectMeta{Name: "test-mut-vds-when-t1"},
		Spec: netv1alpha1.NamespaceNetworkSpec{
			Type: netv1alpha1.NetworkProviderNSXTier1,
			NamespaceNetworkConfig: netv1alpha1.NamespaceNetworkConfig{
				VSphereDistributedConfig: netv1alpha1.VSphereDistributedConfig{
					DefaultNetwork: testNetName,
					Networks:       []netv1alpha1.VSphereDistributedNetworkRef{{Name: testNetName}},
				},
			},
		},
	}
	if err := k8sClient.Create(testCtx, nnc1); err == nil || !strings.Contains(err.Error(), "vsphereDistributedConfig must not be populated when type is not vsphere-distributed") {
		t.Fatalf("expected rejection, got: %v", err)
	}

	// 2. Populating vpcConfig when type is nsx-tier1
	nnc2 := &netv1alpha1.NamespaceNetworkConfiguration{
		ObjectMeta: metav1.ObjectMeta{Name: "test-mut-vpc-when-t1"},
		Spec: netv1alpha1.NamespaceNetworkSpec{
			Type: netv1alpha1.NetworkProviderNSXTier1,
			NamespaceNetworkConfig: netv1alpha1.NamespaceNetworkConfig{
				VPCConfig: netv1alpha1.VPCConfig{
					VPC: "/infra/vpcs/vpc-1",
				},
			},
		},
	}
	if err := k8sClient.Create(testCtx, nnc2); err == nil || !strings.Contains(err.Error(), "vpcConfig must not be populated when type is not vpc") {
		t.Fatalf("expected rejection, got: %v", err)
	}

	// 3. Populating nsxTier1Config when type is vsphere-distributed
	nnc3 := &netv1alpha1.NamespaceNetworkConfiguration{
		ObjectMeta: metav1.ObjectMeta{Name: "test-mut-t1-when-vds"},
		Spec: netv1alpha1.NamespaceNetworkSpec{
			Type: netv1alpha1.NetworkProviderVSphereDistributed,
			NamespaceNetworkConfig: netv1alpha1.NamespaceNetworkConfig{
				VSphereDistributedConfig: netv1alpha1.VSphereDistributedConfig{
					DefaultNetwork: testNetName,
					Networks:       []netv1alpha1.VSphereDistributedNetworkRef{{Name: testNetName}},
				},
				NSXTier1Config: &netv1alpha1.NSXTier1Config{
					NamespaceCIDRs: []string{testNamespaceCIDR},
				},
			},
		},
	}
	if err := k8sClient.Create(testCtx, nnc3); err == nil || !strings.Contains(err.Error(), "nsxTier1Config must not be populated when type is not nsx-tier1") {
		t.Fatalf("expected rejection, got: %v", err)
	}
}
