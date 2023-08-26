package fawkes

import (
	"context"

	"hash/crc32"
	"strconv"
	"time"

	"go-common/library/ecode"
	"go-common/library/log"

	fkmdl "go-gateway/app/app-svr/app-resource/interface/model/fawkes"
)

func (s *Service) loadTestFlight() {
	log.Info("start load fawkes testFlight")
	//nolint:gosimple
	var resTmp map[string]map[string]*fkmdl.TestFlight
	resTmp = make(map[string]map[string]*fkmdl.TestFlight)
	envs := []string{"test", "prod"}
	for _, env := range envs {
		envTmp, err := s.fkDao.TestFlight(context.Background(), env)
		if err != nil {
			log.Error("%+v", err)
			continue
		}
		resTmp[env] = envTmp
	}
	s.testFlightCache = resTmp
}

func (s *Service) TestFlight(param *fkmdl.TestFlightParam, env string, mid int64) (*fkmdl.TestFlightResult, error) {
	switch param.IsTestflight {
	case 0:
		return s.TestFlightOnline(param, env, mid)
	case 1:
		return s.TestFlightTF(param, env)
	}
	return nil, ecode.NotModified
}

func (s *Service) TestFlightOnline(param *fkmdl.TestFlightParam, env string, mid int64) (*fkmdl.TestFlightResult, error) {
	tfEnvData, ok := s.testFlightCache[env]
	if !ok {
		return nil, ecode.NotModified
	}
	tf, ok := tfEnvData[param.MobiApp]
	if !ok {
		return nil, ecode.NotModified
	}
	if tf.TestFlightPack == nil || param.Build >= tf.TestFlightPack.VersionCode {
		return nil, ecode.NotModified
	}
	res := &fkmdl.TestFlightResult{
		IsForce:     false,
		URL:         tf.TestFlightPack.UpdateURL,
		Desc:        tf.TestFlightPack.GuideText,
		Version:     tf.TestFlightPack.VersionCode,
		PackageType: tf.TestFlightPack.PackageType,
		IsWhite:     false,
	}
	// 黑名单
	if tf.BlackList != nil {
		for _, bmid := range tf.BlackList {
			if mid == bmid {
				return nil, ecode.NotModified
			}
		}
	}
	// 白名单
	if tf.WhiteList != nil {
		for _, wmid := range tf.WhiteList {
			if mid == wmid {
				res.IsWhite = true
				return res, nil
			}
		}
	}
	// 千分桶
	if crc32.ChecksumIEEE([]byte(param.Buvid+strconv.FormatInt(tf.TestFlightPack.VersionCode, 10)))%1000 >= tf.TestFlightPack.DisPermil {
		return nil, ecode.NotModified
	}
	return res, nil
}

func (s *Service) TestFlightTF(param *fkmdl.TestFlightParam, env string) (*fkmdl.TestFlightResult, error) {
	tfEnvData, ok := s.testFlightCache[env]
	if !ok {
		return nil, ecode.NotModified
	}
	tf, ok := tfEnvData[param.MobiApp]
	if !ok {
		return nil, ecode.NotModified
	}
	if tf.OnlinePack == nil && tf.TestFlightPack == nil {
		return nil, ecode.NotModified
	}
	var (
		url, forceDesc, remindDesc, packageType string
		version                                 int64
	)
	if tf.OnlinePack != nil {
		url = tf.OnlinePack.UpdateURL
		forceDesc = tf.OnlinePack.ForceText
		remindDesc = tf.OnlinePack.RemindText
		version = tf.OnlinePack.VersionCode
		packageType = tf.OnlinePack.PackageType
	}
	if tf.TestFlightPack != nil && param.Build < tf.TestFlightPack.VersionCode {
		url = tf.TestFlightPack.UpdateURL
		forceDesc = tf.TestFlightPack.ForceText
		remindDesc = tf.TestFlightPack.RemindText
		version = tf.TestFlightPack.VersionCode
		packageType = tf.TestFlightPack.PackageType
	}
	// 是否符合强制更新时间标准
	for _, tfTmp := range tf.Packs {
		if tfTmp == nil {
			continue
		}
		if param.Build != tfTmp.VersionCode {
			continue
		}
		var now = time.Now().Unix()
		// 满足强制更新条件
		if now >= tfTmp.ForceTime.Time().Unix() {
			return &fkmdl.TestFlightResult{IsForce: true, URL: url, Desc: forceDesc, Version: version, PackageType: packageType}, nil
		}
		// 满足建议更新条件
		if now >= tfTmp.RemindTime.Time().Unix() {
			return &fkmdl.TestFlightResult{IsForce: false, URL: url, Desc: remindDesc, Version: version, PackageType: packageType}, nil
		}
		break
	}
	return nil, ecode.NotModified
}
