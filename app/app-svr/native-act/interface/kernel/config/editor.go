package config

import (
	"go-gateway/app/app-svr/native-act/interface/kernel"
)

type Editor struct {
	BaseCfgManager

	Position          Position
	DisplayMoreButton bool
	BgColor           string
	MixExtsReqID      kernel.RequestID
	GetHisReqID       kernel.RequestID
}

type Position struct {
	Position1 string
	Position2 string
	Position3 string
	Position4 string
	Position5 string
}
