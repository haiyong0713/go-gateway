package lottery

import (
	xtime "go-common/library/time"
	"go-gateway/app/web-svr/activity/ecode"
)

const (
	_ = iota
	// TimesBaseType 基础抽奖次数
	TimesBaseType
	// TimesWinType 最多获奖次数类型
	TimesWinType
	// TimesShareType 分享增加次数
	TimesShareType
	// TimesFollowType 关注增加次数
	TimesFollowType
	// TimesArchiveType 投稿增加次数
	TimesArchiveType
	// TimesBuyVipType 购买大会员增加次数
	TimesBuyVipType
	// TimesOtherType 其他行为增加次数
	TimesOtherType
	// TimesCustomizeType 大会员增加次数
	TimesCustomizeType
	// TimesOGVType OGV增加次数
	TimesOGVType
	// TimesFeType 前端增加次数
	TimesFeType
	// TimesLikeType 点赞增加次数
	TimesLikeType
	// TimesCoinType 投币增加次数
	TimesCoinType
	// TimesAdditionalType 额外赠送次数
	TimesAdditionalType
	// TimesActType 任务获得抽奖次数
	TimesActType
	// TimesActPointType 任务节点组获得抽奖次数
	TimesActPointType
)

const (
	// TimesAddTimesStateNone 不能领取
	TimesAddTimesStateNone = 1
	// TimesAddTimesStateWait 待领取
	TimesAddTimesStateWait = 2
	// TimesAddTimesStateAlready 已领取
	TimesAddTimesStateAlready = 3
)

// TimesReply ...
type TimesReply struct {
	Times int `json:"times"`
}

// CountNumReply ...
type CountNumReply struct {
	Num int64 `json:"num"`
}

// CountStateReply ...
type CountStateReply struct {
	State int `json:"state"`
}

// RecordDetail ...
type RecordDetail struct {
	ID       int64             `json:"id"`
	Mid      int64             `json:"mid"`
	IP       int64             `json:"ip"`
	Num      int               `json:"num"`
	GiftID   int64             `json:"gift_id"`
	GiftName string            `json:"gift_name"`
	GiftType int               `json:"gift_type"`
	ImgURL   string            `json:"img_url"`
	Type     int               `json:"type"`
	Ctime    xtime.Time        `json:"ctime"`
	CID      int64             `json:"cid"`
	Extra    map[string]string `json:"extra"`
}

// InsertRecord ...
type InsertRecord struct {
	ID      int64  `json:"id"`
	Mid     int64  `json:"mid"`
	Num     int    `json:"num"`
	Type    int    `json:"type"`
	CID     int64  `json:"cid"`
	OrderNo string `json:"order_no"`
	GiftID  int64  `json:"gift_id"`
}

// TimesConfig lottery times config
type TimesConfig struct {
	ID      int64  `json:"id"`
	Sid     string `json:"sid"`
	Type    int    `json:"type"`
	AddType int    `json:"add_type"`
	Times   int    `json:"times"`
	Info    string `json:"info"`
	Most    int    `json:"most"`
	State   int    `json:"state"`
}

// TimesInfo ...
type TimesInfo struct {
	Counter  string `json:"counter"`
	Activity string `json:"activity"`
	Count    int64  `json:"count"`
}

// ConsumeInfo ...
type ConsumeInfo struct {
	Consume int64 `json:"consume"`
	Send    int   `json:"send"`
}

// AddTimes ...
type AddTimes struct {
	ID    int64      `json:"id"`
	Mid   int64      `json:"mid"`
	Type  int        `json:"type"`
	Num   int        `json:"num"`
	CID   int64      `json:"cid"`
	Ctime xtime.Time `json:"ctime"`
}

// TimesInterface ...
type TimesInterface interface {
	IsInternal() bool
	Record() *RecordDetail
}

// GetTimesByType ...
func GetTimesByType(record *RecordDetail) (TimesInterface, error) {
	switch record.Type {
	case TimesBaseType:
		return &BaseTimes{RecordDetail: record}, nil
	case TimesShareType:
		return &ShareTimes{RecordDetail: record}, nil
	case TimesFollowType:
		return &FollowTimes{RecordDetail: record}, nil
	case TimesArchiveType:
		return &ArchiveTimes{RecordDetail: record}, nil
	case TimesBuyVipType:
		return &BuyVipTimes{RecordDetail: record}, nil
	case TimesOtherType:
		return &OtherTimes{RecordDetail: record}, nil
	case TimesCustomizeType:
		return &CustomizeTimes{RecordDetail: record}, nil
	case TimesOGVType:
		return &OGVTimes{RecordDetail: record}, nil
	case TimesFeType:
		return &FeTimes{RecordDetail: record}, nil
	case TimesLikeType:
		return &LikeTimes{RecordDetail: record}, nil
	case TimesCoinType:
		return &CoinTimes{RecordDetail: record}, nil
	case TimesAdditionalType:
		return &AdditionalTimes{RecordDetail: record}, nil
	case TimesActType:
		return &Actimes{RecordDetail: record}, nil
	case TimesActPointType:
		return &ActPointimes{RecordDetail: record}, nil
	}
	return nil, ecode.ActivityLotteryTimesTypeError
}

// BaseTimes ...
type BaseTimes struct {
	*RecordDetail
}

// IsInternal ...
func (t *BaseTimes) IsInternal() bool {
	return false
}

// Record ...
func (t *BaseTimes) Record() *RecordDetail {
	return t.RecordDetail
}

// ShareTimes ...
type ShareTimes struct {
	*RecordDetail
}

// IsInternal ...
func (t *ShareTimes) IsInternal() bool {
	return false
}

// Record ...
func (t *ShareTimes) Record() *RecordDetail {
	return t.RecordDetail
}

// FollowTimes ...
type FollowTimes struct {
	*RecordDetail
}

// IsInternal ...
func (t *FollowTimes) IsInternal() bool {
	return false
}

// Record ...
func (t *FollowTimes) Record() *RecordDetail {
	return t.RecordDetail
}

// ArchiveTimes ...
type ArchiveTimes struct {
	*RecordDetail
}

// IsInternal ...
func (t *ArchiveTimes) IsInternal() bool {
	return true
}

// Record ...
func (t *ArchiveTimes) Record() *RecordDetail {
	return t.RecordDetail
}

// BuyVipTimes ...
type BuyVipTimes struct {
	*RecordDetail
}

// IsInternal ...
func (t *BuyVipTimes) IsInternal() bool {
	return true
}

// Record ...
func (t *BuyVipTimes) Record() *RecordDetail {
	return t.RecordDetail
}

// OtherTimes ...
type OtherTimes struct {
	*RecordDetail
}

// Record ...
func (t *OtherTimes) Record() *RecordDetail {
	return t.RecordDetail
}

// IsInternal ...
func (t *OtherTimes) IsInternal() bool {
	return true
}

// CustomizeTimes ...
type CustomizeTimes struct {
	*RecordDetail
}

// Record ...
func (t *CustomizeTimes) Record() *RecordDetail {
	return t.RecordDetail
}

// IsInternal ...
func (t *CustomizeTimes) IsInternal() bool {
	return true
}

// FeTimes ...
type FeTimes struct {
	*RecordDetail
}

// IsInternal ...
func (t *FeTimes) IsInternal() bool {
	return true
}

// Record ...
func (t *FeTimes) Record() *RecordDetail {
	return t.RecordDetail
}

// OGVTimes ...
type OGVTimes struct {
	*RecordDetail
}

// Record ...
func (t *OGVTimes) Record() *RecordDetail {
	return t.RecordDetail
}

// IsInternal ...
func (t *OGVTimes) IsInternal() bool {
	return true
}

// LikeTimes ...
type LikeTimes struct {
	*RecordDetail
}

// Record ...
func (t *LikeTimes) Record() *RecordDetail {
	return t.RecordDetail
}

// IsInternal ...
func (t *LikeTimes) IsInternal() bool {
	return false
}

// CoinTimes ...
type CoinTimes struct {
	*RecordDetail
}

// Record ...
func (t *CoinTimes) Record() *RecordDetail {
	return t.RecordDetail
}

// IsInternal ...
func (t *CoinTimes) IsInternal() bool {
	return false
}

// Actimes ...
type Actimes struct {
	*RecordDetail
}

// Record ...
func (t *Actimes) Record() *RecordDetail {
	return t.RecordDetail
}

// IsInternal ...
func (t *Actimes) IsInternal() bool {
	return true
}

// AdditionalTimes ...
type AdditionalTimes struct {
	*RecordDetail
}

// Record ...
func (t *AdditionalTimes) Record() *RecordDetail {
	return t.RecordDetail
}

// IsInternal ...
func (t *AdditionalTimes) IsInternal() bool {
	return true
}

// ActPointimes ...
type ActPointimes struct {
	*RecordDetail
}

// Record ...
func (t *ActPointimes) Record() *RecordDetail {
	return t.RecordDetail
}

// IsInternal ...
func (t *ActPointimes) IsInternal() bool {
	return true
}
