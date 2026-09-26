// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: Apache-2.0

package v1alpha1

import (
	"context"
	"testing"

	netopv1alpha1 "github.com/vmware-tanzu/net-operator-api/api/v1alpha1"
	vpcv1alpha1 "github.com/vmware-tanzu/nsx-operator/pkg/apis/vpc/v1alpha1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrlfake "sigs.k8s.io/controller-runtime/pkg/client/fake"

	"github.com/stretchr/testify/require"
)

func testScheme() *runtime.Scheme {
	scheme := runtime.NewScheme()
	_ = clientgoscheme.AddToScheme(scheme)
	_ = netopv1alpha1.AddToScheme(scheme)
	_ = vpcv1alpha1.AddToScheme(scheme)
	return scheme
}

// ---------------------------------------------------------------------------
// NamespaceNetworkConfiguration <-> Network
// ---------------------------------------------------------------------------

func TestNetworkNamespaceNetworkConfigurationOwner(t *testing.T) {
	tests := []struct {
		name    string
		network *netopv1alpha1.Network
		want    string
		wantOK  bool
	}{
		{
			name:    "nil network",
			network: nil,
			want:    "",
			wantOK:  false,
		},
		{
			name:    "no labels",
			network: &netopv1alpha1.Network{},
			want:    "",
			wantOK:  false,
		},
		{
			name: "missing owner label",
			network: &netopv1alpha1.Network{
				ObjectMeta: metav1.ObjectMeta{Labels: map[string]string{"other-label": "val"}},
			},
			want:   "",
			wantOK: false,
		},
		{
			name: "owner label present",
			network: &netopv1alpha1.Network{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{netopv1alpha1.ManagedByNNCLabelKey: "my-nnc"},
				},
			},
			want:   "my-nnc",
			wantOK: true,
		},
		{
			name: "owner label present but empty",
			network: &netopv1alpha1.Network{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{netopv1alpha1.ManagedByNNCLabelKey: ""},
				},
			},
			want:   "",
			wantOK: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, ok := NetworkNamespaceNetworkConfigurationOwner(tc.network)
			require.Equal(t, tc.want, got)
			require.Equal(t, tc.wantOK, ok)
		})
	}
}

func TestNetworksOwnedByNamespaceNetworkConfiguration(t *testing.T) {
	netB := &netopv1alpha1.Network{ObjectMeta: metav1.ObjectMeta{
		Name:      "net-b",
		Namespace: "ns-1",
		Labels:    map[string]string{netopv1alpha1.ManagedByNNCLabelKey: "my-nnc"},
	}}

	netA := &netopv1alpha1.Network{ObjectMeta: metav1.ObjectMeta{
		Name:      "net-a",
		Namespace: "ns-1",
		Labels:    map[string]string{netopv1alpha1.ManagedByNNCLabelKey: "my-nnc"},
	}}

	netOtherNamespace := &netopv1alpha1.Network{ObjectMeta: metav1.ObjectMeta{
		Name:      "net-c",
		Namespace: "ns-2",
		Labels:    map[string]string{netopv1alpha1.ManagedByNNCLabelKey: "my-nnc"},
	}}

	netUnmanaged := &netopv1alpha1.Network{ObjectMeta: metav1.ObjectMeta{
		Name:      "net-unmanaged",
		Namespace: "ns-1",
		Labels:    map[string]string{netopv1alpha1.ManagedByNNCLabelKey: "other-nnc"},
	}}

	c := ctrlfake.NewClientBuilder().WithScheme(testScheme()).
		WithObjects(netB, netA, netOtherNamespace, netUnmanaged).Build()

	t.Run("filters by namespace", func(t *testing.T) {
		got, err := NetworksOwnedByNamespaceNetworkConfiguration(context.Background(), c, "my-nnc", "ns-1")
		require.NoError(t, err)
		require.Len(t, got, 2)

		requireContainsAll(t, got, uncastObjs(netA, netB)...)
	})

	t.Run("empty namespace searches across all namespaces", func(t *testing.T) {
		got, err := NetworksOwnedByNamespaceNetworkConfiguration(context.Background(), c, "my-nnc", "")
		require.NoError(t, err)
		require.Len(t, got, 3)

		requireContainsAll(t, got, uncastObjs(netA, netB, netOtherNamespace)...)
	})

	t.Run("no matches returns empty, not nil", func(t *testing.T) {
		got, err := NetworksOwnedByNamespaceNetworkConfiguration(context.Background(), c, "unknown-nnc", "")
		require.NoError(t, err)
		require.Empty(t, got)
	})
}

// ---------------------------------------------------------------------------
// NamespaceNetworkConfiguration <-> VPCNetworkConfiguration
// ---------------------------------------------------------------------------

func TestVPCNetworkConfigurationNamespaceNetworkConfigurationOwner(t *testing.T) {
	tests := []struct {
		name      string
		vpcNetCfg *vpcv1alpha1.VPCNetworkConfiguration
		want      string
		wantOK    bool
	}{
		{
			name:      "nil VPCNetworkConfiguration",
			vpcNetCfg: nil,
			want:      "",
			wantOK:    false,
		},
		{
			name:      "no labels",
			vpcNetCfg: &vpcv1alpha1.VPCNetworkConfiguration{},
			want:      "",
			wantOK:    false,
		},
		{
			name: "missing owner label",
			vpcNetCfg: &vpcv1alpha1.VPCNetworkConfiguration{
				ObjectMeta: metav1.ObjectMeta{Labels: map[string]string{"other-label": "val"}},
			},
			want:   "",
			wantOK: false,
		},
		{
			name: "owner label present",
			vpcNetCfg: &vpcv1alpha1.VPCNetworkConfiguration{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{netopv1alpha1.ManagedByNNCLabelKey: "my-nnc"},
				},
			},
			want:   "my-nnc",
			wantOK: true,
		},
		{
			name: "owner label present but empty",
			vpcNetCfg: &vpcv1alpha1.VPCNetworkConfiguration{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{netopv1alpha1.ManagedByNNCLabelKey: ""},
				},
			},
			want:   "",
			wantOK: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, ok := VPCNetworkConfigurationNamespaceNetworkConfigurationOwner(tc.vpcNetCfg)
			require.Equal(t, tc.want, got)
			require.Equal(t, tc.wantOK, ok)
		})
	}
}

func TestVPCNetworkConfigurationOwnedByNamespaceNetworkConfiguration(t *testing.T) {
	vpcCfgA := &vpcv1alpha1.VPCNetworkConfiguration{ObjectMeta: metav1.ObjectMeta{
		Name:   "vpc-cfg-a",
		Labels: map[string]string{netopv1alpha1.ManagedByNNCLabelKey: "my-nnc"},
	}}

	vpcCfgUnmanaged := &vpcv1alpha1.VPCNetworkConfiguration{ObjectMeta: metav1.ObjectMeta{
		Name:   "vpc-cfg-unmanaged",
		Labels: map[string]string{netopv1alpha1.ManagedByNNCLabelKey: "other-nnc"},
	}}

	t.Run("none found returns nil, nil", func(t *testing.T) {
		c := ctrlfake.NewClientBuilder().WithScheme(testScheme()).
			WithObjects(vpcCfgUnmanaged).Build()

		got, err := VPCNetworkConfigurationOwnedByNamespaceNetworkConfiguration(context.Background(), c, "my-nnc")
		require.NoError(t, err)
		require.Nil(t, got)
	})

	t.Run("exactly one found is returned", func(t *testing.T) {
		c := ctrlfake.NewClientBuilder().WithScheme(testScheme()).
			WithObjects(vpcCfgA, vpcCfgUnmanaged).Build()

		got, err := VPCNetworkConfigurationOwnedByNamespaceNetworkConfiguration(context.Background(), c, "my-nnc")
		require.NoError(t, err)
		require.NotNil(t, got)
		require.Equal(t, vpcCfgA.Name, got.Name)
	})

	t.Run("more than one found is an error naming the offenders", func(t *testing.T) {
		vpcCfgB := &vpcv1alpha1.VPCNetworkConfiguration{ObjectMeta: metav1.ObjectMeta{
			Name:   "vpc-cfg-b",
			Labels: map[string]string{netopv1alpha1.ManagedByNNCLabelKey: "my-nnc"},
		}}

		c := ctrlfake.NewClientBuilder().WithScheme(testScheme()).
			WithObjects(vpcCfgA, vpcCfgB).Build()

		got, err := VPCNetworkConfigurationOwnedByNamespaceNetworkConfiguration(context.Background(), c, "my-nnc")
		require.Nil(t, got)
		require.Error(t, err)
		require.ErrorContains(t, err, "2")
		require.ErrorContains(t, err, vpcCfgA.Name)
		require.ErrorContains(t, err, vpcCfgB.Name)

		var multiErr *MultipleVPCNetworkConfigurationsError
		require.ErrorAs(t, err, &multiErr)
		require.Equal(t, "my-nnc", multiErr.NNCName)
		requireContainsAll(t, multiErr.VPCNetworkConfigurationNames, vpcCfgA.Name, vpcCfgB.Name)
	})
}

// requireContainsAll asserts that the given list of entries is a subset of the provided container.
func requireContainsAll[V any](t *testing.T, container []V, entries ...V) {
	for _, entry := range entries {
		require.Contains(t, container, entry)
	}
}

// uncastObjs takes a list of pointer objects and returns them by value.
func uncastObjs[T any](objs ...*T) []T {
	uncasted := make([]T, 0, len(objs))
	for _, obj := range objs {
		uncasted = append(uncasted, *obj)
	}

	return uncasted
}
