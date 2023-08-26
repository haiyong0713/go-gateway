package lottery

import (
	"context"

	"git.bilibili.co/bapis/bapis-go/account/service"
	"go-common/library/log"
	"go-gateway/app/web-svr/activity/admin/model/lottery"
)

const _wxActName = "小程序抽奖"

func (s *Service) WxLotteryLog(c context.Context, mid, giftType int64, pn, ps int64) (list []*lottery.WxLotteryLog, count int64, err error) {
	list, count, err = s.lotDao.WxLotteryLog(c, mid, giftType, pn, ps)
	if err != nil {
		log.Error("WxLotteryLog mid:%d giftType:%d pn:%d ps:%v error:%v", mid, giftType, pn, ps, err)
		return
	}
	if len(list) > 0 {
		var mids []int64
		for _, v := range list {
			if v != nil && v.Mid > 0 {
				mids = append(mids, v.Mid)
			}
		}
		var accs map[int64]*api.Info
		if len(mids) > 0 {
			if arcReply, err := s.accClient.Infos3(c, &api.MidsReq{Mids: mids}); err != nil {
				log.Error("WxLotteryLog Infos3 mids:%v error:%v", mids, err)
				err = nil
			} else if arcReply != nil {
				accs = arcReply.Infos
			}
		}
		for _, v := range list {
			if v == nil {
				continue
			}
			v.ActName = _wxActName
			if acc, ok := accs[v.Mid]; ok && acc != nil {
				v.Uname = acc.Name
			}
			if v.GiftID > 0 {
				v.GiftCount = 1
			}
			if v.GiftType == 2 || v.GiftType == 3 { //大会员券，头像挂件
				v.GiftStatus = 1
			} else if v.GiftType == 4 { // 现金奖
				v.GiftStatus = 2        // 审核中
				if v.OrderStatus == 2 { // 现金已发放
					v.GiftStatus = 1
				}
			}
		}
	}
	return
}
