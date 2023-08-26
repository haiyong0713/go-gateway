package bws

// GameRes .
const (
	GameResWin              = 1
	GameResFail             = 2
	AllType                 = 0
	DpType                  = 1
	GameType                = 2 //游戏点
	ClockinType             = 3 //打卡点
	EggType                 = 4
	HideClockinType         = 5 //隐藏打卡点
	ChargeType              = 6 //充能点
	SignType                = 7 //签到点
	SpecialClockinType      = 8 //特殊打卡点
	AchieveDpType           = 1
	AchieveGameType         = 2
	AchieveClockinType      = 3
	AchieveHideType         = 4
	AchieveSignType         = 5
	AchieveChargeType       = 6
	AchieveBindType         = 7
	AchieveVipCard          = 8
	AchieveIncrPointType    = 9
	AchieveOther            = 10
	AchieveSpeClockinType   = 11 // 特殊隐藏打卡点成就类型
	ExtraGameFirstFail      = 1
	ExtraGameFirstSuee      = 2
	ExtraGameContinueSuee   = 3
	ExtraGameContinueFail   = 4
	ExtraGameContinuePlay   = 5
	ExtraChargeCnt          = 6
	ExtraChargeHp           = 7
	ExtraBindNew            = 8
	ExtraBindOld            = 9
	ExtraBindVip            = 10
	ExtraGameSuee           = 11
	ExtraBindGuang          = 12 //云游万里
	ExtraBind2019           = 13 //神行千里
	Dp                      = "dp"
	Game                    = "game"
	Clockin                 = "clockin"
	Egg                     = "egg"
	HideClockin             = "hide_clockin"
	Recharge                = "recharge"
	Sign                    = "sign"
	SpecialClockin          = "special_clockin"
	RechargeUnlocked        = 1
	RechargeNotLock         = 0
	RechargeHalf            = 1
	RechargeAll             = 2
	AwardStateInit          = "init"
	AwardStatePending       = "pending"
	AwardStateFinish        = "finish"
	AwardStockNoLimit       = -1
	LotteryTimesStateInit   = "init"
	LotteryTimesStateFinish = "finish"
	VoteStateInit           = "init"
	VoteStateFinish         = "finish"
	TaskCateMain            = "main"
	TaskCateOther           = "other"
	TaskCateCatch           = "catch"
	TaskUserFinish          = 1
	TaskHasAward            = 1
	AwardCateLotteryTicket  = "lottery_ticket"
	ReasonPlayGame          = "play_game"
	ReasonVipAddHeart       = "vip_add"
	ReasonInternalAddHeart  = "internal_add"
	ReasonAddHeart          = "add_heart"
	ReasonLockHeart         = "lock_heart"
	ReasonUps               = "catch_ups"
	UpsStarOneMin           = "star_1_min"
	UpsStarOneMax           = "star_1_max"
	UpsStarTwoMin           = "star_2_min"
	UpsStarTwoMax           = "star_2_max"
	UpsStarThreeMin         = "star_3_min"
	UpsStarThreeMax         = "star_3_max"
	StarMapUps              = 0
)

var (
	PointAchieve = map[int64]int64{
		DpType:             AchieveDpType,
		GameType:           AchieveGameType,
		ClockinType:        AchieveClockinType,
		HideClockinType:    AchieveHideType,
		SignType:           AchieveSignType,
		ChargeType:         AchieveChargeType,
		SpecialClockinType: AchieveSpeClockinType,
	}
)

// ParamPoints points param
type ParamPoints struct {
	Bid int64 `form:"bid" validate:"required"`
	Tp  int64 `form:"tp"`
}

// ParamID point or achievements id param
type ParamID struct {
	Bid int64  `form:"bid" validate:"required"`
	ID  int64  `form:"id"`
	Day string `form:"day"`
}

// ParamAward point or achievements id param
type ParamAward struct {
	Bid int64  `form:"bid" validate:"required"`
	Aid int64  `form:"aid" validate:"required"`
	Key string `form:"key"`
	Mid int64  `form:"mid"`
}

// ParamBinding binding param
type ParamBinding struct {
	Bid int64  `form:"bid" validate:"required"`
	Key string `form:"key" validate:"required"`
}

// ParamUnlock .
type ParamUnlock struct {
	Bid        int64  `form:"bid" validate:"required"`
	Pid        int64  `form:"pid" validate:"required"`
	Key        string `form:"key"`
	Mid        int64  `form:"mid"`
	GameResult int    `form:"game_result"`
	Recharge   int    `form:"recharge"`
}

type ParamUnlock20 struct {
	Bid int64  `form:"bid" validate:"required"`
	Pid int64  `form:"pid" validate:"required"`
	Key string `form:"key"`
	Mid int64  `form:"mid"`
}

// ParamRechargeAward  param .
type ParamRechargeAward struct {
	Bid int64 `form:"bid" validate:"required"`
}

// ParamSign .
type ParamSign struct {
	Bid int64  `form:"bid" validate:"required"`
	Key string `form:"key"`
	Mid int64  `form:"mid"`
}

// SignReply .
type SignReply struct {
	Points []*SinglePoints `json:"points"`
}

// SinglePoints .
type SinglePoints struct {
	*Point
	Sign     *SignInfoReply   `json:"sign,omitempty"`
	Recharge []*RechargeAward `json:"recharge,omitempty"`
}

// SignInfoReply .
type SignInfoReply struct {
	ID            int64 `json:"id"`
	SurplusPoints int64 `json:"surplus_points"`
	SignPoints    int64 `json:"sign_points"`
	State         int32 `json:"state"`
	Stime         int64 `json:"stime"`
	Etime         int64 `json:"etime"`
	Points        int64 `json:"points"`
}

type VoteLog struct {
	ID        int64  `json:"id"`
	UserToken string `json:"user_token"`
	PointID   int64  `json:"point_id"`
	Result    int64  `json:"result"`
	Ctime     int64  `json:"ctime"`
	Mtime     int64  `json:"mtime"`
}

// CatchUpper .
type CatchUpper struct {
	Bid    int64  `form:"bid" validate:"required"`
	Key    string `form:"key"`
	UpKeys string `form:"up_keys"`
	Pn     int    `form:"pn"`
	Ps     int    `form:"ps"`
}

// LogReason reason
type LogReason struct {
	Reason string `json:"reason"`
	Params string `json:"params"`
}

// RankRes ...
type RankRes struct {
	List []*Account `json:"list"`
}

// MidScore ...
type MidScore struct {
	Star         int64 `json:"star"`
	LastStarTime int64 `json:"last_star_time"`
}

// Account ...
type Account struct {
	Mid          int64  `json:"mid"`
	Name         string `json:"name"`
	Face         string `json:"face"`
	Sign         string `json:"sign"`
	Sex          string `json:"sex"`
	LastStarTime int64  `json:"last_star_time"`
	Star         int64  `json:"star"`
}

//go:generate kratos t protoc --grpc bws.proto
