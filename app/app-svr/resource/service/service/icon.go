package service

import (
	"context"
	"sync"
	"time"

	"go-common/library/log"
	"go-common/library/sync/errgroup.v2"

	v1 "go-gateway/app/app-svr/resource/service/api/v1"
	"go-gateway/app/app-svr/resource/service/model"
)

func (s *Service) loadIconCache() {
	ics, err := s.show.Icons(context.Background(), time.Now(), time.Now())
	if err != nil {
		log.Error("s.show.Icons error(%v)", err)
		return
	}
	s.IconCache = ics
	log.Info("loadIconCache success")

	preIcs, err := s.show.Icons(context.Background(),
		time.Now(),
		time.Now().Add(time.Duration(s.c.IconCacheConfig.PreloadDuration)*time.Hour))
	if err != nil {
		log.Error("loadPreIconCache s.show.Icons error(%v)", err)
		return
	}
	s.PreIconCache = preIcs
	log.Info("loadPreIconCache success")

}

// MngIcon .
func (s *Service) MngIcon(c context.Context, req *v1.MngIconRequest) (res *v1.MngIconReply, err error) {
	return s.mngIcon(c, req, s.IconCache)
}

// mngIcon .
func (s *Service) mngIcon(c context.Context, req *v1.MngIconRequest, iconCache map[int64]*v1.MngIcon) (res *v1.MngIconReply, err error) {
	res = new(v1.MngIconReply)
	var (
		tmpIcon  = make(map[int64]*v1.MngIcon, len(req.Oids))
		whiteMap = make(map[int64]bool)
	)
	if req.Mid > 0 { //如果是推送到指定用户走业务方接口获取
		var (
			mutex sync.Mutex
			eg    = errgroup.WithContext(c)
		)
		for _, oid := range req.Oids {
			ic, ok := iconCache[oid]
			if !ok || ic.EffectGroup != model.IconEffectGroupSpecial || ic.EffectUrl == "" {
				continue
			}
			tmpID := ic.Id
			tmpURL := ic.EffectUrl
			eg.Go(func(ctx context.Context) (err error) {
				ok, err := s.show.EffectUrl(ctx, req.Mid, tmpURL)
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
		if ic.EffectGroup == model.IconEffectGroupLogin && req.Mid == 0 { //登录用户推送
			continue
		}
		if ic.EffectGroup == model.IconEffectGroupSpecial && (ic.EffectUrl == "" || !whiteMap[ic.Id]) { //指定用户推送
			continue
		}
		tmpIcon[oid] = ic
	}
	res.Info = tmpIcon
	return
}
