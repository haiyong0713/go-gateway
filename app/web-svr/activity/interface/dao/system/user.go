package system

import (
	"context"
	"encoding/json"
	"fmt"
	"go-common/library/log"
	"go-gateway/app/web-svr/activity/ecode"
	model "go-gateway/app/web-svr/activity/interface/model/system"
	"strconv"
)

func WXAccessTokenKey(from string) string {
	return fmt.Sprintf("activity:system:wx:%s:access_token", from)
}

func WXJSAPITicketKey(from string) string {
	return fmt.Sprintf("activity:system:wx:%s:jsapi_ticket", from)
}

func OAAccessTokenKey() string {
	return "activity:system:oa:access_token"
}

// 因为企业微信跟OA同步有bug 企业微信的user_id可能是工号也可能是OA系统ID
func (d *Dao) GetUIDWithWXUserID(ctx context.Context, WXUserID string) (ID string, err error) {
	if res := d.GetUserInfoByWorkCode(ctx, WXUserID); res != nil {
		ID = res.WorkCode
		return
	}
	OAID, err := strconv.ParseInt(WXUserID, 10, 64)
	if err != nil {
		err = ecode.SystemNoUserInOAErr
		return
	}
	if res := d.GetUserInfoByOAID(ctx, OAID); res != nil {
		ID = res.WorkCode
		return
	}
	log.Warnc(ctx, "GetUIDWithWXUserID From Memory No WXUserID:%v", WXUserID)

	err = ecode.SystemNoUserInOAErr
	return
}

// 从内存获取用户信息
func (d *Dao) GetUserInfoWithUID(ctx context.Context, uid string) (res *model.User, err error) {
	if res = d.GetUserInfoByWorkCode(ctx, uid); res != nil {
		return
	}
	log.Warnc(ctx, "GetUserInfo From Memory No WorkCode:%v", uid)

	err = ecode.SystemNoUserInOAErr
	return
}

// 通过工卡号获取用户信息
func (d *Dao) GetUserInfoByWorkCode(ctx context.Context, query string) (res *model.User) {
	if v, ok := d.MapKeyWorkCode[query]; ok {
		res = v
		return
	}
	return
}

// 通过OAID获取用户信息
func (d *Dao) GetUserInfoByOAID(ctx context.Context, OAID int64) (res *model.User) {
	if v, ok := d.MapKeyOAID[OAID]; ok {
		res = v
		return
	}
	return
}

func (d *Dao) BuildMapKeyWorkCode(ctx context.Context, data []*model.User) (err error) {
	if len(data) == 0 {
		err = fmt.Errorf("BuildMapKeyWorkCode Data Empty")
		return
	}
	bucket := make(map[string]*model.User)
	for _, v := range data {
		if v.WorkCode == "" {
			continue
		}
		bucket[v.WorkCode] = v
	}
	d.MapKeyWorkCode = bucket
	return
}

func (d *Dao) BuildMapKeyOAID(ctx context.Context, data []*model.User) (err error) {
	if len(data) == 0 {
		err = fmt.Errorf("BuildMapKeyOAID Data Empty")
		return
	}
	bucket := make(map[int64]*model.User)
	for _, v := range data {
		if v.ID == 0 {
			continue
		}
		bucket[v.ID] = v
	}
	d.MapKeyOAID = bucket
	return
}

// 从cookie获取用户详细信息
func (d *Dao) GetUserInfoWithCookie(ctx context.Context, sessionToken string) (res *model.SystemUser, err error) {
	if sessionToken == "" {
		err = ecode.SystemNoTokenErr
		return
	}
	DBUser, err := d.GetDBUserInfoByToken(ctx, sessionToken)
	if DBUser == nil || DBUser.ID == 0 {
		log.Warnc(ctx, "GetDBUserInfoByToken No DBUser Token:%v", sessionToken)
		err = ecode.SystemNoUserErr
		return
	}
	if err != nil {
		err = ecode.SystemNetWorkBuzyErr
		return
	}
	user, err := d.GetUserInfoWithUID(ctx, DBUser.UID)
	if err != nil {
		return
	}

	res = &model.SystemUser{UID: DBUser.UID, NickName: user.NickName, Avatar: user.Avatar, DepartmentName: user.DepartmentName, LastName: user.LastName, UseKind: user.UseKind}
	return
}

// 根据cookie获取用户uid
func (d *Dao) GetUserIDWithCookie(ctx context.Context, sessionToken string) (uid string, err error) {
	if sessionToken == "" {
		err = ecode.SystemNoTokenErr
		return
	}
	DBUser, err := d.GetDBUserInfoByToken(ctx, sessionToken)
	if DBUser == nil || DBUser.ID == 0 {
		log.Warnc(ctx, "GetDBUserInfoByToken No DBUser Token:%v", sessionToken)
		err = ecode.SystemNoUserErr
		return
	}
	if err != nil {
		err = ecode.SystemNetWorkBuzyErr
		return
	}

	uid = DBUser.UID
	return
}

// 调用企业微信员工接口获取用户信息
func (d *Dao) GetWXUserInfo(ctx context.Context, userID string) (res *model.WXUserDetail, err error) {
	accessToken, err := d.GetWXAccessToken(ctx, "2021party")
	if err != nil {
		return
	}
	url := "https://qyapi.weixin.qq.com/cgi-bin/user/get"
	params := map[string]string{"access_token": accessToken, "userid": userID}
	str, err := d.HTTPGet(ctx, url, params, map[string]string{})
	if err != nil {
		log.Errorc(ctx, "GetWXUserInfo HTTPGet Err params(%+v) res(%s) err(%+v)", params, str, err)
		return
	}
	res = new(model.WXUserDetail)
	if err = json.Unmarshal([]byte(str), res); err != nil {
		log.Errorc(ctx, "GetWXUserInfo json.Unmarshal Err str(%s) err(%+v)", str, err)
		return
	}
	if res.Errcode != 0 || res.Userid == "" {
		err = fmt.Errorf("GetWXUserInfo res.Errcode == 0 || res.Userid == '' res(%+v)", res)
		log.Errorc(ctx, err.Error())
		return
	}
	return
}

// 年会白名单
//func (d *Dao) isWhite(ctx context.Context, userInfo *model.User) bool {
//	if userInfo.UseKind == "全职" || userInfo.UseKind == "实习生转正" {
//		return true
//	}
//}
