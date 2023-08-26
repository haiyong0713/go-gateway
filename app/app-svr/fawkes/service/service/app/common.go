package app

import (
	"context"
	log "go-gateway/app/app-svr/fawkes/service/tools/logger"
)

// System get system version.
func (s *Service) System(c context.Context, platform string) (res []string) {
	switch platform {
	case "iOS", "IOS", "ios":
		res = s.c.System.IOS
	case "android", "Android", "ANDROID":
		res = s.c.System.Android
	}
	return
}

func (s *Service) ServicePing(c context.Context, requestUrl string) (err error) {
	if err = s.fkDao.HookRequest(c, requestUrl, "GET", nil); err != nil {
		log.Error("ServicePing error: %v", err)
	}
	return
}
