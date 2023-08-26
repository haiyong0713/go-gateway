package region

import (
	"context"
	"go-common/library/log"
)

// loadRegionListCache
func (s *Service) loadRegionListCache() error {
	res, err := s.dao.RegionPlat(context.Background())
	if err != nil {
		log.Error("s.dao.RegionPlat error(%+v)", err)
		return err
	}
	if err := s.dao.AddCacheRegionList(context.Background(), res); err != nil {
		log.Error("qdd cache region list error(%+v)", err)
		return err
	}
	return nil
}

// loadRegion regions cache.
func (s *Service) loadRegion() error {
	res, err := s.dao.All(context.Background())
	if err != nil {
		log.Error("s.dao.All error(%+v)", err)
		return err
	}
	if err = s.dao.AddCacheRegion(context.Background(), res); err != nil {
		log.Error("add cache region error(%+v)", err)
		return err
	}
	return nil
}

func (s *Service) loadRegionlist() error {
	res, err := s.dao.AllList(context.Background())
	if err != nil {
		log.Error("s.dao.AllList error(%+v)", err)
		return err
	}
	limit, err := s.dao.Limit(context.Background())
	if err != nil {
		log.Error("s.dao.limit error(%+v)", err)
		return err
	}
	config, err := s.dao.Config(context.Background())
	if err != nil {
		log.Error("s.dao.Config error(%+v)", err)
		return err
	}
	if err = s.dao.AddRegionList(context.Background(), res, limit, config); err != nil {
		log.Error("add region list error(%+v)", err)
		return err
	}
	return nil
}
