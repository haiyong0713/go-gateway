#!/bin/bash

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

go install ./hack/tools/security

files=$(git diff ${PULL_BASE_SHA} --name-only  --diff-filter=ACM)

echo -e "变更的文件如下:\n$files"

security ${files}

exit $?