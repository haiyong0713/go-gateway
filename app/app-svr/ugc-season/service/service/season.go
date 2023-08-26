package service

import (
	"context"
	"strconv"
	"sync"

	"go-common/library/log"
	"go-common/library/sync/errgroup.v2"

	arcapi "go-gateway/app/app-svr/archive/service/api"
	"go-gateway/app/app-svr/ugc-season/service/api"
)

const _maxAids = 50

// Stat is
func (s *Service) Stat(c context.Context, sid int64) (stat *api.Stat, err error) {
	if stat, err = s.d.Stat(c, sid); err != nil {
		log.Error("s.d.Stat(%d) error(%+v)", sid, err)
	}
	return
}

// Stats is
func (s *Service) Stats(c context.Context, sids []int64) (stats map[int64]*api.Stat, err error) {
	if stats, err = s.d.Stats(c, sids); err != nil {
		log.Error("s.d.Stats(%v) error(%+v)", sids, err)
	}
	return
}

// Season is
func (s *Service) Season(c context.Context, sid int64) (*api.Season, error) {
	season, err := s.d.Season(c, sid)
	if err != nil {
		log.Error("s.d.Season(%d) error(%+v)", sid, err)
		return nil, err
	}
	//获取付费合集绑定的商品
	if season.AttrVal(api.SeasonAttrSnPay) == api.AttrSnYes {
		season.GoodsInfo = s.d.GetGoodsInfoFromCache(c, strconv.FormatInt(season.ID, 10))
	}
	return season, nil
}

// Seasons is
func (s *Service) Seasons(c context.Context, sids []int64) (map[int64]*api.Season, error) {
	seasons, err := s.d.Seasons(c, sids)
	if err != nil {
		log.Error("s.d.Seasons(%+v) error(%+v)", sids, err)
		return nil, err
	}
	return seasons, nil
}

// View get a Season view by sid.
func (s *Service) View(c context.Context, sid int64) (view *api.View, err error) {
	var (
		epAids       []int64
		archiveViews map[int64]*arcapi.ViewReply
		arcm         map[int64]*arcapi.Arc
		pagem        map[int64]*arcapi.Page
	)
	if view, err = s.d.ViewWithStat(c, sid); err != nil {
		log.Error("s.d.View(%d) error(%+v)", sid, err)
		return
	}
	for _, sec := range view.Sections {
		for _, ep := range sec.Episodes {
			epAids = append(epAids, ep.Aid)
		}
	}
	if archiveViews, err = s.batchViews(c, epAids); err != nil {
		log.Error("s.batchViews err(%v) aids(%v)", err, epAids)
		return
	}
	arcm = make(map[int64]*arcapi.Arc, len(epAids))
	pagem = make(map[int64]*arcapi.Page)
	for aid, v := range archiveViews {
		arcm[aid] = v.Arc
		for _, p := range v.Pages {
			pagem[p.Cid] = p
		}
	}
	for _, sec := range view.Sections {
		for _, ep := range sec.Episodes {
			var (
				arc  *arcapi.Arc
				page *arcapi.Page
				ok   bool
			)
			if arc, ok = arcm[ep.Aid]; !ok {
				continue
			}
			if page, ok = pagem[ep.Cid]; !ok {
				continue
			}
			ep.Arc = &api.Arc{
				Pic:      arc.Pic,
				PubDate:  arc.PubDate,
				Duration: arc.Duration,
				Stat: &api.ArcStat{
					Aid:     arc.Aid,
					View:    arc.Stat.View,
					Danmaku: arc.Stat.Danmaku,
					Reply:   arc.Stat.Reply,
					Fav:     arc.Stat.Fav,
					Coin:    arc.Stat.Coin,
					Share:   arc.Stat.Share,
					NowRank: arc.Stat.NowRank,
					HisRank: arc.Stat.HisRank,
					Like:    arc.Stat.Like,
				},
				Author: &api.Author{
					Mid:  arc.Author.Mid,
					Name: arc.Author.Name,
					Face: arc.Author.Face,
				},
				Attribute:   int64(arc.Attribute),
				AttributeV2: arc.AttributeV2,
				Title:       arc.Title,
				FirstFrame:  arc.FirstFrame,
			}
			ep.Page = &api.ArcPage{
				Cid:      page.Cid,
				Page:     page.Page,
				From:     page.From,
				Part:     page.Part,
				Duration: page.Duration,
				Vid:      page.Vid,
				Desc:     page.Desc,
				WebLink:  page.WebLink,
				Dimension: api.Dimension{
					Width:  page.Dimension.Width,
					Height: page.Dimension.Height,
					Rotate: page.Dimension.Rotate,
				},
			}
		}
	}
	return
}

// UpSeasonCache update season cache
func (s *Service) UpSeasonCache(c context.Context, sid int64, action string) (err error) {
	log.Warn("UpSeasonCache start sid(%d) action(%s)", sid, action)
	switch action {
	case "update":
		if err = s.d.UpCache(c, sid); err != nil {
			log.Error("s.d.UpSeasonCache(%d) error(%+v)", sid, err)
		}
	case "delete":
		if err = s.d.DelCache(c, sid); err != nil {
			log.Error("s.d.DelSeasonCache(%d) error(%+v)", sid, err)
		}
	}
	return
}

// UpperSeason up season list
func (s *Service) UpperSeason(c context.Context, req *api.UpperListRequest) (seasons []*api.Season, totalCount int64, totalPage int64, err error) {
	seasons, totalCount, totalPage, err = s.d.UpperSeason(c, req)
	return
}

func (s *Service) batchViews(c context.Context, aids []int64) (views map[int64]*arcapi.ViewReply, err error) {
	var (
		aidsLen = len(aids)
		mutex   = sync.Mutex{}
	)
	views = make(map[int64]*arcapi.ViewReply, aidsLen)
	eg := errgroup.WithCancel(c)
	for i := 0; i < aidsLen; i += _maxAids {
		var partAids []int64
		if i+_maxAids > aidsLen {
			partAids = aids[i:]
		} else {
			partAids = aids[i : i+_maxAids]
		}
		eg.Go(func(ctx context.Context) (err error) {
			var tmpRes *arcapi.ViewsReply
			arg := &arcapi.ViewsRequest{Aids: partAids}
			if tmpRes, err = s.d.ArcClient.Views(ctx, arg); err != nil {
				log.Error("s.d.ArcClient.Views arg(%v) err(%v)", arg, err)
				return
			}
			if tmpRes != nil && len(tmpRes.Views) > 0 {
				mutex.Lock()
				for aid, view := range tmpRes.Views {
					views[aid] = view
				}
				mutex.Unlock()
			}
			return err
		})
	}
	if err = eg.Wait(); err != nil {
		return
	}
	return
}

// View get multi Season view by sids.
func (s *Service) Views(c context.Context, sids []int64, epSize int64) (map[int64]*api.View, error) {
	views, err := s.d.ViewsWithStat(c, sids)
	if err != nil {
		log.Error("s.d.View(%+v) error(%+v)", sids, err)
		return nil, err
	}
	res, epAids := func() (map[int64]*api.View, []int64) {
		var (
			allAids []int64
			vs      = make(map[int64]*api.View)
		)
		for sid, sn := range views {
			var (
				aids []int64
				secs []*api.Section
			)
			tmpSn := &api.View{}
			*tmpSn = *sn
			for _, sec := range tmpSn.Sections {
				tmpSec := &api.Section{}
				*tmpSec = *sec
				var eps []*api.Episode
				for _, ep := range tmpSec.Episodes {
					tmpEp := &api.Episode{}
					*tmpEp = *ep
					if len(aids) >= int(epSize) {
						break
					}
					eps = append(eps, ep)
					aids = append(aids, tmpEp.Aid)
					allAids = append(allAids, tmpEp.Aid)
				}
				tmpSec.Episodes = eps
				secs = append(secs, tmpSec)
				if len(aids) >= int(epSize) {
					break
				}
			}
			tmpSn.Sections = secs
			vs[sid] = tmpSn
		}
		return vs, allAids
	}()
	archiveViews, err := s.batchViews(c, epAids)
	if err != nil {
		log.Error("s.batchViews err(%v) aids(%v)", err, epAids)
		return nil, err
	}
	arcm := make(map[int64]*arcapi.Arc, len(epAids))
	pagem := make(map[int64]*arcapi.Page)
	for aid, v := range archiveViews {
		arcm[aid] = v.Arc
		for _, p := range v.Pages {
			pagem[p.Cid] = p
		}
	}
	for _, sn := range res {
		for _, sec := range sn.Sections {
			for _, ep := range sec.Episodes {
				arc, arcOk := arcm[ep.Aid]
				if !arcOk {
					continue
				}
				page, pageOk := pagem[ep.Cid]
				if !pageOk {
					continue
				}
				ep.Arc = &api.Arc{
					Pic:     arc.Pic,
					PubDate: arc.PubDate,
					Stat: &api.ArcStat{
						Aid:     arc.Aid,
						View:    arc.Stat.View,
						Danmaku: arc.Stat.Danmaku,
						Reply:   arc.Stat.Reply,
						Fav:     arc.Stat.Fav,
						Coin:    arc.Stat.Coin,
						Share:   arc.Stat.Share,
						NowRank: arc.Stat.NowRank,
						HisRank: arc.Stat.HisRank,
						Like:    arc.Stat.Like,
					},
					Author: &api.Author{
						Mid:  arc.Author.Mid,
						Name: arc.Author.Name,
						Face: arc.Author.Face,
					},
					Title:      arc.Title,
					FirstFrame: arc.FirstFrame,
				}
				ep.Page = &api.ArcPage{
					Cid:      page.Cid,
					Page:     page.Page,
					From:     page.From,
					Part:     page.Part,
					Duration: page.Duration,
					Vid:      page.Vid,
					Desc:     page.Desc,
					WebLink:  page.WebLink,
					Dimension: api.Dimension{
						Width:  page.Dimension.Width,
						Height: page.Dimension.Height,
						Rotate: page.Dimension.Rotate,
					},
				}
			}
		}
	}
	return res, nil
}
