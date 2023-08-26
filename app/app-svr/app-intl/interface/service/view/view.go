package view

import (
	"context"
	"time"

	"go-common/library/ecode"
	"go-common/library/log"
	"go-common/library/net/metadata"
	"go-common/library/sync/errgroup"
	"go-common/library/text/translate/chinese.v2"

	errgroupv2 "go-common/library/sync/errgroup.v2"

	"go-gateway/app/app-svr/app-intl/interface/model"
	"go-gateway/app/app-svr/app-intl/interface/model/view"
	"go-gateway/app/app-svr/archive/service/api"
	arcgrpc "go-gateway/app/app-svr/archive/service/api"
	steinApi "go-gateway/app/app-svr/steins-gate/service/api"
	xecode "go-gateway/ecode"
	"go-gateway/pkg/idsafe/bvid"

	cfcgrpc "git.bilibili.co/bapis/bapis-go/content-flow-control/service"
	vuApi "git.bilibili.co/bapis/bapis-go/videoup/open/service"
)

const (
	// _descLen is.
	_descLen = 250
	// _avTypeAv is.
	_avTypeAv = 1
	// _businessLike is.
	_businessLike = "archive"
)

// View all view data.
func (s *Service) View(c context.Context, mid, aid, movieID int64, plat int8, build, autoplay int, ak, mobiApp, device, buvid, cdnIP, network, adExtra, from, trackid string, now time.Time, locale, filtered string) (v *view.View, err error) {
	if v, err = s.ViewPage(c, mid, aid, movieID, plat, build, ak, mobiApp, device, cdnIP, true, now, locale, buvid); err != nil {
		ip := metadata.String(c, metadata.RemoteIP)
		if err == ecode.AccessDenied || err == ecode.NothingFound {
			log.Warn("s.ViewPage() mid(%d) aid(%d) movieID(%d) plat(%d) ak(%s) ip(%s) cdn_ip(%s) error(%v)", mid, aid, movieID, plat, ak, ip, cdnIP, err)
		} else {
			log.Error("s.ViewPage() mid(%d) aid(%d) movieID(%d) plat(%d) ak(%s) ip(%s) cdn_ip(%s) error(%v)", mid, aid, movieID, plat, ak, ip, cdnIP, err)
		}
		return
	}
	if v != nil {
		v.Config = &view.Config{
			RelatesTitle: s.c.ViewConfig.RelatesTitle,
		}
	}
	isTW := model.TWLocale(locale)
	g, ctx := errgroup.WithContext(c)
	g.Go(func() (err error) {
		s.initReqUser(ctx, v, mid)
		return
	})
	g.Go(func() (err error) {
		s.initRelateCMTag(ctx, v, plat, build, autoplay, mid, buvid, from, trackid, filtered, isTW)
		return
	})
	if v.AttrVal(arcgrpc.AttrBitIsPGC) != arcgrpc.AttrYes {
		g.Go(func() (err error) {
			s.initDM(ctx, v)
			return
		})
		g.Go(func() (err error) {
			s.initAudios(ctx, v)
			return
		})
		if len([]rune(v.Desc)) > _descLen {
			g.Go(func() (err error) {
				if desc, _ := s.arcDao.Description(ctx, v.Aid); desc != "" {
					v.Desc = desc
				}
				return
			})
		}
	}
	if err = g.Wait(); err != nil {
		log.Error("%+v", err)
	}
	var tagIDs []int64
	for _, tag := range v.Tag {
		tagIDs = append(tagIDs, tag.TagID)
	}
	v.PlayerIcon, _ = s.rscDao.PlayerIcon(c, v.Aid, tagIDs, v.TypeID)
	s.initLabel(v)
	return
}

// ViewPage view page data.
// nolint:gocognit
func (s *Service) ViewPage(c context.Context, mid, aid, movieID int64, plat int8, build int, ak, mobiApp, device, cdnIP string, nMovie bool, now time.Time, locale, buvid string) (v *view.View, err error) {
	if aid == 0 && movieID == 0 {
		err = ecode.NothingFound
		return
	}
	var (
		vs               *view.ViewStatic
		vp               *api.ViewReply
		seasoninfo       map[int64]int64
		ok               bool
		bvID             string
		arcAddit         *vuApi.ArcViewAdditReply
		flowInfosV2Reply *cfcgrpc.FlowCtlInfosV2Reply
	)
	eg := errgroupv2.WithContext(c)
	eg.Go(func(ctx context.Context) error {
		if movieID != 0 {
			if seasoninfo, err = s.banDao.SeasonidAid(ctx, movieID, now); err != nil {
				log.Error("%+v", err)
				return ecode.NothingFound
			}
			if aid, ok = seasoninfo[movieID]; !ok || aid == 0 {
				return ecode.NothingFound
			}
			var a *api.Arc
			if a, err = s.arcDao.Arc(ctx, aid); err != nil {
				log.Error("%+v", err)
				return ecode.NothingFound
			}
			if a == nil {
				return ecode.NothingFound
			}
			vs = &view.ViewStatic{Arc: a}
			s.prom.Incr("from_movieID")
		} else {
			if vp, err = s.arcDao.View(ctx, aid); err != nil {
				log.Error("%+v", err)
				return ecode.NothingFound
			}
			if vp == nil || !vp.IsNormal() || vp.Videos == 0 || vp.AttrVal(arcgrpc.AttrBitIsMovie) == arcgrpc.AttrYes || vp.AttrVal(arcgrpc.AttrBitIsPUGVPay) == arcgrpc.AttrYes {
				log.Error("aid(%d) state(%d) videos(%d) can not view", aid, vp.State, vp.GetVideos())
				return ecode.NothingFound
			}
			vs = &view.ViewStatic{Arc: vp.Arc}
			if s.displaySteins(vs, mobiApp, build) {
				vp.Pages = []*api.Page{}
			} else {
				s.initPages(ctx, vs, vp.Pages)
			}
			s.prom.Incr("from_aid")
		}
		return nil
	})
	eg.Go(func(ctx context.Context) error {
		flowInfosV2Reply, err = s.cfcDao.ContentFlowControlInfosV2(ctx, []int64{aid})
		if err != nil {
			log.Error("s.cfc.ContentFlowControlInfosV2 err%+v", err)
			return nil
		}
		return nil
	})
	if err = eg.Wait(); err != nil {
		log.Error("%+v", err)
		return nil, err
	}
	isTW := model.TWLocale(locale)
	if isTW {
		out := chinese.Converts(c, vs.Title, vs.Desc)
		vs.Title = out[vs.Title]
		vs.Desc = out[vs.Desc]
	}
	vs.Stat.DisLike = 0

	if s.overseaCheck(flowInfosV2Reply, vs.Arc.Aid, plat) {
		err = ecode.AreaLimit
		return
	}
	// check region area limit
	if err = s.areaLimit(c, plat, int16(vs.TypeID)); err != nil {
		return
	}
	if bvID, err = bvid.AvToBv(vs.Aid); err != nil {
		log.Error("avtobv aid:%d err(%v)", vs.Aid, err)
		err = nil
	}
	v = &view.View{ViewStatic: vs, DMSeg: 1, BvID: bvID}
	if v.AttrVal(arcgrpc.AttrBitIsPGC) != arcgrpc.AttrYes {
		// check access
		if err = s.checkAceess(c, mid, v.Aid, int(v.State), int(v.Access), ak); err != nil {
			// archive is ForbitFixed and Transcoding and StateForbitDistributing need analysis history body .
			if v.State != arcgrpc.StateForbidFixed {
				return
			}
			err = nil
		}
		if v.Access > 0 {
			v.Stat.View = 0
		}
	}
	g, ctx := errgroup.WithContext(c)
	// 校验稿件审核屏蔽状态
	g.Go(func() (err error) {
		if arcAddit, err = s.vuDao.ArcViewAddit(ctx, aid); err != nil || arcAddit == nil {
			log.Error("s.vuDao.ArcViewAddit aid(%d) err(%+v) or arcAddit=nil", aid, err)
			err = nil
			return
		}
		if arcAddit.ForbidReco != nil {
			v.ForbidRec = arcAddit.ForbidReco.State
		}
		return
	})

	if s.displaySteins(vs, mobiApp, build) {
		g.Go(func() (err error) {
			var steinView *steinApi.ViewReply
			if steinView, err = s.steinDao.View(c, aid, mid, buvid); err != nil {
				log.Error("s.steinDao.View err(%v)", err)
				if ecode.EqualError(xecode.NonValidGraph, err) {
					err = ecode.NothingFound
				}
				return
			}
			if steinView.Graph == nil {
				err = ecode.NothingFound
				return
			}
			vp.Pages = []*api.Page{view.ArchivePage(steinView.Page)}
			vp.FirstCid = steinView.Page.Cid
			v.Interaction = &view.Interaction{
				GraphVersion: steinView.Graph.Id,
				Mark:         steinView.Mark,
			}
			if steinView.Evaluation != "" { // 稿件综合评分
				v.Interaction.Evaluation = steinView.Evaluation
			}
			if steinView.ToastMsg != "" {
				v.Interaction.Msg = steinView.ToastMsg
			}
			if steinView.CurrentNode != nil {
				v.Interaction.HistoryNode = &view.Node{
					CID:    steinView.CurrentNode.Cid,
					Title:  steinView.CurrentNode.Name,
					NodeID: steinView.CurrentNode.Id,
				}
			}
			s.initPages(c, vs, vp.Pages)
			return
		})
	}
	if ((plat == model.PlatAndroidI && build > s.c.ViewBuildLimit.SteinsSeasonBuildAndroid) || (plat == model.PlatIPhoneI && build > s.c.ViewBuildLimit.SteinsSeasonBuildIOS)) && v.SeasonID != 0 {
		g.Go(func() (err error) {
			if ugcSn, err := s.seasonDao.Season(ctx, v.SeasonID); err == nil && ugcSn != nil { // ugc剧集
				v.UgcSeason = new(view.UgcSeason)
				v.UgcSeason.FromSeason(ugcSn)
				s.prom.Incr("Season_Show")
			}
			return
		})
	}
	if mid != 0 {
		g.Go(func() (err error) {
			v.History, _ = s.arcDao.Progress(ctx, v.Aid, mid)
			return
		})
	}
	if v.AttrVal(arcgrpc.AttrBitIsPGC) == arcgrpc.AttrYes {
		g.Go(func() error {
			return s.initPGC(ctx, v, mid, build, mobiApp, device)
		})
	} else {
		g.Go(func() (err error) {
			if err = s.initDownload(ctx, v, mid, cdnIP); err != nil {
				ip := metadata.String(ctx, metadata.RemoteIP)
				log.Error("aid(%d) mid(%d) ip(%s) cdn_ip(%s) error(%+v)", v.Aid, mid, ip, cdnIP, err)
			}
			return
		})
	}
	if err = g.Wait(); err != nil {
		log.Error("%+v", err)
	}
	return
}

func (s *Service) displaySteins(a *view.ViewStatic, mobiApp string, build int) bool {
	return a.AttrVal(arcgrpc.AttrBitSteinsGate) == arcgrpc.AttrYes && ((mobiApp == "android_i" && build > s.c.Custom.SteinsSeasonBuild) || mobiApp == "iphone_i")
}

func HideArcAttribute(arc *api.Arc) {
	if arc == nil {
		return
	}
	arc.Access = 0
	arc.Attribute = 0
	arc.AttributeV2 = 0
}
