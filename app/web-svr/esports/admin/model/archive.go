package model

import (
	"fmt"
	"strings"
)

const (
	_arcBatchAddSQL        = "INSERT INTO es_archives(`aid`) VALUES %s"
	_whitesBatchInsertSQL  = "REPLACE INTO es_archive_whites(`mid`,`game_ids`,`match_ids`) VALUES %s"
	_tagBatchInsertSQL     = "REPLACE INTO es_archive_tags(`tag`,`game_ids`,`match_ids`) VALUES %s"
	_keywordBatchInsertSQL = "REPLACE INTO es_archive_keywords(`keyword`,`game_ids`,`match_ids`) VALUES %s"
)

// Arc .
type Arc struct {
	ID        int64 `json:"id"`
	Aid       int64 `json:"aid"`
	IsDeleted int   `json:"is_deleted"`
	Source    int   `json:"source" gorm:"-"`
}

// ArcAddParam .
type ArcAddParam struct {
	Aids     []int64  `form:"aids,split" validate:"dive,gt=1"`
	AidsStr  []string `form:"aids_str,split"`
	Gids     []int64  `form:"gids,split"`
	MatchIDs []int64  `form:"match_ids,split"`
	TeamIDs  []int64  `form:"team_ids,split"`
	TagIDs   []int64  `form:"tag_ids,split"`
	Years    []int64  `form:"years,split"`
}

// ArcImportParam .
type ArcImportParam struct {
	Aid      int64   `form:"aid" validate:"min=1"`
	Gids     []int64 `form:"gids,split"`
	MatchIDs []int64 `form:"match_ids,split"`
	TeamIDs  []int64 `form:"team_ids,split"`
	TagIDs   []int64 `form:"tag_ids,split"`
	Years    []int64 `form:"years,split"`
}

// ArcListParam .
type ArcListParam struct {
	Title      string  `form:"title"`
	Aid        string  `form:"aid"`
	TypeID     int64   `form:"type_id"`
	Copyright  int     `form:"copyright"`
	State      string  `form:"state"`
	CheckState []int64 `form:"check_state,split" validate:"dive,gte=0"`
	Source     int64   `form:"source"`
	Rules      []int64 `form:"rules,split"`
	Pn         int     `form:"pn" validate:"min=0" default:"1"`
	Ps         int     `form:"ps" validate:"min=0,max=30" default:"20"`
}

// SearchArc .
type SearchArc struct {
	ID     int64   `json:"id"`
	Aid    int64   `json:"aid"`
	TypeID int64   `json:"typeid"`
	Title  string  `json:"title"`
	State  int64   `json:"state"`
	Mid    int64   `json:"mid"`
	Gid    []int64 `json:"gid"`
	Tags   []int64 `json:"tags"`
	Matchs []int64 `json:"matchs"`
	Teams  []int64 `json:"teams"`
	Year   []int64 `json:"year"`
	Source int64   `json:"source"`
	Ctime  string  `json:"ctime"`
}

// ArcResult .
type ArcResult struct {
	ID           int64    `json:"id"`
	Aid          int64    `json:"aid"`
	BvID         string   `json:"bvid"`
	TypeID       int64    `json:"type_id"`
	Title        string   `json:"title"`
	State        int64    `json:"state"`
	Mid          int64    `json:"mid"`
	Uname        string   `json:"uname"`
	Games        []*Game  `json:"games"`
	Tags         []*Tag   `json:"tags"`
	Matchs       []*Match `json:"matchs"`
	Teams        []*Team  `json:"teams"`
	Years        []int64  `json:"years"`
	Source       int64    `json:"source"`
	RuleMid      int64    `json:"rule_mid"`
	RuleTags     []string `json:"rule_tags"`
	RuleKeywords []string `json:"rule_keywords"`
	Ctime        string   `json:"ctime"`
}

// ArcRelation .
type ArcRelation struct {
	AddGids     []*GIDMap
	UpAddGids   []int64
	UpDelGids   []int64
	AddMatchs   []*MatchMap
	UpAddMatchs []int64
	UpDelMatchs []int64
	AddTags     []*TagMap
	UpAddTags   []int64
	UpDelTags   []int64
	AddTeams    []*TeamMap
	UpAddTeams  []int64
	UpDelTeams  []int64
	AddYears    []*YearMap
	UpAddYears  []int64
	UpDelYears  []int64
}

// EsArchiveHit archive hit rule.
type EsArchiveHit struct {
	ID         int64  `json:"id"`
	ArcsID     int64  `json:"arcs_id"`
	WhiteMid   int64  `json:"white_mid"`
	TagIDs     string `json:"tag_ids"`
	KeywordIDs string `json:"keyword_ids"`
}

// ArchiveRule  archive hit tags,keywords.
type ArchiveRule struct {
	HitTags []int64
	HitKeys []int64
}

// EsArchiveWhite .
type EsArchiveWhite struct {
	ID        int64       `json:"id" form:"id"`
	Mid       int64       `json:"mid" form:"mid" validate:"required"`
	IsDeleted int         `json:"is_deleted" form:"is_deleted"`
	Uname     string      `json:"uname" gorm:"-"`
	Games     []*BaseInfo `json:"games" gorm:"-"`
	Matchs    []*BaseInfo `json:"matchs" gorm:"-"`
	GameIDs   string      `json:"-" form:"game_ids"`
	MatchIDs  string      `json:"-" form:"match_ids"`
	UserInfo  *BaseInfo   `json:"-" gorm:"-"`
}

// WhiteImportParam.
type WhiteImportParam struct {
	Mid      int64  `form:"mid" validate:"min=1"`
	Gids     string `form:"gids"`
	MatchIDs string `form:"match_ids"`
}

// EsArchiveTag.
type EsArchiveTag struct {
	ID        int64       `json:"id" form:"id"`
	Tag       string      `json:"tag" form:"tag" validate:"required"`
	IsDeleted int         `json:"is_deleted" form:"is_deleted"`
	Games     []*BaseInfo `json:"games" gorm:"-"`
	Matchs    []*BaseInfo `json:"matchs" gorm:"-"`
	GameIDs   string      `json:"-" form:"game_ids"`
	MatchIDs  string      `json:"-" form:"match_ids"`
	UserInfo  *BaseInfo   `json:"-" gorm:"-"`
}

// TagImportParam.
type TagImportParam struct {
	Tag      string `form:"tag" validate:"min=1"`
	Gids     string `form:"gids"`
	MatchIDs string `form:"match_ids"`
}

// EsArchiveKeyWord.
type EsArchiveKeyword struct {
	ID        int64       `json:"id" form:"id"`
	Keyword   string      `json:"keyword" form:"keyword" validate:"required"`
	IsDeleted int         `json:"is_deleted" form:"is_deleted"`
	Games     []*BaseInfo `json:"games" gorm:"-"`
	Matchs    []*BaseInfo `json:"matchs" gorm:"-"`
	GameIDs   string      `json:"-" form:"game_ids"`
	MatchIDs  string      `json:"-" form:"match_ids"`
	UserInfo  *BaseInfo   `json:"-" gorm:"-"`
}

// KeywordImportParam.
type KeywordImportParam struct {
	Keyword  string `form:"keyword" validate:"min=1"`
	Gids     string `form:"gids"`
	MatchIDs string `form:"match_ids"`
}

// BaseInfo.
type BaseInfo struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

// TableName .
func (a Arc) TableName() string {
	return "es_archives"
}

// ArcBatchAddSQL .
func ArcBatchAddSQL(aids []int64) (sql string, param []interface{}) {
	if len(aids) == 0 {
		return "", []interface{}{}
	}
	var rowStrings []string
	for _, aid := range aids {
		rowStrings = append(rowStrings, "(?)")
		param = append(param, aid)
	}
	return fmt.Sprintf(_arcBatchAddSQL, strings.Join(rowStrings, ",")), param
}

// WhiteBatchAddSQL .
func WhiteBatchAddSQL(list []*WhiteImportParam) (sql string, param []interface{}) {
	if len(list) == 0 {
		return "", []interface{}{}
	}
	var rowStrings []string
	for _, v := range list {
		rowStrings = append(rowStrings, "(?,?,?)")
		param = append(param, v.Mid, v.Gids, v.MatchIDs)
	}
	return fmt.Sprintf(_whitesBatchInsertSQL, strings.Join(rowStrings, ",")), param
}

func TagBatchAddSQL(list []*TagImportParam) (sql string, param []interface{}) {
	if len(list) == 0 {
		return "", []interface{}{}
	}
	var rowStrings []string
	for _, v := range list {
		rowStrings = append(rowStrings, "(?,?,?)")
		param = append(param, v.Tag, v.Gids, v.MatchIDs)
	}
	return fmt.Sprintf(_tagBatchInsertSQL, strings.Join(rowStrings, ",")), param
}

func KeywordBatchAddSQL(list []*KeywordImportParam) (sql string, param []interface{}) {
	if len(list) == 0 {
		return "", []interface{}{}
	}
	var rowStrings []string
	for _, v := range list {
		rowStrings = append(rowStrings, "(?,?,?)")
		param = append(param, v.Keyword, v.Gids, v.MatchIDs)
	}
	return fmt.Sprintf(_keywordBatchInsertSQL, strings.Join(rowStrings, ",")), param
}
