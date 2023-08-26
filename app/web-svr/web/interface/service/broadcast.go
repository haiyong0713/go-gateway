package service

import (
	"context"

	"go-common/library/log"
	"go-gateway/app/web-svr/web/interface/model"

	warden "git.bilibili.co/bapis/bapis-go/infra/service/broadcast"
)

// BroadServers broadcast server list.
func (s *Service) BroadServers(c context.Context, platform string) (res *warden.ServerListReply, err error) {
	if res, err = s.broadcastGRPC.ServerList(c, &warden.ServerListReq{Platform: platform}); err != nil {
		log.Error("s.broadCastGRPC.ServerList(%s) error(%v)", platform, err)
		res = model.DefaultServer
		err = nil
	}
	return
}
