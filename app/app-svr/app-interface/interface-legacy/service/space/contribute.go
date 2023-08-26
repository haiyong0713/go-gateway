package space

import (
	"context"
	"sort"
	"time"

	"go-common/library/log"
	"go-common/library/sync/errgroup"
	xtime "go-common/library/time"

	"go-gateway/app/app-svr/archive/service/api"

	"go-gateway/app/app-svr/app-interface/interface-legacy/model"
	"go-gateway/app/app-svr/app-interface/interface-legacy/model/audio"
	"go-gateway/app/app-svr/app-interface/interface-legacy/model/comic"
	"go-gateway/app/app-svr/app-interface/interface-legacy/model/space"
	ugcSeasonGrpc "go-gateway/app/app-svr/ugc-season/service/api"

	article "git.bilibili.co/bapis/bapis-go/article/model"
	uparcgrpc "git.bilibili.co/bapis/bapis-go/up-archive/service"
)

const (
	_androidAudio = 516009
	_iosAudio     = 6160
)

// Contribute func
//
//nolint:gocognit
func (s *Service) Contribute(c context.Context, plat int8, build int, vmid int64, pn, ps int, now time.Time, mobiApp, device string, mid int64) (res *space.Contributes, err error) {
	var (
		attrs                               *space.Attrs
		items                               []*space.Item
		isCooperation, isUGCSeason, isComic bool
	)
	if (model.IsAndroid(plat) && build > s.c.SpaceBuildLimit.CooperationAndroid) || (model.IsIOS(plat) && build > s.c.SpaceBuildLimit.CooperationIOS) {
		isCooperation = true
	}
	if (mobiApp == "android" && build > s.c.SpaceBuildLimit.UGCSeasonAndroid) || (model.IsIPhone(plat) && build > s.c.SpaceBuildLimit.UGCSeasonIOS) || (mobiApp == "android_i" && build > s.c.SpaceBuildLimit.UGCSeasonAndroidI) {
		isUGCSeason = true
	}
	if (mobiApp == "android" && build > s.c.SpaceBuildLimit.ComicAndroid) || (model.IsIOS(plat) && build > s.c.SpaceBuildLimit.ComicIOS) || (mobiApp == "android_i" && build > s.c.SpaceBuildLimit.ComicAndroidI) {
		isComic = true
	}
	if pn == 1 {
		var (
			ctime  xtime.Time
			cached bool
		)
		size := ps
		if items, err = s.bplusDao.RangeContributeCache(c, vmid, pn, ps, isCooperation, isComic); err != nil {
			log.Error("%+v", err)
		} else if len(items) != 0 {
			ctime = items[0].CTime
		} else {
			size = 50
			cached = true
		}
		if res, err = s.firstContribute(c, vmid, size, now, isCooperation, isComic, isUGCSeason); err != nil {
			log.Error("%+v", err)
			err = nil
		}
		if res != nil && len(res.Items) != 0 {
			if res.Items[0].CTime > ctime {
				if err = s.bplusDao.NotifyContribute(c, vmid, nil, ctime, isCooperation, isComic); err != nil {
					log.Error("%+v", err)
					err = nil
				}
			}
			if cached {
				ris := res.Items
				s.addCache(func() {
					//nolint:errcheck
					s.bplusDao.AddContributeCache(context.Background(), vmid, nil, ris, isCooperation, isComic)
				})
			}
			if len(items) == 0 {
				ris := make([]*space.Item, 0, ps)
				for _, item := range res.Items {
					item.FormatKey()
					switch item.Goto {
					case model.GotoAudio:
						if (plat == model.PlatAndroid && build > _androidAudio) || (plat == model.PlatIPhone && build > _iosAudio) || plat == model.PlatAndroidB {
							ris = append(ris, item)
						}
					default:
						ris = append(ris, item)
					}
					if len(ris) == ps {
						break
					}
				}
				res.Items = ris
				return
			}
		}
	} else {
		if items, err = s.bplusDao.RangeContributeCache(c, vmid, pn, ps, isCooperation, isComic); err != nil {
			return
		}
	}
	if len(items) != 0 {
		if attrs, err = s.bplusDao.AttrCache(c, vmid, isCooperation, isComic); err != nil {
			log.Error("%+v", err)
		}
		// merge res
		if res, err = s.dealContribute(c, plat, build, vmid, attrs, items, now, mobiApp, device, mid); err != nil {
			log.Error("%+v", err)
		}
	}
	if res == nil {
		res = &space.Contributes{Tab: &space.Tab{}, Items: []*space.Item{}, Links: &space.Links{}}
	}
	return
}

// Contribution func
func (s *Service) Contribution(c context.Context, plat int8, build int, vmid int64, cursor *model.Cursor, now time.Time, mobiApp, device string, mid int64) (res *space.Contributes, err error) {
	var (
		attrs *space.Attrs
		items []*space.Item
	)
	if cursor.Latest() {
		var (
			ctime  xtime.Time
			cached bool
		)
		size := cursor.Size
		if items, err = s.bplusDao.RangeContributeCache(c, vmid, 1, 1, false, false); err != nil {
			log.Error("%+v", err)
		} else if len(items) != 0 {
			ctime = items[0].CTime
		} else {
			size = 50
			cached = true
		}
		if res, err = s.firstContribute(c, vmid, size, now, false, false, false); err != nil {
			log.Error("%+v", err)
		}
		if res != nil && len(res.Items) != 0 {
			if cached {
				ris := res.Items
				s.addCache(func() {
					//nolint:errcheck
					s.bplusDao.AddContributeCache(context.Background(), vmid, nil, ris, false, false)
				})
			}
			if res.Items[0].CTime > ctime {
				if len(items) != 0 {
					if attrs, err = s.bplusDao.AttrCache(c, vmid, false, false); err != nil {
						log.Error("%+v", err)
					}
				}
				if err = s.bplusDao.NotifyContribute(c, vmid, attrs, ctime, false, false); err != nil {
					log.Error("%+v", err)
					err = nil
				}
			}
			ris := make([]*space.Item, 0, cursor.Size)
			for _, item := range res.Items {
				item.FormatKey()
				ris = append(ris, item)
				if len(ris) == cursor.Size {
					break
				}
			}
			if len(ris) != 0 {
				res.Items = ris
				res.Links.Link(0, int64(ris[len(ris)-1].Member))
			}
			return
		}
	}
	if items, err = s.bplusDao.RangeContributionCache(c, vmid, cursor); err != nil {
		return
	}
	if len(items) != 0 {
		if attrs, err = s.bplusDao.AttrCache(c, vmid, false, false); err != nil {
			log.Error("%+v", err)
		}
		// merge res
		if res, err = s.dealContribute(c, plat, build, vmid, attrs, items, now, mobiApp, device, mid); err != nil {
			log.Error("%+v", err)
		}
	}
	if res == nil {
		res = &space.Contributes{Tab: &space.Tab{}, Items: []*space.Item{}, Links: &space.Links{}}
	}
	return
}

//nolint:gocognit
func (s *Service) firstContribute(c context.Context, vmid int64, size int, _ time.Time, isCooperation, isComic, isUGCSeason bool) (res *space.Contributes, err error) {
	res = &space.Contributes{Tab: &space.Tab{}, Items: []*space.Item{}, Links: &space.Links{}}
	res.Tab.Clip = false  // 空间小视频下线
	res.Tab.Album = false // 空间相簿下线
	g, ctx := errgroup.WithContext(c)
	var (
		arcItem, artItem, audioItem, comicItem, items []*space.Item
		ugcSeasons                                    *ugcSeasonGrpc.UpperListReply
	)
	g.Go(func() (err error) {
		var arcs []*uparcgrpc.Arc
		if isCooperation {
			if arcs, _, err = s.upArcDao.ArcPassed(ctx, vmid, 1, int64(size), "", nil); err != nil {
				log.Error("s.upArcDao.ArcPassed(%d,%d,%d) error(%v)", vmid, 1, size, err)
				err = nil
				return
			}
		} else {
			without := []uparcgrpc.Without{uparcgrpc.Without_staff}
			if arcs, _, err = s.upArcDao.ArcPassed(ctx, vmid, 1, int64(size), "", without); err != nil {
				log.Error("s.upArcDao.ArcPassed(%d,%d,%d) withoutStaff error(%v)", vmid, 1, size, err)
				err = nil
				return
			}
		}
		if len(arcs) != 0 {
			arcItem = make([]*space.Item, 0, len(arcs))
			for _, v := range arcs {
				if v == nil {
					continue
				}
				if arc := space.FromUpArcToArc(v); arc.IsNormal() {
					si := &space.Item{}
					si.FromArc3(arc, s.hotAids)
					arcItem = append(arcItem, si)
				}
			}
		}
		return
	})
	g.Go(func() (err error) {
		var arts []*article.Meta
		if arts, _, err = s.artDao.UpArticles(ctx, vmid, 1, size); err != nil {
			log.Error("s.artDao.UpArticles(%d,%d,%d) error(%v)", vmid, 1, size, err)
			err = nil
			return
		}
		if len(arts) != 0 {
			artItem = make([]*space.Item, 0, len(arts))
			for _, v := range arts {
				if v.AttrVal(article.AttrBitNoDistribute) {
					continue
				}
				si := &space.Item{}
				si.FromArticle(v)
				artItem = append(artItem, si)
			}
		}
		return
	})
	if isComic {
		g.Go(func() (err error) {
			var comics *comic.Comics
			if comics, err = s.comicDao.UpComics(ctx, vmid, 1, size); err != nil {
				log.Error("s.comicDao.UpComics(%d) error(%v)", vmid, err)
				err = nil
				return
			}
			if len(comics.ComicList) != 0 {
				comicItem = make([]*space.Item, 0, len(comics.ComicList))
				for _, v := range comics.ComicList {
					ci := &space.Item{}
					ci.FromComic(v)
					comicItem = append(comicItem, ci)
				}
			}
			return
		})
	}
	g.Go(func() (err error) {
		var audio []*audio.Audio
		if audio, err = s.audioDao.AllAudio(ctx, vmid); err != nil {
			log.Error("s.audioDao.AllAudio(%d) error(%v)", vmid, err)
			err = nil
			return
		}
		if len(audio) != 0 {
			audioItem = make([]*space.Item, 0, len(audio))
			for _, v := range audio {
				si := &space.Item{}
				si.FromAudio(v)
				audioItem = append(audioItem, si)
			}
		}
		return
	})
	if isUGCSeason {
		g.Go(func() (err error) {
			if ugcSeasons, err = s.ugcSeasonDao.UpperList(ctx, vmid, 1, 20); err != nil {
				log.Error("%+v", err)
				err = nil
			}
			return
		})
	}
	if err = g.Wait(); err != nil {
		log.Error("Contribute errgroup.WithContext error(%v)", err)
	}
	items = make([]*space.Item, 0, len(arcItem)+len(artItem)+len(audioItem)+len(comicItem))
	if len(arcItem) != 0 {
		res.Tab.Archive = true
		items = append(items, arcItem...)
	}
	if len(artItem) != 0 {
		res.Tab.Article = true
		items = append(items, artItem...)
	}
	if len(audioItem) != 0 {
		res.Tab.Audios = true
		items = append(items, audioItem...)
	}
	if ugcSeasons.GetTotalCount() != 0 {
		res.Tab.UGCSeason = true
		// ugc season is archive.
	}
	if len(comicItem) != 0 {
		res.Tab.Comic = true
		items = append(items, comicItem...)
	}
	sort.Sort(space.Items(items))
	res.Items = items
	return
}

//nolint:gocognit
func (s *Service) dealContribute(c context.Context, plat int8, build int, _ int64, attrs *space.Attrs, items []*space.Item, _ time.Time, mobiApp, device string, mid int64) (res *space.Contributes, err error) {
	res = &space.Contributes{Tab: &space.Tab{}, Items: []*space.Item{}, Links: &space.Links{}}
	var aids, cvids, auids, comicids []int64
	if attrs == nil {
		attrs = &space.Attrs{}
	} else if !((plat == model.PlatAndroid && build > _androidAudio) || (plat == model.PlatIPhone && build > _iosAudio) || plat == model.PlatAndroidB) {
		attrs.Audio = false
	}
	for _, item := range items {
		if item.ID == 0 {
			continue
		}
		switch item.Goto {
		case model.GotoAv:
			aids = append(aids, item.ID)
		case model.GotoArticle:
			cvids = append(cvids, item.ID)
		case model.GotoAudio:
			if (plat == model.PlatAndroid && build > _androidAudio) || (plat == model.PlatIPhone && build > _iosAudio) || plat == model.PlatAndroidB {
				auids = append(auids, item.ID)
			}
		case model.GotoComic:
			comicids = append(comicids, item.ID)
		}
	}
	var (
		am     map[int64]*api.Arc
		atm    map[int64]*article.Meta
		aum    map[int64]*audio.Audio
		comicm map[int64]*comic.Comic
	)
	g, ctx := errgroup.WithContext(c)
	if len(aids) != 0 {
		g.Go(func() (err error) {
			if am, err = s.arcDao.Archives(ctx, aids, mobiApp, device, mid); err != nil {
				log.Error("s.arcDao.Archives(%v) error(%v)", aids, err)
				err = nil
			}
			return
		})
	}
	if len(cvids) != 0 {
		g.Go(func() (err error) {
			if atm, err = s.artDao.Articles(ctx, cvids); err != nil {
				log.Error("s.artDao.Articles(%v) error(%v)", cvids, err)
				err = nil
			}
			return
		})
	}
	if len(auids) != 0 {
		g.Go(func() (err error) {
			if aum, err = s.audioDao.AudioDetail(ctx, auids); err != nil {
				log.Error("s.audioDao.AudioDetail(%v) error(%v)", auids, err)
				err = nil
			}
			return
		})
	}
	if len(comicids) != 0 {
		g.Go(func() (err error) {
			if comicm, err = s.comicDao.Comics(ctx, comicids); err != nil {
				log.Error("s.comicDao.Comics(%v) error(%v)", auids, err)
				err = nil
			}
			return
		})
	}
	if err = g.Wait(); err != nil {
		log.Error("Contribute errgroup.WithContext error(%v)", err)
		return
	}
	if len(am) != 0 || attrs.Archive {
		res.Tab.Archive = true
	}
	if len(atm) != 0 || attrs.Article {
		res.Tab.Article = true
	}
	if len(aum) != 0 || attrs.Audio {
		res.Tab.Audios = true
	}
	if len(comicm) != 0 || attrs.Comic {
		res.Tab.Audios = true
	}
	ris := make([]*space.Item, 0, len(items))
	for _, item := range items {
		ri := &space.Item{}
		switch item.Goto {
		case model.GotoAv:
			if a, ok := am[item.ID]; ok && a.IsNormal() {
				ri.FromArc3(a, s.hotAids)
			}
		case model.GotoArticle:
			if at, ok := atm[item.ID]; ok {
				ri.FromArticle(at)
			}
		case model.GotoAudio:
			if au, ok := aum[item.ID]; ok {
				ri.FromAudio(au)
			}
		case model.GotoComic:
			if comic, ok := comicm[item.ID]; ok {
				ri.FromComic(comic)
			}
		}
		if ri.Goto != "" {
			ri.FormatKey()
			ris = append(ris, ri)
		}
	}
	res.Items = ris
	res.Links.Link(int64(items[0].Member), int64(items[len(items)-1].Member))
	return
}

// AddContribute func
func (s *Service) AddContribute(c context.Context, vmid int64, attrs *space.Attrs, items []*space.Item, isCooperation, isComic bool) (err error) {
	if err = s.bplusDao.AddContributeCache(c, vmid, attrs, items, isCooperation, isComic); err != nil {
		log.Error("%+v", err)
	}
	return
}
