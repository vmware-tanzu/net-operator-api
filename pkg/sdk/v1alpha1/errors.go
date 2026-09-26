// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: Apache-2.0

package v1alpha1

import "fmt"

// MultipleVPCNetworkConfigurationsError indicates that more than one VPCNetworkConfiguration
// was found labeled as owned by a single NamespaceNetworkConfiguration, even though at most
// one is expected. This signals a data-integrity problem.
type MultipleVPCNetworkConfigurationsError struct {
	// NNCName is the name of the NamespaceNetworkConfiguration that unexpectedly owns
	// more than one VPCNetworkConfiguration.
	NNCName string

	// VPCNetworkConfigurationNames are the names of the offending VPCNetworkConfigurations.
	VPCNetworkConfigurationNames []string
}

func (e *MultipleVPCNetworkConfigurationsError) Error() string {
	return fmt.Sprintf("found %d VPCNetworkConfigurations owned by NamespaceNetworkConfiguration '%s', expected at most 1: %v",
		len(e.VPCNetworkConfigurationNames), e.NNCName, e.VPCNetworkConfigurationNames)
}
