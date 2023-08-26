package model

type MemberInReq struct {
	Business  string `form:"business" validate:"required"`
	Name      string `form:"name"`                     // names 人群包名称
	Version   int64  `form:"version" validate:"min=0"` // version 人群包版本，默认不传或者小于等于0，都是使用人群包当前最新的
	Dimension int    `form:"dinmension" default:"1" validate:"min=1,max=2"`
	Mid       int64  `form:"-"`
	Buvid     string `form:"-"`
}
