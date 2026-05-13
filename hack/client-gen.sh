#!/usr/bin/env bash

set -o errexit
set -o pipefail
set -o nounset

TOPLEVEL="$(git rev-parse --show-toplevel)"
TOOLS_PATH=hack/tools/bin
PKG=github.com/vmware-tanzu/net-operator-api

VERSION=v1alpha1

CLIENTGEN_PATH=$PKG/pkg/client/clientset_generated
LISTERGEN_PATH=$PKG/pkg/client/listers_generated
INFORMERGEN_PATH=$PKG/pkg/client/informers_generated
CLIENTSET_NAME=clientset
HEADER_FILE=hack/boilerplate/boilerplate.go.txt

$TOOLS_PATH/client-gen --go-header-file $HEADER_FILE --input-base $PKG/api --input /$VERSION \
    --clientset-path $CLIENTGEN_PATH --clientset-name $CLIENTSET_NAME

mv $TOPLEVEL/$PKG/pkg/client/clientset_generated/clientset/typed/v1alpha1/_client.go $TOPLEVEL/$PKG/pkg/client/clientset_generated/clientset/typed/v1alpha1/client.go

$TOOLS_PATH/lister-gen --input-dirs $PKG/api/$VERSION --go-header-file $HEADER_FILE --output-package $LISTERGEN_PATH

$TOOLS_PATH/informer-gen --single-directory --input-dirs $PKG/api/$VERSION --go-header-file $HEADER_FILE \
    --output-package $INFORMERGEN_PATH --listers-package $LISTERGEN_PATH --versioned-clientset-package $CLIENTGEN_PATH/$CLIENTSET_NAME

# Move to top level so that samples can consume the via the top-level import.
# while taking care to avoid other artifacts potentially in pkg.
mkdir -p "$TOPLEVEL/pkg/client"
rm -rf $TOPLEVEL/pkg/client/*_generated
mv -f $TOPLEVEL/$PKG/pkg/client/* pkg/client
rm -rf "$TOPLEVEL/github.com"