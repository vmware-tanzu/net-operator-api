// Copyright (c) 2024-2026 Broadcom. All Rights Reserved.
// Broadcom Confidential. The term "Broadcom" refers to Broadcom Inc.
// and/or its subsidiaries.

package cel_test

import (
	"testing"

	netv1alpha1 "github.com/vmware-tanzu/net-operator-api/api/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/controller-runtime/pkg/client"
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

// TestAviLoadBalancerConfig_EmptyServer_Admitted verifies that API admits an empty Server string.
// spec.server has no omitempty tag, so "" is sent even if field is unset by a structured/typed client.
func TestAviLoadBalancerConfig_EmptyServer_Rejected(t *testing.T) {
	obj := validAviLoadBalancerConfig("avic-empty-server")
	obj.Spec.Server = ""
	if err := k8sClient.Create(testCtx, obj); err != nil {
		t.Fatalf("expected admission for empty server, got: %v", err)
	}
	_ = k8sClient.Delete(testCtx, obj)
}

// TestAviLoadBalancerConfig_OmittedCloudName_DefaultedOnReadback verifies that omitting cloudName
// via the typed Go struct (CloudName zero value + omitempty tag) is admitted, and that the field
// is defaulted to "Default-Cloud" on readback. The Go marshaler omits a zero-value CloudName from
// the JSON payload, so the API server applies the kubebuilder default before persisting.
func TestAviLoadBalancerConfig_OmittedCloudName_DefaultedOnReadback(t *testing.T) {
	obj := &netv1alpha1.AviLoadBalancerConfig{
		ObjectMeta: metav1.ObjectMeta{Name: "avic-omitted-cloud"},
		Spec: netv1alpha1.AviLoadBalancerConfigSpec{
			Server: "https://10.0.0.1",
			// CloudName is intentionally the zero value; omitempty omits it from the JSON
			// payload so the server applies the "Default-Cloud" default.
			CredentialSecretRef: netv1alpha1.ClientSecretReference{
				Name:      "avi-creds",
				Namespace: "default",
			},
		},
	}
	if err := k8sClient.Create(testCtx, obj); err != nil {
		t.Fatalf("expected admission for omitted cloudName, got: %v", err)
	}
	t.Cleanup(func() { _ = k8sClient.Delete(testCtx, obj) })

	got := &netv1alpha1.AviLoadBalancerConfig{}
	if err := k8sClient.Get(testCtx, client.ObjectKeyFromObject(obj), got); err != nil {
		t.Fatalf("failed to read back object: %v", err)
	}
	if got.Spec.CloudName != "Default-Cloud" {
		t.Errorf("expected cloudName defaulted to %q, got %q", "Default-Cloud", got.Spec.CloudName)
	}
}

// TestAviLoadBalancerConfig_ExplicitEmptyCloudName_DefaultedOnReadback verifies that a typed Go
// client that explicitly sets CloudName = "" is admitted and receives the "Default-Cloud" default
// on readback. Setting "" is indistinguishable from omitting the field: omitempty drops both from
// the JSON payload, so MinLength=1 never fires and the API server applies the kubebuilder default.
func TestAviLoadBalancerConfig_ExplicitEmptyCloudName_DefaultedOnReadback(t *testing.T) {
	obj := validAviLoadBalancerConfig("avic-explicit-empty-cloud")
	obj.Spec.CloudName = "" // explicitly zeroed — omitempty will omit this from the wire
	if err := k8sClient.Create(testCtx, obj); err != nil {
		t.Fatalf("expected admission for explicit empty cloudName via typed client, got: %v", err)
	}
	t.Cleanup(func() { _ = k8sClient.Delete(testCtx, obj) })

	got := &netv1alpha1.AviLoadBalancerConfig{}
	if err := k8sClient.Get(testCtx, client.ObjectKeyFromObject(obj), got); err != nil {
		t.Fatalf("failed to read back object: %v", err)
	}
	if got.Spec.CloudName != "Default-Cloud" {
		t.Errorf("expected cloudName defaulted to %q, got %q", "Default-Cloud", got.Spec.CloudName)
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
