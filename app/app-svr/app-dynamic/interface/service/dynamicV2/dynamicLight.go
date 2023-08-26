package dynamicV2

import (
	"context"

	"go-common/library/log"
	api "go-gateway/app/app-svr/app-dynamic/interface/api/v2"
	mdlv2 "go-gateway/app/app-svr/app-dynamic/interface/model/dynamicV2"
	xmetric "go-gateway/app/app-svr/app-dynamic/interface/model/metric"
)

func (s *Service) DynLight(c context.Context, general *mdlv2.GeneralParam, req *api.DynLightReq) (*api.DynLightReply, error) {
	var (
		dynList *mdlv2.DynListRes
	)
	err := func(ctx context.Context) (err error) {
		dynTypeList := []string{"2"}
		if req.FromType == api.LightFromType_from_unlogin {
			dynList, err = s.dynDao.DynUnLoginLight(ctx, general, req, dynTypeList)
			if err != nil {
				xmetric.DynamicCoreAPI.Inc("未登录轻浏览页", "request_error")
				log.Error("DynLight s.dynDao.DynUnLoginLight mid(%d) error(%v)", req.FakeUid, err)
				return err
			}
		} else {
			// Step 1. 获取用户关注链信息(关注的up、追番、购买的课程）
			following, pgcFollowing, cheese, ugcSeason, batchListFavorite, err := s.followings(ctx, general.Mid, true, true, general)
			if err != nil {
				return err
			}
			attentions := mdlv2.GetAttentionsParams(general.Mid, following, pgcFollowing, cheese, ugcSeason, batchListFavorite)
			dynList, err = s.dynDao.DynLight(c, general, req, dynTypeList, attentions)
			if err != nil {
				xmetric.DynamicCoreAPI.Inc("轻浏览页", "request_error")
				log.Error("DynLight s.dynDao.DynLight mid(%d) error(%v)", general.Mid, err)
				return err
			}
		}
		return nil
	}(c)
	if err != nil {
		return nil, err
	}
	// Step 3. 初始化返回值 & 获取物料信息
	reply := &api.DynLightReply{
		DynamicList: &api.DynamicList{
			HistoryOffset: dynList.HistoryOffset,
			HasMore:       dynList.HasMore,
		},
	}
	if len(dynList.Dynamics) == 0 {
		return reply, nil
	}
	dynCtx, err := s.getMaterial(c, getMaterialOption{general: general, dynamics: dynList.Dynamics})
	if err != nil {
		return nil, err
	}
	// Step 4. 对物料信息处理，获取详情列表
	foldList := s.procListReply(c, dynList.Dynamics, dynCtx, general, _handleTypeLight)
	// Step 5. 回填
	s.procBackfill(c, dynCtx, general, foldList)
	// Step 6. 折叠判断
	reply.DynamicList.List = s.procFold(foldList, dynCtx, general)
	return reply, nil
}
