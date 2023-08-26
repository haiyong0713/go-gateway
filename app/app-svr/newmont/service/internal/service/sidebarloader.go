package service

import (
	"context"
	"fmt"
	"time"

	"go-gateway/app/app-svr/newmont/service/api"
	secmdl "go-gateway/app/app-svr/newmont/service/internal/model/section"

	"go-common/library/log"
	"go-common/library/sync/errgroup.v2"
)

type sidebarLoader struct {
	sideBarByModule   map[string][]*secmdl.SideBar
	sideBarLimitCache map[int64][]*secmdl.SideBarLimit
	language          map[int64]string
	entryModuleCache  map[int32][]*secmdl.ModuleInfo
	mineModuleCache   map[int32][]*secmdl.ModuleInfo
	hiddenCache       []*api.HiddenInfo
	hiddenLimits      map[int64][]*api.HiddenLimit
	sidebarTusValues  []string
	load              func() error
}

func (s *sidebarLoader) Timer() string {
	return "@every 5s"
}

func (s *sidebarLoader) Load() error {
	return s.load()
}

func (s *Service) loadSideBarCache() error {
	var (
		now          = time.Now()
		sidebar      []*secmdl.SideBar
		sbm          = make(map[string][]*secmdl.SideBar)
		limits       map[int64][]*secmdl.SideBarLimit
		sModules     map[int32][]*secmdl.ModuleInfo
		eModules     map[int32][]*secmdl.ModuleInfo
		language     map[int64]string
		hiddens      []*api.Hidden
		hiddenLimits map[int64][]*api.HiddenLimit
		hiddenInfos  []*api.HiddenInfo
	)
	eg := errgroup.WithCancel(context.Background())
	eg.Go(func(ctx context.Context) (err error) {
		if sidebar, err = s.sectionDao.SideBar(ctx, now); err != nil {
			return err
		}
		return nil
	})
	eg.Go(func(ctx context.Context) (err error) {
		// 获取我的页模块属性配置
		if sModules, err = s.sectionDao.SideBarModules(ctx, MType_MineSection); err != nil {
			return err
		}
		return nil
	})
	eg.Go(func(ctx context.Context) (err error) {
		// 获取首页tab模块属性配置
		if eModules, err = s.sectionDao.SideBarModules(ctx, MType_HomeTab); err != nil {
			return err
		}
		return nil
	})
	eg.Go(func(ctx context.Context) (err error) {
		if limits, err = s.sectionDao.SidebarLimit(context.Background()); err != nil {
			return err
		}
		return nil
	})
	eg.Go(func(ctx context.Context) (err error) {
		if language, err = s.sectionDao.SidebarLang(context.Background()); err != nil {
			return err
		}
		return nil
	})
	eg.Go(func(ctx context.Context) (err error) {
		if hiddens, err = s.sectionDao.Hiddens(context.Background(), now); err != nil {
			return err
		}
		return nil
	})
	eg.Go(func(ctx context.Context) (err error) {
		if hiddenLimits, err = s.sectionDao.HiddenLimits(context.Background()); err != nil {
			return err
		}
		return nil
	})
	if err := eg.Wait(); err != nil {
		log.Error("loadSideBarCache err(%+v)", err)
		return err
	}
	for _, item := range sidebar {
		item.Language = language[item.LanguageID]
	}
	for _, item := range sidebar {
		key := fmt.Sprintf(_initSidebarKey, item.Plat, item.Module, item.Language)
		sbm[key] = append(sbm[key], item)
	}
	for _, v := range hiddens {
		limit := hiddenLimits[v.Id]
		hiddenInfo := &api.HiddenInfo{
			Info:  v,
			Limit: limit,
		}
		hiddenInfos = append(hiddenInfos, hiddenInfo)
	}
	s.sideBarByModule = sbm
	s.sideBarLimitCache = limits
	s.mineModuleCache = sModules
	s.entryModuleCache = eModules
	s.hiddenCache = hiddenInfos
	s.sidebarTusValues = collectSidebarTusValues(sidebar)
	return nil
}

func collectSidebarTusValues(sidebars []*secmdl.SideBar) []string {
	var (
		tmpTusValues = make(map[string]struct{})
		tusValues    []string
	)
	for _, v := range sidebars {
		if v.TusValue == "" {
			continue
		}
		tmpTusValues[v.TusValue] = struct{}{}
	}
	for tusvalue := range tmpTusValues {
		tusValues = append(tusValues, tusvalue)
	}
	return tusValues
}
