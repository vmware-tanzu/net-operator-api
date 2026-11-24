// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: Apache-2.0

package cel_test

import (
	"strings"
	"testing"

	netv1alpha1 "github.com/vmware-tanzu/net-operator-api/api/v1alpha1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func validNSXLoadBalancerConfig(name string) *netv1alpha1.NSXLoadBalancerConfig {
	return &netv1alpha1.NSXLoadBalancerConfig{
		ObjectMeta: metav1.ObjectMeta{Name: name},
		Spec: netv1alpha1.NSXLoadBalancerConfigSpec{
			CertificateSecretRef: netv1alpha1.ClientSecretReference{
				Name:      "nsx-ingress-tls",
				Namespace: "vmware-system-netop",
			},
		},
	}
}

func TestNSXLoadBalancerConfig_Valid_Admitted(t *testing.T) {
	obj := validNSXLoadBalancerConfig("nsxlbc-valid")
	if err := k8sClient.Create(testCtx, obj); err != nil {
		t.Fatalf("expected admission, got: %v", err)
	}
	t.Cleanup(func() { _ = k8sClient.Delete(testCtx, obj) })
}

func TestNSXLoadBalancerConfig_StatusSubresource_NotServed(t *testing.T) {
	obj := validNSXLoadBalancerConfig("nsxlbc-no-status")
	if err := k8sClient.Create(testCtx, obj); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = k8sClient.Delete(testCtx, obj) })
	if err := k8sClient.Status().Update(testCtx, obj); !apierrors.IsNotFound(err) {
		t.Fatalf("expected no status subresource, got: %v", err)
	}
}

func TestNSXLoadBalancerConfig_RequiredReference_Rejected(t *testing.T) {
	for _, tc := range []struct {
		name     string
		spec     map[string]interface{}
		message  string
		omitSpec bool
	}{
		{name: "omitted-spec", omitSpec: true, message: "spec: Required value"},
		{name: "omitted-ref", spec: map[string]interface{}{}, message: "certificateSecretRef: Required value"},
		{name: "empty-ref", spec: map[string]interface{}{"certificateSecretRef": map[string]interface{}{}}, message: "name: Required value"},
		{name: "omitted-name", spec: map[string]interface{}{"certificateSecretRef": map[string]interface{}{"namespace": "vmware-system-netop"}}, message: "name: Required value"},
		{name: "empty-name", spec: map[string]interface{}{"certificateSecretRef": map[string]interface{}{"name": "", "namespace": "vmware-system-netop"}}, message: "certificateSecretRef.name and namespace must be non-empty"},
		{name: "empty-namespace", spec: map[string]interface{}{"certificateSecretRef": map[string]interface{}{"name": "nsx-ingress-tls", "namespace": ""}}, message: "certificateSecretRef.name and namespace must be non-empty"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			obj := &unstructured.Unstructured{Object: map[string]interface{}{
				"apiVersion": "netoperator.vmware.com/v1alpha1",
				"kind":       "NSXLoadBalancerConfig",
				"metadata":   map[string]interface{}{"name": "nsxlbc-" + tc.name},
			}}
			if !tc.omitSpec {
				obj.Object["spec"] = tc.spec
			}
			err := k8sClient.Create(testCtx, obj)
			if !isRejected(err) || !strings.Contains(err.Error(), tc.message) {
				t.Fatalf("expected rejection containing %q, got: %v", tc.message, err)
			}
		})
	}
}

func TestNSXLoadBalancerConfig_OmittedNamespace_DefaultedOnReadback(t *testing.T) {
	obj := validNSXLoadBalancerConfig("nsxlbc-default-namespace")
	obj.Spec.CertificateSecretRef.Namespace = ""
	if err := k8sClient.Create(testCtx, obj); err != nil {
		t.Fatalf("expected admission with defaulted namespace, got: %v", err)
	}
	t.Cleanup(func() { _ = k8sClient.Delete(testCtx, obj) })
	got := &netv1alpha1.NSXLoadBalancerConfig{}
	if err := k8sClient.Get(testCtx, client.ObjectKeyFromObject(obj), got); err != nil {
		t.Fatal(err)
	}
	if got.Spec.CertificateSecretRef.Namespace != "default" {
		t.Fatalf("expected namespace default, got %q", got.Spec.CertificateSecretRef.Namespace)
	}
}

func TestNSXLoadBalancerConfig_InvalidSecretName_Rejected(t *testing.T) {
	for _, tc := range []struct {
		name  string
		value string
	}{
		{name: "uppercase", value: "Invalid"},
		{name: "slash", value: "invalid/name"},
		{name: "underscore", value: "invalid_name"},
		{name: "too-long", value: strings.Repeat("a", 254)},
	} {
		t.Run(tc.name, func(t *testing.T) {
			obj := validNSXLoadBalancerConfig("nsxlbc-name-" + tc.name)
			obj.Spec.CertificateSecretRef.Name = tc.value
			if err := k8sClient.Create(testCtx, obj); !isRejected(err) {
				t.Fatalf("expected invalid Secret name rejection, got: %v", err)
			}
		})
	}
}

func TestNSXLoadBalancerConfig_SecretReferenceMutable_Admitted(t *testing.T) {
	obj := validNSXLoadBalancerConfig("nsxlbc-ref-update")
	if err := k8sClient.Create(testCtx, obj); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = k8sClient.Delete(testCtx, obj) })
	obj.Spec.CertificateSecretRef = netv1alpha1.ClientSecretReference{Name: "replacement-tls", Namespace: "default"}
	if err := k8sClient.Update(testCtx, obj); err != nil {
		t.Fatalf("expected Secret reference update, got: %v", err)
	}
	got := &netv1alpha1.NSXLoadBalancerConfig{}
	if err := k8sClient.Get(testCtx, client.ObjectKeyFromObject(obj), got); err != nil {
		t.Fatal(err)
	}
	if got.Spec != obj.Spec {
		t.Fatalf("reference did not round-trip: got %+v, want %+v", got.Spec, obj.Spec)
	}
}
