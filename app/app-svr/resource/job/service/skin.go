package service

import (
	"context"
	"fmt"
	"go-common/library/log"
	feedMenuModel "go-gateway/app/app-svr/app-feed/admin/model/menu"
	api "go-gateway/app/app-svr/resource/service/api/v1"
	serviceModel "go-gateway/app/app-svr/resource/service/model"
	"time"
)

func (s *Service) loadSkinExt() {
	log.Info("Service: loadSkinExt Start")
	skinExtMap, err := s.getOnlineSkinExtMap()
	if err != nil {
		log.Error("Service: loadSkinExt getOnlineSkinExts Error (%v)", err)
		return
	}
	log.Info("Service: loadSkinExt getOnlineSkinExtMap Success")
	skinInfos := make([]*api.SkinInfo, 0)
	if len(skinExtMap) > 0 {
		for key := range skinExtMap {
			skinInfos = append(skinInfos, skinExtMap[key])
		}
	}
	if err = s.dao.SetSkinInfo2Cache(context.Background(), serviceModel.SkinExtCacheKey, skinInfos); err != nil {
		log.Error("Service: loadSkinExt SetSkinExt2Cache Error (%v)", err)
		return
	}
	log.Info("Service: loadSkinExt Success")
}

func (s *Service) getOnlineSkinExtMap() (reply map[string]*api.SkinInfo, err error) {
	var (
		exts   []*feedMenuModel.SkinExt
		limits map[int64][]*feedMenuModel.SkinLimit
		ids    []int64
	)

	// 获取当前生效的主题配置信息
	if exts, err = s.dao.GetSkinExts(time.Now()); err != nil {
		log.Error("Service: getOnlineSkinExts GetSkinExts Error (%v)", err)
		return
	} else if len(exts) == 0 {
		log.Info("Service: getOnlineSkinExts GetSkinExts got no online config")
		return
	}

	// 获取主题配置版本管理信息
	for _, v := range exts {
		ids = append(ids, v.ID)
	}
	if limits, err = s.dao.GetSkinLimits(ids); err != nil {
		return
	}

	// 构建主题配置信息
	reply = make(map[string]*api.SkinInfo)
	for _, ext := range exts {
		if lVal, ok := limits[ext.ID]; !ok || len(lVal) == 0 {
			continue
		}
		apiInfo := &api.SkinExt{
			ID:                ext.ID,
			SkinID:            ext.SkinID,
			SkinName:          ext.SkinName,
			Attribute:         ext.Attribute,
			State:             api.SkinExtState_Enum(ext.State),
			Ctime:             ext.Ctime,
			Mtime:             ext.Mtime,
			Stime:             ext.Stime,
			Etime:             ext.Etime,
			LocationPolicyGID: ext.LocationPolicyGroupID,
			UserScopeType:     ext.UserScopeType,
			UserScopeValue:    ext.UserScopeValue,
			DressUpType:       ext.DressUpType,
			DressUpValue:      ext.DressUpValue,
		}
		apiLimits := make([]*api.SkinLimit, 0)
		for _, limit := range limits[ext.ID] {
			apiLimits = append(apiLimits, &api.SkinLimit{
				ID:         limit.ID,
				SID:        limit.SID,
				Conditions: limit.Conditions,
				Build:      limit.Build,
				State:      api.SkinLimitState_Enum(limit.State),
				Plat:       int32(limit.Plat),
				Mtime:      limit.Mtime,
				Ctime:      limit.Ctime,
			})
		}
		tKey := fmt.Sprintf(serviceModel.InitSkinExtKey, ext.ID)
		reply[tKey] = &api.SkinInfo{
			Info:  apiInfo,
			Limit: apiLimits,
		}
	}
	return
}
