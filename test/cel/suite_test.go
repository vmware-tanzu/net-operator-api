// Copyright (c) 2026 Broadcom. All Rights Reserved.
// Broadcom Confidential. The term "Broadcom" refers to Broadcom Inc.
// and/or its subsidiaries.

// Package cel_test contains envtest-based tests that verify CEL and OpenAPI
// validation rules fire correctly on admission for the CRD types in this repo.
package cel_test

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	netv1alpha1 "github.com/vmware-tanzu/net-operator-api/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	apiruntime "k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
)

var (
	k8sClient  client.Client
	testEnv    *envtest.Environment
	testCtx    context.Context
	testCancel context.CancelFunc
)

func TestMain(m *testing.M) {
	testCtx, testCancel = context.WithCancel(context.Background())
	defer testCancel()

	// Locate CRD base directory relative to this file.
	_, filename, _, _ := runtime.Caller(0)
	repoRoot := filepath.Join(filepath.Dir(filename), "..", "..")
	crdDir := filepath.Join(repoRoot, "config", "crd", "bases")

	testEnv = &envtest.Environment{
		CRDDirectoryPaths:     []string{crdDir},
		ErrorIfCRDPathMissing: true,
	}

	cfg, err := testEnv.Start()
	if err != nil {
		panic(err)
	}

	testScheme := apiruntime.NewScheme()
	if err := corev1.AddToScheme(testScheme); err != nil {
		panic(err)
	}
	if err := netv1alpha1.AddToScheme(testScheme); err != nil {
		panic(err)
	}

	k8sClient, err = client.New(cfg, client.Options{Scheme: testScheme})
	if err != nil {
		panic(err)
	}

	code := m.Run()

	if err := testEnv.Stop(); err != nil {
		panic(err)
	}
	os.Exit(code)
}

// isRejected returns true when the server rejects an admission request.
// CEL and OpenAPI validation failures surface as 422 Unprocessable Entity or
// 400 Bad Request from the kube-apiserver.
func isRejected(err error) bool {
	return apierrors.IsInvalid(err) || apierrors.IsBadRequest(err)
}

// ensureNamespace idempotently creates a namespace for namespace-scoped tests.
func ensureNamespace(t *testing.T, name string) {
	t.Helper()
	ns := &corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: name}}
	err := k8sClient.Create(testCtx, ns)
	if err != nil && !apierrors.IsAlreadyExists(err) {
		t.Fatalf("create namespace %s: %v", name, err)
	}
}
