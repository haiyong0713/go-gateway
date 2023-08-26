package white_list

import (
	"go-gateway/app/web-svr/native-page/interface/model"
)

const (
	FromWhitelist = "whitelist" //白名单接口
	// 白名单操作来源
	WhitelistOpAdd        = "add"
	WhitelistOpWaitSave   = "wait_save"
	WhitelistOpOnlineSave = "online_save"
)

type WhiteList struct {
	ID          int           `json:"id"`
	Mid         int64         `json:"mid"`
	Creator     string        `json:"creator"`
	CreatorUID  int           `json:"creator_uid"`
	Modifier    string        `json:"modifier"`
	ModifierUID int           `json:"modifier_uid"`
	FromType    string        `json:"from_type"`
	State       int           `json:"state"`
	Ctime       model.StrTime `json:"ctime"`
	Mtime       model.StrTime `json:"mtime"`
}
