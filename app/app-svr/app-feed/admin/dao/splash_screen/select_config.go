package splash_screen

import (
	"encoding/json"
	"fmt"
	"github.com/jinzhu/gorm"
	"github.com/pkg/errors"
	"strings"
	"time"

	xecode "go-common/library/ecode"
	xtime "go-common/library/time"

	"go-gateway/app/app-svr/app-feed/admin/api"
	splashModel "go-gateway/app/app-svr/app-feed/admin/model/splash_screen"
	"go-gateway/app/app-svr/app-feed/ecode"
)

const zeroTime = "2000-01-01 00:00:00"

/**
 * -------------- 自选配置分类  --------------
 */

// SaveCategories 保存所有分类
func (d *Dao) SaveCategories(categories []*splashModel.Category, username string) (res []*splashModel.Category, err error) {
	var (
		toSaveCategoryIDs         = make([]int64, 0)
		toSaveCategoriesMap       = make(map[string]*splashModel.Category)
		existingCategories        = make([]*splashModel.Category, 0)
		existingCategoriesIDMap   = make(map[int64]*splashModel.Category)
		existingCategoriesNameMap = make(map[string]*splashModel.Category)
		tx                        = d.DB.Begin()
	)
	defer func() {
		if err != nil {
			if _err := tx.Rollback().Error; _err != nil {
				err = errors.Wrap(_err, "tx.Rollback error")
			}
			return
		}
		if err = tx.Commit().Error; err != nil {
			err = errors.Wrap(err, "tx.Commit error")
		}
	}()

	// 检查并处理待保存id array
	for _, category := range categories {
		if category.Name == "" {
			return nil, xecode.Error(xecode.RequestErr, "分类名为空")
		}
		_category := category
		toSaveCategoryIDs = append(toSaveCategoryIDs, _category.ID)
		if _, exists := toSaveCategoriesMap[_category.Name]; exists {
			return nil, ecode.SplashScreenCategoryExists
		} else {
			toSaveCategoriesMap[_category.Name] = _category
		}
	}

	// 删除分类
	if err = d.DeleteCategoriesByIDsNotIn(toSaveCategoryIDs, tx); err != nil {
		err = errors.Wrap(err, "Delete unused categories error")
		return
	}
	// 删除分类配置关系
	if err = d.DeleteCategoriesConfigRelsByIDsNotIn(toSaveCategoryIDs, tx); err != nil {
		err = errors.Wrap(err, "Delete unused categories error")
		return
	}

	// 查询所有生效分类
	if err = tx.Model(splashModel.Category{}).Where("is_deleted = ?", splashModel.NotDeleted).Find(&existingCategories).Error; err != nil {
		err = errors.Wrap(err, "Find existing categories error")
		return
	}
	for _, category := range existingCategories {
		_category := category
		existingCategoriesIDMap[_category.ID] = _category
		existingCategoriesNameMap[_category.Name] = _category
	}

	// 更新及保存分类
	if res, err = d.processAndSaveCategories(categories, existingCategoriesIDMap, existingCategoriesNameMap, username, tx); err != nil {
		err = errors.Wrap(err, "processAndSaveCategories")
	}

	return
}

// processAndSaveCategories 保存预处理过后的所有分类
func (d *Dao) processAndSaveCategories(categories []*splashModel.Category, existingCategoriesIDMap map[int64]*splashModel.Category, existingCategoriesNameMap map[string]*splashModel.Category, username string, tx *gorm.DB) (res []*splashModel.Category, err error) {
	if len(categories) == 0 {
		return
	}

	res = make([]*splashModel.Category, 0, len(categories))
	for i, toSaveCategory := range categories {
		// 若id大于0且存在，则更新
		if toSaveCategory.ID != 0 {
			if existingCategory, exists := existingCategoriesIDMap[toSaveCategory.ID]; exists {
				name := toSaveCategory.Name
				toSaveCategory = existingCategory
				toSaveCategory.Name = name
				toSaveCategory.Sort = int64(len(categories) - i)
				toSaveCategory.MTime = xtime.Time(time.Now().Unix())
				toSaveCategory.MUser = username
				if err = tx.Model(splashModel.Category{}).Where("id = ?", toSaveCategory.ID).Updates(toSaveCategory).Error; err != nil {
					err = errors.Wrap(err, fmt.Sprintf("Update category %d error", toSaveCategory.ID))
					return
				}
			} else {
				name := toSaveCategory.Name
				toSaveCategory.ID = 0
				toSaveCategory.Name = name
				toSaveCategory.Sort = int64(len(categories) - i)
				toSaveCategory.CTime = xtime.Time(time.Now().Unix())
				toSaveCategory.CUser = username
				if err = tx.Model(splashModel.Category{}).Create(toSaveCategory).Error; err != nil {
					err = errors.Wrap(err, fmt.Sprintf("Create category %s error", toSaveCategory.Name))
					return
				}
			}
		} else if existingCategory, exists := existingCategoriesNameMap[toSaveCategory.Name]; exists {
			// 若同名分类已存在，则更新
			toSaveCategory = existingCategory
			toSaveCategory.Sort = int64(len(categories) - i)
			toSaveCategory.MTime = xtime.Time(time.Now().Unix())
			toSaveCategory.MUser = username
			if err = tx.Model(splashModel.Category{}).Where("id = ?", toSaveCategory.ID).Updates(toSaveCategory).Error; err != nil {
				err = errors.Wrap(err, fmt.Sprintf("Update category %d error", toSaveCategory.ID))
				return
			}
		} else {
			// 否则新建
			toSaveCategory.Sort = int64(len(categories) - i)
			toSaveCategory.CTime = xtime.Time(time.Now().Unix())
			toSaveCategory.CUser = username
			if err = tx.Model(splashModel.Category{}).Create(toSaveCategory).Error; err != nil {
				err = errors.Wrap(err, fmt.Sprintf("Create category %s error", toSaveCategory.Name))
				return
			}
		}
		res = append(res, toSaveCategory)
	}
	return
}

// GetAllCategories 获取所有自选分类
func (d *Dao) GetAllCategories() (res []*splashModel.Category, err error) {
	res = make([]*splashModel.Category, 0)
	err = d.DB.Model(splashModel.Category{}).Where("is_deleted = ?", splashModel.NotDeleted).Order("sort DESC").Find(&res).Error
	return
}

// DeleteCategoriesByIDsNotIn 删除没有指定的所有分类
func (d *Dao) DeleteCategoriesByIDsNotIn(ids []int64, tx *gorm.DB) (err error) {
	if tx == nil {
		return xecode.Error(xecode.ServerErr, "transaction nil")
	}
	deleteQuery := tx.Model(splashModel.Category{}).Where("is_deleted = ?", splashModel.NotDeleted)
	if len(ids) > 0 {
		deleteQuery = deleteQuery.Where("id NOT IN (?)", ids)
	}
	if err = deleteQuery.UpdateColumn("is_deleted", splashModel.IsDeleted).Error; err != nil {
		return
	}
	return
}

// DeleteCategoriesConfigRelsByIDsNotIn 删除没有指定的所有分类与配置的关系
func (d *Dao) DeleteCategoriesConfigRelsByIDsNotIn(configIDs []int64, tx *gorm.DB) (err error) {
	if tx == nil {
		return xecode.Error(xecode.ServerErr, "transaction nil")
	}
	deleteRelQuery := tx.Model(splashModel.SelectConfigCategoryRel{}).Where("is_deleted = ?", splashModel.NotDeleted)
	if len(configIDs) > 0 {
		deleteRelQuery = deleteRelQuery.Where("config_id NOT IN (?)", configIDs)
	}
	if err = deleteRelQuery.UpdateColumn("is_deleted", splashModel.IsDeleted).Error; err != nil {
		return
	}
	return
}

// GetAllCategoriesWithConfigsCount 获取所有分类及其配置数量
func (d *Dao) GetAllCategoriesWithConfigsCount(state api.SplashScreenConfigState_Enum) (res []*splashModel.CategoryWithConfigCount, err error) {
	res = make([]*splashModel.CategoryWithConfigCount, 0)
	var joinConfigSQL = "LEFT JOIN splash_screen_select_config ON splash_screen_select_config_category_rel.config_id=splash_screen_select_config.id AND splash_screen_select_config.is_deleted = ?"
	var joinConfigParams = []interface{}{splashModel.NotDeleted}
	query := d.DB.Model(splashModel.Category{}).
		Joins("LEFT JOIN splash_screen_select_config_category_rel ON splash_screen_select_category.id = splash_screen_select_config_category_rel.category_id AND splash_screen_select_config_category_rel.is_deleted = ?", splashModel.NotDeleted)
	if state >= 0 {
		now := time.Now()
		switch state {
		case api.SplashScreenConfigState_REJECTED, api.SplashScreenConfigState_DEPRECATED:
			joinConfigSQL += " AND (splash_screen_select_config.audit_state = ? OR (splash_screen_select_config.etime < ? AND splash_screen_select_config.etime > ?))"
			joinConfigParams = append(joinConfigParams, api.SplashScreenConfigAuditStatus_OFFLINE, now, zeroTime)
		case api.SplashScreenConfigState_NOTPASSED:
			joinConfigSQL += " AND (splash_screen_select_config.audit_state = ? AND splash_screen_select_config.stime < ? AND (splash_screen_select_config.etime < ? OR splash_screen_select_config.etime > ?))"
			joinConfigParams = append(joinConfigParams, api.SplashScreenConfigAuditStatus_NOTPASSED, now, zeroTime, now)
		case api.SplashScreenConfigState_WAITINGONLINE:
			joinConfigSQL += " AND (splash_screen_select_config.audit_state = ? AND splash_screen_select_config.stime > ?)"
			joinConfigParams = append(joinConfigParams, api.SplashScreenConfigAuditStatus_PASSED, now)
		case api.SplashScreenConfigState_ONLINE:
			joinConfigSQL += " AND (splash_screen_select_config.audit_state = ? AND splash_screen_select_config.stime <= ? AND (splash_screen_select_config.etime < ? OR splash_screen_select_config.etime > ?))"
			joinConfigParams = append(joinConfigParams, api.SplashScreenConfigAuditStatus_PASSED, now, zeroTime, now)
		}
	}
	query = query.Joins(joinConfigSQL, joinConfigParams...).
		Where("splash_screen_select_category.is_deleted = ?", splashModel.NotDeleted)

	err = query.Group("splash_screen_select_category.id").
		Order("splash_screen_select_category.sort DESC").
		Select("splash_screen_select_category.*, COUNT(splash_screen_select_config.id) as config_count").
		Find(&res).Error

	return
}

// GetCategorySelectConfigCount 获取对应分类配置计数。0为全部分类
func (d *Dao) GetCategorySelectConfigCount(categoryID int64, state api.SplashScreenConfigState_Enum) (count int32, err error) {
	query := d.DB.Model(splashModel.SelectConfig{})
	if categoryID > 0 || state >= 0 {
		query = query.Joins("LEFT JOIN splash_screen_select_config_category_rel ON splash_screen_select_config.id = splash_screen_select_config_category_rel.config_id AND splash_screen_select_config_category_rel.is_deleted = ?", splashModel.NotDeleted).
			Group("splash_screen_select_config.id")
	}
	if categoryID > 0 {
		query = query.Where("splash_screen_select_config_category_rel.category_id = ?", categoryID)
	}
	if state >= 0 {
		now := time.Now()
		switch state {
		case api.SplashScreenConfigState_REJECTED, api.SplashScreenConfigState_DEPRECATED:
			query = query.Where("splash_screen_select_config.audit_state = ? OR (splash_screen_select_config.etime < ? AND splash_screen_select_config.etime > ?)", api.SplashScreenConfigAuditStatus_OFFLINE, now, zeroTime)
		case api.SplashScreenConfigState_NOTPASSED:
			query = query.Where("splash_screen_select_config.audit_state = ? AND splash_screen_select_config.stime < ? AND (splash_screen_select_config.etime < ? OR splash_screen_select_config.etime > ?)", api.SplashScreenConfigAuditStatus_NOTPASSED, now, zeroTime, now)
		case api.SplashScreenConfigState_WAITINGONLINE:
			query = query.Where("splash_screen_select_config.audit_state = ? AND splash_screen_select_config.stime > ?", api.SplashScreenConfigAuditStatus_PASSED, now)
		case api.SplashScreenConfigState_ONLINE:
			query = query.Where("splash_screen_select_config.audit_state = ? AND splash_screen_select_config.stime <= ? AND (splash_screen_select_config.etime < ? OR splash_screen_select_config.etime > ?)", api.SplashScreenConfigAuditStatus_PASSED, now, zeroTime, now)
		}
	}
	if err = query.
		Where("splash_screen_select_config.is_deleted = ?", splashModel.NotDeleted).
		Count(&count).Error; err != nil {
		if err.Error() == "sql: no rows in result set" {
			err = nil
			count = 0
		}
	}

	return
}

/**
 * -------------- 自选配置 --------------
 */

// GetSelectConfigOnline 根据新自选配置列表生成老的自选配置结构
func (d *Dao) GetSelectConfigOnline() (configOnline *splashModel.SplashScreenConfig, err error) {
	var (
		notDeprecatedSelectConfigs = make([]*splashModel.SelectConfig, 0)
		configDetails              = make([]*splashModel.ConfigDetail, 0)
	)
	nowTime := time.Now()
	// 根据自选配置列表生成对应的通用配置体
	// 若无有效配置则返回空
	// 找出 stime < current_time，没有 etime 或者 etime > current_time，审核通过，没被删除的，按照 stime 顺序的第一个作为起始时间
	if err = d.DB.Model(splashModel.SelectConfig{}).
		Joins("JOIN splash_screen_images ON splash_screen_select_config.image_id = splash_screen_images.id AND splash_screen_images.is_deleted = ?", splashModel.NotDeleted).
		Where("splash_screen_select_config.stime <= ?", nowTime).
		Where("splash_screen_select_config.etime <= ? OR splash_screen_select_config.etime > ?", xtime.Time(0), nowTime).
		Where("splash_screen_select_config.audit_state = ?", splashModel.AuditStatePass).
		Where("splash_screen_select_config.is_deleted = ?", splashModel.NotDeleted).
		Select("splash_screen_select_config.*").
		Order("splash_screen_select_config.stime ASC").Find(&notDeprecatedSelectConfigs).Error; err != nil {
		err = errors.Wrap(err, "GetSelectConfigOnline dao First error")
		return
	}
	if len(notDeprecatedSelectConfigs) == 0 {
		return
	}
	configOnline = &splashModel.SplashScreenConfig{
		ID:         1,
		STime:      notDeprecatedSelectConfigs[0].STime,
		ETime:      xtime.Time(time.Now().AddDate(1, 0, 0).Unix()),
		State:      api.SplashScreenConfigState_ONLINE,
		AuditState: api.SplashScreenConfigAuditStatus_PASSED,
		ShowMode:   splashModel.ShowModeSelect,
	}
	for i, config := range notDeprecatedSelectConfigs {
		_config := config
		configDetails = append(configDetails, &splashModel.ConfigDetail{
			Position: i,
			ImgId:    _config.ImageID,
			Sort:     _config.Sort,
		})
	}
	if configJSON, _err := json.Marshal(configDetails); _err != nil {
		err = errors.Wrap(_err, fmt.Sprintf("Marshal %+v", configDetails))
		return
	} else {
		configOnline.ConfigJson = string(configJSON)
	}

	return
}

// GetConfigInDays 根据 show_mode，获取即将生效的n天内的配置
func (d *Dao) GetSelectConfigsInDays(nDays int) (configList []*splashModel.SplashScreenConfig, err error) {
	if nDays == 0 {
		return
	}

	var (
		passedAndToBeOnlineSelectConfigs = make([]*splashModel.SelectConfig, 0)
	)
	configList = make([]*splashModel.SplashScreenConfig, 0)

	nowTime := time.Now()
	nDaysAfter := nowTime.AddDate(0, 0, nDays)
	// 找出 stime 在 nDays 内的，审核通过，没被删除的，按照 stime, etime 顺序
	if err = d.DB.Model(&splashModel.SelectConfig{}).
		Joins("JOIN splash_screen_images ON splash_screen_select_config.image_id = splash_screen_images.id AND splash_screen_images.is_deleted = ?", splashModel.NotDeleted).
		Where("splash_screen_select_config.stime > ?", nowTime).
		Where("splash_screen_select_config.stime < ?", nDaysAfter).
		Where("splash_screen_select_config.audit_state = ?", api.SplashScreenConfigAuditStatus_PASSED).
		Where("splash_screen_select_config.is_deleted = ?", splashModel.NotDeleted).
		Select("splash_screen_select_config.*").
		Order("splash_screen_select_config.stime ASC, splash_screen_select_config.etime ASC").
		Find(&passedAndToBeOnlineSelectConfigs).Error; err != nil {
		if err == gorm.ErrRecordNotFound || err == xecode.NothingFound {
			err = nil
		} else {
			err = errors.Wrap(err, "GetSelectConfigsInDays dao Find error")
		}
		return
	}
	for i, config := range passedAndToBeOnlineSelectConfigs {
		_config := config
		configDetail := &splashModel.ConfigDetail{
			Position: i,
			ImgId:    _config.ImageID,
			Sort:     _config.Sort,
		}
		if configJSON, _err := json.Marshal([]*splashModel.ConfigDetail{configDetail}); _err != nil {
			err = errors.Wrap(_err, fmt.Sprintf("Marshal %+v", configDetail))
			return
		} else {
			configList = append(configList, &splashModel.SplashScreenConfig{
				ID:         int64(i),
				STime:      _config.STime,
				ETime:      _config.ETime,
				State:      api.SplashScreenConfigState_WAITINGONLINE,
				AuditState: api.SplashScreenConfigAuditStatus_PASSED,
				ShowMode:   splashModel.ShowModeSelect,
				ConfigJson: string(configJSON),
			})
		}
	}

	return
}

var getSelectConfigListSortableColumns = map[string]int{
	"id":   1,
	"sort": 1,
}

// GetSelectConfigList 分页获取所有自选配置
func (d *Dao) GetSelectConfigList(sorting string) (configList []*splashModel.SelectConfig, count int32, err error) {
	if sorting == "" {
		sorting = "sort DESC"
	} else {
		sortingSlice := strings.Split(sorting, " ")
		//nolint:gomnd
		if len(sortingSlice) > 2 {
			err = xecode.Error(xecode.RequestErr, "sorting错误")
			return
		}
		if _, exists := getSelectConfigListSortableColumns[sortingSlice[0]]; !exists {
			err = xecode.Error(xecode.RequestErr, "sorting错误")
			return
		}
		//nolint:gomnd
		if len(sortingSlice) == 2 {
			if strings.ToLower(sortingSlice[1]) != "asc" && strings.ToLower(sortingSlice[1]) != "desc" {
				err = xecode.Error(xecode.RequestErr, "sorting错误")
				return
			}
		}
	}

	query := d.DB.Model(&splashModel.SelectConfig{}).
		Where("is_deleted = ?", splashModel.NotDeleted)

	if err = query.Count(&count).Error; err != nil {
		err = errors.Wrap(err, "query.Count Error")
		return
	}

	if err = query.Order(sorting).
		Find(&configList).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			err = nil
		}
	}

	// 分类id
	var (
		configIDs         = make([]int64, 0, len(configList))
		configCategoryMap = make(map[int64][]int64)
	)
	for _, config := range configList {
		configIDs = append(configIDs, config.ID)
	}
	relList := make([]*splashModel.SelectConfigCategoryRel, 0)
	if err = d.DB.Model(splashModel.SelectConfigCategoryRel{}).Where("config_id IN (?) AND is_deleted = ?", configIDs, splashModel.NotDeleted).Find(&relList).Error; err != nil {
		return
	}
	for _, rel := range relList {
		if _, exists := configCategoryMap[rel.ConfigID]; !exists {
			configCategoryMap[rel.ConfigID] = make([]int64, 0)
		}
		configCategoryMap[rel.ConfigID] = append(configCategoryMap[rel.ConfigID], rel.CategoryID)
	}
	// 处理分类和state
	for _, config := range configList {
		if _, exists := configCategoryMap[config.ID]; exists {
			config.CategoryIDs = configCategoryMap[config.ID]
		} else {
			config.CategoryIDs = make([]int64, 0)
		}

		config.State = d.GetStateForSelectConfig(config)
	}

	return
}

// GetSelectConfigListByPage 分页获取所有自选配置
func (d *Dao) GetSelectConfigsByIDs(ids []int64) (configList []*splashModel.SelectConfig, err error) {
	if len(ids) == 0 {
		return
	}

	configList = make([]*splashModel.SelectConfig, 0)

	query := d.DB.Model(&splashModel.SelectConfig{}).
		Where("id in (?) AND is_deleted = ?", ids, splashModel.NotDeleted)

	if err = query.Order("sort desc").
		Find(&configList).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			err = nil
		}
	}

	// 分类id
	var (
		configIDs         = make([]int64, 0, len(configList))
		configCategoryMap = make(map[int64][]int64)
	)
	for _, config := range configList {
		_config := config
		configIDs = append(configIDs, _config.ID)
	}
	relList := make([]*splashModel.SelectConfigCategoryRel, 0)
	if err = d.DB.Model(splashModel.SelectConfigCategoryRel{}).Where("id IN (?)", configIDs).Find(&relList).Error; err != nil {
		return
	}
	for _, rel := range relList {
		_rel := rel
		if _, exists := configCategoryMap[_rel.ConfigID]; !exists {
			configCategoryMap[_rel.ConfigID] = make([]int64, 0)
		}
		configCategoryMap[_rel.ConfigID] = append(configCategoryMap[_rel.ConfigID], _rel.CategoryID)
	}
	// 处理分类和state
	for _, config := range configList {
		if _, exists := configCategoryMap[config.ID]; exists {
			config.CategoryIDs = configCategoryMap[config.ID]
		} else {
			config.CategoryIDs = make([]int64, 0)
		}

		config.State = d.GetStateForSelectConfig(config)
	}

	return
}

// GetStateForSelectConfig 计算自选配置的状态
func (d *Dao) GetStateForSelectConfig(config *splashModel.SelectConfig) (state api.SplashScreenConfigState_Enum) {
	if config == nil || config.ID == 0 {
		return
	}
	now := time.Now()
	switch config.AuditState {
	case api.SplashScreenConfigAuditStatus_NOTPASSED:
		// 未通过审核
		if config.ETime > 0 && config.ETime.Time().Before(now) {
			// 已失效
			return api.SplashScreenConfigState_DEPRECATED
		} else {
			// 未通过
			return api.SplashScreenConfigState_NOTPASSED
		}
	case api.SplashScreenConfigAuditStatus_OFFLINE:
		// 手动下线
		return api.SplashScreenConfigState_REJECTED
	case api.SplashScreenConfigAuditStatus_PASSED:
		if config.STime.Time().Before(now) {
			// 本该生效的
			if config.ETime > 0 {
				// 非永久生效
				if config.ETime.Time().After(now) {
					// 生效中
					return api.SplashScreenConfigState_ONLINE
				} else {
					// 过期，已失效
					return api.SplashScreenConfigState_DEPRECATED
				}
			} else {
				// 永久有效
				return api.SplashScreenConfigState_ONLINE
			}
		} else {
			// 待生效
			return api.SplashScreenConfigState_WAITINGONLINE
		}
	}

	return
}

// GetSelectConfig 根据id查询自选配置
func (d *Dao) GetSelectConfig(id int64) (res *splashModel.SelectConfig, err error) {
	if id == 0 {
		return
	}

	res = &splashModel.SelectConfig{}
	if err = d.DB.Model(&splashModel.SelectConfig{}).
		Where("id = ? AND is_deleted = ?", id, splashModel.NotDeleted).
		First(&res).Error; err != nil {
		if err == gorm.ErrRecordNotFound || err == xecode.NothingFound {
			err = nil
		}
	}
	return
}

// GetSelectConfigByImageID 根据物料id查询自选闪屏配置
func (d *Dao) GetSelectConfigByImageID(imageID int64) (res *splashModel.SelectConfig, err error) {
	if imageID == 0 {
		return
	}

	res = &splashModel.SelectConfig{}
	if err = d.DB.Model(&splashModel.SelectConfig{}).
		Where("image_id = ? AND is_deleted = ?", imageID, splashModel.NotDeleted).
		First(res).Error; err != nil {
		if err == gorm.ErrRecordNotFound || err == xecode.NothingFound {
			err = nil
		}
	}
	return
}

// AddSelectConfig 新建自选配置
func (d *Dao) AddSelectConfig(toAddConfig *splashModel.SelectConfig, username string) (res *splashModel.SelectConfig, err error) {
	if toAddConfig == nil {
		return
	}

	toAddConfig.CTime = xtime.Time(time.Now().Unix())
	toAddConfig.CUser = username
	if err = d.DB.Model(splashModel.SelectConfig{}).Create(toAddConfig).Error; err != nil {
		return
	}
	return toAddConfig, nil
}

// UpdateSelectConfig 更新自选配置
func (d *Dao) UpdateSelectConfig(id int64, updateMap map[string]interface{}, username string) (res *splashModel.SelectConfig, err error) {
	if id == 0 || len(updateMap) == 0 {
		return nil, xecode.RequestErr
	}

	updateMap["m_user"] = username
	updateMap["mtime"] = xtime.Time(time.Now().Unix())

	query := d.DB.Model(splashModel.SelectConfig{}).Where("id = ?", id)
	if err = query.UpdateColumns(updateMap).Error; err != nil {
		return nil, errors.Wrap(err, "UpdateColumns")
	}
	res = &splashModel.SelectConfig{}
	err = query.Find(res).Error
	return
}

// UpdateSelectConfigAuditState 更新自选配置审核状态
func (d *Dao) UpdateSelectConfigAuditState(id int64, auditState api.SplashScreenConfigAuditStatus_Enum, username string) (err error) {
	updateMap := map[string]interface{}{
		"m_user":      username,
		"mtime":       xtime.Time(time.Now().Unix()),
		"audit_state": auditState,
	}
	if err = d.DB.Model(splashModel.SelectConfig{}).
		Where("id = ?", id).
		UpdateColumns(updateMap).Error; err != nil {
		return
	}
	return
}

// DeleteSelectConfig 删除自选配置
func (d *Dao) DeleteSelectConfig(id int64, username string) (err error) {
	if id == 0 {
		return
	}
	updateMap := map[string]interface{}{
		"is_deleted": splashModel.IsDeleted,
		"m_user":     username,
		"mtime":      xtime.Time(time.Now().Unix()),
	}
	err = d.DB.Model(splashModel.SelectConfig{}).Where("id = ? AND is_deleted = ?", id, splashModel.NotDeleted).UpdateColumns(updateMap).Error

	return
}

// DeleteSelectConfig 删除自选配置
func (d *Dao) DeleteSelectConfigs(ids []int64, username string) (err error) {
	if len(ids) == 0 {
		return
	}
	updateMap := map[string]interface{}{
		"is_deleted": splashModel.IsDeleted,
		"m_user":     username,
		"mtime":      xtime.Time(time.Now().Unix()),
	}
	err = d.DB.Model(splashModel.SelectConfig{}).Where("id IN (?) AND is_deleted = ?", ids, splashModel.NotDeleted).UpdateColumns(updateMap).Error

	return
}

// GetNewSelectConfigSort 获取新的自选配置的sort值，自增
func (d *Dao) GetNewSelectConfigSort() (newSort int64, err error) {
	sortModel := &splashModel.SelectConfigSortSeq{}
	if err = d.DB.Model(splashModel.SelectConfigSortSeq{}).Create(sortModel).Error; err != nil {
		return
	} else {
		newSort = sortModel.ID
	}

	return
}

// updateSelectConfigSortBottomProcessOthersSQL 更新
const updateSelectConfigSortBottomProcessOthersSQL = "UPDATE splash_screen_select_config SET sort=sort+1 WHERE id != ? AND is_deleted = ?"

// UpdateSelectConfigSortBottom 将自选配置置底
func (d *Dao) UpdateSelectConfigSortBottom(id int64, username string) (err error) {
	if id == 0 {
		return
	}

	if _, err = d.GetNewSelectConfigSort(); err != nil {
		return errors.Wrap(err, "GetNewSelectConfigSort")
	}
	updateMap := map[string]interface{}{
		"sort":   1,
		"m_user": username,
		"mtime":  xtime.Time(time.Now().Unix()),
	}
	if err = d.DB.Model(splashModel.SelectConfig{}).Where("id = ?", id).UpdateColumns(updateMap).Error; err != nil {
		return
	}

	err = d.DB.Exec(updateSelectConfigSortBottomProcessOthersSQL, id, splashModel.NotDeleted).Error
	return
}

// DeleteSelectConfigCategoryRelsByConfig 根据自选配置id删除已有分类关系
func (d *Dao) DeleteSelectConfigCategoryRelsByConfig(configID int64, username string) (err error) {
	if configID == 0 {
		return xecode.RequestErr
	}

	updateMap := map[string]interface{}{
		"is_deleted": splashModel.IsDeleted,
		"m_user":     username,
		"mtime":      xtime.Time(time.Now().Unix()),
	}
	err = d.DB.Model(splashModel.SelectConfigCategoryRel{}).Where("config_id = ? AND is_deleted = ?", configID, splashModel.NotDeleted).UpdateColumns(updateMap).Error
	return
}

// AddSelectConfigCategoryRels 批量添加自选配置分类关系
func (d *Dao) AddSelectConfigCategoryRels(toAddRels []*splashModel.SelectConfigCategoryRel, username string) (res []*splashModel.SelectConfigCategoryRel, err error) {
	if len(toAddRels) == 0 {
		return
	}

	tx := d.DB.Begin()
	defer func() {
		if err != nil {
			err = tx.Rollback().Error
		} else {
			err = tx.Commit().Error
		}
		//nolint:gosimple
		return
	}()

	for _, rel := range toAddRels {
		rel.CUser = username
		rel.CTime = xtime.Time(time.Now().Unix())
		if err = tx.Model(splashModel.SelectConfigCategoryRel{}).Create(rel).Error; err != nil {
			return
		}
	}
	res = toAddRels

	return
}
