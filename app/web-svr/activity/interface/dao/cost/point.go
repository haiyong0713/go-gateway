package cost

import (
	"context"
	"fmt"
	"go-common/library/cache/redis"
	"go-common/library/log"
	"go-common/library/sync/errgroup.v2"
	xtime "go-common/library/time"
	"go-gateway/app/web-svr/activity/ecode"
	"go-gateway/app/web-svr/activity/interface/api"
	"go-gateway/app/web-svr/activity/interface/client"
	cmdl "go-gateway/app/web-svr/activity/interface/model/cost"
	"strings"
	"time"
)

const (
	COST_TYPE_LOTTERY  = 1 // 抽奖
	COST_TYPE_EXCHANGE = 2 // 积分兑换
)

// UserCostForExchange 积分兑换奖品
func (d *dao) UserCostForExchange(ctx context.Context, activityId string, awardId string, mid int64, orderId string) error {
	var (
		totalLeft  int64 = 0
		isExchange       = true
		costValue  int64 = 0
		//awardMtime xtime.Time
		stockId int64
	)

	eg := errgroup.WithContext(ctx)
	// 1.查用户积分余额
	eg.Go(func(ctx context.Context) (err error) {
		totalLeft, _, _, err = d.GetUserTotalPoint(ctx, mid, activityId)
		if err != nil {
			log.Errorc(ctx, "UserCostForExchange err ,err is (%v).", err)
			return ecode.PointNotEnough
		}
		return
	})
	// 2.查用户今天是否兑换过
	eg.Go(func(ctx context.Context) (err error) {
		isExchange, err = d.IsUserExchOrder(ctx, mid, orderId)
		if err != nil || isExchange {
			log.Errorc(ctx, "UserCostForExchange err or today has exchange,err is (%v).isExchange flag is (%v)", err, isExchange)
			return ecode.UserHasExchanged
		}
		return
	})
	// 3.查奖品库存&需要消耗
	eg.Go(func(ctx context.Context) (err error) {
		hasStock, awardDb, err := d.rewardConfDao.IsAwardCanExchange(ctx, activityId, awardId, mid)
		if err != nil || hasStock == false {
			log.Errorc(ctx, "UserCostForExchange err or stock is not enough,err is (%v).flag is (%v)", err, hasStock)
			return ecode.AwardStockErr
		}
		costValue = int64(awardDb.CostValue)
		stockId = awardDb.StockId
		//awardMtime = awardDb.Mtime
		return
	})
	if err := eg.Wait(); err != nil {
		return err
	}
	// check余额是否够兑换
	if totalLeft < costValue {
		log.Errorc(ctx, "UserCostForExchange totalLeft not enouth for costvalue,total is (%d),cost is (%d).", totalLeft, costValue)
		return ecode.PointNotEnough
	}
	// 4.扣减库存
	//awardIdInt, err := strconv.ParseInt(awardId, 10, 64)
	//if err != nil {
	//	log.Errorc(ctx, "UserCostForExchange strconv.ParseInt err,err is (%v).", err)
	//	return ecode.ExchangeErr
	//}
	//_, err = d.stockDao.ConsumerStock(ctx, activityId, awardIdInt, int64(awardMtime), 1)
	nowTime := time.Now().Unix()
	_, err := client.ActivityClient.ConsumerStockById(ctx, &api.ConsumerStockReq{
		StockId: stockId,
		RetryId: fmt.Sprintf("#{%s}-{%s}", mid, nowTime),
		Num:     1,
		Ts:      nowTime,
		Mid:     mid,
	})
	if err != nil {
		log.Errorc(ctx, "UserCostForExchange client.ActivityClient.ConsumerStockById err,err is (%v).", err)
		return ecode.ExchangeErr
	}
	// 5.积分消耗入库
	record := &cmdl.UserCostInfoDB{
		Mid:        mid,
		OrderId:    orderId,
		AwardId:    awardId,
		ActivityId: activityId,
		CostType:   COST_TYPE_EXCHANGE,
		CostValue:  int(costValue),
		Status:     1,
	}
	errI := d.PointReduceOne(ctx, record)
	if errI != nil {
		if !strings.Contains(errI.Error(), "Duplicate entry") {
			// 可重入
			return errI
		}
	}
	// 6.填入今日已兑换flag
	err = d.CacheSetUserExchangeFlag(ctx, orderId, 1)
	if err != nil {
		// 吞掉
		log.Errorc(ctx, "UserCostForExchange d.CacheSetUserExchangeFlag err,err is :(%v).", err)
	}
	return nil
}

// UserCostForLottery 积分抽奖
func (d *dao) UserCostForLottery(ctx context.Context, mid int64, orderId string, awardPoolId string, activityId string) error {
	// 1.查用户积分余额是否充足
	totalLeft, _, _, err := d.GetUserTotalPoint(ctx, mid, activityId)
	if err != nil {
		log.Errorc(ctx, "UserCostForLottery err ,err is (%v).", err)
		return ecode.PointNotEnough
	}
	// 2.查需要消耗
	hasStock, awardDb, err := d.rewardConfDao.IsAwardCanLottery(ctx, activityId, d.c.SummerCampConf.LotteryPoolId)
	if err != nil || hasStock == false {
		log.Errorc(ctx, "UserCostForLottery err or stock is not enough,err is (%v).flag is (%v)", err, hasStock)
		return ecode.AwardStockErr
	}
	if totalLeft < int64(awardDb.CostValue) {
		log.Errorc(ctx, "UserCostForLottery totalLeft not enouth for costvalue,total is (%d),cost is (%d).", totalLeft, awardDb.CostValue)
		return ecode.PointNotEnough
	}
	// 3.积分消耗入库
	record := &cmdl.UserCostInfoDB{
		Mid:        mid,
		OrderId:    orderId,
		AwardId:    awardPoolId,
		ActivityId: activityId,
		CostType:   COST_TYPE_LOTTERY,
		CostValue:  int(awardDb.CostValue),
		Status:     1,
	}
	errI := d.PointReduceOne(ctx, record)
	if errI != nil {
		if !strings.Contains(errI.Error(), "Duplicate entry") {
			// 可重入
			return errI
		}
	}
	return nil
}

// GetUserTotalPoint 获取用户总积分=用户获取-用户消耗
func (d *dao) GetUserTotalPoint(ctx context.Context, mid int64, activityId string) (totalPoint int64, costPoint int64, obtainPoint int64, err error) {
	// 并发 todo
	totalPoint = 0
	costPoint = 0
	obtainPoint = 0

	// 用户获得
	obtainPoint, err = d.TaskFormulaTotal(ctx, mid, activityId)
	if err != nil {
		log.Errorc(ctx, "GetUserTotalPoint get TaskFormulaTotal err ,err is (%v).", err)
		return
	}
	// 用户消耗 1.先读cache
	costPoint, err = d.CacheGetUserCostPoint(ctx, mid, activityId)
	if err != nil {
		log.Errorc(ctx, "GetUserTotalPoint CacheGetUserCostPoint err,err is (%v).", err)
		if err == redis.ErrNil {
			err = nil
		} else {
			return
		}
	}

	// 用户消耗 2.回源
	costPointInt, _, errC := d.GetUserAllCost(ctx, activityId, mid, false)
	if errC != nil {
		log.Errorc(ctx, "GetUserTotalPoint d.GetUserAllCost err,err is (%v).", errC)
		err = errC
		return
	}
	costPoint = int64(costPointInt)

	// 用户消耗 3.写cache
	errS := d.CacheSetUserCostPoint(ctx, mid, activityId, int64(costPoint))
	if errS != nil {
		// 写缓存失败 吞掉
		log.Errorc(ctx, "GetUserTotalPoint CacheSetUserCostPoint err,err is (%v).", errS)
	}

	// 计算返回
	totalPoint = obtainPoint - costPoint
	if totalPoint < 0 {
		log.Errorc(ctx, "GetUserTotalPoint totalPoint less than 0,mid is (%d),activityId is (%s),ObPoint is (%d),costPoint is (%d).",
			mid, activityId, obtainPoint, costPoint)
		totalPoint = 0
		return
	}
	return

}

func (d *dao) IsUserExchOrder(ctx context.Context, mid int64, orderId string) (bool, error) {
	// 读缓存
	flag, err := d.CacheGetUserExchangeFlag(ctx, orderId)
	if err != nil {
		log.Errorc(ctx, "IsUserExchOrder CacheGetUserExchangeFlag err,err is (%v).", err)
		if err == redis.ErrNil {
			err = nil
			flag = 0
		} else {
			return true, err
		}
	}
	if flag == 1 {
		return true, nil
	}

	// 回源 one永远不会为nil
	one, err := d.getUserCostByOrderId(ctx, orderId)
	if err != nil {
		return true, err
	}
	if one.ID == 0 {
		// 写缓存 没兑换过
		errS := d.CacheSetUserExchangeFlag(ctx, orderId, 0)
		if errS != nil {
			// 写缓存失败 吞掉
			log.Errorc(ctx, "IsUserExchOrder CacheSetUserExchangeFlag err,err is (%v).", errS)
		}
		return false, nil
	} else {
		// 写缓存 兑换过
		errS := d.CacheSetUserExchangeFlag(ctx, orderId, 1)
		if errS != nil {
			// 写缓存失败 吞掉
			log.Errorc(ctx, "IsUserExchOrder CacheSetUserExchangeFlag err,err is (%v).", errS)
		}
		return true, nil
	}

}

// PointReduceOne 积分消耗，插入一条记录，并更新数据库
func (d *dao) PointReduceOne(ctx context.Context, record *cmdl.UserCostInfoDB) error {
	// 1.记录入db
	_, err := d.InsertOneUserCost(ctx, record, false)
	if err != nil {
		return err
	}
	// 2.删除缓存（用户消耗总积分）
	err = d.CacheDelUserCostPoint(ctx, record.Mid, record.ActivityId)
	if err != nil {
		// 吞掉
		log.Errorc(ctx, "PointReduceOne d.CacheDelUserCostPoint err,err is :(%v).", err)
	}

	return nil

}

// TodayUserHasExchangedPrizes
func (d *dao) TodayUserHasExchangedPrizes(ctx context.Context, mid int64, activityId string) (list []*cmdl.UserCostInfoDB, err error) {
	// todo 加缓存
	todayZero := time.Date(time.Now().Year(), time.Now().Month(), time.Now().Day(), 00, 0, 0, 0, time.Local).Unix()
	list, err = d.GetUserCostListByDate(ctx, mid, activityId, COST_TYPE_EXCHANGE, xtime.Time(todayZero))
	return

}
