package match_component

import (
	"context"
	"fmt"

	"go-common/library/cache/redis"
	"go-common/library/ecode"

	"go-gateway/app/web-svr/esports/interface/component"
	"go-gateway/app/web-svr/esports/interface/tool"
)

const (
	cacheKey4GuessListBySeasonID = "season:guess:list:%v:%v"
	cacheKey4SeasonGuessVersion  = "season:guess:version:%v"

	limitKey2FetchSeasonGuessVersion = "season_guess_version"

	sql2FetchGuessVersionBySeasonID = `
SELECT guess_version
FROM es_seasons
WHERE id = ?
`
	sql2IncrSeasonGuessVersion = `
UPDATE es_seasons
SET guess_version = guess_version + 1
WHERE id = ?
`
)

func IncrSeasonGuessVersion(ctx context.Context, seasonID int64) (err error) {
	_, err = component.GlobalDBOfMaster.Exec(ctx, sql2IncrSeasonGuessVersion, seasonID)

	return
}

func cacheKey4SeasonGuessVersionBySeasonID(seasonId int64) string {
	return fmt.Sprintf(cacheKey4SeasonGuessVersion, seasonId)
}

func DeleteSeasonGuessVersionBySeasonID(ctx context.Context, seasonID int64) (err error) {
	key := cacheKey4SeasonGuessVersionBySeasonID(seasonID)
	_, err = component.GlobalAutoSubCache.Do(ctx, "DEL", key)

	return
}

func FetchSeasonGuessVersionBySeasonID(ctx context.Context, seasonID int64) (version int64, err error) {
	version, err = FetchSeasonGuessVersionBySeasonIDFromCache(ctx, seasonID)
	if err != nil {
		return
	}

	if version == 0 {
		if tool.IsLimiterAllowedByUniqBizKey(limitKey2FetchSeasonGuessVersion, limitKey2FetchSeasonGuessVersion) {
			version, err = FetchSeasonGuessVersionBySeasonIDFromDB(ctx, seasonID)
			if err == nil && version > 0 {
				for i := 0; i < 3; i++ {
					_, cacheErr := component.GlobalAutoSubCache.Do(
						ctx,
						"SETEX",
						cacheKey4SeasonGuessVersionBySeasonID(seasonID),
						tool.CalculateExpiredSeconds(1),
						version)
					if cacheErr == nil {
						break
					}
				}
			}
		} else {
			err = ecode.LimitExceed
		}
	}

	return
}

func FetchSeasonGuessVersionBySeasonIDFromCache(ctx context.Context, seasonID int64) (version int64, err error) {
	version, err = redis.Int64(component.GlobalAutoSubCache.Do(ctx, "GET", cacheKey4SeasonGuessVersionBySeasonID(seasonID)))
	if err == redis.ErrNil {
		err = nil
	}

	return
}

func FetchSeasonGuessVersionBySeasonIDFromDB(ctx context.Context, seasonID int64) (version int64, err error) {
	err = component.GlobalDBOfMaster.QueryRow(ctx, sql2FetchGuessVersionBySeasonID, seasonID).Scan(&version)

	return
}
