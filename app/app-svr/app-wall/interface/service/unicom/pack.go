package unicom

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"go-common/library/ecode"
	"go-common/library/log"
	"go-common/library/sync/errgroup.v2"
	xecode "go-gateway/app/app-svr/app-wall/ecode"
	"go-gateway/app/app-svr/app-wall/interface/model"
	"go-gateway/app/app-svr/app-wall/interface/model/unicom"

	account "git.bilibili.co/bapis/bapis-go/account/service"
	"github.com/pkg/errors"
)

// 1代表会员购
const _couponV2channel = 1

// nolint:gomnd
func (s *Service) PackReceive(ctx context.Context, mid int64, packID int64, now time.Time) (msg string, err error) {
	// 获取用户绑定信息
	bind, err := s.unicomBindInfo(ctx, mid)
	if err != nil {
		return "", err
	}
	order, err := s.verifyUserBiliBiliCard(ctx, bind, now)
	if err != nil {
		log.Error("[service.PackReceive] verifyUserBiliBiliCard error:%v", err)
		return "", err
	}
	// 获取礼包信息
	pack, err := s.unicomPackInfo(ctx, packID)
	if err != nil {
		return "", err
	}
	// 返回message构建
	defer func() {
		if err != nil {
			return
		}
		switch pack.Type {
		case 4: // b漫礼包
			msg = "兑换成功,快去“哔哩哔哩”漫画app使用吧"
		default:
			msg = pack.Desc + ",领取成功"
		}
	}()
	switch pack.Kind {
	case 1: // 可用流量兑换
		switch pack.Type {
		case 0: // 流量包兑换 消耗可用流量
			return s.flowPack(ctx, bind, pack, now)
		default:
			return "", xecode.AppWelfareClubPackNotExist
		}
	case 0: // 福利点兑换
		switch pack.Type {
		case 0: // 流量包兑换 消耗福利点
			return s.flowPack(ctx, bind, pack, now)
		case 5: // 直播礼包 不消化福利点，每张卡只能领取一次
			return "", s.livePack(ctx, bind.Usermob, mid, order)
		default: // 消耗福利点的礼包兑换
			return s.integralPack(ctx, bind, pack, now)
		}
	default:
		return "", xecode.AppWelfareClubPackNotExist
	}
}

// nolint:gocognit,gomnd
func (s *Service) integralPack(ctx context.Context, bind *unicom.UserBind, pack *unicom.UserPack, now time.Time) (msg string, err error) {
	// capped:0-无上限 1-有上限 amount礼包总量
	// 很早的逻辑设定 一直未做进一步功能开发
	// 接口不可用可使用
	// capped:2-禁止
	if pack.Capped == 2 || (pack.Capped == 1 && pack.Amount == 0) {
		return "", xecode.AppWelfareClubPackLack
	}
	if bind.Integral < pack.Integral {
		return "", xecode.AppWelfareClubLackIntegral
	}
	ok, err := s.allowExchangePack(ctx, strconv.Itoa(bind.Phone), now)
	if err != nil {
		return "", err
	}
	if !ok {
		return "", xecode.AppWelfareClubNoAllowPackExchange
	}
	action := model.ActionIntegralPack
	kind := model.ConsumeIntegral
	requestNo, err := s.seqdao.SeqID(ctx)
	if err != nil {
		return "", err
	}
	var preFunc, payFunc packFunc
	switch pack.Type {
	case 1:
		var (
			batchID int64
			appKey  string
		)
		preFunc = func(ctx context.Context, bind *unicom.UserBind, pack *unicom.UserPack) (string, error) {
			vip := s.c.AccountVIP[pack.Desc]
			if vip == nil {
				log.Error("日志告警 找不到大会员兑换参数,desc:%v", pack.Desc)
				return "", xecode.AppWelfareClubPackNotExist
			}
			customCheck, _ := s.gaiaEngine.InitCheck(ctx, "unicom_welfare_rewards")
			customCheck.Put("subscene", "兑换")
			customCheck.Put("phone_num", bind.Phone)
			customCheck.Put("mid", bind.Mid)
			customCheck.Put("vip_days", vip.Days)
			report, _ := customCheck.Do()
			if report != nil && report.CheckHit("reject") {
				return "", xecode.AppWelfareClubRejectPackExchange
			}
			batchID = vip.BatchID
			appKey = vip.AppKey
			return "", nil
		}
		payFunc = func(ctx context.Context, bind *unicom.UserBind, pack *unicom.UserPack) (string, error) {
			if err := s.accd.AddVIP(ctx, bind.Mid, batchID, requestNo, pack.Desc, appKey); err != nil {
				log.Error("msg:%v,err:%+v", msg, err)
				// 69006 大会员库存不足
				if ecode.EqualError(ecode.Int(69006), err) {
					if err := s.forbidenPack(ctx, pack.ID); err != nil {
						log.Error("%+v", err)
					}
					return "", xecode.AppWelfareClubPackLack
				}
				if err := s.SendPackQueue(ctx, action, kind, bind, pack, "", "", now); err != nil {
					log.Error("%+v", err)
					return "", err
				}
				return "", xecode.AppWelfareClubWaitResult
			}
			return "", nil
		}
	case 2:
		var day int
		preFunc = func(ctx context.Context, bind *unicom.UserBind, pack *unicom.UserPack) (string, error) {
			day, err = strconv.Atoi(pack.Param)
			return "", err
		}
		payFunc = func(ctx context.Context, bind *unicom.UserBind, pack *unicom.UserPack) (string, error) {
			if msg, err := s.live.AddVIP(ctx, bind.Mid, day); err != nil {
				log.Error("msg:%v,err:%+v", msg, err)
				if err := s.SendPackQueue(ctx, action, kind, bind, pack, "", "", now); err != nil {
					log.Error("%+v", err)
					return msg, err
				}
				return "", xecode.AppWelfareClubWaitResult
			}
			return "", nil
		}
	case 3:
		var info *account.Info
		preFunc = func(ctx context.Context, bind *unicom.UserBind, pack *unicom.UserPack) (string, error) {
			info, err = s.accd.Info(ctx, bind.Mid)
			return "", err
		}
		payFunc = func(ctx context.Context, bind *unicom.UserBind, pack *unicom.UserPack) (string, error) {
			param := &unicom.CouponParam{}
			param.AssetRequest.Channel = _couponV2channel
			param.AssetRequest.SourceBizId = strconv.FormatInt(requestNo, 10)
			param.AssetRequest.Mid = bind.Mid
			param.SourceAuthorityId = pack.NewParam
			param.SourceId = s.c.CouponV2.SourceID
			if msg, err := s.shop.CouponV2(ctx, param); err != nil {
				log.Error("msg:%v,err:%+v", msg, err)
				// 83110020 会员购优惠券到期
				if ecode.EqualError(ecode.Int(83110020), err) {
					if err := s.forbidenPack(ctx, pack.ID); err != nil {
						log.Error("%+v", err)
					}
					return "", xecode.AppWelfareClubPackLack
				}
				// 83110005 该优惠券领取次数超过上限
				if ecode.EqualError(ecode.Int(83110005), err) {
					return "", xecode.AppWelfareClubPackCountLimit
				}
				// 同步设置
				bind.Name = info.Name
				if err := s.SendPackQueue(ctx, action, kind, bind, pack, "", "", now); err != nil {
					log.Error("%+v", err)
					return msg, err
				}
				return "", xecode.AppWelfareClubWaitResult
			}
			s.cache.Do(ctx, func(ctx context.Context) {
				if err = s.dao.AddCouponV2ReqCache(ctx, param); err != nil {
					log.Error("[s.integralPack] AddCouponV2ReqCache error, error:%v", err)
				}
			})
			return "", nil
		}
	case 4:
		var amount int
		preFunc = func(ctx context.Context, bind *unicom.UserBind, pack *unicom.UserPack) (string, error) {
			if amount, err = strconv.Atoi(pack.Param); err != nil {
				return "", err
			}
			isComicUser, err := s.comicdao.ComicUser(ctx, bind.Mid)
			if err != nil {
				return "", err
			}
			if !isComicUser {
				return "", xecode.AppComicUserNotExist
			}
			return "", nil
		}
		payFunc = func(ctx context.Context, bind *unicom.UserBind, pack *unicom.UserPack) (string, error) {
			if msg, err := s.comicdao.Coupon(ctx, bind.Mid, amount); err != nil {
				log.Error("msg:%v,err:%+v", msg, err)
				if err := s.SendPackQueue(ctx, action, kind, bind, pack, "", "", now); err != nil {
					log.Error("%+v", err)
					return msg, err
				}
				return "", xecode.AppWelfareClubWaitResult
			}
			return "", nil
		}
	default:
		return "", xecode.AppWelfareClubPackNotExist
	}
	if msg, err := s.pay(ctx, bind, pack, model.ConsumeIntegral, preFunc, payFunc, now); err != nil {
		return msg, err
	}
	s.packLog(ctx, model.ConsumeIntegral, bind, pack, strconv.FormatInt(requestNo, 10))
	return "", nil
}

type packFunc func(ctx context.Context, bind *unicom.UserBind, pack *unicom.UserPack) (string, error)

func (s *Service) pay(ctx context.Context, bind *unicom.UserBind, pack *unicom.UserPack, kind model.ConsumeKind, preFunc packFunc, payFunc packFunc, now time.Time) (string, error) {
	if payFunc == nil {
		return "", nil
	}
	if preFunc != nil {
		if msg, err := preFunc(ctx, bind, pack); err != nil {
			return msg, err
		}
	}
	switch kind {
	case model.ConsumeIntegral:
		if err := s.comsumeUserIntegral(ctx, bind.Mid, bind.Phone, pack.Integral, bind.Usermob, now); err != nil {
			return "", err
		}
	case model.ConsumeFlow:
		if err := s.consumeUserFlow(ctx, bind.Mid, bind.Phone, flowValue(pack), bind.Usermob, now); err != nil {
			return "", err
		}
	}
	msg, err := payFunc(ctx, bind, pack)
	if err != nil {
		// 稍后查看结果，表示会重试，暂不返还福利点
		// 重试逻辑会根据情况返还
		if err == xecode.AppWelfareClubWaitResult {
			return "", err
		}
		// 失败返还
		switch kind {
		case model.ConsumeIntegral:
			if err := s.addUserIntegral(ctx, bind.Mid, bind.Phone, pack.Integral, bind.Usermob, now); err != nil {
				return msg, err
			}
		case model.ConsumeFlow:
			if err := s.addUserFlow(ctx, bind.Mid, bind.Phone, flowValue(pack), bind.Usermob, now); err != nil {
				return msg, err
			}
		}
		return msg, err
	}
	return msg, nil
}

func (s *Service) FlowPack(ctx context.Context, mid int64, flowID string, now time.Time) (msg string, err error) {
	// 获取用户绑定信息
	bind, err := s.unicomBindInfo(ctx, mid)
	if err != nil {
		return "", err
	}
	order, err := s.verifyUserBiliBiliCard(ctx, bind, now)
	if err != nil {
		log.Error("[service.FlowPack] verifyUserBiliBiliCard error:%v", err)
		return "", err
	}
	log.Info("FlowPack mid:%v,order:%+v", mid, order)
	pack := &unicom.UserPack{
		Kind:  1,
		Type:  0,
		Param: flowID,
	}
	return s.flowPack(ctx, bind, pack, now)
}

func (s *Service) flowPack(ctx context.Context, bind *unicom.UserBind, pack *unicom.UserPack, now time.Time) (msg string, err error) {
	var kind model.ConsumeKind
	switch pack.Kind {
	case 1: // 流量兑换
		switch pack.Type {
		case 0: // 可用流量兑换流量包
			flow := flowValue(pack)
			pack.Desc = flowDesc(pack)
			if flow == 0 {
				return "", xecode.AppWelfareClubPackNotExist
			}
			if bind.Flow < flow {
				return "", xecode.AppWelfareClubLackFlow
			}
			kind = model.ConsumeFlow
		default:
			return "", xecode.AppWelfareClubPackNotExist
		}
	case 0: // 礼包兑换
		switch pack.Type {
		case 0: // 福利点兑换流量包
			if bind.Integral < pack.Integral {
				return "", xecode.AppWelfareClubLackIntegral
			}
			ok, err := s.allowExchangePack(ctx, strconv.Itoa(bind.Phone), now)
			if err != nil {
				return "", err
			}
			if !ok {
				return "", xecode.AppWelfareClubNoAllowPackExchange
			}
			kind = model.ConsumeIntegral
		default:
			return "", xecode.AppWelfareClubPackNotExist
		}
	default:
		return "", xecode.AppWelfareClubPackNotExist
	}
	var (
		preFunc, payFunc packFunc
		requestNo        int64
		orderID          string
	)
	preFunc = func(ctx context.Context, bind *unicom.UserBind, pack *unicom.UserPack) (string, error) {
		// redis锁 实现用户间隔一分钟才能进行下一次兑换
		lcokKey := flowPackLockKey(bind.Mid)
		locked, err := s.lockdao.TryLock(ctx, lcokKey, 60)
		if err != nil {
			log.Error("TryLock fail key：%s err:%+v", lcokKey, err)
			return "", xecode.AppWelfareClubWaitOneMinute
		}
		if !locked {
			log.Warn("TryLock fail key:%s", lcokKey)
			return "", xecode.AppWelfareClubWaitOneMinute
		}
		if requestNo, err = s.seqdao.SeqID(ctx); err != nil {
			return "", err
		}
		// 防刷检查
		if msg, err := s.dao.FlowPre(ctx, bind.Phone, requestNo, now); err != nil {
			return msg, err
		}
		return "", nil
	}
	payFunc = func(ctx context.Context, bind *unicom.UserBind, pack *unicom.UserPack) (string, error) {
		var outOrderID string
		if orderID, outOrderID, msg, err = s.dao.FlowExchange(ctx, bind.Phone, pack.Param, requestNo, now); err != nil {
			log.Error("msg:%v,err:%+v", msg, err)
			return msg, err
		}
		// 加入到缓存队列，job消费
		// 高并发，会造成数据丢失，不可靠
		// 使用databus，扣除失败后，会返回并落日志
		if err := s.SendPackQueue(ctx, model.ActionFlowPack, kind, bind, pack, orderID, outOrderID, now); err != nil {
			log.Error("%+v", err)
		}
		return "", nil
	}
	if msg, err = s.pay(ctx, bind, pack, kind, preFunc, payFunc, now); err != nil {
		return msg, err
	}
	s.packLog(ctx, kind, bind, pack, orderID)
	return msg, nil
}

// nolint:gomnd
func (s *Service) allowExchangePack(ctx context.Context, phone string, now time.Time) (bool, error) {
	for _, val := range s.c.Unicom.ExchangeLimit.PhoneWhitelist {
		if phone == val {
			return true, nil
		}
	}
	reducePackIntegral, err := s.dao.ReducePackIntegral(ctx, phone, now)
	if err != nil {
		return false, err
	}
	if reducePackIntegral < 5000 {
		return true, nil
	}
	return false, nil
}

// nolint:gomnd
func flowValue(pack *unicom.UserPack) (flow int) {
	if pack.Kind != 1 && pack.Type != 0 {
		return 0
	}
	switch pack.Param {
	case "01":
		return 100
	case "02":
		return 200
	case "03":
		return 300
	case "04":
		return 500
	case "05":
		return 1024
	case "06":
		return 2048
	default:
		return 0
	}
}

func flowDesc(pack *unicom.UserPack) (desc string) {
	if pack.Kind != 1 && pack.Type != 0 {
		return pack.Desc
	}
	if pack.Desc == "" {
		switch pack.Param {
		case "01":
			return "100MB流量包"
		case "02":
			return "200MB流量包"
		case "03":
			return "300MB流量包"
		case "04":
			return "500MB流量包"
		case "05":
			return "1024MB流量包"
		case "06":
			return "2048MB流量包"
		default:
			return ""
		}
	}
	// TODO 后续修改数据库，流量改为流量包，此逻辑可去掉
	if !strings.HasSuffix(pack.Desc, "包") {
		return pack.Desc + "包"
	}
	return pack.Desc
}

func (s *Service) Pack(c context.Context, usermob string, mid int64, now time.Time) error {
	orderm, err := s.orders(c, usermob, now)
	if err != nil {
		return err
	}
	if len(orderm) == 0 {
		return xecode.AppWelfareClubNotFree
	}
	order, err := s.orderState(orderm, unicom.CardProduct, now)
	if err != nil {
		return err
	}
	return s.livePack(c, usermob, mid, order)
}

// nolint:gomnd
func (s *Service) livePack(c context.Context, usermob string, mid int64, order *unicom.Unicom) (err error) {
	switch order.CardType {
	case 1, 2, 3:
	default:
		return xecode.AppWelfareClubOnlySupportCard
	}
	tx, err := s.dao.BeginTran(c)
	if err != nil {
		return
	}
	defer func() {
		if r := recover(); r != nil {
			_ = tx.Rollback()
			log.Error("%v", r)
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
	rows, err := s.dao.InPack(tx, usermob, mid)
	if err != nil {
		log.Error("%+v", err)
		return
	}
	if rows == 0 {
		err = xecode.AppWelfareClubOnlyOnce
		return
	}
	if err = s.live.Pack(c, mid, int64(order.CardType)); err != nil {
		log.Error("%+v", err)
		return
	}
	return
}

func (s *Service) packLog(ctx context.Context, kind model.ConsumeKind, bind *unicom.UserBind, pack *unicom.UserPack, orderID string) {
	log.Warn("packLog sucess mid:%v,orderID:%v,%v,领取成功", bind.Mid, orderID, pack.Desc)
	var (
		integral int
		desc     string
	)
	switch kind {
	case model.ConsumeIntegral:
		integral = pack.Integral
		desc = "福利点"
	case model.ConsumeFlow:
		integral = flowValue(pack)
		desc = "MB流量"
	}
	userDesc := fmt.Sprintf("您当前已领取%v，扣除%v%v", pack.Desc, integral, desc)
	s.cache.Do(ctx, func(ctx context.Context) {
		v := &unicom.UserPackLog{
			Phone:     bind.Phone,
			Usermob:   bind.Usermob,
			Mid:       bind.Mid,
			RequestNo: orderID,
			Type:      pack.Type,
			Desc:      pack.Desc,
			UserDesc:  userDesc,
			Integral:  integral,
		}
		s.addUserPackLog(ctx, v)
	})
}

func flowPackLockKey(mid int64) string {
	return fmt.Sprintf("flow_pack_lock_%d", mid)
}

// nolint:bilirailguncheck
func (s *Service) SendPackQueue(ctx context.Context, action model.Action, kind model.ConsumeKind, bind *unicom.UserBind, pack *unicom.UserPack, orderID, outOrderID string, now time.Time) error {
	var integral, flow int
	switch kind {
	case model.ConsumeIntegral:
		integral = pack.Integral
	case model.ConsumeFlow:
		flow = flowValue(pack)
	}
	v := unicom.PackMsg{
		Action: action,
		Data: &unicom.PackData{
			Kind:       kind,
			Phone:      bind.Phone,
			Mid:        bind.Mid,
			Usermob:    bind.Usermob,
			Name:       bind.Name,
			Integral:   integral,
			Flow:       flow,
			OrderID:    orderID,
			OutorderID: outOrderID,
			PackID:     pack.ID,
			Desc:       pack.Desc,
			Type:       pack.Type,
			Param:      pack.Param,
			Ctime:      now,
			NewParam:   pack.NewParam,
		},
	}
	key := strconv.FormatInt(bind.Mid, 10)
	if err := s.packPub.Send(ctx, key, v); err != nil {
		return err
	}
	log.Warn("send pack queue action:%v,msg:%+v", v.Action, v.Data)
	return nil
}

func (s *Service) verifyUserBiliBiliCard(ctx context.Context, bind *unicom.UserBind, now time.Time) (*unicom.Unicom, error) {
	var order *unicom.Unicom
	eg := errgroup.WithContext(ctx)
	eg.Go(func(ctx context.Context) error {
		// 获取用户免流卡套餐信息，非免流卡返回错误
		orderm, err := s.orders(ctx, bind.Usermob, now)
		if err != nil {
			log.Error("[service.verifyUserBiliBiliCard] orders error:%v", err)
			return err
		}
		if order, err = s.orderState(orderm, "", now); err != nil {
			log.Error("[service.verifyUserBiliBiliCard] orderState error:%v", err)
			return err
		}
		if order.ProductType != 1 {
			return xecode.AppWelfareClubOnlySupportCard
		}
		return nil
	})
	eg.Go(func(ctx context.Context) error {
		resp, err := s.dao.VerifyBiliBiliCardByUnicom(ctx, strconv.Itoa(bind.Phone))
		if err != nil {
			log.Error("[service.verifyUserBiliBiliCard] VerifyBiliBiliCardByUnicom error:%v", err)
			return xecode.AppWelfareClubRequestUnicom
		}
		if resp.Code != "0000" {
			if resp.Code == "2013" {
				return errors.WithStack(xecode.AppWelfareClubUnicomServiceUpgrade)
			}
			return xecode.AppWelfareClubOnlySupportCardFromUnicom
		}
		return nil
	})
	if err := eg.Wait(); err != nil {
		log.Error("[service.verifyUserBiliBiliCard] error:%v", err)
		return nil, err
	}
	return order, nil
}

func (s *Service) CouponVerify(ctx context.Context, param *unicom.CouponParam) error {
	data, err := s.dao.GetCouponV2ReqCache(ctx, param.AssetRequest.Mid, param.AssetRequest.SourceBizId)
	if err != nil {
		log.Error("s.CouponVerify err:%+v", err)
		return err
	}
	if !data.Verify(*param) {
		log.Error("s.CouponVerify verify fail")
		return errors.WithMessage(ecode.RequestErr, "verify fail")
	}
	return nil
}
