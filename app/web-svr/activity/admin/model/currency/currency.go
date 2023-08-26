package currency

import "go-common/library/time"

// AddArg add arg.
type AddArg struct {
	Name  string `form:"name" validate:"min=1"`
	Unit  string `form:"unit"`
	State int64  `form:"state" validate:"min=0"`
}

// SaveArg save arg.
type SaveArg struct {
	ID    int64  `form:"id" validate:"min=1"`
	Name  string `form:"name" validate:"min=1"`
	Unit  string `form:"unit"`
	State int64  `form:"state" validate:"min=0"`
}

// CurrItem currency list item.
type CurrItem struct {
	*Currency
	Relation []*Relation `json:"relation"`
}

// Currency data model.
type Currency struct {
	ID    int64     `json:"id" gorm:"column:id"`
	Name  string    `json:"name" gorm:"column:name"`
	Unit  string    `json:"unit" gorm:"column:unit"`
	State int64     `json:"state" gorm:"column:state"`
	Ctime time.Time `json:"ctime" gorm:"column:ctime" time_format:"2006-01-02 15:04:05"`
	Mtime time.Time `json:"mtime" gorm:"column:mtime" time_format:"2006-01-02 15:04:05"`
}

// Currency data model.
type CurrencyUser struct {
	ID     int64     `json:"id" gorm:"column:id"`
	Mid    int64     `json:"mid" gorm:"column:mid"`
	Amount int64     `json:"amount" gorm:"column:amount"`
	Ctime  time.Time `json:"ctime" gorm:"column:ctime" time_format:"2006-01-02 15:04:05"`
	Mtime  time.Time `json:"mtime" gorm:"column:mtime" time_format:"2006-01-02 15:04:05"`
}

// RelationArg currency relation change arg.
type RelationArg struct {
	CurrencyID int64 `form:"currency_id" validate:"min=1"`
	BusinessID int64 `form:"business_id" validate:"min=1"`
	ForeignID  int64 `form:"foreign_id" validate:"min=1"`
}

// Relation currency relation model.
type Relation struct {
	ID         int64     `json:"id" gorm:"column:id"`
	CurrencyID int64     `json:"currency_id" gorm:"column:currency_id"`
	BusinessID int64     `json:"business_id" gorm:"column:business_id"`
	ForeignID  int64     `json:"foreign_id" gorm:"column:foreign_id"`
	IsDeleted  int       `json:"-" gorm:"column:is_deleted"`
	Ctime      time.Time `json:"ctime" gorm:"column:ctime" time_format:"2006-01-02 15:04:05"`
	Mtime      time.Time `json:"mtime" gorm:"column:mtime" time_format:"2006-01-02 15:04:05"`
}

func (CurrencyUser) TableName() string {
	return "currency_user"
}

// TableName currency def.
func (Currency) TableName() string {
	return "currency"
}

// TableName currency_relation def.
func (Relation) TableName() string {
	return "currency_relation"
}
