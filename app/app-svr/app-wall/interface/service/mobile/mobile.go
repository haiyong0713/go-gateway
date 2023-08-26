package mobile

import (
	"context"
	"fmt"
	"time"

	"go-common/library/ecode"
	"go-common/library/log"
	"go-common/library/queue/databus/actionlog"
	xecode "go-gateway/app/app-svr/app-wall/ecode"
	"go-gateway/app/app-svr/app-wall/interface/model/mobile"

	"github.com/pkg/errors"
)

const (
	_mobileKey = "mobile"
)

// InOrdersSync insert OrdersSync
func (s *Service) InOrdersSync(c context.Context, ip string, u *mobile.MobileXML, now time.Time) (err error) {
	if !s.iplimit(_mobileKey, ip) {
		err = ecode.AccessDenied
		return
	}
	var result int64
	if result, err = s.dao.InOrdersSync(c, u); err != nil || result == 0 {
		log.Error("InOrdersSync(%+v) error:%+v or result==0", u, err)
	}
	return
}

// FlowSync update OrdersSync
func (s *Service) FlowSync(c context.Context, u *mobile.MobileXML, ip string) (err error) {
	if !s.iplimit(_mobileKey, ip) {
		err = ecode.AccessDenied
		return
	}
	var result int64
	if result, err = s.dao.FlowSync(c, u); err != nil || result == 0 {
		log.Error("FlowSync(%+v) error:%+v or result==0", u, err)
	}
	return
}

func (s *Service) Activation(c context.Context, usermob string, now time.Time) error {
	res, err := s.mobileInfo(c, usermob, now)
	if err != nil {
		return err
	}
	if len(res) == 0 {
		return xecode.AppFlowNotOrdered
	}
	for _, u := range res {
		if u.Actionid == 1 || (u.Actionid == 2 && now.Unix() <= int64(u.Expiretime)) {
			return nil
		}
	}
	return xecode.AppFlowExpired
}

func (s *Service) MobileState(ctx context.Context, usermob string, now time.Time) *mobile.Mobile {
	orders, err := s.mobileInfo(ctx, usermob, now)
	if err != nil {
		log.Error("%+v", err)
	}
	return s.userState(orders, now)
}

func (s *Service) UserMobileState(ctx context.Context, usermob string, now time.Time) *mobile.Mobile {
	orders, err := s.mobileInfo(ctx, usermob, now)
	if err != nil {
		log.Error("%+v", err)
		return &mobile.Mobile{MobileType: 1}
	}
	for _, order := range orders {
		if order.Actionid == 1 || (order.Actionid == 2 && now.Unix() <= int64(order.Expiretime)) {
			res := &mobile.Mobile{}
			*res = *order
			res.MobileType = 2
			return res
		}
	}
	return &mobile.Mobile{MobileType: 1}
}

// userState
func (s *Service) userState(orders []*mobile.Mobile, now time.Time) *mobile.Mobile {
	res := &mobile.Mobile{}
	if len(orders) == 0 {
		res.MobileType = 1
		return res
	}
	for _, order := range orders {
		// 对未知的productid过滤，未知的productid没有product_type
		if order == nil || order.ProductType == 0 {
			continue
		}
		*res = *order
		if order.Actionid == 2 && now.Unix() <= int64(order.Expiretime) {
			res.MobileType = 4
			break
		}
		if order.Actionid == 1 {
			res.MobileType = 2
			break
		}
	}
	// 用户订购状态 1：未激活、2：已激活、3：已退订（过期）、4：已退订（未过期）
	if res.MobileType == 0 {
		res.MobileType = 3
	}
	return res
}

func (s *Service) mobileInfo(ctx context.Context, usermob string, now time.Time) ([]*mobile.Mobile, error) {
	if usermob == "" {
		return nil, nil
	}
	addCache := true
	infos, err := s.dao.MobileCache(ctx, usermob)
	if err != nil {
		log.Error("%+v", err)
		addCache = false
	}
	if len(infos) != 0 {
		s.pHit.Incr("mobile_cache")
	}
	if len(infos) == 0 {
		if infos, err = s.dao.OrdersUserFlow(ctx, usermob); err != nil {
			return nil, err
		}
		if len(infos) == 0 {
			return nil, nil
		}
		// TODO 需要考虑缓存穿透的情况，老模型对字段进行了隐藏处理
		// 使用 infos[0].ID 这种方式会有问题
		// 考虑 mc迁移到redis后再做修改
		s.pMiss.Incr("mobile_cache")
		if addCache {
			s.cache.Do(ctx, func(ctx context.Context) {
				if err := s.dao.AddMobileCache(ctx, usermob, infos); err != nil {
					log.Error("%+v", err)
				}
			})
		}
	}
	// ProductID	套餐名称	是否生效
	// 100000000028	哔哩哔哩9元话费3GB流量包	否
	// 100000000030	哔哩哔哩24元话费30GB流量包	否
	// 100000001142	9元15GBbilibili定向流量包	是
	// 100000001143	0元15GB bilibili定向流量包	否
	// 100000001144	9元15GB bilibili定向流量包(首月返)	否
	// 100000001145	9元15GB bilibili定向流量包(三月返)	否
	// 100000001146	9元15GB bilibili定向流量包(首三月1元/月)	否
	// 100000001199	bilibili折扣流量包	是
	// 100000001278	哔哩哔哩随心看会员-折扣包	是
	// 100000001277	哔哩哔哩随心看会员	是
	// 100000001272	无	是/否
	// 100000001271	无	是/否
	// 300000000450	 花卡专享哔哩哔哩定向流量权益包(270)	是
	// 100000001178	花卡专享哔哩哔哩定向流量权益包	是
	var res []*mobile.Mobile
	for _, info := range infos {
		// 修改关于订单是否生效的逻辑判断。
		// 1.订单 actionid为2，表示退订，此时，不需要判断当前时间大于生效时间，只需要判断当前时间小于过期时间就表示订单有效。
		// 2.订单 actionid为1，表示订购，此时，需要判断当前时间大于生效时间，如果失效时间不为空，就需要判断当前时间小于失效时间。
		if info.Actionid == 1 {
			// 过滤未生效的订购订单
			if int64(info.Effectivetime) > now.Unix() {
				continue
			}
			if info.Expiretime > 0 && int64(info.Expiretime) < now.Unix() {
				continue
			}
		}
		order := &mobile.Mobile{}
		*order = *info
		for _, product := range s.c.Mobile.FlowProduct {
			if order.Productid == product.ID {
				order.ProductID = order.Productid
				order.ProductType = 1
				order.Desc = product.Desc
				order.ProductTag = product.Tag
			}
		}
		for _, product := range s.c.Mobile.CardProduct {
			if order.Productid == product.ID {
				order.ProductID = order.Productid
				order.ProductType = 2
				order.Desc = product.Desc
				order.ProductTag = product.Tag
			}
		}
		order.Productid = ""
		res = append(res, order)
	}
	return res, nil
}

// nolint:gomnd
func (s *Service) ActiveState(ctx context.Context, mid int64, usermob string, now time.Time) (*mobile.ActiveState, error) {
	orders, err := s.mobileInfo(ctx, usermob, now)
	if err != nil {
		return nil, err
	}
	if len(orders) == 0 {
		return nil, errors.WithStack(ecode.Error(xecode.AppWelfareClubOrderFailed, xecode.OrderFailedMessage))
	}
	for _, order := range orders {
		// 对未知的productid过滤，未知的productid没有product_type
		if order == nil || order.ProductType == 0 {
			continue
		}
		if order.Actionid == 1 || (order.Actionid == 2 && now.Unix() <= int64(order.Expiretime)) {
			res := &mobile.ActiveState{}
			res.ProductID = order.ProductID
			switch order.ProductType {
			case 2:
				res.TfType = 1 // 免流类型：0-不免流，1-免流卡，2-免流包
			case 1:
				res.TfType = 2
			}
			res.TfWay = "ip"
			res.ProductDesc = order.Desc
			res.ProductTag = order.ProductTag
			res.ProductType = order.ProductType
			return res, nil
		}
	}
	return nil, errors.WithStack(ecode.Error(xecode.AppWelfareClubOrderFailed, xecode.OrderFailedMessage))
}

func (s *Service) UserActiveLog(param *mobile.UserActiveParam, v *mobile.ActiveState, err error) {
	const (
		_bunissID = 93
		_action   = "active"
	)
	suggest := func() string {
		if err != nil {
			return fmt.Sprintf("移动非免流用户。失败原因：%s", err)
		}
		if v != nil && v.TfType == 0 {
			return "移动非免流用户"
		}
		return "移动免流用户"
	}()
	uInfo := &actionlog.UserInfo{
		Business: _bunissID,
		Mid:      param.Mid,
		Platform: param.Platform,
		Build:    param.Build,
		Buvid:    param.Buvid,
		Action:   _action,
		Ctime:    time.Now(),
		IP:       param.IP,
		Content: map[string]interface{}{
			"suggest":  suggest,
			"param":    param,
			"response": v,
			"error":    ecode.Cause(err).Code(),
		},
	}
	retryFunc := func(doFunc func() error) error {
		var err error
		for i := 0; i < 3; i++ {
			if err = doFunc(); err == nil {
				return nil
			}
			time.Sleep(100 * time.Millisecond)
		}
		return err
	}
	s.cache.Do(context.Background(), func(ctx context.Context) {
		if err := retryFunc(func() error {
			return actionlog.User(uInfo)
		}); err != nil {
			log.Error("actionlog.User data:%+v,error:%+v", uInfo, err)
		}
	})
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
