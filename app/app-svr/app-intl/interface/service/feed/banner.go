package feed

import (
	"context"
	"strconv"

	"go-gateway/app/app-svr/app-card/interface/model/card/banner"
	"go-gateway/app/app-svr/app-feed/interface/model"
	resource "go-gateway/app/app-svr/resource/service/model"
)

var (
	_banners = map[int8]int{
		model.PlatIPhoneB:  467,
		model.PlatIPhone:   467,
		model.PlatAndroid:  631,
		model.PlatAndroidB: 631,
		model.PlatIPad:     771,
		model.PlatIPhoneI:  947,
		model.PlatAndroidG: 1285,
		model.PlatAndroidI: 1707,
		model.PlatIPadI:    1117,
	}
	_bigBanners = map[int8]int{
		model.PlatIPhoneI:  4114,
		model.PlatAndroidI: 4119,
	}
)

// banners get banners by plat, build channel, ip.
func (s *Service) banners(c context.Context, plat int8, build int, mid int64, buvid, network, mobiApp, device, openEvent, adExtra, hash string) (bs []*banner.Banner, version string, err error) {
	var (
		rscID = _banners[plat]
		bm    map[int][]*resource.Banner
	)
	if mobiApp == "iphone_i" || mobiApp == "android_i" && build > 2042030 {
		if tmpID, ok := _bigBanners[plat]; ok {
			rscID = tmpID
		}
	}
	if bm, version, err = s.rsc.Banner(c, plat, build, mid, strconv.Itoa(rscID), "", buvid, network, mobiApp, device, true, openEvent, adExtra, hash); err != nil {
		return
	}
	for _, rb := range bm[rscID] {
		b := &banner.Banner{}
		b.Change(rb)
		bs = append(bs, b)
	}
	return
}
