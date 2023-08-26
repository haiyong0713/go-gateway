package display

import (
	"context"
	"go-common/library/net/metadata"
	"strconv"
	"time"

	"go-common/library/log"
	"go-gateway/app/app-svr/app-interface/interface-legacy/model/display"

	locgrpc "git.bilibili.co/bapis/bapis-go/community/service/location"
)

// DisplayID is display id .
func (s *Service) DisplayID(c context.Context, mid int64, buvid string, now time.Time) (id string) {
	if mid == 0 {
		id = buvid + "-" + strconv.FormatInt(now.Unix(), 10)
	} else {
		id = strconv.FormatInt(mid, 10) + "-" + strconv.FormatInt(now.Unix(), 10)
	}
	return
}

// Zone is zone id and district info .
func (s *Service) Zone(c context.Context, now time.Time) (zone *display.Zone) {
	var (
		info *locgrpc.InfoReply
		err  error
	)
	zone = &display.Zone{}
	if info, err = s.loc.Info(c, metadata.String(c, metadata.RemoteIP)); err != nil || info == nil {
		log.Error("error %v or info is nil", err)
		return
	}
	zone.ID = info.ZoneId
	zone.Addr = info.Addr
	zone.ISP = info.Isp
	zone.Country = info.Country
	zone.Province = info.Province
	zone.City = info.City
	zone.Latitude = info.Latitude
	zone.Longitude = info.Longitude
	zone.CountryCode = int(info.CountryCode)
	return
}
