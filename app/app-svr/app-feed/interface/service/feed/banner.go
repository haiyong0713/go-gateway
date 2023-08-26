package feed

import (
	"context"
	"strconv"

	"go-common/library/net/metadata"
	"go-gateway/app/app-svr/app-card/interface/model/card/ai"
	"go-gateway/app/app-svr/app-card/interface/model/card/banner"
	"go-gateway/app/app-svr/app-feed/interface/common"
	"go-gateway/app/app-svr/app-feed/interface/model/feed"
	"go-gateway/app/app-svr/app-feed/interface/model/sets"
	resourcegrpc "go-gateway/app/app-svr/resource/service/api/v1"
	resource "go-gateway/app/app-svr/resource/service/model"
)

var (
	inlineBannersSet            = sets.NewInt64(4332, 4336, 4340, 4592, 4948, 4953, 4960, 4967)
	inlineBannersWithoutIPadSet = sets.NewInt64(4332, 4336, 4592, 4948, 4960, 4953)
)

// banners get banners by plat, build channel, ip.
func (s *Service) banners(c context.Context, plat int8, build int, mid int64, buvid, network, mobiApp, device, openEvent, adExtra, hash string, splashID int64, _ *feed.Abtest, lessonsMode int, bannerInfoItem []*ai.BannerInfoItem, teenagerMode int) (bs []*banner.Banner, version string, err error) {
	var (
		rscID = common.OldBannerResource(c, plat)
		bm    map[int][]*resource.Banner
		isAd  = true
	)
	// abtest banner
	isNewBannerResource := false
	if s.canEnable169BannerResourceID(mobiApp, build) {
		tmpID := common.InlineBannerResource(c, plat)
		if tmpID > 0 {
			isNewBannerResource = true
			rscID = tmpID
		}
	}
	if lessonsMode == 1 {
		rscID = common.BannerLessonResource(c, plat)
		if rscID == 0 {
			return
		}
	}
	if teenagerMode == 1 {
		rscID = common.BannerTeenagerResource(c, plat)
		if rscID == 0 {
			return
		}
	}
	if isNewBannerResource && len(bannerInfoItem) > 0 {
		reply, err := s.rsc.FeedBanners(c, &resourcegrpc.FeedBannersRequest{
			Meta:      asProtoMeta(bannerInfoItem),
			Plat:      int64(plat),
			Build:     int64(build),
			Mid:       mid,
			ResId:     rscID,
			Channel:   "",
			Ip:        metadata.String(c, metadata.RemoteIP),
			Buvid:     buvid,
			Network:   network,
			MobiApp:   mobiApp,
			Device:    device,
			IsAd:      isAd,
			OpenEvent: openEvent,
			AdExtra:   adExtra,
			Version:   hash,
			SplashId:  splashID,
		})
		if err != nil {
			return nil, "", err
		}
		for _, i := range reply.Banner {
			b := &banner.Banner{}
			b.FromProto(i)
			bs = append(bs, b)
		}
		return bs, reply.Version, nil
	}
	// abtest banner
	if bm, version, err = s.rsc.Banner(c, plat, build, mid, strconv.FormatInt(rscID, 10), "", buvid, network, mobiApp, device, isAd, openEvent, adExtra, hash, splashID); err != nil {
		return
	}
	for _, rb := range bm[int(rscID)] {
		b := &banner.Banner{}
		b.Change(rb)
		bs = append(bs, b)
	}
	return
}

func asProtoMeta(in []*ai.BannerInfoItem) []*resourcegrpc.BannerMeta {
	out := make([]*resourcegrpc.BannerMeta, 0, len(in))
	for _, i := range in {
		out = append(out, &resourcegrpc.BannerMeta{
			Id:         i.ID,
			Type:       i.Type,
			InlineType: i.InlineType,
			InlineId:   i.InlineID,
		})
	}
	return out
}
