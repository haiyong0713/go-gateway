package view

import (
	"context"

	flowcontrolapi "git.bilibili.co/bapis/bapis-go/content-flow-control/service"
	"github.com/thoas/go-funk"

	"go-common/library/ecode"
	"go-common/library/log"
	"go-common/library/net/metadata"
	"go-gateway/app/app-svr/app-view/interface/model"
	"go-gateway/app/app-svr/archive/service/api"

	accApi "git.bilibili.co/bapis/bapis-go/account/service"

	locgrpc "git.bilibili.co/bapis/bapis-go/community/service/location"
)

// ipLimit ip limit
func (s *Service) ipLimit(c context.Context, mid, aid int64, cdnIP string) (down int64, err error) {
	var auth *locgrpc.Auth
	ip := metadata.String(c, metadata.RemoteIP)
	cfg := FromContextOrCreate(c, s.defaultViewConfigCreater())
	if auth, err = cfg.dep.Location.Archive(c, aid, mid, ip, cdnIP); err != nil {
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
func (s *Service) areaLimit(c context.Context, plat int8, rid int) (err error) {
	cfg := FromContextOrCreate(c, s.defaultViewConfigCreater())

	ip := metadata.String(c, metadata.RemoteIP)
	if rm, ok := s.region[plat]; !ok {
		return
	} else if r, ok := rm[rid]; ok && r != nil && r.Area != "" {
		var auths map[string]*locgrpc.Auth
		if auths, err = cfg.dep.Location.AuthPIDs(c, r.Area, ip); err != nil {
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

// checkAccess check user Access
func (s *Service) checkAccess(c context.Context, mid, aid int64, state, access int, arc *api.Arc) (err error) {
	cfg := FromContextOrCreate(c, s.defaultViewConfigCreater())
	//稿件仅自见,引导登录
	if arc.AttrValV2(api.AttrBitV2OnlySely) == api.AttrYes {
		if mid == 0 {
			return ecode.AccessDenied
		}
		if mid != arc.GetAuthor().Mid {
			return ecode.NothingFound
		}
	}
	//state > 0 或者首映稿件
	if (state >= 0 || (state == api.StateForbidUserDelay && arc.AttrValV2(api.AttrBitV2Premiere) == api.AttrYes)) &&
		access == 0 {
		return nil
	}
	if state < 0 {
		log.Error("archive(%d) state(%d) can not view", aid, state)
		return ecode.NothingFound
	}
	if mid == 0 && access > 0 {
		log.Error("not login can not view(%d) state(%d) access(%d)", aid, state, access)
		s.prom.Incr("no_login_access")
		return ecode.AccessDenied
	}
	card, err := cfg.dep.Account.Card3(c, mid)
	if err != nil || card == nil {
		if err != nil {
			log.Error("s.accDao.Card3(%d) error(%v) or card=nil", mid, err)
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
		card *accApi.Card
		err  error
	)
	cfg := FromContextOrCreate(c, s.defaultViewConfigCreater())
	if mid > 0 {
		if card, err = cfg.dep.Account.Card3(c, mid); err != nil || card == nil {
			log.Warn("s.acc.Card3(%d) error(%v) or card=nil", mid, err)
			return
		}
		vip = card.Vip.IsValid()
	}
	return
}

func (s *Service) overseaCheck(a *api.Arc, plat int8) bool {
	if a.AttrVal(api.AttrBitOverseaLock) == api.AttrYes && model.IsOverseas(plat) {
		s.prom.Incr("oversea_access")
		return true
	}
	return false
}

func (s *Service) overseaCheckV2(c context.Context, a *api.Arc, plat int8) bool {
	if !model.IsOverseas(plat) {
		return false
	}
	res, err := s.flowDao.GetCtlInfoV2(c, a.Aid)
	if err != nil {
		log.Error("overseaCheck is err %+v %+v", err, a.Aid)
		return false
	}
	if len(res.Items) == 0 {
		return false
	}
	//数组转map
	r := funk.Map(res.Items, func(item *flowcontrolapi.InfoItem) (string, int32) {
		return item.Key, item.Value
	})
	//InfoItem
	forbiddenItemsMap := r.(map[string]int32)
	forbidden, ok := forbiddenItemsMap["54"]
	if !ok {
		return false
	}
	if forbidden == 1 { //1-禁止
		return true
	}
	return false
}
