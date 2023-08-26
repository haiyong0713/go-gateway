package question

import "go-gateway/app/web-svr/activity/interface/model/question"

const (
	StateInit     = 0
	StateOnline   = 1
	StateOffline  = 2
	State4Process = 3
)

// BaseItem .
type BaseItem struct {
	*question.Base
	DetailIDs []int64
}

type NewBaseItem struct {
	*question.Base
	Details []*question.Detail
}

type AnswerUserInfo struct {
	IsJoin      int64         `json:"is_join"`
	Mid         int64         `json:"mid"`
	UserScore   int64         `json:"user_score"`
	UserRank    int           `json:"user_rank"`
	UserPercent int64         `json:"user_percent"`
	AnswerTimes int64         `json:"answer_times"`
	CanPendant  int64         `json:"can_pendant"`
	HavePendant int64         `json:"have_pendant"`
	LastInfo    *UserLastInfo `json:"last_info"`
	ID          int64         `json:"id"`
	FinishTime  int64         `json:"finish_time"`
}
type UserLastInfo struct {
	LastScore int64 `json:"last_score"`
	LastRank  int64 `json:"last_rank"`
}

type AnswerUser struct {
	ID  int64
	Mid int64
}

type UserRank struct {
	OrderNumber int          `json:"order_number"`
	UserScore   int64        `json:"user_score"`
	AnswerTimes int64        `json:"answer_times"`
	Account     *AccountInfo `json:"account"`
}

type RankInfo struct {
	Mid         int64 `json:"mid"`
	OrderNumber int   `json:"order_number"`
	UserScore   int64 `json:"user_score"`
	AnswerTimes int64 `json:"answer_times"`
}

// AccountInfo
type AccountInfo struct {
	Mid  int64  `json:"mid"`
	Name string `json:"name"`
	Face string `json:"face"`
}

type HourPeople struct {
	WeekPeople  int64           `json:"week_people"`
	PeopleCount map[int64]int64 `json:"people_count"`
}
