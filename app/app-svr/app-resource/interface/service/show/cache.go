package show

import (
	"context"
	"fmt"
	"time"

	"go-common/library/log"

	bubblemdl "go-gateway/app/app-svr/app-resource/interface/model/bubble"
	"go-gateway/app/app-svr/app-resource/interface/model/show"
	resource "go-gateway/app/app-svr/resource/service/model"
)

const (
	_top    = 10
	_tab    = 8
	_bottom = 9
)

func (s *Service) loadTabCache() (err error) {
	var (
		tmp       = map[int64]struct{}{}
		ss        = map[string][]*show.Tab{}
		sideBars  *resource.SideBars
		redDot    = make(map[int8][]*show.SectionURL)
		moduleMap = map[int]struct{}{

			_top:    {},
			_tab:    {},
			_bottom: {},
		}
	)
	if sideBars, err = s.rdao.ResSideBar(context.TODO()); err != nil || sideBars == nil {
		log.Error("s.sideDao.SideBar error(%v) or nil", err)
		return
	}
	for _, v := range sideBars.SideBar {
		if _, ok := tmp[v.ID]; ok {
			continue
		}
		tmp[v.ID] = struct{}{}
		st := &show.Tab{}
		if !st.TabChange(v, _showAbtest, _deafaultTab) {
			continue
		}
		key := fmt.Sprintf(_initTabKey, st.Plat, st.Language)
		ss[key] = append(ss[key], st)
		if _, ok := moduleMap[v.Module]; ok {
			if v.Red != "" {
				redDot[v.Plat] = append(redDot[v.Plat], &show.SectionURL{ID: v.ID, URL: v.Red})
			}
		}
	}
	if len(ss) == 0 && len(s.tabCache) == 0 {
		err = fmt.Errorf("tabCache is null")
		return
	} else if len(ss) == 0 {
		return
	}
	s.tabCache = ss
	s.limitsCahce = sideBars.Limit
	s.redDot = redDot
	log.Info("loadTabCache cache success")
	return
}

func (s *Service) loadSkinExtCache() {
	log.Info("cronLog start loadSkinExtCache")
	rly, err := s.rdao.SkinConf(context.Background())
	if err != nil {
		log.Error("loadSkinExtCache error(%v)", err)
		return
	}
	s.skinCache = rly
	log.Info("loadSkinExtCache cache success")
}

func (s *Service) loadMenusCache(now time.Time) {
	menus, err := s.tdao.Menus(context.TODO(), now)
	if err != nil {
		log.Error("s.tab.Menus error(%v)", err)
		return
	}
	s.menuCache = menus
	log.Info("loadMenusCache cache success")
}

func (s *Service) loadAbTestCache() {
	var (
		groups string
	)
	for _, g := range _showAbtest {
		groups = groups + g + ","
	}
	if gLen := len(groups); gLen > 0 {
		groups = groups[:gLen-1]
	}
	res, err := s.rdao.AbTest(context.TODO(), groups)
	if err != nil {
		log.Error("resource s.rdao.AbTest error(%v)", err)
		return
	}
	s.abtestCache = res
	log.Info("loadAbTestCache cache success")
}

func (s *Service) loadCache() (err error) {
	log.Info("cronLog start show loadCache")
	now := time.Now()
	err = s.loadTabCache()
	s.loadMenusCache(now)
	s.loadAbTestCache()
	s.loadAuditCache()
	return
}

func (s *Service) loadShowTabAids() {
	tmp := map[int64]struct{}{}
	for _, mid := range s.c.ShowTabMids {
		tmp[mid] = struct{}{}
	}
	s.showTabMids = tmp
}

func (s *Service) loadBubbleCache() {
	log.Info("cronLog start loadBubbleCache")
	var (
		tmp map[int64]*bubblemdl.Bubble
		err error
	)
	if tmp, err = s.bubbleDao.Bubble(context.Background()); err != nil {
		log.Error("%v", err)
		return
	}
	s.bubbleCache = tmp
}
