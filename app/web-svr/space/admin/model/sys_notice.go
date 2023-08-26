package model

import (
	"fmt"
	"strings"

	xtime "go-common/library/time"
)

const (
	_UIDInsertSQL = "INSERT INTO system_notice_uid(system_notice_id,uid) VALUES %s"
)

var (
	ScopeDict = map[int]string{
		NOTICE_SCOPE_SPACE:  "UP主空间卡",
		NOTICE_SCOPE_SEARCH: "搜索UP主卡",
	}
)

const (
	_ int64 = iota
	NOTICE_STATUS_ONLINE
	NOTICE_STATUS_OFFLINE
)

const (
	_ int = iota
	NOTICE_SCOPE_SPACE
	NOTICE_SCOPE_SEARCH
)

type SysNotice struct {
	ID         int64      `form:"id" json:"id"`
	Content    string     `form:"content" json:"content"`
	NoticeType int        `json:"notice_type" form:"notice_type"`
	Scopes     string     `json:"scopes" form:"scopes"`
	Url        string     `form:"url" json:"url"`
	Status     int64      `form:"status" json:"status"`
	UidCount   int64      `form:"-" json:"uid_count"`
	Ctime      xtime.Time `form:"ctime" json:"ctime"`
	Mtime      xtime.Time `form:"mtime" json:"mtime"`
}

type SysNoticeList struct {
	Uid    int64 `form:"uid" json:"uid"`
	Scopes []int `form:"scopes,split" json:"scopes"`
	Status int64 `form:"status" json:"status"`
	Ps     int   `form:"ps" json:"ps" default:"20"` // 分页大小
	Pn     int   `form:"pn" json:"pn" default:"1"`  // 第几个分页
}

type SysNoticeInfo struct {
	ID         int64   `json:"id"`
	Content    string  `json:"content"`
	NoticeType int     `json:"notice_type"`
	Scopes     string  `json:"scopes"`
	Url        string  `json:"url"`
	Status     int64   `json:"status"`
	UidCount   int64   `json:"uid_count"`
	Uids       []int64 `json:"uids"`
}

type SysNoticeAdd struct {
	ID         int64  `form:"id" json:"id"`
	Content    string `form:"content" json:"content" validate:"required"`
	NoticeType int    `json:"notice_type" form:"notice_type" validate:"required"`
	Scopes     string `json:"scopes" form:"scopes" validate:"required"`
	Url        string `form:"url" json:"url"`
}

type SysNoticeUp struct {
	ID         int64  `form:"id" json:"id" validate:"required"`
	NoticeType int    `json:"notice_type" form:"notice_type" validate:"required"`
	Content    string `form:"content" json:"content" validate:"required"`
	Scopes     string `json:"scopes" form:"scopes" validate:"required"`
	Url        string `form:"url" json:"url"`
	UidCount   int64  `form:"uid_count" json:"uid_count"`
}

type SysNoticeOpt struct {
	ID     int64 `form:"id" json:"id" validate:"required"`
	Status int64 `form:"status" json:"status" validate:"required"`
}

func (a SysNotice) TableName() string {
	return "system_notice"
}

func (a SysNoticeAdd) TableName() string {
	return "system_notice"
}

func (a SysNoticeUp) TableName() string {
	return "system_notice"
}

func (a SysNoticeOpt) TableName() string {
	return "system_notice"
}

func (a SysNoticeUid) TableName() string {
	return "system_notice_uid"
}

type SysNotUidAddDel struct {
	ID   int64   `form:"id" validate:"required"`
	UIDs []int64 `form:"uids,split" validate:"required,dive,gt=0"`
}

type SysNoticeUidParam struct {
	ID int64 `form:"id" validate:"required"`
	Ps int   `form:"ps" default:"20"` // 分页大小
	Pn int   `form:"pn" default:"1"`  // 第几个分页
}

type SysNoticeUid struct {
	ID             int64      `form:"id" json:"id"`
	SystemNoticeId int64      `form:"system_notice_id" json:"system_notice_id"`
	Uid            int64      `form:"uid" json:"uid"`
	IsDeleted      int64      `form:"is_deleted" json:"is_deleted"`
	Ctime          xtime.Time `form:"ctime" json:"ctime"`
	Mtime          xtime.Time `form:"mtime" json:"mtime"`
}

type SysNoticeUidWithScope struct {
	ID             int64  `json:"id"`
	SystemNoticeId int64  `json:"system_notice_id"`
	Uid            int64  `json:"uid"`
	IsDeleted      int64  `json:"is_deleted"`
	Scopes         string `json:"scopes"`
}

func (a SysNoticeUidWithScope) TableName() string {
	return "system_notice_uid"
}

func BatchAddUIDSQL(ID int64, data []int64) (sql string, param []interface{}) {
	if len(data) == 0 {
		return "", []interface{}{}
	}
	var rowStrings []string
	for _, v := range data {
		rowStrings = append(rowStrings, "(?,?)")
		param = append(param, ID, v)
	}
	return fmt.Sprintf(_UIDInsertSQL, strings.Join(rowStrings, ",")), param
}

// SysNoticePager system notice pager
type SysNoticePager struct {
	Item []*SysNoticeUid `json:"item"`
	Page Page            `json:"page"`
}

// SysNoticeInfoPager system notice info pager
type SysNoticeInfoPager struct {
	Item []*SysNotice `json:"item"`
	Page Page         `json:"page"`
}
