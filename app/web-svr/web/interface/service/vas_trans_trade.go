package service

import (
	"context"
	"go-gateway/app/web-svr/web/interface/model"

	vtradegrpc "git.bilibili.co/bapis/bapis-go/vas/trans/trade/service"
)

func (s *Service) TradeCreate(ctx context.Context, req *model.TradeCreateReq) (res *vtradegrpc.TradeCreateReply, err error) {
	return s.dao.TradeCreate(ctx, req)
}
