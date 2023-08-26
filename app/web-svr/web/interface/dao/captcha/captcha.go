package captcha

import (
	"context"
	"fmt"
	"math/rand"
	"strconv"
	"time"

	"github.com/pkg/errors"
	"go-common/library/cache/redis"
	"go-common/library/log"

	"go-gateway/app/web-svr/web/interface/conf"
)

type Dao struct {
	cfg *conf.Captcha
	rds *redis.Pool
}

func NewDao(cfg *conf.Captcha, rds *redis.Pool) (*Dao, error) {
	if cfg.Biz == "" {
		return nil, errors.New("biz is empty")
	}
	if cfg.Digit == 0 {
		return nil, errors.New("digit is zero")
	}
	fixConfig(cfg)
	return &Dao{cfg: cfg, rds: rds}, nil
}

func fixConfig(cfg *conf.Captcha) {
	if cfg.ValidTime == 0 {
		cfg.ValidTime = 300
	}
}

func (d *Dao) Cfg() *conf.Captcha {
	return d.cfg
}

func (d *Dao) IncrCacheCaptchaIp(c context.Context, ip string) (int64, error) {
	conn := d.rds.Get(c)
	defer conn.Close()
	key := d.CaptchaIpKey(ip)
	counter, err := redis.Int64(conn.Do("INCR", key))
	if err != nil {
		log.Errorc(c, "Fail to incr CacheCaptchaIp, key=%s error=%+v", key, err)
		return 0, err
	}
	if counter == 1 {
		if _, err := conn.Do("EXPIRE", key, 86400); err != nil {
			log.Errorc(c, "Fail to expire CaptchaIpKey, key=%s error=%+v", key, err)
		}
	}
	return counter, nil
}

func (d *Dao) DecrCacheCaptchaIp(c context.Context, ip string) (int64, error) {
	conn := d.rds.Get(c)
	defer conn.Close()
	key := d.CaptchaIpKey(ip)
	counter, err := redis.Int64(conn.Do("DECR", key))
	if err != nil {
		log.Errorc(c, "Fail to decr CacheCaptchaIp, key=%s error=%+v", key, err)
		return 0, err
	}
	return counter, nil
}

func (d *Dao) CleanCacheAfterSendFailed(c context.Context, mobile int64, ip string) error {
	conn := d.rds.Get(c)
	defer conn.Close()
	ipKey := d.CaptchaIpKey(ip)
	if err := conn.Send("DECR", ipKey); err != nil {
		log.Errorc(c, "Fail to decr CacheCaptchaIp, key=%s error=%+v", ipKey, err)
		return err
	}
	captKey := d.CaptchaKey(mobile)
	if err := conn.Send("DEL", captKey); err != nil {
		log.Errorc(c, "Fail to del CacheCaptcha, key=%s error=%+v", captKey, err)
		return err
	}
	if err := conn.Flush(); err != nil {
		log.Errorc(c, "Fail to flush CleanCacheAfterSendFailed, ipKey=%s captKey=%s error=%+v", ipKey, captKey, err)
		return err
	}
	if _, err := conn.Receive(); err != nil {
		log.Errorc(c, "Fail to decr CacheCaptchaIp, key=%s error=%+v", ipKey, err)
		return err
	}
	if _, err := conn.Receive(); err != nil {
		log.Errorc(c, "Fail to del CacheCaptcha, key=%s error=%+v", captKey, err)
		return err
	}
	return nil
}

func (d *Dao) AddCacheCaptcha(c context.Context, mobile int64, captcha string) error {
	conn := d.rds.Get(c)
	defer conn.Close()
	key := d.CaptchaKey(mobile)
	if _, err := conn.Do("SET", key, captcha, "EX", d.cfg.ValidTime); err != nil {
		log.Errorc(c, "Fail to cache captcha, key=%s error=%+v", key, err)
		return err
	}
	return nil
}

func (d *Dao) CacheCaptcha(c context.Context, mobile int64) (string, error) {
	conn := d.rds.Get(c)
	defer conn.Close()
	key := d.CaptchaKey(mobile)
	captcha, err := redis.String(conn.Do("GET", key))
	if err != nil {
		if err == redis.ErrNil {
			return "", nil
		}
		log.Errorc(c, "Fail to get captcha cache, key=%s error=%+v", key, err)
		return "", err
	}
	return captcha, nil
}

func (d *Dao) CleanCacheAfterVerified(c context.Context, mobile int64) error {
	conn := d.rds.Get(c)
	defer conn.Close()
	key := d.CaptchaKey(mobile)
	if err := conn.Send("DEL", key); err != nil {
		log.Errorc(c, "Fail to del captcha cache, key=%s error=%+v", key, err)
		return err
	}
	failedKey := d.captchaFailedKey(mobile)
	if err := conn.Send("DEL", failedKey); err != nil {
		log.Errorc(c, "Fail to del captcha_failed cache, key=%s error=%+v", failedKey, err)
		return err
	}
	if err := conn.Flush(); err != nil {
		log.Errorc(c, "Fail to flush CleanCacheAfterVerified, key=%s failedKey=%s error=%+v", key, failedKey, err)
		return err
	}
	if _, err := conn.Receive(); err != nil {
		log.Errorc(c, "Fail to del captcha cache, key=%s error=%+v", key, err)
		return err
	}
	if _, err := conn.Receive(); err != nil {
		log.Errorc(c, "Fail to del captcha_failed cache, key=%s error=%+v", failedKey, err)
		return err
	}
	return nil
}

func (d *Dao) CaptchaTTL(c context.Context, mobile int64) (int64, error) {
	conn := d.rds.Get(c)
	defer conn.Close()
	key := d.CaptchaKey(mobile)
	ttl, err := redis.Int64(conn.Do("TTL", key))
	if err != nil {
		log.Errorc(c, "Fail to get captcha ttl, key=%s error=%+v", key, err)
		return 0, err
	}
	if ttl < 0 {
		return 0, nil
	}
	return ttl, nil
}

func (d *Dao) IncrCacheCaptchaFailed(c context.Context, mobile int64) (int64, error) {
	conn := d.rds.Get(c)
	defer conn.Close()
	key := d.captchaFailedKey(mobile)
	counter, err := redis.Int64(conn.Do("INCR", key))
	if err != nil {
		log.Errorc(c, "Fail to incr CacheCaptchaFailed, key=%s error=%+v", key, err)
		return 0, err
	}
	if counter == 1 {
		if _, err := conn.Do("EXPIRE", key, 86400); err != nil {
			log.Errorc(c, "Fail to expire captchaFailedKey, key=%s error=%+v", key, err)
		}
	}
	return counter, nil
}

func (d *Dao) GenerateCaptcha() string {
	rd := rand.New(rand.NewSource(time.Now().UnixNano()))
	var captcha string
	for i := 0; i < d.cfg.Digit; i++ {
		captcha += strconv.Itoa(rd.Intn(10))
	}
	return captcha
}

func (d *Dao) CaptchaIpKey(ip string) string {
	return fmt.Sprintf("capt_ip_%s_%s_%s", d.cfg.Biz, ip, time.Now().Format("20060102"))
}

func (d *Dao) CaptchaKey(mobile int64) string {
	return fmt.Sprintf("capt_%s_%d", d.cfg.Biz, mobile)
}

func (d *Dao) captchaFailedKey(mobile int64) string {
	return fmt.Sprintf("capt_failed_%s_%d", d.cfg.Biz, mobile)
}
