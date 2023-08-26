package s10

import (
	xtime "go-common/library/time"
)

const (
	S10Act                            = "S10Activity"
	S10ActSign                        = "sign"
	S10ActPred                        = "pred"
	S10LotterySentinels               = 255
	S10LimitTypeGoodsExhcange         = "exchange"
	S10LimitTypeBackToData            = "back"
	S10LimitBusinessPoints            = "points"
	S10LimitBusinessPointsDetail      = "points_detail"
	S10LimitBusinessUserLotteryGoods  = "lottery"
	S10LimitBusinessUserReceiveGoods  = "receive"
	S10LimitBusinessUserExchangeGoods = "exchange"
	S10LimitBusinessUserFlow          = "flow"
	S10PointsCost                     = 1
	S10PointsTotal                    = 0
)

type S10MainDtb struct {
	Act       string `json:"act"`       // 动作
	Timestamp int64  `json:"timestamp"` // 时间戳
	Mid       int64  `json:"mid"`       // 用户uid
}

type FreeFlowPub struct {
	Message string `json:"message"`
	Type    int32  `json:"type"`
	Source  int32  `json:"source"` // 0-联调；1-移动
}

type Task struct {
	Caption string
	Total   int32
}

type ActTask struct {
	Act  string
	Task []*Task
}

type Points struct {
	Total int32 `json:"total"`
	Rest  int32 `json:"rest"`
}

type CacheExpire struct {
	SignedExpire              xtime.Duration
	TaskProgressExpire        xtime.Duration
	RestPointExpire           xtime.Duration
	CoinExpire                xtime.Duration
	PointExpire               xtime.Duration
	LotteryExpire             xtime.Duration
	ExchangeExpire            xtime.Duration
	RoundExchangeExpire       xtime.Duration
	RestCountGoodsExpire      xtime.Duration
	RoundRestCountGoodsExpire xtime.Duration
	PointDetailExpire         xtime.Duration
	UserFlowExpire            xtime.Duration
}

type Progress struct {
	Completed int32 `json:"completed"`
	MaxTimes  int32 `json:"max_times"`
}

type TaskProgress struct {
	UniqID    string `json:"uniq_id"`
	Status    bool   `json:"status"`
	*Progress `json:"progress"`
}

type Good struct {
	Gid     int32  `json:"gid"`
	Type    int32  `json:"type"`
	Point   int32  `json:"point,omitempty"`
	Res     int32  `json:"res,omitempty"`
	Name    string `json:"name"`
	Picture string `json:"picture"`
}

type CostRecord struct {
	Gid   int32      `json:"gid"`
	Cost  int32      `json:"cost"`
	Ctime xtime.Time `json:"ctime"`
	Name  string     `json:"name"`
}

type MatchCategories struct {
	CurrentTime  int64                `json:"current_time"`
	CurrentRobin int32                `json:"current_robin"` // 当前赛事阶段
	Ongoing      int32                `json:"ongoing"`       // 正在进行中的阶段
	IsLogin      bool                 `json:"is_login"`      // 用户是否登录
	IsDegrade    bool                 `json:"is_degrade"`    // 用户相关信息是否降级
	Lottery      map[int32]*MatchUser `json:"lottery"`
	Matches      []*MatchItem         `json:"matches"`
}

type MatchItem struct {
	IsLottery        bool `json:"is_lottery"`
	*Match           `json:"match"`
	*Other           `json:"other"`
	Bonuses          []*Bonus           `json:"bonuses"`
	UserLotteryInfos []*UserLotteryInfo `json:"user_lottery_infos"`
}

type UserLotteryInfo struct {
	Name  string `json:"name"`
	Bonus string `json:"bonus"`
}

type Other struct {
	MaxPoints int32  `json:"max_points"`
	Name      string `json:"name"`
	Figture   string `json:"figture"`
	Link      string `json:"link"`
}

type ActGoods struct {
	Currtime int64    `json:"currtime"`
	Bonuses  []*Bonus `json:"bonuses"`
	Other    []*Other `json:"other"`
}

type FreeFlow struct {
	CurrentTime     int64 `json:"current_time"`
	Swith           bool  `json:"switch"`
	Points          int32 `json:"points"`
	IsLogIn         bool  `json:"is_login"`
	Receive         int32 `json:"receive"` // 1-未领取；2-已领取
	Mobile          int32 `json:"mobile"`  // 0-是敬请期待 1-定点开抢 2-是可领取 3-当天无库存 4-活动结束
	Unicom          int32 `json:"unicom"`  // 0-是敬请期待 1-定点开抢 2-是可领取 3-当天无库存 4-活动结束
	MobileStartTime int64 `json:"mobile_start_time"`
	MobileEndTime   int64 `json:"mobile_end_time"`
	UnicomStartTime int64 `json:"unicom_start_time"`
	UnicomEndTime   int64 `json:"unicom_end_time"`
}

type Bonus struct {
	ID                 int32      `json:"gid"`
	Score              int32      `json:"score"`
	Robin              int32      `json:"-"`
	Rank               int32      `json:"-"`
	Send               int32      `json:"-"`
	Stock              int32      `json:"-"`
	Type               int32      `json:"type"`
	ExchangeTimes      int32      `json:"exchange_times"`
	RoundStock         int32      `json:"-"`
	RoundExchangeTimes int32      `json:"round_exchange_times"`
	RoundSend          int32      `json:"-"`
	CurrDate           int64      `json:"-"`
	Extra              string     `json:"-"`
	Name               string     `json:"name"`
	Figure             string     `json:"figure"`
	Start              xtime.Time `json:"start"`
	End                xtime.Time `json:"end"`
	IsInfinite         bool       `json:"-"`
	IsRoundInfinite    bool       `json:"-"`
	IsHaust            int32      `json:"is_haust"`
	IsRound            bool       `json:"is_round"`
	LeftTimes          int32      `json:"left_times"`
	Desc               string     `json:"-"`
}

type Match struct {
	Title         string `json:"title"`
	Robin         int32  `json:"robin"`
	Points        int32  `json:"points"`
	Start         int64  `json:"start"`
	End           int64  `json:"end"`
	Lottery       int64  `json:"lottery"`
	LotteryExpire int64  `json:"-"`
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

type MatchesConf struct {
	Matches        []*Match
	OtherActivity  []*Other
	DefaultBonuses []*Bonus
}

type S10TimePeriod struct {
	Points *TimePeriod
	Goods  *TimePeriod
}

type TimePeriod struct {
	Start int64 `json:"start"`
	End   int64 `json:"end"`
}

type RoundGoodStatic struct {
	Send  int32
	Stock int32
}

type UserTitle struct {
	Uids    []int64            `json:"uids"`
	MsgID   string             `json:"msg_id"`
	Source  int32              `json:"source"`
	Rewards []*UserTitleReward `json:"rewards"`
}

type UserTitleReward struct {
	RewardID   int32               `json:"reward_id"`
	ExpireTime int64               `json:"expire_time"`
	Type       int32               `json:"type"`
	ExtraData  *UserTitleExtraData `json:"extra_data"`
}

type UserTitleExtraData struct {
	Score   int32 `json:"score"`
	Upgrade int32 `json:"upgrade"`
}

type BulletExtraData struct {
	Type   string `json:"type"`
	Value  int64  `json:"value"`
	RoomID int32  `json:"roomid"`
}

type Bullet struct {
	Uids    []int64         `json:"uids"`
	MsgID   string          `json:"msg_id"`
	Source  int32           `json:"source"`
	Rewards []*BulletReward `json:"rewards"`
}

type BulletReward struct {
	RewardID   int32            `json:"reward_id"`
	ExpireTime int64            `json:"expire_time"`
	Type       int32            `json:"type"`
	ExtraData  *BulletExtraData `json:"extra_data"`
}
