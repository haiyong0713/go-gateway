package summer_camp

import (
	go_common_library_time "go-common/library/time"
	"go-gateway/app/web-svr/activity/interface/model/risk"
)

const (
	Scene                = "summer_camp_obtain_bonus"
	RiskActionExchange   = "obtain_bonus"
	RiskSubsceneExchange = "obtain_bonus"
)

type SCRiskParams struct {
	*risk.Base
	Subscene string `json:"subscene"`
}

// UserInfoRes summerCampUserInfo接口的返回值
type UserInfoRes struct {
	IsJoin   int       `json:"is_join"`
	SignDays int64     `json:"sign_days"`
	UserInfo *UserInfo `json:"user_info"`
	TaskInfo *TaskInfo `json:"task_info"`
}

type UserInfo struct {
	NickName string `json:"nickname"`
	Mid      int64  `json:"mid"`
	Face     string `json:"face"`
}

type TaskInfo struct {
	TotalPoint     int64 `json:"total_point"`
	ViewVideosTask int64 `json:"view_videos_task"`
	ShareTask      int64 `json:"share_task"`
	TougaoTask     int64 `json:"tougao_task"`
}

// DBCourseCamp db字段
type DBCourseCamp struct {
	CourseID    int64                       `json:"course_id"`
	CourseTitle string                      `json:"course_title"`
	PicCover    string                      `json:"pic_cover"`
	BodanId     string                      `json:"bodan_id"`
	Creator     string                      `json:"creator"`
	Status      int                         `json:"status"`
	Ctime       go_common_library_time.Time `json:"ctime"`
	Mtime       go_common_library_time.Time `json:"ctime"`
}

// CourseListRes 课程列表返回
type CourseListRes struct {
	List  []*OneCourseRes `json:"list"`
	Page  int             `json:"page"`
	Size  int             `json:"size"`
	Total int             `json:"total"`
}

type OneCourseRes struct {
	ID         int64  `json:"id"`
	CourseName string `json:"course_name"`
	CourseIcon string `json:"course_icon"`
	BodanId    string `json:"bodan_id"`
}

// DBUserCourse db字段
type DBUserCourse struct {
	ID          int64                       `json:"id"`
	Mid         int64                       `json:"mid"`
	CourseID    int64                       `json:"course_id"`
	CourseTitle string                      `json:"course_title"`
	Status      int                         `json:"status"`
	JoinTime    go_common_library_time.Time `json:"join_time"`
	Ctime       go_common_library_time.Time `json:"ctime"`
	Mtime       go_common_library_time.Time `json:"ctime"`
}

// UserCourseInfoRes ...
type UserCourseInfoRes struct {
	List  []*UserCourseSignInfo `json:"list"`
	Page  int                   `json:"page"`
	Size  int                   `json:"size"`
	Total int                   `json:"total"`
}

type UserCourseSignInfo struct {
	CourseId   int64                       `json:"id"`
	CourseName string                      `json:"course_name"`
	CourseIcon string                      `json:"course_icon"`
	JoinTime   go_common_library_time.Time `json:"join_time"`
	CourseDays int                         `json:"course_days"`
	SignedDays int                         `json:"signed_days"`
	IsFinished bool                        `json:"is_finished"`
}

// CourseBodanList 每日课程视频列表
type CourseBodanList struct {
	List       []*VideoDetail `json:"list"`
	BodanTitle string         `json:"bodan_title"`
	BodanDesc  string         `json:"bodan_desc"`
	TabList    []*TabBodan    `json:"tab_list,omitempty"`
	Page       int            `json:"page"`
	Size       int            `json:"size"`
	Total      int            `json:"total"`
}

type TabBodan struct {
	Title string `json:"title"`
	Desc  string `json:"desc"`
}

type VideoDetail struct {
	Bvid     string   `json:"bvid"`
	Title    string   `json:"title"`
	Duration int64    `json:"duration"`
	Cover    string   `json:"cover"`
	CntInfo  VideoCnt `json:"cnt_info"`
	Link     string   `json:"link"`
}

type VideoCnt struct {
	Collect int32 `json:"collect"`
	Play    int32 `json:"play"`
	Danmaku int32 `json:"danmaku"`
}

// 开启计划返回
type StartPlanRes struct {
	RewardPoint int `json:"reward_point" default:"100"`
}

// UserPointHistoryRes
type UserPointHistoryRes struct {
	List  []*PointInfo `json:"list"`
	Page  int          `json:"page"`
	Size  int          `json:"size"`
	Total int          `json:"total"`
}

// PointInfo
type PointInfo struct {
	Tim        int64  `json:"time"`
	Type       int    `json:"type"`
	RewardName string `json:"reward_name"`
	Point      int64  `json:"point"`
	Left       int64  `json:"left"`
}

// AwardListRes
type AwardListRes struct {
	List  []*AwardInfo `json:"list"`
	Page  int          `json:"page"`
	Size  int          `json:"size"`
	Total int          `json:"total"`
}

type AwardInfo struct {
	AwardId         string `json:"award_id"`
	AwardName       string `json:"award_name"`
	AwardCost       int64  `json:"award_cost"`
	StockLeft       int64  `json:"stock_left"`
	Icon            string `json:"icon"`
	UserCanExchange bool   `json:"user_can_exchange"`
}
