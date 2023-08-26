package model

import (
	"fmt"
	"strings"
)

const (
	_moduleInsertSQL  = "INSERT INTO es_matchs_module(ma_id,name,oids) VALUES %s"
	_moduleEditSQL    = "UPDATE es_matchs_module  SET name = CASE %s END,oids = CASE %s END WHERE id IN (%s)"
	_actLiveInsertSQL = "INSERT INTO es_active_live(ma_id,title,live_id) VALUES %s"
	_actLiveEditSQL   = "UPDATE es_active_live SET title = CASE %s END,live_id = CASE %s END WHERE id IN (%s)"
)

// ParamMA .
type ParamMA struct {
	MatchActive
	Modules    string `json:"-" form:"modules"`
	Adid       int64  `json:"-" form:"adid" validate:"required"`
	ActiveLive string `json:"-" form:"active_live" gorm:"-"`
}

// Activelive active live
type Activelive struct {
	ID        int64  `gorm:"column:id" json:"id"`
	MaId      int64  `gorm:"column:ma_id" json:"ma_id"`
	Title     string `gorm:"column:title" json:"title"`
	LiveId    int64  `gorm:"column:live_id" json:"live_id"`
	IsDeleted int64  `gorm:"column:is_deleted" json:"is_deleted"`
}

// MatchActive .
type MatchActive struct {
	ID           int64  `json:"id" form:"id"`
	Sid          int64  `json:"sid" form:"sid"`
	Sids         string `json:"sids" form:"sids"`
	Mid          int64  `json:"mid" form:"mid" validate:"required"`
	Background   string `json:"background" form:"background"`
	BackColor    string `json:"back_color" form:"back_color"`
	ColorStep    string `json:"color_step" form:"color_step"`
	LiveID       int64  `json:"live_id" form:"live_id"`
	Intr         string `json:"intr" form:"intr"`
	Focus        string `json:"focus" form:"focus"`
	URL          string `json:"url" form:"url"`
	Status       int    `json:"status" form:"status"`
	H5Background string `json:"h5_background" form:"h5_background"`
	H5BackColor  string `json:"h5_back_color" form:"h5_back_color"`
	H5Focus      string `json:"h5_focus" form:"h5_focus"`
	H5URL        string `json:"h5_url" form:"h5_url"`
	IntrLogo     string `json:"intr_logo" form:"intr_logo"`
	IntrTitle    string `json:"intr_title" form:"intr_title"`
	IntrText     string `json:"intr_text" form:"intr_text"`
	IsLive       int    `json:"is_live" form:"is_live"`
}

// Module .
type Module struct {
	ID     int64  `json:"id"`
	MaID   int64  `json:"ma_id"`
	Name   string `json:"name"`
	Oids   string `json:"oids"`
	Status int    `json:"-" form:"status"`
	Bvids  string `json:"bv_ids"`
}

// MatchModule .
type MatchModule struct {
	*MatchActive
	Modules        []*Module     `json:"modules"`
	MatchTitle     string        `json:"match_title"`
	MatchSubTitle  string        `json:"match_sub_title"`
	SeasonTitle    string        `json:"season_title"`
	SeasonSubTitle string        `json:"season_sub_title"`
	ActiveLive     []*Activelive `json:"active_live"`
}

// TableName Activelive.
func (t Activelive) TableName() string {
	return "es_active_live"
}

// TableName es_matchs_module.
func (t Module) TableName() string {
	return "es_matchs_module"
}

// TableName es_matchs_active.
func (t MatchActive) TableName() string {
	return "es_matchs_active"
}

// BatchAddModuleSQL .
func BatchAddModuleSQL(maID int64, data []*Module) (sql string, param []interface{}) {
	if len(data) == 0 {
		return "", []interface{}{}
	}
	var rowStrings []string
	for _, v := range data {
		rowStrings = append(rowStrings, " (?,?,?) ")
		param = append(param, maID, v.Name, v.Oids)
	}
	return fmt.Sprintf(_moduleInsertSQL, strings.Join(rowStrings, ",")), param
}

// BatchEditModuleSQL .
func BatchEditModuleSQL(mapModuel []*Module) (sql string, params []interface{}) {
	if len(mapModuel) == 0 {
		return "", []interface{}{}
	}
	var (
		nameSql, oidsSql string
		ids              []interface{}
		idSql            []string
	)
	for _, module := range mapModuel {
		nameSql += " WHEN id = ? THEN ?"
		params = append(params, module.ID, module.Name)
		ids = append(ids, module.ID)
		idSql = append(idSql, "?")
	}
	for _, module := range mapModuel {
		oidsSql += " WHEN id = ? THEN ?"
		params = append(params, module.ID, module.Oids)
	}
	params = append(params, ids...)
	return fmt.Sprintf(_moduleEditSQL, nameSql, oidsSql, strings.Join(idSql, ",")), params
}

// BatchAddActLiveSQL .
func BatchAddActLiveSQL(maID int64, data []*Activelive) (sql string, param []interface{}) {
	if len(data) == 0 {
		return "", []interface{}{}
	}
	var rowStrings []string
	for _, v := range data {
		rowStrings = append(rowStrings, "(?,?,?)")
		param = append(param, maID, v.Title, v.LiveId)
	}
	return fmt.Sprintf(_actLiveInsertSQL, strings.Join(rowStrings, ",")), param
}

// BatchEditActLiveSQL .
func BatchEditActLiveSQL(alDatas []*Activelive) (sql string, params []interface{}) {
	if len(alDatas) == 0 {
		return "", []interface{}{}
	}
	var (
		titleSql, liveSql string
		ids               []interface{}
		idSql             []string
	)
	for _, live := range alDatas {
		titleSql += " WHEN id = ? THEN ?"
		params = append(params, live.ID, live.Title)
		ids = append(ids, live.ID)
		idSql = append(idSql, "?")
	}
	for _, live := range alDatas {
		liveSql += " WHEN id = ? THEN ?"
		params = append(params, live.ID, live.LiveId)
	}
	params = append(params, ids...)
	return fmt.Sprintf(_actLiveEditSQL, titleSql, liveSql, strings.Join(idSql, ",")), params
}
