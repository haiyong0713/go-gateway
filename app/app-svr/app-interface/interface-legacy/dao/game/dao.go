package game

import (
	"context"
	"net/url"
	"strconv"
	"time"

	"go-common/library/ecode"
	"go-common/library/log"
	httpx "go-common/library/net/http/blademaster"
	"go-common/library/net/metadata"
	"go-common/library/xstr"
	"go-gateway/app/app-svr/app-interface/interface-legacy/conf"
	gmdl "go-gateway/app/app-svr/app-interface/interface-legacy/model/game"
	"go-gateway/app/app-svr/app-interface/interface-legacy/model/space"

	gameEntryClient "git.bilibili.co/bapis/bapis-go/manager/operation/game-entry"

	"github.com/pkg/errors"
)

const (
	_playGame       = "/game/recent/play"
	_playGameSub    = "/game/recent/play/page"
	_topGame        = "/game/multi_get_game_info/for_intensify_card"
	_mineGameCenter = "/x/operation/game-banner"
	_gameDmp        = "/game_dmp/uid/in"
	_gameInfo       = "/game/multi_get_game_info"
	_topGameButton  = "/x/admin/manager/search/game/buttonInfo"
	_gameSwitch     = "/game/func_switch"
)

// Dao is game dao.
type Dao struct {
	client               *httpx.Client
	playGame             string
	playGameSub          string
	topGame              string
	mineGameCenter       string
	gameMultiInfos       string // 批量获取游戏数据
	cloudGameEntryClient gameEntryClient.OperationItemGameEntryV1Client
	gameDmp              string
	topGameConfig        string
	gameSwitch           string //是否打开游戏的tab
}

// New game dao
func New(c *conf.Config) (d *Dao) {
	d = &Dao{
		client:         httpx.NewClient(c.HTTPGameCo),
		playGame:       c.Host.GameCo + _playGame,
		playGameSub:    c.Host.GameCo + _playGameSub,
		topGame:        c.Host.GameCo + _topGame,
		gameMultiInfos: c.Host.GameCo + _gameInfo,
		mineGameCenter: c.Host.Manager + _mineGameCenter, //manager后台帮游戏业务出了个游戏提示条接口
		gameDmp:        c.Host.GameDmp + _gameDmp,
		topGameConfig:  c.Host.Manager + _topGameButton,
		gameSwitch:     c.Host.GameCo + _gameSwitch,
	}
	var err error
	if d.cloudGameEntryClient, err = gameEntryClient.NewClientOperationItemGameEntryV1(c.GameEntryGRPC); err != nil {
		panic(err)
	}
	return
}

func (d *Dao) CloudGameEntry(ctx context.Context, req *gameEntryClient.MultiShowReq) (*gameEntryClient.MultiShowResp, error) {
	return d.cloudGameEntryClient.MultiShow(ctx, req)
}

func (d *Dao) MultiGameInfos(ctx context.Context, mid int64, ids []int64, build, sdkType int) (map[int64]*gmdl.Game, error) {
	params := url.Values{}
	params.Set("game_base_ids", xstr.JoinInts(ids))
	params.Set("sdk_type", strconv.Itoa(sdkType))
	params.Set("mid", strconv.FormatInt(mid, 10))
	params.Set("source_tag", "3") // 搜索游戏卡
	params.Set("build", strconv.Itoa(build))
	params.Set("ts", strconv.FormatInt(time.Now().Unix()*1000, 10))
	var res struct {
		Code int          `json:"code"`
		Data []*gmdl.Game `json:"data"`
	}
	if err := d.client.Get(ctx, d.gameMultiInfos, metadata.String(ctx, metadata.RemoteIP), params, &res); err != nil {
		return nil, err
	}
	if res.Code != ecode.OK.Code() {
		return nil, errors.Wrap(ecode.Int(res.Code), d.gameMultiInfos+"?"+params.Encode())
	}
	out := make(map[int64]*gmdl.Game, len(res.Data))
	for _, v := range res.Data {
		if !v.IsOnline {
			continue
		}
		out[v.GameBaseID] = v
	}
	return out, nil
}

// RecentPlay .
func (d *Dao) RecentGame(c context.Context, mid int64, pn, ps int, platform string) (rly *gmdl.RecentGame, err error) {
	var platformType int
	switch platform {
	case "android":
		platformType = 1
	case "ios":
		platformType = 2
	default:
		return
	}
	tm := strconv.FormatInt(time.Now().UnixNano()/1000000, 10) // 用毫秒
	params := url.Values{}
	params.Set("uid", strconv.FormatInt(mid, 10))
	params.Set("platform_type", strconv.Itoa(platformType))
	params.Set("page_num", strconv.Itoa(pn))
	params.Set("page_size", strconv.Itoa(ps))
	params.Set("ts", tm)
	var res struct {
		Code int              `json:"code"`
		Data *gmdl.RecentGame `json:"data"`
	}
	if err = d.client.Get(c, d.playGame, metadata.String(c, metadata.RemoteIP), params, &res); err != nil {
		log.Error("RecentPlay d.httpR.Get(%s,%d) error(%v)", d.playGame+params.Encode(), mid, err)
		return
	}
	if res.Code != ecode.OK.Code() {
		log.Error("RecentPlay d.httpR.Get(%s) code error(%d)", d.playGame+"?"+params.Encode(), res.Code)
		err = ecode.Int(res.Code)
		return
	}
	rly = res.Data
	return
}

// RecentPlaySub .
func (d *Dao) RecentGameSub(c context.Context, mid int64, pn, ps int, platform string) (rly *gmdl.RecentGameSub, err error) {
	var platformType int
	switch platform {
	case "android":
		platformType = 1
	case "ios":
		platformType = 2
	default:
		return
	}
	tm := strconv.FormatInt(time.Now().UnixNano()/int64(time.Millisecond), 10) // 用毫秒
	params := url.Values{}
	params.Set("uid", strconv.FormatInt(mid, 10))
	params.Set("platform_type", strconv.Itoa(platformType))
	params.Set("page_num", strconv.Itoa(pn))
	params.Set("page_size", strconv.Itoa(ps))
	params.Set("ts", tm)
	var res struct {
		Code int                 `json:"code"`
		Data *gmdl.RecentGameSub `json:"data"`
	}
	if err = d.client.Get(c, d.playGameSub, metadata.String(c, metadata.RemoteIP), params, &res); err != nil {
		log.Error("RecentPlay d.httpR.Get(%s,%d) error(%v)", d.playGame+params.Encode(), mid, err)
		return
	}
	if res.Code != ecode.OK.Code() {
		log.Error("RecentPlay d.httpR.Get(%s) code error(%d)", d.playGame+"?"+params.Encode(), res.Code)
		err = ecode.Int(res.Code)
		return
	}
	rly = res.Data
	return
}

func (d *Dao) FetchGameTip(c context.Context, mid, build, plat int64, buvid string) ([]*space.GameTip, error) {
	params := url.Values{}
	params.Set("mid", strconv.FormatInt(mid, 10))
	params.Set("buvid", buvid)
	params.Set("build", strconv.FormatInt(build, 10))
	params.Set("plat", strconv.FormatInt(plat, 10))
	var res struct {
		Code int `json:"code"`
		Data struct {
			Banner []*space.GameTip `json:"banner"`
		} `json:"data"`
	}
	if err := d.client.Get(c, d.mineGameCenter, metadata.String(c, metadata.RemoteIP), params, &res); err != nil {
		return nil, errors.WithMessagef(err, "mineGameCenter tip error uri(%s), mid(%d)", d.mineGameCenter, mid)
	}
	if res.Code != ecode.OK.Code() {
		return nil, errors.WithMessagef(ecode.Int(res.Code), "mineGameCenter tip error uri(%s), mid(%d)", d.mineGameCenter, mid)
	}
	return res.Data.Banner, nil
}

func (d *Dao) FetchGameDmps(c context.Context, mid int64, dmpIds []int64) ([]*space.GameDmpReply, error) {
	params := url.Values{}
	params.Set("uid", strconv.FormatInt(mid, 10))
	for _, v := range dmpIds {
		params.Add("ids", strconv.FormatInt(v, 10))
	}
	var res struct {
		Code int                   `json:"code"`
		Data []*space.GameDmpReply `json:"data"`
	}
	if err := d.client.Get(c, d.gameDmp, metadata.String(c, metadata.RemoteIP), params, &res); err != nil {
		return nil, err
	}
	if res.Code != ecode.OK.Code() {
		return nil, errors.Wrapf(ecode.Int(res.Code), "CheckGameDmp error uri(%s), mid(%d)", d.gameDmp, mid)
	}
	return res.Data, nil
}

func (d *Dao) FetchTopGameConfigs(ctx context.Context, gameIds []int64) (*gmdl.TopGameConfig, error) {
	params := url.Values{}
	params.Set("game_ids", xstr.JoinInts(gameIds))
	var res struct {
		Code int                 `json:"code"`
		Data *gmdl.TopGameConfig `json:"data"`
	}
	if err := d.client.Get(ctx, d.topGameConfig, metadata.String(ctx, metadata.RemoteIP), params, &res); err != nil {
		return nil, err
	}
	if res.Code != ecode.OK.Code() {
		return nil, errors.Wrapf(ecode.Int(res.Code), "FetchTopGameConfigs error uri(%s), gameid(%+v)", d.topGameConfig+"?"+params.Encode(), gameIds)
	}
	return res.Data, nil
}

func (d *Dao) GameTabSwitch(c context.Context) (int, error) {
	ip := metadata.String(c, metadata.RemoteIP)
	params := url.Values{}
	tm := strconv.FormatInt(time.Now().UnixNano()/int64(time.Millisecond), 10) // 用毫秒
	params.Set("ts", tm)
	params.Set("key_word", "game_viewed_history_func")
	var res struct {
		Code    int    `json:"code"`
		Data    int    `json:"data"`
		Message string `json:"message"`
	}
	if err := d.client.Get(c, d.gameSwitch, ip, params, &res); err != nil {
		return 0, err
	}
	if res.Code != ecode.OK.Code() {
		return 0, errors.Wrap(ecode.Int(res.Code), res.Message+"?"+params.Encode())
	}
	return res.Data, nil
}
