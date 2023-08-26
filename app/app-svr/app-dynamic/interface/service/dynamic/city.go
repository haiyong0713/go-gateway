package dynamic

import (
	"context"
	"time"

	"go-common/library/conf/env"
	"go-common/library/exp/ab"
	"go-common/library/log"

	"go-common/library/log/infoc.v2"
	"go-gateway/app/app-svr/app-dynamic/interface/api"
	"go-gateway/app/app-svr/app-dynamic/interface/model"
	dynmdl "go-gateway/app/app-svr/app-dynamic/interface/model/dynamic"

	accountgrpc "git.bilibili.co/bapis/bapis-go/account/service"
)

const (
	// abtest
	_abDynAll   = "dyn_tab_all"
	_abDynVideo = "dyn_tab_video"
	_abMiss     = "miss"
	_abHit      = "hit"
)

var (
	tabVideo = &api.DynTab{Title: "视频", Uri: "bilibili://following/index/8", Anchor: dynmdl.AnchorVideo}
	tabAll   = &api.DynTab{Title: "综合", Uri: "bilibili://following/index/268435455", DefaultTab: true, Anchor: dynmdl.AnchorAll}
	tabAllV2 = &api.DynTab{Title: "关注", Uri: "bilibili://following/index/268435455", DefaultTab: true, Anchor: dynmdl.AnchorAll}
)

// 动态tab
func (s *Service) DynTab(c context.Context, mid int64, header *dynmdl.Header, req *api.DynTabReq) (res *api.DynTabReply, err error) {
	var (
		acc *accountgrpc.Profile
	)
	if mid > 0 {
		var err3 error
		acc, err3 = s.accDao.Profile3(c, mid)
		if err != nil {
			log.Error("%v", err3)
		}
	}
	res = &api.DynTabReply{}
	dynTabAll := tabAll
	// abtest 在实验内，综合改为关注
	if s.tabAbtest(c, acc, _abDynAll, _abHit, model.TabAbtestAllflag) {
		dynTabAll = tabAllV2
	}
	// 青少年只返回综合tab
	if req.TeenagersMode == 1 {
		res.DynTab = append(res.DynTab, dynTabAll)
		return
	}
	// 在实验组内只展示关注一个标签，视频标签不下发
	if s.tabAbtest(c, acc, _abDynVideo, _abHit, model.TabAbtestVedioflag) {
		res.DynTab = append(res.DynTab, tabAllV2)
	} else {
		res.DynTab = append(res.DynTab, tabVideo, dynTabAll)
	}
	// 国际版只返回视频和综合tab
	if header.MobiApp == "iphone_i" || header.MobiApp == "android_i" {
		return
	}
	// 灰度和白名单判断
	if s.c.Grayscale != nil && s.c.Grayscale.Tab != nil && s.c.Grayscale.Tab.Switch {
		switch s.c.Grayscale.Tab.GrayCheck(mid, header.Buvid) {
		case 1:
			return
		}
	}
	return
}

func (s *Service) GeoCoder(c context.Context, lat, lng float64, from string) (*api.GeoCoderReply, error) {
	resTmp, err := s.GeoDao.GeoCoder(c, lat, lng, from)
	if err != nil {
		log.Warn("%+v", err)
		return nil, err
	}
	var res = new(api.GeoCoderReply)
	// 整理返回结构
	if resTmp != nil {
		res.Address = resTmp.Address
		if resTmp.AddressComponent != nil {
			res.AddressComponent = &api.AddressComponent{
				Nation:       resTmp.AddressComponent.Nation,
				Province:     resTmp.AddressComponent.Province,
				City:         resTmp.AddressComponent.City,
				District:     resTmp.AddressComponent.District,
				Street:       resTmp.AddressComponent.Street,
				StreetNumber: resTmp.AddressComponent.StreetNumber,
			}
		}
		if resTmp.AdInfo != nil {
			res.AdInfo = &api.AdInfo{
				NationCode: resTmp.AdInfo.NationCode,
				Adcode:     resTmp.AdInfo.Adcode,
				CityCode:   resTmp.AdInfo.CityCode,
				Name:       resTmp.AdInfo.Name,
			}
			if resTmp.AdInfo.Location != nil {
				res.AdInfo.Gps = &api.Gps{
					Lat: resTmp.AdInfo.Location.Lat,
					Lng: resTmp.AdInfo.Location.Lng,
				}
			}
		}
	}
	return res, nil
}

func (s *Service) tabAbtest(c context.Context, acc *accountgrpc.Profile, abtestName, abValue string, flag *ab.StringFlag) bool {
	if acc == nil {
		return false
	}
	// 当前用户的注册时间小于实验时间则跳出实验
	if int64(acc.JoinTime) < s.c.Resource.TabAbtestUserTime {
		return false
	}
	var (
		groupID int64
	)
	t, ok := ab.FromContext(c)
	if !ok {
		return false
	}
	t.Add(ab.KVInt("mid", acc.Mid))
	exp := flag.Value(t)
	if exp == _abMiss {
		return false
	}
	for _, state := range t.Snapshot() {
		if state.Type == ab.ExpHit {
			groupID = state.Value
			break
		}
	}
	event := infoc.NewLogStreamV(s.c.Infoc.TabAbLogID,
		log.Int64(acc.Mid),        // 用户uid
		log.String(abtestName),    // 实验变量名
		log.String(exp),           // 实验值
		log.Int64(groupID),        // 实验分组; 未命中是0
		log.String(env.DeployEnv), // 环境
		log.String(time.Now().Format("2006-01-02 15:04:05"))) // 时间
	if err := s.svideoInfoc.Info(c, event); err != nil {
		log.Error("[Lancer] HotTest report failed for log_id:%s, err:%+v", s.c.Infoc.TabAbLogID, err)
	}
	if exp != abValue {
		return false
	}
	return true
}
