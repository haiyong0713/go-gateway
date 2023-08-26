package model

import (
	pb "go-gateway/app/app-svr/kvo/interface/api"
)

type PlayerConfig struct {
	Mid      int64
	Body     *pb.DmPlayerConfigReq
	Action   string
	Platform string
	Buvid    Buvid
}
