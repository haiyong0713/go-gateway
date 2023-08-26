package dynamicV2

import (
	"context"

	"go-common/library/sync/errgroup.v2"
	api "go-gateway/app/app-svr/app-dynamic/interface/api/v2"
	mdlv2 "go-gateway/app/app-svr/app-dynamic/interface/model/dynamicV2"

	dyngrpc "git.bilibili.co/bapis/bapis-go/dynamic/service/feed"
)

func (s *Service) LbsPoiList(c context.Context, general *mdlv2.GeneralParam, req *api.LbsPoiReq) (*api.LbsPoiReply, error) {
	var (
		poilist   *mdlv2.DynListRes
		poidetail *dyngrpc.LbsPoiDetailRsp
	)
	eg := errgroup.WithCancel(c)
	eg.Go(func(ctx context.Context) (err error) {
		if poilist, err = s.dynDao.LbsPoiList(ctx, general, req); err != nil {
			return err
		}
		return nil
	})
	if req.RefreshType == api.Refresh_refresh_new {
		eg.Go(func(ctx context.Context) (err error) {
			if poidetail, err = s.dynDao.LbsPoiDetail(ctx, req); err != nil {
				return err
			}
			return nil
		})
	}
	if err := eg.Wait(); err != nil {
		return nil, err
	}
	res := &api.LbsPoiReply{
		Offset:  poilist.HistoryOffset,
		HasMore: poilist.HasMore,
	}
	if poidetail != nil {
		res.Detail = &api.LbsPoiDetail{
			Poi:     poidetail.Poi,
			Type:    int64(poidetail.Type),
			BasePic: poidetail.BasePic,
			Cover:   poidetail.Cover,
			Address: poidetail.Address,
			Title:   poidetail.Title,
		}
	}
	dynCtx, err := s.getMaterial(c, getMaterialOption{general: general, dynamics: poilist.Dynamics})
	if err != nil {
		return nil, err
	}
	// Step 4. 对物料信息处理，获取详情列表
	foldList := s.procListReply(c, poilist.Dynamics, dynCtx, general, _handleTypeLBS)
	// Step 5. 回填
	s.procBackfill(c, dynCtx, general, foldList)
	// Step 6. 折叠判断
	retDynList := s.procFold(foldList, dynCtx, general)
	res.List = retDynList
	return res, nil
}
