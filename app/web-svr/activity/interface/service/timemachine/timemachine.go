package timemachine

import (
	"context"
	"sort"
	"strconv"
	"strings"
	"time"

	accapi "git.bilibili.co/bapis/bapis-go/account/service"
	"go-common/library/ecode"
	"go-common/library/log"
	arcapi "go-gateway/app/app-svr/archive/service/api"
	"go-gateway/app/web-svr/activity/interface/client"
	"go-gateway/app/web-svr/activity/interface/model/timemachine"

	artmdl "git.bilibili.co/bapis/bapis-go/article/model"
	artapi "git.bilibili.co/bapis/bapis-go/article/service"
	api "git.bilibili.co/bapis/bapis-go/pgc/service/season/season"

	"go-common/library/sync/errgroup.v2"
)

const (
	_tidScoreLen     = 6
	_upCreateTypeArc = 1
	_upCreateTypeArt = 2
)

func (s *Service) Timemachine(c context.Context, loginMid, mid int64) (*timemachine.Result, error) {
	// 查看其它用户数据需要权限
	if mid > 0 && mid != loginMid {
		if _, ok := s.tmMidMap[loginMid]; !ok {
			return nil, ecode.AccessDenied
		}
	}
	if mid == 0 {
		mid = loginMid
	}
	var (
		infos        map[int64]*accapi.Info
		arcs         map[int64]*arcapi.Arc
		arts         map[int64]*artmdl.Meta
		seasons      map[int32]*api.CardInfoProto
		flags        []*timemachine.Flag
		cacheDataNil bool
	)
	cacheData, err := s.dao.CacheTimemachine(c, mid)
	if err != nil || cacheData == nil {
		log.Error("Timemachine s.dao.CacheTimemachine mid(%d) error(%v) or nil cacheData", mid, err)
		cacheDataNil = true
		cacheData = &timemachine.Item{Mid: mid}
	}
	res := &timemachine.Result{Sid: s.c.Timemachine.FlagSid}
	mids, aids, artIDs, seasonIDs := resourceIDs(cacheData)
	group := errgroup.WithContext(c)
	group.Go(func(ctx context.Context) error {
		reply, infoErr := client.AccountClient.Infos3(ctx, &accapi.MidsReq{Mids: mids})
		if infoErr != nil {
			log.Error("Timemachine s.accClient.Infos3(%v) error(%v)", mids, infoErr)
			return nil
		}
		infos = reply.Infos
		return nil
	})
	if len(aids) > 0 {
		group.Go(func(ctx context.Context) error {
			reply, arcErr := client.ArchiveClient.Arcs(ctx, &arcapi.ArcsRequest{Aids: aids})
			if arcErr != nil {
				log.Error("Timemachine s.arcClient.Arcs(%v) error(%v)", aids, arcErr)
				return nil
			}
			arcs = reply.Arcs
			return nil
		})
	}
	if len(artIDs) > 0 {
		group.Go(func(ctx context.Context) error {
			reply, artErr := client.ArticleClient.ArticleMetas(ctx, &artapi.ArticleMetasReq{Ids: artIDs})
			if artErr != nil {
				log.Error("Timemachine s.artClient.ArticleMetas(%v) error(%v)", artIDs, artErr)
				return nil
			}
			arts = reply.Res
			return nil
		})
	}
	if len(seasonIDs) > 0 {
		group.Go(func(ctx context.Context) error {
			reply, seasonErr := client.SeasonClient.Cards(ctx, &api.SeasonInfoReq{SeasonIds: seasonIDs})
			if seasonErr != nil {
				log.Error("Timemachine s.seasonClient.Cards(%v) error(%v)", seasonIDs, seasonErr)
				return nil
			}
			seasons = reply.Cards
			return nil
		})
	}
	group.Go(func(ctx context.Context) error {
		lidItems, likeErr := s.likeDao.LikeActLids(c, s.c.Timemachine.FlagSid, mid)
		if likeErr != nil {
			log.Error("Timemachine s.dao.LikeActLids(%d,%d) error(%v)", s.c.Timemachine.FlagSid, mid, likeErr)
			return nil
		}
		if len(lidItems) == 0 {
			return nil
		}
		var lids []int64
		for _, v := range lidItems {
			if v != nil {
				lids = append(lids, v.Lid)
			}
		}
		contents, likeErr := s.likeDao.LikeContent(c, lids)
		if likeErr != nil {
			log.Error("Timemachine s.likeDao.LikeContent(%v) error(%v)", lids, likeErr)
			return nil
		}
		for _, v := range lidItems {
			if v != nil {
				if item, ok := contents[v.Lid]; ok && item != nil {
					for i := 0; int64(i) < v.Action; i++ {
						flags = append(flags, &timemachine.Flag{Lid: v.Lid, Message: item.Message})
					}
				}
			}
		}
		return nil
	})
	if err = group.Wait(); err != nil {
		log.Error("Timemachine mid(%d) group.Wait error(%v)", mid, err)
	}
	// 无数据只返回第14页
	if cacheDataNil {
		res.PageFourteen = s.pageFourteen(cacheData, flags)
		return res, nil
	}
	res.PageOne = pageOne(cacheData, infos)
	res.PageTwo = s.pageTwo(cacheData)
	res.PageThree = s.pageThree(cacheData)
	res.PageFour = pageFour(cacheData, arcs)
	res.PageFive = pageFive(cacheData, seasons)
	res.PageSix = pageSix(cacheData, seasons)
	res.PageSeven = s.pageSeven(cacheData)
	res.PageEight = pageEight(cacheData, infos, arcs)
	res.PageNine = pageNine(cacheData, infos)
	res.PageTen = pageTen(cacheData, arcs, arts)
	res.PageEleven = pageEleven(cacheData, infos)
	res.PageTwelve = pageTwelve(cacheData)
	res.PageThirteen = pageThirteen(cacheData, infos)
	res.PageFourteen = s.pageFourteen(cacheData, flags)
	return res, nil
}

func resourceIDs(item *timemachine.Item) (mids, aids []int64, artIDs []int64, seasonIDs []int32) {
	mids = []int64{item.Mid}
	if item.LikeBestUp > 0 {
		mids = append(mids, item.LikeBestUp)
	}
	if item.LikeBestLiveUp > 0 {
		mids = append(mids, item.LikeBestLiveUp)
	}
	if item.BestFanMid > 0 {
		mids = append(mids, item.BestFanMid)
	}
	if item.BestLiveFanMid > 0 {
		mids = append(mids, item.BestLiveFanMid)
	}
	if item.CoinAvid > 0 {
		aids = append(aids, item.CoinAvid)
	}
	if item.LikeUpBestCreate > 0 {
		aids = append(aids, item.LikeUpBestCreate)
	}
	if item.BestCreate > 0 {
		switch item.UpBestCreatType {
		case _upCreateTypeArc:
			aids = append(aids, item.BestCreate)
		case _upCreateTypeArt:
			artIDs = append(artIDs, item.BestCreate)
		}
	}
	if item.PlayAmsDuration > 0 && item.BestLikeSid > 0 {
		seasonIDs = append(seasonIDs, item.BestLikeSid)
	}
	if item.BestLikeYinshi > 0 {
		seasonIDs = append(seasonIDs, item.BestLikeYinshi)
	}
	return
}

func pageOne(item *timemachine.Item, infos map[int64]*accapi.Info) *timemachine.PageOne {
	hourMap := make(map[string]int64)
	for i := 1; i <= 6; i++ {
		hourMap[strconv.Itoa(i)] = 0
	}
	if item.HourVisitDays != "" {
		hours := strings.Split(item.HourVisitDays, ",")
		for _, v := range hours {
			visits := strings.Split(v, ":")
			if len(visits) == 2 {
				times, _ := strconv.ParseInt(visits[1], 10, 64)
				hourMap[visits[0]] = times
			}
		}
	}
	info := infos[item.Mid]
	if info == nil {
		log.Warn("pageOne mid(%d) info not found", item.Mid)
		info = &accapi.Info{Mid: item.Mid}
	}
	// 360天以上补一天
	if item.VisitDays >= 360 {
		item.VisitDays = item.VisitDays + 1
	}
	if item.VisitDays > 365 {
		item.VisitDays = 365
	}
	return &timemachine.PageOne{
		Mid:              info.Mid,
		Name:             info.Name,
		Face:             info.Face,
		VisitDays:        item.VisitDays,
		HourVisitDays:    hourMap,
		MaxVisitDaysHour: item.MaxVisitDaysHour,
	}
}

func (s *Service) pageTwo(item *timemachine.Item) *timemachine.PageTwo {
	var tname string
	typ, ok := s.typeNames[item.MaxVvTid]
	if ok && typ != nil {
		tname = typ.Name
	}
	if tname == "" {
		log.Warn("pageTwo mid(%d) MaxVvTid(%d) not found", item.Mid, item.MaxVvTid)
		return nil
	}
	tids := strings.Split(item.Top6VvTidScore, ",")
	tidScores := make([]*timemachine.TidScore, 0)
	for _, v := range tids {
		tidArr := strings.Split(v, ":")
		if len(tidArr) != 2 {
			log.Warn("pageTwo mid(%d) tid score data(%s)", item.Mid, v)
			continue
		}
		tid, err := strconv.ParseInt(tidArr[0], 10, 64)
		if err != nil {
			log.Warn("pageTwo mid(%d) strconv.ParseInt tid(%s) error(%v)", item.Mid, tidArr[0], err)
			continue
		}
		score, err := strconv.ParseInt(tidArr[1], 10, 64)
		if err != nil {
			log.Warn("pageTwo mid(%d) strconv.ParseInt score(%s) error(%v)", item.Mid, tidArr[1], err)
			continue
		}
		secType, ok := s.typeNames[int32(tid)]
		if !ok {
			continue
		}
		item := &timemachine.TidScore{Tid: tid, Tname: secType.Name, Score: score}
		tidScores = append(tidScores, item)
	}
	sort.Slice(tidScores, func(i, j int) bool {
		return tidScores[i].Score > tidScores[j].Score
	})
	tidScores = func(scores []*timemachine.TidScore) []*timemachine.TidScore {
		// 最多展示6个
		if len(scores) >= _tidScoreLen {
			return scores[:_tidScoreLen]
		}
		tidScoreMap := make(map[int64]*timemachine.TidScore)
		for _, v := range scores {
			tidScoreMap[v.Tid] = v
		}
		for _, v := range timemachine.DefaultTidScores {
			if _, ok := tidScoreMap[v.Tid]; !ok {
				scores = append(scores, v)
			}
		}
		// 按score排序
		return scores[:_tidScoreLen]
	}(tidScores)
	sort.Slice(tidScores, func(i, j int) bool {
		return tidScores[i].Tid < tidScores[j].Tid
	})
	return &timemachine.PageTwo{
		Vv:             item.Vv,
		MaxVvTid:       item.MaxVvTid,
		MaxVvTname:     tname,
		Top6VvTidScore: tidScores,
	}
}

func (s *Service) pageThree(item *timemachine.Item) *timemachine.PageThree {
	topTagIDs := func() []*timemachine.TagScore {
		tagArr := strings.Split(item.Top10VvTag, ",")
		if len(tagArr) == 0 {
			log.Warn("pageThree mid(%d) Top10VvTag(%s) len(tagArr) == 0 ", item.Mid, item.Top10VvTag)
			return nil
		}
		var tagScores []*timemachine.TagScore
		for _, v := range tagArr {
			tagIDArr := strings.Split(v, ":")
			if len(tagIDArr) != 2 {
				log.Warn("pageThree mid(%d) tag(%s) len(tagIDArr) != 2 ", item.Mid, tagArr[0])
				continue
			}
			tagID, err := strconv.ParseInt(tagIDArr[0], 10, 64)
			if err != nil {
				log.Warn("pageThree mid(%d) strconv.ParseInt tagID(%s) error(%v)", item.Mid, tagIDArr[0], err)
				continue
			}
			score, err := strconv.ParseInt(tagIDArr[1], 10, 64)
			if err != nil {
				log.Warn("pageThree mid(%d) strconv.ParseInt score(%s) error(%v)", item.Mid, tagIDArr[1], err)
				continue
			}
			tagScores = append(tagScores, &timemachine.TagScore{Tid: tagID, Score: score})
		}
		// 按score排序
		sort.Slice(tagScores, func(i, j int) bool {
			return tagScores[i].Score > tagScores[j].Score
		})
		return tagScores
	}()
	var (
		topTag *timemachine.Tag
		ok     bool
	)
	for _, v := range topTagIDs {
		topTag, ok = s.tagDescs[v.Tid]
		if ok && topTag != nil {
			break
		}
	}
	// region value
	if topTag == nil {
		regionDesc, ok := s.regionDescs[item.MaxVvSubtid]
		if ok && regionDesc != nil {
			topTag = &timemachine.Tag{
				Name:    regionDesc.Name,
				DescOne: regionDesc.DescOne,
				DescTwo: regionDesc.DescTwo,
				Pic:     regionDesc.Pic,
			}
		}
	}
	// default tag
	if topTag == nil {
		topTag = &timemachine.Tag{
			Name:    s.dftRegionDesc.Name,
			DescOne: s.dftRegionDesc.DescOne,
			DescTwo: s.dftRegionDesc.DescTwo,
			Pic:     s.dftRegionDesc.Pic,
		}
	}
	return &timemachine.PageThree{
		MaxVvSubtid: item.MaxVvSubtid,
		Top10VvTag:  item.Top10VvTag,
		TagName:     topTag.Name,
		TagDescOne:  topTag.DescOne,
		TagDescTwo:  topTag.DescTwo,
		TagPic:      topTag.Pic,
	}
}

func pageFour(item *timemachine.Item, arcs map[int64]*arcapi.Arc) *timemachine.PageFour {
	if item.IsCoin == 0 {
		return nil
	}
	arc, ok := arcs[item.CoinAvid]
	if !ok || arc == nil {
		log.Warn("pageFour mid(%d) CoinAvid(%d) not found", item.Mid, item.CoinAvid)
		return nil
	}
	return &timemachine.PageFour{
		CoinTime:  item.CoinTime,
		CoinUsers: item.CoinUsers,
		Arc: &timemachine.Arc{
			Aid:      arc.Aid,
			Title:    arc.Title,
			Pic:      arc.Pic,
			Duration: arc.Duration,
			Owner:    arc.Author,
		},
	}
}

func pageFive(item *timemachine.Item, seasons map[int32]*api.CardInfoProto) *timemachine.PageFive {
	if item.PlayAmsDuration == 0 {
		return nil
	}
	season, ok := seasons[item.BestLikeSid]
	if !ok || season == nil {
		log.Warn("pageFour mid(%d) BestLikeSid(%d) not found", item.Mid, item.BestLikeSid)
		return nil
	}
	return &timemachine.PageFive{
		PlayBangumi: item.PlayFjs + item.PlayGcs,
		BestLikeSeason: &timemachine.Season{
			SeasonID:       season.SeasonId,
			Title:          season.Title,
			Cover:          season.Cover,
			SeasonType:     season.SeasonType,
			SeasonTypeName: season.SeasonTypeName,
		},
	}
}

func pageSix(item *timemachine.Item, seasons map[int32]*api.CardInfoProto) *timemachine.PageSix {
	if item.BestLikeYinshi == 0 {
		return nil
	}
	season, ok := seasons[item.BestLikeYinshi]
	if !ok || season == nil {
		log.Warn("pageSix mid(%d) BestLikeYinshi(%d) not found", item.Mid, item.BestLikeYinshi)
		return nil
	}
	return &timemachine.PageSix{
		PlayMovies:       item.PlayMovies,
		PlayDramas:       item.PlayDramas,
		PlayDocumentarys: item.PlayDocumentarys,
		PlayZongyi:       item.PlayZongyi,
		BestLikeYinshi: &timemachine.Season{
			SeasonID:       season.SeasonId,
			Title:          season.Title,
			Cover:          season.Cover,
			SeasonType:     season.SeasonType,
			SeasonTypeName: season.SeasonTypeName,
		},
	}
}

func (s *Service) pageSeven(item *timemachine.Item) *timemachine.PageSeven {
	if item.EventID <= 0 {
		return nil
	}
	event, ok := s.events[item.EventID]
	if !ok || event == nil {
		log.Warn("pageSeven mid(%d) EventID(%d) not found", item.Mid, item.EventID)
		return nil
	}
	eventTime, eventErr := time.ParseInLocation("2006-01-02", event.PreTime, time.Local)
	viewTime, viewTimeErr := time.ParseInLocation("20060102", item.FirstViewTime, time.Local)
	if eventErr == nil && viewTimeErr == nil {
		if viewTime.Before(eventTime) {
			log.Warn("pageSeven mid(%d) EventID(%d) eventTime(%s) firstViewTime(%s) not found", item.Mid, item.EventID, event.PreTime, item.FirstViewTime)
			return nil
		}
	} else {
		log.Warn("pageSeven mid(%d) eventErr(%v) viewTimeErr(%v)", item.Mid, eventErr, viewTimeErr)
	}
	return &timemachine.PageSeven{
		ViewTime:   item.FirstViewTime,
		EventID:    item.EventID,
		EventTitle: event.Title,
		EventDesc:  event.Desc,
	}
}

func pageEight(item *timemachine.Item, infos map[int64]*accapi.Info, arcs map[int64]*arcapi.Arc) *timemachine.PageEight {
	if item.LikeBestUp == 0 {
		return nil
	}
	info, ok := infos[item.LikeBestUp]
	if !ok || info == nil {
		log.Warn("pageEight mid(%d) LikeBestUp(%d) not found", item.Mid, item.LikeBestUp)
		return nil
	}
	arc, ok := arcs[item.LikeUpBestCreate]
	if !ok || arc == nil || !arc.IsNormal() {
		log.Warn("pageEight mid(%d) LikeBestUp(%d) LikeUpBestCreate(%d) not found", item.Mid, item.LikeBestUp, item.LikeUpBestCreate)
		return nil
	}
	return &timemachine.PageEight{
		Mid:  info.Mid,
		Name: info.Name,
		Face: info.Face,
		Arc: &timemachine.Arc{
			Aid:      arc.Aid,
			Title:    arc.Title,
			Pic:      arc.Pic,
			Duration: arc.Duration,
			Owner:    arc.Author,
		},
	}
}

func pageNine(item *timemachine.Item, infos map[int64]*accapi.Info) *timemachine.PageNine {
	if item.LikeBestLiveUp == 0 {
		return nil
	}
	info, ok := infos[item.LikeBestLiveUp]
	if !ok || info == nil {
		log.Warn("pageNine mid(%d) LikeBestLiveUp(%d) not found", item.Mid, item.LikeBestLiveUp)
		return nil
	}
	return &timemachine.PageNine{
		Mid:      info.Mid,
		Name:     info.Name,
		Face:     info.Face,
		Duration: item.LikeLiveupPlayDuration,
	}
}

func pageTen(item *timemachine.Item, arcs map[int64]*arcapi.Arc, arts map[int64]*artmdl.Meta) *timemachine.PageTen {
	if item.IsValidup == 0 {
		return nil
	}
	res := &timemachine.PageTen{
		CreateAvs:   item.CreateAvs,
		CreateReads: item.CreateReads,
		AvVv:        item.AvVv,
		ReadVv:      item.ReadVv,
	}
	if item.UpBestCreatType == _upCreateTypeArc {
		res.Type = _upCreateTypeArc
		if arc, ok := arcs[item.BestCreate]; ok && arc != nil && arc.IsNormal() {
			res.Arc = &timemachine.Arc{
				Aid:      arc.Aid,
				Title:    arc.Title,
				Pic:      arc.Pic,
				Duration: arc.Duration,
				Owner:    arc.Author,
			}
		}
	}
	if item.UpBestCreatType == _upCreateTypeArt {
		res.Type = _upCreateTypeArt
		if art, ok := arts[item.BestCreate]; ok && art != nil {
			res.Arc = &timemachine.Arc{
				Aid:   art.ID,
				Title: art.Title,
				Owner: arcapi.Author{
					Mid:  art.Author.Mid,
					Name: art.Author.Name,
					Face: art.Author.Face,
				},
			}
			if len(art.ImageURLs) > 0 {
				res.Arc.Pic = art.ImageURLs[0]
			}
		}
	}
	if res.Arc == nil {
		log.Warn("pageTen mid(%d) BestCreate(%d) type(%d) not found", item.Mid, item.BestCreate, item.UpBestCreatType)
	}
	return res
}

func pageEleven(item *timemachine.Item, infos map[int64]*accapi.Info) *timemachine.PageEleven {
	if item.IsHaveBestFan == 0 {
		return nil
	}
	info, ok := infos[item.BestFanMid]
	if !ok || info == nil {
		log.Warn("pageEleven mid(%d) BestFanMid(%d) not found", item.Mid, item.BestFanMid)
		return nil
	}
	return &timemachine.PageEleven{
		Mid:       info.Mid,
		Name:      info.Name,
		Face:      info.Face,
		BestFanVv: item.BestFanVv,
	}
}

func pageTwelve(item *timemachine.Item) *timemachine.PageTwelve {
	if item.IsValidLiveUp == 0 {
		return nil
	}
	return &timemachine.PageTwelve{
		LiveDays:         item.LiveDays,
		Ratio:            item.Ratio,
		MaxOnlineNumTime: item.MaxOnlineNumDate,
		MaxOnlineNum:     item.MaxOnlineNum,
	}
}

func pageThirteen(item *timemachine.Item, infos map[int64]*accapi.Info) *timemachine.PageThirteen {
	if item.BestLiveFanMid <= 0 {
		return nil
	}
	info, ok := infos[item.BestLiveFanMid]
	if !ok || info == nil {
		log.Warn("pageThirteen mid(%d) BestLiveFanMid(%d) not found", item.Mid, item.BestLiveFanMid)
		return nil
	}
	return &timemachine.PageThirteen{
		Mid:  info.Mid,
		Name: info.Name,
		Face: info.Face,
	}
}

func (s *Service) pageFourteen(item *timemachine.Item, flags []*timemachine.Flag) *timemachine.PageFourteen {
	if len(flags) == 0 {
		flags = make([]*timemachine.Flag, 0)
	}
	var flagDesc string
	regionDesc, ok := s.regionDescs[item.MaxVvSubtid]
	if ok && regionDesc != nil {
		flagDesc = regionDesc.FlagDesc
	}
	if flagDesc == "" {
		flagDesc = s.dftRegionDesc.FlagDesc
	}
	return &timemachine.PageFourteen{
		RegionDesc: flagDesc,
		Flags:      flags,
	}
}
