#!/usr/bin/env bash

set -o errexit
set -o pipefail
set +e

files=$(git diff ${PULL_BASE_SHA} --name-only  --diff-filter=ACM | grep -E -i "CHANGELOG.md")

if [[ "${files}" = "" ]]; then
    echo "Error: 未发现CHANGELOG.md文件变更，请'添加'或'修改'CHANGELOG.md" >&2
    exit 1
else
    echo -e "变更如下:\n${files}"
fi



