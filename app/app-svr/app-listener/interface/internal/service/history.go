package service

import (
	"context"

	"go-common/library/ecode"
	v1 "go-gateway/app/app-svr/app-listener/interface/api/v1"
	"go-gateway/app/app-svr/app-listener/interface/internal/dao"
	"go-gateway/app/app-svr/app-listener/interface/internal/model"

	listenerSvc "git.bilibili.co/bapis/bapis-go/dynamic/service/listener"
	"github.com/pkg/errors"

	"github.com/golang/protobuf/ptypes/empty"
)

const (
	DefaultPageSize = 20
	MaxPageSize     = 30
)

func defaultPager() *v1.PageOption {
	return &v1.PageOption{PageSize: DefaultPageSize, Direction: v1.PageOption_SCROLL_DOWN}
}

func (s *Service) PlayHistory(ctx context.Context, req *v1.PlayHistoryReq) (resp *v1.PlayHistoryResp, err error) {
	// 校验
	if req.PageOpt == nil {
		req.PageOpt = defaultPager()
	} else {
		if req.PageOpt.PageSize > MaxPageSize || req.PageOpt.PageSize <= 0 {
			return nil, errors.WithMessagef(ecode.RequestErr, "illegal req.PageOpt.PageSize")
		}
		if req.PageOpt.LastItem != nil {
			if err := validatePlayItem(ctx, req.PageOpt.LastItem, 0); err != nil {
				return nil, err
			}
		}
	}

	// dao
	dev, _, auth := DevNetAuthFromCtx(ctx)
	list, err := s.dao.PlayHistory(ctx, auth.Mid, dev)
	if err != nil {
		return
	}
	if len(list) == 0 {
		return &v1.PlayHistoryResp{ReachEnd: true, List: []*v1.DetailItem{}}, nil
	}

	itemList := make([]*v1.PlayItem, 0, len(list))
	resp = &v1.PlayHistoryResp{}
	for _, item := range list {
		itm := &v1.PlayItem{ItemType: item.ArcType, Oid: item.Oid}
		itemList = append(itemList, itm)
	}

	pagedList, _, reachEnd := playlistPager(itemList, req.PageOpt, nil)
	if pagedList == nil {
		return nil, errors.WithMessagef(ecode.RequestErr, "playhistoryPager: bad PageOpt: %+v", req.PageOpt)
	}
	resp.ReachEnd = reachEnd
	resp.Total = uint32(len(list))

	if len(pagedList) > 0 {
		detailsResp, err := s.BKArcDetails(model.NewBkArchiveArgs(ctx, model.BkArchiveArg{
			EnableServerFilter: true,
		}), &v1.BKArcDetailsReq{
			Items: pagedList,
		})
		if err != nil {
			return nil, err
		}
		resp.List = detailsResp.List
	}

	// 填入历史记录信息
	lkMap := make(map[string]*v1.DetailItem)
	for i, item := range resp.List {
		lkMap[item.Item.Hash()] = resp.List[i]
		item.Item.SetEventTracking(v1.OpHistory)
	}
	for _, item := range list {
		if ritem, ok := lkMap[item.Hash()]; ok {
			ritem.LastPart = item.LastPlay
			ritem.Progress = item.Progress
			ritem.LastPlayTime = item.Timestamp
			ritem.DeviceType = item.ToAppHistoryDeviceType()
		}
	}
	// 历史记录按播放时间打标
	resp.ApplyHistoryTag(req.LocalTodayZero)

	return
}

func (s *Service) PlayHistoryAdd(ctx context.Context, req *v1.PlayHistoryAddReq) (e *empty.Empty, err error) {
	if req.Duration <= 0 || req.Progress < -1 ||
		req.Item == nil || req.Duration < req.Progress {
		return nil, errors.WithMessagef(ecode.RequestErr, "empty req.Item or illegal req.Duration/Progress")
	}
	e = new(empty.Empty)
	err = validatePlayItem(ctx, req.Item, 1)
	if err != nil {
		return
	}
	req.Item.FixItemTypeByEt()
	dev, _, auth := DevNetAuthFromCtx(ctx)
	err = s.dao.PlayHistoryAdd(ctx, dao.PlayHistoryAddOpt{
		Mid:       auth.Mid,
		ArcType:   req.Item.ItemType,
		Oid:       req.Item.Oid,
		SubID:     req.Item.SubId[0],
		Progress:  req.Progress,
		Duration:  req.Duration,
		Dev:       dev,
		PlayStyle: req.PlayStyle,
		Scene:     req.Item.GetEt().GetOperator(),
	})
	return
}

func (s *Service) PlayHistoryDel(ctx context.Context, req *v1.PlayHistoryDelReq) (e *empty.Empty, err error) {
	e = new(empty.Empty)
	dev, _, auth := DevNetAuthFromCtx(ctx)
	if req.Truncate {
		err = s.dao.PlayHistoryTruncate(ctx, auth.Mid, dev.Buvid)
		return
	}

	if len(req.Items) <= 0 {
		return nil, errors.WithMessagef(ecode.RequestErr, "empty req.Items")
	}
	var items []*listenerSvc.PlayHistoryItem
	for _, item := range req.Items {
		err = validatePlayItem(ctx, item, 1)
		if err != nil {
			return
		}
		tmp := &listenerSvc.PlayHistoryItem{
			Mid:   auth.Mid,
			Buvid: dev.Buvid,
			Type:  int64(item.ItemType),
		}
		switch item.ItemType {
		case model.PlayItemUGC, model.PlayItemAudio:
			tmp.Aid = item.Oid
		case model.PlayItemOGV:
			tmp.Epid = item.Oid
		}
		items = append(items, tmp)
	}
	err = s.dao.PlayHistoryDelete(ctx, dao.PlayHistoryDeleteOpt{
		Mid:   auth.Mid,
		Dev:   dev,
		Items: items,
	})
	return
}

func (s *Service) PlayActionReport(ctx context.Context, req *v1.PlayActionReportReq) (e *empty.Empty, err error) {
	if err = validatePlayItem(ctx, req.Item, 1); err != nil {
		return
	}
	e = new(empty.Empty)

	dev, net, auth := DevNetAuthFromCtx(ctx)
	req.Item.FixItemTypeByEt()
	err = s.dao.PlayActionReport(ctx, dao.PlayActionReportOpt{
		Mid:    auth.Mid,
		Buvid:  dev.Buvid,
		Item:   req.Item,
		Device: dev, Network: net,
		FromSpmId: req.FromSpmid,
	})
	return
}
