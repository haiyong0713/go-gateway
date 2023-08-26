package bangumi

import (
	"context"

	"go-common/library/log"
	"go-common/library/net/metadata"
	"go-common/library/net/rpc/warden"

	pgcsearch "git.bilibili.co/bapis/bapis-go/pgc/service/card/search/v1"
	episodegrpc "git.bilibili.co/bapis/bapis-go/pgc/service/season/episode"
	seasongrpc "git.bilibili.co/bapis/bapis-go/pgc/service/season/season"

	"github.com/pkg/errors"
	"google.golang.org/grpc"
)

// appID ...
const (
	appID = "pgc.service.card"
)

// newClient new xfansmedal grpc client
func newClient(cfg *warden.ClientConfig, opts ...grpc.DialOption) (pgcsearch.SearchClient, error) {
	client := warden.NewClient(cfg, opts...)
	conn, err := client.Dial(context.Background(), "discovery://default/"+appID)
	if err != nil {
		return nil, err
	}
	return pgcsearch.NewSearchClient(conn), nil
}

// CardsInfoReply pgc cards info
func (d *Dao) CardsInfoReply(c context.Context, seasonIds []int32) (res map[int32]*seasongrpc.CardInfoProto, err error) {
	arg := &seasongrpc.SeasonInfoReq{SeasonIds: seasonIds}
	info, err := d.rpcClient.Cards(c, arg)
	if err != nil {
		err = errors.Wrapf(err, "%v", arg)
		return
	}
	res = info.Cards
	return
}

func (d *Dao) CardsByEpisodeIds(c context.Context, episodeIds []int32) (res map[int32]*episodegrpc.EpisodeCardsProto, err error) {
	arg := &episodegrpc.EpReq{Epids: episodeIds}
	info, err := d.episodegClient.Cards(c, arg)
	if err != nil {
		err = errors.Wrapf(err, "%v", arg)
		return
	}
	res = info.Cards
	return
}

func (d *Dao) CardsByAids(c context.Context, aids []int32) (res map[int32]*episodegrpc.EpisodeCardsProto, err error) {
	arg := &episodegrpc.EpAidReq{Aids: aids}
	info, err := d.episodegClient.CardsByAids(c, arg)
	if err != nil {
		err = errors.Wrapf(err, "%v", arg)
		return
	}
	res = info.Cards
	return
}

// SearchPGCCards returns the pgc cards dedicated to highlight the episodes
func (d *Dao) SearchPGCCards(ctx context.Context, seps []*pgcsearch.SeasonEpReq, query, mobiApp, device, platform string, mid int64, fnver, fnval, qn, fourk, build int, isWithPlayURL bool) (result map[int32]*pgcsearch.SearchCardProto, medias map[int32]*pgcsearch.SearchMediaProto, err error) {
	var (
		reply *pgcsearch.SearchCardReply
		req   = &pgcsearch.SearchCardReq{
			Seps:  seps,
			Query: query,
			User: &pgcsearch.UserReq{
				Mid:      mid,
				MobiApp:  mobiApp,
				Device:   device,
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
