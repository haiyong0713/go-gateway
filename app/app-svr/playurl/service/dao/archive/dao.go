package archive

import (
	"context"
	"fmt"
	"runtime"

	"go-common/library/cache/credis"
	"go-common/library/ecode"
	"go-common/library/log"
	"go-common/library/sync/pipeline/fanout"

	arcrpc "go-gateway/app/app-svr/archive/service/api"
	"go-gateway/app/app-svr/playurl/service/conf"
	"go-gateway/app/app-svr/playurl/service/model/archive"
	steinsgrpc "go-gateway/app/app-svr/steins-gate/service/api"

	favgrpc "git.bilibili.co/bapis/bapis-go/community/service/favorite"
)

const (
	_typeVideo = 2
)

// Dao is archive dao.
type Dao struct {
	c *conf.Config
	// rpc
	arcRPC    arcrpc.ArchiveClient
	favClient favgrpc.FavoriteClient
	// cache
	cache *fanout.Fanout
	// steins grpc
	steinsClient steinsgrpc.SteinsGateClient
	// redis
	arcRedis credis.Redis
}

// New new a archive dao.
func New(c *conf.Config) (d *Dao) {
	d = &Dao{
		c: c,
		// cache
		cache:    fanout.New("cache", fanout.Worker(runtime.NumCPU()), fanout.Buffer(1024)),
		arcRedis: credis.NewRedis(c.Redis.ArcRedis), //与云控播放二期代码存在冲突
	}
	var err error
	d.arcRPC, err = arcrpc.NewClient(c.ArchiveClient)
	if err != nil {
		panic(fmt.Sprintf("archive NewClient error(%+v)", err))
	}
	d.steinsClient, err = steinsgrpc.NewClient(c.SteinsClient)
	if err != nil {
		panic(fmt.Sprintf("steinsClient NewClient error(%+v)", err))
	}
	d.favClient, err = favgrpc.NewClient(c.FavClient)
	if err != nil {
		panic(fmt.Sprintf("favClient newClient(%+v)", err))
	}
	return
}

// SteinsGraphRights def.
func (d *Dao) SteinsGraphRights(c context.Context, mobiApp, device string, build int32, aid int64) (allowPlay bool, err error) {
	var (
		reply *steinsgrpc.GraphRightsReply
		req   = &steinsgrpc.GraphRightsReq{
			Aid:     aid,
			Build:   build,
			Device:  device,
			MobiApp: mobiApp,
		}
	)
	if reply, err = d.steinsClient.GraphRights(c, req); err != nil {
		log.Error("SteinsGateRights req %+v, err %v", req, err)
		return
	}
	if reply == nil {
		log.Error("SteinsGateRights req %+v, reply nil", req)
		return // allowPlay = false, err = nil
	}
	allowPlay = reply.AllowPlay
	return
}

// SimpleArc
func (d *Dao) GetSimpleArc(c context.Context, aid, mid int64, mobiApp, device, platform string) (*archive.Info, error) {
	simpleArcReq := arcrpc.SimpleArcRequest{
		Aid:      aid,
		Mid:      mid,
		MobiApp:  mobiApp,
		Device:   device,
		Platform: platform,
	}
	simpleArcRes, err := d.arcRPC.SimpleArc(c, &simpleArcReq)
	if err != nil {
		log.Error("d.arcRPC.SimpleArc(%d) error(%+v)", aid, err)
		return nil, err
	}
	if simpleArcRes.Arc == nil {
		log.Error(" d.arcRPC.SimpleArc res SimpleArcRes.Arc is nil(%d)", aid)
		return nil, ecode.NothingFound
	}
	arc := archive.Info{}
	arc.Aid = simpleArcRes.Arc.Aid
	arc.State = simpleArcRes.Arc.State
	arc.Mid = simpleArcRes.Arc.Mid
	arc.Cids = simpleArcRes.Arc.Cids
	arc.Attribute = simpleArcRes.Arc.Attribute
	arc.AttributeV2 = simpleArcRes.Arc.AttributeV2
	arc.SeasonID = simpleArcRes.Arc.SeasonId
	arc.Copyright = simpleArcRes.Arc.Copyright
	arc.TypeID = simpleArcRes.Arc.TypeId
	arc.Duration = simpleArcRes.Arc.Duration
	arc.Premiere = simpleArcRes.Arc.Premiere
	arc.Pay = simpleArcRes.Arc.Pay
	return &arc, nil
}

// Creators is
func (d *Dao) Creators(c context.Context, aid int64) ([]int64, error) {
	reply, err := d.arcRPC.Creators(c, &arcrpc.CreatorsRequest{Aids: []int64{aid}})
	if err != nil {
		return nil, err
	}
	var staffs []int64
	if cs, ok := reply.GetInfo()[aid]; ok {
		for _, s := range cs.GetStaff() {
			staffs = append(staffs, s.Mid)
		}
	}
	return staffs, nil
}

// IsFav is
func (d *Dao) IsFav(c context.Context, aid, mid int64) (bool, error) {
	reply, err := d.favClient.IsFavored(c, &favgrpc.IsFavoredReq{Typ: _typeVideo, Mid: mid, Oid: aid})
	if err != nil {
		return false, err
	}
	if reply == nil {
		return false, ecode.ServerErr
	}
	return reply.Faved, nil
}

// Close close resource.
func (d *Dao) Close() {
	d.cache.Close()
	d.arcRedis.Close()
}
