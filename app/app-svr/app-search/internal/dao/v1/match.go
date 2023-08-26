package v1

import (
	"context"

	esportGRPC "git.bilibili.co/bapis/bapis-go/esports/service"
	esportsservice "git.bilibili.co/bapis/bapis-go/operational/esportsservice"

	"github.com/pkg/errors"
)

func (d *dao) Matchs(c context.Context, mid int64, matchIDs []int64) (res map[int64]*esportGRPC.Contest, err error) {
	var (
		args   = &esportGRPC.LiveContestsRequest{Mid: mid, Cids: matchIDs}
		matchs *esportGRPC.LiveContestsReply
	)
	if matchs, err = d.esportClient.LiveContests(c, args); err != nil {
		return
	}
	res = make(map[int64]*esportGRPC.Contest)
	for _, match := range matchs.GetContests() {
		if match == nil || match.ID == 0 {
			continue
		}
		res[match.ID] = match
	}
	return
}

func (d *dao) GetSportsEventMatches(ctx context.Context, req *esportsservice.GetSportsEventMatchesReq) (res *esportsservice.GetSportsEventMatchesResponse, err error) {
	reply, err := d.sportClient.GetSportsEventMatches(ctx, req)
	if err != nil {
		return nil, errors.WithMessagef(err, "req=%+v", req)
	}
	return reply, nil
}
