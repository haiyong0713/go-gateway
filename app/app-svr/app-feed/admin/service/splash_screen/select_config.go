package splash_screen

import (
	"context"
	"strings"
	"sync"
	"time"

	xecode "go-common/library/ecode"
	"go-common/library/log"
	"go-common/library/sync/errgroup.v2"
	xtime "go-common/library/time"

	"go-gateway/app/app-svr/app-feed/admin/api"
	splashModel "go-gateway/app/app-svr/app-feed/admin/model/splash_screen"
	Log "go-gateway/app/app-svr/app-feed/admin/util"
	feedEcode "go-gateway/app/app-svr/app-feed/ecode"
)

/**
 * -------------- 自选配置分类  --------------
 */

// SaveAllCategories 保存所有分类
func (s *Service) SaveAllCategories(items []*splashModel.CategoryForSaving, username string, uid int64) (res []*splashModel.Category, err error) {
	toSaveItems := make([]*splashModel.Category, len(items))
	for i := range items {
		if items[i].Name != "" && items[i].Name != "全部" {
			toSaveItems[i] = items[i].Category
		}
	}
	if res, err = s.dao.SaveCategories(toSaveItems, username); err != nil {
		log.Error("Service: SaveAllCategories SaveCategories (%v) error %v", items, err)
		return nil, err
	}

	obj := map[string]interface{}{
		"items": res,
	}
	// 行为日志
	if _err := Log.AddLog(splashModel.ActionLogBusiness, username, uid, 0, "SaveAllCategories", obj); _err != nil {
		log.Error("Service: SaveAllCategories AddLog error(%v)", _err)
	}

	return
}

// GetAllCategories 获取所有分类
func (s *Service) GetAllCategories(withAllPrefix bool, state api.SplashScreenConfigState_Enum) (res []*splashModel.CategoryWithConfigCount, err error) {
	if res, err = s.dao.GetAllCategoriesWithConfigsCount(state); err != nil {
		log.Error("Service: GetAllCategories GetAllCategoriesWithConfigsCount error %v", err)
		return
	}
	if withAllPrefix {
		//nolint:ineffassign
		var allCount int32 = 0
		if allCount, err = s.dao.GetCategorySelectConfigCount(0, state); err != nil {
			log.Error("Service: GetAllCategories GetSelectConfigCount error %v", err)
			return
		}
		var sort int64 = 1
		if len(res) > 0 {
			sort = res[0].Sort + 1
		}
		res = append([]*splashModel.CategoryWithConfigCount{{
			Category: &splashModel.Category{
				ID:   0,
				Name: "全部",
				Sort: sort,
			},
			ConfigCount: allCount,
		}}, res...)
	}

	return
}

/**
 * -------------- 自选配置 --------------
 */

// GetSelectConfigs 获取自选配置
//
//nolint:gocognit
func (s *Service) GetSelectConfigs(imageID int64, categoryID int64, state api.SplashScreenConfigState_Enum, sorting string, pn int32, ps int32) (res []*splashModel.SelectConfig, total int32, err error) {
	if imageID > 0 {
		var config *splashModel.SelectConfig
		if config, err = s.dao.GetSelectConfigByImageID(imageID); err != nil {
			log.Error("Service: GetSelectConfigs GetSelectConfigByImageID %d error %v", imageID, err)
			return
		} else if config == nil || config.ID == 0 {
			err = feedEcode.SplashScreenConfigNotExists
			return
		} else {
			res = []*splashModel.SelectConfig{config}
		}
		return
	} else {
		if pn == 0 {
			pn = 1
		}
		if ps == 0 {
			ps = 20
		}
		//nolint:ineffassign,staticcheck
		allConfigs := make([]*splashModel.SelectConfig, 0)
		if allConfigs, total, err = s.dao.GetSelectConfigList(sorting); err != nil {
			log.Error("Service: GetSelectConfigs GetSelectConfigListByPage (%d, %d) error %v", pn, ps, err)
			return
		}
		if total > 0 {
			var (
				//nolint:ineffassign
				beginI int32 = 0
				//nolint:ineffassign
				endI            int32 = 0
				filteredConfigs       = make([]*splashModel.SelectConfig, 0)
			)
			for _, config := range allConfigs {
				_config := config
				// 过滤
				if categoryID > 0 {
					exists := false
					for _, configCategoryID := range _config.CategoryIDs {
						if categoryID == configCategoryID {
							exists = true
						}
					}
					if !exists {
						continue
					}
				}
				// state
				if state >= 0 {
					switch state {
					case api.SplashScreenConfigState_DEPRECATED, api.SplashScreenConfigState_REJECTED:
						if _config.State != api.SplashScreenConfigState_DEPRECATED && _config.State != api.SplashScreenConfigState_REJECTED {
							continue
						}
					default:
						if state != _config.State {
							continue
						}
					}
				}

				filteredConfigs = append(filteredConfigs, _config)
			}
			total = int32(len(filteredConfigs))
			if (pn-1)*ps >= int32(len(filteredConfigs)) {
				return
			} else if pn*ps > int32(len(filteredConfigs)) {
				endI = int32(len(filteredConfigs))
			} else {
				endI = pn * ps
			}
			beginI = (pn - 1) * ps
			res = filteredConfigs[beginI:endI]
			// show sort
			if strings.ToLower(sorting) == "sort desc" {
				for i, config := range res {
					_config := config
					_config.ShowSort = int64(beginI) + 1 + int64(i)
				}
			} else {
				for i, config := range res {
					_config := config
					_config.ShowSort = int64(len(filteredConfigs)) - int64(beginI) - int64(i)
				}
			}
		}
	}

	return
}

// SaveSelectConfigs 批量保存自选配置
func (s *Service) SaveSelectConfigs(ctx context.Context, items []*splashModel.SelectConfigForSaving, username string, uid int64) (res []*splashModel.SelectConfig, err error) {
	if len(items) == 0 {
		return
	}
	var (
		duplicateMap = make(map[int64]int)
		toUpdateIDs  = make([]int64, 0, len(items))
	)
	for _, item := range items {
		if _, exists := duplicateMap[item.ImageID]; exists {
			err = xecode.Error(xecode.RequestErr, "批次存在重复物料配置")
		} else {
			duplicateMap[item.ImageID] = 1
		}
		if item.ID > 0 {
			toUpdateIDs = append(toUpdateIDs, item.ID)
		}
	}
	res = make([]*splashModel.SelectConfig, 0)

	var (
		//nolint:ineffassign
		existingSelectConfigs = make([]*splashModel.SelectConfig, 0)
		existingSorts         = make([]int64, 0)
	)

	// 获取已存在配置及sort，后续更新交换位置
	if len(toUpdateIDs) > 0 {
		if existingSelectConfigs, err = s.dao.GetSelectConfigsByIDs(toUpdateIDs); err != nil {
			log.Error("Service: SaveSelectConfigs GetSelectConfigsByIDs %v error %v", toUpdateIDs, err)
			return
		}
		if len(existingSelectConfigs) > 0 {
			// DESC
			for _, config := range existingSelectConfigs {
				if len(existingSorts) == 0 || config.Sort > existingSorts[0] {
					existingSorts = append([]int64{config.Sort}, existingSorts...)
				} else {
					existingSorts = append(existingSorts, config.Sort)
				}
			}
			maxSort := existingSorts[0]
			minSort := existingSorts[len(existingSorts)-1]
			if (maxSort - minSort + 1) < int64(len(items)) {
				err = xecode.Error(xecode.RequestErr, "sort值数据错误。请联系后台开发")
				return
			}
		}
	}

	eg := errgroup.WithContext(ctx)
	var lock sync.Mutex
	for i, item := range items {
		_item := item
		_item.AuditState = 0
		if len(existingSorts) > 0 {
			_item.Sort = existingSorts[0] - int64(i)
		}
		eg.Go(func(_ context.Context) error {
			if savedConfig, _err := s.SaveSelectConfig(_item.SelectConfig, username, uid); _err != nil {
				return _err
			} else {
				lock.Lock()
				res = append(res, savedConfig)
				lock.Unlock()
			}
			return nil
		})
	}
	err = eg.Wait()

	return
}

// SaveSelectConfig 新建/更新单个自选配置
func (s *Service) SaveSelectConfig(item *splashModel.SelectConfig, username string, uid int64) (res *splashModel.SelectConfig, err error) {
	if item == nil || item.ImageID == 0 {
		return nil, xecode.RequestErr
	}

	// 先检查/删除已存在的物料对应的自选配置
	var existingImageConfig *splashModel.SelectConfig
	if existingImageConfig, err = s.dao.GetSelectConfigByImageID(item.ImageID); err != nil {
		log.Error("Service: SaveSelectConfig GetSelectConfigByImageID %d error %v", item.ImageID, err)
		return
	} else if existingImageConfig != nil && existingImageConfig.ID > 0 && existingImageConfig.ID != item.ID {
		if err = s.dao.DeleteSelectConfig(existingImageConfig.ID, username); err != nil {
			log.Error("Service: SaveSelectConfig DeleteSelectConfig %d error %v", item.ID, err)
			return
		}
	}

	if item.ID == 0 {
		if item.Sort == 0 {
			if newSort, _err := s.dao.GetNewSelectConfigSort(); _err != nil {
				log.Error("Service: SaveSelectConfig %+v GetNewSelectConfigSort error %v", item, _err)
				err = _err
				return
			} else {
				item.Sort = newSort
			}
		}
		if res, err = s.dao.AddSelectConfig(item, username); err != nil {
			log.Error("Service: SaveSelectConfig AddSelectConfig %+v error %v", item, err)
			return nil, err
		}
	} else {
		if exists, _err := s.CheckSelectConfigExist(item.ID); _err != nil {
			log.Error("Service: SaveSelectConfig CheckSelectConfigExist %d error %v", item.ID, _err)
			return nil, _err
		} else if !exists {
			log.Error("Service: SaveSelectConfig %d 配置不存在", item.ID)
			return nil, feedEcode.SplashScreenConfigNotExists
		}
		updateMap := map[string]interface{}{
			"stime":       item.STime,
			"etime":       item.ETime,
			"sort":        item.Sort,
			"audit_state": item.AuditState,
			"m_user":      username,
			"mtime":       xtime.Time(time.Now().Unix()),
		}
		if res, err = s.dao.UpdateSelectConfig(item.ID, updateMap, username); err != nil {
			log.Error("Service: SaveSelectConfig UpdateSelectConfig (%d, %+v) error %v", item.ID, updateMap, err)
		}
		// 删除原有配置分类关系
		if err = s.dao.DeleteSelectConfigCategoryRelsByConfig(item.ID, username); err != nil {
			log.Error("Service: SaveSelectConfig DeleteSelectConfigCategoryRelsByConfig %d error %v", item.ID, err)
			return
		}
	}
	// 添加配置分类关系
	if len(item.CategoryIDs) > 0 {
		rels := make([]*splashModel.SelectConfigCategoryRel, 0, len(item.CategoryIDs))
		for _, categoryID := range item.CategoryIDs {
			rels = append(rels, &splashModel.SelectConfigCategoryRel{
				ConfigID:   item.ID,
				CategoryID: categoryID,
				CUser:      username,
				CTime:      xtime.Time(time.Now().Unix()),
			})
		}
		if _, err = s.dao.AddSelectConfigCategoryRels(rels, username); err != nil {
			log.Error("Service: SaveSelectConfig AddSelectConfigCategoryRels %+v error %v", rels, err)
			return
		}
	}

	obj := map[string]interface{}{
		"value": item,
	}
	if _err := Log.AddLog(splashModel.ActionLogBusiness, username, uid, res.ID, "SaveSelectConfig", obj); _err != nil {
		log.Error("Service: UpdateAuditState AddLog error(%v)", _err)
	}

	return
}

// UpdateSelectConfigSort 置顶/置底配置
func (s *Service) UpdateSelectConfigSort(id int64, sortType string, username string, uid int64) (err error) {
	if id == 0 || (sortType != "top" && sortType != "bottom") {
		return xecode.RequestErr
	}

	if exists, _err := s.CheckSelectConfigExist(id); _err != nil {
		//nolint:govet
		log.Error("Service: UpdateSelectConfigSort CheckSelectConfigExist %d error %v", exists, _err)
		return _err
	} else if !exists {
		log.Error("Service: UpdateSelectConfigSort %d 配置不存在", id)
		return feedEcode.SplashScreenConfigNotExists
	}

	var sortVal int64
	if sortVal, err = s.dao.GetNewSelectConfigSort(); err != nil {
		log.Error("Service: UpdateSelectConfigSort GetNewSelectConfigSort error %v", err)
		return
	}
	if sortType == "top" {
		// 置顶。则直接用新生成的sort值
		updateMap := map[string]interface{}{
			"sort":   sortVal,
			"m_user": username,
			"mtime":  xtime.Time(time.Now().Unix()),
		}
		if _, err = s.dao.UpdateSelectConfig(id, updateMap, username); err != nil {
			log.Error("Service: UpdateSelectConfigSort UpdateSelectConfig (%d, %+v) error %v", id, updateMap, err)
			return
		}
	} else {
		// 置底。则将此配置sort置为1，并将其余sort+=1
		if err = s.dao.UpdateSelectConfigSortBottom(id, username); err != nil {
			//nolint:govet
			log.Error("Service: UpdateSelectConfigSort UpdateSelectConfig (%d, %+v) error %v", id, err)
			return
		}
	}

	if _err := Log.AddLog(splashModel.ActionLogBusiness, username, uid, id, "UpdateSelectConfigSort", sortType); _err != nil {
		log.Error("Service: UpdateSelectConfigSort AddLog error(%v)", _err)
	}

	return
}

// DeleteSelectConfigs 删除配置
func (s *Service) DeleteSelectConfigs(ids []int64, username string, uid int64) (err error) {
	if len(ids) == 0 {
		return
	}

	if err = s.dao.DeleteSelectConfigs(ids, username); err != nil {
		log.Error("Service: DeleteSelectConfigs DeleteSelectConfigs (%v) error %v", ids, err)
		return
	}

	if _err := Log.AddLog(splashModel.ActionLogBusiness, username, uid, ids[0], "DeleteSelectConfigs", ids); _err != nil {
		log.Error("Service: UpdateSelectConfigSort AddLog error(%v)", _err)
	}

	return
}

// UpdateSelectConfigAuditState 更新自选配置审核状态
func (s *Service) UpdateSelectConfigAuditState(id int64, auditState api.SplashScreenConfigAuditStatus_Enum, username string, uid int64) (err error) {
	if id == 0 {
		return xecode.RequestErr
	}

	if exist, _err := s.CheckSelectConfigExist(id); _err != nil {
		log.Error("Service: UpdateSelectConfigAuditState CheckSelectConfigExist %d error %v", id, _err)
		return _err
	} else if !exist {
		err = xecode.Error(xecode.RequestErr, "配置不存在")
		return
	}

	if err = s.dao.UpdateSelectConfigAuditState(id, auditState, username); err != nil {
		log.Error("Service: UpdateSelectConfigAuditState  UpdateSelectConfigAuditState (%d, %d) error(%v)", id, auditState, err)
		return
	}

	// 手动下线的，当前配置的etime设置为当前时间
	if auditState == splashModel.AuditStateOffline || auditState == splashModel.AuditStateCancel {
		if err = s.dao.UpdateConfig(&splashModel.SplashScreenConfig{
			ID:    id,
			ETime: xtime.Time(time.Now().Unix()),
		}, username); err != nil {
			log.Error("UpdateAuditState s.dao.UpdateConfig error(%v)", err)
			return
		}
	}

	obj := map[string]interface{}{
		"value": auditState,
	}
	if err = Log.AddLog(splashModel.ActionLogBusiness, username, uid, id, "UpdateSelectConfigAuditState", obj); err != nil {
		log.Error("UpdateAuditState AddLog error(%v)", err)
		return
	}
	return
}

// CheckSelectConfigExist 检查自选配置是否存在
func (s *Service) CheckSelectConfigExist(id int64) (exists bool, err error) {
	if config, _err := s.dao.GetSelectConfig(id); _err != nil {
		log.Error("Service: checkSelectConfigExist %d error %v", id, _err)
		return false, _err
	} else if config == nil || config.ID == 0 {
		return false, nil
	}
	return true, nil
}
