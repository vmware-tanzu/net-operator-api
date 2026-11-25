// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: Apache-2.0

package cel_test

import (
	"fmt"
	"slices"
	"strings"
	"testing"

	netv1alpha1 "github.com/vmware-tanzu/net-operator-api/api/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const dnsZonePathPrefix = "/orgs/default/projects/project1/dns-services/default/zones/"

func TestNamespaceNetworkConfiguration_DNSZonesLifecycle(t *testing.T) {
	for _, autoCreate := range []bool{false, true} {
		mode := "pre-created"
		if autoCreate {
			mode = "auto-created"
		}
		t.Run(mode, func(t *testing.T) {
			obj := vpcNNC("dns-zones-"+mode, testVPCPathFull)
			if autoCreate {
				obj = autoVpcNNC(obj.Name, "/orgs/default/projects/project1",
					"/orgs/default/projects/project1/vpc-connectivity-profiles/default")
			}
			if err := k8sClient.Create(testCtx, obj); err != nil {
				t.Fatalf("create without DNS zones: %v", err)
			}
			t.Cleanup(func() { _ = k8sClient.Delete(testCtx, obj) })

			for _, zones := range [][]string{
				{dnsZonePathPrefix + "zone1", dnsZonePathPrefix + "zone2"},
				{dnsZonePathPrefix + "zone2", dnsZonePathPrefix + "zone3"},
				{},
			} {
				obj.Spec.VPCConfig.DNSZones = slices.Clone(zones)
				if err := k8sClient.Update(testCtx, obj); err != nil {
					t.Fatalf("update DNS zones to %v: %v", zones, err)
				}
				got := &netv1alpha1.NamespaceNetworkConfiguration{}
				if err := k8sClient.Get(testCtx, client.ObjectKeyFromObject(obj), got); err != nil {
					t.Fatalf("read DNS zones: %v", err)
				}
				if !slices.Equal(got.Spec.VPCConfig.DNSZones, zones) {
					t.Fatalf("DNS zones were not persisted: got %v, want %v", got.Spec.VPCConfig.DNSZones, zones)
				}
				obj = got
			}
		})
	}
}

func TestNamespaceNetworkConfiguration_DNSZonesValidation(t *testing.T) {
	maxZones := make([]interface{}, 20)
	for i := range maxZones {
		maxZones[i] = fmt.Sprintf("%szone%d", dnsZonePathPrefix, i)
	}
	for _, tc := range []struct {
		name   string
		zones  []interface{}
		reject bool
	}{
		{name: "empty", zones: []interface{}{}},
		{name: "max-items", zones: maxZones},
		{name: "too-many-items", zones: append(slices.Clone(maxZones), dnsZonePathPrefix+"overflow"), reject: true},
		{name: "valid", zones: []interface{}{dnsZonePathPrefix + "zone1"}},
		{name: "max-length", zones: []interface{}{dnsZonePathPrefix + strings.Repeat("a", 2048-len(dnsZonePathPrefix))}},
		{name: "empty-path", zones: []interface{}{""}, reject: true},
		{name: "long-path", zones: []interface{}{strings.Repeat("a", 2049)}, reject: true},
		{name: "duplicates", zones: []interface{}{dnsZonePathPrefix + "zone1", dnsZonePathPrefix + "zone1"}, reject: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			obj := unstrNNC("dns-zones-validation-"+tc.name, map[string]interface{}{
				"type": "vpc",
				"vpcConfig": map[string]interface{}{
					"vpc":      testVPCPathFull,
					"dnsZones": tc.zones,
				},
			})
			err := k8sClient.Create(testCtx, obj)
			if err == nil {
				t.Cleanup(func() { _ = k8sClient.Delete(testCtx, obj) })
			}
			if tc.reject {
				if !isRejected(err) {
					t.Fatalf("expected DNS zones validation rejection, got: %v", err)
				}
			} else if err != nil {
				t.Fatalf("expected valid DNS zones to be admitted, got: %v", err)
			}
		})
	}
}

func TestWorkloadNetworkConfiguration_DNSZonesPersisted(t *testing.T) {
	zones := []string{dnsZonePathPrefix + "zone1"}
	obj := &netv1alpha1.WorkloadNetworkConfiguration{
		ObjectMeta: metav1.ObjectMeta{Name: wncDefaultName},
		Spec: netv1alpha1.WorkloadNetworkConfigurationSpec{
			Providers: []netv1alpha1.NetworkProviderEntry{{
				Type: netv1alpha1.NetworkProviderVPC,
				SystemConfiguration: &netv1alpha1.NamespaceNetworkConfig{
					VPCConfig: netv1alpha1.VPCConfig{VPC: testVPCPathFull, DNSZones: zones},
				},
			}},
			ActiveSystemProvider: netv1alpha1.NetworkProviderVPC,
		},
	}
	if err := k8sClient.Create(testCtx, obj); err != nil {
		t.Fatalf("create workload network configuration with DNS zones: %v", err)
	}
	t.Cleanup(func() { _ = k8sClient.Delete(testCtx, obj) })
	got := &netv1alpha1.WorkloadNetworkConfiguration{}
	if err := k8sClient.Get(testCtx, client.ObjectKeyFromObject(obj), got); err != nil {
		t.Fatalf("read workload network configuration: %v", err)
	}
	if len(got.Spec.Providers) != 1 || got.Spec.Providers[0].SystemConfiguration == nil {
		t.Fatalf("VPC system configuration was not persisted: %+v", got.Spec.Providers)
	}
	if actual := got.Spec.Providers[0].SystemConfiguration.VPCConfig.DNSZones; !slices.Equal(actual, zones) {
		t.Fatalf("DNS zones were not persisted: got %v, want %v", actual, zones)
	}
}
