#!/bin/bash

set -o errexit

export GOPROXY="http://goproxy.bilibili.co"
export GO111MODULE="on"

go env

echo -e "\ngo mod tidy"

go mod tidy -v

diff=$(git diff --name-only)

if [[ -n "${diff}" ]]; then
  echo "ERROR: 请将你的分支合并 master, 并在根目录下面执行 'make mod-update', 并提交 go.mod  或 go.sum." >&2
  git diff
  exit 1
fi
