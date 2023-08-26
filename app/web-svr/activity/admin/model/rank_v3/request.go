package rank

import go_common_library_time "go-common/library/time"

// SourceReq ...
type SourceReq struct {
	ID         int64 `form:"id" json:"id"`
	SourceID   int64 `form:"source_id" json:"source_id"`
	SourceType int   `form:"source_type" json:"source_type"`
}

// Description ...
type Description struct {
	Title       string `form:"title" json:"title"  validate:"required"`
	Content     string `form:"content" json:"content"  validate:"required"`
	ContentType string `form:"content_type" json:"content_type"  validate:"required"`
}

// CreateReq 排行榜基础信息
type CreateReq struct {
	Name         string                      `form:"name" json:"name"  validate:"required"`
	IsShowScore  int                         `form:"is_show_score" json:"is_show_score"`
	RankType     int                         `form:"rank_type" json:"rank_type"  validate:"required"`
	IsType       int                         `form:"is_type" json:"is_type"  validate:"required"`
	Tids         []int64                     `form:"tids" json:"tids"`
	ArchiveStime int64                       `form:"archive_stime" json:"archive_stime"  validate:"required"`
	ArchiveEtime int64                       `form:"archive_etime" json:"archive_etime"  validate:"required"`
	Author       string                      `form:"author" json:"author"`
	Authority    string                      `form:"authority" json:"authority"`
	Ctime        go_common_library_time.Time `json:"ctime"`
	Mtime        go_common_library_time.Time `json:"mtime"`
	Source       []*Source                   `json:"source"`
	BlackWhite   []*BlackWhite               `json:"black"`
}

// CreateRes ...
type CreateRes struct {
	ID int64 ` json:"id"`
}

// ListReq ...
type ListReq struct {
	State     int    `form:"state" json:"state"`
	Keyword   string `form:"keyword" json:"keyword"`
	RankType  int    `form:"rank_type" json:"rank_type"`
	ValidTime int64  `form:"valid_time" json:"valid_time"`
	Pn        int    `form:"pn" json:"pn" default:"1"`
	Ps        int    `form:"ps" json:"ps" default:"20"`
}

// SourceListReq ...
type SourceListReq struct {
	BaseID int64 `form:"base_id" json:"base_id"  validate:"required"`
	Pn     int   `form:"pn" json:"pn" default:"1"`
	Ps     int   `form:"ps" json:"ps" default:"20"`
}

// BaseReq ...
type BaseReq struct {
	ID int64 `form:"id" json:"id"  validate:"required"`
}

// UpdateReq 排行榜基础信息
type UpdateReq struct {
	ID           int64                       `form:"id" json:"id"  validate:"required"`
	Name         string                      `form:"name" json:"name"`
	IsShowScore  int                         `form:"is_show_score" json:"is_show_score"`
	RankType     int                         `form:"rank_type" json:"rank_type"`
	IsType       int                         `form:"is_type" json:"is_type"`
	Tids         []int64                     `form:"tids" json:"tids"`
	ArchiveStime int64                       `form:"archive_stime" json:"archive_stime"`
	ArchiveEtime int64                       `form:"archive_etime" json:"archive_etime"`
	Author       string                      `form:"author" json:"author"`
	Authority    string                      `form:"authority" json:"authority"`
	State        int                         `form:"state" json:"state"`
	Ctime        go_common_library_time.Time `json:"ctime"`
	Mtime        go_common_library_time.Time `json:"mtime"`
	Source       []*Source                   `json:"source"`
	BlackWhite   []*BlackWhite               `json:"black"`
}

// ExportReq ...
type ExportReq struct {
	RankID int64 `form:"rank_id" json:"rank_id" validate:"required"`
}

// PublishReq ...
type PublishReq struct {
	RankID int64 `form:"rank_id" json:"rank_id" validate:"required"`
	BaseID int64 `form:"base_id" json:"base_id" validate:"required"`
}

// ExportRankResultReq ...
type ExportRankResultReq struct {
	RankID     int64   `form:"rank_id" json:"rank_id" validate:"required"`
	BaseID     int64   `form:"base_id" json:"base_id" validate:"required"`
	ObjectType int     `form:"object_type" json:"object_type"`
	Aid        []int64 `form:"aid" json:"aid"`
	MID        []int64 `form:"mid" json:"mid"`
	TagID      []int64 `form:"tag_id" json:"tag_id"`
}

// ResultReq ...
type ResultReq struct {
	BaseID     int64   `form:"base_id" json:"base_id" validate:"required"`
	RankID     int64   `form:"rank_id" json:"rank_id" validate:"required"`
	IsShow     int64   `form:"is_show" json:"is_show"`
	Aid        []int64 `form:"aid" json:"aid"`
	MID        []int64 `form:"mid" json:"mid"`
	LastID     int64   `form:"last_id" json:"last_id"`
	Ps         int64   `form:"ps" json:"ps" default:"20"`
	Pn         int64   `form:"pn" json:"pn" default:"1"`
	ObjectType int     `form:"object_type" json:"object_type"`
	TagID      []int64 `form:"tag_id" json:"tag_id"`
}

// UpdateAdjustObject ...
type UpdateAdjustObject struct {
	RankID     int64 `form:"rank_id" json:"rank_id" validate:"required"`
	BaseID     int64 `form:"base_id" json:"base_id" validate:"required"`
	OID        int64 `form:"oid" json:"oid" validate:"required"`
	ObjectType int   `form:"object_type" json:"object_type" validate:"required"`
	Rank       int64 `form:"rank" json:"rank"`
	IsShow     int   `form:"is_show" json:"is_show"`
	State      int   `form:"state" json:"state"`
	ParentID   int64 `form:"parent_id" json:"parent_id"`
}

// BlackWhiteReq ...
type BlackWhiteReq struct {
	BaseID     int64         `form:"base_id" json:"base_id" validate:"required"`
	RankID     int64         `form:"rank_id" json:"rank_id"`
	BlackWhite []*BlackWhite `json:"black_white" validate:"required"`
}

// RuleUpdateReq ...
type RuleUpdateReq struct {
	ID              int64                       `form:"id" json:"id"`
	BaseID          int64                       `form:"base_id" json:"base_id"  validate:"required"`
	Name            string                      `form:"name" json:"name"  validate:"required"`
	StatisticsType  int                         `form:"statistics_type" json:"statistics_type"  validate:"required"`
	Nums            int                         `form:"nums" json:"nums"  validate:"required"`
	UpdateFrequency int                         `form:"update_frequency" json:"update_frequency"  validate:"required"`
	State           int                         `form:"state" json:"state"`
	UpdateScope     int                         `form:"update_scope" json:"update_scope"  validate:"required"`
	Stime           go_common_library_time.Time `form:"stime" json:"stime"  validate:"required"`
	Etime           go_common_library_time.Time `form:"etime" json:"etime"  validate:"required"`
	Precision       int                         `form:"precision" json:"precision"`
	Unit            int                         `form:"unit" json:"unit"`
	Description     string                      `form:"description" json:"description"`
	ScoreConfig     []*ScoreConfig              `form:"score_config" json:"score_config"  validate:"required"`
}

// RuleUpdateShowReq ...
type RuleUpdateShowReq struct {
	ID          int64  `form:"id" json:"id"  validate:"required"`
	BaseID      int64  `form:"base_id" json:"base_id"  validate:"required"`
	Precision   int    `form:"precision" json:"precision"`
	Unit        int    `form:"unit" json:"unit"`
	Description string `form:"description" json:"description"`
}

// RuleOfflineReq ...
type RuleOfflineReq struct {
	RankID int64 `form:"rank_id" json:"rank_id"`
}
