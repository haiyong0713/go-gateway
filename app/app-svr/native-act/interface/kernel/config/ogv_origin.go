package config

import (
	"go-gateway/app/app-svr/native-act/interface/kernel"
)

type OgvOrigin struct {
	BaseCfgManager

	OgvCommon
	PlaylistID        int32
	SeasonByPlayIdReq kernel.RequestID
}
