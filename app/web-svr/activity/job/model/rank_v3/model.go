package rank

import (
	go_common_library_time "go-common/library/time"
)

const (
	// FrequencyTypeDay 按日更新
	FrequencyTypeDay = 1
	// FrequencyTypeWeek 按周更新
	FrequencyTypeWeek = 2
	// FrequencyTypeMonth 按月更新
	FrequencyTypeMonth = 3
	// FrequencyTypeOnce 更新一次
	FrequencyTypeOnce = 4
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
)

// Base 排行榜基础信息
type Base struct {
	ID           int64                       `form:"id" json:"id"`
	Name         string                      `form:"name" json:"name"`
	RankType     int                         `form:"rank_type" json:"rank_type"`
	IsType       int                         `form:"is_type" json:"is_type"`
	Tids         string                      `form:"tids" json:"tids"`
	TidsStruct   []int64                     `form:"_" json:"_"`
	Cover        string                      `form:"cover" json:"cover"`
	Description  string                      `form:"description" json:"description"`
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

// Rule ...
type Rule struct {
	ID              int64                       `form:"id" json:"id"`
	BaseID          int64                       `form:"base_id" json:"base_id"`
	Name            string                      `form:"name" json:"name"`
	StatisticsType  int                         `form:"statistics_type" json:"statistics_type"`
	Nums            int                         `form:"nums" json:"nums"`
	UpdateFrequency int                         `form:"update_frequency" json:"update_frequency"`
	UpdateScope     int                         `form:"update_scope" json:"update_scope"`
	ScoreConfig     string                      `form:"score_config" json:"score_config"`
	LastBatch       int                         `form:"last_batch" json:"last_batch"`
	LastBatchTime   go_common_library_time.Time `form:"last_batch_time" json:"last_batch_time"`
	ShowBatch       int                         `form:"show_batch" json:"show_batch"`
	ShowBatchTime   go_common_library_time.Time `form:"show_batch_time" json:"show_batch_time"`
	State           int                         `form:"state" json:"state"`
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

// RuleBatchTime ...
type RuleBatchTime struct {
	RuleID        int64                       `form:"rule_id" json:"rule_id"`
	LastBatch     int                         `form:"last_batch" json:"last_batch"`
	LastBatchTime go_common_library_time.Time `form:"last_batch_time" json:"last_batch_time"`
	ShowBatch     int                         `form:"show_batch" json:"show_batch"`
	ShowBatchTime go_common_library_time.Time `form:"show_batch_time" json:"show_batch_time"`
}
