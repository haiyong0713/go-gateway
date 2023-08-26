package college

import (
	"sync"

	go_common_library_time "go-common/library/time"
	"go-gateway/app/web-svr/activity/interface/model/rank"
)

const (
	// MidTypeIsNew 新用户
	MidTypeIsNew = 1
	// MidTypeIsOld 老用户
	MidTypeIsOld = 2
	// NationWideRankType 全国排行
	NationWideRankType = 0
	// FollowKey ...
	FollowKey = "follow"
	// VideoupKey ...
	VideoupKey = "videoup"
	// BonusKey ...
	BonusKey = "bonus"
	// InviteKey ...
	InviteKey = "invite"
	// ViewKey ...
	ViewKey = "view"
	// LikeKey ...
	LikeKey = "like"
	// ShareKey ...
	ShareKey = "share"
	// MidTypeIsNewStr 新用户
	MidTypeIsNewStr = "new"
	// MidTypeIsAllStr 全部
	MidTypeIsAllStr = "member"
	// TaskFollow 任务关注
	TaskFollow = 1
	// TaskArchive 任务投稿
	TaskArchive = 2
	// TaskInvite 任务邀请
	TaskInvite = 3
	// TaskView 任务观看
	TaskView = 4
	// TaskLike 任务点赞
	TaskLike = 5
	// TaskShare 任务分享
	TaskShare = 6
)

// BindReply ...
type BindReply struct {
	College *College `json:"college"`
}

// TaskReply ...
type TaskReply struct {
	TaskList      []*Task `json:"task_list"`
	PersonalScore int64   `json:"personal_score"`
	TabList       []int64 `json:"tab_list"`
}

// Task ...
type Task struct {
	Type   int                    `json:"type"`
	Params map[string]interface{} `json:"params"`
	State  map[string]int64       `json:"state"`
}

// College ...
type College struct {
	ID         int64  `json:"id"`
	Name       string `json:"name"`
	ProvinceID int64  `json:"province_id"`
}

// Version ...
type Version struct {
	Version int   `json:"version"`
	Time    int64 `json:"time"`
}

// Personal 个人信息
type Personal struct {
	MID       int64 `json:"mid"`
	Score     int64 `json:"score"`
	Rank      int   `json:"rank"`
	Diff      int64 `json:"diff"`
	CollegeID int64 `json:"college_id"`
}

// InviterCollegeReply ...
type InviterCollegeReply struct {
	College *College `json:"college"`
	Account *Account `json:"account"`
}

// PersonalCollege ...
type PersonalCollege struct {
	MID         int64  `json:"mid"`
	CollegeID   int64  `json:"college_id"`
	CollegeName string `json:"name"`
}

// Province 省信息
type Province struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

// Detail ...
type Detail struct {
	ID             int64   `json:"id"`
	Name           string  `json:"college_name"`
	ProvinceID     int64   `json:"province_id"`
	Province       string  `json:"province"`
	White          []int64 `json:"white"`
	MID            int64   `json:"mid"`
	RelationMid    []int64 `json:"relation_mid"`
	Score          int64   `json:"score"`
	Aids           []int64 `json:"aids"`
	TabList        []int64 `json:"tab_list"`
	TagID          int64   `json:"tag_id"`
	Initial        string  `json:"initial"`
	NationwideRank int     `json:"nationwide_rank"`
	ProvinceRank   int     `json:"province_rank"`
}

// DetailReply ...
type DetailReply struct {
	College *DetailCollege `json:"college"`
}

// DetailCollege ...
type DetailCollege struct {
	Score      int64   `json:"score"`
	Nationwide string  `json:"nationwide"`
	Province   string  `json:"province"`
	TabList    []int64 `json:"tab_list"`
	Name       string  `json:"name"`
	ID         int64   `json:"id"`
}

// ProvinceCollegeRank ...
type ProvinceCollegeRank struct {
	Info map[int64][]*rank.Redis
	lock sync.RWMutex
}

// Init ...
func (m *ProvinceCollegeRank) Init() {
	m.Info = make(map[int64][]*rank.Redis)
}

// Set ...
func (m *ProvinceCollegeRank) Set(key int64, value []*rank.Redis) {
	m.lock.Lock()
	defer m.lock.Unlock()
	m.Info[key] = value
}

// Get ...
func (m *ProvinceCollegeRank) Get(key int64) []*rank.Redis {
	m.lock.Lock()
	defer m.lock.Unlock()
	data, ok := m.Info[key]
	if ok {
		return data
	}
	return nil
}

// Base ...
type Base struct {
	ID      int64  `json:"id"`
	Name    string `json:"name"`
	Initial string `json:"initial"`
}

// DB ...
type DB struct {
	ID          int64  `json:"id"`
	TagID       int64  `json:"tag_id"`
	Name        string `json:"college_name"`
	ProvinceID  int64  `json:"province_id"`
	Province    string `json:"province"`
	White       string `json:"white"`
	MID         int64  `json:"mid"`
	RelationMid string `json:"relation_mid"`
	Initial     string `json:"initial"`
}

// AllCollegeReply ...
type AllCollegeReply struct {
	College []*Base `json:"college_list"`
}

// ProvinceCollegeRankReply ...
type ProvinceCollegeRankReply struct {
	CollegeList []*RankReply `json:"college_list"`
	Page        *Page        `json:"page"`
	Province    *Province    `json:"province"`
	Time        int64        `json:"time"`
}

// RankReply ...
type RankReply struct {
	ID       int64      `json:"id"`
	Name     string     `json:"name"`
	Initial  string     `json:"initial"`
	Score    int64      `json:"score"`
	Province string     `json:"province"`
	Archive  []*Archive `json:"archive"`
}

// Page ...
type Page struct {
	Num   int `json:"num"`
	Size  int `json:"size"`
	Total int `json:"total"`
}

// Archive ...
type Archive struct {
	Bvid     string                      `json:"bvid"`
	TypeName string                      `json:"tname"`
	Title    string                      `json:"title"`
	Desc     string                      `json:"desc"`
	Duration int64                       `json:"duration"`
	Pic      string                      `json:"pic"`
	View     int32                       `json:"view"`
	Author   *Author                     `json:"author"`
	Ctime    go_common_library_time.Time `json:"ctime"`
	Danmaku  int32                       `json:"danmu"`
}

// Author 稿件作者信息
type Author struct {
	// Up主mid
	Mid int64 `json:"mid"`
	// Up主名称
	Name string `json:"name"`
	// Up主头像地址 绝对地址
	Face string `json:"face"`
}

// Account ...
type Account struct {
	Mid  int64  `json:"mid"`
	Name string `json:"name"`
	Face string `json:"face"`
	Sign string `json:"sign"`
	Sex  string `json:"sex"`
}

// PeopleRankReply ...
type PeopleRankReply struct {
	MemberList []*MemberReply `json:"member_list"`
	Page       *Page          `json:"page"`
	Time       int64          `json:"time"`
}

// MemberReply ...
type MemberReply struct {
	Account *Account `json:"account"`
	Score   int64    `json:"score"`
}

// PersonalReply ...
type PersonalReply struct {
	Account *Account `json:"account"`
	Score   int64    `json:"score"`
	Rank    int      `json:"rank"`
	Diff    int64    `json:"diff"`
	College *College `json:"college"`
}

// ArchiveListReply 稿件列表
type ArchiveListReply struct {
	ArchiveInfo []*ArchiveInfo `json:"archive_list"`
	Page        *Page          `json:"page"`
}

// ArchiveInfo ...
type ArchiveInfo struct {
	Archive    *Archive `json:"archive"`
	IsFollower int      `json:"is_follower"`
}

// FollowReply ...
type FollowReply struct {
}

// ActPlatActivityPoints ...
type ActPlatActivityPoints struct {
	Points    int64  `json:"points"`    // 积分增减值
	Timestamp int64  `json:"timestamp"` // 事件发生的时间戳
	Mid       int64  `json:"mid"`
	Source    int64  `json:"source"`   // 积分原因，一般是关联的资源id，例如关注的up主id，邀请的用户id
	Activity  string `json:"activity"` // 关联活动名，开学季活动此处填 college2020
	Business  string `json:"business"` // 加分相关业务名，关注：follow，邀请：invite，投稿额外加分：bonus
	Extra     string `json:"extra"`    // 保留字段
}
