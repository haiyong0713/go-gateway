package lottery

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	xtime "time"

	"go-common/library/database/sql"
	"go-common/library/ecode"
	"go-common/library/log"
	lotmdl "go-gateway/app/web-svr/activity/admin/model/lottery"

	vipresource "git.bilibili.co/bapis/bapis-go/vip/resource/service"
)

var (
	giftTypeToStr = map[int]string{
		1:  "实物奖品",
		2:  "大会员",
		3:  "头像挂件",
		4:  "优惠券",
		5:  "硬币",
		6:  "大会员抵用券",
		7:  "其他奖品",
		8:  "OGV券",
		9:  "会员购券",
		10: "现金",
	}
)

const (
	moneyLimitMaxNeedWhiteList = 500000
	vipParams                  = "{\"token\":\"%s\",\"app_key\":\"\"}"
	vipCouponParams            = "{\"token\":\"%s\",\"app_key\":\"\"}"
)

// List get lottery information list
func (s *Service) List(c context.Context, request *lotmdl.ListParam) (rsp lotmdl.ListRsp, err error) {
	var (
		total int
		Page  = &lotmdl.Page{}
		list  []*lotmdl.LotInfo
	)
	if total, err = s.lotDao.ListTotal(c, request.State, request.Keyword); err != nil {
		log.Error("s.lotDao.ListTotal() failed. error(%v)", err)
		return
	}
	if list, err = s.lotDao.BaseList(c, request.Pn, request.Ps, request.State, request.Keyword, request.Rank); err != nil {
		log.Error("s.lotDao.BaseList(%v,%v,%v,%v,%v) failed. error(%v)", request.Pn, request.Ps, request.State, request.Keyword, request.Rank, err)
		return
	}
	Page.Num = request.Pn
	Page.Size = request.Ps
	Page.Total = total
	rsp.Page = Page
	rsp.List = list
	return
}

// Add add lottery base information.
func (s *Service) Add(c context.Context, request *lotmdl.AddParam, operator string) (err error) {
	var (
		tx      *sql.Tx
		id      int64
		lotInfo *lotmdl.LotInfo
	)
	if tx, err = s.lotDao.BeginTran(c); err != nil {
		log.Errorc(c, "s.lotDao.BeginTran() failed. error(%v)", err)
		return
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			log.Errorc(c, "%v", r)
			return
		}
		if err != nil {
			if err1 := tx.Rollback(); err1 != nil {
				log.Errorc(c, "tx.Rollback() error(%v)", err1)
			}
			return
		}
		if err = tx.Commit(); err != nil {
			log.Errorc(c, "tx.Commit() error(%v)", err)
		}
	}()
	if id, err = s.lotDao.Create(tx, request.Name, operator, request.Stime, request.Etime, request.Type); err != nil {
		log.Errorc(c, "Add s.lotDao.Add(%v, %v, %v, %v) failed. error(%v)", request.Name, request.Stime, request.Etime, request.Type, err)
		return
	}
	if lotInfo, err = s.lotDao.LotDetailByID(c, id); err != nil {
		log.Errorc(c, "Add s.lotDao.LotDetailByID() failed. error(%v)", err)
		return
	}
	if err = s.lotDao.InitLotDetail(tx, c, lotInfo.LotteryID); err != nil {
		log.Errorc(c, "Add s.lotDao.InitLotDetail() failed. error(%v)", err)
		return
	}
	return
}

// LotteryRecord ..
func (s *Service) LotteryRecord(c context.Context, sid string, mid int64) (res *lotmdl.RecordDetailRes, err error) {
	var lotInfo *lotmdl.LotInfo
	var gift []*lotmdl.GiftInfo
	res = &lotmdl.RecordDetailRes{}
	if lotInfo, err = s.lotDao.LotDetailBySID(c, sid); err != nil {
		log.Errorc(c, "Add s.lotDao.LotDetailBySID() failed. error(%v)", err)
		return
	}
	if lotInfo == nil {
		return nil, ecode.Error(ecode.RequestErr, "抽奖id不存在")
	}
	if gift, err = s.lotDao.AllGift(c, sid); err != nil {
		log.Errorc(c, "s.lotDao.AllGift(%v) failed. error(%v)", sid, err)
		return
	}
	giftMap := make(map[int64]string)
	for _, v := range gift {
		giftMap[v.ID] = v.Name
	}
	rawList, err := s.lotDao.RawLotteryUsedTimes(c, lotInfo.ID, mid)
	if err != nil {
		return nil, err
	}
	for i, v := range rawList {
		giftName, ok := giftMap[v.GiftID]
		if ok {
			rawList[i].GiftName = giftName
		}
	}
	res.List = rawList
	return

}

// Delete ...
func (s *Service) Delete(c context.Context, id int64, operator string) (err error) {
	var tx *sql.Tx
	if tx, err = s.lotDao.BeginTran(c); err != nil {
		log.Errorc(c, "s.lotDao.BeginTran() failed. error(%v)", err)
		return
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			log.Errorc(c, "%v", r)
			return
		}
		if err != nil {
			if err1 := tx.Rollback(); err1 != nil {
				log.Errorc(c, "tx.Rollback() error(%v)", err1)
			}
			return
		}
		if err = tx.Commit(); err != nil {
			log.Errorc(c, "tx.Commit() error(%v)", err)
		}
	}()
	if err = s.lotDao.Delete(tx, c, id, operator); err != nil {
		log.Errorc(c, "s.lotDao.Delete(%v) failed. error(%v)", id, err)
		return
	}
	if err = s.lotDao.DeleteDraft(c, tx, id, operator); err != nil {
		log.Errorc(c, "s.lotDao.DeleteDraft(%v) failed. error(%v)", id, err)
	}
	return
}

// Detail ...
func (s *Service) Detail(c context.Context, sid string) (rsp *lotmdl.LotDetailInfo, err error) {
	var (
		list        *lotmdl.LotInfo
		info        *lotmdl.RuleInfo
		timeConf    []*lotmdl.TimesConf
		gift        []*lotmdl.GiftInfo
		memberGroup []*lotmdl.MemberGroupDB
	)
	if list, err = s.lotDao.LotDetailBySID(c, sid); err != nil {
		log.Errorc(c, "s.lotDao.LotDetailBySID(%v) failed. error(%v)", sid, err)
		return
	}
	if info, err = s.lotDao.GetLotRuleBySID(c, sid); err != nil {
		log.Errorc(c, "s.lotDao.LotRuleBySID(%v) failed. error(%v)", sid, err)
		return
	}
	if timeConf, err = s.lotDao.AllTimesConf(c, sid); err != nil {
		log.Errorc(c, "s.lotDao.AllTimesConf(%v) failed. error(%v)", sid, err)
		return
	}
	if gift, err = s.lotDao.AllGift(c, sid); err != nil {
		log.Errorc(c, "s.lotDao.AllGift(%v) failed. error(%v)", sid, err)
		return
	}
	if memberGroup, err = s.lotDao.AllMemberGroup(c, sid); err != nil {
		log.Errorc(c, "s.lotDao.AllMemberGroup(%v) failed. error(%v)", sid, err)
		return
	}
	rsp = &lotmdl.LotDetailInfo{}
	rsp.List = *list
	rsp.Info = *info
	for _, timeItem := range timeConf {
		if timeItem.Type == 1 {
			rsp.LotteryTimes = timeItem
			continue
		}
		if timeItem.Type == 2 {
			rsp.PriceTimes = timeItem
			continue
		}
		rsp.TimesConf = append(rsp.TimesConf, timeItem)
	}
	if gift != nil {
		for k, v := range gift {
			gift[k].ProbabilityF = gift[k].GetFloatProbability()
			gift[k].Params = s.getParams(c, v.Source, v.Params, v.Type)
		}
	}
	rsp.Gift = gift
	rsp.MemberGroup = memberGroup
	return
}

func (s *Service) getParams(c context.Context, source string, params string, giftType int) string {
	if params == "" {
		if source != "" {
			switch giftType {
			case lotmdl.GiftTypeVIP:
				return fmt.Sprintf(vipParams, source)
			case lotmdl.GiftTypeCoupon:
				return fmt.Sprintf(vipCouponParams, source)
			default:
				return source
			}
		}
		return source
	}
	return params

}

// Edit 编辑
func (s *Service) Edit(c context.Context, request *lotmdl.EditParam, cookie, operator string) (err error) {
	var (
		tx            *sql.Tx
		list          = &lotmdl.LotInfo{}
		rule          *lotmdl.RuleInfo
		timesUpdate   []*lotmdl.TimesConf
		timesAdd      []*lotmdl.TimesConf
		baseConf      = &lotmdl.BaseTimes{}
		actionAddConf = make([]*lotmdl.AddTimes, 0)
		vipCheck      bool
	)

	if tx, err = s.lotDao.BeginTran(c); err != nil {
		log.Errorc(c, "s.lotDao.BeginTran() failed. error(%v)", err)
		return
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			log.Errorc(c, "%v", r)
			err = ecode.Error(ecode.RequestErr, "编辑失败")
			return
		}
		if err != nil {
			if err1 := tx.Rollback(); err1 != nil {
				log.Errorc(c, "tx.Rollback() error(%v)", err1)
			}
			return
		}
		if err = tx.Commit(); err != nil {
			log.Errorc(c, "tx.Commit() error(%v)", err)
		}
	}()
	if list, err = s.lotDao.LotDetailBySID(c, request.SID); err != nil {
		log.Errorc(c, "s.lotDao.LotDetailByID(id: %v) failed. error(%v)", request.ID, err)
		return
	}
	if err = s.lotDao.UpdateLotInfo(tx, c, list.ID, request.IsInternal, request.Name, operator, request.Stime, request.Etime); err != nil {
		log.Errorc(c, "s.lotDao.UpdateLotInfo(id: %v,is_internal:%d name: %v, stime: %v, etime: %v) failed. error(%v)",
			list.ID, request.IsInternal, request.Name, request.Stime, request.Etime, err)
		return
	}
	if rule, err = s.lotDao.GetLotRuleBySID(c, list.LotteryID); err != nil {
		log.Errorc(c, "s.lotDao.GetLotRuleBySID(sid: %v) failed. error(%v)", list.LotteryID, err)
		return
	}
	if request.ActionAdd != "" {
		if err = json.Unmarshal([]byte(request.ActionAdd), &actionAddConf); err != nil {
			log.Errorc(c, "json.Unmarshal(action_add: %v) failed. error(%v)", request.ActionAdd, err)
			return
		}
		for _, item := range actionAddConf {
			if item.Type == lotmdl.AddActionTypeVIP {
				if item.Info == "" {
					err = ecode.Error(ecode.RequestErr, "未填写大会员套餐ID")
					return
				}
				if vipCheck, err = s.lotDao.GetVIPInfo(c, item.Info, cookie); err != nil {
					log.Errorc(c, "s.lotDao.GetVIPInfo(%v) failed. error(%v)", item.Info, err)
					err = ecode.Error(ecode.RequestErr, err.Error())
					return
				}
				if !vipCheck {
					log.Errorc(c, "vip config bad. info: %v", item.Info)
					err = ecode.Error(ecode.RequestErr, "大会员套餐ID错误，请确认后重新提交")
					return
				}
			}
			if item.Type == lotmdl.AddActionTypeCustom || item.Type == lotmdl.AddActionTypeOGV {
				if item.Info == "" {
					err = ecode.Error(ecode.RequestErr, "未填写自定义行为ID")
					return
				}
				var actionRes []*lotmdl.TimesConf
				if actionRes, err = s.lotDao.CheckAction(c, item.Type, item.Info); err != nil {
					log.Errorc(c, "s.lotDao.CheckAction(type: %v, info: %v) failed. error(%v)", item.Type, item.Info, err)
					return
				}
				for _, res := range actionRes {
					if res.ID != item.ID {
						err = ecode.Error(ecode.RequestErr, fmt.Sprintf("当前自定义行为ID已存在,sid: %v", res.Sid))
						return
					}
				}
			}
			tmp := &lotmdl.TimesConf{}
			tmp.ID = item.ID
			tmp.Type = item.Type
			tmp.Info = item.Info
			tmp.Times = item.Times
			tmp.AddType = item.AddType
			tmp.Most = item.Most
			tmp.Sid = list.LotteryID
			switch item.Status {
			case 1:
				tmp.State = 0
				timesUpdate = append(timesUpdate, tmp)
			case 2:
				tmp.State = 1
				timesUpdate = append(timesUpdate, tmp)
			case 3:
				tmp.State = 0
				timesAdd = append(timesAdd, tmp)
			default:
			}
		}
	}
	rule.Level = request.Level
	rule.RegtimeStime = request.RegTimeSTime
	rule.RegtimeEtime = request.RegTimeETime
	rule.VipCheck = request.VipCheck
	rule.Coin = request.CoinCheck
	rule.FsIP = request.FsIP
	rule.GiftRate = request.Rate
	rule.HighType = request.HighType
	rule.HighRate = request.HighRate
	rule.AccountCheck = request.AccountCheck
	rule.SenderMid = request.SenderMid
	rule.ActivityLink = request.ActivityLink
	if _, err = s.lotDao.RuleUpdate(tx, c, rule); err != nil {
		log.Errorc(c, "s.lotDao.RuleUpdate() failed. error(%v)", err)
		return
	}
	if request.LotTimes != "" {
		if err = json.Unmarshal([]byte(request.LotTimes), baseConf); err != nil {
			log.Errorc(c, "json.Unmarshal(lottery_times: %v) failed. error(%v)", request.LotTimes, err)
			return
		}
		tmp := &lotmdl.TimesConf{}
		tmp.ID = baseConf.ID
		tmp.Sid = list.LotteryID
		tmp.AddType = baseConf.AddType
		tmp.Times = baseConf.Times
		tmp.Most = baseConf.Times
		tmp.Type = 1
		if baseConf.ID != 0 {
			timesUpdate = append(timesUpdate, tmp)
		} else {
			timesAdd = append(timesAdd, tmp)
		}
	}
	if request.PriceTimes != "" {
		if err = json.Unmarshal([]byte(request.PriceTimes), baseConf); err != nil {
			log.Errorc(c, "json.Unmarshal(price_times: %v) failed. error(%v)", request.PriceTimes, err)
			return
		}
		tmp := &lotmdl.TimesConf{}
		tmp.ID = baseConf.ID
		tmp.Sid = list.LotteryID
		tmp.AddType = baseConf.AddType
		tmp.Times = baseConf.Times
		tmp.Most = baseConf.Times
		tmp.Type = 2
		if baseConf.ID != 0 {
			timesUpdate = append(timesUpdate, tmp)
		} else {
			timesAdd = append(timesAdd, tmp)
		}
	}
	if len(timesAdd) != 0 {
		if _, err = s.lotDao.TimesAddBatch(tx, c, timesAdd); err != nil {
			log.Errorc(c, "s.lotDao.TimesAddBatch(timesAdd: %+v), error(%v)", timesAdd, err)
			return
		}
	}
	if len(timesUpdate) != 0 {
		if _, err = s.lotDao.TimesUpdateBatch(tx, c, timesUpdate); err != nil {
			log.Errorc(c, "s.lotDao.TimesUpdateBatch(timesUpdate: %+v), error(%v)", timesUpdate, err)
		}
	}
	return
}

// updateOrInsertMemberGroup 更新或插入用户组
func (s *Service) updateOrInsertMemberGroup(c context.Context, tx *sql.Tx, sid string, memberGroup []*lotmdl.MemberGroupDB) (err error) {
	if memberGroup != nil && len(memberGroup) > 0 {
		if err = s.lotDao.BatchInsertOrUpdateMemberGroup(c, tx, sid, memberGroup); err != nil {
			log.Errorc(c, "s.lotDao.BatchInsertOrUpdateMemberGroup(memberGroup: %+v), error(%v)", memberGroup, err)
		}
	}
	return err
}

// numLimit 奖品数量限制
func (s *Service) numLimit(c context.Context, num int, operator string) error {
	if num > moneyLimitMaxNeedWhiteList {
		if s.c.Lottery.NumLimit != nil && len(s.c.Lottery.NumLimit) > 0 {
			var operateInLimit bool
			for _, v := range s.c.Lottery.NumLimit {
				if v == operator {
					operateInLimit = true
					break
				}
			}
			if !operateInLimit {
				err := ecode.Error(ecode.RequestErr, "请联系管理员配置高额数量")
				log.Errorc(c, "operator(%s) err(%v)", operator, err)
				return err
			}
		}
	}
	return nil
}

// GiftAdd add gift information
func (s *Service) GiftAdd(c context.Context, request *lotmdl.GiftAddParam, cookie, operator string) (err error) {
	var (
		tx      *sql.Tx
		lottery *lotmdl.LotInfo
		coupon  *lotmdl.CouponInfo
		aid     int64
	)
	err = s.numLimit(c, request.Num, operator)
	if err != nil {
		return err
	}
	if err = request.GetDayStore(); err != nil {
		return err
	}
	if request.Type == lotmdl.GiftTypeVIP {
		if request.Params == "" && request.Source == "" {
			err = ecode.Error(ecode.RequestErr, "params为空，请检查后再次提交")
			return
		}
		params := lotmdl.GiftTypeVIPParams{}
		if err = json.Unmarshal([]byte(request.Params), &params); err != nil {
			log.Errorc(c, "json.Unmarshal(action_add: %v) failed. error(%v)", request.Params, err)
			return
		}
		if params.Token == "" || params.AppKey == "" {
			err = ecode.Error(ecode.RequestErr, "Token 或者appKey为空")
			return
		}
		request.Source = params.Token
		var resourceInfo *vipresource.ResourceInfoReply
		resourceInfo, err = s.resourceClient.ResourceInfoByTokenV2(c, &vipresource.ResourceInfoByTokenV2Req{Token: params.Token, Appkey: params.AppKey})
		if err != nil {
			log.Errorc(c, "Failed to get resource info by token: %s, %s, %+v", params.Token, params.AppKey, err)
			return
		}
		if resourceInfo == nil || resourceInfo.Resource == nil || resourceInfo.Resource.ID <= 0 {
			err = ecode.Error(ecode.RequestErr, "params为空，请检查后再次提交")
			return
		}
		if lottery, err = s.lotDao.LotDetailBySID(c, request.SID); err != nil {
			log.Errorc(c, "s.lotDao.LotDetailBySID(%v) failed. error(%v)", request.SID, err)
			return
		}
		if lottery.ETime > resourceInfo.Resource.EndTime {
			err = ecode.Error(ecode.RequestErr, "大会员token有效期结束时间小于活动结束时间")
			return
		}
		if resourceInfo.Resource.StartTime > lottery.STime {
			err = ecode.Error(ecode.RequestErr, "大会员token有效开始时间大于活动开始时间")
			return
		}
	} else if request.Type == lotmdl.GiftTypeGrant {
		if request.Params == "" && request.Source == "" {
			err = ecode.Error(ecode.RequestErr, "params为空，请检查后再次提交")
			return
		}
		params := lotmdl.GiftTypeGrantParams{}
		if err = json.Unmarshal([]byte(request.Params), &params); err != nil {
			log.Errorc(c, "json.Unmarshal(action_add: %v) failed. error(%v)", request.Params, err)
			return
		}
		request.Source = request.Params

	} else if request.Type == lotmdl.GiftTypeCoin {
		if request.Params == "" && request.Source == "" {
			err = ecode.Error(ecode.RequestErr, "params为空，请检查后再次提交")
			return
		}
		params := lotmdl.GiftTypeCoinParams{}
		if err = json.Unmarshal([]byte(request.Params), &params); err != nil {
			log.Errorc(c, "json.Unmarshal(action_add: %v) failed. error(%v)", request.Params, err)
			return
		}
		request.Source = strconv.Itoa(params.Coin)
	} else if request.Type == lotmdl.GiftTypeCoupon {
		if request.Params == "" && request.Source == "" {
			err = ecode.Error(ecode.RequestErr, "params为空，请检查后再次提交")
			return
		}
		params := lotmdl.GiftTypeVipCouponParams{}
		if err = json.Unmarshal([]byte(request.Params), &params); err != nil {
			log.Errorc(c, "json.Unmarshal(action_add: %v) failed. error(%v)", request.Params, err)
			return
		}
		if params.Token == "" {
			err = ecode.Error(ecode.RequestErr, "Token为空")
			return
		}
		request.Source = params.Token

		if coupon, err = s.lotDao.GetCouponInfo(c, params.Token, cookie); err != nil {
			log.Errorc(c, "s.lotDao.GetCouponInfo(%v) failed. error(%v) ", request.Source, err)
			return
		}
		if coupon == nil || coupon.ID == 0 {
			log.Errorc(c, "GiftAdd coupon is empty.")
			err = ecode.Error(ecode.RequestErr, "token不存在，请检查输入的token是否正确")
			return
		}
		if lottery, err = s.lotDao.LotDetailBySID(c, request.SID); err != nil {
			log.Errorc(c, "s.lotDao.LotDetailBySID(%v) failed. error(%v)", request.SID, err)
			return
		}
		if lottery.ETime > coupon.ActivateEnd {
			err = ecode.Error(ecode.RequestErr, "优惠券token有效期小于活动结束时间")
			return
		} else if coupon.ActivateStart > lottery.STime {
			err = ecode.Error(ecode.RequestErr, "优惠券token有效开始时间大于活动开始时间")
			return
		}
	} else if request.Type == lotmdl.GiftTypeMoney {
		// 现金白名单
		if s.c.Lottery.MoneyLimit != nil && len(s.c.Lottery.MoneyLimit) > 0 {
			var operateInLimit bool
			for _, v := range s.c.Lottery.MoneyLimit {
				if v == operator {
					operateInLimit = true
					break
				}
			}
			if !operateInLimit {
				err = ecode.Error(ecode.RequestErr, "没有配置现金的权限")
				log.Errorc(c, "operator(%s) err(%v)", operator, err)
				return
			}
		}
		if request.Params == "" {
			err = ecode.Error(ecode.RequestErr, "params为空，请检查后再次提交")
			return
		}
		params := lotmdl.GiftTypeMoneyParams{}
		if err = json.Unmarshal([]byte(request.Params), &params); err != nil {
			log.Errorc(c, "json.Unmarshal(action_add: %v) failed. error(%v)", request.Params, err)
			return
		}
	} else if request.Type == lotmdl.GiftTypeOGV {
		if request.Params == "" {
			err = ecode.Error(ecode.RequestErr, "params为空，请检查后再次提交")
			return
		}
		params := lotmdl.GiftTypeOGVParams{}
		if err = json.Unmarshal([]byte(request.Params), &params); err != nil {
			log.Errorc(c, "json.Unmarshal(action_add: %v) failed. error(%v)", request.Params, err)
			return
		}
		if params.Token == "" {
			err = ecode.Error(ecode.RequestErr, "Token为空")
			return
		}
	} else if request.Type == lotmdl.GiftTypeVipBuy {
		if request.Params == "" {
			err = ecode.Error(ecode.RequestErr, "params为空，请检查后再次提交")
			return
		}
		params := lotmdl.GiftTypeVipBuyParams{}
		if err = json.Unmarshal([]byte(request.Params), &params); err != nil {
			log.Errorc(c, "json.Unmarshal(action_add: %v) failed. error(%v)", request.Params, err)
			return
		}
		if params.Token == "" {
			err = ecode.Error(ecode.RequestErr, "Token为空")
			return
		}
	}
	// if request.Type==
	if tx, err = s.lotDao.BeginTran(c); err != nil {
		log.Errorc(c, "s.lotDao.BeginTran() failed. error(%v)", err)
		return
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			log.Errorc(c, "%v", r)
			return
		}
		if err != nil {
			if err1 := tx.Rollback(); err1 != nil {
				log.Errorc(c, "tx.Rollback() error(%v)", err1)
			}
			return
		}
		if err = tx.Commit(); err != nil {
			log.Errorc(c, "tx.Commit() error(%v)", err)
		}
	}()
	probability := request.GetIntProbability()
	if aid, err = s.lotDao.GiftAdd(tx, c, request.SID, request.Name, request.Source, request.MsgTitle, request.MsgContent,
		request.ImgURL, request.Params, request.MemberGroup, request.DayNum, request.Num, request.Type, probability, request.Extra, request.TimeLimit); err != nil {
		log.Errorc(c, "s.lotDao.GiftAdd(request: %+v) failed. error(%v)", request, err)
	}
	if err = s.lotDao.UpdateOperatorBySID(c, request.SID, operator); err != nil {
		log.Errorc(c, "s.lotDao.UpdateOperatorBySID(sid: %v, operator: %v) failed. error(%v)", request.SID, operator, err)
		return
	}
	key := lotmdl.GetTaskKey(request.SID, aid, request.Type)
	s.GiftTasks[key] = request.TimeLimit.Time().Unix()
	return
}

// GiftEdit update gift information
func (s *Service) GiftEdit(c context.Context, request *lotmdl.GiftEditParam, cookie, operator string) (err error) {
	var (
		tx      *sql.Tx
		gift    *lotmdl.GiftInfo
		lotInfo *lotmdl.LotInfo
		lottery *lotmdl.LotInfo
		coupon  *lotmdl.CouponInfo
	)
	err = s.numLimit(c, request.Num, operator)
	if err != nil {
		return err
	}
	if tx, err = s.lotDao.BeginTran(c); err != nil {
		log.Errorc(c, "s.lotDao.BeginTran() failed. error(%v)", err)
		return
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			log.Error("%v", r)
			err = ecode.Error(ecode.RequestErr, "编辑失败")
			return
		}
		if err != nil {
			if err1 := tx.Rollback(); err1 != nil {
				log.Errorc(c, "tx.Rollback() error(%v)", err1)
			}
			return
		}
		if err = tx.Commit(); err != nil {
			log.Errorc(c, "tx.Commit() error(%v)", err)
		}
	}()
	if err = request.GetDayStore(); err != nil {
		return err
	}
	if gift, err = s.lotDao.GiftDetailByID(c, request.ID); err != nil {
		log.Errorc(c, "s.lotDao.GiftDetailByID(%v) failed. error(%v)", request.ID, err)
		return
	}
	if request.Type == lotmdl.GiftTypeVIP {
		if request.Params == "" && request.Source == "" {
			err = ecode.Error(ecode.RequestErr, "params为空，请检查后再次提交")
			return
		}
		params := lotmdl.GiftTypeVIPParams{}
		if err = json.Unmarshal([]byte(request.Params), &params); err != nil {
			log.Errorc(c, "json.Unmarshal(action_add: %v) failed. error(%v)", request.Params, err)
			return
		}
		if params.Token == "" || params.AppKey == "" {
			err = ecode.Error(ecode.RequestErr, "Token 或者appKey为空")
			return
		}
		request.Source = params.Token
		var resourceInfo *vipresource.ResourceInfoReply
		resourceInfo, err = s.resourceClient.ResourceInfoByTokenV2(c, &vipresource.ResourceInfoByTokenV2Req{Token: params.Token, Appkey: params.AppKey})
		if err != nil {
			log.Errorc(c, "Failed to get resource info by token: %s, %s, %+v", params.Token, params.AppKey, err)
			return
		}
		if resourceInfo == nil || resourceInfo.Resource == nil || resourceInfo.Resource.ID <= 0 {
			err = ecode.Error(ecode.RequestErr, "params为空，请检查后再次提交")
			return
		}
		if lottery, err = s.lotDao.LotDetailBySID(c, request.SID); err != nil {
			log.Errorc(c, "s.lotDao.LotDetailBySID(%v) failed. error(%v)", request.SID, err)
			return
		}
		if lottery.ETime > resourceInfo.Resource.EndTime {
			err = ecode.Error(ecode.RequestErr, "大会员token有效期结束时间小于活动结束时间")
			return
		}
		if resourceInfo.Resource.StartTime > lottery.STime {
			err = ecode.Error(ecode.RequestErr, "大会员token有效开始时间大于活动开始时间")
			return
		}
	} else if request.Type == lotmdl.GiftTypeGrant {
		if request.Params == "" && request.Source == "" {
			err = ecode.Error(ecode.RequestErr, "params为空，请检查后再次提交")
			return
		}
		params := lotmdl.GiftTypeGrantParams{}
		if err = json.Unmarshal([]byte(request.Params), &params); err != nil {
			log.Errorc(c, "json.Unmarshal(action_add: %v) failed. error(%v)", request.Params, err)
			return
		}
		request.Source = request.Params
	} else if request.Type == lotmdl.GiftTypeCoin {
		if request.Params == "" && request.Source == "" {
			err = ecode.Error(ecode.RequestErr, "params为空，请检查后再次提交")
			return
		}
		params := lotmdl.GiftTypeCoinParams{}
		if err = json.Unmarshal([]byte(request.Params), &params); err != nil {
			log.Errorc(c, "json.Unmarshal(action_add: %v) failed. error(%v)", request.Params, err)
			return
		}
		request.Source = strconv.Itoa(params.Coin)
	} else if request.Type == lotmdl.GiftTypeCoupon {
		if request.Params == "" && request.Source == "" {
			err = ecode.Error(ecode.RequestErr, "params为空，请检查后再次提交")
			return
		}
		params := lotmdl.GiftTypeVipCouponParams{}
		if err = json.Unmarshal([]byte(request.Params), &params); err != nil {
			log.Errorc(c, "json.Unmarshal(action_add: %v) failed. error(%v)", request.Params, err)
			return
		}
		if params.Token == "" {
			err = ecode.Error(ecode.RequestErr, "Token为空")
			return
		}
		request.Source = params.Token
		if coupon, err = s.lotDao.GetCouponInfo(c, params.Token, cookie); err != nil {
			log.Errorc(c, "s.lotDao.GetCouponInfo(%v) failed. error(%v) ", request.Source, err)
			return
		}
		if coupon == nil || coupon.ID == 0 {
			log.Errorc(c, "GiftAdd coupon is empty.")
			err = ecode.Error(ecode.RequestErr, "token不存在，请检查输入的token是否正确")
			return
		}
		if lottery, err = s.lotDao.LotDetailBySID(c, request.SID); err != nil {
			log.Errorc(c, "s.lotDao.LotDetailBySID(%v) failed. error(%v)", request.SID, err)
			return
		}
		if lottery.ETime > coupon.ActivateEnd {
			err = ecode.Error(ecode.RequestErr, "优惠券token有效期小于活动结束时间")
			return
		} else if coupon.ActivateStart > lottery.STime {
			err = ecode.Error(ecode.RequestErr, "优惠券token有效开始时间大于活动开始时间")
			return
		}
	} else if request.Type == lotmdl.GiftTypeMoney {
		// 现金白名单
		if s.c.Lottery.MoneyLimit != nil && len(s.c.Lottery.MoneyLimit) > 0 {
			var operateInLimit bool
			for _, v := range s.c.Lottery.MoneyLimit {
				if v == operator {
					operateInLimit = true
					break
				}
			}
			if !operateInLimit {
				err = ecode.Error(ecode.RequestErr, "没有配置现金的权限")
				return
			}
		}
		if request.Params == "" {
			err = ecode.Error(ecode.RequestErr, "params为空，请检查后再次提交")
			return
		}
		params := lotmdl.GiftTypeMoneyParams{}
		if err = json.Unmarshal([]byte(request.Params), &params); err != nil {
			log.Errorc(c, "json.Unmarshal(action_add: %v) failed. error(%v)", request.Params, err)
			return
		}
	} else if request.Type == lotmdl.GiftTypeOGV {
		if request.Params == "" {
			err = ecode.Error(ecode.RequestErr, "params为空，请检查后再次提交")
			return
		}
		params := lotmdl.GiftTypeOGVParams{}
		if err = json.Unmarshal([]byte(request.Params), &params); err != nil {
			log.Errorc(c, "json.Unmarshal(action_add: %v) failed. error(%v)", request.Params, err)
			return
		}
		if params.Token == "" {
			err = ecode.Error(ecode.RequestErr, "Token为空")
			return
		}
	} else if request.Type == lotmdl.GiftTypeVipBuy {
		if request.Params == "" {
			err = ecode.Error(ecode.RequestErr, "params为空，请检查后再次提交")
			return
		}
		params := lotmdl.GiftTypeVipBuyParams{}
		if err = json.Unmarshal([]byte(request.Params), &params); err != nil {
			log.Errorc(c, "json.Unmarshal(action_add: %v) failed. error(%v)", request.Params, err)
			return
		}
		if params.Token == "" {
			err = ecode.Error(ecode.RequestErr, "Token为空")
			return
		}
	}
	if request.Type == lotmdl.GiftTypeSend && request.Effect == lotmdl.EffectY {
		if lotInfo, err = s.lotDao.LotDetailBySID(c, request.SID); err != nil {
			log.Errorc(c, "s.lotDao.LotDetailBySID(%v) failed. error(%v)", request.SID, err)
			return
		}
		var num int
		if num, err = s.lotDao.CountUpload(c, lotInfo.ID, gift.ID); err != nil {
			log.Errorc(c, "s.lotDao.CountUpload(lotID:%v, giftID:%v) failed. error(%v)", lotInfo.ID, gift.ID, err)
			return
		}
		if num < request.Num {
			err = ecode.Error(ecode.RequestErr, fmt.Sprintf("优惠券所上传的兑换码数量小于设置数量，无法进入奖池。当前上传数量 %d", num))
			return
		}
	}
	if request.LeastMark == lotmdl.GiftLeastMarkY {
		var (
			lmCheck []*lotmdl.GiftInfo
		)
		if lmCheck, err = s.lotDao.LeastMarkCheckList(c, request.SID); err != nil {
			log.Errorc(c, "s.lotDao.LeaskMarkCheckList(id: %v) failed. error(%v)", request.ID, err)
			return
		}
		for _, item := range lmCheck {
			if item.ID != request.ID {
				if _, err = s.lotDao.GiftEdit(tx, c, item.ID, item.Name, item.Source, item.MessageTitle, item.MessageContent, item.ImgURL, item.Params, item.MemberGroup, item.DayNum,
					item.Num, item.Type, item.IsShow, lotmdl.GiftLeastMarkN, item.Effect, item.ProbabilityI, item.Extra, item.TimeLimit); err != nil {
					log.Errorc(c, "s.lotDao.GiftEdit(%+v) failed. error(%v)", request, err)
					return
				}
			}
		}
	}

	probability := request.GetIntProbability()
	if _, err = s.lotDao.GiftEdit(tx, c, request.ID, request.Name, request.Source, request.MsgTitle, request.MsgContent, request.ImgURL, request.Params, request.MemberGroup, request.DayNum,
		request.Num, request.Type, request.IsShow, request.LeastMark, request.Effect, probability, request.Extra, request.TimeLimit); err != nil {
		log.Errorc(c, "s.lotDao.GiftEdit(%+v) failed. error(%v)", request, err)
		return
	}
	if err = s.lotDao.UpdateOperatorBySID(c, request.SID, operator); err != nil {
		log.Errorc(c, "s.lotDao.UpdateOperatorBySID(sid: %v, operator:%v) failed. error(%v)", request.SID, operator, err)
		return
	}
	if request.TimeLimit.Time().Unix() > xtime.Now().Unix() {
		key := lotmdl.GetTaskKey(request.SID, request.ID, request.Type)
		s.GiftTasks[key] = request.TimeLimit.Time().Unix()
	}
	return
}

// MemberGroupEdit update membergroup information
func (s *Service) MemberGroupEdit(c context.Context, request *lotmdl.MemberGroupEditParam, cookie, operator string) (err error) {
	var (
		tx *sql.Tx
	)

	if tx, err = s.lotDao.BeginTran(c); err != nil {
		log.Errorc(c, "s.lotDao.BeginTran() failed. error(%v)", err)
		return
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			log.Errorc(c, "%v", r)
			err = ecode.Error(ecode.RequestErr, "编辑失败")

			return
		}
		if err != nil {
			if err1 := tx.Rollback(); err1 != nil {
				log.Error("tx.Rollback() error(%v)", err1)
			}
			return
		}
		if err = tx.Commit(); err != nil {
			log.Error("tx.Commit() error(%v)", err)
		}
	}()
	memberGroup := make([]*lotmdl.MemberGroupDB, 0)
	memberGroup = append(memberGroup, &lotmdl.MemberGroupDB{
		ID:    request.ID,
		SID:   request.SID,
		Name:  request.Name,
		Group: request.Group,
		State: request.State,
	})
	if err = s.updateOrInsertMemberGroup(c, tx, request.SID, memberGroup); err != nil {
		log.Errorc(c, "s.updateOrInsertMemberGroup(sid: %v, memberGroup:%v) failed. error(%v)", request.SID, memberGroup, err)
		return
	}
	// if err = s.lotDao.DeleteMemberGroup(c, request.SID); err != nil {
	// 	log.Errorc(c, "s.lotDao.DeleteMemberGroup(sid: %v, operator:%v) failed. error(%v)", request.SID, operator, err)
	// 	return
	// }
	if err = s.lotDao.UpdateOperatorBySID(c, request.SID, operator); err != nil {
		log.Errorc(c, "s.lotDao.UpdateOperatorBySID(sid: %v, operator:%v) failed. error(%v)", request.SID, operator, err)
		return
	}
	return
}

// GiftList get gift list
func (s *Service) GiftList(c context.Context, request *lotmdl.GiftListParam) (rsp *lotmdl.GiftList, err error) {
	var (
		giftList []*lotmdl.GiftInfo
		page     = lotmdl.Page{}
		//giftNum  map[int64]int
	)
	if page.Total, err = s.lotDao.GiftTotal(c, request.SID, request.State, request.Type); err != nil {
		log.Errorc(c, "s.lot.GiftTotal() failed. error(%v)", err)
		return
	}
	page.Num = request.Pn
	page.Size = request.Ps
	if giftList, err = s.lotDao.GiftList(c, request.SID, request.Rank, request.State, request.Type, request.Pn, request.Ps); err != nil {
		log.Errorc(c, "s.lotDao.GiftList() failed. error(%v)", err)
		return
	}
	for k, v := range giftList {
		v.DBNum = v.Num
		v.RedisNum = v.Num - v.SendNum
		giftList[k].ProbabilityF = giftList[k].GetFloatProbability()
		giftList[k].Params = s.getParams(c, v.Source, v.Params, v.Type)
	}
	rsp = &lotmdl.GiftList{}
	rsp.List = giftList
	rsp.Page = page
	return
}

// MemberGroupList memberGroup list
func (s *Service) MemberGroupList(c context.Context, request *lotmdl.MemberGroupListParam) (rsp *lotmdl.MemberGroupListReply, err error) {
	var (
		memberGroupList []*lotmdl.MemberGroupDB
		page            = lotmdl.Page{}
	)
	if page.Total, err = s.lotDao.MemberGroupTotal(c, request.SID, request.State); err != nil {
		log.Errorc(c, "s.lot.GiftTotal() failed. error(%v)", err)
		return
	}
	page.Num = request.Pn
	page.Size = request.Ps
	if memberGroupList, err = s.lotDao.MemberGroupList(c, request.SID, request.Rank, request.State, request.Pn, request.Ps); err != nil {
		log.Errorc(c, "s.lotDao.MemberGroupList() failed. error(%v)", err)
		return
	}
	rsp = &lotmdl.MemberGroupListReply{}
	rsp.List = memberGroupList
	rsp.Page = page
	return
}

// GiftWinList get gift win list
func (s *Service) GiftWinList(c context.Context, request *lotmdl.GiftWinListParam) (rsp *lotmdl.GiftWinList, err error) {
	var (
		lotInfo *lotmdl.LotInfo
		giftWin []*lotmdl.GiftWinInfo
		page    = lotmdl.Page{}
	)
	if lotInfo, err = s.lotDao.LotDetailBySID(c, request.SID); err != nil {
		log.Errorc(c, "s.lotDao.LotDetailBySID() failed. error(%v)", err)
		return
	}
	if lotInfo == nil {
		err = ecode.Error(ecode.RequestErr, "未找到相应抽奖信息")
		return
	}
	if page.Total, err = s.lotDao.GiftWinTotal(c, lotInfo.ID, request.ID); err != nil {
		log.Errorc(c, "s.lotDao.GiftWinTotal() failed. error(%v)", err)
		return
	}
	page.Num = request.Pn
	page.Size = request.Ps
	if giftWin, err = s.lotDao.GiftWinList(c, lotInfo.ID, request.ID, request.Pn, request.Ps); err != nil {
		log.Errorc(c, "s.lotDao.GiftWinList() failed. error(%v)", err)
		return
	}
	for i, item := range giftWin {
		addrTmp := &lotmdl.Address{}
		if item.GiftAddrID != 0 {
			if addrTmp, err = s.lotDao.GetAddressByID(c, item.GiftAddrID, item.Mid); err != nil {
				log.Errorc(c, "s.lotDao.GetAddressByID() failed. error(%v)", err)
				return
			}
			giftWin[i].Addr = *addrTmp
		}
	}
	rsp = &lotmdl.GiftWinList{}
	rsp.List = giftWin
	rsp.Page = page
	return
}

// GiftUpload upload keys INSERT act_lottery_gift_win batch
func (s *Service) GiftUpload(c context.Context, keys []string, aid int64, sid, operator string) {
	var (
		tx      *sql.Tx
		lotInfo *lotmdl.LotInfoDraft
		key     = lotmdl.GetUploadKey(sid, aid)
		err     error
		gift    *lotmdl.GiftInfo
	)
	if tx, err = s.lotDao.BeginTran(c); err != nil {
		log.Errorc(c, "s.lotDao.BeginTran() failed. error(%v)", err)
		return
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			log.Error("%v", r)
			err = ecode.Error(ecode.RequestErr, "上传失败")

			return
		}
		if err != nil {
			if err1 := tx.Rollback(); err1 != nil {
				log.Error("tx.Rollback() error(%v)", err1)
			}
			return
		}
		if err = tx.Commit(); err != nil {
			log.Error("tx.Commit() error(%v)", err)
		}
	}()
	if lotInfo, err = s.lotDao.LotDraftDetailBySID(c, sid); err != nil {
		log.Errorc(c, "s.lotDao.LotDraftDetailBySID(sid: %v) failed. error(%v)", sid, err)
		s.setUploadInfo(key, lotmdl.UploadFailed)
		return
	}
	total := len(keys)
	keyInsert := []string{}
	for i, item := range keys {
		keyInsert = append(keyInsert, item)
		if (i%2000 == 0) || (i == (total - 1)) {
			if err = s.lotDao.GiftUpload(tx, c, lotInfo.ID, aid, keyInsert); err != nil {
				log.Error("GiftUpload message s.lotDao.GiftUpload() failed. error(%v)", err)
				s.setUploadInfo(key, lotmdl.UploadFailed)
				return
			}
			keyInsert = []string{}
		}
	}
	if gift, err = s.lotDao.GiftDraftDetailByID(c, aid); err != nil {
		log.Error("GiftUpload s.lotDao.GiftDraftDetailByID(%v) failed. error(%v)", aid, err)
		s.setUploadInfo(key, lotmdl.UploadFailed)
		return
	}
	if err = s.lotDao.UpdateOperatorBySID(c, sid, operator); err != nil {
		log.Error("s.lotDao.UpdateOperatorBySID(sid: %v, operator: %v) failed. error(%v)", sid, operator, err)
		s.setUploadInfo(key, lotmdl.UploadFailed)
		return
	}
	taskKey := lotmdl.GetTaskKey(sid, aid, gift.Type)
	s.GiftTasks[taskKey] = gift.TimeLimit.Time().Unix()
	s.setUploadInfo(key, lotmdl.UploadSuccess)
	return
}

func (s *Service) setUploadInfo(key string, status int) {
	if infoTmp, ok := s.UploadInfo[key]; ok && infoTmp != nil {
		s.UploadInfo[key].Status = status
	} else {
		s.UploadInfo[key] = &lotmdl.UploadInfo{Status: status}
	}
}

// UpdUploadStatus update lottery_gift upload .
func (s *Service) UpdUploadStatus(c context.Context, status int, id int64) (err error) {
	if err = s.lotDao.UploadStatusUpdate(c, status, id); err != nil {
		log.Errorc(c, "s.lotDao.UploadStatusUpdate() failed. error(%v)", err)
	}
	return
}

// UpdUploadStatusDraft update lottery_gift upload .
func (s *Service) UpdUploadStatusDraft(c context.Context, status int, id int64) (err error) {
	if err = s.lotDao.UploadStatusUpdateDraft(c, status, id); err != nil {
		log.Errorc(c, "s.lotDao.UploadStatusUpdateDraft() failed. error(%v)", err)
	}
	return
}

// GiftExport export gift win list
func (s *Service) GiftExport(c context.Context, aid int64, sid string) (result [][]string, err error) {
	var (
		lotInfo  *lotmdl.LotInfo
		giftInfo *lotmdl.GiftInfo
		giftWin  []*lotmdl.GiftWinInfo
	)
	if lotInfo, err = s.lotDao.LotDetailBySID(c, sid); err != nil {
		log.Errorc(c, "s.lotDao.LotDetailBySID() failed. error(%v)", err)
		return
	}
	if lotInfo == nil {
		err = ecode.Error(ecode.RequestErr, "未找到相应抽奖信息")
		return
	}
	if giftWin, err = s.lotDao.GiftWinListAll(c, lotInfo.ID, aid); err != nil {
		log.Errorc(c, "s.lotDao.GiftWinList() failed. error(%v)", err)
		return
	}
	if giftInfo, err = s.lotDao.GiftDetailByID(c, aid); err != nil {
		log.Errorc(c, "s.lotDao.GiftDetailByID() failed. error(%v)", err)
		return
	}
	for _, item := range giftWin {
		addrTmp := &lotmdl.Address{}
		if item.GiftAddrID != 0 {
			if addrTmp, err = s.lotDao.GetAddressByID(c, item.GiftAddrID, item.Mid); err != nil {
				log.Errorc(c, "s.lotDao.GetAddressByID() failed. error(%v)", err)
				return
			}
		}
		addr := addrTmp.Prov + " " + addrTmp.City + " " + addrTmp.Area + " " + addrTmp.Addr
		winTime := item.MTime.Time()
		winTimeStr := fmt.Sprintf("%d-%d-%d %d:%d:%d", winTime.Year(), winTime.Month(), winTime.Day(), winTime.Hour(), winTime.Minute(), winTime.Second())
		result = append(result, []string{strconv.Itoa(item.Mid), addrTmp.Name, giftInfo.Name, giftTypeToStr[giftInfo.Type],
			addr, addrTmp.Phone, winTimeStr})
	}
	return
}

// FixLotteryGiftTask
func (s *Service) FixLotteryGiftTask(c context.Context) (err error) {
	var task []*lotmdl.GiftTask
	s.GiftTaskLock.Lock()
	defer s.GiftTaskLock.Unlock()
	if task, err = s.lotDao.GiftTaskCheck(c); err != nil {
		log.Error("activity-admin lottery runGiftTasks init failed. error(%v)", err)
	} else {
		for _, item := range task {
			key := lotmdl.GetTaskKey(item.SID, item.ID, item.Type)
			s.GiftTasks[key] = item.TimeLimit.Time().Unix()
		}
	}
	return
}

// VipCheck check vip id
func (s *Service) VipCheck(c context.Context, vipID, cookie string) (rsp lotmdl.CheckRsp, err error) {
	var vipCheck bool
	if vipCheck, err = s.lotDao.GetVIPInfo(c, vipID, cookie); err != nil {
		log.Errorc(c, "s.lotDao.GetVIPInfo(%v) failed. error(%v)", vipID, err)
		err = ecode.Error(ecode.RequestErr, "大会员套餐ID校验失败")
		return
	}
	if !vipCheck {
		log.Errorc(c, "vip config bad. info: %v", vipID)
		rsp.Check = 2
		return
	}
	rsp.Check = 1
	return
}

// BatchAddTimes batch add lottery times
func (s *Service) BatchAddTimes(c context.Context, sid string, times int64, mids []int64) error {
	lottery, err := s.lotDao.LotDetailBySID(c, sid)
	if err != nil {
		log.Errorc(c, "s.lotDao.LotDetailBySID(%v) failed. error(%v)", sid, err)
		return ecode.Error(ecode.RequestErr, "找不到指定的抽奖id")
	}
	if lottery == nil || lottery.ID == 0 {
		return ecode.Error(ecode.RequestErr, "抽奖id为空")
	}
	timeConf := make([]*lotmdl.TimesConf, 0)
	if timeConf, err = s.lotDao.AllTimesConf(c, sid); err != nil {
		log.Errorc(c, "s.lotDao.AllTimesConf(%v) failed. error(%v)", sid, err)
		return err
	}
	var cid int64
	for _, v := range timeConf {
		if v.Type == lotmdl.AddActionTypeOther {
			cid = v.ID
		}
	}
	confTimes, err := s.lotDao.RawTimesByID(c, cid)
	if err != nil {
		log.Errorc(c, "s.lotDao.RawTimesByID(%v) failed. error(%v)", sid, err)
		return ecode.Error(ecode.RequestErr, "查询配置失败")
	}
	if confTimes < times {
		return ecode.Error(ecode.RequestErr, "增加次数超过配置")
	}
	if len(mids) > lotmdl.MaxMidLen {
		return ecode.Error(ecode.RequestErr, "上传的mid超过限制")
	}
	go func() {
		for i := 1; i <= 1000; i++ {
			orderNo := strconv.FormatInt(xtime.Now().UnixNano()/1000000, 10)
			if len(mids) > i*1000 {
				if err = s.lotDao.BatchAddLotTimes(context.Background(), lottery.ID, times, cid, mids[(i-1)*1000:i*1000], orderNo); err != nil {
					log.Error("BatchAddTimes s.lotDao.BacthAddLotTimes(%v) error(%v)", mids[(i-1)*1000], err)
				}
				xtime.Sleep(xtime.Millisecond * 100)
				continue
			}
			if err = s.lotDao.BatchAddLotTimes(context.Background(), lottery.ID, times, cid, mids[(i-1)*1000:], orderNo); err != nil {
				log.Error("BatchAddTimes s.lotDao.BacthAddLotTimes(%v) error(%v)", mids[(i-1)*1000], err)
			}
			break
		}
	}()
	go func() {
		for _, v := range mids {
			if err = s.lotDao.DeleteLotteryTimesCache(context.Background(), lottery.ID, v); err != nil {
				log.Error("BatchAddTimes s.lotDao.DeleteLotteryTimesCache(%d) error(%v)", v, err)
			}
			xtime.Sleep(xtime.Millisecond * 20)
		}
	}()
	return nil
}

// GiftExportAll export all gift win list
func (s *Service) GiftExportAll(c context.Context, sid string) (result [][]string, err error) {
	var (
		lotInfo  *lotmdl.LotInfo
		giftWin  []*lotmdl.GiftWinInfo
		giftInfo map[int64]*lotmdl.GiftInfo
	)
	if lotInfo, err = s.lotDao.LotDetailBySID(c, sid); err != nil {
		log.Errorc(c, "s.lotDao.LotDetailBySID() failed. error(%v)", err)
		return
	}
	if lotInfo == nil {
		log.Errorc(c, "did not find sid sid(%s)", sid)

		return
	}
	if giftInfo, err = s.lotDao.GiftDetailBySid(c, sid); err != nil {
		log.Errorc(c, "s.lotDao.GiftDetailBySid() failed. error(%v)", err)
		return
	}
	if giftWin, err = s.lotDao.GiftWinListWithoutAid(c, lotInfo.ID); err != nil {
		log.Errorc(c, "s.lotDao.GiftWinListWithoutAid() failed. error(%v)", err)
		return
	}
	for _, item := range giftWin {
		addrTmp := &lotmdl.Address{}
		if item.GiftAddrID != 0 {
			if addrTmp, err = s.lotDao.GetAddressByID(c, item.GiftAddrID, item.Mid); err != nil {
				log.Errorc(c, "s.lotDao.GetAddressByID() failed. error(%v)", err)
				return
			}
		}
		addr := addrTmp.Prov + " " + addrTmp.City + " " + addrTmp.Area + " " + addrTmp.Addr
		winTime := item.MTime.Time()
		winTimeStr := fmt.Sprintf("%d-%d-%d %d:%d:%d", winTime.Year(), winTime.Month(), winTime.Day(), winTime.Hour(), winTime.Minute(), winTime.Second())
		result = append(result, []string{strconv.Itoa(item.Mid), addrTmp.Name, strconv.FormatInt(item.GiftId, 10), giftInfo[item.GiftId].Name, giftTypeToStr[giftInfo[item.GiftId].Type], addr, addrTmp.Phone, winTimeStr})
	}
	return
}
