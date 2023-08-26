package dao

import (
	"context"
	accountGRPC "git.bilibili.co/bapis/bapis-go/account/service"
	"go-gateway/app/app-svr/archive-push/ecode"
	"net/http"
	"net/url"
	"strconv"

	"go-common/library/log"

	"go-gateway/app/app-svr/archive-push/admin/internal/model"
)

const (
	GetOpenIDByMIDURL = "/x/internal/account-oauth2/openid"
	GetMIDByUIDURL    = "/x/internal/account-oauth2/inner-auth/mid"
	GetMidByOpenIDURL = "/x/internal/account-oauth2/mid/by/openid"
)

// GetOpenIDByMID 根据用户MID获取Open ID
func (d *Dao) GetOpenIDByMID(mid int64, oauth2AppKey string) (res *model.GetOpenIDByMIDResp, err error) {
	var req *http.Request
	params := url.Values{}
	params.Set("mid", strconv.FormatInt(mid, 10))
	params.Set("oauth2_appkey", oauth2AppKey)
	if req, err = d.bmClient.NewRequest("GET", d.hosts.API+GetOpenIDByMIDURL, "", params); err != nil {
		log.Error("Dao: GetOpenIDByMID(%d, %s) bmClient.NewRequest error(%v)", mid, oauth2AppKey, err)
		return
	}
	res = &model.GetOpenIDByMIDResp{}
	if err = d.bmClient.Do(context.Background(), req, &res); err != nil {
		log.Error("Dao: GetOpenIDByMID(%d, %s) bmClient.Do error(%v)", mid, oauth2AppKey, err)
		return
	}
	return
}

// GetMIDByUID 根据用户UID获取MID
func (d *Dao) GetMIDByUID(uid string, oauth2AppKey string) (res *model.GetMIDByUIDResp, err error) {
	var req *http.Request
	params := url.Values{}
	params.Set("uid", uid)
	params.Set("oauth2_appkey", oauth2AppKey)
	if req, err = d.bmClient.NewRequest("GET", d.hosts.API+GetMIDByUIDURL, "", params); err != nil {
		log.Error("Dao: GetMIDByUID(%s, %s) bmClient.NewRequest error(%v)", uid, oauth2AppKey, err)
		return
	}
	res = &model.GetMIDByUIDResp{}
	if err = d.bmClient.Do(context.Background(), req, &res); err != nil {
		log.Error("Dao: GetMIDByUID(%s, %s) bmClient.Do error(%v)", uid, oauth2AppKey, err)
		return
	}
	return
}

// GetMIDByUID 根据用户UID获取MID
func (d *Dao) GetMIDByOpenID(openID string, oauth2AppKey string) (res *model.GetMIDByOpenIDResp, err error) {
	var req *http.Request
	params := url.Values{}
	params.Set("openid", openID)
	params.Set("oauth2_appkey", oauth2AppKey)
	if req, err = d.bmClient.NewRequest("GET", d.hosts.API+GetMidByOpenIDURL, "", params); err != nil {
		log.Error("Dao: GetMIDByOpenID(%s, %s) bmClient.NewRequest error(%v)", openID, oauth2AppKey, err)
		return
	}
	res = &model.GetMIDByOpenIDResp{}
	if err = d.bmClient.Do(context.Background(), req, &res); err != nil {
		log.Error("Dao: GetMIDByOpenID(%s, %s) bmClient.Do error(%v)", openID, oauth2AppKey, err)
		return
	}
	return
}

// GetAccountInfoByMID 根据用户MID查询账号信息
func (d *Dao) GetAccountInfoByMID(mid int64) (res *accountGRPC.Info, err error) {
	if mid == 0 {
		err = ecode.AccountPlatRequestError
		return
	}
	var infoReply *accountGRPC.InfoReply
	req := &accountGRPC.MidReq{Mid: mid}
	if infoReply, err = d.accountGRPCClient.Info3(context.Background(), req); err != nil {
		return
	} else if infoReply == nil {
		err = ecode.AccountPlatResponseError
		return
	}
	res = infoReply.Info

	return
}
