package s10

import (
	"encoding/json"

	xtime "go-common/library/time"
)

type Message struct {
	Action string          `json:"action"`
	Table  string          `json:"table"`
	New    json.RawMessage `json:"new"`
	Old    json.RawMessage `json:"old"`
}

type UserCostRecord struct {
	Cost  int32 `json:"cost"`
	State int32 `json:"state"`
	Mid   int64 `json:"mid"`
	Gid   int32 `json:"gid"`
}

type UserLotteryRecord struct {
	Robin int32  `json:"robin"`
	State int32  `json:"state"`
	Mid   int64  `json:"mid"`
	Gid   int32  `json:"gid"`
	Extra string `json:"extra"`
}

type Goods struct {
	Stock int
	State int32
}

type UserInfo struct {
	Mid   int64  `json:"mid"`
	Level int32  `json:"level"`
	Name  string `json:"name"`
}

type CacheExpire struct {
	SignedExpire         xtime.Duration
	TaskProgressExpire   xtime.Duration
	RestPointExpire      xtime.Duration
	CoinExpire           xtime.Duration
	PointExpire          xtime.Duration
	LotteryExpire        xtime.Duration
	ExchangeExpire       xtime.Duration
	RestCountGoodsExpire xtime.Duration
	PointDetailExpire    xtime.Duration
}

type FreeFlow struct {
	Message string `json:"message"`
	Type    int32  `json:"type"`
	Source  int32  `json:"source"` // 0-联调；1-移动
}

type CostRecord struct {
	Gid   int32      `json:"gid"`
	Cost  int32      `json:"cost"`
	Ctime xtime.Time `json:"ctime"`
	Name  string     `json:"name"`
}

type MatchUser struct {
	IsRecieve bool   `json:"is_recieve"`
	IsLottery bool   `json:"is_lottery"`
	Lucky     *Lucky `json:"lucky"`
}

type Lucky struct {
	Gid   int32  `json:"gid"`
	Type  int32  `json:"type"`
	State int32  `json:"state"`
	Name  string `json:"name"`
	Extra string `json:"extra"`
}

type S10General struct {
	Switch bool
}
