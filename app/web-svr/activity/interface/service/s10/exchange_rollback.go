package s10

import (
	"context"
	"fmt"

	"go-gateway/app/web-svr/activity/interface/model/s10"

	"go-common/library/log"
	"go-common/library/stat/prom"
	xtime "go-common/library/time"
)

func (s *Service) correctAllGoodsStock(ctx context.Context, gift *s10.Bonus, id, mid int64, typ string, currdate xtime.Time) {
	switch typ {
	case "cost":
		if s.splitTab {
			if _, err := s.dao.UpdateUserCostRecordStateSub(ctx, id, mid); err != nil {
				str := fmt.Sprintf("gid:%d need:[cost,act_round_goods,act_goods,decr_round_goods,decr_goods] unfinish: index of cost", gift.ID)
				prom.BusinessErrCount.Incr(str)
				log.Errorc(ctx, str)
				return
			}
		} else {
			if _, err := s.dao.UpdateUserCostRecordState(ctx, id); err != nil {
				str := fmt.Sprintf("gid:%d need:[cost,act_round_goods,act_goods,decr_round_goods,decr_goods] unfinish: index of cost", gift.ID)
				prom.BusinessErrCount.Incr(str)
				log.Errorc(ctx, str)
				return
			}
		}
		fallthrough
	case "act_round_goods":
		if !gift.IsRoundInfinite && gift.IsRound && currdate != 0 {
			if _, err := s.dao.CorrectGoodsRoundSendCount(ctx, gift.ID, currdate); err != nil {
				str := fmt.Sprintf("gid:%d  need:[cost,act_round_goods,act_goods,decr_round_goods,decr_goods] unfinish: index of act_round_goods", gift.ID)
				prom.BusinessErrCount.Incr(str)
				log.Errorc(ctx, str)
				return
			}
		}
		fallthrough
	case "act_goods":
		if !gift.IsInfinite {
			if _, err := s.dao.CorrectGoodsSendCount(ctx, gift.ID); err != nil {
				str := fmt.Sprintf("gid:%d need:[cost,act_round_goods,act_goods,decr_round_goods,decr_goods] unfinish: index of act_goods", gift.ID)
				prom.BusinessErrCount.Incr(str)
				log.Errorc(ctx, str)
				return
			}
		}
		fallthrough
	case "decr_round_goods":
		if !gift.IsRoundInfinite && gift.IsRound && currdate != 0 {
			if _, err := s.dao.DecrRoundRestCountByGoodsCache(ctx, gift.ID, currdate); err != nil {
				str := fmt.Sprintf("gid:%d need:[cost,act_round_goods,act_goods,decr_round_goods,decr_goods] unfinish: index of decr_round_goods", gift.ID)
				prom.BusinessErrCount.Incr(str)
				log.Errorc(ctx, str)
				return
			}
		}
		fallthrough
	case "decr_goods":
		if !gift.IsInfinite {
			if _, err := s.dao.DecrRestCountByGoodsCache(ctx, gift.ID); err != nil {
				str := fmt.Sprintf("gid:%d need:[cost,act_round_goods,act_goods,decr_round_goods,decr_goods] unfinish: index of decr_goods", gift.ID)
				prom.BusinessErrCount.Incr(str)
				log.Errorc(ctx, str)
				return
			}
		}

	}
	return
}

func (s *Service) correctReceive(ctx context.Context, mid int64, robin int32) error {
	if _, err := s.dao.CorrectUserLotteryByRobin(ctx, mid, robin); err != nil {
		str := fmt.Sprintf("user receive goods:mid:%d,robin:%d", mid, robin)
		log.Errorc(ctx, str)
		prom.BusinessInfoCount.Incr(str)
	}
	return nil
}
