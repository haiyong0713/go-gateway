package guess

import (
	"go-common/library/time"
	"go-gateway/app/web-svr/activity/job/model"
)

type GuessUser struct {
	ID        int64         `json:"id"`
	Mid       int64         `json:"mid"`
	MainID    int64         `json:"main_id"`
	DetailID  int64         `json:"detail_id"`
	StakeType int64         `json:"stake_type"`
	Stake     int64         `json:"stake"`
	Income    int64         `json:"income"`
	Status    int64         `json:"status"`
	Ctime     model.StrTime `json:"ctime"`
	Mtime     model.StrTime `json:"mtime"`
}

type UserLog struct {
	ID           int64 `json:"id"`
	Business     int64 `json:"business"`
	Mid          int64 `json:"mid"`
	StakeType    int64 `json:"stake_type"`
	TotalGuess   int64 `json:"total_guess"`
	TotalSuccess int64 `json:"total_success"`
	SuccessRate  int64 `json:"success_rate"`
}

// ImMsgParam im messgage param
type ImMsgParam struct {
	SenderUID uint64   `json:"sender_uid"` //官号uid：发送方uid
	MsgKey    uint64   `json:"msg_key"`    //消息唯一标识
	MsgType   int32    `json:"msg_type"`   //文本类型 type = 1
	Content   string   `json:"content"`    //{"content":"test" //文本内容}
	RecverIDs []uint64 `json:"recver_ids"` //多人消息，列表型，限定每次客户端发送<=100
}

// MainMsg canal main message.
type MainMsg struct {
	ID         int64  `json:"id"`
	Business   int64  `json:"business"`
	Oid        int64  `json:"oid"`
	Title      string `json:"title"`
	StakeType  int64  `json:"stake_type"`
	MaxStake   int64  `json:"max_stake"`
	ResultID   int64  `json:"result_id"`
	GuessCount int64  `json:"guess_count"`
	IsDeleted  int64  `json:"is_deleted"`
}

type ContestDetail struct {
	ContestID int64     `json:"contest_id"`
	Timestamp int64     `json:"timestamp"`
	Title     string    `json:"title"`
	Home      *TeamInfo `json:"home"`
	Away      *TeamInfo `json:"away"`
	Status    string    `json:"status"`
	Win       string    `json:"win"`
	Predict   string    `json:"predict"`
	Coins     int64     `json:"coins"`
}

type McMsg struct {
	Contest *Contest `json:"contest"`
	More    *More    `json:"more"`
}

type Contest struct {
	ID        int64     `json:"id"`
	StartTime int64     `json:"start_time"`
	EndTime   int64     `json:"end_time"`
	Title     string    `json:"title"`
	Status    string    `json:"status"`
	Home      *TeamInfo `json:"home"`
	Away      *TeamInfo `json:"away"`
	Win       string    `json:"win"`
	Predict   string    `json:"predict"`
	Coins     int64     `json:"coins"`
}

type More struct {
	Status  string `json:"status"`
	Title   string `json:"title"`
	Link    string `json:"link"`
	OnClick string `json:"on_click"`
}

type TeamInfo struct {
	Icon     string `json:"icon"`
	Name     string `json:"name"`
	Wins     int64  `json:"wins"`
	Region   string `json:"region"`
	RegionID string `json:"region_id"`
}

type PredictDetail struct {
	ContestID  int64  `json:"contest_id"`
	PredTeam   string `json:"pred_team"`
	PredStatus string `json:"pred_status"`
	PredCoins  int    `json:"pred_coins"`
	WinCoins   int    `json:"win_coins"`
}

// DetailOption.
type DetailOption struct {
	MainID   int64  `json:"main_id"`
	DetailID int64  `json:"detail_id"`
	Option   string `json:"option"`
	Oid      int64  `json:"oid"`
}
type FinishGuessFailTask struct {
	//auto increment primary key
	Id                                          int64
	MainID, ResultID, Business, Oid, TableIndex int64
	Odds                                        float64
	CreateTime                                  time.Time
}
