package service

import (
	"context"
	"time"

	"go-common/library/net/trace"

	v1 "go-gateway/app/web-svr/esports/interface/api/v1"
	"go-gateway/app/web-svr/esports/interface/component"
)

const (
	cacheKey4HotSeason = "esport:season:hot:list"
	cacheKey4HotMatch  = "esport:match:hot:list"
	cacheKey4HotTeam   = "esport:match:hot:team"
)

var (
	HotSeasonInMemory       map[int64]int64
	HotMatchInMemory        map[int64][]int64
	HotMatch2SeasonInmemory map[int64]int64
	HotTeamInMemory         map[int64]*v1.Team
)

func init() {
	HotSeasonInMemory = make(map[int64]int64, 0)
	HotMatchInMemory = make(map[int64][]int64, 0)
	HotMatch2SeasonInmemory = make(map[int64]int64, 0)
	HotTeamInMemory = make(map[int64]*v1.Team, 0)
}

func StoreHotData2Memory(ctx context.Context) {
	ticker := time.NewTicker(time.Second)
	defer func() {
		ticker.Stop()
	}()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			t := trace.New("esport_StoreHotData2Memory")
			ctx4Trace := trace.NewContext(context.Background(), t)
			StoreHotSeason2Memory(ctx4Trace)
			StoreHotMatch2Memory(ctx4Trace)
			StoreHotTeam2Memory(ctx4Trace)
		}
	}
}

func StoreHotSeason2Memory(ctx context.Context) {
	tmpM := make(map[int64]int64, 0)
	if err := component.GlobalMemcached.Get(ctx, cacheKey4HotSeason).Scan(&tmpM); err == nil {
		HotSeasonInMemory = tmpM
	}
}

func StoreHotMatch2Memory(ctx context.Context) {
	tmpM := make(map[int64][]int64, 0)
	if err := component.GlobalMemcached.Get(ctx, cacheKey4HotMatch).Scan(&tmpM); err == nil {
		tmpM4Match2Season := make(map[int64]int64, 0)
		HotMatchInMemory = tmpM
		for seasonID, matchIDList := range HotMatchInMemory {
			for _, matchID := range matchIDList {
				tmpM4Match2Season[matchID] = seasonID
			}
		}
		HotMatch2SeasonInmemory = tmpM4Match2Season
	}
}

func StoreHotTeam2Memory(ctx context.Context) {
	tmpM := make(map[int64]int64, 0)
	if err := component.GlobalMemcached.Get(ctx, cacheKey4HotTeam).Scan(&tmpM); err == nil {
		tmpList := make([]int64, 0)
		for k := range tmpM {
			tmpList = append(tmpList, k)
		}

		if len(tmpList) > 0 {
			// TODO
		}
	}
}
