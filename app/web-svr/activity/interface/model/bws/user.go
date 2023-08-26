package bws

import xtime "go-common/library/time"

// User .
type User struct {
	User               *UserInfo                     `json:"user"`
	Achievements       []*UserAchieveDetail          `json:"achievements"`
	UnlockAchievements []*UserAchieveDetail          `json:"unlock_achievements"`
	Items              map[string][]*UserPointDetail `json:"items"`
}

// CategoryAchieve .
type CategoryAchieve struct {
	Achievements       []*UserAchieveDetail `json:"achievements"`
	UnlockAchievements []*UserAchieveDetail `json:"unlock_achievements"`
}

// AchieveRank .
type AchieveRank struct {
	SelfRank  int         `json:"self_rank"`
	SelfPoint int64       `json:"self_point"`
	List      []*UserInfo `json:"list"`
}

// UserInfo .
type UserInfo struct {
	Mid          int64  `json:"mid"`
	Name         string `json:"name"`
	Key          string `json:"key"`
	Face         string `json:"face"`
	Hp           int64  `json:"hp"`
	AchievePoint int64  `json:"achieve_point"`
	AchieveRank  int    `json:"achieve_rank,omitempty"`
}

// LotteryUser .
type LotteryUser struct {
	Mid         int64  `json:"mid"`
	Name        string `json:"name"`
	Face        string `json:"face"`
	AchieveRank int    `json:"achieve_rank"`
}

// LotteryCache .
type LotteryCache struct {
	Bid  int64 `json:"bid"`
	Mid  int64 `json:"mid"`
	Rank int   `json:"rank"`
}

// AdminInfo .
type AdminInfo struct {
	IsAdmin bool        `json:"is_admin"`
	Point   interface{} `json:"point"`
	Award   interface{} `json:"award"`
}

type User2020 struct {
	User        *UserInfo2020 `json:"user"`
	Tasks       []*UserTask   `json:"tasks"`
	LotteryLog  []*LotteryLog `json:"lottery_log"`
	OnlineAward []*LotteryLog `json:"online_award"`
}

type UserInfo2020 struct {
	Mid          int64  `json:"mid"`
	Name         string `json:"name"`
	Key          string `json:"key"`
	Face         string `json:"face"`
	LotteryTimes int64  `json:"lottery_times"`
	Star         int64  `json:"star"`
}

type UserTask struct {
	TaskID      int64              `json:"task_id"`
	NowCount    int64              `json:"now_count"`
	UserState   int64              `json:"user_state"`
	AwardState  int64              `json:"award_state"`
	TaskDay     int64              `json:"task_day"`
	Title       string             `json:"title"`
	FinishCount int64              `json:"finish_count"`
	OrderNum    int64              `json:"order_num"`
	Points      []*UserPointDetail `json:"points"`
}

type LotteryLog struct {
	ID      int64      `json:"id"`
	AwardID int64      `json:"award_id"`
	Title   string     `json:"title"`
	Stage   string     `json:"stage"`
	Intro   string     `json:"intro"`
	State   string     `json:"state"`
	Amount  int64      `json:"amount"`
	Image   string     `json:"image"`
	Ctime   xtime.Time `json:"ctime"`
	Mtime   xtime.Time `json:"mtime"`
}

// UserRank ...
type UserRank struct {
	Mid   int64   `json:"mid"`
	Score float64 `json:"score"`
}

type GameStarRank struct {
	Star int64  `json:"star"`
	Rank string `json:"rank"`
}

// UserDetailReply ...
type UserDetailReply struct {
	*UserDetail
	StarGameDetail map[int64]*GameStarRank `json:"star_game_detail"`
	StarGame       map[int64]int64         `json:"star_game"`
	Rank           int64                   `json:"rank"`
	LotteryRemain  int64                   `json:"lottery_remain"`
	LotteryLog     []*LotteryLog           `json:"lottery_log"`
	User           *UserInfo2020           `json:"user"`
	RankEntryNum   int64                   `json:"rank_entry_num"`
	RankFirstNum   int64                   `json:"rank_first_num"`
}
