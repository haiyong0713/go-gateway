package business

import (
	"bytes"
	"context"
	"crypto/md5"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"os"
	"path"
	"sort"
	"strconv"
	"strings"
	"time"

	"go-common/library/database/sql"
	"go-common/library/ecode"
	"go-common/library/log"

	"go-common/library/sync/errgroup.v2"

	"go-gateway/app/app-svr/fawkes/service/conf"
	"go-gateway/app/app-svr/fawkes/service/model"
	appmdl "go-gateway/app/app-svr/fawkes/service/model/app"
	bizapkmdl "go-gateway/app/app-svr/fawkes/service/model/bizapk"
	busmdl "go-gateway/app/app-svr/fawkes/service/model/business"
	cdmdl "go-gateway/app/app-svr/fawkes/service/model/cd"
	"go-gateway/app/app-svr/fawkes/service/model/mod"
	"go-gateway/app/app-svr/fawkes/service/model/pcdn"
	toolmdl "go-gateway/app/app-svr/fawkes/service/model/tool"
	tribemdl "go-gateway/app/app-svr/fawkes/service/model/tribe"
	"go-gateway/app/app-svr/fawkes/service/tools/utils"
)

// NewestVersion get config version and ff version.
func (s *Service) NewestVersion(c context.Context) (res map[string]map[string]*busmdl.Version, err error) {
	res = make(map[string]map[string]*busmdl.Version)
	for env, av := range s.configVersionCache {
		var (
			re map[string]*busmdl.Version
			ok bool
		)
		if re, ok = res[env]; !ok {
			re = make(map[string]*busmdl.Version)
			res[env] = re
		}
		for appKey, version := range av {
			r := &busmdl.Version{
				Config: version,
			}
			re[appKey] = r
		}
	}
	for env, fv := range s.ffVersionCache {
		var (
			re map[string]*busmdl.Version
			ok bool
		)
		if re, ok = res[env]; !ok {
			re = make(map[string]*busmdl.Version)
			res[env] = re
		}
		for appKey, version := range fv {
			var (
				r  *busmdl.Version
				ok bool
			)
			if r, ok = re[appKey]; !ok {
				r = &busmdl.Version{}
				re[appKey] = r
			}
			r.FF = version
		}
	}
	return
}

// VersionAll get all version
func (s *Service) VersionAll(c context.Context) (res map[string]map[int64]*model.Version, err error) {
	res = s.versionAllCache
	return
}

// UpgradeAll get all version
func (s *Service) UpgradeAll(c context.Context) (res map[string]map[int64]*cdmdl.UpgradConfig, err error) {
	res = s.upgradConfigAllCache
	return
}

// PackAll get all pack
func (s *Service) PackAll(c context.Context) (res map[string]map[int64][]*cdmdl.Pack, err error) {
	res = s.packAllCache
	return
}

// PackLatestStable get latest stable pack
func (s *Service) PackLatestStable(c context.Context, appKey string, versionCode int) (res *cdmdl.Pack, err error) {
	if res, err = s.fkDao.PackLatestStable(c, appKey, versionCode); err != nil {
		log.Error("%v", err)
	}
	return
}

// FilterAll get all filter
func (s *Service) FilterAll(c context.Context) (res map[string]map[int64]*cdmdl.FilterConfig, err error) {
	if res, err = s.fkDao.FilterConfigAll(c); err != nil {
		log.Error("%v", err)
	}
	return
}

// PatchAll get all patch by cache
func (s *Service) PatchAllCache(c context.Context) (res map[string]map[string]*cdmdl.Patch, err error) {
	res = s.patchAllCache
	return
}

// PatchAll get all patch
func (s *Service) PatchAll(c context.Context) (res map[string]map[string]*cdmdl.Patch, err error) {
	var (
		appKeys []string
	)
	if appKeys, err = s.fkDao.PatchAppKeys(c); err != nil {
		log.Error("%v", err)
		return
	}
	// 每个appKey分别获取10个prod和test的pack包的buildId
	var buildIds []int64
	for _, appKey := range appKeys {
		//取最近十个版本
		var (
			versionIdsTest, versionIdsProd, buildIdsTest, buildIdsProd []int64
		)
		if versionIdsTest, err = s.fkDao.LastPackVersionIds(c, appKey, "test"); err != nil {
			log.Error("%v", err)
			return
		}
		if len(versionIdsTest) > 0 {
			if buildIdsTest, err = s.fkDao.PackBuildIdsByVersions(c, appKey, "test", versionIdsTest); err != nil {
				log.Error("%v", err)
				return
			}
			buildIds = append(buildIds, buildIdsTest...)
		}
		if versionIdsProd, err = s.fkDao.LastPackVersionIds(c, appKey, "prod"); err != nil {
			log.Error("%v", err)
			return
		}
		if len(versionIdsProd) > 0 {
			if buildIdsProd, err = s.fkDao.PackBuildIdsByVersions(c, appKey, "prod", versionIdsProd); err != nil {
				log.Error("%v", err)
				return
			}
			buildIds = append(buildIds, buildIdsProd...)
		}
	}
	if len(buildIds) > 0 {
		if res, err = s.fkDao.PatchAll4(c, buildIds); err != nil {
			log.Error("%v", err)
		}
	}
	return
}

// ChannelAll get all channel
func (s *Service) ChannelAll(c context.Context) (res map[string]map[int64]*appmdl.Channel, err error) {
	if res, err = s.fkDao.AppChannelAll(c); err != nil {
		log.Error("%v", err)
	}
	return
}

// FlowAll get all flow
func (s *Service) FlowAll(c context.Context) (res map[string]map[int64]*cdmdl.FlowConfig, err error) {
	res = s.flowConfigAllCache
	return
}

func (s *Service) HotfixAllCache(c context.Context) (res map[string]map[int64][]*appmdl.HfUpgrade, err error) {
	res = s.hotfixAllCache
	return
}

// HotfixAll get all hotfix upgrade information
func (s *Service) HotfixAll(c context.Context) (res map[string]map[int64][]*appmdl.HfUpgrade, err error) {
	var (
		config map[string]map[int64]*appmdl.HotfixConfig
	)
	if res, err = s.fkDao.HotfixAll(c); err != nil {
		log.Error("s.fkDao.HotfixAll() failed. %v", err)
		return
	}
	if config, err = s.fkDao.HotfixConfigAll(c); err != nil {
		log.Error("s.fkDao.HotfixConfigAll() failed. %v", err)
		return
	}
	for key, hotfixs := range res {
		for _, items := range hotfixs {
			for _, item := range items {
				var (
					conf *appmdl.HotfixConfig
					ok   bool
				)
				if conf, ok = config[key][item.BuildID]; !ok {
					continue
				}
				item.Config = conf
			}
		}
	}
	return
}

// LaserAll get all laser.
func (s *Service) LaserAll(c context.Context) (res []*appmdl.Laser, err error) {
	if res, err = s.fkDao.LaserAll(c); err != nil {
		log.Error("%v", err)
	}
	return
}

// Laser get laser info.
func (s *Service) Laser(c context.Context, taskID int64) (res *appmdl.Laser, err error) {
	if res, err = s.fkDao.Laser(c, taskID); err != nil {
		log.Error("%v", err)
	}
	return
}

// LaserAll get all laser.
func (s *Service) LaserAllSilence(c context.Context) (res []*appmdl.Laser, err error) {
	if res, err = s.fkDao.LaserAllSilence(c); err != nil {
		log.Error("%v", err)
	}
	return
}

// BizApkListAll is.
func (s *Service) BizApkListAll(c context.Context) (map[int64]map[string]map[string][]*bizapkmdl.Apk, error) {
	list, err := s.fkDao.BizApkListAll(c)
	if err != nil {
		log.Error("%v", err)
		return nil, err
	}
	filter, err := s.fkDao.BizApkFilterAll(c)
	if err != nil {
		log.Error("%v", err)
		return nil, err
	}
	flow, err := s.fkDao.BizApkFlowAll(c)
	if err != nil {
		log.Error("%v", err)
		return nil, err
	}
	res := map[int64]map[string]map[string][]*bizapkmdl.Apk{}
	for _, v := range list {
		for _, fi := range filter {
			if v.BuildID == fi.BuildID && v.Env == fi.Env {
				v.FilterConfig = fi
			}
		}
		for _, fl := range flow {
			if v.BuildID == fl.BuildID && v.Env == fl.Env {
				v.FlowConfig = fl
			}
		}
		mm, ok := res[v.PackBuildID]
		if !ok {
			mm = map[string]map[string][]*bizapkmdl.Apk{}
			res[v.PackBuildID] = mm
		}
		m, ok := mm[v.Env]
		if !ok {
			m = map[string][]*bizapkmdl.Apk{}
			mm[v.Env] = m
		}
		m[v.Name] = append(m[v.Name], v)
	}
	return res, nil
}

func (s *Service) BizApkListAllCache(c context.Context) (map[int64]map[string]map[string][]*bizapkmdl.Apk, error) {
	return s.bizApkListAllCache, nil
}

// TribeListAll {app_key: {build_id: {env: tribeName: []}}}
// nolint:gocognit
func (s *Service) TribeListAll(c context.Context) (tribePacks map[string]map[int64]map[string]map[string][]*tribemdl.TribeApk, tribeHostRelation map[int64]int64, err error) {
	res := map[string]map[int64]map[string]map[string][]*tribemdl.TribeApk{}
	tribeList, err := s.fkDao.Tribes(c)
	if err != nil {
		log.Error("%v", err)
		return nil, nil, err
	}
	// 获取宿主app依赖关系表
	hostRelation, err := s.fkDao.TribeHostRelationAll(c)
	if err != nil {
		log.Error("%v", err)
		return nil, nil, err
	}
	// 将宿主依赖关系表 映射成 当前 [当前app构建号] ： [前置app构建号]
	hostRelationMap := map[int64]int64{}
	// hostRelationFeatureMap 确定唯一feature
	hostRelationFeatureMap := map[string][]string{}
	for _, rel := range hostRelation {
		hostRelationMap[rel.CurrentBuildID] = rel.ParentBuildID
		fmKey := fmt.Sprintf("%d_%d", rel.CurrentBuildID, rel.ParentBuildID)
		hostRelationFeatureMap[fmKey] = append(hostRelationFeatureMap[fmKey], rel.Feature)
	}
	tribeMap := map[string][]*tribemdl.Tribe{}
	for _, trb := range tribeList {
		_, ok := tribeMap[trb.AppKey]
		if !ok {
			t := []*tribemdl.Tribe{}
			tribeMap[trb.AppKey] = t
		}
		tribeMap[trb.AppKey] = append(tribeMap[trb.AppKey], trb)
		// res 一级结构 app_key
		app, ok := res[trb.AppKey]
		if !ok {
			app = map[int64]map[string]map[string][]*tribemdl.TribeApk{}
			res[trb.AppKey] = app
		}
		for _, relation := range hostRelation {
			if relation.AppKey == trb.AppKey {
				// 二级结构 宿主的job_id
				_, cok := app[relation.CurrentBuildID]
				if !cok {
					app[relation.CurrentBuildID] = map[string]map[string][]*tribemdl.TribeApk{}
				}
				_, pok := app[relation.ParentBuildID]
				if !pok {
					app[relation.CurrentBuildID] = map[string]map[string][]*tribemdl.TribeApk{}
				}
			}
		}
	}
	// 获取所有生效的tribe_pack的数据
	tribePacklist, err := s.fkDao.TribePackListAll(c)
	if err != nil {
		log.Error("%v", err)
		return nil, nil, err
	}
	// 获取所有tribe pack 升级配置的数据
	tribePackfilter, err := s.fkDao.TribePackFilterAll(c)
	if err != nil {
		log.Error("%v", err)
		return nil, nil, err
	}
	// 获取所有tribe pack 配置的版本信息数据
	tribePackUpgrade, err := s.fkDao.TribePackUpgradeAll(c)
	if err != nil {
		log.Error("%v", err)
		return nil, nil, err
	}
	// 标记所有的环境变量，用于补全缺少的部分
	allEnvs := map[string]bool{}
	// 遍历pack数据， 将升级配置 和 版本配置关联到pack数据上
	for _, v := range tribePacklist {
		allEnvs[v.Env] = true
		// 关联上组件pack 的 filter
		for _, fi := range tribePackfilter {
			if v.ID == fi.TribePackId && v.Env == fi.Env {
				v.FilterConfig = fi
			}
		}
		// 关联pack 版本信息
		for _, upg := range tribePackUpgrade {
			if v.ID == upg.TribePackId && v.Env == upg.Env {
				v.UpgradeConfig = upg
			}
		}
		// 通用组件 没有宿主 hostjobid, 默认为-1
		if v.Nohost {
			v.TribeHostJobID = -1
		}
		// 一级结构 app_key
		app, ok := res[v.AppKey]
		if !ok {
			app = map[int64]map[string]map[string][]*tribemdl.TribeApk{}
			res[v.AppKey] = app
		}
		// 二级结构 宿主的job_id
		mm, ok := app[v.TribeHostJobID]
		if !ok {
			mm = map[string]map[string][]*tribemdl.TribeApk{}
			app[v.TribeHostJobID] = mm
		}
		m, ok := mm[v.Env]
		if !ok {
			m = map[string][]*tribemdl.TribeApk{}
			mm[v.Env] = m
		}
		m[v.Name] = append(m[v.Name], v)
	}
	// {
	// 	"android": {
	// 		6201009: {
	// 			"prod": {
	// 				"liveStream": [{xxx}]
	// 			}
	// 		}
	// 	}
	// }
	for appKey, jobIDs := range res {
		for jobID, envs := range jobIDs {
			if jobID == -1 {
				continue
			}
			// 补全缺少的环境变量，用于寻找兼容
			for aenv := range allEnvs {
				if _, ok := envs[aenv]; !ok {
					envs[aenv] = map[string][]*tribemdl.TribeApk{}
				}
			}
			for env, tribeNames := range envs {
				for _, trb := range tribeMap[appKey] {
					// 记录初始兼容的features, 递归查找的时候，以此为标准，对比feature是否匹配
					var initFeatures []string
					depJobID, hok := hostRelationMap[jobID]
					if hok {
						fmKey := fmt.Sprintf("%d_%d", jobID, depJobID)
						initFeatures = hostRelationFeatureMap[fmKey]
					}
					_, ok := tribeNames[trb.Name]
					if !ok {
						// 该版本下缺失的组件，需要去relation表中找到兼容的组件
						// log.Info("[TribeListAll]-版本号(%v)下缺失的组件(%v)", jobID, trb.Name)
						tribes := s.TribeIsUseable(res[appKey], hostRelationMap, jobID, jobID, env, trb.Name, hostRelationFeatureMap, initFeatures)
						// log.Info("[TribeListAll]-版本号(%v)兼容表中找到缺失组件(%v)(%+v), 共(%v)条", jobID, trb.Name, tribes, len(tribes))
						if len(tribes) != 0 {
							res[appKey][jobID][env][trb.Name] = tribes
							// log.Info("[TribeListAll]-res组件%+v", res[appKey][jobID][env][trb.Name])
						}
					} else {
						// 查看该版本下的版本是否可用
						useableTribes := s.TribeVersionUseable(jobID, tribeNames[trb.Name], hostRelationMap, nil)
						if len(useableTribes) == 0 {
							tribes := s.TribeIsUseable(res[appKey], hostRelationMap, jobID, jobID, env, trb.Name, hostRelationFeatureMap, initFeatures)
							// log.Info("[TribeListAll]-当前构建号(%v)存在组件，但所有组件配置在当前构建号下不可用，需要向上查找兼容版本：%+v, 长度%v", jobID, tribes, len(tribes))
							res[appKey][jobID][env][trb.Name] = tribes
						} else {
							res[appKey][jobID][env][trb.Name] = useableTribes
						}
					}
				}
			}
		}
	}
	// log.Warn("[TribeListAll]：%+v", res)
	return res, hostRelationMap, nil
}
func (s *Service) TribeIsUseable(res map[int64]map[string]map[string][]*tribemdl.TribeApk, relationMap map[int64]int64, curJobID, thJobID int64, env, tribeName string, hostRelationFeatureMap map[string][]string, initFeatures []string) (resTribePack []*tribemdl.TribeApk) {
	// log.Info("TribeIsUseable-(组件是否可用): curJobID(%v),thJobID(%v),env(%v),组件名(%v),兼容的Feature(%+v),兼容版本(%+v),组件结构(%+v)", curJobID, thJobID, env, tribeName, hostRelationFeatureMap, relationMap, res)
	depJobID, ok := relationMap[thJobID]

	if !ok {
		// log.Info("TribeIsUseable-(组件是否可用): 不存在兼容版本")
		return
	}
	fmKey := fmt.Sprintf("%d_%d", thJobID, depJobID)
	// intersectionFeature feature 交集， 用于确定兼容关系链
	var intersectionFeature []string
	for _, initfeature := range initFeatures {
		for _, feature := range hostRelationFeatureMap[fmKey] {
			if initfeature == feature {
				intersectionFeature = append(intersectionFeature, initfeature)
			}
		}
	}
	if len(intersectionFeature) == 0 {
		return
	}
	depHost, depOk := res[depJobID]
	if depOk {
		depHostEnv, depEnvOk := depHost[env]
		if depEnvOk {
			depHostTribe, dpeHTOk := depHostEnv[tribeName]
			// 如果前置app有该组件 且 组件配置的使用版本合规 那就赋值到当前的数据下
			if dpeHTOk {
				// log.Info("TribeIsUseable-(组件是否可用): feature(%v) -> TribeVersionUseable", fmKey)
				tribePacks := s.TribeVersionUseable(curJobID, depHostTribe, relationMap, intersectionFeature)
				if len(tribePacks) != 0 {
					resTribePack = tribePacks
					// log.Info("[TribeListAll]-TribeIsUseable: 兼容的前置包合规的数据%v, %+v", len(resTribePack), resTribePack)
					return
				} else {
					resTribePack = s.TribeIsUseable(res, relationMap, curJobID, depJobID, env, tribeName, hostRelationFeatureMap, intersectionFeature)
				}
			} else {
				resTribePack = s.TribeIsUseable(res, relationMap, curJobID, depJobID, env, tribeName, hostRelationFeatureMap, intersectionFeature)
			}
		} else {
			resTribePack = s.TribeIsUseable(res, relationMap, curJobID, depJobID, env, tribeName, hostRelationFeatureMap, intersectionFeature)
		}
	}
	return
}

// nolint:gocognit
func (s *Service) TribeVersionUseable(curJobID int64, tribePacks []*tribemdl.TribeApk, relationMap map[int64]int64, features []string) (useableTribePack []*tribemdl.TribeApk) {
	// log.Info("TribeVersionUseable-(判断版本是否可用-%v)：组件列表(%+v), 兼容关系(%+v), features(%+v)", curJobID, tribePacks, relationMap, features)
	for _, tp := range tribePacks {
		// 判断feature 是否符合
		if features != nil {
			featureAble := false
			for _, feature := range features {
				if tp.DepFeature == feature {
					// log.Info("TribeVersionUseable-(判断版本是否可用-%v): feature(%v)已匹配", curJobID, feature)
					featureAble = true
				}
			}
			if !featureAble {
				// log.Info("TribeVersionUseable-(判断版本是否可用-%v): feature(%v)未匹配", curJobID, tp.DepFeature)
				continue
			}
		}
		// log.Info("TribeVersionUseable-(判断版本是否可用-%v): UpgradeConfig(%+v)", curJobID, tp.UpgradeConfig)
		if tp.UpgradeConfig == nil || (len(tp.UpgradeConfig.ChosenVersionCode) == 0 && len(tp.UpgradeConfig.StartVersionCode) == 0) {
			// log.Info("TribeVersionUseable-(判断版本是否可用-%v): 未进行任何版本配置", curJobID)
			useableTribePack = append(useableTribePack, tp)
			continue
		}
		if len(tp.UpgradeConfig.ChosenVersionCode) != 0 {
			// log.Info("TribeVersionUseable-(判断版本是否可用-%v)-配置的指定版本-%v: 宿主(%v)", curJobID, tp.UpgradeConfig.ChosenVersionCode, tp.TribeHostJobID)
			verArr := strings.Split(tp.UpgradeConfig.ChosenVersionCode, ",")
			for _, v := range verArr {
				if vint, err := strconv.ParseInt(v, 10, 64); err == nil && vint == curJobID {
					// log.Info("TribeVersionUseable-(判断版本是否可用-%v)-配置的指定版本-%v:，宿主(%v)，符合组件(%+v)", curJobID, tp.UpgradeConfig.ChosenVersionCode, tp.TribeHostJobID, tp)
					useableTribePack = append(useableTribePack, tp)
					break
				}
			}
		}
		if len(tp.UpgradeConfig.StartVersionCode) != 0 {
			// log.Info("TribeVersionUseable-(判断版本是否可用-%v)-配置的版本范围-%v: 宿主(%v)", curJobID, tp.UpgradeConfig.StartVersionCode, tp.TribeHostJobID)
			verArr := strings.Split(tp.UpgradeConfig.StartVersionCode, ",")
			for _, v := range verArr {
				if vint, err := strconv.ParseInt(v, 10, 64); err == nil {
					useable := s.FindVersionUseable(vint, curJobID, relationMap)
					if useable {
						// log.Info("TribeVersionUseable-(判断版本是否可用-%v)-配置的版本范围-%v: 宿主(%v),匹配的组件(%+v)", curJobID, tp.UpgradeConfig.StartVersionCode, tp.TribeHostJobID, tp)
						// 因为两个配置可以共存， 所以对结果去重
						isInRes := false
						for _, tpack := range useableTribePack {
							if tpack.ID == tp.ID {
								isInRes = true
								break
							}
						}
						if !isInRes {
							useableTribePack = append(useableTribePack, tp)
						}
						break
					}
				}
			}
		}
	}
	return
}
func (s *Service) FindVersionUseable(buildID, curJobID int64, relationMap map[int64]int64) (useable bool) {
	// log.Info("FindVersionUseable-(寻找兼容的版本)：buildID(%v), curID(%v), relationMap(%+v)", buildID, curJobID, relationMap)
	// 递归去找兼容链路 上是否存在该curJobID 版本
	if buildID == curJobID {
		useable = true
		return
	}
	for curID, devID := range relationMap {
		if devID == buildID {
			// log.Info("===devID(%v)-buildID(%v), curID(%v)-%v-curJobID(%v)", devID, buildID, curID, curID == curJobID, curJobID)
			if curID == curJobID {
				useable = true
				return
			} else {
				res := s.FindVersionUseable(curID, curJobID, relationMap)
				if res {
					useable = true
					return
				}
			}
		}
	}
	return
}
func (s *Service) TribeListAllCache(c context.Context) (map[string]map[int64]map[string]map[string][]*tribemdl.TribeApk, error) {
	return s.tribeListAllCache, nil
}

func (s *Service) TribeRelationAllCache(c context.Context) (map[int64]int64, error) {
	return s.tribeHostRelationCache, nil
}

func (s *Service) AppUseableTribes(c context.Context, appKey, env string, buildID int64) (map[string][]*tribemdl.TribeApk, error) {
	appTribe, ok := s.tribeListAllCache[appKey]
	if !ok {
		return nil, nil
	}
	appTribeEnv, ok := appTribe[buildID]
	if !ok {
		return nil, nil
	}
	tribes, ok := appTribeEnv[env]
	if !ok {
		return nil, nil
	}
	return tribes, nil
}

func (s *Service) TribeHosts(c context.Context, appKey, env, tribeName string, id int64) ([]int64, error) {
	var hostIDs []int64
	appTribe, ok := s.tribeListAllCache[appKey]
	if !ok {
		return nil, nil
	}
	for buildID := range appTribe {
		for _, tribe := range appTribe[buildID][env][tribeName] {
			if tribe.ID == id {
				hostIDs = append(hostIDs, buildID)
				break
			}
		}
	}
	return hostIDs, nil
}

// ReleaseVersionList
func (s *Service) AppCDVersionList(c context.Context, appKey, env string) (res []*model.Version, err error) {
	var (
		list []*model.Version
	)
	// get version list
	if list, err = s.fkDao.AppCDVersionList(c, appKey, env); err != nil {
		log.Error("ReleaseVersionList error: %v", err)
		return
	}
	// filter data
	versions := map[string]*model.Version{}
	for _, version := range list {
		versions[version.Version] = version
	}
	// sort key
	var keys []string
	for k := range versions {
		keys = append(keys, k)
	}
	sort.Sort(sort.Reverse(sort.StringSlice(keys)))

	// generator res
	res = []*model.Version{}
	for _, k := range keys {
		res = append(res, versions[k])
	}
	return
}

// HawkeyeWebhookCrash
//func (s *Service) HawkeyeWebhookCrash(c context.Context, alterParams *apmmdl.AlertWebhookParams) (err error) {
//	var (
//		appKey        string
//		versionCode   int64
//		parseIntError error
//		crashCount    int64
//		setupCount    int64
//	)
//	// 过滤位置应用
//	if appKey = alterParams.Labels["app_key"]; appKey == "" {
//		return
//	}
//	// 异常版本. 则无视版本继续执行告警
//	versionCodeStr := alterParams.Labels["version_code"]
//	if versionCode, parseIntError = strconv.ParseInt(versionCodeStr, 10, 64); parseIntError != nil {
//		versionCode = 0
//	}
//	// 参数解析
//	TriggerValue := alterParams.TriggerValue
//	receivers := strings.Join(alterParams.Receivers, ",")
//	startTime, _ := time.ParseInLocation(time.RFC3339Nano, alterParams.StartAt, time.Local)
//	triggerTime, _ := time.ParseInLocation(time.RFC3339Nano, alterParams.TriggerAt, time.Local)
//	// 查询APP
//	appInfo, _ := s.fkDao.AppPass(c, appKey)
//	if crashInfo, _ := s.fkDao.ApmAggregateCrashInfo(c, appKey, versionCode, triggerTime.Unix()); crashInfo != nil {
//		crashCount = crashInfo.DistinctBuvidCount
//	}
//	if setupInfo, _ := s.fkDao.ApmAggregateSetupInfo(c, appKey, versionCode, triggerTime.Unix()); setupInfo != nil {
//		setupCount = setupInfo.DistinctBuvidCount
//	}
//	messages := []string{}
//	messages = append(messages, fmt.Sprintf("版本：%v", versionCode))
//	messages = append(messages, fmt.Sprintf("崩溃率：%.2f%%", 100*float64(crashCount)/float64(setupCount)))
//	messages = append(messages, fmt.Sprintf("影响用户数：%d", crashCount))
//	messages = append(messages, fmt.Sprintf("用户启动量：%d", setupCount))
//	messages = append(messages, fmt.Sprintf("首次触发时间：%v", startTime.Format("2006-01-02 15:04:05")))
//	messages = append(messages, fmt.Sprintf("本次触发时间：%v", triggerTime.Format("2006-01-02 15:04:05")))
//	messages = append(messages, fmt.Sprintf("消息接收者：%v", receivers))
//	err = s.fkDao.WechatCardMessageNotify(
//		fmt.Sprintf("%v (%v) 崩溃率异常 - %.2f%%", appInfo.Name, appInfo.AppKey, TriggerValue),
//		strings.Join(messages, "\n"),
//		fmt.Sprintf("http://fawkes.bilibili.co/#/apm/crash/overview?app_key=%v", appInfo.AppKey),
//		"",
//		strings.Join(alterParams.Receivers, "|"),
//		s.c.Comet.MonitorAppID)
//	return
//}

// GenerateList
func (s *Service) GenerateList(c context.Context, appKey string, buildID int64) (res []*cdmdl.Generate, err error) {
	if res, err = s.fkDao.PublishGenerateList(c, appKey, buildID); err != nil {
		log.Error("%v", err)
	}
	for _, gen := range res {
		gen.PackCDNURL = strings.Replace(gen.GeneratePath, s.c.LocalPath.LocalDir, fmt.Sprintf("%v/mobile", s.c.Oss.Inland.CDNDomain), -1)
	}
	return
}

func (s *Service) AppkeyList(ctx context.Context) ([]string, error) {
	return s.fkDao.ModBusAppKeyList(ctx)
}

// AppKeyFileList map[pool_name]map[module_name]
func (s *Service) AppKeyFileList(ctx context.Context, appKey string, env mod.Env, md5Val string) (map[string]map[string][]*mod.BusFile, string, error) {
	var (
		ok          bool
		md5CacheVal interface{}
		modVal      interface{}
		md5CacheKey = string(env) + "_" + appKey
	)

	if md5CacheVal, ok = s.md5Cache.Load(md5CacheKey); !ok {
		log.Errorc(ctx, "env: %s appKey: %s md5 cache not found", env, appKey)
		return nil, "", ecode.Error(ecode.RequestErr, fmt.Sprintf("env: %s appKey: %s data not found", env, appKey))
	}
	if md5CacheVal.(string) == md5Val {
		return nil, "", ecode.NotModified
	}
	if modVal, ok = s.modCache.Load(env); !ok {
		log.Errorc(ctx, "env: %s appKey: %s mod cache  not found", env, appKey)
		return nil, "", ecode.Error(ecode.RequestErr, fmt.Sprintf("env: %s appKey: %s data not found", env, appKey))
	}
	data := modVal.(map[string]map[string]map[string][]*mod.BusFile)[appKey]
	return data, md5CacheVal.(string), nil
}

// moduleList map[appKey]map[pool_name]map[module_name]
// nolint:gocognit
func (s *Service) moduleList(ctx context.Context, pool map[string]map[string]*mod.BusPool, module map[int64]map[string]*mod.BusModule, env mod.Env) (map[string]map[string]map[string][]*mod.BusFile, error) {
	var moduleIDs []int64
	for _, v := range module {
		for _, m := range v {
			moduleIDs = append(moduleIDs, m.ID)
		}
	}
	if len(moduleIDs) == 0 {
		return nil, nil
	}
	version, err := s.fkDao.ModBusVersionListByModuleIDs(ctx, moduleIDs, env)
	if err != nil {
		return nil, err
	}
	allSuccessVersion, err := s.fkDao.ModBusVersionListByModuleIDs2(ctx, moduleIDs, env)
	if err != nil {
		return nil, err
	}
	prodVersion := getProdVersion(allSuccessVersion)
	var fromVerIDs []int64
	versionIDm := map[int64]struct{}{}
	// moduleId-version-struct{} module下的版本map
	envVersionMap := make(map[int64]map[int64]struct{})
	for mId, val := range version {
		for _, vs := range val {
			for _, v := range vs {
				versionIDm[v.ID] = struct{}{}
				if v.Env == mod.EnvProd {
					if m, ok := envVersionMap[mId]; ok {
						m[v.Version] = struct{}{}
					} else {
						versionMap := make(map[int64]struct{})
						versionMap[v.Version] = struct{}{}
						envVersionMap[mId] = versionMap
					}
					fromVerIDs = append(fromVerIDs, v.FromVerID)
				}
			}
		}
	}
	if len(fromVerIDs) != 0 {
		ver, err := s.fkDao.ModBusVersionListByIDs(ctx, fromVerIDs)
		if err != nil {
			return nil, err
		}
		for _, v := range ver {
			versionIDm[v.ID] = struct{}{}
		}
	}
	var versionIDs []int64
	for v := range versionIDm {
		versionIDs = append(versionIDs, v)
	}
	if len(versionIDs) == 0 {
		return nil, nil
	}
	var (
		file   map[int64][]*mod.BusFile
		config map[int64]*mod.BusVersionConfig
		gray   map[int64]*mod.BusVersionGray
	)
	g := errgroup.WithCancel(ctx)
	g.Go(func(ctx context.Context) (err error) {
		file, err = s.fkDao.ModBusFileList(ctx, versionIDs)
		return err
	})

	g.Go(func(ctx context.Context) (err error) {
		config, err = s.fkDao.ModBusVersionConfigList(ctx, versionIDs)
		return err
	})
	g.Go(func(ctx context.Context) (err error) {
		gray, err = s.fkDao.ModBusVersionGrayList(ctx, versionIDs)
		return err
	})
	if err := g.Wait(); err != nil {
		return nil, err
	}
	// appKey pool_name module_name
	res := map[string]map[string]map[string][]*mod.BusFile{}
	for appKey, val := range pool {
		data, ok := res[appKey]
		if !ok {
			data = map[string]map[string][]*mod.BusFile{}
			res[appKey] = data
		}
		for name, pool := range val {
			p, ok := data[name]
			if !ok {
				p = map[string][]*mod.BusFile{}
				data[name] = p
			}
			for name, module := range module[pool.ID] {
				for _, vs := range version[module.ID] {
					for _, v := range vs {
						versionID := v.ID
						// prod 获取 file 需要 from_ver_id
						if v.Env == mod.EnvProd {
							versionID = v.FromVerID
						}
						v.PoolID = pool.ID
						v.PoolName = pool.Name
						v.ModuleName = module.Name
						v.Compress = module.Compress
						v.IsWifi = module.IsWifi
						v.ZipCheck = module.ZipCheck
						var sourceFileSize int64 = 0
						// file[versionID] 该版本下的所有file 包含path和源文件
						for _, f := range file[versionID] {
							if !f.IsPatch {
								sourceFileSize = f.Size
								break
							}
						}
						for _, f := range file[versionID] {
							// 过滤掉增量包大于原包的情况
							if f.IsPatch {
								if (len(conf.Conf.Mod.Switch.Patch.FileUrl) != 0 && utils.Contain(f.URL, conf.Conf.Mod.Switch.Patch.FileUrl)) || len(conf.Conf.Mod.Switch.Patch.FileUrl) == 0 {
									if sourceFileSize != 0 && f.Size >= sourceFileSize {
										continue
									}
								}
							}
							// 生产环境需要过滤掉测试的patch包
							if v.Env == mod.EnvProd && f.IsPatch {
								if m, ok := prodVersion[module.ID]; ok {
									if _, ok1 := m[f.FromVer]; !ok1 {
										continue
									}
								}
							}
							f.Version = v
							f.Config = config[v.ID]
							f.Gray = gray[v.ID]
							p[name] = append(p[name], f)
						}
					}
				}
			}
		}
	}
	return res, nil
}

func getProdVersion(version map[int64]map[mod.Env][]*mod.BusVersion) map[int64]map[int64]struct{} {
	prodVersion := make(map[int64]map[int64]struct{})
	for mId, val := range version {
		for _, vs := range val {
			for _, v := range vs {
				if v.Env == mod.EnvProd {
					if m, ok := prodVersion[mId]; ok {
						m[v.Version] = struct{}{}
					} else {
						versionMap := make(map[int64]struct{})
						versionMap[v.Version] = struct{}{}
						prodVersion[mId] = versionMap
					}
				}
			}
		}
	}
	return prodVersion
}

// TestFlightAll Get all testflight infos.
func (s *Service) TestFlightAll(c context.Context, env string) (res []*cdmdl.TestFlightUpdInfo, err error) {
	var appBaseInfos []*cdmdl.TFAppBaseInfo
	if appBaseInfos, err = s.fkDao.TFAppInfos(context.Background()); err != nil {
		log.Error("TFAppInfos %v", err)
		return
	}
	for _, oneApp := range appBaseInfos {
		re := &cdmdl.TestFlightUpdInfo{}
		re.AppKey = oneApp.AppKey
		re.MobiApp = oneApp.MobiApp
		if re.LatestOnline, err = s.fkDao.LatestOnline(context.Background(), oneApp.AppKey); err != nil {
			log.Error("LatestOnline %v", err)
			continue
		}
		// e.g. itms-apps://itunes.apple.com/cn/app/id736536022?mt=8
		if re.LatestOnline != nil {
			re.LatestOnline.UpdateURL = "itms-apps://itunes.apple.com/cn/app/id" + oneApp.StoreAppID + "?mt=8"
		}
		if re.LatestTF, err = s.fkDao.LatestTF(context.Background(), oneApp.AppKey, env); err != nil {
			log.Error("LatestTF %v", err)
			continue
		}
		if re.LatestTF != nil {
			if env == "test" {
				re.LatestTF.UpdateURL = oneApp.PublicLinkTest
			} else {
				re.LatestTF.UpdateURL = oneApp.PublicLink
			}
		}
		if re.TFPacks, err = s.fkDao.TFPackList(context.Background(), oneApp.AppKey, env); err != nil {
			log.Error("LatestTF %v", err)
			continue
		}
		// get black & white list
		var (
			blackList, whiteList []*cdmdl.TestFlightBWList
		)
		if blackList, err = s.fkDao.TFBlackWhiteList(context.Background(), re.AppKey, "black", env); err != nil {
			log.Error("%v", err)
			return
		}
		for _, user := range blackList {
			re.BlackList = append(re.BlackList, user.MID)
		}
		if whiteList, err = s.fkDao.TFBlackWhiteList(context.Background(), re.AppKey, "white", env); err != nil {
			log.Error("%v", err)
			return
		}
		for _, user := range whiteList {
			re.WhiteList = append(re.WhiteList, user.MID)
		}
		res = append(res, re)
	}
	return
}

// PatchSetStatus set patch build status
func (s *Service) PatchSetStatus(c context.Context, id, glJobID int64, status int, appKey string) (err error) {
	var patchInfo *cdmdl.Patch
	if patchInfo, err = s.fkDao.PatchInfo(c, id, appKey); err != nil {
		log.Error("PatchInfo error: %v", err)
		return
	}
	if patchInfo.Status == cdmdl.PatchStatusSuccess {
		err = errors.New("【Patch】: status is already success, don`t trigger again")
		return
	}
	var tx *sql.Tx
	if tx, err = s.fkDao.BeginTran(c); err != nil {
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
	if err = s.fkDao.TxPatchStatus(tx, id, glJobID, status, appKey); err != nil {
		log.Error("%v", err)
	}
	return
}

// PatchUpload upload patch file
func (s *Service) PatchUpload(c context.Context, id int64, file multipart.File, uploadFileMd5, appKey string) (err error) {
	var (
		patchInfo     *cdmdl.Patch
		destFile      *os.File
		tx            *sql.Tx
		cdnPath, fmd5 string
		size          int64
	)
	fileData, err := ioutil.ReadAll(file)
	if err != nil {
		return
	}
	// md5 验证
	h := md5.New()
	defer file.Close()
	_, err = io.Copy(h, bytes.NewReader(fileData))
	if err != nil {
		log.Error("io.Copy error(%v)", err)
		return
	}
	fileMd5Bs := h.Sum(nil)
	if uploadFileMd5 != hex.EncodeToString(fileMd5Bs[:]) {
		log.Error("upload file md5 check failed")
		return errors.New("upload file md5 check failed")
	}
	if patchInfo, err = s.fkDao.PatchInfo(c, id, appKey); err != nil {
		log.Error("PatchInfo error: %v", err)
		return
	}
	if patchInfo.Status == cdmdl.PatchStatusSuccess {
		err = errors.New("【Patch】:status is already success, don`t trigger again")
		return
	}
	folder := path.Join("pack", patchInfo.AppKey, strconv.FormatInt(patchInfo.BuildID, 10), "patch")
	patchName := strconv.FormatInt(patchInfo.OriginBuildID, 10) + "-to-" + strconv.FormatInt(patchInfo.BuildID, 10) + ".patch"
	outPath := path.Join(s.c.LocalPath.LocalDir, folder, patchName)
	inetPath := s.c.LocalPath.LocalDomain + "/" + path.Join(folder, patchName)
	// if patch file exist
	patchDir := path.Join(s.c.LocalPath.LocalDir, folder)
	if _, err = os.Stat(patchDir); err != nil {
		if err = os.MkdirAll(patchDir, 0755); err != nil {
			log.Error("os.MkdirAll error:%v", err)
			return
		}
	}
	if _, err = os.Stat(outPath); err == nil {
		// err = errors.New("Patch file already exist")
		// log.Error("%v Patch file already exist!", outPath)
		// return
		dupliPathName := "duplicate-" + strconv.FormatInt(time.Now().Unix(), 10) + "-" + patchName
		_ = os.Rename(outPath, path.Join(s.c.LocalPath.LocalDir, folder, dupliPathName))
	}
	if destFile, err = os.Create(outPath); err != nil {
		log.Error("%v", err)
		return
	}
	if _, err = io.Copy(destFile, bytes.NewReader(fileData)); err != nil {
		log.Error("%v", err)
		return
	}
	_ = destFile.Close()
	if cdnPath, fmd5, size, err = s.fkDao.FilePutOss(c, folder, patchName, appKey); err != nil {
		log.Error("FilePutOss: %v", err)
		return
	}
	if tx, err = s.fkDao.BeginTran(c); err != nil {
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
	if err = s.fkDao.TxUpPatchFileInfo(tx, id, size, fmd5, outPath, inetPath, cdnPath); err != nil {
		log.Error("%v", err)
	}
	return
}

func (s *Service) ParseIP(c context.Context, addr string) (ipInfo *toolmdl.IPInfo, err error) {
	name := fmt.Sprintf("%v%v", s.c.LocalPath.LocalDir, s.c.IPDB.Ipv4)
	if ipInfo, err = utils.ParseIP(addr, name); err != nil {
		log.Errorc(c, "%v", err)
	}
	return
}

func (s *Service) AddPcdnFile(c context.Context, rid, url, md5, business, versionId string, size int64) (err error) {
	_ = s.fkDao.Transact(c, func(tx *sql.Tx) error {
		if err = s.fkDao.TxAddPcdnFile(tx, rid, url, md5, business, versionId, size); err != nil {
			log.Error("%v", err)
		}
		return err
	})
	return
}

func (s *Service) PcdnFileList(ctx context.Context, versionId, zone string) (resp *pcdn.ListResp, err error) {
	var files []*pcdn.Files
	defer func() {
		if _, err = s.fkDao.AddPcdnQueryLog(ctx, versionId, zone); err != nil {
			log.Errorc(ctx, "%v", err)
			return
		}
	}()
	if len(versionId) == 0 {
		var latestVer string
		if latestVer, err = s.fkDao.LatestPcdnQueryLog(ctx, zone); err != nil {
			log.Errorc(ctx, "%v", err)
			return
		}
		versionId = latestVer
	}
	if files, err = s.fkDao.PcdnFileByVersion(ctx, versionId); err != nil {
		log.Errorc(ctx, "%v", err)
		return
	}
	resp = buildResp(files, versionId)
	return
}

func buildResp(files []*pcdn.Files, versionId string) *pcdn.ListResp {
	var (
		res       = &pcdn.ListResp{}
		items     []*pcdn.Item
		latestVer = versionId
	)
	for _, f := range files {
		item := &pcdn.Item{
			Rid:     f.Rid,
			Url:     f.Url,
			Md5:     f.Md5,
			Size:    f.Size,
			Popular: 50,
		}
		items = append(items, item)
		if f.VersionId > latestVer {
			latestVer = f.VersionId
		}
	}
	res.LatestVersion = latestVer
	res.Resource = items
	return res
}
