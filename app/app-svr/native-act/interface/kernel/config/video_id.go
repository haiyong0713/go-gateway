package config

import (
	"go-gateway/app/app-svr/native-act/interface/kernel"
)

type VideoID struct {
	BaseCfgManager

	VideoCommon
	MixExtsReqID kernel.RequestID
}

type VideoCommon struct {
	ImageTitle         string
	TextTitle          string
	AutoPlay           bool   //是否开启自动播放
	HideTitle          bool   //是否隐藏标题
	DisplayViewMore    bool   //是否展示查看更多按钮
	BgColor            string //背景色
	TitleColor         string //标题文字色
	CardTitleFontColor string //卡片标题文字色
	SubpageTitle       string //二级页面标题
	ViewMoreFontColor  string //查看更多文字色
	ViewMoreBgColor    string //查看更多背景色
	Ps                 int64  //展示条数
	// 老版本所需参数
	PageID int64
}
