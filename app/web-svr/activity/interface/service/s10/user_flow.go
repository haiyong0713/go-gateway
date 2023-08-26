package s10

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"go-gateway/app/web-svr/activity/interface/conf"
	"strings"
	"time"

	"go-gateway/app/web-svr/activity/ecode"
	"go-gateway/app/web-svr/activity/interface/model/s10"

	xecode "go-common/library/ecode"
	"go-common/library/log"
)

func (s *Service) CheckUserProfile(ctx context.Context, source int32, tel, timestamp, sign string) (int64, error) {
	var secretkey string
	if !s.flowSwitch {
		return 2, nil
	}
	switch source {
	case 0:
		secretkey = s.unicomSecretkey
	case 1:
		secretkey = s.mobileSecretkey
	default:
		return 0, xecode.RequestErr
	}
	if !s.checkMd5Key(tel, timestamp, sign, secretkey) {
		return 0, xecode.SignCheckErr
	}
	var (
		mid int64
		err error
	)
	if s.flowRecvLimit {
		mid, err = s.dao.MidByTel(ctx, tel)
		if err != nil {
			code := xecode.Cause(err).Code()
			if code == 75312 || code == -626 { //账号不存在错误码
				log.Errorc(ctx, "userflow UserNotExist error:%v", err)
				return 2, nil
			}
			return 0, ecode.ActivityDataPackageFail
		}
		points, err := s.dao.TotalPoints(ctx, mid, s.s10Act)
		if err != nil {
			return 0, ecode.ActivityDataPackageFail
		}
		if points < s.points {
			log.Errorc(ctx, "userflow points:%d,needPoints", points, s.points)
			return 2, nil
		}
		userRecvState, err := s.userFlow(ctx, mid)
		if err != nil {
			return 0, ecode.ActivityDataPackageFail
		}
		if userRecvState >= 2 {
			log.Errorc(ctx, "userflow userRecvState:%d", userRecvState)
			return 2, nil
		}
	}
	err = s.dao.FreeFlowPubDataBus(ctx, 0, source, tel)
	if err != nil {
		return 0, ecode.ActivityDataPackageFail
	}
	s.dao.DelUserFlow(ctx, mid)
	return 1, nil
}

func (s *Service) checkMd5Key(tel, timestamp, sign, secretkey string) bool {
	str := fmt.Sprintf("tel=%s,timestamp=%s,secretKey=%s", tel, timestamp, secretkey)
	hasher := md5.New()
	hasher.Write([]byte(str))
	result := strings.ToUpper(hex.EncodeToString(hasher.Sum(nil)))
	if result != sign {
		log.Warn("s10 checkMd5Key: str:%s, hash:%s, sign:%s", str, result, sign)
		return false
	}
	return true
}

func (s *Service) AckSendFreeFlow(ctx context.Context, source int32, tel, timestamp, sign string) error {
	var secretkey string
	if !s.flowSwitch {
		return xecode.RequestErr
	}
	switch source {
	case 0:
		secretkey = s.unicomSecretkey
	case 1:
		secretkey = s.mobileSecretkey
	default:
		return xecode.RequestErr
	}
	if !s.checkMd5Key(tel, timestamp, sign, secretkey) {
		return xecode.SignCheckErr
	}
	err := s.dao.FreeFlowPubDataBus(ctx, 1, source, tel)
	if err != nil {
		return ecode.ActivityDataPackageFail
	}
	return nil
}

func (s *Service) UserFlow(ctx context.Context, mid int64) (*s10.FreeFlow, error) {
	res := &s10.FreeFlow{Points: s.points, CurrentTime: time.Now().Unix(), Swith: s.flowSwitch}
	res.Mobile, res.MobileStartTime, res.MobileEndTime = s.generalFlowSwitch(ctx, s.conf.S10General.Mobile)
	res.Unicom, res.UnicomStartTime, res.UnicomEndTime = s.generalFlowSwitch(ctx, s.conf.S10General.Unicom)
	if mid <= 0 || !s.flowSwitch {
		return res, nil
	}
	res.IsLogIn = true
	var err error
	res.Receive, err = s.userFlow(ctx, mid)
	return res, err
}

func (s *Service) generalFlowSwitch(ctx context.Context, conf *conf.S10FlowControl) (int32, int64, int64) {
	if !conf.Switch {
		// 开关统一关闭，再无库存状态
		return 4, 0, 0
	}
	// 开关打开时，看下是否已经到了开始时间，未到返回敬请期待
	if time.Now().Before(conf.StartAt) {
		return 0, 0, 0
	}
	// 已经开始了，并且开关没关闭，检查是否有符合的时间段
	now := time.Now()
	for _, one := range conf.Stock {
		if now.After(one.StartAt) && now.Before(one.EndAt) {
			startTime := time.Date(now.Year(), now.Month(), now.Day(), one.StartHour, 0, 0, 0, time.Local)
			endTime := time.Date(now.Year(), now.Month(), now.Day(), one.Endhour, 59, 59, 0, time.Local)
			if now.Hour() < one.StartHour {
				// 当前小时小于开始时间，定时开抢
				return 1, startTime.Unix(), endTime.Unix()
			}
			if now.Hour() <= one.Endhour {
				// 在时间范围内，可以正常领取
				return 2, startTime.Unix(), endTime.Unix()
			}
			// 过时间段了，当天已抢完
			return 3, 0, 0
		}
	}
	return 3, 0, 0
}

func (s *Service) userFlow(ctx context.Context, mid int64) (int32, error) {
	userRecvState, err := s.dao.UserFlowCache(ctx, mid)
	if err != nil || userRecvState != 0 {
		return userRecvState, err
	}
	userRecvState, source, err := s.dao.UserFlow(ctx, mid)
	if err != nil {
		return 0, ecode.ActivityDataPackageFail
	}
	switch userRecvState {
	case 0:
		userRecvState = 1
	default: //db:source: 0-联通；1-移动；缓存值userRecvState：1-哨兵；2-联通；3-移动
		userRecvState = source + 2
	}
	s.dao.AddUserFlowCache(ctx, mid, userRecvState)
	return userRecvState, nil
}
