package springfestival2021

import (
	xtime "go-common/library/time"
)

const (
	// CardID1 卡片1
	CardID1 = 1
	// CardID2 卡片2
	CardID2 = 2
	// CardID3 卡片3
	CardID3 = 3
	// CardID4 卡片4
	CardID4 = 4
	// CardID5 卡片5
	CardID5 = 5
	// Card1DB ,,,
	Card1DB = "card_1"
	// Card2DB ,,,
	Card2DB = "card_2"
	// Card3DB ,,,
	Card3DB = "card_3"
	// Card4DB ,,,
	Card4DB = "card_4"
	// Card5DB ,,,
	Card5DB = "card_5"
	// StateFinish 任务完成
	StateFinish = 1
	// IsReceived 已领取
	IsReceived = 1
	// IsInStock 有库存
	IsInStock = 1
)

// Card 卡
type Card struct {
	ID       int64      `json:"id"`
	GiftID   int64      `json:"gift_id"`
	GiftName string     `json:"gift_name"`
	ImgURL   string     `json:"img_url"`
	CardID   int64      `json:"card_id"`
	Ctime    xtime.Time `json:"ctime"`
}

// MidCard 用户卡片情况
type MidCard struct {
	Card1   int64 `json:"1"`
	Card2   int64 `json:"2"`
	Card3   int64 `json:"3"`
	Card4   int64 `json:"4"`
	Card5   int64 `json:"5"`
	Compose int64 `json:"compose"`
}

// MidNums ...
type MidNums struct {
	Compose   int64 `json:"compose"`
	Card1     int64 `json:"card1"`
	Card1Used int64 `json:"card1_used"`
	Card2     int64 `json:"card2"`
	Card2Used int64 `json:"card2_used"`
	Card3     int64 `json:"card3"`
	Card3Used int64 `json:"card3_used"`
	Card4     int64 `json:"card4"`
	Card4Used int64 `json:"card4_used"`
	Card5     int64 `json:"card5"`
	Card5Used int64 `json:"card5_used"`
	MID       int64 `json:"mid"`
}

// CardTokenMid ...
type CardTokenMid struct {
	Mid        int64 `json:"mid"`
	CardID     int64 `json:"card_id"`
	IsReceived int   `json:"is_received"`
	ReceiveMid int64 `json:"receive_mid"`
}

// FollowMid ...
type FollowMid struct {
	Mid  int64  `json:"mid"`
	Desc string `json:"desc"`
	Date string `json:"date"`
}

// OgvLink ...
type OgvLink struct {
	Link string `json:"link"`
	Date string `json:"date"`
}

// FollowerReply 关注人信息
type FollowerReply struct {
	List      []*Follower `json:"list"`
	AllFollow bool        `json:"all_follow"`
}

// ShareTokenToMidReply ...
type ShareTokenToMidReply struct {
	Account *Account `json:"account"`
}

// CardTokenToMidReply ...
type CardTokenToMidReply struct {
	Account *Account        `json:"account"`
	Card    *CardIsReceived `json:"card"`
}

// CardIsReceived ...
type CardIsReceived struct {
	CardID     int64 `json:"card_id"`
	IsReceived int   `json:"is_received"`
	IsInStock  int   `json:"is_in_stock"`
	Mid        int64 `json:"mid"`
}

// Follower 关注人信息
type Follower struct {
	Account  *Account `json:"account"`
	IsFollow bool     `json:"is_followed"`
	Desc     string   `json:"desc"`
}

// Account 账号信息
type Account struct {
	Mid  int64  `json:"mid"`
	Name string `json:"name"`
	Face string `json:"face"`
	Sign string `json:"sign"`
	Sex  string `json:"sex"`
}

// CardsReply ...
type CardsReply struct {
	Cards      *MidCard `json:"cards"`
	CanCompose bool     `json:"can_compose"`
}

// CardsNum ...
type CardsNum struct {
	CardID int64 `json:"card_id"`
	Nums   int64 `json:"num"`
}

// Task ...
type Task struct {
	ID          int64  `json:"id"`
	TaskName    string `json:"task_name"`
	LinkName    string `json:"link_name"`
	OrderID     int64  `json:"order_id"`
	Activity    string `json:"activity"`
	ActivityID  int64  `json:"activity_id"`
	Counter     string `json:"counter"`
	Desc        string `json:"desc"`
	Link        string `json:"link"`
	FinishTimes int64  `json:"finish_times"`
	State       int    `json:"state"`
}

// InviteTokenReply ...
type InviteTokenReply struct {
	Token string `json:"token"`
}

// CardTokenReply ...
type CardTokenReply struct {
	Token string `json:"token"`
}

// TaskReply ...
type TaskReply struct {
	List []*TaskDetail `json:"list"`
}

// SimpleTask ...
type SimpleTask struct {
	TaskName    string `json:"task_name"`
	LinkName    string `json:"link_name"`
	Desc        string `json:"desc"`
	Link        string `json:"link"`
	FinishTimes int64  `json:"finish_times"`
}

// TaskDetail 用户任务情况
type TaskDetail struct {
	Task   *SimpleTask `json:"task"`
	Member *TaskMember `json:"member"`
}

// TaskMember ...
type TaskMember struct {
	Counter string                 `json:"counter"`
	Count   int64                  `json:"count"`
	State   int                    `json:"state"`
	Params  map[string]interface{} `json:"params"`
}
