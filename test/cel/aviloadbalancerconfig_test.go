// Copyright (c) 2024-2026 Broadcom. All Rights Reserved.
// Broadcom Confidential. The term "Broadcom" refers to Broadcom Inc.
// and/or its subsidiaries.

package cel_test

import (
	"testing"

	netv1alpha1 "github.com/vmware-tanzu/net-operator-api/api/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func validAviLoadBalancerConfig(name string) *netv1alpha1.AviLoadBalancerConfig {
	return &netv1alpha1.AviLoadBalancerConfig{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Spec: netv1alpha1.AviLoadBalancerConfigSpec{
			Server:    "https://10.0.0.1",
			CloudName: "Default-Cloud",
			CredentialSecretRef: netv1alpha1.ClientSecretReference{
				Name:      "avi-creds",
				Namespace: "default",
			},
		},
	}
}

func TestAviLoadBalancerConfig_Valid_Admitted(t *testing.T) {
	obj := validAviLoadBalancerConfig("avic-valid")
	if err := k8sClient.Create(testCtx, obj); err != nil {
		t.Fatalf("expected admission, got: %v", err)
	}
	_ = k8sClient.Delete(testCtx, obj)
}

// TestAviLoadBalancerConfig_EmptyServer_Rejected verifies that spec.server with MinLength=1 rejects
// an empty string.  spec.server has no omitempty tag, so "" is sent literally.
func TestAviLoadBalancerConfig_EmptyServer_Rejected(t *testing.T) {
	obj := validAviLoadBalancerConfig("avic-empty-server")
	obj.Spec.Server = ""
	if err := k8sClient.Create(testCtx, obj); !isRejected(err) {
		t.Fatalf("expected rejection for empty server, got: %v", err)
	}
}

// TestAviLoadBalancerConfig_EmptyCloudName_Rejected verifies that an explicit empty cloudName string
// is rejected by MinLength=1.  Because the Go struct field carries omitempty,
// the Go JSON marshaler omits an empty CloudName and the server applies the
// "Default-Cloud" default instead.  We use an Unstructured object to bypass this
// and send "cloudName": "" explicitly so that the server's default is not applied.
func TestAviLoadBalancerConfig_EmptyCloudName_Rejected(t *testing.T) {
	obj := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "netoperator.vmware.com/v1alpha1",
			"kind":       "AviLoadBalancerConfig",
			"metadata": map[string]interface{}{
				"name": "avic-empty-cloud",
			},
			"spec": map[string]interface{}{
				"server":    "https://10.0.0.1",
				"cloudName": "",
				"credentialSecretRef": map[string]interface{}{
					"name":      "avi-creds",
					"namespace": "default",
				},
			},
		},
	}
	if err := k8sClient.Create(testCtx, obj); !isRejected(err) {
		t.Fatalf("expected rejection for explicit empty cloudName, got: %v", err)
		_ = k8sClient.Delete(testCtx, obj)
	}
}

// TestAviLoadBalancerConfig_ValidCloudNameAndServer_Admitted verifies a typical non-default cloudName
// and explicit server URL are admitted.
func TestAviLoadBalancerConfig_ValidCloudNameAndServer_Admitted(t *testing.T) {
	obj := validAviLoadBalancerConfig("avic-named-cloud")
	obj.Spec.CloudName = "my-cloud"
	obj.Spec.Server = "https://10.1.2.3"
	if err := k8sClient.Create(testCtx, obj); err != nil {
		t.Fatalf("expected admission for valid cloudName and server, got: %v", err)
	}
	_ = k8sClient.Delete(testCtx, obj)
}

func TestAviLoadBalancerConfig_InvalidCredentialSecretRefName_Rejected(t *testing.T) {
	obj := validAviLoadBalancerConfig("avic-bad-secret")
	// Uppercase letters are not allowed in DNS-1123 subdomains.
	obj.Spec.CredentialSecretRef.Name = "Invalid-Secret-Name"
	if err := k8sClient.Create(testCtx, obj); !isRejected(err) {
		t.Fatalf("expected rejection for invalid credentialSecretRef.name, got: %v", err)
	}
}
