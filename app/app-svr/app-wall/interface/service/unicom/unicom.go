package unicom

import (
	"bytes"
	"context"

	// nolint:gosec
	"crypto/des"
	"encoding/base64"
	"encoding/json"
	"errors"
	"strconv"
	"time"

	"go-common/library/conf/env"
	"go-common/library/ecode"
	log "go-common/library/log"
	"go-common/library/queue/databus/report"
	xecode "go-gateway/app/app-svr/app-wall/ecode"
	"go-gateway/app/app-svr/app-wall/interface/model"
	"go-gateway/app/app-svr/app-wall/interface/model/unicom"

	account "git.bilibili.co/bapis/bapis-go/account/service"
)

const (
	_unicomKey           = "unicom"
	_unicomPackKey       = "unicom_pack"
	_entryPink           = 0
	_entryComic          = 1
	_defalutLogRequestNo = "defalut_log_requestno"
	_fakeIDMonthLayout   = "200601"
)

// InOrdersSync insert OrdersSync
func (s *Service) InOrdersSync(c context.Context, usermob, ip string, u *unicom.UnicomJson, now time.Time) error {
	if env.DeployEnv != env.DeployEnvUat && !s.iplimit(_unicomKey, ip) {
		return ecode.AccessDenied
	}
	if u.Ordertime == "" {
		log.Error("日志告警 联通订单同步 ordertime 为空,order:%+v", u)
	}
	if err := s.dao.InOrdersSync(c, usermob, u, now); err != nil {
		log.Error("unicom_s.dao.OrdersSync usermob:(%s),unicom:(%+v),error:(%v)", usermob, u, err)
		return err
	}
	if u.FakeID == "" || u.FakeIDMonth == "" {
		log.Error("[service.InOrdersSync]联通订单fake_id或者month为空,order:%+v", u)
		return nil
	}
	fakeIDTime, err := time.Parse(_fakeIDMonthLayout, u.FakeIDMonth)
	if err != nil {
		// 解析周期错误
		log.Error("[service.InOrdersSync]日志告警 联通订单fake_id month与期望值不符合, err:%v, order:%+v", err, u)
		return ecode.Error(ecode.RequestErr, "fake_id时间与期望不符")
	}
	period := fakeIDTime.Month()
	if period < time.January || period > time.December {
		log.Error("[service.InOrdersSync]日志告警 联通订单fake_id month与期望值不符合, order:%+v", u)
		return ecode.Error(ecode.RequestErr, "fake_id时间与期望不符")
	}
	info := &unicom.UserMobInfo{
		Usermob: usermob,
		FakeID:  u.FakeID,
		Month:   u.FakeIDMonth,
		Period:  int64(period),
	}
	if err := s.dao.InsertOrUpdateUserMobInfo(c, info); err != nil {
		log.Error("[service.InOrdersSync]日志告警 InsertOrUpdateUserMobInfo error:%v, order:%+v", err, u)
		s.cache.Do(c, func(ctx context.Context) {
			if err := s.dao.InsertOrUpdateUserMobInfo(c, info); err != nil {
				log.Error("[service.InOrdersSync]日志告警 重试InsertOrUpdateUserMobInfo error:%v, order:%+v", err, u)
			}
		})
		return err
	}
	return nil
}

// InAdvanceSync insert AdvanceSync
func (s *Service) InAdvanceSync(c context.Context, usermob, ip string, u *unicom.UnicomJson, now time.Time) (err error) {
	if !s.iplimit(_unicomKey, ip) {
		err = ecode.AccessDenied
		return
	}
	var result int64
	if result, err = s.dao.InAdvanceSync(c, usermob, u, now); err != nil || result == 0 {
		log.Error("unicom_s.dao.InAdvanceSync (%v,%v,%v,%v,%v,%v,%v,%v) error(%v) or result==0",
			usermob, u.Userphone, u.Cpid, u.Spid, u.Ordertypes, u.Channelcode, u.Province, u.Area, err)
	}
	return
}

// FlowSync update OrdersSync
func (s *Service) FlowSync(c context.Context, flowbyte int, usermob, time, ip string, now time.Time) (err error) {
	if !s.iplimit(_unicomKey, ip) {
		err = ecode.AccessDenied
		return
	}
	var result int64
	if result, err = s.dao.FlowSync(c, flowbyte, usermob, time, now); err != nil || result == 0 {
		log.Error("unicom_s.dao.OrdersSync(%v, %v, %v) error(%v) or result==0", usermob, time, flowbyte, err)
	}
	return
}

// InIPSync
func (s *Service) InIPSync(c context.Context, ip string, u *unicom.UnicomIpJson, now time.Time) (err error) {
	if !s.iplimit(_unicomKey, ip) {
		err = ecode.AccessDenied
		return
	}
	var result int64
	if result, err = s.dao.InIPSync(c, u, now); err != nil {
		log.Error("s.dao.InIpSync(%s,%s) error(%v)", u.Ipbegin, u.Ipend, err)
	} else if result == 0 {
		err = ecode.RequestErr
		log.Error("unicom_s.dao.InIpSync(%s,%s) error(%v) result==0", u.Ipbegin, u.Ipend, err)
	}
	return
}

func (s *Service) UserFlow(ctx context.Context, usermob, mobiApp, ip string, build int, now time.Time) (*unicom.Unicom, error) {
	orderm, err := s.orders(ctx, usermob, now)
	if err != nil {
		return nil, err
	}
	order, err := s.orderState(orderm, "", now)
	if err != nil {
		return nil, err
	}
	return order, nil
}

// nolint:gocognit
func (s *Service) orders(ctx context.Context, usermob string, now time.Time) (map[unicom.FreeProduct]*unicom.Unicom, error) {
	cached := true
	infos, err := s.dao.UnicomCache(ctx, usermob)
	if err != nil {
		log.Error("usermob:%s,error:%+v", usermob, err)
		cached = false
	}
	if len(infos) != 0 {
		s.pHit.Incr("unicoms_cache")
	}
	if len(infos) == 0 {
		if infos, err = s.dao.OrdersUserFlow(ctx, usermob); err != nil {
			return nil, err
		}
		if len(infos) == 0 {
			return nil, nil
		}
		// TODO 需要考虑缓存穿透的情况
		s.pMiss.Incr("unicoms_cache")
		if cached {
			s.cache.Do(ctx, func(ctx context.Context) {
				if err := s.dao.AddUnicomCache(ctx, usermob, infos); err != nil {
					log.Error("%+v", err)
				}
			})
		}
	}
	orderFunc := func(infos []*unicom.Unicom) *unicom.Unicom {
		if len(infos) == 0 {
			return nil
		}
		// endtime  卡有可能为0，包一定不为0
		// 有 endtime  and < now.time 映射老的状态 type=0 有效订单，已生效，和待生效
		// 有 endtime  and > now.time 映射老的状态 type=1 退订订单
		// 没有 endtime  映射老的状态 type=0，已生效，和待生效，只有卡
		var (
			effectiveOrder *unicom.Unicom
			cancelOrder    *unicom.Unicom
		)
		for _, u := range infos {
			// 还未失效的订单（卡退订，包退订或者自动到期）|| 生效的卡
			// 订单时间 理论上不为0
			if u == nil || u.Ordertime <= 0 {
				continue
			}
			orderTime := u.Ordertime.Time()
			var endTime time.Time
			if u.Endtime > 0 {
				endTime = u.Endtime.Time()
			}
			if endTime.IsZero() || endTime.After(now) { // 有效订购关系
				// 找最早的有效订购关系
				if effectiveOrder != nil && orderTime.After(effectiveOrder.Ordertime.Time()) {
					continue
				}
				effectiveOrder = &unicom.Unicom{}
				*effectiveOrder = *u
				if endTime.IsZero() { // 有效订单（卡+包）
					effectiveOrder.TypeInt = 0
				} else { // 退订了但没有过期的订单（卡+包）
					effectiveOrder.TypeInt = 1
				}
			} else { // 失效
				// 找最近的失效订购关系
				if cancelOrder != nil && orderTime.Before(cancelOrder.Ordertime.Time()) {
					continue
				}
				cancelOrder = &unicom.Unicom{}
				*cancelOrder = *u
				cancelOrder.TypeInt = 1
			}
		}
		// 无最早的有效订购关系
		if effectiveOrder == nil {
			// 找最近的失效订购关系
			if cancelOrder == nil {
				return nil
			}
			effectiveOrder = cancelOrder
		}
		return effectiveOrder
	}
	var (
		cardOrders []*unicom.Unicom
		flowOrders []*unicom.Unicom
	)

	for _, info := range infos {
		spid := strconv.Itoa(info.Spid)
		for _, product := range s.c.Unicom.CardProduct {
			if spid == product.Spid {
				order := &unicom.Unicom{}
				*order = *info
				// 用来判断是否是免流卡
				order.ProductType = 1
				order.CardType = product.Type
				order.Desc = product.Desc
				order.Flowtype = 1
				order.ProductTag = product.Tag
				order.TfWay = "ip"
				order.TfType = 1
				cardOrders = append(cardOrders, order)
			}
		}
		for _, product := range s.c.Unicom.FlowProduct {
			if spid == product.Spid {
				order := &unicom.Unicom{}
				*order = *info
				order.ProductType = 2
				order.CardType = product.Type
				order.Desc = product.Desc
				// 这里有历史问题，联通的包和卡是两个团队开发的
				// 包团队是cdn免流方式
				// 卡团队是ip免流方式
				// s10的免流包是卡团队开发的
				// 联通s10免流包是披着免流卡的包
				// flowtype 免流类型：1：免流卡，2：免流包
				// 和客户端对接后，确认激活接口依赖的字段和值：
				// cardtype : 5（双端确认不会用的，但是服务端还是下发一个值）
				// flowtype : 1（联通卡团队开发的s10免流包，走ip调度，客户端需要走卡逻辑）
				// product_tag : "s10" （给播放面板的toast用，标识是s10免流包）
				// desc : "s10免流包"
				flowtype := 2
				order.TfWay = "cdn"
				if product.Way == "ip" {
					order.TfWay = "ip"
					flowtype = 1
				}
				order.TfType = 2
				order.Flowtype = flowtype
				order.ProductTag = product.Tag
				flowOrders = append(flowOrders, order)
			}
		}
	}
	flowOrder := orderFunc(flowOrders)
	cardOrder := orderFunc(cardOrders)
	res := map[unicom.FreeProduct]*unicom.Unicom{}
	if flowOrder != nil {
		if flowOrder.Canceltime.Time().IsZero() {
			flowOrder.Canceltime = 0
		}
		if flowOrder.Endtime.Time().IsZero() {
			flowOrder.Endtime = 0
		}
		res[unicom.FlowProduct] = flowOrder
	}
	if cardOrder != nil {
		if cardOrder.Canceltime.Time().IsZero() {
			cardOrder.Canceltime = 0
		}
		if cardOrder.Endtime.Time().IsZero() {
			cardOrder.Endtime = 0
		}
		res[unicom.CardProduct] = cardOrder
	}
	b, _ := json.Marshal(res)
	log.Info("unicom orders usermob:%s,result:%s", usermob, b)
	return res, nil
}

func (s *Service) UserState(ctx context.Context, usermob, mobiApp, ip string, build int, now time.Time) (*unicom.Unicom, error) {
	orderm, err := s.orders(ctx, usermob, now)
	if err != nil {
		return nil, err
	}
	order, err := s.orderState(orderm, "", now)
	if err != nil {
		return nil, err
	}
	return order, nil
}

// UnicomState
func (s *Service) UnicomState(ctx context.Context, usermob, mobiApp, ip string, build int, now time.Time) *unicom.Unicom {
	res := &unicom.Unicom{Unicomtype: 1}
	defer func() {
		log.Info("UnicomStateM usermob:%v res:%+v", usermob, res)
	}()
	orderm, err := s.orders(ctx, usermob, now)
	if err != nil {
		log.Error("UnicomStateM orders usermob:%v,error:%+v", usermob, err)
		return res
	}
	order, err := s.orderState(orderm, "", now)
	if err != nil {
		if err == xecode.AppWelfareClubNotFree {
			res.Unicomtype = 1
		}
		if err == xecode.AppWelfareClubCancelOrExpire {
			res.Unicomtype = 3
		}
		return res
	}
	// 卡的激活状态 1：未激活、2：已激活、3：已退订（过期）、4：已退订（未过期）
	*res = *order
	if res.TypeInt == 0 {
		res.Unicomtype = 2
		return res
	}
	if res.TypeInt == 1 {
		res.Unicomtype = 4
		return res
	}
	return res
}

func (s *Service) UserFlowState(c context.Context, usermob string, now time.Time) *unicom.Unicom {
	res := &unicom.Unicom{Unicomtype: 1}
	orderm, err := s.orders(c, usermob, now)
	if err != nil {
		log.Error("%+v", err)
		return res
	}
	order, err := s.orderState(orderm, "", now)
	if err != nil {
		log.Error("UserFlowState orderState usermob:%v,error%+v", usermob, err)
		return res
	}
	*res = *order
	if res.TypeInt == 1 {
		res.Unicomtype = 3
		return res
	}
	if res.TypeInt == 0 {
		res.Unicomtype = 2
		return res
	}
	return res
}

func (s *Service) orderState(orderm map[unicom.FreeProduct]*unicom.Unicom, product unicom.FreeProduct, now time.Time) (*unicom.Unicom, error) {
	// 解决用户既是联通卡，又订购了联通包的情况
	// 原来运营商沟通，卡和包只能二选一
	// 现在发现有同时订购的情况，觉得从逻辑上区分开
	switch product {
	case unicom.FlowProduct, unicom.CardProduct:
		order, ok := orderm[product]
		if !ok {
			return nil, xecode.AppWelfareClubNotFree
		}
		// 用户退订了,且过期
		if order.TypeInt == 1 && order.Endtime.Time().Before(now) {
			return nil, xecode.AppWelfareClubCancelOrExpire
		}
		return order, nil
	default:
		// 优先级 卡产品>包产品
		var orders []*unicom.Unicom
		if order, ok := orderm[unicom.CardProduct]; ok {
			orders = append(orders, order)
		}
		if order, ok := orderm[unicom.FlowProduct]; ok {
			orders = append(orders, order)
		}
		if len(orders) == 0 {
			return nil, xecode.AppWelfareClubNotFree
		}
		for _, order := range orders {
			// 正常订单
			if order.TypeInt == 0 {
				return order, nil
			}
			// 退订了，没有过期
			if order.TypeInt == 1 && order.Endtime.Time().After(now) {
				return order, nil
			}
		}
		return nil, xecode.AppWelfareClubCancelOrExpire
	}
}

// IsUnciomIP is unicom ip
func (s *Service) IsUnciomIP(ipUint uint32, ipStr, mobiApp string, build int, now time.Time) (err error) {
	if !model.IsIPv4(ipStr) {
		err = ecode.NothingFound
		return
	}
	isValide := s.unciomIPState(ipUint)
	if isValide {
		return
	}
	err = ecode.NothingFound
	return
}

// UserUnciomIP
func (s *Service) UserUnciomIP(ipUint uint32, ipStr, usermob, mobiApp string, build int, now time.Time) (res *unicom.UnicomUserIP) {
	res = &unicom.UnicomUserIP{
		IPStr:    ipStr,
		IsValide: false,
	}
	if !model.IsIPv4(ipStr) {
		return
	}
	if res.IsValide = s.unciomIPState(ipUint); !res.IsValide {
		log.Error("unicom_user_ip:%v unicom_ip_usermob:%v", ipStr, usermob)
	}
	return
}

func (s *Service) Order(ctx context.Context, usermobDes, channel string, ordertype int, now time.Time) (*unicom.BroadbandOrder, string, error) {
	usermob, err := s.decrypt(usermobDes)
	if err != nil {
		return nil, "", err
	}
	orderm, err := s.orders(ctx, usermob, now)
	if err != nil {
		return nil, "", err
	}
	order, _ := s.orderState(orderm, unicom.CardProduct, now)
	if order != nil && order.TypeInt == 0 {
		return nil, "", xecode.AppWelfareClubFlowOrderForbidden
	}
	res, msg, err := s.dao.Order(ctx, usermobDes, channel, ordertype)
	if err != nil {
		log.Error("s.dao.Order usermobDes(%v) error(%v)", usermobDes, err)
		return nil, msg, err
	}
	return res, "", nil
}

// CancelOrder unicom user cancel order
// nolint:gomnd
func (s *Service) CancelOrder(ctx context.Context, usermobDes string, now time.Time) (*unicom.BroadbandOrder, string, error) {
	usermob, err := s.decrypt(usermobDes)
	if err != nil {
		return nil, "", err
	}
	log.Warn("CancelOrder usermobDes:%v usermob:%v", usermobDes, usermob)
	orderm, err := s.orders(ctx, usermob, now)
	if err != nil {
		return nil, "", err
	}
	order, err := s.orderState(orderm, unicom.FlowProduct, now)
	if err != nil {
		return nil, "", err
	}
	if order.Flowtype != 2 {
		return nil, "", xecode.AppWelfareClubOrderCancelFailed
	}
	res, msg, err := s.dao.CancelOrder(ctx, usermobDes, order.Spid)
	if err != nil {
		log.Error("s.dao.CancelOrder usermobDes:%v usermob:%v error:%+v", usermobDes, usermob, err)
		return nil, msg, err
	}
	return res, "", nil
}

// UnicomSMSCode unicom sms code
func (s *Service) UnicomSMSCode(c context.Context, phone string, now time.Time) (msg string, err error) {
	if msg, err = s.dao.SendSmsCode(c, phone); err != nil {
		log.Error("s.dao.SendSmsCode phone(%v) error(%v)", phone, err)
		return
	}
	return
}

func (s *Service) BindUser(ctx context.Context, phone int, code int, mid int64, now time.Time) error {
	customCheck, _ := s.gaiaEngine.InitCheck(ctx, "unicom_welfare_rewards")
	customCheck.Put("subscene", "绑定")
	customCheck.Put("phone_num", phone)
	customCheck.Put("mid", mid)
	report, _ := customCheck.Do()
	phoneStr := strconv.Itoa(phone)
	if report != nil && report.CheckHit("reject") {
		return xecode.AppWelfareClubRejectPackExchange
	}
	usermobDes, msg, err := s.dao.SmsNumber(ctx, phoneStr, code)
	if err != nil {
		log.Error("BindUser SmsNumber phone:%s,code:%d,error:%+v", phoneStr, code, err)
		return ecode.Error(ecode.RequestErr, msg)
	}
	if usermobDes == "" {
		return xecode.AppWelfareClubActiveFailed
	}
	usermob, err := s.decrypt(usermobDes)
	if err != nil {
		log.Error("BindUser %+v", err)
		return err
	}
	orderm, err := s.orders(ctx, usermob, now)
	if err != nil {
		log.Error("BindUser %+v", err)
		return err
	}
	order, err := s.orderState(orderm, "", now)
	if err != nil {
		log.Error("BindUser %+v", err)
		return err
	}
	if order.ProductType != 1 {
		return xecode.AppWelfareClubOnlySupportCard
	}
	if _, err = s.unicomBindInfo(ctx, mid); err == nil {
		return xecode.AppWelfareClubBinded
	}
	if owner := s.unicomBindMIdByPhone(ctx, phoneStr); owner > 0 {
		return xecode.AppWelfareClubRegistered
	}
	rows, err := s.dao.BindUser(ctx, mid, phoneStr, usermob)
	if err != nil {
		log.Error("福利社绑定错误 mid:%d,phone:%s,error:%+v", mid, phoneStr, err)
		return err
	}
	if rows == 0 {
		log.Error("福利社绑定错误 mid:%d,phone:%s,error:重复绑定", mid, phoneStr)
		return ecode.Error(ecode.RequestErr, "请勿重复绑定")
	}
	log.Warn("福利社绑定成功 mid:%d,phone:%s", mid, phoneStr)
	if err := s.updateUserBind(ctx, mid); err != nil {
		log.Error("%+v", err)
		s.cache.Do(ctx, func(ctx context.Context) {
			_ = s.updateUserBind(ctx, mid)
		})
	}
	// databus
	s.addUserBindState(&unicom.UserBindInfo{MID: mid, Phone: phone, Action: "unicom_welfare_bind"})
	return nil
}

func (s *Service) UnbindUser(ctx context.Context, mid int64, phone int) error {
	customCheck, _ := s.gaiaEngine.InitCheck(ctx, "unicom_welfare_rewards")
	customCheck.Put("subscene", "解绑")
	customCheck.Put("phone_num", phone)
	customCheck.Put("mid", mid)
	report, _ := customCheck.Do()
	if report != nil && report.CheckHit("reject") {
		return xecode.AppWelfareClubRejectPackExchange
	}
	ub, err := s.unicomBindInfo(ctx, mid)
	if err != nil {
		log.Error("UnbindUser %+v", err)
		return ecode.Error(ecode.RequestErr, "用户未绑定手机号")
	}
	if ub.Phone != phone {
		return ecode.Error(ecode.RequestErr, "解绑的手机号和已绑定的不一致")
	}
	if ub.State == 0 {
		return ecode.Error(ecode.RequestErr, "请勿重复解绑")
	}
	if err := s.dao.DeleteUserBindCache(ctx, mid); err != nil {
		log.Error("UnbindUser %+v", err)
		return err
	}
	phoneStr := strconv.Itoa(phone)
	rows, err := s.dao.UnbindUser(ctx, mid, phoneStr)
	if err != nil {
		log.Error("福利社解绑错误 mid:%d,phone:%s,error:%+v", mid, phoneStr, err)
		return err
	}
	if rows == 0 {
		log.Error("福利社绑定错误 mid:%d,phone:%s,error:重复解绑", mid, phoneStr)
		return ecode.Error(ecode.RequestErr, "请勿重复解绑")
	}
	log.Warn("福利社解绑成功 mid:%d,phone:%s", mid, phoneStr)
	// databus
	s.addUserBindState(&unicom.UserBindInfo{MID: mid, Phone: phone, Action: "unicom_welfare_untied"})
	return nil
}

// unicomBindInfo unicom bind info
func (s *Service) unicomBindInfo(c context.Context, mid int64) (res *unicom.UserBind, err error) {
	if res, err = s.dao.UserBindCache(c, mid); err == nil {
		s.pHit.Incr("unicoms_userbind_cache")
	} else {
		if res, err = s.dao.UserBind(c, mid); err != nil {
			log.Error("s.dao.UserBind error(%v)", err)
			return
		}
		s.pMiss.Incr("unicoms_userbind_cache")
		if res == nil {
			err = xecode.AppWelfareClubNoBinding
			return
		}
		if err = s.dao.AddUserBindCache(c, mid, res); err != nil {
			log.Error("s.dao.AddUserBindCache mid(%d) error(%v)", mid, err)
			return
		}
	}
	return
}

func (s *Service) unicomBindMIdByPhone(c context.Context, phone string) (mid int64) {
	var err error
	if mid, err = s.dao.UserBindPhoneMid(c, phone); err != nil {
		log.Error("s.dao.UserBindPhoneMid error(%v)", phone)
		return
	}
	return
}

// UserBind user bind
func (s *Service) UserBind(c context.Context, mid int64) (res *unicom.UserBind, msg string, err error) {
	var (
		acc *account.Info
		ub  *unicom.UserBind
	)
	if acc, err = s.accd.Info(c, mid); err != nil {
		log.Error("s.accd.info error(%v)", err)
		return
	}
	res = &unicom.UserBind{
		Name: acc.Name,
		Mid:  acc.Mid,
	}
	if ub, err = s.unicomBindInfo(c, mid); err != nil {
		log.Error("UserBind userBindInfo mid:%v,error:%v", mid, err)
		err = nil
	}
	if ub != nil {
		res.Phone = ub.Phone
		res.Integral = ub.Integral
		res.Flow = ub.Flow
		res.Usermob = ub.Usermob
	}
	return
}

// UnicomPackList unicom pack list
func (s *Service) UnicomPackList(entry int) (res []*unicom.UserPack) {
	switch entry {
	case _entryComic:
		for _, v := range s.unicomPackCache {
			if v.IsComic() || v.IsTraffic() { // 漫画app只出流量包和漫读卷
				res = append(res, v)
			}
		}
	case _entryPink:
		res = s.unicomPackCache
	default:
		res = s.unicomPackCache
	}
	return
}

// UserBindLog user bind week log
func (s *Service) UserBindLog(c context.Context, mid int64, now time.Time) (res []*unicom.UserLog, err error) {
	if res, err = s.dao.SearchUserBindLog(c, mid, now); err != nil {
		log.Error("unicom s.dao.SearchUserBindLog error(%v)", err)
		return
	}
	return
}

// WelfareBindState welfare user bind state
func (s *Service) WelfareBindState(c context.Context, mid int64) (res int) {
	if ub, err := s.dao.UserBindCache(c, mid); err == nil && ub != nil {
		res = 1
	}
	return
}

// unicomPackInfo unicom pack infos
func (s *Service) unicomPackInfo(c context.Context, id int64) (res *unicom.UserPack, err error) {
	defer func() {
		if err != nil {
			res.Desc = flowDesc(res)
		}
	}()
	if res, err = s.dao.UserPackCache(c, id); err != nil {
		log.Error("%+v", err)
	}
	if err == nil {
		s.pHit.Incr("unicoms_pack_cache")
		return
	}
	if res, err = s.dao.UserPackByID(c, id); err != nil {
		log.Error("%+v", err)
		return
	}
	s.pMiss.Incr("unicoms_pack_cache")
	if res == nil {
		err = xecode.AppWelfareClubPackNotExist
		return
	}
	if err = s.dao.AddUserPackCache(c, id, res); err != nil {
		log.Error("s.dao.AddUserPackCache id(%d) error(%v)", id, err)
		return
	}
	return
}

// unciomIPState
func (s *Service) unciomIPState(ipUint uint32) (isValide bool) {
	for _, u := range s.unicomIpCache {
		if u.IPStartUint <= ipUint && u.IPEndUint >= ipUint {
			isValide = true
			break
		}
	}
	if !isValide {
		s.infoProm.Incr("unciom_ip_state_invalide")
	}
	return
}

func (s *Service) iplimit(isp, ip string) bool {
	addrs := s.c.IPLimit.Addrs[isp]
	if len(addrs) == 0 {
		return true
	}
	log.Info("isp:%v,ip limit list:%+v", isp, addrs)
	for _, addr := range addrs {
		if ip == addr {
			return true
		}
	}
	log.Error("日志告警 IP白名单检验不通过,isp:%v,ip:%v", isp, ip)
	return false
}

// DesDecrypt
func (s *Service) DesDecrypt(src, key []byte) ([]byte, error) {
	block, err := des.NewCipher(key)
	if err != nil {
		return nil, err
	}
	out := make([]byte, len(src))
	dst := out
	bs := block.BlockSize()
	if len(src)%bs != 0 {
		return nil, errors.New("crypto/cipher: input not full blocks")
	}
	for len(src) > 0 {
		block.Decrypt(dst, src[:bs])
		src = src[bs:]
		dst = dst[bs:]
	}
	out = s.zeroUnPadding(out)
	return out, nil
}

// zeroUnPadding
func (s *Service) zeroUnPadding(origData []byte) []byte {
	return bytes.TrimFunc(origData,
		func(r rune) bool {
			return r == rune(0)
		})
}

// UserPacksLog user pack logs
func (s *Service) UserPacksLog(ctx context.Context, starttime, now time.Time, start int, ip string) ([]*unicom.UserPackLog, error) {
	if env.DeployEnv == env.DeployEnvProd && !s.iplimit(_unicomPackKey, ip) {
		return nil, ecode.AccessDenied
	}
	if starttime.Month() >= now.Month() && starttime.Year() >= now.Year() {
		return []*unicom.UserPackLog{}, nil
	}
	endInt := starttime.AddDate(0, 1, -1).Day()
	if start > endInt {
		return []*unicom.UserPackLog{}, nil
	}
	endday := starttime.AddDate(0, 0, start)
	if start == endInt {
		endday = starttime.AddDate(0, 1, 0)
	}
	res, err := s.dao.UserPacksLog(ctx, endday.AddDate(0, 0, -1), endday)
	if err != nil {
		return nil, err
	}
	if len(res) == 0 {
		return []*unicom.UserPackLog{}, nil
	}
	return res, nil
}

func (s *Service) addUserPackLog(ctx context.Context, v *unicom.UserPackLog) {
	const _logID = 91
	retryFunc := func(doFunc func() error) {
		if err := doFunc(); err == nil {
			return
		}
		// nolint:biligowordcheck
		go func() {
			var err error
			for i := 0; i < 3; i++ {
				if err = doFunc(); err == nil {
					return
				}
				time.Sleep(100 * time.Millisecond)
			}
			log.Error("%+v", err)
		}()
	}
	if v.RequestNo != _defalutLogRequestNo {
		retryFunc(func() error {
			_, err := s.dao.InUserPackLog(ctx, v)
			return err
		})
	}
	userInfo := &report.UserInfo{
		Mid:      v.Mid,
		Business: _logID,
		Action:   "unicom_userpack_deduct",
		Ctime:    time.Now(),
		Content: map[string]interface{}{
			"phone":     v.Phone,
			"pack_desc": v.UserDesc,
			"integral":  (-v.Integral),
		},
	}
	retryFunc(func() error {
		return report.User(userInfo)
	})
}

func (s *Service) UserBindInfoByPhone(ctx context.Context, phone, ip string, now time.Time) (*unicom.UserBindV2, error) {
	if !s.iplimit(_unicomPackKey, ip) {
		return nil, ecode.AccessDenied
	}
	res, err := s.dao.UserBindInfoByPhone(ctx, phone)
	if err != nil {
		log.Error("UserBindInfoByPhone s.dao.UserBindInfoByPhone phone(%s) error(%v)", phone, err)
		return nil, err
	}
	orderm, err := s.orders(ctx, res.Usermob, now)
	if err != nil {
		return nil, err
	}
	order, err := s.orderState(orderm, "", now)
	if err != nil {
		return nil, err
	}
	if order.ProductType != 1 {
		return nil, xecode.AppWelfareClubOnlySupportCard
	}
	return res, nil
}

func (s *Service) AddUserBindIntegral(ctx context.Context, mids []int64, integral int, ip string) (map[int64]interface{}, error) {
	if !s.iplimit(_unicomPackKey, ip) {
		return nil, ecode.AccessDenied
	}
	max := 100
	if len(mids) > max {
		return nil, xecode.AppQueryExceededLimit
	}
	now := time.Now()
	res := map[int64]interface{}{}
	for _, mid := range mids {
		ub, err := s.unicomBindInfo(ctx, mid)
		if err != nil {
			res[mid] = map[string]interface{}{"state": 0}
			continue
		}
		orderm, err := s.orders(ctx, ub.Usermob, now)
		if err != nil {
			log.Error("AddUserBindIntegral orders mid:%v,usermob:%v,error:%+v", mid, ub.Usermob, err)
			res[mid] = map[string]interface{}{"state": 0}
			continue
		}
		order, err := s.orderState(orderm, "", now)
		if err != nil {
			log.Error("AddUserBindIntegral orderState mid:%v,usermob:%v,error:%+v", mid, ub.Usermob, err)
			res[mid] = map[string]interface{}{"state": 0}
			continue
		}
		if order.ProductType != 1 {
			res[mid] = map[string]interface{}{"state": 0}
			continue
		}
		if err = s.addUserIntegral(ctx, mid, ub.Phone, integral, ub.Usermob, now); err != nil {
			err = nil
			log.Error("AddUserBindIntegral s.updateUserIntegral mid(%d) error(%v)", mid, err)
			res[mid] = map[string]interface{}{"state": 0}
			continue
		}
		s.cache.Do(ctx, func(ctx context.Context) {
			v := &unicom.UserPackLog{
				Phone:     ub.Phone,
				Usermob:   ub.Usermob,
				Mid:       ub.Mid,
				RequestNo: _defalutLogRequestNo,
				Type:      0,
				UserDesc:  "福利点发放",
				Integral:  -integral,
			}
			s.addUserPackLog(ctx, v)
		})
		res[mid] = map[string]interface{}{"state": 1}
	}
	return res, nil
}

func (s *Service) comsumeUserIntegral(ctx context.Context, mid int64, phone int, integral int, usermob string, now time.Time) error {
	rows, err := s.dao.ConsumeUserBindIntegral(ctx, mid, strconv.Itoa(phone), integral, usermob, now)
	if err != nil {
		return err
	}
	if rows == 0 {
		return xecode.AppWelfareClubLackIntegral
	}
	if err := s.updateUserBind(ctx, mid); err != nil {
		log.Error("%+v", err)
		s.cache.Do(ctx, func(ctx context.Context) {
			_ = s.updateUserBind(ctx, mid)
		})
	}
	return nil
}

func (s *Service) consumeUserFlow(ctx context.Context, mid int64, phone int, flow int, usermob string, now time.Time) error {
	rows, err := s.dao.ConsumeUserBindFlow(ctx, mid, strconv.Itoa(phone), flow, usermob, now)
	if err != nil {
		return err
	}
	if rows == 0 {
		return xecode.AppWelfareClubLackFlow
	}
	if err := s.updateUserBind(ctx, mid); err != nil {
		log.Error("%+v", err)
		s.cache.Do(ctx, func(ctx context.Context) {
			_ = s.updateUserBind(ctx, mid)
		})
	}

	return nil
}

func (s *Service) addUserIntegral(ctx context.Context, mid int64, phone int, integral int, usermob string, now time.Time) error {
	if _, err := s.dao.AddUserBindIntegral(ctx, mid, strconv.Itoa(phone), integral, usermob, now); err != nil {
		log.Error("%+v", err)
		return err
	}
	if err := s.updateUserBind(ctx, mid); err != nil {
		log.Error("%+v", err)
		s.cache.Do(ctx, func(ctx context.Context) {
			_ = s.updateUserBind(ctx, mid)
		})
	}
	return nil
}

func (s *Service) addUserFlow(ctx context.Context, mid int64, phone int, flow int, usermob string, now time.Time) error {
	if _, err := s.dao.AddUserBindFlow(ctx, mid, strconv.Itoa(phone), flow, usermob, now); err != nil {
		log.Error("%+v", err)
		return err
	}
	if err := s.updateUserBind(ctx, mid); err != nil {
		log.Error("%+v", err)
		s.cache.Do(ctx, func(ctx context.Context) {
			_ = s.updateUserBind(ctx, mid)
		})
	}
	return nil
}

func (s *Service) updateUserBind(ctx context.Context, mid int64) error {
	res, err := s.dao.UserBind(ctx, mid)
	if err != nil {
		return err
	}
	if res == nil {
		return nil
	}
	if err = s.dao.AddUserBindCache(ctx, mid, res); err != nil {
		log.Error("%+v", err)
		return err
	}
	return nil
}

func (s *Service) forbidenPack(ctx context.Context, id int64) error {
	_, err := s.dao.SetUserPackFlow(ctx, id, 2)
	if err != nil {
		log.Error("%+v", err)
		return err
	}
	return nil
}

// nolint:gomnd
func (s *Service) decrypt(usermob string) (string, error) {
	aesKey := []byte("9ed226d9")
	bs, err := base64.StdEncoding.DecodeString(usermob)
	if err != nil {
		return "", err
	}
	if bs, err = s.DesDecrypt(bs, aesKey); err != nil {
		return "", err
	}
	usermobStr := string(bs)
	if len(bs) > 32 {
		usermobStr = string(bs[:32])
	}
	return usermobStr, nil
}
