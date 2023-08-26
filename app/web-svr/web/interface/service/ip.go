package service

import (
	"context"

	"go-common/library/log"
	"go-common/library/net/metadata"

	api "git.bilibili.co/bapis/bapis-go/community/service/location"
)

// IPZone get ip zone info by ip
func (s *Service) IPZone(c context.Context) (res *api.InfoReply, err error) {
	ip := metadata.String(c, metadata.RemoteIP)
	if res, err = s.locGRPC.Info(c, &api.InfoReq{Addr: ip}); err != nil {
		log.Error("s.locGRPC.Info(%s) error(%v)", ip, err)
	}
	return
}
