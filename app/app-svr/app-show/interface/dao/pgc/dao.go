package pgc

import (
	"context"
	"fmt"

	pgcAppGrpc "git.bilibili.co/bapis/bapis-go/pgc/service/card/app"
	pgcClient "git.bilibili.co/bapis/bapis-go/pgc/service/follow"
	"go-common/library/ecode"
	"go-common/library/log"

	"go-gateway/app/app-svr/app-show/interface/conf"
)

// Dao is rpc dao.
type Dao struct {
	// grpc
	pgcGRPC    pgcClient.FollowClient
	pgcAppGRPC pgcAppGrpc.AppCardClient
}

// New new a pgc dao.
func New(c *conf.Config) (d *Dao) {
	d = &Dao{}
	var err error
	if d.pgcGRPC, err = pgcClient.NewClient(c.PgcFollowGRPC); err != nil {
		panic(fmt.Sprintf("pgcClient NewClientt error (%+v)", err))
	}
	if d.pgcAppGRPC, err = pgcAppGrpc.NewClient(c.PgcAppGRPC); err != nil {
		panic(fmt.Sprintf("pgcAppGrpc NewClientt error (%+v)", err))
	}
	return
}

// StatusByMid .
func (d *Dao) StatusByMid(c context.Context, mid int64, SeasonIDs []int32) (res map[int32]*pgcClient.FollowStatusProto, err error) {
	var (
		rly *pgcClient.FollowStatusByMidReply
	)
	if rly, err = d.pgcGRPC.StatusByMid(c, &pgcClient.FollowStatusByMidReq{Mid: mid, SeasonId: SeasonIDs}); err != nil {
		log.Error("d.pgcGRPC.StatusByMid(%d,%v) error(%v)", mid, SeasonIDs, err)
		return
	}
	if rly != nil {
		res = rly.Result
	}
	return
}

// AddFollow .
func (d *Dao) AddFollow(c context.Context, SeasonID int32, mid int64) (err error) {
	if _, err = d.pgcGRPC.AddFollow(c, &pgcClient.FollowReq{SeasonId: SeasonID, Mid: mid}); err != nil {
		log.Error("d.pgcGRPC.AddFollow(%d,%d) error(%v)", mid, SeasonID, err)
	}
	return
}

// AddFollow .
func (d *Dao) DeleteFollow(c context.Context, SeasonID int32, mid int64) (err error) {
	if _, err = d.pgcGRPC.DeleteFollow(c, &pgcClient.FollowReq{SeasonId: SeasonID, Mid: mid}); err != nil {
		log.Error("d.pgcGRPC.DeleteFollow(%d,%d) error(%v)", mid, SeasonID, err)
	}
	return
}

// SeasonBySeasonId .
func (d *Dao) SeasonBySeasonId(c context.Context, ids []int32, mid int64) (map[int32]*pgcAppGrpc.SeasonCardInfoProto, error) {
	rly, err := d.pgcAppGRPC.SeasonBySeasonId(c, &pgcAppGrpc.SeasonBySeasonIdReq{SeasonIds: ids, User: &pgcAppGrpc.UserReq{Mid: mid}})
	if err != nil {
		return nil, err
	}
	if rly == nil {
		return make(map[int32]*pgcAppGrpc.SeasonCardInfoProto), nil
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
	rly, err := d.pgcAppGRPC.SeasonByPlayId(c, &pgcAppGrpc.SeasonByPlayIdReq{PlaylistId: fid, Offset: offset, PageSize: ps, User: &pgcAppGrpc.UserReq{Mid: mid}})
	if err != nil {
		return nil, err
	}
	if rly == nil {
		return nil, ecode.NothingFound
	}
	return rly, nil
}

func (d *Dao) QueryWid(c context.Context, wid int32, mid int64, mobiApp, device, platform string, build int32) ([]*pgcAppGrpc.QueryWidItem, error) {
	userReq := &pgcAppGrpc.UserReq{
		Mid:      mid,
		MobiApp:  mobiApp,
		Device:   device,
		Platform: platform,
		Build:    build,
	}
	req := &pgcAppGrpc.QueryWidReq{Wid: wid, User: userReq}
	rly, err := d.pgcAppGRPC.QueryWid(c, req)
	if err != nil {
		log.Errorc(c, "Fail to query wid, req=%+v error=%+v", req, err)
		return nil, err
	}
	return rly.GetItems(), nil
}
