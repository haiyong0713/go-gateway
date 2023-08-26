package search_whitelist

import xtime "go-common/library/time"

const (
	StatusDefault    = 0
	StatusWaitAudit1 = 1
	StatusWaitAudit2 = 2
	StatusWaitOnline = 3
	StatusOnline     = 4
	StatusOffline    = 5
	StatusReject     = 6
)

type WhiteListItemWithQueryAndArchive struct {
	*WhiteListItem
	SearchWord []string                `json:"search_word" form:"search_word"`
	Archive    []*WhiteListArchiveItem `json:"archive" form:"archive"`
}

type WhiteListItem struct {
	ID        int64      `json:"id" form:"id" gorm:"column:id"`
	STime     xtime.Time `json:"stime" form:"stime" gorm:"column:stime"`
	ETime     xtime.Time `json:"etime" form:"etime" gorm:"column:etime"`
	CUser     string     `json:"c_user" form:"c_user" gorm:"column:c_user"`
	RoleId    int64      `json:"role_id" form:"role_id" gorm:"column:role_id"`
	Pid       int64      `json:"pid" form:"pid" gorm:"column:pid"`
	Hidden    int        `json:"hidden" form:"hidden" gorm:"column:hidden"`
	Status    int        `json:"status" form:"status" gorm:"column:status"`
	IsDeleted int        `json:"-" form:"-" gorm:"column:is_deleted"`
}

func (*WhiteListItem) TableName() string {
	return "search_whitelist"
}

type WhiteListQueryItem struct {
	ID         int64  `json:"-" form:"-" gorm:"column:id"`
	Pid        int64  `json:"-" form:"-" gorm:"column:pid"`
	SearchWord string `json:"search_word" form:"search_word" gorm:"column:search_word"`
	IsDeleted  int    `json:"-" form:"-" gorm:"column:is_deleted"`
}

func (*WhiteListQueryItem) TableName() string {
	return "search_whitelist_query"
}

type WhiteListArchiveItem struct {
	ID        int64  `json:"-" form:"-" gorm:"column:id"`
	Avid      int64  `json:"avid" form:"avid" gorm:"column:avid"`
	CardType  int    `json:"card_type" form:"card_type" gorm:"-"`
	Title     string `json:"title" form:"title" gorm:"-"`
	Cover     string `json:"cover" form:"cover" gorm:"-"`
	Pid       int64  `json:"pid" form:"pid" gorm:"column:pid"`
	Rank      int64  `json:"rank" form:"rank" gorm:"column:rank"`
	IsDeleted int    `json:"-" form:"-" gorm:"column:is_deleted"`
}

func (*WhiteListArchiveItem) TableName() string {
	return "search_whitelist_archive"
}
