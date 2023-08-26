package dao

import (
	"context"
	"fmt"
	"time"

	"go-common/component/metadata/device"
	"go-common/library/log"
	v1 "go-gateway/app/app-svr/app-listener/interface/api/v1"
	"go-gateway/app/app-svr/app-listener/interface/internal/model"

	arcSvc "go-gateway/app/app-svr/archive/service/api"

	hisSvc "git.bilibili.co/bapis/bapis-go/community/interface/history"
	listenerSvc "git.bilibili.co/bapis/bapis-go/dynamic/service/listener"
	ogvEpisodeSvc "git.bilibili.co/bapis/bapis-go/pgc/service/season/episode"
)

type HistoryAddOpt struct {
	Mid, Aid, Cid, Progress int64
	SeasonId, EpisodeId     int64
	BusinessTypeId          int32
	BusinessSubTypeId       int64
}

// 社区（全站）历史记录上报
func (d *dao) CommunityHistoryAdd(ctx context.Context, opt HistoryAddOpt, dev *device.Device) (err error) {
	req := &hisSvc.ReportReq{
		Mid:    opt.Mid,
		Buvid:  dev.Buvid,
		Oid:    opt.Aid,
		Sid:    opt.SeasonId,
		Epid:   opt.EpisodeId,
		Tp:     opt.BusinessTypeId,
		Stp:    opt.BusinessSubTypeId,
		Cid:    opt.Cid,
		Dt:     int32(dev.DevType()),
		Pro:    opt.Progress,
		ViewAt: time.Now().Unix(),
	}
	_, err = d.hisGRPC.Report(ctx, req)
	err = wrapDaoError(err, "hisGRPC.Report", req)
	return
}

func (d *dao) PlayHistory(ctx context.Context, mid int64, dev *device.Device) (list []model.PlayHistory, err error) {
	req := &listenerSvc.GetPlayHistoryListReq{
		Mid:   mid,
		Buvid: dev.Buvid,
	}
	resp, err := d.listenerGRPC.GetPlayHistoryList(ctx, req)
	if err != nil {
		err = wrapDaoError(err, "listenerGRPC.GetPlayHistoryList", req)
		return
	}
	if !resp.Success {
		return nil, fmt.Errorf("listener svc GetPlayHistoryList failed")
	}

	var lastPlay, aid int64
	for _, item := range resp.GetItems() {
		switch int32(item.Type) {
		case model.PlayItemUGC:
			aid = item.Aid
			lastPlay = item.Cid
		case model.PlayItemOGV:
			aid = item.Epid
			lastPlay = item.Cid
		case model.PlayItemAudio:
			aid = item.Aid
			lastPlay = item.Aid
		default:
			return nil, fmt.Errorf("unexpected PlayHistoryItem Type %d", item.Type)
		}
		list = append(list, model.PlayHistory{
			ArcType:    int32(item.Type),
			Oid:        aid,
			LastPlay:   lastPlay,
			Progress:   item.Progress,
			DeviceType: item.DeviceType,
			Timestamp:  item.Timestamp,
		})
	}
	return
}

//nolint:unused
func (d *dao) firstCid(ctx context.Context, aid int64) (cid int64, err error) {
	resp, err := d.arcGRPC.SimpleArc(ctx, &arcSvc.SimpleArcRequest{Aid: aid})
	if err != nil {
		err = wrapDaoError(err, "arcGRPC.SimpleArc", aid)
		return
	}
	if len(resp.GetArc().Cids) > 0 {
		return resp.Arc.Cids[0], nil
	}
	err = fmt.Errorf("error no cids found with aid %d", aid)
	return
}

func (d *dao) ogvGetEpisode(ctx context.Context, epid int64) (aid int64, err error) {
	resp, err := d.ogvEpisodeGRPC.AvInfos(ctx, &ogvEpisodeSvc.AvInfoReq{EpisodeId: []int32{int32(epid)}})
	if err != nil {
		err = wrapDaoError(err, "ogvEpisodeGRPC.AvInfos", epid)
		return
	}
	if resp.GetInfo() == nil {
		err = fmt.Errorf("error get OGV AvInfos: nil info")
		return
	}
	info, ok := resp.GetInfo()[int32(epid)]
	if !ok || info == nil {
		err = fmt.Errorf("error get OGV AvInfos: no aid found")
		return
	}
	return info.Aid, nil
}

type PlayHistoryAddOpt struct {
	Mid                            int64
	ArcType                        int32
	Oid, SubID, Progress, Duration int64
	Dev                            *device.Device
	PlayStyle                      int32
	Scene                          string
}

var playStyleLookup = map[int32]int64{
	// 默认走连播逻辑
	0: int64(_playRcmd),
	1: int64(_playRcmd),
	2: int64(_playVod),
}

func (d *dao) PlayHistoryAdd(ctx context.Context, opt PlayHistoryAddOpt) (err error) {
	req := &listenerSvc.AddPlayHistoryReq{
		Mid:        opt.Mid,
		Buvid:      opt.Dev.Buvid,
		Progress:   opt.Progress,
		Duration:   opt.Duration,
		Type:       int64(opt.ArcType),
		DeviceType: toHistoryDevType(opt.Dev),
		PlayStyle:  playStyleLookup[opt.PlayStyle],
		Scene:      opt.Scene,
	}
	switch opt.ArcType {
	case model.PlayItemUGC:
		req.Aid, req.Cid = opt.Oid, opt.SubID
	case model.PlayItemOGV:
		req.Sid, req.Epid = 0, opt.Oid
		req.Aid, err = d.ogvGetEpisode(ctx, req.Epid)
		if err != nil {
			return
		}
		//req.Cid, err = d.firstCid(ctx, req.Cid)
	case model.PlayItemAudio:
		req.Aid, req.Cid = opt.Oid, opt.Oid
	default:
		return fmt.Errorf("unexpected ArcType %d", opt.ArcType)
	}

	resp, err := d.listenerGRPC.AddPlayHistory(ctx, req)
	if err != nil {
		err = wrapDaoError(err, "listenerGRPC.AddPlayHistory", req)
		return
	}
	if !resp.Success {
		return fmt.Errorf("failed to AddPlayHistory")
	}
	if opt.ArcType == model.PlayItemAudio {
		// 成功之后针对旧音频报一下播放数
		err = d.MusicClickReport(ctx, MusicClickReportOpt{
			SongId: opt.Oid, ClickTyp: MusicClickPlay, AddMetric: true,
		})
		if err != nil {
			log.Errorc(ctx, "MusicClickReport failed to report play click: %v", err)
		}
	}

	return nil
}

type PlayHistoryDeleteOpt struct {
	Mid   int64
	Dev   *device.Device
	Items []*listenerSvc.PlayHistoryItem
}

func (d *dao) PlayHistoryDelete(ctx context.Context, opt PlayHistoryDeleteOpt) (err error) {
	req := &listenerSvc.DeletePlayHistoryReq{
		Mid:   opt.Mid,
		Buvid: opt.Dev.Buvid,
		Items: opt.Items,
	}
	resp, err := d.listenerGRPC.DeletePlayHistory(ctx, req)
	if err != nil {
		err = wrapDaoError(err, "listenerGRPC.DeletePlayHistory", req)
		return
	}
	if !resp.Success {
		return fmt.Errorf("failed to DeletePlayHistory")
	}
	return nil
}

func (d *dao) PlayHistoryTruncate(ctx context.Context, mid int64, buvid string) error {
	req := &listenerSvc.CleanPlayHistoryReq{Mid: mid, Buvid: buvid}
	resp, err := d.listenerGRPC.CleanPlayHistory(ctx, req)
	if err != nil {
		return wrapDaoError(err, "listenerGRPC.CleanPlayHistory", req)
	}
	if !resp.Success {
		return fmt.Errorf("failed to CleanPlayHistory")
	}
	return nil
}

func (d *dao) PlayHisoryByItemID(ctx context.Context, mid int64, buvid string, item *v1.PlayItem) (subid int64, progress int64, err error) {
	req := &listenerSvc.GetPlayHistoryReq{
		Mid: mid, Buvid: buvid,
		Items: []*listenerSvc.PlayHistoryItem{
			{Type: int64(item.ItemType)},
		},
	}
	switch item.ItemType {
	case model.PlayItemUGC, model.PlayItemAudio:
		req.Items[0].Aid = item.Oid
	case model.PlayItemOGV:
		req.Items[0].Epid = item.Oid
	}
	resp, err := d.listenerGRPC.GetPlayHistoryByIds(ctx, req)
	if err != nil {
		err = wrapDaoError(err, "listenerGRPC.GetPlayHistoryByIds", req)
		return
	}
	if !resp.Success || len(resp.Items) == 0 {
		return 0, 0, model.ErrNoHistoryRecord
	}
	progress = resp.Items[0].Progress
	switch item.ItemType {
	case model.PlayItemUGC:
		subid = resp.Items[0].Cid
	case model.PlayItemOGV:
		subid = resp.Items[0].Epid
	case model.PlayItemAudio:
		subid = resp.Items[0].Aid
	}

	return
}

type PlayHistoryResult struct {
	Item     *v1.PlayItem
	SubId    int64
	Progress int64
}

func (d *dao) PlayHisoryByItemIDs(ctx context.Context, mid int64, buvid string, items ...*v1.PlayItem) (ret map[string]PlayHistoryResult, err error) {
	ret = make(map[string]PlayHistoryResult)
	if len(items) <= 0 {
		return
	}

	req := &listenerSvc.GetPlayHistoryReq{
		Mid: mid, Buvid: buvid,
		Items: make([]*listenerSvc.PlayHistoryItem, 0, len(items)),
	}
	for _, it := range items {
		tgt := &listenerSvc.PlayHistoryItem{
			Type: int64(it.ItemType),
		}
		switch it.ItemType {
		case model.PlayItemUGC, model.PlayItemAudio:
			tgt.Aid = it.Oid
		case model.PlayItemOGV:
			tgt.Epid = it.Oid
		}
		if tgt.Aid != 0 || tgt.Epid != 0 {
			req.Items = append(req.Items, tgt)
		}
	}
	if len(req.Items) <= 0 {
		return
	}

	resp, err := d.listenerGRPC.GetPlayHistoryByIds(ctx, req)
	if err != nil {
		err = wrapDaoError(err, "listenerGRPC.GetPlayHistoryByIds", req)
		return
	}
	if !resp.Success || len(resp.Items) == 0 {
		return
	}
	for _, r := range resp.Items {
		if r.Epid == 0 && r.Aid == 0 {
			continue
		}
		res := PlayHistoryResult{
			Item: &v1.PlayItem{
				ItemType: int32(r.Type),
			},
			Progress: r.Progress,
		}
		switch int32(r.Type) {
		case model.PlayItemUGC:
			res.Item.Oid = r.Aid
			res.Item.SubId = []int64{r.Cid}
		case model.PlayItemOGV:
			res.Item.Oid = r.Epid
			res.Item.SubId = []int64{r.Epid}
		case model.PlayItemAudio:
			res.Item.Oid = r.Aid
			res.Item.SubId = []int64{r.Aid}
		default:
			log.Errorc(ctx, "unknown play history type(%d): %+v", r.Type, r)
			continue
		}
		res.SubId = res.Item.SubId[0]
		ret[res.Item.Hash()] = res
	}
	return
}
