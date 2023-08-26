package dynamicV2

import (
	"context"

	"go-common/library/log"
	api "go-gateway/app/app-svr/app-dynamic/interface/api/v2"
	mdlv2 "go-gateway/app/app-svr/app-dynamic/interface/model/dynamicV2"
	xmetric "go-gateway/app/app-svr/app-dynamic/interface/model/metric"
	feature "go-gateway/app/app-svr/feature/service/sdk"
)

const (
	_dynFromLive  = "live"
	_dynFromSpace = "space"
)

// nolint:gocognit
func (s *Service) DynSpace(c context.Context, general *mdlv2.GeneralParam, req *api.DynSpaceReq) (*api.DynSpaceRsp, error) {
	// Step 1. 获取用户关注链信息(关注的up、追番、购买的课程）
	following, pgcFollowing, cheese, ugcSeason, batchListFavorite, err := s.followings(c, general.Mid, true, true, general)
	if err != nil {
		return nil, err
	}
	attentions := mdlv2.GetAttentionsParams(general.Mid, following, pgcFollowing, cheese, ugcSeason, batchListFavorite)
	// Step 2. 根据 refreshType 获取dynamic_list
	// 空间动态列表
	dynTypeList := []string{"1", "2", "4", "8", "8_1", "8_2", "64", "256", "512", "2048", "2049", "4097", "4098", "4099", "4100", "4200", "4101", "4300", "4301", "4302", "4303", "4305", "4306", "4310", "4311"}
	if general.IsIPhonePick() && general.GetBuild() >= s.c.BuildLimit.DynMatchIOS || general.IsAndroidPick() && general.GetBuild() >= s.c.BuildLimit.DynMatchAndroid {
		dynTypeList = append(dynTypeList, "4312")
	}
	if general.IsIPhonePick() && general.GetBuild() >= s.c.BuildLimit.DynCourUpIOS || general.IsAndroidPick() && general.GetBuild() >= s.c.BuildLimit.DynCourUpAndroid {
		dynTypeList = append(dynTypeList, "4313")
	}
	if s.isUGCSeasonShareCapble(c, general) {
		dynTypeList = append(dynTypeList, "4314")
	}
	if ok := feature.GetBuildLimit(c, s.c.Feature.FeatureBuildLimit.DynSpaceLive, &feature.OriginResutl{
		BuildLimit: (general.IsIPhonePick() && general.GetBuild() >= s.c.BuildLimit.DynSpaceLiveIOS) ||
			(general.IsAndroidPick() && general.GetBuild() > s.c.BuildLimit.DynSpaceLiveAndroid)}); ok && req.From == _dynFromSpace {
		dynTypeList = append(dynTypeList, "4308")
	}
	switch {
	case general.IsPadHD(), general.IsPad():
		dynTypeList = []string{"1", "2", "4", "8", "8_1", "512", "2048", "2049", "4097", "4098", "4099", "4100", "4101", "4310"}
		if feature.GetBuildLimit(c, s.c.Feature.FeatureBuildLimit.DynArticle, &feature.OriginResutl{
			BuildLimit: (general.IsPad() && general.GetBuild() >= s.c.BuildLimit.DynArticleIOSPad) ||
				(general.IsPadHD() && general.GetBuild() >= s.c.BuildLimit.DynArticleIOSHD)}) {
			dynTypeList = append(dynTypeList, "64")
		}
		if general.GetBuild() > 66200100 && general.IsPad() || general.GetBuild() > 33600100 && general.IsPadHD() {
			dynTypeList = append(dynTypeList, []string{"4302", "4303"}...)
		}
		if general.GetBuild() >= s.c.BuildLimit.DynCourUpIOSPAD && general.IsPad() || general.GetBuild() >= s.c.BuildLimit.DynCourUpIOSHD && general.IsPadHD() {
			dynTypeList = append(dynTypeList, "4313")
		}
	case general.IsAndroidHD():
		dynTypeList = []string{"1", "2", "4", "8", "8_1", "8_2", "512", "4097", "4098", "4099", "4100", "4101", "4200"}
		// nolint:gomnd
		if general.GetBuild() > 1140000 {
			dynTypeList = append(dynTypeList, []string{"4302", "4303"}...)
		}
		if general.GetBuild() >= s.c.BuildLimit.DynCourUpAndroidHD {
			dynTypeList = append(dynTypeList, "4313")
		}
	}
	dynList, err := s.dynDao.SpaceHistory(c, general, req, dynTypeList, attentions)
	if err != nil {
		xmetric.DynamicCoreAPI.Inc("动态空间页", "request_error")
		log.Error("dynamicAll mid(%v) SpaceHistory(), error %v", general.Mid, err)
		return nil, err
	}
	// Step 3. 初始化返回值 & 获取物料信息
	reply := &api.DynSpaceRsp{
		HistoryOffset: dynList.HistoryOffset,
		HasMore:       dynList.HasMore,
	}
	if len(dynList.Dynamics) == 0 {
		return reply, nil
	}
	general.AdFrom = _handleTypeSpace
	general.DynFrom = req.From
	if req.From == _dynFromLive {
		general.CloseAutoPlay = true
	}

	dynCtx, err := s.getMaterial(c, getMaterialOption{general: general, dynamics: dynList.Dynamics, fold: dynList.FoldInfo})
	if err != nil {
		return nil, err
	}
	// Step 4. 对物料信息处理，获取详情列表
	foldList := s.procListReply(c, dynList.Dynamics, dynCtx, general, _handleTypeSpace)
	// Step 5. 回填
	s.procBackfill(c, dynCtx, general, foldList)
	// Step 6. 折叠判断
	var retDynList []*api.DynamicItem
	retDynList = s.procFold(foldList, dynCtx, general)
	// 低关注列表
	if dynList.RcmdUps != nil && dynList.RcmdUps.Type == mdlv2.LowFollow {
		lowfollow, pos := s.procLowfollow(dynCtx, general, dynList.RcmdUps)
		pos++ // 服务端的pos从0开始
		if lowfollow != nil && pos >= 0 {
			if pos > len(retDynList) {
				retDynList = append(retDynList, lowfollow)
			} else {
				retDynList = append(retDynList[:pos], append([]*api.DynamicItem{lowfollow}, retDynList[pos:]...)...)
			}
		}
	}
	reply.List = retDynList
	return reply, nil
}
