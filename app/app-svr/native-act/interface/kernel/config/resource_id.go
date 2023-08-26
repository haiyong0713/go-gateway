package config

import (
	"go-gateway/app/app-svr/native-act/interface/kernel"
)

type ResourceID struct {
	BaseCfgManager

	ResourceCommon
	MixExtsReqID kernel.RequestID
}

type ResourceCommon struct {
	ImageTitle          string //图片标题
	TextTitle           string //文字标题
	SubpageTitle        string //二级页面标题
	DisplayUGCBadge     bool   //是否展示ugc角标
	DisplayPGCBadge     bool   //是否展示pgc角标
	DisplayArticleBadge bool   //是否展示专栏角标
	DisplayViewMore     bool   //是否展示查看更多按钮
	DisplayOnlyLive     bool   //是否只展示开播状态直播间
	BgColor             string //背景色
	TitleColor          string //标题文字色
	CardTitleFontColor  string //卡片标题文字色
	CardTitleBgColor    string //卡片标题背景色
	ViewMoreFontColor   string //查看更多文字色
	ViewMoreBgColor     string //查看更多背景色
	Ps                  int64
	// 老版本所需参数
	PageID int64
}
