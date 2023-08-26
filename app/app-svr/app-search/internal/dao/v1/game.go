package v1

import (
	"context"
	"net/url"
	"strconv"
	"time"

	"go-common/library/ecode"
	"go-common/library/net/metadata"
	"go-common/library/xstr"
	"go-gateway/app/app-svr/app-search/internal/model/search"

	gameEntryClient "git.bilibili.co/bapis/bapis-go/manager/operation/game-entry"

	"github.com/pkg/errors"
)

func (d *dao) CloudGameEntry(ctx context.Context, req *gameEntryClient.MultiShowReq) (*gameEntryClient.MultiShowResp, error) {
	return d.cloudGameEntryClient.MultiShow(ctx, req)
}

func (d *dao) MultiGameInfos(ctx context.Context, mid int64, ids []int64, build, sdkType int) (map[int64]*search.NewGame, error) {
	params := url.Values{}
	params.Set("game_base_ids", xstr.JoinInts(ids))
	params.Set("sdk_type", strconv.Itoa(sdkType))
	params.Set("mid", strconv.FormatInt(mid, 10))
	params.Set("source_tag", "3") // 搜索游戏卡
	params.Set("build", strconv.Itoa(build))
	params.Set("ts", strconv.FormatInt(time.Now().Unix()*1000, 10))
	var res struct {
		Code int               `json:"code"`
		Data []*search.NewGame `json:"data"`
	}
	if err := d.gameClient.Get(ctx, d.gameMultiInfos, metadata.String(ctx, metadata.RemoteIP), params, &res); err != nil {
		return nil, err
	}
	if res.Code != ecode.OK.Code() {
		return nil, errors.Wrap(ecode.Int(res.Code), d.gameMultiInfos+"?"+params.Encode())
	}
	out := make(map[int64]*search.NewGame, len(res.Data))
	for _, v := range res.Data {
		if !v.IsOnline {
			continue
		}
		out[v.GameBaseID] = v
	}
	return out, nil
}

func (d *dao) TopGame(ctx context.Context, mid int64, topGameIDs []int64, sdkType int) ([]*search.TopGameData, error) {
	params := url.Values{}
	params.Set("game_base_ids", xstr.JoinInts(topGameIDs))
	params.Set("sdk_type", strconv.Itoa(sdkType))
	params.Set("mid", strconv.FormatInt(mid, 10))
	params.Set("ts", strconv.FormatInt(time.Now().UnixNano()/1000000, 10))
	var res struct {
		Code    int                   `json:"code"`
		Data    []*search.TopGameData `json:"data"`
		Message string                `json:"message"`
	}
	if err := d.gameClient.Get(ctx, d.topGame, metadata.String(ctx, metadata.RemoteIP), params, &res); err != nil {
		return nil, errors.Wrapf(err, "TopGame d.httpR.Get(%s)", d.topGame+"?"+params.Encode())
	}
	if res.Code != ecode.OK.Code() {
		return nil, errors.Wrapf(ecode.Int(res.Code), "TopGame d.httpR.Get(%s)", d.topGame+"?"+params.Encode())
	}
	return res.Data, nil
}

func (d *dao) FetchTopGameConfigs(ctx context.Context, gameIds []int64) (*search.TopGameConfig, error) {
	params := url.Values{}
	params.Set("game_ids", xstr.JoinInts(gameIds))
	var res struct {
		Code int                   `json:"code"`
		Data *search.TopGameConfig `json:"data"`
	}
	if err := d.gameClient.Get(ctx, d.topGameConfig, metadata.String(ctx, metadata.RemoteIP), params, &res); err != nil {
		return nil, err
	}
	if res.Code != ecode.OK.Code() {
		return nil, errors.Wrapf(ecode.Int(res.Code), "FetchTopGameConfigs error uri(%s), gameid(%+v)", d.topGameConfig+"?"+params.Encode(), gameIds)
	}
	return res.Data, nil
}

func (d *dao) FetchTopGameInlineConfigs(ctx context.Context, cardIds []int64) (*search.TopGameInlineInfo, error) {
	params := url.Values{}
	params.Set("card_ids", xstr.JoinInts(cardIds))
	var res struct {
		Code int                       `json:"code"`
		Data *search.TopGameInlineInfo `json:"data"`
	}
	if err := d.gameClient.Get(ctx, d.topGameInlineInfo, metadata.String(ctx, metadata.RemoteIP), params, &res); err != nil {
		return nil, err
	}
	if res.Code != ecode.OK.Code() {
		return nil, errors.Wrapf(ecode.Int(res.Code), "FetchTopGameInlineConfigs error uri(%s), cardids(%+v)", d.topGameInlineInfo+"?"+params.Encode(), cardIds)
	}
	return res.Data, nil
}
