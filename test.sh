#!/usr/bin/env bash

set -e
echo "" > coverage.txt

# Podman client build tags — without `remote` the bindings pull the local
# libpod tree (cgo, btrfs, unix-only rlimit) and the build breaks. See
# docs/adr/0005-podman-native-backend.md.
export GOFLAGS="-mod=vendor -tags=containers_image_openpgp,exclude_graphdriver_btrfs,exclude_graphdriver_devicemapper,remote"

use_go_test=false
if command -v gotest; then
    use_go_test=true
fi

for d in $( find ./* -maxdepth 10 ! -path "./vendor*" ! -path "./.git*" ! -path "./scripts*" -type d); do
    if ls $d/*.go &> /dev/null; then
        # The Docker runtime is gated behind `-tags docker` (its SDK is kept
        # out of the default build), so every file in pkg/runtime/docker is
        # excluded here and `go test` would error with "build constraints
        # exclude all Go files". Skip such packages; the docker job exercises
        # them with the tag.
        if ! go list "$d" &> /dev/null; then
            continue
        fi
        args="-race -coverprofile=profile.out -covermode=atomic $d"
        if [ "$use_go_test" == true ]; then
            gotest $args
        else
            go test $args
        fi
        if [ -f profile.out ]; then
            cat profile.out >> coverage.txt
            rm profile.out
        fi
    fi
done
