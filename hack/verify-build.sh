#!/bin/bash

set -o errexit
set -o nounset
set -o pipefail

export GOPROXY="http://goproxy.bilibili.co"
export GO111MODULE="on"

export dir="$1"
if [[ $dir == "app/app-svr/app-wall/job" || $dir == "app/app-svr/app-wall/interface" || $dir == "app/app-svr/up-archive/job" || $dir == "app/app-svr/up-archive/service" ]];
then
  cd ./$dir
  export options="build cmd/..."
elif [[ -d "$dir/cmd" ]]; then
  dir="$dir/cmd"
  export options="build ./$dir/..."
else
  export options="build ./$dir/..."
fi

go env

echo -e "\n${options}"

go ${options}
