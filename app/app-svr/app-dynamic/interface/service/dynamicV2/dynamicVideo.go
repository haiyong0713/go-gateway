package dynamicV2

import (
	"context"
	"strconv"
	"strings"

	dyngrpc "git.bilibili.co/bapis/bapis-go/dynamic/service/feed"

	"go-common/library/ecode"
	"go-common/library/log"

	"go-common/library/sync/errgroup.v2"

	activitygrpc "git.bilibili.co/bapis/bapis-go/activity/service"
	pgcAppGrpc "git.bilibili.co/bapis/bapis-go/pgc/service/card/app"

	api "go-gateway/app/app-svr/app-dynamic/interface/api/v2"
	mdlv2 "go-gateway/app/app-svr/app-dynamic/interface/model/dynamicV2"
	xmetric "go-gateway/app/app-svr/app-dynamic/interface/model/metric"
	relationmdl "go-gateway/app/app-svr/app-dynamic/interface/model/relation"

	"github.com/pkg/errors"
)

// DynVideo 动态视频列表
// nolint:gocognit
func (s *Service) DynVideo(c context.Context, general *mdlv2.GeneralParam, req *api.DynVideoReq) (*api.DynVideoReply, error) {
	// Step 1. 获取用户关注链信息(关注的up、追番、购买的课程）
	following, pgcFollowing, cheese, ugcSeason, batchListFavorite, err := s.followings(c, general.Mid, true, true, general)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	attentions := mdlv2.GetAttentionsParams(general.Mid, following, pgcFollowing, cheese, ugcSeason, batchListFavorite)
	var (
		dynList    *mdlv2.DynListRes
		followList *pgcAppGrpc.FollowReply
		upList     *dyngrpc.VideoUpListRsp
	)
	dynTypeList := []string{"8", "512", "4097", "4098", "4099", "4100", "4101", "4303", "4310"}
	if general.IsIPhonePick() && general.GetBuild() >= s.c.BuildLimit.DynCourUpIOS || general.IsAndroidPick() && general.GetBuild() >= s.c.BuildLimit.DynCourUpAndroid {
		dynTypeList = append(dynTypeList, "4313")
	}
	switch {
	case general.IsPadHD(), general.IsPad():
		dynTypeList = []string{"8", "512", "4097", "4098", "4099", "4100", "4101"}
		// nolint:gomnd
		if general.GetBuild() > 12510 { // 大于HD 3.7版本
			dynTypeList = append(dynTypeList, "4310")
		}
		if general.GetBuild() >= s.c.BuildLimit.DynCourUpIOSPAD && general.IsPad() || general.GetBuild() >= s.c.BuildLimit.DynCourUpIOSHD && general.IsPadHD() {
			dynTypeList = append(dynTypeList, "4313")
		}
		if general.GetBuild() > 66200100 && general.IsPad() || general.GetBuild() > 33600100 && general.IsPadHD() {
			dynTypeList = append(dynTypeList, []string{"4302", "4303"}...)
		}
	case general.IsAndroidHD():
		dynTypeList = []string{"8", "512", "4097", "4098", "4099", "4100", "4101"}
		if general.GetBuild() >= s.c.BuildLimit.DynCourUpAndroidHD {
			dynTypeList = append(dynTypeList, "4313")
		}
		// nolint:gomnd
		if general.GetBuild() > 1140000 {
			dynTypeList = append(dynTypeList, []string{"4302", "4303"}...)
		}
	}
	// Step 2. 根据 refreshType 获取dynamic_list
	err = func(ctx context.Context) (err error) {
		switch req.RefreshType {
		case api.Refresh_refresh_new:
			eg := errgroup.WithCancel(ctx)
			// 动态列表
			eg.Go(func(ctx context.Context) (err error) {
				dynList, err = s.dynDao.DynVideoList(ctx, general.Mid, req.UpdateBaseline, req.AssistBaseline, dynTypeList, attentions, general.GetBuildStr(), general.GetPlatform(), general.GetMobiApp(),
					general.GetBuvid(), general.GetDevice(), general.IP, req.From)
				if err != nil {
					xmetric.DynamicCoreAPI.Inc("视频页(首刷)", "request_error")
					log.Error("dynamicVideo mid(%v) DynVideoList(), error %v", general.Mid, err)
				}
				return errors.WithStack(err)
			})
			//我的追番
			eg.Go(func(ctx context.Context) (err error) {
				followList, err = s.pgcDao.MyFollows(ctx, general.Mid)
				if err != nil {
					xmetric.DynamicCoreAPI.Inc("我的追番", "request_error")
					log.Error("dynamicVideo mid(%v) MyFollows(%v), error %v", general.Mid, general.Mid, err)
				}
				return nil
			})
			//最近访问up主头像列表
			eg.Go(func(ctx context.Context) error {
				var errTmp error
				upList, errTmp = s.dynDao.VideoUpList(ctx, general, req)
				if errTmp != nil {
					xmetric.DynamicCoreAPI.Inc("最常访问(视频)", "request_error")
					log.Errorc(ctx, "dynamicVideo mid(%v) VideoUpList(%v, %v, %v), error %v", general.Mid, general.Mid, general.GetBuvid(), errors.WithStack(errTmp), errTmp)
				}
				return nil
			})
			err = eg.Wait()
		case api.Refresh_refresh_history:
			dynList, err = s.dynDao.DynVideoHistory(ctx, general.Mid, req.Offset, int64(req.Page), dynTypeList, attentions, general.GetBuildStr(), general.GetPlatform(), general.GetMobiApp(),
				general.GetBuvid(), general.GetDevice(), general.IP, req.From)
			if err != nil {
				xmetric.DynamicCoreAPI.Inc("视频页(翻页)", "request_error")
				log.Error("dynamicVideo mid(%v) DynVideoHistory(), error %v", general.Mid, err)
			}
		}
		return err
	}(c)
	if err != nil {
		return nil, err
	}
	// Step 3. 初始化返回值 & 获取物料信息
	reply := &api.DynVideoReply{
		DynamicList: &api.CardVideoDynList{
			UpdateNum:      dynList.UpdateNum,
			HistoryOffset:  dynList.HistoryOffset,
			UpdateBaseline: dynList.UpdateBaseline,
			HasMore:        dynList.HasMore,
		},
	}
	if len(dynList.Dynamics) == 0 && followList == nil && upList == nil {
		return reply, nil
	}

	dynCtx, err := s.getMaterial(c, getMaterialOption{
		general: general, dynamics: dynList.Dynamics,
		playurlParam: req.PlayurlParam, fold: dynList.FoldInfo,
	})
	if err != nil {
		return nil, err
	}
	// Step 4. 对物料信息处理，获取详情列表
	foldList := s.procListReply(c, dynList.Dynamics, dynCtx, general, _handleTypeVideo)
	// Step 5. 回填
	s.procBackfill(c, dynCtx, general, foldList)
	// Step 6. 折叠判断
	retDynList := s.procFold(foldList, dynCtx, general)
	// 如果有“最近访问”数据，则插入
	if upList != nil {
		reply.VideoUpList = s.procVideoUpList(c, general, upList)
	}
	// 如果有“我的追番”数据，则插入
	if followList != nil {
		reply.VideoFollowList = s.procPGCFollow(c, followList)
	}

	reply.DynamicList.List = retDynList
	return reply, nil
}

// DynDetail 用动态id获取动态详情（折叠展开使用）
func (s *Service) DynDetails(c context.Context, general *mdlv2.GeneralParam, req *api.DynDetailsReq) (*api.DynDetailsReply, error) {
	// Step 1. 根据 refreshType 获取dynamic_list
	dynIDStrs := strings.Split(req.DynamicIds, ",")
	if len(dynIDStrs) == 0 {
		return nil, errors.Wrapf(ecode.RequestErr, "invalid dynamic_ids. ids: %v", req.DynamicIds)
	}
	var dynIds []int64
	for _, item := range dynIDStrs {
		id, err := strconv.ParseInt(item, 10, 64)
		if err != nil {
			return nil, errors.WithStack(err)
		}
		dynIds = append(dynIds, id)
	}
	dynList, err := s.dynDao.DynBriefs(c, dynIds, general.GetBuildStr(), general.GetPlatform(), general.GetMobiApp(), general.GetBuvid(), general.GetDevice(), general.IP, "", "dt.dt-detail.0.0.pv", true, true, general.Mid)
	if err != nil {
		log.Error("dynamicDetails mid(%v) DynBriefs(), error %v", general.Mid, err)
		return nil, err
	}
	// Step 2. 初始化返回值 & 获取物料信息
	reply := &api.DynDetailsReply{}
	if len(dynList.Dynamics) == 0 {
		return reply, nil
	}

	dynCtx, err := s.getMaterial(c, getMaterialOption{
		general: general, dynamics: dynList.Dynamics, playurlParam: req.PlayurlParam, fold: dynList.FoldInfo,
	})
	if err != nil {
		return nil, err
	}
	// Step 3. 对物料信息处理，获取详情列表
	foldList := s.procListReply(c, dynList.Dynamics, dynCtx, general, _handleTypeDetail)
	// Step 4. 回填
	s.procBackfill(c, dynCtx, general, foldList)
	// Step 5. 折叠判断（该接口不需要判断折叠，直接按顺序取列表）
	var retDynList []*api.DynamicItem
	for _, item := range foldList.List {
		retDynList = append(retDynList, item.Item)
	}
	reply.List = append(reply.List, retDynList...)
	return reply, nil
}

// DynVideoPersonal 视频页-最近访问-个人feed流
func (s *Service) DynVideoPersonal(c context.Context, general *mdlv2.GeneralParam, req *api.DynVideoPersonalReq) (*api.DynVideoPersonalReply, error) {
	// Step 0. 获取用户关注链信息(关注的up、追番、购买的课程）
	following, pgcFollowing, cheese, ugcSeason, batchListFavorite, err := s.followings(c, general.Mid, true, true, general)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	attentions := mdlv2.GetAttentionsParams(general.Mid, following, pgcFollowing, cheese, ugcSeason, batchListFavorite)
	var (
		dynList *mdlv2.VideoPersonal
		reserve []*activitygrpc.UpActReserveRelationInfo
	)
	eg := errgroup.WithCancel(c)
	eg.Go(func(ctx context.Context) (err error) {
		// Step 1. 获取 dynamic_list
		dynTypeList := []string{"8", "512", "4097", "4098", "4099", "4100", "4101", "4303", "4310"}
		if general.IsIPhonePick() && general.GetBuild() >= s.c.BuildLimit.DynCourUpIOS || general.IsAndroidPick() && general.GetBuild() >= s.c.BuildLimit.DynCourUpAndroid {
			dynTypeList = append(dynTypeList, "4313")
		}
		switch {
		case general.IsPadHD(), general.IsPad():
			dynTypeList = []string{"8", "512", "4097", "4098", "4099", "4100", "4101"}
			// nolint:gomnd
			if general.GetBuild() > 12510 { // 大于HD 3.7版本
				dynTypeList = append(dynTypeList, "4310")
			}
			if general.GetBuild() >= s.c.BuildLimit.DynCourUpIOSPAD && general.IsPad() || general.GetBuild() >= s.c.BuildLimit.DynCourUpIOSHD && general.IsPadHD() {
				dynTypeList = append(dynTypeList, "4313")
			}
			if general.GetBuild() > 66200100 && general.IsPad() || general.GetBuild() > 33600100 && general.IsPadHD() {
				dynTypeList = append(dynTypeList, []string{"4302", "4303"}...)
			}
		case general.IsAndroidHD():
			dynTypeList = []string{"8", "512", "4097", "4098", "4099", "4100", "4101"}
			if general.GetBuild() >= s.c.BuildLimit.DynCourUpAndroidHD {
				dynTypeList = append(dynTypeList, "4313")
			}
			// nolint:gomnd
			if general.GetBuild() > 1140000 {
				dynTypeList = append(dynTypeList, []string{"4302", "4303"}...)
			}
		}
		dynList, err = s.dynDao.DynVideoPersonal(c, req.HostUid, general.Mid, mdlv2.Int32ToBool(req.IsPreload), req.Offset, general.GetBuildStr(), general.GetPlatform(), general.GetMobiApp(), general.GetBuvid(), general.GetDevice(), general.IP, req.From, req.Footprint, attentions, dynTypeList)
		if err != nil {
			xmetric.DynamicCoreAPI.Inc("视频页(快速消费)", "request_error")
			log.Error("DynVideoPersonal mid(%v) DynVideoPersonal(), error %v", general.Mid, err)
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
				xmetric.DynamicCoreAPI.Inc("视频页(快速消费) UP主预约", "request_error")
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
	reply := &api.DynVideoPersonalReply{
		Offset:     dynList.Offset,
		HasMore:    dynList.HasMore,
		ReadOffset: dynList.ReadOffset,
	}
	if len(dynList.Dynamics) == 0 {
		return reply, nil
	}

	dynCtx, err := s.getMaterial(c, getMaterialOption{
		general: general, dynamics: dynList.Dynamics, reserves: reserve,
		playurlParam: req.PlayurlParam, fold: dynList.FoldInfo,
	})
	if err != nil {
		return nil, err
	}
	// Step 3. 对物料信息处理，获取详情列表
	foldList := s.procListReply(c, dynList.Dynamics, dynCtx, general, _handleTypeVideoPersonal)
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

// DynVideoUpdateOffset 视频页-最近访问-已读进度更新
func (s *Service) DynVideoUpdateOffset(c context.Context, general *mdlv2.GeneralParam, req *api.DynVideoUpdOffsetReq) (*api.NoReply, error) {
	ret := &api.NoReply{}
	err := s.dynDao.DynVideoUpdateOffset(c, general.Mid, req.HostUid, req.ReadOffset, req.Footprint)
	if err != nil {
		log.Error("DynVideoUpdateOffset mid(%v) DynVideoUpdateOffset(), error %v", general.Mid, err)
		return nil, err
	}
	return ret, nil
}

func (s *Service) procVideoUpList(c context.Context, general *mdlv2.GeneralParam, upList *dyngrpc.VideoUpListRsp) *api.CardVideoUpList {
	var list []*api.UpListItem
	for k, item := range upList.List {
		if item == nil || item.UserProfile == nil {
			continue
		}
		userInfo := item.GetUserProfile()
		uid := userInfo.GetUid()
		// mid > int32老版本抛弃当前卡片
		if s.checkMidMaxInt32(c, uid, general) {
			continue
		}
		itemTmp := &api.UpListItem{
			HasUpdate:    item.GetHasUpdate(),
			Face:         userInfo.GetFace(),
			Name:         userInfo.GetUname(),
			Uid:          uid,
			Pos:          int64(k + 1),
			UserItemType: api.UserItemType_user_item_type_normal,
			IsRecall:     item.GetIsReserveRecall(),
		}
		list = append(list, itemTmp)
	}
	if len(list) == 0 {
		return nil
	}
	card := &api.CardVideoUpList{
		Title:       upList.GetModuleTitle(),
		List:        list,
		Footprint:   upList.GetFootprint(),
		TitleSwitch: 1,
	}
	return card
}

func (s *Service) procPGCFollow(_ context.Context, followList *pgcAppGrpc.FollowReply) *api.CardVideoFollowList {
	res := &api.CardVideoFollowList{
		ViewAllLink: followList.SchemaUri,
	}
	if len(followList.Seasons) != 0 {
		for k, season := range followList.Seasons {
			item := &api.FollowListItem{
				SeasonId: int64(season.SeasonId),
				Title:    season.Title,
				Cover:    season.Cover,
				Url:      season.Url,
				Pos:      int64(k + 1),
			}
			var labels []string
			if season.Progress != nil {
				labels = append(labels, season.Progress.IndexShow)
			}
			if season.NewEp != nil {
				newEp := &api.NewEP{
					Id:        season.NewEp.Id,
					IndexShow: season.NewEp.IndexShow,
					Cover:     season.NewEp.Cover,
				}
				labels = append(labels, season.NewEp.IndexShow)
				item.NewEp = newEp
			}
			if len(labels) != 0 {
				item.SubTitle = strings.Join(labels, " / ")
			}
			res.List = append(res.List, item)
		}
	}
	if len(res.List) == 0 {
		return nil
	}
	return res
}
