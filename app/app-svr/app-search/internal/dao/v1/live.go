package v1

import (
	"context"
	"net/url"
	"strconv"
	"sync"

	"go-common/library/ecode"
	"go-common/library/log"
	"go-common/library/net/metadata"
	"go-common/library/sync/errgroup.v2"

	"go-gateway/app/app-svr/app-search/internal/model/search"

	livexfans "git.bilibili.co/bapis/bapis-go/live/xfansmedal"
	livexroom "git.bilibili.co/bapis/bapis-go/live/xroom"
	livexroomgate "git.bilibili.co/bapis/bapis-go/live/xroom-gate"

	"github.com/pkg/errors"
)

const (
	lpSwitchOn = 1
)

func (d *dao) QueryMedalStatus(c context.Context, mid int64) (status int64, err error) {
	var (
		arg  = &livexfans.QueryMedalReq{UpUid: mid}
		resp *livexfans.QueryMedalResp
	)
	if resp, err = d.liveRpcClient.QueryMedal(c, arg); err != nil || resp == nil {
		log.Error("livexfans grpc querymedal error(%v) or is resp null", err)
		return
	}
	if medal := resp.UpMedal; medal != nil {
		status = medal.MasterStatus
	}
	return
}

func (d *dao) LiveGetMultiple(ctx context.Context, roomIDs []int64) (map[int64]*livexroom.Infos, error) {
	var roomIDsFilter []int64
	for _, roomID := range roomIDs {
		if roomID != 0 {
			roomIDsFilter = append(roomIDsFilter, roomID)
		}
	}
	if len(roomIDsFilter) == 0 {
		return nil, ecode.NothingFound
	}
	var max50 = 50
	info := make(map[int64]*livexroom.Infos, len(roomIDsFilter))
	eg, mu := errgroup.WithContext(ctx), sync.Mutex{}
	for i := 0; i < len(roomIDsFilter); i += max50 {
		partRoomIds := roomIDsFilter[i:]
		if i+max50 < len(roomIDsFilter) {
			partRoomIds = roomIDsFilter[i : i+max50]
		}
		eg.Go(func(ctx context.Context) (err error) {
			var (
				arg  = &livexroom.RoomIDsReq{RoomIds: partRoomIds, Attrs: []string{"show", "status", "area"}}
				resp *livexroom.RoomIDsInfosResp
			)
			resp, err = d.roomRPCClient.GetMultiple(ctx, arg)
			if err != nil || resp == nil {
				log.Error("GetMultiple d.roomRPCClient.GetMultiple error(%v)", err)
				return
			}
			mu.Lock()
			for liveId, v := range resp.List {
				if v == nil {
					continue
				}
				info[liveId] = v
			}
			mu.Unlock()
			return nil
		})
	}
	if err := eg.Wait(); err != nil {
		return nil, err
	}
	return info, nil
}

func (d *dao) EntryRoomInfo(ctx context.Context, req *livexroomgate.EntryRoomInfoReq) (map[int64]*livexroomgate.EntryRoomInfoResp_EntryList, error) {
	var roomIDsFilter []int64
	for _, roomID := range req.RoomIds {
		if roomID != 0 {
			roomIDsFilter = append(roomIDsFilter, roomID)
		}
	}
	req.RoomIds = roomIDsFilter

	var upMidsFilter []int64
	for _, upMid := range req.Uids {
		if upMid != 0 {
			upMidsFilter = append(upMidsFilter, upMid)
		}
	}
	req.Uids = upMidsFilter

	if len(req.RoomIds) == 0 && len(req.Uids) == 0 {
		return map[int64]*livexroomgate.EntryRoomInfoResp_EntryList{}, nil
	}
	reply, err := d.roomGateClient.EntryRoomInfo(ctx, req)
	if err != nil {
		return nil, err
	}
	return reply.List, nil
}

func (d *dao) GetMultipleWithPlayUrl(c context.Context, roomIDs []int64, param *search.LiveParam) (map[int64]*livexroom.Infos, map[int64]*livexroom.LivePlayUrlData, error) {
	var roomIDsFilter []int64
	for _, roomID := range roomIDs {
		if roomID != 0 {
			roomIDsFilter = append(roomIDsFilter, roomID)
		}
	}
	if len(roomIDsFilter) == 0 {
		return nil, nil, ecode.NothingFound
	}
	var max50 = 50
	infos, playUrls := make(map[int64]*livexroom.Infos, len(roomIDsFilter)), make(map[int64]*livexroom.LivePlayUrlData, len(roomIDsFilter))
	eg, mu := errgroup.WithContext(c), sync.Mutex{}
	for i := 0; i < len(roomIDsFilter); i += max50 {
		partRoomIds := roomIDsFilter[i:]
		if i+max50 < len(roomIDsFilter) {
			partRoomIds = roomIDsFilter[i : i+max50]
		}
		eg.Go(func(ctx context.Context) (err error) {
			arg := &livexroom.RoomIDsReq{
				RoomIds: partRoomIds,
				Attrs:   []string{"show", "status", "area"},
			}
			if param != nil { //使用直播下发的链接
				arg.Playurl = &livexroom.PlayURLParams{
					Uid:        param.Uid,
					Uipstr:     metadata.String(ctx, metadata.RemoteIP),
					Build:      param.Build,
					Platform:   param.Platform,
					Switch:     lpSwitchOn,
					ReqBiz:     param.ReqBiz,
					DeviceName: param.DeviceName,
					Network:    param.NetWork,
				}
			}
			resp, err := d.roomRPCClient.GetMultiple(ctx, arg)
			if err != nil || resp == nil {
				return errors.Wrapf(err, "GetMultiple d.roomRPCClient.GetMultiple arg(%v)", arg)
			}
			mu.Lock()
			for liveId, info := range resp.List {
				if info == nil {
					continue
				}
				infos[liveId] = info
			}
			for liveId, playUrl := range resp.PlayUrl {
				if playUrl == nil {
					continue
				}
				playUrls[liveId] = playUrl
			}
			mu.Unlock()
			return nil
		})
	}
	if err := eg.Wait(); err != nil {
		return nil, nil, err
	}
	return infos, playUrls, nil
}

func (d *dao) GetMultiple(ctx context.Context, roomIDs []int64) (map[int64]*livexroom.Infos, error) {
	var roomIDsFilter []int64
	for _, roomID := range roomIDs {
		if roomID != 0 {
			roomIDsFilter = append(roomIDsFilter, roomID)
		}
	}
	if len(roomIDsFilter) == 0 {
		return nil, ecode.NothingFound
	}
	var max50 = 50
	info := make(map[int64]*livexroom.Infos, len(roomIDsFilter))
	eg, mu := errgroup.WithContext(ctx), sync.Mutex{}
	for i := 0; i < len(roomIDsFilter); i += max50 {
		partRoomIds := roomIDsFilter[i:]
		if i+max50 < len(roomIDsFilter) {
			partRoomIds = roomIDsFilter[i : i+max50]
		}
		eg.Go(func(ctx context.Context) (err error) {
			var (
				arg  = &livexroom.RoomIDsReq{RoomIds: partRoomIds, Attrs: []string{"show", "status", "area"}}
				resp *livexroom.RoomIDsInfosResp
			)
			resp, err = d.roomRPCClient.GetMultiple(ctx, arg)
			if err != nil || resp == nil {
				log.Error("GetMultiple d.roomRPCClient.GetMultiple error(%v)", err)
				return
			}
			mu.Lock()
			for liveId, v := range resp.List {
				if v == nil {
					continue
				}
				info[liveId] = v
			}
			mu.Unlock()
			return nil
		})
	}
	if err := eg.Wait(); err != nil {
		return nil, err
	}
	return info, nil
}

func (d *dao) AppMRoom(c context.Context, roomids []int64, platform string) (rs map[int64]*search.Room, err error) {
	ip := metadata.String(c, metadata.RemoteIP)
	params := url.Values{}
	params.Set("platform", platform)
	for _, roomid := range roomids {
		params.Add("room_ids", strconv.FormatInt(roomid, 10))
	}
	var res struct {
		Code int `json:"code"`
		Data struct {
			List []*search.Room `json:"list"`
		} `json:"data"`
	}
	if err = d.liveClient.Get(c, d.appMRoom, ip, params, &res); err != nil {
		return
	}
	if res.Code != ecode.OK.Code() {
		err = errors.Wrap(err, d.appMRoom+"?"+params.Encode())
		return
	}
	if list := res.Data.List; len(list) > 0 {
		rs = make(map[int64]*search.Room, len(list))
		for _, r := range list {
			rs[r.RoomID] = r
		}
	}
	return
}

// Glory for live search
func (d *dao) LiveGlory(c context.Context, uid int64) (glory []*search.LiveGlory, err error) {
	ip := metadata.String(c, metadata.RemoteIP)
	params := url.Values{}
	params.Set("uid", strconv.FormatInt(uid, 10))
	var res struct {
		Code int                 `json:"code"`
		Data []*search.LiveGlory `json:"data"`
	}
	if err = d.liveClient.Get(c, d.visibleInfo, ip, params, &res); err != nil {
		return
	}
	if res.Code != ecode.OK.Code() {
		err = errors.Wrap(err, d.visibleInfo+"?"+params.Encode())
		return
	}
	glory = res.Data
	return
}

// UserInfo for live search
func (d *dao) UserInfo(c context.Context, uids []int64) (userInfo map[int64]map[string]*search.Exp, err error) {
	ip := metadata.String(c, metadata.RemoteIP)
	params := url.Values{}
	for _, uid := range uids {
		params.Set("uids[]", strconv.FormatInt(uid, 10))
	}
	params.Set("attributes[]", "exp")
	var res struct {
		Code int                              `json:"code"`
		Data map[int64]map[string]*search.Exp `json:"data"`
	}
	if err = d.liveClient.Get(c, d.userInfo, ip, params, &res); err != nil {
		return
	}
	if res.Code != ecode.OK.Code() {
		err = errors.Wrap(err, d.userInfo+"?"+params.Encode())
		return
	}
	userInfo = res.Data
	return
}
