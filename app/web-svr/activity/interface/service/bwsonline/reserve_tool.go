package bwsonline

import (
	"context"
	"fmt"
	"go-common/library/log"
	"go-common/library/sync/errgroup.v2"
	"go-gateway/app/web-svr/activity/ecode"
	pb "go-gateway/app/web-svr/activity/interface/api"
	"go-gateway/app/web-svr/activity/interface/model/bwsonline"
	"go-gateway/app/web-svr/activity/interface/tool"
	"sort"
	"strconv"
	"strings"
	"time"
)

type HygRepData struct {
	Name string                  `json:"name"`
	List []*bwsonline.TicketInfo `json:"list"`
}

// getStockId 生成预约活动库存的业务前缀
func getStockId(id interface{}, vipTag interface{}) string {
	return fmt.Sprintf("%v-%v", id, vipTag)
}

func convert2ReserveRep(lair []*pb.ActInterReserve, ts int64) (res []*ReserveRep) {
	for _, air := range lair {
		rr := &ReserveRep{
			ID:                air.ID,
			ActType:           air.ActType,
			ActTitle:          air.ActTitle,
			ActImg:            air.ActImg,
			ActBeginTime:      air.ActBeginTime,
			ActEndTime:        air.ActEndTime,
			ReserveBeginTime:  air.ReserveBeginTime,
			ReserveEndTime:    air.ReserveEndTime,
			DescribeInfo:      air.DescribeInfo,
			VipTicketNum:      air.VipTicketNum,
			StandardTicketNum: air.StandardTicketNum,
			ScreenDate:        air.ScreenDate,
			DisplayIndex:      air.DisplayIndex,
		}
		if (rr.ReserveBeginTime <= 0 || rr.ReserveEndTime <= 0 || rr.StandardTicketNum <= 0) || (air.VipTicketNum > 0 && air.VipReserveEndTime > 0 && air.VipReserveEndTime > ts) {
			rr.IsVipTicket = 1
			rr.ReserveBeginTime = air.VipReserveBeginTime
			rr.ReserveEndTime = air.VipReserveEndTime
		}
		if rr.ReserveBeginTime <= 0 || rr.ReserveEndTime <= 0 {
			continue
		}
		res = append(res, rr)
	}
	return
}

func (s *Service) getInterReserveState(ctx context.Context, records []*ReserveRep, ts, mid int64,
	ticketMap map[int64]*bwsonline.TicketInfo) (stateMap map[int64]ReserveState, stockMap, vipStockMap map[int64]int, err error) {

	stateMap = make(map[int64]ReserveState)
	stockMap = make(map[int64]int)
	vipStockMap = make(map[int64]int)
	if len(records) <= 0 {
		return
	}
	var orders []*bwsonline.InterReserveOrder
	actIds := []int64{}
	for _, v := range records {
		actIds = append(actIds, v.ID)
	}
	eg := errgroup.WithContext(ctx)
	eg.Go(func(ctx context.Context) error {
		var err1 error
		orders, err1 = s.dao.RawInterReserveOrderByMid(ctx, mid, s.c.BwsOnline.BwPark.Year)
		return err1
	})

	eg.Go(func(ctx context.Context) error {
		var err2 error
		stockMap, err2 = s.dao.GetGiftStocks(ctx, getStockId(s.c.BwsOnline.BwPark.Year, pb.ActInterReserveTicketType_StandardTicket2021), actIds)
		return err2
	})

	eg.Go(func(ctx context.Context) error {
		var err3 error
		vipStockMap, err3 = s.dao.GetGiftStocks(ctx, getStockId(s.c.BwsOnline.BwPark.Year, pb.ActInterReserveTicketType_VipTicket2021), actIds)
		return err3
	})

	if err = eg.Wait(); err != nil {
		return
	}

	for _, v := range orders {
		stateMap[v.InterReserveId] = Reserved
	}
	for _, v := range records {
		if _, ok := stateMap[v.ID]; ok {
			continue
		}
		if v.IsVipTicket == 1 {
			if tick, ok := ticketMap[v.ScreenDate]; ok && tick.SkuName != s.c.BwsOnline.BwPark.VipTag {
				stateMap[v.ID] = OnlyForVip
				continue
			}
		}
		if ts < v.ReserveBeginTime {
			stateMap[v.ID] = Waiting
			continue
		}

		if v.ReserveEndTime < ts {
			stateMap[v.ID] = Closed
			continue
		}

		if tick, ok := ticketMap[v.ScreenDate]; !ok || tick == nil {
			stateMap[v.ID] = Invalid
			continue
		}

		var (
			stock, vipStock int
			ok              bool
		)
		if v.IsVipTicket == 1 {
			if tick, ok := ticketMap[v.ScreenDate]; ok && tick.SkuName != s.c.BwsOnline.BwPark.VipTag {
				stateMap[v.ID] = OnlyForVip
				continue
			}
			if vipStock, ok = vipStockMap[v.ID]; !ok || vipStock <= 0 {
				if tool.InInt64Slice(v.ID, s.c.BwsOnline.BwPark.BackUpStckIds) {
					vipStock, err = s.dao.GetBackUpStock(ctx, getStockId(s.c.BwsOnline.BwPark.Year, pb.ActInterReserveTicketType_VipTicket2021), v.ID)
					if err != nil {
						return
					}
					vipStockMap[v.ID] = vipStock
				}
				if vipStock <= 0 {
					stateMap[v.ID] = SaleOut
					continue
				}
			}
		}
		if vipStock <= 0 {
			if stock, ok = stockMap[v.ID]; !ok || stock <= 0 {
				if tool.InInt64Slice(v.ID, s.c.BwsOnline.BwPark.BackUpStckIds) {
					stock, err = s.dao.GetBackUpStock(ctx, getStockId(s.c.BwsOnline.BwPark.Year, pb.ActInterReserveTicketType_StandardTicket2021), v.ID)
					if err != nil {
						return
					}
					stockMap[v.ID] = stock
				}
				if stock <= 0 {
					stateMap[v.ID] = SaleOut
					continue
				}
			}
		}

		stateMap[v.ID] = Available
	}

	if len(stateMap) < len(records) {
		log.Errorc(ctx, "getInterReserveState mid:%d ,  stateMap len:%d , records len:%d", mid, len(stateMap), len(records))
		err = ecode.SystemActivityConfigErr
	}
	return
}

// getReserveNo 获取预约号码
func (s *Service) getReserveNo(ctx context.Context, sk string, actReserve *pb.ActInterReserve) (orderNo string, reserveNo int, err error) {
	var orderNos []string
	if orderNos, err = s.dao.ConsumerStock(ctx, sk, actReserve.ID, actReserve.Ctime.Time().Unix(), 1); err != nil {
		return
	}
	if len(orderNos) != 1 {
		err = ecode.BwsOnlineInterReserveFailed
		return
	}
	orderNo = orderNos[0]
	item := strings.Split(orderNo, ":")
	if len(item) > 0 {
		if reserveNo, err = strconv.Atoi(item[len(item)-1]); err != nil {
			return
		}
		if strings.Contains(sk, fmt.Sprint(pb.ActInterReserveTicketType_VipTicket2021)) {
			reserveNo = int(actReserve.VipTicketNum) - reserveNo + 1
		} else {
			reserveNo = int(actReserve.VipTicketNum) + (int(actReserve.StandardTicketNum) - reserveNo) + 1
		}
	}
	return
}

func sortReserveList(reserveList []*ReserveRep) (total []*ReserveRep) {
	var available, other []*ReserveRep
	for _, v := range reserveList {
		if v.State == Available {
			available = append(available, v)
			continue
		}
		other = append(other, v)
	}

	var rrs [][]*ReserveRep
	rrs = append(rrs, available, other)
	for _, item := range rrs {
		if len(item) > 0 {
			sort.Slice(item, func(i, j int) bool {
				return item[i].DisplayIndex < item[j].DisplayIndex
			})
			total = append(total, item...)
		}
	}
	return total
}

func (s *Service) getBindInfo(ctx context.Context, mid int64) (records []*bwsonline.TicketBindRecord, err error) {

	if s.c.BwsOnline.BwPark.OpenCache > 0 {
		var ok bool
		if records, ok, err = s.dao.GetBindInfoFromCache(ctx, mid); err == nil && ok {
			log.Infoc(ctx, "GetBindInfoFromCache  succ , mid:%v , records len:%v", mid, len(records))
			return
		}
		log.Infoc(ctx, "GetBindInfoFromCache  failed , mid:%v , err:%v , ok:%v", mid, err, ok)
	}

	if records, err = s.dao.RawTicketsByMid(ctx, mid, s.c.BwsOnline.BwPark.Year); err != nil {
		log.Errorc(ctx, "RawTicketsByMid failed , err:%+v , records len:%v", err, len(records))
		return
	}
	_ = s.cache.SyncDo(ctx, func(ctx context.Context) {
		var expireTime = s.c.BwsOnline.BwPark.ShortExpireTime * 100
		if records == nil {
			records = []*bwsonline.TicketBindRecord{}
			expireTime = s.c.BwsOnline.BwPark.ShortExpireTime
		}
		err2 := s.dao.AddBindInfoCache(ctx, mid, records, expireTime)
		log.Infoc(ctx, "getBindInfo SyncDo AddBindInfoCache , mid:%v , err:%v", mid, err2)
	})
	return
}

func (s *Service) GetUserTicketInfo(ctx context.Context, pId string, idType int) (string, []*bwsonline.TicketInfo, error) {

	var (
		ticketsCache *bwsonline.TicketInfoFromHYG
		nowTime      = time.Now().Unix()
		err          error
	)
	if s.c.BwsOnline.BwPark.OpenCache > 0 {
		ticketsCache, err = s.dao.GetUserTicketInfosFromCache(ctx, pId, idType)
		if err == nil && ticketsCache != nil && nowTime-ticketsCache.UpdateTime < s.c.BwsOnline.BwPark.DefaultTicketExpireTime {
			log.Infoc(ctx, "GetUserTicketInfosFromCache  succ , pId:%v , time diff:%v", pId, nowTime-ticketsCache.UpdateTime)
			return ticketsCache.Name, ticketsCache.List, nil
		}
		log.Infoc(ctx, "GetUserTicketInfosFromCache  failed , pId:%v , err:%v , ticketsCache:%v", pId, err, ticketsCache)
	}

	var (
		name    string
		tickets []*bwsonline.TicketInfo
		err2    error
	)
	name, tickets, err2 = s.getTicketInfo(ctx, pId, idType)
	if err2 == nil {
		// 缓存已过期，但是主动从会员购拿到了数据，更新缓存
		_ = s.cache.SyncDo(ctx, func(ctx context.Context) {
			err2 := s.dao.AddUserTicketInfosCache(ctx, pId, idType, &bwsonline.TicketInfoFromHYG{
				Name:       name,
				List:       tickets,
				UpdateTime: nowTime,
			})
			log.Infoc(ctx, "GetUserTicketInfo SyncDo AddUserTicketInfosCache , pid:%v , err:%v", pId, err2)
		})
	}
	// 缓存已过期，从会员购取数据失败，使用缓存（兜底）
	if err2 != nil && err == nil && ticketsCache != nil {
		return ticketsCache.Name, ticketsCache.List, nil
	}
	return name, tickets, err2
}
