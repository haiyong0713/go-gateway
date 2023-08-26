package lottery

import (
	"context"

	accapi "git.bilibili.co/bapis/bapis-go/account/service"
	"go-common/library/cache"
	"go-common/library/log"
	"go-common/library/sync/errgroup.v2"
	"go-gateway/app/web-svr/activity/ecode"
	l "go-gateway/app/web-svr/activity/interface/model/lottery_v2"

	"github.com/pkg/errors"
)

// WinList 中奖名单接口
func (s *Service) WinList(c context.Context, sid string, num int64, needCache bool) (res []*l.WinList, err error) {
	log.Infoc(c, "do new lottery")
	var (
		lottery     *l.Lottery
		lotteryGift []*l.Gift
		giftList    []*l.GiftMid
		membersRly  *accapi.InfosReply
		mids        []int64
	)
	eg := errgroup.WithContext(c)
	eg.Go(func(ctx context.Context) (err error) {
		if lottery, err = s.base(ctx, sid); err != nil {
			err = errors.Wrapf(err, "s.dao.Lottery %s", sid)
		}
		return
	})
	eg.Go(func(ctx context.Context) (err error) {
		if lotteryGift, err = s.gift(ctx, sid); err != nil {
			err = errors.Wrapf(err, "s.dao.LotteryGift %s", sid)
		}
		return
	})
	if err = eg.Wait(); err != nil {
		log.Errorc(c, "eg.Wait error(%v)", err)
		return
	}
	if lottery.ID == 0 {
		err = ecode.ActivityNotExist
		return
	}
	if len(lotteryGift) == 0 {
		err = ecode.ActivityNotConfig
		return
	}
	showGift := make([]int64, 0)
	showGiftMap := make(map[int64]*l.Gift)
	for _, v := range lotteryGift {
		if v.IsShow == l.IsShow {
			showGiftMap[v.ID] = v
			showGift = append(showGift, v.ID)
		}

	}
	if giftList, err = s.winList(c, lottery.ID, showGift, num, needCache); err != nil {
		log.Errorc(c, "WinList s.dao.LotteryWinList sid(%d) num(%d) error(%v)", lottery.ID, num, err)
		return
	}
	mids = make([]int64, 0, len(giftList))
	for _, v := range giftList {
		if v.GiftID == -1 {
			continue
		}
		mids = append(mids, v.Mid)
	}
	if len(mids) == 0 {
		return
	}
	if membersRly, err = s.accClient.Infos3(c, &accapi.MidsReq{Mids: mids}); err != nil {
		log.Errorc(c, "s.accRPC.Infos3(%v) error(%v)", mids, err)
		return
	}
	res = make([]*l.WinList, 0, len(giftList))
	for _, v := range giftList {
		if v.GiftID == -1 {
			continue
		}
		gift, ok := showGiftMap[v.GiftID]
		if !ok {
			continue
		}
		v.GiftName = gift.Name
		v.ImgURL = gift.ImgURL
		n := &l.WinList{GiftMid: v}
		if membersRly != nil {
			if val, y := membersRly.Infos[v.Mid]; y {
				n.Name = hideName(val.Name)
				n.Mid = 0
			}
		}
		res = append(res, n)
	}
	return
}

func hideName(name string) (res string) {
	if name == "" {
		return "***"
	}
	tmp := []rune(name)
	l := len(tmp)
	if l <= 3 {
		return "***" + name
	}
	for i := 0; i < 3; i++ {
		tmp[i] = '*'
	}
	res = string(tmp)
	return
}

// winList ...

func (s *Service) winList(c context.Context, sid int64, giftIds []int64, num int64, needCache bool) (res []*l.GiftMid, err error) {
	if needCache {
		res, err = s.lottery.CacheLotteryWinList(c, sid)
		if err != nil {
			log.Errorc(c, " s.lottery.CacheLotteryWinList(%d) err(%v)", sid, err)
		}
		if len(res) != 0 {
			cache.MetricHits.Inc("LotteryWinList")
			return res, nil
		}
	}
	cache.MetricMisses.Inc("LotteryWinList")
	res, err = s.lottery.RawLotteryWinList(c, sid, giftIds, num)
	if err != nil {
		return nil, err
	}
	miss := res
	if len(res) == 0 {
		miss = []*l.GiftMid{{GiftID: -1}}
	}
	s.cache.Do(c, func(c context.Context) {
		s.lottery.AddCacheLotteryWinList(c, sid, miss)
	})
	return res, nil
}

// myRealyWinList ...
func (s *Service) myRealyWinList(c context.Context, id, mid int64, pn, ps int) (res []*l.MidWinList, err error) {
	var (
		start = int64((pn - 1) * ps)
		end   = start + int64(ps) - 1
	)
	addCache := true
	res, err = s.lottery.CacheLotteryWinLog(c, id, mid, start, end)

	if err != nil {
		addCache = false
		err = nil
	}
	if len(res) != 0 {
		cache.MetricHits.Inc("LotteryMyRealyWinList")
		return
	}
	cache.MetricMisses.Inc("LotteryMyRealyWinList")
	res, err = s.lottery.RawLotteryMidWinList(c, id, mid, start, int64(ps))
	if err != nil {
		return
	}
	miss := res
	if len(res) == 0 {
		return
	}
	if !addCache {
		return
	}
	s.cache.Do(c, func(ctx context.Context) {
		s.lottery.AddCacheLotteryWinLog(ctx, id, mid, miss)
	})
	return
}

// CouponWinList 中奖名单接口
func (s *Service) CouponWinList(c context.Context, sid string, mid, giftID int64, pn, ps int) (res *l.RealWinList, err error) {
	log.Infoc(c, "do new lottery")
	res = &l.RealWinList{}

	var (
		lottery     *l.Lottery
		lotteryGift []*l.Gift
		giftList    []*l.MidWinList
	)
	eg := errgroup.WithContext(c)
	eg.Go(func(ctx context.Context) (err error) {
		if lottery, err = s.base(ctx, sid); err != nil {
			err = errors.Wrapf(err, "s.dao.Lottery %s", sid)
		}
		return
	})
	eg.Go(func(ctx context.Context) (err error) {
		if lotteryGift, err = s.gift(ctx, sid); err != nil {
			err = errors.Wrapf(err, "s.dao.LotteryGift %s", sid)
		}
		return
	})
	if err = eg.Wait(); err != nil {
		log.Errorc(c, "eg.Wait error(%v)", err)
		return
	}
	if lottery.ID == 0 {
		err = ecode.ActivityNotExist
		return
	}
	if len(lotteryGift) == 0 {
		err = ecode.ActivityNotConfig
		return
	}
	var canFindGift bool
	for _, v := range lotteryGift {
		if v.ID == giftID && v.Type == giftCoupon {
			canFindGift = true
			continue
		}
	}
	if !canFindGift {
		err = ecode.ActivityLotteryGiftErr
		return
	}
	if giftList, err = s.myRealyWinList(c, lottery.ID, mid, pn, ps); err != nil {
		log.Errorc(c, "WinList s.dao.myRealyWinList sid(%d) mid(%d) giftID (%d) error(%v)", lottery.ID, mid, giftID, err)
		return
	}
	list := make([]*l.MidWinList, 0)
	for _, v := range giftList {
		if v.GiftID == giftID {
			list = append(list, v)
		}
	}
	res.List = list
	return
}
