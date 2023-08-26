package game

import (
	"bytes"
	"context"
	"fmt"
	"net/url"
	"strconv"
	"time"

	"go-common/library/ecode"
	httpx "go-common/library/net/http/blademaster"
	"go-common/library/net/metadata"
	"go-common/library/utils/collection"
	"go-gateway/app/app-svr/app-card/interface/model/card/game"
	"go-gateway/app/app-svr/app-feed/interface/conf"

	"github.com/pkg/errors"
)

const (
	_game = "/game/multi_get_game_info"
)

// Dao .
type Dao struct {
	client *httpx.Client
	host   string
}

func New(c *conf.Config) (d *Dao) {
	d = &Dao{
		client: httpx.NewClient(c.HTTPGame),
		host:   c.Host.GameCenter,
	}
	return
}

func (d *Dao) MultiGameInfos(ctx context.Context, id []int64, mobiApp string, build int, gameParams []*game.GameParam) (map[int64]*game.Game, error) {
	params := url.Values{}
	params.Set("materials_ids", deriveMaterialsIds(gameParams))
	params.Set("game_base_ids", collection.JoinSliceInt(id, ","))
	params.Set("sdk_type", castSDKType(mobiApp))
	params.Set("source_tag", "1")
	params.Set("ts", strconv.FormatInt(time.Now().Unix()*1000, 10))
	params.Set("build", strconv.Itoa(build))
	var res struct {
		Code int          `json:"code"`
		Data []*game.Game `json:"data"`
	}
	if err := d.client.Get(ctx, d.host+_game, metadata.String(ctx, metadata.RemoteIP), params, &res); err != nil {
		return nil, err
	}
	if res.Code != ecode.OK.Code() {
		return nil, errors.Wrap(ecode.Int(res.Code), d.host+_game+"?"+params.Encode())
	}
	out := make(map[int64]*game.Game, len(res.Data))
	for _, v := range res.Data {
		out[v.GameBaseID] = v
	}
	return out, nil
}

func deriveMaterialsIds(params []*game.GameParam) string {
	buffer := bytes.NewBufferString("")
	for i := 0; i < len(params); i++ {
		buffer.WriteString(fmt.Sprintf("%d_%d", params[i].GameId, params[i].CreativeId))
		if i < len(params)-1 {
			buffer.WriteString(",")
		}
	}
	return buffer.String()
}

func castSDKType(mobiApp string) string {
	switch mobiApp {
	case "android", "android_b", "android_i", "android_hd":
		return "1"
	case "iphone", "iphone_i", "iphone_b", "ipad":
		return "2"
	default:
		return "1"
	}
}
