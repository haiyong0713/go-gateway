package rank

import (
	go_common_library_time "go-common/library/time"
)

const (
	// RankStateStart 进行中
	RankStateStart = 1
	// RankStateDelete 已删除
	RankStateDelete = 2
	// RankStateNotStart 未开始
	RankStateNotStart = 3
	// RankStateEnd 已结束
	RankStateEnd = 4
	// RankStateResult 已解绑
	RankStateResult = 5

	// FrequencyTypeDay 按日更新
	FrequencyTypeDay = 1
	// FrequencyTypeWeek 按周更新
	FrequencyTypeWeek = 2
	// FrequencyTypeMonth 按月更新
	FrequencyTypeMonth = 3
	// UpdateScopeIncrement 增量更新
	UpdateScopeIncrement = 1
	// UpdateScopeTotal 累计更新
	UpdateScopeTotal = 2
	// LogStateDelete 日志删除
	LogStateDelete = 0
	// LogStateInit 日志初始化
	LogStateInit = 1
	// LogStateFinish 日志完成
	LogStateFinish = 2
	// LogStatePublish 日志发布
	LogStatePublish = 3
	// SourceTypeActivity 数据来源活动id
	SourceTypeActivity = 1
	// SourceTypeTag 数据来源tagid
	SourceTypeTag = 2
	// SourceTypeMid 数据来源mid
	SourceTypeMid = 3
	// SourceStateOnline 有效
	SourceStateOnline = 1
	// SourceStateOffline 无效
	SourceStateOffline = 2
	// BlackWhiteInterventionTypeWhite 白名单
	BlackWhiteInterventionTypeWhite = 1
	// BlackWhiteInterventionTypeBlack 黑名单
	BlackWhiteInterventionTypeBlack = 2
	// BlackWhiteObjectTypeUp 用户
	BlackWhiteObjectTypeUp = 1
	// BlackWhiteObjectTypeArchive 稿件
	BlackWhiteObjectTypeArchive = 2
	// IsNotType 不分区
	IsNotType = 1
	// IsType 分区
	IsType = 2
	// StatisticsTypeArchive 作品榜
	StatisticsTypeArchive = 1
	// StatisticsTypeDistinctArchive 作品榜去重
	StatisticsTypeDistinctArchive = 2
	// StatisticsTypeUp up榜
	StatisticsTypeUp = 3
	// StatisticsTypeTag tag榜
	StatisticsTypeTag = 4
	// RankTypeAct 活动榜
	RankTypeAct = 1
	// RankTypeTag 品类榜
	RankTypeTag = 2
	// ObjectTypeUp 用户
	ObjectTypeUp = 1
	// ObjectTypeArchive 稿件
	ObjectTypeArchive = 2
	// ObjectTypeTag tag
	ObjectTypeTag = 3
	// IsShow 展示
	IsShow = 1
	// IsNotShow 不展示
	IsNotShow = 2
	// ArchiveNums 保留稿件数量
	ArchiveNums = 3
	// PrecisionZero 保留0位小数
	PrecisionZero = 0
	// PrecisionOne 保留一位小数
	PrecisionOne = 1
	// PrecisionTwo 保留两位小数
	PrecisionTwo = 2
	// PrecisionThree 保留三位小数
	PrecisionThree = 3
	// UnitOne 单位个
	UnitOne = 0
	// UnitTenThousands 单位万
	UnitTenThousands = 1
)

// ResultOid ...
type ResultOid struct {
	ID        int64  `json:"id"`
	BaseID    int64  `json:"base_id"`
	RankID    int64  `json:"rank_id"`
	OID       int64  `json:"oid"`
	Rank      int64  `json:"rank"`
	Score     int64  `json:"score"`
	ShowScore string `json:"show_score"`
	Batch     int    `json:"batch"`
	State     int    `json:"state"`
}

// ResultOidArchive ...
type ResultOidArchive struct {
	ID        int64  `json:"id"`
	BaseID    int64  `json:"base_id"`
	RankID    int64  `json:"rank_id"`
	AID       int64  `json:"aid"`
	OID       int64  `json:"oid"`
	Rank      int64  `json:"rank"`
	Score     int64  `json:"score"`
	ShowScore string `json:"show_score"`
	Batch     int    `json:"batch"`
	State     int    `json:"state"`
}

// IdAndName id name 格式
type IdAndName struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

// Res ...
type Res struct {
	*BaseRes `json:"base"`
	Source   []*IdAndName     `json:"source"`
	Rule     []*Rule          `json:"rule"`
	Black    []*BlackWhiteRes `json:"black"`
}

// BaseRes ...
type BaseRes struct {
	ID           int64                       `form:"id" json:"id"`
	Name         string                      `form:"name" json:"name"`
	RankType     int                         `form:"rank_type" json:"rank_type"`
	IsType       int                         `form:"is_type" json:"is_type"`
	Tids         []*IdAndName                `form:"tids" json:"tids"`
	ArchiveStime int64                       `form:"archive_stime" json:"archive_stime"`
	ArchiveEtime int64                       `form:"archive_etime" json:"archive_etime"`
	State        int                         `form:"state" json:"state"`
	Author       string                      `form:"author" json:"author"`
	Authority    string                      `form:"authority" json:"authority"`
	Stime        go_common_library_time.Time `form:"stime" json:"stime"`
	Etime        go_common_library_time.Time `form:"etime" json:"etime"`
	Ctime        go_common_library_time.Time `json:"ctime"`
	Mtime        go_common_library_time.Time `json:"mtime"`
}

// Base 排行榜基础信息
type Base struct {
	ID           int64                       `form:"id" json:"id"`
	Name         string                      `form:"name" json:"name"`
	IsShowScore  int                         `form:"is_show_score" json:"is_show_score"`
	RankType     int                         `form:"rank_type" json:"rank_type"`
	IsType       int                         `form:"is_type" json:"is_type"`
	Tids         string                      `form:"tids" json:"tids"`
	TidsStruct   []int64                     `form:"_" json:"_"`
	ArchiveStime int64                       `form:"archive_stime" json:"archive_stime"`
	ArchiveEtime int64                       `form:"archive_etime" json:"archive_etime"`
	State        int                         `form:"state" json:"state"`
	Author       string                      `form:"author" json:"author"`
	Authority    string                      `form:"authority" json:"authority"`
	Ctime        go_common_library_time.Time `json:"ctime"`
	Mtime        go_common_library_time.Time `json:"mtime"`
}

// RuleBatchTime ...
type RuleBatchTime struct {
	RuleID        int64 `form:"rule_id" json:"rule_id"`
	LastBatch     int   `form:"last_batch" json:"last_batch"`
	LastBatchTime int64 `form:"last_batch_time" json:"last_batch_time"`
}

// Rule ...
type Rule struct {
	ID              int64                       `form:"id" json:"id"`
	BaseID          int64                       `form:"base_id" json:"base_id"`
	Name            string                      `form:"name" json:"name"`
	StatisticsType  int                         `form:"statistics_type" json:"statistics_type"`
	Nums            int                         `form:"nums" json:"nums"`
	LastBatch       int                         `form:"last_batch" json:"last_batch"`
	LastBatchTime   int64                       `form:"last_batch_time" json:"last_batch_time"`
	ShowBatch       int                         `form:"show_batch" json:"show_batch"`
	ShowBatchTime   int64                       `form:"show_batch_time" json:"show_batch_time"`
	UpdateFrequency int                         `form:"update_frequency" json:"update_frequency"`
	UpdateScope     int                         `form:"update_scope" json:"update_scope"`
	Score           []*ScoreConfig              `form:"update_scope" json:"score"`
	State           int                         `form:"state" json:"state"`
	Precision       int                         `form:"precision" json:"precision"`
	Unit            int                         `form:"unit" json:"unit"`
	Description     string                      `form:"description" json:"description"`
	Stime           go_common_library_time.Time `form:"stime" json:"stime"`
	Etime           go_common_library_time.Time `form:"etime" json:"etime"`
	Ctime           go_common_library_time.Time `json:"ctime"`
	Mtime           go_common_library_time.Time `json:"mtime"`
}

// Log ,,,
type Log struct {
	ID          int64                       `form:"id" json:"id"`
	BaseID      int64                       `form:"base_id" json:"base_id"`
	RankID      int64                       `form:"rank_id" json:"rank_id"`
	Batch       int                         `form:"batch" json:"batch"`
	State       int                         `form:"state" json:"state"`
	ThisDate    string                      `form:"this_date" json:"this_date"`
	LastDate    string                      `form:"last_date" json:"last_date"`
	FinishTime  int64                       `form:"finish_time" json:"finish_time"`
	PublishTime int64                       `form:"publish_time" json:"publish_time"`
	Ctime       go_common_library_time.Time `json:"ctime"`
	Mtime       go_common_library_time.Time `json:"mtime"`
}

// SourceRes ...
type SourceRes struct {
	ID         int64                       `form:"id" json:"id"`
	BaseID     int64                       `form:"base_id" json:"base_id"`
	SourceID   int64                       `form:"source_id" json:"source_id"`
	SourceType int                         `form:"source_type" json:"source_type"`
	Name       string                      `form:"name" json:"name"`
	State      int                         `form:"state" json:"state"`
	Ctime      go_common_library_time.Time `json:"ctime"`
	Mtime      go_common_library_time.Time `json:"mtime"`
}

// Source ...
type Source struct {
	ID         int64                       `form:"id" json:"id"`
	BaseID     int64                       `form:"base_id" json:"base_id"`
	SourceID   int64                       `form:"source_id" json:"source_id"`
	SourceType int                         `form:"source_type" json:"source_type"`
	State      int                         `form:"state" json:"state"`
	Name       string                      `form:"name" json:"name"`
	Ctime      go_common_library_time.Time `json:"ctime"`
	Mtime      go_common_library_time.Time `json:"mtime"`
}

// BlackWhiteRes ...
type BlackWhiteRes struct {
	ID               int64                       `form:"id" json:"id"`
	Name             string                      `form:"name" json:"name"`
	BaseID           int64                       `form:"base_id" json:"base_id"`
	Oid              int64                       `form:"oid" json:"oid"`
	Score            int64                       `form:"score" json:"score"`
	State            int                         `form:"state" json:"state"`
	InterventionType int                         `form:"intervention_type" json:"intervention_type"`
	ObjectType       int                         `form:"object_type" json:"object_type"`
	Ctime            go_common_library_time.Time `json:"ctime"`
	Mtime            go_common_library_time.Time `json:"mtime"`
}

// BlackWhite ...
type BlackWhite struct {
	ID               int64                       `form:"id" json:"id"`
	BaseID           int64                       `form:"base_id" json:"base_id"`
	Oid              int64                       `form:"oid" json:"oid" validate:"required"`
	Score            int64                       `form:"score" json:"score"`
	State            int                         `form:"state" json:"state"`
	InterventionType int                         `form:"intervention_type" json:"intervention_type" validate:"required"`
	ObjectType       int                         `form:"object_type" json:"object_type" validate:"required"`
	Ctime            go_common_library_time.Time `json:"ctime"`
	Mtime            go_common_library_time.Time `json:"mtime"`
}

// Result ...
type Result struct {
	BaseID     int64                       `json:"base_id"`
	ParentID   int64                       `json:"parent_id"`
	AID        int64                       `json:"aid"`
	OID        int64                       `json:"oid"`
	RankID     int64                       `json:"rank_id"`
	MID        int64                       `json:"mid"`
	TagID      int64                       `json:"tag_id"`
	LogID      int64                       `json:"log_id"`
	LogDate    int64                       `json:"log_date"`
	Batch      int                         `json:"batch"`
	WhiteScore int64                       `json:"white_score"`
	FansScore  int64                       `json:"fans_score"`
	CountScore int64                       `json:"count_score"`
	Score      int64                       `json:"score"`
	LastScore  int64                       `json:"last_score"`
	TodayScore int64                       `json:"today_score"`
	RankType   int64                       `json:"rank_type"`
	ID         string                      `json:"id"`
	LikesScore int64                       `json:"likes_score"`
	PlayScore  int64                       `json:"play_score"`
	CoinScore  int64                       `json:"coin_score"`
	ShareScore int64                       `json:"share_score"`
	SourceID   int64                       `json:"source_id"`
	Rank       int64                       `json:"rank"`
	ObjectType int                         `json:"object_type"`
	Ctime      go_common_library_time.Time `json:"ctime"`
	Mtime      go_common_library_time.Time `json:"mtime"`
}

// Score ...
type Score struct {
	Rank      int64  `json:"rank"`
	Total     int64  `json:"total"`
	Play      int64  `json:"play"`
	Like      int64  `json:"like"`
	Coin      int64  `json:"coin"`
	Share     int64  `json:"share"`
	Fans      int64  `json:"fans"`
	Extra     int64  `json:"extra"`
	ShowScore string `json:"show_score"`
}

// ResultDetail ...
type ResultDetail struct {
	AID          int64    `json:"aid"`
	IsHidden     int      `json:"is_hidden"`
	HiddenReason []string `json:"hidden_reason"`
	ManualRank   int64    `json:"manual_rank"`
	Score        *Score   `json:"score"`
	ShowRank     int64    `json:"show_rank"`
	Mid          int64    `json:"mid"`
	TagID        int64    `json:"tag_id"`
	OID          int64    `json:"oid"`
	ObjectType   int      `json:"object_type"`
}

// ResultList ...
type ResultList struct {
	Rule *Rule        `json:"rank"`
	List []*ResultRes `json:"list"`
	Page *Page        `json:"page"`
}

// ResultRes ...
type ResultRes struct {
	AID          int64    `json:"aid"`
	TagID        int64    `json:"tag_id"`
	ObjectType   int      `json:"object_type"`
	IsHidden     int      `json:"is_hidden"`
	HiddenReason []string `json:"hidden_reason"`
	ManualRank   int64    `json:"manual_rank"`
	Score        *Score   `json:"score"`
	ShowRank     int64    `json:"show_rank"`
	Mid          int64    `json:"mid"`
	Account      *Account `json:"account"`
	Archive      *Archive `json:"archive"`
	Tag          *Tag     `json:"tag"`
}

// Tag ...
type Tag struct {
	TID  int64  `json:"tag_id"`
	Name string `json:"name"`
}

// Archive ...
type Archive struct {
	Aid   int64  `json:"aid"`
	Bvid  string `json:"bvid"`
	Title string `json:"title"`
}

// Account ...
type Account struct {
	Mid  int64  `json:"mid"`
	Name string `json:"uname"`
}

// ScoreConfig ...
type ScoreConfig struct {
	ID        int64                       `form:"id" json:"id"`
	RankID    int64                       `form:"rank_id" json:"rank_id"`
	Action    string                      `form:"action" json:"action"`
	Base      int                         `form:"base" json:"base"`
	CntPerDay int                         `form:"cnt_per_day" json:"cnt_per_day"`
	Ctime     go_common_library_time.Time `json:"ctime"`
	Mtime     go_common_library_time.Time `json:"mtime"`
}

// Adjust ...
type Adjust struct {
	ID         int64                       `form:"id" json:"id"`
	BaseID     int64                       `form:"base_id" json:"base_id"`
	ParentID   int64                       `form:"parent_id" json:"parent_id"`
	RankID     int64                       `form:"rank_id" json:"rank_id"`
	OID        int64                       `form:"oid" json:"oid"`
	ObjectType int                         `form:"object_type" json:"object_type"`
	Rank       int64                       `form:"rank" json:"rank"`
	IsShow     int                         `form:"is_show" json:"is_show"`
	State      int                         `form:"state" json:"state"`
	Ctime      go_common_library_time.Time `json:"ctime"`
	Mtime      go_common_library_time.Time `json:"mtime"`
}

// ListRsp ...
type ListRsp struct {
	List []*Res `json:"list"`
	Page *Page  `json:"page"`
}

// SourceListRsp ...
type SourceListRsp struct {
	List []*Source `json:"list"`
	Page *Page     `json:"page"`
}

// Page ...
type Page struct {
	Num   int `json:"num"`
	Size  int `json:"size"`
	Total int `json:"total"`
}
