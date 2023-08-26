package favorite

import (
	"go-gateway/app/app-svr/app-interface/interface-legacy/conf"
	artdao "go-gateway/app/app-svr/app-interface/interface-legacy/dao/article"
	audiodao "go-gateway/app/app-svr/app-interface/interface-legacy/dao/audio"
	bangumidao "go-gateway/app/app-svr/app-interface/interface-legacy/dao/bangumi"
	bplusdao "go-gateway/app/app-svr/app-interface/interface-legacy/dao/bplus"
	"go-gateway/app/app-svr/app-interface/interface-legacy/dao/channel"
	"go-gateway/app/app-svr/app-interface/interface-legacy/dao/checkin"
	cheesedao "go-gateway/app/app-svr/app-interface/interface-legacy/dao/cheese"
	comicdao "go-gateway/app/app-svr/app-interface/interface-legacy/dao/comic"
	favdao "go-gateway/app/app-svr/app-interface/interface-legacy/dao/favorite"
	malldao "go-gateway/app/app-svr/app-interface/interface-legacy/dao/mall"
	"go-gateway/app/app-svr/app-interface/interface-legacy/dao/note"
	spdao "go-gateway/app/app-svr/app-interface/interface-legacy/dao/sp"
	ticketdao "go-gateway/app/app-svr/app-interface/interface-legacy/dao/ticket"
	topicdao "go-gateway/app/app-svr/app-interface/interface-legacy/dao/topic"
	"go-gateway/app/app-svr/app-interface/interface-legacy/dao/workshop"
)

// Service is favorite.
type Service struct {
	c *conf.Config
	// dao
	favDao      *favdao.Dao
	artDao      *artdao.Dao
	spDao       *spdao.Dao
	topicDao    *topicdao.Dao
	bplusDao    *bplusdao.Dao
	audioDao    *audiodao.Dao
	bangumiDao  *bangumidao.Dao
	ticketDao   *ticketdao.Dao
	mallDao     *malldao.Dao
	comicDao    *comicdao.Dao
	cheeseDao   *cheesedao.Dao
	noteDao     *note.Dao
	channelDao  *channel.Dao
	workshopDao *workshop.Dao
	checkinDao  *checkin.Dao
}

// New new favorite。
func New(c *conf.Config) (s *Service) {
	s = &Service{
		c: c,
		// dao
		favDao:      favdao.New(c),
		topicDao:    topicdao.New(c),
		artDao:      artdao.New(c),
		spDao:       spdao.New(c),
		bplusDao:    bplusdao.New(c),
		audioDao:    audiodao.New(c),
		bangumiDao:  bangumidao.New(c),
		ticketDao:   ticketdao.New(c),
		mallDao:     malldao.New(c),
		comicDao:    comicdao.New(c),
		cheeseDao:   cheesedao.New(c),
		noteDao:     note.New(c),
		channelDao:  channel.New(c),
		workshopDao: workshop.New(c),
		checkinDao:  checkin.New(c),
	}
	return s
}
