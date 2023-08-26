package unicom

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"go-common/library/log"
	"go-common/library/net/netutil"
	"go-common/library/railgun"
	"go-common/library/retry"
	xecode "go-gateway/app/app-svr/app-wall/ecode"
	"go-gateway/app/app-svr/app-wall/interface/model"
	unicomdl "go-gateway/app/app-svr/app-wall/interface/model/unicom"
	"go-gateway/app/app-svr/app-wall/job/model/unicom"
)

// 1代表会员购
const _couponV2channel = 1

func (s *Service) initPackRailGun(cfg *railgun.DatabusV1Config, pcfg *railgun.SingleConfig) {
	inputer := railgun.NewDatabusV1Inputer(cfg)
	processor := railgun.NewSingleProcessor(pcfg, s.packUnpack, s.packDo)
	g := railgun.NewRailGun("订阅binlog", nil, inputer, processor)
	s.packRailGun = g
	g.Start()
}

func (s *Service) packUnpack(msg railgun.Message) (*railgun.SingleUnpackMsg, error) {
	var v *unicomdl.PackMsg
	if err := json.Unmarshal(msg.Payload(), &v); err != nil {
		return nil, err
	}
	if v == nil || v.Data == nil {
		return nil, nil
	}
	return &railgun.SingleUnpackMsg{
		Group: time.Now().Unix(),
		Item:  v,
	}, nil
}

func (s *Service) packDo(ctx context.Context, item interface{}) railgun.MsgPolicy {
	data := item.(*unicomdl.PackMsg)
	// 开始时间在这里初始化，防止消费中断恢复后，礼包直接降级
	if data.Data.Stime.IsZero() {
		data.Data.Stime = time.Now()
	}
	switch data.Action {
	case model.ActionIntegralPack:
		if err := s.integralPackPay(ctx, data.Data); err != nil {
			log.Error("%+v", err)
			return railgun.MsgPolicyRetryInfinite
		}
		return railgun.MsgPolicyNormal
	case model.ActionFlowPack:
		key := checkFlowLockKey(data.Data.Mid, data.Data.OrderID, data.Data.OutorderID)
		locked, err := s.lockdao.TryLock(ctx, key, s.lockExpire)
		if err != nil {
			log.Error("日志告警 流量包发放检查获取锁错误,date:%+v,err:%+v", data.Data, err)
			return railgun.MsgPolicyAttempts
		}
		if !locked {
			log.Error("流量包发放已检查过,date:%+v", data.Data)
			return railgun.MsgPolicyNormal
		}
		finish, err := s.flowCheck(ctx, data.Data)
		if err != nil {
			log.Error("%+v", err)
		}
		if !finish {
			// 尝试删除锁
			if err := retry.WithAttempts(ctx, "unlock", 3, netutil.DefaultBackoffConfig, func(ctx context.Context) error {
				return s.lockdao.UnLock(ctx, key)
			}); err != nil {
				log.Error("日志告警 流量包发放检查获取锁错误,date:%+v,err:%+v", data.Data, err)
			}
			if err := retry.WithAttempts(ctx, "send_pack_queue", 3, netutil.DefaultBackoffConfig, func(ctx context.Context) error {
				return s.sendPackQueue(ctx, data)
			}); err != nil {
				log.Error("日志告警 流量包发放检查获取锁错误,date:%+v,err:%+v", data.Data, err)
			}
		}
		return railgun.MsgPolicyNormal
	default:
		return railgun.MsgPolicyNormal
	}
}

func (s *Service) flowCheck(ctx context.Context, data *unicomdl.PackData) (finish bool, err error) {
	requestNo, err := s.seqdao.SeqID(ctx)
	if err != nil {
		return false, err
	}
	orderStatus, msg, err := s.dao.FlowQry(ctx, data.Phone, requestNo, data.OutorderID, data.OrderID, time.Now())
	if err != nil {
		log.Error("status:%v,msg:%v,err%+v", orderStatus, msg, err)
		return false, err
	}
	log.Warn("load unicom userbind flow=%+v orderstatus=%v", data, orderStatus)
	if orderStatus == "01" { // 已到账，成功
		return true, nil
	}
	if orderStatus == "00" { // 未到账，继续等待结果
		if time.Since(data.Stime) > 36*time.Hour { // 针对历史数据
			log.Warn("免流包检查 %+v超过36小时联通接口仍返回未到账,默认已领取", data)
			return true, nil
		}
		if time.Since(data.Stime) < 24*time.Hour {
			log.Warn("免流包检查 %+v小于24小时联通接口仍返回未到账,重新进入消息队列等待重试", data)
			return false, nil
		}
		log.Warn("免流包检查 %+v超过24小时联通接口仍返回未到账,返还点数", data)
	}
	// 其他 orderStatus 发放失败，退还
	var (
		integral int
		desc     string
	)
	switch data.Kind {
	case model.ConsumeIntegral:
		if _, err := s.dao.BackUserBindIntegral(ctx, data.Mid, strconv.Itoa(data.Phone), data.Integral, data.Ctime); err != nil {
			log.Error("%+v", err)
			return false, err
		}
		integral = data.Integral
		desc = "福利点"
	case model.ConsumeFlow:
		if _, err := s.dao.BackUserBindFlow(ctx, data.Mid, strconv.Itoa(data.Phone), data.Flow, data.Ctime); err != nil {
			log.Error("%+v", err)
			return false, err
		}
		integral = data.Flow
		desc = "MB流量"
	}
	userDesc := fmt.Sprintf("%v，领取失败并返还%v%v", data.Desc, integral, desc)
	v := &unicom.UserPackLog{
		Phone:   data.Phone,
		Usermob: data.Usermob,
		Mid:     data.Mid,
		// outOrderID to orderID
		RequestNo: data.OrderID,
		Type:      data.Type,
		Desc:      data.Desc + ",领取失败并返还",
		Integral:  integral,
		UserDesc:  userDesc,
	}
	s.addUserPackLog(v)
	s.addUserIntegralLog(v)
	log.Warn("packLog sucess mid:%v,orderID:%v,%v,领取失败并返还成功", data.Mid, data.OrderID, data.Desc)
	return true, nil
}

func (s *Service) integralPackPay(ctx context.Context, data *unicomdl.PackData) error {
	// 获取礼包信息
	pack, err := s.unicomPackInfo(ctx, data.PackID)
	if err != nil {
		return err
	}
	// 礼包不可兑换，或者超过上限，返还福利点
	if pack.Capped == 2 || (pack.Capped == 1 && pack.Amount == 0) {
		if _, err := s.dao.BackUserBindIntegral(ctx, data.Mid, strconv.Itoa(data.Phone), data.Integral, data.Ctime); err != nil {
			log.Error("%+v", err)
			return err
		}
		log.Warn("返还福利点成功 mid:%v,%+v", data.Mid, data)
		return nil
	}
	requestNo, err := s.seqdao.SeqID(ctx)
	if err != nil {
		return err
	}
	data.OrderID = strconv.FormatInt(requestNo, 10)
	if err := s.packExchange(ctx, requestNo, data); err != nil {
		// 兑换失败后，比较时间，超过5分钟，还未兑换成功，进行下架处理
		if time.Since(data.Stime) <= 5*time.Minute {
			return err
		}
		// 设置礼包不可兑换
		if err1 := s.forbidenPack(ctx, data.PackID); err1 != nil {
			log.Error("%+v", err1)
		} else {
			log.Error("日志告警 设置礼包不可兑换成功 pack_id:%v,pack_desc:%v,err:%+v", data.PackID, data.Desc, err)
		}
		if _, err := s.dao.BackUserBindIntegral(ctx, data.Mid, strconv.Itoa(data.Phone), data.Integral, data.Ctime); err != nil {
			log.Error("%+v", err)
			return err
		}
		log.Warn("返还福利点成功 mid:%v,%+v", data.Mid, data)
		return nil
	}
	userDesc := fmt.Sprintf("您当前已领取%v，扣除%v福利点", data.Desc, data.Integral)
	v := &unicom.UserPackLog{
		Phone:     data.Phone,
		Usermob:   data.Usermob,
		Mid:       data.Mid,
		RequestNo: data.OrderID,
		Type:      data.Type,
		Desc:      data.Desc,
		Integral:  data.Integral,
		UserDesc:  userDesc,
	}
	s.addUserPackLog(v)
	// 行为日志中显示成功扣除，福利点取负
	u := &unicom.UserPackLog{}
	*u = *v
	u.Integral = -data.Integral
	s.addUserIntegralLog(u)
	log.Warn("packLog sucess mid:%v,orderID:%v,%v,领取成功", data.Mid, data.OrderID, data.Desc)
	return nil
}

func (s *Service) packExchange(ctx context.Context, requestNo int64, data *unicomdl.PackData) error {
	const (
		_accountVIP = 1
		_liveVIP    = 2
		_shop       = 3
		_comic      = 4
	)
	switch data.Type {
	case _accountVIP:
		var (
			batchID int64
			appKey  string
		)
		vip := s.c.AccountVIP[data.Desc]
		if vip == nil {
			log.Error("日志告警 找不到大会员兑换参数,desc:%v", data.Desc)
			return xecode.AppWelfareClubPackNotExist
		}
		batchID = vip.BatchID
		appKey = vip.AppKey
		if err := s.accd.AddVIP(ctx, data.Mid, batchID, requestNo, data.Desc, appKey); err != nil {
			log.Error("%+v", err)
			return err
		}
	case _liveVIP:
		day, _ := strconv.Atoi(data.Param)
		if _, err := s.live.AddVIP(ctx, data.Mid, day); err != nil {
			log.Error("%+v", err)
			return err
		}
	case _shop:
		param := &unicomdl.CouponParam{}
		param.AssetRequest.Channel = _couponV2channel
		param.AssetRequest.SourceBizId = strconv.FormatInt(requestNo, 10)
		param.AssetRequest.Mid = data.Mid
		param.SourceAuthorityId = data.NewParam
		param.SourceId = s.c.CouponV2.SourceID
		if _, err := s.shop.CouponV2(ctx, param); err != nil {
			log.Error("%+v", err)
			return err
		}
		s.cache.Do(ctx, func(ctx context.Context) {
			if err := s.dao.AddCouponV2ReqCache(ctx, param); err != nil {
				log.Error("%+v", err)
			}
		})
	case _comic:
		amount, _ := strconv.Atoi(data.Param)
		if _, err := s.comic.Coupon(ctx, data.Mid, amount); err != nil {
			log.Error("%+v", err)
			return err
		}
	}
	return nil
}

func (s *Service) updateUserBind(ctx context.Context, mid int64) error {
	res, err := s.dao.UserBind(ctx, mid)
	if err != nil {
		return err
	}
	// 没有state=1的订购，清除缓存，老逻辑需要考虑缓存穿透
	if res == nil {
		return s.dao.DeleteUserBindCache(ctx, mid)
	}
	return s.dao.AddUserBindCache(ctx, mid, res)
}

func (s *Service) forbidenPack(ctx context.Context, id int64) error {
	_, err := s.dao.SetUserPackFlow(ctx, id, 2)
	return err
}

// unicomPackInfo unicom pack infos
func (s *Service) unicomPackInfo(ctx context.Context, id int64) (res *unicom.UserPack, err error) {
	if res, err = s.dao.UserPackCache(ctx, id); err != nil {
		log.Error("%+v", err)
	}
	if err == nil {
		s.pHit.Incr("unicoms_pack_cache")
		return
	}
	if res, err = s.dao.UserPackByID(ctx, id); err != nil {
		log.Error("%+v", err)
		return
	}
	s.pMiss.Incr("unicoms_pack_cache")
	if res == nil {
		err = xecode.AppWelfareClubPackNotExist
		return
	}
	if err = s.dao.AddUserPackCache(ctx, id, res); err != nil {
		log.Error("s.dao.AddUserPackCache id(%d) error(%v)", id, err)
		return
	}
	return
}

// nolint:bilirailguncheck
func (s *Service) sendPackQueue(ctx context.Context, v *unicomdl.PackMsg) error {
	key := strconv.FormatInt(v.Data.Mid, 10)
	if err := s.packPub.Send(ctx, key, v); err != nil {
		return err
	}
	log.Warn("send pack queue retry msg,action:%v,data:%+v", v.Action, v.Data)
	return nil
}

func checkFlowLockKey(mid int64, orderID, outOrderID string) string {
	// key的前缀是当日在这一个月中的哪一天，按月循环
	// key的超时时间设置为略大于一天，25小时，满足当前按天设限的场景
	return fmt.Sprintf("check_flow_lock_%d_%s_%s", mid, orderID, outOrderID)
}
