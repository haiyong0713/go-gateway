package model

import (
	"encoding/json"
)

const (
	// binlog action
	ActionUpdate = "update"
	ActionInsert = "insert"
	ActionDelete = "delete"
	// table
	TableWhiteList    = "white_list"
	TableNatUserSpace = "native_user_space"
)

var (
	Tables = map[string]struct{}{
		TableWhiteList:    {},
		TableNatUserSpace: {},
	}
)

type BinlogMsg struct {
	Action string          `json:"action"`
	Table  string          `json:"table"`
	New    json.RawMessage `json:"new"`
	Old    json.RawMessage `json:"old"`
}

type WhiteList struct {
	ID          int      `gorm:"AUTO_INCREMENT;column:id;type:INT;primary_key" json:"id"`
	Mid         int64    `gorm:"column:mid;type:BIGINT;default:0;" json:"mid"`
	Creator     string   `gorm:"column:creator;type:VARCHAR;size:32;" json:"creator"`
	CreatorUID  int      `gorm:"column:creator_uid;type:INT;default:0;" json:"creator_uid"`
	Modifier    string   `gorm:"column:modifier;type:VARCHAR;size:32;" json:"modifier"`
	ModifierUID int      `gorm:"column:modifier_uid;type:INT;default:0;" json:"modifier_uid"`
	FromType    string   `gorm:"column:from_type;type:VARCHAR;size:32;" json:"from_type"`
	State       int      `gorm:"column:state;type:TINYINT;default:0;" json:"state"`
	Ctime       DateTime `gorm:"column:ctime;type:DATETIME;default:CURRENT_TIMESTAMP;" json:"ctime"`
	Mtime       DateTime `gorm:"column:mtime;type:DATETIME;default:CURRENT_TIMESTAMP;" json:"mtime"`
}

type NativeUserSpace struct {
	Id           int64    `json:"id"`
	Mid          int64    `json:"mid"`
	Title        string   `json:"title"`
	PageId       int64    `json:"page_id"`
	DisplaySpace int64    `json:"display_space"`
	State        string   `json:"state"`
	Ctime        DateTime `json:"ctime"`
	Mtime        DateTime `json:"mtime"`
}
