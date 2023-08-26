package fawkes

import (
	"context"
	"strconv"
	"strings"

	"go-common/library/ecode"
	"go-common/library/log"
	"go-common/library/net/metadata"
	"go-common/library/xstr"
	"go-gateway/app/app-svr/app-resource/interface/model/fawkes"
	fkmdl "go-gateway/app/app-svr/app-resource/interface/model/fawkes"
	fkappmdl "go-gateway/app/app-svr/fawkes/service/model/app"
	tribemdl "go-gateway/app/app-svr/fawkes/service/model/tribe"

	locgrpc "git.bilibili.co/bapis/bapis-go/community/service/location"
)

func (s *Service) TribeAllList(c context.Context, param *fkmdl.TribeListParam) ([]*fkmdl.TribeApk, error) {
	appBuilds, ok := s.tribeCache[param.AppKey]
	if !ok {
		return nil, ecode.NotModified
	}
	var useableDevID int64
	builds, ok := appBuilds[param.HostVer]
	if !ok {
		useableDevID = s.findDevBuildID(param.HostVer, appBuilds, 1000)
		if useableDevID != 0 {
			builds = appBuilds[useableDevID]
		} else {
			return nil, ecode.NotModified
		}
	}
	tribes, ok := builds[param.Env]
	if !ok {
		return nil, ecode.NotModified
	}
	// 通用组件
	commonTribe := s.getCommonTribe(appBuilds, param.Env)
	cTribeMap := map[string]bool{}
	resTribes := map[string][]*tribemdl.TribeApk{}
	for tName, tribe := range tribes {
		resTribes[tName] = tribe
	}
	for ctName, cts := range commonTribe {
		cTribeMap[ctName] = true
		_, ok := tribes[ctName]
		if !ok {
			resTribes[ctName] = cts
		}
	}
	if param.Bundle != "" {
		tribe, ok := resTribes[param.Bundle]
		if !ok {
			return nil, ecode.NotModified
		}
		resTribes = map[string][]*tribemdl.TribeApk{param.Bundle: tribe}
	}
	ip := metadata.String(c, metadata.RemoteIP)
	info, err := s.locDao.InfoComplete(c, ip)
	if err != nil {
		log.Error("%v", err)
		//nolint:ineffassign
		err = nil
	}
	if info != nil {
		if info.Isp != "移动" && info.Isp != "联通" && info.Isp != "电信" {
			info.Isp = "其他"
		}
	}
	channels := s.channelCache[param.AppKey]
	var res []*fkmdl.TribeApk
	for tribeName, tribeList := range resTribes {
		var data *fkmdl.TribeApk
		// 如果是通用组件 则不需要向上寻找兼容组件
		for _, v := range tribeList {
			// 优先级
			// 传 0 就是 大于等于 0 的都返回
			if param.Priority > v.Priority {
				continue
			}
			if !s.tribeFilter(param.Buvid, param.Nt, param.Channel, param.Ov, param.HostVer, v.FilterConfig, info, channels) {
				log.Error("tribeFilter(%v)失败: %+v", tribeName, v.FilterConfig)
				continue
			}
			if v.Nohost && !s.tribeVer(param.HostVer, v.UpgradeConfig) {
				log.Error("tribeVer(%v)版本不符合: %+v", tribeName, v.UpgradeConfig)
				continue
			}
			data = &fkmdl.TribeApk{Name: v.Name, BundleVer: v.BundleVer, MD5: v.MD5, ApkCdnURL: v.ApkCdnURL, Priority: v.Priority}
			break
		}
		log.Info("findTargetTribe(%+v)", data)
		if data == nil {
			log.Info("not_found_target_tribe")
			if _, ok := cTribeMap[tribeName]; !ok {
				log.Info("(%v) is not commonTribe", tribeName)
				preBuildID := param.HostVer
				if useableDevID != 0 {
					preBuildID = useableDevID
				}
				data, _ = s.findSingleUseableTribe(appBuilds, param, preBuildID, tribeName, info, 1000)
				log.Info("findSingleUseableTribe(%+v)", data)
				if data != nil {
					res = append(res, data)
				}
			}
			continue
		}
		res = append(res, data)
	}
	return res, nil
}

func (s *Service) findSingleUseableTribe(appBuilds map[int64]map[string]map[string][]*tribemdl.TribeApk, param *fkmdl.TribeListParam, buildID int64, tribeName string, info *locgrpc.InfoComplete, recursionLimit int) (*fkmdl.TribeApk, error) {
	if recursionLimit <= 0 {
		log.Error("findSingleUseableTribe_recursionLimit(%v)", recursionLimit)
		return nil, ecode.NotModified
	}
	recursionLimit = recursionLimit - 1
	useableDevID := s.findDevBuildID(buildID, appBuilds, 1000)
	if useableDevID == 0 {
		return nil, ecode.NotModified
	}
	builds := appBuilds[useableDevID]
	tribes, ok := builds[param.Env]
	if !ok {
		return nil, ecode.NotModified
	}
	tribeList, ok := tribes[tribeName]
	if !ok {
		return nil, ecode.NotModified
	}
	channels := s.channelCache[param.AppKey]
	var data *fkmdl.TribeApk
	// 如果是通用组件 则不需要向上寻找兼容组件
	for _, v := range tribeList {
		// 优先级
		// 传 0 就是 大于等于 0 的都返回
		if param.Priority > v.Priority {
			continue
		}
		if !s.tribeFilter(param.Buvid, param.Nt, param.Channel, param.Ov, param.HostVer, v.FilterConfig, info, channels) {
			continue
		}
		data = &fkmdl.TribeApk{Name: v.Name, BundleVer: v.BundleVer, MD5: v.MD5, ApkCdnURL: v.ApkCdnURL, Priority: v.Priority}
		break
	}
	if data == nil {
		preBuildID := buildID
		if useableDevID != 0 {
			preBuildID = useableDevID
		}
		return s.findSingleUseableTribe(appBuilds, param, preBuildID, tribeName, info, recursionLimit)
	}
	return data, nil
}
func (s *Service) getCommonTribe(appBuilds map[int64]map[string]map[string][]*tribemdl.TribeApk, env string) map[string][]*tribemdl.TribeApk {
	commonTribes, ok := appBuilds[-1]
	if !ok {
		return nil
	}
	cTribes, ok := commonTribes[env]
	if !ok {
		return nil
	}
	return cTribes
}
func (s *Service) findDevBuildID(curID int64, appBuilds map[int64]map[string]map[string][]*tribemdl.TribeApk, recursionLimit int) (devID int64) {
	if recursionLimit <= 0 {
		log.Error("findDevBuildID_recursionLimit(%v)", recursionLimit)
		return
	}
	recursionLimit = recursionLimit - 1
	devID, ok := s.tribeRelationCache[curID]
	if !ok || devID == curID {
		return 0
	}
	_, ok = appBuilds[devID]
	if !ok {
		return s.findDevBuildID(devID, appBuilds, recursionLimit)
	}
	return
}

func (s *Service) tribeVer(hostVer int64, upConf *tribemdl.PackUpgrade) bool {
	if upConf == nil {
		return true
	}
	if upConf != nil && len(upConf.ChosenVersionCode) == 0 && len(upConf.ChosenVersionCode) == 0 {
		return true
	}
	if len(upConf.ChosenVersionCode) != 0 {
		verArr := strings.Split(upConf.ChosenVersionCode, ",")
		for _, v := range verArr {
			if vint, err := strconv.ParseInt(v, 10, 64); err == nil && vint == hostVer {
				return true
			}
		}
	}
	if len(upConf.StartVersionCode) != 0 {
		verArr := strings.Split(upConf.StartVersionCode, ",")
		if len(verArr) == 0 {
			return false
		}
		// 通用组件只需要对比起始版本大小，
		sv, err := strconv.ParseInt(verArr[0], 10, 64)
		if err != nil {
			return false
		}
		return hostVer >= sv
	}
	return false
}

//nolint:gocognit
func (s *Service) tribeFilter(buvid, network, channel, ov string, buildID int64, filter *tribemdl.ConfigFilter, info *locgrpc.InfoComplete, channels map[int64]*fkappmdl.Channel) bool {
	// 自定义 升级比例和设备
	// 内存 TODO
	// 全量 过滤规则
	log.Info("tribeFilterConfig:%+v", filter)
	if filter == nil {
		return false
	}
	switch filter.Type {
	case fawkes.FilterStatusCustom:
		log.Info("customMode filter.Type(%v)", filter.Type)
		if filter.Device != "" {
			log.Info("tribeFilter buvid(%v) filter.Device(%v)", buvid, filter.Device)
			for _, device := range strings.Split(filter.Device, ",") {
				if buvid == device {
					return true
				}
			}
		}
		if filter.Percent <= 0 {
			return false
		}
		if bucket := int8(s.flowTest(buvid + strconv.FormatInt(buildID, 10))); bucket < 0 || bucket >= filter.Percent {
			log.Info("bucketFailed:(%v),(%v)", filter.Type, bucket)
			return false
		}
	case fawkes.FilterStatusInner:
		log.Info("innerTest(%v)", filter.Type)
		// TODO inner user list.
		return false
	}
	if filter.Network != "" {
		log.Info("apkFilter network(%v) filter.Network(%v)", network, filter.Network)
		if network == "" {
			return false
		}
		var nts []string
		if nts = strings.Split(filter.Network, ","); len(nts) > 0 {
			var ntb bool
			for _, nt := range nts {
				if network == nt {
					ntb = true
					break
				}
			}
			if !ntb {
				return false
			}
		}
	}
	if filter.ExcludesSystem != "" {
		log.Info("tribeFilter ov(%v) filter.ExcludesSystem(%v)", ov, filter.ExcludesSystem)
		if ov != "" {
			ess := strings.Split(filter.ExcludesSystem, ",")
			for _, es := range ess {
				if ov == es {
					return false
				}
			}
		}
	}
	if filter.Channel != "" {
		log.Info("apkFilter channel(%v) filter.Channel(%v)", channel, filter.Channel)
		if channel == "" {
			return false
		}
		chs, err := xstr.SplitInts(filter.Channel)
		if err != nil {
			log.Error("apkFilter error %v", err)
			return false
		}
		var chb bool
		for _, ch := range chs {
			log.Info("Upgrade channel(%v) ch(%v) channels(%+v)", channel, ch, channels)
			if c, ok := channels[ch]; ok && channel == c.Code {
				chb = true
				break
			}
		}
		if !chb {
			return false
		}
	}
	if filter.Isp != "" {
		log.Info("tribeFilter info(%+v) filter.ISP(%v)", info, filter.Isp)
		if info == nil {
			return false
		}
		var isps []string
		if isps = strings.Split(filter.Isp, ","); len(isps) > 0 {
			var ispb bool
			for _, isp := range isps {
				if info.Isp == isp {
					ispb = true
					break
				}
			}
			if !ispb {
				return false
			}
		}
	}
	if filter.City != "0" && filter.City != "" {
		log.Info("tribeFilter info(%+v) filter.City(%v)", info, filter.City)
		if info == nil {
			return false
		}
		zoonIDs, err := xstr.SplitInts(filter.City)
		if err != nil {
			log.Warn("%v", err)
			return false
		}
		var cityb bool
		for _, zoneID := range info.ZoneId {
			for _, cz := range zoonIDs {
				if zoneID == cz {
					cityb = true
					break
				}
			}
			if cityb {
				break
			}
		}
		if !cityb {
			return false
		}
	}
	return true
}
