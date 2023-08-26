package splash_screen

import (
	"fmt"
	"github.com/jinzhu/gorm"
	"time"

	"go-common/library/log"
	xtime "go-common/library/time"

	splashModel "go-gateway/app/app-svr/app-feed/admin/model/splash_screen"
)

/**
物料
*/
// 获取所有没被删除的物料列表
func (d *Dao) GetImageList() (imageList []*splashModel.SplashScreenImage, err error) {
	if err = d.DB.Model(&splashModel.SplashScreenConfig{}).
		Where("is_deleted = ?", splashModel.NotDeleted).
		Order("ctime desc").Find(&imageList).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			err = nil
		} else {
			log.Error("getImageList dao Find error(%v)", err)
		}
	}
	// 分类id
	var (
		imageIDs         = make([]int64, 0, len(imageList))
		imageCategoryMap = make(map[int64][]int64)
	)
	for _, image := range imageList {
		imageIDs = append(imageIDs, image.ID)
	}
	relList := make([]*splashModel.SelectConfigCategoryRelWithImageID, 0)
	if err = d.DB.Model(splashModel.SelectConfigCategoryRel{}).
		Joins("JOIN splash_screen_select_config ON splash_screen_select_config.id=splash_screen_select_config_category_rel.config_id AND splash_screen_select_config_category_rel.is_deleted = ?", splashModel.NotDeleted).
		Where("splash_screen_select_config.image_id IN (?) AND splash_screen_select_config.is_deleted = ?", imageIDs, splashModel.NotDeleted).
		Select("splash_screen_select_config_category_rel.*, splash_screen_select_config.image_id").
		Find(&relList).Error; err != nil {
		return
	}
	for _, rel := range relList {
		if _, exists := imageCategoryMap[rel.ImageID]; !exists {
			imageCategoryMap[rel.ImageID] = make([]int64, 0)
		}
		imageCategoryMap[rel.ImageID] = append(imageCategoryMap[rel.ImageID], rel.CategoryID)
	}
	// 处理分类和state
	for i, image := range imageList {
		if _, exists := imageCategoryMap[image.ID]; exists {
			imageList[i].CategoryIDs = imageCategoryMap[image.ID]
		} else {
			imageList[i].CategoryIDs = make([]int64, 0)
		}
	}
	return
}

// 获取所有没被删除的物料列表
func (d *Dao) GetImageWithCategoriesList() (imageList []*splashModel.SplashScreenImage, err error) {
	if err = d.DB.Model(&splashModel.SplashScreenConfig{}).
		Where("is_deleted = ?", splashModel.NotDeleted).
		Order("ctime desc").Find(&imageList).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			err = nil
		} else {
			log.Error("getImageList dao Find error(%v)", err)
		}
	}
	return
}

// 根据id获取某一个物料详情
func (d *Dao) GetImageDetail(id int64) (imageDetail *splashModel.SplashScreenImage, err error) {
	imageDetail = &splashModel.SplashScreenImage{}
	if err = d.DB.Model(&splashModel.SplashScreenConfig{}).
		Where("is_deleted = ?", splashModel.NotDeleted).
		Order("ctime desc").First(imageDetail, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			err = nil
		} else {
			log.Error("getImageDetail dao First error(%v)", err)
		}
	}
	return
}

// 新添加一个物料
func (d *Dao) InsertImage(newImage *splashModel.SplashScreenImage, username string) (id int64, err error) {
	insertItem := newImage
	insertItem.CUser = username
	insertItem.MUser = username
	if err = d.DB.Model(&splashModel.SplashScreenImage{}).Create(insertItem).Error; err != nil {
		log.Error("insertImage dao Create error(%v)", err)
	}
	id = insertItem.ID
	return
}

// 删除某个物料
func (d *Dao) DeleteImage(id int64, username string) (err error) {
	if err = d.DB.Model(&splashModel.SplashScreenImage{}).
		Where("id = ?", id).
		Updates(&splashModel.SplashScreenImage{
			IsDeleted: splashModel.IsDeleted,
			MUser:     username,
		}).Error; err != nil {
		log.Error("deleteImage dao Update error(%v)", err)
	}
	return
}

// 更新某一个物料的详情信息
func (d *Dao) UpdateImage(updateImage *splashModel.SplashScreenImage, username string) (err error) {
	toUpdateImage := updateImage
	toUpdateImage.MUser = username
	toUpdateImage.MTime = xtime.Time(time.Now().Unix())
	updateQuery := d.DB.Model(&splashModel.SplashScreenImage{}).
		Where("id = ?", updateImage.ID)
	if err = updateQuery.Updates(toUpdateImage).Error; err != nil {
		log.Error("updateImage dao Update error(%v)", err)
	}
	if toUpdateImage.LogoShowFlag == 1 {
		updateHideFieldMap := make(map[string]interface{})
		updateHideFieldMap["logo_hide"] = splashModel.LogoShow
		if err = updateQuery.Updates(updateHideFieldMap).Error; err != nil {
			log.Error("updateImage dao Updates logo_hide field error(%v)", err)
		}
	}
	return
}

/*
配置列表
*/
// 根据 show_mode，获取所有列表
func (d *Dao) GetConfigListAll(showModeFilter []int, ps, pn int32) (configList []*splashModel.SplashScreenConfig, count int32, err error) {
	query := d.DB.Model(&splashModel.SplashScreenConfig{}).
		Where("show_mode in (?)", showModeFilter).
		Where("is_deleted = ?", splashModel.NotDeleted)

	if err = query.Count(&count).Error; err != nil {
		log.Error("GetConfigListAll dao Count error(%v)", err)
		return
	}

	if err = query.Order("ctime desc").
		Offset(ps * (pn - 1)).Limit(ps).
		Find(&configList).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			err = nil
		} else {
			log.Error("GetConfigListAll dao Find error(%v)", err)
		}
	}
	return
}

// 根据 show_mode 和 stime，获取符合条件配置个数，也就是同样类型配置下，stime 冲突的个数
func (d *Dao) GetConfigConflictCnt(showModeFilter []int, stime xtime.Time, id int64) (cnt int, err error) {
	query := d.DB.Model(&splashModel.SplashScreenConfig{}).
		Where("stime = ?", stime).
		Where("show_mode in (?)", showModeFilter).
		Where("is_deleted = ?", splashModel.NotDeleted)

	if id != 0 {
		query = query.Where("id != ?", id)
	}

	if err = query.Count(&cnt).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			err = nil
		} else {
			log.Error("GetConfigConflictCnt dao Find error(%v)", err)
		}
	}
	return
}

// 根据 show_mode，获取正在生效和上一次生效的配置
func (d *Dao) GetConfigOnline(showModeFilter []int) (configOnline, lastConfig *splashModel.SplashScreenConfig, err error) {
	var (
		configList []*splashModel.SplashScreenConfig
	)
	nowTime := time.Now().Format("2006-01-02 15:04:05")
	// 找出 in showModeFilter，stime < current_time，没有 etime 或者 etime > current_time，审核通过，没被删除的，按照 stime 降序的第一个就是当前配置，第二个是上一个配置
	if err = d.DB.Model(&splashModel.SplashScreenConfig{}).
		Where("show_mode in (?)", showModeFilter).
		Where("stime <= ?", nowTime).
		Where("etime = ? OR etime > ?", xtime.Time(0), nowTime).
		Where("audit_state = ?", splashModel.AuditStatePass).
		Where("is_deleted = ?", splashModel.NotDeleted).
		Order("stime desc").Limit(2).Find(&configList).Error; err != nil {
		log.Error("GetConfigListOnline dao First error(%v)", err)
	}

	//nolint:gomnd
	if len(configList) == 2 {
		// 当前和上一条配置都有
		lastConfig = configList[1]
		configOnline = configList[0]
	} else if len(configList) == 1 {
		// 只有一条配置
		configOnline = configList[0]
	}
	// 其他情况，没有任何配置

	return
}

// GetConfigInDays 根据 show_mode，获取即将生效的n天内的配置
func (d *Dao) GetConfigInDays(showModeFilter []int, nDays int) (configList []*splashModel.SplashScreenConfig, err error) {
	nowTime := time.Now()
	fmt.Println(nowTime.Format("2006-01-02 15:04:05"))
	nDaysAfter := nowTime.AddDate(0, 0, nDays)
	fmt.Println(nDaysAfter.Format("2006-01-02 15:04:05"))
	// 找出 in showModeFilter，stime 在 nDays 内的，审核通过，没被删除的，按照 stime, etime 顺序
	query := d.DB.Model(&splashModel.SplashScreenConfig{}).
		Where("show_mode in (?)", showModeFilter).
		Where("stime > ?", nowTime).
		Where("stime < ?", nDaysAfter).
		Where("audit_state = ?", splashModel.AuditStatePass).
		Where("is_deleted = ?", splashModel.NotDeleted).
		Order("stime asc, etime asc")

	if err := query.Find(&configList).Error; err != nil {
		log.Error("GetConfigInDays dao First error(%v)", err)
	}

	return
}

// 根据id获取某一个配置详情
func (d *Dao) GetConfigDetail(id int64) (configDetail *splashModel.SplashScreenConfig, err error) {
	configDetail = &splashModel.SplashScreenConfig{}
	if err = d.DB.Model(&splashModel.SplashScreenConfig{}).
		Where("is_deleted = ?", splashModel.NotDeleted).
		First(&configDetail, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			err = nil
		} else {
			log.Error("getConfigDetail dao First error(%v)", err)
		}
	}
	return
}

// 添加一个配置
func (d *Dao) InsertConfig(newConfig *splashModel.SplashScreenConfig, username string) (id int64, err error) {
	insertItem := &splashModel.SplashScreenConfig{
		STime:          newConfig.STime,
		ETime:          newConfig.ETime,
		IsImmediately:  newConfig.IsImmediately,
		ShowMode:       newConfig.ShowMode,
		ConfigJson:     newConfig.ConfigJson,
		ForceShowTimes: newConfig.ForceShowTimes,
		CUser:          username,
		MUser:          username,
	}
	if err = d.DB.Model(&splashModel.SplashScreenConfig{}).Create(insertItem).Error; err != nil {
		log.Error("insertConfig dao Create error(%v)", err)
	}
	id = insertItem.ID
	return
}

// 更新某一个配置的详情信息
func (d *Dao) UpdateConfig(updateConfig *splashModel.SplashScreenConfig, username string) (err error) {
	if err = d.DB.Model(&splashModel.SplashScreenConfig{}).
		Where("id = ?", updateConfig.ID).
		Updates(&splashModel.SplashScreenConfig{
			IsImmediately:  updateConfig.IsImmediately,
			STime:          updateConfig.STime,
			ETime:          updateConfig.ETime,
			ShowMode:       updateConfig.ShowMode,
			ConfigJson:     updateConfig.ConfigJson,
			ForceShowTimes: updateConfig.ForceShowTimes,
			MUser:          username,
		}).Error; err != nil {
		log.Error("updateConfig dao Update error(%v)", err)
	}
	return
}

// 更新可能为0值的字段
func (d *Dao) UpdateConfigMap(id int64, updateConfigMap map[string]interface{}, username string) (err error) {
	updateConfigMap["m_user"] = username
	if err = d.DB.Model(&splashModel.SplashScreenConfig{}).
		Where("id = ?", id).
		Updates(updateConfigMap).Error; err != nil {
		log.Error("updateConfig dao Update error(%v)", err)
	}
	return
}

// 更新审核状态
func (d *Dao) UpdateConfigAuditState(id int64, auditState int, username string) (err error) {
	updateMap := map[string]interface{}{
		"m_user":      username,
		"audit_state": auditState,
	}
	if err = d.DB.Model(&splashModel.SplashScreenConfig{}).
		Where("id = ?", id).
		Updates(updateMap).Error; err != nil {
		log.Error("updateConfigAuditState dao Update error(%v)", err)
		return
	}
	return
}

// 失效配置更新后，重新设置配置的etime=0和audit_state=0
func (d *Dao) ResetConfigETimeAndAuditState(id int64, etime xtime.Time) (err error) {
	updateMap := map[string]interface{}{
		"audit_state": 0,
		"etime":       etime,
	}
	if err = d.DB.Model(&splashModel.SplashScreenConfig{}).
		Where("id = ?", id).
		Updates(updateMap).Error; err != nil {
		log.Error("ResetConfigETimeAndAuditState dao Update error(%v)", err)
	}
	return
}
