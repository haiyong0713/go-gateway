package config

import (
	"go-gateway/app/app-svr/native-act/interface/kernel"
)

type Timeline struct {
	BaseCfgManager

	ImageTitle       string //图片标题
	TextTitle        string //文字标题
	NodeType         int64  //时间轴节点类型
	TimePrecision    int64  //时间节点-精度
	BgColor          string //背景色
	CardBgColor      string //卡片背景色
	TimelineColor    string //时间轴色
	ViewMoreType     int64  //查看更多方式
	ViewMoreText     string //查看更多文案
	SupernatantTitle string //浮窗标题
	ShowNum          int64  //外显事件个数
	Ps               int64
	MixExtsReqID     kernel.RequestID
}
