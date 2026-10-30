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
		{"nsx-registered-avi", validNSXRegisteredAviLBC("lbc-valid-nsx-avi")},
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
	for _, provider := range []string{"foundation", "avi", "haproxy", "nsx-registered-avi"} {
		t.Run(provider, func(t *testing.T) {
			obj := lbcWithoutReference("lbc-no-ref-"+provider, provider)
			requireLBCRejection(t, k8sClient.Create(testCtx, obj), "spec.providerRef is required except for nsx")
		})
	}
}

func TestLoadBalancerConfig_ProviderRefKindMismatch_Rejected(t *testing.T) {
	for _, newObject := range []func(string) *netv1alpha1.LoadBalancerConfig{
		validFoundationLBC, validAviLBC, validHAProxyLBC, validNSXRegisteredAviLBC,
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

func TestLoadBalancerConfig_NSXProviderSwitch_Admitted(t *testing.T) {
	obj := lbcWithoutReference("lbc-nsx-switch", "nsx")
	if err := k8sClient.Create(testCtx, obj); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = k8sClient.Delete(testCtx, obj) }()
	assertUnstructuredLBCSpecRoundtrip(t, obj)
	ref := map[string]interface{}{"apiGroup": "netoperator.vmware.com", "apiVersion": "v1alpha1", "kind": "AviLoadBalancerConfig", "name": "my-avi"}
	if err := unstructured.SetNestedMap(obj.Object, ref, "spec", "providerRef"); err != nil {
		t.Fatal(err)
	}
	for _, provider := range []string{"nsx-registered-avi", "nsx", "nsx-registered-avi"} {
		if err := unstructured.SetNestedField(obj.Object, provider, "spec", "type"); err != nil {
			t.Fatal(err)
		}
		if err := k8sClient.Update(testCtx, obj); err != nil {
			t.Fatalf("switch to %s: %v", provider, err)
		}
		assertUnstructuredLBCSpecRoundtrip(t, obj)
		// Reference removal is forbidden in both NSX modes, including on the return transition.
		patch := []byte(`{"spec":{"type":"nsx","providerRef":null}}`)
		requireLBCRejection(t, k8sClient.Patch(testCtx, obj.DeepCopy(), client.RawPatch(types.MergePatchType, patch)), "spec.providerRef is immutable once set")
	}
}

func TestLoadBalancerConfig_ProviderRefImmutable_Rejected(t *testing.T) {
	for _, tc := range []struct {
		name   string
		mutate func(*netv1alpha1.LoadBalancerConfigProviderReference)
	}{
		{"name", func(ref *netv1alpha1.LoadBalancerConfigProviderReference) { ref.Name = "different-name" }},
		{"api-group", func(ref *netv1alpha1.LoadBalancerConfigProviderReference) { ref.APIGroup = "other.example.com" }},
		{"api-version", func(ref *netv1alpha1.LoadBalancerConfigProviderReference) { ref.APIVersion = "v1beta1" }},
		{"api-version-removal", func(ref *netv1alpha1.LoadBalancerConfigProviderReference) { ref.APIVersion = "" }},
		{"kind", func(ref *netv1alpha1.LoadBalancerConfigProviderReference) { ref.Kind = "AviLoadBalancerConfig" }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			obj := validFoundationLBC("lbc-ref-immutable-" + tc.name)
			obj.Spec.ProviderRef.APIVersion = "v1alpha1"
			if err := k8sClient.Create(testCtx, obj); err != nil {
				t.Fatal(err)
			}
			defer func() { _ = k8sClient.Delete(testCtx, obj) }()
			if err := k8sClient.Get(testCtx, client.ObjectKeyFromObject(obj), obj); err != nil {
				t.Fatalf("get before update: %v", err)
			}
			tc.mutate(&obj.Spec.ProviderRef)
			requireLBCRejection(t, k8sClient.Update(testCtx, obj), "spec.providerRef is immutable once set")
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
	// Retagging alone is invalid: an ALBC reference must be supplied atomically.
	if err := unstructured.SetNestedField(obj.Object, "nsx-registered-avi", "spec", "type"); err != nil {
		t.Fatal(err)
	}
	requireLBCRejection(t, k8sClient.Update(testCtx, obj), "spec.providerRef is required except for nsx")
}

func TestLoadBalancerConfig_ProviderRefRemoval_Rejected(t *testing.T) {
	// NSX permits omission on creation, isolating the transition rule from the
	// rule requiring a provider reference for non-NSX types.
	obj := validNSXLBC("lbc-ref-removal")
	obj.Spec.ProviderRef = validFoundationLBC("").Spec.ProviderRef
	if err := k8sClient.Create(testCtx, obj); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = k8sClient.Delete(testCtx, obj) }()
	if err := k8sClient.Get(testCtx, client.ObjectKeyFromObject(obj), obj); err != nil {
		t.Fatalf("get before update: %v", err)
	}
	err := k8sClient.Patch(testCtx, obj, client.RawPatch(types.MergePatchType, []byte(`{"spec":{"providerRef":null}}`)))
	if !isRejected(err) || !strings.Contains(err.Error(), "spec.providerRef is immutable once set") {
		t.Fatalf("expected reference immutability rejection, got: %v", err)
	}
}

func TestLoadBalancerConfig_TypeTransitions_Rejected(t *testing.T) {
	providers := []netv1alpha1.LoadBalancerConfigType{
		netv1alpha1.LoadBalancerConfigTypeFoundation, netv1alpha1.LoadBalancerConfigTypeAvi,
		netv1alpha1.LoadBalancerConfigTypeHAProxy, netv1alpha1.LoadBalancerConfigTypeNSX,
		netv1alpha1.LoadBalancerConfigTypeNSXRegisteredAvi,
	}
	for _, from := range providers {
		for _, to := range providers {
			if from == to || (isNSXLBCType(from) && isNSXLBCType(to)) || (from == netv1alpha1.LoadBalancerConfigTypeAvi && to == netv1alpha1.LoadBalancerConfigTypeNSXRegisteredAvi) {
				continue
			}
			t.Run(string(from)+"-to-"+string(to), func(t *testing.T) {
				obj := validFoundationLBC("lbc-transition-" + string(from) + "-" + string(to))
				obj.Spec.Type = from
				kinds := map[netv1alpha1.LoadBalancerConfigType]string{
					netv1alpha1.LoadBalancerConfigTypeNSXRegisteredAvi: "AviLoadBalancerConfig",
					netv1alpha1.LoadBalancerConfigTypeFoundation:       "FoundationLoadBalancerConfig",
					netv1alpha1.LoadBalancerConfigTypeAvi:              "AviLoadBalancerConfig",
					netv1alpha1.LoadBalancerConfigTypeHAProxy:          "HAProxyLoadBalancerConfig",
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
	return []*netv1alpha1.LoadBalancerConfig{validFoundationLBC("lbc-foundation"), validAviLBC("lbc-avi"), validHAProxyLBC("lbc-haproxy"), legacy, validNSXLBC("lbc-nsx"), validNSXRegisteredAviLBC("lbc-nsx-avi")}
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

func isNSXLBCType(provider netv1alpha1.LoadBalancerConfigType) bool {
	return provider == netv1alpha1.LoadBalancerConfigTypeNSX || provider == netv1alpha1.LoadBalancerConfigTypeNSXRegisteredAvi
}

func requireLBCRejection(t *testing.T, err error, message string) {
	t.Helper()
	if !isRejected(err) || !strings.Contains(err.Error(), message) {
		t.Fatalf("expected admission rejection containing %q, got: %v", message, err)
	}
}

func TestLoadBalancerConfig_ProviderRefInitiallyAbsent_Admitted(t *testing.T) {
	// "Immutable once set" allows setting a previously absent reference, but
	// subsequent changes must be rejected. Use NSX to permit initial absence.
	obj := lbcWithoutReference("lbc-set-ref-nsx", "nsx")
	if err := k8sClient.Create(testCtx, obj); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = k8sClient.Delete(testCtx, obj) }()
	if err := k8sClient.Get(testCtx, client.ObjectKeyFromObject(obj), obj); err != nil {
		t.Fatalf("get before update: %v", err)
	}
	if err := k8sClient.Update(testCtx, obj); err != nil {
		t.Fatalf("unchanged absent reference: %v", err)
	}
	assertUnstructuredLBCSpecRoundtrip(t, obj)
	ref := map[string]interface{}{"apiGroup": "netoperator.vmware.com", "kind": "AviLoadBalancerConfig", "name": "my-avi"}
	if err := unstructured.SetNestedMap(obj.Object, ref, "spec", "providerRef"); err != nil {
		t.Fatal(err)
	}
	if err := k8sClient.Update(testCtx, obj); err != nil {
		t.Fatalf("initial reference assignment: %v", err)
	}
	assertUnstructuredLBCSpecRoundtrip(t, obj)
	if err := k8sClient.Update(testCtx, obj); err != nil {
		t.Fatalf("unchanged present reference: %v", err)
	}
	assertUnstructuredLBCSpecRoundtrip(t, obj)
	if err := unstructured.SetNestedField(obj.Object, "other-avi", "spec", "providerRef", "name"); err != nil {
		t.Fatal(err)
	}
	requireLBCRejection(t, k8sClient.Update(testCtx, obj), "spec.providerRef is immutable once set")
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

func TestLoadBalancerConfig_HAProxyKindChange_Rejected(t *testing.T) {
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

			// Both spellings are accepted, but an existing reference remains immutable.
			obj.Spec.ProviderRef.Kind = "HAProxy"
			if kind == "HAProxy" {
				obj.Spec.ProviderRef.Kind = "HAProxyLoadBalancerConfig"
			}
			requireLBCRejection(t, k8sClient.Update(testCtx, obj), "spec.providerRef is immutable once set")
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

func validNSXRegisteredAviLBC(name string) *netv1alpha1.LoadBalancerConfig {
	obj := validAviLBC(name)
	obj.Spec.Type = netv1alpha1.LoadBalancerConfigTypeNSXRegisteredAvi
	return obj
}

func TestLoadBalancerConfig_AviRetag_Admitted(t *testing.T) {
	obj := validAviLBC("lbc-avi-retag")
	obj.Spec.ProviderRef.APIVersion = "v1alpha1"
	if err := k8sClient.Create(testCtx, obj); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = k8sClient.Delete(testCtx, obj) }()
	latest := getLBC(t, obj)
	expected := latest.Spec
	expected.Type = netv1alpha1.LoadBalancerConfigTypeNSXRegisteredAvi
	latest.Spec = expected
	if err := k8sClient.Update(testCtx, latest); err != nil {
		t.Fatalf("retag: %v", err)
	}
	if got := getLBC(t, obj).Spec; got != expected {
		t.Fatalf("spec: got %+v, want %+v", got, expected)
	}
}

func TestLoadBalancerConfig_AviRetagReferenceChange_Rejected(t *testing.T) {
	for _, patch := range []struct{ name, body string }{
		{"replace", `{"spec":{"type":"nsx-registered-avi","providerRef":{"name":"different-avi"}}}`},
		{"remove", `{"spec":{"type":"nsx-registered-avi","providerRef":null}}`},
	} {
		t.Run(patch.name, func(t *testing.T) {
			obj := validAviLBC("lbc-retag-ref-" + patch.name)
			if err := k8sClient.Create(testCtx, obj); err != nil {
				t.Fatal(err)
			}
			defer func() { _ = k8sClient.Delete(testCtx, obj) }()
			latest := getLBC(t, obj)
			expected := latest.Spec
			requireLBCRejection(t, k8sClient.Patch(testCtx, latest, client.RawPatch(types.MergePatchType, []byte(patch.body))), "spec.providerRef is immutable once set")
			if got := getLBC(t, obj).Spec; got != expected {
				t.Fatalf("rejected patch changed spec: %+v", got)
			}
		})
	}
}

func TestLoadBalancerConfig_NSXToRegisteredAvi_ReferenceReplacement(t *testing.T) {
	for _, source := range []struct {
		name string
		ref  netv1alpha1.LoadBalancerConfigProviderReference
	}{
		{"zero", netv1alpha1.LoadBalancerConfigProviderReference{}},
		{"foundation", validFoundationLBC("").Spec.ProviderRef},
		{"avi", validAviLBC("").Spec.ProviderRef},
	} {
		t.Run(source.name, func(t *testing.T) {
			obj := validNSXLBC("lbc-nsx-ref-" + source.name)
			obj.Spec.ProviderRef = source.ref
			if err := k8sClient.Create(testCtx, obj); err != nil {
				t.Fatal(err)
			}
			defer func() { _ = k8sClient.Delete(testCtx, obj) }()
			latest := getLBC(t, obj)
			// Verify the wire representation as well as the typed value: a typed GET
			// alone cannot distinguish an omitted reference from a zero-valued one.
			wire := lbcWithoutReference(obj.Name, "nsx")
			if err := k8sClient.Get(testCtx, client.ObjectKeyFromObject(obj), wire); err != nil {
				t.Fatal(err)
			}
			if _, found, err := unstructured.NestedMap(wire.Object, "spec", "providerRef"); err != nil || !found {
				t.Fatalf("expected stored providerRef, found=%v, err=%v", found, err)
			}
			if latest.Spec.ProviderRef != source.ref {
				t.Fatalf("unexpected source reference: %+v", latest.Spec.ProviderRef)
			}
			expected := netv1alpha1.LoadBalancerConfigSpec{
				Type: netv1alpha1.LoadBalancerConfigTypeNSXRegisteredAvi,
				ProviderRef: netv1alpha1.LoadBalancerConfigProviderReference{
					APIGroup: "netoperator.vmware.com", APIVersion: "v1alpha1", Kind: "AviLoadBalancerConfig", Name: "replacement-avi",
				},
			}
			latest.Spec = expected
			if err := k8sClient.Update(testCtx, latest); err != nil {
				t.Fatalf("replace reference during NSX transition: %v", err)
			}
			updated := getLBC(t, obj)
			if updated.Spec != expected {
				t.Fatalf("spec: got %+v, want %+v", updated.Spec, expected)
			}
			updated.Spec.ProviderRef.Name = "another-avi"
			requireLBCRejection(t, k8sClient.Update(testCtx, updated), "spec.providerRef is immutable once set")
			updated = getLBC(t, obj)
			updated.Spec.Type = netv1alpha1.LoadBalancerConfigTypeNSX
			updated.Spec.ProviderRef.Name = "another-avi"
			requireLBCRejection(t, k8sClient.Update(testCtx, updated), "spec.providerRef is immutable once set")
		})
	}
}

func TestLoadBalancerConfig_NSXToRegisteredAvi_InvalidReference_Rejected(t *testing.T) {
	for _, tc := range []struct{ name, patch, message string }{
		{"missing", `{"spec":{"type":"nsx-registered-avi","providerRef":null}}`, "spec.providerRef is required except for nsx"},
		{"wrong-kind", `{"spec":{"type":"nsx-registered-avi","providerRef":{"kind":"FoundationLoadBalancerConfig"}}}`, "providerRef.kind must match spec.type"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			obj := validNSXLBC("lbc-nsx-invalid-ref-" + tc.name)
			if err := k8sClient.Create(testCtx, obj); err != nil {
				t.Fatal(err)
			}
			defer func() { _ = k8sClient.Delete(testCtx, obj) }()
			latest := getLBC(t, obj)
			expected := latest.Spec
			requireLBCRejection(t, k8sClient.Patch(testCtx, latest, client.RawPatch(types.MergePatchType, []byte(tc.patch))), tc.message)
			if got := getLBC(t, obj).Spec; got != expected {
				t.Fatalf("rejected patch changed spec: %+v", got)
			}
		})
	}
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

func TestLoadBalancerConfig_UnchangedType_ReferenceChangeRejected(t *testing.T) {
	// The NSX-to-registered-Avi exception must not permit reference replacement
	// when either endpoint type is unchanged, or weaken the other providers.
	for _, obj := range controllerLBCs() {
		t.Run(obj.Name, func(t *testing.T) {
			if err := k8sClient.Create(testCtx, obj); err != nil {
				t.Fatal(err)
			}
			defer func() { _ = k8sClient.Delete(testCtx, obj) }()
			latest := getLBC(t, obj)
			expected := latest.Spec
			latest.Spec.ProviderRef.Name = "different-provider"
			requireLBCRejection(t, k8sClient.Update(testCtx, latest), "spec.providerRef is immutable once set")
			if got := getLBC(t, obj).Spec; got != expected {
				t.Fatalf("rejected update changed spec: %+v", got)
			}
		})
	}
}
