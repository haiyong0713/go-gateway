package config

import (
	"go-gateway/app/app-svr/native-act/interface/kernel"
)

type NewactStatement struct {
	BaseCfgManager

	Sid   int64 //数据源id
	Type  int64 //文本类型
	ReqID kernel.RequestID
}
