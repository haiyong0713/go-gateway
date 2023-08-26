package config

import (
	"go-gateway/app/app-svr/native-act/interface/kernel"
)

type Dynamic struct {
	BaseCfgManager

	ImageTitle    string
	TextTitle     string
	IsFeed        bool
	TopicID       int64
	PickID        int64   //精选内容-数据id
	Contents      []int64 //内容选择
	SortBy        int32
	FontColor     string
	BgColor       string
	PageSize      int32
	HasDynsReqID  kernel.RequestID
	ListDynsReqID kernel.RequestID
	IsMaster      bool //是否是主态访问
	// 老版本所需参数
	ModuleTitle string
	Sort        string
	PageTitle   string
	PageID      int64
}
