package modules

import (
	"context"
	"fmt"

	"go-common/library/database/sql"

	mdlmdl "go-gateway/app/app-svr/fawkes/service/model/modules"
	log "go-gateway/app/app-svr/fawkes/service/tools/logger"
)

// AddModuleSize add module size
func (s *Service) AddModuleSize(c context.Context, req *mdlmdl.ModuleSizeReq) (err error) {
	var (
		moduleID, gitJobId int64
	)
	// 通过平台主键id. 获取gl_job_id
	if gitJobId, err = s.fkDao.GetPackBuildID(context.Background(), req.BuildID); err != nil {
		log.Error("AddModuleSize GetPackBuildID err: %v", err)
		return
	}
	moduleMap := make(map[string]int64)   // { module: module_id }
	midGroupMap := make(map[int64]string) // { module_id: group_name }
	for _, item := range req.Meta {
		if moduleID = moduleMap[item.Name]; moduleID == 0 {
			if moduleID, err = s.fkDao.GetModuleID(context.Background(), req.AppKey, item.Name); err != nil {
				if err == sql.ErrNoRows {
					err = nil
					if moduleID, err = s.addModule(req.AppKey, item.Name); err != nil {
						log.Error("AddModuleSize addModule err: %v", err)
						continue
					}
				} else {
					log.Error("AddModuleSize GetModuleID err: %v", err)
					continue
				}
			}
			moduleMap[item.Name] = moduleID
			// 如果组名为空. 则不会进入自动分组逻辑
			if item.Group != "" {
				midGroupMap[moduleID] = item.Group
			}
		}
		if err = s.addModuleSize(req.AppKey, gitJobId, moduleID, item.LibVer, item.SizeType, item.Size); err != nil {
			log.Error("%v", err)
		}
	}
	// 自动分组 ( 新版本. 旧版本逐步废弃 )
	if req.IsAutoGroup == 1 {
		_ = s.autoGroupV2(c, req.AppKey, midGroupMap)
	} else {
		_ = s.autoGroup(c, req.AppKey, midGroupMap)
	}
	return
}

func (s *Service) addModule(appKey, name string) (moduleID int64, err error) {
	var (
		tx *sql.Tx
	)
	if tx, err = s.fkDao.BeginTran(context.Background()); err != nil {
		log.Error("s.fkDao.BeginTran() error(%v)", err)
		return
	}
	defer func() {
		if r := recover(); r != nil {
			//nolint:errcheck
			tx.Rollback()
			log.Error("%v", r)
		}
		if err != nil {
			if err1 := tx.Rollback(); err1 != nil {
				log.Error("tx.Rollback() error(%v)", err1)
			}
			return
		}
		if err = tx.Commit(); err != nil {
			log.Error("tx.Commit() error(%v)", err)
		}
	}()
	if moduleID, err = s.fkDao.TxAddModule(tx, appKey, name); err != nil {
		log.Error("TxAddModule %v", err)
	}
	return
}

func (s *Service) addModuleSize(appKey string, buildID int64, moduleID int64, libVer, sizeType string, size int64) (err error) {
	var (
		tx *sql.Tx
	)
	if tx, err = s.fkDao.BeginTran(context.Background()); err != nil {
		log.Error("s.fkDao.BeginTran() error(%v)", err)
		return
	}
	defer func() {
		if r := recover(); r != nil {
			//nolint:errcheck
			tx.Rollback()
			log.Error("%v", r)
		}
		if err != nil {
			if err1 := tx.Rollback(); err1 != nil {
				log.Error("tx.Rollback() error(%v)", err1)
			}
			return
		}
		if err = tx.Commit(); err != nil {
			log.Error("tx.Commit() error(%v)", err)
		}
	}()
	if _, err = s.fkDao.TxAddModuleSize(tx, appKey, buildID, moduleID, libVer, sizeType, size); err != nil {
		log.Error("%v", err)
	}
	return
}

// AddGroup add a new group
func (s *Service) AddGroup(c context.Context, appKey, name, cname string) (err error) {
	var (
		tx *sql.Tx
	)
	if tx, err = s.fkDao.BeginTran(context.Background()); err != nil {
		log.Error("s.dao.BeginTran() error(%v)", err)
		return
	}
	defer func() {
		if r := recover(); r != nil {
			//nolint:errcheck
			tx.Rollback()
			log.Error("%v", r)
		}
	}()
	if _, err = s.fkDao.TxAddGroup(tx, appKey, name, cname); err != nil {
		//nolint:errcheck
		tx.Rollback()
		log.Error("%v", err)
		return
	}
	if err = tx.Commit(); err != nil {
		log.Error("tx.Commit error(%v)", err)
	}
	return nil
}

// ChangeGroup change moudle to a group
func (s *Service) ChangeGroup(c context.Context, appKey, mName, gName, operator string) (err error) {
	var (
		tx                *sql.Tx
		moduleID, groupID int64
		relationExist     int
	)
	if tx, err = s.fkDao.BeginTran(context.Background()); err != nil {
		log.Error("s.dao.BeginTran() error(%v)", err)
		return
	}
	defer func() {
		if r := recover(); r != nil {
			//nolint:errcheck
			tx.Rollback()
			log.Error("%v", r)
		}
	}()
	if moduleID, err = s.fkDao.GetModuleID(c, appKey, mName); err != nil {
		log.Error("%v", err)
		return
	}
	if groupID, _, err = s.fkDao.GetGroupID(c, appKey, gName); err != nil {
		log.Error("%v", err)
		return
	}
	if relationExist, err = s.fkDao.ExistsRelation(c, moduleID); err != nil {
		log.Error("%v", err)
		return
	}
	if relationExist > 0 {
		if _, err = s.fkDao.TxSetModuleGroupRalation(tx, moduleID, groupID, operator); err != nil {
			//nolint:errcheck
			tx.Rollback()
			log.Error("%v", err)
			return
		}
	} else {
		if _, err = s.fkDao.TxSetModuleGroupRalation(tx, moduleID, groupID, operator); err != nil {
			//nolint:errcheck
			tx.Rollback()
			log.Error("%v", err)
			return
		}
	}
	if err = tx.Commit(); err != nil {
		log.Error("tx.Commit error(%v)", err)
	}
	return nil
}

// autoGroup auto group module
func (s *Service) autoGroup(c context.Context, appKey string, midGroupMap map[int64]string) (err error) {
	var (
		tx           *sql.Tx
		mdlIDs       []int64
		extMdlIDs    []int64
		notExtMdlIDs []int64
	)
	if tx, err = s.fkDao.BeginTran(context.Background()); err != nil {
		log.Error("s.dao.BeginTran() error(%v)", err)
		return
	}
	defer func() {
		if r := recover(); r != nil {
			//nolint:errcheck
			tx.Rollback()
			log.Error("%v", r)
		}
	}()
	for mdlID := range midGroupMap {
		mdlIDs = append(mdlIDs, mdlID)
	}
	// 查询已经存在分组的module_id 集合
	if extMdlIDs, err = s.fkDao.ExistsRelationModuleIDs(c, mdlIDs); err != nil {
		log.Error("%v", err)
		return
	}
	// 判断如果和当前上传的module_id集合不相等，则筛选出不存在分组关系的 module_ids: notExtMdlIDs
	if len(extMdlIDs) != len(mdlIDs) {
		for _, mID := range mdlIDs {
			isNotExt := true
			for _, emID := range extMdlIDs {
				if mID == emID {
					isNotExt = false
					break
				}
			}
			if isNotExt {
				notExtMdlIDs = append(notExtMdlIDs, mID)
			}
		}
		// 缓存分组信息 groupMap { group_name: group_id}
		groupMap := make(map[string]int64)
		if len(notExtMdlIDs) != 0 {
			var groupList []*mdlmdl.Group
			if groupList, err = s.fkDao.ListAllGroups(c, appKey); err != nil {
				log.Error("%v", err)
				return
			}
			for _, groupInfo := range groupList {
				groupMap[groupInfo.GName] = groupInfo.GID
			}
		}
		// 将不存在分组关系的module_id 关联上分组：midGroupMap{ mid: group_name }
		for _, nemID := range notExtMdlIDs {
			groupID := groupMap[midGroupMap[nemID]]
			if groupID != 0 {
				if _, err = s.fkDao.TxSetModuleGroupRalation(tx, nemID, groupID, "git"); err != nil {
					log.Error("%v", err)
					return
				}
			}
		}
	}
	if err = tx.Commit(); err != nil {
		log.Error("tx.Commit error(%v)", err)
	}
	return nil
}

// autoGroupV2 auto group module
func (s *Service) autoGroupV2(c context.Context, appKey string, midGroupMap map[int64]string) (err error) {
	var (
		existsGroups []*mdlmdl.Group
		groupMap     = make(map[string]int64)
	)
	// 自动添加分组
	if existsGroups, err = s.fkDao.ListAllGroups(c, appKey); err != nil {
		log.Error("%v", err)
		return
	}
	tmp := make(map[string]int64)
	for mid := range midGroupMap {
		gname := midGroupMap[mid]
		tmp[gname] = 0
	}
	for gname := range tmp {
		var (
			isExists = false
		)
		for _, g := range existsGroups {
			groupMap[g.GName] = g.GID
			if g.GName == gname {
				isExists = true
				break
			}
		}
		if isExists == false {
			// 添加group
			_ = s.fkDao.Transact(c, func(tx *sql.Tx) error {
				gid, err := s.fkDao.TxAddGroup(tx, appKey, gname, gname)
				log.Info(fmt.Sprintf("auto group - TxAddGroup: appkey:%s, gname:%v, gname:%v", appKey, gname, gname))
				if err == nil {
					groupMap[gname] = gid
				}
				return err
			})
		}
	}
	// 重组关联关系
	for mid := range midGroupMap {
		gname := midGroupMap[mid]
		gid := groupMap[gname]
		_ = s.fkDao.Transact(c, func(tx *sql.Tx) error {
			_, err = s.fkDao.TxSetModuleGroupRalation(tx, mid, gid, "git")
			log.Info(fmt.Sprintf("auto group - TxSetModuleGroupRalation: appkey:%v, mid:, %v, gid:%v, gname:%v", appKey, mid, gid, gname))
			return err
		})
	}
	return nil
}

// EditGroup Edit group's info
func (s *Service) EditGroup(c context.Context, gID int64, name, cname string) (err error) {
	var (
		tx *sql.Tx
	)
	if tx, err = s.fkDao.BeginTran(context.Background()); err != nil {
		log.Error("s.dao.BeginTran() error(%v)", err)
		return
	}
	defer func() {
		if r := recover(); r != nil {
			//nolint:errcheck
			tx.Rollback()
			log.Error("%v", r)
		}
	}()
	if _, err = s.fkDao.TxEditGroup(tx, gID, name, cname); err != nil {
		//nolint:errcheck
		tx.Rollback()
		log.Error("%v", err)
		return
	}
	if err = tx.Commit(); err != nil {
		log.Error("tx.Commit error(%v)", err)
	}
	return nil
}

// DeleteGroup Delete a group
func (s *Service) DeleteGroup(c context.Context, gID int64) (err error) {
	var (
		tx *sql.Tx
	)
	if tx, err = s.fkDao.BeginTran(context.Background()); err != nil {
		log.Error("s.dao.BeginTran() error(%v)", err)
		return
	}
	defer func() {
		if r := recover(); r != nil {
			//nolint:errcheck
			tx.Rollback()
			log.Error("%v", r)
		}
	}()
	if _, err = s.fkDao.TxDelGroupModuleRelation(tx, gID); err != nil {
		//nolint:errcheck
		tx.Rollback()
		log.Error("%v", err)
		return
	}
	if _, err = s.fkDao.TxDelGroup(tx, gID); err != nil {
		//nolint:errcheck
		tx.Rollback()
		log.Error("%v", err)
		return
	}
	if err = tx.Commit(); err != nil {
		log.Error("tx.Commit error(%v)", err)
	}
	return nil
}

// ListModuleGroup list modules & their groups
func (s *Service) ListModuleGroup(c context.Context, appKey string) (res []*mdlmdl.Group, err error) {
	var (
		modules, unGroupedModules []*mdlmdl.Module
		emptyGroups               []*mdlmdl.Group
	)
	// 已分组 modules
	if modules, err = s.fkDao.ListModuleGroup(c, appKey); err != nil {
		log.Error("%v", err)
		return
	}
	groupMap := make(map[string][]*mdlmdl.Module)
	for _, module := range modules {
		groupMap[module.GName] = append(groupMap[module.GName], module)
	}
	for _, module := range groupMap {
		group := &mdlmdl.Group{
			GID:     module[0].GID,
			GName:   module[0].GName,
			GCName:  module[0].GCName,
			Modules: module,
		}
		res = append(res, group)
	}
	// 未分组 modules
	if unGroupedModules, err = s.fkDao.ListModuleUngroup(c, appKey); err != nil {
		if err != sql.ErrNoRows {
			log.Error("%v", err)
			return
		}
	}
	if len(unGroupedModules) > 0 {
		group := &mdlmdl.Group{
			GID:     0,
			GName:   "ungrouped",
			GCName:  "未分组",
			Modules: unGroupedModules,
		}
		res = append(res, group)
	}
	// 无 modules 的空组
	if emptyGroups, err = s.fkDao.ListEmptyGroups(c, appKey); err != nil {
		if err != sql.ErrNoRows {
			log.Error("%v", err)
			return
		}
	}
	if len(emptyGroups) > 0 {
		res = append(res, emptyGroups...)
	}
	return
}

// ListModuleSize list size of a module
func (s *Service) ListModuleSize(c context.Context, appKey, mName, sizeType string) (res mdlmdl.ModuleSizeRes, err error) {
	var (
		moduleSize []*mdlmdl.ModuleSize
	)
	if moduleSize, err = s.fkDao.ListModuleSize(c, appKey, mName, sizeType); err != nil {
		log.Error("%v", err)
		return
	}
	res.MID = moduleSize[0].ID
	res.MName = moduleSize[0].Name
	res.MCName = moduleSize[0].CName
	if sizeType == "" {
		res.SizeType = "sum"
	} else {
		res.SizeType = sizeType
	}
	res.Meta = append(res.Meta, moduleSize...)
	return
}

// ListGroupSize list size of a group
func (s *Service) ListGroupSize(c context.Context, appKey, gName, sizeType string, resRatio, codeRatio, xcassetsRatio float64, limit int) (res mdlmdl.GroupSizeRes, err error) {
	var (
		groupSize []*mdlmdl.GroupSize
		groupID   int64
		gCName    string
		verCodes  []int64
	)
	if verCodes, err = s.fkDao.LatestVersions(c, appKey, limit); err != nil {
		log.Error("%v", err)
		return
	}
	if groupSize, err = s.fkDao.ListGroupSize(c, appKey, gName, sizeType, verCodes, resRatio, codeRatio, xcassetsRatio); err != nil {
		log.Error("%v", err)
		return
	}
	if groupID, gCName, err = s.fkDao.GetGroupID(c, appKey, gName); err != nil {
		log.Error("%v", err)
		return
	}
	res.GID = groupID
	res.GName = gName
	res.GCName = gCName
	if sizeType == "" {
		res.SizeType = "sum"
	} else {
		res.SizeType = sizeType
	}
	res.Meta = append(res.Meta, groupSize...)
	return
}

// ListAllGroups list all groups
func (s *Service) ListAllGroups(c context.Context, appKey string) (res []*mdlmdl.Group, err error) {
	if res, err = s.fkDao.ListAllGroups(c, appKey); err != nil {
		log.Error("%v", err)
		return
	}
	return
}

// ListModuleSizeInGroup list size of modules in one group in specified version
func (s *Service) ListModuleSizeInGroup(c context.Context, appKey, groupName, sizeType string, buildID int64, resRatio, codeRatio, xcassetsRatio float64) (res []*mdlmdl.ModuleGroupSize, err error) {
	if res, err = s.fkDao.ListModuleSizeInGroup(c, appKey, groupName, sizeType, buildID, resRatio, codeRatio, xcassetsRatio); err != nil {
		log.Error("%v", err)
		return
	}
	return
}

// ListGroupSizeInBuild list groups' size in a build
func (s *Service) ListGroupSizeInBuild(c context.Context, appKey, sizeType string, buildID int64, resRatio, codeRatio, xcassetsRatio float64) (res []*mdlmdl.GroupSizeInBuildRes, err error) {
	if res, err = s.fkDao.ListGroupSizeInBuild(c, appKey, sizeType, buildID, resRatio, codeRatio, xcassetsRatio); err != nil {
		log.Error("%v", err)
		return
	}
	return
}

// ListSizeTypes list all size types of an app
func (s *Service) ListSizeTypes(c context.Context, appKey string, limit int) (res []string, err error) {
	var verCodes []int64
	if verCodes, err = s.fkDao.LatestVersions(c, appKey, limit); err != nil {
		log.Error("%v", err)
		return
	}
	if len(verCodes) == 0 {
		return
	}
	if res, err = s.fkDao.ListSizeTypes(c, appKey, verCodes); err != nil {
		log.Error("%v", err)
		return
	}
	return
}

// ModulesConfTotalSizeSet set module config totalsize
func (s *Service) ModulesConfTotalSizeSet(c context.Context, appKey, version, operator string, moduleGroupIDList []int64, totalSize int64) (err error) {
	var tx *sql.Tx
	if tx, err = s.fkDao.BeginTran(c); err != nil {
		log.Error("s.fkDao.modulesConfSet() error(%v)", err)
		return
	}
	defer func() {
		if r := recover(); r != nil {
			//nolint:errcheck
			tx.Rollback()
			log.Error("%v", r)
		}
		if err != nil {
			if err1 := tx.Rollback(); err1 != nil {
				log.Error("tx.Rollback() error(%v)", err1)
			}
			return
		}
		if err = tx.Commit(); err != nil {
			log.Error("tx.Commit() error(%v)", err)
		}
	}()
	err = s.fkDao.TxModulesConfTotalSizeSet(tx, appKey, version, operator, moduleGroupIDList, totalSize)
	return
}

// ModulesConfSet set module config
func (s *Service) ModulesConfSet(c context.Context, appKey, version, description, operator string, percentage float64, moduleGroupID, totalSize, fixedSize, applyNormalSize, applyForceSize, externalSize int64) (err error) {
	var tx *sql.Tx
	if tx, err = s.fkDao.BeginTran(c); err != nil {
		log.Error("s.fkDao.modulesConfSet() error(%v)", err)
		return
	}
	defer func() {
		if r := recover(); r != nil {
			//nolint:errcheck
			tx.Rollback()
			log.Error("%v", r)
		}
		if err != nil {
			if err1 := tx.Rollback(); err1 != nil {
				log.Error("tx.Rollback() error(%v)", err1)
			}
			return
		}
		if err = tx.Commit(); err != nil {
			log.Error("tx.Commit() error(%v)", err)
		}
	}()
	err = s.fkDao.TxModulesConfSet(tx, appKey, version, description, operator, percentage, moduleGroupID, totalSize, fixedSize, applyNormalSize, applyForceSize, externalSize)
	return
}

// ModulesConfGet get module config
func (s *Service) ModulesConfGet(c context.Context, appKey, version string, getNewest bool) (res []*mdlmdl.ModuleConfig, err error) {
	if res, err = s.fkDao.GetModulesConf(c, appKey, version); err != nil {
		log.Error("%v", err)
		return
	}
	// 当前配置不存在时若需要返回最新配置则取最新配置返回
	if len(res) == 0 && getNewest {
		if version, err = s.fkDao.GetNewestModulesConfVersion(c, appKey); err != nil {
			log.Error("%v", err)
			return
		}
		if version != "" {
			if res, err = s.fkDao.GetModulesConf(c, appKey, version); err != nil {
				log.Error("%v", err)
			}
		}
	}
	return
}

// ModulesConfCopy copy newest config
func (s *Service) ModulesConfCopy(c context.Context, appKey, version string) (err error) {
	var (
		moduleConfigs, preConfigs []*mdlmdl.ModuleConfig
		preVersion                string
	)
	// 检测需要同步配置的版本是否存在 如果存在不做处理
	if moduleConfigs, err = s.fkDao.GetModulesConf(c, appKey, version); err != nil {
		log.Error("%v", err)
		return
	}
	if len(moduleConfigs) > 0 {
		log.Info("ModulesConfCopy config already exsit appKey %s version %s", appKey, version)
		return
	}
	// 确认最近版本
	if preVersion, err = s.fkDao.GetPreciousVersion(c, appKey, version); err != nil {
		log.Error("s.fkDao.GetPreciousVersion() error(%v)", err)
	}
	if preVersion == "" {
		log.Info("ModulesConfCopy preVersion not exsit appKey %s version %s preversion %s", appKey, version, preVersion)
		return
	}
	// todo 如果最近版本存在配置则同步配置
	if preConfigs, err = s.fkDao.GetModulesConf(c, appKey, preVersion); err != nil {
		log.Error("%v", err)
		return
	}
	if len(preConfigs) == 0 {
		log.Info("ModulesConfCopy preVersion config not exsit appKey %s version %s preversion %s", appKey, version, preVersion)
		return
	}
	var tx *sql.Tx
	if tx, err = s.fkDao.BeginTran(c); err != nil {
		log.Error("s.fkDao.modulesConfSet() error(%v)", err)
		return
	}
	defer func() {
		if r := recover(); r != nil {
			//nolint:errcheck
			tx.Rollback()
			log.Error("%v", r)
		}
		if err != nil {
			if err1 := tx.Rollback(); err1 != nil {
				log.Error("tx.Rollback() error(%v)", err1)
			}
			return
		}
		if err = tx.Commit(); err != nil {
			log.Error("tx.Commit() error(%v)", err)
		}
	}()
	for _, preConfig := range preConfigs {
		if err = s.fkDao.TxModulesConfSet(tx, preConfig.AppKey, version, preConfig.Description, preConfig.OPERATOR, preConfig.Percentage, preConfig.ModuleGroupID, preConfig.TotalSize, preConfig.FixedSize, preConfig.ApplyNormalSize, preConfig.ApplyForceSize, preConfig.ExternalSize); err != nil {
			log.Error("s.fkDao.TxModulesConfSet() error(%v)", err)
		}
	}
	return
}
