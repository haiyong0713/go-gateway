package manager

import (
	"go-common/library/time"
)

// PosRecUserMgt .
type PosRecUserMgt struct {
	ID            int       `gorm:"column:id" json:"id"`
	Pid           int       `gorm:"column:pid" json:"pid"`
	Type          int64     `gorm:"column:type" json:"type"`
	Name          string    `gorm:"column:name" json:"name"`
	Description   string    `gorm:"column:description" json:"description"`
	Ctime         time.Time `gorm:"column:ctime" json:"ctime"`
	Mtime         time.Time `gorm:"column:mtime" json:"mtime"`
	Limitoriented int64     `gorm:"column:limitoriented" json:"limitoriented"`
	Limitall      int64     `gorm:"column:limitall" json:"limitall"`
}

// TableName .
func (a PosRecUserMgt) TableName() string {
	return "pos_rec_user_mgt"
}
