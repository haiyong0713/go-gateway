#!/bin/bash

set -o errexit

export GOPROXY="http://goproxy.bilibili.co"
export GO111MODULE="on"

go env

echo -e "\ngo mod tidy"

go mod tidy -v

diff=$(git diff --name-only)

if [[ -n "${diff}" ]]; then
  echo "请将变更 go.mod  或 go.sum 提交到 gitlab" >&2
fi
