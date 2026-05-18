// Copyright (c) 2024-2026 Broadcom. All Rights Reserved.
// Broadcom Confidential. The term "Broadcom" refers to Broadcom Inc.
// and/or its subsidiaries.

package cel_test

import (
	"testing"

	netv1alpha1 "github.com/vmware-tanzu/net-operator-api/api/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
// is rejected by the MinLength=1 constraint on NetworkAddress.
func TestFoundationLoadBalancerConfig_EmptyDNSServer_Rejected(t *testing.T) {
	ensureNamespace(t, flbcNS)
	obj := validFLBC("flbc-empty-dns")
	obj.Spec.NetworkSpec.DNSServers = []netv1alpha1.IPAddress{""}
	if err := k8sClient.Create(testCtx, obj); !isRejected(err) {
		t.Fatalf("expected rejection for empty DNS server string, got: %v", err)
	}
}

// TestFoundationLoadBalancerConfig_NonEmptyDNSServer_Admitted verifies that non-IPv4 addresses (e.g. IPv6) are
// admitted — format validation is deferred to the FLBC webhook.
func TestFoundationLoadBalancerConfig_NonEmptyDNSServer_Admitted(t *testing.T) {
	ensureNamespace(t, flbcNS)
	obj := validFLBC("flbc-ipv6-dns")
	obj.Spec.NetworkSpec.DNSServers = []netv1alpha1.IPAddress{"2001:db8::1"}
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
	obj.Spec.NetworkSpec.DNSServers = []netv1alpha1.IPAddress{"not-an-ip"}
	if err := k8sClient.Create(testCtx, obj); err != nil {
		t.Fatalf("expected admission for non-empty malformed DNS server, got: %v", err)
	}
	_ = k8sClient.Delete(testCtx, obj)
}

// TestFoundationLoadBalancerConfig_EmptyVIPSubnet_Rejected verifies that an empty string item in
// virtualServerSubnets is rejected by the MinLength=1 constraint on NetworkCIDR.
func TestFoundationLoadBalancerConfig_EmptyVIPSubnet_Rejected(t *testing.T) {
	ensureNamespace(t, flbcNS)
	obj := validFLBC("flbc-empty-cidr")
	obj.Spec.NetworkSpec.VirtualServerSubnets = []netv1alpha1.NetworkCIDR{""}
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
	obj.Spec.NetworkSpec.VirtualServerSubnets = []netv1alpha1.NetworkCIDR{"not/a/cidr"}
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
