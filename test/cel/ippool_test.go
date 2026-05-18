// Copyright (c) 2024-2026 Broadcom. All Rights Reserved.
// Broadcom Confidential. The term "Broadcom" refers to Broadcom Inc.
// and/or its subsidiaries.

package cel_test

import (
	"testing"

	netv1alpha1 "github.com/vmware-tanzu/net-operator-api/api/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func validIPPool(name string) *netv1alpha1.IPPool {
	return &netv1alpha1.IPPool{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Spec: netv1alpha1.IPPoolSpec{
			StartingAddress: "10.0.0.1",
			AddressCount:    5,
		},
	}
}

func TestIPPool_Valid_Admitted(t *testing.T) {
	obj := validIPPool("ippool-valid")
	if err := k8sClient.Create(testCtx, obj); err != nil {
		t.Fatalf("expected admission, got: %v", err)
	}
	_ = k8sClient.Delete(testCtx, obj)
}

// TestIPPool_EmptyStartingAddress_Rejected verifies that MinLength=1 rejects an
// empty startingAddress.  Format (IPv4 vs IPv6) is validated by the IPPool webhook.
func TestIPPool_EmptyStartingAddress_Rejected(t *testing.T) {
	obj := validIPPool("ippool-empty-addr")
	obj.Spec.StartingAddress = ""
	if err := k8sClient.Create(testCtx, obj); !isRejected(err) {
		t.Fatalf("expected rejection for empty startingAddress, got: %v", err)
	}
}

// TestIPPool_IPv6StartingAddress_Admitted verifies that an IPv6 address is admitted
// at the CRD level.  The webhook enforces any additional format constraints.
func TestIPPool_IPv6StartingAddress_Admitted(t *testing.T) {
	obj := validIPPool("ippool-ipv6")
	obj.Spec.StartingAddress = "fd00::1"
	if err := k8sClient.Create(testCtx, obj); err != nil {
		t.Fatalf("expected admission for IPv6 startingAddress, got: %v", err)
	}
	_ = k8sClient.Delete(testCtx, obj)
}

// TestIPPool_MalformedStartingAddress_Admitted verifies that a non-empty but
// syntactically invalid address (e.g. "not-an-ip") passes MinLength=1 and is
// admitted by CEL.  Format validation is deferred to the IPPool webhook (#23).
func TestIPPool_MalformedStartingAddress_Admitted(t *testing.T) {
	obj := validIPPool("ippool-malformed-addr")
	obj.Spec.StartingAddress = "not-an-ip"
	if err := k8sClient.Create(testCtx, obj); err != nil {
		t.Fatalf("expected admission for non-empty malformed startingAddress, got: %v", err)
	}
	_ = k8sClient.Delete(testCtx, obj)
}

func TestIPPool_ZeroAddressCount_Rejected(t *testing.T) {
	obj := validIPPool("ippool-zero-count")
	obj.Spec.AddressCount = 0
	if err := k8sClient.Create(testCtx, obj); !isRejected(err) {
		t.Fatalf("expected rejection for addressCount=0, got: %v", err)
	}
}

func TestIPPool_NegativeAddressCount_Rejected(t *testing.T) {
	obj := validIPPool("ippool-neg-count")
	obj.Spec.AddressCount = -1
	if err := k8sClient.Create(testCtx, obj); !isRejected(err) {
		t.Fatalf("expected rejection for addressCount=-1, got: %v", err)
	}
}
