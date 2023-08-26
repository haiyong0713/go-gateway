package dao

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	v1 "go-gateway/app/app-svr/app-listener/interface/api/v1"
	"go-gateway/app/app-svr/app-listener/interface/internal/model"

	listenerSvc "git.bilibili.co/bapis/bapis-go/dynamic/service/listener"
)

func (d *dao) Playlist(ctx context.Context, mid int64, buvid string) (ret model.Playlist, err error) {
	req := &listenerSvc.GetPlayListReq{Mid: mid, Buvid: buvid}
	list, err := d.listenerGRPC.GetPlayList(ctx, req)
	if err != nil {
		err = wrapDaoError(err, "listenerGRPC.GetPlayList", req)
		return
	}
	if !list.Success {
		return model.Playlist{}, fmt.Errorf("listener svc failed to GetPlayList")
	}
	ret = model.Playlist{Items: make([]*v1.PlayItem, 0, len(list.Items))}
	if len(list.TrackId) > 0 {
		const _splitParts = 3
		split := strings.SplitN(list.TrackId, "-", _splitParts)
		if len(split) != _splitParts {
			err = fmt.Errorf("dao.listenerGRPC.GetPlayList: unexpected TrackID(%s) req(%+v)", list.TrackId, req)
			return
		}
		from, err := strconv.ParseInt(split[0], 10, 64)
		if err != nil {
			return ret, fmt.Errorf("error paring PlaylistSource from(%s)", split[0])
		}
		ret.From = v1.PlaylistSource(from)
		ret.Batch, ret.TrackID = split[1], split[2]
	}
	for _, item := range list.Items {
		ret.Items = append(ret.Items, &v1.PlayItem{ItemType: int32(item.Type), Oid: item.Aid})
	}
	return
}

type PlaylistAddOpt struct {
	Mid        int64
	Buvid      string
	Items      []*v1.PlayItem
	AfterItem  *v1.PlayItem
	Head, Tail bool
}

func (d *dao) PlaylistAdd(ctx context.Context, opt PlaylistAddOpt) (err error) {
	req := &listenerSvc.AddPlayItemReq{
		Mid: opt.Mid, Buvid: opt.Buvid,
	}
	var pre *listenerSvc.PlayItem
	if opt.Tail {
		tmpReq := &listenerSvc.GetPlayListReq{Mid: opt.Mid, Buvid: opt.Buvid}
		pList, err := d.listenerGRPC.GetPlayList(ctx, tmpReq)
		if err != nil {
			return wrapDaoError(err, "listenerGRPC.GetPlayList", tmpReq)
		}
		if !pList.Success {
			return fmt.Errorf("listener svc failed to GetPlayList")
		}
		if len(pList.Items) != 0 {
			pre = pList.Items[len(pList.Items)-1]
		}
	} else if opt.AfterItem != nil {
		pre = &listenerSvc.PlayItem{Aid: opt.AfterItem.Oid, Type: int64(opt.AfterItem.ItemType)}
	}

	req.Pre = pre
	for _, item := range opt.Items {
		req.Items = append(req.Items, &listenerSvc.PlayItem{Aid: item.Oid, Type: int64(item.ItemType)})
	}
	resp, err := d.listenerGRPC.AddPlayItem(ctx, req)
	if err != nil {
		err = wrapDaoError(err, "listenerGRPC.AddPlayItem", req)
		return
	}
	if !resp.Success {
		return fmt.Errorf("listener svc failed to AddPlayItem")
	}
	return
}

type PlaylistDeleteOpt struct {
	Mid   int64
	Buvid string
	Items []*v1.PlayItem
}

func (d *dao) PlaylistDelete(ctx context.Context, opt PlaylistDeleteOpt) (err error) {
	req := &listenerSvc.DeletePlayItemsReq{
		Mid: opt.Mid, Buvid: opt.Buvid,
	}
	for _, item := range opt.Items {
		req.Items = append(req.Items, &listenerSvc.PlayItem{Type: int64(item.ItemType), Aid: item.Oid})
	}

	resp, err := d.listenerGRPC.DeletePlayItems(ctx, req)
	if err != nil {
		err = wrapDaoError(err, "listenerGRPC.DeletePlayItems", req)
		return
	}
	if !resp.Success {
		return fmt.Errorf("listener svc failed to DeletePlayItems")
	}
	return
}

func (d *dao) PlaylistTruncate(ctx context.Context, mid int64, buvid string) (err error) {
	req := &listenerSvc.ReplacePlayListReq{
		Mid:   mid,
		Buvid: buvid,
	}
	resp, err := d.listenerGRPC.ReplacePlayList(ctx, req)
	if err != nil {
		err = wrapDaoError(err, "listenerGRPC.ReplacePlayList", req)
		return
	}
	if !resp.Success {
		return fmt.Errorf("listener svc failed to ReplacePlayList")
	}
	return
}

type PlaylistReplaceOpt struct {
	Mid     int64
	Buvid   string
	Items   []*v1.PlayItem
	From    v1.PlaylistSource
	Batch   int64
	TrackID int64
}

func (d *dao) PlaylistReplace(ctx context.Context, opt PlaylistReplaceOpt) (rets []*v1.PlayItem, err error) {
	ritems := make([]*listenerSvc.PlayItem, 0, len(opt.Items))
	for _, item := range opt.Items {
		ritems = append(ritems, &listenerSvc.PlayItem{Aid: item.Oid, Type: int64(item.ItemType)})
	}
	req := &listenerSvc.ReplacePlayListReq{
		Mid: opt.Mid, Buvid: opt.Buvid, Items: ritems, TrackId: fmt.Sprintf("%d-%d-%d", opt.From, opt.Batch, opt.TrackID),
	}
	resp, err := d.listenerGRPC.ReplacePlayList(ctx, req)
	if err != nil {
		err = wrapDaoError(err, "listenerGRPC.ReplacePlayList", req)
		return
	}
	if !resp.GetSuccess() {
		return nil, fmt.Errorf("listener svc failed to ReplacePlayList")
	}
	rets = make([]*v1.PlayItem, 0, len(resp.GetItems()))
	for _, item := range resp.GetItems() {
		rets = append(rets, &v1.PlayItem{
			ItemType: int32(item.Type),
			Oid:      item.Aid,
		})
	}
	return
}
