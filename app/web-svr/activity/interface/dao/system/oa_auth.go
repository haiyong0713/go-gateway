package system

import (
	"context"
	"encoding/json"
	"fmt"
	"go-common/library/log"
	"go-gateway/app/web-svr/activity/ecode"
	model "go-gateway/app/web-svr/activity/interface/model/system"
)

// 获取微信AccessToken 如果没有或异常 会重新创建AccessToken
func (d *Dao) GetOAAccessToken(ctx context.Context) (accessToken string, err error) {
	if accessToken, err = d.GetOAAccessTokenFromRedis(ctx); err != nil {
		err = ecode.SystemGetOAAccessTokenErr
		return
	}
	if accessToken == "" {
		var newToken string
		if newToken, err = d.CreateOAAccessToken(ctx); err != nil {
			err = ecode.SystemGetOAAccessTokenErr
			return
		}
		_ = d.StoreOAAccessTokenInRedis(ctx, newToken)
		accessToken = newToken
	}
	return
}

// HTTP 调用企业微信接口创建AccessToken
func (d *Dao) CreateOAAccessToken(ctx context.Context) (accessToken string, err error) {
	res := new(struct {
		Code int `json:"code"`
		Data struct {
			ExpiredInS int    `json:"expired_in(s)"`
			Token      string `json:"token"`
		} `json:"data"`
		Msg string `json:"msg"`
	})
	var resp string
	params := map[string]string{"client": d.c.System.OAClient, "secret": d.c.System.OASecret}
	if resp, err = d.HTTPGet(ctx, d.c.System.OACreateTokenUrl, params, map[string]string{}); err != nil {
		err = fmt.Errorf("CreateOAAccessToken HTTPGet Params:%v Resp:%v Err:%v", params, resp, err)
		log.Errorc(ctx, err.Error())
		return
	}
	if err = json.Unmarshal([]byte(resp), res); err != nil {
		err = fmt.Errorf("CreateOAAccessToken json.Unmarshal Resp:%v Err:%v", resp, err)
		log.Errorc(ctx, err.Error())
		return
	}
	if res.Code != 0 || res.Data.Token == "" {
		err = fmt.Errorf("CreateOAAccessToken Response Err Res:%v", res)
		log.Errorc(ctx, err.Error())
		return
	}
	accessToken = res.Data.Token
	return
}

func (d *Dao) GetOAAllUsersInfo(ctx context.Context) (data []*model.User, err error) {
	res := new(struct {
		Code int           `json:"code"`
		Data []*model.User `json:"data"`
		Msg  string        `json:"msg"`
	})
	params := map[string]string{}
	var accessToken string
	if accessToken, err = d.GetOAAccessToken(ctx); err != nil {
		return
	}
	var resp string
	if resp, err = d.HTTPGet(ctx, d.c.System.OAGetAllUsersInfoUrl, params, map[string]string{"authorization": "Bearer " + accessToken}); err != nil {
		err = fmt.Errorf("GetOAAllUsersInfo HTTPGet Params:%v Resp:%v Err:%v", params, resp, err)
		log.Errorc(ctx, err.Error())
		return
	}
	if err = json.Unmarshal([]byte(resp), res); err != nil {
		err = fmt.Errorf("GetOAAllUsersInfo json.Unmarshal Resp:%v Err:%v", resp, err)
		log.Errorc(ctx, err.Error())
		return
	}
	if res.Code != 0 || len(res.Data) == 0 {
		err = fmt.Errorf("GetOAAllUsersInfo Response Err Res:%v", res)
		log.Errorc(ctx, err.Error())
		return
	}
	data = res.Data
	return
}
