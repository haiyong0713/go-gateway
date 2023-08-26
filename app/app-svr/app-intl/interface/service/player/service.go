package player

import (
	"context"
	"strings"

	"go-common/library/log"
	"go-gateway/app/app-svr/app-intl/interface/conf"
	playurldao "go-gateway/app/app-svr/app-intl/interface/dao/playurl"
	"go-gateway/app/app-svr/app-intl/interface/model"
	"go-gateway/app/app-svr/app-intl/interface/model/player"
	playurlV2Api "go-gateway/app/app-svr/playurl/service/api/v2"
)

var (
	vipQn = []uint32{116, 112, 74}
)

// Service is space service
type Service struct {
	c          *conf.Config
	playURLDao *playurldao.Dao
	//vip qn
	vipQnMap map[uint32]struct{}
}

// New new space
func New(c *conf.Config) (s *Service) {
	s = &Service{
		c:          c,
		vipQnMap:   make(map[uint32]struct{}),
		playURLDao: playurldao.New(c),
	}
	// type cache
	for _, pn := range vipQn {
		s.vipQnMap[pn] = struct{}{}
	}
	return
}

// PlayURLV2 is
func (s *Service) PlayURLV2(c context.Context, mid int64, params *player.Param, plat int8) (playurl *player.PlayurlV2Reply, err error) {
	var reply *playurlV2Api.PlayURLReply
	if reply, err = s.playURLDao.PlayURLV2(c, params, mid); err != nil {
		log.Error("d.playURLDao.PlayURLV2 error(%+v)", err)
		return
	}
	if reply == nil || reply.Playurl == nil {
		return
	}
	playurl = new(player.PlayurlV2Reply)
	playurl.FormatPlayURL(reply.Playurl)
	// 版本过滤，老版本不支持大会员和付费
	if plat == model.PlatAndroidI && params.Build < 2002000 {
		qualitys := make([]uint32, 0, len(playurl.AcceptQuality))
		descs := make([]string, 0, len(playurl.AcceptDescription))
		formats := make([]string, 0, len(playurl.AcceptQuality))
		acceptFormats := strings.Split(playurl.AcceptFormat, ",")
		for index, quality := range playurl.AcceptQuality {
			if _, ok := s.vipQnMap[quality]; !ok {
				qualitys = append(qualitys, quality)
				if index < len(playurl.AcceptDescription) {
					descs = append(descs, playurl.AcceptDescription[index])
				}
				if index < len(acceptFormats) {
					formats = append(formats, acceptFormats[index])
				}
			}
		}
		playurl.AcceptQuality = qualitys
		playurl.AcceptDescription = descs
		playurl.AcceptFormat = strings.Join(formats, ",")
	}
	return
}
