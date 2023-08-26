package service

import (
	"context"
	"fmt"

	"go-common/library/ecode"
	"go-common/library/log"
	"go-common/library/net/metadata"

	xecode "go-gateway/app/web-svr/web/ecode"
	"go-gateway/app/web-svr/web/interface/dao/captcha"
	"go-gateway/app/web-svr/web/interface/model"
)

func (s *Service) SendCaptcha(c context.Context, req *model.SendCaptchaReq, dao *captcha.Dao) error {
	if !verifyMobile(req.Mobile) {
		return xecode.AppealWrongMobile
	}
	// 限制发送间隔时长
	ttl, err := dao.CaptchaTTL(c, req.Mobile)
	if err != nil {
		return ecode.ServerErr
	}
	interval := dao.Cfg().ValidTime - ttl
	if interval < dao.Cfg().IntervalTime {
		log.Warnc(c, "CaptchaFrequently, cache_key=%s interval=%d", dao.CaptchaKey(req.Mobile), interval)
		return xecode.CaptchaFrequently
	}
	// 限制IP调用频率
	ip := metadata.String(c, metadata.RemoteIP)
	ipCnt, err := dao.IncrCacheCaptchaIp(c, ip)
	if err != nil {
		return ecode.ServerErr
	}
	if ipCnt > dao.Cfg().IpDailyLimit {
		log.Warnc(c, "CaptchaIpLimit, cache_key=%s cnt=%d", dao.CaptchaIpKey(ip), ipCnt)
		return xecode.CaptchaIpLimit
	}
	return func() (err error) {
		defer func() {
			if err == nil {
				return
			}
			_ = s.cache.Do(c, func(ctx context.Context) {
				_ = dao.CleanCacheAfterSendFailed(ctx, req.Mobile, ip)
			})
		}()
		// 生成验证码
		capt := dao.GenerateCaptcha()
		// 存至缓存
		if err = dao.AddCacheCaptcha(c, req.Mobile, capt); err != nil {
			return ecode.ServerErr
		}
		// 发送验证码
		tparam := fmt.Sprintf(`{"%s":"%s"}`, dao.Cfg().TField, capt)
		return s.dao.SendSms(c, req.Mobile, dao.Cfg().Tcode, tparam)
	}()
}

func (s *Service) VerifyCaptcha(c context.Context, mobile int64, captcha string, dao *captcha.Dao) (err error) {
	realCapt, err := dao.CacheCaptcha(c, mobile)
	if err != nil {
		return ecode.ServerErr
	}
	defer func() {
		if err == nil || err == xecode.CaptchaInvalid {
			_ = s.cache.Do(c, func(ctx context.Context) {
				_ = dao.CleanCacheAfterVerified(c, mobile)
			})
		}
	}()
	if realCapt == "" {
		return xecode.CaptchaWrong
	}
	if realCapt == captcha {
		return nil
	}
	if failed, err := dao.IncrCacheCaptchaFailed(c, mobile); err == nil && failed > dao.Cfg().VerifyLimit {
		return xecode.CaptchaInvalid
	}
	return xecode.CaptchaWrong
}

func verifyMobile(mobile int64) bool {
	return mobile >= 10000000000 && mobile < 20000000000
}
