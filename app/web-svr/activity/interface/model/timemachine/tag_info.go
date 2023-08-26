package timemachine

import (
	"time"
)

// bcurd -dsn='main_lottery:RWXcNcw1X43S1K1nvB51x7iNjMMjP0ba@tcp(10.221.34.182:4000)/main_lottery?parseTime=true'  -schema=main_lottery -table=user_report_2020_tag_info -tmpl=bilibili_log.tmpl > tag_info.go

// UserReport2020TagInfo represents a row from 'user_report_2020_tag_info'.
type UserReport2020TagInfo struct {
	ID          int32     `json:"id"`          //
	TagName     string    `json:"tag_name"`    //
	Display     string    `json:"display"`     //
	Description string    `json:"description"` //
	Img         string    `json:"img"`         //
	Ctime       time.Time `json:"ctime"`       //
	Mtime       time.Time `json:"mtime"`       //
}
