package dao

import (
	"context"
	"net/url"
	"strconv"
	"time"

	"github.com/pkg/errors"
	"go-common/library/ecode"
	"go-common/library/net/metadata"
	"go-gateway/app/app-svr/story/internal/model"
)

const _game = "http://game-center-open-api.bilibili.co/game/detail/game_gifts"

func (d *dao) GameGifts(ctx context.Context, param *model.StoryGameParam) (*model.StoryGameReply, error) {
	params := url.Values{}
	params.Set("game_base_id", strconv.FormatInt(param.GameBaseId, 10))
	params.Set("sdk_type", castSDKType(param.MobiApp))
	params.Set("uid", strconv.FormatInt(param.Mid, 10))
	params.Set("ts", strconv.FormatInt(time.Now().Unix()*1000, 10))
	var res struct {
		Code int                   `json:"code"`
		Data *model.StoryGameReply `json:"data"`
	}
	if err := d.gameClient.Get(ctx, _game, metadata.String(ctx, metadata.RemoteIP), params, &res); err != nil {
		return nil, err
	}
	if res.Code != ecode.OK.Code() {
		return nil, errors.Wrap(ecode.Int(res.Code), _game+"?"+params.Encode())
	}
	return res.Data, nil
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
