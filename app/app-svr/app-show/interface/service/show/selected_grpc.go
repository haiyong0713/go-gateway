package show

import (
	"context"

	"go-common/library/log"

	xecode "go-gateway/app/app-svr/app-card/ecode"
	"go-gateway/app/app-svr/app-show/interface/model/selected"
)

// BackToSrcSerie def.
// nolint:gomnd
func (s *Service) BackToSrcSerie(c context.Context, sType string, number int64) (result *selected.SerieFull, err error) {
	var (
		config *selected.SerieConfig
		list   []*selected.SelectedRes
	)
	if config, err = s.cdao.SerieConfig(c, sType, number); err != nil {
		log.Error("SerieConfig sType %s, Number %d, Err %v", sType, number, err)
		return
	}
	if config.MediaID > 0 {
		config.MediaID = config.MediaID*100 + s.c.ShowSelectedCfg.MediaMID%100
	}
	if list, err = s.cdao.SelectedRes(c, config.ID); err != nil {
		log.Error("SerieConfig sType %s, Number %d, Err %v", sType, number, err)
		return
	}
	if len(list) == 0 {
		err = xecode.AppNotData
		return
	}
	config.Init()
	result = &selected.SerieFull{
		Config: config,
		List:   list,
	}
	_ = s.fanout.Do(c, func(c context.Context) {
		_ = s.cdao.AddSerieCache(c, result)
		log.Info("AddSerieCache Updates Cache Stype %s, Number %d", sType, number)
	})
	return
}

// BackToSrcSeries def.
func (s *Service) BackToSrcSeries(c context.Context, sType string) (result []*selected.SerieFilter, err error) {
	if result, err = s.cdao.Series(c, sType); err != nil { // back to source
		log.Error("All_Series Stype %s, DB Err %v", sType, err)
		return
	}
	for _, v := range result { // build name and disaster recovery data
		v.Init()
	}
	log.Warn("s.cdao.Series result(%+v)", result)
	if len(result) > 0 {
		_ = s.fanout.Do(c, func(c context.Context) {
			_ = s.cdao.SetAllSeries(c, sType, result)
			log.Info("All_Series Updates Cache Stype %s", sType)
		})
	}
	return
}
