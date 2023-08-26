package model

import "go-common/library/time"

type SpaceOfficial struct {
	ID         int64     `form:"id" json:"id"`
	Uid        int64     `form:"uid" json:"uid" validate:"required,min=1"`
	Name       string    `form:"name" json:"name" validate:"required"`
	Icon       string    `form:"icon" json:"icon" validate:"required,max=128"`
	Scheme     string    `form:"scheme" json:"scheme" validate:"required"`
	Rcmd       string    `form:"rcmd" json:"rcmd" validate:"required"`
	IosUrl     string    `form:"ios_url" json:"ios_url" validate:"required,max=512"`
	AndroidUrl string    `form:"android_url" json:"android_url" validate:"required,max=512"`
	Button     string    `form:"button" json:"button" validate:"required"`
	Deleted    int64     `form:"deleted" json:"deleted"`
	Ctime      time.Time `form:"ctime" json:"ctime"`
	Mtime      time.Time `form:"mtime" json:"mtime"`
}

// TableName .
func (a SpaceOfficial) TableName() string {
	return "space_official"
}

// SpaceOfficialParam .
type SpaceOfficialParam struct {
	Ps int `form:"ps" default:"20"` // 分页大小
	Pn int `form:"pn" default:"1"`  // 第几个分页
}

// SpaceOfficialPager .
type SpaceOfficialPager struct {
	Item []*SpaceOfficial `json:"item"`
	Page Page             `json:"page"`
}
