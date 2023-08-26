package view

import (
	"context"
	"git.bilibili.co/go-tool/libbdevice/pkg/pd"

	archive "go-gateway/app/app-svr/archive/service/api"
)

func (s *Service) ReportPremiereWatch(c context.Context, isMelloi string, a *archive.Arc, buvid string) {
	if isMelloi != "" || a == nil || a.Premiere == nil || a.Premiere.State == archive.PremiereState_premiere_none ||
		a.Premiere.State == archive.PremiereState_premiere_after {
		return
	}
	if pd.WithContext(c).Where(func(pd *pd.PDContext) {
		pd.IsPlatAndroid().And().Build(">=", int64(6670000))
	}).OrWhere(func(pd *pd.PDContext) {
		pd.IsPlatIPhone().And().Build(">=", int64(66700000))
	}).FinishOr(false) {
		//上报
		_ = s.poDao.ReportWatch(c, a.Aid, buvid)
	}
}
