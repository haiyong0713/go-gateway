package model

// VlogParam .
type VlogParam struct {
	TID         int64  `form:"tid" validate:"required"`
	ChnID       int64  `form:"chn_id" validate:"required"`
	Build       int32  `form:"build" validate:"required"`
	Buvid       string `form:"buvid" validate:"required"`
	Ps          int32  `form:"ps" validate:"min=0,max=50" default:"20"`
	Pn          int32  `form:"pn" validate:"min=1" default:"1"`
	Plat        int32  `form:"plat"`
	Rank        int32  `form:"rank"`
	MID         int64  `form:"-"`
	LoginEnvent int32  `form:"-"`
}

// VlogRankParam
type VlogRankParam struct {
	TID int64 `form:"tid" validate:"required"`
	Ps  int32 `form:"ps" validate:"min=1,max=50" default:"20"`
	Pn  int32 `form:"pn" validate:"min=1" default:"1"`
}
