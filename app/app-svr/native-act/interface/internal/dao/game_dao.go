package dao

import (
	"context"
	"net/url"
	"strconv"
	"time"

	"go-common/library/ecode"
	"go-common/library/log"
	bm "go-common/library/net/http/blademaster"
	"go-common/library/net/metadata"
	"go-common/library/xstr"

	"go-gateway/app/app-svr/native-act/interface/internal/model"

	"github.com/pkg/errors"
)

const (
	_gameURI = "/dynamic_card/multi_game_info"
)

type gameDao struct {
	host       string
	httpClient *bm.Client
}

func (d *gameDao) MultiGameInfo(c context.Context, gameIDs []int64, mid int64, platformType string) ([]*model.GaItem, error) {
	ip := metadata.String(c, metadata.RemoteIP)
	params := url.Values{}
	params.Set("game_base_ids", xstr.JoinInts(gameIDs))
	params.Set("uid", strconv.FormatInt(mid, 10))
	ts := strconv.FormatInt(time.Now().UnixNano()/1e6, 10)
	params.Set("ts", ts)
	params.Set("platform_type", platformType)
	params.Set("source", "1009")
	var res struct {
		Code int             `json:"code"`
		Data []*model.GaItem `json:"data"`
	}
	if err := d.httpClient.Get(c, d.host+_gameURI, ip, params, &res); err != nil {
		log.Errorc(c, "gameDao.MultiGameInfo %s error(%v)", d.host+_gameURI+"?"+params.Encode(), err)
		return nil, err
	}
	if res.Code != ecode.OK.Code() {
		err := errors.Wrap(ecode.Int(res.Code), d.host+_gameURI+"?"+params.Encode())
		log.Errorc(c, "gameDao.MultiGameInfo error(%v)", err)
		return nil, err
	}
	return res.Data, nil
}
