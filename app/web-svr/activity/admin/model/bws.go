package model

import (
	"time"
)

// ActBws def.
type ActBws struct {
	ID    int64     `json:"id" form:"id"`
	Name  string    `json:"name" form:"name"`
	Image string    `json:"image" form:"image"`
	Dic   string    `json:"dic" form:"dic"`
	Del   int8      `json:"del" form:"del"`
	Ctime time.Time `json:"ctime"`
	Mtime time.Time `json:"mtime"`
}

// ActBwsAchievement def.
type ActBwsAchievement struct {
	ID            int64     `json:"id" form:"id"`
	Name          string    `json:"name" form:"name"`
	Icon          string    `json:"icon" form:"icon"`
	Dic           string    `json:"dic" form:"dic"`
	Image         string    `json:"image" form:"image"`
	LinkType      int64     `json:"link_type" form:"link_type"`
	Unlock        int64     `json:"unlock" form:"unlock"`
	BID           int64     `json:"bid" gorm:"column:bid"  form:"bid"`
	IconBig       string    `json:"icon_big" form:"icon_big"`
	IconActive    string    `json:"icon_active" form:"icon_active"`
	IconActiveBig string    `json:"icon_active_big" form:"icon_active_big"`
	Award         int8      `json:"award" form:"award"`
	Ctime         time.Time `json:"ctime"`
	Mtime         time.Time `json:"mtime"`
	Del           int8      `json:"del"  form:"del"`
	SuitID        int64     `json:"suit_id" gorm:"column:suit_id"  form:"suit_id"`
	Level         int       `json:"level" gorm:"column:level"  form:"level"`
	AchievePoint  int       `json:"achieve_point" gorm:"column:achieve_point"  form:"achieve_point"`
	ExtraType     int64     `json:"extra_type" gorm:"column:extra_type"  form:"extra_type"`
}

// ActBwsField def.
type ActBwsField struct {
	ID    int64     `json:"id" form:"id"`
	Name  string    `json:"name" form:"name"`
	Area  string    `json:"area" form:"area"`
	BID   int64     `json:"bid" gorm:"column:bid"  form:"bid"`
	Del   int8      `json:"del" form:"del"`
	Ctime time.Time `json:"ctime"`
	Mtime time.Time `json:"mtime"`
}

// ActBwsPoint def.
type ActBwsPoint struct {
	ID           int64     `json:"id" form:"id"`
	Name         string    `json:"name" form:"name"`
	Icon         string    `json:"icon" form:"icon"`
	FID          int64     `json:"fid" gorm:"column:fid"  form:"fid"`
	Ower         int64     `json:"ower" gorm:"column:ower"  form:"ower"`
	Image        string    `json:"image" form:"image"`
	Unlocked     int64     `json:"unlocked" form:"unlocked"`
	LoseUnlocked int64     `json:"lose_unlocked" form:"lose_unlocked"`
	LockType     int64     `json:"lock_type" form:"lock_type"`
	Dic          string    `json:"dic" form:"dic"`
	Rule         string    `json:"rule" form:"rule"`
	BID          int64     `json:"bid" gorm:"column:bid"  form:"bid"`
	OtherIP      string    `json:"other_ip" gorm:"column:other_ip" form:"other_ip"`
	Del          int8      `json:"del" form:"del"`
	Ctime        time.Time `json:"ctime"`
	Mtime        time.Time `json:"mtime"`
}

// ActBwsPointArg .
type ActBwsPointArg struct {
	ActBwsPoint
	TotalPoint int64 `form:"total_point"`
	SignPoint  int64 `form:"sign_point"`
}

type ActBwsPointSign struct {
	Bid           int64 `gorm:"column:bid"`
	Pid           int64 `gorm:"column:pid"`
	Stime         int64 `gorm:"column:stime"`
	Etime         int64 `gorm:"column:etime"`
	State         int   `gorm:"column:state"`
	Points        int64 `gorm:"column:points"`
	ProvidePoints int64 `gorm:"column:provide_points"`
	SignPoints    int64 `gorm:"column:sign_points"`
	IsDelete      int   `gorm:"column:is_delete"`
}

// ActBwsPointResult .
type ActBwsPointResult struct {
	*ActBwsPoint
	TotalPoint    int64                     `json:"total_point"`
	ProvidePoints int64                     `json:"provide_points"`
	SignPoint     int64                     `json:"sign_point"`
	Level         []*ActBwsPointLevelResult `json:"level"`
}

// ActBwsPointLevelResult .
type ActBwsPointLevelResult struct {
	*ActBwsPointLevel
	Awards []*ActBwsPointAward `json:"awards"`
}

// ActBwsPointLevel .
type ActBwsPointLevel struct {
	ID       int64 `json:"id" form:"-"`
	Pid      int64 `json:"pid" form:"pid" validate:"min=1" gorm:"column:pid"`
	Level    int   `json:"level" form:"level" validate:"min=1" gorm:"column:level"`
	Points   int   `json:"points" form:"points" validate:"min=1" gorm:"column:points"`
	IsDelete int   `json:"is_delete" form:"is_delete" validate:"min=1" gorm:"column:is_delete"`
}

// ActBwsPointAward .
type ActBwsPointAward struct {
	PlID     int64  `json:"pl_id" form:"pl_id" validate:"min=1" gorm:"column:pl_id"`
	Name     string `json:"name" form:"name" validate:"min=1" gorm:"column:name"`
	Icon     string `json:"icon" form:"icon" validate:"min=1" gorm:"column:icon"`
	Points   int    `json:"points" form:"points" validate:"min=1" gorm:"column:points"`
	Amount   int    `json:"amount" form:"amount" validate:"min=1" gorm:"column:amount"`
	IsDelete int    `json:"is_delete" form:"is_delete" validate:"min=1" gorm:"column:is_delete"`
}

// ActBwsUserAchievement def.
type ActBwsUserAchievement struct {
	ID    int64     `json:"id" form:"id"`
	MID   int64     `json:"mid" gorm:"column:mid" form:"mid"`
	AID   int64     `json:"aid" gorm:"column:aid" form:"aid"`
	BID   int64     `json:"bid" gorm:"column:bid" form:"bid"`
	Key   string    `json:"key" form:"key"`
	Del   int8      `json:"del" form:"del"`
	Ctime time.Time `json:"ctime"`
	Mtime time.Time `json:"mtime"`
}

// ActBwsUsers def.
type ActBwsUsers struct {
	ID    int64     `json:"id" form:"id"`
	MID   int64     `json:"mid" gorm:"column:mid" form:"mid"`
	Key   string    `json:"key" form:"key"`
	BID   int64     `json:"bid" gorm:"column:bid"  form:"bid"`
	Del   int8      `json:"del" form:"del"`
	Ctime time.Time `json:"ctime"`
	Mtime time.Time `json:"mtime"`
}

// ActBwsVipUsers def.
type ActBwsVipUsers struct {
	ID    int64     `json:"id" form:"id"`
	MID   int64     `json:"mid" gorm:"column:mid" form:"mid"`
	Key   string    `json:"vip_key" form:"vip_key"`
	Day   string    `json:"bws_date" gorm:"column:bws_date" form:"bws_date"`
	BID   int64     `json:"bid" gorm:"column:bid"  form:"bid"`
	Del   int8      `json:"del" form:"del"`
	Ctime time.Time `json:"ctime"`
	Mtime time.Time `json:"mtime"`
}

// ActBwsUserPoint def.
type ActBwsUserPoint struct {
	ID     int64     `json:"id" form:"id"`
	MID    int64     `json:"mid" gorm:"column:mid"  form:"mid"`
	PID    int64     `json:"pid" gorm:"column:pid"  form:"pid"`
	BID    int64     `json:"bid" gorm:"column:bid"  form:"bid"`
	Key    string    `json:"key" form:"key"`
	Points int64     `json:"points" form:"points"`
	Del    int8      `json:"del" form:"del"`
	Ctime  time.Time `json:"ctime"`
	Mtime  time.Time `json:"mtime"`
}

// ActBwsUser def.
type ActBwsUser struct {
	ID    int64     `json:"id" form:"id"`
	MID   int64     `json:"mid" gorm:"column:mid"  form:"mid"`
	BID   int64     `json:"bid" gorm:"column:bid"  form:"bid"`
	Key   string    `json:"key" form:"key"`
	Del   int8      `json:"del" form:"del"`
	Ctime time.Time `json:"ctime"`
	Mtime time.Time `json:"mtime"`
}

type SaveBwsPointLevelParam struct {
	Levels []*ActBwsPointLevelParam `json:"levels"`
}

type ActBwsPointLevelParam struct {
	Level  int                      `json:"level" form:"level" validate:"min=1" gorm:"column:level"`
	Points int                      `json:"points" form:"points" validate:"min=1" gorm:"column:points"`
	Awards []*ActBwsPointAwardParam `json:"awards"`
}

type ActBwsPointAwardParam struct {
	Name   string `json:"name" form:"name" validate:"min=1" gorm:"column:name"`
	Icon   string `json:"icon" form:"icon" validate:"min=1" gorm:"column:icon"`
	Amount int    `json:"amount" form:"amount" validate:"min=1" gorm:"column:amount"`
}

type ActBwsAward struct {
	ID       int64     `form:"id" json:"id" gorm:"column:id"`
	Title    string    `form:"title" json:"title" gorm:"column:title"`
	Image    string    `form:"image" json:"image" gorm:"column:image"`
	Intro    string    `form:"intro" json:"intro" gorm:"column:intro"`
	Cate     string    `form:"cate" json:"cate" gorm:"column:cate"`
	IsOnline int64     `form:"is_online" json:"is_online" gorm:"column:is_online"`
	Stock    int64     `form:"stock" json:"stock" gorm:"column:stock"`
	OwnerMid int64     `form:"owner_mid" json:"owner_mid" gorm:"column:owner_mid"`
	Stage    string    `form:"stage" json:"stage" gorm:"column:stage"`
	State    int64     `form:"state" json:"state" gorm:"column:state"`
	Ctime    time.Time `json:"ctime" gorm:"column:ctime" time_format:"2006-01-02 15:04:05"`
	Mtime    time.Time `json:"mtime" gorm:"column:mtime" time_format:"2006-01-02 15:04:05"`
}

type AddActBwsAwardArg struct {
	Title    string `form:"title" validate:"required"`
	Image    string `form:"image" validate:"required"`
	Intro    string `form:"intro" validate:"required"`
	Cate     string `form:"cate" validate:"required"`
	IsOnline int64  `form:"is_online"`
	Stock    int64  `form:"stock" validate:"required"`
	OwnerMid int64  `form:"owner_mid" validate:"required"`
	Stage    string `form:"stage" validate:"required"`
	State    int64  `form:"state"`
}

type SaveActBwsAwardArg struct {
	ID int64 `form:"id" validate:"min=1" json:"id" gorm:"column:id"`
	AddActBwsAwardArg
}

type ActBwsTask struct {
	ID          int64     `form:"id" json:"id" gorm:"column:id"`
	Title       string    `form:"title" json:"title" gorm:"column:title"`
	Cate        string    `form:"cate" json:"cate" gorm:"column:cate"`
	FinishCount int64     `form:"finish_count" json:"finish_count" gorm:"column:finish_count"`
	RuleIDs     string    `form:"rule_ids" json:"rule_ids" gorm:"column:rule_ids"`
	OrderNum    int64     `form:"order_num" json:"order_num" gorm:"column:order_num"`
	State       int64     `form:"state" json:"state" gorm:"column:state"`
	Ctime       time.Time `json:"ctime" gorm:"column:ctime" time_format:"2006-01-02 15:04:05"`
	Mtime       time.Time `json:"mtime" gorm:"column:mtime" time_format:"2006-01-02 15:04:05"`
}

type AddActBwsTaskArg struct {
	Title       string `form:"title" validate:"required"`
	Cate        string `form:"cate" validate:"required"`
	FinishCount int64  `form:"finish_count" validate:"min=1"`
	RuleIDs     string `form:"rule_ids"`
	OrderNum    int64  `form:"order_num"`
	State       int64  `form:"state"`
}

type SaveActBwsTaskArg struct {
	ID int64 `form:"id" validate:"min=1"`
	AddActBwsTaskArg
}

// TableName ActBws def.
func (ActBws) TableName() string {
	return "act_bws"
}

func (ActBwsPointLevel) TableName() string {
	return "act_bws_points_level"
}

func (ActBwsPointAward) TableName() string {
	return "act_bws_points_award"
}

func (ActBwsPointSign) TableName() string {
	return "act_bws_point_sign"
}

func (ActBwsTask) TableName() string {
	return "act_bws_task"
}

func (ActBwsAward) TableName() string {
	return "act_bws_award"
}
