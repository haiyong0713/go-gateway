package resource

import (
	"context"
	"math/rand"

	"go-common/library/log"
	"go-common/library/net/metadata"
	"go-gateway/app/web-svr/web-show/interface/dao/ad"
	resmdl "go-gateway/app/web-svr/web-show/interface/model/resource"

	account "git.bilibili.co/bapis/bapis-go/account/service"
)

var (
	_emptyVideoAds = []*resmdl.VideoAD{}
)

// VideoAd get videoad by aid
func (s *Service) VideoAd(c context.Context, arg *resmdl.ArgAid) (res []*resmdl.VideoAD) {
	arg.IP = metadata.String(c, metadata.RemoteIP)
	if arg.Mid > 0 {
		// ignore error
		var (
			resPro *account.Card
			err    error
		)
		if resPro, err = s.user(c, arg.Mid, arg.IP); err == nil {
			if s.normalVip(c, resPro) {
				return
			}
		}
		// NOTE cache?
		if isBp := s.bangumiDao.IsBp(c, arg.Mid, arg.Aid, arg.IP); isBp {
			log.Info("mid(%d) aid(%d) ip(%s) is bp", arg.Mid, arg.Aid, arg.IP)
			res = _emptyVideoAds
			return
		}
	}
	if res = s.videoAdByAid(arg.Aid); len(res) == 0 {
		res = _emptyVideoAds
	}
	return
}

func (s *Service) user(c context.Context, mid int64, ip string) (resPro *account.Card, err error) {
	resp, err := s.accClient.Card3(c, &account.MidReq{Mid: mid, RealIp: ip})
	if err != nil {
		ad.PromError("accClient.Card3", "s.accRPC.Info2() err(%v)", err)
		log.Error("accClient.Card3 err(%v)", err)
		return nil, err
	}
	return resp.GetCard(), nil
}

// checkVip check normal vip
func (s *Service) normalVip(_ context.Context, pro *account.Card) bool {
	if pro == nil {
		return false
	}
	if pro.Vip.Type != 0 && pro.Vip.Status == 1 {
		return true
	}
	return false
}

func (s *Service) videoAdByAid(aid int64) (res []*resmdl.VideoAD) {
	ss := s.videoCache[aid]
	l := len(ss)
	if l == 0 {
		return
	}
	// NOTE this means StrategyOnly
	if l == 1 {
		res = ss[0]
		return
	}
	// NOTE this means StrategyShare
	res = ss[rand.Intn(l)]
	return
}

// loadVideoAd load videoad to cache
func (s *Service) loadVideoAd() {
	if s.videoAdRunning {
		return
	}
	s.videoAdRunning = true
	defer func() {
		s.videoAdRunning = false
	}()
	ads, err := s.resdao.VideoAds(context.Background())
	if err != nil {
		log.Error("s.resdao.VideoAds error(%v)", err)
		return
	}
	tmp := make(map[int64][][]*resmdl.VideoAD)
	for aid, vads := range ads {
		if len(vads) < 1 {
			continue
		}
		if vads[0].Strategy == resmdl.StrategyOnly || vads[0].Strategy == resmdl.StrategyRank {
			tmp[aid] = append(tmp[aid], vads)
		} else if vads[0].Strategy == resmdl.StrategyShare {
			for _, vad := range vads {
				tmp[aid] = append(tmp[aid], []*resmdl.VideoAD{vad})
			}
		}
	}
	s.videoCache = tmp
}
