package live

import (
	"context"
	"net/url"
	"strconv"
	"sync"

	"go-common/library/ecode"
	"go-common/library/log"
	"go-common/library/net/metadata"
	"go-common/library/sync/errgroup.v2"
	mdlv2 "go-gateway/app/app-svr/app-dynamic/interface/model/dynamicV2"
	livemdl "go-gateway/app/app-svr/app-dynamic/interface/model/live"

	livexroom "git.bilibili.co/bapis/bapis-go/live/xroom"
	livexroomfeed "git.bilibili.co/bapis/bapis-go/live/xroom-feed"
	livexroomgate "git.bilibili.co/bapis/bapis-go/live/xroom-gate"

	"github.com/pkg/errors"
)

const (
	_liveDetailsByIDs = "/room/v2/Room/get_by_ids"
)

func (d *Dao) LiveDetailsByIDs(c context.Context, ids []int64) (map[int64]*livemdl.LiveResItem, error) {
	params := url.Values{}
	params.Set("need_uinfo", "1")
	params.Set("need_broadcast_type", "1")
	for _, id := range ids {
		params.Add("ids[]", strconv.FormatInt(id, 10))
	}
	liveDetailURL := d.c.Hosts.LiveCo + _liveDetailsByIDs
	var ret struct {
		Code int                            `json:"code"`
		Msg  string                         `json:"msg"`
		Data map[int64]*livemdl.LiveResItem `json:"data"`
	}
	if err := d.client.Get(c, liveDetailURL, "", params, &ret); err != nil {
		log.Errorc(c, "LiveDetailsByIDs http GET(%s) failed, params:(%s), error(%+v)", liveDetailURL, params.Encode(), err)
		return nil, err
	}
	if ret.Code != 0 {
		log.Errorc(c, "LiveDetailsByIDs http GET(%s) failed, params:(%s), code: %v, msg: %v", liveDetailURL, params.Encode(), ret.Code, ret.Msg)
		err := errors.Wrapf(ecode.Int(ret.Code), "LiveDetailsByIDs url(%v) code(%v) msg(%v)", liveDetailURL, ret.Code, ret.Msg)
		return nil, err
	}
	return ret.Data, nil
}

func (d *Dao) EntryRoomInfo(c context.Context, roomids []int64, uids []int64, mid, build int64, platform string) (map[int64]*livexroomgate.EntryRoomInfoResp_EntryList, error) {
	var max50 = 50
	g := errgroup.WithContext(c)
	mu := sync.Mutex{}
	lives := make(map[int64]*livexroomgate.EntryRoomInfoResp_EntryList)
	for i := 0; i < len(uids); i += max50 {
		var partUids []int64
		if i+max50 > len(uids) {
			partUids = uids[i:]
		} else {
			partUids = uids[i : i+max50]
		}
		g.Go(func(ctx context.Context) error {
			arg := &livexroomgate.EntryRoomInfoReq{
				ReqBiz:    "/bilibili.app.dynamic.v2.Dynamic/DynAll",
				EntryFrom: []string{"NONE", "dt_top_live_card"},
				Uids:      partUids,
				Uid:       mid,
				Platform:  platform,
				Build:     build,
				Network:   "other",
				Uipstr:    metadata.String(ctx, metadata.RemoteIP),
			}
			reply, err := d.livexroomgategrpc.EntryRoomInfo(ctx, arg)
			if err != nil {
				log.Error("%+v", err)
				return err
			}
			mu.Lock()
			for k, v := range reply.List {
				lives[k] = v
			}
			mu.Unlock()
			return nil
		})
	}
	for i := 0; i < len(roomids); i += max50 {
		var partRoomids []int64
		if i+max50 > len(roomids) {
			partRoomids = roomids[i:]
		} else {
			partRoomids = roomids[i : i+max50]
		}
		g.Go(func(ctx context.Context) error {
			arg := &livexroomgate.EntryRoomInfoReq{
				ReqBiz:    "/bilibili.app.dynamic.v2.Dynamic/DynAll",
				EntryFrom: []string{"NONE", "dt_top_live_card"},
				RoomIds:   partRoomids,
				Uid:       mid,
				Platform:  platform,
				Build:     build,
				Network:   "other",
				Uipstr:    metadata.String(ctx, metadata.RemoteIP),
			}
			reply, err := d.livexroomgategrpc.EntryRoomInfo(ctx, arg)
			if err != nil {
				log.Error("%+v", err)
				return err
			}
			mu.Lock()
			for k, v := range reply.List {
				lives[k] = v
			}
			mu.Unlock()
			return nil
		})
	}
	if err := g.Wait(); err != nil {
		log.Error("EntryRoomInfo uids(%+v) roomids(%+v) eg.wait(%+v)", uids, roomids, err)
		return nil, err
	}
	return lives, nil
}

func (d *Dao) LiveInfos(c context.Context, uids []int64, general *mdlv2.GeneralParam) (map[int64]*livexroom.Infos, map[int64]*livexroom.LivePlayUrlData, error) {
	var max50 = 50
	g := errgroup.WithContext(c)
	mu := sync.Mutex{}
	lives := make(map[int64]*livexroom.Infos)
	playurls := make(map[int64]*livexroom.LivePlayUrlData)
	for i := 0; i < len(uids); i += max50 {
		var partUids []int64
		if i+max50 > len(uids) {
			partUids = uids[i:]
		} else {
			partUids = uids[i : i+max50]
		}
		g.Go(func(ctx context.Context) (err error) {
			ls, ps, err := d.LiveInfosSlice(ctx, partUids, general)
			if err != nil {
				return err
			}
			mu.Lock()
			for uid, l := range ls {
				lives[uid] = l
			}
			for uid, p := range ps {
				playurls[uid] = p
			}
			mu.Unlock()
			return
		})
	}
	if err := g.Wait(); err != nil {
		log.Error("LiveInfos uids(%+v) eg.wait(%+v)", uids, err)
		return nil, nil, err
	}
	return lives, playurls, nil
}

func (d *Dao) LiveInfosSlice(c context.Context, uids []int64, general *mdlv2.GeneralParam) (map[int64]*livexroom.Infos, map[int64]*livexroom.LivePlayUrlData, error) {
	resTmp, err := d.livexroomgrpc.GetMultipleByUids(c, &livexroom.UIDsReq{
		Uids:  uids,
		Attrs: []string{"show", "status", "area", "pendants"},
		Playurl: &livexroom.PlayURLParams{
			Switch:   1,
			ReqBiz:   "/bilibili.app.dynamic.v2.Dynamic/DynAll",
			Uipstr:   metadata.String(c, metadata.RemoteIP),
			Uid:      general.Mid,
			Platform: general.GetPlatform(),
			Build:    general.GetBuild(),
		},
	})
	if err != nil {
		log.Error("%+v", err)
		return nil, nil, err
	}
	return resTmp.List, resTmp.PlayUrl, nil
}

func (d *Dao) LiveRcmdInfos(c context.Context, uid, build int64, platform, device string, liveids []uint64) (map[int64]*livexroomfeed.HistoryCardInfo, error) {
	ip := metadata.String(c, metadata.RemoteIP)
	resTmp, err := d.livexroomfeedgrpc.GetHistoryCardInfo(c, &livexroomfeed.GetHistoryCardInfoReq{
		Uid:         uid,
		Platform:    platform,
		Ip:          ip,
		DeviceName:  device,
		Build:       build,
		IsHttps:     false,
		LiveIds:     liveids,
		NeedPlayUrl: true,
	})
	if err != nil {
		log.Error("LiveRcmdInfo mid %v, roomids %v", uid, liveids)
		return nil, err
	}
	var res map[int64]*livexroomfeed.HistoryCardInfo
	for id, card := range resTmp.HistoryCardsMap {
		if res == nil {
			res = make(map[int64]*livexroomfeed.HistoryCardInfo)
		}
		res[int64(id)] = card
	}
	return res, nil
}

func (d *Dao) SessionInfo(c context.Context, liveAdditionals map[int64][]string, general *mdlv2.GeneralParam) (map[string]*livexroomgate.SessionInfos, error) {
	g := errgroup.WithContext(c)
	mu := sync.Mutex{}
	res := make(map[string]*livexroomgate.SessionInfos)
	for mid, liveIds := range liveAdditionals {
		for _, liveId := range liveIds {
			upmid := mid
			lid := liveId
			uplives := map[int64]*livexroomgate.LiveIds{upmid: {LiveIds: []string{lid}}}
			g.Go(func(ctx context.Context) (err error) {
				req := &livexroomgate.SessionInfoBatchReq{
					UidLiveIds: uplives,
					EntryFrom:  []string{"dt_booking_dt"},
					Playurl: &livexroomgate.PlayUrlReq{
						ReqBiz:     "/bilibili.app.dynamic.v2.Dynamic/DynAll",
						Uipstr:     metadata.String(c, metadata.RemoteIP),
						Uid:        general.Mid,
						Platform:   general.GetPlatform(),
						Build:      general.GetBuild(),
						DeviceName: general.GetDevice(),
						Network:    "other",
					},
				}
				reply, err := d.livexroomgategrpc.SessionInfoBatch(ctx, req)
				if err != nil {
					log.Error("%+v", err)
					return err
				}
				mu.Lock()
				if item, ok := reply.List[upmid]; ok {
					res[lid] = item
				}
				mu.Unlock()
				return nil
			})
		}
	}
	if err := g.Wait(); err != nil {
		return nil, err
	}
	return res, nil
}
