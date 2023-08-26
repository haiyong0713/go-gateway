package match

import (
	"context"

	"go-common/library/log"

	"go-gateway/app/app-svr/app-interface/interface-legacy/conf"

	esportGRPC "git.bilibili.co/bapis/bapis-go/esports/service"
	esportsservice "git.bilibili.co/bapis/bapis-go/operational/esportsservice"

	"github.com/pkg/errors"
)

type Dao struct {
	c          *conf.Config
	esportgrpc esportGRPC.EsportsClient
	sportgrpc  esportsservice.EsportsServiceClient
}

func New(c *conf.Config) (d *Dao) {
	d = &Dao{
		c: c,
	}
	var err error
	if d.esportgrpc, err = esportGRPC.NewClient(c.ESportsGRPC); err != nil {
		panic(err)
	}
	if d.sportgrpc, err = esportsservice.NewClient(c.SportsGRPC); err != nil {
		panic(err)
	}
	return
}

func (d *Dao) AddFav(c context.Context, mid, matchID int64) (err error) {
	var args = &esportGRPC.FavRequest{Mid: mid, Cid: matchID}
	if _, err = d.esportgrpc.LiveAddFav(c, args); err != nil {
		log.Error("%v", err)
	}
	return
}

func (d *Dao) DelFav(c context.Context, mid, matchID int64) (err error) {
	var args = &esportGRPC.FavRequest{Mid: mid, Cid: matchID}
	if _, err = d.esportgrpc.LiveDelFav(c, args); err != nil {
		log.Error("%v", err)
	}
	return
}

func (d *Dao) Matchs(c context.Context, mid int64, matchIDs []int64) (res map[int64]*esportGRPC.Contest, err error) {
	var (
		args   = &esportGRPC.LiveContestsRequest{Mid: mid, Cids: matchIDs}
		matchs *esportGRPC.LiveContestsReply
	)
	if matchs, err = d.esportgrpc.LiveContests(c, args); err != nil {
		log.Error("%v", err)
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

func (d *Dao) GetSportsEventMatches(ctx context.Context, req *esportsservice.GetSportsEventMatchesReq) (res *esportsservice.GetSportsEventMatchesResponse, err error) {
	reply, err := d.sportgrpc.GetSportsEventMatches(ctx, req)
	if err != nil {
		return nil, errors.WithMessagef(err, "req=%+v", req)
	}
	return reply, nil
}
