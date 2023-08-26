package unicom

import (
	"time"

	"go-gateway/app/app-svr/app-wall/interface/model"
)

type PackMsg struct {
	Action model.Action `json:"action"`
	Data   *PackData    `json:"data"`
}

type PackData struct {
	Kind       model.ConsumeKind `json:"kind"`
	Phone      int               `json:"phone"`
	Mid        int64             `json:"mid"`
	Usermob    string            `json:"usermob"`
	Name       string            `json:"Name"`
	Integral   int               `json:"integral"`
	Flow       int               `json:"flow"`
	OrderID    string            `json:"order_id"`
	OutorderID string            `json:"outorder_id"`
	PackID     int64             `json:"pack_id"`
	Desc       string            `json:"desc"`
	Type       int               `json:"type"`
	Param      string            `json:"param"`
	Ctime      time.Time         `json:"ctime"`
	Stime      time.Time         `json:"stime"`
	NewParam   string            `json:"new_param"`
}
