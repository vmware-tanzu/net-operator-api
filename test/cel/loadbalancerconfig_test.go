// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: Apache-2.0

package cel_test

import (
	"reflect"
	"strings"
	"testing"

	netv1alpha1 "github.com/vmware-tanzu/net-operator-api/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func validFoundationLBC(name string) *netv1alpha1.LoadBalancerConfig {
	return &netv1alpha1.LoadBalancerConfig{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Spec: netv1alpha1.LoadBalancerConfigSpec{
			Type: netv1alpha1.LoadBalancerConfigTypeFoundation,
			ProviderRef: netv1alpha1.LoadBalancerConfigProviderReference{
				APIGroup: "netoperator.vmware.com",
				Kind:     "FoundationLoadBalancerConfig",
				Name:     "my-flb",
			},
		},
	}
}

func validAviLBC(name string) *netv1alpha1.LoadBalancerConfig {
	return &netv1alpha1.LoadBalancerConfig{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Spec: netv1alpha1.LoadBalancerConfigSpec{
			Type: netv1alpha1.LoadBalancerConfigTypeAvi,
			ProviderRef: netv1alpha1.LoadBalancerConfigProviderReference{
				APIGroup: "netoperator.vmware.com",
				Kind:     "AviLoadBalancerConfig",
				Name:     "my-avi",
			},
		},
	}
}

func validHAProxyLBC(name string) *netv1alpha1.LoadBalancerConfig {
	return &netv1alpha1.LoadBalancerConfig{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Spec: netv1alpha1.LoadBalancerConfigSpec{
			Type: netv1alpha1.LoadBalancerConfigTypeHAProxy,
			ProviderRef: netv1alpha1.LoadBalancerConfigProviderReference{
				APIGroup: "netoperator.vmware.com",
				Kind:     "HAProxyLoadBalancerConfig",
				Name:     "my-haproxy",
			},
		},
	}
}

func validNSXLBC(name string) *netv1alpha1.LoadBalancerConfig {
	return &netv1alpha1.LoadBalancerConfig{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Spec: netv1alpha1.LoadBalancerConfigSpec{
			Type: netv1alpha1.LoadBalancerConfigTypeNSX,
		},
	}
}

// --- Basic Validity & Creation Tests ---

func TestLoadBalancerConfig_Valid_Admitted(t *testing.T) {
	for _, tc := range []struct {
		name string
		obj  *netv1alpha1.LoadBalancerConfig
	}{
		{"foundation", validFoundationLBC("lbc-valid-flb")},
		{"avi", validAviLBC("lbc-valid-avi")},
		{"haproxy", validHAProxyLBC("lbc-valid-haproxy")},
		{"nsx", validNSXLBC("lbc-valid-nsx")},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if err := k8sClient.Create(testCtx, tc.obj); err != nil {
				t.Fatalf("expected admission for %s, got: %v", tc.name, err)
			}
			_ = k8sClient.Delete(testCtx, tc.obj)
		})
	}
}

func TestLoadBalancerConfig_RequiredProviderRef_Rejected(t *testing.T) {
	for _, provider := range []string{"foundation", "avi", "haproxy"} {
		t.Run(provider, func(t *testing.T) {
			obj := lbcWithoutReference("lbc-no-ref-"+provider, provider)
			err := k8sClient.Create(testCtx, obj)
			requireLBCRejection(t, err, "spec.providerRef is required except for nsx")
			if strings.Contains(err.Error(), "no such key") {
				t.Fatalf("unexpected CEL evaluation error: %v", err)
			}
		})
	}
}

func TestLoadBalancerConfig_ProviderRefKindMismatch_Rejected(t *testing.T) {
	for _, newObject := range []func(string) *netv1alpha1.LoadBalancerConfig{
		validFoundationLBC, validAviLBC, validHAProxyLBC,
	} {
		for _, kind := range []string{"FoundationLoadBalancerConfig", "AviLoadBalancerConfig", "HAProxyLoadBalancerConfig", "HAProxy", "Unknown", ""} {
			obj := newObject("lbc-kind-mismatch")
			if obj.Spec.ProviderRef.Kind == kind || (obj.Spec.Type == netv1alpha1.LoadBalancerConfigTypeHAProxy && kind == "HAProxy") {
				continue
			}
			t.Run(string(obj.Spec.Type)+"/"+kind, func(t *testing.T) {
				obj.Spec.ProviderRef.Kind = kind
				requireLBCRejection(t, k8sClient.Create(testCtx, obj), "providerRef.kind must match spec.type")
			})
		}
	}
}

func TestLoadBalancerConfig_InvalidTypeEnum_Rejected(t *testing.T) {
	obj := validFoundationLBC("lbc-invalid-type")
	obj.Spec.Type = "unsupported-provider"
	err := k8sClient.Create(testCtx, obj)
	requireLBCRejection(t, err, "spec.type: Unsupported value")
}

// --- Immutability & Update Tests ---

func TestLoadBalancerConfig_AviToNSX_ProviderRefRemoved(t *testing.T) {
	obj := validAviLBC("lbc-avi-to-nsx")
	if err := k8sClient.Create(testCtx, obj); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = k8sClient.Delete(testCtx, obj) }()
	patch := client.RawPatch(types.MergePatchType, []byte(`{"spec":{"type":"nsx","providerRef":null}}`))
	if err := k8sClient.Patch(testCtx, obj, patch); err != nil {
		t.Fatalf("switch from Avi to NSX while removing provider reference: %v", err)
	}
	wire := lbcWithoutReference(obj.Name, "nsx")
	if err := k8sClient.Get(testCtx, client.ObjectKeyFromObject(obj), wire); err != nil {
		t.Fatal(err)
	}
	if _, found, err := unstructured.NestedFieldNoCopy(wire.Object, "spec", "providerRef"); err != nil || found {
		t.Fatalf("expected omitted providerRef, found=%v err=%v", found, err)
	}
}

func TestLoadBalancerConfig_NSXToAvi_Transition_Admitted(t *testing.T) {
	obj := lbcWithoutReference("lbc-nsx-switch", "nsx")
	if err := k8sClient.Create(testCtx, obj); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = k8sClient.Delete(testCtx, obj) }()
	if err := unstructured.SetNestedField(obj.Object, "avi", "spec", "type"); err != nil {
		t.Fatal(err)
	}
	requireLBCRejection(t, k8sClient.Update(testCtx, obj), "spec.providerRef is required except for nsx")
	invalidRef := map[string]interface{}{"apiGroup": "netoperator.vmware.com", "kind": "FoundationLoadBalancerConfig", "name": "wrong-provider"}
	if err := unstructured.SetNestedMap(obj.Object, invalidRef, "spec", "providerRef"); err != nil {
		t.Fatal(err)
	}
	requireLBCRejection(t, k8sClient.Update(testCtx, obj), "providerRef.kind must match spec.type")
	ref := map[string]interface{}{"apiGroup": "netoperator.vmware.com", "apiVersion": "v1alpha1", "kind": "AviLoadBalancerConfig", "name": "my-avi"}
	if err := unstructured.SetNestedMap(obj.Object, ref, "spec", "providerRef"); err != nil {
		t.Fatal(err)
	}
	if err := k8sClient.Update(testCtx, obj); err != nil {
		t.Fatalf("switch from NSX LB to registered Avi: %v", err)
	}
}

func TestLoadBalancerConfig_ProviderRefMutable_Admitted(t *testing.T) {
	for _, tc := range []struct {
		name   string
		mutate func(*netv1alpha1.LoadBalancerConfigProviderReference)
	}{
		{"name", func(ref *netv1alpha1.LoadBalancerConfigProviderReference) { ref.Name = "different-name" }},
		{"api-group", func(ref *netv1alpha1.LoadBalancerConfigProviderReference) { ref.APIGroup = "other.example.com" }},
		{"api-version", func(ref *netv1alpha1.LoadBalancerConfigProviderReference) { ref.APIVersion = "v1beta1" }},
		{"api-version-removal", func(ref *netv1alpha1.LoadBalancerConfigProviderReference) { ref.APIVersion = "" }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			obj := validFoundationLBC("lbc-ref-mutable-" + tc.name)
			obj.Spec.ProviderRef.APIVersion = "v1alpha1"
			if err := k8sClient.Create(testCtx, obj); err != nil {
				t.Fatal(err)
			}
			defer func() { _ = k8sClient.Delete(testCtx, obj) }()
			if err := k8sClient.Get(testCtx, client.ObjectKeyFromObject(obj), obj); err != nil {
				t.Fatalf("get before update: %v", err)
			}
			tc.mutate(&obj.Spec.ProviderRef)
			if err := k8sClient.Update(testCtx, obj); err != nil {
				t.Fatalf("expected provider reference update to be admitted: %v", err)
			}
		})
	}
}

// Use unstructured objects to exercise actual field absence: the Go value-typed
// ProviderRef is serialized even when it is zero-valued and tagged omitempty.
func lbcWithoutReference(name, provider string) *unstructured.Unstructured {
	return &unstructured.Unstructured{Object: map[string]interface{}{
		"apiVersion": "netoperator.vmware.com/v1alpha1", "kind": "LoadBalancerConfig",
		"metadata": map[string]interface{}{"name": name},
		"spec":     map[string]interface{}{"type": provider},
	}}
}

func TestLoadBalancerConfig_OmittedProviderRef_Admitted(t *testing.T) {
	obj := lbcWithoutReference("lbc-omitted-nsx", "nsx")
	if err := k8sClient.Create(testCtx, obj); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = k8sClient.Delete(testCtx, obj) }()
	assertUnstructuredLBCSpecRoundtrip(t, obj)
}

func TestLoadBalancerConfig_TypeTransitions_Rejected(t *testing.T) {
	providers := []netv1alpha1.LoadBalancerConfigType{
		netv1alpha1.LoadBalancerConfigTypeFoundation, netv1alpha1.LoadBalancerConfigTypeAvi,
		netv1alpha1.LoadBalancerConfigTypeHAProxy, netv1alpha1.LoadBalancerConfigTypeNSX,
	}
	for _, from := range providers {
		for _, to := range providers {
			if from == to ||
				(from == netv1alpha1.LoadBalancerConfigTypeNSX && to == netv1alpha1.LoadBalancerConfigTypeAvi) ||
				(from == netv1alpha1.LoadBalancerConfigTypeAvi && to == netv1alpha1.LoadBalancerConfigTypeNSX) {
				continue
			}
			t.Run(string(from)+"-to-"+string(to), func(t *testing.T) {
				obj := validFoundationLBC("lbc-transition-" + string(from) + "-" + string(to))
				obj.Spec.Type = from
				kinds := map[netv1alpha1.LoadBalancerConfigType]string{
					netv1alpha1.LoadBalancerConfigTypeFoundation: "FoundationLoadBalancerConfig",
					netv1alpha1.LoadBalancerConfigTypeAvi:        "AviLoadBalancerConfig",
					netv1alpha1.LoadBalancerConfigTypeHAProxy:    "HAProxyLoadBalancerConfig",
				}
				if kind, ok := kinds[from]; ok {
					obj.Spec.ProviderRef.Kind = kind
				}
				if err := k8sClient.Create(testCtx, obj); err != nil {
					t.Fatal(err)
				}
				defer func() { _ = k8sClient.Delete(testCtx, obj) }()
				if err := k8sClient.Get(testCtx, client.ObjectKeyFromObject(obj), obj); err != nil {
					t.Fatalf("get before update: %v", err)
				}
				obj.Spec.Type = to
				if kind, ok := kinds[to]; ok {
					obj.Spec.ProviderRef.Kind = kind
				}
				err := k8sClient.Update(testCtx, obj)
				if !isRejected(err) || !strings.Contains(err.Error(), "spec.type is immutable") {
					t.Fatalf("expected type immutability rejection, got: %v", err)
				}
			})
		}
	}
}

func TestLoadBalancerConfig_WCPSVCReconciliation_Admitted(t *testing.T) {
	for _, factory := range []func(string) *netv1alpha1.LoadBalancerConfig{validFoundationLBC, validAviLBC, validHAProxyLBC} {
		obj := factory("lbc-wcpsvc")
		t.Run(string(obj.Spec.Type), func(t *testing.T) {
			// Current wcpsvc writers use the LBC name as the reference name and include apiVersion.
			obj.Spec.ProviderRef.Name = obj.Name
			obj.Spec.ProviderRef.APIVersion = "v1alpha1"
			expected := obj.Spec
			if err := k8sClient.Create(testCtx, obj); err != nil {
				t.Fatal(err)
			}
			defer func() { _ = k8sClient.Delete(testCtx, obj) }()
			latest := getLBC(t, obj)
			latest.Spec = expected
			if err := k8sClient.Update(testCtx, latest); err != nil {
				t.Fatalf("idempotent reconciliation: %v", err)
			}
			if got := getLBC(t, obj).Spec; got != expected {
				t.Fatalf("spec: got %+v, want %+v", got, expected)
			}
		})
	}
}

// Both current providers and the legacy HAProxy kind must permit controller writes.
func controllerLBCs() []*netv1alpha1.LoadBalancerConfig {
	legacy := validHAProxyLBC("lbc-legacy-haproxy")
	legacy.Spec.ProviderRef.Kind = "HAProxy"
	return []*netv1alpha1.LoadBalancerConfig{validFoundationLBC("lbc-foundation"), validAviLBC("lbc-avi"), validHAProxyLBC("lbc-haproxy"), legacy, validNSXLBC("lbc-nsx")}
}

func getLBC(t *testing.T, obj *netv1alpha1.LoadBalancerConfig) *netv1alpha1.LoadBalancerConfig {
	t.Helper()
	latest := &netv1alpha1.LoadBalancerConfig{}
	if err := k8sClient.Get(testCtx, client.ObjectKeyFromObject(obj), latest); err != nil {
		t.Fatalf("get: %v", err)
	}
	return latest
}

func TestLoadBalancerConfig_Metadata_Roundtrip(t *testing.T) {
	for _, obj := range controllerLBCs() {
		t.Run(obj.Name, func(t *testing.T) {
			obj.Spec.ProviderRef.APIVersion = "v1alpha1"
			expectedSpec := obj.Spec
			if err := k8sClient.Create(testCtx, obj); err != nil {
				t.Fatal(err)
			}
			// Remove finalizers even if a round-trip assertion fails.
			t.Cleanup(func() {
				latest := &netv1alpha1.LoadBalancerConfig{}
				if err := k8sClient.Get(testCtx, client.ObjectKeyFromObject(obj), latest); err == nil {
					latest.Finalizers = nil
					_ = k8sClient.Update(testCtx, latest)
					_ = k8sClient.Delete(testCtx, latest)
				}
			})
			latest := getLBC(t, obj)
			before := latest.DeepCopy()
			latest.Labels = map[string]string{"test": "mutable"}
			latest.Annotations = map[string]string{"wcp.vmware.com/ingress-tls-secret": "my-secret"}
			latest.Finalizers = []string{"netoperator.vmware.com/test"}
			if err := k8sClient.Patch(testCtx, latest, client.MergeFrom(before)); err != nil {
				t.Fatalf("metadata patch: %v", err)
			}
			updated := getLBC(t, obj)
			if !reflect.DeepEqual(updated.Labels, latest.Labels) || !reflect.DeepEqual(updated.Annotations, latest.Annotations) || !reflect.DeepEqual(updated.Finalizers, latest.Finalizers) {
				t.Fatalf("metadata did not round-trip: %+v", updated.ObjectMeta)
			}
			if updated.Spec != expectedSpec {
				t.Fatalf("metadata patch changed spec: %+v", updated.Spec)
			}
			updated.Finalizers = nil
			if err := k8sClient.Update(testCtx, updated); err != nil {
				t.Fatalf("finalizer removal: %v", err)
			}
			if got := getLBC(t, obj); len(got.Finalizers) != 0 || got.Spec != expectedSpec {
				t.Fatalf("unexpected object after finalizer removal: %+v", got)
			}
		})
	}
}

func TestLoadBalancerConfig_Conditions_StatusRoundtrip(t *testing.T) {
	for _, obj := range controllerLBCs() {
		t.Run(obj.Name, func(t *testing.T) {
			obj.Spec.ProviderRef.APIVersion = "v1alpha1"
			expectedSpec := obj.Spec
			if err := k8sClient.Create(testCtx, obj); err != nil {
				t.Fatal(err)
			}
			defer func() { _ = k8sClient.Delete(testCtx, obj) }()
			latest := getLBC(t, obj)
			if len(latest.Status.Conditions) != 0 {
				t.Fatalf("expected no initial conditions, got: %+v", latest.Status.Conditions)
			}
			before := latest.DeepCopy()
			expected := []netv1alpha1.LoadBalancerConfigCondition{{Type: netv1alpha1.LoadBalancerConfigReady, Status: corev1.ConditionTrue, Reason: "Reconciled", Message: "load balancer ready"}}
			latest.Status.Conditions = expected
			// LBC has no status subresource; the controller patches the main resource.
			if err := k8sClient.Patch(testCtx, latest, client.MergeFrom(before)); err != nil {
				t.Fatalf("status patch: %v", err)
			}
			updated := getLBC(t, obj)
			if !reflect.DeepEqual(updated.Status.Conditions, expected) {
				t.Fatalf("conditions: got %+v, want %+v", updated.Status.Conditions, expected)
			}
			if updated.Spec != expectedSpec {
				t.Fatalf("status patch changed spec: %+v", updated.Spec)
			}
		})
	}
}

func requireLBCRejection(t *testing.T, err error, message string) {
	t.Helper()
	if !isRejected(err) || !strings.Contains(err.Error(), message) {
		t.Fatalf("expected admission rejection containing %q, got: %v", message, err)
	}
}

func TestLoadBalancerConfig_NSXProviderRef_Admitted(t *testing.T) {
	// NSX may use a provider reference in a future configuration.
	obj := validAviLBC("lbc-nsx-with-ref")
	obj.Spec.Type = netv1alpha1.LoadBalancerConfigTypeNSX
	if err := k8sClient.Create(testCtx, obj); err != nil {
		t.Fatalf("expected NSX with a provider reference to be admitted, got: %v", err)
	}
	defer func() { _ = k8sClient.Delete(testCtx, obj) }()
}

func TestLoadBalancerConfig_SpecRequired_Rejected(t *testing.T) {
	t.Run("create-without-spec", func(t *testing.T) {
		obj := lbcWithoutReference("lbc-no-spec", "nsx")
		unstructured.RemoveNestedField(obj.Object, "spec")
		err := k8sClient.Create(testCtx, obj)
		if err == nil {
			t.Cleanup(func() { _ = k8sClient.Delete(testCtx, obj) })
		}
		requireLBCRejection(t, err, "spec: Required value")
	})
	t.Run("remove-spec", func(t *testing.T) {
		obj := validFoundationLBC("lbc-remove-spec")
		if err := k8sClient.Create(testCtx, obj); err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { _ = k8sClient.Delete(testCtx, obj) })
		if err := k8sClient.Get(testCtx, client.ObjectKeyFromObject(obj), obj); err != nil {
			t.Fatalf("get before update: %v", err)
		}
		err := k8sClient.Patch(testCtx, obj, client.RawPatch(types.MergePatchType, []byte(`{"spec":null}`)))
		requireLBCRejection(t, err, "spec: Required value")
	})
}

func TestLoadBalancerConfig_HAProxyKindChange_Admitted(t *testing.T) {
	for _, kind := range []string{"HAProxy", "HAProxyLoadBalancerConfig"} {
		t.Run(kind, func(t *testing.T) {
			obj := validHAProxyLBC("lbc-kind-" + strings.ToLower(kind))
			obj.Spec.ProviderRef.Kind = kind
			obj.Spec.ProviderRef.APIVersion = "v1alpha1"
			if err := k8sClient.Create(testCtx, obj); err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() { _ = k8sClient.Delete(testCtx, obj) })
			latest := getLBC(t, obj)
			if latest.Spec != obj.Spec {
				t.Fatalf("reference did not round-trip: %+v", latest.Spec)
			}
			obj = latest

			// Both spellings are supported for HAProxy references.
			obj.Spec.ProviderRef.Kind = "HAProxy"
			if kind == "HAProxy" {
				obj.Spec.ProviderRef.Kind = "HAProxyLoadBalancerConfig"
			}
			if err := k8sClient.Update(testCtx, obj); err != nil {
				t.Fatalf("expected HAProxy provider kind update to be admitted: %v", err)
			}
		})
	}
}

func assertUnstructuredLBCSpecRoundtrip(t *testing.T, obj *unstructured.Unstructured) {
	t.Helper()
	expected := obj.DeepCopy().Object["spec"]
	latest := &unstructured.Unstructured{}
	latest.SetGroupVersionKind(obj.GroupVersionKind())
	if err := k8sClient.Get(testCtx, client.ObjectKeyFromObject(obj), latest); err != nil {
		t.Fatalf("get after update: %v", err)
	}
	if !reflect.DeepEqual(latest.Object["spec"], expected) {
		t.Fatalf("spec: got %+v, want %+v", latest.Object["spec"], expected)
	}
	*obj = *latest
}

func TestLoadBalancerConfig_TypeRequired_Rejected(t *testing.T) {
	for _, tc := range []struct {
		name    string
		value   interface{}
		omit    bool
		message string
	}{
		{name: "omitted", omit: true, message: "spec.type: Required value"},
		{name: "null", value: nil, message: "spec.type: Required value"},
		{name: "empty", value: "", message: "spec.type: Unsupported value"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			obj := lbcWithoutReference("lbc-type-"+tc.name, "nsx")
			if tc.omit {
				unstructured.RemoveNestedField(obj.Object, "spec", "type")
			} else if err := unstructured.SetNestedField(obj.Object, tc.value, "spec", "type"); err != nil {
				t.Fatal(err)
			}
			requireLBCRejection(t, k8sClient.Create(testCtx, obj), tc.message)
		})
	}
}

func TestLoadBalancerConfig_UnchangedType_ReferenceChangeAdmitted(t *testing.T) {
	for _, obj := range controllerLBCs() {
		t.Run(obj.Name, func(t *testing.T) {
			if err := k8sClient.Create(testCtx, obj); err != nil {
				t.Fatal(err)
			}
			defer func() { _ = k8sClient.Delete(testCtx, obj) }()
			latest := getLBC(t, obj)
			latest.Spec.ProviderRef.Name = "different-provider"
			if err := k8sClient.Update(testCtx, latest); err != nil {
				t.Fatalf("expected provider reference update to be admitted: %v", err)
			}
			if got := getLBC(t, obj).Spec.ProviderRef.Name; got != "different-provider" {
				t.Fatalf("providerRef.name: got %q, want %q", got, "different-provider")
			}
		})
	}
}
