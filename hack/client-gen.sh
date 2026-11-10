#!/usr/bin/env bash
# © Broadcom. All Rights Reserved.
# The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
# SPDX-License-Identifier: Apache-2.0

set -o errexit
set -o pipefail
set -o nounset

TOPLEVEL="$(git rev-parse --show-toplevel)"
TOOLS_PATH=hack/tools/bin
PKG=github.com/vmware-tanzu/net-operator-api

VERSIONS=(v1alpha1 v1alpha2)
INPUTS=$(IFS=,; echo "${VERSIONS[*]/#/api/}")
PACKAGES=("${VERSIONS[@]/#/$PKG/api/}")

CLIENTSET_NAME=clientset
CLIENT_OUT=$TOPLEVEL/pkg/client
CLIENTGEN_PATH=$PKG/pkg/client/clientset_generated
LISTERGEN_PATH=$PKG/pkg/client/listers_generated
INFORMERGEN_PATH=$PKG/pkg/client/informers_generated
HEADER_FILE=hack/boilerplate/boilerplate.go.txt

$TOOLS_PATH/client-gen --go-header-file $HEADER_FILE --input-base $PKG --input "$INPUTS" \
 --clientset-name $CLIENTSET_NAME \
 --output-dir "$CLIENT_OUT"/clientset_generated \
 --output-pkg $CLIENTGEN_PATH

$TOOLS_PATH/lister-gen --go-header-file $HEADER_FILE \
 --output-dir "$CLIENT_OUT"/listers_generated \
 --output-pkg $LISTERGEN_PATH \
 "${PACKAGES[@]}"

$TOOLS_PATH/informer-gen --single-directory --go-header-file $HEADER_FILE \
 --output-dir "$CLIENT_OUT"/informers_generated \
 --output-pkg $INFORMERGEN_PATH \
 --listers-package $LISTERGEN_PATH \
 --versioned-clientset-package $CLIENTGEN_PATH/$CLIENTSET_NAME \
 "${PACKAGES[@]}"
