package question

import "go-common/library/time"

// state.
const (
	StateInit     = 0
	StateOnline   = 1
	StateOffline  = 2
	State4Process = 3
)

// AddDetailArg add detail arg.
type AddDetailArg struct {
	BaseID      int64  `form:"base_id" validate:"min=1"`
	Name        string `form:"name" validate:"min=1"`
	RightAnswer string `form:"right_answer" validate:"min=1"`
	WrongAnswer string `form:"wrong_answer" validate:"min=1"`
	Attribute   int64  `form:"attribute"`
	State       int64  `form:"state"`
	Pic         string `form:"pic"`
}

// SaveDetailArg save detail arg.
type SaveDetailArg struct {
	ID int64 `form:"id" validate:"min=1"`
	AddDetailArg
}

// Detail .
type Detail struct {
	ID          int64     `json:"id" gorm:"column:id"`
	BaseID      int64     `json:"base_id" gorm:"column:base_id"`
	Name        string    `json:"name" gorm:"column:name"`
	RightAnswer string    `json:"right_answer" gorm:"column:right_answer"`
	WrongAnswer string    `json:"wrong_answer" gorm:"column:wrong_answer"`
	Attribute   int64     `json:"attribute" gorm:"column:attribute"`
	State       int64     `json:"state" gorm:"column:state"`
	Ctime       time.Time `json:"ctime" gorm:"column:ctime" time_format:"2006-01-02 15:04:05"`
	Mtime       time.Time `json:"mtime" gorm:"column:mtime" time_format:"2006-01-02 15:04:05"`
	Pic         string    `json:"pic" gorm:"column:pic"`
}

// TableName .
func (Detail) TableName() string {
	return "question_detail"
}
