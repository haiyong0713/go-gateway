package tianma

import (
	"context"
	"encoding/json"
	"fmt"
	xecode "go-common/library/ecode"
	"go-common/library/log"
	xtime "go-common/library/time"
	"go-gateway/app/app-svr/app-feed/admin/conf"
	model "go-gateway/app/app-svr/app-feed/admin/model/tianma"
	"go-gateway/app/app-svr/app-feed/admin/util"
	"go-gateway/app/app-svr/app-feed/ecode"
	"strconv"
	"time"

	bgroupGRPC "git.bilibili.co/bapis/bapis-go/platform/service/bgroup"
)

// AddPopupConfig 新建弹窗配置
func (s *Service) AddPopupConfig(toAddConfig *model.PopupConfig, username string, uid int64) (addedID int64, err error) {
	if err = s.validatePopupBuildsJSON(toAddConfig.Builds); err != nil {
		return
	}

	if err = s.validateExistingDurationForEditing(toAddConfig); err != nil {
		log.Error("AddPopupConfig s.validateExistingDurationForEditing error(%v)", err)
		return
	}

	// 新建人群包
	groupFilePath := fmt.Sprintf("boss://%s?path=%s", conf.Conf.Boss.Bucket, toAddConfig.CrowdValue)
	addBGroupReq := &bgroupGRPC.AddBGroupReq{
		BusinessName: conf.Conf.Popup.BGroupBusinessName, // 业务id
		Name:         toAddConfig.Description,            // 人群包名称，用配置描述作为名称
		Path:         groupFilePath,                      // 文件地址
		Source:       1,                                  // 人群包类型 1:用户自定义上传文件 2:用户提供bsk地址
		StartTime:    int64(toAddConfig.STime),           // 生效时间
		EndTime:      int64(toAddConfig.ETime),           // 失效时间
		Username:     username,                           // 添加人员
	}
	// 数据维度 1:mid维度 2:buvid维度
	if toAddConfig.CrowdBase == model.PopupCrowdBaseBGroupMID {
		addBGroupReq.Dimension = 1
	} else if toAddConfig.CrowdBase == model.PopupCrowdBaseBGroupBuvid {
		addBGroupReq.Dimension = 2
	} else {
		err = ecode.PopupConfigParameterError
		return
	}
	var addBGroupResp *bgroupGRPC.AddBGroupResp
	addBGroupResp, err = s.dao.AddBGroup(context.Background(), addBGroupReq)
	if err != nil {
		log.Error("AddPopupConfig AddBGroup error (%v)", err)
		err = ecode.BGroupAddError
		return
	}
	if addBGroupResp == nil || addBGroupResp.Id == 0 {
		err = ecode.BGroupAddError
		log.Error("AddPopupConfig AddBGroup error (%v)", err)
		return
	}
	bgroupID := addBGroupResp.Id
	toAddConfig.CrowdValue = strconv.FormatInt(bgroupID, 10)

	if addedID, err = s.dao.AddPopupConfig(toAddConfig, username, uid); err != nil {
		log.Error("AddPopupConfig dao.AddPopupConfig error(%v)", err)
		return
	}

	log.Info("AddPopupConfig succeed params(%v) user(%v)", toAddConfig, username)

	obj := map[string]interface{}{
		"value": toAddConfig,
	}
	if err = util.AddLog(model.PopupActionLogBusinessID, username, uid, addedID, model.PopupActionLogAddPopupConfig, obj); err != nil {
		log.Error("AddPopupConfig AddLog error(%v)", err)
		return
	}
	return
}

// UpdatePopupConfig 更新弹窗配置
//
//nolint:gocognit
func (s *Service) UpdatePopupConfig(updateParams *model.PopupConfig, username string, uid int64) (err error) {
	if updateParams == nil {
		return
	}
	if updateParams.ID == 0 {
		err = ecode.PopupConfigParameterError
		return
	}
	// 检查对应ID配置是否存在，不存在则返回error
	var existingConfig *model.PopupConfig
	if existingConfig, err = s.dao.GetPopupConfigByID(updateParams.ID); err != nil {
		log.Error("popupConfig.service.UpdatePopupConfig.GetPopupConfigByID error (%v)", err)
		return
	}
	if existingConfig == nil || existingConfig.ID == 0 {
		err = ecode.PopupConfigNotFound
		log.Error("popupConfig.service.UpdatePopupConfig.GetPopupConfigByID not found")
		return
	}

	// 检查builds字符串
	if updateParams.Builds != "" {
		if err = s.validatePopupBuildsJSON(updateParams.Builds); err != nil {
			log.Error("popupConfig.service.UpdatePopupConfig.validatePopupBuildsJSON error(%v)", err)
			return
		}
	}
	// 检查时间冲突
	if updateParams.STime != 0 || updateParams.ETime != 0 {
		if err = s.validateExistingDurationForEditing(updateParams); err != nil {
			log.Error("popupConfig.service.UpdatePopupConfig.validateExistingDurationForEditing error(%v)", err)
			return
		}
	}
	// 若变更人群定向指向，则变更人群包
	if updateParams.CrowdValue != "" {
		if updateParams.CrowdType == model.PopupCrowdTypeBGroup || (updateParams.CrowdType == 0 && existingConfig.CrowdType == model.PopupCrowdTypeBGroup) {
			// 人群包参数
			groupFilePath := fmt.Sprintf("boss://%s?path=%s", conf.Conf.Boss.Bucket, updateParams.CrowdValue)
			addBGroupReq := &bgroupGRPC.AddBGroupReq{
				BusinessName: conf.Conf.Popup.BGroupBusinessName, // 业务id
				Name:         updateParams.Description,           // 人群包名称，用配置描述作为名称
				Path:         groupFilePath,                      // 文件地址
				Source:       1,                                  // 人群包类型 1:用户自定义上传文件 2:用户提供bsk地址
				StartTime:    int64(updateParams.STime),          // 生效时间
				EndTime:      int64(updateParams.ETime),          // 失效时间
				Username:     username,                           // 添加人员
			}
			if addBGroupReq.Name == "" {
				addBGroupReq.Name = existingConfig.Description
			}
			if addBGroupReq.StartTime == 0 {
				addBGroupReq.StartTime = int64(existingConfig.STime)
			}
			if addBGroupReq.EndTime == 0 {
				addBGroupReq.EndTime = int64(existingConfig.ETime)
			}
			// 数据维度 1:mid维度 2:buvid维度
			switch updateParams.CrowdBase {
			case 0:
				switch existingConfig.CrowdBase {
				case model.PopupCrowdBaseBGroupMID:
					addBGroupReq.Dimension = 1
					//nolint:gosimple
					break
				case model.PopupCrowdBaseBGroupBuvid:
					addBGroupReq.Dimension = 2
					//nolint:gosimple
					break
				default:
					err = ecode.PopupConfigParameterError
					return
				}
			case model.PopupCrowdBaseBGroupMID:
				addBGroupReq.Dimension = 1
				//nolint:gosimple
				break
			case model.PopupCrowdBaseBGroupBuvid:
				addBGroupReq.Dimension = 2
				//nolint:gosimple
				break
			default:
				err = ecode.PopupConfigParameterError
				return
			}

			// 若为已存在人群包，则更新。否则新建人群包
			if existingConfig.CrowdType == model.PopupCrowdTypeBGroup {
				// 更新人群包
				var bgroupID int64
				if bgroupID, err = strconv.ParseInt(existingConfig.CrowdValue, 10, 64); err != nil {
					log.Error("popupConfig.service.UpdatePopupConfig Parse BGroup ID (%s) error (%v)", existingConfig.CrowdValue, err)
					err = ecode.BGroupIDError
					return
				}
				updateBGroupReq := &bgroupGRPC.UpdateBGroupReq{
					Id:           bgroupID,                      // 人群包ID
					BusinessName: addBGroupReq.BusinessName,     // 业务id
					Name:         addBGroupReq.Name,             // 人群包名称，用配置描述作为名称
					Dimension:    addBGroupReq.Dimension,        // 数据维度 1:mid维度 2:buvid维度
					Path:         addBGroupReq.Path,             // 文件地址
					Source:       addBGroupReq.Source,           // 人群包类型 1:用户自定义上传文件 2:用户提供bsk地址
					StartTime:    int64(addBGroupReq.StartTime), // 生效时间
					EndTime:      int64(addBGroupReq.EndTime),   // 失效时间
					Username:     username,                      // 添加人员
				}
				var updateBGroupResp *bgroupGRPC.UpdateBGroupResp
				updateBGroupResp, err = s.dao.UpdateBGroup(context.Background(), updateBGroupReq)
				if err != nil {
					log.Error("popupConfig.service.UpdatePopupConfig UpdateBGroup error (%v)", err)
					err = ecode.BGroupUpdateError
					return
				}
				if updateBGroupResp == nil || updateBGroupResp.Id == 0 {
					err = ecode.BGroupUpdateError
					log.Error("popupConfig.service.UpdatePopupConfig UpdateBGroup error (%v)", err)
					return
				}
				bgroupID = updateBGroupResp.Id
				updateParams.CrowdValue = strconv.FormatInt(bgroupID, 10)
			} else {
				// 新建人群包
				var addBGroupResp *bgroupGRPC.AddBGroupResp
				addBGroupResp, err = s.dao.AddBGroup(context.Background(), addBGroupReq)
				if err != nil {
					log.Error("popupConfig.service.UpdatePopupConfig AddBGroup error (%v)", err)
					err = ecode.BGroupAddError
					return
				}
				if addBGroupResp == nil || addBGroupResp.Id == 0 {
					err = ecode.BGroupAddError
					log.Error("popupConfig.service.UpdatePopupConfig AddBGroup error (%v)", err)
					return
				}
				bgroupID := addBGroupResp.Id
				updateParams.CrowdValue = strconv.FormatInt(bgroupID, 10)
			}
		}
	}

	if err = s.dao.UpdatePopupConfig(updateParams, username, uid); err != nil {
		log.Error("popupConfig.service.UpdatePopupConfig s.dao.UpdatePopupConfig error(%v)", err)
		return
	}

	log.Info("popupConfig.service.UpdatePopupConfig succeed params(%v) user(%v)", updateParams, username)

	obj := map[string]interface{}{
		"value": updateParams,
	}
	if err = util.AddLog(model.PopupActionLogBusinessID, username, uid, updateParams.ID, model.PopupActionLogUpdatePopupConfig, obj); err != nil {
		log.Error("popupConfig.service.UpdatePopupConfig AddLog error(%v)", err)
		return
	}
	return
}

// GetPopupConfigList 获取弹窗配置，带分页
func (s *Service) GetPopupConfigList(query *model.PopupConfig, ps int, pn int, order string) (configList []*model.PopupConfig, total int64, err error) {
	var tempConfigList []*model.PopupConfig
	// 若根据ID获取数据
	if query != nil && query.ID != 0 {
		configList = make([]*model.PopupConfig, 0, 1)
		var config *model.PopupConfig
		if config, err = s.dao.GetPopupConfigByID(query.ID); err != nil {
			log.Error("GetPopupConfigList s.dao.GetPopupConfigByID error(%v", err)
			return
		}
		tempConfigList = []*model.PopupConfig{config}
		total = 1
	} else {
		// 由于需要计算得出status，且配置数据量通常不大，就不用sql写分页了，取出所有后处理
		if tempConfigList, total, err = s.dao.GetPopupConfigAll(0, 0, order); err != nil {
			log.Error("GetPopupConfigList s.dao.GetPopupConfigAll error(%v)", err)
			return
		}
		configList = make([]*model.PopupConfig, 0, total)
	}
	// 分页默认值
	if ps == 0 {
		ps = 10
	}
	if pn == 0 {
		pn = 1
	}
	var (
		onlineConfig *model.PopupConfig
	)
	if onlineConfig, err = s.dao.GetPopupConfigOnline(); err != nil {
		log.Error("GetPopupConfigList s.dao.GetPopupConfigOnline error(%v)", err)
		return
	}

	// status处理 & filter
	for i, config := range tempConfigList {
		// 配置状态。1-在线，2-已下线，3-待生效，4-已过期
		if onlineConfig != nil && onlineConfig.ID != 0 && config.ID == onlineConfig.ID {
			// 在线
			tempConfigList[i].Status = model.PopupStatusOnline
		} else if config.AuditState == model.PopupAuditStateOffline {
			// 已下线
			tempConfigList[i].Status = model.PopupStatusManualOffline
		} else if config.STime > xtime.Time(time.Now().Unix()) {
			// 待生效
			tempConfigList[i].Status = model.PopupStatusReadyOnline
		} else {
			// 自动失效
			tempConfigList[i].Status = model.PopupStatusAutoOffline
		}

		if query == nil || query.Status == 0 {
			configList = append(configList, tempConfigList[i])
		} else if query.Status != 0 && query.Status == tempConfigList[i].Status {
			configList = append(configList, tempConfigList[i])
		}
	}

	// 分页
	total = int64(len(configList))
	pmin := ps * (pn - 1)
	pmax := ps * pn
	if pmin >= len(configList) {
		configList = make([]*model.PopupConfig, 0)
		return
	}
	if pmax >= len(configList) {
		pmax = len(configList)
	}
	configList = configList[pmin:pmax]

	return
}

// DeletePopupConfig 删除弹窗配置
func (s *Service) DeletePopupConfig(id int64, username string, uid int64) (err error) {
	if id == 0 {
		err = ecode.PopupConfigParameterError
		return
	}
	if exists, _ := s.checkPopupConfigExist(id); !exists {
		err = ecode.PopupConfigNotFound
		return
	}

	if err = s.dao.DeletePopupConfig(&model.PopupConfig{ID: id}, username, uid); err != nil {
		log.Error("DeletePopupConfig s.dao.DeletePopupConfig error(%v)", err)
		return
	}

	log.Info("DeletePopupConfig succeed id(%d) user(%v)", id, username)

	obj := map[string]interface{}{
		"value": map[string]interface{}{"id": id},
	}
	if err = util.AddLog(model.PopupActionLogBusinessID, username, uid, id, model.PopupActionLogDeletePopupConfig, obj); err != nil {
		log.Error("popupConfig.service.UpdatePopupConfig AddLog error(%v)", err)
		return
	}
	return
}

// AuditPopupConfig 弹窗审核
func (s *Service) AuditPopupConfig(id int64, auditState int, username string, uid int64) (err error) {
	if id == 0 || auditState == 0 {
		err = ecode.PopupConfigParameterError
		return
	}
	// 获取对应ID config，若不存在返回错误
	var config *model.PopupConfig
	config, err = s.dao.GetPopupConfigByID(id)
	if err != nil {
		return
	}
	if config == nil || config.ID == 0 {
		err = ecode.PopupConfigNotFound
		return
	}

	// 若重新上线，检查时间冲突
	if config.AuditState == model.PopupAuditStateOffline && auditState != model.PopupAuditStateOffline {
		validateConfig := config
		validateConfig.ID = 0
		if err = s.validateExistingDurationForEditing(validateConfig); err != nil {
			log.Error("AuditPopupConfig s.dao.validateExistingDurationForEditing error(%v)", err)
			return
		}
	}

	if err = s.dao.UpdatePopupConfig(&model.PopupConfig{ID: id, AuditState: auditState}, username, uid); err != nil {
		log.Error("AuditPopupConfig s.dao.UpdatePopupConfig error(%v)", err)
		return
	}

	log.Info("AuditPopupConfig succeed id(%d) with audit state(%d) user(%v)", id, auditState, username)

	obj := map[string]interface{}{
		"value": map[string]interface{}{"id": id, "auditState": auditState},
	}
	if err = util.AddLog(model.PopupActionLogBusinessID, username, uid, id, model.PopupActionLogAuditPopupConfig, obj); err != nil {
		log.Error("popupConfig.service.AuditPopupConfig AddLog error(%v)", err)
		return
	}
	return
}

// checkPopupConfigExist 检查弹窗配置是否存在
func (s *Service) checkPopupConfigExist(id int64) (exist bool, detail *model.PopupConfig) {
	var err error
	exist = false
	if detail, err = s.dao.GetPopupConfigByID(id); err != nil {
		exist = false
		return
	}
	if detail != nil && detail.ID != 0 {
		exist = true
	}
	return
}

// validatePopupBuildsJSON 校验版本限制JSON
func (s *Service) validatePopupBuildsJSON(buildsJSON string) error {
	// 当builds为空，无版本限制，不当作出错
	if buildsJSON == "" || buildsJSON == "[]" {
		return nil
	}
	var builds []*model.PopupConfigBuild
	if err := json.Unmarshal([]byte(buildsJSON), &builds); err != nil {
		return ecode.PopupConfigBuildsParsingError
	}
	return nil
}

// validateExistingDurationForEditing 检查编辑的时间是否有时间冲突
func (s *Service) validateExistingDurationForEditing(popupConfig *model.PopupConfig) (err error) {
	if popupConfig == nil {
		err = xecode.NothingFound
		return
	} else if (popupConfig.ID == 0 && (popupConfig.STime == 0 || popupConfig.ETime == 0)) || popupConfig.STime > popupConfig.ETime {
		err = ecode.PopupConfigParameterError
		return
	}
	var conflictConfigList []*model.PopupConfig
	if conflictConfigList, err = s.dao.GetPopupConfigWithConflictDuration(popupConfig); err != nil {
		return
	}
	//nolint:gosimple
	if conflictConfigList != nil && len(conflictConfigList) > 0 {
		err = ecode.PopupConfigConflictTime
		return
	}
	return
}

//nolint:unused
func (s *Service) pushCrowdPackageAndGetID(popupConfig *model.PopupConfig) (packageID int64, err error) {
	return
}
