package model

import (
	"strconv"
	"strings"
)

const (
	TypeProgress      = "progress"
	TypeClickProgress = "click_progress"
	ClickTypProgress  = 30
	Online            = 1
	CategoryProgress  = 28
	CategoryTab       = 18
	CategorySelect    = 27
	MixTypeTab        = 4
	// counter dim01：数据统计维度
	ProgDimPointTotal = ProgressDimension(1) //总积分
	ProgDimUserTotal  = ProgressDimension(2) //总人数
	ProgDimUser       = ProgressDimension(3) //个人分数
)

type ProgressParam struct {
	ID        int64  `json:"id"`
	PageID    int64  `json:"page_id"`
	GroupID   int64  `json:"group_id"`
	Sid       int64  `json:"sid"`
	Dimension int64  `json:"dimension"`
	WebKey    string `json:"web_key"`
	Type      string `json:"type"`
}

type ClickProgressParam struct {
	ID        int64  `json:"id"`
	ModuleID  int64  `json:"module_id"`
	GroupID   int64  `json:"group_id"`
	Sid       int64  `json:"sid"`
	Dimension string `json:"dimension"`
	WebKey    string `json:"web_key"`
}

type ChildPage struct {
	ModuleID    int64 `json:"module_id"`
	ChildPageID int64 `json:"child_page_id"`
}

type ParentPage struct {
	ModuleID     int64 `json:"module_id"`
	ParentPageID int64 `json:"parent_page_id"`
}

type PointMsg struct {
	// 活动id
	Activity string `json:"activity"`
	// $sid_$groupID_$dim01_$dim02_TEMPLATE
	Counter string `json:"counter"`
	// 用户id
	Mid int64 `json:"mid"`
	// 时间戳
	Timestamp int64 `json:"timestamp"`
	// 用户本次加分值
	Diff int64 `json:"diff"`
	// 用户本次加分后总值
	Total int64 `json:"total"`
}

type ProgressDimension int64

func (pd ProgressDimension) IsUser() bool {
	return pd == ProgDimUser
}

func (pd ProgressDimension) IsTotal() bool {
	return pd == ProgDimUserTotal || pd == ProgDimPointTotal
}

func ProgressStatString(number int64) string {
	if number == 0 {
		return "0"
	}
	return StatString(number, "")
}

// nolint:gomnd
func StatString(number int64, suffix string) (s string) {
	if number == 0 {
		s = "-" + suffix
		return
	}
	if number < 10000 {
		s = strconv.FormatInt(number, 10) + suffix
		return
	}
	if number < 100000000 {
		s = strconv.FormatFloat(float64(number)/10000, 'f', 1, 64)
		return strings.TrimSuffix(s, ".0") + "万" + suffix
	}
	s = strconv.FormatFloat(float64(number)/100000000, 'f', 1, 64)
	return strings.TrimSuffix(s, ".0") + "亿" + suffix
}
