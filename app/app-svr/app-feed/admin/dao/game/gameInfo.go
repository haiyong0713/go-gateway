package game

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
	"time"

	"go-common/library/ecode"
	"go-common/library/log"
	"go-gateway/app/app-svr/app-feed/admin/model/game"
	"go-gateway/app/app-svr/app-feed/admin/util"
)

const (
	_gameInfoApp = "/game/info"
)

func (d *Dao) GameInfoApp(c context.Context, gameBaseID int64, platForm int) (info *game.Info, err error) {
	params := url.Values{}
	params.Set("game_base_id", strconv.FormatInt(gameBaseID, 10))
	params.Set("platform_type", strconv.Itoa(platForm))
	params.Set("ts", strconv.FormatInt(time.Now().UnixNano()/1000000, 10))
	type gameRes struct {
		Code int        `json:"code"`
		Msg  string     `json:"message"`
		Data *game.Info `json:"data"`
	}
	res := &gameRes{}
	if err = d.client.Get(c, d.gameURL+_gameInfoApp, "", params, &res); err != nil {
		log.Error("GameInfoApp gameID(%d) error(%v) res(%+v)", gameBaseID, err, res)
		return nil, fmt.Errorf(util.ErrorNetFmts, util.ErrorNet, d.gameURL+_gameInfoApp+"?"+params.Encode(), err.Error())
	}
	if res.Code != ecode.OK.Code() {
		log.Error("GameInfoApp gameID(%d) error(%v) res(%+v)", gameBaseID, err, res)
		return nil, fmt.Errorf(util.ErrorRes, util.ErrorDataNull, d.userFeed.Game, d.gameURL+_gameInfoApp+"?"+params.Encode())
	}
	if res.Data == nil {
		return nil, fmt.Errorf(util.ErrorRes, util.ErrorDataNull, d.userFeed.Game, d.gameURL+_gameInfoApp+"?"+params.Encode())
	}
	info = res.Data
	return
}
