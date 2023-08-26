package config

import (
	"go-gateway/app/app-svr/native-act/interface/kernel"
)

type Click struct {
	BaseCfgManager

	BgImage     *SizeImage //背景图
	Unlock      *Unlock    //解锁
	PressSave   bool       //是否开启长按保存
	Items       []*ClickItem
	UpRsvIdsReq kernel.RequestID
}

type ClickItem struct {
	AreaId           int64 //区域id
	Area             *Area
	Type             int64  //区域类型
	Url              string //跳转链接
	IosUrl           string //ios端链接
	AndroidUrl       string //android端链接
	Id               int64  //操作的对象id
	Image            string //图片
	DoneImage        string //已完成态图片
	UndoneImage      string //未完成态图片
	DisableImage     string //不可操作图片
	MsgBoxTip        string //弹框的提示
	GroupId          int64  //节点组id
	NodeId           int64  //节点id
	DisplayType      string //展示数值类型
	FontSize         int64  //字号
	FontType         string //字体
	FontColor        string //字体颜色
	ProgSource       int64  //非实时进度条-数据来源
	StatsDimension   int64  //非实时进度条-活动报名量-统计维度
	Activity         string //非实时进度条-任务统计-活动名
	Counter          string //非实时进度条-任务统计-counter名
	StatsPeriod      string //非实时进度条-任务统计-统计周期
	LotteryId        string //非实时进度条-抽奖数量-抽奖ID
	PlatCounterReqID kernel.RequestID
	PlatTotalReqID   kernel.RequestID
	Images           []*SizeImage //浮层图片
	Style            string       //浮层样式
	ShareImage       *SizeImage   //长按分享的图片
	ImageTitle       string       //图片样式-图片
	Title            string       //标题
	TitleColor       string       //标题颜色
	TopColor         string       //顶栏颜色
	SyncHover        bool         //是否与自定义悬浮按钮互通
	Unlock           *Unlock      //解锁
}

type Unlock struct {
	UnlockCondition int64 //解锁条件
	UnlockTime      int64 //解锁时间
	Sid             int64 //预约数据源id
	GroupId         int64 //节点组id
	NodeId          int64 //节点id
}
