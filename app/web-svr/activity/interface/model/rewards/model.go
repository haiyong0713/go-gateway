package rewards

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

// ActPlatActivityPoints ...
type ActPlatActivityPoints struct {
	Points    int64  `json:"points"`    // 积分增减值
	Timestamp int64  `json:"timestamp"` // 事件发生的时间戳
	Mid       int64  `json:"mid"`
	Source    int64  `json:"source"`   // 积分原因，一般是关联的资源id，例如关注的up主id，邀请的用户id
	Activity  string `json:"activity"` // 关联活动名，开学季活动此处填 college2020
	Business  string `json:"business"` // 加分相关业务名，关注：follow，邀请：invite，投稿额外加分：bonus
	Extra     string `json:"extra"`    // 保留字段
}
