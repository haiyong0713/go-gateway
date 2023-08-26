package dynamicV2

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"sync"
	"time"

	"go-common/library/log"

	"go-common/library/sync/errgroup.v2"

	api "go-gateway/app/app-svr/app-dynamic/interface/api/v2"
	"go-gateway/app/app-svr/app-dynamic/interface/model"
	dynmdl "go-gateway/app/app-svr/app-dynamic/interface/model/dynamic"
	mdlv2 "go-gateway/app/app-svr/app-dynamic/interface/model/dynamicV2"
	xmetric "go-gateway/app/app-svr/app-dynamic/interface/model/metric"
	relationmdl "go-gateway/app/app-svr/app-dynamic/interface/model/relation"
	feature "go-gateway/app/app-svr/feature/service/sdk"

	accountgrpc "git.bilibili.co/bapis/bapis-go/account/service"
	relationgrpc "git.bilibili.co/bapis/bapis-go/account/service/relation"
	activitygrpc "git.bilibili.co/bapis/bapis-go/activity/service"
	dyncommongrpc "git.bilibili.co/bapis/bapis-go/dynamic/common"
	dyngrpc "git.bilibili.co/bapis/bapis-go/dynamic/service/feed"

	"github.com/pkg/errors"
)

const (
	infoBatchNum = 200
	retryNum     = 1
)

// nolint:gocognit
func (s *Service) DynAll(c context.Context, general *mdlv2.GeneralParam, req *api.DynAllReq) (reply *api.DynAllReply, retErr error) {
	// Step 1. 获取用户关注链信息(关注的up、追番、购买的课程）
	following, pgcFollowing, cheese, ugcSeason, batchListFavorite, err := s.followings(c, general.Mid, true, true, general)
	if err != nil {
		return nil, err
	}
	attentions := mdlv2.GetAttentionsParams(general.Mid, following, pgcFollowing, cheese, ugcSeason, batchListFavorite)
	// Step 2. 根据 refreshType 获取dynamic_list
	var (
		dynList            *mdlv2.DynListRes
		topicSquare        mdlv2.DynAllTopicSquare
		upList             *dyngrpc.MixUpListRsp
		requestId, adExtra string
	)
	if req.GetAdParam() != nil {
		requestId = req.GetAdParam().GetRequestId()
		adExtra = req.GetAdParam().GetAdExtra()
	}
	dynTypeList := []string{"1", "2", "4", "8", "8_1", "8_2", "64", "256", "512", "2048", "2049", "4097", "4098", "4099", "4100", "4101", "4200", "4300", "4301", "4302", "4303", "4305", "4306", "4308", "4310", "4311"}
	if general.IsIPhonePick() && general.GetBuild() >= s.c.BuildLimit.DynMatchIOS || general.IsAndroidPick() && general.GetBuild() >= s.c.BuildLimit.DynMatchAndroid {
		dynTypeList = append(dynTypeList, "4312") // 漫画追漫卡
	}
	if general.IsIPhonePick() && general.GetBuild() >= s.c.BuildLimit.DynCourUpIOS || general.IsAndroidPick() && general.GetBuild() >= s.c.BuildLimit.DynCourUpAndroid {
		dynTypeList = append(dynTypeList, "4313") // UP触发更新课程卡
	}
	if s.isUGCSeasonShareCapble(c, general) {
		dynTypeList = append(dynTypeList, "4314") // UGC合集分享卡
	}
	if s.isDynNewTopicSet(c, general) {
		dynTypeList = append(dynTypeList, "4315") // 话题集订阅更新卡
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
		dynTypeList = []string{"1", "2", "4", "8", "8_1", "8_2", "512", "4097", "4098", "4099", "4100", "4101", "4200", "4308"}
		// nolint:gomnd
		if general.GetBuild() > 1140000 {
			dynTypeList = append(dynTypeList, []string{"4302", "4303"}...)
		}
		if general.GetBuild() >= s.c.BuildLimit.DynCourUpAndroidHD {
			dynTypeList = append(dynTypeList, "4313")
		}
	}
	err = func(ctx context.Context) (err error) {
		switch req.RefreshType {
		case api.Refresh_refresh_new:
			eg := errgroup.WithCancel(ctx)
			// 动态列表
			eg.Go(func(ctx context.Context) (err error) {
				dynList, err = s.dynDao.DynMixNew(ctx, general, req, dynTypeList, attentions)
				if err != nil {
					xmetric.DynamicCoreAPI.Inc("综合页(首刷)", "request_error")
					log.Error("dynamicAll mid(%v) DynMixNew(), error %v", general.Mid, errors.WithStack(err))
				}
				return err
			})
			// 最近访问up主头像列表
			eg.Go(func(ctx context.Context) (err error) {
				upList, err = s.dynDao.MixUpList(ctx, general, req)
				if err != nil {
					xmetric.DynamicCoreAPI.Inc("综合页(最常访问)", "request_error")
					log.Error("dynamicAll mid(%v) MixUpList(%+v), error %v", general.Mid, general, errors.WithStack(err))
				}
				return nil
			})
			if req.From != s.c.Melloi.From {
				eg.Go(func(ctx context.Context) (err error) {
					if s.isDynNewTopicView(ctx, general) {
						// 粉板6.45以上下发新版本话题广场
						topicSquare, err = s.topDao.RcmdNewTopics(ctx, general)
						if err != nil {
							xmetric.DynamicCoreAPI.Inc("话题广场", "request_error")
							log.Error("dynamicAll mid(%v) RcmdNewTopics(%+v), error %v", general.Mid, general, err)
						}
					} else {
						// 老版本维持不变
						topicSquare, err = s.dynDao.MixTopisSquareOld(ctx, general)
						if err != nil {
							xmetric.DynamicCoreAPI.Inc("话题广场", "request_error")
							log.Error("dynamicAll mid(%v) MixTopisSquareOld(%+v), error %v", general.Mid, general, err)
						}
					}
					return nil
				})
			}
			err = eg.Wait()
		case api.Refresh_refresh_history:
			dynList, err = s.dynDao.DynMixHistory(ctx, general, req, dynTypeList, attentions)
			if err != nil {
				xmetric.DynamicCoreAPI.Inc("综合页(翻页)", "request_error")
				log.Error("dynamicAll mid(%v) DynMixHistory(), error %v", general.Mid, err)
			}
		}
		return err
	}(c)
	if err != nil {
		return nil, err
	}
	// Step 3. 初始化返回值 & 获取物料信息
	reply = &api.DynAllReply{
		DynamicList: &api.DynamicList{
			UpdateNum:      dynList.UpdateNum,
			HistoryOffset:  dynList.HistoryOffset,
			UpdateBaseline: dynList.UpdateBaseline,
			HasMore:        dynList.HasMore,
		},
	}

	if len(dynList.Dynamics) == 0 && dynList.RcmdUps == nil && topicSquare == nil && upList == nil {
		return reply, nil
	}
	general.AdFrom = "feed"
	dynCtx, err := s.getMaterial(c, getMaterialOption{
		general: general, dynamics: dynList.Dynamics, rcmdUps: dynList.RcmdUps,
		upRegionRcmds: dynList.RegionUps, mixUpList: upList, playurlParam: req.PlayurlParam,
		fold: dynList.FoldInfo, requestID: requestId, adExtra: adExtra, storyRcmd: dynList.StoryUpCard,
	})
	if err != nil {
		return nil, err
	}
	// Step 4. 对物料信息处理，获取详情列表
	foldList := s.procListReply(c, dynList.Dynamics, dynCtx, general, _handleTypeAll)
	// Step 5. 回填
	s.procBackfill(c, dynCtx, general, foldList)
	// Step 6. 折叠判断
	retDynList := s.procFold(foldList, dynCtx, general)
	// 最常访问
	if upList != nil {
		reply.UpList = s.procMixUpList(c, dynCtx, general, upList)
	}
	// 空关注列表
	if dynList.RcmdUps != nil && dynList.RcmdUps.Type == mdlv2.NoFollow {
		reply.Unfollow = s.procUnfollow(dynCtx, general, dynList.RcmdUps)
	}
	// 低关注列表
	if dynList.RcmdUps != nil {
		// 空动态列表 =1；低关注 =2; 分区聚类推荐up+视频（空列表）=3
		switch dynList.RcmdUps.Type {
		case mdlv2.LowFollow:
			lowfollow, pos := s.procLowfollow(dynCtx, general, dynList.RcmdUps)
			pos++ // 服务端的pos从0开始
			if lowfollow != nil && pos >= 0 {
				if pos > len(retDynList) {
					retDynList = append(retDynList, lowfollow)
				} else {
					retDynList = append(retDynList[:pos], append([]*api.DynamicItem{lowfollow}, retDynList[pos:]...)...)
				}
			}
		case mdlv2.RegionFollow:
			reply.RegionRcmd = s.proUnLoginRcmd(c, dynCtx, dynList.RegionUps, general)
		default:
			log.Warn("dynList.RcmdUps unexpected type(%d)", dynList.RcmdUps.Type)
		}
	}
	// story横插卡
	if story, pos := s.storyRcmdCard(c, dynList.StoryUpCard, dynCtx, general); story != nil {
		if pos > len(retDynList) {
			retDynList = append(retDynList, story)
		} else {
			retDynList = append(retDynList[:pos], append([]*api.DynamicItem{story}, retDynList[pos:]...)...)
		}
	}
	// 话题广场 空关注不出 有分区聚类推荐up+视频不出
	if reply.Unfollow == nil && topicSquare != nil && reply.RegionRcmd == nil {
		reply.TopicList = s.procTopicSquare(c, general, topicSquare)
	}
	reply.DynamicList.List = retDynList
	return reply, nil
}

// nolint:gocognit
func (s *Service) procMixUpList(c context.Context, dynCtx *mdlv2.DynamicContext, general *mdlv2.GeneralParam, upList *dyngrpc.MixUpListRsp) *api.CardVideoUpList {
	var list []*api.UpListItem
	var pos int64
	for _, item := range upList.List {
		if item == nil || item.UserProfile == nil {
			continue
		}
		uid := item.UserProfile.Uid
		// mid > int32老版本抛弃当前卡片
		if s.checkMidMaxInt32(c, int64(uid), general) {
			continue
		}
		userInfo, ok := dynCtx.GetUser(uid)
		if !ok || userInfo == nil {
			continue
		}
		pos++
		itemTmp := &api.UpListItem{
			Uid:       uid,
			Face:      userInfo.Face,
			Name:      userInfo.Name,
			Pos:       pos,
			IsRecall:  item.IsReserveRecall,
			HasUpdate: item.HasUpdate,
		}
		// nolint:gomnd
		switch item.Type {
		case 1: // 直播用户
			itemTmp.UserItemType = api.UserItemType_user_item_type_live
			// 直播埋点
			itemTmp.StyleId = item.StyleId
			// pad端上有bug，不支持该类型
			if !general.IsPad() && !general.IsPadHD() && !general.IsAndroidHD() {
				if item.StyleId != 0 {
					itemTmp.UserItemType = api.UserItemType_user_item_type_live_custom
				}
			}
			itemTmp.Separator = item.HasPostSeparator
			if item.LiveInfo != nil {
				// 直播间跳转
				itemTmp.Uri = item.LiveInfo.JumpUrl
				if item.LiveInfo.JumpUrl == "" {
					itemTmp.Uri = model.FillURI(model.GotoSpaceDyn, strconv.FormatInt(uid, 10), nil)
				}
				// 直播状态
				if item.LiveInfo.LiveStart == 1 {
					itemTmp.LiveState = api.LiveState_live_live
				}
				// 日/夜样式
				itemTmp.DisplayStyleDay = &api.UserItemStyle{
					RectText:       item.DisplayStyleNormal.GetRectText(),
					RectTextColor:  item.DisplayStyleNormal.GetRectTextColor(),
					RectIcon:       item.DisplayStyleNormal.GetRectIcon(),
					RectBgColor:    item.DisplayStyleNormal.GetRectBgColor(),
					OuterAnimation: item.DisplayStyleNormal.GetOuterAnimation(),
				}
				itemTmp.DisplayStyleNight = &api.UserItemStyle{
					RectText:       item.DisplayStyleDark.GetRectText(),
					RectTextColor:  item.DisplayStyleDark.GetRectTextColor(),
					RectIcon:       item.DisplayStyleDark.GetRectIcon(),
					RectBgColor:    item.DisplayStyleDark.GetRectBgColor(),
					OuterAnimation: item.DisplayStyleDark.GetOuterAnimation(),
				}
			}
		case 2: // 动态用户
			itemTmp.UserItemType = api.UserItemType_user_item_type_normal
			itemTmp.Uri = model.FillURI(model.GotoSpaceDyn, strconv.FormatInt(uid, 10), nil)
		case 3: // 首映
			if general.IsIPhonePick() && general.GetBuild() < s.c.BuildLimit.DynPropertyIOS || general.IsAndroidPick() && general.GetBuild() < s.c.BuildLimit.DynPropertyAndroid || general.IsPad() || general.IsPadHD() || general.IsAndroidHD() {
				continue
			}
			itemTmp.UserItemType = api.UserItemType_user_item_type_premiere
			if item.IsReserveRecall {
				itemTmp.UserItemType = api.UserItemType_user_item_type_premiere_reserve
			}
			itemTmp.Uri = model.FillURI(model.GotoAv, strconv.FormatInt(item.Avid, 10), s.inArchivePremiereArg())
			if item.DisplayStyleNormal != nil && item.DisplayStyleDark != nil {
				// 日/夜样式
				itemTmp.DisplayStyleDay = &api.UserItemStyle{
					RectText:       item.DisplayStyleNormal.GetRectText(),
					RectTextColor:  item.DisplayStyleNormal.GetRectTextColor(),
					RectIcon:       item.DisplayStyleNormal.GetRectIcon(),
					RectBgColor:    item.DisplayStyleNormal.GetRectBgColor(),
					OuterAnimation: item.DisplayStyleNormal.GetOuterAnimation(),
				}
				itemTmp.DisplayStyleNight = &api.UserItemStyle{
					RectText:       item.DisplayStyleDark.GetRectText(),
					RectTextColor:  item.DisplayStyleDark.GetRectTextColor(),
					RectIcon:       item.DisplayStyleDark.GetRectIcon(),
					RectBgColor:    item.DisplayStyleDark.GetRectBgColor(),
					OuterAnimation: item.DisplayStyleDark.GetOuterAnimation(),
				}
			}
		default:
			log.Warn("procMixUpList get unknown user type %v", item.Type)
			continue
		}
		list = append(list, itemTmp)
	}
	if len(list) == 0 {
		return nil
	}
	var moreLabel *api.UpListMoreLabel
	if upList.ViewMore != nil {
		// 右侧查看更多按钮 数据
		list = append(list, &api.UpListItem{
			Face:         s.c.Resource.Icon.DynMixUplistMore,
			Name:         upList.ViewMore.Text,
			Uri:          "bilibili://following/up_more_list",
			UserItemType: api.UserItemType_user_item_type_extend,
		})
		// 右上角查看更多label 数据
		moreLabel = &api.UpListMoreLabel{
			Title: upList.ViewMore.Text,
			Uri:   "bilibili://following/up_more_list",
		}
	}
	var titleSwitch int32
	if upList.ModuleTitleSwitch {
		titleSwitch = 1
	}
	res := &api.CardVideoUpList{
		Title:          upList.ModuleTitle,
		List:           list,
		ShowLiveNum:    int32(upList.ShowLiveNum),
		Footprint:      upList.Footprint,
		TitleSwitch:    titleSwitch,
		MoreLabel:      moreLabel,
		ShowMoreLabel:  upList.ViewMore.GetShowMixFixedEntry(),
		ShowInPersonal: upList.ViewMore.GetShowPersonalFixedEntry(),
		ShowMoreButton: upList.ViewMore.GetShow(),
	}
	return res
}

func (s *Service) procTopicSquare(ctx context.Context, general *mdlv2.GeneralParam, obj mdlv2.DynAllTopicSquare) (res *api.TopicList) {
	if obj == nil {
		return nil
	}
	res = obj.ToDynV2TopicList(s.c)
	if res != nil {
		res.ExpStyle = s.dynAllTopicSquareStyle(ctx, general)
	}
	return
}

func (s *Service) procUnfollow(dynCtx *mdlv2.DynamicContext, _ *mdlv2.GeneralParam, unfollow *mdlv2.RcmdUPCard) *api.Unfollow {
	var list []*api.UnfollowUserItem
	var pos int32
	for _, user := range unfollow.Users {
		userInfo, ok := dynCtx.GetUser(user.Uid)
		if !ok {
			continue
		}
		pos++
		item := &api.UnfollowUserItem{
			Face: userInfo.Face,
			Name: userInfo.Name,
			Uid:  userInfo.Mid,
			Pos:  pos,
			Uri:  model.FillURI(model.GotoSpaceDyn, strconv.FormatInt(user.Uid, 10), nil),
			Official: &api.OfficialVerify{
				Type: int32(userInfo.Official.Type),
				Desc: userInfo.Official.Desc,
			},
			Vip: &api.VipInfo{
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
				NicknameColor:   userInfo.Vip.NicknameColor,
			},
			Button: &api.AdditionalButton{
				Type: api.AddButtonType_bt_button,
				Uncheck: &api.AdditionalButtonStyle{
					Text: s.c.Resource.Text.DynMixUnfollowButtonUncheck,
				},
				Check: &api.AdditionalButtonStyle{
					Text: s.c.Resource.Text.DynMixUnfollowButtonCheck,
				},
				Status: api.AdditionalButtonStatus(1),
			},
		}
		if user.Recommend != nil && user.Recommend.Reason != "" {
			item.Sign = user.Recommend.Reason
		}
		if dynCtx.ResStat != nil && dynCtx.ResStat[user.Uid] != nil {
			item.Label = model.StatString(dynCtx.ResStat[user.Uid].Follower, "粉丝")
		}
		if userLive, ok := dynCtx.GetResUserLive(user.Uid); ok && userLive.Status != nil {
			item.LiveState = api.LiveState(userLive.Status.LiveStatus)
		}
		list = append(list, item)
	}
	if len(list) == 0 {
		return nil
	}
	res := &api.Unfollow{
		Title:   s.c.Resource.Text.DynMixUnfollowTitle,
		TrackId: unfollow.TrackId,
		List:    list,
	}
	return res
}

func (s *Service) procLowfollow(dynCtx *mdlv2.DynamicContext, general *mdlv2.GeneralParam, lowfollow *mdlv2.RcmdUPCard) (*api.DynamicItem, int) {
	var list []*api.ModuleBannerUserItem
	for _, user := range lowfollow.Users {
		if user == nil {
			continue
		}
		userInfo, ok := dynCtx.GetUser(user.Uid)
		if !ok {
			continue
		}
		item := &api.ModuleBannerUserItem{
			Face: userInfo.Face,
			Name: userInfo.Name,
			Uid:  userInfo.Mid,
			Uri:  model.FillURI(model.GotoSpaceDyn, strconv.FormatInt(user.Uid, 10), nil),
			Official: &api.OfficialVerify{
				Type: int32(userInfo.Official.Type),
				Desc: userInfo.Official.Desc,
			},
			Vip: &api.VipInfo{
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
				NicknameColor:   userInfo.Vip.NicknameColor,
			},
			Button: &api.AdditionalButton{
				Type: api.AddButtonType_bt_button,
				Uncheck: &api.AdditionalButtonStyle{
					Text: s.c.Resource.Text.DynMixLowfollowButtonUncheck,
				},
				Check: &api.AdditionalButtonStyle{
					Text: s.c.Resource.Text.DynMixLowfollowButtonCheck,
				},
				Status: api.AdditionalButtonStatus(1),
			},
		}
		if user.Recommend != nil {
			item.Label = user.Recommend.Reason
		}
		if userLive, ok := dynCtx.GetResUserLive(user.Uid); ok && userLive.Status != nil {
			item.LiveState = api.LiveState(userLive.Status.LiveStatus)
		}
		list = append(list, item)
	}
	if len(list) == 0 {
		return nil, 0
	}
	res := &api.DynamicItem{
		CardType: api.DynamicType_banner,
		ItemType: api.DynamicType_banner,
		Extend: &api.Extend{
			OrigDynType: api.DynamicType_banner,
		},
		ServerInfo: lowfollow.ServerInfo,
	}
	followTitle := s.c.Resource.Text.DynMixLowfollowTitle
	if general.GetDisableRcmd() {
		followTitle = "热门UP主"
	}
	module := &api.Module{
		ModuleType: api.DynModuleType_module_banner,
		ModuleItem: &api.Module_ModuleBanner{
			ModuleBanner: &api.ModuleBanner{
				Title: followTitle,
				Type:  api.ModuleBannerType_module_banner_type_user,
				Item: &api.ModuleBanner_User{
					User: &api.ModuleBannerUser{
						List: list,
					},
				},
				DislikeText: s.c.Resource.Text.DynMixUnfollowDislike,
				DislikeIcon: s.c.Resource.Icon.DynMixUnfollowDislike,
			},
		},
	}
	res.Modules = append(res.Modules, module)
	return res, int(lowfollow.Pos)
}

func (s *Service) DynFakeCard(c context.Context, general *mdlv2.GeneralParam, req *api.DynFakeCardReq) (*api.DynFakeCardReply, error) {
	// 格式化入参
	dynamics, err := s.formFakeReq(general, req)
	if err != nil {
		log.Error("DynFakeCard form req invalid error %v", err)
		return nil, err
	}
	// 获取物料详情
	dynCtx, err := s.getMaterial(c, getMaterialOption{general: general, dynamics: dynamics})
	if err != nil {
		return nil, err
	}
	// 聚合卡片
	var res = new(api.DynFakeCardReply)
	foldList := s.procListReply(c, dynamics, dynCtx, general, _handleTypeFake)
	if len(foldList.List) > 0 && foldList.List[0].Item != nil {
		// 回填
		s.procBackfill(c, dynCtx, general, foldList)
		res.Item = foldList.List[0].Item
	}
	return res, nil
}

// nolint:gocognit
func (s *Service) formFakeReq(general *mdlv2.GeneralParam, req *api.DynFakeCardReq) ([]*mdlv2.Dynamic, error) {
	var (
		dynamics     []*mdlv2.Dynamic
		fakeDycnmaic mdlv2.FakeDynamicContent
	)
	// 常规部分
	err := json.Unmarshal([]byte(req.Content), &fakeDycnmaic)
	if err != nil {
		log.Error("DynFakeCard mid(%v) req invalid error %v", general.Mid, err)
		return nil, err
	}
	dynamicID, _ := strconv.ParseInt(fakeDycnmaic.DynamicID, 10, 64)
	dynType, _ := strconv.ParseInt(fakeDycnmaic.Type, 10, 64)
	voteID, _ := strconv.ParseInt(fakeDycnmaic.VoteID, 10, 64)
	duration, _ := strconv.ParseInt(fakeDycnmaic.Duration, 10, 64)
	attachAvID, _ := strconv.ParseInt(fakeDycnmaic.AttachAvID, 10, 64)
	dynamic := &mdlv2.Dynamic{
		UID:         general.Mid,
		DynamicID:   dynamicID,
		Type:        dynType,
		Timestamp:   time.Now().Unix(),
		FakeContent: fakeDycnmaic.Content,
		FakeCover:   fakeDycnmaic.CoverURL,
		FakeImages:  fakeDycnmaic.Images,
		Extend: &mdlv2.Extend{
			Vote: &mdlv2.Vote{
				VoteID: voteID,
			},
			Ctrl: fakeDycnmaic.Ctrls,
			FlagCfg: &mdlv2.FlagCfg{
				AvID: attachAvID,
			},
		},
		Duration: duration,
	}
	dynamics = append(dynamics, dynamic)
	if (general.IsAndroidPick() && general.GetBuild() >= s.c.BuildLimit.FakeExtendAndroid) || (general.IsIPhonePick() && general.GetBuild() >= s.c.BuildLimit.FakeExtendIOS) || (general.IsPadHD() && general.GetBuild() >= s.c.BuildLimit.FakeExtendPad) || general.IsPad() || general.IsAndroidHD() {
		if fakeDycnmaic.Extend != "" {
			var fakeDycnmaicExtend *mdlv2.FakeExtend
			if err := json.Unmarshal([]byte(fakeDycnmaic.Extend), &fakeDycnmaicExtend); err != nil {
				log.Warn("DynFakeCard mid(%v) req.extend invalid error %v", general.Mid, err)
				return dynamics, nil
			}
			// 抽奖
			if fakeDycnmaicExtend.Lottery != nil {
				dynamic.Extend.Lott = new(mdlv2.Lott)
				dynamic.Extend.Lott.FromLott(fakeDycnmaicExtend.Lottery)
			}
			// 投票
			if fakeDycnmaicExtend.Vote != nil {
				dynamic.Extend.Vote = new(mdlv2.Vote)
				dynamic.Extend.Vote.FromVote(fakeDycnmaicExtend.Vote)
				// 新附加卡
				dynamic.AttachCardInfosFake = append(dynamic.AttachCardInfosFake, &mdlv2.AttachCardInfo{AttachCard: &dyncommongrpc.AttachCardInfo{Rid: fakeDycnmaicExtend.Vote.VoteId, CardType: dyncommongrpc.AttachCardType_ATTACH_CARD_VOTE}})
			}
			// 商品
			if fakeDycnmaicExtend.Goods != nil && fakeDycnmaicExtend.Goods.ItemsId != "" && fakeDycnmaicExtend.Goods.LinkItemId != "" {
				dynamic.Extend.OpenGoods = &mdlv2.OpenGoods{
					ItemsId:    fakeDycnmaicExtend.Goods.ItemsId,
					ShopId:     fakeDycnmaicExtend.Goods.ShopId,
					Type:       fakeDycnmaicExtend.Goods.Type,
					LinkItemId: fakeDycnmaicExtend.Goods.LinkItemId,
					Version:    fakeDycnmaicExtend.Goods.Version,
				}
				// 新附加卡
				dynamic.AttachCardInfosFake = append(dynamic.AttachCardInfosFake, &mdlv2.AttachCardInfo{AttachCard: &dyncommongrpc.AttachCardInfo{Rid: fakeDycnmaicExtend.Goods.ShopId, CardType: dyncommongrpc.AttachCardType_ATTACH_CARD_DECORATION}})
			}
			// LBS
			if fakeDycnmaicExtend.LBS != nil {
				if fakeDycnmaicExtend.LBS != nil && fakeDycnmaicExtend.LBS.Location != nil && fakeDycnmaicExtend.LBS.ShowTitle != "" && fakeDycnmaicExtend.LBS.Poi != "" {
					dynamic.Extend.Lbs = new(mdlv2.Lbs)
					dynamic.Extend.Lbs.FromLbs(fakeDycnmaicExtend.LBS)
					// 新附加卡
					dynamic.Tags = append(dynamic.Tags, &dyncommongrpc.Tag{TagType: dyncommongrpc.TagType_TAG_LBS})
				}
			}
			// 附加大卡flg_cfg
			if v := fakeDycnmaicExtend.FlagCfg; v != nil {
				dynamic.Extend.FlagCfg = new(mdlv2.FlagCfg)
				dynamic.Extend.FlagCfg.FromFlagCfg(fakeDycnmaicExtend.FlagCfg)
				// 新附加卡
				if v.GetManga() != nil {
					dynamic.AttachCardInfosFake = append(dynamic.AttachCardInfosFake, &mdlv2.AttachCardInfo{AttachCard: &dyncommongrpc.AttachCardInfo{Rid: v.GetManga().GetMangaId(), CardType: dyncommongrpc.AttachCardType_ATTACH_CARD_MANGA}})
				}
				if v.GetPugv() != nil {
					dynamic.AttachCardInfosFake = append(dynamic.AttachCardInfosFake, &mdlv2.AttachCardInfo{AttachCard: &dyncommongrpc.AttachCardInfo{Rid: v.GetPugv().GetPugvId(), CardType: dyncommongrpc.AttachCardType_ATTACH_CARD_PUGV}})
				}
				if v.GetMatch() != nil {
					dynamic.AttachCardInfosFake = append(dynamic.AttachCardInfosFake, &mdlv2.AttachCardInfo{AttachCard: &dyncommongrpc.AttachCardInfo{Rid: v.GetMatch().GetMatchId(), CardType: dyncommongrpc.AttachCardType_ATTACH_CARD_MATCH}})
				}
				if v.GetGame() != nil {
					dynamic.AttachCardInfosFake = append(dynamic.AttachCardInfosFake, &mdlv2.AttachCardInfo{AttachCard: &dyncommongrpc.AttachCardInfo{Rid: v.GetGame().GetGameId(), CardType: dyncommongrpc.AttachCardType_ATTACH_CARD_GAME}})
				}
				if v.GetOgv() != nil {
					dynamic.AttachCardInfosFake = append(dynamic.AttachCardInfosFake, &mdlv2.AttachCardInfo{AttachCard: &dyncommongrpc.AttachCardInfo{Rid: v.GetOgv().GetOgvId(), CardType: dyncommongrpc.AttachCardType_ATTACH_CARD_OGV}})
				}
				if v.GetDecoration() != nil {
					dynamic.AttachCardInfosFake = append(dynamic.AttachCardInfosFake, &mdlv2.AttachCardInfo{AttachCard: &dyncommongrpc.AttachCardInfo{Rid: v.GetDecoration().GetDecorationId(), CardType: dyncommongrpc.AttachCardType_ATTACH_CARD_DECORATION}})
				}
				if v.GetOfficialActivity() != nil {
					dynamic.AttachCardInfosFake = append(dynamic.AttachCardInfosFake, &mdlv2.AttachCardInfo{AttachCard: &dyncommongrpc.AttachCardInfo{Rid: v.GetOfficialActivity().GetOfficialActivityId(), CardType: dyncommongrpc.AttachCardType_ATTACH_CARD_OFFICIAL_ACTIVITY}})
				}
				if v.GetUgc() != nil {
					dynamic.AttachCardInfosFake = append(dynamic.AttachCardInfosFake, &mdlv2.AttachCardInfo{AttachCard: &dyncommongrpc.AttachCardInfo{Rid: v.GetUgc().GetUgcId(), CardType: dyncommongrpc.AttachCardType_ATTACH_CARD_UGC}})
				}
				if v.GetReserve() != nil {
					dynamic.AttachCardInfosFake = append(dynamic.AttachCardInfosFake, &mdlv2.AttachCardInfo{AttachCard: &dyncommongrpc.AttachCardInfo{Rid: v.GetReserve().GetReserveId(), CardType: dyncommongrpc.AttachCardType_ATTACH_CARD_RESERVE}})
				}
			}
			// 优先级
			var (
				_attachCardIndex = map[dyncommongrpc.AttachCardType]int{
					dyncommongrpc.AttachCardType_ATTACH_CARD_RESERVE:           0,
					dyncommongrpc.AttachCardType_ATTACH_CARD_GOODS:             1,
					dyncommongrpc.AttachCardType_ATTACH_CARD_UGC:               2,
					dyncommongrpc.AttachCardType_ATTACH_CARD_ACTIVITY:          3,
					dyncommongrpc.AttachCardType_ATTACH_CARD_OFFICIAL_ACTIVITY: 4,
					dyncommongrpc.AttachCardType_ATTACH_CARD_TOPIC:             5,
					dyncommongrpc.AttachCardType_ATTACH_CARD_OGV:               6,
					dyncommongrpc.AttachCardType_ATTACH_CARD_GAME:              7,
					dyncommongrpc.AttachCardType_ATTACH_CARD_MATCH:             8,
					dyncommongrpc.AttachCardType_ATTACH_CARD_MANGA:             9,
					dyncommongrpc.AttachCardType_ATTACH_CARD_DECORATION:        10,
					dyncommongrpc.AttachCardType_ATTACH_CARD_PUGV:              11,
					dyncommongrpc.AttachCardType_ATTACH_CARD_VOTE:              12,
				}
				tmps []*mdlv2.AttachCardInfo
			)
			for _, v := range dynamic.AttachCardInfosFake {
				tmp := &mdlv2.AttachCardInfo{}
				*tmp = *v
				tmp.Index = _attachCardIndex[v.AttachCard.CardType]
				tmps = append(tmps, tmp)
			}
			// 从小到大
			sort.Slice(tmps, func(i, j int) bool {
				return tmps[i].Index < tmps[j].Index
			})
			var (
				tmpAttachCard []*dyncommongrpc.AttachCardInfo
				ugcAttachCard *dyncommongrpc.AttachCardInfo
				isVote        bool
			)
		LOOP:
			for _, v := range tmps {
				switch v.AttachCard.CardType {
				// UGC视频卡可以和投票卡同时展示,且投票卡在前
				case dyncommongrpc.AttachCardType_ATTACH_CARD_VOTE, dyncommongrpc.AttachCardType_ATTACH_CARD_UGC:
					isVote = true
					if v.AttachCard.CardType == dyncommongrpc.AttachCardType_ATTACH_CARD_UGC {
						ugcAttachCard = v.AttachCard
						continue
					}
					tmpAttachCard = []*dyncommongrpc.AttachCardInfo{v.AttachCard}
				default:
					if isVote {
						continue
					}
					tmpAttachCard = []*dyncommongrpc.AttachCardInfo{v.AttachCard}
					break LOOP
				}
			}
			// UGC视频卡可以和投票卡同时展示,且投票卡在前
			if isVote && ugcAttachCard != nil {
				tmpAttachCard = append(tmpAttachCard, ugcAttachCard)
			}
			dynamic.AttachCardInfos = tmpAttachCard
			// 附加小卡
			if fakeDycnmaicExtend.BottomBusiness != nil {
				for _, item := range fakeDycnmaicExtend.BottomBusiness.Business {
					if item.Rid != 0 {
						b := &mdlv2.BottomBusiness{}
						b.FromBusiness(item)
						dynamic.Extend.BottomBusiness = append(dynamic.Extend.BottomBusiness, b)
						// 新附加卡
						switch b.Type {
						case mdlv2.BottomBusinessBiliCut:
							dynamic.Tags = append(dynamic.Tags, &dyncommongrpc.Tag{TagType: dyncommongrpc.TagType_TAG_DIVERSION, Rid: b.Rid})
						case mdlv2.BottomBusinessBBQ:
							dynamic.Tags = append(dynamic.Tags, &dyncommongrpc.Tag{TagType: dyncommongrpc.TagType_TAG_BBQ})
						case mdlv2.BottomBusinessAutoPGC:
							dynamic.Tags = append(dynamic.Tags, &dyncommongrpc.Tag{TagType: dyncommongrpc.TagType_TAG_AUTOOGV, Rid: b.Rid})
						}
					}
				}
			}
		}
	}
	return dynamics, nil
}

func (s *Service) DynRcmdUpExchange(c context.Context, general *mdlv2.GeneralParam, req *api.DynRcmdUpExchangeReq) (*api.DynRcmdUpExchangeReply, error) {
	var res = new(api.DynRcmdUpExchangeReply)
	// 获取服务端数据
	rcmdUps, err := s.dynDao.DynRcmdUpExchange(c, general, req)
	if err != nil {
		log.Error("DynRcmdUpExchange mid(%v), error %v", general.Mid, err)
		return nil, err
	}
	var midm = make(map[int64]struct{})
	// 推荐用户
	if rcmdUps != nil {
		for _, item := range rcmdUps.Users {
			if item.Uid != 0 {
				midm[item.Uid] = struct{}{}
			}
		}
	}
	// 聚合获取物料
	var ret = new(mdlv2.DynamicContext)
	eg := errgroup.WithCancel(c)
	if len(midm) > 0 {
		var mids []int64
		for mid := range midm {
			mids = append(mids, mid)
		}
		eg.Go(func(ctx context.Context) error {
			res, err := s.accountDao.Cards3New(ctx, mids)
			if err != nil {
				log.Warn("DynRcmdUpExchange mid(%v) Cards3New mids(%v) error(%v)", general.Mid, mids, err)
			}
			ret.ResUser = res
			return nil
		})
		eg.Go(func(ctx context.Context) error {
			ret.ResRelation = s.accountDao.IsAttention(ctx, mids, general.Mid)
			return nil
		})
		eg.Go(func(ctx context.Context) error {
			res, err := s.relationDao.Stats(ctx, mids)
			if err != nil {
				log.Warn("DynRcmdUpExchange mid(%v) Stats mids(%v) error(%v)", general.Mid, mids, err)
			}
			ret.ResStat = res
			return nil
		})
		// 直播信息
		eg.Go(func(ctx context.Context) error {
			lives, playurls, err := s.liveDao.LiveInfos(ctx, mids, general)
			if err != nil {
				log.Error("DynRcmdUpExchange mid(%v) LiveInfos mids %v, err %v", general.Mid, mids, err)
			}
			ret.ResUserLive = lives
			ret.ResUserLivePlayUrl = playurls
			return nil
		})
	}
	if err := eg.Wait(); err != nil {
		return nil, err
	}
	res.Unfollow = s.procUnfollow(ret, general, rcmdUps)
	return res, nil
}

// DynAllPersonal 视频页-最近访问-个人feed流
// nolint:gocognit
func (s *Service) DynAllPersonal(c context.Context, general *mdlv2.GeneralParam, req *api.DynAllPersonalReq) (*api.DynAllPersonalReply, error) {
	// Step 0. 获取用户关注链信息(关注的up、追番、购买的课程）
	following, pgcFollowing, cheese, ugcSeason, batchListFavorite, err := s.followings(c, general.Mid, true, true, general)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	attentions := mdlv2.GetAttentionsParams(general.Mid, following, pgcFollowing, cheese, ugcSeason, batchListFavorite)
	var (
		dynList *mdlv2.AllPersonal
		reserve []*activitygrpc.UpActReserveRelationInfo
	)
	eg := errgroup.WithCancel(c)
	eg.Go(func(ctx context.Context) (err error) {
		// Step 1. 获取 dynamic_list
		dynTypeList := []string{"1", "2", "4", "8", "8_1", "8_2", "64", "256", "512", "2048", "2049", "4097", "4098", "4099", "4100", "4101", "4200", "4300", "4301", "4302", "4303", "4305", "4306", "4308", "4310", "4311"}
		if general.IsIPhonePick() && general.GetBuild() >= s.c.BuildLimit.DynCourUpIOS || general.IsAndroidPick() && general.GetBuild() >= s.c.BuildLimit.DynCourUpAndroid {
			dynTypeList = append(dynTypeList, "4313")
		}
		if s.isUGCSeasonShareCapble(ctx, general) {
			dynTypeList = append(dynTypeList, "4314")
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
			dynTypeList = []string{"1", "2", "4", "8", "8_1", "8_2", "512", "4097", "4098", "4099", "4100", "4101", "4200", "4308"}
			// nolint:gomnd
			if general.GetBuild() > 1140000 {
				dynTypeList = append(dynTypeList, []string{"4302", "4303"}...)
			}
			if general.GetBuild() >= s.c.BuildLimit.DynCourUpAndroidHD {
				dynTypeList = append(dynTypeList, "4313")
			}
		}
		dynList, err = s.dynDao.DynAllPersonal(ctx, req.HostUid, general.Mid, mdlv2.Int32ToBool(req.IsPreload), req.Offset, general.GetBuildStr(), general.GetPlatform(), general.GetMobiApp(), general.GetBuvid(), general.GetDevice(), general.IP, req.From, attentions, req.Footprint, dynTypeList)
		if err != nil {
			xmetric.DynamicCoreAPI.Inc("综合页(快速消费)", "request_error")
			log.Error("DynAllPersonal mid(%v) DynAllPersonal(), error %v", general.Mid, errors.WithStack(err))
			return err
		}
		return nil
	})
	// 是否展示开关
	if s.c.Resource.ReserveShow {
		eg.Go(func(ctx context.Context) (err error) {
			// Step 1. 获取 UP主预约
			reserve, err = s.activityDao.UpActUserSpaceCard(ctx, req.HostUid, general.Mid)
			if err != nil {
				xmetric.DynamicCoreAPI.Inc("综合页(快速消费) UP主预约", "request_error")
				log.Error("DynAllPersonal mid(%v) UpActUserSpaceCard(), error %v", general.Mid, errors.WithStack(err))
				return nil
			}
			return nil
		})
	}
	if err := eg.Wait(); err != nil {
		return nil, err
	}
	// Step 2. 初始化返回值 & 获取物料信息
	reply := &api.DynAllPersonalReply{
		Offset:     dynList.Offset,
		HasMore:    dynList.HasMore,
		ReadOffset: dynList.ReadOffset,
	}
	if len(dynList.Dynamics) == 0 {
		return reply, nil
	}

	dynCtx, err := s.getMaterial(c, getMaterialOption{
		general: general, dynamics: dynList.Dynamics,
		reserves: reserve, playurlParam: req.PlayurlParam,
		fold: dynList.FoldInfo,
	})
	if err != nil {
		return nil, err
	}
	// Step 3. 对物料信息处理，获取详情列表
	foldList := s.procListReply(c, dynList.Dynamics, dynCtx, general, _handleTypeAllPersonal)
	// Step 4. 回填
	s.procBackfill(c, dynCtx, general, foldList)
	// Step 5. 折叠判断
	retDynList := s.procFold(foldList, dynCtx, general)
	reply.List = append(reply.List, retDynList...)
	if g, ok := dynCtx.Grayscale[s.c.Grayscale.Relation.Key]; ok {
		switch g {
		case 1:
			reply.Relation = relationmdl.RelationChange(req.HostUid, dynCtx.ResRelationUltima)
		}
	}
	// Step 6. UP主预约列表
	reply.AdditionUp = s.UpActReserveRelation(c, reserve, dynCtx, general)
	return reply, nil
}

// DynAllUpdOffset 综合页-最近访问-已读进度更新
func (s *Service) DynAllUpdOffset(c context.Context, general *mdlv2.GeneralParam, req *api.DynAllUpdOffsetReq) (*api.NoReply, error) {
	ret := &api.NoReply{}
	err := s.dynDao.DynAllUpdOffset(c, general, req.HostUid, req.ReadOffset, req.Footprint)
	if err != nil {
		log.Error("DynAllUpdOffset mid(%v) DynAllUpdOffset(), error %v", general.Mid, err)
		return nil, err
	}
	return ret, nil
}

// DynVote 投票操作
func (s *Service) DynVote(c context.Context, general *mdlv2.GeneralParam, req *api.DynVoteReq) (*api.DynVoteReply, error) {
	res := new(api.DynVoteReply)
	// 投票操作 报错记录err日志 但是不抛错
	var err error
	if err, res.Toast = s.dynDao.Vote(c, general, req); err != nil {
		log.Error("DynVote mid(%v) Vote(%+v), error %v", general.Mid, req, err)
		return res, nil
	}
	// 获取详情
	var voteResult *dyncommongrpc.VoteInfo
	if voteResult, err = s.dynDao.VoteResult(c, general, req.VoteId); err != nil {
		log.Error("DynVote mid(%v) VoteResult(%v) err %v", general.Mid, req.VoteId, err)
		return res, nil
	}
	// 固定返回非外露的样式
	voteModel := s.votedResult(voteResult, req)
	res.Item = voteModel.GetVote2()
	return res, nil
}

// 动态综合-查看更多-列表
func (s *Service) DynMixUpListViewMore(c context.Context, arg *api.DynMixUpListViewMoreReq, general *mdlv2.GeneralParam) (*api.DynMixUpListViewMoreReply, error) {
	if (general.IsPadHD() && general.GetBuild() < s.c.BuildLimit.UplistMoreSortTypePad) || (general.IsIPhonePick() && general.GetBuild() < s.c.BuildLimit.UplistMoreSortTypeIOS) || (general.IsAndroidPick() && general.GetBuild() < s.c.BuildLimit.UplistMoreSortTypeAndroid) {
		arg.SortType = mdlv2.UplistMoreSortTypeRcmd
	}
	upListViewMoreReply, err := s.upListViewMore(c, general, arg.SortType)
	if err != nil {
		log.Error("日志告警 DynMixUpListViewMore upListViewMore error:(%+v)", err)
		return upListViewMoreReply, err
	}
	return upListViewMoreReply, nil
}

// upListViewMore 查看更多
func (s *Service) upListViewMore(c context.Context, general *mdlv2.GeneralParam, sortType int32) (*api.DynMixUpListViewMoreReply, error) {
	var (
		uidsMap             map[int64]struct{}
		cardsRes            []*accountgrpc.CardsReply
		upListViewMoreReply = &api.DynMixUpListViewMoreReply{SearchDefaultText: "搜索我的关注"}
		mixUplistRes        []*api.MixUpListItem
		relationUltima      map[int64]*relationgrpc.InterrelationReply
	)
	// 动态获取关注up主信息
	upListViewMoreRsp, err := s.dynDao.UpListViewMore(c, general, sortType)
	if err != nil {
		log.Errorc(c, "upListFollowings UpListViewMore(mid: %+v,sortType:%+v) failed. error(%+v)", general.Mid, sortType, err)
		return upListViewMoreReply, err
	}
	if upListViewMoreRsp != nil && upListViewMoreRsp.SortTypes != nil {
		upListViewMoreReply.ShowMoreSortTypes = upListViewMoreRsp.ShowMoreSortTypes
		upListViewMoreReply.DefaultSortType = upListViewMoreRsp.DefaultSortType
		for _, dynSortType := range upListViewMoreRsp.SortTypes {
			var sortTYpe = &api.SortType{}
			sortTYpe.SortType = dynSortType.SortType
			sortTYpe.SortTypeName = dynSortType.SortTypeName
			upListViewMoreReply.SortTypes = append(upListViewMoreReply.SortTypes, sortTYpe)
		}
	}
	if upListViewMoreRsp == nil || len(upListViewMoreRsp.Items) == 0 {
		return upListViewMoreReply, nil
	}
	uidsMap, mixUplistRes = s.procViewMoreFollowingParams(upListViewMoreRsp)
	uidsNum := len(uidsMap)
	if uidsNum == 0 && mixUplistRes == nil {
		log.Info("upListViewMore mixUplistResMap is empty.")
		return upListViewMoreReply, nil
	}
	// 账号信息
	if uidsNum != 0 {
		var (
			uidSubsMap map[int][]int64
			mids       []int64
		)
		uidSubsMap, mids = s.procUidSubsMap(uidsMap)
		eg := errgroup.WithCancel(c)
		mutex := sync.Mutex{}
		// 账号分批处理，拉取账号信息
		for _, uidSubs := range uidSubsMap {
			tmpUidSubs := uidSubs
			eg.Go(func(ctx context.Context) error {
				if err = s.withRetry(retryNum, func() error {
					cardRes, err := s.accountDao.Cards3(ctx, tmpUidSubs)
					if err != nil {
						log.Errorc(ctx, "upListViewMore Cards3(tmpUidSubs: %+v) failed. error(%+v)", tmpUidSubs, err)
						return err
					}
					mutex.Lock()
					defer mutex.Unlock()
					cardsRes = append(cardsRes, cardRes)
					return nil
				}); err != nil {
					log.Errorc(ctx, "upListViewMore withRetry(tmpUidSubs: %+v) failed. error(%+v)", tmpUidSubs, err)
					return err
				}
				return nil
			})
		}
		if len(mids) > 0 {
			eg.Go(func(ctx context.Context) error {
				relationUltima, err = s.accountDao.Interrelations(ctx, general.Mid, mids)
				if err != nil {
					log.Error("Interrelations mid(%v) mids(%v) error %v", general.Mid, mids, err)
				}
				return nil
			})
		}
		if err := eg.Wait(); err != nil {
			return upListViewMoreReply, err
		}
	}
	mixUplistRes = s.procMixUpListItem(general.Mid, mixUplistRes, cardsRes, relationUltima)
	upListViewMoreReply.Items = mixUplistRes
	return upListViewMoreReply, nil
}

func (s *Service) procMixUpListItem(mid int64, mixUpListRes []*api.MixUpListItem, cardsRes []*accountgrpc.CardsReply, relationUltima map[int64]*relationgrpc.InterrelationReply) []*api.MixUpListItem {
	var cards = make(map[int64]*accountgrpc.Card)
	var mixUpListItems []*api.MixUpListItem
	for _, card := range cardsRes {
		for _, item := range card.Cards {
			if item != nil {
				cards[item.Mid] = item
			}
		}
	}
	for _, res := range mixUpListRes {
		card, ok := cards[res.Uid]
		if !ok {
			continue
		}
		res.Uid = card.Mid
		res.Name = card.Name
		res.Face = card.Face
		official := &api.OfficialVerify{
			Type: int32(card.Official.Type),
			Desc: card.Official.Desc,
		}
		res.Official = official
		vip := &api.VipInfo{
			Type:            card.Vip.Type,
			DueDate:         card.Vip.DueDate,
			ThemeType:       card.Vip.ThemeType,
			Status:          card.Vip.Status,
			AvatarSubscript: card.Vip.AvatarSubscript,
			NicknameColor:   card.Vip.NicknameColor,
			Label: &api.VipLabel{
				Path:       card.Vip.Label.Path,
				Text:       card.Vip.Label.Text,
				LabelTheme: card.Vip.Label.LabelTheme,
			},
		}
		res.Vip = vip
		if s.c.Grayscale != nil && s.c.Grayscale.Relation != nil && s.c.Grayscale.Relation.Switch {
			switch s.c.Grayscale.Relation.GrayCheck(mid, "null") {
			case 1:
				res.Relation = relationmdl.RelationChange(card.Mid, relationUltima)
			}
		}
		mixUpListItems = append(mixUpListItems, res)
	}
	return mixUpListItems
}

func (s *Service) procViewMoreFollowingParams(uplistRsp *dyngrpc.UpListViewMoreRsp) (map[int64]struct{}, []*api.MixUpListItem) {
	var (
		mixUplistResMap []*api.MixUpListItem
	)
	uidsMap := make(map[int64]struct{})
	for _, item := range uplistRsp.Items {
		uidsMap[item.Uid] = struct{}{}
		res := &api.MixUpListItem{}
		if item.LiveInfo != nil {
			liveItem := &api.MixUpListLiveItem{
				RoomId: item.LiveInfo.RoomId,
				Uri:    item.LiveInfo.Link,
			}
			if item.LiveInfo.State == dynmdl.UplistMoreLiving {
				liveItem.Status = true
			}
			res.LiveInfo = liveItem
		}
		res.SpecialAttention = item.SpeacialAttention
		res.ReddotState = item.ReddotState
		res.Uid = item.Uid
		res.PremiereState = item.PremiereState
		if item.PremiereState == 1 {
			res.Uri = model.FillURI(model.GotoAv, strconv.FormatInt(item.Avid, 10), s.inArchivePremiereArg())
		}
		mixUplistResMap = append(mixUplistResMap, res)
	}
	return uidsMap, mixUplistResMap
}

func (s *Service) procUidSubsMap(uidsMap map[int64]struct{}) (map[int][]int64, []int64) {
	var uids []int64
	for uid := range uidsMap {
		uids = append(uids, uid)
	}
	uidsNum := len(uidsMap)
	var uidSubsMap = make(map[int][]int64, infoBatchNum)
	// 账号信息
	if uidsNum != 0 {
		j := 1
		for i := 0; i < uidsNum; i += infoBatchNum {
			if i+infoBatchNum > uidsNum {
				// 不足一批次
				uidSubsMap[j] = uids[i:]
			} else {
				// 满一批次
				uidSubsMap[j] = uids[i : i+infoBatchNum]
			}
			j++
		}
	}
	return uidSubsMap, uids
}

func (s *Service) withRetry(attempts int, f func() error) error {
	if err := f(); err != nil {
		if attempts--; attempts > 0 {
			return s.withRetry(attempts, f)
		}
		return err
	}
	return nil
}

func (s *Service) UpActReserveRelation(c context.Context, reserves []*activitygrpc.UpActReserveRelationInfo, dynCtx *mdlv2.DynamicContext, general *mdlv2.GeneralParam) *api.TopAdditionUP {
	res := &api.TopAdditionUP{}
	var (
		liveReserves     []*activitygrpc.UpActReserveRelationInfo
		arcReserves      []*activitygrpc.UpActReserveRelationInfo
		esportsReserves  []*activitygrpc.UpActReserveRelationInfo
		premiereReserves []*activitygrpc.UpActReserveRelationInfo
	)
	for _, v := range reserves {
		// nolint:exhaustive,nolintlint
		switch v.Type {
		case activitygrpc.UpActReserveRelationType_Archive: // 稿件
			arcReserves = append(arcReserves, v)
		case activitygrpc.UpActReserveRelationType_Live: // 直播
			liveReserves = append(liveReserves, v)
		case activitygrpc.UpActReserveRelationType_ESports: // 赛事
			if _, ok := v.Hide[int64(activitygrpc.UpCreateActReserveFrom_FROMQUICKSHOW)]; ok {
				continue
			}
			esportsReserves = append(esportsReserves, v)
		case activitygrpc.UpActReserveRelationType_Premiere: // 首映
			premiereReserves = append(premiereReserves, v)
		}
	}
	sort.Slice(liveReserves, func(i, j int) bool {
		return liveReserves[i].LivePlanStartTime < liveReserves[j].LivePlanStartTime
	})
	sort.Slice(arcReserves, func(i, j int) bool {
		return arcReserves[i].Sid < arcReserves[j].Sid
	})
	sort.Slice(esportsReserves, func(i, j int) bool {
		return (esportsReserves[i].StartShowTime < esportsReserves[j].StartShowTime) ||
			(esportsReserves[i].StartShowTime == esportsReserves[j].StartShowTime && esportsReserves[i].Sid < esportsReserves[j].Sid)
	})
	sort.Slice(premiereReserves, func(i, j int) bool {
		return premiereReserves[i].Sid < premiereReserves[j].Sid
	})
	// 赛事>直播>首映>稿件
	premiereReserves = append(premiereReserves, arcReserves...)
	liveReserves = append(liveReserves, premiereReserves...)
	esportsReserves = append(esportsReserves, liveReserves...)
	for _, v := range esportsReserves {
		switch upActState(v.State) {
		case _upStart, _upOnline:
			tmpDynCtx := &mdlv2.DynamicContext{}
			*tmpDynCtx = *dynCtx
			tmpDynCtx.From = _handleTypeReservePersonal
			common, _, _ := s.additionalUPInfo(c, v, nil, tmpDynCtx, general)
			if common == nil {
				continue
			}
			res.Up = append(res.Up, common)
		}
	}
	// 折叠数量上限
	res.HasFold = s.c.Resource.ReserveHasFold
	return res
}

func (s *Service) storyRcmdCard(c context.Context, storyRcmd *dyngrpc.StoryUPCard, dynCtx *mdlv2.DynamicContext, general *mdlv2.GeneralParam) (*api.DynamicItem, int) {
	// 老版本不下发
	if feature.GetBuildLimit(c, s.c.Feature.FeatureBuildLimit.DynStoryCard, &feature.OriginResutl{
		BuildLimit: (general.IsIPhonePick() && general.GetBuild() < s.c.BuildLimit.DynStoryCardIOS) ||
			(general.IsAndroidPick() && general.GetBuild() <= s.c.BuildLimit.DynStoryCardAndroid)}) ||
		general.IsPadHD() || general.IsAndroidHD() || general.IsPad() {
		return nil, 0
	}
	if storyRcmd == nil {
		return nil, 0
	}
	storyCard := &api.ModuleStory{
		Title:               "动态视频",
		ShowPublishEntrance: storyRcmd.ShowPublishEntrance,
		Uri:                 s.c.Resource.Others.StoryURI,
	}
	if storyRcmd.Title != "" {
		storyCard.Title = storyRcmd.Title
	}
	if pinfo := storyRcmd.PublishInfo; pinfo != nil {
		storyCard.PublishText = pinfo.Title
		storyCard.Cover = pinfo.Cover
		storyCard.Uri = pinfo.Url
	}
	if storyRcmd.FoldInfo != nil {
		storyCard.FoldState = storyRcmd.FoldInfo.FoldState
	}
	// 修复ios 6.51折叠问题
	if (general.IsIPhonePick() && general.GetBuild() >= 65100000) && (general.IsIPhonePick() && general.GetBuild() < 65200000) {
		storyCard.FoldState = 1
	}
	for index, v := range storyRcmd.StoryUps {
		// mid > int32老版本抛弃当前卡片
		if s.checkMidMaxInt32(c, int64(v.Uid), general) {
			continue
		}
		userInfo, ok := dynCtx.GetUser(v.Uid)
		if !ok {
			continue
		}
		ap, ok := dynCtx.GetArchive(v.Rid)
		if !ok {
			continue
		}
		var archive = ap.Arc
		// 付费合集
		if mdlv2.PayAttrVal(archive) {
			continue
		}
		uri := model.FillURI(model.GotoStory, strconv.FormatInt(archive.Aid, 10), model.AvPlayHandlerGRPCV2(ap, archive.FirstCid, true))
		item := &api.StoryItem{
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
			Desc:   userInfo.Name,
			Status: v.Status,
			Type:   api.RcmdType_rcmd_archive,
			RcmdItem: &api.StoryItem_StoryArchive{
				StoryArchive: &api.StoryArchive{
					Cover: archive.Pic,
					Aid:   archive.Aid,
					Uri:   model.FillURI(model.GotoURL, uri, model.SuffixHandler(fmt.Sprintf("scene=dynamic_insert&vmid=%d&uid_pos=%d&next_uid=%d&aid=%d&offset=%d", v.Uid, index, v.Uid, v.Rid, v.DynId))),
					Dimension: &api.Dimension{
						Height: archive.Dimension.Height,
						Width:  archive.Dimension.Width,
						Rotate: archive.Dimension.Rotate,
					},
				},
			},
		}
		storyCard.Items = append(storyCard.Items, item)
	}
	res := &api.DynamicItem{
		CardType: api.DynamicType_story,
		ItemType: api.DynamicType_story,
		Extend: &api.Extend{
			OrigDynType: api.DynamicType_story,
		},
		Modules: []*api.Module{{
			ModuleType: api.DynModuleType_module_story,
			ModuleItem: &api.Module_ModuleStory{
				ModuleStory: storyCard,
			},
		}},
	}
	return res, int(storyRcmd.Pos)
}

func (s *Service) DynSpaceSearchDetails(c context.Context, general *mdlv2.GeneralParam, req *api.DynSpaceSearchDetailsReq) (*api.DynSpaceSearchDetailsReply, error) {
	// Step 1. 获取dynamic_list
	dynIDs := req.DynamicIds
	dynList, err := s.dynDao.DynBriefs(c, dynIDs, general.GetBuildStr(), general.GetPlatform(), general.GetMobiApp(), general.GetBuvid(), general.GetDevice(), general.IP, "", "dt.space-search.0.0.pv", true, true, general.Mid)
	if err != nil {
		log.Error("DynServerDetails mid(%v) DynBriefs(), error %v", general.Mid, err)
		return nil, err
	}
	// Step 2. 初始化返回值 & 获取物料信息
	reply := &api.DynSpaceSearchDetailsReply{}
	if len(dynList.Dynamics) == 0 {
		return reply, nil
	}
	dynCtx, err := s.getMaterial(c, getMaterialOption{general: general, dynamics: dynList.Dynamics, fold: dynList.FoldInfo})
	if err != nil {
		return nil, err
	}
	// Step 3. 对物料信息处理，获取详情列表
	dynCtx.SearchWords = req.SearchWords
	dynCtx.SearchWordRed = true
	foldList := s.procListReply(c, dynList.Dynamics, dynCtx, general, _handleTypeSpaceSearchDetail)
	// Step 4. 回填
	s.procBackfill(c, dynCtx, general, foldList)
	// Step 5. 折叠判断（该接口不需要判断折叠，直接按顺序取列表）
	retDynList := map[int64]*api.DynamicItem{}
	for _, item := range foldList.List {
		dynid, err := strconv.ParseInt(item.Item.Extend.DynIdStr, 10, 64)
		if err != nil {
			continue
		}
		retDynList[dynid] = item.Item
	}
	reply.Items = retDynList
	return reply, nil
}

func (s *Service) DynServerDetails(c context.Context, general *mdlv2.GeneralParam, req *api.DynServerDetailsReq) (*api.DynServerDetailsReply, error) {
	// Step 1. 根据 refreshType 获取dynamic_list
	dynIDs := req.DynamicIds
	dynIDs = append(dynIDs, req.TopDynamicIds...)
	dynList, err := s.dynDao.DynBriefs(c, dynIDs, general.GetBuildStr(), general.GetPlatform(), general.GetMobiApp(), general.GetBuvid(), general.GetDevice(), general.IP, "", "", true, true, general.Mid)
	if err != nil {
		log.Error("DynServerDetails mid(%v) DynBriefs(), error %v", general.Mid, err)
		return nil, err
	}
	// Step 2. 初始化返回值 & 获取物料信息
	reply := &api.DynServerDetailsReply{}
	if len(dynList.Dynamics) == 0 {
		return reply, nil
	}
	// 主态置顶处理
	if req.IsMaster {
		for _, v := range dynList.Dynamics {
			if v.Property == nil {
				v.Property = &dyncommongrpc.Property{}
			}
			v.Property.IsSpaceTop = true
		}
	}
	dynCtx, err := s.getMaterial(c, getMaterialOption{general: general, dynamics: dynList.Dynamics, fold: dynList.FoldInfo})
	if err != nil {
		return nil, err
	}
	// Step 3. 对物料信息处理，获取详情列表
	foldList := s.procListReply(c, dynList.Dynamics, dynCtx, general, _handleTypeServerDetail)
	// Step 4. 回填
	s.procBackfill(c, dynCtx, general, foldList)
	// Step 5. 折叠判断（该接口不需要判断折叠，直接按顺序取列表）
	retDynList := map[int64]*api.DynamicItem{}
	for _, item := range foldList.List {
		dynid, err := strconv.ParseInt(item.Item.Extend.DynIdStr, 10, 64)
		if err != nil {
			continue
		}
		retDynList[dynid] = item.Item
	}
	reply.Items = retDynList
	return reply, nil
}

func (s *Service) UnfollowMatch(c context.Context, general *mdlv2.GeneralParam, req *api.UnfollowMatchReq) error {
	return s.comicDao.DelFavs(c, general.Mid, req.Cid)
}
