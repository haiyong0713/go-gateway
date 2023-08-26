package unicom

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"go-common/library/net/netutil"
	"go-common/library/retry"
	"strconv"
	"strings"
	"time"

	"go-common/library/log"
	"go-common/library/queue/databus/report"
	"go-common/library/railgun"
	xecode "go-gateway/app/app-svr/app-wall/ecode"
	"go-gateway/app/app-svr/app-wall/job/model/unicom"
)

const (
	_initIPUnicomKey = "ipunicom_%v_%v"
)

func (s *Service) initClickRailGun(cfg *railgun.DatabusV1Config, pcfg *railgun.ReplaceConfig) {
	inputer := railgun.NewDatabusV1Inputer(cfg)
	processor := railgun.NewReplaceProcessor(pcfg, s.clickUnpack, s.clickPre, s.clickDo)
	g := railgun.NewRailGun("视频点击日志", nil, inputer, processor)
	s.clickRailGun = g
	g.Start()
}

func (s *Service) clickUnpack(msg railgun.Message) (replace []railgun.ReplaceUnpackMsg, err error) {
	var bss [][]byte
	if err := json.Unmarshal(msg.Payload(), &bss); err != nil {
		return nil, err
	}
	var vs []railgun.ReplaceUnpackMsg
	for _, bs := range bss {
		v, err := s.checkMsgIllegal(bs)
		if err != nil {
			log.Error("%+v", err)
			continue
		}
		if v == nil || v.MID == 0 {
			continue
		}
		vs = append(vs, railgun.ReplaceUnpackMsg{
			ReplaceKey:   strconv.FormatInt(v.MID, 10),
			PreGroup:     v.MID,
			ReplaceGroup: v.MID,
			Item:         v,
		})
	}
	return vs, nil
}

func (s *Service) clickPre(ctx context.Context, item interface{}) railgun.MsgPolicy {
	return railgun.MsgPolicyNormal
}

// click 点击日志来着免流用户和非免流用户 非免流远大于免流
// 绑定福利社的数据 60万
// 免流用户分福利社绑定用户
// 绑定福利社，并且是有效卡，每天增加一次福利点
func (s *Service) clickDo(ctx context.Context, items map[railgun.ReplaceKey]interface{}) railgun.MsgPolicy {
	const (
		_level0 = 0
		_level1 = 1
		_level2 = 2
		_level3 = 3
		_level4 = 4
		_level5 = 5
		_level6 = 6
	)
	for _, v := range items {
		cli := v.(*unicom.ClickMsg)
		// 从监控看，click 日志的qps是峰值76k
		// checkMsgIllegal 过滤了非移动端的点击
		// 过滤后的 mc qps是峰值37k
		// 总的福利社用户是74万
		// 需要过滤非卡的用户，用户有可能换套餐
		// 过滤未绑定的用户
		ub, err := s.dao.UserBindCache(ctx, cli.MID)
		if err != nil {
			continue
		}
		if ub == nil {
			continue
		}
		now := time.Now()
		// 过滤非免流卡的用户
		orderm, err := s.orders(ctx, ub.Usermob, now)
		if err != nil {
			continue
		}
		order, err := s.orderState(orderm, unicom.CardProduct, now)
		if err != nil {
			continue
		}
		// 过滤后的数据量很大的降低了
		// 这边有并发问题，需要加锁
		// 更新 mc db 失败后del锁
		key := scoreLockKey(now, ub.Mid)
		locked, err := s.lockdao.TryLock(ctx, key, s.lockExpire)
		if err != nil {
			log.Error("TryLock key(%s) %+v", key, err)
			continue
		}
		if !locked {
			log.Warn("TryLock fail key(%s)", key)
			continue
		}
		integral := 10
		var flow int
		switch cli.Lv {
		case _level0, _level1, _level2, _level3:
			flow = 10
		case _level4:
			flow = 15
		case _level5:
			flow = 20
		case _level6:
			flow = 30
		default:
			continue
		}
		rows, err := s.dao.AddUserBindScore(ctx, ub.Mid, strconv.Itoa(ub.Phone), integral, flow, ub.Usermob, now)
		if err != nil || rows == 0 {
			// rows == 0的情况是用户解绑了手机号，此时del锁
			// 尝试删除锁
			if err := retry.WithAttempts(ctx, "unlock", 3, netutil.DefaultBackoffConfig, func(ctx context.Context) error {
				return s.lockdao.UnLock(ctx, key)
			}); err != nil {
				log.Error("%+v", err)
			}
			continue
		}
		log.Info("点击视频增加福利点:%+v,order:%+v,flow:%v", ub, order, flow)
		s.addUserIntegralLog(&unicom.UserPackLog{Phone: ub.Phone, Mid: ub.Mid, Integral: 10, UserDesc: "每日礼包"})
	}
	return railgun.MsgPolicyNormal
}

func (s *Service) checkMsgIllegal(msg []byte) (click *unicom.ClickMsg, err error) {
	var (
		aid        int64
		clickMsg   []string
		plat       int64
		bvID       string
		mid        int64
		lv         int64
		ctime      int64
		stime      int64
		epid       int64
		ip         string
		seasonType int
		userAgent  string
	)
	clickMsg = strings.Split(string(msg), "\001")
	msgLen := 18
	if len(clickMsg) < msgLen {
		err = errors.New("click msg error")
		return
	}
	if aid, err = strconv.ParseInt(clickMsg[1], 10, 64); err != nil {
		err = fmt.Errorf("aid(%s) error", clickMsg[1])
		return
	}
	if aid <= 0 {
		err = fmt.Errorf("aid(%s) error", clickMsg[1])
		return
	}
	if plat, err = strconv.ParseInt(clickMsg[0], 10, 64); err != nil {
		err = fmt.Errorf("plat(%s) error", clickMsg[0])
		return
	}
	if plat != 3 && plat != 4 {
		err = fmt.Errorf("plat(%d) is not android or ios", plat)
		return
	}
	userAgent = clickMsg[10]
	bvID = clickMsg[8]
	if bvID == "" {
		err = fmt.Errorf("bvID(%s) is illegal", clickMsg[8])
		return
	}
	if clickMsg[4] != "" && clickMsg[4] != "0" {
		if mid, err = strconv.ParseInt(clickMsg[4], 10, 64); err != nil {
			err = fmt.Errorf("mid(%s) is illegal", clickMsg[4])
			return
		}
	}
	if clickMsg[5] != "" {
		if lv, err = strconv.ParseInt(clickMsg[5], 10, 64); err != nil {
			err = fmt.Errorf("lv(%s) is illegal", clickMsg[5])
			return
		}
	}
	if ctime, err = strconv.ParseInt(clickMsg[6], 10, 64); err != nil {
		err = fmt.Errorf("ctime(%s) is illegal", clickMsg[6])
		return
	}
	if stime, err = strconv.ParseInt(clickMsg[7], 10, 64); err != nil {
		err = fmt.Errorf("stime(%s) is illegal", clickMsg[7])
		return
	}
	if ip = clickMsg[9]; ip == "" {
		err = errors.New("ip is illegal")
		return
	}
	if clickMsg[17] != "" {
		if epid, err = strconv.ParseInt(clickMsg[17], 10, 64); err != nil {
			err = fmt.Errorf("epid(%s) is illegal", clickMsg[17])
			return
		}
		if clickMsg[15] != "null" {
			if seasonType, err = strconv.Atoi(clickMsg[15]); err != nil {
				err = fmt.Errorf("seasonType(%s) is illegal", clickMsg[15])
				return
			}
		}
	}
	click = &unicom.ClickMsg{
		Plat:       int8(plat),
		AID:        aid,
		MID:        mid,
		Lv:         int8(lv),
		CTime:      ctime,
		STime:      stime,
		BvID:       bvID,
		IP:         ip,
		KafkaBs:    msg,
		EpID:       epid,
		SeasonType: seasonType,
		UserAgent:  userAgent,
	}
	return
}

func scoreLockKey(now time.Time, mid int64) string {
	// key的前缀是当日在这一个月中的哪一天，按月循环
	// key的超时时间设置为略大于一天，25小时，满足当前按天设限的场景
	return fmt.Sprintf("score_lock_%d_%d", now.Day(), mid)
}

// nolint:gocognit
func (s *Service) orders(ctx context.Context, usermob string, now time.Time) (map[unicom.FreeProduct]*unicom.Unicom, error) {
	cached := true
	infos, err := s.dao.UnicomCache(ctx, usermob)
	if err != nil {
		log.Error("%+v", err)
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
			// 无最近的失效订购关系
			if cancelOrder == nil {
				return nil
			}
			effectiveOrder = cancelOrder
		}
		return effectiveOrder
	}
	var (
		flowOrders []*unicom.Unicom
		cardOrders []*unicom.Unicom
	)

	for _, info := range infos {
		spid := strconv.Itoa(info.Spid)
		for _, product := range s.c.Unicom.FlowProduct {
			if spid == product.ID {
				flowOrders = append(flowOrders, info)
			}
		}
		for _, product := range s.c.Unicom.CardProduct {
			if spid == product.ID {
				cardOrders = append(cardOrders, info)
			}
		}
	}
	flowOrder := orderFunc(flowOrders)
	cardOrder := orderFunc(cardOrders)
	res := map[unicom.FreeProduct]*unicom.Unicom{}
	if flowOrder != nil {
		res[unicom.FlowProduct] = flowOrder
	}
	if cardOrder != nil {
		res[unicom.CardProduct] = cardOrder
	}
	b, _ := json.Marshal(res)
	log.Info("unicom orders usermob:%s,result:%s", usermob, b)
	return res, nil
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
		// 用户退订了，且过期
		if order.TypeInt == 1 && order.Endtime.Time().Before(now) {
			return nil, xecode.AppWelfareClubCancelOrExpire
		}
		return order, nil
	default:
		// 优先级 联通卡>联通包
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
			// 退订了，没过期
			if order.TypeInt == 1 && order.Endtime.Time().After(now) {
				return order, nil
			}
		}
		return nil, xecode.AppWelfareClubCancelOrExpire
	}
}

// 需要分布式锁，来解决竞争问题
func (s *Service) upBindAll() {
	ctx := context.Background()
	key := monthScoreLockKey(time.Now())
	locked, err := s.lockdao.TryLock(ctx, key, s.monthLockExpire)
	if err != nil {
		log.Error("日志告警 每月礼包自动发放获取锁失败,需要手动发放,key:%s,error:%+v", key, err)
		return
	}
	if !locked {
		log.Warn("TryLock fail key(%s)", key)
		return
	}
	defer func() {
		// 尝试删除锁
		if err1 := retry.WithAttempts(ctx, "unlock", 3, netutil.DefaultBackoffConfig, func(ctx context.Context) error {
			return s.lockdao.UnLock(ctx, key)
		}); err1 != nil {
			log.Error("%+v", err1)
		}
	}()
	offset := 0
	limit := 1000
	var binds []*unicom.UserBind
	for {
		bind, err := s.dao.BindAll(ctx, offset, limit)
		if err != nil {
			log.Error("s.dao.BindAll error:%+v", err)
			time.Sleep(time.Millisecond * 500)
			continue
		}
		offset += limit
		if len(bind) == 0 {
			break
		}
		binds = append(binds, bind...)
	}
	log.Error("开始发放每月礼包")
	for _, ub := range binds {
		now := time.Now()
		if ub.Monthly.Year() == now.Year() && ub.Monthly.Month() == now.Month() {
			log.Error("已发放过每月礼包 userbind:%+v", ub)
			continue
		}
		// 过滤非免流卡的用户
		orderm, err := s.orders(ctx, ub.Usermob, now)
		if err != nil {
			continue
		}
		order, err := s.orderState(orderm, unicom.CardProduct, now)
		if err != nil {
			continue
		}
		// 从配置中获取增加的福利点数
		var integral int
		spid := strconv.Itoa(order.Spid)
		for _, product := range s.c.Unicom.CardProduct {
			if spid == product.ID {
				integral = product.Integral
				break
			}
		}
		if integral == 0 {
			log.Info("日志告警 获取不到免流卡增加的福利点配置,userbind:%+v,order:%+v,integral:%v", ub, order, integral)
			continue
		}
		// 重试
		var row int64
		if err := retry.WithAttempts(ctx, "monthly_integral", 3, netutil.DefaultBackoffConfig, func(ctx context.Context) error {
			var err1 error
			row, err1 = s.dao.AddUserBindMonthlyIntegral(ctx, ub.Mid, strconv.Itoa(ub.Phone), integral, now, ub.Usermob)
			return err1
		}); err != nil {
			log.Error("日志告警 每月礼包发放失败 mid:%v,phone:%v,integral:%v,error:%+v", ub.Mid, ub.Phone, integral, err)
			continue
		}
		if row == 0 {
			continue
		}
		log.Info("每月礼包发放成功 userbind:%+v,order:%+v,integral:%v", ub, order, integral)
		s.addUserIntegralLog(&unicom.UserPackLog{Phone: ub.Phone, Mid: ub.Mid, Integral: integral, UserDesc: "每月礼包"})
	}
	log.Warn("每月礼包发放结束")
}

func monthScoreLockKey(now time.Time) string {
	// key的前缀是当日在这一个月中的哪一天，按月循环
	// key的超时时间设置为略大于一个月，满足当前按月设限的场景
	return fmt.Sprintf("month_score_lock_%d", now.Month())
}

// loadUnicomIPOrder load unciom ip order update
func (s *Service) loadUnicomIPOrder() {
	var (
		dbips map[string]*unicom.UnicomIP
		err   error
	)
	if dbips, err = s.loadUnicomIP(context.TODO()); err != nil {
		log.Error("s.loadUnicomIP error:%+v", err)
		return
	}
	if len(dbips) == 0 {
		log.Error("db cache ip len 0")
		return
	}
	unicomIP, err := s.dao.UnicomIP(context.TODO(), time.Now())
	if err != nil {
		log.Error("s.dao.UnicomIP(%v)", err)
		return
	}
	if len(unicomIP) == 0 {
		log.Info("unicom ip orders is null")
		return
	}
	tx, err := s.dao.BeginTran(context.TODO())
	if err != nil {
		log.Error("s.dao.BeginTran error(%v)", err)
		return
	}
	for _, uip := range unicomIP {
		key := fmt.Sprintf(_initIPUnicomKey, uip.Ipbegin, uip.Ipend)
		if _, ok := dbips[key]; ok {
			delete(dbips, key)
			continue
		}
		if err = s.dao.InUnicomIPSync(tx, uip, time.Now()); err != nil {
			_ = tx.Rollback()
			log.Error("s.dao.InUnicomIPSync error(%v)", err)
			return
		}
	}
	for _, uold := range dbips {
		if _, err = s.dao.UpUnicomIP(tx, uold.Ipbegin, uold.Ipend, 0, time.Now()); err != nil {
			_ = tx.Rollback()
			log.Error("s.dao.UpUnicomIP error(%v)", err)
			return
		}
	}
	if err = tx.Commit(); err != nil {
		log.Error("tx.Commit error(%v)", err)
		return
	}
	log.Info("update unicom ip success")
}

// loadUnicomIP load unicom ip
func (s *Service) loadUnicomIP(c context.Context) (res map[string]*unicom.UnicomIP, err error) {
	var unicomIP []*unicom.UnicomIP
	unicomIP, err = s.dao.IPSync(c)
	if err != nil {
		log.Error("s.dao.IPSync error(%v)", err)
		return
	}
	tmp := map[string]*unicom.UnicomIP{}
	for _, u := range unicomIP {
		key := fmt.Sprintf(_initIPUnicomKey, u.Ipbegin, u.Ipend)
		tmp[key] = u
	}
	res = tmp
	log.Info("loadUnicomIPCache success")
	return
}

func (s *Service) addUserPackLog(v *unicom.UserPackLog) {
	if err := retry.WithAttempts(context.Background(), "user_pack_log", 3, netutil.DefaultBackoffConfig, func(ctx context.Context) error {
		_, err1 := s.dao.InUserPackLog(context.Background(), v)
		return err1
	}); err != nil {
		log.Error("日志告警 addUserPackLog failed,log:%v,error:%+v", v, err)
	}
}

func (s *Service) addUserIntegralLog(v *unicom.UserPackLog) {
	const logID = 91
	if err := retry.WithAttempts(context.Background(), "user_integral_log", 3, netutil.DefaultBackoffConfig, func(ctx context.Context) error {
		return report.User(&report.UserInfo{
			Mid:      v.Mid,
			Business: logID,
			Action:   "unicom_userpack_add",
			Ctime:    time.Now(),
			Content: map[string]interface{}{
				"phone":     v.Phone,
				"pack_desc": v.UserDesc,
				"integral":  v.Integral,
			},
		})
	}); err != nil {
		log.Error("日志告警 addUserIntegralLog failed,log:%v,error:%+v", v, err)
	}
}
