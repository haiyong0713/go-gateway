#!/usr/bin/env bash

GO111MODULE=on go get golang.org/x/tools/gopls@latest
go mod tidy
go mod download

export PATH=$PATH:`go env GOPATH`/bin

pip3 install -i https://pypi.tuna.tsinghua.edu.cn/simple awscli

export AWS_ACCESS_KEY_ID=ecaababdd179dfe5
export AWS_SECRET_ACCESS_KEY=59e6bc1a3204c8b30545983019f9bc8c
export AWS_DEFAULT_REGION=uat

aws configure list

aws --endpoint-url=http://uat-boss.bilibili.co s3 ls
aws --endpoint-url=http://uat-boss.bilibili.co s3 cp s3://misc/diff-reference-linux-v0.0.15 /tmp/diff-reference-linux
chmod +x /tmp/diff-reference-linux

git diff origin/master -U0 | /tmp/diff-reference-linux

filename=`cat /dev/urandom | tr -dc 'a-zA-Z0-9' | fold -w 32 | head -n 1`-go-gateway.html

aws --endpoint-url=http://uat-boss.bilibili.co s3 cp /tmp/go-gateway-target.html s3://misc/client/$filename

html="http://uat-boss.bilibili.co/misc/client/$filename"
echo $html
body="本次提交改动影响的服务详情：%0a%0a|%20commit%20SHA%20|%20网址%20|%0a|%20------%20|%20------%20|%0a|%20$3%20|%20$html%20|"
curl="curl --request POST --header 'PRIVATE-TOKEN: $2' 'https://git.bilibili.co/api/v4/projects/13721/merge_requests/$1/notes?body=$body'"
eval "$curl"
