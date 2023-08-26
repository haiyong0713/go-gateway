package archive

import (
	"context"

	"git.bilibili.co/go-tool/libbdevice/pkg/pd"

	"go-common/component/metadata/device"
	"go-common/library/cache/redis"
	"go-common/library/ecode"
	"go-common/library/log"
	"go-common/library/naming/discovery"
	bm "go-common/library/net/http/blademaster"
	"go-common/library/net/http/blademaster/resolver"

	viewApi "go-gateway/app/app-svr/app-view/interface/api/view"
	"go-gateway/app/app-svr/app-view/interface/conf"
	arcmid "go-gateway/app/app-svr/archive/middleware"
	"go-gateway/app/app-svr/archive/service/api"

	upgrpc "git.bilibili.co/bapis/bapis-go/up-archive/service"

	hisgrpc "git.bilibili.co/bapis/bapis-go/community/interface/history"
	creativerpc "git.bilibili.co/bapis/bapis-go/creative/open/service"

	"github.com/pkg/errors"
	"github.com/thoas/go-funk"
)

// Dao is archive dao.
type Dao struct {
	// http client
	client        *bm.Client
	httpClient    *bm.Client
	httpAiClient  *bm.Client
	realteURL     string
	commercialURL string
	relateRecURL  string
	biJianURL     string
	conf          *conf.Config
	// grpc
	arcGRPC        api.ArchiveClient
	creativeClient creativerpc.CreativeClient
	hisGRPC        hisgrpc.HistoryClient
	upArcGRPC      upgrpc.UpArchiveClient
	// redis
	redis *redis.Pool
}

// New new a archive dao.
func New(c *conf.Config) (d *Dao) {
	d = &Dao{
		// http client
		client:        bm.NewClient(c.HTTPWrite, bm.SetResolver(resolver.New(nil, discovery.Builder()))),
		httpClient:    bm.NewClient(c.HTTPClient, bm.SetResolver(resolver.New(nil, discovery.Builder()))),
		httpAiClient:  bm.NewClient(c.HTTPAiClient, bm.SetResolver(resolver.New(nil, discovery.Builder()))),
		realteURL:     c.Host.Data + _realteURL,
		commercialURL: c.Host.APICo + _commercialURL,
		relateRecURL:  c.HostDiscovery.Data + _relateRecURL,
		biJianURL:     c.Host.APICo + _biJianURL,
		redis:         redis.NewPool(c.Redis.PlayerRedis),
		conf:          c,
	}
	var err error
	if d.creativeClient, err = creativerpc.NewClient(c.CreativeClient); err != nil {
		panic(err)
	}
	if d.arcGRPC, err = api.NewClient(c.ArcGRPC); err != nil {
		panic(err)
	}
	if d.hisGRPC, err = hisgrpc.NewClient(c.HisClient); err != nil {
		panic(err)
	}
	if d.upArcGRPC, err = upgrpc.NewClient(c.UpArcGRPC); err != nil {
		panic(err)
	}
	return
}

// Archives multi get archives.
func (d *Dao) Archives(c context.Context, aids []int64, mid int64, mobiApp, device string) (as map[int64]*api.Arc, err error) {
	if len(aids) == 0 {
		return
	}
	arcsReply, err := d.arcGRPC.Arcs(c, &api.ArcsRequest{Aids: aids, Mid: mid, MobiApp: mobiApp, Device: device})
	if err != nil {
		log.Error("%+v", err)
		return
	}
	as = arcsReply.GetArcs()
	return
}

func (d *Dao) Shot(c context.Context, aid, cid int64, plat int8, build int64, dev device.Device, model string) (shot *api.VideoShot, err error) {
	arg := &api.VideoShotRequest{
		Aid: aid,
		Cid: cid,
		Common: &api.CommonParam{
			Platform: dev.RawPlatform,
			Build:    int32(dev.Build),
			Buvid:    dev.Buvid,
			MobiApp:  dev.RawMobiApp,
			Device:   dev.Device,
		},
	}
	var vs *api.VideoShotReply
	if vs, err = d.arcGRPC.VideoShot(c, arg); err != nil {
		err = errors.Wrapf(err, "%v", arg)
		return
	}
	if pd.WithContext(c).Where(func(pd *pd.PDContext) {
		pd.IsPlatIPadHD().And().Build(">=", int64(d.conf.BuildLimit.ShotIPadHDBuild))
	}).OrWhere(func(pd *pd.PDContext) {
		pd.IsPlatIPad().And().Build(">=", int64(d.conf.BuildLimit.ShotIPadBuild))
	}).OrWhere(func(pd *pd.PDContext) {
		pd.IsPlatAndroidHD().And().Build(">=", int64(d.conf.BuildLimit.ShotIPadBuild))
	}).MustFinish() {
		//判断是否在黑名单中
		if !funk.ContainsString(d.conf.Custom.ShotBlackModel, model) {
			shot = vs.GetHdVs()
		}
		//兜底返回
		if shot == nil {
			shot = vs.GetVs()
		}
		return
	}
	shot = vs.GetVs()
	return
}

// Progress is  archive plays progress .
func (d *Dao) Progress(c context.Context, aid, mid int64, buvid string) (h *viewApi.History, err error) {
	arg := &hisgrpc.ProgressReq{Mid: mid, Aids: []int64{aid}, Buvid: buvid}
	his, err := d.hisGRPC.Progress(c, arg)
	if err != nil {
		log.Error("d.hisGRPC.Progress(%v) error(%v)", arg, err)
		return
	}
	if his != nil {
		if resVal, ok := his.Res[aid]; ok && resVal != nil {
			h = &viewApi.History{Cid: resVal.Cid, Progress: resVal.Pro}
		}
	}
	return
}

func (d *Dao) Archive(c context.Context, aid int64) (a *api.Arc, err error) {
	arg := &api.ArcRequest{Aid: aid}
	reply, err := d.arcGRPC.Arc(c, arg)
	if err != nil {
		log.Error("d.arcGRPC.Arc(%v) error(%+v)", arg, err)
		return nil, err
	}
	return reply.GetArc(), nil
}

func (d *Dao) ArcsPlayer(c context.Context, arcsPlayAv []*api.PlayAv) (map[int64]*api.ArcPlayer, error) {
	batchArg, _ := arcmid.FromContext(c)
	req := api.ArcsPlayerRequest{
		BatchPlayArg: batchArg,
		PlayAvs:      arcsPlayAv,
	}
	info, err := d.arcGRPC.ArcsPlayer(c, &req)
	if err != nil {
		log.Error("ArcsPlayer(%v) error(%+v)", req, err)
		return nil, err
	}
	if info == nil {
		return nil, ecode.NothingFound
	}
	return info.ArcsPlayer, nil
}

func (d *Dao) SimpleArc(c context.Context, aid int64) (*api.SimpleArc, error) {
	arg := &api.SimpleArcRequest{Aid: aid}
	reply, err := d.arcGRPC.SimpleArc(c, arg)
	if err != nil {
		log.Error("d.arcGRPC.Arc(%v) error(%+v)", arg, err)
		return nil, err
	}
	if reply.Arc == nil {
		return nil, ecode.NothingFound
	}
	return reply.GetArc(), nil
}

func (d *Dao) UpArcCount(c context.Context, mid int64) (int64, error) {
	req := &upgrpc.ArcPassedTotalReq{Mid: mid, WithoutStaff: false}
	reply, err := d.upArcGRPC.ArcPassedTotal(c, req)
	if err != nil {
		log.Error("d.upArcGRPC.ArcPassedTotal(%v) error(%+v)", req, err)
		return 0, err
	}
	if reply == nil {
		return 0, ecode.NothingFound
	}
	return reply.Total, nil
}

// Progress is  archive plays progress .
func (d *Dao) BatchProgress(c context.Context, aids []int64, mid int64, buvid string) (map[int64]*hisgrpc.ModelHistory, error) {
	arg := &hisgrpc.ProgressReq{Mid: mid, Aids: aids, Buvid: buvid}
	his, err := d.hisGRPC.Progress(c, arg)
	if err != nil {
		log.Error("d.hisGRPC.Progress(%v) error(%v)", arg, err)
		return nil, err
	}
	return his.Res, nil
}

func (d *Dao) UpArchiveList(c context.Context, req *upgrpc.ArcPassedReq) (*upgrpc.ArcPassedReply, error) {
	reply, err := d.upArcGRPC.ArcPassed(c, req)
	if err != nil {
		return nil, err
	}
	return reply, nil
}

func (d *Dao) Types(c context.Context) (a *api.TypesReply, err error) {
	arg := &api.NoArgRequest{}
	reply, err := d.arcGRPC.Types(c, arg)
	if err != nil {
		log.Error("d.arcGRPC.Types(%v) error(%+v)", arg, err)
		return nil, err
	}
	return reply, nil
}
