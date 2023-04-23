#!/usr/bin/env bash

set -euo pipefail

VER=${1:-}
REPOSITORY="${REPOSITORY:-https://github.com/ahrtr/etcd-defrag.git}"

if [ -z "$VER" ]; then
  echo "Usage: ${0} VERSION" >> /dev/stderr
  exit 255
fi

function setup_env {
  local ver=${1}
  local proj=${2}

  if [ ! -d "${proj}" ]; then
    git clone "${REPOSITORY}"
  fi

  pushd "${proj}" >/dev/null
    git fetch --all
    git checkout "${ver}"
  popd >/dev/null
}

function package {
  local target=${1}
  local srcdir="${2}/bin"

  local ccdir="${srcdir}/${GOOS}_${GOARCH}"
  if [ -d "${ccdir}" ]; then
    srcdir="${ccdir}"
  fi
  local ext=""
  if [ "${GOOS}" == "windows" ]; then
    ext=".exe"
  fi
  cp "${srcdir}/etcd-defrag" "${target}/etcd-defrag${ext}"

  cp etcd-defrag/README.md "${target}"/README.md
}

function main {
  local proj="etcd-defrag"

  mkdir -p release
  cd release
  setup_env "${VER}" "${proj}"

  local tarcmd=tar

  for os in darwin windows linux; do
    export GOOS=${os}
    TARGET_ARCHS=("amd64")

    if [ ${GOOS} == "linux" ]; then
      TARGET_ARCHS+=("arm64")
      TARGET_ARCHS+=("ppc64le")
      TARGET_ARCHS+=("s390x")
    fi

    if [ ${GOOS} == "darwin" ]; then
      TARGET_ARCHS+=("arm64")
    fi

    for TARGET_ARCH in "${TARGET_ARCHS[@]}"; do
      export GOARCH=${TARGET_ARCH}

      pushd ${proj} >/dev/null
      GO_LDFLAGS="-s -w" ./build.sh
      popd >/dev/null

      TARGET="${proj}-${VER}-${GOOS}-${GOARCH}"
      mkdir "${TARGET}"
      package "${TARGET}" "${proj}"

      if [ ${GOOS} == "linux" ]; then
        ${tarcmd} cfz "${TARGET}.tar.gz" "${TARGET}"
        echo "Wrote release/${TARGET}.tar.gz"
      else
        zip -qr "${TARGET}.zip" "${TARGET}"
        echo "Wrote release/${TARGET}.zip"
      fi
    done
  done

  echo "Generating sha256sums of release artifacts."
  ls . | grep -E '\.tar.gz$|\.zip$' | xargs shasum -a 256 > ./SHA256SUMS
}

main


