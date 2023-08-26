package rank

import (
	"time"

	xtime "go-common/library/time"

	"go-gateway/app/app-svr/app-feed/admin/model/common"
)

// 给网关的返回结果
type OpenListPager struct {
	List []*OpenListItem `json:"list"`
	Page common.Page     `json:"page"`
}

// 网关结果元素
type OpenListItem struct {
	ID         int               `json:"id"`
	RankConfig *RankConfigDetail `json:"rank_config"`
	RankVideos []int64           `json:"rank_videos"`
	RankState  int               `json:"rank_state"`
	FinalRank  []*FinalRankItem  `json:"final_rank"`
}

// 榜单配置详情
type RankConfigDetail struct {
	ID           int            `json:"id"`
	Title        string         `json:"title"`
	Cover        string         `json:"cover"`
	Description  []*Description `json:"description"`
	HelpTips     []string       `json:"help_tips"`
	STime        xtime.Time     `json:"stime"`
	ETime        xtime.Time     `json:"etime"`
	Cycle        int            `json:"cycle"`
	PerUpdate    int            `json:"per_update"`
	Tids         []int          `json:"tids"`
	ActIds       []int          `json:"act_ids"`
	ArchiveSTime xtime.Time     `json:"archive_stime"`
	ArchiveETime xtime.Time     `json:"archive_etime"`
}

// 结榜配置
type FinalRankItem struct {
	Position int     `json:"position"`
	Mode     int     `json:"mode"`
	Title    string  `json:"title"`
	List     []int64 `json:"list"`
}

// 运营后台榜单列表
type RankDetailPager struct {
	Title         string             `json:"title"`
	JobFinishTime int64              `json:"job_finish_time"`
	PublishState  int                `json:"publish_state"`
	List          []RankDetailAVItem `json:"list"`
	Page          common.Page        `json:"page"`
}

// 榜单视频详情
type RankDetailAVItem struct {
	User         *UserItem        `json:"user"`
	Avid         int64            `json:"avid"`
	Bvid         string           `json:"bvid"`
	ShowRank     int              `json:"show_rank"`
	IsHidden     int              `json:"is_hidden"`
	HiddenReason []string         `json:"hidden_reason"`
	ManualRank   int              `json:"manual_rank"`
	Title        string           `json:"title"`
	Score        *RankDetailScore `json:"score"`
}

// 榜单分数
type RankDetailScore struct {
	Rank       int         `json:"rank"`
	Total      int         `json:"total"`
	Play       int         `json:"play"`
	Like       int         `json:"like"`
	Coin       int         `json:"coin"`
	Share      int         `json:"share"`
	ExtraScore *ExtraScore `json:"extra"`
}

// 榜单配置列表
type RankListPager struct {
	Item []*RankListItem `json:"list"`
	Page *common.Page    `json:"page"`
}

// 榜单配置列表项
type RankListItem struct {
	Id     int          `json:"id"`
	Title  string       `json:"title"`
	Tids   []*IdAndName `json:"tids"`
	ActIds []*IdAndName `json:"act_ids"`
	Tags   []*IdAndName `json:"tags"`
	State  int          `json:"state"`
	STime  xtime.Time   `json:"stime"`
	ETime  xtime.Time   `json:"etime"`
	CUser  string       `json:"c_user"`
}

// id name 格式
type IdAndName struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

// 这个结构中包含了所有可能出现的请求字段,各函数按需使用
type RankCommonQuery struct {
	Id int `json:"id" form:"id"`

	State   int    `json:"state" form:"state"`
	Keyword string `json:"keyword" form:"keyword"`
	Time    int64  `json:"time" form:"time"`

	IsHidden int   `json:"is_hidden" form:"is_hidden"`
	Mid      int64 `json:"mid" form:"mid"`
	Avid     int64 `json:"avid" form:"avid"`

	Size int `json:"size" form:"size" default:"20"`
	Page int `json:"page" form:"page" default:"1"`
}

// 新建榜单配置的请求参数
type RankConfigReq struct {
	Title             string         `json:"title" validate:"required"`
	STime             xtime.Time     `json:"stime" validate:"required"`
	ETime             xtime.Time     `json:"etime" validate:"required"`
	Cycle             int            `json:"cycle" validate:"required"`
	PerUpdate         string         `json:"per_update" validate:"required"`
	Tids              []int          `json:"tids" validate:"required"`
	ActIds            []int          `json:"act_ids" validate:"required"`
	ArchiveStime      xtime.Time     `json:"archive_stime" validate:"required"`
	ArchiveEtime      xtime.Time     `json:"archive_etime" validate:"required"`
	ArchiveSelectMode int            `json:"archive_select_mode" validate:"required"`
	ScoreConfig       []*ScoreConfig `json:"score_config" validate:"required"`
	Blacklist         []int          `json:"blacklist"`
	Cover             string         `json:"cover" validate:"required"`
	Description       []*Description `json:"description" validate:"required"`
}

// 榜单配置编辑的请求参数
type EditRankConfigReq struct {
	ID                int            `json:"id" validate:"required"`
	Title             string         `json:"title" validate:"required"`
	STime             xtime.Time     `json:"stime" validate:"required"`
	ETime             xtime.Time     `json:"etime" validate:"required"`
	Cycle             int            `json:"cycle" validate:"required"`
	PerUpdate         string         `json:"per_update" validate:"required"`
	Tids              []int          `json:"tids" validate:"required"`
	ActIds            []int          `json:"act_ids" validate:"required"`
	ArchiveStime      xtime.Time     `json:"archive_stime" validate:"required"`
	ArchiveEtime      xtime.Time     `json:"archive_etime" validate:"required"`
	ArchiveSelectMode int            `json:"archive_select_mode" validate:"required"`
	Blacklist         []int          `json:"blacklist"`
	Cover             string         `json:"cover" validate:"required"`
	Description       []*Description `json:"description" validate:"required"`
}

// 榜单配置的返回结果
type RankConfigRes struct {
	ID                int            `json:"id"`
	Title             string         `json:"title"`
	STime             xtime.Time     `json:"stime"`
	ETime             xtime.Time     `json:"etime"`
	Cycle             int            `json:"cycle"`
	PerUpdate         int            `json:"per_update"`
	Tids              []*IdAndName   `json:"tids"`
	ActIds            []*IdAndName   `json:"act_ids"`
	ArchiveStime      xtime.Time     `json:"archive_stime"`
	ArchiveEtime      xtime.Time     `json:"archive_etime"`
	ArchiveSelectMode int            `json:"archive_select_mode"`
	ScoreConfig       []*ScoreConfig `json:"score_config"`
	Blacklist         []*UserItem    `json:"blacklist"`
	Cover             string         `json:"cover"`
	Description       []*Description `json:"description"`
	AvManuallyList    []int64        `json:"av_manually_added"`
	State             int            `json:"state"`
}

// 用户信息
type UserItem struct {
	Uid   int64  `json:"uid"`
	Uname string `json:"uname"`
}

// 榜单配置，DB
type RankConfig struct {
	ID                int        `json:"id" gorm:"column:id"`
	Title             string     `json:"title" gorm:"column:rank_name"`
	STime             xtime.Time `json:"stime" gorm:"column:stime"`
	ETime             xtime.Time `json:"etime" gorm:"column:etime"`
	Cycle             int        `json:"cycle" gorm:"column:cal_cycle"`
	PerUpdate         int        `json:"per_update" gorm:"column:update_freq"`
	Tids              string     `json:"tids" gorm:"column:tids"`
	ActIds            string     `json:"act_ids" gorm:"column:act_ids"`
	ArchiveStime      xtime.Time `json:"archive_stime" gorm:"column:av_stime"`
	ArchiveEtime      xtime.Time `json:"archive_etime" gorm:"column:av_etime"`
	ArchiveSelectMode int        `json:"archive_select_mode" gorm:"column:av_select_mode"`
	ScoreConfig       string     `json:"score_config" gorm:"column:score_rules"`
	Blacklist         string     `json:"blacklist" gorm:"column:blacklist"`
	Cover             string     `json:"cover" gorm:"column:rank_cover"`
	Description       string     `json:"description" gorm:"column:description"`
	State             int        `json:"state" gorm:"column:rank_status"`
	CUser             string     `json:"c_user" gorm:"column:c_user"`
	HistoryId         int        `json:"rank_history_id" gorm:"column:rank_history_id"`
	AvManuallyList    string     `json:"av_manually_added" gorm:"column:av_manually_added"`
}

func (*RankConfig) TableName() string {
	return "rank_config_list"
}

// 分数配置
type ScoreConfig struct {
	Action    string `json:"action" validate:"required"`
	Base      int    `json:"base" validate:"required"`
	CntPerDay int    `json:"cnt_per_day" validate:"required"`
}

// 榜单描述
type Description struct {
	Title       string `json:"title" validate:"required"`
	ContentType string `json:"content_type" validate:"required"`
	Content     string `json:"content" validate:"required"`
}

// 榜单分数
type RankScore struct {
	ID      int64      `json:"id" gorm:"column:id"`
	RankId  int        `json:"rank_id" gorm:"column:rank_id"`
	Avid    int64      `json:"avid" gorm:"column:avid"`
	Mid     int64      `json:"mid" gorm:"column:mid"`
	Score   int        `json:"score" gorm:"column:score"`
	Play    int        `json:"play" gorm:"column:play_score"`
	Like    int        `json:"like" gorm:"column:like_score"`
	Coin    int        `json:"coin" gorm:"column:coin_score"`
	Share   int        `json:"share" gorm:"column:share_score"`
	LogDate xtime.Time `json:"log_date" gorm:"column:log_date"`
	MTime   xtime.Time `json:"mtime" gorm:"column:mtime"`
}

func (*RankScore) TableName() string {
	return "rank_score"
}

// 榜单干预的请求参数
type RankArchiveIntervention struct {
	Avid     int64       `json:"avid" form:"avid" gorm:"avid"`
	RankId   int         `json:"rank_id" form:"rank_id" gorm:"rank_id"`
	Rank     int         `json:"rank" form:"rank" gorm:"rank"`
	IsHidden int         `json:"is_hidden" form:"is_hidden" gorm:"is_hidden"`
	Extra    *ExtraScore `json:"extra" form:"extra" gorm:"-"`
}

// 榜单干预，DB
type RankArchiveInterventionDB struct {
	RankId     int       `gorm:"column:rank_id"`
	Avid       int64     `gorm:"column:avid"`
	IsHidden   int       `gorm:"column:is_hidden"`
	RankPos    int       `gorm:"column:rank_pos"`
	ExtraScore string    `gorm:"column:extra_score"`
	CUser      string    `gorm:"column:c_user"`
	Ctime      time.Time `gorm:"-"`
	Mtime      time.Time `gorm:"column:mtime"`
}

func (*RankArchiveInterventionDB) TableName() string {
	return "rank_intervention"
}

// 评委分
type ExtraScore struct {
	Complete  int `json:"complete" form:"complete"`
	Reduction int `json:"reduction" form:"reduction"`
	Creative  int `json:"creative" form:"creative"`
}

// 榜单历史，DB
type RankHistoryDB struct {
	ID              int64      `gorm:"column:id"`
	RankId          int        `gorm:"column:rank_id"`
	LogData         xtime.Time `gorm:"column:log_date"`
	ScoreAvids      string     `gorm:"column:score_avids"`
	FinalRankConfig string     `gorm:"column:final_rank_config"`
	CUser           string     `gorm:"column:c_user"`
	Ctime           xtime.Time `gorm:"column:ctime"`
	Mtime           xtime.Time `gorm:"column:mtime"`
}

func (*RankHistoryDB) TableName() string {
	return "rank_history"
}

// 结榜参数
type TernimateContent struct {
	Id        int        `json:"id" form:"id"`
	TmContent *[]TMModel `json:"rank" form:"rank"`
}

// 结榜模块项
type TMModel struct {
	Position          int     `json:"position" form:"position"`
	Mode              int     `json:"mode" form:"mode"`
	ArchiveSelectMode int     `json:"archive_select_mode" form:"archive_select_mode"`
	Title             string  `json:"title" form:"title"`
	Count             int     `json:"count" form:"count"`
	UidList           []int64 `json:"uid_list" form:"uid_list"`
}

// hive任务状态，轮询结果
type ResponseHiveCheck struct {
	Code      int      `json:"code"`
	Msg       string   `json:"msg"`
	StatusId  string   `json:"statusId"`
	StatusMsg string   `json:"statusMsg"`
	HdfsPath  []string `json:"hdfsPath"`
}

// 数据平台返回最终符合活动要求的稿件信息项
type DataPlatAvInfo struct {
	Avid int64
	Mid  int64
}
