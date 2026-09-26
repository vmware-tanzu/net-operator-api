// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: Apache-2.0

package v1alpha1

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMultipleVPCNetworkConfigurationsError_Error(t *testing.T) {
	err := &MultipleVPCNetworkConfigurationsError{
		NNCName:                      "my-nnc",
		VPCNetworkConfigurationNames: []string{"vpc-cfg-a", "vpc-cfg-b"},
	}

	require.ErrorContains(t, err, "2")
	require.ErrorContains(t, err, "my-nnc")
	require.ErrorContains(t, err, "vpc-cfg-a")
	require.ErrorContains(t, err, "vpc-cfg-b")
}

func TestMultipleVPCNetworkConfigurationsError_ErrorsAs(t *testing.T) {
	var err error = &MultipleVPCNetworkConfigurationsError{
		NNCName:                      "my-nnc",
		VPCNetworkConfigurationNames: []string{"vpc-cfg-a", "vpc-cfg-b"},
	}

	var target *MultipleVPCNetworkConfigurationsError
	require.True(t, errors.As(err, &target))
	require.Equal(t, "my-nnc", target.NNCName)
	require.Equal(t, []string{"vpc-cfg-a", "vpc-cfg-b"}, target.VPCNetworkConfigurationNames)
}
