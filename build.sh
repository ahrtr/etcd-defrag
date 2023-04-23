#!/usr/bin/env bash

set -euo pipefail

GIT_SHA=$(git rev-parse --short HEAD || echo "GitNotFound")
VERSION_SHA="main.GitSHA"

# use go env if noset
GOOS=${GOOS:-$(go env GOOS)}
GOARCH=${GOARCH:-$(go env GOARCH)}

GO_BUILD_FLAGS=${GO_BUILD_FLAGS:-}

GO_LDFLAGS=(${GO_LDFLAGS:-} "-X=${VERSION_SHA}=${GIT_SHA}")
GO_BUILD_ENV=("CGO_ENABLED=0" "GO_BUILD_FLAGS=${GO_BUILD_FLAGS}" "GOOS=${GOOS}" "GOARCH=${GOARCH}")

rm -f ./bin/etcd-defrag

env "${GO_BUILD_ENV[@]}" go build $GO_BUILD_FLAGS -trimpath -installsuffix=cgo "-ldflags=${GO_LDFLAGS[*]}" -o ./bin/etcd-defrag

