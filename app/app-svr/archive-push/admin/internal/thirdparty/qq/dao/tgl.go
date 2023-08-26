package dao

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"go-common/library/log"

	qqModel "go-gateway/app/app-svr/archive-push/admin/internal/thirdparty/qq/model"
	"go-gateway/app/app-svr/archive-push/ecode"
)

const (
	ContributeVideoURL         = "/contribute/video"
	OauthAccessTokenURL        = "/oauth/access_token"
	maxRefreshAccessTokenTimes = 10
)

// ContributeVideo /contribute/video 视频投稿
func (d *Dao) ContributeVideo(req *qqModel.ContributeVideoReq) (reply *qqModel.ContributeVideoReply, err error) {
	if d.httpClient == nil {
		d.httpClient = http.DefaultClient
	}
	if req == nil {
		return nil, ecode.PushRequestError
	}
	req.Action = d.Cfg.TGL.Action
	req.GameID = d.Cfg.TGL.GameID
	var reqJSON []byte
	if reqJSON, err = json.Marshal(req); err != nil {
		log.Error("qq tgl: ContributeVideo Marshal error (%v)", err)
		return
	}
	toRequestUrl := d.Cfg.TGL.Host + ContributeVideoURL

	httpReq, _ := http.NewRequest("POST", toRequestUrl, strings.NewReader(string(reqJSON)))
	httpReq.Header.Set(qqModel.ContentTypeKey, qqModel.ContentTypeJSON)
	httpReq.Header.Set(qqModel.TGLAccessTokenHeader, fmt.Sprintf(qqModel.TGLAccessTokenBody, d.TGLAccessToken))

	log.Info("qq tgl: ContributeVideo posting URL=%s", httpReq.URL.String())
	log.Info("qq tgl: ContributeVideo posting Data=%s", string(reqJSON))
	var resRaw *http.Response
	if resRaw, err = d.httpClient.Do(httpReq); err != nil {
		log.Error("qq tgl: ContributeVideo Do Request Error (%v)", err)
		return
	} else if resRaw.Close {
		log.Error("qq tgl: ContributeVideo response closed")
	} else if resRaw.Body != nil {
		defer resRaw.Body.Close()
		resBytes, _ := ioutil.ReadAll(resRaw.Body)
		log.Info("qq tgl: ContributeVideo Response: %+v", string(resBytes))
		reply = &qqModel.ContributeVideoReply{}
		if err = json.Unmarshal(resBytes, reply); err != nil {
			log.Error("qq tgl: ContributeVideo Unmarshal response Error (%v)", err)
		}
	}
	return
}

// GetAccessToken /oauth/access_token 获取Token
func (d *Dao) GetAccessToken() (token string, err error) {
	if d.httpClient == nil {
		d.httpClient = http.DefaultClient
	}
	req := &qqModel.OauthAccessTokenReq{
		GrantType:    qqModel.TGLAccessTokenGrantType,
		ClientID:     d.Cfg.TGL.Oauth2.ClientID,
		ClientSecret: d.Cfg.TGL.Oauth2.Secret,
	}
	var reqJSON []byte
	if reqJSON, err = json.Marshal(req); err != nil {
		log.Error("qq tgl: GetAccessToken Marshal error (%v)", err)
		return
	}
	toRequestUrl := d.Cfg.TGL.Host + OauthAccessTokenURL

	httpReq, _ := http.NewRequest("POST", toRequestUrl, strings.NewReader(string(reqJSON)))
	httpReq.Header.Set(qqModel.ContentTypeKey, qqModel.ContentTypeJSON)

	log.Info("qq tgl: GetAccessToken posting URL=%s", httpReq.URL.String())
	log.Info("qq tgl: GetAccessToken posting Data=%s", string(reqJSON))
	var resRaw *http.Response
	if resRaw, err = d.httpClient.Do(httpReq); err != nil {
		log.Error("qq tgl: GetAccessToken Do Request Error (%v)", err)
		return
	} else if resRaw.Close {
		log.Error("qq tgl: GetAccessToken response closed")
	} else if resRaw.Body != nil {
		defer resRaw.Body.Close()
		resBytes, _ := ioutil.ReadAll(resRaw.Body)
		log.Info("qq tgl: GetAccessToken Response: %+v", string(resBytes))
		reply := &qqModel.OauthAccessTokenReply{}
		if err = json.Unmarshal(resBytes, reply); err != nil {
			log.Error("qq tgl: GetAccessToken Unmarshal response Error (%v)", err)
			return
		}
		if reply.Status > 200 || reply.AccessToken == "" {
			err = ecode.GetAccessTokenError
			return
		} else {
			token = reply.AccessToken
			return
		}
	}
	return
}

func (d *Dao) RefreshAccessToken(times int) (err error) {
	if times >= maxRefreshAccessTokenTimes {
		return ecode.GetAccessTokenError
	}
	log.Info("qq tgl: RefreshAccessToken %d Start", times)
	var gotToken string
	if gotToken, err = d.GetAccessToken(); err != nil {
		log.Error("qq tgl: RefreshAccessToken %d error %v", times, err)
		time.Sleep(5 * time.Second)
		err = d.RefreshAccessToken(times + 1)
		return
	}
	d.TGLAccessToken = gotToken
	log.Info("qq tgl: RefreshAccessToken %d End", times)
	return
}

func (d *Dao) startAccessTokenTicker() (ticker *time.Ticker) {
	ticker = time.NewTicker(40 * time.Minute)
	go func() {
		for range ticker.C {
			if err := d.RefreshAccessToken(0); err != nil {
				log.Error("qq tgl: RefreshAccessToken 达到最大重试上限")
			}
		}
	}()

	return
}
