package dynamicV2

import (
	"context"

	"go-common/library/ecode"
	api "go-gateway/app/app-svr/app-dynamic/interface/api/v2"
	mdlv2 "go-gateway/app/app-svr/app-dynamic/interface/model/dynamicV2"

	dynamicCommon "git.bilibili.co/bapis/bapis-go/dynamic/common"
	dyngrpc "git.bilibili.co/bapis/bapis-go/dynamic/service/feed"
)

const (
	_filterTabAll      = "all"
	_filterTabVideo    = "video"
	_filterTabContinue = "continueWatching"
	_filterTabUnseen   = "unseen"
	_filterTabLive     = "onLive"
)

const (
	_dynFromFilterContinue = "from_continueWatching"
)

func (s *Service) tabFilters(_ context.Context, _ *mdlv2.GeneralParam, _ *api.DynTabReq, needVideo bool) []*api.DynScreenTab {
	ret := []*api.DynScreenTab{
		{
			Title: "全部",
			Name:  _filterTabAll,
		},
		{
			Title: "未观看",
			Name:  _filterTabUnseen,
		},
		{
			Title: "继续观看",
			Name:  _filterTabContinue,
		},
		{
			Title: "视频",
			Name:  _filterTabVideo,
		},
		{
			Title: "直播",
			Name:  _filterTabLive,
		},
	}
	if !needVideo {
		copy(ret[3:], ret[4:])
		ret = ret[:len(ret)-1]
	}
	return ret
}

func (s *Service) FeedFilter(ctx context.Context, general *mdlv2.GeneralParam, req *api.FeedFilterReq) (resp *api.FeedFilterReply, err error) {
	following, pgcFollowing, cheese, ugcSeason, batchListFavorite, err := s.followings(ctx, general.Mid, true, true, general)
	if err != nil {
		return nil, err
	}
	attentions := mdlv2.GetAttentionsParams(general.Mid, following, pgcFollowing, cheese, ugcSeason, batchListFavorite)

	dynReq := &dyngrpc.FeedFilterReq{
		Uid: general.Mid, Offset: req.Offset, Page: req.Page,
		Meta: general.ToDynMetaDataCtrl(func(md *dynamicCommon.MetaDataCtrl) {
			md.ColdStart = req.ColdStart
			md.FromSpmid = "dt.dt.0.0.pv"
		}),
		AttentionInfo: attentions, InfoCtrl: &dynamicCommon.FeedInfoCtrl{
			NeedLikeUsers: true, NeedLimitFoldStatement: true, NeedBottom: true, NeedTopicInfo: true, NeedLikeIcon: true, NeedRepostNum: true,
		},
	}

	switch req.Tab {
	case _filterTabContinue, _filterTabUnseen, _filterTabLive:
		dynReq.Tab = req.Tab
		// 标记来自继续观看
		if req.Tab == _filterTabContinue {
			general.DynFrom = _dynFromFilterContinue
		}
	default:
		// all 和 video都是用现有其他接口
		return nil, ecode.RequestErr
	}

	dynResp, err := s.dynDao.FeedFilter(ctx, dynReq)
	if err != nil {
		return nil, err
	}

	dynCtx, err := s.getMaterial(ctx, getMaterialOption{general: general, dynamics: dynResp.Dynamics})
	if err != nil {
		return nil, err
	}
	resp = &api.FeedFilterReply{
		HasMore: dynResp.HasMore, Offset: dynResp.HistoryOffset,
	}
	// Step 4. 对物料信息处理，获取详情列表
	foldList := s.procListReply(ctx, dynResp.Dynamics, dynCtx, general, _handleTypeAllFilter)
	// Step 5. 回填
	s.procBackfill(ctx, dynCtx, general, foldList)
	// Step 6. 折叠判断
	retDynList := s.procFold(foldList, dynCtx, general)
	resp.List = retDynList
	return
}
