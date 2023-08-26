package dao

import (
	"context"
	"math"
	"sync"

	"go-common/library/ecode"
	"go-common/library/log"
	"go-common/library/time"
	"go-gateway/app/app-svr/ugc-season/service/api"

	"go-common/library/sync/errgroup.v2"
)

// UpperSeason is
func (d *Dao) UpperSeason(c context.Context, req *api.UpperListRequest) (seasons []*api.Season, totalCount int64, totalPage int64, err error) {
	var (
		sids  []int64
		start = (req.PageNum - 1) * req.PageSize
		end   = start + req.PageSize - 1
	)
	if sids, totalCount, err = d.UpperSeasonCache(c, req.Mid, start, end); err != nil {
		log.Error("d.UpperSeasonCache mid(%d) start(%d) end(%d) error(%+v)", req.Mid, start, end, err)
		err = nil
	} else if totalCount == 0 { //up从未投过剧集
		err = ecode.NothingFound
		return
	}
	if totalCount <= 0 {
		//up有过剧集可能是缓存过期不存在需要回源
		var (
			allSids []int64
			ptimes  []time.Time
		)
		if allSids, ptimes, err = d.UpperSeasonInfo(c, req.Mid); err != nil {
			log.Error("d.UpperSeasonInfo(%d) error(%+v)", req.Mid, err)
			return
		}
		if len(allSids) == 0 {
			err = ecode.NothingFound
			_ = d.SetUpperNoSeasonCache(c, req.Mid)
			return
		}
		d.addCache(func() {
			_ = d.AddUpperSeasonCache(c, req.Mid, allSids, ptimes)
		})
		if len(allSids) > int(end+1) {
			sids = allSids[start : end+1]
		} else {
			sids = allSids[start:]
		}
		totalCount = int64(len(allSids))
	}
	totalPage = int64(math.Ceil(float64(totalCount) / float64(req.PageSize)))
	sm, err := d.Seasons(c, sids)
	for _, sid := range sids {
		if s, ok := sm[sid]; ok {
			seasons = append(seasons, s)
		}
	}
	return
}

// Season is
func (d *Dao) Season(c context.Context, sid int64) (season *api.Season, err error) {
	addCache := true
	season, err = d.SeasonRdsCache(c, sid)
	if err != nil {
		log.Error("d.SeasonCache(%d) error(%+v)", sid, err)
		addCache = false
		err = nil
	}
	if season != nil {
		if st, _ := d.Stat(c, sid); st != nil {
			season.Stat = *st
		}
		return
	}
	season, err = d.SeasonInfo(c, sid)
	if err != nil {
		return
	}
	if season == nil {
		err = ecode.NothingFound
		return
	}
	if st, _ := d.Stat(c, sid); st != nil {
		season.Stat = *st
	}
	miss := season
	if !addCache {
		return
	}
	d.addCache(func() {
		_ = d.AddSeasonCache(context.Background(), miss)
	})
	return
}

// Seasons is
func (d *Dao) Seasons(c context.Context, sids []int64) (seasons map[int64]*api.Season, err error) {
	if len(sids) == 0 {
		return
	}
	addCache := true
	if seasons, err = d.SeasonsRdsCache(c, sids); err != nil {
		log.Error("d.SeasonsCache(%+v) error(%+v)", sids, err)
		addCache = false
		seasons = nil
		err = nil
	}
	var miss []int64
	for _, key := range sids {
		if (seasons == nil) || (seasons[key] == nil) {
			miss = append(miss, key)
		}
	}
	stm, _ := d.Stats(c, sids)
	if len(stm) > 0 {
		for sid, s := range seasons {
			if st, ok := stm[sid]; ok {
				s.Stat = *st
			}
		}
	}
	missLen := len(miss)
	if missLen == 0 {
		return
	}
	missData := make(map[int64]*api.Season, missLen)
	var mutex sync.Mutex
	eg := errgroup.WithCancel(c)
	var run = func(ms []int64) {
		eg.Go(func(ctx context.Context) (err error) {
			data, err := d.SeasonsInfo(ctx, ms)
			mutex.Lock()
			for k, v := range data {
				missData[k] = v
			}
			mutex.Unlock()
			return
		})
	}
	var (
		i int
		n = missLen / 50
	)
	for i = 0; i < n; i++ {
		run(miss[i*50 : (i+1)*50])
	}
	if len(miss[i*50:]) > 0 {
		run(miss[i*50:])
	}
	err = eg.Wait()
	if seasons == nil {
		seasons = make(map[int64]*api.Season, len(sids))
	}
	for k, v := range missData {
		seasons[k] = v
	}
	if err != nil {
		return
	}
	if len(stm) > 0 {
		for sid, s := range seasons {
			if st, ok := stm[sid]; ok {
				s.Stat = *st
			}
		}
	}
	if !addCache {
		return
	}
	d.addCache(func() {
		_ = d.AddSeasonsCache(context.Background(), seasons)
	})
	return
}

// View get Season by sid.
func (d *Dao) ViewWithStat(c context.Context, sid int64) (*api.View, error) {
	var (
		stat *api.Stat
		view *api.View
	)
	eg := errgroup.WithContext(c)
	eg.Go(func(ctx context.Context) (err error) {
		if view, err = d.View(ctx, sid); err != nil {
			log.Error("d.View(%d) err(%+v)", sid, err)
			return err
		}
		return nil
	})
	eg.Go(func(ctx context.Context) (err error) {
		if stat, err = d.Stat(ctx, sid); err != nil {
			log.Error("d.Stat(%d) err(%+v)", sid, err)
		}
		return nil
	})
	if err := eg.Wait(); err != nil {
		log.Error("eg.wait() sid(%d) err(%+v)", sid, err)
		return nil, err
	}
	if view == nil {
		return nil, ecode.NothingFound
	}
	if stat != nil && view.Season != nil {
		view.Season.Stat = *stat
	}
	return view, nil
}

func (d *Dao) View(c context.Context, sid int64) (*api.View, error) {
	var (
		season   *api.Season
		sections []*api.Section
		episodes map[int64][]*api.Episode
		cached   = true
	)
	view, err := d.ViewRdsCache(c, sid)
	if view != nil {
		return view, nil
	}
	if err != nil {
		log.Error("d.ViewCache(%d) error(%+v)", sid, err)
		cached = false
	}
	eg := errgroup.WithCancel(c)
	eg.Go(func(ctx context.Context) (err error) {
		if season, err = d.SeasonInfo(ctx, sid); err != nil {
			log.Error("d.SeasonInfo(%d) error(%+v)", sid, err)
		}
		return
	})
	eg.Go(func(ctx context.Context) (err error) {
		if sections, err = d.SectionsInfo(ctx, sid); err != nil {
			log.Error("d.SectionsInfo(%d) error(%+v)", sid, err)
		}
		return
	})
	eg.Go(func(ctx context.Context) (err error) {
		if episodes, err = d.EpisodesInfo(ctx, sid); err != nil {
			log.Error("d.EpisodesInfo(%d) error(%+v)", sid, err)
		}
		return
	})
	if err := eg.Wait(); err != nil {
		log.Error("eg.Wait(%d) err(%+v)", sid, err)
		return nil, err
	}
	if season == nil || len(sections) == 0 || len(episodes) == 0 {
		log.Warn("nothingFound sid(%d) sections(%d) episodes(%d)", sid, len(sections), len(episodes))
		return nil, ecode.NothingFound
	}
	view = d.fillView(season, sections, episodes)
	if cached {
		viewBs, err := view.Marshal()
		if err == nil {
			d.addCache(func() {
				_ = d.AddViewCache(context.Background(), sid, viewBs)
			})
		}
	}
	return view, nil
}

// Stat get season stat.
func (d *Dao) Stat(c context.Context, sid int64) (st *api.Stat, err error) {
	var cached = true
	if st, err = d.StCache(c, sid); err != nil {
		log.Error("d.StatCache(%d) error(%+v)", sid, err)
		cached = false
	}
	if st != nil {
		return
	}
	if st, err = d.StatInfo(c, sid); err != nil {
		log.Error("d.StatInfo(%d) error(%+v)", sid, err)
		return
	}
	if st == nil {
		st = &api.Stat{SeasonID: sid}
		return
	}
	if cached {
		d.addCache(func() {
			_ = d.AddStCache(context.Background(), st)
		})
	}
	return
}

// Stats is
func (d *Dao) Stats(c context.Context, sids []int64) (stats map[int64]*api.Stat, err error) {
	if len(sids) == 0 {
		return
	}
	addCache := true
	if stats, err = d.StsCache(c, sids); err != nil {
		log.Error("d.StsCache(%+v) error(%+v)", sids, err)
		addCache = false
		stats = nil
		err = nil
	}
	var miss []int64
	for _, key := range sids {
		if (stats == nil) || (stats[key] == nil) {
			miss = append(miss, key)
		}
	}
	missLen := len(miss)
	if missLen == 0 {
		return
	}
	missData := make(map[int64]*api.Stat, missLen)
	var mutex sync.Mutex
	eg := errgroup.WithCancel(c)
	var run = func(ms []int64) {
		eg.Go(func(ctx context.Context) (err error) {
			data, err := d.StatsInfo(ctx, ms)
			mutex.Lock()
			for k, v := range data {
				missData[k] = v
			}
			mutex.Unlock()
			return
		})
	}
	var (
		i int
		n = missLen / 50
	)
	for i = 0; i < n; i++ {
		run(miss[i*50 : (i+1)*50])
	}
	if len(miss[i*50:]) > 0 {
		run(miss[i*50:])
	}
	err = eg.Wait()
	if stats == nil {
		stats = make(map[int64]*api.Stat, len(sids))
	}
	for k, v := range missData {
		stats[k] = v
	}
	if err != nil {
		return
	}
	if !addCache {
		return
	}
	d.addCache(func() {
		_ = d.AddStsCache(context.Background(), stats)
	})
	return
}

// UpCache update season & view cache
func (d *Dao) UpCache(c context.Context, sid int64) (err error) {
	var (
		view      *api.View
		season    *api.Season
		seasonSec []*api.Section
		seasonEp  map[int64][]*api.Episode
	)
	if season, err = d.SeasonInfo(c, sid); err != nil || season == nil {
		log.Error("UpSeasonCache season err(%v) or season=nil", err)
		return
	}
	if seasonSec, err = d.SectionsInfo(c, sid); err != nil || len(seasonSec) == 0 {
		log.Error("UpSeasonCache seasonSec err(%v) or seasonSec=nil", err)
		return
	}
	if seasonEp, err = d.EpisodesInfo(c, sid); err != nil || len(seasonEp) == 0 {
		log.Error("UpSeasonCache seasonEp err(%v) or seasonEp=nil", err)
		return
	}
	view = d.fillView(season, seasonSec, seasonEp)
	viewBs, err := view.Marshal()
	if err != nil {
		log.Error("view.Marshal error(%+v)", err)
		return err
	}
	if err = d.AddViewCache(c, sid, viewBs); err != nil {
		log.Error("AddViewCache sid(%d) err(%v)", sid, err)
		return
	}
	if err = d.AddSeasonCache(c, season); err != nil {
		log.Error("AddSeasonCache sid(%d) err(%v)", sid, err)
		return
	}
	log.Warn("UpCache success sid(%d) view(%+v) season(%+v)", sid, view, season)
	return
}

func (d *Dao) fillView(season *api.Season, ss []*api.Section, se map[int64][]*api.Episode) (view *api.View) {
	var tmpSecs []*api.Section
	for _, sec := range ss {
		if _, ok := se[sec.ID]; !ok || len(se[sec.ID]) == 0 {
			continue
		}
		tmpSecs = append(tmpSecs, &api.Section{
			SeasonID: sec.SeasonID,
			ID:       sec.ID,
			Title:    sec.Title,
			Type:     sec.Type,
			Episodes: se[sec.ID],
		})
	}
	view = new(api.View)
	view.Season = season
	view.Sections = tmpSecs
	return
}

// DelCache delete season & view cache
func (d *Dao) DelCache(c context.Context, sid int64) (err error) {
	if err = d.DelViewCache(c, sid); err != nil {
		log.Error("DelViewCache sid(%d) err(%v)", sid, err)
		return
	}
	if err = d.DelSeasonCache(c, sid); err != nil {
		log.Error("DelSeasonCache sid(%d) err(%v)", sid, err)
		return
	}
	log.Warn("DelCache success sid(%d)", sid)
	return
}

// ViewsWithStat get Season by sids.
func (d *Dao) ViewsWithStat(c context.Context, sids []int64) (map[int64]*api.View, error) {
	var (
		stats = make(map[int64]*api.Stat)
		views = make(map[int64]*api.View)
	)
	eg := errgroup.WithContext(c)
	eg.Go(func(ctx context.Context) (err error) {
		if views, err = d.Views(ctx, sids); err != nil {
			log.Error("d.ViewsRdsCache err(%+v) sids(%+v)", err, sids)
		}
		return nil
	})
	eg.Go(func(ctx context.Context) (err error) {
		if stats, err = d.Stats(ctx, sids); err != nil {
			log.Error("d.Stats err(%+v) sids(%+v)", err, sids)
		}
		return nil
	})
	if err := eg.Wait(); err != nil {
		log.Error("eg.Wait(%+v) err(%+v)", sids, err)
	}
	for sid, view := range views {
		if s, ok := stats[sid]; ok {
			view.Season.Stat = *s
		}
	}
	return views, nil
}

func (d *Dao) Views(c context.Context, sids []int64) (map[int64]*api.View, error) {
	cached := true
	views, err := d.ViewsRdsCache(c, sids)
	if err != nil {
		log.Error("d.ViewsRdsCache err(%+v) sids(%+v)", err, sids)
		cached = false
	}
	var (
		missView []int64
		seasons  map[int64]*api.Season
		sections map[int64][]*api.Section //season_id为key
		episodes map[int64][]*api.Episode //section_id为key
	)
	for _, sid := range sids {
		if _, ok := views[sid]; !ok {
			missView = append(missView, sid)
		}
	}
	if len(missView) == 0 {
		log.Warn("sseason all cache sids(%+v)", sids)
		return views, nil
	}
	eg := errgroup.WithCancel(c)
	eg.Go(func(ctx context.Context) (err error) {
		if seasons, err = d.SeasonsInfo(ctx, missView); err != nil {
			log.Error("d.SeasonsInfo(%+v) error(%+v)", missView, err)
		}
		return err
	})
	eg.Go(func(ctx context.Context) (err error) {
		if sections, err = d.SectionsInfos(ctx, missView); err != nil {
			log.Error("d.SectionsInfos(%+v) error(%+v)", missView, err)
		}
		return err
	})
	eg.Go(func(ctx context.Context) (err error) {
		if episodes, err = d.EpisodesInfos(ctx, missView); err != nil {
			log.Error("d.EpisodesInfo(%+v) error(%+v)", missView, err)
		}
		return err
	})
	if err := eg.Wait(); err != nil {
		log.Error("eg.wait(%+v) err(%+v)", missView, err)
		return views, nil
	}
	if len(seasons) == 0 || len(sections) == 0 || len(episodes) == 0 {
		log.Error("nothingFound sids(%+v) seasonslen(%d) sectionslen(%d) episodeslen(%d)", missView, len(seasons), len(sections), len(episodes))
		return views, nil
	}
	var cachedViews []*api.View
	for _, s := range seasons {
		sec, secOk := sections[s.ID]
		if !secOk {
			continue
		}
		v := d.fillView(s, sec, episodes)
		views[s.ID] = v
		cacheView := &api.View{}
		*cacheView = *v
		cachedViews = append(cachedViews, cacheView)
	}
	log.Warn("sseason get missed(%+v)", cachedViews)
	if cached {
		d.addCache(func() {
			_ = d.AddViewCaches(context.Background(), cachedViews)
		})
	}
	return views, nil
}
