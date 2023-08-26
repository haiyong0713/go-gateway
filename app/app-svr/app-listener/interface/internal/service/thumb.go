package service

import (
	"context"
	"strings"

	"go-common/library/ecode"
	v1 "go-gateway/app/app-svr/app-listener/interface/api/v1"
	"go-gateway/app/app-svr/app-listener/interface/internal/dao"
	"go-gateway/app/app-svr/app-listener/interface/internal/model"
	mainErr "go-gateway/ecode"

	coinErr "git.bilibili.co/bapis/bapis-go/community/service/coin/ecode"
	"github.com/pkg/errors"
	"go-common/library/sync/errgroup.v2"
)

func (s *Service) ThumbUp(ctx context.Context, req *v1.ThumbUpReq) (resp *v1.ThumbUpResp, err error) {
	if err = validatePlayItem(ctx, req.Item, 1); err != nil {
		return
	}
	resp = new(v1.ThumbUpResp)

	dev, net, auth := DevNetAuthFromCtx(ctx)
	err = s.dao.ThumbAction(ctx, dao.ThumbActionOpt{
		Mid: auth.Mid, Dev: dev,
		Net:      net,
		ItemType: req.Item.ItemType,
		Oid:      req.Item.Oid, SubID: req.Item.SubId[0],
		Action: req.Action,
	})
	if err != nil {
		if err == dao.ErrSilverBulletHit {
			return nil, errors.WithMessagef(mainErr.SilverBulletLikeReject, "silverBullet rejected thumbup: mid(%d) req(%+v)", auth.Mid, req)
		}
	}
	if req.Action == v1.ThumbUpReq_LIKE {
		resp.Message = s.C.Res.Text.ThumbUp
	} else {
		resp.Message = s.C.Res.Text.ThumbCancel
	}
	return
}

func (s *Service) TripleLike(ctx context.Context, req *v1.TripleLikeReq) (resp *v1.TripleLikeResp, err error) {
	if err = validatePlayItem(ctx, req.Item, 1); err != nil {
		return
	}
	resp = new(v1.TripleLikeResp)

	dev, net, auth := DevNetAuthFromCtx(ctx)
	// 点赞+风控
	err = s.dao.ThumbAction(ctx, dao.ThumbActionOpt{
		Mid: auth.Mid, Dev: dev, Net: net,
		ItemType: req.Item.ItemType, Oid: req.Item.Oid, SubID: req.Item.SubId[0],
		Action: v1.ThumbUpReq_LIKE, IsTripleLike: true,
	})
	if err != nil {
		if err == dao.ErrSilverBulletHit {
			return nil, errors.WithMessagef(mainErr.SilverBulletLikeReject, "silverBullet rejected tripleLike: mid(%d) req(%+v)", auth.Mid, req)
		}
		return
	}
	resp.ThumbOk = true

	// 投币/收藏就不走风控了
	eg := errgroup.WithContext(ctx)
	eg.Go(func(c context.Context) error {
		err := s.dao.CoinAdd(c, dao.CoinAddOpt{
			Mid: auth.Mid, Dev: dev, Net: net,
			ItemType: req.Item.ItemType, Oid: req.Item.Oid, SubID: req.Item.SubId[0],
			NoSilver: true, CoinNum: 0, // 自动计算
		})
		if err == nil || ecode.EqualError(coinErr.CoinOverMax, err) {
			resp.CoinOk = true
		}
		return nil
	})
	eg.Go(func(c context.Context) error {
		// 直接收藏到默认收藏夹
		err := s.dao.FavItemAdd(c, dao.FavItemAddOpt{
			Meta: model.FavItemAddAndDelMeta{
				Mid: auth.Mid, Tp: model.FavTypeVideo,
				Oid: req.Item.Oid, Otype: model.Play2Fav[req.Item.ItemType],
			},
			Device: dev, Network: net, NoSilver: true,
		})
		if err == nil {
			resp.FavOk = true
		}
		return nil
	})
	err = eg.Wait()

	if resp.ThumbOk && resp.CoinOk && resp.FavOk {
		resp.Message = s.C.Res.Text.TripleLike // 三连成功信息
	} else {
		if !resp.ThumbOk && !resp.CoinOk && !resp.FavOk {
			resp.Message = "三连失败"
		} else {
			failedActs := make([]string, 0, 2)
			for _, o := range []struct {
				Text   string
				Failed bool
			}{
				{"点赞", !resp.ThumbOk},
				{"投币", !resp.CoinOk},
				{"收藏", !resp.FavOk},
			} {
				if o.Failed {
					failedActs = append(failedActs, o.Text)
				}
			}
			resp.Message = "未完成三连，" + strings.Join(failedActs, "") + "失败"
		}
	}

	return
}
