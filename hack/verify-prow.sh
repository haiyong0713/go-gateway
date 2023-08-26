#!/usr/bin/env bash

set -o errexit
set -o pipefail

export GOPROXY="http://goproxy.bilibili.co"
export GO111MODULE="on"

ROOT="$(cd "$(dirname "${BASH_SOURCE}")/.." && pwd -P)"

OUTPUT_SUBPATH="${OUTPUT_SUBPATH:-_output/local}"
OUTPUT_BINPATH="${ROOT}/${OUTPUT_SUBPATH}/bin"

# Ensure that we find the binaries we build before anything else.
export GOBIN="${OUTPUT_BINPATH}"
PATH="${GOBIN}:${PATH}"

# Install tools we need, but only from vendor/...
go install ./hack/tools/owner
go install ./hack/tools/mkprow
go install ./hack/tools/labels

# Check owner file
if ! owner; then
    echo "ERROR: 请将你的分支合并 master, 并在根目录下面执行 'make prow-update', 并提交变更内容." >&2
    exit 1
fi

# Check owner file
if ! mkprow; then
    echo "ERROR: 请将你的分支合并 master, 并在根目录下面执行 'make prow-update', 并提交变更内容." >&2
    exit 1
fi

# Check owner file
if ! labels; then
    echo "ERROR: 请将你的分支合并 master, 并在根目录下面执行 'make prow-update', 并提交变更内容." >&2
    exit 1
fi