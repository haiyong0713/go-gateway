package system

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"go-common/library/log"
	"go-gateway/app/web-svr/activity/ecode"
	"io/ioutil"
	"net/http"
)

// 获取微信AccessToken 如果没有或异常 会重新创建AccessToken
func (d *Dao) GetWXAccessToken(ctx context.Context, from string) (accessToken string, err error) {
	if accessToken, err = d.GetWXAccessTokenFromRedis(ctx, from); err != nil {
		err = ecode.SystemGetWXAccessTokenErr
		return
	}
	if accessToken == "" {
		var newToken string
		if newToken, err = d.CreateWXAccessToken(ctx, from); err != nil {
			err = ecode.SystemGetWXAccessTokenErr
			return
		}
		_ = d.StoreWXAccessTokenInRedis(ctx, newToken, from)
		accessToken = newToken
	}
	return
}

// HTTP 调用企业微信接口创建AccessToken
func (d *Dao) CreateWXAccessToken(ctx context.Context, from string) (accessToken string, err error) {
	res := new(struct {
		Errcode     int    `json:"errcode"`
		Errmsg      string `json:"errmsg"`
		AccessToken string `json:"access_token"`
		ExpiresIn   int    `json:"expires_in"`
	})
	var resp string
	params := map[string]string{"corpid": d.c.System.CORPID, "corpsecret": d.c.System.CORPSecret[from]}
	if resp, err = d.HTTPGet(ctx, d.c.System.WXCreateTokenUrl, params, map[string]string{}); err != nil {
		err = fmt.Errorf("CreateWXAccessToken HTTPGet Params:%v Resp:%v Err:%v", params, resp, err)
		log.Errorc(ctx, err.Error())
		return
	}
	if err = json.Unmarshal([]byte(resp), res); err != nil {
		err = fmt.Errorf("CreateWXAccessToken json.Unmarshal Resp:%v Err:%v", resp, err)
		log.Errorc(ctx, err.Error())
		return
	}
	if res.Errcode != 0 || res.AccessToken == "" {
		err = fmt.Errorf("CreateWXAccessToken Response Err Res:%v", res)
		log.Errorc(ctx, err.Error())
		return
	}
	accessToken = res.AccessToken
	return
}

// 获取企业微信UserID
func (d *Dao) GetWXUserUserIDByAccessTokenAndCode(ctx context.Context, accessToken string, code string) (userID string, err error) {
	res := new(struct {
		Errcode  int    `json:"errcode"`
		Errmsg   string `json:"errmsg"`
		UserID   string `json:"UserId"`
		DeviceID string `json:"DeviceId"`
	})
	var resp string
	params := map[string]string{"access_token": accessToken, "code": code}
	if resp, err = d.HTTPGet(ctx, d.c.System.WXGetUserUserIDUrl, params, map[string]string{}); err != nil {
		log.Errorc(ctx, "GetWXUserUserIDByAccessTokenAndCode HTTPGet Err Params:%v Resp:%v Err:%v", params, resp, err)
		err = ecode.SystemGetWXUserIDFailed
		return
	}
	if err = json.Unmarshal([]byte(resp), res); err != nil {
		log.Errorc(ctx, "GetWXUserUserIDByAccessTokenAndCode json.Unmarshal Resp:%v Err:%v", resp, err)
		err = ecode.SystemGetWXUserIDFailed
		return
	}
	if res.Errcode != 0 || res.UserID == "" {
		log.Errorc(ctx, "GetWXUserUserIDByAccessTokenAndCode Response Err Res:%v", res)
		err = ecode.SystemGetWXUserIDFailed
		return
	}
	userID = res.UserID
	return
}

// 创建JSAPITicket
func (d *Dao) CreateWXJSAPITicket(ctx context.Context, from string) (ticket string, err error) {
	var accessToken string
	if accessToken, err = d.GetWXAccessToken(ctx, from); err != nil {
		return
	}
	res := new(struct {
		Errcode   int    `json:"errcode"`
		Errmsg    string `json:"errmsg"`
		Ticket    string `json:"ticket"`
		ExpiresIn int    `json:"expires_in"`
	})
	var resp string
	params := map[string]string{"access_token": accessToken}
	if resp, err = d.HTTPGet(ctx, d.c.System.WXCreateJSAPITicketUrl, params, map[string]string{}); err != nil {
		err = fmt.Errorf("CreateWXJSAPITicket HTTPGet Params:%v Resp:%v Err:%v", params, resp, err)
		log.Errorc(ctx, err.Error())
		return
	}
	if err = json.Unmarshal([]byte(resp), res); err != nil {
		err = fmt.Errorf("CreateWXJSAPITicket json.Unmarshal Resp:%v Err:%v", resp, err)
		log.Errorc(ctx, err.Error())
		return
	}
	if res.Errcode != 0 || res.Ticket == "" {
		err = fmt.Errorf("CreateWXJSAPITicket Response Err Res:%v", res)
		log.Errorc(ctx, err.Error())
		return
	}
	ticket = res.Ticket
	return
}

// 获取微信JSAPITicket 如果没有或异常 会重新创建JSAPITicket
func (d *Dao) GetWXJSAPITicket(ctx context.Context, from string) (JSAPITicket string, err error) {
	if JSAPITicket, err = d.GetWXJSAPITicketFromRedis(ctx, from); err != nil {
		err = ecode.SystemGetWXJSAPITicket
		return
	}
	if JSAPITicket == "" {
		var newJSAPITicket string
		if newJSAPITicket, err = d.CreateWXJSAPITicket(ctx, from); err != nil {
			err = ecode.SystemGetWXJSAPITicket
			return
		}
		_ = d.StoreWXJSAPITicketInRedis(ctx, newJSAPITicket, from)
		JSAPITicket = newJSAPITicket
	}
	return
}

func (d *Dao) HTTPGet(ctx context.Context, url string, params map[string]string, header map[string]string) (res string, err error) {
	// 处理query
	query := ""
	isFirst := true
	if len(params) > 0 {
		for k, v := range params {
			if isFirst {
				query += "?"
				isFirst = false
			} else {
				query += "&"
			}
			query += k + "=" + v
		}
	}
	req, _ := http.NewRequest("GET", url+query, nil)
	// 处理header
	if len(header) > 0 {
		for k, v := range header {
			req.Header.Set(k, v)
		}
	}
	resp, err := (&http.Client{}).Do(req)
	if err != nil {
		err = fmt.Errorf("[HTTPGET]Err url:%s error:%v", url, err)
		log.Warnc(ctx, err.Error())
		return
	}
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	res = string(body)
	log.Infoc(ctx, "[HTTPGET]SUCC url:%s res:%s", url+query, res)

	return
}

func (d *Dao) HTTPPost(ctx context.Context, url string, params string, header map[string]string) (res string, err error) {
	reqBody := bytes.NewBuffer([]byte(params))
	req, _ := http.NewRequest("POST", url, reqBody)
	// 处理header
	if len(header) > 0 {
		for k, v := range header {
			req.Header.Set(k, v)
		}
	}
	resp, err := (&http.Client{}).Do(req)
	fmt.Println(resp)
	fmt.Println(err)
	if err != nil {
		err = fmt.Errorf("[HTTPPOST]Err url:%s body:%+v error:%v", url, params, err)
		log.Warnc(ctx, err.Error())
		return
	}
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	res = string(body)
	log.Infoc(ctx, "[HTTPPOST]SUCC url:%s res:%s", url+params, res)

	return
}
