// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: Apache-2.0

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

func TestWorkloadNetworkConfiguration_VPCEntryWithIPv6AndDualStack_Admitted(t *testing.T) {
	wnc := &netv1alpha1.WorkloadNetworkConfiguration{
		ObjectMeta: metav1.ObjectMeta{Name: wncDefaultName},
		Spec: netv1alpha1.WorkloadNetworkConfigurationSpec{
			Providers: []netv1alpha1.NetworkProviderEntry{
				{
					Type: netv1alpha1.NetworkProviderVPC,
					SystemConfiguration: &netv1alpha1.NamespaceNetworkConfig{
						VPCConfig: netv1alpha1.VPCConfig{
							DefaultIPv6PrefixLength: 64,
							AutoCreateConfig: netv1alpha1.AutoCreateVPCConfig{
								NSXProject:             "/infra/projects/proj-1",
								VPCConnectivityProfile: "/infra/vpc-profiles/prof-1",
								PrivateCIDRs:           []string{"10.0.0.0/16", "fd00:100:64::/48"},
							},
						},
					},
				},
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
// Services (DNS & NTP)
// spec.services: +optional
// spec.services.dns: +optional
// spec.services.dns.maxConcurrentForwards: +optional, Minimum=0, Maximum=65535
// spec.services.dns.servers: +optional, MinItems=1, MaxItems=10, items:MinLength=1, items:MaxLength=45, listType=set, isIP(self)
// spec.services.ntp: +optional
// spec.services.ntp.servers: +required, MinItems=1, MaxItems=10, items:MinLength=1, items:MaxLength=253, listType=set
// -----------------------------------------------------------------------

func TestWorkloadNetworkConfiguration_Services_Admitted(t *testing.T) {
	maxForwards1000 := int32(1000)
	maxForwards0 := int32(0)
	maxForwards65535 := int32(65535)

	tests := []struct {
		name     string
		services *netv1alpha1.WorkloadNetworkServicesConfig
	}{
		{
			name: "full services config with dns and ntp",
			services: &netv1alpha1.WorkloadNetworkServicesConfig{
				DNS: &netv1alpha1.WorkloadDNSConfig{
					Servers:               []string{"10.100.1.10", "10.100.1.11"},
					MaxConcurrentForwards: &maxForwards1000,
				},
				NTP: &netv1alpha1.WorkloadNTPConfig{
					Servers: []string{"ntp.corp.internal", "10.100.1.50"},
				},
			},
		},
		{
			name: "minimal services with only dns",
			services: &netv1alpha1.WorkloadNetworkServicesConfig{
				DNS: &netv1alpha1.WorkloadDNSConfig{
					Servers: []string{"10.100.1.10"},
				},
			},
		},
		{
			name: "minimal services with only ntp",
			services: &netv1alpha1.WorkloadNetworkServicesConfig{
				NTP: &netv1alpha1.WorkloadNTPConfig{
					Servers: []string{"ntp.corp.internal"},
				},
			},
		},
		{
			name: "dns with maxConcurrentForwards set to 0 (unconstrained)",
			services: &netv1alpha1.WorkloadNetworkServicesConfig{
				DNS: &netv1alpha1.WorkloadDNSConfig{
					MaxConcurrentForwards: &maxForwards0,
				},
			},
		},
		{
			name: "dns with maxConcurrentForwards set to 65535 (maximum)",
			services: &netv1alpha1.WorkloadNetworkServicesConfig{
				DNS: &netv1alpha1.WorkloadDNSConfig{
					MaxConcurrentForwards: &maxForwards65535,
				},
			},
		},
		{
			name: "services with IPv6 DNS and NTP servers",
			services: &netv1alpha1.WorkloadNetworkServicesConfig{
				DNS: &netv1alpha1.WorkloadDNSConfig{
					Servers: []string{"2001:db8::1", "2001:db8::2"},
				},
				NTP: &netv1alpha1.WorkloadNTPConfig{
					Servers: []string{"time.nist.gov", "2001:db8::10"},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wnc := vdsWNC()
			wnc.Spec.Services = tt.services
			if err := k8sClient.Create(testCtx, wnc); err != nil {
				t.Fatalf("expected admission, got: %v", err)
			}
			_ = k8sClient.Delete(testCtx, wnc)
		})
	}
}

func TestWorkloadNetworkConfiguration_EmptyServices_Rejected(t *testing.T) {
	obj := unstrWNC(wncDefaultName, map[string]interface{}{
		"providers":            []interface{}{vdsProviderEntry(wncSystemNetName)},
		"activeSystemProvider": string(netv1alpha1.NetworkProviderVSphereDistributed),
		"services":             map[string]interface{}{},
	})
	if err := k8sClient.Create(testCtx, obj); err == nil || !strings.Contains(err.Error(), "at least one of dns or ntp must be configured") {
		t.Fatalf("expected rejection containing %q, got: %v", "at least one of dns or ntp must be configured", err)
	}
}

func TestWorkloadNetworkConfiguration_EmptyDNSAndNTPObjects_Rejected(t *testing.T) {
	obj := unstrWNC(wncDefaultName, map[string]interface{}{
		"providers":            []interface{}{vdsProviderEntry(wncSystemNetName)},
		"activeSystemProvider": string(netv1alpha1.NetworkProviderVSphereDistributed),
		"services": map[string]interface{}{
			"dns": map[string]interface{}{},
			"ntp": map[string]interface{}{},
		},
	})
	if err := k8sClient.Create(testCtx, obj); err == nil || (!strings.Contains(err.Error(), "at least one of") && !strings.Contains(err.Error(), "servers") && !strings.Contains(err.Error(), "Required value")) {
		t.Fatalf("expected rejection containing 'at least one of' or 'Required value', got: %v", err)
	}
}

func TestWorkloadNetworkConfiguration_EmptyDNS_Rejected(t *testing.T) {
	obj := unstrWNC(wncDefaultName, map[string]interface{}{
		"providers":            []interface{}{vdsProviderEntry(wncSystemNetName)},
		"activeSystemProvider": string(netv1alpha1.NetworkProviderVSphereDistributed),
		"services": map[string]interface{}{
			"dns": map[string]interface{}{},
		},
	})
	if err := k8sClient.Create(testCtx, obj); err == nil || !strings.Contains(err.Error(), "at least one of") {
		t.Fatalf("expected rejection containing %q, got: %v", "at least one of", err)
	}
}

func TestWorkloadNetworkConfiguration_EmptyNTP_Rejected(t *testing.T) {
	obj := unstrWNC(wncDefaultName, map[string]interface{}{
		"providers":            []interface{}{vdsProviderEntry(wncSystemNetName)},
		"activeSystemProvider": string(netv1alpha1.NetworkProviderVSphereDistributed),
		"services": map[string]interface{}{
			"ntp": map[string]interface{}{},
		},
	})
	if err := k8sClient.Create(testCtx, obj); err == nil || (!strings.Contains(err.Error(), "servers") && !strings.Contains(err.Error(), "at least one of")) {
		t.Fatalf("expected rejection containing 'servers' or 'at least one of', got: %v", err)
	}
}

func TestWorkloadNetworkConfiguration_NegativeMaxConcurrentForwards_Rejected(t *testing.T) {
	negativeVal := int32(-1)
	wnc := vdsWNC()
	wnc.Spec.Services = &netv1alpha1.WorkloadNetworkServicesConfig{
		DNS: &netv1alpha1.WorkloadDNSConfig{
			MaxConcurrentForwards: &negativeVal,
		},
	}
	if err := k8sClient.Create(testCtx, wnc); err == nil || !strings.Contains(err.Error(), "greater than or equal to 0") {
		t.Fatalf("expected rejection containing %q, got: %v", "greater than or equal to 0", err)
	}
}

func TestWorkloadNetworkConfiguration_ExcessiveMaxConcurrentForwards_Rejected(t *testing.T) {
	excessiveVal := int32(65536)
	wnc := vdsWNC()
	wnc.Spec.Services = &netv1alpha1.WorkloadNetworkServicesConfig{
		DNS: &netv1alpha1.WorkloadDNSConfig{
			MaxConcurrentForwards: &excessiveVal,
		},
	}
	if err := k8sClient.Create(testCtx, wnc); err == nil || !strings.Contains(err.Error(), "less than or equal to 65535") {
		t.Fatalf("expected rejection containing %q, got: %v", "less than or equal to 65535", err)
	}
}

func TestWorkloadNetworkConfiguration_InvalidDNSServerIP_Rejected(t *testing.T) {
	wnc := vdsWNC()
	wnc.Spec.Services = &netv1alpha1.WorkloadNetworkServicesConfig{
		DNS: &netv1alpha1.WorkloadDNSConfig{
			Servers: []string{"not-an-ip"},
		},
	}
	if err := k8sClient.Create(testCtx, wnc); err == nil || !strings.Contains(err.Error(), "valid IPv4 or IPv6 address") {
		t.Fatalf("expected rejection containing %q, got: %v", "valid IPv4 or IPv6 address", err)
	}
}

func TestWorkloadNetworkConfiguration_EmptyDNSServersList_Rejected(t *testing.T) {
	obj := unstrWNC(wncDefaultName, map[string]interface{}{
		"providers":            []interface{}{vdsProviderEntry(wncSystemNetName)},
		"activeSystemProvider": string(netv1alpha1.NetworkProviderVSphereDistributed),
		"services": map[string]interface{}{
			"dns": map[string]interface{}{
				"servers": []interface{}{},
			},
		},
	})
	if err := k8sClient.Create(testCtx, obj); err == nil || !strings.Contains(err.Error(), "at least 1") {
		t.Fatalf("expected rejection containing %q, got: %v", "at least 1", err)
	}
}

func TestWorkloadNetworkConfiguration_DuplicateDNSServers_Rejected(t *testing.T) {
	wnc := vdsWNC()
	wnc.Spec.Services = &netv1alpha1.WorkloadNetworkServicesConfig{
		DNS: &netv1alpha1.WorkloadDNSConfig{
			Servers: []string{"10.100.1.10", "10.100.1.10"},
		},
	}
	if err := k8sClient.Create(testCtx, wnc); err == nil || !strings.Contains(err.Error(), "uplicate") {
		t.Fatalf("expected rejection containing %q, got: %v", "uplicate", err)
	}
}

func TestWorkloadNetworkConfiguration_ExcessiveDNSServers_Rejected(t *testing.T) {
	servers := []string{
		"10.0.0.1", "10.0.0.2", "10.0.0.3", "10.0.0.4", "10.0.0.5",
		"10.0.0.6", "10.0.0.7", "10.0.0.8", "10.0.0.9", "10.0.0.10", "10.0.0.11",
	}
	wnc := vdsWNC()
	wnc.Spec.Services = &netv1alpha1.WorkloadNetworkServicesConfig{
		DNS: &netv1alpha1.WorkloadDNSConfig{
			Servers: servers,
		},
	}
	if err := k8sClient.Create(testCtx, wnc); err == nil || !strings.Contains(err.Error(), "at most 10") {
		t.Fatalf("expected rejection containing %q, got: %v", "at most 10", err)
	}
}

func TestWorkloadNetworkConfiguration_EmptyNTPServersList_Rejected(t *testing.T) {
	obj := unstrWNC(wncDefaultName, map[string]interface{}{
		"providers":            []interface{}{vdsProviderEntry(wncSystemNetName)},
		"activeSystemProvider": string(netv1alpha1.NetworkProviderVSphereDistributed),
		"services": map[string]interface{}{
			"ntp": map[string]interface{}{
				"servers": []interface{}{},
			},
		},
	})
	if err := k8sClient.Create(testCtx, obj); err == nil || (!strings.Contains(err.Error(), "at least 1") && !strings.Contains(err.Error(), "servers must be specified")) {
		t.Fatalf("expected rejection containing 'at least 1' or 'servers must be specified', got: %v", err)
	}
}

func TestWorkloadNetworkConfiguration_DuplicateNTPServers_Rejected(t *testing.T) {
	wnc := vdsWNC()
	wnc.Spec.Services = &netv1alpha1.WorkloadNetworkServicesConfig{
		NTP: &netv1alpha1.WorkloadNTPConfig{
			Servers: []string{"ntp.corp.internal", "ntp.corp.internal"},
		},
	}
	if err := k8sClient.Create(testCtx, wnc); err == nil || !strings.Contains(err.Error(), "uplicate") {
		t.Fatalf("expected rejection containing %q, got: %v", "uplicate", err)
	}
}

func TestWorkloadNetworkConfiguration_ExcessiveNTPServers_Rejected(t *testing.T) {
	servers := []string{
		"ntp1.corp.internal", "ntp2.corp.internal", "ntp3.corp.internal", "ntp4.corp.internal",
		"ntp5.corp.internal", "ntp6.corp.internal", "ntp7.corp.internal", "ntp8.corp.internal",
		"ntp9.corp.internal", "ntp10.corp.internal", "ntp11.corp.internal",
	}
	wnc := vdsWNC()
	wnc.Spec.Services = &netv1alpha1.WorkloadNetworkServicesConfig{
		NTP: &netv1alpha1.WorkloadNTPConfig{
			Servers: servers,
		},
	}
	if err := k8sClient.Create(testCtx, wnc); err == nil || !strings.Contains(err.Error(), "at most 10") {
		t.Fatalf("expected rejection containing %q, got: %v", "at most 10", err)
	}
}

func TestWorkloadNetworkConfiguration_NTPServerItemTooLong_Rejected(t *testing.T) {
	longHostname := strings.Repeat("a", 254)
	wnc := vdsWNC()
	wnc.Spec.Services = &netv1alpha1.WorkloadNetworkServicesConfig{
		NTP: &netv1alpha1.WorkloadNTPConfig{
			Servers: []string{longHostname},
		},
	}
	if err := k8sClient.Create(testCtx, wnc); err == nil || !strings.Contains(err.Error(), "253") {
		t.Fatalf("expected rejection containing %q, got: %v", "253", err)
	}
}

