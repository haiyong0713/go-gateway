package dao

import (
	"context"
	"net/url"
	"strconv"
	"time"

	"go-common/library/ecode"
	"go-common/library/net/metadata"
	"go-common/library/xstr"

	"go-gateway/app/web-svr/native-page/admin/model"

	"github.com/pkg/errors"
)

const (
	_gameURI     = "/dynamic_card/multi_game_info"
	_gameListURI = "/dynamic_card/game_list"
)

func (d *Dao) GameList(c context.Context, gameIDs []int64) (map[int64]*model.GameList, error) {
	ip := metadata.String(c, metadata.RemoteIP)
	params := url.Values{}
	params.Set("game_base_ids", xstr.JoinInts(gameIDs))
	ts := strconv.FormatInt(time.Now().UnixNano()/1e6, 10)
	params.Set("ts", ts)
	var res struct {
		Code int               `json:"code"`
		Data []*model.GameList `json:"data"`
	}
	if err := d.gameClient.Get(c, d.gameListURL, ip, params, &res); err != nil {
		return nil, err
	}
	if res.Code != ecode.OK.Code() {
		err := errors.Wrap(ecode.Int(res.Code), d.gameListURL+"?"+params.Encode())
		return nil, err
	}
	rly := make(map[int64]*model.GameList)
	for _, v := range res.Data {
		if v == nil || v.GameBaseId == 0 {
			continue
		}
		rly[v.GameBaseId] = v
	}
	return rly, nil
}

func (d *Dao) MultiGameInfo(c context.Context, gameIDs []int64, mid int64, platformType string) (map[int64]*model.GameItem, error) {
	ip := metadata.String(c, metadata.RemoteIP)
	params := url.Values{}
	params.Set("game_base_ids", xstr.JoinInts(gameIDs))
	params.Set("uid", strconv.FormatInt(mid, 10))
	ts := strconv.FormatInt(time.Now().UnixNano()/1e6, 10)
	params.Set("ts", ts)
	//平台类型：0=PC，1=安卓，2=IOS
	params.Set("platform_type", platformType)
	params.Set("source", "1009")
	var res struct {
		Code int               `json:"code"`
		Data []*model.GameItem `json:"data"`
	}
	if err := d.gameClient.Get(c, d.gameInfoURL, ip, params, &res); err != nil {
		return nil, err
	}
	if res.Code != ecode.OK.Code() {
		err := errors.Wrap(ecode.Int(res.Code), d.gameInfoURL+"?"+params.Encode())
		return nil, err
	}
	rly := make(map[int64]*model.GameItem)
	for _, v := range res.Data {
		if v == nil || v.GameBaseId == 0 {
			continue
		}
		rly[v.GameBaseId] = v
	}
	return rly, nil
}
