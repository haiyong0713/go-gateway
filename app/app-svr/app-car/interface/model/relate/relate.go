package relate

import (
	"go-gateway/app/app-svr/app-car/interface/model"
	cardm "go-gateway/app/app-svr/app-car/interface/model/card"
)

type RelateParam struct {
	model.DeviceInfo
	ParamStr string `form:"param"`
}

type Item struct {
	Items  []cardm.Handler `json:"items,omitempty"`
	Relate *Relate         `json:"relate,omitempty"`
}

type Relate struct {
	Title string          `json:"title,omitempty"`
	Items []cardm.Handler `json:"items,omitempty"`
}
