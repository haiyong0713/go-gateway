package fawkes

import (
	"context"
	"crypto/md5"
	"encoding/json"
	"fmt"
	"hash/crc32"
	"sort"
	"strconv"
	"strings"

	"go-common/library/ecode"
	"go-common/library/log"
	"go-common/library/net/metadata"
	"go-common/library/xstr"

	"go-gateway/app/app-svr/app-resource/interface/model/fawkes"
	fkmdl "go-gateway/app/app-svr/fawkes/service/model"
	fkappmdl "go-gateway/app/app-svr/fawkes/service/model/app"
	fkcdmdl "go-gateway/app/app-svr/fawkes/service/model/cd"
	"go-main/app/ep/hassan/mock/support/slice"

	locgrpc "git.bilibili.co/bapis/bapis-go/community/service/location"
	"github.com/pkg/errors"
)

// Keys sort vids desc.
type Keys []*fkmdl.Version

func (k Keys) Len() int { return len(k) }
func (k Keys) Less(i, j int) bool {
	var iv, jv int64
	if k[i] != nil {
		iv = k[i].VersionCode
	}
	if k[j] != nil {
		jv = k[j].VersionCode
	}
	return iv > jv
}
func (k Keys) Swap(i, j int) { k[i], k[j] = k[j], k[i] }

// Upgrade get upgrade config.
//
//nolint:gocognit
func (s *Service) Upgrade(c context.Context, appKey, env, vs string, build, buildID int64, buvid, network, channel, system string) (res *fawkes.Item, err error) {
	// 特殊逻辑. 过滤 android64的"6070800"对外的鸿蒙包
	if appKey == "android64" && build == 6070800 {
		return
	}
	log.Info("Upgrade appkey(%v) env(%v) version(%v) versionCode(%v) buildID(%v) buvid(%v) network(%v) channel(%v) system(%v) start", appKey, env, vs, build, buildID, buvid, network, channel, system)
	var ip = metadata.String(c, metadata.RemoteIP)
	var info *locgrpc.InfoComplete
	if info, err = s.locDao.InfoComplete(c, ip); err != nil {
		log.Error("%v", err)
		err = nil
	}
	if info != nil {
		if info.Isp != "移动" && info.Isp != "联通" && info.Isp != "电信" {
			info.Isp = "其他"
		}
	}
	var (
		packs map[int64][]*fkcdmdl.Pack
		ok    bool
	)
	key := fmt.Sprintf("%v_%v", appKey, env)
	if packs, ok = s.packCache[key]; !ok {
		log.Warn("Upgrade appkey(%v) env(%v) version(%v) versionCode(%v) buildID(%v) get packs faild, s.packCache(%+v)", appKey, env, vs, build, buildID, s.packCache)
		return
	}
	log.Info("Upgrade appkey(%v) env(%v) version(%v) versionCode(%v) buildID(%v) get packs(%+v)", appKey, env, vs, build, buildID, packs)
	var (
		versionCount int
		versions     map[int64]*fkmdl.Version
		ks           []*fkmdl.Version
	)
	if versions, ok = s.versionCache[key]; !ok {
		log.Warn("Upgrade appkey(%v) env(%v) version(%v) versionCode(%v) buildID(%v) get versions faild, s.versionCache(%+v)", appKey, env, vs, build, buildID, s.versionCache)
		return
	}
	for vid := range packs {
		var version *fkmdl.Version
		if version, ok = versions[vid]; !ok || version == nil {
			continue
		}
		ks = append(ks, version)
	}
	sort.Sort(Keys(ks))
	for _, version := range ks {
		vid := version.ID
		ps, ok := packs[vid]
		if !ok {
			continue
		}
		log.Info("Upgrade appkey(%v) env(%v) version(%v) versionCode(%v) buildID(%v) vid(%v) get ps(%+v)", appKey, env, vs, build, buildID, vid, ps)
		var (
			channels map[int64]*fkappmdl.Channel
			upgrade  *fkcdmdl.UpgradConfig
			upass    bool
			upType   int8
		)
		log.Info("Upgrade appkey(%v) env(%v) version(%v) versionCode(%v) buildID(%v) vid(%v) get version(%+v)", appKey, env, vs, build, buildID, vid, version)
		if version.VersionCode <= build {
			return
		}
		channels = s.channelCache[appKey]
		log.Info("Upgrade appkey(%v) env(%v) version(%v) versionCode(%v) buildID(%v) get channels(%+v)", appKey, env, vs, build, buildID, channels)
		versionCount++
		if _, ok = s.upgrdConfigCache[key]; !ok {
			log.Warn("Upgrade appkey(%v) env(%v) version(%v) versionCode(%v) buildID(%v) vid(%v) get upgrade faild, s.upgrdConfigCache(%+v)", appKey, env, vs, build, buildID, vid, s.upgrdConfigCache)
			continue
		} else {
			if upgrade, ok = s.upgrdConfigCache[key][vid]; !ok {
				log.Warn("Upgrade appkey(%v) env(%v) version(%v) versionCode(%v) buildID(%v) vid(%v) get upgrade faild, s.upgrdConfigCache[key](%+v) !ok", appKey, env, vs, build, buildID, vid, s.upgrdConfigCache[key])
				continue
			}
			log.Info("Upgrade appkey(%v) env(%v) version(%v) versionCode(%v) buildID(%v) vid(%v) get upgrade(%+v)", appKey, env, vs, build, buildID, vid, upgrade)
			//nolint:gomnd
			if versionCount <= 10 {
				log.Info("Upgrade appkey(%v) env(%v) version(%v) versionCode(%v) buildID(%v) start judge upgrade config", appKey, env, vs, build, buildID)
				upType, upass = s.verse(system, build, buildID, upgrade, versions)
				log.Info("Upgrade appkey(%v) env(%v) version(%v) versionCode(%v) buildID(%v) complete verse get upType(%v) upass(%v)", appKey, env, vs, build, buildID, upType, upass)
			}
		}
		for _, p := range ps {
			if p == nil {
				log.Warn("Upgrade appkey(%v) env(%v) version(%v) versionCode(%v) buildID(%v) p is nil", appKey, env, vs, build, buildID)
				continue
			}
			log.Info("Upgrade appkey(%v) env(%v) version(%v) versionCode(%v) buildID(%v) get p(%+v)", appKey, env, vs, build, buildID, p)
			var (
				ok     bool
				bid    = p.BuildID
				filter *fkcdmdl.FilterConfig
				flow   *fkcdmdl.FlowConfig
				patch  *fkcdmdl.Patch
			)
			if _, ok = s.filterConfigCache[key]; !ok {
				log.Warn("Upgrade appkey(%v) env(%v) version(%v) versionCode(%v) buildID(%v) get filter config failed, s.filterConfigCache(%+v)", appKey, env, vs, build, buildID, s.filterConfigCache)
				continue
			}
			if filter, ok = s.filterConfigCache[key][bid]; !ok {
				log.Warn("Upgrade appkey(%v) env(%v) version(%v) versionCode(%v) buildID(%v) get filter config failed s.filterConfigCache[key](%+v) !ok", appKey, env, vs, build, buildID, s.filterConfigCache[key])
				continue
			}
			if filter == nil {
				log.Warn("Upgrade appkey(%v) env(%v) version(%v) versionCode(%v) buildID(%v) filter is nil", appKey, env, vs, build, buildID)
				continue
			}
			if _, ok = s.flowCache[key]; !ok {
				log.Warn("Upgrade appkey(%v) env(%v) version(%v) versionCode(%v) buildID(%v) get flow config failed, s.flowCache(%+v) !ok", appKey, env, vs, build, buildID, s.flowCache)
				continue
			}
			if flow, ok = s.flowCache[key][bid]; !ok {
				log.Warn("Upgrade appkey(%v) env(%v) version(%v) versionCode(%v) buildID(%v) get flow config failed, s.flowCache[key](%+v) !ok", appKey, env, vs, build, buildID, s.flowCache[key])
				continue
			}
			if flow == nil {
				log.Warn("Upgrade appkey(%v) env(%v) version(%v) versionCode(%v) buildID(%v) flow is nil", appKey, env, vs, build, buildID)
				continue
			}
			if _, ok = s.patchCache[appKey]; ok {
				key2 := fmt.Sprintf("%v_%v", bid, buildID)
				patch = s.patchCache[appKey][key2]
			}
			//nolint:gomnd
			if versionCount <= 10 {
				// buvid first.
				if filter.Status == fawkes.FilterStatusCustom {
					if filter.Device != "" {
						log.Info("Upgrade appkey(%v) env(%v) version(%v) versionCode(%v) buildID(%v) start judge device: filter.Device(%v)", appKey, env, vs, build, buildID, filter.Device)
						for _, device := range strings.Split(filter.Device, ",") {
							if buvid == device {
								goto SUCCESS
							}
						}
					}
					if filter.Percent > 0 {
						if bucket := int8(s.flowTest(buvid + strconv.FormatInt(p.BuildID, 10))); bucket < 0 || bucket >= filter.Percent {
							log.Warn("Upgrade appkey(%v) env(%v) version(%v) versionCode(%v) buildID(%v) filter custom flow bucket(%v) percent(%v)", appKey, env, vs, build, buildID, bucket, filter.Percent)
							continue
						}
					}
				} else if filter.Status == fawkes.FilterStatusInner {
					log.Info("Upgrade appkey(%v) env(%v) version(%v) versionCode(%v) buildID(%v) start judge inner user", appKey, env, vs, build, buildID)
					// TODO inner user list.
					continue
				}
				// upgrade config secound.
				if !upass {
					continue
				}
				// flow config third.
				// var salt = fmt.Sprintf("%v(%v)", vs, build)
				var salt = vs
				log.Info("Upgrade appkey(%v) env(%v) version(%v) versionCode(%v) buildID(%v) start judge flow: flow.Flow(%v) buvid(%v) filter.Salt(%v)", appKey, env, vs, build, buildID, flow.Flow, buvid, salt)
				if buvid == "" {
					log.Warn("Upgrade appkey(%v) env(%v) version(%v) versionCode(%v) buildID(%v) buvid is empty", appKey, env, vs, build, buildID)
					continue
				}
				bucket := int64(s.flowTest(buvid + salt))
				fs, err := xstr.SplitInts(flow.Flow)
				if err != nil || len(fs) != 2 {
					log.Warn("Upgrade appkey(%v) env(%v) version(%v) versionCode(%v) buildID(%v) err(%v) fs(%v)", appKey, env, vs, build, buildID, err, fs)
					continue
				}
				if bucket < fs[0] || bucket > fs[1] {
					log.Warn("Upgrade appkey(%v) env(%v) version(%v) versionCode(%v) buildID(%v) bucket(%v) not matched fs(%+v)", appKey, env, vs, build, buildID, bucket, fs)
					continue
				}
				// filter config Fourth.
				log.Info("Upgrade appkey(%v) env(%v) version(%v) versionCode(%v) buildID(%v) start judge filter", appKey, env, vs, build, buildID)
				if !s.chorus(buvid, network, channel, buildID, filter, patch, info, channels) {
					continue
				}
			}
		SUCCESS:
			log.Info("Upgrade key(%v) vid(%v) bid(%v) success pass all config", key, vid, bid)
			if upgrade.Title == "" {
				log.Warn("Upgrade appkey(%v) env(%v) version(%v) versionCode(%v) buildID(%v) up faild title empty", appKey, env, vs, build, buildID)
				continue
			}
			if upgrade.Content == "" {
				log.Warn("Upgrade appkey(%v) env(%v) version(%v) versionCode(%v) buildID(%v) up faild content empty", appKey, env, vs, build, buildID)
				continue
			}
			if version, ok = versions[p.VersionID]; !ok {
				log.Warn("Upgrade appkey(%v) env(%v) version(%v) versionCode(%v) buildID(%v) up faild version faile, versions(%v)", appKey, env, vs, build, buildID, versions)
				continue
			}
			var url string
			if env == "prod" {
				if p.CDNURL == "" {
					log.Warn("Upgrade appkey(%v) env(%v) version(%v) versionCode(%v) buildID(%v) up faild p.CDNURL empty", appKey, env, vs, build, buildID)
					continue
				}
				url = p.CDNURL
			} else {
				if p.PackURL == "" {
					log.Warn("Upgrade appkey(%v) env(%v) version(%v) versionCode(%v) buildID(%v) up faild p.PackURL empty", appKey, env, vs, build, buildID)
					continue
				}
				url = p.PackURL
			}
			if p.Size == 0 {
				log.Warn("Upgrade appkey(%v) env(%v) version(%v) versionCode(%v) buildID(%v) up faild p.Size empty", appKey, env, vs, build, buildID)
				continue
			}
			if p.MD5 == "" {
				log.Warn("Upgrade appkey(%v) env(%v) version(%v) versionCode(%v) buildID(%v) up faild p.MD5 empty", appKey, env, vs, build, buildID)
				continue
			}
			res = &fawkes.Item{
				Titel:       upgrade.Title,
				Content:     upgrade.Content,
				Version:     version.Version,
				VersionCode: version.VersionCode,
				URL:         url,
				Size:        p.Size,
				MD5:         p.MD5,
			}
			log.Info("Upgrade success pass patch(%v)", patch)
			if patch != nil && patch.CDNURL != "" && patch.Size != 0 && patch.MD5 != "" {
				res.Patch = &fawkes.ItemPatch{
					URL:  patch.CDNURL,
					Size: patch.Size,
					MD5:  patch.MD5,
				}
			}
			if upType == 0 {
				upType = 1
			}
			res.Silent = upgrade.IsSilent
			res.UpType = upType
			res.Cycle = upgrade.Cycle
			res.Policy = upgrade.Policy
			res.PolicyURL = upgrade.PolicyURL
			res.PTime = p.PTime
			return
		}
	}
	return
}

//nolint:gocognit
func (s *Service) UpgradeIOS(ctx context.Context, param *fawkes.UpgradeIOSParam) (*fawkes.IOSItem, error) {
	if param.IsTestflight {
		return nil, errors.Wrapf(ecode.NotModified, "%+v", param)
	}
	info, err := s.locDao.InfoComplete(ctx, param.IP)
	if err != nil {
		log.Error("%+v", err)
	}
	if info != nil {
		if info.Isp != "移动" && info.Isp != "联通" && info.Isp != "电信" {
			info.Isp = "其他"
		}
	}
	key := fmt.Sprintf("%v_%v", param.FawkesAppKey, param.FawkesEnv)
	packs, ok := s.packCache[key]
	if !ok {
		return nil, errors.Wrapf(ecode.NotModified, "%+v", param)
	}
	versions, ok := s.versionCache[key]
	if !ok {
		return nil, errors.Wrapf(ecode.NotModified, "%+v", param)
	}
	var ks []*fkmdl.Version
	for vid := range packs {
		var version *fkmdl.Version
		if version, ok = versions[vid]; !ok || version == nil {
			continue
		}
		ks = append(ks, version)
	}
	sort.Sort(Keys(ks))
	for _, version := range ks {
		ps, ok := packs[version.ID]
		if !ok {
			continue
		}
		if param.Build > version.VersionCode {
			return nil, errors.Wrapf(ecode.NotModified, "%+v,%+v", param, version)
		}
		channels := s.channelCache[param.FawkesAppKey]
		upgrade, ok := s.upgrdConfigCache[key][version.ID]
		if !ok {
			continue
		}
		_, upass := s.verse(param.Ov, param.Build, param.Sn, upgrade, versions)
		for _, p := range ps {
			if p == nil {
				continue
			}
			filter := s.filterConfigCache[key][p.BuildID]
			if filter == nil {
				continue
			}
			flow := s.flowCache[key][p.BuildID]
			if flow == nil {
				continue
			}
			var patch *fkcdmdl.Patch
			if _, ok = s.patchCache[param.FawkesAppKey]; ok {
				key2 := fmt.Sprintf("%v_%v", p.BuildID, param.Sn)
				patch = s.patchCache[param.FawkesAppKey][key2]
			}
			if ok := func() bool {
				// buvid first.
				switch filter.Status {
				case fawkes.FilterStatusCustom:
					if filter.Device != "" {
						for _, device := range strings.Split(filter.Device, ",") {
							if param.Buvid == device {
								return true
							}
						}
					}
					if filter.Percent > 0 {
						if bucket := int8(s.flowTest(param.Buvid + strconv.FormatInt(p.BuildID, 10))); bucket < 0 || bucket >= filter.Percent {
							return false
						}
					}
				case fawkes.FilterStatusInner:
					// TODO inner user list.
					return false
				}
				if !upass {
					return false
				}
				if param.Buvid == "" {
					return false
				}
				bucket := int64(s.flowTest(param.Buvid + param.Vn))
				fs, err := xstr.SplitInts(flow.Flow)
				if err != nil || len(fs) != 2 {
					return false
				}
				if bucket < fs[0] || bucket > fs[1] {
					return false
				}
				// filter config Fourth.
				return s.iosChorus(param.Buvid, param.Nt, filter, patch, info, channels, param.Model)
			}(); !ok {
				continue
			}
			if upgrade.Title == "" {
				continue
			}
			if upgrade.Content == "" {
				continue
			}
			if version, ok = versions[p.VersionID]; !ok {
				continue
			}
			if p.Size == 0 {
				continue
			}
			if p.MD5 == "" {
				continue
			}
			res := &fawkes.IOSItem{
				Title:          upgrade.Title,
				Content:        upgrade.Content,
				Version:        version.Version,
				VersionCode:    version.VersionCode,
				PolicyURL:      upgrade.PolicyURL,
				Cycle:          upgrade.Cycle,
				Ptime:          p.PTime,
				IconURL:        upgrade.IconURL,
				ConfirmBtnText: upgrade.ConfirmBtnText,
				CancelBtnText:  upgrade.CancelBtnText,
			}
			return res, nil
		}
	}
	return nil, nil
}

// verse 根据upgrade配置判断是否满足升级的版本条件
func (s *Service) verse(system string, build, buildID int64, upgrade *fkcdmdl.UpgradConfig, versions map[int64]*fkmdl.Version) (upType int8, pass bool) {
	// upgrade config second.
	if upgrade != nil {
		log.Info("Upgrade verse system(%v) upgrade.System(%v)", system, upgrade.System)
		if len(upgrade.ExcludeSystem) > 0 {
			// 系统版本未知
			if len(system) == 0 {
				return upType, false
			}
			var excluded bool
			for _, esv := range strings.Split(upgrade.ExcludeSystem, ",") {
				if system == esv {
					excluded = true
					break
				}
			}
			// 如果是被排除的版本则不允许升级
			if excluded {
				return upType, false
			}
		}
		if len(upgrade.System) > 0 {
			if system != "" {
				var sb bool
				for _, sv := range strings.Split(upgrade.System, ",") {
					if system == sv {
						sb = true
						break
					}
				}
				if !sb {
					return upType, false
				}
			} else {
				return upType, false
			}
		}
		if upgrade.Force == "" && upgrade.Normal == "" {
			return upType, false
		}
		if upgrade.Force != "" {
			upType = fawkes.UpgradeForce
			pass = s.verseUpgradeConfig(upgrade.Force, upgrade.ExForce, build, buildID, upgrade, versions)
		}
		if !pass && upgrade.Normal != "" {
			upType = fawkes.UpgradeNormal
			pass = s.verseUpgradeConfig(upgrade.Normal, upgrade.ExNormal, build, buildID, upgrade, versions)
		}
		return upType, pass
	}
	return fawkes.UpgradeNormal, true
}

func (s *Service) verseUpgradeConfig(upgradeConfig, upgradeExcludeConfig string, build, buildID int64,
	_ *fkcdmdl.UpgradConfig, versions map[int64]*fkmdl.Version) (pass bool) {
	// upgrade config
	log.Info("Upgrade verse buildID(%v) upgradeConfig(%v)", buildID, upgradeConfig)
	var uc *fkcdmdl.UpgradeVersion
	if err := json.Unmarshal([]byte(upgradeConfig), &uc); err != nil {
		log.Warn("%v", err)
		return false
	}
	if uc.Min == 0 && uc.Max == 0 {
		return false
	}
	if uc.Max != 0 && (uc.Min > uc.Max) {
		return false
	}
	if (build < uc.Min) || (uc.Max != 0 && build > uc.Max) {
		return false
	}
	// exclude upgrade config
	if upgradeExcludeConfig != "" {
		log.Info("Upgrade verse buildID(%v) upgradeExcludeConfig(%v)", buildID, upgradeExcludeConfig)
		var uecs []*fkcdmdl.ExcludeUpgradeVersion
		if err := json.Unmarshal([]byte(upgradeExcludeConfig), &uecs); err != nil {
			log.Warn("%v", err)
			return false
		}
		for _, uec := range uecs {
			if uec == nil {
				return false
			}
			log.Info("Upgrade verse buildID(%v) n(%v)", buildID, uec)
			if len(uec.BuildIDs) > 0 {
				for _, bid := range uec.BuildIDs {
					if buildID == bid {
						return false
					}
				}
			} else if version, ok := versions[uec.VersionID]; ok && build == version.VersionCode {
				return false
			}
		}
	}
	return true
}

//nolint:gocognit
func (s *Service) chorus(_, network, channel string, _ int64, filter *fkcdmdl.FilterConfig, _ *fkcdmdl.Patch, info *locgrpc.InfoComplete, channels map[int64]*fkappmdl.Channel) bool {
	if filter.Network != "" {
		log.Info("Upgrade chorus network(%v) filter.Network(%v)", network, filter.Network)
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
	if filter.Channel != "" {
		log.Info("Upgrade chorus channel(%v) filter.Channel(%v)", channel, filter.Channel)
		if channel == "" {
			return false
		}
		chs, err := xstr.SplitInts(filter.Channel)
		if err != nil {
			log.Warn("Upgrade chorus errror %v", err)
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
	if filter.ISP != "" {
		log.Info("Upgrade chorus info(%+v) filter.ISP(%v)", info, filter.ISP)
		if info == nil {
			return false
		}
		var isps []string
		if isps = strings.Split(filter.ISP, ","); len(isps) > 0 {
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
	if filter.City != "" {
		log.Info("Upgrade chorus info(%+v) filter.City(%v)", info, filter.City)
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

//nolint:gocognit
func (s *Service) iosChorus(_, network string, filter *fkcdmdl.FilterConfig, _ *fkcdmdl.Patch, info *locgrpc.InfoComplete, _ map[int64]*fkappmdl.Channel, model string) bool {
	if filter.Network != "" {
		log.Info("ios upgrade chorus network(%v) filter.Network(%v)", network, filter.Network)
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
	if filter.ISP != "" {
		log.Info("ios upgrade chorus info(%+v) filter.ISP(%v)", info, filter.ISP)
		if info == nil {
			return false
		}
		var isps []string
		if isps = strings.Split(filter.ISP, ","); len(isps) > 0 {
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
	if filter.City != "" {
		log.Info("ios upgrade chorus info(%+v) filter.City(%v)", info, filter.City)
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
	if filter.PhoneModel != "" {
		log.Info("ios upgrade chorus model(%v) filter.PhoneModel(%v)", model, filter.PhoneModel)
		if model == "" {
			return false
		}
		mds := strings.Split(filter.PhoneModel, ",")
		var mdb bool
		for _, md := range mds {
			log.Info("ios upgrade chorus model(%v) md(%v)", model, md)
			if model == md {
				mdb = true
				break
			}
		}
		if !mdb {
			return false
		}
	}
	return true
}

func (s *Service) flowTest(buvid string) (id uint32) {
	//nolint:gomnd
	id = crc32.ChecksumIEEE([]byte(strings.ToUpper(buvid))) % 100
	return
}

// HfUpgrade hotfix upgrade
func (s *Service) HfUpgrade(c context.Context, appKey, env, vn, deviceID, channel, appMD5 string, build, sn, ov int64) (res *fawkes.HfUpgradeInfo, err error) {
	var ip = metadata.String(c, metadata.RemoteIP)
	var info *locgrpc.InfoComplete
	if info, err = s.locDao.InfoComplete(c, ip); err != nil {
		log.Error("%v", err)
		err = nil
	}
	var (
		hfUpgrade map[int64][]*fkappmdl.HfUpgrade
		hotfix    []*fkappmdl.HfUpgrade
		ok        bool
	)
	key := fmt.Sprintf("%v_%v", appKey, env)
	if hfUpgrade, ok = s.hfUpgradeCache[key]; !ok {
		err = ecode.NotModified
		return
	}
	if hotfix, ok = hfUpgrade[sn]; !ok {
		err = ecode.NotModified
		return
	}
	var (
		upgrade = &fkappmdl.HfUpgrade{}
		flag    bool
	)
	for _, hf := range hotfix {
		if hf.Config != nil && hf.Config.Effect == 2 && upgrade.BuildID < hf.BuildID {
			upgrade = hf
			flag = true
		}
	}
	//nolint:gosimple
	if flag == false || appMD5 == upgrade.Md5 {
		err = ecode.NotModified
		return
	}
	config := upgrade.Config
	if !s.hfIsUpgrade(config, info, deviceID, channel, sn) {
		err = ecode.NotModified
		return
	}
	res = &fawkes.HfUpgradeInfo{}
	res.Version = vn
	res.VersionCode = sn
	if env == "test" {
		res.PatchURL = upgrade.HotfixURL
	} else {
		res.PatchURL = upgrade.CDNURL
	}
	res.PatchMd5 = upgrade.Md5
	return
}

//nolint:gocognit
func (s *Service) hfIsUpgrade(config *fkappmdl.HotfixConfig, info *locgrpc.InfoComplete, deviceID, channel string, buildID int64) bool {

	if config.Effect != fkappmdl.EffectYes {
		return false
	}
	//nolint:gomnd
	if config.Gray == 1 {
		return true
	} else if config.Gray == 2 {

	} else if config.Gray == 3 {
		if config.Device != "" {
			for _, device := range strings.Split(config.Device, ",") {
				if device == deviceID {
					return true
				}
			}
		}
		if config.UpgradNum == 0 {
			return false
		}
		md5Ctx := md5.New()
		md5Ctx.Write([]byte(deviceID + strconv.FormatInt(buildID, 10)))
		hs := crc32.Checksum(md5Ctx.Sum(nil), crc32.IEEETable)
		var key int
		//nolint:gomnd
		if key = int(hs) % 100; key < 0 {
			key += 100
		}
		if key >= config.UpgradNum {
			return false
		}
		if config.Channel != "" {
			flag := false
			for _, chItem := range strings.Split(config.Channel, ",") {
				if chItem == channel {
					flag = true
					break
				}
			}
			//nolint:gosimple
			if flag == false {
				return false
			}
		}
		if config.City != "" {
			if info == nil {
				return false
			}
			flag := false
		CITY:
			for _, cityID := range strings.Split(config.City, ",") {
				for _, zone := range info.ZoneId {
					zoneID := strconv.FormatInt(zone, 10)
					if cityID == zoneID {
						flag = true
						break CITY
					}
				}
			}
			//nolint:gosimple
			if flag == false {
				return false
			}
		}
		return true
	}
	return false
}

// LaserReport reoprt laster to fawkes.
func (s *Service) LaserReport(c context.Context, taskID int64, status int, url, errMsg, mobiApp, build, md5, rawUposUri string) (err error) {
	if err = s.fkDao.LaserReport(c, taskID, status, url, errMsg, mobiApp, build, md5, rawUposUri); err != nil {
		log.Error("%v", err)
	}
	return
}

// LaserReport2 add 主动触发的laser任务
func (s *Service) LaserReport2(c context.Context, appkey, buvid, url, errMsg, mobiApp, build, md5, rawUposUri string, mid, taskID int64, status int) (res *fawkes.LaserActive, err error) {
	var fkTaskID int64
	if fkTaskID, err = s.fkDao.LaserReport2(c, appkey, buvid, url, errMsg, mobiApp, build, md5, rawUposUri, mid, taskID, status); err != nil {
		log.Error("%v", err)
		return
	}
	res = &fawkes.LaserActive{TaskID: fkTaskID}
	return
}

// LaserReportSilence reoprt laster to fawkes.
func (s *Service) LaserReportSilence(c context.Context, taskID int64, status int, url, errMsg, mobiApp, build string) (err error) {
	if err = s.fkDao.LaserReportSilence(c, taskID, status, url, errMsg, mobiApp, build); err != nil {
		log.Error("%v", err)
		return
	}
	if status == fawkes.StatusUpSuccess && url != "" {
		var (
			laser           *fkappmdl.Laser
			contents, users []string
		)
		if laser, err = s.fkDao.Laser(c, taskID); err != nil {
			log.Error("%v", err)
			return
		}
		if laser == nil {
			return
		}
		users = append(users, s.c.WeChant.Users...)
		users = append(users, laser.Operator)
		//nolint:gosimple
		contents = append(contents, fmt.Sprintf("【Laser通知】静默推送方式执行成功"))
		contents = append(contents, fmt.Sprintf("taskID: %d", taskID))
		contents = append(contents, fmt.Sprintf("platform: %s", laser.Platform))
		contents = append(contents, fmt.Sprintf("mid: %d", laser.MID))
		contents = append(contents, fmt.Sprintf("buvid: %s", laser.Buvid))
		contents = append(contents, fmt.Sprintf("日志时间: %s", laser.LogDate))
		contents = append(contents, fmt.Sprintf("日志文件: %s", url))
		if err = s.alarmDao.SendWeChart(c, strings.Join(contents, "\n	"), users); err != nil {
			log.Error("%v", err)
			err = nil
		}
	}
	return
}

// LaserCmdReport reoprt laser cmd to fawkes.
func (s *Service) LaserCmdReport(c context.Context, taskID int64, status int, mobiApp, build, url, errorMsg, result, md5, rawUposUri string) error {
	return s.fkDao.LaserCmdReport(c, taskID, status, mobiApp, build, url, errorMsg, result, md5, rawUposUri)
}

//nolint:gocognit
func (s *Service) UpgradeTiny(c context.Context, appKey, env, abi string) (*fawkes.Item, error) {
	androidKey := verifyAbi(abi)
	if androidKey == "" {
		log.Error("s.UpgradeTiny abi 异常, abi:%s", abi)
		return nil, nil
	}
	var (
		packs map[int64][]*fkcdmdl.Pack
		ok    bool
	)
	pinkKey := fmt.Sprintf("%v_%v", androidKey, env)
	if packs, ok = s.packCache[pinkKey]; !ok {
		log.Error("s.UpgradeTiny packs is empty, pinkKey:%s", pinkKey)
		return nil, nil
	}
	var (
		versions map[int64]*fkmdl.Version
		ks       []*fkmdl.Version
	)
	if versions, ok = s.versionCache[pinkKey]; !ok {
		log.Error("s.UpgradeTiny versions is empty, pinkKey:%s", pinkKey)
		return nil, nil
	}
	// 获取有效的版本
	for vid := range packs {
		var version *fkmdl.Version
		if version, ok = versions[vid]; !ok || version == nil {
			continue
		}
		ks = append(ks, version)
	}
	sort.Sort(Keys(ks))
	var (
		pinkPack    *fkcdmdl.Pack
		pinkVersion *fkmdl.Version
		url         string
	)
	func() {
		for _, version := range ks {
			var ps []*fkcdmdl.Pack
			if ps, ok = packs[version.ID]; !ok {
				continue
			}
			for _, p := range ps {
				if p == nil {
					continue
				}
				var (
					ok     bool
					bid    = p.BuildID
					filter *fkcdmdl.FilterConfig
					flow   *fkcdmdl.FlowConfig
				)
				if filter, ok = s.filterConfigCache[pinkKey][bid]; !ok || filter == nil {
					continue
				}
				if flow, ok = s.flowCache[pinkKey][bid]; !ok || flow == nil {
					continue
				}
				fs := strings.Split(flow.Flow, ",")
				//nolint:gomnd
				if len(fs) != 2 {
					continue
				}
				if fs[0] == "0" && fs[1] == "99" && filter.Status == fawkes.FilterStatusAll {
					if env == "prod" {
						if p.CDNURL == "" {
							continue
						}
						url = p.CDNURL
					} else {
						if p.PackURL == "" {
							continue
						}
						url = p.PackURL
					}
					if p.Size == 0 {
						continue
					}
					if p.MD5 == "" {
						continue
					}
					pinkPack = &fkcdmdl.Pack{}
					pinkVersion = &fkmdl.Version{}
					*pinkPack = *p
					*pinkVersion = *version
					return
				}
			}
		}
	}()
	if pinkPack == nil || pinkVersion == nil {
		return nil, nil
	}
	var (
		upgradeMap map[int64]*fkcdmdl.UpgradConfig
		upgrades   []*fkcdmdl.UpgradConfig
		tinyKey    = fmt.Sprintf("%v_%v", appKey, env)
	)
	if upgradeMap, ok = s.upgrdConfigCache[tinyKey]; !ok {
		log.Error("s.UpgradeTiny upgrade config is empty, tinyKey:%s", tinyKey)
		return nil, nil
	}
	for _, upgrade := range upgradeMap {
		upgrades = append(upgrades, upgrade)
	}
	sort.Slice(upgrades, func(i, j int) bool {
		var iv, jv int64
		if upgrades[i] != nil {
			iv = upgrades[i].VersionID
		}
		if upgrades[j] != nil {
			jv = upgrades[j].VersionID
		}
		return iv > jv
	})
	upgrade := &fkcdmdl.UpgradConfig{}
	for _, u := range upgrades {
		if u.Title == "" {
			continue
		}
		if u.Content == "" {
			continue
		}
		upgrade = u
	}
	res := &fawkes.Item{
		Titel:       upgrade.Title,
		Content:     upgrade.Content,
		Version:     pinkVersion.Version,
		VersionCode: pinkVersion.VersionCode,
		URL:         url,
		Size:        pinkPack.Size,
		MD5:         pinkPack.MD5,
		Silent:      upgrade.IsSilent,
		Cycle:       upgrade.Cycle,
		Policy:      upgrade.Policy,
		PolicyURL:   upgrade.PolicyURL,
		PTime:       pinkPack.PTime,
	}
	return res, nil
}

func verifyAbi(abi string) string {
	android32 := []string{"armeabi-v7a", "x86", "mips"}
	android64 := []string{"arm64-v8a", "x86_64"}
	if slice.Contains(android32, abi) {
		return "android"
	}
	if slice.Contains(android64, abi) {
		return "android64"
	}
	return ""
}
