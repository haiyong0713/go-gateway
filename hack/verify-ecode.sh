#!/bin/bash

set -o nounset

export dir=$(dirname $1)

if [[ $dir != "app/web-svr/activity" && $dir != "app/web-svr/esports" ]]; then
  exit 0
fi

if [[ -d "$dir/ecode" ]]; then
  dir="$dir/ecode"
else
  exit 0
fi

dup=$(cat $dir/*.go | egrep -o 'New\([0-9]+\)' | sort | uniq -d | egrep -o '[0-9]+')

if [[ "${dup}" = "" ]]; then
    echo "code检查正常"
    exit 0
else
    echo -e "Error: ${dir}下发现重复code,明细如下:\n${dup}" >&2
    exit 1
fi
