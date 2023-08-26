package service

import (
	"context"
	"encoding/json"
	"strconv"
	"time"

	"go-common/library/log"
	"go-common/library/queue/databus"
	"go-common/library/sync/errgroup"
	xtime "go-common/library/time"
	"go-gateway/app/app-svr/app-job/job/model"
	"go-gateway/app/app-svr/app-job/job/model/space"
	"go-gateway/app/app-svr/archive/service/api"

	article "git.bilibili.co/bapis/bapis-go/article/model"
)

const (
	_sleep             = 100 * time.Millisecond
	_upContributeRetry = 5
)

var vmidm = map[int64]time.Time{}

// contributeConsumeproc consumer contribute
func (s *Service) contributeConsumeproc() {
	defer s.waiter.Done()
	var (
		msg *databus.Message
		ok  bool
		err error
	)
	msgs := s.contributeSub.Messages()
	for {
		if msg, ok = <-msgs; !ok {
			close(s.contributeChan)
			log.Info("arc databus Consumer exit")
			break
		}
		var ms = &model.ContributeMsg{}
		if err = json.Unmarshal(msg.Value, ms); err != nil {
			_ = msg.Commit()
			log.Error("json.Unmarshal(%s) error(%v)", msg.Value, err)
			continue
		}
		now, t := time.Now(), vmidm[ms.Vmid]
		if now.Sub(t) > time.Second {
			s.contributeChan <- ms
			vmidm[ms.Vmid] = now
			log.Info("contributeConsumeproc vmid(%d) success", ms.Vmid)
		} else {
			log.Info("contributeConsumeproc vmid(%d) limited", ms.Vmid)
		}
		_ = msg.Commit()
	}
}

func (s *Service) contributeproc() {
	defer s.waiter.Done()
	var (
		ms *model.ContributeMsg
		ok bool
	)
	for {
		if ms, ok = <-s.contributeChan; !ok {
			log.Error("s.contributeChan id closed")
			break
		}
		_ = s.contributeCache(ms.Vmid, ms.Attrs, ms.CTime, ms.IP, ms.IsCooperation, ms.IsComic, "proc")
	}
}

// nolint:gocognit
func (s *Service) contributeCache(vmid int64, attrs *space.Attrs, ctime xtime.Time, ip string, isCooperation, isComic bool, from string) (err error) {
	if vmid == 0 {
		return
	}
	var (
		items    []*space.Item
		archives []*api.Arc
		articles []*article.Meta
		audios   []*space.Audio
		comics   []*space.Comic
	)
	c := context.Background()
	if attrs == nil {
		attrs = &space.Attrs{}
		if err = s.spdao.DelContributeCache(c, vmid, isCooperation, isComic); err != nil {
			log.Error("%+v", err)
		}
	}
	g, ctx := errgroup.WithContext(c)
	g.Go(func() (err error) {
		var pn, ps int64 = 1, 20
		for {
			var as []*api.Arc
			if err = retry(func() (err error) {
				upArcs, err := s.spdao.UpArcs(ctx, vmid, pn, ps, isCooperation)
				if err != nil {
					log.Error("%+v", err)
				}
				for _, v := range upArcs {
					a := &api.Arc{
						Aid:     v.Aid,
						PubDate: v.PubDate,
					}
					as = append(as, a)
				}
				return
			}, _upContributeRetry, _sleep); err != nil {
				log.Error("%+v", err)
				return
			} else if len(as) == 0 {
				break
			}
			archives = append(archives, as...)
			if attrs.Archive {
				a := as[len(as)-1]
				if a != nil && a.PubDate < ctime {
					break
				}
			}
			pn++
		}
		return
	})
	g.Go(func() (err error) {
		pn, ps := 1, 20
		for {
			var ats []*article.Meta
			if err = retry(func() (err error) {
				if ats, _, err = s.spdao.UpArticles(ctx, vmid, pn, ps); err != nil {
					log.Error("%+v", err)
				}
				return
			}, _upContributeRetry, _sleep); err != nil {
				log.Error("%+v", err)
				return
			} else if len(ats) == 0 {
				break
			}
			articles = append(articles, ats...)
			if attrs.Article {
				at := ats[len(ats)-1]
				if at != nil && at.PublishTime < ctime {
					break
				}
			}
			pn++
		}
		return
	})
	g.Go(func() (err error) {
		pn, ps := 1, 20
		for {
			var (
				aus      []*space.Audio
				hasNext  bool
				nextPage int
			)
			if err = retry(func() (err error) {
				if aus, hasNext, nextPage, err = s.spdao.AudioList(c, vmid, pn, ps, ip); err != nil {
					log.Error("%+v", err)
				}
				return
			}, _upContributeRetry, _sleep); err != nil {
				log.Error("%+v", err)
				return
			}
			if len(aus) != 0 {
				audios = append(audios, aus...)
				if attrs.Audio {
					au := aus[len(aus)-1]
					if au != nil && au.CTime < ctime {
						break
					}
				}
			}
			if !hasNext {
				break
			}
			pn = nextPage
		}
		return
	})
	if isComic {
		g.Go(func() (err error) {
			pn, ps := 1, 20
			for {
				var cs []*space.Comic
				if err = retry(func() (err error) {
					if cs, err = s.spdao.UpComics(ctx, vmid, pn, ps); err != nil {
						log.Error("%v", err)
					}
					return
				}, _upContributeRetry, _sleep); err != nil {
					log.Error("%v", err)
					return
				} else if len(cs) == 0 {
					break
				}
				comics = append(comics, cs...)
				if attrs.Comic {
					c := cs[len(cs)-1]
					if c != nil {
						update, _ := strconv.ParseInt(c.LastUpdateTime, 10, 64)
						if xtime.Time(update) < ctime {
							break
						}
					}
				}
				pn++
			}
			return
		})
	}
	if err = g.Wait(); err != nil {
		_ = s.syncRetry(c, model.ActionUpContributeAid, vmid, 0, attrs, nil, ctime, "", isCooperation, isComic)
		log.Error("%+v", err)
		return
	}
	if len(archives) != 0 {
		attrs.Archive = true
	}
	if len(articles) != 0 {
		attrs.Article = true
	}
	if len(audios) != 0 {
		attrs.Audio = true
	}
	if len(comics) != 0 {
		attrs.Comic = true
	}
	items = make([]*space.Item, 0, len(archives)+len(articles)+len(audios)+len(comics))
	for _, a := range archives {
		if a != nil {
			item := &space.Item{ID: a.Aid, Goto: space.GotoAv, CTime: a.PubDate}
			items = append(items, item)
		}
	}
	for _, a := range articles {
		if a != nil {
			item := &space.Item{ID: a.ID, Goto: space.GotoArticle, CTime: a.PublishTime}
			items = append(items, item)
		}
	}
	for _, a := range audios {
		if a != nil {
			item := &space.Item{ID: a.ID, Goto: space.GotoAudio, CTime: a.CTime}
			items = append(items, item)
		}
	}
	for _, c := range comics {
		if c != nil {
			update, _ := strconv.ParseInt(c.LastUpdateTime, 10, 64)
			item := &space.Item{ID: c.ID, Goto: space.GotoComic, CTime: xtime.Time(update)}
			items = append(items, item)
		}
	}
	log.Info("vmid(%d) attrs(%+v) update contribute start ctime(%d) from(%s) items len(%d)", vmid, attrs, ctime, from, len(items))
	s.updateContribute(vmid, attrs, items, isCooperation, isComic)
	return
}

func (s *Service) updateContribute(vmid int64, attrs *space.Attrs, items []*space.Item, isCooperation, isComic bool) {
	c := context.Background()
	log.Info("updateContribute vmid(%d) len(%d)", vmid, len(items))
	if err := retry(func() (err error) {
		if items, err = s.spdao.AddContributeList(c, vmid, items, isCooperation, isComic); err == nil {
			err = s.spdao.AddContributeAttr(c, vmid, attrs, isCooperation, isComic)
		}
		return
	}, _upContributeRetry, time.Second); err != nil {
		log.Error("vmid(%d) attrs(%+v) update contribute failed error(%v)", vmid, attrs, err)
		_ = s.syncRetry(c, model.ActionUpContribute, vmid, 0, attrs, items, 0, "", isCooperation, isComic)
	} else {
		log.Info("vmid(%d) attrs(%+v) update contribute success", vmid, attrs)
	}
}
