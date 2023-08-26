package model

import "go-gateway/app/app-svr/archive/service/api"

// bvid开关
const BvOpen = 1

func ClearAttrAndAccess(in *api.Arc) {
	if in == nil {
		return
	}
	in.Attribute = 0
	in.AttributeV2 = 0
	in.Access = 0
}
