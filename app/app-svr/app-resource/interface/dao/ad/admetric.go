package ad

import (
	"go-gateway/app/app-svr/app-resource/interface/model/splash"
)

//go:generate easyjson -all admetric.go

//easyjson:json
type SplashListData struct {
	Code int `json:"code"`
	*splash.CmConfig
	RequestID string         `json:"request_id"`
	Data      []*splash.List `json:"data"`
}
