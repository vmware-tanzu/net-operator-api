// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: Apache-2.0

package v1alpha1

import (
	"reflect"
	"testing"
)

func TestEffectivePortGroupIDs_NilFallsBackToPortGroupID(t *testing.T) {
	s := &VSphereDistributedNetworkSpec{PortGroupID: "dvportgroup-1"}
	if got, want := s.EffectivePortGroupIDs(), []string{"dvportgroup-1"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v, want %v", got, want)
	}
}

func TestEffectivePortGroupIDs_EmptyFallsBackToPortGroupID(t *testing.T) {
	s := &VSphereDistributedNetworkSpec{PortGroupID: "dvportgroup-1", PortGroupIDs: []string{}}
	if got, want := s.EffectivePortGroupIDs(), []string{"dvportgroup-1"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v, want %v", got, want)
	}
}

func TestEffectivePortGroupIDs_ReturnsPortGroupIDsWhenSet(t *testing.T) {
	s := &VSphereDistributedNetworkSpec{
		PortGroupID:  "dvportgroup-1",
		PortGroupIDs: []string{"dvportgroup-1", "dvportgroup-2"},
	}
	if got, want := s.EffectivePortGroupIDs(), []string{"dvportgroup-1", "dvportgroup-2"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v, want %v", got, want)
	}
}

func TestEffectivePortGroupIDs_ReturnedSliceIsACopy(t *testing.T) {
	s := &VSphereDistributedNetworkSpec{
		PortGroupID:  "dvportgroup-1",
		PortGroupIDs: []string{"dvportgroup-1", "dvportgroup-2"},
	}
	got := s.EffectivePortGroupIDs()
	got[0] = "changed"
	if s.PortGroupIDs[0] != "dvportgroup-1" {
		t.Fatalf("spec changed through returned slice: %v", s.PortGroupIDs)
	}

	s = &VSphereDistributedNetworkSpec{PortGroupID: "dvportgroup-1"}
	got = s.EffectivePortGroupIDs()
	got[0] = "changed"
	if s.PortGroupID != "dvportgroup-1" {
		t.Fatalf("spec changed through returned slice: %v", s.PortGroupID)
	}
}
