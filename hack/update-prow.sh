#!/bin/bash

set -o errexit
set -o nounset
set -o pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE}")/.." && pwd -P)"

OUTPUT_SUBPATH="${OUTPUT_SUBPATH:-_output/local}"
OUTPUT_BINPATH="${ROOT}/${OUTPUT_SUBPATH}/bin"

# Ensure that we find the binaries we build before anything else.
export GOBIN="${OUTPUT_BINPATH}"
PATH="${GOBIN}:${PATH}"

export GOPROXY="http://goproxy.bilibili.co"
export GO111MODULE="on"

# Install tools we need, but only from vendor/...
go install ./hack/tools/owner
go install ./hack/tools/mkprow
go install ./hack/tools/labels

# Check owner file
if ! owner; then
    echo "ERROR: 请注意 OWNERS 文件变更, 请提交变更内容, 参考文档: hack/tools/owner/README.MD " >&2
fi

# Check mkprow file
if ! mkprow; then
    echo "ERROR: 请注意 hack/prow/go_gateway_jobs.yaml 有变更, 请提交变更内容, 参考文档: hack/tools/mkprow/README.MD" >&2
fi

# Check labels file
if ! labels; then
    echo "ERROR: 请注意 hack/prow/labels.yaml 有变更, 请提交变更内容, 参考文档: hack/tools/labels/README.MD" >&2
fi
