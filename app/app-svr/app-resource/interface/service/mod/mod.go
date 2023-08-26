package mod

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"hash/crc32"
	"strconv"
	"strings"
	"sync"
	"time"

	"go-common/library/ecode"
	"go-common/library/log"
	xtime "go-common/library/time"
	"go-common/library/xstr"

	"go-gateway/app/app-svr/app-resource/interface/conf"
	moddao "go-gateway/app/app-svr/app-resource/interface/dao/mod"
	"go-gateway/app/app-svr/app-resource/interface/model/mod"
	"go-gateway/app/app-svr/app-resource/interface/model/module"

	"go-common/library/sync/errgroup.v2"

	v1 "git.bilibili.co/bapis/bapis-go/bilibili/app/resource/v1"

	"github.com/robfig/cron"
)

type Service struct {
	dao *moddao.Dao
	c   *conf.Config
	// cron
	cron     *cron.Cron
	modCache sync.Map // map[cacheKey(env,app_key)]map[pool_name]map[module_name]
	md5Cache sync.Map // map[cacheKey(env,app_key)]
}

func New(c *conf.Config) (s *Service) {
	s = &Service{
		dao: moddao.New(c),
		c:   c,
		// cron
		cron: cron.New(),
	}
	if err := s.loadModule(); err != nil {
		panic(err)
	}
	s.initCron()
	s.cron.Start()
	return
}

func (s *Service) Close() {
	if s.cron != nil {
		s.cron.Stop()
	}
	s.dao.Close()
}

func (s *Service) initCron() {
	//nolint:errcheck
	if err := s.cron.AddFunc(s.c.Cron.LoadModCache, func() { s.loadModule() }); err != nil {
		panic(err)
	}
}

func modCacheKey(env mod.Env, appKey string) string {
	return fmt.Sprintf("%s_%s", env, appKey)
}

//nolint:gocognit
func (s *Service) loadModule() error {
	configFunc := func(config *mod.VersionConfig) {
		if config == nil {
			return
		}
		if config.AppVer != "" {
			if err := json.Unmarshal([]byte(config.AppVer), &config.AppVers); err != nil {
				log.Error("日志告警 app_ver 值错误,config:%+v,error:%+v", config, err)
			}
			for _, ver := range config.AppVers {
				for cond := range ver {
					if !cond.Valid() {
						log.Error("日志告警 app_ver 值错误,config:%+v", config)
					}
				}
			}
		}
		if config.SysVer != "" {
			if err := json.Unmarshal([]byte(config.SysVer), &config.SysVers); err != nil {
				log.Error("日志告警 sys_ver 值错误,config:%+v,error:%+v", config, err)
			}
			for _, ver := range config.SysVers {
				for cond := range ver {
					if !cond.Valid() {
						log.Error("日志告警 sys_ver 值错误,config:%+v", config)
					}
				}
			}
		}
		if config.Scale != "" {
			vals := strings.Split(config.Scale, ",")
			config.Scales = map[mod.Scale]struct{}{}
			for _, val := range vals {
				scale := mod.Scale(val)
				if !scale.Valid() {
					log.Error("日志告警 scale 值错误,config:%+v", config)
					continue
				}
				config.Scales[scale] = struct{}{}
			}
		}
		if config.ForbidenDevice != "" {
			vals := strings.Split(config.ForbidenDevice, ",")
			config.ForbidenDevices = map[mod.Device]struct{}{}
			for _, val := range vals {
				device := mod.Device(val)
				if !device.Valid() {
					log.Error("日志告警 forbiden_device 值错误,config:%+v", config)
					continue
				}
				config.ForbidenDevices[device] = struct{}{}
			}
		}
		if config.Arch != "" {
			vals := strings.Split(config.Arch, ",")
			config.Archs = map[mod.Arch]struct{}{}
			for _, val := range vals {
				arch := mod.Arch(val)
				if !arch.Valid() {
					log.Error("日志告警 arch 值错误,config:%+v", config)
					continue
				}
				config.Archs[arch] = struct{}{}
			}
		}
	}
	grayFunc := func(ctx context.Context, gray *mod.VersionGray) {
		if gray == nil {
			return
		}
		gray.Whitelistm = map[int64]struct{}{}
		vals, err := xstr.SplitInts(gray.Whitelist)
		if err != nil {
			log.Error("日志告警 whitelist 值错误,gray:%+v,error:%+v", gray, err)
		}
		for _, val := range vals {
			gray.Whitelistm[val] = struct{}{}
		}
		if gray.WhitelistURL == "" {
			return
		}
		var bs []byte
		for index := 0; index < 5; index++ { // 最多重试5次
			bs, err = s.dao.WhitelistData(ctx, gray.WhitelistURL)
			if err != nil {
				time.Sleep(time.Microsecond * 200)
				continue
			}
			break
		}
		if err != nil {
			log.Error("日志告警 whitelist_url 值错误,gray:%+v,error:%+v", gray, err)
			return
		}
		if vals, err = xstr.SplitInts(string(bs)); err != nil {
			log.Error("日志告警 whitelist_url 值错误,gray:%+v,error:%+v", gray, err)
			return
		}
		for _, val := range vals {
			gray.Whitelistm[val] = struct{}{}
		}
	}
	appKeys, err := s.dao.AppKeyList(context.Background())
	if err != nil {
		return err
	}
	var g errgroup.Group
	envs := []mod.Env{mod.EnvProd, mod.EnvTest}
	for _, val := range appKeys {
		appKey := val
		for _, val2 := range envs {
			env := val2
			cacheKey := modCacheKey(env, appKey)
			var md5Val string
			if val, ok := s.md5Cache.Load(cacheKey); ok {
				md5Val = val.(string)
			}
			g.Go(func(ctx context.Context) error {
				res, newMD5Val, err := s.dao.FileList(ctx, appKey, env, md5Val)
				if err != nil {
					if ecode.EqualError(ecode.NotModified, err) {
						return nil
					}
					log.Error("日志告警 mod 全量缓存构建失败,env:%v,error:%+v", env, err)
					return err
				}
				for _, module := range res {
					for _, version := range module {
						for _, file := range version {
							configFunc(file.Config)
							grayFunc(ctx, file.Gray)
						}
					}
				}
				s.modCache.Store(cacheKey, res)
				s.md5Cache.Store(cacheKey, newMD5Val)
				return nil
			})
		}
	}
	if err = g.Wait(); err != nil {
		return err
	}
	return nil
}

// cache map[cacheKey(env,app_key)]map[pool_name]map[module_name]
// return map[pool_name]
func (s *Service) fileList(env mod.Env, appKey, poolName, moduleName string) map[string][]*mod.File {
	val, ok := s.modCache.Load(modCacheKey(env, appKey))
	if !ok {
		return nil
	}
	pool := val.(map[string]map[string][]*mod.File)
	if poolName != "" {
		module, ok := pool[poolName]
		if !ok {
			return nil
		}
		if moduleName != "" {
			file, ok := module[moduleName]
			if !ok {
				return nil
			}
			module = map[string][]*mod.File{moduleName: file}
		}
		pool = map[string]map[string][]*mod.File{poolName: module}
	}
	res := map[string][]*mod.File{}
	for poolName, module := range pool {
		for _, file := range module {
			res[poolName] = append(res[poolName], file...)
		}
	}
	return res
}

func (s *Service) HTTPList(ctx context.Context, appKey, buvid string, mid int64, device, poolName, env string, build, sysver, level, scale, arch int, versions []*module.Versions, now time.Time) []*module.ResourcePool {
	reply := []*module.ResourcePool{}
	var modEnv mod.Env
	switch env {
	case module.EnvTest:
		modEnv = mod.EnvTest
	case module.EnvRelease:
		modEnv = mod.EnvProd
	default:
		log.Error("日志告警 环境值错误,env:%v", env)
		modEnv = mod.EnvProd
	}
	fileList := s.fileList(modEnv, appKey, poolName, "")
	if len(fileList) == 0 {
		return reply
	}
	var lowPool []*module.ResourcePool
	verList := formCondition(versions)
	for poolName, files := range fileList {
		poolReply := &module.ResourcePool{Name: poolName}
		existedModule := map[string]struct{}{}
		patchFile := map[string]*mod.File{}
		for _, file := range files {
			if _, ok := existedModule[file.Version.ModuleName]; ok {
				continue
			}
			if _, ok := checkCondition(int64(build), int32(sysver), int32(scale), int32(arch), mod.Device(device), now, file.Config); !ok {
				continue
			}
			// versions里如果指定了module，表示允许手动下载
			ver, ok := verList[file.Version.PoolName][file.Version.ModuleName]
			if !checkGray(buvid, mid, ok, file.Gray) {
				continue
			}
			switch file.IsPatch {
			case true: // 增量
				// 未指定版本，不能下发增量包
				// 增量，版本不对过滤
				if !ok || ver != file.FromVer {
					continue
				}
				patchFile[file.Version.ModuleName] = file
			case false: // 全量
				f := &mod.File{}
				*f = *file
				if patch, ok := patchFile[f.Version.ModuleName]; ok {
					*f = *patch
				}
				f.TotalMd5 = file.Md5
				poolReply.Resources = append(poolReply.Resources, buildHTTPReply(f))
				existedModule[f.Version.ModuleName] = struct{}{}
			}
		}
		if len(poolReply.Resources) == 0 {
			continue
		}
		var skip bool
		// 部分大文件mod要放在末尾，客户端是按list依次下载
		for _, lowPoolName := range s.c.ModLowPool {
			if poolReply.Name == lowPoolName {
				lowPool = append(lowPool, poolReply)
				skip = true
				break
			}
		}
		if skip {
			continue
		}
		reply = append(reply, poolReply)
	}
	if len(lowPool) != 0 {
		reply = append(reply, lowPool...)
	}
	return reply
}

func (s *Service) HTTPModule(ctx context.Context, appKey, buvid string, mid int64, device, poolName, moduleName, env string, ver, build, sysver, level, scale, arch int, now time.Time) (*module.Resource, error) {
	var modEnv mod.Env
	switch env {
	case module.EnvTest:
		modEnv = mod.EnvTest
	case module.EnvRelease:
		modEnv = mod.EnvProd
	default:
		log.Error("日志告警 环境值错误,env:%v", env)
		modEnv = mod.EnvProd
	}
	fileList := s.fileList(modEnv, appKey, poolName, moduleName)
	if len(fileList) == 0 {
		return nil, ecode.NothingFound
	}
	patchFile := map[string]*mod.File{}
	for _, files := range fileList {
		for _, file := range files {
			if _, ok := checkCondition(int64(build), int32(sysver), int32(scale), int32(arch), mod.Device(device), now, file.Config); !ok {
				continue
			}
			// versions里如果指定了module，表示允许手动下载
			if !checkGray(buvid, mid, true, file.Gray) {
				continue
			}
			switch file.IsPatch {
			case true: // 增量
				// 未指定版本，不能下发增量包
				// 增量，版本不对过滤
				if ver == 0 || int64(ver) != file.FromVer {
					continue
				}
				patchFile[file.Version.ModuleName] = file
			case false: // 全量
				f := &mod.File{}
				*f = *file
				patch, ok := patchFile[f.Version.ModuleName]
				// 全量，版本符合
				if !ok && int64(ver) == file.Version.Version {
					return nil, ecode.NotModified
				}
				if ok {
					*f = *patch
				}
				f.TotalMd5 = file.Md5
				modReply := buildHTTPReply(f)
				return modReply, nil
			}
		}
	}
	return nil, ecode.NothingFound
}

func (s *Service) GRPCListWrap(ctx context.Context, appKey, buvid string, mid, build int64, device string, req *v1.ListReq, now time.Time) (*v1.ListReply, error) {
	switch req.Lite {
	case mod.LiteV1:
		return s.GRPCListV1(ctx, appKey, buvid, mid, build, device, req, now)
	case mod.LiteV2:
		return s.GRPCListV2(ctx, appKey, buvid, mid, build, device, req, now)
	default:
		return s.GRPCList(ctx, appKey, buvid, mid, build, device, req, now)
	}
}

// nolint:gocognit
func (s *Service) GRPCList(ctx context.Context, appKey, buvid string, mid, build int64, device string, req *v1.ListReq, now time.Time) (*v1.ListReply, error) {
	reply := &v1.ListReply{Env: req.Env.String()}
	var modEnv mod.Env
	switch req.Env {
	case v1.EnvType_Test:
		modEnv = mod.EnvTest
	case v1.EnvType_Release:
		modEnv = mod.EnvProd
	default:
		log.Error("日志告警 环境值错误,env:%v", req.Env)
		modEnv = mod.EnvProd
	}
	fileList := s.fileList(modEnv, appKey, req.GetPoolName(), req.GetModuleName())
	if len(fileList) == 0 {
		return reply, nil
	}
	disableList, err := s.dao.ModuleDisableList(ctx, appKey, modEnv)
	if err != nil {
		return nil, err
	}
	var lowPool []*v1.PoolReply
	verList := verListCondition(req.GetVersionList())
	for poolName, files := range fileList {
		poolReply := &v1.PoolReply{Name: poolName}
		existedModule := map[string]struct{}{}
		patchFile := map[string]*mod.File{}
		for _, file := range files {
			if disableVer, ok := disableList[fmt.Sprintf("%s_%s", poolName, file.Version.ModuleName)]; ok && disableVer >= file.Version.Version {
				s.filterLog(mid, req, file, "disable_mod")
				continue
			}
			if s.ForbidMod(poolName, file.Version.ModuleName) {
				s.filterLog(mid, req, file, "forbid_mod")
				continue
			}
			if _, ok := existedModule[file.Version.ModuleName]; ok {
				s.filterLog(mid, req, file, "existed_mod")
				continue
			}
			if cause, ok := checkCondition(build, req.GetSysVer(), req.GetScale(), req.GetArch(), mod.Device(device), now, file.Config); !ok {
				s.filterLog(mid, req, file, cause)
				continue
			}
			// versions里如果指定了module，表示允许手动下载
			ver, ok := verList[poolName][file.Version.ModuleName]
			if !checkGray(buvid, mid, ok, file.Gray) {
				s.filterLog(mid, req, file, "check_gray")
				continue
			}
			publishTime := file.Version.ReleaseTime
			if file.Config != nil && file.Config.Stime > publishTime {
				publishTime = file.Config.Stime
			}
			// 用户未下载过或者当前版本不匹配，并且未命中灰度，则跳过
			if (!ok || ver != file.Version.Version) && !s.grayMod(req.Env, buvid, publishTime, file.Gray, now) {
				s.filterLog(mid, req, file, "gray_mod")
				continue
			}
			switch file.IsPatch {
			case true: // 增量
				// 未指定版本，不能下发增量包
				// 增量，版本不对过滤
				if !ok || ver != file.FromVer {
					s.filterLog(mid, req, file, "patch_ver")
					continue
				}
				patchFile[file.Version.ModuleName] = file
			case false: // 全量
				f := &mod.File{}
				*f = *file
				if patch, ok := patchFile[f.Version.ModuleName]; ok {
					*f = *patch
				}
				f.TotalMd5 = file.Md5
				poolReply.Modules = append(poolReply.Modules, buildGRPCReply(f))
				existedModule[f.Version.ModuleName] = struct{}{}
			}
		}
		var skip bool
		// 部分大文件mod要放在末尾，客户端是按list依次下载
		for _, lowPoolName := range s.c.ModLowPool {
			if poolReply.Name == lowPoolName {
				lowPool = append(lowPool, poolReply)
				skip = true
				break
			}
		}
		if skip {
			continue
		}
		reply.Pools = append(reply.Pools, poolReply)
	}
	if len(lowPool) != 0 {
		reply.Pools = append(reply.Pools, lowPool...)
	}
	return reply, nil
}

func (s *Service) filterLog(mid int64, req *v1.ListReq, file *mod.File, cause string) {
	if ok := func() bool {
		if s.c.ModLogGray.Open {
			return true
		}
		if mid < 1 {
			return false
		}
		for _, val := range s.c.ModLogGray.Whitelist {
			if mid == val {
				return true
			}
		}
		h := md5.New()
		h.Write([]byte(strconv.FormatInt(mid, 10)))
		b, err := strconv.ParseUint(hex.EncodeToString(h.Sum(nil))[18:], 16, 64)
		if err != nil {
			log.Error("日志告警 分组错误,error:%+v", err)
			return false
		}
		if b%1000000 < s.c.ModLogGray.Bucket {
			return true
		}
		return false
	}(); !ok {
		return
	}
	rbs, _ := json.Marshal(req)
	fbs, _ := json.Marshal(file)
	log.Error("mod过滤,原因:%v,用户ID:%v,资源池ID:%v,资源ID:%v,\n请求:%s,\n资源:%s", cause, mid, file.Version.PoolID, file.Version.ModuleID, rbs, fbs)
}

func verListCondition(vs []*v1.VersionListReq) map[string]map[string]int64 {
	vm := map[string]map[string]int64{}
	for _, pool := range vs {
		poolName := pool.GetPoolName()
		if poolName == "" {
			continue
		}
		pm, ok := vm[poolName]
		if !ok {
			pm = map[string]int64{}
			vm[poolName] = pm
		}
		for _, mod := range pool.GetVersions() {
			pm[mod.GetModuleName()] = mod.GetVersion()
		}
	}
	return vm
}

func formCondition(versions []*module.Versions) map[string]map[string]int64 {
	res := map[string]map[string]int64{}
	for _, pools := range versions {
		var (
			re map[string]int64
			ok bool
		)
		for _, resource := range pools.Resource {
			if re, ok = res[pools.PoolName]; !ok {
				re = make(map[string]int64)
				res[pools.PoolName] = re
			}
			var tmpVer int64
			switch tmp := resource.Version.(type) {
			case string:
				tmpVer, _ = strconv.ParseInt(tmp, 10, 64)
			case float64:
				tmpVer = int64(tmp)
			}
			re[resource.ResourceName] = tmpVer
		}
	}
	return res
}

func checkCondition(build int64, sysVer, scale, arch int32, device mod.Device, now time.Time, config *mod.VersionConfig) (string, bool) {
	if config == nil {
		return "", true
	}
	stime := config.Stime.Time()
	if !stime.IsZero() && stime.After(now) {
		return "stime", false
	}
	etime := config.Etime.Time()
	if !etime.IsZero() && etime.Before(now) {
		return "etime", false
	}
	// 和老逻辑保持一致
	if scale > 0 && len(config.Scales) != 0 {
		var modScale mod.Scale
		//nolint:gomnd
		switch scale {
		case 1:
			modScale = mod.Scale1x
		case 2:
			modScale = mod.Scale2x
		case 3:
			modScale = mod.Scale3x
		}
		if _, ok := config.Scales[modScale]; !ok {
			return "scale", false
		}
	}
	// 和老逻辑保持一致
	if arch > 0 && len(config.Archs) != 0 {
		var modArch mod.Arch
		//nolint:gomnd
		switch arch {
		case 1:
			modArch = mod.ArchArmeabiV7a
		case 2:
			modArch = mod.ArchArm64V8a
		case 3:
			modArch = mod.ArchX86
		}
		if _, ok := config.Archs[modArch]; !ok {
			return "arch", false
		}
	}
	if len(config.ForbidenDevices) != 0 {
		if _, ok := config.ForbidenDevices[device]; ok {
			return "forbiden_device", false
		}
	}
	checkVerFunc := func(srcVer int64, verCfg []map[mod.Condition]int64) bool {
		var cfgVer []map[mod.Condition]int64
		for _, verm := range verCfg { // 处理 [{}]
			if len(verm) == 0 {
				continue
			}
			cfgVer = append(cfgVer, verm)
		}
		if len(cfgVer) == 0 {
			return true
		}
		for _, verm := range cfgVer {
			var ok bool
			for cond, ver := range verm {
				if !checkVer(srcVer, ver, cond) {
					ok = false
					break
				}
				ok = true
			}
			if ok {
				return true
			}
		}
		return false
	}
	if !checkVerFunc(build, config.AppVers) {
		return "app_ver", false
	}
	// 和老逻辑保持一致
	if sysVer > 0 {
		if !checkVerFunc(int64(sysVer), config.SysVers) {
			return "sys_ver", false
		}
	}
	return "", true
}

func checkVer(srcVer, cfgVer int64, cfgCond mod.Condition) bool {
	switch cfgCond {
	case mod.ConditionLt: // 小于
		if srcVer < cfgVer {
			return true
		}
	case mod.ConditionLe: // 小于等于
		if srcVer <= cfgVer {
			return true
		}
	case mod.ConditionGt: // 大于
		if srcVer > cfgVer {
			return true
		}
	case mod.ConditionGe: // 大于等于
		if srcVer >= cfgVer {
			return true
		}
	}
	return false
}

func checkGray(buvid string, mid int64, manualDownload bool, gray *mod.VersionGray) bool {
	if gray == nil {
		return true
	}
	// Comment: 策略 1-(UID MD5) 2-(DEVICE MD5) 3-(UID)
	// 分桶策略
	// 1. UID MD5：针对登录用户，(mid+盐值).md5().hex().uppercase().hash() % 1000 可以根据mid和盐值，将所有登陆后用户分为0-999 1000个分桶
	// 2. DEVICE MD5：针对所有状态用户，(deviceid+盐值).md5().hex().uppercase().hash() % 1000 可以根据deviceid和盐值，将所有用户分为0-999 1000个分桶
	// 3. UID：mid%1000 可以根据mid后三位直接分桶为0-999 1000个分桶
	bucket := int64(-1)
	//nolint:gomnd
	switch gray.Strategy {
	case 1:
		if mid > 0 {
			bucket = bucketValue(strconv.FormatInt(mid, 10), gray.Salt)
		}
		return hitCheck(bucket, mid, manualDownload, gray)
	case 2:
		if buvid != "" {
			bucket = bucketValue(buvid, gray.Salt)
		}
		return hitCheck(bucket, mid, manualDownload, gray)
	case 3:
		if mid > 0 {
			//nolint:gomnd
			bucket = mid % 1000
		}
		return hitCheck(bucket, mid, manualDownload, gray)
	default:
		log.Error("日志告警 strategy值错误,gray:%+v", gray)
		return false
	}
}

func hitCheck(bucket, mid int64, manualDownload bool, gray *mod.VersionGray) bool {
	if mid > 0 {
		if _, ok := gray.Whitelistm[mid]; ok {
			return true
		}
	}
	// 命中桶设为-1~-1，表示灰度关闭
	if bucket != -1 && bucket >= gray.BucketStart && bucket <= gray.BucketEnd {
		return true
	}
	// -1/0~999 桶开到100%，表示全量
	if bucket == -1 && gray.BucketStart <= 0 && gray.BucketEnd == 999 {
		return true
	}
	return manualDownload && gray.ManualDownload
}

func bucketValue(val, salt string) int64 {
	code := hashCode(md5HashEncode(val, salt))
	//nolint:gomnd
	return (int64(code)%1000 + 1000) % 1000
}

func md5HashEncode(val, salt string) string {
	h := md5.New()
	h.Write([]byte(val + salt))
	s := hex.EncodeToString(h.Sum(nil))
	return strings.ToUpper(s)
}

// hashCode implement of `java.lang.String` hashCode()
func hashCode(val string) int32 {
	if val == "" {
		return 0
	}
	var hash int32
	b := []byte(val)
	for i := range b {
		char := b[i]
		//nolint:gomnd
		hash = ((hash << 5) - hash) + int32(char)
	}
	return hash
}

func buildGRPCReply(file *mod.File) *v1.ModuleReply {
	// 优先级默认中-2
	var level v1.LevelType
	priority := mod.PriorityMiddle
	if file.Config != nil {
		priority = file.Config.Priority
	}
	switch priority {
	case mod.PriorityHigh:
		level = v1.LevelType_High
	case mod.PriorityMiddle:
		level = v1.LevelType_Middle
	case mod.PriorityLow:
		level = v1.LevelType_Low
	default:
		log.Error("日志告警 优先级值错误,version:%+v,config:%+v", file.Version, file.Config)
		level = v1.LevelType_Middle
	}
	increment := v1.IncrementType_Total
	if file.IsPatch {
		increment = v1.IncrementType_Incremental
	}
	var compress v1.CompressType
	switch file.Version.Compress {
	case mod.CompressOriginal:
		compress = v1.CompressType_Original
	case mod.CompressUnzip:
		compress = v1.CompressType_Unzip
	default:
		log.Error("日志告警 解压类型值错误,version:%+v", file.Version)
		compress = v1.CompressType_Unzip
	}
	publishTime := file.Version.ReleaseTime
	if file.Config != nil && file.Config.Stime > publishTime {
		publishTime = file.Config.Stime
	}
	return &v1.ModuleReply{
		Name:        file.Version.ModuleName,
		Version:     file.Version.Version,
		Url:         file.URL,
		Md5:         file.Md5,
		TotalMd5:    file.TotalMd5,
		Increment:   increment,
		IsWifi:      file.Version.IsWifi,
		ZipCheck:    file.Version.ZipCheck,
		Level:       level,
		Filename:    file.Name,
		FileType:    file.ContentType,
		FileSize:    file.Size,
		Compress:    compress,
		PublishTime: int64(publishTime),
		PoolId:      file.Version.PoolID,
		ModuleId:    file.Version.ModuleID,
		VersionId:   file.Version.ID,
		FileId:      file.ID,
	}
}

func buildHTTPReply(file *mod.File) *module.Resource {
	isWifi := 0
	if file.Version.IsWifi {
		isWifi = 1
	}
	// 优先级默认中-2
	var level int
	priority := mod.PriorityMiddle
	if file.Config != nil {
		priority = file.Config.Priority
	}
	switch priority {
	case mod.PriorityHigh:
		level = 1
	case mod.PriorityMiddle:
		level = 2
	case mod.PriorityLow:
		level = 3
	default:
		log.Error("日志告警 优先级值错误,version:%+v,config:%+v", file.Version, file.Config)
		level = 2
	}
	increment := 0
	if file.IsPatch {
		increment = 1
	}
	var compress int
	switch file.Version.Compress {
	case mod.CompressOriginal:
		compress = 1
	case mod.CompressUnzip:
		compress = 0
	default:
		log.Error("日志告警 解压类型值错误,version:%+v", file.Version)
		compress = 0
	}
	mtime := file.Version.Mtime
	if file.Config != nil && mtime < file.Config.Mtime {
		mtime = file.Config.Mtime
	}
	if file.Gray != nil && mtime < file.Gray.Mtime {
		mtime = file.Gray.Mtime
	}
	return &module.Resource{
		Name:         file.Version.ModuleName,
		Compresstype: compress,
		Type:         file.ContentType,
		URL:          file.URL,
		MD5:          file.Md5,
		TotalMD5:     file.TotalMd5,
		Size:         int(file.Size),
		Version:      int(file.Version.Version),
		Increment:    increment,
		Level:        level,
		IsWifi:       int8(isWifi),
		PoolID:       file.Version.PoolID,
		ModuleID:     file.Version.ModuleID,
		VersionID:    file.Version.ID,
		FileID:       file.ID,
		Mtime:        mtime,
	}
}

func (s *Service) ForbidMod(poolName, moduleName string) bool {
	for _, modName := range s.c.Mod.ModuleForbid[poolName] {
		if moduleName == modName {
			return true
		}
	}
	return false
}

func (s *Service) grayMod(env v1.EnvType, buvid string, releaseTime xtime.Time, gray *mod.VersionGray, now time.Time) bool {
	// 测试环境直接分发
	// 若没有配置灰度策略，默认20分钟内把发布的mod资源全量分发完毕，每分钟增量分发5%
	// 若已配置灰度策略，默认10分钟内在灰度范围内把发布的mod资源分发完毕，每分钟增量10%
	if env == v1.EnvType_Test {
		return true
	}
	rlsTime := releaseTime
	var isGray bool
	if gray != nil {
		isGray = true
		if gray.Mtime > rlsTime {
			rlsTime = gray.Mtime
		}
	}
	nowDur := now.Sub(rlsTime.Time())
	grayDur := time.Duration(s.c.Mod.GrayDuration)
	if nowDur > grayDur {
		return true
	}
	nowDurSec := nowDur / time.Second
	grayDurSec := grayDur / time.Second
	if isGray {
		grayDurSec /= 2
	}
	bucket := crc32.ChecksumIEEE([]byte(buvid)) % uint32(grayDurSec)
	//nolint:gosimple
	if bucket < uint32(nowDurSec) {
		return true
	}
	return false
}

// GRPCListV1 response体积优化V1版本
// 重新构造response返回体
// 1.zipCheck不返回
// 2.host统一外移
// 3.比较版本是否相同，相同则不返回
func (s *Service) GRPCListV1(ctx context.Context, appKey string, buvid string, mid int64, build int64, device string, req *v1.ListReq, now time.Time) (*v1.ListReply, error) {
	reply, err := s.GRPCList(ctx, appKey, buvid, mid, build, device, req, now)
	if err != nil {
		return reply, err
	}
	versionMap := buildReqMap(req)
	var pools []*v1.PoolReply
	for _, p := range reply.Pools {
		var pool = new(v1.PoolReply)
		var modules []*v1.ModuleReply

		for _, m := range p.Modules {
			var newModule *v1.ModuleReply
			if versionMap[p.Name][m.Name] == m.Version {
				// 请求和返回版本相同
				newModule = &v1.ModuleReply{
					Name:    m.Name,
					Version: m.Version,
				}
			} else {
				m.Url = replace(m.Url, s.c.Mod.FileHost)
				newModule = m
			}
			modules = append(modules, newModule)
		}
		pool.Name = p.Name
		pool.Modules = modules
		pools = append(pools, pool)
	}
	rep := &v1.ListReply{
		Env:   reply.Env,
		Pools: pools,
		Host: &v1.Host{
			Boss: s.c.Mod.FileHost.BOSS,
			Bfs:  s.c.Mod.FileHost.BFS,
		},
	}
	reply = rep

	return reply, err
}

// GRPCListV2 在V1的基础上，将手动下载的资源轻量返回
func (s *Service) GRPCListV2(ctx context.Context, appKey string, buvid string, mid int64, build int64, device string, req *v1.ListReq, now time.Time) (*v1.ListReply, error) {
	reply, err := s.GRPCList(ctx, appKey, buvid, mid, build, device, req, now)
	if err != nil {
		return reply, err
	}
	versionMap := buildReqMap(req)
	var pools []*v1.PoolReply
	for _, p := range reply.Pools {
		var pool = new(v1.PoolReply)
		var modules []*v1.ModuleReply

		for _, m := range p.Modules {
			var newModule *v1.ModuleReply
			if m.Level == v1.LevelType_Low || versionMap[p.Name][m.Name] == m.Version {
				// 手动下载或者请求和返回版本相同
				newModule = &v1.ModuleReply{
					Name:    m.Name,
					Version: m.Version,
				}
			} else {
				m.Url = replace(m.Url, s.c.Mod.FileHost)
				newModule = m
			}
			modules = append(modules, newModule)
		}
		pool.Name = p.Name
		pool.Modules = modules
		pools = append(pools, pool)
	}
	rep := &v1.ListReply{
		Env:   reply.Env,
		Pools: pools,
		Host: &v1.Host{
			Boss: s.c.Mod.FileHost.BOSS,
			Bfs:  s.c.Mod.FileHost.BFS,
		},
	}
	reply = rep

	return reply, err
}

func replace(url string, host conf.FileHost) string {
	if strings.HasPrefix(url, host.BOSS) {
		url = strings.Replace(url, host.BOSS, "boss://", -1)
	} else if strings.HasPrefix(url, host.BFS) {
		url = strings.Replace(url, host.BFS, "bfs://", -1)
	}
	return url
}

// versionMap poolName-modName-version
func buildReqMap(req *v1.ListReq) (versionMap map[string]map[string]int64) {
	versionMap = make(map[string]map[string]int64)
	for _, v := range req.VersionList {
		modVerMap := map[string]int64{}
		for _, modVer := range v.Versions {
			modVerMap[modVer.ModuleName] = modVer.Version
		}
		versionMap[v.PoolName] = modVerMap
	}
	return
}
