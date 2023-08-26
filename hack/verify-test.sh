#!/bin/bash

CI_SERVER_URL="http://git.bilibili.co"
CI_UATSVEN_URL="http://sven.bilibili.co"
CI_PROJECT_ID=${REPO_OWNER}"%2F"${REPO_NAME}
CI_COMMIT_SHA=${PULL_PULL_SHA}
CI_COMMIT_USER=${TRIGGER}

export APP_ID=$1
export GOPROXY="http://goproxy.bilibili.co"
export GO111MODULE="on"

# get packages
declare -a packages
declare -a projects
declare mergeUser   
declare mergeID
declare appsJson

function GetPackages(){
    cd $GOPATH/src/${REPO_NAME}
    reg=".*/dao/.*\.go"
    if [[ ${UNIT_TEST_ALL} = "true" ]];then
        reg+="|${APP_ID}.*/service/.*\.go"
    fi
    files=$(git diff ${PULL_BASE_SHA}...${PULL_PULL_SHA} --name-only  --diff-filter=ACM | grep -E -i "${APP_ID}" | grep  -E "${reg}")
    if [[ "${files}" = "" ]]; then
        echo "shell.GetPackages: no change files"
        exit 0
    fi
    for value in ${files}
    do
        if [[ "${value}" =~ "/mock" ]]; then
            continue
        fi
        package="${REPO_NAME}/$(dirname ${value})"
        if [[ ${packages} =~ ${package} ]]; then
            continue
        fi
        packages+=${package}" "
        if [[ $1 =~ "library/" ]]; then
            continue
        fi
    done
    if [[ ${packages} = "" ]]; then
        echo "shell.GetPackages no change packages"
        exit 0
    fi
}

function GetProjects(){
    projects=$(find app -type d -name 'cmd' | sed 's/\/cmd//g')
    for project in ${projects}
    do
        if [ ! -f "${project}/OWNERS" ];then
            echo "project(${project}) 没有找到 OWNERS 文件!\n";
            continue
        fi
        owner=""
        substr=${project#*REPO_NAME}
        while read line
        do
            if [[ "${line}" = "#"* ]] || [[ "${line}" = "" ]] || [[ "${line}" = "approvers:" ]];then
                continue
            elif [[ "${line}" = "labels:"* ]];then
                break
            else
                owner+=$(echo "${line[@]:2}," | sed 's/\"//g')
            fi
        done < "${project}/OWNERS"
        appsJson+="{\"path\":\"${REPO_NAME}/${substr}\",\"owner\":\"${owner%,}\",\"repo_name\":\"${REPO_OWNER}/${REPO_NAME}\"},"
    done
    # delete "," at the end of value
    appsJson="[${appsJson%,}]"
}

# GetUserInfo get userinfo by gitlab result.
function GetUserInfo(){
    gitMergeRequestUrl="${CI_SERVER_URL}/api/v4/projects/${CI_PROJECT_ID}/repository/commits/${CI_COMMIT_SHA}/merge_requests?private_token=${PROW_TOKEN}"
    mergeJson=$(curl -s ${gitMergeRequestUrl})
    if [[ "${mergeJson}" = "[]" || "${mergeJson}" = "" ]]; then
        echo "Test not run, maybe you should try create a merge request first!"
        exit 0
    fi
    mergeID=$(echo ${mergeJson} | tr '\r\n' ' ' | jq -r '.[0].iid')
    mergeUser=$(echo ${mergeJson} | tr '\r\n' ' ' | jq -r '.[0].author.username')
}

# Magic ignore method Check()
function Magic(){
    url="http://git.bilibili.co/api/v4/projects/${CI_PROJECT_ID}/merge_requests/${mergeID}/notes?private_token=${PROW_TOKEN}"
    json=$(curl -s ${url})
    admin="haoguanwei,chenjianrong,fengshanshan,zhaobingqing"
    len=$(echo ${json} | jq '.|length')
    for i in $(seq 0 $len)
    do
        comment=$(echo ${json} | jq -r ".[$i].body")
        user=$(echo ${json} | jq -r ".[$i].author.username")
        if [[ ${comment} = "+skiput" && ${admin} =~ ${user} ]]; then
             exit 0
        fi
    done
}

# Check determine whether the standard is up to standard
#$1: commit_id
function Check(){
    curl -s "${CI_UATSVEN_URL}/x/admin/ut/git/report?repo_name=${CI_PROJECT_ID}&merge_id=${mergeID}&commit_id=${CI_COMMIT_SHA}"
    checkURL="${CI_UATSVEN_URL}/x/admin/ut/check?repo_name=${REPO_OWNER}/${REPO_NAME}&commit_id=${CI_COMMIT_SHA}"
    json=$(curl -s ${checkURL})
    code=$(echo ${json} | jq -r '.code')
    if [[ ${code} -ne 0 ]]; then
        echo -e "curl ${checkURL} response(${json})"
        exit 1
    fi
    package=$(echo ${json} | jq -r '.data.package')
    coverage=$(echo ${json} | jq -r '.data.coverage')
    passRate=$(echo ${json} | jq -r '.data.pass_rate')
    standard=$(echo ${json} | jq -r '.data.standard')
    increase=$(echo ${json} | jq -r '.data.increase')
    tyrant=$(echo ${json} | jq -r '.data.tyrant')
    lastCID=$(echo ${json} | jq -r '.data.last_cid')
    if ${tyrant}; then
        echo -e "\t续命失败!\n\t大佬，本次执行结果未达标哦(灬ꈍ ꈍ灬)，请再次优化ut重新提交🆙"
        echo -e "\t---------------------------------------------------------------------"
        printf "\t%-14s %-14s %-14s %-14s\n" "本次覆盖率(%)" "本次通过率(%)" "本次增长量(%)" "执行pkg"
        printf "\t%-13.2f %-13.2f %-13.2f %-12s\n" ${coverage} ${passRate} ${increase} ${package}
        echo -e "\t(达标标准：覆盖率>=${standard} && 通过率=100% && 同比当前package历史最高覆盖率的增长率>=0)"
        echo -e "\t---------------------------------------------------------------------"
        echo -e "本次执行详细结果查询地址请访问：${CI_UATSVEN_URL}/#/ut?merge_id=${mergeID}&&pn=1&ps=20"
        exit 1
    else
        echo -e "\t恭喜你，续命成功，可以请求合并MR了!"
    fi
}

# UTLint check the *_test.go files in the pkg
# $1: pkg
function UTLint(){
    path=${1//$REPO_NAME\//}
    declare -i numCase=0
    declare -i numAssertion=0
    files=$(ls ${path} | grep -E "(.*)_test\.go")
    if [[ ${#files} -eq 0 ]];then
        echo "RunPKGUT.UTLint no *_test.go files in pkg($1)"
        exit 1
    fi
    for file in ${files}
    do
        numCase+=`grep -c -E "^func Test(.+)\(t \*testing\.T\) \{$" ${path}/${file}`
        numAssertion+=`grep -c -E "^(.*)So\((.+)\)$" ${path}/${file}`
    done
    if [[ ${path} =~ "/library" && ${numCase} -eq 0 ]];then
        echo -e "RunPKGUT.UTLint no test case in pkg($1)"
        exit 1
    fi
    if [[ ${path} =~ "/app" && (${numCase} -eq 0 || ${numAssertion} -eq 0) ]];then
        echo -e "RunPKGUT.UTLint no test case or assertion in pkg($1)"
        exit 1
    fi
}

# GoTest execute go test and go tool
# $1: pkg
function GoTest(){
    go test -v -gcflags=-l $1 -coverprofile=cover.out -covermode=set -convey-json -timeout=60s > result.out
    #echo "##### TestCli Run #####"
    #testcli -f $GOPATH/src/${REPO_NAME}/${APP_ID}/resource/docker-compose.yaml run go test -v -gcflags=-l $1 -coverprofile=cover.out -covermode=set -convey-json -timeout=60s > result.out
    if [[ ! -s result.out ]]; then
        echo "==================================WARNING!======================================"
        echo "No test case found,请完善如下路径测试用例： $1 "
        exit 1
    else
        go tool cover -html=cover.out -o cover.html
    fi
}

# upload data to apm
# $1: file result.out path
function Upload () {
    if [[ ! -f "result.out" ]] || [[ ! -f "cover.html" ]] || [[ ! -f "cover.out" ]]; then
        echo "==================================WARNING!======================================"
        echo "No test found!~ 请完善如下路径测试用例： ${1} "
        exit 1
    fi
    url="${CI_UATSVEN_URL}/x/admin/ut/upload?repo_name=${REPO_OWNER}/${REPO_NAME}&merge_id=${mergeID}&username=${mergeUser}&author=${CI_COMMIT_USER}&commit_id=${CI_COMMIT_SHA}&pkg=${1}"
    json=$(curl -s ${url} -H "Content-type: multipart/form-data" -F "html_file=@cover.html" -F "report_file=@result.out" -F "data_file=@cover.out")
    if [[ "${json}" = "" ]]; then
        echo "RunPKGUT.Upload curl ${url} fail"
        exit 1
    fi
    msg=$(echo ${json} | jq -r '.message')
    data=$(echo ${json} | jq -r '.data')
    code=$(echo ${json} | jq -r '.code')
    if [[ ${code} -ne 0 ]]; then
        echo "=============================================================================="
        echo -e "RunPKGUT.Upload Response. message(${msg})"
        echo -e "RunPKGUT.Upload Response. data(${data})\n\n"
        echo -e "RunPKGUT.Upload Upload Fail! status(${code})"
        exit ${code}
    fi
}

function RunPKGUT(){
    for package in $packages
    do
        echo "RunPKGUT.UTLint Start. pkg(${package})"
        UTLint ${package}
        echo "RunPKGUT.GoTest Start. pkg(${package})"
        GoTest ${package}
        echo "RunPKGUT.Upload Start. pkg(${package})"
        Upload ${package}
    done
    return 0
}

function UploadApp(){
    url="${CI_UATSVEN_URL}/x/admin/ut/upload/app"
    json=$(curl -s $url -H 'content-type: application/json' -d $appsJson)
    if [[ "${json}" = "" ]]; then
        echo "UploadApp curl ${url} fail"
        exit 1
    fi
    msg=$(echo ${json} | jq -r '.message')
    data=$(echo ${json} | jq -r '.data')
    code=$(echo ${json} | jq -r '.code')
    if [[ ${code} -ne 0 ]]; then
        echo "=============================================================================="
        echo -e "UploadApp Response. message(${msg})"
        echo -e "UploadApp Response. data(${data})\n\n"
        echo -e "UploadApp Upload Fail! status(${code})"
        exit ${code}
    fi
}

# run
GetPackages
GetProjects
GetUserInfo
Magic
RunPKGUT
UploadApp
Check