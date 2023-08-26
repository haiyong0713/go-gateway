package config

import (
	"go-gateway/app/app-svr/native-act/interface/kernel"
)

type DynamicAct struct {
	BaseCfgManager

	ImageTitle    string
	TextTitle     string
	IsFeed        bool
	Sid           int64 //数据源id
	SortType      int64 //排序
	SortList      []*SortListItem
	FontColor     string
	BgColor       string
	PageSize      int32
	SubpageTitle  string
	ActLikesReqID kernel.RequestID
	// 老版本所需参数
	PageID int64
}
