package rewards

import (
	xtime "go-common/library/time"
)

type Configs struct {
	AwardMap map[int64] /*awardId*/ *GenericConfig
}

type AddAwardParam struct {
	ActivityId  int64  `form:"activity_id" validate:"min=1,required"`
	Type        string `form:"type" validate:"required"`
	DisplayName string `form:"display_name" validate:"required"`
	//是否发送私信通知
	ShouldSendNotify int64 `form:"should_send_notify" `
	//发奖通知发送人
	NotifySenderId int64 `form:"notify_sender_id" `
	//卡片通知码
	NotifyCode string `form:"notify_code" `
	//发奖通知内容
	NotifyMessage string `form:"notify_message"`
	//发奖跳转链接1
	NotifyJumpUri1 string `form:"notify_jump_uri1"`
	//发奖跳转链接2
	NotifyJumpUri2 string            `form:"notify_jump_uri2"`
	JsonStr        string            `form:"json_str" validate:"required"`
	ExtraInfo      map[string]string `form:"extra_info"`
	IconUrl        string            `form:"icon_url"`
}

type UpdateAwardParam struct {
	Id          int64  `form:"id" validate:"min=1,required"`
	ActivityId  int64  `form:"activity_id" validate:"min=1,required"`
	Type        string `form:"type" validate:"required"`
	DisplayName string `form:"display_name" validate:"required"`
	//是否发送私信通知
	ShouldSendNotify int64 `form:"should_send_notify" `
	//发奖通知发送人
	NotifySenderId int64 `form:"notify_sender_id" `
	//卡片通知码
	NotifyCode string `form:"notify_code" `
	//发奖通知内容
	NotifyMessage string `form:"notify_message"`
	//发奖跳转链接1
	NotifyJumpUri1 string `form:"notify_jump_uri1"`
	//发奖跳转链接2
	NotifyJumpUri2 string            `form:"notify_jump_uri2"`
	JsonStr        string            `form:"json_str" validate:"required"`
	ExtraInfo      map[string]string `form:"extra_info"`
	IconUrl        string            `form:"icon_url"`
}

type GenericConfig struct {
	//配置ID
	Id           int64  `json:"id" validate:"min=1,required"`
	ActivityId   int64  `json:"activity_id" validate:"min=1,required"`
	ActivityName string `json:"activity_name"`
	Type         string `json:"type" validate:"required"`
	DisplayName  string `json:"display_name" validate:"required"`
	//是否发送私信通知
	ShouldSendNotify bool `json:"should_send_notify"`
	//发奖通知发送人
	NotifySenderId int64 `json:"notify_sender_id" `
	//卡片通知码
	NotifyCode string `json:"notify_code" `
	//发奖通知内容
	NotifyMessage string `json:"notify_message"`
	//发奖跳转链接1
	NotifyJumpUri1 string `json:"notify_jump_uri1"`
	//发奖跳转链接2
	NotifyJumpUri2 string            `json:"notify_jump_uri2"`
	JsonStr        string            `json:"json_str" validate:"required"`
	ExtraInfo      map[string]string `json:"extra_info"`
	IconUrl        string            `json:"icon_url"`
}

type AwardSentInfo struct {
	Mid          int64             `json:"mid"`
	AwardId      int64             `json:"award_id"`
	AwardName    string            `json:"award_name" validate:"required"`
	ActivityId   int64             `json:"activity_id,omitempty"`
	ActivityName string            `json:"activity_name,omitempty"`
	Type         string            `json:"type"`
	IconUrl      string            `json:"icon"`
	SentTime     xtime.Time        `json:"receive_time"`
	ExtraInfo    map[string]string `json:"extra_info"`
}

type ComicsCouponConfig struct {
	Type int64 `validate:"min=1,required"`
}

type DanmukuConfig struct {
	Color      int64 `validate:"min=1,required"`
	ExpireDays int64 `validate:"min=1,required"`
	RoomIds    []int64
}

type SuitConfig struct {
	Id         int64 `validate:"min=1,required"`
	ExpireDays int64 `validate:"required"`
}

type DressUpConfig struct {
	Id         int64 `validate:"min=1,required"`
	ExpireDays int64 `validate:"required"`
}

type MallCouponConfig struct {
	CouponId string `validate:"required"`
}

type MallCouponConfigV2 struct {
	SourceAuthorityId string `validate:"required"`
}

type MallCouponPayConfigV2 struct {
	SourceAuthorityId string `validate:"required"`
}

type MallPrizeConfigV2 struct {
	SourceAuthorityId string `validate:"required"`
}

type MallPrizeConfig struct {
	PrizeNo     int64  `validate:"min=1,required"`
	PrizePoolId int64  `validate:"min=1,required"`
	GameId      string `validate:"required"`
}

type VipCouponConfig struct {
	BatchToken string `validate:"required"`
	AppKey     string `validate:"required"`
}

type VipConfig struct {
	BatchToken string `validate:"required"`
	AppKey     string `validate:"required"`
}

type ActCounterConfig struct {
	Points   int64  `validate:"min=1,required"`
	Activity string `validate:"required"`
	Business string `validate:"required"`
	Extra    string
}

type ClassCouponConfig struct {
	BatchToken string `validate:"required"`
	SendVc     bool
}

type CashConfig struct {
	CustomerId   string `validate:"required"`
	ActivityID   string `validate:"required"`
	TransBalance int64
	StartTme     int64
}

type CdKeyConfig struct {
}

type LiveGoldConfig struct {
	OrderSource int64 `validate:"min=1"`
	Type        int64 `validate:"min=1"`
	Count       int64 `validate:"min=1"`
	Remark      string
	PoolId      int64 `validate:"min=1"`
}

type Bnj2021Lottery struct {
	Count int64 `validate:"min=1,required"`
}

type GarbCouponConfig struct {
	BatchToken string `validate:"required"`
}

type GarbDiyToolConfig struct {
	ActivityId int64 `validate:"required"`
	ExpireDays int64 `validate:"required"`
}

type EmptyConfig struct {
}

type EntityConfig struct {
}

type TencentGameConfig struct {
	AccountInfoId int64  `validate:"required"`
	FlowId        string `validate:"required"`
}

type PackageConfig struct {
	AwardIds []int64 `validate:"required"`
}

type UserSendHistory struct {
	Mid         int64  `json:"mid"`
	ActivityId  int64  `json:"activity_id"`
	AwardId     int64  `json:"award_id"`
	DisplayName string `json:"display_name"`
	State       int64  `json:"state"`
}

type AsyncSendingAwardInfo struct {
	Mid       int64  `json:"mid"`
	UniqueId  string `json:"unique_id"`
	Business  string `json:"business"`
	AwardId   int64  `json:"award_id"`
	AwardType string `json:"award_type`
	SendTime  int64  `json:"send_time"`
}

type AddressInfo struct {
	ID      int64  `json:"id"`
	Type    int64  `json:"type"`
	Def     int64  `json:"def"`
	ProvID  int64  `json:"prov_id"`
	CityID  int64  `json:"city_id"`
	AreaID  int64  `json:"area_id"`
	Name    string `json:"name"`
	Phone   string `json:"phone"`
	Addr    string `json:"addr"`
	ZipCode string `json:"zip_code"`
	Prov    string `json:"prov"`
	City    string `json:"city"`
	Area    string `json:"area"`
}

type CdKeyInfo struct {
	Mid       int64      `json:"mid"`
	CdKeyName string     `json:"cdkey_name"`
	Cdkey     string     `json:"cdkey"`
	Mtime     xtime.Time `json:"mtime"`
}
