package config

import (
	"go-gateway/app/app-svr/native-act/interface/kernel"
)

type Ogv struct {
	BaseCfgManager

	OgvCommon
	MixExtsReqID kernel.RequestID
}

type OgvCommon struct {
	ImageTitle       string
	TextTitle        string
	Color            *OgvColor
	IsThreeCard      bool   //是否是三列卡
	Ps               int64  //当前页面展示数量
	ViewMoreText     string //查看更多文案
	SupernatantTitle string //更多内容浮层标题
	DisplayPayBadge  bool   //是否展示付费角标
	DisplayScore     bool   //是否展示评分
	DisplayRcmd      bool   //是否展示推荐语
	DisplaySubtitle  bool   //是否展示副标题
	DisplayMore      bool   //是否展示查看更多
}

type OgvColor struct {
	BgColor           string //组件背景色
	CardBgColor       string //卡片背景色
	ViewMoreFontColor string //查看更多文字色
	ViewMoreBgColor   string //查看更多按钮色
	RcmdFontColor     string //推荐语文字色
	TitleColor        string //剧集标题色
	SubtitleFontColor string //副标题文字色
}
