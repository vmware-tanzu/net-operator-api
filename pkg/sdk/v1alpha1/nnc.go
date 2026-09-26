// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: Apache-2.0

package v1alpha1

import (
	"context"
	"fmt"

	netopv1alpha1 "github.com/vmware-tanzu/net-operator-api/api/v1alpha1"
	vpcv1alpha1 "github.com/vmware-tanzu/nsx-operator/pkg/apis/vpc/v1alpha1"

	ctrlclient "sigs.k8s.io/controller-runtime/pkg/client"
)

// labeledObject constrains T to a pointer to U that also implements ctrlclient.Object,
// allowing namespaceNetworkConfigurationOwner below to accept any such CRD type while
// still supporting a nil check on T.
type labeledObject[U any] interface {
	*U
	ctrlclient.Object
}

// namespaceNetworkConfigurationOwner returns the value of the ManagedByNNCLabelKey label on
// obj, and whether the label is present.
func namespaceNetworkConfigurationOwner[T labeledObject[U], U any](obj T) (string, bool) {
	if obj == nil {
		return "", false
	}

	owner, ok := obj.GetLabels()[netopv1alpha1.ManagedByNNCLabelKey]
	return owner, ok
}

// ---------------------------------------------------------------------------
// NamespaceNetworkConfiguration <-> Network
// ---------------------------------------------------------------------------

// NetworkNamespaceNetworkConfigurationOwner returns the corresponding NamespaceNetworkConfiguration
// that manages this given Network, and a boolean indicating if the label is present.
// If the boolean is false, this Network is not managed via a NamespaceNetworkConfiguration.
func NetworkNamespaceNetworkConfigurationOwner(network *netopv1alpha1.Network) (string, bool) {
	return namespaceNetworkConfigurationOwner(network)
}

// NetworksOwnedByNamespaceNetworkConfiguration retrieves a list of Network resources
// that are managed by the specified NamespaceNetworkConfiguration.
//
// As NamespaceNetworkConfiguration's may be scoped to multiple Kubernetes namespaces, an optional
// namespace parameter is provided to help filter the query further. If an empty string ("") is
// provided, it will search for matching Networks across all namespaces at the cluster scope.
// Otherwise, it will only return Networks within the specified namespace.
//
// Example usage:
//
//	// Search across all namespaces
//	nets, err := NetworksOwnedByNamespaceNetworkConfiguration(ctx, c, "my-nnc", "")
//
//	// Search within a specific namespace
//	nets, err := NetworksOwnedByNamespaceNetworkConfiguration(ctx, c, "my-nnc", "default")
func NetworksOwnedByNamespaceNetworkConfiguration(ctx context.Context, c ctrlclient.Client, nncName, namespace string) ([]netopv1alpha1.Network, error) {
	listOpts := []ctrlclient.ListOption{
		ctrlclient.MatchingLabels{netopv1alpha1.ManagedByNNCLabelKey: nncName},
	}

	// Conditionally append the namespace filter.
	if namespace != "" {
		listOpts = append(listOpts, ctrlclient.InNamespace(namespace))
	}

	var networkList netopv1alpha1.NetworkList
	if err := c.List(ctx, &networkList, listOpts...); err != nil {
		return nil, fmt.Errorf("error listing Networks owned by NamespaceNetworkConfiguration '%s' in namespace '%s': %w", nncName, namespace, err)
	}

	return networkList.Items, nil
}

// ---------------------------------------------------------------------------
// NamespaceNetworkConfiguration <-> VPCNetworkConfiguration
// ---------------------------------------------------------------------------

// VPCNetworkConfigurationNamespaceNetworkConfigurationOwner returns the corresponding
// NamespaceNetworkConfiguration that manages this given VPCNetworkConfiguration, and a boolean
// indicating if the label is present. If the boolean is false, this VPCNetworkConfiguration is
// not managed via a NamespaceNetworkConfiguration.
func VPCNetworkConfigurationNamespaceNetworkConfigurationOwner(vpcNetCfg *vpcv1alpha1.VPCNetworkConfiguration) (string, bool) {
	return namespaceNetworkConfigurationOwner(vpcNetCfg)
}

// VPCNetworkConfigurationOwnedByNamespaceNetworkConfiguration returns the VPCNetworkConfiguration
// managed by the specified NamespaceNetworkConfiguration. VPCNetworkConfiguration is cluster-scoped,
// so at most one is expected per NamespaceNetworkConfiguration.
//
// If no VPCNetworkConfiguration is found, it returns (nil, nil). This is not an error: it means
// the NamespaceNetworkConfiguration has not yet been reconciled to a VPCNetworkConfiguration.
//
// If more than one VPCNetworkConfiguration is found, a *MultipleVPCNetworkConfigurationsError is
// returned naming the offending VPCNetworkConfigurations, since this indicates a data-integrity
// problem.
func VPCNetworkConfigurationOwnedByNamespaceNetworkConfiguration(ctx context.Context, c ctrlclient.Client, nncName string) (*vpcv1alpha1.VPCNetworkConfiguration, error) {
	var vpcNetCfgList vpcv1alpha1.VPCNetworkConfigurationList
	if err := c.List(ctx, &vpcNetCfgList, ctrlclient.MatchingLabels{netopv1alpha1.ManagedByNNCLabelKey: nncName}); err != nil {
		return nil, fmt.Errorf("error listing VPCNetworkConfigurations owned by NamespaceNetworkConfiguration '%s': %w", nncName, err)
	}

	switch len(vpcNetCfgList.Items) {
	case 0:
		return nil, nil
	case 1:
		return &vpcNetCfgList.Items[0], nil
	default:
		names := make([]string, 0, len(vpcNetCfgList.Items))
		for _, item := range vpcNetCfgList.Items {
			names = append(names, item.Name)
		}
		return nil, &MultipleVPCNetworkConfigurationsError{NNCName: nncName, VPCNetworkConfigurationNames: names}
	}
}
