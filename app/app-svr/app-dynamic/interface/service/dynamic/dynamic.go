package dynamic

import (
	"context"
	"math"

	"go-common/library/ecode"
	"go-common/library/log"
	"go-common/library/net/metadata"

	"go-common/library/sync/errgroup.v2"

	"go-gateway/app/app-svr/app-dynamic/interface/api"
	dynmdl "go-gateway/app/app-svr/app-dynamic/interface/model/dynamic"
	archivegrpc "go-gateway/app/app-svr/archive/service/api"
	feature "go-gateway/app/app-svr/feature/service/sdk"
	"go-gateway/pkg/idsafe/bvid"

	accountgrpc "git.bilibili.co/bapis/bapis-go/account/service"
	articleMdl "git.bilibili.co/bapis/bapis-go/article/model"
	dynSvrFeedGrpc "git.bilibili.co/bapis/bapis-go/dynamic/service/feed"
	pgcAppGrpc "git.bilibili.co/bapis/bapis-go/pgc/service/card/app"
	pgcShareGrpc "git.bilibili.co/bapis/bapis-go/pgc/service/card/share"
)

type HandlerFunc func(context.Context, *dynmdl.DynContext) error

type foldHandler func(context.Context, *dynmdl.FoldList, map[string]*dynmdl.FoldItem) (map[string]*dynmdl.FoldItem, error)

// FeedInfo IM分享/推送卡片获取详情（稿件、专栏）
func (s *Service) FeedInfo(c context.Context, aids, artids []int64, epids []int32, bvids []string, mobiApp, device string, mid int64) (res *dynmdl.Dynamic, err error) {
	var (
		archiveIDs, articleIDs []int64
		eps                    []int32
		archiveIDm             = make(map[int64]struct{})
		articleIDm             = make(map[int64]struct{})
		epm                    = make(map[int32]struct{})
	)
	for _, aid := range aids {
		archiveIDm[aid] = struct{}{}
	}
	for _, bid := range bvids {
		var bvidTmp int64
		if bvidTmp, _ = bvid.BvToAv(bid); bvidTmp == 0 {
			continue
		}
		archiveIDm[bvidTmp] = struct{}{}
	}
	for archiveID := range archiveIDm {
		archiveIDs = append(archiveIDs, archiveID)
	}
	// nolint:gomnd
	if len(archiveIDs) > 50 {
		err = ecode.RequestErr
		log.Error("FeedInfo too many ids %v", archiveIDs)
		return
	}
	for _, id := range artids {
		articleIDm[id] = struct{}{}
	}
	for id := range articleIDm {
		articleIDs = append(articleIDs, id)
	}
	for _, id := range epids {
		epm[id] = struct{}{}
	}
	for id := range epm {
		eps = append(eps, id)
	}
	eg := errgroup.WithCancel(c)
	var arcm map[int64]*archivegrpc.Arc
	if len(archiveIDs) > 0 {
		eg.Go(func(ctx context.Context) error {
			if arcm, err = s.arcDao.Archive(ctx, archiveIDs, mobiApp, device, mid, ""); err != nil {
				log.Error("%v", err)
				return err
			}
			return nil
		})
	}
	var artm map[int64]*articleMdl.Meta
	if len(articleIDs) > 0 {
		eg.Go(func(ctx context.Context) error {
			if artm, err = s.dynDao.ArticleMetasMc(ctx, articleIDs); err != nil {
				log.Error("%v", err)
				return err
			}
			return nil
		})
	}
	var em []*pgcShareGrpc.ShareMessageResBody
	if len(eps) > 0 {
		eg.Go(func(ctx context.Context) error {
			if em, err = s.pgcDao.ShareMessage(ctx, eps); err != nil {
				log.Error("%v", err)
				return err
			}
			return nil
		})
	}
	// if len
	if err = eg.Wait(); err != nil {
		return
	}
	res = &dynmdl.Dynamic{}
	if arcm != nil {
		var arcs []*dynmdl.Archive
		for _, arc := range arcm {
			if arc == nil {
				continue
			}
			a := &dynmdl.Archive{}
			a.FormArc(arc)
			a.BVID, _ = bvid.AvToBv(a.AID)
			arcs = append(arcs, a)
		}
		res.Archive = arcs
	}
	if artm != nil {
		var articles []*dynmdl.Article
		for _, item := range artm {
			if item == nil {
				continue
			}
			a := &dynmdl.Article{}
			a.FromArt(item)
			if a.ImgURLs == nil {
				a.ImgURLs = []string{""}
			}
			articles = append(articles, a)
		}
		res.Article = articles
	}
	for _, item := range em {
		if item == nil {
			continue
		}
		ep := &dynmdl.PGCShare{}
		ep.FromPgcShare(item)
		res.PGC = append(res.PGC, ep)
	}
	return
}

// DynVideo 动态视频页列表
func (s *Service) DynVideo(c context.Context, header *dynmdl.Header, req *dynmdl.DynVideoReq) (*api.DynVideoReqReply, error) {
	var (
		dynList    *dynmdl.DynVideoListRes
		followList *pgcAppGrpc.FollowReply
		upList     *dynmdl.VdUpListRsp
	)
	// Step 1. 根据refreshType 获取dynamic_list
	err := func(ctx context.Context) error {
		var e error
		if req.Refresh == 1 { // 首刷刷新
			eg := errgroup.WithContext(ctx)
			// 动态列表
			eg.Go(func(ctx2 context.Context) error {
				dynList, e = s.dynDao.DynVideoList(ctx2, req.UpdateBaseLine, req.Teenager, req.Mid)
				return e
			})
			//我的追番
			eg.Go(func(ctx2 context.Context) error {
				followList, e = s.dynDao.MyFollows(ctx2, req.Mid)
				if e != nil {
					log.Errorc(c, "s.dynDao.MyFollows(mid:%v) failed. error(%v)", req.Mid, e)
				}
				return nil
			})
			//最近访问up主头像列表
			eg.Go(func(ctx2 context.Context) error {
				upList, e = s.dynDao.VdUpList(ctx2, req.Teenager, req.Mid, header.Buvid)
				if e != nil {
					log.Errorc(c, "s.dynDao.VdUpList(mid:%v) failed. error(%v)", req.Mid, e)
				}
				return nil
			})
			return eg.Wait()
		}
		// nolint:gomnd
		if req.Refresh == 2 { // 向下翻页
			dynList, e = s.dynDao.DynVideoHistory(ctx, req.Offset, req.Teenager, req.Page, req.Mid)
			return e
		}
		return ecode.RequestErr
	}(c)
	if err != nil {
		return nil, err
	}
	reply := &api.DynVideoReqReply{
		UpdateNum:      int32(dynList.UpdateNum),
		UpdateBaseline: dynList.UpdateBaseline,
		HistoryOffset:  dynList.HistoryOffset,
		HasMore:        int32(dynList.HasMore),
	}
	if dynList.HasMore == 0 {
		reply.HasMore = 2
	}
	if len(dynList.Dynamics) == 0 && followList == nil && upList == nil {
		return reply, nil
	}
	// Step 2. 获取物料信息
	materialParams := &dynmdl.MaterialParams{
		Header:    header,
		VideoMate: req.VideoMate,
		Teenager:  req.Teenager,
		Mid:       req.Mid,
		Dynamics:  dynList.Dynamics,
		UpList:    upList,
	}
	dynCtx, err := s.getMaterial(c, materialParams)
	if err != nil || dynCtx == nil {
		log.Errorc(c, "s.getMaterial() failed. error(%+v)", err)
		return nil, err
	}
	dynCtx.Mid = req.Mid
	// Step 3. 对物料信息处理，获取详情列表
	foldList := s.procVideoListReply(c, dynList.Dynamics, dynCtx, header)
	// Step 4. 折叠判断
	replyList := s.foldConveyer(c, foldList, s.foldUnite, s.foldPublish, s.foldLimit)
	// 如果有“我的追番”数据，则插入
	if followList != nil {
		follow, ok := s.procFollowList(c, followList)
		if ok {
			replyList = append([]*api.DynamicItem{follow}, replyList...)
		}
	}
	// 如果有“最近访问”数据，则插入
	if upList != nil {
		list, ok := s.procUpList(c, upList, dynCtx.ResUid, header)
		if ok {
			replyList = append([]*api.DynamicItem{list}, replyList...)
		}
	}
	reply.List = replyList
	return reply, nil
}

// getVideoListReply 视频列表，处理物料，得到全量的详情列表
func (s *Service) procVideoListReply(c context.Context, dynamics []*dynmdl.Dynamics, dynCtx *dynmdl.DynContext, header *dynmdl.Header) *dynmdl.FoldList {
	// Step 1. 循环处理DynItem容器
	var (
		list     []*api.DynamicItem
		foldList = &dynmdl.FoldList{}
	)
	dynCtx.Emoji = make(map[string]struct{})
	for _, dyn := range dynamics {
		if s.checkMidMaxInt32(c, dyn.UID, header) {
			continue
		}
		dynCtx.DynamicItem = &api.DynamicItem{}
		dynCtx.DynInfo = dyn
		dynCtx.Interim = &dynmdl.Interim{}
		err := s.conveyer(c, dynCtx, s.baseInfo, s.author, s.dispute, s.description, s.dynCard, s.likeUser, s.extend, s.state)
		if err != nil {
			log.Errorc(c, "s.conveyer() failed. error(%+v)", err)
			continue
		}
		list = append(list, dynCtx.DynamicItem)
		foldItem := &dynmdl.FoldItem{
			DynamicItem: dynCtx.DynamicItem,
			Rid:         dynCtx.DynInfo.Rid,
			Uid:         dynCtx.DynInfo.UID,
			Acl:         dynCtx.DynInfo.ACL,
			Timestamp:   dynCtx.DynInfo.Timestamp,
			Type:        dynCtx.DynInfo.Type,
		}
		foldList.List = append(foldList.List, foldItem)
	}
	// Step 2. emoji uri获取
	if len(dynCtx.Emoji) == 0 {
		return foldList
	}
	var emoji []string
	for e := range dynCtx.Emoji {
		emoji = append(emoji, e)
	}
	resEmoji, err := s.dynDao.GetEmoji(c, emoji)
	if err != nil {
		log.Errorc(c, "procVideoListReply getEmoji failed. error(%+v)", err)
		return foldList
	}
	dynCtx.ResEmoji = resEmoji
	for _, dyn := range list {
		dynCtx.DynamicItem = dyn
		s.procEmoji(c, dynCtx)
	}
	return foldList
}

// procFollowList 我的追番数据处理，返回一个 api.DynamicItem 便于插入到列表中
func (s *Service) procFollowList(c context.Context, followList *pgcAppGrpc.FollowReply) (*api.DynamicItem, bool) {
	mdlList := &api.ModuleFollowList{ViewAllLink: followList.SchemaUri}
	if followList.Seasons != nil && len(followList.Seasons) != 0 {
		var itemList []*api.FollowListItem
		for _, season := range followList.Seasons {
			if season == nil {
				continue
			}
			item := &api.FollowListItem{
				SeasonId: season.SeasonId,
				Title:    season.Title,
				Cover:    season.Cover,
				Url:      season.Url,
			}
			if season.NewEp != nil {
				newEp := &api.NewEP{
					Id:        season.NewEp.Id,
					IndexShow: season.NewEp.IndexShow,
					Cover:     season.NewEp.Cover,
				}
				item.NewEp = newEp
			}
			itemList = append(itemList, item)
		}
		mdlList.List = itemList
	}
	if len(mdlList.List) == 0 {
		log.Errorc(c, "followList is empty.")
		return nil, false
	}
	module := &api.Module{
		ModuleType: dynmdl.DynMdlFollowType,
		ModuleItem: &api.Module_ModuleFollowList{
			ModuleFollowList: mdlList,
		},
	}
	item := &api.DynamicItem{
		CardType: dynmdl.DynMdlFollowType,
		Modules:  []*api.Module{module},
	}
	return item, true
}

// procUpList 最近访问 - up主列表
func (s *Service) procUpList(c context.Context, upList *dynmdl.VdUpListRsp, userInfo *accountgrpc.CardsReply, header *dynmdl.Header) (*api.DynamicItem, bool) {
	if userInfo == nil || userInfo.Cards == nil || upList.Items == nil || len(upList.Items) == 0 {
		return nil, false
	}
	upl := &api.ModuleDynUpList{
		ModuleTitle: upList.ModuleTitle,
		ShowAll:     upList.ShowAll,
	}
	var itemList []*api.UpListItem
	for _, item := range upList.Items {
		if s.checkMidMaxInt32(c, item.UID, header) {
			continue
		}
		itemTmp := &api.UpListItem{
			HasUpdate: int32(item.HasUpdate),
			Uid:       item.UID,
		}
		res, ok := userInfo.Cards[item.UID]
		if !ok {
			continue
		}
		itemTmp.Face = res.Face
		itemTmp.Name = res.Name
		itemList = append(itemList, itemTmp)
	}
	if len(itemList) == 0 {
		return nil, false
	}
	upl.List = itemList
	module := &api.Module{
		ModuleType: dynmdl.DynMdlUpList,
		ModuleItem: &api.Module_ModuleUpList{
			ModuleUpList: upl,
		},
	}
	item := &api.DynamicItem{
		CardType: dynmdl.DynMdlUpList,
		Modules:  []*api.Module{module},
	}
	return item, true
}

// nolint:unparam,gocognit
func (s *Service) getMaterial(c context.Context, params *dynmdl.MaterialParams) (*dynmdl.DynContext, error) {
	header := params.Header
	videoMate := params.VideoMate
	var (
		epids, batch, season, dynIds []int64
		aidMap                       = make(map[int64]struct{})
		uidMap                       = make(map[int64]struct{})
		myLike                       = make(map[string][]*dynmdl.LikeBusiItem)
		bottom                       = &dynmdl.DynBottomReq{}
	)
	for _, dyn := range params.Dynamics {
		if dyn.IsAv() {
			aidMap[dyn.Rid] = struct{}{}
			uidMap[dyn.UID] = struct{}{}
		}
		if dyn.IsPGC() {
			epids = append(epids, dyn.Rid)
		}
		if dyn.IsCurrSeason() {
			season = append(season, dyn.Rid)
		}
		if dyn.IsCurrBatch() {
			batch = append(batch, dyn.Rid)
		}
		if dyn.Extend.TopicInfo.IsAttachTopic == 1 {
			dynIds = append(dynIds, dyn.DynamicID)
		}
		if dyn.Extend.Bottom != nil {
			bot := &dynmdl.BottomReqItem{
				DynId:  dyn.DynamicID,
				Bottom: dyn.Extend.Bottom,
			}
			bottom.Dynamics = append(bottom.Dynamics, bot)
		}
		if dyn.Display.LikeUsers != nil && len(dyn.Display.LikeUsers) != 0 {
			for _, uid := range dyn.Display.LikeUsers {
				uidMap[uid] = struct{}{}
			}
		}
		busiItem, business, ok := s.likeBusiProc(dyn)
		if ok {
			myLike[business] = append(myLike[business], busiItem)
		}
	}
	if params.UpList != nil && params.UpList.Items != nil {
		for _, item := range params.UpList.Items {
			uidMap[item.UID] = struct{}{}
		}
	}
	result := &dynmdl.DynContext{}
	eg := errgroup.WithContext(c)
	// 稿件详情
	if len(aidMap) != 0 {
		aids := dynmdl.MapToAids(aidMap)
		eg.Go(func(ctx context.Context) error {
			res, err := s.arcDao.ArcsPlayer(ctx, aids, false, "")
			if err != nil {
				log.Errorc(ctx, "getMaterial Arcs(aids:%+v) failed. error(%+v)", aids, err)
			}
			result.ResArcs = res
			return nil
		})
	}
	// PGC详情
	if len(epids) != 0 {
		eg.Go(func(ctx context.Context) error {
			res, err := s.dynDao.EpList(ctx, epids, header.MobiApp, header.Platform, header.Device, metadata.String(c, metadata.RemoteIP), header.Build,
				videoMate.Fnver, videoMate.Fnval)
			if err != nil {
				log.Errorc(ctx, "getMaterial EpList(epids: %+v) failed. error(%+v)", epids, err)
			}
			result.ResPGC = res
			return nil
		})
	}
	// 账号信息 & 装扮卡片
	if len(uidMap) != 0 {
		uids := dynmdl.MapToInt64(uidMap)
		eg.Go(func(ctx context.Context) error {
			res, err := s.dynDao.Cards3(ctx, uids)
			if err != nil {
				log.Errorc(ctx, "getMaterial Cards3(uids: %+v) failed. error(%+v)", uids, err)
			}
			result.ResUid = res
			return nil
		})
		eg.Go(func(ctx context.Context) error {
			res, err := s.dynDao.DecorateCards(ctx, uids)
			if err != nil {
				log.Errorc(ctx, "getMaterial DecorateCards(uids: %+v) failed. error(%+v)", uids, err)
			}
			result.ResDecorate = res
			return nil
		})
	}
	// 话题信息 & 点赞动画
	if len(dynIds) != 0 {
		eg.Go(func(ctx context.Context) error {
			res, err := s.dynDao.TopicInfos(ctx, dynIds, header.MobiApp, header.Platform, header.Device, header.Build)
			if err != nil {
				log.Errorc(ctx, "getMaterial TopicInfos(dynIds: %+v) failed. error(%+v)", dynIds, err)
				return nil
			}
			var (
				topicRes  = make(map[int64]*dynmdl.TopicResItems)
				lkIconQry []*dynmdl.QryIcon
			)
			if res.Items == nil {
				log.Errorc(ctx, "getMaterial Topic res.Items is empty.")
				return nil
			}
			for _, v := range res.Items {
				topicTmp := &dynmdl.TopicResItems{
					DynamicID:   v.DynamicID,
					FromContent: v.FromContent,
				}
				if v.TopicActivity != nil {
					topicTmp.TopicActivity = v.TopicActivity
				}
				topicRes[v.DynamicID] = topicTmp
				lkIconTmp := &dynmdl.QryIcon{
					DynID: v.DynamicID,
				}
				for _, tpc := range v.FromContent {
					lkIconTmp.TopicNames = append(lkIconTmp.TopicNames, tpc.TopicName)
				}
				lkIconQry = append(lkIconQry, lkIconTmp)
			}
			result.ResTopic = topicRes
			lkIconReq := dynmdl.LikeIconReq{
				Uid:   params.Mid,
				Items: lkIconQry,
			}
			lkIconPlt, ok := dynmdl.LkIconMap[params.Header.Platform]
			if ok {
				lkIconReq.Platform = lkIconPlt
			} else {
				lkIconReq.Platform = dynmdl.LkIconPltAll
			}
			lkIconReply, err := s.dynDao.LikeIcon(ctx, lkIconReq)
			if err != nil {
				log.Errorc(ctx, "getMaterial LikeIcon() failed. error(%+v)", err)
				return nil
			}
			lkIconRes := make(map[int64]*dynmdl.LikeIconItems)
			for _, v := range lkIconReply.Items {
				lkIconRes[v.DynamicID] = v
			}
			result.ResLikeIcon = lkIconRes
			return nil
		})
	}
	// 付费更新卡详情
	if len(batch) != 0 {
		eg.Go(func(ctx context.Context) error {
			res, err := s.dynDao.PGCBatch(ctx, batch)
			if err != nil {
				log.Errorc(ctx, "getMaterial PGCBatch(batch: %+v) failed. error(%+v)", batch, err)
			}
			result.ResPGCBatch = res
			return nil
		})
	}
	// 付费批次卡详情
	if len(season) != 0 {
		eg.Go(func(ctx context.Context) error {
			res, err := s.dynDao.PGCSeason(ctx, season)
			if err != nil {
				log.Errorc(ctx, "getMaterial PGCSeason(batch: %+v) failed. error(%+v)", season, err)
			}
			result.ResPGCSeason = res
			return nil
		})
	}
	// 用户与动态的点赞关系
	if len(myLike) != 0 {
		eg.Go(func(ctx context.Context) error {
			stats, err := s.dynDao.MultiStats(c, params.Mid, myLike)
			if err != nil {
				log.Errorc(c, "getMaterial MultiStats() failed. error(%+v)", err)
				return nil
			}
			result.ResThumStats = stats
			return nil
		})
	}
	// 游戏小卡信息
	if len(bottom.Dynamics) != 0 {
		eg.Go(func(ctx context.Context) error {
			res, err := s.dynDao.GetBottom(ctx, bottom)
			if err != nil {
				log.Errorc(ctx, "getMaterial GetBottom(bottom: %+v) failed. error(%+v)", bottom, err)
				return nil
			}
			if res.Items == nil {
				return nil
			}
			var resm = make(map[int64]*dynmdl.BottomItem)
			for _, item := range res.Items {
				tmp := &dynmdl.BottomItem{
					DynamicID:  item.DynamicID,
					BottomInfo: item.BottomInfo,
				}
				resm[item.DynamicID] = tmp
			}
			result.ResBottom = resm
			return nil
		})
	}
	_ = eg.Wait()
	return result, nil
}

// DynDetails 批量动态ID获取详情列表，目前供折叠展开调用
func (s *Service) DynDetails(c context.Context, header *dynmdl.Header, req *dynmdl.DynDetailsReq) (*api.DynDetailsReply, error) {
	var dynList *dynmdl.DynDetailRsp
	dynList, err := s.dynDao.DynBriefs(c, req.Teenager, req.Mid, req.DynIDs)
	if err != nil {
		return nil, err
	}
	reply := &api.DynDetailsReply{}
	if len(dynList.Dynamics) == 0 {
		return reply, nil
	}
	materialParams := &dynmdl.MaterialParams{
		Header:    header,
		VideoMate: req.VideoMate,
		Teenager:  req.Teenager,
		Mid:       req.Mid,
		Dynamics:  dynList.Dynamics,
	}
	dynCtx, err := s.getMaterial(c, materialParams)
	if err != nil || dynCtx == nil {
		log.Errorc(c, "s.getMaterial() failed. error(%+v)", err)
		return nil, err
	}
	dynCtx.Mid = req.Mid
	// Step 3. 对物料信息处理，获取详情列表
	foldList := s.procVideoListReply(c, dynList.Dynamics, dynCtx, header)
	// Step 4. format foldList.
	replyList := s.foldFinish(foldList)
	reply.List = replyList
	return reply, nil
}

// conveyer DynamicItem 处理过程，执行完一批 handlerFunc 后得到一个 DynamicItem
func (s *Service) conveyer(c context.Context, dynCtx *dynmdl.DynContext, f ...HandlerFunc) error {
	for _, v := range f {
		err := v(c, dynCtx)
		if err != nil {
			log.Errorc(c, "Conveyer failed. dynamic: %v, error(%+v)", dynCtx.DynInfo.DynamicID, err)
			return err
		}
	}
	return nil
}

// foldConveyer 动态列表 折叠处理过程。
func (s *Service) foldConveyer(c context.Context, list *dynmdl.FoldList, handles ...foldHandler) []*api.DynamicItem {
	var (
		ignore = make(map[string]*dynmdl.FoldItem)
		err    error
	)
	for _, f := range handles {
		ignore, err = f(c, list, ignore)
		if err != nil {
			log.Error("fold proc handle failed. error(%+v)\n", err)
		}
	}
	rsp := s.foldFinish(list)
	return rsp
}

// likeBusiProc 点赞业务方获取
func (s *Service) likeBusiProc(dyn *dynmdl.Dynamics) (*dynmdl.LikeBusiItem, string, bool) {
	if dyn.Type == dynmdl.DynTypeVideo {
		item := &dynmdl.LikeBusiItem{
			MsgID: dyn.Rid,
		}
		return item, dynmdl.BusTypeVideo, true
	}
	if dyn.Type == dynmdl.DynTypePGCBangumi || dyn.Type == dynmdl.DynTypePGCMovie || dyn.Type == dynmdl.DynTypePGCTv ||
		dyn.Type == dynmdl.DynTypePGCGuoChuang || dyn.Type == dynmdl.DynTypePGCDocumentary || dyn.Type == dynmdl.DynTypeBangumi {
		item := &dynmdl.LikeBusiItem{
			MsgID:  dyn.Rid,
			OrigID: dyn.UID,
		}
		return item, dynmdl.BusTypePGC, true
	}
	if dyn.Type == dynmdl.DynTypeCheeseBatch {
		item := &dynmdl.LikeBusiItem{
			MsgID:  dyn.Rid,
			OrigID: dyn.UID,
		}
		return item, dynmdl.BusTypeCheese, true
	}
	return nil, "", false
}

// DynVideoPersonal 最近访问 - 个人feed流列表
func (s *Service) DynVideoPersonal(c context.Context, header *dynmdl.Header, req *dynmdl.DynVideoPersonalReq) (*api.DynVideoPersonalReply, error) {
	// Step 1. 获取个人feed流列表
	dynList, err := s.dynDao.VideoPersonal(c, req)
	if err != nil {
		return nil, err
	}
	reply := &api.DynVideoPersonalReply{
		Offset:  dynList.Offset,
		HasMore: int32(dynList.HasMore),
	}
	if dynList.HasMore == 0 {
		reply.HasMore = 2
	}
	if len(dynList.Dynamics) == 0 {
		return reply, nil
	}
	// Step 2. 获取物料信息
	materialParams := &dynmdl.MaterialParams{
		Header:    header,
		VideoMate: req.VideoMate,
		Teenager:  req.Teenager,
		Mid:       req.Mid,
		Dynamics:  dynList.Dynamics,
	}
	dynCtx, err := s.getMaterial(c, materialParams)
	if err != nil || dynCtx == nil {
		log.Errorc(c, "s.getMaterial() failed. error(%+v)", err)
		return nil, err
	}
	dynCtx.Mid = req.Mid
	// Step 3. 对物料信息处理，获取详情列表
	foldList := s.procVideoListReply(c, dynList.Dynamics, dynCtx, header)
	// Step 4. 折叠判断
	replyList := s.foldConveyer(c, foldList, s.foldLimit)
	reply.List = replyList
	return reply, nil
}

// DynUpdOffset 最近访问 - 更新已读进度
func (s *Service) DynUpdOffset(c context.Context, req *dynmdl.DynUpdOffsetReq) (*api.NoReply, error) {
	res := &api.NoReply{}
	err := s.dynDao.DynUpdOffset(c, req)
	if err != nil {
		return res, err
	}
	return res, nil
}

// 动态红点接口
func (s *Service) DynRed(c context.Context, header *dynmdl.Header, req *api.DynRedReq, mid int64) (res *api.DynRedReply, err error) {
	res = &api.DynRedReply{
		RedType:    dynmdl.TabNoPoint,
		DefaultTab: dynmdl.AnchorAll,
	}
	if s.c.Ctrl.RedClose {
		return
	}
	var updateNum *dynSvrFeedGrpc.UpdateNumResp
	if updateNum, err = s.dynDao.UpdateNum(c, mid, req, header.MobiApp, header); err != nil {
		log.Error("%v", err)
		return
	}
	// 默认tab部分
	if updateNum.GetDefaultTab() == "视频" {
		res.DefaultTab = dynmdl.AnchorVideo
	}
	// 红点部分
	res.RedType = updateNum.GetRedType()
	if ok := feature.GetBuildLimit(c, s.c.Feature.FeatureBuildLimit.DynRedLive, &feature.OriginResutl{
		MobiApp:    header.MobiApp,
		Device:     header.Device,
		Build:      int64(header.Build),
		BuildLimit: (header.IsAndroid() && int64(header.Build) >= s.c.BuildLimit.DynRedLiveAndroid)}); ok || header.IsPhone() {
		if updateNum != nil {
			if style := updateNum.SpecialStyle; style != nil && !s.checkMidMaxInt32(c, style.User.Uid, header) {
				res.RedStyle = &api.DynRedStyle{
					DisplayTime: style.DisplayTime,
					CornerMark:  style.CornerMark,
					BgType:      api.BgType_bg_type_default,
					CornerType:  api.CornerType_corner_type_text,
				}
				if style.User != nil {
					res.RedStyle.Up = &api.DynRedStyleUp{
						Uid:  style.User.Uid,
						Face: style.User.Face,
					}
				}
				if style.BgType == dynSvrFeedGrpc.BgType_BG_TYPE_FACE {
					res.RedStyle.BgType = api.BgType_bg_type_face
				}
				if style.CornerType == dynSvrFeedGrpc.CornerType_CORNER_TYPE_ANIMATION {
					res.RedStyle.CornerType = api.CornerType_corner_type_animation
				}
				// 头像样式类型
				switch style.Type {
				case dynSvrFeedGrpc.StyleType_STYLE_TYPE_LIVE:
					res.RedStyle.Type = api.StyleType_STYLE_TYPE_LIVE
				case dynSvrFeedGrpc.StyleType_STYLE_TYPE_DYN_UP:
					res.RedStyle.Type = api.StyleType_STYLE_TYPE_DYN_UP
				default:
					res.RedStyle.Type = api.StyleType_STYLE_TYPE_NONE
				}
			}
		}
	}
	if res.RedType == dynmdl.TabRedTypeCount {
		if updateNum.GetUpdateNum() > 0 {
			res.DynRedItem = &api.DynRedItem{
				Count: updateNum.GetUpdateNum(),
			}
			return
		}
		res.RedType = dynmdl.TabNoPoint
	}
	return
}

func (s *Service) checkMidMaxInt32(c context.Context, mid int64, header *dynmdl.Header) bool {
	if feature.GetBuildLimit(c, s.c.Feature.FeatureBuildLimit.DynMidInt32, &feature.OriginResutl{
		BuildLimit: (header.IsPhone() && int64(header.Build) < s.c.BuildLimit.DynMidInt32IOS) ||
			(header.IsPad() && int64(header.Build) < s.c.BuildLimit.DynMidInt32IOS) ||
			(header.IsAndroid() && int64(header.Build) < s.c.BuildLimit.DynMidInt32Android) ||
			(header.IsAndroidHD() && int64(header.Build) < s.c.BuildLimit.DynMidInt32AndroidHD) ||
			(header.IsPadHD() && int64(header.Build) < s.c.BuildLimit.DynMidInt32IOSHD)}) {
		if mid > math.MaxInt32 {
			return true
		}
	}
	return false
}
