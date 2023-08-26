package currency

import (
	xtime "go-common/library/time"
	"go-gateway/app/web-svr/activity/interface/model/like"
)

const (
	BusinessAct     = 1
	TenUnlockAmount = 1000000
	TenLuckNum      = "090626"
	MikuAmount1     = 8000000
	MikuAmount2     = 16000000
	MikuAmount3     = 25000000
	MikuAmount4     = 39000000
	HasLock         = 1
	HasAward        = 1
	MikuGrant       = 0
	MikuProps       = 1
	MikuHead        = 2
	MikuStep1       = 0
	MikuStep2       = 1
	MikuStep3       = 2
	MikuStep4       = 3
)

// UserCurrency .
type UserCurrency struct {
	*Currency
	Amount int64 `json:"amount"`
}

// AllCurrency .
type AllCurrency struct {
	Amount    int64 `json:"amount"`
	ForeignID int64 `json:"foreign_id"`
}

// CurCurrencyReply .
type CurCurrencyReply struct {
	Amount    int64  `json:"amount"`
	AllAmount int64  `json:"all_amount"`
	Password  string `json:"password"`
	Image     string `json:"image"`
	H5Image   string `json:"h5_image"`
	HasCoupon int    `json:"has_coupon"`
}

// MikuState .
type MikuState struct {
	ID      int   `json:"id"`
	HasLock int64 `json:"has_lock"`
	Award   int64 `json:"award"`
}

// MikuAward .
type MikuAward struct {
	ID    int   `json:"id"`
	Award int64 `json:"award"`
}

// SingleAward .
type SingleAward struct {
	ID    int   `json:"id"`
	Award int64 `json:"award"`
}

type SingleAwardRes struct {
	List       []*SingleAward `json:"list"`
	UserAmount int64          `json:"user_amount"`
}

type SingleRank struct {
	StudyRank   []*like.MissionFriends `json:"study_rank"`
	TeacherRank []*like.MissionFriends `json:"teacher_rank"`
}

// MikuReply .
type MikuReply struct {
	List       []*MikuState `json:"list"`
	UserAmount int64        `json:"user_amount"`
	CurrAmount int64        `json:"curr_amount"`
}

// LiveHeadMsg .
type LiveHeadMsg struct {
	Uids    []int64        `json:"uids"`
	MsgID   string         `json:"msg_id"`
	Source  int64          `json:"source"`
	Rewards []*HeadRewards `json:"rewards"`
}

// HeadRewards .
type HeadRewards struct {
	RewardID   int64          `json:"reward_id"`
	ExpireTime xtime.Time     `json:"expire_time"`
	Type       int64          `json:"type"`
	ExtarData  *HeadExtraData `json:"extra_data"`
}

// HeadExtarData .
type HeadExtraData struct {
	Score   int64 `json:"score"`
	Upgrade int64 `json:"upgrade"`
}

// LivePropsMsg .
type LivePropsMsg struct {
	Uids    []int64         `json:"uids"`
	MsgID   string          `json:"msg_id"`
	Source  int64           `json:"source"`
	Rewards []*PropsRewards `json:"rewards"`
}

// PropsRewards .
type PropsRewards struct {
	RewardID   int64           `json:"reward_id"`
	ExpireTime xtime.Time      `json:"expire_time"`
	Type       int64           `json:"type"`
	Num        int64           `json:"num"`
	ExtraData  *PropsExtraData `json:"extra_data"`
}

// PropsExtraData
type PropsExtraData struct {
	MsgID  string `json:"msg_id"`
	Source int64  `json:"source"`
}

// CertificateMsg
type CertificateMsg struct {
	List []*Certificate `json:"list"`
}

type Certificate struct {
	Data *CertificateData `json:"data"`
}

type CertificateData struct {
	First  int64  `json:"first"`
	TopTen string `json:"top_ten"`
}
