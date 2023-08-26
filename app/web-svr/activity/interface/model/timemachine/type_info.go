package timemachine

import (
	"time"
)

// bcurd -dsn='main_lottery:RWXcNcw1X43S1K1nvB51x7iNjMMjP0ba@tcp(10.221.34.182:4000)/main_lottery?parseTime=true'  -schema=main_lottery -table=user_report_2020_type_info -tmpl=bilibili_log.tmpl > type_info.go

// UserReport2020TypeInfo represents a row from 'user_report_2020_type_info'.
type UserReport2020TypeInfo struct {
	Tid         int64     `json:"tid"`          //
	Pid         int64     `json:"pid"`          //
	TidName     string    `json:"tid_name"`     //
	SubTidName  string    `json:"sub_tid_name"` //
	Display     string    `json:"display"`      //
	Description string    `json:"description"`  //
	Img         string    `json:"img"`          //
	Ctime       time.Time `json:"ctime"`        //
	Mtime       time.Time `json:"mtime"`        //
}
