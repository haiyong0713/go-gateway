package dao

import (
	"context"
	"net/url"
	"strconv"
	"time"

	"go-common/library/ecode"
	"go-common/library/log"
	"go-common/library/net/metadata"

	"go-gateway/app/web-svr/web/interface/model"
)

const (
	_gameInfoURI    = "/game/pc/info"
	_searchGameInfo = "/game/pc/search/info"
)

// GameInfo get game info.
func (d *Dao) GameInfo(c context.Context, id int64) (data *model.Game, err error) {
	params := url.Values{}
	params.Set("game_base_id", strconv.FormatInt(id, 10))
	params.Set("ts", strconv.FormatInt(time.Now().Unix()*1000, 10))
	var res struct {
		Code    int         `json:"code"`
		Data    *model.Game `json:"data"`
		Message string      `json:"message"`
	}
	if err = d.httpGame.Get(c, d.gameInfoURL, metadata.String(c, metadata.RemoteIP), params, &res); err != nil {
		log.Error("GameInfo d.httpGame.Get(%s) error(%v)", d.gameInfoURL+"?"+params.Encode(), err)
		return
	}
	if res.Code != ecode.OK.Code() {
		log.Error("GameInfo d.httpR.Get(%s) code(%v) msg(%s)", d.gameInfoURL+"?"+params.Encode(), res.Code, res.Message)
		err = ecode.Int(res.Code)
		return
	}
	data = res.Data
	data.GameBaseID = id
	return
}

// SearchGameInfo get search result game data by game id.
func (d *Dao) SearchGameInfo(c context.Context, id int64) (data *model.SearchGameCard, err error) {
	params := url.Values{}
	params.Set("game_base_id", strconv.FormatInt(id, 10))
	params.Set("ts", strconv.FormatInt(time.Now().Unix()*1000, 10))
	var res struct {
		Code    int                   `json:"code"`
		Data    *model.SearchGameCard `json:"data"`
		Message string                `json:"message"`
	}
	if err = d.httpGame.Get(c, d.searchGameInfoURL, metadata.String(c, metadata.RemoteIP), params, &res); err != nil {
		log.Error("GameInfo d.httpGame.Get(%s) error(%v)", d.searchGameInfoURL+"?"+params.Encode(), err)
		return
	}
	if res.Code != ecode.OK.Code() {
		log.Error("GameInfo d.httpR.Get(%s) code(%v) msg(%s)", d.searchGameInfoURL+"?"+params.Encode(), res.Code, res.Message)
		err = ecode.Int(res.Code)
		return
	}
	data = res.Data
	return
}
