package mod

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	xsql "go-common/library/database/sql"
	"go-common/library/ecode"

	"go-common/library/sync/errgroup.v2"

	"go-gateway/app/app-svr/fawkes/service/conf"
	"go-gateway/app/app-svr/fawkes/service/model/mod"
	"go-gateway/app/app-svr/fawkes/service/model/template"
	log "go-gateway/app/app-svr/fawkes/service/tools/logger"
	"go-gateway/app/app-svr/fawkes/service/tools/utils"
)

// VersionReleaseCheck 发布检查
func (s *Service) VersionReleaseCheck(ctx context.Context, versionID int64, user string) (resp *mod.ReleaseCheckResponse, err error) {
	resp = &mod.ReleaseCheckResponse{}
	t, err := s.ModReleaseTrafficEstimate(ctx, versionID, user)
	if err != nil {
		err = ecode.Error(ecode.ServerErr, fmt.Sprintf("ModReleaseTrafficEstimate error:%v", err))
		log.Errorc(ctx, err.Error())
		return
	}
	notify := getDetail(t)
	return &mod.ReleaseCheckResponse{
		CostLevel:            int64(notify.Cost),
		Percentage:           notify.Percentage,
		Advice:               notify.Advice,
		DocUrl:               s.c.Mod.TrafficMoni.DocURL,
		DownloadCount:        notify.DownloadCount,
		OriginFileSize:       notify.OriginFileSize,
		PatchFileSize:        notify.PatchFileSize,
		AvgFileSize:          notify.AvgFileSize,
		CDNBandwidthOnline:   notify.DownloadCDNBandwidthOnline,
		CDNBandwidthEstimate: notify.DownloadCDNBandwidthEstimate,
		CDNBandwidthTotal:    notify.DownloadCDNBandwidthTotal,
		IsManual:             notify.IsManual,
	}, err
}

// Alert 根据发布预估的结果发送预警信息
func (s *Service) Alert(ctx context.Context, traffic *mod.Traffic, operateType mod.OperateType) (err error) {
	var notifyContent string
	receiver := conf.Conf.Mod.TrafficMoni.NotifyReceiver
	trafficDetail := getDetail(traffic)
	if int64(trafficDetail.Cost) >= s.c.Mod.TrafficMoni.Threshold {
		receiver = append(receiver, traffic.Operator)
		var adminList []*mod.Permission
		adminList, err = s.fkDao.ModPermissionList(ctx, traffic.Pool.ID)
		if err != nil {
			return
		}
		for _, v := range adminList {
			receiver = append(receiver, v.Username)
		}
	}
	notify := mod.TrafficNotify{
		TrafficDetail: trafficDetail,
		OperateType:   operateType,
	}
	if notifyContent, err = s.fkDao.TemplateAlter(notify, template.ModTrafficNotify); err != nil {
		log.Error("%v", err)
		return
	}
	if err = s.fkDao.WechatMessageNotify(notifyContent, strings.Join(receiver, "|"), conf.Conf.Comet.FawkesAppID); err != nil {
		log.Error("%v", err)
		return
	}
	return
}

// ModReleaseTrafficEstimate 发布带来的流量预估
func (s *Service) ModReleaseTrafficEstimate(ctx context.Context, versionID int64, user string) (traffic *mod.Traffic, err error) {
	var (
		pool                               *mod.Pool
		module                             *mod.Module
		version                            *mod.Version
		config                             *mod.Config
		gray                               *mod.Gray
		file                               *mod.File
		patches                            []*mod.Patch
		versionList                        []*mod.Version
		start, end                         time.Time
		setUpUserCount, onlineDownloadSize float64
		off, slice                         time.Duration
		tConfig                            = s.c.Mod.TrafficMoni
	)
	log.Infoc(ctx, fmt.Sprintf("MOD发布，版本号:%v,生效时间：%v", versionID, time.Now()))
	if version, err = s.fkDao.ModVersionByID(ctx, versionID); err != nil && err != xsql.ErrNoRows {
		log.Errorc(ctx, "%v", err)
		return
	}
	if version == nil || version.Env == mod.EnvTest {
		err = ecode.Error(ecode.RequestErr, "版本不存在|测试环境不计算带宽")
		log.Errorc(ctx, "%v", err)
		return
	}
	defer func() {
		if err != nil {
			s.errNotify(pool, module, err)
		}
	}()
	g := errgroup.WithCancel(ctx)
	g.Go(func(ctx context.Context) error {
		if module, err = s.fkDao.ModModuleByID(ctx, version.ModuleID); err != nil && err != xsql.ErrNoRows {
			log.Errorc(ctx, "%v", err)
			return err
		}
		if module == nil {
			return nil
		}
		if pool, err = s.fkDao.ModPoolByID(ctx, module.PoolID); err != nil && err != xsql.ErrNoRows {
			log.Errorc(ctx, "%v", err)
			return err
		}
		if file, err = s.fkDao.ModFile(ctx, version.FromVerID); err != nil && err != xsql.ErrNoRows {
			log.Errorc(ctx, "%v", err)
			return err
		}
		if patches, err = s.fkDao.ModPatchList(ctx, version.FromVerID); err != nil && err != xsql.ErrNoRows {
			log.Errorc(ctx, "%v", err)
			return err
		}
		return nil
	})
	g.Go(func(ctx context.Context) error {
		if versionList, err = s.fkDao.ModVersionList(ctx, version.ModuleID, mod.EnvProd, 0, -1); err != nil {
			log.Errorc(ctx, "%v", err)
			return err
		}
		return nil
	})
	g.Go(func(ctx context.Context) error {
		if config, err = s.fkDao.ModVersionConfig(ctx, versionID); err != nil && err != xsql.ErrNoRows {
			log.Errorc(ctx, "%v", err)
			return err
		}
		return nil
	})
	g.Go(func(ctx context.Context) error {
		if gray, err = s.fkDao.ModVersionGray(ctx, versionID); err != nil && err != xsql.ErrNoRows {
			log.Errorc(ctx, "%v", err)
			return err
		}
		return nil
	})
	if err = g.Wait(); err != nil {
		log.Errorc(ctx, "%v", err)
		return
	}
	if file == nil {
		err = ecode.Error(ecode.RequestErr, fmt.Sprintf("versionID:%v 找不到文件信息", versionID))
		log.Errorc(ctx, err.Error())
		return
	}
	if off, err = time.ParseDuration(tConfig.TimeOffSet); err != nil {
		log.Errorc(ctx, "err %v", err)
		return
	}
	if slice, err = time.ParseDuration(tConfig.TimeSlice); err != nil {
		log.Errorc(ctx, "err %v", err)
		return
	}
	if start, end, err = timeRange(ctx, config, off, slice); err != nil {
		err = ecode.Error(ecode.RequestErr, fmt.Sprintf("时间范围计算失败 error:%v", err))
		log.Errorc(ctx, err.Error())
		return
	}
	if setUpUserCount, err = s.ModDownloadCountEstimate(ctx, pool.AppKey, config, gray, start, end); err != nil {
		log.Errorc(ctx, "err %v", err)
		return
	}
	log.Infoc(ctx, fmt.Sprintf("app:%v,资源池:%v,资源:%v,版本号:%v,文件信息：%v, 发布生效，生效时间：%v, 在线人数：%v", pool.AppKey, pool.Name, module.Name, version.Version, file, time.Now(), setUpUserCount))
	if onlineDownloadSize, err = s.fkDao.ModDownloadSizeSum(ctx, "", "", "", start, end); err != nil {
		err = ecode.Error(ecode.RequestErr, fmt.Sprintf("在线流量查询失败 error:%v", err))
		log.Errorc(ctx, err.Error())
		return
	}
	traffic = &mod.Traffic{
		Pool:                    pool,
		Module:                  module,
		Version:                 version,
		File:                    file,
		Patches:                 getProdPatch(versionList, patches),
		Config:                  config,
		Gray:                    gray,
		SetUpUserCount:          setUpUserCount,
		DownloadSizeOnlineBytes: onlineDownloadSize,
		Operator:                user,
	}
	return
}

// ModConfigChangeTrafficEstimate 配置变更带来的流量变化预估
func (s *Service) ModConfigChangeTrafficEstimate(ctx context.Context, pool *mod.Pool, module *mod.Module, version *mod.Version, config *mod.Config, gray *mod.Gray, file *mod.File, patch []*mod.Patch) (traffic float64, err error) {
	var (
		estimateCount float64
		off, slice    time.Duration
	)
	if off, err = time.ParseDuration(s.c.Mod.TrafficMoni.TimeOffSet); err != nil {
		log.Errorc(ctx, "err %v", err)
		return
	}
	if slice, err = time.ParseDuration(s.c.Mod.TrafficMoni.TimeSlice); err != nil {
		log.Errorc(ctx, "err %v", err)
		return
	}
	start, end, _ := timeRange(ctx, config, off, slice)
	if estimateCount, err = s.ModDownloadCountEstimate(ctx, pool.AppKey, config, gray, start, end); err != nil {
		log.Errorc(ctx, "err %v", err)
		return
	}
	traffic = estimateCount * float64(file.Size)
	return
}

// ModGrayChangeTrafficEstimate 灰度变更带来的流量变化预估
func (s *Service) ModGrayChangeTrafficEstimate(ctx context.Context, pool *mod.Pool, module *mod.Module, version *mod.Version, config *mod.Config, gray *mod.Gray, file *mod.File, patch []*mod.Patch) (traffic float64, err error) {
	var (
		estimateCount float64
		off, slice    time.Duration
	)
	if off, err = time.ParseDuration(s.c.Mod.TrafficMoni.TimeOffSet); err != nil {
		log.Errorc(ctx, "err %v", err)
		return
	}
	if slice, err = time.ParseDuration(s.c.Mod.TrafficMoni.TimeSlice); err != nil {
		log.Errorc(ctx, "err %v", err)
		return
	}
	start, end, _ := timeRange(ctx, config, off, slice)
	if estimateCount, err = s.ModDownloadCountEstimate(ctx, pool.AppKey, config, gray, start, end); err != nil {
		log.Errorc(ctx, "err %v", err)
		return
	}
	traffic = estimateCount * float64(file.Size)
	return
}

// ModDownloadCountEstimate 计算一段时间内的活跃用户人数
func (s *Service) ModDownloadCountEstimate(ctx context.Context, appKey string, config *mod.Config, gray *mod.Gray, start, end time.Time) (activeUsersCount float64, err error) {
	var (
		onlineUserCount int64
		grayBucket      int64 = 1000
	)
	log.Infoc(ctx, fmt.Sprintf("[%v]活跃人数计算——配置：%v，灰度：%v 时间范围：[%v~%v]", appKey, config, gray, start, end))
	var appVer []map[mod.Condition]int64
	if config != nil && config.AppVer != "" {
		if err = json.Unmarshal([]byte(config.AppVer), &appVer); err != nil {
			log.Errorc(ctx, "%v", err)
			return
		}
	}
	if onlineUserCount, err = s.fkDao.ActiveUserCount(ctx, appKey, appVer, start, end); err != nil {
		log.Errorc(ctx, "%v", err)
		return
	}
	log.Infoc(ctx, fmt.Sprintf("[%v]活跃人数计算——版本限制：%v， 起止时间[%v - %v], 在线人数 %v", appKey, appVer, start, end, onlineUserCount))
	if gray != nil && !(gray.BucketEnd == -1 && gray.BucketStart == -1) {
		grayBucket = gray.BucketEnd - gray.BucketStart + 1
		log.Infoc(ctx, fmt.Sprintf("[%v]活跃人数计算——灰度百分比 %v%%", appKey, grayBucket/10))
	} else {
		log.Infoc(ctx, fmt.Sprintf("[%v]活跃人数计算——没有配置灰度", appKey))
	}
	activeUsersCount = float64(onlineUserCount) * float64(grayBucket) / 1000
	log.Infoc(ctx, fmt.Sprintf("[%v]活跃人数计算—— 最终人数：%v", appKey, activeUsersCount))
	return
}

func (s *Service) errNotify(pool *mod.Pool, module *mod.Module, errStack error) {
	var err error
	var notifyContent string
	tn := &mod.TrafficDetail{
		AppKey:   pool.AppKey,
		PoolName: pool.Name,
		ModName:  module.Name,
		ModUrl:   fmt.Sprintf(s.c.Mod.TrafficMoni.ModUrl, pool.AppKey, pool.ID, module.ID, string(mod.EnvProd)),
		ErrorMsg: errStack.Error(),
	}
	if notifyContent, err = s.fkDao.TemplateAlter(tn, template.ModTrafficNotifyErr); err != nil {
		log.Error("%v", err)
		return
	}
	_ = s.fkDao.WechatMessageNotify(notifyContent, strings.Join(conf.Conf.Mod.TrafficMoni.NotifyReceiver, "|"), conf.Conf.Comet.FawkesAppID)
}

// 计算起止时间 offset-向前偏移的量，因为数据落库存在延迟 dur-时间范围 offset\dur均大于0
func timeRange(_ context.Context, config *mod.Config, offset, dur time.Duration) (start, end time.Time, err error) {
	// 如果配置了sTime则使用配置 否则使用当前时间
	sTime, now := time.Now(), time.Now()
	if config != nil {
		configSTime := time.Unix(int64(config.Stime), 0)
		if configSTime.After(time.Date(1990, 1, 1, 0, 0, 0, 0, time.Local)) {
			sTime = configSTime
		}
	}
	hour, min, sec := sTime.Clock()
	if now.Hour() < sTime.Hour() && now.Minute() < sTime.Minute() && now.Second() < sTime.Second() {
		yesterday := now.Add(-24 * time.Hour)
		sTime = time.Date(yesterday.Year(), yesterday.Month(), yesterday.Day(), hour, min, sec, 0, time.Local)
	} else {
		sTime = time.Date(now.Year(), now.Month(), now.Day(), hour, min, sec, 0, time.Local)
	}
	start = sTime.Add(-offset).Add(-dur)
	end = sTime.Add(-offset)
	return
}

func getDetail(t *mod.Traffic) *mod.TrafficDetail {
	tc := conf.Conf.Mod.TrafficMoni
	// 预估下载次数
	estimateCount := t.SetUpUserCount * tc.DownloadRate
	// 预估平均下载文件大小
	avgFileSize := calcAvgFileSize(t.File, t.Patches)
	// 预估下载量
	downloadSizeBytes := estimateCount * avgFileSize
	per := (downloadSizeBytes) / t.DownloadSizeOnlineBytes
	var patchFileSize float64
	if len(t.Patches) != 0 {
		patchFileSize = float64(t.Patches[0].Size)
	}
	priority := calcPriority(t.Config)
	tn := &mod.TrafficDetail{
		AppKey:                       t.Pool.AppKey,
		PoolName:                     t.Pool.Name,
		ModName:                      t.Module.Name,
		Operator:                     t.Operator,
		VerNum:                       t.Version.Version,
		OriginFileSize:               utils.HumanFileSize(float64(t.File.Size)),
		PatchFileSize:                utils.HumanFileSize(patchFileSize),
		AvgFileSize:                  utils.HumanFileSize(calcAvgFileSize(t.File, t.Patches)),
		DownloadCount:                int64(estimateCount),
		DownloadSizeEstimate:         utils.HumanFileSize(downloadSizeBytes),
		DownloadSizeOnline:           utils.HumanFileSize(t.DownloadSizeOnlineBytes / tc.SamplingRate),
		DownloadSizeOnlineTotal:      utils.HumanFileSize(downloadSizeBytes + (t.DownloadSizeOnlineBytes / tc.SamplingRate)),
		DownloadCDNBandwidthEstimate: utils.HumanFileSize(downloadSizeBytes * tc.CDNRatio),
		DownloadCDNBandwidthOnline:   utils.HumanFileSize((t.DownloadSizeOnlineBytes / tc.SamplingRate) * tc.CDNRatio),
		DownloadCDNBandwidthTotal:    utils.HumanFileSize((downloadSizeBytes + (t.DownloadSizeOnlineBytes / tc.SamplingRate)) * tc.CDNRatio),
		ModUrl:                       fmt.Sprintf(tc.ModUrl, t.Pool.AppKey, t.Pool.ID, t.Module.ID, string(mod.EnvProd)),
		Percentage:                   fmt.Sprintf("%.3f%%", per*10),
		IsManual:                     priority == mod.PriorityLow,
		Cost:                         calcCost(per, priority),
		Advice:                       getAdvice(t.Config),
		Doc:                          conf.Conf.Mod.TrafficMoni.DocURL,
	}
	return tn
}

func getAdvice(config *mod.Config) []string {
	var (
		advice []string
		err    error
	)
	pFlag, err := isPeak(time.Now(), conf.Conf.Mod.Peak)
	if err != nil {
		log.Error("getAdvice error: %v", err)
		return nil
	}
	if pFlag {
		advice = append(advice, conf.Conf.Mod.TrafficMoni.Advice["peak"])
	}
	if config == nil || config.Priority != mod.PriorityLow {
		advice = append(advice, conf.Conf.Mod.TrafficMoni.Advice["priority"])
	}
	return advice
}

// 当前时间是否在峰值区间
func isPeak(now time.Time, peak []*conf.Peak) (bool, error) {
	parseNow, err := time.Parse(time.Kitchen, now.Format(time.Kitchen))
	if err != nil {
		return false, err
	}
	for _, v := range peak {
		parseStart, err := time.Parse(time.Kitchen, v.Start)
		if err != nil {
			return false, err
		}
		parseEnd, err := time.Parse(time.Kitchen, v.End)
		if err != nil {
			return false, err
		}
		if parseStart.Before(parseNow) && parseNow.Before(parseEnd) {
			return true, err
		}
	}
	return false, err
}

// 计算本次发布的成本
func calcCost(percentage float64, priority mod.Priority) mod.TrafficCost {
	bMap := conf.Conf.Mod.TrafficMoni.Boundary
	if priority == mod.PriorityLow {
		return mod.CostLow
	}
	percentage /= 10
	if percentage < bMap["low"] {
		return mod.CostLow
	} else if bMap["low"] <= percentage && percentage < bMap["middle"] {
		return mod.CostMiddle
	} else if bMap["middle"] <= percentage && percentage < bMap["high"] {
		return mod.CostHigh
	} else {
		return mod.CostVeryHigh
	}
}

// 计算平均下载的文件大小
func calcAvgFileSize(file *mod.File, patches []*mod.Patch) float64 {
	var fileSize float64
	if len(patches) != 0 {
		fileSize = float64(file.Size)*(1-conf.Conf.Mod.TrafficMoni.PatchRate) + float64(patches[0].Size)*conf.Conf.Mod.TrafficMoni.PatchRate
	} else {
		fileSize = float64(file.Size)
	}
	return fileSize
}

func calcPriority(config *mod.Config) mod.Priority {
	if config == nil {
		return mod.PriorityMiddle
	} else {
		return config.Priority
	}
}

// 从patch列表里过滤出生产的patch包
func getProdPatch(versions []*mod.Version, patches []*mod.Patch) []*mod.Patch {
	var (
		versionMap  = make(map[string]*mod.Version)
		prodPatches []*mod.Patch
	)
	for _, v := range versions {
		if v.Env == mod.EnvProd && v.State == mod.VersionSucceeded {
			versionMap[strconv.FormatInt(v.Version, 10)] = v
		}
	}
	for _, p := range patches {
		if _, ok := versionMap[p.FromVer]; ok {
			prodPatches = append(prodPatches, p)
		}
	}
	return prodPatches
}
