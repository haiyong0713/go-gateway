package service

import (
	"context"
	"encoding/json"
	"strconv"
	"strings"

	"go-common/library/ecode"
	"go-common/library/log"
	"go-common/library/net/metadata"
	"go-common/library/sync/errgroup.v2"
	"go-gateway/app/web-svr/web/interface/model"
	"go-gateway/app/web-svr/web/interface/model/search"
	"go-gateway/pkg/idsafe/bvid"

	arcmdl "go-gateway/app/app-svr/archive/service/api"

	accgrpc "git.bilibili.co/bapis/bapis-go/account/service"
	accmdl "git.bilibili.co/bapis/bapis-go/account/service"
	relmdl "git.bilibili.co/bapis/bapis-go/account/service/relation"
	locationgrpc "git.bilibili.co/bapis/bapis-go/community/service/location"
	watchedGRPC "git.bilibili.co/bapis/bapis-go/live/watched/v1"

	esportConfGRPC "git.bilibili.co/bapis/bapis-go/ai/search/mgr/interface"
	esportGRPC "git.bilibili.co/bapis/bapis-go/operational/esportsservice"
	seasonmdl "git.bilibili.co/bapis/bapis-go/pgc/service/card/search/v1"
)

const (
	_searchEggWebPlat  = 6
	_searchPGCCardPlat = "web"
	_actGame           = 2
	_activity          = "activity"
	_article           = "article"
	_biliUser          = "bili_user"
	_card              = "card"
	_comic             = "comic"
	_liveRoom          = "live_room"
	_mediaBangumi      = "media_bangumi"
	_mediaFt           = "media_ft"
	_operationCard     = "operation_card"
	_special           = "special"
	_star              = "star"
	_tag               = "tag"
	_topic             = "topic"
	_tv                = "tv"
	_twitter           = "twitter"
	_user              = "user"
	_video             = "video"
	_webGame           = "web_game"
	_tips              = "tips"
	//_living            = 7
	_esports = "esports"
)

var _emptyUpRec = make([]*search.UpRecInfo, 0)

// SearchAll search type all.
func (s *Service) SearchAll(c context.Context, mid int64, arg *search.SearchAllArg, buvid, ua, typ string) (data *search.Search, err error) {
	data, err = s.dao.SearchAll(c, mid, arg, buvid, ua, typ)
	if data != nil && data.Result != nil {
		var seps []*seasonmdl.SeasonEpReq
		for _, v := range data.Result.MediaBangumi {
			if v.SeasonID > 0 {
				seps = append(seps, &seasonmdl.SeasonEpReq{SeasonId: v.SeasonID, EpIds: splitEpids(v.HitEpids)})
			}
		}
		for _, v := range data.Result.MediaFt {
			if v.SeasonID > 0 {
				seps = append(seps, &seasonmdl.SeasonEpReq{SeasonId: v.SeasonID, EpIds: splitEpids(v.HitEpids)})
			}
		}
		if len(seps) > 0 {
			pgcReq := &seasonmdl.SearchCardReq{
				Seps:  seps,
				Query: arg.Keyword,
				User: &seasonmdl.UserReq{
					Platform: _searchPGCCardPlat,
					Ip:       metadata.String(c, metadata.RemoteIP),
				}}
			if mid > 0 {
				pgcReq.User.Mid = mid
			}
			if seasonReply, e := s.pgcSearchGRPC.Card(c, pgcReq); e != nil {
				log.Error("SearchAll s.seasonGRPC.SearchCard Seps(%+v) error(%v)", pgcReq, e)
			} else if seasonReply != nil {
				for _, v := range data.Result.MediaBangumi {
					if card, ok := seasonReply.Cards[v.SeasonID]; ok && card != nil {
						v.Fill(card)
					}
				}
				for _, v := range data.Result.MediaFt {
					if card, ok := seasonReply.Cards[v.SeasonID]; ok && card != nil {
						v.Fill(card)
					}
				}
			}
		}

		if arg.Platform == "wechat" {
			s.searchFilterBindOid(c, &data.Result.Video)
		}

		s.handleVideoInfo(&data.Result.Video)
	}
	return
}

func (s *Service) handleVideoInfo(videoSlice *[]*search.SearchVideo) {
	for _, v := range *videoSlice {
		v.Bvid = s.avToBv(v.Aid)
		if len(v.NewRecTags) == 0 {
			v.NewRecTags = make([]*search.SearchNewRecTag, 0)
		}
	}
}

func inIntSlice(itmes []int64, item int64) bool {
	for _, e := range itmes {
		if e == item {
			return true
		}
	}
	return false
}

// nolint: gocognit
func (s *Service) SearchAllV2(ctx context.Context, mid int64, arg *search.SearchAllArg, buvid, ua, typ string) (*search.SearchAll, error) {
	data, err := s.dao.SearchAll(ctx, mid, arg, buvid, ua, typ)
	if err != nil {
		log.Error("SearchAll s.dao.SearchAll(%+v) error(%v)", arg, err)
		return nil, err
	}
	res := &search.SearchAll{
		SearchAllCommon: &search.SearchAllCommon{
			Code:           data.Code,
			SeID:           data.SeID,
			Page:           data.Page,
			PageSize:       data.PageSize,
			Total:          data.Total,
			NumResults:     data.NumResults,
			NumPages:       data.NumPages,
			SuggestKeyword: data.SuggestKeyword,
			RqtType:        data.RqtType,
			CostTime:       data.CostTime,
			ExpList:        data.ExpList,
			EggHit:         data.EggHit,
			PageInfo:       data.PageInfo,
			TopTList:       data.TopTList,
			EggInfo:        data.EggInfo,
			ShowColumn:     data.ShowColumn,
			ShowModuleList: data.ShowModuleList,
		},
		Result: []*search.ResultInfo{},
	}
	res.InBlackKey = s.B2i(s.dao.SearchKeyInBlack(arg.Keyword))
	res.InWhiteKey = s.B2i(s.dao.SearchKeyEqualWhite(arg.Keyword))
	if data == nil || data.Result == nil {
		log.Warn("SearchAllV2 data(%+v) nil", data)
		return res, nil
	}
	var (
		seps                               []*seasonmdl.SeasonEpReq
		contestIDs, matchIDs, mids, arcIDs []int64
		isPowerUp                          map[int64]struct{}
		faceNftMap                         map[int64]int32
		isSeniorMap                        map[int64]int32
	)
	for _, v := range data.Result.MediaBangumi {
		if v.SeasonID > 0 {
			seps = append(seps, &seasonmdl.SeasonEpReq{SeasonId: v.SeasonID, EpIds: splitEpids(v.HitEpids)})
		}
	}
	for _, v := range data.Result.MediaFt {
		if v.SeasonID > 0 {
			seps = append(seps, &seasonmdl.SeasonEpReq{SeasonId: v.SeasonID, EpIds: splitEpids(v.HitEpids)})
		}
	}
	for _, v := range data.Result.Video {
		v.Bvid = s.avToBv(v.Aid)
		if len(v.NewRecTags) == 0 {
			v.NewRecTags = make([]*search.SearchNewRecTag, 0)
		}
		arcIDs = append(arcIDs, v.Aid)
	}
	for _, item := range data.Result.Card {
		for _, v := range item.VideoList {
			if v != nil {
				v.Bvid = s.avToBv(v.Aid)
			}
		}
	}
	for _, item := range data.Result.BiliUser {
		if item == nil {
			continue
		}
		mids = append(mids, item.Mid)
		for _, v := range item.Res {
			if v == nil {
				continue
			}
			v.Bvid = s.avToBv(v.Aid)
			arcIDs = append(arcIDs, v.Aid)
		}
	}
	for _, item := range data.Result.User {
		mids = append(mids, item.Mid)
		for _, v := range item.Res {
			if v == nil {
				continue
			}
			v.Bvid = s.avToBv(v.Aid)
			arcIDs = append(arcIDs, v.Aid)
		}
	}
	for _, eSport := range data.Result.Esports {
		if eSport == nil {
			continue
		}
		if eSport.ID != 0 {
			contestIDs = append(contestIDs, eSport.ID)
		}
		for _, match := range eSport.MatchList {
			if match == nil || match.ID == 0 {
				continue
			}
			matchIDs = append(matchIDs, match.ID)
		}
	}
	var (
		gameData     []*model.SearchGameCard
		tipData      []*model.SearchTipCard
		eSportConfig map[int64]*esportConfGRPC.EsportConfigInfo
		matches      map[int64]*esportGRPC.ContestDetail
		archives     map[int64]*arcmdl.Arc
	)
	group := errgroup.WithContext(ctx)
	if len(seps) > 0 {
		group.Go(func(ctx context.Context) error {
			pgcReq := &seasonmdl.SearchCardReq{
				Seps:  seps,
				Query: arg.Keyword,
				User: &seasonmdl.UserReq{
					Platform: _searchPGCCardPlat,
					Ip:       metadata.String(ctx, metadata.RemoteIP),
				}}
			if mid > 0 {
				pgcReq.User.Mid = mid
			}
			seasonReply, err := s.pgcSearchGRPC.Card(ctx, pgcReq)
			if err != nil {
				log.Error("SearchAllV2 pgcSearchGRPC.Card req=%+v,error=%v", pgcReq, err)
				return nil
			}
			if seasonReply == nil {
				return nil
			}
			for _, v := range data.Result.MediaBangumi {
				if card, ok := seasonReply.Cards[v.SeasonID]; ok && card != nil {
					v.Fill(card)
				}
			}
			for _, v := range data.Result.MediaFt {
				if card, ok := seasonReply.Cards[v.SeasonID]; ok && card != nil {
					v.Fill(card)
				}
			}
			return nil
		})
	}
	group.Go(func(ctx context.Context) error {
		var gameBaseID int64
		for _, tp := range data.ShowModuleList {
			if tp == _webGame {
				for _, game := range data.Result.WebGame {
					if game.CardType == _actGame {
						if gameBaseID, err = strconv.ParseInt(game.CardValue, 10, 64); err != nil {
							log.Error("SearchAllV2 strconv.ParseInt(%s) error(%v)", game.CardValue, err)
							err = nil
							return nil
						}
					}
				}
				break
			}
		}
		if gameBaseID <= 0 {
			return nil
		}
		reply, err := s.dao.SearchGameInfo(ctx, gameBaseID)
		if err != nil {
			log.Error("SearchAllV2 s.dao.SearchGameInfo(%d) error(%v)", gameBaseID, err)
			return nil
		}
		if reply == nil {
			return nil
		}
		gameData = []*model.SearchGameCard{reply}
		return nil
	})
	group.Go(func(ctx context.Context) error {
		for _, val := range data.Result.Tips {
			if val == nil || val.Value == nil {
				continue
			}
			detail, ok := s.searchTipDetailCache[val.Value.ID]
			if !ok {
				continue
			}
			tip := &model.SearchTipCard{}
			tip.FromTip(detail)
			tipData = append(tipData, tip)
		}
		return nil
	})
	group.Go(func(ctx context.Context) error {
		if len(contestIDs) == 0 {
			return nil
		}
		var err error
		if eSportConfig, err = s.matchDao.GetEsportConfigs(ctx, contestIDs); err != nil {
			log.Error("%+v", err)
		}
		return nil
	})
	group.Go(func(ctx context.Context) error {
		if len(matchIDs) == 0 {
			return nil
		}
		if matches, err = s.matchDao.LiveContests(ctx, mid, matchIDs); err != nil {
			log.Error("%+v", err)
		}
		return nil
	})
	group.Go(func(ctx context.Context) error {
		if len(mids) == 0 {
			return nil
		}
		reply, err := s.accGRPC.ProfileWithoutPrivacy3(ctx, &accgrpc.MidReq{Mid: mids[0]})
		if err != nil {
			log.Error("%+v", err)
			return nil
		}
		profile := reply.GetProfileWithoutPrivacy()
		if profile == nil {
			return nil
		}
		isPowerUp = make(map[int64]struct{})
		if profile.IsLatest_100Honour == 1 {
			isPowerUp[mids[0]] = struct{}{}
		}
		faceNftMap = make(map[int64]int32)
		isSeniorMap = make(map[int64]int32)
		faceNftMap[mids[0]] = profile.FaceNftNew
		isSeniorMap[mids[0]] = profile.IsSeniorMember
		return nil
	})
	group.Go(func(ctx context.Context) error {
		if len(arcIDs) == 0 {
			return nil
		}
		reply, err := s.batchArchives(ctx, arcIDs)
		if err != nil {
			log.Error("%+v", err)
		}
		archives = reply
		return nil
	})
	if err := group.Wait(); err != nil {
		log.Error("%+v", err)
	}
	midNFTRegionMap := s.BatchNFTRegion(ctx, mids)
	var (
		eSportsData  []*model.SearchESportsCard
		biliUserData []*model.SearchBiliUserCard
		userData     []*model.SearchUserCard
	)
	for _, eSport := range data.Result.Esports {
		card := &model.SearchESportsCard{}
		if val, ok := eSportConfig[eSport.ID]; ok {
			card.ConfigInfo = val
		}
		for _, match := range eSport.MatchList {
			if val, ok := matches[match.ID]; ok {
				card.Contest = append(card.Contest, val)
			}
		}
		if len(card.Contest) == 0 {
			continue
		}
		eSportsData = append(eSportsData, card)
	}
	for _, biliUser := range data.Result.BiliUser {
		_, isPowerUp := isPowerUp[biliUser.Mid]
		biliUser.FaceNft = faceNftMap[biliUser.Mid]
		biliUser.FaceNftType = midNFTRegionMap[biliUser.Mid]
		biliUser.IsSeniorMember = isSeniorMap[biliUser.Mid]
		for _, v := range biliUser.Res {
			v.Fill(archives[v.Aid])
		}
		card := &model.SearchBiliUserCard{
			SearchUser: biliUser,
			Expand: &model.SearchBiliUserCardExpand{
				IsPowerUp:    isPowerUp,
				SystemNotice: s.systemNoticeCache[biliUser.Mid],
			},
		}
		biliUserData = append(biliUserData, card)
	}
	for _, user := range data.Result.User {
		_, isPowerUp := isPowerUp[user.Mid]
		user.FaceNft = faceNftMap[user.Mid]
		user.FaceNftType = midNFTRegionMap[user.Mid]
		user.IsSeniorMember = isSeniorMap[user.Mid]
		for _, v := range user.Res {
			v.Fill(archives[v.Aid])
		}
		card := &model.SearchUserCard{
			SearchUser: user,
			Expand: &model.SearchBiliUserCardExpand{
				IsPowerUp:    isPowerUp,
				SystemNotice: s.systemNoticeCache[user.Mid],
			},
		}
		userData = append(userData, card)
	}
	// 获取稿件数据
	for _, video := range data.Result.Video {
		video.Fill(archives[video.Aid])
	}
	if len(data.ShowModuleList) > 0 {
		var resultList []*search.ResultInfo
		for _, keyName := range data.ShowModuleList {
			switch keyName {
			case _esports:
				if len(eSportsData) == 0 {
					eSportsData = []*model.SearchESportsCard{}
				}
				resultList = append(resultList, &search.ResultInfo{ResultType: keyName, Data: eSportsData})
			case _webGame:
				if len(gameData) == 0 {
					gameData = []*model.SearchGameCard{}
				}
				resultList = append(resultList, &search.ResultInfo{ResultType: keyName, Data: gameData})
			case _tips:
				if len(tipData) == 0 {
					tipData = []*model.SearchTipCard{}
				}
				resultList = append(resultList, &search.ResultInfo{ResultType: keyName, Data: tipData})
			case _biliUser:
				if len(biliUserData) == 0 {
					biliUserData = []*model.SearchBiliUserCard{}
				}
				resultList = append(resultList, &search.ResultInfo{ResultType: keyName, Data: biliUserData})
			case _user:
				if len(userData) == 0 {
					userData = []*model.SearchUserCard{}
				}
				resultList = append(resultList, &search.ResultInfo{ResultType: keyName, Data: userData})
			default:
				resultList = append(resultList, s.resultFmt(keyName, data.Result))
			}
		}
		if len(resultList) > 0 {
			res.Result = resultList
		}
	}
	return res, nil
}

func (s *Service) resultFmt(tp string, result *search.Result) (rs *search.ResultInfo) {
	rs = &search.ResultInfo{
		ResultType: tp,
	}
	switch tp {
	case _activity:
		rs.Data = result.Activity
	case _article:
		rs.Data = result.Article
	case _card:
		rs.Data = result.Card
	case _comic:
		rs.Data = result.Comic
	case _liveRoom:
		rs.Data = result.LiveRoom
	case _mediaBangumi:
		rs.Data = result.MediaBangumi
	case _mediaFt:
		rs.Data = result.MediaFt
	case _operationCard:
		rs.Data = result.OperationCard
	case _special:
		rs.Data = result.Special
	case _star:
		rs.Data = result.Star
	case _tag:
		rs.Data = result.Tag
	case _topic:
		rs.Data = result.Topic
	case _tv:
		rs.Data = result.Tv
	case _twitter:
		rs.Data = result.Twitter
	case _video:
		rs.Data = result.Video
	case _tips:
		rs.Data = result.Tips
	case _esports:
		rs.Data = result.Esports
	}
	return
}

// SearchByType type video,bangumi,pgc,live,live_user,article,special,topic,bili_user,photo
func (s *Service) SearchByType(c context.Context, mid int64, arg *search.SearchTypeArg, buvid, ua string) (res *search.SearchTypeRes, err error) {
	switch arg.SearchType {
	case search.SearchTypeLive:
		if res, err = s.dao.SearchLive(c, mid, arg, buvid, ua); err != nil {
			return
		}
		if fillResult, err := s.fillLiveRoomWatched(c, mid, res.Result); err == nil {
			res.Result = fillResult
		}
	case search.SearchTypeLiveRoom:
		if res, err = s.dao.SearchLiveRoom(c, mid, arg, buvid, ua); err != nil {
			return
		}
	case search.SearchTypeLiveUser:
		if res, err = s.dao.SearchLiveUser(c, mid, arg, buvid, ua); err != nil {
			return
		}
	case search.SearchTypeArticle:
		if res, err = s.dao.SearchArticle(c, mid, arg, buvid, ua); err != nil {
			return
		}
	case search.SearchTypeSpecial:
		if res, err = s.dao.SearchSpecial(c, mid, arg, buvid, ua); err != nil {
			return
		}
	case search.SearchTypeTopic:
		if res, err = s.dao.SearchTopic(c, mid, arg, buvid, ua); err != nil {
			return
		}
	case search.SearchTypePhoto:
		if res, err = s.dao.SearchPhoto(c, mid, arg, buvid, ua); err != nil {
			return
		}
	default:
		err = ecode.RequestErr
		return
	}
	if res != nil {
		res.InBlackKey = s.B2i(s.dao.SearchKeyInBlack(arg.Keyword))
		res.InWhiteKey = s.B2i(s.dao.SearchKeyEqualWhite(arg.Keyword))
	}
	return
}

// SearchPGC search pgc data.
func (s *Service) SearchPGC(c context.Context, mid int64, arg *search.SearchTypeArg, buvid, ua string) (res *search.SearchPGCRes, err error) {
	switch arg.SearchType {
	case search.SearchTypeBangumi:
		if res, err = s.dao.SearchBangumi(c, mid, arg, buvid, ua); err != nil {
			return
		}
	case search.SearchTypeMovie:
		if res, err = s.dao.SearchMovie(c, mid, arg, buvid, ua); err != nil {
			return
		}
	default:
		err = ecode.RequestErr
		return
	}
	if res != nil {
		res.InBlackKey = s.B2i(s.dao.SearchKeyInBlack(arg.Keyword))
		res.InWhiteKey = s.B2i(s.dao.SearchKeyEqualWhite(arg.Keyword))

		var seps []*seasonmdl.SeasonEpReq
		for _, v := range res.Result {
			if v.SeasonID > 0 {
				seps = append(seps, &seasonmdl.SeasonEpReq{SeasonId: v.SeasonID, EpIds: splitEpids(v.HitEpids)})
			}
		}
		if len(seps) > 0 {
			pgcReq := &seasonmdl.SearchCardReq{
				Seps:  seps,
				Query: arg.Keyword,
				User: &seasonmdl.UserReq{
					Platform: _searchPGCCardPlat,
					Ip:       metadata.String(c, metadata.RemoteIP),
				}}
			if mid > 0 {
				pgcReq.User.Mid = mid
			}
			if seasonReply, e := s.pgcSearchGRPC.Card(c, pgcReq); e != nil {
				log.Error("SearchPGC s.seasonGRPC.SearchCard Seps(%+v) error(%v)", pgcReq, e)
			} else if seasonReply != nil {
				for _, v := range res.Result {
					if card, ok := seasonReply.Cards[v.SeasonID]; ok && card != nil {
						v.Fill(card)
					}
				}
			}
		}
	}
	return
}

func splitEpids(hitEpids string) (data []int32) {
	idStrs := strings.Split(hitEpids, ",")
	for _, v := range idStrs {
		if id, err := strconv.ParseInt(v, 10, 64); err != nil {
			log.Warn("splitEpids strconv.ParseInt(%s) warn(%v)", v, err)
		} else {
			data = append(data, int32(id))
		}
	}
	return
}

// SearchVideo search type video.
func (s *Service) SearchVideo(c context.Context, mid int64, arg *search.SearchTypeArg, buvid, ua string) (*search.SearchVideoRes, error) {
	res, err := s.dao.SearchVideo(c, mid, arg, buvid, ua)
	if err != nil {
		return nil, err
	}
	if res != nil {
		res.InBlackKey = s.B2i(s.dao.SearchKeyInBlack(arg.Keyword))
		res.InWhiteKey = s.B2i(s.dao.SearchKeyEqualWhite(arg.Keyword))
	}
	if res == nil || len(res.Result) == 0 {
		return res, nil
	}

	if arg.Platform == "wechat" {
		s.searchFilterBindOid(c, &res.Result)
	}

	var arcIDs []int64
	for _, v := range res.Result {
		if v != nil {
			v.Bvid = s.avToBv(v.Aid)
			arcIDs = append(arcIDs, v.Aid)
		}
	}
	archives, err := s.batchArchives(c, arcIDs)
	if err != nil {
		log.Error("s.batchArchives error,%+v", err)
		return res, nil
	}
	for _, video := range res.Result {
		video.Fill(archives[video.Aid])
	}
	return res, nil
}

func (s *Service) searchFilterBindOid(c context.Context, containOidSlice *[]*search.SearchVideo) {
	if len(*containOidSlice) == 0 {
		return
	}
	var oidList []int64
	for _, v := range *containOidSlice {
		if v != nil {
			oidList = append(oidList, v.Aid)
		}
	}

	bindOidList, err := s.dao.TagBind(c, oidList)
	k := 0
	for _, v := range *containOidSlice {
		if err != nil || bindOidList == nil || v == nil || !inIntSlice(bindOidList, v.Aid) {
			(*containOidSlice)[k] = v
			k++
		}
	}
	*containOidSlice = (*containOidSlice)[:k]
}

// SearchUser search type user.
func (s *Service) SearchUser(c context.Context, mid int64, arg *search.SearchTypeArg, buvid, ua string) (res *search.SearchUserRes, err error) {

	if res, err = s.dao.SearchUser(c, mid, arg, buvid, ua); err != nil {
		return
	}
	if res != nil {
		res.InBlackKey = s.B2i(s.dao.SearchKeyInBlack(arg.Keyword))
		res.InWhiteKey = s.B2i(s.dao.SearchKeyEqualWhite(arg.Keyword))
	}
	var mids []int64
	for _, item := range res.Result {
		mids = append(mids, item.Mid)
		for _, v := range item.Res {
			if v != nil {
				v.Bvid = s.avToBv(v.Aid)
			}
		}
	}
	if len(mids) > 0 {
		// 获取face_nft
		func() {
			reply, err := s.accGRPC.Cards3(c, &accgrpc.MidsReq{Mids: mids})
			if err != nil {
				log.Error("s.SearchUser Cards3 error:%+v", err)
				return
			}
			for _, item := range res.Result {
				if val, ok := reply.GetCards()[item.Mid]; ok {
					item.FaceNft = val.FaceNftNew
					item.IsSeniorMember = val.IsSeniorMember
				}
			}
		}()
		// 获取face_nft_type
		midNFTRegionMap := s.BatchNFTRegion(c, mids)
		for _, item := range res.Result {
			item.FaceNftType = midNFTRegionMap[item.Mid]
		}
	}
	return
}

// SearchRec search recommend data.
func (s *Service) SearchRec(c context.Context, mid int64, pn, ps int, keyword, fromSource, buvid, ua string) (data *search.SearchRec, err error) {
	data, err = s.dao.SearchRec(c, mid, pn, ps, keyword, fromSource, buvid, ua)
	return
}

// SearchDefault get search default word.
func (s *Service) SearchDefault(c context.Context, mid int64, fromSource, buvid, ua string) (data *search.SearchDefault, err error) {
	data, err = s.dao.SearchDefault(c, mid, fromSource, buvid, ua)
	return
}

// UpRec get up recommend
func (s *Service) UpRec(c context.Context, mid int64, arg *search.SearchUpRecArg) (data *search.UpRecData, err error) {
	var (
		ups        []*search.SearchUpRecRes
		trackID    string
		mids       []int64
		cardsReply *accmdl.CardsReply
		cardErr    error
	)
	if ups, trackID, err = s.dao.UpRecommend(c, mid, arg); err != nil {
		return
	}
	data = &search.UpRecData{TrackID: trackID}
	if len(ups) == 0 {
		data.List = _emptyUpRec
		return
	}
	for _, v := range ups {
		mids = append(mids, v.UpID)
	}
	relInfos := make(map[int64]*relmdl.StatReply, len(mids))
	group := errgroup.WithContext(c)
	group.Go(func(ctx context.Context) error {
		if cardsReply, cardErr = s.accGRPC.Cards3(ctx, &accmdl.MidsReq{Mids: mids}); cardErr != nil {
			log.Error("UpRec s.accGRPC.Cards3(%v) error(%v)", mids, cardErr)
			return cardErr
		}
		return nil
	})
	group.Go(func(ctx context.Context) error {
		if relReply, relErr := s.relationGRPC.Stats(ctx, &relmdl.MidsReq{Mids: mids}); relErr != nil {
			log.Error("UpRec s.relationGRPC.Stats(%d,%v) error(%v)", mid, mids, relErr)
		} else if relReply != nil {
			relInfos = relReply.StatReplyMap
		}
		return nil
	})
	if err = group.Wait(); err != nil {
		return
	}
	for _, v := range ups {
		if info, ok := cardsReply.Cards[v.UpID]; ok && info != nil && info.Silence == 0 {
			upInfo := &search.UpRecInfo{
				Mid:       info.Mid,
				Name:      info.Name,
				Face:      info.Face,
				Official:  info.Official,
				RecReason: v.RecReason,
				Tid:       v.Tid,
				SecondTid: v.SecondTid,
				Sign:      info.Sign,
			}
			upInfo.Vip.Type = info.Vip.Type
			upInfo.Vip.Status = info.Vip.Status
			if stat, ok := relInfos[v.UpID]; ok {
				upInfo.Follower = stat.Follower
			}
			if typ, ok := s.typeNames[int32(v.Tid)]; ok {
				upInfo.Tname = typ.Name
			}
			if typ, ok := s.typeNames[int32(v.SecondTid)]; ok {
				upInfo.SecondTname = typ.Name
			}
			data.List = append(data.List, upInfo)
		}
	}
	if len(data.List) == 0 {
		data.List = _emptyUpRec
	}
	return
}

// SearchEgg get search egg by egg id.
func (s *Service) SearchEgg(c context.Context, eggID int64) (data *search.SearchEggRes, err error) {
	if _, ok := s.searchEggs[eggID]; !ok {
		err = ecode.NothingFound
		return
	}
	data = s.searchEggs[eggID]
	return
}

// SearchGameInfo get search game info.
func (s *Service) SearchGameInfo(c context.Context, gameBaseID int64) (data *model.SearchGameCard, err error) {
	if data, err = s.dao.SearchGameInfo(c, gameBaseID); err != nil {
		log.Error("SearchGameInfo s.dao.SearchGameInfo(%d) error(%v)", gameBaseID, err)
		err = ecode.NothingFound
	}
	return
}

func (s *Service) loadSearchEgg() {
	if s.searchEggRunning {
		return
	}
	s.searchEggRunning = true
	defer func() {
		s.searchEggRunning = false
	}()
	eggs, err := s.dao.SearchEgg(context.Background())
	if err != nil {
		log.Error("s.dao.SearchEgg error(%v)", err)
		return
	}
	data := make(map[int64]*search.SearchEggRes, len(eggs))
	for _, v := range eggs {
		if source, ok := v.Plat[_searchEggWebPlat]; ok {
			for _, egg := range source {
				if _, isSet := data[egg.EggID]; !isSet {
					data[egg.EggID] = &search.SearchEggRes{
						EggID:     egg.EggID,
						ShowCount: v.ShowCount,
					}
				}
				source := &search.SearchEggSource{URL: egg.URL, MD5: egg.MD5, Size: egg.Size}
				data[egg.EggID].Source = append(data[egg.EggID].Source, source)
			}
		}
	}
	s.searchEggs = data
}

func (s *Service) SearchSquare(ctx context.Context, mid int64, buvid string, limit int, isInner int64, platform string) (*search.SquareResult, error) {
	ip := metadata.String(ctx, metadata.RemoteIP)
	res, err := s.locGRPC.Info(ctx, &locationgrpc.InfoReq{Addr: ip})
	if err != nil {
		log.Error("SearchSquare error:%+v", err)
	}
	zoneID := res.GetZoneId()
	trending, err := s.trending(ctx, buvid, mid, limit, isInner, zoneID, platform)
	if err != nil {
		return nil, err
	}
	return &search.SquareResult{
		Trending: trending,
	}, nil
}

func (s *Service) trending(ctx context.Context, buvid string, mid int64, limit int, isInner, zoneID int64, platform string) (*search.SquareTrending, error) {
	reply, err := s.dao.Trending(ctx, buvid, mid, limit, isInner, zoneID, platform)
	if err != nil {
		log.Error("%+v", err)
		return nil, err
	}
	if reply == nil {
		return nil, nil
	}
	res := &search.SquareTrending{
		Title:   "热搜",
		Trackid: reply.SeID,
	}
	for _, val := range reply.List {
		//if val.WordType == _living {
		//	val.Icon = ""
		//}
		item := &search.SquareList{
			Keyword:  val.Keyword,
			ShowName: val.ShowName,
			Icon:     val.Icon,
		}
		param := val.GotoValue
		switch val.GotoType {
		case search.HotTypeArchive:
			item.Goto = model.GotoBv
			avid, err := strconv.ParseInt(param, 10, 64)
			if err != nil {
				log.Error("%+v", err)
				continue
			}
			param, _ = bvid.AvToBv(avid)
		case search.HotTypeArticle:
			item.Goto = model.GotoArticle
		case search.HotTypePGC:
			item.Goto = model.GotoPGCEP
		case search.HotTypeURL:
			item.Goto = model.GotoURL
		default:
		}
		item.URI = model.FillURI(item.Goto, param, nil)
		res.List = append(res.List, item)
	}
	return res, nil
}

func (s *Service) fillLiveRoomWatched(ctx context.Context, mid int64, result json.RawMessage) (json.RawMessage, error) {
	searchRes := struct {
		LiveRoom []*search.SearchLiveRoom `json:"live_room"`
		LiveUser []*search.SearchLiveUser `json:"live_user"`
	}{}
	err := json.Unmarshal(result, &searchRes)
	if err != nil {
		log.Error("s.fillLiveRoomWatched err:%+v", err)
		return nil, err
	}
	var roomIDs []int64
	for _, room := range searchRes.LiveRoom {
		if room == nil {
			continue
		}
		roomIDs = append(roomIDs, room.Roomid)
	}
	if len(roomIDs) == 0 {
		return result, nil
	}
	req := &watchedGRPC.MultiShowReq{
		RoomIds: roomIDs,
		Uid:     mid,
	}
	res, err := s.watchedGRPC.MultiShow(ctx, req)
	if err != nil {
		log.Error("s.fillLiveRoomWatched MultiShow req:%+v, err:%+v", req, err)
		return nil, err
	}
	for _, room := range searchRes.LiveRoom {
		if room == nil {
			continue
		}
		room.WatchedShow = res.GetList()[room.Roomid]
	}
	data, err := json.Marshal(searchRes)
	if err != nil {
		log.Error("s.fillLiveRoomWatched Marshal err:%+v", err)
		return nil, err
	}
	return data, nil
}
