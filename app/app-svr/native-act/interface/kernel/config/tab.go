package config

import (
	"go-gateway/app/app-svr/native-act/interface/internal/model"
)

type Tab struct {
	BaseCfgManager

	Style int64      //样式
	Items []*TabItem //tab选项
	// 纯色
	BgColor             string //背景色
	SelectedFontColor   string //选中色
	UnselectedFontColor string //未选中色
	DisplayUnfoldButton bool   //展示展开收起按钮
	// 图片
	BgImage SizeImage //tab栏背景图
}

type TabItem struct {
	PageID int64 //子页面id
	TabItemExt
}

type TabItemExt = model.TabExt
