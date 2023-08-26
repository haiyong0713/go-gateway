package service

import (
	"context"
	"encoding/json"
	"fmt"
	"runtime/debug"
	"sort"
	"sync/atomic"
	"time"

	"go-common/library/database/bfs"
	favmdl "go-main/app/community/favorite/service/model"

	actApi "go-gateway/app/web-svr/activity/interface/api"
	"go-gateway/app/web-svr/esports/interface/client"
	"go-gateway/app/web-svr/esports/interface/component"
	"go-gateway/app/web-svr/esports/interface/conf"
	"go-gateway/app/web-svr/esports/interface/model"

	favApi "go-main/app/community/favorite/service/api"
)

const (
	contestSortByStartTime = iota
	contestSortByEndTime
	contestSortByStatus
	contestSortByLive

	bfsBackupFilename4LOLContests      = "contests"
	bfsBackupFilename4LOLContestsLive  = "contests_live"
	bfsBackupFilename4LOLScoreAnalysis = "analysis_aggregation"

	cover4PosterNotify    = "https://i0.hdslb.com/bfs/activity-plat/static/20200901/8a3e1fa14e30dc3be9c5324f604e5991/RH6PwP3w2.png"
	coverKey4PosterNotify = "poster_notify"
)

const (
	bfsTypeOfS104DefaultBiz = iota
	bfsTypeOfS104LiveBiz
	bfsTypeOfS104ScoreAnalysis
)

var (
	contestCardListMap     atomic.Value
	contestCardList4LPLMap atomic.Value
	contestMap             atomic.Value
	contestSeriesMap       atomic.Value
	avCIDMap               atomic.Value

	contestCardList4Live []*model.ContestSeries

	currentSeries *model.ContestSeries

	posterList *model.PosterList4S10

	contestCardList  []*model.ContestCard
	contestBroadcast []*model.ContestCard

	contestCardListMap4All map[int64][]*model.ContestCard
	contestCardTsList4All  []int64
	contestCardListMap4LPL map[int64][]*model.ContestCard
	contestCardTsList4LPL  []int64

	contestCardList4AllInTab []*model.ContestCard
	contestCardList4LPLInTab []*model.ContestCard
)

func init() {
	contestCardList4AllInTab = make([]*model.ContestCard, 0)
	contestCardList4LPLInTab = make([]*model.ContestCard, 0)

	contestCardListMap4All = make(map[int64][]*model.ContestCard, 0)
	contestCardTsList4All = make([]int64, 0)
	contestCardListMap4LPL = make(map[int64][]*model.ContestCard, 0)
	contestCardTsList4LPL = make([]int64, 0)

	contestCardList = make([]*model.ContestCard, 0)

	cardListMap := make(map[int64][]*model.ContestCard, 0)
	contestCardListMap.Store(cardListMap)

	cardListMap4LPL := make(map[int64][]*model.ContestCard, 0)
	contestCardList4LPLMap.Store(cardListMap4LPL)

	contestM := make(map[int64]*model.Contest4Frontend, 0)
	contestMap.Store(contestM)

	contestSeriesM := make(map[int64]*model.ContestSeries, 0)
	contestSeriesMap.Store(contestSeriesM)

	currentSeries = new(model.ContestSeries)

	avCIDM := make(map[int64]int64, 0)
	avCIDMap.Store(avCIDM)

	contestCardList4Live = make([]*model.ContestSeries, 0)

	tmpPosterList := new(model.PosterList4S10)
	{
		tmpPosterList.List = make([]*model.Poster4S10, 0)
	}
	posterList = tmpPosterList

	contestBroadcast = make([]*model.ContestCard, 0)
}

func deepCopyContestCard(card *model.ContestCard) *model.ContestCard {
	tmpCard := new(model.ContestCard)
	{
		tmpMoreList := make([]*model.ContestMore, 0)
		tmpContest4Frontend := new(model.Contest4Frontend)
		*tmpContest4Frontend = *card.Contest

		tmpCard.More = tmpMoreList
		tmpCard.Contest = tmpContest4Frontend
		tmpCard.Timestamp = card.Timestamp
	}

	for _, v := range card.More {
		tmpMore := new(model.ContestMore)
		*tmpMore = *v

		tmpCard.More = append(tmpCard.More, tmpMore)
	}

	return tmpCard
}

func deepCopyPoster4S10(poster *model.Poster4S10) *model.Poster4S10 {
	tmp := new(model.Poster4S10)
	{
		tmp.More = make([]*model.ContestMore, 0)
		tmp.BackGround = poster.BackGround
		tmp.Contest = poster.Contest
		tmp.ContestID = poster.ContestID
		tmp.InCenter = poster.InCenter
	}

	for _, v := range poster.More {
		tmpMore := new(model.ContestMore)
		*tmpMore = *v
		tmp.More = append(tmp.More, tmpMore)
	}

	return tmp
}

func deepCopySeries(series *model.ContestSeries) *model.ContestSeries {
	tmpSeries := new(model.ContestSeries)
	{
		tmpSeries.ID = series.ID
		tmpSeries.Detail = make([]*model.ContestCardList4Live, 0)
		tmpSeries.InTheSeries = series.InTheSeries
		tmpSeries.StartTime = series.StartTime
		tmpSeries.EndTime = series.EndTime
		tmpSeries.ParentTitle = series.ParentTitle
		tmpSeries.ScoreID = series.ScoreID
		tmpSeries.ChildTitle = series.ChildTitle
	}

	if series.Detail != nil && len(series.Detail) > 0 {
		for _, v := range series.Detail {
			tmpDetail := new(model.ContestCardList4Live)
			*tmpDetail = *v
			tmpSeries.Detail = append(tmpSeries.Detail, tmpDetail)
		}
	}

	return tmpSeries
}

func watchSeasonBiz() {
	go watchAvCIDMap(context.Background())
	go watchS10ContestList(context.Background())
	go watchS10ContestSeries(context.Background())
	go watchS10ScoreAnalysis(context.Background())
	go watchS10PosterList(context.Background())
}

func watchS10PosterList(ctx context.Context) {
	ticker := time.NewTicker(time.Second * 5)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			watchS10PosterListBySeasonWatch(ctx)
		case <-ctx.Done():
			return
		}
	}
}

func watchS10PosterListBySeasonWatch(ctx context.Context) {
	d := conf.LoadSeasonContestWatch()
	if d == nil || !d.CanWatch() {
		return
	}

	newList := new(model.PosterList4S10)
	if err := component.GlobalMemcached.Get(ctx, d.CacheKey4PosterList).Scan(&newList); err == nil {
		posterList = newList
	}
}

func loadPosterList() ([]*model.Poster4S10, []int64) {
	newList := make([]*model.Poster4S10, 0)
	contestIDList := make([]int64, 0)
	for _, v := range posterList.List {
		tmpPoster := deepCopyPoster4S10(v)
		newList = append(newList, tmpPoster)

		if tmpPoster.Contest.Status == model.ContestStatusOfNotStart {
			contestIDList = append(contestIDList, v.ContestID)
		}
	}

	return newList, contestIDList
}

func (s *Service) ResetBfsBackup(ctx context.Context) {
	ticker := time.NewTicker(time.Minute * 3)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			_ = s.genBfsBackupDataByBizType(ctx, bfsTypeOfS104DefaultBiz)
			_ = s.genBfsBackupDataByBizType(ctx, bfsTypeOfS104LiveBiz)
			_ = s.genBfsBackupDataByBizType(ctx, bfsTypeOfS104ScoreAnalysis)
		case <-ctx.Done():
			return
		}
	}
}

func (s *Service) genBfsBackupDataByBizType(ctx context.Context, bizType int64) (m map[string]interface{}) {
	m = make(map[string]interface{}, 0)
	switch bizType {
	case bfsTypeOfS104DefaultBiz:
		m["tab"] = s.S10Tab4Contest(ctx, 0, 0)
		m["live_list"] = allAvCIDContest()
		m["contest_more"] = s.S10MoreContest(ctx, 0)
	case bfsTypeOfS104LiveBiz:
		m["series"] = s.S10LiveContestSeries(ctx, 0)
		m["contest_more"] = s.S10LiveMoreContest(ctx, 0)
	case bfsTypeOfS104ScoreAnalysis:
		tmpM := make(map[string]interface{}, 0)
		{
			tmpM["analysisTeam"] = s.S10ScoreAnalysis(
				ctx,
				&model.ScoreAnalysisRequest{
					AnalysisType: analysisType4Team,
					SortType:     sortDesc,
					SortKey:      sortKey4TeamOfTotalRound,
				})
			tmpM["analysisPlayer"] = s.S10ScoreAnalysis(
				ctx,
				&model.ScoreAnalysisRequest{
					AnalysisType: analysisType4Player,
					SortType:     sortDesc,
					SortKey:      sortKey4TeamOfTotalRound,
				})
			tmpM["analysisHero"] = s.S10ScoreAnalysis(
				ctx,
				&model.ScoreAnalysisRequest{
					AnalysisType: analysisType4Hero,
					SortType:     sortDesc,
					SortKey:      sortKey4TeamOfTotalRound,
				})
		}
		m["code"] = 0
		m["data"] = tmpM
	}

	bs, _ := json.Marshal(m)
	fileName := ""
	switch bizType {
	case bfsTypeOfS104DefaultBiz:
		fileName = bfsBackupFilename4LOLContests
	case bfsTypeOfS104LiveBiz:
		fileName = bfsBackupFilename4LOLContestsLive
	case bfsTypeOfS104ScoreAnalysis:
		fileName = bfsBackupFilename4LOLScoreAnalysis
	}

	if fileName != "" {
		now := time.Now()
		var (
			url string
			err error
		)

		defer func() {
			fmt.Println(
				fmt.Sprintf(
					"bfs_backup_biz: bizType(%v) >>> url(%v), err(%v), cost(%v)",
					bizType, url, err, time.Since(now)))
		}()

		bfsClient := bfs.New(nil)
		url, err = bfsClient.Upload(context.Background(), &bfs.Request{
			Filename:    fileName,
			Bucket:      "esport",
			Dir:         "LOL/S10/2020",
			ContentType: "application/json",
			File:        bs,
		})
	}

	return
}

func watchS10ContestList(ctx context.Context) {
	ticker := time.NewTicker(time.Second * 1)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			watchS10ContestListBySeasonWatch(ctx)
		case <-ctx.Done():
			return
		}
	}
}

func watchS10ContestSeries(ctx context.Context) {
	ticker := time.NewTicker(time.Minute * 1)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			watchS10ContestSeriesBySeasonWatch(ctx)
		case <-ctx.Done():
			return
		}
	}
}

func watchS10ContestSeriesBySeasonWatch(ctx context.Context) {
	d := conf.LoadSeasonContestWatch()
	if d == nil || !d.CanWatch() {
		return
	}

	tmpM := make(map[int64][]*model.ContestCard, 0)
	mcScanErr := component.GlobalMemcached.Get(ctx, d.ContestListCacheKey).Scan(&tmpM)
	if mcScanErr == nil && len(tmpM) > 0 {
		seriesMap := make(map[int64]*model.ContestSeries, 0)
		mcScanErr = component.GlobalMemcached.Get(ctx, d.ContestSeriesMapCacheKey).Scan(&seriesMap)
		if mcScanErr == nil && len(seriesMap) > 0 {
			contestSeriesMap.Store(seriesMap)
			updateContestSeriesByCardAndSeriesMap(tmpM, seriesMap)
			resetCurrentSeries(seriesMap)
		}
	}
}

func watchS10ContestListBySeasonWatch(ctx context.Context) {
	d := conf.LoadSeasonContestWatch()
	if d == nil || !d.CanWatch() {
		return
	}

	tmpM := make(map[int64][]*model.ContestCard, 0)
	mcScanErr := component.GlobalMemcached.Get(ctx, d.ContestListCacheKey).Scan(&tmpM)
	if mcScanErr == nil && len(tmpM) > 0 {
		rebuildContestList(tmpM)
	}
}

func resetCurrentSeries(seriesM map[int64]*model.ContestSeries) {
	seriesList := make([]*model.ContestSeries, 0)
	for _, v := range seriesM {
		tmpSeries := deepCopySeries(v)
		seriesList = append(seriesList, tmpSeries)
	}

	sort.SliceStable(seriesList, func(i, j int) bool {
		return seriesList[i].StartTime < seriesList[j].StartTime
	})
	now := time.Now().Unix()
	preSeries := new(model.ContestSeries)
	for _, series := range seriesList {
		if now < series.EndTime {
			currentSeries = deepCopySeries(series)

			return
		}

		preSeries = deepCopySeries(series)
	}

	currentSeries = preSeries
}

func updateContestSeriesByCardAndSeriesMap(cardM map[int64][]*model.ContestCard, seriesM map[int64]*model.ContestSeries) {
	seriesList := make([]*model.ContestSeries, 0)
	isFirst4Series := true
	var lastDiffSeconds4Series, lastLocatedID4Series int64

	now := time.Now().Unix()
	for _, v := range seriesM {
		if isFirst4Series {
			lastLocatedID4Series = v.ID
			lastDiffSeconds4Series = v.StartTime - now
			if lastDiffSeconds4Series < 0 {
				lastDiffSeconds4Series = -lastDiffSeconds4Series
			}

			isFirst4Series = false
		}

		diffSeconds := v.StartTime - now
		if diffSeconds < 0 {
			diffSeconds = -diffSeconds
		}

		anotherDiffSeconds := v.EndTime - now
		if anotherDiffSeconds < 0 {
			anotherDiffSeconds = -anotherDiffSeconds
		}

		if anotherDiffSeconds < diffSeconds {
			diffSeconds = anotherDiffSeconds
		}

		if diffSeconds < lastDiffSeconds4Series {
			lastLocatedID4Series = v.ID
			lastDiffSeconds4Series = diffSeconds
		}

		tmpSeries := deepCopySeries(v)
		tmpSeries.Detail = make([]*model.ContestCardList4Live, 0)
		seriesList = append(seriesList, tmpSeries)
	}

	var (
		lastDiffSeconds, lastLocatedTs int64
		dayUnix                        = currentDateUnix()
		isFirst                        = true
	)

	// generate series firstly
	for _, v := range seriesList {
		for dateUnix, cardList := range cardM {
			tmpCardList := make([]*model.ContestCard, 0)

			for _, card := range cardList {
				if card.Contest.Series.ID == v.ID {
					tmpCard := deepCopyContestCard(card)
					tmpCardList = append(tmpCardList, tmpCard)
				}
			}

			if len(tmpCardList) > 0 {
				sort.SliceStable(tmpCardList, func(i, j int) bool {
					return tmpCardList[i].Contest.CalculateTimestampDiff() < tmpCardList[j].Contest.CalculateTimestampDiff()
				})

				cardList4Live := new(model.ContestCardList4Live)
				{
					cardList4Live.Timestamp = dateUnix
					cardList4Live.CardList = tmpCardList
				}

				diffSeconds := dateUnix - dayUnix
				if diffSeconds < 0 {
					diffSeconds = -diffSeconds
				}

				if isFirst {
					lastLocatedTs = dateUnix
					lastDiffSeconds = diffSeconds
					isFirst = false
				}

				if diffSeconds < lastDiffSeconds {
					lastLocatedTs = dateUnix
					lastDiffSeconds = diffSeconds
				}

				v.Detail = append(v.Detail, cardList4Live)
			}
		}

		if len(v.Detail) > 1 {
			tmpDetail := v.Detail

			sort.SliceStable(tmpDetail, func(i, j int) bool {
				return tmpDetail[i].Timestamp < tmpDetail[j].Timestamp
			})

			v.Detail = tmpDetail
		}
	}

	if len(seriesList) > 1 {
		sort.SliceStable(seriesList, func(i, j int) bool {
			return seriesList[i].StartTime < seriesList[j].StartTime
		})
	}

	for _, v := range seriesList {
		lastLocatedID4Series = v.ID
		if now <= v.EndTime {
			break
		}
	}

	for _, series := range seriesList {
		if series.ID == lastLocatedID4Series {
			series.InTheSeries = true
		}

		for _, card := range series.Detail {
			if card.Timestamp == lastLocatedTs {
				card.IsLocated = true
			}
		}
	}
	resetBroadcast()
	contestCardList4Live = seriesList
}

func resetBroadcast() {
	now := time.Now().Unix()
	beforeList := make([]*model.ContestCard, 0)
	afterList := make([]*model.ContestCard, 0)

	for _, v := range contestCardList {
		tmpCard := deepCopyContestCard(v)
		if tmpCard.Contest.EndTime < now {
			beforeList = append(beforeList, tmpCard)
		} else {
			afterList = append(afterList, tmpCard)
		}
	}

	sortContestList(beforeList, contestSortByEndTime)
	if len(beforeList) > 10 {
		beforeList = beforeList[0:10]
	}

	sortContestList(afterList, contestSortByStartTime)
	if len(afterList) > 20 {
		afterList = afterList[0:20]
	}

	list := make([]*model.ContestCard, 0)
	list = append(list, beforeList...)
	list = append(list, afterList...)

	contestBroadcast = list
}

func loadContestSeries() (list []*model.ContestSeries) {
	list = make([]*model.ContestSeries, 0)
	if contestCardList4Live != nil && len(contestCardList4Live) > 0 {
		for _, v := range contestCardList4Live {
			tmpSeries := deepCopySeries(v)
			tmpSeries.Detail = make([]*model.ContestCardList4Live, 0)
			//tmpSeries.Rebuild()

			for _, cardList4Date := range v.Detail {
				tmpDetail := make([]*model.ContestCardList4Live, 0)
				if tmpSeries.Detail != nil {
					tmpDetail = tmpSeries.Detail
				}

				tmpLiveCardList := new(model.ContestCardList4Live)
				*tmpLiveCardList = *cardList4Date
				tmpLiveCardList.CardList = make([]*model.ContestCard, 0)

				for _, card := range cardList4Date.CardList {
					tmpCard := deepCopyContestCard(card)
					tmpLiveCardList.CardList = append(tmpLiveCardList.CardList, tmpCard)
				}

				tmpDetail = append(tmpDetail, tmpLiveCardList)
				tmpSeries.Detail = tmpDetail
			}

			list = append(list, tmpSeries)
		}
	}

	return
}

func rebuildContestList(m map[int64][]*model.ContestCard) {
	defer func() {
		if err := recover(); err != nil {
			fmt.Println("rebuildContestList >>> panic: ", err, string(debug.Stack()))
		}
	}()

	contestCardListMap.Store(m)
	resetContestCardRelations(false)
	rebuildContestCardList4LPL(m)

	timestampList := make([]int64, 0)
	contestM := make(map[int64]*model.Contest4Frontend, 0)
	tmpContestCardList := make([]*model.ContestCard, 0)

	for k, cardList := range m {
		timestampList = append(timestampList, k)
		for _, card := range cardList {
			tmpContest := new(model.Contest4Frontend)
			*tmpContest = *card.Contest
			contestM[card.Contest.ID] = tmpContest
			tmpContestCardList = append(tmpContestCardList, deepCopyContestCard(card))
		}
	}

	contestMap.Store(contestM)
	sort.SliceStable(timestampList, func(i, j int) bool {
		return timestampList[i] < timestampList[j]
	})
	contestCardTsList4All = timestampList
	sort.SliceStable(tmpContestCardList, func(i, j int) bool {
		return tmpContestCardList[i].Contest.StartTime < tmpContestCardList[j].Contest.StartTime
	})
	contestCardList = tmpContestCardList

	dayUnix := currentDateUnix()
	contestCardList4AllInTab = contestCardListByTimestamp(dayUnix, false)
	contestCardList4LPLInTab = contestCardListByTimestamp(dayUnix, true)
}

func genContestCardListInTab(fromLPL bool) []*model.ContestCard {
	list := make([]*model.ContestCard, 0)
	if fromLPL {
		for _, v := range contestCardList4LPLInTab {
			list = append(list, deepCopyContestCard(v))
		}
	} else {
		for _, v := range contestCardList4AllInTab {
			list = append(list, deepCopyContestCard(v))
		}
	}

	return list
}

func rebuildContestCardList4LPL(m map[int64][]*model.ContestCard) {
	newM := make(map[int64][]*model.ContestCard, 0)
	tmpContestCardList := make([]*model.ContestCard, 0)
	for k, cardList := range m {
		newCardList := make([]*model.ContestCard, 0)
		if d, ok := newM[k]; ok {
			newCardList = d
		}

		for _, card := range cardList {
			if card.FromLPL() {
				tmpCard := deepCopyContestCard(card)
				newCardList = append(newCardList, tmpCard)
			}
		}

		if len(newCardList) > 0 {
			newM[k] = newCardList
			tmpContestCardList = append(tmpContestCardList, newCardList...)
		}
	}

	if len(newM) > 0 {
		contestCardList4LPLMap.Store(newM)

		timestampList := make([]int64, 0)
		for k := range newM {
			timestampList = append(timestampList, k)
		}
		sort.SliceStable(timestampList, func(i, j int) bool {
			return timestampList[i] < timestampList[j]
		})

		contestCardTsList4LPL = timestampList
		resetContestCardRelations(true)
	}
}

func resetContestCardRelations(fromLPL bool) {
	tmp := make(map[int64][]*model.ContestCard, 0)
	m := make(map[int64][]*model.ContestCard, 0)

	if !fromLPL {
		tmp = contestCardListMap.Load().(map[int64][]*model.ContestCard)
	} else {
		tmp = contestCardList4LPLMap.Load().(map[int64][]*model.ContestCard)
	}

	for k, cardList := range tmp {
		newCardList := make([]*model.ContestCard, 0)
		if d, ok := m[k]; ok {
			newCardList = d
		}

		for _, card := range cardList {
			newCard := deepCopyContestCard(card)
			newCardList = append(newCardList, newCard)
		}

		if len(newCardList) > 0 {
			m[k] = newCardList
		}
	}

	if !fromLPL {
		contestCardListMap4All = m
	} else {
		contestCardListMap4LPL = m
	}
}

func watchAvCIDMap(ctx context.Context) {
	ticker := time.NewTicker(time.Second * 1)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			watchAvCIDMapBySeasonWatch(ctx)
		case <-ctx.Done():
			return
		}
	}
}

func watchAvCIDMapBySeasonWatch(ctx context.Context) {
	d := conf.LoadSeasonContestWatch()
	if d == nil || !d.CanWatch() {
		return
	}

	tmpM := make(map[int64]int64, 0)
	mcScanErr := component.GlobalMemcached.Get(ctx, d.ContestAvCIDListCacheKey).Scan(&tmpM)
	if mcScanErr == nil {
		avCIDMap.Store(tmpM)
	}
}

func allAvCIDContest() map[int64]*model.Live4Frontend {
	m := make(map[int64]*model.Live4Frontend, 0)
	avCIDM := loadAvCIDMap()
	for avCID := range avCIDM {
		m[avCID] = loadAvCIDContest(avCID)
	}

	return m
}

func loadAvCIDMap() (m map[int64]int64) {
	tmpM := avCIDMap.Load().(map[int64]int64)
	if tmpM != nil {
		m = tmpM
	}

	return
}

func loadAvCIDContest(avCID int64) *model.Live4Frontend {
	live := new(model.Live4Frontend)
	if avCID > 0 {
		m := loadAvCIDMap()
		if d, ok := m[avCID]; ok {
			if contestM := contestMap.Load().(map[int64]*model.Contest4Frontend); contestM != nil {
				if contest, ok := contestM[d]; ok {
					{
						live.IsLive = true
						live.ContestID = contest.ID
						live.Home = contest.Home
						live.Away = contest.Away
					}
				}
			}
		}
	}

	return live
}

func loadContestCardRelations(fromLPL bool) (map[int64][]*model.ContestCard, []int64) {
	m := make(map[int64][]*model.ContestCard, 0)
	timestampList := make([]int64, 0)

	if !fromLPL {
		for k, v := range contestCardListMap4All {
			tmpList := make([]*model.ContestCard, 0)
			for _, d := range v {
				tmpCard := deepCopyContestCard(d)
				tmpList = append(tmpList, tmpCard)
			}

			m[k] = tmpList
		}

		for _, v := range contestCardTsList4All {
			timestampList = append(timestampList, v)
		}
	} else {
		for k, v := range contestCardListMap4LPL {
			tmpList := make([]*model.ContestCard, 0)
			for _, d := range v {
				tmpCard := deepCopyContestCard(d)
				tmpList = append(tmpList, tmpCard)
			}

			m[k] = tmpList
		}

		for _, v := range contestCardTsList4LPL {
			timestampList = append(timestampList, v)
		}
	}

	return m, timestampList
}

func (s *Service) S10LiveMoreContest(ctx context.Context, mid int64) map[string]interface{} {
	var locatedTs, lastSecondsDiff int64
	cardListMap, tsList := loadContestCardRelations(false)
	needTsList := genLiveNeedTimestampList(tsList)
	list := make([]*model.ContestCard, 0)
	for _, ts := range needTsList {
		if cardList, ok := cardListMap[ts]; ok {
			for _, card := range cardList {
				tmpCard := deepCopyContestCard(card)
				tmpCard.Timestamp = ts

				list = append(list, tmpCard)
			}
		}
	}

	now := currentDateUnix()
	for _, v := range needTsList {
		if now == v {
			locatedTs = v

			break
		}

		diff := v - now
		if diff < 0 {
			diff = -diff
		}

		if lastSecondsDiff == 0 || diff < lastSecondsDiff {
			lastSecondsDiff = diff
			locatedTs = v
		}
	}

	if locatedTs == 0 && len(needTsList) > 0 {
		locatedTs = needTsList[len(needTsList)-1]
	}

	list = sortContestList(list, contestSortByStatus)
	m := make(map[string]interface{}, 2)
	{
		m["ts"] = needTsList
		m["list"] = s.rebuildMore(ctx, mid, list)
		m["located_ts"] = locatedTs
	}

	return m
}

func currentDateUnix() int64 {
	year, month, day := time.Now().Date()

	return time.Date(year, month, day, 0, 0, 0, 0, time.Local).Unix()
}

func genLiveNeedTimestampList(all []int64) (list []int64) {
	list = make([]int64, 0)
	dayUnix := currentDateUnix()

	for _, v := range all {
		if v >= dayUnix {
			list = append(list, v)
		}
	}

	if len(list) > 4 {
		list = list[:4]
	}

	return
}

func (s *Service) S10CurrentSeries(ctx context.Context) (m map[string]interface{}) {
	seasonCfg := conf.LoadS10SeasonCfg()
	m = make(map[string]interface{})
	{
		m["off_season"] = seasonCfg.OffSeason
		m["champion"] = seasonCfg.Champion
		m["desc_link"] = seasonCfg.Desc
		m["series"] = currentSeries
	}

	return
}

func (s *Service) S10Poster4Activity(ctx context.Context, mid int64) map[string]interface{} {
	tmpPosterList, contestIDList := loadPosterList()
	if len(contestIDList) > 0 {
		favorites := s.fetchFavoriteMap(ctx, mid, contestIDList)
		guessMap := s.fetchContestGuessMap(ctx, mid, contestIDList)

		for _, v := range tmpPosterList {
			tmpMoreList := make([]*model.ContestMore, 0)

			for _, more := range v.More {
				tmpMore := new(model.ContestMore)
				switch more.Status {
				case model.MoreStatusOfSubscribe:
					if d, ok := favorites[v.ContestID]; ok && d {
						more.OnClick = model.ClickStatusOfDisabled
						more.Title = model.MoreDisplayOfSubscribed

						*tmpMore = *more
						tmpMoreList = append(tmpMoreList, tmpMore)
					} else {
						if v.Contest.StartTime >= time.Now().Unix() {
							*tmpMore = *more
							tmpMoreList = append(tmpMoreList, tmpMore)
						}
					}
				case model.MoreStatusOfPrediction:
					if d, ok := guessMap[v.ContestID]; ok && d {
						more.OnClick = model.ClickStatusOfDisabled
						more.Title = model.MoreDisplayOfPredicted

						*tmpMore = *more
						tmpMoreList = append(tmpMoreList, tmpMore)
					} else {
						if v.Contest.StartTime-time.Now().Unix() > secondsOf10Minutes {
							*tmpMore = *more
							tmpMoreList = append(tmpMoreList, tmpMore)
						}
					}
				default:
					*tmpMore = *more
					tmpMoreList = append(tmpMoreList, tmpMore)
				}
			}

			v.More = tmpMoreList
		}
	}

	m := make(map[string]interface{}, 2)
	{
		m["poster_list"] = tmpPosterList
		m["contest_list"] = contestBroadcast
		m["banner"] = cover4PosterNotify
		if d, ok := conf.GlobalTabCovers[coverKey4PosterNotify]; ok && d != "" {
			m["banner"] = d
		}
	}

	return m
}

func (s *Service) S10MoreContest(ctx context.Context, mid int64) map[string]interface{} {
	year, month, day := time.Now().Date()
	dayUnix := time.Date(year, month, day, 0, 0, 0, 0, time.Local).Unix()

	beforeToday := make([]int64, 0)
	afterToday := make([]int64, 0)
	cardListMap, tsList := loadContestCardRelations(false)

	for _, v := range tsList {
		if v < dayUnix {
			if len(beforeToday) >= 4 {
				beforeToday = append(beforeToday[1:], v)
			} else {
				beforeToday = append(beforeToday, v)
			}
		} else if v >= dayUnix {
			if len(afterToday) >= 4 {
				//afterToday = append(afterToday[1:], v)
			} else {
				afterToday = append(afterToday, v)
			}
		}
	}

	needTsList := make([]int64, 0)
	switch len(afterToday) {
	case 0:
		needTsList = beforeToday
	case 1:
		if len(beforeToday) > 3 {
			needTsList = beforeToday[len(beforeToday)-3:]
		}
		needTsList = append(needTsList, afterToday...)
	case 2:
		if len(beforeToday) > 2 {
			needTsList = beforeToday[len(beforeToday)-2:]
		}
		needTsList = append(needTsList, afterToday...)
	default:
		if len(beforeToday) > 0 {
			needTsList = beforeToday[len(beforeToday)-1:]
		}

		needTsList = append(needTsList, afterToday...)
	}

	if len(needTsList) > 4 {
		needTsList = needTsList[:4]
	}

	list := make([]*model.MoreContestCard, 0)
	for _, ts := range needTsList {
		if d, ok := cardListMap[ts]; ok {
			tmpCardList := make([]*model.ContestCard, 0)
			for _, v := range d {
				tmpCard := deepCopyContestCard(v)
				tmpCardList = append(tmpCardList, tmpCard)
			}

			tmpCardList = sortContestList(tmpCardList, contestSortByStatus)
			moreContestCard := new(model.MoreContestCard)
			{
				moreContestCard.Timestamp = ts
				moreContestCard.ContestCards = tmpCardList
			}

			list = append(list, moreContestCard)
		}
	}

	m := make(map[string]interface{}, 2)
	{
		m["show_lpl"] = conf.LoadShowLPL()
		m["more"] = s.rebuildMore4MoreContestCard(ctx, mid, list)
	}

	return m
}

func (s *Service) rebuildMore4MoreContestCard(ctx context.Context, mid int64, list []*model.MoreContestCard) []*model.MoreContestCard {
	contestIDList4Predict := make([]int64, 0)
	for _, cards := range list {
		for _, card := range cards.ContestCards {
			if card.Contest.Status == model.ContestStatusOfNotStart {
				contestIDList4Predict = append(contestIDList4Predict, card.Contest.ID)
			}
		}
	}

	if mid > 0 {
		favorites := s.fetchFavoriteMap(ctx, mid, contestIDList4Predict)
		guessMap := s.fetchContestGuessMap(ctx, mid, contestIDList4Predict)

		if len(favorites) > 0 || len(guessMap) > 0 {
			for _, cards := range list {
				for _, card := range cards.ContestCards {
					card.ResetMore(favorites, guessMap)
				}
			}
		}
	}

	return list
}

func (s *Service) LiveBanner() (banner string) {
	tabCovers := conf.LoadS10TabCovers()
	if d, ok := tabCovers["banner"]; ok {
		banner = d
	}

	return
}

func (s *Service) PointsActWebBanner() (banner string) {
	tabCovers := conf.LoadS10TabCovers()
	if d, ok := tabCovers["pointsActWeb"]; ok {
		banner = d
	}

	return
}

func (s *Service) PointsActBanner() (banner string) {
	tabCovers := conf.LoadS10TabCovers()
	if d, ok := tabCovers["pointsAct"]; ok {
		banner = d
	}

	return
}

func (s *Service) LiveWebBanner() (banner string) {
	tabCovers := conf.LoadS10TabCovers()
	if d, ok := tabCovers["bannerWeb"]; ok {
		banner = d
	}

	return
}
func (s *Service) S10Tab4Contest(ctx context.Context, mid, avCID int64) *model.DataInContestArea {
	tabCovers4Frontend := model.TabCovers{}
	tabCovers := conf.LoadS10TabCovers()
	{
		if d, ok := tabCovers["top"]; ok {
			tabCovers4Frontend.Top = d
		}

		if d, ok := tabCovers["middle"]; ok {
			tabCovers4Frontend.Middle = d
		}

		if d, ok := tabCovers["bottom"]; ok {
			tabCovers4Frontend.Bottom = d
		}
	}

	contestCardList4Recent := genContestCardListInTab(false)
	contestCardList4LPL := genContestCardListInTab(true)

	contestCardList4Recent = s.rebuildMore(ctx, mid, contestCardList4Recent)
	contestCardList4LPL = s.rebuildMore(ctx, mid, contestCardList4LPL)

	live := loadAvCIDContest(avCID)

	data := new(model.DataInContestArea)
	{
		data.TabCovers = tabCovers4Frontend
		data.Live = live
		data.LPL = contestCardList4LPL
		data.Recent = contestCardList4Recent
		data.ShowLPL = conf.LoadShowLPL()
	}

	return data
}

func (s *Service) S10LiveContestSeries(ctx context.Context, mid int64) (list []*model.ContestSeries) {
	list = loadContestSeries()

	return s.rebuildContestSeriesMore(ctx, mid, list)
}

func contestCardListByTimestamp(ts int64, fromLPL bool) []*model.ContestCard {
	list := make([]*model.ContestCard, 0)
	cardListMap, tsList := loadContestCardRelations(fromLPL)

	todayCards, ok := cardListMap[ts]
	if ok {
		if len(todayCards) >= 3 {
			if isTodayContestListEnd(todayCards) {
				list = contestsInTheFuture(cardListMap, tsList, ts)
			}

			todayCards = sortContestList(todayCards, contestSortByStatus)
			if len(list) > 2 {
				list = list[:2]
			}

			needLen := 3 - len(list)
			list = append(list, todayCards[:needLen]...)
			list = sortContestList(list, contestSortByStatus)

			return list
		}
	}

	needLen := 3 - len(todayCards)
	list = contestsInTheFuture(cardListMap, tsList, ts)
	if len(list) >= needLen {
		list = sortContestList(list, contestSortByStatus)
		list = list[:needLen]
	}

	list = append(list, todayCards...)

	needLen = 3 - len(list)
	if needLen > 0 {
		if tmpList := contestsBeforeCurrentDay(cardListMap, tsList, ts); len(tmpList) > 0 {
			tmpList = sortContestList(tmpList, contestSortByEndTime)
			if len(tmpList) < needLen {
				needLen = len(tmpList)
			}

			tmpList = tmpList[0:needLen]
			list = append(list, tmpList...)
		}
	}

	return sortContestList(list, contestSortByStatus)
}

func (s *Service) rebuildMore(ctx context.Context, mid int64, list []*model.ContestCard) []*model.ContestCard {
	contestIDList := make([]int64, 0)
	for _, v := range list {
		if v.Contest.Status == model.ContestStatusOfNotStart {
			contestIDList = append(contestIDList, v.Contest.ID)
		}
	}

	if mid > 0 && len(contestIDList) > 0 {
		favorites := s.fetchFavoriteMap(ctx, mid, contestIDList)
		guessMap := s.fetchContestGuessMap(ctx, mid, contestIDList)
		if len(favorites) > 0 || len(guessMap) > 0 {
			for _, card := range list {
				card.ResetMore(favorites, guessMap)
			}
		}
	}

	return list
}

func (s *Service) rebuildContestSeriesMore(ctx context.Context, mid int64, list []*model.ContestSeries) []*model.ContestSeries {
	contestIDList := make([]int64, 0)
	for _, v := range list {
		for _, cardList4Live := range v.Detail {
			for _, card := range cardList4Live.CardList {
				if card.Contest.Status == model.ContestStatusOfNotStart {
					contestIDList = append(contestIDList, card.Contest.ID)
				}
			}
		}
	}

	if mid > 0 && len(contestIDList) > 0 {
		favorites := s.fetchFavoriteMap(ctx, mid, contestIDList)
		guessMap := s.fetchContestGuessMap(ctx, mid, contestIDList)

		if len(favorites) > 0 || len(guessMap) > 0 {
			for _, v := range list {
				for _, cardList4Live := range v.Detail {
					for _, card := range cardList4Live.CardList {
						card.ResetMore(favorites, guessMap)
					}
				}
			}
		}
	}

	return list
}

func (s *Service) fetchFavoriteMap(ctx context.Context, mid int64, contestIDList []int64) map[int64]bool {
	favorites := make(map[int64]bool, 0)
	if len(contestIDList) > 0 {
		req := new(favApi.IsFavoredsReq)
		{
			req.Typ = int32(favmdl.TypeEsports)
			req.Mid = mid
			req.Oids = contestIDList
		}

		if res, err := client.FavoriteRpcCalling(
			ctx,
			client.Path4FavoriteOfIsFavoreds,
			client.FavoriteSvrIsFavoreds,
			req); err == nil && res != nil {
			favorites = res.(*favApi.IsFavoredsReply).Faveds
		}
	}

	return favorites
}

func (s *Service) fetchContestGuessMap(ctx context.Context, mid int64, contestIDList []int64) map[int64]bool {
	m := make(map[int64]bool, 0)
	if len(contestIDList) > 0 {
		req := new(actApi.HasUserPredictReq)
		{
			req.Mid = mid
			req.ContestIds = contestIDList
		}
		if res, err := s.actClient.HasUserPredict(ctx, req); err == nil && res != nil { // 这个grpc 方法只支持LOL,从缓存中取的，要重新写一个方法处理
			m = res.Records
		}
	}

	return m
}

func sortContestList(list []*model.ContestCard, sortType int) []*model.ContestCard {
	sort.SliceStable(list, func(i, j int) bool {
		switch sortType {
		case contestSortByEndTime:
			return list[i].Contest.EndTime > list[j].Contest.EndTime
		case contestSortByStatus:
			switch list[i].Contest.Status {
			case model.ContestStatusOfOngoing:
				switch list[j].Contest.Status {
				case model.ContestStatusOfEnd:
					return true
				default:
					return list[i].Contest.StartTime < list[j].Contest.StartTime
				}
			case model.ContestStatusOfNotStart:
				switch list[j].Contest.Status {
				case model.ContestStatusOfOngoing:
					return false
				case model.ContestStatusOfEnd:
					return true
				default:
					return list[i].Contest.StartTime < list[j].Contest.StartTime
				}
			case model.ContestStatusOfEnd:
				switch list[j].Contest.Status {
				case model.ContestStatusOfEnd:
					return list[i].Contest.EndTime < list[j].Contest.EndTime
				default:
					return false
				}
			}
		case contestSortByLive:
			// TODO
		}

		return list[i].Contest.StartTime < list[j].Contest.StartTime
	})

	return list
}

func contestsBeforeCurrentDay(cardListMap map[int64][]*model.ContestCard, tsList []int64, ts int64) []*model.ContestCard {
	newList := make([]*model.ContestCard, 0)

	for _, v := range tsList {
		if v < ts {
			if d, ok := cardListMap[v]; ok {
				tmpCardList := make([]*model.ContestCard, 0)
				for _, v := range d {
					tmpCardList = append(tmpCardList, deepCopyContestCard(v))
				}

				newList = append(newList, tmpCardList...)
			}
		}
	}

	newList = sortContestList(newList, contestSortByEndTime)
	if len(newList) > 3 {
		newList = newList[:3]
	}

	return newList
}

func contestsInTheFuture(cardListMap map[int64][]*model.ContestCard, tsList []int64, ts int64) []*model.ContestCard {
	newList := make([]*model.ContestCard, 0)
	var count int

	for _, v := range tsList {
		if v > ts {
			if d, ok := cardListMap[v]; ok {
				tmpCardList := make([]*model.ContestCard, 0)
				for _, v := range d {
					tmpCardList = append(tmpCardList, deepCopyContestCard(v))
				}

				newList = append(newList, tmpCardList...)
			}

			if count >= 3 {
				break
			}
		}
	}

	return sortContestList(newList, contestSortByStartTime)
}

func isTodayContestListEnd(list []*model.ContestCard) bool {
	for _, v := range list {
		if v.Contest.Status != model.ContestStatusOfEnd {
			return false
		}
	}

	return true
}
