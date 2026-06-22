#!/usr/bin/env bash

set -o errexit
set -o pipefail
set -o nounset

TOPLEVEL="$(git rev-parse --show-toplevel)"
TOOLS_PATH=hack/tools/bin
PKG=github.com/vmware-tanzu/net-operator-api

VERSION=v1alpha1

CLIENTSET_NAME=clientset
CLIENT_OUT=$TOPLEVEL/pkg/client
CLIENTGEN_PATH=$PKG/pkg/client/clientset_generated
LISTERGEN_PATH=$PKG/pkg/client/listers_generated
INFORMERGEN_PATH=$PKG/pkg/client/informers_generated
HEADER_FILE=hack/boilerplate/boilerplate.go.txt

$TOOLS_PATH/client-gen --go-header-file $HEADER_FILE --input-base $PKG/api --input /$VERSION \
 --output-dir $CLIENT_OUT/clientset_generated/$CLIENTSET_NAME \
 --output-pkg $CLIENTGEN_PATH/$CLIENTSET_NAME

$TOOLS_PATH/lister-gen --go-header-file $HEADER_FILE \
 --output-dir $CLIENT_OUT/listers_generated \
 --output-pkg $LISTERGEN_PATH \
 $PKG/api/$VERSION

$TOOLS_PATH/informer-gen --single-directory --go-header-file $HEADER_FILE \
 --output-dir $CLIENT_OUT/informers_generated \
 --output-pkg $INFORMERGEN_PATH \
 --listers-package $LISTERGEN_PATH \
 --versioned-clientset-package $CLIENTGEN_PATH/$CLIENTSET_NAME \
 $PKG/api/$VERSION
