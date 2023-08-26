package show

import (
	"fmt"
	"strings"

	xtime "go-common/library/time"
	"go-gateway/app/app-svr/app-feed/admin/model/common"
)

const (
	_specialQueryInsertSQL  = "INSERT INTO search_web_special_query(sid,value) VALUES %s"
	_specialModuleInsertSQL = "INSERT INTO search_web_special_module(sid,value,`order`) VALUES %s"
	_specialQueryEditSQL    = "UPDATE search_web_special_query SET value = CASE %s END WHERE id IN (%s)"
	_specialModuleEditSQL   = "UPDATE search_web_special_module SET value = CASE %s END,`order` = CASE %s END WHERE id IN (%s)"
)

// SearchWebModule .
type SearchWebModule struct {
	ID       int64                    `gorm:"column:id" json:"id"`
	Reason   string                   `gorm:"column:reason" json:"reason"`
	UserName string                   `gorm:"column:uname" json:"uname"`
	Check    int64                    `gorm:"column:check" json:"check"`
	Ctime    xtime.Time               `gorm:"column:ctime" json:"ctime"`
	Mtime    xtime.Time               `gorm:"column:mtime" json:"mtime"`
	Querys   []*SearchWebModuleQuery  `gorm:"-" json:"query"`
	Modules  []*SearchWebModuleModule `gorm:"-" json:"modules"`
}

// SearchWebModuleQuery .
type SearchWebModuleQuery struct {
	ID      int64      `gorm:"column:id" json:"id"`
	Sid     int64      `gorm:"column:sid" json:"sid"`
	Value   string     `gorm:"column:value" json:"value"`
	Deleted int64      `gorm:"column:deleted" json:"deleted"`
	Ctime   xtime.Time `gorm:"column:ctime" json:"ctime"`
	Mtime   xtime.Time `gorm:"column:mtime" json:"mtime"`
}

// SearchWebModuleModule .
type SearchWebModuleModule struct {
	ID      int64      `gorm:"column:id" json:"id"`
	Sid     int64      `gorm:"column:sid" json:"sid"`
	Order   int        `gorm:"column:order" json:"order"`
	Value   string     `gorm:"column:value" json:"value"`
	Deleted int64      `gorm:"column:deleted" json:"deleted"`
	Ctime   xtime.Time `gorm:"column:ctime" json:"ctime"`
	Mtime   xtime.Time `gorm:"column:mtime" json:"mtime"`
}

// SearchWebModuleAP add param
type SearchWebModuleAP struct {
	ID       int64      `gorm:"column:id"`
	Reason   string     `gorm:"column:reason" form:"reason"`
	Check    int        `gorm:"column:check" form:"check"`
	UserName string     `gorm:"column:uname"`
	Ctime    xtime.Time `gorm:"column:ctime"`
	Mtime    xtime.Time `gorm:"column:mtime"`
	Query    string     `form:"query" gorm:"-" validate:"required"`
	Module   string     `form:"modules" gorm:"-" validate:"required"`
}

// SearchWebModuleLP list param
type SearchWebModuleLP struct {
	Check int    `form:"check"`
	Query string `form:"query"`
	Ps    int    `form:"ps" default:"20" validate:"max=30"`
	Pn    int    `form:"pn" default:"1"`
}

// SearchWebModulePager .
type SearchWebModulePager struct {
	Item []*SearchWebModule `json:"item"`
	Page common.Page        `json:"page"`
}

// SearchWebModuleOption option web card
type SearchWebModuleOption struct {
	ID       int64  `form:"id" validate:"required"`
	Check    int    `form:"check" validate:"required"`
	UserName string `gorm:"-"`
	UID      int64  `gorm:"-"`
}

// SearchWebModuleAP update param
type SearchWebModuleUP struct {
	ID       int64      `gorm:"column:id" form:"id" validate:"required"`
	Reason   string     `form:"reason" gorm:"column:reason"`
	UserName string     `gorm:"-" json:"uname"`
	Ctime    xtime.Time `gorm:"column:ctime"`
	Mtime    xtime.Time `gorm:"column:mtime"`
	Query    string     `form:"query" gorm:"-" validate:"required"`
	Module   string     `form:"modules" gorm:"-" validate:"required"`
}

// SpecialQuerySQL .
func SpecialQuerySQL(sID int64, data []*SearchWebModuleQuery) (sql string, param []interface{}) {
	if len(data) == 0 {
		return "", []interface{}{}
	}
	var rowStrings []string
	for _, v := range data {
		rowStrings = append(rowStrings, "(?,?)")
		param = append(param, sID, v.Value)
	}
	return fmt.Sprintf(_specialQueryInsertSQL, strings.Join(rowStrings, ",")), param
}

// SpecialQueryUpSQL .
func SpecialQueryUpSQL(querys []*SearchWebModuleQuery) (sql string, params []interface{}) {
	if len(querys) == 0 {
		return "", []interface{}{}
	}
	var (
		ids   []interface{}
		idSQL []string
	)
	for _, query := range querys {
		sql = sql + " WHEN id = ? THEN ? "
		params = append(params, query.ID, query.Value)
		idSQL = append(idSQL, "?")
		ids = append(ids, query.ID)
	}
	params = append(params, ids...)
	return fmt.Sprintf(_specialQueryEditSQL, sql, strings.Join(idSQL, ",")), params
}

// SpecialModuleUpSQL .
func SpecialModuleUpSQL(querys []*SearchWebModuleModule) (sql string, params []interface{}) {
	if len(querys) == 0 {
		return "", []interface{}{}
	}
	var (
		ids                 []interface{}
		idSql               []string
		valuesSQL, orderSQL string
	)
	for _, query := range querys {
		valuesSQL += valuesSQL + " WHEN id = ? THEN ? "
		params = append(params, query.ID, query.Value)
		idSql = append(idSql, "?")
		ids = append(ids, query.ID)
	}
	for _, query := range querys {
		orderSQL += " WHEN id = ? THEN ?"
		params = append(params, query.ID, query.Order)
	}
	params = append(params, ids...)
	return fmt.Sprintf(_specialModuleEditSQL, valuesSQL, orderSQL, strings.Join(idSql, ",")), params
}

// SpecialModuleSQL .
func SpecialModuleSQL(sID int64, data []*SearchWebModuleModule) (sql string, param []interface{}) {
	if len(data) == 0 {
		return "", []interface{}{}
	}
	var rowStrings []string
	for _, v := range data {
		rowStrings = append(rowStrings, "(?,?,?)")
		param = append(param, sID, v.Value, v.Order)
	}
	return fmt.Sprintf(_specialModuleInsertSQL, strings.Join(rowStrings, ",")), param
}

// TableName .
func (a SearchWebModule) TableName() string {
	return "search_web_special"
}

// TableName .
func (a SearchWebModuleUP) TableName() string {
	return "search_web_special"
}

// TableName .
func (a SearchWebModuleAP) TableName() string {
	return "search_web_special"
}

// TableName .
func (a SearchWebModuleQuery) TableName() string {
	return "search_web_special_query"
}

// TableName .
func (a SearchWebModuleModule) TableName() string {
	return "search_web_special_module"
}

// TableName .
func (a SearchWebModuleOption) TableName() string {
	return "search_web_special"
}
