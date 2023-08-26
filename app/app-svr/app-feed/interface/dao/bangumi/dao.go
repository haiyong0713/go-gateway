package bangumi

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
	"time"

	"go-common/library/ecode"
	httpx "go-common/library/net/http/blademaster"
	"go-common/library/net/metadata"
	"go-common/library/net/rpc/warden"
	"go-common/library/xstr"
	"go-gateway/app/app-svr/app-card/interface/model/card/bangumi"
	"go-gateway/app/app-svr/app-feed/interface/conf"
	"google.golang.org/grpc"

	feed "git.bilibili.co/bapis/bapis-go/community/service/feed"
	deliverygrpc "git.bilibili.co/bapis/bapis-go/pgc/servant/delivery"
	pgcCardClient "git.bilibili.co/bapis/bapis-go/pgc/service/card"
	pgcAppGrpc "git.bilibili.co/bapis/bapis-go/pgc/service/card/app"
	pgcinline "git.bilibili.co/bapis/bapis-go/pgc/service/card/inline"
	pgcstory "git.bilibili.co/bapis/bapis-go/pgc/service/card/story"
	pgcFollowClient "git.bilibili.co/bapis/bapis-go/pgc/service/follow"
	episodegrpc "git.bilibili.co/bapis/bapis-go/pgc/service/season/episode"
	"github.com/pkg/errors"
)

const (
	_updates     = "/internal_api/follow_update"
	_pullSeasons = "/internal_api/follow_seasons"
	_followPull  = "/pgc/internal/moe/2018/follow/pull"
	_remind      = "/pgc/internal/dynamic/remind"
	_epPlayer    = "/pgc/internal/dynamic/v3/ep/list"
)

// Dao is show dao.
type Dao struct {
	// http client
	client *httpx.Client
	// bangumi
	updates     string
	pullSeasons string
	followPull  string
	remind      string
	epPlayer    string
	// grpc
	rpcClient       episodegrpc.EpisodeClient
	pgcinlineClient pgcinline.InlineCardClient
	pgcAppClient    pgcAppGrpc.AppCardClient
	pgcCardClient   pgcCardClient.CardClient
	pgcStoryClient  pgcstory.StoryClient
	deliveryClient  deliverygrpc.DeliveryClient
	pgcFollowClient pgcFollowClient.FollowClient
}

// New new a bangumi dao.
func New(c *conf.Config) (d *Dao) {
	d = &Dao{
		// http clients
		client:      httpx.NewClient(c.HTTPBangumi),
		updates:     c.Host.Bangumi + _updates,
		pullSeasons: c.Host.Bangumi + _pullSeasons,
		followPull:  c.Host.APICo + _followPull,
		remind:      c.Host.APICo + _remind,
		epPlayer:    c.Host.APICo + _epPlayer,
	}
	var err error
	if d.rpcClient, err = episodegrpc.NewClient(c.PGCRPC); err != nil {
		panic(fmt.Sprintf("episodegrpc NewClientt error (%+v)", err))
	}
	if d.pgcinlineClient, err = pgcinline.NewClient(c.PGCInline); err != nil {
		panic(fmt.Sprintf("pgcinline NewClientt error (%+v)", err))
	}
	if d.pgcAppClient, err = pgcAppGrpc.NewClient(c.PgcClient); err != nil {
		panic(fmt.Sprintf("pgcAppGrpc NewClient error (%+v)", err))
	}
	if d.pgcCardClient, err = pgcCardClient.NewClient(c.PgcCardClient); err != nil {
		panic(fmt.Sprintf("pgcCardClient NewClient error (%+v)", err))
	}
	if d.pgcStoryClient, err = pgcstory.NewClientStory(c.PgcStoryClient); err != nil {
		panic(fmt.Sprintf("pgcStoryClient NewClient error (%+v)", err))
	}
	if d.pgcFollowClient, err = pgcFollowClient.NewClient(c.PgcFollowClient); err != nil {
		panic(fmt.Sprintf("pgcFollowClient NewClient error (%+v)", err))
	}
	if d.deliveryClient, err = NewDeliveryClient(c.DeliveryClient); err != nil {
		panic(fmt.Sprintf("deliveryClient NewClient error (%+v)", err))
	}
	return d
}

func NewDeliveryClient(cfg *warden.ClientConfig, opts ...grpc.DialOption) (deliverygrpc.DeliveryClient, error) {
	client := warden.NewClient(cfg, opts...)
	conn, err := client.Dial(context.Background(), "discovery://default/"+"ogv.operation.servant")
	if err != nil {
		return nil, err
	}
	return deliverygrpc.NewDeliveryClient(conn), nil
}

// SeasonBySeasonId acquires season card info from season ids
func (d *Dao) SeasonBySeasonId(c context.Context, ids []int32, mid int64) (map[int32]*pgcAppGrpc.SeasonCardInfoProto, error) {
	rly, err := d.pgcAppClient.SeasonBySeasonId(c, &pgcAppGrpc.SeasonBySeasonIdReq{SeasonIds: ids, User: &pgcAppGrpc.UserReq{Mid: mid}})
	if err != nil {
		return nil, err
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

func (d *Dao) Updates(c context.Context, mid int64, now time.Time) (u *bangumi.Update, err error) {
	ip := metadata.String(c, metadata.RemoteIP)
	params := url.Values{}
	params.Set("mid", strconv.FormatInt(mid, 10))
	var res struct {
		Code   int             `json:"code"`
		Result *bangumi.Update `json:"result"`
	}
	if err = d.client.Get(c, d.updates, ip, params, &res); err != nil {
		return
	}
	if res.Code != ecode.OK.Code() {
		err = errors.Wrap(ecode.Int(res.Code), d.updates+"?"+params.Encode())
		return
	}
	u = res.Result
	return
}

func (d *Dao) PullSeasons(c context.Context, seasonIDs []int64, now time.Time) (psm map[int64]*feed.Bangumi, err error) {
	ip := metadata.String(c, metadata.RemoteIP)
	params := url.Values{}
	params.Set("season_ids", xstr.JoinInts(seasonIDs))
	var res struct {
		Code   int             `json:"code"`
		Result []*feed.Bangumi `json:"result"`
	}
	if err = d.client.Get(c, d.pullSeasons, ip, params, &res); err != nil {
		return
	}
	if res.Code != ecode.OK.Code() {
		err = errors.Wrap(ecode.Int(res.Code), d.pullSeasons+"?"+params.Encode())
		return
	}
	psm = make(map[int64]*feed.Bangumi, len(res.Result))
	for _, p := range res.Result {
		psm[p.EpisodeID] = p
	}
	return
}

func (d *Dao) FollowPull(c context.Context, mid int64, mobiApp, device string, now time.Time) (moe *bangumi.Moe, err error) {
	ip := metadata.String(c, metadata.RemoteIP)
	params := url.Values{}
	params.Set("mid", strconv.FormatInt(mid, 10))
	params.Set("mobi_app", mobiApp)
	params.Set("device", device)
	var res struct {
		Code   int          `json:"code"`
		Result *bangumi.Moe `json:"result"`
	}
	if err = d.client.Get(c, d.followPull, ip, params, &res); err != nil {
		return
	}
	if res.Code != ecode.OK.Code() {
		err = errors.Wrap(ecode.Int(res.Code), d.followPull+"?"+params.Encode())
		return
	}
	moe = res.Result
	return
}

func (d *Dao) Remind(c context.Context, mid int64) (remind *bangumi.Remind, err error) {
	ip := metadata.String(c, metadata.RemoteIP)
	params := url.Values{}
	params.Set("mid", strconv.FormatInt(mid, 10))
	var res struct {
		Code   int             `json:"code"`
		Result *bangumi.Remind `json:"result"`
	}
	if err = d.client.Get(c, d.remind, ip, params, &res); err != nil {
		return
	}
	if res.Code != ecode.OK.Code() {
		err = errors.Wrap(ecode.Int(res.Code), d.remind+"?"+params.Encode())
		return
	}
	remind = res.Result
	return
}

func (d *Dao) EpPlayer(c context.Context, epIDs []int64, mobiApp, platform, device string, build, fnver, fnval int) (epPlayer map[int64]*bangumi.EpPlayer, err error) {
	ip := metadata.String(c, metadata.RemoteIP)
	params := url.Values{}
	params.Set("ep_ids", xstr.JoinInts(epIDs))
	params.Set("mobi_app", mobiApp)
	params.Set("platform", platform)
	params.Set("device", device)
	params.Set("build", strconv.Itoa(build))
	params.Set("ip", ip)
	params.Set("fnval", strconv.Itoa(fnval))
	params.Set("fnver", strconv.Itoa(fnver))
	var res struct {
		Code   int                         `json:"code"`
		Result map[int64]*bangumi.EpPlayer `json:"result"`
	}
	if err = d.client.Get(c, d.epPlayer, ip, params, &res); err != nil {
		return
	}
	if res.Code != ecode.OK.Code() {
		err = errors.Wrap(ecode.Int(res.Code), d.epPlayer+"?"+params.Encode())
		return
	}
	epPlayer = res.Result
	return
}

func (d *Dao) OgvPlaylist(ctx context.Context, arg *pgcstory.StoryPlayListReq) (*pgcstory.StoryPlayListReply, error) {
	resp, err := d.pgcStoryClient.QueryPlayList(ctx, arg)
	if err != nil {
		return nil, err
	}
	return resp, nil
}
