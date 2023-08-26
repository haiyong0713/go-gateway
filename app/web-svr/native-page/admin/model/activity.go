package model

import (
	"time"

	natpagegrpc "go-gateway/app/web-svr/native-page/interface/api"
)

const (
	// 数据源活动类型
	ActTypeVote    = 4  //视频投票活动
	ActTypeCollect = 13 //视频收集活动
	ActTypeShoot   = 16 //拍摄活动
	// 活动作者
	ActAuthorUp = "up-sponsor"
)

var ActType2FromType = map[int64]int{
	ActTypeVote:    natpagegrpc.PageFromNewactVote,
	ActTypeCollect: natpagegrpc.PageFromNewactCollect,
	ActTypeShoot:   natpagegrpc.PageFromNewactShoot,
}

type AddActSubjectReq struct {
	Type   int64 //活动类型
	Stime  time.Time
	Etime  time.Time
	Author string //活动作者
	Name   string //活动名称
	Types  string //视频分区ids，逗号分隔
	Tags   string //tag标签
}
