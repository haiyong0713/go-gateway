package region

import "go-gateway/app/app-svr/app-car/interface/model"

type RegionParam struct {
	model.DeviceInfo
	ParamStr string `form:"param"`
	Rid      int64  `form:"rid"`
	Pn       int    `form:"pn" default:"1" validate:"min=1"`
	Ps       int    `form:"ps" default:"20" validate:"min=1,max=20"`
	FromType string `form:"from_type"`
}
