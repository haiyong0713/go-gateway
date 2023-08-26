package model

type ParamAdd struct {
	Cid    int64  `form:"cid" validate:"required"`
	Groups string `form:"groups" validate:"required"`
}

type ParamDel struct {
	Cid    int64 `form:"cid" validate:"required"`
	MainID int64 `form:"main_id" validate:"required"`
}

type ParamRes struct {
	Cid      int64 `form:"cid" validate:"required"`
	MainID   int64 `form:"main_id" validate:"required"`
	DetailID int64 `form:"detail_id" validate:"required"`
}
