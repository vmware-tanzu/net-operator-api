// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: Apache-2.0

package v1alpha1

import (
	"encoding/json"
	"fmt"

	"sigs.k8s.io/controller-runtime/pkg/conversion"

	"github.com/vmware-tanzu/net-operator-api/api/v1alpha2"
)

var _ conversion.Convertible = &VSphereDistributedNetwork{}

// ConvertTo converts this VSphereDistributedNetwork to the v1alpha2 hub. The two versions have the
// same spec except that v1alpha1 spec.ipPools is v1alpha2 status.ipPools, so every other field is
// copied as it is.
func (src *VSphereDistributedNetwork) ConvertTo(dstRaw conversion.Hub) error {
	dst, ok := dstRaw.(*v1alpha2.VSphereDistributedNetwork)
	if !ok {
		return fmt.Errorf("unexpected hub type %T", dstRaw)
	}

	dst.ObjectMeta = *src.ObjectMeta.DeepCopy()
	if err := copyFields(&src.Spec, &dst.Spec); err != nil {
		return fmt.Errorf("converting spec: %w", err)
	}
	if err := copyFields(&src.Status, &dst.Status); err != nil {
		return fmt.Errorf("converting status: %w", err)
	}
	dst.Status.IPPools = nil
	if len(src.Spec.IPPools) > 0 {
		dst.Status.IPPools = make([]v1alpha2.IPPoolReference, len(src.Spec.IPPools))
		for i, ref := range src.Spec.IPPools {
			dst.Status.IPPools[i] = v1alpha2.IPPoolReference{Name: ref.Name, APIVersion: ref.APIVersion}
		}
	}
	return nil
}

// ConvertFrom converts the v1alpha2 hub to this VSphereDistributedNetwork. v1alpha2 status.ipPools
// becomes spec.ipPools.
func (dst *VSphereDistributedNetwork) ConvertFrom(srcRaw conversion.Hub) error {
	src, ok := srcRaw.(*v1alpha2.VSphereDistributedNetwork)
	if !ok {
		return fmt.Errorf("unexpected hub type %T", srcRaw)
	}

	dst.ObjectMeta = *src.ObjectMeta.DeepCopy()
	if err := copyFields(&src.Spec, &dst.Spec); err != nil {
		return fmt.Errorf("converting spec: %w", err)
	}
	if err := copyFields(&src.Status, &dst.Status); err != nil {
		return fmt.Errorf("converting status: %w", err)
	}
	// v1alpha1 spec.ipPools defaults to an empty list, and clients that compare specs with
	// reflect.DeepEqual see nil and empty as different, so it is never nil.
	dst.Spec.IPPools = make([]IPPoolReference, len(src.Status.IPPools))
	for i, ref := range src.Status.IPPools {
		dst.Spec.IPPools[i] = IPPoolReference{Name: ref.Name, APIVersion: ref.APIVersion}
	}
	return nil
}

// copyFields copies the fields that both versions have under the same names. The two spec and
// status types are identical except for ipPools, which the callers move by hand. A field added to
// one version only is dropped here, which the round-trip tests catch.
func copyFields(src, dst any) error {
	raw, err := json.Marshal(src)
	if err != nil {
		return err
	}
	return json.Unmarshal(raw, dst)
}
