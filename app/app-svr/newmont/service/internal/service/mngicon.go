package service

import (
	"context"
	"sync"

	"go-gateway/app/app-svr/newmont/service/api"

	"go-common/library/log"
	"go-common/library/sync/errgroup.v2"
)

const (
	IconEffectGroupLogin   = 2
	IconEffectGroupSpecial = 3
)

func (s *Service) MngIcon(ctx context.Context, req *api.MngIconRequest) (*api.MngIconReply, error) {
	return s.mngIcon(ctx, req, s.IconCache)
}

func (s *Service) mngIcon(c context.Context, req *api.MngIconRequest, iconCache map[int64]*api.MngIcon) (res *api.MngIconReply, err error) {
	res = new(api.MngIconReply)
	var (
		tmpIcon  = make(map[int64]*api.MngIcon, len(req.Oids))
		whiteMap = make(map[int64]bool)
	)
	if req.Mid > 0 { //如果是推送到指定用户走业务方接口获取
		var (
			mutex sync.Mutex
			eg    = errgroup.WithContext(c)
		)
		for _, oid := range req.Oids {
			ic, ok := iconCache[oid]
			if !ok || ic.EffectGroup != IconEffectGroupSpecial || ic.EffectUrl == "" {
				continue
			}
			tmpID := ic.Id
			tmpURL := ic.EffectUrl
			eg.Go(func(ctx context.Context) (err error) {
				ok, err := s.sectionDao.EffectUrl(ctx, req.Mid, tmpURL)
				if err != nil {
					log.Error("s.show.EffectUrl error(%+v) mid(%d)", err, req.Mid)
					return nil
				}
				if ok {
					mutex.Lock()
					whiteMap[tmpID] = ok
					mutex.Unlock()
				}
				return nil
			})
		}
		if err = eg.Wait(); err != nil {
			log.Error("MngIcon eg wait err: %s", err)
		}
	}
	for _, oid := range req.Oids {
		ic, ok := iconCache[oid]
		if !ok {
			continue
		}
		if ic.EffectGroup == IconEffectGroupLogin && req.Mid == 0 { //登录用户推送
			continue
		}
		if ic.EffectGroup == IconEffectGroupSpecial && (ic.EffectUrl == "" || !whiteMap[ic.Id]) { //指定用户推送
			continue
		}
		tmpIcon[oid] = ic
	}
	res.Info = tmpIcon
	return
}
