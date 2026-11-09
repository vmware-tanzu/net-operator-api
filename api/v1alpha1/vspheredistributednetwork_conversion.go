// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: Apache-2.0

package v1alpha1

import (
	"encoding/json"
	"fmt"
	"net"
	"reflect"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/conversion"

	"github.com/vmware-tanzu/net-operator-api/api/v1alpha2"
)

const (
	// VSphereDistributedNetworkV1alpha1DataAnnotation holds, on the v1alpha2 object, the parts of
	// a v1alpha1 VSphereDistributedNetwork that v1alpha2 represents differently, so that a
	// v1alpha1 client reads back exactly what it wrote. It is never visible through v1alpha1.
	VSphereDistributedNetworkV1alpha1DataAnnotation = "netoperator.vmware.com/v1alpha1-conversion-data"

	// VSphereDistributedNetworkV1alpha2DataAnnotation holds, on the v1alpha1 view, the v1alpha2
	// fields that v1alpha1 has no place for: the IPv4 address ranges and the IPv6 configuration.
	// It is restored when the object is written back through v1alpha1, so a v1alpha1 client that
	// updates a network keeps its ranges. It is never stored on the v1alpha2 object.
	VSphereDistributedNetworkV1alpha2DataAnnotation = "netoperator.vmware.com/v1alpha2-conversion-data"
)

// v1alpha1Data is the content of VSphereDistributedNetworkV1alpha1DataAnnotation. Each field is
// restored only while the v1alpha2 value it was derived from is unchanged.
//
//nolint:kubeapilinter // internal conversion state stored in an annotation, not an API type
type v1alpha1Data struct {
	// IPAssignmentModeUnset records that ipAssignmentMode was unset, which v1alpha1 treats as
	// staticpool.
	IPAssignmentModeUnset bool `json:"ipAssignmentModeUnset,omitempty"`
	// SubnetMask is subnetMask when formatting ipv4.prefixLength doesn't reproduce it, for example
	// a non-contiguous mask.
	SubnetMask *string `json:"subnetMask,omitempty"`
	// SubnetMaskPrefixLength is the ipv4.prefixLength SubnetMask was converted to.
	SubnetMaskPrefixLength *int32 `json:"subnetMaskPrefixLength,omitempty"`
}

// v1alpha2Data is the content of VSphereDistributedNetworkV1alpha2DataAnnotation.
//
//nolint:kubeapilinter // internal conversion state stored in an annotation, not an API type
type v1alpha2Data struct {
	// IPv4AddressRanges is spec.ipv4.addressRanges.
	IPv4AddressRanges []v1alpha2.VSphereDistributedNetworkIPRange `json:"ipv4AddressRanges,omitempty"`
	// IPv6 is spec.ipv6.
	IPv6 *v1alpha2.VSphereDistributedNetworkIPv6Config `json:"ipv6,omitempty"`
	// IPPools is status.ipPools of a network with IPv6 configuration whose
	// spec.ipv4.assignmentMode isn't staticpool. v1alpha1 requires spec.ipPools to be empty then,
	// so the network's (IPv6) pools are kept here.
	IPPools []v1alpha2.IPPoolReference `json:"ipPools,omitempty"`
}

var _ conversion.Convertible = &VSphereDistributedNetwork{}

// ConvertTo converts this VSphereDistributedNetwork to the v1alpha2 hub. v1alpha1 spec.ipPools
// becomes status.ipPools, and the IPv4 fields move to spec.ipv4.
func (src *VSphereDistributedNetwork) ConvertTo(dstRaw conversion.Hub) error {
	dst, ok := dstRaw.(*v1alpha2.VSphereDistributedNetwork)
	if !ok {
		return fmt.Errorf("unexpected hub type %T", dstRaw)
	}

	dst.ObjectMeta = *src.ObjectMeta.DeepCopy()
	annotations := dst.GetAnnotations()

	// A malformed stash fails the write rather than silently dropping the network's ranges,
	// which would make the operator delete their IPPools.
	var hubOnly v1alpha2Data
	if raw, ok := annotations[VSphereDistributedNetworkV1alpha2DataAnnotation]; ok {
		if err := json.Unmarshal([]byte(raw), &hubOnly); err != nil {
			return fmt.Errorf("reading annotation %s: %w", VSphereDistributedNetworkV1alpha2DataAnnotation, err)
		}
	}
	delete(annotations, VSphereDistributedNetworkV1alpha2DataAnnotation)

	var spokeOnly v1alpha1Data
	dst.Spec = convertSpecToV1alpha2(&src.Spec, &spokeOnly)
	dst.Spec.IPv4.AddressRanges = hubOnly.IPv4AddressRanges
	dst.Spec.IPv6 = v1alpha2.VSphereDistributedNetworkIPv6Config{}
	if hubOnly.IPv6 != nil {
		dst.Spec.IPv6 = *hubOnly.IPv6
	}

	status := v1alpha2.VSphereDistributedNetworkStatus{
		Conditions:        convertConditionsToV1alpha2(src.Status.Conditions),
		IPPools:           convertIPPoolRefsToV1alpha2(src.Spec.IPPools),
		DefaultPortConfig: convertPortConfigToV1alpha2(src.Status.DefaultPortConfig),
	}
	if len(status.IPPools) == 0 && len(hubOnly.IPPools) > 0 {
		status.IPPools = hubOnly.IPPools
	}
	dst.Status = nil
	if !reflect.ValueOf(status).IsZero() {
		dst.Status = &status
	}

	delete(annotations, VSphereDistributedNetworkV1alpha1DataAnnotation)
	if err := setAnnotationJSON(&annotations, VSphereDistributedNetworkV1alpha1DataAnnotation, spokeOnly); err != nil {
		return err
	}
	dst.SetAnnotations(annotations)
	return nil
}

// ConvertFrom converts the v1alpha2 hub to this VSphereDistributedNetwork. v1alpha2
// status.ipPools becomes spec.ipPools, and spec.ipv4 becomes the v1alpha1 IPv4 fields.
func (dst *VSphereDistributedNetwork) ConvertFrom(srcRaw conversion.Hub) error {
	src, ok := srcRaw.(*v1alpha2.VSphereDistributedNetwork)
	if !ok {
		return fmt.Errorf("unexpected hub type %T", srcRaw)
	}

	dst.ObjectMeta = *src.ObjectMeta.DeepCopy()
	annotations := dst.GetAnnotations()

	// Malformed data is ignored rather than failing the conversion, which would fail every read
	// of the object (and every list that includes it) through v1alpha1.
	var spokeOnly v1alpha1Data
	if raw, ok := annotations[VSphereDistributedNetworkV1alpha1DataAnnotation]; ok {
		_ = json.Unmarshal([]byte(raw), &spokeOnly)
	}
	delete(annotations, VSphereDistributedNetworkV1alpha1DataAnnotation)

	status := src.Status
	if status == nil {
		status = &v1alpha2.VSphereDistributedNetworkStatus{}
	}
	dst.Spec = convertSpecFromV1alpha2(&src.Spec, &spokeOnly)
	dst.Spec.IPPools = convertIPPoolRefsFromV1alpha2(status.IPPools)
	dst.Status = VSphereDistributedNetworkStatus{
		Conditions:        convertConditionsFromV1alpha2(status.Conditions),
		DefaultPortConfig: convertPortConfigFromV1alpha2(status.DefaultPortConfig),
	}

	hubOnly := v1alpha2Data{IPv4AddressRanges: src.Spec.IPv4.AddressRanges}
	if !reflect.ValueOf(src.Spec.IPv6).IsZero() {
		hubOnly.IPv6 = src.Spec.IPv6.DeepCopy()
		// Only a network with IPv6 configuration, which v1alpha1 can't express, has pools that
		// v1alpha1 must not show. Objects written through v1alpha1 keep their spec.ipPools as is.
		if src.Spec.IPv4.AssignmentMode != v1alpha2.IPAssignmentModeStaticPool && len(status.IPPools) > 0 {
			hubOnly.IPPools = status.IPPools
			dst.Spec.IPPools = []IPPoolReference{}
		}
	}
	delete(annotations, VSphereDistributedNetworkV1alpha2DataAnnotation)
	if err := setAnnotationJSON(&annotations, VSphereDistributedNetworkV1alpha2DataAnnotation, hubOnly); err != nil {
		return err
	}
	dst.SetAnnotations(annotations)
	return nil
}

// setAnnotationJSON stores data as JSON under key, or leaves key unset when data is its zero
// value. An empty map is replaced with nil.
func setAnnotationJSON(annotations *map[string]string, key string, data any) error {
	if !reflect.ValueOf(data).IsZero() {
		raw, err := json.Marshal(data)
		if err != nil {
			return fmt.Errorf("writing annotation %s: %w", key, err)
		}
		if *annotations == nil {
			*annotations = map[string]string{}
		}
		(*annotations)[key] = string(raw)
	}
	if len(*annotations) == 0 {
		*annotations = nil
	}
	return nil
}

func convertSpecToV1alpha2(src *VSphereDistributedNetworkSpec, data *v1alpha1Data) v1alpha2.VSphereDistributedNetworkSpec {
	dst := v1alpha2.VSphereDistributedNetworkSpec{
		PortGroupID:  src.PortGroupID,
		PortGroupIDs: append([]string(nil), src.PortGroupIDs...),
		IPv4: v1alpha2.VSphereDistributedNetworkIPv4Config{
			AssignmentMode: v1alpha2.IPAssignmentModeType(src.IPAssignmentMode),
			Gateway:        src.Gateway,
		},
	}
	if len(dst.PortGroupIDs) == 0 {
		dst.PortGroupIDs = nil
	}
	if src.IPAssignmentMode == "" {
		dst.IPv4.AssignmentMode = v1alpha2.IPAssignmentModeStaticPool
		data.IPAssignmentModeUnset = true
	}
	if src.SubnetMask != "" {
		prefix, ok := subnetMaskToPrefixLength(src.SubnetMask)
		if ok {
			dst.IPv4.PrefixLength = &prefix
		}
		if !ok || prefixLengthToSubnetMask(prefix) != src.SubnetMask {
			data.SubnetMask = new(src.SubnetMask)
			data.SubnetMaskPrefixLength = copyInt32(dst.IPv4.PrefixLength)
		}
	}
	return dst
}

func convertSpecFromV1alpha2(src *v1alpha2.VSphereDistributedNetworkSpec, data *v1alpha1Data) VSphereDistributedNetworkSpec {
	dst := VSphereDistributedNetworkSpec{
		PortGroupID:      src.PortGroupID,
		PortGroupIDs:     append([]string(nil), src.PortGroupIDs...),
		IPAssignmentMode: IPAssignmentModeType(src.IPv4.AssignmentMode),
		Gateway:          src.IPv4.Gateway,
	}
	if len(dst.PortGroupIDs) == 0 {
		dst.PortGroupIDs = nil
	}
	if data.IPAssignmentModeUnset && src.IPv4.AssignmentMode == v1alpha2.IPAssignmentModeStaticPool {
		dst.IPAssignmentMode = ""
	}
	switch {
	case data.SubnetMask != nil && reflect.DeepEqual(data.SubnetMaskPrefixLength, src.IPv4.PrefixLength):
		dst.SubnetMask = *data.SubnetMask
	case src.IPv4.PrefixLength != nil:
		dst.SubnetMask = prefixLengthToSubnetMask(*src.IPv4.PrefixLength)
	}
	return dst
}

// subnetMaskToPrefixLength returns the prefix length of a contiguous dotted-quad IPv4 mask.
func subnetMaskToPrefixLength(mask string) (int32, bool) {
	ip := net.ParseIP(mask).To4()
	if ip == nil {
		return 0, false
	}
	ones, bits := net.IPv4Mask(ip[0], ip[1], ip[2], ip[3]).Size()
	if bits != 32 {
		return 0, false
	}
	return int32(ones), true
}

func prefixLengthToSubnetMask(prefix int32) string {
	if prefix < 0 || prefix > 32 {
		return ""
	}
	return net.IP(net.CIDRMask(int(prefix), 32)).String()
}

// convertIPPoolRefsToV1alpha2 returns nil for an empty list, matching the omitempty v1alpha2
// status field.
func convertIPPoolRefsToV1alpha2(refs []IPPoolReference) []v1alpha2.IPPoolReference {
	if len(refs) == 0 {
		return nil
	}
	out := make([]v1alpha2.IPPoolReference, len(refs))
	for i, r := range refs {
		out[i] = v1alpha2.IPPoolReference{Name: r.Name, APIVersion: r.APIVersion}
	}
	return out
}

// convertIPPoolRefsFromV1alpha2 never returns nil: v1alpha1 spec.ipPools defaults to an empty
// list, and clients that compare specs with reflect.DeepEqual see nil and empty as different.
func convertIPPoolRefsFromV1alpha2(refs []v1alpha2.IPPoolReference) []IPPoolReference {
	out := make([]IPPoolReference, len(refs))
	for i, r := range refs {
		out[i] = IPPoolReference{Name: r.Name, APIVersion: r.APIVersion}
	}
	return out
}

// convertConditionsToV1alpha2 maps v1alpha1 conditions to metav1.Condition. v1alpha1 has no
// observedGeneration, so it is zero until the operator next writes the condition.
func convertConditionsToV1alpha2(conditions []VSphereDistributedNetworkCondition) []metav1.Condition {
	if len(conditions) == 0 {
		return nil
	}
	out := make([]metav1.Condition, len(conditions))
	for i, c := range conditions {
		out[i] = metav1.Condition{
			Type:               string(c.Type),
			Status:             metav1.ConditionStatus(c.Status),
			Reason:             c.Reason,
			Message:            c.Message,
			LastTransitionTime: c.LastTransitionTime,
		}
	}
	return out
}

func convertConditionsFromV1alpha2(conditions []metav1.Condition) []VSphereDistributedNetworkCondition {
	if len(conditions) == 0 {
		return nil
	}
	out := make([]VSphereDistributedNetworkCondition, len(conditions))
	for i, c := range conditions {
		out[i] = VSphereDistributedNetworkCondition{
			Type:               VSphereDistributedNetworkConditionType(c.Type),
			Status:             corev1.ConditionStatus(c.Status),
			Reason:             c.Reason,
			Message:            c.Message,
			LastTransitionTime: c.LastTransitionTime,
		}
	}
	return out
}

// convertPortConfigToV1alpha2 and convertPortConfigFromV1alpha2 copy through JSON: the port
// configuration types are identical in both versions.
func convertPortConfigToV1alpha2(src *VSphereDistributedPortConfig) *v1alpha2.VSphereDistributedPortConfig {
	if src == nil {
		return nil
	}
	dst := &v1alpha2.VSphereDistributedPortConfig{}
	copyViaJSON(src, dst)
	return dst
}

func convertPortConfigFromV1alpha2(src *v1alpha2.VSphereDistributedPortConfig) *VSphereDistributedPortConfig {
	if src == nil {
		return nil
	}
	dst := &VSphereDistributedPortConfig{}
	copyViaJSON(src, dst)
	return dst
}

func copyViaJSON(src, dst any) {
	raw, err := json.Marshal(src)
	if err != nil {
		return
	}
	_ = json.Unmarshal(raw, dst)
}

func copyInt32(p *int32) *int32 {
	if p == nil {
		return nil
	}
	return new(*p)
}
