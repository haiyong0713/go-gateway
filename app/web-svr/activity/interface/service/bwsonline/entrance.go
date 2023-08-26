package bwsonline

import (
	"context"

	"go-common/library/log"
	"go-common/library/sync/errgroup.v2"
	"go-gateway/app/web-svr/activity/ecode"
	dao "go-gateway/app/web-svr/activity/interface/dao/bwsonline"
	bwsmdl "go-gateway/app/web-svr/activity/interface/model/bws"
	"go-gateway/app/web-svr/activity/interface/model/like"
)

func (s *Service) tabResData(status int, err error) (interface{}, error) {
	return struct {
		Status int `json:"status"`
	}{
		Status: status,
	}, err
}

func (s *Service) midToKey(c context.Context, bid, mid int64) (key string, err error) {
	var users *bwsmdl.Users
	if users, err = s.bwsdao.UsersMid(c, bid, mid); err != nil {
		err = ecode.ActivityKeyFail
		return
	}
	if users == nil || users.Key == "" {
		err = ecode.ActivityNotBind
		return
	}
	key = users.Key
	return
}

func (s *Service) TabEntrance(ctx context.Context, mid int64) (interface{}, error) {
	// 统一cache判断
	status, err := s.dao.CacheUserEntrance(ctx, mid)
	if err != nil {
		log.Errorc(ctx, "TabEntrance s.dao.CacheUserEntrance(ctx, %d) err[%v]", mid, err)
		return s.tabResData(0, err)
	}
	switch status {
	case dao.EntranceStatusOpened:
		{
			return s.tabResData(1, err)
		}
	case dao.EntranceStatusClosed:
		{
			return s.tabResData(0, err)
		}
	}

	// 并发判断4个人群集合条件
	var currentCurrency, prevCurrency map[int64]int64
	var reserved, buyTicket *like.HasReserve
	var key string
	eg := errgroup.WithContext(ctx)
	eg.Go(func(ctx context.Context) error {
		// 广州场是否玩过判断
		var err error
		currentCurrency, err = s.dao.UserCurrency(ctx, mid, s.c.BwsOnline.DefaultBid)
		if err != nil {
			log.Errorc(ctx, "TabEntrance s.dao.UserCurrency(ctx, %d, %d) err[%v]", mid, s.c.BwsOnline.DefaultBid, err)
		}
		return err
	})
	eg.Go(func(ctx context.Context) error {
		// 上海场是否玩过判断
		var err error
		prevCurrency, err = s.dao.UserCurrency(ctx, mid, s.c.BwsOnline.PrevBid)
		if err != nil {
			log.Errorc(ctx, "TabEntrance s.dao.UserCurrency(ctx, %d, %d) err[%v]", mid, s.c.BwsOnline.PrevBid, err)
		}
		return err
	})
	eg.Go(func(ctx context.Context) error {
		// 预约过本次BW2020广州场
		var err error
		reserved, err = s.likeDao.ReserveOnly(ctx, s.c.BwsOnline.ReserveSid, mid)
		if err != nil {
			log.Errorc(ctx, "TabEntrance s.likeDao.ReserveOnly(ctx, %d, %d) err[%v]", s.c.BwsOnline.ReserveSid, mid, err)
		}
		return err
	})
	eg.Go(func(ctx context.Context) error {
		// 为本届BW2020 广州场 买票
		var err error
		buyTicket, err = s.likeDao.ReserveOnly(ctx, s.c.BwsOnline.BuyTicketSid, mid)
		if err != nil {
			log.Errorc(ctx, "TabEntrance s.likeDao.ReserveOnly(ctx, %d, %d) err[%v]", s.c.BwsOnline.BuyTicketSid, mid, err)
		}
		return err
	})
	eg.Go(func(ctx context.Context) error {
		var err error
		var users *bwsmdl.Users
		if users, err = s.bwsdao.UsersMid(ctx, s.c.Bws.Bws202012Bid, mid); err != nil {
			err = ecode.ActivityKeyFail
			return err
		}
		if users == nil || users.Key == "" {
			key = ""
			return nil
		}
		key = users.Key
		return nil
	})

	if err := eg.Wait(); err != nil {
		return s.tabResData(0, err)
	}

	// 设置缓存
	if len(currentCurrency) > 0 ||
		len(prevCurrency) > 0 ||
		(reserved != nil && reserved.State == 1) ||
		(key != "") ||
		(buyTicket != nil && buyTicket.State == 1) {
		s.dao.AddCacheUserEntrance(ctx, mid, dao.EntranceStatusOpened)
		return s.tabResData(1, err)
	}
	s.dao.AddCacheUserEntrance(ctx, mid, dao.EntranceStatusClosed)
	return s.tabResData(0, err)
}
