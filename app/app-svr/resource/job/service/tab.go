package service

import (
	"context"
	"fmt"
	"go-common/library/log"
	"go-gateway/app/app-svr/resource/job/model/show"
	"time"
)

const initTabExtKey = "tab_ext_%d_%d"

const tabExtCacheKey = "tab_ext"

func (s *Service) loadTabExt() {
	tabExts, err := s.buildTabExt()
	if err != nil {
		log.Error("cron:loadTabExt() fail:%+v", err)
		return
	}
	err = s.dao.SetTabExt2Cache(context.Background(), tabExtCacheKey, tabExts)
	if err != nil {
		log.Error("cron:loadTabExt() fail:%+v", err)
		return
	}
}

func (s *Service) buildTabExt() (rly map[string]*show.MenuExt, err error) {
	var (
		menus  []*show.TabExt
		ids    []int64
		limits map[int64][]*show.TabLimit
	)

	// 获取当前有效的tab资源配置信息
	if menus, err = s.dao.GetTabExts(time.Now()); err != nil || len(menus) == 0 {
		return
	}

	// 获取tab配置版本管理信息
	for _, v := range menus {
		ids = append(ids, v.ID)
	}
	if limits, err = s.dao.GetTabLimits(ids, show.MenuLimitType); err != nil {
		return
	}

	// 构建tab配置信息
	rly = make(map[string]*show.MenuExt)
	for _, val := range menus {
		if s.invalidAttribute(val) {
			continue
		}
		if lVal, okk := limits[val.ID]; !okk || len(lVal) == 0 {
			continue
		}
		tKey := fmt.Sprintf(initTabExtKey, val.TabID, val.Type)
		rly[tKey] = &show.MenuExt{TabExt: val, Limit: limits[val.ID]}
	}
	return
}

func (s *Service) invalidAttribute(val *show.TabExt) bool {
	return val.AttrVal(show.AttrBitImage) != show.AttrYes && val.AttrVal(show.AttrBitColor) != show.AttrYes &&
		val.AttrVal(show.AttrBitFollowBusiness) != show.AttrYes && val.AttrVal(show.AttrBitBgImage) != show.AttrYes
}
