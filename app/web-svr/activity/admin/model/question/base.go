package question

import "go-common/library/time"

// AddBaseArg .
type AddBaseArg struct {
	BusinessID     int64     `form:"business_id" validate:"min=1"`
	ForeignID      int64     `form:"foreign_id" validate:"min=1"`
	Count          int64     `form:"count" validate:"min=1"`
	OneTs          int64     `form:"one_ts" validate:"min=1,max=86400"`
	RetryTs        int64     `form:"retry_ts" validate:"min=1,max=86400"`
	Name           string    `form:"name" validate:"min=1"`
	Separator      string    `form:"separator" default:","`
	DistributeType int64     `form:"distribute_type" default:"1"`
	Stime          time.Time `form:"stime" validate:"min=1"`
	Etime          time.Time `form:"etime" validate:"min=1"`
}

// SaveBaseArg .
type SaveBaseArg struct {
	ID int64 `form:"id" validate:"min=1"`
	AddBaseArg
}

// Base .
type Base struct {
	ID             int64     `json:"id" gorm:"column:id"`
	BusinessID     int64     `json:"business_id" gorm:"column:business_id"`
	ForeignID      int64     `json:"foreign_id" gorm:"column:foreign_id"`
	Count          int64     `json:"count" gorm:"column:count"`
	OneTs          int64     `json:"one_ts" gorm:"column:one_ts"`
	RetryTs        int64     `json:"retry_ts" gorm:"column:retry_ts"`
	Name           string    `json:"name" gorm:"column:name"`
	Separator      string    `json:"separator" gorm:"column:separator"`
	DistributeType int64     `json:"distribute_type" gorm:"column:distribute_type"`
	Stime          time.Time `json:"stime" gorm:"column:stime" time_format:"2006-01-02 15:04:05"`
	Etime          time.Time `json:"etime" gorm:"column:etime" time_format:"2006-01-02 15:04:05"`
	Ctime          time.Time `json:"ctime" gorm:"column:ctime" time_format:"2006-01-02 15:04:05"`
	Mtime          time.Time `json:"mtime" gorm:"column:mtime" time_format:"2006-01-02 15:04:05"`
}

// TableName .
func (Base) TableName() string {
	return "question_base"
}
