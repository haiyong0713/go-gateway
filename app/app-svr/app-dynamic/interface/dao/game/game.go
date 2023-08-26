package game

import (
	"context"
	"net/url"
	"strconv"
	"time"

	"go-common/library/ecode"
	"go-common/library/log"
	"go-common/library/net/metadata"
	"go-common/library/xstr"

	"go-gateway/app/app-svr/app-dynamic/interface/model/game"
	xmetric "go-gateway/app/app-svr/app-dynamic/interface/model/metric"

	"github.com/pkg/errors"
)

const (
	_gameURL       = "/dynamic_card/multi_game_info"
	_gameActionURL = "/dynamic_card/action/multi_game_info"
)

// 获取游戏附加卡 Games
func (d *Dao) Games(ctx context.Context, uid int64, platform string, gameBaseIds []int64) (map[int64]*game.Game, error) {
	var ret struct {
		Code int          `json:"code"`
		Msg  string       `json:"msg"`
		Data []*game.Game `json:"data"`
	}
	var err error
	gameBaseIdsStr := xstr.JoinInts(gameBaseIds)
	params := url.Values{}
	platformType := strconv.Itoa(d.GetPlatformType(platform))
	ts := strconv.FormatInt(time.Now().UnixNano()/1e6, 10)
	params.Set("game_base_ids", gameBaseIdsStr)
	params.Set("platform_type", platformType)
	params.Set("ts", ts)
	params.Set("uid", strconv.FormatInt(uid, 10))
	queryUrl := d.c.Hosts.Game + _gameURL
	if err = d.client.Get(ctx, queryUrl, "", params, &ret); err != nil {
		xmetric.DyanmicItemAPI.Inc(queryUrl, "request_error")
		log.Error("%s query failed, error(%v),uid(%d),platform(%s),gameBaseIds(%s),ts(%s)", queryUrl, err, uid, platformType, gameBaseIdsStr, ts)
		return nil, err
	}
	if ret.Code != ecode.OK.Code() {
		xmetric.DyanmicItemAPI.Inc(queryUrl, "reply_code_error")
		log.Error("%s query failed, error(%v),uid(%d),platform(%s),gameBaseIds(%s),ts(%s)", queryUrl, err, uid, platformType, gameBaseIdsStr, ts)
		err = errors.Wrap(ecode.Int(ret.Code), queryUrl+"?"+params.Encode())
		return nil, err
	}
	if len(ret.Data) == 0 {
		xmetric.DyanmicItemAPI.Inc(queryUrl, "reply_date_error")
		log.Error("%s query failed, error(%v),uid(%d),platform(%s),gameBaseIds(%s),ts(%s)", queryUrl, err, uid, platformType, gameBaseIdsStr, ts)
		return nil, nil
	}
	//按入参gameBaseIds的顺序append
	gameInfoMap := make(map[int64]*game.Game, len(ret.Data))
	for _, gameInfo := range ret.Data {
		if gameInfo != nil {
			gameInfoMap[gameInfo.GameBaseId] = gameInfo
		}
	}
	return gameInfoMap, nil
}

// nolint:gomnd
func (d *Dao) GetPlatformType(platform string) int {
	if platform == "android" {
		return 1
	} else if platform == "ios" {
		return 2
	} else {
		return 0
	}
}

func (d *Dao) GameAction(ctx context.Context, uid int64, platform string, gameBaseIds []int64) (map[int64]*game.Game, error) {
	var ret struct {
		Code int          `json:"code"`
		Msg  string       `json:"msg"`
		Data []*game.Game `json:"data"`
	}
	params := url.Values{}
	params.Set("game_base_ids", xstr.JoinInts(gameBaseIds))
	params.Set("platform_type", strconv.Itoa(d.GetPlatformType(platform)))
	params.Set("uid", strconv.FormatInt(uid, 10))
	params.Set("ts", strconv.FormatInt(time.Now().UnixNano()/1e6, 10))
	url := d.c.Hosts.Game + _gameActionURL
	if err := d.client.Get(ctx, url, metadata.String(ctx, metadata.RemoteIP), params, &ret); err != nil {
		xmetric.DyanmicItemAPI.Inc(url, "request_error")
		log.Error("GameAction d.client.Get url(%s) error(%v)", url+"?"+params.Encode(), err)
		return nil, err
	}
	if ret.Code != ecode.OK.Code() {
		xmetric.DyanmicItemAPI.Inc(url, "reply_code_error")
		err := errors.Wrap(ecode.Int(ret.Code), url+"?"+params.Encode())
		return nil, err
	}
	res := map[int64]*game.Game{}
	for _, v := range ret.Data {
		tmp := &game.Game{}
		*tmp = *v
		tmp.FromGameStatus()
		res[v.GameBaseId] = tmp
	}
	return res, nil
}
