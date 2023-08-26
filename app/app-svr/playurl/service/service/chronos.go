package service

import (
	"context"
	"hash/crc32"

	v2 "go-gateway/app/app-svr/playurl/service/api/v2"
	arcmdl "go-gateway/app/app-svr/playurl/service/model/archive"
)

func (s *Service) ChronosPkg(c context.Context, req *v2.ChronosPkgReq) (*v2.ChronosPkgReply, error) {
	chronos := s.checkChronos(req)
	if chronos == nil {
		return new(v2.ChronosPkgReply), nil
	}
	return &v2.ChronosPkgReply{
		Md5:  chronos.Md5,
		File: chronos.File,
	}, nil
}

func (s *Service) checkChronos(req *v2.ChronosPkgReq) *v2.Chronos {
	// 6.6.0版本开始chronos功能迁移至ViewProgress接口实现
	// feature Chronos
	if req == nil || (req.MobiApp == "iphone" && req.Build > s.c.Custom.IOSChronosBuild) || (req.MobiApp == "android" && req.Build > s.c.Custom.AndChronosBuild) {
		return nil
	}
	conf := s.chronosConf
	for _, v := range conf {
		if v == nil {
			continue
		}
		// avid check
		if !v.AllAvids && !arcmdl.IsInIDs(v.Avids, req.Aid) {
			continue
		}
		// mid check
		if !v.AllMids && !arcmdl.IsInIDs(v.Mids, req.Mid) {
			continue
		}
		// build check
		bis, ok := v.BuildLimit[req.Platform]
		if !ok || len(bis) == 0 {
			continue
		}
		if ok := func() bool {
			for _, b := range bis {
				// 有配置限制才下发皮肤
				if arcmdl.InvalidBuild(req.Build, int32(b.Value), b.Condition) {
					// 有一个版本校验不通过时，则认为不满足条件
					return false
				}
			}
			return true
		}(); !ok {
			continue
		}
		// gray check
		if crc32.ChecksumIEEE([]byte(req.Buvid))%arcmdl.MaxGray < uint32(v.Gray) {
			return &v2.Chronos{File: v.File, Md5: v.MD5}
		}
	}
	return nil
}
