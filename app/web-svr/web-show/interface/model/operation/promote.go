package operation

import "go-gateway/app/app-svr/archive/service/api"

// Promote strcuct
type Promote struct {
	IsAd    int8     `json:"is_ad"`
	Archive *api.Arc `json:"archive"`
}

// ArgPromote strcuct
type ArgPromote struct {
	Tp    string `form:"tp" validate:"required"`
	Count int    `form:"count" validate:"min=0"`
	Rank  int    `form:"rank" validate:"min=0"`
}
