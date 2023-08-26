package config

import (
	"go-gateway/app/app-svr/native-act/interface/kernel"
)

type ResourceTopic struct {
	BaseCfgManager

	ResourceCommon
	BriefDynsReqID kernel.RequestID
}
