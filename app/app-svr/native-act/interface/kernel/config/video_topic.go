package config

import (
	"go-gateway/app/app-svr/native-act/interface/kernel"
)

type VideoTopic struct {
	BaseCfgManager

	VideoCommon
	BriefDynsReqID kernel.RequestID
}
