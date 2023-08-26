package bnj

import "go-gateway/app/web-svr/activity/interface/api"

const (
	ActionTypeIncr = 1
	ActionTypeDecr = 2

	RewardTypeOfNotLottery  = 0
	RewardTypeOfNotReceived = 1
	RewardTypeOfReceived    = 2

	SceneID4ARDraw   = 1
	SceneID4LiveView = 2
	SceneID4Reserve  = 3
)

type RewardConfig struct {
	ReserveActivityID     int64 `json:"reserve_activity_id" toml:"reserve_activity_id"`
	RewardID4Reserve      int64 `json:"reward_id_4_reserve" toml:"reward_id_4_reserve"`
	RewardID4LiveAR       int64 `json:"reward_id_4_live_AR" toml:"reward_id_4_live_AR"`
	RewardID4LiveDuration int64 `json:"reward_id_4_live_duration" toml:"reward_id_4_live_duration"`
	PCVID                 int64 `json:"pc_vid"`
}

type UserARDrawCoupon struct {
	MID    int64 `json:"mid"`
	Coupon int64 `json:"coupon"`
}

type UserCouponLogInLiveRoom struct {
	MID         int64 `json:"mid"`
	ReceiveUnix int64 `json:"related_id"`
	No          int64 `json:"no"`
}

type UserRewardInLiveRoom struct {
	MID         int64                      `json:"mid"`
	SceneID     int64                      `json:"scene_id"`
	ReceiveUnix int64                      `json:"related_id"`
	No          int64                      `json:"no"`
	Duration    int64                      `json:"duration"`
	Reward      *api.RewardsSendAwardReply `json:"reward"`
}

type UserInLiveRoomFor2021 struct {
	ID       int64  `json:"-"`
	MID      int64  `json:"uid"`
	AID      int64  `json:"aid"`
	AType    string `json:"atype"`
	Duration int64  `json:"duration"`
	UniqueID string `json:"unique_id"`
}

type UserLotteryFor2021 struct {
	Duration int64
	MID      int64
}

type ExamStatsRule struct {
	Url       string `json:"url"`
	StartTime int64  `json:"start_time"`
	EndTime   int64  `json:"end_time"`
}

type LotteryRuleFor2021 struct {
	Duration  int64 `json:"duration"`
	StartTime int64 `json:"start_time"`
	EndTime   int64 `json:"end_time"`
}

type ReserveRewardRuleFor2021 struct {
	Count      int64 `json:"count"`
	StartTime  int64 `json:"start_time"`
	EndTime    int64 `json:"end_time"`
	RewardID   int64 `json:"reward_id"`
	ActivityID int64 `json:"activity_id"`
	StartID    int64 `json:"start_id"`
	EndID      int64 `json:"end_id"`
}

type ReservedUser struct {
	ID  int64
	MID int64
}

// ResetMsg .
type ResetMsg struct {
	Mid int64 `json:"mid"`
	Ts  int64 `json:"ts"`
}

// Push .
type Push struct {
	Second        int64  `json:"second"`
	Name          string `json:"name"`
	TimelinePic   string `json:"timeline_pic,omitempty"`
	H5TimelinePic string `json:"h5_timeline_pic,omitempty"`
}

type PushHotpotBnj20 struct {
	Type             int    `json:"type"`
	Value            int64  `json:"value"`
	Msg              string `json:"msg"`
	BlockGame        int    `json:"block_game,omitempty"`
	BlockGameAction  int    `json:"block_game_action,omitempty"`
	BlockMaterial    int    `json:"block_material,omitempty"`
	TimelinePic      string `json:"timeline_pic,omitempty"`
	H5TimelinePic    string `json:"h5_timeline_pic,omitempty"`
	ShareTimelinePic string `json:"share_timeline_pic,omitempty"`
}

type PushReserveBnj20 struct {
	Type          int   `json:"type"`
	ReservedCount int64 `json:"reserved_count"`
}

type Action struct {
	Mid     int64  `json:"mid"`
	Type    int    `json:"type"`
	Num     int64  `json:"num"`
	Message string `json:"message"`
	Ts      int64  `json:"ts"`
}

type AwardAction struct {
	Mid          int64  `json:"mid"`
	ID           int64  `json:"id"`
	Type         int    `json:"type"`
	SourceID     string `json:"source_id"`
	SourceExpire int64  `json:"source_expire"`
	TaskID       int64  `json:"task_id"`
	Mirror       string `json:"mirror"`
}

type LiveMsg struct {
	Uids    []int64       `json:"uids"`
	MsgID   string        `json:"msg_id"`
	Source  int64         `json:"source"`
	Rewards []*LiveReward `json:"rewards"`
}

type LiveReward struct {
	RewardID   int64       `json:"reward_id"`
	StartTime  int64       `json:"start_time,omitempty"`
	ExpireTime int64       `json:"expire_time"`
	Type       int         `json:"type"`
	Num        int         `json:"num,omitempty"`
	ExtraData  interface{} `json:"extra_data,omitempty"`
}

func (rule *LotteryRuleFor2021) DeepCopy() *LotteryRuleFor2021 {
	newOne := new(LotteryRuleFor2021)
	{
		newOne.Duration = rule.Duration
		newOne.StartTime = rule.StartTime
		newOne.EndTime = rule.EndTime
	}

	return newOne
}

func (rule *ReserveRewardRuleFor2021) DeepCopy() *ReserveRewardRuleFor2021 {
	newOne := new(ReserveRewardRuleFor2021)
	{
		newOne.Count = rule.Count
		newOne.StartTime = rule.StartTime
		newOne.EndTime = rule.EndTime
		newOne.ActivityID = rule.ActivityID
		newOne.RewardID = rule.RewardID
	}

	return newOne
}

func (pool *RewardConfig) DeepCopy() *RewardConfig {
	newOne := new(RewardConfig)
	{
		newOne.ReserveActivityID = pool.ReserveActivityID
		newOne.RewardID4Reserve = pool.RewardID4Reserve
		newOne.RewardID4LiveAR = pool.RewardID4LiveAR
		newOne.RewardID4LiveDuration = pool.RewardID4LiveDuration
	}

	return newOne
}

func (info *ReservedUser) DeepCopy() *ReservedUser {
	newOne := new(ReservedUser)
	{
		newOne.ID = info.ID
		newOne.MID = info.MID
	}

	return newOne
}

func (user *UserInLiveRoomFor2021) DeepCopy() *UserInLiveRoomFor2021 {
	newOne := new(UserInLiveRoomFor2021)
	{
		newOne.ID = user.ID
		newOne.MID = user.MID
		newOne.AID = user.AID
		newOne.AType = user.AType
		newOne.Duration = user.Duration
		newOne.UniqueID = user.UniqueID
	}

	return newOne
}
