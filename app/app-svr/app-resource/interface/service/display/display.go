package display

import (
	"context"
	//nolint:gosec
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"time"

	"go-common/library/ecode"
	"go-common/library/log"
	"go-gateway/app/app-svr/app-resource/interface/conf"
	wechatdao "go-gateway/app/app-svr/app-resource/interface/dao/wechat"
)

// Service display service.
type Service struct {
	c   *conf.Config
	dao *wechatdao.Dao
}

// New new display service.
func New(c *conf.Config) (s *Service) {
	s = &Service{
		c:   c,
		dao: wechatdao.New(c),
	}
	return
}

func (s *Service) DisplayId(c context.Context, mid int64, buvid string, now time.Time) string {
	if mid == 0 {
		return fmt.Sprintf("%s-%d", buvid, now.Unix())
	}
	return fmt.Sprintf("%d-%d", mid, now.Unix())
}

func (s *Service) WechatAuth(c context.Context, nonce, timestamp, currentUrl string) (string, error) {
	wechatTicket, err := s.dao.WeChatCache(c)
	if err != nil {
		if wechatTicket, err = s.dao.WechatAuth(c, nonce, timestamp, currentUrl); err != nil {
			log.Error("%+v", err)
			return "", ecode.ServerErr
		}
		//nolint:errcheck
		s.dao.AddWeChatCache(c, wechatTicket)
	}
	res := "jsapi_ticket=" + wechatTicket + "&noncestr=" + nonce + "&timestamp=" + timestamp + "&url=" + currentUrl
	bs := sha1.Sum([]byte(res))
	return hex.EncodeToString(bs[:]), nil
}
