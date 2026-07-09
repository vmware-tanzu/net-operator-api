// Copyright (c) 2026 Broadcom. All Rights Reserved.
// Broadcom Confidential. The term "Broadcom" refers to Broadcom Inc.
// and/or its subsidiaries.

package cel_test

// Network provider enum values used in unstructured map literals.
const (
	providerVPC = "vpc"
)

// WNC singleton resource identity.
const (
	wncAPIVersion    = "netoperator.vmware.com/v1alpha1"
	wncKind          = "WorkloadNetworkConfiguration"
	wncDefaultName   = "default"
	wncSystemNetName = "system-net"
)

// NNC resource identity.
const (
	nncAPIVersion = "netoperator.vmware.com/v1alpha1"
	nncKind       = "NamespaceNetworkConfiguration"
)

// NetworkSettings resource identity.
const (
	nsAPIVersion = "netoperator.vmware.com/v1alpha1"
	nsKind       = "NetworkSettings"
	nsNamespace  = "default"
)

// Shared test path and name values.
const (
	testNetName        = "net"
	testVPCPath        = "/infra/vpcs/v1"
	testVPCPathFull    = "/infra/orgs/default/projects/p1/vpcs/v1"
	testNSXProject     = "/infra/orgs/default/projects/p1"
	testVPCConnProfile = "/infra/vpc-conn-profiles/default"
	testSubnetPath1    = "/infra/subnets/s1"
	testSubnetPath2    = "/infra/subnets/s2"
	testSubnetNameA    = "subnet-a"
	testSubnetNameB    = "subnet-b"
	testUppercaseName  = "UPPERCASE"
	testLabelTrue      = "true"
	testCIDR1          = "10.0.0.0/24"
	testWNCVPCPath     = "/test/vpc/path"
	testNamespaceA     = "ns-a"
	testNamespaceB     = "ns-b"
	testFinalizer      = "test-fin-remove"
)

// Shared condition test data used across NNC and WNC validation tests.
const (
	testConditionReason = "TestReason"
	testConditionReady  = "Ready"
)
