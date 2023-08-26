package main

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestListFile(t *testing.T) {
	path := `./test-proto`
	bapis := &Bapi{Hmap: make(map[string][]*BFile)}
	err := listFile(path, "", bapis)
	assert.Nil(t, err)
	files, ok := bapis.Hmap[""]
	fmt.Println(files)
	assert.Equal(t, ok, true)
	assert.Equal(t, 2, len(files))
}

func TestParseProto(t *testing.T) {
	bapis := &Bapi{Hmap: map[string][]*BFile{
		"": {
			{
				Name: "api2.proto",
				Content: `syntax = "proto3";
package dynamic.service.feed.v1;
import "extension/wdcli/wdcli.proto";

option go_package = "git.bilibili.co/bapis/bapis-go/dynamic/interface/feed;api";
option java_package = "com.bapis.dynamic.interfaces.feed";
option java_multiple_files = true;

option (wdcli.appid) = "main.dynamic.feed";

enum TabType1 {
    INVALID_TAB_TYPE = 0;
    TAB_TYPE_GENERAL = 1;
    TAB_TYPE_VIDEO   = 2;
}

message OffsetInfo1 {
    int32  tab         = 1;
    string type_list   = 2;
    string offset      = 3;
}

message UpdateNumReq1 {
    uint64 uid                  = 1;
    repeated OffsetInfo1 offsets = 2;
}

message UpdateNumResp1 {
    string red_type    = 1; // 红点类型 - count-数字红点 point-普通红点 no_point-没有红点
    uint64 update_num  = 2; // 更新数量 - 仅当 red_type = 2时有意义
    string default_tab = 3;
}

service Feed {
    // 网关调用 - 获取动态更新数量（客户端）
    rpc UpdateNum(UpdateNumReq1) returns (UpdateNumResp1);
}`,
				Prefix: "",
				Path:   "",
			},
			{
				Name: "extension/wdcli/wdcli.proto",
				Content: `syntax = "proto3";
package wdcli;

import "google/protobuf/descriptor.proto";

option go_package = "git.bilibili.co/bapis/bapis-go/extension/wdcli;wdcli";
option java_package = "com.bapis.extension.wdcli";
option java_multiple_files = true;

extend google.protobuf.FileOptions {
    string appid = 1000;
}
`,
			},
		},
	},
	}
	protoDesc, err := parseProto(bapis)
	assert.Nil(t, err)
	assert.Equal(t, 2, len(protoDesc))
	if len(protoDesc) != 1 {
		return
	}
	desc := protoDesc[0]
	assert.Equal(t, "api2.proto", desc.GetName())
	assert.Equal(t, "Feed", desc.GetServices()[0].GetName())
}
