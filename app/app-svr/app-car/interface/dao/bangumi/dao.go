package bangumi

import (
	"context"
	"fmt"
	"math"
	"sync"

	"go-common/library/ecode"
	"go-common/library/log"
	"go-common/library/naming/discovery"
	bm "go-common/library/net/http/blademaster"
	"go-common/library/net/http/blademaster/resolver"
	"go-common/library/net/metadata"
	"go-common/library/net/rpc/warden"
	errgroup "go-common/library/sync/errgroup.v2"
	"go-gateway/app/app-svr/app-car/interface/conf"
	arcmid "go-gateway/app/app-svr/archive/middleware"

	pgccard "git.bilibili.co/bapis/bapis-go/pgc/service/card"
	cardappgrpc "git.bilibili.co/bapis/bapis-go/pgc/service/card/app"
	pgcinline "git.bilibili.co/bapis/bapis-go/pgc/service/card/inline"
	followgrpc "git.bilibili.co/bapis/bapis-go/pgc/service/follow"
	episodegrpc "git.bilibili.co/bapis/bapis-go/pgc/service/season/episode"
	seasongrpc "git.bilibili.co/bapis/bapis-go/pgc/service/season/season"
	strategygrpc "git.bilibili.co/bapis/bapis-go/pgc/service/strategy"
	"google.golang.org/grpc"
)

type Dao struct {
	c *conf.Config
	// api
	client      *bm.Client
	module      string
	view        string
	playurlH5   string
	playurlProj string
	playurlApp  string
	// rpc
	rpcClient       seasongrpc.SeasonClient
	epClient        episodegrpc.EpisodeClient
	followClient    followgrpc.FollowClient
	cardappClient   cardappgrpc.AppCardClient
	pgcinlineClient pgcinline.InlineCardClient
	strategyClient  strategygrpc.StrategyClient
}

func New(c *conf.Config) *Dao {
	d := &Dao{
		c:           c,
		client:      bm.NewClient(c.HTTPPGC, bm.SetResolver(resolver.New(nil, discovery.Builder()))),
		module:      c.Host.APICo + _model,
		view:        c.Host.APICo + _view,
		playurlH5:   c.HostDiscovery.PGCPlayer + _playurlH5,
		playurlProj: c.HostDiscovery.PGCPlayer + _playurlProj,
		playurlApp:  c.HostDiscovery.PGCPlayer + _playurlApp,
	}
	var err error
	if d.rpcClient, err = seasongrpc.NewClient(c.PGCRPC); err != nil {
		panic(fmt.Sprintf("seasongrpc NewClientt error (%+v)", err))
	}
	if d.epClient, err = episodegrpc.NewClient(c.PGCRPC); err != nil {
		panic(fmt.Sprintf("pgcep NewClient error (%+v)", err))
	}
	if d.followClient, err = followgrpc.NewClient(c.PGCRPC); err != nil {
		panic(fmt.Sprintf("follow NewClient error (%+v)", err))
	}
	if d.cardappClient, err = cardappgrpc.NewClient(c.PGCRPC); err != nil {
		panic(fmt.Sprintf("cardapp NewClient error (%+v)", err))
	}
	if d.pgcinlineClient, err = pgcinline.NewClient(c.PGCRPC); err != nil {
		panic(fmt.Sprintf("pgcinline NewClientt error (%+v)", err))
	}
	if d.strategyClient, err = strategyNewClient(c.PGCRPC); err != nil {
		panic(fmt.Sprintf("strategygrpc NewClientt error (%+v)", err))
	}
	return d
}

// NewClient new a app.dynamic grpc client
func strategyNewClient(cfg *warden.ClientConfig, opts ...grpc.DialOption) (strategygrpc.StrategyClient, error) {
	const (
		_appID = "pgc.service.strategy"
	)
	client := warden.NewClient(cfg, opts...)
	conn, err := client.Dial(context.Background(), "discovery://default/"+_appID)
	if err != nil {
		return nil, err
	}
	return strategygrpc.NewStrategyClient(conn), nil
}

func (d *Dao) Cards(ctx context.Context, seasonIds []int32) (map[int32]*seasongrpc.CardInfoProto, error) {
	reply, err := d.rpcClient.Cards(ctx, &seasongrpc.SeasonInfoReq{SeasonIds: seasonIds})
	if err != nil {
		log.Error("%+v", err)
		return nil, err
	}
	return reply.GetCards(), nil
}

// EpCards pgc epid Cards
func (d *Dao) EpCards(ctx context.Context, epids []int32) (map[int32]*episodegrpc.EpisodeCardsProto, error) {
	req := &episodegrpc.EpReq{Epids: epids}
	reply, err := d.epClient.Cards(ctx, req)
	if err != nil {
		log.Error("%+v", err)
		return nil, err
	}
	return reply.GetCards(), nil
}

func (d *Dao) MyRelations(ctx context.Context, mid int64) ([]*followgrpc.FollowSeasonProto, error) {
	req := &followgrpc.MyRelationsReq{Mid: mid, WithAnime: true, WithCinema: true}
	reply, err := d.followClient.MyRelations(ctx, req)
	if err != nil {
		log.Error("%+v", err)
		return nil, err
	}
	return reply.GetRelations(), nil
}

func (d *Dao) MyFollows(c context.Context, mid int64, followType, pn, ps int) ([]*cardappgrpc.CardSeasonProto, error) {
	if mid <= 0 {
		return nil, ecode.RequestErr
	}
	req := &cardappgrpc.FollowReq{Mid: mid, Pn: int32(pn), Ps: int32(ps), FollowType: int32(followType), From: 3}
	reply, err := d.cardappClient.MyFollows(c, req)
	if err != nil {
		log.Error("%+v", err)
		return nil, err
	}
	return reply.GetSeasons(), nil
}

func (d *Dao) CardsByAids(c context.Context, aids []int64) (map[int32]*episodegrpc.EpisodeCardsProto, error) {
	var tmpAids []int32
	for _, aid := range aids {
		tmpAids = append(tmpAids, int32(aid))
	}
	req := &episodegrpc.EpAidReq{Aids: tmpAids}
	reply, err := d.epClient.CardsByAids(c, req)
	if err != nil {
		log.Error("%+v", err)
		return nil, err
	}
	return reply.GetCards(), nil
}

func (d *Dao) SeasonCardNewEp(ctx context.Context, seasonId int32) (*seasongrpc.NewEpProto, error) {
	reply, err := d.rpcClient.Cards(ctx, &seasongrpc.SeasonInfoReq{SeasonIds: []int32{seasonId}})
	if err != nil {
		log.Error("%+v", err)
		return nil, err
	}
	season, ok := reply.GetCards()[seasonId]
	if !ok {
		return nil, ecode.NothingFound
	}
	if season.NewEp == nil {
		return nil, ecode.NothingFound
	}
	return season.NewEp, nil
}

func (d *Dao) AvInfo(ctx context.Context, epid int32) (*episodegrpc.AvInfoProto, error) {
	reply, err := d.epClient.AvInfos(ctx, &episodegrpc.AvInfoReq{EpisodeId: []int32{epid}})
	if err != nil {
		log.Error("%+v", err)
		return nil, err
	}
	ep, ok := reply.GetInfo()[epid]
	if !ok {
		return nil, ecode.NothingFound
	}
	return ep, nil
}

func (d *Dao) CardsByAidsAll(c context.Context, aids []int64) (map[int32]*episodegrpc.EpisodeCardsProto, error) {
	const (
		_max = 50
	)
	var (
		forNum     = int(math.Ceil(float64(len(aids)) / float64(_max)))
		mutex      = sync.Mutex{}
		start, end int
	)
	res := map[int32]*episodegrpc.EpisodeCardsProto{}
	g := errgroup.WithContext(c)
	for i := 0; i < forNum; i++ {
		start = i * _max
		end = start + _max
		var (
			tmpaids []int64
		)
		if len(aids) >= end {
			tmpaids = aids[start:end]
		} else if len(aids) < end {
			tmpaids = aids[start:]
		} else if len(aids) < start {
			break
		}
		g.Go(func(cc context.Context) error {
			reply, err := d.CardsByAids(cc, tmpaids)
			if err != nil {
				log.Error("d.CardsByAids(%v) error(%v)", tmpaids, err)
				return err
			}
			if reply != nil {
				mutex.Lock()
				for _, v := range reply {
					res[v.Aid] = v
				}
				mutex.Unlock()
			}
			return nil
		})
	}
	if err := g.Wait(); err != nil {
		log.Error("%+v", err)
		return nil, err
	}
	return res, nil
}

func (d *Dao) CardsAll(c context.Context, seasonIds []int32) (map[int32]*seasongrpc.CardInfoProto, error) {
	const (
		_max = 50
	)
	var (
		forNum     = int(math.Ceil(float64(len(seasonIds)) / float64(_max)))
		mutex      = sync.Mutex{}
		start, end int
	)
	res := map[int32]*seasongrpc.CardInfoProto{}
	g := errgroup.WithContext(c)
	for i := 0; i < forNum; i++ {
		start = i * _max
		end = start + _max
		var (
			tmpsids []int32
		)
		if len(seasonIds) >= end {
			tmpsids = seasonIds[start:end]
		} else if len(seasonIds) < end {
			tmpsids = seasonIds[start:]
		} else if len(seasonIds) < start {
			break
		}
		g.Go(func(cc context.Context) error {
			reply, err := d.Cards(cc, tmpsids)
			if err != nil {
				log.Error("d.CardsAll(%v) error(%v)", tmpsids, err)
				return err
			}
			if reply != nil {
				mutex.Lock()
				for _, v := range reply {
					res[v.SeasonId] = v
				}
				mutex.Unlock()
			}
			return nil
		})
	}
	if err := g.Wait(); err != nil {
		log.Error("%+v", err)
		return nil, err
	}
	return res, nil
}

func (d *Dao) inlineCards(c context.Context, epIDs []int32, mobiApp, platform, device string, build int) (map[int32]*pgcinline.EpisodeCard, error) {
	batchArg, _ := arcmid.FromContext(c)
	arg := &pgcinline.EpReq{
		EpIds: epIDs,
		User: &pgcinline.UserReq{
			Mid:      batchArg.Mid,
			MobiApp:  mobiApp,
			Device:   device,
			Platform: platform,
			Ip:       metadata.String(c, metadata.RemoteIP),
			Fnver:    uint32(batchArg.Fnver),
			Fnval:    uint32(batchArg.Fnval),
			Qn:       uint32(batchArg.Qn),
			Build:    int32(build),
			Fourk:    int32(batchArg.Fourk),
			NetType:  pgccard.NetworkType(batchArg.NetType),
			TfType:   pgccard.TFType(batchArg.TfType),
		},
	}
	info, err := d.pgcinlineClient.EpCard(c, arg)
	if err != nil {
		log.Error("pgc inline error(%v) arg(%v)", err, arg)
		return nil, err
	}
	return info.Infos, nil
}

func (d *Dao) InlineCardsAll(c context.Context, epIDs []int32, mobiApp, platform, device string, build int) (map[int32]*pgcinline.EpisodeCard, error) {
	const (
		_max = 50
	)
	var (
		forNum     = int(math.Ceil(float64(len(epIDs)) / float64(_max)))
		mutex      = sync.Mutex{}
		start, end int
	)
	res := map[int32]*pgcinline.EpisodeCard{}
	g := errgroup.WithContext(c)
	for i := 0; i < forNum; i++ {
		start = i * _max
		end = start + _max
		var (
			tmpepids []int32
		)
		if len(epIDs) >= end {
			tmpepids = epIDs[start:end]
		} else if len(epIDs) < end {
			tmpepids = epIDs[start:]
		} else if len(epIDs) < start {
			break
		}
		g.Go(func(cc context.Context) error {
			reply, err := d.inlineCards(cc, tmpepids, mobiApp, platform, device, build)
			if err != nil {
				log.Error("d.inlineCards(%v) error(%v)", tmpepids, err)
				return err
			}
			if reply != nil {
				mutex.Lock()
				for _, v := range reply {
					res[v.EpisodeId] = v
				}
				mutex.Unlock()
			}
			return nil
		})
	}
	if err := g.Wait(); err != nil {
		log.Error("%+v", err)
		return nil, err
	}
	return res, nil
}

// Activity
func (d *Dao) Activity(c context.Context, mid int64, buvid string) ([]*strategygrpc.SortedResult, error) {
	const (
		_actType    = 1
		_actSubType = 2
	)
	arg := &strategygrpc.ActivityReq{
		Mid:     mid,
		Buvid:   buvid,
		Type:    _actType,
		SubType: _actSubType,
	}
	reply, err := d.strategyClient.ActivityRcmd(c, arg)
	if err != nil {
		return nil, err
	}
	return reply.GetRecommends(), nil
}
