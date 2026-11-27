// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: Apache-2.0

package v1alpha1

import "testing"

func TestWorkloadNetworkConfigurationDeepCopyTier1NCPConfig(t *testing.T) {
	original := &WorkloadNetworkConfiguration{
		Spec: WorkloadNetworkConfigurationSpec{
			Providers: []NetworkProviderEntry{{NSXTier1NCPConfig: &NSXTier1NCPConfig{EdgeCluster: "edge-original"}}},
		},
	}
	copied := original.DeepCopy()
	copied.Spec.Providers[0].NSXTier1NCPConfig.EdgeCluster = "edge-copy"
	if got := original.Spec.Providers[0].NSXTier1NCPConfig.EdgeCluster; got != "edge-original" {
		t.Fatalf("modifying the copy changed the original edge cluster: %q", got)
	}
}
