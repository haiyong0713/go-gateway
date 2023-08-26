package dynamicV2

import (
	"context"
	"fmt"
	"strconv"

	"go-common/library/log"
	api "go-gateway/app/app-svr/app-dynamic/interface/api/v2"
	"go-gateway/app/app-svr/app-dynamic/interface/model"
	mdlv2 "go-gateway/app/app-svr/app-dynamic/interface/model/dynamicV2"
	xmetric "go-gateway/app/app-svr/app-dynamic/interface/model/metric"
	relationmdl "go-gateway/app/app-svr/app-dynamic/interface/model/relation"
	arcgrpc "go-gateway/app/app-svr/archive/service/api"
	feature "go-gateway/app/app-svr/feature/service/sdk"

	dyngrpc "git.bilibili.co/bapis/bapis-go/dynamic/service/feed"
)

const (
	// abtest
	_abUnloginFeed = "dt_unlogin_region_ups"
	_abMiss        = "miss"
	_abUnloginHit  = "feed"
)

func (s *Service) DynUnLoginRcmd(c context.Context, general *mdlv2.GeneralParam, req *api.DynRcmdReq) (*api.DynRcmdReply, error) {
	const (
		_dynFeed = "feed"
		_dynRcmd = "rcmd"
	)
	var (
		regionRcmd *dyngrpc.UnLoginRsp
	)
	dynList := &mdlv2.DynListRes{}
	// 命中实验则暂时feed流动态，否则展示分区推荐UP主
	abtest := _dynRcmd
	// 6.29才走动态未登录feed流实验
	if feature.GetBuildLimit(c, s.c.Feature.FeatureBuildLimit.DynUnLogin, &feature.OriginResutl{
		BuildLimit: (general.IsIPhonePick() && general.GetBuild() >= s.c.BuildLimit.DynUnLoginIOS) ||
			(general.IsAndroidPick() && general.GetBuild() > s.c.BuildLimit.DynUnLoginAndroid)}) {
		if s.unloginAbtest(c, general, req.FakeUid, _abUnloginFeed, _abUnloginHit, model.UnloginAbtestFlag) {
			abtest = _dynFeed
		}
	}
	err := func(ctx context.Context) (err error) {
		if abtest == _dynFeed {
			if dynList, err = s.dynDao.UnLoginFeed(ctx, req); err != nil {
				xmetric.DynamicCoreAPI.Inc("未登录页推荐", "request_error")
				log.Error("DynUnLoginRcmd fake_uid(%d) buvid(%s) UnLoginFeed error(%v)", req.FakeUid, general.Device.Buvid, err)
				return err
			}
		} else {
			regionRcmd, err = s.dynDao.UnLogin(ctx, general)
			if err != nil {
				xmetric.DynamicCoreAPI.Inc("未登录页分区推荐UP主", "request_error")
				log.Error("DynUnLoginRcmd fake_uid(%d) buvid(%s) UnLogin error(%v)", req.FakeUid, general.Device.Buvid, err)
				return err
			}
		}
		return nil
	}(c)
	if err != nil {
		return nil, err
	}
	dynCtx, err := s.getMaterial(c, getMaterialOption{general: general, dynamics: dynList.Dynamics, upRegionRcmds: regionRcmd})
	if err != nil {
		return nil, err
	}
	res := &api.DynRcmdReply{}
	if abtest == _dynFeed {
		if dynList != nil {
			res.DynamicList = &api.DynamicList{
				HasMore:       dynList.HasMore,
				HistoryOffset: dynList.HistoryOffset,
			}
			foldList := s.procListReply(c, dynList.Dynamics, dynCtx, general, _handleTypeUnLogin)
			s.procBackfill(c, dynCtx, general, foldList)
			retDynList := s.procFold(foldList, dynCtx, general)
			res.DynamicList.List = retDynList
		}
	} else {
		res.RegionRcmd = s.proUnLoginRcmd(c, dynCtx, regionRcmd, general)
	}
	return res, nil
}

func (s *Service) proUnLoginRcmd(c context.Context, dynCtx *mdlv2.DynamicContext, regionUps *dyngrpc.UnLoginRsp, general *mdlv2.GeneralParam) *api.DynRegionRcmd {
	if regionUps == nil {
		return nil
	}
	res := &api.DynRegionRcmd{
		Opts: &api.RcmdOption{
			ShowTitle: regionUps.Opts.ShowTitle,
		},
	}
	var regionRcmds []*api.DynRegionRcmdItem
	for _, v := range regionUps.RegionUps {
		var itemRcmds []*api.ModuleRcmd
		for _, upRcmd := range v.UpVideos {
			// mid > int32老版本抛弃当前卡片
			if s.checkMidMaxInt32(c, upRcmd.Uid, general) {
				continue
			}
			userInfo, ok := dynCtx.ResUser[upRcmd.Uid]
			if !ok {
				continue
			}
			moduleRcmd := &api.ModuleRcmd{
				Author: &api.RcmdAuthor{
					Author: &api.UserInfo{
						Mid:  userInfo.Mid,
						Name: userInfo.Name,
						Face: userInfo.Face,
						Official: &api.OfficialVerify{ // 认证
							Type: int32(userInfo.Official.Type),
							Desc: userInfo.Official.Desc,
						},
						Vip: &api.VipInfo{ // 会员
							Type:    userInfo.Vip.Type,
							Status:  userInfo.Vip.Status,
							DueDate: userInfo.Vip.DueDate,
							Label: &api.VipLabel{
								Path:       userInfo.Vip.Label.Path,
								Text:       userInfo.Vip.Label.Text,
								LabelTheme: userInfo.Vip.Label.LabelTheme,
							},
							ThemeType:       userInfo.Vip.ThemeType,
							AvatarSubscript: userInfo.Vip.AvatarSubscript,
						},
						Nameplate: &api.Nameplate{ // 勋章
							Nid:        int64(userInfo.Nameplate.Nid),
							Name:       userInfo.Nameplate.Name,
							Image:      userInfo.Nameplate.Image,
							ImageSmall: userInfo.Nameplate.ImageSmall,
							Level:      userInfo.Nameplate.Level,
							Condition:  userInfo.Nameplate.Condition,
						},
						Uri:        model.FillURI(model.GotoSpaceDyn, strconv.FormatInt(userInfo.Mid, 10), nil),
						FaceNft:    userInfo.FaceNft,
						FaceNftNew: userInfo.FaceNftNew,
					},
					Relation: relationmdl.RelationChange(upRcmd.Uid, dynCtx.ResRelationUltima),
				},
			}
			if relation, ok := dynCtx.ResStat[upRcmd.Uid]; ok {
				moduleRcmd.Author.Desc = fmt.Sprintf("粉丝：%s", model.StatString(relation.Follower, ""))
			}
			// 推荐理由
			if upRcmd.RcmdReason != "" {
				moduleRcmd.Author.Desc = moduleRcmd.Author.Desc + " " + upRcmd.RcmdReason
			}
			var items []*api.RcmdItem
			for _, aid := range upRcmd.AvIds {
				ap, ok := dynCtx.GetArchive(aid)
				if !ok {
					continue
				}
				var archive = ap.Arc
				cardArc := &api.RcmdArchive{
					Cover:           archive.Pic,
					CoverLeftIcon_1: api.CoverIcon_cover_icon_play,
					CoverLeftText_1: s.numTransfer(int(archive.Stat.View)),
					Uri:             model.FillURI(model.GotoAv, strconv.FormatInt(archive.Aid, 10), model.AvPlayHandlerGRPCV2(ap, archive.FirstCid, true)),
					Aid:             archive.Aid,
				}
				// 展示稿件标题实验
				if res.Opts.ShowTitle {
					cardArc.Title = archive.Title
				}
				// PGC特殊逻辑
				if archive.AttrVal(arcgrpc.AttrBitIsPGC) == arcgrpc.AttrYes && archive.RedirectURL != "" {
					cardArc.Uri = archive.RedirectURL
					cardArc.IsPgc = true
				}
				item := &api.RcmdItem{
					Type: api.RcmdType_rcmd_archive,
					RcmdItem: &api.RcmdItem_RcmdArchive{
						RcmdArchive: cardArc,
					},
				}
				items = append(items, item)
			}
			// 小于3个直接抛弃
			// nolint:gomnd
			if len(items) < 3 {
				continue
			}
			moduleRcmd.Items = items
			moduleRcmd.ServerInfo = regionUps.ServerInfo
			itemRcmds = append(itemRcmds, moduleRcmd)
		}
		// 小于2个不展示
		// nolint:gomnd
		if len(itemRcmds) < 2 {
			continue
		}
		regionRcmd := &api.DynRegionRcmdItem{
			Rid:   v.Tid,
			Title: v.Name + "区UP主",
			Items: itemRcmds,
		}
		// 如果分区id没有title文案透传或者固定返回
		if v.Tid == 0 {
			regionRcmd.Title = "猜你喜欢的UP主"
			if v.Name != "" {
				regionRcmd.Title = v.Name
			}
			if general.GetDisableRcmd() {
				regionRcmd.Title = "热门UP主"
			}
		}
		regionRcmds = append(regionRcmds, regionRcmd)
	}
	if len(regionRcmds) == 0 {
		return nil
	}
	res.Items = regionRcmds
	return res
}
