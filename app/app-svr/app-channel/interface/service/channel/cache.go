package channel

import (
	"context"
	"time"

	"go-common/library/log"
	"go-gateway/app/app-svr/app-card/interface/model/card/operate"
	"go-gateway/app/app-svr/app-channel/interface/model/card"
	"go-gateway/app/app-svr/app-channel/interface/model/tab"
)

// loadCardCache card cache
func (s *Service) loadCardCache() {
	log.Info("cronLog start loadCardCache")
	var (
		now     = time.Now()
		tmp     map[int64][]*card.Card
		tmpPlat map[string][]*card.CardPlat
		tmpUp   map[int64]*operate.Follow
		err     error
		c       = context.TODO()
	)
	if tmp, err = s.cd.Card(c, now); err != nil {
		log.Error("card s.cd.Card error(%v)", err)
		return
	}
	s.cardCache = tmp
	log.Info("loadCardCache success")
	if tmpPlat, err = s.cd.CardPlat(c); err != nil {
		log.Error("card s.cd.CardPlat error(%v)", err)
		return
	}
	s.cardPlatCache = tmpPlat
	log.Info("loadCardPlatCache success")
	if tmpUp, err = s.cd.UpCard(c); err != nil {
		log.Error("card s.cd.UpCard error(%v)", err)
		return
	}
	s.upCardCache = tmpUp
	log.Info("loadUpCardCache success")
}

func (s *Service) loadConvergeCache() {
	log.Info("cronLog start loadConvergeCache")
	var (
		tmp map[int64]*operate.Converge
		err error
		c   = context.TODO()
	)
	if tmp, err = s.ce.Cards(c); err != nil {
		log.Error("converge s.ce.Cards error(%v)", err)
		return
	}
	s.convergeCardCache = tmp
	log.Info("loadConvergeCache success")
}

func (s *Service) loadSpecialCache() {
	log.Info("cronLog start loadSpecialCache")
	var (
		tmp map[int64]*operate.Special
		err error
		c   = context.TODO()
	)
	if tmp, err = s.sl.Card(c); err != nil {
		log.Error("special s.sl.Card error(%v)", err)
		return
	}
	s.specialCardCache = tmp
	log.Info("loadSpecialCache success")
}

func (s *Service) loadLiveCardCache() {
	log.Info("cronLog start loadLiveCardCache")
	csm, err := s.lv.Card(context.TODO())
	if err != nil {
		log.Error("live s.lv.Card error(%v)", err)
		return
	}
	s.liveCardCache = csm
	log.Info("loadLiveCardCache success")
}

func (s *Service) loadGameDownloadCache() {
	log.Info("cronLog start loadGameDownloadCache")
	var (
		download map[int64]*operate.Download
		err      error
	)
	c := context.TODO()
	if download, err = s.g.DownLoad(c); err != nil {
		log.Error("%+v", err)
		return
	}
	s.gameDownloadCache = download
}

func (s *Service) loadCardSetCache() {
	log.Info("cronLog start loadCardSetCache")
	var (
		cards map[int64]*operate.CardSet
		err   error
	)
	if cards, err = s.cd.CardSet(context.TODO()); err != nil {
		log.Error("%+v", err)
		return
	}
	s.cardSetCache = cards
}

func (s *Service) loadMenusCache() {
	log.Info("cronLog start loadMenusCache")
	var (
		now   = time.Now()
		menus map[int64][]*tab.Menu
		err   error
	)
	if menus, err = s.tab.Menus(context.TODO(), now); err != nil {
		log.Error("%+v", err)
		return
	}
	s.menuCache = menus
}
