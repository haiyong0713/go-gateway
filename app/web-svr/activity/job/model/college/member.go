package college

// MidInfo ...
type MidInfo struct {
	MID     int64 `json:"mid"`
	Inviter int64 `json:"inviter"`
	MidType int64 `json:"mid_type"`
}

// Personal 个人信息
type Personal struct {
	MID       int64 `json:"mid"`
	Score     int64 `json:"score"`
	Rank      int   `json:"rank"`
	Diff      int64 `json:"diff"`
	CollegeID int64 `json:"college_id"`
}

// Archive ...
type Archive struct {
	AID   int64 `json:"aid"`
	Score int64 `json:"score"`
}

// MIDCtime ...
type MIDCtime struct {
	MID   int64 `json:"mid"`
	Ctime int64 `json:"ctime"`
	AID   int64 `json:"aid"`
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
