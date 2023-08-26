package bwsonline

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	xcode "go-common/library/ecode"
	"go-common/library/log"
	"go-common/library/sync/errgroup.v2"
	"go-gateway/app/web-svr/activity/ecode"
	pb "go-gateway/app/web-svr/activity/interface/api"
	"go-gateway/app/web-svr/activity/interface/model/bwsonline"
	"go-gateway/app/web-svr/activity/interface/tool"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"
)

type ReserveState int

const (
	Invalid ReserveState = iota
	Waiting
	Available
	Closed
	Reserved
	SaleOut
	OnlyForVip
)

type OnlineState int

const (
	OnlineStateNormal OnlineState = iota
	OnlineStateFinish
	OnlineStateOverdue
	OnlineStateDoing
)

type ReserveRep struct {
	ID                int64        `json:"reserve_id"`
	ActType           string       `json:"act_type"`
	ActTitle          string       `json:"act_title"`
	ActImg            string       `json:"act_img"`
	ActBeginTime      int64        `json:"act_begin_time"`
	ActEndTime        int64        `json:"act_end_time"`
	ReserveBeginTime  int64        `json:"reserve_begin_time"`
	ReserveEndTime    int64        `json:"reserve_end_time"`
	DescribeInfo      string       `json:"describe_info"`
	VipTicketNum      int32        `json:"vip_ticket_num"`
	StandardTicketNum int32        `json:"standard_ticket_num"`
	ScreenDate        int64        `json:"screen_date"`
	IsVipTicket       int          `json:"is_vip_ticket"`
	State             ReserveState `json:"state,omitempty"`
	OnlineState       OnlineState  `json:"online_state"`
	DisplayIndex      int32        `json:"display_index"`
	VipStock          int          `json:"vip_stock"`
	StandardStock     int          `json:"standard_stock"`
}

type ReserveList struct {
	UserTicketInfo  map[int64]*bwsonline.TicketInfo `json:"user_ticket_info,omitempty"`
	ReserveInfoList map[int64][]*ReserveRep         `json:"reserve_list"`
}

type MyReserveListOfflineReply struct {
	Data map[int64][]*UserReserveInfo `json:"data"`
}

type CheckReserveNo struct {
	ReserveNo int `json:"reserve_no"`
}

// BindTicketReserve 哔哩乐园-门票绑定接口
func (s *Service) BindTicketReserve(ctx context.Context, mid int64, uName string, pId string, idType int, ticketNo string) (int64, error) {
	var (
		name    string
		tickets []*bwsonline.TicketInfo
		idSum   string
		err     error
		records []*bwsonline.TicketBindRecord
	)

	if records, err = s.getBindInfo(ctx, mid); err != nil {
		return 0, err
	}
	if len(records) > 0 {
		return 0, ecode.BwsOnlineMidHasBind
	}

	if name, tickets, err = s.GetUserTicketInfo(ctx, pId, idType); err != nil {
		return 0, err
	}
	if name != uName {
		return 0, ecode.BwsOnlineTicketInfoNotMatch
	}

	for _, v := range tickets {
		if v.Ticket == ticketNo || (len(ticketNo) >= 4 && strings.HasSuffix(v.Ticket, ticketNo)) {
			log.Infoc(ctx, "compareTailStr , origin:%v , input:%v", v.Ticket, ticketNo)
			idSum = s.md5(pId)
			break
		}
	}
	if len(idSum) > 0 {
		pIdEncrypt := tool.CFBEncrypt(pId, s.c.BwsOnline.BwPark.AppSecret)
		return s.dao.AddTicketBindRecord(ctx, uName, mid, pIdEncrypt, idType, idSum, s.c.BwsOnline.BwPark.Year)
	}
	return 0, ecode.BwsOnlineTicketInfoNotMatch
}

// InterReserveList 哔哩乐园-预约列表 & 用户门票信息
func (s *Service) InterReserveList(ctx context.Context, mid int64, screenDates []int64) (rList ReserveList, err error) {
	var (
		tickets []*bwsonline.TicketInfo
	)
	if tickets, err = s.FindUserTickets(ctx, mid); err != nil || len(tickets) <= 0 {
		log.Warnc(ctx, "InterReserveList FindUserInfo err :%+v , tickets len:%v", err, len(tickets))
		return
	}

	rList = ReserveList{
		UserTicketInfo:  make(map[int64]*bwsonline.TicketInfo),
		ReserveInfoList: make(map[int64][]*ReserveRep),
	}
	nowTime := time.Now()
	today, _ := strconv.ParseInt(nowTime.Format("20060102"), 10, 64)
	defultDate := int64(20990730)
	for _, v := range tickets {
		var (
			sdate int64
			ok    bool
		)
		if sdate, ok = s.c.BwsOnline.BwPark.ScreenNameMap[v.ScreenName]; !ok {
			log.Warnc(ctx, "err ScreenName:%v", v.ScreenName)
			continue
		}
		// 同一个证件下，一天可能买了多张票，以vip票为准
		if _, ok = rList.UserTicketInfo[sdate]; !ok || v.SkuName == s.c.BwsOnline.BwPark.VipTag {
			tmpTicket := v
			tmpTicket.Tel = ""
			rList.UserTicketInfo[sdate] = tmpTicket
		}
		// 默认返回当天的预约信息
		if sdate >= today && sdate < defultDate {
			defultDate = sdate
		}
	}
	if len(rList.UserTicketInfo) <= 0 {
		return rList, ecode.SystemActivityParamsErr
	}

	if len(screenDates) <= 0 {
		screenDates = []int64{defultDate}
	}

	var newScreenDates []int64
	for _, sd := range screenDates {
		if ut, ok := rList.UserTicketInfo[sd]; ok && ut != nil {
			newScreenDates = append(newScreenDates, sd)
		}
	}

	if len(newScreenDates) == 0 {
		return
	}
	screenDates = newScreenDates

	var (
		list                  []*pb.ActInterReserve
		stateMap              map[int64]ReserveState
		stockMap, vipStockMap map[int64]int
	)
	if list, err = s.dao.RawInterReserveByDate(ctx, screenDates, s.c.BwsOnline.BwPark.Year); err != nil {
		return
	}

	reserveRepList := convert2ReserveRep(list, nowTime.Unix())
	if stateMap, stockMap, vipStockMap, err = s.getInterReserveState(ctx, reserveRepList, nowTime.Unix(), mid, rList.UserTicketInfo); err != nil {
		return
	}

	for _, rr := range reserveRepList {
		rr.State = stateMap[rr.ID]
		rr.StandardStock = stockMap[rr.ID]
		rr.VipStock = vipStockMap[rr.ID]

		rl := rList.ReserveInfoList[rr.ScreenDate]
		rList.ReserveInfoList[rr.ScreenDate] = append(rl, rr)
	}

	for key, value := range rList.ReserveInfoList {
		rList.ReserveInfoList[key] = sortReserveList(value)
	}
	return
}

// ReserveDo  哔哩乐园-预约互动活动
func (s *Service) ReserveDo(ctx context.Context, mid int64, interReserveId int64, ticketNo string) (rNo map[string]int, err error) {

	// 检查当前用户，对于选定互动场次的可预约状态
	var (
		actReserve *pb.ActInterReserve
		ticket     *bwsonline.TicketInfo
	)
	eg := errgroup.Group{}
	eg.Go(func(ctx context.Context) error {
		var err1 error
		if actReserve, err1 = s.dao.RawInterReserveById(ctx, interReserveId, s.c.BwsOnline.BwPark.Year); err1 != nil {
			return err1
		}
		if actReserve == nil || actReserve.ID != interReserveId {
			return ecode.ActivityNotExist
		}
		return nil
	})

	eg.Go(func(ctx context.Context) error {
		var (
			err2    error
			tickets []*bwsonline.TicketInfo
		)

		if tickets, err2 = s.FindUserTickets(ctx, mid); err2 != nil || len(tickets) <= 0 {
			log.Warnc(ctx, "ReserveDo FindUserInfo err :%+v , tickets len:%v", err2, len(tickets))
			return ecode.ActivityNotExist
		}
		for _, v := range tickets {
			if v.Ticket == ticketNo {
				if ticket == nil || v.SkuName == s.c.BwsOnline.BwPark.VipTag {
					ticket = v
				}
			}
		}
		if ticket == nil {
			return ecode.SystemActivityParamsErr
		}
		return nil
	})
	if err = eg.Wait(); err != nil {
		return
	}

	sdate, ok := s.c.BwsOnline.BwPark.ScreenNameMap[ticket.ScreenName]
	if !ok || sdate != actReserve.ScreenDate {
		log.Warnc(ctx, "ReserveDo err date map , ok:%v , sdate:%v , ScreenDate:%v", ok, sdate, actReserve.ScreenDate)
		err = ecode.SystemActivityParamsErr
		return
	}

	var (
		stateMap    map[int64]ReserveState
		vipStockMap map[int64]int
		nowTime     int64
	)
	nowTime = time.Now().Unix()
	reserveRepList := convert2ReserveRep([]*pb.ActInterReserve{actReserve}, nowTime)
	stateMap, _, vipStockMap, err = s.getInterReserveState(ctx, reserveRepList, nowTime, mid, map[int64]*bwsonline.TicketInfo{sdate: ticket})
	if err != nil {
		return
	}

	if state, ok := stateMap[actReserve.ID]; state != Available || len(reserveRepList) != 1 || reserveRepList[0] == nil {
		log.Infoc(ctx, "ReserveDo failed , state is :%v ,ok:%v", state, ok)
		err = ecode.BwsOnlineInterReserveFailed
		return
	}

	var reserveNo int
	var orderNo string
	// 当前场次是vip预约场，或者vip场已经结束（但是票没有卖完），则优先扣减vip的库存
	vipstock, _ := vipStockMap[reserveRepList[0].ID]
	if reserveRepList[0].IsVipTicket == 1 ||
		(actReserve.VipReserveEndTime < nowTime && actReserve.VipReserveEndTime <= actReserve.ReserveBeginTime && vipstock > 0) {
		// vip 票扣减预约库存
		sk := getStockId(s.c.BwsOnline.BwPark.Year, pb.ActInterReserveTicketType_VipTicket2021)
		orderNo, reserveNo, err = s.getReserveNo(ctx, sk, actReserve)
	}

	if reserveRepList[0].IsVipTicket != 1 && (err != nil || orderNo == "" || reserveNo <= 0) {
		// 普通票扣减预约库存
		sk := getStockId(s.c.BwsOnline.BwPark.Year, pb.ActInterReserveTicketType_StandardTicket2021)
		orderNo, reserveNo, err = s.getReserveNo(ctx, sk, actReserve)
	}

	if err != nil || orderNo == "" || reserveNo <= 0 || reserveNo > int(reserveRepList[0].StandardTicketNum+reserveRepList[0].VipTicketNum) {
		log.Errorc(ctx, "ReserveDo failed , order_no:%v ,reserve_no:%v , err:%+v", orderNo, reserveNo, err)
		if err == nil {
			err = ecode.BwsOnlineInterReserveFailed
		}
		return
	}
	//var  lastId int64
	var _ int64
	if _, err = s.dao.AddInterReserveOrder(ctx, mid, ticketNo, interReserveId, orderNo, s.c.BwsOnline.BwPark.Year, reserveNo); err != nil {
		return
	}
	return map[string]int{
		"reserve_no": reserveNo,
	}, nil
}

type UserReserveInfo struct {
	*bwsonline.InterReserveOrder
	*ReserveRep
}

// ReservedList 获取我的预约列表
func (s *Service) MyReservedList(ctx context.Context, mid int64) (rList map[int64][]*UserReserveInfo, err error) {
	var orders []*bwsonline.InterReserveOrder
	if orders, err = s.dao.RawInterReserveOrderByMid(ctx, mid, s.c.BwsOnline.BwPark.Year); err != nil {
		return
	}
	reservedIds := []int64{}

	orderMap := make(map[int64]*bwsonline.InterReserveOrder)
	for _, v := range orders {
		reservedIds = append(reservedIds, v.InterReserveId)
		orderMap[v.InterReserveId] = v
	}

	rList = make(map[int64][]*UserReserveInfo)
	if len(reservedIds) > 0 {
		var reserveList []*pb.ActInterReserve
		if reserveList, err = s.dao.RawInterReserveByIds(ctx, reservedIds, s.c.BwsOnline.BwPark.Year); err != nil {
			return
		}
		for _, record := range reserveList {

			order := orderMap[record.ID]
			reserveInfo := convert2ReserveRep([]*pb.ActInterReserve{record}, order.Ctime.Time().Unix())
			if len(reserveInfo) <= 0 {
				continue
			}
			rl := rList[record.ScreenDate]
			rList[record.ScreenDate] = append(rl, &UserReserveInfo{
				InterReserveOrder: order,
				ReserveRep:        reserveInfo[0],
			})
		}
	}
	return
}

func (s *Service) FindUserTickets(ctx context.Context, mid int64) (tickets []*bwsonline.TicketInfo, err error) {
	var records []*bwsonline.TicketBindRecord
	if records, err = s.getBindInfo(ctx, mid); err != nil {
		return
	}
	if len(records) <= 0 {
		return nil, ecode.BwsOnlineNotBindTicket
	}
	pIdDecrypt := tool.CFBDecrypt(records[0].PersonalId, s.c.BwsOnline.BwPark.AppSecret)
	if len(pIdDecrypt) <= 0 {
		log.Infoc(ctx, "PersonalId CFBDecrypt failed, personal_id:%v", records[0].PersonalId)
		return
	}
	_, tickets, err = s.GetUserTicketInfo(ctx, pIdDecrypt, records[0].PersonalIdType)
	return
}

func (s *Service) getTicketInfo(ctx context.Context, pId string, idType int) (name string, tickets []*bwsonline.TicketInfo, err error) {
	params := make(map[string]interface{})
	params["timestamp"] = time.Now().Unix()
	params["appKey"] = s.c.BwsOnline.BwPark.AppKey
	params["projectId"] = s.c.BwsOnline.BwPark.ProjectId
	params["personalId"] = pId
	params["idType"] = idType
	var strs []string
	for key := range params {
		strs = append(strs, key)
	}
	sort.Strings(strs)
	var md5Url string
	for _, k := range strs {
		md5Url = md5Url + fmt.Sprintf("%s=%v&", k, params[k])
	}
	log.Infoc(ctx, "GetTicketInfo md5Url :%v", strings.TrimSuffix(md5Url, "&"))
	params["sign"] = s.md5(strings.TrimSuffix(md5Url, "&"))

	bytesData, err := json.Marshal(params)
	if err != nil {
		log.Errorc(ctx, "GetTicketInfo json.Marshal params(%+v) error(%v)", params, err)
		return
	}
	var (
		req  *http.Request
		resp = struct {
			ErrNo  int         `json:"errno"`
			ErrTag int         `json:"errtag"`
			Msg    string      `json:"msg"`
			Data   interface{} `json:"data"`
		}{}
	)
	if req, err = http.NewRequest(http.MethodPost, s.c.BwsOnline.BwPark.SearchByIdUrl, bytes.NewReader(bytesData)); err != nil {
		log.Errorc(ctx, "GetTicketInfo http.NewRequest url(%s) error(%v)", s.c.BwsOnline.BwPark.SearchByIdUrl+"?"+string(bytesData), err)
		return
	}
	req.Header.Set("Content-Type", "application/json")
	if err = s.httpClient.Do(ctx, req, &resp); err != nil {
		log.Errorc(ctx, "GetTicketInfo d.httpClient.Post sendMsgURL(%s) error(%v)", s.c.BwsOnline.BwPark.SearchByIdUrl+"?"+string(bytesData), err)
		return
	}
	respStr, _ := json.Marshal(resp)
	log.Infoc(ctx, "GetTicketInfo success by sendMsgURL(%s) , resp:%s", s.c.BwsOnline.BwPark.SearchByIdUrl+"?"+string(bytesData), string(respStr))

	if resp.ErrNo != xcode.OK.Code() {
		err = errors.Wrapf(ecode.BwsOnlineTicketServerErr, "error code(%v)", resp.ErrNo)
		return
	}

	if value, ok := resp.Data.(bool); ok && !value {
		err = ecode.BwsOnlineTicketInfoNotFind
		return
	}

	bytes, _ := json.Marshal(resp.Data)
	hdata := HygRepData{}
	if err = json.Unmarshal(bytes, &hdata); err != nil {
		return
	}

	return hdata.Name, hdata.List, err
}

func (s *Service) IncrReserveStock(ctx context.Context, req *pb.GiftStockReq) (replay int64, err error) {
	return s.dao.IncrStock(ctx, req.SID, req.GiftID, req.GiftVer, int(req.GiftNum))
}

func (s *Service) HasVipTickets(ctx context.Context, mid int64, ticketDate int64) (res []*bwsonline.TicketInfo, err error) {
	var (
		allTickets []*bwsonline.TicketInfo
	)
	if allTickets, err = s.FindUserTickets(ctx, mid); err != nil || len(allTickets) <= 0 {
		log.Warnc(ctx, "HasVipTickets FindUserInfo err :%+v , tickets len:%v", err, len(allTickets))
		return
	}
	sa, _ := json.Marshal(allTickets)
	log.Infoc(ctx, "has vip tickets allTickets(%s) mid(%d) data(%d)", sa, mid, ticketDate)
	for _, v := range allTickets {
		var (
			sdate int64
			ok    bool
		)
		if sdate, ok = s.c.BwsOnline.BwPark.ScreenNameMap[v.ScreenName]; !ok {
			log.Warnc(ctx, "err ScreenName:%v", v.ScreenName)
			continue
		}
		if sdate == ticketDate && v.SkuName == s.c.BwsOnline.BwPark.VipTag {
			tmpTicket := v
			tmpTicket.Tel = ""
			res = append(res, tmpTicket)
		}
	}
	return
}

// CheckedReserve 核销
func (s *Service) CheckedReserve(ctx context.Context, mid int64, reserveID int64) (r *CheckReserveNo, err error) {
	r = &CheckReserveNo{}
	// 是否预约
	res, err := s.dao.RawMidInterReserveID(ctx, mid, reserveID, s.c.BwsOnline.BwPark.Year)
	if err != nil {
		log.Errorc(ctx, "s.dao.RawMidInterReserveID err(%v)", err)
		return
	}
	if res == nil {
		err = ecode.BwsNotReserveError
		return
	}
	update, err := s.dao.CheckReserveByID(ctx, res.Id, s.c.BwsOnline.BwPark.Year)
	if update == 0 {
		err = ecode.BwsNotReserveDuplicateError
		return
	}
	r.ReserveNo = res.ReserveNo
	if err != nil {
		log.Errorc(ctx, "s.dao.CheckReserveByID err(%v)", err)
	}
	return
}

// OfflineMyReserveList 线下当日预约
func (s *Service) OfflineMyReserveList(ctx context.Context, mid int64) (res *MyReserveListOfflineReply, err error) {
	day := todayDate()
	res = new(MyReserveListOfflineReply)
	res.Data = make(map[int64][]*UserReserveInfo)
	now := time.Now().Unix()
	reserve, err := s.MyReservedList(ctx, mid)
	if err != nil {
		log.Errorc(ctx, "s.MyReservedList mid(%d) err(%v)", mid, err)
		return
	}
	todayList := make([]*UserReserveInfo, 0)

	if reserve != nil {
		for k, v := range reserve {
			if _, ok := res.Data[k]; !ok {
				res.Data[k] = make([]*UserReserveInfo, 0)
			}
			if k < day {
				for _, re := range v {
					if re.IsChecked != 1 {
						re.OnlineState = OnlineStateOverdue
					} else {
						re.OnlineState = OnlineStateFinish
					}
					res.Data[k] = append(res.Data[k], re)

				}
				continue
			}
			if k == day {
				for _, re := range v {
					if re.ActBeginTime > now {
						todayList = append(todayList, re)
						continue
					}
					if re.ActEndTime >= now && re.ActBeginTime <= now {
						if re.IsChecked != 1 {
							re.OnlineState = OnlineStateDoing
						} else {
							re.OnlineState = OnlineStateFinish
						}
						todayList = append(todayList, re)
						continue
					}
					if re.ActEndTime < now {
						if re.IsChecked != 1 {
							re.OnlineState = OnlineStateOverdue
						} else {
							re.OnlineState = OnlineStateFinish
						}
						todayList = append(todayList, re)
					}
				}
				continue
			}

			for _, re := range v {
				res.Data[k] = append(res.Data[k], re)
			}
		}
		if len(todayList) > 0 {
			res.Data[day] = make([]*UserReserveInfo, 0)
			orderList := make([]OnlineState, 0)
			orderList = append(orderList, OnlineStateDoing, OnlineStateNormal, OnlineStateFinish, OnlineStateOverdue)
			for _, v := range orderList {
				for _, t := range todayList {
					if v == t.OnlineState {
						res.Data[day] = append(res.Data[day], t)
					}
				}
			}
		}
	}
	return
}
func (s *Service) SyncStock(ctx context.Context, req *pb.GiftStockReq) (syncResp *pb.SyncGiftStockResp, err error) {
	syncResp = new(pb.SyncGiftStockResp)
	sid := getStockId(s.c.BwsOnline.BwPark.Year, req.SID)
	var stocks []string
	if stocks, err = s.dao.RandGetStockNo(ctx, sid, req.GiftID, int(req.GiftNum)); err != nil {
		return
	}
	log.Infoc(ctx, "SyncStock sid:%s , stocks:%v", sid, stocks)
	if len(stocks) > 0 {
		var orders []*bwsonline.InterReserveOrder
		if orders, err = s.dao.RawInterReserveOrderByOrderNos(ctx, stocks, s.c.BwsOnline.BwPark.Year); err != nil {
			return
		}
		log.Infoc(ctx, "SyncStock sid:%s , orders num:%v", sid, len(orders))

		orderCheckMap := make(map[string]struct{})
		for _, v := range orders {
			orderCheckMap[v.OrderNo] = struct{}{}
		}

		//  redis中的库存号码，跟DB 中的库存号码进行对比
		var repeatStock []string
		for _, stockKey := range stocks {
			if stockKey == "" {
				continue
			}

			if _, ok := orderCheckMap[stockKey]; ok {
				repeatStock = append(repeatStock, stockKey)
				continue
			}

			var replay int
			if replay, err = s.dao.SmoveStockNo(ctx, sid, req.GiftID, stockKey); err != nil {
				return
			}
			syncResp.FixNum = syncResp.FixNum + int32(replay)
		}

		// 已经确认的库存号码，进行删除
		if len(repeatStock) > 0 {
			var ackNum int64
			ackNum, err = s.dao.AckStockNo(ctx, sid, req.GiftID, repeatStock)
			syncResp.AckNum = int32(ackNum)
			log.Infoc(ctx, "SyncStock  repeatStock:%v , ackNum:%v , err:%+v", repeatStock, ackNum, err)
		}
	}
	return
}

func (s *Service) GetBeginReserve(ctx context.Context, req *pb.BwParkBeginReserveReq) (list []*pb.ActInterReserve, err error) {
	var (
		listStand, listVip []*pb.ActInterReserve
		reserveMap         map[int64]*pb.ActInterReserve
	)
	if listStand, err = s.dao.RawInterReserveByTime(ctx, req.BeginTime, req.EndTime, s.c.BwsOnline.BwPark.Year, false); err != nil {
		return
	}
	if listVip, err = s.dao.RawInterReserveByTime(ctx, req.BeginTime, req.EndTime, s.c.BwsOnline.BwPark.Year, true); err != nil {
		return
	}
	reserveMap = make(map[int64]*pb.ActInterReserve)
	listStand = append(listStand, listVip...)
	for _, v := range listStand {
		reserveMap[v.ID] = v
	}
	for _, value := range reserveMap {
		list = append(list, value)
	}
	return
}
