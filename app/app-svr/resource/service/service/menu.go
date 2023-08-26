package service

import (
	"context"
	"sort"
	"time"

	"go-common/library/ecode"
	"go-common/library/log"

	api "go-gateway/app/app-svr/resource/service/api/v1"
	"go-gateway/app/app-svr/resource/service/model"
)

func (s *Service) Menu(c context.Context, arg *api.NoArgRequest) (res *api.MenuReply, err error) {
	list, err := s.menuDao.Menus(c)
	if err != nil {
		log.Error("%+v", err)
	}
	res = &api.MenuReply{List: list}
	return
}

func (s *Service) Active(c context.Context, arg *api.NoArgRequest) (res *api.ActiveReply, err error) {
	list, err := s.menuDao.Actives(c)
	if err != nil {
		log.Error("%+v", err)
	}
	res = &api.ActiveReply{List: list}
	return
}

func (s *Service) AppActive(c context.Context, arg *api.AppActiveRequest) (*api.AppActiveReply, error) {
	res := &api.AppActiveReply{}
	// 必须要有内容
	acs, ok := s.activemCache[arg.Id]
	if !ok {
		return nil, ecode.NothingFound
	}
	// 背景图可以为空
	res.Cover = s.activcovermCache[arg.Id]
	res.List = acs
	return res, nil
}

func (s *Service) loadActiveCache() {
	const (
		_background = "common"
	)
	acs, err := s.menuDao.Actives(context.Background())
	if err != nil {
		log.Error("%+v", err)
		return
	}
	coverm := make(map[int64]string, len(acs))
	parentm := make(map[int64]struct{}, len(acs))
	for _, ac := range acs {
		if ac.Type == _background {
			parentm[ac.Id] = struct{}{}
			coverm[ac.Id] = ac.Background
		}
	}
	sort.Slice(acs, func(i, j int) bool {
		return acs[i].Id < acs[j].Id
	})
	tabm := make(map[int64][]*api.Active, len(acs))
	for parentID := range parentm {
		for _, ac := range acs {
			if ac.ParentID == parentID {
				tabm[ac.ParentID] = append(tabm[ac.ParentID], ac)
			}
		}
	}
	s.activemCache = tabm
	s.activcovermCache = coverm
}

func (s *Service) AppMenu(c context.Context, arg *api.AppMenusRequest) (*api.AppMenuReply, error) {
	memuCache := s.menuCache
	menus := make([]*api.AppMenu, 0, len(memuCache))
	res := &api.AppMenuReply{}
	now := time.Now()
LOOP:
	for _, m := range memuCache {
		if vs, ok := m.Versions[arg.Plat]; ok {
			for _, v := range vs {
				if model.InvalidBuild(int(arg.Build), v.Build, v.Condition) {
					continue LOOP
				}
			}
			if m.Status == 1 && (m.STime == 0 || now.After(m.STime.Time())) && (m.ETime == 0 || now.Before(m.ETime.Time())) {
				menus = append(menus, &api.AppMenu{
					TabId: m.TabID,
					Name:  m.Name,
					Img:   m.Img,
					Icon:  m.Icon,
					Color: m.Color,
					Id:    m.ID,
				})
			}
		}
	}
	res.List = menus
	return res, nil
}

func (s *Service) loadAppMenuCache() {
	list, err := s.menuDao.AllMenus(context.Background())
	if err != nil {
		log.Error("%+v", err)
		return
	}
	s.menuCache = list
}

func (s *Service) loadSkinCache() (skinInfos []*api.SkinInfo, err error) {
	if skinInfos, err = s.menuDao.GetSkinInfosFromRedis(context.Background(), model.SkinExtCacheKey); err != nil {
		log.Error("s.show.RawSkinExts() error(%v)", err)
		return
	}
	if len(skinInfos) == 0 {
		skinInfos = make([]*api.SkinInfo, 0)
		//log.Warn("resource-service.Service.loadSkinCache get empty SkinExts")
		return
	}
	log.Info("loadSkinCache success")
	return
}

// SkinConf .
func (s *Service) SkinConf(_ context.Context, _ *api.NoArgRequest) (rly *api.SkinConfReply, err error) {
	var skinExtCache []*api.SkinInfo
	if skinExtCache, err = s.loadSkinCache(); err != nil {
		log.Error("resource-service.Service.SkinConf Error (%v)", err)
		return
	}
	rly = &api.SkinConfReply{
		List: skinExtCache,
	}
	return
}
