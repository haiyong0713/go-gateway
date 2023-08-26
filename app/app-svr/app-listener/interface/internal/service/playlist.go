package service

import (
	"context"
	"strconv"

	"go-common/library/ecode"
	"go-common/library/log"
	v1 "go-gateway/app/app-svr/app-listener/interface/api/v1"
	"go-gateway/app/app-svr/app-listener/interface/internal/dao"
	"go-gateway/app/app-svr/app-listener/interface/internal/model"

	"github.com/golang/protobuf/ptypes/empty"
	"github.com/pkg/errors"
)

func calculateResult(list []*v1.PlayItem, start, end int) ([]*v1.PlayItem, bool, bool) {
	var reachStart, reachEnd bool

	if start <= 0 {
		end = end - start
		start = 0
		reachStart = true
	}
	if end >= len(list) {
		end = len(list)
		reachEnd = true
	}
	if start >= len(list) {
		return []*v1.PlayItem{}, false, true
	}
	if end <= 0 {
		return []*v1.PlayItem{}, true, false
	}
	return list[start:end], reachStart, reachEnd
}

// page, reachStart?, reachEnd?
func playlistPager(list []*v1.PlayItem, opt *v1.PageOption, anchor *v1.PlayItem) ([]*v1.PlayItem, bool, bool) {
	start := 0
	end := len(list)
	// 锚点首页
	if anchor != nil {
		found := false
		for i, item := range list {
			if item.ItemType == anchor.ItemType && item.Oid == anchor.Oid {
				found = true
				halfBefore := int(opt.PageSize / 2)
				halfBehind := int(opt.PageSize) - halfBefore - 1
				start = i - halfBefore
				end = i + 1 + halfBehind
				break
			}
		}
		if !found {
			return nil, false, false
		}
		return calculateResult(list, start, end)
	}
	// 分页
	if opt.LastItem == nil {
		return calculateResult(list, 0, int(opt.PageSize))
	}
	found := false
	for i, item := range list {
		if item.ItemType == opt.LastItem.ItemType && item.Oid == opt.LastItem.Oid {
			found = true
			switch opt.Direction {
			case v1.PageOption_SCROLL_DOWN:
				start = i + 1
				end = start + int(opt.PageSize)
			case v1.PageOption_SCROLL_UP:
				end = i
				start = end - int(opt.PageSize)
			default:
				log.Error("playlistPager: unknown page Direction(%v)", opt.Direction)
				return nil, false, false
			}
			break
		}
	}
	if !found {
		return nil, false, false
	}

	return calculateResult(list, start, end)
}

//nolint:gocognit
func (s *Service) Playlist(ctx context.Context, req *v1.PlaylistReq) (reply *v1.PlaylistResp, err error) {
	// 校验
	if req.PageOpt == nil {
		req.PageOpt = defaultPager()
		if req.Anchor != nil {
			if err = validatePlayItem(ctx, req.Anchor, 0); err != nil {
				return
			}
		}
	} else {
		if req.PageOpt.PageSize > MaxPageSize || req.PageOpt.PageSize <= 0 {
			return nil, errors.WithMessagef(ecode.RequestErr, "illegal req.PageOpt.PageSize")
		}
		if req.PageOpt.LastItem != nil {
			if err := validatePlayItem(ctx, req.PageOpt.LastItem, 0); err != nil {
				return nil, err
			}
			// 如果在设置了翻页的情况下设置了初始页锚点，那么视作锚点设置无效
			req.Anchor = nil
		}
	}

	dev, _, auth := DevNetAuthFromCtx(ctx)
	reply = new(v1.PlaylistResp)
	var itemList []*v1.PlayItem
	var from v1.PlaylistSource
	var eventTrackingOpts []v1.EventTrackingOpt
	idStr := strconv.FormatInt(req.Id, 10)
	extraIdStr := strconv.FormatInt(req.ExtraId, 10)
	bkArcOpt := &v1.BKArcDetailsReq{PlayerArgs: req.PlayerArgs}
	// 后处理是否跳过
	var isPostProcessSkip func(item *v1.DetailItem) (skip bool)
	// 后处理的单个操作
	var postProcess func(item *v1.DetailItem)

	switch req.From {
	case v1.PlaylistSource_DEFAULT:
		// 服务端默认播单
		list, err := s.dao.Playlist(ctx, auth.Mid, dev.Buvid)
		if err != nil {
			return nil, err
		}
		itemList = list.Items
		setTrack := func(et *v1.EventTracking) {
			et.TrackId = list.TrackID
			et.Batch = list.Batch
		}
		eventTrackingOpts = append(eventTrackingOpts, v1.OpByPlaylistSource(list.From, list.Batch), setTrack)
		// 播单合集续播 收藏夹合集均不出合集信息
		if (list.From == v1.PlaylistSource_MEDIA_LIST && list.Batch == "8") ||
			(list.From == v1.PlaylistSource_USER_FAVOURITE && list.Batch == "21") {
			postProcess = func(item *v1.DetailItem) {
				item.UgcSeasonInfo = nil
			}
		}

	case v1.PlaylistSource_USER_FAVOURITE:
		if req.ExtraId <= 0 || req.Id <= 0 {
			return nil, ecode.RequestErr
		}
		folderMeta := model.FavFolderMeta{Typ: int32(req.ExtraId), Fid: req.Id}
		setOp := v1.OpFavorite
		if folderMeta.Typ != model.FavTypeUgcSeason {
			folderMeta.Fid, folderMeta.Mid = extractFidAndMid(req.Id)
		} else {
			setOp = v1.OpUGCSeason
			postProcess = func(item *v1.DetailItem) {
				// 播单进入场景不需要ugc合集信息
				item.UgcSeasonInfo = nil
			}
		}
		anchorMeta := model.FavItemMeta{}
		if req.Anchor != nil {
			anchorMeta.Otype, anchorMeta.Oid = model.Play2Fav[req.Anchor.ItemType], req.Anchor.Oid
		}
		fDetail, err := s.dao.FavFolderDetail(ctx, dao.FavFolderDetailOpt{
			Mid:    auth.Mid,
			Folder: folderMeta,
			Anchor: anchorMeta,
		})
		if err != nil {
			if err == model.ErrAnchorNotFound {
				return nil, errors.WithMessagef(ecode.RequestErr, "dao.FavFolderDetail: %v, anchor(%+v) page(%+v)", err, req.Anchor, req.PageOpt)
			}
			return nil, err
		}
		for _, dt := range fDetail {
			pa := dt.ToV1PlayItem()
			if pa != nil {
				itemList = append(itemList, pa)
			}
		}
		from = v1.PlaylistSource_USER_FAVOURITE
		setTrack := func(et *v1.EventTracking) {
			et.TrackId = idStr
			et.Batch = extraIdStr
		}
		eventTrackingOpts = append(eventTrackingOpts, setOp, setTrack)
		isPostProcessSkip = func(item *v1.DetailItem) (skip bool) {
			if item.Item.ItemType != model.PlayItemAudio {
				return false
			}
			// 音频失效稿件直接滤掉
			return !item.IsPlayable()
		}

	case v1.PlaylistSource_PICK_CARD:
		if req.ExtraId <= 0 || req.Id <= 0 {
			return nil, ecode.RequestErr
		}
		cardDetail, err := s.dao.CardDetail(ctx, dao.CardDetailsOpt{CardId: req.Id, PickId: req.ExtraId})
		if err != nil {
			return nil, err
		}
		itemList = cardDetail.ToV1PlayItems()
		from = v1.PlaylistSource_PICK_CARD
		setTrack := func(et *v1.EventTracking) {
			et.TrackId = idStr
			et.Batch = extraIdStr
		}
		eventTrackingOpts = append(eventTrackingOpts, v1.OpFinding, setTrack)

	case v1.PlaylistSource_AUDIO_COLLECTION:
		if req.Id <= 0 {
			return nil, ecode.RequestErr
		}
		menuDetail, err := s.dao.MusicMenuDetailV1(ctx, dao.MusicMenuDetailOpt{MenuId: req.Id})
		if err != nil {
			return nil, err
		}
		itemList = menuDetail.ToV1PlayItems()
		from = v1.PlaylistSource_AUDIO_COLLECTION
		setTrack := func(et *v1.EventTracking) {
			et.TrackId = idStr
			et.Batch = extraIdStr
		}
		eventTrackingOpts = append(eventTrackingOpts, v1.OpAudioMenu, setTrack)

	case v1.PlaylistSource_AUDIO_CARD:
		if req.Id <= 0 {
			return nil, ecode.RequestErr
		}
		itemList = []*v1.PlayItem{
			{
				ItemType: model.PlayItemAudio,
				Oid:      req.Id,
			},
		}
		from = v1.PlaylistSource_AUDIO_CARD
		setTrack := func(et *v1.EventTracking) {
			et.TrackId = idStr
			//et.Batch = extraIdStr
		}
		eventTrackingOpts = append(eventTrackingOpts, v1.OpAudioSingle, setTrack)

	case v1.PlaylistSource_MEM_SPACE:
		if req.Id <= 0 {
			return nil, ecode.RequestErr
		}
		resp, err := s.dao.SpaceSongList(ctx, dao.SpaceSongListOpt{Mid: req.Id, WithCollaborator: true})
		if err != nil {
			return nil, err
		}
		for _, s := range resp {
			itemList = append(itemList, s.ToV1PlayItem())
		}
		from = v1.PlaylistSource_MEM_SPACE
		setTrack := func(et *v1.EventTracking) {
			et.TrackId = idStr
			//et.Batch = extraIdStr
		}
		eventTrackingOpts = append(eventTrackingOpts, v1.OpSpaceAudio, setTrack)

	case v1.PlaylistSource_MEDIA_LIST:
		if req.Id <= 0 || req.ExtraId <= 0 {
			return nil, ecode.RequestErr
		}
		mediaListResp, err := s.mediaListResources(ctx, mediaListResOpt{
			Id: req.Id, Extra: req.ExtraId, Auth: auth,
			Anchor: req.Anchor, PageOpt: req.PageOpt,
		})
		if err != nil {
			return nil, err
		}
		itemList = mediaListResp.ItemList
		from = v1.PlaylistSource_MEDIA_LIST
		setTrack := func(et *v1.EventTracking) {
			et.TrackId = idStr
			et.Batch = extraIdStr
		}
		eventTrackingOpts = append(eventTrackingOpts, v1.OpMediaList, setTrack)
		isPostProcessSkip = func(item *v1.DetailItem) (skip bool) {
			if item.Item.ItemType != model.PlayItemAudio {
				return false
			}
			// 音频失效稿件直接滤掉
			return !item.IsPlayable()
		}
		// 如果已经是合集续播则不出相关合集信息
		if req.ExtraId == model.MediaListTypUGCSeason {
			postProcess = func(item *v1.DetailItem) {
				item.UgcSeasonInfo = nil
			}
		}

	case v1.PlaylistSource_UP_ARCHIVE:
		if req.Id <= 0 {
			return nil, ecode.RequestErr
		}
		itemList = []*v1.PlayItem{
			{ItemType: model.PlayItemUGC, Oid: req.Id},
		}
		from = v1.PlaylistSource_UP_ARCHIVE
		setTrack := func(et *v1.EventTracking) {
			et.TrackId = idStr
		}
		eventTrackingOpts = append(eventTrackingOpts, v1.OpMusicPv, setTrack)

	default:
		return nil, errors.WithMessagef(ecode.RequestErr, "unknown req.From(%d)", req.From)
	}

	// 替换默认播单
	// 单曲情况下不替换
	if req.From != v1.PlaylistSource_DEFAULT && req.From != v1.PlaylistSource_AUDIO_CARD && req.From != v1.PlaylistSource_UP_ARCHIVE {
		// 应用排序
		itemList = req.SortOpt.ApplyOrderToV1PlayItems(itemList, &v1.ApplyOrderOpt{Anchor: req.Anchor})
		itemList, err = s.dao.PlaylistReplace(ctx, dao.PlaylistReplaceOpt{
			Mid: auth.Mid, Buvid: dev.Buvid, Items: itemList, From: from, TrackID: req.Id, Batch: req.ExtraId,
		})
		if err != nil {
			return nil, err
		}
	}
	if len(itemList) == 0 {
		reply.ReachStart, reply.ReachEnd = true, true
		reply.List = []*v1.DetailItem{}
		return reply, nil
	}

	pagedList, reachStart, reachEnd := playlistPager(itemList, req.PageOpt, req.Anchor)
	if pagedList == nil {
		// 兼容ios登录后没刷新列表的问题
		if auth.Mid != 0 && req.From == v1.PlaylistSource_DEFAULT && len(dev.Buvid) > 0 {
			// 已登录用户查一下相同buvid有没有记录
			list, err := s.dao.Playlist(ctx, 0, dev.Buvid)
			if err != nil {
				return nil, err
			}
			itemList = list.Items
			setTrack := func(et *v1.EventTracking) {
				et.TrackId = list.TrackID
				et.Batch = list.Batch
			}
			// 重新写入上面的两个tracking option
			eventTrackingOpts = eventTrackingOpts[0 : len(eventTrackingOpts)-2]
			eventTrackingOpts = append(eventTrackingOpts, v1.OpByPlaylistSource(list.From, list.Batch), setTrack)
			pagedList, reachStart, reachEnd = playlistPager(itemList, req.PageOpt, req.Anchor)
			if pagedList != nil {
				// 确认分页正常的话把buvid下面的列表刷到mid下面去
				trackId, _ := strconv.ParseInt(list.TrackID, 10, 64)
				batchId, _ := strconv.ParseInt(list.Batch, 10, 64)
				itemList, err = s.dao.PlaylistReplace(ctx, dao.PlaylistReplaceOpt{
					Mid: auth.Mid, Buvid: dev.Buvid, Items: itemList, From: list.From, TrackID: trackId, Batch: batchId,
				})
				if err != nil {
					log.Warnc(ctx, "buvid playlist fallback: failed to replace mid playlist using a valid buvid playlist: %v", err)
				}
				goto _CONTINUE
			}
		}
		// 否则直接返回错误
		return nil, errors.WithMessagef(ecode.RequestErr, "playlistPager: bad PageOpt: %+v  Anchor: %+v", req.PageOpt, req.Anchor)
	}

_CONTINUE:
	reply.Total = uint32(len(itemList))
	reply.ReachStart = reachStart
	reply.ReachEnd = reachEnd
	if len(pagedList) == 0 {
		reply.List = []*v1.DetailItem{}
		return reply, nil
	}
	bkArcOpt.Items = pagedList
	respList, err := s.BKArcDetails(ctx, bkArcOpt)
	if err != nil {
		return nil, err
	}

	// 回填所有埋点数据/post processing
	reply.List = make([]*v1.DetailItem, 0, len(respList.List))
	var skipItems []*v1.PlayItem
	for i, arc := range respList.List {
		arc.Item.SetEventTracking(eventTrackingOpts...)
		// isPostProcessSkip 用来跳过一些需要后处理时判定的稿件
		if isPostProcessSkip != nil && isPostProcessSkip(arc) {
			skipItems = append(skipItems, arc.Item)
		} else {
			reply.List = append(reply.List, respList.List[i])
		}
		// postProcess 用来直接对一些稿件进行后处理
		if postProcess != nil {
			postProcess(arc)
		}
	}
	// 主动从列表里删掉跳过的数据
	if len(skipItems) > 0 {
		sideErr := s.dao.PlaylistDelete(ctx, dao.PlaylistDeleteOpt{
			Mid: auth.Mid, Buvid: dev.Buvid, Items: skipItems,
		})
		if sideErr != nil {
			log.Warnc(ctx, "playlist postprocess: failed to remove skipped items from list(%v): %v", skipItems, sideErr)
		}
	}

	// 秒开稿件解析
	var targetItem *v1.DetailItem
	if len(reply.List) > 0 {
		targetItem = reply.List[0]
	}
	defer func() {
		s.fillPlayerArgs(ctx, req.PlayerArgs, targetItem)
	}()

	// 锚点稿件或者列表最近播放字段处理
	if req.Anchor != nil && reply.LastPlay == nil {
		// 确认这个东西一定是有效的
		for _, item := range reply.List {
			if item.Item.ItemType == req.Anchor.ItemType && item.Item.Oid == req.Anchor.Oid {
				if len(item.Parts) <= 0 || item.Playable != model.PlayableYES {
					// 无效直接返回
					return
				} else {
					targetItem = item
					targetItem.Item.SubId = req.Anchor.SubId
					break
				}
			}
		}
		lastCid, progress, err := s.dao.PlayHisoryByItemID(ctx, auth.Mid, dev.Buvid, req.Anchor)
		if err != nil {
			if err != model.ErrNoHistoryRecord {
				log.Errorc(ctx, "failed to get playHistory for anchor %+v: %v Discarded", req.Anchor, err)
			}
		} else {
			reply.LastPlay = &v1.PlayItem{ItemType: req.Anchor.ItemType, Oid: req.Anchor.Oid, SubId: []int64{lastCid}}
			reply.LastProgress = progress
			if targetItem != nil && targetItem.Item != nil && len(targetItem.Item.SubId) <= 0 {
				targetItem.Item.SubId = []int64{lastCid}
			}
		}
	}

	return
}

func (s *Service) PlaylistAdd(ctx context.Context, req *v1.PlaylistAddReq) (e *empty.Empty, err error) {
	if len(req.Items) == 0 {
		return nil, errors.WithMessagef(ecode.RequestErr, "empty req.Items")
	}
	e = new(empty.Empty)
	for _, item := range req.Items {
		err = validatePlayItem(ctx, item, 0)
		if err != nil {
			return
		}
	}
	dev, _, auth := DevNetAuthFromCtx(ctx)
	opt := dao.PlaylistAddOpt{
		Mid: auth.Mid, Buvid: dev.Buvid,
		Items: req.Items,
	}
	if req.Pos != nil {
		switch pos := req.Pos.(type) {
		case *v1.PlaylistAddReq_After:
			opt.AfterItem = pos.After
		case *v1.PlaylistAddReq_Head:
			opt.Head = pos.Head
		case *v1.PlaylistAddReq_Tail:
			opt.Tail = pos.Tail
		default:
			return nil, ecode.RequestErr
		}
	} else {
		opt.Tail = true
	}

	err = s.dao.PlaylistAdd(ctx, opt)
	return
}

func (s *Service) PlaylistDel(ctx context.Context, req *v1.PlaylistDelReq) (e *empty.Empty, err error) {
	if len(req.Items) == 0 {
		return nil, errors.WithMessagef(ecode.RequestErr, "empty req.Items")
	}
	e = new(empty.Empty)
	dev, _, auth := DevNetAuthFromCtx(ctx)
	if req.Truncate {
		err = s.dao.PlaylistTruncate(ctx, auth.Mid, dev.Buvid)
		return
	}

	for _, item := range req.Items {
		err = validatePlayItem(ctx, item, 0)
		if err != nil {
			return
		}
	}

	err = s.dao.PlaylistDelete(ctx, dao.PlaylistDeleteOpt{
		Mid:   auth.Mid,
		Buvid: dev.Buvid,
		Items: req.Items,
	})
	return
}
