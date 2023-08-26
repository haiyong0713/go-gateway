package dao

import (
	"context"
	"net/url"
	"strconv"
	"time"

	"go-common/library/ecode"
	bm "go-common/library/net/http/blademaster"
	"go-common/library/net/metadata"
	"go-common/library/xstr"
	"go-gateway/app/app-svr/app-card/interface/model/card/bangumi"
	"go-gateway/app/app-svr/app-feed/interface-ng/internal/model"

	pgcinline "git.bilibili.co/bapis/bapis-go/pgc/service/card/inline"
	episodegrpc "git.bilibili.co/bapis/bapis-go/pgc/service/season/episode"
	"github.com/pkg/errors"
)

const (
	_epPlayer = "/pgc/internal/dynamic/v3/ep/list"
	_updates  = "/internal_api/follow_update"
	_remind   = "/pgc/internal/dynamic/remind"
)

type bangumiConfig struct {
	Host string
}

type bangumiDao struct {
	bangumi         episodegrpc.EpisodeClient
	client          *bm.Client
	cfg             bangumiConfig
	pgcinlineClient pgcinline.InlineCardClient
}

func (d *bangumiDao) CardsInfoReply(ctx context.Context, episodeIds []int32) (map[int32]*episodegrpc.EpisodeCardsProto, error) {
	arg := &episodegrpc.EpReq{Epids: episodeIds}
	info, err := d.bangumi.Cards(ctx, arg)
	if err != nil {
		return nil, err
	}
	return info.Cards, nil
}

func (d *bangumiDao) CardsByAids(ctx context.Context, aids []int32) (map[int32]*episodegrpc.EpisodeCardsProto, error) {
	arg := &episodegrpc.EpAidReq{Aids: aids}
	info, err := d.bangumi.CardsByAids(ctx, arg)
	if err != nil {
		return nil, err
	}
	return info.Cards, nil
}

func (d *bangumiDao) EpPlayer(ctx context.Context, req *model.EpPlayerReq) (map[int64]*bangumi.EpPlayer, error) {
	ip := metadata.String(ctx, metadata.RemoteIP)
	params := url.Values{}
	params.Set("ep_ids", xstr.JoinInts(req.EpIDs))
	params.Set("mobi_app", req.MobiApp)
	params.Set("platform", req.Platform)
	params.Set("device", req.Device)
	params.Set("build", strconv.Itoa(req.Build))
	params.Set("ip", ip)
	params.Set("fnval", strconv.Itoa(req.Fnval))
	params.Set("fnver", strconv.Itoa(req.Fnver))
	var res struct {
		Code   int                         `json:"code"`
		Result map[int64]*bangumi.EpPlayer `json:"result"`
	}
	if err := d.client.Get(ctx, d.cfg.Host+_epPlayer, ip, params, &res); err != nil {
		return nil, err
	}
	if res.Code != ecode.OK.Code() {
		return nil, errors.Wrap(ecode.Int(res.Code), d.cfg.Host+_epPlayer+"?"+params.Encode())
	}
	return res.Result, nil
}

func (d *bangumiDao) InlineCards(ctx context.Context, req *pgcinline.EpReq) (map[int32]*pgcinline.EpisodeCard, error) {
	info, err := d.pgcinlineClient.EpCard(ctx, req)
	if err != nil {
		return nil, err
	}
	return info.Infos, nil
}

func (d *bangumiDao) Updates(ctx context.Context, mid int64, now time.Time) (*bangumi.Update, error) {
	ip := metadata.String(ctx, metadata.RemoteIP)
	params := url.Values{}
	params.Set("mid", strconv.FormatInt(mid, 10))
	var res struct {
		Code   int             `json:"code"`
		Result *bangumi.Update `json:"result"`
	}
	if err := d.client.Get(ctx, d.cfg.Host+_updates, ip, params, &res); err != nil {
		return nil, err
	}
	if res.Code != ecode.OK.Code() {
		return nil, errors.Wrap(ecode.Int(res.Code), d.cfg.Host+_updates+"?"+params.Encode())
	}
	return res.Result, nil
}

func (d *bangumiDao) Remind(ctx context.Context, mid int64) (*bangumi.Remind, error) {
	ip := metadata.String(ctx, metadata.RemoteIP)
	params := url.Values{}
	params.Set("mid", strconv.FormatInt(mid, 10))
	var res struct {
		Code   int             `json:"code"`
		Result *bangumi.Remind `json:"result"`
	}
	if err := d.client.Get(ctx, d.cfg.Host+_remind, ip, params, &res); err != nil {
		return nil, err
	}
	if res.Code != ecode.OK.Code() {
		return nil, errors.Wrap(ecode.Int(res.Code), d.cfg.Host+_remind+"?"+params.Encode())
	}
	return res.Result, nil
}
