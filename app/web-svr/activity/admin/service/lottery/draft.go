package lottery

import (
	"context"
	"encoding/json"
	"fmt"
	"go-common/library/database/sql"
	xsql "go-common/library/database/sql"
	"go-common/library/ecode"
	"go-common/library/log"
	"go-common/library/sync/errgroup.v2"
	"go-gateway/app/web-svr/activity/admin/component"
	componentmdl "go-gateway/app/web-svr/activity/admin/model/component"
	"time"

	lotmdl "go-gateway/app/web-svr/activity/admin/model/lottery"
	"strconv"
	"strings"
	xtime "time"

	vipresource "git.bilibili.co/bapis/bapis-go/vip/resource/service"
)

// AddDraft add lottery base information.
func (s *Service) AddDraft(c context.Context, request *lotmdl.AddParam, operator string) (err error) {
	var (
		tx      *sql.Tx
		id      int64
		lotInfo *lotmdl.LotInfoDraft
	)
	if tx, err = s.lotDao.BeginTran(c); err != nil {
		log.Errorc(c, "s.lotDao.BeginTran() failed. error(%v)", err)
		return
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			log.Errorc(c, "%v", r)
			err = ecode.Error(ecode.RequestErr, "新增抽奖草稿失败")
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
			return
		}
		return
	}()
	if id, err = s.lotDao.CreateDraft(c, tx, request.Name, operator, request.Stime, request.Etime, lotmdl.LotteryDraftStateDraft, request.Type); err != nil {
		log.Errorc(c, "Add s.lotDao.CreateDraft(%v, %v, %v, %v) failed. error(%v)", request.Name, request.Stime, request.Etime, request.Type, err)
		return
	}
	if lotInfo, err = s.lotDao.LotDraftDetailTxByID(c, tx, id); err != nil {
		log.Errorc(c, "Add s.lotDao.LotDraftDetailTxByID() failed. error(%v)", err)
		return
	}
	if err = s.lotDao.InitLotDetailDraft(c, tx, lotInfo.LotteryID); err != nil {
		log.Errorc(c, "Add s.lotDao.InitLotDetailDraft() failed. error(%v)", err)
		return
	}
	if err = s.lotDao.CreateWin(tx, id); err != nil {
		log.Error("lottery@Add d.CreateWin(%d) failed. error(%v)", id, err)
		return
	}
	return
}

func (s *Service) draftMidUpdateState(c context.Context, state int) error {
	if state == lotmdl.LotteryDraftStateWaitReviewed || state == lotmdl.LotteryDraftStateWaitReject {
		return ecode.Error(ecode.RequestErr, "没有权限修改为审核或拒绝")
	}
	return nil
}

func (s *Service) draftStateCanUpdate(c context.Context, lotInfo *lotmdl.LotInfoDraft) error {
	if lotInfo == nil {
		return ecode.Error(ecode.RequestErr, "抽奖不存在")
	}
	if lotInfo.State == lotmdl.LotteryDraftStateWaitReview {
		return ecode.Error(ecode.RequestErr, "待审中，不可修改")
	}
	if lotInfo.State == lotmdl.LotteryDraftStateOffline {
		return ecode.Error(ecode.RequestErr, "已下线，不可修改")
	}
	return nil
}

// EditDraft 编辑抽奖
func (s *Service) EditDraft(c context.Context, request *lotmdl.EditParam, cookie, operator string) (err error) {
	var (
		tx            *sql.Tx
		list          = &lotmdl.LotInfoDraft{}
		rule          *lotmdl.RuleInfo
		timesUpdate   []*lotmdl.TimesConf
		timesAdd      []*lotmdl.TimesConf
		baseConf      = &lotmdl.BaseTimes{}
		actionAddConf = make([]*lotmdl.AddTimes, 0)
		vipCheck      bool
	)
	if request.SenderMid == 0 {
		if s.c.Lottery.SenderMidLimit == nil || len(s.c.Lottery.SenderMidLimit) == 0 {
			err = ecode.Error(ecode.RequestErr, "抽奖信息发布账号必填")
			return
		}
		var operateInLimit bool
		for _, v := range s.c.Lottery.SenderMidLimit {
			if v == operator {
				operateInLimit = true
				break
			}
		}
		if !operateInLimit {
			err = ecode.Error(ecode.RequestErr, "抽奖信息发布账号必填")
			log.Errorc(c, "operator(%s) err(%v)", operator, err)
			return
		}
	}
	if err = s.draftMidUpdateState(c, request.State); err != nil {
		return
	}
	if tx, err = s.lotDao.BeginTran(c); err != nil {
		log.Errorc(c, "s.lotDao.BeginTran() failed. error(%v)", err)
		return
	}
	if list, err = s.lotDao.LotDraftDetailBySID(c, request.SID); err != nil {
		log.Errorc(c, "s.lotDao.LotDetailByID(id: %v) failed. error(%v)", request.ID, err)
		return
	}
	//  验证当前抽奖是否可修改
	if err = s.draftStateCanUpdate(c, list); err != nil {
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
				log.Error("tx.Rollback() error(%v)", err1)
			}
			return
		}
		if err = tx.Commit(); err != nil {
			log.Error("tx.Commit() error(%v)", err)
		}
	}()

	// 有资格审核人
	canReviewer := strings.Join(s.c.Lottery.Reviewers, ",")
	if err = s.lotDao.UpdateLotDraftInfo(c, tx, list.ID, request.IsInternal, request.State, request.Name, operator, canReviewer, request.Stime, request.Etime); err != nil {
		log.Errorc(c, "s.lotDao.UpdateLotInfo(id: %v,is_internal:%d name: %v, stime: %v, etime: %v) failed. error(%v)",
			list.ID, request.IsInternal, request.Name, request.Stime, request.Etime, err)
		return
	}
	if rule, err = s.lotDao.GetLotRuleDraftBySID(c, list.LotteryID); err != nil {
		log.Errorc(c, "s.lotDao.GetLotRuleBySID(sid: %v) failed. error(%v)", list.LotteryID, err)
		return
	}
	var coinCount int
	var likeCount int
	if request.ActionAdd != "" {
		if err = json.Unmarshal([]byte(request.ActionAdd), &actionAddConf); err != nil {
			log.Errorc(c, "json.Unmarshal(action_add: %v) failed. error(%v)", request.ActionAdd, err)
			return
		}
		additionalMap := make(map[int64]struct{})
		for _, item := range actionAddConf {
			if item.Status != 2 {
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
				if item.Type == lotmdl.AddActionTypeCustom || item.Type == lotmdl.AddActionTypeOGV || item.Type == lotmdl.AddActionTypeAct {
					if item.Info == "" {
						err = ecode.Error(ecode.RequestErr, "未填写自定义行为ID")
						return
					}
					if item.Type == lotmdl.AddActionTypeAct {
						infoSplit := strings.Split(item.Info, ".")
						if len(infoSplit) != 2 {
							err = ecode.Error(ecode.RequestErr, "活动id及活动行为填写有误")
							return
						}
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
				if item.Type == lotmdl.AddActionTypeLike && item.Status != 2 {
					likeCount++
				}
				if item.Type == lotmdl.AddActionTypeCoin && item.Status != 2 {
					coinCount++
				}
				if likeCount > 1 {
					err = ecode.Error(ecode.RequestErr, "点赞获取抽奖次数只能配置1个")
					return
				}
				if coinCount > 1 {
					err = ecode.Error(ecode.RequestErr, "投币获取抽奖次数只能配置1个")
					return
				}
				if item.Type == lotmdl.AddActionTypeLike || item.Type == lotmdl.AddActionTypeCoin {
					if item.Info == "" {
						err = ecode.Error(ecode.RequestErr, "未填写活动id及counter，请在预约数据源配置")
						return
					}
					info := &lotmdl.CoinLikeInfo{}
					err = json.Unmarshal([]byte(item.Info), info)
					if err != nil {
						err = ecode.Error(ecode.RequestErr, "请正确填写活动id及counter，请在预约数据源配置")
						return
					}
					if info.Activity == "" || info.Counter == "" || info.Count == 0 {
						err = ecode.Error(ecode.RequestErr, "请正确填写活动id及counter，请在预约数据源配置")
						return
					}

				}
				if item.Type == lotmdl.AddActionTypeAdditional {
					if item.Info == "" {
						err = ecode.Error(ecode.RequestErr, "未填写消耗x次，赠送x次的配置信息")
						return
					}
					info := &lotmdl.ConsumeInfo{}
					err = json.Unmarshal([]byte(item.Info), info)
					if err != nil {
						err = ecode.Error(ecode.RequestErr, "请填写消耗x次，赠送x次的配置信息")
						return
					}
					if info.Consume == 0 || info.Send == 0 {
						err = ecode.Error(ecode.RequestErr, "请填写消耗x次，赠送x次的配置信息")
						return
					}
					if item.Status != 2 {
						if _, ok := additionalMap[info.Consume]; ok {
							err = ecode.Error(ecode.RequestErr, "相同消耗次数只能配置一次")
							return
						}
						additionalMap[info.Consume] = struct{}{}
					}
				}
				if item.Type == lotmdl.AddActionTypeTaskPoint {
					if item.Info == "" {
						err = ecode.Error(ecode.RequestErr, "未填写预约数据源id和节点组")
						return
					}
					info := &lotmdl.ActivityInfo{}
					err = json.Unmarshal([]byte(item.Info), info)
					if err != nil {
						err = ecode.Error(ecode.RequestErr, "未填写预约数据源id和节点组")
						return
					}
					if info.GroupID == 0 || info.Sid == 0 {
						err = ecode.Error(ecode.RequestErr, "未填写预约数据源id和节点组")
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
	rule.FigureScore = request.FigureScore
	rule.SpyScore = request.SpyScore
	if _, err = s.lotDao.RuleDraftUpdate(c, tx, rule); err != nil {
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
		if _, err = s.lotDao.TimesDraftAddBatch(c, tx, timesAdd); err != nil {
			log.Errorc(c, "s.lotDao.TimesAddBatch(timesAdd: %+v), error(%v)", timesAdd, err)
			return
		}
	}
	if len(timesUpdate) != 0 {
		if _, err = s.lotDao.TimesDraftUpdateBatch(c, tx, timesUpdate); err != nil {
			log.Errorc(c, "s.lotDao.TimesDraftUpdateBatch(timesUpdate: %+v), error(%v)", timesUpdate, err)
		}
	}
	// 如果提交审核，则需要通知审核人员
	if request.State == lotmdl.LotteryDraftStateWaitReview {
		// 通知审核人
		to := make([]*componentmdl.Address, 0)
		to = append(to, &componentmdl.Address{Address: fmt.Sprintf("%s@bilibili.com", operator), Name: operator})
		if len(s.c.Lottery.Reviewers) > 0 {
			for _, v := range s.c.Lottery.Reviewers {
				to = append(to, &componentmdl.Address{Address: fmt.Sprintf("%s@bilibili.com", v), Name: v})
			}
		}
		wechatTo := s.c.Lottery.Reviewers
		wechatTo = append(wechatTo, operator)
		err = s.sendMessageWechatEmail(c, wechatTo, to, to, to, s.c.Lottery.AuditSubject, s.messageTextBuild(c, request.SID))
		if err != nil {
			log.Errorc(c, "s.sendMessageWechatEmail err(%v)", err)
		}
	}
	return
}

func (s *Service) messageTextBuild(c context.Context, sid string) string {
	return fmt.Sprintf("抽奖待审，sid:%s，链接：%s%s", sid, s.c.Lottery.EditLink, sid)
}

// SendMessageWechatEmail 发送微信和邮箱
func (s *Service) sendMessageWechatEmail(c context.Context, user []string, to, cc, bcc []*componentmdl.Address, subject, message string) error {
	eg := errgroup.WithContext(c)
	eg.Go(func(ctx context.Context) (err error) {
		err = s.sendEmail(c, subject, message, to, cc, bcc)
		return
	})
	eg.Go(func(ctx context.Context) (err error) {
		if len(user) > 0 {
			err = s.sendWechat(c, subject, message, strings.Join(user, ","))
		}
		return
	})
	if err := eg.Wait(); err != nil {
		log.Errorc(c, "eg.Wait error(%v)", err)
		return err
	}
	return nil
}

// sendEmail 发送邮件
func (s *Service) sendEmail(c context.Context, subject, message string, to, cc, bcc []*componentmdl.Address) (err error) {
	base := &componentmdl.Base{
		Host:    s.maiInfo.Host,
		Port:    s.maiInfo.Port,
		Address: s.maiInfo.Address,
		Pwd:     s.maiInfo.Pwd,
		Name:    s.maiInfo.Name,
	}
	mail := &componentmdl.Mail{
		ToAddresses:  to,
		CcAddresses:  cc,
		BccAddresses: bcc,
		Subject:      subject,
		Type:         componentmdl.TypeTextHTML,
		Body:         message,
	}
	err = component.SendMail(mail, base, nil)
	if err != nil {
		log.Errorc(c, "s.dao.SendMail error(%v)", err)
	}
	return
}

// sendWechat 发送微信
func (s *Service) sendWechat(c context.Context, title, message, user string) (err error) {
	err = s.dao.SendWeChat(c, s.c.Lottery.PublicKey, title, message, user)
	if err != nil {
		log.Errorc(c, "s.dao.SendWechat error(%v)", err)
	}
	return
}

// GiftAddDraft add gift information
func (s *Service) GiftAddDraft(c context.Context, request *lotmdl.GiftAddParam, cookie, operator string) (err error) {
	var (
		tx      *sql.Tx
		lottery *lotmdl.LotInfoDraft
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
	if lottery, err = s.lotDao.LotDraftDetailBySID(c, request.SID); err != nil {
		log.Errorc(c, "s.lotDao.LotDetailBySID(%v) failed. error(%v)", request.SID, err)
		return
	}
	//  验证当前抽奖是否可修改
	if err = s.draftStateCanUpdate(c, lottery); err != nil {
		return
	}
	if len(request.Extra) > lotmdl.ExtraLengthMax {
		err = ecode.Error(ecode.RequestErr, "额外参数太长")
		return
	}
	extra := make(map[string]string)
	if err = json.Unmarshal([]byte(request.Extra), &extra); err != nil {
		log.Errorc(c, "json.Unmarshal(Extra: %v) failed. error(%v)", request.Extra, err)
		err = ecode.Error(ecode.RequestErr, "extra参数有误")
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
			log.Error("Failed to get resource info by token: %s, %s, %+v", params.Token, params.AppKey, err)
			return
		}
		if resourceInfo == nil || resourceInfo.Resource == nil || resourceInfo.Resource.ID <= 0 {
			err = ecode.Error(ecode.RequestErr, "params为空，请检查后再次提交")
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
	} else if request.Type == lotmdl.GiftTypeAward {
		if request.Params == "" {
			err = ecode.Error(ecode.RequestErr, "params为空，请检查后再次提交")
			return
		}
		params := lotmdl.GiftTypeAwardParams{}
		if err = json.Unmarshal([]byte(request.Params), &params); err != nil {
			log.Errorc(c, "json.Unmarshal(action_add: %v) failed. error(%v)", request.Params, err)
			return
		}
		if params.AwardID == 0 {
			err = ecode.Error(ecode.RequestErr, "awardID为空")
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
			log.Error("%v", r)
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
	probability := request.GetIntProbability()
	if aid, err = s.lotDao.GiftDraftAdd(c, tx, request.SID, request.Name, request.Source, request.MsgTitle, request.MsgContent,
		request.ImgURL, request.Params, request.MemberGroup, request.DayNum, request.Num, request.Type, probability, request.Extra, request.TimeLimit); err != nil {
		log.Errorc(c, "s.lotDao.GiftDraftAdd(request: %+v) failed. error(%v)", request, err)
	}
	if err = s.lotDao.UpdateOperatorDraftBySIDAndState(c, request.SID, operator, lotmdl.LotteryDraftStateDraft); err != nil {
		log.Errorc(c, "s.lotDao.UpdateOperatorDraftBySIDAndState(sid: %v, operator: %v) failed. error(%v)", request.SID, operator, err)
		return
	}
	key := lotmdl.GetTaskKey(request.SID, aid, request.Type)
	s.GiftTasks[key] = request.TimeLimit.Time().Unix()
	return
}

// GiftEditDraft update gift information
func (s *Service) GiftEditDraft(c context.Context, request *lotmdl.GiftEditParam, cookie, operator string) (err error) {
	var (
		tx      *sql.Tx
		gift    *lotmdl.GiftInfo
		lotInfo *lotmdl.LotInfoDraft
		lottery *lotmdl.LotInfoDraft
		coupon  *lotmdl.CouponInfo
	)
	err = s.numLimit(c, request.Num, operator)
	if err != nil {
		return err
	}
	if lottery, err = s.lotDao.LotDraftDetailBySID(c, request.SID); err != nil {
		log.Errorc(c, "s.lotDao.LotDraftDetailBySID(%v) failed. error(%v)", request.SID, err)
		return
	}
	if len(request.Extra) > lotmdl.ExtraLengthMax {
		err = ecode.Error(ecode.RequestErr, "额外参数太长")
		return
	}
	extra := make(map[string]string)
	if err = json.Unmarshal([]byte(request.Extra), &extra); err != nil {
		log.Errorc(c, "json.Unmarshal(Extra: %v) failed. error(%v)", request.Extra, err)
		err = ecode.Error(ecode.RequestErr, "extra参数有误")
		return
	}
	//  验证当前抽奖是否可修改
	if err = s.draftStateCanUpdate(c, lottery); err != nil {
		return
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
				log.Error("tx.Rollback() error(%v)", err1)
			}
			return
		}
		if err = tx.Commit(); err != nil {
			log.Error("tx.Commit() error(%v)", err)
		}
	}()
	if err = request.GetDayStore(); err != nil {
		return err
	}
	if gift, err = s.lotDao.GiftDraftDetailByID(c, request.ID); err != nil {
		log.Errorc(c, "s.lotDao.GiftDraftDetailByID(%v) failed. error(%v)", request.ID, err)
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
	} else if request.Type == lotmdl.GiftTypeAward {
		if request.Params == "" {
			err = ecode.Error(ecode.RequestErr, "params为空，请检查后再次提交")
			return
		}
		params := lotmdl.GiftTypeAwardParams{}
		if err = json.Unmarshal([]byte(request.Params), &params); err != nil {
			log.Errorc(c, "json.Unmarshal(action_add: %v) failed. error(%v)", request.Params, err)
			return
		}
		if params.AwardID == 0 {
			err = ecode.Error(ecode.RequestErr, "awardID为空")
			return
		}
	}
	if request.Type == lotmdl.GiftTypeSend && request.Effect == lotmdl.EffectY {
		if lotInfo, err = s.lotDao.LotDraftDetailBySID(c, request.SID); err != nil {
			log.Errorc(c, "s.lotDao.LotDraftDetailBySID(%v) failed. error(%v)", request.SID, err)
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
		if lmCheck, err = s.lotDao.LeastMarkCheckDraftList(c, request.SID); err != nil {
			log.Errorc(c, "s.lotDao.LeaskMarkCheckList(id: %v) failed. error(%v)", request.ID, err)
			return
		}
		for _, item := range lmCheck {
			if item.ID != request.ID {
				if _, err = s.lotDao.GiftDraftEdit(c, tx, item.ID, item.Name, item.Source, item.MessageTitle, item.MessageContent, item.ImgURL, item.Params, item.MemberGroup, item.DayNum,
					item.Num, item.Type, item.IsShow, lotmdl.GiftLeastMarkN, item.Effect, item.ProbabilityI, item.Extra, item.TimeLimit); err != nil {
					log.Errorc(c, "s.lotDao.GiftEdit(%+v) failed. error(%v)", request, err)
					return
				}
			}
		}
	}

	probability := request.GetIntProbability()
	if _, err = s.lotDao.GiftDraftEdit(c, tx, request.ID, request.Name, request.Source, request.MsgTitle, request.MsgContent, request.ImgURL, request.Params, request.MemberGroup, request.DayNum,
		request.Num, request.Type, request.IsShow, request.LeastMark, request.Effect, probability, request.Extra, request.TimeLimit); err != nil {
		log.Errorc(c, "s.lotDao.GiftEdit(%+v) failed. error(%v)", request, err)
		return
	}
	if err = s.lotDao.UpdateOperatorDraftBySIDAndState(c, request.SID, operator, lotmdl.LotteryDraftStateDraft); err != nil {
		log.Errorc(c, "s.lotDao.UpdateOperatorDraftBySIDAndState(sid: %v, operator:%v) failed. error(%v)", request.SID, operator, err)
		return
	}

	return
}

// MemberGroupDraftEdit update membergroup information
func (s *Service) MemberGroupDraftEdit(c context.Context, request *lotmdl.MemberGroupEditParam, cookie, operator string) (err error) {
	var (
		tx      *sql.Tx
		lottery *lotmdl.LotInfoDraft
	)
	if lottery, err = s.lotDao.LotDraftDetailBySID(c, request.SID); err != nil {
		log.Errorc(c, "s.lotDao.LotDraftDetailBySID(%v) failed. error(%v)", request.SID, err)
		return
	}
	//  验证当前抽奖是否可修改
	if err = s.draftStateCanUpdate(c, lottery); err != nil {
		return
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
	if err = s.updateOrInsertMemberGroupDraft(c, tx, request.SID, memberGroup); err != nil {
		log.Errorc(c, "s.updateOrInsertMemberGroup(sid: %v, memberGroup:%v) failed. error(%v)", request.SID, memberGroup, err)
		return
	}
	// if err = s.lotDao.DeleteMemberGroup(c, request.SID); err != nil {
	// 	log.Errorc(c, "s.lotDao.DeleteMemberGroup(sid: %v, operator:%v) failed. error(%v)", request.SID, operator, err)
	// 	return
	// }
	if err = s.lotDao.UpdateOperatorDraftBySIDAndState(c, request.SID, operator, lotmdl.LotteryDraftStateDraft); err != nil {
		log.Errorc(c, "s.lotDao.UpdateOperatorDraftBySIDAndState(sid: %v, operator:%v) failed. error(%v)", request.SID, operator, err)
		return
	}
	return
}

// updateOrInsertMemberGroupDraft 更新或插入用户组
func (s *Service) updateOrInsertMemberGroupDraft(c context.Context, tx *sql.Tx, sid string, memberGroup []*lotmdl.MemberGroupDB) (err error) {
	if memberGroup != nil && len(memberGroup) > 0 {
		if err = s.lotDao.BacthInsertOrUpdateMemberGroupDraft(c, tx, sid, memberGroup); err != nil {
			log.Errorc(c, "s.lotDao.BacthInsertOrUpdateMemberGroupDraft(memberGroup: %+v), error(%v)", memberGroup, err)
		}
	}
	return err
}

// GiftListDraft get gift list
func (s *Service) GiftListDraft(c context.Context, request *lotmdl.GiftListParam) (rsp *lotmdl.GiftList, err error) {
	var (
		giftList []*lotmdl.GiftInfo
		page     = lotmdl.Page{}
		//giftNum  map[int64]int
	)
	if page.Total, err = s.lotDao.GiftDraftTotal(c, request.SID, request.State, request.Type); err != nil {
		log.Errorc(c, "s.lot.GiftDraftTotal() failed. error(%v)", err)
		return
	}
	page.Num = request.Pn
	page.Size = request.Ps
	if giftList, err = s.lotDao.GiftDraftList(c, request.SID, request.Rank, request.State, request.Type, request.Pn, request.Ps); err != nil {
		log.Errorc(c, "s.lotDao.GiftDraftList() failed. error(%v)", err)
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

// DetailDraft ...
func (s *Service) DetailDraft(c context.Context, sid string, author string) (rsp *lotmdl.LotDetailInfoDraft, err error) {
	var (
		list        *lotmdl.LotInfoDraft
		info        *lotmdl.RuleInfo
		timeConf    []*lotmdl.TimesConf
		giftDraft   []*lotmdl.GiftInfo
		memberGroup []*lotmdl.MemberGroupDB
	)
	if list, err = s.lotDao.LotDraftDetailBySID(c, sid); err != nil {
		log.Errorc(c, "s.lotDao.LotDetailBySID(%v) failed. error(%v)", sid, err)
		return
	}
	if info, err = s.lotDao.GetLotRuleDraftBySID(c, sid); err != nil {
		log.Errorc(c, "s.lotDao.LotRuleBySID(%v) failed. error(%v)", sid, err)
		return
	}
	if timeConf, err = s.lotDao.AllTimesConfDraft(c, sid); err != nil {
		log.Errorc(c, "s.lotDao.AllTimesConfDraft(%v) failed. error(%v)", sid, err)
		return
	}
	if giftDraft, err = s.lotDao.AllGiftDraft(c, sid); err != nil {
		log.Errorc(c, "s.lotDao.AllGiftDraft(%v) failed. error(%v)", sid, err)
		return
	}

	if memberGroup, err = s.lotDao.AllMemberGroupDraft(c, sid); err != nil {
		log.Errorc(c, "s.lotDao.AllMemberGroupDraft(%v) failed. error(%v)", sid, err)
		return
	}
	rsp = &lotmdl.LotDetailInfoDraft{}
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
	if giftDraft != nil {
		for k, v := range giftDraft {
			giftDraft[k].ProbabilityF = giftDraft[k].GetFloatProbability()
			giftDraft[k].Params = s.getParams(c, v.Source, v.Params, v.Type)
		}
	}

	rsp.Gift = giftDraft
	rsp.MemberGroup = memberGroup
	rsp.CanAudit = s.canAudit(c, list.CanReviewer, author)
	return
}

// MemberGroupListDraft memberGroup list
func (s *Service) MemberGroupListDraft(c context.Context, request *lotmdl.MemberGroupListParam) (rsp *lotmdl.MemberGroupListReply, err error) {
	var (
		memberGroupList []*lotmdl.MemberGroupDB
		page            = lotmdl.Page{}
	)
	if page.Total, err = s.lotDao.MemberGroupDraftTotal(c, request.SID, request.State); err != nil {
		log.Errorc(c, "s.lot.MemberGroupDraftTotal() failed. error(%v)", err)
		return
	}
	page.Num = request.Pn
	page.Size = request.Ps
	if memberGroupList, err = s.lotDao.MemberGroupDraftList(c, request.SID, request.Rank, request.State, request.Pn, request.Ps); err != nil {
		log.Errorc(c, "s.lotDao.MemberGroupDraftList() failed. error(%v)", err)
		return
	}
	rsp = &lotmdl.MemberGroupListReply{}
	rsp.List = memberGroupList
	rsp.Page = page
	return
}

// Audit 审核
func (s *Service) Audit(c context.Context, sid string, state int, reviewer string, rejectReason string) (err error) {
	var (
		lottery *lotmdl.LotInfoDraft
	)
	if lottery, err = s.lotDao.LotDraftDetailBySID(c, sid); err != nil {
		log.Errorc(c, "s.lotDao.LotDraftDetailBySID(sid: %v) failed. error(%v)", sid, err)
		return
	}
	if err = s.auditPrefix(c, lottery, state, rejectReason, reviewer); err != nil {
		return err
	}
	if state == lotmdl.LotteryDraftStateWaitReject {
		return s.rejectLottery(c, sid, reviewer, rejectReason, lottery.Author)
	}
	if state == lotmdl.LotteryDraftStateWaitReviewed {
		return s.auditPassLottery(c, lottery, reviewer, lottery.Author)
	}
	return nil

}

// rejectLottery 拒绝抽奖
func (s *Service) rejectLottery(c context.Context, sid, reviewer, rejectReason, author string) error {
	err := s.lotDao.UpdateRejectReasonDraftBySID(c, sid, reviewer, rejectReason, lotmdl.LotteryDraftStateWaitReject)
	if err != nil {
		log.Errorc(c, "s.lotDao.UpdateRejectReasonDraftBySID(%s,%s,%s) err(%v)", sid, reviewer, rejectReason, err)
		return err
	}
	// 发送消息给用户
	to := make([]*componentmdl.Address, 0)
	to = append(to, &componentmdl.Address{Address: fmt.Sprintf("%s@bilibili.com", author), Name: author})
	err = s.sendMessageWechatEmail(c, []string{author}, to, to, to, s.c.Lottery.AuditRejectSubject, fmt.Sprintf("您的抽奖id：【%s】审批未通过，原因【%s】", sid, rejectReason))
	if err != nil {
		log.Errorc(c, "s.sendMessageWechatEmail err(%v)", err)
		return err
	}
	return nil
}

// auditPassLottery 审批通过
func (s *Service) auditPassLottery(c context.Context, lottery *lotmdl.LotInfoDraft, reviewer, author string) (err error) {
	var (
		tx          *sql.Tx
		rule        *lotmdl.RuleInfo
		times       []*lotmdl.TimesConf
		gift        []*lotmdl.GiftInfo
		memberGroup []*lotmdl.MemberGroupDB
	)
	if tx, err = s.lotDao.BeginTran(c); err != nil {
		log.Errorc(c, "s.lotDao.BeginTran() failed. error(%v)", err)
		return
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			log.Errorc(c, "%v", r)
			err = ecode.Error(ecode.RequestErr, "抽奖审核失败")
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

	// 创建或者更新抽奖
	if err = s.addOrUpdateNewLottery(c, tx, lottery); err != nil {
		return
	}
	// 创建或者更新规则
	if rule, err = s.addOrUpdateRule(c, tx, lottery); err != nil {
		return
	}
	// 更新次数配置
	if times, err = s.addOrUpdateTimes(c, tx, lottery); err != nil {
		return
	}
	// 更新奖品配置
	if gift, err = s.addOrUpdateGift(c, tx, lottery); err != nil {
		return
	}
	for _, v := range gift {
		if v.TimeLimit.Time().Unix() > xtime.Now().Unix() {
			key := lotmdl.GetTaskKey(v.Sid, v.ID, v.Type)
			s.GiftTasks[key] = v.TimeLimit.Time().Unix()
		}
	}
	// 更新用户组配置
	if memberGroup, err = s.addOrUpdateMemberGroup(c, tx, lottery); err != nil {
		return
	}
	now := time.Now().Unix()
	// 更新草稿信息
	if err = s.lotDao.UpdateLotDraftStatePass(c, tx, lottery.LotteryID, lotmdl.LotteryDraftStateWaitReviewed, reviewer, now); err != nil {
		return
	}
	// 通知
	to := make([]*componentmdl.Address, 0)
	to = append(to, &componentmdl.Address{Address: fmt.Sprintf("%s@bilibili.com", author), Name: author})

	canReviewer := strings.Split(lottery.CanReviewer, ",")
	wechat := make([]string, 0)
	wechat = append(wechat, canReviewer...)
	var authorInReviewer = false
	if canReviewer != nil && len(canReviewer) > 0 {
		for _, v := range canReviewer {
			to = append(to, &componentmdl.Address{Address: fmt.Sprintf("%s@bilibili.com", v), Name: v})
			if author == v {
				authorInReviewer = true
			}
		}

	}
	if !authorInReviewer {
		wechat = append(wechat, author)

	}
	err = s.sendMessageWechatEmail(c, wechat, to, to, to, s.c.Lottery.AuditPassSubject, s.lotteryEmailContentBuild(c, lottery, rule, times, gift, memberGroup))
	if err != nil {
		log.Errorc(c, "s.sendMessageWechatEmail err(%v)", err)
		return err
	}

	return nil
}

func (s *Service) lotteryEmailContentBuild(c context.Context, lottery *lotmdl.LotInfoDraft, rule *lotmdl.RuleInfo, times []*lotmdl.TimesConf, gift []*lotmdl.GiftInfo, memberGroup []*lotmdl.MemberGroupDB) string {
	title := fmt.Sprintf("<p>hi,all</p> <p>活动平台即将上线抽奖，基本信息如下：</p>")
	lotterySid := fmt.Sprintf("<p>抽奖ID：%v</p>", lottery.LotteryID)
	lotteryName := fmt.Sprintf("<p>活动名：%v</p>", lottery.Name)
	lotteryTime := fmt.Sprintf("<p>持续时间：%s-%s</p>", time.Unix(int64(lottery.STime), 0).Format("2006-01-02 15:04:05"), time.Unix(int64(lottery.ETime), 0).Format("2006-01-02 15:04:05"))
	mostWinTimes := 0
	mostWinType := ""
	if times != nil && len(times) > 0 {
		for _, v := range times {
			if v.Type == lotmdl.TimesTypePrice {
				mostWinTimes = v.Most
				if v.AddType == lotmdl.TimesAddTypeDay {
					mostWinType = "每日"
				} else {
					mostWinType = "活动内"

				}
			}
		}
	}
	var giftInfo string
	if gift != nil && len(gift) > 0 {
		for _, v := range gift {
			giftInfo += fmt.Sprintf("商品名：%v 库存：%d <br>", v.Name, v.Num)
		}
	}
	lotteryTimes := fmt.Sprintf("<p>用户最大中奖次数：%s最多%d次</p>", mostWinType, mostWinTimes)
	lotteryWin := fmt.Sprintf("<p>用户中奖概率：1/%d</p>", rule.GiftRate)
	lotteryGift := fmt.Sprintf("<p>奖品清单：</p> %v", giftInfo)
	return title + lotterySid + lotteryName + lotteryTime + lotteryTimes + lotteryWin + lotteryGift
}

// ListDraft get lottery information list
func (s *Service) ListDraft(c context.Context, request *lotmdl.ListParam) (rsp lotmdl.ListDraftRsp, err error) {
	var (
		total int
		Page  = &lotmdl.Page{}
		list  []*lotmdl.LotInfoDraft
	)
	if total, err = s.lotDao.ListTotalDraft(c, request.State, request.Keyword); err != nil {
		log.Error("s.lotDao.ListTotal() failed. error(%v)", err)
		return
	}
	if list, err = s.lotDao.BaseListDraft(c, request.Pn, request.Ps, request.State, request.Keyword, request.Rank); err != nil {
		log.Error("s.lotDao.BaseListDraft(%v,%v,%v,%v,%v) failed. error(%v)", request.Pn, request.Ps, request.State, request.Keyword, request.Rank, err)
		return
	}
	Page.Num = request.Pn
	Page.Size = request.Ps
	Page.Total = total
	rsp.Page = Page
	rsp.List = list
	return
}

// addOrUpdateTimes 新增或更新次数配置
func (s *Service) addOrUpdateTimes(c context.Context, tx *xsql.Tx, lottery *lotmdl.LotInfoDraft) (result []*lotmdl.TimesConf, err error) {
	result, err = s.lotDao.AllTimesConfTxDraft(c, tx, lottery.LotteryID)
	if err != nil {
		log.Errorc(c, "s.lotDao.AllTimesConfTxDraft error(%v)", err)

		return
	}
	err = s.lotDao.BatchInsertOrUpdateTimes(c, tx, result)
	if err != nil {
		log.Errorc(c, "s.lotDao.BacthInsertOrUpdateTimes error(%v)", err)
		return
	}
	return
}

// addOrUpdateTimes 新增或更新次数配置
func (s *Service) addOrUpdateGift(c context.Context, tx *xsql.Tx, lottery *lotmdl.LotInfoDraft) (result []*lotmdl.GiftInfo, err error) {
	result, err = s.lotDao.AllGiftTxDraft(c, tx, lottery.LotteryID)
	if err != nil {
		log.Errorc(c, "s.lotDao.AllGiftTxDraft error(%v)", err)

		return
	}
	err = s.lotDao.BatchInsertOrGift(c, tx, result)
	if err != nil {
		log.Errorc(c, "s.lotDao.BatchInsertOrGift error(%v)", err)
		return
	}
	return
}

// addOrUpdateRule 更新或者新增规则
func (s *Service) addOrUpdateRule(c context.Context, tx *xsql.Tx, lottery *lotmdl.LotInfoDraft) (rule *lotmdl.RuleInfo, err error) {
	if rule, err = s.lotDao.GetLotRuleDraftBySID(c, lottery.LotteryID); err != nil {
		log.Errorc(c, "s.lotDao.GetLotRuleDraftBySID(sid: %v) failed. error(%v)", lottery.LotteryID, err)
		return
	}
	if err = s.lotDao.BatchInsertOrUpdateRules(c, tx, rule); err != nil {
		log.Errorc(c, "s.lotDao.BatchInsertOrUpdateRules() failed. error(%v)", err)
		return
	}
	return
}

// addOrUpdateMemberGroup 更新或者新增规则
func (s *Service) addOrUpdateMemberGroup(c context.Context, tx *xsql.Tx, lottery *lotmdl.LotInfoDraft) (memberGroup []*lotmdl.MemberGroupDB, err error) {
	if memberGroup, err = s.lotDao.AllMemberGroupDraftTx(c, tx, lottery.LotteryID); err != nil {
		log.Errorc(c, "s.lotDao.AllMemberGroupTx(sid: %v) failed. error(%v)", lottery.LotteryID, err)
		return
	}
	if err = s.lotDao.BatchInsertOrUpdateMemberGroup(c, tx, lottery.LotteryID, memberGroup); err != nil {
		log.Errorc(c, "s.lotDao.BatchInsertOrUpdateMemberGroup() failed. error(%v)", err)
		return
	}
	return
}

// addOrUpdateNewLottery 创建或者修改抽奖
func (s *Service) addOrUpdateNewLottery(c context.Context, tx *xsql.Tx, lottery *lotmdl.LotInfoDraft) (err error) {
	// 查询是否已经存在抽奖
	detail, err := s.lotDao.LotDetailBySID(c, lottery.LotteryID)
	if err != nil {
		return err
	}
	if detail == nil {
		if err = s.lotDao.CreateNew(tx, lottery.ID, lottery.LotteryID, lottery.Name, lottery.Author, lottery.STime, lottery.ETime, lottery.Type); err != nil {
			log.Errorc(c, "Add s.lotDao.Add(%s, %s, %v, %v,%d) failed. error(%v)", lottery.Name, lottery.Author, lottery.STime, lottery.ETime, lottery.Type, err)
			return
		}
		return
	}
	if err = s.lotDao.UpdateLotInfo(tx, c, detail.ID, lottery.IsInternal, lottery.Name, lottery.Author, lottery.STime, lottery.ETime); err != nil {
		log.Errorc(c, "s.lotDao.UpdateLotInfo(id: %v,is_internal:%d name: %v, stime: %v, etime: %v) failed. error(%v)",
			detail.ID, lottery.IsInternal, lottery.Name, lottery.STime, lottery.ETime, err)
		return
	}
	return
}

func (s *Service) auditPrefix(c context.Context, lottery *lotmdl.LotInfoDraft, state int, rejectReason, reviewer string) (err error) {
	if lottery == nil {
		return ecode.Error(ecode.RequestErr, "找不到抽奖数据")
	}
	if lottery.State != lotmdl.LotteryDraftStateWaitReview {
		return ecode.Error(ecode.RequestErr, "待审中的抽奖才能被审核")
	}
	if state == lotmdl.LotteryDraftStateWaitReject && rejectReason == "" {
		return ecode.Error(ecode.RequestErr, "审批不通过，拒绝原因不能为空")
	}
	reviewerCan := s.canAudit(c, lottery.CanReviewer, reviewer)
	if !reviewerCan {
		return ecode.Error(ecode.RequestErr, "没有审核权限")
	}
	return nil
}

func (s *Service) canAudit(c context.Context, canReviewer string, nowReviewer string) bool {
	var reviewerCan = true
	if canReviewer != "" {
		reviewerCan = false
		canReviewer := strings.Split(canReviewer, ",")
		for _, v := range canReviewer {
			if v == nowReviewer {
				reviewerCan = true
			}
		}
	}
	return reviewerCan
}
