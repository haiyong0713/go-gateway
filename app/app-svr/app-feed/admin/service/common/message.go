package common

import (
	"fmt"

	"go-gateway/app/app-svr/app-feed/admin/model/common"
)

// Notify .
func (s *Service) Notify(mids []int64, bus int64, title, msg string) (err error) {
	if bus == common.NotifyBusnessTianma {
		return s.messageDao.NotifyTianma(mids, title, msg)
	}
	return fmt.Errorf("参数错误")
}
