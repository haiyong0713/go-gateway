package lottery

import (
	"context"

	"go-common/library/log"

	"go-gateway/app/web-svr/activity/ecode"
	l "go-gateway/app/web-svr/activity/interface/model/lottery_v2"

	"github.com/pkg/errors"
)

// AddLotteryAddress 实物中奖添加地址接口
func (s *Service) AddLotteryAddress(c context.Context, sid string, id, mid int64) (err error) {
	log.Infoc(c, "do new lottery")
	var (
		addrID  int64
		val     *l.AddressInfo
		lottery *l.Lottery
	)
	if lottery, err = s.base(c, sid); err != nil {
		err = errors.Wrapf(err, "s.dao.Lottery %s", sid)
		return
	}
	if addrID, err = s.lotteryAddr(c, lottery.ID, mid); err != nil {
		log.Errorc(c, "AddLotteryAddress s.dao.LotteryAddr mid(%d) error(%v)", mid, err)
		return
	}
	if addrID == id {
		err = ecode.ActivityAddrHasAdd
		return
	}
	if addrID == 0 {
		// 校验传输的地址id是否有效
		if val, err = s.lottery.GetMemberAddress(c, id, mid); err != nil {
			log.Errorc(c, "AddLotteryAddress s.dao.GetMemberAddress id(%d) mid(%d) error(%v)", id, mid, err)
			return
		}
		if val == nil || val.ID == 0 {
			err = ecode.ActivityAddrAddFail
			return
		}
		if _, err = s.lottery.InsertLotteryAddr(c, lottery.ID, mid, id); err != nil {
			log.Errorc(c, "AddLotteryAddress s.dao.InsertLotteryAddr sid(%d) mid(%d) id(%d) error(%v)", lottery.ID, mid, id, err)
			return
		}
		s.cache.Do(c, func(c context.Context) {
			s.lottery.AddCacheLotteryAddrCheck(c, lottery.ID, mid, id)
		})
	}
	return
}

// LotteryAddress ...
func (s *Service) LotteryAddress(c context.Context, sid string, mid int64) (res *l.AddressInfo, err error) {
	log.Infoc(c, "do new lottery")
	var (
		addrID  int64
		lottery *l.Lottery
	)
	if lottery, err = s.base(c, sid); err != nil {
		err = errors.Wrapf(err, "s.dao.Lottery %s", sid)
		return
	}
	if addrID, err = s.lotteryAddr(c, lottery.ID, mid); err != nil {
		log.Errorc(c, "LotteryAddress s.lotteryAddr mid(%d) error(%v)", mid, err)
		return
	}
	if addrID == 0 {
		err = ecode.ActivityAddrNotAdd
		return
	}
	if res, err = s.lottery.GetMemberAddress(c, addrID, mid); err != nil {
		log.Errorc(c, "LotteryAddress s.dao.GetMemberAddress(%d,%d) error(%v)", addrID, mid, err)
	}
	return
}
