package service

import (
	"context"

	"go-common/component/metadata/device"
	"go-common/library/ecode"
	"go-common/library/log"
	v1 "go-gateway/app/app-svr/app-listener/interface/api/v1"
	"go-gateway/app/app-svr/app-listener/interface/internal/dao"
	"go-gateway/app/app-svr/app-listener/interface/internal/model"

	"git.bilibili.co/bapis/bapis-go/bilibili/pagination"
	"github.com/pkg/errors"
	"go-common/library/sync/errgroup.v2"
)

func (s *Service) lastNPlayHistory(ctx context.Context, n int, mid int64, dev *device.Device) []*v1.DetailItem {
	resp, err := s.dao.PlayHistory(ctx, mid, dev)
	if err != nil {
		log.Errorc(ctx, "error get last %d PlayHistory dao error: %v Discarded", n, err)
		return nil
	}
	if len(resp) <= 0 {
		return nil
	}
	if len(resp) >= n {
		resp = resp[0:n]
	}
	reqItem := make([]*v1.PlayItem, 0, len(resp))
	historyMap := make(map[string]model.PlayHistory)
	for i, h := range resp {
		reqItem = append(reqItem, h.ToV1PlayItem())
		historyMap[h.Hash()] = resp[i]
	}
	arcDetails, err := s.BKArcDetails(model.NewBkArchiveArgs(ctx, model.BkArchiveArg{
		EnableServerFilter: true,
	}), &v1.BKArcDetailsReq{Items: reqItem})
	if err != nil {
		log.Errorc(ctx, "error get detail for last %d history: %v Discarded", n, err)
		return nil
	}
	ret := arcDetails.List
	for _, item := range ret {
		if h, ok := historyMap[item.Item.Hash()]; ok {
			item.LastPart = h.LastPlay
			item.Progress = h.Progress
		}
	}
	return ret
}

func (s *Service) RcmdPlaylist(ctx context.Context, req *v1.RcmdPlaylistReq) (resp *v1.RcmdPlaylistResp, err error) {
	resp = new(v1.RcmdPlaylistResp)
	dev, net, auth := DevNetAuthFromCtx(ctx)
	eg := errgroup.WithContext(ctx)
	if req.Annotations == nil {
		req.Annotations = make(map[string]string)
	}

	ro := new(v1.RcmdOffset)
	if s.isRcmdOffsetCapable(ctx) {
		if err = ro.UnmarshalFromBase64(req.GetPage().GetNext()); err != nil {
			return nil, errors.WithMessagef(ecode.RequestErr, "invalid rcmdPageOffset: %v(%s)", err, req.GetPage().GetNext())
		}
		// 首次请求的时候初始化好相关参数
		if len(req.GetPage().GetNext()) <= 0 {
			ro.RcmdFrom = req.GetServerRcmdFromType()
			ro.Id = req.GetId()
			ro.SessionId = req.Annotations["session_id"]
			ro.FromTrackid = req.Annotations["from_trackid"]
		}
	}

	var hisPart []*v1.DetailItem
	if req.NeedHistory {
		eg.Go(func(ctx context.Context) error {
			hisPart = s.lastNPlayHistory(ctx, 2, auth.Mid, dev)
			return nil
		})
	}
	var topCards []model.RcmdTopCard
	if req.NeedTopCards {
		eg.Go(func(ctx context.Context) (err error) {
			topCards, err = s.dao.RcmdTopCards(ctx, dao.RcmdTopCardsOpt{
				Mid: auth.Mid, Dev: dev,
			})
			if err != nil {
				log.Errorc(ctx, "error get the RcmdTopCards for mid(%d) Discarded: %v", auth.Mid, err)
			}
			topCards, err = s.fillTopCards(ctx, auth.Mid, dev.Buvid, topCards)
			if err != nil {
				log.Errorc(ctx, "error fill the RcmdTopCards for mid(%d) Discarded: %v", auth.Mid, err)
			}
			return nil
		})
	}
	var extraIds []int64
	switch req.From {
	case v1.RcmdPlaylistReq_UP_ARCHIVE, v1.RcmdPlaylistReq_ARCHIVE_VIEW:
		if req.Id != 0 {
			extraIds = append(extraIds, req.Id)
		}
	default:
		// do nothing
	}
	var rcmdPart []*v1.DetailItem
	var detailItems []*v1.DetailItem
	eg.Go(func(ctx context.Context) error {
		rcmds, err := s.dao.RecommendArchives(ctx, dao.RecommendArchivesOpt{
			Dev:         dev,
			Net:         net,
			Mid:         auth.Mid,
			ExtraAids:   extraIds,
			OriginAid:   ro.Id,
			RcmdFrom:    ro.RcmdFrom,
			Page:        int64(ro.Page),
			SessionId:   ro.SessionId,
			FromTrackId: ro.FromTrackid,
		})
		if err != nil {
			return err
		}
		setTrackID := func(et *v1.EventTracking) {
			et.TrackId = rcmds.TrackID
		}
		for i := range rcmds.Arcs {
			arc := rcmds.Arcs[i].ToV1DetailItem(model.PlayItemUGC)
			// 首个稿件填充秒开
			if i == 0 {
				detailItems = append(detailItems, arc)
			}
			if i < len(extraIds) {
				arc.Item.SetEventTracking(v1.OpManual)
			} else {
				arc.Item.SetEventTracking(v1.OpRecommend, setTrackID)
			}
			rcmdPart = append(rcmdPart, arc)
		}
		return nil
	})
	err = eg.Wait()
	if err != nil {
		return
	}

	// 处理推荐列表信息以及prepend历史条目
	resp.List = append(resp.List, rcmdPart...)
	fillHistoryPart(resp, hisPart)
	// 填充头部卡片
	for i, topCard := range topCards {
		if c, dt := topCard.ToV1TopCard(auth.Mid, int64(i+1), dev); c != nil {
			resp.TopCards = append(resp.TopCards, c)
			if topCard.PlayImmediately() && dt != nil {
				detailItems = append(detailItems, dt)
			}
		} else {
			log.Warnc(ctx, "unexpected nil topCard result. Discarded: %+v", topCard)
		}
	}
	// 处理秒开
	s.fillPlayerArgs(ctx, req.PlayerArgs, detailItems...)

	// 处理翻页
	if s.isRcmdOffsetCapable(ctx) {
		if len(resp.List) > 0 {
			ro.Page += 1
			resp.NextPage = &pagination.PaginationReply{
				Next: ro.MarshalToBase64(ctx),
			}
		}
	}

	return
}

func (s *Service) fillTopCards(ctx context.Context, mid int64, buvid string, cards []model.RcmdTopCard) (ret []model.RcmdTopCard, err error) {
	items := make([]*v1.PlayItem, 0, len(cards))
	lkMap := make(map[string]int)
	for idx, c := range cards {
		if arc := c.Card.GetArchive(); arc != nil {
			m := &v1.PlayItem{
				ItemType: int32(arc.Type),
				Oid:      arc.Aid,
			}
			items = append(items, m)
			lkMap[m.Hash()] = idx
		}
	}
	if len(items) > 0 {
		var historyRes map[string]dao.PlayHistoryResult
		eg := errgroup.WithContext(ctx)
		eg.Go(func(ctx context.Context) error {
			resp, err := s.BKArcDetails(ctx, &v1.BKArcDetailsReq{Items: items})
			if err != nil {
				return err
			}
			for _, res := range resp.GetList() {
				if idx, ok := lkMap[res.Item.Hash()]; ok {
					card := cards[idx]
					card.Detail = res
					cards[idx] = card
				}
			}
			return nil
		})
		eg.Go(func(ctx context.Context) (err error) {
			historyRes, err = s.dao.PlayHisoryByItemIDs(ctx, mid, buvid, items...)
			if err != nil {
				return nil
			}
			return nil
		})
		if err = eg.Wait(); err != nil {
			return nil, err
		}
		// 填充历史记录
		if historyRes != nil {
			for _, c := range cards {
				if c.Detail == nil {
					continue
				}
				if r, ok := historyRes[c.Detail.Item.Hash()]; ok {
					c.Detail.LastPart = r.SubId
					c.Detail.Progress = r.Progress
				}
			}
		}
	}
	return cards, nil
}

// 在列表头部填充历史内容
func fillHistoryPart(resp *v1.RcmdPlaylistResp, hisPart []*v1.DetailItem) {
	if len(hisPart) <= 0 {
		return
	}
	// 先和列表内已有内容去重
	resp.HistoryLen = 0
	for i, his := range hisPart {
		isDup := false
		for _, r := range resp.List {
			if r.Item.ItemType == his.Item.ItemType && r.Item.Oid == his.Item.Oid {
				isDup = true
				break
			}
		}
		if isDup {
			continue
		}
		// 无重复的内容写进列表头部
		arc := hisPart[i]
		arc.Item.SetEventTracking(v1.OpHistory)
		resp.List = append([]*v1.DetailItem{arc}, resp.List...)
		resp.HistoryLen++
	}
}
