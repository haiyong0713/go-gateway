package examination

import (
	"context"
	"go-common/library/log"
	"go-gateway/app/web-svr/activity/interface/client"
	exammdl "go-gateway/app/web-svr/activity/interface/model/examination"
	"time"

	"go-common/library/sync/errgroup.v2"
	likemdl "go-gateway/app/web-svr/activity/interface/model/like"

	liveapi "git.bilibili.co/bapis/bapis-go/live/xroom"
)

const (
	notPlayUrl = 1
)

func (s *Service) UpInfo(ctx context.Context, mid int64, path string, req *exammdl.UpReq) (res *exammdl.UpRes, err error) {
	mids := make([]int64, 0)
	sids := make([]int64, 0)
	todayMids := make([]int64, 0)
	member := make(map[int64]*exammdl.Account)
	live := make(map[int64]*exammdl.Live)
	reserve := make(map[int64]*exammdl.Reserve)
	res = &exammdl.UpRes{
		TimeStamp: time.Now().UnixNano() / 1e6,
	}
	alreadyInMapMids := make(map[int64]struct{})
	alreadyInMapSids := make(map[int64]struct{})
	alreadyIntodayMids := make(map[int64]struct{})
	if len(req.UpOther) == 0 && len(req.UpToday) == 0 {
		return
	}
	if req.UpOther != nil {
		for _, v := range req.UpOther {
			if _, ok := alreadyInMapMids[v.MID]; !ok {
				mids = append(mids, v.MID)
				alreadyInMapMids[v.MID] = struct{}{}
			}
			if _, ok := alreadyInMapSids[v.SID]; !ok {
				sids = append(sids, v.SID)
				alreadyInMapSids[v.SID] = struct{}{}
			}

		}
	}
	if req.UpToday != nil {
		for _, v := range req.UpToday {
			if _, ok := alreadyInMapMids[v.MID]; !ok {
				mids = append(mids, v.MID)
				alreadyInMapMids[v.MID] = struct{}{}
			}
			if _, ok := alreadyIntodayMids[v.MID]; !ok {
				todayMids = append(todayMids, v.MID)
				alreadyIntodayMids[v.MID] = struct{}{}
			}
			if _, ok := alreadyInMapSids[v.SID]; !ok {
				sids = append(sids, v.SID)
				alreadyInMapSids[v.SID] = struct{}{}
			}
		}
	}
	eg := errgroup.WithContext(ctx)
	eg.Go(func(ctx context.Context) (err error) {
		if member, err = s.accountInfo(ctx, mids); err != nil {
			log.Errorc(ctx, "s.accountInfo err(%v)", err)
		}
		return
	})
	if len(todayMids) > 0 {
		eg.Go(func(ctx context.Context) (err error) {
			if live, err = s.liveInfo(ctx, path, todayMids); err != nil {
				log.Errorc(ctx, "s.liveInfo err(%v)", err)
			}
			return
		})
	}
	if mid > 0 {
		eg.Go(func(ctx context.Context) (err error) {
			if reserve, err = s.reserveInfo(ctx, mid, sids); err != nil {
				log.Errorc(ctx, "s.reserveInfo err(%v)", err)
			}
			return
		})
	}

	if err = eg.Wait(); err != nil {
		log.Errorc(ctx, "eg.Wait error(%v)", err)
		return
	}
	if len(req.UpToday) > 0 {
		for _, v := range req.UpToday {
			upInfo := &exammdl.UpInfo{}
			if m, ok := member[v.MID]; ok {
				upInfo.Account = m
			} else {
				log.Errorc(ctx, "account can not find mid(%d)", v.MID)
				continue
			}

			if r, ok := reserve[v.SID]; ok {
				upInfo.Reserve = r
			} else {
				upInfo.Reserve = &exammdl.Reserve{SID: v.SID}
			}
			if l, ok := live[v.MID]; ok {
				upInfo.Live = l
			}
			res.UpToday = append(res.UpToday, upInfo)
		}
	}
	if len(req.UpOther) > 0 {
		for _, v := range req.UpOther {
			upInfo := &exammdl.UpInfo{}
			if m, ok := member[v.MID]; ok {
				upInfo.Account = m
			} else {
				log.Errorc(ctx, "account can not find mid(%d)", v.MID)
				continue
			}
			if r, ok := reserve[v.SID]; ok {
				upInfo.Reserve = r
			} else {
				upInfo.Reserve = &exammdl.Reserve{SID: v.SID}
			}
			res.UpOther = append(res.UpOther, upInfo)
		}
	}

	return
}

// accountInfo
func (s *Service) accountInfo(ctx context.Context, mids []int64) (res map[int64]*exammdl.Account, err error) {
	res = make(map[int64]*exammdl.Account)
	data, err := s.account.MemberInfo(ctx, mids)
	if err != nil {
		log.Errorc(ctx, "s.account.MemberInfo err(%v)", err)
		return
	}
	if len(data) > 0 {
		for _, v := range data {
			res[v.Mid] = &exammdl.Account{
				MID:  v.Mid,
				Name: v.Name,
				Sex:  v.Sex,
				Face: v.Face,
				Sign: v.Sign,
			}
		}
	}
	return
}

// liveInfo
func (s *Service) liveInfo(ctx context.Context, path string, mids []int64) (res map[int64]*exammdl.Live, err error) {
	newMids := mids
	if len(mids) > 50 {
		newMids = mids[:50]
	}
	res = make(map[int64]*exammdl.Live)
	var liveRes = &liveapi.EntryRoomInfoResp{}
	data := &liveapi.EntryRoomInfoReq{
		Uids:       newMids,
		NotPlayurl: notPlayUrl,
		EntryFrom:  []string{"None"},
		ReqBiz:     path,
	}
	liveRes, err = client.LiveClient.EntryRoomInfo(ctx, data)
	if err != nil {
		log.Errorc(ctx, " client.LiveClient.EntryRoomInfo data(%+v) err(%v)", data, err)
		return
	}
	if liveRes != nil {
		if len(liveRes.List) > 0 {
			for _, v := range liveRes.List {
				res[v.Uid] = &exammdl.Live{
					LiveStatus: v.LiveStatus,
					Title:      v.Title,
				}
			}
		}
	}
	return
}

func (s *Service) reserveInfo(ctx context.Context, mid int64, sid []int64) (res map[int64]*exammdl.Reserve, err error) {
	var reserveRes = make(map[int64]*likemdl.ActFollowingReply)
	reserveRes, err = s.likeSvr.ReserveFollowings(ctx, sid, mid)
	res = make(map[int64]*exammdl.Reserve)
	if err != nil {
		log.Errorc(ctx, " s.likeSvr.ReserveFollowings err(%v)", err)
		return
	}
	if reserveRes != nil {
		for sid, v := range reserveRes {
			res[sid] = &exammdl.Reserve{
				IsFollowing: v.IsFollowing,
				SID:         sid,
			}
		}
	}
	return
}
