package region

import (
	"context"
	"sort"
	"strconv"
	"time"

	"go-common/library/log"

	"go-gateway/app/app-svr/app-show/interface/model"
	"go-gateway/app/app-svr/app-show/interface/model/banner"
	feature "go-gateway/app/app-svr/feature/service/sdk"
	resource "go-gateway/app/app-svr/resource/service/model"

	locgrpc "git.bilibili.co/bapis/bapis-go/community/service/location"
)

var (
	_banners = map[int]map[int8]int{
		13: {
			model.PlatIPhone:   454,
			model.PlatIPad:     788,
			model.PlatAndroid:  617,
			model.PlatIPhoneI:  1022,
			model.PlatAndroidG: 1360,
			model.PlatAndroidI: 1791,
			model.PlatIPadI:    1192,
		},
		1: {
			model.PlatIPhone:   453,
			model.PlatIPad:     787,
			model.PlatAndroid:  616,
			model.PlatIPhoneI:  1017,
			model.PlatAndroidG: 1355,
			model.PlatAndroidI: 1785,
			model.PlatIPadI:    1187,
		},
		3: {
			model.PlatIPhone:   455,
			model.PlatIPad:     789,
			model.PlatAndroid:  618,
			model.PlatIPhoneI:  1028,
			model.PlatAndroidG: 1366,
			model.PlatAndroidI: 1798,
			model.PlatIPadI:    1198,
		},
		129: {
			model.PlatIPhone:   456,
			model.PlatIPad:     790,
			model.PlatAndroid:  619,
			model.PlatIPhoneI:  1033,
			model.PlatAndroidG: 1371,
			model.PlatAndroidI: 1804,
			model.PlatIPadI:    1203,
		},
		4: {
			model.PlatIPhone:   457,
			model.PlatIPad:     791,
			model.PlatAndroid:  620,
			model.PlatIPhoneI:  1038,
			model.PlatAndroidG: 1376,
			model.PlatAndroidI: 1810,
			model.PlatIPadI:    1208,
		},
		36: {
			model.PlatIPhone:   458,
			model.PlatIPad:     792,
			model.PlatAndroid:  621,
			model.PlatIPhoneI:  1043,
			model.PlatAndroidG: 1381,
			model.PlatAndroidI: 1816,
			model.PlatIPadI:    1213,
		},
		160: {
			model.PlatIPhone:   459,
			model.PlatIPad:     793,
			model.PlatAndroid:  622,
			model.PlatIPhoneI:  1048,
			model.PlatAndroidG: 1386,
			model.PlatAndroidI: 1822,
			model.PlatIPadI:    1218,
		},
		119: {
			model.PlatIPhone:   460,
			model.PlatIPad:     794,
			model.PlatAndroid:  623,
			model.PlatIPhoneI:  1053,
			model.PlatAndroidG: 1391,
			model.PlatAndroidI: 1828,
			model.PlatIPadI:    1223,
		},
		155: {
			model.PlatIPhone:   462,
			model.PlatIPad:     795,
			model.PlatAndroid:  624,
			model.PlatIPhoneI:  1058,
			model.PlatAndroidG: 1396,
			model.PlatAndroidI: 1834,
			model.PlatIPadI:    1228,
		},
		5: {
			model.PlatIPhone:   463,
			model.PlatIPad:     796,
			model.PlatAndroid:  625,
			model.PlatIPhoneI:  1063,
			model.PlatAndroidG: 1401,
			model.PlatAndroidI: 1840,
			model.PlatIPadI:    1233,
		},
		23: {
			model.PlatIPhone:   464,
			model.PlatIPad:     797,
			model.PlatAndroid:  626,
			model.PlatIPhoneI:  1068,
			model.PlatAndroidG: 1406,
			model.PlatAndroidI: 1846,
			model.PlatIPadI:    1238,
		},
		11: {
			model.PlatIPhone:   465,
			model.PlatIPad:     798,
			model.PlatAndroid:  627,
			model.PlatIPhoneI:  1073,
			model.PlatAndroidG: 1411,
			model.PlatAndroidI: 1852,
			model.PlatIPadI:    1243,
		},
		655: {
			model.PlatIPhone:   466,
			model.PlatIPad:     799,
			model.PlatAndroid:  628,
			model.PlatIPhoneI:  1079,
			model.PlatAndroidG: 1417,
			model.PlatAndroidI: 1859,
			model.PlatIPadI:    1249,
		},
		165: {
			model.PlatIPhone:   1473,
			model.PlatIPad:     1485,
			model.PlatAndroid:  1479,
			model.PlatIPhoneI:  1491,
			model.PlatAndroidG: 1497,
			model.PlatAndroidI: 1873,
			model.PlatIPadI:    1503,
		},
		167: {
			model.PlatIPhone:  1934,
			model.PlatIPad:    1932,
			model.PlatAndroid: 1933,
		},
		181: {
			model.PlatIPhone:  2225,
			model.PlatIPad:    2239,
			model.PlatAndroid: 2232,
		},
		177: {
			model.PlatIPhone:  2275,
			model.PlatIPad:    2289,
			model.PlatAndroid: 2282,
		},
		188: {
			model.PlatIPhone:   2996,
			model.PlatIPad:     3008,
			model.PlatAndroid:  3002,
			model.PlatIPhoneI:  3014,
			model.PlatAndroidG: 3020,
			model.PlatAndroidI: 3032,
			model.PlatIPadI:    3026,
		},
		211: {
			model.PlatIPhone:  4271,
			model.PlatIPad:    4295,
			model.PlatAndroid: 4283,
		},
		217: {
			model.PlatIPhone:  4388,
			model.PlatIPad:    4400,
			model.PlatAndroid: 4394,
		},
		223: {
			model.PlatIPhone:   4448,
			model.PlatAndroid:  4454,
			model.PlatIPad:     4460,
			model.PlatIPhoneI:  4466,
			model.PlatAndroidG: 4472,
			model.PlatIPadI:    4478,
			model.PlatAndroidI: 4484,
		},
		234: {
			model.PlatIPhone:  4659,
			model.PlatAndroid: 4665,
			model.PlatIPad:    4671,
		},
	}
	_bannersPlat = map[int8]string{
		model.PlatIPhone:   "454,453,455,456,457,458,459,460,462,463,464,465,466,1473,1934,2225,2275",
		model.PlatIPad:     "788,787,789,790,791,792,793,794,795,796,797,798,799,1485,1932,2239,2289",
		model.PlatAndroid:  "617,616,618,619,620,621,622,623,624,625,626,627,628,1479,1933,2232,2282",
		model.PlatIPhoneI:  "1022,1017,1028,1033,1038,1043,1048,1053,1058,1063,1068,1073,1079,1491",
		model.PlatAndroidG: "1360,1355,1366,1371,1376,1381,1386,1391,1396,1401,1406,1411,1417,1497",
		model.PlatAndroidI: "1791,1785,1798,1804,1810,1816,1822,1828,1834,1840,1846,1852,1859,1873",
		model.PlatIPadI:    "1192,1187,1198,1203,1208,1213,1218,1223,1228,1233,1238,1243,1249,1503",
	}
	_bannersPGC = map[int8]map[int]int{
		model.PlatAndroid: {
			13:  83,
			167: 85,
			177: 232,
			11:  220,
			23:  49,
		},
		model.PlatIPhone: {
			13:  97,
			167: 98,
			177: 233,
			11:  221,
			23:  50,
		},
		model.PlatIPad: {
			13:  332,
			167: 333,
			177: 334,
			11:  336,
			23:  335,
		},
	}
)

// getBanners get banners by plat, build channel, ip.
func (s *Service) getBanners(c context.Context, plat int8, build, rid int, mid int64, channel, ip, buvid, network, mobiApp, device, adExtra string) (res map[string][]*banner.Banner) {
	var (
		resID = _banners[rid][plat]
		bs    []*banner.Banner
		isAd  bool
	)
	res = map[string][]*banner.Banner{}
	if bs, isAd = s.bgmBanners(c, plat, rid, build, resID, mid, mobiApp, device, buvid, network, ip, adExtra); len(bs) == 0 {
		bs = s.resBanners(c, plat, build, mid, resID, channel, ip, buvid, network, mobiApp, device, adExtra, isAd)
	}
	if len(bs) > 0 {
		res["top"] = bs
	}
	return
}

// resBannersplat
func (s *Service) resBanners(c context.Context, plat int8, build int, mid int64, resID int, channel, ip, buvid, network, mobiApp, device, adExtra string, isAd bool) (res []*banner.Banner) {
	var (
		plm   = s.bannerCache[plat] // operater banner
		err   error
		resbs map[int][]*resource.Banner
		tmp   []*resource.Banner
	)
	resIDStr := strconv.Itoa(resID)
	if resbs, err = s.res.ResBanner(c, plat, build, mid, resIDStr, channel, ip, buvid, network, mobiApp, device, adExtra, isAd); err != nil || len(resbs) == 0 {
		log.Error("s.res.ResBanner is null or err(%v)", err)
		resbs = plm
	}
	tmp = resbs[resID]
	for _, rb := range tmp {
		b := &banner.Banner{}
		b.ResChangeBanner(rb)
		res = append(res, b)
	}
	return
}

// bgmBanners bangumi banner
func (s *Service) bgmBanners(c context.Context, plat int8, rid, build, resID int, mid int64, mobiApp, device, buvid, network, ipaddr, adExtra string) (bgmBanner []*banner.Banner, isAd bool) {
	var (
		bgmb      = s.bannerBmgCache[plat][rid]
		cpmResBus map[int]map[int]*banner.Banner
		cpmBus    map[int]*banner.Banner
		allRank   []int // ad index
		ok        bool
	)
	// pgc banner
	for i, bb := range bgmb {
		b := &banner.Banner{}
		b.BgmChangeBanner(bb)
		b.RequestId = strconv.FormatInt(time.Now().UnixNano()/1000000, 10)
		b.Index = i + 1
		b.ResourceID = resID
		bgmBanner = append(bgmBanner, b)
	}
	isAd = true
	if len(bgmBanner) == 0 {
		return
	}
	if feature.GetBuildLimit(c, s.c.Feature.FeatureBuildLimit.BgmBanners, &feature.OriginResutl{
		BuildLimit: mobiApp == "iphone" && device == "pad" && build <= 8960,
	}) {
		isAd = false
		return
	}
	// ad banner
	cpmResBus = s.adBanners(c, mid, build, strconv.Itoa(resID), mobiApp, device, buvid, network, ipaddr, adExtra)
	if cpmBus, ok = cpmResBus[resID]; ok && len(cpmBus) > 0 {
		var (
			cpmMs = map[int]*banner.Banner{}
		)
		for _, cpm := range cpmBus {
			if cpm.IsAdReplace { // 是广告并且是广告位
				cpmMs[cpm.Rank] = cpm
				allRank = append(allRank, cpm.Rank)
				delete(cpmBus, cpm.Rank)
			}
		}
		if len(allRank) > 0 {
			sort.Ints(allRank)
		}
		for _, index := range allRank {
			if index == 0 {
				continue
			}
			if ad, ok := cpmMs[index]; ok {
				if len(bgmBanner) < index {
					bgmBanner = append(bgmBanner, ad)
					continue
				}
				bgmBanner = append(bgmBanner[:index-1], append([]*banner.Banner{ad}, bgmBanner[index-1:]...)...)
			}
		}
		for i, b := range bgmBanner {
			if ad, ok := cpmBus[i+1]; ok && !ad.IsAdReplace { // 是广告位但不是广告
				b.IsAdLoc = true
				b.IsAd = ad.IsAd
				b.CmMark = ad.CmMark
				b.SrcId = ad.SrcId
				b.RequestId = ad.RequestId
				b.ClientIp = ad.ClientIp
			}
			b.Index = i + 1
			b.ResourceID = resID
		}
	}
	return
}

func (s *Service) adBanners(c context.Context, mid int64, build int, resource, mobiApp, device, buvid, network, ipaddr, adExtra string) (banners map[int]map[int]*banner.Banner) {
	ipInfo, err := s.loc.Info(c, ipaddr)
	if err != nil || ipInfo == nil {
		log.Error("adBanners s.loc ip(%s) error(%v) or ipinfo is nil", ipaddr, err)
		ipInfo = &locgrpc.InfoReply{Addr: ipaddr}
	}
	adr, err := s.ad.ADRequest(c, mid, build, buvid, resource, ipaddr, ipInfo.Country, ipInfo.Province, ipInfo.City, network, mobiApp, device, adExtra)
	if err != nil {
		log.Error("%+v", err)
		return
	}
	banners = adr.ConvertBanner(ipaddr, mobiApp, build)
	return
}
