#!/bin/bash

set -e
test -z ${1} && {
    echo "Usage: $0 /path/to/file.pcap"
    exit 1
}
test -e ${1}

URL="http://portal.bilibili.co/x/free/external/pcap"
echo Uploading: $URL
curl -X POST \
    $URL \
    -F file=@$1
