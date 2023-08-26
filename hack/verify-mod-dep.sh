#!/bin/bash

set -o errexit
set -o nounset

export GOPROXY="http://goproxy.bilibili.co"
export GO111MODULE="on"

go env

set +e
changefiles=$(git diff ${PULL_BASE_SHA} ${PULL_PULL_SHA} go.mod |grep -v go-common |grep -v = |egrep "\+\t" |awk '{split($0,a,"[\tv]");print a[2]}')
set -e

if [[ "${changefiles}" = "" ]]; then
    exit 0
fi

set +e
changes=$(git diff ${PULL_BASE_SHA} ${PULL_PULL_SHA} go.mod |grep -v go-common |grep -v = |egrep "\+\t" |awk '{split($0,a,"[\tv]");print a[2]}' |xargs go mod why |grep -v "#" | grep go-main)
set -e
#check deps change expect replace
if [[ "${changes}" = "" ]]; then
  echo ${changes}|xargs go build -v
fi

set +e
changes=$(git diff ${PULL_BASE_SHA} ${PULL_PULL_SHA} go.mod |grep -v go-common |grep + |grep =|awk '{split($0,a,"(replace|=)");print a[2]}'|xargs go mod why|grep -v "#" |grep go-main)
set -e
#check deps change in replace
if [[ "${changes}" = "" ]]; then
  echo ${changes}| xargs go build -v
fi
