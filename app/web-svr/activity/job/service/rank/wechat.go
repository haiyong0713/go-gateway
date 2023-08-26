package rank

import (
	"context"
	"go-common/library/log"
)

// sendWechat 发送微信
func (s *Service) sendWechat(c context.Context, title, message, user string) (err error) {
	err = s.rankDao.SendWeChat(c, s.c.Rank.PublicKey, title, message, user)
	if err != nil {
		log.Errorc(c, "s.dao.SendWechat error(%v)", err)
	}
	return
}
