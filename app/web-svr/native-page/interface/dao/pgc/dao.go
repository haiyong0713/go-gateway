package pgc

import (
	"context"

	pgcAppGrpc "git.bilibili.co/bapis/bapis-go/pgc/service/card/app"
	"go-common/library/ecode"
	"go-common/library/log"

	"go-gateway/app/web-svr/native-page/interface/conf"
)

type Dao struct {
	pgcAppClient pgcAppGrpc.AppCardClient
}

func New(c *conf.Config) (d *Dao) {
	d = &Dao{}
	var err error
	if d.pgcAppClient, err = pgcAppGrpc.NewClient(c.PgcClient); err != nil {
		panic(err)
	}
	return
}

// SeasonBySeasonId .
func (d *Dao) SeasonBySeasonId(c context.Context, ids []int32, mid int64) (map[int32]*pgcAppGrpc.SeasonCardInfoProto, error) {
	rly, err := d.pgcAppClient.SeasonBySeasonId(c, &pgcAppGrpc.SeasonBySeasonIdReq{SeasonIds: ids, User: &pgcAppGrpc.UserReq{Mid: mid}})
	if err != nil {
		return nil, err
	}
	if rly == nil {
		return nil, ecode.NothingFound
	}
	res := make(map[int32]*pgcAppGrpc.SeasonCardInfoProto)
	for _, v := range rly.SeasonInfos {
		if v == nil {
			continue
		}
		res[v.SeasonId] = v
	}
	return res, nil
}

// SeasonByPlayId .
func (d *Dao) SeasonByPlayId(c context.Context, fid, offset, ps int32, mid int64) (*pgcAppGrpc.SeasonByPlayIdReply, error) {
	rly, err := d.pgcAppClient.SeasonByPlayId(c, &pgcAppGrpc.SeasonByPlayIdReq{PlaylistId: fid, Offset: offset, PageSize: ps, User: &pgcAppGrpc.UserReq{Mid: mid}})
	if err != nil {
		return nil, err
	}
	if rly == nil {
		return nil, ecode.NothingFound
	}
	return rly, nil
}

func (d *Dao) QueryWid(c context.Context, wid int32, mid int64) ([]*pgcAppGrpc.QueryWidItem, error) {
	req := &pgcAppGrpc.QueryWidReq{Wid: wid, User: &pgcAppGrpc.UserReq{Mid: mid}}
	rly, err := d.pgcAppClient.QueryWid(c, req)
	if err != nil {
		log.Errorc(c, "Fail to query wid, req=%+v error=%+v", req, err)
		return nil, err
	}
	return rly.GetItems(), nil
}
