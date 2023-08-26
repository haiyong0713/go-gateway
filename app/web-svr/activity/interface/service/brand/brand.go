package brand

import (
	"context"
	"fmt"
	"strconv"
	"time"

	xecode "go-common/library/ecode"
	"go-common/library/log"
	"go-common/library/sync/errgroup.v2"
	"go-gateway/app/web-svr/activity/ecode"
	brandMdl "go-gateway/app/web-svr/activity/interface/model/brand"

	couponapi "git.bilibili.co/bapis/bapis-go/account/service/coupon"
	passportinfoapi "git.bilibili.co/bapis/bapis-go/passport/service/user"
	silverbulletapi "git.bilibili.co/bapis/bapis-go/silverbullet/service/silverbullet-proxy"
	vipresourceapi "git.bilibili.co/bapis/bapis-go/vip/resource/service"
	"github.com/pkg/errors"
)

const (
	strategyName     = "activity_brand"
	experienceType   = 1
	couponType       = 2
	couponTypeTimes1 = 3
	couponTypeTimes2 = 4
	defaultBuvid     = "activity_brand_oJINVWsgl3bSf2CUI"
	period           = 3600
)

// AddCoupon add coupon
func (s *Service) AddCoupon(c context.Context, mid int64, frontEndParams *brandMdl.FrontEndParams) (data *brandMdl.CouponReply, err error) {
	var (
		infoReply *passportinfoapi.CheckFreshUserReply
		isAdd     bool
		couponErr error
	)
	infoReply, isAdd, err = s.preCheck(c, mid, frontEndParams)
	defer func() {
		// 如果发券失败，且返回错误非超时，则减去用户领取次数
		if err != nil && isAdd && !xecode.EqualError(xecode.Deadline, err) && !xecode.EqualError(ecode.ActivityBrandAwardOnceErr, err) {
			log.Info("send coupon mid(%d) error(%v),set minus coupon times", mid, err)
			if err1 := s.dao.CacheSetMinusCouponTimes(c, mid); err1 != nil {
				log.Error("s.dao.CacheSetMinusCouponTimes(%d) error(%v)", mid, err1)
			}
		}
	}()
	if err != nil {
		return &brandMdl.CouponReply{CouponType: 0}, err
	}
	// 如果用户是新大会员
	if infoReply.IsNew {
		// 发放体验券
		couponErr = s.experienceVip(c, mid)
		if couponErr == nil {
			return &brandMdl.CouponReply{CouponType: experienceType}, nil
		}
	}
	// 如果发放体验券失败 且失败错误码不等于发完，则错误
	if couponErr != nil && !xecode.EqualError(xecode.Int(brandMdl.VipBatchNotEnoughErr), couponErr) {
		if xecode.EqualError(ecode.ActivityQPSLimitErr, couponErr) {
			return &brandMdl.CouponReply{CouponType: 0}, couponErr
		}
		return &brandMdl.CouponReply{CouponType: 0}, ecode.ActivityBrandCouponErr
	}
	// 发放折扣券
	err = s.coupon(c, mid)
	if err != nil {
		return &brandMdl.CouponReply{CouponType: 0}, ecode.ActivityBrandCouponErr
	}
	return &brandMdl.CouponReply{CouponType: couponType}, nil
}

// preCheck 检查
func (s *Service) preCheck(c context.Context, mid int64, frontEndParams *brandMdl.FrontEndParams) (infoReply *passportinfoapi.CheckFreshUserReply, isAdd bool, err error) {
	eg := errgroup.WithContext(c)
	isAdd = false
	var (
		count     int64
		riskReply *silverbulletapi.RiskInfoReply
	)
	// redis 中增加次数
	eg.Go(func(ctx context.Context) (err error) {
		if count, err = s.dao.CacheAddCouponTimes(c, mid); err != nil {
			log.Error("s.dao.CacheAddCouponTimes(%d) error(%v)", mid, err)
			err = errors.Wrapf(err, "s.dao.CacheAddCouponTimes %d", mid)
			return
		}
		isAdd = true
		return
	})
	// 用户是否新用户
	eg.Go(func(ctx context.Context) (err error) {
		if infoReply, err = s.passportClient.CheckFreshUser(c, &passportinfoapi.CheckFreshUserReq{Mid: mid, Buvid: defaultBuvid, Period: period}); err == nil || infoReply == nil {
			log.Error("s.passportClient.CheckFreshUser(%d) error(%v)", mid, err)
			err = errors.Wrapf(err, "s.passportClient.CheckFreshUser %d", mid)
		}
		return
	})
	// 风控
	eg.Go(func(ctx context.Context) (err error) {
		if riskReply, err = s.silverbulletClient.RiskInfo(c, &silverbulletapi.RiskInfoReq{
			Mid: mid, Ip: frontEndParams.IP,
			DeviceId:     frontEndParams.DeviceID,
			Ua:           frontEndParams.Ua,
			Api:          frontEndParams.API,
			Referer:      frontEndParams.Referer,
			StrategyName: []string{strategyName},
		}); err != nil || riskReply == nil || riskReply.Infos == nil {
			log.Error("s.silverbulletClient.RiskInfo(%d) error(%v)", mid, err)
			err = errors.Wrapf(err, "s.silverbulletClient.RiskInfo %d", mid)
		}
		return
	})

	if err = eg.Wait(); err != nil {
		log.Error("eg.Wait error(%v)", err)
		return nil, isAdd, ecode.ActivityBrandMidErr
	}
	// 统计次数不等于1
	if count != 1 {
		return nil, isAdd, ecode.ActivityBrandAwardOnceErr
	}
	// 风险命中判断
	riskMap, ok := riskReply.Infos[strategyName]
	if !ok || riskMap == nil {
		return nil, isAdd, ecode.ActivityBrandRiskErr
	}
	if riskMap != nil && riskMap.Level > 0 {
		return nil, isAdd, ecode.ActivityBrandRiskErr
	}
	return
}

// getOrderNo 获取order_no
func (s *Service) getOrderNo(c context.Context, mid int64, couponType int) string {
	return fmt.Sprintf("%d_%d", mid, couponType)
}

// experienceVip 体验券
func (s *Service) experienceVip(c context.Context, mid int64) error {
	if s.QPSLimit(c, strconv.Itoa(experienceType), s.QPSLimitResourceCoupon) != nil {
		log.Error("qps limit limitType (%d)", experienceType)
		return ecode.ActivityQPSLimitErr
	}
	timestamp := time.Now().Unix()
	_, err := s.vipResourceClient.ResourceUse(c, &vipresourceapi.ResourceUseReq{
		Mid:        mid,
		BatchToken: s.ResourceBatchToken,
		OrderNo:    s.getOrderNo(c, mid, experienceType),
		Remark:     s.CouponExperienceRemark,
		Appkey:     s.ResourceAppkey,
		Ts:         timestamp,
	})
	if err != nil {
		log.Error("s.vipResourceClient.ResourceUse(%d) error(%v)", mid, err)
		return err
	}
	return nil
}

// coupon 折扣券
func (s *Service) coupon(c context.Context, mid int64) error {
	var err, err1, err2 error
	eg := errgroup.WithContext(c)
	eg.Go(func(ctx context.Context) error {
		_, err1 = s.couponClient.AllowanceReceive(c, &couponapi.AllowanceReceiveReq{
			Mid:        mid,
			BatchToken: s.CouponBatchToken,
			OrderNo:    s.getOrderNo(c, mid, couponTypeTimes1),
			Appkey:     s.ResourceAppkey,
		})
		if err1 != nil {
			log.Error("s.couponClient.AllowanceReceive(%d) times1 error(%v)", mid, err1)
			err = errors.Wrapf(err1, "s.couponClient.AllowanceReceive times1 %d", mid)
		}
		return err
	})
	eg.Go(func(ctx context.Context) error {
		_, err2 = s.couponClient.AllowanceReceive(c, &couponapi.AllowanceReceiveReq{
			Mid:        mid,
			BatchToken: s.CouponBatchToken2,
			OrderNo:    s.getOrderNo(c, mid, couponTypeTimes2),
			Appkey:     s.ResourceAppkey,
		})
		if err2 != nil {
			log.Error("s.couponClient.AllowanceReceive(%d) times2 error(%v)", mid, err2)
			err = errors.Wrapf(err2, "s.couponClient.AllowanceReceive times2 %d", mid)
		}
		return err
	})
	if err = eg.Wait(); err != nil {
		log.Error("eg.Wait error(%v)", err)
		return err
	}
	return nil
}

// QPSLimit ...
func (s *Service) QPSLimit(c context.Context, limitType string, maxLimit int64) error {
	limit, err := s.dao.CacheQPSLimit(c, limitType)
	if err != nil {
		log.Error("s.dao.CacheQPSLimit(%s) error(%v)", limitType, err)
		return err
	}
	if limit > maxLimit {
		return ecode.ActivityQPSLimitErr
	}
	return nil
}
