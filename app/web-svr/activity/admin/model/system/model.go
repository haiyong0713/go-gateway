package system

import (
	xtime "go-common/library/time"
)

const (
	SystemActStateNormal  = 0
	SystemActStateOffline = 1
	SystemActStateDelete  = 2

	SystemActivityTypeSign     = 1 // 签到活动
	SystemActivityTypeVote     = 2 // 投票活动
	SystemActivityTypeQuestion = 3 // 提问活动
)

const SystemActTypeSign = 1

type SystemSignStatistics struct {
	ID  int64  `gorm:"primaryKey"`
	AID int64  `gorm:"column:aid"`
	UID string `gorm:"column:uid"`
}

func (SystemSignStatistics) TableName() string {
	return "system_activity_sign_statistics"
}

type SystemSignUser struct {
	ID       int64      `gorm:"column:id"`
	AID      int64      `gorm:"column:aid"`
	UID      string     `gorm:"column:uid"`
	Location string     `gorm:"column:location"`
	Ctime    xtime.Time `gorm:"column:ctime"`
}

func (SystemSignUser) TableName() string {
	return "system_activity_sign"
}

type GetUsersInfo struct {
	Code    int                `json:"code"`
	Message string             `json:"message"`
	Data    []*UsersInfoDetail `json:"data"`
}

type UsersInfoDetail struct {
	UID            string `json:"uid"`
	NickName       string `json:"nick_name"`
	Avatar         string `json:"avatar"`
	DepartmentName string `json:"department_name"`
	LastName       string `json:"last_name"`
}

type GetSignListRes struct {
	Page ManagerPage    `json:"page"`
	List []*GetSignList `json:"list"`
}

type GetSignList struct {
	UID      string     `json:"uid"`
	Location string     `json:"location"`
	Time     xtime.Time `json:"time"`
	NickName string     `json:"nick_name"`
	LastName string     `json:"last_name"`
}

type SystemSignStatisticsList struct {
	AID int64  `gorm:"column:aid"`
	UID string `gorm:"column:uid"`
}

func (SystemSignStatisticsList) TableName() string {
	return "system_activity_sign_statistics"
}

type GetSignVipListRes struct {
	Count int64          `json:"count"`
	List  []*GetSignList `json:"list"`
}

type GetSignVipListDetailRes struct {
	Page ManagerPage          `json:"page"`
	List []*GetSignListDetail `json:"list"`
}

type GetSignListDetail struct {
	IsSign int `json:"is_sign"`
	GetSignList
}

type SystemActAddArgs struct {
	ID int64 `json:"id" gorm:"id"`
	SystemAct
}

type SystemActEditArgs struct {
	ID int64 `json:"id" form:"id" gorm:"id" validate:"required"`
	SystemAct
}

type SystemActInfo struct {
	ID int64 `json:"id" gorm:"id" validate:"required"`
	SystemAct
	State int64 `json:"state" gorm:"state" validate:"required"`
}

type SystemActInfoList struct {
	Page ManagerPage      `json:"page"`
	List []*SystemActInfo `json:"list"`
}

type ManagerPage struct {
	Num   int64 `json:"num"`
	Size  int64 `json:"size"`
	Total int64 `json:"total"`
}

type SystemAct struct {
	Name   string     `form:"name" gorm:"name" validate:"required" json:"name"`
	Stime  xtime.Time `form:"stime" time_format:"2006-01-02 15:04:05" gorm:"stime" validate:"required" json:"stime"`
	Etime  xtime.Time `form:"etime" time_format:"2006-01-02 15:04:05" gorm:"etime" validate:"required" json:"etime"`
	Type   int64      `form:"type" gorm:"type" validate:"required" json:"type"`
	Config string     `form:"config" gorm:"config" validate:"required" json:"config"`
	Create string     `form:"create" gorm:"create" json:"create"`
	Update string     `form:"update" gorm:"update" json:"update"`
}

func (SystemAct) TableName() string {
	return "system_activity"
}

type SystemActSignConfig struct {
	JumpURL  string `json:"jump_url"`
	Location int    `json:"location"`
	ShowSeat int    `json:"show_seat"`
}

type UIDSeat struct {
	UID     string `json:"uid"`
	AID     int64  `json:"aid"`
	Content string `json:"content"`
}

type SystemActSeat struct {
	ID      int64  `gorm:"id"`
	AID     int64  `gorm:"aid"`
	UID     string `gorm:"uid"`
	Content string `gorm:"content"`
}

func (SystemActSeat) TableName() string {
	return "system_activity_seat"
}

type GetSeatList struct {
	UID      string `json:"uid"`
	NickName string `json:"nick_name"`
	LastName string `json:"last_name"`
	Content  string `json:"content"`
}

type ActivitySystemVote struct {
	ID       int64  `json:"-" gorm:"id"`
	AID      int64  `json:"-" gorm:"aid"`
	UID      string `json:"-" gorm:"uid"`
	ItemID   int64  `json:"item_id" gorm:"item_id"`
	OptionID int64  `json:"option_id" gorm:"option_id"`
	Score    int64  `json:"score" gorm:"score"`
}

func (ActivitySystemVote) TableName() string {
	return "system_activity_vote"
}

type VoteDetail struct {
	UID      string `json:"uid"`
	NickName string `json:"nick_name"`
	LastName string `json:"last_name"`
}

type DetailUIDs struct {
	UID string `json:"uid"`
}

func (DetailUIDs) TableName() string {
	return "system_activity_vote"
}

type VoteConfig struct {
	Items []struct {
		Title   string `json:"title"`
		Type    int    `json:"type"`
		Options struct {
			Name []struct {
				Desc string `json:"desc"`
			} `json:"name"`
			LimitNum int `json:"limit_num"`
			Score    int `json:"score"`
		} `json:"options"`
	} `json:"items"`
}

type ActivitySystemQuestion struct {
	ID       int64      `json:"id" gorm:"id"`
	AID      int64      `json:"aid" gorm:"column:aid"`
	QID      int64      `json:"qid" gorm:"column:qid"`
	Question string     `json:"question" gorm:"column:question"`
	UID      string     `json:"uid" gorm:"column:uid"`
	State    int64      `json:"state" gorm:"column:state"`
	Ctime    xtime.Time `json:"ctime" gorm:"column:ctime"`
}

func (ActivitySystemQuestion) TableName() string {
	return "system_activity_question"
}

type ActivitySystemQuestionList struct {
	ActivitySystemQuestion
	NickName string `json:"nick_name"`
	LastName string `json:"last_name"`
}
