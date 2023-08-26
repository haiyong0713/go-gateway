package channel

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"sync"
	"time"

	"go-common/library/log"
	"go-common/library/sync/errgroup"
	cdm "go-gateway/app/app-svr/app-card/interface/model"
	cardm "go-gateway/app/app-svr/app-card/interface/model/card"
	"go-gateway/app/app-svr/app-card/interface/model/card/audio"
	"go-gateway/app/app-svr/app-card/interface/model/card/bangumi"
	"go-gateway/app/app-svr/app-card/interface/model/card/live"
	"go-gateway/app/app-svr/app-card/interface/model/card/operate"
	shopping "go-gateway/app/app-svr/app-card/interface/model/card/show"
	"go-gateway/app/app-svr/app-channel/interface/model"
	"go-gateway/app/app-svr/app-channel/interface/model/card"
	"go-gateway/app/app-svr/app-channel/interface/model/feed"
	arcgrpc "go-gateway/app/app-svr/archive/service/api"

	accountgrpc "git.bilibili.co/bapis/bapis-go/account/service"
	relationgrpc "git.bilibili.co/bapis/bapis-go/account/service/relation"
	article "git.bilibili.co/bapis/bapis-go/article/model"
	taggrpc "git.bilibili.co/bapis/bapis-go/community/interface/tag"
	cfcgrpc "git.bilibili.co/bapis/bapis-go/content-flow-control/service"
	episodegrpc "git.bilibili.co/bapis/bapis-go/pgc/service/season/episode"
	seasongrpc "git.bilibili.co/bapis/bapis-go/pgc/service/season/season"

	farm "go-farm"
)

const (
	_fTypeOperation = "operation"
	_fTypeRecommend = "recommend"
	_newThreePoint  = 1
)

// Index channel index
func (s *Service) Index2(c context.Context, mid, channelID, idx int64, plat int8, mobiApp, device, platform, buvid, channelName, fromSpmid, fromPage string,
	build, loginEvent, displayID, teenagersMode int, pull bool, now time.Time) (res *feed.Show2, err error) {
	var (
		aiCards           []*card.Card
		requestCnt        = 10
		isIpad            = plat == model.PlatIPad
		topic             cardm.Handler
		item              []cardm.Handler
		channelResource   *taggrpc.ChannelResourcesReply
		topChannel, isRec int
		infocs            []*feed.Item
		abtest            *feed.Abtest
	)
	if isIpad {
		requestCnt = 20
	}
	if channelID > 0 {
		channelName = ""
	}
	g, ctx := errgroup.WithContext(c)
	g.Go(func() (err error) {
		if channelResource, err = s.tg.Resources(ctx, plat, channelID, mid, channelName, buvid, build, requestCnt, loginEvent, displayID, teenagersMode); err != nil {
			log.Error("index s.tg.Resources error(%v)", err)
			return
		}
		if channelResource != nil {
			aids := channelResource.Oids
			for _, aid := range aids {
				t := &card.Card{
					Type:     model.GotoAv,
					Value:    aid,
					FromType: _fTypeRecommend,
				}
				aiCards = append(aiCards, t)
			}
			if channelResource.Failover {
				isRec = 0
			} else {
				isRec = 1
			}
			if channelResource.WhetherChannel {
				topChannel = 1
			} else {
				topChannel = 0
			}
		}
		return
	})
	g.Go(func() (err error) {
		var t *taggrpc.ChannelReply
		if t, err = s.tg.ChannelDetail(c, mid, channelID, channelName, s.isOverseas(plat)); err != nil {
			log.Error("s.tag.ChannelDetail(%d, %d, %s) error(%v)", mid, channelID, channelName, err)
			return
		}
		channelID = t.GetChannel().Id
		channelName = t.GetChannel().Name
		return
	})
	err = g.Wait()
	//infoc
	infoc := &feedInfoc{
		mobiApp:     mobiApp,
		device:      device,
		build:       strconv.Itoa(build),
		now:         now.Format("2006-01-02 15:04:05"),
		pull:        strconv.FormatBool(pull),
		loginEvent:  strconv.Itoa(loginEvent),
		channelID:   strconv.FormatInt(channelID, 10),
		channelName: channelName,
		mid:         strconv.FormatInt(mid, 10),
		buvid:       buvid,
		displayID:   strconv.Itoa(displayID),
		isRec:       strconv.Itoa(isRec),
		topChannel:  strconv.Itoa(topChannel),
		ServerCode:  "0",
		FromSpmid:   fromSpmid,
		FromPage:    fromPage,
	}
	//infoc
	if err != nil {
		log.Error("RankUser errgroup.WithContext error(%v)", err)
		res = &feed.Show2{
			Feed: []cardm.Handler{},
		}
		infoc.Items = []*feed.Item{}
		infoc.ServerCode = err.Error()
		s.infoc(infoc)
		return
	}
	var (
		tmps = []*card.Card{}
	)
	if loginEvent == 1 || loginEvent == 2 {
		if cards, ok := s.cardCache[channelID]; ok {
			isShowCard := s.isShowOperationCards(c, buvid, channelID, cards, now)
			for _, c := range cards {
				if !isShowCard && c.Type != model.GotoTopstick {
					continue
				}
				t := &card.Card{}
				*t = *c
				t.FromType = _fTypeOperation
				tmps = append(tmps, t)
			}
			tmps = append(tmps, aiCards...)
		} else {
			tmps = aiCards
		}
	} else {
		tmps = aiCards
	}
	topic, item, infocs, err = s.dealItem2(c, mid, idx, plat, build, mobiApp, platform, pull, now, tmps, abtest)
	res = &feed.Show2{
		Topic: topic,
		Feed:  item,
	}
	infoc.Items = infocs
	s.infoc(infoc)
	return
}

// nolint:gocognit
func (s *Service) dealItem2(c context.Context, mid, idx int64, plat int8, build int, mobiApp, platform string, pull bool,
	now time.Time, cards []*card.Card, abtest *feed.Abtest) (top cardm.Handler, is []cardm.Handler, infocs []*feed.Item, err error) {
	if len(cards) == 0 {
		is = []cardm.Handler{}
		return
	}
	var (
		aids, shopIDs, audioIDs, sids, roomIDs, metaIDs      []int64
		upIDs, tids, rmUpIDs, mtUpIDs, avUpIDs, avUpCountIDs []int64
		seasonIDs, epIDs                                     []int32
		amplayer                                             map[int64]*arcgrpc.ArcPlayer
		tagm                                                 map[int64]*taggrpc.Tag
		rm                                                   map[int64]*live.Room
		sm                                                   map[int64]*bangumi.Season
		metam                                                map[int64]*article.Meta
		shopm                                                map[int64]*shopping.Shopping
		audiom                                               map[int64]*audio.Audio
		cardAids                                             = map[int64]struct{}{}
		ac                                                   map[int64]*accountgrpc.Card
		statm                                                map[int64]*relationgrpc.StatReply
		isAtten                                              map[int64]int8
		upAvCount                                            = map[int64]int{}
		channelCards                                         []*card.Card
		seasonm                                              map[int32]*seasongrpc.CardInfoProto
		epidsCards                                           map[int32]*episodegrpc.EpisodeCardsProto
		flowInfosV2Reply                                     *cfcgrpc.FlowCtlInfosV2Reply
		// haslike                                              map[int64]int8
		// key
		_initCardPlatKey = "card_platkey_%d_%d"
	)
	specialm := map[int64]*operate.Card{}
	convergem := map[int64]*operate.Card{}
	downloadm := map[int64]*operate.Card{}
	followm := map[int64]*operate.Card{}
	liveUpm := map[int64][]*live.Card{}
	cardSet := map[int64]*operate.Card{}
LOOP:
	for _, card := range cards {
		key := fmt.Sprintf(_initCardPlatKey, plat, card.ID)
		if cardPlat, ok := s.cardPlatCache[key]; ok {
			for _, l := range cardPlat {
				if model.InvalidBuild(build, l.Build, l.Condition) {
					continue LOOP
				}
			}
		} else if card.FromType == _fTypeOperation {
			continue LOOP
		}
		channelCards = append(channelCards, card)
		switch card.Type {
		case model.GotoAv, model.GotoPlayer, model.GotoUpRcmdAv:
			if card.Value != 0 {
				aids = append(aids, card.Value)
				cardAids[card.Value] = struct{}{}
			}
		case model.GotoLive, model.GotoPlayerLive:
			if card.Value != 0 {
				roomIDs = append(roomIDs, card.Value)
			}
		case model.GotoBangumi:
			if card.Value != 0 {
				sids = append(sids, card.Value)
			}
		case model.GotoPGC:
			if card.Value != 0 {
				epIDs = append(epIDs, int32(card.Value))
			}
		case model.GotoConverge:
			if card.Value != 0 {
				cardm, aid, roomID, metaID := s.convergeCard2(c, 3, card.Value)
				for id, card := range cardm {
					convergem[id] = card
				}
				aids = append(aids, aid...)
				roomIDs = append(roomIDs, roomID...)
				metaIDs = append(metaIDs, metaID...)
			}
		case model.GotoGameDownload, model.GotoGameDownloadS:
			if card.Value != 0 {
				cardm := s.downloadCard(c, card.Value)
				for id, card := range cardm {
					downloadm[id] = card
				}
			}
		case model.GotoArticle, model.GotoArticleS:
			if card.Value != 0 {
				metaIDs = append(metaIDs, card.Value)
			}
		case model.GotoShoppingS:
			if card.Value != 0 {
				shopIDs = append(shopIDs, card.Value)
			}
		case model.GotoAudio:
			if card.Value != 0 {
				audioIDs = append(audioIDs, card.Value)
			}
		case model.GotoChannelRcmd:
			cardm, aid, tid := s.channelRcmdCard(c, card.Value)
			for id, card := range cardm {
				followm[id] = card
			}
			aids = append(aids, aid...)
			tids = append(tids, tid...)
		case model.GotoLiveUpRcmd:
			if card.Value != 0 {
				cardm, upID := s.liveUpRcmdCard(c, card.Value)
				for id, card := range cardm {
					liveUpm[id] = card
				}
				upIDs = append(upIDs, upID...)
			}
		case model.GotoSubscribe:
			if card.Value != 0 {
				cardm, upID, tid := s.subscribeCard(c, card.Value)
				for id, card := range cardm {
					followm[id] = card
				}
				upIDs = append(upIDs, upID...)
				tids = append(tids, tid...)
			}
		case model.GotoSpecial, model.GotoSpecialS:
			cardm := s.specialCard(c, card.Value)
			for id, card := range cardm {
				specialm[id] = card
			}
		case model.GotoTopstick:
			cardm := s.topstickCard(c, card.Value)
			for id, card := range cardm {
				specialm[id] = card
			}
		case model.GotoPgcsRcmd:
			cardm, ssid := s.cardSetChange(c, card.Value)
			seasonIDs = append(seasonIDs, ssid...)
			for id, card := range cardm {
				cardSet[id] = card
			}
		case model.GotoUpRcmdS:
			if card.Value != 0 {
				avUpCountIDs = append(avUpCountIDs, card.Value)
			}
		}
	}
	g, ctx := errgroup.WithContext(c)
	if len(aids) != 0 {
		var aidsV2 []*arcgrpc.PlayAv
		for _, aid := range aids {
			if aid != 0 {
				aidsV2 = append(aidsV2, &arcgrpc.PlayAv{Aid: aid})
			}
		}
		g.Go(func() (err error) {
			if amplayer, err = s.Archives(ctx, aidsV2, true); err != nil {
				return
			}
			for _, a := range amplayer {
				if a != nil && a.Arc != nil {
					avUpIDs = append(avUpIDs, a.Arc.Author.Mid)
				}
			}
			return
		})
		g.Go(func() error {
			req := makeContentFlowControlInfosV2Params(s.c.CfcSvrConfig, aids)
			flowInfosV2Reply, err = s.arc.ContentFlowControlInfosV2(ctx, req)
			if err != nil {
				log.Error("s.arcDao.ContentFlowControlInfosV2 err=%+v", err)
				return nil
			}
			return nil
		})
	}
	if len(tids) != 0 {
		g.Go(func() (err error) {
			if tagm, err = s.tg.Tags(ctx, mid, tids); err != nil {
				log.Error("%+v", err)
				err = nil
			}
			return
		})
	}
	if len(roomIDs) != 0 {
		g.Go(func() (err error) {
			if rm, err = s.lv.AppMRoom(ctx, roomIDs, platform); err != nil {
				log.Error("%+v", err)
				err = nil
			}
			for _, r := range rm {
				rmUpIDs = append(rmUpIDs, r.UID)
			}
			return
		})
	}
	if len(sids) != 0 {
		g.Go(func() (err error) {
			if sm, err = s.bgm.Seasons(ctx, sids, now); err != nil {
				log.Error("%+v", err)
				err = nil
			}
			return
		})
	}
	if len(seasonIDs) != 0 {
		g.Go(func() (err error) {
			if seasonm, err = s.bgm.CardsInfoReply(ctx, seasonIDs); err != nil {
				log.Error("%+v", err)
				err = nil
			}
			return
		})
	}
	if len(epIDs) != 0 {
		g.Go(func() (err error) {
			if epidsCards, err = s.bgm.EpidsCardsInfoReply(ctx, epIDs); err != nil {
				log.Error("%+v", err)
				err = nil
			}
			return
		})
	}
	if len(metaIDs) != 0 {
		g.Go(func() (err error) {
			if metam, err = s.art.Articles(ctx, metaIDs); err != nil {
				log.Error("%+v", err)
				err = nil
			}
			for _, meta := range metam {
				if meta.Author != nil {
					mtUpIDs = append(mtUpIDs, meta.Author.Mid)
				}
			}
			return
		})
	}
	if len(shopIDs) != 0 {
		g.Go(func() (err error) {
			if shopm, err = s.sp.Card(ctx, shopIDs); err != nil {
				log.Error("%+v", err)
				err = nil
			}
			return
		})
	}
	if len(audioIDs) != 0 {
		g.Go(func() (err error) {
			if audiom, err = s.audio.Audios(ctx, audioIDs); err != nil {
				log.Error("%+v", err)
				err = nil
			}
			return
		})
	}
	if len(avUpCountIDs) != 0 {
		var mutex sync.Mutex
		for _, upid := range avUpCountIDs {
			var (
				tmpupid = upid
			)
			g.Go(func() (err error) {
				var cnt int
				if cnt, err = s.arc.UpCount(ctx, tmpupid); err != nil {
					log.Error("%+v", err)
					err = nil
				}
				mutex.Lock()
				upAvCount[tmpupid] = cnt
				mutex.Unlock()
				return
			})
		}
	}
	if err = g.Wait(); err != nil {
		log.Error("%+v", err)
		return
	}
	upIDs = append(upIDs, avUpIDs...)
	upIDs = append(upIDs, rmUpIDs...)
	upIDs = append(upIDs, mtUpIDs...)
	upIDs = append(upIDs, avUpCountIDs...)
	g, ctx = errgroup.WithContext(c)
	if len(upIDs) != 0 {
		g.Go(func() (err error) {
			if ac, err = s.acc.Cards3GRPC(ctx, upIDs); err != nil {
				log.Error("%+v", err)
				err = nil
			}
			return
		})
		g.Go(func() (err error) {
			if statm, err = s.rel.StatsGRPC(ctx, upIDs); err != nil {
				log.Error("%+v", err)
				err = nil
			}
			return
		})
		if mid != 0 {
			g.Go(func() error {
				isAtten = s.acc.IsAttentionGRPC(ctx, upIDs, mid)
				return nil
			})
		}
	}
	if err = g.Wait(); err != nil {
		log.Error("%+v", err)
		return
	}
	for _, card := range channelCards {
		var (
			r    = card.CardToAiChange()
			main interface{}
		)
		switch r.Goto {
		case model.GotoAv, model.GotoUpRcmdAv, model.GotoPlayer:
			r.HideButton = true
		}
		h := cardm.Handle(plat, cdm.CardGt(r.Goto), "", cdm.ColumnSvrSingle, r, tagm, isAtten, nil, statm, ac, nil)
		if h == nil {
			continue
		}
		op := &operate.Card{}
		op.From(cdm.CardGt(r.Goto), r.ID, 0, plat, build, mobiApp)
		switch r.Goto {
		case model.GotoAv, model.GotoUpRcmdAv, model.GotoPlayer:
			op.ShowUGCPay = true
			if a, ok := amplayer[r.ID]; ok && a != nil && a.Arc != nil {
				if getAttrBitValueFromInfosV2(flowInfosV2Reply, r.ID, model.OverseaBlockKey) == 0 || !model.IsOverseas(plat) {
					main = amplayer
				}
			}
			op.Switch = cdm.SwitchCooperationHide
		case model.GotoLive, model.GotoPlayerLive:
			main = rm
		case model.GotoBangumi:
			main = sm
		case model.GotoPGC:
			main = epidsCards
		case model.GotoSpecial, model.GotoSpecialS, model.GotoTopstick:
			op = specialm[r.ID]
		case model.GotoGameDownload, model.GotoGameDownloadS:
			op = downloadm[r.ID]
		case model.GotoArticle, model.GotoArticleS:
			main = metam
		case model.GotoShoppingS:
			main = shopm
		case model.GotoAudio:
			main = audiom
		case model.GotoChannelRcmd:
			main = amplayer
			op = followm[r.ID]
		case model.GotoSubscribe:
			op = followm[r.ID]
		case model.GotoLiveUpRcmd:
			main = liveUpm
		case model.GotoConverge:
			main = map[cdm.Gt]interface{}{cdm.GotoAv: amplayer, cdm.GotoLive: rm, cdm.GotoArticle: metam}
			op = convergem[r.ID]
		case model.GotoPgcsRcmd:
			main = seasonm
			op = cardSet[r.ID]
		case model.GotoUpRcmdS:
			op.Limit = upAvCount[r.ID]
		}
		_ = h.From(main, op)
		if h.Get() == nil {
			continue
		}
		h.Get().FromType = card.FromType
		if h.Get().Right {
			switch card.FromType {
			case _fTypeOperation:
				if abtest != nil && abtest.ThreePoint == _newThreePoint {
					h.Get().ThreePointWatchLaterV3()
				} else {
					h.Get().ThreePointWatchLater()
				}
			case _fTypeRecommend:
				if abtest != nil && abtest.ThreePoint == _newThreePoint {
					h.Get().ThreePointChannelV3()
				} else {
					h.Get().ThreePointChannel()
				}
			}
			switch r.Goto {
			case model.GotoTopstick:
				top = h
			default:
				is = append(is, h)
			}
		}
		// infoc
		tinfo := &feed.Item{
			Goto:     card.Type,
			Param:    strconv.FormatInt(card.Value, 10),
			URI:      h.Get().URI,
			FromType: card.FromType,
		}
		infocs = append(infocs, tinfo)
	}
	rl := len(is)
	if rl == 0 {
		is = []cardm.Handler{}
		return
	}
	if idx == 0 {
		idx = now.Unix()
	}
	for i, h := range is {
		if pull {
			h.Get().Idx = idx + int64(rl-i)
		} else {
			h.Get().Idx = idx - int64(i+1)
		}
	}
	return
}

// isShowOperationCards is show operation cards by buvid
func (s *Service) isShowOperationCards(c context.Context, buvid string, channelID int64, cards []*card.Card, now time.Time) (isShow bool) {
	var (
		md5, mcmd5 string
	)
	g, ctx := errgroup.WithContext(c)
	g.Go(func() (err error) {
		if mcmd5, err = s.cd.ChannelCardCache(ctx, buvid, channelID); err != nil {
			isShow = true
			return
		}
		return nil
	})
	g.Go(func() (err error) {
		md5 = s.hashCards(cards)
		return nil
	})
	_ = g.Wait()
	if md5 != mcmd5 {
		isShow = true
		_ = s.cd.AddChannelCardCache(c, buvid, md5, channelID, now)
	}
	return
}

func (s *Service) hashCards(v []*card.Card) string {
	bs, err := json.Marshal(v)
	if err != nil {
		log.Error("json.Marshal error(%v)", err)
		return ""
	}
	return strconv.FormatUint(farm.Hash64(bs), 10)
}
