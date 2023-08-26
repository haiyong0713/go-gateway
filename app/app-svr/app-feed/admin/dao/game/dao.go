package game

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
	"time"

	"go-common/library/ecode"
	"go-common/library/log"
	httpx "go-common/library/net/http/blademaster"
	"go-gateway/app/app-svr/app-feed/admin/conf"
	"go-gateway/app/app-svr/app-feed/admin/model/game"
	"go-gateway/app/app-svr/app-feed/admin/util"
)

const (
	_searchPc         = "/game/pc/search/info"
	_rcmdPc           = "/game/pc/info"
	_gameInfoURI      = "/game/info"
	_gameEntryInfoURI = "/api/game/base/get_by_id"
)

const (
	PLAT_APP = 1
	PLAT_WEB = 2
)

// Dao is vip dao.
type Dao struct {
	client       *httpx.Client
	entryClient  *EntryClient
	gameURL      string
	entryGameURL string

	userFeed *conf.UserFeed
}

// New vip dao.
func New(c *conf.Config) (d *Dao) {
	d = &Dao{
		client:       httpx.NewClient(c.HTTPClient.Game),
		entryClient:  NewEntryClient(c.HTTPClient.EntryGame),
		gameURL:      c.Host.Game,
		entryGameURL: c.Host.EntryGame,
		userFeed:     c.UserFeed,
	}
	return
}

// SearchGame search game
func (d *Dao) SearchGame(c context.Context, gameID int64) (title string, err error) {
	return d.GamePC(c, _searchPc, gameID)
}

// WebRcmdGame web recommend game
func (d *Dao) WebRcmdGame(c context.Context, gameID int64) (title string, err error) {
	return d.GamePC(c, _rcmdPc, gameID)
}

// GamePC get pc game info
func (d *Dao) GamePC(c context.Context, ur string, gameID int64) (title string, err error) {
	params := url.Values{}
	params.Set("game_base_id", strconv.FormatInt(gameID, 10))
	params.Set("ts", strconv.FormatInt(time.Now().Unix()*1000, 10))
	res := &game.GameInfo{}
	if err = d.client.Get(c, d.gameURL+ur, "", params, &res); err != nil {
		log.Error("GamePC gameID(%d) error(%v) res(%+v)", gameID, err, res)
		return "", fmt.Errorf(util.ErrorNetFmts, util.ErrorNet, d.gameURL+"?"+ur+params.Encode(), err.Error())
	}
	if res.Code != ecode.OK.Code() {
		log.Error("GamePC gameID(%d) error(%v) res(%+v)", gameID, err, res)
		return "", fmt.Errorf(util.ErrorRes, util.ErrorDataNull, d.userFeed.Game, d.gameURL+ur+"?"+params.Encode())
	}
	if res.Data == nil {
		return "", fmt.Errorf(util.ErrorRes, util.ErrorDataNull, d.userFeed.Game, d.gameURL+ur+"?"+params.Encode())
	}
	return res.Data.GameName, nil
}

// GamesPCInfo get pc game info
func (d *Dao) GamesPCInfo(c context.Context, gameID int64) (g *game.Game, err error) {
	params := url.Values{}
	params.Set("game_base_id", strconv.FormatInt(gameID, 10))
	params.Set("ts", strconv.FormatInt(time.Now().Unix()*1000, 10))
	res := &game.GameInfo{}
	if err = d.client.Get(c, d.gameURL+_rcmdPc, "", params, &res); err != nil {
		log.Error("GamesPCInfo gameID(%d) error(%v) res(%+v)", gameID, err, res)
		return nil, fmt.Errorf(util.ErrorNetFmts, util.ErrorNet, d.gameURL+"?"+d.gameURL+_rcmdPc+params.Encode(), err.Error())
	}
	if res.Code != ecode.OK.Code() {
		log.Error("GamesPCInfo gameID(%d) error(%v) res(%+v)", gameID, err, res)
		return nil, fmt.Errorf(util.ErrorRes, util.ErrorDataNull, d.userFeed.Game, d.gameURL+_rcmdPc+"?"+params.Encode())
	}
	if res.Data == nil {
		return nil, fmt.Errorf(util.ErrorRes, util.ErrorDataNull, d.userFeed.Game, d.gameURL+_rcmdPc+"?"+params.Encode())
	}
	return &game.Game{ID: gameID, Title: res.Data.GameName, Image: res.Data.GameIcon}, err
}

// GameInfo game info
func (d *Dao) GameInfo(c context.Context, gameBaseID int64, platForm int) (info *game.Info, err error) {
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
	if err = d.client.Get(c, d.gameURL+_gameInfoURI, "", params, &res); err != nil {
		log.Error("GameInfo gameID(%d) error(%v) res(%+v)", gameBaseID, err, res)
		return nil, fmt.Errorf(util.ErrorNetFmts, util.ErrorNet, d.gameURL+"?"+_gameInfoURI+params.Encode(), err.Error())
	}
	if res.Code != ecode.OK.Code() {
		log.Error("GameInfo gameID(%d) error(%v) res(%+v)", gameBaseID, err, res)
		return nil, fmt.Errorf(util.ErrorRes, util.ErrorDataNull, d.userFeed.Game, d.gameURL+_gameInfoURI+"?"+params.Encode())
	}
	if res.Data == nil {
		return nil, fmt.Errorf(util.ErrorRes, util.ErrorDataNull, d.userFeed.Game, d.gameURL+_gameInfoURI+"?"+params.Encode())
	}
	return res.Data, nil
}

func (d *Dao) GameListApp(c context.Context, ids []int64) (ret map[int64]*game.Info, err error) {
	ret = make(map[int64]*game.Info)
	for _, id := range ids {
		info, err1 := d.GameInfo(c, id, PLAT_APP)
		if err1 != nil {
			log.Error("GameListApp gameID(%d) error(%v)", id, err1)
			continue
		}
		ret[id] = info
	}
	return
}

func (d *Dao) GameListWeb(c context.Context, ids []int64) (ret map[int64]string, err error) {
	ret = make(map[int64]string)
	for _, id := range ids {
		title, err1 := d.SearchGame(c, id)
		if err1 != nil {
			log.Error("GameListWeb gameID(%d) error(%v)", id, err1)
			continue
		}
		ret[id] = title
	}
	return
}

// GameEntryInfo game entry info 入库游戏信息查询
func (d *Dao) GameEntryInfo(c context.Context, gameBaseID int64) (info *game.EntryInfo, err error) {
	params := url.Values{}
	params.Set("game_base_id", strconv.FormatInt(gameBaseID, 10))
	res := &game.EntryInfo{}
	if err = d.entryClient.Get(c, d.entryGameURL+_gameEntryInfoURI, params, res); err != nil {
		log.Error("GameEntryInfo gameID(%d) error(%v) res(%+v)", gameBaseID, err, res)
		return nil, fmt.Errorf(util.ErrorNetFmts, util.ErrorNet, d.entryGameURL+_gameEntryInfoURI+"?"+params.Encode(), err.Error())
	}
	return res, nil
}
