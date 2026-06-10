// Copyright (c) 2024-2026 Broadcom. All Rights Reserved.
// Broadcom Confidential. The term "Broadcom" refers to Broadcom Inc.
// and/or its subsidiaries.

package cel_test

import (
	"testing"

	netv1alpha1 "github.com/vmware-tanzu/net-operator-api/api/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func validHAC(name string) *netv1alpha1.HAProxyLoadBalancerConfig {
	return &netv1alpha1.HAProxyLoadBalancerConfig{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Spec: netv1alpha1.HAProxyLoadBalancerConfigSpec{
			EndPointURLs: []string{"https://haproxy.example.com:5556/v2"},
			CredentialSecretRef: netv1alpha1.ClientSecretReference{
				Name:      "hac-creds",
				Namespace: "default",
			},
		},
	}
}

func TestHAProxyLoadBalancerConfig_Valid_Admitted(t *testing.T) {
	obj := validHAC("hac-valid")
	if err := k8sClient.Create(testCtx, obj); err != nil {
		t.Fatalf("expected admission, got: %v", err)
	}
	_ = k8sClient.Delete(testCtx, obj)
}

func TestHAProxyLoadBalancerConfig_EmptyEndPointURLs_Rejected(t *testing.T) {
	obj := validHAC("hac-empty-urls")
	// MinItems=1 on spec.endPointURLs must reject an empty list.
	obj.Spec.EndPointURLs = []string{}
	if err := k8sClient.Create(testCtx, obj); !isRejected(err) {
		t.Fatalf("expected rejection for empty endPointURLs (MinItems=1), got: %v", err)
	}
}

func TestHAProxyLoadBalancerConfig_EmptyItemInEndPointURLs_Rejected(t *testing.T) {
	obj := validHAC("hac-empty-item")
	// MinLength=1 on each EndPointURL item must reject an empty string entry.
	obj.Spec.EndPointURLs = []string{""}
	if err := k8sClient.Create(testCtx, obj); !isRejected(err) {
		t.Fatalf("expected rejection for empty string item in endPointURLs (MinLength=1), got: %v", err)
	}
}

func TestHAProxyLoadBalancerConfig_InvalidCredentialSecretRefName_Rejected(t *testing.T) {
	obj := validHAC("hac-bad-secret")
	// Uppercase letters are not allowed in DNS-1123 subdomains.
	obj.Spec.CredentialSecretRef.Name = "UPPER-CASE"
	if err := k8sClient.Create(testCtx, obj); !isRejected(err) {
		t.Fatalf("expected rejection for invalid credentialSecretRef.name, got: %v", err)
	}
}
