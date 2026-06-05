// Copyright (c) 2024-2026 Broadcom. All Rights Reserved.
// Broadcom Confidential. The term "Broadcom" refers to Broadcom Inc.
// and/or its subsidiaries.

package cel_test

import (
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

// TestFoundationLoadBalancerConfig_InvalidAvailabilityMode_Rejected verifies that a value outside the
// active-passive/single-node enum is rejected by OpenAPI validation.
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

// TestFoundationLoadBalancerConfig_EmptyDNSServer_Rejected verifies that an empty string item in dnsServers
// is rejected by the items:MinLength=1 constraint.
func TestFoundationLoadBalancerConfig_EmptyDNSServer_Rejected(t *testing.T) {
	ensureNamespace(t, flbcNS)
	obj := validFLBC("flbc-empty-dns")
	obj.Spec.NetworkSpec.DNSServers = []string{""}
	if err := k8sClient.Create(testCtx, obj); !isRejected(err) {
		t.Fatalf("expected rejection for empty DNS server string, got: %v", err)
	}
}

// TestFoundationLoadBalancerConfig_NonEmptyDNSServer_Admitted verifies that non-IPv4 addresses (e.g. IPv6) are
// admitted — format validation is deferred to the FLBC webhook.
func TestFoundationLoadBalancerConfig_NonEmptyDNSServer_Admitted(t *testing.T) {
	ensureNamespace(t, flbcNS)
	obj := validFLBC("flbc-ipv6-dns")
	obj.Spec.NetworkSpec.DNSServers = []string{"2001:db8::1"}
	if err := k8sClient.Create(testCtx, obj); err != nil {
		t.Fatalf("expected admission for non-IPv4 DNS server, got: %v", err)
	}
	_ = k8sClient.Delete(testCtx, obj)
}

// TestFoundationLoadBalancerConfig_MalformedDNSServer_Admitted verifies that a non-empty but malformed
// DNS server string (e.g. "not-an-ip") passes MinLength=1 and is admitted by CEL.
// Format validation (valid IP or hostname) is deferred to the FLBC webhook (#20).
func TestFoundationLoadBalancerConfig_MalformedDNSServer_Admitted(t *testing.T) {
	ensureNamespace(t, flbcNS)
	obj := validFLBC("flbc-malformed-dns")
	obj.Spec.NetworkSpec.DNSServers = []string{"not-an-ip"}
	if err := k8sClient.Create(testCtx, obj); err != nil {
		t.Fatalf("expected admission for non-empty malformed DNS server, got: %v", err)
	}
	_ = k8sClient.Delete(testCtx, obj)
}

// TestFoundationLoadBalancerConfig_EmptyVIPSubnet_Rejected verifies that an empty string item in
// virtualServerSubnets is rejected by the items:MinLength=1 constraint.
func TestFoundationLoadBalancerConfig_EmptyVIPSubnet_Rejected(t *testing.T) {
	ensureNamespace(t, flbcNS)
	obj := validFLBC("flbc-empty-cidr")
	obj.Spec.NetworkSpec.VirtualServerSubnets = []string{""}
	if err := k8sClient.Create(testCtx, obj); !isRejected(err) {
		t.Fatalf("expected rejection for empty VIP subnet string, got: %v", err)
	}
}

// TestFoundationLoadBalancerConfig_MalformedVIPSubnet_Admitted verifies that a non-empty but malformed CIDR
// string (e.g. "not/a/cidr") passes MinLength=1 and is admitted by CEL.
// Format validation (valid CIDR) is deferred to the FLBC webhook (#20).
func TestFoundationLoadBalancerConfig_MalformedVIPSubnet_Admitted(t *testing.T) {
	ensureNamespace(t, flbcNS)
	obj := validFLBC("flbc-malformed-cidr")
	obj.Spec.NetworkSpec.VirtualServerSubnets = []string{"not/a/cidr"}
	if err := k8sClient.Create(testCtx, obj); err != nil {
		t.Fatalf("expected admission for non-empty malformed CIDR, got: %v", err)
	}
	_ = k8sClient.Delete(testCtx, obj)
}

func TestFoundationLoadBalancerConfig_EmptyVirtualServerIPPools_Rejected(t *testing.T) {
	ensureNamespace(t, flbcNS)
	obj := validFLBC("flbc-empty-pools")
	obj.Spec.NetworkSpec.VirtualServerIPPools = []netv1alpha1.IPPoolReference{}
	if err := k8sClient.Create(testCtx, obj); !isRejected(err) {
		t.Fatalf("expected rejection for empty virtualServerIPPools, got: %v", err)
	}
}

func TestFoundationLoadBalancerConfig_DowngradeAvailabilityMode_Rejected(t *testing.T) {
	ensureNamespace(t, flbcNS)
	obj := validFLBC("flbc-downgrade")
	if err := k8sClient.Create(testCtx, obj); err != nil {
		t.Fatalf("create: %v", err)
	}
	defer func() { _ = k8sClient.Delete(testCtx, obj) }()

	// Fetch the latest resourceVersion before update.
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

// TestFoundationLoadBalancerConfig_TooManyZones_Rejected verifies that providing more than 256
// zones is rejected by the MaxItems=256 constraint.
func TestFoundationLoadBalancerConfig_TooManyZones_Rejected(t *testing.T) {
	ensureNamespace(t, flbcNS)
	obj := validFLBC("flbc-zones-overflow")
	obj.Spec.DeploymentSpec.Zones = makeZones(257)
	if err := k8sClient.Create(testCtx, obj); !isRejected(err) {
		t.Fatalf("expected rejection for >256 zones, got: %v", err)
	}
}

// TestFoundationLoadBalancerConfig_ZonesAtMaxCount_Admitted verifies that exactly 256 zones are
// admitted (boundary condition for MaxItems=256).
func TestFoundationLoadBalancerConfig_ZonesAtMaxCount_Admitted(t *testing.T) {
	ensureNamespace(t, flbcNS)
	obj := validFLBC("flbc-zones-maxcount")
	obj.Spec.DeploymentSpec.Zones = makeZones(256)
	if err := k8sClient.Create(testCtx, obj); err != nil {
		t.Fatalf("expected admission for 256 zones, got: %v", err)
	}
	_ = k8sClient.Delete(testCtx, obj)
}

// TestFoundationLoadBalancerConfig_ZoneNameTooLong_Rejected verifies that a zone name exceeding
// 253 characters is rejected by the items:MaxLength=253 constraint.
func TestFoundationLoadBalancerConfig_ZoneNameTooLong_Rejected(t *testing.T) {
	ensureNamespace(t, flbcNS)
	obj := validFLBC("flbc-zone-toolong")
	obj.Spec.DeploymentSpec.Zones = []string{strings.Repeat("z", 254)}
	if err := k8sClient.Create(testCtx, obj); !isRejected(err) {
		t.Fatalf("expected rejection for zone name >253 chars, got: %v", err)
	}
}

// TestFoundationLoadBalancerConfig_ZoneNameAtMaxLength_Admitted verifies that a zone name of
// exactly 253 characters is admitted (boundary condition).
func TestFoundationLoadBalancerConfig_ZoneNameAtMaxLength_Admitted(t *testing.T) {
	ensureNamespace(t, flbcNS)
	obj := validFLBC("flbc-zone-maxlen")
	obj.Spec.DeploymentSpec.Zones = []string{strings.Repeat("z", 253)}
	if err := k8sClient.Create(testCtx, obj); err != nil {
		t.Fatalf("expected admission for zone name of 253 chars, got: %v", err)
	}
	_ = k8sClient.Delete(testCtx, obj)
}

// TestFoundationLoadBalancerConfig_NoZones_Admitted verifies that omitting zones is valid;
// net-operator will treat all supervisor AvailabilityZones as eligible placement targets.
func TestFoundationLoadBalancerConfig_NoZones_Admitted(t *testing.T) {
	ensureNamespace(t, flbcNS)
	obj := validFLBC("flbc-no-zones")
	obj.Spec.DeploymentSpec.Zones = nil
	if err := k8sClient.Create(testCtx, obj); err != nil {
		t.Fatalf("expected admission without zones, got: %v", err)
	}
	_ = k8sClient.Delete(testCtx, obj)
}

// TestFoundationLoadBalancerConfig_ExplicitZones_Admitted verifies that supplying one or more
// named zones is still admitted.
func TestFoundationLoadBalancerConfig_ExplicitZones_Admitted(t *testing.T) {
	ensureNamespace(t, flbcNS)
	obj := validFLBC("flbc-explicit-zones")
	obj.Spec.DeploymentSpec.Zones = []string{"zone-a", "zone-b"}
	if err := k8sClient.Create(testCtx, obj); err != nil {
		t.Fatalf("expected admission with explicit zones, got: %v", err)
	}
	_ = k8sClient.Delete(testCtx, obj)
}

// TestFoundationLoadBalancerConfig_EmptyNTPServer_Rejected verifies that an empty string item in
// ntpServers is rejected by the items:MinLength=1 constraint.
func TestFoundationLoadBalancerConfig_EmptyNTPServer_Rejected(t *testing.T) {
	ensureNamespace(t, flbcNS)
	obj := validFLBC("flbc-empty-ntp")
	obj.Spec.NetworkSpec.NTPServers = []string{""}
	if err := k8sClient.Create(testCtx, obj); !isRejected(err) {
		t.Fatalf("expected rejection for empty NTP server string, got: %v", err)
	}
}

// TestFoundationLoadBalancerConfig_NonEmptyNTPServer_Admitted verifies that a non-empty NTP server
// hostname is admitted — format validation is deferred to the FLBC webhook.
func TestFoundationLoadBalancerConfig_NonEmptyNTPServer_Admitted(t *testing.T) {
	ensureNamespace(t, flbcNS)
	obj := validFLBC("flbc-valid-ntp")
	obj.Spec.NetworkSpec.NTPServers = []string{"ntp.example.com"}
	if err := k8sClient.Create(testCtx, obj); err != nil {
		t.Fatalf("expected admission for non-empty NTP server, got: %v", err)
	}
	_ = k8sClient.Delete(testCtx, obj)
}

// TestFoundationLoadBalancerConfig_NoStoragePolicy_Admitted verifies that omitting storagePolicy
// is valid; the supervisor control plane's storage policy is used as the default.
func TestFoundationLoadBalancerConfig_NoStoragePolicy_Admitted(t *testing.T) {
	ensureNamespace(t, flbcNS)
	obj := validFLBC("flbc-no-storage-policy")
	obj.Spec.DeploymentSpec.StoragePolicy = ""
	if err := k8sClient.Create(testCtx, obj); err != nil {
		t.Fatalf("expected admission without storagePolicy, got: %v", err)
	}
	_ = k8sClient.Delete(testCtx, obj)
}

// TestFoundationLoadBalancerConfig_EmptyStoragePolicy_Rejected verifies that an explicitly empty
// storagePolicy string is rejected by MinLength=1.  Because the Go struct field carries omitempty,
// the Go JSON marshaler omits an empty StoragePolicy; we use an Unstructured object to bypass this
// and send "storagePolicy": "" so the server's MinLength constraint is exercised directly.
func TestFoundationLoadBalancerConfig_EmptyStoragePolicy_Rejected(t *testing.T) {
	ensureNamespace(t, flbcNS)
	obj := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "netoperator.vmware.com/v1alpha1",
			"kind":       "FoundationLoadBalancerConfig",
			"metadata": map[string]interface{}{
				"name":      "flbc-empty-storage",
				"namespace": flbcNS,
			},
			"spec": map[string]interface{}{
				"deploymentSpec": map[string]interface{}{
					"size":             "small",
					"availabilityMode": "active-passive",
					"storagePolicy":    "",
					"zones":            []interface{}{"zone-a"},
					"activePassiveSpec": map[string]interface{}{
						"replicas": int64(2),
					},
				},
				"virtualIPNetwork": map[string]interface{}{
					"name": "vip-net",
				},
				"networkSpec": map[string]interface{}{
					"virtualServerIPPools": []interface{}{
						map[string]interface{}{"name": "pool-1"},
					},
				},
			},
		},
	}
	if err := k8sClient.Create(testCtx, obj); !isRejected(err) {
		t.Fatalf("expected rejection for explicit empty storagePolicy, got: %v", err)
		_ = k8sClient.Delete(testCtx, obj)
	}
}
