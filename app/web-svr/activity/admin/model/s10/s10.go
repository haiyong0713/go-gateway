package s10

import xtime "go-common/library/time"

const (
	S10LotterySentinels = 255
)

type Goods struct {
	Stock int
	State int32
	Type  int32
	Gname string
	Gid   int32
	Score int32
}

type UserInfo struct {
	Mid   int64  `json:"mid"`
	Level int32  `json:"level"`
	Name  string `json:"name"`
}

type UserCostRecord struct {
	ID       int32      `json:"id"`
	Gid      int32      `json:"gid"`
	Ack      int32      `json:"ack"`
	Cost     int32      `json:"cost"`
	State    int32      `json:"state"`
	Name     string     `json:"name"`
	UniqueID string     `json:"unique_id"`
	Ctime    xtime.Time `json:"ctime"`
}

type SuperLotteryUserInfo struct {
	Gid   int32  `json:"gid"`
	Robin int32  `json:"robin"`
	Mid   int64  `json:"mid"`
	Name  string `json:"name"`
	Gname string `json:"gname"`
}

type UserGiftRecord struct {
	ID       int32      `json:"id"`
	Gid      int32      `json:"gid"`
	Robin    int32      `json:"robin"`
	Ack      int32      `json:"ack"`
	State    int32      `json:"state"`
	Name     string     `json:"name"`
	UniqueID string     `json:"unique_id"`
	Ctime    xtime.Time `json:"ctime"`
}

type RealGiftRecord struct {
	ID       int32  `json:"id"`
	Gid      int32  `json:"gid"`
	Robin    int32  `json:"robin"`
	State    int32  `json:"state"`
	Mid      int64  `json:"mid"`
	UserName string `json:"user_name"`
	Number   string `json:"number"`
	Addr     string `json:"addr"`
}

type CostRecord struct {
	Gid   int32      `json:"gid"`
	Cost  int32      `json:"cost"`
	Ctime xtime.Time `json:"ctime"`
	Name  string     `json:"name"`
}

type MatchUser struct {
	IsRecieve  bool   `json:"is_recieve"`
	IsLottery  bool   `json:"is_lottery"`
	ExpireTime int64  `json:"expire_time"`
	Lucky      *Lucky `json:"lucky"`
}

type Lucky struct {
	Gid     int32  `json:"gid"`
	Type    int32  `json:"type"`
	State   int32  `json:"state"`
	Name    string `json:"name"`
	Extra   string `json:"extra"`
	Desc    string `json:"desc"`
	Figture string `json:"figture"`
}

type Gift struct {
	Gid   int32  `json:"gid"`
	State int32  `json:"state"`
	Act   int32  `json:"act"`
	Name  string `json:"name"`
}
