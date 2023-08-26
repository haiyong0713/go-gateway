package component_biz

import (
	"context"
	"time"

	"go-common/library/cache/memcache"
	"go-common/library/net/trace"

	"go-gateway/app/web-svr/esports/job/component"
	"go-gateway/app/web-svr/esports/job/dao"
)

const (
	cacheKey4HotSeason     = "esport:season:hot:list"
	cacheKey4HotMatch      = "esport:match:hot:list"
	cacheKey4HotTeam       = "esport:match:hot:team"
	expiredSeconds4HotData = 86400
)

func HotDataHandler(ctx context.Context) {
	ticker := time.NewTicker(time.Second)
	defer func() {
		ticker.Stop()
	}()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			t := trace.New("esport_HotDataHandler")
			ctx4Trace := trace.NewContext(context.Background(), t)
			seasonM, err := dao.FetchAllHotSeasonList(ctx4Trace)
			if err != nil || len(seasonM) == 0 {
				break
			}

			item4Season := &memcache.Item{
				Key:        cacheKey4HotSeason,
				Object:     seasonM,
				Expiration: expiredSeconds4HotData,
				Flags:      memcache.FlagJSON,
			}
			_ = component.GlobalMC.Set(ctx4Trace, item4Season)

			seasonList := make([]int64, 0)
			for k := range seasonM {
				seasonList = append(seasonList, k)
			}
			matchMap, teamIDMap, err := dao.FetchAllHotMatchList(ctx4Trace, seasonList)
			if err == nil {
				item4Match := &memcache.Item{
					Key:        cacheKey4HotMatch,
					Object:     matchMap,
					Expiration: expiredSeconds4HotData,
					Flags:      memcache.FlagJSON,
				}
				_ = component.GlobalMC.Set(ctx4Trace, item4Match)

				setHotTeamInfoIntoCache(ctx4Trace, teamIDMap)
			}
		}
	}
}

func setHotTeamInfoIntoCache(ctx context.Context, m map[int64]int64) {
	if len(m) > 0 {
		list := make([]int64, 0)
		for k := range m {
			list = append(list, k)
		}

		if d, err := dao.FetchTeamInfoByLargeIDList(ctx, list); err == nil {
			item4Team := &memcache.Item{
				Key:        cacheKey4HotTeam,
				Object:     d,
				Expiration: expiredSeconds4HotData,
				Flags:      memcache.FlagJSON,
			}

			_ = component.GlobalMC.Set(ctx, item4Team)
		}
	}
}
