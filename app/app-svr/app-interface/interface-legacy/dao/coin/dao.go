package coin

import (
	"context"
	"fmt"
	"time"

	"go-common/library/log"
	errgroupv2 "go-common/library/sync/errgroup.v2"
	"go-gateway/app/app-svr/app-interface/interface-legacy/conf"
	"go-gateway/app/app-svr/archive/service/api"

	"github.com/pkg/errors"

	coinclient "git.bilibili.co/bapis/bapis-go/community/service/coin"
	pgccardgrpc "git.bilibili.co/bapis/bapis-go/pgc/service/card"
)

// Dao is coin dao
type Dao struct {
	coinClient    coinclient.CoinClient
	arcClient     api.ArchiveClient
	pgcCardClient pgccardgrpc.CardClient
}

// New initial coin dao
func New(c *conf.Config) (d *Dao) {
	d = &Dao{}
	var err error
	if d.coinClient, err = coinclient.NewClient(c.CoinClient); err != nil {
		panic(err)
	}
	if d.arcClient, err = api.NewClient(c.ArchiveGRPC); err != nil {
		panic(err)
	}
	if d.pgcCardClient, err = pgccardgrpc.NewClient(c.PGCRPC); err != nil {
		panic(fmt.Sprintf("pgccard NewClient error (%+v)", err))
	}
	return
}

// CoinList coin archive list
func (d *Dao) CoinList(c context.Context, mid int64, pn, ps int, mobiApp, device string) ([]*api.Arc, map[int64]*pgccardgrpc.EpisodeCard, int, error) {
	var (
		coinReply *coinclient.ListReply
		aids      []int64
		err       error
	)
	if coinReply, err = d.coinClient.List(c, &coinclient.ListReq{Mid: mid, Business: "archive", Ts: time.Now().Unix()}); err != nil {
		return nil, nil, 0, err
	}
	existAids := make(map[int64]int64, len(coinReply.List))
	for _, v := range coinReply.List {
		if _, ok := existAids[v.Aid]; ok {
			continue
		}
		aids = append(aids, v.Aid)
		existAids[v.Aid] = v.Aid
	}
	count := len(aids)
	start := (pn - 1) * ps
	end := pn * ps
	switch {
	case start > count:
		aids = aids[:0]
	case end >= count:
		aids = aids[start:]
	default:
		aids = aids[start:end]
	}
	coinArc := make([]*api.Arc, 0)
	if len(aids) == 0 {
		return coinArc, nil, 0, nil
	}

	eg := errgroupv2.WithContext(c)
	var arcsReply *api.ArcsReply
	eg.Go(func(ctx context.Context) (err error) {
		if arcsReply, err = d.arcClient.Arcs(ctx, &api.ArcsRequest{
			Aids:    aids,
			Mid:     mid,
			MobiApp: mobiApp,
			Device:  device,
		}); err != nil {
			return err
		}
		return nil
	})
	var eps map[int64]*pgccardgrpc.EpisodeCard
	eg.Go(func(ctx context.Context) (err error) {
		eps, err = d.EpCardsFromPgcByAids(ctx, aids)
		if err != nil {
			log.Error("%+v", err)
			return nil
		}
		return nil
	})
	if err := eg.Wait(); err != nil {
		log.Error("%+v", err)
		return coinArc, nil, 0, err
	}

	for _, aid := range aids {
		if arc, ok := arcsReply.Arcs[aid]; ok && arc.IsNormal() {
			//nolint:gomnd
			if arc.Access >= 10000 {
				arc.Stat.View = 0
			}
			coinArc = append(coinArc, arc)
		}
	}
	if ps > count {
		count = len(coinArc)
	}
	return coinArc, eps, count, nil
}

// UpMemberState .
func (d *Dao) UpMemberState(c context.Context, aid, mid int64, business string) (err error) {
	if _, err = d.coinClient.UpMemberState(c, &coinclient.UpMemberStateReq{Aid: aid, Mid: mid, Business: business}); err != nil {
		log.Error("d.coinClient.UpMemberState(%d) error(%v)", aid, err)
	}
	return
}

func (d *Dao) ArchiveUserCoins(ctx context.Context, aids []int64, mid int64) (map[int64]int64, error) {
	const _coinBizAv = "archive"

	arg := &coinclient.ItemsUserCoinsReq{
		Mid:      mid,
		Aids:     aids,
		Business: _coinBizAv,
	}
	reply, err := d.coinClient.ItemsUserCoins(ctx, arg)
	if err != nil {
		return nil, err
	}
	return reply.Numbers, nil
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
