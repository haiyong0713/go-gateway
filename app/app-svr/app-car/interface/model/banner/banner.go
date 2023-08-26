package banner

import "go-gateway/app/app-svr/app-car/interface/model"

type Banner struct {
	ID    int64  `json:"id,omitempty"`
	URI   string `json:"uri,omitempty"`
	Image string `json:"image,omitempty"`
}

type ShowBannerParam struct {
	model.DeviceInfo
}
