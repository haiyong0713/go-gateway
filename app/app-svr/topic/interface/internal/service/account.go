package service

import (
	"context"
	"net/url"
	"sync"

	"go-common/library/ecode"
	"go-common/library/log"
	"go-common/library/net/metadata"
	"go-common/library/sync/errgroup.v2"
	"go-common/library/xstr"
	mdlaccount "go-gateway/app/app-svr/app-dynamic/interface/model/account"

	accountapi "git.bilibili.co/bapis/bapis-go/account/service"
	relationgrpc "git.bilibili.co/bapis/bapis-go/account/service/relation"
	playurlgrpc "git.bilibili.co/bapis/bapis-go/playurl/service"

	"github.com/pkg/errors"
)

func (s *Service) getDecorateCards(c context.Context, uids []int64) (map[int64]*mdlaccount.DecoCards, error) {
	var max50 = 50
	g := errgroup.WithContext(c)
	mu := sync.Mutex{}
	res := make(map[int64]*mdlaccount.DecoCards)
	for i := 0; i < len(uids); i += max50 {
		var partUids []int64
		if i+max50 > len(uids) {
			partUids = uids[i:]
		} else {
			partUids = uids[i : i+max50]
		}
		g.Go(func(ctx context.Context) (err error) {
			dcs, err := s.decorateCardsSlice(ctx, partUids)
			if err != nil {
				return err
			}
			mu.Lock()
			for uid, dc := range dcs {
				res[uid] = dc
			}
			mu.Unlock()
			return
		})
	}
	if err := g.Wait(); err != nil {
		log.Error("getDecorateCards uids(%+v) eg.wait(%+v)", uids, err)
		return nil, err
	}
	return res, nil
}

func (s *Service) decorateCardsSlice(c context.Context, uids []int64) (map[int64]*mdlaccount.DecoCards, error) {
	params := url.Values{}
	params.Set("mids", xstr.JoinInts(uids))
	var ret struct {
		Code int                             `json:"code"`
		Msg  string                          `json:"message"`
		Data map[int64]*mdlaccount.DecoCards `json:"data"`
	}
	if err := s.httpMgr.Get(c, s.decorateCards, "", params, &ret); err != nil {
		log.Errorc(c, "PGCBatch http GET(%s) failed, params:(%s), error(%+v)", s.decorateCards, params.Encode(), err)
		return nil, err
	}
	if ret.Code != 0 {
		log.Errorc(c, "PGCBatch http GET(%s) failed, params:(%s), code: %v, msg: %v", s.decorateCards, params.Encode(), ret.Code, ret.Msg)
		err := errors.Wrapf(ecode.Int(ret.Code), "PGCBatch url(%v) code(%v) msg(%v)", s.decorateCards, ret.Code, ret.Msg)
		return nil, err
	}
	return ret.Data, nil
}

func (s *Service) cards3New(c context.Context, uids []int64) (map[int64]*accountapi.Card, error) {
	var max50 = 50
	g := errgroup.WithContext(c)
	mu := sync.Mutex{}
	res := make(map[int64]*accountapi.Card)
	for i := 0; i < len(uids); i += max50 {
		var partUids []int64
		if i+max50 > len(uids) {
			partUids = uids[i:]
		} else {
			partUids = uids[i : i+max50]
		}
		g.Go(func(ctx context.Context) (err error) {
			cs, err := s.cards3Slice(ctx, partUids)
			if err != nil {
				return err
			}
			mu.Lock()
			for uid, card := range cs {
				res[uid] = card
			}
			mu.Unlock()
			return
		})
	}
	if err := g.Wait(); err != nil {
		log.Error("Cards3 uids(%+v) eg.wait(%+v)", uids, err)
		return nil, err
	}
	return res, nil
}

func (s *Service) cards3Slice(c context.Context, uids []int64) (map[int64]*accountapi.Card, error) {
	cardReply, err := s.accGRPC.Cards3(c, &accountapi.MidsReq{Mids: uids})
	if err != nil || cardReply == nil {
		log.Error("Failed to call Cards3(). uids: %+v. error: %+v", uids, errors.WithStack(err))
		return nil, err
	}
	return cardReply.GetCards(), nil
}

func (s *Service) isAttention(c context.Context, owners []int64, mid int64) (isAtten map[int64]int32) {
	var max50 = 50
	g := errgroup.WithContext(c)
	mu := sync.Mutex{}
	res := make(map[int64]int32)
	for i := 0; i < len(owners); i += max50 {
		var partUids []int64
		if i+max50 > len(owners) {
			partUids = owners[i:]
		} else {
			partUids = owners[i : i+max50]
		}
		g.Go(func(ctx context.Context) (err error) {
			as := s.isAttentionSlice(ctx, partUids, mid)
			mu.Lock()
			for uid, a := range as {
				res[uid] = a
			}
			mu.Unlock()
			return
		})
	}
	if err := g.Wait(); err != nil {
		log.Error("isAttention owners(%+v) eg.wait(%+v)", owners, err)
		return nil
	}
	return res
}

func (s *Service) isAttentionSlice(c context.Context, owners []int64, mid int64) (isAtten map[int64]int32) {
	if len(owners) == 0 || mid == 0 {
		return
	}
	ip := metadata.String(c, metadata.RemoteIP)
	arg := &accountapi.RelationsReq{Owners: owners, Mid: mid, RealIp: ip}
	res, err := s.accGRPC.Relations3(c, arg)
	if err != nil {
		log.Error("s.accGRPC.Relations3 arg=%+v, err=%+v", arg, err)
		return
	}
	isAtten = make(map[int64]int32, len(res.Relations))
	for mid, rel := range res.Relations {
		if rel.Following {
			isAtten[mid] = 1
		}
	}
	return
}

func (s *Service) stats(c context.Context, uids []int64) (map[int64]*relationgrpc.StatReply, error) {
	var max50 = 50
	g := errgroup.WithContext(c)
	mu := sync.Mutex{}
	res := make(map[int64]*relationgrpc.StatReply)
	for i := 0; i < len(uids); i += max50 {
		var partUids []int64
		if i+max50 > len(uids) {
			partUids = uids[i:]
		} else {
			partUids = uids[i : i+max50]
		}
		g.Go(func(ctx context.Context) (err error) {
			ss, err := s.statsSlice(ctx, partUids)
			if err != nil {
				return err
			}
			mu.Lock()
			for uid, s := range ss {
				res[uid] = s
			}
			mu.Unlock()
			return
		})
	}
	if err := g.Wait(); err != nil {
		log.Error("stats uids(%+v) eg.wait(%+v)", uids, err)
		return nil, err
	}
	return res, nil
}

func (s *Service) statsSlice(ctx context.Context, mids []int64) (map[int64]*relationgrpc.StatReply, error) {
	var (
		arg        = &relationgrpc.MidsReq{Mids: mids}
		statsReply *relationgrpc.StatsReply
		err        error
	)
	if statsReply, err = s.relGRPC.Stats(ctx, arg); err != nil {
		err = errors.Wrapf(err, "%v", arg)
		return nil, err
	}
	return statsReply.StatReplyMap, nil
}

func (s *Service) interrelations(ctx context.Context, mid int64, owners []int64) (map[int64]*relationgrpc.InterrelationReply, error) {
	fidsMap := make(map[int64]int64)
	var fids []int64
	for _, fid := range owners {
		if _, ok := fidsMap[fid]; ok {
			continue
		}
		fidsMap[fid] = fid
		fids = append(fids, fid)
	}
	const _fidMax = 20
	g := errgroup.WithContext(ctx)
	mu := sync.Mutex{}
	res := make(map[int64]*relationgrpc.InterrelationReply)
	for i := 0; i < len(fids); i += _fidMax {
		var partFids []int64
		if i+_fidMax > len(fids) {
			partFids = fids[i:]
		} else {
			partFids = fids[i : i+_fidMax]
		}
		g.Go(func(ctx context.Context) (err error) {
			var (
				reply *relationgrpc.InterrelationMapReply
				arg   = &relationgrpc.RelationsReq{
					Mid: mid,
					Fid: partFids,
				}
			)
			if reply, err = s.relGRPC.Interrelations(ctx, arg); err != nil {
				log.Error("d.relGRPC.interrelations(%v) error(%v)", arg, err)
				return nil
			}
			if reply == nil {
				return nil
			}
			mu.Lock()
			for k, v := range reply.InterrelationMap {
				res[k] = v
			}
			mu.Unlock()
			return nil
		})
	}
	if err := g.Wait(); err != nil {
		log.Error("interrelations mid(%d) eg.wait(%+v)", mid, err)
		return nil, err
	}
	return res, nil
}

func (s *Service) playOnline(ctx context.Context, aidm map[int64]int64) (map[int64]*playurlgrpc.PlayOnlineReply, error) {
	g := errgroup.WithContext(ctx)
	mu := sync.Mutex{}
	res := map[int64]*playurlgrpc.PlayOnlineReply{}
	for k, v := range aidm {
		aid := k
		cid := v
		g.Go(func(ctx context.Context) error {
			req := &playurlgrpc.PlayOnlineReq{
				Aid:      aid,
				Cid:      cid,
				Business: playurlgrpc.OnlineBusiness_OnlineUGC,
			}
			reply, err := s.playurlGRPC.PlayOnline(ctx, req)
			if err != nil {
				return err
			}
			mu.Lock()
			res[aid] = reply
			mu.Unlock()
			return nil
		})
	}
	if err := g.Wait(); err != nil {
		return nil, err
	}
	return res, nil
}
