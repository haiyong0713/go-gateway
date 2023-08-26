package view

import (
	"context"

	account "git.bilibili.co/bapis/bapis-go/account/service"
	"go-common/library/ecode"
	"go-common/library/log"
	"go-common/library/net/metadata"
	"go-gateway/app/app-svr/app-intl/interface/model"
	arcgrpc "go-gateway/app/app-svr/archive/service/api"

	locgrpc "git.bilibili.co/bapis/bapis-go/community/service/location"
	cfcgrpc "git.bilibili.co/bapis/bapis-go/content-flow-control/service"
)

// ipLimit ip limit
func (s *Service) ipLimit(c context.Context, mid, aid int64, cdnIP string) (down int64, err error) {
	var auth *locgrpc.Auth
	ip := metadata.String(c, metadata.RemoteIP)
	if auth, err = s.locDao.Archive(c, aid, mid, ip, cdnIP); err != nil {
		log.Error("%v", err)
		err = nil // NOTE: return or ignore err???
	}
	if auth != nil {
		down = auth.Down
		switch auth.Play {
		case int64(locgrpc.Status_Forbidden):
			err = ecode.AccessDenied
			s.prom.Incr("ip_limit_access")
		}
	}
	return
}

// areaLimit area limit
func (s *Service) areaLimit(c context.Context, plat int8, rid int16) (err error) {
	ip := metadata.String(c, metadata.RemoteIP)
	if rm, ok := s.region[plat]; !ok {
		return
	} else if r, ok := rm[rid]; ok && r != nil && r.Area != "" {
		var auths map[string]*locgrpc.Auth
		if auths, err = s.locDao.AuthPIDs(c, r.Area, ip); err != nil {
			log.Error("error(%v) area(%v) ip(%v)", err, r.Area, ip)
			err = nil
			return
		}
		if auth, ok := auths[r.Area]; ok && auth.Play == int64(locgrpc.Status_Forbidden) {
			log.Error("zlimit region area(%s) ip(%v) forbid", r.Area, ip)
			err = ecode.NothingFound
			s.prom.Incr("region_limit_access")
		}
	}
	return
}

// checkAceess check user Aceess
func (s *Service) checkAceess(c context.Context, mid, aid int64, state, access int, _ string) (err error) {
	if state >= 0 && access == 0 {
		return
	}
	if state < 0 {
		if state == arcgrpc.StateForbidFixed {
			log.Warn("archive(%d) is fixed", aid)
		} else if state == arcgrpc.StateForbidUpDelete {
			log.Warn("archive(%d) is deleted", aid)
		} else {
			log.Warn("mid(%d) have not access view not pass archive(%d) ", mid, aid)
		}
		err = ecode.NothingFound
		return
	}
	if mid == 0 {
		log.Warn("not login can not view(%d) state(%d) access(%d)", aid, state, access)
		err = ecode.AccessDenied
		s.prom.Incr("no_login_access")
		return
	}
	card, err := s.accDao.Card3(c, mid)
	if err != nil || card == nil {
		if err != nil {
			log.Error("s.accDao.Info(%d) error(%v)", mid, err)
		}
		err = ecode.AccessDenied
		log.Warn("s.accDao.Info failed can not view(%d) state(%d) access(%d)", aid, state, access)
		s.prom.Incr("err_login_access")
		return
	}
	if access > 0 && int(card.Rank) < access && (card.Vip.Type == 0 || card.Vip.Status == 0 || card.Vip.Status == 2 || card.Vip.Status == 3) {
		err = ecode.AccessDenied
		log.Warn("mid(%d) rank(%d) vip(tp:%d,status:%d) have not access(%d) view archive(%d) ", mid, card.Rank, card.Vip.Type, card.Vip.Status, access, aid)
		s.prom.Incr("login_access")
	}
	return
}

// checkVIP check user is vip or no .
func (s *Service) checkVIP(c context.Context, mid int64) (vip bool) {
	var (
		card *account.Card
		err  error
	)
	if mid > 0 {
		if card, err = s.accDao.Card3(c, mid); err != nil || card == nil {
			log.Warn("s.acc.Info(%d) error(%v)", mid, err)
			return
		}
		vip = card.Vip.Type > 0 && card.Vip.Status == 1
	}
	return
}

func (s *Service) overseaCheck(reply *cfcgrpc.FlowCtlInfosV2Reply, aid int64, plat int8) bool {
	if getAttrBitValueFromInfosV2(reply, aid, model.OverseaBlockKey) && model.IsOverseas(plat) {
		s.prom.Incr("oversea_access")
		return true
	}
	return false
}

func getAttrBitValueFromInfosV2(reply *cfcgrpc.FlowCtlInfosV2Reply, aid int64, arcsAttrKey string) bool {
	//获取不到禁止项，默认为空
	if reply == nil {
		return false
	}
	val, ok := reply.ItemsMap[aid]
	if !ok {
		return false
	}
	for _, item := range val.Items {
		//处理reply
		if item.Key == arcsAttrKey {
			return item.Value == 1
		}
	}
	return false
}
