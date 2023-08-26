package model

import (
	xtime "go-common/library/time"
	"go-gateway/app/web-svr/activity/admin/model/stime"
)

// VIDEO actiivty types .
const (
	VIDEO                = 1
	PICTURE              = 2
	DRAWYOO              = 3
	VIDEOLIKE            = 4
	PICTURELIKE          = 5
	DRAWYOOLIKE          = 6
	TEXT                 = 7
	TEXTLIKE             = 8
	ONLINEVOTE           = 9
	QUESTION             = 10
	LOTTERY              = 11
	ARTICLE              = 12
	VIDEO2               = 13
	MUSIC                = 15
	PHONEVIDEO           = 16
	SMALLVIDEO           = 17
	RESERVATION          = 18
	MISSIONGROUP         = 19
	CLOCKIN              = 22
	USERACTIONSTAT       = 23
	SubRuleOffline       = 0
	SubRuleOnline        = 1
	SubRuleLock          = 2
	SubRuleDelete        = 3
	RuleAttrBitCountType = 0
	AttrYes              = 1
	SubOnLine            = 1
	SubOffLine           = -1
	// TagTypeUp uptag
	TagTypeUp = 1
	// TagTypeActivity 活动tag
	TagTypeActivity   = 4
	FLAGRESERVEPUSH   = uint(26)
	FLAGQUESTIONNAIRE = uint(27)
)

// SidSub def
type SidSub struct {
	Type int     `form:"type" validate:"required"`
	Lids []int64 `form:"lids,split" validate:"max=50,min=1,dive,min=1"`
}

// ListSub def
type ListSub struct {
	Page     int     `form:"page" default:"1" validate:"min=1"`
	PageSize int     `form:"pagesize" default:"15" validate:"min=1"`
	Keyword  string  `form:"keyword"`
	States   []int   `form:"state,split" default:"0"`
	Types    []int   `form:"type,split" default:"0"`
	Sctime   int64   `form:"sctime"`
	Ectime   int64   `form:"ectime"`
	IDs      []int64 `form:"ids"`
	Name     string  `form:"name"`
}

// SubListRes .
type SubListRes struct {
	List     []*ActSubject `json:"list"`
	Page     int           `json:"page"`
	PageSize int           `json:"pagesize"`
	Count    int64         `json:"count"`
}

// PageRes .
type PageRes struct {
	Num   int   `json:"num"`
	Size  int   `json:"size"`
	Total int64 `json:"total"`
}

// AddList def
type AddList struct {
	// ActSubject
	ID                   int64  `json:"id,omitempty" form:"id" gorm:"column:id"`
	Oid                  int64  `json:"oid,omitempty" form:"oid"`
	Type                 int    `json:"type,omitempty" form:"type"`
	State                int    `json:"state,omitempty" form:"state"`
	Level                int    `json:"level,omitempty" form:"level"`
	Flag                 int64  `json:"flag,omitempty" form:"flag"`
	Rank                 int64  `json:"rank,omitempty" form:"rank"`
	Stime                string `json:"stime,omitempty" form:"stime" time_format:"2006-01-02 15:04:05"`
	Etime                string `json:"etime,omitempty" form:"etime" time_format:"2006-01-02 15:04:05"`
	Ctime                string `json:"ctime,omitempty" form:"ctime" time_format:"2006-01-02 15:04:05"`
	Mtime                string `json:"mtime,omitempty" form:"mtime" time_format:"2006-01-02 15:04:05"`
	Lstime               string `json:"lstime,omitempty" form:"lstime" time_format:"2006-01-02 15:04:05"`
	Letime               string `json:"letime,omitempty" form:"letime" time_format:"2006-01-02 15:04:05"`
	Uetime               string `json:"uetime,omitempty" form:"uetime" time_format:"2006-01-02 15:04:05"`
	Ustime               string `json:"ustime,omitempty" form:"ustime" time_format:"2006-01-02 15:04:05"`
	Name                 string `json:"name,omitempty" form:"name" validate:"lt=64"`
	Author               string `json:"author,omitempty" form:"author" validate:"lt=64"`
	ActURL               string `json:"act_url,omitempty" form:"act_url" validate:"lt=255"`
	Cover                string `json:"cover,omitempty" form:"cover" validate:"lt=255"`
	Dic                  string `json:"dic,omitempty" form:"dic" validate:"lt=64"`
	H5Cover              string `json:"h5_cover,omitempty" form:"h5_cover" validate:"lt=255"`
	LikeLimit            int    `json:"like_limit" form:"like_limit"`
	AndroidURL           string `json:"android_url" form:"android_url" validate:"lt=255"`
	IosURL               string `json:"ios_url" form:"ios_url" validate:"lt=255"`
	DailyLikeLimit       int64  `json:"daily_like_limit" form:"daily_like_limit"`
	DailySingleLikeLimit int64  `json:"daily_single_like_limit" form:"daily_single_like_limit"`
	UpLevel              int64  `json:"up_level" form:"up_level"`
	UpScore              int64  `json:"up_score" form:"up_score"`
	UpUetime             string `json:"up_uetime" form:"up_uetime" time_format:"2006-01-02 15:04:05"`
	UpUstime             string `json:"up_ustime" form:"up_ustime" time_format:"2006-01-02 15:04:05"`
	FanLimitMax          int64  `json:"fan_limit_max" form:"fan_limit_max"`
	FanLimitMin          int64  `json:"fan_limit_min" form:"fan_limit_min"`
	ChildSids            string `json:"child_sids" form:"child_sids" validate:"lt=255"`
	MonthScore           int    `json:"month_score" form:"month_score"`
	YearScore            int    `json:"year_score" form:"year_score"`
	UpFigureScore        int    `json:"up_figure_score" form:"up_figure_score"`
	ForbidenTime         string `json:"forbiden_time" form:"forbiden_time" time_format:"2006-01-02 15:04:05"`
	Contacts             string `json:"contacts" form:"contacts" validate:"lt=255"`
	ShieldFlag           int64  `json:"shield_flag" form:"shield_flag"`
	RelationID           int64  `json:"relation_id" form:"relation_id"`
	// ActSubject

	Protocol        string `form:"protocol"`
	Types           string `form:"types" validate:"lt=512"`
	Pubtime         string `form:"pubtime" time_format:"2006-01-02 15:04:05"`
	Deltime         string `form:"deltime" time_format:"2006-01-02 15:04:05"`
	Editime         string `form:"editime" time_format:"2006-01-02 15:04:05"`
	Tags            string `form:"tags" validate:"lt=255"`
	Interval        int    `form:"interval"`
	Tlimit          int    `form:"tlimit"`
	Ltime           int    `form:"ltime"`
	Hot             int    `form:"hot"`
	BgmID           int64  `form:"bgm_id"`
	PasterID        int64  `form:"paster_id"`
	Oids            string `form:"oids" validate:"lt=64"`
	ScreenSet       int    `form:"screen_set" default:"1"`
	AwardUrl        string `form:"award_url" validate:"lt=255"`
	Award           string `form:"award" validate:"lt=64"`
	InstepID        int    `form:"instep_id"`
	PriorityRegion  string `form:"priority_region" validate:"lt=255"`
	RegionWeight    int    `form:"region_weight"`
	GlobalWeight    int    `form:"global_weight"`
	WeightStime     string `form:"weight_stime" time_format:"2006-01-02 15:04:05"`
	WeightEtime     string `form:"weight_etime" time_format:"2006-01-02 15:04:05"`
	TagShowPlatform int    `json:"tag_show_platform" form:"tag_show_platform"`

	// other
	AuditLevel int `json:"audit_level" form:"audit_level"`

	// 扩展字段
	Calendar      string `json:"calendar" form:"calendar"`
	AuditPlatform string `json:"audit_platform" form:"audit_platform"`
}

// ActSubjectProtocol def
type ActSubjectProtocol struct {
	ID              int64      `json:"id" form:"id" gorm:"column:id"`
	Sid             int64      `json:"sid" form:"sid"`
	Protocol        string     `json:"protocol" form:"protocol"`
	Mtime           stime.Time `json:"mtime" form:"mtime" time_format:"2006-01-02 15:04:05"`
	Ctime           stime.Time `json:"ctime" form:"ctime" time_format:"2006-01-02 15:04:05"`
	Types           string     `json:"types" form:"types"`
	Tags            string     `json:"tags" form:"tags"`
	Hot             int        `json:"hot" form:"hot"`
	Pubtime         stime.Time `json:"pubtime" form:"pubtime" time_format:"2006-01-02 15:04:05"`
	Deltime         stime.Time `json:"deltime" form:"deltime" time_format:"2006-01-02 15:04:05"`
	Editime         stime.Time `json:"editime" form:"editime" time_format:"2006-01-02 15:04:05"`
	BgmID           int64      `json:"bgm_id" form:"bgm_id" gorm:"column:bgm_id"`
	PasterID        int64      `json:"paster_id" form:"paster_id" gorm:"column:paster_id"`
	Oids            string     `json:"oids" form:"oids" gorm:"column:oids"`
	AwardUrl        string     `json:"award_url" form:"award_url" gorm:"column:award_url"`
	Award           string     `json:"award" form:"award" gorm:"column:award"`
	ScreenSet       int        `json:"screen_set" form:"screen_set" gorm:"column:screen_set"`
	InstepID        int        `json:"instep_id" form:"instep_id" gorm:"column:instep_id"`
	PriorityRegion  string     `json:"priority_region" form:"priority_region"`
	RegionWeight    int        `json:"region_weight" form:"region_weight"`
	GlobalWeight    int        `json:"global_weight" form:"global_weight"`
	TagShowPlatform int        `json:"tag_show_platform" form:"tag_show_platform"`
	WeightStime     stime.Time `json:"weight_stime" form:"weight_stime" time_format:"2006-01-02 15:04:05"`
	WeightEtime     stime.Time `json:"weight_etime" form:"weight_etime" time_format:"2006-01-02 15:04:05"`
}

// ActTimeConfig def
type ActTimeConfig struct {
	ID       int64      `json:"id" form:"id" gorm:"column:id"`
	Sid      int64      `json:"sid" form:"sid"`
	Interval int        `json:"interval" form:"interval"`
	Ctime    xtime.Time `json:"ctime" form:"ctime" time_format:"2006-01-02 15:04:05"`
	Mtime    xtime.Time `json:"mtime" form:"mtime" time_format:"2006-01-02 15:04:05"`
	Tlimit   int        `json:"tlimit" form:"tlimit"`
	Ltime    int        `json:"ltime" form:"ltime"`
}

// ActSubject def.
type ActSubject struct {
	ID                   int64      `json:"id,omitempty" form:"id" gorm:"column:id"`
	Oid                  int64      `json:"oid,omitempty" form:"oid"`
	Type                 int        `json:"type,omitempty" form:"type" validate:"min=1"`
	State                int        `json:"state,omitempty" form:"state"`
	Level                int        `json:"level,omitempty" form:"level"`
	Flag                 int64      `json:"flag,omitempty" form:"flag"`
	Rank                 int64      `json:"rank,omitempty" form:"rank"`
	Stime                stime.Time `json:"stime,omitempty" form:"stime" time_format:"2006-01-02 15:04:05"`
	Etime                stime.Time `json:"etime,omitempty" form:"etime" time_format:"2006-01-02 15:04:05"`
	Ctime                stime.Time `json:"ctime,omitempty" form:"ctime" time_format:"2006-01-02 15:04:05"`
	Mtime                stime.Time `json:"mtime,omitempty" form:"mtime" time_format:"2006-01-02 15:04:05"`
	Lstime               stime.Time `json:"lstime,omitempty" form:"lstime" time_format:"2006-01-02 15:04:05"`
	Letime               stime.Time `json:"letime,omitempty" form:"letime" time_format:"2006-01-02 15:04:05"`
	Uetime               stime.Time `json:"uetime,omitempty" form:"uetime" time_format:"2006-01-02 15:04:05"`
	Ustime               stime.Time `json:"ustime,omitempty" form:"ustime" time_format:"2006-01-02 15:04:05"`
	Name                 string     `json:"name,omitempty" form:"name"`
	Author               string     `json:"author,omitempty" form:"author"`
	ActURL               string     `json:"act_url,omitempty" form:"act_url"`
	Cover                string     `json:"cover,omitempty" form:"cover"`
	Dic                  string     `json:"dic,omitempty" form:"dic"`
	H5Cover              string     `json:"h5_cover,omitempty" form:"h5_cover"`
	LikeLimit            int        `json:"like_limit" form:"like_limit"`
	AndroidURL           string     `json:"android_url" form:"android_url"`
	IosURL               string     `json:"ios_url" form:"ios_url"`
	DailyLikeLimit       int64      `json:"daily_like_limit" form:"daily_like_limit"`
	DailySingleLikeLimit int64      `json:"daily_single_like_limit" form:"daily_single_like_limit"`
	UpLevel              int64      `json:"up_level" form:"up_level"`
	UpScore              int64      `json:"up_score" form:"up_score"`
	UpUetime             stime.Time `json:"up_uetime" form:"up_uetime"`
	UpUstime             stime.Time `json:"up_ustime" form:"up_ustime"`
	FanLimitMax          int64      `json:"fan_limit_max" form:"fan_limit_max"`
	FanLimitMin          int64      `json:"fan_limit_min" form:"fan_limit_min"`
	ChildSids            string     `json:"child_sids" form:"child_sids"`
	MonthScore           int        `json:"month_score" form:"month_score"`
	YearScore            int        `json:"year_score" form:"year_score"`
	UpFigureScore        int        `json:"up_figure_score" form:"up_figure_score"`
	ForbidenTime         stime.Time `json:"forbiden_time" form:"forbiden_time"`
	Contacts             string     `json:"contacts" form:"contacts"`
	PushStart            stime.Time `json:"push_start,omitempty" form:"-" time_format:"2006-01-02 15:04:05"`
	PushEnd              stime.Time `json:"push_end,omitempty" form:"-" time_format:"2006-01-02 15:04:05"`
	IsPush               int64      `json:"is_push" form:"-"`
	ShieldFlag           int64      `json:"shield_flag" form:"flag"`
	RelationID           int64      `json:"relation_id" form:"relation_id"`
	Calendar             string     `json:"calendar" form:"calendar"`
	AuditPlatform        string     `json:"audit_platform" form:"audit_platform"`
}

func (s *ActSubject) IsVideoSource() bool {
	return s.Type == VIDEO || s.Type == VIDEOLIKE || s.Type == VIDEO2 || s.Type == PHONEVIDEO || s.Type == SMALLVIDEO
}

// IsQuestionnaire ..
func (sub *ActSubject) IsQuestionnaire() bool {
	return ((sub.Flag >> FLAGQUESTIONNAIRE) & int64(1)) == 1
}

// SubProtocol .
type SubProtocol struct {
	*ActSubject
	Protocol *ActSubjectProtocol `json:"protocol"`
}

// ActSubjectResult .
type ActSubjectResult struct {
	*ActSubject
	Aids []int64 `json:"aids,omitempty"`
}

// Like def.
type Like struct {
	ID       int64       `json:"id" form:"id" gorm:"column:id"`
	Sid      int64       `json:"sid" form:"sid"`
	Type     int         `json:"type" form:"type"`
	Mid      int64       `json:"mid" form:"mid"`
	Wid      int64       `json:"wid" form:"wid"`
	State    int         `json:"state" form:"state"`
	StickTop int         `json:"stick_top" form:"stick_top"`
	Ctime    xtime.Time  `json:"ctime" form:"ctime" time_format:"2006-01-02 15:04:05"`
	Mtime    xtime.Time  `json:"mtime" form:"mtime" time_format:"2006-01-02 15:04:05"`
	Object   interface{} `json:"object,omitempty" gorm:"-"`
	Like     int64       `json:"like,omitempty" gorm:"-"`
}

// LikeAction def
type LikeAction struct {
	ID     int64      `form:"id" gorm:"column:id"`
	Lid    int64      `form:"lid"`
	Mid    int64      `form:"mid"`
	Action int64      `form:"action"`
	Ctime  xtime.Time `form:"ctime" time_format:"2006-01-02 15:04:05"`
	Mtime  xtime.Time `form:"mtime" time_format:"2006-01-02 15:04:05"`
	Sid    int64      `form:"sid"`
	IP     int64      `form:"ip" gorm:"column:ip"`
}

type LikeExport struct {
	*Like
	*LikeContent
}

// OptVideoListSub def
type OptVideoListSub struct {
	Page     int    `form:"page" default:"1" validate:"min=1"`
	PageSize int    `form:"pagesize" default:"15" validate:"min=1"`
	Keyword  string `form:"keyword"`
	Types    string `json:"types" form:"types"`
}

// TableName LikeAction def
func (LikeAction) TableName() string {
	return "like_action"
}

// TableName ActMatchs def.
func (ActSubject) TableName() string {
	return "act_subject"
}

// TableName Likes def
func (Like) TableName() string {
	return "likes"
}

// TableName ActSubjectProtocol def
func (ActSubjectProtocol) TableName() string {
	return "act_subject_protocol"
}

// TableName ActTimeConfig def
func (ActTimeConfig) TableName() string {
	return "act_time_config"
}

// SubjectStat.
type SubjectStat struct {
	ID    int64      `json:"id" gorm:"column:id"`
	Sid   int64      `json:"sid" gorm:"column:sid"`
	Num   int64      `json:"num" gorm:"column:num"`
	Ctime xtime.Time `json:"ctime" time_format:"2006-01-02 15:04:05" gorm:"column:ctime"`
	Mtime xtime.Time `json:"mtime" time_format:"2006-01-02 15:04:05" gorm:"column:mtime"`
}

// TableName SubjectStat def
func (SubjectStat) TableName() string {
	return "subject_stat"
}

type Award struct {
	ID           int64      `json:"id" gorm:"column:id"`
	Name         string     `json:"name" gorm:"column:name"`
	Etime        xtime.Time `json:"etime" time_format:"2006-01-02 15:04:05" gorm:"column:etime"`
	Sid          int64      `json:"sid" gorm:"column:sid"`
	Type         int        `json:"type" gorm:"column:type"`
	SourceID     string     `json:"source_id" gorm:"column:source_id"`
	SourceExpire int64      `json:"source_expire" gorm:"column:source_expire"`
	State        int64      `json:"state" gorm:"column:state"`
	Author       string     `json:"author" gorm:"column:author"`
	Ctime        xtime.Time `json:"ctime" time_format:"2006-01-02 15:04:05" gorm:"column:ctime"`
	Mtime        xtime.Time `json:"mtime" time_format:"2006-01-02 15:04:05" gorm:"column:mtime"`
	SidType      int        `json:"sid_type" gorm:"column:sid_type"`
	OtherSids    string     `json:"other_sids" gorm:"column:other_sids"`
	TaskID       int64      `json:"task_id" gorm:"column:task_id"`
}

// TableName SubjectStat def
func (Award) TableName() string {
	return "act_award_subject"
}

// AddAwardArg award add arg.
type AddAwardArg struct {
	Name         string     `form:"name" validate:"min=1"`
	Etime        xtime.Time `form:"etime" validate:"min=1"`
	Sid          int64      `form:"sid" validate:"min=1"`
	Type         int        `form:"type" validate:"min=1"`
	SourceID     []int64    `form:"source_id,split" validate:"max=3,min=1,dive,min=1"`
	SourceExpire int64      `form:"source_expire" validate:"min=1"`
	State        int64      `form:"state"`
	Author       string     `form:"-"`
	SidType      int        `form:"sid_type" validate:"min=0,max=1"`
	OtherSids    string     `form:"other_sids"`
	TaskID       int64      `form:"-"`
}

// SaveAwardArg award save arg.
type SaveAwardArg struct {
	ID int64 `form:"id" validate:"min=1"`
	AddAwardArg
}

type SubjectRule struct {
	ID        int64      `gorm:"column:id" json:"id"`
	Sid       int64      `gorm:"column:sid" json:"sid"`
	Category  int64      `gorm:"column:category" json:"category"`
	TypeIDs   string     `gorm:"column:type_ids" json:"type_ids"`
	Tags      string     `gorm:"column:tags" json:"tags"`
	TaskID    int64      `gorm:"column:task_id" json:"-"`
	State     int64      `gorm:"column:state" json:"state"`
	Attribute int64      `gorm:"column:attribute" json:"attribute"`
	RuleName  string     `gorm:"column:rule_name" json:"rule_name"`
	Ctime     xtime.Time `json:"ctime" time_format:"2006-01-02 15:04:05" gorm:"column:ctime"`
	Mtime     xtime.Time `json:"mtime" time_format:"2006-01-02 15:04:05" gorm:"column:mtime"`
	Stime     xtime.Time `json:"stime" time_format:"2006-01-02 15:04:05" gorm:"column:stime"`
	Etime     xtime.Time `json:"etime" time_format:"2006-01-02 15:04:05" gorm:"column:etime"`
}

type AddSubjectRuleArg struct {
	Sid       int64      `form:"sid" validate:"min=1"`
	Category  int64      `form:"category" validate:"min=1"`
	TypeIDs   string     `form:"type_ids" validate:"lt=400"`
	Tags      string     `form:"tags"`
	State     int64      `form:"state" validate:"min=0"`
	Attribute int64      `form:"attribute"`
	Stime     xtime.Time `form:"stime" time_format:"2006-01-02 15:04:05" gorm:"column:stime" validate:"required"`
	Etime     xtime.Time `form:"etime" time_format:"2006-01-02 15:04:05" gorm:"column:etime" validate:"required"`
}

// SaveAwardArg award save arg.
type SaveSubjectRuleArg struct {
	ID int64 `form:"id" validate:"min=1"`
	AddSubjectRuleArg
}

func (SubjectRule) TableName() string {
	return "act_subject_rule"
}

type SubRuleUserState struct {
	ID     int64 `json:"id"`
	Sid    int64 `json:"sid"`
	TaskID int64 `json:"task_id"`
	Total  int64 `json:"total"`
}

type SubRuleUserStateRes struct {
	List  []*SubRuleUserState `json:"list"`
	Total int64               `json:"total"`
}

func (s *SubjectRule) IsDayCount() bool {
	return s.attrVal(RuleAttrBitCountType) == AttrYes
}

func (s *SubjectRule) attrVal(bit uint) int64 {
	return (s.Attribute >> bit) & 1
}
