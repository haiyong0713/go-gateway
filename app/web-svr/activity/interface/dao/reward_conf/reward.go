package reward_conf

import (
	"context"
	"go-common/library/cache"
	"go-common/library/log"
	xtime "go-common/library/time"
	"go-gateway/app/web-svr/activity/interface/api"
	"go-gateway/app/web-svr/activity/interface/client"
	cmdl "go-gateway/app/web-svr/activity/interface/model/cost"
	"time"
)

const (
	COST_TYPE_LOTTERY  = 1 // 抽奖
	COST_TYPE_EXCHANGE = 2 // 积分兑换
)

// GetTodayAwardList 查询今日可兑换奖品列表
func (d *dao) GetTodayAwardList(ctx context.Context, activityId string, costType int) (res []*cmdl.AwardConfigDataDB, err error) {
	timeStr := time.Now().Format("20060102")
	addCache := true
	res, err = d.CacheAwardList(ctx, activityId, timeStr)
	if err != nil {
		addCache = false
		err = nil
	}
	if len(res) != 0 {
		cache.MetricHits.Inc("bts:AwardList")
		return
	}
	cache.MetricMisses.Inc("bts:AwardList")
	res, err = d.FetchAwardFromDB(ctx, activityId, costType, xtime.Time(time.Now().Unix()))

	if err != nil {
		return
	}
	miss := res
	if !addCache {
		return
	}
	d.AddCacheAwardList(ctx, activityId, timeStr, miss)
	return

}

// 查询兑换奖品库存是否充足
func (d *dao) IsAwardCanExchange(ctx context.Context, activityId string, awardId string, mid int64) (hasStock bool, res *cmdl.AwardConfigDataDB, err error) {
	// 查询兑换奖品库存&需要消耗
	res, err = d.GetAwardConfByIdAndDate(ctx, activityId, awardId, COST_TYPE_EXCHANGE, xtime.Time(time.Now().Unix()))
	if err != nil {
		log.Errorc(ctx, "IsAwardCanExchange err,err is :(%v).", err)
		return false, nil, err
	}
	if res != nil {
		// 查库存
		//stockRes, err := d.stockDao.GetGiftStocks(ctx, res.ActivityId, []int64{awardIdInt})
		stockRes, err := client.ActivityClient.GetStocksByIds(ctx, &api.GetStocksReq{
			StockIds:  []int64{res.StockId},
			SkipCache: false,
			Mid:       mid,
		})

		if err != nil || stockRes == nil {
			log.Errorc(ctx, "IsAwardCanExchange client.ActivityClient.GetStocksByIds err,err is (%v).", err)
			return false, nil, err
		}
		if stockRes.StockMap != nil {
			if mp, ok := stockRes.StockMap[res.StockId]; ok {
				if len(mp.List) <= 0 {
					return false, res, nil
				} else {
					if mp.List[0].StockNum > 0 {
						// 库存充足
						return true, res, nil
					}
				}
			} else {
				// 库存不存在
				return false, res, nil
			}
		}
		// 库存nil 不足
		return false, res, nil
	}
	return false, nil, nil
}

// IsAwardCanLottery
func (d *dao) IsAwardCanLottery(ctx context.Context, activityId string, awardId string) (canLottery bool, res *cmdl.AwardConfigDataDB, err error) {
	//todayZero := time.Date(time.Now().Year(), time.Now().Month(), time.Now().Day(), 00, 0, 0, 0, time.Local).Unix()
	res, err = d.GetAwardConfByIdAndDate(ctx, activityId, awardId, COST_TYPE_LOTTERY, xtime.Time(time.Now().Unix()))
	if err != nil || res == nil {
		log.Errorc(ctx, "IsAwardCanLottery err or res is nil,err is :(%v).", err)
		return false, nil, err
	}
	if res.ID <= 0 {
		log.Errorc(ctx, "IsAwardCanLottery res is nil.")
		return false, nil, nil
	}
	return true, res, nil
}
