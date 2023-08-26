package common

import (
	"context"
	"fmt"

	"go-gateway/app/app-svr/app-feed/ecode"

	accgrpc "git.bilibili.co/bapis/bapis-go/account/service"
)

// Archives .
func (s *Service) UpInfo(c context.Context, id int64) (accCard *accgrpc.Card, err error) {
	if accCard, err = s.accDao.Card3(c, id); err != nil {
		if err.Error() == ecode.MemberNotExist.Error() {
			return nil, fmt.Errorf("无效up主ID(%d)", id)
		}
	}
	return
}
