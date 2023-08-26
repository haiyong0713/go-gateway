package show

import (
	"context"
	"time"

	"go-common/library/log"
)

func (s *Service) loadPopEntrances() error {
	res, err := s.dao.Entrances(context.Background())
	if err != nil || len(res) == 0 {
		log.Error("Popular Entrance err (%+v)", err)
		return err
	}
	if err = s.dao.AddCacheEntrances(context.Background(), res); err != nil {
		log.Error("Failed to add cache entrances: res(%+v), err(%+v)", res, err)
		return err
	}
	return nil
}

func (s *Service) loadLargeCards() error {
	res, err := s.dao.LargeCards(context.Background())
	if err != nil || len(res) == 0 {
		log.Error("Popular LargeCards err (%+v)", err)
		return err
	}
	if err = s.dao.AddCacheLargeCards(context.Background(), res); err != nil {
		log.Error("Failed to add cache large cards: res(%+v), err(%+v)", res, err)
		return err
	}
	return nil
}

func (s *Service) loadLiveCards() error {
	res, err := s.dao.LiveCards(context.Background())
	if err != nil || len(res) == 0 {
		log.Error("Popular LiveCards err (%+v)", err)
		return err
	}
	if err = s.dao.AddCacheLiveCards(context.Background(), res); err != nil {
		log.Error("Failed to add cache live cards: res(%+v), err(%+v)", res, err)
		return err
	}
	return nil
}

// loadShowCache load all show cache
func (s *Service) loadShowCache() error {
	hdm, err := s.dao.Heads(context.Background())
	if err != nil {
		log.Error("s.dao.Heads error(%+v)", err)
		return err
	}
	itm, err := s.dao.Items(context.Background())
	if err != nil {
		log.Error("s.dao.Items error(%+v)", err)
		return err
	}
	if err = s.dao.AddCacheShow(context.Background(), hdm, itm); err != nil {
		log.Error("Failed to add cache show: headMap(%+v), itemMap(%+v), err(%+v)", hdm, itm, err)
		return err
	}
	return nil
}

// loadShowTempCache load all show temp cache
func (s *Service) loadShowTempCache() error {
	hdm, err := s.dao.TempHeads(context.Background())
	if err != nil {
		log.Error("s.dao.TempHeads error(%+v)", err)
		return err
	}
	itm, err := s.dao.TempItems(context.Background())
	if err != nil {
		log.Error("s.dao.TempItems error(%+v)", err)
		return err
	}
	if err = s.dao.AddTempCacheShow(context.Background(), hdm, itm); err != nil {
		log.Error("Failed to add temp cache show: headMap(%+v), itemMap(%+v), err(%+v)", hdm, itm, err)
		return err
	}
	return nil
}

func (s *Service) loadArticleCardsCache() error {
	cards, err := s.dao.ArticleCard(context.Background())
	if err != nil {
		log.Error("s.dao.ArticleCard error(%+v)", err)
		return err
	}
	if err = s.dao.AddCacheArticleCards(context.Background(), cards); err != nil {
		log.Error("Failed to add cache article card: cards(%+v), err(%+v)", cards, err)
		return err
	}
	return nil
}

func (s *Service) loadCardSetCache() error {
	cards, err := s.cdao.CardSet(context.Background())
	if err != nil {
		log.Error("s.cdao.CardSet(%+v)", err)
		return err
	}
	if err = s.cdao.AddCacheCardSet(context.Background(), cards); err != nil {
		log.Error("Failed to add cache card set: cards(%+v), err(%+v)", cards, err)
		return err
	}
	return nil
}

func (s *Service) loadEventTopicCache() error {
	eventTopic, err := s.cdao.EventTopic(context.Background())
	if err != nil {
		log.Error("s.cdao.eventTopic error(%+v)", err)
		return err
	}
	if err = s.cdao.AddCacheEventTopic(context.Background(), eventTopic); err != nil {
		log.Error("Failed to add cache event topic: eventTopic(%+v), err(%+v)", eventTopic, err)
		return err
	}
	return nil
}

func (s *Service) loadColumnListCache() error {
	columns, err := s.cdao.ColumnList(context.Background(), time.Now())
	if err != nil {
		log.Error("s.cdao.ColumnList error(%+v)", err)
		return err
	}
	if err = s.cdao.AddCacheColumnList(context.Background(), columns); err != nil {
		log.Error("Failed to add column list: columns(%+v), err(%+v)", columns, err)
		return err
	}
	return nil
}

// loadColumnsCache load all columns cache
func (s *Service) loadColumnsCache() error {
	res, err := s.cdao.Columns(context.Background())
	if err != nil {
		log.Error("s.cdao.Columns error(%+v)", err)
		return err
	}
	if err = s.cdao.AddCacheColumns(context.Background(), res); err != nil {
		log.Error("Failed to add cache columns: res(%+v), err(%+v)", res, err)
		return err
	}
	return nil
}

func (s *Service) loadNperCache() error {
	now := time.Now()
	hdm, err := s.cdao.ColumnNpers(context.Background(), now)
	if err != nil {
		log.Error("s.cdao.ColumnNpers error(%+v)", err)
		return err
	}
	itm, aids, err := s.cdao.NperContents(context.Background(), now)
	if err != nil {
		log.Error("s.cdao.NperContents error(%+v)", err)
		return err
	}
	if err = s.cdao.AddCacheNper(context.Background(), hdm, itm, aids); err != nil {
		log.Error("Failed to add cache Nper: headMap(%+v), itemMap(%+v), err(%+v)", hdm, itm, err)
		return err
	}
	return nil
}

// loadCardCache load all card cache
func (s *Service) loadCardCache() error {
	now := time.Now()
	hdm, err := s.cdao.PosRecs(context.Background(), now)
	if err != nil {
		log.Error("s.cdao.PosRecs error(%+v)", err)
		return err
	}
	itm, aids, err := s.cdao.RecContents(context.Background(), now)
	if err != nil {
		log.Error("s.cdao.RecContents error(%+v)", err)
		return err
	}
	if err = s.cdao.AddCacheCard(context.Background(), hdm, itm, aids); err != nil {
		log.Error("Failed to add cache card: headMap(%+v), itemMap(%+v), err(%+v)", hdm, itm, err)
		return err
	}
	return nil
}

func (s *Service) loadAuditCache() error {
	as, err := s.adao.Audits(context.Background())
	if err != nil {
		log.Error("s.adt.Audits error(%+v)", err)
		return err
	}
	if err = s.adao.AddCacheAudit(context.Background(), as); err != nil {
		log.Error("Failed to add cache audit: audit(%+v), err(%+v)", as, err)
		return err
	}
	return nil
}
