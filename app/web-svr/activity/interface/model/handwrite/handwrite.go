package handwrite

const (
	// GodName 神仙模式
	GodName = "god"
	// TiredName 爆肝模式
	TiredName = "tired"
	// NewName 新人
	NewName = "new"
	// HandWriteKey 手书key
	HandWriteKey = "handWrite"
	// AwardCanGet 可获奖
	AwardCanGet = 1
	// Tired1Name 爆肝模式
	Tired1Name = "tired1"
	// Tired2Name 爆肝模式
	Tired2Name = "tired2"
	// Tired3Name 爆肝模式
	Tired3Name = "tired3"
)

// MidTaskAll 用户任务完成情况
type MidTaskAll struct {
	Mid         int64 `json:"mid"`
	God         int64 `json:"god"`
	TiredLevel1 int64 `json:"tired_level_1"`
	TiredLevel2 int64 `json:"tired_level_2"`
	TiredLevel3 int64 `json:"tired_level_3"`
}

// AwardCountNew 任务完成情况统计
type AwardCountNew struct {
	God         int64 `json:"god"`
	TiredLevel1 int64 `json:"tired_level_1"`
	TiredLevel2 int64 `json:"tired_level_2"`
	TiredLevel3 int64 `json:"tired_level_3"`
}

// MidAward 用户获奖情况
type MidAward struct {
	God   int   `json:"god"`
	Tired int   `json:"tired"`
	New   int   `json:"new"`
	Score int64 `json:"score"`
	Rank  int   `json:"rank"`
}

// AwardCount 用户获奖人数统计
type AwardCount struct {
	God   int `json:"god"`
	Tired int `json:"tired"`
	New   int `json:"new"`
}

// AwardCountMoney 获奖用户及总金额
type AwardCountMoney struct {
	Money int `json:"money"`
	Count int `json:"count"`
}

// AwardMemberReply 获奖用户及总金额返回结构
type AwardMemberReply struct {
	MoneyCount map[string]*AwardCountMoney `json:"money_count"`
}

// Account 账号信息
type Account struct {
	Mid  int64  `json:"mid"`
	Name string `json:"name"`
	Face string `json:"face"`
	Sign string `json:"sign"`
	Sex  string `json:"sex"`
}

// RankMember 用户积分信息
type RankMember struct {
	Account *Account `json:"account"`
	Score   int64    `json:"score"`
}

// RankReply 排行榜返回结构
type RankReply struct {
	Rank []*RankMember `json:"rank"`
}

// PersonalReply 个人信息情况返回
type PersonalReply struct {
	Money   int      `json:"money"`
	Score   int64    `json:"score"`
	Rank    int      `json:"rank"`
	Account *Account `json:"account"`
}

// Personal2021Reply 个人信息情况返回
type Personal2021Reply struct {
	Money int64 `json:"money"`
}

// AddTimesReply 增加抽奖次数返回
type AddTimesReply struct {
}

// CoinReply 硬币数
type CoinReply struct {
	Coin        int64 `json:"coin"`
	CanAddTimes bool  `json:"can_addtimes"`
}
