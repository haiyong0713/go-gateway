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
	bizapkmdl "go-gateway/app/app-svr/fawkes/service/model/bizapk"

	locgrpc "git.bilibili.co/bapis/bapis-go/community/service/location"
)

func (s *Service) ApkList(c context.Context, param *fkmdl.ApkListParam) ([]*fkmdl.Apk, error) {
	cmm, ok := s.apkCache[param.Sn]
	if !ok {
		return nil, ecode.NotModified
	}
	sm, ok := cmm[param.Env]
	if !ok {
		return nil, ecode.NotModified
	}
	if param.Bundle != "" {
		m, ok := sm[param.Bundle]
		if !ok {
			return nil, ecode.NotModified
		}
		sm = map[string][]*bizapkmdl.Apk{param.Bundle: m}
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
	var res []*fkmdl.Apk
	for _, ks := range sm {
		var data *fkmdl.Apk
		for _, v := range ks {
			// 优先级
			// 传 0 就是 大于等于 0 的都返回
			if param.Priority > v.Priority {
				continue
			}
			if !s.apkFilter(param.Buvid, param.Nt, param.Channel, param.Vn, param.Ov, param.Sn, v.FilterConfig, v.FlowConfig, info, channels) {
				continue
			}
			data = &fkmdl.Apk{Name: v.Name, BundleVer: v.BundleVer, MD5: v.MD5, ApkCdnURL: v.ApkCdnURL, Priority: v.Priority}
			break
		}
		if data == nil {
			continue
		}
		res = append(res, data)
	}
	return res, nil
}

//nolint:gocognit
func (s *Service) apkFilter(buvid, network, channel, vn, ov string, buildID int64, filter *bizapkmdl.FilterConfig, flow *bizapkmdl.FlowConfig, info *locgrpc.InfoComplete, channels map[int64]*fkappmdl.Channel) bool {
	// 自定义 升级比例和设备
	// 内存 TODO
	// 全量 过滤规则
	if filter == nil || flow == nil {
		return false
	}
	switch filter.Status {
	case fawkes.FilterStatusCustom:
		if filter.Device != "" {
			log.Info("apkFilter buvid(%v) filter.Device(%v)", buvid, filter.Device)
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
			return false
		}
	case fawkes.FilterStatusInner:
		// TODO inner user list.
		return false
	}
	if !s.apkSplitFlow(buvid, vn, flow) {
		log.Info("apkFilter apkSplitFlow buvid(%v) version(%v) flow(%v)", buvid, vn, flow)
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
		log.Info("apkFilter ov(%v) filter.ExcludesSystem(%v)", ov, filter.ExcludesSystem)
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
			log.Warn("apkFilter errror %v", err)
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
		log.Info("apkFilter info(%+v) filter.ISP(%v)", info, filter.ISP)
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
	if filter.City != "0" && filter.City != "" {
		log.Info("apkFilter info(%+v) filter.City(%v)", info, filter.City)
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

func (s *Service) apkSplitFlow(buvid, salt string, flow *bizapkmdl.FlowConfig) bool {
	if flow == nil {
		return false
	}
	bucket := int64(s.flowTest(buvid + salt))
	fs, err := xstr.SplitInts(flow.Flow)
	if err != nil {
		log.Error("%+v", err)
		return false
	}
	//nolint:gomnd
	if len(fs) != 2 {
		log.Error("split flow_config lenth no equal 2")
		return false
	}
	if fs[0] == 0 && fs[1] == 0 {
		return false
	}
	return bucket >= fs[0] && bucket <= fs[1]
}
