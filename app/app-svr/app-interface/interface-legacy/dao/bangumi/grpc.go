package bangumi

import (
	"context"

	"go-common/component/metadata/device"
	"go-common/library/log"
	"go-common/library/net/metadata"
	"go-common/library/net/rpc/warden"

	arcmid "go-gateway/app/app-svr/archive/middleware"

	pgccardgrpc "git.bilibili.co/bapis/bapis-go/pgc/service/card"
	pgcinline "git.bilibili.co/bapis/bapis-go/pgc/service/card/inline"
	pgcsearch "git.bilibili.co/bapis/bapis-go/pgc/service/card/search/v1"
	episodegrpc "git.bilibili.co/bapis/bapis-go/pgc/service/season/episode"
	seasongrpc "git.bilibili.co/bapis/bapis-go/pgc/service/season/season"
	pgcstat "git.bilibili.co/bapis/bapis-go/pgc/service/stat/v1"

	"github.com/pkg/errors"
	"google.golang.org/grpc"
)

// appID ...
const (
	appID     = "pgc.service.card"
	statAppID = "pgc.stat.service"
)

// newClient new xfansmedal grpc client
func newClient(cfg *warden.ClientConfig, opts ...grpc.DialOption) (pgcsearch.SearchClient, pgcinline.InlineCardClient, error) {
	client := warden.NewClient(cfg, opts...)
	conn, err := client.Dial(context.Background(), "discovery://default/"+appID)
	if err != nil {
		return nil, nil, err
	}
	return pgcsearch.NewSearchClient(conn), pgcinline.NewInlineCardClient(conn), nil
}

// newStatClient new xfansmedal grpc client
func newStatClient(cfg *warden.ClientConfig, opts ...grpc.DialOption) (pgcstat.StatServiceClient, error) {
	client := warden.NewClient(cfg, opts...)
	conn, err := client.Dial(context.Background(), "discovery://default/"+statAppID)
	if err != nil {
		return nil, err
	}
	return pgcstat.NewStatServiceClient(conn), nil
}

// Cards get bangumis.
func (d *Dao) Cards(ctx context.Context, seasonIds []int32) (res map[int32]*seasongrpc.CardInfoProto, err error) {
	arg := &seasongrpc.SeasonInfoReq{
		SeasonIds: seasonIds,
	}
	info, err := d.rpcClient.Cards(ctx, arg)
	if err != nil {
		log.Error("d.rpcClient.Cards error(%v)", err)
		return nil, err
	}
	res = info.Cards
	return
}

// SearchPGCCards returns the pgc cards dedicated to highlight the episodes
func (d *Dao) SearchPGCCards(ctx context.Context, seps []*pgcsearch.SeasonEpReq, query, mobiApp, device_, platform string, mid int64, fnver, fnval, qn, fourk, build int64, isWithPlayURL bool) (result map[int32]*pgcsearch.SearchCardProto, medias map[int32]*pgcsearch.SearchMediaProto, err error) {
	var (
		reply *pgcsearch.SearchCardReply
		req   = &pgcsearch.SearchCardReq{
			Seps:  seps,
			Query: query,
			User: &pgcsearch.UserReq{
				Mid:      mid,
				MobiApp:  mobiApp,
				Device:   device_,
				Platform: platform,
				Ip:       metadata.String(ctx, metadata.RemoteIP),
				Fnver:    uint32(fnver),
				Fnval:    uint32(fnval),
				Qn:       uint32(qn),
				Build:    int32(build),
				Fourk:    int32(fourk),
			},
		}
	)

	dev, ok := device.FromContext(ctx)
	if ok {
		req.User.Buvid = dev.Buvid
	}
	bpa, ok := arcmid.FromContext(ctx)
	if ok {
		req.User.NetType = pgccardgrpc.NetworkType(bpa.NetType)
		req.User.TfType = pgccardgrpc.TFType(bpa.TfType)
	}
	if !isWithPlayURL {
		req.User.LimitPlay = true
	}
	if reply, err = d.pgcsearchClient.Card(ctx, req); err != nil {
		log.Error("SearchCards Query %s, Mid %d, Err %v", query, mid, err)
		return
	}
	result = reply.Cards
	medias = reply.Medias
	return
}

// SeasonsStatGRPC pgc seasons stat
func (d *Dao) SeasonsStatGRPC(ctx context.Context, seasonIds []int32) (result map[int32]*pgcstat.SeasonStatProto, err error) {
	var (
		req   = &pgcstat.SeasonStatsReq{SeasonIds: seasonIds}
		reply *pgcstat.SeasonStatsReply
	)
	if reply, err = d.pgcstatClient.Seasons(ctx, req); err != nil {
		log.Error("SeasonsStatGRPC seasons error(%v)", err)
		return
	}
	if reply != nil {
		result = reply.Infos
	}
	return
}

// EpCards pgc epid Cards
func (d *Dao) EpCards(ctx context.Context, epids []int32) (res map[int32]*episodegrpc.EpisodeCardsProto, err error) {
	var (
		req   = &episodegrpc.EpReq{Epids: epids}
		reply *episodegrpc.EpisodeCardsReply
	)
	if reply, err = d.pgcepClient.Cards(ctx, req); err != nil {
		log.Error("EpCards episode error(%v)", err)
		return
	}
	if reply != nil {
		res = reply.Cards
	}
	return
}

func (d *Dao) SugOGV(c context.Context, ssids []int32) (res map[int32]*pgcsearch.SearchCardProto, err error) {
	var (
		args   = &pgcsearch.SugReq{SeasonIds: ssids}
		resTmp *pgcsearch.SearchCardReply
	)
	if resTmp, err = d.pgcsearchClient.Sug(c, args); err != nil {
		log.Error("%v", err)
		return
	}
	res = resTmp.GetCards()
	return
}

func (d *Dao) InlineCards(c context.Context, epIDs []int32, mobiApp, platform, device string, build int, mid int64) (map[int32]*pgcinline.EpisodeCard, error) {
	batchArg, _ := arcmid.FromContext(c)
	arg := &pgcinline.EpReq{
		EpIds: epIDs,
		User: &pgcinline.UserReq{
			Mid:      mid,
			MobiApp:  mobiApp,
			Device:   device,
			Platform: platform,
			Ip:       metadata.String(c, metadata.RemoteIP),
			Fnver:    uint32(batchArg.Fnver),
			Fnval:    uint32(batchArg.Fnval),
			Qn:       uint32(batchArg.Qn),
			Build:    int32(build),
			Fourk:    int32(batchArg.Fourk),
			NetType:  pgccardgrpc.NetworkType(batchArg.NetType),
			TfType:   pgccardgrpc.TFType(batchArg.TfType),
		},
	}
	info, err := d.pgcinlineClient.EpCard(c, arg)
	if err != nil {
		log.Error("pgc inline error(%v) arg(%v)", err, arg)
		return nil, err
	}
	return info.Infos, nil
}

func (d *Dao) EpCardsFromPgcByEpids(c context.Context, epIds []int32) (map[int32]*pgccardgrpc.EpisodeCard, error) {
	arg := &pgccardgrpc.EpCardsReq{
		EpId: epIds,
	}
	info, err := d.pgcCardClient.EpCards(c, arg)
	if err != nil {
		return nil, errors.Wrapf(err, "%+v", arg)
	}
	return info.Cards, nil
}

func (d *Dao) EpCardsFromPgcByAids(c context.Context, aids []int64) (map[int64]*pgccardgrpc.EpisodeCard, error) {
	arg := &pgccardgrpc.EpCardsReq{
		Aid: aids,
	}
	info, err := d.pgcCardClient.EpCards(c, arg)
	if err != nil {
		return nil, errors.Wrapf(err, "%+v", arg)
	}
	return info.AidCards, nil
}

func (d *Dao) CardsByMediaBizIds(c context.Context, bizIDs []int32) (map[int32]*seasongrpc.CardInfoProto, error) {
	rly, err := d.rpcClient.CardsByMediaBizIds(c, &seasongrpc.SeasonMediaBizIdReq{MediaBizId: bizIDs})
	if err != nil {
		return nil, err
	}
	return rly.GetCard(), nil
}
