package model

import (
	xtime "go-common/library/time"
)

type Task struct {
	Id           int64
	Uid          int64
	Goods        int64
	GoodsState   int64
	GoodsAddress int64
	Mtime        xtime.Time
}

type InviteItem struct {
	Id      int64  `json:"id"`
	Uid     int64  `json:"uid"`
	Mid     int64  `json:"mid"`
	Picture string `json:"picture"`
	Score   int64  `json:"score"`
}

type Invite struct {
	Id    int64
	Uid   int64
	Mid   int64
	Score int64
	Ctime xtime.Time
}

type InviteListItem struct {
	Score int64
	Ctime xtime.Time
}
