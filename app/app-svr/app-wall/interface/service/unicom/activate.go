package unicom

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"go-common/library/ecode"
	log "go-common/library/log"
	"go-common/library/queue/databus/actionlog"
	"go-common/library/sync/errgroup.v2"
	xecode "go-gateway/app/app-svr/app-wall/ecode"
	"go-gateway/app/app-svr/app-wall/interface/model/unicom"

	"github.com/pkg/errors"
)

const (
	_monthLen     = 6 // 时间month最大长度
	_minFakeIdLen = 7 // 原始fakeID长度 + 时间month最大长度
)

func (s *Service) Activate(ctx context.Context, pip, ip string, mid int64) *unicom.Activate {
	result, err := s.dao.Activate(ctx, pip, ip)
	if err != nil {
		log.Error("activate mid:%d,ip:%s,pip:%s,error:%+v", mid, ip, pip, err)
		return &unicom.Activate{Usertype: 2}
	}
	if result.Result != "0" {
		return &unicom.Activate{Usertype: 2}
	}
	switch result.Flag {
	case "0":
		res := &unicom.Activate{}
		res.Usertype = 1
		// 90157638 哔哩哔哩22卡
		// 90157639 哔哩哔哩33卡
		// 90157799 哔哩哔哩小电视卡
		// 90987322 s10免流包
		// 联通接口暂时不支持免流包
		for _, product := range s.c.Unicom.CardProduct {
			if result.Product == product.ID {
				res.Cardtype = product.Type
				res.Desc = product.Desc
				res.Flowtype = 1
			}
		}
		for _, product := range s.c.Unicom.FlowProduct {
			if result.Product == product.ID {
				res.Cardtype = product.Type
				res.Desc = product.Desc
				res.ProductTag = product.Tag
				flowtype := 2
				if product.Way == "ip" {
					flowtype = 1
				}
				res.Flowtype = flowtype
			}
		}
		if res.Flowtype != 0 {
			log.Warn("activate response mid:%d,ip:%s,pip:%s,result:%+v,response:%+v", mid, ip, pip, result, res)
			return res
		}
		// 做好边界问题，接口文档不能保证
		log.Error("日志告警 联通自动激活 product 识别失败,mid:%d,ip:%s,pip:%s,result:%+v", mid, ip, pip, result)
		return res
	case "1":
		return &unicom.Activate{Usertype: 2}
	// 做好边界问题，接口文档不能保证
	default:
		log.Error("日志告警 联通自动激活 flag 识别失败,mid:%d,ip:%s,pip:%s,result:%+v", mid, ip, pip, result)
		return &unicom.Activate{Usertype: 2}
	}
}

func (s *Service) ActiveState(ctx context.Context, mid int64, usermob string, now time.Time) (*unicom.ActiveState, error) {
	orderm, err := s.orders(ctx, usermob, now)
	if err != nil {
		return nil, err
	}
	order, err := s.orderState(orderm, "", now)
	if err != nil {
		log.Error("orderState mid:%v,usermob:%v,error:%+v", mid, usermob, err)
		if ecode.EqualError(xecode.AppWelfareClubNotFree, err) {
			return nil, errors.WithStack(ecode.Error(xecode.AppWelfareClubOrderFailed, "该卡号尚未开通哔哩哔哩专属免流服务"))
		}
		if ecode.EqualError(xecode.AppWelfareClubCancelOrExpire, err) {
			return nil, errors.WithStack(ecode.Error(xecode.AppWelfareClubOrderFailed, "该卡号哔哩哔哩专属免流服务已退订且已过期"))
		}
		return nil, err
	}
	return &unicom.ActiveState{
		ProductID:   strconv.Itoa(order.Spid),
		TfType:      order.TfType,
		TfWay:       order.TfWay,
		ProductDesc: order.Desc,
		ProductTag:  order.ProductTag,
		ProductType: order.CardType,
		Usermob:     usermob,
	}, nil
}

func (s *Service) AutoActiveState(ctx context.Context, pip, ip string) (*unicom.ActiveState, error) {
	reply, err := s.dao.Activate(ctx, pip, ip)
	if err != nil {
		return nil, err
	}
	switch reply.Flag {
	case "0":
		// 90157638 哔哩哔哩22卡
		// 90157639 哔哩哔哩33卡
		// 90157799 哔哩哔哩小电视卡
		// 90987322 s10免流包
		// 联通接口暂时不支持免流包
		if reply.Product == "" {
			return nil, errors.Wrapf(ecode.Error(xecode.AppWelfareClubOrderFailed, xecode.OrderFailedMessage), "接口响应:%+v", reply)
		}
		res := &unicom.ActiveState{}
		for _, product := range s.c.Unicom.CardProduct {
			if reply.Product == product.ID {
				res.ProductID = reply.Product
				res.ProductDesc = product.Desc
				res.ProductTag = product.Tag
				res.ProductType = product.Type
				res.TfType = 1 // 免流类型：0-不免流，1-免流卡，2-免流包
				res.TfWay = "ip"
				return res, nil
			}
		}
		for _, product := range s.c.Unicom.FlowProduct {
			if reply.Product == product.ID {
				res.ProductID = reply.Product
				res.ProductDesc = product.Desc
				res.ProductTag = product.Tag
				res.ProductType = product.Type
				res.TfType = 2 // 免流类型：0-不免流，1-免流卡，2-免流包
				res.TfWay = "cdn"
				if product.Way == "ip" {
					res.TfWay = "ip"
				}
				return res, nil
			}
		}
		// 做好边界问题，接口文档不能保证
		return nil, errors.Wrapf(ecode.Error(xecode.AppWelfareClubOrderFailed, xecode.OrderFailedMessage), "接口响应:%+v", reply)
	case "1":
		return nil, errors.Wrapf(ecode.Error(xecode.AppWelfareClubOrderFailed, xecode.OrderFailedMessage), "接口响应:%+v", reply)
	// 做好边界问题，接口文档不能保证
	default:
		return nil, errors.Wrapf(ecode.Error(xecode.AppWelfareClubOrderFailed, xecode.OrderFailedMessage), "接口响应:%+v", reply)
	}
}

func (s *Service) UserActiveLog(param *unicom.UserActiveParam, v *unicom.ActiveState, err error, from string) {
	const (
		_businessID = 93
		_action     = "active"
	)
	suggest := func() string {
		if err != nil {
			return fmt.Sprintf("联通非免流用户。失败原因：%s", err)
		}
		if v != nil && v.TfType == 0 {
			return "联通非免流用户"
		}
		return "联通免流用户"
	}()
	uInfo := &actionlog.UserInfo{
		Business: _businessID,
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

func (s *Service) AutoActiveStateByUsermob(ctx context.Context, param *unicom.UserActiveParam) (*unicom.ActiveState, error) {
	var (
		autoActive, usermobActive *unicom.ActiveState
		autoError, usermobError   error
	)
	usermob := param.Usermob
	fakeID := param.FakeID
	eg := errgroup.WithContext(ctx)
	eg.Go(func(ctx context.Context) error {
		if autoActive, autoError = s.AutoActiveState(ctx, param.SinglePip, param.IP); autoError != nil {
			log.Error("[service.AutoActiveStateByUsermob] AutoActiveState, param:%+v, error:%+v", param, autoError)
		}
		return nil
	})
	eg.Go(func(ctx context.Context) error {
		if !param.NeedFlowAuto {
			return nil
		}
		if usermob == "" {
			var info *unicom.UserMobInfo
			if info, usermobError = s.GetUserMobInfo(ctx, param); usermobError != nil {
				log.Error("[service.AutoActiveStateByUsermob] GetUserMobInfo, param:%+v, error:%+v", param, usermobError)
				return nil
			}
			// 传参fakeID = 原fakeID + 时间month
			fakeID = info.FakeID + info.Month
			// 处理空usermob的情况
			if info.Usermob == "" {
				usermobActive = &unicom.ActiveState{}
				return nil
			}
			usermob = info.Usermob
		}
		var err error
		if usermobActive, err = s.ActiveState(ctx, param.Mid, usermob, time.Now()); err != nil {
			log.Error("[service.AutoActiveStateByUsermob] ActiveState, param:%+v, error:%+v", param, err)
		}
		if usermobActive == nil {
			usermobActive = &unicom.ActiveState{}
		}
		return nil
	})
	if err := eg.Wait(); err != nil {
		log.Error("[service.AutoActiveStateByUsermob] param:%+v, error:%+v", param, err)
	}
	if !param.NeedFlowAuto {
		return autoActive, autoError
	}
	if autoError != nil && usermobError != nil {
		return nil, autoError
	}
	if autoError == nil {
		autoActive.Usermob = usermob
		autoActive.FakeID = fakeID
		return autoActive, nil
	}
	usermobActive.Usermob = usermob
	usermobActive.FakeID = fakeID
	return usermobActive, nil
}

func (s *Service) GetUserMobInfo(ctx context.Context, param *unicom.UserActiveParam) (*unicom.UserMobInfo, error) {
	var isNewFakeID bool // 是否为最新fake_id
	parseOGFakeIDAndMonth := func() (string, string, error) {
		if param.FakeID == "" {
			ogFakeID, month, err := s.getFakeIDInfo(ctx, param.SinglePip, param.IP)
			if err != nil {
				log.Error("[service.GetUserMobInfo] getFakeIDInfo error, param:%+v, err:%v", param, err)
				return "", "", err
			}
			isNewFakeID = true
			return ogFakeID, month, nil
		}
		if len(param.FakeID) < _minFakeIdLen {
			log.Error("[service.GetUserMobInfo] fakeID长度不符合预期, param:%+v", param)
			return "", "", errors.New("fakeID长度不符合预期")
		}
		month := param.FakeID[len(param.FakeID)-_monthLen:]
		ogFakeID := param.FakeID[:len(param.FakeID)-_monthLen]
		return ogFakeID, month, nil
	}
	ogFakeID, month, err := parseOGFakeIDAndMonth()
	if err != nil {
		log.Error("[service.GetUserMobInfo] parseOGFakeIDAndPeriod error, param:%+v, error:%+v", param, err)
		return nil, err
	}
	// 检查fakeID是否过期
	monthTime, err := time.Parse(_fakeIDMonthLayout, month)
	if err != nil {
		log.Error("[service.GetUserMobInfo] month 转换失败 error:%v", err)
		return nil, err
	}
	if !isNewFakeID && s.isExpired(monthTime, time.Now()) {
		ogFakeID, month, err = s.getFakeIDInfo(ctx, param.SinglePip, param.IP)
		if err != nil {
			log.Error("[service.GetUserMobInfo] getFakeIDInfo error, param:%+v, error:%+v", param, err)
			return nil, err
		}
		monthTime, _ = time.Parse(_fakeIDMonthLayout, month)
		isNewFakeID = true
	}
	info, err := s.getUserMobInfo(ctx, ogFakeID, int64(monthTime.Month()))
	if err != nil {
		log.Error("[service.GetUserMobInfo] getUserMobInfo error, param:%+v, error:%+v", param, err)
		return nil, err
	}
	// 替换最新fake_id和时间month
	info.FakeID = ogFakeID
	info.Month = month
	return info, nil
}

func (s *Service) getUserMobInfo(ctx context.Context, fakeID string, period int64) (*unicom.UserMobInfo, error) {
	info, err := s.dao.GetUsermobInfoCache(ctx, fakeID, period)
	if err == nil {
		s.pHit.Incr("unicom_usermob_info")
		return info, nil
	}
	if info, err = s.dao.SelectUserMobInfo(ctx, fakeID, period); err != nil {
		log.Error("[service.getUserMobInfo] SelectUserMobInfo error, error:%v", err)
		return nil, err
	}
	miss := &unicom.UserMobInfo{}
	*miss = *info
	s.pMiss.Incr("unicom_usermob_info")
	s.cache.Do(ctx, func(ctx context.Context) {
		if err = s.dao.AddUsermobInfoCache(ctx, fakeID, period, miss); err != nil {
			log.Error("[service.getUserMobInfo] AddUsermobInfoCache error, error:%v", err)
		}
	})
	return info, nil
}

func (s *Service) getFakeIDInfo(ctx context.Context, pip, ip string) (fakeID, month string, err error) {
	resp, err := s.dao.GetFakeIDInfo(ctx, pip, ip)
	if err != nil {
		return "", "", err
	}
	orderTime, err := strconv.ParseInt(resp.OrderTime, 10, 64)
	if err != nil {
		return "", "", err
	}
	fakeID = resp.PCode
	month = time.Unix(orderTime, 0).Format(_fakeIDMonthLayout)
	return fakeID, month, nil
}

func (s *Service) isExpired(infoMonth, nowTime time.Time) bool {
	return infoMonth.Year() != nowTime.Year() || infoMonth.Month() != nowTime.Month()
}

func (s *Service) UnicomFlowTryout(ctx context.Context, fakeID, pip, ip string) (string, error) {
	ogFakeID, month, err := s.getOGFakeIDAndMonth(ctx, fakeID, pip, ip)
	if err != nil {
		return "", err
	}
	fakeID = ogFakeID + month
	if err := s.dao.UnicomFlowTryout(ctx, ogFakeID); err != nil {
		log.Error("[s.UnicomFlowTryout] fakeID:%s, error:%+v", fakeID, err)
		return fakeID, err
	}
	return fakeID, nil
}

func (s *Service) getOGFakeIDAndMonth(ctx context.Context, fakeID, pip, ip string) (ogFakeID, month string, err error) {
	if fakeID == "" {
		ogFakeID, month, err = s.getFakeIDInfo(ctx, pip, ip)
		if err != nil {
			log.Error("[s.getOGFakeIDAndMonth] pip:%s, ip:%s, error:%+v", pip, ip, err)
			return
		}
		return
	}
	ogFakeID = fakeID[:len(fakeID)-_monthLen]
	month = fakeID[len(fakeID)-_monthLen:]
	// 检查fakeID是否过期
	monthTime, err := time.Parse(_fakeIDMonthLayout, month)
	if err != nil {
		log.Error("[s.getOGFakeIDAndMonth] month 转换失败 error:%v", err)
		err = ecode.RequestErr
		return
	}
	if s.isExpired(monthTime, time.Now()) {
		ogFakeID, month, err = s.getFakeIDInfo(ctx, pip, ip)
		if err != nil {
			log.Error("[s.getOGFakeIDAndMonth] getFakeIDInfo pip:%s, ip:%s, error:%+v", pip, ip, err)
			return
		}
		return
	}
	return
}
