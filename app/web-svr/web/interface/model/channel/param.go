package channel

// Param is
type Param struct {
	Offset    string `form:"offset" default:""`
	MID       int64  `form:"mid"`
	ChannelID int64  `form:"channel_id" validate:"gt=0"`
	Sort      string `form:"sort"`
	Theme     string `form:"theme"`
}

// ParamHot login hot list.
type ParamHot struct {
	Offset string `form:"offset"`
	Ps     int32  `form:"ps"`
	Count  int32  `form:"count"`
}

// ParamSort  subscribe top param.
type ParamSort struct {
	Action int32  `form:"action" validate:"min=1" default:"2"`
	Tops   string `form:"tops"`
	Cids   string `form:"cids"`
}
