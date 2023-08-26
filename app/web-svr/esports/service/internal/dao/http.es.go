package dao

import (
	"context"
	"go-common/library/log"
	"go-gateway/app/web-svr/esports/service/internal/model"
)

const (
	_contestIndex    = "esports_contests02"
	_videoIndex      = "esports"
	_contestFavIndex = "esports_fav"
)

func (d *dao) SearchContestsByTime(ctx context.Context, contestsQueryParams *model.ContestsQueryParamsModel) (contestIds []int64, total int, err error) {
	res := new(model.ContestQueryResponse)
	contestIds = make([]int64, 0)
	r := d.elastic.NewRequest(_contestIndex).Index(_contestIndex)
	r.Fields("id")
	contestsQueryParams.ContestQueryStatusParams(r)
	contestsQueryParams.ContestQueryPageParams(r)
	if contestsQueryParams.Gid > 0 {
		r.WhereEq("gmap.gid", contestsQueryParams.Gid).WhereEq("gmap.type", model.OidContestType).WhereEq("gmap.is_deleted", 0)
	}
	if contestsQueryParams.MatchId > 0 {
		r.WhereEq("mid", contestsQueryParams.MatchId)
	}
	contestsQueryParams.ContestQueryTeamsParams(r)
	contestsQueryParams.ContestQueryTimeParams(r)
	contestsQueryParams.ContestQueryGuessParams(r)
	contestsQueryParams.ContestQuerySyncPlatformParams(r)
	if len(contestsQueryParams.Sids) > 0 {
		r.WhereIn("sid", contestsQueryParams.Sids)
	}
	if len(contestsQueryParams.ContestIds) > 0 {
		r.WhereIn("id", contestsQueryParams.ContestIds)
	}
	// live grpc  use.
	if len(contestsQueryParams.RoomIds) > 0 {
		r.WhereIn("live_room", contestsQueryParams.RoomIds)
	}
	if contestsQueryParams.Debug {
		log.Infoc(ctx, "[Dao][searchContestsByTime][Debug][Info], req: %+v", r.Params())
	}
	if err = r.Scan(ctx, &res); err != nil {
		log.Errorc(ctx, "[Dao][searchContestsByTime][Error], err: %+v", err)
		return
	}
	total = res.Page.Total
	for _, contest := range res.Result {
		contestIds = append(contestIds, contest.ID)
	}
	return
}
