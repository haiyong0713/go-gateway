package bind

import (
	"context"
	"encoding/json"
	api2 "git.bilibili.co/bapis/bapis-go/account/service"
	api3 "git.bilibili.co/bapis/bapis-go/account/service/oauth2"
	api "git.bilibili.co/bapis/bapis-go/passport/service/sns"
	xecode "go-common/library/ecode"
	"go-common/library/log"
	"go-gateway/app/web-svr/activity/interface/client"
	"go-gateway/app/web-svr/activity/interface/model/bind"
)

func (d *Dao) GetCode(ctx context.Context, clientId string, business string, mid int64) (code string, err error) {
	resp, err := client.PassportClient.CommonAuthorizeCode(ctx, &api.CommonAuthorizeCodeReq{
		ClientId: clientId,
		Business: business,
		Mid:      mid,
	})
	if err != nil {
		log.Errorc(ctx, "[GetCode][CommonAuthorizeCode][Error], err:%+v", err)
		return
	}
	if resp == nil || resp.Code == "" {
		err = xecode.Errorf(xecode.ServerErr, "code生成异常")
		log.Errorc(ctx, "[GetCode][CommonAuthorizeCode][Error], err:%+v", err)
		return
	}
	code = resp.Code
	return
}

func (d *Dao) GetUserInfo(ctx context.Context, mid int64) (userInfo *api2.InfoReply, err error) {
	return client.AccountClient.Info3(ctx, &api2.MidReq{
		Mid: mid,
	})
}

func (d *Dao) GetTencentBindInfo(ctx context.Context, clientId string, business string, oid string, actId string, refresh int32, mid int64) (bindInfo *bind.BindInfo, err error) {
	resp, err := client.PassportClient.CommonUserBindInfo(ctx, &api.CommonUserBindInfoReq{
		Mid:      mid,
		ClientId: clientId,
		Business: business,
		Oid:      oid,
		Refresh:  refresh,
		Actid:    actId,
	})
	if err != nil {
		log.Errorc(ctx, "[GetTencentBindInfo][CommonUserBindInfo][Error], err:%+v", err)
		return
	}

	bindInfo = new(bind.BindInfo)
	bindInfo.BindType = bind.IsBindFalse
	if resp.BindInfo == "" {
		return
	}
	tencentGameBindInfo := new(bind.TencentGameBindInfo)
	if err = json.Unmarshal([]byte(resp.BindInfo), &tencentGameBindInfo); err != nil {
		log.Errorc(ctx, "[GetTencentBindInfo][CommonUserBindInfo][Unmarshal][Error], err:%+v, resp:%+v", err, resp)
		return
	}
	if tencentGameBindInfo.GameAcc != nil {
		bindInfo.AccountInfo = &bind.AccountInfo{
			AccountType: tencentGameBindInfo.GameAcc.Type,
		}
	}
	bindInfo.BindType = bind.IsBindFalse
	if tencentGameBindInfo.IsBind ||
		(tencentGameBindInfo.GameAcc != nil && tencentGameBindInfo.GameAcc.Type != "") {
		bindInfo.BindType = bind.IsBindTrue
	}
	if tencentGameBindInfo.GameRole != nil {
		bindInfo.RoleInfo = &bind.RoleInfo{
			RoleName:      tencentGameBindInfo.GameRole.RoleName,
			AreaName:      tencentGameBindInfo.GameRole.AreaName,
			PartitionName: tencentGameBindInfo.GameRole.PartitionName,
			PlatName:      tencentGameBindInfo.GameRole.PlatName,
		}
	}
	return
}

func (d *Dao) GetOpenIdByMid(ctx context.Context, appKey string, mid int64) (openId string, err error) {
	resp, err := client.BiliOAuth2Client.UserOpenID(ctx, &api3.UserOpenIDReq{
		Oauth2Appkey: appKey,
		Mid:          mid,
	})
	// 122003 标识用户未授权
	if err != nil {
		log.Errorc(ctx, "[GetOpenIdByMid][UserOpenID][Error], err:%+v", err)
		return
	}
	openId = resp.Openid
	return
}

func (d *Dao) GetMidByOpenId(ctx context.Context, appKey string, openId string) (mid int64, err error) {
	resp, err := client.BiliOAuth2Client.MidByOpenID(ctx, &api3.MidByOpenIDReq{
		Oauth2Appkey: appKey,
		Openid:       openId,
	})
	// 122003 标识用户未授权
	if err != nil {
		log.Errorc(ctx, "[GetMidByOpenId][MidByOpenID][Error], err:%+v", err)
		return
	}
	mid = resp.Mid
	return
}
