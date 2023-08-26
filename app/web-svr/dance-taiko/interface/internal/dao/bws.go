package dao

import (
	"bytes"
	"context"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"go-common/library/ecode"
	"go-common/library/log"
	"go-gateway/app/web-svr/dance-taiko/interface/internal/model"

	"github.com/pkg/errors"
)

const (
	_createRoom = "/createRoom"
	_midInfo    = "/getMidInfo"
	_end        = "/endGame"
	_start      = "/startGame"
	_joinRoom   = "/joinRoom"
	_resetRoom  = "/reset"
)

func (d *dao) BwsCreateRoom(c context.Context) (int, error) {
	var (
		params = url.Values{}
		url    = d.conf.BwsCfg.Host + _createRoom
	)
	params.Set("gameId", fmt.Sprintf("%d", d.conf.BwsCfg.GameId))
	params.Set("capacity", fmt.Sprintf("%d", 20))
	req, err := d.client.NewRequest(http.MethodPost, url, "", params)
	if err != nil {
		return 0, errors.Wrapf(err, "BwsCreateRoom time(%v)", time.Now())
	}
	var reply = struct {
		Code int    `json:"code"`
		Msg  string `json:"message"`
		Data int    `json:"data"`
	}{}
	if err := d.client.Do(c, req, &reply); err != nil {
		return 0, errors.Wrapf(err, "BwsCreateRoom req(%v)", req)
	}
	if reply.Code != 0 {
		return 0, errors.Wrapf(ecode.Int(reply.Code), "BwsCreateRoom req(%v)", req)
	}
	return reply.Data, nil
}

func (d *dao) BwsMidInfo(c context.Context, mid int64, gameId int) (*model.BwsPlayInfo, error) {
	var (
		params = url.Values{}
		url    = d.conf.BwsCfg.Host + _midInfo
	)
	params.Set("mid", fmt.Sprintf("%d", mid))
	params.Set("roomId", fmt.Sprintf("%d", gameId))
	req, err := d.client.NewRequest(http.MethodGet, url, "", params)
	if err != nil {
		return nil, errors.Wrapf(err, "BwsMidInfo mid(%d) game(%d)", mid, gameId)
	}
	var reply = struct {
		Code int                `json:"code"`
		Msg  string             `json:"message"`
		Data *model.BwsPlayInfo `json:"data"`
	}{}
	if err := d.client.Do(c, req, &reply); err != nil {
		return nil, errors.Wrapf(err, "BwsMidInfo req(%v)", req)
	}
	if reply.Code != 0 {
		return nil, errors.Wrapf(ecode.Int(reply.Code), "BwsMidInfo req(%v)", req)
	}
	reply.Data.Mid = mid
	return reply.Data, nil
}

func (d *dao) BwsStartGame(c context.Context, gameId int) error {
	var (
		params = url.Values{}
		url    = d.conf.BwsCfg.Host + _start
	)
	params.Set("roomId", fmt.Sprintf("%d", gameId))
	req, err := d.client.NewRequest(http.MethodPost, url, "", params)
	if err != nil {
		return errors.Wrapf(err, "BwsStartGame gameId(%d)", gameId)
	}
	var reply = struct {
		Code int    `json:"code"`
		Msg  string `json:"message"`
	}{}
	if err := d.client.Do(c, req, &reply); err != nil {
		return errors.Wrapf(err, "BwsStartGame req(%v)", req)
	}
	if reply.Code != 0 {
		return errors.Wrapf(ecode.Int(reply.Code), "BwsStartGame req(%v)", req)
	}
	return nil
}

func (d *dao) BwsJoinRoom(c context.Context, gameId int, mid int64) error {
	var (
		params = url.Values{}
		url    = d.conf.BwsCfg.Host + _joinRoom
	)
	params.Set("roomId", fmt.Sprintf("%d", gameId))
	params.Set("mid", fmt.Sprintf("%d", mid))
	req, err := d.client.NewRequest(http.MethodPost, url, "", params)
	if err != nil {
		return errors.Wrapf(err, "BwsJoinRoom gameId(%d) mid(%d)", gameId, mid)
	}
	var reply = struct {
		Code int    `json:"code"`
		Msg  string `json:"message"`
	}{}
	if err := d.client.Do(c, req, &reply); err != nil {
		return errors.Wrapf(err, "BwsJoinRoom req(%v)", req)
	}
	if reply.Code != 0 {
		return errors.Wrapf(ecode.Int(reply.Code), "BwsJoinRoom req(%v)", req)
	}
	return nil
}

func (d *dao) BwsEndGame(c context.Context, gameId int, players []*model.BwsPlayResult) error {
	query, err := d.sign(url.Values{})
	if err != nil {
		return err
	}
	var url = fmt.Sprintf("%s%s?%s", d.conf.BwsCfg.Host, _end, query)
	var reqInfo = struct {
		RoomId int                    `json:"roomId"`
		Result []*model.BwsPlayResult `json:"result"`
	}{gameId, players}
	result, err := json.Marshal(reqInfo)
	if err != nil {
		return errors.Wrapf(err, "BwsEndGame gameId(%d) players(%v)", gameId, players)
	}
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(result))
	if err != nil {
		return errors.Wrapf(err, "BwsEndGame data(%v)", result)
	}
	defer req.Body.Close()
	req.Header.Set("Content-Type", "application/json")
	var reply = struct {
		Code int    `json:"code"`
		Msg  string `json:"message"`
	}{}
	if err := d.client.Do(c, req, &reply); err != nil {
		return errors.Wrapf(err, "BwsEndGame req(%v)", req)
	}
	if reply.Code != 0 {
		return errors.Wrapf(ecode.Int(reply.Code), "BwsEndGame req(%v) reply(%v)", req, reply)
	}
	return nil
}

// sign calc appkey and appsecret sign.
func (d *dao) sign(params url.Values) (query string, err error) {
	key := d.httpCfg.DanceClient.Key
	secret := d.httpCfg.DanceClient.Secret
	if params == nil {
		params = url.Values{}
	}
	params.Set("appkey", key)
	if params.Get("appsecret") != "" {
		log.Warn("utils http get must not have parameter appSecret")
	}
	if params.Get("ts") == "" {
		params.Set("ts", strconv.FormatInt(time.Now().Unix(), 10))
	}
	tmp := params.Encode()
	if strings.IndexByte(tmp, '+') > -1 {
		tmp = strings.Replace(tmp, "+", "%20", -1)
	}
	var b bytes.Buffer
	b.WriteString(tmp)
	b.WriteString(secret)
	mh := md5.Sum(b.Bytes())
	// query
	var qb bytes.Buffer
	qb.WriteString(tmp)
	qb.WriteString("&sign=")
	qb.WriteString(hex.EncodeToString(mh[:]))
	query = qb.String()
	return
}

func (d *dao) BwsReset(c context.Context) {
	var (
		params = url.Values{}
		url    = d.conf.BwsCfg.Host + _resetRoom
	)
	params.Set("gameId", fmt.Sprintf("%d", d.conf.BwsCfg.GameId))
	req, err := d.client.NewRequest(http.MethodPost, url, "", params)
	if err != nil {
		log.Error("BwsReset req(%v) err(%v)", req, err)
		return
	}
	var reply = struct {
		Code int    `json:"code"`
		Msg  string `json:"message"`
		Data bool   `json:"data"`
	}{}
	if err := d.client.Do(c, req, &reply); err != nil {
		log.Error("BwsReset req(%v) err(%v)", req, err)
		return
	}
	if reply.Code != 0 || !reply.Data {
		log.Error("BwsReset reply(%v)", reply)
		return
	}
	return
}
