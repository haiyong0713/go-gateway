package newstar

import (
	xtime "go-common/library/time"
)

// Newstar .
type Newstar struct {
	ID            int64      `json:"id"`
	ActivityUID   string     `json:"activity_uid"`
	VStatus       int64      `json:"v_status"`
	Mid           int64      `json:"mid"`
	InviterMid    int64      `json:"inviter_mid"`
	IsName        int64      `json:"is_name"`
	IsMobile      int64      `json:"is_mobile"`
	IsIdentity    int64      `json:"is_identity"`
	FansCount     int64      `json:"fans_count"`
	UpArchives    int64      `json:"up_archives"`
	FinishTask    int64      `json:"finish_task"`
	FinishTime    int64      `json:"finish_time"`
	Ctime         xtime.Time `json:"ctime"`
	ReceiveAward  int64      `json:"receive_award"`
	RemainingDays int64      `json:"remaining_days"`
}

// NewstarAward .
type NewstarAward struct {
	ID          int64  `json:"id"`
	ActivityUID string `json:"activity_uid"`
	AwardType   int64  `json:"award_type"`
	Condition   int64  `json:"condition"`
	FinishMoney int64  `json:"finish_money"`
	InviteMoney int64  `json:"invite_money"`
}

// NewstarInvite .
type NewstarInvite struct {
	VStatus     int64       `json:"v_status"`
	Mid         int64       `json:"mid"`
	Name        string      `json:"name"`
	Face        string      `json:"face"`
	InviteAward int64       `json:"invite_award"`
	List        []*UserInfo `json:"list"`
	Page        *Page       `json:"page"`
}

// UserInfo .
type UserInfo struct {
	ID            int64      `json:"id"`
	Mid           int64      `json:"mid"`
	Name          string     `json:"name"`
	Face          string     `json:"face"`
	BaseTask      bool       `json:"base_task"`
	FansCount     int64      `json:"fans_count"`
	RemainingDays int64      `json:"remaining_days"`
	Ctime         xtime.Time `json:"ctime"`
}

// Page .
type Page struct {
	Num   int `json:"num"`
	Size  int `json:"size"`
	Total int `json:"total"`
}
