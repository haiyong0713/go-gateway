package lottery

import (
	"encoding/json"
	"go-common/library/ecode"
	"go-common/library/log"
	xtime "go-common/library/time"
	"math"
)

const (
	GiftLeastMarkY          = 1
	GiftLeastMarkN          = 0
	GiftShow                = 1
	GiftNotShow             = 0
	TimesTypeBase           = 1
	TimesTypePrice          = 2
	TimesAddTypeAll         = 0
	TimesAddTypeDay         = 1
	GiftUploadWait          = 0
	GiftUploadSuccess       = 1
	GiftUploadFailed        = 2
	AddActionTypeVIP        = 6
	AddActionTypeCustom     = 8
	AddActionTypeLike       = 11
	AddActionTypeCoin       = 12
	AddActionTypeAdditional = 13
	AddActionTypeTaskPoint  = 15
	GiftTypeVIP             = 2
	GiftTypeGrant           = 3
	GiftTypeSend            = 4
	GiftTypeCoin            = 5
	GiftTypeCoupon          = 6
	GiftTypeOGV             = 8
	GiftTypeVipBuy          = 9
	GiftTypeMoney           = 10
	GiftTypeAward           = 11
	UploadNon               = 0
	UploadStart             = 1
	UploadSuccess           = 2
	UploadFailed            = 3
	StateYes                = 0
	StateNo                 = 1
	EffectY                 = 1
	EffectN                 = 0
	FsIPOn                  = 1
	FsIPOff                 = 0
	InitLevel               = 1
	AddActionTypeOther      = 7
	AddActionTypeOGV        = 9
	MaxMidLen               = 1000000
	AddActionTypeAct        = 14

	// ProbabilityBit 3bit
	ProbabilityBit = 3
	// ExtraLengthMax 额外参数
	ExtraLengthMax = 1024
)

// CoinLikeInfo ...
type CoinLikeInfo struct {
	Activity string `form:"activity" validate:"required"`
	Counter  string `form:"counter" validate:"required"`
	Count    int64  `form:"count" validate:"required"`
}

// ConsumeInfo ...
type ConsumeInfo struct {
	Consume int64 `json:"consume"`
	Send    int   `json:"send"`
}

// ActivityInfo ...
type ActivityInfo struct {
	Sid     int64 `json:"sid"`
	GroupID int64 `json:"group_id"`
}

// LotteryInfo activity lottery base information
type AddParam struct {
	Name  string     `form:"name" validate:"required"`
	Stime xtime.Time `form:"stime" validate:"required"`
	Etime xtime.Time `form:"etime" validate:"required" `
	Type  int        `form:"type" gorm:"type"`
}

// UsedParam activity lottery used information
type UsedParam struct {
	SID string `form:"sid" validate:"required"`
	MID int64  `form:"mid" validate:"required"`
}

// ListParam
type ListParam struct {
	State   int    `form:"state"`
	Keyword string `form:"keyword"`
	Rank    string `form:"rank"`
	Pn      int    `form:"pn" default:"1"`
	Ps      int    `form:"ps" default:"20"`
}

// ListRsp
type ListRsp struct {
	List []*LotInfo `json:"list"`
	Page *Page      `json:"page"`
}

// ListDraftRsp ...
type ListDraftRsp struct {
	List []*LotInfoDraft `json:"list"`
	Page *Page           `json:"page"`
}

// Page
type Page struct {
	Num   int `json:"num"`
	Size  int `json:"size"`
	Total int `json:"total"`
}

// RecordDetailRes ...
type RecordDetailRes struct {
	List []*RecordDetail `json:"list"`
}

// RecordDetail ...
type RecordDetail struct {
	ID       int64      `json:"id"`
	Mid      int64      `json:"mid"`
	IP       int64      `json:"ip"`
	Num      int        `json:"num"`
	GiftID   int64      `json:"gift_id"`
	GiftName string     `json:"gift_name"`
	Type     int        `json:"type"`
	Ctime    xtime.Time `json:"ctime"`
	CID      int64      `json:"cid"`
}

// LotDetailInfo ...
type LotDetailInfo struct {
	List         LotInfo          `json:"list"`
	Info         RuleInfo         `json:"info"`
	LotteryTimes *TimesConf       `json:"lottery_times"`
	PriceTimes   *TimesConf       `json:"price_times"`
	TimesConf    []*TimesConf     `json:"timesConf"`
	Gift         []*GiftInfo      `json:"gift"`
	MemberGroup  []*MemberGroupDB `json:"member_group"`
}

// LotDetailInfoDraft
type LotDetailInfoDraft struct {
	List         LotInfoDraft     `json:"list"`
	Info         RuleInfo         `json:"info"`
	LotteryTimes *TimesConf       `json:"lottery_times"`
	PriceTimes   *TimesConf       `json:"price_times"`
	TimesConf    []*TimesConf     `json:"timesConf"`
	Gift         []*GiftInfo      `json:"gift"`
	MemberGroup  []*MemberGroupDB `json:"member_group"`
	CanAudit     bool             `json:"can_audit"`
}

// RuleInfo
type RuleInfo struct {
	ID           int64  `json:"id"`
	Sid          string `json:"sid"`
	Level        int    `json:"level"`
	RegtimeStime int    `json:"regtime_stime"`
	RegtimeEtime int    `json:"regtime_etime"`
	VipCheck     int    `json:"vip_check"`
	AccountCheck int    `json:"account_check"`
	Coin         int    `json:"coin"`
	FsIP         int    `json:"fs_ip"`
	GiftRate     int    `json:"gift_rate"`
	HighType     int    `json:"high_type"`
	HighRate     int    `json:"high_rate"`
	SenderMid    int64  `json:"sender_mid"`
	FigureScore  int64  `json:"figure_score"`
	SpyScore     int64  `json:"spy_score"`
	State        int    `json:"state"`
	ActivityLink string `json:"activity_link"`
}

func (t RuleInfo) TableName() string {
	return "act_lottery_info"
}

type AddTimesLog struct {
	ID       int64      `json:"id"`
	Author   string     `json:"author"`
	Sid      string     `json:"sid"`
	Cid      int64      `json:"cid"`
	State    int        `json:"state"`
	FileName string     `json:"filename"`
	Ctime    xtime.Time `json:"ctime"`
	Mtime    xtime.Time `json:"mtime"`
}

// TimesConf
type TimesConf struct {
	ID      int64      `json:"id"`
	Sid     string     `json:"sid"`
	Type    int        `json:"type"`
	Info    string     `json:"info"`
	Times   int        `json:"times"`
	AddType int        `json:"add_type"`
	Most    int        `json:"most"`
	Ctime   xtime.Time `json:"ctime"`
	Mtime   xtime.Time `json:"mtime"`
	State   int        `json:"state"`
}

// GiftInfo ...
type GiftInfo struct {
	ID             int64      `json:"id"`
	Sid            string     `json:"sid"`
	Name           string     `json:"name"`
	Num            int        `json:"num"`
	Type           int        `json:"type"`
	Source         string     `json:"source"`
	ImgURL         string     `json:"img_url"`
	IsShow         int        `json:"isshow"`
	LeastMark      int        `json:"least_mark"`
	Effect         int        `json:"effect"`
	TimeLimit      xtime.Time `json:"time_limit"`
	MessageTitle   string     `json:"message_title"`
	MessageContent string     `json:"message_content"`
	Upload         int        `json:"upload"`
	Ctime          xtime.Time `json:"ctime"`
	Mtime          xtime.Time `json:"mtime"`
	State          string     `json:"state"`
	DBNum          int        `json:"db_num"`
	RedisNum       int        `json:"redis_num"`
	SendNum        int        `json:"send_num"`
	Params         string     `json:"params"`
	MemberGroup    string     `json:"member_group"`
	DayNum         string     `json:"day_num"`
	ProbabilityF   float64    `json:"probability"`
	ProbabilityI   int        `json:"_"`
	Extra          string     `json:"extra"`
}

// EditParam
type EditParam struct {
	ID           int64      `form:"id" validate:"required"`
	SID          string     `form:"sid" validate:"required"`
	Name         string     `form:"name" validate:"required"`
	Stime        xtime.Time `form:"stime"`
	Etime        xtime.Time `form:"etime"`
	LotTimes     string     `form:"lottery_times"`
	PriceTimes   string     `form:"price_times"`
	Level        int        `form:"level"`
	RegTimeSTime int        `form:"regtime_stime"`
	RegTimeETime int        `form:"regtime_etime"`
	VipCheck     int        `form:"vip_check"`
	AccountCheck int        `form:"account_check"`
	FsIP         int        `form:"fs_ip"`
	CoinCheck    int        `form:"coin"`
	ActionAdd    string     `form:"action_add"`
	Rate         int        `form:"gift_rate"`
	HighRate     int        `form:"high_rate"`
	HighType     int        `form:"high_type"`
	SenderMid    int64      `form:"sender_mid"`
	IsInternal   int        `form:"is_internal"`
	State        int        `form:"state"`
	ActivityLink string     `form:"activity_link"`
	SpyScore     int64      `form:"spy_score"`
	FigureScore  int64      `form:"figure_score"`
}

// MemberGroupDB  EditParam.MemberGroup json Unmarshal
type MemberGroupDB struct {
	ID    int64      `json:"id"`
	SID   string     `json:"sid"`
	Name  string     `json:"name"`
	Group string     `json:"group"`
	State int        `json:"state"`
	Ctime xtime.Time `json:"ctime"`
	Mtime xtime.Time `json:"mtime"`
}

// MemberGroup ...
type MemberGroup struct {
	ID    int64       `json:"id"`
	SID   string      `json:"sid"`
	Name  string      `json:"name"`
	Group interface{} `json:"group"`
	State int         `json:"state"`
}

// ActionAdd  EditParam.ActionAdd json Unmarshal
type ActionAdd struct {
	ID     int64 `json:"id"`
	Type   int   `json:"type"`
	Info   int   `json:"info"`
	Times  int   `json:"times"`
	State  int   `json:"state"`
	Most   int   `json:"most"`
	Status int   `json:"status"`
}

// BaseTimes EditParam lottery_times or price_times Unmarshal
type BaseTimes struct {
	ID      int64 `json:"id"`
	AddType int   `json:"add_type"`
	Times   int   `json:"times"`
}

// AddTimes EditParam action_add Unmarshal
type AddTimes struct {
	ID      int64  `json:"id"`
	Type    int    `json:"type"`
	Info    string `json:"info"`
	Times   int    `json:"times"`
	State   int    `json:"state"`
	AddType int    `json:"add_type"`
	Most    int    `json:"most"`
	Status  int    `json:"status"`
}

// GiftAddParam gift add request params
type GiftAddParam struct {
	SID         string     `form:"sid" validate:"required"`
	Name        string     `form:"name" validate:"required"`
	Num         int        `form:"num" validate:"required"`
	Type        int        `form:"type" validate:"required"`
	ImgURL      string     `form:"img_url"`
	Source      string     `form:"source"`
	TimeLimit   xtime.Time `form:"time_limit"`
	MsgTitle    string     `form:"message_title"`
	MsgContent  string     `form:"message_content"`
	Params      string     `form:"params"`
	MemberGroup string     `form:"member_group"`
	DayNum      string     `form:"day_num"`
	Probability float64    `form:"probability"`
	Extra       string     `form:"extra" default:"{}"`
}

// AuditParam ...
type AuditParam struct {
	SID          string `form:"sid" validate:"required"`
	RejectReason string `form:"reject_reason"`
	State        int    `form:"state"`
}

// GiftEditParam gift edit request params
type GiftEditParam struct {
	ID          int64      `form:"id" validate:"required"`
	SID         string     `form:"sid" validate:"required"`
	Name        string     `form:"name" validate:"required"`
	Num         int        `form:"num" validate:"required"`
	Type        int        `form:"type" validate:"required"`
	Source      string     `form:"source"`
	IsShow      int        `form:"isshow"`
	LeastMark   int        `form:"least_mark"`
	Effect      int        `form:"effect"`
	TimeLimit   xtime.Time `form:"time_limit"`
	MsgTitle    string     `form:"message_title"`
	MsgContent  string     `form:"message_content"`
	ImgURL      string     `form:"img_url"`
	Params      string     `form:"params"`
	MemberGroup string     `form:"member_group"`
	DayNum      string     `form:"day_num"`
	Probability float64    `form:"probability"`
	Extra       string     `form:"extra" default:"{}"`
}

// MemberGroupEditParam membergroup edit request params
type MemberGroupEditParam struct {
	ID    int64  `form:"id"`
	SID   string `form:"sid" validate:"required"`
	Name  string `form:"name" validate:"required"`
	State int    `form:"state"`
	Group string `form:"group"`
}

// GiftListParam gift list request params
type GiftListParam struct {
	SID   string `form:"sid" validate:"required"`
	State int    `form:"state"`
	Type  int    `form:"type"`
	Rank  string `form:"rank" default:"ctime"`
	Pn    int    `form:"pn" default:"1"`
	Ps    int    `form:"ps" default:"20"`
}

// BatchAddTimesParams gift list request params
type BatchAddTimesParams struct {
	SID string `form:"sid" `
	Pn  int    `form:"pn" default:"1"`
	Ps  int    `form:"ps" default:"20"`
}

// MemberGroupListParam membergroup list request params
type MemberGroupListParam struct {
	SID   string `form:"sid" validate:"required"`
	State int    `form:"state" default:"1"`
	Rank  string `form:"rank" default:"ctime"`
	Pn    int    `form:"pn" default:"1"`
	Ps    int    `form:"ps" default:"20"`
}

// GiftList is gift/list response
type GiftList struct {
	List []*GiftInfo `json:"list"`
	Page Page        `json:"page"`
}

// AddTimesBatchLogList is gift/list response
type AddTimesBatchLogList struct {
	List []*AddTimesLog `json:"list"`
	Page Page           `json:"page"`
}

type LotteryAddTimes struct {
	ID    int64      `json:"id"`
	Mid   int64      `json:"mid"`
	Type  int        `json:"type"`
	Num   int        `json:"num"`
	CID   int64      `json:"cid"`
	Ctime xtime.Time `json:"ctime"`
}

// LotteryAddTimesReply ...
type LotteryAddTimesReply struct {
	List []*LotteryAddTimes `json:"list"`
	Page Page               `json:"page"`
}

// MemberGroupListReply is memberGroup/list response
type MemberGroupListReply struct {
	List []*MemberGroupDB `json:"list"`
	Page Page             `json:"page"`
}

// GiftWinListParam is gift/win params
type GiftWinListParam struct {
	SID string `form:"sid" validate:"required"`
	ID  int64  `form:"id" validate:"required"`
	Pn  int    `form:"pn" validate:"required" default:"1"`
	Ps  int    `form:"ps" validate:"required" default:"20"`
}

// GiftWinList is gift/win response
type GiftWinList struct {
	List []*GiftWinInfo `json:"list"`
	Page Page           `json:"page"`
}

// GiftWinInfo gift win list information
type GiftWinInfo struct {
	ID         int64      `json:"id"`
	Mid        int        `json:"mid"`
	GiftId     int64      `json:"gift_id"`
	CDKey      string     `json:"cdkey"`
	CTime      xtime.Time `json:"ctime"`
	MTime      xtime.Time `json:"mtime"`
	Addr       Address    `json:"address"`
	GiftAddrID int64      `json:"gift_addr_id"`
}

// Address address info
type Address struct {
	Status  int    `json:"status"`
	ProvID  int    `json:"prov_id"`
	CityID  int    `json:"city_id"`
	AreaID  int    `json:"area_id"`
	Name    string `json:"name"`
	Phone   string `json:"phone"`
	Addr    string `json:"addr"`
	ZipCode string `json:"zip_code"`
	Prov    string `json:"prov"`
	City    string `json:"city"`
	Area    string `json:"area"`
}

// UploadInfo giftUpload sync information
type UploadInfo struct {
	Status int
	Update bool
}

// GIftTask
type GiftTask struct {
	ID        int64
	SID       string
	TimeLimit xtime.Time
	Type      int
}

// GetIntProbability get int probability
func (g *GiftAddParam) GetIntProbability() int {
	var multiplier float64
	multiplier = math.Pow(10, ProbabilityBit)
	return int(g.Probability * multiplier)
}

// GetIntProbability get int probability
func (g *GiftEditParam) GetIntProbability() int {
	var multiplier float64
	multiplier = math.Pow(10, ProbabilityBit)
	return int(g.Probability * multiplier)
}

// GetDayStore ...
func (g *GiftAddParam) GetDayStore() error {
	params := &giftDayStore{}
	if err := json.Unmarshal([]byte(g.DayNum), params); err != nil {
		log.Error("json.Unmarshal(%s) error(%v)", string(g.DayNum), err)
		return ecode.Error(ecode.RequestErr, "请配置单日中奖上限，若无上限则设置为0")
	}
	dayNum, _ := json.Marshal(params)
	g.DayNum = string(dayNum)
	return nil
}

type giftDayStore map[string]int

// GetDayStore get day store
func (g *GiftEditParam) GetDayStore() error {
	params := &giftDayStore{}
	if err := json.Unmarshal([]byte(g.DayNum), params); err != nil {
		log.Error("json.Unmarshal(%s) error(%v)", string(g.DayNum), err)
		return ecode.Error(ecode.RequestErr, "请配置单日中奖上限，若无上限则设置为0")
	}
	dayNum, _ := json.Marshal(params)
	g.DayNum = string(dayNum)
	return nil
}

// GetFloatProbability get int probability
func (g *GiftInfo) GetFloatProbability() float64 {
	var multiplier float64
	multiplier = math.Pow(10, ProbabilityBit)
	return float64(g.ProbabilityI) / multiplier
}
