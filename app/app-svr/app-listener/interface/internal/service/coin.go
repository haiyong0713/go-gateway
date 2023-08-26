package service

import (
	"context"

	v1 "go-gateway/app/app-svr/app-listener/interface/api/v1"
	"go-gateway/app/app-svr/app-listener/interface/internal/dao"
	mainErr "go-gateway/ecode"

	"github.com/pkg/errors"
)

func (s *Service) CoinAdd(ctx context.Context, req *v1.CoinAddReq) (resp *v1.CoinAddResp, err error) {
	if err = validatePlayItem(ctx, req.Item, 1); err != nil || req.Num <= 0 {
		return
	}
	resp = new(v1.CoinAddResp)

	dev, net, auth := DevNetAuthFromCtx(ctx)
	err = s.dao.CoinAdd(ctx, dao.CoinAddOpt{
		Mid: auth.Mid, Dev: dev, Net: net,
		ItemType: req.Item.ItemType,
		Oid:      req.Item.Oid, SubID: req.Item.SubId[0],
		CoinNum: req.Num, ThumbUp: req.ThumbUp,
	})
	if err != nil {
		if err == dao.ErrSilverBulletHit {
			return nil, errors.WithMessagef(mainErr.SilverBulletCoinReject, "silverBullet rejected coinAdd: mid(%d) req(%+v)", auth.Mid, req)
		}
	}
	resp.Message = s.C.Res.Text.CoinOK
	return
}
