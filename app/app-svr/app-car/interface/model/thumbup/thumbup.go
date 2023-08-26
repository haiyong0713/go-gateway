package thumbup

import (
	thumbup "git.bilibili.co/bapis/bapis-go/community/service/thumbup"
	"go-gateway/app/app-svr/app-car/interface/model"
)

type LikeReq struct {
	model.DeviceInfo
	Mid      int64
	Buvid    string
	UpMid    int64
	Business string
	MsgId    int64
	Action   thumbup.Action
	WithStat bool
}
