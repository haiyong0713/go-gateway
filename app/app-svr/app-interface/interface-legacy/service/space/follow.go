package space

import (
	"context"
	"strconv"

	"go-common/library/log"

	xecode "go-gateway/app/app-svr/app-card/ecode"
	cdm "go-gateway/app/app-svr/app-card/interface/model"
	"go-gateway/app/app-svr/app-card/interface/model/card"
	"go-gateway/app/app-svr/app-card/interface/model/card/operate"
	"go-gateway/app/app-svr/app-interface/interface-legacy/model"

	account "git.bilibili.co/bapis/bapis-go/account/service"
	relationgrpc "git.bilibili.co/bapis/bapis-go/account/service/relation"

	"go-common/library/sync/errgroup.v2"
)

// searchFollow picks upper recommend data from search API and then build the operation card to return
func (s *Service) searchFollow(c context.Context, platform, mobiApp, device, buvid string, build int, mid, vmid int64) (follow *operate.Card, err error) {
	const _title = "关注TA的也关注了"
	ups, trackID, err := s.srchDao.Follow(c, platform, mobiApp, device, buvid, build, mid, vmid)
	if err != nil {
		return
	}
	items := make([]*operate.Card, 0, len(ups))
	for _, up := range ups {
		if up.Mid != 0 {
			item := &operate.Card{ID: up.Mid, Goto: cdm.GotoMid, Param: strconv.FormatInt(up.Mid, 10), URI: strconv.FormatInt(up.Mid, 10), Desc: up.RecReason}
			items = append(items, item)
		}
	}
	//nolint:gomnd
	if len(items) < 3 {
		return
	}
	id, _ := strconv.ParseInt(trackID, 10, 64)
	if id < 1 {
		return
	}
	follow = &operate.Card{ID: id, Param: trackID, Items: items, Title: _title, CardGoto: cdm.CardGotoSearchUpper}
	return
}

// UpperRecmd picks upper recommend data from search API and then combine data got from RPC and return the final card structure
func (s *Service) UpperRecmd(c context.Context, plat int8, platform, mobiApp, device, buvid string, build int, mid, vimd int64) (res card.Handler, err error) {
	var (
		upIDs     []int64
		follow    *operate.Card
		cardm     map[int64]*account.Card
		statm     map[int64]*relationgrpc.StatReply
		isAtten   map[int64]int8
		relationm map[int64]*relationgrpc.InterrelationReply
	)
	if follow, err = s.searchFollow(c, platform, mobiApp, device, buvid, build, mid, vimd); err != nil {
		log.Error("%+v", err)
		return
	}
	if follow == nil {
		err = xecode.AppNotData
		log.Error("follow is nil")
		return
	}
	for _, item := range follow.Items {
		upIDs = append(upIDs, item.ID)
	}
	g := errgroup.WithCancel(c)
	if len(upIDs) != 0 {
		g.Go(func(ctx context.Context) (err error) {
			if cardm, err = s.accDao.Cards3(ctx, upIDs); err != nil {
				log.Error("%+v", err)
				err = nil
			}
			return
		})
		g.Go(func(ctx context.Context) (err error) {
			if statm, err = s.relDao.StatsGRPC(ctx, upIDs); err != nil {
				log.Error("%+v", err)
				err = nil
			}
			return
		})
		if mid != 0 {
			g.Go(func(ctx context.Context) error {
				isAtten = make(map[int64]int8)
				follow := s.accDao.Relations3(ctx, upIDs, mid)
				for mid, v := range follow {
					if v {
						isAtten[mid] = 1
					}
				}
				return nil
			})
			g.Go(func(ctx context.Context) error {
				if relationm, err = s.relDao.Interrelations(ctx, mid, upIDs); err != nil {
					log.Error("%v", err)
				}
				return nil
			})
		}
	}
	//nolint:errcheck
	g.Wait()
	op := &operate.Card{}
	op.From(cdm.CardGt(model.GotoSearchUpper), 0, 0, plat, build, mobiApp)
	h := card.Handle(plat, cdm.CardGt(model.GotoSearchUpper), "", cdm.ColumnSvrSingle, nil, nil, isAtten, nil, statm, cardm, relationm)
	if h == nil {
		err = xecode.AppNotData
		return
	}
	op = follow
	//nolint:errcheck
	h.From(nil, op)
	if h.Get().Right {
		res = h
	} else {
		err = xecode.AppNotData
	}
	return
}
