package frontpage

import (
	"context"
	"go-gateway/app/app-svr/app-feed/admin/util"
	"time"

	xecode "go-common/library/ecode"
	"go-common/library/log"
	"go-common/library/sync/errgroup.v2"

	model "go-gateway/app/app-svr/app-feed/admin/model/frontpage"
	"go-gateway/app/app-svr/app-feed/ecode"
)

func (s *Service) GetConfigs(ctx context.Context, resourceID int64, pn int64, ps int64) (res model.FrontpagesForFE, total int64, err error) {
	if pn == 0 {
		pn = 1
	}
	if ps == 0 {
		ps = 5
	}
	res = model.FrontpagesForFE{}
	eg := errgroup.WithContext(ctx)
	// online
	eg.Go(func(_ context.Context) (_err error) {
		if res.OnlineConfigs, _err = s.dao.GetOnlineConfigs(resourceID); _err != nil {
			log.Error("Service: GetConfigs GetOnlineConfig %d error %v", resourceID, _err)
		}
		return
	})
	// hidden
	eg.Go(func(_ context.Context) (_err error) {
		if res.HiddenConfigs, total, _err = s.dao.GetHiddenConfigs(resourceID, pn, ps); _err != nil {
			log.Error("Service: GetConfigs GetOnlineConfig %d error %v", resourceID, _err)
		}
		return
	})
	// default
	eg.Go(func(_ context.Context) (_err error) {
		if res.DefaultConfig, _err = s.dao.GetBaseDefaultConfig(); _err != nil {
			log.Error("Service: GetConfigs GetDefaultConfig error %v", _err)
		}
		return
	})

	if err = eg.Wait(); err != nil {
		log.Error("Service: GetConfigs eg.Wait error %v", err)
	}

	return
}

func (s *Service) GetConfig(resourceID int64, id int64) (res *model.Config, err error) {
	if resourceID < 0 || id <= 0 {
		err = xecode.RequestErr
		return
	}
	if res, err = s.dao.GetConfig(id); err != nil {
		log.Error("Service: GetConfig (%d, %d) error %v", resourceID, id, err)
		return
	}

	return
}

func (s *Service) CheckConfigDuplicated(resourceID int64, stime time.Time, etime time.Time, locationPolicyGroupID int64, newConfigID int64) (duplicated bool, err error) {
	var (
		existsConfig *model.Config
	)
	if existsConfig, err = s.dao.GetConfigByTimeAndLoc(resourceID, stime, etime, locationPolicyGroupID); err != nil {
		log.Error("Service: CheckConfigDuplicated GetConfigBySTime (%d, %s, %d) error %v", resourceID, stime.Format(model.DefaultTimeLayout), locationPolicyGroupID, err)
		return
	} else if existsConfig != nil && existsConfig.ID > 0 && newConfigID != existsConfig.ID {
		duplicated = true
	}
	return
}

func (s *Service) AddConfig(toAddConfig model.Config, username string) (id int64, err error) {
	var (
		addedConfig = &toAddConfig
	)
	if addedConfig.STime.Time().After(addedConfig.ETime.Time()) {
		return 0, xecode.Error(xecode.RequestErr, "结束时间不能在起始时间前")
	}

	addedConfig.CUser = username
	if addedConfig, err = s.dao.AddConfig(addedConfig); err != nil {
		log.Error("Service: AddConfig AddConfig %+v error %v", addedConfig, err)
		return
	}
	id = addedConfig.ID

	// 行为日志
	logParams := []interface{}{
		toAddConfig.ConfigName,
	}
	if _err := util.AddFrontpageConfigLogs(username, 0, id, "add", int(toAddConfig.ResourceID), logParams, toAddConfig); _err != nil {
		log.Error("Service: AddConfig AddFrontpageConfigLogs error %v", _err)
	}

	return
}

func (s *Service) EditConfig(configID int64, resourceID int64, updateMap map[string]interface{}, username string) (err error) {
	if configID == 0 {
		return xecode.Error(xecode.RequestErr, "Config ID缺失")
	}
	if len(updateMap) == 0 {
		return
	}

	// validate exists
	var (
		existsConfig *model.Config
	)
	if existsConfig, err = s.dao.GetConfig(configID); err != nil {
		log.Error("Service: EditConfig GetConfig (%d) error %v", configID, err)
		return
	} else if existsConfig == nil || existsConfig.ID == 0 {
		log.Error("Service: EditConfig GetConfig (%d) 未找到对应配置", configID)
		err = ecode.FrontPageConfigNotFound
		return
	}

	// compare stime <= etime
	var (
		toCompareSTime = existsConfig.STime.Time()
		toCompareETime = existsConfig.ETime.Time()
	)
	if _, exists := updateMap["stime"]; exists {
		if stimeTime, ok := updateMap["stime"].(time.Time); ok {
			toCompareSTime = stimeTime
		}
	}
	if _, exists := updateMap["etime"]; exists {
		if etimeTime, ok := updateMap["etime"].(time.Time); ok {
			toCompareETime = etimeTime
		}
	}
	if toCompareSTime.After(toCompareETime) {
		return xecode.Error(xecode.RequestErr, "结束时间不能在起始时间前")
	}

	// 最终兜底图不允许修改时间和loc
	if configID == model.DefaultConfigID {
		delete(updateMap, "stime")
		delete(updateMap, "etime")
		delete(updateMap, "loc_policy_group_id")
	}

	// validate dup
	if configID != model.DefaultConfigID {
		var (
			duplicated            bool
			stime                 time.Time
			etime                 time.Time
			locationPolicyGroupID int64
		)
		if stimeVal, exists := updateMap["stime"]; exists {
			stime = stimeVal.(time.Time)
		} else {
			stime = existsConfig.STime.Time()
		}
		if etimeVal, exists := updateMap["etime"]; exists {
			etime = etimeVal.(time.Time)
		} else {
			etime = existsConfig.ETime.Time()
		}
		if locationPolicyGroupIDVal, exists := updateMap["loc_policy_group_id"]; exists {
			locationPolicyGroupID = locationPolicyGroupIDVal.(int64)
		} else {
			locationPolicyGroupID = existsConfig.LocPolicyGroupID
		}
		if duplicated, err = s.CheckConfigDuplicated(resourceID, stime, etime, locationPolicyGroupID, configID); err != nil {
			log.Error("Service: EditConfig CheckConfigDuplicated (%d, %s, %d, %d) error %v", resourceID, stime.Format(model.DefaultTimeLayout), locationPolicyGroupID, configID, err)
			return
		} else if duplicated {
			err = ecode.FrontPageConfigDuplicated
			return
		}
	}

	updateMap["username"] = username
	if err = s.dao.UpdateConfig(configID, updateMap); err != nil {
		log.Error("Service: EditConfig UpdateConfig %d %+v error %v", configID, updateMap, err)
		return
	}

	// 行为日志
	logParams := []interface{}{
		existsConfig.ConfigName,
		updateMap["config_name"],
	}
	if _err := util.AddFrontpageConfigLogs(username, 0, configID, "edit", int(resourceID), logParams, updateMap); _err != nil {
		log.Error("Service: EditConfig AddFrontpageConfigLogs error %v", _err)
	}

	return
}

// ActionConfig 操作配置上下线
func (s *Service) ActionConfig(configID int64, action string, username string) (err error) {
	if configID == 0 || action == "" {
		return xecode.RequestErr
	}

	// validate exists
	var (
		existsConfig *model.Config
	)
	if existsConfig, err = s.dao.GetConfig(configID); err != nil {
		log.Error("Service: ActionConfig GetConfig (%d) error %v", configID, err)
		return
	} else if existsConfig == nil || existsConfig.ID == 0 {
		log.Error("Service: ActionConfig GetConfig (%d) 未找到对应配置", configID)
		err = ecode.FrontPageConfigNotFound
		return
	}

	// 最终兜底图不允许操作
	if configID == model.DefaultConfigID {
		err = ecode.FrontPageNotEditable
		return
	}

	// 若要上线，先检查线上是否存在冲突时间配置
	if action == model.ActionOnline {
		// validate dup
		if configID != model.DefaultConfigID {
			var (
				duplicated bool
			)
			if duplicated, err = s.CheckConfigDuplicated(existsConfig.ResourceID, existsConfig.STime.Time(), existsConfig.ETime.Time(), existsConfig.LocPolicyGroupID, configID); err != nil {
				log.Error("Service: ActionConfig CheckConfigDuplicated (%d, %s, %d, %d) error %v", existsConfig.ResourceID, existsConfig.STime.Time().Format(model.DefaultTimeLayout), existsConfig.LocPolicyGroupID, configID, err)
				return
			} else if duplicated {
				err = ecode.FrontPageConfigDuplicated
				return
			}
		}
	}

	updateMap := make(map[string]interface{})
	switch action {
	case model.ActionOnline:
		updateMap["state"] = 0
	case model.ActionHidden:
		updateMap["state"] = 1
	case model.ActionDelete:
		updateMap["state"] = -1
	default:
		err = xecode.RequestErr
		return
	}
	updateMap["username"] = username
	if err = s.dao.UpdateConfig(configID, updateMap); err != nil {
		log.Error("Service: ActionConfig UpdateConfig %d %+v error %v", configID, updateMap, err)
		return
	}

	// 行为日志
	logParams := []interface{}{
		existsConfig.ConfigName,
	}
	if _err := util.AddFrontpageConfigLogs(username, 0, configID, action, int(existsConfig.ResourceID), logParams, updateMap); _err != nil {
		log.Error("Service: ActionConfig AddFrontpageConfigLogs error %v", _err)
	}

	return
}

func (s *Service) GetConfigHistories(resourceID int64, pn int64, ps int64) (res []*model.Config, total int64, err error) {
	if res, total, err = s.dao.GetConfigHistories(resourceID, pn, ps); err != nil {
		log.Error("Service: GetConfigHistories (%d) error %v", resourceID, err)
	}

	return
}
