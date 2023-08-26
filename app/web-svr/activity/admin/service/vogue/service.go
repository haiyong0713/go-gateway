package vogue

import (
	"context"
	"os"

	xhttp "go-common/library/net/http/blademaster"
	"go-gateway/app/web-svr/activity/admin/conf"
	lotteryDao "go-gateway/app/web-svr/activity/admin/dao/lottery"
	vogueDao "go-gateway/app/web-svr/activity/admin/dao/vogue"
	voguemdl "go-gateway/app/web-svr/activity/admin/model/vogue"
)

// Service struct
type Service struct {
	c                 *conf.Config
	dao               *vogueDao.Dao
	lotDao            *lotteryDao.Dao
	httpClient        *xhttp.Client
	exportData        *voguemdl.CreditExportData
	weChatBlockStatus bool
	exportState       int64
}

// Close service
func (s *Service) Close() {
	if s.dao != nil {
		s.dao.Close()
	}
}

// New service
func New(c *conf.Config) (s *Service) {
	s = &Service{
		c:                 c,
		dao:               vogueDao.New(c),
		lotDao:            lotteryDao.New(c),
		httpClient:        xhttp.NewClient(c.HTTPClient),
		exportData:        &voguemdl.CreditExportData{},
		weChatBlockStatus: false,
	}
	// 活动有效期间，进行监测
	if c.VogueActivity.Active > 0 {
		if os.Getenv("DEPLOY_ENV") == "prod" {
			// 微信域名监测
			go s.WeChatHostMonitor(context.Background())
			// 检测是否在双倍时间内
			go s.DoubleScoreMonitor(context.Background())
		}
	}

	return
}
