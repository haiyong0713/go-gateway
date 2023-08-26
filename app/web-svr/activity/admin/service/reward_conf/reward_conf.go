package reward_conf

import (
	"context"
	"encoding/json"
	"fmt"
	"go-common/library/log"
	xtime "go-common/library/time"
	"go-gateway/app/web-svr/activity/admin/client"
	reward_conf2 "go-gateway/app/web-svr/activity/admin/dao/reward_conf"
	"go-gateway/app/web-svr/activity/admin/model/reward_conf"
	"go-gateway/app/web-svr/activity/ecode"
	"go-gateway/app/web-svr/activity/interface/api"
	"strconv"
	"strings"
	"time"
)

const (
	ActivityCycleTYpe      = 1
	HasStock          int  = 1
	CostTypeLottery   int8 = 1
	CostTypeExchange  int8 = 2
)

// AddOneRewardConf service层添加一条奖品配置
func (s *Service) AddOneRewardConf(ctx context.Context, req *reward_conf.AddOneRewardReq, username string) (err error) {
	var stockId int32 = 0
	if req.CostType == CostTypeExchange {
		if req.StoreNum <= 0 {
			err = ecode.StockNumErr
			return err
		}
		cycleLimit := []reward_conf.StockCycleLimit{
			{
				CycleType: ActivityCycleTYpe,
				LimitType: HasStock,
				Store:     req.StoreNum,
			},
		}
		b, err := json.Marshal(cycleLimit)
		if err != nil {
			log.Errorc(ctx, "AddOneRewardConf json.marshal err,err is (%v).", err)
			return err
		}
		resp, err := client.ActivityClient.CreateStockRecord(ctx, &api.CreateStockRecordReq{
			ResourceId:     req.ActivityId,
			ResourceVer:    time.Now().Unix(),
			ForeignActId:   fmt.Sprintf("{%s}_{%s}_{%v}", req.ActivityId, req.AwardId, req.ShowTime),
			StockStartTime: req.ShowTime,
			StockEndTime:   getNextMonthTime(3),
			CycleLimit:     string(b[:]),
		})
		if err != nil || resp == nil {
			log.Errorc(ctx, "AddOneRewardConf client.ActivityClient.CreateStockRecord err,err is (%v).req is (%v).", err, req)
			err = ecode.StockErr
			return err
		}
		stockId = int32(resp.StockId)
	}
	if req.Creator != "" {
		username = req.Creator
	}
	record := &reward_conf2.AwardConfigData{
		AwardID:    req.AwardId,
		StockID:    stockId,
		CostType:   req.CostType,
		CostValue:  req.CostValue,
		ShowTime:   req.ShowTime,
		Order:      req.Order,
		Creator:    username,
		Status:     1,
		ActivityID: req.ActivityId,
		EndTime:    req.EndTime,
	}
	err = s.dao.AddOneRewardConf(ctx, record)
	if err != nil {
		log.Errorc(ctx, "AddOneRewardConf dao.AddOneRewardConf err,err is (%v).", err)
		return
	}
	return
	// todo 删缓存
}

// UpdateOneRewardConf service层更新一条奖品配置
func (s *Service) UpdateOneRewardConf(ctx context.Context, req *reward_conf.UpdateOneRewardReq, username string) (err error) {
	if req.StockId != 0 && req.CostType == CostTypeExchange && req.StoreNum != 0 {
		cycleLimit := []reward_conf.StockCycleLimit{
			{
				CycleType: ActivityCycleTYpe,
				LimitType: HasStock,
				Store:     req.StoreNum,
			},
		}
		b, err := json.Marshal(cycleLimit)
		if err != nil {
			log.Errorc(ctx, "UpdateOneRewardConf json.marshal err,err is (%v).", err)
			return err
		}
		_, err = client.ActivityClient.UpdateStockRecord(ctx, &api.CreateStockRecordReq{
			StockId:        req.StockId,
			ResourceId:     req.ActivityId,
			ResourceVer:    time.Now().Unix(),
			ForeignActId:   fmt.Sprintf("{%s}_{%s}_{%v}", req.ActivityId, req.AwardId, req.ShowTime),
			StockStartTime: req.ShowTime,
			StockEndTime:   getNextMonthTime(3),
			CycleLimit:     string(b[:]),
		})
	}
	needUpdateMap := make(map[string]interface{}, 0)
	if req.CostType == CostTypeLottery && strings.Compare(req.AwardId, "0") != 0 {
		needUpdateMap["award_id"] = req.AwardId
	}
	if req.CostType != 0 {
		needUpdateMap["cost_type"] = req.CostType
	}
	if req.CostValue != 0 {
		needUpdateMap["cost_value"] = req.CostValue
	}
	if req.ShowTime != 0 {
		needUpdateMap["show_time"] = req.ShowTime
	}
	if req.EndTime != 0 {
		needUpdateMap["end_time"] = req.EndTime
	}
	if req.Order != 0 {
		needUpdateMap["order"] = req.Order
	}
	needUpdateMap["status"] = req.Status
	if req.Creator != "" {
		needUpdateMap["creator"] = req.Creator
	} else {
		needUpdateMap["creator"] = username
	}

	err = s.dao.UpdateOneRewardByID(ctx, req.Id, needUpdateMap)
	if err != nil {
		log.Errorc(ctx, "UpdateOneRewardByID err,err is (%v).", err)
		return
	}

	return
	// todo 删缓存
}

// SearchList 查询列表
func (s *Service) SearchList(ctx context.Context, req *reward_conf.SearchReq) (res *reward_conf.SearchRes, err error) {
	res = new(reward_conf.SearchRes)
	list := make([]*reward_conf2.AwardConfigData, 0)
	list, err = s.dao.Search(ctx, req.ActivityId, req.STime, req.ETime, req.CostType, req.Pn, req.Ps)
	if err != nil {
		log.Errorc(ctx, "SearchList s.dao.Search err,err is (%v).", err)
	}
	if len(list) <= 0 {
		return
	}
	for _, v := range list {
		// 查询奖品详情
		var (
			awardName   string          = ""
			awardIcon   string          = ""
			stockNumMap map[int64]int32 = make(map[int64]int32, 0)
		)
		var stocknum int32 = 0
		if v.CostType == CostTypeExchange {
			awardIdInt, err := strconv.ParseInt(v.AwardID, 10, 64)
			if err != nil {
				log.Errorc(ctx, "SearchList strconv.ParseInt awardid err,"+
					"err is (%v),awardId is (%v).", err, v.AwardID)
				continue
			}
			info, err := client.ActivityClient.RewardsGetAwardConfigById(ctx, &api.RewardsGetAwardConfigByIdReq{
				Id: awardIdInt,
			})
			if err != nil || info == nil {
				log.Errorc(ctx, "SearchList strconv.ParseInt awardid err,"+
					"err is (%v),awardId is (%v).", err, v.AwardID)
				continue
			}
			stock, err := client.ActivityClient.GetStocksByIds(ctx, &api.GetStocksReq{
				StockIds:  []int64{int64(v.StockID)},
				SkipCache: false,
			})
			if err != nil || stock == nil || stock.StockMap == nil {
				log.Errorc(ctx, "SearchList client.ActivityClient.GetStocksByIds err,err is (%v).", err)
				err = nil
			}
			for k, v := range stock.StockMap {
				stockNumMap[k] = v.List[0].StockNum
			}
			awardName = info.Name
			awardIcon = info.IconUrl
			if val, ok := stockNumMap[int64(v.StockID)]; ok {
				stocknum = val
			}
		}

		tmp := &reward_conf.OneAwardRes{
			ID:         v.ID,
			AwardID:    v.AwardID,
			AwardName:  awardName,
			AwardIcon:  awardIcon,
			StockID:    v.StockID,
			StockNum:   stocknum,
			CostType:   v.CostType,
			CostValue:  v.CostValue,
			ShowTime:   v.ShowTime,
			Order:      v.Order,
			Creator:    v.Creator,
			Status:     v.Status,
			Ctime:      v.Ctime,
			Mtime:      v.Mtime,
			ActivityID: v.ActivityID,
			EndTime:    v.EndTime,
		}
		res.List = append(res.List, tmp)
	}
	res.Page = req.Pn
	res.Size = len(list)
	res.Total = len(list)
	return
}

func getNextMonthTime(monthNum int) xtime.Time {
	year, month, _ := time.Now().Date()
	thisMonth := time.Date(year, month, 0, 0, 0, 0, 0, time.Local)
	start := thisMonth.AddDate(0, monthNum, 0).Unix()
	return xtime.Time(start)
}
