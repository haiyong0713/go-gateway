package telecom

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"go-common/library/ecode"
	"go-common/library/log"
	"go-common/library/queue/databus/actionlog"
	"go-common/library/sync/errgroup.v2"
	xecode "go-gateway/app/app-svr/app-wall/ecode"
	"go-gateway/app/app-svr/app-wall/interface/model/telecom"

	"github.com/pkg/errors"
)

const _timeLayout = "2006-01-02 15:04:05"

func (s *Service) InCardOrderSync(c context.Context, t *telecom.CardOrderJson, ip string) (err error) {
	if !s.iplimit(ip) {
		err = ecode.AccessDenied
		return
	}
	if err = s.dao.InCardOrderSync(c, t.Biz); err != nil {
		log.Error("telecom s.dao.InCardOrderSync error(%v)", err)
		return
	}
	phone, _ := strconv.Atoi(t.Biz.Phone)
	if err = s.dao.DeleteTelecomCardCache(c, phone); err != nil {
		log.Error("telecom s.dao.DeleteTelecomCardCache phone(%d) error(%v)", phone, err)
	}
	return
}

func (s *Service) CardOrder(c context.Context, phone int) (res *telecom.CardOrder, err error) {
	res, err = s.orderlist(c, phone)
	return
}

func (s *Service) orderlist(ctx context.Context, phone int) (*telecom.CardOrder, error) {
	requestNo, err := s.seqdao.SeqID(ctx)
	if err != nil {
		log.Error("orderlist SeqID phone:%v,error:%+v", phone, err)
		return nil, err
	}
	var (
		state    bool
		stateErr error
		auth     *telecom.CardAuth
		authMsg  string
		authErr  error
	)
	g := errgroup.WithContext(ctx)
	g.Go(func(ctx context.Context) error {
		if state, stateErr = s.dao.ActiveState(ctx, strconv.Itoa(phone)); err != nil {
			log.Error("orderlist ActiveState phone:%v,error:%+v", phone, stateErr)
		}
		return nil
	})
	g.Go(func(ctx context.Context) error {
		if auth, authErr = s.dao.PhoneAuth(ctx, requestNo, phone); authErr != nil {
			log.Error("orderlist PhoneAuth phone:%v,error:%+v", phone, authErr)
		}
		return nil
	})
	if err := g.Wait(); err != nil {
		return nil, err
	}
	log.Info("电信激活接口结果,phone:%v,stateErr:%+v,auth:%+v,authMsg:%v,authErr:%+v", phone, stateErr, auth, authMsg, authErr)
	// 先判断电信派卡
	res := &telecom.CardOrder{OrderState: 1}
	if state {
		// 老接口不支持 product_type 扩展，只能复用3
		res.ProductType = 3
		res.Desc = "青年一派卡"
		res.OrderState = 2 // 激活状态 1：未激活、2：已激活
		return res, nil
	}
	if authErr != nil {
		return nil, authErr
	}
	res.CardAuthChange(auth)
	return res, nil
}

func (s *Service) PhoneCardCode(c context.Context, phone int, captcha string, now time.Time) (res *telecom.CardOrder, err error) {
	var (
		captchaStr string
	)
	captchaStr, err = s.dao.PhoneCode(c, phone)
	if err != nil {
		log.Error("telecom_s.dao.PhoneCode error (%v)", err)
		err = xecode.AppVerificationExpired
		return
	}
	if captchaStr == "" || captchaStr != captcha {
		err = xecode.AppVerificationError
		return
	}
	if res, err = s.orderlist(c, phone); err != nil {
		return
	}
	switch res.OrderState {
	case 1:
		err = xecode.AppFlowNotOrdered
		// case 3:
		// 	err = ecode.AppFlowExpired
	}
	return
}

// PhoneCardSendSMS
func (s *Service) PhoneCardSendSMS(c context.Context, phone int) (err error) {
	if err = s.sendSMS(c, phone, s.smsTemplate); err != nil {
		log.Error("sendSMS error(%v)", err)
		return
	}
	return
}

// VipPackLog vip pack log
func (s *Service) VipPackLog(c context.Context, starttime, now time.Time, startDay int, ip string) (res []*telecom.CardVipLog, err error) {
	if !s.iplimit(ip) {
		err = ecode.AccessDenied
		return
	}
	var (
		endday time.Time
	)
	if starttime.Month() >= now.Month() && starttime.Year() >= now.Year() {
		res = []*telecom.CardVipLog{}
		return
	}
	if endInt := starttime.AddDate(0, 1, -1).Day(); startDay > endInt {
		res = []*telecom.CardVipLog{}
		return
	} else if startDay == endInt {
		endday = starttime.AddDate(0, 1, 0)
	} else {
		endday = starttime.AddDate(0, 0, startDay)
	}
	if res, err = s.dao.VipPackLog(c, endday.AddDate(0, 0, -1), endday); err != nil {
		log.Error("user pack logs s.dao.UserPacksLog error(%v)", err)
		return
	}
	if len(res) == 0 {
		res = []*telecom.CardVipLog{}
	}
	return
}

func (s *Service) ActiveState(ctx context.Context, phone int, captcha string) (*telecom.ActiveState, error) {
	if captcha != "" {
		exceptCaptcha, err := s.dao.PhoneCode(ctx, phone)
		if err != nil {
			return nil, errors.Wrapf(xecode.AppVerificationExpired, "%s", err)
		}
		if exceptCaptcha != captcha {
			return nil, xecode.AppVerificationError
		}
	}
	requestNo, err := s.seqdao.SeqID(ctx)
	if err != nil {
		return nil, err
	}
	var (
		state, auth       *telecom.ActiveState
		stateErr, authErr error
	)
	g := errgroup.WithContext(ctx)
	g.Go(func(ctx context.Context) error {
		ok, err := s.dao.ActiveState(ctx, strconv.Itoa(phone))
		if err != nil {
			stateErr = err
			return nil
		}
		if ok {
			for _, product := range s.c.Telecom.CardProduct {
				if product.ID == "qnyp" { // 青年一派
					state = &telecom.ActiveState{
						ProductDesc: product.Desc,
						ProductTag:  product.Tag,
						ProductType: product.Type,
						TfType:      1,
						TfWay:       "ip",
					}
					return nil
				}
			}
		}
		return nil
	})
	g.Go(func(ctx context.Context) error {
		res, err := s.dao.PhoneAuth(ctx, requestNo, phone)
		if err != nil {
			authErr = err
			return nil
		}
		if res == nil || !res.Result {
			return nil
		}
		for _, info := range res.CombosInfo {
			startTime, err := time.ParseInLocation(_timeLayout, info.StartTime, time.Local)
			if err != nil {
				log.Error("s.ActiveState telecom parse startTime:%s, error:%+v", info.StartTime, err)
				continue
			}
			endTime, err := time.ParseInLocation(_timeLayout, info.EndTime, time.Local)
			if err != nil {
				log.Error("s.ActiveState telecom parse endTime:%s, error:%+v", info.EndTime, err)
				continue
			}
			for _, product := range s.c.Telecom.CardProduct {
				if time.Now().Before(startTime) || time.Now().After(endTime) {
					log.Error("s.ActiveState product invalid, nowTime:%v, startTime:%v, endTime:%v", time.Now(), startTime, endTime)
					continue
				}
				if info.Nbr == product.ID {
					auth = &telecom.ActiveState{
						ProductID:   product.ID,
						ProductDesc: product.Desc,
						ProductTag:  product.Tag,
						ProductType: product.Type,
						TfType:      1,
						TfWay:       "ip",
					}
					return nil
				}
			}
		}
		return nil
	})
	if err := g.Wait(); err != nil {
		return nil, err
	}
	if state != nil {
		return state, nil
	}
	if auth != nil {
		return auth, nil
	}
	if stateErr != nil {
		return nil, stateErr
	}
	if authErr != nil {
		return nil, authErr
	}
	return nil, errors.WithStack(ecode.Error(xecode.AppWelfareClubOrderFailed, xecode.OrderFailedMessage))
}

func (s *Service) UserActiveLog(param *telecom.UserActiveParam, v *telecom.ActiveState, err error) {
	const (
		_bunissID = 93
		_action   = "active"
	)
	suggest := func() string {
		if err != nil {
			return fmt.Sprintf("电信非免流用户。失败原因：%s", err)
		}
		if v != nil && v.TfType == 0 {
			return "电信非免流用户"
		}
		return "电信免流用户"
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
