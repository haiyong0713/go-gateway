package domain

import xtime "go-common/library/time"

type Record struct {
	Id           int64      `form:"id" json:"id" validate:"required"`
	ActName      string     `form:"act_name" json:"act_name" validate:"required"`
	PageLink     string     `form:"page_link" json:"page_link" validate:"required"`
	FirstDomain  string     `form:"first_domain" json:"first_domain" validate:"required"`
	SecondDomain string     `form:"second_domain" json:"second_domain"  default:"" `
	Stime        xtime.Time `form:"stime"  json:"stime" validate:"required"`
	Etime        xtime.Time `form:"etime"  json:"etime" validate:"required"`
	Ctime        xtime.Time `json:"ctime"`
	Mtime        xtime.Time `json:"mtime"`
}

type AddDomainParam struct {
	ActName      string     `form:"act_name" json:"act_name" validate:"required"`
	PageLink     string     `form:"page_link" json:"page_link" validate:"required"`
	FirstDomain  string     `form:"first_domain" json:"first_domain" validate:"required"`
	SecondDomain string     `form:"second_domain" json:"second_domain"  default:"" `
	Stime        xtime.Time `form:"stime"  json:"stime" validate:"required"`
	Etime        xtime.Time `form:"etime"  json:"etime" validate:"required"`
}

type Search struct {
	Id       int64  `form:"id" json:"id" `
	ActName  string `form:"act_name" json:"act_name"`
	PageNo   int    `form:"page_no" json:"page_no" default:"1"`
	PageSize int    `form:"page_size" json:"page_size" default:"50"`
}

type UpdateRespone struct {
	Rows int64 `json:"rows"`
}
