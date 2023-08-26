package show

import (
	"context"
	"strconv"

	"go-common/library/log"
	"go-gateway/app/app-svr/app-show/interface/model"
	"go-gateway/app/app-svr/app-show/interface/model/show"

	locmdl "git.bilibili.co/bapis/bapis-go/community/service/location"
)

// cpmRecommend
func (s *Service) cpmRecommend(c context.Context, mid int64, build int, buvid, resource, network, mobiApp, device, ipaddr string) (sis map[int]*show.Item) {
	if _, ok := s.cpmRcmmndMid[mid]; !ok && (mid == 0 || int(mid)%100 >= s.cpmRcmmndNum) && !s.cpmRcmmndAll {
		return
	}
	var (
		info                    *locmdl.InfoReply
		country, province, city string
		err                     error
	)
	if info, err = s.loc.Info(c, ipaddr); err != nil {
		log.Warn("s.loc.Info(%v) error(%v)", ipaddr, err)
	}
	if info != nil {
		country = info.Country
		province = info.Province
		city = info.Province
	}
	adr, err := s.ad.ADRequest(c, mid, build, buvid, resource, ipaddr, country, province, city, network, mobiApp, device, "")
	if err != nil {
		log.Error("s.ad.ADRequest error(%v)", err)
		return
	}
	sis = map[int]*show.Item{}
	sAdis := adr.ADIndexs[resource]
	if len(sAdis) == 0 {
		log.Info("mobi_app:%v-build:%v-resource:%v-is_ad_loc:%v", mobiApp, build, resource, false)
		return
	}
	var aids []int64
	for sidStr, adi := range sAdis {
		sid, _ := strconv.Atoi(sidStr)
		var si = &show.Item{
			IsAdLoc:     true,
			IsAd:        adi.IsAd,
			IsAdReplace: false,
			CmMark:      adi.CmMark,
			Rank:        adi.Index,
			SrcId:       sid,
			RequestId:   adr.RequestID,
			ClientIp:    ipaddr,
		}
		if adInfo := adi.Info; adInfo != nil {
			aids = append(aids, adInfo.CreativeContent.VideoID)
			// si
			si.Goto = model.GotoAv
			si.Param = strconv.FormatInt(adInfo.CreativeContent.VideoID, 10)
			si.URI = model.FillURI(model.GotoAv, si.Param, nil)
			si.Title = adInfo.CreativeContent.Title
			si.Cover = model.CoverURL(adInfo.CreativeContent.ImageURL)
			si.IsAdReplace = true
			si.AdCb = adInfo.AdCb
			si.CreativeId = adInfo.CreativeID
			si.ShowUrl = adInfo.CreativeContent.ShowURL
			si.ClickUrl = adInfo.CreativeContent.ClickURL
		}
		sis[si.Rank] = si
		log.Info("mobi_app:%v-build:%v-resource:%v-is_ad_loc:%v", mobiApp, build, resource, true)
	}
	if len(aids) == 0 {
		return
	}
	as, err := s.arc.ArchivesPB(c, aids, mid, mobiApp, device)
	if err != nil {
		log.Error("s.arc.ArchivesPB(%v) error(%v)", aids, err)
		return
	}
	for _, si := range sis {
		aid, _ := strconv.ParseInt(si.Param, 10, 64)
		if a, ok := as[aid]; ok && a != nil {
			si.Play = int(a.Stat.View)
			si.Danmaku = int(a.Stat.Danmaku)
			if si.Title == "" {
				si.Title = a.Title
			}
			if si.Cover == "" {
				si.Cover = model.CoverURL(a.Pic)
			}
			si.URI = model.AvHandler(a)(si.URI)
		}
	}
	return
}
