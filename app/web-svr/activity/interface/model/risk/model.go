package risk

const (
	// PlatformWeb 平台web
	PlatformWeb = "web"
	// ActionLottery 抽奖场景名
	ActionLottery = "activity_lottery"
	// Reject 命中
	Reject = "reject"
	// ActionSf21Sign 签到场景
	ActionSf21Sign = "activity_ny_card_sign_action"
	// ActionSf21Follow 关注场景
	ActionSf21Follow = "activity_ny_card_relation_action"
	// ActionSf21SendCard 送卡场景
	ActionSf21SendCard = "activity_ny_card_share_action"
	// ActionSf21Join 邀请场景
	ActionSf21Join = "activity_ny_card_friend_action"
	// ActionSf21Compose 合成场景
	ActionSf21Compose = "activity_ny_card_fix_action"

	// ActionCardsSign 签到场景
	ActionCardsSign = "activity_ny_card_sign_action"
	// ActionCardsFollow 关注场景
	ActionCardsFollow = "activity_ny_card_relation_action"
	// ActionCardsSendCard 送卡场景
	ActionCardsSendCard = "activity_ny_card_share_action"
	// ActionCardsJoin 邀请场景
	ActionCardsJoin = "activity_ny_card_friend_action"
	// ActionCardsCompose 合成场景
	ActionCardsCompose = "activity_ny_card_fix_action"
	//ActionVoteNew: 新投票活动
	ActionVoteNew  = "activity_common_vote_new"
	ActionCompose  = "compose"
	ActionSendCard = "donate"
)

// Base ...
type Base struct {
	Buvid      string `form:"buvid" json:"buvid"`
	Origin     string `form:"origin" json:"origin"`
	Referer    string `form:"referer" json:"referer"`
	IP         string `form:"ip" json:"ip"`
	Ctime      string `form:"ctime" json:"ctime"`
	UserAgent  string `form:"user_agent" json:"user_agent"`
	Build      string `form:"build" json:"build"`
	Platform   string `form:"platform" json:"platform"`
	Action     string `form:"action" json:"action"`
	MID        int64  `form:"mid" json:"mid"`
	API        string `form:"api" json:"api"`
	EsTime     int64  `form:"estime" json:"estime"`
	ActivityID int64  `form:"activity_id" json:"activity_id"`
}

// Lottery ...
type Lottery struct {
	Base
	RecordID int64  `form:"record_id" json:"record_id"`
	SID      string `form:"id" json:"id"`
	Name     string `form:"lottery_name" json:"lottery_name"`
	Author   string `form:"operator" json:"operator"`
	Stime    string `form:"start_time" json:"start_time"`
	Etime    string `form:"end_time" json:"end_time"`
	RcLevel  int    `form:"rc_level" json:"rc_level"`
}

// Sf21Sign 春节签到
type Sf21Sign struct {
	Base
	Mid         int64  `json:"mid"`
	ActivityUID string `json:"activity_uid"`
	MobiApp     string `json:"mobi_app"`
}

// Sf21Follow 春节关注
type Sf21Follow struct {
	Base
	Mid         int64  `json:"mid"`
	Fid         string `json:"fid"`
	ActivityUID string `json:"activity_uid"`
	MobiApp     string `json:"mobi_app"`
}

// Sf21SendCard 春节送卡
type Sf21SendCard struct {
	Base
	Mid         int64  `json:"mid"`
	InvitedMid  int64  `json:"invited_mid"`
	CardType    int64  `json:"card_type"`
	ActivityUID string `json:"activity_uid"`
	MobiApp     string `json:"mobi_app"`
}

// Sf21Invited 春节邀请
type Sf21Invited struct {
	Base
	Mid         int64  `json:"mid"`
	InvitedMid  int64  `json:"invited_mid"`
	ActivityUID string `json:"activity_uid"`
	MobiApp     string `json:"mobi_app"`
}

// Sf21Compose 春节合成
type Sf21Compose struct {
	Base
	Mid         int64  `json:"mid"`
	ActivityUID string `json:"activity_uid"`
	MobiApp     string `json:"mobi_app"`
}

// Sf21Compose 春节合成
type VoteNew struct {
	Base
	ActivityUID string `json:"activity_uid"`
	TargetId    int64  `json:"target_id"`
	TargetName  string `json:"target_name"`
	TargetType  string `json:"target_type"`
	Id          int64  `json:"id"`
	Score       int64  `json:"score"`
}

// Task 任务通用
type Task struct {
	Base
	Mid         int64  `json:"mid"`
	TargetMid   int64  `json:"target_mid"`
	ActivityUID string `json:"activity_uid"`
	Subscene    string `json:"subscene"`
	Action      string `json:"action"`
}
