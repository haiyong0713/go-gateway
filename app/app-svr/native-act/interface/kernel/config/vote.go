package config

import (
	"go-gateway/app/app-svr/native-act/interface/kernel"
)

type Vote struct {
	BaseCfgManager

	BgImage         SizeImage     //背景图
	DisplayNum      bool          //是否展示得票数
	DoneButtonImage string        //完成态按钮
	VoteButtons     []*VoteButton //投票按钮
	VoteProgress    *VoteProgress //投票进度条
	VoteLeftNum     *VoteNum      //投票剩余数
	Sid             int64         //数据源id
	Gid             int64         //数据组id
	SourceType      string        //数据源类型
	VoteRankReqID   kernel.RequestID
}

type VoteButton struct {
	Area        Area
	UndoneImage string //未完成态按钮图片
}

type VoteProgress struct {
	Area         Area
	OptionColors []string //进度条颜色
	Style        string   //进度条样式
}

type VoteNum struct {
	Area Area
}
