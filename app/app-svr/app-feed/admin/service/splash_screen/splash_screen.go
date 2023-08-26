package splash_screen

import (
	"encoding/json"
	"sort"
	"sync"
	"time"

	"go-common/library/ecode"
	"go-common/library/log"
	"go-common/library/sync/errgroup"
	xtime "go-common/library/time"

	"go-gateway/app/app-svr/app-feed/admin/conf"
	splashModel "go-gateway/app/app-svr/app-feed/admin/model/splash_screen"
	Log "go-gateway/app/app-svr/app-feed/admin/util"
	feedEcode "go-gateway/app/app-svr/app-feed/ecode"
)

const (
	// PrepareConfigDaysNum 获取预生效配置的天数（从今天开始往后）
	PrepareConfigDaysNum = 14
)

// 获取给网关用的结束时间，如果没结束时间，就设置成当前时间一年后
func getETime(config *splashModel.SplashScreenConfig) (etime xtime.Time) {
	if config.ETime != 0 {
		etime = config.ETime
		return
	}
	// 如果没有etime，就设置为当前时间一年后
	etime = xtime.Time(time.Now().AddDate(1, 0, 0).Unix())
	return
}

/**
 * 获取给网关用的预生效配置的结束时间.
 * 原config list按stime asc排序
 * 如果有etime且（小于下个config的stime或没有下个config），则不变
 * 如果有etime且大于等于下个config的stime，则设置为下个config的stime-1s
 * 如果没有etime且有下个config，则设置为下个config的stime-1s
 * 如果没有etime且没有下个config，则设置成当前时间一年后
 */
func getETimeForPrepare(conf *splashModel.SplashScreenConfig, nextConf *splashModel.SplashScreenConfig) (etime xtime.Time) {
	if conf.ETime != 0 {
		if nextConf == nil || conf.ETime < nextConf.STime {
			etime = conf.ETime
		} else {
			etime = nextConf.STime - 1
		}
		return
	}
	if nextConf == nil {
		// 如果没有etime，就设置为当前时间一年后
		etime = xtime.Time(time.Now().AddDate(1, 0, 0).Unix())
	} else {
		etime = nextConf.STime - 1
	}
	return
}

var (
	outputImageMapLock sync.Mutex
)

// GetSplashConfigOnline 给网关的配置
//
//nolint:gocognit
func (s *Service) GetSplashConfigOnline() (onlineConfig *splashModel.GatewayConfig, err error) {
	var (
		imageList             []*splashModel.SplashScreenImage
		categoryList          []*splashModel.Category
		imageListForGateway   []*splashModel.SplashScreenImageForGateway
		rawImageMap           = make(map[int64]*splashModel.SplashScreenImage)
		imageMap              = splashModel.ImageMap{}
		outputImageMap        = splashModel.ImageMap{}
		defaultConfig         *splashModel.SplashScreenConfig
		selectConfig          *splashModel.SplashScreenConfig
		prepareDefaultConfigs []*splashModel.SplashScreenConfig
		prepareSelectConfigs  []*splashModel.SplashScreenConfig
	)

	eg := errgroup.Group{}

	eg.Go(func() (e error) {
		imageList, e = s.dao.GetImageList()
		if e != nil {
			return e
		}
		imageListForGateway = make([]*splashModel.SplashScreenImageForGateway, 0, len(imageList))
		for _, image := range imageList {
			_image := image

			logoConfig := &splashModel.LogoConfig{
				Mode: _image.LogoMode,
			}
			logoConfig.Show = _image.LogoHideFlag == splashModel.LogoShow
			switch logoConfig.Mode {
			case splashModel.LogoModePink:
				logoConfig.ImageUrl = conf.Conf.SplashScreen.Logo.Pink
				//nolint:gosimple
				break
			case splashModel.LogoModeWhite:
				logoConfig.ImageUrl = conf.Conf.SplashScreen.Logo.White
			case splashModel.LogoModeUser:
				logoConfig.ImageUrl = _image.LogoImageUrl
				//nolint:gosimple
				break
			}
			imageForGateway := &splashModel.SplashScreenImageForGateway{
				ID:              _image.ID,
				ImageName:       _image.ImageName,
				Mode:            _image.Mode,
				LogoConfig:      logoConfig,
				KeepNewDays:     conf.Conf.SplashScreen.Img.KeepNewDays,
				InitialPushTime: _image.InitialPushTime.Time().Unix(),
				CategoryIDs:     _image.CategoryIDs,
			}
			switch image.Mode {
			case splashModel.ImageModeHalfScreen:
				imageForGateway.ImageUrl = _image.ImageUrl
				//nolint:gosimple
				break
			case splashModel.ImageModeFullScreen:
				imageForGateway.FullScreenImageUrl = &splashModel.FullScreenImageUrl{
					Normal: _image.ImageUrlNormal,
					Full:   _image.ImageUrlFull,
					Pad:    _image.ImageUrlPad,
				}
			}
			imageListForGateway = append(imageListForGateway, imageForGateway)
			rawImageMap[_image.ID] = _image
		}
		return
	})

	// 获取自选配置分类列表
	eg.Go(func() (err error) {
		if categoryList, err = s.dao.GetAllCategories(); err != nil {
			log.Error("Service: GetSplashConfigOnline GetAllCategories error %v", err)
			return err
		}
		return
	})

	// 默认配置
	eg.Go(func() (e error) {
		defaultConfig, _, e = s.dao.GetConfigOnline([]int{
			splashModel.ShowModeForceOrder,
			splashModel.ShowModeForceRate,
			splashModel.ShowModeDefaultOrder,
			splashModel.ShowModeDefaultRate,
		})
		if e != nil {
			return e
		}
		if defaultConfig != nil {
			defaultConfig.State = 3
		}
		return
	})

	// 自选配置
	eg.Go(func() (e error) {
		if selectConfig, e = s.dao.GetSelectConfigOnline(); err != nil {
			log.Error("Service: GetSplashConfigOnline GetSelectConfigOnline error %v", err)
			return err
		}
		return
	})

	// 预生效配置获取处理
	eg.Go(func() (e error) {
		// 预生效默认配置列
		//nolint:staticcheck
		if prepareDefaultConfigs, e = s.dao.GetConfigInDays([]int{
			splashModel.ShowModeForceOrder,
			splashModel.ShowModeForceRate,
			splashModel.ShowModeDefaultOrder,
			splashModel.ShowModeDefaultRate,
		}, PrepareConfigDaysNum); err != nil {
			log.Error("Service: GetSplashConfigOnline GetConfigInDays error %v", err)
			return err
		}
		// 预生效自选配置列
		if prepareSelectConfigs, e = s.dao.GetSelectConfigsInDays(PrepareConfigDaysNum); err != nil {
			log.Error("Service: GetSplashConfigOnline GetSelectConfigsInDays error %v", err)
			return err
		}
		return
	})

	err = eg.Wait()
	if err != nil {
		log.Error("GetSplashConfigOnline errgroup error(%v)", err)
		err = ecode.Error(ecode.RequestErr, "配置拉取失败")
		return
	}

	for _, imageItem := range imageListForGateway {
		imageMap[imageItem.ID] = imageItem
	}

	if defaultConfig != nil {
		// 如果没结束时间，就设置成当前时间一年后
		defaultConfig.ETime = getETime(defaultConfig)
		// 失效的图片对应配置过滤
		defaultConfig.ConfigJson = dealWithMissingConfig(defaultConfig, imageMap, outputImageMap)

		// 检查生效配置物料初次下发时间，若值为空则更新
		_defaultConfig := defaultConfig
		if _defaultConfig.ConfigJson != "" {
			var configJsons []*splashModel.ConfigDetail
			if _err := json.Unmarshal([]byte(_defaultConfig.ConfigJson), &configJsons); _err != nil {
				log.Error("Service: GetSplashConfigOnline Unmarshal (%s) error %v", _defaultConfig.ConfigJson, _err)
				return
			}
			for _, config := range configJsons {
				_config := config
				if _config.ImgId != 0 {
					if img, exists := rawImageMap[_config.ImgId]; exists {
						_img := img
						if _img.InitialPushTime.Time().Unix() <= 0 {
							_img.InitialPushTime = xtime.Time(time.Now().Unix())
							//nolint:biligowordcheck
							go func() {
								if _err := s.dao.UpdateImage(_img, "system"); _err != nil {
									log.Error("Service: GetSplashConfigOnline UpdateImage (%+v) error %v", _img, _err)
								}
							}()
						}
					}
				}
			}
		}
	}

	if selectConfig != nil {
		selectConfig.ETime = getETime(selectConfig)
		selectConfig.ConfigJson = dealWithMissingConfig(selectConfig, imageMap, outputImageMap)

		// 检查生效配置物料初次下发时间，若值为空则更新
		_selectConfig := selectConfig
		if _selectConfig.ConfigJson != "" {
			var configJsons []*splashModel.ConfigDetail
			if _err := json.Unmarshal([]byte(_selectConfig.ConfigJson), &configJsons); _err != nil {
				log.Error("Service: GetSplashConfigOnline Unmarshal (%s) error %v", _selectConfig.ConfigJson, _err)
				return
			}
			for _, config := range configJsons {
				_config := config
				if _config.ImgId != 0 {
					if img, exists := rawImageMap[_config.ImgId]; exists {
						_img := img
						if _img.InitialPushTime.Time().Unix() <= 0 {
							_img.InitialPushTime = xtime.Time(time.Now().Unix())
							//nolint:biligowordcheck
							go func() {
								if _err := s.dao.UpdateImage(_img, "system"); _err != nil {
									log.Error("Service: GetSplashConfigOnline UpdateImage (%+v) error %v", _img, _err)
								}
								log.Info("Service: GetSplashConfigOnline UpdateImage (%+v) done", _img)
							}()
						}
					}
				}
			}
		}
	}
	if conf.Conf.SplashScreen.BaseDefaultConfig != nil && conf.Conf.SplashScreen.BaseDefaultConfig.ConfigJson != "" {
		dealWithMissingConfig(conf.Conf.SplashScreen.BaseDefaultConfig, imageMap, outputImageMap)
	}

	// 处理预生效配置
	egForPrepare := errgroup.Group{}
	egForPrepare.Go(func() (e error) {
		//nolint:gosimple
		if prepareDefaultConfigs != nil && len(prepareDefaultConfigs) > 0 {
			for i, conf := range prepareDefaultConfigs {
				var nextConf *splashModel.SplashScreenConfig
				if i < len(prepareDefaultConfigs)-1 {
					nextConf = prepareDefaultConfigs[i+1]
				}
				conf.State = 1 // 预生效
				conf.ETime = getETimeForPrepare(conf, nextConf)
				conf.ConfigJson = dealWithMissingConfig(conf, imageMap, outputImageMap)
			}
		}
		return
	})
	egForPrepare.Go(func() (e error) {
		//nolint:gosimple
		if prepareSelectConfigs != nil && len(prepareSelectConfigs) > 0 {
			for i, _conf := range prepareSelectConfigs {
				var nextConf *splashModel.SplashScreenConfig
				if i < len(prepareSelectConfigs)-1 {
					nextConf = prepareSelectConfigs[i+1]
				}
				_conf.State = 1 // 预生效
				_conf.ETime = getETimeForPrepare(_conf, nextConf)
				_conf.ConfigJson = dealWithMissingConfig(_conf, imageMap, outputImageMap)
			}
		}
		return
	})
	err = egForPrepare.Wait()
	if err != nil {
		log.Error("GetSplashConfigOnline egForPrepare error(%v)", err)
		err = ecode.Error(ecode.RequestErr, "预生效配置处理失败")
		return
	}

	onlineConfig = &splashModel.GatewayConfig{
		ImageMap:              outputImageMap,
		DefaultConfig:         defaultConfig,
		BaseDefaultConfig:     conf.Conf.SplashScreen.BaseDefaultConfig,
		SelectConfig:          selectConfig,
		PrepareDefaultConfigs: prepareDefaultConfigs,
		PrepareSelectConfigs:  prepareSelectConfigs,
		Categories:            categoryList,
	}

	return
}

type sortConfigList []*splashModel.ConfigDetail

// Len() 排序用
func (s sortConfigList) Len() int {
	return len(s)
}

// Less() 排序用
func (s sortConfigList) Less(i, j int) bool {
	return s[i].Sort > s[j].Sort
}

// Swap() 排序用
func (s sortConfigList) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

// 对configJson进行校验，去除已经不存在的图片，修正排序，等比扩充概率
func dealWithMissingConfig(config *splashModel.SplashScreenConfig, imageMap splashModel.ImageMap, outputImageMap splashModel.ImageMap) (configJson string) {
	var (
		err                 error
		configList          []*splashModel.ConfigDetail
		configListAvailable sortConfigList
		leftRate            = 100
	)

	// 始终返回 json
	defer func() {
		if len(configListAvailable) == 0 {
			configJson = "[]"
		} else {
			j, e := json.Marshal(&configListAvailable)
			if e != nil {
				configJson = "[]"
				return
			}
			configJson = string(j)
		}
	}()

	if err = json.Unmarshal([]byte(config.ConfigJson), &configList); err != nil {
		log.Error("dealWithMissingConfig json.Unmarshal error(%v)", err)
		return
	}

	// 处理丢失的图片配置：
	// 顺序模式，顺序补位
	// 概率模式，概率等比增加，取整
	switch config.ShowMode {
	case splashModel.ShowModeForceOrder:
		fallthrough
	case splashModel.ShowModeDefaultOrder:
		fallthrough
	case splashModel.ShowModeSelect:
		for _, configItem := range configList {
			// 图片不存在以后，就不用返回配置了
			if _, ok := imageMap[configItem.ImgId]; ok {
				configListAvailable = append(configListAvailable, configItem)
				outputImageMapLock.Lock()
				outputImageMap[configItem.ImgId] = imageMap[configItem.ImgId]
				outputImageMapLock.Unlock()
			}
		}
		// 按照sort DESC
		sort.Sort(&configListAvailable)
		for i := range configListAvailable {
			configListAvailable[i].Position = i + 1
		}
	case splashModel.ShowModeForceRate:
		fallthrough
	case splashModel.ShowModeDefaultRate:
		for _, configItem := range configList {
			// 图片不存在以后，就不用返回配置了
			if _, ok := imageMap[configItem.ImgId]; ok {
				configListAvailable = append(configListAvailable, configItem)
				leftRate -= configItem.Rate
				outputImageMapLock.Lock()
				outputImageMap[configItem.ImgId] = imageMap[configItem.ImgId]
				outputImageMapLock.Unlock()
			}
		}
		if leftRate > 0 && len(configListAvailable) > 0 {
			// 说明总共的概率和不到 100%，需要修正
			averageAddRate := leftRate / len(configListAvailable)
			lastAddRate := leftRate - averageAddRate*(len(configListAvailable)-1)

			for i := range configListAvailable {
				if i == (len(configListAvailable) - 1) {
					configListAvailable[i].Rate += lastAddRate
				} else {
					configListAvailable[i].Rate += averageAddRate
				}
			}
		}
	}

	return
}

func validateSplashImageModification(image *splashModel.SplashScreenImage) (err error) {
	// 若LOGO显示且为自定义，那必须有上传的图片
	if (image.LogoShowFlag == 1 || image.LogoHideFlag == splashModel.LogoShow) && image.LogoMode == splashModel.LogoModeUser && image.LogoImageUrl == "" {
		err = ecode.Error(ecode.RequestErr, "未上传LOGO图片")
		return
	}
	if image.Mode == splashModel.ImageModeHalfScreen {
		// 若为半屏物料，则必须有上传的图片
		if image.ImageUrl == "" {
			err = ecode.Error(ecode.RequestErr, "未上传物料图片")
			return
		}
	} else if image.Mode == splashModel.ImageModeFullScreen {
		// 若为全屏物料，则必需有上传的三张图片
		if image.ImageUrlNormal == "" || image.ImageUrlFull == "" || image.ImageUrlPad == "" {
			err = ecode.Error(ecode.RequestErr, "未上传完整物料图片")
			return
		}
	} else {
		// 闪屏模式参数不正确
		err = ecode.Error(ecode.RequestErr, "闪屏模式参数不正确")
		return
	}
	return
}

// 新增图片物料
func (s *Service) AddSplashImage(newImage *splashModel.SplashScreenImage, username string, uid int64) (id int64, err error) {
	lNewImage := newImage

	// valiate request
	if err = validateSplashImageModification(lNewImage); err != nil {
		log.Error("AddSplashImage s.validateSplashImageModification error(%v)", err)
		return
	}

	// 若LOGO选为白色或粉色，则一直为默认的配置中的LOGO URL
	if lNewImage.LogoMode == splashModel.LogoModePink || lNewImage.LogoMode == splashModel.LogoModeWhite {
		lNewImage.LogoImageUrl = ""
	}

	if id, err = s.dao.InsertImage(lNewImage, username); err != nil {
		log.Error("AddSplashImage s.dao.InsertImage error(%v)", err)
		return
	}
	log.Info("AddSplashImage succeed params(%v) user(%v)", newImage, username)

	obj := map[string]interface{}{
		"value": lNewImage,
	}
	if err = Log.AddLog(splashModel.ActionLogBusiness, username, uid, id, "AddSplashImage", obj); err != nil {
		log.Error("AddSplashImage AddLog error(%v)", err)
		return
	}
	return
}

//nolint:deadcode,unused
type updateImageReq struct {
	ID             int64  `json:"id" form:"id" validate:"required"`
	ImageName      string `json:"img_name" form:"img_name"`
	ImageUrl       string `json:"img_url" form:"img_url"`
	Mode           int    `json:"mode" form:"mode"`
	ImageUrlNormal string `json:"img_url_normal" form:"img_url_normal"`
	ImageUrlFull   string `json:"img_url_full" form:"img_url_full"`
	ImageUrlPad    string `json:"img_url_pad" form:"img_url_pad"`
	LogoShow       int    `json:"logo_show" form:"logo_show"`
	LogoMode       int    `json:"logo_mode" form:"logo_mode"`
	LogoImageUrl   string `json:"logo_img_url" form:"logo_img_url"`
}

// 更新图片物料
func (s *Service) UpdateSplashImage(updateImage *splashModel.SplashScreenImage, username string, uid int64) (err error) {
	lUpdateImage := updateImage

	if exist := s.checkImageExist(updateImage.ID); exist != 1 {
		err = ecode.Error(ecode.RequestErr, "配置不存在")
		return
	}

	// valiate request
	if err = validateSplashImageModification(lUpdateImage); err != nil {
		log.Error("AddSplashImage s.validateSplashImageModification error(%v)", err)
		return
	}

	// 若LOGO选为白色或粉色，则一直为默认的配置中的LOGO URL
	if lUpdateImage.LogoMode == splashModel.LogoModePink || lUpdateImage.LogoMode == splashModel.LogoModeWhite {
		lUpdateImage.LogoImageUrl = ""
	}

	if err = s.dao.UpdateImage(lUpdateImage, username); err != nil {
		log.Error("UpdateSplashImage s.dao.UpdateImage error(%v)", err)
		return
	}

	log.Info("UpdateSplashImage succeed params(%v) user(%v)", updateImage, username)

	obj := map[string]interface{}{
		"value": lUpdateImage,
	}
	if err = Log.AddLog(splashModel.ActionLogBusiness, username, uid, updateImage.ID, "UpdateSplashImage", obj); err != nil {
		log.Error("UpdateSplashImage AddLog error(%v)", err)
		return
	}
	return
}

// 删除图片物料
func (s *Service) DeleteSplashImage(id int64, username string, uid int64) (err error) {
	if exist := s.checkImageExist(id); exist != 1 {
		err = ecode.Error(ecode.RequestErr, "配置不存在")
		return
	}

	if err = s.dao.DeleteImage(id, username); err != nil {
		log.Error("DeleteSplashImage s.dao.DeleteImage error(%v)", err)
		return
	}

	log.Info("DeleteSplashImage succeed params(%v) user(%v)", id, username)

	obj := map[string]interface{}{
		"value": id,
	}
	if err = Log.AddLog(splashModel.ActionLogBusiness, username, uid, id, "DeleteSplashImage", obj); err != nil {
		log.Error("DeleteSplashImage AddLog error(%v)", err)
		return
	}
	return
}

// GetSplashImageList 获取全部的物料列表
func (s *Service) GetSplashImageList() (splashImageList []*splashModel.SplashScreenImage, err error) {
	if splashImageList, err = s.dao.GetImageList(); err != nil {
		log.Error("GetSplashImageList s.dao.GetImageList error(%v)", err)
	}
	for _, image := range splashImageList {
		// 将默认LOGO图片URL传出
		switch image.LogoMode {
		case splashModel.LogoModeWhite:
			image.LogoImageUrl = conf.Conf.SplashScreen.Logo.White
			//nolint:gosimple
			break
		case splashModel.LogoModePink:
			image.LogoImageUrl = conf.Conf.SplashScreen.Logo.Pink
			//nolint:gosimple
			break
		}
		// 为前端做转换，1为显示2为隐藏
		image.LogoShowFlag = 1
		if image.LogoHideFlag == splashModel.LogoHide {
			image.LogoShowFlag = 2
		}
	}
	return
}

// 检查用户传来的configJson内容
func checkConfigJson(splashConfig *splashModel.SplashScreenConfig) (err error) {
	var (
		configList    []*splashModel.ConfigDetail
		totalRate     = 0
		configObjTemp = map[int]int{}
	)
	if err = json.Unmarshal([]byte(splashConfig.ConfigJson), &configList); err != nil {
		log.Error("checkConfigJson json.Unmarshal error(%v)", err)
		err = ecode.Error(ecode.RequestErr, "JSON 解析错误")
		return
	}

	for _, configItem := range configList {
		if configItem.ImgId < 1 {
			err = ecode.Error(ecode.RequestErr, "配置内图片不能为空")
			return
		}

		switch splashConfig.ShowMode {
		case splashModel.ShowModeForceOrder:
			fallthrough
		case splashModel.ShowModeDefaultOrder:
			fallthrough
		case splashModel.ShowModeSelect:
			if !(configItem.Position > 0) {
				err = ecode.Error(ecode.RequestErr, "配置内顺序需是大于0的整数")
				return
			}

			if configObjTemp[configItem.Position] == 1 {
				err = ecode.Error(ecode.RequestErr, "配置内不能有重复位置")
				return
			} else {
				configObjTemp[configItem.Position] = 1
			}
		case splashModel.ShowModeForceRate:
			fallthrough
		case splashModel.ShowModeDefaultRate:
			if !(configItem.Rate > 0 && configItem.Rate <= 100) {
				err = ecode.Error(ecode.RequestErr, "配置内概率需是大于0且不大于100的整数")
				return
			}
			totalRate += configItem.Rate
		}
	}

	if (splashConfig.ShowMode == splashModel.ShowModeDefaultRate || splashConfig.ShowMode == splashModel.ShowModeForceRate) &&
		totalRate != 100 {
		err = ecode.Error(ecode.RequestErr, "配置内概率概率之需等于100%")
		return
	}

	return
}

// 通过一个showMode，得到同类型的showMode列表
func getShowModeList(showMode int) (showModeList []int) {
	if showMode == splashModel.ShowModeSelect {
		showModeList = []int{splashModel.ShowModeSelect}
	} else {
		showModeList = []int{
			splashModel.ShowModeForceOrder,
			splashModel.ShowModeForceRate,
			splashModel.ShowModeDefaultOrder,
			splashModel.ShowModeDefaultRate,
		}
	}
	return
}

// 检查策略配置是否存在
func (s *Service) checkConfigExist(id int64) (exist int, detail *splashModel.SplashScreenConfig) {
	var (
		err error
	)
	if detail, err = s.dao.GetConfigDetail(id); err != nil {
		exist = 0
		return
	}
	if detail.ID != 0 {
		exist = 1
	}
	return
}

// 检查图片物料是否存在
func (s *Service) checkImageExist(id int64) (res int) {
	var (
		detail *splashModel.SplashScreenImage
		err    error
	)
	if detail, err = s.dao.GetImageDetail(id); err != nil {
		res = 0
		return
	}
	if detail.ID != 0 {
		res = 1
	}
	return
}

// 新增配置策略
func (s *Service) AddSplashConfig(newConfig *splashModel.SplashScreenConfig, username string, uid int64) (id int64, err error) {
	var (
		//nolint:ineffassign
		conflictCnt = 0
	)

	// 如果是立即生效，就设置stime为当前时间，否则取用前端传来的stime
	if newConfig.IsImmediately > 0 {
		newConfig.STime = xtime.Time(time.Now().Unix())
	} else if newConfig.STime == 0 {
		err = ecode.Error(ecode.RequestErr, "请传入正确的开始时间")
		return
	}

	if err = checkConfigJson(newConfig); err != nil {
		return
	}

	if conflictCnt, err = s.dao.GetConfigConflictCnt(getShowModeList(newConfig.ShowMode), newConfig.STime, 0); err != nil {
		log.Error("AddSplashConfig s.dao.GetConfigConflictCnt error(%v)", err)
		return
	}
	if conflictCnt > 0 {
		err = ecode.Error(ecode.RequestErr, "配置的开始时间不能和已有配置相同")
		return
	}

	if id, err = s.dao.InsertConfig(newConfig, username); err != nil {
		log.Error("AddSplashConfig s.dao.InsertConfig error(%v)", err)
		return
	}

	log.Info("AddSplashConfig succeed params(%v) user(%v)", newConfig, username)

	obj := map[string]interface{}{
		"value": newConfig,
	}
	if err = Log.AddLog(splashModel.ActionLogBusiness, username, uid, id, "AddSplashConfig", obj); err != nil {
		log.Error("AddSplashConfig AddLog error(%v)", err)
		return
	}
	return
}

// 更新配置策略
func (s *Service) UpdateSplashConfig(updateConfig *splashModel.SplashScreenConfig, username string, uid int64) (err error) {
	var (
		//nolint:ineffassign
		conflictCnt = 0
		//nolint:ineffassign
		exist      = 0
		lastDetail *splashModel.SplashScreenConfig
	)

	if err = checkConfigJson(updateConfig); err != nil {
		return
	}

	if exist, lastDetail = s.checkConfigExist(updateConfig.ID); exist != 1 {
		err = ecode.Error(ecode.RequestErr, "配置不存在")
		return
	}

	if updateConfig.IsImmediately == 1 {
		// 定时生效，想要变成立即生效，就把stime设置成当前时间
		updateConfig.STime = xtime.Time(time.Now().Unix())
	}

	if conflictCnt, err = s.dao.GetConfigConflictCnt(
		getShowModeList(updateConfig.ShowMode),
		updateConfig.STime, updateConfig.ID); err != nil {
		log.Error("UpdateSplashConfig s.dao.GetConfigConflictCnt error(%v)", err)
		return
	}
	if conflictCnt > 0 {
		err = ecode.Error(ecode.RequestErr, "配置的开始时间不能和已有配置相同")
		return
	}

	updateConfigMap := map[string]interface{}{
		"immediately":      updateConfig.IsImmediately,
		"stime":            updateConfig.STime,
		"etime":            updateConfig.ETime,
		"show_mode":        updateConfig.ShowMode,
		"config_json":      updateConfig.ConfigJson,
		"force_show_times": updateConfig.ForceShowTimes,
	}

	if err = s.dao.UpdateConfigMap(updateConfig.ID, updateConfigMap, username); err != nil {
		log.Error("UpdateSplashConfig s.dao.UpdateConfigMap error(%v)", err)
		return
	}

	// 更新配置的时候，如果原本配置已经失效、手动下线，那么就将etime重新生效，同时设置状态为待审核
	if lastDetail.ETime != 0 &&
		lastDetail.ETime < xtime.Time(time.Now().Unix()) &&
		(lastDetail.AuditState == splashModel.AuditStatePass || lastDetail.AuditState == splashModel.AuditStateOffline) {
		if err = s.dao.ResetConfigETimeAndAuditState(updateConfig.ID, updateConfig.ETime); err != nil {
			log.Error("UpdateSplashConfig s.dao.ResetConfigETimeAndAuditState error(%v)", err)
			return
		}
	}

	log.Info("UpdateSplashConfig succeed params(%v) user(%v)", updateConfig, username)

	obj := map[string]interface{}{
		"value": updateConfig,
	}
	if err = Log.AddLog(splashModel.ActionLogBusiness, username, uid, updateConfig.ID, "UpdateSplashConfig", obj); err != nil {
		log.Error("UpdateSplashConfig AddLog error(%v)", err)
		return
	}
	return
}

// 获取showMode列表内的配置，带分页
func (s *Service) GetSplashConfigList(showMode int, ps, pn int32) (configList []*splashModel.SplashScreenConfig, count int32, err error) {
	if configList, count, err = s.dao.GetConfigListAll(getShowModeList(showMode), ps, pn); err != nil {
		log.Error("GetSplashConfigList s.dao.GetConfigListAll error(%v)", err)
		return
	}

	var (
		onlineConfig *splashModel.SplashScreenConfig
	)
	if onlineConfig, _, err = s.dao.GetConfigOnline(getShowModeList(showMode)); err != nil {
		log.Error("GetSplashConfigList s.dao.GetConfigOnline error(%v)", err)
	}

	for i, config := range configList {
		// state 表示状态：
		// 0、待通过
		// 1、待生效
		// 2、已失效
		// 3、生效中
		// 4、手动下线
		if onlineConfig != nil && onlineConfig.ID != 0 && config.ID == onlineConfig.ID {
			// 生效中
			configList[i].State = 3
		} else if config.AuditState == splashModel.AuditStatePass && config.STime > xtime.Time(time.Now().Unix()) {
			// 待生效
			configList[i].State = 1
		} else if config.AuditState == splashModel.AuditStateCancel && (config.ETime > xtime.Time(time.Now().Unix()) || config.ETime == 0) {
			// 待通过
			configList[i].State = 0
		} else if config.AuditState == splashModel.AuditStateOffline {
			// 手动下线
			configList[i].State = 4
		} else {
			// 已失效
			configList[i].State = 2
		}
	}

	return
}

// UpdateAuditState 更新默认配置审核状态
func (s *Service) UpdateAuditState(id int64, auditState int, username string, uid int64) (err error) {
	if exist, _ := s.checkConfigExist(id); exist != 1 {
		err = feedEcode.SplashScreenConfigNotExists
		return
	}

	if err = s.dao.UpdateConfigAuditState(id, auditState, username); err != nil {
		log.Error("UpdateAuditState s.dao.UpdateConfigAuditState error(%v)", err)
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
	if err = Log.AddLog(splashModel.ActionLogBusiness, username, uid, id, "UpdateAuditState", obj); err != nil {
		log.Error("UpdateAuditState AddLog error(%v)", err)
		return
	}
	return
}

// 更新失效配置的etime，被ETimeEditMonitor()调用
func (s *Service) CheckAndUpdateETime() (err error) {
	var (
		currentDefaultConfig *splashModel.SplashScreenConfig
		lastDefaultConfig    *splashModel.SplashScreenConfig
		currentSelectConfig  *splashModel.SplashScreenConfig
		lastSelectConfig     *splashModel.SplashScreenConfig
	)
	if currentDefaultConfig, lastDefaultConfig, err = s.dao.GetConfigOnline([]int{
		splashModel.ShowModeForceOrder,
		splashModel.ShowModeForceRate,
		splashModel.ShowModeDefaultOrder,
		splashModel.ShowModeDefaultRate,
	}); err != nil {
		return
	}

	if currentDefaultConfig != nil && currentDefaultConfig.ID != 0 &&
		lastDefaultConfig != nil && lastDefaultConfig.ID != 0 &&
		(lastDefaultConfig.ETime == xtime.Time(0) || lastDefaultConfig.ETime > currentDefaultConfig.STime) {
		// 老配置结束时间更新成新配置开始时间
		if err = s.dao.UpdateConfig(&splashModel.SplashScreenConfig{
			ID:    lastDefaultConfig.ID,
			ETime: currentDefaultConfig.STime,
		}, "system"); err != nil {
			log.Error("CheckAndUpdateETime s.dao.UpdateConfig error(%v)", err)
			return
		}
	}

	if currentSelectConfig, lastSelectConfig, err = s.dao.GetConfigOnline([]int{
		splashModel.ShowModeSelect,
	}); err != nil {
		return
	}

	if currentSelectConfig != nil && currentSelectConfig.ID != 0 &&
		lastSelectConfig != nil && lastSelectConfig.ID != 0 &&
		(lastSelectConfig.ETime == xtime.Time(0) || lastSelectConfig.ETime > currentSelectConfig.STime) {
		// 老配置结束时间更新成新配置开始时间
		if err = s.dao.UpdateConfig(&splashModel.SplashScreenConfig{
			ID:    lastSelectConfig.ID,
			ETime: currentSelectConfig.STime,
		}, "system"); err != nil {
			log.Error("CheckAndUpdateETime s.dao.UpdateConfig error(%v)", err)
			return
		}
	}

	return
}

// 定时更新失效配置的etime，用于管理后台展示
func (s *Service) ETimeEditMonitor() (err error) {
	log.Info("ETimeEditMonitor loop start at time(%v)", time.Now())
	for {
		log.Info("ETimeEditMonitor check at time(%v)", time.Now())
		err = s.CheckAndUpdateETime()
		if err != nil {
			log.Error("s.CheckAndUpdateETime error %v", err)
		}
		time.Sleep(9 * time.Second)
	}
}
