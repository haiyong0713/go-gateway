package common

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"sync"

	"go-common/library/ecode"
	"go-common/library/log"
	"go-common/library/sync/errgroup.v2"

	"go-gateway/app/app-svr/app-feed/admin/dao/dynamic"
	"go-gateway/app/app-svr/app-feed/admin/model/article"
	cardModel "go-gateway/app/app-svr/app-feed/admin/model/card"
	"go-gateway/app/app-svr/app-feed/admin/model/common"
	"go-gateway/app/app-svr/app-feed/admin/model/live"
	"go-gateway/app/app-svr/app-feed/admin/model/show"
	showModel "go-gateway/app/app-svr/app-feed/admin/model/show"
	"go-gateway/app/app-svr/app-feed/admin/util"
	xecode "go-gateway/app/app-svr/app-feed/ecode"
	"go-gateway/pkg/idsafe/bvid"

	account "git.bilibili.co/bapis/bapis-go/account/service"
	api "git.bilibili.co/bapis/bapis-go/archive/service"

	epgrpc "git.bilibili.co/bapis/bapis-go/pgc/service/season/episode"
	seasondao "git.bilibili.co/bapis/bapis-go/pgc/service/season/season"
)

const (

// errorCommon = "查询错误"
)

// CardPreview card preview
//
//nolint:gocognit
func (s *Service) CardPreview(c context.Context, cType string, id int64) (title string, raw interface{}, err error) {
	var (
		accCard    *account.Card
		appActive  *showModel.AppActive
		eventTopic *showModel.EventTopic
		webCard    *showModel.SearchWebCard
		webRcmd    *showModel.WebRcmdCard
		seaCards   map[int32]*seasondao.CardInfoProto
		seaCard    *seasondao.CardInfoProto
		arcCard    *api.Arc
		article    *article.Article
		pgcEP      map[int32]*epgrpc.EpisodeCardsProto
		dynamics   map[int64]*dynamic.Dynamic
		tmpDynamic *dynamic.Dynamic
		show       map[int64]*show.Shopping
		ok         bool
	)
	switch cType {
	case common.CardPgc:
		v := []int32{int32(id)}
		if seaCards, err = s.pgcDao.CardsInfoReply(c, v); err != nil {
			return
		}
		if v, ok := seaCards[int32(id)]; ok {
			return v.Title, v, nil
		}
		return "", nil, fmt.Errorf("错误类型（无效ID %d）"+util.ErrorPersonFmt, id, s.feedUser.Pgc)
	case common.CardPgcFan:
		if seaCards, err = s.pgcDao.CardsInfoReply(c, []int32{int32(id)}); err != nil {
			return
		}
		if seaCard, ok = seaCards[int32(id)]; !ok {
			return "", nil, fmt.Errorf("错误类型（无效ID %d）"+util.ErrorPersonFmt, id, s.feedUser.Pgc)
		}
		if !util.IsIntInArray([]int{common.PgcSeasonTypeFollowJP, common.PgcSeasonTypeFollowCN}, int(seaCard.SeasonType)) {
			return "", nil, fmt.Errorf("无效ID(%d)，不是番剧/国创", id)
		}
		return seaCard.Title, seaCard, nil
	case common.CardPgcMovie:
		if seaCards, err = s.pgcDao.CardsInfoReply(c, []int32{int32(id)}); err != nil {
			return
		}
		if seaCard, ok = seaCards[int32(id)]; !ok {
			return "", nil, fmt.Errorf("错误类型（无效ID %d）"+util.ErrorPersonFmt, id, s.feedUser.Pgc)
		}
		if !util.IsIntInArray([]int{
			common.PgcSeasonTypeMovie,
			common.PgcSeasonTypeDoc,
			common.PgcSeasonTypeTV,
			common.PgcSeasonTypeShow,
		}, int(seaCard.SeasonType)) {
			return "", nil, fmt.Errorf("无效ID(%d)，不是电影/纪录片/电视剧/综艺", id)
		}
		return seaCard.Title, seaCard, nil
	case common.CardAv:
		if arcCard, err = s.arcDao.Arc(c, id); err != nil {
			if ecode.EqualError(ecode.NothingFound, err) {
				return "", nil, fmt.Errorf("错误类型（无效ID %d）"+util.ErrorPersonFmt, id, s.feedUser.Archive)
			}
			return "", nil, fmt.Errorf(util.ErrorRpcFmts, err.Error(), s.feedUser.Archive, "ArchiveClient.Arc")
		}
		return arcCard.Title, arcCard, nil
	case common.CardUp:
		if accCard, err = s.accDao.Card3(c, id); err != nil {
			if ecode.EqualError(xecode.MemberNotExist, err) {
				return "", nil, fmt.Errorf("错误类型（无效ID %d）"+util.ErrorPersonFmt, id, s.feedUser.Account)
			}
			return "", nil, fmt.Errorf(util.ErrorRpcFmts, err.Error(), s.feedUser.Account, "rpc.client.Card3")
		}
		return accCard.Name, accCard, nil
	case common.CardUpStat:
		upInfo, err := s.accDao.ProfileWithStat3(c, id)
		if err != nil || upInfo == nil {
			log.Error("accDao.ProfileWithStat3 id(%v) error(%v)", id, err)
			return "", nil, err
		}
		profile := upInfo.GetProfile()
		if profile == nil {
			log.Error("upInfo.GetProfile upInfo(%+v) failed", upInfo)
			return "", nil, fmt.Errorf("错误类型 （无效用户mid %d)", id)
		}
		return profile.Name, upInfo, nil
	case common.CardChannelTab:
		if appActive, err = s.showDao.AAFindByID(c, int64(id)); err != nil {
			return "", nil, fmt.Errorf(util.ErrorDBFmts, err.Error(), s.feedUser.Feed)
		}
		if appActive == nil {
			return "", nil, fmt.Errorf("错误类型（无效ID %d）"+util.ErrorPersonFmt, id, s.feedUser.Feed)
		}
		return appActive.Name, appActive, nil
	case common.CardEventTopic:
		if eventTopic, err = s.showDao.ETFindByID(id); err != nil {
			return "", nil, fmt.Errorf(util.ErrorDBFmts, err.Error(), s.feedUser.Feed)
		}
		if eventTopic == nil {
			return "", nil, fmt.Errorf("错误类型（无效ID %d）"+util.ErrorPersonFmt, id, s.feedUser.Feed)
		}
		return eventTopic.Title, eventTopic, nil
	case common.CardSearchWeb:
		if webCard, err = s.showDao.SWBFindByID(id); err != nil {
			return "", nil, fmt.Errorf(util.ErrorDBFmts, err.Error(), s.feedUser.Feed)
		}
		if webCard == nil {
			return "", nil, fmt.Errorf("错误类型（无效ID %d）"+util.ErrorPersonFmt, id, s.feedUser.Feed)
		}
		return webCard.Title, webCard, nil
	case common.CardWebRcmdSpecial:
		if webRcmd, err = s.showDao.WebRcmdCardFindByID(id); err != nil {
			return "", nil, fmt.Errorf(util.ErrorDBFmts, err.Error(), s.feedUser.Feed)
		}
		if webRcmd == nil {
			return "", nil, fmt.Errorf("错误类型（无效ID %d）"+util.ErrorPersonFmt, id, s.feedUser.Feed)
		}
		return webRcmd.Title, webRcmd, nil
	case common.CardWebRcmdGame:
		if title, err = s.GameDao.WebRcmdGame(ctx, id); err != nil {
			return
		}
		if title == "" {
			return "", nil, fmt.Errorf("错误类型（无效ID %d）"+util.ErrorPersonFmt, id, s.feedUser.Game)
		}
		return title, title, nil
		/*
			case common.CardSearchWebGame:
				if title, err = s.GameDao.SearchGame(ctx, id); err != nil {
					return
				}
				if title == "" {
					return "", nil, fmt.Errorf("错误类型（无效ID %d）"+util.ErrorPersonFmt, id, s.feedUser.Game)
				}
				return title, title, nil

		*/
	case common.CardComic:
		if title, err = s.comic.ComicTitle(ctx, id); err != nil {
			return
		}
		if title == "" {
			return "", nil, fmt.Errorf("错误类型（无效ID %d）"+util.ErrorPersonFmt, id, s.feedUser.Comic)
		}
		return title, title, nil
	case common.CardArticle:
		if article, err = s.articleDao.Article(ctx, []int64{id}); err != nil {
			if err == ecode.NothingFound {
				err = fmt.Errorf("错误类型（无效ID %d）"+util.ErrorPersonFmt, id, s.feedUser.Article)
				return
			}
			return
		}
		if article == nil {
			return "", nil, fmt.Errorf("错误类型（无效ID %d）"+util.ErrorPersonFmt, id, s.feedUser.Article)
		}
		return article.Data.Title, article.Data, nil
	case common.CardPgcEP:
		var (
			ep *epgrpc.EpisodeCardsProto
			ok bool
		)
		id32 := int32(id)
		if pgcEP, err = s.pgcDao.CardsEpInfoReply(ctx, []int32{id32}); err != nil {
			return
		}
		if ep, ok = pgcEP[id32]; !ok {
			return "", nil, fmt.Errorf("错误类型（无效ID %d)"+util.ErrorPersonFmt, id32, s.feedUser.Pgc)
		}
		return ep.Title, ep, nil
	case common.CardLive:
		var (
			rooms map[int64]*live.Room
		)
		if rooms, err = s.liveDao.LiveRoom(c, []int64{id}); err != nil {
			return
		}
		if len(rooms) == 0 {
			err = fmt.Errorf("错误类型（无效ID %d） "+util.ErrorPersonFmt, id, s.feedUser.Live)
			return
		}
		for _, v := range rooms {
			return v.Title, v, nil
		}
		if _, ok := rooms[id]; ok {
			return rooms[id].Title, rooms[id], nil
		} else {
			for _, v := range rooms {
				return v.Title, v, nil
			}
		}

		err = fmt.Errorf("错误类型（无效ID %d） "+util.ErrorPersonFmt, id, s.feedUser.Live)
		return

	case common.CardDynamic:
		dynamics, err = s.dynamic.DynamicDetail(c, []int64{id})
		if err != nil {
			return
		}
		if len(dynamics) == 0 {
			err = fmt.Errorf("错误类型（无效ID %d） "+util.ErrorPersonFmt, id, s.feedUser.Dynamic)
			return
		}
		if tmpDynamic, ok = dynamics[id]; !ok {
			err = fmt.Errorf("错误类型（无效ID %d） "+util.ErrorPersonFmt, id, s.feedUser.Dynamic)
			return
		}
		if tmpDynamic.DeleteStatus != dynamic.DynamicNotDel {
			err = fmt.Errorf("错误类型（无效ID %d 动态已删除！）"+util.ErrorPersonFmt, id, s.feedUser.Dynamic)
			return
		}
		if tmpDynamic.AuditStatus != dynamic.DynamicAuditPass {
			err = fmt.Errorf("id错误（无效ID %d 动态已删除！）"+util.ErrorPersonFmt, id, s.feedUser.Dynamic)
			return
		}
		return dynamics[id].DynamicText, dynamics[id], nil
	case common.CardGoods:
		return "", nil, nil
	case common.CardShow:
		show, err = s.vip.Show(c, []int64{id})
		if err != nil {
			return
		}
		if _, ok := show[id]; !ok {
			err = fmt.Errorf("错误类型（无效ID %d）", id)
			return
		}
		return show[id].Name, show[id], nil
	case common.CardAppGame:
		game, err := s.AppGameInfo(c, id)
		if err != nil {
			return "", nil, err
		}
		return game.Name, game, nil
	case common.CardTopic:
		return "", nil, nil
	case common.CardSearchArchive:
		searchArchive, err := s.SearchArchiveCheck(c, id)
		if err != nil {
			return "", nil, err
		}
		return searchArchive.Title, searchArchive, nil
	case common.CardNavigation:
		rCard, err := s.cardDao.ResourceCardQuery(c, id, cardModel.CardTypeNavigation)
		if err != nil {
			return "", nil, err
		}
		return rCard.Title, rCard, nil
	case common.CardOgvInline:
		id32 := int32(id)
		ogvInlines, err := s.pgcDao.InlineCardEpInfoReply(c, []int32{id32})
		if err != nil {
			return "", nil, err
		}
		if _, ok := ogvInlines[id32]; !ok {
			err = fmt.Errorf("错误类型（无效ID %d）", id32)
			return "", nil, err
		}
		return ogvInlines[id32].ShowTitle, ogvInlines[id32], nil
	case common.CardSearchContent:
		rCard, err := s.cardDao.ResourceCardQuery(c, id, cardModel.CardTypeContent)
		if err != nil {
			return "", nil, err
		}
		contentCard, err := cardModel.ParseContentCard(rCard)
		if err != nil {
			return rCard.Title, nil, err
		}
		return contentCard.Title, contentCard, nil
	case common.CardEntryGame, common.CardSearchWebGame:
		gCard, err := s.GameDao.GameEntryInfo(c, id)
		if err != nil {
			return "", nil, err
		}
		return gCard.Name, gCard, nil
	case common.CardMediaList:
		mCard, err := s.mediaListDao.MediaListInfo(c, id)
		if err != nil {
			return "", nil, err
		}
		return mCard.Name, mCard, nil
	default:
		err = fmt.Errorf("参数错误")
		return "", nil, err
	}
}

// CardPreviewBatch card preview batch
func (s *Service) CardPreviewBatch(c context.Context, cType string, reqIds []string) (ret map[string]*common.CardPreview, err error) {
	var (
		reqIdMap = make(map[int64]string)
		int64Ids = make([]int64, 0, len(reqIds))
		int32Ids = make([]int32, 0, len(reqIds))

		eg   = errgroup.WithContext(c)
		lock sync.Mutex
	)
	ret = make(map[string]*common.CardPreview, len(reqIds))

	for _, reqId := range reqIds {
		var id int64
		if strings.HasPrefix(reqId, "bv") {
			err = ecode.Errorf(ecode.RequestErr, "bv id 需要全部大写")
			return
		}

		if strings.HasPrefix(reqId, "BV") {
			if id, err = bvid.BvToAv(reqId); err != nil {
				return
			}
		} else {
			if id, err = strconv.ParseInt(reqId, 10, 64); err != nil {
				return
			}
		}

		reqIdMap[id] = reqId
		int64Ids = append(int64Ids, id)
		int32Ids = append(int32Ids, int32(id))
	}

	switch cType {
	case common.CardSearchWeb:
		cards, err := s.showDao.SWBFindByIDs(int64Ids)
		if err != nil {
			return nil, err
		}
		for _, card := range cards {
			id := strconv.Itoa(int(card.ID))
			cp := &common.CardPreview{Title: card.Title, Raw: card}
			ret[id] = cp
		}
	case common.CardSearchWebGame:
		for _, id := range int64Ids {
			var gameId = id
			eg.Go(func(ctx context.Context) (egErr error) {
				gCard, err := s.GameDao.GameEntryInfo(c, gameId)
				if err != nil {
					log.Warn("CardPreviewBatch GameDao.GameEntryInfo id(%v) err(%v)", gameId, err)
					return nil
				}
				idStr, ok := reqIdMap[gameId]
				if !ok {
					log.Warn("CardPreviewBatch failed to find req id(%v)", gameId)
					return nil
				}
				lock.Lock()
				ret[idStr] = &common.CardPreview{Title: gCard.Name, Raw: gCard}
				lock.Unlock()
				return
			})
		}
	case common.CardSearchArchive:
		for _, id := range int64Ids {
			var avid = id
			eg.Go(func(ctx context.Context) (egErr error) {
				av, err := s.SearchArchiveCheck(ctx, avid)
				if err != nil || av == nil {
					log.Warn("CardPreviewBatch s.SearchArchiveCheck id(%v) err(%v)", avid, err)
					return nil
				}
				idStr, ok := reqIdMap[avid]
				if !ok {
					log.Warn("CardPreviewBatch failed to find req id(%v)", avid)
					return nil
				}
				lock.Lock()
				ret[idStr] = &common.CardPreview{Title: av.Title, Raw: av}
				lock.Unlock()
				return
			})
		}
	case common.CardUpStat:
	case common.CardPgcFan:
	case common.CardPgcMovie:
	default:
		err = fmt.Errorf("参数错误")
		return
	}
	if err = eg.Wait(); err != nil {
		log.Error("CardPreviewBatch eg wait error(%v)", err)
		err = nil
	}
	return
}
