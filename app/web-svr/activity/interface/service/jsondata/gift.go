package jsondata

import (
	"context"
	"go-common/library/log"
	mdl "go-gateway/app/web-svr/activity/interface/model/jsondata"
	"strings"
	"time"
)

func today() string {
	return time.Now().Format("2006-01-02")
}

// initGift
func (s *Service) initSummerGift(ctx context.Context) (err error) {
	summerGift, err := s.dao.GetSummerGift(ctx, time.Now().Unix())
	if err != nil {
		log.Errorc(ctx, "s.dao.GetSummerGift err(%v)", err)
		return err
	}
	today := today()
	summerGiftNew := make([]*mdl.SummerGift, 0)
	giftIdAlready := make(map[int64]struct{})
	if len(summerGift) > 0 {
		for _, v := range summerGift {
			if v.Date != "" {
				date := strings.Split(v.Date, ",")
				if len(date) > 0 {
					for _, day := range date {
						if day == today {
							if _, ok := giftIdAlready[v.ID]; ok {
								continue
							}
							summerGiftNew = append(summerGiftNew, &mdl.SummerGift{
								ID:     v.ID,
								Name:   v.Name,
								ImgUrl: v.ImgUrl,
								Order:  v.Order,
							})
							giftIdAlready[v.ID] = struct{}{}
						}
					}
				}
			}
		}
	}
	s.summerGift = summerGiftNew
	return nil
}

// GetSummerGift 夏日活动奖品
func (s *Service) GetSummerGift(ctx context.Context) []*mdl.SummerGift {
	return s.summerGift
}
