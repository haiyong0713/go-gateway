package game

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"go-common/library/ecode"
	httpx "go-common/library/net/http/blademaster"
	"go-common/library/net/metadata"
	"go-common/library/xstr"

	"go-gateway/app/app-svr/app-view/interface/conf"
	"go-gateway/app/app-svr/app-view/interface/model"
	"go-gateway/app/app-svr/app-view/interface/model/game"

	"github.com/pkg/errors"
)

const (
	_infoURL   = "/game/info"
	_game_info = "/game/multi_get_game_info"
)

type Dao struct {
	client      *httpx.Client
	infoURL     string
	gameInfoUrl string
	key         string
	secret      string
}

func New(c *conf.Config) (d *Dao) {
	d = &Dao{
		client:      httpx.NewClient(c.HTTPGame),
		infoURL:     c.Host.Game + _infoURL,
		gameInfoUrl: c.Host.Game + _game_info,
		key:         c.HTTPGame.Key,
		secret:      c.HTTPGame.Secret,
	}
	return
}

func (d *Dao) Info(c context.Context, gameID int64, plat int8) (info *game.Info, err error) {
	var platType int
	if model.IsAndroid(plat) {
		platType = 1
	} else if model.IsIOS(plat) {
		platType = 2
	}
	if platType == 0 {
		return
	}
	var req *http.Request
	params := url.Values{}
	params.Set("appkey", d.key)
	params.Set("game_base_id", strconv.FormatInt(gameID, 10))
	params.Set("platform_type", strconv.Itoa(platType))
	params.Set("ts", strconv.FormatInt(time.Now().UnixNano()/1e6, 10))
	mh := md5.Sum([]byte(params.Encode() + d.secret))
	params.Set("sign", hex.EncodeToString(mh[:]))
	if req, err = d.client.NewRequest("GET", d.infoURL, "", params); err != nil {
		return
	}
	var res struct {
		Code int        `json:"code"`
		Data *game.Info `json:"data"`
	}
	if err = d.client.Do(c, req, &res); err != nil {
		return
	}
	if res.Code != ecode.OK.Code() {
		err = errors.Wrap(ecode.Int(res.Code), d.infoURL+"?"+params.Encode())
		return
	}
	info = res.Data
	return
}

func (d *Dao) MultiGameInfos(ctx context.Context, id []int64, mobiApp string, build int) (map[int64]*game.Game, error) {
	params := url.Values{}
	params.Set("game_base_ids", xstr.JoinInts(id))
	params.Set("sdk_type", CastSDKType(mobiApp))
	params.Set("source_tag", "4")
	params.Set("build", strconv.Itoa(build))

	params.Set("appkey", d.key)
	params.Set("ts", strconv.FormatInt(time.Now().UnixNano()/1e6, 10))
	mh := md5.Sum([]byte(params.Encode() + d.secret))
	params.Set("sign", hex.EncodeToString(mh[:]))

	var res struct {
		Code int          `json:"code"`
		Data []*game.Game `json:"data"`
	}
	if err := d.client.Get(ctx, d.gameInfoUrl, metadata.String(ctx, metadata.RemoteIP), params, &res); err != nil {
		return nil, err
	}
	if res.Code != ecode.OK.Code() {
		return nil, errors.Wrap(ecode.Int(res.Code), d.gameInfoUrl+"?"+params.Encode())
	}

	out := make(map[int64]*game.Game, len(res.Data))
	for _, v := range res.Data {
		out[v.GameBaseID] = v
	}
	return out, nil
}

func CastSDKType(mobiApp string) string {
	switch mobiApp {
	case "android", "android_b", "android_i", "android_hd":
		return "1"
	case "iphone", "iphone_i", "iphone_b", "ipad":
		return "2"
	default:
		return "1"
	}
}
