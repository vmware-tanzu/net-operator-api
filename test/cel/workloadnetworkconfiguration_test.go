// Copyright (c) 2026 Broadcom. All Rights Reserved.
// Broadcom Confidential. The term "Broadcom" refers to Broadcom Inc.
// and/or its subsidiaries.

package cel_test

import (
	"strings"
	"testing"
	"time"

	netv1alpha1 "github.com/vmware-tanzu/net-operator-api/api/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// vdsSystemConfig returns a minimal valid vsphere-distributed system configuration.
func vdsSystemConfig() *netv1alpha1.NamespaceNetworkConfig {
	return &netv1alpha1.NamespaceNetworkConfig{
		VSphereDistributedConfig: netv1alpha1.VSphereDistributedConfig{
			Networks:       []netv1alpha1.VSphereDistributedNetworkRef{{Name: wncSystemNetName}},
			DefaultNetwork: wncSystemNetName,
		},
	}
}

// vdsWNC builds a minimal valid vsphere-distributed WNC named "default".
func vdsWNC() *netv1alpha1.WorkloadNetworkConfiguration {
	return &netv1alpha1.WorkloadNetworkConfiguration{
		ObjectMeta: metav1.ObjectMeta{Name: wncDefaultName},
		Spec: netv1alpha1.WorkloadNetworkConfigurationSpec{
			Providers: []netv1alpha1.NetworkProviderEntry{
				{
					Type:                netv1alpha1.NetworkProviderVSphereDistributed,
					SystemConfiguration: vdsSystemConfig(),
				},
			},
			ActiveSystemProvider: netv1alpha1.NetworkProviderVSphereDistributed,
		},
	}
}

// unstrWNC creates a bare unstructured WNC. Pass nil for spec to omit the spec key entirely.
func unstrWNC(name string, spec map[string]interface{}) *unstructured.Unstructured {
	obj := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": wncAPIVersion,
			"kind":       wncKind,
			"metadata":   map[string]interface{}{"name": name},
		},
	}
	if spec != nil {
		obj.Object["spec"] = spec
	}
	return obj
}

// vdsProviderEntry returns an unstructured vsphere-distributed provider entry.
func vdsProviderEntry(networkName string) map[string]interface{} {
	return map[string]interface{}{
		"type": string(netv1alpha1.NetworkProviderVSphereDistributed),
		"systemConfiguration": map[string]interface{}{
			"vsphereDistributedConfig": map[string]interface{}{
				"networks":       []interface{}{map[string]interface{}{"name": networkName}},
				"defaultNetwork": networkName,
			},
		},
	}
}

// makeWNCCondition returns a complete, valid metav1.Condition for WNC status tests.
func makeWNCCondition(condType string, status metav1.ConditionStatus) metav1.Condition {
	return metav1.Condition{
		Type:               condType,
		Status:             status,
		Reason:             testConditionReason,
		Message:            "test message",
		LastTransitionTime: metav1.NewTime(time.Now()),
	}
}

// -----------------------------------------------------------------------
// Name Enforcement
// Rule: self.metadata.name == 'default'
// -----------------------------------------------------------------------

func TestWorkloadNetworkConfiguration_NotNamedDefault_Rejected(t *testing.T) {
	wnc := vdsWNC()
	wnc.Name = "not-default"
	if err := k8sClient.Create(testCtx, wnc); err == nil || !strings.Contains(err.Error(), "must be named 'default'") {
		t.Fatalf("expected rejection containing %q, got: %v", "must be named 'default'", err)
	}
}

func TestWorkloadNetworkConfiguration_NamedDefault_Admitted(t *testing.T) {
	wnc := vdsWNC()
	if err := k8sClient.Create(testCtx, wnc); err != nil {
		t.Fatalf("expected admission, got: %v", err)
	}
	defer func() { _ = k8sClient.Delete(testCtx, wnc) }()
}

// -----------------------------------------------------------------------
// Providers List — structural validations
// providers: +required, minItems=1, maxItems=3, listType=map (key=type)
// -----------------------------------------------------------------------

func TestWorkloadNetworkConfiguration_ProvidersAbsent_Rejected(t *testing.T) {
	wnc := &netv1alpha1.WorkloadNetworkConfiguration{
		ObjectMeta: metav1.ObjectMeta{Name: wncDefaultName},
		Spec: netv1alpha1.WorkloadNetworkConfigurationSpec{
			// Providers is nil — serialized as absent with omitempty.
			ActiveSystemProvider: netv1alpha1.NetworkProviderVSphereDistributed,
		},
	}
	if err := k8sClient.Create(testCtx, wnc); err == nil || !strings.Contains(err.Error(), "providers") {
		t.Fatalf("expected rejection containing %q, got: %v", "providers", err)
	}
}

func TestWorkloadNetworkConfiguration_EmptyProvidersList_Rejected(t *testing.T) {
	obj := unstrWNC(wncDefaultName, map[string]interface{}{
		"providers":            []interface{}{},
		"activeSystemProvider": string(netv1alpha1.NetworkProviderVSphereDistributed),
	})
	if err := k8sClient.Create(testCtx, obj); err == nil || !strings.Contains(err.Error(), "at least 1") {
		t.Fatalf("expected rejection containing %q, got: %v", "at least 1", err)
	}
}

func TestWorkloadNetworkConfiguration_UnknownProviderType_Rejected(t *testing.T) {
	obj := unstrWNC(wncDefaultName, map[string]interface{}{
		"providers": []interface{}{
			map[string]interface{}{
				"type":                "unknown-provider",
				"systemConfiguration": map[string]interface{}{},
			},
		},
		"activeSystemProvider": "unknown-provider",
	})
	if err := k8sClient.Create(testCtx, obj); err == nil || !strings.Contains(err.Error(), "Unsupported value") {
		t.Fatalf("expected rejection containing %q, got: %v", "Unsupported value", err)
	}
}

func TestWorkloadNetworkConfiguration_DuplicateProviderTypes_Rejected(t *testing.T) {
	wnc := &netv1alpha1.WorkloadNetworkConfiguration{
		ObjectMeta: metav1.ObjectMeta{Name: wncDefaultName},
		Spec: netv1alpha1.WorkloadNetworkConfigurationSpec{
			Providers: []netv1alpha1.NetworkProviderEntry{
				{Type: netv1alpha1.NetworkProviderVSphereDistributed, SystemConfiguration: vdsSystemConfig()},
				{Type: netv1alpha1.NetworkProviderVSphereDistributed, SystemConfiguration: vdsSystemConfig()},
			},
			ActiveSystemProvider: netv1alpha1.NetworkProviderVSphereDistributed,
		},
	}
	if err := k8sClient.Create(testCtx, wnc); err == nil || !strings.Contains(err.Error(), "uplicate") {
		t.Fatalf("expected rejection containing %q, got: %v", "uplicate", err)
	}
}

func TestWorkloadNetworkConfiguration_OneValidProvider_Admitted(t *testing.T) {
	wnc := vdsWNC()
	if err := k8sClient.Create(testCtx, wnc); err != nil {
		t.Fatalf("expected admission, got: %v", err)
	}
	defer func() { _ = k8sClient.Delete(testCtx, wnc) }()
}

func TestWorkloadNetworkConfiguration_TwoSupportedProviders_Admitted(t *testing.T) {
	wnc := &netv1alpha1.WorkloadNetworkConfiguration{
		ObjectMeta: metav1.ObjectMeta{Name: wncDefaultName},
		Spec: netv1alpha1.WorkloadNetworkConfigurationSpec{
			Providers: []netv1alpha1.NetworkProviderEntry{
				{Type: netv1alpha1.NetworkProviderVSphereDistributed, SystemConfiguration: vdsSystemConfig()},
				{Type: netv1alpha1.NetworkProviderVPC, SystemConfiguration: &netv1alpha1.NamespaceNetworkConfig{
					VPCConfig: netv1alpha1.VPCConfig{VPC: testWNCVPCPath},
				}},
			},
			ActiveSystemProvider: netv1alpha1.NetworkProviderVSphereDistributed,
		},
	}
	if err := k8sClient.Create(testCtx, wnc); err != nil {
		t.Fatalf("expected admission, got: %v", err)
	}
	defer func() { _ = k8sClient.Delete(testCtx, wnc) }()
}

// -----------------------------------------------------------------------
// Per-Entry vsphereDistributedConfig CEL
// Rule 1: type == 'vsphere-distributed' → vsphereDistributedConfig must be present
// Rule 2: type != 'vsphere-distributed' → vsphereDistributedConfig must be absent
// -----------------------------------------------------------------------

func TestWorkloadNetworkConfiguration_VDSEntryWithoutVDSConfig_Rejected(t *testing.T) {
	obj := unstrWNC(wncDefaultName, map[string]interface{}{
		"providers": []interface{}{
			map[string]interface{}{
				"type":                string(netv1alpha1.NetworkProviderVSphereDistributed),
				"systemConfiguration": map[string]interface{}{},
			},
		},
		"activeSystemProvider": string(netv1alpha1.NetworkProviderVSphereDistributed),
	})
	if err := k8sClient.Create(testCtx, obj); err == nil || !strings.Contains(err.Error(), "vsphereDistributedConfig must be set") {
		t.Fatalf("expected rejection containing %q, got: %v", "vsphereDistributedConfig must be set", err)
	}
}

func TestWorkloadNetworkConfiguration_VDSEntryWithVDSConfig_Admitted(t *testing.T) {
	obj := unstrWNC(wncDefaultName, map[string]interface{}{
		"providers":            []interface{}{vdsProviderEntry(wncSystemNetName)},
		"activeSystemProvider": string(netv1alpha1.NetworkProviderVSphereDistributed),
	})
	if err := k8sClient.Create(testCtx, obj); err != nil {
		t.Fatalf("expected admission, got: %v", err)
	}
	defer func() { _ = k8sClient.Delete(testCtx, obj) }()
}

func TestWorkloadNetworkConfiguration_NSXTier1EntryWithVDSConfig_Rejected(t *testing.T) {
	obj := unstrWNC(wncDefaultName, map[string]interface{}{
		"providers": []interface{}{
			map[string]interface{}{
				"type": string(netv1alpha1.NetworkProviderNSXTier1),
				"systemConfiguration": map[string]interface{}{
					"vsphereDistributedConfig": map[string]interface{}{
						"networks":       []interface{}{map[string]interface{}{"name": wncSystemNetName}},
						"defaultNetwork": wncSystemNetName,
					},
				},
			},
		},
		"activeSystemProvider": string(netv1alpha1.NetworkProviderNSXTier1),
	})
	if err := k8sClient.Create(testCtx, obj); err == nil || !strings.Contains(err.Error(), "may only be set") {
		t.Fatalf("expected rejection containing %q, got: %v", "may only be set", err)
	}
}

func TestWorkloadNetworkConfiguration_VPCEntryWithVPCConfig_Admitted(t *testing.T) {
	wnc := &netv1alpha1.WorkloadNetworkConfiguration{
		ObjectMeta: metav1.ObjectMeta{Name: wncDefaultName},
		Spec: netv1alpha1.WorkloadNetworkConfigurationSpec{
			Providers: []netv1alpha1.NetworkProviderEntry{
				{Type: netv1alpha1.NetworkProviderVPC, SystemConfiguration: &netv1alpha1.NamespaceNetworkConfig{
					VPCConfig: netv1alpha1.VPCConfig{VPC: testWNCVPCPath},
				}},
			},
			ActiveSystemProvider: netv1alpha1.NetworkProviderVPC,
		},
	}
	if err := k8sClient.Create(testCtx, wnc); err != nil {
		t.Fatalf("expected admission, got: %v", err)
	}
	defer func() { _ = k8sClient.Delete(testCtx, wnc) }()
}

// -----------------------------------------------------------------------
// activeSystemProvider
// Rule: self.providers.exists(p, p.type == self.activeSystemProvider)
// Schema enum: vsphere-distributed | vpc (nsx-tier1 not yet supported)
// -----------------------------------------------------------------------

func TestWorkloadNetworkConfiguration_ActiveSystemProviderNotInProvidersList_Rejected(t *testing.T) {
	wnc := &netv1alpha1.WorkloadNetworkConfiguration{
		ObjectMeta: metav1.ObjectMeta{Name: wncDefaultName},
		Spec: netv1alpha1.WorkloadNetworkConfigurationSpec{
			Providers: []netv1alpha1.NetworkProviderEntry{
				{Type: netv1alpha1.NetworkProviderVSphereDistributed, SystemConfiguration: vdsSystemConfig()},
			},
			// Only vsphere-distributed declared, but activeSystemProvider references vpc.
			ActiveSystemProvider: netv1alpha1.NetworkProviderVPC,
		},
	}
	if err := k8sClient.Create(testCtx, wnc); err == nil || !strings.Contains(err.Error(), "must reference a provider type") {
		t.Fatalf("expected rejection containing %q, got: %v", "must reference a provider type", err)
	}
}

func TestWorkloadNetworkConfiguration_SwitchActiveSystemProviderToDeclaredProvider_Admitted(t *testing.T) {
	wnc := &netv1alpha1.WorkloadNetworkConfiguration{
		ObjectMeta: metav1.ObjectMeta{Name: wncDefaultName},
		Spec: netv1alpha1.WorkloadNetworkConfigurationSpec{
			Providers: []netv1alpha1.NetworkProviderEntry{
				{Type: netv1alpha1.NetworkProviderVSphereDistributed, SystemConfiguration: vdsSystemConfig()},
				{Type: netv1alpha1.NetworkProviderVPC, SystemConfiguration: &netv1alpha1.NamespaceNetworkConfig{
					VPCConfig: netv1alpha1.VPCConfig{VPC: testWNCVPCPath},
				}},
			},
			ActiveSystemProvider: netv1alpha1.NetworkProviderVSphereDistributed,
		},
	}
	if err := k8sClient.Create(testCtx, wnc); err != nil {
		t.Fatalf("create: %v", err)
	}
	defer func() { _ = k8sClient.Delete(testCtx, wnc) }()

	fetched := &netv1alpha1.WorkloadNetworkConfiguration{}
	if err := k8sClient.Get(testCtx, client.ObjectKey{Name: wncDefaultName}, fetched); err != nil {
		t.Fatalf("get: %v", err)
	}

	fetched.Spec.ActiveSystemProvider = netv1alpha1.NetworkProviderVPC
	if err := k8sClient.Update(testCtx, fetched); err != nil {
		t.Fatalf("expected admission for activeSystemProvider switch, got: %v", err)
	}
}

// -----------------------------------------------------------------------
// Status Subresource
// conditions: maxItems=8, listType=map (key=type), each condition requires
//   lastTransitionTime, message, reason (CamelCase pattern),
//   status (True|False|Unknown enum), type
// -----------------------------------------------------------------------

func TestWorkloadNetworkConfiguration_ValidCondition_Admitted(t *testing.T) {
	wnc := vdsWNC()
	if err := k8sClient.Create(testCtx, wnc); err != nil {
		t.Fatalf("create: %v", err)
	}
	defer func() { _ = k8sClient.Delete(testCtx, wnc) }()
	if err := k8sClient.Get(testCtx, client.ObjectKey{Name: wncDefaultName}, wnc); err != nil {
		t.Fatalf("get: %v", err)
	}

	wnc.Status = &netv1alpha1.WorkloadNetworkConfigurationStatus{
		Conditions: []metav1.Condition{
			makeWNCCondition(netv1alpha1.WorkloadNetworkConditionReady, metav1.ConditionTrue),
		},
	}
	if err := k8sClient.Status().Update(testCtx, wnc); err != nil {
		t.Fatalf("expected admission for status update, got: %v", err)
	}
}

func TestWorkloadNetworkConfiguration_DuplicateConditionTypes_Rejected(t *testing.T) {
	wnc := vdsWNC()
	if err := k8sClient.Create(testCtx, wnc); err != nil {
		t.Fatalf("create: %v", err)
	}
	defer func() { _ = k8sClient.Delete(testCtx, wnc) }()
	if err := k8sClient.Get(testCtx, client.ObjectKey{Name: wncDefaultName}, wnc); err != nil {
		t.Fatalf("get: %v", err)
	}

	wnc.Status = &netv1alpha1.WorkloadNetworkConfigurationStatus{
		Conditions: []metav1.Condition{
			makeWNCCondition(testConditionReady, metav1.ConditionTrue),
			makeWNCCondition(testConditionReady, metav1.ConditionFalse),
		},
	}
	if err := k8sClient.Status().Update(testCtx, wnc); err == nil || !strings.Contains(err.Error(), "uplicate") {
		t.Fatalf("expected rejection containing %q, got: %v", "uplicate", err)
	}
}

func TestWorkloadNetworkConfiguration_ConditionReasonNotCamelCase_Rejected(t *testing.T) {
	wnc := vdsWNC()
	if err := k8sClient.Create(testCtx, wnc); err != nil {
		t.Fatalf("create: %v", err)
	}
	defer func() { _ = k8sClient.Delete(testCtx, wnc) }()
	if err := k8sClient.Get(testCtx, client.ObjectKey{Name: wncDefaultName}, wnc); err != nil {
		t.Fatalf("get: %v", err)
	}

	cond := makeWNCCondition(testConditionReady, metav1.ConditionTrue)
	cond.Reason = "not-camel-case"
	wnc.Status = &netv1alpha1.WorkloadNetworkConfigurationStatus{
		Conditions: []metav1.Condition{cond},
	}
	if err := k8sClient.Status().Update(testCtx, wnc); err == nil || !strings.Contains(err.Error(), "Invalid value") {
		t.Fatalf("expected rejection containing %q, got: %v", "Invalid value", err)
	}
}

func TestWorkloadNetworkConfiguration_ConditionStatusOutsideEnum_Rejected(t *testing.T) {
	wnc := vdsWNC()
	if err := k8sClient.Create(testCtx, wnc); err != nil {
		t.Fatalf("create: %v", err)
	}
	defer func() { _ = k8sClient.Delete(testCtx, wnc) }()
	if err := k8sClient.Get(testCtx, client.ObjectKey{Name: wncDefaultName}, wnc); err != nil {
		t.Fatalf("get: %v", err)
	}

	wnc.Status = &netv1alpha1.WorkloadNetworkConfigurationStatus{
		Conditions: []metav1.Condition{
			{
				Type:               testConditionReady,
				Status:             "bad-status",
				Reason:             testConditionReason,
				Message:            "msg",
				LastTransitionTime: metav1.NewTime(time.Now()),
			},
		},
	}
	if err := k8sClient.Status().Update(testCtx, wnc); err == nil || !strings.Contains(err.Error(), "Unsupported value") {
		t.Fatalf("expected rejection containing %q, got: %v", "Unsupported value", err)
	}
}

// -----------------------------------------------------------------------
// defaultNamespaceConfiguration
// Rule: self.type == 'vpc' || !has(self.defaultNamespaceConfiguration.vpcConfig)
// NetworkProviderDefaultConfig: +kubebuilder:validation:MinProperties=1
// DefaultVPCConfig.privateCIDRs: +optional, minItems=1, maxItems=16, CIDR pattern
// -----------------------------------------------------------------------

// vpcWNC builds a minimal valid vpc-provider WNC named "default".
func vpcWNC() *netv1alpha1.WorkloadNetworkConfiguration {
	return &netv1alpha1.WorkloadNetworkConfiguration{
		ObjectMeta: metav1.ObjectMeta{Name: wncDefaultName},
		Spec: netv1alpha1.WorkloadNetworkConfigurationSpec{
			Providers: []netv1alpha1.NetworkProviderEntry{
				{
					Type: netv1alpha1.NetworkProviderVPC,
					SystemConfiguration: &netv1alpha1.NamespaceNetworkConfig{
						VPCConfig: netv1alpha1.VPCConfig{
							AutoCreateConfig: netv1alpha1.AutoCreateVPCConfig{
								NSXProject:             testNSXProject,
								VPCConnectivityProfile: testVPCConnProfile,
								PrivateCIDRs:           []string{testCIDR1},
							},
						},
					},
				},
			},
			ActiveSystemProvider: netv1alpha1.NetworkProviderVPC,
		},
	}
}

func TestWorkloadNetworkConfiguration_ValidDefaultVPCConfig_Admitted(t *testing.T) {
	wnc := vpcWNC()
	wnc.Spec.Providers[0].DefaultNamespaceConfiguration = netv1alpha1.NetworkProviderDefaultConfig{
		VPCConfig: &netv1alpha1.DefaultVPCConfig{
			PrivateCIDRs: []string{"10.1.0.0/24"},
		},
	}
	if err := k8sClient.Create(testCtx, wnc); err != nil {
		t.Fatalf("expected admission, got: %v", err)
	}
	defer func() { _ = k8sClient.Delete(testCtx, wnc) }()
}

func TestWorkloadNetworkConfiguration_DefaultVPCConfigOnNonVPCType_Rejected(t *testing.T) {
	wnc := vdsWNC()
	wnc.Spec.Providers[0].DefaultNamespaceConfiguration = netv1alpha1.NetworkProviderDefaultConfig{
		VPCConfig: &netv1alpha1.DefaultVPCConfig{
			PrivateCIDRs: []string{"10.1.0.0/24"},
		},
	}
	err := k8sClient.Create(testCtx, wnc)
	if err == nil || !strings.Contains(err.Error(), "defaultNamespaceConfiguration.vpcConfig may only be set when type is vpc") {
		t.Fatalf("expected rejection for defaultNamespaceConfiguration.vpcConfig on non-vpc type, got: %v", err)
	}
}

func TestWorkloadNetworkConfiguration_EmptyDefaultNamespaceConfiguration_Rejected(t *testing.T) {
	obj := unstrWNC(wncDefaultName, map[string]interface{}{
		"activeSystemProvider": string(netv1alpha1.NetworkProviderVPC),
		"providers": []interface{}{
			map[string]interface{}{
				"type": string(netv1alpha1.NetworkProviderVPC),
				"systemConfiguration": map[string]interface{}{
					"vpcConfig": map[string]interface{}{
						"autoCreateConfig": map[string]interface{}{
							"nsxProject":             testNSXProject,
							"vpcConnectivityProfile": testVPCConnProfile,
						},
					},
				},
				"defaultNamespaceConfiguration": map[string]interface{}{},
			},
		},
	})
	if err := k8sClient.Create(testCtx, obj); !isRejected(err) {
		t.Fatalf("expected rejection for empty defaultNamespaceConfiguration (MinProperties=1), got: %v", err)
	}
}

// TestWorkloadNetworkConfiguration_EmptyPrivateCIDRsList_Rejected sends the
// request as Unstructured because the typed client's `omitempty` drops an
// empty (but non-nil) []string entirely on marshal, which would silently
// turn this into "field absent" instead of testing the MinItems=1 rule.
func TestWorkloadNetworkConfiguration_EmptyPrivateCIDRsList_Rejected(t *testing.T) {
	obj := unstrWNC(wncDefaultName, map[string]interface{}{
		"activeSystemProvider": string(netv1alpha1.NetworkProviderVPC),
		"providers": []interface{}{
			map[string]interface{}{
				"type": string(netv1alpha1.NetworkProviderVPC),
				"systemConfiguration": map[string]interface{}{
					"vpcConfig": map[string]interface{}{
						"autoCreateConfig": map[string]interface{}{
							"nsxProject":             testNSXProject,
							"vpcConnectivityProfile": testVPCConnProfile,
						},
					},
				},
				"defaultNamespaceConfiguration": map[string]interface{}{
					"vpcConfig": map[string]interface{}{
						"privateCIDRs": []interface{}{},
					},
				},
			},
		},
	})
	if err := k8sClient.Create(testCtx, obj); !isRejected(err) {
		t.Fatalf("expected rejection for explicit empty privateCIDRs list (MinItems=1), got: %v", err)
	}
}

func TestWorkloadNetworkConfiguration_DefaultPrivateCIDRsFullReplace_Admitted(t *testing.T) {
	wnc := vpcWNC()
	wnc.Spec.Providers[0].DefaultNamespaceConfiguration = netv1alpha1.NetworkProviderDefaultConfig{
		VPCConfig: &netv1alpha1.DefaultVPCConfig{
			PrivateCIDRs: []string{"10.1.0.0/24"},
		},
	}
	if err := k8sClient.Create(testCtx, wnc); err != nil {
		t.Fatalf("create: %v", err)
	}
	defer func() { _ = k8sClient.Delete(testCtx, wnc) }()

	latest := &netv1alpha1.WorkloadNetworkConfiguration{}
	if err := k8sClient.Get(testCtx, client.ObjectKeyFromObject(wnc), latest); err != nil {
		t.Fatalf("get: %v", err)
	}
	// Full replace with a CIDR that was never in the original list. Unlike
	// systemConfiguration.vpcConfig.autoCreateConfig.privateCIDRs (append-only),
	// defaultNamespaceConfiguration.vpcConfig.privateCIDRs has no append-only
	// constraint, so dropping 10.1.0.0/24 entirely must be admitted.
	latest.Spec.Providers[0].DefaultNamespaceConfiguration.VPCConfig.PrivateCIDRs = []string{"10.2.0.0/24"}
	if err := k8sClient.Update(testCtx, latest); err != nil {
		t.Fatalf("expected admission for full-replace of defaultNamespaceConfiguration privateCIDRs, got: %v", err)
	}
}
